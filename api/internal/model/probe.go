package model

import "time"

const (
	ProbeCheckTypeHTTP = "http"
	ProbeCheckTypeTCP  = "tcp"
	ProbeCheckTypePing = "ping"
	ProbeCheckTypeDNS  = "dns"

	ProbeStatusUnknown  = "unknown"
	ProbeStatusUp       = "up"
	ProbeStatusDown     = "down"
	ProbeStatusDisabled = "disabled"
	ProbeStatusDegraded = "degraded"
	ProbeStatusNoData   = "no_data"

	ProbeIncidentStatusInvestigating = "investigating"
	ProbeIncidentStatusIdentified    = "identified"
	ProbeIncidentStatusMonitoring    = "monitoring"
	ProbeIncidentStatusResolved      = "resolved"

	ProbeContractBlockedCode = "BLOCKED_BY_CONTRACT"
)

type ProbeHTTPConfig struct {
	Method        string            `json:"method,omitempty"`
	ExpectedCodes []int             `json:"expected_codes,omitempty"`
	Headers       map[string]string `json:"headers,omitempty"`
	BodyContains  string            `json:"body_contains,omitempty"`
}

type ProbeDNSConfig struct {
	RecordType    string   `json:"record_type,omitempty"`
	ExpectedValue []string `json:"expected_value,omitempty"`
}

type ProbeCheck struct {
	ID              string                 `gorm:"primaryKey;size:64" json:"id"`
	Name            string                 `gorm:"size:160;not null" json:"name"`
	Type            string                 `gorm:"size:16;index" json:"type"`
	URL             string                 `gorm:"size:1024" json:"url,omitempty"`
	Target          string                 `gorm:"size:512" json:"target,omitempty"`
	Port            int                    `json:"port,omitempty"`
	HTTPConfig      ProbeHTTPConfig        `gorm:"serializer:json" json:"http_config,omitempty"`
	DNSConfig       ProbeDNSConfig         `gorm:"serializer:json" json:"dns_config,omitempty"`
	IntervalSeconds int                    `json:"interval_seconds"`
	TimeoutSeconds  int                    `json:"timeout_seconds"`
	Retries         int                    `json:"retries"`
	Status          string                 `gorm:"size:32;index" json:"status"`
	Enabled         bool                   `gorm:"index" json:"enabled"`
	BusinessGroup   string                 `gorm:"size:160;index" json:"business_group,omitempty"`
	Labels          map[string]string      `gorm:"serializer:json" json:"labels,omitempty"`
	LastResult      *ProbeCheckResult      `gorm:"-" json:"last_result,omitempty"`
	Uptime90d       *float64               `gorm:"-" json:"uptime_90d,omitempty"`
	ResponseTimeMs  *int                   `gorm:"-" json:"response_time_ms,omitempty"`
	StatusBar       []ProbeStatusBarBucket `gorm:"-" json:"status_bar,omitempty"`
	CreatedAt       time.Time              `json:"created_at"`
	UpdatedAt       time.Time              `json:"updated_at"`
}

type ProbeCheckResult struct {
	ID             string            `gorm:"primaryKey;size:64" json:"id"`
	CheckID        string            `gorm:"size:64;index" json:"check_id"`
	Status         string            `gorm:"size:32;index" json:"status"`
	ResponseTimeMs int               `json:"response_time_ms,omitempty"`
	StatusCode     int               `json:"status_code,omitempty"`
	Error          string            `gorm:"type:text" json:"error,omitempty"`
	Region         string            `gorm:"size:80" json:"region,omitempty"`
	EvidenceRef    string            `gorm:"size:512" json:"evidence_ref,omitempty"`
	Metadata       map[string]string `gorm:"serializer:json" json:"metadata,omitempty"`
	CheckedAt      time.Time         `gorm:"index" json:"checked_at"`
	CreatedAt      time.Time         `json:"created_at"`
}

type ProbeStatusBarBucket struct {
	Date           string   `json:"date"`
	Status         string   `json:"status"`
	Reason         string   `json:"reason,omitempty"`
	UptimePercent  *float64 `json:"uptime_percent,omitempty"`
	ResponseTimeMs *int     `json:"response_time_ms,omitempty"`
	EvidenceRef    string   `json:"evidence_ref,omitempty"`
}

type ProbeStatusGroup struct {
	ID       string   `json:"id"`
	Name     string   `json:"name"`
	CheckIDs []string `json:"check_ids"`
}

type ProbeStatusPage struct {
	ID            string             `gorm:"primaryKey;size:64" json:"id"`
	Slug          string             `gorm:"size:120;uniqueIndex;not null" json:"slug"`
	Title         string             `gorm:"size:160;not null" json:"title"`
	Description   string             `gorm:"size:512" json:"description,omitempty"`
	BusinessGroup string             `gorm:"size:160;index" json:"business_group,omitempty"`
	Visibility    string             `gorm:"size:32" json:"visibility"`
	Groups        []ProbeStatusGroup `gorm:"serializer:json;column:status_groups" json:"groups"`
	Labels        map[string]string  `gorm:"serializer:json" json:"labels,omitempty"`
	CreatedAt     time.Time          `json:"created_at"`
	UpdatedAt     time.Time          `json:"updated_at"`
}

type ProbeIncident struct {
	ID            string            `gorm:"primaryKey;size:64" json:"id"`
	CheckID       string            `gorm:"size:64;index" json:"check_id,omitempty"`
	StatusPageID  string            `gorm:"size:64;index" json:"status_page_id,omitempty"`
	Title         string            `gorm:"size:180;not null" json:"title"`
	Status        string            `gorm:"size:32;index" json:"status"`
	Severity      string            `gorm:"size:32" json:"severity,omitempty"`
	Message       string            `gorm:"type:text" json:"message,omitempty"`
	BusinessGroup string            `gorm:"size:160;index" json:"business_group,omitempty"`
	Labels        map[string]string `gorm:"serializer:json" json:"labels,omitempty"`
	StartedAt     time.Time         `json:"started_at"`
	ResolvedAt    *time.Time        `json:"resolved_at,omitempty"`
	CreatedAt     time.Time         `json:"created_at"`
	UpdatedAt     time.Time         `json:"updated_at"`
}

type ProbeSubscription struct {
	ID           string     `gorm:"primaryKey;size:64" json:"id"`
	StatusPageID string     `gorm:"size:64;index" json:"status_page_id"`
	Receiver     string     `gorm:"size:180" json:"receiver"`
	ChannelType  string     `gorm:"size:32" json:"channel_type"`
	Enabled      bool       `json:"enabled"`
	VerifiedAt   *time.Time `json:"verified_at,omitempty"`
	CreatedAt    time.Time  `json:"created_at"`
	UpdatedAt    time.Time  `json:"updated_at"`
}

type ProbeNotificationBinding struct {
	ID             string            `gorm:"primaryKey;size:64" json:"id"`
	CheckID        string            `gorm:"size:64;index" json:"check_id,omitempty"`
	StatusPageID   string            `gorm:"size:64;index" json:"status_page_id,omitempty"`
	ChannelID      string            `gorm:"size:64;index" json:"channel_id"`
	Enabled        bool              `json:"enabled"`
	ReceiptMode    string            `gorm:"size:64" json:"receipt_mode,omitempty"`
	LastReceiptRef string            `gorm:"size:512" json:"last_receipt_ref,omitempty"`
	Labels         map[string]string `gorm:"serializer:json" json:"labels,omitempty"`
	CreatedAt      time.Time         `json:"created_at"`
	UpdatedAt      time.Time         `json:"updated_at"`
}

type ProbeAlertBinding struct {
	ID              string            `gorm:"primaryKey;size:64" json:"id"`
	CheckID         string            `gorm:"size:64;index" json:"check_id"`
	AlertRuleID     string            `gorm:"size:64;index" json:"alert_rule_id"`
	Enabled         bool              `json:"enabled"`
	LastEvidenceRef string            `gorm:"size:512" json:"last_evidence_ref,omitempty"`
	Labels          map[string]string `gorm:"serializer:json" json:"labels,omitempty"`
	CreatedAt       time.Time         `json:"created_at"`
	UpdatedAt       time.Time         `json:"updated_at"`
}

type ProbeStatusPageView struct {
	Status         string                 `json:"status"`
	StatusReason   string                 `json:"status_reason,omitempty"`
	UpdatedAt      time.Time              `json:"updated_at"`
	Page           *ProbeStatusPage       `json:"page,omitempty"`
	Summary        ProbeStatusSummary     `json:"summary"`
	Groups         []ProbeStatusGroupView `json:"groups"`
	Incidents      []ProbeIncident        `json:"incidents"`
	Capabilities   []ProbeCapability      `json:"capabilities,omitempty"`
	ContractMatrix []ProbeCapability      `json:"contract_matrix,omitempty"`
	Meta           map[string]any         `json:"meta,omitempty"`
}

type ProbeStatusGroupView struct {
	ID     string       `json:"id"`
	Name   string       `json:"name"`
	Checks []ProbeCheck `json:"checks"`
}

type ProbeStatusSummary struct {
	Uptime90d              *float64 `json:"uptime_90d,omitempty"`
	RunningChecks          int      `json:"running_checks"`
	TotalChecks            int      `json:"total_checks"`
	AverageResponse30dMs   *int     `json:"average_response_30d_ms,omitempty"`
	IncidentCount90d       int      `json:"incident_count_90d"`
	HasRunEvidence         bool     `json:"has_run_evidence"`
	MissingRunEvidenceNote string   `json:"missing_run_evidence_note,omitempty"`
}

type ProbeCapability struct {
	ID               string   `json:"id"`
	Capability       string   `json:"capability"`
	Domain           string   `json:"domain"`
	Status           string   `json:"status"`
	ContractGapID    string   `json:"contract_gap_id,omitempty"`
	MissingContracts []string `json:"missing_contracts,omitempty"`
	Message          string   `json:"message,omitempty"`
	SafeToRetry      bool     `json:"safe_to_retry"`
}

type ProbeBlockedResponse struct {
	Code             string            `json:"code"`
	Message          string            `json:"message"`
	ContractGapID    string            `json:"contract_gap_id"`
	Status           string            `json:"status"`
	SafeToRetry      bool              `json:"safe_to_retry"`
	MissingContracts []string          `json:"missing_contracts"`
	Capability       ProbeCapability   `json:"capability"`
	ContractMatrix   []ProbeCapability `json:"contract_matrix"`
	Meta             map[string]any    `json:"meta,omitempty"`
}
