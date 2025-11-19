package repository

import (
	"context"
	"testing"
	"time"

	"github.com/gomockserver/mockserver/internal/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func setupTestDB(t *testing.T) (*mongo.Database, func()) {
	ctx := context.Background()

	// 连接测试数据库
	client, err := mongo.Connect(ctx, options.Client().ApplyURI("mongodb://localhost:27017"))
	require.NoError(t, err)

	// 使用测试数据库
	db := client.Database("mockserver_test_logs")

	// 清理函数
	cleanup := func() {
		_ = db.Drop(ctx)
		_ = client.Disconnect(ctx)
	}

	return db, cleanup
}

func TestRequestLogRepository_Create(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	repo := NewMongoRequestLogRepository(db)
	ctx := context.Background()

	log := &models.RequestLog{
		RequestID:     "test-req-001",
		ProjectID:     "proj-001",
		EnvironmentID: "env-001",
		Protocol:      models.ProtocolHTTP,
		Method:        "GET",
		Path:          "/api/test",
		Request:       map[string]interface{}{"query": "test"},
		Response:      map[string]interface{}{"status": "ok"},
		StatusCode:    200,
		Duration:      150,
		SourceIP:      "127.0.0.1",
	}

	err := repo.Create(ctx, log)
	assert.NoError(t, err)
	assert.NotEmpty(t, log.ID)
	assert.False(t, log.Timestamp.IsZero())
}

func TestRequestLogRepository_FindByID(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	repo := NewMongoRequestLogRepository(db)
	ctx := context.Background()

	// 创建测试日志
	log := &models.RequestLog{
		RequestID:     "test-req-002",
		ProjectID:     "proj-001",
		EnvironmentID: "env-001",
		Protocol:      models.ProtocolHTTP,
		Method:        "POST",
		Path:          "/api/test",
		StatusCode:    201,
		Duration:      200,
		SourceIP:      "127.0.0.1",
	}

	err := repo.Create(ctx, log)
	require.NoError(t, err)

	// 查询日志
	found, err := repo.FindByID(ctx, log.ID)
	assert.NoError(t, err)
	assert.NotNil(t, found)
	assert.Equal(t, log.RequestID, found.RequestID)
	assert.Equal(t, log.Method, found.Method)
	assert.Equal(t, log.StatusCode, found.StatusCode)

	// 查询不存在的日志
	notFound, err := repo.FindByID(ctx, "non-existent-id")
	assert.NoError(t, err)
	assert.Nil(t, notFound)
}

func TestRequestLogRepository_List(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	repo := NewMongoRequestLogRepository(db)
	ctx := context.Background()

	// 创建多条测试日志
	logs := []*models.RequestLog{
		{
			RequestID:     "req-001",
			ProjectID:     "proj-001",
			EnvironmentID: "env-001",
			Protocol:      models.ProtocolHTTP,
			Method:        "GET",
			Path:          "/api/users",
			StatusCode:    200,
			Duration:      100,
			SourceIP:      "192.168.1.1",
		},
		{
			RequestID:     "req-002",
			ProjectID:     "proj-001",
			EnvironmentID: "env-001",
			Protocol:      models.ProtocolHTTP,
			Method:        "POST",
			Path:          "/api/users",
			StatusCode:    201,
			Duration:      150,
			SourceIP:      "192.168.1.2",
		},
		{
			RequestID:     "req-003",
			ProjectID:     "proj-002",
			EnvironmentID: "env-002",
			Protocol:      models.ProtocolWebSocket,
			Path:          "/ws/chat",
			StatusCode:    101,
			Duration:      50,
			SourceIP:      "192.168.1.3",
		},
	}

	for _, log := range logs {
		err := repo.Create(ctx, log)
		require.NoError(t, err)
	}

	t.Run("List all logs", func(t *testing.T) {
		result, total, err := repo.List(ctx, RequestLogFilter{
			Page:     1,
			PageSize: 10,
		})
		assert.NoError(t, err)
		assert.Equal(t, int64(3), total)
		assert.Len(t, result, 3)
	})

	t.Run("Filter by project", func(t *testing.T) {
		result, total, err := repo.List(ctx, RequestLogFilter{
			ProjectID: "proj-001",
			Page:      1,
			PageSize:  10,
		})
		assert.NoError(t, err)
		assert.Equal(t, int64(2), total)
		assert.Len(t, result, 2)
	})

	t.Run("Filter by protocol", func(t *testing.T) {
		result, total, err := repo.List(ctx, RequestLogFilter{
			Protocol: models.ProtocolWebSocket,
			Page:     1,
			PageSize: 10,
		})
		assert.NoError(t, err)
		assert.Equal(t, int64(1), total)
		assert.Len(t, result, 1)
		assert.Equal(t, "/ws/chat", result[0].Path)
	})

	t.Run("Filter by method", func(t *testing.T) {
		result, total, err := repo.List(ctx, RequestLogFilter{
			Method:   "POST",
			Page:     1,
			PageSize: 10,
		})
		assert.NoError(t, err)
		assert.Equal(t, int64(1), total)
		assert.Len(t, result, 1)
		assert.Equal(t, "req-002", result[0].RequestID)
	})

	t.Run("Pagination", func(t *testing.T) {
		result, total, err := repo.List(ctx, RequestLogFilter{
			Page:     1,
			PageSize: 2,
		})
		assert.NoError(t, err)
		assert.Equal(t, int64(3), total)
		assert.Len(t, result, 2)

		result, total, err = repo.List(ctx, RequestLogFilter{
			Page:     2,
			PageSize: 2,
		})
		assert.NoError(t, err)
		assert.Equal(t, int64(3), total)
		assert.Len(t, result, 1)
	})
}

func TestRequestLogRepository_DeleteBefore(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	repo := NewMongoRequestLogRepository(db)
	ctx := context.Background()

	// 创建不同时间的日志
	now := time.Now()
	logs := []*models.RequestLog{
		{
			RequestID: "old-1",
			ProjectID: "proj-001",
			Timestamp: now.AddDate(0, 0, -10), // 10天前
			Duration:  100,
			SourceIP:  "127.0.0.1",
		},
		{
			RequestID: "old-2",
			ProjectID: "proj-001",
			Timestamp: now.AddDate(0, 0, -8), // 8天前
			Duration:  100,
			SourceIP:  "127.0.0.1",
		},
		{
			RequestID: "recent",
			ProjectID: "proj-001",
			Timestamp: now.AddDate(0, 0, -2), // 2天前
			Duration:  100,
			SourceIP:  "127.0.0.1",
		},
	}

	for _, log := range logs {
		err := repo.Create(ctx, log)
		require.NoError(t, err)
	}

	// 删除7天前的日志
	deletedCount, err := repo.DeleteBefore(ctx, now.AddDate(0, 0, -7))
	assert.NoError(t, err)
	assert.Equal(t, int64(2), deletedCount)

	// 验证剩余日志
	result, total, err := repo.List(ctx, RequestLogFilter{
		Page:     1,
		PageSize: 10,
	})
	assert.NoError(t, err)
	assert.Equal(t, int64(1), total)
	assert.Equal(t, "recent", result[0].RequestID)
}

func TestRequestLogRepository_GetStatistics(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	repo := NewMongoRequestLogRepository(db)
	ctx := context.Background()

	// 创建测试数据
	now := time.Now()
	logs := []*models.RequestLog{
		{
			RequestID:     "stat-1",
			ProjectID:     "proj-001",
			EnvironmentID: "env-001",
			Protocol:      models.ProtocolHTTP,
			StatusCode:    200,
			Duration:      100,
			SourceIP:      "127.0.0.1",
			Timestamp:     now.Add(-1 * time.Hour),
		},
		{
			RequestID:     "stat-2",
			ProjectID:     "proj-001",
			EnvironmentID: "env-001",
			Protocol:      models.ProtocolHTTP,
			StatusCode:    200,
			Duration:      200,
			SourceIP:      "127.0.0.1",
			Timestamp:     now.Add(-30 * time.Minute),
		},
		{
			RequestID:     "stat-3",
			ProjectID:     "proj-001",
			EnvironmentID: "env-001",
			Protocol:      models.ProtocolHTTP,
			StatusCode:    500,
			Duration:      300,
			SourceIP:      "127.0.0.1",
			Timestamp:     now.Add(-10 * time.Minute),
		},
	}

	for _, log := range logs {
		err := repo.Create(ctx, log)
		require.NoError(t, err)
	}

	// 获取统计
	stats, err := repo.GetStatistics(ctx, "proj-001", "env-001", now.Add(-2*time.Hour), now)
	assert.NoError(t, err)
	assert.NotNil(t, stats)
	assert.Equal(t, int64(3), stats.TotalRequests)
	assert.Equal(t, int64(2), stats.SuccessRequests)
	assert.Equal(t, int64(1), stats.ErrorRequests)
	assert.Equal(t, float64(200), stats.AvgDuration) // (100+200+300)/3
	assert.Equal(t, int64(300), stats.MaxDuration)
	assert.Equal(t, int64(100), stats.MinDuration)
}
