# EiTec VPN 项目构建脚本

# 变量定义
BINARY_DIR=bin
SERVER_BINARY=$(BINARY_DIR)/eitec-vpn-server
MODULE_BINARY=$(BINARY_DIR)/eitec-vpn-module
GO_VERSION=1.21
LDFLAGS=-ldflags "-X main.version=$(shell git describe --tags --always --dirty) -X main.buildTime=$(shell date -u +%Y%m%d.%H%M%S)"

# 默认目标
.PHONY: all
all: clean deps build

# 安装依赖
.PHONY: deps
deps:
	@echo "下载依赖包..."
	go mod download
	go mod tidy

# 构建所有应用
.PHONY: build
build: build-server build-module

# 构建服务器端
.PHONY: build-server
build-server:
	@echo "构建服务器端应用..."
	@mkdir -p $(BINARY_DIR)
	CGO_ENABLED=1 GOOS=linux GOARCH=amd64 go build $(LDFLAGS) -o $(SERVER_BINARY) cmd/server/main.go

# 构建模块端
.PHONY: build-module
build-module:
	@echo "构建模块端应用..."
	@mkdir -p $(BINARY_DIR)
	CGO_ENABLED=1 GOOS=linux GOARCH=amd64 go build $(LDFLAGS) -o $(MODULE_BINARY) cmd/module/main.go

# 本地构建 (用于开发测试)
.PHONY: build-local
build-local:
	@echo "本地构建..."
	@mkdir -p $(BINARY_DIR)
	go build $(LDFLAGS) -o $(SERVER_BINARY) cmd/server/main.go
	go build $(LDFLAGS) -o $(MODULE_BINARY) cmd/module/main.go

# ARM64构建 (树莓派)
.PHONY: build-arm64
build-arm64:
	@echo "构建ARM64版本 (树莓派)..."
	@mkdir -p $(BINARY_DIR)
	CGO_ENABLED=1 GOOS=linux GOARCH=arm64 go build $(LDFLAGS) -o $(MODULE_BINARY)-arm64 cmd/module/main.go

# 清理构建文件
.PHONY: clean
clean:
	@echo "清理构建文件..."
	rm -rf $(BINARY_DIR)
	rm -rf data/*.db
	rm -rf logs/*.log

# 运行测试
.PHONY: test
test:
	@echo "运行测试..."
	go test -v ./...

# 代码检查
.PHONY: lint
lint:
	@echo "代码检查..."
	golangci-lint run

# 格式化代码
.PHONY: fmt
fmt:
	@echo "格式化代码..."
	go fmt ./...

# 启动服务器端 (开发模式)
.PHONY: run-server
run-server: build-local
	@echo "启动服务器端..."
	./$(SERVER_BINARY) --config configs/server.yaml

# 启动模块端 (开发模式)
.PHONY: run-module
run-module: build-local
	@echo "启动模块端..."
	./$(MODULE_BINARY) --config configs/module.yaml

# 安装WireGuard (Ubuntu/Debian)
.PHONY: install-wireguard
install-wireguard:
	@echo "安装WireGuard..."
	sudo apt update
	sudo apt install -y wireguard wireguard-tools

# 生成WireGuard密钥对
.PHONY: generate-keys
generate-keys:
	@echo "生成WireGuard密钥对..."
	wg genkey | tee server.key | wg pubkey > server.pub
	@echo "服务器私钥: $$(cat server.key)"
	@echo "服务器公钥: $$(cat server.pub)"

# 创建systemd服务
.PHONY: install-service
install-service:
	@echo "安装systemd服务..."
	sudo cp scripts/eitec-vpn-server.service /etc/systemd/system/
	sudo cp scripts/eitec-vpn-module.service /etc/systemd/system/
	sudo systemctl daemon-reload

# Docker构建
.PHONY: docker-build
docker-build:
	@echo "构建Docker镜像..."
	docker build -t eitec-vpn-server -f scripts/docker/Dockerfile.server .
	docker build -t eitec-vpn-module -f scripts/docker/Dockerfile.module .

# 开发环境设置
.PHONY: dev-setup
dev-setup:
	@echo "设置开发环境..."
	mkdir -p data logs
	touch logs/server.log logs/module.log

# 显示版本信息
.PHONY: version
version:
	@echo "Go版本: $(shell go version)"
	@echo "Git版本: $(shell git describe --tags --always --dirty)"
	@echo "构建时间: $(shell date -u +%Y%m%d.%H%M%S)"

# 帮助信息
.PHONY: help
help:
	@echo "EiTec VPN 构建脚本"
	@echo ""
	@echo "可用命令:"
	@echo "  all              - 清理、下载依赖并构建所有应用"
	@echo "  deps             - 下载Go依赖包"
	@echo "  build            - 构建所有应用"
	@echo "  build-server     - 构建服务器端应用"
	@echo "  build-module     - 构建模块端应用"
	@echo "  build-local      - 本地构建 (开发用)"
	@echo "  build-arm64      - 构建ARM64版本 (树莓派)"
	@echo "  clean            - 清理构建文件"
	@echo "  test             - 运行测试"
	@echo "  lint             - 代码检查"
	@echo "  fmt              - 格式化代码"
	@echo "  run-server       - 启动服务器端"
	@echo "  run-module       - 启动模块端"
	@echo "  install-wireguard- 安装WireGuard"
	@echo "  generate-keys    - 生成WireGuard密钥对"
	@echo "  install-service  - 安装systemd服务"
	@echo "  docker-build     - 构建Docker镜像"
	@echo "  dev-setup        - 设置开发环境"
	@echo "  version          - 显示版本信息"
	@echo "  help             - 显示此帮助信息" 