package store

import (
	"sort"
	"sync"
	"time"

	"ai-workbench-api/internal/model"
)

var (
	notificationChannels   = map[string]*model.NotificationChannel{}
	notificationChannelsMu sync.RWMutex
)

// ListNotificationChannels returns copies of all configured channels.
func ListNotificationChannels() []model.NotificationChannel {
	if mysqlOK {
		if channels, err := loadNotificationChannelsFromDB(); err == nil {
			return channels
		}
	}
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
	ch, ok := notificationChannels[id]
	notificationChannelsMu.RUnlock()
	if ok {
		cp := *ch
		return &cp, true
	}
	if mysqlOK {
		return loadNotificationChannelFromDB(id)
	}
	return nil, false
}

// PutNotificationChannel stores or updates the channel.
func PutNotificationChannel(ch *model.NotificationChannel) {
	if ch == nil || ch.ID == "" {
		return
	}
	notificationChannelsMu.Lock()
	cp := *ch
	notificationChannels[cp.ID] = &cp
	notificationChannelsMu.Unlock()
	if mysqlOK {
		persistNotificationChannel(&cp)
	}
}

// DeleteNotificationChannel removes the channel; returns true if deleted.
func DeleteNotificationChannel(id string) bool {
	notificationChannelsMu.Lock()
	_, ok := notificationChannels[id]
	if ok {
		delete(notificationChannels, id)
	}
	notificationChannelsMu.Unlock()
	if mysqlOK {
		db.Exec(`DELETE FROM notification_channels WHERE id=?`, id)
	}
	return ok
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

func persistNotificationChannel(ch *model.NotificationChannel) {
	if ch.CreatedAt.IsZero() {
		ch.CreatedAt = time.Now()
	}
	ch.UpdatedAt = time.Now()
	_, _ = db.Exec(`INSERT INTO notification_channels (id,name,type,webhook,endpoint,enabled,config,created_at,updated_at) VALUES (?,?,?,?,?,?,'{}',?,?) ON DUPLICATE KEY UPDATE name=VALUES(name),type=VALUES(type),webhook=VALUES(webhook),endpoint=VALUES(endpoint),enabled=VALUES(enabled),updated_at=VALUES(updated_at)`,
		ch.ID, ch.Name, ch.Type, ch.Webhook, ch.Endpoint, ch.Enabled, ch.CreatedAt, ch.UpdatedAt)
}

func loadNotificationChannelsFromDB() ([]model.NotificationChannel, error) {
	rows, err := db.Query(`SELECT id,name,type,COALESCE(webhook,''),COALESCE(endpoint,''),enabled,created_at,updated_at FROM notification_channels ORDER BY name`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []model.NotificationChannel
	for rows.Next() {
		var ch model.NotificationChannel
		if err := rows.Scan(&ch.ID, &ch.Name, &ch.Type, &ch.Webhook, &ch.Endpoint, &ch.Enabled, &ch.CreatedAt, &ch.UpdatedAt); err != nil {
			continue
		}
		out = append(out, ch)
	}
	return out, nil
}

func loadNotificationChannelFromDB(id string) (*model.NotificationChannel, bool) {
	row := db.QueryRow(`SELECT id,name,type,COALESCE(webhook,''),COALESCE(endpoint,''),enabled,created_at,updated_at FROM notification_channels WHERE id=?`, id)
	var ch model.NotificationChannel
	if err := row.Scan(&ch.ID, &ch.Name, &ch.Type, &ch.Webhook, &ch.Endpoint, &ch.Enabled, &ch.CreatedAt, &ch.UpdatedAt); err != nil {
		return nil, false
	}
	return &ch, true
}
