# Mock Server

一个功能强大、灵活可配置的 Mock Server 系统，支持多协议模拟、可视化配置和多种部署方式。

## 特性

### 当前版本（v0.6.1）

- ✅ **HTTP/HTTPS 协议支持**：完整的 HTTP Mock 能力
- ✅ **WebSocket 协议支持**：实时双向通信，心跳保活，连接管理（最大1000个并发连接）
- ✅ **灵活的规则匹配**：支持路径、方法、Header、Query 参数匹配
- ✅ **正则表达式匹配**：支持正则匹配（含 LRU 缓存优化）
- ✅ **脚本匹配**：基于 JavaScript 的复杂匹配逻辑，安全沙箱隔离（5秒执行超时，10MB 内存限制）
- ✅ **CIDR IP 段匹配**：支持 IP 段白名单过滤
- ✅ **动态响应模板**：基于 Go template 引擎，13 个内置函数
- ✅ **代理模式**：HTTP 反向代理，支持请求/响应修改和错误注入
- ✅ **文件路径引用**：支持从本地文件读取响应内容
- ✅ **静态响应配置**：支持 JSON、XML、HTML、Text、二进制数据（Base64）
- ✅ **高级延迟策略**：固定、随机、正态分布、阶梯延迟（含计数器隔离）
- ✅ **项目和环境管理**：支持多项目、多环境的规则隔离
- ✅ **RESTful 管理 API**：完整的规则 CRUD 接口
- ✅ **MongoDB 持久化**：企业级数据存储
- ✅ **Docker 部署**：容器化部署支持
- ✅ **Web 管理界面**：React + TypeScript + Ant Design 5
- ✅ **统计分析 API**：Dashboard 统计、项目统计、规则统计等
- ✅ **CORS 中间件**：支持前后端分离开发（v0.6.0 新增）
- ✅ **配置导入导出**：JSON/YAML 格式，支持三种冲突策略（v0.6.0 新增）
- ✅ **统计分析增强**：协议分布、Top 项目分析（v0.6.0 新增）
- ✅ **前端环境配置**：支持开发和生产环境（v0.6.0 新增）
- ✅ **Docker 多阶段构建**：包含前端的完整镜像（v0.6.0 新增）

#### v0.6.0 新增 - 企业特性
- ✅ **CORS 中间件**：支持前后端分离开发
- ✅ **配置导入导出**：JSON/YAML 格式，支持 skip/overwrite/append 三种冲突策略
- ✅ **统计分析增强**：协议分布、Top 项目分析
- ✅ **前端环境配置**：支持开发和生产环境
- ✅ **Docker 多阶段构建**：包含前端的完整镜像
- ✅ **Makefile 优化**：新增 build-fullstack、docker-build-full 等命令
- ✅ **质量提升**：单元测试覆盖率提升至69.3%+，核心模块80%+

#### v0.5.0 新增 - 可观测性增强
- ✅ **请求日志系统**：完整的请求/响应日志记录，支持查询、过滤、统计
- ✅ **实时监控**：Prometheus 指标采集，慢请求检测，请求追踪
- ✅ **统计增强**：实时数据、趋势分析、对比分析
- ✅ **质量提升**：单元测试覆盖率达到68%+，核心模块80%+

### 未来版本规划

- 🔄 Redis 缓存支持 - v0.7.0
- 🔄 gRPC 协议支持 - v0.8.0
- 🔄 用户认证和权限管理 - v0.9.0

## 快速开始

### 前置要求

- Go 1.24+
- MongoDB 6.0+
- Docker & Docker Compose（可选）

### 使用 Docker Compose（推荐）

1. 克隆项目
```bash
git clone https://github.com/gomockserver/mockserver.git
cd mockserver
```

2. 启动服务
```bash
docker-compose up -d
```

3. 验证服务
```bash
# 检查健康状态
curl http://localhost:8080/api/v1/system/health

# 查看版本信息
curl http://localhost:8080/api/v1/system/version
```

### 本地开发

#### 方式一：一键启动（推荐）

这是最简单的启动方式，自动启动 MongoDB、后端服务和前端开发服务器。

```bash
# 一键启动全栈应用（MongoDB + 后端 + 前端）
make start-all

# 停止所有服务
make stop-all
```

**访问地址**：
- 🎨 **前端管理界面**：http://localhost:5173
- 🔧 **后端管理 API**：http://localhost:8080/api/v1
- 🚀 **Mock 服务 API**：http://localhost:9090

#### 方式二：手动启动

1. 安装依赖
```bash
go mod download
cd web && npm install && cd ..
```

2. 启动 MongoDB
```bash
make start-mongo
# 或使用 Docker
docker run -d -p 27017:27017 --name mongodb mongo:6.0
```

3. 启动后端服务
```bash
make start-backend
# 或直接运行
go run cmd/mockserver/main.go -config config.dev.yaml
```

4. 启动前端（新终端）
```bash
make start-frontend
# 或手动运行
cd web && npm run dev
```

## 使用示例

### 1. 创建项目

```bash
curl -X POST http://localhost:8080/api/v1/projects \
  -H "Content-Type: application/json" \
  -d '{
    "name": "测试项目",
    "workspace_id": "default",
    "description": "这是一个测试项目"
  }'
```

响应示例：
```json
{
  "id": "507f1f77bcf86cd799439011",
  "name": "测试项目",
  "workspace_id": "default",
  "description": "这是一个测试项目",
  "created_at": "2025-01-01T00:00:00Z",
  "updated_at": "2025-01-01T00:00:00Z"
}
```

### 2. 创建环境

```bash
curl -X POST http://localhost:8080/api/v1/environments \
  -H "Content-Type: application/json" \
  -d '{
    "name": "开发环境",
    "project_id": "507f1f77bcf86cd799439011",
    "base_url": "http://localhost:9090"
  }'
```

### 3. 创建 Mock 规则

```bash
curl -X POST http://localhost:8080/api/v1/rules \
  -H "Content-Type: application/json" \
  -d '{
    "name": "用户列表接口",
    "project_id": "507f1f77bcf86cd799439011",
    "environment_id": "507f1f77bcf86cd799439012",
    "protocol": "HTTP",
    "match_type": "Simple",
    "priority": 100,
    "enabled": true,
    "match_condition": {
      "method": "GET",
      "path": "/api/users"
    },
    "response": {
      "type": "Static",
      "content": {
        "status_code": 200,
        "content_type": "JSON",
        "headers": {
          "Content-Type": "application/json"
        },
        "body": {
          "code": 0,
          "message": "success",
          "data": [
            {
              "id": 1,
              "name": "张三",
              "email": "zhangsan@example.com"
            },
            {
              "id": 2,
              "name": "李四",
              "email": "lisi@example.com"
            }
          ]
        }
      }
    }
  }'
```

### 4. 测试 Mock 接口

```bash
curl http://localhost:9090/507f1f77bcf86cd799439011/507f1f77bcf86cd799439012/api/users
```

响应：
```json
{
  "code": 0,
  "message": "success",
  "data": [
    {
      "id": 1,
      "name": "张三",
      "email": "zhangsan@example.com"
    },
    {
      "id": 2,
      "name": "李四",
      "email": "lisi@example.com"
    }
  ]
}
```

### 5. 查询规则列表

```bash
curl "http://localhost:8080/api/v1/rules?project_id=507f1f77bcf86cd799439011&environment_id=507f1f77bcf86cd799439012"
```

### 6. 启用/禁用规则

```bash
# 禁用规则
curl -X POST http://localhost:8080/api/v1/rules/507f1f77bcf86cd799439013/disable

# 启用规则
curl -X POST http://localhost:8080/api/v1/rules/507f1f77bcf86cd799439013/enable
```

## API 文档

### 规则管理 API

| 接口 | 方法 | 说明 |
|------|------|------|
| /api/v1/rules | GET | 查询规则列表 |
| /api/v1/rules | POST | 创建规则 |
| /api/v1/rules/:id | GET | 获取规则详情 |
| /api/v1/rules/:id | PUT | 更新规则 |
| /api/v1/rules/:id | DELETE | 删除规则 |
| /api/v1/rules/:id/enable | POST | 启用规则 |
| /api/v1/rules/:id/disable | POST | 禁用规则 |

### 项目管理 API

| 接口 | 方法 | 说明 |
|------|------|------|
| /api/v1/projects | POST | 创建项目 |
| /api/v1/projects/:id | GET | 获取项目详情 |
| /api/v1/projects/:id | PUT | 更新项目 |
| /api/v1/projects/:id | DELETE | 删除项目 |

### 环境管理 API

| 接口 | 方法 | 说明 |
|------|------|------|
| /api/v1/environments | GET | 列出环境（需要 project_id 参数） |
| /api/v1/environments | POST | 创建环境 |
| /api/v1/environments/:id | GET | 获取环境详情 |
| /api/v1/environments/:id | PUT | 更新环境 |
| /api/v1/environments/:id | DELETE | 删除环境 |

### 系统管理 API

| 接口 | 方法 | 说明 |
|------|------|------|
| /api/v1/system/health | GET | 健康检查 |
| /api/v1/system/version | GET | 版本信息 |
| /api/v1/system/info | GET | 系统详细信息（v0.5.0） |

### 请求日志 API（v0.5.0 新增）

| 接口 | 方法 | 说明 |
|------|------|------|
| /api/v1/request-logs | GET | 查询请求日志（支持分页、过滤） |
| /api/v1/request-logs/:id | GET | 获取日志详情 |
| /api/v1/request-logs/cleanup | DELETE | 手动清理日志 |
| /api/v1/request-logs/statistics | GET | 日志统计 |

### 监控 API（v0.5.0 新增）

| 接口 | 方法 | 说明 |
|------|------|------|
| /api/v1/health/metrics | GET | Prometheus 指标端点 |

### 统计分析 API（v0.6.0 增强）

| 接口 | 方法 | 说明 |
|------|------|------|
| /api/v1/statistics/dashboard | GET | 获取 Dashboard 统计数据（包含协议分布、Top项目） |
| /api/v1/statistics/projects | GET | 获取项目统计列表 |
| /api/v1/statistics/rules | GET | 获取规则统计（按项目/环境分组） |
| /api/v1/statistics/request-trend | GET | 获取请求趋势数据（7天/30天） |
| /api/v1/statistics/response-time-distribution | GET | 获取响应时间分布 |

### 导入导出 API（v0.6.0 新增）

| 接口 | 方法 | 说明 |
|------|------|------|
| /api/v1/import-export/projects/:id/export | GET | 导出项目配置（含规则、环境） |
| /api/v1/import-export/rules/export | POST | 批量导出规则 |
| /api/v1/import-export/import | POST | 导入数据（支持冲突策略） |
| /api/v1/import-export/validate | POST | 验证导入数据 |

## 配置说明

配置文件位于 `config.yaml`，主要配置项：

```yaml
server:
  admin:
    host: "0.0.0.0"
    port: 8080  # 管理 API 端口
  mock:
    host: "0.0.0.0"
    port: 9090  # Mock 服务端口

database:
  mongodb:
    uri: "mongodb://localhost:27017"
    database: "mockserver"

logging:
  level: "info"  # debug, info, warn, error
  format: "json"  # json, text
```

## Mock 请求格式

Mock 服务的请求格式：

```
http://{host}:{port}/{project_id}/{environment_id}/{path}
```

例如：
```
http://localhost:9090/507f1f77bcf86cd799439011/507f1f77bcf86cd799439012/api/users
```

## 规则匹配逻辑

1. 按项目 ID 和环境 ID 过滤规则
2. 只匹配启用状态的规则
3. 按优先级从高到低排序
4. 逐条匹配规则条件
5. 返回第一个匹配成功的规则
6. 如果没有匹配的规则，返回 404

### HTTP 匹配条件

- **Method**：支持单个方法或方法数组
- **Path**：支持精确匹配和路径参数（如 `/api/users/:id`）
- **Query**：查询参数键值对匹配
- **Headers**：请求头匹配（不区分大小写）
- **IP Whitelist**：IP 白名单限制

## 项目结构

```
.
├── cmd/
│   └── mockserver/          # 主程序入口
├── internal/
│   ├── adapter/             # 协议适配器
│   ├── api/                 # API 处理器
│   ├── config/              # 配置管理
│   ├── engine/              # 规则匹配引擎
│   ├── executor/            # Mock 执行器
│   ├── models/              # 数据模型
│   ├── repository/          # 数据访问层
│   └── service/             # 服务层
├── pkg/
│   ├── logger/              # 日志工具
│   └── utils/               # 通用工具
├── web/                     # 前端管理界面
│   └── frontend/            # React + TypeScript 前端项目
│       ├── src/             # 源码目录
│       │   ├── api/         # API 接口层
│       │   ├── components/  # 通用组件
│       │   ├── pages/       # 页面组件
│       │   ├── hooks/       # 自定义 Hooks
│       │   └── types/       # TypeScript 类型
│       └── package.json     # 前端依赖
├── config.yaml              # 生产环境配置
├── config.dev.yaml          # 开发环境配置
├── docker-compose.yml       # Docker Compose 配置
├── Dockerfile               # Docker 镜像构建
├── Makefile                 # 工程化命令
└── README.md                # 项目文档
```

## 常见问题

### 1. MongoDB 连接失败

**问题**: `dial tcp: lookup mongodb: no such host`

**解决方案**:
- 本地开发时使用 `config.dev.yaml` 配置文件（MongoDB URI 为 `localhost:27017`）
- 或使用 Docker Compose 部署（使用默认 `config.yaml`）

### 2. 规则不生效

检查规则的 `enabled` 字段是否为 `true`，以及 `project_id` 和 `environment_id` 是否正确。

### 3. 端口冲突

**问题**: `bind: address already in use`

**解决方案**:
```bash
# 停止所有服务并清理端口
make stop-all

# 或手动清理端口
lsof -ti:8080 | xargs kill -9  # 后端端口
lsof -ti:5173 | xargs kill -9  # 前端端口
lsof -ti:9090 | xargs kill -9  # Mock 服务端口
```

### 4. 前端访问 404

确保后端服务已启动，健康检查通过：
```bash
curl http://localhost:8080/api/v1/system/health
# 应返回: {"status":"healthy"}
```

## 开发计划

查看 [设计文档](.qoder/quests/mock-server-implementation.md) 了解详细的实施路线图。

### 已完成

**阶段一：MVP 版本（v0.1.0）**
- ✅ HTTP/HTTPS 协议支持
- ✅ MongoDB 持久化
- ✅ RESTful 管理 API
- ✅ Docker 部署支持

**阶段一强化：质量改进（v0.1.1）**
- ✅ 测试覆盖率提升至 70%+
- ✅ 统一错误码体系
- ✅ 健康检查增强
- ✅ 请求追踪与性能监控

**阶段二：全栈管理界面（v0.1.3）**
- ✅ Web 管理界面（React + TypeScript + Ant Design）
- ✅ 统计分析 API（5 个统计端点）
- ✅ 一键启动脚本（make start-all）
- ✅ 开发环境配置优化

**阶段三：核心功能增强（v0.2.0）**
- ✅ CIDR IP 段匹配
- ✅ 正则表达式匹配（含 LRU 缓存）
- ✅ 二进制数据处理（Base64）
- ✅ 正态分布延迟
- ✅ 阶梯延迟

**阶段四：动态能力增强（v0.3.0）**
- ✅ 动态响应模板（Go template）
- ✅ 代理模式（Proxy）
- ✅ 文件路径引用
- ✅ 阶梯延迟优化

**阶段五：协议扩展（v0.4.0）**
- ✅ WebSocket 协议支持
- ✅ 脚本化匹配引擎

**阶段六：可观测性增强（v0.5.0）**
- ✅ 请求日志系统
- ✅ 实时监控（Prometheus）
- ✅ 统计分析增强
- ✅ 单元测试覆盖率提升68%+，核心模块80%+

**阶段七：企业特性（v0.6.0）**
- ✅ CORS 中间件（支持前后端分离）
- ✅ 配置导入导出（支持冲突策略）
- ✅ 统计分析增强（协议分布、Top项目）
- ✅ 前端环境变量配置
- ✅ Docker 多阶段构建（包含前端）
- 🗓️ 用户认证和权限体系（降低优先级，移至 v0.9.0）
- 🗓️ 规则版本控制和回滚（降低优先级，移至 v0.9.0）

### 计划中

**阶段八：性能优化（v0.7.0）**
- 🔴 Redis 缓存集成
- 🔴 数据库查询优化
- 🔴 并发优化

**阶段九：企业级特性（v0.8.0）**
- 🔴 用户认证和权限体系（从 v0.6.0 移至）
- 🔴 规则版本控制和回滚（从 v0.6.0 移至）

**阶段十：协议扩展（v0.9.0）**
- 🔴 gRPC 协议支持
- 🔴 TCP/UDP 协议支持

## 贡献指南

欢迎贡献代码、报告问题或提出建议！

## 许可证

MIT License

## 联系方式

- 项目主页：https://github.com/gomockserver/mockserver
- 问题反馈：https://github.com/gomockserver/mockserver/issues
