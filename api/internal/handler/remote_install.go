package handler

import (
	"fmt"
	"strings"
)

func buildRDPWinRMBootstrapCmd(reportURL string) string {
	return strings.TrimSpace(fmt.Sprintf(`# RDP 引导命令：请在目标 Windows 主机的“管理员 PowerShell”中执行
# 作用：启用 WinRM、开放远程管理防火墙、解除本地管理员远程 UAC 令牌过滤。
# 完成后回到平台选择 Windows + WinRM 安装 Catpaw。

$ErrorActionPreference = "Stop"
Write-Host "[1/6] Enable WinRM service"
Enable-PSRemoting -Force

Write-Host "[2/6] Configure WinRM for local administrator remote management"
winrm quickconfig -quiet
winrm set winrm/config/service '@{AllowUnencrypted="true"}'
winrm set winrm/config/service/auth '@{Basic="true"}'

Write-Host "[3/6] Open Windows firewall rules"
Enable-NetFirewallRule -DisplayGroup "Windows Remote Management" -ErrorAction SilentlyContinue
Enable-NetFirewallRule -DisplayGroup "Windows Management Instrumentation (WMI)" -ErrorAction SilentlyContinue
Enable-NetFirewallRule -DisplayGroup "Remote Service Management" -ErrorAction SilentlyContinue

Write-Host "[4/6] Disable remote UAC token filtering for local admin accounts"
New-Item -Path "HKLM:\SOFTWARE\Microsoft\Windows\CurrentVersion\Policies\System" -Force | Out-Null
Set-ItemProperty -Path "HKLM:\SOFTWARE\Microsoft\Windows\CurrentVersion\Policies\System" -Name LocalAccountTokenFilterPolicy -Type DWord -Value 1

Write-Host "[5/6] Verify WinRM listener"
winrm enumerate winrm/config/listener

Write-Host "[6/6] Platform URL for Catpaw callback: %s"
Write-Host "RDP bootstrap completed. Now install from AI WorkBench with protocol: WinRM, port: 5985."

# 平台侧也需要以管理员 PowerShell 执行一次（把 <TARGET_IP> 改成目标 IP，例如 192.168.1.7）：
# Set-Item -Path WSMan:\localhost\Client\TrustedHosts -Value "<TARGET_IP>" -Force
# Set-Item -Path WSMan:\localhost\Client\AllowUnencrypted -Value $true -Force
`, reportURL))
}

func buildInstallScript(ip, reportURL, mode string) string {
	_ = ip
	return strings.TrimSpace(fmt.Sprintf(`
set -e
ARCH=$(uname -m)
[ "$ARCH" = "x86_64" ] && ARCH=amd64 || ARCH=arm64
# 获取本机 IP
HOST_IP=$(hostname -I | awk '{print $1}')
pkill -x catpaw 2>/dev/null || true
sleep 1
curl -kfsSL "%s/download/catpaw_linux_${ARCH}" -o /tmp/catpaw.$$
chmod +x /tmp/catpaw.$$
mv /tmp/catpaw.$$ /usr/local/bin/catpaw
mkdir -p /etc/catpaw/conf.d
cat > /etc/catpaw/conf.d/config.toml << EOF
[global.labels]
from_hostip = "${HOST_IP}"

[notify.webapi]
enabled = true
url = "%s/api/v1/catpaw/report"
method = "POST"

[notify.heartbeat]
enabled = true
url = "%s/api/v1/catpaw/heartbeat"
interval = "60s"

[ai]
enabled = true
model_priority = []
max_rounds = 10
request_timeout = "120s"
max_retries = 1
retry_backoff = "2s"
tool_timeout = "20s"
queue_full_policy = "wait"
language = "zh"

[ai.gateway]
enabled = true
base_url = "%s/api/v1/agent/llm"
max_retries = 1
request_timeout = "120s"
fallback_to_direct = false
EOF
nohup catpaw --configs /etc/catpaw/conf.d %s > /var/log/catpaw.log 2>&1 &
echo "catpaw started (ip=${HOST_IP}) in %s mode"
`, reportURL, reportURL, reportURL, reportURL, mode, mode))
}

func buildCurlInstallCmd(reportURL string) string {
	return fmt.Sprintf(`ARCH=$(uname -m); [ "$ARCH" = "x86_64" ] && ARCH=amd64 || ARCH=arm64
HOST_IP=$(hostname -I | awk '{print $1}')
pkill -x catpaw 2>/dev/null || true
sleep 1
curl -kfsSL "%s/download/catpaw_linux_${ARCH}" -o /tmp/catpaw.$$ && chmod +x /tmp/catpaw.$$ && mv /tmp/catpaw.$$ /usr/local/bin/catpaw
mkdir -p /etc/catpaw/conf.d
cat > /etc/catpaw/conf.d/config.toml << EOF
[global.labels]
from_hostip = "${HOST_IP}"
[notify.webapi]
enabled = true
url = "%s/api/v1/catpaw/report"
method = "POST"
[notify.heartbeat]
enabled = true
url = "%s/api/v1/catpaw/heartbeat"
interval = "60s"

[ai]
enabled = true
model_priority = []
max_rounds = 10
request_timeout = "120s"
max_retries = 1
retry_backoff = "2s"
tool_timeout = "20s"
queue_full_policy = "wait"
language = "zh"

[ai.gateway]
enabled = true
base_url = "%s/api/v1/agent/llm"
max_retries = 1
request_timeout = "120s"
fallback_to_direct = false
EOF
nohup catpaw --configs /etc/catpaw/conf.d run > /var/log/catpaw.log 2>&1 &
echo "catpaw started (ip=${HOST_IP})"`, reportURL, reportURL, reportURL, reportURL)
}

func buildWMIInstallScript(reportURL string) string {
	return fmt.Sprintf(`$ErrorActionPreference = 'Stop'
$ProgressPreference = 'SilentlyContinue'
New-Item -ItemType Directory -Force -Path C:\catpaw\conf.d | Out-Null
$hostIP = (Get-NetIPAddress -AddressFamily IPv4 | Where-Object {$_.InterfaceAlias -notlike '*Loopback*' -and $_.IPAddress -notlike '169.254.*'} | Select-Object -First 1).IPAddress
certutil -urlcache -f "%s/download/catpaw_windows_amd64.exe" C:\catpaw\catpaw.exe | Out-Null
$cfg = @"
[global.labels]
from_hostip = "$hostIP"
[notify.webapi]
enabled = true
url = "%s/api/v1/catpaw/report"
method = "POST"
[notify.heartbeat]
enabled = true
url = "%s/api/v1/catpaw/heartbeat"
interval = "30s"
[ai.gateway]
enabled = true
base_url = "%s/api/v1/agent/llm"
fallback_to_direct = false
"@
[System.IO.File]::WriteAllText('C:\catpaw\conf.d\config.toml', $cfg, [System.Text.UTF8Encoding]::new($false))
$bat = @'
@echo off
cd /d C:\catpaw
C:\catpaw\catpaw.exe run --configs C:\catpaw\conf.d >> C:\catpaw\catpaw.stdout.log 2>> C:\catpaw\catpaw.stderr.log
'@
[System.IO.File]::WriteAllText('C:\catpaw\start-catpaw.bat', $bat, [System.Text.ASCIIEncoding]::new())
schtasks /End /TN Catpaw /F 2>$null
schtasks /Delete /TN Catpaw /F 2>$null
schtasks /Create /TN Catpaw /SC ONSTART /RL HIGHEST /F /TR 'C:\catpaw\start-catpaw.bat' | Out-Null
Start-Process -FilePath 'C:\catpaw\start-catpaw.bat' -WindowStyle Hidden
Start-Sleep -Seconds 3
if (-not (Get-Process catpaw -ErrorAction SilentlyContinue)) { throw "catpaw did not start" }
Write-Output "catpaw started"
`, reportURL, reportURL, reportURL, reportURL)
}

func buildWinRMInstallCmd(reportURL string) string {
	return fmt.Sprintf(`# PowerShell 安装（在目标机器上执行）
$hostIP = (Get-NetIPAddress -AddressFamily IPv4 | Where-Object {$_.InterfaceAlias -notlike '*Loopback*'} | Select-Object -First 1).IPAddress
New-Item -ItemType Directory -Force -Path C:\catpaw\conf.d | Out-Null
certutil -urlcache -f "%s/download/catpaw_windows_amd64.exe" C:\catpaw\catpaw.exe
@"
[global.labels]
from_hostip = "$hostIP"

[notify.webapi]
enabled = true
url = "%s/api/v1/catpaw/report"
method = "POST"

[notify.heartbeat]
enabled = true
url = "%s/api/v1/catpaw/heartbeat"
interval = "60s"

[ai]
enabled = true
model_priority = []
max_rounds = 10
request_timeout = "120s"
max_retries = 1
retry_backoff = "2s"
tool_timeout = "20s"
queue_full_policy = "wait"
language = "zh"

[ai.gateway]
enabled = true
base_url = "%s/api/v1/agent/llm"
max_retries = 1
request_timeout = "120s"
fallback_to_direct = false
"@ | Out-File C:\catpaw\conf.d\config.toml -Encoding UTF8
schtasks /End /TN Catpaw /F 2>$null
schtasks /Delete /TN Catpaw /F 2>$null
schtasks /Create /TN Catpaw /SC ONSTART /RL HIGHEST /F /TR 'C:\catpaw\catpaw.exe run --configs C:\catpaw\conf.d' | Out-Null
schtasks /Run /TN Catpaw | Out-Null
Start-Sleep -Seconds 2
if (-not (Get-Process catpaw -ErrorAction SilentlyContinue)) { throw "catpaw scheduled task did not start" }
Write-Output "catpaw scheduled task started"`, reportURL, reportURL, reportURL, reportURL)
}
