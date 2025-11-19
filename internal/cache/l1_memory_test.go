package cache

import (
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"go.uber.org/zap/zaptest"
)

func TestMemoryL1Cache_BasicOperations(t *testing.T) {
	logger := zaptest.NewLogger(t)
	cache := NewMemoryL1Cache(100, 50, 1*time.Minute, logger)

	// 测试设置和获取
	t.Run("Set and Get", func(t *testing.T) {
		key := "test_key"
		value := "test_value"

		err := cache.Set(key, value, 10*time.Minute)
		assert.NoError(t, err)

		entry, found := cache.Get(key)
		assert.True(t, found)
		assert.NotNil(t, entry)
		assert.Equal(t, value, entry.Value)
		assert.Equal(t, key, entry.Key)
	})

	// 测试获取不存在的键
	t.Run("Get non-existent key", func(t *testing.T) {
		_, found := cache.Get("non_existent_key")
		assert.False(t, found)
	})
}

func TestMemoryL1Cache_Delete(t *testing.T) {
	logger := zaptest.NewLogger(t)
	cache := NewMemoryL1Cache(100, 50, 1*time.Minute, logger)

	key := "test_key"
	value := "test_value"

	// 设置值
	err := cache.Set(key, value, 10*time.Minute)
	assert.NoError(t, err)

	// 确认值存在
	entry, found := cache.Get(key)
	assert.True(t, found)
	assert.Equal(t, value, entry.Value)

	// 删除值
	err = cache.Delete(key)
	assert.NoError(t, err)

	// 确认值已被删除
	_, found = cache.Get(key)
	assert.False(t, found)
}

func TestMemoryL1Cache_Clear(t *testing.T) {
	logger := zaptest.NewLogger(t)
	cache := NewMemoryL1Cache(100, 50, 1*time.Minute, logger)

	// 添加一些值
	for i := 0; i < 10; i++ {
		key := "key_" + string(rune('0'+i))
		value := "value_" + string(rune('0'+i))
		err := cache.Set(key, value, 10*time.Minute)
		assert.NoError(t, err)
	}

	// 确认值存在
	stats := cache.Stats()
	assert.Equal(t, int64(10), stats.Entries)

	// 清空缓存
	err := cache.Clear()
	assert.NoError(t, err)

	// 确认缓存已空
	stats = cache.Stats()
	assert.Equal(t, int64(0), stats.Entries)
}

func TestMemoryL1Cache_Cleanup(t *testing.T) {
	logger := zaptest.NewLogger(t)
	cache := NewMemoryL1Cache(100, 50, 50*time.Millisecond, logger)

	// 添加一些值，其中一些会很快过期
	for i := 0; i < 5; i++ {
		key := "short_" + string(rune('0'+i))
		value := "value_" + string(rune('0'+i))
		err := cache.Set(key, value, 20*time.Millisecond) // 短TTL
		assert.NoError(t, err)
	}

	for i := 0; i < 5; i++ {
		key := "long_" + string(rune('0'+i))
		value := "value_" + string(rune('0'+i))
		err := cache.Set(key, value, 1*time.Hour) // 长TTL
		assert.NoError(t, err)
	}

	// 等待短TTL的值过期
	time.Sleep(100 * time.Millisecond)

	// 执行清理
	err := cache.Cleanup()
	assert.NoError(t, err)

	// 短TTL的值应该被清理
	for i := 0; i < 5; i++ {
		key := "short_" + string(rune('0'+i))
		_, found := cache.Get(key)
		assert.False(t, found)
	}

	// 长TTL的值应该还存在
	for i := 0; i < 5; i++ {
		key := "long_" + string(rune('0'+i))
		_, found := cache.Get(key)
		assert.True(t, found)
	}
}

func TestMemoryL1Cache_GetStats(t *testing.T) {
	logger := zaptest.NewLogger(t)
	cache := NewMemoryL1Cache(100, 50, 1*time.Minute, logger)

	// 初始统计
	stats := cache.Stats()
	assert.Equal(t, int64(0), stats.Entries)
	assert.Equal(t, int64(0), stats.Hits)
	assert.Equal(t, int64(0), stats.Misses)

	// 添加一些值
	for i := 0; i < 3; i++ {
		key := "key_" + string(rune('0'+i))
		value := "value_" + string(rune('0'+i))
		err := cache.Set(key, value, 10*time.Minute)
		assert.NoError(t, err)
	}

	stats = cache.Stats()
	assert.Equal(t, int64(3), stats.Entries)

	// 命中测试
	_, found := cache.Get("key_0")
	assert.True(t, found)

	stats = cache.Stats()
	assert.Equal(t, int64(1), stats.Hits)

	// 未命中测试
	_, found = cache.Get("non_existent")
	assert.False(t, found)

	stats = cache.Stats()
	assert.Equal(t, int64(1), stats.Misses)
}

func TestMemoryL1Cache_CapacityLimit(t *testing.T) {
	logger := zaptest.NewLogger(t)
	capacity := 3
	cache := NewMemoryL1Cache(capacity, 50, 1*time.Minute, logger)

	// 填满缓存
	for i := 0; i < capacity; i++ {
		key := "key_" + string(rune('A'+i))
		value := "value_" + string(rune('A'+i))
		err := cache.Set(key, value, 10*time.Minute)
		assert.NoError(t, err)
	}

	// 所有值都应该存在
	for i := 0; i < capacity; i++ {
		key := "key_" + string(rune('A'+i))
		_, found := cache.Get(key)
		assert.True(t, found)
	}

	// 添加新值应该淘汰最旧的值
	newKey := "key_new"
	newValue := "value_new"
	err := cache.Set(newKey, newValue, 10*time.Minute)
	assert.NoError(t, err)

	// 第一个值应该被淘汰
	firstKey := "key_A"
	_, found := cache.Get(firstKey)
	assert.False(t, found)

	// 新值应该存在
	_, found = cache.Get(newKey)
	assert.True(t, found)
}

// 基准测试
func BenchmarkMemoryL1Cache_Set(b *testing.B) {
	logger := zaptest.NewLogger(b)
	cache := NewMemoryL1Cache(10000, 1000, 1*time.Minute, logger)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		key := "benchmark_key_" + string(rune(i%1000))
		value := "benchmark_value_" + string(rune(i%1000))
		cache.Set(key, value, 10*time.Minute)
	}
}

func TestMemoryL1Cache_GetTopKeys(t *testing.T) {
	logger := zaptest.NewLogger(t)
	cache := NewMemoryL1Cache(100, 50, 1*time.Minute, logger)

	// 添加不同访问次数的键
	keys := []string{"key_1", "key_2", "key_3", "key_4", "key_5"}
	for i, key := range keys {
		value := "value_" + key
		err := cache.Set(key, value, 10*time.Minute)
		assert.NoError(t, err)

		// 模拟不同的访问次数
		for j := 0; j <= i; j++ {
			cache.Get(key)
		}
	}

	// 测试获取前3个热点键
	topKeys := cache.GetTopKeys(3)
	assert.Len(t, topKeys, 3)
	// key_5应该排在最前面（访问次数最多）
	assert.Equal(t, "key_5", topKeys[0])
	assert.Equal(t, "key_4", topKeys[1])
	assert.Equal(t, "key_3", topKeys[2])

	// 测试limit大于总键数
	allKeys := cache.GetTopKeys(10)
	assert.Len(t, allKeys, 5)

	// 测试limit为0
	zeroKeys := cache.GetTopKeys(0)
	assert.Len(t, zeroKeys, 5) // 应该返回所有键

	// 测试空缓存
	emptyCache := NewMemoryL1Cache(100, 50, 1*time.Minute, logger)
	emptyKeys := emptyCache.GetTopKeys(5)
	assert.Empty(t, emptyKeys)
}

func TestMemoryL1Cache_GetMemoryUsage(t *testing.T) {
	logger := zaptest.NewLogger(t)
	cache := NewMemoryL1Cache(100, 50, 1*time.Minute, logger)

	// 初始内存使用情况
	usage := cache.GetMemoryUsage()
	assert.Equal(t, int64(0), usage["used_bytes"])
	assert.Equal(t, int64(50*1024*1024), usage["capacity_bytes"]) // 50MB
	assert.Equal(t, int64(0), usage["entries"])
	assert.Contains(t, usage, "usage_percent")

	// 添加一些数据
	for i := 0; i < 5; i++ {
		key := "key_" + string(rune('0'+i))
		value := "value_" + string(rune('0'+i))
		err := cache.Set(key, value, 10*time.Minute)
		assert.NoError(t, err)
	}

	// 检查内存使用情况
	usage = cache.GetMemoryUsage()
	assert.Greater(t, usage["used_bytes"], int64(0))
	assert.Equal(t, int64(5), usage["entries"])
	assert.Greater(t, usage["usage_percent"], float64(0))
}

func TestMemoryL1Cache_Stop(t *testing.T) {
	logger := zaptest.NewLogger(t)
	cache := NewMemoryL1Cache(100, 50, 1*time.Minute, logger)

	// 添加一些数据
	for i := 0; i < 5; i++ {
		key := "key_" + string(rune('0'+i))
		value := "value_" + string(rune('0'+i))
		err := cache.Set(key, value, 10*time.Minute)
		assert.NoError(t, err)
	}

	// 确认数据存在
	stats := cache.Stats()
	assert.Equal(t, int64(5), stats.Entries)

	// 停止缓存
	cache.Stop()

	// 确认缓存已清空
	stats = cache.Stats()
	assert.Equal(t, int64(0), stats.Entries)
}

func TestMemoryL1Cache_EdgeCases(t *testing.T) {
	logger := zaptest.NewLogger(t)

	t.Run("Empty cache operations", func(t *testing.T) {
		cache := NewMemoryL1Cache(10, 10, 1*time.Minute, logger)

		// 对空缓存进行操作
		_, found := cache.Get("non_existent")
		assert.False(t, found)

		err := cache.Delete("non_existent")
		assert.NoError(t, err) // 删除不存在的键不应该报错

		stats := cache.Stats()
		assert.Equal(t, int64(0), stats.Entries)
		assert.Equal(t, int64(0), stats.Hits)
		assert.Equal(t, int64(1), stats.Misses) // Get操作会增加miss计数
	})

	t.Run("Nil values", func(t *testing.T) {
		cache := NewMemoryL1Cache(10, 10, 1*time.Minute, logger)

		// 测试设置nil值（应该被拒绝）
		err := cache.Set("nil_key", nil, 1*time.Minute)
		// 注意：这里取决于实现是否允许nil值
		// 如果允许，后续测试需要相应调整
		if err == nil {
			// 如果允许nil值，检查是否能正确获取
			entry, found := cache.Get("nil_key")
			assert.True(t, found)
			assert.Nil(t, entry.Value)
		}
	})

	t.Run("Very large values", func(t *testing.T) {
		smallCache := NewMemoryL1Cache(2, 1, 1*time.Minute, logger) // 1MB容量

		// 创建一个大于缓存容量的值
		largeValue := make([]byte, 2*1024*1024) // 2MB
		err := smallCache.Set("large_key", largeValue, 1*time.Minute)
		// 根据实现可能成功或失败
		if err == nil {
			// 如果成功，检查是否能正确处理
			entry, found := smallCache.Get("large_key")
			assert.True(t, found)
			assert.NotNil(t, entry.Value)
		}
	})
}

func TestMemoryL1Cache_EstimateSize(t *testing.T) {
	logger := zaptest.NewLogger(t)
	cache := NewMemoryL1Cache(100, 50, 1*time.Minute, logger)

	// 测试不同类型值的大小估算
	testCases := []struct {
		name  string
		value interface{}
	}{
		{"string", "hello world"},
		{"int", 42},
		{"float", 3.14},
		{"bool", true},
		{"slice", []int{1, 2, 3, 4, 5}},
		{"map", map[string]interface{}{"key": "value"}},
		{"struct", struct {
			Name string
			Age  int
		}{"John", 30}},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			err := cache.Set("test_key", tc.value, 1*time.Minute)
			assert.NoError(t, err)

			// 验证值被正确存储
			entry, found := cache.Get("test_key")
			assert.True(t, found)
			assert.Equal(t, tc.value, entry.Value)

			// 清理以便下次测试
			cache.Delete("test_key")
		})
	}
}

func TestMemoryL1Cache_ConcurrentAccess(t *testing.T) {
	logger := zaptest.NewLogger(t)
	cache := NewMemoryL1Cache(1000, 50, 1*time.Minute, logger)

	// 并发写入
	done := make(chan bool, 10)
	for i := 0; i < 10; i++ {
		go func(id int) {
			for j := 0; j < 100; j++ {
				key := "key_" + string(rune('0'+id)) + "_" + string(rune('0'+j))
				value := "value_" + string(rune('0'+id)) + "_" + string(rune('0'+j))
				err := cache.Set(key, value, 10*time.Minute)
				assert.NoError(t, err)
			}
			done <- true
		}(i)
	}

	// 等待所有写入完成
	for i := 0; i < 10; i++ {
		<-done
	}

	// 并发读取
	for i := 0; i < 10; i++ {
		go func(id int) {
			for j := 0; j < 100; j++ {
				key := "key_" + string(rune('0'+id)) + "_" + string(rune('0'+j))
				expectedValue := "value_" + string(rune('0'+id)) + "_" + string(rune('0'+j))
				entry, found := cache.Get(key)
				assert.True(t, found)
				assert.Equal(t, expectedValue, entry.Value)
			}
			done <- true
		}(i)
	}

	// 等待所有读取完成
	for i := 0; i < 10; i++ {
		<-done
	}

	// 验证最终统计
	stats := cache.Stats()
	assert.Equal(t, int64(1000), stats.Entries)
	assert.Equal(t, int64(1000), stats.Hits)
}

func TestMemoryL1Cache_ConcurrentMixedOperations(t *testing.T) {
	logger := zaptest.NewLogger(t)
	cache := NewMemoryL1Cache(500, 50, 1*time.Minute, logger)

	done := make(chan bool, 20)

	// 并发写入
	for i := 0; i < 5; i++ {
		go func(id int) {
			for j := 0; j < 50; j++ {
				key := fmt.Sprintf("write_%d_%d", id, j)
				value := fmt.Sprintf("value_%d_%d", id, j)
				cache.Set(key, value, 5*time.Minute)
			}
			done <- true
		}(i)
	}

	// 并发读取
	for i := 0; i < 5; i++ {
		go func(id int) {
			for j := 0; j < 50; j++ {
				key := fmt.Sprintf("read_%d_%d", id, j)
				cache.Get(key) // 这些键可能不存在，测试miss情况
			}
			done <- true
		}(i)
	}

	// 并发删除
	for i := 0; i < 5; i++ {
		go func(id int) {
			for j := 0; j < 20; j++ {
				key := fmt.Sprintf("write_%d_%d", id, j)
				cache.Delete(key)
			}
			done <- true
		}(i)
	}

	// 并发获取统计信息
	for i := 0; i < 5; i++ {
		go func() {
			for j := 0; j < 30; j++ {
				cache.Stats()
				cache.GetMemoryUsage()
			}
			done <- true
		}()
	}

	// 等待所有操作完成
	for i := 0; i < 20; i++ {
		<-done
	}

	// 验证缓存仍然正常工作
	finalStats := cache.Stats()
	assert.GreaterOrEqual(t, finalStats.Entries, int64(0))
	assert.GreaterOrEqual(t, finalStats.Hits, int64(0))
	assert.GreaterOrEqual(t, finalStats.Misses, int64(0))
}

func BenchmarkMemoryL1Cache_Get(b *testing.B) {
	logger := zaptest.NewLogger(b)
	cache := NewMemoryL1Cache(10000, 1000, 1*time.Minute, logger)

	// 预填充数据
	for i := 0; i < 1000; i++ {
		key := "benchmark_key_" + string(rune(i))
		value := "benchmark_value_" + string(rune(i))
		cache.Set(key, value, 10*time.Minute)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		key := "benchmark_key_" + string(rune(i%1000))
		cache.Get(key)
	}
}
