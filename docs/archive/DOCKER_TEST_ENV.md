# Docker 测试环境文档

**创建时间**: 2025-11-14  
**版本**: 1.0.0  
**维护者**: AI Agent

## 📋 概述

Docker 测试环境提供了一个完全隔离、可重复的测试基础设施，支持单元测试、集成测试、端到端测试和性能测试。

## 🎯 设计目标

1. **环境隔离**: 测试环境与开发/生产环境完全隔离
2. **快速启动**: 一键启动完整测试环境
3. **可重复性**: 每次测试都在相同的环境中运行
4. **易于维护**: 简单的配置和管理
5. **灵活性**: 支持多种测试场景

## 📁 文件结构

```
gomockserver/
├── docker-compose.test.yml      # 测试环境 Docker Compose 配置
├── Dockerfile.test              # 测试版本 Mock Server 镜像
├── Dockerfile.test-runner       # 测试运行器镜像
├── config.test.yaml             # 测试环境配置文件
├── scripts/
│   └── test-env.sh             # 测试环境管理脚本
└── tests/
    └── data/
        └── init-mongo.js        # MongoDB 初始化脚本
```

## 🚀 快速开始

### 前置条件

- Docker 20.10+
- Docker Compose 2.0+
- 至少 2GB 可用内存
- 至少 5GB 可用磁盘空间

### 1. 启动基础测试环境

```bash
# 使用管理脚本
./scripts/test-env.sh up

# 或直接使用 docker-compose
docker-compose -f docker-compose.test.yml up -d
```

这将启动：
- MongoDB 测试数据库 (端口 27018)
- Mock Server 测试实例 (端口 8081, 9091)

### 2. 验证环境

```bash
# 查看运行状态
./scripts/test-env.sh ps

# 查看健康状态
docker-compose -f docker-compose.test.yml ps

# 测试 API
curl http://localhost:8081/api/v1/system/health
```

### 3. 运行测试

```bash
# 运行集成测试
./scripts/test-env.sh test

# 或手动运行
docker-compose -f docker-compose.test.yml run --rm test-runner
```

### 4. 停止环境

```bash
# 停止并删除容器
./scripts/test-env.sh down

# 清理所有数据（包括volumes）
./scripts/test-env.sh clean
```

## 🔧 服务配置

### 服务列表

| 服务名 | 镜像 | 端口映射 | 说明 |
|--------|------|---------|------|
| mongodb-test | mongo:6.0 | 27018:27017 | 测试数据库 |
| mockserver-test | custom | 8081:8080, 9091:9090 | Mock Server |
| redis-test | redis:7-alpine | 6380:6379 | Redis缓存（可选） |
| wrk-test | williamyeh/wrk | - | 性能测试工具 |
| test-runner | custom | - | 测试执行器 |

### 端口说明

**测试端口与生产端口对照**:

| 服务 | 生产端口 | 测试端口 | 说明 |
|------|---------|---------|------|
| Admin API | 8080 | 8081 | 管理接口 |
| Mock API | 9090 | 9091 | Mock服务 |
| MongoDB | 27017 | 27018 | 数据库 |
| Redis | 6379 | 6380 | 缓存 |

> 使用不同端口避免与生产环境冲突，允许同时运行

### 网络配置

所有测试服务都在独立的 `mockserver-test-network` 网络中运行：

```yaml
networks:
  test-network:
    name: mockserver-test-network
    driver: bridge
```

**优势**:
- 服务间可以通过服务名访问
- 与主机网络隔离
- 更好的安全性

## 📝 使用场景

### 场景 1: 基础测试环境

**用途**: 日常开发测试、单元测试

```bash
# 启动
./scripts/test-env.sh up

# 访问
curl http://localhost:8081/api/v1/system/health

# 停止
./scripts/test-env.sh down
```

### 场景 2: 完整测试环境（含 Redis）

**用途**: 缓存功能测试

```bash
# 启动
./scripts/test-env.sh up-full

# 或
docker-compose -f docker-compose.test.yml --profile with-redis up -d

# 访问 Redis
redis-cli -p 6380 ping
```

### 场景 3: 集成测试

**用途**: 端到端业务流程测试

```bash
# 方式1: 使用管理脚本
./scripts/test-env.sh test

# 方式2: 手动启动
docker-compose -f docker-compose.test.yml --profile integration up

# 查看测试日志
docker logs -f mockserver-test-runner
```

### 场景 4: 性能测试

**用途**: 压力测试、性能基准测试

```bash
# 启动性能测试环境
./scripts/test-env.sh up-performance

# 运行性能测试
./scripts/test-env.sh perf

# 或手动执行
docker-compose -f docker-compose.test.yml exec wrk-test wrk \
    -t4 -c100 -d30s \
    -H "X-Project-ID: test" \
    -H "X-Environment-ID: test" \
    http://mockserver-test:9090/api/test
```

## 🛠️ 管理脚本使用

### 脚本命令

`scripts/test-env.sh` 提供了完整的测试环境管理功能：

```bash
# 查看帮助
./scripts/test-env.sh help

# 启动命令
./scripts/test-env.sh up              # 基础环境
./scripts/test-env.sh up-full         # 完整环境（含Redis）
./scripts/test-env.sh up-performance  # 性能测试环境
./scripts/test-env.sh up-integration  # 集成测试环境

# 停止和清理
./scripts/test-env.sh down     # 停止环境
./scripts/test-env.sh restart  # 重启环境
./scripts/test-env.sh clean    # 清理所有数据

# 查看和调试
./scripts/test-env.sh ps               # 查看状态
./scripts/test-env.sh logs             # 查看所有日志
./scripts/test-env.sh logs mongodb-test  # 查看特定服务日志

# 执行命令
./scripts/test-env.sh exec mockserver-test sh  # 进入容器shell
./scripts/test-env.sh exec mongodb-test mongosh  # MongoDB shell

# 运行测试
./scripts/test-env.sh test  # 集成测试
./scripts/test-env.sh perf  # 性能测试

# 重新构建
./scripts/test-env.sh build  # 重新构建镜像
```

### 常用操作

#### 1. 查看日志

```bash
# 所有服务日志
./scripts/test-env.sh logs

# 特定服务日志
./scripts/test-env.sh logs mockserver-test

# 实时跟踪日志
docker-compose -f docker-compose.test.yml logs -f mockserver-test
```

#### 2. 进入容器

```bash
# 进入 Mock Server 容器
./scripts/test-env.sh exec mockserver-test sh

# 进入 MongoDB 容器
./scripts/test-env.sh exec mongodb-test bash

# 执行 MongoDB 命令
./scripts/test-env.sh exec mongodb-test mongosh mockserver_test
```

#### 3. 重启服务

```bash
# 重启所有服务
./scripts/test-env.sh restart

# 重启特定服务
docker-compose -f docker-compose.test.yml restart mockserver-test
```

## 🔍 配置说明

### config.test.yaml

测试环境配置文件的关键差异：

```yaml
# 数据库配置
database:
  mongodb:
    uri: "mongodb://mongodb-test:27017"  # 使用服务名
    database: "mockserver_test"          # 测试数据库名

# 日志配置
logging:
  level: "debug"  # 测试环境使用debug级别
  
# 性能配置
performance:
  log_retention_days: 3  # 缩短保留期
  cache:
    rule_ttl: 60  # 缩短缓存时间
  rate_limit:
    enabled: false  # 禁用限流
```

**与生产环境的差异**:

| 配置项 | 生产环境 | 测试环境 | 原因 |
|--------|---------|---------|------|
| 日志级别 | info | debug | 便于调试 |
| 数据保留期 | 7天 | 3天 | 节省空间 |
| 缓存TTL | 300s | 60s | 快速更新 |
| 限流 | 启用 | 禁用 | 避免测试干扰 |

### Docker Compose 配置

#### Healthcheck

所有服务都配置了健康检查：

```yaml
healthcheck:
  test: wget --quiet --tries=1 --spider http://localhost:8080/api/v1/system/health
  interval: 10s
  timeout: 3s
  retries: 5
  start_period: 15s
```

**好处**:
- 确保服务真正就绪后再运行测试
- 自动重启不健康的服务
- 依赖关系管理更可靠

#### Profiles

使用 Profiles 控制服务启动：

```yaml
profiles:
  - with-redis    # Redis服务
  - performance   # 性能测试工具
  - integration   # 集成测试运行器
```

**使用方式**:

```bash
# 启动基础环境（不包含 profile 服务）
docker-compose -f docker-compose.test.yml up -d

# 启动包含 Redis
docker-compose -f docker-compose.test.yml --profile with-redis up -d

# 启动性能测试环境
docker-compose -f docker-compose.test.yml --profile performance up -d
```

## 📊 测试数据管理

### MongoDB 初始化

`tests/data/init-mongo.js` 自动创建索引：

```javascript
// 项目索引
db.projects.createIndex({ "workspace_id": 1 });
db.projects.createIndex({ "created_at": -1 });

// 环境索引
db.environments.createIndex({ "project_id": 1 });
db.environments.createIndex({ "project_id": 1, "name": 1 }, { unique: true });

// 规则索引
db.rules.createIndex({ "project_id": 1, "environment_id": 1, "enabled": 1 });

// 请求日志索引（3天自动删除）
db.request_logs.createIndex({ "timestamp": -1 }, { expireAfterSeconds: 259200 });
```

### 数据卷管理

测试数据存储在命名卷中：

```yaml
volumes:
  mongodb_test_data:      # MongoDB 数据
  test_logs:              # 应用日志
  performance_results:    # 性能测试结果
  test_coverage:          # 测试覆盖率数据
```

**清理数据**:

```bash
# 清理所有测试数据
./scripts/test-env.sh clean

# 手动删除特定卷
docker volume rm mockserver_test_mongodb_data
```

## 🐛 故障排查

### 问题 1: 服务启动失败

**症状**: `docker-compose up` 后服务异常退出

**排查步骤**:

```bash
# 1. 查看日志
./scripts/test-env.sh logs

# 2. 检查容器状态
docker-compose -f docker-compose.test.yml ps

# 3. 查看特定服务日志
docker logs mockserver-test-mongodb
docker logs mockserver-test-app

# 4. 检查健康检查
docker inspect mockserver-test-app | grep -A 10 Health
```

**常见原因**:
- MongoDB 未就绪时 Mock Server 尝试连接
- 端口已被占用
- 镜像构建失败

**解决方案**:
```bash
# 停止所有容器
./scripts/test-env.sh down

# 清理数据
./scripts/test-env.sh clean

# 重新构建
./scripts/test-env.sh build

# 重新启动
./scripts/test-env.sh up
```

### 问题 2: 测试连接失败

**症状**: 集成测试无法连接到服务

**排查**:

```bash
# 检查网络
docker network ls | grep test

# 检查服务在网络中
docker network inspect mockserver-test-network

# 测试连接
docker-compose -f docker-compose.test.yml exec test-runner \
    wget -O- http://mockserver-test:8080/api/v1/system/health
```

### 问题 3: 端口冲突

**症状**: 端口已被占用

**解决**:

```bash
# 查找占用端口的进程
lsof -i :8081
lsof -i :9091
lsof -i :27018

# 修改 docker-compose.test.yml 中的端口映射
ports:
  - "18081:8080"  # 使用其他端口
  - "19091:9090"
```

### 问题 4: MongoDB 连接超时

**症状**: Mock Server 报错 "failed to ping MongoDB"

**原因**: MongoDB 启动较慢

**解决**:

```yaml
# 增加 healthcheck 等待时间
healthcheck:
  start_period: 30s  # 从 15s 增加到 30s
  
# 或使用 restart 策略
restart: on-failure
```

## 📈 性能优化

### 1. 镜像构建优化

使用多阶段构建减小镜像大小：

```dockerfile
# 构建阶段
FROM golang:1.21-alpine AS builder
# ... 编译

# 运行阶段
FROM alpine:latest
# 只包含运行时依赖
```

### 2. 数据卷优化

对于 macOS/Windows，使用委托提高性能：

```yaml
volumes:
  - ./config.test.yaml:/root/config.yaml:ro,delegated
  - test_logs:/root/logs:delegated
```

### 3. 资源限制

限制容器资源使用：

```yaml
deploy:
  resources:
    limits:
      cpus: '1'
      memory: 512M
    reservations:
      memory: 256M
```

## 🔐 安全考虑

### 1. 网络隔离

测试网络与主机网络隔离，只暴露必要端口。

### 2. 敏感信息

测试环境使用独立的配置文件，避免泄露生产密钥：

```yaml
security:
  jwt:
    secret: "test-secret-key-do-not-use-in-production"
```

### 3. 数据清理

定期清理测试数据：

```bash
# 每次测试后清理
./scripts/test-env.sh clean

# 或设置自动删除策略
db.request_logs.createIndex(
  { "timestamp": -1 }, 
  { expireAfterSeconds: 259200 }  // 3天
);
```

## 📚 最佳实践

### 1. 测试隔离

每个测试套件使用独立的项目ID和环境ID：

```bash
PROJECT_ID="test-$(date +%s)"
ENVIRONMENT_ID="env-$(date +%s)"
```

### 2. 数据清理

测试后清理创建的数据：

```bash
# 集成测试脚本中
cleanup() {
    curl -X DELETE "$ADMIN_API/projects/$PROJECT_ID"
}
trap cleanup EXIT
```

### 3. 环境变量

使用环境变量而非硬编码：

```bash
export ADMIN_API=http://localhost:8081/api/v1
export MOCK_API=http://localhost:9091
```

### 4. CI/CD 集成

在 CI 环境中使用：

```yaml
# GitHub Actions 示例
steps:
  - name: Start test environment
    run: ./scripts/test-env.sh up
    
  - name: Wait for services
    run: sleep 15
    
  - name: Run tests
    run: ./scripts/test-env.sh test
    
  - name: Cleanup
    if: always()
    run: ./scripts/test-env.sh down
```

## 🎓 进阶用法

### 1. 自定义镜像

修改 `Dockerfile.test` 添加额外工具：

```dockerfile
RUN apk add --no-cache \
    ca-certificates \
    tzdata \
    curl \
    wget \
    bash \
    jq \          # JSON处理
    httpie \      # HTTP客户端
    postgresql-client  # 如需连接其他数据库
```

### 2. 多环境配置

创建多个配置文件：

```bash
config.test.yaml        # 基础测试环境
config.test.perf.yaml   # 性能测试环境
config.test.stress.yaml # 压力测试环境
```

### 3. 测试数据生成

创建测试数据生成脚本：

```bash
# tests/data/seed-data.sh
#!/bin/bash

# 创建测试项目
curl -X POST $ADMIN_API/projects -d '{
  "name": "Test Project",
  "workspace_id": "test"
}'

# 创建测试规则
# ...
```

## 📞 支持和反馈

### 相关文档

- [集成测试文档](tests/integration/README.md)
- [性能测试文档](PERFORMANCE_TESTS.md)
- [CI/CD 配置](CI_CD_PIPELINE.md)

### 获取帮助

```bash
# 查看管理脚本帮助
./scripts/test-env.sh help

# 查看 Docker Compose 配置
docker-compose -f docker-compose.test.yml config

# 查看服务健康状态
docker-compose -f docker-compose.test.yml ps
```

---

**版本历史**:
- v1.0.0 (2025-11-14): 初始版本，基础测试环境配置

**维护者**: AI Agent  
**最后更新**: 2025-11-14
