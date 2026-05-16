package model

import "time"

// PluginCategory 插件分类
const (
	PluginCategoryCollect  = "collect"  // 采集插件
	PluginCategoryDiagnose = "diagnose" // 诊断插件
	PluginCategoryAPM      = "apm"      // APM 探针
)

// FindXAgentPlugin 表示一个可用的 Agent 插件
type FindXAgentPlugin struct {
	ID                    string            `json:"id"`
	Name                  string            `json:"name"`
	Category              string            `json:"category"` // collect/diagnose/apm
	Description           string            `json:"description"`
	DefaultConfig         string            `json:"default_config"`
	ConfigFormat          string            `json:"config_format"` // toml/yaml
	SupportedOS           []string          `json:"supported_os"`
	Enabled               bool              `json:"enabled"`
	CredentialRequired    bool              `json:"credential_required"`
	CredentialSchema      map[string]string `json:"credential_schema,omitempty"`
	DashboardRefs         []string          `json:"dashboard_refs,omitempty"`
	SecurityLevel         string            `json:"security_level,omitempty"`
	MissingContracts      []string          `json:"missing_contracts,omitempty"`
	Blockers              []string          `json:"blockers,omitempty"`
	ConfigSchemaContracts []string          `json:"config_schema_contracts,omitempty"`
}

// FindXAgentPluginConfig 表示某个 agent 上某个插件的配置
type FindXAgentPluginConfig struct {
	ID        string    `gorm:"primaryKey;size:64" json:"id"`
	AgentID   string    `gorm:"size:64;index" json:"agent_id"`
	PluginID  string    `gorm:"size:64;index" json:"plugin_id"`
	Enabled   bool      `json:"enabled"`
	Config    string    `gorm:"type:text" json:"config"`
	Status    string    `gorm:"size:32" json:"status"` // local config state only
	UpdatedAt time.Time `json:"updated_at"`
	CreatedAt time.Time `json:"created_at"`
}

// FindXAgentEnvironment 表示 agent 所在主机的环境信息
type FindXAgentPluginAssignment struct {
	ID                   string    `gorm:"primaryKey;size:64" json:"id"`
	SourceRolloutID      string    `gorm:"size:64;index" json:"source_rollout_id"`
	HostRef              string    `gorm:"size:128;index" json:"host_ref"`
	AgentRef             string    `gorm:"size:128;index" json:"agent_ref"`
	PluginID             string    `gorm:"size:255;index" json:"plugin_id"`
	PluginVersion        string    `gorm:"size:128" json:"plugin_version,omitempty"`
	ConfigSnippetRef     string    `gorm:"size:255" json:"config_snippet_ref,omitempty"`
	ConfigFormat         string    `gorm:"size:32" json:"config_format,omitempty"`
	ProviderMode         string    `gorm:"size:64" json:"provider_mode,omitempty"`
	TargetBindingRef     string    `gorm:"size:64;index" json:"target_binding_ref,omitempty"`
	AuditRef             string    `gorm:"size:128;index" json:"audit_ref,omitempty"`
	AssignmentContract   string    `gorm:"size:128" json:"assignment_contract"`
	Status               string    `gorm:"size:32;index" json:"status"`
	Blocker              string    `gorm:"size:255" json:"blocker,omitempty"`
	CredentialRefPresent bool      `json:"credential_ref_present"`
	DashboardRefsJSON    string    `gorm:"column:dashboard_refs;type:text" json:"-"`
	DashboardRefsCount   int       `json:"dashboard_refs_count"`
	MissingContractsJSON string    `gorm:"type:text" json:"-"`
	MetadataJSON         string    `gorm:"type:text" json:"-"`
	CreatedAt            time.Time `json:"created_at"`
	UpdatedAt            time.Time `json:"updated_at"`
}

type FindXAgentPluginTargetBinding struct {
	ID                   string    `gorm:"primaryKey;size:64" json:"id"`
	AssignmentID         string    `gorm:"size:64;index" json:"assignment_id"`
	SourceRolloutID      string    `gorm:"size:64;index" json:"source_rollout_id"`
	HostRef              string    `gorm:"size:128;index" json:"host_ref"`
	TargetID             string    `gorm:"size:128;index" json:"target_id"`
	AgentRef             string    `gorm:"size:128;index" json:"agent_ref"`
	PluginID             string    `gorm:"size:255;index" json:"plugin_id"`
	BindingType          string    `gorm:"size:64" json:"binding_type"`
	Status               string    `gorm:"size:32;index" json:"status"`
	Blocker              string    `gorm:"size:255" json:"blocker,omitempty"`
	CredentialRefPresent bool      `json:"credential_ref_present"`
	DashboardRefsCount   int       `json:"dashboard_refs_count"`
	ContractID           string    `gorm:"size:128" json:"contract_id"`
	AuditRef             string    `gorm:"size:128;index" json:"audit_ref,omitempty"`
	CreatedAt            time.Time `json:"created_at"`
	UpdatedAt            time.Time `json:"updated_at"`
}

func (FindXAgentPluginAssignment) TableName() string {
	return "findx_agent_plugin_assignments"
}

func (FindXAgentPluginTargetBinding) TableName() string {
	return "findx_agent_plugin_target_bindings"
}

type FindXAgentEnvironment struct {
	AgentID          string   `json:"agent_id"`
	OS               string   `json:"os"`
	Arch             string   `json:"arch"`
	Hostname         string   `json:"hostname"`
	KernelVersion    string   `json:"kernel_version"`
	DetectedServices []string `json:"detected_services"`
	CPUCores         int      `json:"cpu_cores"`
	MemoryMB         int64    `json:"memory_mb"`
	DiskGB           int64    `json:"disk_gb"`
}

// ConfigPushRequest 批量配置下发请求
type ConfigPushRequest struct {
	AgentIDs []string           `json:"agent_ids"`
	Plugins  []ConfigPushPlugin `json:"plugins"`
	Strategy string             `json:"strategy"` // all/incremental
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
	Status   string `json:"status"` // blocked until executor and receipts exist
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
