package store

import (
	"sort"
	"time"

	"ai-workbench-api/internal/model"

	"github.com/sirupsen/logrus"
)

var cmdbRelationActionRequests []model.CmdbRelationActionRequest
var cmdbRelationActionReceipts []model.CmdbRelationActionReceipt

func SaveCmdbRelationActionRequest(action *model.CmdbRelationActionRequest) (*model.CmdbRelationActionRequest, error) {
	now := time.Now()
	if action.ID == "" {
		action.ID = "cmdb-relation-action-" + NewID()
	}
	if action.CreatedAt.IsZero() {
		action.CreatedAt = now
	}
	action.UpdatedAt = now
	if GormOK() {
		if err := GetDB().Save(action).Error; err != nil {
			return nil, err
		}
		out := *action
		return &out, nil
	}
	mu.Lock()
	defer mu.Unlock()
	for i := range cmdbRelationActionRequests {
		if cmdbRelationActionRequests[i].ID == action.ID {
			cmdbRelationActionRequests[i] = *action
			out := cmdbRelationActionRequests[i]
			return &out, nil
		}
	}
	cmdbRelationActionRequests = append(cmdbRelationActionRequests, *action)
	out := *action
	return &out, nil
}

func GetCmdbRelationActionRequest(id string) (*model.CmdbRelationActionRequest, bool) {
	if GormOK() {
		var row model.CmdbRelationActionRequest
		if err := GetDB().Where("id = ?", id).First(&row).Error; err != nil {
			return nil, false
		}
		return &row, true
	}
	mu.RLock()
	defer mu.RUnlock()
	for i := range cmdbRelationActionRequests {
		if cmdbRelationActionRequests[i].ID == id {
			out := cmdbRelationActionRequests[i]
			return &out, true
		}
	}
	return nil, false
}

func ListCmdbRelationActionRequests(instanceID string) []model.CmdbRelationActionRequest {
	if GormOK() {
		var rows []model.CmdbRelationActionRequest
		err := GetDB().
			Where("instance_id = ?", instanceID).
			Order("updated_at desc").
			Find(&rows).Error
		if err != nil {
			logrus.WithError(err).Warn("cmdb: list relation action requests failed")
			return nil
		}
		return rows
	}
	mu.RLock()
	defer mu.RUnlock()
	out := make([]model.CmdbRelationActionRequest, 0)
	for _, item := range cmdbRelationActionRequests {
		if item.InstanceID == instanceID {
			out = append(out, item)
		}
	}
	sort.Slice(out, func(i, j int) bool { return out[i].UpdatedAt.After(out[j].UpdatedAt) })
	return out
}

func SaveCmdbRelationActionReceipt(receipt *model.CmdbRelationActionReceipt) (*model.CmdbRelationActionReceipt, error) {
	if receipt.ID == "" {
		receipt.ID = "cmdb-relation-action-receipt-" + NewID()
	}
	if receipt.CreatedAt.IsZero() {
		receipt.CreatedAt = time.Now()
	}
	if GormOK() {
		if err := GetDB().Create(receipt).Error; err != nil {
			return nil, err
		}
		out := *receipt
		return &out, nil
	}
	mu.Lock()
	defer mu.Unlock()
	cmdbRelationActionReceipts = append(cmdbRelationActionReceipts, *receipt)
	out := *receipt
	return &out, nil
}

func ListCmdbRelationActionReceipts(actionRequestID string) []model.CmdbRelationActionReceipt {
	if GormOK() {
		var rows []model.CmdbRelationActionReceipt
		err := GetDB().
			Where("action_request_id = ?", actionRequestID).
			Order("created_at asc").
			Find(&rows).Error
		if err != nil {
			logrus.WithError(err).Warn("cmdb: list relation action receipts failed")
			return nil
		}
		return rows
	}
	mu.RLock()
	defer mu.RUnlock()
	out := make([]model.CmdbRelationActionReceipt, 0)
	for _, receipt := range cmdbRelationActionReceipts {
		if receipt.ActionRequestID == actionRequestID {
			out = append(out, receipt)
		}
	}
	sort.Slice(out, func(i, j int) bool { return out[i].CreatedAt.Before(out[j].CreatedAt) })
	return out
}
