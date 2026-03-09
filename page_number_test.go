package main

import (
	"testing"
)

// TestIsPageNumberPattern 测试页码模式识别
func TestIsPageNumberPattern(t *testing.T) {
	tests := []struct {
		name     string
		content  string
		pageNum  int64
		wantBool bool
		wantNum  int64
	}{
		{
			name:     "纯数字页码",
			content:  "5",
			pageNum:  5,
			wantBool: true,
			wantNum:  5,
		},
		{
			name:     "带前缀的页码 - 中文",
			content:  "第 3 页",
			pageNum:  3,
			wantBool: true,
			wantNum:  3,
		},
		{
			name:     "带前缀的页码 - 英文",
			content:  "Page 7",
			pageNum:  7,
			wantBool: true,
			wantNum:  7,
		},
		{
			name:     "带分隔符的页码",
			content:  "- 10 -",
			pageNum:  10,
			wantBool: true,
			wantNum:  10,
		},
		{
			name:     "带后缀的页码",
			content:  "15 页",
			pageNum:  15,
			wantBool: true,
			wantNum:  15,
		},
		{
			name:     "罗马数字页码",
			content:  "iii",
			pageNum:  3,
			wantBool: true,
			wantNum:  3,
		},
		{
			name:     "非页码文本",
			content:  "这是正文内容",
			pageNum:  5,
			wantBool: false,
			wantNum:  0,
		},
		{
			name:     "数字不匹配页码",
			content:  "100",
			pageNum:  5,
			wantBool: false,
			wantNum:  0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotBool, gotNum := isPageNumberPattern(tt.content, tt.pageNum)
			if gotBool != tt.wantBool {
				t.Errorf("isPageNumberPattern() gotBool = %v, want %v", gotBool, tt.wantBool)
			}
			if gotNum != tt.wantNum {
				t.Errorf("isPageNumberPattern() gotNum = %v, want %v", gotNum, tt.wantNum)
			}
		})
	}
}

// TestDetectPageNumberSequence 测试页码序列检测
func TestDetectPageNumberSequence(t *testing.T) {
	tests := []struct {
		name        string
		blocks      []TextBlock
		wantBool    bool
		wantPattern string
	}{
		{
			name: "递增的页码序列",
			blocks: []TextBlock{
				{Content: "1", Page: 1, YPos: 10, Height: 800},
				{Content: "2", Page: 2, YPos: 10, Height: 800},
				{Content: "3", Page: 3, YPos: 10, Height: 800},
				{Content: "4", Page: 4, YPos: 10, Height: 800},
			},
			wantBool:    true,
			wantPattern: "{n}",
		},
		{
			name: "带格式的递增页码序列",
			blocks: []TextBlock{
				{Content: "- 1 -", Page: 1, YPos: 10, Height: 800},
				{Content: "- 2 -", Page: 2, YPos: 10, Height: 800},
				{Content: "- 3 -", Page: 3, YPos: 10, Height: 800},
				{Content: "- 4 -", Page: 4, YPos: 10, Height: 800},
			},
			wantBool:    true,
			wantPattern: "- {n} -",
		},
		{
			name: "样本不足",
			blocks: []TextBlock{
				{Content: "1", Page: 1, YPos: 10, Height: 800},
				{Content: "2", Page: 2, YPos: 10, Height: 800},
			},
			wantBool:    false,
			wantPattern: "",
		},
		{
			name: "非页码序列",
			blocks: []TextBlock{
				{Content: "标题", Page: 1, YPos: 10, Height: 800},
				{Content: "标题", Page: 2, YPos: 10, Height: 800},
				{Content: "标题", Page: 3, YPos: 10, Height: 800},
			},
			wantBool:    false,
			wantPattern: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotBool, gotPattern := detectPageNumberSequence(tt.blocks)
			if gotBool != tt.wantBool {
				t.Errorf("detectPageNumberSequence() gotBool = %v, want %v", gotBool, tt.wantBool)
			}
			if gotPattern != tt.wantPattern {
				t.Errorf("detectPageNumberSequence() gotPattern = %v, want %v", gotPattern, tt.wantPattern)
			}
		})
	}
}

// TestIsRomanNumeral 测试罗马数字识别
func TestIsRomanNumeral(t *testing.T) {
	tests := []struct {
		name string
		s    string
		want bool
	}{
		{name: "小写罗马数字 i", s: "i", want: true},
		{name: "小写罗马数字 iv", s: "iv", want: true},
		{name: "小写罗马数字 ix", s: "ix", want: true},
		{name: "大写罗马数字 X", s: "X", want: true},
		{name: "大写罗马数字 XV", s: "XV", want: true},
		{name: "非罗马数字", s: "abc", want: false},
		{name: "空字符串", s: "", want: false},
		{name: "数字", s: "123", want: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := isRomanNumeral(tt.s); got != tt.want {
				t.Errorf("isRomanNumeral() = %v, want %v", got, tt.want)
			}
		})
	}
}

// TestRomanToInt 测试罗马数字转换
func TestRomanToInt(t *testing.T) {
	tests := []struct {
		name string
		s    string
		want int
	}{
		{name: "i = 1", s: "i", want: 1},
		{name: "ii = 2", s: "ii", want: 2},
		{name: "iii = 3", s: "iii", want: 3},
		{name: "iv = 4", s: "iv", want: 4},
		{name: "v = 5", s: "v", want: 5},
		{name: "ix = 9", s: "ix", want: 9},
		{name: "x = 10", s: "x", want: 10},
		{name: "XV = 15", s: "XV", want: 15},
		{name: "XX = 20", s: "XX", want: 20},
		{name: "L = 50", s: "L", want: 50},
		{name: "C = 100", s: "C", want: 100},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := romanToInt(tt.s); got != tt.want {
				t.Errorf("romanToInt() = %v, want %v", got, tt.want)
			}
		})
	}
}
