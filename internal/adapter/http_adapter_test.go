package adapter

import (
	"bytes"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/gomockserver/mockserver/internal/models"
	"github.com/stretchr/testify/assert"
)

func TestNewHTTPAdapter(t *testing.T) {
	adapter := NewHTTPAdapter()
	assert.NotNil(t, adapter)
}

func TestHTTPAdapter_Parse(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name          string
		method        string
		path          string
		query         string
		headers       map[string]string
		body          string
		expectedPath  string
		expectedQuery map[string]string
	}{
		{
			name:          "GET请求解析",
			method:        "GET",
			path:          "/api/users",
			query:         "",
			headers:       map[string]string{"Content-Type": "application/json"},
			body:          "",
			expectedPath:  "/api/users",
			expectedQuery: map[string]string{},
		},
		{
			name:          "POST请求解析",
			method:        "POST",
			path:          "/api/users",
			query:         "",
			headers:       map[string]string{"Content-Type": "application/json"},
			body:          `{"name":"test"}`,
			expectedPath:  "/api/users",
			expectedQuery: map[string]string{},
		},
		{
			name:   "带Query参数的GET请求",
			method: "GET",
			path:   "/api/users",
			query:  "status=active&page=1",
			headers: map[string]string{
				"Content-Type": "application/json",
				"User-Agent":   "test-client",
			},
			body:         "",
			expectedPath: "/api/users",
			expectedQuery: map[string]string{
				"status": "active",
				"page":   "1",
			},
		},
		{
			name:   "带多个Header的请求",
			method: "POST",
			path:   "/api/orders",
			query:  "",
			headers: map[string]string{
				"Content-Type":  "application/json",
				"Authorization": "Bearer token123",
			},
			body:          `{"product":"item1"}`,
			expectedPath:  "/api/orders",
			expectedQuery: map[string]string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// 创建HTTP请求
			req := httptest.NewRequest(tt.method, tt.path+"?"+tt.query, bytes.NewBufferString(tt.body))
			for k, v := range tt.headers {
				req.Header.Set(k, v)
			}

			// 创建gin.Context
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			c.Request = req

			// 执行解析
			adapter := NewHTTPAdapter()
			result, err := adapter.Parse(c)

			// 验证
			assert.NoError(t, err)
			assert.NotNil(t, result)
			assert.NotEmpty(t, result.ID)
			assert.Equal(t, models.ProtocolHTTP, result.Protocol)
			assert.Equal(t, tt.expectedPath, result.Path)
			assert.Equal(t, []byte(tt.body), result.Body)

			// 验证Headers
			for k, v := range tt.headers {
				assert.Equal(t, v, result.Headers[k])
			}

			// 验证Metadata中的method
			method, ok := result.Metadata["method"].(string)
			assert.True(t, ok)
			assert.Equal(t, tt.method, method)

			// 验证Query参数
			query, ok := result.Metadata["query"].(map[string]string)
			assert.True(t, ok)
			for k, v := range tt.expectedQuery {
				assert.Equal(t, v, query[k])
			}

			// 验证时间戳
			assert.False(t, result.ReceivedAt.IsZero())
		})
	}
}

func TestHTTPAdapter_Parse_EmptyBody(t *testing.T) {
	gin.SetMode(gin.TestMode)

	req := httptest.NewRequest("GET", "/api/users", nil)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = req

	adapter := NewHTTPAdapter()
	result, err := adapter.Parse(c)

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Empty(t, result.Body)
}

func TestHTTPAdapter_Parse_InvalidInput(t *testing.T) {
	adapter := NewHTTPAdapter()

	// 传入非gin.Context类型
	result, err := adapter.Parse("invalid")

	assert.NoError(t, err)
	assert.Nil(t, result)
}

func TestHTTPAdapter_Build(t *testing.T) {
	adapter := NewHTTPAdapter()

	response := &Response{
		StatusCode: 200,
		Headers: map[string]string{
			"Content-Type": "application/json",
		},
		Body: []byte(`{"status":"ok"}`),
	}

	result, err := adapter.Build(response)

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, response, result)
}

func TestHTTPAdapter_WriteResponse(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name           string
		statusCode     int
		headers        map[string]string
		body           []byte
		expectedStatus int
		expectedBody   string
	}{
		{
			name:       "200 JSON响应",
			statusCode: 200,
			headers: map[string]string{
				"Content-Type": "application/json",
			},
			body:           []byte(`{"code":0,"message":"success"}`),
			expectedStatus: 200,
			expectedBody:   `{"code":0,"message":"success"}`,
		},
		{
			name:       "404错误响应",
			statusCode: 404,
			headers: map[string]string{
				"Content-Type": "application/json",
			},
			body:           []byte(`{"error":"not found"}`),
			expectedStatus: 404,
			expectedBody:   `{"error":"not found"}`,
		},
		{
			name:       "自定义Header响应",
			statusCode: 200,
			headers: map[string]string{
				"Content-Type":    "text/plain",
				"X-Custom-Header": "custom-value",
			},
			body:           []byte("plain text response"),
			expectedStatus: 200,
			expectedBody:   "plain text response",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)

			response := &Response{
				StatusCode: tt.statusCode,
				Headers:    tt.headers,
				Body:       tt.body,
			}

			adapter := NewHTTPAdapter()
			adapter.WriteResponse(c, response)

			// 验证状态码
			assert.Equal(t, tt.expectedStatus, w.Code)

			// 验证响应体
			assert.Equal(t, tt.expectedBody, w.Body.String())

			// 验证Headers
			for k, v := range tt.headers {
				assert.Equal(t, v, w.Header().Get(k))
			}
		})
	}
}

func TestGetContentType(t *testing.T) {
	tests := []struct {
		name     string
		headers  map[string]string
		expected string
	}{
		{
			name: "标准Content-Type",
			headers: map[string]string{
				"Content-Type": "application/json",
			},
			expected: "application/json",
		},
		{
			name: "小写content-type",
			headers: map[string]string{
				"content-type": "text/html",
			},
			expected: "text/html",
		},
		{
			name: "混合大小写",
			headers: map[string]string{
				"CoNtEnT-TyPe": "text/xml",
			},
			expected: "text/xml",
		},
		{
			name:     "没有Content-Type",
			headers:  map[string]string{},
			expected: "application/json",
		},
		{
			name: "有其他Header但没有Content-Type",
			headers: map[string]string{
				"Authorization": "Bearer token",
			},
			expected: "application/json",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := getContentType(tt.headers)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestHTTPAdapter_Parse_ClientIP(t *testing.T) {
	gin.SetMode(gin.TestMode)

	req := httptest.NewRequest("GET", "/api/test", nil)
	req.RemoteAddr = "192.168.1.100:12345"

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = req

	adapter := NewHTTPAdapter()
	result, err := adapter.Parse(c)

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.NotEmpty(t, result.SourceIP)
}

func TestHTTPAdapter_Parse_Metadata(t *testing.T) {
	gin.SetMode(gin.TestMode)

	req := httptest.NewRequest("POST", "/api/users?page=1", bytes.NewBufferString(`{"name":"test"}`))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("User-Agent", "Mozilla/5.0")
	req.Host = "example.com"

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = req

	adapter := NewHTTPAdapter()
	result, err := adapter.Parse(c)

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.NotNil(t, result.Metadata)

	// 验证Metadata字段
	assert.Equal(t, "POST", result.Metadata["method"])
	assert.Equal(t, "example.com", result.Metadata["host"])
	assert.Equal(t, "Mozilla/5.0", result.Metadata["user_agent"])
	assert.Equal(t, "application/json", result.Metadata["content_type"])
	assert.Equal(t, "page=1", result.Metadata["raw_query"])

	query := result.Metadata["query"].(map[string]string)
	assert.Equal(t, "1", query["page"])
}
