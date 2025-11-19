package cache

import (
	"context"
	"time"
)

// CacheLevel 缓存级别枚举
type CacheLevel int

const (
	L1_HOT  CacheLevel = iota // L1缓存：内存中的热点数据
	L2_WARM                   // L2缓存：Redis中的温数据
	L3_COLD                   // L3缓存：MongoDB中的冷数据
)

// CacheEntry 缓存条目
type CacheEntry struct {
	Key       string        `json:"key"`
	Value     interface{}   `json:"value"`
	Level     CacheLevel    `json:"level"`
	TTL       time.Duration `json:"ttl"`
	CreatedAt time.Time     `json:"created_at"`
	AccessAt  time.Time     `json:"access_at"`
	HitCount  int64         `json:"hit_count"`
	ExpireAt  time.Time     `json:"expire_at"`
}

// CacheStats 缓存统计信息
type CacheStats struct {
	TotalRequests   int64         `json:"total_requests"`
	L1HitCount      int64         `json:"l1_hit_count"`
	L2HitCount      int64         `json:"l2_hit_count"`
	L3HitCount      int64         `json:"l3_hit_count"`
	MissCount       int64         `json:"miss_count"`
	L1HitRate       float64       `json:"l1_hit_rate"`
	L2HitRate       float64       `json:"l2_hit_rate"`
	TotalHitRate    float64       `json:"total_hit_rate"`
	AvgResponseTime time.Duration `json:"avg_response_time"`
	TotalEntries    int64         `json:"total_entries"`
	L1Entries       int64         `json:"l1_entries"`
	L2Entries       int64         `json:"l2_entries"`
}

// CacheStrategy 缓存策略配置
type CacheStrategy struct {
	// L1缓存配置
	L1MaxEntries      int           `json:"l1_max_entries"`
	L1TTL             time.Duration `json:"l1_ttl"`
	L1CleanupInterval time.Duration `json:"l1_cleanup_interval"`

	// L2缓存配置
	L2TTL             time.Duration `json:"l2_ttl"`
	L2CleanupInterval time.Duration `json:"l2_cleanup_interval"`

	// 策略配置
	HotDataThreshold  float64       `json:"hot_data_threshold"`  // 热点数据访问频率阈值
	WarmDataThreshold float64       `json:"warm_data_threshold"` // 温数据访问频率阈值
	PreloadEnabled    bool          `json:"preload_enabled"`     // 是否启用预热
	AccessFreqWindow  time.Duration `json:"access_freq_window"`  // 访问频率统计窗口

	// 预热配置
	PreloadKeys        []string `json:"preload_keys"`        // 预热键列表
	PreloadConcurrency int      `json:"preload_concurrency"` // 预热并发数
}

// DefaultCacheStrategy 返回默认缓存策略
func DefaultCacheStrategy() *CacheStrategy {
	return &CacheStrategy{
		L1MaxEntries:       10000,
		L1TTL:              1 * time.Minute,
		L1CleanupInterval:  5 * time.Minute,
		L2TTL:              10 * time.Minute,
		L2CleanupInterval:  30 * time.Minute,
		HotDataThreshold:   0.8,
		WarmDataThreshold:  0.2,
		PreloadEnabled:     true,
		AccessFreqWindow:   1 * time.Hour,
		PreloadConcurrency: 10,
	}
}

// Manager 缓存管理器接口
type Manager interface {
	// 基础操作
	Get(ctx context.Context, key string) (interface{}, error)
	Set(ctx context.Context, key string, value interface{}, ttl time.Duration) error
	Delete(ctx context.Context, key string) error
	Exists(ctx context.Context, key string) (bool, error)

	// 批量操作
	MGet(ctx context.Context, keys []string) (map[string]interface{}, error)
	MSet(ctx context.Context, entries map[string]interface{}, ttl time.Duration) error
	MDelete(ctx context.Context, keys []string) error

	// 缓存控制
	Clear(ctx context.Context, level CacheLevel) error
	GetStats(ctx context.Context) (*CacheStats, error)
	Preload(ctx context.Context, keys []string) error

	// 策略管理
	UpdateStrategy(strategy *CacheStrategy) error
	GetStrategy() *CacheStrategy

	// 生命周期
	Start(ctx context.Context) error
	Stop(ctx context.Context) error
}

// L1Cache L1内存缓存接口
type L1Cache interface {
	Get(key string) (*CacheEntry, bool)
	Set(key string, value interface{}, ttl time.Duration) error
	Delete(key string) error
	Clear() error
	Stats() *L1Stats
	Cleanup() error
}

// L1Stats L1缓存统计
type L1Stats struct {
	Entries   int64   `json:"entries"`
	Hits      int64   `json:"hits"`
	Misses    int64   `json:"misses"`
	Evictions int64   `json:"evictions"`
	HitRate   float64 `json:"hit_rate"`
}

// L2Cache L2 Redis缓存接口
type L2Cache interface {
	Get(ctx context.Context, key string) (interface{}, error)
	Set(ctx context.Context, key string, value interface{}, ttl time.Duration) error
	Delete(ctx context.Context, key string) error
	Exists(ctx context.Context, key string) (bool, error)
	Clear(ctx context.Context) error
	Ping(ctx context.Context) error
}

// AccessFrequency 访问频率记录
type AccessFrequency struct {
	Key       string    `json:"key"`
	Count     int64     `json:"count"`
	Window    time.Time `json:"window"`
	Frequency float64   `json:"frequency"`
}

// FrequencyTracker 访问频率跟踪器接口
type FrequencyTracker interface {
	RecordAccess(key string)
	GetFrequency(key string) float64
	CleanupExpiredWindow()
	GetTopKeys(limit int) []string
}
