//go:build ignore
// +build ignore

package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/unidoc/unipdf/v3/extractor"
	"github.com/unidoc/unipdf/v3/model"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Println("用法: go run extract_text.go <PDF文件>")
		os.Exit(1)
	}

	f, err := os.Open(os.Args[1])
	if err != nil {
		fmt.Printf("错误: %v\n", err)
		os.Exit(1)
	}
	defer f.Close()

	pdfReader, err := model.NewPdfReader(f)
	if err != nil {
		fmt.Printf("错误: %v\n", err)
		os.Exit(1)
	}

	numPages, _ := pdfReader.GetNumPages()
	fmt.Printf("总页数: %d\n\n", numPages)

	for pageNum := 1; pageNum <= numPages; pageNum++ {
		page, err := pdfReader.GetPage(pageNum)
		if err != nil {
			continue
		}

		ex, err := extractor.New(page)
		if err != nil {
			continue
		}

		text, err := ex.ExtractText()
		if err != nil {
			continue
		}

		fmt.Printf("========== 第 %d 页 ==========\n", pageNum)

		// 查找包含数字开头的行
		lines := strings.Split(text, "\n")
		for i, line := range lines {
			trimmed := strings.TrimSpace(line)
			if len(trimmed) > 0 && trimmed[0] >= '0' && trimmed[0] <= '9' {
				fmt.Printf("行 %d: %s\n", i+1, trimmed)
			}
		}
		fmt.Println()
	}
}
