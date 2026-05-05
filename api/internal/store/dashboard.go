package store

import (
	"crypto/sha256"
	"database/sql"
	"encoding/json"
	"fmt"
	"sort"
	"strings"
	"time"

	"ai-workbench-api/internal/model"
)

func ListMonitorDashboards() ([]model.MonitorDashboard, error) {
	if mysqlOK {
		rows, err := db.Query(`SELECT id,title,description,workspace_id,resource_group_id,COALESCE(tags,'[]'),COALESCE(variables,'{}'),COALESCE(panels,'[]'),version,status,shared,share_token_hash,created_by,updated_by,created_at,updated_at FROM monitor_dashboards ORDER BY updated_at DESC LIMIT 1000`)
		if err != nil {
			return nil, err
		}
		defer rows.Close()
		return scanMonitorDashboards(rows)
	}
	mu.RLock()
	defer mu.RUnlock()
	out := make([]model.MonitorDashboard, 0, len(monitorDashboards))
	for _, item := range monitorDashboards {
		out = append(out, *copyMonitorDashboard(item))
	}
	sort.Slice(out, func(i, j int) bool { return out[i].UpdatedAt.After(out[j].UpdatedAt) })
	return out, nil
}

func GetMonitorDashboard(id string) (*model.MonitorDashboard, bool, error) {
	if mysqlOK {
		row := db.QueryRow(`SELECT id,title,description,workspace_id,resource_group_id,COALESCE(tags,'[]'),COALESCE(variables,'{}'),COALESCE(panels,'[]'),version,status,shared,share_token_hash,created_by,updated_by,created_at,updated_at FROM monitor_dashboards WHERE id=?`, id)
		item, err := scanMonitorDashboardRow(row)
		if err == sql.ErrNoRows {
			return nil, false, nil
		}
		if err != nil {
			return nil, false, err
		}
		return item, true, nil
	}
	mu.RLock()
	defer mu.RUnlock()
	item, ok := monitorDashboards[id]
	return copyMonitorDashboard(item), ok, nil
}

func SaveMonitorDashboard(input *model.MonitorDashboard, actor string) (*model.MonitorDashboard, error) {
	now := time.Now()
	existing, exists, err := GetMonitorDashboard(input.ID)
	if err != nil {
		return nil, err
	}
	item := normalizeMonitorDashboard(input, existing, actor, now)
	if !exists {
		item.Version = 1
		item.CreatedAt = now
		item.CreatedBy = actor
	}
	mu.Lock()
	monitorDashboards[item.ID] = copyMonitorDashboard(item)
	mu.Unlock()
	if mysqlOK {
		if err := persistMonitorDashboard(item); err != nil {
			return copyMonitorDashboard(item), err
		}
	}
	return copyMonitorDashboard(item), nil
}

func UpdateMonitorDashboard(input *model.MonitorDashboard, actor string) (*model.MonitorDashboard, bool, error) {
	existing, exists, err := GetMonitorDashboard(input.ID)
	if err != nil || !exists {
		return nil, exists, err
	}
	now := time.Now()
	item := normalizeMonitorDashboard(input, existing, actor, now)
	mu.Lock()
	monitorDashboards[item.ID] = copyMonitorDashboard(item)
	mu.Unlock()
	if mysqlOK {
		if err := persistMonitorDashboard(item); err != nil {
			return copyMonitorDashboard(item), true, err
		}
	}
	return copyMonitorDashboard(item), true, nil
}

func DeleteMonitorDashboard(id string) (bool, error) {
	found := false
	mu.Lock()
	if _, ok := monitorDashboards[id]; ok {
		delete(monitorDashboards, id)
		found = true
	}
	mu.Unlock()
	if mysqlOK {
		res, err := db.Exec(`DELETE FROM monitor_dashboards WHERE id=?`, id)
		if err != nil {
			return found, err
		}
		rows, err := res.RowsAffected()
		if err != nil {
			return found, err
		}
		found = found || rows > 0
	}
	return found, nil
}

func CloneMonitorDashboard(id, actor string) (*model.MonitorDashboard, bool, error) {
	item, ok, err := GetMonitorDashboard(id)
	if err != nil || !ok {
		return nil, ok, err
	}
	item.ID = ""
	item.Title = item.Title + " 副本"
	item.Version = 0
	item.Shared = false
	item.ShareTokenSet = false
	item.ShareSummary = ""
	item.CreatedAt = time.Time{}
	out, err := SaveMonitorDashboard(item, actor)
	return out, true, err
}

func ShareMonitorDashboard(id, actor string) (model.MonitorDashboardShareResult, bool, error) {
	item, ok, err := GetMonitorDashboard(id)
	if err != nil || !ok {
		return model.MonitorDashboardShareResult{}, ok, err
	}
	item.Shared = true
	item.ShareTokenSet = true
	item.ShareSummary = "仪表盘分享已启用"
	item.UpdatedBy = actor
	item.UpdatedAt = time.Now()
	mu.Lock()
	if stored := monitorDashboards[item.ID]; stored != nil {
		stored.Shared = true
		stored.ShareTokenSet = true
		stored.ShareSummary = item.ShareSummary
		stored.UpdatedBy = item.UpdatedBy
		stored.UpdatedAt = item.UpdatedAt
	}
	mu.Unlock()
	if mysqlOK {
		tokenHash := dashboardShareHash(item.ID, actor, item.UpdatedAt)
		_, err := db.Exec(`UPDATE monitor_dashboards SET shared=1,share_token_hash=?,updated_by=?,updated_at=? WHERE id=?`, tokenHash, actor, item.UpdatedAt, item.ID)
		if err != nil {
			return model.MonitorDashboardShareResult{}, true, err
		}
	}
	return model.MonitorDashboardShareResult{ID: item.ID, ShareEnabled: true, ShareSummary: item.ShareSummary}, true, nil
}

func normalizeMonitorDashboard(input, existing *model.MonitorDashboard, actor string, now time.Time) *model.MonitorDashboard {
	item := copyMonitorDashboard(input)
	if item.ID == "" {
		item.ID = NewID()
	}
	if existing != nil {
		item.CreatedAt = existing.CreatedAt
		item.CreatedBy = existing.CreatedBy
		item.Version = existing.Version + 1
		item.Shared = existing.Shared
		item.ShareTokenSet = existing.ShareTokenSet
		item.ShareSummary = existing.ShareSummary
	}
	item.Title = strings.TrimSpace(item.Title)
	item.Description = strings.TrimSpace(item.Description)
	item.WorkspaceID = strings.TrimSpace(item.WorkspaceID)
	item.ResourceGroupID = strings.TrimSpace(item.ResourceGroupID)
	item.Tags = dedupeStrings(item.Tags)
	item.Status = firstNonEmpty(item.Status, model.MonitorDashboardStatusActive)
	item.UpdatedBy = actor
	item.UpdatedAt = now
	return item
}

func persistMonitorDashboard(item *model.MonitorDashboard) error {
	tags, err := json.Marshal(item.Tags)
	if err != nil {
		return fmt.Errorf("marshal dashboard tags: %w", err)
	}
	_, err = db.Exec(`REPLACE INTO monitor_dashboards (id,title,description,workspace_id,resource_group_id,tags,variables,panels,version,status,shared,share_token_hash,created_by,updated_by,created_at,updated_at) VALUES (?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?)`,
		item.ID, item.Title, item.Description, item.WorkspaceID, item.ResourceGroupID, string(tags),
		string(item.Variables), string(item.Panels), item.Version, item.Status, item.Shared,
		shareTokenHashForPersist(item), item.CreatedBy, item.UpdatedBy, item.CreatedAt, item.UpdatedAt)
	if err != nil {
		return fmt.Errorf("persist dashboard: %w", err)
	}
	return nil
}

func scanMonitorDashboards(rows *sql.Rows) ([]model.MonitorDashboard, error) {
	out := []model.MonitorDashboard{}
	for rows.Next() {
		item, err := scanMonitorDashboardRow(rows)
		if err != nil {
			return out, err
		}
		out = append(out, *item)
	}
	return out, rows.Err()
}

type monitorDashboardScanner interface {
	Scan(dest ...any) error
}

func scanMonitorDashboardRow(row monitorDashboardScanner) (*model.MonitorDashboard, error) {
	item := model.MonitorDashboard{}
	var tags, variables, panels string
	var shareHash sql.NullString
	if err := row.Scan(&item.ID, &item.Title, &item.Description, &item.WorkspaceID, &item.ResourceGroupID, &tags, &variables, &panels, &item.Version, &item.Status, &item.Shared, &shareHash, &item.CreatedBy, &item.UpdatedBy, &item.CreatedAt, &item.UpdatedAt); err != nil {
		return nil, err
	}
	if err := json.Unmarshal([]byte(tags), &item.Tags); err != nil {
		return nil, fmt.Errorf("scan dashboard tags: %w", err)
	}
	if !json.Valid([]byte(variables)) || !json.Valid([]byte(panels)) {
		return nil, fmt.Errorf("scan dashboard json payload")
	}
	item.Variables = json.RawMessage(variables)
	item.Panels = json.RawMessage(panels)
	item.ShareTokenSet = shareHash.Valid && strings.TrimSpace(shareHash.String) != ""
	if item.Shared {
		item.ShareSummary = "仪表盘分享已启用"
	}
	return &item, nil
}

func copyMonitorDashboard(in *model.MonitorDashboard) *model.MonitorDashboard {
	if in == nil {
		return nil
	}
	cp := *in
	cp.Tags = append([]string{}, in.Tags...)
	cp.Variables = append([]byte{}, in.Variables...)
	cp.Panels = append([]byte{}, in.Panels...)
	return &cp
}

func shareTokenHashForPersist(item *model.MonitorDashboard) string {
	if !item.ShareTokenSet {
		return ""
	}
	return dashboardShareHash(item.ID, item.UpdatedBy, item.UpdatedAt)
}

func dashboardShareHash(id, actor string, at time.Time) string {
	sum := sha256.Sum256([]byte(id + ":" + actor + ":" + at.Format(time.RFC3339Nano)))
	return fmt.Sprintf("%x", sum)
}
