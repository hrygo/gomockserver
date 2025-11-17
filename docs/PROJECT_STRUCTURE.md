# MockServer 项目结构说明

本文档说明 MockServer 项目的目录结构和组织规范。

## 📁 根目录结构

```
gomockserver/
├── cmd/                    # 应用程序入口
├── internal/               # 内部代码
├── pkg/                    # 公共包
├── web/                    # Web 前端
├── tests/                  # 测试目录
├── docs/                   # 📚 文档目录（详见下方）
├── scripts/                # 🔧 核心脚本（详见下方）
├── docker/                 # Docker 配置
├── bin/                    # 编译产物
├── .github/                # GitHub Actions
├── CHANGELOG.md            # 变更日志
├── CONTRIBUTING.md         # 贡献指南
├── LICENSE                 # 开源协议
├── README.md               # 项目说明
├── DEPLOYMENT.md           # 部署文档
├── PROJECT_SUMMARY.md      # 项目总结
├── TECHNICAL_DEBT.md       # 技术债务
├── Makefile                # 构建工具
├── Dockerfile              # Docker 镜像
├── docker-compose.yml      # Docker 编排
└── go.mod                  # Go 依赖
```

## 📚 docs/ 目录结构（规范化组织）

```
docs/
├── ARCHITECTURE.md              # 📐 系统架构文档
├── PROJECT_STRUCTURE.md         # 📁 本文档
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
│   │   ├── TEST_REPORT.md
│   │   ├── coverage_improvement_report.md
│   │   └── test_results.txt
│   ├── coverage/                # 覆盖率数据
│   │   ├── coverage_summary.txt
│   │   └── *.html               # 覆盖率HTML报告
│   ├── scripts/                 # 测试脚本
│   │   └── test.sh
│   └── plans/                   # 测试计划
│
├── api/                         # 📡 API 文档（预留）
├── guides/                      # 📖 使用指南（预留）
└── archive/                     # 📦 历史文档归档
    └── (已废弃的文档)
```

## 🔧 scripts/ 目录结构（仅保留核心脚本）

```
scripts/
├── README.md                    # 脚本使用说明
├── run_unit_tests.sh           # ✅ 核心：单元测试执行
├── test-env.sh                 # ✅ 核心：环境测试
└── coverage/                   # 覆盖率生成脚本
    └── (保留核心覆盖率脚本)
```

### 脚本维护原则
- ✅ **保留**：核心执行脚本（run_unit_tests.sh, test-env.sh）
- 🗄️ **归档**：临时性脚本移至 `docs/testing/scripts/`
- 🗄️ **归档**：版本特定脚本移至 `docs/releases/`
- ❌ **清理**：功能重复或已废弃的脚本

## 📋 必需文件清单（开源项目标准）

### 根目录必需文件
- ✅ `CHANGELOG.md` - 变更日志
- ✅ `README.md` - 项目说明
- ✅ `Makefile` - 构建工具
- ✅ `go.mod` - 依赖管理

注：`CONTRIBUTING.md` 和 `LICENSE` 不常更新，不包含在版本发布检查中。

### 文档目录必需文件
- ✅ `docs/ARCHITECTURE.md` - 架构文档
- ✅ `docs/PROJECT_STRUCTURE.md` - 本文档
- ✅ `docs/releases/RELEASE_NOTES_v*.md` - 发布说明
- ✅ `docs/releases/RELEASE_CHECKLIST.md` - 版本发布清单

## 🎯 目录组织规范

### 1. 总结报告管理规则（v0.6.0+新增）
代码修改任务的执行总结必须归档至 `docs/summaries/` 目录：
- 实施总结 → `docs/summaries/{version}-{module}-implementation-summary.md`
- 测试总结 → `docs/summaries/{version}-test-report.md`
- 功能总结 → `docs/summaries/{feature}-summary.md`
- ❌ 禁止位置：`.qoder/quests/`（仅存放计划文档）

### 2. 发布文档管理规则
版本发布相关文档统一管理：
- 发布说明 → `docs/releases/RELEASE_NOTES_v*.md`（每版本必需，仅此一个文件）
- 发布清单 → `docs/releases/RELEASE_CHECKLIST.md`（所有版本共用）
- 验证脚本 → `docs/releases/verify_release_v*.sh`（可选）
- ❌ 不需要：`RELEASE_SUMMARY_v*.md`（已合并到 RELEASE_NOTES）

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
- ✅ 允许：工作计划、任务规划、设计规划、测试计划
- ❌ 禁止：总结报告、测试报告、实施报告、发布报告

## 📦 版本发布前的目录检查

每次发布新版本前，必须执行以下检查：

### 1. 目录结构检查
```bash
# 验证必需目录存在
docs/
docs/summaries/              # v0.6.0+新增
docs/releases/
docs/testing/{reports,coverage,scripts,plans}/
docs/archive/
scripts/
.qoder/quests/               # 仅存放计划文档
```

### 2. 文档完整性检查
```bash
# 验证必需文件
CHANGELOG.md
README.md
docs/ARCHITECTURE.md
docs/PROJECT_STRUCTURE.md
docs/releases/RELEASE_NOTES_v*.md
docs/releases/RELEASE_CHECKLIST.md
```

### 3. 清理临时文件
```bash
# 移动根目录的临时文件
*.txt (测试结果) → docs/testing/reports/
*_report.md (测试报告) → docs/testing/reports/
coverage*.* (覆盖率) → docs/testing/coverage/

# 移动 .qoder/quests/ 中的总结报告
*-summary.md (总结报告) → docs/summaries/
*-report.md (测试报告) → docs/summaries/
```

### 4. 脚本维护
```bash
# 检查 scripts/ 目录
- 仅保留核心执行脚本
- 版本脚本移至 docs/releases/
- 测试脚本移至 docs/testing/scripts/
```

## 🔄 持续维护建议

### 每次开发周期结束后
1. 整理总结报告到 `docs/summaries/`（从 `.qoder/quests/` 移除）
2. 整理测试产出物到 `docs/testing/`
3. 归档旧版本文档到 `docs/archive/`
4. 清理根目录临时文件
5. 清理 `.qoder/quests/` 中的非计划文档
6. 更新本文档（如有结构变更）

### 每次版本发布前
1. 执行目录结构检查
2. 验证文档完整性
3. 清理和归档临时文件
4. 更新 `CHANGELOG.md`

### 定期清理（每季度）
1. 检查 `docs/archive/` 是否需要进一步整理
2. 清理 `scripts/` 中的废弃脚本
3. 验证所有文档链接有效性

## 📊 目录结构变更历史

### v0.5.0 (2025-01-17)
- ✅ 创建 `docs/testing/` 目录结构
- ✅ 创建 `docs/releases/` 目录
- ✅ 移动测试报告至 `docs/testing/reports/`
- ✅ 移动覆盖率数据至 `docs/testing/coverage/`
- ✅ 移动发布文档至 `docs/releases/`
- ✅ 移动版本脚本至 `docs/releases/`
- ✅ 移动测试脚本至 `docs/testing/scripts/`
- ✅ 创建本结构说明文档

### v0.6.0 (2025-11-17)
- ✅ 创建 `docs/summaries/` 目录（存放任务总结报告）
- ✅ 移动总结报告从 `.qoder/quests/` 至 `docs/summaries/`
- ✅ 明确 `.qoder/quests/` 仅存放计划文档
- ✅ 规范发布文档：仅需 `RELEASE_NOTES_v*.md`（不需要 SUMMARY）
- ✅ 更新目录组织规范和检查清单

### 后续优化计划
- 📝 创建 API 文档模板
- 📝 创建使用指南模板
- 📝 完善测试计划文档

---

**文档版本**: 1.1  
**最后更新**: 2025-11-17  
**维护者**: MockServer Team
