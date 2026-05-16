package store

import (
	"context"
	"database/sql"
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
	mu                             sync.RWMutex
	records                        []*model.DiagnoseRecord
	agents                         = map[string]*model.CatpawAgent{}
	credentials                    = map[string]*model.Credential{}
	alerts                         []*model.AlertRecord
	chatSessions                   = map[string]*model.ChatSession{}
	chatMessages                   = map[string][]model.ChatMessage{}
	topologyNodes                  = map[string]*model.TopologyNode{}
	topologyEdges                  = map[string]*model.TopologyEdge{}
	topologyBusinesses             = map[string]*model.TopologyBusiness{}
	diagnosisCases                 = map[string]*model.DiagnosisCase{}
	metricsMappings                = map[string]*model.MetricsMapping{}
	monitorTargets                 = map[string]*model.MonitorTarget{}
	findxAgents                    = map[string]*model.FindXAgent{}
	findxAgentInstallPlans         = map[string]*model.FindXAgentInstallPlan{}
	findxAgentInstallExecutions    = map[string]*model.FindXAgentInstallExecution{}
	findxAgentConfigRollouts       = map[string]*model.FindXAgentConfigRollout{}
	findxAgentExecutionTasks       = map[string]*model.FindXAgentExecutionTask{}
	findxAgentDataArrivalEvidence  = map[string]*model.FindXAgentDataArrivalEvidence{}
	findxAgentPluginAssignments    = map[string]*model.FindXAgentPluginAssignment{}
	findxAgentPluginTargetBindings = map[string]*model.FindXAgentPluginTargetBinding{}
	probeChecks                    = map[string]*model.ProbeCheck{}
	probeCheckResults              = map[string]*model.ProbeCheckResult{}
	probeStatusPages               = map[string]*model.ProbeStatusPage{}
	probeIncidents                 = map[string]*model.ProbeIncident{}
	probeNotificationBindings      = map[string]*model.ProbeNotificationBinding{}
	probeAlertBindings             = map[string]*model.ProbeAlertBinding{}
	contractMatrixEntries          = map[string]*model.ContractMatrixEntry{}
	monitorDashboards              = map[string]*model.MonitorDashboard{}
	monitorBuiltinComponents       = map[string]*model.MonitoringBuiltinComponent{}
	monitorBuiltinPayloads         = map[string]*model.MonitoringBuiltinPayload{}
	monitorSystemIntegrations      = map[string]*model.MonitoringSystemIntegration{}
	monitorAlertRules              = map[string]*model.MonitorAlertRule{}
	monitorRuleVersions            = map[string][]model.MonitorAlertRuleVersion{}
	monitorEventsCurrent           = map[string]*model.MonitorAlertEvent{}
	monitorEventsHistory           = map[string]*model.MonitorAlertEvent{}
	monitorEventActions            = map[string][]model.MonitorAlertAction{}
	notificationRules              = map[string]*model.NotificationRule{}
	notificationTemplates          = map[string]*model.NotificationTemplate{}
	runbooks                       = map[string]*model.Runbook{}
	runbookExecs                   = map[string]*model.RunbookExecution{}
	diagnosisFeedbacks             []*model.DiagnosisFeedback
	aiSettings                     = map[string]*model.AISetting{}
	knowledgeDocs                  = map[string]*model.KnowledgeDocument{}
	searchEvents                   []*model.KnowledgeSearchEvent
	searchBadcases                 []*model.KnowledgeSearchBadcase
	auditEvents                    []AuditEvent
	users                          = map[string]*model.User{}
	maxRecs                        = 100

	db          *sql.DB
	redisClient *redis.Client
	mysqlOK     bool
	redisOK     bool
	lastDBError string
)

const (
	newIDNumberBase             = 10
	newIDSeparator              = "-"
	defaultMySQLStartupAttempts = 12
	defaultRedisStartupAttempts = 12
	defaultMySQLStartupInterval = 2 * time.Second
	defaultRedisStartupInterval = 2 * time.Second
	defaultMySQLPingTimeout     = 3 * time.Second
	defaultRedisPingTimeout     = 2 * time.Second
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

// Init initializes MySQL, GORM, Redis, topology seed, and fallback snapshot.
func Init() {
	initMySQL()
	InitGormDB()
	initRedis()
	seedPlatformTopology()
	loadFallbackSnapshot()
	SeedOrgDefaults()
	InitAgentPackageStore()
	InitPluginConfigStore()
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

	pingTimeout := storageRetryDuration("mysql.ping_timeout", defaultMySQLPingTimeout)
	retry := storageStartupRetry("mysql.startup_retry_attempts", "mysql.startup_retry_interval", defaultMySQLStartupAttempts, defaultMySQLStartupInterval)
	if err := retry.run("mysql ping", func() error {
		ctx, cancel := context.WithTimeout(context.Background(), pingTimeout)
		defer cancel()
		return conn.PingContext(ctx)
	}); err != nil {
		lastDBError = err.Error()
		logrus.Warnf("mysql unavailable, using memory store: %v", err)
		return
	}
	db = conn
	if err := retry.run("mysql migrate", migrate); err != nil {
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
	pingTimeout := storageRetryDuration("redis.ping_timeout", defaultRedisPingTimeout)
	retry := storageStartupRetry("redis.startup_retry_attempts", "redis.startup_retry_interval", defaultRedisStartupAttempts, defaultRedisStartupInterval)
	if err := retry.run("redis ping", func() error {
		ctx, cancel := context.WithTimeout(context.Background(), pingTimeout)
		defer cancel()
		return client.Ping(ctx).Err()
	}); err != nil {
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

// NewID generates an ID with a timestamp and process-local sequence.
func NewID() string {
	now := time.Now().UnixNano()
	seq := atomic.AddUint64(&newIDSequence, 1)
	return strconv.FormatInt(now, newIDNumberBase) + newIDSeparator + strconv.FormatUint(seq, newIDNumberBase)
}
