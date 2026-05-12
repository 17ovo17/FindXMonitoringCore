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
	entries := []model.ContractMatrixRegisterRequest{
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
			model.ContractStatusBlocked,
			[]string{`D:\项目迁移文件\平台源码\fe-main\src\services\dashboardV2.ts`},
			"Template center dashboard import aggregate is represented by child contract gaps",
			monitoringContractScopeMetadata(
				"/integrations?section=templates",
				"dashboard_template_import_aggregate",
				"dashboard-template-import-aggregate",
				"Child gaps own builtin detail, batch result, conflict rollback, and docs drawer; excludes CRUD/public/export/migrate/annotations",
			),
		),
	}
	entries = append(entries, monitoringTemplateCenterDashboardImportGapSeeds()...)
	return entries
}

func monitoringTemplateCenterDashboardImportGapSeeds() []model.ContractMatrixRegisterRequest {
	return []model.ContractMatrixRegisterRequest{
		monitoringTemplateCenterDashboardImportGapSeed(
			"FX-CONTRACT-N9E-TEMPLATE-CENTER-BUILTIN-BOARD-DETAIL",
			"Template center builtin dashboard detail",
			[]string{
				`D:\项目迁移文件\平台源码\fe-main\src\services\dashboardV2.ts`,
				`D:\项目迁移文件\平台源码\fe-main\src\pages\builtInComponents\Dashboards\Detail.tsx`,
				`D:\项目迁移文件\平台源码\fe-main\src\pages\dashboard\Detail\Detail.tsx`,
			},
			"template_center_builtin_board_detail",
			"POST /api/n9e/builtin-boards-detail",
			"Template center builtin dashboard detail backend contract is missing",
		),
		monitoringTemplateCenterDashboardImportGapSeed(
			"FX-CONTRACT-N9E-TEMPLATE-CENTER-DASHBOARD-IMPORT-BATCH-RESULT",
			"Template center dashboard import batch result",
			[]string{
				`D:\项目迁移文件\平台源码\fe-main\src\pages\builtInComponents\Dashboards\Import.tsx`,
				`D:\项目迁移文件\平台源码\fe-main\src\pages\builtInComponents\Dashboards\services.ts`,
			},
			"dashboard_template_import_batch_result",
			"template-import-batch-result",
			"Dashboard template import batch result contract is missing",
		),
		monitoringTemplateCenterDashboardImportGapSeed(
			"FX-CONTRACT-N9E-TEMPLATE-CENTER-DASHBOARD-IMPORT-CONFLICT-ROLLBACK",
			"Template center dashboard import conflict rollback",
			[]string{
				`D:\项目迁移文件\平台源码\fe-main\src\pages\builtInComponents\Dashboards\Import.tsx`,
				`D:\项目迁移文件\平台源码\fe-main\src\services\dashboardV2.ts`,
			},
			"dashboard_template_import_conflict_rollback",
			"template-import-conflict-rollback",
			"Dashboard template import conflict and rollback contract is missing",
		),
		monitoringTemplateCenterDashboardImportGapSeed(
			"FX-CONTRACT-N9E-TEMPLATE-CENTER-DOCUMENT-DRAWER",
			"Template center document drawer",
			[]string{
				`D:\项目迁移文件\平台源码\fe-main\src\components\DocumentDrawer\index.tsx`,
				`D:\项目迁移文件\平台源码\fe-main\src\components\DocumentDrawer\Document.tsx`,
			},
			"template_center_document_drawer",
			"/n9e-docs/{path}/{language}.md",
			"Template center document drawer content contract is missing",
		),
	}
}

func monitoringTemplateCenterDashboardImportGapSeed(id, capability string, sourceRefs []string, gapType, upstreamRef, blockedReason string) model.ContractMatrixRegisterRequest {
	return monitoringContractSeed(
		id,
		capability,
		model.ContractStatusMissingBackend,
		sourceRefs,
		blockedReason,
		monitoringContractMetadata("/integrations?section=templates", gapType, upstreamRef),
	)
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
			"FX-CONTRACT-N9E-DASHBOARD-CRUD",
			"Dashboard CRUD aggregate",
			model.ContractStatusBlocked,
			dashboardV2SourceRefs(),
			"Dashboard CRUD aggregate is represented by child contract gaps",
			monitoringContractScopeMetadata(
				"/dashboards?section=list",
				"dashboard_crud_aggregate",
				"dashboard-crud-aggregate",
				"dashboard CRUD child gaps own list, create, detail, update, clone, delete, pure, and names; excludes public/export/migrate/annotations",
			),
		),
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
	entries := []model.ContractMatrixRegisterRequest{
		monitoringDashboardCRUDGapSeed("FX-CONTRACT-N9E-DASHBOARD-LIST-BY-BUSI-GROUP", "Dashboard list by business group", "dashboard_list_by_busi_group", "GET /api/n9e/busi-group/{id}/boards"),
		monitoringDashboardCRUDGapSeed("FX-CONTRACT-N9E-DASHBOARD-LIST-BY-BUSI-GROUPS", "Dashboard list by business groups", "dashboard_list_by_busi_groups", "GET /api/n9e/busi-groups/boards"),
		monitoringDashboardCRUDGapSeed("FX-CONTRACT-N9E-DASHBOARD-CREATE", "Dashboard create", "dashboard_create", "POST /api/n9e/busi-group/{id}/boards"),
		monitoringDashboardCRUDGapSeed("FX-CONTRACT-N9E-DASHBOARD-DETAIL", "Dashboard detail", "dashboard_detail", "GET /api/n9e/board/{id}"),
		monitoringDashboardCRUDGapSeed("FX-CONTRACT-N9E-DASHBOARD-UPDATE-METADATA", "Dashboard metadata update", "dashboard_update_metadata", "PUT /api/n9e/board/{id}"),
		monitoringDashboardCRUDGapSeed("FX-CONTRACT-N9E-DASHBOARD-UPDATE-CONFIGS", "Dashboard configs update", "dashboard_update_configs", "PUT /api/n9e/board/{id}/configs"),
		monitoringDashboardCRUDGapSeed("FX-CONTRACT-N9E-DASHBOARD-CLONE", "Dashboard clone", "dashboard_clone", "POST /api/n9e/busi-group/{busiId}/board/{id}/clone"),
		monitoringDashboardCRUDGapSeed("FX-CONTRACT-N9E-DASHBOARD-CLONES", "Dashboard batch clone", "dashboard_clones", "POST /api/n9e/busi-groups/boards/clones"),
		monitoringDashboardCRUDGapSeed("FX-CONTRACT-N9E-DASHBOARD-DELETE", "Dashboard delete", "dashboard_delete", "DELETE /api/n9e/boards"),
		monitoringDashboardCRUDGapSeed("FX-CONTRACT-N9E-DASHBOARD-PURE-DETAIL", "Dashboard pure detail", "dashboard_pure_detail", "GET /api/n9e/board/{id}/pure"),
		monitoringDashboardCRUDGapSeed("FX-CONTRACT-N9E-DASHBOARD-NAMES", "Dashboard names lookup", "dashboard_names", "GET /api/n9e/boards?bids={ids}"),
		monitoringDashboardActionGapSeed("FX-CONTRACT-N9E-DASHBOARD-ANNOTATIONS-LIST", "Dashboard annotations list", "/dashboards?section=list", "dashboard_annotation_list", "GET /api/n9e/dashboard-annotations"),
		monitoringDashboardActionGapSeed("FX-CONTRACT-N9E-DASHBOARD-ANNOTATIONS-CREATE", "Dashboard annotations create", "/dashboards?section=list", "dashboard_annotation_create", "POST /api/n9e/dashboard-annotations"),
		monitoringDashboardActionGapSeed("FX-CONTRACT-N9E-DASHBOARD-ANNOTATIONS-UPDATE", "Dashboard annotations update", "/dashboards?section=list", "dashboard_annotation_update", "PUT /api/n9e/dashboard-annotation/{id}"),
		monitoringDashboardActionGapSeed("FX-CONTRACT-N9E-DASHBOARD-ANNOTATIONS-DELETE", "Dashboard annotations delete", "/dashboards?section=list", "dashboard_annotation_delete", "DELETE /api/n9e/dashboard-annotation/{id}"),
		monitoringDashboardActionGapSeed("FX-CONTRACT-N9E-DASHBOARD-PUBLIC-LIST", "Dashboard public list", "/dashboards?section=list", "dashboard_public_list", "GET /api/n9e/busi-groups/public-boards"),
		monitoringDashboardActionGapSeed("FX-CONTRACT-N9E-DASHBOARD-PUBLIC-UPDATE", "Dashboard public update", "/dashboards?section=list", "dashboard_public_update", "PUT /api/n9e/board/{id}/public"),
		monitoringDashboardActionGapSeed("FX-CONTRACT-N9E-DASHBOARD-EXPORT", "Dashboard export", "/dashboards?section=list", "dashboard_export", "POST /api/n9e/busi-group/{busiId}/dashboards/export"),
		monitoringDashboardActionGapSeed("FX-CONTRACT-N9E-DASHBOARD-MIGRATE", "Dashboard migrate", "/dashboards?section=list", "dashboard_migrate", "PUT /api/n9e/dashboard/{id}/migrate"),
	}
	return entries
}

func monitoringDashboardCRUDGapSeed(id, capability, gapType, upstreamRef string) model.ContractMatrixRegisterRequest {
	return monitoringDashboardActionGapSeed(id, capability, "/dashboards?section=list", gapType, upstreamRef)
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
	entries := []model.ContractMatrixRegisterRequest{
		monitoringContractSeed(
			"FX-CONTRACT-N9E-ALERT-RULE-GROUPS",
			"Alert rule groups",
			model.ContractStatusBlocked,
			warningSourceRefs(),
			"Alert rule groups aggregate is represented by child contract gaps",
			monitoringContractScopeMetadata(
				"/alerts?section=rules",
				"alert_rule_group_aggregate",
				"alert-rule-groups-aggregate",
				"Child gaps own list/detail/create/update/delete/favorite/group rules; excludes alert rule CRUD/import/export/clone/status, alert mute/shield, alert subscribe, notification rules/templates/channels/contacts",
			),
		),
		monitoringContractSeed(
			"FX-CONTRACT-N9E-ALERT-RULE-LIFECYCLE",
			"Alert rule lifecycle aggregate",
			model.ContractStatusBlocked,
			alertRuleLifecycleSourceRefs(),
			"Alert rule lifecycle aggregate is represented by child contract gaps",
			monitoringContractScopeMetadata(
				"/alerts?section=rules",
				"alert_rule_lifecycle_aggregate",
				"alert-rule-lifecycle-aggregate",
				"owns detail,pure,create,update,delete,import,prom-import,clone,bulk-fields,enable,validate,tryrun,timezones,callbacks; excludes groups,shield,subscribe,notification,dashboard,template,metric,event",
			),
		),
		monitoringContractSeed(
			"FX-CONTRACT-N9E-NOTIFICATION-FINDX-ADAPTER",
			"Notification adapter aggregate",
			model.ContractStatusBlocked,
			notificationAdapterSourceRefs(),
			"Notification adapter aggregate represented by child contract gaps",
			monitoringContractScopeMetadata(
				"/notifications?section=rules",
				"notification_adapter_aggregate",
				"notification-adapter-aggregate",
				"owns rules,channel configs,templates,contacts,manage lookup,test,statistics; excludes alert rule groups,lifecycle,mute,shield,subscribe,event,action,share,ack,pipeline,query,dashboard,dashboard template center,metric",
			),
		),
		monitoringContractSeed(
			"FX-CONTRACT-N9E-ALERT-MUTE-SHIELD",
			"Alert mute shield aggregate",
			model.ContractStatusBlocked,
			alertMuteShieldSourceRefs(),
			"Alert mute shield aggregate is represented by child contract gaps",
			monitoringContractScopeMetadata(
				"/alerts?section=mutes",
				"alert_mute_shield_aggregate",
				"alert-mute-shield-aggregate",
				"owns only mute list,detail,create,update,delete,bulk-fields,preview,tryrun; excludes alert rule groups,lifecycle,subscribe,notification,dashboard,template,metric,event",
			),
		),
		monitoringContractSeed(
			"FX-CONTRACT-N9E-ALERT-EVENT-LIFECYCLE",
			"Alert event lifecycle aggregate",
			model.ContractStatusBlocked,
			alertEventLifecycleSourceRefs(),
			"Alert event lifecycle aggregate is represented by child contract gaps",
			monitoringContractScopeMetadata(
				"/alerts?section=events",
				"alert_event_lifecycle_aggregate",
				"alert-event-lifecycle-aggregate",
				"owns:current,history,list,detail,delete,notify,card,ds,csv,share-read; excludes:rule,mute,shield,subscribe,notif,pipeline,query,dashboard,template,metric,ack,share-cred",
			),
		),
		monitoringContractSeed(
			"FX-CONTRACT-N9E-ALERT-EVENT-ACTION-PIPELINE-QUERY",
			"Alert event action pipeline query aggregate",
			model.ContractStatusBlocked,
			alertEventActionPipelineQuerySourceRefs(),
			"Alert event action, acknowledgement, sharing, pipeline, and query aggregate is represented by child contract gaps",
			monitoringContractScopeMetadata(
				"/alerts?section=events",
				"alert_event_action_pipeline_query_aggregate",
				"alert-event-action-pipeline-query-aggregate",
				"owns ack,unack,share credential,shared detail,event pipeline crud,tryrun,tag lookup,enrich preview,executions,execution detail,event query/test selector; excludes 119B15/119B16",
			),
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
	entries = append(entries, monitoringAlertRuleGroupGapSeeds()...)
	entries = append(entries, monitoringAlertRuleLifecycleGapSeeds()...)
	entries = append(entries, monitoringNotificationAdapterGapSeeds()...)
	entries = append(entries, monitoringAlertMuteShieldGapSeeds()...)
	entries = append(entries, monitoringAlertSubscribeGapSeeds()...)
	entries = append(entries, monitoringAlertEventLifecycleGapSeeds()...)
	entries = append(entries, monitoringAlertEventActionPipelineQueryGapSeeds()...)
	return entries
}

func monitoringNotificationAdapterGapSeeds() []model.ContractMatrixRegisterRequest {
	return []model.ContractMatrixRegisterRequest{
		monitoringNotificationAdapterGapSeed("FX-CONTRACT-N9E-NOTIFY-RULES-LIST", "Notification rules list", model.ContractStatusMissingBackend, "notification_rules_list", "GET /api/n9e/notify-rules"),
		monitoringNotificationAdapterGapSeed("FX-CONTRACT-N9E-NOTIFY-RULE-DETAIL", "Notification rule detail", model.ContractStatusMissingBackend, "notification_rule_detail", "GET /api/n9e/notify-rule/{id}"),
		monitoringNotificationAdapterGapSeed("FX-CONTRACT-N9E-NOTIFY-RULES-CREATE", "Notification rules create", model.ContractStatusMissingBackend, "notification_rules_create", "POST /api/n9e/notify-rules"),
		monitoringNotificationAdapterGapSeed("FX-CONTRACT-N9E-NOTIFY-RULE-UPDATE", "Notification rule update", model.ContractStatusMissingBackend, "notification_rule_update", "PUT /api/n9e/notify-rule/{id}"),
		monitoringNotificationAdapterGapSeed("FX-CONTRACT-N9E-NOTIFY-RULES-DELETE", "Notification rules delete", model.ContractStatusMissingBackend, "notification_rules_delete", "DELETE /api/n9e/notify-rules"),
		monitoringNotificationAdapterGapSeed("FX-CONTRACT-N9E-NOTIFY-RULE-CUSTOM-PARAMS", "Notification rule custom params", model.ContractStatusMissingBackend, "notification_rule_custom_params", "GET /api/n9e/notify-rule/custom-params"),
		monitoringNotificationAdapterGapSeed("FX-CONTRACT-N9E-NOTIFY-RULE-TEST", "Notification rule test", model.ContractStatusMissingExecutor, "notification_rule_test", "POST /api/n9e/notify-rule/test"),
		monitoringNotificationAdapterGapSeed("FX-CONTRACT-N9E-NOTIFY-STATISTICS", "Notification statistics", model.ContractStatusMissingDatasource, "notification_statistics", "GET /api/n9e-plus/notify/{id}/statistics"),
		monitoringNotificationAdapterGapSeed("FX-CONTRACT-N9E-NOTIFY-EVENTS", "Notification events", model.ContractStatusMissingDatasource, "notification_events", "GET /api/n9e-plus/notify/{id}/alert-cur-events"),
		monitoringNotificationAdapterGapSeed("FX-CONTRACT-N9E-NOTIFY-ALERT-RULES", "Notification linked alert rules", model.ContractStatusMissingBackend, "notification_alert_rules", "GET /api/n9e-plus/notify/{id}/alert-rules"),
		monitoringNotificationAdapterGapSeed("FX-CONTRACT-N9E-NOTIFY-SUBSCRIBE-RULES", "Notification linked subscribe rules", model.ContractStatusMissingBackend, "notification_subscribe_rules", "GET /api/n9e-plus/notify/{id}/sub-alert-rules"),
		monitoringNotificationAdapterGapSeed("FX-CONTRACT-N9E-NOTIFY-EVENT-TAGKEYS", "Notification event tag keys", model.ContractStatusMissingDatasource, "notification_event_tagkeys", "GET /api/n9e/event-tagkeys"),
		monitoringNotificationAdapterGapSeed("FX-CONTRACT-N9E-NOTIFY-FEISHU-GROUPS", "Notification Feishu groups lookup", model.ContractStatusMissingExecutor, "notification_feishu_groups", "GET /api/n9e/feishu-visible-chats/{id}"),
		monitoringNotificationAdapterGapSeed("FX-CONTRACT-N9E-NOTIFY-FLASHDUTY-CHANNELS", "Notification FlashDuty channels lookup", model.ContractStatusMissingExecutor, "notification_flashduty_channels", "GET /api/n9e/flashduty-channel-list/{id}"),
		monitoringNotificationAdapterGapSeed("FX-CONTRACT-N9E-NOTIFY-PAGERDUTY-SERVICES", "Notification PagerDuty services lookup", model.ContractStatusMissingExecutor, "notification_pagerduty_services", "pagerduty-services-lookup"),
		monitoringNotificationAdapterGapSeed("FX-CONTRACT-N9E-NOTIFY-PAGERDUTY-CONNECTOR-LOOKUP", "Notification PagerDuty connector lookup", model.ContractStatusMissingExecutor, "notification_pagerduty_connector_lookup", "pagerduty-connector-lookup"),
		monitoringNotificationAdapterGapSeed("FX-CONTRACT-N9E-NOTIFY-CHANNEL-CONFIGS-LIST", "Notification channel configs list", model.ContractStatusMissingBackend, "notification_channel_configs_list", "GET /api/n9e/notify-channel-configs"),
		monitoringNotificationAdapterGapSeed("FX-CONTRACT-N9E-NOTIFY-CHANNEL-CONFIGS-SIMPLIFIED", "Notification channel configs simplified", model.ContractStatusMissingBackend, "notification_channel_configs_simplified", "GET /api/n9e/simplified-notify-channel-configs"),
		monitoringNotificationAdapterGapSeed("FX-CONTRACT-N9E-NOTIFY-CHANNEL-CONFIGS-CREATE", "Notification channel configs create", model.ContractStatusMissingBackend, "notification_channel_configs_create", "POST /api/n9e/notify-channel-configs"),
		monitoringNotificationAdapterGapSeed("FX-CONTRACT-N9E-NOTIFY-CHANNEL-CONFIG-UPDATE", "Notification channel config update", model.ContractStatusMissingBackend, "notification_channel_config_update", "PUT /api/n9e/notify-channel-config/{id}"),
		monitoringNotificationAdapterGapSeed("FX-CONTRACT-N9E-NOTIFY-CHANNEL-CONFIG-DETAIL", "Notification channel config detail", model.ContractStatusMissingBackend, "notification_channel_config_detail", "GET /api/n9e/notify-channel-config/{id}"),
		monitoringNotificationAdapterGapSeed("FX-CONTRACT-N9E-NOTIFY-CHANNEL-CONFIG-BY-IDENT", "Notification channel config by ident", model.ContractStatusMissingBackend, "notification_channel_config_by_ident", "GET /api/n9e/notify-channel-config?ident={ident}"),
		monitoringNotificationAdapterGapSeed("FX-CONTRACT-N9E-NOTIFY-CHANNEL-CONFIGS-DELETE", "Notification channel configs delete", model.ContractStatusMissingBackend, "notification_channel_configs_delete", "DELETE /api/n9e/notify-channel-configs"),
		monitoringNotificationAdapterGapSeed("FX-CONTRACT-N9E-NOTIFY-CHANNEL-CONFIG-IDENTS", "Notification channel config idents", model.ContractStatusMissingBackend, "notification_channel_config_idents", "GET /api/n9e/notify-channel-config/idents"),
		monitoringNotificationAdapterGapSeed("FX-CONTRACT-N9E-MESSAGE-TEMPLATES-LIST", "Message templates list", model.ContractStatusMissingBackend, "message_templates_list", "GET /api/n9e/message-templates"),
		monitoringNotificationAdapterGapSeed("FX-CONTRACT-N9E-MESSAGE-TEMPLATE-DETAIL", "Message template detail", model.ContractStatusMissingBackend, "message_template_detail", "GET /api/n9e/message-template/{id}"),
		monitoringNotificationAdapterGapSeed("FX-CONTRACT-N9E-MESSAGE-TEMPLATES-CREATE", "Message templates create", model.ContractStatusMissingBackend, "message_templates_create", "POST /api/n9e/message-templates"),
		monitoringNotificationAdapterGapSeed("FX-CONTRACT-N9E-MESSAGE-TEMPLATE-UPDATE", "Message template update", model.ContractStatusMissingBackend, "message_template_update", "PUT /api/n9e/message-template/{id}"),
		monitoringNotificationAdapterGapSeed("FX-CONTRACT-N9E-MESSAGE-TEMPLATES-DELETE", "Message templates delete", model.ContractStatusMissingBackend, "message_templates_delete", "DELETE /api/n9e/message-templates"),
		monitoringNotificationAdapterGapSeed("FX-CONTRACT-N9E-MESSAGE-TEMPLATE-PREVIEW", "Message template preview", model.ContractStatusMissingDatasource, "message_template_preview", "POST /api/n9e/events-message"),
		monitoringNotificationAdapterGapSeed("FX-CONTRACT-N9E-NOTIFY-CONTACTS-LIST", "Notification contacts list", model.ContractStatusMissingBackend, "notification_contacts_list", "GET /api/n9e/notify-contact"),
		monitoringNotificationAdapterGapSeed("FX-CONTRACT-N9E-NOTIFY-CONTACTS-UPDATE", "Notification contacts update", model.ContractStatusMissingBackend, "notification_contacts_update", "PUT /api/n9e/notify-contact"),
		monitoringNotificationAdapterGapSeed("FX-CONTRACT-N9E-MANAGE-NOTIFY-CHANNELS", "Manage notify channels lookup", model.ContractStatusMissingBackend, "manage_notify_channels", "GET /api/n9e/notify-channels"),
		monitoringNotificationAdapterGapSeed("FX-CONTRACT-N9E-MANAGE-CONTACT-CHANNELS", "Manage contact channels lookup", model.ContractStatusMissingBackend, "manage_contact_channels", "GET /api/n9e/contact-channels"),
		monitoringNotificationAdapterGapSeed("FX-CONTRACT-N9E-MANAGE-CONTACT-KEYS", "Manage contact keys lookup", model.ContractStatusMissingBackend, "manage_contact_keys", "GET /api/n9e/contact-keys"),
	}
}

func monitoringNotificationAdapterGapSeed(id, capability, status, gapType, upstreamRef string) model.ContractMatrixRegisterRequest {
	reason := "Notification adapter backend contract is missing"
	if status == model.ContractStatusMissingDatasource {
		reason = "Notification adapter datasource contract is missing"
	}
	if status == model.ContractStatusMissingExecutor {
		reason = "Notification adapter executor contract is missing"
	}
	return monitoringContractSeed(
		id,
		capability,
		status,
		notificationAdapterSourceRefs(),
		reason,
		monitoringContractMetadata("/notifications?section=rules", gapType, upstreamRef),
	)
}

func monitoringAlertRuleGroupGapSeeds() []model.ContractMatrixRegisterRequest {
	return []model.ContractMatrixRegisterRequest{
		monitoringAlertRuleGroupGapSeed("FX-CONTRACT-N9E-ALERT-RULE-GROUP-LIST", "Alert rule group list", warningSourceRefs(), "alert_rule_group_list", "GET /api/n9e/alert-rule-groups"),
		monitoringAlertRuleGroupGapSeed("FX-CONTRACT-N9E-ALERT-RULE-GROUP-CREATE", "Alert rule group create", warningSourceRefs(), "alert_rule_group_create", "POST /api/n9e/alert-rule-groups"),
		monitoringAlertRuleGroupGapSeed("FX-CONTRACT-N9E-ALERT-RULE-GROUP-DETAIL", "Alert rule group detail", warningSourceRefs(), "alert_rule_group_detail", "GET /api/n9e/alert-rule-group/{id}"),
		monitoringAlertRuleGroupGapSeed("FX-CONTRACT-N9E-ALERT-RULE-GROUP-UPDATE", "Alert rule group update", warningSourceRefs(), "alert_rule_group_update", "PUT /api/n9e/alert-rule-group/{id}"),
		monitoringAlertRuleGroupGapSeed("FX-CONTRACT-N9E-ALERT-RULE-GROUP-DELETE", "Alert rule group delete", warningSourceRefs(), "alert_rule_group_delete", "DELETE /api/n9e/alert-rule-group/{id}"),
		monitoringAlertRuleGroupGapSeed("FX-CONTRACT-N9E-ALERT-RULE-GROUP-RULES-LIST", "Alert rule group rules list", warningAndRuleModalSourceRefs(), "alert_rule_group_rules_list", "GET /api/n9e/busi-group/{id}/alert-rules"),
		monitoringAlertRuleGroupGapSeed("FX-CONTRACT-N9E-ALERT-RULE-GROUPS-MULTI-RULES-LIST", "Alert rule groups multi rules list", warningSourceRefs(), "alert_rule_group_multi_rules_list", "GET /api/n9e/busi-groups/alert-rules"),
		monitoringAlertRuleGroupGapSeed("FX-CONTRACT-N9E-ALERT-RULE-GROUP-FAVORITES-LIST", "Alert rule group favorites list", warningSourceRefs(), "alert_rule_group_favorites_list", "GET /api/n9e/alert-rule-groups/favorites"),
		monitoringAlertRuleGroupGapSeed("FX-CONTRACT-N9E-ALERT-RULE-GROUP-FAVORITE-ADD", "Alert rule group favorite add", warningSourceRefs(), "alert_rule_group_favorite_add", "POST /api/n9e/alert-rule-group/{id}/favorites"),
		monitoringAlertRuleGroupGapSeed("FX-CONTRACT-N9E-ALERT-RULE-GROUP-FAVORITE-DELETE", "Alert rule group favorite delete", warningSourceRefs(), "alert_rule_group_favorite_delete", "DELETE /api/n9e/alert-rule-group/{id}/favorites"),
	}
}

func monitoringAlertRuleGroupGapSeed(id, capability string, sourceRefs []string, gapType, upstreamRef string) model.ContractMatrixRegisterRequest {
	return monitoringContractSeed(
		id,
		capability,
		model.ContractStatusMissingBackend,
		sourceRefs,
		"Alert rule group backend contract is missing",
		monitoringContractMetadata("/alerts?section=rules", gapType, upstreamRef),
	)
}

func monitoringAlertRuleLifecycleGapSeeds() []model.ContractMatrixRegisterRequest {
	return []model.ContractMatrixRegisterRequest{
		monitoringAlertRuleLifecycleGapSeed("FX-CONTRACT-N9E-ALERT-RULE-DETAIL", "Alert rule detail", model.ContractStatusMissingBackend, "alert_rule_detail", "GET /api/n9e/alert-rule/{id}"),
		monitoringAlertRuleLifecycleGapSeed("FX-CONTRACT-N9E-ALERT-RULE-PURE-DETAIL", "Alert rule pure detail", model.ContractStatusMissingBackend, "alert_rule_pure_detail", "GET /api/n9e/alert-rule/{id}/pure"),
		monitoringAlertRuleLifecycleGapSeed("FX-CONTRACT-N9E-ALERT-RULE-CREATE", "Alert rule create", model.ContractStatusMissingBackend, "alert_rule_create", "POST /api/n9e/busi-group/{busiId}/alert-rules"),
		monitoringAlertRuleLifecycleGapSeed("FX-CONTRACT-N9E-ALERT-RULE-UPDATE", "Alert rule update", model.ContractStatusMissingBackend, "alert_rule_update", "PUT /api/n9e/busi-group/{busiId}/alert-rule/{strategyId}"),
		monitoringAlertRuleLifecycleGapSeed("FX-CONTRACT-N9E-ALERT-RULE-DELETE-BY-BUSI-GROUP", "Alert rule delete by business group", model.ContractStatusMissingBackend, "alert_rule_delete_by_busi_group", "DELETE /api/n9e/busi-group/{busiId}/alert-rules"),
		monitoringAlertRuleLifecycleGapSeed("FX-CONTRACT-N9E-ALERT-RULE-DELETE-BY-RULE-GROUP", "Alert rule delete by rule group", model.ContractStatusMissingBackend, "alert_rule_delete_by_rule_group", "DELETE /api/n9e/alert-rule-group/{ruleGroupId}/alert-rules"),
		monitoringAlertRuleLifecycleGapSeed("FX-CONTRACT-N9E-ALERT-RULE-IMPORT-JSON", "Alert rule import JSON", model.ContractStatusMissingBackend, "alert_rule_import_json", "POST /api/n9e/busi-group/{busiId}/alert-rules/import"),
		monitoringAlertRuleLifecycleGapSeed("FX-CONTRACT-N9E-ALERT-RULE-IMPORT-PROM-RULE", "Alert rule import Prometheus rule", model.ContractStatusMissingBackend, "alert_rule_import_prom_rule", "POST /api/n9e/busi-group/{busiId}/alert-rules/import-prom-rule"),
		monitoringAlertRuleLifecycleGapSeed("FX-CONTRACT-N9E-ALERT-RULE-BULK-FIELDS-UPDATE", "Alert rule bulk fields update", model.ContractStatusMissingBackend, "alert_rule_bulk_fields_update", "PUT /api/n9e/busi-group/{busiId}/alert-rules/fields"),
		monitoringAlertRuleLifecycleGapSeed("FX-CONTRACT-N9E-ALERT-RULE-ENABLE-DISABLE", "Alert rule enable disable", model.ContractStatusMissingBackend, "alert_rule_enable_disable", "PUT /api/n9e/busi-group/{busiId}/alert-rules/fields"),
		monitoringAlertRuleLifecycleGapSeed("FX-CONTRACT-N9E-ALERT-RULE-STATUS-BATCH", "Alert rule status batch", model.ContractStatusMissingBackend, "alert_rule_status_batch", "PUT /api/n9e/alert-rules/status"),
		monitoringAlertRuleLifecycleGapSeed("FX-CONTRACT-N9E-ALERT-RULE-CLONE-TO-HOSTS", "Alert rule clone to hosts", model.ContractStatusMissingBackend, "alert_rule_clone_to_hosts", "POST /api/n9e/busi-group/{gid}/alert-rules/clone"),
		monitoringAlertRuleLifecycleGapSeed("FX-CONTRACT-N9E-ALERT-RULE-CLONE-TO-BUSI-GROUPS", "Alert rule clone to business groups", model.ContractStatusMissingBackend, "alert_rule_clone_to_busi_groups", "POST /api/n9e/busi-groups/alert-rules/clones"),
		monitoringAlertRuleLifecycleGapSeed("FX-CONTRACT-N9E-ALERT-RULE-VALIDATE", "Alert rule validate", model.ContractStatusMissingBackend, "alert_rule_validate", "PUT /api/n9e/busi-group/alert-rule/validate"),
		monitoringAlertRuleLifecycleGapSeed("FX-CONTRACT-N9E-ALERT-RULE-ENABLE-TRYRUN", "Alert rule enable tryrun", model.ContractStatusMissingDatasource, "alert_rule_enable_tryrun", "POST /api/n9e/busi-group/alert-rules/enable-tryrun"),
		monitoringAlertRuleLifecycleGapSeed("FX-CONTRACT-N9E-ALERT-RULE-CALLBACKS-LIST", "Alert rule callbacks list", model.ContractStatusMissingBackend, "alert_rule_callbacks_list", "GET /api/n9e/alert-rules/callbacks"),
		monitoringAlertRuleLifecycleGapSeed("FX-CONTRACT-N9E-ALERT-RULE-TIMEZONES", "Alert rule timezones", model.ContractStatusMissingBackend, "alert_rule_timezones", "GET /api/n9e/timezones"),
	}
}

func monitoringAlertRuleLifecycleGapSeed(id, capability, status, gapType, upstreamRef string) model.ContractMatrixRegisterRequest {
	return monitoringContractSeed(
		id,
		capability,
		status,
		alertRuleLifecycleSourceRefs(),
		"Alert rule lifecycle backend contract is missing",
		monitoringContractMetadata("/alerts?section=rules", gapType, upstreamRef),
	)
}

func warningSourceRefs() []string {
	return []string{`D:\项目迁移文件\平台源码\fe-main\src\services\warning.ts`}
}

func notificationAdapterSourceRefs() []string {
	return notificationAdapterMatureSourceRefs(
		`fe-main\src\pages\notificationRules\services.ts`,
		`fe-main\src\pages\notificationRules\pages\List.tsx`,
		`fe-main\src\pages\notificationRules\pages\Form\TestButton.tsx`,
		`fe-main\src\pages\notificationRules\pages\Detail\index.tsx`,
		`fe-main\src\pages\notificationRules\pages\Detail\Events.tsx`,
		`fe-main\src\pages\notificationRules\pages\Detail\AlertRules.tsx`,
		`fe-main\src\pages\notificationRules\pages\Detail\SubscribeRules.tsx`,
		`fe-main\src\pages\notificationChannels\services.ts`,
		`fe-main\src\pages\notificationChannels\pages\ListNG\index.tsx`,
		`fe-main\src\pages\notificationChannels\pages\Form\index.tsx`,
		`fe-main\src\pages\notificationTemplates\services.ts`,
		`fe-main\src\pages\notificationTemplates\pages\List\index.tsx`,
		`fe-main\src\pages\contacts\services.ts`,
		`fe-main\src\pages\contacts\pages\List.tsx`,
		`fe-main\src\services\manage.ts`,
	)
}

func notificationAdapterMatureSourceRefs(relativePaths ...string) []string {
	refs := make([]string, 0, len(relativePaths))
	for _, relativePath := range relativePaths {
		refs = append(refs, `D:\项目迁移文件\平台源码\`+relativePath)
	}
	return refs
}

func monitoringAlertEventLifecycleGapSeeds() []model.ContractMatrixRegisterRequest {
	return []model.ContractMatrixRegisterRequest{
		monitoringAlertEventLifecycleGapSeed("FX-CONTRACT-N9E-ALERT-EVENT-CURRENT-LIST", "Alert current event list", model.ContractStatusMissingBackend, "alert_event_current_list", "GET /api/n9e/alert-cur-events/list; GET /api/n9e-plus/alert-cur-events/list", "Alert current event list backend contract is missing"),
		monitoringAlertEventLifecycleGapSeed("FX-CONTRACT-N9E-ALERT-EVENT-CURRENT-DATASOURCES", "Alert current event datasources", model.ContractStatusMissingBackend, "alert_event_current_datasources", "GET /api/n9e/alert-cur-events-datasources", "Alert current event datasource list backend contract is missing"),
		monitoringAlertEventLifecycleGapSeed("FX-CONTRACT-N9E-ALERT-EVENT-CURRENT-CARD-LIST", "Alert current event card list", model.ContractStatusMissingDatasource, "alert_event_current_card_list", "GET /api/n9e/alert-cur-events/card", "Alert current event card datasource contract is missing"),
		monitoringAlertEventLifecycleGapSeed("FX-CONTRACT-N9E-ALERT-EVENT-CURRENT-CARD-DETAILS", "Alert current event card details", model.ContractStatusMissingDatasource, "alert_event_current_card_details", "POST /api/n9e/alert-cur-events/card/details; POST /api/n9e-plus/alert-cur-events/card/details", "Alert current event card details datasource contract is missing"),
		monitoringAlertEventLifecycleGapSeed("FX-CONTRACT-N9E-ALERT-EVENT-DETAIL", "Alert event detail", model.ContractStatusMissingBackend, "alert_event_detail", "GET /api/n9e/alert-his-event/{eventId}; GET /api/n9e-plus/alert-his-event/{eventId}", "Alert event detail backend contract is missing"),
		monitoringAlertEventLifecycleGapSeed("FX-CONTRACT-N9E-ALERT-EVENT-CURRENT-DELETE", "Alert current event delete", model.ContractStatusMissingBackend, "alert_event_current_delete", "DELETE /api/n9e/alert-cur-events", "Alert current event delete backend contract is missing"),
		monitoringAlertEventLifecycleGapSeed("FX-CONTRACT-N9E-ALERT-EVENT-HISTORY-LIST", "Alert history event list", model.ContractStatusMissingBackend, "alert_event_history_list", "GET /api/n9e/alert-his-events/list", "Alert history event list backend contract is missing"),
		monitoringAlertEventLifecycleGapSeed("FX-CONTRACT-N9E-ALERT-EVENT-HISTORY-BY-IDS", "Alert history events by IDs", model.ContractStatusMissingBackend, "alert_event_history_by_ids", "GET /api/n9e-plus/alert-his-events/{ids}", "Alert history events by IDs backend contract is missing"),
		monitoringAlertEventLifecycleGapSeed("FX-CONTRACT-N9E-ALERT-EVENT-HISTORY-CLEANUP", "Alert history event cleanup", model.ContractStatusMissingBackend, "alert_event_history_cleanup", "DELETE /api/n9e/alert-his-events", "Alert history event cleanup backend contract is missing"),
		monitoringAlertEventLifecycleGapSeed("FX-CONTRACT-N9E-ALERT-EVENT-NOTIFY-RECORDS", "Alert event notify records", model.ContractStatusMissingBackend, "alert_event_notify_records", "GET /api/n9e/event-notify-records/{eventId}", "Alert event notify records backend contract is missing"),
	}
}

func monitoringAlertEventLifecycleGapSeed(id, capability, status, gapType, upstreamRef, blockedReason string) model.ContractMatrixRegisterRequest {
	return monitoringContractSeed(
		id,
		capability,
		status,
		alertEventLifecycleSourceRefs(),
		blockedReason,
		monitoringContractMetadata("/alerts?section=events", gapType, upstreamRef),
	)
}

func monitoringAlertEventActionPipelineQueryGapSeeds() []model.ContractMatrixRegisterRequest {
	return []model.ContractMatrixRegisterRequest{
		monitoringAlertEventActionPipelineQueryGapSeed("FX-CONTRACT-N9E-ALERT-EVENT-ACK", "Alert current event acknowledge", model.ContractStatusMissingBackend, "alert_event_ack", "POST /api/n9e-plus/alert-cur-events/{action}", "Alert event acknowledgement backend contract is missing"),
		monitoringAlertEventActionPipelineQueryGapSeed("FX-CONTRACT-N9E-ALERT-EVENT-SHARE-CREDENTIAL", "Alert event share credential issue", model.ContractStatusMissingBackend, "alert_event_share_credential", "share-credential-issue", "Alert event share credential backend contract is missing"),
		monitoringAlertEventActionPipelineQueryGapSeed("FX-CONTRACT-N9E-ALERT-EVENT-SHARED-DETAIL", "Alert event shared detail", model.ContractStatusMissingBackend, "alert_event_shared_detail", "GET /api/n9e/alert-his-event/{eventId}; GET /api/n9e-plus/alert-his-event/{eventId}", "Alert event shared detail backend contract is missing"),
		monitoringAlertEventActionPipelineQueryGapSeed("FX-CONTRACT-N9E-EVENT-PIPELINE-CRUD", "Event pipeline CRUD aggregate", model.ContractStatusBlocked, "event_pipeline_crud_aggregate", "event-pipeline-crud-aggregate", "Event pipeline CRUD aggregate is represented by child contract gaps"),
		monitoringAlertEventActionPipelineQueryGapSeed("FX-CONTRACT-N9E-EVENT-PIPELINE-LIST", "Event pipeline list", model.ContractStatusMissingBackend, "event_pipeline_list", "GET /api/n9e/event-pipelines", "Event pipeline list backend contract is missing"),
		monitoringAlertEventActionPipelineQueryGapSeed("FX-CONTRACT-N9E-EVENT-PIPELINE-DETAIL", "Event pipeline detail", model.ContractStatusMissingBackend, "event_pipeline_detail", "GET /api/n9e/event-pipeline/{id}", "Event pipeline detail backend contract is missing"),
		monitoringAlertEventActionPipelineQueryGapSeed("FX-CONTRACT-N9E-EVENT-PIPELINE-CREATE", "Event pipeline create", model.ContractStatusMissingBackend, "event_pipeline_create", "POST /api/n9e/event-pipeline", "Event pipeline create backend contract is missing"),
		monitoringAlertEventActionPipelineQueryGapSeed("FX-CONTRACT-N9E-EVENT-PIPELINE-UPDATE", "Event pipeline update", model.ContractStatusMissingBackend, "event_pipeline_update", "PUT /api/n9e/event-pipeline", "Event pipeline update backend contract is missing"),
		monitoringAlertEventActionPipelineQueryGapSeed("FX-CONTRACT-N9E-EVENT-PIPELINE-DELETE", "Event pipeline delete", model.ContractStatusMissingBackend, "event_pipeline_delete", "DELETE /api/n9e/event-pipelines", "Event pipeline delete backend contract is missing"),
		monitoringAlertEventActionPipelineQueryGapSeed("FX-CONTRACT-N9E-EVENT-PROCESSOR-TRYRUN", "Event processor tryrun", model.ContractStatusMissingExecutor, "event_processor_tryrun", "POST /api/n9e/event-processor-tryrun", "Event processor tryrun executor contract is missing"),
		monitoringAlertEventActionPipelineQueryGapSeed("FX-CONTRACT-N9E-EVENT-PIPELINE-TRYRUN", "Event pipeline tryrun", model.ContractStatusMissingExecutor, "event_pipeline_tryrun", "POST /api/n9e/event-pipeline-tryrun", "Event pipeline tryrun executor contract is missing"),
		monitoringAlertEventActionPipelineQueryGapSeed("FX-CONTRACT-N9E-EVENT-TAGKEYS", "Event tag keys", model.ContractStatusMissingDatasource, "event_tagkeys", "GET /api/n9e/event-tagkeys", "Event tag keys datasource contract is missing"),
		monitoringAlertEventActionPipelineQueryGapSeed("FX-CONTRACT-N9E-EVENT-TAGVALUES", "Event tag values", model.ContractStatusMissingDatasource, "event_tagvalues", "GET /api/n9e/event-tagvalues?key={key}", "Event tag values datasource contract is missing"),
		monitoringAlertEventActionPipelineQueryGapSeed("FX-CONTRACT-N9E-EVENT-ENRICH-DATA-PREVIEW", "Event enrich data preview", model.ContractStatusMissingDatasource, "event_enrich_data_preview", "POST /api/n9e-plus/event-enrich-data-preview", "Event enrich data preview datasource contract is missing"),
		monitoringAlertEventActionPipelineQueryGapSeed("FX-CONTRACT-N9E-EVENT-PIPELINE-EXECUTIONS-LIST", "Event pipeline executions list", model.ContractStatusMissingBackend, "event_pipeline_executions_list", "GET /api/n9e/event-pipeline-executions", "Event pipeline executions list backend contract is missing"),
		monitoringAlertEventActionPipelineQueryGapSeed("FX-CONTRACT-N9E-EVENT-PIPELINE-EXECUTION-DETAIL", "Event pipeline execution detail", model.ContractStatusMissingBackend, "event_pipeline_execution_detail", "GET /api/n9e/event-pipeline-execution/{id}", "Event pipeline execution detail backend contract is missing"),
		monitoringAlertEventActionPipelineQueryGapSeed("FX-CONTRACT-N9E-ALERT-EVENT-RULE-TESTER", "Alert event rule tester", model.ContractStatusBlocked, "alert_event_rule_tester", "AlertEventRuleTesterWithButton onClick/onTest event selector", "Alert event rule tester selector contract is blocked"),
	}
}

func monitoringAlertEventActionPipelineQueryGapSeed(id, capability, status, gapType, upstreamRef, blockedReason string) model.ContractMatrixRegisterRequest {
	return monitoringContractSeed(
		id,
		capability,
		status,
		alertEventActionPipelineQuerySourceRefs(),
		blockedReason,
		monitoringContractMetadata("/alerts?section=events", gapType, upstreamRef),
	)
}

func monitoringAlertMuteShieldGapSeeds() []model.ContractMatrixRegisterRequest {
	return []model.ContractMatrixRegisterRequest{
		monitoringAlertMuteShieldGapSeed("FX-CONTRACT-N9E-ALERT-MUTE-LIST-BY-BUSI-GROUP", "Alert mute list by business group", model.ContractStatusMissingBackend, "alert_mute_list_by_busi_group", "GET /api/n9e/busi-group/{id}/alert-mutes", "Alert mute shield backend contract is missing"),
		monitoringAlertMuteShieldGapSeed("FX-CONTRACT-N9E-ALERT-MUTE-LIST-BY-BUSI-GROUPS", "Alert mute list by business groups", model.ContractStatusMissingBackend, "alert_mute_list_by_busi_groups", "GET /api/n9e/busi-groups/alert-mutes", "Alert mute shield backend contract is missing"),
		monitoringAlertMuteShieldGapSeed("FX-CONTRACT-N9E-ALERT-MUTE-DETAIL", "Alert mute detail", model.ContractStatusMissingBackend, "alert_mute_detail", "GET /api/n9e/busi-group/{busiId}/alert-mute/{id}", "Alert mute shield backend contract is missing"),
		monitoringAlertMuteShieldGapSeed("FX-CONTRACT-N9E-ALERT-MUTE-CREATE", "Alert mute create", model.ContractStatusMissingBackend, "alert_mute_create", "POST /api/n9e/busi-group/{busiId}/alert-mutes", "Alert mute shield backend contract is missing"),
		monitoringAlertMuteShieldGapSeed("FX-CONTRACT-N9E-ALERT-MUTE-UPDATE", "Alert mute update", model.ContractStatusMissingBackend, "alert_mute_update", "PUT /api/n9e/busi-group/{busiId}/alert-mute/{muteId}", "Alert mute shield backend contract is missing"),
		monitoringAlertMuteShieldGapSeed("FX-CONTRACT-N9E-ALERT-MUTE-DELETE", "Alert mute delete", model.ContractStatusMissingBackend, "alert_mute_delete", "DELETE /api/n9e/busi-group/{busiId}/alert-mutes", "Alert mute shield backend contract is missing"),
		monitoringAlertMuteShieldGapSeed("FX-CONTRACT-N9E-ALERT-MUTE-BULK-FIELDS-UPDATE", "Alert mute bulk fields update", model.ContractStatusMissingBackend, "alert_mute_bulk_fields_update", "PUT /api/n9e/busi-group/{busiId}/alert-mutes/fields", "Alert mute shield backend contract is missing"),
		monitoringAlertMuteShieldGapSeed("FX-CONTRACT-N9E-ALERT-MUTE-PREVIEW-EVENTS", "Alert mute preview events", model.ContractStatusMissingDatasource, "alert_mute_preview_events", "POST /api/n9e/busi-group/{busiId}/alert-mutes/preview", "Alert mute shield datasource contract is missing"),
		monitoringAlertMuteShieldGapSeed("FX-CONTRACT-N9E-ALERT-MUTE-TRYRUN", "Alert mute tryrun", model.ContractStatusMissingDatasource, "alert_mute_tryrun", "POST /api/n9e/alert-mute-tryrun", "Alert mute shield datasource contract is missing"),
	}
}

func monitoringAlertMuteShieldGapSeed(id, capability, status, gapType, upstreamRef, blockedReason string) model.ContractMatrixRegisterRequest {
	return monitoringContractSeed(
		id,
		capability,
		status,
		alertMuteShieldSourceRefs(),
		blockedReason,
		monitoringContractMetadata("/alerts?section=mutes", gapType, upstreamRef),
	)
}

func alertRuleLifecycleSourceRefs() []string {
	return []string{
		`D:\项目迁移文件\平台源码\fe-main\src\services\warning.ts`,
		`D:\项目迁移文件\平台源码\fe-main\src\pages\alertRules\services.ts`,
		`D:\项目迁移文件\平台源码\fe-main\src\pages\alertRules\Edit.tsx`,
		`D:\项目迁移文件\平台源码\fe-main\src\pages\alertRules\Form\index.tsx`,
		`D:\项目迁移文件\平台源码\fe-main\src\pages\alertRules\List\ListNG.tsx`,
		`D:\项目迁移文件\平台源码\fe-main\src\pages\alertRules\List\MoreOperations.tsx`,
		`D:\项目迁移文件\平台源码\fe-main\src\pages\alertRules\List\Import\ImportBase.tsx`,
		`D:\项目迁移文件\平台源码\fe-main\src\pages\alertRules\List\Import\ImportPrometheus.tsx`,
		`D:\项目迁移文件\平台源码\fe-main\src\pages\alertRules\List\CloneToHosts\index.tsx`,
		`D:\项目迁移文件\平台源码\fe-main\src\pages\alertRules\List\CloneToBgids\index.tsx`,
		`D:\项目迁移文件\平台源码\fe-main\src\pages\alertRules\Form\Notify\index.tsx`,
		`D:\项目迁移文件\平台源码\fe-main\src\pages\alertRules\Form\Effective\index.tsx`,
		`D:\项目迁移文件\平台源码\fe-main\src\pages\alertRules\List\EditModal.tsx`,
	}
}

func alertMuteShieldSourceRefs() []string {
	return []string{
		`D:\项目迁移文件\平台源码\fe-main\src\services\shield.ts`,
		`D:\项目迁移文件\平台源码\fe-main\src\pages\warning\shield\index.tsx`,
		`D:\项目迁移文件\平台源码\fe-main\src\pages\warning\shield\add.tsx`,
		`D:\项目迁移文件\平台源码\fe-main\src\pages\warning\shield\edit.tsx`,
		`D:\项目迁移文件\平台源码\fe-main\src\pages\warning\shield\components\operateForm.tsx`,
		`D:\项目迁移文件\平台源码\fe-main\src\pages\warning\shield\components\PreviewMutedEvents.tsx`,
		`D:\项目迁移文件\平台源码\fe-main\src\pages\warning\shield\components\utils.ts`,
		`D:\项目迁移文件\平台源码\fe-main\src\pages\warning\shield\components\CateSelect\index.tsx`,
	}
}

func monitoringAlertSubscribeGapSeeds() []model.ContractMatrixRegisterRequest {
	entries := []model.ContractMatrixRegisterRequest{
		monitoringContractSeed(
			"FX-CONTRACT-N9E-ALERT-SUBSCRIBE",
			"Alert subscribe aggregate",
			model.ContractStatusBlocked,
			alertSubscribeSourceRefs(),
			"Alert subscribe aggregate is represented by child contract gaps",
			monitoringContractScopeMetadata(
				"/alerts?section=subscribes",
				"alert_subscribe_aggregate",
				"alert-subscribe-aggregate",
				"owns only subscribe list,detail,create,update,delete,tryrun; excludes alert rule groups,lifecycle,mute,shield,notification,dashboard,template,metric,event",
			),
		),
	}
	entries = append(entries,
		monitoringAlertSubscribeGapSeed("FX-CONTRACT-N9E-ALERT-SUBSCRIBE-LIST-BY-BUSI-GROUP", "Alert subscribe list by business group", model.ContractStatusMissingBackend, "alert_subscribe_list_by_busi_group", "GET /api/n9e/busi-group/{id}/alert-subscribes", "Alert subscribe backend contract is missing"),
		monitoringAlertSubscribeGapSeed("FX-CONTRACT-N9E-ALERT-SUBSCRIBE-LIST-BY-BUSI-GROUPS", "Alert subscribe list by business groups", model.ContractStatusMissingBackend, "alert_subscribe_list_by_busi_groups", "GET /api/n9e/busi-groups/alert-subscribes", "Alert subscribe backend contract is missing"),
		monitoringAlertSubscribeGapSeed("FX-CONTRACT-N9E-ALERT-SUBSCRIBE-DETAIL", "Alert subscribe detail", model.ContractStatusMissingBackend, "alert_subscribe_detail", "GET /api/n9e/alert-subscribe/{subscribeId}", "Alert subscribe backend contract is missing"),
		monitoringAlertSubscribeGapSeed("FX-CONTRACT-N9E-ALERT-SUBSCRIBE-CREATE", "Alert subscribe create", model.ContractStatusMissingBackend, "alert_subscribe_create", "POST /api/n9e/busi-group/{busiId}/alert-subscribes", "Alert subscribe backend contract is missing"),
		monitoringAlertSubscribeGapSeed("FX-CONTRACT-N9E-ALERT-SUBSCRIBE-UPDATE", "Alert subscribe update", model.ContractStatusMissingBackend, "alert_subscribe_update", "PUT /api/n9e/busi-group/{busiId}/alert-subscribes", "Alert subscribe backend contract is missing"),
		monitoringAlertSubscribeGapSeed("FX-CONTRACT-N9E-ALERT-SUBSCRIBE-DELETE", "Alert subscribe delete", model.ContractStatusMissingBackend, "alert_subscribe_delete", "DELETE /api/n9e/busi-group/{busiId}/alert-subscribes", "Alert subscribe backend contract is missing"),
		monitoringAlertSubscribeGapSeed("FX-CONTRACT-N9E-ALERT-SUBSCRIBE-TRYRUN", "Alert subscribe tryrun", model.ContractStatusMissingDatasource, "alert_subscribe_tryrun", "POST /api/n9e/alert-subscribe/alert-subscribes-tryrun", "Alert subscribe datasource contract is missing"),
	)
	return entries
}

func monitoringAlertSubscribeGapSeed(id, capability, status, gapType, upstreamRef, blockedReason string) model.ContractMatrixRegisterRequest {
	return monitoringContractSeed(
		id,
		capability,
		status,
		alertSubscribeSourceRefs(),
		blockedReason,
		monitoringContractMetadata("/alerts?section=subscribes", gapType, upstreamRef),
	)
}

func alertSubscribeSourceRefs() []string {
	return alertSubscribeMatureSourceRefs(
		`fe-main\src\services\subscribe.ts`,
		`fe-main\src\pages\warning\subscribe\index.tsx`,
		`fe-main\src\pages\warning\subscribe\ListNG.tsx`,
		`fe-main\src\pages\warning\subscribe\add.tsx`,
		`fe-main\src\pages\warning\subscribe\edit.tsx`,
		`fe-main\src\pages\warning\subscribe\components\operateForm.tsx`,
		`fe-main\src\pages\warning\subscribe\components\ruleModal.tsx`,
		`fe-main\src\pages\warning\subscribe\constants.ts`,
	)
}

func alertSubscribeMatureSourceRefs(relativePaths ...string) []string {
	refs := make([]string, 0, len(relativePaths))
	for _, relativePath := range relativePaths {
		refs = append(refs, `D:\项目迁移文件\平台源码\`+relativePath)
	}
	return refs
}

func alertEventLifecycleSourceRefs() []string {
	return alertEventLifecycleMatureSourceRefs(
		`fe-main\src\pages\event\services.ts`,
		`fe-main\src\pages\event\index.tsx`,
		`fe-main\src\pages\event\Table.tsx`,
		`fe-main\src\pages\event\card.tsx`,
		`fe-main\src\pages\event\DetailNG\index.tsx`,
		`fe-main\src\pages\event\DetailNG\Actions.tsx`,
		`fe-main\src\pages\event\DetailNG\SharingLinkModal.tsx`,
		`fe-main\src\pages\event\EventNotifyRecords\services.ts`,
		`fe-main\src\pages\alertCurEvent\services.ts`,
		`fe-main\src\pages\alertCurEvent\pages\List\index.tsx`,
		`fe-main\src\pages\alertCurEvent\pages\List\AlertTable.tsx`,
		`fe-main\src\pages\alertCurEvent\utils\deleteAlertEventsModal.tsx`,
		`fe-main\src\pages\historyEvents\services.ts`,
		`fe-main\src\pages\historyEvents\ListNG\index.tsx`,
		`fe-main\src\pages\historyEvents\ListNG\DeleteEventsModal.tsx`,
		`fe-main\src\pages\historyEvents\exportEvents.ts`,
		`fe-main\src\pages\alertRules\List\EventsDrawer\index.tsx`,
		`fe-main\src\services\warning.ts`,
	)
}

func alertEventLifecycleMatureSourceRefs(relativePaths ...string) []string {
	refs := make([]string, 0, len(relativePaths))
	for _, relativePath := range relativePaths {
		refs = append(refs, `D:\项目迁移文件\平台源码\`+relativePath)
	}
	return refs
}

func warningAndRuleModalSourceRefs() []string {
	return append(warningSourceRefs(), `D:\项目迁移文件\平台源码\fe-main\src\pages\warning\subscribe\components\ruleModal.tsx`)
}

func alertEventActionPipelineQuerySourceRefs() []string {
	return alertEventActionPipelineQueryMatureSourceRefs(
		`fe-main\src\pages\event\services.ts`,
		`fe-main\src\pages\event\DetailNG\Actions.tsx`,
		`fe-main\src\pages\event\DetailNG\SharingLinkModal.tsx`,
		`fe-main\src\pages\event\DetailNG\SharedDetail.tsx`,
		`fe-main\src\pages\eventPipeline\services.ts`,
		`fe-main\src\pages\eventPipeline\pages\Form\index.tsx`,
		`fe-main\src\pages\eventPipeline\pages\Form\Processor\index.tsx`,
		`fe-main\src\pages\eventPipeline\pages\Form\TestModal\index.tsx`,
		`fe-main\src\pages\eventPipeline\pages\Form\TestModal\EventsTable.tsx`,
		`fe-main\src\pages\eventPipeline\pages\List\index.tsx`,
		`fe-main\src\pages\eventPipeline\pages\Executions\index.tsx`,
		`fe-main\src\pages\eventPipeline\pages\Executions\Detail.tsx`,
		`fe-main\src\components\AlertEventRuleTesterWithButton\index.tsx`,
		`fe-main\src\pages\alertCurEvent\services.ts`,
		`fe-main\src\services\warning.ts`,
	)
}

func alertEventActionPipelineQueryMatureSourceRefs(relativePaths ...string) []string {
	refs := make([]string, 0, len(relativePaths))
	for _, relativePath := range relativePaths {
		refs = append(refs, `D:\项目迁移文件\平台源码\`+relativePath)
	}
	return refs
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
