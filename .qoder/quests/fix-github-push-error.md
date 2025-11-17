# 修复 GitHub Actions 推送错误问题

## 问题分析

经过仔细分析 GitHub Actions 工作流文件和相关配置，发现两个潜在问题：

1. 根据 GitHub 的公告，从 2024 年 7 月 9 日开始，GitHub 托管的运行器镜像中将移除 Docker Compose v1。虽然当前配置已使用正确的 `docker compose` 命令（不带连字符），但可能仍存在兼容性问题。

2. 在 `docker-compose.test.yml` 文件中，`test-runner` 服务使用了 `profiles: integration` 配置，这意味着需要在运行 `docker compose` 命令时显式指定 `--profile integration` 参数才能启动该服务。

## 解决方案

1. 确保所有 `docker compose` 命令都使用正确的语法（不带连字符）
2. 在需要运行集成测试的命令中添加 `--profile integration` 参数

## 实施步骤

1. 检查 `.github/workflows/docker.yml` 文件中的所有 `docker compose` 命令
2. 确保在运行测试的命令中包含 `--profile integration` 参数
3. 验证修改后的配置文件语法正确性

## 验证计划

1. 提交修改后的 GitHub Actions 配置文件
2. 推送更改到 GitHub 仓库
3. 观察 GitHub Actions 构建是否成功完成
4. 如果构建成功，则问题已解决

## 后续步骤

一旦确认修复有效，可以继续进行版本标签的推送和验证工作。

## 修改详情

我们需要在 `.github/workflows/docker.yml` 文件中检查以下部分：

1. 确保 `docker compose` 命令使用正确的语法（不带连字符）

2. 确保在运行测试的命令中包含 `--profile integration` 参数：

```yaml
- name: Run integration tests
  run: |
    docker compose -f docker-compose.test.yml --profile integration run --rm test-runner
```

3. 确保 `Build test runner image` 步骤包含 `tags` 参数：

```yaml
- name: Build test runner image
  uses: docker/build-push-action@v5
  with:
    context: .
    file: ./docker/Dockerfile.test-runner
    push: false
    tags: mockserver:test-runner
```