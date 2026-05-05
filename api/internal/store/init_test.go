package store

import "testing"

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
