# Mock Server 使用指南

**版本**: 1.0.0  
**更新时间**: 2025-11-14  
**目标用户**: 开发者、测试工程师

## 📋 目录

1. [快速开始](#快速开始)
2. [基本概念](#基本概念)
3. [创建Mock规则](#创建mock规则)
4. [Mock请求](#mock请求)
5. [高级用法](#高级用法)
6. [最佳实践](#最佳实践)
7. [常见问题](#常见问题)

## 快速开始

### 5分钟上手

#### 1. 启动服务

```bash
# 使用Docker（推荐）
docker-compose up -d

# 或本地运行
go run ./cmd/mockserver -config=config.yaml
```

#### 2. 创建项目

```bash
curl -X POST http://localhost:8080/api/v1/projects \
  -H "Content-Type: application/json" \
  -d '{
    "name": "我的第一个项目",
    "workspace_id": "my-workspace"
  }'
```

响应示例：
```json
{
  "id": "6565a1b2c3d4e5f6g7h8i9j0",
  "name": "我的第一个项目",
  "workspace_id": "my-workspace",
  "created_at": "2025-11-14T10:00:00Z"
}
```

#### 3. 创建环境

```bash
curl -X POST http://localhost:8080/api/v1/projects/{PROJECT_ID}/environments \
  -H "Content-Type: application/json" \
  -d '{
    "name": "开发环境",
    "base_url": "http://dev.example.com"
  }'
```

#### 4. 创建Mock规则

```bash
curl -X POST http://localhost:8080/api/v1/rules \
  -H "Content-Type: application/json" \
  -d '{
    "name": "获取用户信息",
    "project_id": "{PROJECT_ID}",
    "environment_id": "{ENVIRONMENT_ID}",
    "protocol": "HTTP",
    "match_type": "Simple",
    "priority": 100,
    "enabled": true,
    "match_condition": {
      "method": "GET",
      "path": "/api/user/123"
    },
    "response": {
      "type": "Static",
      "content": {
        "status_code": 200,
        "content_type": "JSON",
        "body": {
          "id": 123,
          "name": "张三",
          "email": "zhangsan@example.com"
        }
      }
    }
  }'
```

#### 5. 发起Mock请求

```bash
curl -X GET http://localhost:9090/api/user/123 \
  -H "X-Project-ID: {PROJECT_ID}" \
  -H "X-Environment-ID: {ENVIRONMENT_ID}"
```

响应：
```json
{
  "id": 123,
  "name": "张三",
  "email": "zhangsan@example.com"
}
```

🎉 **恭喜！** 你已经成功创建了第一个Mock规则！

## 基本概念

### 核心概念

```
工作空间 (Workspace)
  └─ 项目 (Project)
      └─ 环境 (Environment)
          └─ 规则 (Rule)
              ├─ 匹配条件 (Match Condition)
              └─ 响应配置 (Response)
```

#### 工作空间 (Workspace)
- 组织的最高层级
- 用于隔离不同团队或业务线
- 一个工作空间可以包含多个项目

#### 项目 (Project)
- 代表一个具体的应用或服务
- 包含多个环境
- 例如：用户服务、订单服务

#### 环境 (Environment)
- 项目的不同运行环境
- 例如：开发、测试、预发布
- 每个环境可以有独立的配置

#### 规则 (Rule)
- 定义如何匹配请求和返回响应
- 包含匹配条件和响应配置
- 支持优先级排序

### 协议支持

当前支持的协议：
- ✅ **HTTP/HTTPS** - 完全支持
- 🚧 gRPC - 计划中
- 🚧 WebSocket - 计划中
- 🚧 TCP/UDP - 计划中

### 匹配类型

| 类型 | 说明 | 使用场景 |
|------|------|---------|
| Simple | 简单匹配 | 精确匹配路径、方法等 |
| Regex | 正则表达式 | 复杂路径模式 |
| Script | 脚本匹配 | 自定义匹配逻辑 |

### 响应类型

| 类型 | 说明 | 使用场景 |
|------|------|---------|
| Static | 静态响应 | 固定的返回数据 |
| Dynamic | 动态响应 | 根据请求生成响应 |
| Proxy | 代理转发 | 转发到真实服务 |
| Script | 脚本响应 | 自定义响应逻辑 |

## 创建Mock规则

### HTTP Mock 规则

#### 基础规则

```json
{
  "name": "基础GET请求",
  "protocol": "HTTP",
  "match_type": "Simple",
  "match_condition": {
    "method": "GET",
    "path": "/api/hello"
  },
  "response": {
    "type": "Static",
    "content": {
      "status_code": 200,
      "content_type": "JSON",
      "body": {
        "message": "Hello, World!"
      }
    }
  }
}
```

#### 带参数的规则

```json
{
  "name": "带Query参数",
  "match_condition": {
    "method": "GET",
    "path": "/api/users",
    "query": {
      "page": "1",
      "size": "10"
    }
  },
  "response": {
    "type": "Static",
    "content": {
      "status_code": 200,
      "content_type": "JSON",
      "body": {
        "page": 1,
        "size": 10,
        "total": 100,
        "data": [
          {"id": 1, "name": "用户1"},
          {"id": 2, "name": "用户2"}
        ]
      }
    }
  }
}
```

#### 带Header验证

```json
{
  "name": "需要认证的请求",
  "match_condition": {
    "method": "GET",
    "path": "/api/secure/data",
    "headers": {
      "Authorization": "Bearer token123"
    }
  },
  "response": {
    "type": "Static",
    "content": {
      "status_code": 200,
      "content_type": "JSON",
      "body": {
        "secure_data": "这是受保护的数据"
      }
    }
  }
}
```

#### POST请求

```json
{
  "name": "创建用户",
  "match_condition": {
    "method": "POST",
    "path": "/api/users",
    "body": {
      "name": "required"
    }
  },
  "response": {
    "type": "Static",
    "content": {
      "status_code": 201,
      "content_type": "JSON",
      "headers": {
        "Location": "/api/users/123"
      },
      "body": {
        "id": 123,
        "name": "新用户",
        "created_at": "2025-11-14T10:00:00Z"
      }
    }
  }
}
```

### 响应配置

#### 不同内容类型

**JSON响应**:
```json
{
  "content_type": "JSON",
  "body": {
    "key": "value"
  }
}
```

**XML响应**:
```json
{
  "content_type": "XML",
  "body": "<?xml version=\"1.0\"?><root><item>value</item></root>"
}
```

**HTML响应**:
```json
{
  "content_type": "HTML",
  "body": "<html><body><h1>Hello</h1></body></html>"
}
```

**纯文本**:
```json
{
  "content_type": "Text",
  "body": "Plain text response"
}
```

#### 自定义Headers

```json
{
  "headers": {
    "X-Custom-Header": "value",
    "Cache-Control": "no-cache",
    "Content-Language": "zh-CN"
  },
  "body": {...}
}
```

#### 延迟配置

**固定延迟**:
```json
{
  "delay": {
    "type": "fixed",
    "fixed": 1000  // 1秒
  }
}
```

**随机延迟**:
```json
{
  "delay": {
    "type": "random",
    "min": 100,   // 100ms
    "max": 500    // 500ms
  }
}
```

**正态分布延迟**（TODO）:
```json
{
  "delay": {
    "type": "normal",
    "mean": 300,
    "stddev": 50
  }
}
```

### 高级匹配

#### IP白名单

```json
{
  "match_condition": {
    "method": "GET",
    "path": "/api/internal",
    "ip_whitelist": [
      "192.168.1.100",
      "10.0.0.0/24"
    ]
  }
}
```

#### 正则表达式匹配（TODO）

```json
{
  "match_type": "Regex",
  "match_condition": {
    "method": "GET",
    "path": "/api/users/\\d+"  // 匹配 /api/users/123
  }
}
```

#### 优先级

规则按优先级从高到低匹配，数字越大优先级越高：

```json
{
  "priority": 100  // 高优先级
}
```

```json
{
  "priority": 50   // 低优先级
}
```

## Mock请求

### 请求格式

所有Mock请求都需要在Header中指定项目和环境：

```bash
curl -X GET http://localhost:9090/your/api/path \
  -H "X-Project-ID: {PROJECT_ID}" \
  -H "X-Environment-ID: {ENVIRONMENT_ID}"
```

### 响应格式

成功的Mock请求返回规则中定义的响应：

```
HTTP/1.1 200 OK
Content-Type: application/json
X-Custom-Header: value

{
  "key": "value"
}
```

### 错误响应

#### 404 - 无匹配规则

当没有规则匹配时返回：

```json
{
  "error": "No matching rule found"
}
```

#### 400 - 缺少必要Header

```json
{
  "error": "X-Project-ID header is required"
}
```

## 高级用法

### 场景1: 模拟不同的错误状态

```json
{
  "name": "服务器错误",
  "match_condition": {
    "method": "GET",
    "path": "/api/error"
  },
  "response": {
    "type": "Static",
    "content": {
      "status_code": 500,
      "content_type": "JSON",
      "body": {
        "error": "Internal Server Error",
        "message": "数据库连接失败"
      }
    }
  }
}
```

### 场景2: 模拟认证失败

```json
{
  "name": "未授权",
  "match_condition": {
    "method": "GET",
    "path": "/api/protected"
  },
  "response": {
    "type": "Static",
    "content": {
      "status_code": 401,
      "content_type": "JSON",
      "headers": {
        "WWW-Authenticate": "Bearer realm=\"API\""
      },
      "body": {
        "error": "Unauthorized",
        "message": "需要有效的访问令牌"
      }
    }
  }
}
```

### 场景3: 模拟限流

```json
{
  "name": "请求过多",
  "match_condition": {
    "method": "POST",
    "path": "/api/submit"
  },
  "response": {
    "type": "Static",
    "content": {
      "status_code": 429,
      "content_type": "JSON",
      "headers": {
        "X-RateLimit-Limit": "100",
        "X-RateLimit-Remaining": "0",
        "Retry-After": "60"
      },
      "body": {
        "error": "Too Many Requests",
        "message": "请求过于频繁，请稍后重试"
      }
    }
  }
}
```

### 场景4: 分页数据

```json
{
  "name": "用户列表-第1页",
  "match_condition": {
    "method": "GET",
    "path": "/api/users",
    "query": {
      "page": "1"
    }
  },
  "response": {
    "type": "Static",
    "content": {
      "status_code": 200,
      "content_type": "JSON",
      "body": {
        "page": 1,
        "size": 10,
        "total": 100,
        "total_pages": 10,
        "data": [
          {"id": 1, "name": "用户1"},
          {"id": 2, "name": "用户2"},
          "..."
        ],
        "links": {
          "next": "/api/users?page=2",
          "last": "/api/users?page=10"
        }
      }
    }
  }
}
```

### 场景5: 复杂业务逻辑

```json
{
  "name": "订单详情",
  "match_condition": {
    "method": "GET",
    "path": "/api/orders/12345"
  },
  "response": {
    "type": "Static",
    "content": {
      "status_code": 200,
      "content_type": "JSON",
      "body": {
        "order_id": "12345",
        "status": "shipped",
        "customer": {
          "id": 789,
          "name": "李四",
          "email": "lisi@example.com"
        },
        "items": [
          {
            "product_id": "P001",
            "name": "商品A",
            "quantity": 2,
            "price": 99.99
          }
        ],
        "total_amount": 199.98,
        "shipping_address": {
          "province": "北京",
          "city": "北京市",
          "district": "朝阳区",
          "detail": "xxx路xxx号"
        },
        "timeline": [
          {
            "status": "created",
            "timestamp": "2025-11-01T10:00:00Z"
          },
          {
            "status": "paid",
            "timestamp": "2025-11-01T10:05:00Z"
          },
          {
            "status": "shipped",
            "timestamp": "2025-11-02T09:00:00Z"
          }
        ]
      }
    }
  }
}
```

## 最佳实践

### 1. 命名规范

✅ **好的命名**:
- `获取用户列表`
- `创建订单-成功场景`
- `更新配置-权限不足`

❌ **不好的命名**:
- `test1`
- `规则1`
- `abc`

### 2. 优先级设置

```
1000+ : 特殊情况（错误场景、IP限制等）
100-999: 具体路径匹配
1-99  : 通配符匹配
```

### 3. 环境隔离

为不同环境创建独立的规则集：

```
开发环境 (dev)   - 返回详细信息，包含调试数据
测试环境 (test)  - 模拟各种场景
预发布 (staging) - 接近生产的数据
```

### 4. 版本控制

利用项目的版本控制功能：

- 重要变更前创建版本
- 记录变更原因
- 便于回滚

### 5. 文档化

为每个规则添加描述：

```json
{
  "name": "获取用户信息",
  "description": "返回指定ID的用户详细信息，包含基本信息、权限和偏好设置"
}
```

### 6. 测试数据真实性

Mock数据应该尽可能真实：

✅ **真实的数据结构**:
```json
{
  "email": "user@example.com",
  "phone": "+86 138-0000-0000",
  "created_at": "2025-11-14T10:00:00Z"
}
```

❌ **假数据**:
```json
{
  "email": "test@test.com",
  "phone": "123456",
  "created_at": "2020-01-01"
}
```

### 7. 错误场景覆盖

除了正常场景，也要覆盖错误场景：

- 400 - 参数错误
- 401 - 未授权
- 403 - 权限不足
- 404 - 资源不存在
- 429 - 请求过多
- 500 - 服务器错误
- 503 - 服务不可用

## 常见问题

### Q1: 为什么我的请求返回404？

**原因**:
1. 未指定 `X-Project-ID` 或 `X-Environment-ID`
2. 项目ID或环境ID错误
3. 规则未启用 (`enabled: false`)
4. 匹配条件不符合

**解决方案**:
```bash
# 检查Header
curl -v -H "X-Project-ID: xxx" -H "X-Environment-ID: yyy" ...

# 检查规则状态
curl http://localhost:8080/api/v1/rules/{RULE_ID}

# 查看规则列表
curl "http://localhost:8080/api/v1/rules?project_id=xxx&environment_id=yyy"
```

### Q2: 如何调试匹配失败？

1. **检查请求日志**:
```bash
curl http://localhost:8080/api/v1/logs?project_id=xxx
```

2. **启用调试日志**:
修改 `config.yaml`:
```yaml
logging:
  level: debug
```

3. **逐步验证**:
- 先创建简单规则（只匹配path）
- 逐步添加条件（method, headers, query）

### Q3: 延迟不生效？

检查延迟配置：
```json
{
  "delay": {
    "type": "fixed",  // 确保type正确
    "fixed": 1000     // 单位是毫秒
  }
}
```

### Q4: 如何模拟超时？

设置极大延迟：
```json
{
  "delay": {
    "type": "fixed",
    "fixed": 60000  // 60秒
  }
}
```

客户端应该会超时。

### Q5: 多个规则都匹配时如何选择？

按优先级（`priority`）从高到低匹配，返回第一个匹配的规则。

### Q6: 如何临时禁用规则？

```bash
curl -X PUT http://localhost:8080/api/v1/rules/{RULE_ID} \
  -H "Content-Type: application/json" \
  -d '{"enabled": false}'
```

### Q7: 支持HTTPS吗？

支持。Mock Server 的配置：

```yaml
server:
  mock:
    tls:
      enabled: true
      cert_file: /path/to/cert.pem
      key_file: /path/to/key.pem
```

### Q8: 如何批量导入规则？

使用管理API的批量接口（TODO）或脚本：

```bash
#!/bin/bash
for rule in rules/*.json; do
  curl -X POST http://localhost:8080/api/v1/rules \
    -H "Content-Type: application/json" \
    -d @$rule
done
```

## 下一步

- 📖 [测试用例文档](TEST_CASES.md)
- 🐳 [Docker测试环境](DOCKER_TEST_ENV.md)
- 🔧 [API参考文档](../api/)
- 💡 [最佳实践集锦](BEST_PRACTICES.md)

---

**文档版本**: 1.0.0  
**贡献者**: 欢迎提交改进建议  
**支持**: 查看项目 GitHub Issues
