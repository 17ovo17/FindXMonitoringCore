package handler

import (
	"encoding/json"
	"net/http"
	"strings"

	"ai-workbench-api/internal/model"

	"github.com/gin-gonic/gin"
)

func writeFindXPluginRuntimeActionBlocked(c *gin.Context, action string) {
	c.JSON(http.StatusConflict, gin.H{
		"code":              "pending",
		"status":            "pending",
		"contract_id":       "cmdb.agent.plugin.runtime.action.v1",
		"action":            action,
		"missing_contracts": []string{"cmdb_agent_plugin_runtime_executor_contract", "cmdb_agent_plugin_runtime_receipt_contract", "cmdb_operation_risk_policy_contract", "cmdb_action_audit_receipt_contract"},
		"safe_to_retry":     false,
		"findx_audit_query": "scope=cmdb/resource_type=cmdb_agent_plugin/action=" + action,
	})
}

func writeFindXPluginConfigCredentialBlocked(c *gin.Context, agentID, pluginID string) {
	c.JSON(http.StatusConflict, gin.H{
		"code":        "pending",
		"status":      "pending",
		"contract_id": "cmdb.agent.plugin.credential.v1",
		"message":     "PENDING: plugin config must use credential_ref and schema contracts instead of raw sensitive values",
		"agent_id":    agentID,
		"plugin_id":   pluginID,
		"missing_contracts": []string{
			"cmdb_agent_plugin_credential_contract",
			"cmdb_credential_ref_resolve_contract",
			"cmdb_plugin_config_schema_contract",
		},
		"safe_to_retry":     false,
		"findx_audit_query": "scope=cmdb/resource_type=cmdb_agent_plugin/action=config_update/agent_id=" + agentID,
	})
}

func sanitizeFindXAgentPluginConfigs(configs []model.FindXAgentPluginConfig) []model.FindXAgentPluginConfig {
	out := make([]model.FindXAgentPluginConfig, 0, len(configs))
	for _, cfg := range configs {
		out = append(out, sanitizeFindXAgentPluginConfig(cfg))
	}
	return out
}

func sanitizeFindXAgentPluginConfig(cfg model.FindXAgentPluginConfig) model.FindXAgentPluginConfig {
	if findXAgentPluginConfigHasSensitivePayload(cfg.Config) {
		cfg.Config = "REDACTED_CONFIG"
	}
	return cfg
}

func findXAgentPluginConfigHasSensitivePayload(config string) bool {
	lower := strings.ToLower(config)
	if strings.Contains(lower, "://") && strings.Contains(lower, "@") {
		return true
	}
	for _, marker := range []string{"password", "passwd", "token", "cookie", "secret", "dsn", "bearer", "api_key", "apikey", "private_key", "authorization"} {
		if strings.Contains(lower, marker) {
			return true
		}
	}
	if strings.TrimSpace(config) == "" {
		return false
	}
	var payload any
	if err := json.Unmarshal([]byte(config), &payload); err == nil {
		return findXAgentPluginConfigJSONHasSensitivePayload(payload)
	}
	return false
}

func findXAgentPluginConfigJSONHasSensitivePayload(value any) bool {
	switch typed := value.(type) {
	case map[string]any:
		for key, child := range typed {
			if findXAgentPluginConfigHasSensitivePayload(key) || findXAgentPluginConfigJSONHasSensitivePayload(child) {
				return true
			}
		}
	case []any:
		for _, child := range typed {
			if findXAgentPluginConfigJSONHasSensitivePayload(child) {
				return true
			}
		}
	case string:
		return findXAgentPluginConfigHasSensitivePayload(typed)
	}
	return false
}
