package executor

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"math/rand"
	"net/http"
	"time"

	"github.com/gomockserver/mockserver/internal/adapter"
	"github.com/gomockserver/mockserver/pkg/logger"
	"go.uber.org/zap"
)

// ProxyConfig 代理配置
type ProxyConfig struct {
	TargetURL       string                 `json:"target_url"`
	Timeout         int                    `json:"timeout"`          // 超时时间（秒）
	ModifyRequest   *RequestModifier       `json:"modify_request"`   // 请求修改器
	ModifyResponse  *ResponseModifier      `json:"modify_response"`  // 响应修改器
	InjectDelay     int                    `json:"inject_delay"`     // 注入延迟（毫秒）
	ErrorRate       float64                `json:"error_rate"`       // 错误率（0-1）
	ErrorStatusCode int                    `json:"error_status_code"` // 错误状态码
	FollowRedirect  bool                   `json:"follow_redirect"`  // 是否跟随重定向
}

// RequestModifier 请求修改器
type RequestModifier struct {
	Headers map[string]string      `json:"headers"`  // 添加/修改的请求头
	Query   map[string]string      `json:"query"`    // 添加/修改的查询参数
	Body    map[string]interface{} `json:"body"`     // 修改请求体（仅JSON）
	RemoveHeaders []string           `json:"remove_headers"` // 移除的请求头
}

// ResponseModifier 响应修改器
type ResponseModifier struct {
	Headers      map[string]string      `json:"headers"`       // 添加/修改的响应头
	BodyReplace  map[string]interface{} `json:"body_replace"`  // 替换响应体字段
	StatusCode   int                    `json:"status_code"`   // 修改状态码
	RemoveHeaders []string              `json:"remove_headers"` // 移除的响应头
}

// ProxyExecutor 代理执行器
type ProxyExecutor struct {
	client *http.Client
}

// NewProxyExecutor 创建代理执行器
func NewProxyExecutor() *ProxyExecutor {
	return &ProxyExecutor{
		client: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// Execute 执行代理请求
func (p *ProxyExecutor) Execute(request *adapter.Request, config *ProxyConfig) (*adapter.Response, error) {
	// 检查是否应该模拟错误
	if config.ErrorRate > 0 {
		if shouldInjectError(config.ErrorRate) {
			statusCode := config.ErrorStatusCode
			if statusCode == 0 {
				statusCode = 500 // 默认500错误
			}
			logger.Info("injecting error response",
				zap.Float64("error_rate", config.ErrorRate),
				zap.Int("status_code", statusCode))
			
			return &adapter.Response{
				StatusCode: statusCode,
				Headers: map[string]string{
					"Content-Type": "application/json",
				},
				Body: []byte(fmt.Sprintf(`{"error": "injected error", "status_code": %d}`, statusCode)),
			}, nil
		}
	}

	// 注入延迟
	if config.InjectDelay > 0 {
		time.Sleep(time.Duration(config.InjectDelay) * time.Millisecond)
	}

	// 构建目标URL
	targetURL := config.TargetURL
	if request.Path != "" {
		targetURL = targetURL + request.Path
	}

	// 提取HTTP方法
	method := "GET"
	if request.Metadata != nil {
		if m, ok := request.Metadata["method"].(string); ok {
			method = m
		}
	}

	// 创建HTTP请求
	var reqBody io.Reader
	if len(request.Body) > 0 {
		// 如果需要修改请求体
		if config.ModifyRequest != nil && config.ModifyRequest.Body != nil {
			modifiedBody, err := p.modifyRequestBody(request.Body, config.ModifyRequest.Body)
			if err != nil {
				logger.Error("failed to modify request body", zap.Error(err))
				return nil, err
			}
			reqBody = bytes.NewReader(modifiedBody)
		} else {
			reqBody = bytes.NewReader(request.Body)
		}
	}

	httpReq, err := http.NewRequest(method, targetURL, reqBody)
	if err != nil {
		logger.Error("failed to create proxy request", zap.Error(err))
		return nil, err
	}

	// 复制原始请求头
	for key, value := range request.Headers {
		httpReq.Header.Set(key, value)
	}

	// 应用请求修改器
	if config.ModifyRequest != nil {
		p.applyRequestModifier(httpReq, config.ModifyRequest)
	}

	// 设置超时
	if config.Timeout > 0 {
		p.client.Timeout = time.Duration(config.Timeout) * time.Second
	}

	// 设置是否跟随重定向
	if !config.FollowRedirect {
		p.client.CheckRedirect = func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		}
	}

	// 执行请求
	logger.Info("proxying request",
		zap.String("method", method),
		zap.String("target_url", targetURL))
	
	resp, err := p.client.Do(httpReq)
	if err != nil {
		logger.Error("failed to proxy request", zap.Error(err))
		return nil, err
	}
	defer resp.Body.Close()

	// 读取响应体
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		logger.Error("failed to read proxy response", zap.Error(err))
		return nil, err
	}

	// 构建响应
	response := &adapter.Response{
		StatusCode: resp.StatusCode,
		Headers:    make(map[string]string),
		Body:       respBody,
		Metadata:   make(map[string]interface{}),
	}

	// 复制响应头
	for key, values := range resp.Header {
		if len(values) > 0 {
			response.Headers[key] = values[0]
		}
	}

	// 应用响应修改器
	if config.ModifyResponse != nil {
		err = p.applyResponseModifier(response, config.ModifyResponse)
		if err != nil {
			logger.Error("failed to apply response modifier", zap.Error(err))
			return nil, err
		}
	}

	return response, nil
}

// modifyRequestBody 修改请求体
func (p *ProxyExecutor) modifyRequestBody(originalBody []byte, modifications map[string]interface{}) ([]byte, error) {
	// 解析原始body
	var bodyMap map[string]interface{}
	if err := json.Unmarshal(originalBody, &bodyMap); err != nil {
		// 如果不是JSON，返回原始body
		return originalBody, nil
	}

	// 应用修改
	for key, value := range modifications {
		bodyMap[key] = value
	}

	// 重新序列化
	return json.Marshal(bodyMap)
}

// applyRequestModifier 应用请求修改器
func (p *ProxyExecutor) applyRequestModifier(req *http.Request, modifier *RequestModifier) {
	// 添加/修改请求头
	if modifier.Headers != nil {
		for key, value := range modifier.Headers {
			req.Header.Set(key, value)
		}
	}

	// 移除请求头
	if modifier.RemoveHeaders != nil {
		for _, key := range modifier.RemoveHeaders {
			req.Header.Del(key)
		}
	}

	// 修改查询参数
	if modifier.Query != nil {
		q := req.URL.Query()
		for key, value := range modifier.Query {
			q.Set(key, value)
		}
		req.URL.RawQuery = q.Encode()
	}
}

// applyResponseModifier 应用响应修改器
func (p *ProxyExecutor) applyResponseModifier(resp *adapter.Response, modifier *ResponseModifier) error {
	// 修改状态码
	if modifier.StatusCode > 0 {
		resp.StatusCode = modifier.StatusCode
	}

	// 添加/修改响应头
	if modifier.Headers != nil {
		for key, value := range modifier.Headers {
			resp.Headers[key] = value
		}
	}

	// 移除响应头
	if modifier.RemoveHeaders != nil {
		for _, key := range modifier.RemoveHeaders {
			delete(resp.Headers, key)
		}
	}

	// 替换响应体字段（仅JSON）
	if modifier.BodyReplace != nil && len(modifier.BodyReplace) > 0 {
		var bodyMap map[string]interface{}
		if err := json.Unmarshal(resp.Body, &bodyMap); err == nil {
			// 应用替换
			for key, value := range modifier.BodyReplace {
				bodyMap[key] = value
			}
			// 重新序列化
			modifiedBody, err := json.Marshal(bodyMap)
			if err != nil {
				return err
			}
			resp.Body = modifiedBody
		}
	}

	return nil
}

// shouldInjectError 判断是否应该注入错误
func shouldInjectError(errorRate float64) bool {
	if errorRate <= 0 {
		return false
	}
	if errorRate >= 1.0 {
		return true
	}
	return rand.Float64() < errorRate
}
