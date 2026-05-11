package store

import (
	"sync"
	"time"

	"ai-workbench-api/internal/model"

	"github.com/sirupsen/logrus"
)

var (
	agentPackagesMu    sync.RWMutex
	agentPackagesMemory = map[string]*model.AgentPackage{}

	agentLifecycleEventsMu    sync.RWMutex
	agentLifecycleEventsMemory = map[string]*model.AgentLifecycleEvent{}
)

func init() {
	// Will be called after InitGormDB if GORM is available
}

func autoMigrateAgentPackages() {
	if !GormOK() {
		return
	}
	err := GetDB().AutoMigrate(&model.AgentPackage{}, &model.AgentLifecycleEvent{})
	if err != nil {
		logrus.WithError(err).Warn("agent_packages: auto migrate failed, using memory fallback")
	}
}

// --- AgentPackage CRUD ---

func ListAgentPackages() ([]model.AgentPackage, error) {
	if GormOK() {
		var rows []model.AgentPackage
		if err := GetDB().Order("created_at DESC").Find(&rows).Error; err == nil {
			return rows, nil
		}
	}
	agentPackagesMu.RLock()
	defer agentPackagesMu.RUnlock()
	out := make([]model.AgentPackage, 0, len(agentPackagesMemory))
	for _, item := range agentPackagesMemory {
		out = append(out, *item)
	}
	return out, nil
}

func GetAgentPackage(id string) (model.AgentPackage, bool, error) {
	if GormOK() {
		var row model.AgentPackage
		if err := GetDB().Where("id = ?", id).First(&row).Error; err == nil {
			return row, true, nil
		}
	}
	agentPackagesMu.RLock()
	defer agentPackagesMu.RUnlock()
	item, ok := agentPackagesMemory[id]
	if !ok {
		return model.AgentPackage{}, false, nil
	}
	return *item, true, nil
}

func SaveAgentPackage(pkg model.AgentPackage) (model.AgentPackage, error) {
	if pkg.CreatedAt.IsZero() {
		pkg.CreatedAt = time.Now()
	}
	if GormOK() {
		if err := GetDB().Save(&pkg).Error; err != nil {
			logrus.WithError(err).Warn("agent_packages: gorm save failed, falling back to memory")
		} else {
			return pkg, nil
		}
	}
	agentPackagesMu.Lock()
	cp := pkg
	agentPackagesMemory[pkg.ID] = &cp
	agentPackagesMu.Unlock()
	return pkg, nil
}

func DeleteAgentPackage(id string) error {
	if GormOK() {
		if err := GetDB().Where("id = ?", id).Delete(&model.AgentPackage{}).Error; err != nil {
			logrus.WithError(err).Warn("agent_packages: gorm delete failed")
		}
	}
	agentPackagesMu.Lock()
	delete(agentPackagesMemory, id)
	agentPackagesMu.Unlock()
	return nil
}

// --- AgentLifecycleEvent ---

func SaveAgentLifecycleEvent(event model.AgentLifecycleEvent) (model.AgentLifecycleEvent, error) {
	if event.CreatedAt.IsZero() {
		event.CreatedAt = time.Now()
	}
	if GormOK() {
		if err := GetDB().Save(&event).Error; err == nil {
			return event, nil
		}
	}
	agentLifecycleEventsMu.Lock()
	cp := event
	agentLifecycleEventsMemory[event.ID] = &cp
	agentLifecycleEventsMu.Unlock()
	return event, nil
}

func ListAgentLifecycleEvents(agentID string) ([]model.AgentLifecycleEvent, error) {
	if GormOK() {
		var rows []model.AgentLifecycleEvent
		if err := GetDB().Where("agent_id = ?", agentID).Order("created_at DESC").Limit(100).Find(&rows).Error; err == nil {
			return rows, nil
		}
	}
	agentLifecycleEventsMu.RLock()
	defer agentLifecycleEventsMu.RUnlock()
	out := make([]model.AgentLifecycleEvent, 0)
	for _, item := range agentLifecycleEventsMemory {
		if item.AgentID == agentID {
			out = append(out, *item)
		}
	}
	return out, nil
}

// InitAgentPackageStore should be called after InitGormDB.
func InitAgentPackageStore() {
	autoMigrateAgentPackages()
}
