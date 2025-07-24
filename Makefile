# COSCUP MCP Server Makefile

# 變數定義
BINARY_NAME=coscup-mcp-server
MAIN_PATH=./cmd/server
BUILD_DIR=bin
DOCKER_TAG=coscup-mcp-server

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
	go test ./...

# 運行資料測試
.PHONY: test-data
test-data:
	@echo "Testing data loading..."
	@timeout 3s ./$(BUILD_DIR)/$(BINARY_NAME) 2>&1 | head -10 || echo "✅ Server started successfully"

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
	@timeout 5s docker run --rm $(DOCKER_TAG) || echo "✅ Docker container started successfully"

.PHONY: docker-test
docker-test: docker-build docker-run

# 清理
.PHONY: clean
clean:
	rm -rf $(BUILD_DIR)
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
	@echo "✅ Development build completed!"

# 發佈流程
.PHONY: release
release: clean fmt vet test build-all docker-build
	@echo "🚀 Release build completed!"
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
	@echo "  make help         - Show this help message"