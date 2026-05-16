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
	HostID          string                `json:"host_id"`
	Ident           string                `json:"ident"`
	Hostname        string                `json:"hostname,omitempty"`
	IPList          []string              `json:"ip_list"`
	OS              string                `json:"os,omitempty"`
	Arch            string                `json:"arch,omitempty"`
	WorkspaceID     string                `json:"workspace_id,omitempty"`
	ResourceGroupID string                `json:"resource_group_id,omitempty"`
	Tags            []string              `json:"tags"`
	AgentID         string                `json:"agent_id,omitempty"`
	AgentStatus     string                `json:"agent_status,omitempty"`
	AgentVersion    string                `json:"agent_version,omitempty"`
	LastSeenAt      *time.Time            `json:"last_seen_at,omitempty"`
	Status          string                `json:"status"`
	Source          string                `json:"source"`
	Labels          map[string]string     `json:"labels"`
	CmdbInstance    *HostAssetCmdbRef     `json:"cmdb_instance,omitempty"`
	CmdbColumns     []HostAssetCmdbColumn `json:"cmdb_columns,omitempty"`
	CmdbValues      map[string]any        `json:"cmdb_values,omitempty"`
	UpdatedAt       time.Time             `json:"updated_at"`
}

type HostAssetCmdbRef struct {
	InstanceID string `json:"instance_id"`
	ObjectID   string `json:"object_id"`
	ObjectName string `json:"object_name"`
	Source     string `json:"source"`
}

type HostAssetCmdbColumn struct {
	Attr      string `json:"attr"`
	Label     string `json:"label"`
	ValueType string `json:"value_type"`
	Tag       string `json:"tag"`
	Unit      string `json:"unit,omitempty"`
	Sort      int    `json:"sort"`
	Visible   bool   `json:"visible"`
	Sensitive bool   `json:"sensitive"`
	Masked    bool   `json:"masked"`
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
