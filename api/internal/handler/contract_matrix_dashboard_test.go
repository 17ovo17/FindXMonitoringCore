package handler

import (
	"net/http"
	"testing"

	"ai-workbench-api/internal/model"
	"ai-workbench-api/internal/store"
)

func TestContractMatrixSeededDashboardAnnotationsSplitGapsAreBlocked(t *testing.T) {
	store.ResetContractMatrixForTest()
	r := contractMatrixTestRouter()
	for _, id := range []string{
		"FX-CONTRACT-N9E-DASHBOARD-ANNOTATIONS-LIST",
		"FX-CONTRACT-N9E-DASHBOARD-ANNOTATIONS-CREATE",
		"FX-CONTRACT-N9E-DASHBOARD-ANNOTATIONS-UPDATE",
		"FX-CONTRACT-N9E-DASHBOARD-ANNOTATIONS-DELETE",
	} {
		t.Run(id, func(t *testing.T) {
			w := performContractMatrixRequest(t, r, http.MethodGet, "/contract-matrix/"+id, nil)
			if w.Code != http.StatusConflict {
				t.Fatalf("%s should return PENDING 409, got %d body=%s", id, w.Code, w.Body.String())
			}
			var payload model.ContractMatrixBlockedResponse
			decodeContractMatrixResponse(t, w, &payload)
			if payload.Code != model.ContractBlockedByContractCode ||
				payload.ContractGapID != id ||
				payload.Status != model.ContractStatusMissingBackend ||
				payload.SafeToRetry {
				t.Fatalf("dashboard annotations split gap response mismatch for %s: %#v", id, payload)
			}
			assertContractMatrixNoSensitiveLeak(t, w.Body.String())
			assertContractMatrixBlockedShape(t, w.Body.String())
		})
	}
}

func TestContractMatrixSeededDashboardPublicExportMigrateGapsAreBlocked(t *testing.T) {
	store.ResetContractMatrixForTest()
	r := contractMatrixTestRouter()
	for _, id := range []string{
		"FX-CONTRACT-N9E-DASHBOARD-PUBLIC-LIST",
		"FX-CONTRACT-N9E-DASHBOARD-PUBLIC-UPDATE",
		"FX-CONTRACT-N9E-DASHBOARD-EXPORT",
		"FX-CONTRACT-N9E-DASHBOARD-MIGRATE",
	} {
		t.Run(id, func(t *testing.T) {
			w := performContractMatrixRequest(t, r, http.MethodGet, "/contract-matrix/"+id, nil)
			if w.Code != http.StatusConflict {
				t.Fatalf("%s should return PENDING 409, got %d body=%s", id, w.Code, w.Body.String())
			}
			var payload model.ContractMatrixBlockedResponse
			decodeContractMatrixResponse(t, w, &payload)
			if payload.Code != model.ContractBlockedByContractCode ||
				payload.ContractGapID != id ||
				payload.Status != model.ContractStatusMissingBackend ||
				payload.SafeToRetry {
				t.Fatalf("dashboard public/export/migrate split gap response mismatch for %s: %#v", id, payload)
			}
			assertContractMatrixNoSensitiveLeak(t, w.Body.String())
			assertContractMatrixBlockedShape(t, w.Body.String())
			assertContractMatrixNoFakeRuntimeState(t, w.Body.String())
		})
	}
}
