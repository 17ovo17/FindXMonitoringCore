package handler

import (
	"encoding/json"
	"net/http"
	"strings"

	"ai-workbench-api/internal/model"
	"ai-workbench-api/internal/store"

	"github.com/gin-gonic/gin"
)

// McpTool 描述 MCP Server 支持的单个工具。
type McpTool struct {
	Name        string         `json:"name"`
	Description string         `json:"description"`
	Parameters  map[string]any `json:"parameters,omitempty"`
}

// mcpToolRegistry 按 server type 注册可用工具。
var mcpToolRegistry = map[string][]McpTool{
	"cmdb": {
		{Name: "list_hosts", Description: "列出 CMDB 中所有主机", Parameters: map[string]any{}},
		{Name: "get_host", Description: "根据 ID 获取主机详情", Parameters: map[string]any{"id": "string (required)"}},
		{Name: "search_hosts", Description: "按关键字搜索主机", Parameters: map[string]any{"keyword": "string (required)"}},
	},
}

// McpListTools POST /api/v1/mcp/servers/:id/list-tools
func McpListTools(c *gin.Context) {
	server, ok := store.GetMcpServer(c.Param("id"))
	if !ok {
		c.JSON(http.StatusNotFound, gin.H{"error": "mcp server not found"})
		return
	}
	tools := mcpToolRegistry[strings.ToLower(server.Type)]
	if tools == nil {
		tools = []McpTool{}
	}
	c.JSON(http.StatusOK, gin.H{"server_id": server.ID, "tools": tools})
}

// mcpCallToolRequest 调用工具的请求体。
type mcpCallToolRequest struct {
	ToolName  string         `json:"tool_name" binding:"required"`
	Arguments map[string]any `json:"arguments"`
}

// McpCallTool POST /api/v1/mcp/servers/:id/call-tool
func McpCallTool(c *gin.Context) {
	server, ok := store.GetMcpServer(c.Param("id"))
	if !ok {
		c.JSON(http.StatusNotFound, gin.H{"error": "mcp server not found"})
		return
	}
	var req mcpCallToolRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "tool_name is required"})
		return
	}

	// 验证工具是否存在
	tools := mcpToolRegistry[strings.ToLower(server.Type)]
	found := false
	for _, t := range tools {
		if t.Name == req.ToolName {
			found = true
			break
		}
	}
	if !found {
		c.JSON(http.StatusBadRequest, gin.H{"error": "tool not found: " + req.ToolName})
		return
	}

	// 根据 server type 分发执行
	result, err := executeMcpTool(server, req.ToolName, req.Arguments)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"server_id": server.ID,
		"tool_name": req.ToolName,
		"result":    result,
	})
}

func executeMcpTool(server *model.McpServer, toolName string, args map[string]any) (any, error) {
	switch strings.ToLower(server.Type) {
	case "cmdb":
		return executeCmdbTool(toolName, args)
	default:
		return nil, nil
	}
}

func executeCmdbTool(toolName string, args map[string]any) (any, error) {
	switch toolName {
	case "list_hosts":
		return cmdbListHosts(), nil
	case "get_host":
		id := argString(args, "id")
		if id == "" {
			return nil, nil
		}
		return cmdbGetHost(id), nil
	case "search_hosts":
		keyword := argString(args, "keyword")
		return cmdbSearchHosts(keyword), nil
	default:
		return nil, nil
	}
}

// cmdbListHosts 从 CMDB 实例中获取主机列表。
func cmdbListHosts() []map[string]any {
	// 获取所有 object，找到主机类型的 objectID
	objects := store.ListCmdbObjects("")
	hosts := make([]map[string]any, 0)
	for _, obj := range objects {
		instances, _ := store.ListCmdbInstances(obj.ID, 1, 100)
		for _, inst := range instances {
			hosts = append(hosts, cmdbInstanceToMap(inst))
		}
	}
	return hosts
}

// cmdbGetHost 根据 ID 获取单个主机。
func cmdbGetHost(id string) map[string]any {
	inst, ok := store.GetCmdbInstance(id)
	if !ok {
		return nil
	}
	return cmdbInstanceToMap(*inst)
}

// cmdbSearchHosts 按关键字搜索主机。
func cmdbSearchHosts(keyword string) []map[string]any {
	objects := store.ListCmdbObjects("")
	keyword = strings.ToLower(keyword)
	results := make([]map[string]any, 0)
	for _, obj := range objects {
		instances, _ := store.ListCmdbInstances(obj.ID, 1, 100)
		for _, inst := range instances {
			if matchCmdbInstanceData(inst, keyword) {
				results = append(results, cmdbInstanceToMap(inst))
			}
		}
	}
	return results
}

func cmdbInstanceToMap(inst model.CmdbInstance) map[string]any {
	m := map[string]any{
		"id":        inst.ID,
		"object_id": inst.ObjectID,
	}
	var data map[string]any
	if err := json.Unmarshal([]byte(inst.Data), &data); err == nil {
		m["data"] = data
	} else {
		m["data"] = inst.Data
	}
	return m
}

func matchCmdbInstanceData(inst model.CmdbInstance, keyword string) bool {
	if keyword == "" {
		return true
	}
	if strings.Contains(strings.ToLower(inst.ID), keyword) {
		return true
	}
	if strings.Contains(strings.ToLower(inst.Data), keyword) {
		return true
	}
	return false
}

func argString(args map[string]any, key string) string {
	if args == nil {
		return ""
	}
	v, ok := args[key]
	if !ok {
		return ""
	}
	s, ok := v.(string)
	if !ok {
		return ""
	}
	return strings.TrimSpace(s)
}
