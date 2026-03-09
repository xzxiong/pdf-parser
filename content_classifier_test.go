package main

import (
	"testing"
)

func TestNewContentClassifier(t *testing.T) {
	classifier := NewContentClassifier()

	if classifier.HeaderThreshold != 0.15 {
		t.Errorf("HeaderThreshold = %f, 期望 0.15", classifier.HeaderThreshold)
	}

	if classifier.FooterThreshold != 0.15 {
		t.Errorf("FooterThreshold = %f, 期望 0.15", classifier.FooterThreshold)
	}
}

func TestClassifyTextBlock_Header(t *testing.T) {
	classifier := NewContentClassifier()

	// 测试页眉区域（页面顶部 15%）
	block := TextBlock{
		Content: "Header Text",
		Page:    1,
		YPos:    900.0, // 接近页面顶部
		Height:  1000.0,
	}

	result := classifier.ClassifyTextBlock(block)
	if result != "header" {
		t.Errorf("ClassifyTextBlock() = %s, 期望 'header'", result)
	}
}

func TestClassifyTextBlock_Footer(t *testing.T) {
	classifier := NewContentClassifier()

	// 测试页脚区域（页面底部 15%）
	block := TextBlock{
		Content: "Footer Text",
		Page:    1,
		YPos:    100.0, // 接近页面底部
		Height:  1000.0,
	}

	result := classifier.ClassifyTextBlock(block)
	if result != "footer" {
		t.Errorf("ClassifyTextBlock() = %s, 期望 'footer'", result)
	}
}

func TestClassifyTextBlock_Body(t *testing.T) {
	classifier := NewContentClassifier()

	// 测试正文区域（页面中间）
	block := TextBlock{
		Content: "Body Text",
		Page:    1,
		YPos:    500.0, // 页面中间
		Height:  1000.0,
	}

	result := classifier.ClassifyTextBlock(block)
	if result != "body" {
		t.Errorf("ClassifyTextBlock() = %s, 期望 'body'", result)
	}
}

func TestClassifyTextBlock_Boundary(t *testing.T) {
	classifier := NewContentClassifier()

	// 测试边界情况：刚好在页眉阈值上 (> 85%)
	block := TextBlock{
		Content: "Boundary Text",
		Page:    1,
		YPos:    860.0, // 86% 位置，应该在页眉区域
		Height:  1000.0,
	}

	result := classifier.ClassifyTextBlock(block)
	if result != "header" {
		t.Errorf("ClassifyTextBlock() = %s, 期望 'header'", result)
	}

	// 测试边界情况：刚好在页脚阈值下 (< 15%)
	block2 := TextBlock{
		Content: "Boundary Text",
		Page:    1,
		YPos:    140.0, // 14% 位置，应该在页脚区域
		Height:  1000.0,
	}

	result2 := classifier.ClassifyTextBlock(block2)
	if result2 != "footer" {
		t.Errorf("ClassifyTextBlock() = %s, 期望 'footer'", result2)
	}
}

func TestDetectRepeatingPatterns_NoRepetition(t *testing.T) {
	classifier := NewContentClassifier()

	blocks := []TextBlock{
		{Content: "Text 1", Page: 1, YPos: 100, Height: 1000},
		{Content: "Text 2", Page: 2, YPos: 100, Height: 1000},
		{Content: "Text 3", Page: 3, YPos: 100, Height: 1000},
	}

	patterns := classifier.DetectRepeatingPatterns(blocks)

	if len(patterns) != 0 {
		t.Errorf("DetectRepeatingPatterns() 返回 %d 个模式, 期望 0 (没有重复)", len(patterns))
	}
}

func TestDetectRepeatingPatterns_WithRepetition(t *testing.T) {
	classifier := NewContentClassifier()

	blocks := []TextBlock{
		{Content: "Header", Page: 1, YPos: 900, Height: 1000},
		{Content: "Header", Page: 2, YPos: 900, Height: 1000},
		{Content: "Header", Page: 3, YPos: 900, Height: 1000},
		{Content: "Body", Page: 1, YPos: 500, Height: 1000},
		{Content: "Body", Page: 2, YPos: 500, Height: 1000},
	}

	patterns := classifier.DetectRepeatingPatterns(blocks)

	if len(patterns) != 1 {
		t.Errorf("DetectRepeatingPatterns() 返回 %d 个模式, 期望 1", len(patterns))
	}

	if pages, ok := patterns["Header"]; !ok {
		t.Error("未找到 'Header' 模式")
	} else if len(pages) != 3 {
		t.Errorf("'Header' 模式出现在 %d 个页面, 期望 3", len(pages))
	}

	// "Body" 只出现在 2 个页面，不应被识别为重复模式
	if _, ok := patterns["Body"]; ok {
		t.Error("'Body' 不应被识别为重复模式（只出现在 2 个页面）")
	}
}

func TestDetectRepeatingPatterns_EmptyContent(t *testing.T) {
	classifier := NewContentClassifier()

	blocks := []TextBlock{
		{Content: "", Page: 1, YPos: 100, Height: 1000},
		{Content: "  ", Page: 2, YPos: 100, Height: 1000},
		{Content: "\n", Page: 3, YPos: 100, Height: 1000},
	}

	patterns := classifier.DetectRepeatingPatterns(blocks)

	if len(patterns) != 0 {
		t.Errorf("DetectRepeatingPatterns() 返回 %d 个模式, 期望 0 (空内容应被忽略)", len(patterns))
	}
}

func TestDetectRepeatingPatterns_DuplicatePages(t *testing.T) {
	classifier := NewContentClassifier()

	// 同一内容在同一页面出现多次
	blocks := []TextBlock{
		{Content: "Repeated", Page: 1, YPos: 100, Height: 1000},
		{Content: "Repeated", Page: 1, YPos: 200, Height: 1000},
		{Content: "Repeated", Page: 2, YPos: 100, Height: 1000},
		{Content: "Repeated", Page: 3, YPos: 100, Height: 1000},
	}

	patterns := classifier.DetectRepeatingPatterns(blocks)

	if len(patterns) != 1 {
		t.Errorf("DetectRepeatingPatterns() 返回 %d 个模式, 期望 1", len(patterns))
	}

	if pages, ok := patterns["Repeated"]; !ok {
		t.Error("未找到 'Repeated' 模式")
	} else if len(pages) != 3 {
		t.Errorf("'Repeated' 模式出现在 %d 个不同页面, 期望 3", len(pages))
	}
}

func TestClassifyTextBlock_CustomThresholds(t *testing.T) {
	// 测试自定义阈值
	classifier := &ContentClassifier{
		HeaderThreshold: 0.20, // 20% 页眉区域
		FooterThreshold: 0.10, // 10% 页脚区域
	}

	// 测试 15% 位置（在默认阈值下是页眉，但在自定义阈值下是正文）
	block := TextBlock{
		Content: "Text",
		Page:    1,
		YPos:    850.0, // 85% 位置
		Height:  1000.0,
	}

	result := classifier.ClassifyTextBlock(block)
	if result != "header" {
		t.Errorf("ClassifyTextBlock() = %s, 期望 'header' (85%% > 80%%)", result)
	}

	// 测试 12% 位置（在默认阈值下是页脚，但在自定义阈值下是正文）
	block2 := TextBlock{
		Content: "Text",
		Page:    1,
		YPos:    120.0, // 12% 位置
		Height:  1000.0,
	}

	result2 := classifier.ClassifyTextBlock(block2)
	if result2 != "body" {
		t.Errorf("ClassifyTextBlock() = %s, 期望 'body' (12%% > 10%%)", result2)
	}
}
