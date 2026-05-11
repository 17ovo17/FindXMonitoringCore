package model

import "time"

// AgentPackage represents a distributable agent binary package.
type AgentPackage struct {
	ID        string    `gorm:"primaryKey;size:32" json:"id"`
	Name      string    `gorm:"size:64;not null" json:"name"`
	Version   string    `gorm:"size:32;not null" json:"version"`
	Platform  string    `gorm:"size:16" json:"platform"` // linux/windows/darwin
	Arch      string    `gorm:"size:16" json:"arch"`     // amd64/arm64
	Size      int64     `json:"size"`
	Checksum  string    `gorm:"size:64" json:"checksum"`
	URL       string    `gorm:"size:512" json:"url"`
	CreatedAt time.Time `json:"created_at"`
}

// AgentLifecycleEvent represents a single lifecycle evidence entry.
type AgentLifecycleEvent struct {
	ID        string    `gorm:"primaryKey;size:32" json:"id"`
	AgentID   string    `gorm:"size:64;index" json:"agent_id"`
	Action    string    `gorm:"size:32" json:"action"` // install/upgrade/rollback/uninstall/config-push
	Status    string    `gorm:"size:32" json:"status"` // pending/running/succeeded/failed/blocked
	Detail    string    `gorm:"type:text" json:"detail,omitempty"`
	Operator  string    `gorm:"size:128" json:"operator,omitempty"`
	CreatedAt time.Time `json:"created_at"`
}
