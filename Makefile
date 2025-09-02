.PHONY: help
help:
	@echo "可用命令:"
	@echo "  make install    - 安装 buf 和依赖"
	@echo "  make update     - 更新 buf 依赖"
	@echo "  make lint       - 检查 proto 文件"
	@echo "  make generate   - 生成代码"
	@echo "  make build      - 构建插件"
	@echo "  make clean      - 清理生成的文件"
	@echo "  make run        - 运行示例服务器"

# 安装 buf CLI
.PHONY: install
install:
	@echo "安装 buf..."
	go install github.com/bufbuild/buf/cmd/buf@latest
	@echo "安装 Go 依赖..."
	go mod download
	@echo "更新 buf 依赖..."
	buf mod update

# 更新 buf 依赖
.PHONY: update
update:
	buf mod update

# 检查 proto 文件
.PHONY: lint
lint:
	buf lint

# 构建自定义插件
.PHONY: build-plugin
build-plugin:
	@echo "构建 protoc-gen-gin 插件..."
	go build -o cmd/protoc-gen-gin/protoc-gen-gin ./cmd/protoc-gen-gin

# 生成代码
.PHONY: generate
generate: build-plugin
	@echo "生成代码..."
	buf generate

# 清理生成的文件
.PHONY: clean
clean:
	rm -rf gen/
	rm -f cmd/protoc-gen-gin/protoc-gen-gin

# 运行示例服务器
.PHONY: run
run: generate
	go run examples/server/main.go

# 调试模式
.PHONY: debug
debug:
	go run cmd/protoc-gen-gin/main_debug.go debug

# 格式化 proto 文件
.PHONY: format
format:
	buf format -w

# 破坏性变更检查
.PHONY: breaking
breaking:
	buf breaking --against '.git#branch=main'