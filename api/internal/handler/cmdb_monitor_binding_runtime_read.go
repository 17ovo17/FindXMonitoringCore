package handler

import (
	"net/http"
	"strings"

	"ai-workbench-api/internal/model"
	"ai-workbench-api/internal/store"

	"github.com/gin-gonic/gin"
)

func cmdbMonitorBindingRuntimeReadBlockedEnvelope(instanceID string, bindings []model.CmdbMonitorBinding) (gin.H, bool) {
	for _, binding := range bindings {
		if _, ok := cmdbMonitorBindingRuntimeTemplate(binding.TemplateID); !ok {
			return cmdbMonitorBindingRuntimeReadBlockedContract(instanceID, binding, "monitor template runtime contract is missing or not executable", []string{
				"monitor_template_runtime_contract",
				"cmdb_monitor_template_runtime_content_contract",
			}), true
		}
		if _, ok := store.GetMonitorTarget(binding.HostID); !ok {
			return cmdbMonitorBindingRuntimeReadBlockedContract(instanceID, binding, "monitor host target contract is missing", []string{
				"monitor_host_binding_contract",
			}), true
		}
	}
	return nil, false
}

func cmdbMonitorBindingRuntimeReadBlockedContract(instanceID string, binding model.CmdbMonitorBinding, reason string, missing []string) gin.H {
	contracts := append([]string{"cmdb_monitor_binding_runtime_read_contract"}, missing...)
	contracts = append(contracts, "cmdb_monitor_binding_delivery_receipt_contract", "binding_rollback_contract")
	contract := cmdbMonitorBindingsReadBlockedContract()
	return gin.H{
		"code":              http.StatusConflict,
		"status":            "pending",
		"error":             "pending",
		"message":           "CMDB monitor binding read requires current runtime template and monitor target validation; stale bindings are not returned as ready.",
		"reason":            reason,
		"contract":          "cmdb.monitor_binding.runtime.read.v1",
		"missing_contracts": contracts,
		"instance_id":       strings.TrimSpace(instanceID),
		"binding_id":        strings.TrimSpace(binding.ID),
		"hostid":            strings.TrimSpace(binding.HostID),
		"templateid":        strings.TrimSpace(binding.TemplateID),
		"expected_schema":   contract["expected_schema"],
		"field_matrix": []gin.H{
			cmdbContractFieldGroup("monitor_host_binding", cmdbMonitorBindingRuntimeReadStatus(missing, "monitor_host_binding_contract"), []string{
				"hostid", "monitor_target.id", "monitor_target.ident", "monitor_target.status",
			}, "monitor_host_binding_contract"),
			cmdbContractFieldGroup("monitor_template", cmdbMonitorBindingRuntimeReadStatus(missing, "monitor_template_runtime_contract"), []string{
				"templateid", "type", "content", "runtime", "executor_ref", "config_snippet_ref",
			}, "monitor_template_runtime_contract"),
			cmdbContractFieldGroup("cmdb_monitor_binding_runtime_read", "blocked", []string{
				"binding_id", "instance_id", "hostid", "templateid",
			}, "cmdb_monitor_binding_runtime_read_contract"),
		},
		"source_evidence": contract["source_evidence"],
		"safe_to_retry":   false,
		"meta":            cmdbCompatibleMeta{Persistence: cmdbPersistenceStatus()},
	}
}

func cmdbMonitorBindingRuntimeReadStatus(missing []string, contract string) string {
	for _, item := range missing {
		if item == contract {
			return "blocked"
		}
	}
	return "ready"
}
