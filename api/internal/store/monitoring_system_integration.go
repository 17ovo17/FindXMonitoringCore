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

const systemIntegrationBlockedStatus = model.MonitoringSystemIntegrationStatusBlockedByContract

var ErrMonitoringSystemIntegrationBuiltin = fmt.Errorf("builtin system integration cannot be deleted")

func ListMonitoringSystemIntegrations(filter model.MonitoringSystemIntegrationFilter) ([]model.MonitoringSystemIntegration, error) {
	if err := ensureMonitoringSystemIntegrationsSeeded(); err != nil {
		return nil, err
	}
	items, err := monitoringSystemIntegrationItems()
	if err != nil {
		return nil, err
	}
	out := make([]model.MonitoringSystemIntegration, 0, len(items))
	for _, item := range items {
		item = withSystemIntegrationContract(item)
		if monitoringSystemIntegrationMatches(item, filter) {
			out = append(out, copyMonitoringSystemIntegration(item))
		}
	}
	sortMonitoringSystemIntegrations(out)
	return out, nil
}

func GetMonitoringSystemIntegration(id string) (model.MonitoringSystemIntegration, bool, error) {
	if err := ensureMonitoringSystemIntegrationsSeeded(); err != nil {
		return model.MonitoringSystemIntegration{}, false, err
	}
	needle := strings.TrimSpace(id)
	if needle == "" {
		return model.MonitoringSystemIntegration{}, false, nil
	}
	if mysqlOK {
		row := db.QueryRow(monitoringSystemIntegrationSelectSQL()+` WHERE id=?`, needle)
		item, err := scanMonitoringSystemIntegrationRow(row)
		if err == sql.ErrNoRows {
			return model.MonitoringSystemIntegration{}, false, nil
		}
		if err != nil {
			return model.MonitoringSystemIntegration{}, false, err
		}
		return withSystemIntegrationContract(item), true, nil
	}
	mu.RLock()
	item, ok := monitorSystemIntegrations[needle]
	mu.RUnlock()
	if !ok {
		return model.MonitoringSystemIntegration{}, false, nil
	}
	return withSystemIntegrationContract(copyMonitoringSystemIntegration(*item)), true, nil
}

func SaveMonitoringSystemIntegration(input model.MonitoringSystemIntegration, actor string) (model.MonitoringSystemIntegration, error) {
	if err := ensureMonitoringSystemIntegrationsSeeded(); err != nil {
		return model.MonitoringSystemIntegration{}, err
	}
	now := time.Now()
	existing, exists, err := GetMonitoringSystemIntegration(input.ID)
	if err != nil {
		return model.MonitoringSystemIntegration{}, err
	}
	item := normalizeMonitoringSystemIntegration(input, &existing, exists, actor, now)
	if !exists {
		setMonitoringSystemIntegrationCreatedAt(&item, now, actor)
	}
	if err := persistMonitoringSystemIntegration(item); err != nil {
		return model.MonitoringSystemIntegration{}, err
	}
	return withSystemIntegrationContract(copyMonitoringSystemIntegration(item)), nil
}

func UpdateMonitoringSystemIntegration(input model.MonitoringSystemIntegration, actor string) (model.MonitoringSystemIntegration, bool, error) {
	if err := ensureMonitoringSystemIntegrationsSeeded(); err != nil {
		return model.MonitoringSystemIntegration{}, false, err
	}
	existing, ok, err := GetMonitoringSystemIntegration(input.ID)
	if err != nil || !ok {
		return model.MonitoringSystemIntegration{}, ok, err
	}
	item := normalizeMonitoringSystemIntegration(input, &existing, true, actor, time.Now())
	if err := persistMonitoringSystemIntegration(item); err != nil {
		return model.MonitoringSystemIntegration{}, true, err
	}
	return withSystemIntegrationContract(copyMonitoringSystemIntegration(item)), true, nil
}

func DeleteMonitoringSystemIntegration(id string) (bool, error) {
	item, ok, err := GetMonitoringSystemIntegration(id)
	if err != nil || !ok {
		return ok, err
	}
	if item.Builtin {
		return true, ErrMonitoringSystemIntegrationBuiltin
	}
	if mysqlOK {
		if _, err := db.Exec(`DELETE FROM monitor_system_integrations WHERE id=?`, item.ID); err != nil {
			return true, err
		}
		return true, nil
	}
	mu.Lock()
	delete(monitorSystemIntegrations, item.ID)
	mu.Unlock()
	return true, nil
}

func UpdateMonitoringSystemIntegrationWeights(inputs []model.MonitoringSystemIntegrationWeightInput, actor string) ([]model.MonitoringSystemIntegration, bool, error) {
	if err := ensureMonitoringSystemIntegrationsSeeded(); err != nil {
		return nil, false, err
	}
	if len(inputs) == 0 {
		return nil, false, fmt.Errorf("weights payload is empty")
	}
	seen := map[string]bool{}
	now := time.Now()
	updates := []model.MonitoringSystemIntegration{}
	for _, input := range inputs {
		id := strings.TrimSpace(input.ID)
		if id == "" || seen[id] {
			return nil, false, fmt.Errorf("invalid integration id")
		}
		seen[id] = true
		item, ok, err := GetMonitoringSystemIntegration(id)
		if err != nil || !ok {
			return nil, ok, err
		}
		item.Weight = input.Weight
		item.UpdateAt = now.Unix()
		item.UpdateBy = actor
		updates = append(updates, item)
	}
	for _, item := range updates {
		if err := persistMonitoringSystemIntegration(item); err != nil {
			return nil, true, err
		}
	}
	items, err := ListMonitoringSystemIntegrations(model.MonitoringSystemIntegrationFilter{})
	return items, true, err
}

func SetMonitoringSystemIntegrationHide(id string, hide bool, actor string) (model.MonitoringSystemIntegration, bool, error) {
	item, ok, err := GetMonitoringSystemIntegration(id)
	if err != nil || !ok {
		return model.MonitoringSystemIntegration{}, ok, err
	}
	item.Hide = hide
	item.ShowInMenu = !hide
	item.UpdateAt = time.Now().Unix()
	item.UpdateBy = actor
	if err := persistMonitoringSystemIntegration(item); err != nil {
		return model.MonitoringSystemIntegration{}, true, err
	}
	return withSystemIntegrationContract(copyMonitoringSystemIntegration(item)), true, nil
}

func ensureMonitoringSystemIntegrationsSeeded() error {
	if mysqlOK {
		var count int
		if err := db.QueryRow(`SELECT COUNT(*) FROM monitor_system_integrations`).Scan(&count); err != nil {
			return err
		}
		if count > 0 {
			return nil
		}
		for _, item := range monitoringSystemIntegrationDefaults() {
			if err := persistMonitoringSystemIntegration(item); err != nil {
				return err
			}
		}
		return nil
	}
	mu.Lock()
	defer mu.Unlock()
	if len(monitorSystemIntegrations) > 0 {
		return nil
	}
	for _, item := range monitoringSystemIntegrationDefaults() {
		cp := copyMonitoringSystemIntegration(item)
		monitorSystemIntegrations[cp.ID] = &cp
	}
	return nil
}

func monitoringSystemIntegrationItems() ([]model.MonitoringSystemIntegration, error) {
	if mysqlOK {
		rows, err := db.Query(monitoringSystemIntegrationSelectSQL() + ` ORDER BY weight ASC,id ASC`)
		if err != nil {
			return nil, err
		}
		defer rows.Close()
		out := []model.MonitoringSystemIntegration{}
		for rows.Next() {
			item, err := scanMonitoringSystemIntegrationRow(rows)
			if err != nil {
				return nil, err
			}
			out = append(out, item)
		}
		return out, rows.Err()
	}
	mu.RLock()
	defer mu.RUnlock()
	out := make([]model.MonitoringSystemIntegration, 0, len(monitorSystemIntegrations))
	for _, item := range monitorSystemIntegrations {
		out = append(out, copyMonitoringSystemIntegration(*item))
	}
	return out, nil
}

func persistMonitoringSystemIntegration(item model.MonitoringSystemIntegration) error {
	item = withSystemIntegrationContract(item)
	if mysqlOK {
		teams, err := json.Marshal(item.TeamIDs)
		if err != nil {
			return fmt.Errorf("marshal integration teams: %w", err)
		}
		_, err = db.Exec(`REPLACE INTO monitor_system_integrations (id,weight,name,url,config_preview,is_private,team_ids,hide,status,builtin,created_by,updated_by,created_at,updated_at) VALUES (?,?,?,?,?,?,?,?,?,?,?,?,FROM_UNIXTIME(?),FROM_UNIXTIME(?))`,
			item.ID, item.Weight, item.Name, item.URL, item.ConfigPreview, item.IsPrivate, string(teams), item.Hide, item.Status, item.Builtin, item.CreateBy, item.UpdateBy, item.CreateAt, item.UpdateAt)
		return err
	}
	mu.Lock()
	cp := copyMonitoringSystemIntegration(item)
	monitorSystemIntegrations[cp.ID] = &cp
	mu.Unlock()
	return nil
}

func normalizeMonitoringSystemIntegration(input model.MonitoringSystemIntegration, existing *model.MonitoringSystemIntegration, exists bool, actor string, now time.Time) model.MonitoringSystemIntegration {
	item := copyMonitoringSystemIntegration(input)
	if item.ID == "" {
		item.ID = NewID()
	}
	if exists {
		item.CreateAt = existing.CreateAt
		item.CreateBy = existing.CreateBy
		item.Builtin = existing.Builtin
	}
	item.Name = strings.TrimSpace(item.Name)
	item.URL = strings.TrimSpace(item.URL)
	item.ConfigPreview = strings.TrimSpace(item.ConfigPreview)
	item.TeamIDs = dedupePositiveInts(item.TeamIDs)
	item.ShowInMenu = !item.Hide
	item.Status = model.MonitoringSystemIntegrationStatusActive
	item.UpdateAt = now.Unix()
	item.UpdateBy = actor
	return item
}

func monitoringSystemIntegrationDefaults() []model.MonitoringSystemIntegration {
	return []model.MonitoringSystemIntegration{
		newSystemIntegration("findx-console-overview", 10, "FindX Console Overview", "/platform?section=overview", false, nil, false, 1764547200, 1764547200),
		newSystemIntegration("findx-infrastructure-workspace", 20, "FindX Infrastructure Workspace", "/assets?section=hosts", true, []int{101, 102}, true, 1764547200, 1764633600),
		newSystemIntegration("findx-operations-knowledge", 30, "FindX Operations Knowledge Portal", "/aiops?section=knowledge", false, nil, true, 1764547200, 1764720000),
	}
}

func newSystemIntegration(id string, weight int, name, previewURL string, private bool, teams []int, hide bool, createAt, updateAt int64) model.MonitoringSystemIntegration {
	return withSystemIntegrationContract(model.MonitoringSystemIntegration{
		ID:            id,
		Weight:        weight,
		Name:          name,
		URL:           previewURL,
		ConfigPreview: previewURL,
		IsPrivate:     private,
		TeamIDs:       append([]int{}, teams...),
		Hide:          hide,
		ShowInMenu:    !hide,
		Status:        model.MonitoringSystemIntegrationStatusActive,
		CreateAt:      createAt,
		UpdateAt:      updateAt,
		CreateBy:      "findx-system",
		UpdateBy:      "findx-system",
		Builtin:       true,
	})
}

func withSystemIntegrationContract(item model.MonitoringSystemIntegration) model.MonitoringSystemIntegration {
	item.ShowInMenu = !item.Hide
	item.Capabilities = systemIntegrationCapabilities()
	item.BlockedActions = systemIntegrationBlockedActions()
	if item.Status == "" || item.Status == "readonly" {
		item.Status = model.MonitoringSystemIntegrationStatusActive
	}
	return item
}

func systemIntegrationCapabilities() model.MonitoringSystemIntegrationCapabilities {
	return model.MonitoringSystemIntegrationCapabilities{
		List:          true,
		Detail:        true,
		Read:          true,
		Write:         true,
		Sort:          true,
		MenuEmbedding: false,
		OpenEmbedded:  false,
		Status:        systemIntegrationBlockedStatus,
		Reason:        "Create, update, delete, sort, and menu visibility are available; menu embedding and embedded opening remain blocked by contract.",
	}
}

func systemIntegrationBlockedActions() []model.MonitoringSystemIntegrationBlockedAction {
	return []model.MonitoringSystemIntegrationBlockedAction{{
		Action: "open_embedded",
		Status: systemIntegrationBlockedStatus,
		Reason: "Embedded iframe/open behavior has no safe backend contract in this slice.",
	}, {
		Action: "menu_embedding",
		Status: systemIntegrationBlockedStatus,
		Reason: "Dynamic menu embedding is not wired in this backend contract; only show/hide metadata is persisted.",
	}}
}

func monitoringSystemIntegrationMatches(item model.MonitoringSystemIntegration, filter model.MonitoringSystemIntegrationFilter) bool {
	status := strings.TrimSpace(filter.Status)
	if strings.EqualFold(status, "readonly") {
		status = model.MonitoringSystemIntegrationStatusActive
	}
	if status != "" && !strings.EqualFold(item.Status, status) {
		return false
	}
	visibility := strings.ToLower(strings.TrimSpace(filter.Visibility))
	if visibility != "" {
		switch visibility {
		case "private":
			if !item.IsPrivate {
				return false
			}
		case "public":
			if item.IsPrivate {
				return false
			}
		case "menu":
			if !item.ShowInMenu {
				return false
			}
		case "hidden":
			if !item.Hide {
				return false
			}
		default:
			return false
		}
	}
	query := strings.ToLower(strings.TrimSpace(filter.Query))
	if query == "" {
		return true
	}
	fields := []string{item.ID, item.Name, item.ConfigPreview, item.Status, item.CreateBy, item.UpdateBy}
	for _, field := range fields {
		if strings.Contains(strings.ToLower(field), query) {
			return true
		}
	}
	return false
}

func copyMonitoringSystemIntegration(in model.MonitoringSystemIntegration) model.MonitoringSystemIntegration {
	in.TeamIDs = append([]int{}, in.TeamIDs...)
	in.BlockedActions = append([]model.MonitoringSystemIntegrationBlockedAction{}, in.BlockedActions...)
	return in
}

func sortMonitoringSystemIntegrations(items []model.MonitoringSystemIntegration) {
	sort.Slice(items, func(i, j int) bool {
		if items[i].Weight == items[j].Weight {
			return items[i].ID < items[j].ID
		}
		return items[i].Weight < items[j].Weight
	})
}

func setMonitoringSystemIntegrationCreatedAt(item *model.MonitoringSystemIntegration, now time.Time, actor string) {
	item.CreateAt = now.Unix()
	item.CreateBy = actor
}
