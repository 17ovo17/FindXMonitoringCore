package scheduler

import (
	"sync"
	"time"

	"ai-workbench-api/internal/model"
	"ai-workbench-api/internal/monitoring"
	"ai-workbench-api/internal/store"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"
)

// AlertWorkerManager 管理所有 AlertRuleWorker 的生命周期。
// 参考 n9e 的 Scheduler 设计：定期同步规则列表，对新增/修改/删除的规则
// 分别启动/重启/停止对应的 worker。
type AlertWorkerManager struct {
	workers map[string]*AlertRuleWorker // key: rule.ID
	mu      sync.Mutex
	quit    chan struct{}

	gateway     *monitoring.PrometheusGateway
	datasources []monitoring.Datasource
	timeout     time.Duration
	syncInterval time.Duration
}

// newAlertWorkerManager 创建 worker 管理器。
func newAlertWorkerManager() *AlertWorkerManager {
	timeout := timeoutSecondsFromConfig()
	return &AlertWorkerManager{
		workers:      make(map[string]*AlertRuleWorker),
		quit:         make(chan struct{}),
		gateway:      monitoring.NewPrometheusGateway(nil),
		datasources:  monitoring.PrometheusDatasourcesFromConfig(),
		timeout:      timeout,
		syncInterval: ruleSyncIntervalFromConfig(),
	}
}

// Start 启动管理器：立即同步一次规则，然后定期同步。
func (m *AlertWorkerManager) Start() {
	m.syncRules()
	go m.syncLoop()
	log.Infof("worker-manager: started, sync_interval=%s", m.syncInterval)
}

// Stop 停止管理器及所有 worker。
func (m *AlertWorkerManager) Stop() {
	close(m.quit)
	m.mu.Lock()
	defer m.mu.Unlock()
	for id, w := range m.workers {
		w.Stop()
		delete(m.workers, id)
	}
	log.Info("worker-manager: stopped all workers")
}

func (m *AlertWorkerManager) syncLoop() {
	ticker := time.NewTicker(m.syncInterval)
	defer ticker.Stop()
	for {
		select {
		case <-m.quit:
			return
		case <-ticker.C:
			m.syncRules()
		}
	}
}

// syncRules 从 store 加载最新规则列表，与当前 workers 对比，执行增删改。
func (m *AlertWorkerManager) syncRules() {
	m.mu.Lock()
	defer m.mu.Unlock()

	// 刷新 datasources 配置
	m.datasources = monitoring.PrometheusDatasourcesFromConfig()

	rules := store.ListMonitorAlertRules()
	activeRuleIDs := make(map[string]bool, len(rules))

	for _, rule := range rules {
		if !rule.Enabled || rule.Status != model.MonitorAlertRuleStatusActive {
			// 规则被禁用，如果有 worker 则停止
			if w, exists := m.workers[rule.ID]; exists {
				w.Stop()
				delete(m.workers, rule.ID)
				log.Infof("worker-manager: stopped worker for disabled rule %s", rule.ID)
			}
			continue
		}

		activeRuleIDs[rule.ID] = true

		if w, exists := m.workers[rule.ID]; exists {
			// 规则已有 worker，检查是否需要更新（版本变化）
			if w.Rule.Version != rule.Version || w.Rule.Query != rule.Query {
				w.Stop()
				newWorker := newAlertRuleWorker(rule, m.gateway, m.datasources, m.timeout)
				newWorker.Start()
				m.workers[rule.ID] = newWorker
				log.Infof("worker-manager: restarted worker for updated rule %s (v%d->v%d)", rule.ID, w.Rule.Version, rule.Version)
			}
		} else {
			// 新规则，启动 worker
			newWorker := newAlertRuleWorker(rule, m.gateway, m.datasources, m.timeout)
			newWorker.Start()
			m.workers[rule.ID] = newWorker
			log.Infof("worker-manager: started worker for rule %s", rule.ID)
		}
	}

	// 删除已不存在的规则对应的 worker
	for id, w := range m.workers {
		if !activeRuleIDs[id] {
			w.Stop()
			delete(m.workers, id)
			log.Infof("worker-manager: stopped worker for removed rule %s", id)
		}
	}
}

// ruleSyncIntervalFromConfig 从配置读取规则同步间隔，默认 30 秒。
func ruleSyncIntervalFromConfig() time.Duration {
	seconds := viper.GetInt("monitoring.alert_scheduler.rule_sync_interval_seconds")
	if seconds <= 0 {
		seconds = 30
	}
	if seconds < 5 {
		seconds = 5
	}
	return time.Duration(seconds) * time.Second
}
