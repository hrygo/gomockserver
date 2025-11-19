package cache

import (
	"container/list"
	"sync"
	"time"

	"go.uber.org/zap"
)

// MemoryL1Cache L1内存缓存实现
type MemoryL1Cache struct {
	maxEntries int
	capacity   int64
	size       int64
	mu         sync.RWMutex
	data       map[string]*list.Element
	lru        *list.List
	stats      *L1Stats
	logger     *zap.Logger
	cleanupTTL time.Duration
	stopCh     chan struct{}
}

// lruItem LRU缓存项
type lruItem struct {
	key       string
	value     interface{}
	entry     *CacheEntry
	size      int64
	expiresAt time.Time
}

// NewMemoryL1Cache 创建内存L1缓存
func NewMemoryL1Cache(maxEntries int, maxMemoryMB int, cleanupTTL time.Duration, logger *zap.Logger) *MemoryL1Cache {
	if maxEntries <= 0 {
		maxEntries = 10000
	}
	if maxMemoryMB <= 0 {
		maxMemoryMB = 100 // 默认100MB
	}
	if cleanupTTL <= 0 {
		cleanupTTL = 5 * time.Minute
	}

	cache := &MemoryL1Cache{
		maxEntries: maxEntries,
		capacity:   int64(maxMemoryMB) * 1024 * 1024, // 转换为字节
		data:       make(map[string]*list.Element),
		lru:        list.New(),
		stats:      &L1Stats{},
		logger:     logger,
		cleanupTTL: cleanupTTL,
		stopCh:     make(chan struct{}),
	}

	// 启动定期清理过期条目
	go cache.startCleanup()

	logger.Info("L1 memory cache initialized",
		zap.Int("max_entries", maxEntries),
		zap.Int64("capacity_mb", int64(maxMemoryMB)),
		zap.Duration("cleanup_ttl", cleanupTTL),
	)

	return cache
}

// Get 获取缓存值
func (m *MemoryL1Cache) Get(key string) (*CacheEntry, bool) {
	m.mu.Lock()
	defer m.mu.Unlock()

	element, exists := m.data[key]
	if !exists {
		m.stats.Misses++
		return nil, false
	}

	item := element.Value.(*lruItem)

	// 检查是否过期
	if !item.expiresAt.IsZero() && time.Now().After(item.expiresAt) {
		m.removeElement(element)
		m.stats.Misses++
		return nil, false
	}

	// 更新LRU位置
	m.lru.MoveToFront(element)

	// 更新访问信息
	item.entry.AccessAt = time.Now()
	item.entry.HitCount++

	m.stats.Hits++
	m.updateHitRate()

	// 创建返回的缓存条目副本
	entryCopy := *item.entry
	return &entryCopy, true
}

// Set 设置缓存值
func (m *MemoryL1Cache) Set(key string, value interface{}, ttl time.Duration) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	// 计算值大小（估算）
	valueSize := m.estimateSize(value)

	// 检查容量限制
	for (m.size+valueSize > m.capacity || len(m.data) >= m.maxEntries) && len(m.data) > 0 {
		m.evictLRU()
	}

	// 创建缓存条目
	now := time.Now()
	var expiresAt time.Time
	if ttl > 0 {
		expiresAt = now.Add(ttl)
	}

	entry := &CacheEntry{
		Key:       key,
		Value:     value,
		Level:     L1_HOT,
		TTL:       ttl,
		CreatedAt: now,
		AccessAt:  now,
		HitCount:  0,
		ExpireAt:  expiresAt,
	}

	item := &lruItem{
		key:       key,
		value:     value,
		entry:     entry,
		size:      valueSize,
		expiresAt: expiresAt,
	}

	// 如果键已存在，先删除旧的
	if element, exists := m.data[key]; exists {
		m.removeElement(element)
	}

	// 添加新项到前面
	element := m.lru.PushFront(item)
	m.data[key] = element
	m.size += valueSize

	m.logger.Debug("L1 cache set",
		zap.String("key", key),
		zap.Int64("size", valueSize),
		zap.Duration("ttl", ttl),
		zap.Int("total_entries", len(m.data)),
		zap.Int64("total_size_mb", m.size/(1024*1024)),
	)

	return nil
}

// Delete 删除缓存
func (m *MemoryL1Cache) Delete(key string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if element, exists := m.data[key]; exists {
		m.removeElement(element)
		delete(m.data, key)
		m.logger.Debug("L1 cache deleted", zap.String("key", key))
	}

	return nil
}

// Clear 清空所有缓存
func (m *MemoryL1Cache) Clear() error {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.data = make(map[string]*list.Element)
	m.lru.Init()
	m.size = 0

	m.logger.Info("L1 cache cleared")
	return nil
}

// Stats 获取统计信息
func (m *MemoryL1Cache) Stats() *L1Stats {
	m.mu.RLock()
	defer m.mu.RUnlock()

	statsCopy := *m.stats
	statsCopy.Entries = int64(len(m.data))
	return &statsCopy
}

// Cleanup 清理过期条目
func (m *MemoryL1Cache) Cleanup() error {
	m.mu.Lock()
	defer m.mu.Unlock()

	now := time.Now()
	var removed int

	for key, element := range m.data {
		item := element.Value.(*lruItem)
		if !item.expiresAt.IsZero() && now.After(item.expiresAt) {
			m.removeElement(element)
			delete(m.data, key)
			removed++
		}
	}

	if removed > 0 {
		m.logger.Debug("L1 cache cleanup completed",
			zap.Int("expired_entries", removed),
			zap.Int("remaining_entries", len(m.data)),
		)
	}

	return nil
}

// evictLRU 淘汰最久未使用的项
func (m *MemoryL1Cache) evictLRU() {
	if m.lru.Len() == 0 {
		return
	}

	element := m.lru.Back()
	if element != nil {
		item := element.Value.(*lruItem)
		m.removeElement(element)
		delete(m.data, item.key)

		m.stats.Evictions++
		m.logger.Debug("L1 cache evicted LRU item",
			zap.String("key", item.key),
			zap.Int64("size", item.size),
		)
	}
}

// removeElement 移除元素
func (m *MemoryL1Cache) removeElement(element *list.Element) {
	item := element.Value.(*lruItem)
	m.size -= item.size
	m.lru.Remove(element)
}

// updateHitRate 更新命中率
func (m *MemoryL1Cache) updateHitRate() {
	total := m.stats.Hits + m.stats.Misses
	if total > 0 {
		m.stats.HitRate = float64(m.stats.Hits) / float64(total)
	}
}

// estimateSize 估算值大小（简单实现）
func (m *MemoryL1Cache) estimateSize(value interface{}) int64 {
	const baseSize = 64 // 基础开销

	switch v := value.(type) {
	case string:
		return int64(len(v) + baseSize)
	case []byte:
		return int64(len(v) + baseSize)
	case int, int8, int16, int32, int64:
		return 8 + baseSize
	case float32, float64:
		return 8 + baseSize
	case bool:
		return 1 + baseSize
	default:
		// 对于复杂类型，使用较大的估算值
		return 1024 + baseSize
	}
}

// startCleanup 启动定期清理
func (m *MemoryL1Cache) startCleanup() {
	ticker := time.NewTicker(m.cleanupTTL)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			m.Cleanup()
		case <-m.stopCh:
			return
		}
	}
}

// Stop 停止缓存
func (m *MemoryL1Cache) Stop() {
	close(m.stopCh)
	m.Clear()
	m.logger.Info("L1 memory cache stopped")
}

// GetTopKeys 获取热点键
func (m *MemoryL1Cache) GetTopKeys(limit int) []string {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if limit <= 0 || limit > len(m.data) {
		limit = len(m.data)
	}

	type keyHitCount struct {
		key  string
		hits int64
	}

	var topKeys []keyHitCount

	// 收集所有键的命中次数
	for element := m.lru.Front(); element != nil; element = element.Next() {
		item := element.Value.(*lruItem)
		topKeys = append(topKeys, keyHitCount{
			key:  item.key,
			hits: item.entry.HitCount,
		})
	}

	// 简单排序（按命中次数）
	for i := 0; i < len(topKeys)-1; i++ {
		for j := i + 1; j < len(topKeys); j++ {
			if topKeys[i].hits < topKeys[j].hits {
				topKeys[i], topKeys[j] = topKeys[j], topKeys[i]
			}
		}
	}

	// 返回前N个键名
	result := make([]string, 0, limit)
	for i := 0; i < limit && i < len(topKeys); i++ {
		result = append(result, topKeys[i].key)
	}

	return result
}

// GetMemoryUsage 获取内存使用情况
func (m *MemoryL1Cache) GetMemoryUsage() map[string]interface{} {
	m.mu.RLock()
	defer m.mu.RUnlock()

	return map[string]interface{}{
		"used_bytes":     m.size,
		"capacity_bytes": m.capacity,
		"used_mb":        m.size / (1024 * 1024),
		"capacity_mb":    m.capacity / (1024 * 1024),
		"usage_percent":  float64(m.size) / float64(m.capacity) * 100,
		"entries":        len(m.data),
		"max_entries":    m.maxEntries,
	}
}