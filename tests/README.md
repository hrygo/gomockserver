# MockServer 测试框架

MockServer 拥有完整的测试体系，包括单元测试、集成测试、性能测试等多种测试类型，确保系统的稳定性和可靠性。

## 📁 测试目录结构

```
tests/
├── scripts/                     # 测试工具脚本
│   ├── run_unit_tests.sh       # 单元测试执行脚本
│   ├── test-env.sh             # Docker 测试环境管理
│   ├── coverage/               # 覆盖率HTML报告
│   └── README.md               # 脚本使用说明
├── integration/                # 集成测试目录
│   ├── README.md               # 集成测试文档
│   ├── run_all_e2e_tests.sh    # E2E测试执行脚本
│   ├── e2e_test.sh             # 基础功能测试
│   ├── advanced_e2e_test.sh    # 高级功能测试
│   ├── stress_e2e_test.sh      # 压力测试
│   ├── websocket_e2e_test.sh   # WebSocket测试
│   ├── edge_case_e2e_test.sh   # 边界条件测试
│   ├── install_tools.sh        # 测试工具安装脚本
│   ├── lib/                    # 测试框架库
│   │   ├── test_framework.sh   # 测试框架核心
│   │   └── tool_installer.sh   # 工具安装器
│   └── platform_compatibility_test.sh  # 平台兼容性测试
├── coverage/                   # 覆盖率数据文件
│   └── *.out                   # 覆盖率原始数据
└── data/                       # 测试数据
    └── init-mongo.js           # MongoDB初始化数据
```

## 🧪 测试类型和用途

### 1. 单元测试 (Unit Tests)
**位置**: Go源码文件中的 `*_test.go` 文件

**用途**: 测试单个函数和方法的功能正确性

**运行方式**:
```bash
# 使用脚本运行
./tests/scripts/run_unit_tests.sh

# 使用 Makefile
make test-unit
make test-coverage
```

**覆盖率报告**: `tests/coverage/unit-coverage-*.html`

### 2. 集成测试 (Integration Tests)
**位置**: `tests/integration/` 目录

**用途**: 测试系统各组件之间的协作，验证完整的业务流程

**测试覆盖**:
- **基础功能测试** - CRUD操作、Mock服务
- **高级功能测试** - 复杂匹配、动态响应
- **WebSocket测试** - 实时通信功能
- **边界条件测试** - 异常场景处理
- **压力测试** - 性能和负载测试
- **平台兼容性测试** - 跨平台验证

**运行方式**:
```bash
# 运行完整E2E测试套件
./tests/integration/run_all_e2e_tests.sh

# 运行单个测试
./tests/integration/e2e_test.sh              # 基础功能
./tests/integration/advanced_e2e_test.sh      # 高级功能
./tests/integration/stress_e2e_test.sh        # 压力测试
./tests/integration/websocket_e2e_test.sh     # WebSocket
./tests/integration/edge_case_e2e_test.sh     # 边界条件
```

**测试结果**: 94% 通过率 (48/51 测试用例)

### 3. 性能测试 (Performance Tests)
**位置**: `tests/integration/stress_e2e_test.sh`

**用途**: 验证系统在不同负载下的性能表现

**性能指标**:
- **峰值 QPS**: 3,366
- **最大并发**: 200+ 连接
- **响应时间**: P95 < 100ms

## 🚀 快速开始

### 环境准备
```bash
# 安装测试工具
./tests/integration/install_tools.sh --all

# 或使用 Makefile
make test-install-tools
```

### 运行测试

#### 1. 开发时测试
```bash
# 单元测试 + 覆盖率
./tests/scripts/run_unit_tests.sh

# 查看覆盖率报告
open tests/coverage/unit-coverage-all.html
```

#### 2. 完整集成测试
```bash
# 运行所有集成测试
./tests/integration/run_all_e2e_tests.sh

# 查看测试报告
cat docs/testing/TEST_EXECUTION_SUMMARY.md
```

#### 3. 生产前验证
```bash
# 使用 Makefile 进行完整验证
make qa                    # 质量检查
make pre-push             # 推送前检查
make verify               # 完整验证
```

## 📊 测试覆盖率

### 当前覆盖率状态
- **单元测试覆盖率**: 69.3%+ (核心模块 80%+)
- **E2E测试通过率**: 94% (48/51 测试用例)
- **功能覆盖**: 100% (核心功能完全验证)

### 覆盖率报告查看
```bash
# HTML覆盖率报告
open tests/coverage/unit-coverage-all.html

# 各模块覆盖率
ls tests/coverage/unit-coverage-*.html
```

## 🔧 测试工具

### 自动化工具安装
**脚本**: `tests/integration/install_tools.sh`

**支持的工具**:
- `curl` - HTTP请求工具
- `jq` - JSON处理工具
- `wrk` - 压力测试工具
- `websocat` - WebSocket测试工具
- `python3` - 脚本支持

**使用方式**:
```bash
# 安装所有工具
./tests/integration/install_tools.sh --all

# 安装特定类型工具
./tests/integration/install_tools.sh --basic     # 基础工具
./tests/integration/install_tools.sh --stress    # 压力测试工具
./tests/integration/install_tools.sh --websocket # WebSocket工具

# 检查工具状态
./tests/integration/install_tools.sh --check
```

### 跨平台支持
测试框架支持以下平台：
- ✅ **macOS** (Darwin) - 完全支持
- ✅ **Linux** (Ubuntu/CentOS) - 完全支持
- 🔄 **Windows** - 计划中

## 🐳 Docker测试环境

### 使用脚本管理
```bash
# 启动测试环境
./tests/scripts/test-env.sh start

# 停止测试环境
./tests/scripts/test-env.sh stop

# 查看状态
./tests/scripts/test-env.sh status

# 运行冒烟测试
./tests/scripts/test-env.sh test
```

### 使用 Docker Compose
```bash
# 启动测试环境
docker-compose -f docker-compose.test.yml up -d

# 运行测试
docker-compose -f docker-compose.test.yml --profile integration run --rm test-runner

# 清理环境
docker-compose -f docker-compose.test.yml down -v
```

## 📈 测试结果分析

### 测试报告位置
- **执行总结**: `docs/testing/TEST_EXECUTION_SUMMARY.md`
- **完整报告**: `docs/testing/E2E_TEST_REPORT.md`
- **性能基准**: 包含在完整报告中

### 关键性能指标
| 负载级别 | 并发数 | 平均 QPS | 评级 |
|----------|--------|----------|------|
| 轻量级 | 10 | 2,295 | 🟢 优秀 |
| 中等负载 | 50 | 3,228 | 🟢 优秀 |
| 高负载 | 100 | 3,297 | 🟢 优秀 |
| 极高负载 | 200 | 3,366 | 🟢 优秀 |

## 🛠️ 开发指南

### 添加新的测试用例

#### 1. 添加单元测试
```bash
# 在相应模块下创建 *_test.go 文件
# 例如: internal/service/myservice_test.go

# 运行测试
go test ./internal/service/...
```

#### 2. 添加集成测试
```bash
# 在 tests/integration/ 下创建测试脚本
# 使用测试框架库: tests/integration/lib/test_framework.sh

# 示例
cat > tests/integration/my_test.sh << 'EOF'
#!/bin/bash
source "$(dirname "$0")/lib/test_framework.sh"

# 测试逻辑
test_case "我的测试用例" {
    # 测试步骤
}
EOF

chmod +x tests/integration/my_test.sh
```

### 测试最佳实践

1. **测试命名**: 使用描述性的测试名称
2. **测试隔离**: 每个测试用例独立运行
3. **数据清理**: 测试后自动清理数据
4. **错误处理**: 适当的错误检查和报告
5. **文档更新**: 添加新测试后更新文档

## 🔍 故障排除

### 常见问题

#### 1. 工具缺失
```bash
# 重新安装工具
./tests/integration/install_tools.sh --all

# 检查工具状态
./tests/integration/install_tools.sh --check
```

#### 2. 权限问题
```bash
# 添加执行权限
chmod +x tests/integration/*.sh
chmod +x tests/scripts/*.sh
```

#### 3. 端口冲突
```bash
# 检查端口占用
lsof -i :8080  # Admin API
lsof -i :9090  # Mock API

# 停止冲突服务
make stop-all
```

#### 4. Docker问题
```bash
# 检查Docker状态
docker ps
docker-compose ps

# 重建环境
docker-compose -f docker-compose.test.yml down -v
docker-compose -f docker-compose.test.yml up -d
```

### 调试技巧

1. **查看详细日志**:
```bash
# 启用详细输出
DEBUG=1 ./tests/integration/run_all_e2e_tests.sh
```

2. **单步调试**:
```bash
# 运行单个测试
./tests/integration/e2e_test.sh

# 跳过服务器启动
SKIP_SERVER_START=true ./tests/integration/e2e_test.sh
```

3. **环境检查**:
```bash
# 检查服务状态
curl http://localhost:8080/api/v1/system/health
curl http://localhost:9090/health
```

## 📚 相关文档

- [集成测试详细文档](integration/README.md)
- [脚本工具使用说明](scripts/README.md)
- [项目主文档](../README.md)
- [贡献指南](../CONTRIBUTING.md)
- [Makefile 命令参考](../Makefile)
- [系统架构文档](../docs/ARCHITECTURE.md)
- [知识体系文档](../docs/KNOWLEDGE_SYSTEM.md)

## 🆕 Makefile 快捷命令

| 功能 | 命令 | 说明 |
|------|------|------|
| 单元测试 | `make test-unit` | 运行所有单元测试 |
| 覆盖率报告 | `make test-coverage` | 生成覆盖率报告 |
| 集成测试 | `make test-integration` | 运行集成测试 |
| 完整测试 | `make test-all` | 运行所有测试 |
| 质量检查 | `make qa` | 格式化+静态分析+测试 |
| 完整验证 | `make verify` | 推送前完整检查 |
| Docker测试 | `make test-docker` | Docker环境测试 |

**命令别名**:
- `make t` → `make test`
- `make c` → `make test-coverage`

---

**文档版本**: 2.0
**创建日期**: 2025-11-18
**最后更新**: 2025-11-18
**维护者**: MockServer 团队