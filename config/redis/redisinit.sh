#!/bin/sh

# Redis初始化脚本
# 在Redis启动后执行一些初始化操作

echo "正在初始化Redis缓存配置..."

# 等待Redis完全启动
sleep 2

# 设置一些基础的缓存配置
redis-cli CONFIG SET maxmemory-policy allkeys-lru
redis-cli CONFIG SET timeout 300

# 创建一些缓存空间（可选）
redis-cli SET "mockserver:cache:init" "$(date '+%Y-%m-%d %H:%M:%S')"
redis-cli EXPIRE "mockserver:cache:init" 86400

# 设置缓存相关的监控键
redis-cli LPUSH "mockserver:cache:stats:operations" "init:$(date '+%s')"

echo "Redis缓存初始化完成！"