package handler

import (
	"encoding/json"
	"net/http"
	"testing"

	"ai-workbench-api/internal/model"
	"ai-workbench-api/internal/store"
)

func TestContractMatrixSeededAlertEventLifecycleDetailsAreBlocked(t *testing.T) {
	store.ResetContractMatrixForTest()
	r := contractMatrixTestRouter()

	for _, tt := range alertEventLifecycleHandlerExpectations() {
		t.Run(tt.id, func(t *testing.T) {
			w := performContractMatrixRequest(t, r, http.MethodGet, "/contract-matrix/"+tt.id, nil)
			if w.Code != http.StatusConflict {
				t.Fatalf("%s should return PENDING 409, got %d body=%s", tt.id, w.Code, w.Body.String())
			}
			var payload model.ContractMatrixBlockedResponse
			decodeContractMatrixResponse(t, w, &payload)
			if payload.Code != model.ContractBlockedByContractCode ||
				payload.ContractGapID != tt.id ||
				payload.Status != tt.status ||
				payload.SafeToRetry {
				t.Fatalf("alert event lifecycle blocked response mismatch for %s: %#v", tt.id, payload)
			}
			assertContractMatrixNoSensitiveLeak(t, w.Body.String())
			assertContractMatrixBlockedShape(t, w.Body.String())
			assertContractMatrixNoFakeRuntimeState(t, w.Body.String())
		})
	}
}

func TestContractMatrixAlertEventLifecycleStatusDistribution(t *testing.T) {
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
		if _, ok := alertEventLifecycleExpectationByID()[item.ID]; !ok {
			continue
		}
		counts[item.Status]++
		if item.SafeToRetry {
			t.Fatalf("%s must not be safe_to_retry: %#v", item.ID, item)
		}
	}
	if counts[model.ContractStatusBlocked] != 1 ||
		counts[model.ContractStatusMissingBackend] != 8 ||
		counts[model.ContractStatusMissingDatasource] != 2 {
		t.Fatalf("alert event lifecycle status distribution mismatch: %#v", counts)
	}
	assertContractMatrixNoSensitiveLeak(t, resp.Body.String())
	assertContractMatrixNoFakeRuntimeState(t, resp.Body.String())
}

func TestContractMatrixAlertEventLifecycleStatusFilters(t *testing.T) {
	store.ResetContractMatrixForTest()
	r := contractMatrixTestRouter()

	missingBackend := performContractMatrixRequest(t, r, http.MethodGet, "/contract-matrix?domain=monitoring&status=missing_backend", nil)
	if missingBackend.Code != http.StatusOK {
		t.Fatalf("missing_backend filter should return 200, got %d body=%s", missingBackend.Code, missingBackend.Body.String())
	}
	for _, want := range []string{
		"FX-CONTRACT-N9E-ALERT-EVENT-CURRENT-LIST",
		"FX-CONTRACT-N9E-ALERT-EVENT-DETAIL",
		"FX-CONTRACT-N9E-ALERT-EVENT-CURRENT-DELETE",
		"FX-CONTRACT-N9E-ALERT-EVENT-HISTORY-CLEANUP",
		"FX-CONTRACT-N9E-ALERT-EVENT-NOTIFY-RECORDS",
	} {
		if !contractMatrixResponseContains(missingBackend.Body.String(), want) {
			t.Fatalf("missing_backend filter missing %s: %s", want, missingBackend.Body.String())
		}
	}
	for _, forbidden := range []string{
		"FX-CONTRACT-N9E-ALERT-EVENT-CURRENT-CARD-LIST",
		"FX-CONTRACT-N9E-ALERT-EVENT-CURRENT-CARD-DETAILS",
	} {
		if contractMatrixResponseContains(missingBackend.Body.String(), forbidden) {
			t.Fatalf("missing_backend filter included datasource gap %s: %s", forbidden, missingBackend.Body.String())
		}
	}

	missingDatasource := performContractMatrixRequest(t, r, http.MethodGet, "/contract-matrix?domain=monitoring&status=missing_datasource", nil)
	if missingDatasource.Code != http.StatusOK {
		t.Fatalf("missing_datasource filter should return 200, got %d body=%s", missingDatasource.Code, missingDatasource.Body.String())
	}
	for _, want := range []string{
		"FX-CONTRACT-N9E-ALERT-EVENT-CURRENT-CARD-LIST",
		"FX-CONTRACT-N9E-ALERT-EVENT-CURRENT-CARD-DETAILS",
	} {
		if !contractMatrixResponseContains(missingDatasource.Body.String(), want) {
			t.Fatalf("missing_datasource filter missing %s: %s", want, missingDatasource.Body.String())
		}
	}
	for _, forbidden := range []string{
		"FX-CONTRACT-N9E-ALERT-EVENT-CURRENT-LIST",
		"FX-CONTRACT-N9E-ALERT-EVENT-DETAIL",
		"FX-CONTRACT-N9E-ALERT-EVENT-HISTORY-LIST",
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

func TestContractMatrixAlertEventLifecycleListDoesNotExposeReadyRuntimeShape(t *testing.T) {
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
		if _, ok := alertEventLifecycleExpectationByID()[item.ID]; !ok {
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

func TestContractMatrixAlertEventLifecycleDoesNotRegressReadySingleQuery(t *testing.T) {
	store.ResetContractMatrixForTest()
	r := contractMatrixTestRouter()

	w := performContractMatrixRequest(t, r, http.MethodGet, "/contract-matrix/FX-CONTRACT-FINDX-PROMETHEUS-SINGLE-QUERY", nil)
	if w.Code != http.StatusOK {
		t.Fatalf("single query adapter should stay ready, got %d body=%s", w.Code, w.Body.String())
	}
	var item model.ContractMatrixEntry
	decodeContractMatrixResponse(t, w, &item)
	if item.Status != model.ContractStatusReady || !item.SafeToRetry ||
		item.Handler == "" || item.Backend == "" || item.Datasource == "" || item.Executor == "" ||
		len(item.EvidenceRefs) == 0 {
		t.Fatalf("single query ready contract regressed: %#v", item)
	}
	assertContractMatrixNoSensitiveLeak(t, w.Body.String())
}

func alertEventLifecycleHandlerExpectations() []struct {
	id     string
	status string
} {
	return []struct {
		id     string
		status string
	}{
		{id: "FX-CONTRACT-N9E-ALERT-EVENT-LIFECYCLE", status: model.ContractStatusBlocked},
		{id: "FX-CONTRACT-N9E-ALERT-EVENT-CURRENT-LIST", status: model.ContractStatusMissingBackend},
		{id: "FX-CONTRACT-N9E-ALERT-EVENT-CURRENT-DATASOURCES", status: model.ContractStatusMissingBackend},
		{id: "FX-CONTRACT-N9E-ALERT-EVENT-CURRENT-CARD-LIST", status: model.ContractStatusMissingDatasource},
		{id: "FX-CONTRACT-N9E-ALERT-EVENT-CURRENT-CARD-DETAILS", status: model.ContractStatusMissingDatasource},
		{id: "FX-CONTRACT-N9E-ALERT-EVENT-DETAIL", status: model.ContractStatusMissingBackend},
		{id: "FX-CONTRACT-N9E-ALERT-EVENT-CURRENT-DELETE", status: model.ContractStatusMissingBackend},
		{id: "FX-CONTRACT-N9E-ALERT-EVENT-HISTORY-LIST", status: model.ContractStatusMissingBackend},
		{id: "FX-CONTRACT-N9E-ALERT-EVENT-HISTORY-BY-IDS", status: model.ContractStatusMissingBackend},
		{id: "FX-CONTRACT-N9E-ALERT-EVENT-HISTORY-CLEANUP", status: model.ContractStatusMissingBackend},
		{id: "FX-CONTRACT-N9E-ALERT-EVENT-NOTIFY-RECORDS", status: model.ContractStatusMissingBackend},
	}
}

func alertEventLifecycleExpectationByID() map[string]string {
	out := map[string]string{}
	for _, item := range alertEventLifecycleHandlerExpectations() {
		out[item.id] = item.status
	}
	return out
}
