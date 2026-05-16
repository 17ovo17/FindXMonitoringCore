package handler

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"sync"
	"time"

	"ai-workbench-api/internal/model"
	"ai-workbench-api/internal/store"

	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	"golang.org/x/crypto/ssh"
)

const (
	categrafBaseDir     = "/opt/categraf/conf"
	categrafReloadCmd   = "kill -HUP $(pidof categraf) 2>/dev/null || systemctl reload categraf 2>/dev/null || true"
	deploySSHTimeout    = 30 * time.Second
	verifyPollInterval  = 5 * time.Second
	verifyDefaultTimeout = 60 * time.Second
)

// DeployCategrafConfig 部署 categraf 配置到目标主机
// POST /api/v1/categraf/deploy
func DeployCategrafConfig(c *gin.Context) {
	var req model.CategrafDeployRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// 获取模板
	tmpl, ok := integrationTemplateMap[req.TemplateID]
	if !ok {
		c.JSON(http.StatusNotFound, gin.H{"error": fmt.Sprintf("模板 %s 不存在", req.TemplateID)})
		return
	}
	// 验证参数
	if err := validateTemplateParams(tmpl.Params, req.Params); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// 填充默认值并渲染
	params := applyDefaultParams(tmpl.Params, req.Params)
	rendered, err := renderTomlTemplate(tmpl.ID, tmpl.TomlTemplate, params)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "模板渲染失败: " + err.Error()})
		return
	}

	// 获取凭据
	cred, ok := store.GetCredential(req.CredentialID)
	if !ok {
		c.JSON(http.StatusBadRequest, gin.H{"error": "credential_id 不存在"})
		return
	}

	port := req.Port
	if port == 0 {
		port = cred.Port
	}
	if port == 0 {
		port = 22
	}

	pluginDir := pluginDirName(req.TemplateID)
	remotePath := fmt.Sprintf("%s/%s/%s.toml", categrafBaseDir, pluginDir, req.TemplateID)

	// 并发部署到各目标
	var wg sync.WaitGroup
	results := make([]model.CategrafDeployResult, len(req.TargetIPs))

	for i, ip := range req.TargetIPs {
		wg.Add(1)
		go func(idx int, targetIP string) {
			defer wg.Done()
			result := deployToHost(targetIP, port, *cred, rendered, pluginDir, remotePath)
			results[idx] = result
		}(i, ip)
	}
	wg.Wait()

	// 统计结果
	successCount := 0
	for _, r := range results {
		if r.Status == "success" {
			successCount++
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"results":       results,
		"total":         len(results),
		"success_count": successCount,
		"failed_count":  len(results) - successCount,
		"template_id":   req.TemplateID,
		"plugin_dir":    pluginDir,
	})
}

// deployToHost 通过 SSH 部署配置到单台主机
func deployToHost(ip string, port int, cred model.Credential, config, pluginDir, remotePath string) model.CategrafDeployResult {
	result := model.CategrafDeployResult{IP: ip}

	sshClient, err := buildSSHClient(ip, port, cred)
	if err != nil {
		result.Status = "failed"
		result.Message = fmt.Sprintf("SSH 连接失败: %v", err)
		log.WithError(err).WithField("ip", ip).Error("categraf 部署: SSH 连接失败")
		return result
	}
	defer sshClient.Close()

	// 创建目录
	mkdirCmd := fmt.Sprintf("mkdir -p %s/%s", categrafBaseDir, pluginDir)
	if err := runSSHCommand(sshClient, mkdirCmd); err != nil {
		result.Status = "failed"
		result.Message = fmt.Sprintf("创建目录失败: %v", err)
		return result
	}

	// 写入配置文件
	writeCmd := fmt.Sprintf("cat > %s << 'CATEGRAF_EOF'\n%s\nCATEGRAF_EOF", remotePath, config)
	if err := runSSHCommand(sshClient, writeCmd); err != nil {
		result.Status = "failed"
		result.Message = fmt.Sprintf("写入配置失败: %v", err)
		return result
	}

	// 重载 categraf
	if err := runSSHCommand(sshClient, categrafReloadCmd); err != nil {
		result.Status = "failed"
		result.Message = fmt.Sprintf("重载 categraf 失败: %v", err)
		return result
	}

	result.Status = "success"
	result.Message = fmt.Sprintf("配置已部署到 %s", remotePath)
	log.WithFields(log.Fields{"ip": ip, "path": remotePath}).Info("categraf 配置部署成功")
	return result
}

// buildSSHClient 构建 SSH 客户端连接
func buildSSHClient(ip string, port int, cred model.Credential) (*ssh.Client, error) {
	var authMethods []ssh.AuthMethod
	if cred.SSHKey != "" {
		signer, err := ssh.ParsePrivateKey([]byte(cred.SSHKey))
		if err != nil {
			return nil, fmt.Errorf("解析 SSH 密钥失败: %w", err)
		}
		authMethods = append(authMethods, ssh.PublicKeys(signer))
	}
	if cred.Password != "" {
		authMethods = append(authMethods, ssh.Password(cred.Password))
	}

	cfg := &ssh.ClientConfig{
		User:            cred.Username,
		Auth:            authMethods,
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
		Timeout:         deploySSHTimeout,
	}

	addr := fmt.Sprintf("%s:%d", ip, port)
	client, err := ssh.Dial("tcp", addr, cfg)
	if err != nil {
		return nil, fmt.Errorf("连接 %s 失败: %w", addr, err)
	}
	return client, nil
}

// runSSHCommand 在 SSH 连接上执行命令
func runSSHCommand(client *ssh.Client, cmd string) error {
	session, err := client.NewSession()
	if err != nil {
		return fmt.Errorf("创建 SSH 会话失败: %w", err)
	}
	defer session.Close()

	output, err := session.CombinedOutput(cmd)
	if err != nil {
		return fmt.Errorf("命令执行失败: %w, output: %s", err, strings.TrimSpace(string(output)))
	}
	return nil
}

// VerifyCategrafMetricArrival 验证 categraf 指标是否到达 Prometheus
// POST /api/v1/categraf/verify-arrival
func VerifyCategrafMetricArrival(c *gin.Context) {
	var req model.CategrafVerifyArrivalRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	timeout := time.Duration(req.TimeoutSec) * time.Second
	if timeout <= 0 {
		timeout = verifyDefaultTimeout
	}

	promURL := strings.TrimSpace(viper.GetString("prometheus.url"))
	if promURL == "" {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "Prometheus 未配置"})
		return
	}

	ctx, cancel := context.WithTimeout(c.Request.Context(), timeout)
	defer cancel()

	result := pollMetricArrival(ctx, promURL, req.TargetIP, req.MetricPrefix)
	c.JSON(http.StatusOK, gin.H{"result": result})
}

// pollMetricArrival 轮询 Prometheus 检查指标是否到达
func pollMetricArrival(ctx context.Context, promURL, targetIP, metricPrefix string) model.CategrafVerifyArrivalResult {
	// 构造 PromQL 查询: {__name__=~"prefix.*", instance=~".*targetIP.*"}
	query := fmt.Sprintf(`{__name__=~"%s.*",instance=~".*%s.*"}`, metricPrefix, targetIP)

	ticker := time.NewTicker(verifyPollInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return model.CategrafVerifyArrivalResult{
				Arrived: false,
				Message: fmt.Sprintf("超时未检测到指标 (prefix=%s, ip=%s)", metricPrefix, targetIP),
			}
		case <-ticker.C:
			metrics, err := queryPrometheusInstant(promURL, query)
			if err != nil {
				log.WithError(err).Debug("categraf verify-arrival: prometheus 查询失败，继续重试")
				continue
			}
			if len(metrics) > 0 {
				names := extractMetricNames(metrics)
				sample := ""
				if len(names) > 0 {
					sample = names[0]
				}
				return model.CategrafVerifyArrivalResult{
					Arrived:      true,
					MetricCount:  len(names),
					SampleMetric: sample,
					Metrics:      truncateSlice(names, 20),
					Message:      fmt.Sprintf("检测到 %d 个指标", len(names)),
				}
			}
		}
	}
}

// queryPrometheusInstant 执行 Prometheus 即时查询
func queryPrometheusInstant(baseURL, query string) ([]map[string]interface{}, error) {
	url := fmt.Sprintf("%s/api/v1/query?query=%s", strings.TrimRight(baseURL, "/"), query)

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var promResp struct {
		Status string `json:"status"`
		Data   struct {
			ResultType string                   `json:"resultType"`
			Result     []map[string]interface{} `json:"result"`
		} `json:"data"`
	}
	if err := json.Unmarshal(body, &promResp); err != nil {
		return nil, err
	}
	if promResp.Status != "success" {
		return nil, fmt.Errorf("prometheus 返回非 success 状态: %s", promResp.Status)
	}
	return promResp.Data.Result, nil
}

// extractMetricNames 从 Prometheus 查询结果中提取指标名
func extractMetricNames(results []map[string]interface{}) []string {
	seen := make(map[string]struct{})
	var names []string
	for _, r := range results {
		metric, ok := r["metric"].(map[string]interface{})
		if !ok {
			continue
		}
		name, ok := metric["__name__"].(string)
		if !ok || name == "" {
			continue
		}
		if _, exists := seen[name]; !exists {
			seen[name] = struct{}{}
			names = append(names, name)
		}
	}
	return names
}

// truncateSlice 截断切片到指定长度
func truncateSlice(s []string, max int) []string {
	if len(s) <= max {
		return s
	}
	return s[:max]
}
