package main

import (
	"ai-workbench-api/internal/handler"

	"github.com/gin-gonic/gin"
)

func registerCmdbRoutes(v1 *gin.RouterGroup, mw routeMiddleware) {
	cmdb := v1.Group("/cmdb")
	{
		// 模型管理
		cmdb.GET("/resource-projections", mw.monitorRequired("cmdb.resource.projection", "read"), handler.GetCmdbResourceProjection)
		cmdb.GET("/resource-statistics", mw.monitorRequired("cmdb.stats", "read"), handler.GetCmdbResourceStatistics)
		cmdb.GET("/approvals", mw.monitorRequired("cmdb.approval", "read"), handler.GetCmdbResourceApprovals)
		cmdb.POST("/approvals", mw.monitorRequired("cmdb.approval", "approve"), handler.CreateCmdbResourceApprovalRequest)
		cmdb.GET("/approvals/:id", mw.monitorRequired("cmdb.approval", "read"), handler.GetCmdbResourceApproval)
		cmdb.POST("/approvals/:id/review", mw.monitorRequired("cmdb.approval", "review"), handler.ReviewCmdbResourceApproval)
		cmdb.GET("/tree", mw.monitorRequired("cmdb.model", "read"), handler.CmdbTree)
		cmdb.GET("/objects", mw.monitorRequired("cmdb.model", "read"), handler.ListCmdbObjects)
		cmdb.GET("/objects/:id", mw.monitorRequired("cmdb.model", "read"), handler.GetCmdbObject)
		cmdb.POST("/objects", mw.monitorRequired("cmdb.model", "create"), handler.CreateCmdbObject)
		cmdb.PUT("/objects/:id", mw.monitorRequired("cmdb.model", "update"), handler.UpdateCmdbObject)
		cmdb.DELETE("/objects/:id", mw.monitorRequired("cmdb.model", "delete"), handler.DeleteCmdbObject)
		cmdb.GET("/objects/:id/attributes", mw.monitorRequired("cmdb.attribute", "read"), handler.ListCmdbAttributes)
		cmdb.POST("/objects/:id/attributes", mw.monitorRequired("cmdb.attribute", "create"), handler.CreateCmdbAttribute)
		cmdb.PUT("/attributes/:id", mw.monitorRequired("cmdb.attribute", "update"), handler.UpdateCmdbAttribute)
		cmdb.DELETE("/attributes/:id", mw.monitorRequired("cmdb.attribute", "delete"), handler.DeleteCmdbAttribute)
		cmdb.GET("/objects/:id/instances-compatible", mw.monitorRequired("cmdb.instance", "read"), handler.ListCmdbInstancesCompatible)
		cmdb.GET("/objects/:id/instances", mw.monitorRequired("cmdb.instance", "read"), handler.ListCmdbInstances)
		cmdb.POST("/objects/:id/instances", mw.monitorRequired("cmdb.instance", "create"), handler.CreateCmdbInstance)
		cmdb.GET("/instances/:id/detail-compatible", mw.monitorRequired("cmdb.instance", "read"), handler.GetCmdbInstanceDetailCompatible)
		cmdb.GET("/instances/:id/topology", mw.monitorRequired("cmdb.relation", "read"), handler.GetCmdbInstanceTopologyBlocked)
		cmdb.GET("/instances/:id", mw.monitorRequired("cmdb.instance", "read"), handler.GetCmdbInstance)
		cmdb.PUT("/instances/:id", mw.monitorRequired("cmdb.instance", "update"), handler.UpdateCmdbInstance)
		cmdb.DELETE("/instances/:id", mw.monitorRequired("cmdb.instance", "delete"), handler.DeleteCmdbInstance)
		cmdb.GET("/monitor-bindings/*path", mw.monitorRequired("cmdb.relation", "read"), handler.GetCmdbMonitorBindingsBlocked)
		cmdb.POST("/monitor-bindings/*path", mw.monitorRequired("cmdb.relation", "create"), handler.CreateCmdbMonitorBindingsBlocked)
		cmdb.POST("/relation-actions", mw.monitorRequired("cmdb.relation", "create"), handler.CreateCmdbRelationActionRequest)
		cmdb.GET("/relation-actions", mw.monitorRequired("cmdb.relation", "read"), handler.ListCmdbRelationActionRequests)
		cmdb.GET("/relation-actions/:id/receipts", mw.monitorRequired("cmdb.relation", "read"), handler.GetCmdbRelationActionReceipts)
		cmdb.GET("/relation-actions/:id", mw.monitorRequired("cmdb.relation", "read"), handler.GetCmdbRelationActionRequest)
		cmdb.POST("/discover", mw.monitorRequired("cmdb.import", "import"), handler.CmdbDiscover)

		// C03: SSH 终端 WebSocket
		cmdb.GET("/hosts/:id/terminal", mw.monitorRequired("cmdb.terminal", "open"), handler.CmdbHostTerminal)
		cmdb.GET("/hosts/:id/ai-session", mw.monitorRequired("aiops.session", "read"), handler.GetCmdbHostAISessionPreflight)
		cmdb.POST("/hosts/:id/ai-session", mw.monitorRequired("aiops.session", "create"), handler.CreateCmdbHostAISessionPreflight)

		// C04: 文件上传
		cmdb.POST("/hosts/:id/upload", mw.monitorRequired("cmdb.file", "upload"), handler.CmdbHostUpload)

		// C05: 命令执行
		cmdb.POST("/hosts/:id/exec", mw.monitorRequired("cmdb.command", "exec"), handler.CmdbHostExec)

		// C07: 数据库资产
		cmdb.GET("/databases", mw.monitorRequired("cmdb.database", "read"), handler.CmdbListDatabases)
		cmdb.POST("/databases", mw.monitorRequired("cmdb.database", "create"), handler.CmdbCreateDatabase)
		cmdb.GET("/databases/:id", mw.monitorRequired("cmdb.database", "read"), handler.CmdbGetDatabase)
		cmdb.DELETE("/databases/:id", mw.monitorRequired("cmdb.database", "delete"), handler.CmdbDeleteDatabase)
		cmdb.POST("/databases/:id/test", mw.monitorRequired("cmdb.database", "test"), handler.CmdbTestDatabaseConn)

		// C09: 部署任务
		cmdb.POST("/deploy-tasks", mw.monitorRequired("cmdb.command", "exec"), handler.CmdbCreateDeployTask)
		cmdb.GET("/deploy-tasks", mw.monitorRequired("cmdb.command", "read"), handler.CmdbListDeployTasks)
		cmdb.GET("/deploy-tasks/:id", mw.monitorRequired("cmdb.command", "read"), handler.CmdbGetDeployTask)

		// C10: 导入
		cmdb.POST("/import/excel", mw.monitorRequired("cmdb.import", "import"), handler.CmdbImportExcel)
		cmdb.POST("/import/confirm", mw.monitorRequired("cmdb.import", "confirm"), handler.CmdbImportConfirm)
		cmdb.POST("/import/cloud", mw.monitorRequired("cmdb.import", "import"), handler.CmdbImportCloud)

		// F1: 预置资源模型
		cmdb.GET("/model-presets", mw.monitorRequired("cmdb.model", "read"), handler.CmdbListModelPresets)
		cmdb.GET("/model-presets/:id", mw.monitorRequired("cmdb.model", "read"), handler.CmdbGetModelPreset)
		cmdb.POST("/model-presets/:id/apply", mw.monitorRequired("cmdb.model", "create"), handler.CmdbApplyModelPreset)

		// F2: 批量操作
		cmdb.POST("/batch/update-group", mw.monitorRequired("cmdb.instance", "update"), handler.CmdbBatchUpdateGroup)
		cmdb.POST("/batch/update-owner", mw.monitorRequired("cmdb.instance", "update"), handler.CmdbBatchUpdateOwner)
		cmdb.POST("/batch/update-maintenance", mw.monitorRequired("cmdb.instance", "update"), handler.CmdbBatchUpdateMaintenance)
		cmdb.POST("/batch/delete", mw.monitorRequired("cmdb.instance", "delete"), handler.CmdbBatchDelete)
		cmdb.POST("/batch/deploy-agent", mw.monitorRequired("cmdb.command", "exec"), handler.CmdbBatchDeployAgent)

		// F3: 数据中心/机柜视图
		cmdb.GET("/datacenters", mw.monitorRequired("cmdb.datacenter", "read"), handler.CmdbListDatacenters)
		cmdb.POST("/datacenters", mw.monitorRequired("cmdb.datacenter", "create"), handler.CmdbCreateDatacenter)
		cmdb.GET("/datacenters/:id/racks", mw.monitorRequired("cmdb.datacenter", "read"), handler.CmdbListRacks)
		cmdb.POST("/datacenters/:id/racks", mw.monitorRequired("cmdb.datacenter", "create"), handler.CmdbCreateRack)
		cmdb.GET("/datacenters/:id/racks/:rackId/units", mw.monitorRequired("cmdb.datacenter", "read"), handler.CmdbGetRackUnits)
		cmdb.POST("/datacenters/:id/racks/:rackId/units", mw.monitorRequired("cmdb.datacenter", "create"), handler.CmdbAssignRackUnit)
		cmdb.GET("/racks/:id/devices", mw.monitorRequired("cmdb.datacenter", "read"), handler.CmdbGetRackDevices)

		// F4: 自动发现增强
		cmdb.GET("/discovery/rules", mw.monitorRequired("cmdb.discovery", "read"), handler.CmdbListDiscoveryRules)
		cmdb.POST("/discovery/rules", mw.monitorRequired("cmdb.discovery", "create"), handler.CmdbCreateDiscoveryRule)
		cmdb.POST("/discovery/rules/:id/run", mw.monitorRequired("cmdb.discovery", "exec"), handler.CmdbRunDiscoveryRule)
		cmdb.GET("/discovery/results", mw.monitorRequired("cmdb.discovery", "read"), handler.CmdbListDiscoveryResults)
		cmdb.POST("/discovery/results/:id/approve", mw.monitorRequired("cmdb.discovery", "approve"), handler.CmdbApproveDiscoveryResult)

		// F5: 审批工作流
		cmdb.GET("/approval-workflow", mw.monitorRequired("cmdb.approval", "read"), handler.CmdbListApprovalsPending)
		cmdb.GET("/approval-workflow/my-requests", mw.monitorRequired("cmdb.approval", "read"), handler.CmdbListApprovalsMyRequests)
		cmdb.POST("/approval-workflow", mw.monitorRequired("cmdb.approval", "create"), handler.CmdbCreateApprovalRequest)
		cmdb.POST("/approval-workflow/:id/approve", mw.monitorRequired("cmdb.approval", "approve"), handler.CmdbApproveApproval)
		cmdb.POST("/approval-workflow/:id/reject", mw.monitorRequired("cmdb.approval", "approve"), handler.CmdbRejectApproval)
		cmdb.GET("/approval-workflow/archived", mw.monitorRequired("cmdb.approval", "read"), handler.CmdbListApprovalsArchived)
	}
}
