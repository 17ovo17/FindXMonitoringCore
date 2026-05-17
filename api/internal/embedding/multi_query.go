package embedding

import (
	"context"
	"fmt"
	"sort"
	"strings"

	log "github.com/sirupsen/logrus"
)

// MultiQueryConfig 多查询检索配置
type MultiQueryConfig struct {
	Enabled    bool // 是否启用多查询检索
	NumQueries int  // 生成几个变体，默认 3
}

// MultiQueryResult 多查询检索结果
type MultiQueryResult struct {
	OriginalQuery string         `json:"original_query"`
	Variants      []string       `json:"variants"`
	SearchResults []SearchResult `json:"search_results"`
}

// LoadMultiQueryConfig 从配置加载多查询检索设置
func LoadMultiQueryConfig() MultiQueryConfig {
	return MultiQueryConfig{
		Enabled:    getSettingOrViperBool("multi_query.enabled", "multi_query.enabled", false),
		NumQueries: getSettingOrViperInt("multi_query.num_queries", "multi_query.num_queries", 3),
	}
}

// SearchMultiQuery 多查询检索
// 核心思路：一个用户问题可能有多个角度，
// 生成多个变体查询，分别搜索，合并去重结果。
func (h *HybridSearcher) SearchMultiQuery(ctx context.Context, query string, topK int) (*MultiQueryResult, error) {
	cfg := LoadMultiQueryConfig()
	result := &MultiQueryResult{
		OriginalQuery: query,
	}

	// 如果未启用或无 LLM 客户端，退回普通搜索
	if !cfg.Enabled || llmClientInstance == nil {
		results, err := h.Search(ctx, query, topK)
		if err != nil {
			return nil, err
		}
		result.SearchResults = results
		return result, nil
	}

	// 1. 生成查询变体
	variants, err := generateQueryVariants(ctx, query, cfg.NumQueries)
	if err != nil {
		log.WithError(err).Warn("embedding: multi-query variant generation failed, falling back")
		results, err := h.Search(ctx, query, topK)
		if err != nil {
			return nil, err
		}
		result.SearchResults = results
		return result, nil
	}
	result.Variants = variants

	// 2. 对每个变体执行搜索（包括原始 query）
	allQueries := append([]string{query}, variants...)
	docScores := make(map[string]float64)
	docMap := make(map[string]SearchResult)

	for _, q := range allQueries {
		results, searchErr := h.Search(ctx, q, defaultSearchTop)
		if searchErr != nil {
			continue
		}
		for _, r := range results {
			// 取最高分
			if r.Score > docScores[r.DocID] {
				docScores[r.DocID] = r.Score
				docMap[r.DocID] = r
			}
		}
	}

	// 3. 按分数排序
	type scoredItem struct {
		docID string
		score float64
	}
	ranked := make([]scoredItem, 0, len(docScores))
	for id, s := range docScores {
		ranked = append(ranked, scoredItem{docID: id, score: s})
	}
	sort.Slice(ranked, func(i, j int) bool {
		return ranked[i].score > ranked[j].score
	})

	// 4. 取 topK
	results := make([]SearchResult, 0, topK)
	for i, item := range ranked {
		if i >= topK {
			break
		}
		doc := docMap[item.docID]
		doc.Score = item.score
		results = append(results, doc)
	}
	result.SearchResults = results
	return result, nil
}

// generateQueryVariants 生成查询变体
// 策略：
// - 同义词替换
// - 角度变换（"怎么解决" → "排查方法"/"最佳实践"/"配置优化"）
// - 具体化（加入可能的技术术语）
func generateQueryVariants(ctx context.Context, query string, numVariants int) ([]string, error) {
	if llmClientInstance == nil {
		return nil, fmt.Errorf("LLM client not available")
	}

	prompt := fmt.Sprintf(
		"请为以下运维搜索查询生成 %d 个不同角度的变体查询，用于扩展搜索范围。"+
			"每个变体一行，不要编号，不要解释。\n\n"+
			"原始查询：%s\n\n变体：", numVariants, query)

	response, err := llmClientInstance.Generate(ctx, prompt, 200, 0.8)
	if err != nil {
		return nil, err
	}

	// 解析响应：每行一个变体
	lines := strings.Split(response, "\n")
	var variants []string
	for _, line := range lines {
		line = strings.TrimSpace(line)
		// 去除可能的编号前缀
		line = trimNumberPrefix(line)
		if line != "" && line != query {
			variants = append(variants, line)
		}
		if len(variants) >= numVariants {
			break
		}
	}
	return variants, nil
}

// trimNumberPrefix 去除 "1. " "1、" "- " 等前缀
func trimNumberPrefix(s string) string {
	s = strings.TrimSpace(s)
	if len(s) == 0 {
		return s
	}
	// 去除 "1. " "2. " 等
	for i, r := range s {
		if r >= '0' && r <= '9' {
			continue
		}
		if r == '.' || r == '、' || r == ')' || r == '）' {
			return strings.TrimSpace(s[i+1:])
		}
		break
	}
	// 去除 "- "
	if strings.HasPrefix(s, "- ") {
		return strings.TrimSpace(s[2:])
	}
	return s
}
