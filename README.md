# Mock Server

> 🚀 功能强大、灵活可配置的 Mock Server 系统
> 📊 支持多协议模拟、可视化配置和企业级部署
> 🎯 当前版本：v0.6.2

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
- **CIDR IP段** - IP白名单过滤

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
- **配置导入导出** - JSON/YAML格式，冲突处理

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

## 📖 使用示例

### 1. 创建项目
```bash
curl -X POST http://localhost:8080/api/v1/projects \
  -H "Content-Type: application/json" \
  -d '{
    "name": "测试项目",
    "workspace_id": "default"
  }'
```

### 2. 创建Mock规则
```bash
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
```

### 3. 测试Mock接口
```bash
# Mock服务格式：http://host:port/{project_id}/{environment_id}/{path}
curl http://localhost:9090/PROJECT_ID/ENV_ID/api/users
```

---

## 📚 API文档

### 🔧 核心API
| 接口 | 方法 | 说明 |
|------|------|------|
| `/api/v1/projects` | POST | 创建项目 |
| `/api/v1/rules` | POST | 创建规则 |
| `/api/v1/rules/:id` | PUT | 更新规则 |
| `/api/v1/rules/:id/enable` | POST | 启用规则 |
| `/api/v1/system/health` | GET | 健康检查 |
| `/api/v1/system/version` | GET | 版本信息 |

### 📊 统计API
| 接口 | 方法 | 说明 |
|------|------|------|
| `/api/v1/statistics/dashboard` | GET | Dashboard统计 |
| `/api/v1/statistics/projects` | GET | 项目统计 |
| `/api/v1/statistics/rules` | GET | 规则统计 |
| `/api/v1/request-logs` | GET | 请求日志 |

> **详细API文档**: 查看 [完整API文档](docs/ARCHITECTURE.md)

---

## ⚙️ 配置说明

### 基础配置
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

### 环境配置
- `config.yaml` - 生产环境
- `config.dev.yaml` - 开发环境

---

## 🛠️ 开发指南

### 项目结构
```
.
├── cmd/                     # 主程序入口
├── internal/                # 内部代码
│   ├── api/               # API处理器
│   ├── engine/            # 规则匹配引擎
│   └── service/           # 服务层
├── tests/                  # 测试框架
│   ├── scripts/           # 测试脚本
│   └── integration/      # 集成测试
├── web/frontend/           # React前端
├── docs/                   # 文档
├── Makefile               # 工程化命令
└── docker-compose.yml     # Docker配置
```

### 常用命令
```bash
# 开发命令
make start-all          # 启动全栈应用
make stop-all           # 停止所有服务
make build              # 构建后端
make test-coverage      # 测试覆盖率

# Docker命令
docker-compose up -d    # 启动服务
docker-compose down    # 停止服务
docker-compose logs -f  # 查看日志
```

### 测试框架
```bash
# 运行完整测试套件
./tests/integration/run_all_e2e_tests.sh

# 测试覆盖率报告
make test-coverage

# 查看测试文档
cat tests/README.md
```

---

## ❓ 常见问题

### 🔧 连接问题
```bash
# MongoDB连接失败
make start-mongo

# 端口冲突
make stop-all
lsof -ti:8080 | xargs kill -9
```

### 🐳 Docker问题
```bash
# 重建镜像
docker-compose down -v
docker-compose build --no-cache
docker-compose up -d
```

### 🌐 服务检查
```bash
# 健康检查
curl http://localhost:8080/api/v1/system/health

# 版本信息
curl http://localhost:8080/api/v1/system/version
```

---

## 📈 版本历史

### v0.6.2 (2025-11-18) - 结构优化
- ✅ 测试框架重组，94% E2E测试通过率
- ✅ 目录结构优化，文档体系完善
- ✅ 跨平台兼容性改进

### v0.6.0 (2025-11-17) - 企业特性
- ✅ CORS中间件支持
- ✅ 配置导入导出功能
- ✅ 统计分析增强

### v0.5.0 (2025-01-17) - 可观测性
- ✅ 请求日志系统
- ✅ Prometheus监控
- ✅ 实时统计分析

### v0.4.0 (2024-12-15) - 协议扩展
- ✅ WebSocket协议支持
- ✅ JavaScript脚本引擎

---

## 🗺️ 未来规划

### v0.7.0 - 性能优化
- 🔴 Redis缓存集成
- 🔴 数据库查询优化
- 🔴 并发性能提升

### v0.8.0 - 企业级特性
- 🔴 用户认证和权限体系
- 🔴 规则版本控制
- 🔴 多租户支持

### v0.9.0 - 协议扩展
- 🔴 gRPC协议支持
- 🔴 TCP/UDP协议支持

---

## 🤝 贡献指南

欢迎贡献代码、报告问题或提出建议！

### 贡献流程
1. Fork项目
2. 创建特性分支
3. 提交代码
4. 创建Pull Request

### 开发规范
- 遵循Go代码规范
- 添加单元测试
- 更新相关文档

> **详细贡献指南**: 查看 [CONTRIBUTING.md](CONTRIBUTING.md)

---

## 📄 许可证

MIT License - 查看 [LICENSE](LICENSE) 文件

---

## 🔗 相关链接

- **项目主页**: https://github.com/gomockserver/mockserver
- **问题反馈**: https://github.com/gomockserver/mockserver/issues
- **文档中心**: [docs/](docs/)
- **测试指南**: [tests/README.md](tests/README.md)
- **架构设计**: [docs/ARCHITECTURE.md](docs/ARCHITECTURE.md)
- **部署指南**: [DEPLOYMENT.md](DEPLOYMENT.md)

---

<div align="center">

**Mock Server** - 让API Mock变得简单而强大

[![GitHub stars](https://img.shields.io/github/stars/gomockserver/mockserver?style=social&label=Star)](https://github.com/gomockserver/mockserver)
[![GitHub forks](https://img.shields.io/github/forks/gomockserver/mockserver?style=social&label=Fork)](https://github.com/gomockserver/mockserver/fork)
[![GitHub issues](https://img.shields.io/github/issues/gomockserver/mockserver?style=social&label=Issues)](https://github.com/gomockserver/mockserver/issues)

</div>