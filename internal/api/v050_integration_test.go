package api

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gomockserver/mockserver/internal/models"
	"github.com/gomockserver/mockserver/internal/repository"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// 集成测试 - 测试完整的API流程
func TestRequestLogAPIIntegration(t *testing.T) {
	gin.SetMode(gin.TestMode)
	
	// 连接测试数据库
	ctx := context.Background()
	client, err := mongo.Connect(ctx, options.Client().ApplyURI("mongodb://localhost:27017"))
	require.NoError(t, err)
	defer client.Disconnect(ctx)
	
	db := client.Database("mockserver_api_integration_test")
	defer db.Drop(ctx)
	
	// 创建路由和处理器
	router := gin.New()
	requestLogRepo := repository.NewMongoRequestLogRepository(db)
	requestLogHandler := NewRequestLogHandler(requestLogRepo)
	
	api := router.Group("/api/v1")
	requestLogHandler.RegisterRoutes(api)
	
	// 创建测试数据
	repo := repository.NewMongoRequestLogRepository(db)
	testLog := &models.RequestLog{
		RequestID:     "integration-test-001",
		ProjectID:     "proj-001",
		EnvironmentID: "env-001",
		Protocol:      models.ProtocolHTTP,
		Method:        "GET",
		Path:          "/api/test",
		StatusCode:    200,
		Duration:      100,
		SourceIP:      "127.0.0.1",
	}
	err = repo.Create(ctx, testLog)
	require.NoError(t, err)
	
	t.Run("List request logs", func(t *testing.T) {
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/api/v1/request-logs", nil)
		router.ServeHTTP(w, req)
		
		assert.Equal(t, http.StatusOK, w.Code)
	})
	
	t.Run("Get request log by ID", func(t *testing.T) {
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/api/v1/request-logs/"+testLog.ID, nil)
		router.ServeHTTP(w, req)
		
		assert.Equal(t, http.StatusOK, w.Code)
	})
	
	t.Run("Get statistics", func(t *testing.T) {
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/api/v1/request-logs/statistics?period=24h", nil)
		router.ServeHTTP(w, req)
		
		assert.Equal(t, http.StatusOK, w.Code)
	})
}

func TestHealthAPIIntegration(t *testing.T) {
	gin.SetMode(gin.TestMode)
	
	// 连接测试数据库
	ctx := context.Background()
	client, err := mongo.Connect(ctx, options.Client().ApplyURI("mongodb://localhost:27017"))
	require.NoError(t, err)
	defer client.Disconnect(ctx)
	
	db := client.Database("mockserver_health_integration_test")
	defer db.Drop(ctx)
	
	router := gin.New()
	handler := NewHealthHandler(db, nil)
	
	api := router.Group("/api/v1")
	{
		api.GET("/health", handler.Health)
		api.GET("/metrics", handler.Metrics)
		api.GET("/live", handler.Live)
		api.GET("/ready", handler.Ready)
	}
	
	t.Run("Health check", func(t *testing.T) {
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/api/v1/health", nil)
		router.ServeHTTP(w, req)
		
		assert.Equal(t, http.StatusOK, w.Code)
	})
	
	t.Run("System metrics", func(t *testing.T) {
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/api/v1/metrics", nil)
		router.ServeHTTP(w, req)
		
		assert.Equal(t, http.StatusOK, w.Code)
	})
	
	t.Run("Liveness probe", func(t *testing.T) {
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/api/v1/live", nil)
		router.ServeHTTP(w, req)
		
		assert.Equal(t, http.StatusOK, w.Code)
	})
	
	t.Run("Readiness probe", func(t *testing.T) {
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/api/v1/ready", nil)
		router.ServeHTTP(w, req)
		
		assert.Equal(t, http.StatusOK, w.Code)
	})
}

func TestStatisticsAPIIntegration(t *testing.T) {
	gin.SetMode(gin.TestMode)
	
	// 连接测试数据库
	ctx := context.Background()
	client, err := mongo.Connect(ctx, options.Client().ApplyURI("mongodb://localhost:27017"))
	require.NoError(t, err)
	defer client.Disconnect(ctx)
	
	db := client.Database("mockserver_stats_integration_test")
	defer db.Drop(ctx)
	
	router := gin.New()
	requestLogRepo := repository.NewMongoRequestLogRepository(db)
	statsHandler := NewStatisticsHandler(requestLogRepo, db)
	
	api := router.Group("/api/v1")
	statsHandler.RegisterRoutes(api)
	
	// 创建测试数据
	now := time.Now()
	testLogs := []*models.RequestLog{
		{
			RequestID:     "stats-001",
			ProjectID:     "proj-001",
			EnvironmentID: "env-001",
			StatusCode:    200,
			Duration:      100,
			SourceIP:      "127.0.0.1",
			Timestamp:     now.Add(-1 * time.Hour),
		},
		{
			RequestID:     "stats-002",
			ProjectID:     "proj-001",
			EnvironmentID: "env-001",
			StatusCode:    500,
			Duration:      200,
			SourceIP:      "127.0.0.1",
			Timestamp:     now.Add(-30 * time.Minute),
		},
	}
	
	for _, log := range testLogs {
		err := requestLogRepo.Create(ctx, log)
		require.NoError(t, err)
	}
	
	t.Run("Get overview statistics", func(t *testing.T) {
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/api/v1/statistics/overview", nil)
		router.ServeHTTP(w, req)
		
		assert.Equal(t, http.StatusOK, w.Code)
	})
	
	t.Run("Get realtime statistics", func(t *testing.T) {
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/api/v1/statistics/realtime", nil)
		router.ServeHTTP(w, req)
		
		assert.Equal(t, http.StatusOK, w.Code)
	})
	
	t.Run("Get trend analysis", func(t *testing.T) {
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/api/v1/statistics/trend?period=day", nil)
		router.ServeHTTP(w, req)
		
		assert.Equal(t, http.StatusOK, w.Code)
	})
	
	t.Run("Get comparison analysis", func(t *testing.T) {
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/api/v1/statistics/comparison?period=day", nil)
		router.ServeHTTP(w, req)
		
		assert.Equal(t, http.StatusOK, w.Code)
	})
}
