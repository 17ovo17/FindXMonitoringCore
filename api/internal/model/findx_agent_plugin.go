package model

import "time"

// PluginCategory 插件分类
const (
	PluginCategoryCollect  = "collect"  // 采集插件（Categraf）
	PluginCategoryDiagnose = "diagnose" // 诊断插件（Catpaw）
	PluginCategoryAPM      = "apm"      // APM 探针（SkyWalking）
)

// FindXAgentPlugin 表示一个可用的 Agent 插件
type FindXAgentPlugin struct {
	ID            string   `json:"id"`
	Name          string   `json:"name"`
	Category      string   `json:"category"` // collect/diagnose/apm
	Description   string   `json:"description"`
	DefaultConfig string   `json:"default_config"`
	ConfigFormat  string   `json:"config_format"` // toml/yaml
	SupportedOS   []string `json:"supported_os"`
	Enabled       bool     `json:"enabled"`
}

// FindXAgentPluginConfig 表示某个 agent 上某个插件的配置
type FindXAgentPluginConfig struct {
	ID        string    `gorm:"primaryKey;size:64" json:"id"`
	AgentID   string    `gorm:"size:64;index" json:"agent_id"`
	PluginID  string    `gorm:"size:64;index" json:"plugin_id"`
	Enabled   bool      `json:"enabled"`
	Config    string    `gorm:"type:text" json:"config"`
	Status    string    `gorm:"size:32" json:"status"` // running/stopped/error
	UpdatedAt time.Time `json:"updated_at"`
	CreatedAt time.Time `json:"created_at"`
}

// FindXAgentEnvironment 表示 agent 所在主机的环境信息
type FindXAgentEnvironment struct {
	AgentID           string   `json:"agent_id"`
	OS                string   `json:"os"`
	Arch              string   `json:"arch"`
	Hostname          string   `json:"hostname"`
	KernelVersion     string   `json:"kernel_version"`
	InstalledServices []string `json:"installed_services"`
	CPUCores          int      `json:"cpu_cores"`
	MemoryMB          int64    `json:"memory_mb"`
	DiskGB            int64    `json:"disk_gb"`
}

// ConfigPushRequest 批量配置下发请求
type ConfigPushRequest struct {
	AgentIDs []string              `json:"agent_ids"`
	Plugins  []ConfigPushPlugin    `json:"plugins"`
	Strategy string                `json:"strategy"` // all/incremental
}

// ConfigPushPlugin 下发的单个插件配置
type ConfigPushPlugin struct {
	ID      string `json:"id"`
	Enabled bool   `json:"enabled"`
	Config  string `json:"config"`
}

// ConfigPushResult 配置下发结果
type ConfigPushResult struct {
	AgentID  string `json:"agent_id"`
	PluginID string `json:"plugin_id"`
	Status   string `json:"status"` // success/failed
	Message  string `json:"message,omitempty"`
}

// PluginRecommendation 自动适配推荐
type PluginRecommendation struct {
	PluginID    string `json:"plugin_id"`
	PluginName  string `json:"plugin_name"`
	Reason      string `json:"reason"`
	Confidence  int    `json:"confidence"` // 0-100
	SuggestedOn bool   `json:"suggested_on"`
}
