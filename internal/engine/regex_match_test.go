package engine

import (
	"testing"

	"github.com/gomockserver/mockserver/internal/adapter"
	"github.com/gomockserver/mockserver/internal/models"
	"github.com/stretchr/testify/assert"
)

// TestRegexMatch_PathMatching 测试路径正则匹配
func TestRegexMatch_PathMatching(t *testing.T) {
	engine := NewMatchEngine(nil)

	tests := []struct {
		name     string
		path     string
		pattern  string
		expected bool
	}{
		{"简单路径匹配", "/api/users/123", "/api/users/[0-9]+", true},
		{"简单路径不匹配", "/api/users/abc", "/api/users/[0-9]+", false},
		{"复杂路径匹配", "/api/v1/users/123/posts/456", "/api/v[0-9]+/users/[0-9]+/posts/[0-9]+", true},
		{"复杂路径不匹配", "/api/v1/users/123/comments", "/api/v[0-9]+/users/[0-9]+/posts/[0-9]+", false},
		{"通配符匹配", "/api/any/path", "/api/.*", true},
		{"通配符不匹配", "/other/path", "/api/.*", false},
		{"精确匹配", "/api/exact", "/api/exact", true},
		{"精确不匹配", "/api/different", "/api/exact", false},
		{"空模式", "/api/test", "", true}, // 空模式应该匹配所有路径
		{"无效正则", "/api/test", "[invalid", false}, // 无效正则应该返回false而不是panic
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			request := &adapter.Request{
				Protocol: models.ProtocolHTTP,
				Path:     tt.path,
				Metadata: map[string]interface{}{
					"method": "GET",
				},
			}

			rule := &models.Rule{
				Protocol:  models.ProtocolHTTP,
				MatchType: models.MatchTypeRegex,
				MatchCondition: map[string]interface{}{
					"path": tt.pattern,
				},
			}

			matched, err := engine.regexMatch(request, rule)
			if tt.pattern == "[invalid" {
				// 无效正则应该返回false（而不是错误），因为regexMatch函数内部处理了错误
				assert.NoError(t, err)
				assert.False(t, matched)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expected, matched)
			}
		})
	}
}

// TestRegexMatch_QueryMatching 测试查询参数正则匹配
func TestRegexMatch_QueryMatching(t *testing.T) {
	engine := NewMatchEngine(nil)

	tests := []struct {
		name     string
		query    map[string]string
		patterns map[string]string
		expected bool
	}{
		{
			name: "单参数匹配",
			query: map[string]string{
				"page": "1",
			},
			patterns: map[string]string{
				"page": "[0-9]+",
			},
			expected: true,
		},
		{
			name: "单参数不匹配",
			query: map[string]string{
				"page": "abc",
			},
			patterns: map[string]string{
				"page": "[0-9]+",
			},
			expected: false,
		},
		{
			name: "多参数匹配",
			query: map[string]string{
				"page":     "1",
				"limit":    "10",
				"category": "electronics",
			},
			patterns: map[string]string{
				"page":     "[0-9]+",
				"limit":    "[0-9]+",
				"category": "[a-z]+",
			},
			expected: true,
		},
		{
			name: "多参数部分不匹配",
			query: map[string]string{
				"page":     "1",
				"limit":    "abc", // 不匹配数字模式
				"category": "electronics",
			},
			patterns: map[string]string{
				"page":     "[0-9]+",
				"limit":    "[0-9]+",
				"category": "[a-z]+",
			},
			expected: false,
		},
		{
			name: "缺少参数",
			query: map[string]string{
				"page": "1",
			},
			patterns: map[string]string{
				"page":  "[0-9]+",
				"limit": "[0-9]+", // 缺少limit参数
			},
			expected: false,
		},
		{
			name: "复杂模式匹配",
			query: map[string]string{
				"email": "user@example.com",
			},
			patterns: map[string]string{
				"email": `^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`,
			},
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			request := &adapter.Request{
				Protocol: models.ProtocolHTTP,
				Path:     "/api/test",
				Metadata: map[string]interface{}{
					"method": "GET",
					"query":  tt.query,
				},
			}

			rule := &models.Rule{
				Protocol:  models.ProtocolHTTP,
				MatchType: models.MatchTypeRegex,
				MatchCondition: map[string]interface{}{
					"query": tt.patterns,
				},
			}

			matched, err := engine.regexMatch(request, rule)
			assert.NoError(t, err)
			assert.Equal(t, tt.expected, matched)
		})
	}
}

// TestRegexMatch_HeaderMatching 测试请求头正则匹配
func TestRegexMatch_HeaderMatching(t *testing.T) {
	engine := NewMatchEngine(nil)

	tests := []struct {
		name     string
		headers  map[string]string
		patterns map[string]string
		expected bool
	}{
		{
			name: "单头匹配",
			headers: map[string]string{
				"Content-Type": "application/json",
			},
			patterns: map[string]string{
				"Content-Type": "application/.*",
			},
			expected: true,
		},
		{
			name: "头部大小写不敏感匹配",
			headers: map[string]string{
				"content-type": "application/json", // 小写
			},
			patterns: map[string]string{
				"Content-Type": "application/.*", // 大写模式
			},
			expected: true,
		},
		{
			name: "混合大小写匹配",
			headers: map[string]string{
				"X-Request-ID": "abc123xyz",
			},
			patterns: map[string]string{
				"x-request-id": "[a-z0-9]+", // 小写模式
			},
			expected: true,
		},
		{
			name: "多头部匹配",
			headers: map[string]string{
				"Authorization": "Bearer token123",
				"User-Agent":    "Mozilla/5.0",
				"Accept":        "application/json",
			},
			patterns: map[string]string{
				"Authorization": "Bearer .*",
				"User-Agent":    "Mozilla/.*",
				"Accept":        "application/.*",
			},
			expected: true,
		},
		{
			name: "缺少头部",
			headers: map[string]string{
				"Content-Type": "application/json",
			},
			patterns: map[string]string{
				"Content-Type": "application/.*",
				"Authorization": "Bearer .*", // 缺少Authorization头部
			},
			expected: false,
		},
		{
			name: "头部值不匹配",
			headers: map[string]string{
				"Content-Type": "text/plain",
			},
			patterns: map[string]string{
				"Content-Type": "application/.*",
			},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			request := &adapter.Request{
				Protocol: models.ProtocolHTTP,
				Path:     "/api/test",
				Headers:  tt.headers,
				Metadata: map[string]interface{}{
					"method": "GET",
				},
			}

			rule := &models.Rule{
				Protocol:  models.ProtocolHTTP,
				MatchType: models.MatchTypeRegex,
				MatchCondition: map[string]interface{}{
					"headers": tt.patterns,
				},
			}

			matched, err := engine.regexMatch(request, rule)
			assert.NoError(t, err)
			assert.Equal(t, tt.expected, matched)
		})
	}
}

// TestRegexMatch_CombinedMatching 测试组合匹配
func TestRegexMatch_CombinedMatching(t *testing.T) {
	engine := NewMatchEngine(nil)

	request := &adapter.Request{
		Protocol: models.ProtocolHTTP,
		Path:     "/api/v1/users/123",
		Headers: map[string]string{
			"Content-Type": "application/json",
			"Authorization": "Bearer token123",
		},
		Metadata: map[string]interface{}{
			"method": "GET",
			"query": map[string]string{
				"page":  "1",
				"limit": "10",
			},
		},
	}

	rule := &models.Rule{
		Protocol:  models.ProtocolHTTP,
		MatchType: models.MatchTypeRegex,
		MatchCondition: map[string]interface{}{
			"method":  "GET",
			"path":    "/api/v[0-9]+/users/[0-9]+",
			"query": map[string]string{
				"page":  "[0-9]+",
				"limit": "[0-9]+",
			},
			"headers": map[string]string{
				"Content-Type":  "application/.*",
				"Authorization": "Bearer .*",
			},
		},
	}

	matched, err := engine.regexMatch(request, rule)
	assert.NoError(t, err)
	assert.True(t, matched)
}

// TestRegexMatch_Performance 性能测试
func TestRegexMatch_Performance(t *testing.T) {
	if testing.Short() {
		t.Skip("跳过性能测试")
	}

	engine := NewMatchEngine(nil)

	// 创建单个规则，重复使用以测试缓存效果
	rule := &models.Rule{
		Protocol:  models.ProtocolHTTP,
		MatchType: models.MatchTypeRegex,
		MatchCondition: map[string]interface{}{
			"path":    "/api/test/[0-9]+",
			"method":  "GET",
			"headers": map[string]string{
				"X-Test": "value.*",
			},
		},
	}

	request := &adapter.Request{
		Protocol: models.ProtocolHTTP,
		Path:     "/api/test/123",
		Headers: map[string]string{
			"X-Test": "value123",
		},
		Metadata: map[string]interface{}{
			"method": "GET",
		},
	}

	// 性能测试：执行1000次匹配同一个规则
	for i := 0; i < 1000; i++ {
		_, err := engine.regexMatch(request, rule)
		assert.NoError(t, err)
	}

	// 检查缓存统计
	stats := engine.GetCacheStats()
	assert.Greater(t, stats.Hits, int64(0)) // 应该有缓存命中
	assert.Greater(t, stats.Misses, int64(0)) // 第一次应该有缓存未命中
}

// TestRegexMatch_CacheEviction 测试缓存淘汰
func TestRegexMatch_CacheEviction(t *testing.T) {
	// 创建容量为1的小缓存
	engine := &MatchEngine{
		regexCache: NewLRURegexCache(1), // 只能缓存1个正则表达式
	}

	// 第一个请求
	request1 := &adapter.Request{
		Protocol: models.ProtocolHTTP,
		Path:     "/api/test1",
		Metadata: map[string]interface{}{
			"method": "GET",
		},
	}

	// 第二个请求
	request2 := &adapter.Request{
		Protocol: models.ProtocolHTTP,
		Path:     "/api/test2",
		Metadata: map[string]interface{}{
			"method": "GET",
		},
	}

	// 第一个规则
	rule1 := &models.Rule{
		Protocol:  models.ProtocolHTTP,
		MatchType: models.MatchTypeRegex,
		MatchCondition: map[string]interface{}{
			"path": "/api/test1",
		},
	}

	// 第二个规则
	rule2 := &models.Rule{
		Protocol:  models.ProtocolHTTP,
		MatchType: models.MatchTypeRegex,
		MatchCondition: map[string]interface{}{
			"path": "/api/test2",
		},
	}

	// 匹配第一个规则（会缓存）
	matched, err := engine.regexMatch(request1, rule1)
	assert.NoError(t, err)
	assert.True(t, matched)

	stats := engine.GetCacheStats()
	assert.Equal(t, int64(0), stats.Hits)
	assert.Equal(t, int64(1), stats.Misses)

	// 匹配第二个规则（会淘汰第一个缓存）
	matched, err = engine.regexMatch(request2, rule2)
	assert.NoError(t, err)
	assert.True(t, matched) // 路径匹配

	stats = engine.GetCacheStats()
	assert.Equal(t, int64(0), stats.Hits)
	assert.Equal(t, int64(2), stats.Misses)

	// 再次匹配第一个规则（缓存已被淘汰）
	matched, err = engine.regexMatch(request1, rule1)
	assert.NoError(t, err)
	assert.True(t, matched)

	stats = engine.GetCacheStats()
	assert.Equal(t, int64(0), stats.Hits)
	assert.Equal(t, int64(3), stats.Misses)
}