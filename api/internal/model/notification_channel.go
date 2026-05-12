package model

import "time"

// NotificationChannel describes a delivery endpoint for alert events.
// It lives in the model package so that scheduler/notifier code can
// reference it without depending on the handler package.
type NotificationChannel struct {
	ID        string    `json:"id"`
	Type      string    `json:"type"`
	Name      string    `json:"name"`
	Endpoint  string    `json:"endpoint,omitempty"`
	Receiver  string    `json:"receiver,omitempty"`
	Secret    string    `json:"secret,omitempty"`
	Webhook   string    `json:"webhook"`
	Enabled   bool      `json:"enabled"`
	CreatedAt time.Time `json:"created_at,omitempty"`
	UpdatedAt time.Time `json:"updated_at,omitempty"`
}
