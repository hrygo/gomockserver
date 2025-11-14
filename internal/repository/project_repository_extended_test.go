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

// setupProjectTestDB 设置项目测试数据库
func setupProjectTestDB(t *testing.T) (*mongo.Client, *mongo.Database, ProjectRepository) {
	ctx := context.Background()

	// 从环境变量读取，如果没有则使用默认值
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

	// 验证连接
	err = client.Ping(ctx, nil)
	if err != nil {
		client.Disconnect(ctx)
		t.Skipf("Skipping integration test: cannot ping MongoDB - %v", err)
		return nil, nil, nil
	}

	db := client.Database("mockserver_test_project_" + primitive.NewObjectID().Hex())
	t.Logf("Using test database: %s", db.Name())

	collection := db.Collection("projects")
	repo := &projectRepository{collection: collection}

	return client, db, repo
}

// teardownProjectTestDB 清理项目测试数据库
func teardownProjectTestDB(t *testing.T, client *mongo.Client, db *mongo.Database) {
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

// TestProjectRepository_Create_Success 测试成功创建项目
func TestProjectRepository_Create_Success(t *testing.T) {
	client, db, repo := setupProjectTestDB(t)
	if client == nil {
		return
	}
	defer teardownProjectTestDB(t, client, db)

	ctx := context.Background()

	tests := []struct {
		name    string
		project *models.Project
	}{
		{
			name: "完整字段项目",
			project: &models.Project{
				Name:        "完整项目",
				WorkspaceID: "workspace-001",
				Description: "这是一个包含所有字段的项目",
			},
		},
		{
			name: "最小字段项目",
			project: &models.Project{
				Name:        "最小项目",
				WorkspaceID: "workspace-002",
			},
		},
		{
			name: "中文项目名",
			project: &models.Project{
				Name:        "测试项目中文名称",
				WorkspaceID: "workspace-003",
				Description: "包含中文描述的项目",
			},
		},
		{
			name: "长描述项目",
			project: &models.Project{
				Name:        "长描述项目",
				WorkspaceID: "workspace-004",
				Description: "这是一个非常长的描述文本，用于测试数据库是否能够正确存储较长的描述内容。" +
					"描述可以包含多行文本，各种特殊字符，以及详细的项目说明信息。",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := repo.Create(ctx, tt.project)
			require.NoError(t, err, "创建项目应该成功")

			assert.NotEmpty(t, tt.project.ID, "ID应该被设置")
			assert.False(t, tt.project.CreatedAt.IsZero(), "CreatedAt应该被设置")
			assert.False(t, tt.project.UpdatedAt.IsZero(), "UpdatedAt应该被设置")
			assert.Equal(t, tt.project.CreatedAt, tt.project.UpdatedAt, "初始创建时两个时间戳应该相同")

			// 验证ID格式
			_, err = primitive.ObjectIDFromHex(tt.project.ID)
			assert.NoError(t, err, "ID应该是有效的ObjectID格式")
		})
	}
}

// TestProjectRepository_FindByID_EdgeCases 测试查询项目的边界情况
func TestProjectRepository_FindByID_EdgeCases(t *testing.T) {
	client, db, repo := setupProjectTestDB(t)
	if client == nil {
		return
	}
	defer teardownProjectTestDB(t, client, db)

	ctx := context.Background()

	// 首先创建一个测试项目
	project := &models.Project{
		Name:        "测试项目",
		WorkspaceID: "workspace-001",
	}
	err := repo.Create(ctx, project)
	require.NoError(t, err)

	tests := []struct {
		name        string
		id          string
		shouldExist bool
		shouldError bool
	}{
		{
			name:        "有效ID查询存在的项目",
			id:          project.ID,
			shouldExist: true,
			shouldError: false,
		},
		{
			name:        "有效ID查询不存在的项目",
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
			name:        "长度不够的ID",
			id:          "12345",
			shouldExist: false,
			shouldError: true,
		},
		{
			name:        "超长ID",
			id:          "507f1f77bcf86cd799439011507f1f77bcf86cd799439011",
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
				require.NotNil(t, found, "应该找到项目")
				assert.Equal(t, tt.id, found.ID)
			} else {
				assert.Nil(t, found, "不应该找到项目")
			}
		})
	}
}

// TestProjectRepository_Update_Success 测试更新项目成功场景
func TestProjectRepository_Update_Success(t *testing.T) {
	client, db, repo := setupProjectTestDB(t)
	if client == nil {
		return
	}
	defer teardownProjectTestDB(t, client, db)

	ctx := context.Background()

	// 创建初始项目
	project := &models.Project{
		Name:        "初始项目",
		WorkspaceID: "workspace-001",
		Description: "初始描述",
	}
	err := repo.Create(ctx, project)
	require.NoError(t, err)

	originalID := project.ID
	originalCreatedAt := project.CreatedAt
	time.Sleep(10 * time.Millisecond) // 确保时间戳不同

	tests := []struct {
		name       string
		updateFunc func(*models.Project)
		verifyFunc func(*testing.T, *models.Project)
	}{
		{
			name: "更新所有字段",
			updateFunc: func(p *models.Project) {
				p.Name = "更新后的名称"
				p.Description = "更新后的描述"
				p.WorkspaceID = "workspace-002"
			},
			verifyFunc: func(t *testing.T, p *models.Project) {
				assert.Equal(t, "更新后的名称", p.Name)
				assert.Equal(t, "更新后的描述", p.Description)
				assert.Equal(t, "workspace-002", p.WorkspaceID)
			},
		},
		{
			name: "只更新名称",
			updateFunc: func(p *models.Project) {
				p.Name = "仅更新名称"
			},
			verifyFunc: func(t *testing.T, p *models.Project) {
				assert.Equal(t, "仅更新名称", p.Name)
			},
		},
		{
			name: "清空描述",
			updateFunc: func(p *models.Project) {
				p.Description = ""
			},
			verifyFunc: func(t *testing.T, p *models.Project) {
				assert.Empty(t, p.Description)
			},
		},
		{
			name: "使用特殊字符",
			updateFunc: func(p *models.Project) {
				p.Name = "特殊字符!@#$%^&*()"
				p.Description = "包含换行\n和制表符\t的描述"
			},
			verifyFunc: func(t *testing.T, p *models.Project) {
				assert.Contains(t, p.Name, "!@#$%")
				assert.Contains(t, p.Description, "\n")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// 应用更新
			tt.updateFunc(project)

			err := repo.Update(ctx, project)
			require.NoError(t, err, "更新应该成功")

			// 验证ID和CreatedAt没有变化
			assert.Equal(t, originalID, project.ID, "ID不应该改变")
			assert.Equal(t, originalCreatedAt, project.CreatedAt, "CreatedAt不应该改变")
			assert.True(t, project.UpdatedAt.After(originalCreatedAt), "UpdatedAt应该更新")

			// 从数据库重新查询验证
			updated, err := repo.FindByID(ctx, project.ID)
			require.NoError(t, err)
			require.NotNil(t, updated)

			tt.verifyFunc(t, updated)
		})
	}
}

// TestProjectRepository_Update_InvalidID 测试使用无效ID更新
func TestProjectRepository_Update_InvalidID(t *testing.T) {
	client, db, repo := setupProjectTestDB(t)
	if client == nil {
		return
	}
	defer teardownProjectTestDB(t, client, db)

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
			project := &models.Project{
				ID:          tt.id,
				Name:        "测试",
				WorkspaceID: "workspace-001",
			}

			err := repo.Update(ctx, project)
			assert.Error(t, err, "应该返回错误")
		})
	}
}

// TestProjectRepository_Delete_Success 测试删除项目
func TestProjectRepository_Delete_Success(t *testing.T) {
	client, db, repo := setupProjectTestDB(t)
	if client == nil {
		return
	}
	defer teardownProjectTestDB(t, client, db)

	ctx := context.Background()

	// 创建项目
	project := &models.Project{
		Name:        "待删除项目",
		WorkspaceID: "workspace-001",
	}
	err := repo.Create(ctx, project)
	require.NoError(t, err)

	// 验证项目存在
	found, err := repo.FindByID(ctx, project.ID)
	require.NoError(t, err)
	require.NotNil(t, found)

	// 删除项目
	err = repo.Delete(ctx, project.ID)
	assert.NoError(t, err, "删除应该成功")

	// 验证项目已删除
	deleted, err := repo.FindByID(ctx, project.ID)
	assert.NoError(t, err)
	assert.Nil(t, deleted, "项目应该已被删除")
}

// TestProjectRepository_Delete_NotExist 测试删除不存在的项目
func TestProjectRepository_Delete_NotExist(t *testing.T) {
	client, db, repo := setupProjectTestDB(t)
	if client == nil {
		return
	}
	defer teardownProjectTestDB(t, client, db)

	ctx := context.Background()

	// 删除不存在的项目（应该不报错）
	err := repo.Delete(ctx, "507f1f77bcf86cd799439011")
	assert.NoError(t, err, "删除不存在的项目应该不报错")
}

// TestProjectRepository_Delete_InvalidID 测试使用无效ID删除
func TestProjectRepository_Delete_InvalidID(t *testing.T) {
	client, db, repo := setupProjectTestDB(t)
	if client == nil {
		return
	}
	defer teardownProjectTestDB(t, client, db)

	ctx := context.Background()

	tests := []struct {
		name string
		id   string
	}{
		{"无效ID", "invalid-id"},
		{"空ID", ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := repo.Delete(ctx, tt.id)
			assert.Error(t, err, "应该返回错误")
		})
	}
}

// TestProjectRepository_FindByWorkspace 测试按工作空间查询
func TestProjectRepository_FindByWorkspace(t *testing.T) {
	client, db, repo := setupProjectTestDB(t)
	if client == nil {
		return
	}
	defer teardownProjectTestDB(t, client, db)

	ctx := context.Background()

	// 创建多个工作空间的项目
	projects := []*models.Project{
		{Name: "项目1", WorkspaceID: "workspace-001"},
		{Name: "项目2", WorkspaceID: "workspace-001"},
		{Name: "项目3", WorkspaceID: "workspace-001"},
		{Name: "项目4", WorkspaceID: "workspace-002"},
		{Name: "项目5", WorkspaceID: "workspace-002"},
	}

	for _, p := range projects {
		err := repo.Create(ctx, p)
		require.NoError(t, err)
	}

	// 查询 workspace-001
	found1, err := repo.FindByWorkspace(ctx, "workspace-001")
	require.NoError(t, err)
	assert.Len(t, found1, 3, "应该找到3个项目")

	// 查询 workspace-002
	found2, err := repo.FindByWorkspace(ctx, "workspace-002")
	require.NoError(t, err)
	assert.Len(t, found2, 2, "应该找到2个项目")

	// 查询不存在的workspace
	found3, err := repo.FindByWorkspace(ctx, "workspace-999")
	require.NoError(t, err)
	assert.Len(t, found3, 0, "应该找到0个项目")
}

// TestProjectRepository_List_Pagination 测试分页查询
func TestProjectRepository_List_Pagination(t *testing.T) {
	client, db, repo := setupProjectTestDB(t)
	if client == nil {
		return
	}
	defer teardownProjectTestDB(t, client, db)

	ctx := context.Background()

	// 创建20个项目
	for i := 0; i < 20; i++ {
		project := &models.Project{
			Name:        "项目" + string(rune('A'+i)),
			WorkspaceID: "workspace-001",
		}
		err := repo.Create(ctx, project)
		require.NoError(t, err)
		time.Sleep(1 * time.Millisecond) // 确保创建时间不同
	}

	tests := []struct {
		name          string
		skip          int64
		limit         int64
		expectedCount int
		expectedTotal int64
	}{
		{"第一页-每页10条", 0, 10, 10, 20},
		{"第二页-每页10条", 10, 10, 10, 20},
		{"第三页-每页10条", 20, 10, 0, 20}, // 已经没有数据了
		{"每页5条", 0, 5, 5, 20},
		{"每页20条", 0, 20, 20, 20},
		{"每页25条(超过总数)", 0, 25, 20, 20},
		{"跳过15条-每页10条", 15, 10, 5, 20},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			projects, total, err := repo.List(ctx, tt.skip, tt.limit)
			require.NoError(t, err)
			assert.Equal(t, tt.expectedTotal, total, "总数应该匹配")
			assert.Len(t, projects, tt.expectedCount, "返回的项目数应该匹配")
		})
	}
}

// TestProjectRepository_List_Ordering 测试排序
func TestProjectRepository_List_Ordering(t *testing.T) {
	client, db, repo := setupProjectTestDB(t)
	if client == nil {
		return
	}
	defer teardownProjectTestDB(t, client, db)

	ctx := context.Background()

	// 创建项目（间隔时间确保顺序）
	var createdProjects []*models.Project
	for i := 0; i < 5; i++ {
		project := &models.Project{
			Name:        "项目" + string(rune('A'+i)),
			WorkspaceID: "workspace-001",
		}
		err := repo.Create(ctx, project)
		require.NoError(t, err)
		createdProjects = append(createdProjects, project)
		time.Sleep(2 * time.Millisecond) // 确保创建时间不同
	}

	// 查询（默认按创建时间降序）
	projects, _, err := repo.List(ctx, 0, 10)
	require.NoError(t, err)
	require.Len(t, projects, 5)

	// 验证是按创建时间降序（最新的在前面）
	for i := 0; i < len(projects)-1; i++ {
		assert.True(t,
			projects[i].CreatedAt.After(projects[i+1].CreatedAt) ||
				projects[i].CreatedAt.Equal(projects[i+1].CreatedAt),
			"应该按创建时间降序排列")
	}
}

// TestProjectRepository_List_EmptyResult 测试空结果集
func TestProjectRepository_List_EmptyResult(t *testing.T) {
	client, db, repo := setupProjectTestDB(t)
	if client == nil {
		return
	}
	defer teardownProjectTestDB(t, client, db)

	ctx := context.Background()

	// 不创建任何数据，直接查询
	projects, total, err := repo.List(ctx, 0, 10)
	require.NoError(t, err)
	assert.Equal(t, int64(0), total, "总数应该为0")
	assert.Len(t, projects, 0, "应该返回空数组")
	assert.NotNil(t, projects, "应该返回空数组而不是nil")
}

// TestProjectRepository_ConcurrentCreation 测试并发创建
func TestProjectRepository_ConcurrentCreation(t *testing.T) {
	client, db, repo := setupProjectTestDB(t)
	if client == nil {
		return
	}
	defer teardownProjectTestDB(t, client, db)

	ctx := context.Background()
	concurrency := 10
	done := make(chan bool, concurrency)
	errors := make(chan error, concurrency)

	// 并发创建项目
	for i := 0; i < concurrency; i++ {
		go func(index int) {
			project := &models.Project{
				Name:        "并发项目" + string(rune('A'+index)),
				WorkspaceID: "workspace-001",
			}
			err := repo.Create(ctx, project)
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

	// 验证所有项目都创建成功
	projects, total, err := repo.List(ctx, 0, 20)
	require.NoError(t, err)
	assert.Equal(t, int64(concurrency), total, "应该创建了所有项目")
	assert.Len(t, projects, concurrency)
}
