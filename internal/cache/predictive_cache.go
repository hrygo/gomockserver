package cache

import (
	"context"
	"fmt"
	"math"
	"sync"
	"time"

	"go.uber.org/zap"
)

// PredictiveCache 预测性缓存
type PredictiveCache struct {
	mu                sync.RWMutex
	accessPatterns    map[string]*AccessPattern
	predictionEngine  *PredictionEngine
	cacheManager      Manager
	predictionConfig  *PredictionConfig
	logger            *zap.Logger
}

// AccessPattern 访问模式
type AccessPattern struct {
	Key               string    `json:"key"`
	AccessTimes       []time.Time `json:"access_times"`
	Periodicity       float64   `json:"periodicity"`       // 周期性（秒）
	Predictability    float64   `json:"predictability"`   // 可预测性
	LastPredicted     time.Time `json:"last_predicted"`
	PredictionAccuracy float64   `json:"prediction_accuracy"`
	SeasonalFactor    float64   `json:"seasonal_factor"`   // 季节性因子
}

// PredictionEngine 预测引擎
type PredictionEngine struct {
	mu                sync.RWMutex
	models            map[string]*PredictionModel
	predictionHistory map[string][]Prediction
	learningRate       float64
	logger            *zap.Logger
}

// PredictionModel 预测模型
type PredictionModel struct {
	Key               string            `json:"key"`
	ModelType         string            `json:"model_type"`     // linear, seasonal, trending
	Parameters        map[string]float64 `json:"parameters"`
	LastTraining      time.Time         `json:"last_training"`
	TrainingSamples   int               `json:"training_samples"`
	Accuracy          float64           `json:"accuracy"`
}

// Prediction 预测结果
type Prediction struct {
	Key               string    `json:"key"`
	PredictedTime     time.Time `json:"predicted_time"`
	Confidence        float64   `json:"confidence"`
	Reason            string    `json:"reason"`
	PreloadRecommended bool      `json:"preload_recommended"`
}

// PredictionConfig 预测配置
type PredictionConfig struct {
	EnablePreload        bool          `json:"enable_preload"`
	PreloadWindow        time.Duration `json:"preload_window"`        // 预测窗口
	MinAccessCount       int           `json:"min_access_count"`      // 最小访问次数
	MinPredictability    float64       `json:"min_predictability"`    // 最小可预测性
	TrainingInterval     time.Duration `json:"training_interval"`     // 训练间隔
	MaxPredictionTime    time.Duration `json:"max_prediction_time"`    // 最大预测时间
	PreloadConcurrency   int           `json:"preload_concurrency"`   // 预加载并发数
	PatternHistoryLimit  int           `json:"pattern_history_limit"` // 模式历史限制
}

// DefaultPredictionConfig 默认预测配置
func DefaultPredictionConfig() *PredictionConfig {
	return &PredictionConfig{
		EnablePreload:       true,
		PreloadWindow:       1 * time.Hour,
		MinAccessCount:      5,
		MinPredictability:   0.7,
		TrainingInterval:    10 * time.Minute,
		MaxPredictionTime:   6 * time.Hour,
		PreloadConcurrency:  5,
		PatternHistoryLimit: 1000,
	}
}

// NewPredictiveCache 创建预测性缓存
func NewPredictiveCache(cacheManager Manager, predictionConfig *PredictionConfig, logger *zap.Logger) *PredictiveCache {
	if predictionConfig == nil {
		predictionConfig = DefaultPredictionConfig()
	}

	pc := &PredictiveCache{
		accessPatterns:   make(map[string]*AccessPattern),
		cacheManager:     cacheManager,
		predictionConfig: predictionConfig,
		logger:           logger.Named("predictive_cache"),
	}

	// 创建预测引擎
	pc.predictionEngine = &PredictionEngine{
		models:            make(map[string]*PredictionModel),
		predictionHistory: make(map[string][]Prediction),
		learningRate:       0.01,
		logger:           logger.Named("prediction_engine"),
	}

	// 启动后台任务
	go pc.startBackgroundTasks()

	logger.Info("Predictive cache initialized",
		zap.Bool("enable_preload", predictionConfig.EnablePreload),
		zap.Duration("preload_window", predictionConfig.PreloadWindow),
		zap.Float64("min_predictability", predictionConfig.MinPredictability),
	)

	return pc
}

// RecordAccess 记录访问并更新预测模型
func (pc *PredictiveCache) RecordAccess(key string) {
	pc.mu.Lock()
	defer pc.mu.Unlock()

	// 获取或创建访问模式
	pattern, exists := pc.accessPatterns[key]
	if !exists {
		pattern = &AccessPattern{
			Key:            key,
			AccessTimes:    make([]time.Time, 0),
			SeasonalFactor: 1.0,
		}
		pc.accessPatterns[key] = pattern
	}

	// 记录访问时间
	now := time.Now()
	pattern.AccessTimes = append(pattern.AccessTimes, now)

	// 限制历史记录数量
	if len(pattern.AccessTimes) > pc.predictionConfig.PatternHistoryLimit {
		pattern.AccessTimes = pattern.AccessTimes[1:]
	}

	// 更新模式特征
	pc.updatePatternFeatures(pattern)

	// 如果有足够的数据，训练预测模型
	if len(pattern.AccessTimes) >= pc.predictionConfig.MinAccessCount {
		pc.trainPredictionModel(key, pattern)
	}
}

// updatePatternFeatures 更新访问模式特征
func (pc *PredictiveCache) updatePatternFeatures(pattern *AccessPattern) {
	if len(pattern.AccessTimes) < 2 {
		return
	}

	// 计算访问间隔
	intervals := make([]float64, len(pattern.AccessTimes)-1)
	for i := 1; i < len(pattern.AccessTimes); i++ {
		interval := pattern.AccessTimes[i].Sub(pattern.AccessTimes[i-1]).Seconds()
		intervals[i-1] = interval
	}

	// 计算周期性（简化实现）
	pattern.Periodicity = pc.calculatePeriodicity(intervals)

	// 计算可预测性
	pattern.Predictability = pc.calculatePredictability(intervals)

	// 计算季节性因子
	pattern.SeasonalFactor = pc.calculateSeasonalFactor(pattern.AccessTimes)
}

// calculatePeriodicity 计算周期性
func (pc *PredictiveCache) calculatePeriodicity(intervals []float64) float64 {
	if len(intervals) < 3 {
		return 0.0
	}

	// 计算平均间隔
	var sum float64
	for _, interval := range intervals {
		sum += interval
	}
	mean := sum / float64(len(intervals))

	// 计算方差
	var variance float64
	for _, interval := range intervals {
		diff := interval - mean
		variance += diff * diff
	}
	variance /= float64(len(intervals))

	// 方差越小，周期性越强
	if variance == 0 {
		return 1.0
	}

	// 使用变异系数的倒数作为周期性指标
	cv := math.Sqrt(variance) / mean
	return 1.0 / (1.0 + cv)
}

// calculatePredictability 计算可预测性
func (pc *PredictiveCache) calculatePredictability(intervals []float64) float64 {
	if len(intervals) < 3 {
		return 0.0
	}

	// 简单的可预测性计算：基于间隔的一致性
	mean := 0.0
	for _, interval := range intervals {
		mean += interval
	}
	mean /= float64(len(intervals))

	errors := 0.0
	for _, interval := range intervals {
		error := math.Abs(interval-mean) / mean
		errors += error
	}

	avgError := errors / float64(len(intervals))
	predictability := math.Max(0, 1.0-avgError)

	return predictability
}

// calculateSeasonalFactor 计算季节性因子
func (pc *PredictiveCache) calculateSeasonalFactor(accessTimes []time.Time) float64 {
	if len(accessTimes) < 24 { // 至少需要24个访问点
		return 1.0
	}

	// 按小时分组访问次数
	hourlyCounts := make(map[int]int)
	for _, accessTime := range accessTimes {
		hour := accessTime.Hour()
		hourlyCounts[hour]++
	}

	// 计算方差
	var mean float64
	for _, count := range hourlyCounts {
		mean += float64(count)
	}
	mean /= float64(len(hourlyCounts))

	var variance float64
	for _, count := range hourlyCounts {
		diff := float64(count) - mean
		variance += diff * diff
	}
	variance /= float64(len(hourlyCounts))

	// 季节性因子：方差越大，季节性越强
	if mean == 0 {
		return 1.0
	}

	cv := math.Sqrt(variance) / mean
	return math.Min(3.0, 1.0+cv) // 限制最大值
}

// trainPredictionModel 训练预测模型
func (pc *PredictiveCache) trainPredictionModel(key string, pattern *AccessPattern) {
	model := &PredictionModel{
		Key:             key,
		ModelType:       pc.determineModelType(pattern),
		Parameters:      make(map[string]float64),
		LastTraining:    time.Now(),
		TrainingSamples: len(pattern.AccessTimes),
	}

	// 根据模型类型训练参数
	switch model.ModelType {
	case "linear":
		pc.trainLinearModel(model, pattern)
	case "seasonal":
		pc.trainSeasonalModel(model, pattern)
	case "trending":
		pc.trainTrendingModel(model, pattern)
	}

	// 评估模型准确性
	model.Accuracy = pc.evaluateModelAccuracy(model, pattern)

	// 更新预测引擎
	pc.predictionEngine.mu.Lock()
	pc.predictionEngine.models[key] = model
	pc.predictionEngine.mu.Unlock()

	pc.logger.Debug("Prediction model trained",
		zap.String("key", key),
		zap.String("model_type", model.ModelType),
		zap.Float64("accuracy", model.Accuracy),
	)
}

// determineModelType 确定模型类型
func (pc *PredictiveCache) determineModelType(pattern *AccessPattern) string {
	if pattern.SeasonalFactor > 2.0 {
		return "seasonal"
	}

	// 简单的趋势检测
	if len(pattern.AccessTimes) >= 10 {
		firstAccess := pattern.AccessTimes[0]
		lastAccess := pattern.AccessTimes[len(pattern.AccessTimes)-1]
		timeSpan := lastAccess.Sub(firstAccess).Hours()
		accessRate := float64(len(pattern.AccessTimes)) / timeSpan

		if accessRate > 1.0 { // 每小时超过1次访问
			return "trending"
		}
	}

	return "linear"
}

// trainLinearModel 训练线性模型
func (pc *PredictiveCache) trainLinearModel(model *PredictionModel, pattern *AccessPattern) {
	if len(pattern.AccessTimes) < 2 {
		return
	}

	// 简单的线性回归：基于前一个访问时间预测下一个
	n := float64(len(pattern.AccessTimes))
	var sumX, sumY, sumXY, sumX2 float64

	for i, accessTime := range pattern.AccessTimes {
		x := float64(i)
		y := float64(accessTime.Unix())
		sumX += x
		sumY += y
		sumXY += x * y
		sumX2 += x * x
	}

	// 计算线性回归参数
	denominator := n*sumX2 - sumX*sumX
	if denominator != 0 {
		slope := (n*sumXY - sumX*sumY) / denominator
		intercept := (sumY - slope*sumX) / n

		model.Parameters["slope"] = slope
		model.Parameters["intercept"] = intercept
	}
}

// trainSeasonalModel 训练季节性模型
func (pc *PredictiveCache) trainSeasonalModel(model *PredictionModel, pattern *AccessPattern) {
	// 分析访问时间的小时分布
	hourlyCounts := make(map[int]int)
	for _, accessTime := range pattern.AccessTimes {
		hour := accessTime.Hour()
		hourlyCounts[hour]++
	}

	// 找出最活跃的小时
	maxCount := 0
	peakHour := 0
	for hour, count := range hourlyCounts {
		if count > maxCount {
			maxCount = count
			peakHour = hour
		}
	}

	model.Parameters["peak_hour"] = float64(peakHour)
	model.Parameters["peak_count"] = float64(maxCount)
	model.Parameters["seasonal_factor"] = pattern.SeasonalFactor
}

// trainTrendingModel 训练趋势模型
func (pc *PredictiveCache) trainTrendingModel(model *PredictionModel, pattern *AccessPattern) {
	if len(pattern.AccessTimes) < 2 {
		return
	}

	// 计算访问频率趋势
	firstAccess := pattern.AccessTimes[0]
	lastAccess := pattern.AccessTimes[len(pattern.AccessTimes)-1]
	timeSpan := lastAccess.Sub(firstAccess).Seconds()
	accessRate := float64(len(pattern.AccessTimes)) / timeSpan

	model.Parameters["access_rate"] = accessRate
	model.Parameters["time_span"] = timeSpan
}

// evaluateModelAccuracy 评估模型准确性
func (pc *PredictiveCache) evaluateModelAccuracy(model *PredictionModel, pattern *AccessPattern) float64 {
	if len(pattern.AccessTimes) < 3 {
		return 0.0
	}

	// 使用最后几个访问点来验证模型
	validationSize := minInt(5, len(pattern.AccessTimes)/2)
	if validationSize < 1 {
		return 0.0
	}

	var totalError float64
	validPredictions := 0

	for i := len(pattern.AccessTimes) - validationSize; i < len(pattern.AccessTimes); i++ {
		actualTime := pattern.AccessTimes[i]

		// 使用历史数据预测
		predictedTime := pc.predictWithModel(model, i, pattern.AccessTimes[:i])
		if !predictedTime.IsZero() {
			error := math.Abs(float64(predictedTime.Sub(actualTime).Seconds()))
			totalError += error
			validPredictions++
		}
	}

	if validPredictions == 0 {
		return 0.0
	}

	avgError := totalError / float64(validPredictions)
	accuracy := math.Max(0, 1.0-avgError/3600.0) // 假设1小时的误差为完全不准确

	return accuracy
}

// predictWithModel 使用模型进行预测
func (pc *PredictiveCache) predictWithModel(model *PredictionModel, index int, history []time.Time) time.Time {
	switch model.ModelType {
	case "linear":
		return pc.predictWithLinearModel(model, index)
	case "seasonal":
		return pc.predictWithSeasonalModel(model, history)
	case "trending":
		return pc.predictWithTrendingModel(model, history)
	}
	return time.Time{}
}

// predictWithLinearModel 线性模型预测
func (pc *PredictiveCache) predictWithLinearModel(model *PredictionModel, index int) time.Time {
	slope, ok1 := model.Parameters["slope"]
	intercept, ok2 := model.Parameters["intercept"]
	if !ok1 || !ok2 {
		return time.Time{}
	}

	x := float64(index + 1)
	y := intercept + slope*x
	return time.Unix(int64(y), 0)
}

// predictWithSeasonalModel 季节性模型预测
func (pc *PredictiveCache) predictWithSeasonalModel(model *PredictionModel, history []time.Time) time.Time {
	peakHour, ok := model.Parameters["peak_hour"]
	if !ok {
		return time.Time{}
	}

	// 预测下一个在峰值小时
	now := time.Now()
	nextPeak := time.Date(now.Year(), now.Month(), now.Day(), int(peakHour), 0, 0, 0, now.Location())

	// 如果已过今天的峰值，预测明天的
	if nextPeak.Before(now) {
		nextPeak = nextPeak.Add(24 * time.Hour)
	}

	return nextPeak
}

// predictWithTrendingModel 趋势模型预测
func (pc *PredictiveCache) predictWithTrendingModel(model *PredictionModel, history []time.Time) time.Time {
	accessRate, ok := model.Parameters["access_rate"]
	if !ok || accessRate <= 0 {
		return time.Time{}
	}

	if len(history) == 0 {
		return time.Time{}
	}

	// 基于平均访问频率预测下一次访问
	nextInterval := 1.0 / accessRate
	return history[len(history)-1].Add(time.Duration(nextInterval * float64(time.Second)))
}

// PredictNextAccess 预测下次访问时间
func (pc *PredictiveCache) PredictNextAccess(ctx context.Context, key string) (*Prediction, error) {
	pc.mu.RLock()
	pattern, exists := pc.accessPatterns[key]
	pc.mu.RUnlock()

	if !exists || len(pattern.AccessTimes) < pc.predictionConfig.MinAccessCount {
		return nil, fmt.Errorf("insufficient data for prediction")
	}

	// 检查可预测性
	if pattern.Predictability < pc.predictionConfig.MinPredictability {
		return nil, fmt.Errorf("pattern not predictable enough")
	}

	// 获取预测模型
	pc.predictionEngine.mu.RLock()
	model, modelExists := pc.predictionEngine.models[key]
	pc.predictionEngine.mu.RUnlock()

	if !modelExists || model.Accuracy < 0.5 {
		return nil, fmt.Errorf("no reliable prediction model")
	}

	// 进行预测
	lastIndex := len(pattern.AccessTimes) - 1
	predictedTime := pc.predictWithModel(model, lastIndex, pattern.AccessTimes)

	if predictedTime.IsZero() {
		return nil, fmt.Errorf("prediction failed")
	}

	// 检查预测时间是否在合理范围内
	now := time.Now()
	maxPredictionTime := now.Add(pc.predictionConfig.MaxPredictionTime)
	if predictedTime.After(maxPredictionTime) {
		return nil, fmt.Errorf("prediction too far in the future")
	}

	// 计算置信度
	confidence := pc.calculateConfidence(model, pattern, predictedTime)

	// 确定是否建议预加载
	preloadRecommended := confidence > 0.8 && predictedTime.Sub(now) < pc.predictionConfig.PreloadWindow

	prediction := &Prediction{
		Key:                key,
		PredictedTime:      predictedTime,
		Confidence:         confidence,
		Reason:             fmt.Sprintf("Model: %s, Accuracy: %.2f", model.ModelType, model.Accuracy),
		PreloadRecommended: preloadRecommended,
	}

	// 更新最后预测时间
	pc.mu.Lock()
	pattern.LastPredicted = now
	pc.mu.Unlock()

	// 记录预测历史
	pc.predictionEngine.mu.Lock()
	pc.predictionEngine.predictionHistory[key] = append(pc.predictionEngine.predictionHistory[key], *prediction)
	pc.predictionEngine.mu.Unlock()

	return prediction, nil
}

// calculateConfidence 计算预测置信度
func (pc *PredictiveCache) calculateConfidence(model *PredictionModel, pattern *AccessPattern, predictedTime time.Time) float64 {
	// 基于模型准确性
	accuracy := model.Accuracy

	// 基于可预测性
	predictability := pattern.Predictability

	// 基于预测时间的合理性
	now := time.Now()
	timeToPrediction := predictedTime.Sub(now)
	timeScore := 1.0
	if timeToPrediction > pc.predictionConfig.MaxPredictionTime {
		timeScore = 0.0
	} else {
		timeScore = 1.0 - float64(timeToPrediction)/float64(pc.predictionConfig.MaxPredictionTime)
	}

	// 综合置信度
	confidence := (accuracy + predictability + timeScore) / 3.0
	return math.Min(1.0, confidence)
}

// PreloadIfNeeded 根据预测结果预加载缓存
func (pc *PredictiveCache) PreloadIfNeeded(ctx context.Context, predictions []*Prediction) error {
	if !pc.predictionConfig.EnablePreload {
		return nil
	}

	// 过滤需要预加载的预测
	var preloadPredictions []*Prediction
	for _, prediction := range predictions {
		if prediction.PreloadRecommended && prediction.Confidence > 0.8 {
			preloadPredictions = append(preloadPredictions, prediction)
		}
	}

	if len(preloadPredictions) == 0 {
		return nil
	}

	pc.logger.Info("Starting predictive preload",
		zap.Int("predictions", len(preloadPredictions)),
	)

	// 并发预加载
	semaphore := make(chan struct{}, pc.predictionConfig.PreloadConcurrency)
	var wg sync.WaitGroup
	var preloadErrors []error

	for _, prediction := range preloadPredictions {
		wg.Add(1)
		go func(p *Prediction) {
			defer wg.Done()
			semaphore <- struct{}{}
			defer func() { <-semaphore }()

			// 这里应该调用实际的预加载逻辑
			// 例如从数据源加载数据并存入缓存
			pc.logger.Debug("Preloading predicted key",
				zap.String("key", p.Key),
				zap.Time("predicted_time", p.PredictedTime),
				zap.Float64("confidence", p.Confidence),
			)
		}(prediction)
	}

	wg.Wait()

	if len(preloadErrors) > 0 {
		return fmt.Errorf("preload errors: %v", preloadErrors)
	}

	pc.logger.Info("Predictive preload completed",
		zap.Int("preloaded_keys", len(preloadPredictions)),
	)

	return nil
}

// startBackgroundTasks 启动后台任务
func (pc *PredictiveCache) startBackgroundTasks() {
	// 定期训练模型
	trainingTicker := time.NewTicker(pc.predictionConfig.TrainingInterval)
	defer trainingTicker.Stop()

	// 定期清理过期的访问模式
	cleanupTicker := time.NewTicker(1 * time.Hour)
	defer cleanupTicker.Stop()

	for {
		select {
		case <-trainingTicker.C:
			pc.retrainAllModels()
		case <-cleanupTicker.C:
			pc.cleanupExpiredPatterns()
		}
	}
}

// retrainAllModels 重新训练所有模型
func (pc *PredictiveCache) retrainAllModels() {
	pc.mu.RLock()
	patterns := make(map[string]*AccessPattern)
	for k, v := range pc.accessPatterns {
		patterns[k] = v
	}
	pc.mu.RUnlock()

	for key, pattern := range patterns {
		if len(pattern.AccessTimes) >= pc.predictionConfig.MinAccessCount {
			pc.trainPredictionModel(key, pattern)
		}
	}

	pc.logger.Debug("Model retraining completed",
		zap.Int("models_trained", len(patterns)),
	)
}

// cleanupExpiredPatterns 清理过期的访问模式
func (pc *PredictiveCache) cleanupExpiredPatterns() {
	pc.mu.Lock()
	defer pc.mu.Unlock()

	now := time.Now()
	expiredKeys := make([]string, 0)

	for key, pattern := range pc.accessPatterns {
		// 如果超过7天没有访问，删除模式
		if len(pattern.AccessTimes) > 0 {
			lastAccess := pattern.AccessTimes[len(pattern.AccessTimes)-1]
			if now.Sub(lastAccess) > 7*24*time.Hour {
				expiredKeys = append(expiredKeys, key)
			}
		}
	}

	for _, key := range expiredKeys {
		delete(pc.accessPatterns, key)
		// 同时删除预测模型
		pc.predictionEngine.mu.Lock()
		delete(pc.predictionEngine.models, key)
		delete(pc.predictionEngine.predictionHistory, key)
		pc.predictionEngine.mu.Unlock()
	}

	if len(expiredKeys) > 0 {
		pc.logger.Info("Cleaned up expired access patterns",
			zap.Int("expired_patterns", len(expiredKeys)),
		)
	}
}

// GetPatternStats 获取访问模式统计
func (pc *PredictiveCache) GetPatternStats() map[string]interface{} {
	pc.mu.RLock()
	defer pc.mu.RUnlock()

	stats := make(map[string]interface{})
	stats["total_patterns"] = len(pc.accessPatterns)
	stats["predictable_patterns"] = 0
	stats["high_predictability_patterns"] = 0

	predictabilitySum := 0.0
	count := 0

	for _, pattern := range pc.accessPatterns {
		count++
		predictabilitySum += pattern.Predictability
		if pattern.Predictability >= pc.predictionConfig.MinPredictability {
			stats["predictable_patterns"] = stats["predictable_patterns"].(int) + 1
		}
		if pattern.Predictability >= 0.9 {
			stats["high_predictability_patterns"] = stats["high_predictability_patterns"].(int) + 1
		}
	}

	if count > 0 {
		stats["average_predictability"] = predictabilitySum / float64(count)
	}

	// 添加预测引擎统计
	pc.predictionEngine.mu.RLock()
	stats["total_models"] = len(pc.predictionEngine.models)
	stats["total_predictions"] = 0
	for _, history := range pc.predictionEngine.predictionHistory {
		stats["total_predictions"] = stats["total_predictions"].(int) + len(history)
	}
	pc.predictionEngine.mu.RUnlock()

	return stats
}

// minInt 返回最小值
func minInt(a, b int) int {
	if a < b {
		return a
	}
	return b
}