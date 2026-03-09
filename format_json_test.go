package main

import (
	"encoding/json"
	"strings"
	"testing"
)

// TestFormatAsJSON 测试 formatAsJSON 函数
func TestFormatAsJSON(t *testing.T) {
	tests := []struct {
		name    string
		content *PDFContent
		wantErr bool
	}{
		{
			name: "空内容",
			content: &PDFContent{
				Headings: []Heading{},
				Body:     []BodyText{},
				Headers:  []HeaderFooter{},
				Footers:  []HeaderFooter{},
			},
			wantErr: false,
		},
		{
			name: "包含标题",
			content: &PDFContent{
				Headings: []Heading{
					{Title: "第一章", Level: 1, Page: 1},
					{Title: "1.1 引言", Level: 2, Page: 2},
				},
				Body:    []BodyText{},
				Headers: []HeaderFooter{},
				Footers: []HeaderFooter{},
			},
			wantErr: false,
		},
		{
			name: "包含所有类型内容",
			content: &PDFContent{
				Headings: []Heading{
					{Title: "第一章", Level: 1, Page: 1},
				},
				Body: []BodyText{
					{Content: "这是正文内容", Page: 1},
				},
				Headers: []HeaderFooter{
					{Content: "文档标题", PageRange: []int64{1, 2, 3}, Type: "header"},
				},
				Footers: []HeaderFooter{
					{Content: "第 {n} 页", PageRange: []int64{1, 2, 3}, Type: "footer"},
				},
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := formatAsJSON(tt.content)
			if (err != nil) != tt.wantErr {
				t.Errorf("formatAsJSON() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if err == nil {
				// 验证返回的字符串是有效的 JSON
				var result PDFContent
				if err := json.Unmarshal([]byte(got), &result); err != nil {
					t.Errorf("formatAsJSON() 返回的不是有效的 JSON: %v", err)
					return
				}

				// 验证内容是否正确
				if len(result.Headings) != len(tt.content.Headings) {
					t.Errorf("标题数量不匹配: got %d, want %d", len(result.Headings), len(tt.content.Headings))
				}
				if len(result.Body) != len(tt.content.Body) {
					t.Errorf("正文数量不匹配: got %d, want %d", len(result.Body), len(tt.content.Body))
				}
				if len(result.Headers) != len(tt.content.Headers) {
					t.Errorf("页眉数量不匹配: got %d, want %d", len(result.Headers), len(tt.content.Headers))
				}
				if len(result.Footers) != len(tt.content.Footers) {
					t.Errorf("页脚数量不匹配: got %d, want %d", len(result.Footers), len(tt.content.Footers))
				}
			}
		})
	}
}

// TestFormatAsJSON_ValidJSON 测试 JSON 输出的有效性
func TestFormatAsJSON_ValidJSON(t *testing.T) {
	content := &PDFContent{
		Headings: []Heading{
			{Title: "第一章 引言", Level: 1, Page: 1},
			{Title: "1.1 背景", Level: 2, Page: 2},
			{Title: "1.2 目标", Level: 2, Page: 3},
		},
		Body: []BodyText{
			{Content: "这是第一页的正文内容。\n包含多行文本。", Page: 1},
			{Content: "这是第二页的正文内容。", Page: 2},
		},
		Headers: []HeaderFooter{
			{Content: "文档标题 | 第一章", PageRange: []int64{1, 2, 3, 4, 5}, Type: "header"},
		},
		Footers: []HeaderFooter{
			{Content: "第 {n} 页", PageRange: []int64{1, 2, 3, 4, 5}, Type: "footer"},
		},
	}

	jsonStr, err := formatAsJSON(content)
	if err != nil {
		t.Fatalf("formatAsJSON() 失败: %v", err)
	}

	// 验证 JSON 格式
	var result PDFContent
	if err := json.Unmarshal([]byte(jsonStr), &result); err != nil {
		t.Fatalf("返回的不是有效的 JSON: %v\nJSON 内容:\n%s", err, jsonStr)
	}

	// 验证数据完整性
	if len(result.Headings) != 3 {
		t.Errorf("标题数量错误: got %d, want 3", len(result.Headings))
	}
	if len(result.Body) != 2 {
		t.Errorf("正文数量错误: got %d, want 2", len(result.Body))
	}
	if len(result.Headers) != 1 {
		t.Errorf("页眉数量错误: got %d, want 1", len(result.Headers))
	}
	if len(result.Footers) != 1 {
		t.Errorf("页脚数量错误: got %d, want 1", len(result.Footers))
	}

	// 验证具体内容
	if result.Headings[0].Title != "第一章 引言" {
		t.Errorf("第一个标题错误: got %s, want 第一章 引言", result.Headings[0].Title)
	}
	if result.Body[0].Content != "这是第一页的正文内容。\n包含多行文本。" {
		t.Errorf("第一段正文内容错误")
	}
}

// TestFormatAsJSON_Indentation 测试 JSON 输出的格式化（缩进）
func TestFormatAsJSON_Indentation(t *testing.T) {
	content := &PDFContent{
		Headings: []Heading{
			{Title: "测试标题", Level: 1, Page: 1},
		},
		Body:    []BodyText{},
		Headers: []HeaderFooter{},
		Footers: []HeaderFooter{},
	}

	jsonStr, err := formatAsJSON(content)
	if err != nil {
		t.Fatalf("formatAsJSON() 失败: %v", err)
	}

	// 验证 JSON 包含缩进（格式化的 JSON 应该包含换行符和空格）
	if !containsIndentation(jsonStr) {
		t.Errorf("JSON 输出没有正确格式化（缺少缩进）:\n%s", jsonStr)
	}
}

// containsIndentation 检查字符串是否包含缩进（换行符和空格）
func containsIndentation(s string) bool {
	// 格式化的 JSON 应该包含换行符和多个空格
	return len(s) > 0 && (len(s) != len(strings.TrimSpace(s)) || strings.Contains(s, "\n"))
}
