package handler

import (
	"encoding/json"
	"net/http"
	"testing"

	"ai-workbench-api/internal/model"
	"ai-workbench-api/internal/store"
)

func TestContractMatrixAlertEventActionPipelineQueryDetailsAreBlocked(t *testing.T) {
	store.ResetContractMatrixForTest()
	r := contractMatrixTestRouter()

	for _, tt := range alertEventActionPipelineQueryHandlerExpectations() {
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
				t.Fatalf("blocked response mismatch for %s: %#v", tt.id, payload)
			}
			assertContractMatrixNoSensitiveLeak(t, w.Body.String())
			assertContractMatrixBlockedShape(t, w.Body.String())
			assertContractMatrixNoFakeRuntimeState(t, w.Body.String())
		})
	}
}

func TestContractMatrixAlertEventActionPipelineQueryStatusDistribution(t *testing.T) {
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

	counts := map[string]int{}
	for _, item := range payload.Items {
		if _, ok := alertEventActionPipelineQueryExpectationByID()[item.ID]; !ok {
			continue
		}
		counts[item.Status]++
		if item.SafeToRetry ||
			item.Handler != "" ||
			item.Backend != "" ||
			item.Datasource != "" ||
			item.Executor != "" ||
			len(item.EvidenceRefs) != 0 {
			t.Fatalf("%s must not expose ready runtime fields: %#v", item.ID, item)
		}
	}
	if counts[model.ContractStatusBlocked] != 3 ||
		counts[model.ContractStatusMissingBackend] != 10 ||
		counts[model.ContractStatusMissingDatasource] != 3 ||
		counts[model.ContractStatusMissingExecutor] != 2 {
		t.Fatalf("status distribution mismatch: %#v", counts)
	}
	assertContractMatrixNoSensitiveLeak(t, resp.Body.String())
	assertContractMatrixNoFakeRuntimeState(t, resp.Body.String())
}

func TestContractMatrixAlertEventActionPipelineQueryStatusFilters(t *testing.T) {
	store.ResetContractMatrixForTest()
	r := contractMatrixTestRouter()

	tests := []struct {
		status    string
		want      []string
		forbidden []string
	}{
		{
			status: model.ContractStatusMissingBackend,
			want: []string{
				"FX-CONTRACT-N9E-ALERT-EVENT-ACK",
				"FX-CONTRACT-N9E-EVENT-PIPELINE-LIST",
				"FX-CONTRACT-N9E-EVENT-PIPELINE-UPDATE",
				"FX-CONTRACT-N9E-EVENT-PIPELINE-EXECUTION-DETAIL",
			},
			forbidden: []string{
				"FX-CONTRACT-N9E-EVENT-PIPELINE-TRYRUN",
				"FX-CONTRACT-N9E-EVENT-ENRICH-DATA-PREVIEW",
			},
		},
		{
			status: model.ContractStatusMissingDatasource,
			want: []string{
				"FX-CONTRACT-N9E-EVENT-TAGKEYS",
				"FX-CONTRACT-N9E-EVENT-TAGVALUES",
				"FX-CONTRACT-N9E-EVENT-ENRICH-DATA-PREVIEW",
			},
			forbidden: []string{
				"FX-CONTRACT-N9E-ALERT-EVENT-ACK",
				"FX-CONTRACT-N9E-EVENT-PIPELINE-TRYRUN",
			},
		},
		{
			status: model.ContractStatusMissingExecutor,
			want: []string{
				"FX-CONTRACT-N9E-EVENT-PROCESSOR-TRYRUN",
				"FX-CONTRACT-N9E-EVENT-PIPELINE-TRYRUN",
			},
			forbidden: []string{
				"FX-CONTRACT-N9E-EVENT-PIPELINE-LIST",
				"FX-CONTRACT-N9E-ALERT-EVENT-SHARE-CREDENTIAL",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.status, func(t *testing.T) {
			resp := performContractMatrixRequest(t, r, http.MethodGet, "/contract-matrix?domain=monitoring&status="+tt.status, nil)
			if resp.Code != http.StatusOK {
				t.Fatalf("%s filter should return 200, got %d body=%s", tt.status, resp.Code, resp.Body.String())
			}
			for _, want := range tt.want {
				if !contractMatrixResponseContains(resp.Body.String(), want) {
					t.Fatalf("%s filter missing %s: %s", tt.status, want, resp.Body.String())
				}
			}
			for _, forbidden := range tt.forbidden {
				if contractMatrixResponseContains(resp.Body.String(), forbidden) {
					t.Fatalf("%s filter included %s: %s", tt.status, forbidden, resp.Body.String())
				}
			}
			assertContractMatrixNoSensitiveLeak(t, resp.Body.String())
			assertContractMatrixNoFakeRuntimeState(t, resp.Body.String())
		})
	}
}

func alertEventActionPipelineQueryHandlerExpectations() []struct {
	id     string
	status string
} {
	return []struct {
		id     string
		status string
	}{
		{id: "FX-CONTRACT-N9E-ALERT-EVENT-ACTION-PIPELINE-QUERY", status: model.ContractStatusBlocked},
		{id: "FX-CONTRACT-N9E-ALERT-EVENT-ACK", status: model.ContractStatusMissingBackend},
		{id: "FX-CONTRACT-N9E-ALERT-EVENT-SHARE-CREDENTIAL", status: model.ContractStatusMissingBackend},
		{id: "FX-CONTRACT-N9E-ALERT-EVENT-SHARED-DETAIL", status: model.ContractStatusMissingBackend},
		{id: "FX-CONTRACT-N9E-EVENT-PIPELINE-CRUD", status: model.ContractStatusBlocked},
		{id: "FX-CONTRACT-N9E-EVENT-PIPELINE-LIST", status: model.ContractStatusMissingBackend},
		{id: "FX-CONTRACT-N9E-EVENT-PIPELINE-DETAIL", status: model.ContractStatusMissingBackend},
		{id: "FX-CONTRACT-N9E-EVENT-PIPELINE-CREATE", status: model.ContractStatusMissingBackend},
		{id: "FX-CONTRACT-N9E-EVENT-PIPELINE-UPDATE", status: model.ContractStatusMissingBackend},
		{id: "FX-CONTRACT-N9E-EVENT-PIPELINE-DELETE", status: model.ContractStatusMissingBackend},
		{id: "FX-CONTRACT-N9E-EVENT-PROCESSOR-TRYRUN", status: model.ContractStatusMissingExecutor},
		{id: "FX-CONTRACT-N9E-EVENT-PIPELINE-TRYRUN", status: model.ContractStatusMissingExecutor},
		{id: "FX-CONTRACT-N9E-EVENT-TAGKEYS", status: model.ContractStatusMissingDatasource},
		{id: "FX-CONTRACT-N9E-EVENT-TAGVALUES", status: model.ContractStatusMissingDatasource},
		{id: "FX-CONTRACT-N9E-EVENT-ENRICH-DATA-PREVIEW", status: model.ContractStatusMissingDatasource},
		{id: "FX-CONTRACT-N9E-EVENT-PIPELINE-EXECUTIONS-LIST", status: model.ContractStatusMissingBackend},
		{id: "FX-CONTRACT-N9E-EVENT-PIPELINE-EXECUTION-DETAIL", status: model.ContractStatusMissingBackend},
		{id: "FX-CONTRACT-N9E-ALERT-EVENT-RULE-TESTER", status: model.ContractStatusBlocked},
	}
}

func alertEventActionPipelineQueryExpectationByID() map[string]string {
	out := map[string]string{}
	for _, item := range alertEventActionPipelineQueryHandlerExpectations() {
		out[item.id] = item.status
	}
	return out
}
