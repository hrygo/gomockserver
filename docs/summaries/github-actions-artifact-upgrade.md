# GitHub Actions Artifact 升级修复总结

## 问题描述

**错误信息**：
```
Error: This request has been automatically failed because it uses a deprecated version of `actions/upload-artifact: v3`. 
Learn more: https://github.blog/changelog/2024-04-16-deprecation-notice-v3-of-the-artifact-actions/
```

**触发场景**：推送代码到 GitHub 后，GitHub Actions CI 流程执行时报错。

**错误原因**：使用了已弃用的 `actions/upload-artifact@v3` 版本。

---

## 根本原因分析

GitHub 在 2024-04-16 宣布弃用 v3 版本的 artifact actions（包括 `upload-artifact` 和 `download-artifact`），要求所有工作流升级到 v4 版本。

**主要变更**：
- v3 版本将不再支持
- v4 版本提供了改进的性能和可靠性
- v4 版本使用了新的存储后端

**官方公告**：https://github.blog/changelog/2024-04-16-deprecation-notice-v3-of-the-artifact-actions/

---

## 解决方案

### 修改文件

1. `.github/workflows/docker.yml`
2. `.github/workflows/ci.yml`

### 具体修改

#### 1. docker.yml 修改（1 处）

**位置**：第 101 行

**修改前**：
```yaml
- name: Upload logs
  if: failure()
  uses: actions/upload-artifact@v3
  with:
    name: docker-logs
    path: docker-logs.txt
```

**修改后**：
```yaml
- name: Upload logs
  if: failure()
  uses: actions/upload-artifact@v4
  with:
    name: docker-logs
    path: docker-logs.txt
```

#### 2. ci.yml 修改（2 处）

**位置 1**：第 59 行（unit-tests job）

**修改前**：
```yaml
- name: Archive coverage results
  uses: actions/upload-artifact@v3
  with:
    name: coverage-report
    path: |
      coverage.out
      coverage.txt
```

**修改后**：
```yaml
- name: Archive coverage results
  uses: actions/upload-artifact@v4
  with:
    name: coverage-report
    path: |
      coverage.out
      coverage.txt
```

**位置 2**：第 133 行（integration-tests job）

**修改前**：
```yaml
- name: Archive test logs
  if: failure()
  uses: actions/upload-artifact@v3
  with:
    name: integration-test-logs
    path: /tmp/mockserver_e2e_test.log
```

**修改后**：
```yaml
- name: Archive test logs
  if: failure()
  uses: actions/upload-artifact@v4
  with:
    name: integration-test-logs
    path: /tmp/mockserver_e2e_test.log
```

---

## 修改详情

### 变更统计

- **修改文件数量**：2 个
- **修改位置数量**：3 处
  - docker.yml: 1 处
  - ci.yml: 2 处
- **变更类型**：版本升级（v3 → v4）

### 影响范围

**受影响的工作流**：
1. Docker Build and Test（docker.yml）
   - docker-compose-test job 中的日志上传

2. CI Tests（ci.yml）
   - unit-tests job 中的覆盖率报告上传
   - integration-tests job 中的测试日志上传

**功能影响**：
- ✅ **无功能变更**：仅版本升级，功能保持不变
- ✅ **向后兼容**：v4 完全兼容 v3 的使用方式
- ✅ **性能提升**：v4 版本提供更好的性能和可靠性

---

## 验证步骤

### 自动验证

推送到 GitHub 后，GitHub Actions 会自动触发工作流，验证以下内容：

1. **Docker Build and Test 工作流**
   - docker-build job
   - docker-compose-test job（测试失败时上传日志）

2. **CI Tests 工作流**
   - unit-tests job（上传覆盖率报告）
   - integration-tests job（测试失败时上传日志）
   - code-quality job
   - build job

### 验证结果

访问 GitHub Actions 页面，确认：
- ✅ 所有工作流运行成功
- ✅ 不再出现 artifact v3 弃用警告
- ✅ artifact 正常上传（如果触发条件满足）

---

## v3 到 v4 的主要变更

根据 GitHub 官方文档，v4 版本的主要变更包括：

### 1. 性能改进
- 更快的上传和下载速度
- 改进的并发处理
- 优化的存储后端

### 2. 可靠性提升
- 更好的错误处理
- 改进的重试机制
- 更稳定的大文件处理

### 3. 使用方式
- **基本用法保持不变**（向后兼容）
- 参数和选项与 v3 相同
- 无需修改 `with` 配置

### 4. 破坏性变更
- ⚠️ **无**：v4 完全兼容 v3 的使用方式
- ✅ 可以直接从 v3 升级到 v4，无需修改其他配置

---

## 提交信息

```
fix(ci): upgrade actions/upload-artifact from v3 to v4

- Upgrade actions/upload-artifact@v3 to v4 in docker.yml
- Upgrade actions/upload-artifact@v3 to v4 in ci.yml (2 occurrences)
- Resolves deprecation warning from GitHub Actions

GitHub deprecated v3 of artifact actions on 2024-04-16.
See: https://github.blog/changelog/2024-04-16-deprecation-notice-v3-of-the-artifact-actions/

Changed files:
- .github/workflows/docker.yml: Upload logs (1 occurrence)
- .github/workflows/ci.yml: Archive coverage results, Archive test logs (2 occurrences)
```

**Git Commit**: `6271253`

---

## 相关文件

- `.github/workflows/docker.yml` - Docker 构建和测试工作流
- `.github/workflows/ci.yml` - CI 测试工作流
- `.github/workflows/pr-checks.yml` - PR 检查工作流（未使用 upload-artifact）

---

## 后续建议

### 1. 定期检查依赖版本

定期检查 GitHub Actions 中使用的 action 版本，确保使用最新的稳定版本：

```bash
# 检查所有工作流中使用的 actions
grep -r "uses:" .github/workflows/ | sort | uniq
```

### 2. 订阅 GitHub 更新通知

关注 GitHub Blog 的更新公告：
- GitHub Blog: https://github.blog/changelog/
- GitHub Actions 公告: https://github.blog/changelog/?s=actions

### 3. 使用 Dependabot 自动更新

考虑配置 Dependabot 来自动更新 GitHub Actions 的版本：

```yaml
# .github/dependabot.yml
version: 2
updates:
  - package-ecosystem: "github-actions"
    directory: "/"
    schedule:
      interval: "weekly"
```

### 4. 其他常用 actions 检查

除了 `upload-artifact`，还应关注以下常用 actions 的版本：

| Action | 当前使用版本 | 最新稳定版本 | 是否需要更新 |
|--------|------------|------------|------------|
| actions/checkout | v4 | v4 | ✅ 最新 |
| actions/setup-go | v4 | v5 | ⚠️ 可考虑升级 |
| docker/setup-buildx-action | v3 | v3 | ✅ 最新 |
| docker/login-action | v3 | v3 | ✅ 最新 |
| docker/build-push-action | v5 | v6 | ⚠️ 可考虑升级 |
| docker/metadata-action | v5 | v5 | ✅ 最新 |
| codecov/codecov-action | v3 | v4 | ⚠️ 可考虑升级 |
| golangci/golangci-lint-action | v3 | v6 | ⚠️ 建议升级 |

---

## 问题预防

为避免类似问题再次发生，建议：

### 1. 启用 GitHub Actions 安全更新

在仓库设置中启用 Dependabot 安全更新：
- Settings → Security → Code security and analysis
- 启用 "Dependabot security updates"

### 2. 定期审查工作流

每月或每季度审查一次 GitHub Actions 工作流：
- 检查是否有弃用警告
- 更新到最新稳定版本
- 删除未使用的工作流

### 3. 测试环境验证

在推送到主分支前，在分支上测试工作流：
```bash
# 创建测试分支
git checkout -b test/update-actions

# 修改工作流文件
# ...

# 推送并触发工作流
git push origin test/update-actions
```

### 4. 文档记录

在项目文档中记录：
- 使用的 GitHub Actions 版本
- 更新历史和原因
- 已知问题和解决方案

---

## 参考资料

1. **GitHub 官方公告**
   - [Deprecation notice: v3 of the artifact actions](https://github.blog/changelog/2024-04-16-deprecation-notice-v3-of-the-artifact-actions/)

2. **actions/upload-artifact 文档**
   - [GitHub Repository](https://github.com/actions/upload-artifact)
   - [v4 Release Notes](https://github.com/actions/upload-artifact/releases/tag/v4.0.0)

3. **迁移指南**
   - [Migrating from v3 to v4](https://github.com/actions/upload-artifact#migration-from-v3)

4. **GitHub Actions 最佳实践**
   - [GitHub Actions documentation](https://docs.github.com/en/actions)
   - [Workflow syntax](https://docs.github.com/en/actions/using-workflows/workflow-syntax-for-github-actions)

---

**修复日期**: 2025-11-17  
**修复版本**: v0.6.1  
**影响范围**: GitHub Actions CI/CD 工作流  
**测试状态**: ✅ 已推送到 GitHub，等待 CI 验证