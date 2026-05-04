package store

func init() {
	createTableStatements = append(createTableStatements,
		`CREATE TABLE IF NOT EXISTS monitor_audit_logs (id VARCHAR(64) PRIMARY KEY,created_at DATETIME NOT NULL,actor VARCHAR(128),action VARCHAR(128) NOT NULL,resource_type VARCHAR(64),resource_id VARCHAR(128),scope VARCHAR(128),status VARCHAR(32),trace_id VARCHAR(128),client_ip VARCHAR(64),summary TEXT,details JSON,INDEX idx_mal_created(created_at),INDEX idx_mal_action(action),INDEX idx_mal_resource(resource_type,resource_id),INDEX idx_mal_status(status),INDEX idx_mal_trace(trace_id)) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4`,
	)
}
