# 后端调整待办事项

## 前后端联调准备工作

### 1. CORS 配置（紧急）

为了支持前端开发环境（http://localhost:5173）调用后端 API，需要在 Go 后端添加 CORS 中间件。

**位置**: `internal/service/middleware.go` 或创建新的 CORS 中间件

**配置要求**:
```go
// 允许的源
AllowOrigins: []string{
    "http://localhost:5173",  // 前端开发服务器
    "http://localhost:8080",  // 前端生产环境（集成部署）
}

// 允许的方法
AllowMethods: []string{
    "GET", "POST", "PUT", "DELETE", "OPTIONS", "PATCH",
}

// 允许的头
AllowHeaders: []string{
    "Content-Type", 
    "Authorization",  // 预留
    "X-Request-ID",   // 前端添加的请求追踪 ID
}

// 允许凭证
AllowCredentials: true

// 暴露的头（前端可以访问）
ExposeHeaders: []string{
    "X-Request-ID",
}
```

### 2. 静态文件托管（Week 4 Day 26-28）

在 `cmd/mockserver/main.go` 中添加静态文件服务：

```go
// 静态文件托管（生产环境）
if _, err := os.Stat("web/dist"); err == nil {
    adminRouter.Static("/", "web/dist")
    adminRouter.NoRoute(func(c *gin.Context) {
        // SPA 路由回退
        c.File("web/dist/index.html")
    })
}
```

**路由优先级**:
1. API 路由 `/api/v1/*` - 最高优先级
2. 静态文件 `/*` - 次优先级
3. SPA 回退 - 404 时返回 index.html

### 3. 统计 API 开发（Week 4 Day 20-22）

需要新增的统计接口：

#### 3.1 总体统计
```
GET /api/v1/statistics/overview

Response:
{
  "project_count": 10,
  "rule_count": 100,
  "enabled_rule_count": 80,
  "disabled_rule_count": 20,
  "today_new_rule_count": 5
}
```

#### 3.2 规则分布统计
```
GET /api/v1/statistics/rules/distribution

Response:
{
  "by_protocol": {
    "HTTP": 80,
    "HTTPS": 20
  },
  "by_project": {
    "project_id_1": 50,
    "project_id_2": 30
  },
  "by_status": {
    "enabled": 80,
    "disabled": 20
  }
}
```

#### 3.3 最近活动
```
GET /api/v1/statistics/recent-activities

Response:
{
  "recent_rules": [...],      // 最近创建/更新的规则（10条）
  "recent_projects": [...]    // 最近创建的项目（10条）
}
```

### 4. 环境检查清单

**开发环境联调前检查**:
- [ ] 后端服务运行在 http://localhost:8080
- [ ] 后端已添加 CORS 中间件
- [ ] 前端服务运行在 http://localhost:5173
- [ ] 前端可以成功调用后端健康检查接口
- [ ] MongoDB 服务正常运行

**测试步骤**:
```bash
# 1. 启动 MongoDB（如果未启动）
docker-compose up -d mongodb

# 2. 启动后端服务
cd /Users/huangzhonghui/aicoding/gomockserver
go run cmd/mockserver/main.go

# 3. 启动前端开发服务器
cd web/frontend
npm run dev

# 4. 测试 API 连通性
# 在浏览器控制台执行：
fetch('/api/v1/system/health')
  .then(res => res.json())
  .then(console.log)
```

### 5. 已完成的前端配置

✅ API 代理配置（vite.config.ts）
```typescript
server: {
  proxy: {
    '/api': {
      target: 'http://localhost:8080',
      changeOrigin: true,
    },
  },
}
```

✅ API 客户端配置（src/api/client.ts）
- 统一的请求拦截器（添加 Request ID）
- 统一的响应拦截器（错误处理）
- 基础 URL 配置

✅ TypeScript 类型定义
- Project, Environment, Rule 完整类型
- API 响应类型
- 错误响应类型

### 6. 后续集成工作

**Week 1 Day 3-5**: 项目管理功能开发
- 需要后端 `/api/v1/projects` 接口正常工作
- 测试 CRUD 操作

**Week 2**: 环境和规则管理
- 需要相关 API 接口支持

**Week 4 Day 26-28**: 前后端集成部署
- 静态文件托管
- Docker 镜像更新
- Makefile 命令增强

---

**优先级**: 
1. 🔴 **CORS 配置**（立即需要，用于开发联调）
2. 🟡 **统计 API**（Week 4 需要）
3. 🟢 **静态托管**（Week 4 需要）
