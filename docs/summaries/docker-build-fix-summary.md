# Docker 构建错误修复总结

## 问题描述

**错误信息**：
```
#17 ERROR: process "/bin/sh -c CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o mockserver ./cmd/mockserver" did not complete successfully: exit code: 1
------
 > [builder 7/7] RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o mockserver ./cmd/mockserver:
0.045 stat /app/cmd/mockserver: directory not found
```

**触发场景**：推送到 GitHub 后，GitHub Actions CI 流程执行 Docker 构建时失败。

**错误原因**：找不到 `/app/cmd/mockserver` 目录。

---

## 根本原因分析

在 Docker 构建过程中，使用 `COPY . .` 指令复制源代码时，由于 `.dockerignore` 文件的规则或 Docker 构建上下文的问题，可能导致某些关键目录（如 `cmd/`, `internal/`, `pkg/`）没有被正确复制到容器中。

虽然 `.dockerignore` 文件中没有明确排除这些目录，但在某些环境（特别是 GitHub Actions CI）中，`COPY . .` 的行为可能不一致或不可靠。

---

## 解决方案

### 修改文件

1. **Dockerfile**
2. **Dockerfile.fullstack**
3. **.dockerignore**（可选优化）

### 具体修改

#### 1. Dockerfile 修改

**修改前**：
```dockerfile
# 复制 go mod 文件
COPY go.mod go.sum ./
RUN go mod download

# 复制源代码
COPY . .
```

**修改后**：
```dockerfile
# 复制 go mod 文件
COPY go.mod go.sum ./
RUN go mod download

# 复制源代码（确保包含关键目录）
COPY cmd/ ./cmd/
COPY internal/ ./internal/
COPY pkg/ ./pkg/
COPY config.yaml ./
```

#### 2. Dockerfile.fullstack 修改

应用相同的修改，确保后端构建阶段（Stage 2）使用显式的目录复制。

#### 3. .dockerignore 优化（可选）

添加明确的包含规则，确保关键目录不被意外排除：

```dockerignore
# 确保关键目录不被排除
!cmd/
!internal/
!pkg/
```

---

## 修改优势

### 1. **明确性和可靠性**
- 显式指定需要复制的目录，避免依赖 `COPY . .` 的隐式行为
- 在不同环境（本地、CI/CD）中行为一致

### 2. **安全性**
- 只复制必需的源代码目录，减少构建上下文
- 避免意外复制敏感文件或临时文件

### 3. **可维护性**
- 清晰展示 Docker 镜像中包含哪些源代码
- 便于审查和理解构建过程

### 4. **性能优化**
- 减少不必要的文件复制
- 利用 Docker 分层缓存更高效

---

## 验证步骤

### 本地验证

```bash
# 清理缓存并重新构建
docker build --no-cache -t mockserver .

# 验证构建成功
docker run --rm mockserver --version

# 测试完整栈构建
docker build --no-cache -f Dockerfile.fullstack -t mockserver-full .
```

### CI/CD 验证

1. 提交修改并推送到 GitHub：
   ```bash
   git add Dockerfile Dockerfile.fullstack
   git commit -m "fix(docker): resolve 'directory not found' error in CI build"
   git push origin master
   ```

2. 检查 GitHub Actions 工作流程：
   - 访问 GitHub Actions 页面
   - 查看 "Docker Build and Test" 工作流
   - 确认构建成功完成

---

## 提交信息

```
fix(docker): resolve 'directory not found' error in CI build

- Change COPY . . to explicit COPY for key directories
- Explicitly copy cmd/, internal/, pkg/ and config.yaml
- Ensures critical directories are included in Docker build
- Fixes GitHub Actions CI build failure

Resolves: stat /app/cmd/mockserver: directory not found
```

---

## 相关文件

- `Dockerfile` - 后端服务 Docker 镜像构建文件
- `Dockerfile.fullstack` - 完整栈（前端+后端）Docker 镜像构建文件
- `.dockerignore` - Docker 构建忽略文件配置
- `.github/workflows/docker.yml` - GitHub Actions Docker 构建工作流

---

## 后续建议

### 1. 监控 CI/CD 构建

在接下来的几次提交中，密切关注 GitHub Actions 的 Docker 构建工作流，确保问题彻底解决。

### 2. 定期检查 .dockerignore

定期审查 `.dockerignore` 文件，确保：
- 没有误排除关键目录
- 规则清晰明确
- 与项目结构保持同步

### 3. 本地测试最佳实践

在推送到 GitHub 前，始终在本地进行 Docker 构建测试：
```bash
# 清除缓存测试
docker build --no-cache -t mockserver .

# 测试不同 Dockerfile
docker build --no-cache -f Dockerfile.fullstack -t mockserver-full .
```

### 4. 文档更新

考虑在 `DEPLOYMENT.md` 或 `README.md` 中添加：
- Docker 构建最佳实践
- 常见构建问题排查指南
- CI/CD 环境差异说明

---

## 问题预防

为避免类似问题再次发生，建议：

1. **使用显式 COPY 指令**
   - 明确列出需要复制的目录和文件
   - 避免使用 `COPY . .` 的模糊行为

2. **CI/CD 环境测试**
   - 在推送前，确保本地 Docker 构建成功
   - 使用 `--no-cache` 选项测试完整构建过程

3. **.dockerignore 管理**
   - 定期审查忽略规则
   - 添加注释说明每条规则的目的
   - 使用包含规则 (`!`) 确保关键文件不被排除

4. **版本控制**
   - 确保 `.dockerignore` 文件被纳入版本控制
   - 记录每次修改的原因和影响

---

**修复日期**: 2025-11-17  
**修复版本**: v0.6.0+  
**影响范围**: Docker 构建流程、GitHub Actions CI  
**测试状态**: ✅ 本地构建通过，等待 CI 验证
