package model

import "time"

type MonitorAlertRule struct {
	ID             string            `json:"id"`
	Name           string            `json:"name"`
	Query          string            `json:"query"`
	Severity       string            `json:"severity"`
	DatasourceID   string            `json:"datasource_id"`
	TargetSelector map[string]string `json:"target_selector,omitempty"`
	Labels         map[string]string `json:"labels,omitempty"`
	Annotations    map[string]string `json:"annotations,omitempty"`
	Enabled        bool              `json:"enabled"`
	Version        int               `json:"version"`
	ForDuration    string            `json:"for_duration"`
	NoDataPolicy   string            `json:"no_data_policy"`
	Status         string            `json:"status"`
	CreatedBy      string            `json:"created_by,omitempty"`
	UpdatedBy      string            `json:"updated_by,omitempty"`
	CreatedAt      time.Time         `json:"created_at"`
	UpdatedAt      time.Time         `json:"updated_at"`
}

type MonitorAlertRuleVersion struct {
	ID             string            `json:"id"`
	RuleID         string            `json:"rule_id"`
	Version        int               `json:"version"`
	Name           string            `json:"name"`
	Query          string            `json:"query"`
	Severity       string            `json:"severity"`
	DatasourceID   string            `json:"datasource_id"`
	TargetSelector map[string]string `json:"target_selector,omitempty"`
	Labels         map[string]string `json:"labels,omitempty"`
	Annotations    map[string]string `json:"annotations,omitempty"`
	Enabled        bool              `json:"enabled"`
	ForDuration    string            `json:"for_duration"`
	NoDataPolicy   string            `json:"no_data_policy"`
	Status         string            `json:"status"`
	CreatedBy      string            `json:"created_by,omitempty"`
	CreatedAt      time.Time         `json:"created_at"`
}

type MonitorAlertEvalLog struct {
	ID           string         `json:"id"`
	RuleID       string         `json:"rule_id"`
	RuleVersion  int            `json:"rule_version"`
	Status       string         `json:"status"`
	Message      string         `json:"message,omitempty"`
	Details      map[string]any `json:"details,omitempty"`
	StartedAt    time.Time      `json:"started_at"`
	FinishedAt   time.Time      `json:"finished_at"`
	DurationMs   int64          `json:"duration_ms"`
	DatasourceID string         `json:"datasource_id"`
	QueryHash    string         `json:"query_hash,omitempty"`
}

type MonitorAlertEvent struct {
	ID           string               `json:"id"`
	RuleID       string               `json:"rule_id,omitempty"`
	RuleVersion  int                  `json:"rule_version,omitempty"`
	EventKey     string               `json:"event_key"`
	Name         string               `json:"name"`
	Severity     string               `json:"severity"`
	Status       string               `json:"status"`
	DatasourceID string               `json:"datasource_id,omitempty"`
	TargetID     string               `json:"target_id,omitempty"`
	TargetIdent  string               `json:"target_ident,omitempty"`
	Labels       map[string]string    `json:"labels,omitempty"`
	Annotations  map[string]string    `json:"annotations,omitempty"`
	Value        string               `json:"value,omitempty"`
	Fingerprint  string               `json:"fingerprint"`
	Count        int                  `json:"count"`
	FirstSeen    time.Time            `json:"first_seen"`
	LastSeen     time.Time            `json:"last_seen"`
	AckBy        string               `json:"ack_by,omitempty"`
	Assignee     string               `json:"assignee,omitempty"`
	Resolution   string               `json:"resolution,omitempty"`
	ArchivedAt   *time.Time           `json:"archived_at,omitempty"`
	ResolvedAt   *time.Time           `json:"resolved_at,omitempty"`
	ActionLog    []MonitorAlertAction `json:"action_log,omitempty"`
	CreatedAt    time.Time            `json:"created_at"`
	UpdatedAt    time.Time            `json:"updated_at"`
}

type MonitorAlertAction struct {
	ID        string    `json:"id"`
	EventID   string    `json:"event_id"`
	Action    string    `json:"action"`
	Actor     string    `json:"actor,omitempty"`
	From      string    `json:"from,omitempty"`
	To        string    `json:"to,omitempty"`
	Reason    string    `json:"reason,omitempty"`
	Assignee  string    `json:"assignee,omitempty"`
	TraceID   string    `json:"trace_id,omitempty"`
	CreatedAt time.Time `json:"created_at"`
}

type MonitorAlertTryRunResult struct {
	OK      bool                `json:"ok"`
	Status  string              `json:"status"`
	Checks  []MonitorTryCheck   `json:"checks"`
	Rule    *MonitorAlertRule   `json:"rule,omitempty"`
	EvalLog MonitorAlertEvalLog `json:"eval_log"`
}

type MonitorTryCheck struct {
	Name    string `json:"name"`
	Status  string `json:"status"`
	Message string `json:"message"`
}
