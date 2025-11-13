# Mock Server 测试指南

本文档介绍如何对 Mock Server 进行测试。

## 快速开始

### 1. 运行静态检查测试

最快速的验证方式，不需要任何外部依赖：

```bash
make test-static
```

或直接运行脚本：

```bash
chmod +x mvp-test.sh
./mvp-test.sh
```

**检查内容**:
- Go 环境验证
- 代码编译检查
- 模块完整性验证
- 配置文件检查
- 文档完整性
- 项目结构验证

**预期结果**: 26/27 项通过

### 2. 快速验证

运行代码格式化、静态分析和编译检查：

```bash
make verify
```

这个命令会自动执行：
1. 代码格式化 (`gofmt`)
2. 静态分析 (`go vet`)
3. 项目编译
4. 静态测试

### 3. 代码格式化

```bash
make fmt
```

### 4. 编译项目

```bash
make build
```

编译后的二进制文件位于 `bin/mockserver`

## 测试类型

### 静态测试 ✅ 已实现

**运行命令**:
```bash
make test-static
```

**测试内容**:
- 环境检查（Go、Docker）
- 代码质量检查（格式、编译）
- 模块功能验证
- 配置文件验证
- 文档完整性检查
- 代码结构检查

**输出**: 
- 终端输出测试结果
- 生成 `test-report-*.md` 测试报告

### 单元测试 ⏸️ 待实现

**计划运行命令**:
```bash
make test-unit
```

**测试范围**:
- 规则匹配引擎 (`internal/engine`)
- Mock 执行器 (`internal/executor`)
- HTTP 适配器 (`internal/adapter`)
- Repository 层 (`internal/repository`)

**目标覆盖率**: > 80%

**实现步骤**:
1. 为每个模块创建 `*_test.go` 文件
2. 使用 `testify/assert` 进行断言
3. 使用 `testify/mock` 进行接口 Mock
4. 运行 `go test -v ./internal/...`

### 集成测试 ⏸️ 部分实现

**运行命令**:
```bash
# 启动 MongoDB
make docker-up

# 运行集成测试
make test-integration
```

**现有测试脚本**: `test.sh`

**测试流程**:
1. 创建测试项目
2. 创建测试环境
3. 创建多个 Mock 规则
4. 发送 HTTP 请求验证
5. 测试规则启用/禁用
6. 验证环境隔离

**前置条件**:
- MongoDB 服务运行中
- 管理 API 服务（8080端口）
- Mock 服务（9090端口）

### 性能测试 ⏸️ 待实现

**目标指标**:
- QPS > 10,000
- 平均响应时间 < 10ms
- P99 响应时间 < 50ms
- 支持并发连接 > 5,000

**推荐工具**:
- Apache JMeter
- wrk
- Go benchmark

**测试场景**:
1. 基准性能测试（简单规则）
2. 大规模规则匹配（1000+ 规则）
3. 并发压力测试（逐步增加并发）
4. 响应延迟性能测试

### 覆盖率测试 ⏸️ 待实现

**运行命令**:
```bash
make test-coverage
```

**输出**:
- `coverage.out`: 覆盖率数据文件
- `coverage.html`: HTML 格式覆盖率报告

**查看报告**:
```bash
open coverage.html
```

## 测试环境准备

### 本地开发环境

1. **安装 Go 1.21+**
```bash
go version
```

2. **安装依赖**
```bash
make deps
```

3. **验证依赖**
```bash
make deps-check
```

### Docker 环境

1. **启动所有服务**
```bash
make docker-up
```

这会启动：
- MongoDB (27017端口)
- Mock Server 管理 API (8080端口)
- Mock Server Mock 服务 (9090端口)

2. **查看服务状态**
```bash
docker-compose ps
```

3. **查看日志**
```bash
make docker-logs
```

4. **停止服务**
```bash
make docker-down
```

### 测试数据库

**使用 Docker MongoDB**:
```bash
docker run -d -p 27017:27017 --name mongodb-test mongo:6.0
```

**使用 testcontainers-go** (推荐用于单元测试):
```go
import "github.com/testcontainers/testcontainers-go"
```

## 测试工具

### 已集成工具

| 工具 | 用途 | 命令 |
|------|------|------|
| gofmt | 代码格式化 | `make fmt` |
| go vet | 静态分析 | `make vet` |
| go build | 编译检查 | `make build` |
| Makefile | 测试自动化 | `make help` |

### 推荐安装工具

#### 1. golangci-lint (代码检查)

```bash
# macOS
brew install golangci-lint

# 运行检查
make lint
```

#### 2. testify (测试框架)

```bash
go get github.com/stretchr/testify
```

#### 3. testcontainers-go (容器化测试)

```bash
go get github.com/testcontainers/testcontainers-go
```

## 测试数据

### 标准测试数据

**项目配置**:
```json
{
  "name": "测试项目",
  "workspace_id": "test-workspace",
  "description": "用于测试的项目"
}
```

**环境配置**:
```json
{
  "name": "开发环境",
  "project_id": "<project_id>",
  "base_url": "http://localhost:9090"
}
```

**规则配置**:
```json
{
  "name": "获取用户列表",
  "project_id": "<project_id>",
  "environment_id": "<env_id>",
  "protocol": "HTTP",
  "match_type": "Simple",
  "priority": 100,
  "enabled": true,
  "match_condition": {
    "method": "GET",
    "path": "/api/users"
  },
  "response": {
    "type": "Static",
    "content": {
      "status_code": 200,
      "content_type": "JSON",
      "body": {
        "code": 0,
        "message": "success",
        "data": []
      }
    }
  }
}
```

## 常见问题

### Q1: Docker 启动失败

**问题**: `docker-compose up -d` 超时

**解决方案**:
1. 配置 Docker 镜像加速
2. 检查网络连接
3. 尝试使用本地 MongoDB

```bash
docker run -d -p 27017:27017 --name mongodb mongo:6.0
```

### Q2: 测试编译失败

**问题**: 找不到依赖包

**解决方案**:
```bash
go mod tidy
go mod download
```

### Q3: 代码格式检查失败

**解决方案**:
```bash
make fmt
```

### Q4: 端口被占用

**问题**: 8080 或 9090 端口已被使用

**解决方案**:
1. 修改 `config.yaml` 中的端口配置
2. 或停止占用端口的服务

```bash
lsof -i :8080
lsof -i :9090
```

## 测试报告

### 自动生成报告

运行测试后会自动生成以下报告：

1. **测试执行报告**: `test-report-<timestamp>.md`
   - 测试统计信息
   - 详细测试结果
   - 功能验证清单

2. **测试执行总结**: `TEST_EXECUTION_SUMMARY.md`
   - 测试执行情况
   - 功能验证状态
   - 代码质量评估
   - 后续行动计划

### 查看报告

```bash
# 查看最新测试报告
ls -lt test-report-*.md | head -1 | xargs cat

# 查看测试执行总结
cat TEST_EXECUTION_SUMMARY.md
```

## 持续集成

### GitHub Actions 配置示例

```yaml
name: Test

on: [push, pull_request]

jobs:
  test:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v2
      - uses: actions/setup-go@v2
        with:
          go-version: 1.21
      - name: Run static tests
        run: make test-static
      - name: Run unit tests
        run: make test-unit
      - name: Generate coverage
        run: make test-coverage
```

## 贡献测试

### 编写测试用例

1. 在对应模块目录创建 `*_test.go` 文件
2. 使用 table-driven tests 模式
3. 包含正常、边界、异常场景
4. 添加清晰的测试描述

**示例**:
```go
func TestRuleMatching(t *testing.T) {
    tests := []struct {
        name        string
        rule        *models.Rule
        request     *adapter.Request
        shouldMatch bool
    }{
        {
            name: "精确路径匹配",
            rule: createTestRule("/api/users", "GET"),
            request: createTestRequest("/api/users", "GET"),
            shouldMatch: true,
        },
        // 更多测试用例...
    }
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            // 测试逻辑
        })
    }
}
```

### 测试检查清单

提交测试代码前确认：
- [ ] 测试命名清晰（Test开头）
- [ ] 包含测试描述
- [ ] 覆盖正常场景
- [ ] 覆盖边界场景
- [ ] 覆盖异常场景
- [ ] 测试独立运行
- [ ] 测试可重复执行
- [ ] 添加必要注释

## 参考资料

- [Go 测试官方文档](https://golang.org/pkg/testing/)
- [Testify 文档](https://github.com/stretchr/testify)
- [测试方案设计](.qoder/quests/perfect-mvp-testing-plan.md)
- [测试执行总结](TEST_EXECUTION_SUMMARY.md)

---

**最后更新**: 2025-11-13  
**维护人**: Mock Server 团队
