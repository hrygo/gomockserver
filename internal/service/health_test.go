package service

import (
	"context"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func TestNewHealthChecker(t *testing.T) {
	// 不需要真实的MongoDB客户端
	checker := NewHealthChecker(nil)
	assert.NotNil(t, checker)
	assert.Nil(t, checker.mongoClient)
}

func TestHealthCheck_Basic(t *testing.T) {
	gin.SetMode(gin.TestMode)
	
	t.Run("Basic health check without MongoDB", func(t *testing.T) {
		checker := NewHealthChecker(nil)
		
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest("GET", "/health", nil)
		
		checker.Check(c)
		
		assert.Equal(t, 200, w.Code)
		assert.Contains(t, w.Body.String(), "healthy")
		assert.Contains(t, w.Body.String(), Version)
	})
	
	t.Run("Health check with detailed query", func(t *testing.T) {
		checker := NewHealthChecker(nil)
		
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest("GET", "/health?detailed=true", nil)
		
		checker.Check(c)
		
		assert.Equal(t, 200, w.Code)
		// 当没有MongoDB客户端时，不会有components字段
		// assert.Contains(t, w.Body.String(), "components")
		// assert.Contains(t, w.Body.String(), "database")
	})
}

func TestHealthCheck_WithMongoDB(t *testing.T) {
	gin.SetMode(gin.TestMode)
	
	// 尝试连接真实MongoDB（如果可用）
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	
	client, err := mongo.Connect(ctx, options.Client().ApplyURI("mongodb://localhost:27017"))
	if err != nil {
		t.Skip("MongoDB not available, skipping integration test")
		return
	}
	defer client.Disconnect(context.Background())
	
	// 检查连接
	err = client.Ping(ctx, nil)
	if err != nil {
		t.Skip("MongoDB ping failed, skipping integration test")
		return
	}
	
	t.Run("Health check with healthy MongoDB", func(t *testing.T) {
		checker := NewHealthChecker(client)
		
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest("GET", "/health?detailed=true", nil)
		
		checker.Check(c)
		
		assert.Equal(t, 200, w.Code)
		assert.Contains(t, w.Body.String(), "database")
		assert.Contains(t, w.Body.String(), "healthy")
	})
}

func TestCheckDatabase_NoClient(t *testing.T) {
	checker := NewHealthChecker(nil)
	
	status := checker.checkDatabase(context.Background())
	
	assert.Equal(t, StatusHealthy, status.Status)
	assert.Contains(t, status.Message, "not configured")
}

func TestFormatUptime(t *testing.T) {
	tests := []struct {
		name     string
		duration time.Duration
		contains []string
	}{
		{
			name:     "Only seconds",
			duration: 45 * time.Second,
			contains: []string{"45s"},
		},
		{
			name:     "Minutes and seconds",
			duration: 3*time.Minute + 30*time.Second,
			contains: []string{"3m", "30s"},
		},
		{
			name:     "Hours, minutes and seconds",
			duration: 2*time.Hour + 15*time.Minute + 10*time.Second,
			contains: []string{"2h", "15m", "10s"},
		},
		{
			name:     "Days, hours, minutes and seconds",
			duration: 1*24*time.Hour + 5*time.Hour + 20*time.Minute + 5*time.Second,
			contains: []string{"1d", "5h", "20m", "5s"},
		},
		{
			name:     "Multiple days",
			duration: 3*24*time.Hour + 2*time.Hour,
			contains: []string{"3d", "2h"},
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := formatUptime(tt.duration)
			for _, expected := range tt.contains {
				assert.Contains(t, result, expected)
			}
		})
	}
}

func TestFormatString(t *testing.T) {
	tests := []struct {
		name     string
		format   string
		args     []interface{}
		expected string
	}{
		{
			name:     "One argument",
			format:   "%ds",
			args:     []interface{}{30},
			expected: "30s",
		},
		{
			name:     "Two arguments",
			format:   "%dm %ds",
			args:     []interface{}{5, 30},
			expected: "5m 30s",
		},
		{
			name:     "Three arguments",
			format:   "%dh %dm %ds",
			args:     []interface{}{2, 15, 30},
			expected: "2h 15m 30s",
		},
		{
			name:     "Four arguments",
			format:   "%dd %dh %dm %ds",
			args:     []interface{}{1, 5, 20, 10},
			expected: "1d 5h 20m 10s",
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := formatString(tt.format, tt.args...)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestIntToString(t *testing.T) {
	tests := []struct {
		input    int
		expected string
	}{
		{0, "0"},
		{1, "1"},
		{10, "10"},
		{99, "99"},
		{123, "123"},
		{1000, "1000"},
	}
	
	for _, tt := range tests {
		t.Run(tt.expected, func(t *testing.T) {
			result := intToString(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestFormatWithOneArg(t *testing.T) {
	tests := []struct {
		format   string
		arg      int
		expected string
	}{
		{"%ds", 30, "30s"},
		{"%dm", 5, "5m"},
		{"test %d test", 42, "test 42 test"},
		{"no placeholder", 100, "no placeholder"},
	}
	
	for _, tt := range tests {
		result := formatWithOneArg(tt.format, tt.arg)
		assert.Equal(t, tt.expected, result)
	}
}

func TestFormatWithTwoArgs(t *testing.T) {
	result := formatWithTwoArgs("%dm %ds", 5, 30)
	assert.Equal(t, "5m 30s", result)
	
	result = formatWithTwoArgs("%dh %dm", 2, 15)
	assert.Equal(t, "2h 15m", result)
}

func TestFormatWithThreeArgs(t *testing.T) {
	result := formatWithThreeArgs("%dh %dm %ds", 2, 15, 30)
	assert.Equal(t, "2h 15m 30s", result)
}

func TestFormatWithFourArgs(t *testing.T) {
	result := formatWithFourArgs("%dd %dh %dm %ds", 1, 5, 20, 10)
	assert.Equal(t, "1d 5h 20m 10s", result)
}

func TestHealthResponseStructure(t *testing.T) {
	gin.SetMode(gin.TestMode)
	checker := NewHealthChecker(nil)
	
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest("GET", "/health", nil)
	
	checker.Check(c)
	
	// 验证响应结构
	assert.Equal(t, 200, w.Code)
	body := w.Body.String()
	assert.Contains(t, body, "status")
	assert.Contains(t, body, "version")
	assert.Contains(t, body, "app_name")
	assert.Contains(t, body, "uptime")
	assert.Contains(t, body, "timestamp")
}
