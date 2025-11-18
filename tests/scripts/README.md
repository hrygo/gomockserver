# 测试脚本工具

本目录包含 MockServer 项目的测试脚本工具，用于简化测试执行、环境管理和覆盖率生成。

## 📁 脚本列表

### 核心脚本

#### `run_unit_tests.sh`
**用途**: 执行完整的单元测试套件并生成覆盖率报告

**功能**:
- 运行所有单元测试
- 生成模块级覆盖率报告
- 分析测试结果并输出统计信息
- 自动清理历史测试文件

**使用**:
```bash
./tests/scripts/run_unit_tests.sh
```

**输出**:
- `tests/coverage/unit-coverage-*.html` - 各模块覆盖率报告
- `docs/testing/reports/` - 测试报告和总结

---

#### `test-env.sh`
**用途**: Docker 测试环境管理

**功能**:
- 启动测试环境（MongoDB + Mock Server）
- 停止并清理测试环境
- 查看测试环境状态
- 运行冒烟测试

**使用**:
```bash
# 启动测试环境
./tests/scripts/test-env.sh up

# 启动完整环境（包含 Redis）
./tests/scripts/test-env.sh up-full

# 停止测试环境
./tests/scripts/test-env.sh down

# 重启测试环境
./tests/scripts/test-env.sh restart

# 查看状态
./tests/scripts/test-env.sh ps

# 查看日志
./tests/scripts/test-env.sh logs

# 运行集成测试
./tests/scripts/test-env.sh test

# 运行性能测试
./tests/scripts/test-env.sh perf

# 清理环境
./tests/scripts/test-env.sh clean

# 重建镜像
./tests/scripts/test-env.sh build

# 显示帮助
./tests/scripts/test-env.sh help
```

**环境变量**:
- `ADMIN_API`: 管理API地址 (默认: http://localhost:8080/api/v1)
- `MOCK_API`: Mock服务地址 (默认: http://localhost:9090)

---

### 覆盖率报告

#### HTML 覆盖率报告
`coverage/` 目录包含各模块的测试覆盖率报告（HTML格式）：

- `unit-coverage-all.html` - 总体覆盖率
- `unit-coverage-adapter.html` - Adapter 模块
- `unit-coverage-api.html` - API 模块
- `unit-coverage-engine.html` - Engine 模块
- `unit-coverage-executor.html` - Executor 模块
- `unit-coverage-repository.html` - Repository 模块
- `unit-coverage-service.html` - Service 模块

**查看方式**:
```bash
# 在浏览器中打开
open tests/coverage/unit-coverage-all.html  # macOS
xdg-open tests/coverage/unit-coverage-all.html  # Linux
```

---

## 🚀 使用示例

### 开发工作流

#### 1. 日常开发测试
```bash
# 运行单元测试和覆盖率
./tests/scripts/run_unit_tests.sh

# 查看覆盖率报告
open tests/coverage/unit-coverage-all.html
```

#### 2. 集成测试环境
```bash
# 启动测试环境
./tests/scripts/test-env.sh up

# 运行集成测试
../integration/run_all_e2e_tests.sh

# 停止环境
./tests/scripts/test-env.sh down
```

#### 3. 性能测试
```bash
# 启动性能测试环境
./tests/scripts/test-env.sh up-performance

# 运行性能测试
./tests/scripts/test-env.sh perf

# 查看性能报告
cat /tmp/mockserver_perf_results.txt
```

### 调试和故障排除

#### 查看详细日志
```bash
# 查看服务日志
./tests/scripts/test-env.sh logs mockserver-test

# 查看所有服务日志
./tests/scripts/test-env.sh logs
```

#### 检查环境状态
```bash
# 检查服务状态
./tests/scripts/test-env.sh ps

# 检查健康状态
curl http://localhost:8081/api/v1/system/health
curl http://localhost:9091/health
```

#### 重置环境
```bash
# 完全清理并重建
./tests/scripts/test-env.sh clean
./tests/scripts/test-env.sh build
./tests/scripts/test-env.sh up
```

---

## 📊 与 Makefile 集成

推荐使用 Makefile 命令代替直接执行脚本：

| 脚本命令 | Makefile 命令 | 说明 |
|---------|--------------|------|
| `./tests/scripts/run_unit_tests.sh` | `make test-coverage` | 单元测试+覆盖率报告 |
| `./tests/scripts/test-env.sh up` | `make docker-test-up` | 启动测试环境 |
| `./tests/scripts/test-env.sh down` | `make docker-test-down` | 停止测试环境 |
| `./tests/scripts/test-env.sh test` | `make test-integration` | 运行集成测试 |

**查看所有可用命令**:
```bash
make help
```

---

## 🛠️ 脚本维护

### 添加新脚本

1. 脚本应放在 `tests/scripts/` 目录下
2. 文件名使用小写字母和连字符（如 `my-script.sh`）
3. 添加可执行权限：`chmod +x tests/scripts/my-script.sh`
4. 在文件开头添加清晰的注释说明用途
5. 更新本 README.md 文件

### 脚本规范

- 使用 `#!/bin/bash` 作为 shebang
- 设置 `set -e` 在错误时退出
- 使用有意义的变量名
- 添加错误处理和用户提示
- 使用彩色输出增强可读性

### 路径计算

脚本中使用以下方式计算项目根目录：
```bash
PROJECT_ROOT="$(cd "$(dirname "$0")/../.." && pwd)"
```

---

## 🔗 相关文档

- [测试框架总览](../README.md)
- [集成测试文档](../integration/README.md)
- [项目主文档](../../README.md)
- [Makefile 命令参考](../../Makefile)
- [Docker 测试环境](../../docker-compose.test.yml)

---

**最后更新**: 2025-11-18
**维护者**: MockServer 团队