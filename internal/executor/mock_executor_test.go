package executor

import (
	"testing"
	"time"

	"github.com/gomockserver/mockserver/internal/adapter"
	"github.com/gomockserver/mockserver/internal/models"
	"github.com/stretchr/testify/assert"
)

// TestCalculateDelay 测试延迟计算
func TestCalculateDelay(t *testing.T) {
	executor := NewMockExecutor()

	tests := []struct {
		name        string
		config      *models.DelayConfig
		minExpected int
		maxExpected int
	}{
		{
			name: "固定延迟",
			config: &models.DelayConfig{
				Type:  "fixed",
				Fixed: 100,
			},
			minExpected: 100,
			maxExpected: 100,
		},
		{
			name: "随机延迟",
			config: &models.DelayConfig{
				Type: "random",
				Min:  50,
				Max:  200,
			},
			minExpected: 50,
			maxExpected: 200,
		},
		{
			name: "正态分布延迟(暂返回均值)",
			config: &models.DelayConfig{
				Type: "normal",
				Mean: 150,
			},
			minExpected: 150,
			maxExpected: 150,
		},
		{
			name: "无效延迟类型",
			config: &models.DelayConfig{
				Type: "invalid",
			},
			minExpected: 0,
			maxExpected: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			delay := executor.calculateDelay(tt.config)
			assert.GreaterOrEqual(t, delay, tt.minExpected, "延迟不应小于最小值")
			assert.LessOrEqual(t, delay, tt.maxExpected, "延迟不应大于最大值")
		})
	}
}

// TestGetDefaultContentType 测试默认Content-Type获取
func TestGetDefaultContentType(t *testing.T) {
	executor := NewMockExecutor()

	tests := []struct {
		name        string
		contentType models.ContentType
		expected    string
	}{
		{"JSON", models.ContentTypeJSON, "application/json"},
		{"XML", models.ContentTypeXML, "application/xml"},
		{"HTML", models.ContentTypeHTML, "text/html"},
		{"Text", models.ContentTypeText, "text/plain"},
		{"Binary", models.ContentTypeBinary, "application/octet-stream"},
		{"默认", models.ContentType("unknown"), "application/json"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := executor.getDefaultContentType(tt.contentType)
			assert.Equal(t, tt.expected, result)
		})
	}
}

// TestStaticJSONResponse 测试JSON静态响应
func TestStaticJSONResponse(t *testing.T) {
	executor := NewMockExecutor()

	rule := &models.Rule{
		Protocol: models.ProtocolHTTP,
		Response: models.Response{
			Type: models.ResponseTypeStatic,
			Content: map[string]interface{}{
				"status_code":  200,
				"content_type": "JSON",
				"body": map[string]interface{}{
					"code":    0,
					"message": "success",
					"data":    []interface{}{},
				},
				"headers": map[string]interface{}{
					"X-Custom": "value",
				},
			},
		},
	}

	request := &adapter.Request{
		Protocol: models.ProtocolHTTP,
	}

	response, err := executor.Execute(request, rule)

	assert.NoError(t, err, "执行不应该出错")
	assert.NotNil(t, response, "响应不应该为空")
	assert.Equal(t, 200, response.StatusCode, "状态码应该是200")
	assert.Contains(t, response.Headers, "Content-Type", "应该包含Content-Type")
	assert.Contains(t, response.Headers, "X-Custom", "应该包含自定义Header")
	assert.NotEmpty(t, response.Body, "响应体不应该为空")
}

// TestStaticTextResponse 测试文本静态响应
func TestStaticTextResponse(t *testing.T) {
	executor := NewMockExecutor()

	rule := &models.Rule{
		Protocol: models.ProtocolHTTP,
		Response: models.Response{
			Type: models.ResponseTypeStatic,
			Content: map[string]interface{}{
				"status_code":  200,
				"content_type": "Text",
				"body":         "Hello, World!",
			},
		},
	}

	request := &adapter.Request{
		Protocol: models.ProtocolHTTP,
	}

	response, err := executor.Execute(request, rule)

	assert.NoError(t, err)
	assert.Equal(t, 200, response.StatusCode)
	assert.Equal(t, "Hello, World!", string(response.Body))
}

// TestResponseWithDelay 测试带延迟的响应
func TestResponseWithDelay(t *testing.T) {
	executor := NewMockExecutor()

	rule := &models.Rule{
		Protocol: models.ProtocolHTTP,
		Response: models.Response{
			Type: models.ResponseTypeStatic,
			Delay: &models.DelayConfig{
				Type:  "fixed",
				Fixed: 50, // 50ms延迟
			},
			Content: map[string]interface{}{
				"status_code":  200,
				"content_type": "JSON",
				"body":         map[string]interface{}{"message": "delayed"},
			},
		},
	}

	request := &adapter.Request{
		Protocol: models.ProtocolHTTP,
	}

	start := time.Now()
	response, err := executor.Execute(request, rule)
	duration := time.Since(start)

	assert.NoError(t, err)
	assert.NotNil(t, response)
	assert.GreaterOrEqual(t, duration.Milliseconds(), int64(50), "应该有延迟")
}

// TestGetDefaultResponse 测试默认404响应
func TestGetDefaultResponse(t *testing.T) {
	executor := NewMockExecutor()

	response := executor.GetDefaultResponse()

	assert.NotNil(t, response)
	assert.Equal(t, 404, response.StatusCode)
	assert.Contains(t, response.Headers, "Content-Type")
	assert.Contains(t, string(response.Body), "No matching rule found")
}

// TestUnsupportedResponseType 测试不支持的响应类型
func TestUnsupportedResponseType(t *testing.T) {
	executor := NewMockExecutor()

	tests := []struct {
		name         string
		responseType models.ResponseType
	}{
		{"Dynamic响应", models.ResponseTypeDynamic},
		{"Script响应", models.ResponseTypeScript},
		{"Proxy响应", models.ResponseTypeProxy},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rule := &models.Rule{
				Protocol: models.ProtocolHTTP,
				Response: models.Response{
					Type: tt.responseType,
				},
			}

			request := &adapter.Request{
				Protocol: models.ProtocolHTTP,
			}

			response, err := executor.Execute(request, rule)

			assert.Error(t, err, "应该返回错误")
			assert.Nil(t, response, "响应应该为空")
		})
	}
}

// TestDifferentStatusCodes 测试不同状态码
func TestDifferentStatusCodes(t *testing.T) {
	executor := NewMockExecutor()

	statusCodes := []int{200, 201, 204, 400, 404, 500, 503}

	for _, statusCode := range statusCodes {
		t.Run(string(rune(statusCode)), func(t *testing.T) {
			rule := &models.Rule{
				Protocol: models.ProtocolHTTP,
				Response: models.Response{
					Type: models.ResponseTypeStatic,
					Content: map[string]interface{}{
						"status_code":  statusCode,
						"content_type": "JSON",
						"body":         map[string]interface{}{"status": statusCode},
					},
				},
			}

			request := &adapter.Request{
				Protocol: models.ProtocolHTTP,
			}

			response, err := executor.Execute(request, rule)

			assert.NoError(t, err)
			assert.Equal(t, statusCode, response.StatusCode)
		})
	}
}

// TestXMLResponse 测试XML响应
func TestXMLResponse(t *testing.T) {
	executor := NewMockExecutor()

	rule := &models.Rule{
		Protocol: models.ProtocolHTTP,
		Response: models.Response{
			Type: models.ResponseTypeStatic,
			Content: map[string]interface{}{
				"status_code":  200,
				"content_type": "XML",
				"body":         "<users><user>张三</user></users>",
			},
		},
	}

	request := &adapter.Request{
		Protocol: models.ProtocolHTTP,
	}

	response, err := executor.Execute(request, rule)

	assert.NoError(t, err)
	assert.Equal(t, 200, response.StatusCode)
	assert.Contains(t, response.Headers["Content-Type"], "xml")
	assert.Contains(t, string(response.Body), "<users>")
}

// TestHTMLResponse 测试HTML响应
func TestHTMLResponse(t *testing.T) {
	executor := NewMockExecutor()

	rule := &models.Rule{
		Protocol: models.ProtocolHTTP,
		Response: models.Response{
			Type: models.ResponseTypeStatic,
			Content: map[string]interface{}{
				"status_code":  200,
				"content_type": "HTML",
				"body":         "<html><body>Hello</body></html>",
			},
		},
	}

	request := &adapter.Request{
		Protocol: models.ProtocolHTTP,
	}

	response, err := executor.Execute(request, rule)

	assert.NoError(t, err)
	assert.Equal(t, 200, response.StatusCode)
	assert.Contains(t, response.Headers["Content-Type"], "html")
}
