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
	return []model.ContractMatrixRegisterRequest{
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
			"Datasource proxy by id",
			model.ContractStatusMissingDatasource,
			[]string{`D:\项目迁移文件\平台源码\fe-main\src\services\dashboardV2.ts`, `D:\项目迁移文件\平台源码\fe-main\src\services\metricViews.ts`, `D:\项目迁移文件\平台源码\fe-main\src\services\warning.ts`},
			"Datasource proxy adapter is not available",
			monitoringContractMetadata("/query?section=metrics", "datasource_proxy", "/api/n9e/proxy/{datasource_id}"),
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
			monitoringContractMetadata("/integrations?section=templates", "dashboard_template", "/api/n9e/dashboard/{id}/migrate"),
		),
	}
}

func monitoringQueryDashboardContractSeeds() []model.ContractMatrixRegisterRequest {
	return []model.ContractMatrixRegisterRequest{
		monitoringContractSeed(
			"FX-CONTRACT-N9E-METRIC-VIEWS-CRUD",
			"Metric views CRUD",
			model.ContractStatusMissingBackend,
			[]string{`D:\项目迁移文件\平台源码\fe-main\src\services\metricViews.ts`},
			"Metric views CRUD backend contract is missing",
			monitoringContractMetadata("/query?section=metric-views", "metric_view_contract", "/api/n9e/metric-views"),
		),
		monitoringContractSeed(
			"FX-CONTRACT-N9E-METRIC-QUERY-BATCH",
			"Metric query batch",
			model.ContractStatusMissingDatasource,
			[]string{`D:\项目迁移文件\平台源码\fe-main\src\services\metric.ts`, `D:\项目迁移文件\平台源码\fe-main\src\services\metricViews.ts`, `D:\项目迁移文件\平台源码\fe-main\src\services\warning.ts`},
			"Metric query datasource contract is missing",
			monitoringContractMetadata("/query?section=metrics", "metric_query_contract", "/api/n9e/tag-metrics"),
		),
		monitoringContractSeed(
			"FX-CONTRACT-N9E-DASHBOARD-ANNOTATIONS",
			"Dashboard annotations",
			model.ContractStatusMissingBackend,
			[]string{`D:\项目迁移文件\平台源码\fe-main\src\services\dashboardV2.ts`},
			"Dashboard annotation backend contract is missing",
			monitoringContractMetadata("/dashboards?section=list", "dashboard_annotation", "/api/n9e/dashboard-annotations"),
		),
	}
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

func monitoringContractAdapterMetadata(findxRoute, adapter, upstreamRef string) map[string]string {
	return map[string]string{
		"findx_route":   findxRoute,
		"findx_adapter": adapter,
		"gap_type":      "adapter_contract",
		"upstream_ref":  upstreamRef,
	}
}
