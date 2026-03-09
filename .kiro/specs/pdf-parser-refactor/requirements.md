# 需求文档

## 简介

PDF 标题解析工具是一个用 Golang 编写的命令行工具，用于从 PDF 文档中提取结构化内容，包括标题、正文、页眉和页脚。该工具支持从 PDF 大纲（目录/书签）和文本内容中识别标题，并支持多种标题格式（如 '1. ', '1.1 ', '2.3.4 ' 等）。

本需求文档旨在整理现有实现，添加正文和页眉页脚提取功能，解决代码质量问题，并提升工具的性能和可维护性。

## 术语表

- **Parser**: PDF 内容解析器，负责从 PDF 文件中提取结构化内容
- **Heading**: 标题结构，包含标题文字、级别和页码
- **Body_Text**: 正文内容，PDF 文档的主要文本内容（不包括标题、页眉、页脚）
- **Header**: 页眉，出现在每页顶部的重复性文本
- **Footer**: 页脚，出现在每页底部的重复性文本（通常包含页码）
- **Outline**: PDF 大纲（目录/书签），PDF 文件内置的导航结构
- **Text_Extractor**: 文本提取器，从 PDF 页面内容中提取文本
- **Content_Classifier**: 内容分类器，识别和分类文本内容类型（标题、正文、页眉、页脚）
- **Heading_Merger**: 标题合并器，负责合并和去重不同来源的标题
- **CLI**: 命令行接口，用户与工具交互的界面
- **Formatter**: 格式化器，负责输出内容的格式化显示

## 需求

### 需求 1: 从 PDF 大纲提取标题

**用户故事:** 作为用户，我想从 PDF 大纲中提取标题，以便快速获取文档的结构化目录信息。

#### 验收标准

1. WHEN 提供包含大纲的 PDF 文件时，THE Parser SHALL 提取所有大纲项作为标题
2. THE Parser SHALL 为每个大纲项记录标题文字、层级和页码
3. THE Parser SHALL 递归提取嵌套的大纲子项
4. THE Parser SHALL 根据大纲层级结构正确设置标题级别（从 1 开始）
5. IF PDF 文件不包含大纲，THEN THE Parser SHALL 继续尝试从文本内容提取标题

### 需求 2: 从文本内容识别标题

**用户故事:** 作为用户，我想从 PDF 文本内容中识别标题格式，以便在没有大纲的情况下也能提取标题。

#### 验收标准

1. THE Text_Extractor SHALL 支持识别 "数字. " 格式的标题（如 "1. ", "7. "）
2. THE Text_Extractor SHALL 支持识别 "数字.数字 " 格式的标题（如 "1.1 ", "2.3.4 "）
3. THE Text_Extractor SHALL 支持识别 "数字.数字. " 格式的标题（带结尾点号，如 "1.1.", "2.3.4."）
4. THE Text_Extractor SHALL 根据点号数量计算标题级别（点号数量 + 1）
5. WHEN 在文本行中匹配到标题格式时，THE Text_Extractor SHALL 记录完整行内容、级别和页码
6. THE Text_Extractor SHALL 逐页扫描 PDF 文档的所有页面
7. THE Text_Extractor SHALL 忽略空行和纯空白行

### 需求 3: 合并和去重标题

**用户故事:** 作为用户，我想自动合并来自大纲和文本的标题并去除重复项，以便获得完整且不重复的标题列表。

#### 验收标准

1. THE Heading_Merger SHALL 优先保留大纲中的标题
2. THE Heading_Merger SHALL 添加文本中提取的标题到结果列表
3. WHEN 检测到重复标题时，THE Heading_Merger SHALL 仅保留一个副本
4. THE Heading_Merger SHALL 通过页码和标题前缀相似度判断重复（前 10 个字符）
5. IF 两个标题在同一页且标题开头相似，THEN THE Heading_Merger SHALL 认为它们是重复的

### 需求 4: 命令行接口

**用户故事:** 作为用户，我想通过命令行参数控制工具行为，以便灵活使用工具。

#### 验收标准

1. THE CLI SHALL 接受 PDF 文件路径作为必需参数
2. WHERE 用户提供 "-d" 或 "--debug" 参数，THE CLI SHALL 启用调试模式显示详细信息
3. WHERE 用户提供 "-h" 或 "--help" 参数，THE CLI SHALL 显示使用说明并退出
4. WHERE 用户提供 "--extract-all" 参数，THE CLI SHALL 提取所有内容类型
5. WHERE 用户提供 "--extract-body" 参数，THE CLI SHALL 仅提取正文
6. WHERE 用户提供 "--extract-header" 参数，THE CLI SHALL 仅提取页眉
7. WHERE 用户提供 "--extract-footer" 参数，THE CLI SHALL 仅提取页脚
8. WHERE 用户提供 "--format json" 参数，THE CLI SHALL 以 JSON 格式输出
9. IF 用户未提供任何参数，THEN THE CLI SHALL 显示使用说明并以错误码退出
10. IF 用户提供的文件路径无效，THEN THE CLI SHALL 返回清晰的错误消息

### 需求 5: 格式化输出标题

**用户故事:** 作为用户，我想以清晰的层级结构查看提取的标题，以便理解文档结构。

#### 验收标准

1. THE Formatter SHALL 显示提取到的标题总数
2. THE Formatter SHALL 为每个标题显示级别、标题文字和页码
3. THE Formatter SHALL 根据标题级别添加缩进（每级 2 个空格）
4. THE Formatter SHALL 按提取顺序输出标题
5. IF 未提取到任何标题，THEN THE Formatter SHALL 显示友好提示消息

### 需求 6: 性能优化

**用户故事:** 作为开发者，我想优化代码性能，以便工具能高效处理大型 PDF 文件。

#### 验收标准

1. THE Parser SHALL 使用 strings.Builder 替代字符串拼接来构建缩进
2. THE Parser SHALL 使用 Go 1.24+ 内置的 min 函数替代自定义实现
3. WHERE Go 版本支持 strings.SplitSeq，THE Text_Extractor SHALL 使用 SplitSeq 提高行分割效率
4. THE Parser SHALL 在文本提取失败时仍返回大纲中的标题（容错处理）
5. THE Parser SHALL 在单页提取失败时继续处理其他页面（容错处理）

### 需求 7: 错误处理

**用户故事:** 作为用户，我想在遇到错误时获得清晰的错误信息，以便快速定位和解决问题。

#### 验收标准

1. IF 文件无法打开，THEN THE Parser SHALL 返回包含文件路径的错误消息
2. IF PDF 文件格式无效，THEN THE Parser SHALL 返回描述性错误消息
3. THE Parser SHALL 使用 fmt.Errorf 和 %w 包装底层错误以保留错误链
4. THE CLI SHALL 使用 log.Fatalf 输出致命错误并以非零状态码退出
5. THE Parser SHALL 在部分提取失败时继续执行并返回已提取的结果

### 需求 8: 代码质量和可维护性

**用户故事:** 作为开发者，我想保持代码的高质量和可维护性，以便团队协作和长期维护。

#### 验收标准

1. THE Parser SHALL 将核心功能分解为独立的函数（单一职责原则）
2. THE Parser SHALL 为所有导出类型和函数提供清晰的注释
3. THE Parser SHALL 遵循 Go 语言惯用法和最佳实践
4. THE Parser SHALL 通过所有静态分析检查（无 lint 警告）
5. THE Parser SHALL 保持现有的单元测试覆盖率
6. THE Parser SHALL 确保所有测试用例通过

### 需求 9: 测试覆盖

**用户故事:** 作为开发者，我想保持完整的测试覆盖，以便确保代码重构不会引入回归问题。

#### 验收标准

1. THE Parser SHALL 保留所有现有的单元测试
2. THE Parser SHALL 确保测试覆盖标题提取、合并、格式化等核心功能
3. THE Parser SHALL 包含针对真实 PDF 文件的集成测试
4. THE Parser SHALL 包含性能基准测试
5. WHERE 修改核心逻辑，THE Parser SHALL 更新相应的测试用例以保持一致性

### 需求 10: 向后兼容性

**用户故事:** 作为用户，我想确保重构后的工具保持相同的行为，以便无缝升级。

#### 验收标准

1. THE Parser SHALL 保持相同的命令行接口和参数
2. THE Parser SHALL 保持相同的输出格式
3. THE Parser SHALL 保持相同的 Heading 数据结构
4. THE Parser SHALL 保持相同的标题提取逻辑和结果
5. THE Parser SHALL 保持相同的错误处理行为

### 需求 11: 提取正文内容

**用户故事:** 作为用户，我想从 PDF 文档中提取正文内容，以便获取文档的主要文本信息。

#### 验收标准

1. THE Parser SHALL 从每个 PDF 页面提取正文文本内容
2. THE Parser SHALL 排除标题、页眉和页脚，仅保留正文内容
3. THE Parser SHALL 保持正文的段落结构和换行
4. THE Parser SHALL 为每段正文记录所在页码
5. THE Parser SHALL 按页面顺序组织正文内容
6. WHERE 正文跨越多页，THE Parser SHALL 正确连接内容
7. THE Parser SHALL 处理多栏布局的正文内容

#### 正确性属性

1. **完整性**: 所有非标题、非页眉、非页脚的文本都应被识别为正文
2. **准确性**: 正文内容不应包含标题、页眉或页脚文本
3. **顺序性**: 正文内容应按照文档的阅读顺序排列

### 需求 12: 识别和提取页眉

**用户故事:** 作为用户，我想识别和提取 PDF 文档的页眉，以便了解文档的元信息和导航信息。

#### 验收标准

1. THE Parser SHALL 识别出现在每页顶部的重复性文本作为页眉
2. THE Parser SHALL 通过位置信息判断页眉（页面顶部区域）
3. THE Parser SHALL 通过跨页重复性判断页眉（在多页中重复出现）
4. THE Parser SHALL 记录页眉的文本内容和出现的页码范围
5. WHERE 页眉在不同页面有变化，THE Parser SHALL 识别为不同的页眉
6. THE Parser SHALL 支持左右页不同的页眉（奇偶页页眉）
7. IF 文档没有页眉，THEN THE Parser SHALL 返回空的页眉列表

#### 正确性属性

1. **位置准确性**: 页眉应位于页面顶部区域（通常是页面高度的前 10-15%）
2. **重复性**: 页眉内容应在多个页面中重复出现
3. **一致性**: 相同的页眉文本应被识别为同一页眉

### 需求 13: 识别和提取页脚

**用户故事:** 作为用户，我想识别和提取 PDF 文档的页脚，以便获取页码和其他页面级别的信息。

#### 验收标准

1. THE Parser SHALL 识别出现在每页底部的重复性文本作为页脚
2. THE Parser SHALL 通过位置信息判断页脚（页面底部区域）
3. THE Parser SHALL 通过跨页重复性判断页脚（在多页中重复出现）
4. THE Parser SHALL 识别页脚中的页码信息
5. THE Parser SHALL 记录页脚的文本内容和出现的页码范围
6. WHERE 页脚在不同页面有变化（如页码递增），THE Parser SHALL 识别页脚模式
7. THE Parser SHALL 支持左右页不同的页脚（奇偶页页脚）
8. IF 文档没有页脚，THEN THE Parser SHALL 返回空的页脚列表

#### 正确性属性

1. **位置准确性**: 页脚应位于页面底部区域（通常是页面高度的后 10-15%）
2. **重复性**: 页脚内容或模式应在多个页面中重复出现
3. **页码识别**: 页脚中的数字序列应被正确识别为页码

### 需求 14: 内容分类和输出格式

**用户故事:** 作为用户，我想以结构化的方式查看提取的所有内容，以便清晰地理解文档结构。

#### 验收标准

1. THE CLI SHALL 支持 "--extract-all" 参数提取所有内容（标题、正文、页眉、页脚）
2. THE CLI SHALL 支持 "--extract-body" 参数仅提取正文
3. THE CLI SHALL 支持 "--extract-header" 参数仅提取页眉
4. THE CLI SHALL 支持 "--extract-footer" 参数仅提取页脚
5. THE Formatter SHALL 以清晰的分区显示不同类型的内容
6. THE Formatter SHALL 为每种内容类型显示统计信息（数量、页码范围等）
7. THE Formatter SHALL 支持输出为 JSON 格式（通过 "--format json" 参数）
8. THE Formatter SHALL 支持输出为纯文本格式（默认）

#### 输出格式示例

```
=== 标题 (5 个) ===
级别 1: 第一章 引言 (第 1 页)
  级别 2: 1.1 背景 (第 2 页)
  
=== 页眉 ===
"文档标题 | 第一章" (第 1-5 页)

=== 页脚 ===
"第 {n} 页" (第 1-10 页)

=== 正文 (第 1 页) ===
这是第一页的正文内容...

=== 正文 (第 2 页) ===
这是第二页的正文内容...
```

### 需求 15: 内容分类算法

**用户故事:** 作为开发者，我想实现准确的内容分类算法，以便正确区分标题、正文、页眉和页脚。

#### 验收标准

1. THE Content_Classifier SHALL 使用位置信息（Y 坐标）作为主要分类依据
2. THE Content_Classifier SHALL 使用文本重复性作为页眉页脚的辅助判断
3. THE Content_Classifier SHALL 使用字体大小和样式作为标题的辅助判断（如果可用）
4. THE Content_Classifier SHALL 定义可配置的页眉区域阈值（默认页面顶部 15%）
5. THE Content_Classifier SHALL 定义可配置的页脚区域阈值（默认页面底部 15%）
6. THE Content_Classifier SHALL 在分类不确定时优先归类为正文
7. THE Content_Classifier SHALL 处理边界情况（如标题位于页眉区域）

#### 正确性属性

1. **互斥性**: 同一文本块不应同时被分类为多种类型
2. **完整性**: 所有提取的文本都应被分类到某一类型
3. **准确性**: 分类错误率应低于 5%（基于测试数据集）
