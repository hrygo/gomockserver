//go:build integration
// +build integration

package repository

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/gomockserver/mockserver/internal/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// setupEnvironmentTestDB 设置环境测试数据库
func setupEnvironmentTestDB(t *testing.T) (*mongo.Client, *mongo.Database, EnvironmentRepository) {
	ctx := context.Background()

	uri := "mongodb://localhost:27017"
	if envURI := os.Getenv("MONGODB_URI"); envURI != "" {
		uri = envURI
	}

	clientOpts := options.Client().ApplyURI(uri).SetServerSelectionTimeout(2 * time.Second)
	client, err := mongo.Connect(ctx, clientOpts)
	if err != nil {
		t.Skipf("Skipping integration test: MongoDB not available - %v", err)
		return nil, nil, nil
	}

	err = client.Ping(ctx, nil)
	if err != nil {
		client.Disconnect(ctx)
		t.Skipf("Skipping integration test: cannot ping MongoDB - %v", err)
		return nil, nil, nil
	}

	db := client.Database("mockserver_test_env_" + primitive.NewObjectID().Hex())
	t.Logf("Using test database: %s", db.Name())

	collection := db.Collection("environments")
	repo := &environmentRepository{collection: collection}

	return client, db, repo
}

// teardownEnvironmentTestDB 清理环境测试数据库
func teardownEnvironmentTestDB(t *testing.T, client *mongo.Client, db *mongo.Database) {
	if db != nil {
		ctx := context.Background()
		err := db.Drop(ctx)
		if err != nil {
			t.Logf("Warning: failed to drop test database: %v", err)
		}
	}

	if client != nil {
		ctx := context.Background()
		err := client.Disconnect(ctx)
		if err != nil {
			t.Logf("Warning: failed to disconnect: %v", err)
		}
	}
}

// TestEnvironmentRepository_Create_Success 测试成功创建环境
func TestEnvironmentRepository_Create_Success(t *testing.T) {
	client, db, repo := setupEnvironmentTestDB(t)
	if client == nil {
		return
	}
	defer teardownEnvironmentTestDB(t, client, db)

	ctx := context.Background()

	tests := []struct {
		name string
		env  *models.Environment
	}{
		{
			name: "最小字段环境",
			env: &models.Environment{
				Name:      "开发环境",
				ProjectID: "project-001",
			},
		},
		{
			name: "包含BaseURL的环境",
			env: &models.Environment{
				Name:      "测试环境",
				ProjectID: "project-001",
				BaseURL:   "http://localhost:9090",
			},
		},
		{
			name: "包含变量的环境",
			env: &models.Environment{
				Name:      "生产环境",
				ProjectID: "project-001",
				BaseURL:   "https://api.prod.example.com",
				Variables: map[string]interface{}{
					"api_key":     "prod-key-123",
					"timeout":     30,
					"retry_count": 3,
				},
			},
		},
		{
			name: "中文名称环境",
			env: &models.Environment{
				Name:      "预发布环境-中文",
				ProjectID: "project-002",
				BaseURL:   "http://pre.example.com",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := repo.Create(ctx, tt.env)
			require.NoError(t, err, "创建环境应该成功")

			assert.NotEmpty(t, tt.env.ID, "ID应该被设置")
			assert.False(t, tt.env.CreatedAt.IsZero(), "CreatedAt应该被设置")
			assert.False(t, tt.env.UpdatedAt.IsZero(), "UpdatedAt应该被设置")

			// 验证ID格式
			_, err = primitive.ObjectIDFromHex(tt.env.ID)
			assert.NoError(t, err, "ID应该是有效的ObjectID格式")
		})
	}
}

// TestEnvironmentRepository_FindByID_EdgeCases 测试查询环境的边界情况
func TestEnvironmentRepository_FindByID_EdgeCases(t *testing.T) {
	client, db, repo := setupEnvironmentTestDB(t)
	if client == nil {
		return
	}
	defer teardownEnvironmentTestDB(t, client, db)

	ctx := context.Background()

	// 创建测试环境
	env := &models.Environment{
		Name:      "测试环境",
		ProjectID: "project-001",
	}
	err := repo.Create(ctx, env)
	require.NoError(t, err)

	tests := []struct {
		name        string
		id          string
		shouldExist bool
		shouldError bool
	}{
		{
			name:        "有效ID查询存在的环境",
			id:          env.ID,
			shouldExist: true,
			shouldError: false,
		},
		{
			name:        "有效ID查询不存在的环境",
			id:          "507f1f77bcf86cd799439011",
			shouldExist: false,
			shouldError: false,
		},
		{
			name:        "无效的ObjectID格式",
			id:          "invalid-id",
			shouldExist: false,
			shouldError: true,
		},
		{
			name:        "空ID",
			id:          "",
			shouldExist: false,
			shouldError: true,
		},
		{
			name:        "短ID",
			id:          "123",
			shouldExist: false,
			shouldError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			found, err := repo.FindByID(ctx, tt.id)

			if tt.shouldError {
				assert.Error(t, err, "应该返回错误")
				return
			}

			assert.NoError(t, err, "不应该返回错误")

			if tt.shouldExist {
				require.NotNil(t, found, "应该找到环境")
				assert.Equal(t, tt.id, found.ID)
			} else {
				assert.Nil(t, found, "不应该找到环境")
			}
		})
	}
}

// TestEnvironmentRepository_Update_Success 测试更新环境成功场景
func TestEnvironmentRepository_Update_Success(t *testing.T) {
	client, db, repo := setupEnvironmentTestDB(t)
	if client == nil {
		return
	}
	defer teardownEnvironmentTestDB(t, client, db)

	ctx := context.Background()

	// 创建初始环境
	env := &models.Environment{
		Name:      "初始环境",
		ProjectID: "project-001",
		BaseURL:   "http://old.example.com",
		Variables: map[string]interface{}{
			"old_var": "old_value",
		},
	}
	err := repo.Create(ctx, env)
	require.NoError(t, err)

	originalID := env.ID
	originalCreatedAt := env.CreatedAt
	time.Sleep(10 * time.Millisecond)

	tests := []struct {
		name       string
		updateFunc func(*models.Environment)
		verifyFunc func(*testing.T, *models.Environment)
	}{
		{
			name: "更新所有字段",
			updateFunc: func(e *models.Environment) {
				e.Name = "更新后的环境"
				e.BaseURL = "http://new.example.com"
				e.Variables = map[string]interface{}{
					"new_var": "new_value",
					"count":   100,
				}
			},
			verifyFunc: func(t *testing.T, e *models.Environment) {
				assert.Equal(t, "更新后的环境", e.Name)
				assert.Equal(t, "http://new.example.com", e.BaseURL)
				assert.Contains(t, e.Variables, "new_var")
				assert.NotContains(t, e.Variables, "old_var")
			},
		},
		{
			name: "只更新名称",
			updateFunc: func(e *models.Environment) {
				e.Name = "仅名称更新"
			},
			verifyFunc: func(t *testing.T, e *models.Environment) {
				assert.Equal(t, "仅名称更新", e.Name)
			},
		},
		{
			name: "清空BaseURL",
			updateFunc: func(e *models.Environment) {
				e.BaseURL = ""
			},
			verifyFunc: func(t *testing.T, e *models.Environment) {
				assert.Empty(t, e.BaseURL)
			},
		},
		{
			name: "添加复杂变量",
			updateFunc: func(e *models.Environment) {
				e.Variables = map[string]interface{}{
					"nested": map[string]interface{}{
						"key1": "value1",
						"key2": 123,
					},
					"array": []interface{}{"item1", "item2"},
				}
			},
			verifyFunc: func(t *testing.T, e *models.Environment) {
				assert.Contains(t, e.Variables, "nested")
				assert.Contains(t, e.Variables, "array")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.updateFunc(env)

			err := repo.Update(ctx, env)
			require.NoError(t, err, "更新应该成功")

			// 验证ID和CreatedAt没有变化
			assert.Equal(t, originalID, env.ID, "ID不应该改变")
			assert.Equal(t, originalCreatedAt, env.CreatedAt, "CreatedAt不应该改变")
			assert.True(t, env.UpdatedAt.After(originalCreatedAt), "UpdatedAt应该更新")

			// 从数据库重新查询验证
			updated, err := repo.FindByID(ctx, env.ID)
			require.NoError(t, err)
			require.NotNil(t, updated)

			tt.verifyFunc(t, updated)
		})
	}
}

// TestEnvironmentRepository_Update_InvalidID 测试使用无效ID更新
func TestEnvironmentRepository_Update_InvalidID(t *testing.T) {
	client, db, repo := setupEnvironmentTestDB(t)
	if client == nil {
		return
	}
	defer teardownEnvironmentTestDB(t, client, db)

	ctx := context.Background()

	tests := []struct {
		name string
		id   string
	}{
		{"无效ID格式", "invalid-id"},
		{"空ID", ""},
		{"短ID", "123"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			env := &models.Environment{
				ID:        tt.id,
				Name:      "测试",
				ProjectID: "project-001",
			}

			err := repo.Update(ctx, env)
			assert.Error(t, err, "应该返回错误")
		})
	}
}

// TestEnvironmentRepository_Delete_Success 测试删除环境
func TestEnvironmentRepository_Delete_Success(t *testing.T) {
	client, db, repo := setupEnvironmentTestDB(t)
	if client == nil {
		return
	}
	defer teardownEnvironmentTestDB(t, client, db)

	ctx := context.Background()

	// 创建环境
	env := &models.Environment{
		Name:      "待删除环境",
		ProjectID: "project-001",
	}
	err := repo.Create(ctx, env)
	require.NoError(t, err)

	// 验证环境存在
	found, err := repo.FindByID(ctx, env.ID)
	require.NoError(t, err)
	require.NotNil(t, found)

	// 删除环境
	err = repo.Delete(ctx, env.ID)
	assert.NoError(t, err, "删除应该成功")

	// 验证环境已删除
	deleted, err := repo.FindByID(ctx, env.ID)
	assert.NoError(t, err)
	assert.Nil(t, deleted, "环境应该已被删除")
}

// TestEnvironmentRepository_Delete_NotExist 测试删除不存在的环境
func TestEnvironmentRepository_Delete_NotExist(t *testing.T) {
	client, db, repo := setupEnvironmentTestDB(t)
	if client == nil {
		return
	}
	defer teardownEnvironmentTestDB(t, client, db)

	ctx := context.Background()

	// 删除不存在的环境（应该不报错）
	err := repo.Delete(ctx, "507f1f77bcf86cd799439011")
	assert.NoError(t, err, "删除不存在的环境应该不报错")
}

// TestEnvironmentRepository_FindByProject 测试按项目查询环境
func TestEnvironmentRepository_FindByProject(t *testing.T) {
	client, db, repo := setupEnvironmentTestDB(t)
	if client == nil {
		return
	}
	defer teardownEnvironmentTestDB(t, client, db)

	ctx := context.Background()

	// 创建多个项目的环境
	environments := []*models.Environment{
		{Name: "环境1", ProjectID: "project-001", BaseURL: "http://env1.example.com"},
		{Name: "环境2", ProjectID: "project-001", BaseURL: "http://env2.example.com"},
		{Name: "环境3", ProjectID: "project-001", BaseURL: "http://env3.example.com"},
		{Name: "环境4", ProjectID: "project-002", BaseURL: "http://env4.example.com"},
		{Name: "环境5", ProjectID: "project-002", BaseURL: "http://env5.example.com"},
	}

	for _, e := range environments {
		err := repo.Create(ctx, e)
		require.NoError(t, err)
	}

	// 查询 project-001
	found1, err := repo.FindByProject(ctx, "project-001")
	require.NoError(t, err)
	assert.Len(t, found1, 3, "应该找到3个环境")

	// 验证所有环境都属于 project-001
	for _, env := range found1 {
		assert.Equal(t, "project-001", env.ProjectID)
	}

	// 查询 project-002
	found2, err := repo.FindByProject(ctx, "project-002")
	require.NoError(t, err)
	assert.Len(t, found2, 2, "应该找到2个环境")

	// 查询不存在的项目
	found3, err := repo.FindByProject(ctx, "project-999")
	require.NoError(t, err)
	assert.Len(t, found3, 0, "应该找到0个环境")
}

// TestEnvironmentRepository_FindByProject_EmptyResult 测试空结果集
func TestEnvironmentRepository_FindByProject_EmptyResult(t *testing.T) {
	client, db, repo := setupEnvironmentTestDB(t)
	if client == nil {
		return
	}
	defer teardownEnvironmentTestDB(t, client, db)

	ctx := context.Background()

	// 不创建任何数据，直接查询
	envs, err := repo.FindByProject(ctx, "non-existent-project")
	require.NoError(t, err)
	assert.NotNil(t, envs)
	assert.Len(t, envs, 0)
}

// TestEnvironmentRepository_VariableTypes 测试变量的各种类型
func TestEnvironmentRepository_VariableTypes(t *testing.T) {
	client, db, repo := setupEnvironmentTestDB(t)
	if client == nil {
		return
	}
	defer teardownEnvironmentTestDB(t, client, db)

	ctx := context.Background()

	env := &models.Environment{
		Name:      "类型测试环境",
		ProjectID: "project-001",
		Variables: map[string]interface{}{
			"string_var": "string value",
			"int_var":    123,
			"float_var":  45.67,
			"bool_var":   true,
			"null_var":   nil,
			"array_var":  []interface{}{1, 2, 3},
			"nested_var": map[string]interface{}{
				"key1": "value1",
				"key2": 456,
			},
		},
	}

	err := repo.Create(ctx, env)
	require.NoError(t, err)

	// 查询并验证
	found, err := repo.FindByID(ctx, env.ID)
	require.NoError(t, err)
	require.NotNil(t, found)

	assert.Contains(t, found.Variables, "string_var")
	assert.Contains(t, found.Variables, "int_var")
	assert.Contains(t, found.Variables, "float_var")
	assert.Contains(t, found.Variables, "bool_var")
	assert.Contains(t, found.Variables, "array_var")
	assert.Contains(t, found.Variables, "nested_var")
}

// TestEnvironmentRepository_MultipleUpdates 测试多次更新
func TestEnvironmentRepository_MultipleUpdates(t *testing.T) {
	client, db, repo := setupEnvironmentTestDB(t)
	if client == nil {
		return
	}
	defer teardownEnvironmentTestDB(t, client, db)

	ctx := context.Background()

	// 创建环境
	env := &models.Environment{
		Name:      "多次更新环境",
		ProjectID: "project-001",
		BaseURL:   "http://v1.example.com",
	}
	err := repo.Create(ctx, env)
	require.NoError(t, err)

	previousUpdatedAt := env.UpdatedAt

	// 进行多次更新
	for i := 0; i < 3; i++ {
		time.Sleep(10 * time.Millisecond)
		env.Name = "更新" + string(rune('A'+i))
		env.BaseURL = "http://v" + string(rune('2'+i)) + ".example.com"

		err := repo.Update(ctx, env)
		require.NoError(t, err)

		assert.True(t, env.UpdatedAt.After(previousUpdatedAt), "UpdatedAt应该递增")
		previousUpdatedAt = env.UpdatedAt
	}

	// 验证最终状态
	found, err := repo.FindByID(ctx, env.ID)
	require.NoError(t, err)
	assert.Equal(t, "更新C", found.Name)
	assert.Equal(t, "http://v4.example.com", found.BaseURL)
}

// TestEnvironmentRepository_ConcurrentCreation 测试并发创建
func TestEnvironmentRepository_ConcurrentCreation(t *testing.T) {
	client, db, repo := setupEnvironmentTestDB(t)
	if client == nil {
		return
	}
	defer teardownEnvironmentTestDB(t, client, db)

	ctx := context.Background()
	concurrency := 10
	done := make(chan bool, concurrency)
	errors := make(chan error, concurrency)

	// 并发创建环境
	for i := 0; i < concurrency; i++ {
		go func(index int) {
			env := &models.Environment{
				Name:      "并发环境" + string(rune('A'+index)),
				ProjectID: "project-001",
			}
			err := repo.Create(ctx, env)
			if err != nil {
				errors <- err
			}
			done <- true
		}(i)
	}

	// 等待所有goroutine完成
	for i := 0; i < concurrency; i++ {
		<-done
	}
	close(errors)

	// 检查是否有错误
	for err := range errors {
		t.Errorf("并发创建失败: %v", err)
	}

	// 验证所有环境都创建成功
	envs, err := repo.FindByProject(ctx, "project-001")
	require.NoError(t, err)
	assert.Len(t, envs, concurrency)
}

// TestEnvironmentRepository_SpecialCharacters 测试特殊字符处理
func TestEnvironmentRepository_SpecialCharacters(t *testing.T) {
	client, db, repo := setupEnvironmentTestDB(t)
	if client == nil {
		return
	}
	defer teardownEnvironmentTestDB(t, client, db)

	ctx := context.Background()

	env := &models.Environment{
		Name:      "特殊字符!@#$%^&*()",
		ProjectID: "project-001",
		BaseURL:   "http://example.com/path?param=value&other=123",
		Variables: map[string]interface{}{
			"special": "包含\n换行\t制表符",
			"json":    `{"key": "value"}`,
		},
	}

	err := repo.Create(ctx, env)
	require.NoError(t, err)

	// 查询并验证
	found, err := repo.FindByID(ctx, env.ID)
	require.NoError(t, err)
	require.NotNil(t, found)

	assert.Contains(t, found.Name, "!@#$%")
	assert.Contains(t, found.BaseURL, "?param=value")
	assert.Contains(t, found.Variables["special"], "\n")
}
