package store

import (
	"sort"
	"time"

	"ai-workbench-api/internal/model"

	"github.com/sirupsen/logrus"
)

var cmdbMonitorBindings []model.CmdbMonitorBinding
var cmdbMonitorBindingReceipts []model.CmdbMonitorBindingReceipt

func SaveCmdbMonitorBinding(binding *model.CmdbMonitorBinding) (*model.CmdbMonitorBinding, error) {
	now := time.Now()
	if binding.ID == "" {
		binding.ID = "cmdb-monitor-binding-" + NewID()
	}
	if binding.CreatedAt.IsZero() {
		binding.CreatedAt = now
	}
	binding.UpdatedAt = now
	if GormOK() {
		if err := GetDB().Save(binding).Error; err != nil {
			return nil, err
		}
		out := *binding
		return &out, nil
	}
	mu.Lock()
	defer mu.Unlock()
	for i := range cmdbMonitorBindings {
		if cmdbMonitorBindings[i].ID == binding.ID {
			cmdbMonitorBindings[i] = *binding
			out := cmdbMonitorBindings[i]
			return &out, nil
		}
	}
	cmdbMonitorBindings = append(cmdbMonitorBindings, *binding)
	out := *binding
	return &out, nil
}

func ListCmdbMonitorBindings(instanceID string) []model.CmdbMonitorBinding {
	if GormOK() {
		var rows []model.CmdbMonitorBinding
		err := GetDB().
			Where("instance_id = ?", instanceID).
			Order("updated_at desc").
			Find(&rows).Error
		if err != nil {
			logrus.WithError(err).Warn("cmdb: list monitor bindings failed")
			return nil
		}
		return rows
	}
	mu.RLock()
	defer mu.RUnlock()
	out := make([]model.CmdbMonitorBinding, 0)
	for _, binding := range cmdbMonitorBindings {
		if binding.InstanceID == instanceID {
			out = append(out, binding)
		}
	}
	sort.Slice(out, func(i, j int) bool { return out[i].UpdatedAt.After(out[j].UpdatedAt) })
	return out
}

func GetCmdbMonitorBinding(id string) (*model.CmdbMonitorBinding, bool) {
	if GormOK() {
		var row model.CmdbMonitorBinding
		if err := GetDB().Where("id = ?", id).First(&row).Error; err != nil {
			return nil, false
		}
		return &row, true
	}
	mu.RLock()
	defer mu.RUnlock()
	for i := range cmdbMonitorBindings {
		if cmdbMonitorBindings[i].ID == id {
			out := cmdbMonitorBindings[i]
			return &out, true
		}
	}
	return nil, false
}

func SaveCmdbMonitorBindingReceipt(receipt *model.CmdbMonitorBindingReceipt) (*model.CmdbMonitorBindingReceipt, error) {
	if receipt.ID == "" {
		receipt.ID = "cmdb-monitor-binding-receipt-" + NewID()
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
	cmdbMonitorBindingReceipts = append(cmdbMonitorBindingReceipts, *receipt)
	out := *receipt
	return &out, nil
}

func GetCmdbMonitorBindingReceipt(bindingID, receiptType string) (*model.CmdbMonitorBindingReceipt, bool) {
	if GormOK() {
		var row model.CmdbMonitorBindingReceipt
		if err := GetDB().
			Where("binding_id = ? AND receipt_type = ?", bindingID, receiptType).
			Order("created_at asc").
			First(&row).Error; err != nil {
			return nil, false
		}
		return &row, true
	}
	mu.RLock()
	defer mu.RUnlock()
	for i := range cmdbMonitorBindingReceipts {
		if cmdbMonitorBindingReceipts[i].BindingID == bindingID && cmdbMonitorBindingReceipts[i].ReceiptType == receiptType {
			out := cmdbMonitorBindingReceipts[i]
			return &out, true
		}
	}
	return nil, false
}

func UpdateCmdbMonitorBindingReceipt(receipt *model.CmdbMonitorBindingReceipt) (*model.CmdbMonitorBindingReceipt, error) {
	if receipt == nil || receipt.ID == "" {
		return nil, nil
	}
	if receipt.CreatedAt.IsZero() {
		receipt.CreatedAt = time.Now()
	}
	if GormOK() {
		if err := GetDB().Save(receipt).Error; err != nil {
			return nil, err
		}
		out := *receipt
		return &out, nil
	}
	mu.Lock()
	defer mu.Unlock()
	for i := range cmdbMonitorBindingReceipts {
		if cmdbMonitorBindingReceipts[i].ID == receipt.ID {
			cmdbMonitorBindingReceipts[i] = *receipt
			out := cmdbMonitorBindingReceipts[i]
			return &out, nil
		}
	}
	cmdbMonitorBindingReceipts = append(cmdbMonitorBindingReceipts, *receipt)
	out := *receipt
	return &out, nil
}

func ListCmdbMonitorBindingReceipts(bindingID string) []model.CmdbMonitorBindingReceipt {
	if GormOK() {
		var rows []model.CmdbMonitorBindingReceipt
		err := GetDB().
			Where("binding_id = ?", bindingID).
			Order("created_at asc").
			Find(&rows).Error
		if err != nil {
			logrus.WithError(err).Warn("cmdb: list monitor binding receipts failed")
			return nil
		}
		return rows
	}
	mu.RLock()
	defer mu.RUnlock()
	out := make([]model.CmdbMonitorBindingReceipt, 0)
	for _, receipt := range cmdbMonitorBindingReceipts {
		if receipt.BindingID == bindingID {
			out = append(out, receipt)
		}
	}
	sort.Slice(out, func(i, j int) bool { return out[i].CreatedAt.Before(out[j].CreatedAt) })
	return out
}

func ListCmdbMonitorBindingReceiptsForInstance(instanceID string) []model.CmdbMonitorBindingReceipt {
	if GormOK() {
		var rows []model.CmdbMonitorBindingReceipt
		err := GetDB().
			Where("instance_id = ?", instanceID).
			Order("created_at asc").
			Find(&rows).Error
		if err != nil {
			logrus.WithError(err).Warn("cmdb: list monitor binding receipts by instance failed")
			return nil
		}
		return rows
	}
	mu.RLock()
	defer mu.RUnlock()
	out := make([]model.CmdbMonitorBindingReceipt, 0)
	for _, receipt := range cmdbMonitorBindingReceipts {
		if receipt.InstanceID == instanceID {
			out = append(out, receipt)
		}
	}
	sort.Slice(out, func(i, j int) bool { return out[i].CreatedAt.Before(out[j].CreatedAt) })
	return out
}
