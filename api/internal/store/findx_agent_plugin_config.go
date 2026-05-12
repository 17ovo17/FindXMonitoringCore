package store

import (
	"sync"
	"time"

	"ai-workbench-api/internal/model"

	"github.com/sirupsen/logrus"
)

var (
	pluginConfigsMu sync.RWMutex
	pluginConfigs   = map[string]*model.FindXAgentPluginConfig{} // key: agentID+":"+pluginID
)

// InitPluginConfigStore 初始化插件配置存储
func InitPluginConfigStore() {
	if GormOK() {
		err := GetDB().AutoMigrate(&model.FindXAgentPluginConfig{})
		if err != nil {
			logrus.WithError(err).Warn("plugin_config: auto migrate failed")
		}
	}
}

// ListPluginConfigs 获取某个 agent 的所有插件配置
func ListPluginConfigs(agentID string) ([]model.FindXAgentPluginConfig, error) {
	if GormOK() {
		var rows []model.FindXAgentPluginConfig
		if err := GetDB().Where("agent_id = ?", agentID).Find(&rows).Error; err == nil {
			return rows, nil
		}
	}
	pluginConfigsMu.RLock()
	defer pluginConfigsMu.RUnlock()
	out := make([]model.FindXAgentPluginConfig, 0)
	for _, item := range pluginConfigs {
		if item.AgentID == agentID {
			out = append(out, *item)
		}
	}
	return out, nil
}

// GetPluginConfig 获取单个插件配置
func GetPluginConfig(agentID, pluginID string) (model.FindXAgentPluginConfig, bool) {
	key := agentID + ":" + pluginID
	if GormOK() {
		var row model.FindXAgentPluginConfig
		err := GetDB().Where("agent_id = ? AND plugin_id = ?", agentID, pluginID).First(&row).Error
		if err == nil {
			return row, true
		}
	}
	pluginConfigsMu.RLock()
	defer pluginConfigsMu.RUnlock()
	item, ok := pluginConfigs[key]
	if !ok {
		return model.FindXAgentPluginConfig{}, false
	}
	return *item, true
}

// SavePluginConfig 保存插件配置
func SavePluginConfig(cfg model.FindXAgentPluginConfig) (model.FindXAgentPluginConfig, error) {
	if cfg.CreatedAt.IsZero() {
		cfg.CreatedAt = time.Now()
	}
	cfg.UpdatedAt = time.Now()
	if cfg.ID == "" {
		cfg.ID = cfg.AgentID + ":" + cfg.PluginID
	}
	if GormOK() {
		if err := GetDB().Save(&cfg).Error; err == nil {
			return cfg, nil
		}
	}
	pluginConfigsMu.Lock()
	cp := cfg
	pluginConfigs[cfg.ID] = &cp
	pluginConfigsMu.Unlock()
	return cfg, nil
}

// UpdatePluginEnabled 更新插件启用状态
func UpdatePluginEnabled(agentID, pluginID string, enabled bool) error {
	key := agentID + ":" + pluginID
	if GormOK() {
		err := GetDB().Model(&model.FindXAgentPluginConfig{}).
			Where("agent_id = ? AND plugin_id = ?", agentID, pluginID).
			Update("enabled", enabled).Error
		if err == nil {
			return nil
		}
	}
	pluginConfigsMu.Lock()
	defer pluginConfigsMu.Unlock()
	item, ok := pluginConfigs[key]
	if ok {
		item.Enabled = enabled
		item.UpdatedAt = time.Now()
	} else {
		pluginConfigs[key] = &model.FindXAgentPluginConfig{
			ID:        key,
			AgentID:   agentID,
			PluginID:  pluginID,
			Enabled:   enabled,
			Status:    "stopped",
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}
	}
	return nil
}

// UpdatePluginStatus 更新插件运行状态
func UpdatePluginStatus(agentID, pluginID, status string) error {
	key := agentID + ":" + pluginID
	if GormOK() {
		err := GetDB().Model(&model.FindXAgentPluginConfig{}).
			Where("agent_id = ? AND plugin_id = ?", agentID, pluginID).
			Update("status", status).Error
		if err == nil {
			return nil
		}
	}
	pluginConfigsMu.Lock()
	defer pluginConfigsMu.Unlock()
	item, ok := pluginConfigs[key]
	if ok {
		item.Status = status
		item.UpdatedAt = time.Now()
	}
	return nil
}
