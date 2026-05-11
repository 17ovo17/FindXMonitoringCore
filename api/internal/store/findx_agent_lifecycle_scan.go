package store

import (
	"database/sql"
	"encoding/json"
	"sort"
	"strings"

	"ai-workbench-api/internal/model"
)

func scanFindXAgentInstallPlans(rows *sql.Rows) ([]model.FindXAgentInstallPlan, error) {
	out := []model.FindXAgentInstallPlan{}
	for rows.Next() {
		var item model.FindXAgentInstallPlan
		var targetIDs, evidenceRefs, metadata string
		if err := rows.Scan(&item.ID, &item.PackageID, &item.OS, &item.Method, &targetIDs, &item.CredentialRefPresent, &item.Status, &item.Blocker, &item.Audit, &evidenceRefs, &metadata, &item.CreatedAt, &item.UpdatedAt); err != nil {
			return out, err
		}
		if err := decodeLifecycleJSON(targetIDs, &item.TargetIDs); err != nil {
			return out, err
		}
		if err := decodeLifecycleJSON(evidenceRefs, &item.EvidenceRefs); err != nil {
			return out, err
		}
		if err := decodeLifecycleJSON(metadata, &item.Metadata); err != nil {
			return out, err
		}
		out = append(out, item)
	}
	return out, rows.Err()
}

func scanFindXAgentConfigRollouts(rows *sql.Rows) ([]model.FindXAgentConfigRollout, error) {
	out := []model.FindXAgentConfigRollout{}
	for rows.Next() {
		item, err := scanFindXAgentConfigRollout(rows)
		if err != nil {
			return out, err
		}
		out = append(out, item)
	}
	return out, rows.Err()
}

func findXAgentConfigRolloutSelectSQL() string {
	return `SELECT id,template_id,COALESCE(agent_ids,'[]'),COALESCE(target_ids,'[]'),config_version,config_snippet_ref,config_format,provider_mode,plugin_id,plugin_version,reload_strategy,restart_strategy,rollout_strategy,rollback_ref,audit_reason,change_ticket,remote_mutation,canary_percent,credential_ref_present,status,blocker,audit,COALESCE(evidence_refs,'[]'),COALESCE(metadata,'{}'),created_at,updated_at FROM findx_agent_config_rollouts`
}

func scanFindXAgentConfigRollout(row interface{ Scan(dest ...any) error }) (model.FindXAgentConfigRollout, error) {
	var item model.FindXAgentConfigRollout
	var agentIDs, targetIDs, evidenceRefs, metadata string
	err := row.Scan(&item.ID, &item.TemplateID, &agentIDs, &targetIDs, &item.ConfigVersion, &item.ConfigSnippetRef, &item.ConfigFormat, &item.ProviderMode, &item.PluginID, &item.PluginVersion, &item.ReloadStrategy, &item.RestartStrategy, &item.RolloutStrategy, &item.RollbackRef, &item.AuditReason, &item.ChangeTicket, &item.RemoteMutation, &item.CanaryPercent, &item.CredentialRefPresent, &item.Status, &item.Blocker, &item.Audit, &evidenceRefs, &metadata, &item.CreatedAt, &item.UpdatedAt)
	if err != nil {
		return item, err
	}
	if err := decodeLifecycleJSON(agentIDs, &item.AgentIDs); err != nil {
		return item, err
	}
	if err := decodeLifecycleJSON(targetIDs, &item.TargetIDs); err != nil {
		return item, err
	}
	if err := decodeLifecycleJSON(evidenceRefs, &item.EvidenceRefs); err != nil {
		return item, err
	}
	if err := decodeLifecycleJSON(metadata, &item.Metadata); err != nil {
		return item, err
	}
	return item, nil
}

func scanFindXAgentExecutionTasks(rows *sql.Rows) ([]model.FindXAgentExecutionTask, error) {
	out := []model.FindXAgentExecutionTask{}
	for rows.Next() {
		var item model.FindXAgentExecutionTask
		var agentIDs, targetIDs, evidenceRefs, metadata string
		if err := rows.Scan(&item.ID, &item.Action, &agentIDs, &targetIDs, &item.PackageID, &item.ConfigVersion, &item.CredentialRefPresent, &item.Status, &item.Blocker, &item.Audit, &evidenceRefs, &metadata, &item.CreatedAt, &item.UpdatedAt); err != nil {
			return out, err
		}
		if err := decodeLifecycleJSON(agentIDs, &item.AgentIDs); err != nil {
			return out, err
		}
		if err := decodeLifecycleJSON(targetIDs, &item.TargetIDs); err != nil {
			return out, err
		}
		if err := decodeLifecycleJSON(evidenceRefs, &item.EvidenceRefs); err != nil {
			return out, err
		}
		if err := decodeLifecycleJSON(metadata, &item.Metadata); err != nil {
			return out, err
		}
		out = append(out, item)
	}
	return out, rows.Err()
}

func scanFindXAgentDataArrivalEvidence(rows *sql.Rows) ([]model.FindXAgentDataArrivalEvidence, error) {
	out := []model.FindXAgentDataArrivalEvidence{}
	for rows.Next() {
		var item model.FindXAgentDataArrivalEvidence
		var evidenceRefs, metadata string
		if err := rows.Scan(&item.ID, &item.Kind, &item.AgentID, &item.TargetID, &item.Status, &item.Blocker, &evidenceRefs, &metadata, &item.CreatedAt, &item.UpdatedAt); err != nil {
			return out, err
		}
		if err := decodeLifecycleJSON(evidenceRefs, &item.EvidenceRefs); err != nil {
			return out, err
		}
		if err := decodeLifecycleJSON(metadata, &item.Metadata); err != nil {
			return out, err
		}
		out = append(out, item)
	}
	return out, rows.Err()
}

func decodeLifecycleJSON(raw string, dest any) error {
	if strings.TrimSpace(raw) == "" {
		return nil
	}
	return json.Unmarshal([]byte(raw), dest)
}

func copyFindXAgentInstallPlan(item model.FindXAgentInstallPlan) model.FindXAgentInstallPlan {
	item.TargetIDs = append([]string{}, item.TargetIDs...)
	item.EvidenceRefs = append([]string{}, item.EvidenceRefs...)
	item.Metadata = copyStringMap(item.Metadata)
	return item
}

func copyFindXAgentConfigRollout(item model.FindXAgentConfigRollout) model.FindXAgentConfigRollout {
	item.AgentIDs = append([]string{}, item.AgentIDs...)
	item.TargetIDs = append([]string{}, item.TargetIDs...)
	item.EvidenceRefs = append([]string{}, item.EvidenceRefs...)
	item.Metadata = copyStringMap(item.Metadata)
	return item
}

func copyFindXAgentExecutionTask(item model.FindXAgentExecutionTask) model.FindXAgentExecutionTask {
	item.AgentIDs = append([]string{}, item.AgentIDs...)
	item.TargetIDs = append([]string{}, item.TargetIDs...)
	item.EvidenceRefs = append([]string{}, item.EvidenceRefs...)
	item.Metadata = copyStringMap(item.Metadata)
	return item
}

func copyFindXAgentDataArrivalEvidence(item model.FindXAgentDataArrivalEvidence) model.FindXAgentDataArrivalEvidence {
	item.EvidenceRefs = append([]string{}, item.EvidenceRefs...)
	item.Metadata = copyStringMap(item.Metadata)
	return item
}

func sortFindXAgentInstallPlans(items []model.FindXAgentInstallPlan) {
	sort.Slice(items, func(i, j int) bool { return items[i].UpdatedAt.After(items[j].UpdatedAt) })
}

func sortFindXAgentConfigRollouts(items []model.FindXAgentConfigRollout) {
	sort.Slice(items, func(i, j int) bool { return items[i].UpdatedAt.After(items[j].UpdatedAt) })
}

func sortFindXAgentExecutionTasks(items []model.FindXAgentExecutionTask) {
	sort.Slice(items, func(i, j int) bool { return items[i].UpdatedAt.After(items[j].UpdatedAt) })
}

func sortFindXAgentDataArrivalEvidence(items []model.FindXAgentDataArrivalEvidence) {
	sort.Slice(items, func(i, j int) bool { return items[i].UpdatedAt.After(items[j].UpdatedAt) })
}
