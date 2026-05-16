package model

// IntegrationTemplate 定义 categraf 集成模板
type IntegrationTemplate struct {
	ID           string          `json:"id"`
	Name         string          `json:"name"`
	Description  string          `json:"description"`
	Category     string          `json:"category"`
	Params       []TemplateParam `json:"params"`
	TomlTemplate string          `json:"toml_template"`
}

// TemplateParam 定义模板参数
type TemplateParam struct {
	Key         string   `json:"key"`
	Label       string   `json:"label"`
	Type        string   `json:"type"` // string, password, bool, number, array
	Required    bool     `json:"required"`
	Default     string   `json:"default"`
	Placeholder string   `json:"placeholder"`
	Options     []string `json:"options,omitempty"`
}

// CategrafDeployRequest 部署请求
type CategrafDeployRequest struct {
	TargetIPs    []string          `json:"target_ips" binding:"required"`
	CredentialID string            `json:"credential_id" binding:"required"`
	TemplateID   string            `json:"template_id" binding:"required"`
	Params       map[string]string `json:"params"`
	Port         int               `json:"port"`
}

// CategrafDeployResult 单主机部署结果
type CategrafDeployResult struct {
	IP      string `json:"ip"`
	Status  string `json:"status"` // success, failed
	Message string `json:"message"`
}

// CategrafRenderRequest 渲染请求
type CategrafRenderRequest struct {
	TemplateID string            `json:"template_id" binding:"required"`
	Params     map[string]string `json:"params" binding:"required"`
}

// CategrafVerifyArrivalRequest 指标到达验证请求
type CategrafVerifyArrivalRequest struct {
	TargetIP     string `json:"target_ip" binding:"required"`
	MetricPrefix string `json:"metric_prefix" binding:"required"`
	TimeoutSec   int    `json:"timeout_sec"`
}

// CategrafVerifyArrivalResult 指标到达验证结果
type CategrafVerifyArrivalResult struct {
	Arrived      bool     `json:"arrived"`
	MetricCount  int      `json:"metric_count"`
	SampleMetric string   `json:"sample_metric,omitempty"`
	Metrics      []string `json:"metrics,omitempty"`
	Message      string   `json:"message"`
}
