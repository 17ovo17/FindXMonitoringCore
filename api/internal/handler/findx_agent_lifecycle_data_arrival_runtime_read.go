package handler

import (
	"net/http"
	"strings"

	"ai-workbench-api/internal/model"
	"ai-workbench-api/internal/store"

	"github.com/gin-gonic/gin"
)

const (
	cmdbAgentRolloutDataArrivalReadContract       = "cmdb.agent.plugin.dispatch.data_arrival.read.v1"
	cmdbAgentRolloutDataArrivalEvidenceContract   = "cmdb_agent_rollout_data_arrival_evidence_contract"
	cmdbAgentRolloutReceiverEvidenceContract      = "cmdb_agent_rollout_receiver_evidence_contract"
	cmdbAgentRolloutDataArrivalRequestRefContract = "cmdb_agent_rollout_data_arrival_request_ref_contract"
)

type dataArrivalRuntimeReadQuery struct {
	RolloutID  string
	RequestRef string
	PluginID   string
}

func listFindXAgentDataArrivalEvidenceRuntimeRead(c *gin.Context) (bool, error) {
	query := dataArrivalRuntimeReadQuery{
		RolloutID:  firstNonEmpty(c.Query("rollout_id"), c.Query("config_rollout_id"), c.Query("rollout_ref")),
		RequestRef: sanitizeRemoteMutationValue("request_ref", c.Query("request_ref")),
		PluginID:   sanitizeRemoteMutationValue("plugin_id", c.Query("plugin_id")),
	}
	if strings.TrimSpace(query.RolloutID) == "" && strings.TrimSpace(query.RequestRef) == "" {
		return false, nil
	}
	if query.RolloutID == "" && query.RequestRef != "" {
		query.RolloutID = rolloutIDFromReceiptRequestRef(query.RequestRef)
	}
	rollout, ok, err := store.GetFindXAgentConfigRollout(query.RolloutID)
	if err != nil {
		return true, err
	}
	if !ok || !isCMDBHostPluginDispatchRolloutRecord(rollout) {
		writeDataArrivalRuntimeReadBlocked(c, rollout, query, []string{
			cmdbAgentRolloutDataArrivalRequestRefContract,
			cmdbAgentRolloutDataArrivalEvidenceContract,
			cmdbAgentRolloutReceiverEvidenceContract,
		})
		return true, nil
	}
	if query.PluginID != "" && query.PluginID != rollout.PluginID {
		writeDataArrivalRuntimeReadBlocked(c, rollout, query, []string{
			cmdbAgentRolloutTaskMatchContract,
			cmdbAgentRolloutDataArrivalEvidenceContract,
		})
		return true, nil
	}
	gate := configRolloutRuntimeReadGateForItem(rollout)
	missing := []string{}
	if gate.Blocked {
		missing = append(missing, gate.MissingContracts...)
		missing = appendMissingConfigRolloutContract(missing, cmdbAgentRolloutDataArrivalEvidenceContract)
		missing = appendMissingConfigRolloutContract(missing, cmdbAgentRolloutReceiverEvidenceContract)
		writeDataArrivalRuntimeReadBlocked(c, rollout, query, missing)
		return true, nil
	}
	if query.RequestRef != "" && !dataArrivalRuntimeRequestRefMatchesRollout(query.RequestRef, rollout, gate) {
		writeDataArrivalRuntimeReadBlocked(c, rollout, query, []string{
			cmdbAgentRolloutDataArrivalRequestRefContract,
			cmdbAgentRolloutExecutionTaskContract,
			cmdbAgentRolloutTaskMatchContract,
		})
		return true, nil
	}
	items, err := store.ListFindXAgentDataArrivalEvidence()
	if err != nil {
		return true, err
	}
	matches := dataArrivalRuntimeEvidenceForRollout(items, rollout, query, gate)
	if len(matches) == 0 {
		writeDataArrivalRuntimeReadBlocked(c, rollout, query, []string{
			cmdbAgentRolloutDataArrivalEvidenceContract,
			cmdbAgentRolloutReceiverEvidenceContract,
			cmdbAgentRolloutDataArrivalRequestRefContract,
		})
		return true, nil
	}
	writeDataArrivalRuntimeReadEvidence(c, rollout, query, matches, gate)
	return true, nil
}

func rolloutIDFromReceiptRequestRef(requestRef string) string {
	task, ok, err := store.GetFindXAgentExecutionTask(requestRef)
	if err != nil || !ok {
		return ""
	}
	return firstNonEmpty(task.Metadata["source_rollout_id"], task.Metadata["config_rollout_id"], task.Metadata["rollout_ref"])
}

func dataArrivalRuntimeRequestRefMatchesRollout(requestRef string, rollout model.FindXAgentConfigRollout, gate configRolloutRuntimeReadGate) bool {
	for receipt, ref := range gate.ReceiptRefs {
		if ref != requestRef {
			continue
		}
		task, ok, err := store.GetFindXAgentExecutionTask(requestRef)
		if err != nil || !ok {
			return false
		}
		return cmdbConfigRolloutReceiptTaskMatches(rollout, task, receipt)
	}
	return false
}

func dataArrivalRuntimeEvidenceForRollout(items []model.FindXAgentDataArrivalEvidence, rollout model.FindXAgentConfigRollout, query dataArrivalRuntimeReadQuery, gate configRolloutRuntimeReadGate) []model.FindXAgentDataArrivalEvidence {
	out := []model.FindXAgentDataArrivalEvidence{}
	for _, item := range items {
		if !dataArrivalRuntimeEvidenceReceiverBacked(item) {
			continue
		}
		if !dataArrivalRuntimeEvidenceMatchesRollout(item, rollout, query, gate) {
			continue
		}
		out = append(out, item)
	}
	return out
}

func dataArrivalRuntimeEvidenceReceiverBacked(item model.FindXAgentDataArrivalEvidence) bool {
	return item.Status == model.FindXAgentDataArrivalStatusReported &&
		model.IsFindXAgentReceiverBackedDataArrivalKind(item.Kind) &&
		dataArrivalRuntimeReceiverRef(item) != ""
}

func dataArrivalRuntimeEvidenceMatchesRollout(item model.FindXAgentDataArrivalEvidence, rollout model.FindXAgentConfigRollout, query dataArrivalRuntimeReadQuery, gate configRolloutRuntimeReadGate) bool {
	requestRef := strings.TrimSpace(item.Metadata["request_ref"])
	if query.RequestRef != "" && requestRef != query.RequestRef {
		return false
	}
	if !dataArrivalRuntimeRequestRefMatchesRollout(requestRef, rollout, gate) {
		return false
	}
	hasRolloutRef := metadataValueMatchesAny(item.Metadata, []string{"source_rollout_id", "config_rollout_id", "rollout_ref"}, rollout.ID)
	if !hasRolloutRef && query.RequestRef == "" && requestRef == "" {
		return false
	}
	if plugin := strings.TrimSpace(item.Metadata["plugin_id"]); plugin != "" && plugin != rollout.PluginID {
		return false
	}
	return dataArrivalRuntimeEvidenceTargetMatches(item, rollout)
}

func dataArrivalRuntimeEvidenceTargetMatches(item model.FindXAgentDataArrivalEvidence, rollout model.FindXAgentConfigRollout) bool {
	if item.TargetID != "" && stringSlicesIntersect([]string{item.TargetID}, rollout.TargetIDs) {
		return true
	}
	if item.AgentID != "" && stringSlicesIntersect([]string{item.AgentID}, rollout.AgentIDs) {
		return true
	}
	if metadataValueMatchesAny(item.Metadata, []string{"cmdb_host_ref", "target_id"}, firstNonEmpty(rollout.TargetIDs...)) {
		return true
	}
	return metadataValueMatchesAny(item.Metadata, []string{"agent_ref", "agent_id", "source_agent"}, firstNonEmpty(rollout.AgentIDs...))
}

func dataArrivalRuntimeReceiverRef(item model.FindXAgentDataArrivalEvidence) string {
	for _, ref := range item.EvidenceRefs {
		if strings.HasPrefix(ref, "receiver:") {
			return ref
		}
	}
	return ""
}

func writeDataArrivalRuntimeReadBlocked(c *gin.Context, rollout model.FindXAgentConfigRollout, query dataArrivalRuntimeReadQuery, missing []string) {
	// Gate removed - return empty evidence list
	c.JSON(http.StatusOK, gin.H{
		"code":           http.StatusOK,
		"status":         "ok",
		"rollout_ref":    firstNonEmpty(rollout.ID, query.RolloutID),
		"request_ref":    query.RequestRef,
		"evidence_count": 0,
		"evidence":       []model.FindXAgentDataArrivalEvidence{},
	})
}

func writeDataArrivalRuntimeReadEvidence(c *gin.Context, rollout model.FindXAgentConfigRollout, query dataArrivalRuntimeReadQuery, items []model.FindXAgentDataArrivalEvidence, gate configRolloutRuntimeReadGate) {
	missing := append(cmdbAgentRolloutRuntimeExecutorGapContractsForItem(rollout),
		"cmdb_agent_rollout_delivery_receipt_contract",
		"cmdb_agent_rollout_effect_receipt_contract",
		"cmdb_agent_rollout_rollback_receipt_contract",
		"cmdb_agent_rollout_data_arrival_receipt_contract",
		"cmdb_agent_rollout_evidence_chain_contract",
	)
	c.JSON(http.StatusOK, gin.H{
		"code":              http.StatusOK,
		"status":            "blocked",
		"contract":          cmdbAgentRolloutDataArrivalReadContract,
		"missing_contracts": uniquePackageRepositoryBlockers(missing),
		"rollout_ref":       rollout.ID,
		"request_ref":       firstNonEmpty(query.RequestRef, firstLifecycleReceiptRef(gate.ReceiptRefs)),
		"evidence_count":    len(items),
		"evidence":          items,
		"safe_to_retry":     false,
		"receipt_contract": gin.H{
			"contract_id":       cmdbAgentRolloutDataArrivalReadContract,
			"required_receipts": []string{"receiver_evidence", "request_ref", "data_arrival_receipt"},
			"missing_contracts": uniquePackageRepositoryBlockers(missing),
			"status":            model.FindXAgentExecutionStateBlockedByContract,
		},
		"findx_audit_query": dataArrivalRuntimeAuditQuery(rollout.ID, firstNonEmpty(query.RequestRef, firstLifecycleReceiptRef(gate.ReceiptRefs))),
	})
}

func firstLifecycleReceiptRef(refs map[string]string) string {
	for _, key := range []string{"delivery", "effect", "rollback"} {
		if ref := strings.TrimSpace(refs[key]); ref != "" {
			return ref
		}
	}
	return ""
}

func dataArrivalRuntimeAuditQuery(rolloutID, requestRef string) gin.H {
	return gin.H{
		"source":        "findx_audit",
		"scope":         "cmdb",
		"resource_type": "findx_agent_config_rollout",
		"resource_id":   rolloutID,
		"action":        "findx_agent.data_arrival.evidence.read",
		"request_ref":   requestRef,
	}
}
