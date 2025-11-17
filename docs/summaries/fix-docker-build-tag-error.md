# 修复 Docker 构建标签错误问题

## 问题分析

根据 GitHub Actions 的错误信息 `ERROR: failed to build: tag is needed when pushing to registry`，问题出现在 Docker 构建推送过程中。虽然我们在配置中使用了 `docker/metadata-action` 来生成标签，但在通过标签触发构建时，可能没有生成正确的标签信息。

在 `.github/workflows/docker.yml` 文件中，我们有以下配置：

```yaml
- name: Extract metadata
  id: meta
  uses: docker/metadata-action@v5
  with:
    images: ${{ env.REGISTRY }}/${{ env.IMAGE_NAME }}
    tags: |
      type=ref,event=branch
      type=ref,event=pr
      type=semver,pattern={{version}}
      type=semver,pattern={{major}}.{{minor}}
```

当通过标签 `v0.6.1.bugfix6` 触发构建时，`type=semver` 应该能够匹配并生成标签，但似乎没有正确工作。

## 解决方案

我们需要修改 `docker/metadata-action` 的配置，确保在标签推送时能够正确生成标签。我们将添加一个 fallback 标签，以确保在任何情况下都有标签可用。

## 实施步骤

1. 修改 `.github/workflows/docker.yml` 文件中的 "Extract metadata" 步骤
2. 添加针对 tag 事件的标签生成规则
3. 添加 fallback 标签以确保始终有标签可用

## 验证计划

1. 提交修改后的 GitHub Actions 配置文件
2. 推送更改到 GitHub 仓库
3. 推送新的 bugfix 标签 `v0.6.1.bugfix7` 触发构建
4. 观察 GitHub Actions 构建是否成功完成
5. 如果构建成功，则问题已解决

## 后续步骤

一旦确认修复有效，可以继续进行版本标签的推送和验证工作。