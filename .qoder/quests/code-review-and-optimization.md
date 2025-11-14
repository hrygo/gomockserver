# 代码审查和优化设计文档

## 文档元数据

| 项目 | 内容 |
|------|------|
| 文档名称 | 代码审查和优化设计文档 |
| 创建时间 | 2025-01-21 |
| 目标版本 | v0.1.2+ |
| 文档类型 | 代码质量改进 |
| 优先级 | P0 - 高优先级 |

## 执行概述

### 目标

对 gomockserver 项目进行全面的代码审查，发现并解决代码质量问题，整合和归并项目文档与脚本资产，保持工程结构清晰明了。

### 范围

- **代码审查**: 全面审查所有 Go 源代码，识别质量问题和技术债务
- **问题修复**: 解决发现的代码质量问题、潜在缺陷和性能隐患
- **文档整合**: 归并和整理项目文档，删除过期和冗余文档
- **脚本优化**: 整合重复脚本，优化脚本功能
- **工程优化**: 改进项目结构，提升可维护性

### 原则

- 保持设计简洁，只解决实际存在的问题
- 不过度工程化，避免引入不必要的复杂性
- 确保所有改动都有明确的价值
- 优先修复高风险和高影响的问题

---

## 一、代码审查发现的问题

### 1.1 代码质量问题

#### 问题 1.1.1: TODO 注释未处理

**位置**:
- `internal/engine/match_engine.go` (3处)
- `internal/executor/mock_executor.go` (6处)

**问题描述**:
代码中存在 9 个 TODO 注释标记，表示未完成的功能实现，这些功能标记为"阶段三实现"但未在任何迭代计划中明确。

**影响**:
- 功能不完整，用户可能遇到未实现的特性
- 代码可维护性差，缺乏明确的实现计划
- 版本发布时容易遗漏功能

**解决方案**:
- 创建技术债务清单文档，记录所有 TODO 项
- 为每个 TODO 分配优先级和目标版本
- 近期不实现的 TODO 改为明确的错误提示

#### 问题 1.1.2: 循环依赖风险

**位置**:
- `internal/api` 与 `internal/service` 包

**问题描述**:
根据 `TASK_COMPLETION_SUMMARY.md` 记录，存在包循环依赖问题导致 API Handler 无法集成。

**影响**:
- 阻碍新功能集成
- 代码架构不清晰
- 难以进行单元测试

**解决方案**:
- 引入接口抽象层，将服务接口定义从实现中分离
- 考虑创建 `internal/interfaces` 包存放接口定义
- 遵循依赖倒置原则（DIP）

#### 问题 1.1.3: Context 超时管理不一致

**位置**:
- `internal/repository/database.go`
- `internal/service/health.go`

**问题描述**:
不同模块使用的 Context 超时时间不一致，且部分硬编码：
- 数据库初始化: 使用配置文件的 timeout
- 健康检查: 硬编码 2 秒
- 数据库关闭: 硬编码 10 秒

**影响**:
- 超时配置不统一，难以调优
- 硬编码值难以在不同环境调整
- 可能导致资源泄漏或超时不合理

**解决方案**:
- 统一 Context 超时管理策略
- 将超时配置提取到配置文件
- 为不同操作类型定义合理的默认超时

#### 问题 1.1.4: 错误处理不完整

**位置**:
- `internal/service/batch_operation_service.go`
- `internal/service/import_export_service.go`

**问题描述**:
批量操作和导入导出服务中使用 `fmt.Sprintf` 格式化错误消息，但错误信息未使用统一的错误码体系。

**影响**:
- 错误信息不规范
- 客户端难以根据错误码进行错误处理
- 不符合已建立的错误码规范

**解决方案**:
- 为批量操作和导入导出服务定义专用错误码
- 使用 `internal/models/errors.go` 中的错误码体系
- 确保所有错误都有对应的错误码

### 1.2 测试覆盖率问题

#### 问题 1.2.1: 新增服务缺少测试

**位置**:
- `internal/service/import_export_service.go` (无测试)
- `internal/service/batch_operation_service.go` (无测试)

**问题描述**:
Sprint 02 新增的导入导出和批量操作服务尚未编写单元测试，覆盖率为 0%。

**影响**:
- Service 层总体覆盖率下降
- 新功能质量无法保证
- 重构风险高

**解决方案**:
- 创建 `import_export_service_test.go`
- 创建 `batch_operation_service_test.go`
- 使用 Mock 框架模拟 Repository 层
- 目标覆盖率: 80%+

#### 问题 1.2.2: 数据库初始化代码未测试

**位置**:
- `internal/repository/database.go`

**问题描述**:
数据库初始化、索引创建等核心代码覆盖率为 0%，这是关键的基础设施代码。

**影响**:
- 数据库连接问题难以发现
- 索引配置错误可能导致生产问题
- 整体项目覆盖率被拉低

**解决方案**:
- 创建数据库初始化集成测试
- 验证索引创建逻辑
- 测试连接失败场景

### 1.3 性能和资源管理问题

#### 问题 1.3.1: MongoDB Cursor 资源泄漏风险

**位置**:
- `internal/repository/project_repository.go`
- `internal/repository/rule_repository.go`

**问题描述**:
部分查询使用 `defer cursor.Close(ctx)`，但在错误情况下可能未正确关闭 cursor。

**影响**:
- 资源泄漏
- 连接池耗尽
- 性能下降

**解决方案**:
- 检查所有 cursor 使用，确保正确关闭
- 在 cursor.All() 错误时也要关闭 cursor
- 添加 lint 规则检测资源泄漏

#### 问题 1.3.2: 无并发控制

**位置**:
- `internal/service/batch_operation_service.go`

**问题描述**:
批量操作使用串行处理，没有并发控制，大批量操作性能较差。

**影响**:
- 批量操作性能低下
- 用户体验差
- 无法充分利用系统资源

**解决方案**:
- 引入并发处理机制
- 使用 worker pool 模式
- 添加并发数量配置
- 实现优雅的错误聚合

### 1.4 配置管理问题

#### 问题 1.4.1: 硬编码配置值

**位置**:
- 健康检查超时: 2秒 (硬编码)
- 数据库关闭超时: 10秒 (硬编码)
- 慢请求阈值: 1秒 (硬编码)

**问题描述**:
多处使用硬编码的超时和阈值配置，无法根据不同环境调整。

**影响**:
- 配置不灵活
- 不同环境难以适配
- 调优困难

**解决方案**:
- 将硬编码值移至配置文件
- 提供合理的默认值
- 支持环境变量覆盖

---

## 二、文档整合方案

### 2.1 文档现状分析

#### 2.1.1 根目录文档 (16个)

**保留的核心文档**:
- `README.md` - 项目主文档
- `CHANGELOG.md` - 变更日志
- `CONTRIBUTING.md` - 贡献指南
- `DEPLOYMENT.md` - 部署文档
- `LICENSE` - 许可证
- `PROJECT_SUMMARY.md` - 项目概述

**需要归档的文档**:
- `MVP_MILESTONE_SUMMARY.md` - 归档至 `docs/archive/milestones/`
- `MVP_RELEASE_CHECKLIST.md` - 归档至 `docs/archive/milestones/`
- `RELEASE_NOTES_v0.1.0.md` - 归档至 `docs/archive/releases/`
- `SPRINT_01_COMPLETION_REPORT.md` - 归档至 `docs/archive/sprints/`
- `SPRINT_01_SUMMARY.md` - 归档至 `docs/archive/sprints/`
- `SPRINT_02_EXECUTION_SUMMARY.md` - 归档至 `docs/archive/sprints/`
- `SPRINT_02_SUMMARY.md` - 归档至 `docs/archive/sprints/`
- `TASK_COMPLETION_SUMMARY.md` - 归档至 `docs/archive/tasks/`

**归档原则**:
- MVP 里程碑文档已完成，应归档保存
- Sprint 执行报告已完成，应归档保存
- 保持根目录简洁，只保留用户常用文档

#### 2.1.2 docs/archive 目录 (22个文件)

**当前状态**:
包含大量历史文档和脚本，部分已过期。

**优化方案**:
1. 按类型创建子目录结构:
   - `docs/archive/milestones/` - 里程碑文档
   - `docs/archive/sprints/` - Sprint 报告
   - `docs/archive/tasks/` - 任务完成报告
   - `docs/archive/testing/` - 测试相关文档
   - `docs/archive/scripts/` - 废弃脚本
   - `docs/archive/reports/` - 已存在，保留

2. 移除完全过期的文档:
   - 识别内容已被新文档覆盖的文档
   - 识别不再需要的临时文档

3. 添加归档索引:
   - 创建 `docs/archive/INDEX.md`
   - 记录归档文档的分类和用途
   - 方便查找历史文档

### 2.2 文档整合操作

#### 整合操作表

| 操作 | 源文件 | 目标位置 | 原因 |
|------|--------|---------|------|
| 移动 | `MVP_MILESTONE_SUMMARY.md` | `docs/archive/milestones/` | MVP 已完成 |
| 移动 | `MVP_RELEASE_CHECKLIST.md` | `docs/archive/milestones/` | MVP 已完成 |
| 移动 | `RELEASE_NOTES_v0.1.0.md` | `docs/archive/releases/` | 历史发布记录 |
| 移动 | `SPRINT_01_COMPLETION_REPORT.md` | `docs/archive/sprints/` | Sprint 已完成 |
| 移动 | `SPRINT_01_SUMMARY.md` | `docs/archive/sprints/` | Sprint 已完成 |
| 移动 | `SPRINT_02_EXECUTION_SUMMARY.md` | `docs/archive/sprints/` | Sprint 已完成 |
| 移动 | `SPRINT_02_SUMMARY.md` | `docs/archive/sprints/` | Sprint 已完成 |
| 移动 | `TASK_COMPLETION_SUMMARY.md` | `docs/archive/tasks/` | 任务已完成 |
| 移动 | `docs/archive/cleanup_docs.sh` | `docs/archive/scripts/` | 废弃脚本 |
| 移动 | `docs/archive/test-completion-report.sh` | `docs/archive/scripts/` | 废弃脚本 |
| 创建 | - | `docs/archive/INDEX.md` | 归档索引 |

### 2.3 创建文档管理规范

#### 文档分类标准

| 文档类型 | 位置 | 命名规范 | 生命周期 |
|---------|------|---------|---------|
| 用户文档 | 根目录 | 大写字母+下划线 | 长期维护 |
| 架构文档 | `docs/` | 大写字母+下划线 | 长期维护 |
| API 文档 | `docs/api/` | 小写字母+连字符 | 随代码更新 |
| 指南文档 | `docs/guides/` | 小写字母+连字符 | 长期维护 |
| 里程碑文档 | `docs/archive/milestones/` | 语义化命名 | 归档保存 |
| Sprint 报告 | `docs/archive/sprints/` | sprint-XX-xxx.md | 归档保存 |
| 测试报告 | `docs/archive/testing/` | 带时间戳 | 定期清理 |

---

## 三、脚本资产整合方案

### 3.1 脚本现状分析

#### 3.1.1 当前脚本清单

| 脚本名称 | 功能 | 状态 | 建议 |
|---------|------|------|------|
| `scripts/run_unit_tests.sh` | 单元测试执行 | 活跃使用 | 保留优化 |
| `scripts/test-env.sh` | 测试环境管理 | 活跃使用 | 保留 |
| `scripts/test.sh` | 快速测试 | 活跃使用 | 保留 |
| `scripts/mvp-test.sh` | MVP 综合测试 | 历史用途 | 评估后归档 |
| `tests/integration/e2e_test.sh` | E2E 测试 | 活跃使用 | 保留 |
| `tests/performance/run_perf_tests.sh` | 性能测试 | 活跃使用 | 保留 |
| `tests/smoke/smoke_test.sh` | 冒烟测试 | 待评估 | 评估功能 |

### 3.2 脚本优化重点

#### 优化 3.2.1: run_unit_tests.sh 改进

**当前问题**:
- 清理逻辑较复杂
- 报告文件管理策略需统一
- 与 Makefile 命令有重叠

**优化方案**:
- 简化清理逻辑，统一清理策略
- 与 Makefile test-coverage 命令集成
- 添加失败时的诊断信息输出

#### 优化 3.2.2: 脚本功能重叠处理

**发现的重叠**:
1. `mvp-test.sh` 与 `run_unit_tests.sh` + `e2e_test.sh` 功能重叠
2. Makefile 中的 test 相关命令与脚本重复

**解决方案**:
- 评估 `mvp-test.sh` 是否仍需要
- 如不需要，归档至 `docs/archive/scripts/`
- 在 `scripts/README.md` 中明确各脚本用途
- 推荐使用 Makefile 命令代替直接运行脚本

### 3.3 脚本管理规范

#### 脚本组织结构

建议的脚本目录结构:

```
scripts/
├── README.md              # 脚本使用说明
├── run_unit_tests.sh      # 单元测试
├── test-env.sh            # 环境管理
├── test.sh                # 快速测试
└── coverage/              # 覆盖率报告输出
```

#### 脚本命名和编写规范

| 规范项 | 要求 | 示例 |
|-------|------|------|
| 命名 | 小写字母+连字符 | run-unit-tests.sh |
| Shebang | 使用 `#!/bin/bash` | - |
| 错误处理 | 使用 `set -e` | - |
| 变量引用 | 使用双引号 | `"$variable"` |
| 函数命名 | 小写字母+下划线 | `cleanup_files()` |
| 注释 | 关键逻辑添加注释 | - |

---

## 四、工程结构优化

### 4.1 项目目录结构优化

#### 4.1.1 当前结构

```
gomockserver/
├── cmd/                  # 命令行入口
├── internal/             # 内部代码
├── pkg/                  # 可复用包
├── scripts/              # 脚本
├── tests/                # 测试
├── docs/                 # 文档
├── web/                  # Web 资源
└── (16个根目录文档)       # 过多
```

#### 4.1.2 优化后结构

```
gomockserver/
├── cmd/                  # 命令行入口
├── internal/             # 内部代码
│   ├── adapter/
│   ├── api/
│   ├── config/
│   ├── engine/
│   ├── executor/
│   ├── models/
│   ├── repository/
│   └── service/
├── pkg/                  # 可复用包
├── scripts/              # 脚本 (简化)
├── tests/                # 测试
├── docs/                 # 文档 (结构化)
│   ├── api/
│   ├── guides/
│   └── archive/          # 归档 (细分子目录)
│       ├── milestones/
│       ├── sprints/
│       ├── tasks/
│       ├── testing/
│       ├── scripts/
│       └── INDEX.md
├── web/                  # Web 资源
└── (6个核心文档)          # 精简后
```

### 4.2 Makefile 优化

#### 4.2.1 当前问题

- 命令分组清晰，但部分命令可合并
- test-repository、test-service、test-api 可能使用频率不高
- 缺少代码质量检查的快捷命令

#### 4.2.2 优化建议

**新增命令**:
- `make qa` - 快速质量检查 (fmt + vet + lint + test-unit)
- `make pre-push` - 推送前检查 (qa + test-integration)

**命令别名**:
- `make t` - 别名 `make test`
- `make c` - 别名 `make coverage`

### 4.3 配置文件优化

#### 4.3.1 配置结构改进

**新增配置项** (config.yaml):

```yaml
# 性能和超时配置
performance:
  health_check_timeout: 2s
  database_close_timeout: 10s
  slow_request_threshold: 1s
  batch_operation_concurrency: 10
  context_timeout:
    default: 30s
    database_query: 10s
    database_write: 15s
    health_check: 2s
```

#### 4.3.2 配置验证

**需要实现**:
- 配置文件 schema 验证
- 启动时配置合法性检查
- 配置热重载支持 (可选)

---

## 五、代码修复优先级

### 5.1 P0 - 必须立即修复

| 问题 | 影响 | 修复工作量 | 目标版本 |
|------|------|----------|---------|
| 循环依赖问题 | 阻碍功能开发 | 中 | v0.1.2 |
| 新增服务缺少测试 | 质量风险高 | 中 | v0.1.2 |
| 错误码体系未统一 | API 不规范 | 小 | v0.1.2 |

### 5.2 P1 - 重要但不紧急

| 问题 | 影响 | 修复工作量 | 目标版本 |
|------|------|----------|---------|
| Context 超时管理不一致 | 运维调优困难 | 小 | v0.1.3 |
| 硬编码配置值 | 配置不灵活 | 小 | v0.1.3 |
| 数据库初始化未测试 | 基础设施风险 | 中 | v0.1.3 |
| TODO 注释未处理 | 功能不完整 | 大 | v0.2.0 |

### 5.3 P2 - 优化改进

| 问题 | 影响 | 修复工作量 | 目标版本 |
|------|------|----------|---------|
| 批量操作无并发控制 | 性能不佳 | 中 | v0.2.0 |
| MongoDB Cursor 资源泄漏风险 | 资源管理 | 小 | v0.1.3 |

---

## 六、执行计划

### 6.1 第一阶段: 紧急问题修复 (v0.1.2)

#### 任务清单

**代码修复** (优先级 P0):
1. 解决包循环依赖问题
   - 分析依赖关系
   - 设计接口抽象方案
   - 重构代码结构
   - 验证编译通过

2. 为新增服务添加单元测试
   - 创建 `import_export_service_test.go`
   - 创建 `batch_operation_service_test.go`
   - Mock Repository 层依赖
   - 达到 80%+ 覆盖率

3. 统一错误码使用
   - 为批量操作定义错误码
   - 为导入导出定义错误码
   - 替换 fmt.Sprintf 错误消息
   - 更新错误码文档

**文档整合**:
1. 创建 docs/archive 子目录结构
2. 移动根目录历史文档到归档目录
3. 创建归档索引文档
4. 更新 README.md 引用

**脚本优化**:
1. 评估 mvp-test.sh 是否仍需要
2. 更新 scripts/README.md
3. 统一测试脚本清理策略

**预期成果**:
- 循环依赖问题解决，API Handler 可集成
- Service 层覆盖率提升至 60%+
- 根目录文档减少至 6-8 个
- 文档结构清晰，易于查找

### 6.2 第二阶段: 配置和超时优化 (v0.1.3)

#### 任务清单

**配置管理**:
1. 识别所有硬编码配置
2. 添加配置项到 config.yaml
3. 更新配置加载逻辑
4. 添加配置验证

**Context 管理**:
1. 设计统一的超时策略
2. 从配置文件读取超时配置
3. 更新所有 Context 使用
4. 添加超时监控日志

**测试补充**:
1. 创建数据库初始化测试
2. 测试索引创建逻辑
3. 测试连接失败场景

**预期成果**:
- 配置管理规范化
- 超时配置统一且可调整
- 数据库初始化有测试覆盖

### 6.3 第三阶段: 性能和功能完善 (v0.2.0)

#### 任务清单

**性能优化**:
1. 批量操作并发控制
2. Cursor 资源管理优化
3. 性能测试和基准

**功能完善**:
1. 处理所有 TODO 注释
2. 实现阶段三功能或明确废弃
3. 更新功能文档

**工程化**:
1. Makefile 命令优化
2. golangci-lint 规则完善
3. CI/CD 流程改进

**预期成果**:
- 批量操作性能提升 5-10 倍
- 无未处理的 TODO
- 工程化水平显著提升

---

## 七、质量保证措施

### 7.1 代码审查清单

每次代码修改前检查:

- [ ] 是否引入了新的 TODO 注释?
- [ ] 是否有硬编码的配置值?
- [ ] Context 是否正确管理和取消?
- [ ] 错误是否使用统一错误码?
- [ ] 资源 (cursor, connection) 是否正确关闭?
- [ ] 是否有足够的单元测试?
- [ ] 是否更新了相关文档?

### 7.2 测试要求

| 测试类型 | 覆盖率目标 | 检查点 |
|---------|-----------|--------|
| 单元测试 | 80%+ | 核心业务逻辑 |
| 集成测试 | 关键路径全覆盖 | 数据库操作 |
| E2E 测试 | 主要业务场景 | 完整流程 |

### 7.3 文档要求

| 文档类型 | 更新时机 | 检查点 |
|---------|---------|--------|
| 代码注释 | 代码编写时 | 导出函数必须有注释 |
| API 文档 | API 变更时 | 接口签名和示例 |
| CHANGELOG | 功能完成时 | 语义化版本规范 |
| README | 重大变更时 | 用户可见的变化 |

### 7.4 CI/CD 检查

**推送前本地检查**:
```bash
make fmt        # 代码格式化
make vet        # 静态检查
make lint       # Lint 检查
make test-unit  # 单元测试
```

**CI 流程检查**:
- 代码格式检查
- 静态分析
- 单元测试
- 集成测试
- 覆盖率检查 (70%+)
- 构建检查

---

## 八、风险评估

### 8.1 技术风险

| 风险 | 概率 | 影响 | 缓解措施 |
|------|------|------|---------|
| 重构引入新 bug | 中 | 高 | 充分的测试覆盖 |
| 配置变更影响现有部署 | 低 | 中 | 提供向后兼容的默认值 |
| 性能优化引入复杂性 | 中 | 中 | 分阶段实施，充分测试 |

### 8.2 进度风险

| 风险 | 概率 | 影响 | 缓解措施 |
|------|------|------|---------|
| 循环依赖重构耗时超预期 | 中 | 高 | 提前进行技术验证 |
| 测试编写工作量大 | 高 | 中 | 优先核心功能 |
| 文档整理琐碎耗时 | 低 | 低 | 使用脚本自动化 |

---

## 九、成功标准

### 9.1 代码质量指标

| 指标 | 当前值 | 目标值 | 测量方式 |
|------|--------|--------|---------|
| 单元测试覆盖率 | ~50% | 70%+ | go test -cover |
| Service 层覆盖率 | 45.6% | 75%+ | 分层测试 |
| golangci-lint 问题数 | 未知 | 0 | make lint |
| TODO 注释数量 | 9 | 0 或有明确计划 | grep TODO |

### 9.2 工程化指标

| 指标 | 当前值 | 目标值 | 测量方式 |
|------|--------|--------|---------|
| 根目录文档数量 | 16 | ≤ 8 | ls *.md |
| 硬编码配置数量 | 多处 | 0 | code review |
| 脚本功能重复 | 存在 | 无 | 人工评估 |
| 文档结构层级 | 混乱 | 清晰 | 目录树 |

### 9.3 验收标准

**必须满足** (P0):
- ✅ 所有 P0 问题已修复
- ✅ 新增服务有 80%+ 测试覆盖率
- ✅ 循环依赖问题已解决
- ✅ 错误码使用统一规范
- ✅ 根目录文档减少至 8 个以内
- ✅ golangci-lint 无报错

**应该满足** (P1):
- ⭕ 配置管理规范化
- ⭕ Context 超时统一管理
- ⭕ 文档归档结构清晰
- ⭕ 脚本功能无重复

**可选满足** (P2):
- 💡 批量操作并发优化
- 💡 性能基准测试
- 💡 所有 TODO 已处理

---

## 十、后续改进建议

### 10.1 中期改进 (v0.2.x)

1. **引入代码生成工具**
   - 使用 swag 生成 API 文档
   - 使用 mockgen 生成 Mock 对象
   - 减少手工维护工作

2. **完善监控体系**
   - 集成 Prometheus metrics
   - 添加性能监控
   - 实现分布式追踪

3. **自动化测试增强**
   - 添加混沌测试
   - 添加压力测试
   - 完善性能基准

### 10.2 长期改进 (v0.3.x+)

1. **架构演进**
   - 考虑插件化架构
   - 支持多种数据库后端
   - 实现配置热重载

2. **开发体验优化**
   - 提供开发环境容器
   - 集成 IDE 插件
   - 完善开发者文档

3. **社区建设**
   - 提供示例项目
   - 编写最佳实践指南
   - 建立贡献者社区

---

## 附录

### 附录 A: 错误码分配方案

**批量操作错误码** (6000-6999):

| 错误码 | 名称 | 消息 |
|-------|------|------|
| 6001 | ErrBatchOperationFailed | Batch operation failed |
| 6002 | ErrBatchPartialSuccess | Batch operation partially succeeded |
| 6003 | ErrBatchInvalidInput | Invalid batch operation input |

**导入导出错误码** (7000-7999):

| 错误码 | 名称 | 消息 |
|-------|------|------|
| 7001 | ErrImportDataInvalid | Import data is invalid |
| 7002 | ErrExportFailed | Export operation failed |
| 7003 | ErrImportConflict | Import data conflicts with existing data |
| 7004 | ErrUnsupportedVersion | Unsupported import data version |

### 附录 B: Context 超时配置表

| 操作类型 | 配置键 | 默认值 | 说明 |
|---------|--------|--------|------|
| 数据库查询 | `context_timeout.database_query` | 10s | 普通查询操作 |
| 数据库写入 | `context_timeout.database_write` | 15s | 写入操作 |
| 健康检查 | `context_timeout.health_check` | 2s | 健康检查 ping |
| 默认超时 | `context_timeout.default` | 30s | 其他操作 |
| 数据库关闭 | `performance.database_close_timeout` | 10s | 关闭连接 |

### 附录 C: 文档归档索引模板

建议创建 `docs/archive/INDEX.md`:

```markdown
# 归档文档索引

## 里程碑文档 (milestones/)
- MVP_MILESTONE_SUMMARY.md - MVP 版本里程碑总结
- MVP_RELEASE_CHECKLIST.md - MVP 版本发布检查清单

## Sprint 报告 (sprints/)
- SPRINT_01_COMPLETION_REPORT.md - Sprint 01 完成报告
- SPRINT_01_SUMMARY.md - Sprint 01 总结
- SPRINT_02_EXECUTION_SUMMARY.md - Sprint 02 执行总结
- SPRINT_02_SUMMARY.md - Sprint 02 总结

## 任务完成报告 (tasks/)
- TASK_COMPLETION_SUMMARY.md - 任务完成汇总

## 测试文档 (testing/)
- (测试相关历史文档)

## 废弃脚本 (scripts/)
- cleanup_docs.sh - 文档清理脚本 (已废弃)
- test-completion-report.sh - 测试报告生成脚本 (已废弃)
```
