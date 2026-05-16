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
		if !strings.Contains(row.Blocker, "PENDING") || !strings.Contains(row.Blocker, "不能替代") {
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
			Metadata: map[string]string{"password": "do-not-store", "safe": "kept", "state": "succeeded"},
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
		if saved.Status != model.FindXAgentDataArrivalStatusBlocked || !strings.Contains(saved.Blocker, "PENDING") {
			t.Fatalf("unsupported/non receiver-backed evidence should be blocked, got %#v", saved)
		}
		if strings.Contains(strings.ToLower(saved.Metadata["password"]), "do-not-store") {
			t.Fatalf("sensitive metadata should not be retained: %#v", saved.Metadata)
		}
		if saved.Metadata["state"] == "succeeded" {
			t.Fatalf("fake completion state should not be retained: %#v", saved.Metadata)
		}
		assertDataArrivalEvidenceMetadataContract(t, saved)
	}
}

func TestDataArrivalEvidenceSnapshotExposesSignalSummary(t *testing.T) {
	resetCategrafReceiverTestState(t)
	older, err := store.SaveFindXAgentDataArrivalEvidence(model.FindXAgentDataArrivalEvidence{
		Kind:         model.FindXAgentDataArrivalKindLogs,
		AgentID:      "agent-log-summary",
		Status:       model.FindXAgentDataArrivalStatusReported,
		EvidenceRefs: []string{"receiver:/findx-agent/logs-compatible"},
		Metadata: map[string]string{
			"package_version": "1.2.3",
			"config_version":  "cfg-7",
			"log_id":          "log-1",
			"password":        "do-not-store",
		},
		CreatedAt: time.Now().Add(-time.Hour),
	})
	if err != nil {
		t.Fatalf("save older evidence: %v", err)
	}
	newer, err := store.SaveFindXAgentDataArrivalEvidence(model.FindXAgentDataArrivalEvidence{
		Kind:         model.FindXAgentDataArrivalKindLogs,
		AgentID:      "agent-log-summary",
		Status:       model.FindXAgentDataArrivalStatusReported,
		EvidenceRefs: []string{"receiver:/findx-agent/logs-compatible"},
		Metadata: map[string]string{
			"package_version": "1.2.4",
			"config_version":  "cfg-8",
			"trace_id":        "trace-1",
		},
	})
	if err != nil {
		t.Fatalf("save newer evidence: %v", err)
	}
	row := store.DataArrivalEvidenceSnapshot()[model.FindXAgentDataArrivalKindLogs]
	if row.Status != model.FindXAgentDataArrivalStatusReported || row.EvidenceCount != 2 {
		t.Fatalf("expected reported snapshot with two evidence records, got %#v", row)
	}
	if row.SourceAgent != "agent-log-summary" || row.BackendReceiver != "receiver:/findx-agent/logs-compatible" {
		t.Fatalf("snapshot should expose source agent and receiver: %#v", row)
	}
	if row.PackageVersion != "1.2.4" || row.ConfigVersion != "cfg-8" || row.SampleEvidence == "" {
		t.Fatalf("snapshot should expose latest package/config/sample evidence: %#v", row)
	}
	if !row.FirstSeen.Equal(older.CreatedAt) || !row.LastSeen.Equal(newer.UpdatedAt) || !row.LastSeenAt.Equal(newer.UpdatedAt) {
		t.Fatalf("snapshot should expose first/last seen from evidence: older=%s newer=%s row=%#v", older.CreatedAt, newer.UpdatedAt, row)
	}
	if len(row.RelatedIDs) != 1 || row.RelatedIDs[0] != "trace-1" {
		t.Fatalf("snapshot should expose safe related ids from latest evidence: %#v", row.RelatedIDs)
	}
}

func TestDataArrivalEvidenceSnapshotKeepsReceiverGateForHistoricalRows(t *testing.T) {
	resetCategrafReceiverTestState(t)
	if _, err := store.SaveFindXAgentDataArrivalEvidence(model.FindXAgentDataArrivalEvidence{
		Kind:         model.FindXAgentDataArrivalKindMetrics,
		AgentID:      "agent-snapshot-gate",
		Status:       model.FindXAgentDataArrivalStatusReported,
		EvidenceRefs: []string{"capability:metrics"},
		Metadata: map[string]string{
			"backend_receiver": "receiver:metrics",
			"sample_evidence":  "none",
		},
	}); err != nil {
		t.Fatalf("save metrics evidence without receiver ref: %v", err)
	}
	row := store.DataArrivalEvidenceSnapshot()[model.FindXAgentDataArrivalKindMetrics]
	if row.Status == model.FindXAgentDataArrivalStatusReported {
		t.Fatalf("snapshot must not report receiver-backed kind without receiver evidence ref: %#v", row)
	}
	if row.BackendReceiver != "" || strings.HasPrefix(row.SampleEvidence, "receiver:") {
		t.Fatalf("snapshot must not fabricate receiver evidence: %#v", row)
	}
}

func TestDataArrivalEvidenceSnapshotIncludesRelatedMetricIDs(t *testing.T) {
	resetCategrafReceiverTestState(t)
	if _, err := store.SaveFindXAgentDataArrivalEvidence(model.FindXAgentDataArrivalEvidence{
		Kind:         model.FindXAgentDataArrivalKindMetrics,
		AgentID:      "agent-related-metric",
		Status:       model.FindXAgentDataArrivalStatusReported,
		EvidenceRefs: []string{"receiver:/metrics/write-compatible"},
		Metadata: map[string]string{
			"related_metric_ids": "metric-a,metric-b",
		},
	}); err != nil {
		t.Fatalf("save metrics evidence with related metric ids: %v", err)
	}
	row := store.DataArrivalEvidenceSnapshot()[model.FindXAgentDataArrivalKindMetrics]
	if len(row.RelatedIDs) != 1 || row.RelatedIDs[0] != "metric-a,metric-b" {
		t.Fatalf("snapshot should keep related_metric_ids: %#v", row.RelatedIDs)
	}
}

func TestDataArrivalEvidenceRequiresReceiverEvidencePerSignal(t *testing.T) {
	resetCategrafReceiverTestState(t)
	if _, err := store.SaveFindXAgentDataArrivalEvidence(model.FindXAgentDataArrivalEvidence{
		Kind:         model.FindXAgentDataArrivalKindHeartbeat,
		AgentID:      "agent-heartbeat-only",
		Status:       model.FindXAgentDataArrivalStatusReported,
		EvidenceRefs: []string{"receiver:/agent/heartbeat-compatible"},
		Metadata:     map[string]string{"package_version": "1.0.0"},
	}); err != nil {
		t.Fatalf("save heartbeat evidence: %v", err)
	}
	rows := mergeDataArrivalEvidence(agentDataArrival([]*model.FindXAgent{{
		ID:           "agent-heartbeat-only",
		Status:       "online",
		Capabilities: []string{"metrics", "logs", "tracing", "profiling", "rum", "gateway"},
		LastSeen:     time.Now(),
	}}), store.DataArrivalEvidenceSnapshot())
	if row := dataArrivalRowForTest(t, rows, model.FindXAgentDataArrivalKindHeartbeat); row.Status != model.FindXAgentDataArrivalStatusReported {
		t.Fatalf("heartbeat should be reported from heartbeat receiver evidence: %#v", row)
	}
	for _, kind := range []string{
		model.FindXAgentDataArrivalKindMetrics,
		model.FindXAgentDataArrivalKindLogs,
		model.FindXAgentDataArrivalKindTracing,
		model.FindXAgentDataArrivalKindProfiling,
		model.FindXAgentDataArrivalKindInspection,
		model.FindXAgentDataArrivalKindRUM,
		model.FindXAgentDataArrivalKindGatewayTrace,
		model.FindXAgentDataArrivalKindTopology,
	} {
		row := dataArrivalRowForTest(t, rows, kind)
		if row.Status == model.FindXAgentDataArrivalStatusReported || row.EvidenceCount != 0 {
			t.Fatalf("heartbeat/capability must not promote %s to reported: %#v", kind, row)
		}
	}
}

func TestReceiverBackedDataArrivalRequiresReceiverRefToReport(t *testing.T) {
	resetCategrafReceiverTestState(t)
	kinds := []string{
		model.FindXAgentDataArrivalKindHeartbeat,
		model.FindXAgentDataArrivalKindMetrics,
		model.FindXAgentDataArrivalKindLogs,
		model.FindXAgentDataArrivalKindTracing,
	}
	for _, kind := range kinds {
		saved, err := store.SaveFindXAgentDataArrivalEvidence(model.FindXAgentDataArrivalEvidence{
			Kind:         kind,
			AgentID:      "agent-no-receiver-" + kind,
			Status:       model.FindXAgentDataArrivalStatusReported,
			EvidenceRefs: []string{"heartbeat:seen", "capability:" + kind},
			Metadata:     map[string]string{"package_version": "1.0.0"},
		})
		if err != nil {
			t.Fatalf("save %s evidence without receiver ref: %v", kind, err)
		}
		if saved.Status != model.FindXAgentDataArrivalStatusBlocked {
			t.Fatalf("%s without receiver ref must be blocked, got %#v", kind, saved)
		}
		if !strings.Contains(saved.Blocker, "receiver evidence ref is required") {
			t.Fatalf("%s should expose receiver-ref blocker, got %q", kind, saved.Blocker)
		}
		if saved.Metadata["blocked_reason"] != saved.Blocker {
			t.Fatalf("%s should write blocked_reason metadata, got %#v", kind, saved.Metadata)
		}
		if saved.Metadata["backend_receiver"] == "receiver:"+kind {
			t.Fatalf("%s must not fabricate backend_receiver from kind: %#v", kind, saved.Metadata)
		}
		if strings.HasPrefix(saved.Metadata["backend_receiver"], "receiver:") {
			t.Fatalf("%s must not expose receiver metadata without receiver ref: %#v", kind, saved.Metadata)
		}
	}
}

func TestReceiverBackedDataArrivalMetadataContract(t *testing.T) {
	resetCategrafReceiverTestState(t)
	kinds := []string{
		model.FindXAgentDataArrivalKindHeartbeat,
		model.FindXAgentDataArrivalKindMetrics,
		model.FindXAgentDataArrivalKindLogs,
		model.FindXAgentDataArrivalKindTracing,
	}
	for _, kind := range kinds {
		saved, err := store.SaveFindXAgentDataArrivalEvidence(model.FindXAgentDataArrivalEvidence{
			Kind:         kind,
			AgentID:      "agent-" + kind,
			Status:       model.FindXAgentDataArrivalStatusReported,
			EvidenceRefs: []string{"receiver:/unit/" + kind},
			Metadata: map[string]string{
				"package_version": "1.0.0",
				"config_version":  "cfg-1",
				"trace_id":        "trace-" + kind,
				"log_id":          "log-" + kind,
				"metric_ids":      "metric-" + kind,
				"token":           "do-not-store",
			},
		})
		if err != nil {
			t.Fatalf("save %s evidence: %v", kind, err)
		}
		if saved.Status != model.FindXAgentDataArrivalStatusReported {
			t.Fatalf("%s should stay reported when receiver-backed evidence exists: %#v", kind, saved)
		}
		assertDataArrivalEvidenceMetadataContract(t, saved)
	}
}

func TestNonReceiverDataArrivalSignalsAlwaysBlocked(t *testing.T) {
	resetCategrafReceiverTestState(t)
	for _, kind := range []string{
		model.FindXAgentDataArrivalKindProfiling,
		model.FindXAgentDataArrivalKindInspection,
		model.FindXAgentDataArrivalKindRUM,
		model.FindXAgentDataArrivalKindGatewayTrace,
		model.FindXAgentDataArrivalKindTopology,
	} {
		saved, err := store.SaveFindXAgentDataArrivalEvidence(model.FindXAgentDataArrivalEvidence{
			Kind:         kind,
			AgentID:      "agent-" + kind,
			Status:       model.FindXAgentDataArrivalStatusReported,
			EvidenceRefs: []string{"receiver:/fake/" + kind},
			Metadata:     map[string]string{"package_version": "1.0.0"},
		})
		if err != nil {
			t.Fatalf("save %s evidence: %v", kind, err)
		}
		if saved.Status != model.FindXAgentDataArrivalStatusBlocked {
			t.Fatalf("%s must be blocked without real receiver/evidence contract: %#v", kind, saved)
		}
		if saved.Metadata["blocked_reason"] == "" || !strings.Contains(saved.Metadata["blocked_reason"], "receiver/evidence/data-arrival validator") {
			t.Fatalf("%s should expose precise blocked_reason: %#v", kind, saved.Metadata)
		}
	}
}

func TestVerifyAgentHeartbeatDoesNotCreateDataArrivalEvidence(t *testing.T) {
	resetCategrafReceiverTestState(t)
	rows := mergeDataArrivalEvidence(agentDataArrival([]*model.FindXAgent{{
		ID:           "agent-prom-check",
		Status:       "online",
		Capabilities: []string{"metrics", "logs", "tracing", "profiling"},
		LastSeen:     time.Now(),
	}}), store.DataArrivalEvidenceSnapshot())
	for _, kind := range requiredDataArrivalKindsForTest() {
		row := dataArrivalRowForTest(t, rows, kind)
		if row.Status == model.FindXAgentDataArrivalStatusReported || row.EvidenceCount != 0 {
			t.Fatalf("without SaveFindXAgentDataArrivalEvidence, heartbeat verify/prometheus checks must not report %s: %#v", kind, row)
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

func assertDataArrivalEvidenceMetadataContract(t *testing.T, item model.FindXAgentDataArrivalEvidence) {
	t.Helper()
	metadata := item.Metadata
	for _, key := range []string{
		"signal_type",
		"source_agent",
		"first_seen_at",
		"last_seen_at",
		"sample_evidence",
		"backend_receiver",
	} {
		if strings.TrimSpace(metadata[key]) == "" {
			t.Fatalf("data-arrival metadata missing %s: %#v", key, metadata)
		}
	}
	if metadata["signal_type"] != item.Kind {
		t.Fatalf("data-arrival metadata should match signal: %#v item=%#v", metadata, item)
	}
	if item.AgentID != "" && metadata["source_agent"] != item.AgentID {
		t.Fatalf("data-arrival metadata should match signal and agent: %#v item=%#v", metadata, item)
	}
	if item.Status == model.FindXAgentDataArrivalStatusBlocked && strings.TrimSpace(metadata["blocked_reason"]) == "" {
		t.Fatalf("blocked data-arrival metadata should expose blocked_reason: %#v", metadata)
	}
	for _, forbidden := range []string{"password", "token", "cookie", "secret", "Bearer", "succeeded", "success", "applied", "installed", "service_registered"} {
		if strings.Contains(strings.ToLower(metadata["state"]), strings.ToLower(forbidden)) ||
			strings.Contains(strings.ToLower(metadata["status"]), strings.ToLower(forbidden)) ||
			strings.Contains(strings.ToLower(metadata["token"]), strings.ToLower(forbidden)) {
			t.Fatalf("data-arrival metadata leaked sensitive or fake completion marker %q: %#v", forbidden, metadata)
		}
	}
}
