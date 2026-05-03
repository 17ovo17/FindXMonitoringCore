package handler

import (
	"fmt"
	"math"
	"net/http"
	"sort"
	"strings"

	"ai-workbench-api/internal/model"
	"ai-workbench-api/internal/store"

	"github.com/gin-gonic/gin"
)

func ListMonitorMetrics(c *gin.Context) {
	dsID := normalizeMonitoringDatasourceID(c.Query("datasource_id"))
	if _, _, ok := resolveMonitoringPrometheus(c, dsID); !ok {
		return
	}
	page := boundedPositiveInt(c.DefaultQuery("page", "1"), 1, math.MaxInt32)
	limit := boundedPositiveInt(c.DefaultQuery("limit", "50"), 50, 500)
	filtered := filterMonitorMetrics(listAllMetricsMappings(dsID, c.Query("status")), c)
	sort.Slice(filtered, func(i, j int) bool { return fmt.Sprint(filtered[i]["raw_name"]) < fmt.Sprint(filtered[j]["raw_name"]) })
	start, end := (page-1)*limit, page*limit
	if start > len(filtered) {
		start = len(filtered)
	}
	if end > len(filtered) {
		end = len(filtered)
	}
	c.JSON(http.StatusOK, gin.H{"datasource_id": dsID, "items": filtered[start:end], "total": len(filtered), "page": page, "limit": limit})
}

func filterMonitorMetrics(items []model.MetricsMapping, c *gin.Context) []gin.H {
	q := strings.ToLower(strings.TrimSpace(c.Query("q")))
	category, exporter := strings.TrimSpace(c.Query("category")), strings.TrimSpace(c.Query("exporter"))
	out := make([]gin.H, 0, len(items))
	for _, mapping := range items {
		item := metricMappingResponse(mapping)
		haystack := strings.ToLower(strings.Join([]string{
			fmt.Sprint(item["raw_name"]), fmt.Sprint(item["standard_name"]),
			fmt.Sprint(item["description"]), fmt.Sprint(item["exporter"]),
		}, " "))
		if (q == "" || strings.Contains(haystack, q)) && (category == "" || item["category"] == category) && (exporter == "" || item["exporter"] == exporter) {
			out = append(out, item)
		}
	}
	return out
}

func metricMappingResponse(m model.MetricsMapping) gin.H {
	category, categoryName, _ := categoryForMetric(m.RawName)
	exporter := firstNonEmpty(m.Exporter, monitoringGuessExporter(m.RawName))
	promql := strings.TrimSpace(m.Transform)
	if promql == "" || strings.EqualFold(promql, "none") || promql == "{}" {
		promql = defaultMappedPromQL(m.RawName, `instance=~".*"`)
	}
	return gin.H{
		"id": m.ID, "datasource_id": m.DatasourceID, "raw_name": m.RawName,
		"standard_name": m.StandardName, "exporter": exporter, "description": m.Description,
		"transform": m.Transform, "status": m.Status, "category": category,
		"category_name": categoryName, "promql": promql,
	}
}

func listAllMetricsMappings(datasourceID, status string) []model.MetricsMapping {
	out := []model.MetricsMapping{}
	for page := 1; ; page++ {
		items, _ := store.ListMetricsMappings(datasourceID, strings.TrimSpace(status), page, 500)
		out = append(out, items...)
		if len(items) < 500 {
			return out
		}
	}
}

func monitoringGuessExporter(name string) string {
	for _, rule := range []struct{ prefix, exporter string }{
		{"node_", "node_exporter"}, {"process_", "process_exporter"},
		{"mysql_", "mysqld_exporter"}, {"redis_", "redis_exporter"}, {"go_", "go_runtime"},
	} {
		if strings.HasPrefix(name, rule.prefix) {
			return rule.exporter
		}
	}
	if strings.HasPrefix(name, "container_") || strings.HasPrefix(name, "kube_") {
		return "kubernetes"
	}
	return ""
}
