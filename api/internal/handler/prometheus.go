package handler

import (
	"encoding/json"
	"fmt"
	"io"
	"math"
	"net/http"
	"net/url"
	"regexp"
	"strings"
	"time"

	"github.com/spf13/viper"
)

type promResult struct {
	Status string `json:"status"`
	Data   struct {
		Result []struct {
			Metric map[string]string `json:"metric"`
			Value  []interface{}     `json:"value"`
		} `json:"result"`
	} `json:"data"`
}

type MetricQuery struct {
	Name  string `mapstructure:"name"`
	Query string `mapstructure:"query"`
}

type promTarget struct {
	LabelKey   string              `json:"label_key"`
	LabelVal   string              `json:"label_val"`
	Series     []map[string]string `json:"-"`
	Metrics    []string            `json:"metrics"`
	Categories []metricCategory    `json:"categories"`
	TargetOnly bool                `json:"target_only"`
}

type metricCategory struct {
	Key         string   `json:"key"`
	Name        string   `json:"name"`
	Description string   `json:"description"`
	Metrics     []string `json:"metrics"`
}

type metricSample struct {
	Name     string `json:"name"`
	Category string `json:"category"`
	Query    string `json:"query"`
	Value    string `json:"value"`
}

type promHostSummary struct {
	IP         string   `json:"ip"`
	Labels     []string `json:"labels"`
	MetricCnt  int      `json:"metric_count"`
	TargetOnly bool     `json:"target_only"`
}

var ipRe = regexp.MustCompile(`\b(\d{1,3}(?:\.\d{1,3}){3})\b`)

var targetOnlyMetrics = map[string]bool{
	"up":                                    true,
	"scrape_duration_seconds":               true,
	"scrape_samples_scraped":                true,
	"scrape_samples_post_metric_relabeling": true,
	"scrape_series_added":                   true,
}

var categoryRules = []struct {
	Key         string
	Name        string
	Description string
	Prefixes    []string
	Exact       []string
}{
	{Key: "cpu", Name: "CPU", Description: "Categraf CPU 使用率、分态占比、核心数", Prefixes: []string{"cpu_"}},
	{Key: "memory", Name: "内存", Description: "Categraf 内存、Swap、可用率", Prefixes: []string{"mem_", "swap_"}},
	{Key: "disk", Name: "磁盘容量", Description: "Categraf 文件系统容量、inode、挂载点", Prefixes: []string{"disk_"}},
	{Key: "diskio", Name: "磁盘 IO", Description: "Categraf 磁盘吞吐、IOPS、等待、利用率", Prefixes: []string{"diskio_"}},
	{Key: "network", Name: "网络", Description: "Categraf 网卡带宽、包量、丢包、错误包、协议栈", Prefixes: []string{"net_", "netstat_", "sockstat_", "conntrack_", "ethtool_", "net_response_", "ping_"}},
	{Key: "system", Name: "系统", Description: "Categraf load、内核、进程总览、文件句柄", Prefixes: []string{"system_", "kernel_", "kernel_vmstat_", "linux_sysctl_fs_", "processes", "procstat_"}},
	{Key: "container", Name: "容器", Description: "Docker、Kubernetes、cAdvisor、Kubelet 等容器指标", Prefixes: []string{"docker_", "container_", "cadvisor_", "kubernetes_", "kube_", "kubelet_"}},
	{Key: "mysql", Name: "MySQL", Description: "MySQL 状态、InnoDB、复制、慢查询", Prefixes: []string{"mysql_"}},
	{Key: "redis", Name: "Redis", Description: "Redis 连接、内存、命中率、命令、慢日志", Prefixes: []string{"redis_"}},
	{Key: "nginx", Name: "Nginx", Description: "Nginx stub_status、VTS、upstream", Prefixes: []string{"nginx_", "nginx_vts_", "nginx_upstream_", "tengine_"}},
	{Key: "database", Name: "数据库", Description: "PostgreSQL、MongoDB、Oracle、SQLServer、ClickHouse、Greenplum", Prefixes: []string{"postgresql_", "postgres_", "mongodb_", "mongo_", "oracle_", "sqlserver_", "clickhouse_", "greenplum_"}},
	{Key: "middleware", Name: "中间件", Description: "RabbitMQ、Kafka、ZooKeeper、NATS、NSQ、RocketMQ、Elasticsearch", Prefixes: []string{"rabbitmq_", "kafka_", "zookeeper_", "nats_", "nsq_", "rocketmq_", "elasticsearch_", "logstash_"}},
	{Key: "web", Name: "Web/应用", Description: "Apache、HAProxy、Tomcat、JVM、HTTP 探测、PHP-FPM", Prefixes: []string{"apache_", "haproxy_", "tomcat_", "jvm_", "http_", "phpfpm_", "spring_"}},
	{Key: "storage", Name: "存储", Description: "NFS、SMART、Ceph/对象存储、文件计数", Prefixes: []string{"nfs_", "nfsclient_", "smart_", "filecount_", "xskyapi_"}},
	{Key: "network_device", Name: "网络设备/SNMP", Description: "SNMP、IPMI、Redfish、交换机与设备侧指标", Prefixes: []string{"snmp_", "ipmi_", "redfish_", "switch_"}},
	{Key: "cloud", Name: "云与虚拟化", Description: "CloudWatch、阿里云、Google Cloud、vSphere", Prefixes: []string{"cloudwatch_", "aliyun_", "googlecloud_", "vsphere_"}},
	{Key: "prometheus", Name: "Prometheus 抓取", Description: "Prometheus target、scrape、自监控数据", Prefixes: []string{"scrape_", "prometheus_"}, Exact: []string{"up"}},
}

func queryProm(query string) (string, error) {
	base := strings.TrimRight(viper.GetString("prometheus.url"), "/")
	if base == "" {
		return "", fmt.Errorf("prometheus.url is empty")
	}
	if val := queryPromInstant(base, query, 0); val != "" {
		return val, nil
	}
	return queryPromRange(base, query), nil
}

func queryPromInstant(base, query string, offsetSec int64) string {
	endpoint := base + "/api/v1/query?query=" + url.QueryEscape(query)
	if offsetSec != 0 {
		endpoint += "&time=" + fmt.Sprintf("%d", time.Now().Unix()-offsetSec)
	}
	resp, err := http.Get(endpoint)
	if err != nil {
		return ""
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)
	var result promResult
	if err := json.Unmarshal(body, &result); err != nil || result.Status != "success" || len(result.Data.Result) == 0 {
		return ""
	}
	return formatPromResult(result.Data.Result)
}

func queryPromRange(base, query string) string {
	now := time.Now()
	start := now.Add(-30 * 24 * time.Hour)
	step := int64(math.Ceil(now.Sub(start).Seconds() / 10000))
	if step < 60 {
		step = 60
	}
	endpoint := fmt.Sprintf("%s/api/v1/query_range?query=%s&start=%d&end=%d&step=%d", base, url.QueryEscape(query), start.Unix(), now.Unix(), step)
	resp, err := http.Get(endpoint)
	if err != nil {
		return "无数据"
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)
	var result struct {
		Status string `json:"status"`
		Data   struct {
			Result []struct {
				Metric map[string]string `json:"metric"`
				Values [][]interface{}   `json:"values"`
			} `json:"result"`
		} `json:"data"`
	}
	if err := json.Unmarshal(body, &result); err != nil || result.Status != "success" {
		return "无数据"
	}
	parts := []string{}
	for _, item := range result.Data.Result {
		if len(item.Values) == 0 {
			continue
		}
		last := item.Values[len(item.Values)-1]
		if len(last) < 2 {
			continue
		}
		value, ok := last[1].(string)
		if !ok || value == "" {
			continue
		}
		parts = append(parts, formatValueWithLabels(item.Metric, value))
	}
	if len(parts) == 0 {
		return "无数据"
	}
	return strings.Join(parts, "; ")
}

func formatPromResult(results []struct {
	Metric map[string]string `json:"metric"`
	Value  []interface{}     `json:"value"`
}) string {
	parts := []string{}
	for _, item := range results {
		if len(item.Value) < 2 {
			continue
		}
		value, ok := item.Value[1].(string)
		if !ok || value == "" {
			continue
		}
		parts = append(parts, formatValueWithLabels(item.Metric, value))
	}
	return strings.Join(parts, "; ")
}

func formatValueWithLabels(labels map[string]string, value string) string {
	keys := []string{"device", "iface", "interface", "path", "mountpoint", "state", "proto", "service", "name", "container", "pod"}
	shown := []string{}
	for _, key := range keys {
		if val := labels[key]; val != "" {
			shown = append(shown, key+"="+val)
		}
	}
	if len(shown) > 0 {
		return strings.Join(shown, ",") + ": " + value
	}
	return value
}
