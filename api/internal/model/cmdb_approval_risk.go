package model

import "time"

type CmdbResourceApproval struct {
	ID            string     `gorm:"primaryKey;size:64" json:"id"`
	View          string     `gorm:"size:32;index" json:"view"`
	Requester     string     `gorm:"size:64;index" json:"requester"`
	Approver      string     `gorm:"size:64;index" json:"approver"`
	ResourceType  string     `gorm:"size:64;index" json:"resource_type"`
	ResourceID    string     `gorm:"size:128;index" json:"resource_id"`
	Action        string     `gorm:"size:64;index" json:"action"`
	RiskLevel     string     `gorm:"size:32;index" json:"risk_level"`
	Status        string     `gorm:"size:32;index" json:"status"`
	Title         string     `gorm:"size:256" json:"title"`
	Summary       string     `gorm:"type:text" json:"summary"`
	BusinessGroup string     `gorm:"size:128;index" json:"business_group"`
	WorkflowState string     `gorm:"size:64;index" json:"workflow_state"`
	RiskRecordID  string     `gorm:"size:64;index" json:"risk_record_id"`
	ContextJSON   string     `gorm:"type:text" json:"-"`
	DiffJSON      string     `gorm:"type:text" json:"-"`
	AuditRef      string     `gorm:"size:128" json:"audit_ref"`
	CreatedAt     time.Time  `json:"created_at"`
	UpdatedAt     time.Time  `json:"updated_at"`
	DecidedAt     *time.Time `json:"decided_at,omitempty"`
	DecisionActor string     `gorm:"size:64" json:"decision_actor"`
	DecisionNote  string     `gorm:"type:text" json:"decision_note"`
}

type CmdbOperationRiskRecord struct {
	ID            string    `gorm:"primaryKey;size:64" json:"id"`
	ResourceType  string    `gorm:"size:64;index" json:"resource_type"`
	ResourceID    string    `gorm:"size:128;index" json:"resource_id"`
	Action        string    `gorm:"size:64;index" json:"action"`
	RiskLevel     string    `gorm:"size:32;index" json:"risk_level"`
	PolicyID      string    `gorm:"size:128;index" json:"policy_id"`
	Status        string    `gorm:"size:32;index" json:"status"`
	Actor         string    `gorm:"size:64;index" json:"actor"`
	BusinessGroup string    `gorm:"size:128;index" json:"business_group"`
	Reason        string    `gorm:"type:text" json:"reason"`
	ContextJSON   string    `gorm:"type:text" json:"-"`
	AuditRef      string    `gorm:"size:128" json:"audit_ref"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
}
