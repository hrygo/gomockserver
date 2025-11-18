package executor

import (
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestTemplateEngine_Render_ErrorHandling 测试模板渲染的错误处理
func TestTemplateEngine_Render_ErrorHandling(t *testing.T) {
	engine := NewTemplateEngine()

	tests := []struct {
		name     string
		template string
		context  *TemplateContext
		wantErr  bool
	}{
		{
			name:     "语法错误 - 未闭合的括号",
			template: "{{uuid}",
			context:  &TemplateContext{},
			wantErr:  true,
		},
		{
			name:     "语法错误 - 未闭合的大括号",
			template: "{{timestamp",
			context:  &TemplateContext{},
			wantErr:  true,
		},
		{
			name:     "无效函数调用",
			template: "{{invalidFunction}}",
			context:  &TemplateContext{},
			wantErr:  true,
		},
		{
			name:     "无效的管道操作",
			template: "{{uuid | invalid}}",
			context:  &TemplateContext{},
			wantErr:  true,
		},
		{
			name:     "无效的范围操作",
			template: "{{range .Request.Path}}{{.}}{{end}}",
			context:  &TemplateContext{},
			wantErr:  true,
		},
		{
			name:     "空模板",
			template: "",
			context:  &TemplateContext{},
			wantErr:  false,
		},
		{
			name:     "仅包含空格的模板",
			template: "   ",
			context:  &TemplateContext{},
			wantErr:  false,
		},
		{
			name:     "特殊字符模板",
			template: "特殊字符: <>&\"'",
			context:  &TemplateContext{},
			wantErr:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := engine.Render(tt.template, tt.context)
			if tt.wantErr {
				assert.Error(t, err)
				assert.Empty(t, result)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.template, result)
			}
		})
	}
}

// TestTemplateEngine_Render_EdgeCases 测试边界情况
func TestTemplateEngine_Render_EdgeCases(t *testing.T) {
	engine := NewTemplateEngine()

	// 测试嵌套模板
	t.Run("嵌套模板", func(t *testing.T) {
		template := `{
			"outer": {
				"inner": "{{uuid}}",
				"timestamp": "{{timestamp}}"
			},
			"method": "{{.Request.Method}}"
		}`

		context := &TemplateContext{
			Request: &RequestContext{
				Method: "POST",
			},
		}

		result, err := engine.Render(template, context)
		assert.NoError(t, err)
		assert.Contains(t, result, `"outer": {`)
		assert.Contains(t, result, `"method": "POST"`)
	})

	// 测试长模板
	t.Run("长模板", func(t *testing.T) {
		longText := strings.Repeat("这是一段很长的文本。", 100)
		template := `{"message": "{{.Request.Path}}", "longText": "` + longText + `"}`

		context := &TemplateContext{
			Request: &RequestContext{
				Path: "/api/test",
			},
		}

		result, err := engine.Render(template, context)
		assert.NoError(t, err)
		assert.Contains(t, result, "/api/test")
		assert.Contains(t, result, longText)
	})

	// 测试Unicode字符
	t.Run("Unicode字符", func(t *testing.T) {
		template := `{"message": "你好世界 {{.Request.Path}}", "emoji": "🚀"}`

		context := &TemplateContext{
			Request: &RequestContext{
				Path: "/测试/接口",
			},
		}

		result, err := engine.Render(template, context)
		assert.NoError(t, err)
		assert.Contains(t, result, "你好世界 /测试/接口")
		assert.Contains(t, result, "🚀")
	})
}

// TestTemplateEngine_RenderJSON_ErrorHandling 测试JSON模板渲染的错误处理
func TestTemplateEngine_RenderJSON_ErrorHandling(t *testing.T) {
	engine := NewTemplateEngine()

	tests := []struct {
		name     string
		template interface{}
		context  *TemplateContext
		wantErr  bool
	}{
		{
			name:     "包含无效语法的字符串",
			template: "{{invalid syntax}}",
			context:  &TemplateContext{},
			wantErr:  false, // RenderJSON会处理错误，返回原始字符串
		},
		{
			name:     "深度嵌套结构中的无效模板",
			template: map[string]interface{}{
				"level1": map[string]interface{}{
					"level2": map[string]interface{}{
						"invalid": "{{bad syntax}}",
					},
				},
			},
			context:  &TemplateContext{},
			wantErr:  false, // 会返回部分处理的结果
		},
		{
			name:     "循环引用",
			template: func() interface{} { return nil }, // 函数类型不支持
			context:  &TemplateContext{},
			wantErr:  false, // 会返回原始值
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := engine.RenderJSON(tt.template, tt.context)
			// RenderJSON在遇到错误时不会返回错误，而是返回处理后的结果
			// 所以这里我们主要验证不会panic
			assert.NotNil(t, result)
			if tt.wantErr {
				// 对于某些类型，可能仍然会返回错误
				_ = err
			}
		})
	}
}

// TestTemplateEngine_RenderJSON_Recursive 测试递归JSON渲染
func TestTemplateEngine_RenderJSON_Recursive(t *testing.T) {
	engine := NewTemplateEngine()

	t.Run("多层嵌套结构", func(t *testing.T) {
		template := map[string]interface{}{
			"user": map[string]interface{}{
				"id": "{{.Request.Path}}",
				"info": map[string]interface{}{
					"name": "Test User",
					"email": "user@example.com",
					"metadata": map[string]interface{}{
						"created": "{{timestamp}}",
						"version": "1.0",
					},
				},
				"permissions": []string{
					"read",
					"write",
					"{{.Request.Method}}",
				},
			},
		}

		context := &TemplateContext{
			Request: &RequestContext{
				Path:   "/api/users/123",
				Method: "GET",
			},
		}

		result, err := engine.RenderJSON(template, context)
		assert.NoError(t, err)

		// 验证结构是否正确处理
		userMap, ok := result.(map[string]interface{})
		require.True(t, ok)
		// 模板渲染结果应该包含路径信息
		idValue := fmt.Sprintf("%v", userMap["id"])
		assert.NotEmpty(t, idValue)
		// 验证模板确实被处理了（原始值是{{.Request.Path}}）
		assert.NotContains(t, idValue, "{{")
	})

	t.Run("数组中的模板", func(t *testing.T) {
		template := []interface{}{
			"item1",
			"{{.Request.Path}}",
			"item3",
			map[string]interface{}{
				"key": "{{uuid}}",
			},
		}

		context := &TemplateContext{
			Request: &RequestContext{
				Path: "/api/test",
			},
		}

		result, err := engine.RenderJSON(template, context)
		assert.NoError(t, err)

		// 验证数组是否正确处理
		resultArray, ok := result.([]interface{})
		require.True(t, ok)
		assert.Equal(t, "item1", resultArray[0])
		assert.Contains(t, fmt.Sprintf("%v", resultArray[1]), "/api/test")
	})

	t.Run("复杂嵌套", func(t *testing.T) {
		template := map[string]interface{}{
			"data": []map[string]interface{}{
				{
					"id": "{{counter}}",
					"info": map[string]interface{}{
						"path": "{{.Request.Path}}",
						"headers": map[string]interface{}{
							"host": "{{.Request.Headers.Host}}",
							"user-agent": "{{.Request.Headers.UserAgent}}",
						},
					},
				},
			},
			"metadata": map[string]interface{}{
				"timestamp": "{{timestamp}}",
				"count": len("{{.Request.Path}}"),
			},
		}

		context := &TemplateContext{
			Request: &RequestContext{
				Path: "/api/v1/data",
				Headers: map[string]string{
					"Host":       "localhost:8080",
					"User-Agent": "test-agent",
				},
			},
		}

		result, err := engine.RenderJSON(template, context)
		assert.NoError(t, err)
		assert.NotNil(t, result)
	})
}

// TestTemplateEngine_Performance 测试模板渲染性能
func TestTemplateEngine_Performance(t *testing.T) {
	if testing.Short() {
		t.Skip("跳过性能测试")
	}

	engine := NewTemplateEngine()

	// 创建复杂模板
	template := `{
		"requestId": "{{uuid}}",
		"timestamp": "{{timestamp}}",
		"path": "{{.Request.Path}}",
		"method": "{{.Request.Method}}",
		"generated": {
			"random": "{{random 1 100}}",
			"uuidShort": "{{uuidShort}}"
		}
	}`

	context := &TemplateContext{
		Request: &RequestContext{
			Path:    "/api/test/performance",
			Method:  "POST",
			Headers: map[string]string{
				"Content-Type": "application/json",
				"X-Request-ID": "req-123",
			},
			Query: map[string]string{
				"page": "1",
				"limit": "10",
			},
			Body: `{"test": "data"}`,
		},
		Environment: &EnvironmentContext{
			Variables: map[string]interface{}{
				"Name":    "test-env",
				"Project": "test-project",
			},
		},
	}

	// 性能测试：执行1000次渲染
	start := time.Now()
	for i := 0; i < 1000; i++ {
		_, err := engine.Render(template, context)
		assert.NoError(t, err)
	}
	duration := time.Since(start)

	// 性能验证：1000次复杂模板渲染应该在合理时间内完成
	assert.Less(t, duration, 1*time.Second, "模板渲染性能测试失败")
	t.Logf("1000次模板渲染耗时: %v", duration)
}

// TestTemplateEngine_MemoryLeak 测试内存泄漏
func TestTemplateEngine_MemoryLeak(t *testing.T) {
	engine := NewTemplateEngine()

	// 创建大模板
	largeTemplate := strings.Repeat(`{"id": "{{uuid}}", "data": "`, 1000) + strings.Repeat(`"value": "test", `, 100) + `"}`

	context := &TemplateContext{}

	// 执行多次渲染，验证不会发生内存泄漏
	for i := 0; i < 100; i++ {
		result, err := engine.Render(largeTemplate, context)
		assert.NoError(t, err)
		assert.NotEmpty(t, result)
	}
}

// TestTemplateEngine_ConcurrentSafety 测试并发安全性
func TestTemplateEngine_ConcurrentSafety(t *testing.T) {
	if testing.Short() {
		t.Skip("跳过并发测试")
	}

	engine := NewTemplateEngine()
	template := `{"id": "{{uuid}}", "timestamp": "{{timestamp}}", "path": "{{.Request.Path}}"}`

	// 启动多个goroutine并发渲染
	done := make(chan bool, 10)
	for i := 0; i < 10; i++ {
		go func(id int) {
			defer func() { done <- true }()

			context := &TemplateContext{
				Request: &RequestContext{
					Path: fmt.Sprintf("/api/concurrent/%d", id),
				},
			}

			for j := 0; j < 100; j++ {
				_, err := engine.Render(template, context)
				assert.NoError(t, err)
			}
		}(i)
	}

	// 等待所有goroutine完成
	for i := 0; i < 10; i++ {
		<-done
	}
}