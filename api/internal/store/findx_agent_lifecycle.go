package store

import (
	"database/sql"
	"time"

	"ai-workbench-api/internal/model"
)

const findXAgentBlockedStatus = "blocked"

func SaveFindXAgentInstallPlan(input model.FindXAgentInstallPlan) (model.FindXAgentInstallPlan, error) {
	item := normalizeFindXAgentInstallPlan(input, time.Now())
	cp := copyFindXAgentInstallPlan(item)
	mu.Lock()
	findxAgentInstallPlans[item.ID] = &cp
	mu.Unlock()
	if mysqlOK {
		if err := persistFindXAgentInstallPlan(item); err != nil {
			return copyFindXAgentInstallPlan(item), err
		}
	}
	return copyFindXAgentInstallPlan(item), nil
}

func ListFindXAgentInstallPlans() ([]model.FindXAgentInstallPlan, error) {
	if mysqlOK {
		rows, err := db.Query(findXAgentInstallPlanSelectSQL()+` WHERE status=? ORDER BY updated_at DESC LIMIT 500`, findXAgentBlockedStatus)
		if err == nil {
			defer rows.Close()
			return scanFindXAgentInstallPlans(rows)
		}
	}
	mu.RLock()
	defer mu.RUnlock()
	out := make([]model.FindXAgentInstallPlan, 0, len(findxAgentInstallPlans))
	for _, item := range findxAgentInstallPlans {
		if item.Status != findXAgentBlockedStatus {
			continue
		}
		out = append(out, copyFindXAgentInstallPlan(*item))
	}
	sortFindXAgentInstallPlans(out)
	return out, nil
}

func GetFindXAgentInstallPlan(id string) (model.FindXAgentInstallPlan, bool, error) {
	if mysqlOK {
		rows, err := db.Query(findXAgentInstallPlanSelectSQL()+` WHERE id=? AND status=?`, id, findXAgentBlockedStatus)
		if err == nil {
			defer rows.Close()
			items, scanErr := scanFindXAgentInstallPlans(rows)
			if scanErr != nil {
				return model.FindXAgentInstallPlan{}, false, scanErr
			}
			if len(items) > 0 {
				return items[0], true, nil
			}
		} else if err != sql.ErrNoRows {
			return model.FindXAgentInstallPlan{}, false, err
		}
	}
	mu.RLock()
	defer mu.RUnlock()
	item, ok := findxAgentInstallPlans[id]
	if !ok || item.Status != findXAgentBlockedStatus {
		return model.FindXAgentInstallPlan{}, false, nil
	}
	return copyFindXAgentInstallPlan(*item), true, nil
}

func findXAgentInstallPlanSelectSQL() string {
	return `SELECT id,package_id,os,method,COALESCE(target_ids,'[]'),credential_ref_present,status,blocker,audit,COALESCE(evidence_refs,'[]'),COALESCE(metadata,'{}'),created_at,updated_at FROM findx_agent_install_plans`
}

func SaveFindXAgentConfigRollout(input model.FindXAgentConfigRollout) (model.FindXAgentConfigRollout, error) {
	item := normalizeFindXAgentConfigRollout(input, time.Now())
	cp := copyFindXAgentConfigRollout(item)
	mu.Lock()
	findxAgentConfigRollouts[item.ID] = &cp
	mu.Unlock()
	if mysqlOK {
		if err := persistFindXAgentConfigRollout(item); err != nil {
			return copyFindXAgentConfigRollout(item), err
		}
	}
	return copyFindXAgentConfigRollout(item), nil
}

func ListFindXAgentConfigRollouts() ([]model.FindXAgentConfigRollout, error) {
	if mysqlOK {
		rows, err := db.Query(findXAgentConfigRolloutSelectSQL()+` WHERE status=? ORDER BY updated_at DESC LIMIT 500`, findXAgentBlockedStatus)
		if err == nil {
			defer rows.Close()
			return scanFindXAgentConfigRollouts(rows)
		}
	}
	mu.RLock()
	defer mu.RUnlock()
	out := make([]model.FindXAgentConfigRollout, 0, len(findxAgentConfigRollouts))
	for _, item := range findxAgentConfigRollouts {
		if item.Status != findXAgentBlockedStatus {
			continue
		}
		out = append(out, copyFindXAgentConfigRollout(*item))
	}
	sortFindXAgentConfigRollouts(out)
	return out, nil
}

func GetFindXAgentConfigRollout(id string) (model.FindXAgentConfigRollout, bool, error) {
	if mysqlOK {
		row := db.QueryRow(findXAgentConfigRolloutSelectSQL()+` WHERE id=? AND status=?`, id, findXAgentBlockedStatus)
		item, err := scanFindXAgentConfigRollout(row)
		if err == nil {
			return item, true, nil
		}
		if err != nil && err != sql.ErrNoRows {
			return model.FindXAgentConfigRollout{}, false, err
		}
	}
	mu.RLock()
	defer mu.RUnlock()
	item, ok := findxAgentConfigRollouts[id]
	if !ok || item.Status != findXAgentBlockedStatus {
		return model.FindXAgentConfigRollout{}, false, nil
	}
	return copyFindXAgentConfigRollout(*item), true, nil
}

func SaveFindXAgentExecutionTask(input model.FindXAgentExecutionTask) (model.FindXAgentExecutionTask, error) {
	item := normalizeFindXAgentExecutionTask(input, time.Now())
	cp := copyFindXAgentExecutionTask(item)
	mu.Lock()
	findxAgentExecutionTasks[item.ID] = &cp
	mu.Unlock()
	if mysqlOK {
		if err := persistFindXAgentExecutionTask(item); err != nil {
			return copyFindXAgentExecutionTask(item), err
		}
	}
	return copyFindXAgentExecutionTask(item), nil
}

func ListFindXAgentExecutionTasks() ([]model.FindXAgentExecutionTask, error) {
	if mysqlOK {
		rows, err := db.Query(findXAgentExecutionTaskSelectSQL()+` WHERE status=? ORDER BY updated_at DESC LIMIT 500`, findXAgentBlockedStatus)
		if err == nil {
			defer rows.Close()
			return scanFindXAgentExecutionTasks(rows)
		}
	}
	mu.RLock()
	defer mu.RUnlock()
	out := make([]model.FindXAgentExecutionTask, 0, len(findxAgentExecutionTasks))
	for _, item := range findxAgentExecutionTasks {
		if item.Status != findXAgentBlockedStatus {
			continue
		}
		out = append(out, copyFindXAgentExecutionTask(*item))
	}
	sortFindXAgentExecutionTasks(out)
	return out, nil
}

func GetFindXAgentExecutionTask(id string) (model.FindXAgentExecutionTask, bool, error) {
	if mysqlOK {
		rows, err := db.Query(findXAgentExecutionTaskSelectSQL()+` WHERE id=? AND status=?`, id, findXAgentBlockedStatus)
		if err == nil {
			defer rows.Close()
			items, scanErr := scanFindXAgentExecutionTasks(rows)
			if scanErr != nil {
				return model.FindXAgentExecutionTask{}, false, scanErr
			}
			if len(items) > 0 {
				return items[0], true, nil
			}
		} else if err != sql.ErrNoRows {
			return model.FindXAgentExecutionTask{}, false, err
		}
	}
	mu.RLock()
	defer mu.RUnlock()
	item, ok := findxAgentExecutionTasks[id]
	if !ok || item.Status != findXAgentBlockedStatus {
		return model.FindXAgentExecutionTask{}, false, nil
	}
	return copyFindXAgentExecutionTask(*item), true, nil
}

func findXAgentExecutionTaskSelectSQL() string {
	return `SELECT id,action,COALESCE(agent_ids,'[]'),COALESCE(target_ids,'[]'),package_id,config_version,credential_ref_present,status,blocker,audit,COALESCE(evidence_refs,'[]'),COALESCE(metadata,'{}'),created_at,updated_at FROM findx_agent_execution_tasks`
}

func ListFindXAgentDataArrivalEvidence() ([]model.FindXAgentDataArrivalEvidence, error) {
	if mysqlOK {
		rows, err := db.Query(`SELECT id,kind,agent_id,target_id,status,blocker,COALESCE(evidence_refs,'[]'),COALESCE(metadata,'{}'),created_at,updated_at FROM findx_agent_data_arrival_evidence ORDER BY updated_at DESC LIMIT 500`)
		if err == nil {
			defer rows.Close()
			return scanFindXAgentDataArrivalEvidence(rows)
		}
	}
	mu.RLock()
	defer mu.RUnlock()
	out := make([]model.FindXAgentDataArrivalEvidence, 0, len(findxAgentDataArrivalEvidence))
	for _, item := range findxAgentDataArrivalEvidence {
		out = append(out, copyFindXAgentDataArrivalEvidence(*item))
	}
	sortFindXAgentDataArrivalEvidence(out)
	return out, nil
}

func ResetFindXAgentLifecycleForTest() {
	mu.Lock()
	defer mu.Unlock()
	findxAgentInstallPlans = map[string]*model.FindXAgentInstallPlan{}
	findxAgentConfigRollouts = map[string]*model.FindXAgentConfigRollout{}
	findxAgentExecutionTasks = map[string]*model.FindXAgentExecutionTask{}
	findxAgentDataArrivalEvidence = map[string]*model.FindXAgentDataArrivalEvidence{}
	mysqlOK = false
}

func persistFindXAgentInstallPlan(item model.FindXAgentInstallPlan) error {
	targetIDs, evidenceRefs, metadata := lifecycleJSON(item.TargetIDs), lifecycleJSON(item.EvidenceRefs), lifecycleJSON(item.Metadata)
	_, err := db.Exec(`REPLACE INTO findx_agent_install_plans (id,package_id,os,method,target_ids,credential_ref_present,status,blocker,audit,evidence_refs,metadata,created_at,updated_at) VALUES (?,?,?,?,?,?,?,?,?,?,?,?,?)`,
		item.ID, item.PackageID, item.OS, item.Method, targetIDs, item.CredentialRefPresent, item.Status, item.Blocker, item.Audit, evidenceRefs, metadata, item.CreatedAt, item.UpdatedAt)
	return err
}

func persistFindXAgentConfigRollout(item model.FindXAgentConfigRollout) error {
	agentIDs, targetIDs := lifecycleJSON(item.AgentIDs), lifecycleJSON(item.TargetIDs)
	evidenceRefs, metadata := lifecycleJSON(item.EvidenceRefs), lifecycleJSON(item.Metadata)
	_, err := db.Exec(`REPLACE INTO findx_agent_config_rollouts (id,template_id,agent_ids,target_ids,config_version,config_snippet_ref,config_format,provider_mode,plugin_id,plugin_version,reload_strategy,restart_strategy,rollout_strategy,rollback_ref,audit_reason,change_ticket,remote_mutation,canary_percent,credential_ref_present,status,blocker,audit,evidence_refs,metadata,created_at,updated_at) VALUES (?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?)`,
		item.ID, item.TemplateID, agentIDs, targetIDs, item.ConfigVersion, item.ConfigSnippetRef, item.ConfigFormat, item.ProviderMode, item.PluginID, item.PluginVersion, item.ReloadStrategy, item.RestartStrategy, item.RolloutStrategy, item.RollbackRef, item.AuditReason, item.ChangeTicket, item.RemoteMutation, item.CanaryPercent, item.CredentialRefPresent, item.Status, item.Blocker, item.Audit, evidenceRefs, metadata, item.CreatedAt, item.UpdatedAt)
	return err
}

func persistFindXAgentExecutionTask(item model.FindXAgentExecutionTask) error {
	agentIDs, targetIDs := lifecycleJSON(item.AgentIDs), lifecycleJSON(item.TargetIDs)
	evidenceRefs, metadata := lifecycleJSON(item.EvidenceRefs), lifecycleJSON(item.Metadata)
	_, err := db.Exec(`REPLACE INTO findx_agent_execution_tasks (id,action,agent_ids,target_ids,package_id,config_version,credential_ref_present,status,blocker,audit,evidence_refs,metadata,created_at,updated_at) VALUES (?,?,?,?,?,?,?,?,?,?,?,?,?,?)`,
		item.ID, item.Action, agentIDs, targetIDs, item.PackageID, item.ConfigVersion, item.CredentialRefPresent, item.Status, item.Blocker, item.Audit, evidenceRefs, metadata, item.CreatedAt, item.UpdatedAt)
	return err
}
