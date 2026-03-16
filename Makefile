.PHONY: build build-all clean test lint

# 版本信息
VERSION ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
COMMIT ?= $(shell git rev-parse --short HEAD 2>/dev/null || echo "none")
DATE ?= $(shell date -u +"%Y-%m-%dT%H:%M:%SZ")
LDFLAGS := -ldflags "-s -w -X main.version=$(VERSION) -X main.commit=$(COMMIT) -X main.date=$(DATE)"

# 输出目录
DIST_DIR := dist

# 默认构建当前平台
build:
	go build $(LDFLAGS) -o github-buddy ./cmd/github-buddy/

# 交叉编译三平台
build-all: clean
	@mkdir -p $(DIST_DIR)
	@echo "构建 Linux amd64 ..."
	GOOS=linux GOARCH=amd64 go build $(LDFLAGS) -o $(DIST_DIR)/github-buddy-linux-amd64 ./cmd/github-buddy/
	@echo "构建 Linux arm64 ..."
	GOOS=linux GOARCH=arm64 go build $(LDFLAGS) -o $(DIST_DIR)/github-buddy-linux-arm64 ./cmd/github-buddy/
	@echo "构建 macOS amd64 ..."
	GOOS=darwin GOARCH=amd64 go build $(LDFLAGS) -o $(DIST_DIR)/github-buddy-darwin-amd64 ./cmd/github-buddy/
	@echo "构建 macOS arm64 ..."
	GOOS=darwin GOARCH=arm64 go build $(LDFLAGS) -o $(DIST_DIR)/github-buddy-darwin-arm64 ./cmd/github-buddy/
	@echo "构建 Windows amd64 ..."
	GOOS=windows GOARCH=amd64 go build $(LDFLAGS) -o $(DIST_DIR)/github-buddy-windows-amd64.exe ./cmd/github-buddy/
	@echo "✅ 所有平台构建完成，输出目录: $(DIST_DIR)/"

# 清理构建产物
clean:
	rm -rf $(DIST_DIR) github-buddy github-buddy.exe

# 运行测试
test:
	go test ./... -v -count=1

# 运行测试（简洁输出）
test-short:
	go test ./... -count=1

# 代码格式检查
lint:
	go vet ./...
	gofmt -l .
