# MockServer 版本发布清单

## 📋 发布前必检项目

本清单确保每次版本发布前完成所有必要的准备工作，适用于所有版本类型。

---

## ✅ 阶段一：代码和测试

### 1.1 代码完成度
- [ ] 所有计划功能已实现
- [ ] 代码已通过 Code Review
- [ ] 无已知的严重 Bug（P0/P1）
- [ ] 代码符合编码规范（通过 golangci-lint）

### 1.2 测试覆盖
- [ ] 单元测试覆盖率 ≥ 70%（总体）
- [ ] 核心模块覆盖率 ≥ 80%
  - [ ] executor 模块 ≥ 80%
  - [ ] service 模块 ≥ 80%
  - [ ] engine 模块 ≥ 80%
- [ ] 所有单元测试通过
- [ ] 集成测试通过
- [ ] 性能测试通过（如有）

### 1.3 构建验证
- [ ] 项目构建成功（`make build`）
- [ ] Docker 镜像构建成功
- [ ] 跨平台编译成功（Linux, macOS, Windows）

---

## ✅ 阶段二：文档更新

### 2.1 必需文档（根目录）
- [ ] `CHANGELOG.md` - 包含本版本完整变更记录
- [ ] `README.md` - 项目说明已更新版本信息
- [ ] `DEPLOYMENT.md` - 部署文档已更新（如有变化）
- [ ] `PROJECT_SUMMARY.md` - 项目总结已更新

### 2.2 架构文档
- [ ] `docs/ARCHITECTURE.md` - 架构文档已更新
- [ ] `docs/PROJECT_STRUCTURE.md` - 项目结构说明存在

### 2.3 发布文档
- [ ] `docs/releases/RELEASE_NOTES_v*.md` - 详细发布说明
- [ ] `docs/releases/RELEASE_CHECKLIST_v*.md` - 本清单文件
- [ ] `docs/releases/RELEASE_STATUS_v*.md` - 发布状态报告（可选）
- [ ] `docs/releases/RELEASE_SUMMARY_v*.md` - 发布总结（可选）

### 2.4 API 文档
- [ ] API 端点文档已更新（README.md 或独立文件）
- [ ] 新增 API 已添加使用示例
- [ ] 破坏性变更已明确标注

### 2.5 测试文档
- [ ] 测试报告已归档至 `docs/testing/reports/`
- [ ] 覆盖率数据已归档至 `docs/testing/coverage/`

---

## ✅ 阶段三：版本管理

### 3.1 版本号更新
- [ ] `internal/service/health.go` - Version 常量已更新
- [ ] `README.md` - 当前版本号已更新
- [ ] `CHANGELOG.md` - 版本号和日期已更新
- [ ] `package.json` - 前端版本号已更新（如适用）

### 3.2 版本号一致性
- [ ] 所有文档中的版本号一致
- [ ] Git 标签版本号与代码一致
- [ ] Docker 镜像标签与版本一致

---

## ✅ 阶段四：目录结构优化

### 4.1 测试文档归档
- [ ] 测试报告移至 `docs/testing/reports/`
- [ ] 覆盖率数据移至 `docs/testing/coverage/`
- [ ] 测试脚本移至 `docs/testing/scripts/`
- [ ] 测试计划移至 `docs/testing/plans/`（如有）

### 4.2 发布文档归档
- [ ] 发布说明移至 `docs/releases/`
- [ ] 发布清单移至 `docs/releases/`
- [ ] 版本验证脚本移至 `docs/releases/`
- [ ] 旧版本文档保留在 `docs/releases/`

### 4.3 脚本维护
- [ ] 核心脚本保留在 `scripts/`
- [ ] 临时脚本已归档或删除
- [ ] 版本特定脚本移至 `docs/releases/`
- [ ] 功能重复脚本已清理

### 4.4 根目录清理
- [ ] 临时文件已移除或归档
- [ ] 测试产出物已移至 `docs/testing/`
- [ ] 构建产物仅保留在 `bin/` 或 `.gitignore`
- [ ] 不必要的文件已删除

---

## ✅ 阶段五：质量验证

### 5.1 自动化验证
- [ ] 运行版本验证脚本（如 `verify_release_v*.sh`）
- [ ] 所有验证项通过
- [ ] 无编译错误或警告
- [ ] 无测试失败

### 5.2 手动验证
- [ ] 在本地环境启动服务成功
- [ ] 验证版本信息正确（`/api/v1/system/health`）
- [ ] 验证新增功能可用
- [ ] 验证 API 文档准确性

### 5.3 部署验证
- [ ] Docker Compose 启动成功
- [ ] 服务健康检查通过
- [ ] 数据库连接正常
- [ ] 前端访问正常

---

## ✅ 阶段六：发布准备

### 6.1 Git 操作
- [ ] 所有变更已提交
- [ ] 提交信息清晰明确
- [ ] 已创建版本标签（如 `v0.5.0`）
- [ ] 标签信息完整

### 6.2 远程仓库
- [ ] 代码已推送到主分支
- [ ] 标签已推送到远程
- [ ] GitHub/GitLab Release 已创建（如适用）
- [ ] Release Notes 已填写

### 6.3 构建产物
- [ ] Docker 镜像已构建
- [ ] Docker 镜像已推送到仓库（如适用）
- [ ] 二进制文件已编译（多平台）
- [ ] Release 附件已上传（如适用）

---

## ✅ 阶段七：发布后验证

### 7.1 部署验证
- [ ] 从 Docker Hub 拉取镜像成功
- [ ] 使用新版本镜像启动服务成功
- [ ] 验证版本号正确
- [ ] 验证新功能可用

### 7.2 文档验证
- [ ] README 在 GitHub 上显示正确
- [ ] CHANGELOG 格式正确
- [ ] Release Notes 链接有效
- [ ] API 文档可访问

### 7.3 通知和公告
- [ ] 更新项目主页（如有）
- [ ] 发布版本公告（如有社区）
- [ ] 通知相关团队或用户
- [ ] 更新文档网站（如有）

---

## 📊 发布检查总结

### 快速检查命令

```bash
# 1. 检查版本号一致性
grep -r "0.5.0" internal/service/health.go README.md CHANGELOG.md

# 2. 运行测试
make test

# 3. 检查覆盖率
make test-coverage

# 4. 构建项目
make build

# 5. 验证 Docker
docker-compose up -d
curl http://localhost:8080/api/v1/system/health

# 6. 运行验证脚本
./docs/releases/verify_release_v0.5.0.sh
```

### 必需文件检查

```bash
# 验证必需文件存在
ls -1 \
  CHANGELOG.md \
  README.md \
  docs/ARCHITECTURE.md \
  docs/PROJECT_STRUCTURE.md \
  docs/releases/RELEASE_NOTES_v*.md \
  docs/releases/RELEASE_CHECKLIST.md
```

### 目录结构检查

```bash
# 验证目录结构
tree -L 3 docs/
tree -L 2 scripts/
```

---

## 🎯 发布标准

### 所有版本必须满足
1. ✅ 计划功能完整可用
2. ✅ 测试覆盖率达标（70%+）
3. ✅ 文档完整准确
4. ✅ 无已知严重 Bug
5. ✅ 构建和部署成功

### 建议但非必需
- 📝 性能测试报告
- 📝 压力测试报告
- 📝 安全审计报告
- 📝 用户使用指南
- 📝 视频演示

---

## 📝 版本发布记录

### v0.5.0 (2025-01-17)
- ✅ 所有检查项通过
- ✅ 测试覆盖率 70.7%
- ✅ 核心模块覆盖率 80%+
- ✅ 文档完整更新
- ✅ 目录结构优化完成

---

## 🔄 持续改进

### 下一版本优化计划
- [ ] 增加性能测试环节
- [ ] 完善 API 文档生成流程
- [ ] 自动化版本验证脚本
- [ ] 添加安全扫描检查

---

**清单版本**: 1.0  
**适用版本**: v0.5.0+  
**最后更新**: 2025-01-17  
**维护者**: MockServer Team
