package handler

import (
	"net/http"
	"testing"

	"ai-workbench-api/internal/model"
	"ai-workbench-api/internal/store"
)

func TestContractMatrixSeededDashboardCRUDGapsAreBlocked(t *testing.T) {
	store.ResetContractMatrixForTest()
	r := contractMatrixTestRouter()
	for _, id := range dashboardCRUDContractGapIDs() {
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
				t.Fatalf("dashboard CRUD split gap response mismatch for %s: %#v", id, payload)
			}
			assertContractMatrixNoSensitiveLeak(t, w.Body.String())
			assertContractMatrixBlockedShape(t, w.Body.String())
			assertContractMatrixNoFakeRuntimeState(t, w.Body.String())
		})
	}
}

func TestContractMatrixSeededDashboardCRUDAggregateIsBlocked(t *testing.T) {
	store.ResetContractMatrixForTest()
	r := contractMatrixTestRouter()

	w := performContractMatrixRequest(t, r, http.MethodGet, "/contract-matrix/FX-CONTRACT-N9E-DASHBOARD-CRUD", nil)
	if w.Code != http.StatusConflict {
		t.Fatalf("dashboard CRUD aggregate should return BLOCKED_BY_CONTRACT 409, got %d body=%s", w.Code, w.Body.String())
	}
	var payload model.ContractMatrixBlockedResponse
	decodeContractMatrixResponse(t, w, &payload)
	if payload.Code != model.ContractBlockedByContractCode ||
		payload.ContractGapID != "FX-CONTRACT-N9E-DASHBOARD-CRUD" ||
		payload.Status != model.ContractStatusBlocked ||
		payload.SafeToRetry {
		t.Fatalf("dashboard CRUD aggregate response mismatch: %#v", payload)
	}
	assertContractMatrixNoSensitiveLeak(t, w.Body.String())
	assertContractMatrixBlockedShape(t, w.Body.String())
	assertContractMatrixNoFakeRuntimeState(t, w.Body.String())
}

func dashboardCRUDContractGapIDs() []string {
	return []string{
		"FX-CONTRACT-N9E-DASHBOARD-LIST-BY-BUSI-GROUP",
		"FX-CONTRACT-N9E-DASHBOARD-LIST-BY-BUSI-GROUPS",
		"FX-CONTRACT-N9E-DASHBOARD-CREATE",
		"FX-CONTRACT-N9E-DASHBOARD-DETAIL",
		"FX-CONTRACT-N9E-DASHBOARD-UPDATE-METADATA",
		"FX-CONTRACT-N9E-DASHBOARD-UPDATE-CONFIGS",
		"FX-CONTRACT-N9E-DASHBOARD-CLONE",
		"FX-CONTRACT-N9E-DASHBOARD-CLONES",
		"FX-CONTRACT-N9E-DASHBOARD-DELETE",
		"FX-CONTRACT-N9E-DASHBOARD-PURE-DETAIL",
		"FX-CONTRACT-N9E-DASHBOARD-NAMES",
	}
}
