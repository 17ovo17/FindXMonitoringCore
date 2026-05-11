package handler

import (
	"encoding/base64"
	"fmt"
	"net"
	"os/exec"
	"strings"
	"time"
	"unicode/utf16"

	"golang.org/x/crypto/ssh"
)

func encodePowerShell(script string) string {
	encoded := utf16.Encode([]rune(script))
	bytes := make([]byte, len(encoded)*2)
	for i, value := range encoded {
		bytes[i*2] = byte(value)
		bytes[i*2+1] = byte(value >> 8)
	}
	return base64.StdEncoding.EncodeToString(bytes)
}

func cleanPowerShellOutput(out []byte) string {
	text := string(out)
	text = strings.ReplaceAll(text, "#< CLIXML\r\n", "")
	text = strings.ReplaceAll(text, "#< CLIXML\n", "")
	if idx := strings.Index(text, "<Objs Version="); idx >= 0 {
		text = text[:idx]
	}
	return strings.TrimSpace(text)
}

func execWMI(req RemoteExecRequest) (string, error) {
	port := req.Port
	if port == 0 {
		port = 135
	}
	conn, err := net.DialTimeout("tcp", fmt.Sprintf("%s:%d", req.IP, port), 5*time.Second)
	if err != nil {
		return "", fmt.Errorf("WMI 连接失败: %s:%d 不可达: %v", req.IP, port, err)
	}
	conn.Close()
	localTarget := req.IP == "127.0.0.1" || strings.EqualFold(req.IP, "localhost")
	remoteCommand := fmt.Sprintf("powershell.exe -NoProfile -ExecutionPolicy Bypass -EncodedCommand %s", encodePowerShell(req.Command))
	script := ""
	if localTarget && strings.TrimSpace(req.Password) == "" {
		script = fmt.Sprintf(`
$ErrorActionPreference = 'Stop'
$ProgressPreference = 'SilentlyContinue'
[Console]::OutputEncoding = [System.Text.Encoding]::UTF8
$result = Invoke-WmiMethod -ComputerName %q -Class Win32_Process -Name Create -ArgumentList %q -ErrorAction Stop
if ($result.ReturnValue -ne 0) { throw "WMI process create failed: ReturnValue=$($result.ReturnValue)" }
Write-Output "WMI process started: ProcessId=$($result.ProcessId)"
`, req.IP, remoteCommand)
	} else {
		if strings.TrimSpace(req.Password) == "" {
			return "", fmt.Errorf("WMI 连接失败: password is required")
		}
		script = fmt.Sprintf(`
$ErrorActionPreference = 'Stop'
[Console]::OutputEncoding = [System.Text.Encoding]::UTF8
$secure = ConvertTo-SecureString %q -AsPlainText -Force
$cred = New-Object System.Management.Automation.PSCredential(%q, $secure)
$result = Invoke-WmiMethod -ComputerName %q -Credential $cred -Class Win32_Process -Name Create -ArgumentList %q -ErrorAction Stop
if ($result.ReturnValue -ne 0) { throw "WMI process create failed: ReturnValue=$($result.ReturnValue)" }
Write-Output "WMI process started: ProcessId=$($result.ProcessId)"
`, req.Password, req.Username, req.IP, remoteCommand)
	}
	cmd := exec.Command("powershell.exe", "-NoProfile", "-NonInteractive", "-ExecutionPolicy", "Bypass", "-EncodedCommand", encodePowerShell(script))
	out, err := cmd.CombinedOutput()
	if err != nil {
		return string(out), fmt.Errorf("WMI 连接失败: %v；请确认目标已开放 WMI/DCOM、445/135 与动态 RPC 端口、防火墙允许远程管理，且凭据属于本地管理员", err)
	}
	if strings.Contains(string(out), "拒绝访问") || strings.Contains(strings.ToLower(string(out)), "access is denied") || strings.Contains(string(out), "UnauthorizedAccessException") {
		return string(out), fmt.Errorf("WMI 连接失败: 目标拒绝访问；请确认用户名格式、管理员权限、远程 UAC 本地账号限制与 WMI/DCOM 权限")
	}
	return cleanPowerShellOutput(out), nil
}

func execSSH(req RemoteExecRequest) (string, error) {
	port := req.Port
	if port == 0 {
		port = 22
	}
	var auth []ssh.AuthMethod
	if req.SSHKey != "" {
		signer, err := ssh.ParsePrivateKey([]byte(req.SSHKey))
		if err != nil {
			return "", fmt.Errorf("解析 SSH 密钥失败: %v", err)
		}
		auth = append(auth, ssh.PublicKeys(signer))
	}
	if req.Password != "" {
		auth = append(auth, ssh.Password(req.Password))
	}
	cfg := &ssh.ClientConfig{
		User:            req.Username,
		Auth:            auth,
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
		Timeout:         30 * time.Second,
	}
	client, err := ssh.Dial("tcp", fmt.Sprintf("%s:%d", req.IP, port), cfg)
	if err != nil {
		return "", fmt.Errorf("SSH 连接失败: %v", err)
	}
	defer client.Close()

	session, err := client.NewSession()
	if err != nil {
		return "", err
	}
	defer session.Close()

	out, err := session.CombinedOutput(req.Command)
	return string(out), err
}

func execWinRM(req RemoteExecRequest) (string, error) {
	port := req.Port
	if port == 0 {
		port = 5985
	}
	conn, err := net.DialTimeout("tcp", fmt.Sprintf("%s:%d", req.IP, port), 5*time.Second)
	if err != nil {
		return "", fmt.Errorf("WinRM 连接失败: %s:%d 不可达: %v", req.IP, port, err)
	}
	conn.Close()
	if strings.TrimSpace(req.Password) == "" {
		return "", fmt.Errorf("WinRM 连接失败: password is required")
	}
	script := fmt.Sprintf(`
$ErrorActionPreference = 'Stop'
$ProgressPreference = 'SilentlyContinue'
[Console]::OutputEncoding = [System.Text.Encoding]::UTF8
$secure = ConvertTo-SecureString %q -AsPlainText -Force
$cred = New-Object System.Management.Automation.PSCredential(%q, $secure)
$sessionOption = New-PSSessionOption -SkipCACheck -SkipCNCheck -SkipRevocationCheck
$session = $null
$lastError = $null
foreach ($auth in @('Negotiate','Basic')) {
  try {
    $session = New-PSSession -ComputerName %q -Port %d -Credential $cred -Authentication $auth -SessionOption $sessionOption
    break
  } catch {
    $lastError = $_
  }
}
if (-not $session) { throw $lastError }
try {
  Invoke-Command -Session $session -ScriptBlock { %s }
} finally {
  if ($session) { Remove-PSSession $session }
}
`, req.Password, req.Username, req.IP, port, req.Command)
	cmd := exec.Command("powershell.exe", "-NoProfile", "-NonInteractive", "-ExecutionPolicy", "Bypass", "-EncodedCommand", encodePowerShell(script))
	out, err := cmd.CombinedOutput()
	if err != nil {
		return string(out), fmt.Errorf("WinRM 连接失败: %v；如目标为 Windows/IP 直连，请确认目标端已执行 RDP 引导，且平台侧已以管理员执行 WinRM 客户端配置：TrustedHosts、AllowUnencrypted", err)
	}
	return cleanPowerShellOutput(out), nil
}
