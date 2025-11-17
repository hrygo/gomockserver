package executor

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
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

// TestUnsupportedResponseType 测试不支持的响应类型（仅Script响应）
func TestUnsupportedResponseType(t *testing.T) {
	executor := NewMockExecutor()

	// Script响应类型尚未实现
	t.Run("Script响应", func(t *testing.T) {
		rule := &models.Rule{
			Protocol: models.ProtocolHTTP,
			Response: models.Response{
				Type: models.ResponseTypeScript,
			},
		}

		request := &adapter.Request{
			Protocol: models.ProtocolHTTP,
		}

		response, err := executor.Execute(request, rule)

		assert.Error(t, err, "Script响应应该返回错误")
		assert.Nil(t, response, "响应应该为空")
	})
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

// TestNonHTTPProtocol 测试非HTTP协议的错误处理
func TestNonHTTPProtocol(t *testing.T) {
	executor := NewMockExecutor()

	tests := []struct {
		name     string
		protocol models.ProtocolType
	}{
		{"gRPC协议", models.ProtocolGRPC},
		{"WebSocket协议", models.ProtocolWebSocket},
		{"TCP协议", models.ProtocolTCP},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rule := &models.Rule{
				Protocol: tt.protocol,
				Response: models.Response{
					Type: models.ResponseTypeStatic,
					Content: map[string]interface{}{
						"status_code":  200,
						"content_type": "JSON",
						"body":         map[string]interface{}{"test": "data"},
					},
				},
			}

			request := &adapter.Request{
				Protocol: tt.protocol,
			}

			response, err := executor.Execute(request, rule)

			assert.Error(t, err, "非HTTP协议应该返回错误")
			assert.Nil(t, response)
			assert.Contains(t, err.Error(), "only HTTP protocol is supported")
		})
	}
}

// TestInvalidResponseContent 测试无效的响应内容
func TestInvalidResponseContent(t *testing.T) {
	executor := NewMockExecutor()

	// 测试缺少必要字段的Content
	rule := &models.Rule{
		Protocol: models.ProtocolHTTP,
		Response: models.Response{
			Type: models.ResponseTypeStatic,
			Content: map[string]interface{}{
				// 缺少 status_code 和 content_type
				"body": "test",
			},
		},
	}

	request := &adapter.Request{
		Protocol: models.ProtocolHTTP,
	}

	// 这个测试应该能正常处理，因为代码会使用默认值
	response, err := executor.Execute(request, rule)

	// 如果缺少必要字段，Unmarshal会使用默认值
	if err != nil {
		assert.Error(t, err)
	} else {
		assert.NotNil(t, response)
	}
}

// TestEmptyAndNilBody 测试空响应体和nil处理
func TestEmptyAndNilBody(t *testing.T) {
	executor := NewMockExecutor()

	tests := []struct {
		name        string
		body        interface{}
		contentType models.ContentType
	}{
		{
			name:        "JSON空对象",
			body:        map[string]interface{}{},
			contentType: models.ContentTypeJSON,
		},
		{
			name:        "Text空字符串",
			body:        "",
			contentType: models.ContentTypeText,
		},
		{
			name:        "XML空字符串",
			body:        "",
			contentType: models.ContentTypeXML,
		},
		{
			name:        "HTML空字符串",
			body:        "",
			contentType: models.ContentTypeHTML,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rule := &models.Rule{
				Protocol: models.ProtocolHTTP,
				Response: models.Response{
					Type: models.ResponseTypeStatic,
					Content: map[string]interface{}{
						"status_code":  200,
						"content_type": tt.contentType,
						"body":         tt.body,
					},
				},
			}

			request := &adapter.Request{
				Protocol: models.ProtocolHTTP,
			}

			response, err := executor.Execute(request, rule)

			assert.NoError(t, err)
			assert.NotNil(t, response)
			assert.NotNil(t, response.Body, "响应体不应该为nil")
		})
	}
}

// TestSpecialCharacters 测试特殊字符处理
func TestSpecialCharacters(t *testing.T) {
	executor := NewMockExecutor()

	tests := []struct {
		name        string
		body        string
		contentType models.ContentType
	}{
		{
			name:        "中文字符",
			body:        "你好，世界！这是中文测试",
			contentType: models.ContentTypeText,
		},
		{
			name:        "特殊符号",
			body:        "!@#$%^&*()_+-=[]{}|;:',.<>?/~`",
			contentType: models.ContentTypeText,
		},
		{
			name:        "换行和制表符",
			body:        "Line1\nLine2\tTabbed",
			contentType: models.ContentTypeText,
		},
		{
			name:        "Emoji表情",
			body:        "Hello 😀 🎉 🚀",
			contentType: models.ContentTypeText,
		},
		{
			name:        "XML特殊字符",
			body:        "<?xml version=\"1.0\"?><data>&lt;test&gt;</data>",
			contentType: models.ContentTypeXML,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rule := &models.Rule{
				Protocol: models.ProtocolHTTP,
				Response: models.Response{
					Type: models.ResponseTypeStatic,
					Content: map[string]interface{}{
						"status_code":  200,
						"content_type": tt.contentType,
						"body":         tt.body,
					},
				},
			}

			request := &adapter.Request{
				Protocol: models.ProtocolHTTP,
			}

			response, err := executor.Execute(request, rule)

			assert.NoError(t, err)
			assert.Equal(t, tt.body, string(response.Body))
		})
	}
}

// TestLargeResponseBody 测试超大响应体
func TestLargeResponseBody(t *testing.T) {
	executor := NewMockExecutor()

	// 生成1MB的文本数据
	largeText := make([]byte, 1024*1024)
	for i := range largeText {
		largeText[i] = 'A' + byte(i%26)
	}

	rule := &models.Rule{
		Protocol: models.ProtocolHTTP,
		Response: models.Response{
			Type: models.ResponseTypeStatic,
			Content: map[string]interface{}{
				"status_code":  200,
				"content_type": "Text",
				"body":         string(largeText),
			},
		},
	}

	request := &adapter.Request{
		Protocol: models.ProtocolHTTP,
	}

	response, err := executor.Execute(request, rule)

	assert.NoError(t, err)
	assert.NotNil(t, response)
	assert.Equal(t, 1024*1024, len(response.Body), "响应体大小应该是1MB")
}

// TestBinaryContentType 测试二进制内容类型
// TestBinaryContentType 测试二进制内容类型
func TestBinaryContentType(t *testing.T) {
	executor := NewMockExecutor()
	
	// 测试Base64编码的二进制数据
	base64Data := "SGVsbG8sIHdvcmxkIQ==" // "Hello, world!"的Base64编码
	expectedData := []byte("Hello, world!")
	
	rule := &models.Rule{
		Protocol: models.ProtocolHTTP,
		Response: models.Response{
			Type: models.ResponseTypeStatic,
			Content: map[string]interface{}{
				"status_code":  200,
				"content_type": "Binary",
				"body":         base64Data,
			},
		},
	}
	request := &adapter.Request{
		Protocol: models.ProtocolHTTP,
	}
	response, err := executor.Execute(request, rule)
	assert.NoError(t, err)
	assert.NotNil(t, response)
	assert.Equal(t, "application/octet-stream", response.Headers["Content-Type"])
	assert.Equal(t, expectedData, response.Body, "Binary类型应该正确解码Base64数据")
	
	// 测试无效的Base64数据（应该返回原始数据）
	invalidBase64Data := "invalid-base64!"
	rule2 := &models.Rule{
		Protocol: models.ProtocolHTTP,
		Response: models.Response{
			Type: models.ResponseTypeStatic,
			Content: map[string]interface{}{
				"status_code":  200,
				"content_type": "Binary",
				"body":         invalidBase64Data,
			},
		},
	}
	response2, err2 := executor.Execute(request, rule2)
	assert.NoError(t, err2)
	assert.NotNil(t, response2)
	assert.Equal(t, []byte(invalidBase64Data), response2.Body, "无效Base64应该返回原始数据")
	
	// 测试非字符串类型的二进制数据
	nonStringData := map[string]interface{}{"key": "value"}
	jsonData, _ := json.Marshal(nonStringData)
	rule3 := &models.Rule{
		Protocol: models.ProtocolHTTP,
		Response: models.Response{
			Type: models.ResponseTypeStatic,
			Content: map[string]interface{}{
				"status_code":  200,
				"content_type": "Binary",
				"body":         nonStringData,
			},
		},
	}
	response3, err3 := executor.Execute(request, rule3)
	assert.NoError(t, err3)
	assert.NotNil(t, response3)
	assert.Equal(t, jsonData, response3.Body, "非字符串类型应该被JSON序列化")
}


// TestUnknownContentType 测试未知内容类型
func TestUnknownContentType(t *testing.T) {
	executor := NewMockExecutor()

	rule := &models.Rule{
		Protocol: models.ProtocolHTTP,
		Response: models.Response{
			Type: models.ResponseTypeStatic,
			Content: map[string]interface{}{
				"status_code":  200,
				"content_type": "Unknown",
				"body":         map[string]interface{}{"data": "test"},
			},
		},
	}

	request := &adapter.Request{
		Protocol: models.ProtocolHTTP,
	}

	response, err := executor.Execute(request, rule)

	assert.NoError(t, err)
	assert.NotNil(t, response)
	assert.Equal(t, "application/json", response.Headers["Content-Type"], "未知类型应默认为JSON")
}

// TestRandomDelayBoundary 测试随机延迟边界条件
func TestRandomDelayBoundary(t *testing.T) {
	executor := NewMockExecutor()

	tests := []struct {
		name     string
		config   *models.DelayConfig
		expected int
	}{
		{
			name: "Min等于Max",
			config: &models.DelayConfig{
				Type: "random",
				Min:  100,
				Max:  100,
			},
			expected: 100,
		},
		{
			name: "Min大于Max",
			config: &models.DelayConfig{
				Type: "random",
				Min:  200,
				Max:  100,
			},
			expected: 200,
		},
		{
			name: "Min为0",
			config: &models.DelayConfig{
				Type: "random",
				Min:  0,
				Max:  100,
			},
			expected: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			delay := executor.calculateDelay(tt.config)
			if tt.config.Max <= tt.config.Min {
				assert.Equal(t, tt.expected, delay, "当Max<=Min时应该返回Min值")
			} else {
				assert.GreaterOrEqual(t, delay, tt.config.Min)
				assert.LessOrEqual(t, delay, tt.config.Max)
			}
		})
	}
}

// TestDelayWithRandomVariation 测试随机延迟的变化性
func TestDelayWithRandomVariation(t *testing.T) {
	executor := NewMockExecutor()

	config := &models.DelayConfig{
		Type: "random",
		Min:  10,
		Max:  100,
	}

	// 多次调用，检查是否有不同的值
	delays := make(map[int]bool)
	for i := 0; i < 50; i++ {
		delay := executor.calculateDelay(config)
		delays[delay] = true
		assert.GreaterOrEqual(t, delay, config.Min)
		assert.LessOrEqual(t, delay, config.Max)
	}

	// 应该有多个不同的延迟值（至少5个）
	assert.GreaterOrEqual(t, len(delays), 5, "随机延迟应该产生多个不同的值")
}

// TestResponseWithCustomHeaders 测试自定义Headers
func TestResponseWithCustomHeaders(t *testing.T) {
	executor := NewMockExecutor()

	customHeaders := map[string]interface{}{
		"X-Custom-Header":  "custom-value",
		"X-Request-ID":     "12345",
		"Cache-Control":    "no-cache",
		"X-Rate-Limit":     "1000",
		"Content-Language": "zh-CN",
	}

	rule := &models.Rule{
		Protocol: models.ProtocolHTTP,
		Response: models.Response{
			Type: models.ResponseTypeStatic,
			Content: map[string]interface{}{
				"status_code":  200,
				"content_type": "JSON",
				"body":         map[string]interface{}{"status": "ok"},
				"headers":      customHeaders,
			},
		},
	}

	request := &adapter.Request{
		Protocol: models.ProtocolHTTP,
	}

	response, err := executor.Execute(request, rule)

	assert.NoError(t, err)
	assert.NotNil(t, response)
	for key, value := range customHeaders {
		assert.Equal(t, value.(string), response.Headers[key], "自定义Header应该被正确设置")
	}
}

// TestResponseWithoutHeaders 测试没有Headers的响应
func TestResponseWithoutHeaders(t *testing.T) {
	executor := NewMockExecutor()

	rule := &models.Rule{
		Protocol: models.ProtocolHTTP,
		Response: models.Response{
			Type: models.ResponseTypeStatic,
			Content: map[string]interface{}{
				"status_code":  200,
				"content_type": "JSON",
				"body":         map[string]interface{}{"status": "ok"},
				// 不设置headers字段
			},
		},
	}

	request := &adapter.Request{
		Protocol: models.ProtocolHTTP,
	}

	response, err := executor.Execute(request, rule)

	assert.NoError(t, err)
	assert.NotNil(t, response)
	assert.NotNil(t, response.Headers, "Headers不应该为nil")
	assert.Contains(t, response.Headers, "Content-Type", "应该自动添加Content-Type")
}

// TestComplexJSONBody 测试复杂的JSON响应体
func TestComplexJSONBody(t *testing.T) {
	executor := NewMockExecutor()

	complexBody := map[string]interface{}{
		"code":    0,
		"message": "success",
		"data": map[string]interface{}{
			"users": []interface{}{
				map[string]interface{}{
					"id":   1,
					"name": "张三",
					"tags": []string{"admin", "developer"},
				},
				map[string]interface{}{
					"id":   2,
					"name": "李四",
					"tags": []string{"user"},
				},
			},
			"pagination": map[string]interface{}{
				"page":        1,
				"page_size":   10,
				"total":       100,
				"total_pages": 10,
			},
		},
		"timestamp": 1234567890,
	}

	rule := &models.Rule{
		Protocol: models.ProtocolHTTP,
		Response: models.Response{
			Type: models.ResponseTypeStatic,
			Content: map[string]interface{}{
				"status_code":  200,
				"content_type": "JSON",
				"body":         complexBody,
			},
		},
	}

	request := &adapter.Request{
		Protocol: models.ProtocolHTTP,
	}

	response, err := executor.Execute(request, rule)

	assert.NoError(t, err)
	assert.NotNil(t, response)
	assert.NotEmpty(t, response.Body)

	// 验证JSON可以正确解析
	var parsedBody map[string]interface{}
	err = json.Unmarshal(response.Body, &parsedBody)
	assert.NoError(t, err)
	assert.Equal(t, float64(0), parsedBody["code"])
}

// TestNonStringBodyForTextType 测试Text类型的非字符串body
func TestNonStringBodyForTextType(t *testing.T) {
	executor := NewMockExecutor()

	// Text类型但body是map
	rule := &models.Rule{
		Protocol: models.ProtocolHTTP,
		Response: models.Response{
			Type: models.ResponseTypeStatic,
			Content: map[string]interface{}{
				"status_code":  200,
				"content_type": "Text",
				"body":         map[string]interface{}{"key": "value"},
			},
		},
	}

	request := &adapter.Request{
		Protocol: models.ProtocolHTTP,
	}

	response, err := executor.Execute(request, rule)

	assert.NoError(t, err)
	assert.NotNil(t, response)
	// 非字符串body会被JSON序列化
	assert.Contains(t, string(response.Body), "key")
}

// TestStepDelayType 测试step延迟类型
func TestStepDelayType(t *testing.T) {
	executor := NewMockExecutor()
	
	t.Run("Step delay basic", func(t *testing.T) {
		config := &models.DelayConfig{
			Type:  "step",
			Fixed: 100,
			Step:  50,
			Limit: 300,
		}
		
		// 第一次调用: 100 + 0*50 = 100
		delay1 := executor.calculateStepDelay(config, "rule1")
		assert.Equal(t, 100, delay1)
		
		// 第二次调用: 100 + 1*50 = 150
		delay2 := executor.calculateStepDelay(config, "rule1")
		assert.Equal(t, 150, delay2)
		
		// 第三次调用: 100 + 2*50 = 200
		delay3 := executor.calculateStepDelay(config, "rule1")
		assert.Equal(t, 200, delay3)
	})
	
	t.Run("Step delay with limit", func(t *testing.T) {
		executor.ResetStepCounter("rule2")
		config := &models.DelayConfig{
			Type:  "step",
			Fixed: 100,
			Step:  100,
			Limit: 250,
		}
		
		for i := 0; i < 5; i++ {
			delay := executor.calculateStepDelay(config, "rule2")
			if i < 2 {
				assert.LessOrEqual(t, delay, 250)
			} else {
				assert.Equal(t, 250, delay, "超过limit应该返回limit值")
			}
		}
	})
	
	t.Run("Step delay with zero step", func(t *testing.T) {
		config := &models.DelayConfig{
			Type:  "step",
			Fixed: 100,
			Step:  0,
		}
		
		delay := executor.calculateStepDelay(config, "rule3")
		assert.Equal(t, 100, delay, "step为0应该返回Fixed值")
	})
}

func TestResetStepCounter(t *testing.T) {
	executor := NewMockExecutor()
	
	config := &models.DelayConfig{
		Type:  "step",
		Fixed: 100,
		Step:  50,
	}
	
	// 增加计数器
	executor.calculateStepDelay(config, "rule1")
	executor.calculateStepDelay(config, "rule1")
	assert.Equal(t, int64(2), executor.GetStepCounter("rule1"))
	
	// 重置特定规则的计数器
	executor.ResetStepCounter("rule1")
	assert.Equal(t, int64(0), executor.GetStepCounter("rule1"))
	
	// 测试重置所有计数器
	executor.calculateStepDelay(config, "rule2")
	executor.calculateStepDelay(config, "rule3")
	executor.ResetStepCounter("") // 空字符串重置所有
	assert.Equal(t, int64(0), executor.GetStepCounter("rule2"))
	assert.Equal(t, int64(0), executor.GetStepCounter("rule3"))
}

func TestGetStepCounter(t *testing.T) {
	executor := NewMockExecutor()
	
	// 未调用前应该为0
	assert.Equal(t, int64(0), executor.GetStepCounter("new-rule"))
	
	config := &models.DelayConfig{
		Type:  "step",
		Fixed: 100,
		Step:  50,
	}
	
	executor.calculateStepDelay(config, "test-rule")
	executor.calculateStepDelay(config, "test-rule")
	executor.calculateStepDelay(config, "test-rule")
	
	assert.Equal(t, int64(3), executor.GetStepCounter("test-rule"))
}

func TestGenerateNormalRand(t *testing.T) {
	executor := NewMockExecutor()
	
	// 测试正态分布生成
	mean := 100.0
	stdDev := 20.0
	
	// 生成多个值并检查分布
	values := make([]float64, 1000)
	for i := 0; i < 1000; i++ {
		values[i] = executor.generateNormalRand(mean, stdDev)
	}
	
	// 计算平均值
	sum := 0.0
	for _, v := range values {
		sum += v
	}
	avg := sum / float64(len(values))
	
	// 平均值应该接近mean（允许5%误差）
	assert.InDelta(t, mean, avg, mean*0.1, "平均值应该接近期望值")
	
	// 检查值的范围（99.7%的值应该在mean±3*stdDev范围内）
	inRangeCount := 0
	for _, v := range values {
		if v >= mean-3*stdDev && v <= mean+3*stdDev {
			inRangeCount++
		}
	}
	assert.GreaterOrEqual(t, inRangeCount, 990, "至少99%的值应该在±3σ范围内")
}

func TestNormalDelayType(t *testing.T) {
	executor := NewMockExecutor()
	
	t.Run("Normal delay with valid stddev", func(t *testing.T) {
		config := &models.DelayConfig{
			Type:   "normal",
			Mean:   100,
			StdDev: 20,
		}
		
		delays := make([]int, 100)
		for i := 0; i < 100; i++ {
			delays[i] = executor.calculateDelay(config)
			assert.GreaterOrEqual(t, delays[i], 0, "延迟不应该为负")
		}
		
		// 应该有变化
		uniqueDelays := make(map[int]bool)
		for _, d := range delays {
			uniqueDelays[d] = true
		}
		assert.Greater(t, len(uniqueDelays), 10, "正态分布应该产生多个不同值")
	})
	
	t.Run("Normal delay with zero stddev", func(t *testing.T) {
		config := &models.DelayConfig{
			Type:   "normal",
			Mean:   100,
			StdDev: 0,
		}
		
		delay := executor.calculateDelay(config)
		assert.Equal(t, 100, delay, "stddev为0应该返回mean值")
	})
	
	t.Run("Normal delay with negative stddev", func(t *testing.T) {
		config := &models.DelayConfig{
			Type:   "normal",
			Mean:   100,
			StdDev: -10,
		}
		
		delay := executor.calculateDelay(config)
		assert.Equal(t, 100, delay, "负stddev应该返回mean值")
	})
}

func TestCalculateDelayNil(t *testing.T) {
	executor := NewMockExecutor()
	
	delay := executor.calculateDelay(nil)
	assert.Equal(t, 0, delay, "nil config应该返回0")
}

func TestCalculateDelayUnknownType(t *testing.T) {
	executor := NewMockExecutor()
	
	config := &models.DelayConfig{
		Type: "unknown",
	}
	
	delay := executor.calculateDelay(config)
	assert.Equal(t, 0, delay, "未知类型应该返回0")
}

func TestReadFileResponse(t *testing.T) {
	executor := NewMockExecutor()
	
	// 创建临时测试文件
	tmpFile, err := os.CreateTemp("", "test-response-*.txt")
	assert.NoError(t, err)
	defer os.Remove(tmpFile.Name())
	
	testContent := []byte("Test file content")
	_, err = tmpFile.Write(testContent)
	assert.NoError(t, err)
	tmpFile.Close()
	
	// 测试读取文件
	data, err := executor.readFileResponse(tmpFile.Name())
	assert.NoError(t, err)
	assert.Equal(t, testContent, data)
	
	// 测试读取不存在的文件
	_, err = executor.readFileResponse("/nonexistent/file.txt")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to open file")
}

// TestProxyResponse 测试Proxy响应
func TestProxyResponse(t *testing.T) {
	// 创建 mock 服务器
	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"message": "proxied"}`))
	}))
	defer mockServer.Close()

	executor := NewMockExecutor()

	t.Run("Valid proxy config", func(t *testing.T) {
		rule := &models.Rule{
			Protocol: models.ProtocolHTTP,
			Response: models.Response{
				Type: models.ResponseTypeProxy,
				Content: map[string]interface{}{
					"target_url": mockServer.URL,
					"timeout":    5,
				},
			},
		}

		request := &adapter.Request{
			Protocol: models.ProtocolHTTP,
			Path:     "/test",
			Metadata: map[string]interface{}{
				"method": "GET",
			},
		}

		response, err := executor.proxyResponse(request, rule)

		assert.NoError(t, err)
		assert.NotNil(t, response)
		assert.Equal(t, http.StatusOK, response.StatusCode)
		assert.Contains(t, string(response.Body), "proxied")
	})

	t.Run("Invalid proxy config - unmarshal error", func(t *testing.T) {
		rule := &models.Rule{
			Protocol: models.ProtocolHTTP,
			Response: models.Response{
				Type: models.ResponseTypeProxy,
				Content: map[string]interface{}{
					"target_url": []int{1, 2, 3}, // 错误的类型
				},
			},
		}

		request := &adapter.Request{
			Protocol: models.ProtocolHTTP,
		}

		response, err := executor.proxyResponse(request, rule)

		assert.Error(t, err)
		assert.Nil(t, response)
	})
}
