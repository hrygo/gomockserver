#!/bin/bash

# 启动 MongoDB 测试实例
echo "Starting MongoDB test instance..."
docker run -d --name mongodb-test -p 27018:27017 mongo:6.0 --quiet

# 等待 MongoDB 启动
echo "Waiting for MongoDB to start..."
sleep 15

# 构建测试镜像
echo "Building test image..."
docker build -t mockserver-test -f Dockerfile.test .

# 运行测试容器
echo "Running test container..."
docker run -d --name mockserver-test -p 8080:8080 -p 9090:9090 mockserver-test

# 等待应用启动
echo "Waiting for application to start..."
sleep 15

# 测试健康检查端点，最多重试5次
echo "Testing health check endpoint..."
for i in {1..5}; do
  echo "Attempt $i..."
  if curl -f http://localhost:8080/api/v1/system/health; then
    echo "Health check passed!"
    break
  else
    echo "Health check failed, retrying in 5 seconds..."
    sleep 5
  fi
done

# 清理
echo "Cleaning up..."
docker stop mockserver-test mongodb-test
docker rm mockserver-test mongodb-test

echo "Test completed!"