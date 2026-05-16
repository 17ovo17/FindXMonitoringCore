package handler

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"net/http"
	"strings"

	"ai-workbench-api/internal/model"
	"ai-workbench-api/internal/store"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"github.com/sirupsen/logrus"
)

var wsUpgrader = websocket.Upgrader{
	ReadBufferSize:  4096,
	WriteBufferSize: 4096,
	CheckOrigin: func(r *http.Request) bool {
		return false
	},
}

const cmdbHostInstanceMappingContract = "cmdb_host_instance_mapping_contract"

// CmdbHostTerminal WebSocket 终端端点。缺少真实 WebSSH 通道和 Origin 策略时拒绝升级。
func CmdbHostTerminal(c *gin.Context) {
	hostID := c.Param("id")
	if _, ok := resolveCmdbHostForOps(c, hostID, "cmdb.host.terminal.v1", []string{
		"webssh_channel_contract",
		"credential_ref_resolver",
		"websocket_origin_policy",
		"terminal_audit_contract",
	}); !ok {
		return
	}

	logrus.WithFields(logrus.Fields{
		"host_id": hostID,
		"origin":  c.Request.Header.Get("Origin"),
		"action":  "terminal_blocked",
	}).Warn("cmdb: terminal blocked by missing webssh contract")

	c.JSON(http.StatusConflict, cmdbHighRiskApprovalGate(c, cmdbHighRiskApprovalInput{
		ContractID:    "cmdb.host.terminal.v1",
		ResourceType:  "cmdb_host_terminal",
		ResourceID:    hostID,
		Action:        "terminal",
		RiskLevel:     "critical",
		Title:         "CMDB host terminal review",
		Summary:       "Open terminal request requires approval and runtime receipts before execution.",
		Reason:        "missing webssh channel, credential resolver, websocket origin policy and terminal audit contracts",
		Context:       map[string]any{"host_id": hostID, "origin_present": c.Request.Header.Get("Origin") != ""},
		Missing:       []string{"webssh_channel_contract", "credential_ref_resolver", "websocket_origin_policy", "terminal_audit_contract"},
		ExecutionState: "webssh session remains blocked",
	}))
}

// CmdbHostUpload 文件上传到目标主机。缺少真实远程 writer 时阻断。
func CmdbHostUpload(c *gin.Context) {
	hostID := c.Param("id")
	host, ok := resolveCmdbHostForOps(c, hostID, "cmdb.host.file_upload.v1", []string{
		"remote_file_writer_contract",
		"credential_ref_resolver",
		"artifact_checksum_contract",
		"upload_audit_contract",
	})
	if !ok {
		return
	}

	if _, _, err := c.Request.FormFile("file"); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "文件参数缺失"})
		return
	}
	connInfo := parseHostConnInfo(host.Data)
	logrus.WithFields(logrus.Fields{
		"host_id": hostID,
		"ip":      connInfo.IP,
		"action":  "file_upload_blocked",
	}).Warn("cmdb: file upload blocked by missing remote writer contract")

	c.JSON(http.StatusConflict, cmdbHighRiskApprovalGate(c, cmdbHighRiskApprovalInput{
		ContractID:   "cmdb.host.file_upload.v1",
		ResourceType: "cmdb_host_file_upload",
		ResourceID:   hostID,
		Action:       "upload",
		RiskLevel:    "high",
		Title:        "CMDB host file upload review",
		Summary:      "Remote file upload requires approval, artifact checksum and upload receipt contracts.",
		Reason:       "missing remote writer, credential resolver, checksum and upload audit contracts",
		Context: map[string]any{
			"host_id": hostID,
			"ip":      connInfo.IP,
		},
		Missing:        []string{"remote_file_writer_contract", "credential_ref_resolver", "artifact_checksum_contract", "upload_audit_contract"},
		ExecutionState: "remote writer remains blocked",
	}))
}

type execRequest struct {
	Command string `json:"command" binding:"required"`
	Timeout int    `json:"timeout"`
	Sudo    bool   `json:"sudo"`
}

// CmdbHostExec 远程命令执行。缺少真实执行器时阻断，不返回 exit_code=0。
func CmdbHostExec(c *gin.Context) {
	hostID := c.Param("id")
	host, ok := resolveCmdbHostForOps(c, hostID, "cmdb.host.command_exec.v1", []string{
		"remote_command_executor_contract",
		"credential_ref_resolver",
		"command_audit_contract",
		"stdout_stderr_capture_contract",
	})
	if !ok {
		return
	}

	var req execRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "请求参数无效"})
		return
	}
	if req.Timeout <= 0 {
		req.Timeout = 30
	}
	if req.Timeout > 300 {
		req.Timeout = 300
	}

	connInfo := parseHostConnInfo(host.Data)
	logrus.WithFields(logrus.Fields{
		"host_id":        hostID,
		"ip":             connInfo.IP,
		"command_length": len(req.Command),
		"command_digest": commandDigest(req.Command),
		"sudo":           req.Sudo,
		"timeout":        req.Timeout,
		"action":         "command_exec_blocked",
	}).Warn("cmdb: command execution blocked by missing executor contract")

	c.JSON(http.StatusConflict, cmdbHighRiskApprovalGate(c, cmdbHighRiskApprovalInput{
		ContractID:   "cmdb.host.command_exec.v1",
		ResourceType: "cmdb_host_command_exec",
		ResourceID:   hostID,
		Action:       "exec",
		RiskLevel:    "critical",
		Title:        "CMDB host command execution review",
		Summary:      "Remote command execution requires approval and command output receipts before execution.",
		Reason:       "missing remote command executor, credential resolver, command audit and stdout/stderr capture contracts",
		Context: map[string]any{
			"host_id":        hostID,
			"ip":             connInfo.IP,
			"command_length": len(req.Command),
			"command_digest": commandDigest(req.Command),
			"sudo":           req.Sudo,
			"timeout":        req.Timeout,
		},
		Missing:        []string{"remote_command_executor_contract", "credential_ref_resolver", "command_audit_contract", "stdout_stderr_capture_contract"},
		ExecutionState: "remote command executor remains blocked",
	}))
}

func resolveCmdbHostForOps(c *gin.Context, hostID string, contractID string, missing []string) (*model.CmdbInstance, bool) {
	host, ok := store.GetCmdbInstance(hostID)
	if ok {
		return host, true
	}
	if _, ok := store.GetMonitorTarget(hostID); ok {
		c.JSON(http.StatusConflict, cmdbBlockedContractEnvelope(
			contractID,
			cmdbHighRiskMissingContracts(append([]string{cmdbHostInstanceMappingContract}, missing...)...),
		))
		return nil, false
	}
	c.JSON(http.StatusNotFound, gin.H{"error": "主机不存在"})
	return nil, false
}

type hostConnInfo struct {
	IP       string
	Port     int
	Username string
	Hostname string
}

func parseHostConnInfo(dataJSON string) hostConnInfo {
	info := hostConnInfo{Port: 22, Username: "root"}
	if dataJSON == "" {
		return info
	}
	var data map[string]any
	if err := json.Unmarshal([]byte(dataJSON), &data); err != nil {
		return info
	}
	for _, key := range []string{"ip_address", "OS001", "ssh_ip", "sshIp"} {
		if ip, ok := data[key].(string); ok && ip != "" {
			info.IP = ip
			break
		}
	}
	if port := intFromAny(data["ssh_port"], 22); port > 0 {
		info.Port = port
	}
	if port := intFromAny(data["sshPort"], 22); port > 0 {
		info.Port = port
	}
	for _, key := range []string{"ssh_user", "sshName", "ssh_name"} {
		if user, ok := data[key].(string); ok && user != "" {
			info.Username = user
			break
		}
	}
	if name, ok := data["name"].(string); ok && name != "" {
		info.Hostname = name
	}
	return info
}

func sanitizeCommand(cmd string) string {
	if len(cmd) > 500 {
		return cmd[:500] + "..."
	}
	return cmd
}

func commandDigest(cmd string) string {
	sum := sha256.Sum256([]byte(cmd))
	return hex.EncodeToString(sum[:])
}

func sanitizeOutput(output string) string {
	replacer := strings.NewReplacer(
		"password=", "password=<REDACTED>",
		"passwd=", "passwd=<REDACTED>",
		"secret=", "secret=<REDACTED>",
		"token=", "token=<REDACTED>",
	)
	return replacer.Replace(output)
}
