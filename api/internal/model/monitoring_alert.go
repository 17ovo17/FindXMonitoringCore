package model

import (
	"crypto/sha256"
	"fmt"
	"sort"
	"strings"
	"time"
)

const (
	MonitorAlertSeverityCritical = "critical"
	MonitorAlertSeverityWarning  = "warning"
	MonitorAlertSeverityInfo     = "info"
	MonitorAlertSeverityP0       = "p0"
	MonitorAlertSeverityP1       = "p1"
	MonitorAlertSeverityP2       = "p2"
	MonitorAlertSeverityP3       = "p3"

	MonitorAlertRuleStatusActive   = "active"
	MonitorAlertRuleStatusDisabled = "disabled"

	MonitorAlertEventStatusFiring       = "firing"
	MonitorAlertEventStatusAcknowledged = "acknowledged"
	MonitorAlertEventStatusAssigned     = "assigned"
	MonitorAlertEventStatusMuted        = "muted"
	MonitorAlertEventStatusResolved     = "resolved"
	MonitorAlertEventStatusArchived     = "archived"

	MonitorAlertEventActionAck     = "ack"
	MonitorAlertEventActionAssign  = "assign"
	MonitorAlertEventActionMute    = "mute"
	MonitorAlertEventActionResolve = "resolve"
	MonitorAlertEventActionArchive = "archive"

	MonitorNoDataPolicyKeepState = "keep_state"
	MonitorNoDataPolicyAlerting  = "alerting"
	MonitorNoDataPolicyOK        = "ok"
)

var monitorFingerprintSensitiveKeys = []string{
	"api_key", "apikey", "auth", "authorization", "cookie", "dsn", "password", "private", "secret", "token",
}

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

func GenerateMonitorAlertEventFingerprint(event *MonitorAlertEvent) string {
	if event == nil {
		return ""
	}
	// P1 新平台基座语义：current event 幂等键只来自规范字段，外部 fingerprint 与敏感 label 不参与计算。
	parts := []string{
		"rule_id=" + strings.TrimSpace(event.RuleID),
		"datasource_id=" + strings.TrimSpace(event.DatasourceID),
		"event_key=" + strings.TrimSpace(event.EventKey),
		"target_id=" + strings.TrimSpace(event.TargetID),
		"target_ident=" + strings.TrimSpace(event.TargetIdent),
		"labels=" + canonicalAlertLabels(event.Labels),
	}
	sum := sha256.Sum256([]byte(strings.Join(parts, "\n")))
	return fmt.Sprintf("%x", sum)
}

func ValidateMonitorAlertEventTransition(status, action string) (string, error) {
	status = strings.TrimSpace(status)
	if status == "" {
		status = MonitorAlertEventStatusFiring
	}
	action = strings.TrimSpace(action)
	if !IsKnownMonitorAlertEventStatus(status) {
		return "", fmt.Errorf("invalid alert event status: %s", status)
	}
	if IsTerminalMonitorAlertEventStatus(status) {
		return "", fmt.Errorf("terminal alert event cannot be changed")
	}
	switch action {
	case MonitorAlertEventActionAck:
		return MonitorAlertEventStatusAcknowledged, nil
	case MonitorAlertEventActionAssign:
		return MonitorAlertEventStatusAssigned, nil
	case MonitorAlertEventActionMute:
		return MonitorAlertEventStatusMuted, nil
	case MonitorAlertEventActionResolve:
		return MonitorAlertEventStatusResolved, nil
	case MonitorAlertEventActionArchive:
		return MonitorAlertEventStatusArchived, nil
	default:
		return "", fmt.Errorf("invalid alert event action: %s", action)
	}
}

func IsKnownMonitorAlertEventStatus(status string) bool {
	switch strings.TrimSpace(status) {
	case MonitorAlertEventStatusFiring, MonitorAlertEventStatusAcknowledged,
		MonitorAlertEventStatusAssigned, MonitorAlertEventStatusMuted,
		MonitorAlertEventStatusResolved, MonitorAlertEventStatusArchived:
		return true
	default:
		return false
	}
}

func IsTerminalMonitorAlertEventStatus(status string) bool {
	switch strings.TrimSpace(status) {
	case MonitorAlertEventStatusResolved, MonitorAlertEventStatusArchived:
		return true
	default:
		return false
	}
}

func canonicalAlertLabels(labels map[string]string) string {
	keys := make([]string, 0, len(labels))
	for key := range labels {
		if !isMonitorFingerprintSensitiveKey(key) {
			keys = append(keys, key)
		}
	}
	sort.Strings(keys)
	parts := make([]string, 0, len(keys))
	for _, key := range keys {
		parts = append(parts, strings.TrimSpace(key)+"="+strings.TrimSpace(labels[key]))
	}
	return strings.Join(parts, "\n")
}

func isMonitorFingerprintSensitiveKey(key string) bool {
	key = strings.ToLower(strings.TrimSpace(key))
	for _, sensitive := range monitorFingerprintSensitiveKeys {
		if strings.Contains(key, sensitive) {
			return true
		}
	}
	return false
}
