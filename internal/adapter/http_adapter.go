package adapter

import (
	"io"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gomockserver/mockserver/internal/models"
	"github.com/google/uuid"
)

// HTTPAdapter HTTP 协议适配器
type HTTPAdapter struct{}

// NewHTTPAdapter 创建 HTTP 适配器
func NewHTTPAdapter() *HTTPAdapter {
	return &HTTPAdapter{}
}

// Parse 解析 HTTP 请求为统一模型
func (a *HTTPAdapter) Parse(rawRequest interface{}) (*Request, error) {
	c, ok := rawRequest.(*gin.Context)
	if !ok {
		return nil, nil
	}

	// 生成请求ID
	requestID := uuid.New().String()

	// 读取请求体
	body, err := io.ReadAll(c.Request.Body)
	if err != nil {
		return nil, err
	}

	// 提取Headers
	headers := make(map[string]string)
	for key, values := range c.Request.Header {
		if len(values) > 0 {
			headers[key] = values[0]
		}
	}

	// 提取 Query 参数
	query := make(map[string]string)
	for key, values := range c.Request.URL.Query() {
		if len(values) > 0 {
			query[key] = values[0]
		}
	}

	// 获取客户端IP
	clientIP := c.ClientIP()

	// 获取实际的API路径（移除 /:projectID/:environmentID 前缀）
	// Gin的路由格式：/:projectID/:environmentID/*path
	// c.Param("path") 会返回包含前导斜杠的路径，如 "/api/users/1"
	actualPath := c.Param("path")
	if actualPath == "" {
		// 如果没有path参数，使用完整路径
		actualPath = c.Request.URL.Path
	}

	// 创建统一请求模型
	request := &Request{
		ID:         requestID,
		Protocol:   models.ProtocolHTTP,
		Path:       actualPath,
		Headers:    headers,
		Body:       body,
		SourceIP:   clientIP,
		ReceivedAt: time.Now(),
		Metadata: map[string]interface{}{
			"method":       c.Request.Method,
			"query":        query,
			"raw_query":    c.Request.URL.RawQuery,
			"host":         c.Request.Host,
			"user_agent":   c.Request.UserAgent(),
			"content_type": c.ContentType(),
		},
	}

	return request, nil
}

// Build 构建 HTTP 响应
func (a *HTTPAdapter) Build(response *Response) (interface{}, error) {
	// 返回响应配置，由调用方设置到 gin.Context
	return response, nil
}

// WriteResponse 将响应写入 gin.Context
func (a *HTTPAdapter) WriteResponse(c *gin.Context, response *Response) {
	// 设置响应头
	for key, value := range response.Headers {
		c.Header(key, value)
	}

	// 设置状态码和响应体
	c.Data(response.StatusCode, getContentType(response.Headers), response.Body)
}

// getContentType 从响应头中获取 Content-Type
func getContentType(headers map[string]string) string {
	for key, value := range headers {
		if strings.ToLower(key) == "content-type" {
			return value
		}
	}
	return "application/json"
}
