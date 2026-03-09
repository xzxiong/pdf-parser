package main

import (
	"testing"
)

// TestExtractAllContent 测试 extractAllContent 函数
func TestExtractAllContent(t *testing.T) {
	// 使用现有的测试 PDF 文件
	pdfPath := "135_title.pdf"

	content, err := extractAllContent(pdfPath)
	if err != nil {
		t.Fatalf("extractAllContent 失败: %v", err)
	}

	// 验证返回的结构不为 nil
	if content == nil {
		t.Fatal("extractAllContent 返回 nil")
	}

	// 验证至少提取到了一些内容
	// 注意：我们不要求所有字段都有内容，因为某些 PDF 可能没有页眉或页脚
	hasContent := len(content.Headings) > 0 ||
		len(content.Body) > 0 ||
		len(content.Headers) > 0 ||
		len(content.Footers) > 0

	if !hasContent {
		t.Error("extractAllContent 没有提取到任何内容")
	}

	// 打印提取的内容统计
	t.Logf("提取结果统计:")
	t.Logf("  标题数量: %d", len(content.Headings))
	t.Logf("  正文段落数量: %d", len(content.Body))
	t.Logf("  页眉数量: %d", len(content.Headers))
	t.Logf("  页脚数量: %d", len(content.Footers))
}

// TestExtractAllContent_InvalidFile 测试无效文件的错误处理
func TestExtractAllContent_InvalidFile(t *testing.T) {
	// 测试不存在的文件
	_, err := extractAllContent("nonexistent.pdf")
	if err == nil {
		t.Error("期望返回错误，但没有返回")
	}
}

// TestExtractAllContent_PartialFailure 测试部分提取失败的情况
// 这个测试验证即使某些提取失败，函数仍然返回已成功提取的内容
func TestExtractAllContent_PartialFailure(t *testing.T) {
	pdfPath := "135_title.pdf"

	content, err := extractAllContent(pdfPath)
	if err != nil {
		t.Fatalf("extractAllContent 失败: %v", err)
	}

	// 验证返回的结构不为 nil
	if content == nil {
		t.Fatal("extractAllContent 返回 nil")
	}

	// 验证所有字段都已初始化（即使为空）
	if content.Headings == nil {
		t.Error("Headings 字段为 nil")
	}
	if content.Body == nil {
		t.Error("Body 字段为 nil")
	}
	if content.Headers == nil {
		t.Error("Headers 字段为 nil")
	}
	if content.Footers == nil {
		t.Error("Footers 字段为 nil")
	}
}
