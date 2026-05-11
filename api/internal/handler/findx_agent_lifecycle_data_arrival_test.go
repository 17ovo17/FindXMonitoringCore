package handler

import (
	"strings"
	"testing"
	"time"

	"ai-workbench-api/internal/model"
	"ai-workbench-api/internal/store"
)

func TestFindXAgentDataArrivalMatrixBlocksWithoutReceiverEvidence(t *testing.T) {
	rows := agentDataArrival([]*model.FindXAgent{{
		ID:           "agent-matrix",
		Status:       "online",
		Capabilities: []string{"metrics", "logs", "tracing", "profiling", "inspection", "rum", "gateway"},
		LastSeen:     time.Now(),
	}})
	for _, kind := range requiredDataArrivalKindsForTest() {
		row := dataArrivalRowForTest(t, rows, kind)
		if row.Status != model.FindXAgentDataArrivalStatusBlocked || row.EvidenceCount != 0 {
			t.Fatalf("kind %s should stay blocked without receiver evidence, got %#v", kind, row)
		}
		if !strings.Contains(row.Blocker, "BLOCKED_BY_CONTRACT") || !strings.Contains(row.Blocker, "不能替代") {
			t.Fatalf("kind %s should explain blocked validator and no substitution, got %q", kind, row.Blocker)
		}
	}
	if dataArrivalRowForTest(t, rows, model.FindXAgentDataArrivalKindGatewayTrace).AgentCount != 1 {
		t.Fatalf("gateway capability should be counted without becoming reported: %#v", rows)
	}
}

func TestMergeDataArrivalEvidenceReportsOnlyReceiverBackedKinds(t *testing.T) {
	rows := agentDataArrival([]*model.FindXAgent{{
		ID:           "agent-merge",
		Status:       "online",
		Capabilities: []string{"metrics", "logs", "tracing", "profiling", "inspection", "rum", "gateway"},
		LastSeen:     time.Now(),
	}})
	merged := mergeDataArrivalEvidence(rows, map[string]model.FindXAgentDataArrival{
		model.FindXAgentDataArrivalKindHeartbeat:    reportedArrivalForTest(model.FindXAgentDataArrivalKindHeartbeat),
		model.FindXAgentDataArrivalKindMetrics:      reportedArrivalForTest(model.FindXAgentDataArrivalKindMetrics),
		model.FindXAgentDataArrivalKindLogs:         reportedArrivalForTest(model.FindXAgentDataArrivalKindLogs),
		model.FindXAgentDataArrivalKindTracing:      reportedArrivalForTest(model.FindXAgentDataArrivalKindTracing),
		model.FindXAgentDataArrivalKindProfiling:    reportedArrivalForTest(model.FindXAgentDataArrivalKindProfiling),
		model.FindXAgentDataArrivalKindInspection:   reportedArrivalForTest(model.FindXAgentDataArrivalKindInspection),
		model.FindXAgentDataArrivalKindRUM:          reportedArrivalForTest(model.FindXAgentDataArrivalKindRUM),
		model.FindXAgentDataArrivalKindGatewayTrace: reportedArrivalForTest(model.FindXAgentDataArrivalKindGatewayTrace),
		"unknown_signal":                            reportedArrivalForTest("unknown_signal"),
	})
	for _, kind := range []string{
		model.FindXAgentDataArrivalKindHeartbeat,
		model.FindXAgentDataArrivalKindMetrics,
		model.FindXAgentDataArrivalKindLogs,
		model.FindXAgentDataArrivalKindTracing,
	} {
		row := dataArrivalRowForTest(t, merged, kind)
		if row.Status != model.FindXAgentDataArrivalStatusReported || row.EvidenceCount != 1 || row.Blocker != "" {
			t.Fatalf("receiver-backed kind %s should be reported from evidence, got %#v", kind, row)
		}
	}
	for _, kind := range []string{
		model.FindXAgentDataArrivalKindProfiling,
		model.FindXAgentDataArrivalKindInspection,
		model.FindXAgentDataArrivalKindRUM,
		model.FindXAgentDataArrivalKindGatewayTrace,
	} {
		row := dataArrivalRowForTest(t, merged, kind)
		if row.Status == model.FindXAgentDataArrivalStatusReported || row.EvidenceCount != 0 {
			t.Fatalf("non receiver-backed kind %s must not be promoted by evidence map, got %#v", kind, row)
		}
	}
	if hasDataArrivalKindForTest(merged, "unknown_signal") {
		t.Fatalf("unknown data-arrival evidence must not be appended to summary: %#v", merged)
	}
}

func TestDataArrivalEvidenceStoreBlocksUnsupportedAndNonReceiverKinds(t *testing.T) {
	resetCategrafReceiverTestState(t)
	for _, input := range []model.FindXAgentDataArrivalEvidence{
		{
			Kind:     model.FindXAgentDataArrivalKindProfiling,
			Status:   model.FindXAgentDataArrivalStatusReported,
			Metadata: map[string]string{"password": "do-not-store", "safe": "kept"},
		},
		{
			Kind:   "unknown_signal",
			Status: model.FindXAgentDataArrivalStatusReported,
		},
	} {
		saved, err := store.SaveFindXAgentDataArrivalEvidence(input)
		if err != nil {
			t.Fatalf("save evidence: %v", err)
		}
		if saved.Status != model.FindXAgentDataArrivalStatusBlocked || !strings.Contains(saved.Blocker, "BLOCKED_BY_CONTRACT") {
			t.Fatalf("unsupported/non receiver-backed evidence should be blocked, got %#v", saved)
		}
		if strings.Contains(strings.ToLower(saved.Metadata["password"]), "do-not-store") {
			t.Fatalf("sensitive metadata should not be retained: %#v", saved.Metadata)
		}
	}
}

func requiredDataArrivalKindsForTest() []string {
	return []string{
		model.FindXAgentDataArrivalKindHeartbeat,
		model.FindXAgentDataArrivalKindMetrics,
		model.FindXAgentDataArrivalKindLogs,
		model.FindXAgentDataArrivalKindTracing,
		model.FindXAgentDataArrivalKindProfiling,
		model.FindXAgentDataArrivalKindInspection,
		model.FindXAgentDataArrivalKindTopology,
		model.FindXAgentDataArrivalKindRUM,
		model.FindXAgentDataArrivalKindGatewayTrace,
	}
}

func reportedArrivalForTest(kind string) model.FindXAgentDataArrival {
	return model.FindXAgentDataArrival{
		Kind:          kind,
		Status:        model.FindXAgentDataArrivalStatusReported,
		EvidenceCount: 1,
		LastSeen:      time.Now(),
	}
}

func dataArrivalRowForTest(t *testing.T, rows []model.FindXAgentDataArrival, kind string) model.FindXAgentDataArrival {
	t.Helper()
	for _, row := range rows {
		if row.Kind == kind {
			return row
		}
	}
	t.Fatalf("missing data-arrival kind %s in %#v", kind, rows)
	return model.FindXAgentDataArrival{}
}

func hasDataArrivalKindForTest(rows []model.FindXAgentDataArrival, kind string) bool {
	for _, row := range rows {
		if row.Kind == kind {
			return true
		}
	}
	return false
}
