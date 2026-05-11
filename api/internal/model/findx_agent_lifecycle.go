package model

import (
	"time"
)

const (
	FindXAgentExecutionStatusBlocked           = "blocked"
	FindXAgentExecutionStatusQueued            = "queued"
	FindXAgentExecutionStatusRunning           = "running"
	FindXAgentExecutionStatusSucceeded         = "succeeded"
	FindXAgentExecutionStatusFailed            = "failed"
	FindXAgentExecutionStatusCancelled         = "cancelled"
	FindXAgentExecutionStatusRollbackRequired  = "rollback_required"
	FindXAgentExecutionStatusUninstallVerified = "uninstall_verified"
)

const (
	FindXAgentDataArrivalKindHeartbeat    = "heartbeat"
	FindXAgentDataArrivalKindMetrics      = "metrics"
	FindXAgentDataArrivalKindLogs         = "logs"
	FindXAgentDataArrivalKindTracing      = "tracing"
	FindXAgentDataArrivalKindProfiling    = "profiling"
	FindXAgentDataArrivalKindInspection   = "inspection"
	FindXAgentDataArrivalKindRUM          = "rum"
	FindXAgentDataArrivalKindGatewayTrace = "gateway_trace"
	FindXAgentDataArrivalKindTopology     = "topology"

	FindXAgentDataArrivalStatusReported = "reported"
	FindXAgentDataArrivalStatusBlocked  = "blocked"
	FindXAgentDataArrivalStatusError    = "error"
)

func IsFindXAgentDataArrivalKind(kind string) bool {
	switch kind {
	case FindXAgentDataArrivalKindHeartbeat,
		FindXAgentDataArrivalKindMetrics,
		FindXAgentDataArrivalKindLogs,
		FindXAgentDataArrivalKindTracing,
		FindXAgentDataArrivalKindProfiling,
		FindXAgentDataArrivalKindInspection,
		FindXAgentDataArrivalKindRUM,
		FindXAgentDataArrivalKindGatewayTrace,
		FindXAgentDataArrivalKindTopology:
		return true
	default:
		return false
	}
}

func IsFindXAgentReceiverBackedDataArrivalKind(kind string) bool {
	switch kind {
	case FindXAgentDataArrivalKindHeartbeat,
		FindXAgentDataArrivalKindMetrics,
		FindXAgentDataArrivalKindLogs,
		FindXAgentDataArrivalKindTracing:
		return true
	default:
		return false
	}
}

type FindXAgentPackage struct {
	ID                 string                                  `json:"id"`
	Name               string                                  `json:"name"`
	CapabilityDomain   string                                  `json:"capability_domain"`
	Runtime            string                                  `json:"runtime"`
	SupportedOS        []string                                `json:"supported_os"`
	PackageShape       string                                  `json:"package_shape"`
	TelemetryKinds     []string                                `json:"telemetry_kinds"`
	ConfigKeys         []string                                `json:"config_keys"`
	ConfigTemplateIDs  []string                                `json:"config_template_ids"`
	PluginConfig       *FindXAgentPluginConfigSpec             `json:"plugin_config,omitempty"`
	InstallEnvironment FindXAgentInstallEnvironment            `json:"install_environment"`
	EnvironmentMatrix  []FindXAgentPackageEnvironmentMatrixRow `json:"environment_matrix"`
	InstallMethods     []string                                `json:"install_methods"`
	SourceState        string                                  `json:"source_state"`
	Status             string                                  `json:"status"`
	Blockers           []string                                `json:"blockers,omitempty"`
	Signature          string                                  `json:"signature"`
	UpdatedAt          string                                  `json:"updated_at"`
}

type FindXAgentPackageEnvironmentMatrixRow struct {
	Platform            string `json:"platform"`
	InstallMethod       string `json:"install_method"`
	ToolEvidence        string `json:"tool_evidence"`
	SourceState         string `json:"source_state"`
	PackageState        string `json:"package_state"`
	Executor            string `json:"executor"`
	ServiceRegistration string `json:"service_registration"`
	ConfigDelivery      string `json:"config_delivery"`
	Uninstall           string `json:"uninstall"`
	Rollback            string `json:"rollback"`
	DataArrival         string `json:"data_arrival"`
	Blocker             string `json:"blocker"`
}

type FindXAgentInstallEnvironment struct {
	Status    string                          `json:"status"`
	Platforms []string                        `json:"platforms,omitempty"`
	Tools     []FindXAgentInstallToolEvidence `json:"tools,omitempty"`
	Blockers  []string                        `json:"blockers,omitempty"`
}

type FindXAgentInstallToolEvidence struct {
	Name        string `json:"name"`
	OS          string `json:"os,omitempty"`
	Arch        string `json:"arch,omitempty"`
	Required    bool   `json:"required"`
	Bundled     bool   `json:"bundled"`
	EvidenceRef string `json:"evidence_ref,omitempty"`
	Status      string `json:"status"`
	Blocker     string `json:"blocker,omitempty"`
}

type FindXAgentLifecyclePhase struct {
	Key     string `json:"key"`
	Name    string `json:"name"`
	Status  string `json:"status"`
	Blocker string `json:"blocker,omitempty"`
}

type FindXAgentInstallPlanRequest struct {
	PackageID     string            `json:"package_id"`
	OS            string            `json:"os"`
	Method        string            `json:"method"`
	Mode          string            `json:"mode,omitempty"`
	Execute       bool              `json:"execute,omitempty"`
	TargetID      string            `json:"target_id,omitempty"`
	TargetIDs     []string          `json:"target_ids,omitempty"`
	CredentialRef string            `json:"credential_ref,omitempty"`
	Metadata      map[string]string `json:"metadata,omitempty"`
}

type FindXAgentInstallPlan struct {
	ID                   string            `json:"id"`
	PackageID            string            `json:"package_id"`
	OS                   string            `json:"os"`
	Method               string            `json:"method"`
	TargetIDs            []string          `json:"target_ids"`
	Status               string            `json:"status"`
	Blocker              string            `json:"blocker,omitempty"`
	Audit                string            `json:"audit"`
	EvidenceRefs         []string          `json:"evidence_refs,omitempty"`
	Metadata             map[string]string `json:"metadata,omitempty"`
	CredentialRefPresent bool              `json:"credential_ref_present"`
	CreatedAt            time.Time         `json:"created_at"`
	UpdatedAt            time.Time         `json:"updated_at"`
}

type FindXAgentInstallExecutionStep struct {
	Name         string    `json:"name"`
	Status       string    `json:"status"`
	EvidenceRef  string    `json:"evidence_ref,omitempty"`
	ErrorSummary string    `json:"error_summary,omitempty"`
	UpdatedAt    time.Time `json:"updated_at"`
}

type FindXAgentInstallExecution struct {
	ID           string                           `json:"id"`
	PlanID       string                           `json:"plan_id"`
	TargetID     string                           `json:"target_id"`
	Runner       string                           `json:"runner"`
	Status       string                           `json:"status"`
	ExitCode     *int                             `json:"exit_code"`
	Steps        []FindXAgentInstallExecutionStep `json:"steps"`
	EvidenceRefs []string                         `json:"evidence_refs"`
	ErrorSummary string                           `json:"error_summary"`
	CreatedAt    time.Time                        `json:"created_at"`
	StartedAt    *time.Time                       `json:"started_at"`
	FinishedAt   *time.Time                       `json:"finished_at"`
	UpdatedAt    time.Time                        `json:"updated_at"`
}

func IsFindXAgentInstallExecutionStatus(status string) bool {
	switch status {
	case FindXAgentExecutionStatusBlocked,
		FindXAgentExecutionStatusQueued,
		FindXAgentExecutionStatusRunning,
		FindXAgentExecutionStatusSucceeded,
		FindXAgentExecutionStatusFailed,
		FindXAgentExecutionStatusCancelled,
		FindXAgentExecutionStatusRollbackRequired,
		FindXAgentExecutionStatusUninstallVerified:
		return true
	default:
		return false
	}
}

type FindXAgentConfigTemplate struct {
	ID                 string                      `json:"id"`
	Name               string                      `json:"name"`
	Scope              string                      `json:"scope"`
	ConfigKind         string                      `json:"config_kind"`
	Fields             []string                    `json:"fields"`
	TargetScopes       []string                    `json:"target_scopes"`
	RolloutScopes      []string                    `json:"rollout_scopes"`
	RolloutStrategies  []string                    `json:"rollout_strategies"`
	RemoteDistribution bool                        `json:"remote_distribution"`
	RollbackPolicy     string                      `json:"rollback_policy"`
	CapabilityPackages []string                    `json:"capability_packages"`
	PluginConfig       *FindXAgentPluginConfigSpec `json:"plugin_config,omitempty"`
	Status             string                      `json:"status"`
	Blocker            string                      `json:"blocker,omitempty"`
	UpdatedAt          string                      `json:"updated_at"`
}

type FindXAgentPluginConfigSpec struct {
	PluginID              string   `json:"plugin_id"`
	PluginVersion         string   `json:"plugin_version"`
	ConfigFormat          string   `json:"config_format"`
	ConfigSnippetRef      string   `json:"config_snippet_ref"`
	ProviderModes         []string `json:"provider_modes"`
	ReloadStrategy        string   `json:"reload_strategy"`
	RestartStrategy       string   `json:"restart_strategy"`
	RemoteMutation        bool     `json:"remote_mutation"`
	RemoteMutationStatus  string   `json:"remote_mutation_status"`
	RolloutMetadata       []string `json:"rollout_metadata"`
	CredentialRefRequired bool     `json:"credential_ref_required"`
	AuditEvent            string   `json:"audit_event"`
	SourceEvidence        []string `json:"source_evidence"`
}

type FindXAgentConfigRolloutRequest struct {
	TemplateID       string            `json:"template_id"`
	AgentIDs         []string          `json:"agent_ids,omitempty"`
	TargetIDs        []string          `json:"target_ids,omitempty"`
	ConfigVersion    string            `json:"config_version,omitempty"`
	ConfigSnippetRef string            `json:"config_snippet_ref,omitempty"`
	ConfigFormat     string            `json:"config_format,omitempty"`
	ProviderMode     string            `json:"provider_mode,omitempty"`
	PluginID         string            `json:"plugin_id,omitempty"`
	PluginVersion    string            `json:"plugin_version,omitempty"`
	ReloadStrategy   string            `json:"reload_strategy,omitempty"`
	RestartStrategy  string            `json:"restart_strategy,omitempty"`
	RolloutStrategy  string            `json:"rollout_strategy,omitempty"`
	RollbackRef      string            `json:"rollback_ref,omitempty"`
	AuditReason      string            `json:"audit_reason,omitempty"`
	ChangeTicket     string            `json:"change_ticket,omitempty"`
	RemoteMutation   bool              `json:"remote_mutation,omitempty"`
	CanaryPercent    int               `json:"canary_percent,omitempty"`
	CredentialRef    string            `json:"credential_ref,omitempty"`
	Metadata         map[string]string `json:"metadata,omitempty"`
}

type FindXAgentTaskRequest struct {
	Action        string            `json:"action"`
	AgentIDs      []string          `json:"agent_ids,omitempty"`
	TargetIDs     []string          `json:"target_ids,omitempty"`
	PackageID     string            `json:"package_id,omitempty"`
	ConfigVersion string            `json:"config_version,omitempty"`
	CredentialRef string            `json:"credential_ref,omitempty"`
	Metadata      map[string]string `json:"metadata,omitempty"`
}

type FindXAgentConfigRollout struct {
	ID                   string            `json:"id"`
	TemplateID           string            `json:"template_id"`
	AgentIDs             []string          `json:"agent_ids,omitempty"`
	TargetIDs            []string          `json:"target_ids,omitempty"`
	ConfigVersion        string            `json:"config_version,omitempty"`
	ConfigSnippetRef     string            `json:"config_snippet_ref,omitempty"`
	ConfigFormat         string            `json:"config_format,omitempty"`
	ProviderMode         string            `json:"provider_mode,omitempty"`
	PluginID             string            `json:"plugin_id,omitempty"`
	PluginVersion        string            `json:"plugin_version,omitempty"`
	ReloadStrategy       string            `json:"reload_strategy,omitempty"`
	RestartStrategy      string            `json:"restart_strategy,omitempty"`
	RolloutStrategy      string            `json:"rollout_strategy,omitempty"`
	RollbackRef          string            `json:"rollback_ref,omitempty"`
	AuditReason          string            `json:"audit_reason,omitempty"`
	ChangeTicket         string            `json:"change_ticket,omitempty"`
	RemoteMutation       bool              `json:"remote_mutation"`
	CanaryPercent        int               `json:"canary_percent,omitempty"`
	Status               string            `json:"status"`
	Blocker              string            `json:"blocker,omitempty"`
	Audit                string            `json:"audit"`
	EvidenceRefs         []string          `json:"evidence_refs,omitempty"`
	Metadata             map[string]string `json:"metadata,omitempty"`
	CredentialRefPresent bool              `json:"credential_ref_present"`
	CreatedAt            time.Time         `json:"created_at"`
	UpdatedAt            time.Time         `json:"updated_at"`
}

type FindXAgentExecutionTask struct {
	ID                   string            `json:"id"`
	Action               string            `json:"action"`
	AgentIDs             []string          `json:"agent_ids,omitempty"`
	TargetIDs            []string          `json:"target_ids,omitempty"`
	PackageID            string            `json:"package_id,omitempty"`
	ConfigVersion        string            `json:"config_version,omitempty"`
	Status               string            `json:"status"`
	Blocker              string            `json:"blocker,omitempty"`
	Audit                string            `json:"audit"`
	EvidenceRefs         []string          `json:"evidence_refs,omitempty"`
	Metadata             map[string]string `json:"metadata,omitempty"`
	CredentialRefPresent bool              `json:"credential_ref_present"`
	CreatedAt            time.Time         `json:"created_at"`
	UpdatedAt            time.Time         `json:"updated_at"`
}

type FindXAgentDataArrivalEvidence struct {
	ID           string            `json:"id"`
	Kind         string            `json:"kind"`
	AgentID      string            `json:"agent_id,omitempty"`
	TargetID     string            `json:"target_id,omitempty"`
	Status       string            `json:"status"`
	Blocker      string            `json:"blocker,omitempty"`
	EvidenceRefs []string          `json:"evidence_refs,omitempty"`
	Metadata     map[string]string `json:"metadata,omitempty"`
	CreatedAt    time.Time         `json:"created_at"`
	UpdatedAt    time.Time         `json:"updated_at"`
}

type FindXAgentArtifactMetadata struct {
	PackageID            string            `json:"package_id,omitempty"`
	TemplateID           string            `json:"template_id,omitempty"`
	PluginID             string            `json:"plugin_id,omitempty"`
	PluginVersion        string            `json:"plugin_version,omitempty"`
	ArtifactRef          string            `json:"artifact_ref,omitempty"`
	SignatureRef         string            `json:"signature_ref,omitempty"`
	CredentialRefPresent bool              `json:"credential_ref_present"`
	Metadata             map[string]string `json:"metadata,omitempty"`
}

type FindXAgentDataArrival struct {
	Kind          string    `json:"kind"`
	Name          string    `json:"name"`
	Status        string    `json:"status"`
	AgentCount    int       `json:"agent_count"`
	LastSeen      time.Time `json:"last_seen,omitempty"`
	Blocker       string    `json:"blocker,omitempty"`
	EvidenceCount int       `json:"evidence_count"`
}
