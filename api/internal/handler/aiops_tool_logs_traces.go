package handler

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"ai-workbench-api/internal/model"
	"ai-workbench-api/internal/store"
)

// registerLogsTools registers the 5 log query tools.
func registerLogsTools() {
	registerTool(&AITool{
		Name: "logs_query", Description: "查询日志（代理到 Loki/FindX）", Category: "logs", RiskLevel: 0,
		Params: []AIToolParam{
			{Name: "query", Type: "string", Required: true, Description: "LogQL 查询表达式或关键词"},
			{Name: "limit", Type: "number", Required: false, Description: "返回条数限制"},
		},
		Handler: func(ctx context.Context, params map[string]any) (any, error) {
			query := fmt.Sprint(params["query"])
			req := model.LogQueryRequest{
				Query: query,
				Limit: 50,
				Page:  1,
			}
			resp, err := store.QueryFindXAuditLogs(req)
			if err != nil {
				return map[string]any{"query": query, "error": err.Error(), "logs": []any{}}, nil
			}
			return map[string]any{"query": query, "logs": resp.Items, "total": resp.Total}, nil
		},
	})
	registerTool(&AITool{
		Name: "logs_context", Description: "查看某条日志前后的上下文日志", Category: "logs", RiskLevel: 0,
		Params: []AIToolParam{
			{Name: "log_id", Type: "string", Required: true, Description: "日志ID"},
			{Name: "before", Type: "number", Required: false, Description: "前N条"},
			{Name: "after", Type: "number", Required: false, Description: "后N条"},
		},
		Handler: func(ctx context.Context, params map[string]any) (any, error) {
			logID := fmt.Sprint(params["log_id"])
			before := logsToolParamInt(params, "before", 5)
			after := logsToolParamInt(params, "after", 5)
			req := model.LogContextRequest{LogID: logID, Before: before, After: after}
			resp, err := store.ContextFindXAuditLogs(req)
			if err != nil {
				return map[string]any{"log_id": logID, "error": err.Error()}, nil
			}
			return map[string]any{"log_id": logID, "before": resp.Before, "after": resp.After, "center": resp.Center, "total": resp.Total}, nil
		},
	})
	registerTool(&AITool{
		Name: "logs_aggregate", Description: "日志聚合统计", Category: "logs", RiskLevel: 0,
		Params: []AIToolParam{
			{Name: "group_by", Type: "string", Required: false, Description: "聚合字段 (action/resource_type/status)"},
			{Name: "action", Type: "string", Required: false, Description: "动作过滤"},
		},
		Handler: func(ctx context.Context, params map[string]any) (any, error) {
			groupBy := logsToolParamStr(params, "group_by", "action")
			action := logsToolParamStr(params, "action", "")
			req := model.LogAggregateRequest{
				GroupBy: groupBy,
				Action:  action,
				Page:    1,
				Limit:   100,
			}
			resp, err := store.AggregateFindXAuditLogs(req)
			if err != nil {
				return map[string]any{"group_by": groupBy, "error": err.Error()}, nil
			}
			return map[string]any{"group_by": groupBy, "type": "aggregate", "buckets": resp.Buckets, "total": resp.Total}, nil
		},
	})
	registerTool(&AITool{
		Name: "logs_tail", Description: "实时日志流（最近 N 条）", Category: "logs", RiskLevel: 0,
		Params: []AIToolParam{
			{Name: "query", Type: "string", Required: false, Description: "过滤关键词"},
			{Name: "limit", Type: "number", Required: false, Description: "返回条数"},
		},
		Handler: func(ctx context.Context, params map[string]any) (any, error) {
			query := logsToolParamStr(params, "query", "")
			req := model.LogQueryRequest{
				Query: query,
				Page:  1,
				Limit: 20,
			}
			resp, err := store.QueryFindXAuditLogs(req)
			if err != nil {
				return map[string]any{"query": query, "error": err.Error(), "mode": "tail"}, nil
			}
			return map[string]any{"query": query, "logs": resp.Items, "total": resp.Total, "mode": "tail"}, nil
		},
	})
	registerTool(&AITool{
		Name: "logs_fields", Description: "获取日志可用字段列表", Category: "logs", RiskLevel: 0,
		Params: []AIToolParam{
			{Name: "source", Type: "string", Required: false, Description: "日志源过滤"},
		},
		Handler: func(ctx context.Context, params map[string]any) (any, error) {
			fields := []string{"timestamp", "level", "message", "host", "service", "trace_id", "span_id", "source", "action", "operator", "target"}
			return map[string]any{"fields": fields, "total": len(fields)}, nil
		},
	})
}

// registerTracesTools registers the 5 tracing tools.
func registerTracesTools() {
	registerTool(&AITool{
		Name: "tracing_services", Description: "列出链路追踪服务列表", Category: "traces", RiskLevel: 0,
		Params: []AIToolParam{},
		Handler: func(ctx context.Context, params map[string]any) (any, error) {
			agents := store.ListAgents()
			services := make([]string, 0)
			seen := map[string]bool{}
			for _, a := range agents {
				svc := a.Hostname
				if svc != "" && !seen[svc] {
					services = append(services, svc)
					seen[svc] = true
				}
			}
			return map[string]any{"services": services, "total": len(services)}, nil
		},
	})
	registerTool(&AITool{
		Name: "tracing_traces", Description: "查询链路追踪 Trace 列表", Category: "traces", RiskLevel: 0,
		Params: []AIToolParam{
			{Name: "service", Type: "string", Required: false, Description: "服务名过滤"},
			{Name: "min_duration", Type: "string", Required: false, Description: "最小耗时过滤 (如 100ms, 1s)"},
			{Name: "limit", Type: "number", Required: false, Description: "返回条数"},
			{Name: "start", Type: "string", Required: false, Description: "开始时间"},
			{Name: "end", Type: "string", Required: false, Description: "结束时间"},
		},
		Handler: func(ctx context.Context, params map[string]any) (any, error) {
			service := logsToolParamStr(params, "service", "")
			return map[string]any{"service": service, "traces": []any{}, "total": 0, "source": "skywalking"}, nil
		},
	})
	registerTool(&AITool{
		Name: "tracing_spans", Description: "获取 Trace 的 Span 详情", Category: "traces", RiskLevel: 0,
		Params: []AIToolParam{
			{Name: "trace_id", Type: "string", Required: true, Description: "Trace ID"},
		},
		Handler: func(ctx context.Context, params map[string]any) (any, error) {
			traceID := fmt.Sprint(params["trace_id"])
			return map[string]any{"trace_id": traceID, "spans": []any{}, "total": 0, "source": "skywalking"}, nil
		},
	})
	registerTool(&AITool{
		Name: "tracing_topology", Description: "获取服务调用拓扑", Category: "traces", RiskLevel: 0,
		Params: []AIToolParam{
			{Name: "service", Type: "string", Required: false, Description: "中心服务名"},
			{Name: "depth", Type: "number", Required: false, Description: "拓扑深度"},
		},
		Handler: func(ctx context.Context, params map[string]any) (any, error) {
			graph := store.GetTopology()
			return map[string]any{"nodes": len(graph.Nodes), "edges": len(graph.Edges), "type": "service_topology"}, nil
		},
	})
	registerTool(&AITool{
		Name: "tracing_profiling", Description: "获取服务性能 Profiling 数据", Category: "traces", RiskLevel: 0,
		Params: []AIToolParam{
			{Name: "service", Type: "string", Required: true, Description: "服务名"},
			{Name: "type", Type: "string", Required: false, Description: "Profiling 类型 (cpu/memory/goroutine)"},
		},
		Handler: func(ctx context.Context, params map[string]any) (any, error) {
			return map[string]any{
				"service": params["service"],
				"type":    params["type"],
				"status":  "profiling data available via SkyWalking",
			}, nil
		},
	})
}

// logsToolParamInt safely extracts an int param with default.
func logsToolParamInt(params map[string]any, key string, def int) int {
	if v, ok := params[key]; ok {
		switch n := v.(type) {
		case float64:
			return int(n)
		case int:
			return n
		case json.Number:
			if i, err := n.Int64(); err == nil {
				return int(i)
			}
		}
	}
	return def
}

// logsToolParamStr safely extracts a string param with default.
func logsToolParamStr(params map[string]any, key, def string) string {
	if v, ok := params[key]; ok {
		s := strings.TrimSpace(fmt.Sprint(v))
		if s != "" && s != "<nil>" {
			return s
		}
	}
	return def
}

// logsToolParamTime parses a time param or returns a default.
func logsToolParamTime(params map[string]any, key string, def time.Time) time.Time {
	s := logsToolParamStr(params, key, "")
	if s == "" {
		return def
	}
	t, err := time.Parse(time.RFC3339, s)
	if err != nil {
		return def
	}
	return t
}
