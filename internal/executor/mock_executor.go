package executor

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"time"

	"github.com/gomockserver/mockserver/internal/adapter"
	"github.com/gomockserver/mockserver/internal/models"
	"github.com/gomockserver/mockserver/pkg/logger"
	"go.uber.org/zap"
)

// MockExecutor Mock 执行器
type MockExecutor struct{}

// NewMockExecutor 创建 Mock 执行器
func NewMockExecutor() *MockExecutor {
	return &MockExecutor{}
}

// Execute 执行 Mock 响应生成
func (e *MockExecutor) Execute(request *adapter.Request, rule *models.Rule) (*adapter.Response, error) {
	// 应用延迟
	if rule.Response.Delay != nil {
		delay := e.calculateDelay(rule.Response.Delay)
		if delay > 0 {
			time.Sleep(time.Duration(delay) * time.Millisecond)
		}
	}

	// 根据响应类型生成响应
	switch rule.Response.Type {
	case models.ResponseTypeStatic:
		return e.staticResponse(request, rule)
	case models.ResponseTypeDynamic:
		// TODO: 阶段三实现
		return nil, fmt.Errorf("dynamic response not implemented yet")
	case models.ResponseTypeScript:
		// TODO: 阶段三实现
		return nil, fmt.Errorf("script response not implemented yet")
	case models.ResponseTypeProxy:
		// TODO: 阶段三实现
		return nil, fmt.Errorf("proxy response not implemented yet")
	default:
		return nil, fmt.Errorf("unsupported response type: %s", rule.Response.Type)
	}
}

// staticResponse 生成静态响应
func (e *MockExecutor) staticResponse(request *adapter.Request, rule *models.Rule) (*adapter.Response, error) {
	if rule.Protocol != models.ProtocolHTTP {
		return nil, fmt.Errorf("only HTTP protocol is supported in static response")
	}

	// 解析 HTTP 响应配置
	contentBytes, err := json.Marshal(rule.Response.Content)
	if err != nil {
		logger.Error("failed to marshal response content", zap.Error(err))
		return nil, err
	}

	var httpResp models.HTTPResponse
	if err := json.Unmarshal(contentBytes, &httpResp); err != nil {
		logger.Error("failed to unmarshal http response", zap.Error(err))
		return nil, err
	}

	// 构建响应体
	var body []byte
	switch httpResp.ContentType {
	case models.ContentTypeJSON:
		body, err = json.Marshal(httpResp.Body)
		if err != nil {
			return nil, err
		}
	case models.ContentTypeText, models.ContentTypeHTML, models.ContentTypeXML:
		if str, ok := httpResp.Body.(string); ok {
			body = []byte(str)
		} else {
			body, err = json.Marshal(httpResp.Body)
			if err != nil {
				return nil, err
			}
		}
	case models.ContentTypeBinary:
		// TODO: 处理二进制数据
		body = []byte{}
	default:
		body, err = json.Marshal(httpResp.Body)
		if err != nil {
			return nil, err
		}
	}

	// 设置默认 Content-Type
	if httpResp.Headers == nil {
		httpResp.Headers = make(map[string]string)
	}
	if _, ok := httpResp.Headers["Content-Type"]; !ok {
		httpResp.Headers["Content-Type"] = e.getDefaultContentType(httpResp.ContentType)
	}

	// 构建统一响应模型
	response := &adapter.Response{
		StatusCode: httpResp.StatusCode,
		Headers:    httpResp.Headers,
		Body:       body,
		Metadata:   make(map[string]interface{}),
	}

	return response, nil
}

// calculateDelay 计算延迟时间（毫秒）
func (e *MockExecutor) calculateDelay(config *models.DelayConfig) int {
	switch config.Type {
	case "fixed":
		return config.Fixed
	case "random":
		if config.Max <= config.Min {
			return config.Min
		}
		return config.Min + rand.Intn(config.Max-config.Min)
	case "normal":
		// TODO: 实现正态分布延迟
		return config.Mean
	case "step":
		// TODO: 实现阶梯延迟
		return config.Fixed
	default:
		return 0
	}
}

// getDefaultContentType 获取默认 Content-Type
func (e *MockExecutor) getDefaultContentType(contentType models.ContentType) string {
	switch contentType {
	case models.ContentTypeJSON:
		return "application/json"
	case models.ContentTypeXML:
		return "application/xml"
	case models.ContentTypeHTML:
		return "text/html"
	case models.ContentTypeText:
		return "text/plain"
	case models.ContentTypeBinary:
		return "application/octet-stream"
	default:
		return "application/json"
	}
}

// GetDefaultResponse 获取默认 404 响应
func (e *MockExecutor) GetDefaultResponse() *adapter.Response {
	return &adapter.Response{
		StatusCode: 404,
		Headers: map[string]string{
			"Content-Type": "application/json",
		},
		Body: []byte(`{"error": "No matching rule found"}`),
	}
}
