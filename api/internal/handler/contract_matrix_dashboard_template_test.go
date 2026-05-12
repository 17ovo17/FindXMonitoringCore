package handler

import (
	"net/http"
	"testing"

	"ai-workbench-api/internal/model"
	"ai-workbench-api/internal/store"
)

func TestContractMatrixSeededDashboardTemplateBuiltinGapsAreBlocked(t *testing.T) {
	store.ResetContractMatrixForTest()
	r := contractMatrixTestRouter()
	for _, id := range []string{
		"FX-CONTRACT-N9E-TEMPLATE-CENTER-BUILTIN-BOARD-DETAIL",
		"FX-CONTRACT-N9E-TEMPLATE-CENTER-DASHBOARD-IMPORT-BATCH-RESULT",
		"FX-CONTRACT-N9E-TEMPLATE-CENTER-DASHBOARD-IMPORT-CONFLICT-ROLLBACK",
		"FX-CONTRACT-N9E-TEMPLATE-CENTER-DOCUMENT-DRAWER",
	} {
		t.Run(id, func(t *testing.T) {
			w := performContractMatrixRequest(t, r, http.MethodGet, "/contract-matrix/"+id, nil)
			if w.Code != http.StatusConflict {
				t.Fatalf("%s should return BLOCKED_BY_CONTRACT 409, got %d body=%s", id, w.Code, w.Body.String())
			}
			var payload model.ContractMatrixBlockedResponse
			decodeContractMatrixResponse(t, w, &payload)
			if payload.Code != model.ContractBlockedByContractCode ||
				payload.ContractGapID != id ||
				payload.Status != model.ContractStatusMissingBackend ||
				payload.SafeToRetry {
				t.Fatalf("dashboard template builtin split gap response mismatch for %s: %#v", id, payload)
			}
			assertContractMatrixNoSensitiveLeak(t, w.Body.String())
			assertContractMatrixBlockedShape(t, w.Body.String())
			assertContractMatrixNoFakeRuntimeState(t, w.Body.String())
		})
	}
}

func TestContractMatrixSeededDashboardTemplateAggregateStaysBlocked(t *testing.T) {
	store.ResetContractMatrixForTest()
	r := contractMatrixTestRouter()

	w := performContractMatrixRequest(t, r, http.MethodGet, "/contract-matrix/FX-CONTRACT-N9E-TEMPLATE-CENTER-DASHBOARD-IMPORT", nil)
	if w.Code != http.StatusConflict {
		t.Fatalf("dashboard template import aggregate should return BLOCKED_BY_CONTRACT 409, got %d body=%s", w.Code, w.Body.String())
	}
	var payload model.ContractMatrixBlockedResponse
	decodeContractMatrixResponse(t, w, &payload)
	if payload.Code != model.ContractBlockedByContractCode ||
		payload.ContractGapID != "FX-CONTRACT-N9E-TEMPLATE-CENTER-DASHBOARD-IMPORT" ||
		payload.Status != model.ContractStatusBlocked ||
		payload.SafeToRetry {
		t.Fatalf("dashboard template import aggregate response mismatch: %#v", payload)
	}
	assertContractMatrixNoSensitiveLeak(t, w.Body.String())
	assertContractMatrixBlockedShape(t, w.Body.String())
	assertContractMatrixNoFakeRuntimeState(t, w.Body.String())
}
