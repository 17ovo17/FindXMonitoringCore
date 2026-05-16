package store

import (
	"encoding/json"
	"strings"
	"time"

	"ai-workbench-api/internal/model"
)

const findXAgentExecutorBlockedReason = "PENDING: executor not enabled / execution protocol not open"

func normalizeFindXAgentInstallPlan(item model.FindXAgentInstallPlan, now time.Time) model.FindXAgentInstallPlan {
	if item.ID == "" {
		item.ID = NewID()
	}
	item.PackageID = strings.TrimSpace(item.PackageID)
	item.OS = strings.TrimSpace(item.OS)
	item.Method = strings.TrimSpace(item.Method)
	item.TargetIDs = cleanStringList(item.TargetIDs)
	item.Status = findXAgentBlockedStatus
	if strings.TrimSpace(item.Blocker) == "" {
		item.Blocker = findXAgentExecutorBlockedReason
	}
	item.Audit = firstLifecycleValue(item.Audit, "findx_agent.install_plan.requested")
	item.Metadata = sanitizeLifecycleMetadata(item.Metadata)
	if item.CreatedAt.IsZero() {
		item.CreatedAt = now
	}
	item.UpdatedAt = now
	return item
}

func normalizeFindXAgentConfigRollout(item model.FindXAgentConfigRollout, now time.Time) model.FindXAgentConfigRollout {
	if item.ID == "" {
		item.ID = NewID()
	}
	item.TemplateID = strings.TrimSpace(item.TemplateID)
	item.AgentIDs = cleanStringList(item.AgentIDs)
	item.TargetIDs = cleanStringList(item.TargetIDs)
	item.Status = findXAgentBlockedStatus
	if strings.TrimSpace(item.Blocker) == "" {
		item.Blocker = findXAgentExecutorBlockedReason
	}
	item.Audit = firstLifecycleValue(item.Audit, "findx_agent.config_rollout.requested")
	item.Metadata = sanitizeLifecycleMetadata(item.Metadata)
	if item.CreatedAt.IsZero() {
		item.CreatedAt = now
	}
	item.UpdatedAt = now
	return item
}

func normalizeFindXAgentExecutionTask(item model.FindXAgentExecutionTask, now time.Time) model.FindXAgentExecutionTask {
	if item.ID == "" {
		item.ID = NewID()
	}
	item.Action = strings.ToLower(strings.TrimSpace(item.Action))
	item.AgentIDs = cleanStringList(item.AgentIDs)
	item.TargetIDs = cleanStringList(item.TargetIDs)
	item.Status = findXAgentBlockedStatus
	if strings.TrimSpace(item.Blocker) == "" {
		item.Blocker = findXAgentExecutorBlockedReason
	}
	item.Audit = firstLifecycleValue(item.Audit, "findx_agent.task.requested")
	item.Metadata = sanitizeLifecycleMetadata(item.Metadata)
	if item.CreatedAt.IsZero() {
		item.CreatedAt = now
	}
	item.UpdatedAt = now
	return item
}

func lifecycleJSON(value any) string {
	raw, err := json.Marshal(value)
	if err != nil {
		return "{}"
	}
	return string(raw)
}

func firstLifecycleValue(values ...string) string {
	for _, value := range values {
		if clean := strings.TrimSpace(value); clean != "" {
			return clean
		}
	}
	return ""
}

func cleanStringList(values []string) []string {
	out := []string{}
	seen := map[string]bool{}
	for _, value := range values {
		clean := strings.TrimSpace(value)
		if clean != "" && !seen[clean] {
			seen[clean] = true
			out = append(out, clean)
		}
	}
	return out
}

func sanitizeLifecycleMetadata(in map[string]string) map[string]string {
	out := map[string]string{}
	for key, value := range in {
		cleanKey := strings.TrimSpace(key)
		allowedRef := allowedLifecycleReferenceKey(cleanKey)
		if cleanKey == "" || (!allowedRef && isSensitiveKey(value)) {
			continue
		}
		if isSensitiveKey(key) && !allowedLifecycleReferenceKey(cleanKey) {
			continue
		}
		cleanValue := strings.TrimSpace(value)
		if allowedRef {
			cleanValue = sanitizeLifecycleReferenceValue(value)
		}
		if cleanValue == "" {
			continue
		}
		out[cleanKey] = cleanValue
	}
	return out
}

func allowedLifecycleReferenceKey(key string) bool {
	switch strings.TrimSpace(key) {
	case "provider_auth_ref":
		return true
	default:
		return false
	}
}

func sanitizeLifecycleReferenceValue(value string) string {
	clean := strings.TrimSpace(value)
	if clean == "" || isSensitiveReferenceValue(clean) {
		return ""
	}
	const maxLifecycleReferenceLen = 120
	runes := []rune(clean)
	if len(runes) > maxLifecycleReferenceLen {
		clean = string(runes[:maxLifecycleReferenceLen])
	}
	return clean
}

func isSensitiveReferenceValue(value string) bool {
	normalized := strings.NewReplacer("-", "_", " ", "_", ".", "_").Replace(strings.ToLower(value))
	for _, marker := range []string{"password", "passwd", "secret", "token", "cookie", "bearer", "api_key", "apikey", "access_key", "private_key", "session", "dsn"} {
		if strings.Contains(normalized, marker) {
			return true
		}
	}
	return false
}
