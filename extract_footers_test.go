package main

import (
	"os"
	"testing"

	"github.com/unidoc/unipdf/v3/model"
)

// TestExtractFooters 测试 extractFooters 函数
func TestExtractFooters(t *testing.T) {
	// 打开测试 PDF 文件
	pdfPath := "135_title.pdf"
	f, err := os.Open(pdfPath)
	if err != nil {
		t.Fatalf("无法打开测试文件: %v", err)
	}
	defer f.Close()

	pdfReader, err := model.NewPdfReader(f)
	if err != nil {
		t.Fatalf("无法创建 PDF reader: %v", err)
	}

	classifier := NewContentClassifier()

	// 调用 extractFooters
	footers, err := extractFooters(pdfReader, classifier)
	if err != nil {
		t.Fatalf("extractFooters 失败: %v", err)
	}

	// 验证返回的页脚列表不为 nil
	if footers == nil {
		t.Error("footers 不应为 nil")
	}

	// 打印页脚信息用于调试
	t.Logf("提取到 %d 个页脚", len(footers))
	for i, footer := range footers {
		t.Logf("页脚 %d: 内容='%s', 类型='%s', 页码范围=%v",
			i+1, footer.Content, footer.Type, footer.PageRange)
	}

	// 验证页脚的基本属性
	for i, footer := range footers {
		if footer.Type != "footer" {
			t.Errorf("页脚 %d 的类型应为 'footer'，实际为 '%s'", i+1, footer.Type)
		}

		if len(footer.PageRange) == 0 {
			t.Errorf("页脚 %d 的页码范围不应为空", i+1)
		}

		if footer.Content == "" {
			t.Errorf("页脚 %d 的内容不应为空", i+1)
		}
	}
}

// TestExtractFooters_EmptyPDF 测试空 PDF 或无页脚的情况
func TestExtractFooters_EmptyPDF(t *testing.T) {
	// 这个测试需要一个没有页脚的 PDF 文件
	// 由于我们使用的是 135_title.pdf，这里只是演示测试结构
	pdfPath := "135_title.pdf"
	f, err := os.Open(pdfPath)
	if err != nil {
		t.Skipf("跳过测试: 无法打开测试文件: %v", err)
	}
	defer f.Close()

	pdfReader, err := model.NewPdfReader(f)
	if err != nil {
		t.Fatalf("无法创建 PDF reader: %v", err)
	}

	classifier := NewContentClassifier()

	footers, err := extractFooters(pdfReader, classifier)
	if err != nil {
		t.Fatalf("extractFooters 失败: %v", err)
	}

	// 验证返回的是空列表而不是 nil
	if footers == nil {
		t.Error("即使没有页脚，也应返回空列表而不是 nil")
	}
}

// TestExtractFooters_PageNumberDetection 测试页码检测
func TestExtractFooters_PageNumberDetection(t *testing.T) {
	pdfPath := "135_title.pdf"
	f, err := os.Open(pdfPath)
	if err != nil {
		t.Fatalf("无法打开测试文件: %v", err)
	}
	defer f.Close()

	pdfReader, err := model.NewPdfReader(f)
	if err != nil {
		t.Fatalf("无法创建 PDF reader: %v", err)
	}

	classifier := NewContentClassifier()

	footers, err := extractFooters(pdfReader, classifier)
	if err != nil {
		t.Fatalf("extractFooters 失败: %v", err)
	}

	// 检查是否有页脚包含页码模式
	hasPageNumber := false
	for _, footer := range footers {
		if footer.Content == "{n}" || containsPageNumberPattern(footer.Content) {
			hasPageNumber = true
			t.Logf("检测到页码模式: '%s', 页码范围: %v", footer.Content, footer.PageRange)
		}
	}

	t.Logf("是否检测到页码模式: %v", hasPageNumber)
}

// containsPageNumberPattern 辅助函数，检查内容是否包含页码模式
func containsPageNumberPattern(content string) bool {
	// 简单检查是否包含 {n} 占位符或常见页码关键词
	return content == "{n}" ||
		len(content) > 0 && (content[0] >= '0' && content[0] <= '9')
}

// TestExtractFooters_ContentNotEmpty 测试页脚内容不为空
func TestExtractFooters_ContentNotEmpty(t *testing.T) {
	pdfPath := "135_title.pdf"
	f, err := os.Open(pdfPath)
	if err != nil {
		t.Fatalf("无法打开测试文件: %v", err)
	}
	defer f.Close()

	pdfReader, err := model.NewPdfReader(f)
	if err != nil {
		t.Fatalf("无法创建 PDF reader: %v", err)
	}

	classifier := NewContentClassifier()

	footers, err := extractFooters(pdfReader, classifier)
	if err != nil {
		t.Fatalf("extractFooters 失败: %v", err)
	}

	// 验证所有页脚的内容都不为空
	for i, footer := range footers {
		if footer.Content == "" {
			t.Errorf("页脚 %d 的内容不应为空", i+1)
		}

		if len(footer.PageRange) == 0 {
			t.Errorf("页脚 %d 的页码范围不应为空", i+1)
		}
	}
}
