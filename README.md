# Mock Server

> 🚀 功能强大、灵活可配置的 Mock Server 系统
> 📊 支持多协议模拟、GraphQL API 和企业级部署
> 🎯 **生产就绪版本：v0.8.1**

---

## ✨ 核心特性

### 🌐 多协议支持
- **HTTP/HTTPS** - 完整的 RESTful API Mock 能力
- **WebSocket** - 实时双向通信，支持1000+并发连接
- **GraphQL API** - 现代化查询语言，实时数据同步
- **代理模式** - HTTP 反向代理，支持请求/响应修改

### 🎯 智能匹配
- **灵活规则匹配** - 路径、方法、Header、Query参数
- **正则表达式** - 复杂模式匹配，LRU缓存优化
- **脚本化匹配** - JavaScript 引擎，安全沙箱隔离

### 📦 动态响应
- **模板引擎** - Go template，13个内置函数
- **静态配置** - JSON、XML、HTML、二进制数据
- **文件引用** - 从本地文件读取响应内容
- **高级延迟** - 固定、随机、正态分布延迟

### 🏢️ 企业级功能
- **项目环境管理** - 多项目、多环境隔离
- **三级缓存架构** - 内存+Redis+预测性缓存，性能提升300%
- **现代化Web界面** - React 18 + TypeScript 5 + Apollo Client
- **实时监控仪表盘** - ECharts图表、统计分析、趋势分析
- **GraphQL管理** - 类型安全的API查询和变更
- **Docker容器化** - 生产就绪，健康检查，多阶段构建
- **性能优化** - 启动时间优化20-28%，渐进式健康检查
- **集成测试框架** - 95%+测试通过率，CI/CD就绪

---

## 🚀 快速开始

### 📋 前置要求
- **Go 1.24+**
- **MongoDB 6.0+**
- **Docker & Docker Compose** (可选)

### 🐳 Docker Compose (推荐)

```bash
# 1. 克隆项目
git clone https://github.com/gomockserver/mockserver.git
cd mockserver

# 2. 启动服务
docker-compose up -d

# 3. 验证服务
curl http://localhost:8080/api/v1/system/health
```

### 🛠️ 本地开发

#### 一键启动（最简单）
```bash
# 启动全栈应用（MongoDB + 后端 + 前端）
make start-all

# 停止所有服务
make stop-all
```

**访问地址**：
- 🎨 **前端管理界面**: http://localhost:5173
- 🔧 **后端管理API**: http://localhost:8080/api/v1
- 🚀 **Mock服务API**: http://localhost:9090

#### 手动启动
```bash
# 1. 安装依赖
go mod download
cd web && npm install && cd ..

# 2. 启动 MongoDB
make start-mongo

# 3. 启动后端服务
make start-backend

# 4. 启动前端（新终端）
make start-frontend
```

---

## 📖 基础使用

### 创建项目和规则
```bash
# 1. 创建项目
curl -X POST http://localhost:8080/api/v1/projects \
  -H "Content-Type: application/json" \
  -d '{"name": "测试项目", "workspace_id": "default"}'

# 2. 创建Mock规则
curl -X POST http://localhost:8080/api/v1/rules \
  -H "Content-Type: application/json" \
  -d '{
    "name": "用户列表接口",
    "project_id": "PROJECT_ID",
    "environment_id": "ENV_ID",
    "protocol": "HTTP",
    "match_type": "Simple",
    "match_condition": {
      "method": "GET",
      "path": "/api/users"
    },
    "response": {
      "type": "Static",
      "content": {
        "status_code": 200,
        "content_type": "JSON",
        "body": {"code": 0, "data": [{"id": 1, "name": "张三"}]}
      }
    }
  }'

# 3. 测试Mock接口
curl http://localhost:9090/PROJECT_ID/ENV_ID/api/users
```

### GraphQL API 使用
```bash
# GraphQL 查询示例
curl -X POST http://localhost:8080/graphql \
  -H "Content-Type: application/json" \
  -d '{
    "query": "query { projects { id name environments { id name } } }"
  }'

# GraphQL 变更示例
curl -X POST http://localhost:8080/graphql \
  -H "Content-Type: application/json" \
  -d '{
    "query": "mutation CreateProject($input: CreateProjectInput!) { createProject(input: $input) { id name } }",
    "variables": {"input": {"name": "GraphQL项目", "workspaceId": "default"}}
  }'
```

### Web界面管理
访问 **http://localhost:5173** 使用现代化Web界面进行：
- 🎨 **可视化管理** - 项目、环境、规则配置
- 📊 **实时监控** - ECharts图表、统计分析
- 🔍 **GraphQL管理** - 类型安全的API查询
- ⚡ **智能代码提示** - Monaco编辑器集成

---

## 📚 文档导航

| 文档 | 用途 | 读者 |
|------|------|------|
| 📖 [文档中心](docs/README.md) | 所有文档入口 | 所有用户 |
| 🎯 [快速入门](docs/getting-started/GETTING_STARTED.md) | 15分钟快速上手 | 新用户 |
| 🔧 [故障排查](docs/getting-started/TROUBLESHOOTING.md) | 常见问题解决 | 所有用户 |
| 📋 [常见问题](docs/getting-started/FAQ.md) | FAQ大全 | 所有用户 |
| 🚀 [部署指南](docs/user-guide/DEPLOYMENT.md) | 生产环境部署配置 | DevOps、运维 |
| 🔧 [贡献指南](docs/user-guide/CONTRIBUTING.md) | 开发和贡献流程 | 开源贡献者 |
| 📐 [系统架构](docs/architecture/ARCHITECTURE.md) | 详细架构设计 | 架构师、高级开发者 |
| 🛠️ [开发指南](docs/developer-guide/DEVELOPMENT_SETUP.md) | 开发环境搭建 | 开发者 |
| 📋 [更新日志](docs/project-docs/CHANGELOG.md) | 版本变更记录 | 所有用户 |
| 🧪 [测试框架](tests/README.md) | 测试框架和执行 | QA、开发者 |

---

## ⚙️ 核心配置

```yaml
server:
  admin:
    host: "0.0.0.0"
    port: 8080  # 管理 API
  mock:
    host: "0.0.0.0"
    port: 9090  # Mock 服务

database:
  mongodb:
    uri: "mongodb://localhost:27017"
    database: "mockserver"

logging:
  level: "info"
  format: "json"
```

> 📖 **完整配置**: 查看 [部署指南](DEPLOYMENT.md#配置说明)

---

## 🚀 常用命令

```bash
# 开发环境
make start-all          # 启动全栈应用
make stop-all           # 停止服务
make test-coverage      # 测试覆盖率

# Docker
docker-compose up -d    # 启动服务
docker-compose logs -f  # 查看日志

# 服务检查
curl http://localhost:8080/api/v1/system/health
```

---

## 📈 版本历史

- **v0.8.1** (2025-12-03) - 🐛 **Bug修复版本** - Redis缓存和WebSocket测试稳定性大幅提升
- **v0.8.0** (2025-12-01) - 🚀 **GraphQL查询执行引擎** - 完整的GraphQL支持，类型安全，实时数据同步
- **v0.7.0** (2025-11-20) - 📊 **三级缓存架构** - 内存+Redis+预测性缓存，性能提升300%
- **v0.6.4** (2025-11-18) - 🔧 **测试框架重大修复** - 解决变量导出和跨平台兼容问题
- **v0.6.3** (2025-11-18) - 🎉 **100%测试通过率** - 企业级稳定性版本
- **v0.6.2** (2025-11-18) - 测试框架重组，目录结构优化

> 📊 **完整日志**: 查看 [CHANGELOG.md](docs/project-docs/CHANGELOG.md)

---

## 🤝 参与贡献

1. Fork 项目
2. 创建特性分支
3. 提交代码
4. 创建 Pull Request

> 🔧 **开发指南**: 查看 [贡献指南](docs/user-guide/CONTRIBUTING.md) | [开发环境](docs/developer-guide/DEVELOPMENT_SETUP.md)

---

## 📄 开源协议

MIT License

---

<div align="center">

**Mock Server** - 让API Mock变得简单而强大

[![GitHub stars](https://img.shields.io/github/stars/gomockserver/mockserver?style=social&label=Star)](https://github.com/gomockserver/mockserver)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)

</div>