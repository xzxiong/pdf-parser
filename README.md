# PDF 标题解析工具

这是一个用 Golang 编写的 PDF 解析工具，可以提取 PDF 文档中的标题文字和标题级别。

## 功能

- 提取 PDF 文档的标题（基于 PDF 大纲/书签）
- 获取标题的层级结构
- 显示标题所在的页码

## 安装依赖

```bash
go mod download
```

## 使用方法

```bash
go run main.go <PDF文件路径>
```

### 示例

```bash
go run main.go example.pdf
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
go build -o pdf-parser
./pdf-parser example.pdf
```

## 注意事项

- 此工具依赖 PDF 文件中的大纲（Outlines/Bookmarks）信息
- 如果 PDF 没有大纲结构，将无法提取标题
- unipdf 库可能需要许可证用于商业用途

## 依赖库

- [unipdf](https://github.com/unidoc/unipdf) - 纯 Go 实现的 PDF 处理库
