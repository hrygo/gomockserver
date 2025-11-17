# Mock Server v0.4.0 发布说明

## 版本概览

- **版本**: v0.4.0
- **发布日期**: 2025-01-18
- **主题**: 协议扩展
- **代码行数**: 约 600 行新增代码
- **测试覆盖率**: 保持 85%+

## 核心功能

### 1. WebSocket 协议支持

v0.4.0 版本引入了完整的 WebSocket 协议支持，使 Mock Server 能够模拟实时双向通信场景。

**主要特性**:
- 实时双向通信
- 心跳保活机制（Ping/Pong）
- 连接管理（最大1000个并发连接）
- 支持单播和广播消息
- 消息大小限制（512KB）
- 写超时控制（10秒）

**技术实现**:
- 基于 gorilla/websocket v1.5.3 库
- 读/写/心跳三个协程分离
- 可扩展消息处理器接口

### 2. 脚本匹配引擎

引入基于 JavaScript 的脚本匹配引擎，支持复杂业务逻辑匹配。

**主要特性**:
- 基于 goja JavaScript 引擎实现
- 安全沙箱环境
- 资源限制（5秒执行超时，10MB 内存限制）
- 中断机制防止无限循环
- 审计日志支持

**内置 API**:
- `request` 对象（id、protocol、path、headers、body、metadata）
- `rule` 对象（id、name、project_id、priority）
- 工具函数：log、match、hasHeader、getHeader、hasQuery、getQuery

## 技术债务解决

本版本解决了以下技术债务：
- TD-006: 脚本匹配
- TD-007: WebSocket 支持

## 新增文件

- `internal/adapter/websocket_adapter.go` - WebSocket 适配器实现
- `internal/adapter/websocket_adapter_test.go` - WebSocket 适配器测试
- `internal/engine/script_engine.go` - 脚本引擎实现
- `internal/engine/script_engine_test.go` - 脚本引擎测试

## 新增依赖

- `github.com/gorilla/websocket` v1.5.3 - WebSocket 库
- `github.com/dop251/goja` v0.0.0-20251103141225-af2ceb9156d7 - JavaScript 引擎

## 修复问题

- 修复 WebSocket 连接数限制检查中的 nil 指针错误
- 修复脚本执行超时不生效的问题

## 性能优化

- WebSocket 心跳保活减少无效连接
- 脚本执行超时中断

## 安全性增强

- JavaScript 脚本沙箱隔离
- 执行时间和内存限制
- 危险功能禁用（require、eval、Function）

## 前端增强

- 支持 WebSocket 协议类型配置
- 支持脚本匹配和脚本响应类型
- 支持所有四种延迟类型配置（固定、随机、正态分布、阶梯延迟）
- 支持 Binary 内容类型配置

## 升级指南

### 后端升级

1. 更新代码到 v0.4.0 版本
2. 运行 `go mod tidy` 更新依赖
3. 重启服务

### 前端升级

1. 更新前端代码
2. 运行 `npm install` 安装新依赖
3. 重启前端服务

## 向后兼容性

v0.4.0 版本完全向后兼容之前的版本，现有配置和规则无需修改即可正常工作。

## 已知限制

- 脚本匹配功能为实验性特性，建议在生产环境中谨慎使用
- WebSocket 连接数默认限制为 1000，可根据需要调整

## 下一版本规划 (v0.5.0)

- 请求日志系统
- 实时监控
- 统计分析增强