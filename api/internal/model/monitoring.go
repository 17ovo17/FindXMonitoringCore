package model

import "time"

type MonitorTarget struct {
	ID            string            `json:"id"`
	Ident         string            `json:"ident"`
	Name          string            `json:"name"`
	IP            string            `json:"ip"`
	Hostname      string            `json:"hostname,omitempty"`
	OS            string            `json:"os,omitempty"`
	Arch          string            `json:"arch,omitempty"`
	Environment   string            `json:"environment,omitempty"`
	BusinessGroup string            `json:"business_group,omitempty"`
	Owner         string            `json:"owner,omitempty"`
	Status        string            `json:"status"`
	Source        string            `json:"source"`
	Labels        map[string]string `json:"labels,omitempty"`
	Metadata      map[string]string `json:"metadata,omitempty"`
	LastSeen      *time.Time        `json:"last_seen,omitempty"`
	CreatedAt     time.Time         `json:"created_at"`
	UpdatedAt     time.Time         `json:"updated_at"`
}

type FindXAgent struct {
	ID            string            `json:"id"`
	Ident         string            `json:"ident"`
	TargetID      string            `json:"target_id,omitempty"`
	IP            string            `json:"ip"`
	Hostname      string            `json:"hostname"`
	OS            string            `json:"os,omitempty"`
	Arch          string            `json:"arch,omitempty"`
	Version       string            `json:"version,omitempty"`
	Collector     string            `json:"collector,omitempty"`
	Status        string            `json:"status"`
	Capabilities  []string          `json:"capabilities,omitempty"`
	GlobalLabels  map[string]string `json:"global_labels,omitempty"`
	ConfigVersion string            `json:"config_version,omitempty"`
	LastSeen      time.Time         `json:"last_seen"`
	CreatedAt     time.Time         `json:"created_at"`
	UpdatedAt     time.Time         `json:"updated_at"`
}

type FindXAgentHeartbeat struct {
	Ident         string            `json:"ident"`
	IP            string            `json:"ip"`
	Hostname      string            `json:"hostname"`
	OS            string            `json:"os,omitempty"`
	Arch          string            `json:"arch,omitempty"`
	Version       string            `json:"version,omitempty"`
	Collector     string            `json:"collector,omitempty"`
	Capabilities  []string          `json:"capabilities,omitempty"`
	GlobalLabels  map[string]string `json:"global_labels,omitempty"`
	ConfigVersion string            `json:"config_version,omitempty"`
	UnixTime      int64             `json:"unixtime,omitempty"`
}

type MonitorHealth struct {
	Status      string         `json:"status"`
	Mode        string         `json:"mode"`
	Storage     map[string]any `json:"storage"`
	Targets     int            `json:"targets"`
	Agents      int            `json:"agents"`
	AgentOnline int            `json:"agent_online"`
	GeneratedAt time.Time      `json:"generated_at"`
}
