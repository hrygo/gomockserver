package cache

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
)

// RedisL2Cache Redis L2缓存实现
type RedisL2Cache struct {
	client  *redis.Client
	prefix  string
	logger  *zap.Logger
	metrics *L2Metrics
}

// L2Metrics L2缓存指标
type L2Metrics struct {
	Commands    int64         `json:"commands"`
	Hits        int64         `json:"hits"`
	Misses      int64         `json:"misses"`
	Errors      int64         `json:"errors"`
	AvgLatency  time.Duration `json:"avg_latency"`
	LastLatency time.Duration `json:"last_latency"`
}

// RedisConfig Redis配置
type RedisConfig struct {
	Host         string        `json:"host"`
	Port         int           `json:"port"`
	Password     string        `json:"password"`
	Database     int           `json:"database"`
	PoolSize     int           `json:"pool_size"`
	MinIdleConns int           `json:"min_idle_conns"`
	DialTimeout  time.Duration `json:"dial_timeout"`
	ReadTimeout  time.Duration `json:"read_timeout"`
	WriteTimeout time.Duration `json:"write_timeout"`
	PoolTimeout  time.Duration `json:"pool_timeout"`
	KeyPrefix    string        `json:"key_prefix"`
}

// DefaultRedisConfig 返回默认Redis配置
func DefaultRedisConfig() *RedisConfig {
	return &RedisConfig{
		Host:         "localhost",
		Port:         6379,
		Password:     "",
		Database:     0,
		PoolSize:     20,
		MinIdleConns: 5,
		DialTimeout:  5 * time.Second,
		ReadTimeout:  3 * time.Second,
		WriteTimeout: 3 * time.Second,
		PoolTimeout:  4 * time.Second,
		KeyPrefix:    "mockserver:cache:",
	}
}

// NewRedisL2Cache 创建Redis L2缓存
func NewRedisL2Cache(config *RedisConfig, logger *zap.Logger) (*RedisL2Cache, error) {
	if config == nil {
		config = DefaultRedisConfig()
	}

	rdb := redis.NewClient(&redis.Options{
		Addr:         fmt.Sprintf("%s:%d", config.Host, config.Port),
		Password:     config.Password,
		DB:           config.Database,
		PoolSize:     config.PoolSize,
		MinIdleConns: config.MinIdleConns,
		DialTimeout:  config.DialTimeout,
		ReadTimeout:  config.ReadTimeout,
		WriteTimeout: config.WriteTimeout,
		PoolTimeout:  config.PoolTimeout,
	})

	// 测试连接
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := rdb.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("failed to connect to Redis: %w", err)
	}

	cache := &RedisL2Cache{
		client:  rdb,
		prefix:  config.KeyPrefix,
		logger:  logger,
		metrics: &L2Metrics{},
	}

	logger.Info("Redis L2 cache initialized successfully",
		zap.String("addr", rdb.Options().Addr),
		zap.String("prefix", config.KeyPrefix),
	)

	return cache, nil
}

// Get 获取缓存值
func (r *RedisL2Cache) Get(ctx context.Context, key string) (interface{}, error) {
	start := time.Now()
	defer func() {
		r.metrics.LastLatency = time.Since(start)
		r.updateAvgLatency()
	}()

	fullKey := r.prefix + key

	r.metrics.Commands++

	result, err := r.client.Get(ctx, fullKey).Result()
	if err != nil {
		if err == redis.Nil {
			r.metrics.Misses++
			return nil, nil // 未找到，不返回错误
		}
		r.metrics.Errors++
		r.logger.Error("Redis GET error",
			zap.String("key", key),
			zap.Error(err),
		)
		return nil, fmt.Errorf("redis get error: %w", err)
	}

	r.metrics.Hits++

	// 尝试解析JSON
	var value interface{}
	if err := json.Unmarshal([]byte(result), &value); err != nil {
		// 如果不是JSON格式，直接返回字符串
		value = result
	}

	r.logger.Debug("Redis cache hit",
		zap.String("key", key),
		zap.Duration("latency", r.metrics.LastLatency),
	)

	return value, nil
}

// Set 设置缓存值
func (r *RedisL2Cache) Set(ctx context.Context, key string, value interface{}, ttl time.Duration) error {
	start := time.Now()
	defer func() {
		r.metrics.LastLatency = time.Since(start)
		r.updateAvgLatency()
	}()

	fullKey := r.prefix + key

	r.metrics.Commands++

	// 序列化值
	var data string
	switch v := value.(type) {
	case string:
		data = v
	case []byte:
		data = string(v)
	default:
		jsonBytes, err := json.Marshal(value)
		if err != nil {
			r.metrics.Errors++
			return fmt.Errorf("json marshal error: %w", err)
		}
		data = string(jsonBytes)
	}

	err := r.client.Set(ctx, fullKey, data, ttl).Err()
	if err != nil {
		r.metrics.Errors++
		r.logger.Error("Redis SET error",
			zap.String("key", key),
			zap.Duration("ttl", ttl),
			zap.Error(err),
		)
		return fmt.Errorf("redis set error: %w", err)
	}

	r.logger.Debug("Redis cache set",
		zap.String("key", key),
		zap.Duration("ttl", ttl),
		zap.Duration("latency", r.metrics.LastLatency),
	)

	return nil
}

// Delete 删除缓存
func (r *RedisL2Cache) Delete(ctx context.Context, key string) error {
	start := time.Now()
	defer func() {
		r.metrics.LastLatency = time.Since(start)
		r.updateAvgLatency()
	}()

	fullKey := r.prefix + key
	r.metrics.Commands++

	err := r.client.Del(ctx, fullKey).Err()
	if err != nil {
		r.metrics.Errors++
		r.logger.Error("Redis DELETE error",
			zap.String("key", key),
			zap.Error(err),
		)
		return fmt.Errorf("redis delete error: %w", err)
	}

	r.logger.Debug("Redis cache deleted",
		zap.String("key", key),
		zap.Duration("latency", r.metrics.LastLatency),
	)

	return nil
}

// Exists 检查键是否存在
func (r *RedisL2Cache) Exists(ctx context.Context, key string) (bool, error) {
	start := time.Now()
	defer func() {
		r.metrics.LastLatency = time.Since(start)
		r.updateAvgLatency()
	}()

	fullKey := r.prefix + key
	r.metrics.Commands++

	result, err := r.client.Exists(ctx, fullKey).Result()
	if err != nil {
		r.metrics.Errors++
		return false, fmt.Errorf("redis exists error: %w", err)
	}

	exists := result > 0

	r.logger.Debug("Redis exists check",
		zap.String("key", key),
		zap.Bool("exists", exists),
		zap.Duration("latency", r.metrics.LastLatency),
	)

	return exists, nil
}

// Clear 清空所有缓存
func (r *RedisL2Cache) Clear(ctx context.Context) error {
	start := time.Now()
	defer func() {
		r.metrics.LastLatency = time.Since(start)
		r.updateAvgLatency()
	}()

	r.metrics.Commands++

	// 使用SCAN命令获取所有匹配前缀的键
	iter := r.client.Scan(ctx, 0, r.prefix+"*", 0).Iterator()
	var keys []string

	for iter.Next(ctx) {
		keys = append(keys, iter.Val())
	}

	if err := iter.Err(); err != nil {
		r.metrics.Errors++
		return fmt.Errorf("redis scan error: %w", err)
	}

	if len(keys) > 0 {
		err := r.client.Del(ctx, keys...).Err()
		if err != nil {
			r.metrics.Errors++
			return fmt.Errorf("redis delete keys error: %w", err)
		}
	}

	r.logger.Info("Redis cache cleared",
		zap.Int("keys_deleted", len(keys)),
		zap.Duration("latency", r.metrics.LastLatency),
	)

	return nil
}

// Ping 测试连接
func (r *RedisL2Cache) Ping(ctx context.Context) error {
	start := time.Now()
	defer func() {
		r.metrics.LastLatency = time.Since(start)
		r.updateAvgLatency()
	}()

	r.metrics.Commands++

	err := r.client.Ping(ctx).Err()
	if err != nil {
		r.metrics.Errors++
		return fmt.Errorf("redis ping error: %w", err)
	}

	r.logger.Debug("Redis ping successful",
		zap.Duration("latency", r.metrics.LastLatency),
	)

	return nil
}

// GetMetrics 获取缓存指标
func (r *RedisL2Cache) GetMetrics() *L2Metrics {
	return r.metrics
}

// updateAvgLatency 更新平均延迟
func (r *RedisL2Cache) updateAvgLatency() {
	if r.metrics.Commands == 0 {
		return
	}

	// 简单的移动平均
	r.metrics.AvgLatency = time.Duration(
		(int64(r.metrics.AvgLatency) + int64(r.metrics.LastLatency)) / 2,
	)
}

// Close 关闭连接
func (r *RedisL2Cache) Close() error {
	if r.client != nil {
		err := r.client.Close()
		if err != nil {
			r.logger.Error("Failed to close Redis client", zap.Error(err))
			return err
		}
		r.logger.Info("Redis L2 cache closed")
	}
	return nil
}

// GetClient 获取原始Redis客户端（用于高级操作）
func (r *RedisL2Cache) GetClient() *redis.Client {
	return r.client
}
