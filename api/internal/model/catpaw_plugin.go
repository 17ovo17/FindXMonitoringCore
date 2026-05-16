package model

import "time"

// CatpawPlugin 表示一个 catpaw 巡检插件模板
type CatpawPlugin struct {
	ID            string `json:"id"`
	Name          string `json:"name"`
	Category      string `json:"category"`
	Description   string `json:"description"`
	DefaultConfig string `json:"default_config"`
	Enabled       bool   `json:"enabled"`
}

// CatpawDeployRequest 批量部署 catpaw 插件配置的请求
type CatpawDeployRequest struct {
	TargetIPs    []string          `json:"target_ips"`
	CredentialID string            `json:"credential_id"`
	Plugins      []string          `json:"plugins"`
	CustomConfig map[string]string `json:"custom_config"` // plugin_id -> custom toml
}

// CatpawDeployResult 单台主机的部署结果
type CatpawDeployResult struct {
	IP      string `json:"ip"`
	Status  string `json:"status"` // success, failed
	Message string `json:"message"`
}

// CatpawInspectionResult 巡检结果
type CatpawInspectionResult struct {
	ID         string            `json:"id"`
	IP         string            `json:"ip"`
	PluginID   string            `json:"plugin_id"`
	Severity   string            `json:"severity"` // ok, warning, critical
	Summary    string            `json:"summary"`
	Detail     string            `json:"detail"`
	Labels     map[string]string `json:"labels,omitempty"`
	CollectedAt time.Time        `json:"collected_at"`
	ExpiresAt  time.Time         `json:"expires_at"`
}

// CatpawDiagnoseRequest AI 诊断请求
type CatpawDiagnoseRequest struct {
	IP      string `json:"ip" binding:"required"`
	Prompt  string `json:"prompt"`
	Plugins []string `json:"plugins"` // 指定分析哪些插件的结果
}
