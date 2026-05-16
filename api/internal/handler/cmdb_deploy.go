package handler

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"net/http"
	"sort"
	"strings"
	"time"

	"ai-workbench-api/internal/model"
	"ai-workbench-api/internal/store"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

type deployTask struct {
	ID               string         `json:"id"`
	Name             string         `json:"name"`
	TargetHosts      []string       `json:"target_hosts"`
	ScriptLength     int            `json:"script_length"`
	ScriptDigest     string         `json:"script_digest"`
	Status           string         `json:"status"`
	Progress         int            `json:"progress"`
	Creator          string         `json:"creator"`
	CreatedAt        string         `json:"created_at"`
	UpdatedAt        string         `json:"updated_at"`
	Logs             []string       `json:"logs,omitempty"`
	Code             string         `json:"code"`
	ContractID       string         `json:"contract_id"`
	MissingContracts []string       `json:"missing_contracts"`
	SafeToRetry      bool           `json:"safe_to_retry"`
	AuditRef         string         `json:"audit_ref"`
	LogRef           string         `json:"log_ref"`
	Meta             map[string]any `json:"meta,omitempty"`
}

// CmdbCreateDeployTask 创建部署任务阻断记录。脚本只用于 preflight，不持久化和回显原文。
func CmdbCreateDeployTask(c *gin.Context) {
	var req struct {
		Name        string   `json:"name" binding:"required"`
		TargetHosts []string `json:"target_hosts" binding:"required"`
		Script      string   `json:"script" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "请求参数无效: name/target_hosts/script 必填"})
		return
	}

	if len(req.TargetHosts) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "目标主机列表不能为空"})
		return
	}
	if len(req.Script) > 10000 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "脚本内容超过 10000 字符限制"})
		return
	}

	now := time.Now()
	contractID := "cmdb.deploy.executor.v1"
	missing := cmdbHighRiskMissingContracts("remote_executor_contract", "credential_ref_resolver", "artifact_transfer_contract", "deploy_audit_contract")
	task := deployTask{
		Name:             strings.TrimSpace(req.Name),
		TargetHosts:      append([]string(nil), req.TargetHosts...),
		ScriptLength:     len(req.Script),
		ScriptDigest:     deployScriptDigest(req.Script),
		Status:           strings.ToLower(cmdbBlockedByContract),
		Progress:         0,
		Creator:          requestActor(c),
		CreatedAt:        now.Format(time.RFC3339),
		UpdatedAt:        now.Format(time.RFC3339),
		Logs:             []string{"preflight blocked: remote deploy executor contract is not connected"},
		Code:             cmdbBlockedByContract,
		ContractID:       contractID,
		MissingContracts: missing,
		SafeToRetry:      false,
		AuditRef:         "audit://cmdb/deploy/" + now.Format(time.RFC3339),
		LogRef:           "log://cmdb/deploy/" + now.Format(time.RFC3339),
		Meta: map[string]any{
			"preflight":             "failed",
			"reason":                "缺少真实远程执行器、凭据解析、文件传输和审计契约，不能进入部署执行闭环",
			"persistence":           cmdbPersistenceStatus(),
			"script_storage_policy": "raw_script_not_persisted",
		},
	}

	record := deployTaskToModel(task)
	if err := store.CreateCmdbDeployTask(&record); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "保存部署阻断任务失败"})
		return
	}
	task = deployTaskFromModel(record)

	logrus.WithFields(logrus.Fields{
		"task_id": task.ID,
		"name":    task.Name,
		"hosts":   len(task.TargetHosts),
		"action":  "deploy_task_blocked",
	}).Warn("cmdb: deploy task blocked by missing executor contract")

	gate := cmdbHighRiskApprovalGate(c, cmdbHighRiskApprovalInput{
		ContractID:   contractID,
		ResourceType: "cmdb_deploy_task",
		ResourceID:   task.ID,
		Action:       "deploy",
		RiskLevel:    "critical",
		Title:        "CMDB deploy task review",
		Summary:      "Deploy task requires approval, transfer and execution receipts before execution.",
		Reason:       "missing remote executor, credential resolver, artifact transfer and deploy audit contracts",
		Context: map[string]any{
			"task_id":       task.ID,
			"name":          task.Name,
			"target_count":  len(task.TargetHosts),
			"script_length": task.ScriptLength,
			"script_digest": task.ScriptDigest,
		},
		Missing:        []string{"remote_executor_contract", "credential_ref_resolver", "artifact_transfer_contract", "deploy_audit_contract"},
		ExecutionState: "deploy executor remains blocked",
	})
	c.JSON(http.StatusConflict, gin.H{
		"task": task,
		"gate": gate,
	})
}

// CmdbListDeployTasks 从 store 层读取部署任务列表。
func CmdbListDeployTasks(c *gin.Context) {
	page, limit := cmdbPageAndLimit(c)
	rows, total := store.ListCmdbDeployTasks(page, limit)

	out := make([]deployTask, 0, len(rows))
	for _, row := range rows {
		out = append(out, deployTaskFromModel(row))
	}
	for i := range out {
		out[i].Logs = nil
	}
	sort.Slice(out, func(i, j int) bool { return out[i].CreatedAt > out[j].CreatedAt })

	c.JSON(http.StatusOK, gin.H{
		"items": out,
		"total": total,
		"page":  page,
		"limit": limit,
		"meta":  cmdbDeployPersistenceMeta(),
	})
}

// CmdbGetDeployTask 从 store 层读取部署任务详情。
func CmdbGetDeployTask(c *gin.Context) {
	task, ok := store.GetCmdbDeployTask(c.Param("id"))
	if !ok {
		c.JSON(http.StatusNotFound, gin.H{"error": "部署任务不存在"})
		return
	}
	c.JSON(http.StatusOK, deployTaskFromModel(*task))
}

func deployTaskToModel(task deployTask) model.CmdbDeployTask {
	return model.CmdbDeployTask{
		ID:              task.ID,
		Name:            task.Name,
		TargetHostsJSON: mustMarshalJSON(task.TargetHosts),
		ScriptLength:    task.ScriptLength,
		ScriptDigest:    task.ScriptDigest,
		Status:          task.Status,
		Progress:        task.Progress,
		Creator:         task.Creator,
		LogsJSON:        mustMarshalJSON(task.Logs),
		Code:            task.Code,
		ContractID:      task.ContractID,
		MissingJSON:     mustMarshalJSON(task.MissingContracts),
		SafeToRetry:     task.SafeToRetry,
		AuditRef:        task.AuditRef,
		LogRef:          task.LogRef,
		MetaJSON:        mustMarshalJSON(task.Meta),
	}
}

func deployTaskFromModel(task model.CmdbDeployTask) deployTask {
	return deployTask{
		ID:               task.ID,
		Name:             task.Name,
		TargetHosts:      unmarshalStringSlice(task.TargetHostsJSON),
		ScriptLength:     task.ScriptLength,
		ScriptDigest:     task.ScriptDigest,
		Status:           task.Status,
		Progress:         task.Progress,
		Creator:          task.Creator,
		CreatedAt:        task.CreatedAt.Format(time.RFC3339),
		UpdatedAt:        task.UpdatedAt.Format(time.RFC3339),
		Logs:             unmarshalStringSlice(task.LogsJSON),
		Code:             task.Code,
		ContractID:       task.ContractID,
		MissingContracts: unmarshalStringSlice(task.MissingJSON),
		SafeToRetry:      task.SafeToRetry,
		AuditRef:         task.AuditRef,
		LogRef:           task.LogRef,
		Meta:             unmarshalStringAnyMap(task.MetaJSON),
	}
}

func deployScriptDigest(script string) string {
	sum := sha256.Sum256([]byte(script))
	return hex.EncodeToString(sum[:])
}

func cmdbDeployPersistenceMeta() gin.H {
	meta := gin.H{"persistence": cmdbPersistenceStatus()}
	if !store.GormOK() {
		meta["risk"] = "memory_fallback/dev-only, not production persistence"
	}
	return meta
}

func mustMarshalJSON(value any) string {
	data, err := json.Marshal(value)
	if err != nil {
		return "{}"
	}
	return string(data)
}

func unmarshalStringSlice(data string) []string {
	var values []string
	if err := json.Unmarshal([]byte(data), &values); err != nil {
		return nil
	}
	return values
}

func unmarshalStringAnyMap(data string) map[string]any {
	values := make(map[string]any)
	if err := json.Unmarshal([]byte(data), &values); err != nil {
		return values
	}
	return values
}
