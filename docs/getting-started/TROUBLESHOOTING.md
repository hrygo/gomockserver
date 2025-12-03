# MockServer 故障排查指南

> 🔧 **常见问题解决方案**
> ⏱️ **更新时间**: 2025年12月3日
> 🎯 **适用版本**: v0.8.1

---

## 📋 目录

1. [安装问题](#安装问题)
2. [启动问题](#启动问题)
3. [连接问题](#连接问题)
4. [性能问题](#性能问题)
5. [缓存问题](#缓存问题)
6. [WebSocket问题](#WebSocket问题)
7. [GraphQL问题](#GraphQL问题)
8. [测试框架问题](#测试框架问题)
9. [日志分析](#日志分析)
10. [获取帮助](#获取帮助)

---

## 安装问题

### Docker安装失败

**问题**: `docker-compose up` 失败

**解决方案**:
```bash
# 1. 检查Docker版本
docker --version
docker-compose --version

# 2. 清理旧的容器和镜像
docker-compose down -v
docker system prune -a

# 3. 重新拉取镜像
docker-compose pull

# 4. 检查端口占用
lsof -i :8080
lsof -i :9090
lsof -i :5173
lsof -i :27017

# 5. 如果端口被占用，修改docker-compose.yml
ports:
  - "8081:8080"  # 修改为其他端口
```

### 权限问题

**问题**: `permission denied` 错误

**解决方案**:
```bash
# Linux/macOS
sudo chown -R $USER:$USER .

# 给脚本执行权限
chmod +x scripts/*.sh
chmod +x tests/integration/*.sh
```

---

## 启动问题

### 服务无法启动

**症状**: 服务启动后立即退出

**排查步骤**:
```bash
# 1. 查看详细日志
docker-compose logs mockserver

# 2. 检查配置文件
cat config/config.yaml

# 3. 验证MongoDB连接
docker-compose logs mongo
```

**常见原因及解决**:

1. **MongoDB连接失败**
   ```yaml
   # config/config.yaml
   database:
     mongodb:
       uri: "mongodb://mongo:27017"  # Docker中使用服务名
   ```

2. **端口冲突**
   ```yaml
   server:
     admin:
       port: 8081  # 改为其他端口
     mock:
       port: 9091
   ```

3. **内存不足**
   ```bash
   # 增加Docker内存限制
   docker-compose up -d --scale mockserver=1
   ```

### 健康检查失败

**症状**: `/api/v1/system/health` 返回错误

**解决方案**:
```bash
# 1. 检查所有服务状态
docker-compose ps

# 2. 等待服务完全启动
sleep 30
curl http://localhost:8080/api/v1/system/health

# 3. 检查依赖服务
docker-compose exec mongo mongosh --eval "db.adminCommand('ismaster')"
```

---

## 连接问题

### API请求超时

**问题**: 请求API时超时

**解决方案**:
```bash
# 1. 检查服务是否运行
curl -v http://localhost:8080/api/v1/system/health

# 2. 增加超时时间
curl --max-time 30 http://localhost:8080/api/v1/projects

# 3. 检查防火墙设置
# Ubuntu/Debian
sudo ufw status

# CentOS/RHEL
sudo firewall-cmd --list-all
```

### 前端无法访问后端

**症状**: Web界面显示"网络错误"

**解决方案**:
```bash
# 1. 检查后端CORS配置
curl -I http://localhost:8080/api/v1/system/health
# 查看响应头是否有 Access-Control-Allow-Origin

# 2. 修改前端配置
# web/frontend/.env
VITE_API_BASE_URL=http://localhost:8080

# 3. 重新构建前端
cd web/frontend
npm run build
```

---

## 性能问题

### 响应缓慢

**排查步骤**:
```bash
# 1. 查看系统资源
docker stats

# 2. 查看响应时间
time curl http://localhost:9090/test/test/api

# 3. 分析慢查询
# MongoDB慢查询日志
docker-compose logs mongo | grep "slow query"
```

**优化建议**:
```yaml
# config/config.yaml
cache:
  enabled: true
  l1_cache:
    max_size: 1000  # 增加内存缓存
    ttl: 300s
  l2_cache:
    enabled: true
    ttl: 3600s
```

### 内存使用过高

**解决方案**:
```bash
# 1. 监控内存使用
docker stats --no-stream

# 2. 限制容器内存
# docker-compose.yml
services:
  mockserver:
    mem_limit: 1g
    memswap_limit: 1g

# 3. 调整缓存大小
cache:
  l1_cache:
    max_size: 500  # 减少缓存大小
```

---

## 缓存问题

### Redis连接失败

**问题**: Redis缓存不可用

**排查步骤**:
```bash
# 1. 检查Redis服务
docker-compose logs redis

# 2. 测试Redis连接
docker-compose exec redis redis-cli ping

# 3. 查看Redis配置
cat config/config.yaml | grep -A 10 redis
```

**解决方案**:
```yaml
# config/config.yaml
cache:
  l2_cache:
    enabled: true
    redis:
      addr: "redis:6379"
      password: ""
      db: 0
      pool_size: 10
      min_idle_conns: 5
```

### 缓存不生效

**解决方案**:
```bash
# 1. 清空缓存
curl -X POST http://localhost:8080/api/v1/cache/clear

# 2. 检查缓存统计
curl http://localhost:8080/api/v1/cache/stats

# 3. 启用调试日志
# config/config.yaml
logging:
  level: "debug"
  modules:
    - "cache"
```

---

## WebSocket问题

### 连接被拒绝

**问题**: WebSocket连接失败

**排查步骤**:
```javascript
// 浏览器控制台
// 查看错误信息
// 常见错误：WebSocket is closed before the connection is established
```

**解决方案**:
```bash
# 1. 检查WebSocket服务
curl -i -N \
  -H "Connection: Upgrade" \
  -H "Upgrade: websocket" \
  -H "Sec-WebSocket-Key: test" \
  -H "Sec-WebSocket-Version: 13" \
  http://localhost:9090/ws/test

# 2. 检查防火墙
# 确保WebSocket端口开放

# 3. 调整超时设置
# config/config.yaml
websocket:
  read_timeout: 30s
  write_timeout: 30s
  ping_period: 30s
```

### 消息发送失败

**解决方案**:
```bash
# 1. 检查连接数限制
curl http://localhost:8080/api/v1/websocket/stats

# 2. 增加连接限制
websocket:
  max_connections: 2000  # 增加连接数
```

---

## GraphQL问题

### 查询执行失败

**问题**: GraphQL查询返回错误

**排查步骤**:
```bash
# 1. 测试GraphQL端点
curl -X POST http://localhost:8080/graphql \
  -H "Content-Type: application/json" \
  -d '{"query":"{ __schema { types { name } } }"}'

# 2. 查看GraphQL日志
docker-compose logs mockserver | grep GraphQL
```

**常见错误**:
- `Cannot query field`: 字段不存在
- `Must provide query string`: 缺少查询语句
- `Syntax Error`: 语法错误

### Schema构建失败

**解决方案**:
```bash
# 1. 重新加载Schema
curl -X POST http://localhost:8080/api/v1/graphql/reload

# 2. 检查Schema定义
find . -name "*.graphql" -type f
```

---

## 测试框架问题

### 测试失败

**问题**: 集成测试失败

**排查步骤**:
```bash
# 1. 运行单个测试
./tests/integration/simple_cache_test.sh

# 2. 查看测试日志
./tests/integration/lib/test_framework.sh debug

# 3. 检查环境变量
env | grep MOCKSERVER
```

**常见解决**:
```bash
# 1. 清理测试环境
./tests/cleanup.sh

# 2. 重新初始化
./tests/integration/lib/test_framework.sh init

# 3. 跳过服务器启动
SKIP_SERVER_START=true ./tests/integration/run_all_e2e_tests.sh
```

### 变量未定义

**问题**: `TEST_PROJECT_ID` 等变量未定义

**解决方案**:
```bash
# 手动导出变量
export TEST_PROJECT_ID="test_project_$(date +%s)"
export TEST_ENV_ID="test_env_$(date +%s)"

# 或使用框架自动创建
source ./tests/integration/lib/test_framework.sh
init_test_framework
```

---

## 日志分析

### 查看实时日志

```bash
# Docker方式
docker-compose logs -f mockserver

# 本地方式
tail -f logs/mockserver.log
```

### 日志级别设置

```yaml
# config/config.yaml
logging:
  level: "info"  # debug, info, warn, error
  format: "json" # json, text
  output: "stdout" # stdout, file
```

### 关键日志位置

```
日志路径:
- Docker: docker-compose logs
- 本地: logs/
- 测试: tests/reports/
```

---

## 获取帮助

### 自动诊断

```bash
# 运行诊断脚本
./scripts/diagnose.sh
```

### 社区支持

- 📖 [官方文档](https://docs.gomockserver.com)
- 🐛 [GitHub Issues](https://github.com/gomockserver/mockserver/issues)
- 💬 [GitHub Discussions](https://github.com/gomockserver/mockserver/discussions)
- 📧 [邮件支持](mailto:support@gomockserver.com)

### 提交Issue时请包含

1. **版本信息**：
   ```bash
   curl http://localhost:8080/api/v1/system/version
   ```

2. **系统信息**：
   ```bash
   uname -a
   docker --version
   docker-compose --version
   ```

3. **配置文件**（敏感信息请脱敏）
4. **错误日志**：
   ```bash
   docker-compose logs --tail=100 mockserver
   ```

5. **重现步骤**

---

## 快速命令参考

```bash
# 服务管理
docker-compose up -d          # 启动服务
docker-compose down -v        # 停止并清理
docker-compose restart        # 重启服务
docker-compose logs -f        # 查看日志

# 健康检查
curl http://localhost:8080/api/v1/system/health
curl http://localhost:8080/api/v1/system/stats

# 缓存管理
curl -X POST http://localhost:8080/api/v1/cache/clear
curl http://localhost:8080/api/v1/cache/stats

# 测试框架
./tests/integration/run_all_e2e_tests.sh
SKIP_SERVER_START=true ./tests/integration/simple_cache_test.sh

# 常用修复
docker system prune -a         # 清理Docker
chmod +x scripts/*.sh         # 修复权限
make build                    # 重新构建
```

---

<div align="center">

**🔧 遇到问题？我们在这里帮助您！**

[返回文档首页](../../README.md) | [查看API文档](../api/README.md)

</div>