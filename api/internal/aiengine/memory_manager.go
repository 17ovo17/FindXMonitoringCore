package aiengine

import (
	"sort"
	"strings"
	"sync"
	"time"

	"ai-workbench-api/internal/model"
	"ai-workbench-api/internal/store"

	"github.com/google/uuid"
)

// ---------------------------------------------------------------------------
// MemoryManager — 跨会话记忆持久化
// ---------------------------------------------------------------------------
//
// 存储：诊断结论、修复方案、学到的模式、事实
// 参考 Hermes memory_manager.py：单一集成点，管理多种记忆类型

// MemoryType 记忆类型
type MemoryType string

const (
	MemoryDiagnosis MemoryType = "diagnosis" // 诊断结论
	MemoryRunbook   MemoryType = "runbook"   // 修复方案
	MemoryPattern   MemoryType = "pattern"   // 学到的模式
	MemoryFact      MemoryType = "fact"      // 事实（如"Redis 主从切换需要 30s"）
)

// Memory 表示一条记忆
type Memory struct {
	ID        string     `json:"id"`
	Type      MemoryType `json:"type"`
	Content   string     `json:"content"`
	Tags      []string   `json:"tags"`                // 关联标签（主机/服务/告警规则）
	Score     float64    `json:"score"`               // 相关性评分
	UsedCount int        `json:"used_count"`          // 被引用次数
	CreatedAt time.Time  `json:"created_at"`
	ExpiresAt *time.Time `json:"expires_at,omitempty"` // 过期时间
}

// MemoryManager 管理所有记忆的存储和召回
type MemoryManager struct {
	mu       sync.RWMutex
	memories map[string]*Memory // id -> Memory
	index    map[string][]string // tag -> []memory_id（倒排索引）
}

// NewMemoryManager 创建记忆管理器
func NewMemoryManager() *MemoryManager {
	return &MemoryManager{
		memories: make(map[string]*Memory),
		index:    make(map[string][]string),
	}
}

// Store 存储一条记忆
func (mm *MemoryManager) Store(memType MemoryType, content string, tags []string, expiresAt *time.Time) string {
	mm.mu.Lock()
	defer mm.mu.Unlock()

	id := uuid.New().String()
	m := &Memory{
		ID:        id,
		Type:      memType,
		Content:   content,
		Tags:      tags,
		Score:     1.0,
		UsedCount: 0,
		CreatedAt: time.Now(),
		ExpiresAt: expiresAt,
	}
	mm.memories[id] = m

	// 更新倒排索引
	for _, tag := range tags {
		mm.index[tag] = append(mm.index[tag], id)
	}

	// 持久化到 MySQL + Redis
	mm.persist(m)

	return id
}

// Recall 根据 query 召回相关记忆（基于关键词匹配 + 标签匹配）
func (mm *MemoryManager) Recall(query string, limit int) []*Memory {
	mm.mu.RLock()
	defer mm.mu.RUnlock()

	now := time.Now()
	scored := make([]*Memory, 0, len(mm.memories))

	for _, m := range mm.memories {
		// 跳过过期记忆
		if m.ExpiresAt != nil && now.After(*m.ExpiresAt) {
			continue
		}
		// 计算相关性评分
		score := mm.calculateRelevance(query, m)
		if score > 0 {
			// 创建副本避免修改原始数据
			copy := *m
			copy.Score = score
			scored = append(scored, &copy)
		}
	}

	// 按评分降序排列
	sort.Slice(scored, func(i, j int) bool {
		return scored[i].Score > scored[j].Score
	})

	if limit > 0 && len(scored) > limit {
		scored = scored[:limit]
	}

	// 更新引用计数
	mm.mu.RUnlock()
	mm.mu.Lock()
	for _, m := range scored {
		if orig, ok := mm.memories[m.ID]; ok {
			orig.UsedCount++
		}
	}
	mm.mu.Unlock()
	mm.mu.RLock()

	return scored
}

// RecallByType 按类型召回相关记忆
func (mm *MemoryManager) RecallByType(query string, memType MemoryType, limit int) []*Memory {
	mm.mu.RLock()
	defer mm.mu.RUnlock()

	now := time.Now()
	scored := make([]*Memory, 0)

	for _, m := range mm.memories {
		if m.Type != memType {
			continue
		}
		if m.ExpiresAt != nil && now.After(*m.ExpiresAt) {
			continue
		}
		score := mm.calculateRelevance(query, m)
		if score > 0 {
			copy := *m
			copy.Score = score
			scored = append(scored, &copy)
		}
	}

	sort.Slice(scored, func(i, j int) bool {
		return scored[i].Score > scored[j].Score
	})

	if limit > 0 && len(scored) > limit {
		scored = scored[:limit]
	}
	return scored
}

// RecallByTags 按标签召回记忆
func (mm *MemoryManager) RecallByTags(tags []string, limit int) []*Memory {
	mm.mu.RLock()
	defer mm.mu.RUnlock()

	now := time.Now()
	seen := make(map[string]bool)
	result := make([]*Memory, 0)

	for _, tag := range tags {
		ids, ok := mm.index[tag]
		if !ok {
			continue
		}
		for _, id := range ids {
			if seen[id] {
				continue
			}
			seen[id] = true
			m, exists := mm.memories[id]
			if !exists {
				continue
			}
			if m.ExpiresAt != nil && now.After(*m.ExpiresAt) {
				continue
			}
			result = append(result, m)
		}
	}

	if limit > 0 && len(result) > limit {
		result = result[:limit]
	}
	return result
}

// Forget 删除过期或低分记忆
func (mm *MemoryManager) Forget() int {
	mm.mu.Lock()
	defer mm.mu.Unlock()

	now := time.Now()
	removed := 0

	for id, m := range mm.memories {
		shouldRemove := false
		// 过期删除
		if m.ExpiresAt != nil && now.After(*m.ExpiresAt) {
			shouldRemove = true
		}
		// 长期未使用且评分低的 pattern 类记忆删除
		if m.Type == MemoryPattern && m.UsedCount == 0 {
			age := now.Sub(m.CreatedAt)
			if age > 30*24*time.Hour { // 30 天未使用
				shouldRemove = true
			}
		}
		if shouldRemove {
			// 清理倒排索引
			for _, tag := range m.Tags {
				mm.removeFromIndex(tag, id)
			}
			delete(mm.memories, id)
			removed++
		}
	}
	return removed
}

// Learn 从诊断结果中自动提取记忆
func (mm *MemoryManager) Learn(memType MemoryType, content string, tags []string) string {
	// 检查是否已有相似记忆（避免重复）
	mm.mu.RLock()
	for _, m := range mm.memories {
		if m.Type == memType && mm.isSimilar(m.Content, content) {
			mm.mu.RUnlock()
			// 更新已有记忆的引用计数
			mm.mu.Lock()
			m.UsedCount++
			mm.mu.Unlock()
			return m.ID
		}
	}
	mm.mu.RUnlock()

	// 存储新记忆
	return mm.Store(memType, content, tags, nil)
}

// Count 返回记忆总数
func (mm *MemoryManager) Count() int {
	mm.mu.RLock()
	defer mm.mu.RUnlock()
	return len(mm.memories)
}

// ---------------------------------------------------------------------------
// 内部方法
// ---------------------------------------------------------------------------

// calculateRelevance 计算 query 与记忆的相关性（简化版 BM25）
func (mm *MemoryManager) calculateRelevance(query string, m *Memory) float64 {
	queryLower := strings.ToLower(query)
	contentLower := strings.ToLower(m.Content)

	score := 0.0

	// 关键词匹配
	words := strings.Fields(queryLower)
	for _, word := range words {
		if len(word) < 2 {
			continue
		}
		if strings.Contains(contentLower, word) {
			score += 1.0
		}
	}

	// 标签匹配加分
	for _, tag := range m.Tags {
		if strings.Contains(queryLower, strings.ToLower(tag)) {
			score += 2.0
		}
	}

	// 使用频率加分（被引用越多越可能有价值）
	if m.UsedCount > 0 {
		score += float64(m.UsedCount) * 0.1
	}

	// 时间衰减（越新越相关）
	age := time.Since(m.CreatedAt).Hours()
	if age > 0 {
		score *= 1.0 / (1.0 + age/720) // 30 天半衰期
	}

	return score
}

// isSimilar 简单判断两段内容是否相似（Jaccard 系数）
func (mm *MemoryManager) isSimilar(a, b string) bool {
	wordsA := strings.Fields(strings.ToLower(a))
	wordsB := strings.Fields(strings.ToLower(b))

	if len(wordsA) == 0 || len(wordsB) == 0 {
		return false
	}

	setA := make(map[string]bool, len(wordsA))
	for _, w := range wordsA {
		setA[w] = true
	}

	intersection := 0
	for _, w := range wordsB {
		if setA[w] {
			intersection++
		}
	}

	union := len(setA) + len(wordsB) - intersection
	if union == 0 {
		return false
	}

	jaccard := float64(intersection) / float64(union)
	return jaccard > 0.7 // 相似度阈值
}

// removeFromIndex 从倒排索引中移除
func (mm *MemoryManager) removeFromIndex(tag, id string) {
	ids, ok := mm.index[tag]
	if !ok {
		return
	}
	for i, existingID := range ids {
		if existingID == id {
			mm.index[tag] = append(ids[:i], ids[i+1:]...)
			break
		}
	}
	if len(mm.index[tag]) == 0 {
		delete(mm.index, tag)
	}
}

// persist 将记忆持久化到 MySQL + Redis（通过 store 层）
func (mm *MemoryManager) persist(m *Memory) {
	go store.SaveAIMemory(&model.AIMemory{
		ID:        m.ID,
		Type:      string(m.Type),
		Content:   m.Content,
		Tags:      m.Tags,
		Score:     m.Score,
		UsedCount: m.UsedCount,
		CreatedAt: m.CreatedAt,
		ExpiresAt: m.ExpiresAt,
	})
}
