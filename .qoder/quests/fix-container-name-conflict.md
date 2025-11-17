# 修复容器名称冲突问题

## 问题分析

根据错误信息 `dependency failed to start: container mockserver-test-app is unhealthy`，问题出在 `mockserver-test-app` 容器不健康。经过进一步分析，发现问题可能出在 Docker Compose 配置中的 `container_name` 设置上。

在 `docker-compose.test.yml` 文件中，我们为多个服务设置了 `container_name`，这可能会导致以下问题：
1. 服务间的网络连接可能无法正常工作
2. 健康检查可能无法正确执行
3. 依赖关系可能无法正确建立

## 解决方案

我们需要移除 Docker Compose 配置中的 `container_name` 设置，让 Docker Compose 使用默认的服务名称。这样可以确保服务间的网络连接和依赖关系能够正常工作。

## 实施步骤

1. 修改 `docker-compose.test.yml` 文件
   - 移除 `mongodb-test` 服务的 `container_name` 配置
   - 移除 `mockserver-test` 服务的 `container_name` 配置
   - 移除 `redis-test` 服务的 `container_name` 配置
   - 移除 `wrk-test` 服务的 `container_name` 配置
   - 移除 `test-runner` 服务的 `container_name` 配置
2. 验证修改后的配置文件语法正确性

## 验证计划

1. 提交修改后的 `docker-compose.test.yml` 文件
2. 推送更改到 GitHub 仓库
3. 推送新的 bugfix 标签 `v0.6.1.bugfix14` 触发构建
4. 观察 GitHub Actions 构建是否成功完成
5. 如果构建成功，则问题已解决

## 后续步骤

一旦确认修复有效，可以继续进行版本标签的推送和验证工作。