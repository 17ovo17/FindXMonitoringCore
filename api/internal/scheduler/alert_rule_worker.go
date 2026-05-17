package scheduler

import (
	"context"
	"sync"
	"time"

	"ai-workbench-api/internal/evaluator"
	"ai-workbench-api/internal/model"
	"ai-workbench-api/internal/monitoring"
	"ai-workbench-api/internal/store"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"
)

// pendingEntry 记录处于 pending 状态的告警事件，等待 ForDuration 满足后升级为 firing。
type pendingEntry struct {
	Event     *model.MonitorAlertEvent
	FirstSeen time.Time
}

// AlertRuleWorker 每条告警规则对应一个独立的 worker，拥有自己的定时器和状态。
// 参考 n9e AlertRuleWorker + Processor 架构：每个 worker 维护 fires/pendings 两个 map，
// 独立执行 eval 循环，独立管理事件生命周期。
type AlertRuleWorker struct {
	Rule     model.MonitorAlertRule
	Interval time.Duration
	quit     chan struct{}

	// fires 存储当前正在触发的事件，key 为 fingerprint
	fires map[string]*model.MonitorAlertEvent
	// pendings 存储等待 ForDuration 的事件，key 为 fingerprint
	pendings map[string]*pendingEntry
	// lastNotifyTime 记录每个 fingerprint 上次发送通知的时间，用于重复通知间隔控制
	lastNotifyTime map[string]time.Time
	mu             sync.Mutex

	// 依赖
	gateway     *monitoring.PrometheusGateway
	datasources []monitoring.Datasource
	timeout     time.Duration
}

// newAlertRuleWorker 创建一个新的规则 worker。
func newAlertRuleWorker(rule model.MonitorAlertRule, gateway *monitoring.PrometheusGateway, datasources []monitoring.Datasource, timeout time.Duration) *AlertRuleWorker {
	interval := schedulerIntervalFromConfig()
	return &AlertRuleWorker{
		Rule:           rule,
		Interval:       interval,
		quit:           make(chan struct{}),
		fires:          make(map[string]*model.MonitorAlertEvent),
		pendings:       make(map[string]*pendingEntry),
		lastNotifyTime: make(map[string]time.Time),
		gateway:        gateway,
		datasources:    datasources,
		timeout:        timeout,
	}
}

// Start 启动 worker 的独立 eval 循环。
func (w *AlertRuleWorker) Start() {
	go w.loop()
}

// Stop 停止 worker。
func (w *AlertRuleWorker) Stop() {
	close(w.quit)
}

func (w *AlertRuleWorker) loop() {
	ticker := time.NewTicker(w.Interval)
	defer ticker.Stop()
	for {
		select {
		case <-w.quit:
			return
		case <-ticker.C:
			w.eval()
		}
	}
}

// eval 执行一次规则评估，更新 fires/pendings 状态，触发通知或恢复。
func (w *AlertRuleWorker) eval() {
	w.mu.Lock()
	defer w.mu.Unlock()

	ctx, cancel := context.WithTimeout(context.Background(), w.timeout)
	defer cancel()

	rule := w.Rule
	now := time.Now()

	// 验证 PromQL
	if err := monitoring.ValidatePromQL(rule.Query); err != nil {
		log.Warnf("worker[%s]: invalid promql: %v", rule.ID, err)
		return
	}

	// 解析 datasource
	dsID, base, err := monitoring.ResolvePrometheusDatasource(
		w.datasources, rule.DatasourceID,
		monitoring.DefaultPrometheusDatasourceID,
		viper.GetString("prometheus.url"),
	)
	if err != nil {
		log.Warnf("worker[%s]: datasource not found: %v", rule.ID, err)
		return
	}
	rule.DatasourceID = dsID

	// 查询 Prometheus
	prom, err := w.gateway.QueryInstant(ctx, monitoring.PrometheusQueryRequest{
		BaseURL: base, Query: rule.Query, Timeout: w.timeout,
	})
	if err != nil {
		log.Warnf("worker[%s]: prometheus query failed: %v", rule.ID, err)
		return
	}

	// 评估规则
	queryHash := monitoring.QueryHash(rule.Query)
	evalRule := evaluator.Rule{
		ID: rule.ID, Name: rule.Name, Severity: rule.Severity, DatasourceID: dsID,
		QueryHash: queryHash, ForDuration: rule.ForDuration, NoDataPolicy: rule.NoDataPolicy,
		Labels: rule.Labels, Annotations: rule.Annotations,
	}
	eval, evalErr := evaluator.EvaluateRule(evalRule, prom)
	if evalErr != nil {
		log.Warnf("worker[%s]: evaluation error: %v", rule.ID, evalErr)
		return
	}

	// 写 eval log
	started := now
	addMonitorAlertEvalLog(model.MonitorAlertEvalLog{
		RuleID: rule.ID, RuleVersion: rule.Version, Status: eval.State,
		Message: "worker eval", Details: evaluator.SafeDetails(eval),
		StartedAt: started, FinishedAt: time.Now(),
		DurationMs: time.Since(started).Milliseconds(), DatasourceID: dsID,
		QueryHash: queryHash,
	})

	// 处理 no_data_keep_state：不改变任何状态
	if eval.State == "no_data_keep_state" {
		return
	}

	// 解析 ForDuration
	forDurationMS, _ := evaluator.ParsePromDurationMillis(rule.ForDuration)
	forDuration := time.Duration(forDurationMS) * time.Millisecond

	// 获取重复通知间隔（默认 1 小时）
	repeatInterval := repeatNotifyIntervalFromConfig()

	// 当前评估产生的活跃 fingerprint 集合
	activeFingerprints := make(map[string]bool)

	if eval.Triggered && len(eval.Candidates) > 0 {
		for _, candidate := range eval.Candidates {
			event := w.buildEvent(rule, candidate, now)
			fingerprint := model.GenerateMonitorAlertEventFingerprint(event)
			activeFingerprints[fingerprint] = true

			if fired, exists := w.fires[fingerprint]; exists {
				// 已在 fires 中：更新 LastSeen 和 Count，检查是否需要重复通知
				fired.LastSeen = now
				fired.Count++
				w.checkRepeatNotify(fingerprint, fired, repeatInterval, now)
			} else if pending, exists := w.pendings[fingerprint]; exists {
				// 已在 pendings 中：检查是否满足 ForDuration
				pending.Event.LastSeen = now
				if forDuration <= 0 || now.Sub(pending.FirstSeen) >= forDuration {
					// 满足 ForDuration，升级到 fires 并发送通知
					w.promoteToFiring(fingerprint, pending.Event, now)
				}
			} else {
				// 新出现的 fingerprint
				if forDuration <= 0 {
					// 无 ForDuration 要求，直接 fire
					w.promoteToFiring(fingerprint, event, now)
				} else {
					// 加入 pendings 等待
					w.pendings[fingerprint] = &pendingEntry{
						Event:     event,
						FirstSeen: now,
					}
				}
			}
		}
	}

	// 恢复检测：fires 中不在当前活跃集合的 → 恢复
	for fp, event := range w.fires {
		if activeFingerprints[fp] {
			continue
		}
		w.recoverEvent(fp, event)
	}

	// pendings 中不在当前活跃集合的 → 移除（未达到 firing 就恢复了）
	for fp := range w.pendings {
		if activeFingerprints[fp] {
			continue
		}
		delete(w.pendings, fp)
	}
}

// promoteToFiring 将事件从 pendings 升级到 fires，持久化并发送通知。
func (w *AlertRuleWorker) promoteToFiring(fingerprint string, event *model.MonitorAlertEvent, now time.Time) {
	delete(w.pendings, fingerprint)
	event.FirstSeen = now
	event.LastSeen = now
	event.Count = 1
	w.fires[fingerprint] = event

	// 抑制检查：收集当前所有 firing 事件作为源告警
	firingEvents := make([]*model.MonitorAlertEvent, 0, len(w.fires))
	for _, fe := range w.fires {
		firingEvents = append(firingEvents, fe)
	}
	if GetInhibitManager().IsInhibited(event, firingEvents) {
		log.Infof("worker[%s]: event %s inhibited, skip notification", w.Rule.ID, fingerprint)
		// 仍然持久化事件，但不发送通知
		store.UpsertMonitorAlertEvent(event)
		return
	}

	// 持久化并发送通知
	stored, err := store.UpsertMonitorAlertEvent(event)
	if err != nil {
		log.Warnf("worker[%s]: upsert event failed: %v", w.Rule.ID, err)
		return
	}
	processed, shouldDrop := ApplyPipelines(stored)
	if !shouldDrop {
		dispatchAlertEvent(processed)
	}
	w.lastNotifyTime[fingerprint] = now
}

// checkRepeatNotify 检查是否需要发送重复通知。
func (w *AlertRuleWorker) checkRepeatNotify(fingerprint string, event *model.MonitorAlertEvent, repeatInterval time.Duration, now time.Time) {
	if repeatInterval <= 0 {
		return
	}
	lastSent, exists := w.lastNotifyTime[fingerprint]
	if !exists || now.Sub(lastSent) >= repeatInterval {
		// 更新持久化
		store.UpsertMonitorAlertEvent(event)
		processed, shouldDrop := ApplyPipelines(event)
		if !shouldDrop {
			dispatchAlertEvent(processed)
		}
		w.lastNotifyTime[fingerprint] = now
	}
}

// recoverEvent 将事件标记为恢复，从 fires 中移除，持久化恢复状态并发送恢复通知。
func (w *AlertRuleWorker) recoverEvent(fingerprint string, event *model.MonitorAlertEvent) {
	delete(w.fires, fingerprint)
	delete(w.lastNotifyTime, fingerprint)

	if event.ID == "" {
		return
	}
	_, ok, err := store.ApplyMonitorAlertEventAction(event.ID, model.MonitorAlertAction{
		Action: model.MonitorAlertEventActionResolve,
		Actor:  monitorAlertSchedulerActor,
		Reason: "worker recovered",
	})
	if err != nil || !ok {
		log.Warnf("worker[%s]: recover event failed id=%s: %v", w.Rule.ID, event.ID, err)
	}
}

// buildEvent 从评估候选项构建告警事件。
func (w *AlertRuleWorker) buildEvent(rule model.MonitorAlertRule, candidate evaluator.Candidate, now time.Time) *model.MonitorAlertEvent {
	labels := sanitizeMonitorAlertStringMap(mergeMonitorStringMaps(rule.Labels, candidate.Labels))
	annotations := sanitizeMonitorAlertStringMap(mergeMonitorStringMaps(rule.Annotations, map[string]string{"reason": candidate.Reason}))
	return &model.MonitorAlertEvent{
		RuleID:       rule.ID,
		RuleVersion:  rule.Version,
		EventKey:     candidate.EventKey,
		Name:         rule.Name,
		Severity:     rule.Severity,
		Status:       model.MonitorAlertEventStatusFiring,
		DatasourceID: rule.DatasourceID,
		TargetIdent:  monitorCandidateTargetIdent(labels),
		Labels:       labels,
		Annotations:  annotations,
		Value:        candidate.Value,
		Count:        1,
		FirstSeen:    now,
		LastSeen:     now,
	}
}

// repeatNotifyIntervalFromConfig 从配置读取重复通知间隔，默认 3600 秒。
func repeatNotifyIntervalFromConfig() time.Duration {
	seconds := viper.GetInt("monitoring.alert_scheduler.repeat_interval_seconds")
	if seconds <= 0 {
		seconds = 3600
	}
	return time.Duration(seconds) * time.Second
}
