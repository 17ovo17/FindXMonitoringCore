package model

import "time"

// AlertSubscribe allows a user to subscribe to alert events matching
// specific rules, labels, or severities and receive notifications
// through their preferred channels.
type AlertSubscribe struct {
	ID         string            `json:"id"`
	Name       string            `json:"name"`
	UserID     string            `json:"user_id"`
	Labels     map[string]string `json:"labels"`
	Severities []string          `json:"severities"`
	RuleIDs    []string          `json:"rule_ids"`
	ChannelIDs []string          `json:"channel_ids"`
	Enabled    bool              `json:"enabled"`
	CreatedAt  time.Time         `json:"created_at"`
	UpdatedAt  time.Time         `json:"updated_at"`
}
