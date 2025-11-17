# 修复 Docker Compose 测试错误问题

## 问题分析

根据 GitHub Actions 的错误信息 `failed to read dockerfile: open Dockerfile.test: no such file or directory`，问题出在 Docker Compose 测试过程中找不到 `Dockerfile.test` 文件。

在 `docker-compose.test.yml` 文件中，`mockserver-test` 服务的 Dockerfile 路径被设置为 `docker/Dockerfile.test`，但该项目中实际上并没有这个文件。在 `docker` 目录中只有以下文件：
- `Dockerfile` - 主要的 Dockerfile
- `Dockerfile.fullstack` - 完整栈 Dockerfile
- `Dockerfile.test-runner` - 测试运行器 Dockerfile

## 解决方案

我们需要修改 `docker-compose.test.yml` 文件，让 `mockserver-test` 服务使用现有的 `docker/Dockerfile` 文件，而不是不存在的 `docker/Dockerfile.test` 文件。

## 实施步骤

1. 修改 `docker-compose.test.yml` 文件中的 `mockserver-test` 服务配置
2. 将 `dockerfile: docker/Dockerfile.test` 更改为 `dockerfile: docker/Dockerfile`
3. 验证修改后的配置文件语法正确性

## 验证计划

1. 提交修改后的 `docker-compose.test.yml` 文件
2. 推送更改到 GitHub 仓库
3. 推送新的 bugfix 标签 `v0.6.1.bugfix10` 触发构建
4. 观察 GitHub Actions 构建是否成功完成
5. 如果构建成功，则问题已解决

## 后续步骤

一旦确认修复有效，可以继续进行版本标签的推送和验证工作。