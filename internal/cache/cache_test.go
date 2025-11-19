package cache

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestCacheStrategy_DefaultStrategy(t *testing.T) {
	strategy := DefaultCacheStrategy()

	assert.NotNil(t, strategy)
	assert.Equal(t, 10000, strategy.L1MaxEntries)
	assert.Equal(t, 1*time.Minute, strategy.L1TTL)
	assert.Equal(t, 10*time.Minute, strategy.L2TTL)
	assert.Equal(t, 0.8, strategy.HotDataThreshold)
	assert.Equal(t, 0.2, strategy.WarmDataThreshold)
	assert.True(t, strategy.PreloadEnabled)
}

func TestCacheEntry_Creation(t *testing.T) {
	entry := CacheEntry{
		Key:       "test_key",
		Value:     "test_value",
		Level:     L1_HOT,
		TTL:       10 * time.Minute,
		CreatedAt: time.Now(),
		AccessAt:  time.Now(),
		HitCount:  5,
		ExpireAt:  time.Now().Add(10 * time.Minute),
	}

	assert.Equal(t, "test_key", entry.Key)
	assert.Equal(t, "test_value", entry.Value)
	assert.Equal(t, L1_HOT, entry.Level)
	assert.Equal(t, 10*time.Minute, entry.TTL)
	assert.Equal(t, int64(5), entry.HitCount)
}

func TestCacheStats_Creation(t *testing.T) {
	stats := &CacheStats{
		TotalRequests:   1000,
		L1HitCount:      800,
		L2HitCount:      150,
		L3HitCount:      50,
		MissCount:       200,
		L1HitRate:       0.8,
		L2HitRate:       0.15,
		TotalHitRate:    0.8,
		AvgResponseTime: 25 * time.Millisecond,
		TotalEntries:    800,
		L1Entries:       600,
		L2Entries:       200,
	}

	assert.Equal(t, int64(1000), stats.TotalRequests)
	assert.Equal(t, int64(800), stats.L1HitCount)
	assert.Equal(t, int64(150), stats.L2HitCount)
	assert.Equal(t, int64(50), stats.L3HitCount)
	assert.Equal(t, int64(200), stats.MissCount)
	assert.Equal(t, 0.8, stats.L1HitRate)
	assert.Equal(t, 0.15, stats.L2HitRate)
	assert.Equal(t, 0.8, stats.TotalHitRate)
	assert.Equal(t, 25*time.Millisecond, stats.AvgResponseTime)
	assert.Equal(t, int64(800), stats.TotalEntries)
	assert.Equal(t, int64(600), stats.L1Entries)
	assert.Equal(t, int64(200), stats.L2Entries)
}

func TestCacheLevels(t *testing.T) {
	// 测试缓存级别常量
	assert.Equal(t, CacheLevel(0), L1_HOT)
	assert.Equal(t, CacheLevel(1), L2_WARM)
	assert.Equal(t, CacheLevel(2), L3_COLD)
}

// 基准测试
func BenchmarkCacheStrategy_DefaultStrategy(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		DefaultCacheStrategy()
	}
}

func BenchmarkCacheEntry_Creation(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = CacheEntry{
			Key:       "test_key",
			Value:     "test_value",
			Level:     L1_HOT,
			TTL:       10 * time.Minute,
			CreatedAt: time.Now(),
			AccessAt:  time.Now(),
			HitCount:  5,
			ExpireAt:  time.Now().Add(10 * time.Minute),
		}
	}
}
