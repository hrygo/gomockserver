package api

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGraphQLHandler_HandleGraphQL_POST(t *testing.T) {
	handler := NewGraphQLHandler()

	router := gin.New()
	handler.RegisterRoutes(router)

	tests := []struct {
		name           string
		requestBody    string
		expectedStatus int
		expectData      bool
		expectError     bool
	}{
		{
			name:           "基础查询",
			requestBody:    `{"query": "{ hello }"}`,
			expectedStatus: http.StatusOK,
			expectData:      true,
			expectError:     false,
		},
		{
			name:           "查询用户",
			requestBody:    `{"query": "{ user }"}`,
			expectedStatus: http.StatusOK,
			expectData:      true,
			expectError:     false,
		},
		{
			name:           "查询多个字段",
			requestBody:    `{"query": "{ hello status user }"}`,
			expectedStatus: http.StatusOK,
			expectData:      true,
			expectError:     false,
		},
		{
			name:           "带变量的查询",
			requestBody:    `{"query": "query GetUser($id: ID!) { user(id: $id) { id name } }", "variables": {"id": "test"}}`,
			expectedStatus: http.StatusOK,
			expectData:      true,
			expectError:     false,
		},
		{
			name:           "空查询",
			requestBody:    `{"query": ""}`,
			expectedStatus: http.StatusBadRequest,
			expectData:      false,
			expectError:     true,
		},
		{
			name:           "无效JSON",
			requestBody:    `{"query": "{ hello }",}`,
			expectedStatus: http.StatusBadRequest,
			expectData:      false,
			expectError:     true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest("POST", "/graphql", bytes.NewBufferString(tt.requestBody))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()

			router.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)

			var response map[string]interface{}
			err := json.Unmarshal(w.Body.Bytes(), &response)
			require.NoError(t, err)

			if tt.expectData {
				assert.Contains(t, response, "data")
				assert.NotNil(t, response["data"])
			}

			if tt.expectError {
				assert.Contains(t, response, "errors")
				assert.NotEmpty(t, response["errors"])
			}
		})
	}
}

func TestGraphQLHandler_HandleGraphQL_GET(t *testing.T) {
	handler := NewGraphQLHandler()

	router := gin.New()
	handler.RegisterRoutes(router)

	tests := []struct {
		name           string
		queryParam      string
		expectedStatus int
		expectData      bool
	}{
		{
			name:           "GET基础查询",
			queryParam:      "{ hello }",
			expectedStatus: http.StatusOK,
			expectData:      true,
		},
		{
			name:           "GET查询用户",
			queryParam:      "{ user }",
			expectedStatus: http.StatusOK,
			expectData:      true,
		},
		{
			name:           "GET带变量查询",
			queryParam:      `query GetUser($id: ID!) { user(id: $id) { id name } }`,
			expectedStatus: http.StatusOK,
			expectData:      true,
		},
		{
			name:           "GET空查询",
			queryParam:      "",
			expectedStatus: http.StatusBadRequest,
			expectData:      false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest("GET", "/graphql?query="+url.QueryEscape(tt.queryParam), nil)
			w := httptest.NewRecorder()

			router.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)

			var response map[string]interface{}
			err := json.Unmarshal(w.Body.Bytes(), &response)
			require.NoError(t, err)

			if tt.expectData {
				assert.Contains(t, response, "data")
			} else {
				assert.Contains(t, response, "errors")
			}
		})
	}
}

func TestGraphQLHandler_HandleGraphQL_UnsupportedMethod(t *testing.T) {
	handler := NewGraphQLHandler()

	router := gin.New()
	handler.RegisterRoutes(router)

	req := httptest.NewRequest("PUT", "/graphql", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusMethodNotAllowed, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)
	assert.Contains(t, response, "errors")
}

func TestGraphQLHandler_HandlePlayground(t *testing.T) {
	handler := NewGraphQLHandler()

	router := gin.New()
	handler.RegisterRoutes(router)

	req := httptest.NewRequest("GET", "/graphql-playground", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), "GraphQL Playground")
	assert.Contains(t, w.Body.String(), "MockServer")
}

func TestGraphQLHandler_HandleSchemaIntrospection(t *testing.T) {
	handler := NewGraphQLHandler()

	router := gin.New()
	handler.RegisterRoutes(router)

	req := httptest.NewRequest("GET", "/graphql/schema", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)
	assert.Contains(t, response, "data")
}

func TestGraphQLHandler_HandleSchemaIntrospection_SDL(t *testing.T) {
	handler := NewGraphQLHandler()

	router := gin.New()
	handler.RegisterRoutes(router)

	req := httptest.NewRequest("GET", "/graphql/schema?sdl=true", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)
	assert.Contains(t, response, "sdl")
	assert.Contains(t, response["sdl"], "type Query")
}

func TestGraphQLHandler_HandleHealth(t *testing.T) {
	handler := NewGraphQLHandler()

	router := gin.New()
	handler.RegisterRoutes(router)

	req := httptest.NewRequest("GET", "/graphql/health", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	assert.Equal(t, "healthy", response["status"])
	assert.Equal(t, "MockServer GraphQL", response["service"])
	assert.Equal(t, "0.8.0", response["version"])
}

func TestGraphQLHandler_ParseVariables(t *testing.T) {
	handler := NewGraphQLHandler()

	// 测试有效的JSON变量
	variablesStr := `{"id": "123", "name": "test"}`
	result := handler.parseVariables(variablesStr)
	assert.Equal(t, "123", result["id"])
	assert.Equal(t, "test", result["name"])

	// 测试空变量
	emptyStr := ""
	result = handler.parseVariables(emptyStr)
	assert.Empty(t, result)

	// 测试无效JSON
	invalidStr := `{"id": 123, "name":}`
	result = handler.parseVariables(invalidStr)
	assert.Empty(t, result)
}

func TestGraphQLHandler_GetHeaders(t *testing.T) {
	handler := NewGraphQLHandler()

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest("GET", "/", nil)
	c.Request.Header.Set("Content-Type", "application/json")
	c.Request.Header.Set("User-Agent", "test-agent")

	headers := handler.getHeaders(c)

	assert.Equal(t, "application/json", headers["Content-Type"])
	assert.Equal(t, "test-agent", headers["User-Agent"])
}

func TestGraphQLHandler_GetMetadata(t *testing.T) {
	handler := NewGraphQLHandler()

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest("POST", "/test", nil)
	c.Request.Header.Set("User-Agent", "test-agent")
	c.Request.RemoteAddr = "127.0.0.1:8000"

	metadata := handler.getMetadata(c)

	assert.Equal(t, "127.0.0.1", metadata["clientIP"])
	assert.Equal(t, "test-agent", metadata["userAgent"])
	assert.Equal(t, "POST", metadata["method"])
	assert.Equal(t, "/test", metadata["path"])
}

func TestGraphQLHandler_InferOperationType(t *testing.T) {
	handler := NewGraphQLHandler()

	tests := []struct {
		query string
		want  string
	}{
		{
			query: "{ hello }",
			want:  "QUERY",
		},
		{
			query: "mutation { createUser }",
			want:  "MUTATION",
		},
		{
			query: "subscription { userUpdated }",
			want:  "SUBSCRIPTION",
		},
		{
			query: "query GetUser { user(id: $id) { name } }",
			want:  "QUERY",
		},
	}

	for _, tt := range tests {
		t.Run(tt.query, func(t *testing.T) {
			result := handler.inferOperationType(tt.query)
			assert.Equal(t, tt.want, result)
		})
	}
}

func TestGraphQLResponse_Extensions(t *testing.T) {
	handler := NewGraphQLHandler()

	router := gin.New()
	handler.RegisterRoutes(router)

	req := httptest.NewRequest("POST", "/graphql", bytes.NewBufferString(`{"query": "{ hello }"}`))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	// 检查是否有扩展信息
	if extensions, ok := response["extensions"].(map[string]interface{}); ok {
		assert.Contains(t, extensions, "timestamp")
		assert.Contains(t, extensions, "requestId")
	}
}