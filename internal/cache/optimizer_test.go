package cache

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"go.uber.org/zap/zaptest"
)

func TestCacheOptimizer_Creation(t *testing.T) {
	logger := zaptest.NewLogger(t)
	strategy := DefaultCacheStrategy()
	config := DefaultOptimizationConfig()

	optimizer := NewCacheOptimizer(strategy, config, logger)

	assert.NotNil(t, optimizer)
	assert.Equal(t, strategy, optimizer.currentStrategy)
	assert.Equal(t, config, optimizer.config)
	assert.False(t, optimizer.isOptimizing)
}

func TestCacheOptimizer_GenerateRandomStrategy(t *testing.T) {
	logger := zaptest.NewLogger(t)
	strategy := DefaultCacheStrategy()
	config := DefaultOptimizationConfig()
	optimizer := NewCacheOptimizer(strategy, config, logger)

	randomStrategy := optimizer.generateRandomStrategy()

	assert.NotNil(t, randomStrategy)
	assert.GreaterOrEqual(t, randomStrategy.HotDataThreshold, config.HotThresholdRange[0])
	assert.LessOrEqual(t, randomStrategy.HotDataThreshold, config.HotThresholdRange[1])
	assert.GreaterOrEqual(t, randomStrategy.WarmDataThreshold, config.WarmThresholdRange[0])
	assert.LessOrEqual(t, randomStrategy.WarmDataThreshold, config.WarmThresholdRange[1])
	assert.GreaterOrEqual(t, randomStrategy.L1MaxEntries, config.L1CapacityRange[0])
	assert.LessOrEqual(t, randomStrategy.L1MaxEntries, config.L1CapacityRange[1])
}

func TestCacheOptimizer_EvaluateStrategy(t *testing.T) {
	logger := zaptest.NewLogger(t)
	strategy := DefaultCacheStrategy()
	config := DefaultOptimizationConfig()
	optimizer := NewCacheOptimizer(strategy, config, logger)

	performance := optimizer.evaluateStrategy(strategy)

	assert.NotNil(t, performance)
	assert.GreaterOrEqual(t, performance.HitRate, 0.0)
	assert.LessOrEqual(t, performance.HitRate, 1.0)
	assert.Greater(t, performance.AvgResponseTime, time.Duration(0))
	assert.GreaterOrEqual(t, performance.MemoryUsage, int64(0))
	assert.GreaterOrEqual(t, performance.CPULoad, 0.0)
	assert.LessOrEqual(t, performance.CPULoad, 1.0)
}

func TestCacheOptimizer_CalculateFitness(t *testing.T) {
	logger := zaptest.NewLogger(t)
	strategy := DefaultCacheStrategy()
	config := DefaultOptimizationConfig()
	optimizer := NewCacheOptimizer(strategy, config, logger)

	tests := []struct {
		name       string
		metrics    *PerformanceMetrics
		expectedGT float64
	}{
		{
			name: "Good performance",
			metrics: &PerformanceMetrics{
				HitRate:         0.9,
				AvgResponseTime: 10 * time.Millisecond,
				MemoryUsage:     50 * 1024 * 1024, // 50MB
				CPULoad:         0.3,
				LastUpdateTime:  time.Now(),
			},
			expectedGT: 0.5,
		},
		{
			name: "Poor performance",
			metrics: &PerformanceMetrics{
				HitRate:         0.3,
				AvgResponseTime: 200 * time.Millisecond,
				MemoryUsage:     200 * 1024 * 1024, // 200MB
				CPULoad:         0.9,
				LastUpdateTime:  time.Now(),
			},
			expectedGT: 0.0,
		},
		{
			name: "Below minimum hit rate",
			metrics: &PerformanceMetrics{
				HitRate:         0.4, // Below MinHitRate (0.6)
				AvgResponseTime: 20 * time.Millisecond,
				MemoryUsage:     50 * 1024 * 1024,
				CPULoad:         0.3,
				LastUpdateTime:  time.Now(),
			},
			expectedGT: 0.0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fitness := optimizer.calculateFitness(tt.metrics)
			assert.GreaterOrEqual(t, fitness, tt.expectedGT)
			assert.LessOrEqual(t, fitness, 1.0)
		})
	}
}

func TestCacheOptimizer_ValidateStrategy(t *testing.T) {
	logger := zaptest.NewLogger(t)
	strategy := DefaultCacheStrategy()
	config := DefaultOptimizationConfig()
	optimizer := NewCacheOptimizer(strategy, config, logger)

	tests := []struct {
		name        string
		strategy    *CacheStrategy
		expectError bool
	}{
		{
			name: "Valid strategy",
			strategy: &CacheStrategy{
				HotDataThreshold:  0.8,
				WarmDataThreshold: 0.2,
				L1MaxEntries:      5000,
				L1TTL:             30 * time.Minute,
				L2TTL:             2 * time.Hour,
			},
			expectError: false,
		},
		{
			name: "Hot threshold <= warm threshold",
			strategy: &CacheStrategy{
				HotDataThreshold:  0.3,
				WarmDataThreshold: 0.5,
				L1MaxEntries:      5000,
				L1TTL:             30 * time.Minute,
				L2TTL:             2 * time.Hour,
			},
			expectError: true,
		},
		{
			name: "Negative capacity",
			strategy: &CacheStrategy{
				HotDataThreshold:  0.8,
				WarmDataThreshold: 0.2,
				L1MaxEntries:      -100,
				L1TTL:             30 * time.Minute,
				L2TTL:             2 * time.Hour,
			},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := optimizer.ValidateStrategy(tt.strategy)
			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestCacheOptimizer_InitializePopulation(t *testing.T) {
	logger := zaptest.NewLogger(t)
	strategy := DefaultCacheStrategy()
	config := &OptimizationConfig{
		PopulationSize: 10,
	}
	optimizer := NewCacheOptimizer(strategy, config, logger)

	population := optimizer.initializePopulation()

	assert.Len(t, population, config.PopulationSize)

	// 第一个个体应该是当前策略
	assert.Equal(t, strategy, population[0].Strategy)

	// 其他个体应该是不同的随机策略
	for i := 1; i < len(population); i++ {
		assert.NotEqual(t, strategy, population[i].Strategy)
		assert.NoError(t, optimizer.ValidateStrategy(population[i].Strategy))
	}
}

func TestCacheOptimizer_Selection(t *testing.T) {
	logger := zaptest.NewLogger(t)
	strategy := DefaultCacheStrategy()
	config := DefaultOptimizationConfig()
	optimizer := NewCacheOptimizer(strategy, config, logger)

	// 创建测试种群
	population := make([]StrategyIndividual, 5)
	fitnesses := []float64{0.2, 0.8, 0.5, 0.9, 0.1}

	for i, fitness := range fitnesses {
		population[i] = StrategyIndividual{
			Strategy: optimizer.generateRandomStrategy(),
			Fitness:  fitness,
		}
	}

	// 测试选择
	selectedCount := make(map[int]int)
	for i := 0; i < 100; i++ {
		selected := optimizer.selectParent(population)
		// 找到选中的个体索引
		for j, individual := range population {
			if individual.Strategy == selected.Strategy {
				selectedCount[j]++
				break
			}
		}
	}

	// 适应度高的个体应该被选中更多次
	assert.Greater(t, selectedCount[3], selectedCount[4]) // 0.9 > 0.1
	assert.Greater(t, selectedCount[1], selectedCount[0]) // 0.8 > 0.2
}

func TestCacheOptimizer_Crossover(t *testing.T) {
	logger := zaptest.NewLogger(t)
	strategy := DefaultCacheStrategy()
	config := DefaultOptimizationConfig()
	optimizer := NewCacheOptimizer(strategy, config, logger)

	parent1 := &CacheStrategy{
		HotDataThreshold:  0.8,
		WarmDataThreshold: 0.3,
		L1MaxEntries:      6000,
		L1TTL:             30 * time.Minute,
		L2TTL:             2 * time.Hour,
	}

	parent2 := &CacheStrategy{
		HotDataThreshold:  0.7,
		WarmDataThreshold: 0.2,
		L1MaxEntries:      5000,
		L1TTL:             45 * time.Minute,
		L2TTL:             3 * time.Hour,
	}

	child := optimizer.crossover(parent1, parent2)

	assert.NotNil(t, child)
	assert.NoError(t, optimizer.ValidateStrategy(child))

	// 子代应该有来自双亲的特征
	assert.True(t, child.HotDataThreshold == parent1.HotDataThreshold ||
		child.HotDataThreshold == parent2.HotDataThreshold)
	assert.True(t, child.WarmDataThreshold == parent1.WarmDataThreshold ||
		child.WarmDataThreshold == parent2.WarmDataThreshold)
}

func TestCacheOptimizer_Mutation(t *testing.T) {
	logger := zaptest.NewLogger(t)
	strategy := DefaultCacheStrategy()
	config := DefaultOptimizationConfig()
	optimizer := NewCacheOptimizer(strategy, config, logger)

	original := &CacheStrategy{
		HotDataThreshold:  0.8,
		WarmDataThreshold: 0.3,
		L1MaxEntries:      6000,
		L1TTL:             30 * time.Minute,
		L2TTL:             2 * time.Hour,
	}

	mutated := optimizer.mutate(original)

	assert.NotNil(t, mutated)
	assert.NoError(t, optimizer.ValidateStrategy(mutated))

	// 变异后的值应该在有效范围内
	assert.GreaterOrEqual(t, mutated.HotDataThreshold, config.HotThresholdRange[0])
	assert.LessOrEqual(t, mutated.HotDataThreshold, config.HotThresholdRange[1])
	assert.GreaterOrEqual(t, mutated.WarmDataThreshold, config.WarmThresholdRange[0])
	assert.LessOrEqual(t, mutated.WarmDataThreshold, config.WarmThresholdRange[1])
}

func TestCacheOptimizer_HasConverged(t *testing.T) {
	logger := zaptest.NewLogger(t)
	strategy := DefaultCacheStrategy()
	config := DefaultOptimizationConfig()
	optimizer := NewCacheOptimizer(strategy, config, logger)

	// 测试收敛的种群（适应度相似）
	convergedPopulation := make([]StrategyIndividual, 10)
	baseFitness := 0.8
	for i := range convergedPopulation {
		convergedPopulation[i] = StrategyIndividual{
			Strategy: optimizer.generateRandomStrategy(),
			Fitness:  baseFitness + float64(i)*0.001, // 很小的差异
		}
	}

	assert.True(t, optimizer.hasConverged(convergedPopulation))

	// 测试未收敛的种群（适应度差异大）
	divergedPopulation := make([]StrategyIndividual, 10)
	for i := range divergedPopulation {
		divergedPopulation[i] = StrategyIndividual{
			Strategy: optimizer.generateRandomStrategy(),
			Fitness:  float64(i) / 10.0, // 0.0 到 0.9
		}
	}

	assert.False(t, optimizer.hasConverged(divergedPopulation))
}

func TestCacheOptimizer_GetOptimizationStats(t *testing.T) {
	logger := zaptest.NewLogger(t)
	strategy := DefaultCacheStrategy()
	config := DefaultOptimizationConfig()
	optimizer := NewCacheOptimizer(strategy, config, logger)

	stats := optimizer.GetOptimizationStats()

	assert.NotNil(t, stats)
	assert.Contains(t, stats, "optimization_count")
	assert.Contains(t, stats, "is_optimizing")
	assert.Contains(t, stats, "best_fitness")
	assert.Contains(t, stats, "best_strategy")
	assert.Contains(t, stats, "best_performance")
	assert.Equal(t, 0, stats["optimization_count"])
	assert.False(t, stats["is_optimizing"].(bool))
}

func TestCacheOptimizer_ResetOptimization(t *testing.T) {
	logger := zaptest.NewLogger(t)
	strategy := DefaultCacheStrategy()
	config := DefaultOptimizationConfig()
	optimizer := NewCacheOptimizer(strategy, config, logger)

	// 添加一些历史记录
	optimizer.optimizationHistory = append(optimizer.optimizationHistory, OptimizationResultRecord{
		Timestamp:   time.Now(),
		Iteration:   1,
		Fitness:     0.8,
		Improvement: 0.1,
		Algorithm:   "test",
	})

	assert.Len(t, optimizer.optimizationHistory, 1)

	// 重置
	optimizer.ResetOptimization()

	assert.Len(t, optimizer.optimizationHistory, 0)
	assert.Equal(t, strategy, optimizer.bestStrategy)
}

func TestAutoTuner_Creation(t *testing.T) {
	logger := zaptest.NewLogger(t)
	strategy := DefaultCacheStrategy()
	config := DefaultAutoTuningConfig()

	tuner := NewAutoTuner(strategy, config, logger)

	assert.NotNil(t, tuner)
	assert.Equal(t, strategy, tuner.currentStrategy)
	assert.Equal(t, config, tuner.config)
	assert.False(t, tuner.isActive)
}

func TestAutoTuner_CollectPerformanceSnapshot(t *testing.T) {
	logger := zaptest.NewLogger(t)
	strategy := DefaultCacheStrategy()
	config := DefaultAutoTuningConfig()
	tuner := NewAutoTuner(strategy, config, logger)

	snapshot, err := tuner.collectPerformanceSnapshot(context.Background())

	assert.NoError(t, err)
	assert.NotNil(t, snapshot)
	assert.NotNil(t, snapshot.Metrics)
	assert.GreaterOrEqual(t, snapshot.LoadFactor, 0.0)
	assert.LessOrEqual(t, snapshot.LoadFactor, 1.0)
	assert.NotEmpty(t, snapshot.AccessPattern)
}

func TestAutoTuner_CalculatePerformanceTrend(t *testing.T) {
	logger := zaptest.NewLogger(t)
	strategy := DefaultCacheStrategy()
	config := DefaultAutoTuningConfig()
	tuner := NewAutoTuner(strategy, config, logger)

	// 添加性能历史
	now := time.Now()
	for i := 0; i < 5; i++ {
		snapshot := PerformanceSnapshot{
			Timestamp: now.Add(time.Duration(i) * time.Minute),
			Metrics: &PerformanceMetrics{
				HitRate:         0.7 + float64(i)*0.05,                                     // 递增
				AvgResponseTime: 30*time.Millisecond - time.Duration(i)*2*time.Millisecond, // 递减
				MemoryUsage:     int64(100*1024*1024 + i*10*1024*1024),                     // 递增
				LastUpdateTime:  now.Add(time.Duration(i) * time.Minute),
			},
			LoadFactor:    0.5,
			AccessPattern: "test",
		}
		tuner.performanceHistory = append(tuner.performanceHistory, snapshot)
	}

	trend := tuner.calculatePerformanceTrend()

	assert.NotNil(t, trend)
	assert.Greater(t, trend.HitRateTrend, 0.0)      // 递增趋势
	assert.Greater(t, trend.ResponseTimeTrend, 0.0) // 响应时间递减，取反后为正
	assert.Less(t, trend.MemoryTrend, 0.0)          // 内存递增，取反后为负
}

func TestAutoTuner_AnalyzeHitRateTrend(t *testing.T) {
	logger := zaptest.NewLogger(t)
	strategy := DefaultCacheStrategy()
	config := DefaultAutoTuningConfig()
	tuner := NewAutoTuner(strategy, config, logger)

	tests := []struct {
		name              string
		trend             *PerformanceTrend
		expectAction      bool
		expectedDirection string
	}{
		{
			name: "Good hit rate and trend",
			trend: &PerformanceTrend{
				HitRateTrend:        0.1,
				CurrentHitRate:      0.9,
				CurrentResponseTime: 20 * time.Millisecond,
			},
			expectAction: false,
		},
		{
			name: "Low hit rate",
			trend: &PerformanceTrend{
				HitRateTrend:        -0.1,
				CurrentHitRate:      0.6,
				CurrentResponseTime: 20 * time.Millisecond,
			},
			expectAction:      true,
			expectedDirection: "decrease",
		},
		{
			name: "Declining hit rate",
			trend: &PerformanceTrend{
				HitRateTrend:        -0.2,
				CurrentHitRate:      0.8,
				CurrentResponseTime: 20 * time.Millisecond,
			},
			expectAction:      true,
			expectedDirection: "decrease",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			action := tuner.analyzeHitRateTrend(tt.trend)

			if tt.expectAction {
				assert.NotNil(t, action)
				assert.Equal(t, "HotDataThreshold", action.Parameter)
				assert.Equal(t, tt.expectedDirection, getDirectionFromValue(action.OldValue, action.NewValue))
			} else {
				assert.Nil(t, action)
			}
		})
	}
}

func TestAutoTuner_PredictPerformanceTrend(t *testing.T) {
	logger := zaptest.NewLogger(t)
	strategy := DefaultCacheStrategy()
	config := DefaultAutoTuningConfig()
	tuner := NewAutoTuner(strategy, config, logger)

	// 添加稳定的性能历史
	now := time.Now()
	for i := 0; i < 5; i++ {
		snapshot := PerformanceSnapshot{
			Timestamp: now.Add(time.Duration(i) * time.Minute),
			Metrics: &PerformanceMetrics{
				HitRate:         0.8,
				AvgResponseTime: 20 * time.Millisecond,
				MemoryUsage:     100 * 1024 * 1024,
				LastUpdateTime:  now.Add(time.Duration(i) * time.Minute),
			},
			LoadFactor:    0.5,
			AccessPattern: "stable",
		}
		tuner.performanceHistory = append(tuner.performanceHistory, snapshot)
	}

	prediction := tuner.predictPerformanceTrend()

	assert.NotNil(t, prediction)
	assert.GreaterOrEqual(t, prediction.PredictedHitRate, 0.0)
	assert.LessOrEqual(t, prediction.PredictedHitRate, 1.0)
	assert.Greater(t, prediction.PredictedResponseTime, time.Duration(0))
	assert.GreaterOrEqual(t, prediction.Confidence, 0.0)
	assert.LessOrEqual(t, prediction.Confidence, 1.0)
}

func TestAutoTuner_GetTuningStats(t *testing.T) {
	logger := zaptest.NewLogger(t)
	strategy := DefaultCacheStrategy()
	config := DefaultAutoTuningConfig()
	tuner := NewAutoTuner(strategy, config, logger)

	stats := tuner.GetTuningStats()

	assert.NotNil(t, stats)
	assert.Contains(t, stats, "is_active")
	assert.Contains(t, stats, "total_tuning_actions")
	assert.Contains(t, stats, "performance_snapshots")
	assert.Contains(t, stats, "current_strategy")
	assert.False(t, stats["is_active"].(bool))
	assert.Equal(t, 0, stats["total_tuning_actions"])
}

// 辅助函数
func getDirectionFromValue(oldValue, newValue interface{}) string {
	oldFloat, oldOk := oldValue.(float64)
	newFloat, newOk := newValue.(float64)

	if oldOk && newOk {
		if newFloat < oldFloat {
			return "decrease"
		} else if newFloat > oldFloat {
			return "increase"
		}
	}

	return "unknown"
}

func TestAutoTuner_Reset(t *testing.T) {
	logger := zaptest.NewLogger(t)
	strategy := DefaultCacheStrategy()
	config := DefaultAutoTuningConfig()
	tuner := NewAutoTuner(strategy, config, logger)

	// 添加一些历史数据
	tuner.performanceHistory = append(tuner.performanceHistory, PerformanceSnapshot{
		Timestamp: time.Now(),
		Metrics:   &PerformanceMetrics{HitRate: 0.8},
	})
	tuner.tuningHistory = append(tuner.tuningHistory, TuningAction{
		Timestamp: time.Now(),
		Action:    "test",
	})

	assert.Len(t, tuner.performanceHistory, 1)
	assert.Len(t, tuner.tuningHistory, 1)

	// 重置
	tuner.Reset()

	assert.Len(t, tuner.performanceHistory, 0)
	assert.Len(t, tuner.tuningHistory, 0)
}

// 基准测试
func BenchmarkCacheOptimizer_EvaluateStrategy(b *testing.B) {
	logger := zaptest.NewLogger(b)
	strategy := DefaultCacheStrategy()
	config := DefaultOptimizationConfig()
	optimizer := NewCacheOptimizer(strategy, config, logger)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		testStrategy := optimizer.generateRandomStrategy()
		optimizer.evaluateStrategy(testStrategy)
	}
}

func BenchmarkAutoTuner_CollectPerformanceSnapshot(b *testing.B) {
	logger := zaptest.NewLogger(b)
	strategy := DefaultCacheStrategy()
	config := DefaultAutoTuningConfig()
	tuner := NewAutoTuner(strategy, config, logger)
	ctx := context.Background()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		tuner.collectPerformanceSnapshot(ctx)
	}
}
