package main

import (
	"fmt"
	"log"
	"os"
	"regexp"
	"strings"

	"github.com/unidoc/unipdf/v3/extractor"
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

	var headings []Heading

	// 首先尝试从大纲中提取标题
	outlines, err := pdfReader.GetOutlines()
	if err == nil && outlines != nil {
		extractOutlineItems(outlines.Entries, 1, &headings)
	}

	// 然后从文本内容中提取标题
	textHeadings, err := extractHeadingsFromText(pdfReader)
	if err != nil {
		return headings, nil // 如果文本提取失败，至少返回大纲中的标题
	}

	// 合并标题，去重
	headings = mergeHeadings(headings, textHeadings)

	return headings, nil
}

// extractHeadingsFromText 从 PDF 文本内容中提取标题
func extractHeadingsFromText(pdfReader *model.PdfReader) ([]Heading, error) {
	var headings []Heading

	numPages, err := pdfReader.GetNumPages()
	if err != nil {
		return nil, err
	}

	// 多个匹配标题的正则表达式
	patterns := []*regexp.Regexp{
		// 匹配 "1. ", "7. ", "8. " 等
		regexp.MustCompile(`^(\d+)\.\s+(.+)$`),
		// 匹配 "1.1 ", "2.3.4 " 等
		regexp.MustCompile(`^(\d+\.\d+(?:\.\d+)*)\s+(.+)$`),
		// 匹配 "1.1.", "2.3.4." 等（带结尾点号）
		regexp.MustCompile(`^(\d+\.\d+(?:\.\d+)*)\.\s+(.+)$`),
	}

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

		// 按行分割文本
		lines := strings.Split(text, "\n")
		for _, line := range lines {
			line = strings.TrimSpace(line)
			if line == "" {
				continue
			}

			// 尝试所有模式
			for _, pattern := range patterns {
				matches := pattern.FindStringSubmatch(line)
				if len(matches) >= 3 {
					numberPart := matches[1]

					// 计算标题级别（根据点号的数量）
					level := strings.Count(numberPart, ".") + 1

					headings = append(headings, Heading{
						Title: line,
						Level: level,
						Page:  int64(pageNum),
					})
					break // 匹配成功后跳出模式循环
				}
			}
		}
	}

	return headings, nil
}

// mergeHeadings 合并两个标题列表，去重
func mergeHeadings(outlineHeadings, textHeadings []Heading) []Heading {
	// 使用 map 来去重
	seen := make(map[string]bool)
	var result []Heading

	// 首先添加大纲中的标题
	for _, h := range outlineHeadings {
		key := fmt.Sprintf("%d-%s", h.Page, h.Title)
		if !seen[key] {
			seen[key] = true
			result = append(result, h)
		}
	}

	// 然后添加文本中提取的标题（如果不重复）
	for _, h := range textHeadings {
		// 检查是否已存在相似的标题
		found := false
		for _, existing := range result {
			// 如果页码相同且标题开头相似，认为是重复的
			if existing.Page == h.Page &&
				(strings.HasPrefix(existing.Title, h.Title[:min(10, len(h.Title))]) ||
					strings.HasPrefix(h.Title, existing.Title[:min(10, len(existing.Title))])) {
				found = true
				break
			}
		}
		if !found {
			result = append(result, h)
		}
	}

	return result
}

// min 返回两个整数中的较小值
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
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
	fmt.Println("  pdf-parser -d | --debug <PDF文件路径>")
	fmt.Println("  pdf-parser -h | --help")
	fmt.Println()
	fmt.Println("参数:")
	fmt.Println("  <PDF文件路径>    要解析的 PDF 文件路径")
	fmt.Println("  -d, --debug      调试模式，显示从文本提取的标题")
	fmt.Println("  -h, --help       显示此帮助信息")
	fmt.Println()
	fmt.Println("示例:")
	fmt.Println("  pdf-parser document.pdf")
	fmt.Println("  pdf-parser -d document.pdf")
	fmt.Println("  pdf-parser /path/to/file.pdf")
	fmt.Println()
	fmt.Println("说明:")
	fmt.Println("  此工具从 PDF 文件中提取标题信息")
	fmt.Println("  - 首先从 PDF 大纲（目录/书签）中提取")
	fmt.Println("  - 然后从文本内容中识别标题格式（如 '1. ', '1.1 ' 等）")
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

	// 处理调试模式
	debugMode := false
	pdfPath := os.Args[1]
	if os.Args[1] == "-d" || os.Args[1] == "--debug" {
		if len(os.Args) < 3 {
			printUsage()
			os.Exit(1)
		}
		debugMode = true
		pdfPath = os.Args[2]
	}

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

	if debugMode {
		fmt.Printf("找到 %d 个标题（包含从文本提取的）:\n\n", len(headings))
	} else {
		fmt.Printf("找到 %d 个标题:\n\n", len(headings))
	}

	for _, h := range headings {
		// 根据级别添加缩进
		indent := ""
		for i := 1; i < h.Level; i++ {
			indent += "  "
		}
		fmt.Printf("%s级别 %d: %s (第 %d 页)\n", indent, h.Level, h.Title, h.Page)
	}
}
