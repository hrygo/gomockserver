# MockServer 快速入门指南

> 📚 **目标读者**: 新用户、开发者、测试工程师
> ⏱️ **阅读时间**: 15分钟
> 🎯 **学习目标**: 掌握MockServer的基本使用和核心功能

---

## 📖 目录

1. [系统概述](#系统概述)
2. [环境准备](#环境准备)
3. [快速安装](#快速安装)
4. [第一个Mock API](#第一个mock-api)
5. [高级功能](#高级功能)
6. [常见问题](#常见问题)
7. [下一步](#下一步)

---

## 系统概述

MockServer是一个功能强大的API模拟工具，支持：

- 🌐 **多协议支持** - HTTP/HTTPS、WebSocket、GraphQL
- 🎯 **智能匹配** - 路径、方法、Header、Body匹配
- 📦 **动态响应** - 模板引擎、脚本支持、文件引用
- 🏢️ **企业功能** - 项目管理、缓存系统、实时监控

---

## 环境准备

### 最低要求
- **操作系统**: Linux/macOS/Windows
- **内存**: 2GB RAM
- **磁盘**: 1GB可用空间
- **网络**: 能够访问Docker Hub和npm仓库

### 软件依赖
- **Docker** 20.10+ (推荐方式)
- **Docker Compose** 2.0+ (推荐方式)
- **Go** 1.24+ (本地开发)
- **Node.js** 18+ (前端开发)
- **MongoDB** 6.0+ (如果不使用Docker)

---

## 快速安装

### 方式一：Docker Compose（推荐）

```bash
# 1. 克隆项目
git clone https://github.com/gomockserver/mockserver.git
cd mockserver

# 2. 启动所有服务
docker-compose up -d

# 3. 等待服务就绪（约30秒）
docker-compose logs -f mockserver

# 4. 验证安装
curl http://localhost:8080/api/v1/system/health
```

**服务地址**：
- 🎨 Web管理界面: http://localhost:5173
- 🔧 管理API: http://localhost:8080
- 🚀 Mock服务: http://localhost:9090

### 方式二：本地开发环境

```bash
# 1. 克隆项目
git clone https://github.com/gomockserver/mockserver.git
cd mockserver

# 2. 安装后端依赖
go mod download

# 3. 安装前端依赖
cd web/frontend && npm install && cd ../..

# 4. 启动MongoDB（需要先安装并启动MongoDB）
make start-mongo

# 5. 启动后端服务
make start-backend

# 6. 启动前端服务（新终端）
make start-frontend
```

---

## 第一个Mock API

### 使用Web界面（推荐新手）

1. **访问管理界面**
   ```
   http://localhost:5173
   ```

2. **创建项目**
   - 点击"新建项目"
   - 输入项目名称：`我的API项目`
   - 选择工作空间：`default`
   - 点击"创建"

3. **创建环境**
   - 在项目详情页点击"新建环境"
   - 输入环境名称：`开发环境`
   - 点击"创建"

4. **创建Mock规则**
   - 在环境详情页点击"新建规则"
   - 配置规则：
     ```
     规则名称: 用户列表API
     请求方法: GET
     请求路径: /api/users
     响应状态码: 200
     响应内容:
     {
       "code": 0,
       "message": "success",
       "data": [
         {"id": 1, "name": "张三", "email": "zhangsan@example.com"},
         {"id": 2, "name": "李四", "email": "lisi@example.com"}
       ]
     }
     ```
   - 点击"保存"

5. **测试Mock API**
   ```bash
   # 使用项目ID和环境ID（从URL获取）
   curl http://localhost:9090/{PROJECT_ID}/{ENV_ID}/api/users
   ```

### 使用API（推荐开发者）

```bash
# 1. 创建项目
PROJECT_RESPONSE=$(curl -s -X POST http://localhost:8080/api/v1/projects \
  -H "Content-Type: application/json" \
  -d '{"name": "API测试项目", "workspace_id": "default"}')

PROJECT_ID=$(echo $PROJECT_RESPONSE | jq -r '.data.id')

# 2. 创建环境
ENV_RESPONSE=$(curl -s -X POST http://localhost:8080/api/v1/environments \
  -H "Content-Type: application/json" \
  -d "{\"name\": \"测试环境\", \"project_id\": \"$PROJECT_ID\"}")

ENV_ID=$(echo $ENV_RESPONSE | jq -r '.data.id')

# 3. 创建Mock规则
curl -s -X POST http://localhost:8080/api/v1/rules \
  -H "Content-Type: application/json" \
  -d "{
    \"name\": \"用户信息API\",
    \"project_id\": \"$PROJECT_ID\",
    \"environment_id\": \"$ENV_ID\",
    \"protocol\": \"HTTP\",
    \"match_type\": \"Simple\",
    \"match_condition\": {
      \"method\": \"GET\",
      \"path\": \"/api/users/:id\"
    },
    \"response\": {
      \"type\": \"Template\",
      \"content\": {
        \"status_code\": 200,
        \"content_type\": \"JSON\",
        \"body\": \"{\\\"code\\\": 0, \\\"data\\\": {\\\"id\\\": {{.path.id}}, \\\"name\\\": \\\"用户{{.path.id}}\\\", \\\"email\\\": \\\"user{{.path.id}}@example.com\\\"}}\"
      }
    }
  }"

# 4. 测试Mock API
curl http://localhost:9090/$PROJECT_ID/$ENV_ID/api/users/123
```

---

## 高级功能

### 1. 动态响应模板

```json
{
  "response": {
    "type": "Template",
    "content": {
      "body": "{{if eq .header.Authorization \"Bearer token123\"}}{\"authenticated\": true, \"user\": \"admin\"}{{else}}{\"error\": \"Unauthorized\"}{{end}}"
    }
  }
}
```

### 2. JavaScript脚本匹配

```javascript
// 匹配条件脚本
function match(request) {
  const auth = request.headers.authorization;
  const token = auth ? auth.split(' ')[1] : null;

  // 验证JWT token
  if (token && token.startsWith('valid_')) {
    return true;
  }

  return false;
}
```

### 3. WebSocket Mock

```bash
# 创建WebSocket规则
curl -X POST http://localhost:8080/api/v1/rules \
  -H "Content-Type: application/json" \
  -d '{
    "name": "WebSocket聊天",
    "protocol": "WebSocket",
    "match_condition": {
      "path": "/ws/chat"
    },
    "response": {
      "type": "WebSocket",
      "content": {
        "messages": [
          {"type": "welcome", "data": "欢迎来到聊天室"},
          {"type": "notification", "data": "新用户加入"}
        ]
      }
    }
  }'
```

### 4. GraphQL API

```graphql
# 查询项目和规则
query {
  projects {
    id
    name
    environments {
      id
      name
      rules {
        id
        name
        protocol
      }
    }
  }
}

# 创建新项目
mutation {
  createProject(input: {
    name: "GraphQL项目"
    workspaceId: "default"
  }) {
    id
    name
  }
}
```

---

## 常见问题

### Q: 如何修改端口？

**A**: 编辑 `config/config.yaml` 文件：
```yaml
server:
  admin:
    port: 8080  # 修改管理端口
  mock:
    port: 9090  # 修改Mock服务端口
```

### Q: 如何启用HTTPS？

**A**: 在配置文件中添加：
```yaml
server:
  admin:
    tls:
      enabled: true
      cert_file: "path/to/cert.pem"
      key_file: "path/to/key.pem"
```

### Q: 如何导入/导出数据？

**A**: 使用API接口：
```bash
# 导出项目数据
curl http://localhost:8080/api/v1/projects/{PROJECT_ID}/export > export.json

# 导入项目数据
curl -X POST http://localhost:8080/api/v1/projects/import \
  -H "Content-Type: application/json" \
  -d @export.json
```

### Q: 如何查看日志？

**A**:
- Docker方式：`docker-compose logs -f mockserver`
- 本地方式：查看 `logs/mockserver.log`

### Q: 如何设置延迟？

**A**: 在规则响应中配置：
```json
{
  "response": {
    "delay": {
      "type": "fixed",
      "value": 1000
    }
  }
}
```

支持的延迟类型：
- `fixed`: 固定延迟（毫秒）
- `random`: 随机延迟（min-max毫秒）
- `normal`: 正态分布延迟

---

## 下一步

恭喜！您已经掌握了MockServer的基础使用。继续学习：

- 📚 [高级配置指南](ADVANCED_USAGE.md)
- 🎯 [API完整文档](../api/README.md)
- 🏗️ [架构设计文档](../ARCHITECTURE.md)
- 🧪 [测试框架使用](../tests/README.md)
- 🚀 [部署最佳实践](../DEPLOYMENT.md)

---

## 获取帮助

- 📖 [官方文档](https://docs.gomockserver.com)
- 🐛 [问题反馈](https://github.com/gomockserver/mockserver/issues)
- 💬 [社区讨论](https://github.com/gomockserver/mockserver/discussions)
- 📧 [邮件支持](mailto:support@gomockserver.com)

---

<div align="center">

**🎉 开始使用MockServer，让API开发更高效！**

[返回首页](../../README.md)

</div>