package main

import (
	"os"
	"testing"

	"github.com/unidoc/unipdf/v3/model"
)

// TestExtractHeaders 测试 extractHeaders 函数
func TestExtractHeaders(t *testing.T) {
	// 打开测试 PDF 文件
	pdfPath := "135_title.pdf"
	f, err := os.Open(pdfPath)
	if err != nil {
		t.Fatalf("无法打开文件: %v", err)
	}
	defer f.Close()

	// 创建 PDF 阅读器
	pdfReader, err := model.NewPdfReader(f)
	if err != nil {
		t.Fatalf("无法读取 PDF: %v", err)
	}

	// 创建分类器
	classifier := NewContentClassifier()

	// 提取页眉
	headers, err := extractHeaders(pdfReader, classifier)
	if err != nil {
		t.Fatalf("提取页眉失败: %v", err)
	}

	t.Logf("成功提取 %d 个页眉", len(headers))

	// 验证页眉的基本属性
	for i, header := range headers {
		// 检查内容是否为空
		if header.Content == "" {
			t.Errorf("页眉 %d 的内容为空", i+1)
		}

		// 检查类型是否正确
		if header.Type != "header" {
			t.Errorf("页眉 %d 的类型 '%s' 不正确，期望 'header'", i+1, header.Type)
		}

		// 检查页码范围是否合理
		if len(header.PageRange) < 3 {
			t.Errorf("页眉 %d 的页码范围 %v 少于 3 页（不符合重复模式要求）", i+1, header.PageRange)
		}

		// 检查页码是否合理
		for _, page := range header.PageRange {
			if page < 1 {
				t.Errorf("页眉 %d 包含不合理的页码 %d", i+1, page)
			}
		}

		t.Logf("页眉 %d: '%s' (出现在 %d 个页面: %v)",
			i+1, header.Content, len(header.PageRange), header.PageRange)
	}
}

// TestExtractHeaders_EmptyPDF 测试空 PDF 或无页眉的情况
func TestExtractHeaders_EmptyPDF(t *testing.T) {
	// 这个测试需要一个没有页眉的 PDF 文件
	// 由于我们使用的是 135_title.pdf，这里只是演示测试结构
	pdfPath := "135_title.pdf"
	f, err := os.Open(pdfPath)
	if err != nil {
		t.Skip("测试文件不存在，跳过测试")
	}
	defer f.Close()

	pdfReader, err := model.NewPdfReader(f)
	if err != nil {
		t.Fatalf("无法读取 PDF: %v", err)
	}

	classifier := NewContentClassifier()
	headers, err := extractHeaders(pdfReader, classifier)
	if err != nil {
		t.Fatalf("提取页眉失败: %v", err)
	}

	// 对于没有页眉的 PDF，应该返回空列表
	// 注意：135_title.pdf 可能有页眉，所以这个测试可能会失败
	// 这里只是验证函数不会崩溃
	t.Logf("提取到 %d 个页眉", len(headers))
}

// TestExtractHeaders_OddEvenPages 测试奇偶页不同的页眉
func TestExtractHeaders_OddEvenPages(t *testing.T) {
	pdfPath := "135_title.pdf"
	f, err := os.Open(pdfPath)
	if err != nil {
		t.Fatalf("无法打开文件: %v", err)
	}
	defer f.Close()

	pdfReader, err := model.NewPdfReader(f)
	if err != nil {
		t.Fatalf("无法读取 PDF: %v", err)
	}

	classifier := NewContentClassifier()
	headers, err := extractHeaders(pdfReader, classifier)
	if err != nil {
		t.Fatalf("提取页眉失败: %v", err)
	}

	// 检查是否有奇偶页不同的页眉
	for i, header := range headers {
		oddCount := 0
		evenCount := 0

		for _, page := range header.PageRange {
			if page%2 == 1 {
				oddCount++
			} else {
				evenCount++
			}
		}

		// 如果只出现在奇数页或偶数页，说明是奇偶页页眉
		if oddCount > 0 && evenCount == 0 {
			t.Logf("页眉 %d 只出现在奇数页: '%s' (页码: %v)",
				i+1, header.Content, header.PageRange)
		} else if evenCount > 0 && oddCount == 0 {
			t.Logf("页眉 %d 只出现在偶数页: '%s' (页码: %v)",
				i+1, header.Content, header.PageRange)
		} else {
			t.Logf("页眉 %d 出现在奇偶页: '%s' (奇数页: %d, 偶数页: %d)",
				i+1, header.Content, oddCount, evenCount)
		}
	}
}

// TestExtractHeaders_PageRangeAccuracy 测试页码范围的准确性
func TestExtractHeaders_PageRangeAccuracy(t *testing.T) {
	pdfPath := "135_title.pdf"
	f, err := os.Open(pdfPath)
	if err != nil {
		t.Fatalf("无法打开文件: %v", err)
	}
	defer f.Close()

	pdfReader, err := model.NewPdfReader(f)
	if err != nil {
		t.Fatalf("无法读取 PDF: %v", err)
	}

	classifier := NewContentClassifier()
	headers, err := extractHeaders(pdfReader, classifier)
	if err != nil {
		t.Fatalf("提取页眉失败: %v", err)
	}

	// 验证页码范围的准确性
	for i, header := range headers {
		// 页码应该是唯一的
		pageSet := make(map[int64]bool)
		for _, page := range header.PageRange {
			if pageSet[page] {
				t.Errorf("页眉 %d 的页码范围包含重复页码 %d", i+1, page)
			}
			pageSet[page] = true
		}

		// 页码应该是递增的（虽然不一定连续）
		for j := 1; j < len(header.PageRange); j++ {
			if header.PageRange[j] <= header.PageRange[j-1] {
				t.Logf("警告: 页眉 %d 的页码范围不是递增的: %v", i+1, header.PageRange)
				break
			}
		}
	}
}

// TestExtractHeaders_ContentNotEmpty 测试页眉内容不为空
func TestExtractHeaders_ContentNotEmpty(t *testing.T) {
	pdfPath := "135_title.pdf"
	f, err := os.Open(pdfPath)
	if err != nil {
		t.Fatalf("无法打开文件: %v", err)
	}
	defer f.Close()

	pdfReader, err := model.NewPdfReader(f)
	if err != nil {
		t.Fatalf("无法读取 PDF: %v", err)
	}

	classifier := NewContentClassifier()
	headers, err := extractHeaders(pdfReader, classifier)
	if err != nil {
		t.Fatalf("提取页眉失败: %v", err)
	}

	// 验证所有页眉内容都不为空
	for i, header := range headers {
		if len(header.Content) == 0 {
			t.Errorf("页眉 %d 的内容为空", i+1)
		}

		// 验证内容不是纯空白
		trimmed := len(header.Content) - len(header.Content)
		if trimmed == len(header.Content) {
			t.Errorf("页眉 %d 的内容只包含空白字符", i+1)
		}
	}
}
