package store

import (
	"database/sql"
	"testing"
	"time"

	"ai-workbench-api/internal/model"
)

func resetFindXAgentLifecycleMemoryForTest() {
	mu.Lock()
	defer mu.Unlock()
	findxAgentInstallPlans = map[string]*model.FindXAgentInstallPlan{}
	resetFindXAgentInstallExecutionsLocked()
	findxAgentConfigRollouts = map[string]*model.FindXAgentConfigRollout{}
	findxAgentExecutionTasks = map[string]*model.FindXAgentExecutionTask{}
	findxAgentDataArrivalEvidence = map[string]*model.FindXAgentDataArrivalEvidence{}
	mysqlOK = false
}

func TestFindXAgentLifecycleDecodeRejectsBadJSON(t *testing.T) {
	var values []string
	if err := decodeLifecycleJSON(`{bad-json`, &values); err == nil {
		t.Fatal("expected bad lifecycle json to return error")
	}
}

func TestFindXAgentInstallExecutionMemoryFallbackListSortsAndCopies(t *testing.T) {
	resetFindXAgentLifecycleMemoryForTest()
	now := time.Now()
	blocked, err := SaveFindXAgentInstallExecution(model.FindXAgentInstallExecution{
		PlanID:   "plan-old",
		TargetID: "target-old",
		Runner:   "ssh",
		Status:   model.FindXAgentExecutionStatusBlocked,
		Steps: []model.FindXAgentInstallExecutionStep{{
			Name:      "preflight",
			Status:    model.FindXAgentExecutionStatusBlocked,
			UpdatedAt: now,
		}},
		EvidenceRefs: []string{"install-plan:plan-old"},
	})
	if err != nil {
		t.Fatalf("save blocked execution: %v", err)
	}
	time.Sleep(time.Millisecond)
	nonBlocked, err := SaveFindXAgentInstallExecution(model.FindXAgentInstallExecution{
		PlanID:       "plan-new",
		TargetID:     "target-new",
		Runner:       "local",
		Status:       model.FindXAgentExecutionStatusFailed,
		EvidenceRefs: []string{"install-plan:plan-new"},
	})
	if err != nil {
		t.Fatalf("save non-blocked execution: %v", err)
	}

	items, err := ListFindXAgentInstallExecutions()
	if err != nil {
		t.Fatalf("list executions: %v", err)
	}
	if len(items) != 2 || items[0].ID != nonBlocked.ID || items[1].ID != blocked.ID {
		t.Fatalf("expected non-blocked input coerced to blocked and sorted, got %#v", items)
	}
	for _, item := range items {
		if item.Status != model.FindXAgentExecutionStatusBlocked || item.ErrorSummary == "" {
			t.Fatalf("execution must be stored as blocked with summary, got %#v", item)
		}
	}
	items[0].EvidenceRefs[0] = "mutated"
	again, err := ListFindXAgentInstallExecutions()
	if err != nil {
		t.Fatalf("list executions again: %v", err)
	}
	if again[0].EvidenceRefs[0] != "install-plan:plan-new" {
		t.Fatalf("expected copied evidence refs, got %#v", again[0])
	}
}

func TestFindXAgentInstallExecutionMemoryFallbackCoercesAndReadsBlockedOnly(t *testing.T) {
	resetFindXAgentLifecycleMemoryForTest()
	blocked, err := SaveFindXAgentInstallExecution(model.FindXAgentInstallExecution{
		PlanID:       "plan-blocked",
		TargetID:     "target-blocked",
		Runner:       "ssh",
		Status:       model.FindXAgentExecutionStatusBlocked,
		Steps:        []model.FindXAgentInstallExecutionStep{{Name: "preflight", Status: model.FindXAgentExecutionStatusBlocked, UpdatedAt: time.Now()}},
		EvidenceRefs: []string{"install-plan:plan-blocked"},
	})
	if err != nil {
		t.Fatalf("save blocked execution: %v", err)
	}
	nonBlocked, err := SaveFindXAgentInstallExecution(model.FindXAgentInstallExecution{
		PlanID:       "plan-running",
		TargetID:     "target-running",
		Runner:       "ssh",
		Status:       model.FindXAgentExecutionStatusRunning,
		Steps:        []model.FindXAgentInstallExecutionStep{{Name: "execute", Status: model.FindXAgentExecutionStatusSucceeded, UpdatedAt: time.Now()}},
		EvidenceRefs: []string{"install-plan:plan-running"},
	})
	if err != nil {
		t.Fatalf("save non-blocked execution: %v", err)
	}
	if nonBlocked.Status != model.FindXAgentExecutionStatusBlocked || nonBlocked.ErrorSummary == "" {
		t.Fatalf("non-blocked execution input must be coerced to blocked, got %#v", nonBlocked)
	}
	if len(nonBlocked.Steps) != 1 || nonBlocked.Steps[0].Status != model.FindXAgentExecutionStatusBlocked {
		t.Fatalf("non-blocked execution step must be coerced to blocked, got %#v", nonBlocked.Steps)
	}

	items, err := ListFindXAgentInstallExecutions()
	if err != nil {
		t.Fatalf("list executions: %v", err)
	}
	if len(items) != 2 {
		t.Fatalf("expected blocked-only list, got %#v", items)
	}

	item, ok, err := GetFindXAgentInstallExecution(blocked.ID)
	if err != nil {
		t.Fatalf("get blocked execution: %v", err)
	}
	if !ok || item.ID != blocked.ID || item.Status != model.FindXAgentExecutionStatusBlocked {
		t.Fatalf("expected blocked execution detail, got ok=%v item=%#v", ok, item)
	}

	coerced, ok, err := GetFindXAgentInstallExecution(nonBlocked.ID)
	if err != nil {
		t.Fatalf("get coerced execution: %v", err)
	}
	if !ok || coerced.Status != model.FindXAgentExecutionStatusBlocked || coerced.ErrorSummary == "" {
		t.Fatalf("expected coerced blocked execution detail, got ok=%v item=%#v", ok, coerced)
	}
	if _, ok, err := GetFindXAgentInstallExecution("missing-execution"); err != nil {
		t.Fatalf("get missing execution: %v", err)
	} else if ok {
		t.Fatalf("expected missing execution to stay hidden")
	}
}

func TestFindXAgentLifecycleMemoryFallbackListSortsAndReadsBack(t *testing.T) {
	resetFindXAgentLifecycleMemoryForTest()
	older, err := SaveFindXAgentInstallPlan(model.FindXAgentInstallPlan{
		PackageID: "agent-core",
		OS:        "linux",
		Method:    "linux-curl",
		TargetIDs: []string{"target-old"},
		Status:    "blocked",
		Metadata:  map[string]string{"ticket": "CHG-1"},
	})
	if err != nil {
		t.Fatalf("save older install plan: %v", err)
	}
	time.Sleep(time.Millisecond)
	newer, err := SaveFindXAgentInstallPlan(model.FindXAgentInstallPlan{
		PackageID: "host-collector",
		OS:        "windows",
		Method:    "powershell",
		TargetIDs: []string{"target-new"},
		Status:    model.FindXAgentExecutionStatusRunning,
		Metadata:  map[string]string{"ticket": "CHG-2"},
	})
	if err != nil {
		t.Fatalf("save newer install plan: %v", err)
	}

	items, err := ListFindXAgentInstallPlans()
	if err != nil {
		t.Fatalf("list install plans: %v", err)
	}
	if len(items) != 2 {
		t.Fatalf("expected two install plans, got %d", len(items))
	}
	if items[0].ID != newer.ID || items[1].ID != older.ID {
		t.Fatalf("expected updated_at desc order, got %#v", items)
	}
	if items[0].TargetIDs[0] != "target-new" || items[0].Metadata["ticket"] != "CHG-2" {
		t.Fatalf("expected read back newer plan data, got %#v", items[0])
	}
	if items[0].Status != findXAgentBlockedStatus || items[0].Blocker != findXAgentExecutorBlockedReason {
		t.Fatalf("expected install plan to be coerced to blocked, got %#v", items[0])
	}
	rollout, err := SaveFindXAgentConfigRollout(model.FindXAgentConfigRollout{
		TemplateID: "host-plugin",
		TargetIDs:  []string{"target-new"},
		Status:     model.FindXAgentExecutionStatusSucceeded,
	})
	if err != nil {
		t.Fatalf("save config rollout: %v", err)
	}
	if rollout.Status != findXAgentBlockedStatus || rollout.Blocker != findXAgentExecutorBlockedReason {
		t.Fatalf("expected config rollout to be coerced to blocked, got %#v", rollout)
	}
}

func TestScanFindXAgentInstallExecutionCoercesStaleStepStates(t *testing.T) {
	item, err := scanFindXAgentInstallExecution(staleInstallExecutionRow{now: time.Now()})
	if err != nil {
		t.Fatalf("scan stale install execution: %v", err)
	}
	if item.Status != model.FindXAgentExecutionStatusBlocked || item.ErrorSummary != findXAgentExecutorBlockedReason {
		t.Fatalf("expected stale execution row to be blocked-safe, got %#v", item)
	}
	if len(item.Steps) != 2 || item.Steps[0].Status != model.FindXAgentExecutionStatusBlocked || item.Steps[1].Status != model.FindXAgentExecutionStatusBlocked {
		t.Fatalf("expected stale step states to be coerced to blocked, got %#v", item.Steps)
	}
}

type staleInstallExecutionRow struct {
	now time.Time
}

func (row staleInstallExecutionRow) Scan(dest ...any) error {
	*dest[0].(*string) = "execution-stale"
	*dest[1].(*string) = "plan-stale"
	*dest[2].(*string) = "target-stale"
	*dest[3].(*string) = "ssh"
	*dest[4].(*string) = model.FindXAgentExecutionStatusBlocked
	*dest[5].(*sql.NullInt64) = sql.NullInt64{}
	*dest[6].(*string) = `[{"name":"preflight","status":"running"},{"name":"verify","status":"succeeded"}]`
	*dest[7].(*string) = `["install-plan:plan-stale"]`
	*dest[8].(*string) = ""
	*dest[9].(*time.Time) = row.now
	*dest[10].(*sql.NullTime) = sql.NullTime{}
	*dest[11].(*sql.NullTime) = sql.NullTime{}
	*dest[12].(*time.Time) = row.now
	return nil
}

func TestFindXAgentExecutionTaskMemoryFallbackCoercesToBlocked(t *testing.T) {
	resetFindXAgentLifecycleMemoryForTest()
	task, err := SaveFindXAgentExecutionTask(model.FindXAgentExecutionTask{
		Action:    " Restart ",
		TargetIDs: []string{"target-a"},
		Status:    model.FindXAgentExecutionStatusSucceeded,
	})
	if err != nil {
		t.Fatalf("save execution task: %v", err)
	}
	if task.Status != model.FindXAgentExecutionStatusBlocked || task.Blocker != findXAgentExecutorBlockedReason {
		t.Fatalf("execution task must be coerced to blocked, got %#v", task)
	}

	items, err := ListFindXAgentExecutionTasks()
	if err != nil {
		t.Fatalf("list execution tasks: %v", err)
	}
	if len(items) != 1 || items[0].ID != task.ID || items[0].Status != model.FindXAgentExecutionStatusBlocked {
		t.Fatalf("expected blocked task in list, got %#v", items)
	}
	detail, ok, err := GetFindXAgentExecutionTask(task.ID)
	if err != nil {
		t.Fatalf("get execution task: %v", err)
	}
	if !ok || detail.Status != model.FindXAgentExecutionStatusBlocked || detail.Blocker != findXAgentExecutorBlockedReason {
		t.Fatalf("expected blocked task detail, got ok=%v item=%#v", ok, detail)
	}
}

func TestSaveFindXAgentDataArrivalEvidenceDoesNotCacheOnPersistFailure(t *testing.T) {
	resetFindXAgentLifecycleMemoryForTest()
	previousDB, previousMySQLOK := db, mysqlOK
	failingDB, err := sql.Open("mysql", "invalid:invalid@tcp(127.0.0.1:1)/invalid?timeout=1ms")
	if err != nil {
		t.Fatalf("open failing db: %v", err)
	}
	t.Cleanup(func() {
		failingDB.Close()
		resetFindXAgentLifecycleMemoryForTest()
		db = previousDB
		mysqlOK = previousMySQLOK
	})

	db = failingDB
	mysqlOK = true
	_, err = SaveFindXAgentDataArrivalEvidence(model.FindXAgentDataArrivalEvidence{
		Kind:   model.FindXAgentDataArrivalKindLogs,
		Status: model.FindXAgentDataArrivalStatusReported,
	})
	if err == nil {
		t.Fatal("expected persistence failure")
	}
	mu.RLock()
	cached := len(findxAgentDataArrivalEvidence)
	mu.RUnlock()
	if cached != 0 {
		t.Fatalf("failed persistence must not leave reported memory evidence, got %d entries", cached)
	}
}
