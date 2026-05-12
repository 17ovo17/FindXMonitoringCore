package store

import (
	"strings"
	"testing"

	"ai-workbench-api/internal/model"
)

var monitoringContractSeedIDs = []string{
	"FX-CONTRACT-N9E-DATASOURCE-BRIEF-LIST",
	"FX-CONTRACT-N9E-DATASOURCE-PROXY-BY-ID",
	"FX-CONTRACT-N9E-DATASOURCE-TEST-CONNECTION",
	"FX-CONTRACT-N9E-SYSTEM-INTEGRATION-CATALOG",
	"FX-CONTRACT-N9E-TEMPLATE-CENTER-DASHBOARD-IMPORT",
	"FX-CONTRACT-N9E-METRIC-VIEWS-CRUD",
	"FX-CONTRACT-N9E-METRIC-QUERY-BATCH",
	"FX-CONTRACT-N9E-DASHBOARD-ANNOTATIONS",
	"FX-CONTRACT-N9E-ALERT-RULE-GROUPS",
	"FX-CONTRACT-N9E-NOTIFICATION-FINDX-ADAPTER",
	"FX-CONTRACT-N9E-BUSI-GROUP-RESOURCE-GROUP-MAP",
}

func TestMonitoringContractMatrixSeedContainsRequiredGaps(t *testing.T) {
	ResetContractMatrixForTest()

	items, err := ListContractMatrixEntries("", "monitoring")
	if err != nil {
		t.Fatalf("list monitoring contract matrix: %v", err)
	}
	got := map[string]model.ContractMatrixEntry{}
	for _, item := range items {
		got[item.ID] = item
	}

	for _, id := range monitoringContractSeedIDs {
		item, ok := got[id]
		if !ok {
			t.Fatalf("missing monitoring contract seed %s", id)
		}
		assertMonitoringContractSeedEntry(t, item)
	}
}

func TestMonitoringContractMatrixDomainFilter(t *testing.T) {
	ResetContractMatrixForTest()
	_, err := SaveContractMatrixEntry(model.ContractMatrixRegisterRequest{
		ID:         "FX-CONTRACT-AGENT-EXECUTOR-SSH",
		Domain:     "agent",
		Capability: "ssh executor",
		Status:     model.ContractStatusMissingExecutor,
	})
	if err != nil {
		t.Fatalf("save non-monitoring entry: %v", err)
	}

	items, err := ListContractMatrixEntries("", "monitoring")
	if err != nil {
		t.Fatalf("filter monitoring contract matrix: %v", err)
	}
	for _, item := range items {
		if item.Domain != "monitoring" {
			t.Fatalf("domain filter returned non-monitoring entry: %#v", item)
		}
	}
	if len(items) < len(monitoringContractSeedIDs) {
		t.Fatalf("domain filter did not include monitoring seeds, got %d", len(items))
	}
}

func TestMonitoringContractMatrixSeedDoesNotOverwriteUserEntry(t *testing.T) {
	ResetContractMatrixForTest()
	const customReason = "operator registered custom contract gap"
	_, err := SaveContractMatrixEntry(model.ContractMatrixRegisterRequest{
		ID:            "FX-CONTRACT-N9E-METRIC-VIEWS-CRUD",
		Domain:        "monitoring",
		Capability:    "custom metric views",
		Status:        model.ContractStatusBlocked,
		SourceRefs:    []string{"custom-source-ref"},
		BlockedReason: customReason,
		Metadata:      map[string]string{"findx_route": "/custom-monitoring"},
	})
	if err != nil {
		t.Fatalf("save user entry: %v", err)
	}

	item, ok, err := GetContractMatrixEntry("FX-CONTRACT-N9E-METRIC-VIEWS-CRUD")
	if err != nil {
		t.Fatalf("get seeded id after user entry: %v", err)
	}
	if !ok {
		t.Fatal("expected user entry to remain available")
	}
	if item.BlockedReason != customReason || item.Metadata["findx_route"] != "/custom-monitoring" {
		t.Fatalf("seed overwrote user entry: %#v", item)
	}
}

func TestContractMatrixMetadataDropsSensitiveAndFakeSuccessText(t *testing.T) {
	ResetContractMatrixForTest()
	item, err := SaveContractMatrixEntry(model.ContractMatrixRegisterRequest{
		ID:         "FX-CONTRACT-N9E-SAFE-METADATA",
		Domain:     "monitoring",
		Capability: "metadata sanitize",
		Status:     model.ContractStatusMissingBackend,
		Metadata: map[string]string{
			"findx_route": "/monitoring/safe",
			"password":    "hidden",
			"phase":       "running",
			"result":      "succeeded",
			"note":        "queued for install",
		},
	})
	if err != nil {
		t.Fatalf("save sanitized metadata entry: %v", err)
	}
	if item.Metadata["findx_route"] != "/monitoring/safe" {
		t.Fatalf("safe metadata was dropped: %#v", item.Metadata)
	}
	for _, key := range []string{"password", "phase", "result", "note"} {
		if _, ok := item.Metadata[key]; ok {
			t.Fatalf("unsafe metadata key/value survived: %s=%q in %#v", key, item.Metadata[key], item.Metadata)
		}
	}
}

func assertMonitoringContractSeedEntry(t *testing.T, item model.ContractMatrixEntry) {
	t.Helper()
	if item.Status == model.ContractStatusReady {
		t.Fatalf("monitoring seed must not claim ready without evidence: %#v", item)
	}
	if !model.IsContractMatrixStatus(item.Status) {
		t.Fatalf("monitoring seed used invalid status: %#v", item)
	}
	if len(item.SourceRefs) == 0 {
		t.Fatalf("monitoring seed missing source refs: %#v", item)
	}
	if item.Metadata["findx_route"] == "" {
		t.Fatalf("monitoring seed missing findx_route metadata: %#v", item)
	}
	if item.Metadata["gap_type"] == "" && item.Metadata["findx_adapter"] == "" {
		t.Fatalf("monitoring seed missing gap metadata: %#v", item)
	}
	body := strings.ToLower(strings.Join(append(item.SourceRefs, item.BlockedReason), " "))
	for _, forbidden := range []string{"queued", "running", "succeeded", "success", "applied", "rolled-back", "installed", "data_arrived"} {
		if strings.Contains(body, forbidden) {
			t.Fatalf("monitoring seed exposed fake success state %q: %#v", forbidden, item)
		}
	}
}
