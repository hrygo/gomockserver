# 修复服务依赖错误问题

## 问题分析

根据错误信息 `dependency failed to start: container mockserver-test-app is unhealthy`，问题出在 `mockserver-test-app` 容器不健康，导致依赖它的服务无法启动。这表明服务启动失败或健康检查失败。

在之前的修复中，我们已经解决了 MongoDB 连接配置问题，但可能还有其他因素导致服务无法正常启动或通过健康检查。

## 解决方案

我们需要从以下几个方面来解决这个问题：

1. 增加健康检查的超时时间和重试次数，给服务更多时间启动
2. 确保所有相关的环境变量都正确设置
3. 验证服务间的网络连接

## 实施步骤

1. 修改 `docker-compose.test.yml` 文件中的健康检查配置
   - 增加 `mockserver-test` 服务的健康检查超时时间和重试次数
   - 增加 `mongodb-test` 服务的健康检查重试次数
2. 修改 `docker/Dockerfile.test-runner` 文件
   - 添加缺失的 `MONGODB_URI` 环境变量
3. 验证修改后的配置文件语法正确性

## 验证计划

1. 提交修改后的配置文件
2. 推送更改到 GitHub 仓库
3. 推送新的 bugfix 标签 `v0.6.1.bugfix13` 触发构建
4. 观察 GitHub Actions 构建是否成功完成
5. 如果构建成功，则问题已解决

## 后续步骤

一旦确认修复有效，可以继续进行版本标签的推送和验证工作。