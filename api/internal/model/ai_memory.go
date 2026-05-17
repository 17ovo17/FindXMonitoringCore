package model

import "time"

// AIMemory 表示一条 AI 记忆（MySQL 基石 + Redis 缓存 + 内置向量加速）
type AIMemory struct {
	ID        string    `json:"id"`
	Type      string    `json:"type"`
	Content   string    `json:"content"`
	Tags      []string  `json:"tags"`
	Score     float64   `json:"score"`
	UsedCount int       `json:"used_count"`
	SessionID string    `json:"session_id,omitempty"`
	CreatedAt time.Time `json:"created_at"`
	ExpiresAt *time.Time `json:"expires_at,omitempty"`
}
