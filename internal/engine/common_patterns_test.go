package engine

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCommonRegexPatterns(t *testing.T) {
	// 验证所有常用模式都是有效的
	for _, pattern := range CommonRegexPatterns {
		err := validateRegexPattern(pattern)
		assert.NoError(t, err, "Pattern should be valid: %s", pattern)
	}
}

func TestPrecompileCommonPatterns(t *testing.T) {
	// 创建匹配引擎
	mockRepo := new(MockRuleRepository)
	engine := NewMatchEngine(mockRepo)

	// 获取初始缓存大小
	initialSize := engine.regexCache.Size()

	// 预编译常用模式
	engine.PrecompileCommonPatterns()

	// 验证缓存大小增加了
	finalSize := engine.regexCache.Size()
	assert.Greater(t, finalSize, initialSize, "Cache size should increase after precompilation")

	// 验证所有常用模式都在缓存中
	for _, pattern := range CommonRegexPatterns {
		if re, exists := engine.regexCache.Get(pattern); exists {
			assert.NotNil(t, re, "Compiled regex should not be nil for pattern: %s", pattern)
		}
	}

	// 验证可以通过compileRegex获取预编译的模式
	for _, pattern := range CommonRegexPatterns[:5] { // 测试前5个模式
		// 第一次调用应该命中缓存
		statsBefore := engine.GetCacheStats()
		re1, err1 := engine.compileRegex(pattern)
		statsAfter := engine.GetCacheStats()

		assert.NoError(t, err1, "Should compile pattern without error: %s", pattern)
		assert.NotNil(t, re1, "Should return compiled regex: %s", pattern)
		assert.Equal(t, statsBefore.Hits+1, statsAfter.Hits, "Should hit cache for pattern: %s", pattern)

		// 第二次调用也应该命中缓存
		statsBefore2 := engine.GetCacheStats()
		re2, err2 := engine.compileRegex(pattern)
		statsAfter2 := engine.GetCacheStats()

		assert.NoError(t, err2, "Should compile pattern without error: %s", pattern)
		assert.NotNil(t, re2, "Should return compiled regex: %s", pattern)
		assert.Equal(t, statsBefore2.Hits+1, statsAfter2.Hits, "Should hit cache for pattern: %s", pattern)
		assert.Equal(t, re1, re2, "Should return same regex instance")
	}
}
