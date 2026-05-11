package store

import (
	"database/sql"
	"encoding/json"

	"ai-workbench-api/internal/model"
)

func persistLogPipeline(item *model.LogPipeline) error {
	_, err := db.Exec(`REPLACE INTO log_pipelines (id,name,version_id,description,enabled,stages,config,created_by,updated_by,created_at,updated_at) VALUES (?,?,?,?,?,?,?,?,?,?,?)`,
		item.ID, item.Name, item.Version, item.Description, item.Enabled, string(item.Stages), string(item.Config), item.CreatedBy, item.UpdatedBy, item.CreatedAt, item.UpdatedAt)
	return err
}

func persistExplorerSavedView(item *model.ExplorerSavedView) error {
	_, err := db.Exec(`REPLACE INTO explorer_saved_views (id,source_page,name,description,query_json,filters,columns_json,time_range,layout,created_by,updated_by,created_at,updated_at) VALUES (?,?,?,?,?,?,?,?,?,?,?,?,?)`,
		item.ID, item.SourcePage, item.Name, item.Description, string(item.Query), string(item.Filters), string(item.Columns), string(item.TimeRange), string(item.Layout), item.CreatedBy, item.UpdatedBy, item.CreatedAt, item.UpdatedAt)
	return err
}

func scanLogPipelines(rows *sql.Rows) ([]model.LogPipeline, error) {
	out := []model.LogPipeline{}
	for rows.Next() {
		item, err := scanLogPipelineRow(rows)
		if err != nil {
			return out, err
		}
		out = append(out, *item)
	}
	return out, rows.Err()
}

type rowScanner interface {
	Scan(dest ...any) error
}

func scanLogPipelineRow(row rowScanner) (*model.LogPipeline, error) {
	item := model.LogPipeline{}
	var stages, config string
	if err := row.Scan(&item.ID, &item.Name, &item.Version, &item.Description, &item.Enabled, &stages, &config, &item.CreatedBy, &item.UpdatedBy, &item.CreatedAt, &item.UpdatedAt); err != nil {
		return nil, err
	}
	item.Stages = json.RawMessage(stages)
	item.Config = json.RawMessage(config)
	return &item, nil
}

func scanExplorerSavedViews(rows *sql.Rows) ([]model.ExplorerSavedView, error) {
	out := []model.ExplorerSavedView{}
	for rows.Next() {
		item, err := scanExplorerSavedViewRow(rows)
		if err != nil {
			return out, err
		}
		out = append(out, *item)
	}
	return out, rows.Err()
}

func scanExplorerSavedViewRow(row rowScanner) (*model.ExplorerSavedView, error) {
	item := model.ExplorerSavedView{}
	var query, filters, columns, timeRange, layout string
	err := row.Scan(&item.ID, &item.SourcePage, &item.Name, &item.Description, &query, &filters, &columns, &timeRange, &layout, &item.CreatedBy, &item.UpdatedBy, &item.CreatedAt, &item.UpdatedAt)
	if err != nil {
		return nil, err
	}
	item.Query = json.RawMessage(query)
	item.Filters = json.RawMessage(filters)
	item.Columns = json.RawMessage(columns)
	item.TimeRange = json.RawMessage(timeRange)
	item.Layout = json.RawMessage(layout)
	return &item, nil
}
