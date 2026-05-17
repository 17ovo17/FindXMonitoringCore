package model

import "time"

// AlertInhibitRule 定义告警抑制规则。
// 当存在一条 firing 的源告警匹配 SourceMatch，且当前告警匹配 TargetMatch，
// 且 Equal 中指定的标签值相同时，当前告警被抑制（不发送通知）。
type AlertInhibitRule struct {
	ID          string            `json:"id"`
	Name        string            `json:"name"`
	SourceMatch map[string]string `json:"source_match"` // 源告警匹配条件
	TargetMatch map[string]string `json:"target_match"` // 目标告警匹配条件
	Equal       []string          `json:"equal"`        // 必须相等的标签
	Enabled     bool              `json:"enabled"`
	CreatedBy   string            `json:"created_by,omitempty"`
	CreatedAt   time.Time         `json:"created_at"`
	UpdatedAt   time.Time         `json:"updated_at"`
}
