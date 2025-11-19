package cache

import (
	"context"
	"fmt"
	"math"
	"sync"
	"time"

	"go.uber.org/zap"
)

// EnhancedCacheManager 增强的缓存管理器
type EnhancedCacheManager struct {
	*ThreeLevelCacheManager  // 嵌入基础管理器
	adaptiveStrategy       *AdaptiveStrategy
	predictiveCache        *PredictiveCache
	performanceMonitor     *PerformanceMonitor
	optimizer              *CacheOptimizer
	autoTuner              *AutoTuner
	mu                    sync.RWMutex
	logger                *zap.Logger
}

// PerformanceMonitor 性能监控器
type PerformanceMonitor struct {
	mu                sync.RWMutex
	metrics           *PerformanceMetrics
	adjustmentHistory []PerformanceAdjustment
	alertThresholds   *AlertThresholds
	logger            *zap.Logger
}

// PerformanceMetrics 性能指标
type PerformanceMetrics struct {
	TotalRequests       int64         `json:"total_requests"`
	AvgResponseTime     time.Duration `json:"avg_response_time"`
	P95ResponseTime    time.Duration `json:"p95_response_time"`
	HitRate            float64       `json:"hit_rate"`
	QPS                float64       `json:"qps"`
	MemoryUsage        int64         `json:"memory_usage"`
	CPULoad            float64       `json:"cpu_load"`
	NetworkLatency     time.Duration `json:"network_latency"`
	DiskIOPS           int64         `json:"disk_iops"`
	LastUpdateTime     time.Time     `json:"last_update_time"`
}

// PerformanceAdjustment 性能调整记录
type PerformanceAdjustment struct {
	Timestamp time.Time `json:"timestamp"`
	Metric    string    `json:"metric"`
	Action    string    `json:"action"`
	OldValue  float64   `json:"old_value"`
	NewValue  float64   `json:"new_value"`
	Reason    string    `json:"reason"`
}

// AlertThresholds 告警阈值
type AlertThresholds struct {
	ResponseTimeThreshold time.Duration `json:"response_time_threshold"`
	HitRateThreshold       float64       `json:"hit_rate_threshold"`
	QPSThreshold           float64       `json:"qps_threshold"`
	MemoryThreshold        int64         `json:"memory_threshold"`
	CPUThreshold           float64       `json:"cpu_threshold"`
}

// EnhancedCacheConfig 增强缓存配置
type EnhancedCacheConfig struct {
	*CacheStrategy
	*PredictionConfig
	*StrategyTuningConfig
	*AlertThresholds
	*OptimizationConfig
	*AutoTuningConfig
	EnableAdaptive     bool `json:"enable_adaptive"`
	EnablePredictive    bool `json:"enable_predictive"`
	EnableMonitoring    bool `json:"enable_monitoring"`
	EnableOptimizer     bool `json:"enable_optimizer"`
	EnableAutoTuning    bool `json:"enable_auto_tuning"`
}

// DefaultEnhancedCacheConfig 默认增强缓存配置
func DefaultEnhancedCacheConfig() *EnhancedCacheConfig {
	return &EnhancedCacheConfig{
		CacheStrategy:        DefaultCacheStrategy(),
		PredictionConfig:     DefaultPredictionConfig(),
		StrategyTuningConfig: DefaultTuningConfig(),
		AlertThresholds: &AlertThresholds{
			ResponseTimeThreshold: 50 * time.Millisecond,
			HitRateThreshold:       0.7,
			QPSThreshold:           20000,
			MemoryThreshold:        100 * 1024 * 1024, // 100MB
			CPUThreshold:           0.8,
		},
		OptimizationConfig:  DefaultOptimizationConfig(),
		AutoTuningConfig:    DefaultAutoTuningConfig(),
		EnableAdaptive:      true,
		EnablePredictive:    true,
		EnableMonitoring:    true,
		EnableOptimizer:     true,
		EnableAutoTuning:    true,
	}
}

// NewEnhancedCacheManager 创建增强的缓存管理器
func NewEnhancedCacheManager(
	l1Cache L1Cache,
	l2Cache L2Cache,
	tracker FrequencyTracker,
	config *EnhancedCacheConfig,
	logger *zap.Logger,
) *EnhancedCacheManager {
	if config == nil {
		config = DefaultEnhancedCacheConfig()
	}

	// 创建基础缓存管理器
	baseManager := NewThreeLevelCacheManager(l1Cache, l2Cache, tracker, config.CacheStrategy, logger)

	enhanced := &EnhancedCacheManager{
		ThreeLevelCacheManager: baseManager,
		logger:                logger.Named("enhanced_cache_manager"),
	}

	// 创建自适应策略
	if config.EnableAdaptive {
		enhanced.adaptiveStrategy = NewAdaptiveStrategy(config.CacheStrategy, config.StrategyTuningConfig, logger)
	}

	// 创建预测性缓存
	if config.EnablePredictive {
		enhanced.predictiveCache = NewPredictiveCache(enhanced, config.PredictionConfig, logger)
	}

	// 创建性能监控
	if config.EnableMonitoring {
		enhanced.performanceMonitor = NewPerformanceMonitor(config.AlertThresholds, logger)
	}

	// 创建优化器
	if config.EnableOptimizer {
		enhanced.optimizer = NewCacheOptimizer(config.CacheStrategy, config.OptimizationConfig, logger)
	}

	// 创建自动调优器
	if config.EnableAutoTuning {
		enhanced.autoTuner = NewAutoTuner(config.CacheStrategy, config.AutoTuningConfig, logger)
		// 设置性能指标收集器
		enhanced.autoTuner.SetMetricsCollector(enhanced)
	}

	// 启动后台任务
	go enhanced.startBackgroundTasks()

	logger.Info("Enhanced cache manager initialized",
		zap.Bool("adaptive", config.EnableAdaptive),
		zap.Bool("predictive", config.EnablePredictive),
		zap.Bool("monitoring", config.EnableMonitoring),
		zap.Bool("optimizer", config.EnableOptimizer),
		zap.Bool("auto_tuning", config.EnableAutoTuning),
	)

	return enhanced
}

// Get 获取缓存值（增强版）
func (ecm *EnhancedCacheManager) Get(ctx context.Context, key string) (interface{}, error) {
	defer func() {
		// 记录访问时间用于性能监控
		if ecm.performanceMonitor != nil {
			ecm.performanceMonitor.RecordRequest(true)
		}
	}()

	// 记录访问用于预测
	if ecm.predictiveCache != nil {
		ecm.predictiveCache.RecordAccess(key)
	}

	// 调用基础Get方法
	value, err := ecm.ThreeLevelCacheManager.Get(ctx, key)
	if err != nil {
		return nil, err
	}

	// 更新性能监控
	if ecm.performanceMonitor != nil {
		ecm.performanceMonitor.RecordRequest(true)
	}

	return value, nil
}

// Set 设置缓存值（增强版）
func (ecm *EnhancedCacheManager) Set(ctx context.Context, key string, value interface{}, ttl time.Duration) error {
	// 调用基础Set方法
	err := ecm.ThreeLevelCacheManager.Set(ctx, key, value, ttl)
	if err != nil {
		return err
	}

	// 更新性能监控
	if ecm.performanceMonitor != nil {
		ecm.performanceMonitor.RecordRequest(false)
	}

	return nil
}

// GetAdaptiveStrategy 获取自适应策略
func (ecm *EnhancedCacheManager) GetAdaptiveStrategy() *AdaptiveStrategy {
	ecm.mu.RLock()
	defer ecm.mu.RUnlock()
	return ecm.adaptiveStrategy
}

// GetPredictiveCache 获取预测性缓存
func (ecm *EnhancedCacheManager) GetPredictiveCache() *PredictiveCache {
	ecm.mu.RLock()
	defer ecm.mu.RUnlock()
	return ecm.predictiveCache
}

// GetPerformanceMonitor 获取性能监控器
func (ecm *EnhancedCacheManager) GetPerformanceMonitor() *PerformanceMonitor {
	ecm.mu.RLock()
	defer ecm.mu.RUnlock()
	return ecm.performanceMonitor
}

// GetOptimizer 获取优化器
func (ecm *EnhancedCacheManager) GetOptimizer() *CacheOptimizer {
	ecm.mu.RLock()
	defer ecm.mu.RUnlock()
	return ecm.optimizer
}

// GetAutoTuner 获取自动调优器
func (ecm *EnhancedCacheManager) GetAutoTuner() *AutoTuner {
	ecm.mu.RLock()
	defer ecm.mu.RUnlock()
	return ecm.autoTuner
}

// 实现 MetricsCollector 接口

// GetCurrentMetrics 获取当前性能指标
func (ecm *EnhancedCacheManager) GetCurrentMetrics(ctx context.Context) (*PerformanceMetrics, error) {
	if ecm.performanceMonitor == nil {
		return &PerformanceMetrics{
			HitRate:        0.0,
			AvgResponseTime: time.Hour,
			MemoryUsage:    0,
			CPULoad:        0,
			LastUpdateTime: time.Now(),
		}, nil
	}
	return ecm.performanceMonitor.GetCurrentMetrics(), nil
}

// GetLoadFactor 获取负载因子
func (ecm *EnhancedCacheManager) GetLoadFactor(ctx context.Context) (float64, error) {
	stats, err := ecm.GetEnhancedStats(ctx)
	if err != nil {
		return 0.5, nil // 默认中等负载
	}

	if stats.PerformanceMetrics == nil {
		return 0.5, nil
	}

	// 基于QPS和响应时间计算负载因子
	qps := stats.PerformanceMetrics.QPS
	maxQPS := 50000.0
	loadFactor := math.Min(1.0, qps/maxQPS)

	// 考虑响应时间影响
	responseTimeRatio := float64(stats.PerformanceMetrics.AvgResponseTime) / float64(100*time.Millisecond)
	loadFactor = math.Max(loadFactor, responseTimeRatio)

	return math.Min(1.0, loadFactor), nil
}

// AnalyzeAccessPattern 分析访问模式
func (ecm *EnhancedCacheManager) AnalyzeAccessPattern(ctx context.Context) (string, error) {
	if ecm.predictiveCache != nil {
		stats := ecm.predictiveCache.GetPatternStats()

		totalPatterns, ok := stats["total_patterns"].(int)
		if !ok || totalPatterns == 0 {
			return "unknown", nil
		}

		predictablePatterns, ok := stats["predictable_patterns"].(int)
		if !ok {
			predictablePatterns = 0
		}

		predictableRatio := float64(predictablePatterns) / float64(totalPatterns)

		if predictableRatio > 0.7 {
			return "highly_predictable", nil
		} else if predictableRatio > 0.4 {
			return "moderately_predictable", nil
		} else if predictableRatio > 0.2 {
			return "slightly_predictable", nil
		} else {
			return "random_access", nil
		}
	}

	return "unknown", nil
}

// GetEnhancedStats 获取增强统计信息
func (ecm *EnhancedCacheManager) GetEnhancedStats(ctx context.Context) (*EnhancedStats, error) {
	// 获取基础统计
	baseStats, err := ecm.ThreeLevelCacheManager.GetStats(ctx)
	if err != nil {
		return nil, err
	}

	enhancedStats := &EnhancedStats{
		CacheStats: baseStats,
	}

	// 添加自适应策略统计
	if ecm.adaptiveStrategy != nil {
		enhancedStats.AdaptiveStrategy = ecm.adaptiveStrategy.GetCurrentStats()
		enhancedStats.AdjustmentHistory = ecm.adaptiveStrategy.GetAdjustmentHistory(10)
	}

	// 添加预测性缓存统计
	if ecm.predictiveCache != nil {
		enhancedStats.PredictiveStats = ecm.predictiveCache.GetPatternStats()
	}

	// 添加性能监控统计
	if ecm.performanceMonitor != nil {
		enhancedStats.PerformanceMetrics = ecm.performanceMonitor.GetCurrentMetrics()
		enhancedStats.PerformanceHistory = ecm.performanceMonitor.GetRecentHistory(50)
	}

	return enhancedStats, nil
}

// EnhancedStats 增强统计信息
type EnhancedStats struct {
	*CacheStats
	AdaptiveStrategy    *StrategyStats            `json:"adaptive_strategy"`
	AdjustmentHistory    []StrategyAdjustmentRecord `json:"adjustment_history"`
	PredictiveStats     map[string]interface{}    `json:"predictive_stats"`
	PerformanceMetrics *PerformanceMetrics      `json:"performance_metrics"`
	PerformanceHistory []PerformanceAdjustment  `json:"performance_history"`
}

// UpdatePerformanceMetrics 更新性能指标
func (ecm *EnhancedCacheManager) UpdatePerformanceMetrics(metrics *PerformanceMetrics) {
	if ecm.adaptiveStrategy != nil {
		// 直接使用基础缓存统计
		stats := &CacheStats{}
		ecm.adaptiveStrategy.UpdateStats(stats, metrics.MemoryUsage)
	}

	if ecm.performanceMonitor != nil {
		ecm.performanceMonitor.UpdateMetrics(metrics)
	}
}

// PredictNextAccess 预测下次访问
func (ecm *EnhancedCacheManager) PredictNextAccess(ctx context.Context, key string) (*Prediction, error) {
	if ecm.predictiveCache == nil {
		return nil, fmt.Errorf("predictive cache not enabled")
	}

	return ecm.predictiveCache.PredictNextAccess(ctx, key)
}

// PreloadPredictions 预加载预测的缓存
func (ecm *EnhancedCacheManager) PreloadPredictions(ctx context.Context) error {
	if ecm.predictiveCache == nil {
		return nil
	}

	// 收集所有预测
	predictions := make([]*Prediction, 0)
	ecm.mu.RLock()
		if ecm.predictiveCache != nil {
		// 这里需要暴露predictiveCache的内部数据
		// 简化实现，直接预测常用键
		commonKeys := []string{"rules:project:default", "rules:environment:default:dev"}
		for _, key := range commonKeys {
			if prediction, err := ecm.PredictNextAccess(ctx, key); err == nil {
				predictions = append(predictions, prediction)
			}
		}
	}
	ecm.mu.RUnlock()

	// 执行预加载
	return ecm.predictiveCache.PreloadIfNeeded(ctx, predictions)
}

// OptimizeCache 优化缓存性能
func (ecm *EnhancedCacheManager) OptimizeCache(ctx context.Context) (*OptimizationResult, error) {
	result := &OptimizationResult{
		Timestamp: time.Now(),
		Actions:   make([]string, 0),
		Improvements: make(map[string]float64),
	}

	// 获取当前性能指标
	currentStats, err := ecm.GetEnhancedStats(ctx)
	if err != nil {
		return nil, err
	}

	// 分析性能问题并制定优化策略
	if currentStats.PerformanceMetrics != nil {
		metrics := currentStats.PerformanceMetrics

	// 响应时间优化
		if metrics.AvgResponseTime > 50*time.Millisecond {
			result.Actions = append(result.Actions, "response_time_optimization")
			if ecm.adaptiveStrategy != nil {
				// 触发自适应调整
				ecm.triggerAdaptiveAdjustment("slow_response_time")
			}
			result.Improvements["response_time"] = 0.3 // 预期30%改善
		}

	// 命中率优化
		if metrics.HitRate < 0.8 {
			result.Actions = append(result.Actions, "hit_rate_optimization")
			if ecm.adaptiveStrategy != nil {
				ecm.triggerAdaptiveAdjustment("low_hit_rate")
			}
			result.Improvements["hit_rate"] = 0.2 // 预期20%改善
		}

	// 内存使用优化
		if metrics.MemoryUsage > 90*1024*1024 { // 90MB
			result.Actions = append(result.Actions, "memory_optimization")
			if ecm.adaptiveStrategy != nil {
				ecm.triggerAdaptiveAdjustment("high_memory_usage")
			}
			result.Improvements["memory_efficiency"] = 0.15 // 预期15%改善
		}
	}

	return result, nil
}

// OptimizationResult 优化结果
type OptimizationResult struct {
	Timestamp   time.Time           `json:"timestamp"`
	Actions     []string           `json:"actions"`
	Improvements map[string]float64 `json:"improvements"`
}

// triggerAdaptiveAdjustment 触发自适应调整
func (ecm *EnhancedCacheManager) triggerAdaptiveAdjustment(reason string) {
	if ecm.adaptiveStrategy == nil {
		return
	}

	// 这里可以通过设置特定的性能指标来触发调整
	// 简化实现
	ecm.logger.Info("Triggering adaptive adjustment",
		zap.String("reason", reason),
	)
}

// startBackgroundTasks 启动后台任务
func (ecm *EnhancedCacheManager) startBackgroundTasks() {
	// 性能监控任务
	monitoringTicker := time.NewTicker(1 * time.Minute)
	defer monitoringTicker.Stop()

	// 预测预加载任务
	preloadTicker := time.NewTicker(30 * time.Minute)
	defer preloadTicker.Stop()

	// 优化器启动任务
	optimizerTicker := time.NewTicker(5 * time.Minute)
	defer optimizerTicker.Stop()

	for {
		select {
		case <-monitoringTicker.C:
			ecm.performPerformanceMonitoring()
		case <-preloadTicker.C:
			ecm.performPredictivePreload()
		case <-optimizerTicker.C:
			ecm.startOptimizationIfNeeded()
		}
	}
}

// performPerformanceMonitoring 执行性能监控
func (ecm *EnhancedCacheManager) performPerformanceMonitoring() {
	if ecm.performanceMonitor == nil {
		return
	}

	// 收集性能指标
	stats, err := ecm.GetEnhancedStats(context.Background())
	if err != nil {
		ecm.logger.Error("Failed to get enhanced stats", zap.Error(err))
		return
	}

	// 更新性能指标
	if stats.PerformanceMetrics != nil {
		ecm.UpdatePerformanceMetrics(stats.PerformanceMetrics)
	}

	// 检查告警条件
	ecm.checkAlerts(stats.PerformanceMetrics)
}

// performPredictivePreload 执行预测性预加载
func (ecm *EnhancedCacheManager) performPredictivePreload() {
	if ecm.predictiveCache == nil {
		return
	}

	// 执行预加载
	err := ecm.PreloadPredictions(context.Background())
	if err != nil {
		ecm.logger.Debug("Predictive preload failed", zap.Error(err))
	}
}

// checkAlerts 检查告警条件
func (ecm *EnhancedCacheManager) checkAlerts(metrics *PerformanceMetrics) {
	if metrics == nil || ecm.performanceMonitor == nil {
		return
	}

	thresholds := ecm.performanceMonitor.alertThresholds
	alerts := make([]string, 0)

	// 响应时间告警
	if metrics.AvgResponseTime > thresholds.ResponseTimeThreshold {
		alerts = append(alerts, fmt.Sprintf("High response time: %v", metrics.AvgResponseTime))
	}

	// 命中率告警
	if metrics.HitRate < thresholds.HitRateThreshold {
		alerts = append(alerts, fmt.Sprintf("Low hit rate: %.2f%%", metrics.HitRate*100))
	}

	// QPS告警
	if metrics.QPS < thresholds.QPSThreshold {
		alerts = append(alerts, fmt.Sprintf("Low QPS: %.0f", metrics.QPS))
	}

	// 内存使用告警
	if metrics.MemoryUsage > thresholds.MemoryThreshold {
		alerts = append(alerts, fmt.Sprintf("High memory usage: %d MB", metrics.MemoryUsage/(1024*1024)))
	}

	// CPU使用告警
	if metrics.CPULoad > thresholds.CPUThreshold {
		alerts = append(alerts, fmt.Sprintf("High CPU load: %.2f%%", metrics.CPULoad*100))
	}

	// 发送告警
	if len(alerts) > 0 {
		ecm.logger.Warn("Performance alerts triggered",
			zap.Strings("alerts", alerts),
			zap.Time("timestamp", time.Now()),
		)
	}
}

// startOptimizationIfNeeded 根据需要启动优化
func (ecm *EnhancedCacheManager) startOptimizationIfNeeded() {
	ctx := context.Background()

	// 启动优化器
	if ecm.optimizer != nil {
		go func() {
			defer func() {
				if r := recover(); r != nil {
					ecm.logger.Error("Optimizer panic recovered", zap.Any("panic", r))
				}
			}()

			ecm.optimizer.StartOptimization(ctx)
		}()
	}

	// 启动自动调优器
	if ecm.autoTuner != nil {
		go func() {
			defer func() {
				if r := recover(); r != nil {
					ecm.logger.Error("Auto tuner panic recovered", zap.Any("panic", r))
				}
			}()

			if err := ecm.autoTuner.StartAutoTuning(ctx); err != nil {
				ecm.logger.Error("Failed to start auto tuning", zap.Error(err))
			}
		}()
	}
}

// StartOptimization 手动启动优化
func (ecm *EnhancedCacheManager) StartOptimization(ctx context.Context) error {
	if ecm.optimizer == nil {
		return fmt.Errorf("optimizer not enabled")
	}

	go ecm.optimizer.StartOptimization(ctx)
	return nil
}

// StopOptimization 停止优化
func (ecm *EnhancedCacheManager) StopOptimization() {
	if ecm.optimizer != nil {
		ecm.optimizer.StopOptimization()
	}
	if ecm.autoTuner != nil {
		ecm.autoTuner.StopAutoTuning()
	}
}

// GetOptimizationResults 获取优化结果
func (ecm *EnhancedCacheManager) GetOptimizationResults() (map[string]interface{}, error) {
	results := make(map[string]interface{})

	// 优化器结果
	if ecm.optimizer != nil {
		results["optimizer"] = ecm.optimizer.GetOptimizationStats()
		results["optimization_history"] = ecm.optimizer.GetOptimizationHistory(10)
		results["current_optimal_strategy"] = ecm.optimizer.GetCurrentStrategy()
	}

	// 自动调优器结果
	if ecm.autoTuner != nil {
		results["auto_tuner"] = ecm.autoTuner.GetTuningStats()
		results["tuning_history"] = ecm.autoTuner.GetTuningHistory(10)
		results["current_strategy"] = ecm.autoTuner.GetCurrentStrategy()
		results["performance_history"] = ecm.autoTuner.GetPerformanceHistory(20)
	}

	return results, nil
}

// UpdateOptimizationConfig 更新优化配置
func (ecm *EnhancedCacheManager) UpdateOptimizationConfig(optimizationConfig *OptimizationConfig, autoTuningConfig *AutoTuningConfig) error {
	ecm.mu.Lock()
	defer ecm.mu.Unlock()

	if optimizationConfig != nil && ecm.optimizer != nil {
		ecm.optimizer.SetOptimizationConfig(optimizationConfig)
	}

	if autoTuningConfig != nil && ecm.autoTuner != nil {
		ecm.autoTuner.UpdateConfig(autoTuningConfig)
	}

	ecm.logger.Info("Optimization configs updated")
	return nil
}

// ResetOptimization 重置优化状态
func (ecm *EnhancedCacheManager) ResetOptimization() {
	ecm.mu.Lock()
	defer ecm.mu.Unlock()

	if ecm.optimizer != nil {
		ecm.optimizer.ResetOptimization()
	}

	if ecm.autoTuner != nil {
		ecm.autoTuner.Reset()
	}

	ecm.logger.Info("Optimization state reset")
}

// NewPerformanceMonitor 创建性能监控器
func NewPerformanceMonitor(thresholds *AlertThresholds, logger *zap.Logger) *PerformanceMonitor {
	if thresholds == nil {
		thresholds = &AlertThresholds{
			ResponseTimeThreshold: 50 * time.Millisecond,
			HitRateThreshold:       0.7,
			QPSThreshold:           20000,
			MemoryThreshold:        100 * 1024 * 1024,
			CPUThreshold:           0.8,
		}
	}

	return &PerformanceMonitor{
		metrics:           &PerformanceMetrics{LastUpdateTime: time.Now()},
		adjustmentHistory: make([]PerformanceAdjustment, 0),
		alertThresholds:   thresholds,
		logger:           logger.Named("performance_monitor"),
	}
}

// RecordRequest 记录请求
func (pm *PerformanceMonitor) RecordRequest(hit bool) {
	pm.mu.Lock()
	defer pm.mu.Unlock()

	pm.metrics.TotalRequests++
	if hit {
		// 这里需要从缓存管理器获取实际的命中率
		// 简化实现
	}
	pm.metrics.LastUpdateTime = time.Now()
}

// UpdateMetrics 更新性能指标
func (pm *PerformanceMonitor) UpdateMetrics(metrics *PerformanceMetrics) {
	pm.mu.Lock()
	defer pm.mu.Unlock()

	pm.metrics = metrics
}

// GetCurrentMetrics 获取当前性能指标
func (pm *PerformanceMonitor) GetCurrentMetrics() *PerformanceMetrics {
	pm.mu.RLock()
	defer pm.mu.RUnlock()

	return &PerformanceMetrics{
		TotalRequests:    pm.metrics.TotalRequests,
		AvgResponseTime:   pm.metrics.AvgResponseTime,
		P95ResponseTime:  pm.metrics.P95ResponseTime,
		HitRate:          pm.metrics.HitRate,
		QPS:              pm.metrics.QPS,
		MemoryUsage:      pm.metrics.MemoryUsage,
		CPULoad:          pm.metrics.CPULoad,
		NetworkLatency:   pm.metrics.NetworkLatency,
		DiskIOPS:         pm.metrics.DiskIOPS,
		LastUpdateTime:   pm.metrics.LastUpdateTime,
	}
}

// GetRecentHistory 获取最近的历史记录
func (pm *PerformanceMonitor) GetRecentHistory(limit int) []PerformanceAdjustment {
	pm.mu.RLock()
	defer pm.mu.RUnlock()

	if limit <= 0 || limit > len(pm.adjustmentHistory) {
		limit = len(pm.adjustmentHistory)
	}

	result := make([]PerformanceAdjustment, limit)
	copy(result, pm.adjustmentHistory[len(pm.adjustmentHistory)-limit:])
	return result
}