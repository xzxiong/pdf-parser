.PHONY: build test clean run install help

# 变量定义
BINARY_NAME=pdf-parser
GO=go
GOFLAGS=-v

# 默认目标
all: build

# 编译项目
build:
	@echo "正在编译..."
	$(GO) build $(GOFLAGS) -o $(BINARY_NAME) .
	@echo "编译完成: $(BINARY_NAME)"

# 运行测试
test:
	@echo "运行测试..."
	$(GO) test $(GOFLAGS) ./...

# 安装依赖
install:
	@echo "安装依赖..."
	$(GO) mod download
	$(GO) mod tidy

# 运行程序（需要提供 PDF 文件路径）
run:
	@if [ -z "$(PDF)" ]; then \
		echo "用法: make run PDF=<文件路径>"; \
		exit 1; \
	fi
	$(GO) run . $(PDF)

# 清理编译产物
clean:
	@echo "清理编译产物..."
	@rm -f $(BINARY_NAME)
	@echo "清理完成"

# 显示帮助信息
help:
	@echo "可用的命令:"
	@echo "  make build    - 编译项目"
	@echo "  make test     - 运行测试"
	@echo "  make install  - 安装依赖"
	@echo "  make run PDF=<文件路径> - 运行程序"
	@echo "  make clean    - 清理编译产物"
	@echo "  make help     - 显示此帮助信息"
