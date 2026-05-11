package handler

import (
	"context"
	"encoding/json"
	"fmt"
	"math"
	"net/http"
	"strings"
	"time"

	"ai-workbench-api/internal/model"
	"ai-workbench-api/internal/monitoring"
	"ai-workbench-api/internal/store"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

const (
	discoveryObjectID  = "obj-os"
	discoveryTimeout   = 10 * time.Second
	prometheusFallback = "http://localhost:9090"
)

// CmdbDiscover 自动发现：从 Prometheus node_exporter 指标创建/更新 CMDB 实例
func CmdbDiscover(c *gin.Context) {
	promURL := resolveDiscoveryPrometheusURL()
	if promURL == "" {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "prometheus URL not configured"})
		return
	}

	ctx := c.Request.Context()
	gw := monitoring.NewPrometheusGateway(nil)

	// 1. 查询 node_uname_info 获取所有主机
	hosts, err := queryNodeUnameInfo(ctx, gw, promURL)
	if err != nil {
		logrus.WithError(err).Error("cmdb discovery: query node_uname_info failed")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "query prometheus failed: " + err.Error()})
		return
	}

	if len(hosts) == 0 {
		c.JSON(http.StatusOK, gin.H{"discovered": 0, "created": 0, "updated": 0, "errors": []string{}})
		return
	}

	// 2. 获取现有实例，按 IP 索引
	existingByIP := indexInstancesByIP(discoveryObjectID)

	var created, updated int
	var errors []string

	for _, host := range hosts {
		ip := extractIPFromPromInstance(host.Instance)
		if ip == "" {
			errors = append(errors, fmt.Sprintf("skip host %s: cannot extract IP", host.Instance))
			continue
		}

		// 查询额外指标
		extra := queryHostMetrics(ctx, gw, promURL, host.Instance)

		// 构建 data JSON
		data := buildDiscoveryData(host, extra, ip)
		dataJSON, _ := json.Marshal(data)

		if existing, ok := existingByIP[ip]; ok {
			// 更新已有实例（仅更新 discovery 字段）
			mergedData := mergeDiscoveryData(existing.Data, data)
			mergedJSON, _ := json.Marshal(mergedData)
			existing.Data = string(mergedJSON)
			existing.Updater = "auto-discovery"
			if err := store.UpdateCmdbInstance(existing); err != nil {
				errors = append(errors, fmt.Sprintf("update %s failed: %v", ip, err))
				continue
			}
			updated++
		} else {
			// 创建新实例
			inst := &model.CmdbInstance{
				ObjectID: discoveryObjectID,
				Data:     string(dataJSON),
				Creator:  "auto-discovery",
				Updater:  "auto-discovery",
			}
			if err := store.CreateCmdbInstance(inst); err != nil {
				errors = append(errors, fmt.Sprintf("create %s failed: %v", ip, err))
				continue
			}
			created++
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"discovered": len(hosts),
		"created":    created,
		"updated":    updated,
		"errors":     errors,
	})
}

// --- Helper types and functions ---

type nodeUnameHost struct {
	Instance string
	Nodename string
	Sysname  string
	Release  string
	Machine  string
}

type hostExtraMetrics struct {
	MemoryTotal float64
	CPUCores    float64
	DiskTotal   float64
	UptimeDays  float64
}

func resolveDiscoveryPrometheusURL() string {
	_, base, err := monitoring.ResolvePrometheusDatasourceFromConfig("")
	if err == nil && base != "" {
		return base
	}
	return prometheusFallback
}

func queryNodeUnameInfo(ctx context.Context, gw *monitoring.PrometheusGateway, promURL string) ([]nodeUnameHost, error) {
	result, err := gw.QueryInstant(ctx, monitoring.PrometheusQueryRequest{
		BaseURL: promURL,
		Query:   "node_uname_info",
		Timeout: discoveryTimeout,
	})
	if err != nil {
		return nil, err
	}

	rows := extractVectorResults(result.Data)
	var hosts []nodeUnameHost
	for _, row := range rows {
		metric, _ := row["metric"].(map[string]any)
		if metric == nil {
			continue
		}
		hosts = append(hosts, nodeUnameHost{
			Instance: getString(metric, "instance"),
			Nodename: getString(metric, "nodename"),
			Sysname:  getString(metric, "sysname"),
			Release:  getString(metric, "release"),
			Machine:  getString(metric, "machine"),
		})
	}
	return hosts, nil
}

func queryHostMetrics(ctx context.Context, gw *monitoring.PrometheusGateway, promURL, instance string) hostExtraMetrics {
	var extra hostExtraMetrics

	// memory
	if val, err := queryScalar(ctx, gw, promURL, fmt.Sprintf(`node_memory_MemTotal_bytes{instance="%s"}`, instance)); err == nil {
		extra.MemoryTotal = val
	}

	// cpu cores
	if val, err := queryScalar(ctx, gw, promURL, fmt.Sprintf(`count(node_cpu_seconds_total{instance="%s",mode="idle"})`, instance)); err == nil {
		extra.CPUCores = val
	}

	// disk total (root mountpoint)
	if val, err := queryScalar(ctx, gw, promURL, fmt.Sprintf(`node_filesystem_size_bytes{instance="%s",mountpoint="/"}`, instance)); err == nil {
		extra.DiskTotal = val
	}

	// uptime days
	if val, err := queryScalar(ctx, gw, promURL, fmt.Sprintf(`(node_time_seconds{instance="%s"} - node_boot_time_seconds{instance="%s"}) / 86400`, instance, instance)); err == nil {
		extra.UptimeDays = val
	}

	return extra
}

func queryScalar(ctx context.Context, gw *monitoring.PrometheusGateway, promURL, query string) (float64, error) {
	result, err := gw.QueryInstant(ctx, monitoring.PrometheusQueryRequest{
		BaseURL: promURL,
		Query:   query,
		Timeout: discoveryTimeout,
	})
	if err != nil {
		return 0, err
	}
	rows := extractVectorResults(result.Data)
	if len(rows) == 0 {
		return 0, fmt.Errorf("no data")
	}
	return extractValue(rows[0]), nil
}

func extractVectorResults(data map[string]any) []map[string]any {
	if data == nil {
		return nil
	}
	result, _ := data["result"].([]any)
	var out []map[string]any
	for _, item := range result {
		if row, ok := item.(map[string]any); ok {
			out = append(out, row)
		}
	}
	return out
}

func extractValue(row map[string]any) float64 {
	// instant query: "value": [timestamp, "value_string"]
	value, _ := row["value"].([]any)
	if len(value) >= 2 {
		if s, ok := value[1].(string); ok {
			var f float64
			fmt.Sscanf(s, "%f", &f)
			return f
		}
	}
	return 0
}

func getString(m map[string]any, key string) string {
	v, _ := m[key].(string)
	return v
}

func extractIPFromPromInstance(instance string) string {
	// instance format: "ip:port" or just "ip"
	parts := strings.Split(instance, ":")
	if len(parts) >= 1 && parts[0] != "" {
		return parts[0]
	}
	return ""
}

func indexInstancesByIP(objectID string) map[string]*model.CmdbInstance {
	// Load all instances for the OS object (paginated, get all)
	result := make(map[string]*model.CmdbInstance)
	page := 1
	for {
		items, total := store.ListCmdbInstances(objectID, page, 100)
		for i := range items {
			ip := extractIPFromInstanceData(items[i].Data)
			if ip != "" {
				inst := items[i]
				result[ip] = &inst
			}
		}
		if int64(page*100) >= total {
			break
		}
		page++
	}
	return result
}

func extractIPFromInstanceData(dataJSON string) string {
	if dataJSON == "" {
		return ""
	}
	var data map[string]any
	if err := json.Unmarshal([]byte(dataJSON), &data); err != nil {
		return ""
	}
	ip, _ := data["ip_address"].(string)
	return ip
}

func buildDiscoveryData(host nodeUnameHost, extra hostExtraMetrics, ip string) map[string]any {
	uname := strings.TrimSpace(host.Sysname + " " + host.Nodename + " " + host.Release)
	data := map[string]any{
		"name":         host.Nodename,
		"ip_address":   ip,
		"os_version":   host.Release,
		"uname":        uname,
		"agent_status": "在线",
	}
	if host.Machine != "" {
		data["arch"] = host.Machine
	}
	if extra.MemoryTotal > 0 {
		data["memory_total"] = int64(extra.MemoryTotal)
	}
	if extra.CPUCores > 0 {
		data["cpu_cores"] = int(extra.CPUCores)
	}
	if extra.DiskTotal > 0 {
		data["disk_total"] = int64(extra.DiskTotal)
	}
	if extra.UptimeDays > 0 {
		data["uptime_days"] = int(math.Round(extra.UptimeDays))
	}
	data["discovery_time"] = time.Now().Format(time.RFC3339)
	return data
}

func mergeDiscoveryData(existingJSON string, discovered map[string]any) map[string]any {
	existing := make(map[string]any)
	if existingJSON != "" {
		json.Unmarshal([]byte(existingJSON), &existing)
	}
	// Only overwrite discovery-flagged fields
	discoveryFields := []string{
		"name", "ip_address", "os_version", "uname", "arch",
		"memory_total", "cpu_cores", "disk_total", "uptime_days",
		"agent_status", "discovery_time",
	}
	for _, field := range discoveryFields {
		if val, ok := discovered[field]; ok {
			existing[field] = val
		}
	}
	return existing
}
