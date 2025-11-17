package service

import (
	"runtime"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gomockserver/mockserver/internal/api"
	"github.com/gomockserver/mockserver/pkg/logger"
	"go.uber.org/zap"
)

// AdminService 管理服务
type AdminService struct {
	ruleHandler       *api.RuleHandler
	projectHandler    *api.ProjectHandler
	statisticsHandler *api.StatisticsHandler
	mockHandler       *api.MockHandler
}

// NewAdminService 创建管理服务
func NewAdminService(ruleHandler *api.RuleHandler, projectHandler *api.ProjectHandler, statisticsHandler *api.StatisticsHandler) *AdminService {
	return &AdminService{
		ruleHandler:       ruleHandler,
		projectHandler:    projectHandler,
		statisticsHandler: statisticsHandler,
		mockHandler:       api.NewMockHandler(),
	}
}

// StartAdminServer 启动管理服务器
func StartAdminServer(addr string, service *AdminService) error {
	gin.SetMode(gin.ReleaseMode)
	r := gin.New()
	r.Use(gin.Recovery())
	r.Use(CORSMiddleware())

	// API 路由组
	v1 := r.Group("/api/v1")
	{
		// 规则管理 API
		rules := v1.Group("/rules")
		{
			rules.GET("", service.ruleHandler.ListRules)
			rules.POST("", service.ruleHandler.CreateRule)
			rules.GET("/:id", service.ruleHandler.GetRule)
			rules.PUT("/:id", service.ruleHandler.UpdateRule)
			rules.DELETE("/:id", service.ruleHandler.DeleteRule)
			rules.POST("/:id/enable", service.ruleHandler.EnableRule)
			rules.POST("/:id/disable", service.ruleHandler.DisableRule)
		}

		// 项目管理 API
		projects := v1.Group("/projects")
		{
			projects.GET("", service.projectHandler.ListProjects)
			projects.POST("", service.projectHandler.CreateProject)
			projects.GET("/:id", service.projectHandler.GetProject)
			projects.PUT("/:id", service.projectHandler.UpdateProject)
			projects.DELETE("/:id", service.projectHandler.DeleteProject)
		}

		// 环境管理 API
		environments := v1.Group("/environments")
		{
			environments.GET("", service.projectHandler.ListEnvironments)
			environments.POST("", service.projectHandler.CreateEnvironment)
			environments.GET("/:id", service.projectHandler.GetEnvironment)
			environments.PUT("/:id", service.projectHandler.UpdateEnvironment)
			environments.DELETE("/:id", service.projectHandler.DeleteEnvironment)
		}

		// 系统管理 API
		system := v1.Group("/system")
		{
			system.GET("/health", HealthCheck)
			system.GET("/version", GetVersion)
			system.GET("/info", GetSystemInfo)
		}

		// 统计 API
		statistics := v1.Group("/statistics")
		{
			statistics.GET("/dashboard", service.statisticsHandler.GetOverview)
			statistics.GET("/projects", service.statisticsHandler.GetOverview)
			statistics.GET("/rules", service.statisticsHandler.GetOverview)
			statistics.GET("/request-trend", service.statisticsHandler.GetTrend)
			statistics.GET("/response-time-distribution", service.statisticsHandler.GetTrend)
		}

		// Mock API
		mock := v1.Group("/mock")
		{
			mock.POST("/test", service.mockHandler.SendMockRequest)
			mock.GET("/history", service.mockHandler.GetMockHistory)
			mock.DELETE("/history", service.mockHandler.ClearMockHistory)
			mock.DELETE("/history/:id", service.mockHandler.DeleteMockHistoryItem)
		}
	}

	logger.Info("starting admin server", zap.String("address", addr))
	return r.Run(addr)
}

// CORSMiddleware CORS 中间件
func CORSMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, accept, origin, Cache-Control, X-Requested-With")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS, GET, PUT, DELETE")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

		c.Next()
	}
}

var startTime = time.Now()

func HealthCheck(c *gin.Context) {
	uptime := time.Since(startTime).Seconds()
	c.JSON(200, gin.H{
		"status":   "healthy",
		"database": true,
		"cache":    true,
		"uptime":   int(uptime),
	})
}

func GetVersion(c *gin.Context) {
	c.JSON(200, gin.H{
		"version": "0.1.1",
		"name":    "MockServer",
	})
}

func GetSystemInfo(c *gin.Context) {
	c.JSON(200, gin.H{
		"version":          "0.2.0",
		"build_time":       "2025-11-14",
		"go_version":       runtime.Version(),
		"admin_api_url":    "http://localhost:8080/api/v1",
		"mock_service_url": "http://localhost:9090",
	})
}
