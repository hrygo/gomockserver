package cache

import (
	"sort"
	"sync"
	"time"
)

// SimpleFrequencyTracker 简单的访问频率跟踪器
type SimpleFrequencyTracker struct {
	mu           sync.RWMutex
	accessData   map[string]*AccessData
	windowStart  time.Time
	windowLength time.Duration
}

// AccessData 访问数据
type AccessData struct {
	Key       string    `json:"key"`
	Count     int64     `json:"count"`
	LastSeen  time.Time `json:"last_seen"`
	Frequency float64   `json:"frequency"`
}

// NewSimpleFrequencyTracker 创建频率跟踪器
func NewSimpleFrequencyTracker(windowLength time.Duration) *SimpleFrequencyTracker {
	if windowLength <= 0 {
		windowLength = 1 * time.Hour
	}

	tracker := &SimpleFrequencyTracker{
		accessData:   make(map[string]*AccessData),
		windowStart:  time.Now(),
		windowLength: windowLength,
	}

	// 启动定期清理
	go tracker.startWindowRotation()

	return tracker
}

// RecordAccess 记录访问
func (t *SimpleFrequencyTracker) RecordAccess(key string) {
	t.mu.Lock()
	defer t.mu.Unlock()

	// 检查是否需要重置窗口
	if time.Since(t.windowStart) > t.windowLength {
		t.resetWindow()
	}

	data, exists := t.accessData[key]
	if !exists {
		data = &AccessData{
			Key:       key,
			Count:     0,
			LastSeen:  time.Now(),
			Frequency: 0.0,
		}
		t.accessData[key] = data
	}

	data.Count++
	data.LastSeen = time.Now()
}

// GetFrequency 获取访问频率
func (t *SimpleFrequencyTracker) GetFrequency(key string) float64 {
	t.mu.RLock()
	defer t.mu.RUnlock()

	data, exists := t.accessData[key]
	if !exists {
		return 0.0
	}

	// 计算频率：访问次数 / 窗口时长（小时）
	windowHours := t.windowLength.Hours()
	if windowHours > 0 {
		data.Frequency = float64(data.Count) / windowHours
	}

	return data.Frequency
}

// CleanupExpiredWindow 清理过期窗口
func (t *SimpleFrequencyTracker) CleanupExpiredWindow() {
	t.mu.Lock()
	defer t.mu.Unlock()

	if time.Since(t.windowStart) > t.windowLength {
		t.resetWindow()
	}
}

// GetTopKeys 获取热点键
func (t *SimpleFrequencyTracker) GetTopKeys(limit int) []string {
	t.mu.RLock()
	defer t.mu.RUnlock()

	if limit <= 0 {
		limit = 10
	}

	// 收集所有访问数据
	var dataList []*AccessData
	for _, data := range t.accessData {
		// 更新频率
		windowHours := t.windowLength.Hours()
		if windowHours > 0 {
			data.Frequency = float64(data.Count) / windowHours
		}
		dataList = append(dataList, data)
	}

	// 按频率排序
	sort.Slice(dataList, func(i, j int) bool {
		return dataList[i].Frequency > dataList[j].Frequency
	})

	// 返回前N个键
	if len(dataList) < limit {
		limit = len(dataList)
	}

	result := make([]string, 0, limit)
	for i := 0; i < limit; i++ {
		result = append(result, dataList[i].Key)
	}

	return result
}

// resetWindow 重置统计窗口
func (t *SimpleFrequencyTracker) resetWindow() {
	// 清理过期的访问数据
	now := time.Now()
	for key, data := range t.accessData {
		// 如果超过2个窗口长度没有访问，删除数据
		if now.Sub(data.LastSeen) > 2*t.windowLength {
			delete(t.accessData, key)
		}
	}

	// 重置窗口开始时间
	t.windowStart = now
}

// startWindowRotation 启动窗口轮转
func (t *SimpleFrequencyTracker) startWindowRotation() {
	ticker := time.NewTicker(t.windowLength / 4) // 每1/4窗口时间检查一次
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			t.CleanupExpiredWindow()
		}
	}
}

// GetAccessData 获取访问数据
func (t *SimpleFrequencyTracker) GetAccessData() map[string]*AccessData {
	t.mu.RLock()
	defer t.mu.RUnlock()

	result := make(map[string]*AccessData)
	for key, data := range t.accessData {
		// 复制数据
		dataCopy := *data
		result[key] = &dataCopy
	}

	return result
}

// GetStats 获取统计信息
func (t *SimpleFrequencyTracker) GetStats() map[string]interface{} {
	t.mu.RLock()
	defer t.mu.RUnlock()

	var totalAccesses int64
	var activeKeys int

	windowRemaining := t.windowLength - time.Since(t.windowStart)

	for _, data := range t.accessData {
		totalAccesses += data.Count
		if time.Since(data.LastSeen) < t.windowLength {
			activeKeys++
		}
	}

	return map[string]interface{}{
		"total_accesses":     totalAccesses,
		"active_keys":        activeKeys,
		"total_keys":         len(t.accessData),
		"window_start":       t.windowStart,
		"window_length":      t.windowLength,
		"window_remaining":   windowRemaining,
		"avg_access_per_key": float64(totalAccesses) / float64(len(t.accessData)),
	}
}
