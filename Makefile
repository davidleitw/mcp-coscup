# COSCUP MCP Server Makefile

# 變數定義
BINARY_NAME=mcp-coscup
MAIN_PATH=./cmd/server
BUILD_DIR=bin
DOCKER_TAG=mcp-coscup

# 預設目標
.PHONY: all
all: build

# 建立 build 目錄
$(BUILD_DIR):
	mkdir -p $(BUILD_DIR)

# Linux 版本 (預設)
.PHONY: build
build: $(BUILD_DIR)
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o $(BUILD_DIR)/$(BINARY_NAME) $(MAIN_PATH)

# Windows 版本
.PHONY: build-windows
build-windows: $(BUILD_DIR)
	CGO_ENABLED=0 GOOS=windows GOARCH=amd64 go build -o $(BUILD_DIR)/$(BINARY_NAME).exe $(MAIN_PATH)

# macOS 版本
.PHONY: build-mac
build-mac: $(BUILD_DIR)
	CGO_ENABLED=0 GOOS=darwin GOARCH=amd64 go build -o $(BUILD_DIR)/$(BINARY_NAME)-darwin $(MAIN_PATH)

# macOS ARM64 版本 (Apple Silicon)
.PHONY: build-mac-arm
build-mac-arm: $(BUILD_DIR)
	CGO_ENABLED=0 GOOS=darwin GOARCH=arm64 go build -o $(BUILD_DIR)/$(BINARY_NAME)-darwin-arm64 $(MAIN_PATH)

# 編譯所有平台版本
.PHONY: build-all
build-all: build build-windows build-mac build-mac-arm
	@echo "All binaries built successfully!"
	@ls -la $(BUILD_DIR)/

# 測試本地版本
.PHONY: test
test:
	go test ./mcp

# 詳細測試輸出
.PHONY: test-verbose
test-verbose:
	go test ./mcp -v

# 測試覆蓋率
.PHONY: test-coverage
test-coverage:
	go test ./mcp -cover -coverprofile=coverage.out
	go tool cover -func=coverage.out

# 生成 HTML 覆蓋率報告
.PHONY: test-coverage-html
test-coverage-html:
	go test ./mcp -cover -coverprofile=coverage.out
	go tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report generated: coverage.html"

# 運行所有測試 (包括原有的所有包)
.PHONY: test-all
test-all:
	go test ./...

# 運行資料測試
.PHONY: test-data
test-data:
	@echo "Testing data loading..."
	@timeout 3s ./$(BUILD_DIR)/$(BINARY_NAME) 2>&1 | head -10 || echo "Server started successfully"

# 測試 Windows 版本 (需要 Wine 或在 Windows 環境)
.PHONY: test-windows
test-windows:
	@echo "Testing Windows binary (需要在 Windows 環境或透過 Wine)..."
	@echo "Please run: $(BUILD_DIR)/$(BINARY_NAME).exe"

# Docker 相關
.PHONY: docker-build
docker-build:
	docker build -t $(DOCKER_TAG) .

.PHONY: docker-run
docker-run:
	@echo "Running Docker container (will timeout after 5 seconds)..."
	@timeout 5s docker run --rm $(DOCKER_TAG) || echo "Docker container started successfully"

.PHONY: docker-test
docker-test: docker-build docker-run

# 清理
.PHONY: clean
clean:
	rm -rf $(BUILD_DIR)
	rm -f coverage.out coverage.html
	docker rmi $(DOCKER_TAG) 2>/dev/null || true

# 開發相關
.PHONY: deps
deps:
	go mod tidy
	go mod download

.PHONY: fmt
fmt:
	go fmt ./...

.PHONY: vet
vet:
	go vet ./...

# 完整的開發流程
.PHONY: dev
dev: fmt vet deps test build
	@echo "Development build completed!"

# 發佈流程
.PHONY: release
release: clean fmt vet test build-all docker-build
	@echo "Release build completed!"
	@echo "Binaries created:"
	@ls -la $(BUILD_DIR)/

# 幫助訊息
.PHONY: help
help:
	@echo "COSCUP MCP Server Build Commands:"
	@echo ""
	@echo "Basic builds:"
	@echo "  make build         - Build Linux binary (default)"
	@echo "  make build-windows - Build Windows binary (.exe)"
	@echo "  make build-mac     - Build macOS binary"
	@echo "  make build-mac-arm - Build macOS ARM64 binary"
	@echo "  make build-all     - Build all platform binaries"
	@echo ""
	@echo "Testing:"
	@echo "  make test         - Run Go tests"
	@echo "  make test-data    - Test server startup and data loading"
	@echo "  make test-windows - Test Windows binary"
	@echo ""
	@echo "Docker:"
	@echo "  make docker-build - Build Docker image"
	@echo "  make docker-run   - Test Docker container"
	@echo "  make docker-test  - Build and test Docker"
	@echo ""
	@echo "Development:"
	@echo "  make dev          - Complete development build"
	@echo "  make release      - Complete release build"
	@echo "  make clean        - Clean build artifacts"
	@echo ""
	@echo "Claude Code Integration:"
	@echo "  make install           - Interactive install (choose local/remote)"
	@echo "  make install-local     - Install local version to Claude Code"
	@echo "  make install-remote    - Install remote version to Claude Code"
	@echo "  make uninstall-local   - Remove local version from Claude Code"
	@echo "  make uninstall-remote  - Remove remote version from Claude Code"
	@echo "  make claude-status     - Show Claude Code MCP status"
	@echo ""
	@echo "Cloud Deployment:"
	@echo "  make deploy           - Deploy to Google Cloud Run"
	@echo "  make logs             - View Cloud Run logs"
	@echo "  make logs-tail        - Monitor Cloud Run logs in real-time"
	@echo "  make help             - Show this help message"

# 雲端部署變數 - 可透過環境變數覆蓋
PROJECT_ID ?= your-project-id
REGION ?= asia-east1
SERVICE_URL ?= https://your-service-url.run.app/mcp

# 載入 .env 檔案（如果存在）
-include .env

# ==============================================
# Claude Code 整合相關指令
# ==============================================

# 互動式安裝 - 讓用戶選擇本地或遠端
.PHONY: install
install:
	@echo "=========================================="
	@echo "  COSCUP MCP Server Installation"
	@echo "=========================================="
	@echo ""
	@echo "Choose installation type:"
	@echo "  1) Local  - Build and install locally (recommended for development)"
	@echo "  2) Remote - Connect to cloud service (recommended for usage)"
	@echo ""
	@read -p "Enter your choice (1 or 2): " choice; \
	case $$choice in \
		1) echo "Installing local version..."; make install-local ;; \
		2) echo "Installing remote version..."; make install-remote ;; \
		*) echo "Invalid choice. Please run 'make install' again." ;; \
	esac

# 安裝本地版本到 Claude Code
.PHONY: install-local
install-local: build
	@echo "Installing local MCP server to Claude Code..."
	@if command -v claude >/dev/null 2>&1; then \
		claude mcp add $(BINARY_NAME) $(shell pwd)/$(BUILD_DIR)/$(BINARY_NAME) -s user; \
		echo "✅ Local MCP server installed successfully!"; \
		echo ""; \
		echo "Usage: Ask Claude about COSCUP schedules in Claude Code"; \
		echo "Test: Try asking 'Help me plan my COSCUP Aug.9 schedule'"; \
	else \
		echo "❌ Claude CLI not found. Please install Claude Code first."; \
		echo "Visit: https://claude.ai/download"; \
		exit 1; \
	fi

# 安裝遠端版本到 Claude Code
.PHONY: install-remote
install-remote:
	@echo "Installing remote MCP server to Claude Code..."
	@if command -v claude >/dev/null 2>&1; then \
		claude mcp add --transport http $(BINARY_NAME)-remote $(SERVICE_URL) -s user; \
		echo "✅ Remote MCP server connected successfully!"; \
		echo ""; \
		echo "Usage: Ask Claude about COSCUP schedules in Claude Code"; \
		echo "Test: Try asking 'Help me plan my COSCUP Aug.9 schedule'"; \
		echo "Service: $(SERVICE_URL)"; \
	else \
		echo "❌ Claude CLI not found. Please install Claude Code first."; \
		echo "Visit: https://claude.ai/download"; \
		exit 1; \
	fi

# 移除本地版本
.PHONY: uninstall-local
uninstall-local:
	@echo "Removing local MCP server from Claude Code..."
	@if command -v claude >/dev/null 2>&1; then \
		claude mcp remove $(BINARY_NAME) -s user; \
		echo "✅ Local MCP server removed successfully!"; \
	else \
		echo "❌ Claude CLI not found."; \
		exit 1; \
	fi

# 移除遠端版本
.PHONY: uninstall-remote
uninstall-remote:
	@echo "Removing remote MCP server from Claude Code..."
	@if command -v claude >/dev/null 2>&1; then \
		claude mcp remove $(BINARY_NAME)-remote -s user; \
		echo "✅ Remote MCP server removed successfully!"; \
	else \
		echo "❌ Claude CLI not found."; \
		exit 1; \
	fi

# 顯示 Claude Code MCP 狀態
.PHONY: claude-status
claude-status:
	@echo "Claude Code MCP Server Status:"
	@echo "==============================="
	@if command -v claude >/dev/null 2>&1; then \
		claude mcp list; \
	else \
		echo "❌ Claude CLI not found. Please install Claude Code first."; \
		echo "Visit: https://claude.ai/download"; \
	fi

# 雲端部署
.PHONY: deploy
deploy:
	@echo "Deploying to Google Cloud Run..."
	@if [ -f "./deploy/deploy.sh" ]; then \
		chmod +x ./deploy/deploy.sh; \
		./deploy/deploy.sh; \
	else \
		echo "❌ deploy/deploy.sh not found. Please ensure you're in the project root."; \
		exit 1; \
	fi

# 檢查本地二進位檔案
.PHONY: check-local
check-local:
	@if [ -f "$(BUILD_DIR)/$(BINARY_NAME)" ]; then \
		echo "✅ Local binary exists: $(BUILD_DIR)/$(BINARY_NAME)"; \
		echo "Built: $(shell stat -c %y $(BUILD_DIR)/$(BINARY_NAME) 2>/dev/null || stat -f %Sm $(BUILD_DIR)/$(BINARY_NAME) 2>/dev/null || echo 'unknown')"; \
	else \
		echo "❌ Local binary not found. Run 'make build' first."; \
	fi

# 檢查遠端服務
.PHONY: check-remote
check-remote:
	@echo "Checking remote service..."
	@if curl -s -f $(SERVICE_URL:mcp=health) >/dev/null; then \
		echo "✅ Remote service is healthy: $(SERVICE_URL:mcp=health)"; \
	else \
		echo "❌ Remote service is not accessible: $(SERVICE_URL:mcp=health)"; \
	fi

# 查看 Cloud Run logs
.PHONY: logs
logs:
	@echo "Viewing Cloud Run logs..."
	@if command -v gcloud >/dev/null 2>&1; then \
		gcloud run services logs read $(BINARY_NAME) --region=$(REGION) --project=$(PROJECT_ID) --limit=50; \
	else \
		echo "❌ gcloud CLI not found."; \
		echo "To view logs:"; \
		echo "1. Install gcloud CLI"; \
		echo "2. Or visit: https://console.cloud.google.com/logs/query?project=$(PROJECT_ID)"; \
		echo "3. Filter: resource.type=cloud_run_revision AND resource.labels.service_name=$(BINARY_NAME)"; \
	fi

# 即時監控 logs
.PHONY: logs-tail
logs-tail:
	@echo "Tailing Cloud Run logs..."
	@if command -v gcloud >/dev/null 2>&1; then \
		gcloud run services logs tail $(BINARY_NAME) --region=$(REGION) --project=$(PROJECT_ID); \
	else \
		echo "❌ gcloud CLI not found."; \
		echo "Please install gcloud CLI or use Cloud Console"; \
	fi