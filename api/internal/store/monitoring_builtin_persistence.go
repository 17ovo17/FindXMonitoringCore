package store

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"sort"
	"strings"
	"time"

	"ai-workbench-api/internal/model"
)

func customMonitoringBuiltinComponents() []model.MonitoringBuiltinComponent {
	if mysqlOK {
		rows, err := db.Query(`SELECT id,ident,name,COALESCE(logo,''),COALESCE(readme,''),disabled,COALESCE(updated_by,'') FROM monitor_builtin_components ORDER BY ident ASC`)
		if err == nil {
			defer rows.Close()
			out := []model.MonitoringBuiltinComponent{}
			for rows.Next() {
				item := model.MonitoringBuiltinComponent{}
				if scanErr := rows.Scan(&item.ID, &item.Ident, &item.Name, &item.Logo, &item.Readme, &item.Disabled, &item.UpdatedBy); scanErr == nil {
					out = append(out, copyMonitoringBuiltinComponent(item))
				}
			}
			if rows.Err() == nil {
				return out
			}
		}
	}
	mu.RLock()
	defer mu.RUnlock()
	out := make([]model.MonitoringBuiltinComponent, 0, len(monitorBuiltinComponents))
	for _, item := range monitorBuiltinComponents {
		out = append(out, copyMonitoringBuiltinComponent(*item))
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Ident < out[j].Ident })
	return out
}

func customMonitoringBuiltinPayloads() []model.MonitoringBuiltinPayload {
	if mysqlOK {
		rows, err := db.Query(`SELECT id,uuid,type,component_id,COALESCE(cate,''),name,content,COALESCE(updated_by,'') FROM monitor_builtin_payloads ORDER BY type ASC,id ASC`)
		if err == nil {
			defer rows.Close()
			out := []model.MonitoringBuiltinPayload{}
			for rows.Next() {
				item, scanErr := scanMonitoringBuiltinPayload(rows)
				if scanErr == nil {
					out = append(out, copyMonitoringBuiltinPayload(item))
				}
			}
			if rows.Err() == nil {
				return out
			}
		}
	}
	mu.RLock()
	defer mu.RUnlock()
	out := make([]model.MonitoringBuiltinPayload, 0, len(monitorBuiltinPayloads))
	for _, item := range monitorBuiltinPayloads {
		out = append(out, copyMonitoringBuiltinPayload(*item))
	}
	sort.Slice(out, func(i, j int) bool {
		if out[i].Type == out[j].Type {
			return out[i].ID < out[j].ID
		}
		return out[i].Type < out[j].Type
	})
	return out
}

func getCustomMonitoringBuiltinComponent(id string) (model.MonitoringBuiltinComponent, bool) {
	needle := strings.TrimSpace(id)
	if mysqlOK {
		row := db.QueryRow(`SELECT id,ident,name,COALESCE(logo,''),COALESCE(readme,''),disabled,COALESCE(updated_by,'') FROM monitor_builtin_components WHERE id=?`, needle)
		item := model.MonitoringBuiltinComponent{}
		err := row.Scan(&item.ID, &item.Ident, &item.Name, &item.Logo, &item.Readme, &item.Disabled, &item.UpdatedBy)
		return item, err != sql.ErrNoRows && err == nil
	}
	mu.RLock()
	defer mu.RUnlock()
	item, ok := monitorBuiltinComponents[needle]
	if !ok {
		return model.MonitoringBuiltinComponent{}, false
	}
	return copyMonitoringBuiltinComponent(*item), true
}

func getCustomMonitoringBuiltinPayload(id string) (model.MonitoringBuiltinPayload, bool) {
	needle := strings.TrimSpace(id)
	if mysqlOK {
		row := db.QueryRow(`SELECT id,uuid,type,component_id,COALESCE(cate,''),name,content,COALESCE(updated_by,'') FROM monitor_builtin_payloads WHERE id=?`, needle)
		item, err := scanMonitoringBuiltinPayload(row)
		return item, err != sql.ErrNoRows && err == nil
	}
	mu.RLock()
	defer mu.RUnlock()
	item, ok := monitorBuiltinPayloads[needle]
	if !ok {
		return model.MonitoringBuiltinPayload{}, false
	}
	return copyMonitoringBuiltinPayload(*item), true
}

func persistMonitoringBuiltinComponent(item model.MonitoringBuiltinComponent) error {
	if mysqlOK {
		_, err := db.Exec(`REPLACE INTO monitor_builtin_components (id,ident,name,logo,readme,disabled,updated_by,updated_at) VALUES (?,?,?,?,?,?,?,?)`,
			item.ID, item.Ident, item.Name, item.Logo, item.Readme, item.Disabled, item.UpdatedBy, time.Now())
		return err
	}
	mu.Lock()
	cp := copyMonitoringBuiltinComponent(item)
	monitorBuiltinComponents[cp.ID] = &cp
	mu.Unlock()
	return nil
}

func persistMonitoringBuiltinPayload(item model.MonitoringBuiltinPayload) error {
	if mysqlOK {
		_, err := db.Exec(`REPLACE INTO monitor_builtin_payloads (id,uuid,type,component_id,cate,name,content,updated_by,updated_at) VALUES (?,?,?,?,?,?,?,?,?)`,
			item.ID, item.UUID, item.Type, item.ComponentID, item.Cate, item.Name, string(item.Content), item.UpdatedBy, time.Now())
		return err
	}
	mu.Lock()
	cp := copyMonitoringBuiltinPayload(item)
	monitorBuiltinPayloads[cp.ID] = &cp
	mu.Unlock()
	return nil
}

type monitoringBuiltinPayloadScanner interface {
	Scan(dest ...any) error
}

func scanMonitoringBuiltinPayload(row monitoringBuiltinPayloadScanner) (model.MonitoringBuiltinPayload, error) {
	item := model.MonitoringBuiltinPayload{}
	var content string
	if err := row.Scan(&item.ID, &item.UUID, &item.Type, &item.ComponentID, &item.Cate, &item.Name, &content, &item.UpdatedBy); err != nil {
		return item, err
	}
	if !json.Valid([]byte(content)) {
		return item, fmt.Errorf("invalid builtin payload json")
	}
	item.Title = item.Name
	item.Content = json.RawMessage(content)
	return item, nil
}
