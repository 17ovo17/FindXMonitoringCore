package store

import (
	"errors"
	"sort"
	"strings"
	"time"

	"ai-workbench-api/internal/model"
)

var (
	alertInhibitRules          = map[string]*model.AlertInhibitRule{}
	ErrAlertInhibitNotFound    = errors.New("alert inhibit rule not found")
	ErrAlertInhibitValidation  = errors.New("alert inhibit rule validation failed")
)

// ListAlertInhibitRules 返回所有抑制规则，按更新时间倒序。
func ListAlertInhibitRules() []model.AlertInhibitRule {
	mu.RLock()
	defer mu.RUnlock()
	out := make([]model.AlertInhibitRule, 0, len(alertInhibitRules))
	for _, r := range alertInhibitRules {
		out = append(out, copyAlertInhibitRule(r))
	}
	sort.Slice(out, func(i, j int) bool { return out[i].UpdatedAt.After(out[j].UpdatedAt) })
	return out
}

// ListEnabledAlertInhibitRules 返回所有已启用的抑制规则。
func ListEnabledAlertInhibitRules() []model.AlertInhibitRule {
	mu.RLock()
	defer mu.RUnlock()
	out := make([]model.AlertInhibitRule, 0)
	for _, r := range alertInhibitRules {
		if r.Enabled {
			out = append(out, copyAlertInhibitRule(r))
		}
	}
	return out
}

// CreateAlertInhibitRule 创建一条新的抑制规则。
func CreateAlertInhibitRule(rule model.AlertInhibitRule) (model.AlertInhibitRule, error) {
	if strings.TrimSpace(rule.Name) == "" {
		return model.AlertInhibitRule{}, ErrAlertInhibitValidation
	}
	now := time.Now()
	rule.ID = NewID()
	rule.CreatedAt = now
	rule.UpdatedAt = now
	if rule.SourceMatch == nil {
		rule.SourceMatch = map[string]string{}
	}
	if rule.TargetMatch == nil {
		rule.TargetMatch = map[string]string{}
	}

	mu.Lock()
	cp := copyAlertInhibitRule(&rule)
	alertInhibitRules[cp.ID] = &cp
	mu.Unlock()

	return copyAlertInhibitRule(&rule), nil
}

// UpdateAlertInhibitRule 更新一条已有的抑制规则。
func UpdateAlertInhibitRule(id string, rule model.AlertInhibitRule) (model.AlertInhibitRule, error) {
	if strings.TrimSpace(rule.Name) == "" {
		return model.AlertInhibitRule{}, ErrAlertInhibitValidation
	}

	mu.Lock()
	defer mu.Unlock()

	existing, ok := alertInhibitRules[id]
	if !ok {
		return model.AlertInhibitRule{}, ErrAlertInhibitNotFound
	}

	rule.ID = id
	rule.CreatedAt = existing.CreatedAt
	rule.CreatedBy = existing.CreatedBy
	rule.UpdatedAt = time.Now()
	if rule.SourceMatch == nil {
		rule.SourceMatch = map[string]string{}
	}
	if rule.TargetMatch == nil {
		rule.TargetMatch = map[string]string{}
	}

	cp := copyAlertInhibitRule(&rule)
	alertInhibitRules[id] = &cp
	return copyAlertInhibitRule(&rule), nil
}

// DeleteAlertInhibitRule 删除一条抑制规则。
func DeleteAlertInhibitRule(id string) error {
	mu.Lock()
	defer mu.Unlock()
	if _, ok := alertInhibitRules[id]; !ok {
		return ErrAlertInhibitNotFound
	}
	delete(alertInhibitRules, id)
	return nil
}

func copyAlertInhibitRule(r *model.AlertInhibitRule) model.AlertInhibitRule {
	cp := *r
	if r.SourceMatch != nil {
		cp.SourceMatch = make(map[string]string, len(r.SourceMatch))
		for k, v := range r.SourceMatch {
			cp.SourceMatch[k] = v
		}
	}
	if r.TargetMatch != nil {
		cp.TargetMatch = make(map[string]string, len(r.TargetMatch))
		for k, v := range r.TargetMatch {
			cp.TargetMatch[k] = v
		}
	}
	if r.Equal != nil {
		cp.Equal = make([]string, len(r.Equal))
		copy(cp.Equal, r.Equal)
	}
	return cp
}
