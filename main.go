package main

import (
	"fmt"
	"log"
	"os"

	"github.com/unidoc/unipdf/v3/model"
)

// Heading 表示 PDF 中的标题
type Heading struct {
	Title string // 标题文字
	Level int    // 标题级别（1, 2, 3...）
	Page  int64  // 所在页码
}

// extractHeadings 从 PDF 文件中提取标题
func extractHeadings(pdfPath string) ([]Heading, error) {
	// 打开 PDF 文件
	f, err := os.Open(pdfPath)
	if err != nil {
		return nil, fmt.Errorf("无法打开文件: %w", err)
	}
	defer f.Close()

	// 创建 PDF 阅读器
	pdfReader, err := model.NewPdfReader(f)
	if err != nil {
		return nil, fmt.Errorf("无法读取 PDF: %w", err)
	}

	// 获取 PDF 大纲（目录/书签）
	outlines, err := pdfReader.GetOutlines()
	if err != nil {
		return nil, fmt.Errorf("无法获取大纲: %w", err)
	}

	if outlines == nil {
		return []Heading{}, nil
	}

	// 递归提取所有标题
	var headings []Heading
	extractOutlineItems(outlines.Entries(), 1, &headings, pdfReader)

	return headings, nil
}

// extractOutlineItems 递归提取大纲项
func extractOutlineItems(items []*model.OutlineItem, level int, headings *[]Heading, reader *model.PdfReader) {
	for _, item := range items {
		// 获取页码
		pageNum := getPageNumber(item, reader)

		// 添加标题
		*headings = append(*headings, Heading{
			Title: item.Title,
			Level: level,
			Page:  pageNum,
		})

		// 递归处理子项
		if len(item.Entries()) > 0 {
			extractOutlineItems(item.Entries(), level+1, headings, reader)
		}
	}
}

// getPageNumber 获取大纲项对应的页码
func getPageNumber(item *model.OutlineItem, reader *model.PdfReader) int64 {
	if item.Dest != nil {
		// 尝试从目标获取页码
		if page := item.Dest.GetPage(reader); page != nil {
			numPages, _ := reader.GetNumPages()
			for i := 1; i <= numPages; i++ {
				p, _ := reader.GetPage(i)
				if p == page {
					return int64(i)
				}
			}
		}
	}
	return 0
}

func main() {
	if len(os.Args) < 2 {
		fmt.Println("用法: go run main.go <PDF文件路径>")
		os.Exit(1)
	}

	pdfPath := os.Args[1]

	// 提取标题
	headings, err := extractHeadings(pdfPath)
	if err != nil {
		log.Fatalf("错误: %v", err)
	}

	// 输出结果
	if len(headings) == 0 {
		fmt.Println("该 PDF 文件没有标题大纲")
		return
	}

	fmt.Printf("找到 %d 个标题:\n\n", len(headings))
	for _, h := range headings {
		// 根据级别添加缩进
		indent := ""
		for i := 1; i < h.Level; i++ {
			indent += "  "
		}
		fmt.Printf("%s级别 %d: %s (第 %d 页)\n", indent, h.Level, h.Title, h.Page)
	}
}
