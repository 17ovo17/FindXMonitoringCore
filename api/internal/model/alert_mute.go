package model

import (
	"strings"
	"time"
)

// AlertMute defines a rule that suppresses alert notifications for matching
// events during a configured time window.
type AlertMute struct {
	ID         string            `json:"id"`
	Name       string            `json:"name"`
	Labels     map[string]string `json:"labels"`      // label matchers (key=value)
	Severities []string          `json:"severities"`  // empty = all severities
	RuleIDs    []string          `json:"rule_ids"`    // empty = all rules
	StartTime  int64             `json:"start_time"`  // unix timestamp
	EndTime    int64             `json:"end_time"`    // unix timestamp, 0 = permanent
	Reason     string            `json:"reason"`
	CreatedBy  string            `json:"created_by"`
	Enabled    bool              `json:"enabled"`
	CreatedAt  time.Time         `json:"created_at"`
	UpdatedAt  time.Time         `json:"updated_at"`
}

// alertMuteProvider is set by the store package to supply active mutes
// without creating a circular import.
var alertMuteProvider func() []AlertMute

// RegisterAlertMuteProvider allows the store layer to inject the active-mute
// listing function at init time.
func RegisterAlertMuteProvider(fn func() []AlertMute) {
	alertMuteProvider = fn
}

// IsAlertMuted returns true if the given event matches any active mute rule.
func IsAlertMuted(event *MonitorAlertEvent) bool {
	if event == nil {
		return false
	}
	if alertMuteProvider == nil {
		return false
	}
	mutes := alertMuteProvider()
	for i := range mutes {
		if alertMuteMatches(&mutes[i], event) {
			return true
		}
	}
	return false
}

func alertMuteMatches(mute *AlertMute, event *MonitorAlertEvent) bool {
	// 检查 severity 过滤
	if len(mute.Severities) > 0 {
		matched := false
		evSev := strings.ToLower(strings.TrimSpace(event.Severity))
		for _, s := range mute.Severities {
			if strings.ToLower(strings.TrimSpace(s)) == evSev {
				matched = true
				break
			}
		}
		if !matched {
			return false
		}
	}

	// 检查 rule_id 过滤
	if len(mute.RuleIDs) > 0 {
		matched := false
		for _, id := range mute.RuleIDs {
			if strings.TrimSpace(id) == event.RuleID {
				matched = true
				break
			}
		}
		if !matched {
			return false
		}
	}

	// 检查 label 匹配（所有 mute label 必须在 event label 中存在且相等）
	for k, v := range mute.Labels {
		ev, ok := event.Labels[k]
		if !ok || strings.TrimSpace(ev) != strings.TrimSpace(v) {
			return false
		}
	}

	return true
}
