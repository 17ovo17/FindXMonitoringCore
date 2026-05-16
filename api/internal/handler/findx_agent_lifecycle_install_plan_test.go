package handler

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"ai-workbench-api/internal/model"
	"ai-workbench-api/internal/store"

	"github.com/gin-gonic/gin"
)

func TestFindXAgentInstallPlanPersistsBlockedRecord(t *testing.T) {
	gin.SetMode(gin.TestMode)
	resetAgentLifecycleRecordsForTest(t)
	body := strings.NewReader(`{
		"package_id":"agent-core",
		"os":"linux",
		"method":"linux-curl",
		"target_ids":["target-a"],
		"credential_ref":"<CREDENTIAL_REF>",
		"metadata":{"ticket":"CHG-1","token":"secret"}
	}`)
	w := performAgentLifecyclePost("/api/v1/findx-agents/install-plans", body, CreateFindXAgentInstallPlan)
	if w.Code != http.StatusConflict {
		t.Fatalf("install plan should stay blocked, got %d body=%s", w.Code, w.Body.String())
	}
	var payload struct {
		Data model.FindXAgentInstallPlan `json:"data"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &payload); err != nil {
		t.Fatalf("invalid response: %v", err)
	}
	if payload.Data.ID == "" || payload.Data.Status != "blocked" || !payload.Data.CredentialRefPresent {
		t.Fatalf("blocked plan should be persisted with safe metadata: %#v", payload.Data)
	}
	assertNoCredentialEcho(t, w.Body.String())
}

func TestFindXAgentInstallPlanListCoercesNonBlockedRecords(t *testing.T) {
	gin.SetMode(gin.TestMode)
	resetAgentLifecycleRecordsForTest(t)

	blocked, err := store.SaveFindXAgentInstallPlan(model.FindXAgentInstallPlan{
		PackageID: "agent-core",
		TargetIDs: []string{"target-blocked"},
		Status:    "blocked",
		Blocker:   "PENDING: executor not enabled",
		Metadata:  map[string]string{"evidence_chain_ref": "evidence-ref"},
	})
	if err != nil {
		t.Fatalf("save blocked install plan: %v", err)
	}
	nonBlocked, err := store.SaveFindXAgentInstallPlan(model.FindXAgentInstallPlan{
		PackageID: "agent-core",
		TargetIDs: []string{"target-queued"},
		Status:    "queued",
		Metadata:  map[string]string{"evidence_chain_ref": "queued-evidence-ref"},
	})
	if err != nil {
		t.Fatalf("save non-blocked install plan: %v", err)
	}
	if nonBlocked.Status != "blocked" || nonBlocked.Blocker == "" {
		t.Fatalf("store must coerce install plan to blocked, got %#v", nonBlocked)
	}

	w := performAgentLifecycleGet("/api/v1/findx-agents/install-plans", ListFindXAgentInstallPlans)
	if w.Code != http.StatusOK {
		t.Fatalf("install plan list should return 200, got %d body=%s", w.Code, w.Body.String())
	}
	body := w.Body.String()
	if !strings.Contains(body, blocked.ID) || !strings.Contains(body, nonBlocked.ID) {
		t.Fatalf("list should include store-coerced blocked install plans, body=%s", body)
	}
	assertNoInstallPlanSuccessStates(t, body)
	assertNoInstallPlanSensitiveEcho(t, body)
}

func TestFindXAgentInstallPlanDetailReturnsBlockedSafeRefs(t *testing.T) {
	gin.SetMode(gin.TestMode)
	resetAgentLifecycleRecordsForTest(t)
	post := performBlockedInstallPlanDetailPost(t)
	if post.Code != http.StatusConflict {
		t.Fatalf("blocked install plan should be created before detail query, code=%d body=%s", post.Code, post.Body.String())
	}
	payload := decodeBlockedInstallPlanResponse(t, post)
	w := performAgentLifecycleGet("/api/v1/findx-agents/install-plans?id="+payload.Data.ID, ListFindXAgentInstallPlans)
	if w.Code != http.StatusOK {
		t.Fatalf("install plan detail should return 200, got %d body=%s", w.Code, w.Body.String())
	}
	detail := decodeBlockedInstallPlanDetail(t, w)
	if detail.ID != payload.Data.ID || detail.Status != "blocked" || !strings.Contains(detail.Blocker, "PENDING") {
		t.Fatalf("detail should return the blocked install plan record, got %#v", detail)
	}
	for _, key := range []string{"package_repository_ref", "signature_ref", "checksum", "script_manifest_ref", "executor_ref", "provider_auth_ref"} {
		if detail.Metadata[key] == "" {
			t.Fatalf("detail should retain safe metadata ref %s, metadata=%#v", key, detail.Metadata)
		}
	}
	for _, key := range []string{"token", "cookie", "session", "dsn", "password", "private_key", "credential_ref"} {
		if detail.Metadata[key] != "" {
			t.Fatalf("detail must not expose sensitive metadata key %s, metadata=%#v", key, detail.Metadata)
		}
	}
	assertNoInstallPlanSuccessStates(t, w.Body.String())
	assertNoInstallPlanSensitiveEcho(t, w.Body.String())
}

func performBlockedInstallPlanDetailPost(t *testing.T) *httptest.ResponseRecorder {
	t.Helper()
	body := strings.NewReader(`{
		"package_id":"agent-core",
		"os":"linux",
		"method":"linux-curl",
		"target_ids":["target-a"],
		"credential_ref":"<CREDENTIAL_REF>",
		"metadata":{
			"package_repository_ref":"repo-ref",
			"signature_ref":"signature-ref",
			"checksum":"sha256:abc123",
			"script_manifest_ref":"script-ref",
			"executor_ref":"executor-ref",
			"provider_auth_ref":"vault-ref",
			"token":"secret-token",
			"cookie":"secret-cookie",
			"credential_ref":"credential-secret",
			"dsn":"mysql://user:pass@host/db",
			"private_key":"secret-private-key"
		}
	}`)
	return performAgentLifecyclePost("/api/v1/findx-agents/install-plans", body, CreateFindXAgentInstallPlan)
}

func decodeBlockedInstallPlanResponse(t *testing.T, w *httptest.ResponseRecorder) struct {
	Data model.FindXAgentInstallPlan `json:"data"`
} {
	t.Helper()
	var payload struct {
		Data model.FindXAgentInstallPlan `json:"data"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &payload); err != nil {
		t.Fatalf("invalid install plan response: %v body=%s", err, w.Body.String())
	}
	return payload
}

func decodeBlockedInstallPlanDetail(t *testing.T, w *httptest.ResponseRecorder) model.FindXAgentInstallPlan {
	t.Helper()
	var detail model.FindXAgentInstallPlan
	if err := json.Unmarshal(w.Body.Bytes(), &detail); err != nil {
		t.Fatalf("invalid detail response: %v body=%s", err, w.Body.String())
	}
	return detail
}

func TestFindXAgentInstallPlanDetailCoercesNonBlockedAndMissingReturns404(t *testing.T) {
	gin.SetMode(gin.TestMode)
	resetAgentLifecycleRecordsForTest(t)
	nonBlocked, err := store.SaveFindXAgentInstallPlan(model.FindXAgentInstallPlan{
		PackageID: "agent-core",
		TargetIDs: []string{"target-running"},
		Status:    "running",
		Metadata:  map[string]string{"evidence_chain_ref": "running-evidence-ref"},
	})
	if err != nil {
		t.Fatalf("save non-blocked install plan: %v", err)
	}

	w := performAgentLifecycleGet("/api/v1/findx-agents/install-plans?id="+nonBlocked.ID, ListFindXAgentInstallPlans)
	if w.Code != http.StatusOK || !strings.Contains(w.Body.String(), `"status":"blocked"`) || !strings.Contains(w.Body.String(), "executor not enabled") {
		t.Fatalf("coerced install plan detail should return blocked gate, got %d body=%s", w.Code, w.Body.String())
	}
	assertNoInstallPlanSuccessStates(t, w.Body.String())
	assertNoInstallPlanSensitiveEcho(t, w.Body.String())

	missing := performAgentLifecycleGet("/api/v1/findx-agents/install-plans?id=missing-plan", ListFindXAgentInstallPlans)
	if missing.Code != http.StatusNotFound {
		t.Fatalf("missing install plan detail should return 404, got %d body=%s", missing.Code, missing.Body.String())
	}
	assertNoInstallPlanSuccessStates(t, missing.Body.String())
	assertNoInstallPlanSensitiveEcho(t, missing.Body.String())
}
