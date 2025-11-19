package cache

import (
	"context"
	"fmt"
	"math"
	"sort"
	"sync"
	"sync/atomic"
	"time"

	"go.uber.org/zap"
)

// CacheMonitor 缓存监控器
type CacheMonitor struct {
	mu                    sync.RWMutex
	config                *MonitorConfig
	metrics               *DetailedMetrics
	alertManager          *AlertManager
	dashboard             *Dashboard
	exporters             map[string]MetricsExporter
	isRunning             bool
	stopCh                chan struct{}
	logger                *zap.Logger
}

// MonitorConfig 监控配置
type MonitorConfig struct {
	// 数据收集配置
	CollectionInterval    time.Duration `json:"collection_interval"`
	RetentionPeriod      time.Duration `json:"retention_period"`
	EnableDetailedStats  bool          `json:"enable_detailed_stats"`
	EnableRealTimeAlerts bool          `json:"enable_real_time_alerts"`

	// 指标配置
	MetricsEnabled       map[string]bool `json:"metrics_enabled"`
	Percentiles          []float64      `json:"percentiles"`
	HistogramBuckets     []float64      `json:"histogram_buckets"`

	// 告警配置
	AlertThresholds      *AlertThresholds `json:"alert_thresholds"`
	AlertCooldown        time.Duration    `json:"alert_cooldown"`
	MaxAlertsPerMinute   int              `json:"max_alerts_per_minute"`

	// 导出配置
	EnablePrometheus     bool     `json:"enable_prometheus"`
	EnableInfluxDB       bool     `json:"enable_influxdb"`
	EnableJSONExport     bool     `json:"enable_json_export"`
	ExportInterval       time.Duration `json:"export_interval"`
	ExportPath           string   `json:"export_path"`
}

// DetailedMetrics 详细指标
type DetailedMetrics struct {
	// 基础指标
	TotalRequests        int64     `json:"total_requests"`
	TotalHits            int64     `json:"total_hits"`
	TotalMisses          int64     `json:"total_misses"`
	HitRate              float64   `json:"hit_rate"`

	// 响应时间指标
	ResponseTimeStats    *ResponseTimeStats `json:"response_time_stats"`

	// 吞吐量指标
	ThroughputStats      *ThroughputStats   `json:"throughput_stats"`

	// 缓存层级指标
	LevelMetrics         map[string]*LevelMetrics `json:"level_metrics"`

	// 键级别指标
	TopKeys              []*KeyMetrics `json:"top_keys"`
	HotKeys              []*KeyMetrics `json:"hot_keys"`
	ColdKeys             []*KeyMetrics `json:"cold_keys"`

	// 内存指标
	MemoryMetrics        *MemoryMetrics `json:"memory_metrics"`

	// 网络指标
	NetworkMetrics       *NetworkMetrics `json:"network_metrics"`

	// 错误指标
	ErrorMetrics         *ErrorMetrics `json:"error_metrics"`

	// 时间窗口指标
	TimeWindowMetrics    map[string]*WindowMetrics `json:"time_window_metrics"`

	// 趋势指标
	TrendMetrics         *TrendMetrics `json:"trend_metrics"`

	// 最后更新时间
	LastUpdateTime       time.Time `json:"last_update_time"`
}

// ResponseTimeStats 响应时间统计
type ResponseTimeStats struct {
	Min                  time.Duration `json:"min"`
	Max                  time.Duration `json:"max"`
	Mean                 time.Duration `json:"mean"`
	Median               time.Duration `json:"median"`
	P50                  time.Duration `json:"p50"`
	P75                  time.Duration `json:"p75"`
	P90                  time.Duration `json:"p90"`
	P95                  time.Duration `json:"p95"`
	P99                  time.Duration `json:"p99"`
	P999                 time.Duration `json:"p999"`
	StdDev               time.Duration `json:"std_dev"`
	Percentiles          map[float64]time.Duration `json:"percentiles"`
}

// ThroughputStats 吞吐量统计
type ThroughputStats struct {
	CurrentQPS           float64 `json:"current_qps"`
	PeakQPS              float64 `json:"peak_qps"`
	AverageQPS           float64 `json:"average_qps"`
	RequestsPerSecond    []float64 `json:"requests_per_second"` // 最近60秒
	RequestsPerMinute    []float64 `json:"requests_per_minute"` // 最近60分钟
}

// LevelMetrics 缓存层级指标
type LevelMetrics struct {
	Level                string        `json:"level"`
	TotalRequests        int64         `json:"total_requests"`
	Hits                 int64         `json:"hits"`
	Misses               int64         `json:"misses"`
	HitRate              float64       `json:"hit_rate"`
	AvgResponseTime      time.Duration `json:"avg_response_time"`
	DataSize             int64         `json:"data_size"`
	EntryCount           int64         `json:"entry_count"`
	Evictions            int64         `json:"evictions"`
	Expirations          int64         `json:"expirations"`
	ErrorRate            float64       `json:"error_rate"`
}

// KeyMetrics 键指标
type KeyMetrics struct {
	Key                  string        `json:"key"`
	AccessCount          int64         `json:"access_count"`
	HitCount             int64         `json:"hit_count"`
	MissCount            int64         `json:"miss_count"`
	HitRate              float64       `json:"hit_rate"`
	LastAccessTime       time.Time     `json:"last_access_time"`
	AvgResponseTime      time.Duration `json:"avg_response_time"`
	DataSize             int64         `json:"data_size"`
	TTL                  time.Duration `json:"ttl"`
	Level                string        `json:"level"`
	AccessFrequency      float64       `json:"access_frequency"`
}

// MemoryMetrics 内存指标
type MemoryMetrics struct {
	TotalMemory          int64         `json:"total_memory"`
	UsedMemory           int64         `json:"used_memory"`
	FreeMemory           int64         `json:"free_memory"`
	MemoryUsagePercent   float64       `json:"memory_usage_percent"`
	FragmentationRatio   float64       `json:"fragmentation_ratio"`
	EvictionMemory       int64         `json:"eviction_memory"`
	OverheadMemory       int64         `json:"overhead_memory"`
	MemoryGrowthRate     float64       `json:"memory_growth_rate"`
}

// NetworkMetrics 网络指标
type NetworkMetrics struct {
	BytesTransferred     int64         `json:"bytes_transferred"`
	ConnectionsCount     int64         `json:"connections_count"`
	ActiveConnections    int64         `json:"active_connections"`
	NetworkLatency       time.Duration `json:"network_latency"`
	ThroughputMBps       float64       `json:"throughput_mbps"`
	ErrorRate            float64       `json:"error_rate"`
	Retries              int64         `json:"retries"`
}

// ErrorMetrics 错误指标
type ErrorMetrics struct {
	TotalErrors          int64         `json:"total_errors"`
	ErrorRate            float64       `json:"error_rate"`
	TimeoutErrors        int64         `json:"timeout_errors"`
	ConnectionErrors     int64         `json:"connection_errors"`
	SerializationErrors  int64         `json:"serialization_errors"`
	ValidationErrors     int64         `json:"validation_errors"`
	OtherErrors          int64         `json:"other_errors"`
	ErrorsByType         map[string]int64 `json:"errors_by_type"`
	RecentErrors         []*ErrorEntry `json:"recent_errors"`
}

// ErrorEntry 错误条目
type ErrorEntry struct {
	Timestamp            time.Time     `json:"timestamp"`
	ErrorType            string        `json:"error_type"`
	ErrorMessage         string        `json:"error_message"`
	Key                  string        `json:"key"`
	Level                string        `json:"level"`
	Duration             time.Duration `json:"duration"`
}

// WindowMetrics 时间窗口指标
type WindowMetrics struct {
	Window               time.Duration `json:"window"`
	StartTime            time.Time     `json:"start_time"`
	EndTime              time.Time     `json:"end_time"`
	TotalRequests        int64         `json:"total_requests"`
	TotalHits            int64         `json:"total_hits"`
	HitRate              float64       `json:"hit_rate"`
	AvgResponseTime      time.Duration `json:"avg_response_time"`
	Throughput           float64       `json:"throughput"`
	ErrorRate            float64       `json:"error_rate"`
}

// TrendMetrics 趋势指标
type TrendMetrics struct {
	HitRateTrend         float64 `json:"hit_rate_trend"`
	ResponseTimeTrend    float64 `json:"response_time_trend"`
	ThroughputTrend      float64 `json:"throughput_trend"`
	ErrorRateTrend       float64 `json:"error_rate_trend"`
	MemoryUsageTrend     float64 `json:"memory_usage_trend"`
	Prediction           *TrendPrediction `json:"prediction"`
}

// TrendPrediction 趋势预测
type TrendPrediction struct {
	HitRate              float64   `json:"hit_rate"`
	ResponseTime         time.Duration `json:"response_time"`
	Throughput           float64   `json:"throughput"`
	ErrorRate            float64   `json:"error_rate"`
	MemoryUsage          int64     `json:"memory_usage"`
	Confidence           float64   `json:"confidence"`
	PredictionTime       time.Time `json:"prediction_time"`
}

// AlertManager 告警管理器
type AlertManager struct {
	mu                   sync.RWMutex
	config               *AlertThresholds
	activeAlerts         map[string]*Alert
	alertHistory         []*Alert
	rateLimiter          *RateLimiter
	logger               *zap.Logger
}

// Alert 告警
type Alert struct {
	ID                   string        `json:"id"`
	Type                 string        `json:"type"`
	Severity             string        `json:"severity"`
	Title                string        `json:"title"`
	Description          string        `json:"description"`
	CurrentValue         interface{}   `json:"current_value"`
	ThresholdValue       interface{}   `json:"threshold_value"`
	Timestamp            time.Time     `json:"timestamp"`
	Duration             time.Duration `json:"duration"`
	Status               string        `json:"status"`
	Metadata             map[string]interface{} `json:"metadata"`
}

// Dashboard 仪表板
type Dashboard struct {
	mu                   sync.RWMutex
	metrics              *DetailedMetrics
	refreshInterval      time.Duration
	logger               *zap.Logger
}

// MetricsExporter 指标导出器接口
type MetricsExporter interface {
	Export(metrics *DetailedMetrics) error
	GetName() string
	IsEnabled() bool
}

// RateLimiter 速率限制器
type RateLimiter struct {
	mu                   sync.Mutex
	tokens               int
	maxTokens            int
	refillRate           int
	lastRefill           time.Time
}

// DefaultMonitorConfig 默认监控配置
func DefaultMonitorConfig() *MonitorConfig {
	return &MonitorConfig{
		CollectionInterval:   30 * time.Second,
		RetentionPeriod:     24 * time.Hour,
		EnableDetailedStats: true,
		EnableRealTimeAlerts: true,
		MetricsEnabled: map[string]bool{
			"hit_rate":        true,
			"response_time":   true,
			"throughput":      true,
			"memory":          true,
			"errors":          true,
			"level_metrics":   true,
			"key_metrics":     true,
		},
		Percentiles: []float64{0.5, 0.75, 0.9, 0.95, 0.99, 0.999},
		HistogramBuckets: []float64{0.001, 0.005, 0.01, 0.025, 0.05, 0.1, 0.25, 0.5, 1, 2.5, 5, 10},
		AlertThresholds: &AlertThresholds{
			ResponseTimeThreshold: 100 * time.Millisecond,
			HitRateThreshold:       0.8,
			QPSThreshold:           10000,
			MemoryThreshold:        1024 * 1024 * 1024, // 1GB
			CPUThreshold:           0.8,
		},
		AlertCooldown:        5 * time.Minute,
		MaxAlertsPerMinute:   10,
		EnablePrometheus:     false,
		EnableInfluxDB:       false,
		EnableJSONExport:     true,
		ExportInterval:       1 * time.Minute,
		ExportPath:          "./metrics",
	}
}

// NewCacheMonitor 创建缓存监控器
func NewCacheMonitor(config *MonitorConfig, logger *zap.Logger) *CacheMonitor {
	if config == nil {
		config = DefaultMonitorConfig()
	}

	monitor := &CacheMonitor{
		config:     config,
		metrics:    &DetailedMetrics{},
		exporters:  make(map[string]MetricsExporter),
		stopCh:     make(chan struct{}),
		logger:     logger.Named("cache_monitor"),
	}

	// 初始化告警管理器
	monitor.alertManager = &AlertManager{
		config:       config.AlertThresholds,
		activeAlerts: make(map[string]*Alert),
		alertHistory: make([]*Alert, 0),
		rateLimiter:  NewRateLimiter(config.MaxAlertsPerMinute, time.Minute),
		logger:       logger.Named("alert_manager"),
	}

	// 初始化仪表板
	monitor.dashboard = &Dashboard{
		refreshInterval: 10 * time.Second,
		logger:          logger.Named("dashboard"),
	}

	// 初始化指标
	monitor.initializeMetrics()

	return monitor
}

// initializeMetrics 初始化指标
func (cm *CacheMonitor) initializeMetrics() {
	cm.metrics = &DetailedMetrics{
		ResponseTimeStats: &ResponseTimeStats{
			Percentiles: make(map[float64]time.Duration),
		},
		ThroughputStats: &ThroughputStats{
			RequestsPerSecond: make([]float64, 60),
			RequestsPerMinute: make([]float64, 60),
		},
		LevelMetrics: make(map[string]*LevelMetrics),
		TopKeys:      make([]*KeyMetrics, 0),
		HotKeys:      make([]*KeyMetrics, 0),
		ColdKeys:     make([]*KeyMetrics, 0),
		MemoryMetrics: &MemoryMetrics{},
		NetworkMetrics: &NetworkMetrics{},
		ErrorMetrics: &ErrorMetrics{
			ErrorsByType: make(map[string]int64),
			RecentErrors: make([]*ErrorEntry, 0),
		},
		TimeWindowMetrics: make(map[string]*WindowMetrics),
		TrendMetrics: &TrendMetrics{
			Prediction: &TrendPrediction{},
		},
		LastUpdateTime: time.Now(),
	}

	// 初始化时间窗口指标
	windows := []time.Duration{time.Minute, 5 * time.Minute, 15 * time.Minute, time.Hour}
	for _, window := range windows {
		cm.metrics.TimeWindowMetrics[window.String()] = &WindowMetrics{
			Window: window,
		}
	}
}

// Start 启动监控
func (cm *CacheMonitor) Start(ctx context.Context) error {
	cm.mu.Lock()
	if cm.isRunning {
		cm.mu.Unlock()
		return fmt.Errorf("monitor is already running")
	}
	cm.isRunning = true
	cm.mu.Unlock()

	cm.logger.Info("Starting cache monitor",
		zap.Duration("collection_interval", cm.config.CollectionInterval),
		zap.Bool("detailed_stats", cm.config.EnableDetailedStats),
	)

	// 启动数据收集
	go cm.startCollection(ctx)

	// 启动导出器
	if cm.config.EnableJSONExport {
		go cm.startExporting(ctx)
	}

	return nil
}

// Stop 停止监控
func (cm *CacheMonitor) Stop() {
	cm.mu.Lock()
	defer cm.mu.Unlock()

	if cm.isRunning {
		cm.isRunning = false
		close(cm.stopCh)
		cm.stopCh = make(chan struct{})
		cm.logger.Info("Cache monitor stopped")
	}
}

// startCollection 启动数据收集
func (cm *CacheMonitor) startCollection(ctx context.Context) {
	ticker := time.NewTicker(cm.config.CollectionInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-cm.stopCh:
			return
		case <-ticker.C:
			cm.collectMetrics(ctx)
		}
	}
}

// collectMetrics 收集指标
func (cm *CacheMonitor) collectMetrics(ctx context.Context) {
	cm.mu.Lock()
	defer cm.mu.Unlock()

	// 更新基础指标
	cm.updateBasicMetrics()

	// 更新响应时间指标
	cm.updateResponseTimeMetrics()

	// 更新吞吐量指标
	cm.updateThroughputMetrics()

	// 更新时间窗口指标
	cm.updateTimeWindowMetrics()

	// 更新趋势指标
	cm.updateTrendMetrics()

	// 检查告警
	if cm.config.EnableRealTimeAlerts {
		go cm.checkAlerts()
	}

	cm.metrics.LastUpdateTime = time.Now()
}

// updateBasicMetrics 更新基础指标
func (cm *CacheMonitor) updateBasicMetrics() {
	// 这里需要从实际的缓存管理器获取数据
	// 简化实现，使用模拟数据
	totalRequests := atomic.LoadInt64(&cm.metrics.TotalRequests)
	totalHits := atomic.LoadInt64(&cm.metrics.TotalHits)

	if totalRequests > 0 {
		cm.metrics.HitRate = float64(totalHits) / float64(totalRequests)
	}
}

// updateResponseTimeMetrics 更新响应时间指标
func (cm *CacheMonitor) updateResponseTimeMetrics() {
	// 模拟响应时间数据
	responseTimes := []time.Duration{
		10 * time.Millisecond,
		15 * time.Millisecond,
		20 * time.Millisecond,
		25 * time.Millisecond,
		30 * time.Millisecond,
		50 * time.Millisecond,
		100 * time.Millisecond,
		200 * time.Millisecond,
	}

	if len(responseTimes) == 0 {
		return
	}

	// 计算统计值
	sort.Slice(responseTimes, func(i, j int) bool {
		return responseTimes[i] < responseTimes[j]
	})

	cm.metrics.ResponseTimeStats.Min = responseTimes[0]
	cm.metrics.ResponseTimeStats.Max = responseTimes[len(responseTimes)-1]

	// 计算平均值
	var sum time.Duration
	for _, rt := range responseTimes {
		sum += rt
	}
	cm.metrics.ResponseTimeStats.Mean = sum / time.Duration(len(responseTimes))

	// 计算中位数
	n := len(responseTimes)
	if n%2 == 0 {
		cm.metrics.ResponseTimeStats.Median = (responseTimes[n/2-1] + responseTimes[n/2]) / 2
	} else {
		cm.metrics.ResponseTimeStats.Median = responseTimes[n/2]
	}

	// 计算百分位数
	for _, p := range cm.config.Percentiles {
		index := int(float64(n) * p)
		if index >= n {
			index = n - 1
		}
		cm.metrics.ResponseTimeStats.Percentiles[p] = responseTimes[index]
	}

	// 设置常用百分位数
	cm.metrics.ResponseTimeStats.P50 = cm.metrics.ResponseTimeStats.Percentiles[0.5]
	cm.metrics.ResponseTimeStats.P75 = cm.metrics.ResponseTimeStats.Percentiles[0.75]
	cm.metrics.ResponseTimeStats.P90 = cm.metrics.ResponseTimeStats.Percentiles[0.9]
	cm.metrics.ResponseTimeStats.P95 = cm.metrics.ResponseTimeStats.Percentiles[0.95]
	cm.metrics.ResponseTimeStats.P99 = cm.metrics.ResponseTimeStats.Percentiles[0.99]
	cm.metrics.ResponseTimeStats.P999 = cm.metrics.ResponseTimeStats.Percentiles[0.999]
}

// updateThroughputMetrics 更新吞吐量指标
func (cm *CacheMonitor) updateThroughputMetrics() {
	// 模拟吞吐量数据
	now := time.Now()
	currentQPS := 1500.0 + 500.0*math.Sin(float64(now.Unix())/10.0) // 模拟波动

	cm.metrics.ThroughputStats.CurrentQPS = currentQPS

	if currentQPS > cm.metrics.ThroughputStats.PeakQPS {
		cm.metrics.ThroughputStats.PeakQPS = currentQPS
	}

	// 更新每秒请求数（滚动窗口）
	copy(cm.metrics.ThroughputStats.RequestsPerSecond[1:], cm.metrics.ThroughputStats.RequestsPerSecond[:59])
	cm.metrics.ThroughputStats.RequestsPerSecond[59] = currentQPS

	// 计算平均QPS
	var sum float64
	count := 0
	for _, qps := range cm.metrics.ThroughputStats.RequestsPerSecond {
		if qps > 0 {
			sum += qps
			count++
		}
	}
	if count > 0 {
		cm.metrics.ThroughputStats.AverageQPS = sum / float64(count)
	}
}

// updateTimeWindowMetrics 更新时间窗口指标
func (cm *CacheMonitor) updateTimeWindowMetrics() {
	for _, window := range cm.metrics.TimeWindowMetrics {
		// 模拟时间窗口数据
		window.StartTime = time.Now().Add(-window.Window)
		window.EndTime = time.Now()

		// 根据窗口大小调整指标
		multiplier := window.Window.Seconds() / 60.0
		window.TotalRequests = int64(float64(cm.metrics.ThroughputStats.CurrentQPS) * multiplier * 60)
		window.TotalHits = int64(float64(window.TotalRequests) * cm.metrics.HitRate)
		window.HitRate = cm.metrics.HitRate
		window.AvgResponseTime = cm.metrics.ResponseTimeStats.Mean
		window.Throughput = cm.metrics.ThroughputStats.CurrentQPS
		window.ErrorRate = cm.metrics.ErrorMetrics.ErrorRate
	}
}

// updateTrendMetrics 更新趋势指标
func (cm *CacheMonitor) updateTrendMetrics() {
	// 模拟趋势计算
	cm.metrics.TrendMetrics.HitRateTrend = 0.05           // 5% 增长趋势
	cm.metrics.TrendMetrics.ResponseTimeTrend = -0.02      // 2% 改善趋势
	cm.metrics.TrendMetrics.ThroughputTrend = 0.10         // 10% 增长趋势
	cm.metrics.TrendMetrics.ErrorRateTrend = -0.01         // 1% 减少趋势
	cm.metrics.TrendMetrics.MemoryUsageTrend = 0.03        // 3% 增长趋势

	// 简单的预测
	cm.metrics.TrendMetrics.Prediction = &TrendPrediction{
		HitRate:        math.Min(1.0, cm.metrics.HitRate*(1+cm.metrics.TrendMetrics.HitRateTrend)),
		ResponseTime:   time.Duration(float64(cm.metrics.ResponseTimeStats.Mean) * (1 + cm.metrics.TrendMetrics.ResponseTimeTrend)),
		Throughput:     cm.metrics.ThroughputStats.CurrentQPS * (1 + cm.metrics.TrendMetrics.ThroughputTrend),
		ErrorRate:      math.Max(0, cm.metrics.ErrorMetrics.ErrorRate*(1+cm.metrics.TrendMetrics.ErrorRateTrend)),
		MemoryUsage:    int64(float64(cm.metrics.MemoryMetrics.UsedMemory) * (1 + cm.metrics.TrendMetrics.MemoryUsageTrend)),
		Confidence:     0.75,
		PredictionTime: time.Now().Add(time.Hour),
	}
}

// RecordRequest 记录请求
func (cm *CacheMonitor) RecordRequest(key string, hit bool, responseTime time.Duration, level string) {
	atomic.AddInt64(&cm.metrics.TotalRequests, 1)

	if hit {
		atomic.AddInt64(&cm.metrics.TotalHits, 1)
	} else {
		atomic.AddInt64(&cm.metrics.TotalMisses, 1)
	}

	// 更新键级别指标
	cm.updateKeyMetrics(key, hit, responseTime, level)

	// 更新层级指标
	cm.updateLevelMetrics(level, hit, responseTime)
}

// RecordError 记录错误
func (cm *CacheMonitor) RecordError(errorType, message, key, level string, duration time.Duration) {
	atomic.AddInt64(&cm.metrics.ErrorMetrics.TotalErrors, 1)
	cm.metrics.ErrorMetrics.ErrorsByType[errorType]++

	// 添加到最近错误列表
	errorEntry := &ErrorEntry{
		Timestamp:    time.Now(),
		ErrorType:    errorType,
		ErrorMessage: message,
		Key:          key,
		Level:        level,
		Duration:     duration,
	}

	cm.mu.Lock()
	cm.metrics.ErrorMetrics.RecentErrors = append(cm.metrics.ErrorMetrics.RecentErrors, errorEntry)

	// 保持最近100个错误
	if len(cm.metrics.ErrorMetrics.RecentErrors) > 100 {
		cm.metrics.ErrorMetrics.RecentErrors = cm.metrics.ErrorMetrics.RecentErrors[1:]
	}
	cm.mu.Unlock()
}

// startExporting 启动指标导出
func (cm *CacheMonitor) startExporting(ctx context.Context) {
	ticker := time.NewTicker(cm.config.ExportInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-cm.stopCh:
			return
		case <-ticker.C:
			cm.exportMetrics(ctx)
		}
	}
}

// exportMetrics 导出指标
func (cm *CacheMonitor) exportMetrics(ctx context.Context) {
	// 简化实现，记录导出操作
	cm.logger.Debug("Exporting metrics", zap.Time("time", time.Now()))
}

// checkAlerts 检查告警
func (cm *CacheMonitor) checkAlerts() {
	cm.mu.RLock()
	defer cm.mu.RUnlock()

	if cm.alertManager == nil {
		return
	}

	// 简化的告警检查
	if cm.metrics.HitRate < 0.5 {
		cm.logger.Warn("Low hit rate detected", zap.Float64("hit_rate", cm.metrics.HitRate))
	}

	if cm.metrics.ResponseTimeStats != nil && cm.metrics.ResponseTimeStats.Mean > 100*time.Millisecond {
		cm.logger.Warn("High response time detected", zap.Duration("response_time", cm.metrics.ResponseTimeStats.Mean))
	}
}

// updateKeyMetrics 更新键指标
func (cm *CacheMonitor) updateKeyMetrics(key string, hit bool, responseTime time.Duration, level string) {
	cm.mu.Lock()
	defer cm.mu.Unlock()

	// 查找或创建键指标
	var keyMetrics *KeyMetrics
	for _, km := range cm.metrics.TopKeys {
		if km.Key == key {
			keyMetrics = km
			break
		}
	}

	if keyMetrics == nil {
		keyMetrics = &KeyMetrics{
			Key:             key,
			Level:           level,
			AccessFrequency: 0,
		}
		cm.metrics.TopKeys = append(cm.metrics.TopKeys, keyMetrics)
	}

	// 更新指标
	keyMetrics.AccessCount++
	if hit {
		keyMetrics.HitCount++
	} else {
		keyMetrics.MissCount++
	}

	if keyMetrics.AccessCount > 0 {
		keyMetrics.HitRate = float64(keyMetrics.HitCount) / float64(keyMetrics.AccessCount)
	}

	keyMetrics.LastAccessTime = time.Now()
	keyMetrics.Level = level

	// 更新平均响应时间
	if keyMetrics.AccessCount == 1 {
		keyMetrics.AvgResponseTime = responseTime
	} else {
		keyMetrics.AvgResponseTime = time.Duration(
			(float64(keyMetrics.AvgResponseTime)*float64(keyMetrics.AccessCount-1) + float64(responseTime)) / float64(keyMetrics.AccessCount),
		)
	}

	// 计算访问频率（每分钟访问次数）
	now := time.Now()
	if !keyMetrics.LastAccessTime.IsZero() {
		duration := now.Sub(keyMetrics.LastAccessTime)
		if duration > 0 {
			keyMetrics.AccessFrequency = float64(time.Minute) / duration.Seconds()
		}
	}

	// 保持Top 100键
	if len(cm.metrics.TopKeys) > 100 {
		sort.Slice(cm.metrics.TopKeys, func(i, j int) bool {
			return cm.metrics.TopKeys[i].AccessCount > cm.metrics.TopKeys[j].AccessCount
		})
		cm.metrics.TopKeys = cm.metrics.TopKeys[:100]
	}
}

// updateLevelMetrics 更新层级指标
func (cm *CacheMonitor) updateLevelMetrics(level string, hit bool, responseTime time.Duration) {
	cm.mu.Lock()
	defer cm.mu.Unlock()

	levelMetrics, exists := cm.metrics.LevelMetrics[level]
	if !exists {
		levelMetrics = &LevelMetrics{
			Level: level,
		}
		cm.metrics.LevelMetrics[level] = levelMetrics
	}

	levelMetrics.TotalRequests++
	if hit {
		levelMetrics.Hits++
	} else {
		levelMetrics.Misses++
	}

	if levelMetrics.TotalRequests > 0 {
		levelMetrics.HitRate = float64(levelMetrics.Hits) / float64(levelMetrics.TotalRequests)
	}

	// 更新平均响应时间
	if levelMetrics.TotalRequests == 1 {
		levelMetrics.AvgResponseTime = responseTime
	} else {
		levelMetrics.AvgResponseTime = time.Duration(
			(float64(levelMetrics.AvgResponseTime)*float64(levelMetrics.TotalRequests-1) + float64(responseTime)) / float64(levelMetrics.TotalRequests),
		)
	}
}

// GetMetrics 获取当前指标
func (cm *CacheMonitor) GetMetrics() *DetailedMetrics {
	cm.mu.RLock()
	defer cm.mu.RUnlock()

	// 返回指标的深拷贝
	return cm.deepCopyMetrics(cm.metrics)
}

// deepCopyMetrics 深拷贝指标
func (cm *CacheMonitor) deepCopyMetrics(original *DetailedMetrics) *DetailedMetrics {
	copied := &DetailedMetrics{
		TotalRequests:  original.TotalRequests,
		TotalHits:      original.TotalHits,
		TotalMisses:    original.TotalMisses,
		HitRate:        original.HitRate,
		LastUpdateTime: original.LastUpdateTime,
	}

	// 拷贝响应时间统计
	if original.ResponseTimeStats != nil {
		copied.ResponseTimeStats = &ResponseTimeStats{
			Min:         original.ResponseTimeStats.Min,
			Max:         original.ResponseTimeStats.Max,
			Mean:        original.ResponseTimeStats.Mean,
			Median:      original.ResponseTimeStats.Median,
			P50:         original.ResponseTimeStats.P50,
			P75:         original.ResponseTimeStats.P75,
			P90:         original.ResponseTimeStats.P90,
			P95:         original.ResponseTimeStats.P95,
			P99:         original.ResponseTimeStats.P99,
			P999:        original.ResponseTimeStats.P999,
			StdDev:      original.ResponseTimeStats.StdDev,
			Percentiles: make(map[float64]time.Duration),
		}
		for k, v := range original.ResponseTimeStats.Percentiles {
			copied.ResponseTimeStats.Percentiles[k] = v
		}
	}

	// 拷贝吞吐量统计
	if original.ThroughputStats != nil {
		copied.ThroughputStats = &ThroughputStats{
			CurrentQPS:        original.ThroughputStats.CurrentQPS,
			PeakQPS:           original.ThroughputStats.PeakQPS,
			AverageQPS:        original.ThroughputStats.AverageQPS,
			RequestsPerSecond: make([]float64, len(original.ThroughputStats.RequestsPerSecond)),
			RequestsPerMinute: make([]float64, len(original.ThroughputStats.RequestsPerMinute)),
		}
		copy(copied.ThroughputStats.RequestsPerSecond, original.ThroughputStats.RequestsPerSecond)
		copy(copied.ThroughputStats.RequestsPerMinute, original.ThroughputStats.RequestsPerMinute)
	}

	return copied
}

// NewRateLimiter 创建速率限制器
func NewRateLimiter(maxTokens int, refillInterval time.Duration) *RateLimiter {
	return &RateLimiter{
		maxTokens:  maxTokens,
		tokens:     maxTokens,
		refillRate: maxTokens,
		lastRefill: time.Now(),
	}
}

// Allow 检查是否允许
func (rl *RateLimiter) Allow() bool {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	now := time.Now()
	elapsed := now.Sub(rl.lastRefill)

	// 补充令牌 (假设每秒补充refillRate个令牌)
	tokensToAdd := int(float64(elapsed.Seconds()) * float64(rl.refillRate))
	rl.tokens = min(rl.maxTokens, rl.tokens+tokensToAdd)
	rl.lastRefill = now

	if rl.tokens > 0 {
		rl.tokens--
		return true
	}

	return false
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}