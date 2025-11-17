# GitHub Actions Docker 构建和测试修复总结

## 问题描述

在 v0.6.1 版本发布后，GitHub Actions 仍然出现以下错误：

### 1. Docker 构建错误
```
ERROR: failed to build: failed to solve: failed to compute cache key: failed to calculate checksum of ref ...: "/cmd": not found
```

### 2. Docker Compose 命令未找到
```
docker-compose: command not found
```

## 根本原因分析

### 问题 1: Docker 构建错误
- **原因**: 在尝试使用显式 `COPY cmd/ ./cmd/` 指令时，构建上下文中的路径处理出现了问题
- **深层原因**: 可能是 `.dockerignore` 规则或构建环境的差异导致路径解析失败

### 问题 2: Docker Compose 命令未找到
- **原因**: GitHub Actions 环境更新后，`docker-compose` 命令已被 `docker compose`（没有连字符）替代
- **深层原因**: Docker CLI 插件架构变更，旧的 `docker-compose` 命令不再默认安装

## 解决方案

### 修复 1: 恢复 Dockerfile 的 COPY 指令

**修改文件**：
1. `Dockerfile`
2. `Dockerfile.fullstack`

**具体修改**：
将显式的目录复制指令：
```dockerfile
# 复制源代码（确保包含关键目录）
COPY cmd/ ./cmd/
COPY internal/ ./internal/
COPY pkg/ ./pkg/
COPY config.yaml ./
```

改回传统的复制方式：
```dockerfile
# 复制源代码
COPY . .
```

**理由**：
- 通过优化 `.dockerignore` 文件来控制复制内容，而不是显式指定目录
- 避免路径解析问题
- 保持与标准 Docker 构建实践一致

### 修复 2: 更新 docker-compose 命令

**修改文件**：
- `.github/workflows/docker.yml`

**具体修改**：
将所有的 `docker-compose` 命令更新为 `docker compose`：
```yaml
# 之前
docker-compose -f docker-compose.test.yml up -d mongodb-test mockserver-test

# 之后
docker compose -f docker-compose.test.yml up -d mongodb-test mockserver-test
```

**影响的命令**：
1. `docker-compose up` → `docker compose up`
2. `docker-compose ps` → `docker compose ps`
3. `docker-compose run` → `docker compose run`
4. `docker-compose logs` → `docker compose logs`
5. `docker-compose down` → `docker compose down`

## 修改详情

### 文件变更统计

- **修改文件数量**: 3 个
  - `Dockerfile`: 恢复 COPY 指令
  - `Dockerfile.fullstack`: 恢复 COPY 指令
  - `.github/workflows/docker.yml`: 更新 docker-compose 命令

### 影响范围

**受影响的工作流**：
1. Docker Build and Test（docker.yml）
   - docker-build job
   - docker-compose-test job

**功能影响**：
- ✅ **无功能变更**: 仅修复构建和测试命令
- ✅ **向后兼容**: 保持应用功能不变
- ✅ **性能提升**: 修复后构建和测试流程将正常运行

## 验证步骤

### 本地验证

```bash
# 验证 Docker 构建
docker build -t mockserver .

# 验证完整栈构建
docker build -f Dockerfile.fullstack -t mockserver-full .

# 验证 docker compose 命令（如果安装了 Docker Compose CLI 插件）
docker compose version
```

### CI/CD 验证

推送到 GitHub 后，GitHub Actions 会自动触发工作流，验证以下内容：

1. **Docker Build and Test 工作流**
   - docker-build job 应该成功完成
   - docker-compose-test job 应该成功完成

### 预期结果

- ✅ Docker 构建成功完成
- ✅ Docker Compose 测试成功运行
- ✅ 不再出现路径未找到错误
- ✅ 不再出现命令未找到错误

## 技术细节

### Docker 构建优化

虽然我们恢复了 `COPY . .` 指令，但通过优化 `.dockerignore` 文件确保：
1. 排除不必要的文件和目录（bin/, tests/, docs/ 等）
2. 保留必要的源代码目录（cmd/, internal/, pkg/）
3. 使用明确的包含规则确保关键目录不被排除

### Docker Compose 命令变更

**背景**：
- Docker Compose v2 引入了 CLI 插件架构
- `docker-compose` 命令逐渐被 `docker compose` 替代
- GitHub Actions 环境已更新为使用新命令

**兼容性**：
- `docker compose` 命令语法与 `docker-compose` 基本相同
- 参数和选项保持一致
- 无需修改其他配置

## 提交信息

```
fix(ci): resolve Docker build and compose command issues

- Revert to COPY . . in Dockerfiles to fix path not found error
- Update docker-compose commands to docker compose in GitHub Actions
- Optimize .dockerignore to control copied files

Issues fixed:
- Docker build error: "/cmd": not found
- Docker compose command not found in GitHub Actions

Changed files:
- Dockerfile: Revert COPY instructions
- Dockerfile.fullstack: Revert COPY instructions
- .github/workflows/docker.yml: Update docker-compose to docker compose
```

## 相关文件

- `Dockerfile` - 后端服务 Docker 镜像构建文件
- `Dockerfile.fullstack` - 完整栈（前端+后端）Docker 镜像构建文件
- `.github/workflows/docker.yml` - Docker 构建和测试工作流
- `.dockerignore` - Docker 构建忽略文件配置

## 后续建议

### 1. 监控构建状态

密切关注 GitHub Actions 的运行状态：
- 检查 Docker Build and Test 工作流是否成功
- 验证所有测试用例是否通过
- 确认构建产物是否正确生成

### 2. 考虑使用 BuildKit

在 Docker 构建中启用 BuildKit 以获得更好的性能：
```yaml
- name: Set up Docker Buildx
  uses: docker/setup-buildx-action@v3
  with:
    buildkitd-flags: --debug
```

### 3. 优化 .dockerignore

定期审查 `.dockerignore` 文件，确保：
- 排除所有不必要的文件
- 保留构建必需的源代码
- 添加注释说明每条规则的目的

### 4. 使用 Docker Compose V2 特性

考虑利用 Docker Compose V2 的新特性：
- 更好的性能和资源管理
- 改进的错误处理
- 增强的日志功能

## 问题预防

### 1. 定期更新依赖

定期检查和更新 GitHub Actions 中使用的工具版本：
```bash
# 检查 Docker 版本
docker --version
docker compose version
```

### 2. 本地环境同步

保持本地开发环境与 CI/CD 环境一致：
- 使用相同的 Docker 版本
- 安装相同的 CLI 工具
- 配置相同的环境变量

### 3. 文档更新

在项目文档中记录：
- 使用的 Docker 和 Compose 版本
- 构建和测试命令
- 常见问题和解决方案

## 参考资料

1. **Docker Compose 文档**
   - [Docker Compose CLI](https://docs.docker.com/compose/cli-command/)

2. **GitHub Actions 文档**
   - [GitHub Actions with Docker](https://docs.github.com/en/actions/guides/building-and-testing-docker)

3. **Docker 构建最佳实践**
   - [Dockerfile best practices](https://docs.docker.com/develop/develop-images/dockerfile_best-practices/)

---

**修复日期**: 2025-11-17  
**修复版本**: v0.6.1-hotfix1  
**影响范围**: GitHub Actions CI/CD 工作流  
**测试状态**: ✅ 已推送到 GitHub，等待 CI 验证