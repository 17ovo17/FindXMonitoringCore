package handler

import (
	"net/http"
	"strings"
	"testing"

	"ai-workbench-api/internal/model"
	"ai-workbench-api/internal/store"
)

func TestContractMatrixSeededAlertRuleGroupsAggregateIsBlocked(t *testing.T) {
	store.ResetContractMatrixForTest()
	r := contractMatrixTestRouter()

	w := performContractMatrixRequest(t, r, http.MethodGet, "/contract-matrix/FX-CONTRACT-N9E-ALERT-RULE-GROUPS", nil)
	if w.Code != http.StatusConflict {
		t.Fatalf("alert rule groups aggregate should return BLOCKED_BY_CONTRACT 409, got %d body=%s", w.Code, w.Body.String())
	}
	var payload model.ContractMatrixBlockedResponse
	decodeContractMatrixResponse(t, w, &payload)
	if payload.Code != model.ContractBlockedByContractCode ||
		payload.ContractGapID != "FX-CONTRACT-N9E-ALERT-RULE-GROUPS" ||
		payload.Status != model.ContractStatusBlocked ||
		payload.SafeToRetry {
		t.Fatalf("alert rule groups aggregate response mismatch: %#v", payload)
	}
	assertContractMatrixNoSensitiveLeak(t, w.Body.String())
	assertContractMatrixBlockedShape(t, w.Body.String())
	assertContractMatrixNoFakeRuntimeState(t, w.Body.String())
}

func TestContractMatrixSeededAlertRuleGroupGapsAreBlocked(t *testing.T) {
	store.ResetContractMatrixForTest()
	r := contractMatrixTestRouter()
	for _, tt := range []struct {
		id     string
		status string
	}{
		{id: "FX-CONTRACT-N9E-ALERT-RULE-GROUP-LIST", status: model.ContractStatusMissingBackend},
		{id: "FX-CONTRACT-N9E-ALERT-RULE-GROUP-CREATE", status: model.ContractStatusMissingBackend},
		{id: "FX-CONTRACT-N9E-ALERT-RULE-GROUP-DETAIL", status: model.ContractStatusMissingBackend},
		{id: "FX-CONTRACT-N9E-ALERT-RULE-GROUP-UPDATE", status: model.ContractStatusMissingBackend},
		{id: "FX-CONTRACT-N9E-ALERT-RULE-GROUP-DELETE", status: model.ContractStatusMissingBackend},
		{id: "FX-CONTRACT-N9E-ALERT-RULE-GROUP-RULES-LIST", status: model.ContractStatusMissingBackend},
		{id: "FX-CONTRACT-N9E-ALERT-RULE-GROUPS-MULTI-RULES-LIST", status: model.ContractStatusMissingBackend},
		{id: "FX-CONTRACT-N9E-ALERT-RULE-GROUP-FAVORITES-LIST", status: model.ContractStatusMissingBackend},
		{id: "FX-CONTRACT-N9E-ALERT-RULE-GROUP-FAVORITE-ADD", status: model.ContractStatusMissingBackend},
		{id: "FX-CONTRACT-N9E-ALERT-RULE-GROUP-FAVORITE-DELETE", status: model.ContractStatusMissingBackend},
	} {
		t.Run(tt.id, func(t *testing.T) {
			w := performContractMatrixRequest(t, r, http.MethodGet, "/contract-matrix/"+tt.id, nil)
			if w.Code != http.StatusConflict {
				t.Fatalf("%s should return BLOCKED_BY_CONTRACT 409, got %d body=%s", tt.id, w.Code, w.Body.String())
			}
			var payload model.ContractMatrixBlockedResponse
			decodeContractMatrixResponse(t, w, &payload)
			if payload.Code != model.ContractBlockedByContractCode ||
				payload.ContractGapID != tt.id ||
				payload.Status != tt.status ||
				payload.SafeToRetry {
				t.Fatalf("alert rule group gap response mismatch for %s: %#v", tt.id, payload)
			}
			assertContractMatrixNoSensitiveLeak(t, w.Body.String())
			assertContractMatrixBlockedShape(t, w.Body.String())
			assertContractMatrixNoFakeRuntimeState(t, w.Body.String())
		})
	}
}

func TestContractMatrixAlertRuleGroupsDoNotLeakRuntimeState(t *testing.T) {
	store.ResetContractMatrixForTest()
	r := contractMatrixTestRouter()

	for _, id := range []string{
		"FX-CONTRACT-N9E-ALERT-RULE-GROUPS",
		"FX-CONTRACT-N9E-ALERT-RULE-GROUP-LIST",
		"FX-CONTRACT-N9E-ALERT-RULE-GROUP-RULES-LIST",
	} {
		t.Run(id, func(t *testing.T) {
			w := performContractMatrixRequest(t, r, http.MethodGet, "/contract-matrix/"+id, nil)
			body := w.Body.String()
			if strings.Contains(strings.ToLower(body), "running") || strings.Contains(strings.ToLower(body), "installed") {
				t.Fatalf("%s must not expose fake runtime state: %s", id, body)
			}
			assertContractMatrixNoSensitiveLeak(t, body)
		})
	}
}
