package aiengine

import (
	"strings"
	"sync"
	"time"
)

// ---------------------------------------------------------------------------
// Curator — 策展器：选择最相关的上下文注入到 prompt
// ---------------------------------------------------------------------------
//
// 核心理念：不是把所有历史都塞进去，而是选择最相关的。
// 参考 Hermes curator.py：基于活跃度和相关性的智能选择。

// ConversationTurn 表示一轮对话
type ConversationTurn struct {
	Role      string    `json:"role"`
	Content   string    `json:"content"`
	Timestamp time.Time `json:"timestamp"`
	Intent    IntentType `json:"intent,omitempty"`
}

// SessionHistory 会话历史
type SessionHistory struct {
	SessionID string             `json:"session_id"`
	Turns     []ConversationTurn `json:"turns"`
	Summary   string             `json:"summary"` // 压缩后的摘要
	UpdatedAt time.Time          `json:"updated_at"`
}

// Curator 策展器
type Curator struct {
	mu            sync.RWMutex
	memoryManager *MemoryManager
	maxHistory    int // 最多保留几轮对话
	sessions      map[string]*SessionHistory
}

// NewCurator 创建策展器
func NewCurator(mm *MemoryManager, maxHistory int) *Curator {
	if maxHistory <= 0 {
		maxHistory = 5
	}
	return &Curator{
		memoryManager: mm,
		maxHistory:    maxHistory,
		sessions:      make(map[string]*SessionHistory),
	}
}

// AddTurn 记录一轮对话
func (c *Curator) AddTurn(sessionID string, role string, content string, intent IntentType) {
	c.mu.Lock()
	defer c.mu.Unlock()

	session, ok := c.sessions[sessionID]
	if !ok {
		session = &SessionHistory{
			SessionID: sessionID,
			Turns:     make([]ConversationTurn, 0, c.maxHistory*2),
		}
		c.sessions[sessionID] = session
	}

	session.Turns = append(session.Turns, ConversationTurn{
		Role:      role,
		Content:   content,
		Timestamp: time.Now(),
		Intent:    intent,
	})
	session.UpdatedAt = time.Now()

	// 超过最大轮数时压缩旧历史
	maxTurns := c.maxHistory * 2 // user + assistant 各算一轮
	if len(session.Turns) > maxTurns+4 { // 保留缓冲
		c.compressOldTurns(session)
	}
}

// SelectRelevantHistory 从会话历史中选择与当前 query 最相关的片段
// 策略：
// 1. 最近 2 轮对话始终保留（短期记忆）
// 2. 更早的对话只保留与当前 query 相关的（通过关键词匹配）
// 3. 跨会话的记忆通过 MemoryManager 召回
func (c *Curator) SelectRelevantHistory(sessionID string, query string) []ContextBlock {
	c.mu.RLock()
	defer c.mu.RUnlock()

	blocks := make([]ContextBlock, 0, 3)

	session, ok := c.sessions[sessionID]
	if !ok {
		return blocks
	}

	turns := session.Turns
	if len(turns) == 0 {
		return blocks
	}

	// 1. 最近 2 轮始终保留
	recentStart := len(turns) - 4 // 2 轮 = 4 条消息（user+assistant）
	if recentStart < 0 {
		recentStart = 0
	}
	recentTurns := turns[recentStart:]
	if len(recentTurns) > 0 {
		var sb strings.Builder
		sb.WriteString("[最近对话]\n")
		for _, t := range recentTurns {
			sb.WriteString(t.Role)
			sb.WriteString(": ")
			// 截断过长内容
			content := t.Content
			if len([]rune(content)) > 200 {
				content = string([]rune(content)[:200]) + "..."
			}
			sb.WriteString(content)
			sb.WriteString("\n")
		}
		content := sb.String()
		blocks = append(blocks, ContextBlock{
			Role:     BlockRoleHistory,
			Content:  content,
			Priority: 55,
			Tokens:   estimateTokens(content),
			Source:   "recent_history",
		})
	}

	// 2. 更早的对话：只保留与 query 相关的
	if recentStart > 0 {
		olderTurns := turns[:recentStart]
		relevant := c.filterRelevantTurns(olderTurns, query)
		if len(relevant) > 0 {
			var sb strings.Builder
			sb.WriteString("[相关历史]\n")
			for _, t := range relevant {
				sb.WriteString(t.Role)
				sb.WriteString(": ")
				content := t.Content
				if len([]rune(content)) > 150 {
					content = string([]rune(content)[:150]) + "..."
				}
				sb.WriteString(content)
				sb.WriteString("\n")
			}
			content := sb.String()
			blocks = append(blocks, ContextBlock{
				Role:     BlockRoleHistory,
				Content:  content,
				Priority: 35,
				Tokens:   estimateTokens(content),
				Source:   "relevant_history",
			})
		}
	}

	// 3. 注入压缩摘要（如果有）
	if session.Summary != "" {
		blocks = append(blocks, ContextBlock{
			Role:     BlockRoleHistory,
			Content:  "[会话摘要] " + session.Summary,
			Priority: 30,
			Tokens:   estimateTokens(session.Summary),
			Source:   "session_summary",
		})
	}

	return blocks
}

// InjectRelevantMemories 注入跨会话记忆
func (c *Curator) InjectRelevantMemories(query string, tags []string) []ContextBlock {
	blocks := make([]ContextBlock, 0, 2)

	// 按标签召回
	if len(tags) > 0 {
		memories := c.memoryManager.RecallByTags(tags, 3)
		if len(memories) > 0 {
			var sb strings.Builder
			sb.WriteString("[跨会话记忆 - 标签匹配]\n")
			for _, m := range memories {
				sb.WriteString("- [")
				sb.WriteString(string(m.Type))
				sb.WriteString("] ")
				sb.WriteString(m.Content)
				sb.WriteString("\n")
			}
			content := sb.String()
			blocks = append(blocks, ContextBlock{
				Role:     BlockRoleHistory,
				Content:  content,
				Priority: 45,
				Tokens:   estimateTokens(content),
				Source:   "cross_session_memory",
			})
		}
	}

	// 按语义召回
	memories := c.memoryManager.Recall(query, 3)
	if len(memories) > 0 {
		var sb strings.Builder
		sb.WriteString("[跨会话记忆 - 语义匹配]\n")
		for _, m := range memories {
			sb.WriteString("- [")
			sb.WriteString(string(m.Type))
			sb.WriteString("] ")
			sb.WriteString(m.Content)
			sb.WriteString("\n")
		}
		content := sb.String()
		blocks = append(blocks, ContextBlock{
			Role:     BlockRoleHistory,
			Content:  content,
			Priority: 40,
			Tokens:   estimateTokens(content),
			Source:   "semantic_memory",
		})
	}

	return blocks
}

// GetSessionSummary 获取会话摘要
func (c *Curator) GetSessionSummary(sessionID string) string {
	c.mu.RLock()
	defer c.mu.RUnlock()
	if session, ok := c.sessions[sessionID]; ok {
		return session.Summary
	}
	return ""
}

// ClearSession 清除会话历史
func (c *Curator) ClearSession(sessionID string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	delete(c.sessions, sessionID)
}

// ---------------------------------------------------------------------------
// 内部方法
// ---------------------------------------------------------------------------

// filterRelevantTurns 从历史对话中筛选与 query 相关的轮次
func (c *Curator) filterRelevantTurns(turns []ConversationTurn, query string) []ConversationTurn {
	queryLower := strings.ToLower(query)
	words := strings.Fields(queryLower)

	relevant := make([]ConversationTurn, 0)
	for _, t := range turns {
		contentLower := strings.ToLower(t.Content)
		matchCount := 0
		for _, word := range words {
			if len(word) < 2 {
				continue
			}
			if strings.Contains(contentLower, word) {
				matchCount++
			}
		}
		// 至少匹配 2 个关键词才认为相关
		if matchCount >= 2 || (len(words) == 1 && matchCount >= 1) {
			relevant = append(relevant, t)
		}
	}

	// 最多返回 3 条相关历史
	if len(relevant) > 3 {
		relevant = relevant[len(relevant)-3:]
	}
	return relevant
}

// compressOldTurns 压缩旧对话为摘要
func (c *Curator) compressOldTurns(session *SessionHistory) {
	maxTurns := c.maxHistory * 2
	if len(session.Turns) <= maxTurns {
		return
	}

	// 保留最近 maxTurns 条，压缩更早的
	oldTurns := session.Turns[:len(session.Turns)-maxTurns]
	session.Turns = session.Turns[len(session.Turns)-maxTurns:]

	// 生成摘要（简化版：提取关键信息）
	var sb strings.Builder
	if session.Summary != "" {
		sb.WriteString(session.Summary)
		sb.WriteString(" | ")
	}
	topics := make(map[string]bool)
	for _, t := range oldTurns {
		if t.Role == "user" {
			// 提取前 50 字符作为话题
			content := t.Content
			if len([]rune(content)) > 50 {
				content = string([]rune(content)[:50])
			}
			topics[content] = true
		}
	}
	sb.WriteString("历史话题: ")
	for topic := range topics {
		sb.WriteString(topic)
		sb.WriteString("; ")
	}
	session.Summary = sb.String()
}
