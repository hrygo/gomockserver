# 架构设计

> 🏗️ MockServer 系统架构文档
>
> 适合架构师和高级开发者了解系统设计

## 📚 文档列表

- [系统架构](ARCHITECTURE.md) - 核心架构设计详情
- [最佳实践](ARCHITECTURE_BEST_PRACTICES.md) - 架构设计原则和规范

---

## 🏗️ 架构概览

MockServer 采用分层架构设计：

```
┌─────────────────────┐
│   Web Frontend      │  React + TypeScript
└─────────┬───────────┘
          │
┌─────────▼───────────┐
│   Admin API         │  REST & GraphQL
└─────────┬───────────┘
          │
┌─────────▼───────────┐
│   Business Layer    │  Core Services
└─────────┬───────────┘
          │
┌─────────▼───────────┐
│   Engine Layer      │  Mock Engine
└─────────┬───────────┘
          │
┌─────────▼───────────┐
│   Adapter Layer     │  HTTP/WS/GraphQL
└─────────┬───────────┘
          │
┌─────────▼───────────┐
│   Data Layer        │  MongoDB + Redis
└─────────────────────┘
```

---

## 🎯 核心特性

- **多协议支持** - HTTP/HTTPS、WebSocket、GraphQL
- **智能匹配** - 灵活的规则匹配引擎
- **三级缓存** - 内存+Redis+预测性缓存
- **实时同步** - WebSocket 实时数据更新
- **插件化** - 可扩展的适配器架构

---

## 🔗 相关文档

- [项目总结](../project-docs/PROJECT_SUMMARY.md) - 技术栈和功能
- [开发指南](../developer-guide/) - 开发相关文档
- [API文档](../developer-guide/api/) - 接口设计

---

<div align="center">

[返回文档中心](../README.md) | [查看系统架构](ARCHITECTURE.md)

</div>