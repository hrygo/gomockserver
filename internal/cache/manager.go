package cache

import (
	"context"
	"fmt"
	"sync"
	"time"

	"go.uber.org/zap"
)

// ThreeLevelCacheManager 三级缓存管理器
type ThreeLevelCacheManager struct {
	// 缓存组件
	l1Cache L1Cache
	l2Cache L2Cache
	tracker FrequencyTracker

	// 配置和状态
	strategy *CacheStrategy
	stats    *CacheStats
	logger   *zap.Logger

	// 控制
	ctx     context.Context
	cancel  context.CancelFunc
	started bool
	mu      sync.RWMutex
}

// NewThreeLevelCacheManager 创建三级缓存管理器
func NewThreeLevelCacheManager(
	l1Cache L1Cache,
	l2Cache L2Cache,
	tracker FrequencyTracker,
	strategy *CacheStrategy,
	logger *zap.Logger,
) *ThreeLevelCacheManager {
	if strategy == nil {
		strategy = DefaultCacheStrategy()
	}

	ctx, cancel := context.WithCancel(context.Background())

	manager := &ThreeLevelCacheManager{
		l1Cache:  l1Cache,
		l2Cache:  l2Cache,
		tracker:  tracker,
		strategy: strategy,
		stats:    &CacheStats{},
		logger:   logger,
		ctx:      ctx,
		cancel:   cancel,
	}

	logger.Info("Three-level cache manager created",
		zap.Float64("hot_threshold", strategy.HotDataThreshold),
		zap.Float64("warm_threshold", strategy.WarmDataThreshold),
	)

	return manager
}

// Start 启动缓存管理器
func (m *ThreeLevelCacheManager) Start(ctx context.Context) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.started {
		return fmt.Errorf("cache manager already started")
	}

	// 启动后台任务
	go m.backgroundTasks()

	// 如果启用了预热，执行预热
	if m.strategy.PreloadEnabled && len(m.strategy.PreloadKeys) > 0 {
		go m.preloadKeys(ctx, m.strategy.PreloadKeys)
	}

	m.started = true
	m.logger.Info("Three-level cache manager started")
	return nil
}

// Stop 停止缓存管理器
func (m *ThreeLevelCacheManager) Stop(ctx context.Context) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if !m.started {
		return nil
	}

	m.cancel()
	m.started = false

	// 停止L1缓存（如果支持）
	if stopper, ok := m.l1Cache.(interface{ Stop() }); ok {
		stopper.Stop()
	}

	// 关闭L2缓存（如果支持）
	if closer, ok := m.l2Cache.(interface{ Close() error }); ok {
		closer.Close()
	}

	m.logger.Info("Three-level cache manager stopped")
	return nil
}

// Get 获取缓存值
func (m *ThreeLevelCacheManager) Get(ctx context.Context, key string) (interface{}, error) {
	start := time.Now()
	defer func() {
		m.updateAvgResponseTime(time.Since(start))
	}()

	m.stats.TotalRequests++

	// 记录访问频率
	m.tracker.RecordAccess(key)

	// L1缓存查找
	if entry, found := m.l1Cache.Get(key); found {
		m.stats.L1HitCount++
		m.updateHitRates()
		m.logger.Debug("L1 cache hit", zap.String("key", key))
		return entry.Value, nil
	}

	// L2缓存查找
	value, err := m.l2Cache.Get(ctx, key)
	if err == nil && value != nil {
		m.stats.L2HitCount++
		m.updateHitRates()

		// 将数据提升到L1缓存
		frequency := m.tracker.GetFrequency(key)
		if m.shouldPromoteToL1(frequency) {
			ttl := m.calculateL1TTL(frequency)
			if err := m.l1Cache.Set(key, value, ttl); err != nil {
				m.logger.Warn("Failed to promote to L1 cache",
					zap.String("key", key),
					zap.Error(err),
				)
			}
		}

		m.logger.Debug("L2 cache hit", zap.String("key", key))
		return value, nil
	}

	// 缓存未命中
	m.stats.MissCount++
	m.updateHitRates()
	m.logger.Debug("Cache miss", zap.String("key", key))
	return nil, nil
}

// Set 设置缓存值
func (m *ThreeLevelCacheManager) Set(ctx context.Context, key string, value interface{}, ttl time.Duration) error {
	// 记录访问频率
	m.tracker.RecordAccess(key)

	frequency := m.tracker.GetFrequency(key)
	cacheLevel := m.determineCacheLevel(frequency)

	switch cacheLevel {
	case L1_HOT:
		// 存储到L1和L2
		l1TTL := m.calculateL1TTL(frequency)
		l2TTL := m.calculateL2TTL(ttl)

		if err := m.l1Cache.Set(key, value, l1TTL); err != nil {
			m.logger.Error("Failed to set L1 cache", zap.String("key", key), zap.Error(err))
			return err
		}

		if err := m.l2Cache.Set(ctx, key, value, l2TTL); err != nil {
			m.logger.Error("Failed to set L2 cache", zap.String("key", key), zap.Error(err))
			return err
		}

	case L2_WARM:
		// 只存储到L2
		l2TTL := m.calculateL2TTL(ttl)
		if err := m.l2Cache.Set(ctx, key, value, l2TTL); err != nil {
			m.logger.Error("Failed to set L2 cache", zap.String("key", key), zap.Error(err))
			return err
		}

	case L3_COLD:
		// 不缓存到L1或L2，直接存储到数据库
		m.logger.Debug("Data marked as cold, not caching", zap.String("key", key))
	}

	m.logger.Debug("Cache set",
		zap.String("key", key),
		zap.String("level", cacheLevelString(cacheLevel)),
		zap.Duration("ttl", ttl),
	)

	return nil
}

// Delete 删除缓存
func (m *ThreeLevelCacheManager) Delete(ctx context.Context, key string) error {
	// 从所有缓存层删除
	if err := m.l1Cache.Delete(key); err != nil {
		m.logger.Warn("Failed to delete from L1 cache", zap.String("key", key), zap.Error(err))
	}

	if err := m.l2Cache.Delete(ctx, key); err != nil {
		m.logger.Warn("Failed to delete from L2 cache", zap.String("key", key), zap.Error(err))
	}

	m.logger.Debug("Cache deleted", zap.String("key", key))
	return nil
}

// Exists 检查键是否存在
func (m *ThreeLevelCacheManager) Exists(ctx context.Context, key string) (bool, error) {
	// 先检查L1
	if _, found := m.l1Cache.Get(key); found {
		return true, nil
	}

	// 再检查L2
	exists, err := m.l2Cache.Exists(ctx, key)
	if err != nil {
		return false, fmt.Errorf("L2 cache exists error: %w", err)
	}

	return exists, nil
}

// MGet 批量获取
func (m *ThreeLevelCacheManager) MGet(ctx context.Context, keys []string) (map[string]interface{}, error) {
	result := make(map[string]interface{})
	var missedKeys []string

	// 先从L1获取
	for _, key := range keys {
		if entry, found := m.l1Cache.Get(key); found {
			result[key] = entry.Value
			m.stats.L1HitCount++
		} else {
			missedKeys = append(missedKeys, key)
		}
	}

	// 从L2获取剩余的键
	for _, key := range missedKeys {
		if value, err := m.l2Cache.Get(ctx, key); err == nil && value != nil {
			result[key] = value
			m.stats.L2HitCount++

			// 尝试提升到L1
			frequency := m.tracker.GetFrequency(key)
			if m.shouldPromoteToL1(frequency) {
				ttl := m.calculateL1TTL(frequency)
				m.l1Cache.Set(key, value, ttl)
			}
		} else {
			m.stats.MissCount++
		}
	}

	m.stats.TotalRequests += int64(len(keys))
	m.updateHitRates()

	return result, nil
}

// MSet 批量设置
func (m *ThreeLevelCacheManager) MSet(ctx context.Context, entries map[string]interface{}, ttl time.Duration) error {
	for key, value := range entries {
		if err := m.Set(ctx, key, value, ttl); err != nil {
			return fmt.Errorf("failed to set key %s: %w", key, err)
		}
	}
	return nil
}

// MDelete 批量删除
func (m *ThreeLevelCacheManager) MDelete(ctx context.Context, keys []string) error {
	for _, key := range keys {
		if err := m.Delete(ctx, key); err != nil {
			m.logger.Warn("Failed to delete key", zap.String("key", key), zap.Error(err))
		}
	}
	return nil
}

// Clear 清空指定级别的缓存
func (m *ThreeLevelCacheManager) Clear(ctx context.Context, level CacheLevel) error {
	switch level {
	case L1_HOT:
		return m.l1Cache.Clear()
	case L2_WARM:
		return m.l2Cache.Clear(ctx)
	default:
		// 清空所有
		if err := m.l1Cache.Clear(); err != nil {
			return err
		}
		return m.l2Cache.Clear(ctx)
	}
}

// GetStats 获取缓存统计
func (m *ThreeLevelCacheManager) GetStats(ctx context.Context) (*CacheStats, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	// 更新条目数统计
	if l1Stats := m.l1Cache.Stats(); l1Stats != nil {
		m.stats.L1Entries = l1Stats.Entries
	}

	m.stats.TotalEntries = m.stats.L1Entries + m.stats.L2Entries

	return m.stats, nil
}

// Preload 预热缓存
func (m *ThreeLevelCacheManager) Preload(ctx context.Context, keys []string) error {
	return m.preloadKeys(ctx, keys)
}

// UpdateStrategy 更新缓存策略
func (m *ThreeLevelCacheManager) UpdateStrategy(strategy *CacheStrategy) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.strategy = strategy
	m.logger.Info("Cache strategy updated")
	return nil
}

// GetStrategy 获取当前策略
func (m *ThreeLevelCacheManager) GetStrategy() *CacheStrategy {
	m.mu.RLock()
	defer m.mu.RUnlock()

	return m.strategy
}

// 辅助方法

// determineCacheLevel 根据访问频率确定缓存级别
func (m *ThreeLevelCacheManager) determineCacheLevel(frequency float64) CacheLevel {
	if frequency >= m.strategy.HotDataThreshold {
		return L1_HOT
	} else if frequency >= m.strategy.WarmDataThreshold {
		return L2_WARM
	}
	return L3_COLD
}

// shouldPromoteToL1 判断是否应该提升到L1缓存
func (m *ThreeLevelCacheManager) shouldPromoteToL1(frequency float64) bool {
	return frequency >= m.strategy.HotDataThreshold
}

// calculateL1TTL 计算L1缓存TTL
func (m *ThreeLevelCacheManager) calculateL1TTL(frequency float64) time.Duration {
	// 频率越高，TTL越长
	baseTTL := m.strategy.L1TTL
	if frequency > m.strategy.HotDataThreshold*2 {
		return baseTTL * 2
	}
	return baseTTL
}

// calculateL2TTL 计算L2缓存TTL
func (m *ThreeLevelCacheManager) calculateL2TTL(originalTTL time.Duration) time.Duration {
	if originalTTL > 0 {
		return originalTTL
	}
	return m.strategy.L2TTL
}

// updateHitRates 更新命中率
func (m *ThreeLevelCacheManager) updateHitRates() {
	total := m.stats.L1HitCount + m.stats.L2HitCount + m.stats.MissCount
	if total > 0 {
		m.stats.L1HitRate = float64(m.stats.L1HitCount) / float64(total)
		m.stats.L2HitRate = float64(m.stats.L2HitCount) / float64(total)
		m.stats.TotalHitRate = float64(m.stats.L1HitCount+m.stats.L2HitCount) / float64(total)
	}
}

// updateAvgResponseTime 更新平均响应时间
func (m *ThreeLevelCacheManager) updateAvgResponseTime(duration time.Duration) {
	if m.stats.TotalRequests == 0 {
		m.stats.AvgResponseTime = duration
		return
	}

	// 简单的移动平均
	m.stats.AvgResponseTime = time.Duration(
		(int64(m.stats.AvgResponseTime) + int64(duration)) / 2,
	)
}

// preloadKeys 预热键
func (m *ThreeLevelCacheManager) preloadKeys(ctx context.Context, keys []string) error {
	m.logger.Info("Starting cache preload", zap.Int("keys_count", len(keys)))

	// 分批并发预热
	semaphore := make(chan struct{}, m.strategy.PreloadConcurrency)
	var wg sync.WaitGroup

	for _, key := range keys {
		wg.Add(1)
		go func(k string) {
			defer wg.Done()
			semaphore <- struct{}{}
			defer func() { <-semaphore }()

			// 尝试从数据源加载并缓存
			if value := m.loadFromDataSource(ctx, k); value != nil {
				m.Set(ctx, k, value, m.strategy.L2TTL)
			}
		}(key)
	}

	wg.Wait()
	m.logger.Info("Cache preload completed")
	return nil
}

// loadFromDataSource 从数据源加载数据（需要实现）
func (m *ThreeLevelCacheManager) loadFromDataSource(ctx context.Context, key string) interface{} {
	// 这里应该根据具体业务实现数据加载逻辑
	// 例如：从数据库、API或其他服务加载
	m.logger.Debug("Loading data from source", zap.String("key", key))
	return nil
}

// backgroundTasks 后台任务
func (m *ThreeLevelCacheManager) backgroundTasks() {
	ticker := time.NewTicker(1 * time.Minute)
	defer ticker.Stop()

	for {
		select {
		case <-m.ctx.Done():
			return
		case <-ticker.C:
			// 定期清理和统计
			m.tracker.CleanupExpiredWindow()
			if l1Cache, ok := m.l1Cache.(interface{ Cleanup() error }); ok {
				l1Cache.Cleanup()
			}
		}
	}
}

// cacheLevelString 缓存级别字符串
func cacheLevelString(level CacheLevel) string {
	switch level {
	case L1_HOT:
		return "L1_HOT"
	case L2_WARM:
		return "L2_WARM"
	case L3_COLD:
		return "L3_COLD"
	default:
		return "UNKNOWN"
	}
}
