package store

import (
	"errors"
	"sort"
	"strings"
	"time"

	"ai-workbench-api/internal/model"

	"gorm.io/gorm"
)

var (
	cmdbResourceApprovals    = map[string]*model.CmdbResourceApproval{}
	cmdbOperationRiskRecords = map[string]*model.CmdbOperationRiskRecord{}
)

func CreateCmdbOperationRiskRecord(record *model.CmdbOperationRiskRecord) (*model.CmdbOperationRiskRecord, error) {
	normalized := normalizeCmdbOperationRiskRecord(record)
	if GormOK() {
		if err := GetDB().Create(&normalized).Error; err != nil {
			return nil, err
		}
		out := normalized
		return &out, nil
	}
	mu.Lock()
	cmdbOperationRiskRecords[normalized.ID] = copyCmdbOperationRiskRecord(normalized)
	mu.Unlock()
	out := normalized
	return &out, nil
}

func CreateCmdbResourceApproval(approval *model.CmdbResourceApproval) (*model.CmdbResourceApproval, error) {
	normalized := normalizeCmdbResourceApproval(approval)
	if GormOK() {
		if err := GetDB().Create(&normalized).Error; err != nil {
			return nil, err
		}
		out := normalized
		return &out, nil
	}
	mu.Lock()
	cmdbResourceApprovals[normalized.ID] = copyCmdbResourceApproval(normalized)
	mu.Unlock()
	out := normalized
	return &out, nil
}

func ListCmdbResourceApprovals(view, actor string) []model.CmdbResourceApproval {
	view = strings.ToLower(strings.TrimSpace(view))
	actor = strings.TrimSpace(actor)
	if GormOK() {
		var rows []model.CmdbResourceApproval
		query := GetDB().Model(&model.CmdbResourceApproval{})
		query = applyCmdbApprovalListScope(query, view, actor)
		if err := query.Order("updated_at desc").Find(&rows).Error; err != nil {
			return nil
		}
		return rows
	}
	mu.RLock()
	out := make([]model.CmdbResourceApproval, 0, len(cmdbResourceApprovals))
	for _, approval := range cmdbResourceApprovals {
		item := copyCmdbResourceApprovalValue(approval)
		if cmdbApprovalMatchesListScope(item, view, actor) {
			out = append(out, item)
		}
	}
	mu.RUnlock()
	sort.Slice(out, func(i, j int) bool { return out[i].UpdatedAt.After(out[j].UpdatedAt) })
	return out
}

func GetCmdbResourceApproval(id string) (*model.CmdbResourceApproval, bool) {
	id = strings.TrimSpace(id)
	if id == "" {
		return nil, false
	}
	if GormOK() {
		var row model.CmdbResourceApproval
		err := GetDB().Where("id = ?", id).First(&row).Error
		if err == nil {
			return &row, true
		}
		return nil, false
	}
	mu.RLock()
	defer mu.RUnlock()
	approval, ok := cmdbResourceApprovals[id]
	if !ok {
		return nil, false
	}
	out := copyCmdbResourceApprovalValue(approval)
	return &out, true
}

func DecideCmdbResourceApproval(id, decision, actor, note string) (*model.CmdbResourceApproval, bool, error) {
	id = strings.TrimSpace(id)
	if id == "" {
		return nil, false, nil
	}
	status, err := normalizeCmdbApprovalDecision(decision)
	if err != nil {
		return nil, false, err
	}
	actor = strings.TrimSpace(actor)
	note = strings.TrimSpace(note)
	now := time.Now()
	if GormOK() {
		var row model.CmdbResourceApproval
		err := GetDB().Where("id = ?", id).First(&row).Error
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, false, nil
		}
		if err != nil {
			return nil, false, err
		}
		if row.Status != "pending_review" {
			return nil, true, errors.New("cmdb approval request already reviewed")
		}
		row.Status = status
		row.WorkflowState = status
		row.DecisionActor = actor
		row.DecisionNote = note
		row.DecidedAt = &now
		row.UpdatedAt = now
		if err := GetDB().Save(&row).Error; err != nil {
			return nil, true, err
		}
		out := row
		return &out, true, nil
	}
	mu.Lock()
	defer mu.Unlock()
	approval, ok := cmdbResourceApprovals[id]
	if !ok {
		return nil, false, nil
	}
	if approval.Status != "pending_review" {
		return nil, true, errors.New("cmdb approval request already reviewed")
	}
	approval.Status = status
	approval.WorkflowState = status
	approval.DecisionActor = actor
	approval.DecisionNote = note
	approval.DecidedAt = &now
	approval.UpdatedAt = now
	out := copyCmdbResourceApprovalValue(approval)
	return &out, true, nil
}

func DeleteCmdbResourceApproval(id string) (bool, error) {
	id = strings.TrimSpace(id)
	if id == "" {
		return false, nil
	}
	found := false
	if GormOK() {
		res := GetDB().Where("id = ?", id).Delete(&model.CmdbResourceApproval{})
		if res.Error != nil {
			return false, res.Error
		}
		found = res.RowsAffected > 0
	}
	mu.Lock()
	if _, ok := cmdbResourceApprovals[id]; ok {
		delete(cmdbResourceApprovals, id)
		found = true
	}
	mu.Unlock()
	return found, nil
}

func DeleteCmdbOperationRiskRecord(id string) (bool, error) {
	id = strings.TrimSpace(id)
	if id == "" {
		return false, nil
	}
	found := false
	if GormOK() {
		res := GetDB().Where("id = ?", id).Delete(&model.CmdbOperationRiskRecord{})
		if res.Error != nil {
			return false, res.Error
		}
		found = res.RowsAffected > 0
	}
	mu.Lock()
	if _, ok := cmdbOperationRiskRecords[id]; ok {
		delete(cmdbOperationRiskRecords, id)
		found = true
	}
	mu.Unlock()
	return found, nil
}

func ResetCmdbApprovalRiskForTest() {
	mu.Lock()
	defer mu.Unlock()
	cmdbResourceApprovals = map[string]*model.CmdbResourceApproval{}
	cmdbOperationRiskRecords = map[string]*model.CmdbOperationRiskRecord{}
}

func normalizeCmdbResourceApproval(approval *model.CmdbResourceApproval) model.CmdbResourceApproval {
	out := model.CmdbResourceApproval{}
	if approval != nil {
		out = *approval
	}
	now := time.Now()
	out.ID = strings.TrimSpace(out.ID)
	if out.ID == "" {
		out.ID = "cmdb-approval-" + NewID()
	}
	out.View = strings.TrimSpace(out.View)
	out.Requester = strings.TrimSpace(out.Requester)
	out.Approver = strings.TrimSpace(out.Approver)
	out.ResourceType = strings.TrimSpace(out.ResourceType)
	out.ResourceID = strings.TrimSpace(out.ResourceID)
	out.Action = strings.TrimSpace(out.Action)
	out.RiskLevel = strings.TrimSpace(out.RiskLevel)
	out.Status = strings.TrimSpace(out.Status)
	if out.Status == "" {
		out.Status = "pending_review"
	}
	if out.View == "" && out.Status == "pending_review" {
		out.View = "todo"
	}
	out.Title = strings.TrimSpace(out.Title)
	out.Summary = strings.TrimSpace(out.Summary)
	out.BusinessGroup = strings.TrimSpace(out.BusinessGroup)
	out.WorkflowState = strings.TrimSpace(out.WorkflowState)
	if out.WorkflowState == "" {
		out.WorkflowState = out.Status
	}
	out.RiskRecordID = strings.TrimSpace(out.RiskRecordID)
	out.ContextJSON = strings.TrimSpace(out.ContextJSON)
	out.DiffJSON = strings.TrimSpace(out.DiffJSON)
	out.AuditRef = strings.TrimSpace(out.AuditRef)
	out.DecisionActor = strings.TrimSpace(out.DecisionActor)
	out.DecisionNote = strings.TrimSpace(out.DecisionNote)
	if out.CreatedAt.IsZero() {
		out.CreatedAt = now
	}
	out.UpdatedAt = now
	return out
}

func normalizeCmdbOperationRiskRecord(record *model.CmdbOperationRiskRecord) model.CmdbOperationRiskRecord {
	out := model.CmdbOperationRiskRecord{}
	if record != nil {
		out = *record
	}
	now := time.Now()
	out.ID = strings.TrimSpace(out.ID)
	if out.ID == "" {
		out.ID = "cmdb-risk-" + NewID()
	}
	out.ResourceType = strings.TrimSpace(out.ResourceType)
	out.ResourceID = strings.TrimSpace(out.ResourceID)
	out.Action = strings.TrimSpace(out.Action)
	out.RiskLevel = strings.TrimSpace(out.RiskLevel)
	out.PolicyID = strings.TrimSpace(out.PolicyID)
	out.Status = strings.TrimSpace(out.Status)
	if out.Status == "" {
		out.Status = "risk_recorded"
	}
	out.Actor = strings.TrimSpace(out.Actor)
	out.BusinessGroup = strings.TrimSpace(out.BusinessGroup)
	out.Reason = strings.TrimSpace(out.Reason)
	out.ContextJSON = strings.TrimSpace(out.ContextJSON)
	out.AuditRef = strings.TrimSpace(out.AuditRef)
	if out.CreatedAt.IsZero() {
		out.CreatedAt = now
	}
	out.UpdatedAt = now
	return out
}

func normalizeCmdbApprovalDecision(decision string) (string, error) {
	switch strings.ToLower(strings.TrimSpace(decision)) {
	case "accept", "approve":
		return "review_accept_recorded", nil
	case "reject", "rejected":
		return "review_reject_recorded", nil
	case "deny", "denied":
		return "review_deny_recorded", nil
	case "cancel", "canceled", "cancelled":
		return "review_cancel_recorded", nil
	default:
		return "", errors.New("invalid cmdb approval decision")
	}
}

func applyCmdbApprovalListScope(query *gorm.DB, view, actor string) *gorm.DB {
	switch view {
	case "mine":
		if actor != "" {
			return query.Where("requester = ?", actor)
		}
	case "todo":
		query = query.Where("status = ?", "pending_review")
		if actor != "" {
			return query.Where("approver = ?", actor)
		}
		return query
	case "archive":
		query = query.Where("status <> ?", "pending_review")
		if actor != "" {
			return query.Where("requester = ? OR approver = ?", actor, actor)
		}
		return query
	default:
		if actor != "" {
			return query.Where("requester = ? OR approver = ?", actor, actor)
		}
	}
	return query
}

func cmdbApprovalMatchesListScope(item model.CmdbResourceApproval, view, actor string) bool {
	switch view {
	case "mine":
		return actor == "" || item.Requester == actor
	case "todo":
		return item.Status == "pending_review" && (actor == "" || item.Approver == actor)
	case "archive":
		return item.Status != "pending_review" && (actor == "" || item.Requester == actor || item.Approver == actor)
	default:
		return actor == "" || item.Requester == actor || item.Approver == actor
	}
}

func copyCmdbResourceApproval(in model.CmdbResourceApproval) *model.CmdbResourceApproval {
	out := in
	return &out
}

func copyCmdbResourceApprovalValue(in *model.CmdbResourceApproval) model.CmdbResourceApproval {
	if in == nil {
		return model.CmdbResourceApproval{}
	}
	return *copyCmdbResourceApproval(*in)
}

func copyCmdbOperationRiskRecord(in model.CmdbOperationRiskRecord) *model.CmdbOperationRiskRecord {
	out := in
	return &out
}
