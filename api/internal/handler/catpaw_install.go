package handler

import (
	"fmt"
	"net/http"
	"strings"
	"sync"

	"ai-workbench-api/internal/model"
	"ai-workbench-api/internal/security"
	"ai-workbench-api/internal/store"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

// CatpawBatchInstall 批量安装 catpaw agent 到目标主机
func CatpawBatchInstall(c *gin.Context) {
	var req struct {
		TargetIPs    []string `json:"target_ips" binding:"required"`
		CredentialID string   `json:"credential_id"`
		Credential   RemoteCredential `json:"credential"`
		Mode         string   `json:"mode"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if len(req.TargetIPs) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "target_ips is required"})
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

	cred := req.Credential
	if req.CredentialID != "" {
		applySavedCredential(&cred, req.CredentialID)
	}
	if strings.TrimSpace(cred.Username) == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "username is required (via credential or credential_id)"})
		return
	}

	var (
		mu      sync.Mutex
		results []model.CatpawDeployResult
		wg      sync.WaitGroup
	)

	for _, ip := range req.TargetIPs {
		cleanedIP, ok := cleanIP(ip)
		if !ok {
			mu.Lock()
			results = append(results, model.CatpawDeployResult{IP: ip, Status: "failed", Message: "invalid IP address"})
			mu.Unlock()
			continue
		}

		if decision := security.ValidateRemoteHost(cleanedIP); !decision.Allowed {
			mu.Lock()
			results = append(results, model.CatpawDeployResult{IP: cleanedIP, Status: "failed", Message: decision.Reason})
			mu.Unlock()
			continue
		}

		wg.Add(1)
		go func(targetIP string) {
			defer wg.Done()
			result := installCatpawOnHost(targetIP, cred, reportURL, req.Mode)
			mu.Lock()
			results = append(results, result)
			mu.Unlock()
		}(cleanedIP)
	}

	wg.Wait()
	c.JSON(http.StatusOK, gin.H{"results": results})
}

// installCatpawOnHost 在单台主机上安装 catpaw
func installCatpawOnHost(ip string, cred RemoteCredential, reportURL, mode string) model.CatpawDeployResult {
	hostCred := cred
	hostCred.IP = ip
	if hostCred.Port == 0 {
		hostCred.Port = 22
	}

	// 检查是否已安装
	checkReq := RemoteExecRequest{RemoteCredential: hostCred, Command: "test -f /usr/local/bin/catpaw && echo INSTALLED || echo NOT_INSTALLED"}
	checkOut, err := execSSH(checkReq)
	if err != nil {
		logrus.WithError(err).WithField("ip", ip).Warn("catpaw install: SSH check failed")
		return model.CatpawDeployResult{IP: ip, Status: "failed", Message: fmt.Sprintf("SSH 连接失败: %v", err)}
	}

	alreadyInstalled := strings.Contains(strings.TrimSpace(checkOut), "INSTALLED")

	// 执行安装脚本
	script := buildCatpawInstallScript(ip, reportURL, mode, alreadyInstalled)
	installReq := RemoteExecRequest{RemoteCredential: hostCred, Command: script}
	out, err := execSSH(installReq)
	if err != nil {
		logrus.WithError(err).WithField("ip", ip).Warn("catpaw install: install script failed")
		return model.CatpawDeployResult{IP: ip, Status: "failed", Message: fmt.Sprintf("安装失败: %v, output: %s", err, truncate(out, 200))}
	}

	logrus.WithField("ip", ip).Info("catpaw install: success")
	msg := "安装成功"
	if alreadyInstalled {
		msg = "已更新并重启"
	}
	return model.CatpawDeployResult{IP: ip, Status: "success", Message: msg}
}

// buildCatpawInstallScript 构建安装脚本
func buildCatpawInstallScript(ip, reportURL, mode string, alreadyInstalled bool) string {
	_ = ip
	stopCmd := ""
	if alreadyInstalled {
		stopCmd = "systemctl stop catpaw 2>/dev/null || pkill -x catpaw 2>/dev/null || true; sleep 1"
	}

	return strings.TrimSpace(fmt.Sprintf(`set -e
ARCH=$(uname -m)
[ "$ARCH" = "x86_64" ] && ARCH=amd64 || ARCH=arm64
HOST_IP=$(hostname -I | awk '{print $1}')
%s
curl -kfsSL "%s/download/catpaw_linux_${ARCH}" -o /tmp/catpaw.$$
chmod +x /tmp/catpaw.$$
mv /tmp/catpaw.$$ /usr/local/bin/catpaw
mkdir -p /opt/catpaw/conf

cat > /opt/catpaw/conf/config.toml << 'CATPAW_EOF'
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
CATPAW_EOF

sed -i "s/\${HOST_IP}/$HOST_IP/g" /opt/catpaw/conf/config.toml

cat > /etc/systemd/system/catpaw.service << 'SYSTEMD_EOF'
[Unit]
Description=Catpaw Inspection Agent
After=network-online.target
Wants=network-online.target

[Service]
Type=simple
ExecStart=/usr/local/bin/catpaw %s --configs /opt/catpaw/conf/
Restart=always
RestartSec=5
LimitNOFILE=65536

[Install]
WantedBy=multi-user.target
SYSTEMD_EOF

systemctl daemon-reload
systemctl enable catpaw
systemctl restart catpaw
echo "CATPAW_INSTALL_OK"
`, stopCmd, reportURL, reportURL, reportURL, mode))
}

// ListCatpawPlugins 列出所有可用插件模板
func ListCatpawPlugins(c *gin.Context) {
	category := c.Query("category")
	if category != "" {
		c.JSON(http.StatusOK, store.ListCatpawPluginsByCategory(category))
		return
	}
	c.JSON(http.StatusOK, store.ListCatpawPlugins())
}

// GetCatpawPluginDetail 获取单个插件详情
func GetCatpawPluginDetail(c *gin.Context) {
	id := c.Param("id")
	plugin, ok := store.GetCatpawPlugin(id)
	if !ok {
		c.JSON(http.StatusNotFound, gin.H{"error": "plugin not found"})
		return
	}
	c.JSON(http.StatusOK, plugin)
}

func truncate(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen] + "..."
}
