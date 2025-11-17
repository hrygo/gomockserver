# 修复健康检查参数问题

## 问题分析

根据错误信息 `dependency failed to start: container gomockserver-mockserver-test-1 is unhealthy`，问题出在 `gomockserver-mockserver-test-1` 容器不健康。虽然我们已经移除了 `container_name` 配置，但问题仍然存在。

经过进一步分析，发现问题可能出在健康检查参数设置上。服务启动可能需要更多时间，而当前的健康检查参数不足以应对较长的启动时间。

## 解决方案

我们需要增加健康检查的参数值，给服务更多时间来启动和通过健康检查。

## 实施步骤

1. 修改 `docker-compose.test.yml` 文件中的健康检查配置
   - 增加 `mockserver-test` 服务的健康检查超时时间从 5s 到 10s
   - 增加 `mockserver-test` 服务的健康检查重试次数从 10 次到 15 次
   - 增加 `mockserver-test` 服务的健康检查启动期从 30s 到 60s
   - 增加 `mongodb-test` 服务的健康检查超时时间从 5s 到 10s
   - 增加 `mongodb-test` 服务的健康检查重试次数从 10 次到 15 次
   - 增加 `mongodb-test` 服务的健康检查启动期从 20s 到 60s
2. 验证修改后的配置文件语法正确性

## 验证计划

1. 提交修改后的 `docker-compose.test.yml` 文件
2. 推送更改到 GitHub 仓库
3. 推送新的 bugfix 标签 `v0.6.1.bugfix15` 触发构建
4. 观察 GitHub Actions 构建是否成功完成
5. 如果构建成功，则问题已解决

## 后续步骤

一旦确认修复有效，可以继续进行版本标签的推送和验证工作。