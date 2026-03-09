package main

import (
	"os"
	"strings"
	"testing"

	"github.com/unidoc/unipdf/v3/model"
)

func TestExtractHeadings135Title(t *testing.T) {
	// 测试文件路径
	pdfPath := "135_title.pdf"

	// 提取标题
	headings, err := extractHeadings(pdfPath)
	if err != nil {
		t.Fatalf("提取标题失败: %v", err)
	}

	// 验证标题数量
	if len(headings) == 0 {
		t.Fatal("未提取到任何标题")
	}

	t.Logf("成功提取 %d 个标题", len(headings))

	// 期望的关键标题（基于大纲）
	// 这里列出一些期望存在的标题作为参考
	_ = []struct {
		title string
		level int
		page  int64
	}{
		{"1.Project background and objectives", 1, 2},
		{"1.1.DCC：Document Control Center", 2, 2},
		{"1.2.DCN：Document Change Notice(for Company Rules and R", 2, 2},
		{"1.3.DMS：Document Management System", 2, 2},
		{"1.4.Background description", 2, 2},
		{"1.4.1.Industry prospects1", 3, 2},
		{"1.4.2.Industry prospects2", 3, 2},
		{"1.4.3.Industry prospects3", 3, 2},
		{"1.4.4.Industry prospects4", 3, 2},
		{"1.4.5.Industry prospects", 3, 2},
		{"1.4.6.Main pain points", 3, 2},
		{"1.5.Construction objectives", 2, 2},
		{"2.Overall scheme design", 1, 3},
		{"2.1.Overview of system architecture", 2, 3},
		{"2.2.技术选型说明", 2, 3},
		{"2.2.1.后端与基础设施", 3, 3},
		{"2.2.2.模型相关", 3, 3},
		{"2.2.3.其他模型相关", 3, 3},
		{"2.3.Technical selection description", 2, 4},
		{"3.Core functional modules", 1, 5},
		{"4.4. Implementation plan and milestones", 1, 6},
		{"4.1.Categories and hierarchy of documents", 2, 6},
		{"4.2.实施阶段划分", 2, 6},
		{"4.3.风险与应对", 2, 6},
		{"4.3.1.技术风险", 3, 6},
		{"4.3.2.业务风险", 3, 6},
		{"5.Attachment", 1, 7},
		{"6.总结与展望", 1, 7},
		{"7.SMIC document security levels are categorized accounting to the security control level’s defined in “CR-GWNL-01-1001 SMIC Classified Information Protection Policy (CIPP)”as follows", 1, 8},
		{"8.TECN: Temporary engineering changes can be requested as necessary", 1, 8},
		{"8.1.Effective period of TECN", 2, 9},
		{"8.1.1.The maximum effective period of a TECN is 14 working days. If in special siituation, e.g. raw material or part specification evaluation, the maxium effective period is 6 months.", 3, 9},
		{"8.1.2.TECN renewal:TECN which are preferable,suitable,and necessary for advancing into permanent engineering changes can be renewed for three time at most.", 3, 9},
	}

	// 验证关键标题是否存在
	foundTitles := make(map[string]bool)
	for _, h := range headings {
		foundTitles[h.Title] = true
	}

	// 检查一些关键标题
	keyTitles := []string{
		"1.Project background and objectives",
		"2.Overall scheme design",
		"3.Core functional modules",
		"5.Attachment",
		"6.总结与展望",
	}

	for _, title := range keyTitles {
		found := false
		for _, h := range headings {
			if h.Title == title {
				found = true
				t.Logf("✓ 找到关键标题: %s (级别 %d, 第 %d 页)", title, h.Level, h.Page)
				break
			}
		}
		if !found {
			t.Errorf("✗ 未找到关键标题: %s", title)
		}
	}

	// 验证标题层级结构
	levelCounts := make(map[int]int)
	for _, h := range headings {
		levelCounts[h.Level]++
	}

	t.Logf("标题层级分布:")
	for level := 1; level <= 3; level++ {
		if count, ok := levelCounts[level]; ok {
			t.Logf("  级别 %d: %d 个标题", level, count)
		}
	}

	// 验证至少有一级标题
	if levelCounts[1] == 0 {
		t.Error("未找到任何一级标题")
	}

	// 验证页码范围
	for _, h := range headings {
		if h.Page < 1 || h.Page > 10 {
			t.Errorf("标题 '%s' 的页码 %d 超出合理范围", h.Title, h.Page)
		}
	}
}

func TestExtractHeadingsFromOutline(t *testing.T) {
	pdfPath := "135_title.pdf"

	headings, err := extractHeadings(pdfPath)
	if err != nil {
		t.Fatalf("提取标题失败: %v", err)
	}

	// 验证是否至少提取到了大纲中的标题
	if len(headings) < 50 {
		t.Errorf("提取的标题数量 (%d) 少于预期 (至少 50 个)", len(headings))
	}

	// 输出所有标题用于调试
	t.Log("提取到的所有标题:")
	for i, h := range headings {
		var indent strings.Builder
		indent.Grow((h.Level - 1) * 2)
		for j := 1; j < h.Level; j++ {
			indent.WriteString("  ")
		}
		t.Logf("%d. %s级别 %d: %s (第 %d 页)", i+1, indent.String(), h.Level, h.Title, h.Page)
	}
}

func TestHeadingStructure(t *testing.T) {
	pdfPath := "135_title.pdf"

	headings, err := extractHeadings(pdfPath)
	if err != nil {
		t.Fatalf("提取标题失败: %v", err)
	}

	// 验证标题结构的合理性
	prevLevel := 0
	for i, h := range headings {
		// 检查级别跳跃是否合理（不应该从1级直接跳到3级）
		if i > 0 && h.Level > prevLevel+1 {
			t.Logf("警告: 标题 '%s' 的级别 (%d) 从上一个标题的级别 (%d) 跳跃过大",
				h.Title, h.Level, prevLevel)
		}
		prevLevel = h.Level

		// 检查标题是否为空
		if h.Title == "" {
			t.Errorf("第 %d 个标题为空", i+1)
		}

		// 检查级别是否合理
		if h.Level < 1 || h.Level > 5 {
			t.Errorf("标题 '%s' 的级别 %d 不合理", h.Title, h.Level)
		}
	}
}

func TestMergeHeadings(t *testing.T) {
	// 测试标题合并功能
	outline := []Heading{
		{Title: "1. Introduction", Level: 1, Page: 1},
		{Title: "1.1 Background", Level: 2, Page: 1},
	}

	text := []Heading{
		{Title: "1. Introduction", Level: 1, Page: 1}, // 重复
		{Title: "2. Methods", Level: 1, Page: 2},      // 新标题
	}

	merged := mergeHeadings(outline, text)

	// 验证合并后的数量
	expectedCount := 3 // 1个重复被去除，应该有3个标题
	if len(merged) != expectedCount {
		t.Errorf("合并后的标题数量 %d，期望 %d", len(merged), expectedCount)
	}

	// 验证是否包含所有唯一标题
	titles := make(map[string]bool)
	for _, h := range merged {
		titles[h.Title] = true
	}

	if !titles["1. Introduction"] {
		t.Error("缺少标题: 1. Introduction")
	}
	if !titles["1.1 Background"] {
		t.Error("缺少标题: 1.1 Background")
	}
	if !titles["2. Methods"] {
		t.Error("缺少标题: 2. Methods")
	}
}

// 基准测试
func BenchmarkExtractHeadings(b *testing.B) {
	pdfPath := "135_title.pdf"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := extractHeadings(pdfPath)
		if err != nil {
			b.Fatalf("提取标题失败: %v", err)
		}
	}
}

// TestExtractTextBlocks 测试 extractTextBlocks 函数
func TestExtractTextBlocks(t *testing.T) {
	// 打开测试 PDF 文件
	pdfPath := "135_title.pdf"
	f, err := os.Open(pdfPath)
	if err != nil {
		t.Fatalf("无法打开文件: %v", err)
	}
	defer f.Close()

	// 创建 PDF 阅读器
	pdfReader, err := model.NewPdfReader(f)
	if err != nil {
		t.Fatalf("无法读取 PDF: %v", err)
	}

	// 提取文本块
	blocks, err := extractTextBlocks(pdfReader)
	if err != nil {
		t.Fatalf("提取文本块失败: %v", err)
	}

	// 验证提取到了文本块
	if len(blocks) == 0 {
		t.Fatal("未提取到任何文本块")
	}

	t.Logf("成功提取 %d 个文本块", len(blocks))

	// 验证文本块的基本属性
	pageCount := make(map[int64]int)
	for _, block := range blocks {
		// 检查页码是否合理
		if block.Page < 1 {
			t.Errorf("文本块页码 %d 不合理", block.Page)
		}

		// 检查内容是否为空
		if block.Content == "" {
			t.Error("发现空文本块")
		}

		// 检查高度是否合理
		if block.Height <= 0 {
			t.Errorf("文本块高度 %f 不合理", block.Height)
		}

		// 检查 Y 坐标是否在合理范围内
		if block.YPos < 0 || block.YPos > block.Height {
			t.Errorf("文本块 Y 坐标 %f 超出页面高度 %f", block.YPos, block.Height)
		}

		pageCount[block.Page]++
	}

	// 验证至少有多个页面的文本块
	if len(pageCount) < 2 {
		t.Errorf("只提取到 %d 个页面的文本块，期望至少 2 个", len(pageCount))
	}

	t.Logf("文本块分布在 %d 个页面", len(pageCount))

	// 输出前几个文本块用于调试
	t.Log("前 5 个文本块:")
	for i := 0; i < min(5, len(blocks)); i++ {
		block := blocks[i]
		content := block.Content
		if len(content) > 50 {
			content = content[:50] + "..."
		}
		t.Logf("  块 %d: 页 %d, Y=%.2f, 高度=%.2f, 内容='%s'",
			i+1, block.Page, block.YPos, block.Height, content)
	}
}

// TestExtractTextBlocksPositionInfo 测试文本块的位置信息
func TestExtractTextBlocksPositionInfo(t *testing.T) {
	pdfPath := "135_title.pdf"
	f, err := os.Open(pdfPath)
	if err != nil {
		t.Fatalf("无法打开文件: %v", err)
	}
	defer f.Close()

	pdfReader, err := model.NewPdfReader(f)
	if err != nil {
		t.Fatalf("无法读取 PDF: %v", err)
	}

	blocks, err := extractTextBlocks(pdfReader)
	if err != nil {
		t.Fatalf("提取文本块失败: %v", err)
	}

	// 按页面分组
	pageBlocks := make(map[int64][]TextBlock)
	for _, block := range blocks {
		pageBlocks[block.Page] = append(pageBlocks[block.Page], block)
	}

	// 验证每个页面的文本块位置分布
	for page, blocks := range pageBlocks {
		if len(blocks) == 0 {
			continue
		}

		// 计算相对位置分布
		var topCount, middleCount, bottomCount int
		for _, block := range blocks {
			relativePos := block.YPos / block.Height
			if relativePos > 0.85 {
				topCount++
			} else if relativePos < 0.15 {
				bottomCount++
			} else {
				middleCount++
			}
		}

		t.Logf("第 %d 页: 顶部 %d 块, 中间 %d 块, 底部 %d 块",
			page, topCount, middleCount, bottomCount)

		// 验证至少有一些文本块在中间区域（正文区域）
		if middleCount == 0 && len(blocks) > 3 {
			t.Logf("警告: 第 %d 页没有中间区域的文本块", page)
		}
	}
}
