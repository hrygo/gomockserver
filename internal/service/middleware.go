package service

import (
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gomockserver/mockserver/pkg/logger"
	"go.uber.org/zap"
)

const (
	// RequestIDHeader 请求ID头部名称
	RequestIDHeader = "X-Request-ID"
	// RequestIDKey 请求ID在上下文中的key
	RequestIDKey = "request_id"
)

// RequestIDMiddleware 请求追踪中间件
// 为每个请求生成唯一的 request_id，并在整个调用链路中传递
func RequestIDMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 尝试从请求头获取 request_id
		requestID := c.GetHeader(RequestIDHeader)

		// 如果请求头中没有，则生成一个新的
		if requestID == "" {
			requestID = generateRequestID()
		}

		// 将 request_id 存储到上下文中
		c.Set(RequestIDKey, requestID)

		// 在响应头中返回 request_id
		c.Header(RequestIDHeader, requestID)

		c.Next()
	}
}

// PerformanceMiddleware 性能监控中间件
// 记录请求处理时长和基本信息
func PerformanceMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 记录开始时间
		startTime := time.Now()
		path := c.Request.URL.Path
		method := c.Request.Method

		// 获取 request_id
		requestID, _ := c.Get(RequestIDKey)

		// 处理请求
		c.Next()

		// 计算耗时
		duration := time.Since(startTime)
		statusCode := c.Writer.Status()

		// 记录日志
		logger.Info("request completed",
			zap.String("request_id", requestID.(string)),
			zap.String("method", method),
			zap.String("path", path),
			zap.Int("status", statusCode),
			zap.Duration("duration", duration),
			zap.String("client_ip", c.ClientIP()),
		)

		// 如果请求耗时过长（超过1秒），记录警告
		if duration > time.Second {
			logger.Warn("slow request detected",
				zap.String("request_id", requestID.(string)),
				zap.String("method", method),
				zap.String("path", path),
				zap.Duration("duration", duration),
			)
		}
	}
}

// LoggingMiddleware 日志中间件
// 记录请求的基本信息
func LoggingMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		requestID, exists := c.Get(RequestIDKey)
		if !exists {
			requestID = "unknown"
		}

		logger.Debug("incoming request",
			zap.String("request_id", requestID.(string)),
			zap.String("method", c.Request.Method),
			zap.String("path", c.Request.URL.Path),
			zap.String("client_ip", c.ClientIP()),
			zap.String("user_agent", c.Request.UserAgent()),
		)

		c.Next()
	}
}

// generateRequestID 生成唯一的请求ID
func generateRequestID() string {
	// 使用时间戳 + 随机数生成简单的 request_id
	// 生产环境建议使用 UUID 或其他更强的唯一性保证
	timestamp := time.Now().UnixNano()
	return "req-" + int64ToString(timestamp)
}

// int64ToString 将 int64 转换为字符串
func int64ToString(n int64) string {
	if n == 0 {
		return "0"
	}

	isNegative := n < 0
	if isNegative {
		n = -n
	}

	var result string
	for n > 0 {
		digit := n % 10
		result = string(rune('0'+digit)) + result
		n /= 10
	}

	if isNegative {
		result = "-" + result
	}

	return result
}
