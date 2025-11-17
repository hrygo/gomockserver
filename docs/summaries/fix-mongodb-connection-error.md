# 修复 MongoDB 连接错误问题

## 问题分析

根据 GitHub Actions 的错误信息 `dial tcp: lookup host.docker.internal on 127.0.0.11:53: no such host`，问题出在 `mockserver-test` 服务无法连接到 MongoDB。错误信息显示它尝试连接到 `host.docker.internal:27018`，但无法解析 `host.docker.internal` 主机名。

在 `config.test.yaml` 文件中，MongoDB 的 URI 被设置为 `mongodb://host.docker.internal:27018`，这在某些环境中可能无法工作，因为 `host.docker.internal` 主机名可能无法解析。

## 解决方案

我们需要修改 `config.test.yaml` 文件，使其使用 Docker 内部网络来连接 MongoDB，而不是通过 `host.docker.internal`。

## 实施步骤

1. 修改 `config.test.yaml` 文件中的 MongoDB URI 配置
2. 将 `uri: "mongodb://host.docker.internal:27018"` 更改为 `uri: "mongodb://mongodb-test:27017"`
3. 验证修改后的配置文件语法正确性

## 验证计划

1. 提交修改后的 `config.test.yaml` 文件
2. 推送更改到 GitHub 仓库
3. 推送新的 bugfix 标签 `v0.6.1.bugfix12` 触发构建
4. 观察 GitHub Actions 构建是否成功完成
5. 如果构建成功，则问题已解决

## 后续步骤

一旦确认修复有效，可以继续进行版本标签的推送和验证工作。