package notifier

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"
	"time"

	"ai-workbench-api/internal/model"
	"ai-workbench-api/internal/store"
)

func resetChannelsForTest(t *testing.T) {
	t.Helper()
	for _, ch := range store.ListNotificationChannels() {
		store.DeleteNotificationChannel(ch.ID)
	}
}

func TestSendToChannelLogTypeSucceeds(t *testing.T) {
	ch := &model.NotificationChannel{ID: "t1", Type: "log", Name: "log-ch", Enabled: true}
	event := BuildMockAlertEvent("log-ch")
	if err := SendToChannel(ch, event); err != nil {
		t.Fatalf("expected log dispatch to succeed, got %v", err)
	}
}

func TestSendToChannelWebhookPostsJSON(t *testing.T) {
	received := make(chan map[string]any, 1)
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		defer r.Body.Close()
		var payload map[string]any
		_ = json.Unmarshal(body, &payload)
		received <- payload
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	ch := &model.NotificationChannel{
		ID: "wh", Type: "webhook", Name: "wh", Enabled: true, Webhook: server.URL,
	}
	event := BuildMockAlertEvent("wh")
	if err := SendToChannel(ch, event); err != nil {
		t.Fatalf("webhook send failed: %v", err)
	}
	select {
	case body := <-received:
		if body["name"] != event.Name {
			t.Fatalf("webhook payload missing event name, got %#v", body)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("webhook payload not received in time")
	}
}

func TestSendToChannelUnsupportedReturnsError(t *testing.T) {
	ch := &model.NotificationChannel{ID: "x", Type: "pager", Name: "pager", Enabled: true}
	if err := SendToChannel(ch, BuildMockAlertEvent("pager")); err == nil {
		t.Fatal("expected unsupported channel type to return error")
	}
}

func TestDispatchAlertEventFiltersByAlertRuleID(t *testing.T) {
	resetChannelsForTest(t)

	hits := int32(0)
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		atomic.AddInt32(&hits, 1)
		w.WriteHeader(http.StatusNoContent)
	}))
	defer server.Close()

	ch := &model.NotificationChannel{ID: "ch1", Type: "webhook", Name: "ch1", Enabled: true, Webhook: server.URL}
	store.PutNotificationChannel(ch)
	t.Cleanup(func() { store.DeleteNotificationChannel(ch.ID) })

	rule := &model.NotificationRule{
		Name:          "match rule",
		Enabled:       true,
		AlertRuleIDs:  []string{"alert-001"},
		NotifyConfigs: []model.NotificationConfig{{ChannelID: "ch1"}},
	}
	saved, err := store.SaveNotificationRule(rule, "tester")
	if err != nil {
		t.Fatalf("save rule: %v", err)
	}
	t.Cleanup(func() { store.DeleteNotificationRules([]string{saved.ID}) })

	// Event for a different rule id should not fire.
	DispatchAlertEvent(&model.MonitorAlertEvent{Name: "other", RuleID: "alert-999", Severity: "info"})
	// Event for the matching rule id should fire.
	DispatchAlertEvent(&model.MonitorAlertEvent{Name: "match", RuleID: "alert-001", Severity: "info"})

	deadline := time.After(2 * time.Second)
	for atomic.LoadInt32(&hits) < 1 {
		select {
		case <-deadline:
			t.Fatalf("expected at least one webhook hit, got %d", atomic.LoadInt32(&hits))
		default:
			time.Sleep(25 * time.Millisecond)
		}
	}
	// give the filtered event a moment; it must not fire
	time.Sleep(150 * time.Millisecond)
	if got := atomic.LoadInt32(&hits); got != 1 {
		t.Fatalf("expected exactly 1 webhook hit, got %d", got)
	}
}
