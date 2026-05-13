package main

import (
	"ai-workbench-api/internal/handler"

	"github.com/gin-gonic/gin"
)

func registerCmdbRoutes(v1 *gin.RouterGroup, mw routeMiddleware) {
	cmdb := v1.Group("/cmdb")
	{
		// 模型管理
		cmdb.GET("/tree", handler.CmdbTree)
		cmdb.GET("/objects", handler.ListCmdbObjects)
		cmdb.GET("/objects/:id", handler.GetCmdbObject)
		cmdb.POST("/objects", mw.adminRequired, handler.CreateCmdbObject)
		cmdb.PUT("/objects/:id", mw.adminRequired, handler.UpdateCmdbObject)
		cmdb.DELETE("/objects/:id", mw.adminRequired, handler.DeleteCmdbObject)
		cmdb.GET("/objects/:id/attributes", handler.ListCmdbAttributes)
		cmdb.POST("/objects/:id/attributes", mw.adminRequired, handler.CreateCmdbAttribute)
		cmdb.PUT("/attributes/:id", mw.adminRequired, handler.UpdateCmdbAttribute)
		cmdb.DELETE("/attributes/:id", mw.adminRequired, handler.DeleteCmdbAttribute)
		cmdb.GET("/objects/:id/instances-compatible", handler.ListCmdbInstancesCompatible)
		cmdb.GET("/objects/:id/instances", handler.ListCmdbInstances)
		cmdb.POST("/objects/:id/instances", mw.adminRequired, handler.CreateCmdbInstance)
		cmdb.GET("/instances/:id/detail-compatible", handler.GetCmdbInstanceDetailCompatible)
		cmdb.GET("/instances/:id/topology", handler.GetCmdbInstanceTopologyBlocked)
		cmdb.GET("/instances/:id", handler.GetCmdbInstance)
		cmdb.PUT("/instances/:id", mw.adminRequired, handler.UpdateCmdbInstance)
		cmdb.DELETE("/instances/:id", mw.adminRequired, handler.DeleteCmdbInstance)
		cmdb.GET("/monitor-bindings/*path", handler.GetCmdbMonitorBindingsBlocked)
		cmdb.POST("/monitor-bindings/*path", mw.adminRequired, handler.CreateCmdbMonitorBindingsBlocked)
		cmdb.POST("/discover", mw.adminRequired, handler.CmdbDiscover)

		// C03: SSH 终端 WebSocket
		cmdb.GET("/hosts/:id/terminal", handler.CmdbHostTerminal)

		// C04: 文件上传
		cmdb.POST("/hosts/:id/upload", mw.adminRequired, handler.CmdbHostUpload)

		// C05: 命令执行
		cmdb.POST("/hosts/:id/exec", mw.adminRequired, handler.CmdbHostExec)

		// C07: 数据库资产
		cmdb.GET("/databases", handler.CmdbListDatabases)
		cmdb.POST("/databases", mw.adminRequired, handler.CmdbCreateDatabase)
		cmdb.GET("/databases/:id", handler.CmdbGetDatabase)
		cmdb.DELETE("/databases/:id", mw.adminRequired, handler.CmdbDeleteDatabase)
		cmdb.POST("/databases/:id/test", handler.CmdbTestDatabaseConn)

		// C09: 部署任务
		cmdb.POST("/deploy-tasks", mw.adminRequired, handler.CmdbCreateDeployTask)
		cmdb.GET("/deploy-tasks", handler.CmdbListDeployTasks)
		cmdb.GET("/deploy-tasks/:id", handler.CmdbGetDeployTask)

		// C10: 导入
		cmdb.POST("/import/excel", mw.adminRequired, handler.CmdbImportExcel)
		cmdb.POST("/import/confirm", mw.adminRequired, handler.CmdbImportConfirm)
		cmdb.POST("/import/cloud", mw.adminRequired, handler.CmdbImportCloud)
	}
}
