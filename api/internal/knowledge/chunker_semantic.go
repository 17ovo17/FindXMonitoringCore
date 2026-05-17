package knowledge

import (
	"strings"
	"unicode"
	"unicode/utf8"
)

// SemanticChunkConfig 语义分块配置
type SemanticChunkConfig struct {
	MaxChunkSize        int     // 最大块大小（字符），默认 1200
	MinChunkSize        int     // 最小块大小，默认 100
	SimilarityThreshold float64 // 相似度阈值，低于此值则切分，默认 0.3
	WindowSize          int     // 滑动窗口大小（句子数），默认 3
}

func normalizeSemanticConfig(cfg SemanticChunkConfig) SemanticChunkConfig {
	if cfg.MaxChunkSize <= 0 {
		cfg.MaxChunkSize = 1200
	}
	if cfg.MinChunkSize <= 0 {
		cfg.MinChunkSize = 100
	}
	if cfg.SimilarityThreshold <= 0 {
		cfg.SimilarityThreshold = 0.3
	}
	if cfg.WindowSize <= 0 {
		cfg.WindowSize = 3
	}
	return cfg
}

// ChunkSemantic 语义分块：按语义主题切分文档，而非固定长度。
// 算法：
// 1. 先按句子分割
// 2. 计算相邻窗口的词重叠度（Jaccard 相似度）
// 3. 当相似度低于阈值时，认为是主题切换点
// 4. 在切换点处分块
//
// 降级方案：如果句子数不足以做语义分析，退回到段落分块。
func ChunkSemantic(content string, cfg SemanticChunkConfig) []Chunk {
	cfg = normalizeSemanticConfig(cfg)

	// 短文档直接返回
	if utf8.RuneCountInString(content) <= cfg.MaxChunkSize {
		return []Chunk{{Index: 0, Content: content, Start: 0, End: len(content)}}
	}

	sentences := splitSentences(content)
	// 句子数不足，降级到段落分块
	if len(sentences) < cfg.WindowSize*2+1 {
		return ChunkDocument(content, ChunkConfig{MaxChunkSize: cfg.MaxChunkSize, Overlap: 100})
	}

	breakpoints := findSemanticBreakpoints(sentences, cfg)
	chunks := splitAtBreakpoints(sentences, breakpoints, cfg)
	return reindexChunks(chunks)
}

// splitSentences 按中英文句子边界分割文本
func splitSentences(content string) []string {
	var sentences []string
	var buf strings.Builder
	runes := []rune(content)

	for i := 0; i < len(runes); i++ {
		buf.WriteRune(runes[i])
		if isSentenceEnd(runes[i]) {
			s := strings.TrimSpace(buf.String())
			if s != "" {
				sentences = append(sentences, s)
			}
			buf.Reset()
		}
	}
	// 剩余内容
	if s := strings.TrimSpace(buf.String()); s != "" {
		sentences = append(sentences, s)
	}
	return sentences
}

// isSentenceEnd 判断是否为句子结束符
func isSentenceEnd(r rune) bool {
	return r == '。' || r == '！' || r == '？' || r == '.' || r == '!' || r == '?' || r == '\n'
}

// findSemanticBreakpoints 找到语义断点。
// 使用滑动窗口 + 词重叠度（Jaccard）：
// - 窗口 A = 前 N 个句子的词集合
// - 窗口 B = 后 N 个句子的词集合
// - 如果 Jaccard(A, B) < threshold，则此处是断点
func findSemanticBreakpoints(sentences []string, cfg SemanticChunkConfig) []int {
	var breakpoints []int
	w := cfg.WindowSize

	for i := w; i < len(sentences)-w; i++ {
		// 构建前窗口词集合
		windowA := buildWordSet(sentences[i-w : i])
		// 构建后窗口词集合
		windowB := buildWordSet(sentences[i : i+w])

		sim := jaccardSimilarity(windowA, windowB)
		if sim < cfg.SimilarityThreshold {
			breakpoints = append(breakpoints, i)
		}
	}
	return breakpoints
}

// splitAtBreakpoints 在断点处切分句子为块
func splitAtBreakpoints(sentences []string, breakpoints []int, cfg SemanticChunkConfig) []Chunk {
	var chunks []Chunk
	start := 0
	byteOffset := 0

	allBreaks := append(breakpoints, len(sentences))
	for _, bp := range allBreaks {
		if bp <= start {
			continue
		}
		text := joinSentences(sentences[start:bp])
		runeLen := utf8.RuneCountInString(text)

		// 如果块太大，进一步切分
		if runeLen > cfg.MaxChunkSize {
			subChunks := ChunkDocument(text, ChunkConfig{MaxChunkSize: cfg.MaxChunkSize, Overlap: 50})
			for _, sc := range subChunks {
				sc.Start += byteOffset
				sc.End += byteOffset
				chunks = append(chunks, sc)
			}
		} else if runeLen >= cfg.MinChunkSize {
			chunks = append(chunks, Chunk{
				Content: text,
				Start:   byteOffset,
				End:     byteOffset + len(text),
			})
		} else if len(chunks) > 0 {
			// 太小的块合并到前一个
			prev := &chunks[len(chunks)-1]
			prev.Content += "\n" + text
			prev.End = byteOffset + len(text)
		} else {
			chunks = append(chunks, Chunk{
				Content: text,
				Start:   byteOffset,
				End:     byteOffset + len(text),
			})
		}

		byteOffset += len(text) + 1 // +1 for separator
		start = bp
	}
	return chunks
}

// joinSentences 将句子列表合并为文本
func joinSentences(sentences []string) string {
	return strings.Join(sentences, "")
}

// buildWordSet 从句子列表中提取词集合
func buildWordSet(sentences []string) map[string]struct{} {
	set := make(map[string]struct{})
	for _, s := range sentences {
		words := tokenizeForJaccard(s)
		for _, w := range words {
			set[w] = struct{}{}
		}
	}
	return set
}

// tokenizeForJaccard 简单分词：按空格和标点分割，中文按字/双字切分
func tokenizeForJaccard(text string) []string {
	var tokens []string
	var word strings.Builder
	runes := []rune(strings.ToLower(text))

	for i := 0; i < len(runes); i++ {
		r := runes[i]
		if unicode.IsLetter(r) || unicode.IsDigit(r) {
			if unicode.Is(unicode.Han, r) {
				// 中文：单字 + 双字 gram
				if word.Len() > 0 {
					tokens = append(tokens, word.String())
					word.Reset()
				}
				tokens = append(tokens, string(r))
				if i+1 < len(runes) && unicode.Is(unicode.Han, runes[i+1]) {
					tokens = append(tokens, string(runes[i:i+2]))
				}
			} else {
				word.WriteRune(r)
			}
		} else {
			if word.Len() > 0 {
				tokens = append(tokens, word.String())
				word.Reset()
			}
		}
	}
	if word.Len() > 0 {
		tokens = append(tokens, word.String())
	}
	return tokens
}

// jaccardSimilarity 计算两个词集合的 Jaccard 相似度
func jaccardSimilarity(a, b map[string]struct{}) float64 {
	if len(a) == 0 && len(b) == 0 {
		return 1.0
	}
	intersection := 0
	for k := range a {
		if _, ok := b[k]; ok {
			intersection++
		}
	}
	union := len(a) + len(b) - intersection
	if union == 0 {
		return 1.0
	}
	return float64(intersection) / float64(union)
}
