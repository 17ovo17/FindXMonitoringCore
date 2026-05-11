package handler

import (
	"fmt"
	"regexp"
	"sort"
	"strings"
)

func availableMetrics(ip string) string {
	target := discoverPromTarget(ip)
	if target.LabelKey == "" {
		hosts := discoveredPromHosts(50)
		if len(hosts) == 0 {
			return "Prometheus 已连接，但没有发现任何主机标签或指标序列。"
		}
		return fmt.Sprintf("Prometheus 已全量检索 label values 与 series，未精确匹配到 IP %s。当前发现主机：%s。精确 IP 匹配不会把 198.18.20.12 误配到 198.18.20.122/123。", ip, strings.Join(hosts, ", "))
	}
	categoryText := []string{}
	for _, category := range target.Categories {
		categoryText = append(categoryText, fmt.Sprintf("%s:%d", category.Name, len(category.Metrics)))
	}
	if target.TargetOnly {
		return fmt.Sprintf("Prometheus 已发现测试目标：%s=\"%s\"，IP=%s。当前只有 up/scrape/target 序列；测试环境中 up=0/offline 仍代表目标已被发现，但不能判断主机资源压力。序列数=%d，指标=%s", target.LabelKey, target.LabelVal, ip, len(target.Series), strings.Join(target.Metrics, ", "))
	}
	return fmt.Sprintf("Prometheus 已发现 Categraf/Prometheus 性能指标：%s=\"%s\"，IP=%s。序列数=%d，指标数=%d，分类=%s，指标=%s", target.LabelKey, target.LabelVal, ip, len(target.Series), len(target.Metrics), strings.Join(categoryText, ", "), strings.Join(target.Metrics, ", "))
}

func buildMonitorContext(ip, userMsg string) string {
	_ = userMsg
	if ip == "" {
		return ""
	}
	target := discoverPromTarget(ip)
	if target.LabelKey == "" {
		return fmt.Sprintf("\n\n[Prometheus 全量发现结果 - 主机: %s]\n- 数据源: prometheus\n- 解析结果: 未精确匹配到该 IP 的 ident/instance/ip/host/hostname/target/address 标签，也未在全量 series 标签值中匹配到该 IP。\n- 已发现主机: %s\n- 说明: 已使用精确 IP 匹配，不会把 198.18.20.12 误配到 198.18.20.122/123。\n[数据结束]\n", ip, strings.Join(discoveredPromHosts(50), ", "))
	}
	queries := buildCategrafQueries(target.LabelKey, target.LabelVal, target.Metrics)
	samples := runMetricQueries(queries)
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("\n\n[Prometheus 全量发现数据 - 主机: %s，匹配标签: %s=%s]\n", ip, target.LabelKey, target.LabelVal))
	sb.WriteString("- 数据源: prometheus\n")
	if target.TargetOnly {
		sb.WriteString("- 发现类型: target/scrape 数据。测试环境中 up=0/offline 仍代表该目标已被发现；在线状态只作为健康字段，不作为发现失败标准。\n")
	} else {
		sb.WriteString("- 发现类型: Categraf/Prometheus 性能指标。已动态适配 Categraf 的主机、网络、磁盘、容器、中间件与应用指标前缀。\n")
	}
	sb.WriteString(fmt.Sprintf("- 目标解析: 发现 %d 条序列、%d 个唯一指标。\n", len(target.Series), len(target.Metrics)))
	for _, category := range target.Categories {
		sb.WriteString(fmt.Sprintf("- 指标分类: %s，数量=%d，说明=%s\n", category.Name, len(category.Metrics), category.Description))
		sb.WriteString(fmt.Sprintf("  指标: %s\n", strings.Join(category.Metrics, ", ")))
	}
	if len(samples) == 0 {
		sb.WriteString("- 实际数值: 未查询到可用最新值或历史值。\n")
	} else {
		for _, sample := range samples {
			sb.WriteString(fmt.Sprintf("- %s/%s: %s\n", sample.Category, sample.Name, sample.Value))
		}
	}
	sb.WriteString("[数据结束。若只有 target/scrape 数据，只能判断目标被发现和抓取健康，不能判断主机资源压力；不要把 up=0 当作测试失败。]\n")
	return sb.String()
}

func buildCategrafQueries(labelKey, labelVal string, metrics []string) []MetricQuery {
	selector := fmt.Sprintf(`%s="%s"`, labelKey, labelVal)
	queries := []MetricQuery{}
	seen := map[string]bool{}
	add := func(title, query string) {
		if seen[query] {
			return
		}
		seen[query] = true
		queries = append(queries, MetricQuery{Name: title, Query: query})
	}
	metricSet := map[string]bool{}
	for _, metric := range metrics {
		metricSet[metric] = true
	}
	for _, spec := range priorityMetricSpecs() {
		if metricSet[spec.Metric] {
			add(spec.Title, fmt.Sprintf(spec.Query, selector))
		}
	}
	for _, metric := range metrics {
		if seen[metric] {
			continue
		}
		add(friendlyMetricName(metric), fmt.Sprintf(`%s{%s}`, metric, selector))
	}
	return queries
}

func runMetricQueries(queries []MetricQuery) []metricSample {
	samples := []metricSample{}
	for _, query := range queries {
		value, err := queryProm(query.Query)
		if err != nil || strings.TrimSpace(value) == "" || value == "无数据" {
			continue
		}
		samples = append(samples, metricSample{Name: query.Name, Category: categoryNameForMetric(metricNameFromQuery(query.Query)), Query: query.Query, Value: value})
	}
	return samples
}

func metricNameFromQuery(query string) string {
	cleaned := strings.TrimSpace(query)
	for _, fn := range []string{"rate", "irate", "sum", "avg", "max", "min", "increase"} {
		cleaned = strings.TrimPrefix(cleaned, fn+"(")
	}
	match := regexp.MustCompile(`[a-zA-Z_:][a-zA-Z0-9_:]*`).FindString(cleaned)
	return match
}

func friendlyMetricName(name string) string {
	return categoryNameForMetric(name) + "/" + name
}

func categoryNameForMetric(metric string) string {
	for _, rule := range categoryRules {
		for _, exact := range rule.Exact {
			if metric == exact {
				return rule.Name
			}
		}
		for _, prefix := range rule.Prefixes {
			if strings.HasPrefix(metric, prefix) {
				return rule.Name
			}
		}
	}
	return "其他"
}

func categorizeMetrics(metrics []string) []metricCategory {
	byKey := map[string]*metricCategory{}
	order := []string{}
	for _, metric := range metrics {
		key, name, desc := categoryForMetric(metric)
		if byKey[key] == nil {
			byKey[key] = &metricCategory{Key: key, Name: name, Description: desc}
			order = append(order, key)
		}
		byKey[key].Metrics = append(byKey[key].Metrics, metric)
	}
	for _, category := range byKey {
		sort.Strings(category.Metrics)
	}
	out := make([]metricCategory, 0, len(order))
	for _, key := range order {
		out = append(out, *byKey[key])
	}
	return out
}

func categoryForMetric(metric string) (string, string, string) {
	for _, rule := range categoryRules {
		for _, exact := range rule.Exact {
			if metric == exact {
				return rule.Key, rule.Name, rule.Description
			}
		}
		for _, prefix := range rule.Prefixes {
			if strings.HasPrefix(metric, prefix) {
				return rule.Key, rule.Name, rule.Description
			}
		}
	}
	return "other", "其他", "未识别前缀，仍按原始 Prometheus 指标纳入适配"
}

type priorityMetricSpec struct {
	Title  string
	Metric string
	Query  string
}

func priorityMetricSpecs() []priorityMetricSpec {
	specs := append(priorityHostMetricSpecs(), priorityNetworkMetricSpecs()...)
	return append(specs, priorityServiceRuntimeMetricSpecs()...)
}

func priorityHostMetricSpecs() []priorityMetricSpec {
	return []priorityMetricSpec{
		{Title: "CPU usage active(%)", Metric: "cpu_usage_active", Query: `cpu_usage_active{%s}`},
		{Title: "CPU usage idle(%)", Metric: "cpu_usage_idle", Query: `cpu_usage_idle{%s}`},
		{Title: "CPU iowait(%)", Metric: "cpu_usage_iowait", Query: `cpu_usage_iowait{%s}`},
		{Title: "CPU steal(%)", Metric: "cpu_usage_steal", Query: `cpu_usage_steal{%s}`},
		{Title: "Memory used(%)", Metric: "mem_used_percent", Query: `mem_used_percent{%s}`},
		{Title: "Memory available(%)", Metric: "mem_available_percent", Query: `mem_available_percent{%s}`},
		{Title: "Swap used(%)", Metric: "swap_used_percent", Query: `swap_used_percent{%s}`},
		{Title: "Disk used(%)", Metric: "disk_used_percent", Query: `disk_used_percent{%s}`},
		{Title: "Disk read bytes/s", Metric: "diskio_read_bytes", Query: `rate(diskio_read_bytes{%s}[5m])`},
		{Title: "Disk write bytes/s", Metric: "diskio_write_bytes", Query: `rate(diskio_write_bytes{%s}[5m])`},
		{Title: "Disk IO utilization(%)", Metric: "diskio_io_util", Query: `diskio_io_util{%s}`},
		{Title: "Disk IO await(ms)", Metric: "diskio_io_await", Query: `diskio_io_await{%s}`},
		{Title: "System load 1m", Metric: "system_load1", Query: `system_load1{%s}`},
		{Title: "System load 5m", Metric: "system_load5", Query: `system_load5{%s}`},
		{Title: "System load 15m", Metric: "system_load15", Query: `system_load15{%s}`},
		{Title: "CPU cores", Metric: "system_n_cpus", Query: `system_n_cpus{%s}`},
		{Title: "Process count", Metric: "processes", Query: `processes{%s}`},
		{Title: "Zombie processes", Metric: "processes_zombies", Query: `processes_zombies{%s}`},
	}
}

func priorityNetworkMetricSpecs() []priorityMetricSpec {
	return []priorityMetricSpec{
		{Title: "Network receive bits/s", Metric: "net_bits_recv", Query: `net_bits_recv{%s}`},
		{Title: "Network send bits/s", Metric: "net_bits_sent", Query: `net_bits_sent{%s}`},
		{Title: "Network receive packets", Metric: "net_packets_recv", Query: `net_packets_recv{%s}`},
		{Title: "Network send packets", Metric: "net_packets_sent", Query: `net_packets_sent{%s}`},
		{Title: "Network inbound drops", Metric: "net_drop_in", Query: `net_drop_in{%s}`},
		{Title: "Network outbound drops", Metric: "net_drop_out", Query: `net_drop_out{%s}`},
		{Title: "Network inbound errors", Metric: "net_err_in", Query: `net_err_in{%s}`},
		{Title: "Network outbound errors", Metric: "net_err_out", Query: `net_err_out{%s}`},
		{Title: "TCP inuse", Metric: "netstat_tcp_inuse", Query: `netstat_tcp_inuse{%s}`},
		{Title: "TCP TIME_WAIT", Metric: "netstat_tcp_tw", Query: `netstat_tcp_tw{%s}`},
		{Title: "Socket used", Metric: "netstat_sockets_used", Query: `netstat_sockets_used{%s}`},
	}
}

func priorityServiceRuntimeMetricSpecs() []priorityMetricSpec {
	return []priorityMetricSpec{
		{Title: "MySQL up", Metric: "mysql_up", Query: `mysql_up{%s}`},
		{Title: "MySQL connections", Metric: "mysql_global_status_threads_connected", Query: `mysql_global_status_threads_connected{%s}`},
		{Title: "MySQL QPS", Metric: "mysql_global_status_queries", Query: `rate(mysql_global_status_queries{%s}[5m])`},
		{Title: "Redis up", Metric: "redis_up", Query: `redis_up{%s}`},
		{Title: "Redis connected clients", Metric: "redis_connected_clients", Query: `redis_connected_clients{%s}`},
		{Title: "Redis OPS", Metric: "redis_instantaneous_ops_per_sec", Query: `redis_instantaneous_ops_per_sec{%s}`},
		{Title: "Nginx up", Metric: "nginx_up", Query: `nginx_up{%s}`},
		{Title: "Nginx active connections", Metric: "nginx_active", Query: `nginx_active{%s}`},
		{Title: "Nginx request rate", Metric: "nginx_requests", Query: `rate(nginx_requests{%s}[5m])`},
		{Title: "Docker container count", Metric: "docker_n_containers", Query: `docker_n_containers{%s}`},
		{Title: "JVM GC max pause(s)", Metric: "jvm_gc_pause_seconds_max", Query: `jvm_gc_pause_seconds_max{%s}`},
		{Title: "JVM GC pause sum(s)", Metric: "jvm_gc_pause_seconds_sum", Query: `jvm_gc_pause_seconds_sum{%s}`},
		{Title: "JVM GC pause count", Metric: "jvm_gc_pause_seconds_count", Query: `jvm_gc_pause_seconds_count{%s}`},
		{Title: "JVM memory used(bytes)", Metric: "jvm_memory_used_bytes", Query: `jvm_memory_used_bytes{%s}`},
		{Title: "JVM memory max(bytes)", Metric: "jvm_memory_max_bytes", Query: `jvm_memory_max_bytes{%s}`},
		{Title: "JVM live threads", Metric: "jvm_threads_live_threads", Query: `jvm_threads_live_threads{%s}`},
		{Title: "JVM loaded classes", Metric: "jvm_classes_loaded_classes", Query: `jvm_classes_loaded_classes{%s}`},
		{Title: "Process CPU usage", Metric: "process_cpu_usage", Query: `process_cpu_usage{%s}`},
	}
}
