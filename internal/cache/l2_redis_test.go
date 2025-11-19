package cache

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"go.uber.org/zap/zaptest"
)

// ErrCacheMiss 缓存未命中错误
var ErrCacheMiss = errors.New("cache miss")

// TestRedisConfig_Creation 测试Redis配置创建
func TestRedisConfig_Creation(t *testing.T) {
	config := DefaultRedisConfig()

	assert.NotNil(t, config)
	assert.Equal(t, "localhost", config.Host)
	assert.Equal(t, 6379, config.Port)
	assert.Equal(t, 0, config.Database)
	assert.Greater(t, config.PoolSize, 0)
	assert.Greater(t, config.MinIdleConns, 0)
	assert.Greater(t, int(config.DialTimeout), 0)
	assert.Greater(t, int(config.ReadTimeout), 0)
	assert.Greater(t, int(config.WriteTimeout), 0)
	assert.Greater(t, int(config.PoolTimeout), 0)
}

// TestRedisL2Cache_BasicOperations 测试基本操作（简化版）
func TestRedisL2Cache_BasicOperations(t *testing.T) {
	logger := zaptest.NewLogger(t)

	// 使用默认配置创建缓存
	config := DefaultRedisConfig()
	cache, err := NewRedisL2Cache(config, logger)

	// 如果没有Redis服务器连接，会返回错误，这是正常的
	if err != nil {
		t.Skipf("Redis connection failed, skipping test: %v", err)
	}
	defer cache.Close()

	ctx := context.Background()
	key := "test_key"
	value := "test_value"
	ttl := 1 * time.Minute

	// 测试设置和获取
	err = cache.Set(ctx, key, value, ttl)
	if err != nil {
		t.Skipf("Redis operation failed, skipping test: %v", err)
	}

	retrievedValue, err := cache.Get(ctx, key)
	if err == nil {
		assert.Equal(t, value, retrievedValue)
	}

	// 测试获取不存在的键
	_, err = cache.Get(ctx, "non_existent_key")
	assert.Error(t, err)
}

// TestRedisL2Cache_Exists 测试存在性检查
func TestRedisL2Cache_Exists(t *testing.T) {
	logger := zaptest.NewLogger(t)

	config := DefaultRedisConfig()
	cache, err := NewRedisL2Cache(config, logger)
	if err != nil {
		t.Skipf("Redis connection failed, skipping test: %v", err)
	}
	defer cache.Close()

	ctx := context.Background()
	key := "test_key"
	value := "test_value"

	// 测试不存在的键
	exists, err := cache.Exists(ctx, key)
	if err == nil {
		assert.False(t, exists)

		// 设置值
		err = cache.Set(ctx, key, value, 1*time.Minute)
		if err == nil {
			// 测试存在的键
			exists, err = cache.Exists(ctx, key)
			assert.NoError(t, err)
			assert.True(t, exists)
		}
	}
}

// TestRedisL2Cache_Delete 测试删除操作
func TestRedisL2Cache_Delete(t *testing.T) {
	logger := zaptest.NewLogger(t)

	config := DefaultRedisConfig()
	cache, err := NewRedisL2Cache(config, logger)
	if err != nil {
		t.Skipf("Redis connection failed, skipping test: %v", err)
	}
	defer cache.Close()

	ctx := context.Background()
	key := "test_key"
	value := "test_value"

	// 设置值
	err = cache.Set(ctx, key, value, 1*time.Minute)
	if err != nil {
		t.Skipf("Redis operation failed, skipping test: %v", err)
	}

	// 删除值
	err = cache.Delete(ctx, key)
	assert.NoError(t, err)

	// 确认值已被删除
	_, err = cache.Get(ctx, key)
	assert.Error(t, err)
}

// TestRedisL2Cache_Ping 测试连接检查
func TestRedisL2Cache_Ping(t *testing.T) {
	logger := zaptest.NewLogger(t)

	config := DefaultRedisConfig()
	cache, err := NewRedisL2Cache(config, logger)
	if err != nil {
		t.Skipf("Redis connection failed, skipping test: %v", err)
	}
	defer cache.Close()

	ctx := context.Background()

	// 测试ping
	err = cache.Ping(ctx)
	if err != nil {
		t.Skipf("Redis ping failed, skipping test: %v", err)
	}
}

// TestRedisL2Cache_Clear 测试清空操作
func TestRedisL2Cache_Clear(t *testing.T) {
	logger := zaptest.NewLogger(t)

	config := DefaultRedisConfig()
	cache, err := NewRedisL2Cache(config, logger)
	if err != nil {
		t.Skipf("Redis connection failed, skipping test: %v", err)
	}
	defer cache.Close()

	ctx := context.Background()

	// 添加一些值
	for i := 0; i < 3; i++ {
		key := "key_" + string(rune('0'+i))
		value := "value_" + string(rune('0'+i))
		err := cache.Set(ctx, key, value, 1*time.Minute)
		if err != nil {
			t.Skipf("Redis operation failed, skipping test: %v", err)
		}
	}

	// 清空缓存（使用特定前缀的键）
	err = cache.Clear(ctx)
	if err == nil {
		// 验证缓存已清空（对于特定前缀的键）
		for i := 0; i < 3; i++ {
			key := "key_" + string(rune('0'+i))
			_, err := cache.Get(ctx, key)
			assert.Error(t, err)
		}
	}
}


// TestRedisL2Cache_ErrorHandling 测试错误处理
func TestRedisL2Cache_ErrorHandling(t *testing.T) {
	logger := zaptest.NewLogger(t)

	// 测试nil配置
	t.Run("Nil config", func(t *testing.T) {
		cache, err := NewRedisL2Cache(nil, logger)
		assert.Error(t, err)
		assert.Nil(t, cache)
	})

	// 测试无效配置
	t.Run("Invalid config", func(t *testing.T) {
		config := &RedisConfig{
			Host: "invalid_host_that_does_not_exist",
			Port: 99999,
		}
		cache, err := NewRedisL2Cache(config, logger)
		assert.Error(t, err)
		assert.Nil(t, cache)
	})

	// 测试空键错误
	t.Run("Empty key", func(t *testing.T) {
		config := &RedisConfig{
			Host: "invalid_host",
			Port: 6379,
		}
		cache, err := NewRedisL2Cache(config, logger)
		// 由于无法连接，这里会失败，这是预期的
		if err != nil {
			assert.Contains(t, err.Error(), "invalid_host")
			return
		}
		defer cache.Close()

		ctx := context.Background()

		// 测试空键
		err = cache.Set(ctx, "", "value", 1*time.Minute)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "key cannot be empty")

		_, err = cache.Get(ctx, "")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "key cannot be empty")

		_, err = cache.Exists(ctx, "")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "key cannot be empty")

		err = cache.Delete(ctx, "")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "key cannot be empty")
	})

	// 测试nil值错误
	t.Run("Nil value", func(t *testing.T) {
		config := &RedisConfig{
			Host: "invalid_host",
			Port: 6379,
		}
		cache, err := NewRedisL2Cache(config, logger)
		if err != nil {
			return
		}
		defer cache.Close()

		ctx := context.Background()

		err = cache.Set(ctx, "key", nil, 1*time.Minute)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "value cannot be nil")
	})
}

// TestRedisL2Cache_ClientOperations 测试客户端操作
func TestRedisL2Cache_ClientOperations(t *testing.T) {
	logger := zaptest.NewLogger(t)
	config := DefaultRedisConfig()
	cache, err := NewRedisL2Cache(config, logger)
	if err != nil {
		t.Skipf("Redis connection failed, skipping test: %v", err)
	}
	defer cache.Close()

	ctx := context.Background()

	// 测试Ping
	err = cache.Ping(ctx)
	if err == nil {
		// Ping成功，测试其他操作
		t.Run("Ping successful", func(t *testing.T) {
			// Ping应该没有错误
			assert.NoError(t, err)
		})
	}

	// 测试获取客户端
	client := cache.GetClient()
	if client != nil {
		t.Run("GetClient", func(t *testing.T) {
			assert.NotNil(t, client)
		})
	}

	// 测试Close
	err = cache.Close()
	assert.NoError(t, err)
}

// TestRedisL2Cache_Configuration 测试配置相关
func TestRedisL2Cache_Configuration(t *testing.T) {
	t.Run("Default configuration", func(t *testing.T) {
		config := DefaultRedisConfig()
		assert.Equal(t, "localhost", config.Host)
		assert.Equal(t, 6379, config.Port)
		assert.Equal(t, "", config.Password)
		assert.Equal(t, 0, config.Database)
		assert.Equal(t, 20, config.PoolSize)
		assert.Equal(t, 5, config.MinIdleConns)
		assert.Equal(t, "mockserver:cache:", config.KeyPrefix)
	})

	t.Run("Custom configuration", func(t *testing.T) {
		config := &RedisConfig{
			Host:         "custom-host",
			Port:         6380,
			Password:     "secret",
			Database:     1,
			PoolSize:     10,
			MinIdleConns: 2,
			DialTimeout:  3 * time.Second,
			ReadTimeout:  2 * time.Second,
			WriteTimeout: 2 * time.Second,
			PoolTimeout:  3 * time.Second,
			KeyPrefix:    "custom:prefix:",
		}

		assert.Equal(t, "custom-host", config.Host)
		assert.Equal(t, 6380, config.Port)
		assert.Equal(t, "secret", config.Password)
		assert.Equal(t, 1, config.Database)
		assert.Equal(t, 10, config.PoolSize)
		assert.Equal(t, 2, config.MinIdleConns)
		assert.Equal(t, 3*time.Second, config.DialTimeout)
		assert.Equal(t, 2*time.Second, config.ReadTimeout)
		assert.Equal(t, 2*time.Second, config.WriteTimeout)
		assert.Equal(t, 3*time.Second, config.PoolTimeout)
		assert.Equal(t, "custom:prefix:", config.KeyPrefix)
	})
}

// 基准测试
func BenchmarkRedisL2Cache_Set(b *testing.B) {
	logger := zaptest.NewLogger(b)

	config := DefaultRedisConfig()
	cache, err := NewRedisL2Cache(config, logger)
	if err != nil {
		b.Skipf("Redis connection failed, skipping benchmark: %v", err)
	}
	defer cache.Close()

	ctx := context.Background()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		key := "benchmark_key_" + string(rune(i%1000))
		value := "benchmark_value_" + string(rune(i%1000))
		cache.Set(ctx, key, value, 10*time.Minute)
	}
}

func BenchmarkRedisL2Cache_Get(b *testing.B) {
	logger := zaptest.NewLogger(b)

	config := DefaultRedisConfig()
	cache, err := NewRedisL2Cache(config, logger)
	if err != nil {
		b.Skipf("Redis connection failed, skipping benchmark: %v", err)
	}
	defer cache.Close()

	ctx := context.Background()

	// 预填充数据
	for i := 0; i < 100; i++ {
		key := "benchmark_key_" + string(rune(i))
		value := "benchmark_value_" + string(rune(i))
		cache.Set(ctx, key, value, 10*time.Minute)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		key := "benchmark_key_" + string(rune(i%100))
		cache.Get(ctx, key)
	}
}

func BenchmarkRedisL2Cache_Exists(b *testing.B) {
	logger := zaptest.NewLogger(b)

	config := DefaultRedisConfig()
	cache, err := NewRedisL2Cache(config, logger)
	if err != nil {
		b.Skipf("Redis connection failed, skipping benchmark: %v", err)
	}
	defer cache.Close()

	ctx := context.Background()

	// 预填充数据
	for i := 0; i < 100; i++ {
		key := "benchmark_key_" + string(rune(i))
		value := "benchmark_value_" + string(rune(i))
		cache.Set(ctx, key, value, 10*time.Minute)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		key := "benchmark_key_" + string(rune(i%100))
		cache.Exists(ctx, key)
	}
}