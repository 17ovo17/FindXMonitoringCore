package handler

import (
	"net/http"
	"strings"

	"ai-workbench-api/internal/store"

	"github.com/gin-gonic/gin"
)

// CategrafHTTPProviderConfigs 处理 Categraf agent 的配置拉取请求
// Categraf HTTP Provider 协议：
// 1. Agent 发送 GET 请求，携带 query params: agent_hostname, version, timestamp, 以及 global labels
// 2. Server 根据 agent 标识查找分配的配置
// 3. 返回 version + configs map（inputName -> checksum -> ConfigWithFormat）
// 4. Agent 比较 version，如果不同则应用差异（新增/删除插件）
//
// GET /categraf/configs
func CategrafHTTPProviderConfigs(c *gin.Context) {
	if !validCategrafProviderToken(c) {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid agent provider token"})
		return
	}

	agentIdent := resolveCategrafAgentIdent(c)
	if agentIdent == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "agent identity is required (agent_hostname, ident, host, or agent query param)",
			"version": "",
			"configs": map[string]interface{}{},
		})
		return
	}

	// 获取 agent 当前上报的 version，用于判断是否需要返回完整配置
	clientVersion := strings.TrimSpace(c.Query("version"))

	// 构建 provider 响应
	resp := store.BuildCategrafProviderResponse(agentIdent)

	// 如果 agent 上报的 version 与服务端一致，返回空 configs 表示无变更
	if clientVersion != "" && clientVersion == resp.Version {
		c.JSON(http.StatusOK, gin.H{
			"version": resp.Version,
			"configs": map[string]interface{}{},
		})
		return
	}

	c.JSON(http.StatusOK, resp)
}

// resolveCategrafAgentIdent 从请求中解析 agent 标识
// Categraf 源码中通过 query param 传递: agent_hostname, 以及 global labels
// 兼容多种标识方式
func resolveCategrafAgentIdent(c *gin.Context) string {
	candidates := []string{
		c.Query("agent_hostname"),
		c.Query("ident"),
		c.Query("host"),
		c.Query("hostname"),
		c.Query("agent"),
		c.Query("agent_id"),
		c.Query("target_id"),
		c.GetHeader("X-Agent-Ident"),
	}
	for _, v := range candidates {
		if clean := strings.TrimSpace(v); clean != "" {
			if looksSensitive(clean) {
				continue
			}
			return clean
		}
	}
	return ""
}
