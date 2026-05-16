package handler

import (
	"net/http"
	"strings"
	"unicode"

	"ai-workbench-api/internal/model"
	"ai-workbench-api/internal/store"

	"github.com/gin-gonic/gin"
)

const agentBlocked = "pending"

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
	if isWindowsInstallerInstallPlan(req) {
		plan = safeWindowsInstallPlanResponse(plan)
	}
	saved, err := store.SaveFindXAgentInstallPlan(plan)
	if err != nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "install plan persistence unavailable"})
		return
	}
	auditEvent(c, "findx_agent.install_plan.requested", saved.ID, "medium", "blocked", saved.Blocker, c.GetHeader("X-Test-Batch-Id"))
	if mode == "execute" {
		if isRemoteInstallerInstallPlan(req) {
			createBlockedFindXAgentRemoteInstallExecution(c, req, saved)
			return
		}
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
	if isCMDBHostPluginRollout(req, metadata) {
		if configRolloutOperationIdentityConflict(req, metadata) {
			metadata["plugin_action_conflict"] = "blocked"
		}
		metadata["plugin_action"] = configRolloutOperationMode(req, metadata)
	}
	if !hasConfigRolloutTarget(req) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "target_ids or agent_ids is required"})
		return
	}
	credentialCtx := cmdbResolvePluginCredentialContext(req, metadata)
	assignmentCtx := cmdbResolvePluginAssignmentContext(req, metadata)
	missing := missingConfigRolloutRefs(req, metadata)
	missing = cmdbApplyCredentialResolveGate(missing, credentialCtx)
	missing = cmdbApplyCredentialScopePolicyGate(missing, credentialCtx, configRolloutOperationMode(req, metadata))
	if isCMDBHostPluginRollout(req, metadata) && configRolloutOperationMode(req, metadata) == configRolloutPluginOperationDispatch {
		missing = cmdbFilterMissingContractsForDispatch(missing, assignmentCtx)
	}
	rollout := newBlockedFindXAgentConfigRollout(req, metadata)
	rollout.Blocker = configRolloutBlockerFromMissing(missing)
	saved, err := store.SaveFindXAgentConfigRollout(rollout)
	if err != nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "config rollout persistence unavailable"})
		return
	}
	if isCMDBHostPluginRollout(req, metadata) && configRolloutOperationMode(req, metadata) == configRolloutPluginOperationAssign {
		assignmentCtx, err = cmdbPersistPluginAssignment(c, req, metadata, saved, missing)
		if err != nil {
			c.JSON(http.StatusServiceUnavailable, gin.H{"error": "plugin assignment persistence unavailable"})
			return
		}
		if assignmentCtx.AssignmentReady {
			missing = cmdbFilterMissingContractsForAssignment(missing, true)
			saved.Blocker = configRolloutBlockerFromMissing(missing)
			saved.Metadata["assignment_ref"] = assignmentCtx.Assignment.ID
			saved.Metadata["target_binding_ref"] = assignmentCtx.TargetBinding.ID
			saved, err = store.SaveFindXAgentConfigRollout(saved)
			if err != nil {
				c.JSON(http.StatusServiceUnavailable, gin.H{"error": "config rollout persistence unavailable"})
				return
			}
		}
	}
	if isCMDBHostPluginRollout(req, metadata) && configRolloutOperationMode(req, metadata) == configRolloutPluginOperationDispatch {
		saved, err = cmdbAttachConfigRolloutReceiptRequestRefs(saved, req, metadata, requestActor(c))
		if err != nil {
			c.JSON(http.StatusServiceUnavailable, gin.H{"error": "config rollout receipt request persistence unavailable"})
			return
		}
	}
	auditEvent(c, "findx_agent.config_rollout.requested", saved.ID, "medium", "blocked", saved.Blocker, c.GetHeader("X-Test-Batch-Id"))
	c.JSON(http.StatusConflict, gin.H{
		"code":                http.StatusConflict,
		"error":               saved.Blocker,
		"status":              "blocked",
		"state_machine":       blockedExecutionStateMachine(saved.Blocker),
		"operation_contract":  cmdbOperationContractWithAssignmentAndCredential(configRolloutOperationContract(req, metadata, missing), assignmentCtx, credentialCtx),
		"receipt_contract":    configRolloutReceiptContract(req, metadata, saved.CredentialRefPresent, missing),
		"receipt_matrix":      configRolloutReceiptContractMatrix(),
		"blockers":            configRolloutResponseBlockers(missing),
		"missing_contracts":   missing,
		"assignment_contract": cmdbPluginAssignmentResponse(assignmentCtx),
		"credential_contract": cmdbPluginCredentialContractResponse(credentialCtx),
		"safe_to_retry":       false,
		"data":                safeConfigRolloutRuntimeReadDetail(saved),
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
		"state_machine":     blockedExecutionStateMachine(task.Blocker),
		"receipt_contract":  taskReceiptContract(action, task.Metadata, task.CredentialRefPresent, missing),
		"receipt_matrix":    findXAgentReceiptContractMatrix(),
		"blockers":          agentTaskResponseBlockers(missing),
		"missing_contracts": missing,
		"safe_to_retry":     false,
		"data":              taskResponseWithSafeExecutionMetadata(task),
	})
}

func blockedExecutionStateMachine(blocker string) model.FindXAgentExecutionStateMachine {
	return model.FindXAgentExecutionStateMachine{
		CurrentState:  model.FindXAgentExecutionStateBlockedByContract,
		AllowedStates: []string{model.FindXAgentExecutionStateBlockedByContract},
		Terminal:      true,
		SafeToRetry:   false,
		Blocker:       sanitizeInstallExecutionSummary(blocker),
	}
}

func installReceiptContract(scope string, req model.FindXAgentInstallPlanRequest, runner string, missing []string) model.FindXAgentReceiptContract {
	return model.FindXAgentReceiptContract{
		ID:                 scope + "_install_receipt_contract",
		Scope:              scope,
		Transport:          installPlanTransport(req),
		Runner:             runner,
		RequiredReceipts:   requiredReceiptNamesForScope(scope),
		MissingContracts:   uniquePackageRepositoryBlockers(missing),
		CredentialRequired: credentialRequiredForScope(scope),
		CredentialProvided: strings.TrimSpace(req.CredentialRef) != "",
		Status:             model.FindXAgentExecutionStateBlockedByContract,
		Blocker:            agentBlocked + ": executor, receipt, service, heartbeat, data-arrival and evidence contracts are not open",
	}
}

func taskReceiptContract(action string, metadata map[string]string, credentialProvided bool, missing []string) model.FindXAgentReceiptContract {
	return model.FindXAgentReceiptContract{
		ID:                 action + "_task_receipt_contract",
		Scope:              taskReceiptScope(metadata),
		Transport:          taskReceiptTransport(metadata),
		Runner:             taskReceiptRunner(metadata),
		RequiredReceipts:   []string{"execution_receipt", "service_receipt", "heartbeat_receipt", "data_arrival_receipt", "evidence_chain"},
		MissingContracts:   uniquePackageRepositoryBlockers(missing),
		CredentialRequired: true,
		CredentialProvided: credentialProvided,
		Status:             model.FindXAgentExecutionStateBlockedByContract,
		Blocker:            agentBlocked + ": task executor and receipt protocol are not open",
	}
}

func findXAgentReceiptContractMatrix() []model.FindXAgentReceiptContractMatrixRow {
	baseMissing := []string{"executor_contract", "execution_receipt_contract", "service_receipt_contract", "heartbeat_receipt_contract", "data_arrival_receipt_contract", "evidence_chain_contract"}
	return []model.FindXAgentReceiptContractMatrixRow{
		receiptMatrixRow("linux_local", "linux", "local curl -kfsSL + systemd", baseMissing),
		receiptMatrixRow("windows_local", "windows", "certutil / PowerShell + Windows Service", baseMissing),
		receiptMatrixRow("ssh", "linux", "SSH remote execution", append(baseMissing, "ssh_runner_contract", "host_key_contract")),
		receiptMatrixRow("winrm", "windows", "WinRM remote execution", append(baseMissing, "winrm_transport_contract")),
		receiptMatrixRow("systemd", "linux", "systemd service lifecycle", append(baseMissing, "systemd_unit_contract")),
		receiptMatrixRow("windows_service", "windows", "Windows Service lifecycle", append(baseMissing, "windows_service_contract")),
		receiptMatrixRow("iis", "windows", "IIS site and app pool lifecycle", append(baseMissing, "iis_receipt_contract")),
		receiptMatrixRow("docker", "linux/windows", "Docker container lifecycle", append(baseMissing, "container_receipt_contract")),
		receiptMatrixRow("helm", "kubernetes", "Helm release lifecycle", append(baseMissing, "helm_release_contract", "cluster_rbac_contract")),
		receiptMatrixRow("operator", "kubernetes", "Operator and CRD reconciliation", append(baseMissing, "operator_controller_contract", "crd_contract")),
		receiptMatrixRow("daemonset", "kubernetes", "DaemonSet rollout", append(baseMissing, "daemonset_rollout_contract")),
		receiptMatrixRow("sidecar", "kubernetes", "Sidecar injection", append(baseMissing, "sidecar_injection_contract")),
		receiptMatrixRow("initcontainer", "kubernetes", "InitContainer injection", append(baseMissing, "init_container_contract")),
	}
}

func receiptMatrixRow(scope, platform, surface string, missing []string) model.FindXAgentReceiptContractMatrixRow {
	return model.FindXAgentReceiptContractMatrixRow{
		Scope:             scope,
		Platform:          platform,
		ExecutionSurface:  surface,
		RequiredContracts: []string{"executor", "receipt", "service", "heartbeat", "data_arrival", "evidence"},
		MissingContracts:  uniquePackageRepositoryBlockers(missing),
		Status:            model.FindXAgentExecutionStateBlockedByContract,
		Blocker:           agentBlocked + ": " + scope + " executor receipt contract is not open",
	}
}

func installPlanTransport(req model.FindXAgentInstallPlanRequest) string {
	metadata := req.Metadata
	if transport := normalizeInstallPlanTransport(metadata["transport"]); transport != "" {
		return transport
	}
	method := strings.ToLower(strings.TrimSpace(req.Method))
	switch {
	case strings.Contains(method, "ssh"):
		return "ssh"
	case strings.Contains(method, "winrm"):
		return "winrm"
	case isKubernetesInstallerInstallPlan(req):
		return "kubernetes"
	default:
		return "local"
	}
}

func normalizeInstallPlanTransport(value string) string {
	transport := strings.ToLower(strings.TrimSpace(removeControlRunes(value)))
	transport = strings.NewReplacer("-", "_", " ", "_").Replace(transport)
	switch transport {
	case "ssh", "winrm", "kubernetes", "local", "helm", "operator", "daemonset", "sidecar", "initcontainer", "docker", "iis", "systemd", "windows_service":
		return transport
	default:
		return ""
	}
}

func taskReceiptScope(metadata map[string]string) string {
	text := agentTaskMatrixText(metadata)
	for _, scope := range []string{"winrm", "ssh", "systemd", "windows-service", "iis", "docker", "helm", "operator", "daemonset", "sidecar", "initcontainer", "init-container"} {
		if strings.Contains(text, scope) {
			return strings.ReplaceAll(scope, "-", "_")
		}
	}
	if strings.EqualFold(strings.TrimSpace(metadata["target_os"]), "windows") {
		return "windows_local"
	}
	if strings.Contains(strings.ToLower(strings.TrimSpace(metadata["target_os"])), "k8s") ||
		strings.Contains(strings.ToLower(strings.TrimSpace(metadata["target_os"])), "kubernetes") {
		return "kubernetes"
	}
	return "linux_local"
}

func taskReceiptTransport(metadata map[string]string) string {
	if transport := normalizeInstallPlanTransport(metadata["transport"]); transport != "" {
		return transport
	}
	if runner := normalizeInstallPlanTransport(metadata["runner"]); runner != "" {
		return runner
	}
	switch taskReceiptScope(metadata) {
	case "ssh":
		return "ssh"
	case "winrm":
		return "winrm"
	case "systemd":
		return "systemd"
	case "windows_service":
		return "windows_service"
	case "iis":
		return "iis"
	case "docker":
		return "docker"
	case "helm":
		return "helm"
	case "operator":
		return "operator"
	case "daemonset":
		return "daemonset"
	case "sidecar":
		return "sidecar"
	case "initcontainer":
		return "initcontainer"
	case "kubernetes":
		return "kubernetes"
	default:
		return "local"
	}
}

func taskReceiptRunner(metadata map[string]string) string {
	if runner := normalizeInstallPlanTransport(metadata["runner"]); runner != "" {
		return runner
	}
	switch taskReceiptScope(metadata) {
	case "ssh":
		return "ssh"
	case "winrm":
		return "winrm"
	default:
		return ""
	}
}

func taskResponseWithSafeExecutionMetadata(task model.FindXAgentExecutionTask) model.FindXAgentExecutionTask {
	task.Metadata = safeTaskResponseExecutionMetadata(task.Metadata)
	return task
}

func safeTaskResponseExecutionMetadata(input map[string]string) map[string]string {
	out := map[string]string{}
	for key, value := range input {
		switch strings.TrimSpace(key) {
		case "transport":
			if transport := normalizeInstallPlanTransport(value); transport != "" {
				out[key] = transport
			}
		case "runner":
			if runner := normalizeInstallPlanTransport(value); runner != "" {
				out[key] = runner
			}
		default:
			out[key] = value
		}
	}
	return out
}

func requiredReceiptNamesForScope(scope string) []string {
	switch scope {
	case "kubernetes", "helm", "operator", "daemonset", "sidecar", "initcontainer":
		return []string{"rollout_receipt", "service_account_receipt", "workload_status_receipt", "data_arrival_receipt", "evidence_chain"}
	case "windows_local", "windows_service", "iis":
		return []string{"install_receipt", "windows_service_receipt", "service_status_receipt", "heartbeat_receipt", "data_arrival_receipt", "evidence_chain"}
	default:
		return []string{"install_receipt", "systemd_receipt", "service_status_receipt", "heartbeat_receipt", "data_arrival_receipt", "evidence_chain"}
	}
}

func credentialRequiredForScope(scope string) bool {
	switch scope {
	case "linux_local", "windows_local":
		return false
	default:
		return true
	}
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
		"private_key", "session", "dsn", "marker", "sensitive",
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
		return "PENDING: 能力包源码、内置包仓库、签名证据、安装计划和配置下发契约未接入"
	}
	return "PENDING: 安装器生成、执行回执和审计协议未开放"
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
		return "PENDING: " + name + "仅有 Agent 能力上报线索，" + suffix
	}
	return "PENDING: " + name + suffix
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
