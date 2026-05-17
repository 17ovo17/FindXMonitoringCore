package embedding

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"os"
	"sort"
	"strings"
	"sync"
	"time"

	log "github.com/sirupsen/logrus"
)

// FeedbackType 反馈类型
type FeedbackType string

const (
	FeedbackLike    FeedbackType = "like"
	FeedbackDislike FeedbackType = "dislike"
	FeedbackClick   FeedbackType = "click"
	FeedbackSkip    FeedbackType = "skip"
)

// feedbackWeights 各反馈类型对应的分数调整权重
var feedbackWeights = map[FeedbackType]float64{
	FeedbackLike:    0.15,
	FeedbackDislike: -0.20,
	FeedbackClick:   0.05,
	FeedbackSkip:    -0.03,
}

// FeedbackEntry 反馈记录
type FeedbackEntry struct {
	Query     string       `json:"query"`
	DocID     string       `json:"doc_id"`
	Type      FeedbackType `json:"type"`
	UserID    string       `json:"user_id"`
	Timestamp time.Time    `json:"timestamp"`
}

// FeedbackStats 反馈统计
type FeedbackStats struct {
	TotalFeedback int              `json:"total_feedback"`
	ByType        map[string]int   `json:"by_type"`
	TopLiked      []DocFeedback    `json:"top_liked"`
	TopDisliked   []DocFeedback    `json:"top_disliked"`
	RecentEntries []FeedbackEntry  `json:"recent_entries"`
}

// DocFeedback 文档反馈统计
type DocFeedback struct {
	DocID    string `json:"doc_id"`
	Likes    int    `json:"likes"`
	Dislikes int    `json:"dislikes"`
	Score    int    `json:"score"` // likes - dislikes
}

// FeedbackLearner 从用户反馈中学习，调整搜索排序
type FeedbackLearner struct {
	// queryHash -> docID -> score_adjustment
	adjustments map[string]map[string]float64
	// 所有反馈记录（用于统计）
	entries []FeedbackEntry
	// 文档级统计
	docLikes    map[string]int
	docDislikes map[string]int
	mu          sync.RWMutex
	persistPath string
}

// globalFeedback 全局反馈学习实例
var (
	globalFeedback *FeedbackLearner
	feedbackOnce   sync.Once
)

// GetFeedbackLearner 获取全局反馈学习实例
func GetFeedbackLearner() *FeedbackLearner {
	feedbackOnce.Do(func() {
		globalFeedback = NewFeedbackLearner("data/feedback.json")
		globalFeedback.Load()
	})
	return globalFeedback
}

// NewFeedbackLearner 创建反馈学习器
func NewFeedbackLearner(persistPath string) *FeedbackLearner {
	return &FeedbackLearner{
		adjustments: make(map[string]map[string]float64),
		entries:     make([]FeedbackEntry, 0),
		docLikes:    make(map[string]int),
		docDislikes: make(map[string]int),
		persistPath: persistPath,
	}
}

// Record 记录反馈
func (f *FeedbackLearner) Record(entry FeedbackEntry) {
	f.mu.Lock()
	defer f.mu.Unlock()

	if entry.Timestamp.IsZero() {
		entry.Timestamp = time.Now()
	}

	// 记录到历史
	f.entries = append(f.entries, entry)

	// 更新 query-doc 调整分数
	qHash := queryHash(entry.Query)
	if f.adjustments[qHash] == nil {
		f.adjustments[qHash] = make(map[string]float64)
	}
	weight := feedbackWeights[entry.Type]
	f.adjustments[qHash][entry.DocID] += weight

	// 更新文档级统计
	switch entry.Type {
	case FeedbackLike:
		f.docLikes[entry.DocID]++
	case FeedbackDislike:
		f.docDislikes[entry.DocID]++
	}

	// 每 50 条反馈持久化一次
	if len(f.entries)%50 == 0 {
		go f.persist()
	}
}

// AdjustScores 根据反馈调整搜索结果分数
func (f *FeedbackLearner) AdjustScores(query string, results []SearchResult) []SearchResult {
	f.mu.RLock()
	defer f.mu.RUnlock()

	qHash := queryHash(query)
	docAdj := f.adjustments[qHash]

	adjusted := make([]SearchResult, len(results))
	copy(adjusted, results)

	for i := range adjusted {
		// query-doc 级别调整
		if docAdj != nil {
			if adj, ok := docAdj[adjusted[i].DocID]; ok {
				adjusted[i].Score += adj
			}
		}
		// 文档全局声誉调整
		likes := f.docLikes[adjusted[i].DocID]
		dislikes := f.docDislikes[adjusted[i].DocID]
		if likes+dislikes > 0 {
			reputation := float64(likes-dislikes) / float64(likes+dislikes) * 0.1
			adjusted[i].Score += reputation
		}
	}

	// 重新排序
	sort.SliceStable(adjusted, func(i, j int) bool {
		return adjusted[i].Score > adjusted[j].Score
	})
	return adjusted
}

// GetStats 获取反馈统计
func (f *FeedbackLearner) GetStats() FeedbackStats {
	f.mu.RLock()
	defer f.mu.RUnlock()

	stats := FeedbackStats{
		TotalFeedback: len(f.entries),
		ByType:        make(map[string]int),
	}

	for _, e := range f.entries {
		stats.ByType[string(e.Type)]++
	}

	// Top liked/disliked
	var docFeedbacks []DocFeedback
	allDocs := make(map[string]bool)
	for docID := range f.docLikes {
		allDocs[docID] = true
	}
	for docID := range f.docDislikes {
		allDocs[docID] = true
	}
	for docID := range allDocs {
		docFeedbacks = append(docFeedbacks, DocFeedback{
			DocID:    docID,
			Likes:    f.docLikes[docID],
			Dislikes: f.docDislikes[docID],
			Score:    f.docLikes[docID] - f.docDislikes[docID],
		})
	}

	sort.Slice(docFeedbacks, func(i, j int) bool {
		return docFeedbacks[i].Score > docFeedbacks[j].Score
	})

	if len(docFeedbacks) > 10 {
		stats.TopLiked = docFeedbacks[:10]
	} else {
		stats.TopLiked = docFeedbacks
	}

	// Top disliked（从末尾取）
	sort.Slice(docFeedbacks, func(i, j int) bool {
		return docFeedbacks[i].Score < docFeedbacks[j].Score
	})
	if len(docFeedbacks) > 10 {
		stats.TopDisliked = docFeedbacks[:10]
	} else {
		stats.TopDisliked = docFeedbacks
	}

	// 最近 20 条
	recentCount := 20
	if len(f.entries) < recentCount {
		recentCount = len(f.entries)
	}
	stats.RecentEntries = f.entries[len(f.entries)-recentCount:]

	return stats
}

// persist 持久化反馈数据到文件
func (f *FeedbackLearner) persist() {
	f.mu.RLock()
	defer f.mu.RUnlock()

	if f.persistPath == "" {
		return
	}

	data := struct {
		Adjustments map[string]map[string]float64 `json:"adjustments"`
		Entries     []FeedbackEntry               `json:"entries"`
		DocLikes    map[string]int                `json:"doc_likes"`
		DocDislikes map[string]int                `json:"doc_dislikes"`
	}{
		Adjustments: f.adjustments,
		Entries:     f.entries,
		DocLikes:    f.docLikes,
		DocDislikes: f.docDislikes,
	}

	bytes, err := json.Marshal(data)
	if err != nil {
		log.WithError(err).Warn("feedback: persist marshal failed")
		return
	}

	if err := os.MkdirAll("data", 0o755); err != nil {
		log.WithError(err).Warn("feedback: create data dir failed")
		return
	}

	if err := os.WriteFile(f.persistPath, bytes, 0o644); err != nil {
		log.WithError(err).Warn("feedback: persist write failed")
	}
}

// Load 从文件加载反馈数据
func (f *FeedbackLearner) Load() {
	if f.persistPath == "" {
		return
	}

	data, err := os.ReadFile(f.persistPath)
	if err != nil {
		return // 文件不存在是正常的
	}

	var stored struct {
		Adjustments map[string]map[string]float64 `json:"adjustments"`
		Entries     []FeedbackEntry               `json:"entries"`
		DocLikes    map[string]int                `json:"doc_likes"`
		DocDislikes map[string]int                `json:"doc_dislikes"`
	}

	if err := json.Unmarshal(data, &stored); err != nil {
		log.WithError(err).Warn("feedback: load unmarshal failed")
		return
	}

	f.mu.Lock()
	defer f.mu.Unlock()

	if stored.Adjustments != nil {
		f.adjustments = stored.Adjustments
	}
	if stored.Entries != nil {
		f.entries = stored.Entries
	}
	if stored.DocLikes != nil {
		f.docLikes = stored.DocLikes
	}
	if stored.DocDislikes != nil {
		f.docDislikes = stored.DocDislikes
	}

	log.WithField("entries", len(f.entries)).Info("feedback: loaded from disk")
}

// queryHash 对 query 做哈希（归一化后）
func queryHash(query string) string {
	normalized := strings.ToLower(strings.TrimSpace(query))
	h := sha256.Sum256([]byte(normalized))
	return hex.EncodeToString(h[:8]) // 取前 8 字节足够
}
