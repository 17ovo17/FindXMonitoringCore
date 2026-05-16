package store

import (
	"errors"
	"sort"
	"strings"
	"time"

	"ai-workbench-api/internal/model"
)

var (
	alertMutes            = map[string]*model.AlertMute{}
	ErrAlertMuteNotFound  = errors.New("alert mute not found")
	ErrAlertMuteValidation = errors.New("alert mute validation failed")
)

func init() {
	model.RegisterAlertMuteProvider(ListActiveAlertMutes)
}

// ListAlertMutes returns all mute rules ordered by updated_at descending.
func ListAlertMutes() []model.AlertMute {
	mu.RLock()
	defer mu.RUnlock()
	out := make([]model.AlertMute, 0, len(alertMutes))
	for _, m := range alertMutes {
		out = append(out, copyAlertMute(m))
	}
	sort.Slice(out, func(i, j int) bool { return out[i].UpdatedAt.After(out[j].UpdatedAt) })
	return out
}

// GetAlertMute returns a single mute by ID.
func GetAlertMute(id string) (*model.AlertMute, bool) {
	mu.RLock()
	defer mu.RUnlock()
	m, ok := alertMutes[id]
	if !ok {
		return nil, false
	}
	cp := copyAlertMute(m)
	return &cp, true
}

// CreateAlertMute validates and persists a new mute rule.
func CreateAlertMute(mute model.AlertMute) (model.AlertMute, error) {
	if strings.TrimSpace(mute.Name) == "" {
		return model.AlertMute{}, ErrAlertMuteValidation
	}
	now := time.Now()
	mute.ID = NewID()
	mute.CreatedAt = now
	mute.UpdatedAt = now
	if mute.Labels == nil {
		mute.Labels = map[string]string{}
	}

	mu.Lock()
	cp := copyAlertMute(&mute)
	alertMutes[cp.ID] = &cp
	mu.Unlock()

	return copyAlertMute(&mute), nil
}

// UpdateAlertMute replaces an existing mute rule by ID.
func UpdateAlertMute(id string, mute model.AlertMute) (model.AlertMute, error) {
	if strings.TrimSpace(mute.Name) == "" {
		return model.AlertMute{}, ErrAlertMuteValidation
	}

	mu.Lock()
	defer mu.Unlock()

	existing, ok := alertMutes[id]
	if !ok {
		return model.AlertMute{}, ErrAlertMuteNotFound
	}

	mute.ID = id
	mute.CreatedAt = existing.CreatedAt
	mute.CreatedBy = existing.CreatedBy
	mute.UpdatedAt = time.Now()
	if mute.Labels == nil {
		mute.Labels = map[string]string{}
	}

	cp := copyAlertMute(&mute)
	alertMutes[id] = &cp
	return copyAlertMute(&mute), nil
}

// DeleteAlertMute removes a mute rule by ID.
func DeleteAlertMute(id string) error {
	mu.Lock()
	defer mu.Unlock()
	if _, ok := alertMutes[id]; !ok {
		return ErrAlertMuteNotFound
	}
	delete(alertMutes, id)
	return nil
}

// ListActiveAlertMutes returns mutes that are enabled and within their time window.
func ListActiveAlertMutes() []model.AlertMute {
	mu.RLock()
	defer mu.RUnlock()
	now := time.Now().Unix()
	out := make([]model.AlertMute, 0)
	for _, m := range alertMutes {
		if !m.Enabled {
			continue
		}
		if m.StartTime > 0 && now < m.StartTime {
			continue
		}
		if m.EndTime > 0 && now > m.EndTime {
			continue
		}
		out = append(out, copyAlertMute(m))
	}
	return out
}

func copyAlertMute(m *model.AlertMute) model.AlertMute {
	cp := *m
	if m.Labels != nil {
		cp.Labels = make(map[string]string, len(m.Labels))
		for k, v := range m.Labels {
			cp.Labels[k] = v
		}
	}
	if m.Severities != nil {
		cp.Severities = make([]string, len(m.Severities))
		copy(cp.Severities, m.Severities)
	}
	if m.RuleIDs != nil {
		cp.RuleIDs = make([]string, len(m.RuleIDs))
		copy(cp.RuleIDs, m.RuleIDs)
	}
	return cp
}
