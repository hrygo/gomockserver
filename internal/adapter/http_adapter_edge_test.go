package adapter

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/gomockserver/mockserver/internal/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestHTTPAdapter_Parse_EdgeCases 测试HTTP适配器的边界情况
func TestHTTPAdapter_Parse_EdgeCases(t *testing.T) {
	adapter := NewHTTPAdapter()

	tests := []struct {
		name         string
		setupRequest func() *http.Request
		expectedPath string
		shouldError  bool
	}{
		{
			name: "空路径",
			setupRequest: func() *http.Request {
				return httptest.NewRequest("GET", "/", nil)
			},
			expectedPath: "/",
			shouldError:  false,
		},
		{
			name: "非常长的路径",
			setupRequest: func() *http.Request {
				longPath := "/" + strings.Repeat("segment/", 50) + "end"
				return httptest.NewRequest("GET", longPath, nil)
			},
			expectedPath: "/" + strings.Repeat("segment/", 50) + "end",
			shouldError:  false,
		},
		{
			name: "包含Unicode字符的路径",
			setupRequest: func() *http.Request {
				return httptest.NewRequest("GET", "/api/测试/🚀.json", nil)
			},
			expectedPath: "/api/测试/🚀.json",
			shouldError:  false,
		},
		{
			name: "多个查询参数",
			setupRequest: func() *http.Request {
				return httptest.NewRequest("GET", "/api/test?param1=value1&param2=value2&param3=value3", nil)
			},
			expectedPath: "/api/test",
			shouldError:  false,
		},
		{
			name: "重复的查询参数",
			setupRequest: func() *http.Request {
				return httptest.NewRequest("GET", "/api/test?param=value1&param=value2", nil)
			},
			expectedPath: "/api/test",
			shouldError:  false,
		},
		{
			name: "空请求体",
			setupRequest: func() *http.Request {
				return httptest.NewRequest("POST", "/api/test", bytes.NewReader([]byte{}))
			},
			expectedPath: "/api/test",
			shouldError:  false,
		},
		{
			name: "大请求体",
			setupRequest: func() *http.Request {
				largeBody := strings.Repeat("x", 1024*1024) // 1MB
				return httptest.NewRequest("POST", "/api/test", strings.NewReader(largeBody))
			},
			expectedPath: "/api/test",
			shouldError:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gin.SetMode(gin.TestMode)
			c, _ := gin.CreateTestContext(httptest.NewRecorder())
			c.Request = tt.setupRequest()
			c.Params = gin.Params{
				gin.Param{Key: "projectID", Value: "test-project"},
				gin.Param{Key: "environmentID", Value: "test-env"},
				gin.Param{Key: "path", Value: tt.expectedPath},
			}

			result, err := adapter.Parse(c)

			if tt.shouldError {
				assert.Error(t, err)
				assert.Nil(t, result)
			} else {
				assert.NoError(t, err)
				require.NotNil(t, result)

				assert.Equal(t, models.ProtocolHTTP, result.Protocol)
				assert.Equal(t, tt.expectedPath, result.Path)
				assert.NotEmpty(t, result.ID)
				assert.NotEmpty(t, result.SourceIP)
				assert.NotNil(t, result.ReceivedAt)

				// 验证元数据
				assert.NotNil(t, result.Metadata)
				assert.Equal(t, c.Request.Method, result.Metadata["method"])
				assert.Equal(t, c.Request.Host, result.Metadata["host"])
				assert.NotNil(t, result.Metadata["query"])
			}
		})
	}
}

// TestHTTPAdapter_Parse_ErrorHandling 测试错误处理
func TestHTTPAdapter_Parse_ErrorHandling(t *testing.T) {
	adapter := NewHTTPAdapter()

	tests := []struct {
		name       string
		rawRequest interface{}
		expectNil  bool
	}{
		{
			name:       "nil输入",
			rawRequest: nil,
			expectNil:  true,
		},
		{
			name:       "字符串类型错误",
			rawRequest: "not a gin.Context",
			expectNil:  true,
		},
		{
			name:       "结构体类型错误",
			rawRequest: struct{}{},
			expectNil:  true,
		},
		{
			name:       "整数类型错误",
			rawRequest: 123,
			expectNil:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := adapter.Parse(tt.rawRequest)

			// 对于非gin.Context输入，应该返回nil但不报错
			assert.NoError(t, err)
			if tt.expectNil {
				assert.Nil(t, result)
			}
		})
	}
}

// TestHTTPAdapter_WriteResponse_EdgeCases 测试响应写入的边界情况
func TestHTTPAdapter_WriteResponse_EdgeCases(t *testing.T) {
	adapter := NewHTTPAdapter()

	tests := []struct {
		name           string
		response       *Response
		expectedStatus int
		expectedBody   string
	}{
		{
			name: "空响应体",
			response: &Response{
				StatusCode: 200,
				Headers:    map[string]string{},
				Body:       []byte{},
			},
			expectedStatus: 200,
			expectedBody:   "",
		},
		{
			name: "nil响应体",
			response: &Response{
				StatusCode: 204,
				Headers:    map[string]string{},
				Body:       nil,
			},
			expectedStatus: 204,
			expectedBody:   "",
		},
		{
			name: "JSON响应无Content-Type",
			response: &Response{
				StatusCode: 200,
				Headers:    map[string]string{},
				Body:       []byte(`{"message": "test"}`),
			},
			expectedStatus: 200,
			expectedBody:   `{"message": "test"}`,
		},
		{
			name: "包含特殊字符的响应",
			response: &Response{
				StatusCode: 200,
				Headers: map[string]string{
					"Content-Type": "text/plain; charset=utf-8",
				},
				Body: []byte("测试响应🚀special chars: &<>\"'"),
			},
			expectedStatus: 200,
			expectedBody:   "测试响应🚀special chars: &<>\"'",
		},
		{
			name: "二进制响应",
			response: &Response{
				StatusCode: 200,
				Headers: map[string]string{
					"Content-Type": "application/octet-stream",
				},
				Body: []byte{0x00, 0x01, 0x02, 0xFF, 0xFE},
			},
			expectedStatus: 200,
			expectedBody:   "\x00\x01\x02\xFF\xFE",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gin.SetMode(gin.TestMode)
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)

			adapter.WriteResponse(c, tt.response)

			assert.Equal(t, tt.expectedStatus, w.Code)
			assert.Equal(t, tt.expectedBody, w.Body.String())

			// 验证Content-Type默认设置
			if _, exists := tt.response.Headers["Content-Type"]; !exists {
				assert.Equal(t, "application/json", w.Header().Get("Content-Type"))
			}
		})
	}
}

// TestHTTPAdapter_BuildResponse 测试构建响应
func TestHTTPAdapter_BuildResponse(t *testing.T) {
	adapter := NewHTTPAdapter()

	response := &Response{
		StatusCode: 200,
		Headers: map[string]string{
			"Content-Type": "application/json",
		},
		Body: []byte(`{"test": "value"}`),
	}

	result, err := adapter.Build(response)
	assert.NoError(t, err)
	assert.Equal(t, response, result)
}

// TestHTTPAdapter_ComplexHeaders 测试复杂头部处理
func TestHTTPAdapter_ComplexHeaders(t *testing.T) {
	adapter := NewHTTPAdapter()

	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	// 创建带有复杂头部的请求
	req := httptest.NewRequest("GET", "/api/test", nil)
	req.Header.Set("X-Custom-Header", "value1")
	req.Header.Add("X-Custom-Header", "value2") // 多值头部
	req.Header.Set("Authorization", "Bearer token123")
	req.Header.Set("User-Agent", "TestAgent/1.0")
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Special-Chars", "特殊字符🚀test")

	c.Request = req
	c.Params = gin.Params{
		gin.Param{Key: "path", Value: "/api/test"},
	}

	result, err := adapter.Parse(c)
	assert.NoError(t, err)
	require.NotNil(t, result)

	// 验证头部解析（应该只取第一个值）
	assert.Equal(t, "value1", result.Headers["X-Custom-Header"])
	assert.Equal(t, "Bearer token123", result.Headers["Authorization"])
	assert.Equal(t, "TestAgent/1.0", result.Headers["User-Agent"])
	assert.Equal(t, "application/json", result.Headers["Content-Type"])
	assert.Equal(t, "特殊字符🚀test", result.Headers["X-Special-Chars"])
}

// TestHTTPAdapter_EmptyAndSpecialValues 测试空值和特殊值处理
func TestHTTPAdapter_EmptyAndSpecialValues(t *testing.T) {
	adapter := NewHTTPAdapter()

	tests := []struct {
		name   string
		path   string
		query  string
		expect string
	}{
		{
			name:   "空查询参数",
			path:   "/api/test",
			query:  "",
			expect: "/api/test",
		},
		{
			name:   "空值查询参数",
			path:   "/api/test",
			query:  "param=",
			expect: "/api/test",
		},
		{
			name:   "URL编码的路径",
			path:   "/api/test%20path",
			query:  "",
			expect: "/api/test%20path", // 适配器不自动解码
		},
		{
			name:   "查询参数包含特殊字符",
			path:   "/api/test",
			query:  "key=特殊字符&emoji=🚀",
			expect: "/api/test",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gin.SetMode(gin.TestMode)
			c, _ := gin.CreateTestContext(httptest.NewRecorder())

			url := tt.path
			if tt.query != "" {
				url += "?" + tt.query
			}

			req := httptest.NewRequest("GET", url, nil)
			c.Request = req
			c.Params = gin.Params{
				gin.Param{Key: "path", Value: tt.path},
			}

			result, err := adapter.Parse(c)
			assert.NoError(t, err)
			require.NotNil(t, result)
			assert.Equal(t, tt.expect, result.Path)
		})
	}
}
