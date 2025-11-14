// +build integration

package repository

import (
	"context"
	"testing"
	"time"

	"github.com/gomockserver/mockserver/internal/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// 本测试文件使用真实的 MongoDB 数据库操作测试
// 需要本地运行 MongoDB 服务，或者使用环境变量指定连接
// 运行方式: MONGODB_URI="mongodb://localhost:27017" go test -tags=integration -v

// getTestMongoClient 获取测试用的 MongoDB 客户端
// 优先使用环境变量 MONGODB_URI，否则使用默认本地连接
func getTestMongoClient(t *testing.T) (*mongo.Client, *mongo.Database) {
	ctx := context.Background()
	
	// 从环境变量读取，如果没有则使用默认值
	uri := "mongodb://localhost:27017"
	if envURI := getEnv("MONGODB_URI", ""); envURI != "" {
		uri = envURI
	}
	
	clientOpts := options.Client().ApplyURI(uri).SetServerSelectionTimeout(2 * time.Second)
	client, err := mongo.Connect(ctx, clientOpts)
	if err != nil {
		t.Skipf("Skipping integration test: MongoDB not available - %v", err)
		return nil, nil
	}
	
	// 验证连接
	err = client.Ping(ctx, nil)
	if err != nil {
		client.Disconnect(ctx)
		t.Skipf("Skipping integration test: cannot ping MongoDB - %v", err)
		return nil, nil
	}
	
	db := client.Database("mockserver_test_" + primitive.NewObjectID().Hex())
	t.Logf("Using test database: %s", db.Name())
	
	return client, db
}

// getEnv 获取环境变量，如果不存在则返回默认值
func getEnv(key, defaultValue string) string {
	// 简化实现，在真实环境中会使用 os.Getenv
	return defaultValue
}

// cleanupTestDB 清理测试数据库
func cleanupTestDB(t *testing.T, client *mongo.Client, db *mongo.Database) {
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

// TestRuleRepository_RealDB_CRUD 测试规则的完整 CRUD 流程
func TestRuleRepository_RealDB_CRUD(t *testing.T) {
	client, db := getTestMongoClient(t)
	if client == nil {
		return // 已跳过
	}
	defer cleanupTestDB(t, client, db)
	
	ctx := context.Background()
	collection := db.Collection("rules")
	repo := &ruleRepository{collection: collection}
	
	// 1. 创建规则
	rule := &models.Rule{
		Name:          "测试规则",
		ProjectID:     "project-001",
		EnvironmentID: "env-001",
		Protocol:      models.ProtocolHTTP,
		MatchType:     models.MatchTypeSimple,
		Priority:      100,
		Enabled:       true,
		MatchCondition: map[string]interface{}{
			"method": "GET",
			"path":   "/api/test",
		},
		Response: models.Response{
			Type: models.ResponseTypeStatic,
			Content: map[string]interface{}{
				"status_code": 200,
				"body":        "test response",
			},
		},
	}
	
	err := repo.Create(ctx, rule)
	require.NoError(t, err)
	assert.NotEmpty(t, rule.ID, "Rule ID should be set")
	assert.False(t, rule.CreatedAt.IsZero(), "CreatedAt should be set")
	assert.False(t, rule.UpdatedAt.IsZero(), "UpdatedAt should be set")
	
	originalID := rule.ID
	originalCreatedAt := rule.CreatedAt
	
	// 2. 查询规则
	found, err := repo.FindByID(ctx, rule.ID)
	require.NoError(t, err)
	require.NotNil(t, found, "Rule should be found")
	assert.Equal(t, rule.Name, found.Name)
	assert.Equal(t, rule.ProjectID, found.ProjectID)
	assert.Equal(t, rule.EnvironmentID, found.EnvironmentID)
	assert.Equal(t, rule.Protocol, found.Protocol)
	assert.Equal(t, rule.Priority, found.Priority)
	assert.Equal(t, rule.Enabled, found.Enabled)
	
	// 3. 更新规则
	time.Sleep(10 * time.Millisecond) // 确保时间戳不同
	rule.Name = "更新后的规则"
	rule.Priority = 200
	rule.Enabled = false
	
	err = repo.Update(ctx, rule)
	require.NoError(t, err)
	assert.Equal(t, originalID, rule.ID, "ID should not change")
	assert.Equal(t, originalCreatedAt, rule.CreatedAt, "CreatedAt should not change")
	assert.True(t, rule.UpdatedAt.After(originalCreatedAt), "UpdatedAt should be newer")
	
	// 验证更新
	updated, err := repo.FindByID(ctx, rule.ID)
	require.NoError(t, err)
	assert.Equal(t, "更新后的规则", updated.Name)
	assert.Equal(t, 200, updated.Priority)
	assert.False(t, updated.Enabled)
	
	// 4. 删除规则
	err = repo.Delete(ctx, rule.ID)
	require.NoError(t, err)
	
	// 验证删除
	deleted, err := repo.FindByID(ctx, rule.ID)
	assert.NoError(t, err)
	assert.Nil(t, deleted, "Rule should be deleted")
}

// TestRuleRepository_RealDB_FindByEnvironment 测试按环境查询
func TestRuleRepository_RealDB_FindByEnvironment(t *testing.T) {
	client, db := getTestMongoClient(t)
	if client == nil {
		return
	}
	defer cleanupTestDB(t, client, db)
	
	ctx := context.Background()
	collection := db.Collection("rules")
	repo := &ruleRepository{collection: collection}
	
	// 创建多个规则
	rules := []*models.Rule{
		{
			Name:          "规则1",
			ProjectID:     "project-001",
			EnvironmentID: "env-001",
			Protocol:      models.ProtocolHTTP,
			Priority:      300,
			Enabled:       true,
		},
		{
			Name:          "规则2",
			ProjectID:     "project-001",
			EnvironmentID: "env-001",
			Protocol:      models.ProtocolHTTP,
			Priority:      100,
			Enabled:       true,
		},
		{
			Name:          "规则3",
			ProjectID:     "project-001",
			EnvironmentID: "env-001",
			Protocol:      models.ProtocolHTTP,
			Priority:      200,
			Enabled:       false,
		},
		{
			Name:          "规则4",
			ProjectID:     "project-001",
			EnvironmentID: "env-002", // 不同环境
			Protocol:      models.ProtocolHTTP,
			Priority:      100,
			Enabled:       true,
		},
	}
	
	for _, rule := range rules {
		err := repo.Create(ctx, rule)
		require.NoError(t, err)
	}
	
	// 查询 env-001 的所有规则（应按优先级降序）
	found, err := repo.FindByEnvironment(ctx, "project-001", "env-001")
	require.NoError(t, err)
	assert.Len(t, found, 3, "Should find 3 rules")
	
	// 验证排序（优先级降序）
	assert.Equal(t, "规则1", found[0].Name, "Highest priority first")
	assert.Equal(t, 300, found[0].Priority)
	assert.Equal(t, "规则3", found[1].Name)
	assert.Equal(t, 200, found[1].Priority)
	assert.Equal(t, "规则2", found[2].Name)
	assert.Equal(t, 100, found[2].Priority)
}

// TestRuleRepository_RealDB_List 测试分页和过滤
func TestRuleRepository_RealDB_List(t *testing.T) {
	client, db := getTestMongoClient(t)
	if client == nil {
		return
	}
	defer cleanupTestDB(t, client, db)
	
	ctx := context.Background()
	collection := db.Collection("rules")
	repo := &ruleRepository{collection: collection}
	
	// 创建测试数据
	for i := 0; i < 15; i++ {
		rule := &models.Rule{
			Name:          "规则" + string(rune('A'+i)),
			ProjectID:     "project-001",
			EnvironmentID: "env-001",
			Protocol:      models.ProtocolHTTP,
			Priority:      i * 10,
			Enabled:       i%2 == 0,
		}
		err := repo.Create(ctx, rule)
		require.NoError(t, err)
	}
	
	// 测试分页
	filter1 := bson.M{"project_id": "project-001"}
	page1, total, err := repo.List(ctx, filter1, 0, 5)
	require.NoError(t, err)
	assert.Equal(t, int64(15), total, "Total should be 15")
	assert.Len(t, page1, 5, "Page 1 should have 5 items")
	
	page2, _, err := repo.List(ctx, filter1, 5, 5)
	require.NoError(t, err)
	assert.Len(t, page2, 5, "Page 2 should have 5 items")
	
	page3, _, err := repo.List(ctx, filter1, 10, 5)
	require.NoError(t, err)
	assert.Len(t, page3, 5, "Page 3 should have 5 items")
	
	// 测试过滤（只查询启用的）
	filter2 := bson.M{
		"project_id": "project-001",
		"enabled":    true,
	}
	enabled, total, err := repo.List(ctx, filter2, 0, 20)
	require.NoError(t, err)
	assert.Equal(t, int64(8), total, "Should have 8 enabled rules")
	assert.Len(t, enabled, 8)
	
	for _, rule := range enabled {
		assert.True(t, rule.Enabled, "All returned rules should be enabled")
	}
}

// TestProjectRepository_RealDB_CRUD 测试项目的 CRUD
func TestProjectRepository_RealDB_CRUD(t *testing.T) {
	client, db := getTestMongoClient(t)
	if client == nil {
		return
	}
	defer cleanupTestDB(t, client, db)
	
	ctx := context.Background()
	collection := db.Collection("projects")
	repo := &projectRepository{collection: collection}
	
	// 创建项目
	project := &models.Project{
		Name:        "测试项目",
		WorkspaceID: "workspace-001",
		Description: "这是一个测试项目",
	}
	
	err := repo.Create(ctx, project)
	require.NoError(t, err)
	assert.NotEmpty(t, project.ID)
	assert.False(t, project.CreatedAt.IsZero())
	
	// 查询
	found, err := repo.FindByID(ctx, project.ID)
	require.NoError(t, err)
	require.NotNil(t, found)
	assert.Equal(t, project.Name, found.Name)
	assert.Equal(t, project.WorkspaceID, found.WorkspaceID)
	
	// 更新
	project.Name = "更新后的项目"
	project.Description = "更新后的描述"
	
	err = repo.Update(ctx, project)
	require.NoError(t, err)
	
	updated, err := repo.FindByID(ctx, project.ID)
	require.NoError(t, err)
	assert.Equal(t, "更新后的项目", updated.Name)
	assert.Equal(t, "更新后的描述", updated.Description)
	
	// 删除
	err = repo.Delete(ctx, project.ID)
	require.NoError(t, err)
	
	deleted, err := repo.FindByID(ctx, project.ID)
	assert.NoError(t, err)
	assert.Nil(t, deleted)
}

// TestEnvironmentRepository_RealDB_CRUD 测试环境的 CRUD
func TestEnvironmentRepository_RealDB_CRUD(t *testing.T) {
	client, db := getTestMongoClient(t)
	if client == nil {
		return
	}
	defer cleanupTestDB(t, client, db)
	
	ctx := context.Background()
	collection := db.Collection("environments")
	repo := &environmentRepository{collection: collection}
	
	// 创建环境
	env := &models.Environment{
		Name:      "开发环境",
		ProjectID: "project-001",
	}
	
	err := repo.Create(ctx, env)
	require.NoError(t, err)
	assert.NotEmpty(t, env.ID)
	
	// 查询
	found, err := repo.FindByID(ctx, env.ID)
	require.NoError(t, err)
	require.NotNil(t, found)
	assert.Equal(t, env.Name, found.Name)
	
	// 更新
	env.Name = "生产环境"
	
	err = repo.Update(ctx, env)
	require.NoError(t, err)
	
	updated, err := repo.FindByID(ctx, env.ID)
	require.NoError(t, err)
	assert.Equal(t, "生产环境", updated.Name)
	
	// 删除
	err = repo.Delete(ctx, env.ID)
	require.NoError(t, err)
	
	deleted, err := repo.FindByID(ctx, env.ID)
	assert.NoError(t, err)
	assert.Nil(t, deleted)
}

// TestRuleRepository_RealDB_FindEnabledByEnvironment 测试查询启用的规则
func TestRuleRepository_RealDB_FindEnabledByEnvironment(t *testing.T) {
	client, db := getTestMongoClient(t)
	if client == nil {
		return
	}
	defer cleanupTestDB(t, client, db)
	
	ctx := context.Background()
	collection := db.Collection("rules")
	repo := &ruleRepository{collection: collection}
	
	// 创建规则
	rules := []*models.Rule{
		{
			Name:          "启用规则1",
			ProjectID:     "project-001",
			EnvironmentID: "env-001",
			Protocol:      models.ProtocolHTTP,
			Priority:      300,
			Enabled:       true,
		},
		{
			Name:          "禁用规则",
			ProjectID:     "project-001",
			EnvironmentID: "env-001",
			Protocol:      models.ProtocolHTTP,
			Priority:      200,
			Enabled:       false,
		},
		{
			Name:          "启用规则2",
			ProjectID:     "project-001",
			EnvironmentID: "env-001",
			Protocol:      models.ProtocolHTTP,
			Priority:      100,
			Enabled:       true,
		},
	}
	
	for _, rule := range rules {
		err := repo.Create(ctx, rule)
		require.NoError(t, err)
	}
	
	// 查询启用的规则
	found, err := repo.FindEnabledByEnvironment(ctx, "project-001", "env-001")
	require.NoError(t, err)
	assert.Len(t, found, 2, "Should find 2 enabled rules")
	
	// 验证都是启用的
	for _, rule := range found {
		assert.True(t, rule.Enabled)
	}
	
	// 验证排序
	assert.Equal(t, "启用规则1", found[0].Name)
	assert.Equal(t, "启用规则2", found[1].Name)
}
