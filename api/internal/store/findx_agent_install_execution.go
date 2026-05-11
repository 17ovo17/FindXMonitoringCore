package store

import (
	"database/sql"
	"sort"
	"strings"
	"time"

	"ai-workbench-api/internal/model"
)

func SaveFindXAgentInstallExecution(input model.FindXAgentInstallExecution) (model.FindXAgentInstallExecution, error) {
	item := normalizeFindXAgentInstallExecution(input, time.Now())
	cp := copyFindXAgentInstallExecution(item)
	mu.Lock()
	findxAgentInstallExecutions[item.ID] = &cp
	mu.Unlock()
	if mysqlOK {
		if err := persistFindXAgentInstallExecution(item); err != nil {
			return copyFindXAgentInstallExecution(item), err
		}
	}
	return copyFindXAgentInstallExecution(item), nil
}

func ListFindXAgentInstallExecutions() ([]model.FindXAgentInstallExecution, error) {
	if mysqlOK {
		rows, err := db.Query(`SELECT id,plan_id,target_id,runner,status,exit_code,COALESCE(steps,'[]'),COALESCE(evidence_refs,'[]'),error_summary,created_at,started_at,finished_at,updated_at FROM findx_agent_install_executions WHERE status=? ORDER BY updated_at DESC LIMIT 500`, model.FindXAgentExecutionStatusBlocked)
		if err == nil {
			defer rows.Close()
			return scanFindXAgentInstallExecutions(rows)
		}
	}
	mu.RLock()
	defer mu.RUnlock()
	out := make([]model.FindXAgentInstallExecution, 0, len(findxAgentInstallExecutions))
	for _, item := range findxAgentInstallExecutions {
		if item.Status != model.FindXAgentExecutionStatusBlocked {
			continue
		}
		out = append(out, copyFindXAgentInstallExecution(*item))
	}
	sortFindXAgentInstallExecutions(out)
	return out, nil
}

func GetFindXAgentInstallExecution(id string) (model.FindXAgentInstallExecution, bool, error) {
	if mysqlOK {
		row := db.QueryRow(`SELECT id,plan_id,target_id,runner,status,exit_code,COALESCE(steps,'[]'),COALESCE(evidence_refs,'[]'),error_summary,created_at,started_at,finished_at,updated_at FROM findx_agent_install_executions WHERE id=? AND status=?`, id, model.FindXAgentExecutionStatusBlocked)
		item, err := scanFindXAgentInstallExecution(row)
		if err == nil {
			return item, true, nil
		}
		if err != nil && err != sql.ErrNoRows {
			return model.FindXAgentInstallExecution{}, false, err
		}
	}
	mu.RLock()
	defer mu.RUnlock()
	item, ok := findxAgentInstallExecutions[id]
	if !ok || item.Status != model.FindXAgentExecutionStatusBlocked {
		return model.FindXAgentInstallExecution{}, false, nil
	}
	return copyFindXAgentInstallExecution(*item), true, nil
}

func ResetFindXAgentInstallExecutionsForTest() {
	mu.Lock()
	defer mu.Unlock()
	findxAgentInstallExecutions = map[string]*model.FindXAgentInstallExecution{}
	mysqlOK = false
}

func resetFindXAgentInstallExecutionsLocked() {
	findxAgentInstallExecutions = map[string]*model.FindXAgentInstallExecution{}
}

func normalizeFindXAgentInstallExecution(item model.FindXAgentInstallExecution, now time.Time) model.FindXAgentInstallExecution {
	if item.ID == "" {
		item.ID = NewID()
	}
	item.PlanID = strings.TrimSpace(item.PlanID)
	item.TargetID = strings.TrimSpace(item.TargetID)
	item.Runner = firstLifecycleValue(strings.TrimSpace(item.Runner), "ssh")
	item.Status = model.FindXAgentExecutionStatusBlocked
	item.Steps = cleanInstallExecutionSteps(item.Steps)
	item.EvidenceRefs = cleanStringList(item.EvidenceRefs)
	item.ErrorSummary = sanitizeInstallExecutionText(item.ErrorSummary)
	if item.ErrorSummary == "" {
		item.ErrorSummary = findXAgentExecutorBlockedReason
	}
	if item.CreatedAt.IsZero() {
		item.CreatedAt = now
	}
	item.UpdatedAt = now
	return item
}

func persistFindXAgentInstallExecution(item model.FindXAgentInstallExecution) error {
	steps := lifecycleJSON(item.Steps)
	evidenceRefs := lifecycleJSON(item.EvidenceRefs)
	_, err := db.Exec(`REPLACE INTO findx_agent_install_executions (id,plan_id,target_id,runner,status,exit_code,steps,evidence_refs,error_summary,created_at,started_at,finished_at,updated_at) VALUES (?,?,?,?,?,?,?,?,?,?,?,?,?)`,
		item.ID, item.PlanID, item.TargetID, item.Runner, item.Status, nullableInt(item.ExitCode), steps, evidenceRefs, item.ErrorSummary, item.CreatedAt, nullableInstallExecutionTime(item.StartedAt), nullableInstallExecutionTime(item.FinishedAt), item.UpdatedAt)
	return err
}

func scanFindXAgentInstallExecutions(rows *sql.Rows) ([]model.FindXAgentInstallExecution, error) {
	out := []model.FindXAgentInstallExecution{}
	for rows.Next() {
		item, err := scanFindXAgentInstallExecution(rows)
		if err != nil {
			return out, err
		}
		out = append(out, item)
	}
	return out, rows.Err()
}

func scanFindXAgentInstallExecution(row interface{ Scan(dest ...any) error }) (model.FindXAgentInstallExecution, error) {
	var item model.FindXAgentInstallExecution
	var exitCode sql.NullInt64
	var startedAt, finishedAt sql.NullTime
	var steps, evidenceRefs string
	err := row.Scan(&item.ID, &item.PlanID, &item.TargetID, &item.Runner, &item.Status, &exitCode, &steps, &evidenceRefs, &item.ErrorSummary, &item.CreatedAt, &startedAt, &finishedAt, &item.UpdatedAt)
	if err != nil {
		return item, err
	}
	if exitCode.Valid {
		value := int(exitCode.Int64)
		item.ExitCode = &value
	}
	if startedAt.Valid {
		value := startedAt.Time
		item.StartedAt = &value
	}
	if finishedAt.Valid {
		value := finishedAt.Time
		item.FinishedAt = &value
	}
	if err := decodeLifecycleJSON(steps, &item.Steps); err != nil {
		return item, err
	}
	if err := decodeLifecycleJSON(evidenceRefs, &item.EvidenceRefs); err != nil {
		return item, err
	}
	item.Status = model.FindXAgentExecutionStatusBlocked
	item.Steps = cleanInstallExecutionSteps(item.Steps)
	item.EvidenceRefs = cleanStringList(item.EvidenceRefs)
	item.ErrorSummary = sanitizeInstallExecutionText(item.ErrorSummary)
	if item.ErrorSummary == "" {
		item.ErrorSummary = findXAgentExecutorBlockedReason
	}
	return item, nil
}

func cleanInstallExecutionSteps(steps []model.FindXAgentInstallExecutionStep) []model.FindXAgentInstallExecutionStep {
	out := make([]model.FindXAgentInstallExecutionStep, 0, len(steps))
	for _, step := range steps {
		name := strings.TrimSpace(step.Name)
		if name == "" {
			continue
		}
		step.Status = model.FindXAgentExecutionStatusBlocked
		step.Name = name
		step.EvidenceRef = sanitizeInstallExecutionText(step.EvidenceRef)
		step.ErrorSummary = sanitizeInstallExecutionText(step.ErrorSummary)
		if step.ErrorSummary == "" {
			step.ErrorSummary = findXAgentExecutorBlockedReason
		}
		out = append(out, step)
	}
	return out
}

func copyFindXAgentInstallExecution(item model.FindXAgentInstallExecution) model.FindXAgentInstallExecution {
	if item.ExitCode != nil {
		value := *item.ExitCode
		item.ExitCode = &value
	}
	item.Steps = append([]model.FindXAgentInstallExecutionStep{}, item.Steps...)
	item.EvidenceRefs = append([]string{}, item.EvidenceRefs...)
	item.StartedAt = copyTimePointer(item.StartedAt)
	item.FinishedAt = copyTimePointer(item.FinishedAt)
	return item
}

func sortFindXAgentInstallExecutions(items []model.FindXAgentInstallExecution) {
	sort.Slice(items, func(i, j int) bool { return items[i].UpdatedAt.After(items[j].UpdatedAt) })
}

func nullableInt(value *int) any {
	if value == nil {
		return nil
	}
	return *value
}

func nullableInstallExecutionTime(value *time.Time) any {
	if value == nil || value.IsZero() {
		return nil
	}
	return *value
}

func copyTimePointer(value *time.Time) *time.Time {
	if value == nil {
		return nil
	}
	cp := *value
	return &cp
}

func sanitizeInstallExecutionText(value string) string {
	clean := strings.TrimSpace(removeControlCharacters(value))
	if clean == "" || isSensitiveKey(clean) {
		return ""
	}
	const maxInstallExecutionTextLen = 240
	runes := []rune(clean)
	if len(runes) > maxInstallExecutionTextLen {
		clean = string(runes[:maxInstallExecutionTextLen])
	}
	return clean
}

func removeControlCharacters(value string) string {
	return strings.Map(func(r rune) rune {
		if r < 32 || r == 127 {
			return -1
		}
		return r
	}, value)
}
