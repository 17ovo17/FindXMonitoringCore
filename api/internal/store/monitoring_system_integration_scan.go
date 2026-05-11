package store

import (
	"encoding/json"
	"fmt"
	"sort"

	"ai-workbench-api/internal/model"
)

type monitoringSystemIntegrationScanner interface {
	Scan(dest ...any) error
}

func monitoringSystemIntegrationSelectSQL() string {
	return `SELECT id,weight,name,url,config_preview,is_private,COALESCE(team_ids,'[]'),hide,status,builtin,COALESCE(created_by,''),COALESCE(updated_by,''),UNIX_TIMESTAMP(created_at),UNIX_TIMESTAMP(updated_at) FROM monitor_system_integrations`
}

func scanMonitoringSystemIntegrationRow(row monitoringSystemIntegrationScanner) (model.MonitoringSystemIntegration, error) {
	item := model.MonitoringSystemIntegration{}
	var teams string
	if err := row.Scan(&item.ID, &item.Weight, &item.Name, &item.URL, &item.ConfigPreview, &item.IsPrivate, &teams, &item.Hide, &item.Status, &item.Builtin, &item.CreateBy, &item.UpdateBy, &item.CreateAt, &item.UpdateAt); err != nil {
		return item, err
	}
	if err := json.Unmarshal([]byte(teams), &item.TeamIDs); err != nil {
		return item, fmt.Errorf("scan integration teams: %w", err)
	}
	return item, nil
}

func dedupePositiveInts(in []int) []int {
	seen := map[int]bool{}
	out := []int{}
	for _, value := range in {
		if value <= 0 || seen[value] {
			continue
		}
		seen[value] = true
		out = append(out, value)
	}
	sort.Ints(out)
	return out
}
