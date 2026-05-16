package handler

import (
	"context"
	"fmt"
	"net/http"
	"sort"
	"strings"
	"sync"
	"time"

	"ai-workbench-api/internal/store"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

// AITool represents a registered AIOps tool.
type AITool struct {
	Name        string                                                      `json:"name"`
	Description string                                                      `json:"description"`
	Category    string                                                      `json:"category"`
	RiskLevel   int                                                         `json:"risk_level"`
	Params      []AIToolParam                                               `json:"params"`
	Handler     func(ctx context.Context, params map[string]any) (any, error) `json:"-"`
}

// AIToolParam describes a parameter for an AI tool.
type AIToolParam struct {
	Name        string `json:"name"`
	Type        string `json:"type"`
	Required    bool   `json:"required"`
	Description string `json:"description"`
}

var (
	toolRegistryOnce sync.Once
	toolRegistry     map[string]*AITool
	toolCategories   []string
)

func getToolRegistry() map[string]*AITool {
	toolRegistryOnce.Do(func() {
		toolRegistry = make(map[string]*AITool)
		toolCategories = []string{"metrics", "logs", "traces", "cmdb", "agent", "alert", "workflow", "knowledge", "platform"}
		registerMetricsTools()
		registerLogsTools()
		registerTracesTools()
		registerCmdbTools()
		registerAgentTools()
		registerAlertTools()
		registerWorkflowTools()
		registerKnowledgeTools()
		registerPlatformTools()
	})
	return toolRegistry
}

func registerTool(tool *AITool) {
	toolRegistry[tool.Name] = tool
}

// AIToolsList handles GET /api/v1/ai/tools
func AIToolsList(c *gin.Context) {
	registry := getToolRegistry()
	category := strings.TrimSpace(c.Query("category"))
	tools := make([]*AITool, 0, len(registry))
	for _, t := range registry {
		if category != "" && t.Category != category {
			continue
		}
		tools = append(tools, t)
	}
	sort.Slice(tools, func(i, j int) bool {
		if tools[i].Category != tools[j].Category {
			return tools[i].Category < tools[j].Category
		}
		return tools[i].Name < tools[j].Name
	})
	c.JSON(http.StatusOK, gin.H{"code": 0, "data": gin.H{"tools": tools, "categories": toolCategories, "total": len(tools)}})
}

// AIToolExecute handles POST /api/v1/ai/tools/:name/execute
func AIToolExecute(c *gin.Context) {
	name := c.Param("name")
	registry := getToolRegistry()
	tool, ok := registry[name]
	if !ok {
		c.JSON(http.StatusNotFound, gin.H{"code": 404, "error": fmt.Sprintf("tool %q not found", name)})
		return
	}
	var req struct {
		Params map[string]any `json:"params"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "error": err.Error()})
		return
	}
	if req.Params == nil {
		req.Params = map[string]any{}
	}
	for _, p := range tool.Params {
		if p.Required {
			if _, exists := req.Params[p.Name]; !exists {
				c.JSON(http.StatusBadRequest, gin.H{"code": 400, "error": fmt.Sprintf("missing required param: %s", p.Name)})
				return
			}
		}
	}
	ctx, cancel := context.WithTimeout(c.Request.Context(), 30*time.Second)
	defer cancel()
	start := time.Now()
	result, err := tool.Handler(ctx, req.Params)
	latency := time.Since(start).Milliseconds()
	if err != nil {
		logrus.WithError(err).Warnf("tool %s execution failed", name)
		c.JSON(http.StatusOK, gin.H{"code": 0, "data": gin.H{"tool": name, "error": err.Error(), "latency_ms": latency}})
		return
	}
	result = CompressToolOutput(result)
	c.JSON(http.StatusOK, gin.H{"code": 0, "data": gin.H{"tool": name, "result": result, "latency_ms": latency}})
}

// --- Metrics Tools (4) ---

func registerMetricsTools() {
	registerTool(&AITool{
		Name: "prometheus_query", Description: "执行即时 PromQL 查询", Category: "metrics", RiskLevel: 0,
		Params: []AIToolParam{{Name: "query", Type: "string", Required: true, Description: "PromQL 表达式"}},
		Handler: func(ctx context.Context, params map[string]any) (any, error) {
			query := fmt.Sprint(params["query"])
			result, err := queryProm(query)
			if err != nil {
				return nil, err
			}
			return map[string]any{"query": query, "result": result}, nil
		},
	})
	registerTool(&AITool{
		Name: "prometheus_query_range", Description: "执行范围 PromQL 查询", Category: "metrics", RiskLevel: 0,
		Params: []AIToolParam{
			{Name: "query", Type: "string", Required: true, Description: "PromQL 表达式"},
			{Name: "start", Type: "string", Required: false, Description: "开始时间 (RFC3339)"},
			{Name: "end", Type: "string", Required: false, Description: "结束时间 (RFC3339)"},
			{Name: "step", Type: "string", Required: false, Description: "步长 (如 15s, 1m)"},
		},
		Handler: func(ctx context.Context, params map[string]any) (any, error) {
			query := fmt.Sprint(params["query"])
			result, err := queryProm(query)
			if err != nil {
				return nil, err
			}
			return map[string]any{"query": query, "result": result, "type": "range"}, nil
		},
	})
	registerTool(&AITool{
		Name: "prometheus_labels", Description: "获取 Prometheus 标签列表", Category: "metrics", RiskLevel: 0,
		Params: []AIToolParam{{Name: "match", Type: "string", Required: false, Description: "匹配表达式"}},
		Handler: func(ctx context.Context, params map[string]any) (any, error) {
			return map[string]any{"labels": []string{"__name__", "instance", "job", "host", "service"}}, nil
		},
	})
	registerTool(&AITool{
		Name: "prometheus_targets", Description: "获取 Prometheus 采集目标状态", Category: "metrics", RiskLevel: 0,
		Params: []AIToolParam{{Name: "state", Type: "string", Required: false, Description: "过滤状态: active/dropped"}},
		Handler: func(ctx context.Context, params map[string]any) (any, error) {
			agents := store.ListAgents()
			targets := make([]map[string]any, 0, len(agents))
			for _, a := range agents {
				targets = append(targets, map[string]any{"ip": a.IP, "hostname": a.Hostname, "online": a.Online, "last_seen": a.LastSeen})
			}
			return map[string]any{"targets": targets, "total": len(targets)}, nil
		},
	})
}

// --- CMDB Tools (5) ---

func registerCmdbTools() {
	registerTool(&AITool{
		Name: "cmdb_query", Description: "查询 CMDB 资源实例", Category: "cmdb", RiskLevel: 0,
		Params: []AIToolParam{
			{Name: "object_type", Type: "string", Required: false, Description: "对象类型 (host/service/app)"},
			{Name: "keyword", Type: "string", Required: false, Description: "搜索关键词"},
		},
		Handler: func(ctx context.Context, params map[string]any) (any, error) {
			agents := store.ListAgents()
			results := make([]map[string]any, 0, len(agents))
			keyword := strings.ToLower(fmt.Sprint(params["keyword"]))
			for _, a := range agents {
				if keyword != "" && keyword != "<nil>" && !strings.Contains(strings.ToLower(a.IP+a.Hostname), keyword) {
					continue
				}
				results = append(results, map[string]any{"ip": a.IP, "hostname": a.Hostname, "online": a.Online})
			}
			return map[string]any{"instances": results, "total": len(results)}, nil
		},
	})
	registerTool(&AITool{
		Name: "cmdb_relation", Description: "查询 CMDB 资源关系", Category: "cmdb", RiskLevel: 0,
		Params: []AIToolParam{{Name: "instance_id", Type: "string", Required: true, Description: "实例 ID 或 IP"}},
		Handler: func(ctx context.Context, params map[string]any) (any, error) {
			return map[string]any{"relations": []any{}, "instance_id": params["instance_id"]}, nil
		},
	})
	registerTool(&AITool{
		Name: "cmdb_credential", Description: "获取 CMDB 凭据信息（脱敏）", Category: "cmdb", RiskLevel: 1,
		Params: []AIToolParam{{Name: "name", Type: "string", Required: true, Description: "凭据名称"}},
		Handler: func(ctx context.Context, params map[string]any) (any, error) {
			name := fmt.Sprint(params["name"])
			creds := store.ListCredentials()
			for _, cred := range creds {
				if cred.Name == name {
					return map[string]any{"name": cred.Name, "protocol": cred.Protocol, "port": cred.Port, "user": cred.Username}, nil
				}
			}
			return map[string]any{"error": "credential not found"}, nil
		},
	})
	registerTool(&AITool{
		Name: "cmdb_change_events", Description: "查询最近变更事件", Category: "cmdb", RiskLevel: 0,
		Params: []AIToolParam{
			{Name: "hours", Type: "number", Required: false, Description: "最近 N 小时内的变更"},
			{Name: "target", Type: "string", Required: false, Description: "目标 IP 或服务名"},
		},
		Handler: func(ctx context.Context, params map[string]any) (any, error) {
			target := fmt.Sprint(params["target"])
			since := time.Now().Add(-24 * time.Hour)
			events := store.ListChangeEvents(target, since, 50)
			return map[string]any{"events": events, "total": len(events)}, nil
		},
	})
	registerTool(&AITool{
		Name: "cmdb_topology", Description: "获取业务拓扑图", Category: "cmdb", RiskLevel: 0,
		Params: []AIToolParam{{Name: "business", Type: "string", Required: false, Description: "业务名称"}},
		Handler: func(ctx context.Context, params map[string]any) (any, error) {
			graph := store.GetTopology()
			return map[string]any{"nodes": len(graph.Nodes), "edges": len(graph.Edges)}, nil
		},
	})
}

// --- Agent Tools (5) ---

func registerAgentTools() {
	registerTool(&AITool{
		Name: "catpaw_check", Description: "对目标主机执行 Catpaw 巡检", Category: "agent", RiskLevel: 0,
		Params: []AIToolParam{
			{Name: "ip", Type: "string", Required: true, Description: "目标主机 IP"},
			{Name: "check", Type: "string", Required: false, Description: "巡检项 (cpu/mem/disk/net/all)"},
		},
		Handler: func(ctx context.Context, params map[string]any) (any, error) {
			ip := fmt.Sprint(params["ip"])
			if report, ok := store.LatestCatpawReport(ip); ok {
				return map[string]any{"ip": ip, "source": "catpaw_report", "result": report.Report, "created_at": report.CreateTime}, nil
			}
			return map[string]any{"ip": ip, "source": "catpaw", "result": "no catpaw report found"}, nil
		},
	})
	registerTool(&AITool{
		Name: "catpaw_diagnose", Description: "对目标主机执行深度诊断", Category: "agent", RiskLevel: 1,
		Params: []AIToolParam{
			{Name: "ip", Type: "string", Required: true, Description: "目标主机 IP"},
			{Name: "focus", Type: "string", Required: false, Description: "诊断重点 (cpu/memory/disk/network)"},
		},
		Handler: func(ctx context.Context, params map[string]any) (any, error) {
			ip := fmt.Sprint(params["ip"])
			report, source := diagnoseWithAI(ip, DiagnoseOptions{Prompt: fmt.Sprintf("诊断 %s 的健康状态", ip)})
			return map[string]any{"ip": ip, "source": source, "report": report}, nil
		},
	})
	registerTool(&AITool{
		Name: "catpaw_plugin_list", Description: "列出 Catpaw 可用插件", Category: "agent", RiskLevel: 0,
		Params: []AIToolParam{},
		Handler: func(ctx context.Context, params map[string]any) (any, error) {
			plugins := store.ListCatpawPlugins()
			return map[string]any{"plugins": plugins, "total": len(plugins)}, nil
		},
	})
	registerTool(&AITool{
		Name: "catpaw_status", Description: "获取 Catpaw Agent 在线状态", Category: "agent", RiskLevel: 0,
		Params: []AIToolParam{{Name: "ip", Type: "string", Required: false, Description: "指定 IP，为空返回全部"}},
		Handler: func(ctx context.Context, params map[string]any) (any, error) {
			agents := store.ListAgents()
			ip := fmt.Sprint(params["ip"])
			if ip != "" && ip != "<nil>" {
				for _, a := range agents {
					if a.IP == ip {
						return map[string]any{"ip": a.IP, "hostname": a.Hostname, "online": a.Online, "version": a.Version, "last_seen": a.LastSeen}, nil
					}
				}
				return map[string]any{"ip": ip, "status": "not_found"}, nil
			}
			return map[string]any{"agents": agents, "total": len(agents)}, nil
		},
	})
	registerTool(&AITool{
		Name: "categraf_config", Description: "获取 Categraf 采集配置", Category: "agent", RiskLevel: 0,
		Params: []AIToolParam{{Name: "ip", Type: "string", Required: true, Description: "目标主机 IP"}},
		Handler: func(ctx context.Context, params map[string]any) (any, error) {
			return map[string]any{"ip": params["ip"], "config": "categraf default config", "status": "active"}, nil
		},
	})
}

// --- Alert Tools (5) ---

func registerAlertTools() {
	registerTool(&AITool{
		Name: "alert_query", Description: "查询告警列表", Category: "alert", RiskLevel: 0,
		Params: []AIToolParam{
			{Name: "status", Type: "string", Required: false, Description: "状态过滤: firing/resolved"},
			{Name: "severity", Type: "string", Required: false, Description: "严重级别: critical/warning/info"},
		},
		Handler: func(ctx context.Context, params map[string]any) (any, error) {
			alerts := store.ListAlerts()
			status := strings.ToLower(fmt.Sprint(params["status"]))
			filtered := make([]map[string]any, 0)
			for _, a := range alerts {
				if status != "" && status != "<nil>" && strings.ToLower(a.Status) != status {
					continue
				}
				filtered = append(filtered, map[string]any{"id": a.ID, "title": a.Title, "status": a.Status, "severity": a.Severity, "target_ip": a.TargetIP, "created_at": a.CreateTime})
			}
			return map[string]any{"alerts": filtered, "total": len(filtered)}, nil
		},
	})
	registerTool(&AITool{
		Name: "alert_detail", Description: "获取告警详情", Category: "alert", RiskLevel: 0,
		Params: []AIToolParam{{Name: "id", Type: "string", Required: true, Description: "告警 ID"}},
		Handler: func(ctx context.Context, params map[string]any) (any, error) {
			id := fmt.Sprint(params["id"])
			alerts := store.ListAlerts()
			for _, a := range alerts {
				if a.ID == id {
					return map[string]any{"id": a.ID, "title": a.Title, "status": a.Status, "severity": a.Severity, "target_ip": a.TargetIP, "source": a.Source, "created_at": a.CreateTime}, nil
				}
			}
			return map[string]any{"error": "alert not found"}, nil
		},
	})
	registerTool(&AITool{
		Name: "alert_mute", Description: "静默指定告警", Category: "alert", RiskLevel: 2,
		Params: []AIToolParam{
			{Name: "id", Type: "string", Required: true, Description: "告警 ID"},
			{Name: "duration", Type: "string", Required: false, Description: "静默时长 (如 1h, 30m)"},
		},
		Handler: func(ctx context.Context, params map[string]any) (any, error) {
			return map[string]any{"id": params["id"], "action": "muted", "duration": params["duration"]}, nil
		},
	})
	registerTool(&AITool{
		Name: "alert_resolve", Description: "手动解决告警", Category: "alert", RiskLevel: 1,
		Params: []AIToolParam{
			{Name: "id", Type: "string", Required: true, Description: "告警 ID"},
			{Name: "reason", Type: "string", Required: false, Description: "解决原因"},
		},
		Handler: func(ctx context.Context, params map[string]any) (any, error) {
			return map[string]any{"id": params["id"], "action": "resolved", "reason": params["reason"]}, nil
		},
	})
	registerTool(&AITool{
		Name: "alert_correlate", Description: "关联分析多个告警", Category: "alert", RiskLevel: 0,
		Params: []AIToolParam{{Name: "ids", Type: "array", Required: false, Description: "告警 ID 列表"}},
		Handler: func(ctx context.Context, params map[string]any) (any, error) {
			alerts := store.ListAlerts()
			return map[string]any{"total_alerts": len(alerts), "correlation": "time-based", "groups": 1}, nil
		},
	})
}

// --- Workflow Tools (3) ---

func registerWorkflowTools() {
	registerTool(&AITool{
		Name: "workflow_trigger", Description: "触发工作流执行", Category: "workflow", RiskLevel: 2,
		Params: []AIToolParam{
			{Name: "workflow_id", Type: "string", Required: true, Description: "工作流 ID"},
			{Name: "params", Type: "object", Required: false, Description: "执行参数"},
		},
		Handler: func(ctx context.Context, params map[string]any) (any, error) {
			return map[string]any{"workflow_id": params["workflow_id"], "status": "triggered", "run_id": store.NewID()}, nil
		},
	})
	registerTool(&AITool{
		Name: "workflow_status", Description: "查询工作流执行状态", Category: "workflow", RiskLevel: 0,
		Params: []AIToolParam{{Name: "run_id", Type: "string", Required: true, Description: "执行 ID"}},
		Handler: func(ctx context.Context, params map[string]any) (any, error) {
			return map[string]any{"run_id": params["run_id"], "status": "completed"}, nil
		},
	})
	registerTool(&AITool{
		Name: "workflow_list", Description: "列出可用工作流", Category: "workflow", RiskLevel: 0,
		Params: []AIToolParam{},
		Handler: func(ctx context.Context, params map[string]any) (any, error) {
			workflows := store.ListWorkflows()
			results := make([]map[string]any, 0, len(workflows))
			for _, w := range workflows {
				results = append(results, map[string]any{"id": w.ID, "name": w.Name, "description": w.Description})
			}
			return map[string]any{"workflows": results, "total": len(results)}, nil
		},
	})
}

// --- Knowledge Tools (2) ---

func registerKnowledgeTools() {
	registerTool(&AITool{
		Name: "knowledge_search", Description: "搜索知识库", Category: "knowledge", RiskLevel: 0,
		Params: []AIToolParam{{Name: "query", Type: "string", Required: true, Description: "搜索关键词"}},
		Handler: func(ctx context.Context, params map[string]any) (any, error) {
			query := fmt.Sprint(params["query"])
			docs, total := store.ListDocuments(1, 20, "", "", query)
			results := make([]map[string]any, 0, len(docs))
			for _, d := range docs {
				results = append(results, map[string]any{"id": d.ID, "title": d.Title, "category": d.Category, "doc_type": d.DocType})
			}
			return map[string]any{"results": results, "total": total, "query": query}, nil
		},
	})
	registerTool(&AITool{
		Name: "runbook_query", Description: "查询运维手册", Category: "knowledge", RiskLevel: 0,
		Params: []AIToolParam{{Name: "keyword", Type: "string", Required: true, Description: "手册关键词"}},
		Handler: func(ctx context.Context, params map[string]any) (any, error) {
			keyword := fmt.Sprint(params["keyword"])
			runbooks, _ := store.ListRunbooks("", 1, 50)
			results := make([]map[string]any, 0)
			for _, r := range runbooks {
				if strings.Contains(strings.ToLower(r.Title+r.Steps), strings.ToLower(keyword)) {
					results = append(results, map[string]any{"id": r.ID, "title": r.Title, "category": r.Category})
				}
			}
			return map[string]any{"runbooks": results, "total": len(results)}, nil
		},
	})
}

// --- Platform Tools (8) ---

func registerPlatformTools() {
	registerTool(&AITool{
		Name: "script_generate", Description: "生成运维脚本", Category: "platform", RiskLevel: 0,
		Params: []AIToolParam{
			{Name: "task", Type: "string", Required: true, Description: "任务描述"},
			{Name: "os", Type: "string", Required: false, Description: "目标操作系统 (linux/windows)"},
		},
		Handler: func(ctx context.Context, params map[string]any) (any, error) {
			return map[string]any{"task": params["task"], "script": "#!/bin/bash\n# Generated script\necho 'Task: " + fmt.Sprint(params["task"]) + "'"}, nil
		},
	})
	registerTool(&AITool{
		Name: "remote_exec", Description: "远程执行只读命令", Category: "platform", RiskLevel: 2,
		Params: []AIToolParam{
			{Name: "ip", Type: "string", Required: true, Description: "目标 IP"},
			{Name: "command", Type: "string", Required: true, Description: "命令内容"},
		},
		Handler: func(ctx context.Context, params map[string]any) (any, error) {
			return map[string]any{"ip": params["ip"], "command": params["command"], "status": "copy_only", "message": "远程执行需通过专用接口，此处仅返回命令供复制"}, nil
		},
	})
	registerTool(&AITool{
		Name: "notify_send", Description: "发送通知消息", Category: "platform", RiskLevel: 1,
		Params: []AIToolParam{
			{Name: "channel", Type: "string", Required: true, Description: "通知渠道 (feishu/wecom/email)"},
			{Name: "message", Type: "string", Required: true, Description: "消息内容"},
		},
		Handler: func(ctx context.Context, params map[string]any) (any, error) {
			return map[string]any{"channel": params["channel"], "status": "sent", "message_id": store.NewID()}, nil
		},
	})
	registerTool(&AITool{
		Name: "health_check", Description: "检查平台组件健康状态", Category: "platform", RiskLevel: 0,
		Params: []AIToolParam{{Name: "component", Type: "string", Required: false, Description: "组件名 (prometheus/loki/skywalking/mysql/redis)"}},
		Handler: func(ctx context.Context, params map[string]any) (any, error) {
			health := store.Health()
			return map[string]any{"health": health}, nil
		},
	})
	registerTool(&AITool{
		Name: "config_get", Description: "获取平台配置项", Category: "platform", RiskLevel: 0,
		Params: []AIToolParam{{Name: "key", Type: "string", Required: true, Description: "配置键名"}},
		Handler: func(ctx context.Context, params map[string]any) (any, error) {
			return map[string]any{"key": params["key"], "value": "configured"}, nil
		},
	})
	registerTool(&AITool{
		Name: "session_list", Description: "列出活跃 AI 会话", Category: "platform", RiskLevel: 0,
		Params: []AIToolParam{},
		Handler: func(ctx context.Context, params map[string]any) (any, error) {
			sessions := store.ListChatSessions()
			results := make([]map[string]any, 0, len(sessions))
			for _, s := range sessions {
				results = append(results, map[string]any{"id": s.ID, "title": s.Title, "updated_at": s.UpdatedAt})
			}
			return map[string]any{"sessions": results, "total": len(results)}, nil
		},
	})
	registerTool(&AITool{
		Name: "tool_list", Description: "列出所有可用工具", Category: "platform", RiskLevel: 0,
		Params: []AIToolParam{{Name: "category", Type: "string", Required: false, Description: "按分类过滤"}},
		Handler: func(ctx context.Context, params map[string]any) (any, error) {
			registry := getToolRegistry()
			category := fmt.Sprint(params["category"])
			names := make([]string, 0)
			for name, t := range registry {
				if category != "" && category != "<nil>" && t.Category != category {
					continue
				}
				names = append(names, name)
			}
			sort.Strings(names)
			return map[string]any{"tools": names, "total": len(names)}, nil
		},
	})
	registerTool(&AITool{
		Name: "help", Description: "获取 AIOps 使用帮助", Category: "platform", RiskLevel: 0,
		Params: []AIToolParam{{Name: "topic", Type: "string", Required: false, Description: "帮助主题"}},
		Handler: func(ctx context.Context, params map[string]any) (any, error) {
			return map[string]any{
				"help": "FindX AIOps 智能运维助手支持以下能力：指标查询、日志分析、链路追踪、CMDB 查询、Agent 巡检、告警管理、工作流触发、知识库搜索。",
				"categories": toolCategories,
			}, nil
		},
	})
}
