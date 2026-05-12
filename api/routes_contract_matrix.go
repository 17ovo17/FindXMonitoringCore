package main

import (
	"ai-workbench-api/internal/handler"

	"github.com/gin-gonic/gin"
)

func registerContractMatrixRoutes(v1 *gin.RouterGroup, mw routeMiddleware) {
	v1.GET("/contracts/matrix", mw.monitorRequired("findx_agent", "read"), handler.ListContractMatrixEntries)
	v1.GET("/contracts/matrix/:id", mw.monitorRequired("findx_agent", "read"), handler.GetContractMatrixEntry)
	v1.POST("/contracts/matrix", mw.monitorRequired("findx_agent", "write"), handler.RegisterContractMatrixEntry)
}
