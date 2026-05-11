package model

import "encoding/json"

type MonitoringBuiltinComponent struct {
	ID             string   `json:"id"`
	Ident          string   `json:"ident"`
	Name           string   `json:"name"`
	Logo           string   `json:"logo,omitempty"`
	Readme         string   `json:"readme"`
	Disabled       int      `json:"disabled"`
	Tags           []string `json:"tags,omitempty"`
	DashboardCount int      `json:"dashboard_count"`
	CollectCount   int      `json:"collect_count"`
	MetricCount    int      `json:"metric_count"`
	AlertCount     int      `json:"alert_count"`
	RecordCount    int      `json:"record_count,omitempty"`
	UpdatedBy      string   `json:"updated_by"`
}

type MonitoringBuiltinPayload struct {
	ID          string          `json:"id"`
	UUID        string          `json:"uuid,omitempty"`
	ComponentID string          `json:"component_id"`
	Type        string          `json:"type"`
	Cate        string          `json:"cate,omitempty"`
	Name        string          `json:"name"`
	Title       string          `json:"title"`
	Tags        []string        `json:"tags,omitempty"`
	Description string          `json:"description,omitempty"`
	Note        string          `json:"note,omitempty"`
	UpdatedBy   string          `json:"updated_by"`
	Content     json.RawMessage `json:"content"`
}

type MonitoringBuiltinComponentInput struct {
	ID       string `json:"id"`
	Ident    string `json:"ident"`
	Name     string `json:"name"`
	Logo     string `json:"logo"`
	Readme   string `json:"readme"`
	Disabled int    `json:"disabled"`
}

type MonitoringBuiltinPayloadInput struct {
	ID          string          `json:"id"`
	UUID        string          `json:"uuid"`
	Type        string          `json:"type"`
	ComponentID string          `json:"component_id"`
	Cate        string          `json:"cate"`
	Name        string          `json:"name"`
	Content     json.RawMessage `json:"content"`
}

type MonitoringBuiltinPayloadFilter struct {
	ComponentID string
	Type        string
	Query       string
}
