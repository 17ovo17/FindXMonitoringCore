package handler

import (
	"net/http"
	"sort"
	"strings"
	"time"

	"ai-workbench-api/internal/model"
	"ai-workbench-api/internal/store"

	"github.com/gin-gonic/gin"
)

const (
	cmdbResourceProjectionReadContract = "cmdb.resource.projection.read.v1"
	cmdbResourceProjectionHostAsset    = "host_asset"
)

var cmdbResourceProjectionSourceMissingContracts = []string{
	"cmdb_resource_projection_source_contract",
	"cmdb_resource_projection_model_contract",
}

var cmdbResourceProjectionRuntimeMissingContracts = []string{
	"cmdb_resource_projection_runtime_contract",
	"cmdb_resource_approval_runtime_contract",
	"cmdb_dashboard_import_runtime_contract",
	"cmdb_operation_risk_policy_contract",
}

// GetCmdbResourceProjection builds read-only resource projections from live monitor targets and CMDB instances.
func GetCmdbResourceProjection(c *gin.Context) {
	resourceType := strings.TrimSpace(c.Query("resource_type"))
	if resourceType != cmdbResourceProjectionHostAsset {
		c.JSON(http.StatusConflict, cmdbResourceProjectionBlockedEnvelope(resourceType))
		return
	}
	projection, ok := buildCmdbHostAssetProjection(c)
	if !ok {
		c.JSON(http.StatusConflict, cmdbResourceProjectionBlockedEnvelope(resourceType))
		return
	}
	c.JSON(http.StatusOK, projection)
}

func buildCmdbHostAssetProjection(c *gin.Context) (gin.H, bool) {
	targets := store.ListMonitorTargets()
	agents := agentsByTarget()
	columnsByAttr := map[string]model.HostAssetCmdbColumn{}
	rows := make([]gin.H, 0, len(targets))
	for _, target := range targets {
		asset := hostAssetFromTarget(target, agents[target.ID])
		applyCmdbHostFields(&asset)
		if !matchHostAssetQuery(c, asset) {
			continue
		}
		if asset.CmdbInstance == nil || len(asset.CmdbColumns) == 0 || len(asset.CmdbValues) == 0 {
			continue
		}
		values := gin.H{}
		for key, value := range asset.CmdbValues {
			values[key] = value
		}
		row := gin.H{
			"host_id":          asset.HostID,
			"ident":            asset.Ident,
			"hostname":         asset.Hostname,
			"ip_list":          asset.IPList,
			"agent_id":         asset.AgentID,
			"agent_status":     asset.AgentStatus,
			"cmdb_instance_id": asset.CmdbInstance.InstanceID,
			"cmdb_object_id":   asset.CmdbInstance.ObjectID,
			"cmdb_instance":    asset.CmdbInstance,
			"values":           values,
		}
		for _, column := range asset.CmdbColumns {
			columnsByAttr[column.Attr] = column
			if value, ok := asset.CmdbValues[column.Attr]; ok {
				row[column.Attr] = value
			}
		}
		rows = append(rows, row)
	}
	if len(rows) == 0 || len(columnsByAttr) == 0 {
		return nil, false
	}
	columns := make([]model.HostAssetCmdbColumn, 0, len(columnsByAttr))
	for _, column := range columnsByAttr {
		columns = append(columns, column)
	}
	sort.Slice(columns, func(i, j int) bool {
		if columns[i].Sort == columns[j].Sort {
			return columns[i].Attr < columns[j].Attr
		}
		return columns[i].Sort < columns[j].Sort
	})
	return gin.H{
		"contract_id":       cmdbResourceProjectionReadContract,
		"contract":          cmdbResourceProjectionReadContract,
		"code":              0,
		"status":            cmdbResourceStatisticsStatus,
		"resource_type":     cmdbResourceProjectionHostAsset,
		"identity_fields":   []string{"host_id", "cmdb_instance_id", "cmdb_object_id"},
		"columns":           columns,
		"rows":              rows,
		"missing_contracts": cmdbResourceProjectionRuntimeMissingContracts,
		"blocked_contracts": cmdbResourceProjectionBlockedContracts(),
		"generated_at":      time.Now().UTC(),
		"expected_schema":   cmdbResourceProjectionExpectedSchema(),
		"field_matrix":      cmdbResourceProjectionFieldMatrix("ready"),
		"findx_audit_query": cmdbResourceProjectionAuditQuery(),
		"meta": gin.H{
			"source":      "monitor_target_cmdb_instance",
			"persistence": cmdbPersistenceStatus(),
		},
	}, true
}

func cmdbResourceProjectionBlockedEnvelope(resourceType string) gin.H {
	return cmdbBlockedContractEnvelope(
		cmdbResourceProjectionReadContract,
		cmdbResourceProjectionSourceMissingContracts,
		gin.H{
			"contract":        cmdbResourceProjectionReadContract,
			"status":          cmdbBlockedByContract,
			"error":           cmdbBlockedByContract,
			"resource_type":   resourceType,
			"expected_schema": cmdbResourceProjectionExpectedSchema(),
			"field_matrix":    cmdbResourceProjectionFieldMatrix("blocked"),
		},
	)
}

func cmdbResourceProjectionExpectedSchema() gin.H {
	return gin.H{
		"query":   []string{"resource_type", "workspace_id", "resource_group_id", "status", "online", "tag", "tags", "keyword"},
		"columns": []string{"attr", "label", "value_type", "tag", "unit", "sort", "visible", "sensitive", "masked"},
		"rows":    []string{"host_id", "ident", "hostname", "ip_list", "agent_id", "agent_status", "cmdb_instance", "values"},
		"audit":   []string{"source", "scope", "resource_type", "action"},
	}
}

func cmdbResourceProjectionFieldMatrix(sourceStatus string) []gin.H {
	return []gin.H{
		cmdbContractFieldGroup("resource_projection_source", sourceStatus, []string{
			"monitor_target", "findx_agent", "cmdb_instance",
		}, "cmdb_resource_projection_source_contract"),
		cmdbContractFieldGroup("resource_projection_model", sourceStatus, []string{
			"cmdb_columns", "cmdb_values", "attr", "label", "value_type", "masked",
		}, "cmdb_resource_projection_model_contract"),
		cmdbContractFieldGroup("resource_projection_runtime", "blocked", []string{
			"approval_state", "risk_policy", "dashboard_import_state", "agent_rollout_receipts",
		}, "cmdb_resource_projection_runtime_contract"),
	}
}

func cmdbResourceProjectionBlockedContracts() []model.CmdbResourceStatisticsContractGap {
	return []model.CmdbResourceStatisticsContractGap{
		{ID: "cmdb_resource_projection_runtime_contract", Status: cmdbBlockedByContract},
		{ID: "cmdb_resource_approval_runtime_contract", Status: cmdbBlockedByContract},
		{ID: "cmdb_dashboard_import_runtime_contract", Status: cmdbBlockedByContract},
		{ID: "cmdb_operation_risk_policy_contract", Status: cmdbBlockedByContract},
	}
}

func cmdbResourceProjectionAuditQuery() gin.H {
	return gin.H{
		"source":        "findx_audit",
		"scope":         "cmdb",
		"resource_type": "cmdb_resource_projection",
		"action":        "cmdb.resource_projection.read",
	}
}
