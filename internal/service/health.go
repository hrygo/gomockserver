package service

import (
	"context"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gomockserver/mockserver/pkg/logger"
	"go.mongodb.org/mongo-driver/mongo"
	"go.uber.org/zap"
)

var (
	// ServerStartTime 服务器启动时间
	ServerStartTime time.Time
	// Version 应用版本号
	Version = "0.1.1"
	// AppName 应用名称
	AppName = "MockServer"
)

func init() {
	ServerStartTime = time.Now()
}

// HealthStatus 健康状态
type HealthStatus string

const (
	// StatusHealthy 健康
	StatusHealthy HealthStatus = "healthy"
	// StatusDegraded 降级（部分功能不可用）
	StatusDegraded HealthStatus = "degraded"
	// StatusUnhealthy 不健康
	StatusUnhealthy HealthStatus = "unhealthy"
)

// ComponentStatus 组件状态
type ComponentStatus struct {
	Status  HealthStatus `json:"status"`
	Message string       `json:"message,omitempty"`
	Details interface{}  `json:"details,omitempty"`
}

// HealthResponse 健康检查响应
type HealthResponse struct {
	Status     HealthStatus               `json:"status"`
	Version    string                     `json:"version"`
	AppName    string                     `json:"app_name"`
	Uptime     string                     `json:"uptime"`
	Timestamp  string                     `json:"timestamp"`
	Components map[string]ComponentStatus `json:"components,omitempty"`
}

// HealthChecker 健康检查器
type HealthChecker struct {
	mongoClient *mongo.Client
}

// NewHealthChecker 创建健康检查器
func NewHealthChecker(mongoClient *mongo.Client) *HealthChecker {
	return &HealthChecker{
		mongoClient: mongoClient,
	}
}

// Check 执行健康检查
func (h *HealthChecker) Check(c *gin.Context) {
	ctx := c.Request.Context()
	detailed := c.Query("detailed") == "true"

	response := HealthResponse{
		Status:    StatusHealthy,
		Version:   Version,
		AppName:   AppName,
		Uptime:    formatUptime(time.Since(ServerStartTime)),
		Timestamp: time.Now().UTC().Format(time.RFC3339),
	}

	// 如果需要详细信息，检查各组件状态
	if detailed && h.mongoClient != nil {
		components := make(map[string]ComponentStatus)

		// 检查数据库连接
		dbStatus := h.checkDatabase(ctx)
		components["database"] = dbStatus

		// 如果数据库不健康，整体状态降级
		if dbStatus.Status == StatusUnhealthy {
			response.Status = StatusDegraded
		}

		response.Components = components
	}

	// 根据整体状态返回不同的 HTTP 状态码
	statusCode := 200
	if response.Status == StatusUnhealthy {
		statusCode = 503
	} else if response.Status == StatusDegraded {
		statusCode = 200 // 降级时仍然返回 200，但在响应中标记
	}

	c.JSON(statusCode, response)
}

// checkDatabase 检查数据库连接状态
func (h *HealthChecker) checkDatabase(ctx context.Context) ComponentStatus {
	if h.mongoClient == nil {
		return ComponentStatus{
			Status:  StatusHealthy,
			Message: "database not configured",
		}
	}

	// 使用超时上下文进行 ping
	pingCtx, cancel := context.WithTimeout(ctx, 2*time.Second)
	defer cancel()

	err := h.mongoClient.Ping(pingCtx, nil)
	if err != nil {
		logger.Error("database health check failed", zap.Error(err))
		return ComponentStatus{
			Status:  StatusUnhealthy,
			Message: "database connection failed",
			Details: map[string]interface{}{
				"error": err.Error(),
			},
		}
	}

	return ComponentStatus{
		Status:  StatusHealthy,
		Message: "database connection established",
	}
}

// formatUptime 格式化运行时间
func formatUptime(d time.Duration) string {
	days := int(d.Hours()) / 24
	hours := int(d.Hours()) % 24
	minutes := int(d.Minutes()) % 60
	seconds := int(d.Seconds()) % 60

	if days > 0 {
		return formatString("%dd %dh %dm %ds", days, hours, minutes, seconds)
	}
	if hours > 0 {
		return formatString("%dh %dm %ds", hours, minutes, seconds)
	}
	if minutes > 0 {
		return formatString("%dm %ds", minutes, seconds)
	}
	return formatString("%ds", seconds)
}

func formatString(format string, args ...interface{}) string {
	// 简单格式化实现
	switch len(args) {
	case 1:
		return formatWithOneArg(format, args[0].(int))
	case 2:
		return formatWithTwoArgs(format, args[0].(int), args[1].(int))
	case 3:
		return formatWithThreeArgs(format, args[0].(int), args[1].(int), args[2].(int))
	case 4:
		return formatWithFourArgs(format, args[0].(int), args[1].(int), args[2].(int), args[3].(int))
	default:
		return ""
	}
}

func formatWithOneArg(format string, a int) string {
	// %ds 格式
	result := ""
	for i := 0; i < len(format); i++ {
		if format[i] == '%' && i+1 < len(format) && format[i+1] == 'd' {
			result += intToString(a)
			i++
		} else {
			result += string(format[i])
		}
	}
	return result
}

func formatWithTwoArgs(format string, a, b int) string {
	result := ""
	args := []int{a, b}
	argIdx := 0
	for i := 0; i < len(format); i++ {
		if format[i] == '%' && i+1 < len(format) && format[i+1] == 'd' && argIdx < len(args) {
			result += intToString(args[argIdx])
			argIdx++
			i++
		} else {
			result += string(format[i])
		}
	}
	return result
}

func formatWithThreeArgs(format string, a, b, c int) string {
	result := ""
	args := []int{a, b, c}
	argIdx := 0
	for i := 0; i < len(format); i++ {
		if format[i] == '%' && i+1 < len(format) && format[i+1] == 'd' && argIdx < len(args) {
			result += intToString(args[argIdx])
			argIdx++
			i++
		} else {
			result += string(format[i])
		}
	}
	return result
}

func formatWithFourArgs(format string, a, b, c, d int) string {
	result := ""
	args := []int{a, b, c, d}
	argIdx := 0
	for i := 0; i < len(format); i++ {
		if format[i] == '%' && i+1 < len(format) && format[i+1] == 'd' && argIdx < len(args) {
			result += intToString(args[argIdx])
			argIdx++
			i++
		} else {
			result += string(format[i])
		}
	}
	return result
}

func intToString(n int) string {
	if n == 0 {
		return "0"
	}
	var result string
	for n > 0 {
		digit := n % 10
		result = string(rune('0'+digit)) + result
		n /= 10
	}
	return result
}
