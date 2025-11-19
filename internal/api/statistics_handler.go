package api

import (
	"context"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gomockserver/mockserver/internal/repository"
	"github.com/gomockserver/mockserver/pkg/logger"
	"go.mongodb.org/mongo-driver/mongo"
	"go.uber.org/zap"
)

// StatisticsHandler 统计分析处理器
type StatisticsHandler struct {
	requestLogRepo repository.RequestLogRepository
	db             *mongo.Database
}

// NewStatisticsHandler 创建统计分析处理器
func NewStatisticsHandler(requestLogRepo repository.RequestLogRepository, db *mongo.Database) *StatisticsHandler {
	return &StatisticsHandler{
		requestLogRepo: requestLogRepo,
		db:             db,
	}
}

// RegisterRoutes 注册路由
func (h *StatisticsHandler) RegisterRoutes(r *gin.RouterGroup) {
	stats := r.Group("/statistics")
	{
		stats.GET("/overview", h.GetOverview)
		stats.GET("/realtime", h.GetRealtime)
		stats.GET("/trend", h.GetTrend)
		stats.GET("/comparison", h.GetComparison)
	}
}

// OverviewResponse 概览统计响应
type OverviewResponse struct {
	TotalRequests        int64            `json:"total_requests"`
	TotalProjects        int64            `json:"total_projects"`
	TotalRules           int64            `json:"total_rules"`
	TotalEnvironments    int64            `json:"total_environments"`
	RequestsToday        int64            `json:"requests_today"`
	SuccessRate          float64          `json:"success_rate"`
	AvgResponseTime      float64          `json:"avg_response_time"`
	TopProjects          []ProjectStats   `json:"top_projects"`
	ProtocolDistribution map[string]int64 `json:"protocol_distribution"`
}

// ProjectStats 项目统计
type ProjectStats struct {
	ProjectID    string  `json:"project_id"`
	ProjectName  string  `json:"project_name"`
	RequestCount int64   `json:"request_count"`
	SuccessRate  float64 `json:"success_rate"`
}

// RealtimeResponse 实时统计响应
type RealtimeResponse struct {
	Timestamp         time.Time               `json:"timestamp"`
	RequestsPerMin    int64                   `json:"requests_per_min"`
	ActiveConnections int                     `json:"active_connections"`
	AvgResponseTime   float64                 `json:"avg_response_time"`
	ErrorRate         float64                 `json:"error_rate"`
	ProtocolStats     map[string]ProtocolStat `json:"protocol_stats"`
}

// ProtocolStat 协议统计
type ProtocolStat struct {
	Count       int64   `json:"count"`
	SuccessRate float64 `json:"success_rate"`
	AvgDuration float64 `json:"avg_duration"`
}

// TrendResponse 趋势分析响应
type TrendResponse struct {
	Period     string       `json:"period"` // hour, day, week, month
	DataPoints []TrendPoint `json:"data_points"`
}

// TrendPoint 趋势数据点
type TrendPoint struct {
	Timestamp    time.Time `json:"timestamp"`
	RequestCount int64     `json:"request_count"`
	SuccessCount int64     `json:"success_count"`
	ErrorCount   int64     `json:"error_count"`
	AvgDuration  float64   `json:"avg_duration"`
}

// ComparisonResponse 对比分析响应
type ComparisonResponse struct {
	CurrentPeriod  PeriodStats `json:"current_period"`
	PreviousPeriod PeriodStats `json:"previous_period"`
	Changes        ChangeStats `json:"changes"`
}

// PeriodStats 时段统计
type PeriodStats struct {
	StartTime     time.Time `json:"start_time"`
	EndTime       time.Time `json:"end_time"`
	TotalRequests int64     `json:"total_requests"`
	SuccessRate   float64   `json:"success_rate"`
	AvgDuration   float64   `json:"avg_duration"`
	ErrorCount    int64     `json:"error_count"`
}

// ChangeStats 变化统计
type ChangeStats struct {
	RequestsChange    float64 `json:"requests_change"`     // 百分比
	SuccessRateChange float64 `json:"success_rate_change"` // 百分比
	DurationChange    float64 `json:"duration_change"`     // 百分比
}

// GetOverview 获取概览统计
// @Summary 获取概览统计
// @Description 获取系统整体统计概览
// @Tags Statistics
// @Produce json
// @Success 200 {object} OverviewResponse
// @Router /api/v1/statistics/overview [get]
func (h *StatisticsHandler) GetOverview(c *gin.Context) {
	ctx := c.Request.Context()

	// 获取今天的开始时间
	today := time.Now().Truncate(24 * time.Hour)

	// 查询今日请求统计
	todayStats, err := h.requestLogRepo.GetStatistics(ctx, "", "", today, time.Now())
	if err != nil {
		logger.Error("failed to get today statistics", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get statistics"})
		return
	}

	// 查询总请求统计
	totalStats, err := h.requestLogRepo.GetStatistics(ctx, "", "", time.Time{}, time.Now())
	if err != nil {
		logger.Error("failed to get total statistics", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get statistics"})
		return
	}

	// 查询项目总数
	projectCount, err := h.db.Collection("projects").CountDocuments(ctx, gin.H{})
	if err != nil {
		logger.Error("failed to count projects", zap.Error(err))
		projectCount = 0
	}

	// 查询规则总数
	ruleCount, err := h.db.Collection("rules").CountDocuments(ctx, gin.H{})
	if err != nil {
		logger.Error("failed to count rules", zap.Error(err))
		ruleCount = 0
	}

	// 查询环境总数
	envCount, err := h.db.Collection("environments").CountDocuments(ctx, gin.H{})
	if err != nil {
		logger.Error("failed to count environments", zap.Error(err))
		envCount = 0
	}

	// 计算成功率
	successRate := 0.0
	if totalStats.TotalRequests > 0 {
		successRate = float64(totalStats.SuccessRequests) / float64(totalStats.TotalRequests) * 100
	}

	// 获取协议分布统计
	protocolDistribution, err := h.getProtocolDistribution(ctx)
	if err != nil {
		logger.Error("failed to get protocol distribution", zap.Error(err))
		// 使用默认值
		protocolDistribution = map[string]int64{
			"http":      0,
			"websocket": 0,
		}
	}

	// 获取 Top 项目统计
	topProjects, err := h.getTopProjects(ctx, 5)
	if err != nil {
		logger.Error("failed to get top projects", zap.Error(err))
		topProjects = []ProjectStats{}
	}

	response := OverviewResponse{
		TotalRequests:        totalStats.TotalRequests,
		TotalProjects:        projectCount,
		TotalRules:           ruleCount,
		TotalEnvironments:    envCount,
		RequestsToday:        todayStats.TotalRequests,
		SuccessRate:          successRate,
		AvgResponseTime:      totalStats.AvgDuration,
		TopProjects:          topProjects,
		ProtocolDistribution: protocolDistribution,
	}

	c.JSON(http.StatusOK, response)
}

// GetRealtime 获取实时统计
// @Summary 获取实时统计
// @Description 获取最近一分钟的实时统计数据
// @Tags Statistics
// @Produce json
// @Success 200 {object} RealtimeResponse
// @Router /api/v1/statistics/realtime [get]
func (h *StatisticsHandler) GetRealtime(c *gin.Context) {
	ctx := c.Request.Context()

	// 最近一分钟
	oneMinuteAgo := time.Now().Add(-1 * time.Minute)

	stats, err := h.requestLogRepo.GetStatistics(ctx, "", "", oneMinuteAgo, time.Now())
	if err != nil {
		logger.Error("failed to get realtime statistics", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get statistics"})
		return
	}

	errorRate := 0.0
	if stats.TotalRequests > 0 {
		errorRate = float64(stats.ErrorRequests) / float64(stats.TotalRequests) * 100
	}

	// 计算成功率
	successRate := 0.0
	if stats.TotalRequests > 0 {
		successRate = float64(stats.SuccessRequests) / float64(stats.TotalRequests) * 100
	}

	// 协议统计（简化版本）
	protocolStats := map[string]ProtocolStat{
		"http": {
			Count:       stats.TotalRequests,
			SuccessRate: successRate,
			AvgDuration: stats.AvgDuration,
		},
		"websocket": {
			Count:       0,
			SuccessRate: 0,
			AvgDuration: 0,
		},
	}

	response := RealtimeResponse{
		Timestamp:         time.Now(),
		RequestsPerMin:    stats.TotalRequests,
		ActiveConnections: 0, // TODO: 从WebSocket管理器获取
		AvgResponseTime:   stats.AvgDuration,
		ErrorRate:         errorRate,
		ProtocolStats:     protocolStats,
	}

	c.JSON(http.StatusOK, response)
}

// GetTrend 获取趋势分析
// @Summary 获取趋势分析
// @Description 获取指定时间段的趋势分析数据
// @Tags Statistics
// @Produce json
// @Param period query string false "时间粒度" Enums(hour, day, week, month) default(day)
// @Param duration query int false "持续天数" default(7)
// @Success 200 {object} TrendResponse
// @Router /api/v1/statistics/trend [get]
func (h *StatisticsHandler) GetTrend(c *gin.Context) {
	ctx := c.Request.Context()

	period := c.DefaultQuery("period", "day")
	duration := 7 // 默认7天

	var startTime time.Time
	var interval time.Duration

	switch period {
	case "hour":
		startTime = time.Now().Add(-24 * time.Hour)
		interval = 1 * time.Hour
	case "day":
		startTime = time.Now().AddDate(0, 0, -duration)
		interval = 24 * time.Hour
	case "week":
		startTime = time.Now().AddDate(0, 0, -4*7) // 4周
		interval = 7 * 24 * time.Hour
	case "month":
		startTime = time.Now().AddDate(0, -12, 0) // 12个月
		interval = 30 * 24 * time.Hour
	default:
		startTime = time.Now().AddDate(0, 0, -duration)
		interval = 24 * time.Hour
	}

	// 生成趋势数据点
	dataPoints := []TrendPoint{}
	currentTime := startTime

	for currentTime.Before(time.Now()) {
		nextTime := currentTime.Add(interval)

		stats, err := h.requestLogRepo.GetStatistics(ctx, "", "", currentTime, nextTime)
		if err != nil {
			logger.Error("failed to get trend statistics",
				zap.Time("start", currentTime),
				zap.Time("end", nextTime),
				zap.Error(err))
		} else {
			successCount := int64(float64(stats.TotalRequests) * float64(stats.SuccessRequests) / float64(stats.TotalRequests))
			if stats.TotalRequests == 0 {
				successCount = 0
			}
			dataPoints = append(dataPoints, TrendPoint{
				Timestamp:    currentTime,
				RequestCount: stats.TotalRequests,
				SuccessCount: successCount,
				ErrorCount:   stats.ErrorRequests,
				AvgDuration:  stats.AvgDuration,
			})
		}

		currentTime = nextTime
	}

	response := TrendResponse{
		Period:     period,
		DataPoints: dataPoints,
	}

	c.JSON(http.StatusOK, response)
}

// GetComparison 获取对比分析
// @Summary 获取对比分析
// @Description 对比当前时段和上一时段的统计数据
// @Tags Statistics
// @Produce json
// @Param period query string false "时段类型" Enums(day, week, month) default(day)
// @Success 200 {object} ComparisonResponse
// @Router /api/v1/statistics/comparison [get]
func (h *StatisticsHandler) GetComparison(c *gin.Context) {
	ctx := c.Request.Context()

	period := c.DefaultQuery("period", "day")

	var currentStart, currentEnd, previousStart, previousEnd time.Time
	now := time.Now()

	switch period {
	case "day":
		currentStart = now.Truncate(24 * time.Hour)
		currentEnd = now
		previousStart = currentStart.AddDate(0, 0, -1)
		previousEnd = currentStart
	case "week":
		// 本周一到现在
		weekday := int(now.Weekday())
		if weekday == 0 {
			weekday = 7 // 周日算作第7天
		}
		currentStart = now.AddDate(0, 0, -(weekday - 1)).Truncate(24 * time.Hour)
		currentEnd = now
		previousStart = currentStart.AddDate(0, 0, -7)
		previousEnd = currentStart
	case "month":
		// 本月1号到现在
		currentStart = time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, now.Location())
		currentEnd = now
		previousStart = currentStart.AddDate(0, -1, 0)
		previousEnd = currentStart
	default:
		currentStart = now.Truncate(24 * time.Hour)
		currentEnd = now
		previousStart = currentStart.AddDate(0, 0, -1)
		previousEnd = currentStart
	}

	// 查询当前时段统计
	currentStats, err := h.requestLogRepo.GetStatistics(ctx, "", "", currentStart, currentEnd)
	if err != nil {
		logger.Error("failed to get current period statistics", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get statistics"})
		return
	}

	// 查询上一时段统计
	previousStats, err := h.requestLogRepo.GetStatistics(ctx, "", "", previousStart, previousEnd)
	if err != nil {
		logger.Error("failed to get previous period statistics", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get statistics"})
		return
	}

	// 计算变化百分比
	requestsChange := 0.0
	if previousStats.TotalRequests > 0 {
		requestsChange = float64(currentStats.TotalRequests-previousStats.TotalRequests) / float64(previousStats.TotalRequests) * 100
	}

	// 计算成功率
	currentSuccessRate := 0.0
	if currentStats.TotalRequests > 0 {
		currentSuccessRate = float64(currentStats.SuccessRequests) / float64(currentStats.TotalRequests) * 100
	}
	previousSuccessRate := 0.0
	if previousStats.TotalRequests > 0 {
		previousSuccessRate = float64(previousStats.SuccessRequests) / float64(previousStats.TotalRequests) * 100
	}
	successRateChange := currentSuccessRate - previousSuccessRate

	durationChange := 0.0
	if previousStats.AvgDuration > 0 {
		durationChange = (currentStats.AvgDuration - previousStats.AvgDuration) / previousStats.AvgDuration * 100
	}

	response := ComparisonResponse{
		CurrentPeriod: PeriodStats{
			StartTime:     currentStart,
			EndTime:       currentEnd,
			TotalRequests: currentStats.TotalRequests,
			SuccessRate:   currentSuccessRate,
			AvgDuration:   currentStats.AvgDuration,
			ErrorCount:    currentStats.ErrorRequests,
		},
		PreviousPeriod: PeriodStats{
			StartTime:     previousStart,
			EndTime:       previousEnd,
			TotalRequests: previousStats.TotalRequests,
			SuccessRate:   previousSuccessRate,
			AvgDuration:   previousStats.AvgDuration,
			ErrorCount:    previousStats.ErrorRequests,
		},
		Changes: ChangeStats{
			RequestsChange:    requestsChange,
			SuccessRateChange: successRateChange,
			DurationChange:    durationChange,
		},
	}

	c.JSON(http.StatusOK, response)
}

// getProtocolDistribution 获取协议分布统计
func (h *StatisticsHandler) getProtocolDistribution(ctx context.Context) (map[string]int64, error) {
	// 使用 MongoDB 聚合查询统计协议分布
	pipeline := []gin.H{
		{
			"$group": gin.H{
				"_id":   "$protocol",
				"count": gin.H{"$sum": 1},
			},
		},
	}

	cursor, err := h.db.Collection("request_logs").Aggregate(ctx, pipeline)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	distribution := make(map[string]int64)
	for cursor.Next(ctx) {
		var result struct {
			ID    string `bson:"_id"`
			Count int64  `bson:"count"`
		}
		if err := cursor.Decode(&result); err != nil {
			return nil, err
		}
		if result.ID != "" {
			distribution[result.ID] = result.Count
		}
	}

	if err := cursor.Err(); err != nil {
		return nil, err
	}

	// 确保基本协议类型存在
	if _, ok := distribution["http"]; !ok {
		distribution["http"] = 0
	}
	if _, ok := distribution["websocket"]; !ok {
		distribution["websocket"] = 0
	}

	return distribution, nil
}

// getTopProjects 获取 Top 项目统计
func (h *StatisticsHandler) getTopProjects(ctx context.Context, limit int) ([]ProjectStats, error) {
	// 使用 MongoDB 聚合查询统计每个项目的请求量
	pipeline := []gin.H{
		{
			"$group": gin.H{
				"_id":           "$project_id",
				"request_count": gin.H{"$sum": 1},
				"success_count": gin.H{
					"$sum": gin.H{
						"$cond": []interface{}{
							gin.H{"$and": []gin.H{
								{"$gte": []interface{}{"$status_code", 200}},
								{"$lt": []interface{}{"$status_code", 400}},
							}},
							1,
							0,
						},
					},
				},
			},
		},
		{
			"$sort": gin.H{"request_count": -1},
		},
		{
			"$limit": limit,
		},
	}

	cursor, err := h.db.Collection("request_logs").Aggregate(ctx, pipeline)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var topProjects []ProjectStats
	for cursor.Next(ctx) {
		var result struct {
			ID           string `bson:"_id"`
			RequestCount int64  `bson:"request_count"`
			SuccessCount int64  `bson:"success_count"`
		}
		if err := cursor.Decode(&result); err != nil {
			return nil, err
		}

		// 计算成功率
		successRate := 0.0
		if result.RequestCount > 0 {
			successRate = float64(result.SuccessCount) / float64(result.RequestCount) * 100
		}

		// 查询项目名称
		projectName := result.ID
		var project struct {
			Name string `bson:"name"`
		}
		err := h.db.Collection("projects").FindOne(ctx, gin.H{"_id": result.ID}).Decode(&project)
		if err == nil {
			projectName = project.Name
		}

		topProjects = append(topProjects, ProjectStats{
			ProjectID:    result.ID,
			ProjectName:  projectName,
			RequestCount: result.RequestCount,
			SuccessRate:  successRate,
		})
	}

	if err := cursor.Err(); err != nil {
		return nil, err
	}

	return topProjects, nil
}
