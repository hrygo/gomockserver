package executor

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gomockserver/mockserver/internal/adapter"
	"github.com/stretchr/testify/assert"
)

func TestNewProxyExecutor(t *testing.T) {
	executor := NewProxyExecutor()
	assert.NotNil(t, executor)
	assert.NotNil(t, executor.client)
	assert.Equal(t, 30*time.Second, executor.client.Timeout)
}

func TestProxyExecutor_Execute_BasicProxy(t *testing.T) {
	// 创建mock服务器
	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"message": "success"}`))
	}))
	defer mockServer.Close()

	executor := NewProxyExecutor()
	config := &ProxyConfig{
		TargetURL: mockServer.URL,
	}

	request := &adapter.Request{
		Path:    "/test",
		Headers: map[string]string{"User-Agent": "test"},
		Body:    []byte(`{"key": "value"}`),
		Metadata: map[string]interface{}{
			"method": "POST",
		},
	}

	response, err := executor.Execute(request, config)

	assert.NoError(t, err)
	assert.NotNil(t, response)
	assert.Equal(t, http.StatusOK, response.StatusCode)
	assert.Contains(t, string(response.Body), "success")
}

func TestProxyExecutor_Execute_ErrorInjection(t *testing.T) {
	executor := NewProxyExecutor()

	t.Run("Always inject error", func(t *testing.T) {
		config := &ProxyConfig{
			TargetURL:       "http://example.com",
			ErrorRate:       1.0,
			ErrorStatusCode: 503,
		}

		request := &adapter.Request{
			Path: "/test",
			Metadata: map[string]interface{}{
				"method": "GET",
			},
		}

		response, err := executor.Execute(request, config)

		assert.NoError(t, err)
		assert.NotNil(t, response)
		assert.Equal(t, 503, response.StatusCode)
		assert.Contains(t, string(response.Body), "injected error")
	})

	t.Run("Never inject error", func(t *testing.T) {
		mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{"ok": true}`))
		}))
		defer mockServer.Close()

		config := &ProxyConfig{
			TargetURL: mockServer.URL,
			ErrorRate: 0.0,
		}

		request := &adapter.Request{
			Path: "/test",
			Metadata: map[string]interface{}{
				"method": "GET",
			},
		}

		response, err := executor.Execute(request, config)

		assert.NoError(t, err)
		assert.NotNil(t, response)
		assert.Equal(t, http.StatusOK, response.StatusCode)
	})

	t.Run("Default error status code", func(t *testing.T) {
		config := &ProxyConfig{
			TargetURL: "http://example.com",
			ErrorRate: 1.0,
		}

		request := &adapter.Request{
			Path: "/test",
			Metadata: map[string]interface{}{
				"method": "GET",
			},
		}

		response, err := executor.Execute(request, config)

		assert.NoError(t, err)
		assert.Equal(t, 500, response.StatusCode)
	})
}

func TestProxyExecutor_Execute_InjectDelay(t *testing.T) {
	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"ok": true}`))
	}))
	defer mockServer.Close()

	executor := NewProxyExecutor()
	config := &ProxyConfig{
		TargetURL:   mockServer.URL,
		InjectDelay: 100,
	}

	request := &adapter.Request{
		Path: "/test",
		Metadata: map[string]interface{}{
			"method": "GET",
		},
	}

	start := time.Now()
	response, err := executor.Execute(request, config)
	elapsed := time.Since(start)

	assert.NoError(t, err)
	assert.NotNil(t, response)
	assert.GreaterOrEqual(t, elapsed.Milliseconds(), int64(100))
}

func TestProxyExecutor_Execute_ModifyRequest(t *testing.T) {
	// 创建mock服务器来验证请求修改
	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// 验证请求头
		assert.Equal(t, "custom-value", r.Header.Get("X-Custom-Header"))
		assert.Empty(t, r.Header.Get("X-Remove-Header"))

		// 验证查询参数
		assert.Equal(t, "bar", r.URL.Query().Get("foo"))

		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"ok": true}`))
	}))
	defer mockServer.Close()

	executor := NewProxyExecutor()
	config := &ProxyConfig{
		TargetURL: mockServer.URL,
		ModifyRequest: &RequestModifier{
			Headers: map[string]string{
				"X-Custom-Header": "custom-value",
			},
			Query: map[string]string{
				"foo": "bar",
			},
			RemoveHeaders: []string{"X-Remove-Header"},
		},
	}

	request := &adapter.Request{
		Path: "/test",
		Headers: map[string]string{
			"X-Remove-Header": "should-be-removed",
		},
		Metadata: map[string]interface{}{
			"method": "GET",
		},
	}

	response, err := executor.Execute(request, config)

	assert.NoError(t, err)
	assert.NotNil(t, response)
	assert.Equal(t, http.StatusOK, response.StatusCode)
}

func TestProxyExecutor_Execute_ModifyRequestBody(t *testing.T) {
	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// 读取并验证请求体
		var body map[string]interface{}
		json.NewDecoder(r.Body).Decode(&body)
		
		// 验证修改后的字段
		assert.Equal(t, "modified", body["field1"])
		assert.Equal(t, "original", body["field2"])

		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"ok": true}`))
	}))
	defer mockServer.Close()

	executor := NewProxyExecutor()
	config := &ProxyConfig{
		TargetURL: mockServer.URL,
		ModifyRequest: &RequestModifier{
			Body: map[string]interface{}{
				"field1": "modified",
			},
		},
	}

	request := &adapter.Request{
		Path: "/test",
		Body: []byte(`{"field1": "original", "field2": "original"}`),
		Metadata: map[string]interface{}{
			"method": "POST",
		},
	}

	response, err := executor.Execute(request, config)

	assert.NoError(t, err)
	assert.NotNil(t, response)
}

func TestProxyExecutor_Execute_ModifyResponse(t *testing.T) {
	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("X-Original-Header", "original")
		w.Header().Set("X-Remove-Header", "should-be-removed")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"message": "original", "data": "test"}`))
	}))
	defer mockServer.Close()

	executor := NewProxyExecutor()
	config := &ProxyConfig{
		TargetURL: mockServer.URL,
		ModifyResponse: &ResponseModifier{
			StatusCode: http.StatusCreated,
			Headers: map[string]string{
				"X-Custom-Header": "custom-value",
			},
			BodyReplace: map[string]interface{}{
				"message": "modified",
			},
			RemoveHeaders: []string{"X-Remove-Header"},
		},
	}

	request := &adapter.Request{
		Path: "/test",
		Metadata: map[string]interface{}{
			"method": "GET",
		},
	}

	response, err := executor.Execute(request, config)

	assert.NoError(t, err)
	assert.NotNil(t, response)
	assert.Equal(t, http.StatusCreated, response.StatusCode)
	assert.Equal(t, "custom-value", response.Headers["X-Custom-Header"])
	assert.Empty(t, response.Headers["X-Remove-Header"])

	// 验证响应体修改
	var body map[string]interface{}
	json.Unmarshal(response.Body, &body)
	assert.Equal(t, "modified", body["message"])
	assert.Equal(t, "test", body["data"])
}

func TestProxyExecutor_Execute_Timeout(t *testing.T) {
	// 创建一个慢速服务器
	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(3 * time.Second)
		w.WriteHeader(http.StatusOK)
	}))
	defer mockServer.Close()

	executor := NewProxyExecutor()
	config := &ProxyConfig{
		TargetURL: mockServer.URL,
		Timeout:   1, // 1秒超时
	}

	request := &adapter.Request{
		Path: "/test",
		Metadata: map[string]interface{}{
			"method": "GET",
		},
	}

	_, err := executor.Execute(request, config)

	assert.Error(t, err)
}

func TestProxyExecutor_Execute_NoFollowRedirect(t *testing.T) {
	// 创建重定向服务器
	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/redirect" {
			http.Redirect(w, r, "/target", http.StatusFound)
		} else {
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{"ok": true}`))
		}
	}))
	defer mockServer.Close()

	executor := NewProxyExecutor()
	config := &ProxyConfig{
		TargetURL:      mockServer.URL,
		FollowRedirect: false,
	}

	request := &adapter.Request{
		Path: "/redirect",
		Metadata: map[string]interface{}{
			"method": "GET",
		},
	}

	response, err := executor.Execute(request, config)

	assert.NoError(t, err)
	assert.NotNil(t, response)
	assert.Equal(t, http.StatusFound, response.StatusCode)
}

func TestShouldInjectError(t *testing.T) {
	tests := []struct {
		name      string
		errorRate float64
		expected  bool
	}{
		{
			name:      "Zero error rate",
			errorRate: 0.0,
			expected:  false,
		},
		{
			name:      "Negative error rate",
			errorRate: -0.1,
			expected:  false,
		},
		{
			name:      "100% error rate",
			errorRate: 1.0,
			expected:  true,
		},
		{
			name:      "Over 100% error rate",
			errorRate: 1.5,
			expected:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := shouldInjectError(tt.errorRate)
			assert.Equal(t, tt.expected, result)
		})
	}

	// 测试中间值的随机性
	t.Run("50% error rate", func(t *testing.T) {
		errorRate := 0.5
		trueCount := 0
		iterations := 1000

		for i := 0; i < iterations; i++ {
			if shouldInjectError(errorRate) {
				trueCount++
			}
		}

		// 允许一定误差范围（40%-60%）
		assert.GreaterOrEqual(t, trueCount, 400)
		assert.LessOrEqual(t, trueCount, 600)
	})
}

func TestProxyExecutor_ModifyRequestBody(t *testing.T) {
	executor := NewProxyExecutor()

	t.Run("Modify JSON body", func(t *testing.T) {
		originalBody := []byte(`{"field1": "original", "field2": "keep"}`)
		modifications := map[string]interface{}{
			"field1": "modified",
			"field3": "new",
		}

		result, err := executor.modifyRequestBody(originalBody, modifications)

		assert.NoError(t, err)

		var resultMap map[string]interface{}
		json.Unmarshal(result, &resultMap)

		assert.Equal(t, "modified", resultMap["field1"])
		assert.Equal(t, "keep", resultMap["field2"])
		assert.Equal(t, "new", resultMap["field3"])
	})

	t.Run("Non-JSON body returns original", func(t *testing.T) {
		originalBody := []byte("plain text")
		modifications := map[string]interface{}{
			"field1": "value",
		}

		result, err := executor.modifyRequestBody(originalBody, modifications)

		assert.NoError(t, err)
		assert.Equal(t, originalBody, result)
	})
}

func TestProxyExecutor_ApplyRequestModifier(t *testing.T) {
	executor := NewProxyExecutor()

	req, _ := http.NewRequest("GET", "http://example.com?existing=value", nil)
	req.Header.Set("X-Keep", "keep")
	req.Header.Set("X-Remove", "remove")

	modifier := &RequestModifier{
		Headers: map[string]string{
			"X-Custom": "value",
		},
		Query: map[string]string{
			"new": "param",
		},
		RemoveHeaders: []string{"X-Remove"},
	}

	executor.applyRequestModifier(req, modifier)

	assert.Equal(t, "value", req.Header.Get("X-Custom"))
	assert.Equal(t, "keep", req.Header.Get("X-Keep"))
	assert.Empty(t, req.Header.Get("X-Remove"))
	assert.Equal(t, "param", req.URL.Query().Get("new"))
	assert.Equal(t, "value", req.URL.Query().Get("existing"))
}

func TestProxyExecutor_ApplyResponseModifier(t *testing.T) {
	executor := NewProxyExecutor()

	t.Run("Modify all response fields", func(t *testing.T) {
		response := &adapter.Response{
			StatusCode: 200,
			Headers: map[string]string{
				"X-Keep":   "keep",
				"X-Remove": "remove",
			},
			Body: []byte(`{"field1": "original", "field2": "keep"}`),
		}

		modifier := &ResponseModifier{
			StatusCode: 201,
			Headers: map[string]string{
				"X-Custom": "value",
			},
			BodyReplace: map[string]interface{}{
				"field1": "modified",
			},
			RemoveHeaders: []string{"X-Remove"},
		}

		err := executor.applyResponseModifier(response, modifier)

		assert.NoError(t, err)
		assert.Equal(t, 201, response.StatusCode)
		assert.Equal(t, "value", response.Headers["X-Custom"])
		assert.Equal(t, "keep", response.Headers["X-Keep"])
		assert.Empty(t, response.Headers["X-Remove"])

		var body map[string]interface{}
		json.Unmarshal(response.Body, &body)
		assert.Equal(t, "modified", body["field1"])
		assert.Equal(t, "keep", body["field2"])
	})

	t.Run("Non-JSON body not modified", func(t *testing.T) {
		response := &adapter.Response{
			StatusCode: 200,
			Headers:    map[string]string{},
			Body:       []byte("plain text"),
		}

		modifier := &ResponseModifier{
			BodyReplace: map[string]interface{}{
				"field1": "value",
			},
		}

		err := executor.applyResponseModifier(response, modifier)

		assert.NoError(t, err)
		assert.Equal(t, []byte("plain text"), response.Body)
	})
}
