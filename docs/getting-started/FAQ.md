# MockServer 常见问题 FAQ

> 📝 **最后更新**: 2025年12月3日
- 🎯 **版本**: v0.8.1

---

## 📋 目录

- [安装和部署](#安装和部署)
- [配置和使用](#配置和使用)
- [API和功能](#API和功能)
- [性能和缓存](#性能和缓存)
- [WebSocket和GraphQL](#WebSocket和GraphQL)
- [测试和调试](#测试和调试)
- [开发和贡献](#开发和贡献)
- [企业版问题](#企业版问题)

---

## 安装和部署

### Q: MockServer支持哪些操作系统？
**A**: MockServer支持：
- ✅ Linux (Ubuntu 18.04+, CentOS 7+, Debian 9+)
- ✅ macOS 10.15+
- ✅ Windows 10+ (通过Docker或WSL2)

### Q: 可以在没有Docker的环境中运行吗？
**A**: 可以。MockServer提供两种部署方式：
1. **二进制部署**: 下载编译好的可执行文件
2. **源码部署**: 从源码编译

```bash
# 从源码编译
git clone https://github.com/gomockserver/mockserver.git
cd mockserver
make build
./bin/mockserver
```

### Q: 生产环境需要多少资源？
**A**: 推荐配置：
- **CPU**: 2核心+
- **内存**: 4GB+
- **磁盘**: 20GB SSD
- **网络**: 100Mbps+

### Q: 如何升级到新版本？
**A**: 升级步骤：
```bash
# 1. 备份数据
mongodump --db mockserver

# 2. 停止服务
docker-compose down

# 3. 拉取新版本
git fetch origin
git checkout v0.8.1

# 4. 重新构建
docker-compose build

# 5. 启动服务
docker-compose up -d

# 6. 验证升级
curl http://localhost:8080/api/v1/system/health
```

---

## 配置和使用

### Q: 如何配置HTTPS？
**A**: 在配置文件中添加TLS配置：
```yaml
server:
  admin:
    tls:
      enabled: true
      cert_file: "/path/to/cert.pem"
      key_file: "/path/to/key.pem"
      min_version: "1.2"
  mock:
    tls:
      enabled: true
      cert_file: "/path/to/cert.pem"
      key_file: "/path/to/key.pem"
```

### Q: 如何修改默认端口？
**A**: 编辑配置文件：
```yaml
server:
  admin:
    port: 8081  # 管理API端口
  mock:
    port: 9091  # Mock服务端口
```

### Q: 可以同时运行多个MockServer实例吗？
**A**: 可以。需要：
1. 使用不同的端口
2. 使用不同的数据库
3. 配置负载均衡器

```yaml
# 实例1配置
server:
  admin:
    port: 8080
  mock:
    port: 9090

# 实例2配置
server:
  admin:
    port: 8081
  mock:
    port: 9091
```

### Q: 如何配置认证？
**A**: MockServer支持多种认证方式：
```yaml
auth:
  enabled: true
  type: "jwt"  # jwt, basic, apikey
  jwt:
    secret: "your-secret-key"
    expiration: "24h"
  basic:
    users:
      - username: "admin"
        password: "hashed_password"
```

---

## API和功能

### Q: 如何批量导入Mock规则？
**A**: 使用导入API：
```bash
# 1. 导出现有规则
curl http://localhost:8080/api/v1/projects/{project_id}/export > rules.json

# 2. 修改规则文件
vim rules.json

# 3. 导入规则
curl -X POST http://localhost:8080/api/v1/projects/import \
  -H "Content-Type: application/json" \
  -d @rules.json
```

### Q: 支持哪些响应类型？
**A**: 支持的响应类型：
- **Static**: 静态内容
- **Template**: 模板动态内容
- **File**: 从文件读取
- **Proxy**: 代理到真实服务
- **Script**: JavaScript脚本生成
- **WebSocket**: WebSocket消息

### Q: 如何实现延迟响应？
**A**: 在规则中配置延迟：
```json
{
  "response": {
    "delay": {
      "type": "fixed",      // fixed, random, normal
      "value": 1000,        // 毫秒
      "min": 500,          // 随机延迟最小值
      "max": 2000          // 随机延迟最大值
    }
  }
}
```

### Q: 如何使用正则表达式匹配？
**A**: 设置match_type为Regex：
```json
{
  "match_type": "Regex",
  "match_condition": {
    "path": "^/api/v\\d+/users/\\d+$",
    "body": "\\bemail\\b.*\\b@test\\.com\\b"
  }
}
```

---

## 性能和缓存

### Q: 如何优化性能？
**A**: 性能优化建议：
1. **启用缓存**
```yaml
cache:
  enabled: true
  l1_cache:
    max_size: 10000
    ttl: 300s
```

2. **使用连接池**
```yaml
database:
  mongodb:
    max_pool_size: 100
    min_pool_size: 10
```

3. **启用压缩**
```yaml
server:
  enable_compression: true
  compression_level: 6
```

### Q: 缓存是如何工作的？
**A**: MockServer采用三级缓存：
- **L1 Cache**: 内存缓存（最快）
- **L2 Cache**: Redis缓存（分布式）
- **L3 Cache**: 预测性缓存（AI驱动）

### Q: 如何监控性能？
**A**: 使用监控API：
```bash
# 获取系统统计
curl http://localhost:8080/api/v1/system/stats

# 获取缓存统计
curl http://localhost:8080/api/v1/cache/stats

# 获取请求统计
curl http://localhost:8080/api/v1/statistics/requests
```

---

## WebSocket和GraphQL

### Q: 如何Mock WebSocket？
**A**: 创建WebSocket规则：
```json
{
  "protocol": "WebSocket",
  "match_condition": {
    "path": "/ws/chat"
  },
  "response": {
    "type": "WebSocket",
    "content": {
      "auto_reply": true,
      "messages": [
        {"type": "welcome", "data": "Welcome!"},
        {"type": "ping", "interval": 30}
      ]
    }
  }
}
```

### Q: GraphQL Schema如何定义？
**A**: MockServer支持两种方式：
1. **自动生成**: 基于规则自动生成Schema
2. **自定义定义**:上传GraphQL Schema文件

```bash
# 上传Schema
curl -X POST http://localhost:8080/api/v1/graphql/schema \
  -F "schema=@schema.graphql"
```

### Q: 支持GraphQL订阅吗？
**A**: 暂时不支持GraphQL订阅，但可以通过WebSocket模拟实现。

---

## 测试和调试

### Q: 如何查看请求日志？
**A**: 启用请求日志：
```yaml
logging:
  level: "info"
  log_requests: true
  log_responses: true
```

查询日志：
```bash
curl "http://localhost:8080/api/v1/logs?limit=100&level=info"
```

### Q: 如何调试规则匹配？
**A**: 使用调试模式：
```bash
# 启用调试
curl -X POST http://localhost:8080/api/v1/debug/enable

# 测试匹配
curl -X POST http://localhost:8080/api/v1/debug/match \
  -H "Content-Type: application/json" \
  -d '{
    "method": "GET",
    "path": "/api/users",
    "headers": {"Authorization": "Bearer token"}
  }'
```

### Q: 测试数据如何管理？
**A**: 使用环境隔离：
```bash
# 创建测试环境
curl -X POST http://localhost:8080/api/v1/environments \
  -d '{"name": "test", "project_id": "xxx"}'

# 使用环境变量
export MOCK_ENV=test
curl http://localhost:9090/project/test/api/users
```

---

## 开发和贡献

### Q: 如何设置开发环境？
**A**: 查看详细的开发环境搭建指南：
[开发环境搭建](docs/development/DEVELOPMENT_SETUP.md)

### Q: 代码贡献流程是什么？
**A**: 贡献流程：
1. Fork项目
2. 创建功能分支
3. 编写代码和测试
4. 提交Pull Request
5. 代码审查
6. 合并代码

### Q: 如何运行测试？
**A**: 测试命令：
```bash
# 运行所有测试
make test

# 运行单元测试
make test-unit

# 运行集成测试
make test-integration

# 查看测试覆盖率
make test-coverage
```

---

## 企业版问题

### Q: 开源版和企业版有什么区别？
**A**: 主要区别：
| 功能 | 开源版 | 企业版 |
|------|--------|--------|
| 基础Mock | ✅ | ✅ |
| 缓存系统 | ✅ | ✅ |
| 集群部署 | ❌ | ✅ |
| 高级监控 | ❌ | ✅ |
| SSO集成 | ❌ | ✅ |
| 审计日志 | ❌ | ✅ |
| 24/7支持 | ❌ | ✅ |

### Q: 如何获得企业版？
**A**: 联系销售团队：
- 📧 sales@gomockserver.com
- 🌐 https://gomockserver.com/enterprise

### Q: 可以试用企业版功能吗？
**A**: 可以申请30天免费试用：
```bash
# 申请试用令牌
curl -X POST https://api.gomockserver.com/trial \
  -d '{"email": "your@email.com", "company": "Your Company"}'
```

---

## 其他问题

### Q: 是否有Python/Java客户端？
**A**: 官方支持：
- ✅ Go SDK
- ✅ JavaScript/TypeScript SDK
- 🚧 Python SDK（开发中）
- 🚧 Java SDK（规划中）

### Q: 如何备份数据？
**A**: 备份脚本：
```bash
#!/bin/bash
# backup.sh
DATE=$(date +%Y%m%d_%H%M%S)
BACKUP_DIR="/backup/mockserver"

# 创建备份目录
mkdir -p $BACKUP_DIR

# 备份MongoDB
mongodump --db mockserver --out $BACKUP_DIR/mongo_$DATE

# 备份配置文件
cp -r config $BACKUP_DIR/config_$DATE

# 压缩备份
tar -czf $BACKUP_DIR/backup_$DATE.tar.gz $BACKUP_DIR/*_$DATE
```

### Q: 有在线演示吗？
**A**: 是的，访问：
- 🌐 https://demo.gomockserver.com
- 👤 用户名: demo
- 🔑 密码: demo123

---

## 仍需要帮助？

如果您的疑问没有在FAQ中找到答案，请通过以下方式获取帮助：

- 📖 [完整文档](https://docs.gomockserver.com)
- 💬 [GitHub Discussions](https://github.com/gomockserver/mockserver/discussions)
- 🐛 [报告问题](https://github.com/gomockserver/mockserver/issues)
- 📧 [技术支持](mailto:support@gomockserver.com)
- 💬 [Slack社区](https://gomockserver.slack.com)

---

<div align="center">

**🙏 感谢使用MockServer！**

[返回首页](../../README.md) | [快速入门](docs/guides/GETTING_STARTED.md)

</div>