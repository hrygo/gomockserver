# 正则表达式缓存优化方案

## 当前实现问题

当前的正则表达式缓存实现存在以下问题：
1. 无限制的缓存增长，可能导致内存泄漏
2. 缺乏缓存淘汰机制
3. 缺少缓存统计信息

## 优化方案

### 方案一：LRU缓存实现

```go
package engine

import (
    "container/list"
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
```

### 方案二：带统计信息的缓存

```go
// RegexCacheStats 缓存统计信息
type RegexCacheStats struct {
    Hits      int64
    Misses    int64
    Evictions int64
    Size      int
}

// MatchEngine 匹配引擎
type MatchEngine struct {
    ruleRepo     repository.RuleRepository
    regexCache   *LRURegexCache
    stats        RegexCacheStats
    statsMu      sync.RWMutex
}

// GetCacheStats 获取缓存统计信息
func (e *MatchEngine) GetCacheStats() RegexCacheStats {
    e.statsMu.RLock()
    defer e.statsMu.RUnlock()
    stats := e.stats
    stats.Size = e.regexCache.Size()
    return stats
}

// compileRegex 编译正则表达式并缓存
func (e *MatchEngine) compileRegex(pattern string) (*regexp.Regexp, error) {
    // 先尝试从缓存中获取
    if re, exists := e.regexCache.Get(pattern); exists {
        e.statsMu.Lock()
        e.stats.Hits++
        e.statsMu.Unlock()
        return re, nil
    }
    
    e.statsMu.Lock()
    e.stats.Misses++
    e.statsMu.Unlock()
    
    // 编译正则表达式
    re, err := regexp.Compile(pattern)
    if err != nil {
        return nil, err
    }
    
    // 存入缓存
    e.regexCache.Put(pattern, re)
    
    return re, nil
}
```

### 方案三：配置化的缓存策略

```go
// RegexCacheConfig 正则表达式缓存配置
type RegexCacheConfig struct {
    Enabled     bool `json:"enabled"`
    MaxSize     int  `json:"max_size"`
    MaxPatterns int  `json:"max_patterns"`
    TTL         int  `json:"ttl"` // 缓存过期时间（秒）
}

// DefaultRegexCacheConfig 默认缓存配置
func DefaultRegexCacheConfig() RegexCacheConfig {
    return RegexCacheConfig{
        Enabled:     true,
        MaxSize:     1000,
        MaxPatterns: 100,
        TTL:         3600, // 1小时
    }
}
```

## 实施建议

1. **分阶段实施**：先实现LRU缓存，再添加统计信息，最后实现配置化
2. **性能测试**：在实施前后进行性能基准测试
3. **监控指标**：添加缓存命中率等监控指标
4. **可配置性**：通过配置文件控制缓存行为

## 预期收益

1. **内存使用控制**：限制缓存大小，防止内存泄漏
2. **性能提升**：通过统计信息优化热点正则表达式
3. **可观测性**：通过监控指标了解缓存效果
4. **灵活性**：通过配置调整缓存策略