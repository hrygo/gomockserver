# MockServer API 文档

> 📡 **完整的RESTful API文档** - MockServer v0.7.0
> 🚀 **支持HTTP、WebSocket、缓存管理** - 企业级API接口规范

---

## 📋 API概览

MockServer提供完整的RESTful API来管理Mock服务，包括项目管理、规则配置、实时监控等功能。

### 🌐 API基础信息

- **基础URL**: `http://localhost:8080/api/v1`
- **内容类型**: `application/json`
- **认证方式**: Bearer Token (可选)
- **API版本**: v1.0 (与v0.7.0对应)

### 📚 API文档导航

| 文档 | 描述 | 版本 |
|------|------|------|
| 🔥 **[v0.7.0 API更新](./v0.7.0_API_UPDATE.md)** | 最新版本的API变更和新功能 | v0.7.0+ |
| 📖 **核心API参考** | 基础API接口文档 | 所有版本 |
| 🎯 **WebSocket API** | WebSocket相关接口 | v0.4.0+ |
| 🗄️ **缓存API** | 缓存管理接口 | v0.7.0+ |

---

## 🔧 核心API接口

### 项目管理
```http
# 获取项目列表
GET /api/v1/projects

# 创建项目
POST /api/v1/projects
{
  "name": "项目名称",
  "workspace_id": "default"
}

# 获取项目详情
GET /api/v1/projects/{id}

# 更新项目
PUT /api/v1/projects/{id}

# 删除项目
DELETE /api/v1/projects/{id}
```

### 环境管理
```http
# 获取环境列表
GET /api/v1/environments

# 创建环境
POST /api/v1/environments
{
  "name": "环境名称",
  "project_id": "项目ID"
}

# 获取环境详情
GET /api/v1/environments/{id}

# 删除环境
DELETE /api/v1/environments/{id}
```

### 规则管理
```http
# 获取规则列表
GET /api/v1/rules

# 创建Mock规则
POST /api/v1/rules
{
  "name": "规则名称",
  "project_id": "项目ID",
  "environment_id": "环境ID",
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
      "body": {"code": 0, "data": []}
    }
  }
}

# 更新规则
PUT /api/v1/rules/{id}

# 删除规则
DELETE /api/v1/rules/{id}

# 启用/禁用规则
PATCH /api/v1/rules/{id}/toggle
```

---

## 🚀 Mock服务API

### HTTP Mock服务
```http
# 基础URL格式
http://localhost:9090/{PROJECT_ID}/{ENVIRONMENT_ID}/{REQUEST_PATH}

# 示例请求
curl http://localhost:9090/prod_123/dev_456/api/users
```

### WebSocket Mock服务
```http
# WebSocket连接URL
ws://localhost:9090/{PROJECT_ID}/{ENVIRONMENT_ID}/ws/{endpoint}

# 示例连接
curl --http1.1 -H "Connection: Upgrade" \
     -H "Upgrade: websocket" \
     -H "Sec-WebSocket-Key: test" \
     -H "Sec-WebSocket-Version: 13" \
     http://localhost:9090/prod_123/dev_456/ws/chat
```

---

## 📊 监控和统计API

### 系统健康检查
```http
# 基础健康检查 (v1)
GET /api/v1/system/health

# 增强健康检查 (v2) - 推荐使用
GET /api/v2/system/health
```

**响应示例:**
```json
{
  "status": "healthy",
  "version": "v0.7.0",
  "uptime": "72h30m15s",
  "components": {
    "database": {"status": "healthy", "response_time": "5ms"},
    "cache": {"status": "healthy", "overall_hit_rate": 0.92},
    "websocket": {"status": "operational", "active_connections": 45}
  },
  "performance": {
    "cpu_usage": "25%",
    "memory_usage": "512MB",
    "disk_usage": "2.1GB"
  }
}
```

### 系统指标
```http
# 获取系统指标
GET /api/v1/system/metrics

# 获取系统信息
GET /api/v1/system/info
```

### 使用统计
```http
# 获取使用统计
GET /api/v1/usage/stats

# 获取项目统计
GET /api/v1/projects/{id}/stats
```

---

## 🗄️ 缓存管理API (v0.7.0+)

### 缓存统计
```http
# 获取缓存统计信息
GET /api/v1/cache/stats
```

**响应示例:**
```json
{
  "l1_cache": {
    "hit_rate": 0.95,
    "miss_rate": 0.05,
    "eviction_count": 12,
    "current_size": 850
  },
  "l2_cache": {
    "hit_rate": 0.87,
    "miss_rate": 0.13,
    "connection_pool": {
      "active": 15,
      "idle": 85,
      "total": 100
    }
  },
  "overall": {
    "total_hit_rate": 0.92,
    "total_requests": 10000,
    "cache_size": 1850
  }
}
```

### 缓存操作
```http
# 清除缓存
DELETE /api/v1/cache/clear

# 预热缓存
POST /api/v1/cache/warmup
{
  "keys": ["user:123", "config:app"],
  "ttl": "1h"
}

# 获取缓存配置
GET /api/v1/cache/config

# 更新缓存配置
PUT /api/v1/cache/config
{
  "l1_memory": {
    "max_size": 2000,
    "ttl": "2h"
  },
  "l2_redis": {
    "address": "redis://localhost:6379",
    "pool_size": 200
  }
}
```

---

## 🌐 WebSocket管理API

### WebSocket统计
```http
# 获取WebSocket统计
GET /api/v1/websocket/stats
```

### WebSocket广播
```http
# 广播消息
POST /api/v1/websocket/broadcast
{
  "message": "系统维护通知",
  "target": "all",
  "type": "notification"
}
```

---

## 🔍 高级功能API

### 规则性能分析
```http
# 获取规则性能统计
GET /api/v1/rules/performance
```

### 项目缓存统计
```http
# 获取项目缓存统计
GET /api/v1/projects/{id}/cache-stats
```

### 代理模式管理
```http
# 创建代理规则
POST /api/v1/proxy/rules
{
  "name": "代理规则",
  "upstream": "https://api.example.com",
  "path_rewrite": {
    "from": "/mock",
    "to": "/api"
  }
}
```

---

## 📝 响应格式

### 成功响应
```json
{
  "success": true,
  "data": {},
  "message": "操作成功",
  "timestamp": "2025-11-19T10:30:00Z"
}
```

### 错误响应
```json
{
  "success": false,
  "error": {
    "code": "VALIDATION_ERROR",
    "message": "请求参数验证失败",
    "details": {}
  },
  "timestamp": "2025-11-19T10:30:00Z"
}
```

### 分页响应
```json
{
  "success": true,
  "data": {
    "items": [],
    "pagination": {
      "page": 1,
      "page_size": 20,
      "total": 100,
      "total_pages": 5
    }
  }
}
```

---

## 🛡️ 错误代码

| 错误代码 | HTTP状态码 | 描述 |
|---------|-----------|------|
| `VALIDATION_ERROR` | 400 | 请求参数验证失败 |
| `UNAUTHORIZED` | 401 | 认证失败 |
| `FORBIDDEN` | 403 | 权限不足 |
| `NOT_FOUND` | 404 | 资源不存在 |
| `CONFLICT` | 409 | 资源冲突 |
| `INTERNAL_ERROR` | 500 | 服务器内部错误 |
| `SERVICE_UNAVAILABLE` | 503 | 服务不可用 |

---

## 🔧 SDK和工具

### JavaScript/TypeScript SDK
```javascript
import { MockServerAPI } from '@gomockserver/sdk';

const api = new MockServerAPI({
  baseURL: 'http://localhost:8080/api/v1',
  timeout: 5000
});

// 创建项目
const project = await api.projects.create({
  name: '测试项目',
  workspace_id: 'default'
});

// 创建Mock规则
const rule = await api.rules.create({
  name: '用户API',
  project_id: project.id,
  // ... 其他配置
});
```

### Python SDK
```python
from gomockserver import MockServerAPI

api = MockServerAPI(base_url="http://localhost:8080/api/v1")

# 创建项目
project = api.projects.create(
    name="测试项目",
    workspace_id="default"
)

# 获取缓存统计
stats = api.cache.get_stats()
print(f"缓存命中率: {stats.overall.total_hit_rate}")
```

### CLI工具
```bash
# 安装CLI
npm install -g @gomockserver/cli

# 配置
mock config set server.url http://localhost:8080

# 创建项目
mock project create "测试项目"

# 导入规则
mock rules import rules.json

# 运行测试
mock test run e2e/
```

---

## 🚀 快速开始

### 1. 启动服务
```bash
# Docker方式
docker-compose up -d

# 本地开发
make start-all
```

### 2. 创建第一个Mock API
```bash
# 创建项目
PROJECT_ID=$(curl -s -X POST http://localhost:8080/api/v1/projects \
  -H "Content-Type: application/json" \
  -d '{"name": "我的项目", "workspace_id": "default"}' | \
  jq -r '.data.id')

# 创建环境
ENV_ID=$(curl -s -X POST http://localhost:8080/api/v1/environments \
  -H "Content-Type: application/json" \
  -d "{\"name\": \"开发环境\", \"project_id\": \"$PROJECT_ID\"}" | \
  jq -r '.data.id')

# 创建Mock规则
curl -X POST http://localhost:8080/api/v1/rules \
  -H "Content-Type: application/json" \
  -d "{
    \"name\": \"用户列表API\",
    \"project_id\": \"$PROJECT_ID\",
    \"environment_id\": \"$ENV_ID\",
    \"protocol\": \"HTTP\",
    \"match_type\": \"Simple\",
    \"match_condition\": {
      \"method\": \"GET\",
      \"path\": \"/api/users\"
    },
    \"response\": {
      \"type\": \"Static\",
      \"content\": {
        \"status_code\": 200,
        \"content_type\": \"JSON\",
        \"body\": {\"code\": 0, \"data\": [{\"id\": 1, \"name\": \"张三\"}]}
      }
    }
  }"
```

### 3. 测试Mock API
```bash
# 测试Mock接口
curl http://localhost:9090/$PROJECT_ID/$ENV_ID/api/users

# 响应
{
  "code": 0,
  "data": [
    {"id": 1, "name": "张三"}
  ]
}
```

---

## 📚 相关文档

- **[v0.7.0 API更新详情](./v0.7.0_API_UPDATE.md)** - 最新版本API变更
- **[开发指南](../development/README.md)** - 开发环境配置
- **[部署指南](../DEPLOYMENT.md)** - 生产环境部署
- **[测试指南](../../tests/README.md)** - 测试框架使用

---

## 🔗 外部链接

- **Swagger UI**: http://localhost:8080/swagger/index.html
- **OpenAPI规范**: http://localhost:8080/api/docs/openapi.json
- **GitHub仓库**: https://github.com/hrygo/gomockserver
- **问题反馈**: https://github.com/hrygo/gomockserver/issues

---

**📊 MockServer API** - 企业级Mock Server的完整API解决方案

[![API版本](https://img.shields.io/badge/API-v1.0-blue)](./v0.7.0_API_UPDATE.md)
[![服务器版本](https://img.shields.io/badge/Server-v0.7.0-green)](../releases/RELEASE_NOTES_v0.7.0.md)