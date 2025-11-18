# Mock Server

> 🚀 功能强大、灵活可配置的 Mock Server 系统
> 📊 支持多协议模拟、可视化配置和企业级部署
> 🎯 当前版本：v0.6.3

---

## ✨ 核心特性

### 🌐 多协议支持
- **HTTP/HTTPS** - 完整的 RESTful API Mock 能力
- **WebSocket** - 实时双向通信，支持1000+并发连接
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
- **Web管理界面** - React + TypeScript + Ant Design
- **统计分析** - 实时监控、趋势分析
- **Docker部署** - 容器化，多阶段构建

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

### Web界面管理
访问 **http://localhost:5173** 使用可视化管理界面进行项目管理、规则配置和实时监控。

---

## 📚 文档导航

| 文档 | 用途 | 读者 |
|------|------|------|
| 📖 [项目总览](PROJECT_SUMMARY.md) | 技术架构和功能详解 | 开发者、架构师 |
| 🚀 [部署指南](DEPLOYMENT.md) | 生产环境部署配置 | DevOps、运维 |
| 🔧 [贡献指南](CONTRIBUTING.md) | 开发和贡献流程 | 开源贡献者 |
| 📐 [系统架构](docs/ARCHITECTURE.md) | 详细架构设计 | 架构师、高级开发者 |
| 🧪 [测试指南](tests/README.md) | 测试框架和执行 | QA、开发者 |

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

- **v0.6.3** (2025-11-18) - 🎉 **100%测试通过率** - 企业级稳定性版本
- **v0.6.2** (2025-11-18) - 测试框架重组，目录结构优化
- **v0.6.0** (2025-11-17) - 企业特性和统计分析增强
- **v0.5.0** (2025-01-17) - 可观测性和监控支持
- **v0.4.0** (2024-12-15) - WebSocket协议支持

> 📊 **完整日志**: 查看 [CHANGELOG.md](CHANGELOG.md)

---

## 🤝 参与贡献

1. Fork 项目
2. 创建特性分支
3. 提交代码
4. 创建 Pull Request

> 🔧 **开发指南**: 查看 [CONTRIBUTING.md](CONTRIBUTING.md)

---

## 📄 开源协议

MIT License

---

<div align="center">

**Mock Server** - 让API Mock变得简单而强大

[![GitHub stars](https://img.shields.io/github/stars/gomockserver/mockserver?style=social&label=Star)](https://github.com/gomockserver/mockserver)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)

</div>