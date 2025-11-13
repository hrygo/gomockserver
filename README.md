# Mock Server

一个功能强大、灵活可配置的 Mock Server 系统，支持多协议模拟、可视化配置和多种部署方式。

## 特性

### 当前版本（v0.1.0 - MVP）

- ✅ **HTTP/HTTPS 协议支持**：完整的 HTTP Mock 能力
- ✅ **灵活的规则匹配**：支持路径、方法、Header、Query 参数匹配
- ✅ **静态响应配置**：支持 JSON、XML、HTML、Text 等多种格式
- ✅ **项目和环境管理**：支持多项目、多环境的规则隔离
- ✅ **RESTful 管理 API**：完整的规则 CRUD 接口
- ✅ **MongoDB 持久化**：企业级数据存储
- ✅ **Docker 部署**：容器化部署支持

### 未来版本规划

- 🔄 WebSocket、gRPC、TCP/UDP 协议支持
- 🔄 正则表达式和脚本化匹配
- 🔄 动态响应和模板引擎
- 🔄 Web 管理界面
- 🔄 规则版本控制
- 🔄 请求日志和统计分析
- 🔄 Redis 缓存支持

## 快速开始

### 前置要求

- Go 1.21+
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

1. 安装依赖
```bash
go mod download
```

2. 启动 MongoDB
```bash
docker run -d -p 27017:27017 --name mongodb mongo:6.0
```

3. 启动服务
```bash
go run cmd/mockserver/main.go
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
├── config.yaml              # 配置文件
├── docker-compose.yml       # Docker Compose 配置
├── Dockerfile               # Docker 镜像构建
└── README.md                # 项目文档
```

## 常见问题

### 1. MongoDB 连接失败

确保 MongoDB 服务正在运行，检查配置文件中的连接字符串是否正确。

### 2. 规则不生效

检查规则的 `enabled` 字段是否为 `true`，以及 `project_id` 和 `environment_id` 是否正确。

### 3. 端口冲突

修改 `config.yaml` 中的端口配置，或者停止占用端口的其他服务。

## 开发计划

查看 [设计文档](.qoder/quests/mock-server-implementation.md) 了解详细的实施路线图。

### 阶段二：协议扩展
- WebSocket 协议支持
- gRPC 协议支持
- TCP/UDP 协议支持

### 阶段三：高级匹配
- 正则表达式匹配
- 脚本化匹配引擎
- 动态响应模板

### 阶段四：企业特性
- Web 管理界面
- 用户权限体系
- 版本控制和回滚

## 贡献指南

欢迎贡献代码、报告问题或提出建议！

## 许可证

MIT License

## 联系方式

- 项目主页：https://github.com/gomockserver/mockserver
- 问题反馈：https://github.com/gomockserver/mockserver/issues
