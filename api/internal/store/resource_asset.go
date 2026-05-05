package store

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"sort"
	"time"

	"ai-workbench-api/internal/model"

	"github.com/sirupsen/logrus"
)

var resourceGroups = map[string]*model.ResourceGroup{}

func ListResourceGroups() []model.ResourceGroup {
	if mysqlOK {
		if rows, err := db.Query(`SELECT id,name,description,workspace_id,parent_id,status,COALESCE(tags,'[]'),created_at,updated_at FROM resource_groups ORDER BY updated_at DESC`); err == nil {
			defer rows.Close()
			groups, scanErr := scanResourceGroups(rows)
			if scanErr == nil {
				return groups
			}
			logrus.WithError(scanErr).Warn("resource group mysql scan failed, using memory fallback")
		}
	}
	mu.RLock()
	defer mu.RUnlock()
	out := make([]model.ResourceGroup, 0, len(resourceGroups))
	for _, group := range resourceGroups {
		out = append(out, copyResourceGroup(group))
	}
	sort.Slice(out, func(i, j int) bool { return out[i].UpdatedAt.After(out[j].UpdatedAt) })
	return out
}

func GetResourceGroup(id string) (model.ResourceGroup, bool, error) {
	if mysqlOK {
		var group model.ResourceGroup
		var tags string
		row := db.QueryRow(`SELECT id,name,description,workspace_id,parent_id,status,COALESCE(tags,'[]'),created_at,updated_at FROM resource_groups WHERE id=?`, id)
		err := row.Scan(&group.ID, &group.Name, &group.Description, &group.WorkspaceID, &group.ParentID, &group.Status, &tags, &group.CreatedAt, &group.UpdatedAt)
		if err == nil {
			if err := json.Unmarshal([]byte(tags), &group.Tags); err != nil {
				return model.ResourceGroup{}, false, fmt.Errorf("resource group tags decode failed: %w", err)
			}
			return group, true, nil
		}
		if !errors.Is(err, sql.ErrNoRows) {
			return model.ResourceGroup{}, false, err
		}
	}
	mu.RLock()
	defer mu.RUnlock()
	group, ok := resourceGroups[id]
	if !ok {
		return model.ResourceGroup{}, false, nil
	}
	return copyResourceGroup(group), true, nil
}

func SaveResourceGroup(group model.ResourceGroup) (model.ResourceGroup, error) {
	now := time.Now()
	if group.ID == "" {
		group.ID = NewID()
	}
	if group.CreatedAt.IsZero() {
		group.CreatedAt = now
	}
	group.UpdatedAt = now
	if mysqlOK {
		tags, err := json.Marshal(group.Tags)
		if err != nil {
			return model.ResourceGroup{}, err
		}
		if _, err := db.Exec(`REPLACE INTO resource_groups (id,name,description,workspace_id,parent_id,status,tags,created_at,updated_at) VALUES (?,?,?,?,?,?,?,?,?)`, group.ID, group.Name, group.Description, group.WorkspaceID, group.ParentID, group.Status, string(tags), group.CreatedAt, group.UpdatedAt); err != nil {
			return model.ResourceGroup{}, err
		}
	}
	mu.Lock()
	cp := copyResourceGroup(&group)
	resourceGroups[group.ID] = &cp
	mu.Unlock()
	return group, nil
}

func DeleteResourceGroup(id string) (bool, error) {
	found := false
	if mysqlOK {
		res, err := db.Exec(`DELETE FROM resource_groups WHERE id=?`, id)
		if err != nil {
			return false, err
		}
		if res != nil {
			rows, rowErr := res.RowsAffected()
			if rowErr != nil {
				return false, rowErr
			}
			if rows > 0 {
				found = true
			}
		}
	}
	mu.Lock()
	if _, ok := resourceGroups[id]; ok {
		delete(resourceGroups, id)
		found = true
	}
	mu.Unlock()
	return found, nil
}

func scanResourceGroups(rows *sql.Rows) ([]model.ResourceGroup, error) {
	out := []model.ResourceGroup{}
	for rows.Next() {
		var group model.ResourceGroup
		var tags string
		err := rows.Scan(&group.ID, &group.Name, &group.Description, &group.WorkspaceID, &group.ParentID, &group.Status, &tags, &group.CreatedAt, &group.UpdatedAt)
		if err != nil {
			return nil, fmt.Errorf("resource group row scan failed: %w", err)
		}
		if err := json.Unmarshal([]byte(tags), &group.Tags); err != nil {
			return nil, fmt.Errorf("resource group tags decode failed: %w", err)
		}
		out = append(out, group)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("resource group rows failed: %w", err)
	}
	return out, nil
}

func copyResourceGroup(in *model.ResourceGroup) model.ResourceGroup {
	if in == nil {
		return model.ResourceGroup{}
	}
	cp := *in
	cp.Tags = append([]string{}, in.Tags...)
	return cp
}
