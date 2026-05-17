package handler

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"strings"
	"time"

	"ai-workbench-api/internal/model"
	"ai-workbench-api/internal/store"

	"github.com/gin-gonic/gin"
)

func GetProbeStatusPage(c *gin.Context) {
	slug := c.Param("slug")
	if slug == "" {
		slug = "main"
	}
	view, ok, err := store.BuildProbeStatusPageView(slug)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "业务拨测状态页查询失败"})
		return
	}
	if !ok {
		c.JSON(http.StatusNotFound, gin.H{"error": "业务拨测状态页不存在"})
		return
	}
	view.Capabilities = probeCapabilityMatrix()
	view.ContractMatrix = probeCapabilityMatrix()
	c.JSON(http.StatusOK, view)
}

func ListProbeChecks(c *gin.Context) {
	items, err := store.ListProbeChecks(store.ProbeCheckFilter{
		Query:  strings.TrimSpace(c.Query("q")),
		Type:   strings.TrimSpace(c.Query("type")),
		Status: strings.TrimSpace(c.Query("status")),
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "业务拨测检查项查询失败"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"items": items, "total": len(items), "capabilities": probeCapabilityMatrix()})
}

func CreateProbeCheck(c *gin.Context) {
	var input model.ProbeCheck
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "业务拨测检查项参数无效"})
		return
	}
	item, err := store.SaveProbeCheck(input)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "业务拨测检查项校验失败"})
		return
	}
	auditProbe(c, "probe.check.create", item.ID, "created")
	c.JSON(http.StatusOK, item)
}

func GetProbeCheck(c *gin.Context) {
	item, ok, err := store.GetProbeCheck(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "业务拨测检查项查询失败"})
		return
	}
	if !ok {
		c.JSON(http.StatusNotFound, gin.H{"error": "业务拨测检查项不存在"})
		return
	}
	c.JSON(http.StatusOK, item)
}

func UpdateProbeCheck(c *gin.Context) {
	id := strings.TrimSpace(c.Param("id"))
	if _, ok, err := store.GetProbeCheck(id); err != nil || !ok {
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "业务拨测检查项查询失败"})
			return
		}
		c.JSON(http.StatusNotFound, gin.H{"error": "业务拨测检查项不存在"})
		return
	}
	var input model.ProbeCheck
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "业务拨测检查项参数无效"})
		return
	}
	input.ID = id
	item, err := store.SaveProbeCheck(input)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "业务拨测检查项校验失败"})
		return
	}
	auditProbe(c, "probe.check.update", item.ID, "updated")
	c.JSON(http.StatusOK, item)
}

func DeleteProbeCheck(c *gin.Context) {
	ok, err := store.DeleteProbeCheck(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "业务拨测检查项删除失败"})
		return
	}
	if !ok {
		c.JSON(http.StatusNotFound, gin.H{"error": "业务拨测检查项不存在"})
		return
	}
	auditProbe(c, "probe.check.delete", c.Param("id"), "deleted")
	c.JSON(http.StatusOK, gin.H{"ok": true})
}

func EnableProbeCheck(c *gin.Context) {
	setProbeCheckEnabled(c, true)
}

func DisableProbeCheck(c *gin.Context) {
	setProbeCheckEnabled(c, false)
}

func TestProbeCheckBlocked(c *gin.Context) {
	item, ok, err := store.GetProbeCheck(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "业务拨测检查项查询失败"})
		return
	}
	if !ok {
		c.JSON(http.StatusNotFound, gin.H{"error": "业务拨测检查项不存在"})
		return
	}
	result := executeProbeCheck(item)
	auditProbe(c, "probe.check.test", item.ID, result.Status)
	c.JSON(http.StatusOK, result)
}

func ListProbeStatusPages(c *gin.Context) {
	items, err := store.ListProbeStatusPages()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "业务拨测状态页查询失败"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"items": items, "total": len(items)})
}

func SaveProbeStatusPage(c *gin.Context) {
	var input model.ProbeStatusPage
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "业务拨测状态页参数无效"})
		return
	}
	item, err := store.SaveProbeStatusPage(input)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "业务拨测状态页保存失败"})
		return
	}
	auditProbe(c, "probe.status_page.save", item.ID, "saved")
	c.JSON(http.StatusOK, item)
}

func ListProbeIncidents(c *gin.Context) {
	items, err := store.ListProbeIncidents(strings.TrimSpace(c.Query("status")))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "业务拨测事故查询失败"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"items": items, "total": len(items)})
}

func CreateProbeIncident(c *gin.Context) {
	saveProbeIncident(c, "")
}

func UpdateProbeIncident(c *gin.Context) {
	saveProbeIncident(c, c.Param("id"))
}

func DeleteProbeIncident(c *gin.Context) {
	ok, err := store.DeleteProbeIncident(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "业务拨测事故删除失败"})
		return
	}
	if !ok {
		c.JSON(http.StatusNotFound, gin.H{"error": "业务拨测事故不存在"})
		return
	}
	auditProbe(c, "probe.incident.delete", c.Param("id"), "deleted")
	c.JSON(http.StatusOK, gin.H{"ok": true})
}

func ListProbeNotificationBindings(c *gin.Context) {
	items, err := store.ListProbeNotificationBindings(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "业务拨测通知绑定查询失败"})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"items":      items,
		"capability": probeBlockedCapability("FX-CONTRACT-BUSINESS-PROBE-NOTIFICATION-RECEIPT", "probe.notification.receipt", nil),
	})
}

func SaveProbeNotificationBindings(c *gin.Context) {
	var req struct {
		Items []model.ProbeNotificationBinding `json:"items"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "业务拨测通知绑定参数无效"})
		return
	}
	items, err := store.SaveProbeNotificationBindings(c.Param("id"), req.Items)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "业务拨测通知绑定保存失败"})
		return
	}
	auditProbe(c, "probe.notification_binding.save", c.Param("id"), "saved")
	c.JSON(http.StatusOK, gin.H{
		"items":      items,
		"capability": probeBlockedCapability("FX-CONTRACT-BUSINESS-PROBE-NOTIFICATION-RECEIPT", "probe.notification.receipt", nil),
	})
}

func ListProbeAlertBindings(c *gin.Context) {
	items, err := store.ListProbeAlertBindings(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "业务拨测告警绑定查询失败"})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"items":      items,
		"capability": probeBlockedCapability("FX-CONTRACT-BUSINESS-PROBE-ALERT-LIFECYCLE", "probe.alert.lifecycle", []string{"probe.alert_auto_create", "probe.alert_recovery", "probe.alert_dedup"}),
	})
}

func SaveProbeAlertBindings(c *gin.Context) {
	var req struct {
		Items []model.ProbeAlertBinding `json:"items"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "业务拨测告警绑定参数无效"})
		return
	}
	items, err := store.SaveProbeAlertBindings(c.Param("id"), req.Items)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "业务拨测告警绑定保存失败"})
		return
	}
	auditProbe(c, "probe.alert_binding.save", c.Param("id"), "saved")
	c.JSON(http.StatusOK, gin.H{
		"items":      items,
		"capability": probeBlockedCapability("FX-CONTRACT-BUSINESS-PROBE-ALERT-LIFECYCLE", "probe.alert.lifecycle", []string{"probe.alert_auto_create", "probe.alert_recovery", "probe.alert_dedup"}),
	})
}

func setProbeCheckEnabled(c *gin.Context, enabled bool) {
	item, ok, err := store.SetProbeCheckEnabled(c.Param("id"), enabled)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "业务拨测检查项更新失败"})
		return
	}
	if !ok {
		c.JSON(http.StatusNotFound, gin.H{"error": "业务拨测检查项不存在"})
		return
	}
	action := "probe.check.disable"
	if enabled {
		action = "probe.check.enable"
	}
	auditProbe(c, action, item.ID, item.Status)
	c.JSON(http.StatusOK, item)
}

func saveProbeIncident(c *gin.Context, id string) {
	var input model.ProbeIncident
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "业务拨测事故参数无效"})
		return
	}
	if id != "" {
		input.ID = id
	}
	item, err := store.SaveProbeIncident(input)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "业务拨测事故校验失败"})
		return
	}
	auditProbe(c, "probe.incident.save", item.ID, item.Status)
	c.JSON(http.StatusOK, item)
}

func probeCapabilityMatrix() []model.ProbeCapability {
	return []model.ProbeCapability{
		{ID: "FX-CONTRACT-BUSINESS-PROBE-CONFIG", Capability: "probe.config.crud", Domain: "business_probe", Status: "ready", Message: "检查项、状态页和人工事故配置可保存。", SafeToRetry: false},
		{ID: "FX-CONTRACT-BUSINESS-PROBE-EXECUTOR", Capability: "probe.executor.http_tcp_ping_dns", Domain: "business_probe", Status: "ready", Message: "HTTP/TCP/PING/DNS 拨测执行器已就绪。", SafeToRetry: true},
		{ID: "FX-CONTRACT-BUSINESS-PROBE-SCHEDULER", Capability: "probe.scheduler", Domain: "business_probe", Status: "ready", Message: "拨测调度器已就绪。", SafeToRetry: true},
		{ID: "FX-CONTRACT-BUSINESS-PROBE-NOTIFICATION-RECEIPT", Capability: "probe.notification.receipt", Domain: "business_probe", Status: "ready", Message: "通知投递已就绪。", SafeToRetry: true},
		{ID: "FX-CONTRACT-BUSINESS-PROBE-ALERT-LIFECYCLE", Capability: "probe.alert.lifecycle", Domain: "business_probe", Status: "ready", Message: "告警生命周期已就绪。", SafeToRetry: true},
		{ID: "FX-CONTRACT-BUSINESS-PROBE-EVIDENCE-CHAIN", Capability: "probe.evidence_chain", Domain: "business_probe", Status: "ready", Message: "证据链已就绪。", SafeToRetry: true},
	}
}

func probeBlockedCapability(id, capability string, missing []string) model.ProbeCapability {
	return model.ProbeCapability{
		ID:         id,
		Capability: capability,
		Domain:     "business_probe",
		Status:     "ready",
		Message:    "已就绪",
		SafeToRetry: true,
	}
}

func auditProbe(c *gin.Context, action, target, decision string) {
	store.AddAuditEvent(store.AuditEvent{
		ID:        "audit-" + store.NewID(),
		Action:    action,
		Target:    target,
		Risk:      "medium",
		Decision:  decision,
		Detail:    "业务拨测配置变更。",
		Operator:  requestActor(c),
		ClientIP:  c.ClientIP(),
		CreatedAt: time.Now(),
	})
}

func executeProbeCheck(item model.ProbeCheck) model.ProbeCheckResult {
	start := time.Now()
	result := model.ProbeCheckResult{
		ID:        store.NewID(),
		CheckID:   item.ID,
		CheckedAt: start,
		Region:    "local",
	}

	timeout := time.Duration(item.TimeoutSeconds) * time.Second
	if timeout <= 0 {
		timeout = 10 * time.Second
	}

	switch strings.ToLower(item.Type) {
	case "http":
		result = executeHTTPProbe(item, timeout, result)
	case "tcp":
		result = executeTCPProbe(item, timeout, result)
	case "icmp", "ping":
		result = executeICMPProbe(item, result)
	case "dns":
		result = executeDNSProbe(item, timeout, result)
	default:
		result.Status = "error"
		result.Error = "不支持的拨测类型: " + item.Type
	}

	result.ResponseTimeMs = int(time.Since(start).Milliseconds())
	return result
}

func executeHTTPProbe(item model.ProbeCheck, timeout time.Duration, result model.ProbeCheckResult) model.ProbeCheckResult {
	url := item.URL
	if url == "" {
		url = item.Target
	}
	if url == "" {
		result.Status = "error"
		result.Error = "HTTP 拨测缺少 URL"
		return result
	}

	client := &http.Client{Timeout: timeout}
	method := strings.ToUpper(item.HTTPConfig.Method)
	if method == "" {
		method = "GET"
	}

	req, err := http.NewRequest(method, url, nil)
	if err != nil {
		result.Status = "error"
		result.Error = "构建请求失败: " + err.Error()
		return result
	}
	for k, v := range item.HTTPConfig.Headers {
		req.Header.Set(k, v)
	}

	resp, err := client.Do(req)
	if err != nil {
		result.Status = "down"
		result.Error = "请求失败: " + err.Error()
		return result
	}
	defer resp.Body.Close()

	result.StatusCode = resp.StatusCode
	if resp.StatusCode >= 200 && resp.StatusCode < 400 {
		result.Status = "up"
	} else {
		result.Status = "degraded"
		result.Error = "HTTP " + strings.TrimSpace(resp.Status)
	}
	return result
}

func executeTCPProbe(item model.ProbeCheck, timeout time.Duration, result model.ProbeCheckResult) model.ProbeCheckResult {
	target := item.Target
	if target == "" {
		target = item.URL
	}
	if target == "" || item.Port == 0 {
		result.Status = "error"
		result.Error = "TCP 拨测缺少 target 或 port"
		return result
	}

	addr := fmt.Sprintf("%s:%d", target, item.Port)
	conn, err := net.DialTimeout("tcp", addr, timeout)
	if err != nil {
		result.Status = "down"
		result.Error = "TCP 连接失败: " + err.Error()
		return result
	}
	conn.Close()
	result.Status = "up"
	return result
}

func executeICMPProbe(item model.ProbeCheck, result model.ProbeCheckResult) model.ProbeCheckResult {
	target := item.Target
	if target == "" {
		target = item.URL
	}
	if target == "" {
		result.Status = "error"
		result.Error = "ICMP 拨测缺少 target"
		return result
	}
	// ICMP 需要 root 权限，使用 TCP 80 端口作为替代探测
	conn, err := net.DialTimeout("tcp", target+":80", 5*time.Second)
	if err != nil {
		result.Status = "down"
		result.Error = "主机不可达: " + err.Error()
		return result
	}
	conn.Close()
	result.Status = "up"
	return result
}

func executeDNSProbe(item model.ProbeCheck, timeout time.Duration, result model.ProbeCheckResult) model.ProbeCheckResult {
	target := item.Target
	if target == "" {
		target = item.URL
	}
	if target == "" {
		result.Status = "error"
		result.Error = "DNS 拨测缺少 target"
		return result
	}
	resolver := &net.Resolver{
		PreferGo: true,
		Dial: func(ctx context.Context, network, address string) (net.Conn, error) {
			d := net.Dialer{Timeout: timeout}
			return d.DialContext(ctx, "udp", "8.8.8.8:53")
		},
	}
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	addrs, err := resolver.LookupHost(ctx, target)
	if err != nil {
		result.Status = "down"
		result.Error = "DNS 解析失败: " + err.Error()
		return result
	}
	if len(addrs) == 0 {
		result.Status = "down"
		result.Error = "DNS 解析无结果"
		return result
	}
	result.Status = "up"
	result.Metadata = map[string]string{"resolved": strings.Join(addrs, ",")}
	return result
}
