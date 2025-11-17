package middleware

import (
	"bytes"
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gomockserver/mockserver/internal/models"
	"github.com/gomockserver/mockserver/internal/repository"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestResponseWriter_Write(t *testing.T) {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	
	rw := &responseWriter{
		ResponseWriter: c.Writer,
		body:           bytes.NewBufferString(""),
	}
	
	data := []byte("test data")
	n, err := rw.Write(data)
	
	assert.NoError(t, err)
	assert.Equal(t, len(data), n)
	assert.Equal(t, "test data", rw.body.String())
}

func TestResponseWriter_WriteString(t *testing.T) {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	
	rw := &responseWriter{
		ResponseWriter: c.Writer,
		body:           bytes.NewBufferString(""),
	}
	
	str := "test string"
	n, err := rw.WriteString(str)
	
	assert.NoError(t, err)
	assert.Equal(t, len(str), n)
	assert.Equal(t, str, rw.body.String())
}

func TestGetProtocol(t *testing.T) {
	m := &RequestLoggerMiddleware{}
	
	testCases := []struct {
		name     string
		upgrade  string
		expected models.ProtocolType
	}{
		{"WebSocket", "websocket", models.ProtocolWebSocket},
		{"HTTP", "", models.ProtocolHTTP},
	}
	
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			c.Request = httptest.NewRequest("GET", "/test", nil)
			if tc.upgrade != "" {
				c.Request.Header.Set("Upgrade", tc.upgrade)
			}
			
			protocol := m.getProtocol(c)
			assert.Equal(t, tc.expected, protocol)
		})
	}
}

func TestBuildRequestData(t *testing.T) {
	m := &RequestLoggerMiddleware{}
	
	t.Run("With JSON Body", func(t *testing.T) {
		body := []byte(`{"name":"test","value":123}`)
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest("POST", "/api/test?id=1", bytes.NewReader(body))
		c.Request.Header.Set("Content-Type", "application/json")
		c.Request.Header.Set("User-Agent", "TestAgent/1.0")
		
		data := m.buildRequestData(c, body)
		
		assert.NotNil(t, data)
		assert.Equal(t, "POST", data["method"])
		assert.Equal(t, "/api/test", data["path"])
		assert.Contains(t, data, "headers")
		assert.Contains(t, data, "body")
		
		bodyMap := data["body"].(map[string]interface{})
		assert.Equal(t, "test", bodyMap["name"])
		assert.Equal(t, float64(123), bodyMap["value"])
	})
	
	t.Run("With Empty Body", func(t *testing.T) {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest("GET", "/api/test", nil)
		c.Request.Header.Set("Accept", "application/json")
		
		data := m.buildRequestData(c, []byte{})
		
		assert.NotNil(t, data)
		assert.Contains(t, data, "headers")
		assert.NotContains(t, data, "body")
	})
	
	t.Run("With Sensitive Headers", func(t *testing.T) {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest("GET", "/api/test", nil)
		c.Request.Header.Set("Authorization", "Bearer secret-token")
		c.Request.Header.Set("Cookie", "session=abc123")
		c.Request.Header.Set("X-Api-Key", "secret-key")
		c.Request.Header.Set("User-Agent", "TestAgent/1.0")
		
		data := m.buildRequestData(c, []byte{})
		
		assert.NotNil(t, data)
		headersMap := data["headers"].(map[string]string)
		
		// 敏感信息应该被脱敏
		assert.Equal(t, "***REDACTED***", headersMap["Authorization"])
		assert.Equal(t, "***REDACTED***", headersMap["Cookie"])
		
		// 非敏感信息应该保留
		assert.Equal(t, "TestAgent/1.0", headersMap["User-Agent"])
	})
	
	t.Run("With Invalid JSON Body", func(t *testing.T) {
		body := []byte(`{"invalid json`)
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest("POST", "/api/test", bytes.NewReader(body))
		c.Request.Header.Set("Content-Type", "application/json")
		
		data := m.buildRequestData(c, body)
		
		assert.NotNil(t, data)
		assert.Contains(t, data, "headers")
		// 无效 JSON 应该作为字符串存储
		if bodyData, exists := data["body"]; exists {
			assert.IsType(t, "", bodyData)
		}
	})
	
	t.Run("With Large Body", func(t *testing.T) {
		// 创建大于 1000 字节的body
		largeBody := make([]byte, 1100)
		for i := range largeBody {
			largeBody[i] = 'A'
		}
		
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest("POST", "/api/test", bytes.NewReader(largeBody))
		c.Request.Header.Set("Content-Type", "text/plain")
		
		data := m.buildRequestData(c, largeBody)
		
		assert.NotNil(t, data)
		// 大body应该被截断
		if bodyData, exists := data["body"]; exists {
			bodyStr := bodyData.(string)
			assert.Contains(t, bodyStr, "truncated")
		}
	})
}

func TestBuildResponseData(t *testing.T) {
	m := &RequestLoggerMiddleware{}
	
	t.Run("With JSON Response", func(t *testing.T) {
		body := []byte(`{"status":"success","data":{"id":123}}`)
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Writer.Header().Set("Content-Type", "application/json")
		c.Writer.Header().Set("X-Request-ID", "req-123")
		c.Writer.WriteHeader(200)
		
		data := m.buildResponseData(c, body)
		
		assert.NotNil(t, data)
		assert.Equal(t, 200, data["status_code"])
		assert.Contains(t, data, "headers")
		assert.Contains(t, data, "body")
		
		bodyMap := data["body"].(map[string]interface{})
		assert.Equal(t, "success", bodyMap["status"])
	})
	
	t.Run("With Empty Response", func(t *testing.T) {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Writer.Header().Set("Content-Type", "application/json")
		c.Writer.WriteHeader(204)
		
		data := m.buildResponseData(c, []byte{})
		
		assert.NotNil(t, data)
		assert.Equal(t, 204, data["status_code"])
		assert.Contains(t, data, "headers")
	})
	
	t.Run("With Text Response", func(t *testing.T) {
		body := []byte("Plain text response")
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Writer.Header().Set("Content-Type", "text/plain")
		c.Writer.WriteHeader(200)
		
		data := m.buildResponseData(c, body)
		
		assert.NotNil(t, data)
		assert.Contains(t, data, "body")
		assert.Equal(t, "Plain text response", data["body"])
	})
}

func TestSanitizeHeaders(t *testing.T) {
	m := &RequestLoggerMiddleware{}
	
	headers := http.Header{
		"Authorization": []string{"Bearer token123"},
		"Cookie":        []string{"session=abc; user=xyz"},
		"X-Api-Key":     []string{"api-key-secret"},
		"Content-Type":  []string{"application/json"},
		"User-Agent":    []string{"TestAgent/1.0"},
		"Accept":        []string{"*/*"},
	}
	
	result := m.sanitizeHeaders(headers)
	
	// 敏感字段应该被脱敏
	assert.Equal(t, "***REDACTED***", result["Authorization"])
	assert.Equal(t, "***REDACTED***", result["Cookie"])
	
	// 非敏感字段应该保留
	assert.Equal(t, "application/json", result["Content-Type"])
	assert.Equal(t, "TestAgent/1.0", result["User-Agent"])
	assert.Equal(t, "*/*", result["Accept"])
}

func TestGetStringValue(t *testing.T) {
	testCases := []struct {
		name     string
		value    interface{}
		expected string
	}{
		{"String value", "test", "test"},
		{"Nil value", nil, ""},
		{"Integer value", 123, ""},
		{"Empty string", "", ""},
	}
	
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := getStringValue(tc.value)
			assert.Equal(t, tc.expected, result)
		})
	}
}

func TestMultipleWrites(t *testing.T) {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	
	rw := &responseWriter{
		ResponseWriter: c.Writer,
		body:           bytes.NewBufferString(""),
	}
	
	// 多次写入
	rw.Write([]byte("Part1"))
	rw.WriteString(" Part2")
	rw.Write([]byte(" Part3"))
	
	assert.Equal(t, "Part1 Part2 Part3", rw.body.String())
}

// Mock RequestLogRepository
type MockRequestLogRepo struct {
	mock.Mock
}

func (m *MockRequestLogRepo) Create(ctx context.Context, log *models.RequestLog) error {
	args := m.Called(ctx, log)
	return args.Error(0)
}

func (m *MockRequestLogRepo) FindByID(ctx context.Context, id string) (*models.RequestLog, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.RequestLog), args.Error(1)
}

func (m *MockRequestLogRepo) List(ctx context.Context, filter repository.RequestLogFilter) ([]*models.RequestLog, int64, error) {
	args := m.Called(ctx, filter)
	if args.Get(0) == nil {
		return nil, args.Get(1).(int64), args.Error(2)
	}
	return args.Get(0).([]*models.RequestLog), args.Get(1).(int64), args.Error(2)
}

func (m *MockRequestLogRepo) DeleteBefore(ctx context.Context, before time.Time) (int64, error) {
	args := m.Called(ctx, before)
	return args.Get(0).(int64), args.Error(1)
}

func (m *MockRequestLogRepo) DeleteByProjectID(ctx context.Context, projectID string) error {
	args := m.Called(ctx, projectID)
	return args.Error(0)
}

func (m *MockRequestLogRepo) CountByProjectID(ctx context.Context, projectID string, startTime, endTime time.Time) (int64, error) {
	args := m.Called(ctx, projectID, startTime, endTime)
	return args.Get(0).(int64), args.Error(1)
}

func (m *MockRequestLogRepo) GetStatistics(ctx context.Context, projectID, environmentID string, startTime, endTime time.Time) (*repository.RequestLogStatistics, error) {
	args := m.Called(ctx, projectID, environmentID, startTime, endTime)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*repository.RequestLogStatistics), args.Error(1)
}

func TestNewRequestLoggerMiddleware(t *testing.T) {
	mockRepo := new(MockRequestLogRepo)
	middleware := NewRequestLoggerMiddleware(mockRepo)
	
	assert.NotNil(t, middleware)
	assert.True(t, middleware.enabled)
	assert.Equal(t, mockRepo, middleware.repo)
}

func TestRequestLoggerMiddleware_SetEnabled(t *testing.T) {
	mockRepo := new(MockRequestLogRepo)
	middleware := NewRequestLoggerMiddleware(mockRepo)
	
	middleware.SetEnabled(false)
	assert.False(t, middleware.enabled)
	
	middleware.SetEnabled(true)
	assert.True(t, middleware.enabled)
}

func TestRequestLoggerMiddleware_Handler(t *testing.T) {
	gin.SetMode(gin.TestMode)
	
	t.Run("Successful logging", func(t *testing.T) {
		mockRepo := new(MockRequestLogRepo)
		middleware := NewRequestLoggerMiddleware(mockRepo)
		
		// Mock Create to succeed
		mockRepo.On("Create", mock.Anything, mock.AnythingOfType("*models.RequestLog")).Return(nil).Once()
		
		w := httptest.NewRecorder()
		c, router := gin.CreateTestContext(w)
		
		// Set context values
		c.Set("request_id", "test-req-123")
		c.Set("project_id", "proj-001")
		c.Set("environment_id", "env-001")
		c.Set("rule_id", "rule-001")
		
		router.Use(middleware.Handler())
		router.GET("/test", func(c *gin.Context) {
			c.JSON(200, gin.H{"status": "ok"})
		})
		
		req := httptest.NewRequest("GET", "/test", nil)
		router.ServeHTTP(w, req)
		
		assert.Equal(t, 200, w.Code)
		
		// Wait a bit for async goroutine to complete
		time.Sleep(100 * time.Millisecond)
		mockRepo.AssertExpectations(t)
	})
	
	t.Run("Logging when disabled", func(t *testing.T) {
		mockRepo := new(MockRequestLogRepo)
		middleware := NewRequestLoggerMiddleware(mockRepo)
		middleware.SetEnabled(false)
		
		w := httptest.NewRecorder()
		_, router := gin.CreateTestContext(w)
		
		router.Use(middleware.Handler())
		router.GET("/test", func(c *gin.Context) {
			c.JSON(200, gin.H{"status": "ok"})
		})
		
		req := httptest.NewRequest("GET", "/test", nil)
		router.ServeHTTP(w, req)
		
		assert.Equal(t, 200, w.Code)
		
		// Should not call Create when disabled
		time.Sleep(50 * time.Millisecond)
		mockRepo.AssertNotCalled(t, "Create")
	})
	
	t.Run("Logging with request body", func(t *testing.T) {
		mockRepo := new(MockRequestLogRepo)
		middleware := NewRequestLoggerMiddleware(mockRepo)
		
		mockRepo.On("Create", mock.Anything, mock.AnythingOfType("*models.RequestLog")).Return(nil).Once()
		
		w := httptest.NewRecorder()
		_, router := gin.CreateTestContext(w)
		
		router.Use(middleware.Handler())
		router.POST("/api/test", func(c *gin.Context) {
			c.JSON(201, gin.H{"created": true})
		})
		
		body := bytes.NewBufferString(`{"name":"test"}`)
		req := httptest.NewRequest("POST", "/api/test", body)
		req.Header.Set("Content-Type", "application/json")
		router.ServeHTTP(w, req)
		
		assert.Equal(t, 201, w.Code)
		
		time.Sleep(100 * time.Millisecond)
		mockRepo.AssertExpectations(t)
	})
}
