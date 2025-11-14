package api

import (
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gomockserver/mockserver/internal/repository"
)

// StatisticsHandler 统计 API 处理器
type StatisticsHandler struct {
	projectRepo     repository.ProjectRepository
	ruleRepo        repository.RuleRepository
	environmentRepo repository.EnvironmentRepository
}

// NewStatisticsHandler 创建统计处理器
func NewStatisticsHandler(
	projectRepo repository.ProjectRepository,
	ruleRepo repository.RuleRepository,
	environmentRepo repository.EnvironmentRepository,
) *StatisticsHandler {
	return &StatisticsHandler{
		projectRepo:     projectRepo,
		ruleRepo:        ruleRepo,
		environmentRepo: environmentRepo,
	}
}

// GetDashboardStatistics 获取仪表盘统计数据
// GET /api/v1/statistics/dashboard
func (h *StatisticsHandler) GetDashboardStatistics(c *gin.Context) {
	ctx := c.Request.Context()

	// 获取项目总数（不分页，获取所有）
	projects, _, err := h.projectRepo.List(ctx, 0, 10000)
	if err != nil {
		c.JSON(500, gin.H{"error": "failed to get projects"})
		return
	}

	// 统计环境总数（遍历所有项目）
	totalEnvironments := 0
	for _, project := range projects {
		envs, _ := h.environmentRepo.FindByProject(ctx, project.ID)
		totalEnvironments += len(envs)
	}

	// 获取规则总数和启用/禁用统计
	rules, _, err := h.ruleRepo.List(ctx, map[string]interface{}{}, 0, 10000)
	if err != nil {
		c.JSON(500, gin.H{"error": "failed to get rules"})
		return
	}

	enabledCount := 0
	disabledCount := 0
	for _, rule := range rules {
		if rule.Enabled {
			enabledCount++
		} else {
			disabledCount++
		}
	}

	c.JSON(200, gin.H{
		"total_projects":     len(projects),
		"total_environments": totalEnvironments,
		"total_rules":        len(rules),
		"enabled_rules":      enabledCount,
		"disabled_rules":     disabledCount,
		"total_requests":     0, // TODO: 需要实现请求日志统计
		"requests_today":     0, // TODO: 需要实现今日请求统计
	})
}

// GetProjectStatistics 获取项目统计
// GET /api/v1/statistics/projects
func (h *StatisticsHandler) GetProjectStatistics(c *gin.Context) {
	ctx := c.Request.Context()

	projects, _, err := h.projectRepo.List(ctx, 0, 10000)
	if err != nil {
		c.JSON(500, gin.H{"error": "failed to get projects"})
		return
	}

	var stats []map[string]interface{}
	for _, project := range projects {
		// 获取该项目的环境数
		environments, _ := h.environmentRepo.FindByProject(ctx, project.ID)

		// 获取该项目的规则数
		filter := map[string]interface{}{"project_id": project.ID}
		rules, _, _ := h.ruleRepo.List(ctx, filter, 0, 10000)

		stats = append(stats, map[string]interface{}{
			"project_id":        project.ID,
			"project_name":      project.Name,
			"environment_count": len(environments),
			"rule_count":        len(rules),
			"request_count":     0, // TODO: 需要实现请求日志统计
		})
	}

	c.JSON(200, stats)
}

// GetRuleStatistics 获取规则统计
// GET /api/v1/statistics/rules
func (h *StatisticsHandler) GetRuleStatistics(c *gin.Context) {
	ctx := c.Request.Context()
	projectID := c.Query("project_id")

	filter := map[string]interface{}{}
	if projectID != "" {
		filter["project_id"] = projectID
	}

	rules, _, err := h.ruleRepo.List(ctx, filter, 0, 10000)
	if err != nil {
		c.JSON(500, gin.H{"error": "failed to get rules"})
		return
	}

	var stats []map[string]interface{}
	for _, rule := range rules {
		stats = append(stats, map[string]interface{}{
			"rule_id":           rule.ID,
			"rule_name":         rule.Name,
			"match_count":       0,   // TODO: 需要实现匹配统计
			"avg_response_time": 0.0, // TODO: 需要实现响应时间统计
			"last_matched_at":   nil, // TODO: 需要实现最后匹配时间
		})
	}

	c.JSON(200, stats)
}

// GetRequestTrend 获取请求趋势（最近7天）
// GET /api/v1/statistics/request-trend
func (h *StatisticsHandler) GetRequestTrend(c *gin.Context) {
	// TODO: 需要实现请求日志统计
	// 当前返回模拟数据
	var trend []map[string]interface{}
	now := time.Now()

	for i := 6; i >= 0; i-- {
		date := now.AddDate(0, 0, -i)
		trend = append(trend, map[string]interface{}{
			"date":  date.Format("2006-01-02"),
			"count": 0,
		})
	}

	c.JSON(200, trend)
}

// GetResponseTimeDistribution 获取响应时间分布
// GET /api/v1/statistics/response-time-distribution
func (h *StatisticsHandler) GetResponseTimeDistribution(c *gin.Context) {
	// TODO: 需要实现响应时间统计
	// 当前返回模拟数据
	distribution := []map[string]interface{}{
		{"range": "0-100ms", "count": 0},
		{"range": "100-500ms", "count": 0},
		{"range": "500-1000ms", "count": 0},
		{"range": "1000ms+", "count": 0},
	}

	c.JSON(200, distribution)
}
