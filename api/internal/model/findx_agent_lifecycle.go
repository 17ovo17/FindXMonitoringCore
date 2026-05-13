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
	FindXAgentExecutionStatePlanned           = "planned"
	FindXAgentExecutionStatePreflightFailed   = "preflight_failed"
	FindXAgentExecutionStateBlockedByContract = "blocked_by_contract"
	FindXAgentExecutionStateDispatching       = "dispatching"
	FindXAgentExecutionStateRunning           = "running"
	FindXAgentExecutionStateReceiptPending    = "receipt_pending"
	FindXAgentExecutionStateServiceRegistered = "service_registered"
	FindXAgentExecutionStateHeartbeatSeen     = "heartbeat_seen"
	FindXAgentExecutionStateDataArrivalSeen   = "data_arrival_seen"
	FindXAgentExecutionStateFailed            = "failed"
	FindXAgentExecutionStateRolledBack        = "rolled_back"
	FindXAgentExecutionStateUninstalled       = "uninstalled"
)

type FindXAgentExecutionStateMachine struct {
	CurrentState  string   `json:"current_state"`
	AllowedStates []string `json:"allowed_states"`
	Terminal      bool     `json:"terminal"`
	SafeToRetry   bool     `json:"safe_to_retry"`
	Blocker       string   `json:"blocker"`
}

type FindXAgentReceiptContract struct {
	ID                 string   `json:"id"`
	Scope              string   `json:"scope"`
	Transport          string   `json:"transport"`
	Runner             string   `json:"runner"`
	RequiredReceipts   []string `json:"required_receipts"`
	MissingContracts   []string `json:"missing_contracts"`
	CredentialRequired bool     `json:"credential_required"`
	CredentialProvided bool     `json:"credential_provided"`
	Status             string   `json:"status"`
	Blocker            string   `json:"blocker"`
}

type FindXAgentReceiptContractMatrixRow struct {
	Scope             string   `json:"scope"`
	Platform          string   `json:"platform"`
	ExecutionSurface  string   `json:"execution_surface"`
	RequiredContracts []string `json:"required_contracts"`
	MissingContracts  []string `json:"missing_contracts"`
	Status            string   `json:"status"`
	Blocker           string   `json:"blocker"`
}

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
		FindXAgentExecutionStatusUninstallVerified,
		FindXAgentExecutionStatePlanned,
		FindXAgentExecutionStatePreflightFailed,
		FindXAgentExecutionStateBlockedByContract,
		FindXAgentExecutionStateDispatching,
		FindXAgentExecutionStateReceiptPending,
		FindXAgentExecutionStateServiceRegistered,
		FindXAgentExecutionStateHeartbeatSeen,
		FindXAgentExecutionStateDataArrivalSeen,
		FindXAgentExecutionStateRolledBack,
		FindXAgentExecutionStateUninstalled:
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
	PluginID              string                          `json:"plugin_id"`
	PluginVersion         string                          `json:"plugin_version"`
	ConfigFormat          string                          `json:"config_format"`
	ConfigSnippetRef      string                          `json:"config_snippet_ref"`
	ProviderModes         []string                        `json:"provider_modes"`
	ReloadStrategy        string                          `json:"reload_strategy"`
	RestartStrategy       string                          `json:"restart_strategy"`
	RemoteMutation        bool                            `json:"remote_mutation"`
	RemoteMutationStatus  string                          `json:"remote_mutation_status"`
	RolloutMetadata       []string                        `json:"rollout_metadata"`
	CredentialRefRequired bool                            `json:"credential_ref_required"`
	AuditEvent            string                          `json:"audit_event"`
	SourceEvidence        []string                        `json:"source_evidence"`
	PluginSourceMap       []FindXAgentPluginSourceSpec    `json:"plugin_source_map"`
	PlatformMatrix        []FindXAgentPluginPlatformSpec  `json:"platform_matrix"`
	SecurityProfile       FindXAgentPluginSecurityProfile `json:"security_profile"`
	Blockers              []string                        `json:"blockers,omitempty"`
}

type FindXAgentPluginSourceSpec struct {
	PluginID                 string   `json:"plugin_id"`
	PluginCategory           string   `json:"plugin_category"`
	SourceDirectories        []string `json:"source_directories"`
	ConfigPaths              []string `json:"config_paths"`
	ConfigFormat             string   `json:"config_format"`
	SupportedPlatforms       []string `json:"supported_platforms"`
	SecurityLevel            string   `json:"security_level"`
	UnsafePlugin             bool     `json:"unsafe_plugin"`
	UnsafePluginPolicyRef    string   `json:"unsafe_plugin_policy_ref,omitempty"`
	RemoteMutationStatus     string   `json:"remote_mutation_status"`
	Blockers                 []string `json:"blockers,omitempty"`
	SourceEvidence           []string `json:"source_evidence"`
	SourceEvidenceSummaryRef string   `json:"source_evidence_summary_ref,omitempty"`
}

type FindXAgentPluginPlatformSpec struct {
	Platform            string   `json:"platform"`
	ConfigPath          string   `json:"config_path"`
	ConfigFormat        string   `json:"config_format"`
	ReloadSupport       string   `json:"reload_support"`
	ReloadReceiptRef    string   `json:"reload_receipt_ref,omitempty"`
	RestartReceiptRef   string   `json:"restart_receipt_ref,omitempty"`
	Selectors           []string `json:"selectors,omitempty"`
	ReceiptRefs         []string `json:"receipt_refs,omitempty"`
	ReceiptRequirements []string `json:"receipt_requirements,omitempty"`
	Status              string   `json:"status"`
	Blockers            []string `json:"blockers,omitempty"`
}

type FindXAgentPluginSecurityProfile struct {
	SecurityLevel         string   `json:"security_level"`
	UnsafePluginPolicyRef string   `json:"unsafe_plugin_policy_ref,omitempty"`
	UnsafePluginIDs       []string `json:"unsafe_plugin_ids,omitempty"`
	BlockedPluginIDs      []string `json:"blocked_plugin_ids,omitempty"`
	Blockers              []string `json:"blockers,omitempty"`
	EvidenceRefs          []string `json:"evidence_refs,omitempty"`
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
	Kind            string    `json:"kind"`
	Name            string    `json:"name"`
	Status          string    `json:"status"`
	AgentCount      int       `json:"agent_count"`
	SourceAgent     string    `json:"source_agent,omitempty"`
	PackageVersion  string    `json:"package_version,omitempty"`
	ConfigVersion   string    `json:"config_version,omitempty"`
	FirstSeen       time.Time `json:"first_seen_at,omitempty"`
	LastSeen        time.Time `json:"last_seen,omitempty"`
	LastSeenAt      time.Time `json:"last_seen_at,omitempty"`
	SampleEvidence  string    `json:"sample_evidence,omitempty"`
	BackendReceiver string    `json:"backend_receiver,omitempty"`
	RelatedIDs      []string  `json:"related_ids,omitempty"`
	Blocker         string    `json:"blocker,omitempty"`
	EvidenceCount   int       `json:"evidence_count"`
}
