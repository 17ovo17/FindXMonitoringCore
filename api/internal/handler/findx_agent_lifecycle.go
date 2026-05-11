package handler

import (
	"net/http"
	"strings"
	"unicode"

	"ai-workbench-api/internal/model"
	"ai-workbench-api/internal/store"

	"github.com/gin-gonic/gin"
)

const agentBlocked = "BLOCKED_BY_CONTRACT"

type agentPackageDef struct {
	id                string
	name              string
	domain            string
	runtime           string
	osList            []string
	shape             string
	telemetryKinds    []string
	configKeys        []string
	configTemplateIDs []string
	pluginConfig      *model.FindXAgentPluginConfigSpec
}

type agentConfigTemplateDef struct {
	id                 string
	name               string
	kind               string
	scope              string
	fields             []string
	targetScopes       []string
	rolloutScopes      []string
	rollbackPolicy     string
	capabilityPackages []string
	pluginConfig       *model.FindXAgentPluginConfigSpec
}

func ListFindXAgentPackages(c *gin.Context) {
	c.JSON(http.StatusOK, findXAgentPackages())
}

func FindXAgentLifecycle(c *gin.Context) {
	packages := findXAgentPackages()
	readyPackages := 0
	for _, pkg := range packages {
		if pkg.Status == "ready" {
			readyPackages++
		}
	}
	agents := store.ListFindXAgents()
	online := 0
	for _, agent := range agents {
		if agent.Status == "online" {
			online++
		}
	}
	c.JSON(http.StatusOK, gin.H{
		"packages": packages,
		"phases":   agentLifecyclePhases(readyPackages, len(agents), online),
		"summary":  gin.H{"package_ready": readyPackages, "agents": len(agents), "online": online},
	})
}

func CreateFindXAgentInstallPlan(c *gin.Context) {
	var req model.FindXAgentInstallPlanRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid install plan payload"})
		return
	}
	mode, err := findXAgentInstallPlanMode(req)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	pkg, ok := findAgentPackage(req.PackageID)
	if !ok {
		c.JSON(http.StatusNotFound, gin.H{"error": "package not found"})
		return
	}
	targetIDs := append([]string{}, req.TargetIDs...)
	if strings.TrimSpace(req.TargetID) != "" {
		targetIDs = append(targetIDs, req.TargetID)
	}
	if len(cleanAgentLifecycleValues(targetIDs)) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "target_ids is required"})
		return
	}
	plan := newBlockedFindXAgentInstallPlan(req, pkg, targetIDs)
	saved, err := store.SaveFindXAgentInstallPlan(plan)
	if err != nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "install plan persistence unavailable"})
		return
	}
	auditEvent(c, "findx_agent.install_plan.requested", saved.ID, "medium", "blocked", saved.Blocker, c.GetHeader("X-Test-Batch-Id"))
	if mode == "execute" {
		if isKubernetesInstallerInstallPlan(req) {
			createBlockedFindXAgentKubernetesInstallExecution(c, req, saved)
			return
		}
		if isWindowsInstallerInstallPlan(req) {
			createBlockedFindXAgentWindowsInstallExecution(c, req, saved)
			return
		}
		createBlockedFindXAgentInstallExecution(c, req, saved)
		return
	}
	c.JSON(http.StatusConflict, gin.H{"code": http.StatusConflict, "error": saved.Blocker, "data": saved})
}

func ListFindXAgentConfigTemplates(c *gin.Context) {
	c.JSON(http.StatusOK, findXAgentConfigTemplates())
}

func CreateFindXAgentConfigRollout(c *gin.Context) {
	var req model.FindXAgentConfigRolloutRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid config rollout payload"})
		return
	}
	if strings.TrimSpace(req.TemplateID) == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "template_id is required"})
		return
	}
	if req.CanaryPercent < 0 || req.CanaryPercent > 100 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "canary_percent must be between 0 and 100"})
		return
	}
	metadata := safeAgentLifecycleMetadata(req.Metadata)
	if !hasConfigRolloutTarget(req) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "target_ids or agent_ids is required"})
		return
	}
	missing := missingConfigRolloutRefs(req, metadata)
	rollout := newBlockedFindXAgentConfigRollout(req, metadata)
	saved, err := store.SaveFindXAgentConfigRollout(rollout)
	if err != nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "config rollout persistence unavailable"})
		return
	}
	auditEvent(c, "findx_agent.config_rollout.requested", saved.ID, "medium", "blocked", saved.Blocker, c.GetHeader("X-Test-Batch-Id"))
	c.JSON(http.StatusConflict, gin.H{
		"code":              http.StatusConflict,
		"error":             saved.Blocker,
		"status":            "blocked",
		"blockers":          configRolloutResponseBlockers(missing),
		"missing_contracts": missing,
		"data":              saved,
	})
}

func FindXAgentDataArrival(c *gin.Context) {
	c.JSON(http.StatusOK, mergeDataArrivalEvidence(agentDataArrival(store.ListFindXAgents()), store.DataArrivalEvidenceSnapshot()))
}

func CreateFindXAgentTask(c *gin.Context) {
	var req model.FindXAgentTaskRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid agent task payload"})
		return
	}
	action := strings.ToLower(strings.TrimSpace(req.Action))
	if !validAgentTaskAction(action) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "unsupported agent task action"})
		return
	}
	if !hasAgentTaskTarget(req) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "target_ids or agent_ids is required"})
		return
	}
	task, err := saveBlockedAgentTask(c, req, action)
	if err != nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "agent task persistence unavailable"})
		return
	}
	missing := missingAgentTaskRefs(action, task.Metadata, task.CredentialRefPresent)
	c.JSON(http.StatusConflict, gin.H{
		"code":              http.StatusConflict,
		"error":             task.Blocker,
		"status":            "blocked",
		"blockers":          agentTaskResponseBlockers(missing),
		"missing_contracts": missing,
		"data":              task,
	})
}

func sanitizedConfigRolloutMetadata(req model.FindXAgentConfigRolloutRequest) gin.H {
	data := gin.H{
		"status":          "blocked",
		"remote_mutation": req.RemoteMutation,
		"canary_percent":  req.CanaryPercent,
	}
	for key, value := range map[string]string{
		"template_id":        req.TemplateID,
		"plugin_id":          req.PluginID,
		"plugin_version":     req.PluginVersion,
		"config_snippet_ref": req.ConfigSnippetRef,
		"config_format":      req.ConfigFormat,
		"provider_mode":      req.ProviderMode,
		"rollout_strategy":   req.RolloutStrategy,
		"rollback_ref":       req.RollbackRef,
		"audit_reason":       req.AuditReason,
		"change_ticket":      req.ChangeTicket,
	} {
		if sanitized := sanitizeRemoteMutationValue(key, value); sanitized != "" {
			data[key] = sanitized
		}
	}
	return data
}

func sanitizeRemoteMutationValue(key, value string) string {
	clean := strings.TrimSpace(removeControlRunes(value))
	if clean == "" || looksSensitive(key) || looksSensitive(clean) {
		return ""
	}
	const maxRemoteMutationMetadataLen = 120
	runes := []rune(clean)
	if len(runes) > maxRemoteMutationMetadataLen {
		clean = string(runes[:maxRemoteMutationMetadataLen])
	}
	return clean
}

func removeControlRunes(value string) string {
	return strings.Map(func(r rune) rune {
		if unicode.IsControl(r) {
			return -1
		}
		return r
	}, value)
}

func looksSensitive(value string) bool {
	normalized := strings.NewReplacer("-", "_", " ", "_", ".", "_").Replace(strings.ToLower(value))
	for _, marker := range []string{
		"credential", "password", "passwd", "secret", "token", "cookie",
		"authorization", "bearer", "api_key", "apikey", "access_key",
		"private_key", "session", "dsn",
	} {
		if strings.Contains(normalized, marker) {
			return true
		}
	}
	return false
}

func findAgentPackage(id string) (model.FindXAgentPackage, bool) {
	for _, pkg := range findXAgentPackages() {
		if pkg.ID == id {
			return pkg, true
		}
	}
	return model.FindXAgentPackage{}, false
}

func packageInstallBlocker(pkg model.FindXAgentPackage) string {
	if pkg.Status != "ready" {
		return "BLOCKED_BY_CONTRACT: 能力包源码、内置包仓库、签名证据、安装计划和配置下发契约未接入"
	}
	return "BLOCKED_BY_CONTRACT: 安装器生成、执行回执和审计协议未开放"
}

func agentLifecyclePhases(readyPackages, agents, online int) []model.FindXAgentLifecyclePhase {
	heartbeatBlocker := "心跳详情、服务注册、丢包检测、版本漂移和数据到达校验未开放"
	if agents > 0 || online > 0 {
		heartbeatBlocker = "已有 Agent 心跳清单，但心跳详情、服务注册、丢包检测、版本漂移和数据到达校验未开放"
	}
	return []model.FindXAgentLifecyclePhase{
		phase("package_repository", "内置包仓库", false, "内置包仓库、版本、哈希、许可证和 NOTICE 证据未开放"),
		phase("offline_package", "离线包", false, "离线包构建和存储未开放"),
		phase("signature", "签名校验", false, "能力包签名、摘要和完整性校验未开放"),
		phase("local_install", "本机安装", false, "安装器生成和执行回执未开放"),
		phase("remote_install", "远程安装", false, "远程执行、凭据引用和审计未开放"),
		phase("config_rollout", "配置下发", false, "统一配置模板保存、灰度、全量、回滚和审计未开放"),
		phase("heartbeat", "心跳", false, heartbeatBlocker),
		phase("data_arrival", "数据到达", false, "跨域数据到达验证未开放"),
		phase("upgrade", "升级", false, "升级任务下发和回滚协议未开放"),
		phase("rollback", "回滚", false, "回滚包和状态恢复协议未开放"),
		phase("uninstall", "卸载", false, "卸载任务下发和回执协议未开放"),
	}
}

func phase(key, name string, ready bool, blocker string) model.FindXAgentLifecyclePhase {
	if ready {
		return model.FindXAgentLifecyclePhase{Key: key, Name: name, Status: "ready"}
	}
	return model.FindXAgentLifecyclePhase{Key: key, Name: name, Status: "blocked", Blocker: agentBlocked + ": " + blocker}
}

func agentDataArrival(agents []*model.FindXAgent) []model.FindXAgentDataArrival {
	kinds := []model.FindXAgentDataArrival{
		{Kind: model.FindXAgentDataArrivalKindHeartbeat, Name: "心跳"},
		{Kind: model.FindXAgentDataArrivalKindMetrics, Name: "指标"},
		{Kind: model.FindXAgentDataArrivalKindLogs, Name: "日志"},
		{Kind: model.FindXAgentDataArrivalKindTracing, Name: "链路"},
		{Kind: model.FindXAgentDataArrivalKindProfiling, Name: "性能分析"},
		{Kind: model.FindXAgentDataArrivalKindInspection, Name: "巡检"},
		{Kind: model.FindXAgentDataArrivalKindTopology, Name: "拓扑"},
		{Kind: model.FindXAgentDataArrivalKindRUM, Name: "前端体验"},
		{Kind: model.FindXAgentDataArrivalKindGatewayTrace, Name: "网关链路"},
	}
	for i := range kinds {
		kinds[i] = fillDataArrival(kinds[i], agents)
	}
	return kinds
}

func fillDataArrival(item model.FindXAgentDataArrival, agents []*model.FindXAgent) model.FindXAgentDataArrival {
	for _, agent := range agents {
		if hasCapability(agent, item.Kind) {
			item.AgentCount++
			if agent.LastSeen.After(item.LastSeen) {
				item.LastSeen = agent.LastSeen
			}
		}
	}
	item.Status = model.FindXAgentDataArrivalStatusBlocked
	if item.AgentCount > 0 {
		item.Blocker = dataArrivalBlockedReason(item.Name, true)
		return item
	}
	item.Blocker = dataArrivalBlockedReason(item.Name, false)
	return item
}

func hasCapability(agent *model.FindXAgent, kind string) bool {
	for _, cap := range agent.Capabilities {
		cleanCap := strings.ToLower(strings.ReplaceAll(cap, "-", "_"))
		cleanKind := strings.ToLower(strings.ReplaceAll(kind, "-", "_"))
		if strings.Contains(cleanCap, cleanKind) || dataArrivalCapabilityAlias(cleanCap) == cleanKind {
			return true
		}
	}
	return false
}

func dataArrivalCapabilityAlias(capability string) string {
	if strings.Contains(capability, "gateway") {
		return model.FindXAgentDataArrivalKindGatewayTrace
	}
	if strings.Contains(capability, "trace") || strings.Contains(capability, "tracing") {
		return model.FindXAgentDataArrivalKindTracing
	}
	return ""
}

func dataArrivalBlockedReason(name string, hasCapability bool) string {
	const suffix = "数据到达验证器未开放；heartbeat、任务创建、命令预览或复制按钮不能替代该信号"
	if hasCapability {
		return "BLOCKED_BY_CONTRACT: " + name + "仅有 Agent 能力上报线索，" + suffix
	}
	return "BLOCKED_BY_CONTRACT: " + name + suffix
}

func validAgentTaskAction(action string) bool {
	switch action {
	case "upgrade", "rollback", "uninstall", "restart":
		return true
	case "sync_package_repository", "publish_package", "download_package", "verify_package_signature":
		return true
	default:
		return false
	}
}
