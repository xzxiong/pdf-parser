package main

import (
	"strings"
	"testing"
)

func TestFormatAsText_Headings(t *testing.T) {
	content := &PDFContent{
		Headings: []Heading{
			{Title: "第一章 引言", Level: 1, Page: 1},
			{Title: "1.1 背景", Level: 2, Page: 2},
			{Title: "1.2 目标", Level: 2, Page: 3},
			{Title: "第二章 方法", Level: 1, Page: 4},
		},
		Body:    []BodyText{},
		Headers: []HeaderFooter{},
		Footers: []HeaderFooter{},
	}

	options := FormatOptions{
		ShowHeadings: true,
		ShowBody:     false,
		ShowHeaders:  false,
		ShowFooters:  false,
	}

	result := formatAsText(content, options)

	// 验证包含分区标记
	if !strings.Contains(result, "=== 标题 (4 个) ===") {
		t.Errorf("Expected heading section marker with count, got: %s", result)
	}

	// 验证包含所有标题
	if !strings.Contains(result, "第一章 引言") {
		t.Error("Expected to find '第一章 引言' in output")
	}

	// 验证层级缩进
	if !strings.Contains(result, "  级别 2: 1.1 背景") {
		t.Error("Expected level 2 heading to have 2-space indentation")
	}

	// 验证页码显示
	if !strings.Contains(result, "(第 1 页)") {
		t.Error("Expected page number in output")
	}
}

func TestFormatAsText_Headers(t *testing.T) {
	content := &PDFContent{
		Headings: []Heading{},
		Body:     []BodyText{},
		Headers: []HeaderFooter{
			{Content: "文档标题 | 第一章", PageRange: []int64{1, 2, 3, 5, 6}, Type: "header"},
		},
		Footers: []HeaderFooter{},
	}

	options := FormatOptions{
		ShowHeadings: false,
		ShowBody:     false,
		ShowHeaders:  true,
		ShowFooters:  false,
	}

	result := formatAsText(content, options)

	// 验证包含页眉分区标记
	if !strings.Contains(result, "=== 页眉 ===") {
		t.Error("Expected header section marker")
	}

	// 验证页眉内容
	if !strings.Contains(result, "文档标题 | 第一章") {
		t.Error("Expected header content in output")
	}

	// 验证页码范围格式
	if !strings.Contains(result, "第 1-3, 5-6 页") {
		t.Errorf("Expected formatted page range, got: %s", result)
	}
}

func TestFormatAsText_Footers(t *testing.T) {
	content := &PDFContent{
		Headings: []Heading{},
		Body:     []BodyText{},
		Headers:  []HeaderFooter{},
		Footers: []HeaderFooter{
			{Content: "第 {n} 页", PageRange: []int64{1, 2, 3, 4, 5, 6, 7, 8, 9, 10}, Type: "footer"},
		},
	}

	options := FormatOptions{
		ShowHeadings: false,
		ShowBody:     false,
		ShowHeaders:  false,
		ShowFooters:  true,
	}

	result := formatAsText(content, options)

	// 验证包含页脚分区标记
	if !strings.Contains(result, "=== 页脚 ===") {
		t.Error("Expected footer section marker")
	}

	// 验证页脚内容
	if !strings.Contains(result, "第 {n} 页") {
		t.Error("Expected footer content in output")
	}

	// 验证页码范围格式
	if !strings.Contains(result, "第 1-10 页") {
		t.Errorf("Expected formatted page range, got: %s", result)
	}
}

func TestFormatAsText_Body(t *testing.T) {
	content := &PDFContent{
		Headings: []Heading{},
		Body: []BodyText{
			{Content: "这是第一页的正文内容...", Page: 1},
			{Content: "这是第二页的正文内容...", Page: 2},
		},
		Headers: []HeaderFooter{},
		Footers: []HeaderFooter{},
	}

	options := FormatOptions{
		ShowHeadings: false,
		ShowBody:     true,
		ShowHeaders:  false,
		ShowFooters:  false,
	}

	result := formatAsText(content, options)

	// 验证包含正文分区标记
	if !strings.Contains(result, "=== 正文 (第 1 页) ===") {
		t.Error("Expected body section marker for page 1")
	}

	if !strings.Contains(result, "=== 正文 (第 2 页) ===") {
		t.Error("Expected body section marker for page 2")
	}

	// 验证正文内容
	if !strings.Contains(result, "这是第一页的正文内容...") {
		t.Error("Expected body content from page 1")
	}

	if !strings.Contains(result, "这是第二页的正文内容...") {
		t.Error("Expected body content from page 2")
	}
}

func TestFormatAsText_AllContent(t *testing.T) {
	content := &PDFContent{
		Headings: []Heading{
			{Title: "第一章 引言", Level: 1, Page: 1},
		},
		Body: []BodyText{
			{Content: "正文内容", Page: 1},
		},
		Headers: []HeaderFooter{
			{Content: "页眉", PageRange: []int64{1, 2, 3}, Type: "header"},
		},
		Footers: []HeaderFooter{
			{Content: "页脚", PageRange: []int64{1, 2, 3}, Type: "footer"},
		},
	}

	options := FormatOptions{
		ShowHeadings: true,
		ShowBody:     true,
		ShowHeaders:  true,
		ShowFooters:  true,
	}

	result := formatAsText(content, options)

	// 验证所有分区都存在
	if !strings.Contains(result, "=== 标题") {
		t.Error("Expected headings section")
	}

	if !strings.Contains(result, "=== 页眉 ===") {
		t.Error("Expected headers section")
	}

	if !strings.Contains(result, "=== 页脚 ===") {
		t.Error("Expected footers section")
	}

	if !strings.Contains(result, "=== 正文") {
		t.Error("Expected body section")
	}
}

func TestFormatAsText_EmptyContent(t *testing.T) {
	content := &PDFContent{
		Headings: []Heading{},
		Body:     []BodyText{},
		Headers:  []HeaderFooter{},
		Footers:  []HeaderFooter{},
	}

	options := FormatOptions{
		ShowHeadings: true,
		ShowBody:     true,
		ShowHeaders:  true,
		ShowFooters:  true,
	}

	result := formatAsText(content, options)

	// 空内容应该返回空字符串或只有换行
	if strings.Contains(result, "===") {
		t.Error("Expected no section markers for empty content")
	}
}

func TestFormatAsText_SelectiveDisplay(t *testing.T) {
	content := &PDFContent{
		Headings: []Heading{
			{Title: "标题", Level: 1, Page: 1},
		},
		Body: []BodyText{
			{Content: "正文", Page: 1},
		},
		Headers: []HeaderFooter{
			{Content: "页眉", PageRange: []int64{1}, Type: "header"},
		},
		Footers: []HeaderFooter{
			{Content: "页脚", PageRange: []int64{1}, Type: "footer"},
		},
	}

	// 只显示标题
	options := FormatOptions{
		ShowHeadings: true,
		ShowBody:     false,
		ShowHeaders:  false,
		ShowFooters:  false,
	}

	result := formatAsText(content, options)

	if !strings.Contains(result, "标题") {
		t.Error("Expected headings to be shown")
	}

	if strings.Contains(result, "正文") {
		t.Error("Expected body to be hidden")
	}

	if strings.Contains(result, "页眉") {
		t.Error("Expected headers to be hidden")
	}

	if strings.Contains(result, "页脚") {
		t.Error("Expected footers to be hidden")
	}
}

func TestFormatPageRange_Continuous(t *testing.T) {
	pages := []int64{1, 2, 3, 4, 5}
	result := formatPageRange(pages)

	expected := "第 1-5 页"
	if result != expected {
		t.Errorf("Expected '%s', got '%s'", expected, result)
	}
}

func TestFormatPageRange_Discontinuous(t *testing.T) {
	pages := []int64{1, 2, 3, 5, 6, 7, 10}
	result := formatPageRange(pages)

	expected := "第 1-3, 5-7, 10 页"
	if result != expected {
		t.Errorf("Expected '%s', got '%s'", expected, result)
	}
}

func TestFormatPageRange_SinglePage(t *testing.T) {
	pages := []int64{5}
	result := formatPageRange(pages)

	expected := "第 5 页"
	if result != expected {
		t.Errorf("Expected '%s', got '%s'", expected, result)
	}
}

func TestFormatPageRange_Unsorted(t *testing.T) {
	pages := []int64{5, 1, 3, 2, 4}
	result := formatPageRange(pages)

	expected := "第 1-5 页"
	if result != expected {
		t.Errorf("Expected '%s', got '%s'", expected, result)
	}
}

func TestFormatPageRange_WithDuplicates(t *testing.T) {
	pages := []int64{1, 2, 2, 3, 3, 3, 5}
	result := formatPageRange(pages)

	expected := "第 1-3, 5 页"
	if result != expected {
		t.Errorf("Expected '%s', got '%s'", expected, result)
	}
}

func TestFormatPageRange_Empty(t *testing.T) {
	pages := []int64{}
	result := formatPageRange(pages)

	if result != "" {
		t.Errorf("Expected empty string for empty pages, got '%s'", result)
	}
}

func TestFormatAsText_BackwardCompatibility(t *testing.T) {
	// 测试向后兼容性：标题输出格式应与原有格式一致
	content := &PDFContent{
		Headings: []Heading{
			{Title: "第一章", Level: 1, Page: 1},
			{Title: "1.1 小节", Level: 2, Page: 2},
		},
		Body:    []BodyText{},
		Headers: []HeaderFooter{},
		Footers: []HeaderFooter{},
	}

	options := FormatOptions{
		ShowHeadings: true,
		ShowBody:     false,
		ShowHeaders:  false,
		ShowFooters:  false,
	}

	result := formatAsText(content, options)

	// 验证格式：级别 X: 标题 (第 Y 页)
	if !strings.Contains(result, "级别 1: 第一章 (第 1 页)") {
		t.Error("Expected backward compatible heading format for level 1")
	}

	if !strings.Contains(result, "  级别 2: 1.1 小节 (第 2 页)") {
		t.Error("Expected backward compatible heading format for level 2 with indentation")
	}
}
