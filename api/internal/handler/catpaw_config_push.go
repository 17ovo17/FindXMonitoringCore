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

// CatpawPushConfig 批量推送插件配置到目标主机
func CatpawPushConfig(c *gin.Context) {
	var req model.CatpawDeployRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if len(req.TargetIPs) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "target_ips is required"})
		return
	}
	if len(req.Plugins) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "plugins is required"})
		return
	}

	// 解析凭据
	var cred RemoteCredential
	if req.CredentialID != "" {
		if !applySavedCredential(&cred, req.CredentialID) {
			c.JSON(http.StatusBadRequest, gin.H{"error": "credential_id not found"})
			return
		}
	}
	if strings.TrimSpace(cred.Username) == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "credential with username is required"})
		return
	}

	// 构建每个插件的配置内容
	pluginConfigs := make(map[string]string, len(req.Plugins))
	for _, pluginID := range req.Plugins {
		if customCfg, ok := req.CustomConfig[pluginID]; ok && strings.TrimSpace(customCfg) != "" {
			pluginConfigs[pluginID] = customCfg
		} else {
			plugin, ok := store.GetCatpawPlugin(pluginID)
			if !ok {
				c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("plugin %q not found", pluginID)})
				return
			}
			pluginConfigs[pluginID] = plugin.DefaultConfig
		}
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
			result := pushConfigToHost(targetIP, cred, pluginConfigs)
			mu.Lock()
			results = append(results, result)
			mu.Unlock()
		}(cleanedIP)
	}

	wg.Wait()
	c.JSON(http.StatusOK, gin.H{"results": results})
}

// pushConfigToHost 推送配置到单台主机
func pushConfigToHost(ip string, cred RemoteCredential, pluginConfigs map[string]string) model.CatpawDeployResult {
	hostCred := cred
	hostCred.IP = ip
	if hostCred.Port == 0 {
		hostCred.Port = 22
	}

	// 构建写入脚本：创建目录 → 写入每个插件 toml → 重启 catpaw
	var scriptBuilder strings.Builder
	scriptBuilder.WriteString("set -e\nmkdir -p /opt/catpaw/conf\n")

	for pluginID, config := range pluginConfigs {
		fileName := pluginID + ".toml"
		// 使用 heredoc 写入文件，避免转义问题
		escapedConfig := strings.ReplaceAll(config, "'", "'\\''")
		scriptBuilder.WriteString(fmt.Sprintf("cat > /opt/catpaw/conf/%s << 'PLUGIN_EOF'\n%s\nPLUGIN_EOF\n", fileName, escapedConfig))
	}

	scriptBuilder.WriteString("systemctl restart catpaw 2>/dev/null || (pkill -x catpaw 2>/dev/null; sleep 1; /usr/local/bin/catpaw run --configs /opt/catpaw/conf/ &)\n")
	scriptBuilder.WriteString("echo CONFIG_PUSH_OK\n")

	pushReq := RemoteExecRequest{RemoteCredential: hostCred, Command: scriptBuilder.String()}
	out, err := execSSH(pushReq)
	if err != nil {
		logrus.WithError(err).WithField("ip", ip).Warn("catpaw config push: failed")
		return model.CatpawDeployResult{IP: ip, Status: "failed", Message: fmt.Sprintf("配置推送失败: %v, output: %s", err, truncate(out, 200))}
	}

	if !strings.Contains(out, "CONFIG_PUSH_OK") {
		return model.CatpawDeployResult{IP: ip, Status: "failed", Message: fmt.Sprintf("配置推送未确认完成: %s", truncate(out, 200))}
	}

	logrus.WithField("ip", ip).WithField("plugins", len(pluginConfigs)).Info("catpaw config push: success")
	return model.CatpawDeployResult{IP: ip, Status: "success", Message: fmt.Sprintf("已推送 %d 个插件配置并重启", len(pluginConfigs))}
}
