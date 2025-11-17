package service

import (
	"runtime"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gomockserver/mockserver/internal/api"
	"github.com/gomockserver/mockserver/internal/middleware"
	"github.com/gomockserver/mockserver/internal/models"
	"github.com/gomockserver/mockserver/pkg/logger"
	"go.uber.org/zap"
)

// AdminService 管理服务
type AdminService struct {
	ruleHandler         *api.RuleHandler
	projectHandler      *api.ProjectHandler
	statisticsHandler   *api.StatisticsHandler
	mockHandler         *api.MockHandler
	importExportService ImportExportService
}

// NewAdminService 创建管理服务
func NewAdminService(
	ruleHandler *api.RuleHandler,
	projectHandler *api.ProjectHandler,
	statisticsHandler *api.StatisticsHandler,
	importExportService ImportExportService,
) *AdminService {
	return &AdminService{
		ruleHandler:         ruleHandler,
		projectHandler:      projectHandler,
		statisticsHandler:   statisticsHandler,
		mockHandler:         api.NewMockHandler(),
		importExportService: importExportService,
	}
}

// StartAdminServer 启动管理服务器
func StartAdminServer(addr string, service *AdminService) error {
	gin.SetMode(gin.ReleaseMode)
	r := gin.New()
	r.Use(gin.Recovery())
	// 使用新的 CORS 中间件
	r.Use(middleware.CORS())

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
			
			// 环境管理 API (在项目下)
			environments := projects.Group("/:id/environments")
			{
				environments.GET("", service.projectHandler.ListEnvironments)
				environments.POST("", service.projectHandler.CreateEnvironment)
				environments.GET("/:env_id", service.projectHandler.GetEnvironment)
				environments.PUT("/:env_id", service.projectHandler.UpdateEnvironment)
				environments.DELETE("/:env_id", service.projectHandler.DeleteEnvironment)
			}
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
			statistics.GET("/dashboard", service.getDashboardStatistics)
			statistics.GET("/projects", service.getProjectStatistics)
			statistics.GET("/rules", service.getRuleStatistics)
			statistics.GET("/request-trend", service.getRequestTrend)
			statistics.GET("/response-time-distribution", service.getResponseTimeDistribution)
		}

		// Mock API
		mock := v1.Group("/mock")
		{
			mock.POST("/test", service.mockHandler.SendMockRequest)
			mock.GET("/history", service.mockHandler.GetMockHistory)
			mock.DELETE("/history", service.mockHandler.ClearMockHistory)
			mock.DELETE("/history/:id", service.mockHandler.DeleteMockHistoryItem)
		}

		// 导入导出 API
		if service.importExportService != nil {
			importExport := v1.Group("/import-export")
			{
				importExport.GET("/projects/:id/export", service.ExportProject)
				importExport.POST("/rules/export", service.ExportRules)
				importExport.POST("/import", service.ImportData)
				importExport.POST("/validate", service.ValidateImportData)
			}
		}
	}

	logger.Info("starting admin server", zap.String("address", addr))
	return r.Run(addr)
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
		"version": "0.6.0",
		"name":    "MockServer",
	})
}

func GetSystemInfo(c *gin.Context) {
	c.JSON(200, gin.H{
		"version":          "0.6.0",
		"build_time":       "2025-11-18",
		"go_version":       runtime.Version(),
		"admin_api_url":    "http://localhost:8080/api/v1",
		"mock_service_url": "http://localhost:9090",
	})
}

// ExportProject 导出项目
func (s *AdminService) ExportProject(c *gin.Context) {
	projectID := c.Param("id")
	if projectID == "" {
		c.JSON(400, gin.H{"error": "project_id is required"})
		return
	}

	includeMetadata := c.DefaultQuery("include_metadata", "false") == "true"

	exportData, err := s.importExportService.ExportProject(c.Request.Context(), projectID, includeMetadata)
	if err != nil {
		logger.Error("failed to export project", zap.String("project_id", projectID), zap.Error(err))
		c.JSON(500, gin.H{"error": "failed to export project: " + err.Error()})
		return
	}

	c.JSON(200, exportData)
}

// ExportRules 导出规则
func (s *AdminService) ExportRules(c *gin.Context) {
	var req models.ExportRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"error": "invalid request: " + err.Error()})
		return
	}

	exportData, err := s.importExportService.ExportRules(c.Request.Context(), &req)
	if err != nil {
		logger.Error("failed to export rules", zap.Error(err))
		c.JSON(500, gin.H{"error": "failed to export rules: " + err.Error()})
		return
	}

	c.JSON(200, exportData)
}

// ImportData 导入数据
func (s *AdminService) ImportData(c *gin.Context) {
	var req models.ImportRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"error": "invalid request: " + err.Error()})
		return
	}

	// 设置默认策略
	if req.Strategy == "" {
		req.Strategy = models.ImportStrategySkip
	}

	result, err := s.importExportService.ImportData(c.Request.Context(), &req)
	if err != nil {
		logger.Error("failed to import data", zap.Error(err))
		c.JSON(500, gin.H{"error": "failed to import data: " + err.Error()})
		return
	}

	// 根据结果返回适当的状态码
	statusCode := 200
	if !result.Success {
		statusCode = 207 // 部分成功
	}

	c.JSON(statusCode, result)
}

// ValidateImportData 验证导入数据
func (s *AdminService) ValidateImportData(c *gin.Context) {
	var data models.ExportData
	if err := c.ShouldBindJSON(&data); err != nil {
		c.JSON(400, gin.H{"error": "invalid data format: " + err.Error()})
		return
	}

	if err := s.importExportService.ValidateImportData(c.Request.Context(), &data); err != nil {
		c.JSON(400, gin.H{"error": "validation failed: " + err.Error()})
		return
	}

	c.JSON(200, gin.H{
		"message": "validation successful",
		"data": gin.H{
			"rule_count":        len(data.Data.Rules),
			"environment_count": len(data.Data.Environments),
			"has_project":       data.Data.Project != nil,
		},
	})
}

// getDashboardStatistics 获取仪表盘统计数据
func (s *AdminService) getDashboardStatistics(c *gin.Context) {
	// 这里返回模拟数据，实际应该从数据库查询
	c.JSON(200, gin.H{
		"total_projects":     0,
		"total_environments": 0,
		"total_rules":        0,
		"total_requests":     0,
		"enabled_rules":      0,
		"disabled_rules":     0,
		"requests_today":     0,
	})
}

// getProjectStatistics 获取项目统计
func (s *AdminService) getProjectStatistics(c *gin.Context) {
	// 返回空数组，实际应该从数据库查询
	c.JSON(200, []interface{}{})
}

// getRuleStatistics 获取规则统计
func (s *AdminService) getRuleStatistics(c *gin.Context) {
	// 返回空数组，实际应该从数据库查询
	c.JSON(200, []interface{}{})
}

// getRequestTrend 获取请求趋势
func (s *AdminService) getRequestTrend(c *gin.Context) {
	// 返回空数组，实际应该从数据库查询
	c.JSON(200, []interface{}{})
}

// getResponseTimeDistribution 获取响应时间分布
func (s *AdminService) getResponseTimeDistribution(c *gin.Context) {
	// 返回空数组，实际应该从数据库查询
	c.JSON(200, []interface{}{})
}
