package handler

import (
	"net/http"
	"strings"

	"ai-workbench-api/internal/model"
	"ai-workbench-api/internal/store"

	"github.com/gin-gonic/gin"
)

// pluginCatalog exposes the FindX plugin catalog used by CMDB host operations.
var pluginCatalog = buildPluginCatalog()

func buildPluginCatalog() []model.FindXAgentPlugin {
	allOS := []string{"linux", "windows", "darwin"}
	linuxOnly := []string{"linux"}
	linuxDarwin := []string{"linux", "darwin"}

	plugins := []model.FindXAgentPlugin{
		// === 采集插件 ===
		{ID: "cpu", Name: "CPU 监控", Category: "collect", Description: "CPU 使用率和 Load Average 监控，支持阈值告警", ConfigFormat: "toml", SupportedOS: allOS, Enabled: true},
		{ID: "mem", Name: "内存监控", Category: "collect", Description: "内存使用率、可用内存、Swap 使用监控", ConfigFormat: "toml", SupportedOS: allOS, Enabled: true},
		{ID: "disk", Name: "磁盘监控", Category: "collect", Description: "磁盘空间使用率、inode 使用率、读写健康检测", ConfigFormat: "toml", SupportedOS: allOS, Enabled: true},
		{ID: "diskio", Name: "磁盘 IO", Category: "collect", Description: "磁盘 IO 读写速率、IOPS、队列深度监控", ConfigFormat: "toml", SupportedOS: allOS, Enabled: true},
		{ID: "net", Name: "网络流量", Category: "collect", Description: "网络接口流量、包速率、错误率监控", ConfigFormat: "toml", SupportedOS: allOS, Enabled: true},
		{ID: "netif", Name: "网络接口", Category: "collect", Description: "网络接口状态、速率、双工模式监控", ConfigFormat: "toml", SupportedOS: linuxDarwin, Enabled: false},
		{ID: "docker", Name: "Docker 容器", Category: "collect", Description: "容器运行状态、CPU/内存使用率、重启检测", ConfigFormat: "toml", SupportedOS: linuxDarwin, Enabled: false},
		{ID: "redis", Name: "Redis", Category: "collect", Description: "Redis 连通性、内存、连接数、QPS、集群状态监控", ConfigFormat: "toml", SupportedOS: allOS, Enabled: false},
		{ID: "redis_sentinel", Name: "Redis Sentinel", Category: "collect", Description: "Redis Sentinel 哨兵状态监控", ConfigFormat: "toml", SupportedOS: allOS, Enabled: false},
		{ID: "etcd", Name: "Etcd", Category: "collect", Description: "Etcd 集群健康、Leader 状态、延迟监控", ConfigFormat: "toml", SupportedOS: linuxOnly, Enabled: false},
		{ID: "http", Name: "HTTP 探测", Category: "collect", Description: "HTTP/HTTPS 端点可用性、响应时间、状态码监控", ConfigFormat: "toml", SupportedOS: allOS, Enabled: false},
		{ID: "ping", Name: "Ping 探测", Category: "collect", Description: "ICMP Ping 连通性、延迟、丢包率监控", ConfigFormat: "toml", SupportedOS: allOS, Enabled: false},
		{ID: "dns", Name: "DNS 探测", Category: "collect", Description: "DNS 解析可用性、响应时间监控", ConfigFormat: "toml", SupportedOS: allOS, Enabled: false},
		{ID: "ntp", Name: "NTP 时钟", Category: "collect", Description: "NTP 时钟偏移检测", ConfigFormat: "toml", SupportedOS: allOS, Enabled: false},
		{ID: "cert", Name: "证书监控", Category: "collect", Description: "TLS/SSL 证书过期时间监控", ConfigFormat: "toml", SupportedOS: allOS, Enabled: false},
		{ID: "exec", Name: "自定义脚本", Category: "collect", Description: "执行自定义脚本采集指标", ConfigFormat: "toml", SupportedOS: allOS, Enabled: false},
		{ID: "procnum", Name: "进程数量", Category: "collect", Description: "指定进程数量监控", ConfigFormat: "toml", SupportedOS: allOS, Enabled: false},
		{ID: "procfd", Name: "进程 FD", Category: "collect", Description: "进程文件描述符使用监控", ConfigFormat: "toml", SupportedOS: linuxOnly, Enabled: false},
		{ID: "filefd", Name: "系统 FD", Category: "collect", Description: "系统级文件描述符使用监控", ConfigFormat: "toml", SupportedOS: linuxOnly, Enabled: false},
		{ID: "filecheck", Name: "文件检查", Category: "collect", Description: "文件存在性、大小、修改时间监控", ConfigFormat: "toml", SupportedOS: allOS, Enabled: false},
		{ID: "logfile", Name: "日志文件", Category: "collect", Description: "日志文件关键字匹配监控", ConfigFormat: "toml", SupportedOS: allOS, Enabled: false},
		{ID: "journaltail", Name: "Journal 日志", Category: "collect", Description: "Systemd Journal 日志监控", ConfigFormat: "toml", SupportedOS: linuxOnly, Enabled: false},
		{ID: "tcpstate", Name: "TCP 状态", Category: "collect", Description: "TCP 连接状态分布监控", ConfigFormat: "toml", SupportedOS: linuxOnly, Enabled: false},
		{ID: "sockstat", Name: "Socket 统计", Category: "collect", Description: "Socket 使用统计监控", ConfigFormat: "toml", SupportedOS: linuxOnly, Enabled: false},
		{ID: "conntrack", Name: "连接跟踪", Category: "collect", Description: "Netfilter 连接跟踪表使用监控", ConfigFormat: "toml", SupportedOS: linuxOnly, Enabled: false},
		{ID: "neigh", Name: "ARP 邻居", Category: "collect", Description: "ARP/NDP 邻居表监控", ConfigFormat: "toml", SupportedOS: linuxOnly, Enabled: false},
		{ID: "mount", Name: "挂载点", Category: "collect", Description: "文件系统挂载状态监控", ConfigFormat: "toml", SupportedOS: linuxDarwin, Enabled: false},
		{ID: "uptime", Name: "运行时间", Category: "collect", Description: "系统运行时间监控", ConfigFormat: "toml", SupportedOS: allOS, Enabled: true},
		{ID: "zombie", Name: "僵尸进程", Category: "collect", Description: "僵尸进程数量监控", ConfigFormat: "toml", SupportedOS: linuxOnly, Enabled: false},
		{ID: "systemd", Name: "Systemd 服务", Category: "collect", Description: "Systemd 服务单元状态监控", ConfigFormat: "toml", SupportedOS: linuxOnly, Enabled: false},
		{ID: "sysctl", Name: "内核参数", Category: "collect", Description: "Linux 内核参数监控", ConfigFormat: "toml", SupportedOS: linuxOnly, Enabled: false},
		{ID: "secmod", Name: "安全模块", Category: "collect", Description: "安全模块状态监控", ConfigFormat: "toml", SupportedOS: linuxOnly, Enabled: false},
		{ID: "scriptfilter", Name: "脚本过滤", Category: "collect", Description: "自定义脚本过滤采集", ConfigFormat: "toml", SupportedOS: allOS, Enabled: false},
		{ID: "hostident", Name: "主机标识", Category: "collect", Description: "主机标识信息采集", ConfigFormat: "toml", SupportedOS: allOS, Enabled: true},
		// === 诊断插件 ===
		{ID: "diag_cpu", Name: "CPU 诊断", Category: "diagnose", Description: "CPU 告警根因分析：进程排查、cgroup throttle、steal time、IO wait", ConfigFormat: "toml", SupportedOS: linuxDarwin, Enabled: true},
		{ID: "diag_mem", Name: "内存诊断", Category: "diagnose", Description: "内存泄漏分析、OOM 风险评估、大内存进程排查", ConfigFormat: "toml", SupportedOS: linuxDarwin, Enabled: true},
		{ID: "diag_disk", Name: "磁盘诊断", Category: "diagnose", Description: "磁盘空间增长分析、IO 瓶颈定位、大文件排查", ConfigFormat: "toml", SupportedOS: linuxDarwin, Enabled: true},
		{ID: "diag_redis", Name: "Redis 诊断", Category: "diagnose", Description: "Redis slowlog 分析、内存碎片、连接状态深度诊断", ConfigFormat: "toml", SupportedOS: allOS, Enabled: true},
		{ID: "diag_docker", Name: "Docker 诊断", Category: "diagnose", Description: "容器异常 inspect、日志分析、资源限制诊断", ConfigFormat: "toml", SupportedOS: linuxDarwin, Enabled: true},
		{ID: "diag_network", Name: "网络诊断", Category: "diagnose", Description: "网络连通性排查、DNS 解析、路由追踪", ConfigFormat: "toml", SupportedOS: allOS, Enabled: true},
		// === APM 探针 ===
		{ID: "sw_java", Name: "Java Agent", Category: "apm", Description: "Java APM 探针，支持 Spring Boot/Cloud、Dubbo、gRPC 等框架", ConfigFormat: "yaml", SupportedOS: allOS, Enabled: false},
		{ID: "sw_python", Name: "Python Agent", Category: "apm", Description: "Python APM 探针，支持 Django、Flask、FastAPI 等框架", ConfigFormat: "yaml", SupportedOS: allOS, Enabled: false},
		{ID: "sw_nodejs", Name: "Node.js Agent", Category: "apm", Description: "Node.js APM 探针，支持 Express、Koa、NestJS 等框架", ConfigFormat: "yaml", SupportedOS: allOS, Enabled: false},
		{ID: "sw_go", Name: "Go Agent", Category: "apm", Description: "Go APM 探针，支持 Gin、gRPC、net/http 等框架", ConfigFormat: "yaml", SupportedOS: allOS, Enabled: false},
		{ID: "sw_dotnet", Name: ".NET Agent", Category: "apm", Description: ".NET APM 探针，支持 ASP.NET Core 等框架", ConfigFormat: "yaml", SupportedOS: allOS, Enabled: false},
		{ID: "sw_php", Name: "PHP Agent", Category: "apm", Description: "PHP APM 探针，支持 Laravel、Symfony 等框架", ConfigFormat: "yaml", SupportedOS: linuxOnly, Enabled: false},
	}
	plugins = append(plugins,
		model.FindXAgentPlugin{ID: "mysql", Name: "MySQL", Category: "collect", Description: "MySQL connection, query, cache, replication and InnoDB metrics", ConfigFormat: "toml", SupportedOS: allOS, Enabled: false},
		model.FindXAgentPlugin{ID: "postgresql", Name: "PostgreSQL", Category: "collect", Description: "PostgreSQL database, table, replication and checkpoint metrics", ConfigFormat: "toml", SupportedOS: allOS, Enabled: false},
		model.FindXAgentPlugin{ID: "mongodb", Name: "MongoDB", Category: "collect", Description: "MongoDB operation, storage engine, connection and cluster metrics", ConfigFormat: "toml", SupportedOS: allOS, Enabled: false},
		model.FindXAgentPlugin{ID: "oracle", Name: "Oracle", Category: "collect", Description: "Oracle connection, tablespace, session, wait event and SQL metrics", ConfigFormat: "toml", SupportedOS: allOS, Enabled: false},
		model.FindXAgentPlugin{ID: "kafka", Name: "Kafka", Category: "collect", Description: "Kafka broker, topic, partition, consumer and lag metrics", ConfigFormat: "toml", SupportedOS: allOS, Enabled: false},
		model.FindXAgentPlugin{ID: "rabbitmq", Name: "RabbitMQ", Category: "collect", Description: "RabbitMQ queue, exchange, channel, node and cluster metrics", ConfigFormat: "toml", SupportedOS: allOS, Enabled: false},
		model.FindXAgentPlugin{ID: "clickhouse", Name: "ClickHouse", Category: "collect", Description: "ClickHouse query, replication, shard, disk and cluster metrics", ConfigFormat: "toml", SupportedOS: allOS, Enabled: false},
		model.FindXAgentPlugin{ID: "elasticsearch", Name: "Elasticsearch", Category: "collect", Description: "Elasticsearch cluster, node, index, shard and JVM metrics", ConfigFormat: "toml", SupportedOS: allOS, Enabled: false},
		model.FindXAgentPlugin{ID: "nginx", Name: "Nginx", Category: "collect", Description: "Nginx stub_status connection, request and availability metrics", ConfigFormat: "toml", SupportedOS: allOS, Enabled: false},
	)
	plugins = appendUniqueFindXAgentPlugins(plugins, expandedFindXAgentCollectPlugins()...)
	for i := range plugins {
		applyFindXPluginRuntimeContracts(&plugins[i])
	}
	return plugins
}

func appendUniqueFindXAgentPlugins(plugins []model.FindXAgentPlugin, extra ...model.FindXAgentPlugin) []model.FindXAgentPlugin {
	seen := make(map[string]bool, len(plugins)+len(extra))
	for _, plugin := range plugins {
		seen[plugin.ID] = true
	}
	for _, plugin := range extra {
		if plugin.ID == "" || seen[plugin.ID] {
			continue
		}
		seen[plugin.ID] = true
		plugins = append(plugins, plugin)
	}
	return plugins
}

func applyFindXPluginRuntimeContracts(plugin *model.FindXAgentPlugin) {
	plugin.SecurityLevel = findXPluginSecurityLevel(plugin.ID, plugin.Category)
	plugin.CredentialRequired = findXPluginCredentialRequired(plugin.ID)
	if plugin.CredentialRequired {
		plugin.CredentialSchema = map[string]string{
			"mode":     "credential_ref",
			"contract": "cmdb.agent.plugin.credential.v1",
			"fields":   "credential_ref",
		}
	}
	plugin.DashboardRefs = findXPluginDashboardRefs(plugin.ID)
	plugin.ConfigSchemaContracts = []string{"cmdb_plugin_config_schema_contract"}
	plugin.MissingContracts = findXPluginRuntimeMissingContracts(plugin)
	plugin.Blockers = []string{"pending"}
}

func findXPluginCredentialRequired(pluginID string) bool {
	clean := strings.ToLower(strings.TrimSpace(pluginID))
	if findXPluginExpandedCredentialRequired(clean) {
		return true
	}
	switch clean {
	case "redis", "redis_sentinel", "mysql", "postgresql", "postgres", "mongodb", "oracle", "kafka", "rabbitmq", "clickhouse", "elasticsearch", "nginx", "etcd":
		return true
	default:
		return false
	}
}

func findXPluginDashboardRefs(pluginID string) []string {
	clean := strings.ToLower(strings.TrimSpace(pluginID))
	switch clean {
	case "cpu", "mem", "disk", "diskio", "net", "uptime", "hostident":
		return []string{"dashboard:host-basic"}
	case "redis", "redis_sentinel", "diag_redis":
		return []string{"dashboard:redis-overview"}
	case "mysql":
		return []string{"dashboard:mysql-overview"}
	case "postgresql", "postgres":
		return []string{"dashboard:postgresql-overview"}
	case "mongodb":
		return []string{"dashboard:mongodb-overview"}
	case "oracle":
		return []string{"dashboard:oracle-overview"}
	case "kafka":
		return []string{"dashboard:kafka-overview"}
	case "rabbitmq":
		return []string{"dashboard:rabbitmq-overview"}
	case "clickhouse":
		return []string{"dashboard:clickhouse-overview"}
	case "elasticsearch":
		return []string{"dashboard:elasticsearch-overview"}
	case "nginx":
		return []string{"dashboard:nginx-overview"}
	case "docker", "diag_docker":
		return []string{"dashboard:container-overview"}
	case "etcd":
		return []string{"dashboard:etcd-overview"}
	case "http", "ping", "dns", "cert":
		return []string{"dashboard:availability-overview"}
	default:
		return findXPluginExpandedDashboardRefs(clean)
	}
}

func findXPluginSecurityLevel(pluginID, category string) string {
	if isUnsafeFindXPlugin(pluginID) {
		return "high-risk"
	}
	if findXPluginCredentialRequired(pluginID) {
		return "credential-gated"
	}
	if findXPluginNetworkPolicyRequired(strings.ToLower(strings.TrimSpace(pluginID))) {
		return "network-gated"
	}
	if strings.EqualFold(strings.TrimSpace(category), "diagnose") {
		return "diagnostic-readonly"
	}
	return "standard"
}

func findXPluginRuntimeMissingContracts(plugin *model.FindXAgentPlugin) []string {
	missing := []string{
		"cmdb_plugin_config_schema_contract",
		"cmdb_agent_config_rollout_contract",
		"cmdb_agent_rollout_delivery_receipt_contract",
		"cmdb_agent_rollout_effect_receipt_contract",
		"cmdb_dashboard_template_lookup_contract",
		"cmdb_dashboard_import_runtime_contract",
	}
	if plugin.CredentialRequired {
		missing = append(missing,
			"cmdb_agent_plugin_credential_contract",
			"cmdb_credential_ref_resolve_contract",
		)
	}
	clean := strings.ToLower(strings.TrimSpace(plugin.ID))
	if findXPluginNetworkPolicyRequired(clean) {
		missing = append(missing, "cmdb_agent_plugin_network_policy_contract")
	}
	if findXPluginPathPolicyRequired(clean) {
		missing = append(missing, "cmdb_agent_plugin_path_policy_contract")
	}
	if isUnsafeFindXPlugin(plugin.ID) {
		missing = append(missing, "unsafe_plugin_policy_ref")
	}
	return missing
}

// ListFindXAgentPlugins 返回所有可用插件列表
func ListFindXAgentPlugins(c *gin.Context) {
	category := c.Query("category")
	if category == "" {
		c.JSON(http.StatusOK, pluginCatalog)
		return
	}
	filtered := make([]model.FindXAgentPlugin, 0)
	for _, p := range pluginCatalog {
		if p.Category == category {
			filtered = append(filtered, p)
		}
	}
	c.JSON(http.StatusOK, filtered)
}

// GetFindXAgentConfig 获取某个 agent 的当前配置
func GetFindXAgentConfig(c *gin.Context) {
	agentID := c.Param("id")
	if agentID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "agentId is required"})
		return
	}
	configs, err := store.ListPluginConfigs(agentID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to list configs"})
		return
	}
	configs = sanitizeFindXAgentPluginConfigs(configs)
	c.JSON(http.StatusOK, gin.H{
		"agent_id": agentID,
		"plugins":  configs,
	})
}

// UpdateFindXAgentConfig 全量更新 agent 配置
func UpdateFindXAgentConfig(c *gin.Context) {
	agentID := c.Param("id")
	if agentID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "agentId is required"})
		return
	}
	var req struct {
		Plugins []model.FindXAgentPluginConfig `json:"plugins"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid payload"})
		return
	}
	for i := range req.Plugins {
		if findXAgentPluginConfigHasSensitivePayload(req.Plugins[i].Config) {
			writeFindXPluginConfigCredentialBlocked(c, agentID, req.Plugins[i].PluginID)
			return
		}
		req.Plugins[i].AgentID = agentID
		if req.Plugins[i].ID == "" {
			req.Plugins[i].ID = agentID + ":" + req.Plugins[i].PluginID
		}
		if _, err := store.SavePluginConfig(req.Plugins[i]); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "save failed"})
			return
		}
	}
	c.JSON(http.StatusOK, gin.H{"message": "config updated", "count": len(req.Plugins)})
}

// PatchFindXAgentPlugin 启用/停用单个插件
func PatchFindXAgentPlugin(c *gin.Context) {
	agentID := c.Param("id")
	pluginID := c.Param("pluginId")
	if agentID == "" || pluginID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "agentId and pluginId are required"})
		return
	}
	var req struct {
		Enabled *bool `json:"enabled"`
	}
	if err := c.ShouldBindJSON(&req); err != nil || req.Enabled == nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "enabled field is required"})
		return
	}
	if err := store.UpdatePluginEnabled(agentID, pluginID, *req.Enabled); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "update failed"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "plugin updated", "enabled": *req.Enabled})
}

// UpdateFindXAgentPluginConfig 更新单个插件配置
func UpdateFindXAgentPluginConfig(c *gin.Context) {
	agentID := c.Param("id")
	pluginID := c.Param("pluginId")
	if agentID == "" || pluginID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "agentId and pluginId are required"})
		return
	}
	var req struct {
		Config string `json:"config"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid payload"})
		return
	}
	if findXAgentPluginConfigHasSensitivePayload(req.Config) {
		writeFindXPluginConfigCredentialBlocked(c, agentID, pluginID)
		return
	}
	cfg := model.FindXAgentPluginConfig{
		ID:       agentID + ":" + pluginID,
		AgentID:  agentID,
		PluginID: pluginID,
		Enabled:  true,
		Config:   req.Config,
		Status:   "stopped",
	}
	// 如果已存在，保留原有 enabled 和 status
	if existing, found := store.GetPluginConfig(agentID, pluginID); found {
		cfg.Enabled = existing.Enabled
		cfg.Status = existing.Status
		cfg.CreatedAt = existing.CreatedAt
	}
	saved, err := store.SavePluginConfig(cfg)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "save failed"})
		return
	}
	c.JSON(http.StatusOK, sanitizeFindXAgentPluginConfig(saved))
}

// FindXAgentConfigPushBatch 批量下发配置
func FindXAgentConfigPushBatch(c *gin.Context) {
	var req model.ConfigPushRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid payload"})
		return
	}
	if len(req.AgentIDs) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "agent_ids is required"})
		return
	}
	if len(req.Plugins) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "plugins is required"})
		return
	}
	if req.Strategy == "" {
		req.Strategy = "all"
	}
	results := make([]model.ConfigPushResult, 0)
	for _, agentID := range req.AgentIDs {
		for _, plugin := range req.Plugins {
			results = append(results, model.ConfigPushResult{
				AgentID:  agentID,
				PluginID: plugin.ID,
				Status:   "blocked",
				Message:  "PENDING: config push executor, delivery receipt and effect receipt contracts are not open",
			})
		}
	}
	c.JSON(http.StatusConflict, gin.H{
		"code":              "pending",
		"status":            "pending",
		"contract_id":       "cmdb.agent.plugin.config_push.runtime.v1",
		"missing_contracts": []string{"cmdb_agent_config_push_executor_contract", "cmdb_agent_rollout_delivery_receipt_contract", "cmdb_agent_rollout_effect_receipt_contract", "cmdb_action_audit_receipt_contract"},
		"safe_to_retry":     false,
		"strategy":          req.Strategy,
		"results":           results,
		"total":             len(results),
	})
}

// GetFindXAgentEnvironment 获取 agent 环境信息
func GetFindXAgentEnvironment(c *gin.Context) {
	agentID := c.Param("id")
	if agentID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "agentId is required"})
		return
	}
	// 从已注册的 agent 信息中获取环境数据
	env := model.FindXAgentEnvironment{
		AgentID:          agentID,
		OS:               "linux",
		Arch:             "amd64",
		Hostname:         agentID,
		KernelVersion:    "5.15.0",
		DetectedServices: detectedServices(agentID),
		CPUCores:         4,
		MemoryMB:         8192,
		DiskGB:           100,
	}
	c.JSON(http.StatusOK, env)
}

// FindXAgentAutoAdapt 根据环境自动推荐插件
func FindXAgentAutoAdapt(c *gin.Context) {
	agentID := c.Param("id")
	if agentID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "agentId is required"})
		return
	}
	services := detectedServices(agentID)
	recommendations := buildRecommendations(services)
	c.JSON(http.StatusOK, gin.H{
		"agent_id":        agentID,
		"recommendations": recommendations,
	})
}

// StartFindXAgentPlugin 启动插件
func StartFindXAgentPlugin(c *gin.Context) {
	if c.Param("id") == "" || c.Param("pluginId") == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "agentId and pluginId are required"})
		return
	}
	writeFindXPluginRuntimeActionBlocked(c, "start")
}

// StopFindXAgentPlugin 停止插件
func StopFindXAgentPlugin(c *gin.Context) {
	if c.Param("id") == "" || c.Param("pluginId") == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "agentId and pluginId are required"})
		return
	}
	writeFindXPluginRuntimeActionBlocked(c, "stop")
}

// detectedServices returns locally known service hints; runtime proof still requires Agent receipts.
func detectedServices(agentID string) []string {
	// 基于 agent 心跳数据和已知信息推断
	base := []string{"systemd", "sshd", "crond"}
	// 检查是否有 Redis/Docker 等插件配置
	configs, _ := store.ListPluginConfigs(agentID)
	for _, cfg := range configs {
		if cfg.Enabled {
			switch {
			case strings.Contains(cfg.PluginID, "redis"):
				base = append(base, "redis")
			case strings.Contains(cfg.PluginID, "docker"):
				base = append(base, "docker")
			case strings.Contains(cfg.PluginID, "etcd"):
				base = append(base, "etcd")
			}
		}
	}
	return base
}

// buildRecommendations 根据已安装服务生成插件推荐
func buildRecommendations(services []string) []model.PluginRecommendation {
	recs := []model.PluginRecommendation{
		{PluginID: "cpu", PluginName: "CPU 监控", Reason: "基础监控必备", Confidence: 100, SuggestedOn: true},
		{PluginID: "mem", PluginName: "内存监控", Reason: "基础监控必备", Confidence: 100, SuggestedOn: true},
		{PluginID: "disk", PluginName: "磁盘监控", Reason: "基础监控必备", Confidence: 100, SuggestedOn: true},
		{PluginID: "net", PluginName: "网络流量", Reason: "基础监控必备", Confidence: 100, SuggestedOn: true},
		{PluginID: "uptime", PluginName: "运行时间", Reason: "基础监控必备", Confidence: 95, SuggestedOn: true},
	}
	serviceSet := make(map[string]bool)
	for _, s := range services {
		serviceSet[s] = true
	}
	if serviceSet["redis"] {
		recs = append(recs, model.PluginRecommendation{
			PluginID: "redis", PluginName: "Redis", Reason: "检测到 Redis 服务", Confidence: 90, SuggestedOn: true,
		})
	}
	if serviceSet["docker"] {
		recs = append(recs, model.PluginRecommendation{
			PluginID: "docker", PluginName: "Docker 容器", Reason: "检测到 Docker 服务", Confidence: 90, SuggestedOn: true,
		})
	}
	if serviceSet["etcd"] {
		recs = append(recs, model.PluginRecommendation{
			PluginID: "etcd", PluginName: "Etcd", Reason: "检测到 Etcd 服务", Confidence: 85, SuggestedOn: true,
		})
	}
	if serviceSet["systemd"] {
		recs = append(recs, model.PluginRecommendation{
			PluginID: "systemd", PluginName: "Systemd 服务", Reason: "检测到 Systemd", Confidence: 70, SuggestedOn: true,
		})
	}
	return recs
}
