package handler

import (
	"ai-workbench-api/internal/security"
	"ai-workbench-api/internal/store"

	"fmt"
	"net"
	"net/http"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

type RemoteCredential struct {
	IP       string `json:"ip"`
	Protocol string `json:"protocol"` // ssh | winrm
	Port     int    `json:"port"`
	Username string `json:"username"`
	Password string `json:"password"`
	SSHKey   string `json:"ssh_key"`
}

type RemoteExecRequest struct {
	RemoteCredential
	CredentialID  string `json:"credential_id"`
	Command       string `json:"command" binding:"required"`
	SafetyConfirm string `json:"safety_confirm"`
	TestBatchID   string `json:"test_batch_id"`
}

func CheckRemotePort(c *gin.Context) {
	var req struct {
		IP   string `json:"ip" binding:"required"`
		Port int    `json:"port" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if decision := security.ValidateNetworkProbeTarget(req.IP); !decision.Allowed {
		auditEvent(c, "remote.check_port", req.IP, decision.Level, "reject", decision.Reason, c.GetHeader("X-Test-Batch-Id"))
		c.JSON(http.StatusForbidden, gin.H{"error": decision.Reason, "safety": decision})
		return
	}
	address := fmt.Sprintf("%s:%d", req.IP, req.Port)
	conn, err := net.DialTimeout("tcp", address, 5*time.Second)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"reachable": false, "address": address, "error": err.Error()})
		return
	}
	conn.Close()
	c.JSON(http.StatusOK, gin.H{"reachable": true, "address": address})
}

func hasAsset(name string) bool {
	_, err := os.Stat("./assets/" + name)
	return err == nil
}

func isLocalRemoteTarget(host string) bool {
	trimmed := strings.TrimSpace(host)
	if trimmed == "" {
		return false
	}
	if strings.EqualFold(trimmed, "localhost") {
		return true
	}
	ip := net.ParseIP(trimmed)
	if ip == nil {
		return false
	}
	if ip.IsLoopback() {
		return true
	}
	interfaces, err := net.Interfaces()
	if err != nil {
		return false
	}
	for _, iface := range interfaces {
		addrs, err := iface.Addrs()
		if err != nil {
			continue
		}
		for _, addr := range addrs {
			var candidate net.IP
			switch value := addr.(type) {
			case *net.IPNet:
				candidate = value.IP
			case *net.IPAddr:
				candidate = value.IP
			}
			if candidate != nil && candidate.Equal(ip) {
				return true
			}
		}
	}
	return false
}

func execLocalShell(script string) (string, error) {
	cmd := exec.Command("bash", "-lc", script)
	out, err := cmd.CombinedOutput()
	if err != nil {
		return string(out), fmt.Errorf("本地脚本执行失败: %v", err)
	}
	return string(out), nil
}

func applySavedCredential(cred *RemoteCredential, credentialID string) bool {
	if strings.TrimSpace(credentialID) == "" {
		return false
	}
	saved, ok := store.GetCredential(credentialID)
	if !ok {
		return false
	}
	if cred.Protocol == "" {
		cred.Protocol = saved.Protocol
	}
	if cred.Port == 0 {
		cred.Port = saved.Port
	}
	if cred.Username == "" {
		cred.Username = saved.Username
	}
	if cred.Password == "" || cred.Password == "******" {
		cred.Password = saved.Password
	}
	if cred.SSHKey == "" || cred.SSHKey == "******" {
		cred.SSHKey = saved.SSHKey
	}
	return true
}

func requireSavedCredential(c *gin.Context, cred *RemoteCredential, credentialID string) bool {
	if strings.TrimSpace(credentialID) == "" {
		return true
	}
	if !applySavedCredential(cred, credentialID) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "credential_id not found"})
		return false
	}
	return true
}

// dispatchRemoteExec 根据协议分发远程执行并返回输出
func dispatchRemoteExec(req RemoteExecRequest) (string, error) {
	switch req.Protocol {
	case "wmi":
		return execWMI(req)
	case "winrm":
		return execWinRM(req)
	default:
		return execSSH(req)
	}
}

func RemoteExec(c *gin.Context) {
	var req RemoteExecRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if !requireSavedCredential(c, &req.RemoteCredential, req.CredentialID) {
		return
	}
	if strings.TrimSpace(req.IP) == "" || strings.TrimSpace(req.Username) == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ip and username are required"})
		return
	}
	if decision := security.ValidateRemoteHost(req.IP); !decision.Allowed {
		auditEvent(c, "remote.exec", req.IP, decision.Level, "reject", decision.Reason, req.TestBatchID)
		c.JSON(http.StatusForbidden, gin.H{"error": decision.Reason, "safety": decision})
		return
	}
	commandDecision := security.ValidateConfirm(security.ClassifyCommand(req.Command), req.SafetyConfirm)
	if !commandDecision.Allowed {
		status := http.StatusPreconditionRequired
		if commandDecision.Level == "L4" {
			status = http.StatusForbidden
		}
		auditEvent(c, "remote.exec", req.IP, commandDecision.Level, "reject", commandDecision.Reason, req.TestBatchID)
		c.JSON(status, gin.H{"error": commandDecision.Reason, "safety": commandDecision})
		return
	}
	auditEvent(c, "remote.exec", req.IP, commandDecision.Level, "allow", commandDecision.Reason, req.TestBatchID)
	out, err := dispatchRemoteExec(req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error(), "output": out})
		return
	}
	c.JSON(http.StatusOK, gin.H{"output": out})
}

// executeInstall 根据协议选择安装脚本并执行
func executeInstall(cred RemoteCredential, protocol, ip, reportURL, mode string) (string, error) {
	switch protocol {
	case "winrm":
		if !hasAsset("catpaw_windows_amd64.exe") {
			return "", fmt.Errorf("缺少 Windows 探针二进制: ./assets/catpaw_windows_amd64.exe")
		}
		script := buildWinRMInstallCmd(reportURL)
		return execWinRM(RemoteExecRequest{RemoteCredential: cred, Command: script})
	case "wmi":
		if !hasAsset("catpaw_windows_amd64.exe") {
			return "", fmt.Errorf("缺少 Windows 探针二进制: ./assets/catpaw_windows_amd64.exe")
		}
		script := buildWMIInstallScript(reportURL)
		return execWMI(RemoteExecRequest{RemoteCredential: cred, Command: script})
	default:
		script := buildInstallScript(ip, reportURL, mode)
		if protocol == "local" || isLocalRemoteTarget(ip) {
			return execLocalShell(script)
		}
		return execSSH(RemoteExecRequest{RemoteCredential: cred, Command: script})
	}
}

// InstallCatpaw 通过 SSH/WinRM 一键安装 catpaw 到目标机器
func InstallCatpaw(c *gin.Context) {
	var req struct {
		RemoteCredential
		CredentialID string `json:"credential_id"`
		Mode         string `json:"mode"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if !requireSavedCredential(c, &req.RemoteCredential, req.CredentialID) {
		return
	}
	if strings.TrimSpace(req.IP) == "" || strings.TrimSpace(req.Username) == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ip and username are required"})
		return
	}
	if decision := security.ValidateRemoteHost(req.IP); !decision.Allowed {
		c.JSON(http.StatusForbidden, gin.H{"error": decision.Reason, "safety": decision})
		return
	}
	if req.Mode == "" {
		req.Mode = "run"
	}
	reportURL := c.GetHeader("X-Platform-URL")
	if reportURL == "" {
		reportURL = "http://your-ai-workbench:8080"
	}
	if decision := security.ValidatePlatformURL(reportURL); !decision.Allowed {
		c.JSON(http.StatusForbidden, gin.H{"error": decision.Reason, "safety": decision})
		return
	}
	out, err := executeInstall(req.RemoteCredential, req.Protocol, req.IP, reportURL, req.Mode)
	if err != nil {
		if strings.HasPrefix(err.Error(), "缺少") {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error(), "output": out})
		return
	}
	c.JSON(http.StatusOK, gin.H{"output": out})
}

// GenerateInstallCmd 生成安装命令（不直接执行，供用户复制）
func GenerateInstallCmd(c *gin.Context) {
	var req struct {
		IP          string `json:"ip"`
		OS          string `json:"os"`
		ReportURL   string `json:"report_url"`
		PlatformURL string `json:"platform_url"`
		Protocol    string `json:"protocol"` // ssh | winrm | curl
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if req.ReportURL == "" {
		req.ReportURL = req.PlatformURL
	}
	if req.ReportURL == "" {
		req.ReportURL = "http://your-ai-workbench:8080"
	}
	if decision := security.ValidatePlatformURL(req.ReportURL); !decision.Allowed {
		c.JSON(http.StatusForbidden, gin.H{"error": decision.Reason, "safety": decision})
		return
	}

	var cmd string
	protocol := req.Protocol
	if protocol == "" && req.OS == "windows" {
		protocol = "winrm"
	}
	switch protocol {
	case "rdp-winrm", "winrm-bootstrap":
		cmd = buildRDPWinRMBootstrapCmd(req.ReportURL)
	case "wmi":
		cmd = buildWMIInstallScript(req.ReportURL)
	case "winrm":
		cmd = buildWinRMInstallCmd(req.ReportURL)
	default:
		cmd = buildCurlInstallCmd(req.ReportURL)
	}
	c.JSON(http.StatusOK, gin.H{"command": cmd})
}
