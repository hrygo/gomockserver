package middleware

import (
	"bytes"
	"encoding/json"
	"io"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gomockserver/mockserver/internal/models"
	"github.com/gomockserver/mockserver/internal/repository"
	"github.com/gomockserver/mockserver/pkg/logger"
	"go.uber.org/zap"
)

// RequestLoggerMiddleware 请求日志记录中间件
type RequestLoggerMiddleware struct {
	repo    repository.RequestLogRepository
	enabled bool
}

// NewRequestLoggerMiddleware 创建请求日志中间件
func NewRequestLoggerMiddleware(repo repository.RequestLogRepository) *RequestLoggerMiddleware {
	return &RequestLoggerMiddleware{
		repo:    repo,
		enabled: true,
	}
}

// SetEnabled 设置是否启用
func (m *RequestLoggerMiddleware) SetEnabled(enabled bool) {
	m.enabled = enabled
}

// responseWriter 自定义 ResponseWriter 用于捕获响应
type responseWriter struct {
	gin.ResponseWriter
	body *bytes.Buffer
}

func (w *responseWriter) Write(b []byte) (int, error) {
	w.body.Write(b)
	return w.ResponseWriter.Write(b)
}

func (w *responseWriter) WriteString(s string) (int, error) {
	w.body.WriteString(s)
	return w.ResponseWriter.WriteString(s)
}

// Handler 日志记录处理器
func (m *RequestLoggerMiddleware) Handler() gin.HandlerFunc {
	return func(c *gin.Context) {
		if !m.enabled {
			c.Next()
			return
		}

		// 记录开始时间
		startTime := time.Now()

		// 读取请求体
		var requestBody []byte
		if c.Request.Body != nil {
			requestBody, _ = io.ReadAll(c.Request.Body)
			// 重新设置请求体，供后续处理使用
			c.Request.Body = io.NopCloser(bytes.NewBuffer(requestBody))
		}

		// 创建自定义 ResponseWriter
		blw := &responseWriter{
			ResponseWriter: c.Writer,
			body:           bytes.NewBufferString(""),
		}
		c.Writer = blw

		// 处理请求
		c.Next()

		// 计算耗时
		duration := time.Since(startTime).Milliseconds()

		// 从上下文获取项目和环境信息
		projectID, _ := c.Get("project_id")
		environmentID, _ := c.Get("environment_id")
		ruleID, _ := c.Get("rule_id")
		requestID, _ := c.Get("request_id")

		// 构建请求日志
		requestLog := &models.RequestLog{
			RequestID:     getStringValue(requestID),
			ProjectID:     getStringValue(projectID),
			EnvironmentID: getStringValue(environmentID),
			RuleID:        getStringValue(ruleID),
			Protocol:      m.getProtocol(c),
			Method:        c.Request.Method,
			Path:          c.Request.URL.Path,
			StatusCode:    c.Writer.Status(),
			Duration:      duration,
			SourceIP:      c.ClientIP(),
			Timestamp:     startTime,
			Request:       m.buildRequestData(c, requestBody),
			Response:      m.buildResponseData(c, blw.body.Bytes()),
		}

		// 异步保存日志（避免阻塞请求）
		go func() {
			ctx := c.Request.Context()
			if err := m.repo.Create(ctx, requestLog); err != nil {
				logger.Error("failed to save request log",
					zap.String("request_id", requestLog.RequestID),
					zap.Error(err))
			}
		}()
	}
}

// getProtocol 获取协议类型
func (m *RequestLoggerMiddleware) getProtocol(c *gin.Context) models.ProtocolType {
	// 检查是否是 WebSocket 升级请求
	if c.Request.Header.Get("Upgrade") == "websocket" {
		return models.ProtocolWebSocket
	}
	return models.ProtocolHTTP
}

// buildRequestData 构建请求数据
func (m *RequestLoggerMiddleware) buildRequestData(c *gin.Context, body []byte) map[string]interface{} {
	data := make(map[string]interface{})

	// 基本信息
	data["method"] = c.Request.Method
	data["path"] = c.Request.URL.Path
	data["query"] = c.Request.URL.RawQuery
	data["headers"] = m.sanitizeHeaders(c.Request.Header)

	// 请求体
	if len(body) > 0 {
		// 尝试解析为 JSON
		var jsonBody interface{}
		if err := json.Unmarshal(body, &jsonBody); err == nil {
			data["body"] = jsonBody
		} else {
			// 如果不是 JSON，限制长度存储为字符串
			bodyStr := string(body)
			if len(bodyStr) > 1000 {
				bodyStr = bodyStr[:1000] + "... (truncated)"
			}
			data["body"] = bodyStr
		}
	}

	return data
}

// buildResponseData 构建响应数据
func (m *RequestLoggerMiddleware) buildResponseData(c *gin.Context, body []byte) map[string]interface{} {
	data := make(map[string]interface{})

	data["status_code"] = c.Writer.Status()
	data["headers"] = m.sanitizeHeaders(c.Writer.Header())

	// 响应体
	if len(body) > 0 {
		// 尝试解析为 JSON
		var jsonBody interface{}
		if err := json.Unmarshal(body, &jsonBody); err == nil {
			data["body"] = jsonBody
		} else {
			// 如果不是 JSON，限制长度存储为字符串
			bodyStr := string(body)
			if len(bodyStr) > 1000 {
				bodyStr = bodyStr[:1000] + "... (truncated)"
			}
			data["body"] = bodyStr
		}
	}

	return data
}

// sanitizeHeaders 清理敏感头信息
func (m *RequestLoggerMiddleware) sanitizeHeaders(headers map[string][]string) map[string]string {
	result := make(map[string]string)
	sensitiveHeaders := map[string]bool{
		"authorization": true,
		"cookie":        true,
		"set-cookie":    true,
		"x-api-key":     true,
	}

	for key, values := range headers {
		lowerKey := strings.ToLower(key)
		if sensitiveHeaders[lowerKey] {
			result[key] = "***REDACTED***"
		} else if len(values) > 0 {
			result[key] = values[0]
		}
	}

	return result
}

// getStringValue 安全获取字符串值
func getStringValue(v interface{}) string {
	if v == nil {
		return ""
	}
	if str, ok := v.(string); ok {
		return str
	}
	return ""
}
