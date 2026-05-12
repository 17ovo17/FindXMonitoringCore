package model

import "time"

// MonitorAlertEventRecord 是告警事件的 GORM 模型，用于 AutoMigrate 自动建表。
// 实际业务读写仍通过 store 层的 MonitorAlertEvent + 原生 SQL 完成，
// 此模型仅作为 schema 定义和未来 GORM 迁移的桥梁。
type MonitorAlertEventRecord struct {
	ID          uint       `gorm:"primaryKey" json:"id"`
	RuleID      uint       `json:"rule_id"`
	RuleName    string     `gorm:"size:255" json:"rule_name"`
	Severity    int        `json:"severity"` // 1=critical, 2=warning, 3=info
	Status      string     `gorm:"size:32;index" json:"status"`
	Labels      string     `gorm:"type:json" json:"labels"`
	Annotations string     `gorm:"type:json" json:"annotations"`
	Value       float64    `json:"value"`
	StartsAt    time.Time  `json:"starts_at"`
	EndsAt      *time.Time `json:"ends_at"`
	AckedBy     string     `gorm:"size:128" json:"acked_by"`
	AckedAt     *time.Time `json:"acked_at"`
	AssignedTo  string     `gorm:"size:128" json:"assigned_to"`
	ResolvedAt  *time.Time `json:"resolved_at"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`
}

// TableName 指定表名。
func (MonitorAlertEventRecord) TableName() string {
	return "alert_events"
}
