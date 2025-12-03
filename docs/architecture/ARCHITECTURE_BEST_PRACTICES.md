# MockServer 工程最佳实践指南

> 📅 创建日期：2025-11-19
> 🎯 目标：防止工程目录腐化，保持代码质量和架构一致性
> 📋 检查频率：建议每月定期检查，重大功能开发后必须检查

---

## 📋 目录

1. [项目架构原则](#1-项目架构原则)
2. [目录结构规范](#2-目录结构规范)
3. [代码组织最佳实践](#3-代码组织最佳实践)
4. [配置管理规范](#4-配置管理规范)
5. [测试组织标准](#5-测试组织标准)
6. [文档管理体系](#6-文档管理体系)
7. [构建与部署规范](#7-构建与部署规范)
8. [脚本管理指南](#8-脚本管理指南)
9. [质量检查清单](#9-质量检查清单)
10. [腐化预防机制](#10-腐化预防机制)

---

## 1. 项目架构原则

### 🏗️ 核心设计原则

#### 分层架构
```
┌─────────────────┐
│   前端展示层     │ ← React + TypeScript + Ant Design
├─────────────────┤
│   API网关层      │ ← GraphQL + RESTful API
├─────────────────┤
│   业务逻辑层     │ ← Service + Domain Logic
├─────────────────┤
│   数据访问层     │ ← Repository + Cache
├─────────────────┤
│   基础设施层     │ ← Database + External Services
└─────────────────┘
```

#### 职责分离原则
- **单一职责**：每个模块只负责一个明确的功能域
- **依赖倒置**：高层模块不依赖低层模块，都依赖抽象
- **开闭原则**：对扩展开放，对修改封闭
- **接口隔离**：客户端不应该依赖它不需要的接口

### 🔧 技术栈标准

#### 后端技术栈
- **语言**: Go 1.24+
- **Web框架**: Gin v1.11+
- **数据库**: MongoDB 6.0+
- **缓存**: Redis 6.0+
- **API**: RESTful + GraphQL
- **容器化**: Docker + Docker Compose

#### 前端技术栈
- **框架**: React 18+
- **语言**: TypeScript 5+
- **构建工具**: Vite 5.1+
- **状态管理**: Apollo Client + Zustand
- **UI组件**: Ant Design 5.14+
- **图表**: ECharts 5.6+

---

## 2. 目录结构规范

### 📁 标准目录结构

```
gomockserver/
├── cmd/                          # 应用程序入口
│   └── mockserver/              # 主应用
├── internal/                     # 私有应用代码
│   ├── adapter/                # 协议适配器
│   ├── api/                     # API处理层
│   ├── cache/                   # 缓存模块
│   ├── config/                  # 配置管理
│   ├── engine/                  # 核心匹配引擎
│   ├── executor/                # Mock执行器
│   ├── graphql/                 # GraphQL处理
│   ├── middleware/              # HTTP中间件
│   ├── models/                  # 数据模型
│   ├── monitoring/              # 监控模块
│   ├── repository/              # 数据访问层
│   └── service/                 # 业务服务层
├── pkg/                         # 公共库代码
│   └── logger/                  # 日志库
├── web/                         # 前端代码
│   └── frontend/
│       ├── public/              # 静态资源
│       ├── src/
│       │   ├── api/             # API封装
│       │   ├── components/      # 可复用组件
│       │   ├── graphql/         # GraphQL相关
│       │   ├── hooks/           # 自定义Hooks
│       │   ├── lib/             # 第三方库配置
│       │   ├── pages/           # 页面组件
│       │   ├── providers/       # Context提供者
│       │   ├── store/           # 状态管理
│       │   ├── styles/          # 样式文件
│       │   ├── types/           # TypeScript类型
│       │   └── utils/           # 工具函数
│       ├── package.json
│       ├── tsconfig.json
│       └── vite.config.ts
├── tests/                       # 测试代码
│   ├── unit/                    # 单元测试
│   ├── integration/             # 集成测试
│   ├── e2e/                      # 端到端测试
│   ├── performance/             # 性能测试
│   ├── fixtures/                # 测试固件
│   ├── coverage/                # 覆盖率报告
│   └── scripts/                 # 测试脚本
├── docs/                        # 项目文档
│   ├── api/                     # API文档
│   ├── architecture/            # 架构文档
│   ├── deployment/              # 部署文档
│   ├── development/             # 开发文档
│   └── user-guide/              # 用户指南
├── scripts/                     # 运维脚本
│   ├── dev/                     # 开发脚本
│   ├── deploy/                  # 部署脚本
│   ├── maintenance/             # 维护脚本
│   └── quality/                 # 质量检查脚本
├── docker/                      # Docker相关文件
├── .github/                     # GitHub配置
│   └── workflows/               # CI/CD工作流
├── config.yaml                  # 生产环境配置
├── config.dev.yaml              # 开发环境配置
├── config.test.yaml             # 测试环境配置
├── docker-compose.yml           # Docker编排文件
├── docker-compose.test.yml      # 测试环境Docker
├── Dockerfile                   # 基础Docker镜像
├── Makefile                     # 构建脚本
├── go.mod                       # Go模块定义
├── go.sum                       # Go依赖锁定
├── .golangci.yml               # 代码质量配置
├── .gitignore                   # Git忽略文件
└── README.md                    # 项目说明
```

### 📏 目录命名规范

#### 标准命名约定
- **小写字母**: 所有目录名使用小写字母
- **连字符分隔**: 多词目录使用连字符分隔 (如 `user-guide`)
- **单数形式**: 目录名使用单数形式 (如 `service` 而不是 `services`)
- **语义明确**: 目录名应该清晰表达其用途

#### 禁止的目录名
- ❌ 大写字母目录名
- ❌ 驼峰命名目录名
- ❌ 缩写不明确的目录名
- ❌ 临时目录名 (如 `temp`, `tmp`)

### 📂 深度限制

#### 目录层次规范
- **最大深度**: 不超过4层目录嵌套
- **推荐深度**: 2-3层为最佳
- **扁平化优先**: 功能相关的文件应尽量扁平化组织

#### 当前需要优化的深度问题
```
# 问题示例 (深度=4，需要优化)
internal/graphql/resolvers/
internal/graphql/schema/
internal/graphql/handlers/

# 优化方案 (深度=3)
internal/graphql/
├── resolvers.go
├── schema.go
├── handlers.go
```

---

## 3. 代码组织最佳实践

### 🎯 Go代码组织

#### 包结构规范

**internal/ 包结构**:
```go
// ✅ 正确的包组织
internal/
├── adapter/           // 适配器层 - 处理外部协议
│   ├── http/         // HTTP适配器
│   └── websocket/    // WebSocket适配器
├── api/              // API层 - HTTP路由处理
│   ├── handlers/     // 请求处理器
│   └── middleware/   // 中间件
├── service/          // 业务服务层
│   ├── mock/         // Mock服务
│   └── project/      // 项目管理服务
└── repository/       // 数据访问层
    ├── mongodb/      // MongoDB实现
    └── redis/        // Redis实现
```

**包命名规范**:
- **简短描述性**: 包名应该简短且具有描述性
- **避免重复**: 避免包名与上级目录重复
- **小写字母**: 包名使用小写字母，不使用下划线或连字符

#### 文件组织规范

**单个文件原则**:
- 一个文件只包含一个主要类型或功能
- 相关的常量和辅助函数可以放在同一文件
- 文件名应该与主要内容匹配

**示例**:
```go
// ✅ 正确的文件组织
service/
├── mock_service.go     // MockService主类型
├── project_service.go  // ProjectService主类型
└── interfaces.go       // 共享接口定义

// ❌ 错误的组织方式
service/
├── services.go         // 多个服务混在一个文件
└── utils.go           // 功能不明确
```

### ⚛️ 前端代码组织

#### React组件结构

**组件分类**:
```
src/components/
├── ui/                 // 纯UI组件
│   ├── Button/
│   │   ├── index.tsx
│   │   ├── Button.module.css
│   │   └── Button.test.tsx
│   └── Modal/
├── business/           // 业务组件
│   ├── MockRuleEditor/
│   └── ProjectSelector/
└── layout/             // 布局组件
    ├── Header/
    └── Sidebar/
```

**组件命名规范**:
- **PascalCase**: 组件文件名和组件名使用PascalCase
- **语义化**: 组件名应该清晰表达其功能
- **单一职责**: 每个组件只负责一个明确的UI功能

#### 状态管理规范

**状态分层**:
```
src/store/
├── global/             // 全局状态
│   ├── auth.ts
│   └── settings.ts
├── features/           // 特性状态
│   ├── mock.ts
│   └── project.ts
└── hooks/              // 状态hooks
    ├── useAuth.ts
    └── useMock.ts
```

---

## 4. 配置管理规范

### ⚙️ 配置文件组织

#### 环境配置分离
```
config.yaml              # 生产环境配置 (默认)
config.dev.yaml          # 开发环境配置
config.test.yaml         # 测试环境配置
config.staging.yaml      # 预发布环境配置 (可选)
```

#### 配置文件结构规范

**通用配置结构**:
```yaml
# config.yaml 示例
server:
  admin:
    host: "0.0.0.0"
    port: 8080
  mock:
    host: "0.0.0.0"
    port: 9090

database:
  mongodb:
    uri: "mongodb://localhost:27017"
    database: "mockserver"
  redis:
    addr: "localhost:6379"
    db: 0

logging:
  level: "info"
  format: "json"
  output: "stdout"

features:
  enable_graphql: true
  enable_metrics: true
  enable_tracing: false
```

### 🔐 配置安全规范

#### 敏感信息处理
- ❌ 配置文件中不能包含明文密码
- ❌ 不能提交API密钥到版本控制
- ✅ 使用环境变量存储敏感信息
- ✅ 生产环境使用密钥管理系统

#### 配置验证
```go
// ✅ 配置结构验证示例
type Config struct {
    Server   ServerConfig   `yaml:"server" validate:"required"`
    Database DatabaseConfig `yaml:"database" validate:"required"`
    Logging  LoggingConfig  `yaml:"logging" validate:"required"`
}

func (c *Config) Validate() error {
    return validator.New().Struct(c)
}
```

---

## 5. 测试组织标准

### 🧪 测试目录结构

#### 完整测试体系
```
tests/
├── unit/                    # 单元测试
│   ├── internal/
│   │   ├── service/
│   │   ├── repository/
│   │   └── engine/
│   └── web/frontend/src/
├── integration/             # 集成测试
│   ├── api/
│   ├── database/
│   └── cache/
├── e2e/                     # 端到端测试
│   ├── user-workflows/
│   └── admin-workflows/
├── performance/             # 性能测试
│   ├── load/
│   └── stress/
├── fixtures/                # 测试固件
│   ├── data/
│   └── mock/
├── coverage/                # 覆盖率报告
├── reports/                 # 测试报告
└── scripts/                 # 测试脚本
    ├── run-all.sh
    ├── run-unit.sh
    └── run-e2e.sh
```

### 📊 测试覆盖率标准

#### 覆盖率要求
- **单元测试**: ≥ 80%
- **集成测试**: ≥ 70%
- **E2E测试**: ≥ 60%
- **总体覆盖率**: ≥ 75%

#### 测试命名规范
```go
// ✅ 正确的测试命名
func TestMockService_CreateRule(t *testing.T) {
    // 测试创建规则功能
}

func TestMockService_CreateRule_InvalidInput(t *testing.T) {
    // 测试创建规则的无效输入场景
}

func TestMockService_CreateRule_DuplicateName(t *testing.T) {
    // 测试创建规则的重复名称场景
}
```

---

## 6. 文档管理体系

### 📚 文档组织结构

#### 文档分类体系
```
docs/
├── api/                     # API文档
│   ├── rest/
│   │   ├── v1/
│   │   └── v2/
│   └── graphql/
├── architecture/            # 架构文档
│   ├── system-design.md
│   ├── data-model.md
│   └── deployment.md
├── development/             # 开发文档
│   ├── setup.md
│   ├── coding-standards.md
│   └── testing-guide.md
├── deployment/              # 部署文档
│   ├── docker.md
│   ├── kubernetes.md
│   └── production.md
├── user-guide/              # 用户指南
│   ├── getting-started.md
│   ├── features.md
│   └── troubleshooting.md
└── changelog/               # 变更日志
    ├── v0.8.0.md
    └── v0.7.0.md
```

### 📝 文档编写规范

#### Markdown标准
- **标题层级**: 最多使用4级标题 (H1-H4)
- **代码块**: 指定语言类型，使用语法高亮
- **链接**: 使用相对链接，避免硬编码URL
- **图片**: 优化图片大小，提供alt文本

#### 文档更新流程
1. **功能变更**: 必须同步更新相关文档
2. **API变更**: 必须更新API文档和示例
3. **架构变更**: 必须更新架构图和设计文档
4. **版本发布**: 必须更新CHANGELOG和RELEASE_NOTES

---

## 7. 构建与部署规范

### 🔨 构建脚本组织

#### Makefile结构规范
```makefile
# Makefile 示例结构
.PHONY: help build test clean deploy

# 默认目标
help:
	@echo "Available targets..."

# 开发相关
dev-setup dev-start dev-stop dev-reset

# 构建相关
build build-backend build-frontend build-all

# 测试相关
test test-unit test-integration test-e2e test-all

# 部署相关
deploy-staging deploy-prod docker-build docker-push

# 维护相关
clean format lint security-check

# 发布相关
release version-tag changelog
```

### 🐳 Docker规范

#### 镜像组织
```
docker/
├── Dockerfile               # 基础运行时镜像
├── Dockerfile.build         # 构建时镜像
├── Dockerfile.test          # 测试环境镜像
├── docker-compose.yml       # 开发环境编排
├── docker-compose.prod.yml  # 生产环境编排
├── docker-compose.test.yml  # 测试环境编排
└── scripts/
    ├── build.sh
    ├── push.sh
    └── run.sh
```

#### Dockerfile最佳实践
```dockerfile
# ✅ 多阶段构建示例
FROM golang:1.24-alpine AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o mockserver ./cmd/mockserver

FROM alpine:latest AS runtime
RUN apk --no-cache add ca-certificates tzdata
WORKDIR /root/
COPY --from=builder /app/mockserver .
COPY --from=builder /app/config.yaml .
EXPOSE 8080 9090
CMD ["./mockserver"]
```

---

## 8. 脚本管理指南

### 📜 脚本分类组织

#### 脚本目录结构
```
scripts/
├── dev/                     # 开发环境脚本
│   ├── start-all.sh       # 启动开发环境
│   ├── stop-all.sh        # 停止开发环境
│   ├── reset-db.sh        # 重置数据库
│   └── migrate.sh         # 数据迁移
├── deploy/                  # 部署脚本
│   ├── deploy-staging.sh  # 部署到预发布
│   ├── deploy-prod.sh     # 部署到生产
│   ├── rollback.sh        # 版本回滚
│   └── health-check.sh    # 健康检查
├── maintenance/             # 维护脚本
│   ├── backup.sh          # 数据备份
│   ├── restore.sh         # 数据恢复
│   ├── cleanup.sh         # 清理脚本
│   └── monitor.sh         # 监控脚本
└── quality/                 # 质量检查脚本
    ├── security-check.sh  # 安全检查
    ├── performance-test.sh # 性能测试
    ├── dependency-check.sh # 依赖检查
    └── license-check.sh   # 许可证检查
```

#### 脚本编写规范

**Shell脚本标准**:
```bash
#!/bin/bash

# 脚本头部信息
# Author: MockServer Team
# Created: 2025-11-19
# Description: 启动开发环境

set -euo pipefail  # 严格模式

# 颜色定义
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m'

# 日志函数
log_info() {
    echo -e "${GREEN}[INFO]${NC} $1"
}

log_error() {
    echo -e "${RED}[ERROR]${NC} $1" >&2
}

log_warn() {
    echo -e "${YELLOW}[WARN]${NC} $1"
}

# 主函数
main() {
    log_info "Starting development environment..."
    # 脚本逻辑
}

main "$@"
```

---

## 9. 质量检查清单

### ✅ 定期检查项目

#### 每周检查
- [ ] 代码格式化 (`go fmt`, `npm run format`)
- [ ] 代码质量检查 (`golangci-lint`, `npm run lint`)
- [ ] 单元测试通过率 ≥ 95%
- [ ] 文档更新同步性

#### 每月检查
- [ ] 目录结构完整性
- [ ] 依赖库安全漏洞扫描
- [ ] 性能回归测试
- [ ] API文档一致性
- [ ] 配置文件有效性验证

#### 版本发布前检查
- [ ] 完整测试套件通过 (单元/集成/E2E)
- [ ] 安全审计报告
- [ ] 性能基准测试
- [ ] 文档完整性检查
- [ ] 版本兼容性验证
- [ ] 发布说明完整性

### 📊 质量指标标准

#### 代码质量指标
| 指标 | 目标值 | 检查工具 |
|------|--------|----------|
| 代码覆盖率 | ≥ 75% | go test, npm test |
| 圈复杂度 | ≤ 10 | golangci-lint |
| 重复率 | ≤ 3% | dupl |
| 安全漏洞 | 0 | gosec, npm audit |
| 性能回归 | ≤ 5% | benchcmp |

#### 文档质量指标
| 指标 | 目标值 | 检查方式 |
|------|--------|----------|
| API文档覆盖率 | 100% | 自动化检查 |
| 示例代码有效性 | 100% | 运行验证 |
| 文档更新及时性 | 24小时内 | 版本控制检查 |
| 文档链接有效性 | 100% | 链接检查工具 |

---

## 10. 腐化预防机制

### 🛡️ 自动化防护

#### Git Hooks
```bash
# .git/hooks/pre-commit
#!/bin/bash
# 提交前检查

# 格式化检查
go fmt ./...
if [[ $(git status --porcelain) ]]; then
    echo "代码已格式化，请重新提交"
    exit 1
fi

# 质量检查
golangci-lint run

# 单元测试
go test ./...
```

#### CI/CD集成
```yaml
# .github/workflows/quality-check.yml
name: Quality Check
on: [push, pull_request]

jobs:
  quality:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3
      - name: Setup Go
        uses: actions/setup-go@v3
        with:
          go-version: 1.24

      - name: Run tests
        run: make test-all

      - name: Check code quality
        run: make lint

      - name: Security scan
        run: make security-check
```

### 🔍 持续监控

#### 目录结构监控
```bash
# scripts/quality/structure-check.sh
#!/bin/bash

# 检查目录结构完整性
check_structure() {
    local required_dirs=(
        "cmd/mockserver"
        "internal"
        "pkg"
        "web/frontend"
        "tests"
        "docs"
        "scripts"
    )

    for dir in "${required_dirs[@]}"; do
        if [[ ! -d "$dir" ]]; then
            echo "❌ 缺少必需目录: $dir"
            return 1
        fi
    done

    echo "✅ 目录结构检查通过"
}

# 检查命名规范
check_naming() {
    # 检查是否有大写目录名
    if find . -type d -name "*[A-Z]*" | grep -v "./.git" | grep -v "./node_modules"; then
        echo "❌ 发现大写目录名"
        return 1
    fi

    echo "✅ 命名规范检查通过"
}

check_structure
check_naming
```

#### 依赖管理监控
```bash
# scripts/quality/dependency-check.sh
#!/bin/bash

# Go依赖检查
echo "检查Go依赖漏洞..."
go list -json -m all | nancy sleuth

# Node.js依赖检查
echo "检查Node.js依赖漏洞..."
cd web/frontend && npm audit --audit-level high
```

### 🚨 早期预警机制

#### 腐化信号识别
1. **技术债务增加**: 代码复杂度持续上升
2. **测试覆盖率下降**: 低于设定的阈值
3. **文档滞后**: 代码变更后文档未更新
4. **配置膨胀**: 配置文件数量和复杂度增加
5. **依赖混乱**: 版本冲突和安全漏洞

#### 自动化报告
```bash
# scripts/quality/health-report.sh
#!/bin/bash

# 生成项目健康报告
generate_health_report() {
    echo "# MockServer 项目健康报告 - $(date)" > health-report.md
    echo "" >> health-report.md

    # 代码质量
    echo "## 📊 代码质量" >> health-report.md
    echo "- 测试覆盖率: $(go test -cover ./... | tail -1)" >> health-report.md
    echo "- 代码行数: $(find . -name "*.go" | xargs wc -l | tail -1)" >> health-report.md

    # 依赖状态
    echo "## 📦 依赖状态" >> health-report.md
    echo "- Go模块数量: $(go list -m all | wc -l)" >> health-report.md
    echo "- Node.js依赖: $(cd web/frontend && npm list --depth=0 | grep -c "├\|└")" >> health-report.md

    # 文档状态
    echo "## 📚 文档状态" >> health-report.md
    echo "- API文档: $(find docs/api -name "*.md" | wc -l)" >> health-report.md
    echo "- 架构文档: $(find docs/architecture -name "*.md" | wc -l)" >> health-report.md
}

generate_health_report
echo "健康报告已生成: health-report.md"
```

---

## 📋 检查清单总结

### 🔄 每日检查 (开发者)
- [ ] 代码格式化
- [ ] 单元测试通过
- [ ] 提交信息规范
- [ ] 没有硬编码的敏感信息

### 📅 每周检查 (团队负责人)
- [ ] 代码质量报告
- [ ] 测试覆盖率达标
- [ ] 依赖安全扫描
- [ ] 文档同步更新

### 🗓️ 每月检查 (架构师)
- [ ] 目录结构完整性
- [ ] 架构一致性验证
- [ ] 性能回归测试
- [ ] 技术债务评估

### 🚀 发布前检查 (DevOps)
- [ ] 完整测试套件通过
- [ ] 安全审计报告
- [ ] 文档完整性
- [ ] 部署验证

---

## 🎯 实施建议

### 📈 渐进式实施
1. **第一阶段** (1-2周): 建立检查脚本和Git hooks
2. **第二阶段** (1个月): 完善CI/CD质量检查
3. **第三阶段** (持续): 逐步完善文档和规范

### 👥 团队培训
- **开发者培训**: 代码规范和最佳实践
- **架构师培训**: 架构设计和腐化预防
- **DevOps培训**: 自动化工具和监控

### 📊 持续改进
- **定期回顾**: 每月回顾质量指标
- **规范更新**: 根据实践情况更新规范
- **工具升级**: 及时更新和质量检查工具

---

**维护责任人**: MockServer 架构团队
**更新频率**: 重大变更后必须更新，至少每季度回顾一次
**版本**: v1.0
**最后更新**: 2025-11-19