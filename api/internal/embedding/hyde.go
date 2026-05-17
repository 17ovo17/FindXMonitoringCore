package embedding

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	log "github.com/sirupsen/logrus"
)

// HyDEConfig HyDE（假设文档嵌入）配置
type HyDEConfig struct {
	Enabled     bool    // 是否启用 HyDE
	MaxTokens   int     // 假设文档最大 token 数，默认 200
	Temperature float64 // 生成温度，默认 0.7
}

// HyDEResult HyDE 搜索结果
type HyDEResult struct {
	OriginalQuery   string         `json:"original_query"`
	HypotheticalDoc string         `json:"hypothetical_doc"`
	SearchResults   []SearchResult `json:"search_results"`
	HyDEUsed        bool           `json:"hyde_used"`
}

// LLMClient LLM 调用接口，由外部注入
type LLMClient interface {
	// Generate 生成文本（非流式）
	Generate(ctx context.Context, prompt string, maxTokens int, temperature float64) (string, error)
}

// defaultLLMClient 基于 OpenAI 兼容 API 的默认 LLM 客户端
type defaultLLMClient struct {
	baseURL string
	apiKey  string
	model   string
}

// llmClientInstance 全局 LLM 客户端（由 handler 层注入）
var llmClientInstance LLMClient

// SetLLMClient 设置全局 LLM 客户端
func SetLLMClient(c LLMClient) {
	llmClientInstance = c
}

// GetLLMClient 获取全局 LLM 客户端
func GetLLMClient() LLMClient {
	return llmClientInstance
}

// NewDefaultLLMClient 创建基于 OpenAI 兼容 API 的 LLM 客户端
func NewDefaultLLMClient(baseURL, apiKey, model string) LLMClient {
	return &defaultLLMClient{
		baseURL: strings.TrimRight(baseURL, "/"),
		apiKey:  apiKey,
		model:   model,
	}
}

func (c *defaultLLMClient) Generate(ctx context.Context, prompt string, maxTokens int, temperature float64) (string, error) {
	if c.baseURL == "" || c.apiKey == "" {
		return "", fmt.Errorf("LLM client not configured")
	}

	payload := map[string]interface{}{
		"model": c.model,
		"messages": []map[string]string{
			{"role": "user", "content": prompt},
		},
		"max_tokens":  maxTokens,
		"temperature": temperature,
		"stream":      false,
	}
	body, err := json.Marshal(payload)
	if err != nil {
		return "", fmt.Errorf("marshal LLM request: %w", err)
	}

	url := c.baseURL + "/chat/completions"
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(body))
	if err != nil {
		return "", fmt.Errorf("create LLM request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+c.apiKey)

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("LLM request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		respBody, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("LLM returned status %d: %s", resp.StatusCode, string(respBody))
	}

	var result struct {
		Choices []struct {
			Message struct {
				Content string `json:"content"`
			} `json:"message"`
		} `json:"choices"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", fmt.Errorf("decode LLM response: %w", err)
	}
	if len(result.Choices) == 0 {
		return "", fmt.Errorf("LLM returned no choices")
	}
	return strings.TrimSpace(result.Choices[0].Message.Content), nil
}

// LoadHyDEConfig 从配置加载 HyDE 设置
func LoadHyDEConfig() HyDEConfig {
	return HyDEConfig{
		Enabled:     getSettingOrViperBool("hyde.enabled", "hyde.enabled", false),
		MaxTokens:   getSettingOrViperInt("hyde.max_tokens", "hyde.max_tokens", 200),
		Temperature: 0.7,
	}
}

// SearchWithHyDE 使用 HyDE 增强搜索
// 核心思路：用户查询往往很短，直接做向量搜索效果差。
// HyDE 先让 LLM 生成一个假设的答案文档，再用这个假设文档做向量搜索。
func (h *HybridSearcher) SearchWithHyDE(ctx context.Context, query string, topK int) (*HyDEResult, error) {
	cfg := LoadHyDEConfig()
	result := &HyDEResult{
		OriginalQuery: query,
		HyDEUsed:      false,
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

	// 1. 生成假设文档
	hypoDoc, err := generateHypotheticalDocument(ctx, query, cfg)
	if err != nil {
		log.WithError(err).Warn("embedding: HyDE generation failed, falling back to normal search")
		results, err := h.Search(ctx, query, topK)
		if err != nil {
			return nil, err
		}
		result.SearchResults = results
		return result, nil
	}

	result.HypotheticalDoc = hypoDoc
	result.HyDEUsed = true

	// 2. 混合策略：HyDE 向量结果 + 原始 BM25 结果 + RRF 融合
	h.mu.RLock()
	defer h.mu.RUnlock()

	var hydeResults []SearchResult
	var bm25Results []SearchResult

	// 用假设文档做向量搜索
	if h.vector != nil {
		hydeVec, vecErr := h.vector.SearchByText(hypoDoc, defaultSearchTop)
		if vecErr == nil {
			hydeResults = hydeVec
		}
	}

	// 原始 query 做 BM25 搜索
	if h.bm25 != nil {
		bm25Res, bm25Err := h.bm25.Search(query, defaultSearchTop)
		if bm25Err == nil {
			bm25Results = bm25Res
		}
	}

	// RRF 融合
	if len(hydeResults) > 0 && len(bm25Results) > 0 {
		merged := mergeResults(bm25Results, hydeResults)
		result.SearchResults = limitResults(boostSearchResults(query, merged), topK)
	} else if len(hydeResults) > 0 {
		result.SearchResults = limitResults(boostSearchResults(query, hydeResults), topK)
	} else if len(bm25Results) > 0 {
		result.SearchResults = limitResults(boostSearchResults(query, bm25Results), topK)
	}

	// 如果启用了 reranker，对结果做重排
	if h.reranker != nil && len(result.SearchResults) > 0 {
		reranked, rerankErr := h.reranker.Rerank(ctx, query, result.SearchResults, topK)
		if rerankErr == nil {
			result.SearchResults = reranked
		}
	}

	return result, nil
}

// generateHypotheticalDocument 让 LLM 生成假设答案
func generateHypotheticalDocument(ctx context.Context, query string, cfg HyDEConfig) (string, error) {
	prompt := fmt.Sprintf(
		"请针对以下运维问题，写一段简短的技术解答（200字以内），包含相关的技术术语和概念。"+
			"不需要完全准确，只需要语义相关即可。\n\n问题：%s\n\n解答：", query)

	return llmClientInstance.Generate(ctx, prompt, cfg.MaxTokens, cfg.Temperature)
}

// SearchByText 使用文本直接做向量搜索（不经过分词）
func (e *VectorEngine) SearchByText(text string, topK int) ([]SearchResult, error) {
	vec, err := e.embedder.Embed(context.Background(), text)
	if err != nil {
		return nil, fmt.Errorf("embed text for search: %w", err)
	}

	e.mu.RLock()
	defer e.mu.RUnlock()

	type scored struct {
		doc   *vectorDoc
		score float64
	}
	var results []scored
	for _, doc := range e.docs {
		sim := cosineSimilarity(vec, doc.Embedding)
		results = append(results, scored{doc: doc, score: sim})
	}

	// 排序取 topK
	for i := 0; i < len(results)-1; i++ {
		for j := i + 1; j < len(results); j++ {
			if results[j].score > results[i].score {
				results[i], results[j] = results[j], results[i]
			}
		}
	}
	if len(results) > topK {
		results = results[:topK]
	}

	out := make([]SearchResult, len(results))
	for i, r := range results {
		out[i] = SearchResult{
			DocID:      r.doc.ID,
			Title:      r.doc.Title,
			Content:    r.doc.Content,
			Score:      r.score,
			DocType:    r.doc.DocType,
			Category:   r.doc.Category,
			ParentID:   r.doc.ParentID,
			ChunkIndex: r.doc.ChunkIndex,
		}
	}
	return out, nil
}

