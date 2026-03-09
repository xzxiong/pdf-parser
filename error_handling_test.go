package main

import (
	"errors"
	"fmt"
	"os"
	"testing"
)

// TestErrorWrapping 测试错误包装是否正确使用 %w
func TestErrorWrapping(t *testing.T) {
	// 测试文件不存在的错误
	_, err := extractHeadings("nonexistent_file.pdf")
	if err == nil {
		t.Fatal("期望返回错误，但得到 nil")
	}

	// 验证错误消息包含文件路径
	errMsg := err.Error()
	if !contains(errMsg, "nonexistent_file.pdf") {
		t.Errorf("错误消息应包含文件路径，得到: %s", errMsg)
	}

	// 验证错误消息包含上下文信息
	if !contains(errMsg, "无法打开文件") {
		t.Errorf("错误消息应包含上下文信息，得到: %s", errMsg)
	}

	// 验证错误链是否保留（使用 errors.Unwrap）
	var pathErr *os.PathError
	if !errors.As(err, &pathErr) {
		t.Error("错误链应该包含 os.PathError")
	}

	t.Logf("错误消息: %v", err)
	t.Logf("底层错误类型: %T", errors.Unwrap(err))
}

// TestErrorContextInTextExtraction 测试文本提取中的错误上下文
func TestErrorContextInTextExtraction(t *testing.T) {
	// 创建一个临时的无效 PDF 文件
	tmpFile, err := os.CreateTemp("", "invalid_*.pdf")
	if err != nil {
		t.Fatalf("无法创建临时文件: %v", err)
	}
	defer os.Remove(tmpFile.Name())

	// 写入无效内容
	tmpFile.WriteString("This is not a valid PDF")
	tmpFile.Close()

	// 尝试提取标题
	_, err = extractHeadings(tmpFile.Name())
	if err == nil {
		t.Fatal("期望返回错误，但得到 nil")
	}

	// 验证错误消息包含文件路径
	errMsg := err.Error()
	if !contains(errMsg, tmpFile.Name()) {
		t.Errorf("错误消息应包含文件路径，得到: %s", errMsg)
	}

	// 验证错误消息包含上下文信息
	if !contains(errMsg, "无法读取 PDF 文件") {
		t.Errorf("错误消息应包含上下文信息，得到: %s", errMsg)
	}

	t.Logf("错误消息: %v", err)
}

// TestPartialFailureTolerance 测试部分失败的容错处理
func TestPartialFailureTolerance(t *testing.T) {
	// 使用真实的 PDF 文件测试
	pdfPath := "135_title.pdf"

	headings, err := extractHeadings(pdfPath)

	// 即使文本提取失败（由于许可证问题），也应该返回大纲中的标题
	if err != nil {
		t.Fatalf("不应该返回错误: %v", err)
	}

	// 应该至少提取到大纲中的标题
	if len(headings) == 0 {
		t.Error("应该至少提取到大纲中的标题")
	}

	t.Logf("成功提取 %d 个标题（部分失败容错）", len(headings))
}

// contains 检查字符串是否包含子串
func contains(s, substr string) bool {
	return fmt.Sprintf("%s", s) != "" &&
		len(s) >= len(substr) &&
		(s == substr || len(s) > len(substr) &&
			(s[:len(substr)] == substr ||
				s[len(s)-len(substr):] == substr ||
				containsMiddle(s, substr)))
}

func containsMiddle(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
