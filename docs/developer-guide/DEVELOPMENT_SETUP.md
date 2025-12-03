# MockServer 开发环境搭建指南

> 👨‍💻 **面向开发者**
> ⏱️ **阅读时间**: 20分钟
> 🎯 **目标**: 搭建完整的开发环境并参与贡献

---

## 📋 目录

1. [开发环境要求](#开发环境要求)
2. [环境搭建步骤](#环境搭建步骤)
3. [项目结构说明](#项目结构说明)
4. [开发工作流](#开发工作流)
5. [调试技巧](#调试技巧)
6. [常见问题](#常见问题)
7. [贡献指南](#贡献指南)

---

## 开发环境要求

### 系统要求
- **操作系统**: Linux/macOS/Windows 10+
- **内存**: 8GB+ RAM (推荐16GB)
- **磁盘**: 10GB+ 可用空间
- **网络**: 稳定的互联网连接

### 必需软件

#### 核心工具
- **Go 1.24+**
  ```bash
  # 验证安装
  go version

  # 设置代理（国内）
  go env -w GOPROXY=https://goproxy.cn,direct
  ```

- **Node.js 18+ & npm 8+**
  ```bash
  # 推荐使用 nvm
  curl -o- https://raw.githubusercontent.com/nvm-sh/nvm/v0.39.0/install.sh | bash
  nvm install 18
  nvm use 18
  ```

- **Docker 20.10+ & Docker Compose 2.0+**
  ```bash
  docker --version
  docker-compose --version
  ```

#### 数据库
- **MongoDB 6.0+** (本地开发)或使用Docker容器
- **Redis 6.0+** (缓存功能，可选)

#### 开发工具（推荐）
- **IDE**: VS Code / GoLand / Vim
- **API测试**: Postman / Insomnia
- **Git客户端**: SourceTree / GitKraken / 命令行

---

## 环境搭建步骤

### 1. 克隆项目

```bash
# 1. Fork 项目到你的GitHub账号
# 2. 克隆你的fork
git clone https://github.com/YOUR_USERNAME/mockserver.git
cd mockserver

# 3. 添加上游仓库
git remote add upstream https://github.com/gomockserver/mockserver.git

# 4. 验证
git remote -v
```

### 2. 配置开发环境

```bash
# 1. 安装Go依赖
go mod download
go mod verify

# 2. 安装开发工具
make install-tools

# 3. 安装前端依赖
cd web/frontend
npm install
cd ../..

# 4. 创建开发配置
cp config/config.example.yaml config/config.dev.yaml
```

### 3. 配置IDE

#### VS Code配置

创建 `.vscode/settings.json`:
```json
{
  "go.useLanguageServer": true,
  "go.lintTool": "golangci-lint",
  "go.lintOnSave": "file",
  "go.formatTool": "goimports",
  "go.testFlags": ["-v"],
  "go.coverOnSave": true,
  "go.coverageDecorator": {
    "type": "gutter",
    "coveredHighlightColor": "rgba(64,128,64,0.5)",
    "uncoveredHighlightColor": "rgba(128,64,64,0.25)"
  },
  "typescript.preferences.importModuleSpecifier": "relative",
  "editor.formatOnSave": true,
  "editor.codeActionsOnSave": {
    "source.organizeImports": true
  }
}
```

创建 `.vscode/launch.json`:
```json
{
  "version": "0.2.0",
  "configurations": [
    {
      "name": "Launch MockServer",
      "type": "go",
      "request": "launch",
      "mode": "auto",
      "program": "${workspaceFolder}/cmd/mockserver",
      "env": {
        "CONFIG_FILE": "${workspaceFolder}/config/config.dev.yaml",
        "GIN_MODE": "debug"
      },
      "args": []
    }
  ]
}
```

#### GoLand配置
1. 打开项目
2. File → Settings → Go → GOPATH → 设置项目GOPATH
3. Settings → Editor → Code Style → Go → 导入项目配置
4. Settings → Editor → Inspections → Go → 启用所有检查

### 4. 启动开发服务

#### 方式一：使用Make命令（推荐）

```bash
# 1. 启动所有依赖服务
make start-services

# 2. 启动后端（开发模式）
make dev-backend

# 3. 启动前端（新终端）
make dev-frontend

# 4. 运行测试
make test
```

#### 方式二：手动启动

```bash
# 1. 启动MongoDB和Redis
docker-compose up -d mongo redis

# 2. 启动后端
cd cmd/mockserver
go run main.go --config=../../config/config.dev.yaml

# 3. 启动前端（新终端）
cd web/frontend
npm run dev
```

---

## 项目结构说明

```
gomockserver/
├── cmd/                      # 应用入口
│   └── mockserver/          # 主程序
├── internal/                 # 内部包（不导出）
│   ├── adapter/             # 协议适配器
│   │   ├── http_adapter.go
│   │   └── websocket_adapter.go
│   ├── api/                 # API处理器
│   │   ├── handlers/
│   │   └── middleware/
│   ├── cache/               # 缓存系统
│   │   ├── l1_cache.go
│   │   ├── l2_cache.go
│   │   └── manager.go
│   ├── config/              # 配置管理
│   ├── engine/              # 匹配引擎
│   ├── graphql/             # GraphQL实现
│   ├── models/              # 数据模型
│   ├── repository/          # 数据访问层
│   └── service/             # 业务逻辑层
├── pkg/                     # 公共包（可导出）
│   └── logger/              # 日志工具
├── web/                     # 前端代码
│   └── frontend/            # React应用
│       ├── src/
│       │   ├── components/
│       │   ├── pages/
│       │   ├── hooks/
│       │   ├── services/
│       │   └── utils/
│       ├── public/
│       └── package.json
├── tests/                   # 测试代码
│   ├── integration/         # 集成测试
│   ├── unit/                # 单元测试
│   └── fixtures/            # 测试数据
├── scripts/                 # 脚本工具
├── config/                  # 配置文件
├── docs/                    # 文档
├── docker/                  # Docker相关
└── Makefile                 # 构建脚本
```

---

## 开发工作流

### 1. 创建功能分支

```bash
# 1. 同步主分支
git checkout master
git pull upstream master

# 2. 创建功能分支
git checkout -b feature/your-feature-name

# 3. 设置分支追踪
git push -u origin feature/your-feature-name
```

### 2. 开发阶段

```bash
# 1. 编码
# ... 编写代码 ...

# 2. 运行测试
make test

# 3. 代码检查
make lint

# 4. 格式化代码
make fmt

# 5. 提交前检查
make pre-commit
```

### 3. 提交规范

使用 [Conventional Commits](https://www.conventionalcommits.org/) 规范：

```bash
# 功能添加
git commit -m "feat(api): add support for GraphQL mutations"

# Bug修复
git commit -m "fix(cache): resolve Redis connection timeout issue"

# 文档更新
git commit -m "docs: update API documentation for v0.8.1"

# 性能优化
git commit -m "perf(engine): improve regex matching speed by 50%"

# 测试
git commit -m "test: add integration tests for WebSocket"
```

### 4. 创建Pull Request

```bash
# 1. 推送到你的fork
git push origin feature/your-feature-name

# 2. 在GitHub上创建PR
# 3. 填写PR模板
# 4. 等待代码审查
```

PR模板：
```markdown
## 变更类型
- [ ] Bug修复
- [ ] 新功能
- [ ] 破坏性变更
- [ ] 文档更新

## 变更描述
简要描述本次变更的内容和原因。

## 测试
- [ ] 单元测试通过
- [ ] 集成测试通过
- [ ] 手动测试完成

## 检查清单
- [ ] 代码遵循项目规范
- [ ] 添加了必要的测试
- [ ] 更新了相关文档
- [ ] 没有引入新的警告
```

---

## 调试技巧

### 1. 启用调试模式

```bash
# 设置环境变量
export GIN_MODE=debug
export LOG_LEVEL=debug

# 或修改配置文件
vim config/config.dev.yaml
```

### 2. 使用Delve调试器

```bash
# 1. 安装delve
go install github.com/go-delve/delve/cmd/dlv@latest

# 2. 调试运行
dlv debug cmd/mockserver/main.go

# 3. 常用命令
(Delve) break main.main         # 设置断点
(Delve) continue               # 继续
(Delve) next                   # 下一步
(Delve) print variable         # 打印变量
(Delve) locals                 # 查看局部变量
```

### 3. 热重载

使用 [Air](https://github.com/cosmtrek/air) 实现热重载：

```bash
# 1. 安装air
go install github.com/cosmtrek/air@latest

# 2. 创建配置
cat > .air.toml << EOF
root = "."
testdata_dir = "testdata"
tmp_dir = "tmp"

[build]
  args_bin = []
  bin = "./tmp/main"
  cmd = "go build -o ./tmp/main ./cmd/mockserver"
  delay = 1000
  exclude_dir = ["assets", "tmp", "vendor", "testdata", "web"]
  exclude_file = []
  exclude_regex = ["_test.go"]
  exclude_unchanged = false
  follow_symlink = false
  full_bin = ""
  include_dir = []
  include_ext = ["go", "tpl", "tmpl", "html"]
  kill_delay = "0s"
  log = "build-errors.log"
  send_interrupt = false
  stop_on_root = false

[color]
  app = ""
  build = "yellow"
  main = "magenta"
  runner = "green"
  watcher = "cyan"

[log]
  time = false

[misc]
  clean_on_exit = false
EOF

# 3. 运行
air
```

### 4. 前端调试

```bash
# 1. 启用Source Map
# web/frontend/vite.config.ts
export default defineConfig({
  build: {
    sourcemap: true
  }
})

# 2. 使用React DevTools
npm install --save-dev react-devtools

# 3. Chrome DevTools
# 在浏览器中使用F12打开开发者工具
```

---

## 常见问题

### 1. Go版本问题

**问题**: `go: cannot find main module`

**解决**:
```bash
# 1. 检查go.mod文件
ls -la go.mod

# 2. 初始化模块（如果需要）
go mod init github.com/gomockserver/mockserver

# 3. 同步依赖
go mod tidy
```

### 2. 前端依赖问题

**问题**: npm安装失败

**解决**:
```bash
# 1. 清除缓存
npm cache clean --force

# 2. 删除node_modules
rm -rf node_modules package-lock.json

# 3. 重新安装
npm install

# 4. 使用国内镜像
npm config set registry https://registry.npmmirror.com
```

### 3. 测试失败

**问题**: 测试连接超时

**解决**:
```bash
# 1. 检查服务状态
make status

# 2. 重启服务
make restart

# 3. 跳过集成测试
go test ./... -tags=unit

# 4. 单独运行失败的测试
go test -v ./internal/cache -run TestCacheManager
```

### 4. Docker问题

**问题**: 容器启动失败

**解决**:
```bash
# 1. 查看详细日志
docker-compose logs -f

# 2. 重新构建
docker-compose build --no-cache

# 3. 清理资源
docker-compose down -v
docker system prune -a
```

---

## 贡献指南

### 1. 代码规范

#### Go代码规范
- 遵循 [Go Code Review Comments](https://github.com/golang/go/wiki/CodeReviewComments)
- 使用 `gofmt` 格式化代码
- 使用 `golangci-lint` 进行静态检查
- 包名使用小写单词
- 接口名以-er结尾

#### TypeScript代码规范
- 使用 TypeScript严格模式
- 遵循 ESLint规则
- 使用 Prettier格式化
- 组件使用 PascalCase
- 文件名使用 camelCase

### 2. 提交前检查

```bash
# 运行所有检查
make check

# 或单独运行
make fmt          # 格式化
make lint         # 代码检查
make test         # 运行测试
make coverage     # 测试覆盖率
```

### 3. 性能考虑

- 避免不必要的内存分配
- 使用 sync.Pool 复用对象
- 批量操作减少数据库访问
- 合理使用缓存
- 避免阻塞操作

### 4. 安全考虑

- 输入验证和清理
- SQL注入防护
- XSS防护
- 敏感信息加密
- 使用最小权限原则

---

## 开发工具推荐

### VS Code插件
```json
{
  "recommendations": [
    "golang.go",
    "bradlc.vscode-tailwindcss",
    "esbenp.prettier-vscode",
    "dbaeumer.vscode-eslint",
    "ms-vscode.vscode-typescript-next",
    "formulahendry.auto-rename-tag",
    "christian-kohler.path-intellisense",
    "ms-vscode.vscode-json"
  ]
}
```

### Chrome插件
- React Developer Tools
- Apollo Client Developer Tools
- JSON Viewer
- Postman Interceptor

---

## 获取帮助

- 📖 [项目文档](../../docs/README.md)
- 💬 [开发者讨论](https://github.com/gomockserver/mockserver/discussions)
- 🐛 [报告Bug](https://github.com/gomockserver/mockserver/issues)
- 📧 [邮件联系](mailto:dev@gomockserver.com)

---

<div align="center">

**🚀 感谢您的贡献！**

[返回文档首页](../../README.md) | [查看架构文档](../ARCHITECTURE.md)

</div>