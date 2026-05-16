package model

import "time"

type CmdbResourceStatisticsResponse struct {
	Contract                  string                              `json:"contract"`
	Status                    string                              `json:"status"`
	GeneratedAt               time.Time                           `json:"generated_at"`
	Totals                    CmdbResourceStatisticsTotals        `json:"totals"`
	ModelDistribution         []CmdbResourceStatisticsDimension   `json:"model_distribution"`
	BusinessGroupDistribution []CmdbResourceStatisticsDimension   `json:"business_group_distribution"`
	ResourceGroupDistribution []CmdbResourceStatisticsDimension   `json:"resource_group_distribution"`
	BlockedContracts          []CmdbResourceStatisticsContractGap `json:"blocked_contracts"`
	FindXAuditQuery           map[string]string                   `json:"findx_audit_query,omitempty"`
}

type CmdbResourceStatisticsTotals struct {
	CmdbModels            int `json:"cmdb_models"`
	CmdbInstances         int `json:"cmdb_instances"`
	HostAssets            int `json:"host_assets"`
	FindXAgents           int `json:"findx_agents"`
	ResourceGroups        int `json:"resource_groups"`
	TopologyBusinesses    int `json:"topology_businesses"`
	RelationEdges         int `json:"relation_edges"`
	MonitorBindingRows    int `json:"monitor_binding_rows"`
	MonitorBoundInstances int `json:"monitor_bound_instances"`
}

type CmdbResourceStatisticsDimension struct {
	Key   string `json:"key"`
	Name  string `json:"name,omitempty"`
	Count int    `json:"count"`
}

type CmdbResourceStatisticsContractGap struct {
	ID     string `json:"id"`
	Status string `json:"status"`
}
