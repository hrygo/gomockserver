package cache

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

// MockL1Cache 模拟L1缓存
type MockL1Cache struct {
	data map[string]*CacheEntry
}

func NewMockL1Cache() *MockL1Cache {
	return &MockL1Cache{
		data: make(map[string]*CacheEntry),
	}
}

func (m *MockL1Cache) Get(key string) (*CacheEntry, bool) {
	entry, exists := m.data[key]
	return entry, exists
}

func (m *MockL1Cache) Set(key string, value interface{}, ttl time.Duration) error {
	entry := &CacheEntry{
		Key:   key,
		Value: value,
		Level: L1_HOT,
		TTL:   ttl,
	}
	m.data[key] = entry
	return nil
}

func (m *MockL1Cache) Delete(key string) error {
	delete(m.data, key)
	return nil
}

func (m *MockL1Cache) Clear() error {
	m.data = make(map[string]*CacheEntry)
	return nil
}

func (m *MockL1Cache) Stats() *L1Stats {
	return &L1Stats{
		Entries: int64(len(m.data)),
	}
}

func (m *MockL1Cache) Cleanup() error {
	return nil
}

// MockL2Cache 模拟L2缓存
type MockL2Cache struct {
	data map[string]interface{}
}

func NewMockL2Cache() *MockL2Cache {
	return &MockL2Cache{
		data: make(map[string]interface{}),
	}
}

func (m *MockL2Cache) Get(ctx interface{}, key string) (interface{}, error) {
	value, exists := m.data[key]
	if !exists {
		return nil, ErrCacheMiss
	}
	return value, nil
}

func (m *MockL2Cache) Set(ctx interface{}, key string, value interface{}, ttl time.Duration) error {
	m.data[key] = value
	return nil
}

func (m *MockL2Cache) Delete(ctx interface{}, key string) error {
	delete(m.data, key)
	return nil
}

func (m *MockL2Cache) Exists(ctx interface{}, key string) (bool, error) {
	_, exists := m.data[key]
	return exists, nil
}

func (m *MockL2Cache) Clear(ctx interface{}) error {
	m.data = make(map[string]interface{})
	return nil
}

func (m *MockL2Cache) Ping(ctx interface{}) error {
	return nil
}

// MockFrequencyTracker 模拟频率跟踪器
type MockFrequencyTracker struct {
	frequencies map[string]float64
}

func NewMockFrequencyTracker() *MockFrequencyTracker {
	return &MockFrequencyTracker{
		frequencies: make(map[string]float64),
	}
}

func (m *MockFrequencyTracker) RecordAccess(key string) {
	m.frequencies[key]++
}

func (m *MockFrequencyTracker) GetFrequency(key string) float64 {
	return m.frequencies[key]
}

func (m *MockFrequencyTracker) CleanupExpiredWindow() {
	// Mock implementation
}

func (m *MockFrequencyTracker) GetTopKeys(limit int) []string {
	// Mock implementation - 返回固定键
	keys := make([]string, 0, limit)
	for i := 0; i < limit && i < len(m.frequencies); i++ {
		for key := range m.frequencies {
			keys = append(keys, key)
			if len(keys) >= limit {
				break
			}
		}
	}
	return keys
}

// MockThreeLevelCacheManager 模拟三级缓存管理器
type MockThreeLevelCacheManager struct {
	l1Cache  *MockL1Cache
	l2Cache  *MockL2Cache
	tracker  *MockFrequencyTracker
	strategy *CacheStrategy
	stats    *CacheStats
}

func NewMockThreeLevelCacheManager() *MockThreeLevelCacheManager {
	return &MockThreeLevelCacheManager{
		l1Cache:  NewMockL1Cache(),
		l2Cache:  NewMockL2Cache(),
		tracker:  NewMockFrequencyTracker(),
		strategy: DefaultCacheStrategy(),
		stats:    &CacheStats{},
	}
}

func (m *MockThreeLevelCacheManager) Get(ctx interface{}, key string) (interface{}, error) {
	// 先尝试L1缓存
	entry, found := m.l1Cache.Get(key)
	if found {
		m.stats.L1HitCount++
		m.stats.TotalRequests++
		m.tracker.RecordAccess(key)
		return entry.Value, nil
	}

	// 再尝试L2缓存
	value, err := m.l2Cache.Get(ctx, key)
	if err == nil {
		m.stats.L2HitCount++
		m.stats.TotalRequests++
		m.tracker.RecordAccess(key)
		// 将数据提升到L1缓存
		m.l1Cache.Set(key, value, m.strategy.L1TTL)
		return value, nil
	}

	// 缓存未命中
	m.stats.MissCount++
	m.stats.TotalRequests++
	return nil, ErrCacheMiss
}

func (m *MockThreeLevelCacheManager) Set(ctx interface{}, key string, value interface{}, ttl time.Duration) error {
	// 直接设置到L2缓存
	err := m.l2Cache.Set(ctx, key, value, ttl)
	if err != nil {
		return err
	}

	// 根据策略决定是否也存储到L1缓存
	m.l1Cache.Set(key, value, ttl)
	return nil
}

func (m *MockThreeLevelCacheManager) Delete(ctx interface{}, key string) error {
	m.l1Cache.Delete(key)
	m.l2Cache.Delete(ctx, key)
	return nil
}

func (m *MockThreeLevelCacheManager) Exists(ctx interface{}, key string) (bool, error) {
	// 检查L1缓存
	_, found := m.l1Cache.Get(key)
	if found {
		return true, nil
	}

	// 检查L2缓存
	exists, err := m.l2Cache.Exists(ctx, key)
	if err != nil {
		return false, err
	}
	return exists, nil
}

func (m *MockThreeLevelCacheManager) Clear(ctx interface{}, level CacheLevel) error {
	switch level {
	case L1_HOT:
		return m.l1Cache.Clear()
	case L2_WARM:
		return m.l2Cache.Clear(ctx)
	case L3_COLD:
		// Mock implementation for L3
		return nil
	default:
		return m.l1Cache.Clear()
	}
}

func (m *MockThreeLevelCacheManager) GetStats(ctx interface{}) (*CacheStats, error) {
	// 计算命中率
	total := m.stats.TotalRequests
	if total > 0 {
		m.stats.L1HitRate = float64(m.stats.L1HitCount) / float64(total)
		m.stats.L2HitRate = float64(m.stats.L2HitCount) / float64(total)
		m.stats.TotalHitRate = float64(m.stats.L1HitCount+m.stats.L2HitCount) / float64(total)
	}

	return m.stats, nil
}

func (m *MockThreeLevelCacheManager) Preload(ctx interface{}, keys []string) error {
	// Mock implementation
	return nil
}

func (m *MockThreeLevelCacheManager) UpdateStrategy(strategy *CacheStrategy) error {
	m.strategy = strategy
	return nil
}

func (m *MockThreeLevelCacheManager) GetStrategy() *CacheStrategy {
	return m.strategy
}

func (m *MockThreeLevelCacheManager) Start(ctx interface{}) error {
	return nil
}

func (m *MockThreeLevelCacheManager) Stop(ctx interface{}) error {
	return nil
}

// MGet 批量获取
func (m *MockThreeLevelCacheManager) MGet(ctx interface{}, keys []string) (map[string]interface{}, error) {
	result := make(map[string]interface{})
	for _, key := range keys {
		value, err := m.Get(ctx, key)
		if err == nil {
			result[key] = value
		}
	}
	return result, nil
}

// MSet 批量设置
func (m *MockThreeLevelCacheManager) MSet(ctx interface{}, entries map[string]interface{}, ttl time.Duration) error {
	for key, value := range entries {
		err := m.Set(ctx, key, value, ttl)
		if err != nil {
			return err
		}
	}
	return nil
}

// MDelete 批量删除
func (m *MockThreeLevelCacheManager) MDelete(ctx interface{}, keys []string) error {
	for _, key := range keys {
		err := m.Delete(ctx, key)
		if err != nil {
			return err
		}
	}
	return nil
}

// TestThreeLevelCacheManager_BasicOperations 测试三级缓存管理器基本操作
func TestThreeLevelCacheManager_BasicOperations(t *testing.T) {
	manager := NewMockThreeLevelCacheManager()
	ctx := context.Background()

	// 测试Set和Get
	t.Run("Set and Get", func(t *testing.T) {
		key := "test_key"
		value := "test_value"

		err := manager.Set(ctx, key, value, 10*time.Minute)
		assert.NoError(t, err)

		retrievedValue, err := manager.Get(ctx, key)
		assert.NoError(t, err)
		assert.Equal(t, value, retrievedValue)
	})

	// 测试缓存未命中
	t.Run("Cache miss", func(t *testing.T) {
		_, err := manager.Get(ctx, "non_existent_key")
		assert.Error(t, err)
		assert.Equal(t, ErrCacheMiss, err)
	})

	// 测试Exists
	t.Run("Exists", func(t *testing.T) {
		key := "exists_key"
		value := "exists_value"

		err := manager.Set(ctx, key, value, 10*time.Minute)
		assert.NoError(t, err)

		exists, err := manager.Exists(ctx, key)
		assert.NoError(t, err)
		assert.True(t, exists)

		exists, err = manager.Exists(ctx, "non_existent_key")
		assert.NoError(t, err)
		assert.False(t, exists)
	})

	// 测试Delete
	t.Run("Delete", func(t *testing.T) {
		key := "delete_key"
		value := "delete_value"

		err := manager.Set(ctx, key, value, 10*time.Minute)
		assert.NoError(t, err)

		err = manager.Delete(ctx, key)
		assert.NoError(t, err)

		exists, err := manager.Exists(ctx, key)
		assert.NoError(t, err)
		assert.False(t, exists)
	})
}

// TestThreeLevelCacheManager_Statistics 测试统计功能
func TestThreeLevelCacheManager_Statistics(t *testing.T) {
	manager := NewMockThreeLevelCacheManager()
	ctx := context.Background()

	// 添加一些数据
	for i := 0; i < 10; i++ {
		key := "key_" + string(rune('0'+i))
		value := "value_" + string(rune('0'+i))
		manager.Set(ctx, key, value, 10*time.Minute)
	}

	// 获取一些数据以产生L2命中（因为数据在L2缓存中），然后提升到L1
	for i := 0; i < 5; i++ {
		key := "key_" + string(rune('0'+i))
		manager.Get(ctx, key) // L2命中，提升到L1
	}

	// 再次获取相同的数据以产生L1命中
	for i := 0; i < 3; i++ {
		key := "key_" + string(rune('0'+i))
		manager.Get(ctx, key) // L1命中
	}

	// 获取一些不存在的数据以产生未命中
	for i := 10; i < 15; i++ {
		key := "non_existent_" + string(rune('0'+i))
		manager.Get(ctx, key)
	}

	// 检查统计信息
	stats, err := manager.GetStats(ctx)
	assert.NoError(t, err)
	assert.Equal(t, int64(13), stats.TotalRequests) // 5次L2命中 + 8次L1命中（5次提升+3次直接命中） + 5次未命中
	assert.Equal(t, int64(8), stats.L1HitCount)
	assert.Equal(t, int64(0), stats.L2HitCount) // 提升到L1后，第二次访问直接从L1命中
	assert.Equal(t, int64(5), stats.MissCount)
	assert.Equal(t, 8.0/13.0, stats.L1HitRate)
	assert.Equal(t, 8.0/13.0, stats.TotalHitRate)
}

// TestThreeLevelCacheManager_Strategy 测试策略管理
func TestThreeLevelCacheManager_Strategy(t *testing.T) {
	manager := NewMockThreeLevelCacheManager()

	// 获取默认策略
	strategy := manager.GetStrategy()
	assert.NotNil(t, strategy)
	assert.Equal(t, 0.8, strategy.HotDataThreshold)
	assert.Equal(t, 0.2, strategy.WarmDataThreshold)

	// 更新策略
	newStrategy := &CacheStrategy{
		L1MaxEntries:      20000,
		L1TTL:             2 * time.Minute,
		L2TTL:             20 * time.Minute,
		HotDataThreshold:  0.9,
		WarmDataThreshold: 0.3,
		PreloadEnabled:    false,
	}

	err := manager.UpdateStrategy(newStrategy)
	assert.NoError(t, err)

	// 验证策略已更新
	updatedStrategy := manager.GetStrategy()
	assert.Equal(t, 20000, updatedStrategy.L1MaxEntries)
	assert.Equal(t, 2*time.Minute, updatedStrategy.L1TTL)
	assert.Equal(t, 20*time.Minute, updatedStrategy.L2TTL)
	assert.Equal(t, 0.9, updatedStrategy.HotDataThreshold)
	assert.Equal(t, 0.3, updatedStrategy.WarmDataThreshold)
	assert.False(t, updatedStrategy.PreloadEnabled)
}

// TestThreeLevelCacheManager_BatchOperations 测试批量操作
func TestThreeLevelCacheManager_BatchOperations(t *testing.T) {
	manager := NewMockThreeLevelCacheManager()
	ctx := context.Background()

	// 测试MSet
	t.Run("MSet", func(t *testing.T) {
		entries := map[string]interface{}{
			"key1": "value1",
			"key2": "value2",
			"key3": "value3",
		}

		err := manager.MSet(ctx, entries, 10*time.Minute)
		assert.NoError(t, err)

		// 验证所有值都已设置
		for key, expectedValue := range entries {
			actualValue, err := manager.Get(ctx, key)
			assert.NoError(t, err)
			assert.Equal(t, expectedValue, actualValue)
		}
	})

	// 测试MGet
	t.Run("MGet", func(t *testing.T) {
		keys := []string{"key1", "key2", "key4"}
		result, err := manager.MGet(ctx, keys)
		assert.NoError(t, err)

		assert.Equal(t, "value1", result["key1"])
		assert.Equal(t, "value2", result["key2"])
		assert.NotContains(t, result, "key4") // 不存在的键
	})

	// 测试MDelete
	t.Run("MDelete", func(t *testing.T) {
		keys := []string{"key1", "key3"}

		err := manager.MDelete(ctx, keys)
		assert.NoError(t, err)

		// 验证键已被删除
		exists1, _ := manager.Exists(ctx, "key1")
		assert.False(t, exists1)
		exists3, _ := manager.Exists(ctx, "key3")
		assert.False(t, exists3)

		// 未删除的键应该还存在
		exists2, _ := manager.Exists(ctx, "key2")
		assert.True(t, exists2)
	})
}

// TestThreeLevelCacheManager_Lifecycle 测试生命周期管理
func TestThreeLevelCacheManager_Lifecycle(t *testing.T) {
	manager := NewMockThreeLevelCacheManager()
	ctx := context.Background()

	// 测试启动
	err := manager.Start(ctx)
	assert.NoError(t, err)

	// 执行一些操作
	err = manager.Set(ctx, "test", "value", 1*time.Minute)
	assert.NoError(t, err)

	value, err := manager.Get(ctx, "test")
	assert.NoError(t, err)
	assert.Equal(t, "value", value)

	// 测试停止
	err = manager.Stop(ctx)
	assert.NoError(t, err)
}
