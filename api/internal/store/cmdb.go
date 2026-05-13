package store

import (
	"sort"
	"sync"
	"time"

	"ai-workbench-api/internal/model"

	"github.com/sirupsen/logrus"
)

// 内存回退（GormOK() == false 时使用）
var (
	cmdbCategories        []model.CmdbCategory
	cmdbObjects           []model.CmdbObject
	cmdbAttributes        []model.CmdbAttribute
	cmdbInstances         []model.CmdbInstance
	cmdbDeployTasks       []model.CmdbDeployTask
	cmdbDeployMigrateOnce sync.Once
	cmdbDeployMigrateErr  error
)

func ListCmdbCategories() []model.CmdbCategory {
	if GormOK() {
		var rows []model.CmdbCategory
		if err := GetDB().Order("sort asc").Find(&rows).Error; err != nil {
			logrus.WithError(err).Warn("cmdb: list categories failed")
			return nil
		}
		return rows
	}
	mu.RLock()
	defer mu.RUnlock()
	out := make([]model.CmdbCategory, len(cmdbCategories))
	copy(out, cmdbCategories)
	sort.Slice(out, func(i, j int) bool { return out[i].Sort < out[j].Sort })
	return out
}

func CreateCmdbCategory(cat *model.CmdbCategory) error {
	cat.ID = NewID()
	if GormOK() {
		return GetDB().Create(cat).Error
	}
	mu.Lock()
	cmdbCategories = append(cmdbCategories, *cat)
	mu.Unlock()
	return nil
}

func ListCmdbObjects(categoryID string) []model.CmdbObject {
	if GormOK() {
		var rows []model.CmdbObject
		q := GetDB().Order("updated_at desc")
		if categoryID != "" {
			q = q.Where("category_id = ?", categoryID)
		}
		if err := q.Find(&rows).Error; err != nil {
			logrus.WithError(err).Warn("cmdb: list objects failed")
			return nil
		}
		return rows
	}
	mu.RLock()
	defer mu.RUnlock()
	var out []model.CmdbObject
	for _, o := range cmdbObjects {
		if categoryID == "" || o.CategoryID == categoryID {
			out = append(out, o)
		}
	}
	sort.Slice(out, func(i, j int) bool { return out[i].UpdatedAt.After(out[j].UpdatedAt) })
	return out
}

func GetCmdbObject(id string) (*model.CmdbObject, bool) {
	if GormOK() {
		var obj model.CmdbObject
		if err := GetDB().Where("id = ?", id).First(&obj).Error; err != nil {
			return nil, false
		}
		return &obj, true
	}
	mu.RLock()
	defer mu.RUnlock()
	for i := range cmdbObjects {
		if cmdbObjects[i].ID == id {
			cp := cmdbObjects[i]
			return &cp, true
		}
	}
	return nil, false
}

func CreateCmdbObject(obj *model.CmdbObject) error {
	obj.ID = NewID()
	now := time.Now()
	obj.CreatedAt = now
	obj.UpdatedAt = now
	if GormOK() {
		return GetDB().Create(obj).Error
	}
	mu.Lock()
	cmdbObjects = append(cmdbObjects, *obj)
	mu.Unlock()
	return nil
}

func UpdateCmdbObject(obj *model.CmdbObject) error {
	obj.UpdatedAt = time.Now()
	if GormOK() {
		return GetDB().Save(obj).Error
	}
	mu.Lock()
	for i := range cmdbObjects {
		if cmdbObjects[i].ID == obj.ID {
			cmdbObjects[i] = *obj
			break
		}
	}
	mu.Unlock()
	return nil
}

func DeleteCmdbObject(id string) error {
	if GormOK() {
		return GetDB().Where("id = ?", id).Delete(&model.CmdbObject{}).Error
	}
	mu.Lock()
	for i := range cmdbObjects {
		if cmdbObjects[i].ID == id {
			cmdbObjects = append(cmdbObjects[:i], cmdbObjects[i+1:]...)
			break
		}
	}
	mu.Unlock()
	return nil
}

func ListCmdbAttributes(objectID string) []model.CmdbAttribute {
	if GormOK() {
		var rows []model.CmdbAttribute
		if err := GetDB().Where("object_id = ?", objectID).Order("sort asc").Find(&rows).Error; err != nil {
			logrus.WithError(err).Warn("cmdb: list attributes failed")
			return nil
		}
		return rows
	}
	mu.RLock()
	defer mu.RUnlock()
	var out []model.CmdbAttribute
	for _, a := range cmdbAttributes {
		if a.ObjectID == objectID {
			out = append(out, a)
		}
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Sort < out[j].Sort })
	return out
}

func CreateCmdbAttribute(attr *model.CmdbAttribute) error {
	attr.ID = NewID()
	if GormOK() {
		return GetDB().Create(attr).Error
	}
	mu.Lock()
	cmdbAttributes = append(cmdbAttributes, *attr)
	mu.Unlock()
	return nil
}

func UpdateCmdbAttribute(attr *model.CmdbAttribute) error {
	if GormOK() {
		return GetDB().Save(attr).Error
	}
	mu.Lock()
	for i := range cmdbAttributes {
		if cmdbAttributes[i].ID == attr.ID {
			cmdbAttributes[i] = *attr
			break
		}
	}
	mu.Unlock()
	return nil
}

func DeleteCmdbAttribute(id string) error {
	if GormOK() {
		return GetDB().Where("id = ?", id).Delete(&model.CmdbAttribute{}).Error
	}
	mu.Lock()
	for i := range cmdbAttributes {
		if cmdbAttributes[i].ID == id {
			cmdbAttributes = append(cmdbAttributes[:i], cmdbAttributes[i+1:]...)
			break
		}
	}
	mu.Unlock()
	return nil
}

func ListCmdbInstances(objectID string, page, limit int) ([]model.CmdbInstance, int64) {
	if GormOK() {
		var rows []model.CmdbInstance
		var total int64
		q := GetDB().Where("object_id = ?", objectID)
		q.Model(&model.CmdbInstance{}).Count(&total)
		if limit <= 0 {
			limit = 20
		}
		offset := (page - 1) * limit
		if err := q.Order("created_at desc").Offset(offset).Limit(limit).Find(&rows).Error; err != nil {
			logrus.WithError(err).Warn("cmdb: list instances failed")
			return nil, 0
		}
		return rows, total
	}
	mu.RLock()
	defer mu.RUnlock()
	var filtered []model.CmdbInstance
	for _, inst := range cmdbInstances {
		if inst.ObjectID == objectID {
			filtered = append(filtered, inst)
		}
	}
	sort.Slice(filtered, func(i, j int) bool { return filtered[i].CreatedAt.After(filtered[j].CreatedAt) })
	total := int64(len(filtered))
	start := (page - 1) * limit
	if start >= len(filtered) {
		return nil, total
	}
	end := start + limit
	if end > len(filtered) {
		end = len(filtered)
	}
	return filtered[start:end], total
}

func GetCmdbInstance(id string) (*model.CmdbInstance, bool) {
	if GormOK() {
		var inst model.CmdbInstance
		if err := GetDB().Where("id = ?", id).First(&inst).Error; err != nil {
			return nil, false
		}
		return &inst, true
	}
	mu.RLock()
	defer mu.RUnlock()
	for i := range cmdbInstances {
		if cmdbInstances[i].ID == id {
			cp := cmdbInstances[i]
			return &cp, true
		}
	}
	return nil, false
}

func CreateCmdbInstance(inst *model.CmdbInstance) error {
	inst.ID = NewID()
	now := time.Now()
	inst.CreatedAt = now
	inst.UpdatedAt = now
	if GormOK() {
		return GetDB().Create(inst).Error
	}
	mu.Lock()
	cmdbInstances = append(cmdbInstances, *inst)
	mu.Unlock()
	return nil
}

func UpdateCmdbInstance(inst *model.CmdbInstance) error {
	inst.UpdatedAt = time.Now()
	if GormOK() {
		return GetDB().Save(inst).Error
	}
	mu.Lock()
	for i := range cmdbInstances {
		if cmdbInstances[i].ID == inst.ID {
			cmdbInstances[i] = *inst
			break
		}
	}
	mu.Unlock()
	return nil
}

func DeleteCmdbInstance(id string) error {
	if GormOK() {
		return GetDB().Where("id = ?", id).Delete(&model.CmdbInstance{}).Error
	}
	mu.Lock()
	for i := range cmdbInstances {
		if cmdbInstances[i].ID == id {
			cmdbInstances = append(cmdbInstances[:i], cmdbInstances[i+1:]...)
			break
		}
	}
	mu.Unlock()
	return nil
}

func CountCmdbInstancesByObject() map[string]int64 {
	if GormOK() {
		type result struct {
			ObjectID string
			Count    int64
		}
		var rows []result
		GetDB().Model(&model.CmdbInstance{}).Select("object_id, count(*) as count").Group("object_id").Find(&rows)
		m := make(map[string]int64, len(rows))
		for _, r := range rows {
			m[r.ObjectID] = r.Count
		}
		return m
	}
	mu.RLock()
	defer mu.RUnlock()
	m := make(map[string]int64)
	for _, inst := range cmdbInstances {
		m[inst.ObjectID]++
	}
	return m
}

func CreateCmdbDeployTask(task *model.CmdbDeployTask) error {
	if task.ID == "" {
		task.ID = "deploy-" + NewID()
	}
	now := time.Now()
	task.CreatedAt = now
	task.UpdatedAt = now
	if GormOK() {
		if err := ensureCmdbDeployTaskMigrated(); err != nil {
			return err
		}
		return GetDB().Create(task).Error
	}
	mu.Lock()
	cmdbDeployTasks = append(cmdbDeployTasks, *task)
	mu.Unlock()
	return nil
}

func ListCmdbDeployTasks(page, limit int) ([]model.CmdbDeployTask, int64) {
	if limit <= 0 {
		limit = 20
	}
	if page < 1 {
		page = 1
	}
	if GormOK() {
		if err := ensureCmdbDeployTaskMigrated(); err != nil {
			logrus.WithError(err).Warn("cmdb: migrate deploy tasks failed")
			return nil, 0
		}
		var rows []model.CmdbDeployTask
		var total int64
		q := GetDB().Model(&model.CmdbDeployTask{})
		q.Count(&total)
		offset := (page - 1) * limit
		if err := q.Order("created_at desc").Offset(offset).Limit(limit).Find(&rows).Error; err != nil {
			logrus.WithError(err).Warn("cmdb: list deploy tasks failed")
			return nil, 0
		}
		return rows, total
	}
	mu.RLock()
	defer mu.RUnlock()
	out := make([]model.CmdbDeployTask, len(cmdbDeployTasks))
	copy(out, cmdbDeployTasks)
	sort.Slice(out, func(i, j int) bool { return out[i].CreatedAt.After(out[j].CreatedAt) })
	total := int64(len(out))
	start := (page - 1) * limit
	if start >= len(out) {
		return nil, total
	}
	end := start + limit
	if end > len(out) {
		end = len(out)
	}
	return out[start:end], total
}

func GetCmdbDeployTask(id string) (*model.CmdbDeployTask, bool) {
	if GormOK() {
		if err := ensureCmdbDeployTaskMigrated(); err != nil {
			logrus.WithError(err).Warn("cmdb: migrate deploy tasks failed")
			return nil, false
		}
		var task model.CmdbDeployTask
		if err := GetDB().Where("id = ?", id).First(&task).Error; err != nil {
			return nil, false
		}
		return &task, true
	}
	mu.RLock()
	defer mu.RUnlock()
	for i := range cmdbDeployTasks {
		if cmdbDeployTasks[i].ID == id {
			cp := cmdbDeployTasks[i]
			return &cp, true
		}
	}
	return nil, false
}

func ensureCmdbDeployTaskMigrated() error {
	cmdbDeployMigrateOnce.Do(func() {
		cmdbDeployMigrateErr = GetDB().AutoMigrate(&model.CmdbDeployTask{})
	})
	return cmdbDeployMigrateErr
}
