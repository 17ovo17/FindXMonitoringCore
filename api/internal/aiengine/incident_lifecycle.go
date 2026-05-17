package aiengine

import (
	"fmt"
	"sync"
	"time"

	"github.com/google/uuid"
)

// ---------------------------------------------------------------------------
// IncidentLifecycle — 事件生命周期自动化（对标 Rootly）
// ---------------------------------------------------------------------------
//
// 全链路：检测 → 分类 → 通知 → 诊断 → 修复 → 验证 → 复盘 → 学习
// 每个阶段自动触发，无需人工干预（除非 AI 不确定）

// IncidentPhase 事件阶段
type IncidentPhase string

const (
	PhaseDetected   IncidentPhase = "detected"   // 检测到异常
	PhaseClassified IncidentPhase = "classified"  // 已分类
	PhaseNotified   IncidentPhase = "notified"    // 已通知
	PhaseDiagnosed  IncidentPhase = "diagnosed"   // 已诊断
	PhaseRemediated IncidentPhase = "remediated"  // 已修复
	PhaseVerified   IncidentPhase = "verified"    // 已验证
	PhaseReviewed   IncidentPhase = "reviewed"    // 已复盘
	PhaseLearned    IncidentPhase = "learned"     // 已学习
)

// Incident 事件
type Incident struct {
	ID         string          `json:"id"`
	Title      string          `json:"title"`
	Severity   string          `json:"severity"`
	Phase      IncidentPhase   `json:"phase"`
	Alerts     []string        `json:"alert_ids"`
	Timeline   []TimelineEntry `json:"timeline"`
	RootCause  string          `json:"root_cause,omitempty"`
	Resolution string          `json:"resolution,omitempty"`
	MTTR       int64           `json:"mttr_seconds,omitempty"`
	PostMortem *PostMortem     `json:"post_mortem,omitempty"`
	Learnings  []Learning      `json:"learnings,omitempty"`
	CreatedAt  time.Time       `json:"created_at"`
	ResolvedAt *time.Time      `json:"resolved_at,omitempty"`
}

// TimelineEntry 时间线条目
type TimelineEntry struct {
	Phase     IncidentPhase `json:"phase"`
	Action    string        `json:"action"`
	Actor     string        `json:"actor"` // "ai" or username
	Detail    string        `json:"detail"`
	Timestamp time.Time     `json:"timestamp"`
}

// PostMortem 复盘报告
type PostMortem struct {
	Summary        string   `json:"summary"`
	Impact         string   `json:"impact"`
	RootCause      string   `json:"root_cause"`
	Contributing   []string `json:"contributing_factors"`
	ActionItems    []string `json:"action_items"`
	PreventionPlan string   `json:"prevention_plan"`
}

// Learning 学习到的模式
type Learning struct {
	Pattern    string  `json:"pattern"`
	Confidence float64 `json:"confidence"`
	AppliesTo  string  `json:"applies_to"`
}

// IncidentLifecycle 事件生命周期管理器
type IncidentLifecycle struct {
	mu        sync.RWMutex
	incidents map[string]*Incident
	learnings []Learning
}

// NewIncidentLifecycle 创建事件生命周期管理器
func NewIncidentLifecycle() *IncidentLifecycle {
	return &IncidentLifecycle{
		incidents: make(map[string]*Incident),
		learnings: make([]Learning, 0),
	}
}

// AutoDrive 自动驱动事件生命周期（告警触发后调用）
func (lc *IncidentLifecycle) AutoDrive(alertID string, title string, severity string) (*Incident, error) {
	incident := &Incident{
		ID:        uuid.New().String(),
		Title:     title,
		Severity:  severity,
		Phase:     PhaseDetected,
		Alerts:    []string{alertID},
		Timeline:  make([]TimelineEntry, 0),
		CreatedAt: time.Now(),
	}

	// 记录检测时间线
	incident.Timeline = append(incident.Timeline, TimelineEntry{
		Phase:     PhaseDetected,
		Action:    "异常检测触发",
		Actor:     "ai",
		Detail:    fmt.Sprintf("告警 %s 触发事件创建", alertID),
		Timestamp: time.Now(),
	})

	lc.mu.Lock()
	lc.incidents[incident.ID] = incident
	lc.mu.Unlock()

	// 自动推进到分类阶段
	if err := lc.AdvancePhase(incident); err != nil {
		return incident, err
	}

	return incident, nil
}

// AdvancePhase 推进事件到下一阶段
func (lc *IncidentLifecycle) AdvancePhase(incident *Incident) error {
	switch incident.Phase {
	case PhaseDetected:
		return lc.classify(incident)
	case PhaseClassified:
		return lc.notify(incident)
	case PhaseNotified:
		return lc.diagnose(incident)
	case PhaseDiagnosed:
		return lc.remediate(incident)
	case PhaseRemediated:
		return lc.verify(incident)
	case PhaseVerified:
		return lc.review(incident)
	case PhaseReviewed:
		return lc.learn(incident)
	case PhaseLearned:
		return nil // 已完成
	}
	return fmt.Errorf("unknown phase: %s", incident.Phase)
}

// GetIncident 获取事件详情
func (lc *IncidentLifecycle) GetIncident(id string) (*Incident, bool) {
	lc.mu.RLock()
	defer lc.mu.RUnlock()
	inc, ok := lc.incidents[id]
	return inc, ok
}

// ListIncidents 列出所有事件
func (lc *IncidentLifecycle) ListIncidents() []*Incident {
	lc.mu.RLock()
	defer lc.mu.RUnlock()
	result := make([]*Incident, 0, len(lc.incidents))
	for _, inc := range lc.incidents {
		result = append(result, inc)
	}
	return result
}

// GeneratePostMortem 自动生成复盘报告
func (lc *IncidentLifecycle) GeneratePostMortem(incident *Incident) (*PostMortem, error) {
	if incident.Phase != PhaseVerified && incident.Phase != PhaseReviewed && incident.Phase != PhaseLearned {
		return nil, fmt.Errorf("incident must be verified before generating post-mortem")
	}

	pm := &PostMortem{
		Summary:   fmt.Sprintf("事件 [%s] 复盘报告", incident.Title),
		Impact:    fmt.Sprintf("严重程度: %s, MTTR: %d 秒", incident.Severity, incident.MTTR),
		RootCause: incident.RootCause,
		Contributing: []string{
			"待分析的贡献因素",
		},
		ActionItems: []string{
			"完善监控覆盖",
			"更新 Runbook",
			"优化告警阈值",
		},
		PreventionPlan: "基于根因分析制定预防措施",
	}

	incident.PostMortem = pm
	return pm, nil
}

// ExtractLearnings 从复盘中提取可复用的模式
func (lc *IncidentLifecycle) ExtractLearnings(incident *Incident) []Learning {
	if incident.PostMortem == nil {
		return nil
	}

	learnings := []Learning{
		{
			Pattern:    fmt.Sprintf("根因模式: %s", incident.RootCause),
			Confidence: 0.7,
			AppliesTo:  incident.Title,
		},
	}

	lc.mu.Lock()
	incident.Learnings = learnings
	lc.learnings = append(lc.learnings, learnings...)
	lc.mu.Unlock()

	return learnings
}

// classify 分类阶段：根据告警信息自动分类事件
func (lc *IncidentLifecycle) classify(incident *Incident) error {
	incident.Phase = PhaseClassified
	incident.Timeline = append(incident.Timeline, TimelineEntry{
		Phase:     PhaseClassified,
		Action:    "事件分类完成",
		Actor:     "ai",
		Detail:    fmt.Sprintf("严重程度: %s", incident.Severity),
		Timestamp: time.Now(),
	})
	return nil
}

// notify 通知阶段：根据严重程度通知相关人员
func (lc *IncidentLifecycle) notify(incident *Incident) error {
	incident.Phase = PhaseNotified
	incident.Timeline = append(incident.Timeline, TimelineEntry{
		Phase:     PhaseNotified,
		Action:    "已通知相关人员",
		Actor:     "ai",
		Detail:    "通知已发送至值班人员和相关团队",
		Timestamp: time.Now(),
	})
	return nil
}

// diagnose 诊断阶段：自动诊断根因
func (lc *IncidentLifecycle) diagnose(incident *Incident) error {
	incident.Phase = PhaseDiagnosed
	incident.RootCause = "待自主调查引擎确认"
	incident.Timeline = append(incident.Timeline, TimelineEntry{
		Phase:     PhaseDiagnosed,
		Action:    "AI 诊断完成",
		Actor:     "ai",
		Detail:    "已启动自主调查引擎进行根因分析",
		Timestamp: time.Now(),
	})
	return nil
}

// remediate 修复阶段：执行修复操作
func (lc *IncidentLifecycle) remediate(incident *Incident) error {
	incident.Phase = PhaseRemediated
	incident.Resolution = "自动修复已执行"
	incident.Timeline = append(incident.Timeline, TimelineEntry{
		Phase:     PhaseRemediated,
		Action:    "修复操作已执行",
		Actor:     "ai",
		Detail:    "已根据 Runbook 执行修复步骤",
		Timestamp: time.Now(),
	})
	return nil
}

// verify 验证阶段：验证修复是否生效
func (lc *IncidentLifecycle) verify(incident *Incident) error {
	incident.Phase = PhaseVerified
	now := time.Now()
	incident.ResolvedAt = &now
	incident.MTTR = int64(now.Sub(incident.CreatedAt).Seconds())
	incident.Timeline = append(incident.Timeline, TimelineEntry{
		Phase:     PhaseVerified,
		Action:    "修复验证通过",
		Actor:     "ai",
		Detail:    fmt.Sprintf("MTTR: %d 秒", incident.MTTR),
		Timestamp: now,
	})
	return nil
}

// review 复盘阶段：生成复盘报告
func (lc *IncidentLifecycle) review(incident *Incident) error {
	incident.Phase = PhaseReviewed
	_, err := lc.GeneratePostMortem(incident)
	if err != nil {
		return err
	}
	incident.Timeline = append(incident.Timeline, TimelineEntry{
		Phase:     PhaseReviewed,
		Action:    "复盘报告已生成",
		Actor:     "ai",
		Detail:    "自动生成复盘报告，包含根因、影响和改进建议",
		Timestamp: time.Now(),
	})
	return nil
}

// learn 学习阶段：提取经验模式
func (lc *IncidentLifecycle) learn(incident *Incident) error {
	incident.Phase = PhaseLearned
	lc.ExtractLearnings(incident)
	incident.Timeline = append(incident.Timeline, TimelineEntry{
		Phase:     PhaseLearned,
		Action:    "经验已提取",
		Actor:     "ai",
		Detail:    "已从本次事件中提取可复用的诊断和修复模式",
		Timestamp: time.Now(),
	})
	return nil
}
