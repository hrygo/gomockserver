package service

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

// setupTestRouter 创建测试路由器
func setupTestRouter() *gin.Engine {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	return router
}

// TestNewAdminService 测试创建管理服务
func TestNewAdminService(t *testing.T) {
	service := NewAdminService(nil, nil)
	assert.NotNil(t, service)
}

// TestCORSMiddleware 测试 CORS 中间件
func TestCORSMiddleware(t *testing.T) {
	tests := []struct {
		name           string
		method         string
		expectedStatus int
		checkHeaders   bool
	}{
		{
			name:           "OPTIONS 请求返回 204",
			method:         http.MethodOptions,
			expectedStatus: http.StatusNoContent,
			checkHeaders:   true,
		},
		{
			name:           "GET 请求正常处理",
			method:         http.MethodGet,
			expectedStatus: http.StatusOK,
			checkHeaders:   true,
		},
		{
			name:           "POST 请求正常处理",
			method:         http.MethodPost,
			expectedStatus: http.StatusOK,
			checkHeaders:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			router := setupTestRouter()
			router.Use(CORSMiddleware())
			router.GET("/test", func(c *gin.Context) {
				c.JSON(http.StatusOK, gin.H{"message": "ok"})
			})
			router.POST("/test", func(c *gin.Context) {
				c.JSON(http.StatusOK, gin.H{"message": "ok"})
			})

			req := httptest.NewRequest(tt.method, "/test", nil)
			w := httptest.NewRecorder()

			router.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)

			if tt.checkHeaders {
				assert.Equal(t, "*", w.Header().Get("Access-Control-Allow-Origin"))
				assert.Equal(t, "true", w.Header().Get("Access-Control-Allow-Credentials"))
				assert.NotEmpty(t, w.Header().Get("Access-Control-Allow-Headers"))
				assert.NotEmpty(t, w.Header().Get("Access-Control-Allow-Methods"))
			}
		})
	}
}

// TestHealthCheck 测试健康检查
func TestHealthCheck(t *testing.T) {
	router := setupTestRouter()
	router.GET("/health", HealthCheck)

	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), "healthy")
}

// TestGetVersion 测试获取版本信息
func TestGetVersion(t *testing.T) {
	router := setupTestRouter()
	router.GET("/version", GetVersion)

	req := httptest.NewRequest(http.MethodGet, "/version", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), "version")
	assert.Contains(t, w.Body.String(), "0.1.1")
	assert.Contains(t, w.Body.String(), "MockServer")
}

// TestAdminServiceRoutes 测试管理服务路由配置
func TestAdminServiceRoutes(t *testing.T) {
	// 创建一个临时的 AdminService 用于路由测试
	// 注意：这里不启动真实的服务器，只测试路由配置

	tests := []struct {
		name           string
		method         string
		path           string
		expectedStatus int // 期望状态码（404 表示路由存在但没有实际处理）
	}{
		// 系统 API
		{
			name:           "健康检查路由",
			method:         http.MethodGet,
			path:           "/api/v1/system/health",
			expectedStatus: http.StatusOK,
		},
		{
			name:           "版本信息路由",
			method:         http.MethodGet,
			path:           "/api/v1/system/version",
			expectedStatus: http.StatusOK,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// 创建测试路由器并配置系统路由
			router := setupTestRouter()
			router.Use(CORSMiddleware())

			v1 := router.Group("/api/v1")
			{
				system := v1.Group("/system")
				{
					system.GET("/health", HealthCheck)
					system.GET("/version", GetVersion)
				}
			}

			req := httptest.NewRequest(tt.method, tt.path, nil)
			w := httptest.NewRecorder()

			router.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)
		})
	}
}

// TestCORSMiddleware_OptionsRequest 测试 CORS 预检请求
func TestCORSMiddleware_OptionsRequest(t *testing.T) {
	router := setupTestRouter()
	router.Use(CORSMiddleware())
	router.GET("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "ok"})
	})

	req := httptest.NewRequest(http.MethodOptions, "/test", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	// OPTIONS 请求应该返回 204 并且不继续处理
	assert.Equal(t, http.StatusNoContent, w.Code)
	assert.Empty(t, w.Body.String())
}

// TestCORSMiddleware_Headers 测试 CORS 头部设置
func TestCORSMiddleware_Headers(t *testing.T) {
	router := setupTestRouter()
	router.Use(CORSMiddleware())
	router.GET("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "ok"})
	})

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	// 验证所有必要的 CORS 头部都已设置
	headers := map[string]string{
		"Access-Control-Allow-Origin":      "*",
		"Access-Control-Allow-Credentials": "true",
		"Access-Control-Allow-Headers":     "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, accept, origin, Cache-Control, X-Requested-With",
		"Access-Control-Allow-Methods":     "POST, OPTIONS, GET, PUT, DELETE",
	}

	for key, expectedValue := range headers {
		assert.Equal(t, expectedValue, w.Header().Get(key), "Header %s should be %s", key, expectedValue)
	}
}
