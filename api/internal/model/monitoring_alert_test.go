package model

import (
	"strings"
	"testing"
)

func TestMonitorAlertFingerprintStableAcrossLabelOrder(t *testing.T) {
	first := &MonitorAlertEvent{
		RuleID:       "rule-a",
		DatasourceID: "prometheus-default",
		TargetIdent:  "host-a",
		Labels:       map[string]string{"env": "prod", "service": "api"},
	}
	second := &MonitorAlertEvent{
		RuleID:       "rule-a",
		DatasourceID: "prometheus-default",
		TargetIdent:  "host-a",
		Labels:       map[string]string{"service": "api", "env": "prod"},
	}
	if got, want := GenerateMonitorAlertEventFingerprint(first), GenerateMonitorAlertEventFingerprint(second); got != want {
		t.Fatalf("fingerprint must ignore map order, got %s want %s", got, want)
	}
}

func TestMonitorAlertFingerprintDiffersByRuleTargetAndLabels(t *testing.T) {
	base := &MonitorAlertEvent{RuleID: "rule-a", DatasourceID: "ds-a", TargetIdent: "host-a", Labels: map[string]string{"service": "api"}}
	cases := []*MonitorAlertEvent{
		{RuleID: "rule-b", DatasourceID: "ds-a", TargetIdent: "host-a", Labels: map[string]string{"service": "api"}},
		{RuleID: "rule-a", DatasourceID: "ds-a", TargetIdent: "host-b", Labels: map[string]string{"service": "api"}},
		{RuleID: "rule-a", DatasourceID: "ds-a", TargetIdent: "host-a", Labels: map[string]string{"service": "worker"}},
	}
	fp := GenerateMonitorAlertEventFingerprint(base)
	for _, item := range cases {
		if got := GenerateMonitorAlertEventFingerprint(item); got == fp {
			t.Fatalf("fingerprint should differ for %+v", item)
		}
	}
}

func TestMonitorAlertFingerprintIgnoresSensitiveLabels(t *testing.T) {
	first := &MonitorAlertEvent{RuleID: "rule-a", DatasourceID: "ds-a", TargetIdent: "host-a", Labels: map[string]string{"api_key": "secret-a", "auth": "bearer-a", "token": "secret-a", "env": "prod"}}
	second := &MonitorAlertEvent{RuleID: "rule-a", DatasourceID: "ds-a", TargetIdent: "host-a", Labels: map[string]string{"api_key": "secret-b", "auth": "bearer-b", "token": "secret-b", "env": "prod"}}
	fp := GenerateMonitorAlertEventFingerprint(first)
	if fp != GenerateMonitorAlertEventFingerprint(second) {
		t.Fatal("sensitive labels must not affect fingerprint")
	}
	if strings.Contains(fp, "secret") || strings.Contains(fp, "token") {
		t.Fatalf("fingerprint leaked sensitive text: %s", fp)
	}
}

func TestMonitorAlertEventStateMachineRejectsTerminalActions(t *testing.T) {
	for _, status := range []string{MonitorAlertEventStatusResolved, MonitorAlertEventStatusArchived} {
		for _, action := range []string{
			MonitorAlertEventActionAck,
			MonitorAlertEventActionAssign,
			MonitorAlertEventActionMute,
			MonitorAlertEventActionResolve,
			MonitorAlertEventActionArchive,
		} {
			if _, err := ValidateMonitorAlertEventTransition(status, action); err == nil {
				t.Fatalf("expected terminal transition error for %s/%s", status, action)
			}
		}
	}
}
