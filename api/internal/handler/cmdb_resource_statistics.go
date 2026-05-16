package handler

import (
	"net/http"
	"sort"
	"strings"
	"time"

	"ai-workbench-api/internal/model"
	"ai-workbench-api/internal/store"

	"github.com/gin-gonic/gin"
)

const cmdbResourceStatisticsContract = "cmdb.resource.statistics.read.v1"
const cmdbResourceStatisticsStatus = "ready_with_contract_gaps"
const cmdbContractBlockedStatus = "pending"

func GetCmdbResourceStatistics(c *gin.Context) {
	objects := store.ListCmdbObjects("")
	instanceCounts := store.CountCmdbInstancesByObject()
	targets := store.ListMonitorTargets()
	agents := store.ListFindXAgents()
	resourceGroups := store.ListResourceGroups()
	businesses := store.ListTopologyBusinesses()

	instances := listCmdbStatisticsInstances(objects, instanceCounts)
	relationIDs := map[string]struct{}{}
	boundInstanceIDs := map[string]struct{}{}
	monitorBindingRows := 0
	for _, inst := range instances {
		for _, rel := range store.ListCmdbInstanceRelations(inst.ID) {
			if strings.TrimSpace(rel.ID) != "" {
				relationIDs[rel.ID] = struct{}{}
			}
		}
		bindings := store.ListCmdbMonitorBindings(inst.ID)
		monitorBindingRows += len(bindings)
		if len(bindings) > 0 {
			boundInstanceIDs[inst.ID] = struct{}{}
		}
	}

	resp := model.CmdbResourceStatisticsResponse{
		Contract:    cmdbResourceStatisticsContract,
		Status:      cmdbResourceStatisticsStatus,
		GeneratedAt: time.Now().UTC(),
		Totals: model.CmdbResourceStatisticsTotals{
			CmdbModels:            len(objects),
			CmdbInstances:         len(instances),
			HostAssets:            len(targets),
			FindXAgents:           len(agents),
			ResourceGroups:        len(resourceGroups),
			TopologyBusinesses:    len(businesses),
			RelationEdges:         len(relationIDs),
			MonitorBindingRows:    monitorBindingRows,
			MonitorBoundInstances: len(boundInstanceIDs),
		},
		ModelDistribution:         buildCmdbModelDistribution(objects, instanceCounts),
		BusinessGroupDistribution: buildTargetLabelDistribution(targets, businessesByID(businesses), hostLabelWorkspaceID),
		ResourceGroupDistribution: buildTargetLabelDistribution(targets, resourceGroupsByID(resourceGroups), hostLabelResourceGroupID),
		BlockedContracts: []model.CmdbResourceStatisticsContractGap{
			{ID: "cmdb_resource_approval_runtime_contract", Status: cmdbContractBlockedStatus},
			{ID: "cmdb_dashboard_import_runtime_contract", Status: cmdbContractBlockedStatus},
			{ID: "cmdb_operation_risk_policy_contract", Status: cmdbContractBlockedStatus},
		},
		FindXAuditQuery: map[string]string{
			"source":        "findx_audit",
			"scope":         "cmdb",
			"resource_type": "cmdb_resource_statistics",
			"action":        "cmdb.resource_statistics.read",
		},
	}

	c.JSON(http.StatusOK, resp)
}

func listCmdbStatisticsInstances(objects []model.CmdbObject, instanceCounts map[string]int64) []model.CmdbInstance {
	out := []model.CmdbInstance{}
	for _, obj := range objects {
		total := int(instanceCounts[obj.ID])
		if total <= 0 {
			continue
		}
		rows, _ := store.ListCmdbInstances(obj.ID, 1, total)
		out = append(out, rows...)
	}
	return out
}

func buildCmdbModelDistribution(objects []model.CmdbObject, counts map[string]int64) []model.CmdbResourceStatisticsDimension {
	rows := make([]model.CmdbResourceStatisticsDimension, 0, len(objects))
	for _, obj := range objects {
		rows = append(rows, model.CmdbResourceStatisticsDimension{
			Key:   obj.ID,
			Name:  obj.Name,
			Count: int(counts[obj.ID]),
		})
	}
	sortStatisticsDimensions(rows)
	return rows
}

func buildTargetLabelDistribution(targets []*model.MonitorTarget, names map[string]string, labelKey string) []model.CmdbResourceStatisticsDimension {
	counts := map[string]int{}
	for _, target := range targets {
		if target == nil || target.Labels == nil {
			continue
		}
		key := strings.TrimSpace(target.Labels[labelKey])
		if key != "" {
			counts[key]++
		}
	}
	rows := make([]model.CmdbResourceStatisticsDimension, 0, len(counts))
	for key, count := range counts {
		rows = append(rows, model.CmdbResourceStatisticsDimension{Key: key, Name: names[key], Count: count})
	}
	sortStatisticsDimensions(rows)
	return rows
}

func businessesByID(items []model.TopologyBusiness) map[string]string {
	out := make(map[string]string, len(items))
	for _, item := range items {
		out[item.ID] = item.Name
	}
	return out
}

func resourceGroupsByID(items []model.ResourceGroup) map[string]string {
	out := make(map[string]string, len(items))
	for _, item := range items {
		out[item.ID] = item.Name
	}
	return out
}

func sortStatisticsDimensions(rows []model.CmdbResourceStatisticsDimension) {
	sort.Slice(rows, func(i, j int) bool {
		if rows[i].Count == rows[j].Count {
			return rows[i].Key < rows[j].Key
		}
		return rows[i].Count > rows[j].Count
	})
}
