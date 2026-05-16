package handler

import (
	"encoding/json"
	"net/http"
	"testing"

	"ai-workbench-api/internal/model"
	"ai-workbench-api/internal/store"
)

func TestContractMatrixSeededNotificationAdapterDetailsAreBlocked(t *testing.T) {
	store.ResetContractMatrixForTest()
	r := contractMatrixTestRouter()

	for _, tt := range notificationAdapterHandlerExpectations() {
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
				t.Fatalf("notification adapter blocked response mismatch for %s: %#v", tt.id, payload)
			}
			assertContractMatrixNoSensitiveLeak(t, w.Body.String())
			assertContractMatrixBlockedShape(t, w.Body.String())
			assertContractMatrixNoFakeRuntimeState(t, w.Body.String())
		})
	}
}

func TestContractMatrixNotificationAdapterStatusDistribution(t *testing.T) {
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
		if _, ok := notificationAdapterExpectationByID()[item.ID]; !ok {
			continue
		}
		counts[item.Status]++
		if item.SafeToRetry ||
			item.Handler != "" ||
			item.Backend != "" ||
			item.Datasource != "" ||
			item.Executor != "" ||
			len(item.EvidenceRefs) != 0 {
			t.Fatalf("%s must not expose fake completion or executable evidence: %#v", item.ID, item)
		}
	}
	if counts[model.ContractStatusBlocked] != 1 ||
		counts[model.ContractStatusMissingBackend] != 26 ||
		counts[model.ContractStatusMissingDatasource] != 4 ||
		counts[model.ContractStatusMissingExecutor] != 5 {
		t.Fatalf("notification adapter status distribution mismatch: %#v", counts)
	}
	assertContractMatrixNoSensitiveLeak(t, resp.Body.String())
	assertContractMatrixNoFakeRuntimeState(t, resp.Body.String())
}

func TestContractMatrixNotificationAdapterMetadataSurvivesSanitizer(t *testing.T) {
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
		if _, ok := notificationAdapterExpectationByID()[item.ID]; !ok {
			continue
		}
		if item.Metadata["findx_route"] == "" || item.Metadata["gap_type"] == "" || item.Metadata["upstream_ref"] == "" {
			t.Fatalf("%s metadata was unexpectedly cleaned: %#v", item.ID, item.Metadata)
		}
	}
	assertContractMatrixNoSensitiveLeak(t, resp.Body.String())
	assertContractMatrixNoFakeRuntimeState(t, resp.Body.String())
}

func notificationAdapterHandlerExpectations() []struct {
	id     string
	status string
} {
	return []struct {
		id     string
		status string
	}{
		{id: "FX-CONTRACT-N9E-NOTIFICATION-FINDX-ADAPTER", status: model.ContractStatusBlocked},
		{id: "FX-CONTRACT-N9E-NOTIFY-RULES-LIST", status: model.ContractStatusMissingBackend},
		{id: "FX-CONTRACT-N9E-NOTIFY-RULE-DETAIL", status: model.ContractStatusMissingBackend},
		{id: "FX-CONTRACT-N9E-NOTIFY-RULES-CREATE", status: model.ContractStatusMissingBackend},
		{id: "FX-CONTRACT-N9E-NOTIFY-RULE-UPDATE", status: model.ContractStatusMissingBackend},
		{id: "FX-CONTRACT-N9E-NOTIFY-RULES-DELETE", status: model.ContractStatusMissingBackend},
		{id: "FX-CONTRACT-N9E-NOTIFY-RULE-CUSTOM-PARAMS", status: model.ContractStatusMissingBackend},
		{id: "FX-CONTRACT-N9E-NOTIFY-RULE-TEST", status: model.ContractStatusMissingExecutor},
		{id: "FX-CONTRACT-N9E-NOTIFY-STATISTICS", status: model.ContractStatusMissingDatasource},
		{id: "FX-CONTRACT-N9E-NOTIFY-EVENTS", status: model.ContractStatusMissingDatasource},
		{id: "FX-CONTRACT-N9E-NOTIFY-ALERT-RULES", status: model.ContractStatusMissingBackend},
		{id: "FX-CONTRACT-N9E-NOTIFY-SUBSCRIBE-RULES", status: model.ContractStatusMissingBackend},
		{id: "FX-CONTRACT-N9E-NOTIFY-EVENT-TAGKEYS", status: model.ContractStatusMissingDatasource},
		{id: "FX-CONTRACT-N9E-NOTIFY-FEISHU-GROUPS", status: model.ContractStatusMissingExecutor},
		{id: "FX-CONTRACT-N9E-NOTIFY-FLASHDUTY-CHANNELS", status: model.ContractStatusMissingExecutor},
		{id: "FX-CONTRACT-N9E-NOTIFY-PAGERDUTY-SERVICES", status: model.ContractStatusMissingExecutor},
		{id: "FX-CONTRACT-N9E-NOTIFY-PAGERDUTY-CONNECTOR-LOOKUP", status: model.ContractStatusMissingExecutor},
		{id: "FX-CONTRACT-N9E-NOTIFY-CHANNEL-CONFIGS-LIST", status: model.ContractStatusMissingBackend},
		{id: "FX-CONTRACT-N9E-NOTIFY-CHANNEL-CONFIGS-SIMPLIFIED", status: model.ContractStatusMissingBackend},
		{id: "FX-CONTRACT-N9E-NOTIFY-CHANNEL-CONFIGS-CREATE", status: model.ContractStatusMissingBackend},
		{id: "FX-CONTRACT-N9E-NOTIFY-CHANNEL-CONFIG-UPDATE", status: model.ContractStatusMissingBackend},
		{id: "FX-CONTRACT-N9E-NOTIFY-CHANNEL-CONFIG-DETAIL", status: model.ContractStatusMissingBackend},
		{id: "FX-CONTRACT-N9E-NOTIFY-CHANNEL-CONFIG-BY-IDENT", status: model.ContractStatusMissingBackend},
		{id: "FX-CONTRACT-N9E-NOTIFY-CHANNEL-CONFIGS-DELETE", status: model.ContractStatusMissingBackend},
		{id: "FX-CONTRACT-N9E-NOTIFY-CHANNEL-CONFIG-IDENTS", status: model.ContractStatusMissingBackend},
		{id: "FX-CONTRACT-N9E-MESSAGE-TEMPLATES-LIST", status: model.ContractStatusMissingBackend},
		{id: "FX-CONTRACT-N9E-MESSAGE-TEMPLATE-DETAIL", status: model.ContractStatusMissingBackend},
		{id: "FX-CONTRACT-N9E-MESSAGE-TEMPLATES-CREATE", status: model.ContractStatusMissingBackend},
		{id: "FX-CONTRACT-N9E-MESSAGE-TEMPLATE-UPDATE", status: model.ContractStatusMissingBackend},
		{id: "FX-CONTRACT-N9E-MESSAGE-TEMPLATES-DELETE", status: model.ContractStatusMissingBackend},
		{id: "FX-CONTRACT-N9E-MESSAGE-TEMPLATE-PREVIEW", status: model.ContractStatusMissingDatasource},
		{id: "FX-CONTRACT-N9E-NOTIFY-CONTACTS-LIST", status: model.ContractStatusMissingBackend},
		{id: "FX-CONTRACT-N9E-NOTIFY-CONTACTS-UPDATE", status: model.ContractStatusMissingBackend},
		{id: "FX-CONTRACT-N9E-MANAGE-NOTIFY-CHANNELS", status: model.ContractStatusMissingBackend},
		{id: "FX-CONTRACT-N9E-MANAGE-CONTACT-CHANNELS", status: model.ContractStatusMissingBackend},
		{id: "FX-CONTRACT-N9E-MANAGE-CONTACT-KEYS", status: model.ContractStatusMissingBackend},
	}
}

func notificationAdapterExpectationByID() map[string]string {
	out := map[string]string{}
	for _, item := range notificationAdapterHandlerExpectations() {
		out[item.id] = item.status
	}
	return out
}
