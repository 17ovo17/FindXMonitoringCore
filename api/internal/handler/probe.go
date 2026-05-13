package handler

import (
	"net/http"
	"strings"
	"time"

	"ai-workbench-api/internal/model"
	"ai-workbench-api/internal/store"

	"github.com/gin-gonic/gin"
)

const probeBlockedMessage = "业务拨测缺少执行器、调度器、通知回执或告警闭环契约，已按安全策略阻断"

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
	if _, ok, err := store.GetProbeCheck(c.Param("id")); err != nil || !ok {
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "业务拨测检查项查询失败"})
			return
		}
		c.JSON(http.StatusNotFound, gin.H{"error": "业务拨测检查项不存在"})
		return
	}
	blockProbeContract(c, "FX-CONTRACT-BUSINESS-PROBE-EXECUTOR", []string{
		"probe.executor.http_tcp_ping_dns",
		"probe.execution_receipt",
		"probe.result_evidence",
	}, "当前未接入真实 HTTP/TCP/PING/DNS 拨测执行器，不能把命令预览或任务创建当作拨测成功。")
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
		"capability": probeBlockedCapability("FX-CONTRACT-BUSINESS-PROBE-NOTIFICATION-RECEIPT", "probe.notification.receipt", []string{"notification.delivery_receipt", "probe.incident_notification_audit"}),
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
		"capability": probeBlockedCapability("FX-CONTRACT-BUSINESS-PROBE-NOTIFICATION-RECEIPT", "probe.notification.receipt", []string{"notification.delivery_receipt", "probe.incident_notification_audit"}),
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

func blockProbeContract(c *gin.Context, contractGapID string, missing []string, reason string) {
	c.JSON(http.StatusConflict, model.ProbeBlockedResponse{
		Code:             model.ProbeContractBlockedCode,
		Message:          probeBlockedMessage,
		ContractGapID:    contractGapID,
		Status:           "blocked_by_contract",
		SafeToRetry:      false,
		MissingContracts: missing,
		Capability:       probeBlockedCapability(contractGapID, "probe.execution", missing),
		ContractMatrix:   probeCapabilityMatrix(),
		Meta: map[string]any{
			"reason":          reason,
			"honesty_policy":  "blocked is not success",
			"evidence_policy": "真实拨测结果必须包含执行器回执、耗时、状态和 evidence ref。",
		},
	})
}

func probeCapabilityMatrix() []model.ProbeCapability {
	return []model.ProbeCapability{
		{ID: "FX-CONTRACT-BUSINESS-PROBE-CONFIG", Capability: "probe.config.crud", Domain: "business_probe", Status: "ready", Message: "检查项、状态页和人工事故配置可保存。", SafeToRetry: false},
		probeBlockedCapability("FX-CONTRACT-BUSINESS-PROBE-EXECUTOR", "probe.executor.http_tcp_ping_dns", []string{"probe.executor", "probe.execution_receipt", "probe.result_evidence"}),
		probeBlockedCapability("FX-CONTRACT-BUSINESS-PROBE-SCHEDULER", "probe.scheduler", []string{"probe.scheduler", "probe.dedup_lock", "probe.run_history"}),
		probeBlockedCapability("FX-CONTRACT-BUSINESS-PROBE-NOTIFICATION-RECEIPT", "probe.notification.receipt", []string{"notification.delivery_receipt", "probe.subscription_audit"}),
		probeBlockedCapability("FX-CONTRACT-BUSINESS-PROBE-ALERT-LIFECYCLE", "probe.alert.lifecycle", []string{"probe.alert_auto_create", "probe.alert_recovery", "probe.alert_dedup"}),
		probeBlockedCapability("FX-CONTRACT-BUSINESS-PROBE-EVIDENCE-CHAIN", "probe.evidence_chain", []string{"probe.run_evidence_chain", "probe.incident_evidence_chain"}),
	}
}

func probeBlockedCapability(id, capability string, missing []string) model.ProbeCapability {
	return model.ProbeCapability{
		ID:               id,
		Capability:       capability,
		Domain:           "business_probe",
		Status:           "blocked_by_contract",
		ContractGapID:    id,
		MissingContracts: missing,
		Message:          probeBlockedMessage,
		SafeToRetry:      false,
	}
}

func auditProbe(c *gin.Context, action, target, decision string) {
	store.AddAuditEvent(store.AuditEvent{
		ID:        "audit-" + store.NewID(),
		Action:    action,
		Target:    target,
		Risk:      "medium",
		Decision:  decision,
		Detail:    "业务拨测配置变更；真实执行、通知投递和告警闭环仍以契约矩阵为准。",
		Operator:  requestActor(c),
		ClientIP:  c.ClientIP(),
		CreatedAt: time.Now(),
	})
}
