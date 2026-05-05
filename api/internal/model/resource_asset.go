package model

import "time"

type ResourceGroup struct {
	ID          string    `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description,omitempty"`
	WorkspaceID string    `json:"workspace_id,omitempty"`
	ParentID    string    `json:"parent_id,omitempty"`
	Status      string    `json:"status"`
	Tags        []string  `json:"tags"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type ResourceGroupInput struct {
	Name        string   `json:"name"`
	Description string   `json:"description"`
	WorkspaceID string   `json:"workspace_id"`
	ParentID    string   `json:"parent_id"`
	Status      string   `json:"status"`
	Tags        []string `json:"tags"`
}

type HostAsset struct {
	HostID          string            `json:"host_id"`
	Ident           string            `json:"ident"`
	Hostname        string            `json:"hostname,omitempty"`
	IPList          []string          `json:"ip_list"`
	OS              string            `json:"os,omitempty"`
	Arch            string            `json:"arch,omitempty"`
	WorkspaceID     string            `json:"workspace_id,omitempty"`
	ResourceGroupID string            `json:"resource_group_id,omitempty"`
	Tags            []string          `json:"tags"`
	AgentID         string            `json:"agent_id,omitempty"`
	AgentStatus     string            `json:"agent_status,omitempty"`
	AgentVersion    string            `json:"agent_version,omitempty"`
	LastSeenAt      *time.Time        `json:"last_seen_at,omitempty"`
	Status          string            `json:"status"`
	Source          string            `json:"source"`
	Labels          map[string]string `json:"labels"`
	UpdatedAt       time.Time         `json:"updated_at"`
}

type HostTagsInput struct {
	Tags []string `json:"tags"`
}

type HostResourceGroupInput struct {
	ResourceGroupID string `json:"resource_group_id"`
}

type HostWorkspaceInput struct {
	WorkspaceID string `json:"workspace_id"`
}
