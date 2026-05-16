package handler

import "ai-workbench-api/internal/model"

// integrationTemplates 内置 15 个 P0 集成模板
var integrationTemplates = []model.IntegrationTemplate{
	mysqlTemplate(),
	redisTemplate(),
	nginxTemplate(),
	linuxTemplate(),
	dockerTemplate(),
	postgresqlTemplate(),
	elasticsearchTemplate(),
	kafkaTemplate(),
	mongodbTemplate(),
	kubernetesTemplate(),
	httpResponseTemplate(),
	pingTemplate(),
	netResponseTemplate(),
	prometheusTemplate(),
	javaTemplate(),
}

// integrationTemplateMap 按 ID 索引
var integrationTemplateMap map[string]model.IntegrationTemplate

func init() {
	integrationTemplateMap = make(map[string]model.IntegrationTemplate, len(integrationTemplates))
	for _, t := range integrationTemplates {
		integrationTemplateMap[t.ID] = t
	}
}

func mysqlTemplate() model.IntegrationTemplate {
	return model.IntegrationTemplate{
		ID:          "mysql",
		Name:        "MySQL",
		Description: "监控 MySQL 数据库性能指标，包括连接数、查询吞吐、慢查询、复制延迟等",
		Category:    "database",
		Params: []model.TemplateParam{
			{Key: "address", Label: "数据库地址", Type: "string", Required: true, Default: "127.0.0.1:3306", Placeholder: "host:port"},
			{Key: "username", Label: "用户名", Type: "string", Required: true, Default: "monitor"},
			{Key: "password", Label: "密码", Type: "password", Required: true},
			{Key: "extra_status_metrics", Label: "额外状态指标", Type: "bool", Default: "true"},
			{Key: "extra_innodb_metrics", Label: "InnoDB 指标", Type: "bool", Default: "true"},
			{Key: "gather_slave_status", Label: "采集从库状态", Type: "bool", Default: "false"},
			{Key: "interval", Label: "采集间隔", Type: "string", Default: "15s"},
		},
		TomlTemplate: `[[instances]]
address = "{{.username}}:{{.password}}@tcp({{.address}})/"
interval_times = 1
labels = { instance="{{.address}}" }

extra_status_metrics = {{.extra_status_metrics}}
extra_innodb_metrics = {{.extra_innodb_metrics}}
gather_slave_status = {{.gather_slave_status}}

interval = "{{.interval}}"
`,
	}
}

func redisTemplate() model.IntegrationTemplate {
	return model.IntegrationTemplate{
		ID:          "redis",
		Name:        "Redis",
		Description: "监控 Redis 实例性能指标，包括内存使用、连接数、QPS、持久化状态等",
		Category:    "database",
		Params: []model.TemplateParam{
			{Key: "address", Label: "Redis 地址", Type: "string", Required: true, Default: "127.0.0.1:6379", Placeholder: "host:port"},
			{Key: "password", Label: "密码", Type: "password", Required: false},
			{Key: "username", Label: "用户名(Redis 6+)", Type: "string", Required: false, Default: ""},
			{Key: "interval", Label: "采集间隔", Type: "string", Default: "15s"},
		},
		TomlTemplate: `[[instances]]
targets = ["{{.address}}"]
{{- if .password}}
password = "{{.password}}"
{{- end}}
{{- if .username}}
username = "{{.username}}"
{{- end}}

labels = { instance="{{.address}}" }
interval = "{{.interval}}"
`,
	}
}

func nginxTemplate() model.IntegrationTemplate {
	return model.IntegrationTemplate{
		ID:          "nginx",
		Name:        "Nginx",
		Description: "监控 Nginx 状态指标，包括活跃连接数、请求速率、连接状态分布等",
		Category:    "webserver",
		Params: []model.TemplateParam{
			{Key: "url", Label: "Status URL", Type: "string", Required: true, Default: "http://127.0.0.1/nginx_status", Placeholder: "http://host/nginx_status"},
			{Key: "response_timeout", Label: "响应超时", Type: "string", Default: "5s"},
			{Key: "interval", Label: "采集间隔", Type: "string", Default: "15s"},
		},
		TomlTemplate: `[[instances]]
urls = ["{{.url}}"]
response_timeout = "{{.response_timeout}}"
labels = { instance="{{.url}}" }
interval = "{{.interval}}"
`,
	}
}

func linuxTemplate() model.IntegrationTemplate {
	return model.IntegrationTemplate{
		ID:          "linux",
		Name:        "Linux",
		Description: "监控 Linux 主机基础指标，包括 CPU、内存、磁盘、网络、进程等",
		Category:    "os",
		Params: []model.TemplateParam{
			{Key: "interval", Label: "采集间隔", Type: "string", Default: "15s"},
			{Key: "collect_cpu", Label: "采集 CPU", Type: "bool", Default: "true"},
			{Key: "collect_mem", Label: "采集内存", Type: "bool", Default: "true"},
			{Key: "collect_disk", Label: "采集磁盘", Type: "bool", Default: "true"},
			{Key: "collect_net", Label: "采集网络", Type: "bool", Default: "true"},
		},
		TomlTemplate: `# Linux 基础监控由 categraf 内置插件自动采集
# 以下配置控制各子插件的开关

[cpu]
collect = {{.collect_cpu}}
interval = "{{.interval}}"

[mem]
collect = {{.collect_mem}}
interval = "{{.interval}}"

[disk]
collect = {{.collect_disk}}
interval = "{{.interval}}"

[net]
collect = {{.collect_net}}
interval = "{{.interval}}"
`,
	}
}

func dockerTemplate() model.IntegrationTemplate {
	return model.IntegrationTemplate{
		ID:          "docker",
		Name:        "Docker",
		Description: "监控 Docker 容器状态、资源使用率、重启检测等",
		Category:    "container",
		Params: []model.TemplateParam{
			{Key: "endpoint", Label: "Docker Endpoint", Type: "string", Required: true, Default: "unix:///var/run/docker.sock"},
			{Key: "gather_services", Label: "采集 Services", Type: "bool", Default: "false"},
			{Key: "container_names", Label: "容器名过滤(逗号分隔)", Type: "string", Default: "*", Placeholder: "*,nginx,redis"},
			{Key: "interval", Label: "采集间隔", Type: "string", Default: "15s"},
		},
		TomlTemplate: `[[instances]]
endpoint = "{{.endpoint}}"
gather_services = {{.gather_services}}

container_names = [{{range $i, $v := split .container_names ","}}{{if $i}}, {{end}}"{{trim $v}}"{{end}}]

labels = { endpoint="{{.endpoint}}" }
interval = "{{.interval}}"
`,
	}
}

func postgresqlTemplate() model.IntegrationTemplate {
	return model.IntegrationTemplate{
		ID:          "postgresql",
		Name:        "PostgreSQL",
		Description: "监控 PostgreSQL 数据库性能指标，包括连接数、事务、锁、复制延迟等",
		Category:    "database",
		Params: []model.TemplateParam{
			{Key: "address", Label: "连接地址", Type: "string", Required: true, Default: "host=127.0.0.1 port=5432 sslmode=disable", Placeholder: "host=IP port=5432 user=X password=Y dbname=Z sslmode=disable"},
			{Key: "username", Label: "用户名", Type: "string", Required: true, Default: "monitor"},
			{Key: "password", Label: "密码", Type: "password", Required: true},
			{Key: "dbname", Label: "数据库名", Type: "string", Required: true, Default: "postgres"},
			{Key: "interval", Label: "采集间隔", Type: "string", Default: "15s"},
		},
		TomlTemplate: `[[instances]]
address = "host={{index_field .address "host"}} port={{index_field .address "port"}} user={{.username}} password={{.password}} dbname={{.dbname}} sslmode=disable"
labels = { instance="{{.address}}" }
interval = "{{.interval}}"

ignored_databases = ["template0", "template1"]
`,
	}
}

func elasticsearchTemplate() model.IntegrationTemplate {
	return model.IntegrationTemplate{
		ID:          "elasticsearch",
		Name:        "Elasticsearch",
		Description: "监控 Elasticsearch 集群健康、节点状态、索引性能等",
		Category:    "database",
		Params: []model.TemplateParam{
			{Key: "servers", Label: "ES 节点地址", Type: "string", Required: true, Default: "http://127.0.0.1:9200", Placeholder: "http://host:9200"},
			{Key: "username", Label: "用户名", Type: "string", Required: false},
			{Key: "password", Label: "密码", Type: "password", Required: false},
			{Key: "cluster_health", Label: "采集集群健康", Type: "bool", Default: "true"},
			{Key: "cluster_stats", Label: "采集集群统计", Type: "bool", Default: "true"},
			{Key: "interval", Label: "采集间隔", Type: "string", Default: "30s"},
		},
		TomlTemplate: `[[instances]]
servers = ["{{.servers}}"]
{{- if .username}}
username = "{{.username}}"
password = "{{.password}}"
{{- end}}
cluster_health = {{.cluster_health}}
cluster_stats = {{.cluster_stats}}
labels = { instance="{{.servers}}" }
interval = "{{.interval}}"
`,
	}
}

func kafkaTemplate() model.IntegrationTemplate {
	return model.IntegrationTemplate{
		ID:          "kafka",
		Name:        "Kafka",
		Description: "监控 Kafka 集群指标，包括 Broker 状态、Topic 吞吐、消费者 Lag 等",
		Category:    "middleware",
		Params: []model.TemplateParam{
			{Key: "brokers", Label: "Broker 地址", Type: "string", Required: true, Default: "127.0.0.1:9092", Placeholder: "host1:9092,host2:9092"},
			{Key: "topics", Label: "监控 Topic(逗号分隔,空=全部)", Type: "string", Default: ""},
			{Key: "consumer_groups", Label: "消费组(逗号分隔,空=全部)", Type: "string", Default: ""},
			{Key: "sasl_username", Label: "SASL 用户名", Type: "string", Required: false},
			{Key: "sasl_password", Label: "SASL 密码", Type: "password", Required: false},
			{Key: "interval", Label: "采集间隔", Type: "string", Default: "30s"},
		},
		TomlTemplate: `[[instances]]
brokers = [{{range $i, $v := split .brokers ","}}{{if $i}}, {{end}}"{{trim $v}}"{{end}}]
{{- if .topics}}
topics = [{{range $i, $v := split .topics ","}}{{if $i}}, {{end}}"{{trim $v}}"{{end}}]
{{- end}}
{{- if .consumer_groups}}
consumer_groups = [{{range $i, $v := split .consumer_groups ","}}{{if $i}}, {{end}}"{{trim $v}}"{{end}}]
{{- end}}
{{- if .sasl_username}}
sasl_username = "{{.sasl_username}}"
sasl_password = "{{.sasl_password}}"
sasl_mechanism = "PLAIN"
{{- end}}
labels = { instance="{{.brokers}}" }
interval = "{{.interval}}"
`,
	}
}

func mongodbTemplate() model.IntegrationTemplate {
	return model.IntegrationTemplate{
		ID:          "mongodb",
		Name:        "MongoDB",
		Description: "监控 MongoDB 实例指标，包括连接数、操作计数、复制集状态、WiredTiger 引擎等",
		Category:    "database",
		Params: []model.TemplateParam{
			{Key: "servers", Label: "MongoDB URI", Type: "string", Required: true, Default: "mongodb://127.0.0.1:27017", Placeholder: "mongodb://user:pass@host:27017"},
			{Key: "gather_repl_set_status", Label: "采集复制集状态", Type: "bool", Default: "true"},
			{Key: "gather_cluster_status", Label: "采集集群状态", Type: "bool", Default: "false"},
			{Key: "gather_db_stats", Label: "采集数据库统计", Type: "bool", Default: "true"},
			{Key: "interval", Label: "采集间隔", Type: "string", Default: "30s"},
		},
		TomlTemplate: `[[instances]]
servers = ["{{.servers}}"]
gather_repl_set_status = {{.gather_repl_set_status}}
gather_cluster_status = {{.gather_cluster_status}}
gather_db_stats = {{.gather_db_stats}}
labels = { instance="{{.servers}}" }
interval = "{{.interval}}"
`,
	}
}

func kubernetesTemplate() model.IntegrationTemplate {
	return model.IntegrationTemplate{
		ID:          "kubernetes",
		Name:        "Kubernetes",
		Description: "监控 Kubernetes 集群指标，包括 Node、Pod、Deployment 状态等",
		Category:    "container",
		Params: []model.TemplateParam{
			{Key: "url", Label: "API Server 地址", Type: "string", Required: true, Default: "https://kubernetes.default.svc", Placeholder: "https://host:6443"},
			{Key: "bearer_token_file", Label: "Token 文件路径", Type: "string", Default: "/var/run/secrets/kubernetes.io/serviceaccount/token"},
			{Key: "insecure_skip_verify", Label: "跳过 TLS 验证", Type: "bool", Default: "true"},
			{Key: "interval", Label: "采集间隔", Type: "string", Default: "30s"},
		},
		TomlTemplate: `[[instances]]
url = "{{.url}}"
bearer_token_file = "{{.bearer_token_file}}"
insecure_skip_verify = {{.insecure_skip_verify}}
labels = { instance="{{.url}}" }
interval = "{{.interval}}"
`,
	}
}

func httpResponseTemplate() model.IntegrationTemplate {
	return model.IntegrationTemplate{
		ID:          "http_response",
		Name:        "HTTP Response",
		Description: "HTTP 拨测监控，检测 URL 可达性、响应时间、状态码、证书过期等",
		Category:    "network",
		Params: []model.TemplateParam{
			{Key: "targets", Label: "目标 URL(逗号分隔)", Type: "string", Required: true, Placeholder: "https://example.com,http://api.local/health"},
			{Key: "method", Label: "HTTP 方法", Type: "string", Default: "GET", Options: []string{"GET", "POST", "HEAD"}},
			{Key: "timeout", Label: "超时时间", Type: "string", Default: "5s"},
			{Key: "expect_status", Label: "期望状态码", Type: "string", Default: "200"},
			{Key: "interval", Label: "采集间隔", Type: "string", Default: "30s"},
		},
		TomlTemplate: `[[instances]]
targets = [{{range $i, $v := split .targets ","}}{{if $i}}, {{end}}"{{trim $v}}"{{end}}]
method = "{{.method}}"
timeout = "{{.timeout}}"
interval = "{{.interval}}"

[instances.status_code]
expect = ["{{.expect_status}}*"]
severity = "Warning"

[instances.connectivity]
severity = "Critical"
`,
	}
}

func pingTemplate() model.IntegrationTemplate {
	return model.IntegrationTemplate{
		ID:          "ping",
		Name:        "Ping",
		Description: "ICMP Ping 拨测，检测主机可达性、丢包率、RTT 延迟等",
		Category:    "network",
		Params: []model.TemplateParam{
			{Key: "targets", Label: "目标地址(逗号分隔)", Type: "string", Required: true, Placeholder: "8.8.8.8,192.168.1.1"},
			{Key: "count", Label: "Ping 包数量", Type: "number", Default: "3"},
			{Key: "timeout", Label: "超时时间", Type: "string", Default: "3s"},
			{Key: "interval", Label: "采集间隔", Type: "string", Default: "30s"},
		},
		TomlTemplate: `[[instances]]
targets = [{{range $i, $v := split .targets ","}}{{if $i}}, {{end}}"{{trim $v}}"{{end}}]
count = {{.count}}
timeout = "{{.timeout}}"
interval = "{{.interval}}"

[instances.connectivity]
severity = "Critical"

[instances.packet_loss]
warn_ge = 1.0
critical_ge = 50.0
`,
	}
}

func netResponseTemplate() model.IntegrationTemplate {
	return model.IntegrationTemplate{
		ID:          "net_response",
		Name:        "Net Response",
		Description: "TCP/UDP 端口拨测，检测服务端口可达性和响应时间",
		Category:    "network",
		Params: []model.TemplateParam{
			{Key: "targets", Label: "目标地址(逗号分隔)", Type: "string", Required: true, Placeholder: "host1:22,host2:3306"},
			{Key: "protocol", Label: "协议", Type: "string", Default: "tcp", Options: []string{"tcp", "udp"}},
			{Key: "timeout", Label: "超时时间", Type: "string", Default: "1s"},
			{Key: "send", Label: "发送内容", Type: "string", Default: ""},
			{Key: "expect", Label: "期望返回", Type: "string", Default: ""},
			{Key: "interval", Label: "采集间隔", Type: "string", Default: "30s"},
		},
		TomlTemplate: `[[instances]]
targets = [{{range $i, $v := split .targets ","}}{{if $i}}, {{end}}"{{trim $v}}"{{end}}]
protocol = "{{.protocol}}"
timeout = "{{.timeout}}"
{{- if .send}}
send = "{{.send}}"
{{- end}}
{{- if .expect}}
expect = "{{.expect}}"
{{- end}}
interval = "{{.interval}}"

[instances.connectivity]
severity = "Critical"
`,
	}
}

func prometheusTemplate() model.IntegrationTemplate {
	return model.IntegrationTemplate{
		ID:          "prometheus",
		Name:        "Prometheus",
		Description: "抓取 Prometheus Exporter 暴露的 metrics 端点",
		Category:    "monitoring",
		Params: []model.TemplateParam{
			{Key: "urls", Label: "Metrics URL(逗号分隔)", Type: "string", Required: true, Placeholder: "http://host:9090/metrics"},
			{Key: "username", Label: "Basic Auth 用户名", Type: "string", Required: false},
			{Key: "password", Label: "Basic Auth 密码", Type: "password", Required: false},
			{Key: "metric_name_filter", Label: "指标名过滤(正则)", Type: "string", Default: ""},
			{Key: "interval", Label: "采集间隔", Type: "string", Default: "15s"},
		},
		TomlTemplate: `[[instances]]
urls = [{{range $i, $v := split .urls ","}}{{if $i}}, {{end}}"{{trim $v}}"{{end}}]
{{- if .username}}
username = "{{.username}}"
password = "{{.password}}"
{{- end}}
{{- if .metric_name_filter}}
name_filter = "{{.metric_name_filter}}"
{{- end}}
labels = { instance="{{.urls}}" }
interval = "{{.interval}}"
`,
	}
}

func javaTemplate() model.IntegrationTemplate {
	return model.IntegrationTemplate{
		ID:          "java",
		Name:        "Java (JMX)",
		Description: "通过 JMX 监控 Java 应用指标，包括 JVM 堆内存、GC、线程数等",
		Category:    "application",
		Params: []model.TemplateParam{
			{Key: "jmx_url", Label: "JMX URL", Type: "string", Required: true, Default: "service:jmx:rmi:///jndi/rmi://127.0.0.1:9999/jmxrmi", Placeholder: "service:jmx:rmi:///jndi/rmi://host:port/jmxrmi"},
			{Key: "username", Label: "JMX 用户名", Type: "string", Required: false},
			{Key: "password", Label: "JMX 密码", Type: "password", Required: false},
			{Key: "collect_jvm_heap", Label: "采集 JVM 堆内存", Type: "bool", Default: "true"},
			{Key: "collect_jvm_gc", Label: "采集 GC 指标", Type: "bool", Default: "true"},
			{Key: "collect_jvm_threads", Label: "采集线程指标", Type: "bool", Default: "true"},
			{Key: "interval", Label: "采集间隔", Type: "string", Default: "15s"},
		},
		TomlTemplate: `[[instances]]
jmx_url = "{{.jmx_url}}"
{{- if .username}}
username = "{{.username}}"
password = "{{.password}}"
{{- end}}

collect_jvm_heap = {{.collect_jvm_heap}}
collect_jvm_gc = {{.collect_jvm_gc}}
collect_jvm_threads = {{.collect_jvm_threads}}

labels = { instance="{{.jmx_url}}" }
interval = "{{.interval}}"
`,
	}
}
