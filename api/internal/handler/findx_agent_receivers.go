package handler

import (
	"ai-workbench-api/internal/model"
	"ai-workbench-api/internal/store"
	"bytes"
	"compress/gzip"
	"crypto/subtle"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/spf13/viper"
	"io"
	"net"
	"net/http"
	"os"
	"sort"
	"strconv"
	"strings"
)



func ListFindXAgentInstallPlans(c *gin.Context) {
	if id := strings.TrimSpace(c.Query("id")); id != "" {
		item, ok, err := store.GetFindXAgentInstallPlan(id)
		if err != nil {
			c.JSON(http.StatusServiceUnavailable, gin.H{"error": "install plan detail unavailable"})
			return
		}
		if !ok {
			c.JSON(http.StatusNotFound, gin.H{"error": "install plan not found"})
			return
		}
		c.JSON(http.StatusOK, item)
		return
	}
	items, err := store.ListFindXAgentInstallPlans()
	writeAgentLifecycleList(c, items, err, "install plan list unavailable")
}

func ListFindXAgentInstallExecutions(c *gin.Context) {
	if id := strings.TrimSpace(c.Query("id")); id != "" {
		item, ok, err := store.GetFindXAgentInstallExecution(id)
		if err != nil {
			c.JSON(http.StatusServiceUnavailable, gin.H{"error": "install execution detail unavailable"})
			return
		}
		if !ok {
			c.JSON(http.StatusNotFound, gin.H{"error": "install execution not found"})
			return
		}
		c.JSON(http.StatusOK, item)
		return
	}
	items, err := store.ListFindXAgentInstallExecutions()
	writeAgentLifecycleList(c, items, err, "install execution list unavailable")
}

func ListFindXAgentConfigRollouts(c *gin.Context) {
	if id := strings.TrimSpace(c.Query("id")); id != "" {
		item, ok, err := store.GetFindXAgentConfigRollout(id)
		if err != nil {
			c.JSON(http.StatusServiceUnavailable, gin.H{"error": "config rollout detail unavailable"})
			return
		}
		if !ok {
			c.JSON(http.StatusNotFound, gin.H{"error": "config rollout not found"})
			return
		}
		c.JSON(http.StatusOK, safeConfigRolloutRuntimeReadDetail(item))
		return
	}
	items, err := store.ListFindXAgentConfigRollouts()
	writeAgentLifecycleList(c, items, err, "config rollout list unavailable")
}

func safeConfigRolloutRuntimeReadDetail(item model.FindXAgentConfigRollout) model.FindXAgentConfigRollout {
	metadata := make(map[string]string, len(item.Metadata))
	for key, value := range item.Metadata {
		cleanKey := strings.ToLower(strings.TrimSpace(key))
		if cleanKey == "writer_request_ref" || cleanKey == "cmdb_agent_rollout_writer_request_ref_contract" {
			continue
		}
		metadata[key] = value
	}
	if isCMDBHostPluginDispatchRolloutRecord(item) {
		missing := cmdbAgentRolloutRuntimeExecutorGapContractsForItem(item)
		metadata["runtime_read_status"] = "blocked"
		metadata["runtime_read_contract"] = cmdbAgentRolloutRuntimeReadContract
		metadata["runtime_read_missing_contracts"] = configRolloutRuntimeReadMissingJSON(missing)
	}
	item.Metadata = metadata
	return item
}

func ListFindXAgentTasks(c *gin.Context) {
	if id := strings.TrimSpace(c.Query("id")); id != "" {
		item, ok, err := store.GetFindXAgentExecutionTask(id)
		if err != nil {
			c.JSON(http.StatusServiceUnavailable, gin.H{"error": "agent task detail unavailable"})
			return
		}
		if !ok {
			c.JSON(http.StatusNotFound, gin.H{"error": "agent task not found"})
			return
		}
		c.JSON(http.StatusOK, item)
		return
	}
	items, err := store.ListFindXAgentExecutionTasks()
	writeAgentLifecycleList(c, items, err, "agent task list unavailable")
}

func ListFindXAgentDataArrivalEvidence(c *gin.Context) {
	if handled, err := listFindXAgentDataArrivalEvidenceRuntimeRead(c); handled {
		if err != nil {
			c.JSON(http.StatusServiceUnavailable, gin.H{"error": "data arrival evidence runtime read unavailable"})
		}
		return
	}
	items, err := store.ListFindXAgentDataArrivalEvidence()
	writeAgentLifecycleList(c, items, err, "data arrival evidence list unavailable")
}

func writeAgentLifecycleList(c *gin.Context, items any, err error, message string) {
	if err != nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": message})
		return
	}
	c.JSON(http.StatusOK, items)
}

func saveBlockedAgentTask(c *gin.Context, req model.FindXAgentTaskRequest, action string) (model.FindXAgentExecutionTask, error) {
	metadata := safeAgentLifecycleMetadata(req.Metadata)
	credentialRefPresent := strings.TrimSpace(req.CredentialRef) != ""
	task := model.FindXAgentExecutionTask{
		Action:               action,
		AgentIDs:             cleanAgentLifecycleValues(req.AgentIDs),
		TargetIDs:            cleanAgentLifecycleValues(req.TargetIDs),
		PackageID:            sanitizeRemoteMutationValue("package_id", req.PackageID),
		ConfigVersion:        sanitizeRemoteMutationValue("config_version", req.ConfigVersion),
		Status:               "accepted",
		Blocker:              "",
		Audit:                "findx_agent.task.created",
		CredentialRefPresent: credentialRefPresent,
		Metadata:             metadata,
	}
	saved, err := store.SaveFindXAgentExecutionTask(task)
	if err != nil {
		return model.FindXAgentExecutionTask{}, err
	}
	auditEvent(c, "findx_agent.task.created", saved.ID, "medium", "accepted", "", c.GetHeader("X-Test-Batch-Id"))
	return saved, nil
}

func blockedAgentTaskReason(action string, metadata map[string]string, credentialRefPresent bool) string {
	missing := missingAgentTaskRefs(action, metadata, credentialRefPresent)
	if len(missing) > 0 {
		return fmt.Sprintf("PENDING: missing %s", strings.Join(missing, ", "))
	}
	return "PENDING: executor not enabled / execution protocol not open"
}

func missingAgentTaskRefs(action string, metadata map[string]string, credentialRefPresent bool) []string {
	required := append(requiredAgentTaskRefs(action), requiredRemoteExecutionRefs()...)
	required = append(required, requiredAgentTaskExecutorReceiptRefs(metadata)...)
	if isKubernetesAgentTask(metadata) {
		required = append(required, requiredKubernetesAgentTaskRefs(action, metadata)...)
	}
	missingSet := map[string]bool{}
	for _, key := range required {
		if strings.TrimSpace(metadata[key]) == "" {
			missingSet[key] = true
		}
	}
	if !credentialRefPresent {
		missingSet["credential_ref"] = true
	}
	for _, key := range missingRemoteExecutionChoiceRefs(metadata) {
		missingSet[key] = true
	}
	if isKubernetesAgentTask(metadata) {
		for _, key := range missingKubernetesAgentTaskChoiceRefs(action, metadata) {
			missingSet[key] = true
		}
	}
	missing := make([]string, 0, len(missingSet))
	for key := range missingSet {
		missing = append(missing, key)
	}
	sort.Strings(missing)
	return missing
}

func agentTaskResponseBlockers(missing []string) []string {
	if len(missing) == 0 {
		return []string{agentBlocked, "EXECUTOR_DISABLED_BY_CONTRACT"}
	}
	values := append([]string{agentBlocked, "MISSING_CONTRACTS"}, missing...)
	return uniquePackageRepositoryBlockers(values)
}

func requiredRemoteExecutionRefs() []string {
	return []string{"idempotency_key", "target_os", "timeout_policy_ref"}
}

func requiredAgentTaskExecutorReceiptRefs(metadata map[string]string) []string {
	refs := []string{"data_arrival_validator_ref"}
	text := agentTaskMatrixText(metadata)
	if isLocalLinuxAgentTask(metadata, text) {
		refs = append(refs, "local_executor_ref", "linux_installer_ref", "service_receipt_ref")
	}
	if isLocalWindowsAgentTask(metadata, text) {
		refs = append(refs, "local_executor_ref", "windows_installer_ref", "service_receipt_ref")
	}
	addAgentTaskRefsIf(&refs, text, []string{"local-service"}, "service_manifest_ref", "service_receipt_ref")
	addAgentTaskRefsIf(&refs, text, []string{"ssh"}, "ssh_runner_ref", "remote_executor_ref")
	addAgentTaskRefsIf(&refs, text, []string{"winrm"}, "winrm_endpoint_ref", "winrm_transport_ref", "remote_executor_ref")
	addAgentTaskRefsIf(&refs, text, []string{"systemd"}, "systemd_unit_ref", "systemd_receipt_ref")
	addAgentTaskRefsIf(&refs, text, []string{"windows-service", "windows service"}, "windows_service_ref", "windows_service_receipt_ref")
	addAgentTaskRefsIf(&refs, text, []string{"iis"}, "iis_site_ref", "iis_app_pool_ref", "iis_receipt_ref")
	addAgentTaskRefsIf(&refs, text, []string{"docker"}, "container_ref", "image_ref", "docker_receipt_ref")
	addAgentTaskRefsIf(&refs, text, []string{"operator"}, "operator_ref", "crd_ref", "controller_receipt_ref")
	addAgentTaskRefsIf(&refs, text, []string{"daemonset"}, "daemonset_ref")
	addAgentTaskRefsIf(&refs, text, []string{"sidecar"}, "sidecar_injection_ref")
	addAgentTaskRefsIf(&refs, text, []string{"initcontainer", "init-container"}, "init_container_ref")
	return refs
}

func agentTaskMatrixText(metadata map[string]string) string {
	keys := []string{"target_os", "transport", "runner", "method", "orchestrator", "workload_kind", "service_type", "runtime"}
	values := make([]string, 0, len(keys))
	for _, key := range keys {
		values = append(values, strings.ToLower(strings.TrimSpace(metadata[key])))
	}
	return strings.Join(values, " ")
}

func isLocalLinuxAgentTask(metadata map[string]string, text string) bool {
	targetOS := strings.ToLower(strings.TrimSpace(metadata["target_os"]))
	return strings.Contains(text, "local-linux") || (targetOS == "linux" && strings.TrimSpace(metadata["transport"]) == "" && strings.TrimSpace(metadata["runner"]) == "")
}

func isLocalWindowsAgentTask(metadata map[string]string, text string) bool {
	targetOS := strings.ToLower(strings.TrimSpace(metadata["target_os"]))
	return strings.Contains(text, "local-windows") || (targetOS == "windows" && strings.TrimSpace(metadata["transport"]) == "" && strings.TrimSpace(metadata["runner"]) == "")
}

func addAgentTaskRefsIf(refs *[]string, text string, markers []string, values ...string) {
	for _, marker := range markers {
		if strings.Contains(text, marker) {
			*refs = append(*refs, values...)
			return
		}
	}
}

func missingRemoteExecutionChoiceRefs(metadata map[string]string) []string {
	missing := []string{}
	text := agentTaskMatrixText(metadata)
	if strings.TrimSpace(metadata["transport"]) == "" && strings.TrimSpace(metadata["runner"]) == "" && !isLocalLinuxAgentTask(metadata, text) && !isLocalWindowsAgentTask(metadata, text) {
		missing = append(missing, "transport_or_runner")
	}
	if strings.TrimSpace(metadata["execution_receipt_ref"]) == "" && strings.TrimSpace(metadata["receipt_ref"]) == "" {
		missing = append(missing, "execution_receipt_ref_or_receipt_ref")
	}
	if strings.TrimSpace(metadata["audit_ref"]) == "" && strings.TrimSpace(metadata["evidence_chain_ref"]) == "" {
		missing = append(missing, "audit_ref_or_evidence_chain_ref")
	}
	if strings.Contains(text, "ssh") && strings.TrimSpace(metadata["ssh_host_key"]) == "" && strings.TrimSpace(metadata["ssh_fingerprint"]) == "" {
		missing = append(missing, "ssh_host_key_or_fingerprint")
		missing = append(missing, "ssh_host_key", "ssh_fingerprint")
	}
	return missing
}

func requiredAgentTaskRefs(action string) []string {
	switch action {
	case "sync_package_repository":
		return []string{"package_repository_ref", "manifest_ref", "checksum", "signature_ref", "executor_ref"}
	case "publish_package":
		return []string{"package_id", "package_repository_ref", "release_manifest_ref", "checksum", "signature_ref", "executor_ref"}
	case "download_package":
		return []string{"package_id", "package_repository_ref", "artifact_ref", "checksum", "signature_ref", "public_key_ref", "executor_ref"}
	case "verify_package_signature":
		return []string{"package_id", "package_repository_ref", "checksum", "signature_ref", "public_key_ref", "verifier_ref"}
	case "uninstall":
		return []string{"uninstall_manifest_ref", "executor_ref"}
	case "rollback":
		return []string{"rollback_manifest_ref", "state_snapshot_ref", "executor_ref"}
	case "upgrade":
		return []string{"package_repository_ref", "signature_ref", "checksum", "script_manifest_ref", "executor_ref"}
	case "restart":
		return []string{"service_ref", "executor_ref"}
	default:
		return []string{}
	}
}

func isKubernetesAgentTask(metadata map[string]string) bool {
	for _, key := range kubernetesAgentTaskMarkerKeys() {
		value := strings.ToLower(strings.TrimSpace(metadata[key]))
		for _, marker := range kubernetesAgentTaskMarkers() {
			if strings.Contains(value, marker) {
				return true
			}
		}
	}
	return false
}

func kubernetesAgentTaskMarkerKeys() []string {
	return []string{"target_os", "runner", "transport", "orchestrator", "method", "workload_kind"}
}

func kubernetesAgentTaskMarkers() []string {
	return []string{"kubernetes", "k8s", "helm", "daemonset", "sidecar", "initcontainer", "init-container"}
}

func requiredKubernetesAgentTaskRefs(action string, metadata map[string]string) []string {
	required := []string{
		"cluster_ref",
		"namespace_ref",
		"workload_selector_ref",
		"rbac_ref",
		"service_account_ref",
		"rollout_strategy_ref",
		"rollout_receipt_ref",
		"data_arrival_validator_ref",
	}
	if isHelmAgentTask(metadata) {
		required = append(required, "helm_release_ref")
	}
	switch action {
	case "upgrade":
		required = append(required, "values_ref", "image_ref", "config_map_ref")
	case "rollback":
		required = append(required, "rollback_revision_ref")
	case "uninstall":
		required = append(required, "teardown_plan_ref")
	case "restart":
		required = append(required, "restart_strategy_ref")
	}
	return required
}

func missingKubernetesAgentTaskChoiceRefs(action string, metadata map[string]string) []string {
	if action != "upgrade" {
		return nil
	}
	if strings.TrimSpace(metadata["helm_chart_ref"]) != "" || strings.TrimSpace(metadata["manifest_bundle_ref"]) != "" {
		return nil
	}
	return []string{"helm_chart_ref_or_manifest_bundle_ref"}
}

func isHelmAgentTask(metadata map[string]string) bool {
	for _, key := range kubernetesAgentTaskMarkerKeys() {
		if strings.Contains(strings.ToLower(strings.TrimSpace(metadata[key])), "helm") {
			return true
		}
	}
	return false
}

func hasAgentTaskTarget(req model.FindXAgentTaskRequest) bool {
	return len(cleanAgentLifecycleValues(req.AgentIDs)) > 0 || len(cleanAgentLifecycleValues(req.TargetIDs)) > 0
}

func cleanAgentLifecycleValues(values []string) []string {
	out := []string{}
	seen := map[string]bool{}
	for _, value := range values {
		clean := strings.TrimSpace(value)
		if clean != "" && !seen[clean] {
			seen[clean] = true
			out = append(out, clean)
		}
	}
	return out
}

func safeAgentLifecycleMetadata(input map[string]string) map[string]string {
	out := map[string]string{}
	for key, value := range input {
		cleanKey := strings.TrimSpace(key)
		allowedRef := allowedAgentLifecycleReferenceKey(cleanKey)
		if cleanKey == "" || (!allowedRef && looksSensitive(value)) {
			continue
		}
		if looksSensitive(cleanKey) && !allowedAgentLifecycleReferenceKey(cleanKey) {
			continue
		}
		cleanValue := sanitizeRemoteMutationValue(cleanKey, value)
		if allowedRef {
			cleanValue = sanitizeAgentLifecycleReferenceValue(value)
		}
		if cleanValue != "" {
			out[cleanKey] = cleanValue
		}
	}
	return out
}

func allowedAgentLifecycleReferenceKey(key string) bool {
	switch strings.TrimSpace(key) {
	case "provider_auth_ref":
		return true
	default:
		return false
	}
}

func sanitizeAgentLifecycleReferenceValue(value string) string {
	clean := strings.TrimSpace(removeControlRunes(value))
	if clean == "" || looksSensitiveReferenceValue(clean) {
		return ""
	}
	const maxAgentLifecycleReferenceLen = 120
	runes := []rune(clean)
	if len(runes) > maxAgentLifecycleReferenceLen {
		clean = string(runes[:maxAgentLifecycleReferenceLen])
	}
	return clean
}

func looksSensitiveReferenceValue(value string) bool {
	normalized := strings.NewReplacer("-", "_", " ", "_", ".", "_").Replace(strings.ToLower(value))
	for _, marker := range []string{"password", "passwd", "secret", "token", "cookie", "bearer", "api_key", "apikey", "access_key", "private_key", "session", "dsn"} {
		if strings.Contains(normalized, marker) {
			return true
		}
	}
	return false
}


const categrafReceiverBodyLimit = 2 << 20

type categrafHeartbeatPayload struct {
	Ident        string            `json:"ident"`
	Agent        string            `json:"agent"`
	AgentID      string            `json:"agent_id"`
	Host         string            `json:"host"`
	Hostname     string            `json:"hostname"`
	IP           string            `json:"ip"`
	HostIP       string            `json:"host_ip"`
	Version      string            `json:"version"`
	AgentVersion string            `json:"agent_version"`
	OS           string            `json:"os"`
	Arch         string            `json:"arch"`
	Plugin       string            `json:"plugin"`
	Source       string            `json:"source"`
	Scope        string            `json:"scope"`
	Collector    string            `json:"collector"`
	Tags         map[string]string `json:"tags"`
	Labels       map[string]string `json:"labels"`
	GlobalLabels map[string]string `json:"global_labels"`
	Metadata     map[string]string `json:"metadata"`
	UnixTime     int64             `json:"unixtime"`
	Timestamp    int64             `json:"timestamp"`
	RemoteAddr   string            `json:"-"`
}

func CategrafN9EHeartbeat(c *gin.Context) {
	if !validCategrafReceiverSource(c) {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid agent receiver token"})
		return
	}
	payload, err := readCategrafHeartbeat(c)
	if err != nil {
		writeCategrafReceiverError(c, err)
		return
	}
	heartbeat := payload.toFindXAgentHeartbeat()
	agent, target, err := store.UpsertFindXAgentHeartbeat(heartbeat)
	if err != nil {
		if strings.Contains(err.Error(), "future") {
			c.JSON(http.StatusBadRequest, gin.H{"error": "heartbeat time is too far in the future"})
			return
		}
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "agent heartbeat persistence unavailable"})
		return
	}
	evidence := payload.toEvidence(agent.ID, target.ID)
	saved, err := store.SaveFindXAgentDataArrivalEvidence(evidence)
	if err != nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "heartbeat evidence persistence unavailable"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"ok": true, "status": saved.Status, "agent_id": agent.ID, "target_id": target.ID, "evidence_id": saved.ID})
}

func CategrafPrometheusRemoteWrite(c *gin.Context) {
	if !validCategrafReceiverSource(c) {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid agent receiver token"})
		return
	}
	body, err := readLimitedReceiverBody(c.Request.Body, categrafReceiverBodyLimit)
	if err != nil {
		writeCategrafReceiverError(c, err)
		return
	}
	if len(body) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "remote write body is required"})
		return
	}
	metadata := remoteWriteEvidenceMetadata(c, len(body))
	evidence := model.FindXAgentDataArrivalEvidence{
		Kind:         model.FindXAgentDataArrivalKindMetrics,
		AgentID:      sanitizeReceiverValue("agent_id", firstReceiverValue(c.Query("agent_id"), c.GetHeader("X-FindX-Agent-Id"))),
		TargetID:     sanitizeReceiverValue("target_id", firstReceiverValue(c.Query("target_id"), c.GetHeader("X-FindX-Target-Id"))),
		Status:       model.FindXAgentDataArrivalStatusReported,
		EvidenceRefs: []string{"receiver:/metrics/write-compatible"},
		Metadata:     metadata,
	}
	saved, err := store.SaveFindXAgentDataArrivalEvidence(evidence)
	if err != nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "metrics evidence persistence unavailable"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"ok": true, "status": saved.Status, "evidence_id": saved.ID})
}


func readCategrafHeartbeat(c *gin.Context) (categrafHeartbeatPayload, error) {
	body, err := readReceiverBody(c.Request.Body, c.GetHeader("Content-Encoding"), categrafReceiverBodyLimit)
	if err != nil {
		return categrafHeartbeatPayload{}, err
	}
	var payload categrafHeartbeatPayload
	if err := json.Unmarshal(body, &payload); err != nil {
		return categrafHeartbeatPayload{}, errBadRequest("invalid heartbeat payload")
	}
	payload.RemoteAddr = clientHost(c.Request.RemoteAddr)
	if err := validateCategrafHeartbeat(payload); err != nil {
		return categrafHeartbeatPayload{}, err
	}
	return payload, nil
}

func validateCategrafHeartbeat(payload categrafHeartbeatPayload) error {
	if strings.TrimSpace(payload.ident()) == "" && strings.TrimSpace(payload.agentIP()) == "" && strings.TrimSpace(payload.Hostname) == "" && strings.TrimSpace(payload.Host) == "" {
		return errBadRequest("agent, host, hostname, or ip is required")
	}
	if strings.TrimSpace(payload.agentIP()) != "" {
		if _, ok := cleanIP(payload.agentIP()); !ok {
			return errBadRequest("valid ip is required")
		}
	}
	return nil
}

func (p categrafHeartbeatPayload) toFindXAgentHeartbeat() model.FindXAgentHeartbeat {
	return model.FindXAgentHeartbeat{
		Ident:        firstReceiverValue(p.ident(), p.Host, p.Hostname, p.IP),
		IP:           firstReceiverValue(p.agentIP(), p.RemoteAddr),
		Hostname:     firstReceiverValue(p.Hostname, p.Host),
		OS:           sanitizeReceiverValue("os", p.OS),
		Arch:         sanitizeReceiverValue("arch", p.Arch),
		Version:      sanitizeReceiverValue("version", firstReceiverValue(p.Version, p.AgentVersion)),
		Collector:    "findx-agent-host-collector",
		Capabilities: []string{"metrics", "heartbeat"},
		GlobalLabels: safeReceiverMetadata(p.Labels, p.Tags, p.GlobalLabels, p.Metadata),
		UnixTime:     firstReceiverUnixTime(p.UnixTime, p.Timestamp),
	}
}

func (p categrafHeartbeatPayload) toEvidence(agentID, targetID string) model.FindXAgentDataArrivalEvidence {
	metadata := safeReceiverMetadata(p.Labels, p.Tags, p.GlobalLabels, p.Metadata)
	for key, value := range map[string]string{
		"agent":     p.ident(),
		"host":      firstReceiverValue(p.Host, p.Hostname),
		"hostname":  firstReceiverValue(p.Hostname, p.Host),
		"ip":        firstReceiverValue(p.agentIP(), p.RemoteAddr),
		"os":        p.OS,
		"arch":      p.Arch,
		"version":   firstReceiverValue(p.Version, p.AgentVersion),
		"plugin":    p.Plugin,
		"source":    "findx-agent-compatible",
		"scope":     p.Scope,
		"collector": "findx-agent-host-collector",
	} {
		if clean := sanitizeReceiverValue(key, value); clean != "" {
			metadata[key] = clean
		}
	}
	return model.FindXAgentDataArrivalEvidence{
		Kind:         model.FindXAgentDataArrivalKindHeartbeat,
		AgentID:      agentID,
		TargetID:     targetID,
		Status:       model.FindXAgentDataArrivalStatusReported,
		EvidenceRefs: []string{"receiver:/agent/heartbeat-compatible"},
		Metadata:     metadata,
	}
}

func (p categrafHeartbeatPayload) ident() string {
	return firstReceiverValue(p.Ident, p.AgentID, p.Agent)
}

func (p categrafHeartbeatPayload) agentIP() string {
	return firstReceiverValue(p.IP, p.HostIP)
}


func readReceiverBody(body io.Reader, encoding string, limit int64) ([]byte, error) {
	if !strings.EqualFold(strings.TrimSpace(encoding), "gzip") {
		return readLimitedReceiverBody(body, limit)
	}
	gzipBody, err := gzip.NewReader(body)
	if err != nil {
		return nil, errBadRequest("invalid gzip receiver payload")
	}
	defer gzipBody.Close()
	return readLimitedReceiverBody(gzipBody, limit)
}

func readLimitedReceiverBody(body io.Reader, limit int64) ([]byte, error) {
	var buf bytes.Buffer
	n, err := io.Copy(&buf, io.LimitReader(body, limit+1))
	if err != nil {
		return nil, err
	}
	if n > limit {
		return nil, errReceiverBodyTooLarge{}
	}
	return buf.Bytes(), nil
}

func remoteWriteEvidenceMetadata(c *gin.Context, size int) map[string]string {
	metadata := map[string]string{
		"receiver":     "remote_write_compatible",
		"body_bytes":   strconv.Itoa(size),
		"content_type": sanitizeReceiverValue("content_type", c.GetHeader("Content-Type")),
	}
	for key, value := range map[string]string{
		"agent_id":  firstReceiverValue(c.Query("agent_id"), c.GetHeader("X-FindX-Agent-Id")),
		"target_id": firstReceiverValue(c.Query("target_id"), c.GetHeader("X-FindX-Target-Id")),
		"scope":     c.Query("scope"),
	} {
		if clean := sanitizeReceiverValue(key, value); clean != "" {
			metadata[key] = clean
		}
	}
	if encoding := sanitizeReceiverValue("content_encoding", c.GetHeader("Content-Encoding")); encoding != "" {
		metadata["content_encoding"] = encoding
	}
	if ip := clientHost(c.Request.RemoteAddr); ip != "" {
		metadata["remote_ip"] = ip
	}
	mergeReceiverDispatchRuntimeMetadata(metadata, c)
	return metadata
}

func mergeReceiverDispatchRuntimeMetadata(metadata map[string]string, c *gin.Context) {
	rolloutRef := firstReceiverValue(
		c.Query("source_rollout_id"),
		c.Query("config_rollout_id"),
		c.Query("rollout_id"),
		c.Query("rollout_ref"),
		c.GetHeader("X-FindX-Source-Rollout-Id"),
		c.GetHeader("X-FindX-Config-Rollout-Id"),
		c.GetHeader("X-FindX-Rollout-Id"),
		c.GetHeader("X-FindX-Rollout-Ref"),
	)
	for key, value := range map[string]string{
		"source_rollout_id": rolloutRef,
		"request_ref": firstReceiverValue(
			c.Query("request_ref"),
			c.GetHeader("X-FindX-Request-Ref"),
		),
		"plugin_id": firstReceiverValue(
			c.Query("plugin_id"),
			c.GetHeader("X-FindX-Plugin-Id"),
		),
		"agent_ref": firstReceiverValue(
			c.Query("agent_ref"),
			c.Query("agent_id"),
			c.GetHeader("X-FindX-Agent-Ref"),
			c.GetHeader("X-FindX-Agent-Id"),
		),
		"cmdb_host_ref": firstReceiverValue(
			c.Query("cmdb_host_ref"),
			c.Query("target_id"),
			c.GetHeader("X-FindX-CMDB-Host-Ref"),
			c.GetHeader("X-FindX-Target-Id"),
		),
	} {
		if clean := sanitizeReceiverRuntimeRefValue(key, value); clean != "" {
			metadata[key] = clean
		}
	}
}

func sanitizeReceiverRuntimeRefValue(key, value string) string {
	clean := sanitizeReceiverValue(key, value)
	if clean == "" || receiverValueLooksFakeState(clean) {
		return ""
	}
	return clean
}

func receiverValueLooksFakeState(value string) bool {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case "queued", "running", "applied", "installed", "data_arrived", "service_registered",
		"rolled_back", "rolled-back", "uninstalled", "delivered", "effective", "succeeded", "success", "imported":
		return true
	default:
		return false
	}
}

func safeReceiverMetadata(groups ...map[string]string) map[string]string {
	out := map[string]string{}
	for _, group := range groups {
		for key, value := range group {
			if clean := sanitizeReceiverValue(key, value); clean != "" {
				out[strings.TrimSpace(key)] = clean
			}
		}
	}
	return out
}

func sanitizeReceiverValue(key, value string) string {
	if looksSensitive(key) || looksSensitive(value) {
		return ""
	}
	return sanitizeRemoteMutationValue(key, value)
}

func firstReceiverValue(values ...string) string {
	for _, value := range values {
		if clean := strings.TrimSpace(value); clean != "" {
			return clean
		}
	}
	return ""
}

func firstReceiverUnixTime(values ...int64) int64 {
	for _, value := range values {
		if value > 0 {
			return value
		}
	}
	return 0
}

func clientHost(remoteAddr string) string {
	host, _, err := net.SplitHostPort(remoteAddr)
	if err == nil {
		return host
	}
	return strings.TrimSpace(remoteAddr)
}

func validCategrafReceiverSource(c *gin.Context) bool {
	host := categrafReceiverClientHost(c)
	ip := net.ParseIP(host)
	if ip != nil && ip.IsLoopback() {
		return true
	}
	return validAgentToken(c)
}

func categrafReceiverClientHost(c *gin.Context) string {
	immediateHost := clientHost(c.Request.RemoteAddr)
	immediateIP := net.ParseIP(immediateHost)
	if immediateIP == nil || !immediateIP.IsLoopback() {
		return immediateHost
	}
	if value := forwardedClientHost(c.GetHeader("X-Real-IP"), false); value != "" {
		return value
	}
	if strings.TrimSpace(c.GetHeader("X-Real-IP")) != "" {
		return ""
	}
	if value := forwardedClientHost(c.GetHeader("X-Forwarded-For"), true); value != "" {
		return value
	}
	if strings.TrimSpace(c.GetHeader("X-Forwarded-For")) != "" {
		return ""
	}
	return clientHost(c.Request.RemoteAddr)
}

func forwardedClientHost(value string, last bool) string {
	parts := strings.Split(value, ",")
	if last {
		for i := len(parts) - 1; i >= 0; i-- {
			if host := forwardedClientHostPart(parts[i]); host != "" {
				return host
			}
		}
		return ""
	}
	for _, part := range parts {
		if host := forwardedClientHostPart(part); host != "" {
			return host
		}
	}
	return ""
}

func forwardedClientHostPart(value string) string {
	host := clientHost(strings.TrimSpace(value))
	if net.ParseIP(host) != nil {
		return host
	}
	return ""
}

func validCategrafProviderToken(c *gin.Context) bool {
	expected := strings.TrimSpace(os.Getenv("FINDX_AGENT_TOKEN"))
	if expected == "" {
		expected = strings.TrimSpace(viper.GetString("findx_agents.shared_token"))
	}
	if expected == "" {
		return false
	}
	actual := strings.TrimSpace(c.GetHeader("X-Agent-Token"))
	return subtle.ConstantTimeCompare([]byte(actual), []byte(expected)) == 1
}

func hasCategrafProviderTarget(c *gin.Context) bool {
	for _, key := range []string{"agent", "agent_id", "host", "agent_hostname", "target_id", "ident"} {
		if sanitizeReceiverValue(key, c.Query(key)) != "" {
			return true
		}
	}
	return false
}

func writeCategrafReceiverError(c *gin.Context, err error) {
	var tooLarge errReceiverBodyTooLarge
	if errors.As(err, &tooLarge) {
		c.JSON(http.StatusRequestEntityTooLarge, gin.H{"error": "request body too large"})
		return
	}
	c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
}

type errReceiverBodyTooLarge struct{}

func (e errReceiverBodyTooLarge) Error() string { return "request body too large" }


type findXAgentLogsPayload struct {
	AgentID  string                `json:"agent_id"`
	TargetID string                `json:"target_id"`
	Source   string                `json:"source"`
	Scope    string                `json:"scope"`
	Service  string                `json:"service"`
	TraceID  string                `json:"trace_id"`
	Records  []findXAgentLogRecord `json:"records"`
	Logs     []findXAgentLogRecord `json:"logs"`
	Metadata map[string]string     `json:"metadata"`
	Labels   map[string]string     `json:"labels"`
}

type findXAgentLogRecord struct {
	Body    string `json:"body"`
	Message string `json:"message"`
	Log     string `json:"log"`
}

type findXAgentTracesPayload struct {
	AgentID  string                  `json:"agent_id"`
	TargetID string                  `json:"target_id"`
	TraceID  string                  `json:"trace_id"`
	Source   string                  `json:"source"`
	Scope    string                  `json:"scope"`
	Service  string                  `json:"service"`
	Spans    []findXAgentSpanSummary `json:"spans"`
	Metadata map[string]string       `json:"metadata"`
	Labels   map[string]string       `json:"labels"`
}

type findXAgentSpanSummary struct {
	SpanID string `json:"span_id"`
}

func FindXAgentLogsCompatibleReceiver(c *gin.Context) {
	if !validCategrafReceiverSource(c) {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid agent receiver token"})
		return
	}
	payload, err := readFindXAgentLogsPayload(c)
	if err != nil {
		writeCategrafReceiverError(c, err)
		return
	}
	saved, err := store.SaveFindXAgentDataArrivalEvidence(payload.toEvidence(c))
	if err != nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "logs evidence persistence unavailable"})
		return
	}
	writeFindXAgentReceiverOK(c, saved)
}

func FindXAgentTracesCompatibleReceiver(c *gin.Context) {
	if !validCategrafReceiverSource(c) {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid agent receiver token"})
		return
	}
	payload, err := readFindXAgentTracesPayload(c)
	if err != nil {
		writeCategrafReceiverError(c, err)
		return
	}
	saved, err := store.SaveFindXAgentDataArrivalEvidence(payload.toEvidence(c))
	if err != nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "tracing evidence persistence unavailable"})
		return
	}
	writeFindXAgentReceiverOK(c, saved)
}

func readFindXAgentLogsPayload(c *gin.Context) (findXAgentLogsPayload, error) {
	body, err := readReceiverBody(c.Request.Body, c.GetHeader("Content-Encoding"), categrafReceiverBodyLimit)
	if err != nil {
		return findXAgentLogsPayload{}, err
	}
	if len(body) == 0 {
		return findXAgentLogsPayload{}, errBadRequest("logs payload body is required")
	}
	var payload findXAgentLogsPayload
	if err := json.Unmarshal(body, &payload); err != nil {
		return findXAgentLogsPayload{}, errBadRequest("invalid logs payload")
	}
	if err := validateFindXAgentLogsPayload(payload); err != nil {
		return findXAgentLogsPayload{}, err
	}
	return payload, nil
}

func readFindXAgentTracesPayload(c *gin.Context) (findXAgentTracesPayload, error) {
	body, err := readReceiverBody(c.Request.Body, c.GetHeader("Content-Encoding"), categrafReceiverBodyLimit)
	if err != nil {
		return findXAgentTracesPayload{}, err
	}
	if len(body) == 0 {
		return findXAgentTracesPayload{}, errBadRequest("traces payload body is required")
	}
	var payload findXAgentTracesPayload
	if err := json.Unmarshal(body, &payload); err != nil {
		return findXAgentTracesPayload{}, errBadRequest("invalid traces payload")
	}
	if err := validateFindXAgentTracesPayload(payload); err != nil {
		return findXAgentTracesPayload{}, err
	}
	return payload, nil
}

func validateFindXAgentLogsPayload(payload findXAgentLogsPayload) error {
	if firstReceiverValue(payload.AgentID, payload.TargetID) == "" {
		return errBadRequest("agent_id or target_id is required")
	}
	if findXAgentLogRecordCount(payload) == 0 {
		return errBadRequest("at least one log record is required")
	}
	for _, record := range append(payload.Records, payload.Logs...) {
		if firstReceiverValue(record.Body, record.Message, record.Log) != "" {
			return nil
		}
	}
	return errBadRequest("at least one log record body, message, or log is required")
}

func validateFindXAgentTracesPayload(payload findXAgentTracesPayload) error {
	if firstReceiverValue(payload.AgentID, payload.TargetID) == "" {
		return errBadRequest("agent_id or target_id is required")
	}
	if strings.TrimSpace(payload.TraceID) == "" {
		return errBadRequest("trace_id is required")
	}
	if len(payload.Spans) == 0 {
		return errBadRequest("at least one span is required")
	}
	for _, span := range payload.Spans {
		if strings.TrimSpace(span.SpanID) != "" {
			return nil
		}
	}
	return errBadRequest("at least one span_id is required")
}

func (p findXAgentLogsPayload) toEvidence(c *gin.Context) model.FindXAgentDataArrivalEvidence {
	return model.FindXAgentDataArrivalEvidence{
		Kind:         model.FindXAgentDataArrivalKindLogs,
		AgentID:      sanitizeReceiverValue("agent_id", p.AgentID),
		TargetID:     sanitizeReceiverValue("target_id", p.TargetID),
		Status:       model.FindXAgentDataArrivalStatusReported,
		EvidenceRefs: []string{"receiver:/findx-agent/logs-compatible"},
		Metadata:     p.metadata(c),
	}
}

func (p findXAgentTracesPayload) toEvidence(c *gin.Context) model.FindXAgentDataArrivalEvidence {
	return model.FindXAgentDataArrivalEvidence{
		Kind:         model.FindXAgentDataArrivalKindTracing,
		AgentID:      sanitizeReceiverValue("agent_id", p.AgentID),
		TargetID:     sanitizeReceiverValue("target_id", p.TargetID),
		Status:       model.FindXAgentDataArrivalStatusReported,
		EvidenceRefs: []string{"receiver:/findx-agent/traces-compatible"},
		Metadata:     p.metadata(c),
	}
}

func (p findXAgentLogsPayload) metadata(c *gin.Context) map[string]string {
	metadata := safeReceiverMetadata(p.Labels, p.Metadata)
	for key, value := range map[string]string{
		"count":     strconv.Itoa(findXAgentLogRecordCount(p)),
		"source":    firstReceiverValue(p.Source, "findx-agent-compatible"),
		"scope":     p.Scope,
		"service":   p.Service,
		"trace_id":  p.TraceID,
		"remote_ip": clientHost(c.Request.RemoteAddr),
	} {
		if clean := sanitizeReceiverValue(key, value); clean != "" {
			metadata[key] = clean
		}
	}
	return metadata
}

func (p findXAgentTracesPayload) metadata(c *gin.Context) map[string]string {
	metadata := safeReceiverMetadata(p.Labels, p.Metadata)
	for key, value := range map[string]string{
		"span_count": strconv.Itoa(len(p.Spans)),
		"trace_id":   p.TraceID,
		"source":     firstReceiverValue(p.Source, "findx-agent-compatible"),
		"scope":      p.Scope,
		"service":    p.Service,
		"remote_ip":  clientHost(c.Request.RemoteAddr),
	} {
		if clean := sanitizeReceiverValue(key, value); clean != "" {
			metadata[key] = clean
		}
	}
	return metadata
}

func findXAgentLogRecordCount(payload findXAgentLogsPayload) int {
	return len(payload.Records) + len(payload.Logs)
}

func writeFindXAgentReceiverOK(c *gin.Context, saved model.FindXAgentDataArrivalEvidence) {
	c.JSON(http.StatusOK, gin.H{
		"ok":          true,
		"status":      saved.Status,
		"evidence_id": saved.ID,
		"kind":        saved.Kind,
	})
}
