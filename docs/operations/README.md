# 运维指南

> 🔧 MockServer 运维和部署指南
>
> 适合 DevOps 和运维人员

## 📚 文档列表

- [缓存集成指南](CACHE_INTEGRATION_GUIDE.md) - Redis 缓存配置和优化

---

## 🎯 运维任务

### 性能优化
- **缓存配置** - 查看 [缓存集成指南](CACHE_INTEGRATION_GUIDE.md)
- **连接池优化** - 数据库连接管理
- **并发控制** - 请求限流和熔断

### 监控告警
- **健康检查** - `/api/v1/system/health`
- **性能指标** - `/api/v1/system/stats`
- **日志监控** - 结构化日志输出

### 部署运维
- **容器化部署** - Docker 和 Kubernetes
- **服务发现** - 注册中心集成
- **负载均衡** - Nginx/HAProxy 配置

---

## 🔗 相关文档

- [部署指南](../user-guide/DEPLOYMENT.md) - 部署配置详情
- [故障排查](../getting-started/TROUBLESHOOTING.md) - 常见问题解决
- [系统架构](../architecture/) - 了解系统设计

---

## 📊 关键指标

| 指标 | 监控点 | 正常值 |
|------|--------|--------|
| CPU使用率 | 系统 | < 80% |
| 内存使用 | 系统 | < 85% |
| 请求延迟 | API | < 100ms |
| 缓存命中率 | Redis | > 90% |
| WebSocket连接 | 实时 | < 1000 |

---

<div align="center">

[返回文档中心](../README.md) | [查看缓存指南](CACHE_INTEGRATION_GUIDE.md)

</div>