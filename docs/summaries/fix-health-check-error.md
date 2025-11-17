# 修复健康检查错误问题

## 问题分析

根据 GitHub Actions 的错误信息 `curl: (7) Failed to connect to localhost port 8081 after 0 ms: Couldn't connect to server`，问题出在健康检查步骤中无法连接到 `localhost:8081`。这表明 `mockserver-test` 服务没有正确启动或者没有在指定端口监听。

可能的原因包括：
1. `mockserver-test` 服务启动失败
2. 服务启动需要更多时间
3. 服务没有在正确的端口监听
4. Docker Compose 配置问题

## 解决方案

我们需要修改 GitHub Actions 配置文件，增加等待时间并添加更详细的错误诊断信息。同时，我们也会增加重试机制来确保服务有足够的时间启动。

## 实施步骤

1. 修改 `.github/workflows/docker.yml` 文件中的 "Start test environment" 和 "Check services health" 步骤
2. 增加等待时间从 15 秒到 30 秒
3. 添加重试机制，最多尝试 30 次，每次间隔 2 秒
4. 添加服务日志输出以便诊断问题

## 验证计划

1. 提交修改后的 GitHub Actions 配置文件
2. 推送更改到 GitHub 仓库
3. 推送新的 bugfix 标签 `v0.6.1.bugfix11` 触发构建
4. 观察 GitHub Actions 构建日志以获取更多调试信息
5. 根据调试信息进一步调整解决方案

## 后续步骤

一旦获取到足够的调试信息，可以进一步调整 Docker Compose 配置或服务启动参数来解决问题。