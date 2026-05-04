package model

import "time"

type MonitorAuditLog struct {
	ID           string         `json:"id"`
	CreatedAt    time.Time      `json:"created_at"`
	Actor        string         `json:"actor"`
	Action       string         `json:"action"`
	ResourceType string         `json:"resource_type"`
	ResourceID   string         `json:"resource_id"`
	Scope        string         `json:"scope"`
	Status       string         `json:"status"`
	TraceID      string         `json:"trace_id"`
	ClientIP     string         `json:"client_ip"`
	Summary      string         `json:"summary"`
	Details      map[string]any `json:"details"`
}

type MonitorAuditLogQuery struct {
	Page         int
	Limit        int
	Action       string
	ResourceType string
	ResourceID   string
	Status       string
	TraceID      string
	Scope        string
}

type MonitorAuditLogPage struct {
	Items []MonitorAuditLog `json:"items"`
	Total int               `json:"total"`
	Page  int               `json:"page"`
	Limit int               `json:"limit"`
}
