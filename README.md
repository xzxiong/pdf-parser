# PDF 标题解析工具

这是一个用 Golang 编写的 PDF 解析工具，可以提取 PDF 文档中的标题文字和标题级别。

## 功能

- 从 PDF 大纲（目录/书签）中提取标题
- 从文本内容中识别标题格式（如 '1. ', '1.1 ', '2.3.4 ' 等）
- 获取标题的层级结构
- 显示标题所在的页码

## 安装依赖

```bash
go mod download
```

## 使用方法

```bash
# 基本用法
go run main.go <PDF文件路径>

# 或使用编译后的程序
./pdf-parser <PDF文件路径>

# 调试模式（显示更多信息）
./pdf-parser -d <PDF文件路径>

# 显示帮助
./pdf-parser -h
```

### 示例

```bash
go run main.go example.pdf
./pdf-parser document.pdf
./pdf-parser -d document.pdf
```

### 输出示例

```
找到 5 个标题:

级别 1: 第一章 引言 (第 1 页)
  级别 2: 1.1 背景 (第 2 页)
  级别 2: 1.2 目标 (第 3 页)
级别 1: 第二章 方法 (第 5 页)
  级别 2: 2.1 实验设计 (第 6 页)
```

## 编译

```bash
make build
# 或
go build -o pdf-parser
```

## 测试

项目包含完整的单元测试，用于验证 PDF 标题提取功能。

### 运行测试

```bash
# 运行所有测试
make test

# 或使用 go test
go test -v

# 运行测试并生成覆盖率报告
make test-coverage

# 运行基准测试
make bench
```

### 测试内容

测试文件 `main_test.go` 包含以下测试用例：

1. **TestExtractHeadings135Title** - 测试从 135_title.pdf 提取标题
   - 验证标题数量
   - 检查关键标题是否存在
   - 验证标题层级分布
   - 检查页码范围

2. **TestExtractHeadingsFromOutline** - 测试大纲提取功能
   - 验证提取的标题数量
   - 输出所有标题用于调试

3. **TestHeadingStructure** - 测试标题结构合理性
   - 检查级别跳跃
   - 验证标题非空
   - 检查级别范围

4. **TestMergeHeadings** - 测试标题合并去重功能

5. **BenchmarkExtractHeadings** - 性能基准测试

### 测试结果示例

```
=== RUN   TestExtractHeadings135Title
    main_test.go:22: 成功提取 56 个标题
    main_test.go:81: ✓ 找到关键标题: 1.Project background and objectives (级别 1, 第 2 页)
    main_test.go:81: ✓ 找到关键标题: 2.Overall scheme design (级别 1, 第 3 页)
    main_test.go:96: 标题层级分布:
    main_test.go:99:   级别 1: 6 个标题
    main_test.go:99:   级别 2: 24 个标题
    main_test.go:99:   级别 3: 26 个标题
--- PASS: TestExtractHeadings135Title (0.14s)
PASS
```

## 注意事项

### 标题提取方式

工具使用两种方式提取标题：

1. **从 PDF 大纲提取**（主要方式）
   - 依赖 PDF 文件中的大纲（Outlines/Bookmarks）信息
   - 如果 PDF 没有大纲结构，将无法通过此方式提取标题

2. **从文本内容提取**（辅助方式）
   - 识别符合特定格式的文本行作为标题
   - 支持的格式：`1. 标题`、`1.1 标题`、`2.3.4 标题` 等
   - 注意：unipdf 免费版本对文本提取有限制

### 限制

- unipdf 库的免费版本可能有功能限制
- 商业用途需要购买 unipdf 许可证
- 某些 PDF 文件的标题可能无法被正确识别（取决于 PDF 的内部结构）
- 如果标题跨越多行或使用特殊格式，可能无法被识别

### 解决方案

如果发现某些标题缺失：

1. 检查 PDF 文件是否包含大纲（在 PDF 阅读器中查看目录/书签）
2. 如果大纲不完整，可能需要：
   - 使用 PDF 编辑工具重新生成大纲
   - 使用商业版 PDF 处理库
   - 使用其他 PDF 文本提取工具（如 pdftotext）配合本工具使用

## 依赖库

- [unipdf](https://github.com/unidoc/unipdf) - 纯 Go 实现的 PDF 处理库

## 许可证

本工具使用 unipdf 库，该库对商业用途需要许可证。详见：https://unidoc.io
