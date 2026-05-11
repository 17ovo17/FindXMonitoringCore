package main

import (
	"ai-workbench-api/internal/handler"

	"github.com/gin-gonic/gin"
)

func registerCmdbRoutes(v1 *gin.RouterGroup, mw routeMiddleware) {
	cmdb := v1.Group("/cmdb")
	{
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
		cmdb.GET("/objects/:id/instances", handler.ListCmdbInstances)
		cmdb.POST("/objects/:id/instances", mw.adminRequired, handler.CreateCmdbInstance)
		cmdb.GET("/instances/:id", handler.GetCmdbInstance)
		cmdb.PUT("/instances/:id", mw.adminRequired, handler.UpdateCmdbInstance)
		cmdb.DELETE("/instances/:id", mw.adminRequired, handler.DeleteCmdbInstance)
	}
}
