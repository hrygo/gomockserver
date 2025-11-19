package cache

import (
	"context"
	"fmt"
	"math"
	"sync"
	"time"

	"go.uber.org/zap"
)

// AutoTuner 自动调优器
type AutoTuner struct {
	mu                    sync.RWMutex
	config                *AutoTuningConfig
	currentStrategy       *CacheStrategy
	performanceHistory    []PerformanceSnapshot
	tuningHistory         []TuningAction
	isActive              bool
	stopCh                chan struct{}
	logger                *zap.Logger
	metricsCollector      MetricsCollector
}

// AutoTuningConfig 自动调优配置
type AutoTuningConfig struct {
	// 调优控制
	EnableAutoTuning       bool          `json:"enable_auto_tuning"`
	TuningInterval        time.Duration `json:"tuning_interval"`
	MinDataPoints         int           `json:"min_data_points"`
	ReactionSpeed         float64       `json:"reaction_speed"`         // 反应速度 0.0-1.0
	MaxAdjustmentPerStep  float64       `json:"max_adjustment_per_step"` // 每步最大调整幅度

	// 性能阈值
	TargetHitRate         float64       `json:"target_hit_rate"`
	TargetResponseTime    time.Duration `json:"target_response_time"`
	AcceptableHitRate     float64       `json:"acceptable_hit_rate"`
	MaxResponseTime       time.Duration `json:"max_response_time"`

	// 调优参数范围
	HotThresholdRange     [2]float64    `json:"hot_threshold_range"`
	WarmThresholdRange    [2]float64    `json:"warm_threshold_range"`
	L1CapacityRange       [2]int        `json:"l1_capacity_range"`
	TTLRange              [2]time.Duration `json:"ttl_range"`

	// 调优权重
	HitRateWeight         float64       `json:"hit_rate_weight"`
	ResponseTimeWeight    float64       `json:"response_time_weight"`
	MemoryWeight          float64       `json:"memory_weight"`

	// 安全约束
	MinHitRate            float64       `json:"min_hit_rate"`
	MaxMemoryUsage        int64         `json:"max_memory_usage"`
	MinL1Capacity         int           `json:"min_l1_capacity"`
	MaxL1Capacity         int           `json:"max_l1_capacity"`

	// 高级设置
	EnablePredictiveTuning bool         `json:"enable_predictive_tuning"`
	PredictionWindow       time.Duration `json:"prediction_window"`
	LoadBasedTuning        bool         `json:"load_based_tuning"`
}

// PerformanceSnapshot 性能快照
type PerformanceSnapshot struct {
	Timestamp       time.Time         `json:"timestamp"`
	Metrics         *PerformanceMetrics `json:"metrics"`
	LoadFactor      float64           `json:"load_factor"`
	AccessPattern   string            `json:"access_pattern"`
}

// TuningAction 调优动作
type TuningAction struct {
	Timestamp       time.Time         `json:"timestamp"`
	Action          string            `json:"action"`
	Parameter       string            `json:"parameter"`
	OldValue        interface{}       `json:"old_value"`
	NewValue        interface{}       `json:"new_value"`
	Reason          string            `json:"reason"`
	ExpectedImpact  string            `json:"expected_impact"`
	ActualImpact    float64           `json:"actual_impact"`
	Success         bool              `json:"success"`
}

// MetricsCollector 性能指标收集器接口
type MetricsCollector interface {
	GetCurrentMetrics(ctx context.Context) (*PerformanceMetrics, error)
	GetLoadFactor(ctx context.Context) (float64, error)
	AnalyzeAccessPattern(ctx context.Context) (string, error)
}

// DefaultAutoTuningConfig 默认自动调优配置
func DefaultAutoTuningConfig() *AutoTuningConfig {
	return &AutoTuningConfig{
		EnableAutoTuning:      true,
		TuningInterval:        2 * time.Minute,
		MinDataPoints:         5,
		ReactionSpeed:         0.3,
		MaxAdjustmentPerStep:  0.1,

		TargetHitRate:         0.85,
		TargetResponseTime:    20 * time.Millisecond,
		AcceptableHitRate:     0.75,
		MaxResponseTime:       50 * time.Millisecond,

		HotThresholdRange:     [2]float64{0.6, 0.9},
		WarmThresholdRange:    [2]float64{0.1, 0.4},
		L1CapacityRange:       [2]int{2000, 8000},
		TTLRange:              [2]time.Duration{10 * time.Minute, 1 * time.Hour},

		HitRateWeight:         0.5,
		ResponseTimeWeight:    0.4,
		MemoryWeight:          0.1,

		MinHitRate:            0.6,
		MaxMemoryUsage:        150 * 1024 * 1024, // 150MB
		MinL1Capacity:         1000,
		MaxL1Capacity:         10000,

		EnablePredictiveTuning: true,
		PredictionWindow:       10 * time.Minute,
		LoadBasedTuning:        true,
	}
}

// NewAutoTuner 创建自动调优器
func NewAutoTuner(initialStrategy *CacheStrategy, config *AutoTuningConfig, logger *zap.Logger) *AutoTuner {
	if config == nil {
		config = DefaultAutoTuningConfig()
	}

	tuner := &AutoTuner{
		config:             config,
		currentStrategy:    initialStrategy,
		performanceHistory: make([]PerformanceSnapshot, 0),
		tuningHistory:      make([]TuningAction, 0),
		stopCh:             make(chan struct{}),
		logger:             logger.Named("auto_tuner"),
	}

	return tuner
}

// SetMetricsCollector 设置性能指标收集器
func (at *AutoTuner) SetMetricsCollector(collector MetricsCollector) {
	at.mu.Lock()
	defer at.mu.Unlock()
	at.metricsCollector = collector
}

// StartAutoTuning 启动自动调优
func (at *AutoTuner) StartAutoTuning(ctx context.Context) error {
	if !at.config.EnableAutoTuning {
		at.logger.Info("Auto tuning is disabled")
		return nil
	}

	at.mu.Lock()
	if at.isActive {
		at.mu.Unlock()
		return fmt.Errorf("auto tuning is already active")
	}
	at.isActive = true
	at.mu.Unlock()

	at.logger.Info("Starting auto tuning",
		zap.Duration("interval", at.config.TuningInterval),
		zap.Float64("reaction_speed", at.config.ReactionSpeed),
	)

	ticker := time.NewTicker(at.config.TuningInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			at.StopAutoTuning()
			return ctx.Err()
		case <-at.stopCh:
			return nil
		case <-ticker.C:
			if err := at.performTuningCycle(ctx); err != nil {
				at.logger.Error("Tuning cycle failed", zap.Error(err))
			}
		}
	}
}

// StopAutoTuning 停止自动调优
func (at *AutoTuner) StopAutoTuning() {
	at.mu.Lock()
	defer at.mu.Unlock()

	if at.isActive {
		at.isActive = false
		close(at.stopCh)
		at.stopCh = make(chan struct{})
		at.logger.Info("Auto tuning stopped")
	}
}

// performTuningCycle 执行调优周期
func (at *AutoTuner) performTuningCycle(ctx context.Context) error {
	// 收集当前性能数据
	snapshot, err := at.collectPerformanceSnapshot(ctx)
	if err != nil {
		return fmt.Errorf("failed to collect performance snapshot: %w", err)
	}

	// 添加到历史记录
	at.addPerformanceSnapshot(snapshot)

	// 检查是否有足够的数据进行调优
	if len(at.performanceHistory) < at.config.MinDataPoints {
		at.logger.Debug("Insufficient data for tuning",
			zap.Int("current_points", len(at.performanceHistory)),
			zap.Int("required_points", at.config.MinDataPoints))
		return nil
	}

	// 分析性能趋势并决定调优动作
	actions := at.analyzeAndPlanTuning()

	// 执行调优动作
	for _, action := range actions {
		if err := at.executeTuningAction(ctx, action); err != nil {
			at.logger.Error("Failed to execute tuning action",
				zap.String("action", action.Action),
				zap.Error(err))
			continue
		}
	}

	return nil
}

// collectPerformanceSnapshot 收集性能快照
func (at *AutoTuner) collectPerformanceSnapshot(ctx context.Context) (*PerformanceSnapshot, error) {
	var metrics *PerformanceMetrics
	var loadFactor float64
	var accessPattern string = "unknown"

	// 从性能收集器获取指标
	if at.metricsCollector != nil {
		var err error
		metrics, err = at.metricsCollector.GetCurrentMetrics(ctx)
		if err != nil {
			at.logger.Warn("Failed to get metrics from collector", zap.Error(err))
		}

		loadFactor, err = at.metricsCollector.GetLoadFactor(ctx)
		if err != nil {
			at.logger.Debug("Failed to get load factor", zap.Error(err))
			loadFactor = 0.5 // 默认中等负载
		}

		accessPattern, err = at.metricsCollector.AnalyzeAccessPattern(ctx)
		if err != nil {
			at.logger.Debug("Failed to analyze access pattern", zap.Error(err))
		}
	}

	// 如果没有收集器，使用模拟数据
	if metrics == nil {
		metrics = at.simulateMetrics()
	}

	return &PerformanceSnapshot{
		Timestamp:     time.Now(),
		Metrics:       metrics,
		LoadFactor:    loadFactor,
		AccessPattern: accessPattern,
	}, nil
}

// simulateMetrics 模拟性能指标（用于测试）
func (at *AutoTuner) simulateMetrics() *PerformanceMetrics {
	// 基于当前策略模拟性能
	baseHitRate := 0.7
	hitRateVariation := (at.currentStrategy.HotDataThreshold - 0.75) * 0.5
	hitRate := math.Max(0.5, math.Min(0.95, baseHitRate+hitRateVariation))

	baseResponseTime := 30 * time.Millisecond
	responseTimeImprovement := time.Duration(at.currentStrategy.HotDataThreshold * 20 * float64(time.Millisecond))
	responseTime := time.Duration(math.Max(float64(5*time.Millisecond),
		float64(baseResponseTime-responseTimeImprovement)))

	memoryUsage := int64(5000) * 1024 // 默认5000个条目，每个1KB

	return &PerformanceMetrics{
		HitRate:        hitRate,
		AvgResponseTime: responseTime,
		MemoryUsage:    memoryUsage,
		CPULoad:        0.2 + hitRate*0.3, // 简单的CPU负载模型
		LastUpdateTime: time.Now(),
	}
}

// addPerformanceSnapshot 添加性能快照
func (at *AutoTuner) addPerformanceSnapshot(snapshot *PerformanceSnapshot) {
	at.mu.Lock()
	defer at.mu.Unlock()

	at.performanceHistory = append(at.performanceHistory, *snapshot)

	// 保持历史记录大小在合理范围内
	maxHistory := 100
	if len(at.performanceHistory) > maxHistory {
		at.performanceHistory = at.performanceHistory[len(at.performanceHistory)-maxHistory:]
	}
}

// analyzeAndPlanTuning 分析并规划调优
func (at *AutoTuner) analyzeAndPlanTuning() []TuningAction {
	at.mu.RLock()
	defer at.mu.RUnlock()

	if len(at.performanceHistory) < at.config.MinDataPoints {
		return nil
	}

	actions := make([]TuningAction, 0)

	// 获取最近的性能趋势
	recentTrend := at.calculatePerformanceTrend()

	// 命中率分析
	hitRateAction := at.analyzeHitRateTrend(recentTrend)
	if hitRateAction != nil {
		actions = append(actions, *hitRateAction)
	}

	// 响应时间分析
	responseTimeAction := at.analyzeResponseTimeTrend(recentTrend)
	if responseTimeAction != nil {
		actions = append(actions, *responseTimeAction)
	}

	// 内存使用分析
	memoryAction := at.analyzeMemoryUsageTrend(recentTrend)
	if memoryAction != nil {
		actions = append(actions, *memoryAction)
	}

	// 负载感知调优
	if at.config.LoadBasedTuning {
		loadActions := at.analyzeLoadBasedTuning(recentTrend)
		actions = append(actions, loadActions...)
	}

	// 预测性调优
	if at.config.EnablePredictiveTuning {
		predictiveActions := at.analyzePredictiveTuning()
		actions = append(actions, predictiveActions...)
	}

	return actions
}

// PerformanceTrend 性能趋势
type PerformanceTrend struct {
	HitRateTrend        float64 `json:"hit_rate_trend"`        // 命中率趋势 (-1 to 1)
	ResponseTimeTrend   float64 `json:"response_time_trend"`   // 响应时间趋势 (-1 to 1)
	MemoryTrend         float64 `json:"memory_trend"`         // 内存趋势 (-1 to 1)
	CurrentHitRate      float64 `json:"current_hit_rate"`
	CurrentResponseTime time.Duration `json:"current_response_time"`
	CurrentMemoryUsage  int64   `json:"current_memory_usage"`
}

// calculatePerformanceTrend 计算性能趋势
func (at *AutoTuner) calculatePerformanceTrend() *PerformanceTrend {
	if len(at.performanceHistory) < 3 {
		return nil
	}

	// 获取最近的几个数据点
	recent := at.performanceHistory[len(at.performanceHistory)-3:]

	// 计算趋势（简单线性回归）
	hitRateTrend := at.calculateTrend([]float64{
		recent[0].Metrics.HitRate,
		recent[1].Metrics.HitRate,
		recent[2].Metrics.HitRate,
	})

	responseTimeTrend := at.calculateTrend([]float64{
		float64(recent[0].Metrics.AvgResponseTime),
		float64(recent[1].Metrics.AvgResponseTime),
		float64(recent[2].Metrics.AvgResponseTime),
	})

	memoryTrend := at.calculateTrend([]float64{
		float64(recent[0].Metrics.MemoryUsage),
		float64(recent[1].Metrics.MemoryUsage),
		float64(recent[2].Metrics.MemoryUsage),
	})

	latest := recent[len(recent)-1]

	return &PerformanceTrend{
		HitRateTrend:        hitRateTrend,
		ResponseTimeTrend:   -responseTimeTrend, // 响应时间越小越好，所以取反
		MemoryTrend:         -memoryTrend,        // 内存使用越小越好，所以取反
		CurrentHitRate:      latest.Metrics.HitRate,
		CurrentResponseTime: latest.Metrics.AvgResponseTime,
		CurrentMemoryUsage:  latest.Metrics.MemoryUsage,
	}
}

// calculateTrend 计算趋势（返回-1到1之间的值）
func (at *AutoTuner) calculateTrend(values []float64) float64 {
	if len(values) < 2 {
		return 0
	}

	// 计算简单线性回归斜率
	n := float64(len(values))
	var sumX, sumY, sumXY, sumX2 float64

	for i, y := range values {
		x := float64(i)
		sumX += x
		sumY += y
		sumXY += x * y
		sumX2 += x * x
	}

	denominator := n*sumX2 - sumX*sumX
	if denominator == 0 {
		return 0
	}

	slope := (n*sumXY - sumX*sumY) / denominator

	// 归一化到-1到1范围
	maxSlope := 0.1 // 最大斜率
	return math.Max(-1, math.Min(1, slope/maxSlope))
}

// analyzeHitRateTrend 分析命中率趋势
func (at *AutoTuner) analyzeHitRateTrend(trend *PerformanceTrend) *TuningAction {
	if trend == nil {
		return nil
	}

	targetHitRate := at.config.TargetHitRate
	currentHitRate := trend.CurrentHitRate

	// 如果命中率达标且趋势良好，不需要调整
	if currentHitRate >= targetHitRate && trend.HitRateTrend >= 0 {
		return nil
	}

	// 如果命中率低于可接受水平或呈下降趋势，需要调整
	if currentHitRate < at.config.AcceptableHitRate || trend.HitRateTrend < -0.1 {
		// 降低热点阈值，让更多数据进入L1缓存
		newThreshold := at.currentStrategy.HotDataThreshold * (1 - at.config.MaxAdjustmentPerStep*at.config.ReactionSpeed)
		newThreshold = math.Max(at.config.HotThresholdRange[0], newThreshold)

		return &TuningAction{
			Timestamp:      time.Now(),
			Action:         "adjust_hot_threshold",
			Parameter:      "HotDataThreshold",
			OldValue:       at.currentStrategy.HotDataThreshold,
			NewValue:       newThreshold,
			Reason:         "low_hit_rate_or_declining",
			ExpectedImpact: "increase_hit_rate",
		}
	}

	return nil
}

// analyzeResponseTimeTrend 分析响应时间趋势
func (at *AutoTuner) analyzeResponseTimeTrend(trend *PerformanceTrend) *TuningAction {
	if trend == nil {
		return nil
	}

	targetResponseTime := at.config.TargetResponseTime
	currentResponseTime := trend.CurrentResponseTime

	// 如果响应时间达标且趋势良好，不需要调整
	if currentResponseTime <= targetResponseTime && trend.ResponseTimeTrend >= 0 {
		return nil
	}

	// 如果响应时间超标或呈恶化趋势，需要调整
	if currentResponseTime > at.config.MaxResponseTime || trend.ResponseTimeTrend < -0.1 {
		// 提高热点阈值，保持更多热点数据在L1缓存中
		newThreshold := at.currentStrategy.HotDataThreshold * (1 + at.config.MaxAdjustmentPerStep*at.config.ReactionSpeed)
		newThreshold = math.Min(at.config.HotThresholdRange[1], newThreshold)

		return &TuningAction{
			Timestamp:      time.Now(),
			Action:         "adjust_hot_threshold",
			Parameter:      "HotDataThreshold",
			OldValue:       at.currentStrategy.HotDataThreshold,
			NewValue:       newThreshold,
			Reason:         "high_response_time_or_degrading",
			ExpectedImpact: "reduce_response_time",
		}
	}

	return nil
}

// analyzeMemoryUsageTrend 分析内存使用趋势
func (at *AutoTuner) analyzeMemoryUsageTrend(trend *PerformanceTrend) *TuningAction {
	if trend == nil {
		return nil
	}

	currentMemoryUsage := trend.CurrentMemoryUsage
	maxMemoryUsage := at.config.MaxMemoryUsage

	// 如果内存使用接近上限，需要调整
	if currentMemoryUsage > int64(float64(maxMemoryUsage)*0.9) {
		// 提高热点阈值，减少L1缓存的数据量
		newThreshold := at.currentStrategy.HotDataThreshold * (1 + at.config.MaxAdjustmentPerStep*at.config.ReactionSpeed)
		newThreshold = math.Min(at.config.HotThresholdRange[1], newThreshold)

		return &TuningAction{
			Timestamp:      time.Now(),
			Action:         "adjust_hot_threshold",
			Parameter:      "HotDataThreshold",
			OldValue:       at.currentStrategy.HotDataThreshold,
			NewValue:       newThreshold,
			Reason:         "high_memory_usage",
			ExpectedImpact: "reduce_memory_usage",
		}
	}

	return nil
}

// analyzeLoadBasedTuning 分析基于负载的调优
func (at *AutoTuner) analyzeLoadBasedTuning(trend *PerformanceTrend) []TuningAction {
	actions := make([]TuningAction, 0)

	if len(at.performanceHistory) == 0 {
		return actions
	}

	// 获取最新的负载因子
	latest := at.performanceHistory[len(at.performanceHistory)-1]
	loadFactor := latest.LoadFactor

	// 高负载时的调优策略
	if loadFactor > 0.8 {
		// 提高热点阈值，优化性能
		newThreshold := at.currentStrategy.HotDataThreshold * (1 + at.config.MaxAdjustmentPerStep*0.5)
		newThreshold = math.Min(at.config.HotThresholdRange[1], newThreshold)

		actions = append(actions, TuningAction{
			Timestamp:      time.Now(),
			Action:         "load_based_adjustment",
			Parameter:      "HotDataThreshold",
			OldValue:       at.currentStrategy.HotDataThreshold,
			NewValue:       newThreshold,
			Reason:         "high_load_detected",
			ExpectedImpact: "optimize_performance_under_load",
		})
	}

	// 低负载时的调优策略
	if loadFactor < 0.3 {
		// 降低热点阈值，提高命中率
		newThreshold := at.currentStrategy.HotDataThreshold * (1 - at.config.MaxAdjustmentPerStep*0.5)
		newThreshold = math.Max(at.config.HotThresholdRange[0], newThreshold)

		actions = append(actions, TuningAction{
			Timestamp:      time.Now(),
			Action:         "load_based_adjustment",
			Parameter:      "HotDataThreshold",
			OldValue:       at.currentStrategy.HotDataThreshold,
			NewValue:       newThreshold,
			Reason:         "low_load_detected",
			ExpectedImpact: "improve_hit_rate_under_low_load",
		})
	}

	return actions
}

// analyzePredictiveTuning 分析预测性调优
func (at *AutoTuner) analyzePredictiveTuning() []TuningAction {
	actions := make([]TuningAction, 0)

	if len(at.performanceHistory) < 5 {
		return actions
	}

	// 预测未来性能趋势
	prediction := at.predictPerformanceTrend()

	// 基于预测进行预防性调优
	if prediction.PredictedHitRate < at.config.TargetHitRate*0.9 {
		// 预测命中率将下降，提前调整
		newThreshold := at.currentStrategy.HotDataThreshold * (1 - at.config.MaxAdjustmentPerStep*0.3)
		newThreshold = math.Max(at.config.HotThresholdRange[0], newThreshold)

		actions = append(actions, TuningAction{
			Timestamp:      time.Now(),
			Action:         "predictive_adjustment",
			Parameter:      "HotDataThreshold",
			OldValue:       at.currentStrategy.HotDataThreshold,
			NewValue:       newThreshold,
			Reason:         "predicted_hit_rate_decline",
			ExpectedImpact: "prevent_hit_rate_decline",
		})
	}

	return actions
}

// PerformancePrediction 性能预测
type PerformancePrediction struct {
	PredictedHitRate      float64     `json:"predicted_hit_rate"`
	PredictedResponseTime time.Duration `json:"predicted_response_time"`
	PredictedMemoryUsage  int64       `json:"predicted_memory_usage"`
	Confidence           float64     `json:"confidence"`
}

// predictPerformanceTrend 预测性能趋势
func (at *AutoTuner) predictPerformanceTrend() *PerformancePrediction {
	if len(at.performanceHistory) < 5 {
		return nil
	}

	// 获取最近的5个数据点用于预测
	recent := at.performanceHistory[len(at.performanceHistory)-5:]

	// 简单的线性预测
	hitRateValues := make([]float64, len(recent))
	responseTimeValues := make([]float64, len(recent))
	memoryValues := make([]float64, len(recent))

	for i, snapshot := range recent {
		hitRateValues[i] = snapshot.Metrics.HitRate
		responseTimeValues[i] = float64(snapshot.Metrics.AvgResponseTime)
		memoryValues[i] = float64(snapshot.Metrics.MemoryUsage)
	}

	// 计算趋势并预测下一个值
	hitRateTrend := at.calculateTrend(hitRateValues)
	responseTimeTrend := at.calculateTrend(responseTimeValues)
	memoryTrend := at.calculateTrend(memoryValues)

	latest := recent[len(recent)-1]
	predictedHitRate := math.Max(0, math.Min(1, latest.Metrics.HitRate+hitRateTrend*0.1))
	predictedResponseTime := time.Duration(math.Max(0,
		float64(latest.Metrics.AvgResponseTime)+responseTimeTrend*float64(time.Millisecond)*1000))
	predictedMemoryUsage := int64(math.Max(0,
		float64(latest.Metrics.MemoryUsage)+memoryTrend*1024))

	// 计算预测置信度
	confidence := at.calculatePredictionConfidence(hitRateValues, responseTimeValues, memoryValues)

	return &PerformancePrediction{
		PredictedHitRate:      predictedHitRate,
		PredictedResponseTime: predictedResponseTime,
		PredictedMemoryUsage:  predictedMemoryUsage,
		Confidence:           confidence,
	}
}

// calculatePredictionConfidence 计算预测置信度
func (at *AutoTuner) calculatePredictionConfidence(hitRates, responseTimes, memoryUsages []float64) float64 {
	// 基于数据的稳定性计算置信度
	hitRateStability := 1.0 - at.calculateVariance(hitRates)
	responseTimeStability := 1.0 - at.calculateVariance(responseTimes)
	memoryStability := 1.0 - at.calculateVariance(memoryUsages)

	// 综合置信度
	confidence := (hitRateStability + responseTimeStability + memoryStability) / 3.0
	return math.Max(0, math.Min(1, confidence))
}

// calculateVariance 计算方差
func (at *AutoTuner) calculateVariance(values []float64) float64 {
	if len(values) < 2 {
		return 0
	}

	var sum, sumSquares float64
	for _, v := range values {
		sum += v
		sumSquares += v * v
	}

	mean := sum / float64(len(values))
	variance := (sumSquares / float64(len(values))) - (mean * mean)

	// 归一化方差
	maxVariance := mean * mean
	if maxVariance == 0 {
		return 0
	}

	return math.Min(1, variance/maxVariance)
}

// executeTuningAction 执行调优动作
func (at *AutoTuner) executeTuningAction(ctx context.Context, action TuningAction) error {
	at.mu.Lock()
	defer at.mu.Unlock()

	// 执行具体的调优动作
	switch action.Parameter {
	case "HotDataThreshold":
		if newThreshold, ok := action.NewValue.(float64); ok {
			at.currentStrategy.HotDataThreshold = newThreshold
		}
	case "WarmDataThreshold":
		if newThreshold, ok := action.NewValue.(float64); ok {
			at.currentStrategy.WarmDataThreshold = newThreshold
		}
	default:
		return fmt.Errorf("unknown parameter: %s", action.Parameter)
	}

	// 记录调优动作
	action.Timestamp = time.Now()
	at.tuningHistory = append(at.tuningHistory, action)

	// 限制历史记录大小
	if len(at.tuningHistory) > 100 {
		at.tuningHistory = at.tuningHistory[1:]
	}

	at.logger.Info("Executed tuning action",
		zap.String("action", action.Action),
		zap.String("parameter", action.Parameter),
		zap.Any("old_value", action.OldValue),
		zap.Any("new_value", action.NewValue),
		zap.String("reason", action.Reason),
	)

	return nil
}

// getCurrentPerformanceSnapshot 获取当前性能快照
func (at *AutoTuner) getCurrentPerformanceSnapshot() *PerformanceMetrics {
	if len(at.performanceHistory) > 0 {
		return at.performanceHistory[len(at.performanceHistory)-1].Metrics
	}

	// 返回默认性能指标
	return &PerformanceMetrics{
		HitRate:        0.0,
		AvgResponseTime: time.Hour,
		MemoryUsage:    0,
		CPULoad:        0,
		LastUpdateTime: time.Now(),
	}
}

// GetCurrentStrategy 获取当前策略
func (at *AutoTuner) GetCurrentStrategy() *CacheStrategy {
	at.mu.RLock()
	defer at.mu.RUnlock()

	// 返回策略的副本
	strategyCopy := *at.currentStrategy
	return &strategyCopy
}

// GetTuningHistory 获取调优历史
func (at *AutoTuner) GetTuningHistory(limit int) []TuningAction {
	at.mu.RLock()
	defer at.mu.RUnlock()

	if limit <= 0 || limit > len(at.tuningHistory) {
		limit = len(at.tuningHistory)
	}

	result := make([]TuningAction, limit)
	copy(result, at.tuningHistory[len(at.tuningHistory)-limit:])
	return result
}

// GetPerformanceHistory 获取性能历史
func (at *AutoTuner) GetPerformanceHistory(limit int) []PerformanceSnapshot {
	at.mu.RLock()
	defer at.mu.RUnlock()

	if limit <= 0 || limit > len(at.performanceHistory) {
		limit = len(at.performanceHistory)
	}

	result := make([]PerformanceSnapshot, limit)
	copy(result, at.performanceHistory[len(at.performanceHistory)-limit:])
	return result
}

// GetTuningStats 获取调优统计
func (at *AutoTuner) GetTuningStats() map[string]interface{} {
	at.mu.RLock()
	defer at.mu.RUnlock()

	stats := map[string]interface{}{
		"is_active":              at.isActive,
		"total_tuning_actions":   len(at.tuningHistory),
		"performance_snapshots":  len(at.performanceHistory),
		"current_strategy":       at.currentStrategy,
		"last_tuning_time":       "none",
		"successful_tunings":     0,
		"failed_tunings":         0,
	}

	if len(at.tuningHistory) > 0 {
		lastAction := at.tuningHistory[len(at.tuningHistory)-1]
		stats["last_tuning_time"] = lastAction.Timestamp.Format(time.RFC3339)

		// 统计成功和失败的调优
		successCount := 0
		failedCount := 0
		for _, action := range at.tuningHistory {
			if action.Success {
				successCount++
			} else {
				failedCount++
			}
		}
		stats["successful_tunings"] = successCount
		stats["failed_tunings"] = failedCount
	}

	return stats
}

// UpdateConfig 更新调优配置
func (at *AutoTuner) UpdateConfig(config *AutoTuningConfig) {
	at.mu.Lock()
	defer at.mu.Unlock()

	at.config = config
	at.logger.Info("Auto tuning config updated")
}

// Reset 重置调优器状态
func (at *AutoTuner) Reset() {
	at.mu.Lock()
	defer at.mu.Unlock()

	at.performanceHistory = make([]PerformanceSnapshot, 0)
	at.tuningHistory = make([]TuningAction, 0)
	at.logger.Info("Auto tuner reset")
}