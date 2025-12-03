# MockServer v0.7.0 缓存集成指南

**版本**: v0.7.0-alpha
**更新日期**: 2025年11月18日
**状态**: Week 1 完成 - Redis集成+三级缓存架构

---

## 📋 概述

本文档介绍MockServer v0.7.0版本中新增的三级缓存系统，包括Redis集成、智能缓存策略和性能优化功能。

## 🏗️ 缓存架构

### 三级缓存设计

```
┌─────────────────┐    ┌─────────────────┐    ┌─────────────────┐
│   L1 热点缓存    │    │   L2 温数据缓存  │    │   L3 数据库存储  │
│  (内存缓存)      │    │  (Redis缓存)    │    │  (MongoDB)      │
│                 │    │                 │    │                 │
│ • 1分钟TTL      │    │ • 10分钟TTL     │    │ • 实时查询       │
│ • 10,000条目    │    │ • 无限制        │    │ • 持久化存储     │
│ • LRU淘汰       │    │ • 智能过期      │    │ • 事务安全       │
└─────────────────┘    └─────────────────┘    └─────────────────┘
         │                       │                       │
         └───────── 智能缓存协调 ──────────────────────────┘
```

### 缓存策略

| 数据类型 | 频率阈值 | 缓存级别 | TTL | 用途 |
|---------|---------|---------|-----|------|
| **热点数据** | >80% | L1+L2 | 1分钟/10分钟 | 频繁访问的规则 |
| **温数据** | 20%-80% | L2 | 10分钟 | 中等频率数据 |
| **冷数据** | <20% | L3 | - | 低频率直接查询 |

---

## 🚀 核心功能

### 1. Redis缓存集成

#### 特性
- **高可用性**: 支持Redis集群和哨兵模式
- **数据持久化**: 支持RDB和AOF持久化
- **连接池管理**: 智能连接池，支持连接复用
- **故障转移**: 自动故障检测和恢复

#### 配置示例
```yaml
cache:
  redis:
    host: "localhost"
    port: 6379
    password: ""
    database: 1
    pool_size: 20
    min_idle_conns: 5
    dial_timeout: 5s
    read_timeout: 3s
    write_timeout: 3s
    key_prefix: "mockserver:cache:"
```

### 2. 智能缓存策略

#### 访问频率跟踪
- **实时统计**: 基于滑动窗口的访问频率计算
- **动态调整**: 根据访问模式自动调整缓存级别
- **预热机制**: 支持热点数据预加载

#### 缓存决策算法
```go
// 智能缓存决策
func determineCacheLevel(frequency float64) CacheLevel {
    switch {
    case frequency > 0.8: return L1_HOT    // 高频数据存内存
    case frequency > 0.2: return L2_WARM   // 中频数据存Redis
    default: return L3_COLD                // 低频数据直接查DB
    }
}
```

### 3. 性能监控

#### 缓存指标
- **命中率**: L1/L2/总体命中率
- **响应时间**: 各级别缓存响应时间
- **容量使用**: 内存和Redis存储使用情况
- **QPS**: 缓存系统处理能力

#### 监控API
```bash
# 获取缓存统计
GET /api/v1/cache/stats

# 获取缓存策略
GET /api/v1/cache/strategy

# 更新缓存策略
PUT /api/v1/cache/strategy

# 清空缓存
DELETE /api/v1/cache/clear?level=L1
```

---

## 📊 性能指标

### 基准测试结果

| 指标 | v0.6.4 | v0.7.0目标 | 提升幅度 |
|------|-------|-----------|---------|
| **响应时间** | 50ms | 5ms | **90%** |
| **QPS性能** | 10K | 30K | **200%** |
| **缓存命中率** | - | >80% | **新增** |
| **内存效率** | 2GB | <3GB | **智能管理** |

### 实时性能监控

```bash
# 监控缓存性能
curl http://localhost:8080/api/v1/cache/stats

# 响应示例
{
  "total_requests": 100000,
  "l1_hit_rate": 0.65,
  "l2_hit_rate": 0.25,
  "total_hit_rate": 0.90,
  "avg_response_time": "5ms",
  "l1_entries": 8500,
  "l2_entries": 1500
}
```

---

## 🔧 部署配置

### 1. Redis部署

#### Docker部署
```bash
# 启动Redis
docker run -d \
  --name mockserver-redis \
  -p 6379:6379 \
  -v redis-data:/data \
  redis:7-alpine \
  redis-server --appendonly yes
```

#### 生产环境配置
```bash
# Redis配置文件 redis.conf
maxmemory 1gb
maxmemory-policy allkeys-lru
save 900 1
save 300 10
save 60 10000
appendonly yes
appendfsync everysec
```

### 2. 应用配置

#### 配置文件更新
```yaml
# config.yaml
cache:
  enabled: true
  strategy:
    l1_max_entries: 10000
    l1_ttl: 1m
    l2_ttl: 10m
    hot_data_threshold: 0.8
    warm_data_threshold: 0.2
    preload_enabled: true
    preload_keys:
      - "rules:project:default"
      - "rules:environment:default:dev"

  redis:
    host: "${REDIS_HOST:localhost}"
    port: ${REDIS_PORT:6379}
    password: "${REDIS_PASSWORD:}"
    database: ${REDIS_DB:1}
```

#### 环境变量
```bash
# Redis连接配置
export REDIS_HOST=localhost
export REDIS_PORT=6379
export REDIS_PASSWORD=
export REDIS_DB=1

# 缓存策略配置
export CACHE_ENABLED=true
export CACHE_L1_MAX_ENTRIES=10000
export CACHE_HOT_THRESHOLD=0.8
```

### 3. Docker Compose集成

```yaml
# docker-compose.yml
version: '3.8'
services:
  mockserver:
    build: .
    ports:
      - "8080:8080"
      - "9090:9090"
    environment:
      - REDIS_HOST=redis
      - REDIS_PORT=6379
      - CACHE_ENABLED=true
    depends_on:
      - redis
      - mongodb

  redis:
    image: redis:7-alpine
    ports:
      - "6379:6379"
    volumes:
      - redis-data:/data
    command: redis-server --appendonly yes

  mongodb:
    image: mongo:6.0
    ports:
      - "27017:27017"
    volumes:
      - mongodb-data:/data/db

volumes:
  redis-data:
  mongodb-data:
```

---

## 🧪 测试验证

### 1. 单元测试
```bash
# 运行缓存模块测试
go test ./internal/cache/ -v

# 运行服务集成测试
go test ./internal/service/ -v
```

### 2. 性能测试
```bash
# 缓存性能基准测试
go test -bench=BenchmarkThreeLevelCacheManager_Get ./internal/cache/

# 响应时间测试
curl -w "@curl-format.txt" -s http://localhost:9090/default/dev/api/users
```

### 3. 缓存效果验证
```bash
# 多次请求相同资源
for i in {1..10}; do
  curl -s http://localhost:9090/default/dev/api/users > /dev/null
done

# 查看缓存统计
curl http://localhost:8080/api/v1/cache/stats
```

---

## 🔍 故障排查

### 1. Redis连接问题

#### 检查Redis状态
```bash
# 测试Redis连接
redis-cli ping

# 检查Redis日志
docker logs mockserver-redis
```

#### 常见错误
```
Error: failed to connect to Redis: dial tcp: connection refused
```
**解决方案**: 检查Redis服务是否启动，网络连接是否正常

### 2. 缓存性能问题

#### 监控指标异常
- **命中率低**: 检查缓存策略配置
- **响应时间长**: 检查网络延迟和Redis性能
- **内存不足**: 调整L1缓存大小限制

### 3. 数据一致性问题

#### 缓存失效策略
- **主动失效**: 数据更新时主动清除相关缓存
- **TTL过期**: 设置合理的过期时间
- **版本控制**: 使用版本号管理缓存一致性

---

## 📈 运维指南

### 1. 监控告警

#### 关键指标
- 缓存命中率 < 70%
- 平均响应时间 > 50ms
- Redis连接失败率 > 1%
- L1缓存使用率 > 90%

#### 告警配置
```yaml
# Prometheus告警规则
groups:
  - name: mockserver-cache
    rules:
      - alert: CacheHitRateLow
        expr: cache_hit_rate < 0.7
        for: 5m

      - alert: CacheResponseTimeHigh
        expr: cache_avg_response_time_ms > 50
        for: 3m
```

### 2. 性能调优

#### Redis优化
```bash
# 优化Redis配置
CONFIG SET maxmemory-policy allkeys-lru
CONFIG SET timeout 300
CONFIG SET tcp-keepalive 300
```

#### 应用调优
```yaml
cache:
  strategy:
    l1_ttl: 30s        # 热点数据更短TTL
    l2_ttl: 5m         # 温数据适中TTL
    hot_data_threshold: 0.9    # 提高热点阈值
```

### 3. 容量规划

#### 内存使用估算
```
L1缓存: 10,000条目 × 1KB = 10MB
L2缓存: 活跃数据 × 5KB = 50MB
总内存: ~100MB (含开销)
```

#### Redis容量规划
```
小型环境: 100MB Redis内存
中型环境: 1GB Redis内存
大型环境: 10GB+ Redis内存
```

---

## 🎯 下一步计划

### Week 2: 智能缓存策略优化
- **动态阈值调整**: 基于负载自动调整阈值
- **预测性缓存**: 基于访问模式预加载
- **分布式缓存**: Redis集群支持

### Week 3: 缓存监控完善
- **可视化仪表板**: Grafana集成
- **详细性能分析**: 缓存热点分析
- **自动故障恢复**: 缓存故障自愈机制

---

## 📞 技术支持

如遇到缓存相关问题，请提供以下信息：
1. 错误日志
2. 缓存统计数据
3. Redis配置信息
4. 系统负载情况

**联系方式**: 技术团队内部支持

---

**文档版本**: v1.0
**最后更新**: 2025年11月18日
**审核状态**: 技术团队审核通过