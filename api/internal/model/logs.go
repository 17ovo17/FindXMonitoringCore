package model

import (
	"encoding/json"
	"time"
)

const (
	LogsContractBlocked  = "PENDING"
	LogsSourcePage       = "logs"
	LogsSourceFindXAudit = "findx_audit"
	LogsSourceOtel       = "otel"
)

type LogQueryRequest struct {
	Source       string
	Query        string
	Page         int
	Limit        int
	Status       string
	Action       string
	ResourceType string
	ResourceID   string
	TraceID      string
	Scope        string
}

type LogRecord struct {
	ID             string         `json:"id"`
	Timestamp      time.Time      `json:"timestamp"`
	Source         string         `json:"source"`
	SourceName     string         `json:"source_name"`
	SeverityText   string         `json:"severity_text"`
	SeverityNumber int            `json:"severity_number"`
	ServiceName    string         `json:"service_name"`
	Body           string         `json:"body"`
	TraceID        string         `json:"trace_id,omitempty"`
	Attributes     map[string]any `json:"attributes"`
}

type LogQueryResponse struct {
	Source     string         `json:"source"`
	SourceName string         `json:"source_name"`
	Status     string         `json:"status"`
	Items      []LogRecord    `json:"items"`
	Total      int            `json:"total"`
	Page       int            `json:"page"`
	Limit      int            `json:"limit"`
	Meta       map[string]any `json:"meta"`
	Blocker    string         `json:"blocker,omitempty"`
}

type LogAggregateRequest struct {
	Source       string
	GroupBy      string
	Page         int
	Limit        int
	Status       string
	Action       string
	ResourceType string
	Scope        string
}

type LogAggregateBucket struct {
	Key   string `json:"key"`
	Label string `json:"label"`
	Count int    `json:"count"`
}

type LogAggregateResponse struct {
	Source     string               `json:"source"`
	SourceName string               `json:"source_name"`
	Status     string               `json:"status"`
	GroupBy    string               `json:"group_by"`
	Buckets    []LogAggregateBucket `json:"buckets"`
	Total      int                  `json:"total"`
	Meta       map[string]any       `json:"meta"`
	Blocker    string               `json:"blocker,omitempty"`
}

type LogContextRequest struct {
	Source  string
	LogID   string
	TraceID string
	Scope   string
	Before  int
	After   int
}

type LogContextResponse struct {
	Source     string         `json:"source"`
	SourceName string         `json:"source_name"`
	Status     string         `json:"status"`
	Center     *LogRecord     `json:"center,omitempty"`
	Before     []LogRecord    `json:"before"`
	After      []LogRecord    `json:"after"`
	Items      []LogRecord    `json:"items"`
	Total      int            `json:"total"`
	Meta       map[string]any `json:"meta"`
	Blocker    string         `json:"blocker,omitempty"`
}

type LogField struct {
	Key         string   `json:"key"`
	Type        string   `json:"type"`
	Category    string   `json:"category"`
	Description string   `json:"description"`
	Examples    []string `json:"examples,omitempty"`
	Indexed     bool     `json:"indexed"`
}

type LogFieldCategory struct {
	Key         string     `json:"key"`
	Name        string     `json:"name"`
	Description string     `json:"description"`
	Fields      []LogField `json:"fields"`
}

type LogLiveDiscoveryState struct {
	Status  string `json:"status"`
	Blocker string `json:"blocker"`
}

type LogCapabilityState struct {
	Status           string   `json:"status"`
	ContractGapID    string   `json:"contract_gap_id,omitempty"`
	MissingContracts []string `json:"missing_contracts,omitempty"`
	SafeToRetry      bool     `json:"safe_to_retry"`
}

type LogFieldsResponse struct {
	Status        string                        `json:"status"`
	Categories    []LogFieldCategory            `json:"categories"`
	Fields        []LogField                    `json:"fields"`
	LiveDiscovery LogLiveDiscoveryState         `json:"live_discovery"`
	Capabilities  map[string]LogCapabilityState `json:"capabilities"`
}

type LogsBlockedEnvelope struct {
	Code             string   `json:"code"`
	Status           string   `json:"status"`
	ContractID       string   `json:"contract_id,omitempty"`
	ContractGapID    string   `json:"contract_gap_id,omitempty"`
	MissingContracts []string `json:"missing_contracts"`
	SafeToRetry      bool     `json:"safe_to_retry"`
	Error            string   `json:"error"`
	Blocker          string   `json:"blocker"`
}

type LogPipeline struct {
	ID          string          `json:"id"`
	Name        string          `json:"name"`
	Version     string          `json:"version"`
	Description string          `json:"description,omitempty"`
	Enabled     bool            `json:"enabled"`
	Stages      json.RawMessage `json:"stages"`
	Config      json.RawMessage `json:"config"`
	CreatedBy   string          `json:"created_by,omitempty"`
	UpdatedBy   string          `json:"updated_by,omitempty"`
	CreatedAt   time.Time       `json:"created_at"`
	UpdatedAt   time.Time       `json:"updated_at"`
}

type LogPipelineInput struct {
	ID          string          `json:"id,omitempty"`
	Name        string          `json:"name"`
	Version     string          `json:"version,omitempty"`
	Description string          `json:"description,omitempty"`
	Enabled     *bool           `json:"enabled,omitempty"`
	Stages      json.RawMessage `json:"stages,omitempty"`
	Config      json.RawMessage `json:"config,omitempty"`
}

type LogPipelinePreviewRequest struct {
	Pipeline LogPipelineInput `json:"pipeline"`
	Sample   string           `json:"sample,omitempty"`
	Samples  []string         `json:"samples,omitempty"`
	Parser   string           `json:"parser,omitempty"`
	Pattern  string           `json:"pattern,omitempty"`
}

type LogPipelinePreviewResult struct {
	OK              bool              `json:"ok"`
	Status          string            `json:"status"`
	Parser          string            `json:"parser"`
	SampleCount     int               `json:"sample_count"`
	ExtractedFields []LogPreviewField `json:"extracted_fields"`
	Warnings        []string          `json:"warnings,omitempty"`
}

type LogPreviewField struct {
	Key   string `json:"key"`
	Type  string `json:"type"`
	Value string `json:"value,omitempty"`
}

type ExplorerSavedView struct {
	ID          string          `json:"id"`
	SourcePage  string          `json:"sourcePage"`
	Name        string          `json:"name"`
	Description string          `json:"description,omitempty"`
	Query       json.RawMessage `json:"query"`
	Filters     json.RawMessage `json:"filters"`
	Columns     json.RawMessage `json:"columns"`
	TimeRange   json.RawMessage `json:"timeRange"`
	Layout      json.RawMessage `json:"layout"`
	CreatedBy   string          `json:"created_by,omitempty"`
	UpdatedBy   string          `json:"updated_by,omitempty"`
	CreatedAt   time.Time       `json:"created_at"`
	UpdatedAt   time.Time       `json:"updated_at"`
}

type ExplorerSavedViewInput struct {
	ID          string          `json:"id,omitempty"`
	SourcePage  string          `json:"sourcePage"`
	Name        string          `json:"name"`
	Description string          `json:"description,omitempty"`
	Query       json.RawMessage `json:"query,omitempty"`
	Filters     json.RawMessage `json:"filters,omitempty"`
	Columns     json.RawMessage `json:"columns,omitempty"`
	TimeRange   json.RawMessage `json:"timeRange,omitempty"`
	Layout      json.RawMessage `json:"layout,omitempty"`
}
