package scheduler

import (
	"strings"
	"sync"

	"ai-workbench-api/internal/model"
	"ai-workbench-api/internal/store"

	log "github.com/sirupsen/logrus"
)

// InhibitRule 定义抑制规则（运行时视图）。
type InhibitRule struct {
	ID          string            `json:"id"`
	SourceMatch map[string]string `json:"source_match"`
	TargetMatch map[string]string `json:"target_match"`
	Equal       []string          `json:"equal"`
	Enabled     bool              `json:"enabled"`
}

// InhibitManager 管理抑制规则，提供运行时抑制判断。
type InhibitManager struct {
	rules []InhibitRule
	mu    sync.RWMutex
}

// globalInhibitManager 全局单例。
var globalInhibitManager = &InhibitManager{}

// GetInhibitManager 返回全局 InhibitManager 实例。
func GetInhibitManager() *InhibitManager {
	return globalInhibitManager
}

// Reload 从 store 重新加载已启用的抑制规则。
func (m *InhibitManager) Reload() {
	rules := store.ListEnabledAlertInhibitRules()
	converted := make([]InhibitRule, 0, len(rules))
	for _, r := range rules {
		converted = append(converted, InhibitRule{
			ID:          r.ID,
			SourceMatch: r.SourceMatch,
			TargetMatch: r.TargetMatch,
			Equal:       r.Equal,
			Enabled:     r.Enabled,
		})
	}
	m.mu.Lock()
	m.rules = converted
	m.mu.Unlock()
}

// IsInhibited 检查告警是否被抑制。
// 逻辑：如果存在一条 firing 的源告警匹配 source_match，
//
//	且当前告警匹配 target_match，
//	且 equal 标签值相同，
//	则当前告警被抑制（不发送通知）。
func (m *InhibitManager) IsInhibited(event *model.MonitorAlertEvent, firingEvents []*model.MonitorAlertEvent) bool {
	if event == nil {
		return false
	}
	m.mu.RLock()
	rules := m.rules
	m.mu.RUnlock()

	if len(rules) == 0 {
		return false
	}

	for _, rule := range rules {
		if !rule.Enabled {
			continue
		}
		// 当前告警必须匹配 target_match
		if !matchLabels(event.Labels, rule.TargetMatch) {
			continue
		}
		// 查找是否存在匹配 source_match 的 firing 源告警
		for _, src := range firingEvents {
			if src == nil {
				continue
			}
			// 源告警不能是自身
			if src.Fingerprint == event.Fingerprint {
				continue
			}
			if !matchLabels(src.Labels, rule.SourceMatch) {
				continue
			}
			// 检查 equal 标签是否相同
			if equalLabelsMatch(event.Labels, src.Labels, rule.Equal) {
				log.WithFields(log.Fields{
					"inhibit_rule": rule.ID,
					"source_event": src.Fingerprint,
					"target_event": event.Fingerprint,
				}).Info("alert event inhibited")
				return true
			}
		}
	}
	return false
}

// matchLabels 检查事件标签是否满足匹配条件（所有条件 key=value 必须在 labels 中存在且相等）。
func matchLabels(labels map[string]string, matchers map[string]string) bool {
	for k, v := range matchers {
		lv, ok := labels[k]
		if !ok || strings.TrimSpace(lv) != strings.TrimSpace(v) {
			return false
		}
	}
	return true
}

// equalLabelsMatch 检查两组标签在指定 key 列表上的值是否完全相同。
func equalLabelsMatch(a, b map[string]string, keys []string) bool {
	for _, key := range keys {
		av := strings.TrimSpace(a[key])
		bv := strings.TrimSpace(b[key])
		if av != bv {
			return false
		}
	}
	return true
}
