# Mock Server 部署指南

本文档提供 Mock Server 的详细部署说明，包括多种部署方式和配置选项。

## 目录

- [环境要求](#环境要求)
- [部署方式](#部署方式)
  - [Docker Compose 部署（推荐）](#docker-compose-部署推荐)
  - [Docker 部署](#docker-部署)
  - [本地部署](#本地部署)
  - [Kubernetes 部署](#kubernetes-部署)
- [配置说明](#配置说明)
- [运维管理](#运维管理)
- [故障排查](#故障排查)

## 环境要求

### 最小配置
- CPU: 2核
- 内存: 2GB
- 磁盘: 10GB

### 推荐配置
- CPU: 4核
- 内存: 4GB
- 磁盘: 20GB SSD

### 软件依赖
- Docker 20.10+（容器化部署）
- Docker Compose 2.0+（容器化部署）
- Go 1.24+（源码部署）
- MongoDB 6.0+

## 部署方式

### Docker Compose 部署（推荐）

这是最简单的部署方式，适合快速开始和开发测试环境。

#### 1. 准备工作

```bash
# 克隆项目
git clone https://github.com/gomockserver/mockserver.git
cd mockserver

# 检查 docker-compose.yml
cat docker-compose.yml
```

#### 2. 启动服务

```bash
# 构建并启动所有服务
docker-compose up -d

# 查看服务状态
docker-compose ps

# 查看日志
docker-compose logs -f
```

#### 3. 验证服务

```bash
# 检查健康状态
curl http://localhost:8080/api/v1/system/health

# 查看版本信息
curl http://localhost:8080/api/v1/system/version
```

#### 4. 停止服务

```bash
# 停止服务
docker-compose stop

# 停止并删除容器
docker-compose down

# 停止并删除容器及数据卷
docker-compose down -v
```

### Docker 部署

手动使用 Docker 部署，适合已有 MongoDB 服务的场景。

#### 1. 启动 MongoDB

```bash
docker run -d \
  --name mockserver-mongodb \
  -p 27017:27017 \
  -v mongodb_data:/data/db \
  mongo:6.0
```

#### 2. 构建镜像

```bash
# 构建 Mock Server 镜像
docker build -t mockserver:latest .
```

#### 3. 启动 Mock Server

```bash
docker run -d \
  --name mockserver-app \
  -p 8080:8080 \
  -p 9090:9090 \
  -v $(pwd)/config.yaml:/root/config.yaml \
  -v $(pwd)/logs:/root/logs \
  --link mockserver-mongodb:mongodb \
  mockserver:latest
```

### 本地部署

适合开发环境或需要源码调试的场景。

#### 方式一：一键启动（推荐）

这是最简单的本地开发方式，自动启动 MongoDB、后端服务和前端开发服务器。

**前置要求**：
- Go 1.24+
- Node.js 18+
- Docker（用于 MongoDB）

**启动步骤**：

1. 克隆项目
```bash
git clone https://github.com/gomockserver/mockserver.git
cd mockserver
```

2. 安装依赖
```bash
# 安装 Go 依赖
go mod download

# 安装前端依赖
cd web/frontend
npm install
cd ../..
```

3. 一键启动所有服务
```bash
make start-all
```

这个命令会自动：
- 启动 MongoDB 容器（如果未运行）
- 启动后端服务（使用 `config.dev.yaml`）
- 启动前端开发服务器

4. 访问服务
- 🎨 **前端管理界面**：http://localhost:5173
- 🔧 **后端管理 API**：http://localhost:8080/api/v1
- 🚀 **Mock 服务 API**：http://localhost:9090

5. 停止所有服务
```bash
make stop-all
```

#### 方式二：分步启动

如果需要更细粒度的控制，可以分步启动各个组件。

1. 启动 MongoDB
```bash
make start-mongo
```

2. 启动后端（新终端）
```bash
make start-backend
```

3. 启动前端（新终端）
```bash
make start-frontend
```

4. 分别停止服务
```bash
make stop-frontend
make stop-backend
make stop-mongo
```

#### 方式三：手动启动（调试模式）

适合需要详细日志输出和调试的场景。

#### 方式三：手动启动（调试模式）

适合需要详细日志输出和调试的场景。

#### 1. 安装依赖

```bash
# 安装 Go 1.24+
# 根据你的操作系统下载并安装 Go

# 验证 Go 安装
go version

# 安装 Node.js 18+
# macOS: brew install node@18
# Ubuntu: apt-get install nodejs npm

# 验证 Node.js 安装
node --version
npm --version
```

#### 2. 准备 MongoDB

```bash
# 使用 Docker 启动 MongoDB
docker run -d -p 27017:27017 --name mongodb mongo:6.0

# 或安装本地 MongoDB 服务
# macOS: brew install mongodb-community
# Ubuntu: apt-get install mongodb
```

#### 3. 编译运行后端

```bash
# 克隆项目
git clone https://github.com/gomockserver/mockserver.git
cd mockserver

# 安装 Go 依赖
go mod download

# 编译
go build -o mockserver ./cmd/mockserver

# 运行（使用开发配置）
./mockserver -config config.dev.yaml
```

#### 4. 启动前端（新终端）

```bash
cd web/frontend

# 安装依赖（首次）
npm install

# 启动开发服务器
npm run dev
```

前端服务将运行在 http://localhost:5173

#### 5. 后台运行

```bash
# 使用 nohup 后台运行
nohup ./mockserver -config config.yaml > logs/mockserver.log 2>&1 &

# 查看进程
ps aux | grep mockserver

# 停止服务
pkill mockserver
```

### 前端独立部署

如果后端已经部署，只需要部署前端界面。

#### 1. 构建前端

```bash
cd web/frontend

# 安装依赖
npm install

# 构建生产版本
npm run build
```

构庻产物将输出到 `web/dist` 目录。

#### 2. 部署到静态服务器

**使用 Nginx**：

```nginx
server {
    listen 80;
    server_name mockserver.example.com;

    # 前端静态文件
    root /path/to/web/dist;
    index index.html;

    # SPA 路由支持
    location / {
        try_files $uri $uri/ /index.html;
    }

    # API 代理
    location /api {
        proxy_pass http://backend-server:8080;
        proxy_http_version 1.1;
        proxy_set_header Upgrade $http_upgrade;
        proxy_set_header Connection 'upgrade';
        proxy_set_header Host $host;
        proxy_cache_bypass $http_upgrade;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
    }
}
```

**使用 Apache**：

```apache
<VirtualHost *:80>
    ServerName mockserver.example.com
    DocumentRoot /path/to/web/dist

    <Directory /path/to/web/dist>
        Options -Indexes +FollowSymLinks
        AllowOverride All
        Require all granted

        # SPA 路由支持
        RewriteEngine On
        RewriteBase /
        RewriteRule ^index\.html$ - [L]
        RewriteCond %{REQUEST_FILENAME} !-f
        RewriteCond %{REQUEST_FILENAME} !-d
        RewriteRule . /index.html [L]
    </Directory>

    # API 代理
    ProxyPass /api http://backend-server:8080/api
    ProxyPassReverse /api http://backend-server:8080/api
</VirtualHost>
```

#### 3. 使用 CDN 加速

将构庻后的静态文件上传到 CDN（如 Cloudflare、AWS S3 + CloudFront），提高访问速度。

#### 4. 环境变量配置

在前端项目根目录创建 `.env.production`：

```bash
# API 基础地址
VITE_API_BASE_URL=https://api.mockserver.example.com/api/v1
```

### Kubernetes 部署

适合生产环境和需要高可用的场景。

#### 1. 准备配置文件

创建 `k8s/configmap.yaml`:

```yaml
apiVersion: v1
kind: ConfigMap
metadata:
  name: mockserver-config
data:
  config.yaml: |
    server:
      admin:
        host: "0.0.0.0"
        port: 8080
      mock:
        host: "0.0.0.0"
        port: 9090
    database:
      mongodb:
        uri: "mongodb://mongodb-service:27017"
        database: "mockserver"
        timeout: 10s
        pool:
          min: 10
          max: 100
    logging:
      level: "info"
      format: "json"
      output: "stdout"
```

#### 2. 部署 MongoDB

创建 `k8s/mongodb.yaml`:

```yaml
apiVersion: apps/v1
kind: StatefulSet
metadata:
  name: mongodb
spec:
  serviceName: mongodb-service
  replicas: 1
  selector:
    matchLabels:
      app: mongodb
  template:
    metadata:
      labels:
        app: mongodb
    spec:
      containers:
      - name: mongodb
        image: mongo:6.0
        ports:
        - containerPort: 27017
        volumeMounts:
        - name: mongodb-data
          mountPath: /data/db
  volumeClaimTemplates:
  - metadata:
      name: mongodb-data
    spec:
      accessModes: ["ReadWriteOnce"]
      resources:
        requests:
          storage: 20Gi
---
apiVersion: v1
kind: Service
metadata:
  name: mongodb-service
spec:
  selector:
    app: mongodb
  ports:
  - port: 27017
    targetPort: 27017
```

#### 3. 部署 Mock Server

创建 `k8s/mockserver.yaml`:

```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: mockserver
spec:
  replicas: 3
  selector:
    matchLabels:
      app: mockserver
  template:
    metadata:
      labels:
        app: mockserver
    spec:
      containers:
      - name: mockserver
        image: mockserver:latest
        ports:
        - containerPort: 8080
          name: admin
        - containerPort: 9090
          name: mock
        volumeMounts:
        - name: config
          mountPath: /root/config.yaml
          subPath: config.yaml
        livenessProbe:
          httpGet:
            path: /api/v1/system/health
            port: 8080
          initialDelaySeconds: 30
          periodSeconds: 10
        readinessProbe:
          httpGet:
            path: /api/v1/system/health
            port: 8080
          initialDelaySeconds: 10
          periodSeconds: 5
      volumes:
      - name: config
        configMap:
          name: mockserver-config
---
apiVersion: v1
kind: Service
metadata:
  name: mockserver-admin
spec:
  type: LoadBalancer
  selector:
    app: mockserver
  ports:
  - port: 8080
    targetPort: 8080
    name: admin
---
apiVersion: v1
kind: Service
metadata:
  name: mockserver-mock
spec:
  type: LoadBalancer
  selector:
    app: mockserver
  ports:
  - port: 9090
    targetPort: 9090
    name: mock
```

#### 4. 部署到 Kubernetes

```bash
# 应用配置
kubectl apply -f k8s/configmap.yaml
kubectl apply -f k8s/mongodb.yaml
kubectl apply -f k8s/mockserver.yaml

# 查看部署状态
kubectl get pods
kubectl get services

# 查看日志
kubectl logs -f deployment/mockserver
```

## 配置说明

### 配置文件示例

完整的 `config.yaml` 配置项说明：

```yaml
# 服务器配置
server:
  # 管理 API 服务
  admin:
    host: "0.0.0.0"    # 监听地址
    port: 8080          # 监听端口
  # Mock 服务
  mock:
    host: "0.0.0.0"
    port: 9090

# 数据库配置
database:
  mongodb:
    uri: "mongodb://localhost:27017"  # MongoDB 连接字符串
    database: "mockserver"             # 数据库名称
    timeout: 10s                       # 连接超时
    pool:
      min: 10    # 最小连接数
      max: 100   # 最大连接数

# Redis 配置（可选）
redis:
  enabled: false                # 是否启用 Redis
  host: "localhost"
  port: 6379
  password: ""
  db: 0
  pool:
    min: 5
    max: 50

# 安全配置
security:
  jwt:
    secret: "your-secret-key-change-in-production"  # JWT 密钥
    expiration: 7200                                 # 过期时间（秒）
  api_key:
    enabled: false
  ip_whitelist:
    enabled: false
    ips: []

# 日志配置
logging:
  level: "info"     # debug, info, warn, error
  format: "json"    # json, text
  output: "stdout"  # stdout, file
  file:
    path: "./logs/mockserver.log"
    max_size: 100     # MB
    max_backups: 10
    max_age: 30       # days

# 性能配置
performance:
  log_retention_days: 7  # 日志保留天数
  cache:
    rule_ttl: 300      # 规则缓存时间（秒）
    config_ttl: 1800   # 配置缓存时间（秒）
  rate_limit:
    enabled: true
    ip_limit: 1000       # 每分钟每IP请求数
    global_limit: 10000  # 每秒全局请求数

# 功能开关
features:
  version_control: true  # 规则版本控制
  audit_log: true        # 审计日志
  metrics: true          # 监控指标
```

### 环境变量

支持通过环境变量覆盖配置：

```bash
# 设置管理端口
export SERVER_ADMIN_PORT=8080

# 设置 Mock 端口
export SERVER_MOCK_PORT=9090

# 设置 MongoDB URI
export DATABASE_MONGODB_URI="mongodb://localhost:27017"

# 设置日志级别
export LOGGING_LEVEL="debug"
```

## 运维管理

### 日志管理

#### 查看日志

```bash
# Docker Compose
docker-compose logs -f mockserver

# Docker
docker logs -f mockserver-app

# 本地部署
tail -f logs/mockserver.log
```

#### 日志级别调整

修改 `config.yaml` 中的 `logging.level`：
- `debug`: 调试信息
- `info`: 常规信息
- `warn`: 警告信息
- `error`: 错误信息

### 数据备份

#### MongoDB 备份

```bash
# 导出数据
docker exec mockserver-mongodb mongodump \
  --db mockserver \
  --out /tmp/backup

# 复制备份文件
docker cp mockserver-mongodb:/tmp/backup ./backup

# 恢复数据
docker exec mockserver-mongodb mongorestore \
  --db mockserver \
  /tmp/backup/mockserver
```

### 性能监控

#### 监控指标

访问管理 API 查看系统状态：

```bash
# 健康检查
curl http://localhost:8080/api/v1/system/health

# 版本信息
curl http://localhost:8080/api/v1/system/version
```

## 故障排查

### 常见问题

#### 1. 服务无法启动

**症状**: 服务启动失败或立即退出

**排查步骤**:
```bash
# 查看日志
docker-compose logs mockserver

# 检查端口占用
lsof -i :8080
lsof -i :9090

# 检查配置文件
cat config.yaml
```

#### 2. 无法连接 MongoDB

**症状**: 日志显示数据库连接错误

**解决方案**:
```bash
# 检查 MongoDB 服务状态
docker-compose ps mongodb

# 检查网络连接
docker network ls
docker network inspect mockserver_mockserver-network

# 测试 MongoDB 连接
docker exec -it mockserver-mongodb mongosh
```

#### 3. Mock 规则不生效

**排查步骤**:
1. 检查规则是否启用 (`enabled: true`)
2. 确认 `project_id` 和 `environment_id` 正确
3. 检查规则优先级和匹配条件
4. 查看请求日志

#### 4. 性能问题

**优化建议**:
- 启用 Redis 缓存
- 增加 MongoDB 连接池大小
- 调整日志级别为 `warn` 或 `error`
- 使用 SSD 存储
- 增加服务器资源

### 获取帮助

如果遇到问题，可以：

1. 查看 [GitHub Issues](https://github.com/gomockserver/mockserver/issues)
2. 提交新的 Issue
3. 查看详细日志信息

## 安全建议

### 生产环境配置

1. **更改默认密钥**
   - 修改 `security.jwt.secret` 为强密钥

2. **启用 IP 白名单**
   ```yaml
   security:
     ip_whitelist:
       enabled: true
       ips:
         - "192.168.1.0/24"
         - "10.0.0.0/8"
   ```

3. **使用 HTTPS**
   - 在 Nginx 或负载均衡器配置 SSL 证书

4. **限制日志输出**
   - 设置 `logging.level: "warn"` 或 `"error"`

5. **定期备份**
   - 设置自动备份策略
   - 测试恢复流程

## 性能调优

### 推荐配置（高性能场景）

```yaml
database:
  mongodb:
    pool:
      min: 50
      max: 500

performance:
  cache:
    rule_ttl: 600
  rate_limit:
    global_limit: 50000

redis:
  enabled: true
```

### 监控指标

- QPS: > 10,000
- P99 延迟: < 50ms
- 内存使用: < 2GB
- CPU 使用: < 60%
