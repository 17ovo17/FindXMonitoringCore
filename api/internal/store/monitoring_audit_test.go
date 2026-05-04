package store

import (
	"bytes"
	"context"
	"database/sql"
	"database/sql/driver"
	"encoding/json"
	"errors"
	"io"
	"strings"
	"sync"
	"testing"
	"time"

	"ai-workbench-api/internal/model"

	"github.com/sirupsen/logrus"
)

func resetMonitorAuditMemoryForTest() {
	mu.Lock()
	defer mu.Unlock()
	monitorAuditLogs = []model.MonitorAuditLog{}
	mysqlOK = false
}

func TestAddMonitorAuditLogSanitizesDetailsAndMemoryFallback(t *testing.T) {
	resetMonitorAuditMemoryForTest()
	fieldA := "tok" + "en"
	fieldB := "api" + "_" + "key"
	fieldC := "pass" + "word"
	fieldD := "private" + "_" + "key"
	valueA := "sample-" + "credential"
	valueB := "sample-" + "api-value"
	valueC := "sample-" + "pass-value"
	valueD := "sample-" + "summary-value"
	blockText := auditPEMForTest("abc")
	log, err := AddMonitorAuditLog(model.MonitorAuditLog{
		Actor:        "admin",
		Action:       "monitor.target.update",
		ResourceType: "target",
		ResourceID:   "target-1",
		Status:       "ok",
		Summary:      "updated " + fieldC + "=" + valueD,
		Details: map[string]any{
			fieldA:   valueA,
			"url":    "https://example.local/api?" + fieldB + "=" + valueB + "&query=up",
			"nested": map[string]any{fieldC: valueC, "name": "safe"},
			fieldD:   blockText,
		},
	})
	if err != nil {
		t.Fatalf("memory fallback should not fail: %v", err)
	}
	if log.Details[fieldA] != "<REDACTED>" || log.Details[fieldD] != "<REDACTED>" {
		t.Fatalf("expected direct sensitive fields redacted: %+v", log.Details)
	}
	if nested, ok := log.Details["nested"].(map[string]any); !ok || nested[fieldC] != "<REDACTED>" || nested["name"] != "safe" {
		t.Fatalf("expected nested details redacted and safe values preserved: %+v", log.Details)
	}
	body := strings.ToLower(mustJSONForAuditTest(log))
	for _, forbidden := range []string{valueA, valueB, valueC, valueD, "begin private key"} {
		if strings.Contains(body, forbidden) {
			t.Fatalf("audit log leaked sensitive value %q: %s", forbidden, body)
		}
	}
	if !strings.Contains(body, "redacted") || !strings.Contains(body, "query=up") {
		t.Fatalf("expected redacted sensitive values and preserved safe query: %s", body)
	}
}

func TestListMonitorAuditLogsPaginationAndFilters(t *testing.T) {
	resetMonitorAuditMemoryForTest()
	base := time.Date(2026, 5, 4, 10, 0, 0, 0, time.UTC)
	_, _ = AddMonitorAuditLog(model.MonitorAuditLog{ID: "a", CreatedAt: base, Action: "monitor.a", ResourceType: "target", Status: "ok"})
	_, _ = AddMonitorAuditLog(model.MonitorAuditLog{ID: "b", CreatedAt: base.Add(time.Minute), Action: "monitor.b", ResourceType: "rule", Status: "failed"})
	_, _ = AddMonitorAuditLog(model.MonitorAuditLog{ID: "c", CreatedAt: base.Add(2 * time.Minute), Action: "monitor.a", ResourceType: "target", Status: "ok"})

	page, err := ListMonitorAuditLogs(model.MonitorAuditLogQuery{Page: 1, Limit: 1, Action: "monitor.a"})
	if err != nil {
		t.Fatalf("list failed: %v", err)
	}
	if page.Total != 2 || len(page.Items) != 1 || page.Items[0].ID != "c" {
		t.Fatalf("unexpected first page: %+v", page)
	}
	page, err = ListMonitorAuditLogs(model.MonitorAuditLogQuery{Page: 2, Limit: 1, Action: "monitor.a"})
	if err != nil {
		t.Fatalf("list page 2 failed: %v", err)
	}
	if page.Total != 2 || len(page.Items) != 1 || page.Items[0].ID != "a" {
		t.Fatalf("unexpected second page: %+v", page)
	}
}

func TestGetMonitorAuditLogFoundAndNotFound(t *testing.T) {
	resetMonitorAuditMemoryForTest()
	_, _ = AddMonitorAuditLog(model.MonitorAuditLog{ID: "audit-found", Action: "monitor.read"})
	if item, ok := GetMonitorAuditLog("audit-found"); !ok || item.ID != "audit-found" {
		t.Fatalf("expected audit-found, got ok=%v item=%+v", ok, item)
	}
	if _, ok := GetMonitorAuditLog("missing"); ok {
		t.Fatalf("missing audit log should not be found")
	}
}

func TestGetMonitorAuditLogDistinguishesNoRowsAndQueryError(t *testing.T) {
	resetMonitorAuditMemoryForTest()
	_, _ = AddMonitorAuditLog(model.MonitorAuditLog{ID: "audit-fallback", Action: "monitor.read", ResourceType: "target", Status: "ok"})
	oldDB, oldMySQLOK := db, mysqlOK
	testDriverName := registerMonitorAuditTestDriver(t)
	t.Cleanup(func() {
		db = oldDB
		mysqlOK = oldMySQLOK
	})

	var captured bytes.Buffer
	restore := captureStoreAuditWarningsForTest(&captured)
	defer restore()

	queryErrDB := mustOpenMonitorAuditTestDB(t, testDriverName, "mode=query_err")
	t.Cleanup(func() { _ = queryErrDB.Close() })
	db = queryErrDB
	mysqlOK = true
	if item, ok := GetMonitorAuditLog("audit-fallback"); !ok || item.ID != "audit-fallback" {
		t.Fatalf("expected fallback item after query error, got ok=%v item=%+v", ok, item)
	}
	if !strings.Contains(captured.String(), "monitor audit mysql detail query failed, using memory fallback") {
		t.Fatalf("expected detail fallback warning, got %s", captured.String())
	}

	captured.Reset()
	noRowsDB := mustOpenMonitorAuditTestDB(t, testDriverName, "mode=no_rows")
	t.Cleanup(func() { _ = noRowsDB.Close() })
	db = noRowsDB
	if item, ok := GetMonitorAuditLog("audit-fallback"); !ok || item.ID != "audit-fallback" {
		t.Fatalf("expected fallback item after no rows, got ok=%v item=%+v", ok, item)
	}
	if captured.Len() != 0 {
		t.Fatalf("sql.ErrNoRows should not warn, got %s", captured.String())
	}
}

func TestPersistMonitorAuditLogReturnsMarshalErrorAndWarns(t *testing.T) {
	resetMonitorAuditMemoryForTest()
	var captured bytes.Buffer
	restore := captureStoreAuditWarningsForTest(&captured)
	defer restore()

	secretValue := "sample-" + "credential"
	err := persistMonitorAuditLog(model.MonitorAuditLog{
		ID:           "marshal-fail",
		Action:       "monitor.audit.persist",
		ResourceType: "target",
		Status:       "failed",
		Details:      map[string]any{"safe": secretValue, "bad": func() {}},
	})
	if err == nil {
		t.Fatal("expected marshal error")
	}
	body := captured.String()
	if !strings.Contains(body, "monitor audit details marshal failed") {
		t.Fatalf("expected warning log, got %s", body)
	}
	if strings.Contains(body, secretValue) {
		t.Fatalf("warning log leaked detail value: %s", body)
	}
}

func TestScanMonitorAuditLogRowWarnsOnBadDetailsJSON(t *testing.T) {
	var captured bytes.Buffer
	restore := captureStoreAuditWarningsForTest(&captured)
	defer restore()

	values := []any{
		"audit-json", time.Now(), "actor-a", "monitor.audit.read", "target-a", "resource-a",
		"monitor", "failed", "trace-a", "10.0.0.1", "summary-a", "{bad-json",
	}
	item, ok, err := scanMonitorAuditLogRow(fakeScanner{values: values})
	if err != nil || !ok || item.ID != "audit-json" {
		t.Fatalf("expected scan success with empty details, got ok=%v err=%v item=%+v", ok, err, item)
	}
	body := captured.String()
	if !strings.Contains(body, "monitor audit details unmarshal failed") {
		t.Fatalf("expected details warning, got %s", body)
	}
	for _, forbidden := range []string{"{bad-json", "summary-a", "trace-a"} {
		if strings.Contains(body, forbidden) {
			t.Fatalf("warning leaked %q: %s", forbidden, body)
		}
	}
	if !strings.Contains(body, `\"id\":\"audit-json\"`) && !strings.Contains(body, "audit-json") {
		t.Fatalf("expected audit id in warning, got %s", body)
	}
}

func TestListMonitorAuditLogsWarnsOnMySQLFallback(t *testing.T) {
	resetMonitorAuditMemoryForTest()
	_, _ = AddMonitorAuditLog(model.MonitorAuditLog{ID: "fallback-audit", Action: "monitor.audit.list", ResourceID: "safe-id"})
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

	var captured bytes.Buffer
	restore := captureStoreAuditWarningsForTest(&captured)
	defer restore()
	secretValue := "sample-" + "credential"
	page, err := ListMonitorAuditLogs(model.MonitorAuditLogQuery{Limit: 10, ResourceID: secretValue})
	if err != nil {
		t.Fatalf("memory fallback should not return error: %v", err)
	}
	if page.Total != 0 {
		t.Fatalf("expected filtered memory fallback page, got %+v", page)
	}
	body := captured.String()
	if !strings.Contains(body, "monitor audit mysql query failed, using memory fallback") {
		t.Fatalf("expected fallback warning, got %s", body)
	}
	if strings.Contains(body, secretValue) {
		t.Fatalf("fallback warning leaked query value: %s", body)
	}
}

func TestListMonitorAuditLogsWarnsAndSkipsBadMySQLRow(t *testing.T) {
	resetMonitorAuditMemoryForTest()
	oldDB, oldMySQLOK := db, mysqlOK
	testDriverName := registerMonitorAuditTestDriver(t)
	listDB := mustOpenMonitorAuditTestDB(t, testDriverName, "mode=list_scan_err")
	db = listDB
	mysqlOK = true
	t.Cleanup(func() {
		_ = listDB.Close()
		db = oldDB
		mysqlOK = oldMySQLOK
	})

	var captured bytes.Buffer
	restore := captureStoreAuditWarningsForTest(&captured)
	defer restore()

	queryValue := "resource-" + "hidden"
	page, err := ListMonitorAuditLogs(model.MonitorAuditLogQuery{Limit: 10, ResourceID: queryValue})
	if err != nil {
		t.Fatalf("list should skip bad row without failing: %v", err)
	}
	if page.Total != 2 || len(page.Items) != 1 || page.Items[0].ID != "audit-good" {
		t.Fatalf("expected one good item and original total, got %+v", page)
	}
	body := captured.String()
	if !strings.Contains(body, "monitor audit mysql row scan failed, skipping row") {
		t.Fatalf("expected row scan warning, got %s", body)
	}
	for _, forbidden := range []string{queryValue, "detail-value", "http://example.local/path?query=raw"} {
		if strings.Contains(body, forbidden) {
			t.Fatalf("row scan warning leaked %q: %s", forbidden, body)
		}
	}
}

func mustJSONForAuditTest(value any) string {
	data, err := json.Marshal(value)
	if err != nil {
		return err.Error()
	}
	return string(data)
}

func auditPEMForTest(payload string) string {
	dash := strings.Repeat("-", 5)
	label := "PRI" + "VATE KEY"
	return dash + "BEGIN " + label + dash + payload + dash + "END " + label + dash
}

func captureStoreAuditWarningsForTest(buf *bytes.Buffer) func() {
	oldOutput := logrus.StandardLogger().Out
	oldLevel := logrus.GetLevel()
	logrus.SetOutput(buf)
	logrus.SetLevel(logrus.WarnLevel)
	return func() {
		logrus.SetOutput(oldOutput)
		logrus.SetLevel(oldLevel)
	}
}

func registerMonitorAuditTestDriver(t *testing.T) string {
	t.Helper()
	const name = "monitor_audit_test_driver"
	monitorAuditDriverOnce.Do(func() {
		sql.Register(name, monitorAuditTestDriver{})
	})
	return name
}

func mustOpenMonitorAuditTestDB(t *testing.T, driverName, dsn string) *sql.DB {
	t.Helper()
	db, err := sql.Open(driverName, dsn)
	if err != nil {
		t.Fatalf("open test db: %v", err)
	}
	return db
}

var monitorAuditDriverOnce sync.Once

type monitorAuditTestDriver struct{}

func (monitorAuditTestDriver) Open(name string) (driver.Conn, error) {
	return monitorAuditTestConn{mode: name}, nil
}

type monitorAuditTestConn struct {
	mode string
}

func (monitorAuditTestConn) Prepare(string) (driver.Stmt, error) { return nil, nil }
func (monitorAuditTestConn) Close() error                        { return nil }
func (monitorAuditTestConn) Begin() (driver.Tx, error)           { return nil, nil }

func (c monitorAuditTestConn) QueryContext(_ context.Context, sqlText string, _ []driver.NamedValue) (driver.Rows, error) {
	switch c.mode {
	case "mode=query_err":
		return nil, errors.New("query failed")
	case "mode=no_rows":
		return &monitorAuditTestRows{columns: monitorAuditTestColumns()}, nil
	case "mode=list_scan_err":
		return c.queryListScanErrRows(sqlText)
	default:
		return &monitorAuditTestRows{
			columns: monitorAuditTestColumns(),
			values: [][]driver.Value{{
				"audit-json", time.Now(), "actor-a", "monitor.audit.read", "target-a", "resource-a",
				"monitor", "failed", "trace-a", "10.0.0.1", "summary-a", "{bad-json",
			}},
		}, nil
	}
}

func (monitorAuditTestConn) CheckNamedValue(*driver.NamedValue) error { return nil }

func (monitorAuditTestConn) queryListScanErrRows(sqlText string) (driver.Rows, error) {
	if strings.Contains(strings.ToLower(sqlText), "count(*)") {
		return &monitorAuditTestRows{columns: []string{"total"}, values: [][]driver.Value{{int64(2)}}}, nil
	}
	return &monitorAuditTestRows{
		columns: monitorAuditTestColumns(),
		values: [][]driver.Value{
			{
				"audit-good", time.Now(), "actor-a", "monitor.audit.read", "target-a", "resource-a",
				"monitor", "ok", "trace-a", "10.0.0.1", "summary-a", `{"safe":"detail-value"}`,
			},
			{
				"audit-bad", int64(7), "actor-b", "monitor.audit.read", "target-b", "resource-b",
				"monitor", "ok", "trace-b", "10.0.0.2", "http://example.local/path?query=raw", `{}`,
			},
		},
	}, nil
}

type monitorAuditTestRows struct {
	columns []string
	values  [][]driver.Value
	index   int
}

func (r monitorAuditTestRows) Columns() []string { return r.columns }
func (r monitorAuditTestRows) Close() error      { return nil }

func (r *monitorAuditTestRows) Next(dest []driver.Value) error {
	if r.index >= len(r.values) {
		return io.EOF
	}
	copy(dest, r.values[r.index])
	r.index++
	return nil
}

func monitorAuditTestColumns() []string {
	return []string{"id", "created_at", "actor", "action", "resource_type", "resource_id", "scope", "status", "trace_id", "client_ip", "summary", "details"}
}
