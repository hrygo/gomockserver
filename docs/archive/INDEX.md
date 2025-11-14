# 归档文档索引

本目录包含项目历史文档和已完成阶段的记录，按类型分类归档。

## 📂 目录结构

```
archive/
├── milestones/         # 里程碑文档
├── sprints/            # Sprint 报告
├── tasks/              # 任务完成报告
├── releases/           # 历史发布记录
├── testing/            # 测试相关历史文档
├── scripts/            # 废弃脚本
└── reports/            # 测试报告
```

---

## 🏆 里程碑文档 (milestones/)

| 文档 | 说明 | 日期 |
|------|------|------|
| `MVP_MILESTONE_SUMMARY.md` | MVP 版本里程碑总结 | 2025-01-15 |
| `MVP_RELEASE_CHECKLIST.md` | MVP 版本发布检查清单 | 2025-01-15 |

MVP（最小可行产品）版本的完整记录，包含功能清单、测试结果、发布流程等。

---

## 🏃 Sprint 报告 (sprints/)

| Sprint | 文档 | 说明 | 日期 |
|--------|------|------|------|
| Sprint 01 | `SPRINT_01_COMPLETION_REPORT.md` | Sprint 01 完成报告 | 2025-01-15 |
| Sprint 01 | `SPRINT_01_SUMMARY.md` | Sprint 01 详细总结 | 2025-01-15 |
| Sprint 02 | `SPRINT_02_EXECUTION_SUMMARY.md` | Sprint 02 执行总结 | 2025-01-20 |
| Sprint 02 | `SPRINT_02_SUMMARY.md` | Sprint 02 详细总结 | 2025-01-20 |

每个 Sprint 的执行记录，包含任务完成情况、技术亮点、遇到的问题和解决方案。

### Sprint 内容概览

- **Sprint 01 (v0.1.1)**: 代码质量提升
  - Repository 层测试覆盖率提升 (44.4% → 80%+)
  - 建立统一错误码体系 (42 个错误码)
  - 健康检查增强和请求追踪
  - Makefile 工程化增强

- **Sprint 02 (v0.1.2)**: 功能增强
  - 导入导出功能实现
  - 批量操作服务实现
  - 规则克隆功能

---

## ✅ 任务完成报告 (tasks/)

| 文档 | 说明 | 日期 |
|------|------|------|
| `TASK_COMPLETION_SUMMARY.md` | Sprint 02 任务完成汇总 | 2025-01-20 |

任务级别的详细完成报告，包含代码统计、待完成工作、经验教训等。

---

## 📦 历史发布记录 (releases/)

| 版本 | 文档 | 发布日期 |
|------|------|---------|
| v0.1.0 | `RELEASE_NOTES_v0.1.0.md` | 2025-01-10 |

历史版本的发布说明，包含新功能、改进、修复的 Bug 等。

---

## 🧪 测试文档 (testing/)

测试相关的历史文档，包含：
- 测试方案设计
- 测试执行报告
- 覆盖率分析
- 测试环境配置

---

## 📜 废弃脚本 (scripts/)

| 脚本 | 说明 | 废弃原因 |
|------|------|---------|
| `cleanup_docs.sh` | 文档清理脚本 | 功能已整合到项目管理流程 |
| `test-completion-report.sh` | 测试报告生成脚本 | 由 run_unit_tests.sh 替代 |

这些脚本已不再使用，保留用于参考。

---

## 📊 测试报告 (reports/)

自动生成的测试报告，包含单元测试结果、覆盖率数据等。定期清理旧报告。

---

## 🔍 查找指南

### 查找 MVP 相关信息
```bash
cd docs/archive/milestones/
ls -l
```

### 查找特定 Sprint 报告
```bash
cd docs/archive/sprints/
grep -r "关键词" .
```

### 查看历史发布说明
```bash
cd docs/archive/releases/
cat RELEASE_NOTES_v0.1.0.md
```

---

## 📝 归档说明

### 归档原则
1. **及时归档**: Sprint 完成后立即归档相关文档
2. **分类明确**: 按文档类型归档到对应目录
3. **保留价值**: 仅归档有参考价值的文档
4. **定期清理**: 测试报告等临时文档定期清理

### 文档命名规范
- 里程碑: `{PROJECT}_MILESTONE_SUMMARY.md`
- Sprint: `SPRINT_{NN}_{TYPE}.md`
- 发布: `RELEASE_NOTES_v{X.Y.Z}.md`
- 任务: `TASK_COMPLETION_SUMMARY.md`

---

## 🔗 相关链接

- [项目主文档](../../README.md)
- [变更日志](../../CHANGELOG.md)
- [贡献指南](../../CONTRIBUTING.md)
- [架构文档](../ARCHITECTURE.md)

---

**最后更新**: 2025-01-21  
**维护者**: 项目团队
