// Package engine provides the rule matching engine for the mock server.
package engine

import (
	"container/list"
	"regexp"
	"sync"
)

// LRURegexCache LRU正则表达式缓存
type LRURegexCache struct {
	capacity int
	cache    map[string]*list.Element
	list     *list.List
	mu       sync.RWMutex
}

// regexCacheItem 缓存项
type regexCacheItem struct {
	pattern string
	regex   *regexp.Regexp
}

// NewLRURegexCache 创建LRU正则表达式缓存
func NewLRURegexCache(capacity int) *LRURegexCache {
	return &LRURegexCache{
		capacity: capacity,
		cache:    make(map[string]*list.Element),
		list:     list.New(),
	}
}

// Get 获取缓存的正则表达式
func (c *LRURegexCache) Get(pattern string) (*regexp.Regexp, bool) {
	c.mu.Lock()
	defer c.mu.Unlock()
	
	if element, exists := c.cache[pattern]; exists {
		// 移动到列表头部（最近使用）
		c.list.MoveToFront(element)
		return element.Value.(*regexCacheItem).regex, true
	}
	
	return nil, false
}

// Put 存储正则表达式到缓存
func (c *LRURegexCache) Put(pattern string, regex *regexp.Regexp) {
	c.mu.Lock()
	defer c.mu.Unlock()
	
	if element, exists := c.cache[pattern]; exists {
		// 更新现有项
		c.list.MoveToFront(element)
		element.Value.(*regexCacheItem).regex = regex
	} else {
		// 添加新项
		if c.list.Len() >= c.capacity {
			// 移除最久未使用的项
			back := c.list.Back()
			if back != nil {
				c.list.Remove(back)
				item := back.Value.(*regexCacheItem)
				delete(c.cache, item.pattern)
			}
		}
		
		// 添加到列表头部
		item := &regexCacheItem{pattern: pattern, regex: regex}
		element := c.list.PushFront(item)
		c.cache[pattern] = element
	}
}

// Size 获取缓存大小
func (c *LRURegexCache) Size() int {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return len(c.cache)
}

// Clear 清空缓存
func (c *LRURegexCache) Clear() {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.cache = make(map[string]*list.Element)
	c.list = list.New()
}