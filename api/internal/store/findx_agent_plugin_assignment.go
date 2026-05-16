package store

import (
	"encoding/json"
	"sort"
	"strings"
	"time"

	"ai-workbench-api/internal/model"

	"github.com/sirupsen/logrus"
)

func SaveFindXAgentPluginAssignment(input model.FindXAgentPluginAssignment) (model.FindXAgentPluginAssignment, error) {
	item := normalizeFindXAgentPluginAssignment(input, time.Now())
	if GormOK() {
		if err := GetDB().Save(&item).Error; err != nil {
			return model.FindXAgentPluginAssignment{}, err
		}
		return copyFindXAgentPluginAssignment(item), nil
	}
	mu.Lock()
	cp := copyFindXAgentPluginAssignment(item)
	findxAgentPluginAssignments[item.ID] = &cp
	mu.Unlock()
	return copyFindXAgentPluginAssignment(item), nil
}

func GetFindXAgentPluginAssignment(id string) (model.FindXAgentPluginAssignment, bool, error) {
	id = strings.TrimSpace(id)
	if id == "" {
		return model.FindXAgentPluginAssignment{}, false, nil
	}
	if GormOK() {
		var row model.FindXAgentPluginAssignment
		if err := GetDB().Where("id = ?", id).First(&row).Error; err == nil {
			return copyFindXAgentPluginAssignment(row), true, nil
		}
	}
	mu.RLock()
	defer mu.RUnlock()
	item, ok := findxAgentPluginAssignments[id]
	if !ok {
		return model.FindXAgentPluginAssignment{}, false, nil
	}
	return copyFindXAgentPluginAssignment(*item), true, nil
}

func FindFindXAgentPluginAssignment(hostRef, agentRef, pluginID string) (model.FindXAgentPluginAssignment, bool, error) {
	hostRef = strings.TrimSpace(hostRef)
	agentRef = strings.TrimSpace(agentRef)
	pluginID = strings.TrimSpace(pluginID)
	if hostRef == "" || agentRef == "" || pluginID == "" {
		return model.FindXAgentPluginAssignment{}, false, nil
	}
	if GormOK() {
		var row model.FindXAgentPluginAssignment
		err := GetDB().
			Where("host_ref = ? AND agent_ref = ? AND plugin_id = ?", hostRef, agentRef, pluginID).
			Order("updated_at desc").
			First(&row).Error
		if err == nil {
			return copyFindXAgentPluginAssignment(row), true, nil
		}
	}
	mu.RLock()
	defer mu.RUnlock()
	var found *model.FindXAgentPluginAssignment
	for _, item := range findxAgentPluginAssignments {
		if item.HostRef != hostRef || item.AgentRef != agentRef || item.PluginID != pluginID {
			continue
		}
		if found == nil || item.UpdatedAt.After(found.UpdatedAt) {
			found = item
		}
	}
	if found == nil {
		return model.FindXAgentPluginAssignment{}, false, nil
	}
	return copyFindXAgentPluginAssignment(*found), true, nil
}

func ListFindXAgentPluginAssignments(hostRef, agentRef string) ([]model.FindXAgentPluginAssignment, error) {
	hostRef = strings.TrimSpace(hostRef)
	agentRef = strings.TrimSpace(agentRef)
	if GormOK() {
		query := GetDB().Model(&model.FindXAgentPluginAssignment{})
		if hostRef != "" {
			query = query.Where("host_ref = ?", hostRef)
		}
		if agentRef != "" {
			query = query.Where("agent_ref = ?", agentRef)
		}
		var rows []model.FindXAgentPluginAssignment
		if err := query.Order("updated_at desc").Limit(500).Find(&rows).Error; err != nil {
			return nil, err
		}
		out := make([]model.FindXAgentPluginAssignment, 0, len(rows))
		for _, row := range rows {
			out = append(out, copyFindXAgentPluginAssignment(row))
		}
		return out, nil
	}
	mu.RLock()
	defer mu.RUnlock()
	out := make([]model.FindXAgentPluginAssignment, 0, len(findxAgentPluginAssignments))
	for _, item := range findxAgentPluginAssignments {
		if hostRef != "" && item.HostRef != hostRef {
			continue
		}
		if agentRef != "" && item.AgentRef != agentRef {
			continue
		}
		out = append(out, copyFindXAgentPluginAssignment(*item))
	}
	sortFindXAgentPluginAssignments(out)
	if len(out) > 500 {
		out = out[:500]
	}
	return out, nil
}

func SaveFindXAgentPluginTargetBinding(input model.FindXAgentPluginTargetBinding) (model.FindXAgentPluginTargetBinding, error) {
	item := normalizeFindXAgentPluginTargetBinding(input, time.Now())
	if GormOK() {
		if err := GetDB().Save(&item).Error; err != nil {
			return model.FindXAgentPluginTargetBinding{}, err
		}
		return item, nil
	}
	mu.Lock()
	cp := item
	findxAgentPluginTargetBindings[item.ID] = &cp
	mu.Unlock()
	return item, nil
}

func GetFindXAgentPluginTargetBinding(id string) (model.FindXAgentPluginTargetBinding, bool, error) {
	id = strings.TrimSpace(id)
	if id == "" {
		return model.FindXAgentPluginTargetBinding{}, false, nil
	}
	if GormOK() {
		var row model.FindXAgentPluginTargetBinding
		if err := GetDB().Where("id = ?", id).First(&row).Error; err == nil {
			return row, true, nil
		}
	}
	mu.RLock()
	defer mu.RUnlock()
	item, ok := findxAgentPluginTargetBindings[id]
	if !ok {
		return model.FindXAgentPluginTargetBinding{}, false, nil
	}
	return *item, true, nil
}

func ListFindXAgentPluginTargetBindings(assignmentID string) ([]model.FindXAgentPluginTargetBinding, error) {
	assignmentID = strings.TrimSpace(assignmentID)
	if assignmentID == "" {
		return nil, nil
	}
	if GormOK() {
		var rows []model.FindXAgentPluginTargetBinding
		err := GetDB().
			Where("assignment_id = ?", assignmentID).
			Order("updated_at desc").
			Find(&rows).Error
		if err != nil {
			return nil, err
		}
		return rows, nil
	}
	mu.RLock()
	defer mu.RUnlock()
	out := make([]model.FindXAgentPluginTargetBinding, 0)
	for _, item := range findxAgentPluginTargetBindings {
		if item.AssignmentID == assignmentID {
			out = append(out, *item)
		}
	}
	sort.Slice(out, func(i, j int) bool { return out[i].UpdatedAt.After(out[j].UpdatedAt) })
	return out, nil
}

func ResetFindXAgentPluginAssignmentsForTest() {
	mu.Lock()
	defer mu.Unlock()
	findxAgentPluginAssignments = map[string]*model.FindXAgentPluginAssignment{}
	findxAgentPluginTargetBindings = map[string]*model.FindXAgentPluginTargetBinding{}
}

func normalizeFindXAgentPluginAssignment(item model.FindXAgentPluginAssignment, now time.Time) model.FindXAgentPluginAssignment {
	if item.ID == "" {
		item.ID = "findx-agent-plugin-assignment-" + NewID()
	}
	item.HostRef = strings.TrimSpace(item.HostRef)
	item.AgentRef = strings.TrimSpace(item.AgentRef)
	item.PluginID = strings.TrimSpace(item.PluginID)
	item.PluginVersion = strings.TrimSpace(item.PluginVersion)
	item.ConfigSnippetRef = strings.TrimSpace(item.ConfigSnippetRef)
	item.ConfigFormat = strings.TrimSpace(item.ConfigFormat)
	item.ProviderMode = strings.TrimSpace(item.ProviderMode)
	item.SourceRolloutID = strings.TrimSpace(item.SourceRolloutID)
	item.TargetBindingRef = strings.TrimSpace(item.TargetBindingRef)
	item.AuditRef = strings.TrimSpace(item.AuditRef)
	item.AssignmentContract = firstNonEmpty(item.AssignmentContract, "cmdb.agent.plugin.assignment.v1")
	item.Status = findXAgentBlockedStatus
	if strings.TrimSpace(item.Blocker) == "" {
		item.Blocker = "PENDING: plugin assignment saved; remote dispatch receipts are still closed"
	}
	if item.CreatedAt.IsZero() {
		item.CreatedAt = now
	}
	item.UpdatedAt = now
	item.DashboardRefsJSON = sanitizeAssignmentJSONList(item.DashboardRefsJSON)
	item.MissingContractsJSON = sanitizeAssignmentJSONList(item.MissingContractsJSON)
	item.MetadataJSON = sanitizeAssignmentJSONObject(item.MetadataJSON)
	return item
}

func normalizeFindXAgentPluginTargetBinding(item model.FindXAgentPluginTargetBinding, now time.Time) model.FindXAgentPluginTargetBinding {
	if item.ID == "" {
		item.ID = "findx-agent-plugin-target-binding-" + NewID()
	}
	item.AssignmentID = strings.TrimSpace(item.AssignmentID)
	item.HostRef = strings.TrimSpace(item.HostRef)
	item.TargetID = strings.TrimSpace(item.TargetID)
	item.AgentRef = strings.TrimSpace(item.AgentRef)
	item.PluginID = strings.TrimSpace(item.PluginID)
	item.BindingType = firstNonEmpty(item.BindingType, "cmdb_host_plugin")
	item.SourceRolloutID = strings.TrimSpace(item.SourceRolloutID)
	item.ContractID = firstNonEmpty(item.ContractID, "cmdb.agent.plugin.target_binding.v1")
	item.Status = findXAgentBlockedStatus
	if strings.TrimSpace(item.Blocker) == "" {
		item.Blocker = "PENDING: target binding saved; remote dispatch receipts are still closed"
	}
	item.AuditRef = strings.TrimSpace(item.AuditRef)
	if item.CreatedAt.IsZero() {
		item.CreatedAt = now
	}
	item.UpdatedAt = now
	return item
}

func copyFindXAgentPluginAssignment(item model.FindXAgentPluginAssignment) model.FindXAgentPluginAssignment {
	return item
}

func sortFindXAgentPluginAssignments(items []model.FindXAgentPluginAssignment) {
	sort.Slice(items, func(i, j int) bool { return items[i].UpdatedAt.After(items[j].UpdatedAt) })
}

func sanitizeAssignmentJSONList(raw string) string {
	var values []string
	if err := json.Unmarshal([]byte(strings.TrimSpace(raw)), &values); err != nil {
		return "[]"
	}
	clean := cleanStringList(values)
	out, err := json.Marshal(clean)
	if err != nil {
		logrus.WithError(err).Warn("plugin assignment list json marshal failed")
		return "[]"
	}
	return string(out)
}

func sanitizeAssignmentJSONObject(raw string) string {
	var values map[string]string
	if err := json.Unmarshal([]byte(strings.TrimSpace(raw)), &values); err != nil {
		return "{}"
	}
	out, err := json.Marshal(sanitizeLifecycleMetadata(values))
	if err != nil {
		logrus.WithError(err).Warn("plugin assignment metadata json marshal failed")
		return "{}"
	}
	return string(out)
}
