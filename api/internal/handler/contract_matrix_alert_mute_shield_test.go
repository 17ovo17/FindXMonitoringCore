package handler

import (
	"encoding/json"
	"net/http"
	"testing"

	"ai-workbench-api/internal/model"
	"ai-workbench-api/internal/store"
)

func TestContractMatrixSeededAlertMuteShieldDetailsAreBlocked(t *testing.T) {
	store.ResetContractMatrixForTest()
	r := contractMatrixTestRouter()

	for _, tt := range alertMuteShieldHandlerExpectations() {
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
				t.Fatalf("alert mute shield blocked response mismatch for %s: %#v", tt.id, payload)
			}
			assertContractMatrixNoSensitiveLeak(t, w.Body.String())
			assertContractMatrixBlockedShape(t, w.Body.String())
			assertContractMatrixNoFakeRuntimeState(t, w.Body.String())
		})
	}
}

func TestContractMatrixAlertMuteShieldStatusDistribution(t *testing.T) {
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
		if _, ok := alertMuteShieldExpectationByID()[item.ID]; !ok {
			continue
		}
		counts[item.Status]++
		if item.SafeToRetry {
			t.Fatalf("%s must not be safe_to_retry: %#v", item.ID, item)
		}
	}
	if counts[model.ContractStatusBlocked] != 1 ||
		counts[model.ContractStatusMissingBackend] != 7 ||
		counts[model.ContractStatusMissingDatasource] != 2 {
		t.Fatalf("alert mute shield status distribution mismatch: %#v", counts)
	}
	assertContractMatrixNoSensitiveLeak(t, resp.Body.String())
	assertContractMatrixNoFakeRuntimeState(t, resp.Body.String())
}

func TestContractMatrixAlertMuteShieldStatusFilters(t *testing.T) {
	store.ResetContractMatrixForTest()
	r := contractMatrixTestRouter()

	missingBackend := performContractMatrixRequest(t, r, http.MethodGet, "/contract-matrix?domain=monitoring&status=missing_backend", nil)
	if missingBackend.Code != http.StatusOK {
		t.Fatalf("missing_backend filter should return 200, got %d body=%s", missingBackend.Code, missingBackend.Body.String())
	}
	for _, want := range []string{
		"FX-CONTRACT-N9E-ALERT-MUTE-LIST-BY-BUSI-GROUP",
		"FX-CONTRACT-N9E-ALERT-MUTE-CREATE",
		"FX-CONTRACT-N9E-ALERT-MUTE-UPDATE",
		"FX-CONTRACT-N9E-ALERT-MUTE-BULK-FIELDS-UPDATE",
	} {
		if !contractMatrixResponseContains(missingBackend.Body.String(), want) {
			t.Fatalf("missing_backend filter missing %s: %s", want, missingBackend.Body.String())
		}
	}
	for _, forbidden := range []string{
		"FX-CONTRACT-N9E-ALERT-MUTE-PREVIEW-EVENTS",
		"FX-CONTRACT-N9E-ALERT-MUTE-TRYRUN",
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
		"FX-CONTRACT-N9E-ALERT-MUTE-PREVIEW-EVENTS",
		"FX-CONTRACT-N9E-ALERT-MUTE-TRYRUN",
	} {
		if !contractMatrixResponseContains(missingDatasource.Body.String(), want) {
			t.Fatalf("missing_datasource filter missing %s: %s", want, missingDatasource.Body.String())
		}
	}
	for _, forbidden := range []string{
		"FX-CONTRACT-N9E-ALERT-MUTE-CREATE",
		"FX-CONTRACT-N9E-ALERT-MUTE-UPDATE",
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

func TestContractMatrixAlertMuteShieldListDoesNotExposeReadyRuntimeShape(t *testing.T) {
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
		if _, ok := alertMuteShieldExpectationByID()[item.ID]; !ok {
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

func alertMuteShieldHandlerExpectations() []struct {
	id     string
	status string
} {
	return []struct {
		id     string
		status string
	}{
		{id: "FX-CONTRACT-N9E-ALERT-MUTE-SHIELD", status: model.ContractStatusBlocked},
		{id: "FX-CONTRACT-N9E-ALERT-MUTE-LIST-BY-BUSI-GROUP", status: model.ContractStatusMissingBackend},
		{id: "FX-CONTRACT-N9E-ALERT-MUTE-LIST-BY-BUSI-GROUPS", status: model.ContractStatusMissingBackend},
		{id: "FX-CONTRACT-N9E-ALERT-MUTE-DETAIL", status: model.ContractStatusMissingBackend},
		{id: "FX-CONTRACT-N9E-ALERT-MUTE-CREATE", status: model.ContractStatusMissingBackend},
		{id: "FX-CONTRACT-N9E-ALERT-MUTE-UPDATE", status: model.ContractStatusMissingBackend},
		{id: "FX-CONTRACT-N9E-ALERT-MUTE-DELETE", status: model.ContractStatusMissingBackend},
		{id: "FX-CONTRACT-N9E-ALERT-MUTE-BULK-FIELDS-UPDATE", status: model.ContractStatusMissingBackend},
		{id: "FX-CONTRACT-N9E-ALERT-MUTE-PREVIEW-EVENTS", status: model.ContractStatusMissingDatasource},
		{id: "FX-CONTRACT-N9E-ALERT-MUTE-TRYRUN", status: model.ContractStatusMissingDatasource},
	}
}

func alertMuteShieldExpectationByID() map[string]string {
	out := map[string]string{}
	for _, item := range alertMuteShieldHandlerExpectations() {
		out[item.id] = item.status
	}
	return out
}
