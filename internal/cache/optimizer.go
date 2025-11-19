package cache

import (
	"context"
	"fmt"
	"math"
	"math/rand"
	"sort"
	"sync"
	"time"

	"go.uber.org/zap"
)

// CacheOptimizer 缓存策略参数优化器
type CacheOptimizer struct {
	mu                  sync.RWMutex
	config              *OptimizationConfig
	currentStrategy     *CacheStrategy
	optimizationHistory []OptimizationResultRecord
	bestStrategy        *CacheStrategy
	bestPerformance     *PerformanceMetrics
	isOptimizing        bool
	stopCh              chan struct{}
	logger              *zap.Logger
}

// OptimizationConfig 优化配置
type OptimizationConfig struct {
	// 优化参数范围
	HotThresholdRange  [2]float64       `json:"hot_threshold_range"`  // [min, max]
	WarmThresholdRange [2]float64       `json:"warm_threshold_range"` // [min, max]
	L1CapacityRange    [2]int           `json:"l1_capacity_range"`    // [min, max]
	L2TTLRange         [2]time.Duration `json:"l2_ttl_range"`         // [min, max]

	// 优化控制参数
	OptimizationInterval time.Duration `json:"optimization_interval"` // 优化间隔
	TestDuration         time.Duration `json:"test_duration"`         // 每个策略测试时间
	MaxIterations        int           `json:"max_iterations"`        // 最大迭代次数
	ImprovementThreshold float64       `json:"improvement_threshold"` // 改善阈值
	StagnationLimit      int           `json:"stagnation_limit"`      // 停滞限制

	// 优化算法参数
	PopulationSize int     `json:"population_size"` // 种群大小
	MutationRate   float64 `json:"mutation_rate"`   // 变异率
	CrossoverRate  float64 `json:"crossover_rate"`  // 交叉率
	ElitismRatio   float64 `json:"elitism_ratio"`   // 精英保留比例

	// 性能权重
	HitRateWeight      float64 `json:"hit_rate_weight"`      // 命中率权重
	ResponseTimeWeight float64 `json:"response_time_weight"` // 响应时间权重
	MemoryWeight       float64 `json:"memory_weight"`        // 内存权重
	CPUWeight          float64 `json:"cpu_weight"`           // CPU权重

	// 安全约束
	MinHitRate      float64       `json:"min_hit_rate"`      // 最小命中率
	MaxResponseTime time.Duration `json:"max_response_time"` // 最大响应时间
	MaxMemoryUsage  int64         `json:"max_memory_usage"`  // 最大内存使用
}

// DefaultOptimizationConfig 默认优化配置
func DefaultOptimizationConfig() *OptimizationConfig {
	return &OptimizationConfig{
		HotThresholdRange:  [2]float64{0.5, 0.95},
		WarmThresholdRange: [2]float64{0.05, 0.5},
		L1CapacityRange:    [2]int{1000, 10000},
		L2TTLRange:         [2]time.Duration{5 * time.Minute, 2 * time.Hour},

		OptimizationInterval: 30 * time.Minute,
		TestDuration:         5 * time.Minute,
		MaxIterations:        50,
		ImprovementThreshold: 0.05,
		StagnationLimit:      10,

		PopulationSize: 20,
		MutationRate:   0.2,
		CrossoverRate:  0.8,
		ElitismRatio:   0.1,

		HitRateWeight:      0.4,
		ResponseTimeWeight: 0.4,
		MemoryWeight:       0.1,
		CPUWeight:          0.1,

		MinHitRate:      0.6,
		MaxResponseTime: 100 * time.Millisecond,
		MaxMemoryUsage:  200 * 1024 * 1024, // 200MB
	}
}

// StrategyIndividual 策略个体（用于遗传算法）
type StrategyIndividual struct {
	Strategy    *CacheStrategy      `json:"strategy"`
	Fitness     float64             `json:"fitness"`
	Performance *PerformanceMetrics `json:"performance"`
}

// OptimizationResultRecord 优化结果记录
type OptimizationResultRecord struct {
	Timestamp        time.Time           `json:"timestamp"`
	Iteration        int                 `json:"iteration"`
	Strategy         *CacheStrategy      `json:"strategy"`
	Performance      *PerformanceMetrics `json:"performance"`
	Fitness          float64             `json:"fitness"`
	Improvement      float64             `json:"improvement"`
	Algorithm        string              `json:"algorithm"`
	OptimizationTime time.Duration       `json:"optimization_time"`
}

// NewCacheOptimizer 创建缓存优化器
func NewCacheOptimizer(initialStrategy *CacheStrategy, config *OptimizationConfig, logger *zap.Logger) *CacheOptimizer {
	if config == nil {
		config = DefaultOptimizationConfig()
	}

	optimizer := &CacheOptimizer{
		config:              config,
		currentStrategy:     initialStrategy,
		bestStrategy:        initialStrategy,
		optimizationHistory: make([]OptimizationResultRecord, 0),
		stopCh:              make(chan struct{}),
		logger:              logger.Named("cache_optimizer"),
	}

	// 初始化最佳性能指标
	optimizer.bestPerformance = &PerformanceMetrics{
		HitRate:         0.0,
		AvgResponseTime: time.Hour, // 初始设为很大的值
		MemoryUsage:     0,
		CPULoad:         0,
		LastUpdateTime:  time.Now(),
	}

	return optimizer
}

// StartOptimization 启动自动优化
func (co *CacheOptimizer) StartOptimization(ctx context.Context) {
	co.mu.Lock()
	if co.isOptimizing {
		co.mu.Unlock()
		return
	}
	co.isOptimizing = true
	co.mu.Unlock()

	co.logger.Info("Starting cache strategy optimization")

	ticker := time.NewTicker(co.config.OptimizationInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			co.StopOptimization()
			return
		case <-co.stopCh:
			return
		case <-ticker.C:
			if co.shouldOptimize() {
				co.performOptimization(ctx)
			}
		}
	}
}

// StopOptimization 停止优化
func (co *CacheOptimizer) StopOptimization() {
	co.mu.Lock()
	defer co.mu.Unlock()

	if co.isOptimizing {
		co.isOptimizing = false
		close(co.stopCh)
		co.stopCh = make(chan struct{})
		co.logger.Info("Cache strategy optimization stopped")
	}
}

// shouldOptimize 判断是否需要进行优化
func (co *CacheOptimizer) shouldOptimize() bool {
	co.mu.RLock()
	defer co.mu.RUnlock()

	// 如果历史记录不足，不优化
	if len(co.optimizationHistory) < 3 {
		return true
	}

	// 检查最近的改善情况
	recentResults := co.optimizationHistory[len(co.optimizationHistory)-3:]
	noImprovementCount := 0

	for i := 1; i < len(recentResults); i++ {
		if recentResults[i].Improvement < co.config.ImprovementThreshold {
			noImprovementCount++
		}
	}

	// 如果连续几次没有显著改善，触发优化
	return noImprovementCount >= 2
}

// performOptimization 执行优化
func (co *CacheOptimizer) performOptimization(ctx context.Context) {
	co.logger.Info("Starting cache strategy optimization cycle")

	startTime := time.Now()
	bestResult, err := co.geneticAlgorithmOptimization(ctx)
	optimizationTime := time.Since(startTime)

	if err != nil {
		co.logger.Error("Optimization failed", zap.Error(err))
		return
	}

	if bestResult != nil && bestResult.Fitness > co.calculateFitness(co.bestPerformance) {
		co.mu.Lock()
		co.bestStrategy = bestResult.Strategy
		co.bestPerformance = bestResult.Performance
		co.currentStrategy = bestResult.Strategy
		co.optimizationHistory = append(co.optimizationHistory, *bestResult)
		co.mu.Unlock()

		co.logger.Info("Cache strategy optimized",
			zap.Float64("fitness", bestResult.Fitness),
			zap.Float64("improvement", bestResult.Improvement),
			zap.Duration("optimization_time", optimizationTime),
		)
	} else {
		co.logger.Debug("No better strategy found, keeping current strategy")
	}
}

// geneticAlgorithmOptimization 遗传算法优化
func (co *CacheOptimizer) geneticAlgorithmOptimization(ctx context.Context) (*OptimizationResultRecord, error) {
	// 初始化种群
	population := co.initializePopulation()

	bestIndividual := &StrategyIndividual{
		Strategy: co.currentStrategy,
		Fitness:  co.calculateFitness(co.bestPerformance),
	}

	for iteration := 0; iteration < co.config.MaxIterations; iteration++ {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		default:
		}

		// 评估种群
		for i := range population {
			individual := &population[i]
			individual.Performance = co.evaluateStrategy(individual.Strategy)
			individual.Fitness = co.calculateFitness(individual.Performance)

			// 更新最佳个体
			if individual.Fitness > bestIndividual.Fitness {
				bestIndividual = individual
			}
		}

		// 检查收敛条件
		if co.hasConverged(population) {
			co.logger.Info("Genetic algorithm converged", zap.Int("iteration", iteration))
			break
		}

		// 选择、交叉、变异
		population = co.evolvePopulation(population)

		co.logger.Debug("Genetic algorithm iteration",
			zap.Int("iteration", iteration),
			zap.Float64("best_fitness", bestIndividual.Fitness),
		)
	}

	improvement := (bestIndividual.Fitness - co.calculateFitness(co.bestPerformance)) / co.calculateFitness(co.bestPerformance)

	return &OptimizationResultRecord{
		Timestamp:        time.Now(),
		Iteration:        co.config.MaxIterations,
		Strategy:         bestIndividual.Strategy,
		Performance:      bestIndividual.Performance,
		Fitness:          bestIndividual.Fitness,
		Improvement:      improvement,
		Algorithm:        "genetic_algorithm",
		OptimizationTime: 0, // 将在上层设置
	}, nil
}

// initializePopulation 初始化种群
func (co *CacheOptimizer) initializePopulation() []StrategyIndividual {
	population := make([]StrategyIndividual, co.config.PopulationSize)

	// 包含当前策略
	population[0] = StrategyIndividual{
		Strategy: co.currentStrategy,
		Fitness:  co.calculateFitness(co.bestPerformance),
	}

	// 随机生成其他策略
	for i := 1; i < co.config.PopulationSize; i++ {
		strategy := co.generateRandomStrategy()
		population[i] = StrategyIndividual{
			Strategy: strategy,
		}
	}

	return population
}

// generateRandomStrategy 生成随机策略
func (co *CacheOptimizer) generateRandomStrategy() *CacheStrategy {
	hotRange := co.config.HotThresholdRange
	warmRange := co.config.WarmThresholdRange
	l1Range := co.config.L1CapacityRange
	ttlRange := co.config.L2TTLRange

	// 安全的随机数生成，确保范围有效
	hotThreshold := hotRange[0]
	if hotRange[1] > hotRange[0] {
		hotThreshold = hotRange[0] + rand.Float64()*(hotRange[1]-hotRange[0])
	}

	warmThreshold := warmRange[0]
	if warmRange[1] > warmRange[0] {
		warmThreshold = warmRange[0] + rand.Float64()*(warmRange[1]-warmRange[0])
	}

	// 确保hot threshold > warm threshold
	if hotThreshold <= warmThreshold {
		// 如果生成的hot threshold不大于warm threshold，则重新生成
		hotThreshold = warmThreshold + 0.1 // 确保至少比warm threshold大0.1
		if hotRange[1] > hotRange[0] && hotThreshold > hotRange[1] {
			hotThreshold = hotRange[1] // 如果超出范围，使用最大值
		}
	}

	l1MaxEntries := l1Range[0]
	if l1Range[1] > l1Range[0] && l1Range[0] > 0 {
		l1MaxEntries = l1Range[0] + rand.Intn(l1Range[1]-l1Range[0])
	} else {
		// 如果范围无效或者不是正数，使用默认值
		l1MaxEntries = 100 // 默认100个条目
	}

	l1TTL := ttlRange[0]
	if ttlRange[1] > ttlRange[0] && ttlRange[0] > 0 {
		l1TTL = ttlRange[0] + time.Duration(rand.Int63n(int64(ttlRange[1]-ttlRange[0])))
	} else {
		// 如果范围无效或者不是正数，使用默认值
		l1TTL = 10 * time.Minute // 默认10分钟
	}

	l2TTL := ttlRange[0]
	if ttlRange[1] > ttlRange[0] && ttlRange[0] > 0 {
		l2TTL = ttlRange[0] + time.Duration(rand.Int63n(int64(ttlRange[1]-ttlRange[0])))
	} else {
		// 如果范围无效或者不是正数，使用默认值
		l2TTL = 30 * time.Minute // 默认30分钟
	}

	return &CacheStrategy{
		HotDataThreshold:   hotThreshold,
		WarmDataThreshold:  warmThreshold,
		L1MaxEntries:       l1MaxEntries,
		L1TTL:              l1TTL,
		L2TTL:              l2TTL,
		PreloadEnabled:     rand.Float64() > 0.5,
		PreloadConcurrency: 5 + rand.Intn(10), // 5-15个并发
	}
}

// evaluateStrategy 评估策略性能
func (co *CacheOptimizer) evaluateStrategy(strategy *CacheStrategy) *PerformanceMetrics {
	// 模拟性能评估
	// 在实际实现中，这里应该使用真实的性能测试

	// 基于策略参数预测性能
	hotThreshold := strategy.HotDataThreshold
	warmThreshold := strategy.WarmDataThreshold
	l1Capacity := strategy.L1MaxEntries

	// 模拟性能指标（实际实现中应该进行真实测试）
	hitRate := co.predictHitRate(hotThreshold, warmThreshold, l1Capacity)
	responseTime := co.predictResponseTime(hotThreshold, l1Capacity)
	memoryUsage := co.predictMemoryUsage(l1Capacity)
	cpuLoad := co.predictCPULoad(hotThreshold, warmThreshold)

	return &PerformanceMetrics{
		HitRate:         hitRate,
		AvgResponseTime: responseTime,
		MemoryUsage:     memoryUsage,
		CPULoad:         cpuLoad,
		LastUpdateTime:  time.Now(),
	}
}

// predictHitRate 预测命中率
func (co *CacheOptimizer) predictHitRate(hotThreshold, warmThreshold float64, l1Capacity int) float64 {
	// 简化的预测模型
	baseHitRate := 0.6
	hotBonus := hotThreshold * 0.3
	warmBonus := warmThreshold * 0.1
	capacityBonus := math.Min(float64(l1Capacity)/10000.0, 0.2)

	predicted := baseHitRate + hotBonus + warmBonus + capacityBonus
	return math.Min(0.95, predicted)
}

// predictResponseTime 预测响应时间
func (co *CacheOptimizer) predictResponseTime(hotThreshold float64, l1Capacity int) time.Duration {
	// 热点阈值越高，L1缓存命中率越高，响应时间越短
	baseTime := 50 * time.Millisecond
	hotImprovement := time.Duration(hotThreshold * 40 * float64(time.Millisecond))
	capacityImprovement := time.Duration(float64(l1Capacity) / 10000.0 * 30 * float64(time.Millisecond))

	predicted := baseTime - hotImprovement - capacityImprovement
	return time.Duration(math.Max(float64(5*time.Millisecond), float64(predicted)))
}

// predictMemoryUsage 预测内存使用
func (co *CacheOptimizer) predictMemoryUsage(l1Capacity int) int64 {
	// 基于容量预测内存使用
	entrySize := int64(1024) // 假设每个条目1KB
	return int64(l1Capacity) * entrySize
}

// predictCPULoad 预测CPU负载
func (co *CacheOptimizer) predictCPULoad(hotThreshold, warmThreshold float64) float64 {
	// 阈值越高，缓存策略越复杂，CPU负载越高
	baseLoad := 0.1
	complexityLoad := (hotThreshold + warmThreshold) * 0.3
	return math.Min(0.8, baseLoad+complexityLoad)
}

// calculateFitness 计算适应度
func (co *CacheOptimizer) calculateFitness(performance *PerformanceMetrics) float64 {
	// 检查约束条件
	if performance.HitRate < co.config.MinHitRate {
		return 0.0
	}
	if performance.AvgResponseTime > co.config.MaxResponseTime {
		return 0.0
	}
	if performance.MemoryUsage > co.config.MaxMemoryUsage {
		return 0.0
	}

	// 计算加权适应度
	hitRateScore := performance.HitRate * co.config.HitRateWeight
	responseTimeScore := (1.0 - float64(performance.AvgResponseTime)/float64(co.config.MaxResponseTime)) * co.config.ResponseTimeWeight
	memoryScore := (1.0 - float64(performance.MemoryUsage)/float64(co.config.MaxMemoryUsage)) * co.config.MemoryWeight
	cpuScore := (1.0 - performance.CPULoad) * co.config.CPUWeight

	totalFitness := hitRateScore + responseTimeScore + memoryScore + cpuScore
	return math.Max(0.0, totalFitness)
}

// hasConverged 检查种群是否收敛
func (co *CacheOptimizer) hasConverged(population []StrategyIndividual) bool {
	if len(population) < 2 {
		return true
	}

	// 计算适应度的标准差
	var sum, sumSquares float64
	for _, individual := range population {
		sum += individual.Fitness
		sumSquares += individual.Fitness * individual.Fitness
	}

	mean := sum / float64(len(population))
	variance := (sumSquares / float64(len(population))) - (mean * mean)
	stdDev := math.Sqrt(variance)

	// 如果标准差很小，认为收敛
	return stdDev < 0.01
}

// evolvePopulation 进化种群
func (co *CacheOptimizer) evolvePopulation(population []StrategyIndividual) []StrategyIndividual {
	// 排序，按适应度降序
	sort.Slice(population, func(i, j int) bool {
		return population[i].Fitness > population[j].Fitness
	})

	newPopulation := make([]StrategyIndividual, co.config.PopulationSize)

	// 精英保留
	eliteCount := int(float64(co.config.PopulationSize) * co.config.ElitismRatio)
	for i := 0; i < eliteCount; i++ {
		newPopulation[i] = population[i]
	}

	// 生成新个体
	for i := eliteCount; i < co.config.PopulationSize; i++ {
		if rand.Float64() < co.config.CrossoverRate {
			// 交叉
			parent1 := co.selectParent(population)
			parent2 := co.selectParent(population)
			child := co.crossover(parent1.Strategy, parent2.Strategy)

			if rand.Float64() < co.config.MutationRate {
				// 变异
				child = co.mutate(child)
			}

			newPopulation[i] = StrategyIndividual{Strategy: child}
		} else {
			// 直接选择
			parent := co.selectParent(population)
			newPopulation[i] = StrategyIndividual{Strategy: parent.Strategy}
		}
	}

	return newPopulation
}

// selectParent 选择父代（轮盘赌选择）
func (co *CacheOptimizer) selectParent(population []StrategyIndividual) StrategyIndividual {
	// 计算总适应度
	var totalFitness float64
	for _, individual := range population {
		totalFitness += individual.Fitness
	}

	if totalFitness == 0 {
		return population[rand.Intn(len(population))]
	}

	// 轮盘赌选择
	r := rand.Float64() * totalFitness
	currentSum := 0.0

	for _, individual := range population {
		currentSum += individual.Fitness
		if currentSum >= r {
			return individual
		}
	}

	return population[len(population)-1]
}

// crossover 交叉操作
func (co *CacheOptimizer) crossover(parent1, parent2 *CacheStrategy) *CacheStrategy {
	child := &CacheStrategy{}

	// 单点交叉
	if rand.Float64() < 0.5 {
		child.HotDataThreshold = parent1.HotDataThreshold
		child.WarmDataThreshold = parent2.WarmDataThreshold
		child.L1MaxEntries = parent1.L1MaxEntries
		child.L1TTL = parent2.L1TTL
		child.L2TTL = parent1.L2TTL
	} else {
		child.HotDataThreshold = parent2.HotDataThreshold
		child.WarmDataThreshold = parent1.WarmDataThreshold
		child.L1MaxEntries = parent2.L1MaxEntries
		child.L1TTL = parent1.L1TTL
		child.L2TTL = parent2.L2TTL
	}

	// 其他参数随机选择或取平均值
	child.PreloadEnabled = rand.Float64() > 0.5
	child.PreloadConcurrency = (parent1.PreloadConcurrency + parent2.PreloadConcurrency) / 2
	child.L1CleanupInterval = parent1.L1CleanupInterval
	child.L2CleanupInterval = parent2.L2CleanupInterval
	child.AccessFreqWindow = parent1.AccessFreqWindow

	return child
}

// mutate 变异操作
func (co *CacheOptimizer) mutate(strategy *CacheStrategy) *CacheStrategy {
	mutated := *strategy

	// 随机选择一个参数进行变异
	switch rand.Intn(5) {
	case 0:
		// 变异热点阈值
		hotRange := co.config.HotThresholdRange
		delta := (hotRange[1] - hotRange[0]) * 0.1 * (rand.Float64()*2 - 1)
		mutated.HotDataThreshold = math.Max(hotRange[0], math.Min(hotRange[1], mutated.HotDataThreshold+delta))
	case 1:
		// 变异温点阈值
		warmRange := co.config.WarmThresholdRange
		delta := (warmRange[1] - warmRange[0]) * 0.1 * (rand.Float64()*2 - 1)
		mutated.WarmDataThreshold = math.Max(warmRange[0], math.Min(warmRange[1], mutated.WarmDataThreshold+delta))
	case 2:
		// 变异L1容量
		l1Range := co.config.L1CapacityRange
		delta := int(float64(l1Range[1]-l1Range[0]) * 0.1 * (rand.Float64()*2 - 1))
		newCapacity := mutated.L1MaxEntries + delta
		if newCapacity < l1Range[0] {
			newCapacity = l1Range[0]
		} else if newCapacity > l1Range[1] {
			newCapacity = l1Range[1]
		}
		mutated.L1MaxEntries = newCapacity
	case 3:
		// 变异TTL
		ttlRange := co.config.L2TTLRange
		delta := time.Duration(float64(ttlRange[1]-ttlRange[0]) * 0.1 * (rand.Float64()*2 - 1))
		newTTL := mutated.L2TTL + delta
		if newTTL < ttlRange[0] {
			newTTL = ttlRange[0]
		} else if newTTL > ttlRange[1] {
			newTTL = ttlRange[1]
		}
		mutated.L2TTL = newTTL
	case 4:
		// 翻转布尔标志
		mutated.PreloadEnabled = !mutated.PreloadEnabled
	}

	return &mutated
}

// GetCurrentStrategy 获取当前最优策略
func (co *CacheOptimizer) GetCurrentStrategy() *CacheStrategy {
	co.mu.RLock()
	defer co.mu.RUnlock()

	return co.bestStrategy
}

// GetOptimizationHistory 获取优化历史
func (co *CacheOptimizer) GetOptimizationHistory(limit int) []OptimizationResultRecord {
	co.mu.RLock()
	defer co.mu.RUnlock()

	if limit <= 0 || limit > len(co.optimizationHistory) {
		limit = len(co.optimizationHistory)
	}

	result := make([]OptimizationResultRecord, limit)
	copy(result, co.optimizationHistory[len(co.optimizationHistory)-limit:])
	return result
}

// GetOptimizationStats 获取优化统计信息
func (co *CacheOptimizer) GetOptimizationStats() map[string]interface{} {
	co.mu.RLock()
	defer co.mu.RUnlock()

	stats := map[string]interface{}{
		"optimization_count": len(co.optimizationHistory),
		"is_optimizing":      co.isOptimizing,
		"best_fitness":       co.calculateFitness(co.bestPerformance),
		"best_strategy":      co.bestStrategy,
		"best_performance":   co.bestPerformance,
		"last_optimization":  "none",
		"total_improvement":  0.0,
	}

	if len(co.optimizationHistory) > 0 {
		lastResult := co.optimizationHistory[len(co.optimizationHistory)-1]
		stats["last_optimization"] = lastResult.Timestamp.Format(time.RFC3339)
		stats["total_improvement"] = co.calculateTotalImprovement()
	}

	return stats
}

// calculateTotalImprovement 计算总改善
func (co *CacheOptimizer) calculateTotalImprovement() float64 {
	if len(co.optimizationHistory) < 2 {
		return 0.0
	}

	first := co.optimizationHistory[0]
	last := co.optimizationHistory[len(co.optimizationHistory)-1]

	return (last.Fitness - first.Fitness) / first.Fitness
}

// UpdateCurrentPerformance 更新当前性能指标
func (co *CacheOptimizer) UpdateCurrentPerformance(performance *PerformanceMetrics) {
	co.mu.Lock()
	defer co.mu.Unlock()

	co.bestPerformance = performance
}

// SetOptimizationConfig 设置优化配置
func (co *CacheOptimizer) SetOptimizationConfig(config *OptimizationConfig) {
	co.mu.Lock()
	defer co.mu.Unlock()

	co.config = config
	co.logger.Info("Optimization config updated")
}

// ResetOptimization 重置优化
func (co *CacheOptimizer) ResetOptimization() {
	co.mu.Lock()
	defer co.mu.Unlock()

	co.optimizationHistory = make([]OptimizationResultRecord, 0)
	co.bestStrategy = co.currentStrategy
	co.bestPerformance = &PerformanceMetrics{
		HitRate:         0.0,
		AvgResponseTime: time.Hour,
		MemoryUsage:     0,
		CPULoad:         0,
		LastUpdateTime:  time.Now(),
	}

	co.logger.Info("Optimization reset")
}

// ValidateStrategy 验证策略有效性
func (co *CacheOptimizer) ValidateStrategy(strategy *CacheStrategy) error {
	if strategy.HotDataThreshold <= strategy.WarmDataThreshold {
		return fmt.Errorf("hot threshold must be greater than warm threshold")
	}

	if strategy.HotDataThreshold < 0 || strategy.HotDataThreshold > 1 {
		return fmt.Errorf("hot threshold must be between 0 and 1")
	}

	if strategy.WarmDataThreshold < 0 || strategy.WarmDataThreshold > 1 {
		return fmt.Errorf("warm threshold must be between 0 and 1")
	}

	if strategy.L1MaxEntries <= 0 {
		return fmt.Errorf("L1 max entries must be positive")
	}

	if strategy.L1TTL <= 0 || strategy.L2TTL <= 0 {
		return fmt.Errorf("TTL values must be positive")
	}

	return nil
}
