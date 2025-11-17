package executor

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/gomockserver/mockserver/internal/adapter"
	"github.com/gomockserver/mockserver/internal/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestTemplateEngine_Render(t *testing.T) {
	engine := NewTemplateEngine()

	tests := []struct {
		name     string
		template string
		context  *TemplateContext
		want     string
		wantErr  bool
	}{
		{
			name:     "simple text",
			template: "Hello, World!",
			context:  &TemplateContext{},
			want:     "Hello, World!",
			wantErr:  false,
		},
		{
			name:     "uuid function",
			template: "Request ID: {{uuid}}",
			context:  &TemplateContext{},
			want:     "",  // UUID is random
			wantErr:  false,
		},
		{
			name:     "timestamp function",
			template: "Timestamp: {{timestamp}}",
			context:  &TemplateContext{},
			want:     "",  // timestamp is dynamic
			wantErr:  false,
		},
		{
			name:     "request path",
			template: "Path: {{.Request.Path}}",
			context: &TemplateContext{
				Request: &RequestContext{
					Path: "/api/users",
				},
			},
			want:    "Path: /api/users",
			wantErr: false,
		},
		{
			name:     "request method",
			template: "Method: {{.Request.Method}}",
			context: &TemplateContext{
				Request: &RequestContext{
					Method: "POST",
				},
			},
			want:    "Method: POST",
			wantErr: false,
		},
		{
			name:     "random function",
			template: "Random: {{random 1 100}}",
			context:  &TemplateContext{},
			want:     "",  // random value
			wantErr:  false,
		},
		{
			name:     "randomString function",
			template: "Token: {{randomString 8}}",
			context:  &TemplateContext{},
			want:     "",  // random string
			wantErr:  false,
		},
		{
			name:     "base64 encoding",
			template: "Encoded: {{base64 \"hello\"}}",
			context:  &TemplateContext{},
			want:     "Encoded: aGVsbG8=",
			wantErr:  false,
		},
		{
			name:     "conditional rendering",
			template: "{{if eq .Request.Method \"POST\"}}Creating{{else}}Reading{{end}}",
			context: &TemplateContext{
				Request: &RequestContext{
					Method: "POST",
				},
			},
			want:    "Creating",
			wantErr: false,
		},
		{
			name:     "concat function",
			template: "{{concat \"Hello\" \" \" \"World\"}}",
			context:  &TemplateContext{},
			want:     "Hello World",
			wantErr:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := engine.Render(tt.template, tt.context)
			if tt.wantErr {
				assert.Error(t, err)
				return
			}
			require.NoError(t, err)
			
			if tt.want != "" {
				assert.Equal(t, tt.want, got)
			} else {
				// For dynamic content, just check it's not empty
				assert.NotEmpty(t, got)
			}
		})
	}
}

func TestTemplateEngine_RenderJSON(t *testing.T) {
	engine := NewTemplateEngine()

	tests := []struct {
		name     string
		template interface{}
		context  *TemplateContext
		check    func(t *testing.T, result interface{})
		wantErr  bool
	}{
		{
			name: "simple json object",
			template: map[string]interface{}{
				"message": "Hello, {{.Request.Path}}",
				"status":  "success",
			},
			context: &TemplateContext{
				Request: &RequestContext{
					Path: "/api/users",
				},
			},
			check: func(t *testing.T, result interface{}) {
				m, ok := result.(map[string]interface{})
				require.True(t, ok)
				assert.Equal(t, "Hello, /api/users", m["message"])
				assert.Equal(t, "success", m["status"])
			},
			wantErr: false,
		},
		{
			name: "with uuid and timestamp",
			template: map[string]interface{}{
				"request_id": "{{uuid}}",
				"timestamp":  "{{timestamp}}",
				"path":       "{{.Request.Path}}",
			},
			context: &TemplateContext{
				Request: &RequestContext{
					Path: "/api/products",
				},
			},
			check: func(t *testing.T, result interface{}) {
				m, ok := result.(map[string]interface{})
				require.True(t, ok)
				assert.NotEmpty(t, m["request_id"])
				assert.NotEmpty(t, m["timestamp"])
				assert.Equal(t, "/api/products", m["path"])
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := engine.RenderJSON(tt.template, tt.context)
			if tt.wantErr {
				assert.Error(t, err)
				return
			}
			require.NoError(t, err)
			tt.check(t, got)
		})
	}
}

func TestMockExecutor_DynamicResponse(t *testing.T) {
	executor := NewMockExecutor()

	tests := []struct {
		name    string
		request *adapter.Request
		rule    *models.Rule
		check   func(t *testing.T, resp *adapter.Response)
		wantErr bool
	}{
		{
			name: "dynamic json response",
			request: &adapter.Request{
				Path: "/api/users",
				Metadata: map[string]interface{}{
					"method": "POST",
				},
				Body: []byte(`{"name":"张三"}`),
			},
			rule: &models.Rule{
				ID:       "rule1",
				Name:     "User Creation",
				Protocol: models.ProtocolHTTP,
				Priority: 100,
				Response: models.Response{
					Type: models.ResponseTypeDynamic,
					Content: map[string]interface{}{
						"status_code": 200,
						"content_type": "JSON",
						"body": map[string]interface{}{
							"request_id": "{{uuid}}",
							"timestamp":  "{{timestamp}}",
							"user": map[string]interface{}{
								"name":       "{{.Request.Body.name}}",
								"created_at": "{{now \"2006-01-02 15:04:05\"}}",
							},
							"path": "{{.Request.Path}}",
						},
					},
				},
			},
			check: func(t *testing.T, resp *adapter.Response) {
				assert.Equal(t, 200, resp.StatusCode)
				assert.Equal(t, "application/json", resp.Headers["Content-Type"])

				var body map[string]interface{}
				err := json.Unmarshal(resp.Body, &body)
				require.NoError(t, err)

				assert.NotEmpty(t, body["request_id"])
				assert.NotEmpty(t, body["timestamp"])
				assert.Equal(t, "/api/users", body["path"])

				user, ok := body["user"].(map[string]interface{})
				require.True(t, ok)
				assert.Equal(t, "张三", user["name"])
				assert.NotEmpty(t, user["created_at"])
			},
			wantErr: false,
		},
		{
			name: "dynamic text response",
			request: &adapter.Request{
				Path: "/api/greeting",
				Metadata: map[string]interface{}{
					"method": "GET",
				},
			},
			rule: &models.Rule{
				ID:       "rule2",
				Name:     "Greeting",
				Protocol: models.ProtocolHTTP,
				Priority: 100,
				Response: models.Response{
					Type: models.ResponseTypeDynamic,
					Content: map[string]interface{}{
						"status_code":  200,
						"content_type": "Text",
						"body":         "Hello! Current time: {{now \"15:04:05\"}}",
					},
				},
			},
			check: func(t *testing.T, resp *adapter.Response) {
				assert.Equal(t, 200, resp.StatusCode)
				assert.Contains(t, string(resp.Body), "Hello! Current time:")
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := executor.dynamicResponse(tt.request, tt.rule, nil)
			if tt.wantErr {
				assert.Error(t, err)
				return
			}
			require.NoError(t, err)
			tt.check(t, got)
		})
	}
}

func TestTemplateEngine_BuildContext(t *testing.T) {
	engine := NewTemplateEngine()

	request := &adapter.Request{
		Path:     "/api/users/123",
		Headers:  map[string]string{"Authorization": "Bearer token"},
		Body:     []byte(`{"name":"test"}`),
		SourceIP: "192.168.1.100",
		Metadata: map[string]interface{}{
			"method": "POST",
			"query": map[string]string{
				"page": "1",
			},
		},
	}

	rule := &models.Rule{
		ID:       "rule_123",
		Name:     "Test Rule",
		Priority: 100,
	}

	env := &models.Environment{
		Variables: map[string]interface{}{
			"base_url": "http://localhost:9090",
			"version":  "v1",
		},
	}

	ctx := engine.BuildContext(request, rule, env)

	assert.NotNil(t, ctx)
	assert.NotNil(t, ctx.Request)
	assert.NotNil(t, ctx.Rule)
	assert.NotNil(t, ctx.Environment)

	// Check request context
	assert.Equal(t, "POST", ctx.Request.Method)
	assert.Equal(t, "/api/users/123", ctx.Request.Path)
	assert.Equal(t, "192.168.1.100", ctx.Request.IP)
	assert.Equal(t, "Bearer token", ctx.Request.Headers["Authorization"])
	assert.Equal(t, "1", ctx.Request.Query["page"])

	// Check body parsing
	bodyMap, ok := ctx.Request.Body.(map[string]interface{})
	require.True(t, ok)
	assert.Equal(t, "test", bodyMap["name"])

	// Check rule context
	assert.Equal(t, "rule_123", ctx.Rule.ID)
	assert.Equal(t, "Test Rule", ctx.Rule.Name)
	assert.Equal(t, 100, ctx.Rule.Priority)

	// Check environment context
	assert.Equal(t, "http://localhost:9090", ctx.Environment.Variables["base_url"])
	assert.Equal(t, "v1", ctx.Environment.Variables["version"])
}

func TestTemplateEngine_BuiltInFunctions(t *testing.T) {
	engine := NewTemplateEngine()
	ctx := &TemplateContext{}

	t.Run("timestamp", func(t *testing.T) {
		result, err := engine.Render("{{timestamp}}", ctx)
		require.NoError(t, err)
		assert.NotEmpty(t, result)
		
		// Parse and verify it's a valid timestamp
		var ts int64
		err = json.Unmarshal([]byte(result), &ts)
		require.NoError(t, err)
		assert.Greater(t, ts, int64(0))
	})

	t.Run("uuid", func(t *testing.T) {
		result, err := engine.Render("{{uuid}}", ctx)
		require.NoError(t, err)
		assert.Len(t, result, 36) // UUID format: xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx
	})

	t.Run("now with format", func(t *testing.T) {
		result, err := engine.Render("{{now \"2006-01-02\"}}", ctx)
		require.NoError(t, err)
		
		// Verify date format
		_, err = time.Parse("2006-01-02", result)
		require.NoError(t, err)
	})

	t.Run("random range", func(t *testing.T) {
		result, err := engine.Render("{{random 10 20}}", ctx)
		require.NoError(t, err)
		
		var num int
		err = json.Unmarshal([]byte(result), &num)
		require.NoError(t, err)
		assert.GreaterOrEqual(t, num, 10)
		assert.Less(t, num, 20)
	})

	t.Run("randomString length", func(t *testing.T) {
		result, err := engine.Render("{{randomString 10}}", ctx)
		require.NoError(t, err)
		assert.Len(t, result, 10)
	})
}
