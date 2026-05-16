package handler

import (
	"net/http"
	"sort"
	"sync"
	"time"

	"ai-workbench-api/internal/store"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

// cmdbApprovalRequest 审批工单
type cmdbApprovalRequest struct {
	ID           string         `json:"id"`
	Requester    string         `json:"requester"`
	ResourceType string         `json:"resource_type"`
	ResourceID   string         `json:"resource_id"`
	Action       string         `json:"action"`
	Title        string         `json:"title"`
	Description  string         `json:"description"`
	Status       string         `json:"status"`
	Approver     string         `json:"approver,omitempty"`
	ApprovedAt   string         `json:"approved_at,omitempty"`
	RejectReason string         `json:"reject_reason,omitempty"`
	Context      map[string]any `json:"context,omitempty"`
	CreatedAt    string         `json:"created_at"`
	UpdatedAt    string         `json:"updated_at"`
}

var (
	approvalMu       sync.RWMutex
	approvalRequests []cmdbApprovalRequest
)

// CmdbListApprovalsPending 列出待审批工单
func CmdbListApprovalsPending(c *gin.Context) {
	approvalMu.RLock()
	var out []cmdbApprovalRequest
	for _, r := range approvalRequests {
		if r.Status == "pending" {
			out = append(out, r)
		}
	}
	approvalMu.RUnlock()

	if out == nil {
		out = []cmdbApprovalRequest{}
	}
	sort.Slice(out, func(i, j int) bool { return out[i].CreatedAt > out[j].CreatedAt })
	c.JSON(http.StatusOK, gin.H{"items": out, "total": len(out)})
}

// CmdbListApprovalsMyRequests 列出我提交的审批请求
func CmdbListApprovalsMyRequests(c *gin.Context) {
	actor := requestActor(c)

	approvalMu.RLock()
	var out []cmdbApprovalRequest
	for _, r := range approvalRequests {
		if r.Requester == actor {
			out = append(out, r)
		}
	}
	approvalMu.RUnlock()

	if out == nil {
		out = []cmdbApprovalRequest{}
	}
	sort.Slice(out, func(i, j int) bool { return out[i].CreatedAt > out[j].CreatedAt })
	c.JSON(http.StatusOK, gin.H{"items": out, "total": len(out)})
}

// CmdbApproveApproval 审批通过
func CmdbApproveApproval(c *gin.Context) {
	id := c.Param("id")
	actor := requestActor(c)

	var req struct {
		Note string `json:"note"`
	}
	c.ShouldBindJSON(&req)

	approvalMu.Lock()
	var target *cmdbApprovalRequest
	for i := range approvalRequests {
		if approvalRequests[i].ID == id {
			target = &approvalRequests[i]
			break
		}
	}

	if target == nil {
		approvalMu.Unlock()
		c.JSON(http.StatusNotFound, gin.H{"error": "审批工单不存在"})
		return
	}

	if target.Status != "pending" {
		approvalMu.Unlock()
		c.JSON(http.StatusConflict, gin.H{"error": "该工单已处理", "status": target.Status})
		return
	}

	now := time.Now().Format(time.RFC3339)
	target.Status = "approved"
	target.Approver = actor
	target.ApprovedAt = now
	target.UpdatedAt = now
	result := *target
	approvalMu.Unlock()

	logrus.WithFields(logrus.Fields{
		"approval_id": id,
		"actor":       actor,
		"action":      "approve",
	}).Info("cmdb: approval approved")

	c.JSON(http.StatusOK, gin.H{"message": "审批已通过", "approval": result})
}

// CmdbRejectApproval 审批拒绝
func CmdbRejectApproval(c *gin.Context) {
	id := c.Param("id")
	actor := requestActor(c)

	var req struct {
		Reason string `json:"reason" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "参数无效: reason 必填"})
		return
	}

	approvalMu.Lock()
	var target *cmdbApprovalRequest
	for i := range approvalRequests {
		if approvalRequests[i].ID == id {
			target = &approvalRequests[i]
			break
		}
	}

	if target == nil {
		approvalMu.Unlock()
		c.JSON(http.StatusNotFound, gin.H{"error": "审批工单不存在"})
		return
	}

	if target.Status != "pending" {
		approvalMu.Unlock()
		c.JSON(http.StatusConflict, gin.H{"error": "该工单已处理", "status": target.Status})
		return
	}

	now := time.Now().Format(time.RFC3339)
	target.Status = "rejected"
	target.Approver = actor
	target.ApprovedAt = now
	target.RejectReason = req.Reason
	target.UpdatedAt = now
	result := *target
	approvalMu.Unlock()

	logrus.WithFields(logrus.Fields{
		"approval_id": id,
		"actor":       actor,
		"action":      "reject",
		"reason":      req.Reason,
	}).Info("cmdb: approval rejected")

	c.JSON(http.StatusOK, gin.H{"message": "审批已拒绝", "approval": result})
}

// CmdbListApprovalsArchived 列出已归档审批
func CmdbListApprovalsArchived(c *gin.Context) {
	approvalMu.RLock()
	var out []cmdbApprovalRequest
	for _, r := range approvalRequests {
		if r.Status == "approved" || r.Status == "rejected" {
			out = append(out, r)
		}
	}
	approvalMu.RUnlock()

	if out == nil {
		out = []cmdbApprovalRequest{}
	}
	sort.Slice(out, func(i, j int) bool { return out[i].UpdatedAt > out[j].UpdatedAt })
	c.JSON(http.StatusOK, gin.H{"items": out, "total": len(out)})
}

// CmdbCreateApprovalRequest 创建审批请求（内部使用，供批量操作等触发）
func CmdbCreateApprovalRequest(c *gin.Context) {
	var req struct {
		ResourceType string         `json:"resource_type" binding:"required"`
		ResourceID   string         `json:"resource_id" binding:"required"`
		Action       string         `json:"action" binding:"required"`
		Title        string         `json:"title" binding:"required"`
		Description  string         `json:"description"`
		Context      map[string]any `json:"context"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "参数无效: resource_type, resource_id, action, title 必填"})
		return
	}

	now := time.Now().Format(time.RFC3339)
	approval := cmdbApprovalRequest{
		ID:           "approval-" + store.NewID(),
		Requester:    requestActor(c),
		ResourceType: req.ResourceType,
		ResourceID:   req.ResourceID,
		Action:       req.Action,
		Title:        req.Title,
		Description:  req.Description,
		Status:       "pending",
		Context:      req.Context,
		CreatedAt:    now,
		UpdatedAt:    now,
	}

	approvalMu.Lock()
	approvalRequests = append(approvalRequests, approval)
	approvalMu.Unlock()

	logrus.WithFields(logrus.Fields{
		"approval_id":   approval.ID,
		"resource_type": req.ResourceType,
		"resource_id":   req.ResourceID,
		"action":        req.Action,
	}).Info("cmdb: approval request created")

	c.JSON(http.StatusCreated, approval)
}
