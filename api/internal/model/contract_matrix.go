package model

import "time"

const (
	ContractStatusReady             = "ready"
	ContractStatusBlocked           = "blocked"
	ContractStatusMissingBackend    = "missing_backend"
	ContractStatusMissingDatasource = "missing_datasource"
	ContractStatusMissingExecutor   = "missing_executor"
	ContractStatusUnsafe            = "unsafe"

	ContractBlockedByContractCode = "PENDING"
)

type ContractMatrixEntry struct {
	ID            string            `json:"id"`
	Capability    string            `json:"capability"`
	Domain        string            `json:"domain"`
	Status        string            `json:"status"`
	Handler       string            `json:"handler,omitempty"`
	Backend       string            `json:"backend,omitempty"`
	Datasource    string            `json:"datasource,omitempty"`
	Executor      string            `json:"executor,omitempty"`
	SourceRefs    []string          `json:"source_refs,omitempty"`
	EvidenceRefs  []string          `json:"evidence_refs,omitempty"`
	BlockedReason string            `json:"blocked_reason,omitempty"`
	SafeToRetry   bool              `json:"safe_to_retry"`
	Metadata      map[string]string `json:"metadata,omitempty"`
	CreatedAt     time.Time         `json:"created_at"`
	UpdatedAt     time.Time         `json:"updated_at"`
}

type ContractMatrixRegisterRequest struct {
	ID            string            `json:"id"`
	Capability    string            `json:"capability"`
	Domain        string            `json:"domain"`
	Status        string            `json:"status"`
	Handler       string            `json:"handler,omitempty"`
	Backend       string            `json:"backend,omitempty"`
	Datasource    string            `json:"datasource,omitempty"`
	Executor      string            `json:"executor,omitempty"`
	SourceRefs    []string          `json:"source_refs,omitempty"`
	EvidenceRefs  []string          `json:"evidence_refs,omitempty"`
	BlockedReason string            `json:"blocked_reason,omitempty"`
	SafeToRetry   bool              `json:"safe_to_retry"`
	Metadata      map[string]string `json:"metadata,omitempty"`
}

type ContractMatrixBlockedResponse struct {
	Code          string `json:"code"`
	Message       string `json:"message"`
	ContractGapID string `json:"contract_gap_id"`
	Status        string `json:"status"`
	SafeToRetry   bool   `json:"safe_to_retry"`
}

func IsContractMatrixStatus(status string) bool {
	switch status {
	case ContractStatusReady,
		ContractStatusBlocked,
		ContractStatusMissingBackend,
		ContractStatusMissingDatasource,
		ContractStatusMissingExecutor,
		ContractStatusUnsafe:
		return true
	default:
		return false
	}
}
