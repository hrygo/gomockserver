# 添加调试信息以诊断服务启动问题

## 问题分析

根据错误信息 `dependency failed to start: container gomockserver-mockserver-test-1 is unhealthy`，问题出在 `gomockserver-mockserver-test-1` 容器不健康。用户认为问题不是等待时间的问题，而是其他原因导致的。

经过进一步分析，我们怀疑问题可能出在服务启动过程中，但目前缺乏足够的调试信息来确定具体原因。

## 解决方案

我们需要添加更多的调试信息，以便更好地了解服务启动过程中发生了什么。

## 实施步骤

1. 修改 `docker/Dockerfile` 文件
   - 在 CMD 命令中明确指定配置文件路径
2. 修改 `docker-compose.test.yml` 文件
   - 添加 DEBUG 环境变量以启用更详细的日志
3. 验证修改后的配置文件语法正确性

## 验证计划

1. 提交修改后的 `docker/Dockerfile` 和 `docker-compose.test.yml` 文件
2. 推送更改到 GitHub 仓库
3. 推送新的 bugfix 标签 `v0.6.1.bugfix16` 触发构建
4. 观察 GitHub Actions 构建日志以获取更多调试信息
5. 根据调试信息进一步分析问题原因

## 后续步骤

一旦获取到足够的调试信息，可以进一步调整配置来解决问题。