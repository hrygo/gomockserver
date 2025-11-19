package cache

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

// TestSimpleFrequencyTracker_BasicOperations 测试频率跟踪器基本操作
func TestSimpleFrequencyTracker_BasicOperations(t *testing.T) {
	// 使用短窗口以便快速测试
	tracker := NewSimpleFrequencyTracker(100 * time.Millisecond)

	// 测试记录访问
	tracker.RecordAccess("key1")
	tracker.RecordAccess("key1")
	tracker.RecordAccess("key2")

	// 获取频率
	freq1 := tracker.GetFrequency("key1")
	freq2 := tracker.GetFrequency("key2")
	freq3 := tracker.GetFrequency("key3") // 不存在的键

	assert.Greater(t, freq1, freq2)
	assert.Greater(t, freq2, freq3)

	// 获取热门键
	topKeys := tracker.GetTopKeys(2)
	assert.Len(t, topKeys, 2)
	assert.Equal(t, "key1", topKeys[0])
	assert.Equal(t, "key2", topKeys[1])

	// 清理过期窗口
	tracker.CleanupExpiredWindow()

	// 等待窗口重置
	time.Sleep(150 * time.Millisecond)

	// 再次测试访问
	tracker.RecordAccess("key4")
	freq4 := tracker.GetFrequency("key4")
	assert.Greater(t, freq4, float64(0))
}

// TestSimpleFrequencyTracker_EdgeCases 测试边界情况
func TestSimpleFrequencyTracker_EdgeCases(t *testing.T) {
	t.Run("Zero window length", func(t *testing.T) {
		tracker := NewSimpleFrequencyTracker(0)
		tracker.RecordAccess("key1")
		freq := tracker.GetFrequency("key1")
		assert.GreaterOrEqual(t, freq, float64(0))
	})

	t.Run("Negative window length", func(t *testing.T) {
		tracker := NewSimpleFrequencyTracker(-1 * time.Second)
		tracker.RecordAccess("key1")
		freq := tracker.GetFrequency("key1")
		assert.GreaterOrEqual(t, freq, float64(0))
	})

	t.Run("Empty tracker", func(t *testing.T) {
		tracker := NewSimpleFrequencyTracker(1 * time.Hour)
		topKeys := tracker.GetTopKeys(5)
		assert.Empty(t, topKeys)

		freq := tracker.GetFrequency("non_existent")
		assert.Equal(t, float64(0), freq)
	})

	t.Run("Large limit", func(t *testing.T) {
		tracker := NewSimpleFrequencyTracker(1 * time.Hour)
		tracker.RecordAccess("key1")
		tracker.RecordAccess("key2")

		topKeys := tracker.GetTopKeys(100)
		assert.Len(t, topKeys, 2)
	})
}

// TestAccessData_Creation 测试访问数据创建
func TestAccessData_Creation(t *testing.T) {
	data := &AccessData{
		Key:       "test_key",
		Count:     10,
		LastSeen:  time.Now(),
		Frequency: 5.5,
	}

	assert.Equal(t, "test_key", data.Key)
	assert.Equal(t, int64(10), data.Count)
	assert.Equal(t, 5.5, data.Frequency)
}

// TestSimpleFrequencyTracker_AdvancedMethods 测试高级方法
func TestSimpleFrequencyTracker_AdvancedMethods(t *testing.T) {
	tracker := NewSimpleFrequencyTracker(1 * time.Hour)

	// 记录一些访问
	tracker.RecordAccess("key1")
	tracker.RecordAccess("key1")
	tracker.RecordAccess("key2")

	// 测试GetAccessData
	allAccessData := tracker.GetAccessData()
	assert.NotNil(t, allAccessData)
	assert.Contains(t, allAccessData, "key1")
	assert.Contains(t, allAccessData, "key2")

	// 检查具体数据
	key1Data := allAccessData["key1"]
	assert.NotNil(t, key1Data)
	assert.Equal(t, "key1", key1Data.Key)
	assert.Equal(t, int64(2), key1Data.Count)

	// 测试GetStats
	stats := tracker.GetStats()
	assert.NotNil(t, stats)
	// Stats结构的具体字段取决于实现，这里只测试不为nil
}