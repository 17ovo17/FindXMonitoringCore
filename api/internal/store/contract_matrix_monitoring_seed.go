package store

import (
	"time"

	"ai-workbench-api/internal/model"
)

func EnsureMonitoringContractMatrixSeeded() error {
	now := time.Now()
	entries := monitoringContractMatrixSeedEntries()
	mu.Lock()
	defer mu.Unlock()
	for _, input := range entries {
		item, err := normalizeContractMatrixEntry(input, now)
		if err != nil {
			return err
		}
		if _, exists := contractMatrixEntries[item.ID]; exists {
			continue
		}
		contractMatrixEntries[item.ID] = &item
	}
	return nil
}

func monitoringContractMatrixSeedEntries() []model.ContractMatrixRegisterRequest {
	entries := []model.ContractMatrixRegisterRequest{}
	entries = append(entries, monitoringDatasourceContractSeeds()...)
	entries = append(entries, monitoringCatalogTemplateContractSeeds()...)
	entries = append(entries, monitoringQueryDashboardContractSeeds()...)
	entries = append(entries, monitoringAlertNotificationContractSeeds()...)
	return entries
}

func monitoringDatasourceContractSeeds() []model.ContractMatrixRegisterRequest {
	entries := []model.ContractMatrixRegisterRequest{
		monitoringReadyContractSeed(
			"FX-CONTRACT-N9E-DATASOURCE-BRIEF-LIST",
			"Datasource brief list",
			[]string{`D:\项目迁移文件\平台源码\fe-main\src\services\common.ts`, `D:\项目迁移文件\平台源码\fe-main\src\services\dashboardV2.ts`},
			"ListMonitorDatasources",
			"handler.monitoring_query.ListMonitorDatasources",
			"monitoring.PrometheusDatasourcesFromConfig",
			"/api/v1/monitor/datasources",
			[]string{
				"api/internal/handler/monitoring_query.go ListMonitorDatasources reads configured prometheus datasources",
				"api/internal/monitoring/datasources.go PrometheusDatasourcesFromConfig resolves configured prometheus sources",
				"api/routes_monitor.go GET /monitor/datasources is registered behind monitor datasource read permission",
			},
			monitoringContractMetadata("/integrations?section=datasources", "datasource_contract", "/monitor/datasources"),
		),
		monitoringContractSeed(
			"FX-CONTRACT-N9E-DATASOURCE-PROXY-BY-ID",
			"Datasource proxy aggregate",
			model.ContractStatusBlocked,
			[]string{`D:\项目迁移文件\平台源码\fe-main\src\services\dashboardV2.ts`, `D:\项目迁移文件\平台源码\fe-main\src\services\metricViews.ts`, `D:\项目迁移文件\平台源码\fe-main\src\services\warning.ts`},
			"Datasource proxy aggregate is represented by child contract gaps",
			monitoringContractScopeMetadata(
				"/query?section=metrics",
				"datasource_proxy_aggregate",
				"datasource-proxy-aggregate",
				"Child gaps own labels,label-values,metric-names,series,buildinfo,query,query_range,es_search endpoint refs",
			),
		),
		monitoringContractSeed(
			"FX-CONTRACT-N9E-DATASOURCE-TEST-CONNECTION",
			"Datasource connection test",
			model.ContractStatusMissingExecutor,
			[]string{`D:\项目迁移文件\平台源码\fe-main\src\services\dashboardV2.ts`, `D:\项目迁移文件\平台源码\fe-main\src\services\metricViews.ts`},
			"Datasource connection test executor contract is missing",
			monitoringContractMetadata("/integrations?section=datasources", "datasource_contract", "connection-test"),
		),
	}
	entries = append(entries, monitoringDatasourceProxyGapSeeds()...)
	return entries
}

func monitoringDatasourceProxyGapSeeds() []model.ContractMatrixRegisterRequest {
	return []model.ContractMatrixRegisterRequest{
		monitoringDatasourceProxyGapSeed("FX-CONTRACT-N9E-DATASOURCE-PROXY-LABELS", "Datasource proxy labels", dashboardV2AndMetricViewsSourceRefs(), "/query?section=metrics", "datasource_proxy_labels", "GET /api/{N9E_PATHNAME}/proxy/{datasourceValue}/api/v1/labels"),
		monitoringDatasourceProxyGapSeed("FX-CONTRACT-N9E-DATASOURCE-PROXY-LABEL-VALUES", "Datasource proxy label values", dashboardV2AndMetricViewsSourceRefs(), "/query?section=metrics", "datasource_proxy_label_values", "GET /api/{N9E_PATHNAME}/proxy/{datasourceValue}/api/v1/label/{label}/values"),
		monitoringDatasourceProxyGapSeed("FX-CONTRACT-N9E-DATASOURCE-PROXY-METRIC-NAMES", "Datasource proxy metric names", dashboardV2AndMetricViewsSourceRefs(), "/query?section=metrics", "datasource_proxy_metric_names", "GET /api/{N9E_PATHNAME}/proxy/{datasource_id}/api/v1/label/__name__/values"),
		monitoringDatasourceProxyGapSeed("FX-CONTRACT-N9E-DATASOURCE-PROXY-SERIES", "Datasource proxy series", dashboardV2SourceRefs(), "/query?section=metrics", "datasource_proxy_series", "GET /api/{N9E_PATHNAME}/proxy/{datasourceValue}/api/v1/series"),
		monitoringDatasourceProxyGapSeed("FX-CONTRACT-N9E-DATASOURCE-PROXY-BUILDINFO", "Datasource proxy build info", dashboardV2SourceRefs(), "/query?section=metrics", "datasource_proxy_buildinfo", "GET /api/{N9E_PATHNAME}/proxy/{datasourceValue}/api/v1/status/buildinfo"),
		monitoringDatasourceProxyGapSeed("FX-CONTRACT-N9E-DATASOURCE-PROXY-QUERY", "Datasource proxy instant query", dashboardV2SourceRefs(), "/query?section=metrics", "datasource_proxy_query", "GET /api/{N9E_PATHNAME}/proxy/{datasourceValue}/api/v1/query"),
		monitoringDatasourceProxyGapSeed("FX-CONTRACT-N9E-DATASOURCE-PROXY-QUERY-RANGE", "Datasource proxy range query", metricViewsSourceRefs(), "/query?section=metric-views", "datasource_proxy_query_range", "GET /api/{N9E_PATHNAME}/proxy/{datasourceValue}/api/v1/query_range"),
		monitoringDatasourceProxyGapSeed("FX-CONTRACT-N9E-DATASOURCE-PROXY-ES-SEARCH", "Datasource proxy ES search", dashboardV2SourceRefs(), "/query?section=metrics", "datasource_proxy_es_search", "POST /api/{N9E_PATHNAME}/proxy/{datasourceValue}/{index}/_search"),
	}
}

func monitoringDatasourceProxyGapSeed(id, capability string, sourceRefs []string, findxRoute, gapType, upstreamRef string) model.ContractMatrixRegisterRequest {
	return monitoringContractSeed(
		id,
		capability,
		model.ContractStatusMissingDatasource,
		sourceRefs,
		"Datasource proxy datasource contract is missing",
		monitoringContractMetadata(findxRoute, gapType, upstreamRef),
	)
}

func monitoringCatalogTemplateContractSeeds() []model.ContractMatrixRegisterRequest {
	return []model.ContractMatrixRegisterRequest{
		monitoringContractSeed(
			"FX-CONTRACT-N9E-SYSTEM-INTEGRATION-CATALOG",
			"System integration catalog",
			model.ContractStatusMissingBackend,
			[]string{`D:\项目迁移文件\平台源码\fe-main\src\services\manage.ts`, `D:\项目迁移文件\平台源码\fe-main\src\services\common.ts`},
			"System integration catalog backend contract is missing",
			monitoringContractMetadata("/integrations?section=systems", "integration_catalog", "system-integration-catalog"),
		),
		monitoringContractSeed(
			"FX-CONTRACT-N9E-TEMPLATE-CENTER-DASHBOARD-IMPORT",
			"Template center dashboard import",
			model.ContractStatusMissingBackend,
			[]string{`D:\项目迁移文件\平台源码\fe-main\src\services\dashboardV2.ts`},
			"Dashboard template import contract is missing",
			monitoringContractScopeMetadata(
				"/integrations?section=templates",
				"dashboard_template_import_aggregate",
				"dashboard-template-import-aggregate",
				"Template center import flow depends on dry-run, conflict handling, rollback, migrate, and follow-up child contracts",
			),
		),
	}
}

func monitoringQueryDashboardContractSeeds() []model.ContractMatrixRegisterRequest {
	entries := []model.ContractMatrixRegisterRequest{
		monitoringReadyContractSeed(
			"FX-CONTRACT-FINDX-PROMETHEUS-SINGLE-QUERY",
			"FindX Prometheus single query adapter",
			[]string{
				`D:\项目迁移文件\平台源码\fe-main\src\services\dashboardV2.ts`,
				`D:\项目迁移文件\平台源码\fe-main\src\services\metricViews.ts`,
				`D:\项目迁移文件\平台源码\fe-main\src\services\metric.ts`,
			},
			"MonitorQuery MonitorQueryRange ListMonitorLabels ListMonitorLabelValues",
			"handler.monitoring_query Prometheus single query adapter",
			"monitoring.PrometheusGateway",
			"/api/v1/monitor/query /api/v1/monitor/query-range /api/v1/monitor/labels /api/v1/monitor/label-values",
			[]string{
				"api/routes_monitor.go registers /monitor/query, /monitor/query-range, /monitor/labels, /monitor/label-values",
				"api/internal/handler/monitoring_query.go validates single PromQL query/range and label requests",
				"api/internal/handler/monitoring_query_prometheus.go proxies Prometheus query, query_range, labels, and label values",
				"api/internal/handler/monitoring_query_test.go covers single query, range query, labels validation, datasource miss, and redaction",
			},
			monitoringContractScopeMetadata(
				"/query?section=metrics",
				"findx_prometheus_single_query_adapter",
				"/monitor/query,/monitor/query-range,/monitor/labels,/monitor/label-values",
				"FindX single Prometheus query, range, labels, and label-values only; not batch query or metric views CRUD",
			),
		),
		monitoringContractSeed(
			"FX-CONTRACT-N9E-METRIC-VIEWS-CRUD",
			"Metric views CRUD aggregate",
			model.ContractStatusBlocked,
			[]string{`D:\项目迁移文件\平台源码\fe-main\src\services\metricViews.ts`},
			"Metric views CRUD aggregate is represented by child contract gaps",
			monitoringContractScopeMetadata(
				"/query?section=metric-views",
				"metric_view_contract",
				"metric-views-crud-aggregate",
				"Child gaps carry endpoint refs: FX-CONTRACT-N9E-METRIC-VIEWS-LIST, FX-CONTRACT-N9E-METRIC-VIEWS-CREATE, FX-CONTRACT-N9E-METRIC-VIEWS-UPDATE, FX-CONTRACT-N9E-METRIC-VIEWS-DELETE",
			),
		),
		monitoringContractSeed(
			"FX-CONTRACT-N9E-METRIC-QUERY-BATCH",
			"Metric query aggregate",
			model.ContractStatusBlocked,
			[]string{`D:\项目迁移文件\平台源码\fe-main\src\services\metric.ts`, `D:\项目迁移文件\平台源码\fe-main\src\services\metricViews.ts`, `D:\项目迁移文件\平台源码\fe-main\src\services\warning.ts`},
			"Metric query aggregate is represented by child contract gaps",
			monitoringContractScopeMetadata(
				"/query?section=metrics",
				"metric_query_contract",
				"metric-query-aggregate",
				"Child gaps carry endpoint refs: FX-CONTRACT-N9E-QUERY-RANGE-BATCH, FX-CONTRACT-N9E-QUERY-INSTANT-BATCH, FX-CONTRACT-N9E-PLUS-QUERY-BATCH, FX-CONTRACT-N9E-TAG-PAIRS, FX-CONTRACT-N9E-TAG-METRICS, FX-CONTRACT-N9E-QUERY-DATA, FX-CONTRACT-N9E-QUERY-BENCH, FX-CONTRACT-N9E-PROMETHEUS-COMPAT-API",
			),
		),
	}
	entries = append(entries, monitoringMetricViewsGapSeeds()...)
	entries = append(entries, monitoringMetricQueryGapSeeds()...)
	entries = append(entries,
		monitoringContractSeed(
			"FX-CONTRACT-N9E-DASHBOARD-ANNOTATIONS",
			"Dashboard annotations aggregate",
			model.ContractStatusBlocked,
			[]string{`D:\项目迁移文件\平台源码\fe-main\src\services\dashboardV2.ts`},
			"Dashboard annotations aggregate is represented by child contract gaps",
			monitoringContractScopeMetadata(
				"/dashboards?section=list",
				"dashboard_annotation_aggregate",
				"dashboard-annotations-aggregate",
				"Child gaps own list, create, update, and delete annotation endpoint refs",
			),
		),
	)
	entries = append(entries, monitoringDashboardActionGapSeeds()...)
	return entries
}

func monitoringDashboardActionGapSeeds() []model.ContractMatrixRegisterRequest {
	return []model.ContractMatrixRegisterRequest{
		monitoringDashboardActionGapSeed("FX-CONTRACT-N9E-DASHBOARD-ANNOTATIONS-LIST", "Dashboard annotations list", "/dashboards?section=list", "dashboard_annotation_list", "GET /api/n9e/dashboard-annotations"),
		monitoringDashboardActionGapSeed("FX-CONTRACT-N9E-DASHBOARD-ANNOTATIONS-CREATE", "Dashboard annotations create", "/dashboards?section=list", "dashboard_annotation_create", "POST /api/n9e/dashboard-annotations"),
		monitoringDashboardActionGapSeed("FX-CONTRACT-N9E-DASHBOARD-ANNOTATIONS-UPDATE", "Dashboard annotations update", "/dashboards?section=list", "dashboard_annotation_update", "PUT /api/n9e/dashboard-annotation/{id}"),
		monitoringDashboardActionGapSeed("FX-CONTRACT-N9E-DASHBOARD-ANNOTATIONS-DELETE", "Dashboard annotations delete", "/dashboards?section=list", "dashboard_annotation_delete", "DELETE /api/n9e/dashboard-annotation/{id}"),
		monitoringDashboardActionGapSeed("FX-CONTRACT-N9E-DASHBOARD-PUBLIC-LIST", "Dashboard public list", "/dashboards?section=list", "dashboard_public_list", "GET /api/n9e/busi-groups/public-boards"),
		monitoringDashboardActionGapSeed("FX-CONTRACT-N9E-DASHBOARD-PUBLIC-UPDATE", "Dashboard public update", "/dashboards?section=list", "dashboard_public_update", "PUT /api/n9e/board/{id}/public"),
		monitoringDashboardActionGapSeed("FX-CONTRACT-N9E-DASHBOARD-EXPORT", "Dashboard export", "/dashboards?section=list", "dashboard_export", "POST /api/n9e/busi-group/{busiId}/dashboards/export"),
		monitoringDashboardActionGapSeed("FX-CONTRACT-N9E-DASHBOARD-MIGRATE", "Dashboard migrate", "/dashboards?section=list", "dashboard_migrate", "PUT /api/n9e/dashboard/{id}/migrate"),
	}
}

func monitoringDashboardActionGapSeed(id, capability, findxRoute, gapType, upstreamRef string) model.ContractMatrixRegisterRequest {
	return monitoringContractSeed(
		id,
		capability,
		model.ContractStatusMissingBackend,
		dashboardV2SourceRefs(),
		"Dashboard action backend contract is missing",
		monitoringContractMetadata(findxRoute, gapType, upstreamRef),
	)
}

func monitoringMetricViewsGapSeeds() []model.ContractMatrixRegisterRequest {
	return []model.ContractMatrixRegisterRequest{
		monitoringMetricQueryGapSeed("FX-CONTRACT-N9E-METRIC-VIEWS-LIST", "Metric views list", model.ContractStatusMissingBackend, metricViewsSourceRefs(), "Metric views list backend contract is missing", "/query?section=metric-views", "metric_view_list", "GET /api/n9e/metric-views"),
		monitoringMetricQueryGapSeed("FX-CONTRACT-N9E-METRIC-VIEWS-CREATE", "Metric views create", model.ContractStatusMissingBackend, metricViewsSourceRefs(), "Metric views create backend contract is missing", "/query?section=metric-views", "metric_view_create", "POST /api/n9e/metric-views"),
		monitoringMetricQueryGapSeed("FX-CONTRACT-N9E-METRIC-VIEWS-UPDATE", "Metric views update", model.ContractStatusMissingBackend, metricViewsSourceRefs(), "Metric views update backend contract is missing", "/query?section=metric-views", "metric_view_update", "PUT /api/n9e/metric-views"),
		monitoringMetricQueryGapSeed("FX-CONTRACT-N9E-METRIC-VIEWS-DELETE", "Metric views delete", model.ContractStatusMissingBackend, metricViewsSourceRefs(), "Metric views delete backend contract is missing", "/query?section=metric-views", "metric_view_delete", "DELETE /api/n9e/metric-views"),
	}
}

func monitoringMetricQueryGapSeeds() []model.ContractMatrixRegisterRequest {
	return []model.ContractMatrixRegisterRequest{
		monitoringMetricQueryGapSeed("FX-CONTRACT-N9E-QUERY-RANGE-BATCH", "Metric range batch query", model.ContractStatusMissingDatasource, dashboardV2SourceRefs(), "Metric range batch datasource contract is missing", "/query?section=metrics", "metric_query_range_batch", "/api/{N9E_PATHNAME}/query-range-batch"),
		monitoringMetricQueryGapSeed("FX-CONTRACT-N9E-QUERY-INSTANT-BATCH", "Metric instant batch query", model.ContractStatusMissingDatasource, dashboardV2SourceRefs(), "Metric instant batch datasource contract is missing", "/query?section=metrics", "metric_query_instant_batch", "/api/{N9E_PATHNAME}/query-instant-batch"),
		monitoringMetricQueryGapSeed("FX-CONTRACT-N9E-PLUS-QUERY-BATCH", "Metric plus query batch", model.ContractStatusMissingDatasource, dashboardV2SourceRefs(), "Metric plus query datasource contract is missing", "/query?section=metrics", "metric_query_plus_batch", "/api/n9e-plus/query-batch"),
		monitoringMetricQueryGapSeed("FX-CONTRACT-N9E-TAG-PAIRS", "Metric tag pairs", model.ContractStatusMissingDatasource, metricSourceRefs(), "Metric tag pairs datasource contract is missing", "/query?section=metrics", "metric_tag_pairs", "/api/n9e/tag-pairs"),
		monitoringMetricQueryGapSeed("FX-CONTRACT-N9E-TAG-METRICS", "Metric tag metrics", model.ContractStatusMissingDatasource, metricSourceRefs(), "Metric tag metrics datasource contract is missing", "/query?section=metrics", "metric_tag_metrics", "/api/n9e/tag-metrics"),
		monitoringMetricQueryGapSeed("FX-CONTRACT-N9E-QUERY-DATA", "Metric raw data query", model.ContractStatusMissingDatasource, metricSourceRefs(), "Metric raw data datasource contract is missing", "/query?section=metrics", "metric_query_data", "/api/n9e/query"),
		monitoringMetricQueryGapSeed("FX-CONTRACT-N9E-QUERY-BENCH", "Metric query bench", model.ContractStatusMissingBackend, metricSourceRefs(), "Metric query bench backend contract is missing", "/query?section=metrics", "metric_query_bench", "/api/n9e/query-bench"),
		monitoringMetricQueryGapSeed("FX-CONTRACT-N9E-PROMETHEUS-COMPAT-API", "Prometheus compatibility API", model.ContractStatusMissingDatasource, append(metricSourceRefs(), metricViewsSourceRefs()...), "Prometheus compatibility datasource contract is missing", "/query?section=metrics", "prometheus_compat_api", "/api/n9e/prometheus/api/v1/{path}"),
		monitoringMetricQueryGapSeed("FX-CONTRACT-N9E-SHARE-CHARTS", "Metric share charts", model.ContractStatusMissingBackend, append(metricSourceRefs(), metricViewsSourceRefs()...), "Metric share charts backend contract is missing", "/query?section=metrics", "metric_share_charts", "/api/n9e/share-charts"),
		monitoringMetricQueryGapSeed("FX-CONTRACT-N9E-METRICS-DESC", "Metric description", model.ContractStatusMissingBackend, metricViewsSourceRefs(), "Metric description backend contract is missing", "/query?section=metric-views", "metric_description", "/api/n9e/metrics/desc"),
	}
}

func monitoringMetricQueryGapSeed(id, capability, status string, sourceRefs []string, blockedReason, findxRoute, gapType, upstreamRef string) model.ContractMatrixRegisterRequest {
	return monitoringContractSeed(id, capability, status, sourceRefs, blockedReason, monitoringContractMetadata(findxRoute, gapType, upstreamRef))
}

func dashboardV2SourceRefs() []string {
	return []string{`D:\项目迁移文件\平台源码\fe-main\src\services\dashboardV2.ts`}
}

func metricSourceRefs() []string {
	return []string{`D:\项目迁移文件\平台源码\fe-main\src\services\metric.ts`}
}

func metricViewsSourceRefs() []string {
	return []string{`D:\项目迁移文件\平台源码\fe-main\src\services\metricViews.ts`}
}

func dashboardV2AndMetricViewsSourceRefs() []string {
	return append(dashboardV2SourceRefs(), metricViewsSourceRefs()...)
}

func monitoringAlertNotificationContractSeeds() []model.ContractMatrixRegisterRequest {
	return []model.ContractMatrixRegisterRequest{
		monitoringContractSeed(
			"FX-CONTRACT-N9E-ALERT-RULE-GROUPS",
			"Alert rule groups",
			model.ContractStatusMissingBackend,
			[]string{`D:\项目迁移文件\平台源码\fe-main\src\services\warning.ts`},
			"Alert rule group backend contract is missing",
			monitoringContractMetadata("/alerts?section=rules", "alert_rule_group", "/api/n9e/alert-rule-groups"),
		),
		monitoringContractSeed(
			"FX-CONTRACT-N9E-NOTIFICATION-FINDX-ADAPTER",
			"Notification FindX adapter",
			model.ContractStatusMissingExecutor,
			[]string{`D:\项目迁移文件\平台源码\fe-main\src\services\manage.ts`, `D:\项目迁移文件\平台源码\fe-main\src\services\subscribe.ts`, `D:\项目迁移文件\平台源码\fe-main\src\services\shield.ts`},
			"Notification adapter executor contract is missing",
			monitoringContractAdapterMetadata("/notifications?section=rules", "findx_notification_adapter", "/api/n9e/notify-channels"),
		),
		monitoringContractSeed(
			"FX-CONTRACT-N9E-BUSI-GROUP-RESOURCE-GROUP-MAP",
			"Business group resource group map",
			model.ContractStatusMissingBackend,
			[]string{`D:\项目迁移文件\平台源码\fe-main\src\services\manage.ts`, `D:\项目迁移文件\平台源码\fe-main\src\services\resource.ts`, `D:\项目迁移文件\平台源码\fe-main\src\services\targets.ts`},
			"Business group to resource group mapping contract is missing",
			monitoringContractMetadata("/assets?section=resource-groups", "resource_group_mapping", "/api/n9e/busi-groups"),
		),
	}
}

func monitoringContractSeed(id, capability, status string, sourceRefs []string, blockedReason string, metadata map[string]string) model.ContractMatrixRegisterRequest {
	return model.ContractMatrixRegisterRequest{
		ID:            id,
		Capability:    capability,
		Domain:        "monitoring",
		Status:        status,
		SourceRefs:    sourceRefs,
		BlockedReason: blockedReason,
		Metadata:      metadata,
	}
}

func monitoringReadyContractSeed(id, capability string, sourceRefs []string, handler, backend, datasource, executor string, evidenceRefs []string, metadata map[string]string) model.ContractMatrixRegisterRequest {
	return model.ContractMatrixRegisterRequest{
		ID:           id,
		Capability:   capability,
		Domain:       "monitoring",
		Status:       model.ContractStatusReady,
		Handler:      handler,
		Backend:      backend,
		Datasource:   datasource,
		Executor:     executor,
		SourceRefs:   sourceRefs,
		EvidenceRefs: evidenceRefs,
		SafeToRetry:  true,
		Metadata:     metadata,
	}
}

func monitoringContractMetadata(findxRoute, gapType, upstreamRef string) map[string]string {
	return map[string]string{
		"findx_route":  findxRoute,
		"gap_type":     gapType,
		"upstream_ref": upstreamRef,
	}
}

func monitoringContractScopeMetadata(findxRoute, gapType, upstreamRef, upstreamScope string) map[string]string {
	metadata := monitoringContractMetadata(findxRoute, gapType, upstreamRef)
	metadata["upstream_scope"] = upstreamScope
	return metadata
}

func monitoringContractAdapterMetadata(findxRoute, adapter, upstreamRef string) map[string]string {
	return map[string]string{
		"findx_route":   findxRoute,
		"findx_adapter": adapter,
		"gap_type":      "adapter_contract",
		"upstream_ref":  upstreamRef,
	}
}
