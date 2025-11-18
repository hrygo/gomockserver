# Scripts 目录说明

本目录包含 Mock Server 项目的各类脚本工具，现已整合到 `tests/` 目录下。

## 📁 目录结构

```
tests/
├── scripts/                     # 脚本工具目录
│   ├── coverage/               # 测试覆盖率报告（HTML）
│   ├── run_unit_tests.sh       # 单元测试执行脚本
│   ├── test-env.sh             # Docker 测试环境管理
│   └── README.md               # 本说明文档
├── integration/                # 集成测试目录
├── coverage/                   # 覆盖率数据文件
└── data/                       # 测试数据
```

> **🗑️ 归档说明**: `mvp-test.sh` 已归档至 `docs/archive/scripts/`。该脚本用于 MVP 版本验证，现已被 Makefile 命令替代，建议使用 `make verify` 或 `make qa`。

## 🔧 脚本说明

### 核心测试脚本

#### `run_unit_tests.sh`
**用途**：执行完整的单元测试套件并生成覆盖率报告

**功能**：
- 运行所有单元测试
- 生成模块级覆盖率报告
- 分析测试结果并输出统计信息
- 自动打开覆盖率报告（可选）

**使用**：
```bash
./tests/scripts/run_unit_tests.sh
```

**输出**：
- `tests/coverage/unit-coverage-*.html` - 各模块覆盖率报告
- 终端输出测试统计信息

---

#### `test-env.sh`
**用途**：Docker 测试环境管理

**功能**：
- 启动测试环境（MongoDB + Mock Server）
- 停止并清理测试环境
- 查看测试环境状态
- 运行冒烟测试

**使用**：
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

---

#### `test.sh`
**用途**：快速功能测试脚本

**功能**：
- 验证服务健康状态
- 测试基本 CRUD 操作
- 测试 Mock 接口功能
- 适合快速验证部署是否正常

**使用**：
```bash
# 确保服务已启动
docker-compose up -d

# 运行测试
./tests/scripts/test.sh
```

**注意**：需要服务在 8080（管理API）和 9090（Mock服务）端口运行

---

### 辅助工具脚本

> **⚠️ 已弃用**: `mvp-test.sh` 已归档，不再使用。请使用下面的 Makefile 命令代替。

---

## 📊 Coverage 目录

`coverage/` 目录包含各模块的测试覆盖率报告（HTML格式）：

- `unit-coverage-all.html` - 总体覆盖率
- `unit-coverage-adapter.html` - Adapter 模块
- `unit-coverage-api.html` - API 模块
- `unit-coverage-config.html` - Config 模块
- `unit-coverage-engine.html` - Engine 模块
- `unit-coverage-executor.html` - Executor 模块
- `unit-coverage-repository.html` - Repository 模块
- `unit-coverage-service.html` - Service 模块

**查看方式**：
```bash
# 在浏览器中打开
open tests/coverage/unit-coverage-all.html  # macOS
xdg-open tests/coverage/unit-coverage-all.html  # Linux
```

---

## 🚀 常用工作流

### 开发时运行测试
```bash
# 1. 运行单元测试
./tests/scripts/run_unit_tests.sh

# 2. 查看覆盖率报告
open tests/coverage/unit-coverage-all.html
```

### 启动测试环境验证
```bash
# 1. 启动测试环境
./tests/scripts/test-env.sh start

# 2. 运行快速测试
./tests/scripts/test.sh

# 3. 停止环境
./tests/scripts/test-env.sh stop
```

### 发布前完整测试

**推荐使用 Makefile 命令：**
```bash
# 1. 质量检查（格式化+静态分析+单元测试）
make qa

# 2. 推送前检查（包含集成测试）
make pre-push

# 3. 完整验证
 make verify

# 4. 生成覆盖率报告
make test-coverage
```

**或使用命令别名：**
```bash
make t              # 别名: make test
make c              # 别名: make test-coverage
```

---

## 📝 脚本维护指南

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

---

## 🔗 相关文档

- [README.md](../README.md) - 项目主文档
- [CONTRIBUTING.md](../CONTRIBUTING.md) - 贡献指南
- [Makefile](../Makefile) - 构建脚本（推荐使用 `make help` 查看所有命令）
- [DEPLOYMENT.md](../DEPLOYMENT.md) - 部署指南
- [docs/archive/INDEX.md](../docs/archive/INDEX.md) - 归档文档索引（包含已弃用脚本）

## 🆕 Makefile 快捷命令

**推荐使用 Makefile 命令代替直接执行脚本：**

| 脚本 | Makefile 命令 | 说明 |
|------|--------------|------|
| `run_unit_tests.sh` | `make test-coverage` | 单元测试+覆盖率报告 |
| `test-env.sh start` | `make docker-test-up` | 启动测试环境 |
| `test-env.sh stop` | `make docker-test-down` | 停止测试环境 |
| `mvp-test.sh` (已弃用) | `make verify` 或 `make qa` | 完整验证 |

**查看所有可用命令：**
```bash
make help
```

---

**最后更新**: 2025-01-21  
**维护者**: Mock Server 团队
