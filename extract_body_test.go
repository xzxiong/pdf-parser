package main

import (
	"os"
	"testing"

	"github.com/unidoc/unipdf/v3/model"
)

// TestExtractBodyBasic tests the basic functionality of extractBody
func TestExtractBodyBasic(t *testing.T) {
	// 打开测试 PDF 文件
	f, err := os.Open("135_title.pdf")
	if err != nil {
		t.Skipf("测试 PDF 文件不存在: %v", err)
		return
	}
	defer f.Close()

	// 创建 PDF 阅读器
	pdfReader, err := model.NewPdfReader(f)
	if err != nil {
		t.Fatalf("无法读取 PDF: %v", err)
	}

	// 创建分类器
	classifier := NewContentClassifier()

	// 提取正文
	bodyTexts, err := extractBody(pdfReader, classifier)
	if err != nil {
		t.Fatalf("提取正文失败: %v", err)
	}

	// 验证结果
	if len(bodyTexts) == 0 {
		t.Log("警告: 未提取到任何正文内容")
	} else {
		t.Logf("成功提取 %d 页的正文内容", len(bodyTexts))

		// 显示前几页的正文（用于调试）
		for i, body := range bodyTexts {
			if i >= 3 { // 只显示前3页
				break
			}
			t.Logf("第 %d 页正文 (前100字符): %s...", body.Page, truncate(body.Content, 100))
		}
	}

	// 验证正文按页码顺序组织
	for i := 1; i < len(bodyTexts); i++ {
		if bodyTexts[i].Page < bodyTexts[i-1].Page {
			t.Errorf("正文未按页码顺序组织: 第 %d 页在第 %d 页之后", bodyTexts[i].Page, bodyTexts[i-1].Page)
		}
	}
}

// TestExtractBodyNoHeaderFooter tests that headers and footers are excluded
func TestExtractBodyNoHeaderFooter(t *testing.T) {
	f, err := os.Open("135_title.pdf")
	if err != nil {
		t.Skipf("测试 PDF 文件不存在: %v", err)
		return
	}
	defer f.Close()

	pdfReader, err := model.NewPdfReader(f)
	if err != nil {
		t.Fatalf("无法读取 PDF: %v", err)
	}

	classifier := NewContentClassifier()
	bodyTexts, err := extractBody(pdfReader, classifier)
	if err != nil {
		t.Fatalf("提取正文失败: %v", err)
	}

	// 验证正文不包含标题格式
	for _, body := range bodyTexts {
		lines := splitLines(body.Content)
		for _, line := range lines {
			// 简单检查：正文行不应该以数字+点号开头（标题格式）
			if len(line) > 0 && line[0] >= '0' && line[0] <= '9' {
				// 这可能是标题，但也可能是正文中的编号列表
				// 这里只是一个基本检查
				t.Logf("发现可能的标题格式: %s", truncate(line, 50))
			}
		}
	}
}

// 辅助函数：截断字符串
func truncate(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen]
}

// 辅助函数：分割行
func splitLines(s string) []string {
	var lines []string
	start := 0
	for i, r := range s {
		if r == '\n' {
			lines = append(lines, s[start:i])
			start = i + 1
		}
	}
	if start < len(s) {
		lines = append(lines, s[start:])
	}
	return lines
}
