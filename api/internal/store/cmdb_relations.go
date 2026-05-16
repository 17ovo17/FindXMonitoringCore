package store

import (
	"sort"
	"time"

	"ai-workbench-api/internal/model"

	"github.com/sirupsen/logrus"
)

var (
	cmdbRelationTypes     []model.CmdbRelationType
	cmdbInstanceRelations []model.CmdbInstanceRelation
)

func CreateCmdbRelationType(rel *model.CmdbRelationType) error {
	if rel.ID == "" {
		rel.ID = NewID()
	}
	if GormOK() {
		return GetDB().Create(rel).Error
	}
	mu.Lock()
	defer mu.Unlock()
	for i := range cmdbRelationTypes {
		if cmdbRelationTypes[i].ID == rel.ID {
			cmdbRelationTypes[i] = *rel
			return nil
		}
	}
	cmdbRelationTypes = append(cmdbRelationTypes, *rel)
	return nil
}

func GetCmdbRelationType(id string) (*model.CmdbRelationType, bool) {
	if GormOK() {
		var rel model.CmdbRelationType
		if err := GetDB().Where("id = ?", id).First(&rel).Error; err != nil {
			return nil, false
		}
		return &rel, true
	}
	mu.RLock()
	defer mu.RUnlock()
	for i := range cmdbRelationTypes {
		if cmdbRelationTypes[i].ID == id {
			cp := cmdbRelationTypes[i]
			return &cp, true
		}
	}
	return nil, false
}

func CreateCmdbInstanceRelation(rel *model.CmdbInstanceRelation) error {
	if rel.ID == "" {
		rel.ID = NewID()
	}
	if rel.CreatedAt.IsZero() {
		rel.CreatedAt = time.Now()
	}
	if GormOK() {
		return GetDB().Create(rel).Error
	}
	mu.Lock()
	defer mu.Unlock()
	for i := range cmdbInstanceRelations {
		if cmdbInstanceRelations[i].ID == rel.ID {
			cmdbInstanceRelations[i] = *rel
			return nil
		}
	}
	cmdbInstanceRelations = append(cmdbInstanceRelations, *rel)
	return nil
}

func ListCmdbInstanceRelations(instanceID string) []model.CmdbInstanceRelation {
	if GormOK() {
		var rows []model.CmdbInstanceRelation
		err := GetDB().
			Where("source_instance_id = ? OR target_instance_id = ?", instanceID, instanceID).
			Order("created_at asc, id asc").
			Find(&rows).Error
		if err != nil {
			logrus.WithError(err).Warn("cmdb: list instance relations failed")
			return nil
		}
		return rows
	}
	mu.RLock()
	defer mu.RUnlock()
	rows := make([]model.CmdbInstanceRelation, 0)
	for _, rel := range cmdbInstanceRelations {
		if rel.SourceInstanceID == instanceID || rel.TargetInstanceID == instanceID {
			rows = append(rows, rel)
		}
	}
	sort.Slice(rows, func(i, j int) bool {
		if rows[i].CreatedAt.Equal(rows[j].CreatedAt) {
			return rows[i].ID < rows[j].ID
		}
		return rows[i].CreatedAt.Before(rows[j].CreatedAt)
	})
	return rows
}
