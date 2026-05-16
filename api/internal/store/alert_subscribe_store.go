package store

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"sort"
	"strings"
	"sync"
	"time"

	"ai-workbench-api/internal/model"
)

var (
	alertSubscribes   = map[string]*model.AlertSubscribe{}
	alertSubscribesMu sync.RWMutex

	ErrAlertSubscribeValidation = errors.New("alert subscribe validation failed")
	ErrAlertSubscribeNotFound   = errors.New("alert subscribe not found")
)

func newSubscribeID() string {
	b := make([]byte, 8)
	_, _ = rand.Read(b)
	return hex.EncodeToString(b)
}

// ListAlertSubscribes returns all subscriptions ordered by updated_at desc.
func ListAlertSubscribes() []model.AlertSubscribe {
	alertSubscribesMu.RLock()
	defer alertSubscribesMu.RUnlock()
	out := make([]model.AlertSubscribe, 0, len(alertSubscribes))
	for _, sub := range alertSubscribes {
		out = append(out, copyAlertSubscribe(sub))
	}
	sort.Slice(out, func(i, j int) bool { return out[i].UpdatedAt.After(out[j].UpdatedAt) })
	return out
}

// ListAlertSubscribesByUser returns subscriptions for a specific user.
func ListAlertSubscribesByUser(userID string) []model.AlertSubscribe {
	alertSubscribesMu.RLock()
	defer alertSubscribesMu.RUnlock()
	out := make([]model.AlertSubscribe, 0)
	for _, sub := range alertSubscribes {
		if sub.UserID == userID {
			out = append(out, copyAlertSubscribe(sub))
		}
	}
	sort.Slice(out, func(i, j int) bool { return out[i].UpdatedAt.After(out[j].UpdatedAt) })
	return out
}

// GetAlertSubscribe returns a subscription by ID.
func GetAlertSubscribe(id string) (*model.AlertSubscribe, bool) {
	alertSubscribesMu.RLock()
	defer alertSubscribesMu.RUnlock()
	sub, ok := alertSubscribes[id]
	if !ok {
		return nil, false
	}
	cp := copyAlertSubscribe(sub)
	return &cp, true
}

// CreateAlertSubscribe creates a new subscription.
func CreateAlertSubscribe(sub model.AlertSubscribe) (model.AlertSubscribe, error) {
	if strings.TrimSpace(sub.Name) == "" {
		return model.AlertSubscribe{}, ErrAlertSubscribeValidation
	}
	if strings.TrimSpace(sub.UserID) == "" {
		return model.AlertSubscribe{}, ErrAlertSubscribeValidation
	}
	if len(sub.ChannelIDs) == 0 {
		return model.AlertSubscribe{}, ErrAlertSubscribeValidation
	}
	now := time.Now()
	sub.ID = newSubscribeID()
	sub.CreatedAt = now
	sub.UpdatedAt = now
	alertSubscribesMu.Lock()
	cp := copyAlertSubscribe(&sub)
	alertSubscribes[sub.ID] = &cp
	alertSubscribesMu.Unlock()
	return copyAlertSubscribe(&sub), nil
}

// UpdateAlertSubscribe updates an existing subscription.
func UpdateAlertSubscribe(id string, sub model.AlertSubscribe) (model.AlertSubscribe, error) {
	if strings.TrimSpace(sub.Name) == "" {
		return model.AlertSubscribe{}, ErrAlertSubscribeValidation
	}
	if len(sub.ChannelIDs) == 0 {
		return model.AlertSubscribe{}, ErrAlertSubscribeValidation
	}
	alertSubscribesMu.Lock()
	defer alertSubscribesMu.Unlock()
	existing, ok := alertSubscribes[id]
	if !ok {
		return model.AlertSubscribe{}, ErrAlertSubscribeNotFound
	}
	sub.ID = id
	sub.UserID = existing.UserID
	sub.CreatedAt = existing.CreatedAt
	sub.UpdatedAt = time.Now()
	cp := copyAlertSubscribe(&sub)
	alertSubscribes[id] = &cp
	return copyAlertSubscribe(&sub), nil
}

// DeleteAlertSubscribe removes a subscription by ID.
func DeleteAlertSubscribe(id string) error {
	alertSubscribesMu.Lock()
	defer alertSubscribesMu.Unlock()
	if _, ok := alertSubscribes[id]; !ok {
		return ErrAlertSubscribeNotFound
	}
	delete(alertSubscribes, id)
	return nil
}

// MatchingSubscribes returns all enabled subscriptions that match the event.
func MatchingSubscribes(event *model.MonitorAlertEvent) []model.AlertSubscribe {
	if event == nil {
		return nil
	}
	alertSubscribesMu.RLock()
	defer alertSubscribesMu.RUnlock()
	var out []model.AlertSubscribe
	for _, sub := range alertSubscribes {
		if !sub.Enabled {
			continue
		}
		if !subscribeMatchesEvent(sub, event) {
			continue
		}
		out = append(out, copyAlertSubscribe(sub))
	}
	return out
}

func subscribeMatchesEvent(sub *model.AlertSubscribe, event *model.MonitorAlertEvent) bool {
	if len(sub.RuleIDs) > 0 {
		found := false
		for _, rid := range sub.RuleIDs {
			if strings.TrimSpace(rid) == event.RuleID {
				found = true
				break
			}
		}
		if !found {
			return false
		}
	}
	if len(sub.Severities) > 0 {
		severity := strings.ToLower(strings.TrimSpace(event.Severity))
		found := false
		for _, s := range sub.Severities {
			if strings.ToLower(strings.TrimSpace(s)) == severity {
				found = true
				break
			}
		}
		if !found {
			return false
		}
	}
	if len(sub.Labels) > 0 {
		for key, val := range sub.Labels {
			eventVal, exists := event.Labels[key]
			if !exists || eventVal != val {
				return false
			}
		}
	}
	return true
}

func copyAlertSubscribe(in *model.AlertSubscribe) model.AlertSubscribe {
	cp := *in
	if in.Labels != nil {
		cp.Labels = make(map[string]string, len(in.Labels))
		for k, v := range in.Labels {
			cp.Labels[k] = v
		}
	}
	cp.Severities = append([]string{}, in.Severities...)
	cp.RuleIDs = append([]string{}, in.RuleIDs...)
	cp.ChannelIDs = append([]string{}, in.ChannelIDs...)
	return cp
}
