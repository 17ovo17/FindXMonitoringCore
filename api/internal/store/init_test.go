package store

import (
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/spf13/viper"
)

const newIDRapidCallCount = 1000

func TestNewIDUniqueForRapidCalls(t *testing.T) {
	seen := make(map[string]struct{}, newIDRapidCallCount)
	for i := 0; i < newIDRapidCallCount; i++ {
		id := NewID()
		if id == "" {
			t.Fatalf("expected non-empty id")
		}
		if _, exists := seen[id]; exists {
			t.Fatalf("duplicate id generated at call %d: %s", i, id)
		}
		seen[id] = struct{}{}
	}
}

func TestStartupRetryStopsAfterSuccess(t *testing.T) {
	attempts := 0
	err := startupRetry{attempts: 3, interval: time.Nanosecond}.run("test retry", func() error {
		attempts++
		if attempts < 2 {
			return errors.New("temporary unavailable")
		}
		return nil
	})
	if err != nil {
		t.Fatalf("expected retry to recover, got %v", err)
	}
	if attempts != 2 {
		t.Fatalf("expected 2 attempts, got %d", attempts)
	}
}

func TestStartupRetryReturnsLastError(t *testing.T) {
	want := errors.New("still unavailable")
	attempts := 0
	err := startupRetry{attempts: 3, interval: time.Nanosecond}.run("test retry", func() error {
		attempts++
		return want
	})
	if !errors.Is(err, want) {
		t.Fatalf("expected last error %v, got %v", want, err)
	}
	if attempts != 3 {
		t.Fatalf("expected 3 attempts, got %d", attempts)
	}
}

func TestStorageRetryDurationConfig(t *testing.T) {
	viper.Set("test.retry_duration", "25ms")
	t.Cleanup(func() { viper.Set("test.retry_duration", "") })

	got := storageRetryDuration("test.retry_duration", time.Second)
	if got != 25*time.Millisecond {
		t.Fatalf("expected parsed duration 25ms, got %s", got)
	}
}

func TestStorageStartupRetryDefaultsInvalidValues(t *testing.T) {
	viper.Set("test.retry_attempts", 0)
	viper.Set("test.retry_interval", "-1s")
	t.Cleanup(func() {
		viper.Set("test.retry_attempts", "")
		viper.Set("test.retry_interval", "")
	})

	got := storageStartupRetry("test.retry_attempts", "test.retry_interval", 5, 10*time.Millisecond)
	if got.attempts != 5 {
		t.Fatalf("expected default attempts 5, got %d", got.attempts)
	}
	if got.interval != 10*time.Millisecond {
		t.Fatalf("expected default interval 10ms, got %s", got.interval)
	}
}

func TestCreateTableStatementsIncludeFindXAgentInstallExecutions(t *testing.T) {
	found := false
	for _, stmt := range createTableStatements {
		if strings.Contains(stmt, "findx_agent_install_executions") &&
			strings.Contains(stmt, "plan_id") &&
			strings.Contains(stmt, "steps JSON") {
			found = true
			break
		}
	}
	if !found {
		t.Fatal("expected findx_agent_install_executions table schema")
	}
}
