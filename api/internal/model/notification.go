package model

import (
	"encoding/json"
	"time"
)

type NotificationRule struct {
	ID                   string               `json:"id"`
	Name                 string               `json:"name"`
	Description          string               `json:"description,omitempty"`
	Enabled              bool                 `json:"enabled"`
	UserGroupIDs         []string             `json:"user_group_ids,omitempty"`
	NotifyConfigs        []NotificationConfig `json:"notify_configs"`
	AlertRuleIDs         []string             `json:"alert_rule_ids,omitempty"`
	EventPipelineConfigs []map[string]any     `json:"event_pipeline_configs,omitempty"`
	Conditions           map[string]any       `json:"conditions,omitempty"`
	TimeWindow           map[string]any       `json:"time_window,omitempty"`
	Extra                map[string]any       `json:"extra,omitempty"`
	CreatedBy            string               `json:"created_by,omitempty"`
	UpdatedBy            string               `json:"updated_by,omitempty"`
	CreatedAt            time.Time            `json:"created_at"`
	UpdatedAt            time.Time            `json:"updated_at"`
}

type NotificationConfig struct {
	ChannelID  string           `json:"channel_id"`
	Channel    string           `json:"channel,omitempty"`
	TemplateID string           `json:"template_id,omitempty"`
	Params     map[string]any   `json:"params,omitempty"`
	Receivers  []string         `json:"receivers,omitempty"`
	Severities []string         `json:"severities,omitempty"`
	TimeRanges []map[string]any `json:"time_ranges,omitempty"`
	LabelKeys  []string         `json:"label_keys,omitempty"`
	Attributes map[string]any   `json:"attributes,omitempty"`
}

type NotificationTemplate struct {
	ID                 string            `json:"id"`
	Name               string            `json:"name"`
	Ident              string            `json:"ident"`
	NotifyChannelIdent string            `json:"notify_channel_ident"`
	UserGroupIDs       []string          `json:"user_group_ids,omitempty"`
	Private            int               `json:"private"`
	Content            map[string]string `json:"content"`
	CreatedBy          string            `json:"created_by,omitempty"`
	UpdatedBy          string            `json:"updated_by,omitempty"`
	CreatedAt          time.Time         `json:"created_at"`
	UpdatedAt          time.Time         `json:"updated_at"`
}

type NotificationRuleStatistics struct {
	RuleID      string `json:"rule_id"`
	Days        int    `json:"days"`
	Events      int    `json:"events"`
	AlertRules  int    `json:"alert_rules"`
	Deliveries  int    `json:"deliveries"`
	LastEventAt string `json:"last_event_at,omitempty"`
}

type NotificationRenderResult struct {
	Content map[string]string `json:"content"`
	Event   map[string]any    `json:"event"`
	Missing []string          `json:"missing,omitempty"`
}

func (r *NotificationRule) UnmarshalJSON(data []byte) error {
	type alias NotificationRule
	aux := struct {
		*alias
		Enable  *bool `json:"enable"`
		Enabled *bool `json:"enabled"`
	}{
		alias: (*alias)(r),
	}
	if err := json.Unmarshal(data, &aux); err != nil {
		return err
	}
	if aux.Enabled != nil {
		r.Enabled = *aux.Enabled
	} else if aux.Enable != nil {
		r.Enabled = *aux.Enable
	}
	return nil
}

func (r NotificationRule) MarshalJSON() ([]byte, error) {
	type alias NotificationRule
	return json.Marshal(struct {
		alias
		Enable bool `json:"enable"`
	}{
		alias:  alias(r),
		Enable: r.Enabled,
	})
}
