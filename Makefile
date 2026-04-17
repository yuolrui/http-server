# HTTP Mock Server Makefile

VERSION ?= 1.0.0
OUTPUT_DIR = build
MAIN_PATH = ./cmd/http-server

# 默认目标：编译当前平台
.PHONY: build
build:
	go build -ldflags "-s -w" -o http-server.exe $(MAIN_PATH)

# 清理构建目录
.PHONY: clean
clean:
	rm -rf $(OUTPUT_DIR)
	rm -f http-server.exe

# 创建构建目录
$(OUTPUT_DIR):
	mkdir -p $(OUTPUT_DIR)

# 交叉编译所有平台
.PHONY: build-all
build-all: clean $(OUTPUT_DIR)
	@echo "开始交叉编译..."
	GOOS=windows GOARCH=amd64 go build -ldflags "-s -w" -o $(OUTPUT_DIR)/http-server-$(VERSION)-windows-amd64.exe $(MAIN_PATH)
	GOOS=windows GOARCH=arm64 go build -ldflags "-s -w" -o $(OUTPUT_DIR)/http-server-$(VERSION)-windows-arm64.exe $(MAIN_PATH)
	GOOS=linux GOARCH=amd64 go build -ldflags "-s -w" -o $(OUTPUT_DIR)/http-server-$(VERSION)-linux-amd64 $(MAIN_PATH)
	GOOS=linux GOARCH=arm64 go build -ldflags "-s -w" -o $(OUTPUT_DIR)/http-server-$(VERSION)-linux-arm64 $(MAIN_PATH)
	GOOS=darwin GOARCH=amd64 go build -ldflags "-s -w" -o $(OUTPUT_DIR)/http-server-$(VERSION)-darwin-amd64 $(MAIN_PATH)
	GOOS=darwin GOARCH=arm64 go build -ldflags "-s -w" -o $(OUTPUT_DIR)/http-server-$(VERSION)-darwin-arm64 $(MAIN_PATH)
	@echo "编译完成！"
	@ls -lh $(OUTPUT_DIR)

# 单平台编译
.PHONY: build-windows
build-windows:
	GOOS=windows GOARCH=amd64 go build -ldflags "-s -w" -o http-server.exe $(MAIN_PATH)

.PHONY: build-linux
build-linux:
	GOOS=linux GOARCH=amd64 go build -ldflags "-s -w" -o http-server $(MAIN_PATH)

.PHONY: build-darwin
build-darwin:
	GOOS=darwin GOARCH=amd64 go build -ldflags "-s -w" -o http-server $(MAIN_PATH)

# 运行服务器
.PHONY: run
run:
	go run $(MAIN_PATH)

# 运行测试
.PHONY: test
test:
	go test -v ./...

# 格式化代码
.PHONY: fmt
fmt:
	go fmt ./...

# 静态检查
.PHONY: vet
vet:
	go vet ./...

# 帮助
.PHONY: help
help:
	@echo "可用命令:"
	@echo "  make build          - 编译当前平台"
	@echo "  make build-all      - 交叉编译所有平台"
	@echo "  make build-windows  - 编译 Windows amd64"
	@echo "  make build-linux    - 编译 Linux amd64"
	@echo "  make build-darwin   - 编译 macOS amd64"
	@echo "  make clean          - 清理构建产物"
	@echo "  make run            - 运行服务器"
	@echo "  make test           - 运行测试"
	@echo "  make fmt            - 格式化代码"
	@echo "  make vet            - 静态检查"
	@echo "  make help           - 显示帮助"
	@echo ""
	@echo "可选参数:"
	@echo "  VERSION=1.0.0       - 设置版本号"