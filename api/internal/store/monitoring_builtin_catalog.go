package store

import (
	"encoding/json"
	"sort"

	"ai-workbench-api/internal/model"
)

func builtinComponentCatalog() []model.MonitoringBuiltinComponent {
	return []model.MonitoringBuiltinComponent{
		{
			ID:        "findx-monitor-core",
			Ident:     "findx-monitor-core",
			Name:      "FindX Monitor Core",
			Readme:    "Built-in monitoring component catalog for dashboard templates, collection guidance, metrics, alert examples, and import records.",
			Tags:      []string{"monitor", "template-center"},
			UpdatedBy: builtinUpdatedBy,
		},
	}
}

func builtinPayloadCatalog() []model.MonitoringBuiltinPayload {
	payloads := dashboardTemplatePayloads()
	payloads = append(payloads,
		staticReadmePayload(),
		staticInstructionsPayload(),
		staticCollectPayload(),
		staticMetricPayload(),
		staticAlertPayload(),
		staticRecordPayload(),
	)
	sort.Slice(payloads, func(i, j int) bool {
		if payloads[i].Type == payloads[j].Type {
			return payloads[i].ID < payloads[j].ID
		}
		return payloads[i].Type < payloads[j].Type
	})
	return payloads
}

func dashboardTemplatePayloads() []model.MonitoringBuiltinPayload {
	templates := ListMonitorDashboardTemplates()
	out := make([]model.MonitoringBuiltinPayload, 0, len(templates))
	for _, tpl := range templates {
		content := mustMonitoringBuiltinJSON(map[string]any{
			"source":    "store.ListMonitorDashboardTemplates",
			"template":  tpl.ID,
			"variables": json.RawMessage(tpl.Variables),
			"panels":    json.RawMessage(tpl.Panels),
		})
		out = append(out, model.MonitoringBuiltinPayload{
			ID:          "dashboard:" + tpl.ID,
			ComponentID: "findx-monitor-core",
			Type:        "dashboard",
			Name:        tpl.Title,
			Title:       tpl.Title,
			Tags:        append([]string{}, tpl.Tags...),
			Description: tpl.Description,
			Note:        "Read-only dashboard template preview. Import remains handled by dashboard-template import API.",
			UpdatedBy:   builtinUpdatedBy,
			Content:     content,
		})
	}
	return out
}

func staticReadmePayload() model.MonitoringBuiltinPayload {
	return model.MonitoringBuiltinPayload{
		ID:          "readme:findx-monitor-core",
		ComponentID: "findx-monitor-core",
		Type:        "readme",
		Name:        "FindX Monitor Core README",
		Title:       "FindX Monitor Core README",
		Tags:        []string{"readme", "monitor"},
		Description: "Read-only introduction for the built-in monitoring catalog.",
		Note:        "This endpoint only exposes preview content and does not create, update, or delete templates.",
		UpdatedBy:   builtinUpdatedBy,
		Content: mustMonitoringBuiltinJSON(map[string]any{
			"format": "markdown",
			"body":   "# FindX Monitor Core\n\nThis catalog exposes built-in dashboard previews and contract-visible examples for collection, metrics, alerts, and import records.",
		}),
	}
}

func staticInstructionsPayload() model.MonitoringBuiltinPayload {
	return model.MonitoringBuiltinPayload{
		ID:          "instructions:findx-monitor-core",
		ComponentID: "findx-monitor-core",
		Type:        "instructions",
		Name:        "Read-only integration instructions",
		Title:       "Read-only integration instructions",
		Tags:        []string{"instructions", "contract"},
		Description: "Frontend integration notes for switching from fallback data to backend contract data.",
		Note:        "Write operations from upstream source are intentionally not implemented in this slice.",
		UpdatedBy:   builtinUpdatedBy,
		Content: mustMonitoringBuiltinJSON(map[string]any{
			"mode": "read_only",
			"supported_endpoints": []string{
				"GET /api/v1/monitor/builtin-components",
				"GET /api/v1/monitor/builtin-payloads/cates",
				"GET /api/v1/monitor/builtin-payloads",
				"GET /api/v1/monitor/builtin-payload/:id",
			},
			"blocked_writes": []string{"POST", "PUT", "DELETE"},
		}),
	}
}

func staticCollectPayload() model.MonitoringBuiltinPayload {
	return model.MonitoringBuiltinPayload{
		ID:          "collect:findx-agent-plugin",
		ComponentID: "findx-monitor-core",
		Type:        "collect",
		Name:        "FindX Agent collector plugin contract",
		Title:       "FindX Agent collector plugin contract",
		Tags:        []string{"collect", "agent", "collector-plugin"},
		Description: "Preview of collection inputs expected by FindX Agent collector plugin integration.",
		Note:        "Executable rollout is blocked until lifecycle contract is selected by Agent Center.",
		UpdatedBy:   builtinUpdatedBy,
		Content: blockedMonitoringBuiltinContent("collect_execution_not_available", map[string]any{
			"collector": "findx_agent_collector_plugin",
			"inputs":    []string{"cpu", "mem", "disk", "net", "prometheus"},
			"preview":   "Collection content is visible for template center only.",
		}),
	}
}

func staticMetricPayload() model.MonitoringBuiltinPayload {
	return model.MonitoringBuiltinPayload{
		ID:          "metric:host-core",
		ComponentID: "findx-monitor-core",
		Type:        "metric",
		Name:        "Host core metric examples",
		Title:       "Host core metric examples",
		Tags:        []string{"metric", "host"},
		Description: "Metric names used by the built-in host dashboard templates.",
		Note:        "Metrics are examples for preview and query mapping, not a write contract.",
		UpdatedBy:   builtinUpdatedBy,
		Content: mustMonitoringBuiltinJSON(map[string]any{
			"metrics": []map[string]string{
				{"name": "node_cpu_seconds_total", "kind": "counter", "unit": "seconds"},
				{"name": "node_memory_MemAvailable_bytes", "kind": "gauge", "unit": "bytes"},
				{"name": "node_filesystem_avail_bytes", "kind": "gauge", "unit": "bytes"},
				{"name": "node_network_receive_bytes_total", "kind": "counter", "unit": "bytes"},
			},
		}),
	}
}

func staticAlertPayload() model.MonitoringBuiltinPayload {
	return model.MonitoringBuiltinPayload{
		ID:          "alert:host-availability-preview",
		ComponentID: "findx-monitor-core",
		Type:        "alert",
		Name:        "Host availability alert preview",
		Title:       "Host availability alert preview",
		Tags:        []string{"alert", "host"},
		Description: "Read-only alert rule example for template center preview.",
		Note:        "Rule creation is intentionally not exposed by builtin-payload endpoints.",
		UpdatedBy:   builtinUpdatedBy,
		Content: blockedMonitoringBuiltinContent("alert_write_not_available", map[string]any{
			"expr":     "up{job=~\"node|windows\"} == 0",
			"for":      "5m",
			"severity": "warning",
		}),
	}
}

func staticRecordPayload() model.MonitoringBuiltinPayload {
	return model.MonitoringBuiltinPayload{
		ID:          "record:dashboard-import-preview",
		ComponentID: "findx-monitor-core",
		Type:        "record",
		Name:        "Dashboard import record preview",
		Title:       "Dashboard import record preview",
		Tags:        []string{"record", "dashboard"},
		Description: "Example record shape for displaying template import history when the audit contract is available.",
		Note:        "No import record is persisted by this read-only endpoint.",
		UpdatedBy:   builtinUpdatedBy,
		Content: blockedMonitoringBuiltinContent("record_storage_not_available", map[string]any{
			"record_fields": []string{"template_id", "dashboard_id", "actor", "imported_at"},
		}),
	}
}

func blockedMonitoringBuiltinContent(reason string, preview map[string]any) json.RawMessage {
	return mustMonitoringBuiltinJSON(map[string]any{
		"status":  "PENDING",
		"reason":  reason,
		"preview": preview,
	})
}

func mustMonitoringBuiltinJSON(value any) json.RawMessage {
	data, err := json.Marshal(value)
	if err != nil {
		return json.RawMessage(`{"status":"PENDING","reason":"json_marshal_failed"}`)
	}
	return data
}
