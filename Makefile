.PHONY: help test test-unit test-integration test-e2e test-all test-coverage build clean fmt vet lint docker-build docker-up docker-down docker-test run install dev start-mongo stop-mongo restart-mongo mongo-shell mongo-logs dev-env clean-env test-service-coverage test-api-coverage start-all stop-all start-backend stop-backend start-frontend stop-frontend

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
	@echo "  make build           - 编译后端二进制文件"
	@echo "  make build-frontend  - 构建前端（npm run build）"
	@echo "  make build-fullstack - 构建完整应用（前端+后端）"
	@echo "  make install         - 安装到 GOPATH/bin"
	@echo "  make clean           - 清理构建产物"
	@echo "  make build-platforms - 跨平台编译"
	@echo ""
	@echo "🧪 测试命令:"
	@echo "  make test            - 运行所有测试"
	@echo "  make test-unit       - 运行单元测试"
	@echo "  make test-service    - 运行 Service 层测试"
	@echo "  make test-api        - 运行 API 层测试"
	@echo "  make test-repository - 运行 Repository 层测试"
	@echo "  make test-integration - 运行集成测试"
	@echo "  make test-e2e        - 运行端到端测试"
	@echo "  make test-coverage   - 生成覆盖率报告"
	@echo "  make test-coverage-check - 检查覆盖率门限 (70%)"
	@echo "  make test-docker     - 在Docker环境中测试"
	@echo "  make bench           - 运行性能基准测试"
	@echo ""
	@echo "🔍 代码质量:"
	@echo "  make fmt             - 格式化代码"
	@echo "  make vet             - 运行 go vet"
	@echo "  make lint            - 运行 golangci-lint"
	@echo "  make check           - 运行所有检查"
	@echo "  make code-analysis   - 代码分析"
	@echo "  make security        - 安全扫描"
	@echo "  make qa              - 质量检查 (fmt+vet+lint+test)"
	@echo "  make pre-push        - 推送前检查 (qa+integration)"
	@echo "  make pre-commit      - 提交前检查"
	@echo ""
	@echo "🐳 Docker 命令:"
	@echo "  make docker-build      - 构建后端 Docker 镜像"
	@echo "  make docker-build-full - 构建完整 Docker 镜像（包含前端）"
	@echo "  make docker-up         - 启动服务"
	@echo "  make docker-down       - 停止服务"
	@echo "  make docker-test       - Docker 测试环境"
	@echo "  make docker-logs       - 查看日志"
	@echo ""
	@echo "🚀 运行命令:"
	@echo "  make run             - 本地运行后端"
	@echo "  make dev             - 开发模式运行（带热重载）"
	@echo "  make start-mongo     - 启动 MongoDB 容器"
	@echo "  make stop-mongo      - 停止 MongoDB 容器"
	@echo "  make mongo-shell     - 连接 MongoDB Shell"
	@echo "  make start-all       - 启动全栈应用 (MongoDB + 后端 + 前端)"
	@echo "  make stop-all        - 停止全栈应用"
	@echo "  make start-backend   - 后台运行（使用 dev 配置）"
	@echo "  make start-frontend  - 前端运行"
	@echo ""
	@echo "📚 其他:"
	@echo "  make deps            - 安装依赖"
	@echo "  make deps-check      - 检查依赖"
	@echo "  make deps-upgrade    - 检查依赖升级"
	@echo "  make mock-generate   - 生成 Mock 对象"
	@echo "  make verify          - 快速验证（fmt+vet+build+test）"
	@echo "  make release         - 创建发布版本"
	@echo "  make version         - 显示版本信息"
	@echo "  make dev-env         - 启动开发环境 (MongoDB)"
	@echo "  make clean-env       - 清理开发环境"
	@echo "  make t               - 别名: make test"
	@echo "  make c               - 别名: make test-coverage"
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
build-platforms:
	@echo "🔨 Building for multiple platforms..."
	@mkdir -p $(BIN_DIR)
	@GOOS=linux GOARCH=amd64 go build $(LDFLAGS) -o $(BIN_DIR)/$(BINARY_NAME)-linux-amd64 ./cmd/mockserver
	@GOOS=linux GOARCH=arm64 go build $(LDFLAGS) -o $(BIN_DIR)/$(BINARY_NAME)-linux-arm64 ./cmd/mockserver
	@GOOS=darwin GOARCH=amd64 go build $(LDFLAGS) -o $(BIN_DIR)/$(BINARY_NAME)-darwin-amd64 ./cmd/mockserver
	@GOOS=darwin GOARCH=arm64 go build $(LDFLAGS) -o $(BIN_DIR)/$(BINARY_NAME)-darwin-arm64 ./cmd/mockserver
	@GOOS=windows GOARCH=amd64 go build $(LDFLAGS) -o $(BIN_DIR)/$(BINARY_NAME)-windows-amd64.exe ./cmd/mockserver
	@echo "✅ Cross-compilation complete"

# 构建前端
build-frontend:
	@echo "🎨 Building frontend..."
	@if [ -d "web/frontend" ]; then \
		cd web/frontend && \
		echo "📦 Installing dependencies..." && \
		npm install && \
		echo "🔨 Building frontend..." && \
		npm run build && \
		echo "✅ Frontend build complete: web/frontend/dist"; \
	else \
		echo "❌ Frontend directory not found"; \
		exit 1; \
	fi

# 构建完整应用（前端+后端）
build-fullstack: build-frontend build
	@echo "✅ Fullstack build complete"
	@echo "  - Frontend: web/frontend/dist"
	@echo "  - Backend:  $(BIN_DIR)/$(BINARY_NAME)"

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

# Repository层测试
test-repository:
	@echo "🧪 Running Repository layer tests..."
	@go test -v -race -tags=integration ./internal/repository/...

# Service层测试
test-service:
	@echo "🧪 Running Service layer tests..."
	@go test -v -race ./internal/service/...

# API层测试
test-api:
	@echo "🧪 Running API layer tests..."
	@go test -v -race ./internal/api/...

# Repository层测试覆盖率
test-repository-coverage:
	@echo "📊 Running Repository layer tests with coverage..."
	@mkdir -p scripts/coverage
	@go test -v -race -tags=integration -coverprofile=scripts/coverage/repository-coverage.out ./internal/repository/...
	@go tool cover -html=scripts/coverage/repository-coverage.out -o scripts/coverage/repository-coverage.html
	@go tool cover -func=scripts/coverage/repository-coverage.out | tail -1
	@echo "📈 Coverage report: scripts/coverage/repository-coverage.html"

# Service层测试覆盖率
test-service-coverage:
	@echo "📊 Running Service layer tests with coverage..."
	@mkdir -p scripts/coverage
	@go test -v -race -coverprofile=scripts/coverage/service-coverage.out ./internal/service/...
	@go tool cover -html=scripts/coverage/service-coverage.out -o scripts/coverage/service-coverage.html
	@COVERAGE=$$(go tool cover -func=scripts/coverage/service-coverage.out | grep total | awk '{print $$3}' | sed 's/%//'); \
	echo "📈 Service layer coverage: $$COVERAGE%"; \
	if [ $$(echo "$$COVERAGE < 75" | bc -l) -eq 1 ]; then \
		echo "⚠️  Warning: Service layer coverage $$COVERAGE% is below 75% requirement"; \
	fi
	@echo "📈 Coverage report: scripts/coverage/service-coverage.html"

# API层测试覆盖率
test-api-coverage:
	@echo "📊 Running API layer tests with coverage..."
	@mkdir -p scripts/coverage
	@go test -v -race -coverprofile=scripts/coverage/api-coverage.out ./internal/api/...
	@go tool cover -html=scripts/coverage/api-coverage.out -o scripts/coverage/api-coverage.html
	@go tool cover -func=scripts/coverage/api-coverage.out | tail -1
	@echo "📈 Coverage report: scripts/coverage/api-coverage.html"

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

# 测试覆盖率检查
test-coverage-check:
	@echo "📊 Checking test coverage..."
	@go test -coverprofile=$(COVERAGE_FILE) ./... > /dev/null 2>&1
	@COVERAGE=$$(go tool cover -func=$(COVERAGE_FILE) | grep total | awk '{print $$3}' | sed 's/%//'); \
	if [ $$(echo "$$COVERAGE < 70" | bc) -eq 1 ]; then \
		echo "❌ Coverage $$COVERAGE% is below 70%"; \
		exit 1; \
	else \
		echo "✅ Coverage $$COVERAGE% meets the requirement"; \
	fi

# 代码分析
code-analysis:
	@echo "🔍 Running code analysis..."
	@echo "Running gofmt check..."
	@test -z $$(gofmt -l . | grep -v vendor) || (echo "Please run 'make fmt'"; exit 1)
	@echo "Running go vet..."
	@go vet ./...
	@echo "✅ Code analysis passed"

# Mock对象生成
mock-generate:
	@echo "🎭 Generating mock objects..."
	@if command -v mockgen > /dev/null; then \
		echo "Generating mocks..."; \
		mockgen -source=internal/repository/project_repository.go -destination=internal/repository/mocks/mock_project_repository.go; \
		mockgen -source=internal/repository/rule_repository.go -destination=internal/repository/mocks/mock_rule_repository.go; \
		echo "✅ Mocks generated"; \
	else \
		echo "⚠️  mockgen not installed, install with: go install github.com/golang/mock/mockgen@latest"; \
	fi

# 依赖升级检查
deps-upgrade:
	@echo "⬆️  Checking for dependency upgrades..."
	@go list -u -m all | grep '\['
	@echo "Run 'make deps-update' to upgrade"

# ════════════════════════════════════════════════════════
# Docker 相关
# ════════════════════════════════════════════════════════

# 构建 Docker 镜像
docker-build:
	@echo "🐳 Building Docker image (backend only)..."
	@docker build -t mockserver:$(VERSION) -t mockserver:latest .
	@echo "✅ Docker image built: mockserver:$(VERSION)"

# 构建包含前端的完整 Docker 镜像（多阶段构建）
docker-build-full:
	@echo "🐳 Building full-stack Docker image..."
	@if [ ! -f Dockerfile.fullstack ]; then \
		echo "❌ Dockerfile.fullstack not found"; \
		exit 1; \
	fi
	@docker build -f Dockerfile.fullstack \
		-t mockserver-fullstack:$(VERSION) \
		-t mockserver-fullstack:latest \
		--build-arg VERSION=$(VERSION) \
		--build-arg BUILD_TIME=$(BUILD_TIME) \
		--build-arg GIT_COMMIT=$(GIT_COMMIT) \
		.
	@echo "✅ Full-stack Docker image built: mockserver-fullstack:$(VERSION)"
	@echo "  - Frontend: web/frontend/dist (built inside container)"
	@echo "  - Backend:  mockserver binary with version info"

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

# 启动 MongoDB 容器
start-mongo:
	@echo "🍃 Starting MongoDB container..."
	@if docker ps -a --format '{{.Names}}' | grep -q '^mongodb$$'; then \
		echo "ℹ️  MongoDB container exists, checking status..."; \
		if docker ps --format '{{.Names}}' | grep -q '^mongodb$$'; then \
			echo "✅ MongoDB is already running"; \
		else \
			echo "🔄 Starting existing MongoDB container..."; \
			docker start mongodb || (echo "❌ Failed to start, removing broken container..." && docker rm -f mongodb && \
			docker run -d --name mongodb -p 27017:27017 -v mongodb_data:/data/db m.daocloud.io/docker.io/mongo:6.0); \
		fi; \
	else \
		echo "🚀 Creating and starting MongoDB container..."; \
		docker run -d --name mongodb -p 27017:27017 -v mongodb_data:/data/db m.daocloud.io/docker.io/mongo:6.0; \
	fi
	@echo "✅ MongoDB is running on localhost:27017"

# 停止 MongoDB 容器
stop-mongo:
	@echo "🛑 Stopping MongoDB container..."
	@docker stop mongodb 2>/dev/null || echo "⚠️  MongoDB container not running"
	@echo "✅ MongoDB stopped"

# 重启 MongoDB 容器
restart-mongo: stop-mongo start-mongo
	@echo "✅ MongoDB restarted"

# 连接 MongoDB Shell
mongo-shell:
	@echo "🐚 Connecting to MongoDB shell..."
	@docker exec -it mongodb mongosh

# 查看 MongoDB 日志
mongo-logs:
	@docker logs -f mongodb

# 启动后端服务（使用本地开发配置）
start-backend:
	@echo "🚀 Starting backend server with dev config..."
	@nohup go run ./cmd/mockserver/main.go -config config.dev.yaml > /tmp/mockserver.log 2>&1 &
	@echo $$! > /tmp/mockserver.pid
	@echo "⏳ Waiting for backend to start..."
	@sleep 5
	@if curl -s http://localhost:8080/api/v1/system/health > /dev/null 2>&1; then \
		echo "✅ Backend server started successfully"; \
		echo "📌 Admin API: http://localhost:8080/api/v1"; \
		echo "📌 Mock API: http://localhost:9090"; \
		echo "📋 Logs: tail -f /tmp/mockserver.log"; \
	else \
		echo "❌ Failed to start backend server"; \
		echo "📋 Last 20 lines of log:"; \
		tail -20 /tmp/mockserver.log 2>/dev/null || echo "No logs found"; \
		exit 1; \
	fi

# 停止后端服务
stop-backend:
	@echo "🛑 Stopping backend server..."
	@if [ -f /tmp/mockserver.pid ]; then \
		PID=$$(cat /tmp/mockserver.pid); \
		if ps -p $$PID > /dev/null 2>&1; then \
			kill $$PID 2>/dev/null || true; \
			echo "✅ Backend server stopped (PID: $$PID)"; \
		else \
			echo "ℹ️  Backend server process not found"; \
		fi; \
		rm -f /tmp/mockserver.pid; \
	else \
		echo "ℹ️  Backend server is not running"; \
	fi

# 启动前端开发服务器
start-frontend:
	@echo "🎨 Starting frontend dev server..."
	@cd web/frontend && \
		if [ ! -d "node_modules" ]; then \
			echo "📦 Installing frontend dependencies..."; \
			npm install; \
		fi && \
		nohup npm run dev > /tmp/frontend.log 2>&1 &
	@echo $$! > /tmp/frontend.pid
	@echo "⏳ Waiting for frontend to start..."
	@sleep 6
	@if curl -s http://localhost:5173 > /dev/null 2>&1; then \
		echo "✅ Frontend server started successfully"; \
		echo "📌 Frontend: http://localhost:5173"; \
		echo "📋 Logs: tail -f /tmp/frontend.log"; \
	else \
		echo "❌ Failed to start frontend server"; \
		echo "📋 Last 20 lines of log:"; \
		tail -20 /tmp/frontend.log 2>/dev/null || echo "No logs found"; \
		exit 1; \
	fi

# 停止前端服务
stop-frontend:
	@echo "🛑 Stopping frontend server..."
	@if [ -f /tmp/frontend.pid ]; then \
		PID=$$(cat /tmp/frontend.pid); \
		if ps -p $$PID > /dev/null 2>&1; then \
			kill $$PID 2>/dev/null || true; \
			echo "✅ Frontend server stopped (PID: $$PID)"; \
		else \
			echo "ℹ️  Frontend server process not found"; \
		fi; \
		rm -f /tmp/frontend.pid; \
	else \
		echo "ℹ️  Frontend server is not running"; \
	fi

# 启动全栈应用（MongoDB + 后端 + 前端）
start-all:
	@echo "🚀 Starting full stack application..."
	@echo ""
	@echo "Step 1/3: Starting MongoDB..."
	@make start-mongo
	@echo ""
	@echo "Step 2/3: Starting Backend..."
	@sleep 3
	@make start-backend
	@echo ""
	@echo "Step 3/3: Starting Frontend..."
	@make start-frontend
	@echo ""
	@echo "═══════════════════════════════════════════════════════"
	@echo "✅ Full stack application is running!"
	@echo "═══════════════════════════════════════════════════════"
	@echo ""
	@echo "🌐 Access URLs:"
	@echo "  Frontend:   http://localhost:5173"
	@echo "  Admin API:  http://localhost:8080/api/v1"
	@echo "  Mock API:   http://localhost:9090"
	@echo "  MongoDB:    mongodb://localhost:27017"
	@echo ""
	@echo "📋 View Logs:"
	@echo "  Backend:    tail -f /tmp/mockserver.log"
	@echo "  Frontend:   tail -f /tmp/frontend.log"
	@echo "  MongoDB:    make mongo-logs"
	@echo ""
	@echo "🛑 Stop All:"
	@echo "  make stop-all"
	@echo "═══════════════════════════════════════════════════════"

# 停止全栈应用
stop-all:
	@echo "🛑 Stopping full stack application..."
	@make stop-frontend 2>/dev/null || true
	@make stop-backend 2>/dev/null || true
	@pkill -f "vite" 2>/dev/null || true
	@pkill -f "mockserver/main.go" 2>/dev/null || true
	@lsof -ti:5173 | xargs kill -9 2>/dev/null || true
	@lsof -ti:8080 | xargs kill -9 2>/dev/null || true
	@lsof -ti:9090 | xargs kill -9 2>/dev/null || true
	@make stop-mongo 2>/dev/null || true
	@rm -f /tmp/mockserver.pid /tmp/frontend.pid /tmp/mockserver.log /tmp/frontend.log
	@echo "✅ Full stack application stopped"

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
# 快速验证（格式化+检查+构建+测试）
verify: fmt vet lint build test-unit
	@echo "✅ Quick verification complete!"

# 质量检查 (快捷命令)
qa: fmt vet lint test-unit
	@echo "✅ Quality assurance checks passed!"

# 推送前检查 (包含集成测试)
pre-push: qa
	@echo "🚀 Running integration tests..."
	@if [ -f ./tests/integration/e2e_test.sh ]; then \
		chmod +x ./tests/integration/e2e_test.sh; \
		echo "✅ Pre-push checks passed!"; \
	else \
		echo "⚠️  Integration tests not found, skipping..."; \
	fi

# 预提交检查 (别名 qa)
pre-commit: qa
	@echo "✅ Pre-commit checks passed!"

# 命令别名
t: test
c: test-coverage

# 快速启动开发环境
dev-env: start-mongo
	@echo "✅ Development environment ready!"
	@echo "📌 MongoDB: localhost:27017"
	@echo "🚀 Run 'make run' or 'make dev' to start the server"

# 清理开发环境
clean-env: stop-mongo
	@echo "🧽 Cleaning development environment..."
	@docker volume rm mongodb_data 2>/dev/null || true
	@echo "✅ Environment cleaned"

# 显示版本信息
version:
	@echo "Version:    $(VERSION)"
	@echo "Build Time: $(BUILD_TIME)"
	@echo "Git Commit: $(GIT_COMMIT)"
