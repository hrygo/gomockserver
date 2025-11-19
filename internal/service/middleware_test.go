package service

import (
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func TestRequestIDMiddleware(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("Generate new request ID", func(t *testing.T) {
		w := httptest.NewRecorder()
		_, router := gin.CreateTestContext(w)

		router.Use(RequestIDMiddleware())
		router.GET("/test", func(c *gin.Context) {
			requestID, exists := c.Get(RequestIDKey)
			assert.True(t, exists)
			assert.NotEmpty(t, requestID)
			c.String(200, "OK")
		})

		req := httptest.NewRequest("GET", "/test", nil)
		router.ServeHTTP(w, req)

		assert.Equal(t, 200, w.Code)
		assert.NotEmpty(t, w.Header().Get(RequestIDHeader))
	})

	t.Run("Use existing request ID from header", func(t *testing.T) {
		w := httptest.NewRecorder()
		_, router := gin.CreateTestContext(w)

		existingID := "existing-req-id-12345"

		router.Use(RequestIDMiddleware())
		router.GET("/test", func(c *gin.Context) {
			requestID, _ := c.Get(RequestIDKey)
			assert.Equal(t, existingID, requestID)
			c.String(200, "OK")
		})

		req := httptest.NewRequest("GET", "/test", nil)
		req.Header.Set(RequestIDHeader, existingID)
		router.ServeHTTP(w, req)

		assert.Equal(t, 200, w.Code)
		assert.Equal(t, existingID, w.Header().Get(RequestIDHeader))
	})
}

func TestPerformanceMiddleware(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("Record normal request performance", func(t *testing.T) {
		w := httptest.NewRecorder()
		_, router := gin.CreateTestContext(w)

		router.Use(RequestIDMiddleware())
		router.Use(PerformanceMiddleware())
		router.GET("/test", func(c *gin.Context) {
			c.String(200, "OK")
		})

		req := httptest.NewRequest("GET", "/test", nil)
		router.ServeHTTP(w, req)

		assert.Equal(t, 200, w.Code)
	})

	t.Run("Record slow request", func(t *testing.T) {
		w := httptest.NewRecorder()
		_, router := gin.CreateTestContext(w)

		router.Use(RequestIDMiddleware())
		router.Use(PerformanceMiddleware())
		router.GET("/slow", func(c *gin.Context) {
			// 模拟慢请求
			time.Sleep(1100 * time.Millisecond)
			c.String(200, "OK")
		})

		req := httptest.NewRequest("GET", "/slow", nil)
		start := time.Now()
		router.ServeHTTP(w, req)
		duration := time.Since(start)

		assert.Equal(t, 200, w.Code)
		assert.GreaterOrEqual(t, duration.Milliseconds(), int64(1100))
	})
}

func TestLoggingMiddleware(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("Log request with request ID", func(t *testing.T) {
		w := httptest.NewRecorder()
		_, router := gin.CreateTestContext(w)

		router.Use(RequestIDMiddleware())
		router.Use(LoggingMiddleware())
		router.GET("/test", func(c *gin.Context) {
			c.String(200, "OK")
		})

		req := httptest.NewRequest("GET", "/test", nil)
		req.Header.Set("User-Agent", "test-agent")
		router.ServeHTTP(w, req)

		assert.Equal(t, 200, w.Code)
	})

	t.Run("Log request without request ID", func(t *testing.T) {
		w := httptest.NewRecorder()
		_, router := gin.CreateTestContext(w)

		// 不使用RequestIDMiddleware
		router.Use(LoggingMiddleware())
		router.GET("/test", func(c *gin.Context) {
			c.String(200, "OK")
		})

		req := httptest.NewRequest("GET", "/test", nil)
		router.ServeHTTP(w, req)

		assert.Equal(t, 200, w.Code)
	})
}

func TestGenerateRequestID(t *testing.T) {
	// 生成多个request ID并验证唯一性
	ids := make(map[string]bool)

	for i := 0; i < 100; i++ {
		id := generateRequestID()
		assert.NotEmpty(t, id)
		assert.Contains(t, id, "req-")

		// 验证唯一性
		assert.False(t, ids[id], "Request ID should be unique")
		ids[id] = true

		// 短暂延迟以确保时间戳不同
		time.Sleep(time.Microsecond)
	}

	assert.Equal(t, 100, len(ids))
}

func TestInt64ToString(t *testing.T) {
	tests := []struct {
		input    int64
		expected string
	}{
		{0, "0"},
		{1, "1"},
		{10, "10"},
		{123, "123"},
		{999999, "999999"},
		{-1, "-1"},
		{-123, "-123"},
		{-999999, "-999999"},
		{1234567890, "1234567890"},
		{-1234567890, "-1234567890"},
	}

	for _, tt := range tests {
		t.Run(tt.expected, func(t *testing.T) {
			result := int64ToString(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestMiddlewareChain(t *testing.T) {
	gin.SetMode(gin.TestMode)

	w := httptest.NewRecorder()
	_, router := gin.CreateTestContext(w)

	// 完整的中间件链
	router.Use(RequestIDMiddleware())
	router.Use(LoggingMiddleware())
	router.Use(PerformanceMiddleware())
	router.GET("/test", func(c *gin.Context) {
		requestID, exists := c.Get(RequestIDKey)
		assert.True(t, exists)
		assert.NotEmpty(t, requestID)
		c.JSON(200, gin.H{"status": "ok"})
	})

	req := httptest.NewRequest("GET", "/test", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, 200, w.Code)
	assert.NotEmpty(t, w.Header().Get(RequestIDHeader))
	assert.Contains(t, w.Body.String(), "ok")
}

func TestRequestIDPropagation(t *testing.T) {
	gin.SetMode(gin.TestMode)

	w := httptest.NewRecorder()
	_, router := gin.CreateTestContext(w)

	var capturedID string

	router.Use(RequestIDMiddleware())
	router.GET("/test", func(c *gin.Context) {
		requestID, _ := c.Get(RequestIDKey)
		capturedID = requestID.(string)
		c.String(200, "OK")
	})

	req := httptest.NewRequest("GET", "/test", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, 200, w.Code)
	assert.NotEmpty(t, capturedID)
	assert.Equal(t, capturedID, w.Header().Get(RequestIDHeader))
}
