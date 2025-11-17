package executor

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"math"
	"math/rand"
	"os"
	"sync"
	"time"

	"github.com/gomockserver/mockserver/internal/adapter"
	"github.com/gomockserver/mockserver/internal/models"
	"github.com/gomockserver/mockserver/pkg/logger"
	"go.uber.org/zap"
)

// MockExecutor Mock 执行器
type MockExecutor struct {
	normalRandMu sync.Mutex
	normalRandS  float64
	normalRandV  float64
	normalRandOK bool

	// 阶梯延迟相关字段
	stepCounters   map[string]int64
	stepCountersMu sync.RWMutex

	// 模板引擎
	templateEngine *TemplateEngine

	// 代理执行器
	proxyExecutor *ProxyExecutor
}

// NewMockExecutor 创建 Mock 执行器
func NewMockExecutor() *MockExecutor {
	return &MockExecutor{
		stepCounters:   make(map[string]int64),
		templateEngine: NewTemplateEngine(),
		proxyExecutor:  NewProxyExecutor(),
	}
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
		return e.dynamicResponse(request, rule, nil)
	case models.ResponseTypeScript:
		// TODO: v0.4.0 实现
		return nil, fmt.Errorf("script response not implemented yet")
	case models.ResponseTypeProxy:
		return e.proxyResponse(request, rule)
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

	// 检查是否使用文件路径引用
	if bodyMap, ok := httpResp.Body.(map[string]interface{}); ok {
		if filePath, ok := bodyMap["file_path"].(string); ok {
			// 从文件读取
			body, err = e.readFileResponse(filePath)
			if err != nil {
				logger.Error("failed to read file", zap.String("file_path", filePath), zap.Error(err))
				return nil, fmt.Errorf("failed to read file %s: %w", filePath, err)
			}
		} else {
			// 使用map内容
			body, err = json.Marshal(httpResp.Body)
			if err != nil {
				return nil, err
			}
		}
	} else {
		// 使用内嵌内容
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
			// 处理二进制数据 - 支持Base64编码
			if str, ok := httpResp.Body.(string); ok {
				// 尝试Base64解码
				decoded, err := base64.StdEncoding.DecodeString(str)
				if err != nil {
					// 如果解码失败，记录警告并返回原始数据
					logger.Warn("failed to decode base64 binary data, returning raw data", zap.Error(err))
					body = []byte(str)
				} else {
					body = decoded
				}
			} else {
				// 非字符串类型，尝试JSON序列化
				body, err = json.Marshal(httpResp.Body)
				if err != nil {
					return nil, fmt.Errorf("failed to marshal binary body: %w", err)
				}
			}
		default:
			body, err = json.Marshal(httpResp.Body)
			if err != nil {
				return nil, err
			}
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

// dynamicResponse 生成动态响应
func (e *MockExecutor) dynamicResponse(request *adapter.Request, rule *models.Rule, env *models.Environment) (*adapter.Response, error) {
	if rule.Protocol != models.ProtocolHTTP {
		return nil, fmt.Errorf("only HTTP protocol is supported in dynamic response")
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

	// 构建模板上下文
	ctx := e.templateEngine.BuildContext(request, rule, env)

	// 渲染响应体
	var body []byte
	switch httpResp.ContentType {
	case models.ContentTypeJSON:
		// JSON模板渲染
		rendered, err := e.templateEngine.RenderJSON(httpResp.Body, ctx)
		if err != nil {
			logger.Error("failed to render json template", zap.Error(err))
			return nil, fmt.Errorf("failed to render json template: %w", err)
		}
		body, err = json.Marshal(rendered)
		if err != nil {
			return nil, err
		}
	case models.ContentTypeText, models.ContentTypeHTML, models.ContentTypeXML:
		// 文本模板渲染
		var templateStr string
		if str, ok := httpResp.Body.(string); ok {
			templateStr = str
		} else {
			// 如果不是字符串，先转为JSON
			tempBytes, err := json.Marshal(httpResp.Body)
			if err != nil {
				return nil, err
			}
			templateStr = string(tempBytes)
		}

		rendered, err := e.templateEngine.Render(templateStr, ctx)
		if err != nil {
			logger.Error("failed to render text template", zap.Error(err))
			return nil, fmt.Errorf("failed to render text template: %w", err)
		}
		body = []byte(rendered)
	default:
		// 默认JSON处理
		rendered, err := e.templateEngine.RenderJSON(httpResp.Body, ctx)
		if err != nil {
			logger.Error("failed to render template", zap.Error(err))
			return nil, fmt.Errorf("failed to render template: %w", err)
		}
		body, err = json.Marshal(rendered)
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

// proxyResponse 生成代理响应
func (e *MockExecutor) proxyResponse(request *adapter.Request, rule *models.Rule) (*adapter.Response, error) {
	// 解析代理配置
	contentBytes, err := json.Marshal(rule.Response.Content)
	if err != nil {
		logger.Error("failed to marshal proxy config", zap.Error(err))
		return nil, err
	}

	var proxyConfig ProxyConfig
	if err := json.Unmarshal(contentBytes, &proxyConfig); err != nil {
		logger.Error("failed to unmarshal proxy config", zap.Error(err))
		return nil, err
	}

	// 执行代理请求
	return e.proxyExecutor.Execute(request, &proxyConfig)
}

// calculateDelay 计算延迟时间（毫秒）
func (e *MockExecutor) calculateDelay(config *models.DelayConfig) int {
	if config == nil {
		return 0
	}

	switch config.Type {
	case "fixed":
		return config.Fixed
	case "random":
		if config.Max <= config.Min {
			return config.Min
		}
		return config.Min + rand.Intn(config.Max-config.Min)
	case "normal":
		// 实现正态分布延迟 - 使用Marsaglia polar method
		if config.StdDev <= 0 {
			// 标准差必须为正数
			return config.Mean
		}

		// 生成正态分布随机数
		normalRand := e.generateNormalRand(float64(config.Mean), float64(config.StdDev))

		// 确保结果为非负整数
		result := int(math.Round(normalRand))
		if result < 0 {
			result = 0
		}

		return result
	case "step":
		// 实现阶梯延迟 - 基于请求计数的阶梯延迟算法
		return e.calculateStepDelay(config, "")
	default:
		return 0
	}
}

// calculateStepDelay 计算阶梯延迟
func (e *MockExecutor) calculateStepDelay(config *models.DelayConfig, ruleID string) int {
	// 使用规则ID作为计数器键，实现计数器隔离
	counterKey := "default"
	if ruleID != "" {
		counterKey = ruleID
	}

	// 增加计数器
	e.stepCountersMu.Lock()
	e.stepCounters[counterKey]++
	count := e.stepCounters[counterKey]
	e.stepCountersMu.Unlock()

	// 计算阶梯延迟
	baseDelay := config.Fixed
	step := config.Step
	limit := config.Limit

	if step <= 0 {
		// 步长必须为正数
		return baseDelay
	}

	// 计算阶梯值: baseDelay + (count-1) * step
	delay := baseDelay + int(count-1)*step

	// 如果设置了上限，则不超过上限
	if limit > 0 && delay > limit {
		delay = limit
	}

	return delay
}

// ResetStepCounter 重置阶梯延迟计数器
func (e *MockExecutor) ResetStepCounter(ruleID string) {
	e.stepCountersMu.Lock()
	defer e.stepCountersMu.Unlock()

	if ruleID == "" {
		// 重置所有计数器
		e.stepCounters = make(map[string]int64)
	} else {
		// 重置特定规则的计数器
		delete(e.stepCounters, ruleID)
	}

	logger.Info("reset step delay counter", zap.String("rule_id", ruleID))
}

// GetStepCounter 获取阶梯延迟计数器值
func (e *MockExecutor) GetStepCounter(ruleID string) int64 {
	e.stepCountersMu.RLock()
	defer e.stepCountersMu.RUnlock()

	counterKey := "default"
	if ruleID != "" {
		counterKey = ruleID
	}

	return e.stepCounters[counterKey]
}

// generateNormalRand 使用Marsaglia polar method生成正态分布随机数
func (e *MockExecutor) generateNormalRand(mean, stdDev float64) float64 {
	e.normalRandMu.Lock()
	defer e.normalRandMu.Unlock()

	// 如果有缓存的随机数，直接使用
	if e.normalRandOK {
		e.normalRandOK = false
		return mean + stdDev*e.normalRandV
	}

	// 生成新的正态分布随机数对
	for {
		// 生成[-1, 1]范围内的均匀分布随机数
		u := 2.0*rand.Float64() - 1.0
		v := 2.0*rand.Float64() - 1.0

		// 计算s = u^2 + v^2
		s := u*u + v*v

		// 如果s在(0, 1)范围内，则接受
		if s > 0.0 && s < 1.0 {
			// 计算乘数
			multiplier := math.Sqrt(-2.0 * math.Log(s) / s)

			// 缓存其中一个值供下次使用
			e.normalRandS = s
			e.normalRandV = v * multiplier
			e.normalRandOK = true

			// 返回另一个值
			return mean + stdDev*u*multiplier
		}
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

// readFileResponse 从文件读取响应内容
func (e *MockExecutor) readFileResponse(filePath string) ([]byte, error) {
	// 打开文件
	file, err := os.Open(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to open file: %w", err)
	}
	defer file.Close()

	// 读取文件内容
	data, err := io.ReadAll(file)
	if err != nil {
		return nil, fmt.Errorf("failed to read file: %w", err)
	}

	logger.Info("read file response",
		zap.String("file_path", filePath),
		zap.Int("size", len(data)))

	return data, nil
}
