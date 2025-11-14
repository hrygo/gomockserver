# Mock Server 架构设计

## 整体架构

Mock Server 采用分层架构设计，遵循单一职责原则和依赖倒置原则。

```
┌─────────────────────────────────────────────────────────┐
│                    客户端层                                │
│          Web UI  │  HTTP Client  │  API Testing Tools    │
└────────────────────────┬────────────────────────────────┘
                         │
┌────────────────────────┴────────────────────────────────┐
│                    接入层 (API Layer)                      │
│          Admin API Server  │  Mock Server               │
│          (端口 8080)        │  (端口 9090)               │
└────────────────────────┬────────────────────────────────┘
                         │
┌────────────────────────┴────────────────────────────────┐
│                  服务层 (Service Layer)                    │
│          Admin Service  │  Mock Service                  │
└────────────────────────┬────────────────────────────────┘
                         │
┌────────────────────────┴────────────────────────────────┐
│                 核心层 (Core Layer)                        │
│     Match Engine  │  Mock Executor  │  Protocol Adapter │
└────────────────────────┬────────────────────────────────┘
                         │
┌────────────────────────┴────────────────────────────────┐
│              数据访问层 (Repository Layer)                 │
│    Rule Repository  │  Project Repository               │
└────────────────────────┬────────────────────────────────┘
                         │
┌────────────────────────┴────────────────────────────────┐
│                    数据层                                  │
│                   MongoDB 6.0+                           │
└─────────────────────────────────────────────────────────┘
```

## 核心组件

### 1. 规则匹配引擎 (Match Engine)

**位置**: `internal/engine/`

**职责**:
- 根据请求特征匹配对应的 Mock 规则
- 支持路径、方法、Header、Query 参数匹配
- 按优先级排序规则
- 处理路径参数（如 `/api/users/:id`）

**关键接口**:
```go
type MatchEngine interface {
    Match(ctx context.Context, request *adapter.Request, projectID, environmentID string) (*models.Rule, error)
}
```

### 2. Mock 执行器 (Mock Executor)

**位置**: `internal/executor/`

**职责**:
- 根据规则生成 Mock 响应
- 处理响应延迟
- 支持多种内容类型（JSON、XML、HTML、Text）
- 应用响应头和状态码

**关键接口**:
```go
type MockExecutor interface {
    Execute(request *adapter.Request, rule *models.Rule) (*adapter.Response, error)
}
```

### 3. 协议适配器 (Protocol Adapter)

**位置**: `internal/adapter/`

**职责**:
- 统一不同协议的请求/响应模型
- HTTP 请求解析和响应构建
- 提供协议无关的抽象接口

**关键接口**:
```go
type Adapter interface {
    Parse(rawRequest interface{}) (*Request, error)
    Build(response *Response) interface{}
}
```

### 4. 数据仓库层 (Repository)

**位置**: `internal/repository/`

**职责**:
- 提供数据库操作接口
- 管理 MongoDB 连接
- 创建和维护索引
- 实现 CRUD 操作

**仓库列表**:
- `RuleRepository`: 规则管理
- `ProjectRepository`: 项目管理
- `EnvironmentRepository`: 环境管理

## 数据模型

### 规则 (Rule)

```go
type Rule struct {
    ID              string           // 规则唯一标识
    Name            string           // 规则名称
    ProjectID       string           // 所属项目
    EnvironmentID   string           // 所属环境
    Protocol        ProtocolType     // 协议类型（HTTP/HTTPS）
    MatchType       MatchType        // 匹配类型（Simple）
    Priority        int              // 优先级（越大越优先）
    Enabled         bool             // 是否启用
    MatchCondition  MatchCondition   // 匹配条件
    Response        Response         // 响应配置
    CreatedAt       time.Time
    UpdatedAt       time.Time
}
```

### 匹配条件 (MatchCondition)

```go
type MatchCondition struct {
    Method      string              // HTTP 方法
    Path        string              // 路径模式
    Query       map[string]string   // Query 参数
    Headers     map[string]string   // 请求头
    IPWhitelist []string            // IP 白名单
}
```

### 响应配置 (Response)

```go
type Response struct {
    Type    ResponseType    // 响应类型（Static）
    Delay   *DelayConfig    // 延迟配置
    Content ResponseContent // 响应内容
}

type ResponseContent struct {
    StatusCode  int                 // HTTP 状态码
    ContentType ContentType         // 内容类型
    Headers     map[string]string   // 响应头
    Body        interface{}         // 响应体
}
```

## 请求处理流程

### Mock 请求流程

```
1. HTTP Request → HTTP Adapter
                   ↓
2. Parse Request → Unified Request Model
                   ↓
3. Extract Project & Environment ID from Path
                   ↓
4. Match Engine → Find Matching Rule
                   ↓
5. Mock Executor → Generate Response
                   ↓
6. Apply Delay (if configured)
                   ↓
7. HTTP Adapter → Build HTTP Response
                   ↓
8. Return Response
```

### 规则匹配算法

```
1. 加载指定项目和环境的所有启用规则
2. 过滤协议类型匹配的规则
3. 按优先级降序排序
4. 逐条检查匹配条件:
   - HTTP 方法匹配
   - 路径匹配（支持路径参数）
   - Query 参数匹配
   - Header 匹配
   - IP 白名单检查
5. 返回第一个完全匹配的规则
6. 未找到则返回 404
```

## 性能优化

### 索引设计

```javascript
// Rules 集合
db.rules.createIndex({ "project_id": 1, "environment_id": 1 })
db.rules.createIndex({ "enabled": 1 })
db.rules.createIndex({ "priority": -1 })

// Projects 集合
db.projects.createIndex({ "workspace_id": 1 })

// Environments 集合
db.environments.createIndex({ "project_id": 1 })
```

### 缓存策略

- 规则按项目和环境缓存（计划中）
- 使用 Redis 缓存热点规则（计划中）
- 连接池优化 MongoDB 访问

### 并发处理

- 使用 Goroutine 处理并发请求
- 无锁读取规则配置
- 异步日志记录（计划中）

## 扩展性设计

### 协议扩展

通过实现 `Adapter` 接口可轻松添加新协议支持：

```go
type WebSocketAdapter struct {}

func (a *WebSocketAdapter) Parse(rawRequest interface{}) (*Request, error) {
    // WebSocket 解析逻辑
}

func (a *WebSocketAdapter) Build(response *Response) interface{} {
    // WebSocket 响应构建
}
```

### 匹配策略扩展

支持添加新的匹配类型：

- Simple：简单匹配（已实现）
- Regex：正则表达式匹配（计划中）
- Script：脚本化匹配（计划中）

### 响应类型扩展

支持添加新的响应类型：

- Static：静态响应（已实现）
- Dynamic：动态模板（计划中）
- Proxy：代理转发（计划中）
- Script：脚本生成（计划中）

## 配置管理

### 配置文件结构

```yaml
server:
  admin:
    host: "0.0.0.0"
    port: 8080
  mock:
    host: "0.0.0.0"
    port: 9090

database:
  mongodb:
    uri: "mongodb://localhost:27017"
    database: "mockserver"
    pool:
      min: 10
      max: 100
    timeout: 10s

logging:
  level: "info"
  format: "json"
```

### 环境变量覆盖

支持通过环境变量覆盖配置：

```bash
MOCKSERVER_SERVER_ADMIN_PORT=8081
MOCKSERVER_DATABASE_MONGODB_URI=mongodb://prod:27017
MOCKSERVER_LOGGING_LEVEL=debug
```

## 安全考虑

### 当前实现

- IP 白名单限制
- CORS 配置
- 输入验证
- MongoDB 注入防护

### 未来计划

- JWT 认证
- API Key 管理
- 权限控制
- 审计日志
- 速率限制

## 监控和可观测性

### 日志

使用 Zap 结构化日志：

```go
logger.Info("request matched",
    zap.String("rule_id", rule.ID),
    zap.String("path", request.Path),
    zap.Int("status_code", response.StatusCode))
```

### 健康检查

```
GET /api/v1/system/health
GET /api/v1/system/version
```

### 指标（计划中）

- 请求计数
- 响应时间
- 错误率
- 规则匹配率

## 部署架构

### 单机部署

```
┌─────────────┐
│ Mock Server │ ← Client
│  (Docker)   │
└──────┬──────┘
       │
┌──────┴──────┐
│   MongoDB   │
└─────────────┘
```

### 集群部署（计划中）

```
                     ┌─────────────┐
          ┌─────────►│ Mock Server │
          │          │   Node 1    │
┌─────────┴──┐       └──────┬──────┘
│  Load      │              │
│  Balancer  ├─────────────►│ Mock Server │
│            │              │   Node 2    │
└─────────┬──┘              └──────┬──────┘
          │                        │
          │          ┌─────────────┴──────┐
          │          │    MongoDB         │
          │          │    Replica Set     │
          │          └────────────────────┘
          │          ┌────────────────────┐
          └─────────►│    Redis Cache     │
                     └────────────────────┘
```

## 技术栈

- **语言**: Go 1.21+
- **Web 框架**: Gin
- **数据库**: MongoDB 6.0+
- **配置**: Viper
- **日志**: Zap
- **测试**: testify, gomock
- **容器**: Docker, Docker Compose

## 代码组织

```
gomockserver/
├── cmd/                    # 可执行程序
│   └── mockserver/        # 主程序入口
├── internal/              # 内部代码
│   ├── adapter/          # 协议适配器
│   ├── api/              # API 处理器
│   ├── config/           # 配置管理
│   ├── engine/           # 匹配引擎
│   ├── executor/         # Mock 执行器
│   ├── models/           # 数据模型
│   ├── repository/       # 数据访问层
│   └── service/          # 服务层
├── pkg/                   # 公共库
│   ├── logger/           # 日志工具
│   └── utils/            # 工具函数
└── tests/                 # 测试
    ├── integration/      # 集成测试
    ├── performance/      # 性能测试
    └── smoke/            # 冒烟测试
```

## 最佳实践

1. **依赖注入**: 使用构造函数注入依赖
2. **接口抽象**: 面向接口编程，便于测试和扩展
3. **错误处理**: 使用 `fmt.Errorf` 包装错误，保留错误链
4. **上下文传递**: 使用 `context.Context` 管理请求生命周期
5. **配置管理**: 配置可覆盖，支持多环境
6. **日志规范**: 结构化日志，便于查询和分析
7. **测试覆盖**: 核心模块测试覆盖率 > 80%

## 性能指标

### 当前性能

- QPS: > 10,000
- 平均响应时间: < 10ms
- P99 响应时间: < 50ms
- 支持规则数: > 10,000

### 优化方向

- 规则缓存
- 连接池调优
- 异步日志
- 批量操作
- 索引优化

## 参考资料

- [Gin Web Framework](https://gin-gonic.com/)
- [MongoDB Go Driver](https://www.mongodb.com/docs/drivers/go/current/)
- [Effective Go](https://golang.org/doc/effective_go.html)
- [Go Code Review Comments](https://github.com/golang/go/wiki/CodeReviewComments)
