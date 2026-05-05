package model

import (
	"sort"
	"strings"
)

const (
	MonitorPermissionVersion = "monitor-permission-matrix/v1"
	MonitorPermissionMode    = "role_based"

	MonitorPermissionReasonRoleAllowed = "role_allowed"
	MonitorPermissionReasonRoleDenied  = "role_denied"
	MonitorPermissionReasonUnknown     = "unknown_permission"
)

type MonitorPermissionUser struct {
	ID       string `json:"id"`
	Username string `json:"username"`
	Role     string `json:"role"`
}

type MonitorPermissionScope struct {
	TeamIDs        []string `json:"team_ids"`
	BusinessGroups []string `json:"business_groups"`
	TenantID       string   `json:"tenant_id"`
}

type MonitorPermission struct {
	Resource string `json:"resource"`
	Action   string `json:"action"`
	Key      string `json:"key"`
}

type MonitorPermissionCheckRequest struct {
	Resource   string         `json:"resource"`
	Action     string         `json:"action"`
	ResourceID string         `json:"resource_id,omitempty"`
	Scope      map[string]any `json:"scope,omitempty"`
	TraceID    string         `json:"trace_id,omitempty"`
}

type MonitorPermissionCheckResult struct {
	Known      bool   `json:"known"`
	Allowed    bool   `json:"allowed"`
	Reason     string `json:"reason"`
	Resource   string `json:"resource"`
	Action     string `json:"action"`
	ResourceID string `json:"resource_id,omitempty"`
	TraceID    string `json:"trace_id,omitempty"`
}

type MonitorPermissionMeResponse struct {
	Version     string                     `json:"version"`
	Mode        string                     `json:"mode"`
	User        MonitorPermissionUser      `json:"user"`
	Scope       MonitorPermissionScope     `json:"scope"`
	Permissions []MonitorPermission        `json:"permissions"`
	Matrix      map[string]map[string]bool `json:"matrix"`
}

type MonitorPermissionChecker struct {
	known map[string]map[string]bool
}

func NewMonitorPermissionChecker() MonitorPermissionChecker {
	return MonitorPermissionChecker{known: monitorPermissionMatrix()}
}

func (c MonitorPermissionChecker) Check(role, resource, action string) MonitorPermissionCheckResult {
	resource = strings.TrimSpace(resource)
	action = strings.TrimSpace(action)
	actions, ok := c.known[resource]
	if !ok || !actions[action] {
		return monitorPermissionResult(false, false, resource, action, MonitorPermissionReasonUnknown)
	}
	if strings.EqualFold(strings.TrimSpace(role), "admin") || monitorUserPermissionAllowed(resource, action) {
		return monitorPermissionResult(true, true, resource, action, MonitorPermissionReasonRoleAllowed)
	}
	return monitorPermissionResult(true, false, resource, action, MonitorPermissionReasonRoleDenied)
}

func (c MonitorPermissionChecker) PermissionsForRole(role string) []MonitorPermission {
	out := []MonitorPermission{}
	for resource, actions := range c.known {
		for action := range actions {
			if c.Check(role, resource, action).Allowed {
				out = append(out, MonitorPermission{Resource: resource, Action: action, Key: resource + ":" + action})
			}
		}
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Key < out[j].Key })
	return out
}

func (c MonitorPermissionChecker) MatrixForRole(role string) map[string]map[string]bool {
	out := map[string]map[string]bool{}
	for resource, actions := range c.known {
		out[resource] = map[string]bool{}
		for action := range actions {
			out[resource][action] = c.Check(role, resource, action).Allowed
		}
	}
	return out
}

func monitorPermissionResult(known, allowed bool, resource, action, reason string) MonitorPermissionCheckResult {
	return MonitorPermissionCheckResult{Known: known, Allowed: allowed, Reason: reason, Resource: resource, Action: action}
}

func monitorUserPermissionAllowed(resource, action string) bool {
	switch resource {
	case "monitor.health", "monitor.datasource", "monitor.metric", "monitor.label",
		"monitor.audit_log", "monitor.dashboard", "findx_agent":
		return action == "read"
	case "monitor.query":
		return action == "execute" || action == "execute_range"
	case "monitor.target", "monitor.alert_rule", "monitor.alert_event":
		return action == "read"
	default:
		return false
	}
}

func monitorPermissionMatrix() map[string]map[string]bool {
	return map[string]map[string]bool{
		"monitor.health":      {"read": true},
		"monitor.datasource":  {"read": true},
		"monitor.query":       {"execute": true, "execute_range": true},
		"monitor.metric":      {"read": true},
		"monitor.label":       {"read": true},
		"monitor.target":      {"read": true, "create": true, "update": true, "delete": true},
		"monitor.dashboard":   {"read": true, "create": true, "update": true, "delete": true, "clone": true, "share": true},
		"monitor.alert_rule":  {"read": true, "create": true, "update": true, "delete": true, "enable": true, "disable": true, "clone": true, "tryrun": true, "rollback": true},
		"monitor.alert_event": {"read": true, "ack": true, "assign": true, "resolve": true, "archive": true},
		"monitor.audit_log":   {"read": true},
		"findx_agent":         {"read": true},
	}
}
