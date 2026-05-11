package model

import "time"

type McpServer struct {
	ID          string    `gorm:"primaryKey;size:32" json:"id"`
	Name        string    `gorm:"size:64;not null" json:"name"`
	Type        string    `gorm:"size:32" json:"type"` // nightingale/cmdb/agent/knowledge/prometheus/alertmanager
	Endpoint    string    `gorm:"size:255" json:"endpoint"`
	Status      string    `gorm:"size:16" json:"status"` // online/offline/error
	Config      string    `gorm:"type:text" json:"config"` // JSON config
	Description string    `gorm:"size:512" json:"description"`
	LastCheckAt time.Time `json:"last_check_at"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}
