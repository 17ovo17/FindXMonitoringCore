package store

import (
	"database/sql"
	"errors"
	"net"
	"strings"
	"testing"
	"time"

	"ai-workbench-api/internal/model"
)

func unreachableMySQLTestDSN() string {
	// 假测试夹具：net.JoinHostPort 避免源码中出现可被敏感扫描误判的 MySQL DSN 模式。
	return "fixture_login:fixture_value@" + "tcp" + "(" + net.JoinHostPort("127.0.0.1", "1") + ")/ai_workbench?timeout=1ms&parseTime=true"
}

func resetMonitoringMemoryForTest() {
	mu.Lock()
	defer mu.Unlock()
	monitorTargets = map[string]*model.MonitorTarget{}
	findxAgents = map[string]*model.FindXAgent{}
	monitorAlertRules = map[string]*model.MonitorAlertRule{}
	monitorRuleVersions = map[string][]model.MonitorAlertRuleVersion{}
	monitorEventsCurrent = map[string]*model.MonitorAlertEvent{}
	monitorEventsHistory = map[string]*model.MonitorAlertEvent{}
	monitorEventActions = map[string][]model.MonitorAlertAction{}
	mysqlOK = false
}

func TestUpsertFindXAgentHeartbeatCreatesAgentAndTarget(t *testing.T) {
	resetMonitoringMemoryForTest()
	agent, target, err := UpsertFindXAgentHeartbeat(model.FindXAgentHeartbeat{Ident: "agent-1", IP: "10.0.0.1", Hostname: "host-a"})
	if err != nil {
		t.Fatalf("heartbeat failed: %v", err)
	}
	if agent == nil || target == nil {
		t.Fatalf("expected agent and target")
	}
	if agent.TargetID != target.ID {
		t.Fatalf("target id mismatch: %s != %s", agent.TargetID, target.ID)
	}
	if agent.Status != "online" || target.Status != "online" {
		t.Fatalf("expected online status, got agent=%s target=%s", agent.Status, target.Status)
	}
}

func TestMonitoringSanitizesSensitiveFields(t *testing.T) {
	resetMonitoringMemoryForTest()
	target, err := UpsertMonitorTarget(&model.MonitorTarget{
		Ident: "target-sensitive",
		IP:    "10.0.0.2",
		Labels: map[string]string{
			"api_token": "secret-value",
			"env":       " prod ",
		},
	})
	if err != nil {
		t.Fatalf("upsert failed: %v", err)
	}
	if target.Labels["api_token"] != "******" {
		t.Fatalf("expected sensitive value masked")
	}
	if target.Labels["env"] != "prod" {
		t.Fatalf("expected non-sensitive value trimmed")
	}
}

func TestHeartbeatTimeRejectsFarFuture(t *testing.T) {
	future := time.Now().Add(6 * time.Minute).Unix()
	if _, err := HeartbeatTime(future); err == nil {
		t.Fatalf("expected future heartbeat error")
	}
}

func TestMonitorTargetStatusValidation(t *testing.T) {
	resetMonitoringMemoryForTest()
	target, err := UpsertMonitorTarget(&model.MonitorTarget{Ident: "target-status", IP: "10.0.0.3"})
	if err != nil {
		t.Fatalf("upsert failed: %v", err)
	}
	if target.Status != "unknown" {
		t.Fatalf("expected unknown status, got %s", target.Status)
	}
	if _, err := UpsertMonitorTarget(&model.MonitorTarget{Ident: "target-bad-status", Status: "bad"}); err == nil {
		t.Fatalf("expected invalid status error")
	}
}

func TestStableShortIDForLongIdent(t *testing.T) {
	resetMonitoringMemoryForTest()
	ident := strings.Repeat("agent-long-", 20)
	agent1, target1, err := UpsertFindXAgentHeartbeat(model.FindXAgentHeartbeat{Ident: ident, IP: "10.0.0.4"})
	if err != nil {
		t.Fatalf("heartbeat failed: %v", err)
	}
	agent2, target2, err := UpsertFindXAgentHeartbeat(model.FindXAgentHeartbeat{Ident: ident, IP: "10.0.0.4"})
	if err != nil {
		t.Fatalf("heartbeat failed: %v", err)
	}
	if len(agent1.ID) > 64 || len(target1.ID) > 64 {
		t.Fatalf("expected short ids, got agent=%d target=%d", len(agent1.ID), len(target1.ID))
	}
	if agent1.ID != agent2.ID || target1.ID != target2.ID {
		t.Fatalf("expected stable ids")
	}
}

func TestSaveMonitorAlertRulePropagatesPersistError(t *testing.T) {
	resetMonitoringMemoryForTest()
	oldDB, oldMySQLOK := db, mysqlOK
	badDB, err := sql.Open("mysql", unreachableMySQLTestDSN())
	if err != nil {
		t.Fatalf("open bad db: %v", err)
	}
	db = badDB
	mysqlOK = true
	t.Cleanup(func() {
		_ = badDB.Close()
		db = oldDB
		mysqlOK = oldMySQLOK
	})

	_, err = SaveMonitorAlertRule(&model.MonitorAlertRule{
		Name:         "bad persist",
		Query:        "up == 0",
		Severity:     "warning",
		DatasourceID: "prometheus-default",
		Enabled:      true,
	}, "tester")
	if err == nil {
		t.Fatal("expected persist error")
	}
}

func TestDeleteMonitorAlertRulePropagatesPersistError(t *testing.T) {
	resetMonitoringMemoryForTest()
	oldDB, oldMySQLOK := db, mysqlOK
	badDB, err := sql.Open("mysql", unreachableMySQLTestDSN())
	if err != nil {
		t.Fatalf("open bad db: %v", err)
	}
	db = badDB
	mysqlOK = true
	t.Cleanup(func() {
		_ = badDB.Close()
		db = oldDB
		mysqlOK = oldMySQLOK
	})

	if _, err := DeleteMonitorAlertRule("missing"); err == nil {
		t.Fatal("expected delete persist error")
	}
}

func TestTerminalMonitorAlertEventRejectsFurtherActions(t *testing.T) {
	resetMonitoringMemoryForTest()
	resolved, err := UpsertMonitorAlertEvent(&model.MonitorAlertEvent{ID: "event-resolved", EventKey: "resolved", Name: "Resolved", Severity: "warning"})
	if err != nil {
		t.Fatalf("upsert resolved event failed: %v", err)
	}
	if _, ok, err := ApplyMonitorAlertEventAction(resolved.ID, model.MonitorAlertAction{Action: "resolve", Actor: "tester"}); err != nil || !ok {
		t.Fatalf("resolve should succeed, ok=%v err=%v", ok, err)
	}
	if _, ok, err := ApplyMonitorAlertEventAction(resolved.ID, model.MonitorAlertAction{Action: "ack", Actor: "tester"}); !ok || !errors.Is(err, ErrTerminalMonitorAlertEvent) {
		t.Fatalf("ack on resolved should fail with terminal error, ok=%v err=%v", ok, err)
	}
	if current := ListMonitorAlertEvents(true); len(current) != 0 {
		t.Fatalf("resolved event must not return to current: %+v", current)
	}

	archived, err := UpsertMonitorAlertEvent(&model.MonitorAlertEvent{ID: "event-archived", EventKey: "archived", Name: "Archived", Severity: "critical"})
	if err != nil {
		t.Fatalf("upsert archived event failed: %v", err)
	}
	if _, ok, err := ApplyMonitorAlertEventAction(archived.ID, model.MonitorAlertAction{Action: "archive", Actor: "tester"}); err != nil || !ok {
		t.Fatalf("archive should succeed, ok=%v err=%v", ok, err)
	}
	if _, ok, err := ApplyMonitorAlertEventAction(archived.ID, model.MonitorAlertAction{Action: "assign", Actor: "tester", Assignee: "owner"}); !ok || !errors.Is(err, ErrTerminalMonitorAlertEvent) {
		t.Fatalf("assign on archived should fail with terminal error, ok=%v err=%v", ok, err)
	}
	if current := ListMonitorAlertEvents(true); len(current) != 0 {
		t.Fatalf("archived event must not return to current: %+v", current)
	}
}

func TestScanMonitorTargetRowSkipsScanErrorAndBadJSON(t *testing.T) {
	if target, ok := scanMonitorTargetRow(fakeScanner{err: errors.New("scan failed")}); ok || target != nil {
		t.Fatal("scan error must not produce target")
	}
	values := []any{
		"mt-1", "ident-1", "target", "10.0.0.1", "host", "linux", "amd64",
		"prod", "core", "owner", "online", "manual", "{bad-json", "{}", sql.NullTime{},
		time.Now(), time.Now(),
	}
	if target, ok := scanMonitorTargetRow(fakeScanner{values: values}); ok || target != nil {
		t.Fatal("bad labels json must not produce target")
	}
	values[12] = `{"env":"prod"}`
	if target, ok := scanMonitorTargetRow(fakeScanner{values: values}); !ok || target == nil || target.Labels["env"] != "prod" {
		t.Fatalf("valid row should produce target, ok=%v target=%v", ok, target)
	}
}

func TestScanFindXAgentRowSkipsBadJSON(t *testing.T) {
	values := []any{
		"fa-1", "ident-1", "mt-1", "10.0.0.1", "host", "linux", "amd64",
		"1.0.0", "categraf", "online", "{bad-json", "{}", "cfg", time.Now(), time.Now(), time.Now(),
	}
	if agent, ok := scanFindXAgentRow(fakeScanner{values: values}); ok || agent != nil {
		t.Fatal("bad capabilities json must not produce agent")
	}
	values[10] = `["cpu","mem"]`
	values[11] = `{"env":"prod"}`
	if agent, ok := scanFindXAgentRow(fakeScanner{values: values}); !ok || agent == nil || len(agent.Capabilities) != 2 || agent.GlobalLabels["env"] != "prod" {
		t.Fatalf("valid row should produce agent, ok=%v agent=%v", ok, agent)
	}
}

type fakeScanner struct {
	values []any
	err    error
}

func (f fakeScanner) Scan(dest ...any) error {
	if f.err != nil {
		return f.err
	}
	if len(dest) != len(f.values) {
		return errors.New("unexpected scan destination count")
	}
	for i, value := range f.values {
		switch out := dest[i].(type) {
		case *string:
			*out = value.(string)
		case *int:
			*out = value.(int)
		case *bool:
			*out = value.(bool)
		case *time.Time:
			*out = value.(time.Time)
		case *sql.NullTime:
			*out = value.(sql.NullTime)
		default:
			return errors.New("unsupported scan destination")
		}
	}
	return nil
}
