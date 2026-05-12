package store

import (
	"fmt"
	"strings"

	"ai-workbench-api/internal/model"

	"gorm.io/gorm"
)

// CreateAlertEvent 创建告警事件（GORM 模型）。
func CreateAlertEvent(event *model.MonitorAlertEventRecord) error {
	if !GormOK() {
		return fmt.Errorf("gorm db not available")
	}
	return GetDB().Create(event).Error
}

// ListAlertEventsPaged 分页查询告警事件（GORM 模型），支持状态过滤和搜索。
func ListAlertEventsPaged(status string, page, pageSize int, search string) ([]model.MonitorAlertEventRecord, int64, error) {
	if !GormOK() {
		return nil, 0, fmt.Errorf("gorm db not available")
	}
	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 || pageSize > 100 {
		pageSize = 20
	}

	query := GetDB().Model(&model.MonitorAlertEventRecord{})
	query = applyAlertEventFilters(query, status, search)

	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("count alert events: %w", err)
	}

	var events []model.MonitorAlertEventRecord
	offset := (page - 1) * pageSize
	err := query.Order("created_at DESC").Offset(offset).Limit(pageSize).Find(&events).Error
	if err != nil {
		return nil, 0, fmt.Errorf("query alert events: %w", err)
	}
	return events, total, nil
}

// GetAlertEvent 按 ID 查询单个告警事件。
func GetAlertEvent(id uint) (*model.MonitorAlertEventRecord, error) {
	if !GormOK() {
		return nil, fmt.Errorf("gorm db not available")
	}
	var event model.MonitorAlertEventRecord
	if err := GetDB().First(&event, id).Error; err != nil {
		return nil, err
	}
	return &event, nil
}

// UpdateAlertEventStatus 更新告警事件状态及附加字段。
func UpdateAlertEventStatus(id uint, status string, fields map[string]interface{}) error {
	if !GormOK() {
		return fmt.Errorf("gorm db not available")
	}
	if fields == nil {
		fields = map[string]interface{}{}
	}
	fields["status"] = status
	return GetDB().Model(&model.MonitorAlertEventRecord{}).Where("id = ?", id).Updates(fields).Error
}

func applyAlertEventFilters(query *gorm.DB, status, search string) *gorm.DB {
	if status != "" {
		query = query.Where("status = ?", status)
	}
	if search != "" {
		like := "%" + strings.TrimSpace(search) + "%"
		query = query.Where("rule_name LIKE ?", like)
	}
	return query
}
