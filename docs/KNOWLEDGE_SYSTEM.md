# MockServer 知识体系总结

## 📚 文档概述

本文档整理了 MockServer 项目的完整知识体系，包括项目配置、目录规范、文档管理、工作流程等关键信息。

---

## 1️⃣ 项目配置信息

### 技术栈

#### 后端技术栈
- **编程语言**: Go 1.24.0
- **Web 框架**: Gin
- **数据库**: MongoDB 6.0+
- **配置管理**: Viper
- **日志系统**: Zap
- **容器化**: Docker
- **WebSocket**: gorilla/websocket v1.5.3
- **JavaScript 引擎**: goja

#### 前端技术栈
- **框架**: React 18.3.1
- **语言**: TypeScript 5.3.3
- **构建工具**: Vite 5.1.0
- **UI 组件库**: Ant Design 5.14.0
- **路由**: React Router 6.22.0
- **状态管理**: Zustand 4.5.0
- **数据请求**: TanStack Query 5.20.0
- **HTTP 客户端**: Axios 1.6.7
- **图表**: ECharts 5.6.0
- **代码编辑器**: Monaco Editor 0.46.0

### 当前版本
- **最新版本**: v0.6.0 (Enterprise Foundation)
- **发布日期**: 2025-11-17
- **测试覆盖率**: 总体 69.3%+，核心模块 80%+

### 构建工具
- **Go**: Makefile
- **前端**: Vite 5.1.0, package.json

---

## 2️⃣ 目录结构规范

### 核心目录结构

```
gomockserver/
├── cmd/                    # 应用程序入口
├── internal/               # 内部代码
├── pkg/                    # 公共包
├── web/                    # Web 前端
├── tests/                  # 测试目录
├── docs/                   # 📚 文档目录
│   ├── summaries/          # 任务总结报告（v0.6.0+）
│   ├── releases/           # 版本发布文档
│   ├── testing/            # 测试文档
│   ├── api/                # API 文档（预留）
│   ├── guides/             # 使用指南（预留）
│   └── archive/            # 历史文档归档
├── scripts/                # 🔧 核心脚本
├── docker/                 # Docker 配置
├── bin/                    # 编译产物
├── .github/                # GitHub Actions
├── .qoder/                 # 工作计划和规划
│   └── quests/             # 仅存放计划文档
└── (根目录必需文件)
```

### docs/ 目录详细结构

```
docs/
├── ARCHITECTURE.md              # 系统架构文档
├── PROJECT_STRUCTURE.md         # 项目结构说明
├── KNOWLEDGE_SYSTEM.md          # 本文档
│
├── summaries/                   # 📊 任务总结报告（v0.6.0+新增）
│   ├── v0.6.0-backend-implementation-summary.md
│   ├── v0.6.0-test-report.md
│   └── (其他代码修改任务的总结报告)
│
├── releases/                    # 🎯 版本发布文档
│   ├── RELEASE_NOTES_v*.md      # 发布说明（每版本必需）
│   ├── RELEASE_CHECKLIST.md     # 发布清单（所有版本共用）
│   └── verify_release_v*.sh     # 版本验证脚本（可选）
│
├── testing/                     # 🧪 测试相关文档
│   ├── reports/                 # 测试报告
│   ├── coverage/                # 覆盖率数据
│   ├── scripts/                 # 测试脚本
│   └── plans/                   # 测试计划
│
├── api/                         # 📡 API 文档（预留）
├── guides/                      # 📖 使用指南（预留）
└── archive/                     # 📦 历史文档归档
```

### .qoder/ 目录规范

**.qoder/quests/ 用途定位**（v0.6.0+明确）：
- ✅ **允许存放**：工作计划、任务规划、设计规划、测试计划
- ❌ **禁止存放**：总结报告、测试报告、实施报告、发布报告

**目录结构**：
```
.qoder/
├── quests/                      # 工作计划和任务规划
│   ├── architecture/            # 架构规划
│   ├── improvement/             # 改进计划
│   ├── testing/                 # 测试计划（计划，非报告）
│   └── (各种 planning 文档)
└── archive/                     # 历史规划归档
```

---

## 3️⃣ 文档管理规范

### 总结报告生成和存放规则

#### 存放位置规范
- ❌ **禁止**：不要将执行过程的总结报告放到 `.qoder/quests` 目录
- ✅ **正确位置**：`docs/summaries/` 目录
- ✅ **发布报告**：`docs/releases/` 目录（仅限版本发布）

#### 生成条件判断
| 任务类型 | 是否生成 | 存放位置 | 文件命名示例 |
|---------|---------|---------|-------------|
| 修改代码的任务 | ✅ 是 | docs/summaries/ | v0.6.0-backend-implementation-summary.md |
| 版本发布任务 | ✅ 是（必须） | docs/releases/ | RELEASE_NOTES_v0.6.0.md |
| 短任务（查询/分析/简单配置） | ❌ 否 | - | - |

#### 发布报告特殊规则
- 无论任务长短，**版本发布必须生成发布报告**
- 文件命名：`RELEASE_NOTES_v*.md`（**仅此一个文件，不需要额外的 SUMMARY**）
- 存放位置：`docs/releases/`
- 报告内容：版本信息、核心变更、质量指标、升级指南、已知问题等

#### 判断逻辑流程图
```
任务开始
    ↓
是否为版本发布？
    ├─ 是 → 生成 docs/releases/RELEASE_NOTES_v*.md
    └─ 否 → 是否修改了代码？
            ├─ 是 → 生成 docs/summaries/{task}-summary.md
            └─ 否 → 不生成总结报告
```

### 文档命名规范

| 文档类型 | 位置 | 命名规范 | 示例 |
|---------|------|---------|------|
| 实施总结 | docs/summaries/ | {version}-{module}-implementation-summary.md | v0.6.0-backend-implementation-summary.md |
| 测试总结 | docs/summaries/ | {version}-test-report.md | v0.6.0-test-report.md |
| 功能总结 | docs/summaries/ | {feature}-summary.md | cors-middleware-summary.md |
| 发布说明 | docs/releases/ | RELEASE_NOTES_v{version}.md | RELEASE_NOTES_v0.6.0.md |
| 发布清单 | docs/releases/ | RELEASE_CHECKLIST.md | RELEASE_CHECKLIST.md（所有版本共用） |
| 验证脚本 | docs/releases/ | verify_release_v{version}.sh | verify_release_v0.6.0.sh |

---

## 4️⃣ 版本发布流程

### 发布前必须执行的默认动作

执行顺序：**目录检查 → 结构优化 → 文档更新 → 生成发布报告 → 版本发布**

#### 步骤 1：检查工程目录结构
- 验证是否符合标准开源项目结构
- 检查必需文件完整性：
  - CHANGELOG.md
  - README.md
  - docs/ARCHITECTURE.md
  - docs/PROJECT_STRUCTURE.md
- 验证 `docs/summaries/` 目录存在（v0.6.0+）
- 注意：不检查 `CONTRIBUTING.md` 和 `LICENSE`（这些文件不常更新）

#### 步骤 2：优化产出物目录结构
- **总结报告归档**：`docs/summaries/`（实施总结、任务总结等）
- **测试文档归档**：`docs/testing/` (reports/coverage/scripts/plans/)
- **发布文档归档**：`docs/releases/`（仅 `RELEASE_NOTES_v*.md`）
- **脚本维护**：核心脚本保留在 `scripts/`，临时/废弃脚本归档至 `docs/archive/`
- **历史文档归档**：`docs/archive/`
- **.qoder/quests/ 清理**：移除所有总结报告，仅保留计划文档

#### 步骤 3：更新项目文档
- **必需文件**：
  - `CHANGELOG.md`（版本号、日期）
  - `README.md`（功能列表、版本信息）
  - `PROJECT_SUMMARY.md`（项目状态）
  - `docs/ARCHITECTURE.md`（架构更新）
  - `docs/PROJECT_STRUCTURE.md`（结构说明）
- **发布文档**：
  - `docs/releases/RELEASE_NOTES_v*.md`（发布说明，**仅此一个文件**）
  - `docs/releases/RELEASE_CHECKLIST.md`（发布清单，所有版本共用）
- **版本文档**：版本号一致性、README.md 功能列表、API 文档、项目状态

#### 步骤 4：生成发布报告
- 版本发布必须生成发布报告（无论任务大小）
- 文件命名：`docs/releases/RELEASE_NOTES_v*.md`
- 报告内容：
  - 版本信息（版本号、日期、代号、类型）
  - 版本概述（核心价值）
  - 新增功能详解
  - 技术指标（测试覆盖率、性能指标）
  - 升级指南
  - 已知问题
  - 文档更新列表

#### 步骤 5：版本发布
```bash
# 1. Git 提交
git add -A
git commit -m "release: v{version} - {code_name}

详细的提交信息..."

# 2. 创建 Git 标签
git tag -a v{version} -m "Release v{version} - {code_name}

核心特性:
- 特性1
- 特性2

测试覆盖率: XX%
发布日期: YYYY-MM-DD"

# 3. 推送到远程仓库
git push origin master
git push origin v{version}

# 4. 构建验证
make build
```

### 发布检查清单

详见 `docs/releases/RELEASE_CHECKLIST.md`

---

## 5️⃣ 目录组织规范

### 1. 总结报告管理规则（v0.6.0+新增）
代码修改任务的执行总结必须归档至 `docs/summaries/` 目录：
- 实施总结 → `docs/summaries/{version}-{module}-implementation-summary.md`
- 测试总结 → `docs/summaries/{version}-test-report.md`
- 功能总结 → `docs/summaries/{feature}-summary.md`
- ❌ **禁止位置**：`.qoder/quests/`（仅存放计划文档）

### 2. 发布文档管理规则
版本发布相关文档统一管理：
- 发布说明 → `docs/releases/RELEASE_NOTES_v*.md`（每版本必需，**仅此一个文件**）
- 发布清单 → `docs/releases/RELEASE_CHECKLIST.md`（所有版本共用）
- 验证脚本 → `docs/releases/verify_release_v*.sh`（可选）
- ❌ **不需要**：`RELEASE_SUMMARY_v*.md`（已合并到 RELEASE_NOTES）

### 3. 测试文档归档规则
所有测试相关产出物必须归档至 `docs/testing/` 目录：
- 测试报告 → `docs/testing/reports/`
- 覆盖率数据 → `docs/testing/coverage/`
- 测试脚本 → `docs/testing/scripts/`
- 测试计划 → `docs/testing/plans/`

### 4. 脚本管理规则
- 核心脚本保留在 `scripts/` 根目录
- 临时脚本归档至相应功能目录
- 版本脚本归档至 `docs/releases/`
- 定期清理重复或废弃脚本

### 5. 历史文档归档规则
- 已废弃文档 → `docs/archive/`
- 旧版本设计文档 → `docs/archive/`
- 不再维护的指南 → `docs/archive/`

### 6. .qoder/quests/ 目录规则（v0.6.0+明确）
仅存放工作计划和任务规划文档：
- ✅ **允许**：工作计划、任务规划、设计规划、测试计划
- ❌ **禁止**：总结报告、测试报告、实施报告、发布报告

---

## 6️⃣ 持续维护建议

### 每次开发周期结束后
1. 整理总结报告到 `docs/summaries/`（从 `.qoder/quests/` 移除）
2. 整理测试产出物到 `docs/testing/`
3. 归档旧版本文档到 `docs/archive/`
4. 清理根目录临时文件
5. 清理 `.qoder/quests/` 中的非计划文档
6. 更新相关文档（如有结构变更）

### 每次版本发布前
1. 执行目录结构检查
2. 验证文档完整性
3. 清理和归档临时文件
4. 更新 `CHANGELOG.md`
5. 生成 `RELEASE_NOTES_v*.md`

### 定期清理（每季度）
1. 检查 `docs/archive/` 是否需要进一步整理
2. 清理 `scripts/` 中的废弃脚本
3. 验证所有文档链接有效性
4. 更新知识体系文档

---

## 7️⃣ 版本历史和演进

### v0.5.0 (2025-01-17)
- ✅ 创建 `docs/testing/` 目录结构
- ✅ 创建 `docs/releases/` 目录
- ✅ 移动测试报告至 `docs/testing/reports/`
- ✅ 移动覆盖率数据至 `docs/testing/coverage/`
- ✅ 移动发布文档至 `docs/releases/`

### v0.6.0 (2025-11-17)
- ✅ 创建 `docs/summaries/` 目录（存放任务总结报告）
- ✅ 移动总结报告从 `.qoder/quests/` 至 `docs/summaries/`
- ✅ 明确 `.qoder/quests/` 仅存放计划文档
- ✅ 规范发布文档：仅需 `RELEASE_NOTES_v*.md`（不需要 SUMMARY）
- ✅ 更新目录组织规范和检查清单
- ✅ 创建知识体系总结文档

---

## 8️⃣ 快速参考

### 文件应该放在哪里？

| 文件类型 | 正确位置 | 错误位置 |
|---------|---------|---------|
| 工作计划 | `.qoder/quests/` | `docs/summaries/` |
| 任务规划 | `.qoder/quests/` | `docs/summaries/` |
| 设计规划 | `.qoder/quests/architecture/` | `docs/archive/` |
| 实施总结 | `docs/summaries/` | `.qoder/quests/` |
| 测试报告 | `docs/summaries/` | `.qoder/quests/` |
| 发布说明 | `docs/releases/RELEASE_NOTES_v*.md` | 项目根目录 |
| 测试数据 | `docs/testing/reports/` | 项目根目录 |
| 覆盖率数据 | `docs/testing/coverage/` | 项目根目录 |
| 废弃文档 | `docs/archive/` | 直接删除 |

### 常用命令

```bash
# 检查目录结构
tree -L 2 docs/
tree -L 2 .qoder/

# 验证必需文件
ls -1 CHANGELOG.md README.md docs/ARCHITECTURE.md docs/PROJECT_STRUCTURE.md

# 清理临时文件（移至正确位置）
mv .qoder/quests/*-summary.md docs/summaries/
mv .qoder/quests/*-report.md docs/summaries/

# 构建和测试
make build
make test
make test-coverage

# 版本发布
git tag -l "v0.*"
git push origin v{version}
```

---

## 📝 文档维护

**文档版本**: 1.0  
**创建日期**: 2025-11-17  
**最后更新**: 2025-11-17  
**维护者**: MockServer Team

**相关文档**:
- [项目结构说明](PROJECT_STRUCTURE.md)
- [系统架构文档](ARCHITECTURE.md)
- [发布检查清单](releases/RELEASE_CHECKLIST.md)
- [项目总结](../PROJECT_SUMMARY.md)
