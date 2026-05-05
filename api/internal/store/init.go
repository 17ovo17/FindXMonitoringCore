package store

import (
	"context"
	"database/sql"
	"encoding/json"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"ai-workbench-api/internal/model"

	_ "github.com/go-sql-driver/mysql"
	"github.com/redis/go-redis/v9"
	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
)

var (
	mu                   sync.RWMutex
	records              []*model.DiagnoseRecord
	agents               = map[string]*model.CatpawAgent{}
	credentials          = map[string]*model.Credential{}
	alerts               []*model.AlertRecord
	chatSessions         = map[string]*model.ChatSession{}
	chatMessages         = map[string][]model.ChatMessage{}
	topologyNodes        = map[string]*model.TopologyNode{}
	topologyEdges        = map[string]*model.TopologyEdge{}
	topologyBusinesses   = map[string]*model.TopologyBusiness{}
	diagnosisCases       = map[string]*model.DiagnosisCase{}
	metricsMappings      = map[string]*model.MetricsMapping{}
	monitorTargets       = map[string]*model.MonitorTarget{}
	findxAgents          = map[string]*model.FindXAgent{}
	monitorDashboards    = map[string]*model.MonitorDashboard{}
	monitorAlertRules    = map[string]*model.MonitorAlertRule{}
	monitorRuleVersions  = map[string][]model.MonitorAlertRuleVersion{}
	monitorEventsCurrent = map[string]*model.MonitorAlertEvent{}
	monitorEventsHistory = map[string]*model.MonitorAlertEvent{}
	monitorEventActions  = map[string][]model.MonitorAlertAction{}
	runbooks             = map[string]*model.Runbook{}
	runbookExecs         = map[string]*model.RunbookExecution{}
	diagnosisFeedbacks   []*model.DiagnosisFeedback
	aiSettings           = map[string]*model.AISetting{}
	knowledgeDocs        = map[string]*model.KnowledgeDocument{}
	searchEvents         []*model.KnowledgeSearchEvent
	searchBadcases       []*model.KnowledgeSearchBadcase
	auditEvents          []AuditEvent
	maxRecs              = 100

	db          *sql.DB
	redisClient *redis.Client
	mysqlOK     bool
	redisOK     bool
	lastDBError string
)

const (
	newIDNumberBase = 10
	newIDSeparator  = "-"
)

var newIDSequence uint64

// AuditEvent represents an audit log entry.
type AuditEvent struct {
	ID            string    `json:"id"`
	Action        string    `json:"action"`
	Target        string    `json:"target"`
	Risk          string    `json:"risk"`
	Decision      string    `json:"decision"`
	Detail        string    `json:"detail"`
	Operator      string    `json:"operator"`
	Timestamp     string    `json:"timestamp"`
	Description   string    `json:"description"`
	TestBatchID   string    `json:"test_batch_id"`
	ClientIP      string    `json:"client_ip"`
	CreatedAt     time.Time `json:"created_at"`
	IntegrityHash string    `json:"integrity_hash,omitempty"`
}

type fallbackSnapshot struct {
	ChatSessions       map[string]*model.ChatSession      `json:"chat_sessions"`
	ChatMessages       map[string][]model.ChatMessage     `json:"chat_messages"`
	TopologyBusinesses map[string]*model.TopologyBusiness `json:"topology_businesses"`
}

func fallbackSnapshotPath() string {
	if value := strings.TrimSpace(viper.GetString("storage.fallback_file")); value != "" {
		return value
	}
	return filepath.Join("data", "memory-store.json")
}

func loadFallbackSnapshot() {
	if mysqlOK {
		return
	}
	path := fallbackSnapshotPath()
	data, err := os.ReadFile(path)
	if err != nil || len(data) == 0 {
		return
	}
	var snap fallbackSnapshot
	if err := json.Unmarshal(data, &snap); err != nil {
		logrus.Warnf("memory fallback snapshot ignored: %v", err)
		return
	}
	mu.Lock()
	if snap.ChatSessions != nil {
		chatSessions = snap.ChatSessions
	}
	if snap.ChatMessages != nil {
		chatMessages = snap.ChatMessages
	}
	if snap.TopologyBusinesses != nil {
		topologyBusinesses = snap.TopologyBusinesses
	}
	mu.Unlock()
	logrus.Infof("loaded memory fallback snapshot from %s", path)
}

func persistFallbackSnapshot() {
	if mysqlOK {
		return
	}
	mu.RLock()
	snap := fallbackSnapshot{
		ChatSessions:       map[string]*model.ChatSession{},
		ChatMessages:       map[string][]model.ChatMessage{},
		TopologyBusinesses: map[string]*model.TopologyBusiness{},
	}
	for id, session := range chatSessions {
		cp := *session
		snap.ChatSessions[id] = &cp
	}
	for id, messages := range chatMessages {
		snap.ChatMessages[id] = append([]model.ChatMessage{}, messages...)
	}
	for id, business := range topologyBusinesses {
		cp := *business
		snap.TopologyBusinesses[id] = &cp
	}
	mu.RUnlock()
	path := fallbackSnapshotPath()
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		logrus.Warnf("memory fallback snapshot mkdir failed: %v", err)
		return
	}
	data, err := json.MarshalIndent(snap, "", "  ")
	if err != nil {
		logrus.Warnf("memory fallback snapshot marshal failed: %v", err)
		return
	}
	tmp := path + ".tmp"
	if err := os.WriteFile(tmp, data, 0644); err != nil {
		logrus.Warnf("memory fallback snapshot write failed: %v", err)
		return
	}
	if err := os.Rename(tmp, path); err != nil {
		logrus.Warnf("memory fallback snapshot replace failed: %v", err)
	}
}

// Init initializes MySQL, Redis, topology seed, and fallback snapshot.
func Init() {
	initMySQL()
	initRedis()
	seedPlatformTopology()
	loadFallbackSnapshot()
}

// Health returns the health status of storage backends.
func Health() map[string]any {
	out := map[string]any{"mysql": mysqlOK, "redis": redisOK}
	if lastDBError != "" {
		out["error"] = lastDBError
	}
	return out
}

func initMySQL() {
	dsn := strings.TrimSpace(viper.GetString("mysql.dsn"))
	if dsn == "" {
		lastDBError = "mysql.dsn is empty"
		logrus.Warn("mysql not configured, using memory store")
		return
	}
	conn, err := sql.Open("mysql", dsn)
	if err != nil {
		lastDBError = err.Error()
		logrus.Warnf("mysql disabled: %v", err)
		return
	}
	conn.SetMaxOpenConns(10)
	conn.SetMaxIdleConns(3)
	conn.SetConnMaxLifetime(time.Hour)
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	if err := conn.PingContext(ctx); err != nil {
		lastDBError = err.Error()
		logrus.Warnf("mysql unavailable, using memory store: %v", err)
		return
	}
	db = conn
	if err := migrate(); err != nil {
		lastDBError = err.Error()
		logrus.Warnf("mysql migrate failed, using memory store: %v", err)
		return
	}
	mysqlOK = true
	lastDBError = ""
}

func initRedis() {
	addr := strings.TrimSpace(viper.GetString("redis.addr"))
	if addr == "" {
		logrus.Warn("redis not configured, using mysql/memory for online state")
		return
	}
	client := redis.NewClient(&redis.Options{Addr: addr, Password: viper.GetString("redis.password"), DB: viper.GetInt("redis.db")})
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	if err := client.Ping(ctx).Err(); err != nil {
		logrus.Warnf("redis unavailable, using mysql/memory for online state: %v", err)
		return
	}
	redisClient = client
	redisOK = true
}

func migrate() error {
	for _, stmt := range createTableStatements {
		if _, err := db.Exec(stmt); err != nil {
			return err
		}
	}
	for _, stmt := range tolerantMigrationStatements {
		_, _ = db.Exec(stmt)
	}
	return nil
}

var createTableStatements = []string{
	`CREATE TABLE IF NOT EXISTS diagnose_records (id varchar(64) PRIMARY KEY,target_ip varchar(64),` + "`trigger`" + ` varchar(32),source varchar(64),data_source varchar(64),status varchar(32),report longtext,summary_report longtext,raw_report longtext,alert_title varchar(255),create_time datetime,end_time datetime NULL,INDEX idx_diag_ip_time(target_ip,create_time)) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4`,
	`CREATE TABLE IF NOT EXISTS catpaw_agents (ip varchar(64) PRIMARY KEY,hostname varchar(255),version varchar(64),last_seen datetime) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4`,
	`CREATE TABLE IF NOT EXISTS credentials (id varchar(64) PRIMARY KEY,name varchar(255),protocol varchar(32),username varchar(255),password text,ssh_key longtext,port int,remark text) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4`,
	`CREATE TABLE IF NOT EXISTS alerts (id varchar(64) PRIMARY KEY,title varchar(255),target_ip varchar(64),severity varchar(32),status varchar(32),labels json,source varchar(64),create_time datetime,resolved_at datetime NULL,INDEX idx_alert_time(create_time)) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4`,
	`CREATE TABLE IF NOT EXISTS chat_sessions (id varchar(64) PRIMARY KEY,title varchar(255),model varchar(128),target_ip varchar(64),created_at datetime,updated_at datetime,INDEX idx_chat_updated(updated_at)) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4`,
	`CREATE TABLE IF NOT EXISTS chat_messages (id varchar(64) PRIMARY KEY,session_id varchar(64),role varchar(32),content longtext,model varchar(128),target_ip varchar(64),created_at datetime,INDEX idx_msg_session(session_id,created_at)) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4`,
	`CREATE TABLE IF NOT EXISTS topology_nodes (id varchar(64) PRIMARY KEY,name varchar(255),type varchar(64),ip varchar(64),service_name varchar(255),port int,status varchar(32),x double,y double,meta text,created_at datetime,updated_at datetime) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4`,
	`CREATE TABLE IF NOT EXISTS topology_edges (id varchar(64) PRIMARY KEY,source_id varchar(64),target_id varchar(64),protocol varchar(64),direction varchar(32),label varchar(255),status varchar(32),latency_ms int,error text,created_at datetime,updated_at datetime,UNIQUE KEY uniq_edge(source_id,target_id,protocol)) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4`,
	`CREATE TABLE IF NOT EXISTS topology_businesses (id varchar(64) PRIMARY KEY,name varchar(255),hosts json,endpoints json,attributes json,graph longtext,created_at datetime,updated_at datetime,INDEX idx_topology_business_updated(updated_at)) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4`,
	`CREATE TABLE IF NOT EXISTS audit_events (id varchar(64) PRIMARY KEY,action varchar(128),target varchar(255),risk varchar(32),decision varchar(64),detail text,operator varchar(128),description text,test_batch_id varchar(128),client_ip varchar(64),created_at datetime,INDEX idx_audit_batch(test_batch_id),INDEX idx_audit_time(created_at)) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4`,
	`CREATE TABLE IF NOT EXISTS user_profiles (id varchar(64) PRIMARY KEY,name varchar(255) NOT NULL,hosts json,endpoints json,description text,created_at datetime,updated_at datetime) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4`,
	`CREATE TABLE IF NOT EXISTS users (id varchar(64) PRIMARY KEY,username varchar(128) UNIQUE NOT NULL,password_hash varchar(255) NOT NULL,role varchar(32) DEFAULT 'user',must_change_pwd tinyint(1) DEFAULT 0,created_at datetime,updated_at datetime) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4`,
	// GB-001: knowledge base tables
	`CREATE TABLE IF NOT EXISTS diagnosis_cases (id VARCHAR(64) PRIMARY KEY,metric_snapshot JSON NOT NULL COMMENT '指标异常快照',root_cause_category VARCHAR(64) NOT NULL COMMENT '根因分类',root_cause_description TEXT NOT NULL COMMENT '根因描述',treatment_steps TEXT NOT NULL COMMENT '处置方案',keywords VARCHAR(500) COMMENT '关键词标签，逗号分隔',source_diagnosis_id VARCHAR(64) COMMENT '来源诊断记录ID',dify_document_id VARCHAR(64) COMMENT 'Dify知识库文档ID',created_at DATETIME NOT NULL,created_by VARCHAR(128),evaluation_avg DECIMAL(3,1) DEFAULT 0 COMMENT '平均评价分',INDEX idx_case_category(root_cause_category),FULLTEXT INDEX idx_case_keywords(keywords,root_cause_description)) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4`,
	`CREATE TABLE IF NOT EXISTS diagnosis_feedback (id VARCHAR(64) PRIMARY KEY,diagnosis_id VARCHAR(64) NOT NULL COMMENT '关联诊断记录ID',user VARCHAR(128) NOT NULL,rating ENUM('accurate','partial','inaccurate') NOT NULL,comment TEXT,dify_message_id VARCHAR(64) COMMENT 'Dify消息ID',created_at DATETIME NOT NULL,INDEX idx_feedback_diagnosis(diagnosis_id)) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4`,
	`CREATE TABLE IF NOT EXISTS ai_settings (id VARCHAR(64) PRIMARY KEY,setting_key VARCHAR(128) UNIQUE NOT NULL,setting_value TEXT,updated_at DATETIME NOT NULL) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4`,
	`CREATE TABLE IF NOT EXISTS runbooks (id VARCHAR(64) PRIMARY KEY,title VARCHAR(255) NOT NULL,category VARCHAR(64) NOT NULL COMMENT '分类',trigger_conditions JSON COMMENT '触发条件',steps TEXT NOT NULL COMMENT '处置步骤Markdown',auto_executable TINYINT(1) DEFAULT 0,dify_document_id VARCHAR(64),created_at DATETIME NOT NULL,updated_at DATETIME NOT NULL,INDEX idx_runbook_category(category)) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4`,
	`CREATE TABLE IF NOT EXISTS metrics_mappings (id VARCHAR(64) PRIMARY KEY,datasource_id VARCHAR(64) NOT NULL COMMENT '关联数据源',raw_name VARCHAR(255) NOT NULL COMMENT '原始指标名',standard_name VARCHAR(255) COMMENT '标准名',exporter VARCHAR(64) COMMENT '来源exporter',description VARCHAR(500) COMMENT '中文描述',transform VARCHAR(500) COMMENT 'PromQL转换公式',status ENUM('auto','confirmed','custom','unmapped') DEFAULT 'unmapped',created_at DATETIME NOT NULL,updated_at DATETIME NOT NULL,UNIQUE KEY uniq_ds_raw(datasource_id,raw_name),INDEX idx_mapping_standard(standard_name),INDEX idx_mapping_status(status)) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4`,
	`CREATE TABLE IF NOT EXISTS monitor_targets (id VARCHAR(64) PRIMARY KEY,ident VARCHAR(255) UNIQUE NOT NULL,name VARCHAR(255),ip VARCHAR(64),hostname VARCHAR(255),os VARCHAR(64),arch VARCHAR(64),environment VARCHAR(64),business_group VARCHAR(128),owner VARCHAR(128),status VARCHAR(32),source VARCHAR(64),labels JSON,metadata JSON,last_seen DATETIME NULL,created_at DATETIME NOT NULL,updated_at DATETIME NOT NULL,INDEX idx_monitor_target_ip(ip),INDEX idx_monitor_target_status(status),INDEX idx_monitor_target_bg(business_group)) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4`,
	`CREATE TABLE IF NOT EXISTS findx_agents (id VARCHAR(64) PRIMARY KEY,ident VARCHAR(255) UNIQUE NOT NULL,target_id VARCHAR(64),ip VARCHAR(64),hostname VARCHAR(255),os VARCHAR(64),arch VARCHAR(64),version VARCHAR(128),collector VARCHAR(64),status VARCHAR(32),capabilities JSON,global_labels JSON,config_version VARCHAR(128),last_seen DATETIME NOT NULL,created_at DATETIME NOT NULL,updated_at DATETIME NOT NULL,INDEX idx_findx_agent_ip(ip),INDEX idx_findx_agent_status(status),INDEX idx_findx_agent_target(target_id)) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4`,
	`CREATE TABLE IF NOT EXISTS monitor_dashboards (id VARCHAR(64) PRIMARY KEY,title VARCHAR(255) NOT NULL,description TEXT,workspace_id VARCHAR(64),resource_group_id VARCHAR(64),tags JSON,variables JSON,panels JSON,version INT NOT NULL DEFAULT 1,status VARCHAR(32),shared TINYINT(1) DEFAULT 0,share_token_hash VARCHAR(128),created_by VARCHAR(128),updated_by VARCHAR(128),created_at DATETIME NOT NULL,updated_at DATETIME NOT NULL,INDEX idx_md_workspace(workspace_id),INDEX idx_md_resource_group(resource_group_id),INDEX idx_md_status(status),INDEX idx_md_updated(updated_at)) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4`,
	`CREATE TABLE IF NOT EXISTS resource_groups (id VARCHAR(64) PRIMARY KEY,name VARCHAR(255) NOT NULL,description TEXT,workspace_id VARCHAR(64),parent_id VARCHAR(64),status VARCHAR(32),tags JSON,created_at DATETIME NOT NULL,updated_at DATETIME NOT NULL,INDEX idx_rg_workspace_status(workspace_id,status),INDEX idx_rg_status(status)) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4`,
	`CREATE TABLE IF NOT EXISTS monitor_alert_rules (id VARCHAR(64) PRIMARY KEY,name VARCHAR(255) NOT NULL,query LONGTEXT NOT NULL,severity VARCHAR(32) NOT NULL,datasource_id VARCHAR(64) NOT NULL,target_selector JSON,labels JSON,annotations JSON,enabled TINYINT(1) DEFAULT 1,version INT NOT NULL DEFAULT 1,for_duration VARCHAR(64),no_data_policy VARCHAR(32),status VARCHAR(32),created_by VARCHAR(128),updated_by VARCHAR(128),created_at DATETIME NOT NULL,updated_at DATETIME NOT NULL,INDEX idx_mar_ds(datasource_id),INDEX idx_mar_status(status),INDEX idx_mar_enabled(enabled)) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4`,
	`CREATE TABLE IF NOT EXISTS monitor_alert_rule_versions (id VARCHAR(96) PRIMARY KEY,rule_id VARCHAR(64) NOT NULL,version INT NOT NULL,name VARCHAR(255) NOT NULL,query LONGTEXT NOT NULL,severity VARCHAR(32) NOT NULL,datasource_id VARCHAR(64) NOT NULL,target_selector JSON,labels JSON,annotations JSON,enabled TINYINT(1) DEFAULT 1,for_duration VARCHAR(64),no_data_policy VARCHAR(32),status VARCHAR(32),created_by VARCHAR(128),created_at DATETIME NOT NULL,UNIQUE KEY uniq_marv_rule_version(rule_id,version),INDEX idx_marv_rule(rule_id)) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4`,
	`CREATE TABLE IF NOT EXISTS monitor_alert_rule_eval_logs (id VARCHAR(64) PRIMARY KEY,rule_id VARCHAR(64),rule_version INT,status VARCHAR(32),message TEXT,details JSON,started_at DATETIME NOT NULL,finished_at DATETIME NOT NULL,duration_ms BIGINT DEFAULT 0,datasource_id VARCHAR(64),query_hash VARCHAR(64),INDEX idx_marel_rule(rule_id,started_at),INDEX idx_marel_status(status)) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4`,
	`CREATE TABLE IF NOT EXISTS monitor_alert_events_current (id VARCHAR(64) PRIMARY KEY,rule_id VARCHAR(64),rule_version INT,event_key VARCHAR(255),name VARCHAR(255),severity VARCHAR(32),status VARCHAR(32),datasource_id VARCHAR(64),target_id VARCHAR(64),target_ident VARCHAR(255),labels JSON,annotations JSON,value VARCHAR(255),fingerprint VARCHAR(128),count INT DEFAULT 1,first_seen DATETIME NOT NULL,last_seen DATETIME NOT NULL,ack_by VARCHAR(128),assignee VARCHAR(128),resolution TEXT,archived_at DATETIME NULL,resolved_at DATETIME NULL,created_at DATETIME NOT NULL,updated_at DATETIME NOT NULL,UNIQUE KEY uniq_maec_fp(fingerprint),INDEX idx_maec_status(status),INDEX idx_maec_rule(rule_id),INDEX idx_maec_last(last_seen)) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4`,
	`CREATE TABLE IF NOT EXISTS monitor_alert_events_history (id VARCHAR(64) PRIMARY KEY,rule_id VARCHAR(64),rule_version INT,event_key VARCHAR(255),name VARCHAR(255),severity VARCHAR(32),status VARCHAR(32),datasource_id VARCHAR(64),target_id VARCHAR(64),target_ident VARCHAR(255),labels JSON,annotations JSON,value VARCHAR(255),fingerprint VARCHAR(128),count INT DEFAULT 1,first_seen DATETIME NOT NULL,last_seen DATETIME NOT NULL,ack_by VARCHAR(128),assignee VARCHAR(128),resolution TEXT,archived_at DATETIME NULL,resolved_at DATETIME NULL,created_at DATETIME NOT NULL,updated_at DATETIME NOT NULL,INDEX idx_maeh_status(status),INDEX idx_maeh_rule(rule_id),INDEX idx_maeh_last(last_seen)) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4`,
	`CREATE TABLE IF NOT EXISTS monitor_alert_event_actions (id VARCHAR(64) PRIMARY KEY,event_id VARCHAR(64) NOT NULL,action VARCHAR(32) NOT NULL,actor VARCHAR(128),` + "`from`" + ` VARCHAR(32),` + "`to`" + ` VARCHAR(32),reason TEXT,assignee VARCHAR(128),trace_id VARCHAR(128),created_at DATETIME NOT NULL,INDEX idx_maea_event(event_id),INDEX idx_maea_time(created_at)) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4`,
	// 工作流引擎表
	`CREATE TABLE IF NOT EXISTS workflows (id VARCHAR(64) PRIMARY KEY,name VARCHAR(255) NOT NULL,description TEXT,dsl LONGTEXT,builtin TINYINT(1) DEFAULT 0,created_at DATETIME NOT NULL,updated_at DATETIME NOT NULL,INDEX idx_wf_name(name)) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4`,
	`CREATE TABLE IF NOT EXISTS workflow_versions (id VARCHAR(96) PRIMARY KEY,workflow_id VARCHAR(64) NOT NULL,version INT NOT NULL,name VARCHAR(255) NOT NULL,description TEXT,dsl LONGTEXT,created_at DATETIME NOT NULL,UNIQUE KEY uniq_wfv_workflow_version(workflow_id,version),INDEX idx_wfv_workflow(workflow_id)) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4`,
	`CREATE TABLE IF NOT EXISTS workflow_runs (id VARCHAR(64) PRIMARY KEY,workflow_id VARCHAR(64) NOT NULL,status VARCHAR(32) NOT NULL,inputs LONGTEXT,outputs LONGTEXT,error_message TEXT,elapsed_ms BIGINT DEFAULT 0,created_at DATETIME NOT NULL,INDEX idx_wfr_wf(workflow_id),INDEX idx_wfr_time(created_at)) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4`,
	// 知识库文档管理表
	`CREATE TABLE IF NOT EXISTS knowledge_documents (id VARCHAR(64) PRIMARY KEY,title VARCHAR(500) NOT NULL,content LONGTEXT NOT NULL,doc_type VARCHAR(32) NOT NULL,file_type VARCHAR(32),file_name VARCHAR(500),file_size INT DEFAULT 0,category VARCHAR(64),tags VARCHAR(500),source_id VARCHAR(64),embedding_model VARCHAR(100),chunk_index INT DEFAULT 0,parent_id VARCHAR(64),created_at DATETIME NOT NULL,updated_at DATETIME NOT NULL,FULLTEXT INDEX idx_doc_content(title,content),INDEX idx_doc_type(doc_type),INDEX idx_doc_category(category),INDEX idx_doc_parent(parent_id)) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4`,
	`CREATE TABLE IF NOT EXISTS knowledge_search_events (id VARCHAR(64) PRIMARY KEY,query VARCHAR(500) NOT NULL,hit_count INT DEFAULT 0,top_score DOUBLE DEFAULT 0,engine VARCHAR(32),created_at DATETIME NOT NULL,INDEX idx_kse_created(created_at),INDEX idx_kse_query(query)) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4`,
	`CREATE TABLE IF NOT EXISTS knowledge_search_badcases (id VARCHAR(64) PRIMARY KEY,query VARCHAR(500) NOT NULL,doc_id VARCHAR(64) NOT NULL,reason TEXT,created_by VARCHAR(128),created_at DATETIME NOT NULL,INDEX idx_ksb_created(created_at),INDEX idx_ksb_query(query)) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4`,
	// Runbook 执行记录表
	"CREATE TABLE IF NOT EXISTS runbook_executions (id VARCHAR(64) PRIMARY KEY,runbook_id VARCHAR(64) NOT NULL,target_ip VARCHAR(64),executor VARCHAR(128),status ENUM('running','succeeded','failed','cancelled','manual') DEFAULT 'running',variables JSON,output TEXT,error_message TEXT,started_at DATETIME NOT NULL,finished_at DATETIME,INDEX idx_rb_exec(runbook_id)) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4",
	// 工作流节点执行记录表（断点续跑）
	`CREATE TABLE IF NOT EXISTS workflow_run_steps (id BIGINT AUTO_INCREMENT PRIMARY KEY,run_id VARCHAR(64) NOT NULL,node_id VARCHAR(128) NOT NULL,node_type VARCHAR(64) NOT NULL,status VARCHAR(16) NOT NULL,inputs JSON,outputs JSON,` + "`error`" + ` TEXT,started_at DATETIME,finished_at DATETIME,elapsed_ms BIGINT DEFAULT 0,retry_count INT DEFAULT 0,INDEX idx_wrs_run(run_id)) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4`,
}

var tolerantMigrationStatements = []string{
	`ALTER TABLE alerts ADD COLUMN annotations json`,
	`ALTER TABLE alerts ADD COLUMN business_id varchar(64)`,
	`ALTER TABLE alerts ADD COLUMN fingerprint varchar(128)`,
	`ALTER TABLE alerts ADD COLUMN count int DEFAULT 1`,
	`ALTER TABLE alerts ADD COLUMN first_seen datetime`,
	`ALTER TABLE alerts ADD COLUMN last_seen datetime`,
	`ALTER TABLE alerts ADD COLUMN ack_by varchar(128)`,
	`ALTER TABLE alerts ADD COLUMN assignee varchar(128)`,
	`ALTER TABLE alerts ADD COLUMN muted_until datetime NULL`,
	`ALTER TABLE alerts ADD COLUMN deleted_at datetime NULL`,
	`ALTER TABLE alerts ADD COLUMN test_batch_id varchar(128)`,
	`ALTER TABLE alerts ADD COLUMN audit_trace_id varchar(128)`,
	`ALTER TABLE alerts ADD COLUMN diagnose_record_id varchar(64)`,
	`ALTER TABLE alerts ADD COLUMN linked_business_id varchar(64)`,
	`ALTER TABLE alerts ADD COLUMN resolution text`,
	`ALTER TABLE alerts ADD COLUMN runbook_url text`,
	`ALTER TABLE alerts ADD COLUMN action_log json`,
	`ALTER TABLE alerts ADD COLUMN notification_trail json`,
	`ALTER TABLE alerts ADD INDEX idx_alert_fingerprint(fingerprint)`,
	`ALTER TABLE alerts ADD INDEX idx_alert_status(status)`,
	`ALTER TABLE topology_edges ADD COLUMN status varchar(32)`,
	`ALTER TABLE topology_edges ADD COLUMN latency_ms int`,
	`ALTER TABLE topology_edges ADD COLUMN error text`,
	`ALTER TABLE topology_businesses ADD COLUMN attributes json`,
	// Runbook 增强字段
	`ALTER TABLE runbooks ADD COLUMN version INT DEFAULT 1`,
	`ALTER TABLE runbooks ADD COLUMN severity VARCHAR(32) DEFAULT 'medium'`,
	`ALTER TABLE runbooks ADD COLUMN estimated_time VARCHAR(32) DEFAULT ''`,
	`ALTER TABLE runbooks ADD COLUMN prerequisites TEXT`,
	`ALTER TABLE runbooks ADD COLUMN rollback_steps TEXT`,
	`ALTER TABLE runbooks ADD COLUMN variables JSON`,
	`ALTER TABLE runbooks ADD COLUMN last_executed_at DATETIME`,
	`ALTER TABLE runbooks ADD COLUMN execution_count INT DEFAULT 0`,
	`ALTER TABLE runbooks ADD COLUMN success_rate DECIMAL(5,2) DEFAULT 0`,
	`ALTER TABLE audit_events ADD COLUMN integrity_hash VARCHAR(128) DEFAULT ''`,
	`ALTER TABLE audit_events ADD COLUMN operator VARCHAR(128) DEFAULT ''`,
	`ALTER TABLE audit_events ADD COLUMN description TEXT`,
	// v6: 变更事件表
	`CREATE TABLE IF NOT EXISTS change_events (
			id BIGINT AUTO_INCREMENT PRIMARY KEY,
			target_ip VARCHAR(64) NOT NULL,
			change_type VARCHAR(32) NOT NULL,
			title VARCHAR(512) NOT NULL,
			description TEXT,
			operator VARCHAR(128),
			source VARCHAR(64),
			started_at DATETIME NOT NULL,
			completed_at DATETIME,
			status VARCHAR(16) DEFAULT 'completed',
			metadata JSON,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			INDEX idx_change_ip_time (target_ip, started_at)
		) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4`,
	// v6: 工作流定时调度表
	`CREATE TABLE IF NOT EXISTS workflow_schedules (
			id BIGINT AUTO_INCREMENT PRIMARY KEY,
			workflow_name VARCHAR(128) NOT NULL,
			cron_expr VARCHAR(64) NOT NULL,
			inputs_json JSON,
			enabled BOOLEAN DEFAULT TRUE,
			last_run DATETIME,
			next_run DATETIME,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			updated_at DATETIME DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP
		) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4`,
	// v6: Prompt 模板表
	`CREATE TABLE IF NOT EXISTS prompt_templates (
			id BIGINT AUTO_INCREMENT PRIMARY KEY,
			name VARCHAR(128) NOT NULL UNIQUE,
			category VARCHAR(64),
			template TEXT NOT NULL,
			version INT NOT NULL DEFAULT 1,
			variables JSON,
			description TEXT,
			is_active BOOLEAN DEFAULT TRUE,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			updated_at DATETIME DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP
		) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4`,
	// v6: 诊断验证表
	`CREATE TABLE IF NOT EXISTS diagnosis_verifications (
			id BIGINT AUTO_INCREMENT PRIMARY KEY,
			diagnosis_id VARCHAR(64) NOT NULL,
			alert_resolution_rate FLOAT,
			recurrence_count INT DEFAULT 0,
			runbook_executed BOOLEAN DEFAULT FALSE,
			runbook_success BOOLEAN DEFAULT FALSE,
			score FLOAT,
			verified_at DATETIME,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			INDEX idx_dv_diagnosis (diagnosis_id)
		) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4`,
	// v6: 案例表扩展字段
	`ALTER TABLE diagnosis_cases ADD COLUMN diagnostic_path JSON`,
	`ALTER TABLE diagnosis_cases ADD COLUMN distinguishing_features TEXT`,
	`ALTER TABLE diagnosis_cases ADD COLUMN negative_findings JSON`,
	`ALTER TABLE diagnosis_cases ADD COLUMN avg_feedback_rating FLOAT DEFAULT 0`,
	`ALTER TABLE diagnosis_cases ADD COLUMN verification_status VARCHAR(16) DEFAULT 'auto'`,
	// v6: 工作流版本字段
	`ALTER TABLE workflows ADD COLUMN version INT DEFAULT 1`,
	`ALTER TABLE workflow_runs ADD COLUMN workflow_version INT DEFAULT 1`,
	`ALTER TABLE workflow_runs ADD COLUMN parent_run_id VARCHAR(64)`,
	`INSERT IGNORE INTO workflow_versions (id,workflow_id,version,name,description,dsl,created_at) SELECT CONCAT(id, ':v', COALESCE(version,1)),id,COALESCE(version,1),name,COALESCE(description,''),dsl,COALESCE(updated_at,created_at,NOW()) FROM workflows`,
}

func nullableTime(t *time.Time) any {
	if t == nil {
		return nil
	}
	return *t
}

// RedisClient returns the Redis client and whether it is available.
func RedisClient() (*redis.Client, bool) {
	return redisClient, redisOK
}

// NewID 生成包含时间戳和进程内递增序列的唯一 ID。
func NewID() string {
	now := time.Now().UnixNano()
	seq := atomic.AddUint64(&newIDSequence, 1)
	return strconv.FormatInt(now, newIDNumberBase) + newIDSeparator + strconv.FormatUint(seq, newIDNumberBase)
}
