package handler

import (
	"fmt"
	"net/http"
	"sort"
	"strconv"
	"strings"
	"time"

	"ai-workbench-api/internal/model"
	"ai-workbench-api/internal/store"

	"github.com/gin-gonic/gin"
	"github.com/spf13/viper"
)

func PrometheusHosts(c *gin.Context) {
	base := strings.TrimRight(viper.GetString("prometheus.url"), "/")
	if base == "" {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "prometheus.url is empty"})
		return
	}
	hostMap := map[string]*promHostSummary{}
	for _, label := range []string{"ident", "instance", "ip", "host", "hostname", "target", "address"} {
		for _, value := range promLabelValues(base, label) {
			for _, ip := range ipRe.FindAllString(value, -1) {
				if hostMap[ip] == nil {
					hostMap[ip] = &promHostSummary{IP: ip}
				}
				hostMap[ip].Labels = appendUnique(hostMap[ip].Labels, label+"="+value)
			}
		}
	}
	ips := make([]string, 0, len(hostMap))
	for ip := range hostMap {
		ips = append(ips, ip)
	}
	sort.Strings(ips)
	ensureDefaultMonitoringBusiness(ips)
	items := []promHostSummary{}
	for _, ip := range ips {
		target := discoverPromTarget(ip)
		item := *hostMap[ip]
		item.MetricCnt = len(target.Metrics)
		item.TargetOnly = target.TargetOnly
		items = append(items, item)
	}
	c.JSON(http.StatusOK, gin.H{"hosts": items})
}

func ensureDefaultMonitoringBusiness(hosts []string) {
	unassigned := unassignedPrometheusHosts(hosts)
	if len(unassigned) == 0 {
		return
	}
	const defaultBusinessName = "默认监控业务"
	business, ok := defaultMonitoringBusiness(defaultBusinessName)
	if !ok {
		business = model.TopologyBusiness{Name: defaultBusinessName, Attributes: map[string]string{"source": "prometheus_auto_discovery"}}
	}
	business.Hosts = mergeHosts(business.Hosts, unassigned)
	if business.Attributes == nil {
		business.Attributes = map[string]string{}
	}
	business.Attributes["source"] = "prometheus_auto_discovery"
	syncDefaultMonitoringBusinessGraph(&business, unassigned)
	store.SaveTopologyBusiness(business)
}

func unassignedPrometheusHosts(hosts []string) []string {
	assigned := map[string]bool{}
	for _, business := range store.ListTopologyBusinesses() {
		for _, host := range business.Hosts {
			assigned[strings.TrimSpace(host)] = true
		}
	}
	out := []string{}
	for _, host := range normalizeHosts(hosts) {
		if !assigned[host] {
			out = append(out, host)
		}
	}
	return out
}

func defaultMonitoringBusiness(name string) (model.TopologyBusiness, bool) {
	for _, business := range store.ListTopologyBusinesses() {
		if strings.TrimSpace(business.Name) == name {
			return business, true
		}
	}
	return model.TopologyBusiness{}, false
}

func businessGraphForHosts(hosts []string) model.TopologyGraph {
	now := time.Now()
	graph := model.TopologyGraph{Nodes: []model.TopologyNode{}, Edges: []model.TopologyEdge{}}
	hosts = normalizeHosts(hosts)
	for index, host := range hosts {
		addHostDiscovery(&graph, host, index, now, nil)
	}
	graph.Discovery = buildTopologyDiscoveryPlan(hosts, nil, &graph, false)
	layoutBusinessTree(&graph)
	return graph
}

func syncDefaultMonitoringBusinessGraph(business *model.TopologyBusiness, newHosts []string) {
	if len(business.Graph.Nodes) == 0 {
		business.Graph = businessGraphForHosts(business.Hosts)
		return
	}
	now := time.Now()
	for _, host := range normalizeHosts(newHosts) {
		if !hasNode(&business.Graph, "host-"+sanitizeID(host)) {
			addHostDiscovery(&business.Graph, host, len(business.Graph.Nodes), now, nil)
		}
	}
	business.Graph.Discovery = buildTopologyDiscoveryPlan(business.Hosts, business.Endpoints, &business.Graph, false)
	layoutBusinessTree(&business.Graph)
}

func PrometheusMetrics(c *gin.Context) {
	ip := strings.TrimSpace(c.Query("ip"))
	if ip == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ip is required"})
		return
	}
	target := discoverPromTarget(ip)
	if target.LabelKey == "" {
		c.JSON(http.StatusNotFound, gin.H{"error": "ip not found in prometheus", "hosts": discoveredPromHosts(50)})
		return
	}
	queries := buildCategrafQueries(target.LabelKey, target.LabelVal, target.Metrics)
	samples := runMetricQueries(queries)
	c.JSON(http.StatusOK, gin.H{"ip": ip, "target": target, "samples": samples})
}

func parseFloatValue(text string) float64 {
	fields := strings.Fields(strings.ReplaceAll(text, ";", " "))
	for _, field := range fields {
		field = strings.Trim(field, ",")
		if v, err := strconv.ParseFloat(field, 64); err == nil {
			return v
		}
		if i := strings.LastIndex(field, ":"); i >= 0 {
			if v, err := strconv.ParseFloat(field[i+1:], 64); err == nil {
				return v
			}
		}
	}
	return 0
}

func hasRecentPrometheusData(ip string) bool {
	base := strings.TrimRight(viper.GetString("prometheus.url"), "/")
	if base == "" {
		return false
	}
	query := fmt.Sprintf(`cpu_usage_active{ident=~".*%s.*"}`, ip)
	val := queryPromInstant(base, query, 0)
	return val != "" && !strings.Contains(val, "无数据")
}

func queryPromMetricValue(ip, metricName string) float64 {
	base := strings.TrimRight(viper.GetString("prometheus.url"), "/")
	if base == "" {
		return 0
	}
	target := discoverPromTarget(ip)
	if target.LabelKey == "" {
		return 0
	}
	query := fmt.Sprintf(`%s{%s="%s"}`, metricName, target.LabelKey, target.LabelVal)
	val := queryPromInstant(base, query, 0)
	if val == "" {
		return 0
	}
	return parseFloatValue(val)
}
