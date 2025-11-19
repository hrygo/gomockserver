package cache

import (
	"fmt"
	"math"
	"sync"
	"time"

	"go.uber.org/zap"
)

// AdaptiveStrategy 自适应缓存策略
type AdaptiveStrategy struct {
	mu                    sync.RWMutex
	baseStrategy          *CacheStrategy
	currentStrategy       *CacheStrategy
	stats                  *StrategyStats
	adjustmentInterval     time.Duration
	adjustmentHistory      []StrategyAdjustmentRecord
	logger                 *zap.Logger
}

// StrategyStats 策略统计
type StrategyStats struct {
	TotalRequests      int64     `json:"total_requests"`
	HitRate            float64   `json:"hit_rate"`
	AvgResponseTime     time.Duration `json:"avg_response_time"`
	MemoryUsage         int64     `json:"memory_usage"`
	CPULoad            float64   `json:"cpu_load"`
	OptimalHitRate     float64   `json:"optimal_hit_rate"`
	OptimalResponseTime time.Duration `json:"optimal_response_time"`
}

// StrategyAdjustmentRecord 策略调整记录
type StrategyAdjustmentRecord struct {
	Timestamp    time.Time `json:"timestamp"`
	Type         string    `json:"type"`
	Reason       string    `json:"reason"`
	OldStrategy  string    `json:"old_strategy"`
	NewStrategy  string    `json:"new_strategy"`
	Improvement  float64   `json:"improvement"`
}

// StrategyTuningConfig 策略调优配置
type StrategyTuningConfig struct {
	MinHotThreshold    float64       `json:"min_hot_threshold"`     // 最小热点阈值
	MaxHotThreshold    float64       `json:"max_hot_threshold"`     // 最大热点阈值
	MinWarmThreshold   float64       `json:"min_warm_threshold"`    // 最小温数据阈值
	MaxWarmThreshold   float64       `json:"max_warm_threshold"`    // 最大温数据阈值
	AdjustmentInterval  time.Duration `json:"adjustment_interval"`   // 调整间隔
	MinAdjustmentDelta  float64       `json:"min_adjustment_delta"`  // 最小调整幅度
	MaxAdjustmentDelta  float64       `json:"max_adjustment_delta"`  // 最大调整幅度
	HistoryLimit        int           `json:"history_limit"`        // 历史记录限制
}

// DefaultTuningConfig 默认调优配置
func DefaultTuningConfig() *StrategyTuningConfig {
	return &StrategyTuningConfig{
		MinHotThreshold:   0.6,
		MaxHotThreshold:   0.95,
		MinWarmThreshold:  0.1,
		MaxWarmThreshold:  0.4,
		AdjustmentInterval: 5 * time.Minute,
		MinAdjustmentDelta: 0.05,
		MaxAdjustmentDelta: 0.2,
		HistoryLimit:       100,
	}
}

// NewAdaptiveStrategy 创建自适应策略管理器
func NewAdaptiveStrategy(baseStrategy *CacheStrategy, tuningConfig *StrategyTuningConfig, logger *zap.Logger) *AdaptiveStrategy {
	if tuningConfig == nil {
		tuningConfig = DefaultTuningConfig()
	}

	strategy := &AdaptiveStrategy{
		baseStrategy:       baseStrategy,
		currentStrategy:    baseStrategy,
		stats:              &StrategyStats{},
		adjustmentInterval: tuningConfig.AdjustmentInterval,
		adjustmentHistory:  make([]StrategyAdjustmentRecord, 0),
		logger:             logger.Named("adaptive_strategy"),
	}

	// 启动自适应调整
	go strategy.startAdaptiveAdjustment(tuningConfig)

	logger.Info("Adaptive cache strategy initialized",
		zap.Float64("hot_threshold", baseStrategy.HotDataThreshold),
		zap.Float64("warm_threshold", baseStrategy.WarmDataThreshold),
	)

	return strategy
}

// GetStrategy 获取当前策略
func (as *AdaptiveStrategy) GetStrategy() *CacheStrategy {
	as.mu.RLock()
	defer as.mu.RUnlock()
	return as.currentStrategy
}

// UpdateStats 更新策略统计
func (as *AdaptiveStrategy) UpdateStats(stats *CacheStats, memoryUsage int64) {
	as.mu.Lock()
	defer as.mu.Unlock()

	as.stats.TotalRequests = stats.TotalRequests
	as.stats.HitRate = stats.TotalHitRate
	as.stats.AvgResponseTime = stats.AvgResponseTime
	as.stats.MemoryUsage = memoryUsage
}

// GetCurrentStats 获取当前统计
func (as *AdaptiveStrategy) GetCurrentStats() *StrategyStats {
	as.mu.RLock()
	defer as.mu.RUnlock()

	return &StrategyStats{
		TotalRequests:      as.stats.TotalRequests,
		HitRate:            as.stats.HitRate,
		AvgResponseTime:     as.stats.AvgResponseTime,
		MemoryUsage:         as.stats.MemoryUsage,
		CPULoad:            as.stats.CPULoad,
		OptimalHitRate:     as.stats.OptimalHitRate,
		OptimalResponseTime: as.stats.OptimalResponseTime,
	}
}

// GetAdjustmentHistory 获取调整历史
func (as *AdaptiveStrategy) GetAdjustmentHistory(limit int) []StrategyAdjustmentRecord {
	as.mu.RLock()
	defer as.mu.RUnlock()

	if limit <= 0 || limit > len(as.adjustmentHistory) {
		limit = len(as.adjustmentHistory)
	}

	result := make([]StrategyAdjustmentRecord, limit)
	copy(result, as.adjustmentHistory[len(as.adjustmentHistory)-limit:])
	return result
}

// startAdaptiveAdjustment 启动自适应调整
func (as *AdaptiveStrategy) startAdaptiveAdjustment(config *StrategyTuningConfig) {
	ticker := time.NewTicker(config.AdjustmentInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			as.performAdaptiveAdjustment(config)
		}
	}
}

// performAdaptiveAdjustment 执行自适应调整
func (as *AdaptiveStrategy) performAdaptiveAdjustment(config *StrategyTuningConfig) {
	as.mu.Lock()
	defer as.mu.Unlock()

	// 检查是否有足够的数据进行调整
	if as.stats.TotalRequests < 100 {
		return
	}

	// 计算性能指标
	hitRateScore := as.calculateHitRateScore()
	responseTimeScore := as.calculateResponseTimeScore()
	memoryScore := as.calculateMemoryScore()

	// 综合评分 (权重: 命中率40%, 响应时间40%, 内存使用20%)
	overallScore := hitRateScore*0.4 + responseTimeScore*0.4 + memoryScore*0.2

	// 如果整体性能良好，不需要调整
	if overallScore > 0.8 {
		return
	}

	// 确定调整方向和幅度
	adjustment := as.calculateAdjustment(hitRateScore, responseTimeScore, memoryScore, config)
	if adjustment == nil {
		return
	}

	// 应用调整
	oldThreshold := as.currentStrategy.HotDataThreshold
	newStrategy := as.applyAdjustment(as.currentStrategy, adjustment, config)
	as.currentStrategy = newStrategy

	// 记录调整历史
	adjustmentRecord := StrategyAdjustmentRecord{
		Timestamp:   time.Now(),
		Type:        adjustment.Type,
		Reason:      adjustment.Reason,
		OldStrategy: as.formatStrategy(oldThreshold),
		NewStrategy: as.formatStrategy(newStrategy.HotDataThreshold),
		Improvement:  overallScore,
	}

	as.adjustmentHistory = append(as.adjustmentHistory, adjustmentRecord)

	// 限制历史记录数量
	if len(as.adjustmentHistory) > config.HistoryLimit {
		as.adjustmentHistory = as.adjustmentHistory[1:]
	}

	as.logger.Info("Cache strategy adjusted",
		zap.String("type", adjustment.Type),
		zap.String("reason", adjustment.Reason),
		zap.Float64("old_hot_threshold", oldThreshold),
		zap.Float64("new_hot_threshold", newStrategy.HotDataThreshold),
		zap.Float64("performance_score", overallScore),
	)
}

// StrategyAdjustment 策略调整
type StrategyAdjustment struct {
	Type        string  `json:"type"`        // 调整类型: "threshold", "ttl", "capacity"
	Reason      string  `json:"reason"`      // 调整原因
	Direction   string  `json:"direction"`   // 调整方向: "increase", "decrease"
	Magnitude   float64 `json:"magnitude"`   // 调整幅度
	Target      string  `json:"target"`      // 调整目标
}

// calculateHitRateScore 计算命中率评分
func (as *AdaptiveStrategy) calculateHitRateScore() float64 {
	// 理想命中率为90%
	optimal := 0.9
	if as.stats.HitRate >= optimal {
		return 1.0
	}
	return as.stats.HitRate / optimal
}

// calculateResponseTimeScore 计算响应时间评分
func (as *AdaptiveStrategy) calculateResponseTimeScore() float64 {
	// 理想响应时间为10ms
	optimal := 10 * time.Millisecond
	if as.stats.AvgResponseTime <= optimal {
		return 1.0
	}

	// 响应时间越长，评分越低
	ratio := float64(optimal) / float64(as.stats.AvgResponseTime)
	return math.Min(1.0, ratio)
}

// calculateMemoryScore 计算内存使用评分
func (as *AdaptiveStrategy) calculateMemoryScore() float64 {
	// 假设内存限制为100MB
	threshold := int64(100 * 1024 * 1024)
	if as.stats.MemoryUsage <= threshold {
		return 1.0
	}

	// 内存使用越少，评分越高
	ratio := float64(threshold) / float64(as.stats.MemoryUsage)
	return math.Min(1.0, ratio)
}

// calculateAdjustment 计算调整策略
func (as *AdaptiveStrategy) calculateAdjustment(hitRateScore, responseTimeScore, memoryScore float64, config *StrategyTuningConfig) *StrategyAdjustment {
	adjustments := make([]*StrategyAdjustment, 0)

	// 命中率分析
	if hitRateScore < 0.7 {
		// 命中率低，可能需要调整阈值
		if as.currentStrategy.HotDataThreshold > config.MinHotThreshold {
			adjustments = append(adjustments, &StrategyAdjustment{
				Type:      "threshold",
				Reason:    "hit_rate_low",
				Direction: "decrease",
				Magnitude: config.MinAdjustmentDelta,
				Target:    "hot_threshold",
			})
		}
		if as.currentStrategy.WarmDataThreshold > config.MinWarmThreshold {
			adjustments = append(adjustments, &StrategyAdjustment{
				Type:      "threshold",
				Reason:    "hit_rate_low",
				Direction: "decrease",
				Magnitude: config.MinAdjustmentDelta,
				Target:    "warm_threshold",
			})
		}
	}

	// 响应时间分析
	if responseTimeScore < 0.7 {
		// 响应时间慢，可能需要提高缓存命中率
		if as.currentStrategy.HotDataThreshold < config.MaxHotThreshold {
			adjustments = append(adjustments, &StrategyAdjustment{
				Type:      "threshold",
				Reason:    "response_time_slow",
				Direction: "increase",
				Magnitude: config.MinAdjustmentDelta,
				Target:    "hot_threshold",
			})
		}
	}

	// 内存使用分析
	if memoryScore < 0.6 {
		// 内存使用过高，需要提高阈值减少缓存
		if as.currentStrategy.HotDataThreshold < config.MaxHotThreshold {
			adjustments = append(adjustments, &StrategyAdjustment{
				Type:      "threshold",
				Reason:    "memory_high",
				Direction: "increase",
				Magnitude: config.MaxAdjustmentDelta,
				Target:    "hot_threshold",
			})
		}
	}

	// 选择最重要的调整
	if len(adjustments) > 0 {
		return adjustments[0] // 简单实现，选择第一个调整
	}

	return nil
}

// applyAdjustment 应用调整
func (as *AdaptiveStrategy) applyAdjustment(strategy *CacheStrategy, adjustment *StrategyAdjustment, config *StrategyTuningConfig) *CacheStrategy {
	newStrategy := *strategy // 复制策略

	switch adjustment.Target {
	case "hot_threshold":
		delta := adjustment.Magnitude
		if adjustment.Direction == "decrease" {
			delta = -delta
		}
		newStrategy.HotDataThreshold = math.Max(config.MinHotThreshold,
			math.Min(config.MaxHotThreshold, newStrategy.HotDataThreshold+delta))

	case "warm_threshold":
		delta := adjustment.Magnitude
		if adjustment.Direction == "decrease" {
			delta = -delta
		}
		newStrategy.WarmDataThreshold = math.Max(config.MinWarmThreshold,
			math.Min(config.MaxWarmThreshold, newStrategy.WarmDataThreshold+delta))
	}

	return &newStrategy
}

// formatStrategy 格式化策略信息
func (as *AdaptiveStrategy) formatStrategy(threshold float64) string {
	return fmt.Sprintf("hot_threshold=%.2f", threshold)
}

// SetOptimalTargets 设置优化目标
func (as *AdaptiveStrategy) SetOptimalTargets(hitRate float64, responseTime time.Duration) {
	as.mu.Lock()
	defer as.mu.Unlock()

	as.stats.OptimalHitRate = hitRate
	as.stats.OptimalResponseTime = responseTime
}

// SetCPULoad 设置CPU负载
func (as *AdaptiveStrategy) SetCPULoad(cpuLoad float64) {
	as.mu.Lock()
	defer as.mu.Unlock()

	as.stats.CPULoad = cpuLoad
}

// ResetHistory 重置调整历史
func (as *AdaptiveStrategy) ResetHistory() {
	as.mu.Lock()
	defer as.mu.Unlock()

	as.adjustmentHistory = make([]StrategyAdjustmentRecord, 0)
	as.logger.Info("Strategy adjustment history reset")
}

// GetStrategyConfig 获取策略配置详情
func (as *AdaptiveStrategy) GetStrategyConfig() map[string]interface{} {
	as.mu.RLock()
	defer as.mu.RUnlock()

	return map[string]interface{}{
		"current_hot_threshold":    as.currentStrategy.HotDataThreshold,
		"current_warm_threshold":   as.currentStrategy.WarmDataThreshold,
		"base_hot_threshold":       as.baseStrategy.HotDataThreshold,
		"base_warm_threshold":      as.baseStrategy.WarmDataThreshold,
		"total_adjustments":        len(as.adjustmentHistory),
		"last_adjustment_time":     func() string {
			if len(as.adjustmentHistory) > 0 {
				return as.adjustmentHistory[len(as.adjustmentHistory)-1].Timestamp.Format(time.RFC3339)
			}
			return "none"
		}(),
		"stats":                    as.stats,
	}
}