package handler

import (
	"ai-workbench-api/internal/model"

	"github.com/gin-gonic/gin"
)

// cmdbMonitorBindingRuntimeReadBlockedEnvelope 运行时读取直接返回数据，不再阻断。
func cmdbMonitorBindingRuntimeReadBlockedEnvelope(_ string, _ []model.CmdbMonitorBinding) (gin.H, bool) {
	return nil, false
}
