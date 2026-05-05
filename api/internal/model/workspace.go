package model

import "time"

type Workspace struct {
	ID            string             `json:"id"`
	Name          string             `json:"name"`
	Description   string             `json:"description,omitempty"`
	Owner         string             `json:"owner,omitempty"`
	Status        string             `json:"status"`
	Tags          []string           `json:"tags"`
	Hosts         []string           `json:"hosts"`
	Endpoints     []TopologyEndpoint `json:"endpoints"`
	ResourceCount int                `json:"resource_count"`
	CreatedAt     time.Time          `json:"created_at"`
	UpdatedAt     time.Time          `json:"updated_at"`
}

type WorkspaceInput struct {
	Name        string             `json:"name"`
	Description string             `json:"description"`
	Owner       string             `json:"owner"`
	Status      string             `json:"status"`
	Tags        []string           `json:"tags"`
	Hosts       []string           `json:"hosts"`
	Endpoints   []TopologyEndpoint `json:"endpoints"`
}
