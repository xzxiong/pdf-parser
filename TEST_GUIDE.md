# PDF 标题解析工具 - 测试指南

## 概述

本项目包含完整的单元测试，用于验证从 PDF 文件中提取标题的功能。测试文件针对 `135_title.pdf` 进行了全面的验证。

## 快速开始

### 运行所有测试

```bash
make test
```

### 运行测试并查看详细输出

```bash
go test -v
```

### 生成测试覆盖率报告

```bash
make test-coverage
# 会生成 coverage.html 文件，可在浏览器中打开查看
```

### 运行基准测试

```bash
make bench
```

## 测试用例说明

### 1. TestExtractHeadings135Title

**目的**: 验证从 135_title.pdf 提取标题的核心功能

**测试内容**:
- ✓ 验证能够成功提取标题
- ✓ 检查标题数量是否合理（应该 > 0）
- ✓ 验证关键标题是否存在：
  - "1.Project background and objectives"
  - "2.Overall scheme design"
  - "3.Core functional modules"
  - "5.Attachment"
  - "6.总结与展望"
- ✓ 验证标题层级分布（1级、2级、3级标题的数量）
- ✓ 检查页码范围是否合理（1-10页）

**预期结果**:
```
成功提取 56 个标题
标题层级分布:
  级别 1: 6 个标题
  级别 2: 24 个标题
  级别 3: 26 个标题
```

### 2. TestExtractHeadingsFromOutline

**目的**: 详细验证大纲提取功能

**测试内容**:
- ✓ 验证提取的标题数量（至少 50 个）
- ✓ 输出所有提取的标题用于调试和验证
- ✓ 检查标题的完整性

**输出示例**:
```
1. 级别 1: 1.Project background and objectives (第 2 页)
2.   级别 2: 1.1.DCC：Document Control Center (第 2 页)
3.   级别 2: 1.2.DCN：Document Change Notice... (第 2 页)
...
```

### 3. TestHeadingStructure

**目的**: 验证标题结构的合理性

**测试内容**:
- ✓ 检查标题级别跳跃是否合理（不应该从1级直接跳到3级）
- ✓ 验证所有标题非空
- ✓ 检查标题级别范围（1-5级）

**验证规则**:
- 标题不能为空字符串
- 标题级别应在 1-5 之间
- 级别跳跃不应过大（会产生警告）

### 4. TestMergeHeadings

**目的**: 测试标题合并和去重功能

**测试内容**:
- ✓ 验证重复标题被正确去除
- ✓ 验证来自不同来源的标题被正确合并
- ✓ 检查合并后的标题数量

**测试场景**:
```go
大纲标题: ["1. Introduction", "1.1 Background"]
文本标题: ["1. Introduction", "2. Methods"]  // 第一个重复
合并结果: ["1. Introduction", "1.1 Background", "2. Methods"]  // 3个唯一标题
```

### 5. BenchmarkExtractHeadings

**目的**: 性能基准测试

**测试内容**:
- 测量标题提取的性能
- 提供内存分配统计

**运行方式**:
```bash
go test -bench=BenchmarkExtractHeadings -benchmem
```

## 测试数据

### 测试文件

- **135_title.pdf**: 主要测试文件，包含 56 个标题
- **testdata/expected_135_title.txt**: 预期提取结果的参考文档

### 预期结果

从 135_title.pdf 提取的标题应该包括：

- **6 个一级标题**: 主要章节
- **24 个二级标题**: 子章节
- **26 个三级标题**: 详细条目

详细的预期结果请参考 `testdata/expected_135_title.txt`

## 测试失败排查

### 如果测试失败

1. **检查 PDF 文件**
   ```bash
   ls -lh 135_title.pdf
   ```
   确保文件存在且可读

2. **查看详细错误信息**
   ```bash
   go test -v
   ```

3. **检查 unipdf 库**
   ```bash
   go mod verify
   go mod download
   ```

### 常见问题

**Q: 提示 "Unlicensed copy of UniPDF"**
A: 这是正常的警告信息，不影响测试运行。unipdf 免费版本会显示此信息。

**Q: 标题数量不匹配**
A: 检查 PDF 文件是否被修改。如果 PDF 大纲更新，测试预期值也需要相应更新。

**Q: 某些标题缺失**
A: 工具依赖 PDF 的大纲（书签）信息。如果 PDF 大纲不完整，某些标题可能无法提取。

## 添加新测试

### 测试新的 PDF 文件

1. 将 PDF 文件放在项目根目录
2. 在 `main_test.go` 中添加新的测试函数：

```go
func TestExtractHeadingsNewPDF(t *testing.T) {
    pdfPath := "new_document.pdf"
    
    headings, err := extractHeadings(pdfPath)
    if err != nil {
        t.Fatalf("提取标题失败: %v", err)
    }
    
    // 添加你的验证逻辑
    if len(headings) == 0 {
        t.Error("未提取到任何标题")
    }
}
```

3. 运行测试：
```bash
go test -v -run TestExtractHeadingsNewPDF
```

## 持续集成

### GitHub Actions 示例

```yaml
name: Test
on: [push, pull_request]
jobs:
  test:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v2
      - uses: actions/setup-go@v2
        with:
          go-version: '1.21'
      - run: make test
```

## 测试最佳实践

1. **运行测试前先编译**
   ```bash
   make build && make test
   ```

2. **定期更新依赖**
   ```bash
   go get -u ./...
   go mod tidy
   ```

3. **保持测试数据最新**
   - 如果 PDF 文件更新，同步更新 `testdata/expected_*.txt`
   - 更新测试用例中的预期值

4. **使用表格驱动测试**
   - 对于多个相似的测试场景，使用表格驱动测试可以提高可维护性

## 总结

本测试套件提供了全面的验证，确保 PDF 标题提取功能的正确性和稳定性。通过运行这些测试，你可以：

- ✓ 验证核心功能正常工作
- ✓ 确保代码修改不会破坏现有功能
- ✓ 了解工具的性能特征
- ✓ 快速定位问题

如有问题，请查看测试输出的详细信息或参考本文档的故障排查部分。
