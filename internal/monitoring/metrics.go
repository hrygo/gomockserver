// Package monitoring provides monitoring and metrics collection for the mock server.
package monitoring

import (
	"sync/atomic"
)

// Metrics holds all the application metrics
type Metrics struct {
	// Regex cache metrics
	RegexCacheHits   int64 `json:"regex_cache_hits"`
	RegexCacheMisses int64 `json:"regex_cache_misses"`
	RegexCacheSize   int64 `json:"regex_cache_size"`
	
	// Request metrics
	TotalRequests    int64 `json:"total_requests"`
	MatchedRequests  int64 `json:"matched_requests"`
	UnmatchedRequests int64 `json:"unmatched_requests"`
	
	// Rule metrics
	TotalRules       int64 `json:"total_rules"`
	EnabledRules     int64 `json:"enabled_rules"`
}

// Global metrics instance
var globalMetrics = &Metrics{}

// GetMetrics returns the current metrics
func GetMetrics() *Metrics {
	return globalMetrics
}

// IncrementRegexCacheHits increments the regex cache hits counter
func IncrementRegexCacheHits() {
	atomic.AddInt64(&globalMetrics.RegexCacheHits, 1)
}

// IncrementRegexCacheMisses increments the regex cache misses counter
func IncrementRegexCacheMisses() {
	atomic.AddInt64(&globalMetrics.RegexCacheMisses, 1)
}

// SetRegexCacheSize sets the current regex cache size
func SetRegexCacheSize(size int64) {
	atomic.StoreInt64(&globalMetrics.RegexCacheSize, size)
}

// IncrementTotalRequests increments the total requests counter
func IncrementTotalRequests() {
	atomic.AddInt64(&globalMetrics.TotalRequests, 1)
}

// IncrementMatchedRequests increments the matched requests counter
func IncrementMatchedRequests() {
	atomic.AddInt64(&globalMetrics.MatchedRequests, 1)
}

// IncrementUnmatchedRequests increments the unmatched requests counter
func IncrementUnmatchedRequests() {
	atomic.AddInt64(&globalMetrics.UnmatchedRequests, 1)
}

// SetTotalRules sets the total rules count
func SetTotalRules(count int64) {
	atomic.StoreInt64(&globalMetrics.TotalRules, count)
}

// SetEnabledRules sets the enabled rules count
func SetEnabledRules(count int64) {
	atomic.StoreInt64(&globalMetrics.EnabledRules, count)
}

// ResetMetrics resets all metrics to zero
func ResetMetrics() {
	atomic.StoreInt64(&globalMetrics.RegexCacheHits, 0)
	atomic.StoreInt64(&globalMetrics.RegexCacheMisses, 0)
	atomic.StoreInt64(&globalMetrics.RegexCacheSize, 0)
	atomic.StoreInt64(&globalMetrics.TotalRequests, 0)
	atomic.StoreInt64(&globalMetrics.MatchedRequests, 0)
	atomic.StoreInt64(&globalMetrics.UnmatchedRequests, 0)
	atomic.StoreInt64(&globalMetrics.TotalRules, 0)
	atomic.StoreInt64(&globalMetrics.EnabledRules, 0)
}