package handler

import (
	"encoding/json"
	"net/http"
	"testing"

	"ai-workbench-api/internal/model"
	"ai-workbench-api/internal/store"
)

func TestContractMatrixSeededAlertSubscribeDetailsAreBlocked(t *testing.T) {
	store.ResetContractMatrixForTest()
	r := contractMatrixTestRouter()

	for _, tt := range alertSubscribeHandlerExpectations() {
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
				t.Fatalf("alert subscribe blocked response mismatch for %s: %#v", tt.id, payload)
			}
			assertContractMatrixNoSensitiveLeak(t, w.Body.String())
			assertContractMatrixBlockedShape(t, w.Body.String())
			assertContractMatrixNoFakeRuntimeState(t, w.Body.String())
		})
	}
}

func TestContractMatrixAlertSubscribeStatusDistribution(t *testing.T) {
	store.ResetContractMatrixForTest()
	r := contractMatrixTestRouter()

	resp := performContractMatrixRequest(t, r, http.MethodGet, "/contract-matrix?domain=monitoring", nil)
	if resp.Code != http.StatusOK {
		t.Fatalf("monitoring contract matrix list should return 200, got %d body=%s", resp.Code, resp.Body.String())
	}
	var payload struct {
		Items []model.ContractMatrixEntry `json:"items"`
	}
	decodeContractMatrixResponse(t, resp, &payload)

	counts := map[string]int{}
	for _, item := range payload.Items {
		if _, ok := alertSubscribeExpectationByID()[item.ID]; !ok {
			continue
		}
		counts[item.Status]++
		if item.SafeToRetry {
			t.Fatalf("%s must not be safe_to_retry: %#v", item.ID, item)
		}
	}
	if counts[model.ContractStatusBlocked] != 1 ||
		counts[model.ContractStatusMissingBackend] != 6 ||
		counts[model.ContractStatusMissingDatasource] != 1 {
		t.Fatalf("alert subscribe status distribution mismatch: %#v", counts)
	}
	assertContractMatrixNoSensitiveLeak(t, resp.Body.String())
	assertContractMatrixNoFakeRuntimeState(t, resp.Body.String())
}

func TestContractMatrixAlertSubscribeStatusFilters(t *testing.T) {
	store.ResetContractMatrixForTest()
	r := contractMatrixTestRouter()

	missingBackend := performContractMatrixRequest(t, r, http.MethodGet, "/contract-matrix?domain=monitoring&status=missing_backend", nil)
	if missingBackend.Code != http.StatusOK {
		t.Fatalf("missing_backend filter should return 200, got %d body=%s", missingBackend.Code, missingBackend.Body.String())
	}
	for _, want := range []string{
		"FX-CONTRACT-N9E-ALERT-SUBSCRIBE-LIST-BY-BUSI-GROUP",
		"FX-CONTRACT-N9E-ALERT-SUBSCRIBE-CREATE",
		"FX-CONTRACT-N9E-ALERT-SUBSCRIBE-UPDATE",
		"FX-CONTRACT-N9E-ALERT-SUBSCRIBE-DELETE",
	} {
		if !contractMatrixResponseContains(missingBackend.Body.String(), want) {
			t.Fatalf("missing_backend filter missing %s: %s", want, missingBackend.Body.String())
		}
	}
	if contractMatrixResponseContains(missingBackend.Body.String(), "FX-CONTRACT-N9E-ALERT-SUBSCRIBE-TRYRUN") {
		t.Fatalf("missing_backend filter included datasource gap: %s", missingBackend.Body.String())
	}

	missingDatasource := performContractMatrixRequest(t, r, http.MethodGet, "/contract-matrix?domain=monitoring&status=missing_datasource", nil)
	if missingDatasource.Code != http.StatusOK {
		t.Fatalf("missing_datasource filter should return 200, got %d body=%s", missingDatasource.Code, missingDatasource.Body.String())
	}
	if !contractMatrixResponseContains(missingDatasource.Body.String(), "FX-CONTRACT-N9E-ALERT-SUBSCRIBE-TRYRUN") {
		t.Fatalf("missing_datasource filter missing subscribe tryrun: %s", missingDatasource.Body.String())
	}
	for _, forbidden := range []string{
		"FX-CONTRACT-N9E-ALERT-SUBSCRIBE-CREATE",
		"FX-CONTRACT-N9E-ALERT-SUBSCRIBE-UPDATE",
		"FX-CONTRACT-N9E-ALERT-SUBSCRIBE-DELETE",
	} {
		if contractMatrixResponseContains(missingDatasource.Body.String(), forbidden) {
			t.Fatalf("missing_datasource filter included backend gap %s: %s", forbidden, missingDatasource.Body.String())
		}
	}
	assertContractMatrixNoSensitiveLeak(t, missingBackend.Body.String())
	assertContractMatrixNoSensitiveLeak(t, missingDatasource.Body.String())
	assertContractMatrixNoFakeRuntimeState(t, missingBackend.Body.String())
	assertContractMatrixNoFakeRuntimeState(t, missingDatasource.Body.String())
}

func TestContractMatrixAlertSubscribeListDoesNotExposeReadyRuntimeShape(t *testing.T) {
	store.ResetContractMatrixForTest()
	r := contractMatrixTestRouter()

	resp := performContractMatrixRequest(t, r, http.MethodGet, "/contract-matrix?domain=monitoring", nil)
	if resp.Code != http.StatusOK {
		t.Fatalf("monitoring contract matrix list should return 200, got %d body=%s", resp.Code, resp.Body.String())
	}
	var payload struct {
		Items []model.ContractMatrixEntry `json:"items"`
	}
	if err := json.Unmarshal(resp.Body.Bytes(), &payload); err != nil {
		t.Fatalf("decode list response: %v body=%s", err, resp.Body.String())
	}
	for _, item := range payload.Items {
		if _, ok := alertSubscribeExpectationByID()[item.ID]; !ok {
			continue
		}
		if item.Status == model.ContractStatusReady ||
			item.Handler != "" ||
			item.Backend != "" ||
			item.Datasource != "" ||
			item.Executor != "" ||
			len(item.EvidenceRefs) != 0 {
			t.Fatalf("%s must not expose fake runtime state or executable evidence: %#v", item.ID, item)
		}
	}
	assertContractMatrixNoSensitiveLeak(t, resp.Body.String())
	assertContractMatrixNoFakeRuntimeState(t, resp.Body.String())
}

func alertSubscribeHandlerExpectations() []struct {
	id     string
	status string
} {
	return []struct {
		id     string
		status string
	}{
		{id: "FX-CONTRACT-N9E-ALERT-SUBSCRIBE", status: model.ContractStatusBlocked},
		{id: "FX-CONTRACT-N9E-ALERT-SUBSCRIBE-LIST-BY-BUSI-GROUP", status: model.ContractStatusMissingBackend},
		{id: "FX-CONTRACT-N9E-ALERT-SUBSCRIBE-LIST-BY-BUSI-GROUPS", status: model.ContractStatusMissingBackend},
		{id: "FX-CONTRACT-N9E-ALERT-SUBSCRIBE-DETAIL", status: model.ContractStatusMissingBackend},
		{id: "FX-CONTRACT-N9E-ALERT-SUBSCRIBE-CREATE", status: model.ContractStatusMissingBackend},
		{id: "FX-CONTRACT-N9E-ALERT-SUBSCRIBE-UPDATE", status: model.ContractStatusMissingBackend},
		{id: "FX-CONTRACT-N9E-ALERT-SUBSCRIBE-DELETE", status: model.ContractStatusMissingBackend},
		{id: "FX-CONTRACT-N9E-ALERT-SUBSCRIBE-TRYRUN", status: model.ContractStatusMissingDatasource},
	}
}

func alertSubscribeExpectationByID() map[string]string {
	out := map[string]string{}
	for _, item := range alertSubscribeHandlerExpectations() {
		out[item.id] = item.status
	}
	return out
}
