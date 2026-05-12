package store

import (
	"sort"
	"sync"

	"ai-workbench-api/internal/model"
)

var (
	notificationChannels   = map[string]*model.NotificationChannel{}
	notificationChannelsMu sync.RWMutex
)

// ListNotificationChannels returns copies of all configured channels.
func ListNotificationChannels() []model.NotificationChannel {
	notificationChannelsMu.RLock()
	defer notificationChannelsMu.RUnlock()
	out := make([]model.NotificationChannel, 0, len(notificationChannels))
	for _, ch := range notificationChannels {
		out = append(out, *ch)
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Name < out[j].Name })
	return out
}

// ListActiveNotificationChannels returns only enabled channels.
func ListActiveNotificationChannels() []model.NotificationChannel {
	all := ListNotificationChannels()
	out := make([]model.NotificationChannel, 0, len(all))
	for _, ch := range all {
		if ch.Enabled {
			out = append(out, ch)
		}
	}
	return out
}

// GetNotificationChannel returns the channel for the given ID.
func GetNotificationChannel(id string) (*model.NotificationChannel, bool) {
	notificationChannelsMu.RLock()
	defer notificationChannelsMu.RUnlock()
	ch, ok := notificationChannels[id]
	if !ok {
		return nil, false
	}
	cp := *ch
	return &cp, true
}

// PutNotificationChannel stores or updates the channel.
func PutNotificationChannel(ch *model.NotificationChannel) {
	if ch == nil || ch.ID == "" {
		return
	}
	notificationChannelsMu.Lock()
	defer notificationChannelsMu.Unlock()
	cp := *ch
	notificationChannels[cp.ID] = &cp
}

// DeleteNotificationChannel removes the channel; returns true if deleted.
func DeleteNotificationChannel(id string) bool {
	notificationChannelsMu.Lock()
	defer notificationChannelsMu.Unlock()
	if _, ok := notificationChannels[id]; !ok {
		return false
	}
	delete(notificationChannels, id)
	return true
}

// ListActiveNotificationRules returns enabled notification rules only.
func ListActiveNotificationRules() []model.NotificationRule {
	rules := ListNotificationRules()
	out := make([]model.NotificationRule, 0, len(rules))
	for _, rule := range rules {
		if rule.Enabled {
			out = append(out, rule)
		}
	}
	return out
}
