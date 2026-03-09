// Package main implements a PDF content extraction tool that extracts structured content
// from PDF documents, including headings, body text, headers, and footers.
//
// The tool supports extracting headings from PDF outlines (table of contents/bookmarks)
// and text content, recognizing various heading formats (e.g., "1. ", "1.1 ", "2.3.4 ").
// It can also identify and extract body text, headers, and footers based on position
// and repetition patterns across pages.
//
// Usage:
//
//	pdf-parser [options] <PDF file path>
//
// Options:
//
//	-h, --help          Show help information
//	-d, --debug         Enable debug mode
//	--extract-all       Extract all content types (default)
//	--extract-body      Extract only body text
//	--extract-header    Extract only headers
//	--extract-footer    Extract only footers
//	--format text       Output in text format (default)
//	--format json       Output in JSON format
package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"regexp"
	"strings"
	"time"

	"github.com/unidoc/unipdf/v3/extractor"
	"github.com/unidoc/unipdf/v3/model"
)

// Heading represents a heading in a PDF document.
// It contains the heading text, hierarchical level, and page number.
type Heading struct {
	Title string // Title text of the heading
	Level int    // Hierarchical level (1, 2, 3, ...)
	Page  int64  // Page number where the heading appears
}

// BodyText represents body text content in a PDF document.
// It contains the text content and the page number where it appears.
type BodyText struct {
	Content string // Body text content
	Page    int64  // Page number where the content appears
}

// HeaderFooter represents a header or footer in a PDF document.
// It contains the text content, the range of pages where it appears,
// and the type ("header" or "footer").
type HeaderFooter struct {
	Content   string  // Header or footer text content
	PageRange []int64 // Range of page numbers where it appears
	Type      string  // Type: "header" or "footer"
}

// TextBlock represents a text block in a PDF document (internal use).
// It contains the text content, page number, Y coordinate, and page height
// for position-based classification.
type TextBlock struct {
	Content string  // Text content
	Page    int64   // Page number where the text appears
	YPos    float64 // Y coordinate (for position-based classification)
	Height  float64 // Page height
}

// PDFContent represents all extracted content from a PDF document.
// It contains headings, body text, headers, and footers.
type PDFContent struct {
	Headings []Heading      // List of headings
	Body     []BodyText     // List of body text segments
	Headers  []HeaderFooter // List of headers
	Footers  []HeaderFooter // List of footers
}

// FormatOptions specifies which content types to display in the output.
// Each boolean field controls whether the corresponding content type is shown.
type FormatOptions struct {
	ShowHeadings bool // Whether to show headings
	ShowBody     bool // Whether to show body text
	ShowHeaders  bool // Whether to show headers
	ShowFooters  bool // Whether to show footers
}

// ContentClassifier classifies PDF text blocks by type based on position and repetition.
// It uses configurable thresholds to determine header and footer regions.
type ContentClassifier struct {
	HeaderThreshold float64 // Header region threshold (default 0.15, top 15% of page)
	FooterThreshold float64 // Footer region threshold (default 0.15, bottom 15% of page)
}

// NewContentClassifier creates a ContentClassifier with default thresholds.
// The default thresholds are 0.15 (15%) for both header and footer regions.
func NewContentClassifier() *ContentClassifier {
	return &ContentClassifier{
		HeaderThreshold: 0.15,
		FooterThreshold: 0.15,
	}
}

// ClassifyTextBlock classifies a text block based on its position on the page.
// It returns one of: "header", "footer", "heading", or "body".
// Classification is based on the relative Y position of the text block:
//   - Top 15% (by default) of the page is classified as "header"
//   - Bottom 15% (by default) of the page is classified as "footer"
//   - Middle region defaults to "body" (heading detection is handled separately)
func (c *ContentClassifier) ClassifyTextBlock(block TextBlock) string {
	// 计算相对位置（0.0 到 1.0）
	relativePos := block.YPos / block.Height

	// 判断是否在页眉区域（页面顶部）
	// 在 PDF 坐标系中，Y 坐标从底部开始，所以顶部是较大的值
	if relativePos > (1.0 - c.HeaderThreshold) {
		return "header"
	}

	// 判断是否在页脚区域（页面底部）
	if relativePos < c.FooterThreshold {
		return "footer"
	}

	// 默认归类为正文（标题识别由其他逻辑处理）
	return "body"
}

// DetectRepeatingPatterns detects text patterns that repeat across multiple pages.
// It returns a map where keys are text content and values are lists of page numbers
// where the content appears. Only patterns that appear on at least 3 different pages
// are included in the result.
func (c *ContentClassifier) DetectRepeatingPatterns(blocks []TextBlock) map[string][]int64 {
	// 使用 map 统计每个内容出现的页码
	patterns := make(map[string][]int64)

	for _, block := range blocks {
		// 标准化文本内容（去除首尾空白）
		content := strings.TrimSpace(block.Content)
		if content == "" {
			continue
		}

		// 记录该内容出现的页码
		patterns[content] = append(patterns[content], block.Page)
	}

	// 过滤出至少在 3 个页面重复出现的内容
	result := make(map[string][]int64)
	for content, pages := range patterns {
		// 去重页码
		uniquePages := make(map[int64]bool)
		for _, page := range pages {
			uniquePages[page] = true
		}

		// 如果在至少 3 个不同页面出现，则认为是重复模式
		if len(uniquePages) >= 3 {
			// 转换回切片并排序
			pageList := make([]int64, 0, len(uniquePages))
			for page := range uniquePages {
				pageList = append(pageList, page)
			}
			result[content] = pageList
		}
	}

	return result
}

// extractTextBlocks extracts text blocks with position information from all PDF pages.
// It returns a list of TextBlock structures containing content, page number, Y coordinate,
// and page height for each text block. This function handles extraction failures gracefully
// by logging warnings and continuing with other pages.
func extractTextBlocks(pdfReader *model.PdfReader) ([]TextBlock, error) {
	var blocks []TextBlock

	numPages, err := pdfReader.GetNumPages()
	if err != nil {
		return nil, fmt.Errorf("无法获取 PDF 页数: %w", err)
	}

	for pageNum := 1; pageNum <= numPages; pageNum++ {
		page, err := pdfReader.GetPage(pageNum)
		if err != nil {
			// 记录错误但继续处理其他页面
			log.Printf("警告: 无法读取第 %d 页: %v", pageNum, err)
			continue
		}

		// 获取页面尺寸
		mediaBox, err := page.GetMediaBox()
		if err != nil {
			log.Printf("警告: 无法获取第 %d 页的页面尺寸: %v", pageNum, err)
			continue
		}
		pageHeight := mediaBox.Ury - mediaBox.Lly

		// 创建文本提取器
		ex, err := extractor.New(page)
		if err != nil {
			log.Printf("警告: 无法创建第 %d 页的文本提取器: %v", pageNum, err)
			continue
		}

		// 提取带位置信息的文本
		pageText, _, _, err := ex.ExtractPageText()
		if err != nil {
			log.Printf("警告: 无法从第 %d 页提取文本: %v", pageNum, err)
			continue
		}

		// 获取文本标记（包含位置信息）
		textMarks := pageText.Marks()
		if textMarks == nil || textMarks.Len() == 0 {
			// 如果没有文本标记，尝试提取纯文本
			text := pageText.Text()
			if text != "" {
				blocks = append(blocks, TextBlock{
					Content: text,
					Page:    int64(pageNum),
					YPos:    pageHeight / 2, // 默认位于页面中间
					Height:  pageHeight,
				})
			}
			continue
		}

		// 按行组织文本块
		// 使用 map 按 Y 坐标分组文本
		lineMap := make(map[float64][]string)
		var yPositions []float64

		for _, mark := range textMarks.Elements() {
			// 使用 BBox 的 Y 坐标（底部）
			yPos := mark.BBox.Lly
			// 将相近的 Y 坐标归为同一行（容差 2.0）
			found := false
			for _, existingY := range yPositions {
				if abs(existingY-yPos) < 2.0 {
					lineMap[existingY] = append(lineMap[existingY], mark.Text)
					found = true
					break
				}
			}
			if !found {
				yPositions = append(yPositions, yPos)
				lineMap[yPos] = []string{mark.Text}
			}
		}

		// 为每一行创建一个 TextBlock
		for _, yPos := range yPositions {
			lineText := strings.Join(lineMap[yPos], "")
			lineText = strings.TrimSpace(lineText)
			if lineText != "" {
				blocks = append(blocks, TextBlock{
					Content: lineText,
					Page:    int64(pageNum),
					YPos:    yPos,
					Height:  pageHeight,
				})
			}
		}
	}

	return blocks, nil
}

// abs returns the absolute value of a float64 number.
func abs(x float64) float64 {
	if x < 0 {
		return -x
	}
	return x
}

// extractBody extracts body text content from a PDF document.
// It excludes headings, headers, and footers, keeping only body text.
// The body text is organized by page number, preserving paragraph structure and line breaks.
// Cross-page text connections are handled intelligently based on sentence endings.
func extractBody(pdfReader *model.PdfReader, classifier *ContentClassifier) ([]BodyText, error) {
	// 提取所有文本块
	blocks, err := extractTextBlocks(pdfReader)
	if err != nil {
		return nil, fmt.Errorf("提取文本块失败: %w", err)
	}

	// 按页码组织文本块
	pageBlocksMap := make(map[int64][]TextBlock)
	for _, block := range blocks {
		pageBlocksMap[block.Page] = append(pageBlocksMap[block.Page], block)
	}

	// 提取标题模式以便过滤
	headingPatterns := []*regexp.Regexp{
		regexp.MustCompile(`^(\d+)\.\s+(.+)$`),
		regexp.MustCompile(`^(\d+\.\d+(?:\.\d+)*)\s+(.+)$`),
		regexp.MustCompile(`^(\d+\.\d+(?:\.\d+)*)\.\s+(.+)$`),
	}

	// 检测重复模式（页眉和页脚）
	repeatingPatterns := classifier.DetectRepeatingPatterns(blocks)

	var bodyTexts []BodyText

	// 获取所有页码并排序
	numPages, err := pdfReader.GetNumPages()
	if err != nil {
		return nil, fmt.Errorf("无法获取 PDF 页数: %w", err)
	}

	// 按页码顺序处理
	for pageNum := 1; pageNum <= numPages; pageNum++ {
		pageBlocks := pageBlocksMap[int64(pageNum)]
		if len(pageBlocks) == 0 {
			continue
		}

		var pageBodyLines []string

		for _, block := range pageBlocks {
			// 使用分类器判断内容类型
			blockType := classifier.ClassifyTextBlock(block)

			// 跳过页眉和页脚
			if blockType == "header" || blockType == "footer" {
				continue
			}

			// 检查是否是重复模式（页眉/页脚）
			content := strings.TrimSpace(block.Content)
			if _, isRepeating := repeatingPatterns[content]; isRepeating {
				continue
			}

			// 检查是否是标题
			isHeading := false
			for _, pattern := range headingPatterns {
				if pattern.MatchString(content) {
					isHeading = true
					break
				}
			}
			if isHeading {
				continue
			}

			// 这是正文内容
			pageBodyLines = append(pageBodyLines, content)
		}

		// 如果该页有正文内容，组织成 BodyText
		if len(pageBodyLines) > 0 {
			// 保持段落结构：用换行符连接各行
			bodyContent := strings.Join(pageBodyLines, "\n")

			// 处理跨页连接：如果上一页的正文以不完整的句子结尾，
			// 且当前页以小写字母开头，则可能是跨页连接
			if len(bodyTexts) > 0 {
				lastBody := &bodyTexts[len(bodyTexts)-1]
				lastContent := strings.TrimSpace(lastBody.Content)

				// 检查上一页是否以句子结束符结尾
				endsWithPunctuation := false
				if len(lastContent) > 0 {
					lastRunes := []rune(lastContent)
					if len(lastRunes) > 0 {
						lastChar := lastRunes[len(lastRunes)-1]
						endsWithPunctuation = lastChar == '.' || lastChar == '。' ||
							lastChar == '!' || lastChar == '！' ||
							lastChar == '?' || lastChar == '？'
					}
				}

				// 如果上一页不以句子结束符结尾，且当前页内容不为空
				// 则可能需要连接（添加空格而不是换行）
				if !endsWithPunctuation && len(bodyContent) > 0 {
					// 检查是否是中文（中文不需要空格）
					firstRunes := []rune(bodyContent)
					if len(firstRunes) > 0 {
						firstChar := firstRunes[0]
						if firstChar < 128 { // ASCII 字符，可能是英文
							lastBody.Content = lastContent + " " + bodyContent
							continue
						}
						// 中文直接连接
						lastBody.Content = lastContent + bodyContent
						continue
					}
				}
			}

			bodyTexts = append(bodyTexts, BodyText{
				Content: bodyContent,
				Page:    int64(pageNum),
			})
		}
	}

	return bodyTexts, nil
}

// extractHeaders extracts headers from a PDF document.
// It identifies repeating text at the top of pages as headers, based on position
// and repetition patterns. Supports different headers for odd and even pages.
func extractHeaders(pdfReader *model.PdfReader, classifier *ContentClassifier) ([]HeaderFooter, error) {
	// 提取所有文本块
	blocks, err := extractTextBlocks(pdfReader)
	if err != nil {
		return nil, fmt.Errorf("提取文本块失败: %w", err)
	}

	// 筛选出位于页眉区域的文本块
	var headerBlocks []TextBlock
	for _, block := range blocks {
		blockType := classifier.ClassifyTextBlock(block)
		if blockType == "header" {
			headerBlocks = append(headerBlocks, block)
		}
	}

	// 如果没有页眉区域的文本块，返回空列表
	if len(headerBlocks) == 0 {
		return []HeaderFooter{}, nil
	}

	// 检测重复模式
	repeatingPatterns := classifier.DetectRepeatingPatterns(headerBlocks)

	// 如果没有重复模式，返回空列表
	if len(repeatingPatterns) == 0 {
		return []HeaderFooter{}, nil
	}

	// 将重复模式转换为 HeaderFooter 结构
	var headers []HeaderFooter

	// 检查是否有奇偶页不同的页眉
	// 为每个重复模式检查其页码分布
	for content, pages := range repeatingPatterns {
		// 检查是否只出现在奇数页或偶数页
		oddPages := []int64{}
		evenPages := []int64{}

		for _, page := range pages {
			if page%2 == 1 {
				oddPages = append(oddPages, page)
			} else {
				evenPages = append(evenPages, page)
			}
		}

		// 如果只出现在奇数页或偶数页，标记为奇偶页页眉
		// 但仍然记录所有出现的页码
		headers = append(headers, HeaderFooter{
			Content:   content,
			PageRange: pages,
			Type:      "header",
		})
	}

	return headers, nil
}

// extractFooters extracts footers from a PDF document.
// It identifies repeating text at the bottom of pages, including page number recognition.
// Supports different footers for odd and even pages, and detects page number sequences.
func extractFooters(pdfReader *model.PdfReader, classifier *ContentClassifier) ([]HeaderFooter, error) {
	// 提取所有文本块
	blocks, err := extractTextBlocks(pdfReader)
	if err != nil {
		return nil, fmt.Errorf("提取文本块失败: %w", err)
	}

	// 筛选出位于页脚区域的文本块
	var footerBlocks []TextBlock
	for _, block := range blocks {
		blockType := classifier.ClassifyTextBlock(block)
		if blockType == "footer" {
			footerBlocks = append(footerBlocks, block)
		}
	}

	// 如果没有页脚区域的文本块，返回空列表
	if len(footerBlocks) == 0 {
		return []HeaderFooter{}, nil
	}

	// 检测页码序列
	isPageNumSeq, pageNumPattern := detectPageNumberSequence(footerBlocks)

	// 检测重复模式
	repeatingPatterns := classifier.DetectRepeatingPatterns(footerBlocks)

	// 如果既没有页码序列也没有重复模式，返回空列表
	if !isPageNumSeq && len(repeatingPatterns) == 0 {
		return []HeaderFooter{}, nil
	}

	// 将重复模式转换为 HeaderFooter 结构
	var footers []HeaderFooter

	// 如果检测到页码序列，创建一个页脚项
	if isPageNumSeq && pageNumPattern != "" {
		// 收集所有包含页码的页面
		var pageNumPages []int64
		for _, block := range footerBlocks {
			isPageNum, _ := isPageNumberPattern(block.Content, block.Page)
			if isPageNum {
				pageNumPages = append(pageNumPages, block.Page)
			}
		}

		if len(pageNumPages) > 0 {
			footers = append(footers, HeaderFooter{
				Content:   pageNumPattern,
				PageRange: pageNumPages,
				Type:      "footer",
			})
		}
	}

	// 处理其他重复模式（非页码的页脚）
	for content, pages := range repeatingPatterns {
		// 检查是否已经作为页码模式处理过
		// 如果内容与页码模式相似，跳过
		if isPageNumSeq && pageNumPattern != "" {
			// 检查是否为页码模式的实例
			isPageNum, _ := isPageNumberPattern(content, pages[0])
			if isPageNum {
				continue // 已经作为页码序列处理过了
			}
		}

		// 检查是否只出现在奇数页或偶数页
		oddPages := []int64{}
		evenPages := []int64{}

		for _, page := range pages {
			if page%2 == 1 {
				oddPages = append(oddPages, page)
			} else {
				evenPages = append(evenPages, page)
			}
		}

		// 添加页脚项
		footers = append(footers, HeaderFooter{
			Content:   content,
			PageRange: pages,
			Type:      "footer",
		})
	}

	return footers, nil
}

// isPageNumberPattern detects whether footer text contains a page number pattern.
// It recognizes various page number formats and returns whether it's a page number
// and the extracted page number value. The function checks if the extracted number
// is close to the actual page number (within ±10 pages).
func isPageNumberPattern(content string, pageNum int64) (bool, int64) {
	// 去除首尾空白
	content = strings.TrimSpace(content)

	// 常见页码模式的正则表达式
	patterns := []*regexp.Regexp{
		// 纯数字: "1", "2", "3"
		regexp.MustCompile(`^(\d+)$`),
		// 带前缀: "第 1 页", "Page 1", "- 1 -"
		regexp.MustCompile(`(?:第|Page|page|P\.|p\.)\s*(\d+)`),
		// 带后缀: "1 页", "1 of 10"
		regexp.MustCompile(`^(\d+)\s*(?:页|of)`),
		// 带分隔符: "- 1 -", "| 1 |"
		regexp.MustCompile(`[-|]\s*(\d+)\s*[-|]`),
		// 罗马数字页码: "i", "ii", "iii", "iv", "v"
		regexp.MustCompile(`^([ivxlcdm]+)$`),
	}

	for _, pattern := range patterns {
		matches := pattern.FindStringSubmatch(content)
		if len(matches) >= 2 {
			numStr := matches[1]

			// 尝试解析为数字
			var num int64
			_, err := fmt.Sscanf(numStr, "%d", &num)
			if err == nil {
				// 检查数字是否与页码接近（允许一定偏差，因为可能有封面页等）
				// 如果数字在合理范围内（页码 ± 10），认为是页码
				if num > 0 && abs(float64(num-pageNum)) <= 10 {
					return true, num
				}
			}

			// 处理罗马数字
			if isRomanNumeral(numStr) {
				romanNum := romanToInt(numStr)
				if romanNum > 0 {
					return true, int64(romanNum)
				}
			}
		}
	}

	return false, 0
}

// detectPageNumberSequence detects whether a set of footer text blocks forms
// an incrementing page number sequence.
// Parameters:
//   - footerBlocks: list of footer text blocks
//
// Returns:
//   - bool: whether it's a page number sequence
//   - string: page number pattern template (for identifying footers with the same pattern)
func detectPageNumberSequence(footerBlocks []TextBlock) (bool, string) {
	if len(footerBlocks) < 3 {
		// 至少需要 3 个样本才能判断是否为递增序列
		return false, ""
	}

	// 按页码排序
	sortedBlocks := make([]TextBlock, len(footerBlocks))
	copy(sortedBlocks, footerBlocks)

	// 简单冒泡排序（因为数据量小）
	for i := 0; i < len(sortedBlocks)-1; i++ {
		for j := 0; j < len(sortedBlocks)-i-1; j++ {
			if sortedBlocks[j].Page > sortedBlocks[j+1].Page {
				sortedBlocks[j], sortedBlocks[j+1] = sortedBlocks[j+1], sortedBlocks[j]
			}
		}
	}

	// 检查是否形成递增序列
	consecutiveCount := 0
	var pattern string

	for i := 0; i < len(sortedBlocks); i++ {
		block := sortedBlocks[i]
		isPageNum, extractedNum := isPageNumberPattern(block.Content, block.Page)

		if isPageNum {
			// 检查提取的页码是否与实际页码匹配或接近
			if abs(float64(extractedNum-block.Page)) <= 2 {
				consecutiveCount++

				// 记录页码模式（将数字替换为占位符）
				if pattern == "" {
					// 创建模式模板：将数字替换为 {n}
					pattern = regexp.MustCompile(`\d+`).ReplaceAllString(block.Content, "{n}")
				}
			}
		}
	}

	// 如果至少有 3 个连续的页码匹配，认为是页码序列
	if consecutiveCount >= 3 {
		return true, pattern
	}

	return false, ""
}

// isRomanNumeral checks whether a string is a Roman numeral.
// It returns true if the string contains only valid Roman numeral characters (i, v, x, l, c, d, m).
func isRomanNumeral(s string) bool {
	s = strings.ToLower(s)
	for _, c := range s {
		if c != 'i' && c != 'v' && c != 'x' && c != 'l' && c != 'c' && c != 'd' && c != 'm' {
			return false
		}
	}
	return len(s) > 0
}

// romanToInt converts a Roman numeral string to an integer.
// It supports standard Roman numerals (i, v, x, l, c, d, m) and handles
// subtractive notation (e.g., IV = 4, IX = 9).
func romanToInt(s string) int {
	s = strings.ToLower(s)
	romanMap := map[rune]int{
		'i': 1,
		'v': 5,
		'x': 10,
		'l': 50,
		'c': 100,
		'd': 500,
		'm': 1000,
	}

	result := 0
	prevValue := 0

	for i := len(s) - 1; i >= 0; i-- {
		value := romanMap[rune(s[i])]
		if value < prevValue {
			result -= value
		} else {
			result += value
		}
		prevValue = value
	}

	return result
}

// extractAllContent extracts all types of content from a PDF file.
// It returns a PDFContent structure containing headings, body text, headers, and footers.
// If partial extraction fails, it still returns successfully extracted content with warnings logged.
// This function provides fault tolerance by continuing extraction even when individual components fail.
func extractAllContent(pdfPath string) (*PDFContent, error) {
	// 打开 PDF 文件
	f, err := os.Open(pdfPath)
	if err != nil {
		return nil, fmt.Errorf("无法打开文件 %s: %w", pdfPath, err)
	}
	defer f.Close()

	// 创建 PDF 阅读器
	pdfReader, err := model.NewPdfReader(f)
	if err != nil {
		return nil, fmt.Errorf("无法读取 PDF 文件 %s: %w", pdfPath, err)
	}

	// 初始化结果结构
	content := &PDFContent{
		Headings: []Heading{},
		Body:     []BodyText{},
		Headers:  []HeaderFooter{},
		Footers:  []HeaderFooter{},
	}

	// 创建内容分类器
	classifier := NewContentClassifier()

	// 提取标题（使用现有的 extractHeadings 逻辑）
	// 首先尝试从大纲中提取标题
	outlines, err := pdfReader.GetOutlines()
	if err == nil && outlines != nil {
		extractOutlineItems(outlines.Entries, 1, &content.Headings)
	} else if err != nil {
		log.Printf("警告: 无法读取 PDF 大纲: %v", err)
	}

	// 然后从文本内容中提取标题
	textHeadings, err := extractHeadingsFromText(pdfReader)
	if err != nil {
		log.Printf("警告: 从文本提取标题失败: %v", err)
	} else {
		// 合并标题，去重
		content.Headings = mergeHeadings(content.Headings, textHeadings)
	}

	// 提取正文
	body, err := extractBody(pdfReader, classifier)
	if err != nil {
		log.Printf("警告: 提取正文失败: %v", err)
		// 确保 Body 字段被初始化为空切片而不是 nil
		content.Body = []BodyText{}
	} else {
		// 确保即使 body 为 nil 也设置为空切片
		if body == nil {
			content.Body = []BodyText{}
		} else {
			content.Body = body
		}
	}

	// 提取页眉
	headers, err := extractHeaders(pdfReader, classifier)
	if err != nil {
		log.Printf("警告: 提取页眉失败: %v", err)
		// 确保 Headers 字段被初始化为空切片而不是 nil
		content.Headers = []HeaderFooter{}
	} else {
		// 确保即使 headers 为 nil 也设置为空切片
		if headers == nil {
			content.Headers = []HeaderFooter{}
		} else {
			content.Headers = headers
		}
	}

	// 提取页脚
	footers, err := extractFooters(pdfReader, classifier)
	if err != nil {
		log.Printf("警告: 提取页脚失败: %v", err)
		// 确保 Footers 字段被初始化为空切片而不是 nil
		content.Footers = []HeaderFooter{}
	} else {
		// 确保即使 footers 为 nil 也设置为空切片
		if footers == nil {
			content.Footers = []HeaderFooter{}
		} else {
			content.Footers = footers
		}
	}

	return content, nil
}

// extractHeadings extracts headings from a PDF file.
// It first attempts to extract headings from the PDF outline (table of contents/bookmarks),
// then extracts headings from text content using pattern matching.
// The results are merged and deduplicated, with outline headings taking priority.
func extractHeadings(pdfPath string) ([]Heading, error) {
	// 打开 PDF 文件
	f, err := os.Open(pdfPath)
	if err != nil {
		return nil, fmt.Errorf("无法打开文件 %s: %w", pdfPath, err)
	}
	defer f.Close()

	// 创建 PDF 阅读器
	pdfReader, err := model.NewPdfReader(f)
	if err != nil {
		return nil, fmt.Errorf("无法读取 PDF 文件 %s: %w", pdfPath, err)
	}

	var headings []Heading

	// 首先尝试从大纲中提取标题
	outlines, err := pdfReader.GetOutlines()
	if err == nil && outlines != nil {
		extractOutlineItems(outlines.Entries, 1, &headings)
	} else if err != nil {
		log.Printf("警告: 无法读取 PDF 大纲: %v", err)
	}

	// 然后从文本内容中提取标题
	textHeadings, err := extractHeadingsFromText(pdfReader)
	if err != nil {
		log.Printf("警告: 从文本提取标题失败: %v", err)
		return headings, nil // 如果文本提取失败，至少返回大纲中的标题
	}

	// 合并标题，去重
	headings = mergeHeadings(headings, textHeadings)

	return headings, nil
}

// extractHeadingsFromText extracts headings from PDF text content using pattern matching.
// It recognizes various heading formats:
//   - "1. ", "7. " (single-level)
//   - "1.1 ", "2.3.4 " (multi-level without trailing dot)
//   - "1.1.", "2.3.4." (multi-level with trailing dot)
//
// The heading level is calculated based on the number of dots in the number part.
func extractHeadingsFromText(pdfReader *model.PdfReader) ([]Heading, error) {
	var headings []Heading

	numPages, err := pdfReader.GetNumPages()
	if err != nil {
		return nil, fmt.Errorf("无法获取 PDF 页数: %w", err)
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
			log.Printf("警告: 无法读取第 %d 页: %v", pageNum, err)
			continue
		}

		ex, err := extractor.New(page)
		if err != nil {
			log.Printf("警告: 无法创建第 %d 页的文本提取器: %v", pageNum, err)
			continue
		}

		text, err := ex.ExtractText()
		if err != nil {
			log.Printf("警告: 无法从第 %d 页提取文本: %v", pageNum, err)
			continue
		}

		// 按行分割文本 - 使用 SplitSeq 提高效率（Go 1.24+）
		for line := range strings.SplitSeq(text, "\n") {
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

// mergeHeadings merges two heading lists and removes duplicates.
// Outline headings take priority over text headings. Duplicates are detected
// by comparing page numbers and the first 10 characters of heading titles.
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

// extractOutlineItems recursively extracts outline items from a PDF outline.
// It processes each outline item and its nested children, assigning hierarchical
// levels starting from the provided level parameter.
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

// formatAsText formats PDFContent as text output.
// It selectively displays content based on FormatOptions, adds section markers
// for different content types, adds hierarchical indentation for headings,
// displays statistics, and maintains backward compatibility with the original format.
func formatAsText(content *PDFContent, options FormatOptions) string {
	var output strings.Builder

	// 预估容量：假设平均每个标题 100 字符，每段正文 500 字符
	estimatedSize := len(content.Headings)*100 + len(content.Body)*500 + len(content.Headers)*100 + len(content.Footers)*100
	output.Grow(estimatedSize)

	// 标题部分
	if options.ShowHeadings && len(content.Headings) > 0 {
		// 添加分区标记
		output.WriteString(fmt.Sprintf("=== 标题 (%d 个) ===\n\n", len(content.Headings)))

		// 输出每个标题，保持向后兼容的格式
		for _, h := range content.Headings {
			// 根据级别添加缩进
			var indent strings.Builder
			indent.Grow((h.Level - 1) * 2)
			for i := 1; i < h.Level; i++ {
				indent.WriteString("  ")
			}
			output.WriteString(fmt.Sprintf("%s级别 %d: %s (第 %d 页)\n", indent.String(), h.Level, h.Title, h.Page))
		}
		output.WriteString("\n")
	}

	// 页眉部分
	if options.ShowHeaders && len(content.Headers) > 0 {
		output.WriteString("=== 页眉 ===\n\n")

		for _, header := range content.Headers {
			// 格式化页码范围
			pageRangeStr := formatPageRange(header.PageRange)
			output.WriteString(fmt.Sprintf("\"%s\" (%s)\n", header.Content, pageRangeStr))
		}
		output.WriteString("\n")
	}

	// 页脚部分
	if options.ShowFooters && len(content.Footers) > 0 {
		output.WriteString("=== 页脚 ===\n\n")

		for _, footer := range content.Footers {
			// 格式化页码范围
			pageRangeStr := formatPageRange(footer.PageRange)
			output.WriteString(fmt.Sprintf("\"%s\" (%s)\n", footer.Content, pageRangeStr))
		}
		output.WriteString("\n")
	}

	// 正文部分
	if options.ShowBody && len(content.Body) > 0 {
		for _, body := range content.Body {
			output.WriteString(fmt.Sprintf("=== 正文 (第 %d 页) ===\n", body.Page))
			output.WriteString(body.Content)
			output.WriteString("\n\n")
		}
	}

	return output.String()
}

// formatPageRange formats a page range as a string.
// For example: [1, 2, 3, 5, 6, 7] -> "第 1-3, 5-7 页"
// It handles sorting, deduplication, and grouping consecutive pages into ranges.
func formatPageRange(pages []int64) string {
	if len(pages) == 0 {
		return ""
	}

	// 排序页码
	sortedPages := make([]int64, len(pages))
	copy(sortedPages, pages)

	// 简单冒泡排序
	for i := 0; i < len(sortedPages)-1; i++ {
		for j := 0; j < len(sortedPages)-i-1; j++ {
			if sortedPages[j] > sortedPages[j+1] {
				sortedPages[j], sortedPages[j+1] = sortedPages[j+1], sortedPages[j]
			}
		}
	}

	// 去重
	uniquePages := []int64{sortedPages[0]}
	for i := 1; i < len(sortedPages); i++ {
		if sortedPages[i] != sortedPages[i-1] {
			uniquePages = append(uniquePages, sortedPages[i])
		}
	}

	// 构建范围字符串
	var ranges []string
	start := uniquePages[0]
	end := uniquePages[0]

	for i := 1; i < len(uniquePages); i++ {
		if uniquePages[i] == end+1 {
			// 连续页码，扩展范围
			end = uniquePages[i]
		} else {
			// 不连续，保存当前范围并开始新范围
			if start == end {
				ranges = append(ranges, fmt.Sprintf("%d", start))
			} else {
				ranges = append(ranges, fmt.Sprintf("%d-%d", start, end))
			}
			start = uniquePages[i]
			end = uniquePages[i]
		}
	}

	// 添加最后一个范围
	if start == end {
		ranges = append(ranges, fmt.Sprintf("%d", start))
	} else {
		ranges = append(ranges, fmt.Sprintf("%d-%d", start, end))
	}

	return "第 " + strings.Join(ranges, ", ") + " 页"
}

// formatAsJSON formats PDFContent as a JSON string.
// It uses json.MarshalIndent to generate formatted JSON output with 2-space indentation.
// Returns an error if serialization fails.
func formatAsJSON(content *PDFContent) (string, error) {
	// 使用 MarshalIndent 生成格式化的 JSON，缩进使用 2 个空格
	jsonBytes, err := json.MarshalIndent(content, "", "  ")
	if err != nil {
		return "", fmt.Errorf("无法序列化为 JSON: %w", err)
	}

	return string(jsonBytes), nil
}

// printUsage prints the usage information for the PDF content extraction tool.
// It displays command-line options, content extraction options, output format options,
// examples, and descriptions of content types.
func printUsage() {
	fmt.Println("PDF 内容提取工具")
	fmt.Println()
	fmt.Println("用法:")
	fmt.Println("  pdf-parser [选项] <PDF文件路径>")
	fmt.Println()
	fmt.Println("参数:")
	fmt.Println("  <PDF文件路径>       要解析的 PDF 文件路径（必需）")
	fmt.Println()
	fmt.Println("选项:")
	fmt.Println("  -h, --help          显示此帮助信息并退出")
	fmt.Println("  -d, --debug         启用调试模式，显示详细的提取统计信息")
	fmt.Println("  --sleep [duration]  Sleep 模式，程序持续运行并定期打印日志")
	fmt.Println("                      可选参数：持续时间（如 30m, 2h, 1h30m）")
	fmt.Println("                      默认：1h（1小时）")
	fmt.Println("  --report-interval <duration>")
	fmt.Println("                      设置 sleep 模式下的日志打印间隔")
	fmt.Println("                      参数：时间间隔（如 30s, 1m, 5m）")
	fmt.Println("                      默认：60s（60秒）")
	fmt.Println()
	fmt.Println("内容提取选项:")
	fmt.Println("  --extract-all       提取所有内容类型（标题、正文、页眉、页脚）")
	fmt.Println("                      这是默认行为，如果未指定任何提取选项")
	fmt.Println("  --extract-title     仅提取标题内容")
	fmt.Println("  --extract-body      仅提取正文内容")
	fmt.Println("  --extract-header    仅提取页眉内容")
	fmt.Println("  --extract-footer    仅提取页脚内容")
	fmt.Println("  注意：可以组合多个提取选项，例如同时提取页眉和页脚")
	fmt.Println()
	fmt.Println("输出格式选项:")
	fmt.Println("  --format text       以文本格式输出（默认）")
	fmt.Println("  --format json       以 JSON 格式输出，便于程序处理")
	fmt.Println()
	fmt.Println("示例:")
	fmt.Println("  # 提取所有内容（默认行为）")
	fmt.Println("  pdf-parser document.pdf")
	fmt.Println()
	fmt.Println("  # 仅提取标题内容")
	fmt.Println("  pdf-parser --extract-title document.pdf")
	fmt.Println()
	fmt.Println("  # 仅提取正文内容")
	fmt.Println("  pdf-parser --extract-body document.pdf")
	fmt.Println()
	fmt.Println("  # 同时提取页眉和页脚")
	fmt.Println("  pdf-parser --extract-header --extract-footer document.pdf")
	fmt.Println()
	fmt.Println("  # 以 JSON 格式输出所有内容")
	fmt.Println("  pdf-parser --format json document.pdf")
	fmt.Println()
	fmt.Println("  # 启用调试模式查看提取统计")
	fmt.Println("  pdf-parser -d document.pdf")
	fmt.Println()
	fmt.Println("  # 组合多个选项：以 JSON 格式输出正文，并启用调试")
	fmt.Println("  pdf-parser -d --extract-body --format json document.pdf")
	fmt.Println()
	fmt.Println("  # Sleep 模式：处理完后持续运行 1 小时（默认）")
	fmt.Println("  pdf-parser --sleep document.pdf")
	fmt.Println()
	fmt.Println("  # Sleep 模式：自定义持续时间为 30 分钟")
	fmt.Println("  pdf-parser --sleep 30m document.pdf")
	fmt.Println()
	fmt.Println("  # Sleep 模式：持续运行 2 小时")
	fmt.Println("  pdf-parser --sleep 2h document.pdf")
	fmt.Println()
	fmt.Println("  # Sleep 模式：自定义报告间隔为 30 秒")
	fmt.Println("  pdf-parser --sleep --report-interval 30s document.pdf")
	fmt.Println()
	fmt.Println("  # Sleep 模式：持续 2 小时，每 5 分钟报告一次")
	fmt.Println("  pdf-parser --sleep 2h --report-interval 5m document.pdf")
	fmt.Println()
	fmt.Println("内容类型说明:")
	fmt.Println("  标题 (Headings)")
	fmt.Println("    从 PDF 大纲（目录/书签）和文本内容中识别的标题")
	fmt.Println("    支持多级标题格式：\"1. \", \"1.1 \", \"2.3.4 \" 等")
	fmt.Println()
	fmt.Println("  正文 (Body)")
	fmt.Println("    PDF 文档的主要文本内容")
	fmt.Println("    自动排除标题、页眉和页脚，保持段落结构")
	fmt.Println()
	fmt.Println("  页眉 (Headers)")
	fmt.Println("    出现在每页顶部的重复性文本")
	fmt.Println("    基于位置（页面顶部 15%）和跨页重复性识别")
	fmt.Println()
	fmt.Println("  页脚 (Footers)")
	fmt.Println("    出现在每页底部的重复性文本")
	fmt.Println("    基于位置（页面底部 15%）和跨页重复性识别")
	fmt.Println("    自动识别页码模式")
}

func main() {
	// 检查参数
	if len(os.Args) < 2 {
		printUsage()
		os.Exit(1)
	}

	// 解析命令行参数
	var pdfPath string
	var debugMode bool
	var sleepMode bool
	var sleepDuration time.Duration = 1 * time.Hour     // 默认 sleep 1小时
	var reportInterval time.Duration = 60 * time.Second // 默认每 60 秒报告一次
	var extractAll bool
	var extractTitle bool
	var extractBody bool
	var extractHeader bool
	var extractFooter bool
	var outputFormat string = "text" // 默认文本格式

	// 遍历参数
	i := 1
	for i < len(os.Args) {
		arg := os.Args[i]

		switch arg {
		case "-h", "--help":
			printUsage()
			os.Exit(0)

		case "-d", "--debug":
			debugMode = true
			i++

		case "--sleep":
			sleepMode = true
			// 检查是否有下一个参数作为持续时间
			if i+1 < len(os.Args) && !strings.HasPrefix(os.Args[i+1], "-") {
				i++
				duration, err := time.ParseDuration(os.Args[i])
				if err != nil {
					fmt.Printf("错误: 无效的 sleep 持续时间 '%s'，使用默认值 1h\n", os.Args[i])
				} else {
					sleepDuration = duration
				}
			}
			i++

		case "--report-interval":
			// 下一个参数应该是报告间隔
			if i+1 >= len(os.Args) {
				fmt.Println("错误: --report-interval 参数需要指定时间间隔（如 30s, 1m, 5m）")
				printUsage()
				os.Exit(1)
			}
			i++
			interval, err := time.ParseDuration(os.Args[i])
			if err != nil {
				fmt.Printf("错误: 无效的报告间隔 '%s'，使用默认值 60s\n", os.Args[i])
			} else {
				reportInterval = interval
			}
			i++

		case "--extract-all":
			extractAll = true
			i++

		case "--extract-title":
			extractTitle = true
			i++

		case "--extract-body":
			extractBody = true
			i++

		case "--extract-header":
			extractHeader = true
			i++

		case "--extract-footer":
			extractFooter = true
			i++

		case "--format":
			// 下一个参数应该是格式类型
			if i+1 >= len(os.Args) {
				fmt.Println("错误: --format 参数需要指定格式类型（text 或 json）")
				printUsage()
				os.Exit(1)
			}
			i++
			outputFormat = os.Args[i]
			if outputFormat != "text" && outputFormat != "json" {
				fmt.Printf("错误: 不支持的输出格式 '%s'，仅支持 'text' 或 'json'\n", outputFormat)
				os.Exit(1)
			}
			i++

		default:
			// 如果不是选项，则认为是 PDF 文件路径
			if !strings.HasPrefix(arg, "-") {
				if pdfPath != "" {
					fmt.Printf("错误: 指定了多个 PDF 文件路径\n")
					printUsage()
					os.Exit(1)
				}
				pdfPath = arg
				i++
			} else {
				fmt.Printf("错误: 未知选项 '%s'\n", arg)
				printUsage()
				os.Exit(1)
			}
		}
	}

	// 检查是否指定了 PDF 文件路径
	if pdfPath == "" {
		fmt.Println("错误: 未指定 PDF 文件路径")
		printUsage()
		os.Exit(1)
	}

	// 如果没有指定任何提取选项，默认为 extract-all
	if !extractAll && !extractTitle && !extractBody && !extractHeader && !extractFooter {
		extractAll = true
	}

	// 构建 FormatOptions
	formatOptions := FormatOptions{
		ShowHeadings: extractAll || extractTitle, // extract-all 或 extract-title 时显示标题
		ShowBody:     extractAll || extractBody,
		ShowHeaders:  extractAll || extractHeader,
		ShowFooters:  extractAll || extractFooter,
	}

	// 提取内容
	content, err := extractAllContent(pdfPath)
	if err != nil {
		log.Fatalf("错误: %v", err)
	}

	// 根据输出格式生成输出
	var output string
	if outputFormat == "json" {
		output, err = formatAsJSON(content)
		if err != nil {
			log.Fatalf("错误: %v", err)
		}
	} else {
		output = formatAsText(content, formatOptions)
	}

	// 输出结果
	fmt.Print(output)

	// 调试模式下显示额外信息
	if debugMode {
		fmt.Fprintf(os.Stderr, "\n[调试信息]\n")
		fmt.Fprintf(os.Stderr, "提取的标题数: %d\n", len(content.Headings))
		fmt.Fprintf(os.Stderr, "提取的正文段落数: %d\n", len(content.Body))
		fmt.Fprintf(os.Stderr, "提取的页眉数: %d\n", len(content.Headers))
		fmt.Fprintf(os.Stderr, "提取的页脚数: %d\n", len(content.Footers))
	}

	// Sleep 模式：持续运行并定期打印日志
	if sleepMode {
		fmt.Fprintf(os.Stderr, "\n[Sleep 模式] 程序将运行 %v，每 %v 打印一次日志\n", sleepDuration, reportInterval)
		startTime := time.Now()
		endTime := startTime.Add(sleepDuration)
		ticker := time.NewTicker(reportInterval)
		defer ticker.Stop()

		for {
			select {
			case <-ticker.C:
				elapsed := time.Since(startTime)
				remaining := sleepDuration - elapsed
				if remaining <= 0 {
					fmt.Fprintf(os.Stderr, "[%s] Sleep 时间已到，程序退出\n", time.Now().Format("2006-01-02 15:04:05"))
					return
				}
				fmt.Fprintf(os.Stderr, "[%s] 运行中... 已运行: %v, 剩余: %v\n",
					time.Now().Format("2006-01-02 15:04:05"),
					elapsed.Round(time.Second),
					remaining.Round(time.Second))
			case <-time.After(time.Until(endTime)):
				fmt.Fprintf(os.Stderr, "[%s] Sleep 时间已到，程序退出\n", time.Now().Format("2006-01-02 15:04:05"))
				return
			}
		}
	}
}
