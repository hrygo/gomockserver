# Build stage
FROM golang:alpine AS builder

WORKDIR /app

# 设置 Go 代理和 Alpine 镜像源
ENV GOPROXY=https://mirrors.aliyun.com/goproxy,direct

# 安装必要的构建工具
RUN apk add --no-cache git

# 复制 go mod 文件
COPY go.mod go.sum ./
RUN go mod download

# 复制源代码
COPY . .

# 调试：验证源代码是否正确复制
RUN ls -la ./cmd/
RUN ls -la ./cmd/mockserver/
RUN cat ./cmd/mockserver/main.go | head -20

# 构建应用
# 调试：显示详细的构建信息
RUN echo "Building with Go version: $(go version)"
RUN echo "Current directory: $(pwd)"
RUN echo "Go env: $(go env)"
RUN CGO_ENABLED=0 GOOS=linux go build -v -a -installsuffix cgo -o mockserver ./cmd/mockserver

# Runtime stage
FROM alpine:latest

WORKDIR /root/

# 从构建阶段复制二进制文件
COPY --from=builder /app/mockserver .
COPY --from=builder /app/config.yaml .

# 暴露端口
EXPOSE 8080 9090

# 运行应用
CMD ["./mockserver"]
