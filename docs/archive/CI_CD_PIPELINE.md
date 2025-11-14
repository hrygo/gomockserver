# CI/CD 测试流水线文档

**版本**: 1.0.0  
**创建时间**: 2025-11-14  
**维护者**: AI Agent

## 📋 概述

本项目使用 GitHub Actions 实现完整的 CI/CD 流水线，自动化执行测试、代码质量检查、构建和部署流程。

## 🎯 流水线架构

```
┌─────────────────────────────────────────────────┐
│           代码提交/PR                             │
└─────────────────┬───────────────────────────────┘
                  │
        ┌─────────┴─────────┐
        │                   │
        ▼                   ▼
  ┌──────────┐        ┌──────────┐
  │  CI Jobs │        │ PR Checks │
  └─────┬────┘        └─────┬────┘
        │                   │
        ├───┬───┬───┬───┐  │
        │   │   │   │   │  │
        ▼   ▼   ▼   ▼   ▼  ▼
       单  集  代  构  Docker PR
       元  成  码  建  测试  验证
       测  测  质          
       试  试  量          
```

## 📁 配置文件

### GitHub Actions 工作流

```
.github/workflows/
├── ci.yml          # 主CI流水线
├── docker.yml      # Docker构建和测试
└── pr-checks.yml   # PR检查
```

### 配置文件

```
.golangci.yml       # golangci-lint配置
```

## 🔧 CI 流水线 (ci.yml)

### 触发条件

- **Push事件**: main, develop分支
- **Pull Request**: main, develop分支

### 任务列表

#### 1. 单元测试 (unit-tests)

**运行环境**: Ubuntu Latest  
**Go版本**: 1.21

**步骤**:
1. Checkout代码
2. 设置Go环境
3. 下载依赖
4. 运行测试（带竞态检测和覆盖率）
5. 生成覆盖率报告
6. 上传到Codecov
7. 检查覆盖率阈值(50%)
8. 归档覆盖率结果

**关键命令**:
```bash
go test -v -race -coverprofile=coverage.out -covermode=atomic ./...
```

#### 2. 集成测试 (integration-tests)

**运行环境**: Ubuntu Latest + MongoDB Service

**服务依赖**:
- MongoDB 6.0（作为service容器）

**步骤**:
1. Checkout代码
2. 设置Go环境
3. 构建应用
4. 启动Mock Server
5. 等待服务就绪
6. 运行集成测试脚本
7. 停止服务
8. 归档日志（失败时）

**健康检查**:
```yaml
services:
  mongodb:
    options: >-
      --health-cmd "mongosh --eval 'db.adminCommand(\"ping\")'"
      --health-interval 10s
```

#### 3. 代码质量 (code-quality)

**检查项**:
- golangci-lint
- go vet
- gofmt格式化检查
- gosec安全扫描

**关键步骤**:
```yaml
- name: Run golangci-lint
  uses: golangci/golangci-lint-action@v3
  with:
    version: latest
    args: --timeout=5m
```

#### 4. 构建检查 (build)

**多平台构建**:
- Ubuntu
- macOS  
- Windows

**架构**: amd64

**验证**:
- 构建成功
- 二进制文件可执行

## 🐳 Docker 流水线 (docker.yml)

### 触发条件

- **Push事件**: main, develop分支 + version tags
- **Pull Request**: main分支

### 任务列表

#### 1. Docker构建 (docker-build)

**功能**:
- 构建生产镜像
- 推送到GitHub Container Registry
- 构建测试镜像
- 使用构建缓存加速

**镜像标签**:
- 分支名
- PR编号
- 语义化版本号(v1.0.0)

**示例**:
```yaml
tags: |
  type=ref,event=branch
  type=ref,event=pr
  type=semver,pattern={{version}}
  type=semver,pattern={{major}}.{{minor}}
```

#### 2. Docker Compose测试 (docker-compose-test)

**测试内容**:
- 启动测试环境
- 健康检查
- 运行集成测试
- 收集日志
- 清理环境

**关键步骤**:
```bash
docker-compose -f docker-compose.test.yml up -d mongodb-test mockserver-test
curl -f http://localhost:8081/api/v1/system/health
docker-compose -f docker-compose.test.yml run --rm test-runner
```

## ✅ PR检查流水线 (pr-checks.yml)

### 触发条件

Pull Request的opened, synchronize, reopened事件

### 检查项目

#### 1. PR验证 (pr-checks)

**检查内容**:
- PR标题格式（语义化提交）
- 大文件检查（>1MB）
- 敏感数据扫描（Trufflehog）

#### 2. 覆盖率报告 (coverage-check)

**功能**:
- 运行测试并生成覆盖率
- 在PR中评论覆盖率报告
- 检查覆盖率阈值

**PR评论示例**:
```markdown
## Test Coverage Report

**Total Coverage:** 56.1%

<details>
<summary>Detailed Coverage</summary>

```
adapter:    96.3%
engine:     89.8%
executor:   86.0%
...
```
</details>
```

#### 3. 变更文件分析 (changed-files)

**功能**:
- 检测变更的Go文件
- 提示是否需要测试

#### 4. 依赖审查 (dependency-review)

**功能**:
- 检查依赖安全性
- 中等及以上严重性时失败

## 🔒 代码质量配置

### golangci-lint配置

**启用的linters**:
- errcheck - 错误处理检查
- gosimple - 简化建议
- govet - Go vet
- staticcheck - 静态分析
- gofmt - 格式化
- goimports - import顺序
- misspell - 拼写检查
- gosec - 安全检查
- bodyclose - HTTP body关闭

**配置示例**:
```yaml
linters:
  enable:
    - errcheck
    - gosimple
    - govet
    - staticcheck
  
linters-settings:
  errcheck:
    check-type-assertions: true
    check-blank: true
```

## 📊 覆盖率报告

### Codecov集成

使用Codecov自动上传和跟踪覆盖率：

```yaml
- name: Upload coverage to Codecov
  uses: codecov/codecov-action@v3
  with:
    files: ./coverage.out
    flags: unittests
```

### 覆盖率阈值

**最低要求**: 50%

低于阈值时会产生警告但不会失败（可配置为失败）。

## 🚀 使用指南

### 本地运行CI检查

#### 1. 运行单元测试

```bash
go test -v -race -coverprofile=coverage.out ./...
go tool cover -func=coverage.out
```

#### 2. 运行代码质量检查

```bash
# 安装golangci-lint
go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest

# 运行检查
golangci-lint run --timeout=5m
```

#### 3. 运行集成测试

```bash
./tests/integration/e2e_test.sh
```

#### 4. Docker测试

```bash
docker-compose -f docker-compose.test.yml up -d
docker-compose -f docker-compose.test.yml run --rm test-runner
docker-compose -f docker-compose.test.yml down -v
```

### CI Badge

在README中添加状态徽章：

```markdown
![CI Tests](https://github.com/username/repo/workflows/CI%20Tests/badge.svg)
![Docker Build](https://github.com/username/repo/workflows/Docker%20Build%20and%20Test/badge.svg)
[![codecov](https://codecov.io/gh/username/repo/branch/main/graph/badge.svg)](https://codecov.io/gh/username/repo)
```

## 🔧 故障排查

### 问题1: 单元测试失败

**检查**:
```bash
# 本地运行测试
go test -v ./...

# 查看CI日志
# GitHub Actions -> 失败的workflow -> 查看详细日志
```

### 问题2: 集成测试超时

**原因**: MongoDB启动慢或服务未就绪

**解决**:
```yaml
# 增加等待时间
- name: Wait for server ready
  run: |
    for i in {1..60}; do  # 从30增加到60
      ...
    done
```

### 问题3: golangci-lint失败

**查看具体错误**:
```bash
# 本地运行
golangci-lint run

# 修复格式问题
gofmt -w .
goimports -w .
```

### 问题4: Docker构建失败

**检查**:
```bash
# 本地构建测试
docker build -f Dockerfile.test -t mockserver:test .

# 查看构建日志
# GitHub Actions -> docker-build job -> 查看步骤输出
```

## 📈 性能优化

### 1. 缓存策略

**Go模块缓存**:
```yaml
- uses: actions/setup-go@v4
  with:
    cache: true  # 自动缓存go mod
```

**Docker层缓存**:
```yaml
- uses: docker/build-push-action@v5
  with:
    cache-from: type=gha
    cache-to: type=gha,mode=max
```

### 2. 并行执行

多个job并行运行以提高速度：
- unit-tests
- integration-tests
- code-quality
- build

### 3. 条件执行

```yaml
# 只在PR时运行
if: github.event_name == 'pull_request'

# 只在push时推送镜像
if: github.event_name != 'pull_request'
```

## 🔐 安全最佳实践

### 1. Secret管理

使用GitHub Secrets存储敏感信息：
```yaml
env:
  GITHUB_TOKEN: ${{ secrets.GITHUB_TOKEN }}
```

### 2. 权限控制

最小权限原则：
```yaml
permissions:
  contents: read
  packages: write
```

### 3. 依赖安全

定期审查依赖：
```yaml
- uses: actions/dependency-review-action@v3
  with:
    fail-on-severity: moderate
```

## 📝 自定义配置

### 添加新的检查

在 `.github/workflows/ci.yml` 中添加新job：

```yaml
  custom-check:
    name: Custom Check
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - name: Run custom script
        run: ./scripts/custom-check.sh
```

### 修改覆盖率阈值

```yaml
- name: Check coverage threshold
  run: |
    COVERAGE_NUM=$(echo $COVERAGE | sed 's/%//')
    if (( $(echo "$COVERAGE_NUM < 60.0" | bc -l) )); then  # 从50改为60
      exit 1  # 改为失败而不是警告
    fi
```

### 添加通知

使用Slack或邮件通知：
```yaml
- name: Notify Slack
  if: failure()
  uses: 8398a7/action-slack@v3
  with:
    status: ${{ job.status }}
    webhook_url: ${{ secrets.SLACK_WEBHOOK }}
```

## 📊 监控和报告

### 1. 工作流运行历史

在GitHub仓库中查看：
```
Actions -> All workflows -> 选择workflow -> 查看运行历史
```

### 2. 覆盖率趋势

Codecov dashboard提供：
- 覆盖率趋势图
- 文件级别覆盖率
- PR覆盖率差异

### 3. 构建时间分析

查看每个job的执行时间：
```
Actions -> 选择运行 -> 查看时间线
```

## 🎓 最佳实践

### 1. 快速失败

将快速检查放在前面：
1. 代码格式检查（几秒）
2. 单元测试（几分钟）
3. 集成测试（10-15分钟）

### 2. 清晰的日志

使用有意义的步骤名称：
```yaml
- name: Run unit tests with race detection
  run: go test -v -race ./...
```

### 3. 环境一致性

- 使用固定的Go版本
- 使用固定的Action版本(@v4而不是@latest)
- Docker镜像使用具体版本号

### 4. 及时清理

使用 `if: always()` 确保清理步骤总是执行：
```yaml
- name: Cleanup
  if: always()
  run: docker-compose down -v
```

## 🔄 持续改进

### 定期审查

- 每月审查CI配置
- 更新依赖版本
- 优化执行时间
- 检查失败模式

### 指标跟踪

监控：
- 平均构建时间
- 失败率
- 覆盖率趋势

---

**文档版本**: 1.0.0  
**最后更新**: 2025-11-14  
**下次审核**: 每季度或重大变更时
