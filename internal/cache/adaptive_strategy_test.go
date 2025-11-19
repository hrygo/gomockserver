package cache

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"go.uber.org/zap/zaptest"
)

func TestAdaptiveStrategy_Creation(t *testing.T) {
	logger := zaptest.NewLogger(t)
	strategy := DefaultCacheStrategy()
	config := DefaultTuningConfig()

	adaptive := NewAdaptiveStrategy(strategy, config, logger)

	assert.NotNil(t, adaptive)
	assert.Equal(t, strategy, adaptive.baseStrategy)
	assert.Equal(t, strategy, adaptive.currentStrategy)
	assert.NotNil(t, adaptive.stats)
	assert.NotNil(t, adaptive.adjustmentHistory)
}

func TestAdaptiveStrategy_GetStrategy(t *testing.T) {
	logger := zaptest.NewLogger(t)
	strategy := DefaultCacheStrategy()
	config := DefaultTuningConfig()
	adaptive := NewAdaptiveStrategy(strategy, config, logger)

	// 获取当前策略
	current := adaptive.GetStrategy()
	assert.NotNil(t, current)
	assert.Equal(t, strategy.HotDataThreshold, current.HotDataThreshold)
	assert.Equal(t, strategy.WarmDataThreshold, current.WarmDataThreshold)
}

func TestTuningConfig_Creation(t *testing.T) {
	config := DefaultTuningConfig()

	assert.NotNil(t, config)
	assert.Equal(t, 0.6, config.MinHotThreshold)
	assert.Equal(t, 0.95, config.MaxHotThreshold)
	assert.Equal(t, 0.1, config.MinWarmThreshold)
	assert.Equal(t, 0.4, config.MaxWarmThreshold)
	assert.Equal(t, 5*time.Minute, config.AdjustmentInterval)
	assert.Equal(t, 0.05, config.MinAdjustmentDelta)
	assert.Equal(t, 0.2, config.MaxAdjustmentDelta)
	assert.Equal(t, 100, config.HistoryLimit)
}

func TestAdaptiveStrategy_ConfigValidation(t *testing.T) {
	logger := zaptest.NewLogger(t)
	strategy := DefaultCacheStrategy()

	// 测试有效配置
	config := DefaultTuningConfig()
	assert.NotPanics(t, func() {
		NewAdaptiveStrategy(strategy, config, logger)
	})

	// 测试nil配置
	assert.NotPanics(t, func() {
		NewAdaptiveStrategy(strategy, nil, logger)
	})
}

func TestStrategyStats_Creation(t *testing.T) {
	stats := &StrategyStats{
		TotalRequests:      1000,
		HitRate:            0.85,
		AvgResponseTime:     25 * time.Millisecond,
		MemoryUsage:         100 * 1024 * 1024,
		CPULoad:            0.4,
		OptimalHitRate:     0.9,
		OptimalResponseTime: 20 * time.Millisecond,
	}

	assert.Equal(t, int64(1000), stats.TotalRequests)
	assert.Equal(t, 0.85, stats.HitRate)
	assert.Equal(t, 25*time.Millisecond, stats.AvgResponseTime)
	assert.Equal(t, int64(100*1024*1024), stats.MemoryUsage)
	assert.Equal(t, 0.4, stats.CPULoad)
	assert.Equal(t, 0.9, stats.OptimalHitRate)
	assert.Equal(t, 20*time.Millisecond, stats.OptimalResponseTime)
}

func TestStrategyAdjustmentRecord_Creation(t *testing.T) {
	record := &StrategyAdjustmentRecord{
		Timestamp:   time.Now(),
		Type:        "threshold_adjustment",
		Reason:      "Low hit rate detected",
		OldStrategy: "hot_threshold: 0.8",
		NewStrategy: "hot_threshold: 0.7",
		Improvement: 0.1,
	}

	assert.NotNil(t, record.Timestamp)
	assert.Equal(t, "threshold_adjustment", record.Type)
	assert.Equal(t, "Low hit rate detected", record.Reason)
	assert.Equal(t, "hot_threshold: 0.8", record.OldStrategy)
	assert.Equal(t, "hot_threshold: 0.7", record.NewStrategy)
	assert.Equal(t, 0.1, record.Improvement)
}

// 基准测试
func BenchmarkAdaptiveStrategy_GetStrategy(b *testing.B) {
	logger := zaptest.NewLogger(b)
	strategy := DefaultCacheStrategy()
	config := DefaultTuningConfig()
	adaptive := NewAdaptiveStrategy(strategy, config, logger)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		adaptive.GetStrategy()
	}
}

func BenchmarkTuningConfig_Creation(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		DefaultTuningConfig()
	}
}