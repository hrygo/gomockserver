package api

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gomockserver/mockserver/internal/models"
	"github.com/gomockserver/mockserver/internal/repository"
	"github.com/gomockserver/mockserver/pkg/logger"
	"go.uber.org/zap"
)

// RequestLogHandler 请求日志处理器
type RequestLogHandler struct {
	repo repository.RequestLogRepository
}

// NewRequestLogHandler 创建请求日志处理器
func NewRequestLogHandler(repo repository.RequestLogRepository) *RequestLogHandler {
	return &RequestLogHandler{
		repo: repo,
	}
}

// RegisterRoutes 注册路由
func (h *RequestLogHandler) RegisterRoutes(r *gin.RouterGroup) {
	logs := r.Group("/request-logs")
	{
		logs.GET("", h.ListRequestLogs)
		logs.GET("/:id", h.GetRequestLog)
		logs.DELETE("/cleanup", h.CleanupLogs)
		logs.GET("/statistics", h.GetStatistics)
	}
}

// ListRequestLogsRequest 列表查询请求
type ListRequestLogsRequest struct {
	ProjectID     string `form:"project_id"`
	EnvironmentID string `form:"environment_id"`
	RuleID        string `form:"rule_id"`
	Protocol      string `form:"protocol"`
	Method        string `form:"method"`
	Path          string `form:"path"`
	StatusCode    int    `form:"status_code"`
	SourceIP      string `form:"source_ip"`
	StartTime     string `form:"start_time"` // RFC3339 格式
	EndTime       string `form:"end_time"`   // RFC3339 格式
	Page          int    `form:"page"`
	PageSize      int    `form:"page_size"`
	SortBy        string `form:"sort_by"`
	SortOrder     string `form:"sort_order"` // asc, desc
}

// ListRequestLogsResponse 列表查询响应
type ListRequestLogsResponse struct {
	Data  []*models.RequestLog `json:"data"`
	Total int64                `json:"total"`
	Page  int                  `json:"page"`
	Size  int                  `json:"size"`
}

// ListRequestLogs 列表查询请求日志
// @Summary 列表查询请求日志
// @Tags RequestLogs
// @Accept json
// @Produce json
// @Param project_id query string false "项目ID"
// @Param environment_id query string false "环境ID"
// @Param rule_id query string false "规则ID"
// @Param protocol query string false "协议类型"
// @Param method query string false "HTTP方法"
// @Param path query string false "路径（支持正则）"
// @Param status_code query int false "状态码"
// @Param source_ip query string false "来源IP"
// @Param start_time query string false "开始时间（RFC3339格式）"
// @Param end_time query string false "结束时间（RFC3339格式）"
// @Param page query int false "页码" default(1)
// @Param page_size query int false "每页数量" default(20)
// @Param sort_by query string false "排序字段" default(timestamp)
// @Param sort_order query string false "排序方向" default(desc)
// @Success 200 {object} ListRequestLogsResponse
// @Router /api/v1/request-logs [get]
func (h *RequestLogHandler) ListRequestLogs(c *gin.Context) {
	var req ListRequestLogsRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// 设置默认值
	if req.Page == 0 {
		req.Page = 1
	}
	if req.PageSize == 0 {
		req.PageSize = 20
	}

	// 构建过滤器
	filter := repository.RequestLogFilter{
		ProjectID:     req.ProjectID,
		EnvironmentID: req.EnvironmentID,
		RuleID:        req.RuleID,
		Protocol:      models.ProtocolType(req.Protocol),
		Method:        req.Method,
		Path:          req.Path,
		StatusCode:    req.StatusCode,
		SourceIP:      req.SourceIP,
		Page:          req.Page,
		PageSize:      req.PageSize,
		SortBy:        req.SortBy,
	}

	// 解析时间
	if req.StartTime != "" {
		if t, err := time.Parse(time.RFC3339, req.StartTime); err == nil {
			filter.StartTime = t
		}
	}
	if req.EndTime != "" {
		if t, err := time.Parse(time.RFC3339, req.EndTime); err == nil {
			filter.EndTime = t
		}
	}

	// 解析排序方向
	if req.SortOrder == "asc" {
		filter.SortOrder = 1
	} else {
		filter.SortOrder = -1
	}

	// 查询
	logs, total, err := h.repo.List(c.Request.Context(), filter)
	if err != nil {
		logger.Error("failed to list request logs", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to list request logs"})
		return
	}

	c.JSON(http.StatusOK, ListRequestLogsResponse{
		Data:  logs,
		Total: total,
		Page:  req.Page,
		Size:  len(logs),
	})
}

// GetRequestLog 获取单个请求日志
// @Summary 获取单个请求日志
// @Tags RequestLogs
// @Accept json
// @Produce json
// @Param id path string true "日志ID"
// @Success 200 {object} models.RequestLog
// @Router /api/v1/request-logs/{id} [get]
func (h *RequestLogHandler) GetRequestLog(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "id is required"})
		return
	}

	log, err := h.repo.FindByID(c.Request.Context(), id)
	if err != nil {
		logger.Error("failed to get request log", zap.String("id", id), zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get request log"})
		return
	}

	if log == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "request log not found"})
		return
	}

	c.JSON(http.StatusOK, log)
}

// CleanupLogsRequest 清理日志请求
type CleanupLogsRequest struct {
	BeforeDays int `form:"before_days" binding:"required,min=1"`
}

// CleanupLogs 清理旧日志
// @Summary 清理旧日志
// @Tags RequestLogs
// @Accept json
// @Produce json
// @Param before_days query int true "清理多少天前的日志"
// @Success 200 {object} map[string]interface{}
// @Router /api/v1/request-logs/cleanup [delete]
func (h *RequestLogHandler) CleanupLogs(c *gin.Context) {
	var req CleanupLogsRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	before := time.Now().AddDate(0, 0, -req.BeforeDays)
	count, err := h.repo.DeleteBefore(c.Request.Context(), before)
	if err != nil {
		logger.Error("failed to cleanup logs", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to cleanup logs"})
		return
	}

	logger.Info("cleanup logs completed",
		zap.Int("before_days", req.BeforeDays),
		zap.Int64("deleted_count", count))

	c.JSON(http.StatusOK, gin.H{
		"message":       "cleanup completed",
		"deleted_count": count,
		"before":        before,
	})
}

// GetStatisticsRequest 统计查询请求
type GetStatisticsRequest struct {
	ProjectID     string `form:"project_id"`
	EnvironmentID string `form:"environment_id"`
	StartTime     string `form:"start_time"` // RFC3339 格式
	EndTime       string `form:"end_time"`   // RFC3339 格式
	Period        string `form:"period"`     // 24h, 7d, 30d
}

// GetStatistics 获取统计信息
// @Summary 获取请求日志统计信息
// @Tags RequestLogs
// @Accept json
// @Produce json
// @Param project_id query string false "项目ID"
// @Param environment_id query string false "环境ID"
// @Param start_time query string false "开始时间（RFC3339格式）"
// @Param end_time query string false "结束时间（RFC3339格式）"
// @Param period query string false "时间段（24h/7d/30d）"
// @Success 200 {object} repository.RequestLogStatistics
// @Router /api/v1/request-logs/statistics [get]
func (h *RequestLogHandler) GetStatistics(c *gin.Context) {
	var req GetStatisticsRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var startTime, endTime time.Time

	// 解析时间范围
	if req.StartTime != "" {
		if t, err := time.Parse(time.RFC3339, req.StartTime); err == nil {
			startTime = t
		}
	}
	if req.EndTime != "" {
		if t, err := time.Parse(time.RFC3339, req.EndTime); err == nil {
			endTime = t
		}
	}

	// 如果没有指定时间，根据 period 设置
	if startTime.IsZero() && req.Period != "" {
		endTime = time.Now()
		switch req.Period {
		case "24h":
			startTime = endTime.Add(-24 * time.Hour)
		case "7d":
			startTime = endTime.AddDate(0, 0, -7)
		case "30d":
			startTime = endTime.AddDate(0, 0, -30)
		default:
			startTime = endTime.Add(-24 * time.Hour)
		}
	}

	// 获取统计信息
	stats, err := h.repo.GetStatistics(c.Request.Context(), req.ProjectID, req.EnvironmentID, startTime, endTime)
	if err != nil {
		logger.Error("failed to get statistics", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get statistics"})
		return
	}

	c.JSON(http.StatusOK, stats)
}
