package store

import (
	"errors"
	"sort"
	"strings"
	"time"

	"ai-workbench-api/internal/model"
)

var ErrNotificationValidation = errors.New("notification validation failed")

func ListNotificationRules() []model.NotificationRule {
	mu.RLock()
	defer mu.RUnlock()
	out := make([]model.NotificationRule, 0, len(notificationRules))
	for _, rule := range notificationRules {
		out = append(out, *copyNotificationRule(rule))
	}
	sort.Slice(out, func(i, j int) bool { return out[i].UpdatedAt.After(out[j].UpdatedAt) })
	return out
}

func GetNotificationRule(id string) (*model.NotificationRule, bool) {
	mu.RLock()
	defer mu.RUnlock()
	rule, ok := notificationRules[id]
	return copyNotificationRule(rule), ok
}

func SaveNotificationRules(items []model.NotificationRule, actor string) ([]model.NotificationRule, error) {
	out := make([]model.NotificationRule, 0, len(items))
	for i := range items {
		saved, err := SaveNotificationRule(&items[i], actor)
		if err != nil {
			return out, err
		}
		out = append(out, *saved)
	}
	return out, nil
}

func SaveNotificationRule(rule *model.NotificationRule, actor string) (*model.NotificationRule, error) {
	if strings.TrimSpace(rule.Name) == "" {
		return nil, ErrNotificationValidation
	}
	now := time.Now()
	cp := copyNotificationRule(rule)
	if cp.ID == "" {
		cp.ID = NewID()
		cp.CreatedAt = now
		cp.CreatedBy = actor
	} else if existing, ok := GetNotificationRule(cp.ID); ok {
		cp.CreatedAt = existing.CreatedAt
		cp.CreatedBy = existing.CreatedBy
	}
	cp.Name = strings.TrimSpace(cp.Name)
	cp.UpdatedAt = now
	cp.UpdatedBy = actor
	cp.NotifyConfigs = normalizeNotificationConfigs(cp.NotifyConfigs)
	if len(cp.NotifyConfigs) == 0 {
		return nil, ErrNotificationValidation
	}
	mu.Lock()
	notificationRules[cp.ID] = copyNotificationRule(cp)
	mu.Unlock()
	return copyNotificationRule(cp), nil
}

func DeleteNotificationRules(ids []string) int {
	mu.Lock()
	defer mu.Unlock()
	deleted := 0
	for _, id := range ids {
		if _, ok := notificationRules[id]; ok {
			delete(notificationRules, id)
			deleted++
		}
	}
	return deleted
}

func SetNotificationRuleEnabled(id string, enabled bool, actor string) (*model.NotificationRule, bool, error) {
	rule, ok := GetNotificationRule(id)
	if !ok {
		return nil, false, nil
	}
	rule.Enabled = enabled
	saved, err := SaveNotificationRule(rule, actor)
	return saved, true, err
}

func CloneNotificationRule(id, actor string) (*model.NotificationRule, bool, error) {
	rule, ok := GetNotificationRule(id)
	if !ok {
		return nil, false, nil
	}
	rule.ID = ""
	rule.Name = rule.Name + " Copy"
	rule.Enabled = false
	rule.CreatedAt = time.Time{}
	saved, err := SaveNotificationRule(rule, actor)
	return saved, true, err
}

func NotificationRuleStatistics(id string, days int) model.NotificationRuleStatistics {
	if days <= 0 {
		days = 7
	}
	rule, ok := GetNotificationRule(id)
	if !ok {
		return model.NotificationRuleStatistics{RuleID: id, Days: days}
	}
	alertIDs := map[string]bool{}
	for _, alertID := range rule.AlertRuleIDs {
		alertIDs[alertID] = true
	}
	events := ListMonitorAlertEvents(true)
	count := 0
	last := time.Time{}
	for _, event := range events {
		if alertIDs[event.RuleID] {
			count++
			if event.LastSeen.After(last) {
				last = event.LastSeen
			}
		}
	}
	out := model.NotificationRuleStatistics{RuleID: id, Days: days, Events: count, AlertRules: len(alertIDs), Deliveries: 0}
	if !last.IsZero() {
		out.LastEventAt = last.Format(time.RFC3339)
	}
	return out
}

func ListNotificationTemplates(channelIdent string) []model.NotificationTemplate {
	mu.RLock()
	defer mu.RUnlock()
	out := make([]model.NotificationTemplate, 0, len(notificationTemplates))
	for _, tpl := range notificationTemplates {
		if channelIdent != "" && tpl.NotifyChannelIdent != channelIdent {
			continue
		}
		out = append(out, *copyNotificationTemplate(tpl))
	}
	sort.Slice(out, func(i, j int) bool { return out[i].UpdatedAt.After(out[j].UpdatedAt) })
	return out
}

func GetNotificationTemplate(id string) (*model.NotificationTemplate, bool) {
	mu.RLock()
	defer mu.RUnlock()
	tpl, ok := notificationTemplates[id]
	return copyNotificationTemplate(tpl), ok
}

func SaveNotificationTemplates(items []model.NotificationTemplate, actor string) ([]model.NotificationTemplate, error) {
	out := make([]model.NotificationTemplate, 0, len(items))
	for i := range items {
		saved, err := SaveNotificationTemplate(&items[i], actor)
		if err != nil {
			return out, err
		}
		out = append(out, *saved)
	}
	return out, nil
}

func SaveNotificationTemplate(tpl *model.NotificationTemplate, actor string) (*model.NotificationTemplate, error) {
	if strings.TrimSpace(tpl.Name) == "" || strings.TrimSpace(tpl.NotifyChannelIdent) == "" {
		return nil, ErrNotificationValidation
	}
	now := time.Now()
	cp := copyNotificationTemplate(tpl)
	if cp.ID == "" {
		cp.ID = NewID()
		cp.CreatedAt = now
		cp.CreatedBy = actor
	} else if existing, ok := GetNotificationTemplate(cp.ID); ok {
		cp.CreatedAt = existing.CreatedAt
		cp.CreatedBy = existing.CreatedBy
	}
	if cp.Ident == "" {
		cp.Ident = "tpl-" + cp.ID
	}
	if cp.Content == nil {
		cp.Content = map[string]string{"content": ""}
	}
	cp.Name = strings.TrimSpace(cp.Name)
	cp.NotifyChannelIdent = strings.TrimSpace(cp.NotifyChannelIdent)
	cp.UpdatedAt = now
	cp.UpdatedBy = actor
	mu.Lock()
	notificationTemplates[cp.ID] = copyNotificationTemplate(cp)
	mu.Unlock()
	return copyNotificationTemplate(cp), nil
}

func DeleteNotificationTemplates(ids []string) int {
	mu.Lock()
	defer mu.Unlock()
	deleted := 0
	for _, id := range ids {
		if _, ok := notificationTemplates[id]; ok {
			delete(notificationTemplates, id)
			deleted++
		}
	}
	return deleted
}

func CloneNotificationTemplate(id, actor string) (*model.NotificationTemplate, bool, error) {
	tpl, ok := GetNotificationTemplate(id)
	if !ok {
		return nil, false, nil
	}
	tpl.ID = ""
	tpl.Ident = ""
	tpl.Name = tpl.Name + " Copy"
	tpl.Private = 1
	tpl.CreatedAt = time.Time{}
	saved, err := SaveNotificationTemplate(tpl, actor)
	return saved, true, err
}

func normalizeNotificationConfigs(items []model.NotificationConfig) []model.NotificationConfig {
	out := make([]model.NotificationConfig, 0, len(items))
	for _, item := range items {
		item.ChannelID = strings.TrimSpace(item.ChannelID)
		item.TemplateID = strings.TrimSpace(item.TemplateID)
		if item.ChannelID == "" {
			continue
		}
		out = append(out, item)
	}
	return out
}

func copyNotificationRule(in *model.NotificationRule) *model.NotificationRule {
	if in == nil {
		return nil
	}
	cp := *in
	cp.UserGroupIDs = append([]string{}, in.UserGroupIDs...)
	cp.AlertRuleIDs = append([]string{}, in.AlertRuleIDs...)
	cp.NotifyConfigs = copyNotificationConfigs(in.NotifyConfigs)
	cp.EventPipelineConfigs = append([]map[string]any{}, in.EventPipelineConfigs...)
	cp.Conditions = copyAnyMap(in.Conditions)
	cp.TimeWindow = copyAnyMap(in.TimeWindow)
	cp.Extra = copyAnyMap(in.Extra)
	return &cp
}

func copyNotificationConfigs(in []model.NotificationConfig) []model.NotificationConfig {
	out := make([]model.NotificationConfig, 0, len(in))
	for _, item := range in {
		cp := item
		cp.Params = copyAnyMap(item.Params)
		cp.Receivers = append([]string{}, item.Receivers...)
		cp.Severities = append([]string{}, item.Severities...)
		cp.TimeRanges = append([]map[string]any{}, item.TimeRanges...)
		cp.LabelKeys = append([]string{}, item.LabelKeys...)
		cp.Attributes = copyAnyMap(item.Attributes)
		out = append(out, cp)
	}
	return out
}

func copyNotificationTemplate(in *model.NotificationTemplate) *model.NotificationTemplate {
	if in == nil {
		return nil
	}
	cp := *in
	cp.UserGroupIDs = append([]string{}, in.UserGroupIDs...)
	cp.Content = copyStringMap(in.Content)
	return &cp
}

func copyAnyMap(in map[string]any) map[string]any {
	if in == nil {
		return nil
	}
	out := make(map[string]any, len(in))
	for key, value := range in {
		out[key] = value
	}
	return out
}
