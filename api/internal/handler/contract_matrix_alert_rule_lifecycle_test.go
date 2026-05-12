package handler

import (
	"net/http"
	"testing"

	"ai-workbench-api/internal/model"
	"ai-workbench-api/internal/store"
)

func TestContractMatrixSeededAlertRuleLifecycleAggregateIsBlocked(t *testing.T) {
	store.ResetContractMatrixForTest()
	r := contractMatrixTestRouter()

	w := performContractMatrixRequest(t, r, http.MethodGet, "/contract-matrix/FX-CONTRACT-N9E-ALERT-RULE-LIFECYCLE", nil)
	if w.Code != http.StatusConflict {
		t.Fatalf("alert rule lifecycle aggregate should return BLOCKED_BY_CONTRACT 409, got %d body=%s", w.Code, w.Body.String())
	}
	var payload model.ContractMatrixBlockedResponse
	decodeContractMatrixResponse(t, w, &payload)
	if payload.Code != model.ContractBlockedByContractCode ||
		payload.ContractGapID != "FX-CONTRACT-N9E-ALERT-RULE-LIFECYCLE" ||
		payload.Status != model.ContractStatusBlocked ||
		payload.SafeToRetry {
		t.Fatalf("alert rule lifecycle aggregate response mismatch: %#v", payload)
	}
	assertContractMatrixNoSensitiveLeak(t, w.Body.String())
	assertContractMatrixBlockedShape(t, w.Body.String())
	assertContractMatrixNoFakeRuntimeState(t, w.Body.String())
}

func TestContractMatrixSeededAlertRuleLifecycleGapsAreBlocked(t *testing.T) {
	store.ResetContractMatrixForTest()
	r := contractMatrixTestRouter()
	for _, tt := range []struct {
		id     string
		status string
	}{
		{id: "FX-CONTRACT-N9E-ALERT-RULE-DETAIL", status: model.ContractStatusMissingBackend},
		{id: "FX-CONTRACT-N9E-ALERT-RULE-PURE-DETAIL", status: model.ContractStatusMissingBackend},
		{id: "FX-CONTRACT-N9E-ALERT-RULE-CREATE", status: model.ContractStatusMissingBackend},
		{id: "FX-CONTRACT-N9E-ALERT-RULE-UPDATE", status: model.ContractStatusMissingBackend},
		{id: "FX-CONTRACT-N9E-ALERT-RULE-DELETE-BY-BUSI-GROUP", status: model.ContractStatusMissingBackend},
		{id: "FX-CONTRACT-N9E-ALERT-RULE-DELETE-BY-RULE-GROUP", status: model.ContractStatusMissingBackend},
		{id: "FX-CONTRACT-N9E-ALERT-RULE-IMPORT-JSON", status: model.ContractStatusMissingBackend},
		{id: "FX-CONTRACT-N9E-ALERT-RULE-IMPORT-PROM-RULE", status: model.ContractStatusMissingBackend},
		{id: "FX-CONTRACT-N9E-ALERT-RULE-BULK-FIELDS-UPDATE", status: model.ContractStatusMissingBackend},
		{id: "FX-CONTRACT-N9E-ALERT-RULE-ENABLE-DISABLE", status: model.ContractStatusMissingBackend},
		{id: "FX-CONTRACT-N9E-ALERT-RULE-STATUS-BATCH", status: model.ContractStatusMissingBackend},
		{id: "FX-CONTRACT-N9E-ALERT-RULE-CLONE-TO-HOSTS", status: model.ContractStatusMissingBackend},
		{id: "FX-CONTRACT-N9E-ALERT-RULE-CLONE-TO-BUSI-GROUPS", status: model.ContractStatusMissingBackend},
		{id: "FX-CONTRACT-N9E-ALERT-RULE-VALIDATE", status: model.ContractStatusMissingBackend},
		{id: "FX-CONTRACT-N9E-ALERT-RULE-ENABLE-TRYRUN", status: model.ContractStatusMissingDatasource},
		{id: "FX-CONTRACT-N9E-ALERT-RULE-CALLBACKS-LIST", status: model.ContractStatusMissingBackend},
		{id: "FX-CONTRACT-N9E-ALERT-RULE-TIMEZONES", status: model.ContractStatusMissingBackend},
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
				t.Fatalf("alert rule lifecycle gap response mismatch for %s: %#v", tt.id, payload)
			}
			assertContractMatrixNoSensitiveLeak(t, w.Body.String())
			assertContractMatrixBlockedShape(t, w.Body.String())
			assertContractMatrixNoFakeRuntimeState(t, w.Body.String())
		})
	}
}

func TestContractMatrixAlertRuleLifecycleStatusFilters(t *testing.T) {
	store.ResetContractMatrixForTest()
	r := contractMatrixTestRouter()

	missingBackend := performContractMatrixRequest(t, r, http.MethodGet, "/contract-matrix?domain=monitoring&status=missing_backend", nil)
	if missingBackend.Code != http.StatusOK {
		t.Fatalf("missing_backend filter should return 200, got %d body=%s", missingBackend.Code, missingBackend.Body.String())
	}
	for _, want := range []string{
		"FX-CONTRACT-N9E-ALERT-RULE-CREATE",
		"FX-CONTRACT-N9E-ALERT-RULE-UPDATE",
		"FX-CONTRACT-N9E-ALERT-RULE-IMPORT-JSON",
		"FX-CONTRACT-N9E-ALERT-RULE-CLONE-TO-HOSTS",
		"FX-CONTRACT-N9E-ALERT-RULE-TIMEZONES",
	} {
		if !contractMatrixResponseContains(missingBackend.Body.String(), want) {
			t.Fatalf("missing_backend filter missing %s: %s", want, missingBackend.Body.String())
		}
	}

	missingDatasource := performContractMatrixRequest(t, r, http.MethodGet, "/contract-matrix?domain=monitoring&status=missing_datasource", nil)
	if missingDatasource.Code != http.StatusOK {
		t.Fatalf("missing_datasource filter should return 200, got %d body=%s", missingDatasource.Code, missingDatasource.Body.String())
	}
	if !contractMatrixResponseContains(missingDatasource.Body.String(), "FX-CONTRACT-N9E-ALERT-RULE-ENABLE-TRYRUN") {
		t.Fatalf("missing_datasource filter missing enable tryrun gap: %s", missingDatasource.Body.String())
	}
	assertContractMatrixNoSensitiveLeak(t, missingBackend.Body.String())
	assertContractMatrixNoSensitiveLeak(t, missingDatasource.Body.String())
	assertContractMatrixNoFakeRuntimeState(t, missingBackend.Body.String())
	assertContractMatrixNoFakeRuntimeState(t, missingDatasource.Body.String())
}

func contractMatrixResponseContains(body, value string) bool {
	return len(body) >= len(value) && (body == value || containsSubstring(body, value))
}

func containsSubstring(body, value string) bool {
	for i := 0; i+len(value) <= len(body); i++ {
		if body[i:i+len(value)] == value {
			return true
		}
	}
	return false
}
