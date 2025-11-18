package service

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/gomockserver/mockserver/internal/middleware"
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
	service := NewAdminService(nil, nil, nil, nil)
	assert.NotNil(t, service)
	assert.Nil(t, service.ruleHandler)
	assert.Nil(t, service.projectHandler)
	assert.Nil(t, service.statisticsHandler)
	assert.Nil(t, service.importExportService)
}

// TestCORSMiddleware 测试 CORS 中间件
func TestCORSMiddleware(t *testing.T) {
	tests := []struct {
		name           string
		method         string
		origin         string
		expectedStatus int
		checkHeaders   bool
	}{
		{
			name:           "OPTIONS 请求返回 204",
			method:         http.MethodOptions,
			origin:         "http://localhost:5173",
			expectedStatus: http.StatusNoContent,
			checkHeaders:   true,
		},
		{
			name:           "GET 请求正常处理",
			method:         http.MethodGet,
			origin:         "http://localhost:5173",
			expectedStatus: http.StatusOK,
			checkHeaders:   true,
		},
		{
			name:           "POST 请求正常处理",
			method:         http.MethodPost,
			origin:         "http://localhost:8080",
			expectedStatus: http.StatusOK,
			checkHeaders:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			router := setupTestRouter()
			router.Use(middleware.CORS())
			router.GET("/test", func(c *gin.Context) {
				c.JSON(http.StatusOK, gin.H{"message": "ok"})
			})
			router.POST("/test", func(c *gin.Context) {
				c.JSON(http.StatusOK, gin.H{"message": "ok"})
			})

			req := httptest.NewRequest(tt.method, "/test", nil)
			req.Header.Set("Origin", tt.origin)
			w := httptest.NewRecorder()

			router.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)

			if tt.checkHeaders {
				assert.Equal(t, tt.origin, w.Header().Get("Access-Control-Allow-Origin"))
				assert.Equal(t, "true", w.Header().Get("Access-Control-Allow-Credentials"))
				// gin-contrib/cors 只在 OPTIONS 请求时返回完整的 CORS 头
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
	assert.Contains(t, w.Body.String(), "0.6.0")
	assert.Contains(t, w.Body.String(), "MockServer")
}

// TestGetSystemInfo 测试获取系统信息
func TestGetSystemInfo(t *testing.T) {
	router := setupTestRouter()
	router.GET("/info", GetSystemInfo)

	req := httptest.NewRequest(http.MethodGet, "/info", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), "version")
	assert.Contains(t, w.Body.String(), "0.6.0")
	assert.Contains(t, w.Body.String(), "build_time")
	assert.Contains(t, w.Body.String(), "go_version")
	assert.Contains(t, w.Body.String(), "admin_api_url")
	assert.Contains(t, w.Body.String(), "mock_service_url")
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
			router.Use(middleware.CORS())

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
	router.Use(middleware.CORS())
	router.GET("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "ok"})
	})

	req := httptest.NewRequest(http.MethodOptions, "/test", nil)
	req.Header.Set("Origin", "http://localhost:5173")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	// OPTIONS 请求应该返回 204 并且不继续处理
	assert.Equal(t, http.StatusNoContent, w.Code)
}

// TestCORSMiddleware_Headers 测试 CORS 头部设置
func TestCORSMiddleware_Headers(t *testing.T) {
	router := setupTestRouter()
	router.Use(middleware.CORS())
	router.GET("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "ok"})
	})

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.Header.Set("Origin", "http://localhost:5173")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	// 验证必要的 CORS 头部
	assert.Equal(t, "http://localhost:5173", w.Header().Get("Access-Control-Allow-Origin"))
	assert.Equal(t, "true", w.Header().Get("Access-Control-Allow-Credentials"))
	// 非 OPTIONS 请求可能不返回所有 CORS 头，这是正常的
}

// TestStartAdminServer 测试启动管理服务器
func TestStartAdminServer(t *testing.T) {
	// 创建一个AdminService实例
	service := NewAdminService(nil, nil, nil, nil)

	// 测试无效地址
	err := StartAdminServer("invalid-address", service)
	assert.Error(t, err)
}

// TestStartMockServer 测试启动Mock服务器
func TestStartMockServer(t *testing.T) {
	// 创建必要的依赖
	mockService := NewMockService(nil, nil)

	// 测试无效地址
	err := StartMockServer("invalid-address", mockService)
	assert.Error(t, err)
}
