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
	extractOutlineItems(outlines.Entries, 1, &headings)

	return headings, nil
}

// extractOutlineItems 递归提取大纲项
func extractOutlineItems(items []*model.OutlineItem, level int, headings *[]Heading) {
	for _, item := range items {
		// 添加标题
		*headings = append(*headings, Heading{
			Title: item.Title,
			Level: level,
			Page:  item.Dest.Page,
		})

		// 递归处理子项
		if len(item.Entries) > 0 {
			extractOutlineItems(item.Entries, level+1, headings)
		}
	}
}

// printUsage 打印使用说明
func printUsage() {
	fmt.Println("PDF 标题解析工具")
	fmt.Println()
	fmt.Println("用法:")
	fmt.Println("  pdf-parser <PDF文件路径>")
	fmt.Println("  pdf-parser -h | --help")
	fmt.Println()
	fmt.Println("参数:")
	fmt.Println("  <PDF文件路径>    要解析的 PDF 文件路径")
	fmt.Println("  -h, --help       显示此帮助信息")
	fmt.Println()
	fmt.Println("示例:")
	fmt.Println("  pdf-parser document.pdf")
	fmt.Println("  pdf-parser /path/to/file.pdf")
	fmt.Println()
	fmt.Println("说明:")
	fmt.Println("  此工具从 PDF 文件中提取标题大纲（目录/书签）信息")
	fmt.Println("  输出包括标题文字、层级和所在页码")
}

func main() {
	// 检查参数
	if len(os.Args) < 2 {
		printUsage()
		os.Exit(1)
	}

	// 处理帮助参数
	if os.Args[1] == "-h" || os.Args[1] == "--help" {
		printUsage()
		os.Exit(0)
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
