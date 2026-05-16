package handler

import (
	"net/http"
	"strings"
	"time"

	"ai-workbench-api/internal/model"
	"ai-workbench-api/internal/store"

	"github.com/gin-gonic/gin"
)

const evidenceChainBlockedByContract = "pending"
const evidenceChainRedactedBlocker = evidenceChainBlockedByContract + ": evidence detail redacted"

type aiopsEvidenceChainResponse struct {
	Items      []aiopsEvidenceChainItem     `json:"items"`
	Categories []aiopsEvidenceChainCategory `json:"categories"`
	Blockers   []aiopsEvidenceChainBlocker  `json:"blockers"`
	Summary    aiopsEvidenceChainSummary    `json:"summary"`
}

type aiopsEvidenceChainItem struct {
	ID           string            `json:"id"`
	Category     string            `json:"category"`
	SourceType   string            `json:"source_type"`
	Kind         string            `json:"kind,omitempty"`
	Status       string            `json:"status"`
	Blocker      string            `json:"blocker,omitempty"`
	EvidenceRefs []string          `json:"evidence_refs,omitempty"`
	Metadata     map[string]string `json:"metadata,omitempty"`
	UpdatedAt    time.Time         `json:"updated_at,omitempty"`
}

type aiopsEvidenceChainCategory struct {
	Key         string `json:"key"`
	Total       int    `json:"total"`
	Reported    int    `json:"reported"`
	Blocked     int    `json:"blocked"`
	Error       int    `json:"error"`
	Missing     int    `json:"missing"`
	LatestState string `json:"latest_state"`
}

type aiopsEvidenceChainBlocker struct {
	Key    string   `json:"key"`
	Reason string   `json:"reason"`
	Items  []string `json:"items"`
}

type aiopsEvidenceChainSummary struct {
	TotalItems    int `json:"total_items"`
	ReportedItems int `json:"reported_items"`
	BlockedItems  int `json:"blocked_items"`
	ErrorItems    int `json:"error_items"`
	MissingItems  int `json:"missing_items"`
}

func AIOpsEvidenceChain(c *gin.Context) {
	data, err := buildAIOpsEvidenceChain()
	if err != nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"code": 503, "error": "evidence chain unavailable"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 0, "data": data})
}

func buildAIOpsEvidenceChain() (aiopsEvidenceChainResponse, error) {
	var out aiopsEvidenceChainResponse
	if err := appendEvidenceChainRecords(&out); err != nil {
		return out, err
	}
	appendMissingEvidenceChainContracts(&out)
	sortEvidenceChainItems(out.Items)
	out.Categories = summarizeEvidenceChainCategories(out.Items)
	out.Blockers = summarizeEvidenceChainBlockers(out.Items)
	out.Summary = summarizeEvidenceChainItems(out.Items)
	return out, nil
}

func appendEvidenceChainRecords(out *aiopsEvidenceChainResponse) error {
	evidence, err := store.ListFindXAgentDataArrivalEvidence()
	if err != nil {
		return err
	}
	for _, item := range evidence {
		out.Items = append(out.Items, dataArrivalEvidenceChainItem(item))
	}
	plans, err := store.ListFindXAgentInstallPlans()
	if err != nil {
		return err
	}
	for _, item := range plans {
		out.Items = append(out.Items, installPlanEvidenceChainItem(item))
	}
	executions, err := store.ListFindXAgentInstallExecutions()
	if err != nil {
		return err
	}
	for _, item := range executions {
		out.Items = append(out.Items, installExecutionEvidenceChainItem(item))
	}
	rollouts, err := store.ListFindXAgentConfigRollouts()
	if err != nil {
		return err
	}
	for _, item := range rollouts {
		out.Items = append(out.Items, configRolloutEvidenceChainItem(item))
	}
	tasks, err := store.ListFindXAgentExecutionTasks()
	if err != nil {
		return err
	}
	for _, item := range tasks {
		out.Items = append(out.Items, taskEvidenceChainItem(item))
	}
	return nil
}

func appendMissingEvidenceChainContracts(out *aiopsEvidenceChainResponse) {
	present := map[string]bool{}
	for _, item := range out.Items {
		if item.Category == "data_arrival" && item.Status == model.FindXAgentDataArrivalStatusReported {
			present[normalizeEvidenceChainKind(item.Kind)] = true
		}
	}
	for _, kind := range requiredEvidenceChainContracts() {
		if present[kind] {
			continue
		}
		out.Items = append(out.Items, missingEvidenceChainItem(kind))
	}
}

func dataArrivalEvidenceChainItem(item model.FindXAgentDataArrivalEvidence) aiopsEvidenceChainItem {
	kind := normalizeEvidenceChainKind(item.Kind)
	return aiopsEvidenceChainItem{
		ID:           safeEvidenceChainID(item.ID, "data_arrival", "receiver_evidence", kind),
		Category:     "data_arrival",
		SourceType:   "receiver_evidence",
		Kind:         kind,
		Status:       normalizeEvidenceChainStatus(item.Status),
		Blocker:      cleanEvidenceChainBlocker(item.Blocker),
		EvidenceRefs: safeEvidenceChainRefs(item.EvidenceRefs),
		Metadata:     safeEvidenceChainMetadata(item.Metadata),
		UpdatedAt:    item.UpdatedAt,
	}
}

func installPlanEvidenceChainItem(item model.FindXAgentInstallPlan) aiopsEvidenceChainItem {
	return aiopsEvidenceChainItem{ID: safeEvidenceChainID(item.ID, "install_plan", "install_plan", ""), Category: "install_plan", SourceType: "install_plan", Status: "blocked", Blocker: cleanEvidenceChainBlocker(item.Blocker), EvidenceRefs: safeEvidenceChainRefs(item.EvidenceRefs), Metadata: safeEvidenceChainMetadata(item.Metadata), UpdatedAt: item.UpdatedAt}
}

func installExecutionEvidenceChainItem(item model.FindXAgentInstallExecution) aiopsEvidenceChainItem {
	metadata := map[string]string{"plan_id": item.PlanID, "target_id": item.TargetID, "runner": item.Runner}
	return aiopsEvidenceChainItem{ID: safeEvidenceChainID(item.ID, "install_execution", "install_execution", ""), Category: "install_execution", SourceType: "install_execution", Status: "blocked", Blocker: cleanEvidenceChainBlocker(item.ErrorSummary), EvidenceRefs: safeEvidenceChainRefs(item.EvidenceRefs), Metadata: safeEvidenceChainMetadata(metadata), UpdatedAt: item.UpdatedAt}
}

func configRolloutEvidenceChainItem(item model.FindXAgentConfigRollout) aiopsEvidenceChainItem {
	return aiopsEvidenceChainItem{ID: safeEvidenceChainID(item.ID, "config_rollout", "config_rollout", ""), Category: "config_rollout", SourceType: "config_rollout", Status: "blocked", Blocker: cleanEvidenceChainBlocker(item.Blocker), EvidenceRefs: safeEvidenceChainRefs(item.EvidenceRefs), Metadata: safeEvidenceChainMetadata(item.Metadata), UpdatedAt: item.UpdatedAt}
}

func taskEvidenceChainItem(item model.FindXAgentExecutionTask) aiopsEvidenceChainItem {
	metadata := safeEvidenceChainMetadata(item.Metadata)
	metadata["action"] = item.Action
	return aiopsEvidenceChainItem{ID: safeEvidenceChainID(item.ID, "lifecycle_task", "agent_lifecycle_task", item.Action), Category: "lifecycle_task", SourceType: "agent_lifecycle_task", Status: "blocked", Blocker: cleanEvidenceChainBlocker(item.Blocker), EvidenceRefs: safeEvidenceChainRefs(item.EvidenceRefs), Metadata: metadata, UpdatedAt: item.UpdatedAt}
}

func missingEvidenceChainItem(kind string) aiopsEvidenceChainItem {
	return aiopsEvidenceChainItem{ID: "missing-" + kind, Category: categoryForEvidenceChainKind(kind), SourceType: "contract_gap", Kind: kind, Status: "blocked", Blocker: evidenceChainBlockedByContract + ": missing " + kind + " evidence"}
}

func requiredEvidenceChainContracts() []string {
	return []string{"metrics", "logs", "traces", "profiling", "inspection", "rum", "gateway_trace", "ai_conclusion"}
}

func normalizeEvidenceChainKind(kind string) string {
	if kind == model.FindXAgentDataArrivalKindTracing {
		return "traces"
	}
	return strings.TrimSpace(kind)
}

func categoryForEvidenceChainKind(kind string) string {
	if kind == "ai_conclusion" {
		return "ai_conclusion"
	}
	return "data_arrival"
}

func normalizeEvidenceChainStatus(status string) string {
	switch strings.TrimSpace(status) {
	case model.FindXAgentDataArrivalStatusReported:
		return model.FindXAgentDataArrivalStatusReported
	case model.FindXAgentDataArrivalStatusError:
		return model.FindXAgentDataArrivalStatusError
	default:
		return model.FindXAgentDataArrivalStatusBlocked
	}
}
