package engine

import (
	"regexp"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestLRURegexCache(t *testing.T) {
	// 创建容量为3的LRU缓存
	cache := NewLRURegexCache(3)

	// 创建一些测试正则表达式
	re1, _ := regexp.Compile("test1")
	re2, _ := regexp.Compile("test2")
	re3, _ := regexp.Compile("test3")
	re4, _ := regexp.Compile("test4")

	// 测试Put和Get
	cache.Put("pattern1", re1)
	cache.Put("pattern2", re2)
	cache.Put("pattern3", re3)

	// 验证所有项都能获取到
	if re, exists := cache.Get("pattern1"); assert.True(t, exists) {
		assert.Equal(t, re1, re)
	}

	if re, exists := cache.Get("pattern2"); assert.True(t, exists) {
		assert.Equal(t, re2, re)
	}

	if re, exists := cache.Get("pattern3"); assert.True(t, exists) {
		assert.Equal(t, re3, re)
	}

	// 添加第四个项，应该触发LRU淘汰
	cache.Put("pattern4", re4)

	// pattern1应该是最久未使用的，应该被淘汰
	if _, exists := cache.Get("pattern1"); exists {
		t.Error("pattern1 should have been evicted")
	}

	// 其他项应该仍然存在
	if re, exists := cache.Get("pattern2"); assert.True(t, exists) {
		assert.Equal(t, re2, re)
	}

	if re, exists := cache.Get("pattern3"); assert.True(t, exists) {
		assert.Equal(t, re3, re)
	}

	if re, exists := cache.Get("pattern4"); assert.True(t, exists) {
		assert.Equal(t, re4, re)
	}

	// 访问pattern2，使其变为最近使用
	cache.Get("pattern2")

	// 再添加一个新项，pattern3应该被淘汰（因为pattern2刚被访问过）
	re5, _ := regexp.Compile("test5")
	cache.Put("pattern5", re5)

	if _, exists := cache.Get("pattern3"); exists {
		t.Error("pattern3 should have been evicted")
	}

	if re, exists := cache.Get("pattern2"); assert.True(t, exists) {
		assert.Equal(t, re2, re)
	}

	if re, exists := cache.Get("pattern4"); assert.True(t, exists) {
		assert.Equal(t, re4, re)
	}

	if re, exists := cache.Get("pattern5"); assert.True(t, exists) {
		assert.Equal(t, re5, re)
	}
}

func TestLRURegexCache_Size(t *testing.T) {
	cache := NewLRURegexCache(5)

	assert.Equal(t, 0, cache.Size())

	re1, _ := regexp.Compile("test1")
	cache.Put("pattern1", re1)
	assert.Equal(t, 1, cache.Size())

	re2, _ := regexp.Compile("test2")
	cache.Put("pattern2", re2)
	assert.Equal(t, 2, cache.Size())

	// 重复添加同一个pattern，大小应该不变
	cache.Put("pattern1", re1)
	assert.Equal(t, 2, cache.Size())

	// 添加新pattern
	re3, _ := regexp.Compile("test3")
	cache.Put("pattern3", re3)
	assert.Equal(t, 3, cache.Size())
}

func TestLRURegexCache_Clear(t *testing.T) {
	cache := NewLRURegexCache(5)

	re1, _ := regexp.Compile("test1")
	re2, _ := regexp.Compile("test2")

	cache.Put("pattern1", re1)
	cache.Put("pattern2", re2)

	assert.Equal(t, 2, cache.Size())

	cache.Clear()
	assert.Equal(t, 0, cache.Size())

	// 验证缓存已清空
	if _, exists := cache.Get("pattern1"); exists {
		t.Error("pattern1 should not exist after clear")
	}

	if _, exists := cache.Get("pattern2"); exists {
		t.Error("pattern2 should not exist after clear")
	}
}