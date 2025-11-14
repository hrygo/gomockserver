.PHONY: help test test-unit test-integration test-e2e test-all test-coverage build clean fmt vet lint docker-build docker-up docker-down docker-test run install dev

# 变量定义
BINARY_NAME=mockserver
BIN_DIR=bin
COVERAGE_FILE=coverage.out
COVERAGE_HTML=coverage.html
VERSION=$(shell git describe --tags --always --dirty 2>/dev/null || echo "v0.1.0-dev")
BUILD_TIME=$(shell date -u '+%Y-%m-%d_%H:%M:%S')
GIT_COMMIT=$(shell git rev-parse --short HEAD 2>/dev/null || echo "unknown")

# 构建标志
LDFLAGS=-ldflags "-X main.Version=$(VERSION) -X main.BuildTime=$(BUILD_TIME) -X main.GitCommit=$(GIT_COMMIT)"

# 默认目标
all: fmt vet test build

# 帮助信息
help:
	@echo "════════════════════════════════════════════════════════"
	@echo "  Mock Server - 构建和测试命令"
	@echo "════════════════════════════════════════════════════════"
	@echo ""
	@echo "📦 构建命令:"
	@echo "  make build           - 编译二进制文件"
	@echo "  make install         - 安装到 GOPATH/bin"
	@echo "  make clean           - 清理构建产物"
	@echo ""
	@echo "🧪 测试命令:"
	@echo "  make test            - 运行所有测试"
	@echo "  make test-unit       - 运行单元测试"
	@echo "  make test-integration - 运行集成测试"
	@echo "  make test-e2e        - 运行端到端测试"
	@echo "  make test-coverage   - 生成覆盖率报告"
	@echo "  make test-docker     - 在Docker环境中测试"
	@echo ""
	@echo "🔍 代码质量:"
	@echo "  make fmt             - 格式化代码"
	@echo "  make vet             - 运行 go vet"
	@echo "  make lint            - 运行 golangci-lint"
	@echo "  make check           - 运行所有检查"
	@echo ""
	@echo "🐳 Docker 命令:"
	@echo "  make docker-build    - 构建 Docker 镜像"
	@echo "  make docker-up       - 启动服务"
	@echo "  make docker-down     - 停止服务"
	@echo "  make docker-test     - Docker 测试环境"
	@echo "  make docker-logs     - 查看日志"
	@echo ""
	@echo "🚀 运行命令:"
	@echo "  make run             - 本地运行"
	@echo "  make dev             - 开发模式运行（带热重载）"
	@echo ""
	@echo "📚 其他:"
	@echo "  make deps            - 安装依赖"
	@echo "  make verify          - 快速验证（fmt+vet+build+test）"
	@echo "  make release         - 创建发布版本"
	@echo ""

# ════════════════════════════════════════════════════════
# 构建相关
# ════════════════════════════════════════════════════════

# 编译项目
build:
	@echo "🔨 Building Mock Server $(VERSION)..."
	@mkdir -p $(BIN_DIR)
	@go build $(LDFLAGS) -o $(BIN_DIR)/$(BINARY_NAME) ./cmd/mockserver
	@echo "✅ Build complete: $(BIN_DIR)/$(BINARY_NAME)"

# 安装到系统
install:
	@echo "📦 Installing Mock Server..."
	@go install $(LDFLAGS) ./cmd/mockserver
	@echo "✅ Installed to $(GOPATH)/bin/$(BINARY_NAME)"

# 交叉编译
build-all:
	@echo "🔨 Building for multiple platforms..."
	@mkdir -p $(BIN_DIR)
	@GOOS=linux GOARCH=amd64 go build $(LDFLAGS) -o $(BIN_DIR)/$(BINARY_NAME)-linux-amd64 ./cmd/mockserver
	@GOOS=linux GOARCH=arm64 go build $(LDFLAGS) -o $(BIN_DIR)/$(BINARY_NAME)-linux-arm64 ./cmd/mockserver
	@GOOS=darwin GOARCH=amd64 go build $(LDFLAGS) -o $(BIN_DIR)/$(BINARY_NAME)-darwin-amd64 ./cmd/mockserver
	@GOOS=darwin GOARCH=arm64 go build $(LDFLAGS) -o $(BIN_DIR)/$(BINARY_NAME)-darwin-arm64 ./cmd/mockserver
	@GOOS=windows GOARCH=amd64 go build $(LDFLAGS) -o $(BIN_DIR)/$(BINARY_NAME)-windows-amd64.exe ./cmd/mockserver
	@echo "✅ Cross-compilation complete"

# 清理构建文件
clean:
	@echo "🧹 Cleaning up..."
	@rm -rf $(BIN_DIR)/
	@rm -f $(COVERAGE_FILE) $(COVERAGE_HTML)
	@rm -f test-report-*.md
	@rm -rf docs/testing/coverage/*.html
	@find . -name "*.test" -delete
	@echo "✅ Cleanup complete"

# ════════════════════════════════════════════════════════
# 测试相关
# ════════════════════════════════════════════════════════

# 运行所有测试
test: test-unit

# 完整测试套件
test-all: test-unit test-integration
	@echo "✅ All tests completed!"

# 单元测试
test-unit:
	@echo "🧪 Running unit tests..."
	@go test -v -race -short ./...

# 单元测试（带覆盖率）
test-unit-coverage:
	@echo "🧪 Running unit tests with coverage..."
	@go test -v -race -short -coverprofile=$(COVERAGE_FILE) ./...
	@go tool cover -func=$(COVERAGE_FILE) | tail -1

# 集成测试
test-integration:
	@echo "🔗 Running integration tests..."
	@chmod +x ./tests/integration/e2e_test.sh
	@ADMIN_API=http://localhost:8080/api/v1 MOCK_API=http://localhost:9090 SKIP_SERVER_START=false ./tests/integration/e2e_test.sh

# 端到端测试（Docker环境）
test-e2e:
	@echo "🌐 Running E2E tests in Docker..."
	@docker-compose -f docker-compose.test.yml --profile integration run --rm test-runner

# Docker 测试环境
test-docker:
	@echo "🐳 Running tests in Docker environment..."
	@docker-compose -f docker-compose.test.yml up -d mongodb-test mockserver-test
	@sleep 10
	@docker-compose -f docker-compose.test.yml --profile integration run --rm test-runner
	@docker-compose -f docker-compose.test.yml down

# 生成测试覆盖率报告
test-coverage:
	@echo "📊 Generating coverage report..."
	@go test -v -coverprofile=$(COVERAGE_FILE) -covermode=atomic ./...
	@go tool cover -html=$(COVERAGE_FILE) -o $(COVERAGE_HTML)
	@go tool cover -func=$(COVERAGE_FILE) | tail -1
	@echo "📈 Coverage report: $(COVERAGE_HTML)"

# 性能测试
test-perf:
	@echo "⚡ Running performance tests..."
	@chmod +x ./tests/performance/run_perf_tests.sh
	@./tests/performance/run_perf_tests.sh

# Benchmark测试
bench:
	@echo "📊 Running benchmarks..."
	@go test -bench=. -benchmem ./...

# ════════════════════════════════════════════════════════
# 代码质量
# ════════════════════════════════════════════════════════

# 格式化代码
fmt:
	@echo "✨ Formatting code..."
	@gofmt -w .
	@goimports -w . 2>/dev/null || true
	@echo "✅ Code formatted"

# 运行 go vet
vet:
	@echo "🔍 Running go vet..."
	@go vet ./...
	@echo "✅ Vet check passed"

# 运行 golangci-lint
lint:
	@echo "🔍 Running golangci-lint..."
	@if command -v golangci-lint > /dev/null; then \
		golangci-lint run --timeout=5m; \
	else \
		echo "⚠️  golangci-lint not installed, skipping..."; \
	fi

# 运行所有检查
check: fmt vet lint
	@echo "✅ All checks passed"

# 安全检查
security:
	@echo "🔒 Running security checks..."
	@if command -v gosec > /dev/null; then \
		gosec ./...; \
	else \
		echo "⚠️  gosec not installed, install with: go install github.com/securego/gosec/v2/cmd/gosec@latest"; \
	fi

# ════════════════════════════════════════════════════════
# Docker 相关
# ════════════════════════════════════════════════════════

# 构建 Docker 镜像
docker-build:
	@echo "🐳 Building Docker image..."
	@docker build -t mockserver:$(VERSION) -t mockserver:latest .
	@echo "✅ Docker image built: mockserver:$(VERSION)"

# 启动 Docker 服务
docker-up:
	@echo "🚀 Starting Docker services..."
	@docker-compose up -d
	@echo "⏳ Waiting for services to be ready..."
	@sleep 5
	@docker-compose ps
	@echo "✅ Services are running"

# 停止 Docker 服务
docker-down:
	@echo "🛑 Stopping Docker services..."
	@docker-compose down
	@echo "✅ Services stopped"

# 查看 Docker 日志
docker-logs:
	@docker-compose logs -f

# Docker 测试环境
docker-test-up:
	@echo "🐳 Starting Docker test environment..."
	@docker-compose -f docker-compose.test.yml up -d
	@sleep 10
	@docker-compose -f docker-compose.test.yml ps

docker-test-down:
	@echo "🛑 Stopping Docker test environment..."
	@docker-compose -f docker-compose.test.yml down

# 清理 Docker 资源
docker-clean:
	@echo "🧹 Cleaning Docker resources..."
	@docker-compose down -v
	@docker-compose -f docker-compose.test.yml down -v
	@echo "✅ Docker cleanup complete"

# ════════════════════════════════════════════════════════
# 运行相关
# ════════════════════════════════════════════════════════

# 本地运行
run:
	@echo "🚀 Starting Mock Server..."
	@go run ./cmd/mockserver/main.go

# 后台运行
run-bg:
	@echo "🚀 Starting Mock Server in background..."
	@nohup $(BIN_DIR)/$(BINARY_NAME) > /dev/null 2>&1 &
	@echo "✅ Server started in background"

# 开发模式（自动重载）
dev:
	@echo "🔧 Starting in development mode..."
	@if command -v air > /dev/null; then \
		air; \
	else \
		echo "⚠️  air not installed, install with: go install github.com/cosmtrek/air@latest"; \
		echo "Falling back to normal run..."; \
		make run; \
	fi

# ════════════════════════════════════════════════════════
# 依赖管理
# ════════════════════════════════════════════════════════

# 安装依赖
deps:
	@echo "📦 Installing dependencies..."
	@go mod download
	@go mod tidy
	@echo "✅ Dependencies installed"

# 检查依赖
deps-check:
	@echo "🔍 Checking dependencies..."
	@go mod verify
	@echo "✅ Dependencies verified"

# 更新依赖
deps-update:
	@echo "⬆️  Updating dependencies..."
	@go get -u ./...
	@go mod tidy
	@echo "✅ Dependencies updated"

# ════════════════════════════════════════════════════════
# 发布相关
# ════════════════════════════════════════════════════════

# 创建发布版本
release:
	@echo "📦 Creating release $(VERSION)..."
	@make clean
	@make test-all
	@make build-all
	@echo "✅ Release $(VERSION) ready"

# 快速验证（格式化+检查+构建+测试）
verify: fmt vet lint build test-unit
	@echo "✅ Quick verification complete!"

# 预提交检查
pre-commit: fmt vet lint test-unit
	@echo "✅ Pre-commit checks passed!"

# 显示版本信息
version:
	@echo "Version:    $(VERSION)"
	@echo "Build Time: $(BUILD_TIME)"
	@echo "Git Commit: $(GIT_COMMIT)"
