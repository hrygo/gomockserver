package monitoring

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMetrics(t *testing.T) {
	// Reset metrics to start with a clean state
	ResetMetrics()

	// Test initial state
	metrics := GetMetrics()
	assert.Equal(t, int64(0), metrics.RegexCacheHits)
	assert.Equal(t, int64(0), metrics.RegexCacheMisses)
	assert.Equal(t, int64(0), metrics.RegexCacheSize)
	assert.Equal(t, int64(0), metrics.TotalRequests)
	assert.Equal(t, int64(0), metrics.MatchedRequests)
	assert.Equal(t, int64(0), metrics.UnmatchedRequests)
	assert.Equal(t, int64(0), metrics.TotalRules)
	assert.Equal(t, int64(0), metrics.EnabledRules)

	// Test incrementing counters
	IncrementRegexCacheHits()
	IncrementRegexCacheHits()
	metrics = GetMetrics()
	assert.Equal(t, int64(2), metrics.RegexCacheHits)

	IncrementRegexCacheMisses()
	metrics = GetMetrics()
	assert.Equal(t, int64(1), metrics.RegexCacheMisses)

	// Test setting values
	SetRegexCacheSize(5)
	metrics = GetMetrics()
	assert.Equal(t, int64(5), metrics.RegexCacheSize)

	IncrementTotalRequests()
	IncrementMatchedRequests()
	IncrementUnmatchedRequests()
	SetTotalRules(10)
	SetEnabledRules(8)

	metrics = GetMetrics()
	assert.Equal(t, int64(1), metrics.TotalRequests)
	assert.Equal(t, int64(1), metrics.MatchedRequests)
	assert.Equal(t, int64(1), metrics.UnmatchedRequests)
	assert.Equal(t, int64(10), metrics.TotalRules)
	assert.Equal(t, int64(8), metrics.EnabledRules)

	// Test reset
	ResetMetrics()
	metrics = GetMetrics()
	assert.Equal(t, int64(0), metrics.RegexCacheHits)
	assert.Equal(t, int64(0), metrics.RegexCacheMisses)
	assert.Equal(t, int64(0), metrics.RegexCacheSize)
	assert.Equal(t, int64(0), metrics.TotalRequests)
	assert.Equal(t, int64(0), metrics.MatchedRequests)
	assert.Equal(t, int64(0), metrics.UnmatchedRequests)
	assert.Equal(t, int64(0), metrics.TotalRules)
	assert.Equal(t, int64(0), metrics.EnabledRules)
}
