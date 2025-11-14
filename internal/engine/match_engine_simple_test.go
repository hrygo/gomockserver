package engine

import (
	"testing"

	"github.com/gomockserver/mockserver/internal/adapter"
	"github.com/gomockserver/mockserver/internal/models"
	"github.com/stretchr/testify/assert"
)

// TestMatchMethod 测试HTTP方法匹配
func TestMatchMethod(t *testing.T) {
	tests := []struct {
		name            string
		requestMethod   string
		conditionMethod interface{}
		expected        bool
	}{
		{
			name:            "单个方法精确匹配",
			requestMethod:   "GET",
			conditionMethod: "GET",
			expected:        true,
		},
		{
			name:            "单个方法不匹配",
			requestMethod:   "POST",
			conditionMethod: "GET",
			expected:        false,
		},
		{
			name:            "方法数组匹配",
			requestMethod:   "POST",
			conditionMethod: []interface{}{"GET", "POST"},
			expected:        true,
		},
		{
			name:            "方法数组不匹配",
			requestMethod:   "DELETE",
			conditionMethod: []interface{}{"GET", "POST"},
			expected:        false,
		},
		{
			name:            "大小写不敏感",
			requestMethod:   "get",
			conditionMethod: "GET",
			expected:        true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := matchMethod(tt.requestMethod, tt.conditionMethod)
			assert.Equal(t, tt.expected, result, tt.name)
		})
	}
}

// TestMatchPath 测试路径匹配
func TestMatchPath(t *testing.T) {
	tests := []struct {
		name          string
		requestPath   string
		conditionPath string
		expected      bool
	}{
		{
			name:          "精确路径匹配",
			requestPath:   "/api/users",
			conditionPath: "/api/users",
			expected:      true,
		},
		{
			name:          "路径不匹配",
			requestPath:   "/api/users",
			conditionPath: "/api/products",
			expected:      false,
		},
		{
			name:          "路径参数匹配",
			requestPath:   "/api/users/123",
			conditionPath: "/api/users/:id",
			expected:      true,
		},
		{
			name:          "多级路径参数匹配",
			requestPath:   "/api/v1/users/123",
			conditionPath: "/api/:version/users/:id",
			expected:      true,
		},
		{
			name:          "路径段数不匹配",
			requestPath:   "/api/users/123/profile",
			conditionPath: "/api/users/:id",
			expected:      false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := matchPath(tt.requestPath, tt.conditionPath)
			assert.Equal(t, tt.expected, result, tt.name)
		})
	}
}

// TestMatchQuery 测试查询参数匹配
func TestMatchQuery(t *testing.T) {
	tests := []struct {
		name           string
		requestQuery   map[string]string
		conditionQuery map[string]string
		expected       bool
	}{
		{
			name: "单个参数匹配",
			requestQuery: map[string]string{
				"status": "active",
			},
			conditionQuery: map[string]string{
				"status": "active",
			},
			expected: true,
		},
		{
			name: "多个参数全部匹配",
			requestQuery: map[string]string{
				"status": "active",
				"role":   "admin",
			},
			conditionQuery: map[string]string{
				"status": "active",
				"role":   "admin",
			},
			expected: true,
		},
		{
			name: "参数值不匹配",
			requestQuery: map[string]string{
				"status": "inactive",
			},
			conditionQuery: map[string]string{
				"status": "active",
			},
			expected: false,
		},
		{
			name: "缺少必需参数",
			requestQuery: map[string]string{
				"role": "admin",
			},
			conditionQuery: map[string]string{
				"status": "active",
			},
			expected: false,
		},
		{
			name:           "空条件匹配",
			requestQuery:   map[string]string{"status": "active"},
			conditionQuery: map[string]string{},
			expected:       true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := matchQuery(tt.requestQuery, tt.conditionQuery)
			assert.Equal(t, tt.expected, result, tt.name)
		})
	}
}

// TestMatchHeaders 测试Header匹配
func TestMatchHeaders(t *testing.T) {
	tests := []struct {
		name             string
		requestHeaders   map[string]string
		conditionHeaders map[string]string
		expected         bool
	}{
		{
			name: "Content-Type匹配(不区分大小写)",
			requestHeaders: map[string]string{
				"content-type": "application/json",
			},
			conditionHeaders: map[string]string{
				"Content-Type": "application/json",
			},
			expected: true,
		},
		{
			name: "自定义Header匹配",
			requestHeaders: map[string]string{
				"X-API-Key": "secret123",
			},
			conditionHeaders: map[string]string{
				"X-API-Key": "secret123",
			},
			expected: true,
		},
		{
			name: "Header值不匹配",
			requestHeaders: map[string]string{
				"X-API-Key": "wrong-key",
			},
			conditionHeaders: map[string]string{
				"X-API-Key": "secret123",
			},
			expected: false,
		},
		{
			name: "多个Header匹配",
			requestHeaders: map[string]string{
				"Content-Type": "application/json",
				"X-API-Key":    "secret123",
			},
			conditionHeaders: map[string]string{
				"Content-Type": "application/json",
				"X-API-Key":    "secret123",
			},
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := matchHeaders(tt.requestHeaders, tt.conditionHeaders)
			assert.Equal(t, tt.expected, result, tt.name)
		})
	}
}

// TestSimpleMatch 测试简单匹配引擎
func TestSimpleMatch(t *testing.T) {
	engine := &MatchEngine{}

	tests := []struct {
		name     string
		request  *adapter.Request
		rule     *models.Rule
		expected bool
	}{
		{
			name: "完整HTTP请求匹配",
			request: &adapter.Request{
				Protocol: models.ProtocolHTTP,
				Path:     "/api/users",
				Headers: map[string]string{
					"Content-Type": "application/json",
				},
				Metadata: map[string]interface{}{
					"method": "GET",
					"query": map[string]string{
						"status": "active",
					},
				},
			},
			rule: &models.Rule{
				Protocol:  models.ProtocolHTTP,
				MatchType: models.MatchTypeSimple,
				MatchCondition: map[string]interface{}{
					"method": "GET",
					"path":   "/api/users",
					"query": map[string]interface{}{
						"status": "active",
					},
					"headers": map[string]interface{}{
						"Content-Type": "application/json",
					},
				},
			},
			expected: true,
		},
		{
			name: "路径参数匹配",
			request: &adapter.Request{
				Protocol: models.ProtocolHTTP,
				Path:     "/api/users/123",
				Metadata: map[string]interface{}{
					"method": "GET",
				},
			},
			rule: &models.Rule{
				Protocol:  models.ProtocolHTTP,
				MatchType: models.MatchTypeSimple,
				MatchCondition: map[string]interface{}{
					"method": "GET",
					"path":   "/api/users/:id",
				},
			},
			expected: true,
		},
		{
			name: "方法不匹配",
			request: &adapter.Request{
				Protocol: models.ProtocolHTTP,
				Path:     "/api/users",
				Metadata: map[string]interface{}{
					"method": "POST",
				},
			},
			rule: &models.Rule{
				Protocol:  models.ProtocolHTTP,
				MatchType: models.MatchTypeSimple,
				MatchCondition: map[string]interface{}{
					"method": "GET",
					"path":   "/api/users",
				},
			},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			matched, err := engine.simpleMatch(tt.request, tt.rule)
			assert.NoError(t, err, "匹配过程不应该出错")
			assert.Equal(t, tt.expected, matched, tt.name)
		})
	}
}
