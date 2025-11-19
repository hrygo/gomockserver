package engine

import (
	"regexp"
	"testing"

	"github.com/gomockserver/mockserver/internal/adapter"
	"github.com/gomockserver/mockserver/internal/models"
	"github.com/stretchr/testify/assert"
)

func TestRegexCacheStats(t *testing.T) {
	mockRepo := new(MockRuleRepository)
	engine := NewMatchEngine(mockRepo)

	// 初始状态应该没有缓存统计
	stats := engine.GetCacheStats()
	assert.Equal(t, int64(0), stats.Hits)
	assert.Equal(t, int64(0), stats.Misses)
	assert.Equal(t, 0, stats.Size)

	// 编译一个正则表达式，应该增加Misses
	_, err := engine.compileRegex("test.*pattern")
	assert.NoError(t, err)

	stats = engine.GetCacheStats()
	assert.Equal(t, int64(0), stats.Hits)
	assert.Equal(t, int64(1), stats.Misses)
	assert.Equal(t, 1, stats.Size)

	// 再次编译同一个正则表达式，应该增加Hits
	_, err = engine.compileRegex("test.*pattern")
	assert.NoError(t, err)

	stats = engine.GetCacheStats()
	assert.Equal(t, int64(1), stats.Hits)
	assert.Equal(t, int64(1), stats.Misses)
	assert.Equal(t, 1, stats.Size)

	// 编译不同的正则表达式，应该增加Misses
	_, err = engine.compileRegex("another.*pattern")
	assert.NoError(t, err)

	stats = engine.GetCacheStats()
	assert.Equal(t, int64(1), stats.Hits)
	assert.Equal(t, int64(2), stats.Misses)
	assert.Equal(t, 2, stats.Size)
}

func TestRegexCacheLRU(t *testing.T) {
	// 创建一个小容量的缓存来测试LRU行为
	cache := NewLRURegexCache(2) // 只能缓存2个项

	// 添加3个不同的正则表达式
	re1, _ := regexp.Compile("pattern1")
	re2, _ := regexp.Compile("pattern2")
	re3, _ := regexp.Compile("pattern3")

	cache.Put("pattern1", re1)
	cache.Put("pattern2", re2)
	cache.Put("pattern3", re3)

	// 由于容量限制，第一个pattern应该被淘汰
	assert.Equal(t, 2, cache.Size())

	// pattern1应该不存在
	if _, exists := cache.Get("pattern1"); exists {
		t.Error("pattern1 should have been evicted")
	}

	// pattern2和pattern3应该存在
	if re, exists := cache.Get("pattern2"); assert.True(t, exists) {
		assert.Equal(t, re2, re)
	}

	if re, exists := cache.Get("pattern3"); assert.True(t, exists) {
		assert.Equal(t, re3, re)
	}
}

func TestRegexMatchWithCache(t *testing.T) {
	// 创建一个mock引擎来测试正则匹配中的缓存使用
	mockRepo := new(MockRuleRepository)
	engine := NewMatchEngine(mockRepo)

	request := &adapter.Request{
		Protocol: models.ProtocolHTTP,
		Path:     "/api/test",
		Metadata: map[string]interface{}{
			"method": "GET",
		},
	}

	rule := &models.Rule{
		Protocol:  models.ProtocolHTTP,
		MatchType: models.MatchTypeRegex,
		MatchCondition: map[string]interface{}{
			"path_regex": "/api/.*", // 这将被编译并缓存
		},
	}

	// 初始状态
	stats := engine.GetCacheStats()
	assert.Equal(t, int64(0), stats.Hits)
	assert.Equal(t, int64(0), stats.Misses)

	// 第一次匹配，应该增加Misses
	matched, err := engine.regexMatch(request, rule)
	assert.NoError(t, err)
	assert.True(t, matched)

	stats = engine.GetCacheStats()
	assert.Equal(t, int64(0), stats.Hits)
	assert.Equal(t, int64(1), stats.Misses)

	// 第二次匹配相同规则，应该增加Hits
	matched, err = engine.regexMatch(request, rule)
	assert.NoError(t, err)
	assert.True(t, matched)

	stats = engine.GetCacheStats()
	assert.Equal(t, int64(1), stats.Hits)
	assert.Equal(t, int64(1), stats.Misses)
}
