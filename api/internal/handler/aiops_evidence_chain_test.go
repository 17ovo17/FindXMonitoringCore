package handler

import (
	"encoding/json"
	"net/http"
	"strings"
	"testing"
	"time"

	"ai-workbench-api/internal/model"
	"ai-workbench-api/internal/store"

	"github.com/gin-gonic/gin"
)

func TestAIOpsEvidenceChainEmptyReturnsBlockedContracts(t *testing.T) {
	gin.SetMode(gin.TestMode)
	resetAgentLifecycleRecordsForTest(t)

	w := performAgentLifecycleGet("/api/v1/aiops/evidence-chain", AIOpsEvidenceChain)
	if w.Code != http.StatusOK {
		t.Fatalf("evidence chain should return 200, got %d body=%s", w.Code, w.Body.String())
	}
	payload := decodeAIOpsEvidenceChainResponse(t, w.Body.Bytes())
	if payload.Data.Summary.ReportedItems != 0 || payload.Data.Summary.MissingItems == 0 {
		t.Fatalf("empty evidence chain should report contract gaps only: %#v", payload.Data.Summary)
	}
	assertEvidenceChainContains(t, payload.Data.Items, "logs", "contract_gap", model.FindXAgentDataArrivalStatusBlocked)
	assertEvidenceChainContains(t, payload.Data.Items, "ai_conclusion", "contract_gap", model.FindXAgentDataArrivalStatusBlocked)
	assertNoEvidenceChainLeaks(t, w.Body.String())
}

func TestAIOpsEvidenceChainAggregatesReceiverEvidenceAndBlockedRecords(t *testing.T) {
	gin.SetMode(gin.TestMode)
	resetAgentLifecycleRecordsForTest(t)

	seedEvidenceChainReceiverAndBlockedRecords(t)

	w := performAgentLifecycleGet("/api/v1/aiops/evidence-chain", AIOpsEvidenceChain)
	if w.Code != http.StatusOK {
		t.Fatalf("evidence chain should return 200, got %d body=%s", w.Code, w.Body.String())
	}
	payload := decodeAIOpsEvidenceChainResponse(t, w.Body.Bytes())
	assertEvidenceChainContains(t, payload.Data.Items, "logs", "receiver_evidence", model.FindXAgentDataArrivalStatusReported)
	assertEvidenceChainContains(t, payload.Data.Items, "", "config_rollout", model.FindXAgentDataArrivalStatusBlocked)
	assertEvidenceChainContains(t, payload.Data.Items, "", "agent_lifecycle_task", model.FindXAgentDataArrivalStatusBlocked)
	if payload.Data.Summary.ReportedItems == 0 || payload.Data.Summary.BlockedItems == 0 {
		t.Fatalf("evidence chain should expose reported and blocked facts: %#v", payload.Data.Summary)
	}
	assertNoEvidenceChainLeaks(t, w.Body.String())
}

func TestAIOpsEvidenceChainRedactsUnsafeIDAndBlocker(t *testing.T) {
	gin.SetMode(gin.TestMode)
	resetAgentLifecycleRecordsForTest(t)

	longID, longBlocker := seedEvidenceChainUnsafeRecords(t)
	w := performAgentLifecycleGet("/api/v1/aiops/evidence-chain", AIOpsEvidenceChain)
	if w.Code != http.StatusOK {
		t.Fatalf("evidence chain should return 200, got %d body=%s", w.Code, w.Body.String())
	}
	body := w.Body.String()
	assertEvidenceChainUnsafeValuesRedacted(t, body, longID, longBlocker)
	payload := decodeAIOpsEvidenceChainResponse(t, w.Body.Bytes())
	assertEvidenceChainRedactedItems(t, payload.Data.Items)
	assertEvidenceChainBlockersAreSafe(t, payload.Data.Blockers)
	assertNoEvidenceChainLeaks(t, body)
}

func seedEvidenceChainReceiverAndBlockedRecords(t *testing.T) {
	t.Helper()
	now := time.Now()
	if _, err := store.SaveFindXAgentDataArrivalEvidence(model.FindXAgentDataArrivalEvidence{
		ID:           "ev-logs",
		Kind:         model.FindXAgentDataArrivalKindLogs,
		AgentID:      "agent-a",
		TargetID:     "target-a",
		Status:       model.FindXAgentDataArrivalStatusReported,
		EvidenceRefs: []string{"receiver:/findx-agent/logs-compatible", "token-should-drop"},
		Metadata: map[string]string{
			"receiver":       "findx-agent",
			"credential_ref": "credential-ref",
			"authorization":  "Bearer secret",
		},
		UpdatedAt: now,
	}); err != nil {
		t.Fatalf("save data arrival evidence: %v", err)
	}
	seedEvidenceChainBlockedLifecycleRecords(t)
}

func seedEvidenceChainBlockedLifecycleRecords(t *testing.T) {
	t.Helper()
	if _, err := store.SaveFindXAgentConfigRollout(model.FindXAgentConfigRollout{
		ID:       "rollout-a",
		Status:   "running",
		Blocker:  "BLOCKED_BY_CONTRACT: executor not enabled",
		Metadata: map[string]string{"evidence_chain_ref": "chain-ref", "password": "secret", "state": "succeeded"},
	}); err != nil {
		t.Fatalf("save config rollout: %v", err)
	}
	if _, err := store.SaveFindXAgentExecutionTask(model.FindXAgentExecutionTask{
		ID:       "task-a",
		Action:   "uninstall",
		Status:   "queued",
		Blocker:  "BLOCKED_BY_CONTRACT: executor not enabled",
		Metadata: map[string]string{"audit_ref": "audit-ref", "cookie": "secret-cookie", "phase": "success"},
	}); err != nil {
		t.Fatalf("save task: %v", err)
	}
}

func seedEvidenceChainUnsafeRecords(t *testing.T) (string, string) {
	t.Helper()
	longID := strings.Repeat("evidence-", 20)
	longBlocker := "BLOCKED_BY_CONTRACT: " + strings.Repeat("missing-contract-", 20)
	if _, err := store.SaveFindXAgentInstallPlan(model.FindXAgentInstallPlan{
		ID:           "token-control-\x00id",
		Blocker:      "BLOCKED_BY_CONTRACT: bearer token blocked",
		EvidenceRefs: []string{"safe-ref"},
	}); err != nil {
		t.Fatalf("save install plan: %v", err)
	}
	if _, err := store.SaveFindXAgentConfigRollout(model.FindXAgentConfigRollout{
		ID:      "fake-state-succeeded",
		Blocker: "queued",
	}); err != nil {
		t.Fatalf("save config rollout: %v", err)
	}
	if _, err := store.SaveFindXAgentExecutionTask(model.FindXAgentExecutionTask{
		ID:      longID,
		Action:  "install",
		Blocker: "BLOCKED_BY_CONTRACT:\x00 executor not enabled",
	}); err != nil {
		t.Fatalf("save execution task: %v", err)
	}
	if _, err := store.SaveFindXAgentDataArrivalEvidence(model.FindXAgentDataArrivalEvidence{
		ID:      "ev-control-\x00id",
		Kind:    model.FindXAgentDataArrivalKindLogs,
		Status:  model.FindXAgentDataArrivalStatusError,
		Blocker: longBlocker,
	}); err != nil {
		t.Fatalf("save data arrival evidence: %v", err)
	}
	return longID, longBlocker
}

func assertEvidenceChainUnsafeValuesRedacted(t *testing.T, body, longID, longBlocker string) {
	t.Helper()
	for _, leaked := range []string{"token-control", "fake-state-succeeded", longID, "bearer token", `"blocker":"queued"`, longBlocker, "\u0000"} {
		if strings.Contains(body, leaked) {
			t.Fatalf("evidence chain response leaked unsafe value %q: %s", leaked, body)
		}
	}
}

type aiopsEvidenceChainEnvelope struct {
	Code int                        `json:"code"`
	Data aiopsEvidenceChainResponse `json:"data"`
}

func decodeAIOpsEvidenceChainResponse(t *testing.T, raw []byte) aiopsEvidenceChainEnvelope {
	t.Helper()
	var payload aiopsEvidenceChainEnvelope
	if err := json.Unmarshal(raw, &payload); err != nil {
		t.Fatalf("invalid evidence chain response: %v body=%s", err, string(raw))
	}
	return payload
}

func assertEvidenceChainContains(t *testing.T, items []aiopsEvidenceChainItem, kind, sourceType, status string) {
	t.Helper()
	for _, item := range items {
		if kind != "" && item.Kind != kind {
			continue
		}
		if sourceType != "" && item.SourceType != sourceType {
			continue
		}
		if item.Status == status {
			return
		}
	}
	t.Fatalf("missing evidence item kind=%q source=%q status=%q in %#v", kind, sourceType, status, items)
}

func assertNoEvidenceChainLeaks(t *testing.T, body string) {
	t.Helper()
	for _, forbidden := range []string{
		"credential_ref", "credential-ref", "secret", "token-should-drop",
		"authorization", "Bearer", "cookie", "password",
		`"status":"queued"`, `"status":"running"`, `"status":"succeeded"`,
		`"status":"success"`, `"status":"applied"`, "rolled-back",
	} {
		if strings.Contains(body, forbidden) {
			t.Fatalf("evidence chain response leaked sensitive or fake state %q: %s", forbidden, body)
		}
	}
}

func assertEvidenceChainRedactedItems(t *testing.T, items []aiopsEvidenceChainItem) {
	t.Helper()
	const maxSafeIDLen = 128
	var redacted int
	for _, item := range items {
		if strings.HasPrefix(item.ID, "redacted-") {
			redacted++
		}
		if strings.ContainsRune(item.ID, '\x00') || len([]rune(item.ID)) > maxSafeIDLen {
			t.Fatalf("unsafe evidence chain id was exposed: %#v", item)
		}
	}
	if redacted < 4 {
		t.Fatalf("expected unsafe ids to be redacted, got %d redacted items in %#v", redacted, items)
	}
}

func assertEvidenceChainBlockersAreSafe(t *testing.T, blockers []aiopsEvidenceChainBlocker) {
	t.Helper()
	const maxSafeBlockerLen = 160
	const maxSafeIDLen = 128
	for _, blocker := range blockers {
		if strings.ContainsRune(blocker.Reason, '\x00') {
			t.Fatalf("blocker contains control rune: %#v", blocker)
		}
		if len([]rune(blocker.Reason)) > maxSafeBlockerLen {
			t.Fatalf("blocker exceeds length limit: %#v", blocker)
		}
		for _, id := range blocker.Items {
			if strings.Contains(id, "token") || strings.Contains(id, "succeeded") || len([]rune(id)) > maxSafeIDLen {
				t.Fatalf("blocker item leaked unsafe id: %#v", blocker)
			}
		}
	}
}
