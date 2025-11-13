.PHONY: help test test-unit test-integration test-e2e test-all test-coverage build clean fmt vet lint docker-up docker-down run

# 默认目标
help:
	@echo "Mock Server 测试和构建命令"
	@echo ""
	@echo "测试命令:"
	@echo "  make test-static     - 运行静态检查测试"
	@echo "  make test-unit       - 运行单元测试"
	@echo "  make test-integration - 运行集成测试"
	@echo "  make test-e2e        - 运行端到端测试"
	@echo "  make test-all        - 运行所有测试"
	@echo "  make test-coverage   - 生成测试覆盖率报告"
	@echo ""
	@echo "代码质量:"
	@echo "  make fmt             - 格式化代码"
	@echo "  make vet             - 运行go vet检查"
	@echo "  make lint            - 运行golangci-lint"
	@echo ""
	@echo "构建和运行:"
	@echo "  make build           - 编译项目"
	@echo "  make run             - 本地运行服务"
	@echo "  make clean           - 清理构建文件"
	@echo ""
	@echo "Docker:"
	@echo "  make docker-up       - 启动Docker服务"
	@echo "  make docker-down     - 停止Docker服务"
	@echo "  make docker-logs     - 查看Docker日志"
	@echo ""

# 静态检查测试
test-static:
	@echo "Running static tests..."
	@chmod +x ./mvp-test.sh
	@./mvp-test.sh

# 单元测试
test-unit:
	@echo "Running unit tests..."
	@go test -v -short -race ./internal/...

# 集成测试（需要MongoDB）
test-integration:
	@echo "Running integration tests..."
	@chmod +x ./test.sh
	@./test.sh

# 端到端测试
test-e2e:
	@echo "Running E2E tests..."
	@echo "E2E tests not implemented yet"

# 运行所有测试
test-all: test-static test-unit test-integration
	@echo "All tests completed!"

# 生成测试覆盖率报告
test-coverage:
	@echo "Generating test coverage report..."
	@go test -v -coverprofile=coverage.out ./internal/...
	@go tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report generated: coverage.html"

# 格式化代码
fmt:
	@echo "Formatting code..."
	@gofmt -w .
	@echo "Code formatted successfully!"

# 运行 go vet
vet:
	@echo "Running go vet..."
	@go vet ./...

# 运行 golangci-lint (如果安装)
lint:
	@echo "Running golangci-lint..."
	@if command -v golangci-lint > /dev/null; then \
		golangci-lint run; \
	else \
		echo "golangci-lint not installed, skipping..."; \
	fi

# 编译项目
build:
	@echo "Building Mock Server..."
	@go build -o bin/mockserver ./cmd/mockserver
	@echo "Build complete: bin/mockserver"

# 本地运行
run:
	@echo "Starting Mock Server locally..."
	@go run ./cmd/mockserver/main.go

# 清理构建文件
clean:
	@echo "Cleaning up..."
	@rm -rf bin/
	@rm -f coverage.out coverage.html
	@rm -f test-report-*.md
	@echo "Cleanup complete!"

# Docker 命令
docker-up:
	@echo "Starting Docker services..."
	@docker-compose up -d
	@echo "Waiting for services to be ready..."
	@sleep 5
	@docker-compose ps

docker-down:
	@echo "Stopping Docker services..."
	@docker-compose down

docker-logs:
	@docker-compose logs -f

# 安装依赖
deps:
	@echo "Installing dependencies..."
	@go mod download
	@go mod tidy

# 检查依赖
deps-check:
	@echo "Checking dependencies..."
	@go mod verify

# 快速验证（格式化+编译+静态测试）
verify: fmt vet build test-static
	@echo "Quick verification complete!"
