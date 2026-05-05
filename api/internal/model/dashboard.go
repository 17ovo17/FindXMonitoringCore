package model

import (
	"encoding/json"
	"time"
)

const (
	MonitorDashboardStatusActive   = "active"
	MonitorDashboardStatusArchived = "archived"
	MonitorDashboardStatusDraft    = "draft"
)

type MonitorDashboard struct {
	ID              string          `json:"id"`
	Title           string          `json:"title"`
	Description     string          `json:"description,omitempty"`
	WorkspaceID     string          `json:"workspace_id,omitempty"`
	ResourceGroupID string          `json:"resource_group_id,omitempty"`
	Tags            []string        `json:"tags,omitempty"`
	Variables       json.RawMessage `json:"variables,omitempty"`
	Panels          json.RawMessage `json:"panels,omitempty"`
	Version         int             `json:"version"`
	Status          string          `json:"status"`
	Shared          bool            `json:"shared"`
	ShareTokenSet   bool            `json:"share_token_set"`
	ShareSummary    string          `json:"share_summary,omitempty"`
	CreatedBy       string          `json:"created_by,omitempty"`
	UpdatedBy       string          `json:"updated_by,omitempty"`
	CreatedAt       time.Time       `json:"created_at"`
	UpdatedAt       time.Time       `json:"updated_at"`
}

type MonitorDashboardInput struct {
	Title           string          `json:"title"`
	Description     string          `json:"description,omitempty"`
	WorkspaceID     string          `json:"workspace_id,omitempty"`
	ResourceGroupID string          `json:"resource_group_id,omitempty"`
	Tags            []string        `json:"tags,omitempty"`
	Variables       json.RawMessage `json:"variables,omitempty"`
	Panels          json.RawMessage `json:"panels,omitempty"`
	Status          string          `json:"status,omitempty"`
}

type MonitorDashboardShareResult struct {
	ID           string `json:"id"`
	ShareEnabled bool   `json:"share_enabled"`
	ShareSummary string `json:"share_summary"`
}

type MonitorDashboardTemplate struct {
	ID          string          `json:"id"`
	Title       string          `json:"title"`
	Description string          `json:"description,omitempty"`
	Tags        []string        `json:"tags,omitempty"`
	Variables   json.RawMessage `json:"variables,omitempty"`
	Panels      json.RawMessage `json:"panels,omitempty"`
}

type MonitorDashboardTemplateImportInput struct {
	Title           string          `json:"title,omitempty"`
	WorkspaceID     string          `json:"workspace_id,omitempty"`
	ResourceGroupID string          `json:"resource_group_id,omitempty"`
	Variables       json.RawMessage `json:"variables,omitempty"`
	Tags            []string        `json:"tags,omitempty"`
}
