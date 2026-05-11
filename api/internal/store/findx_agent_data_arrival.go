package store

import (
	"strings"
	"time"

	"ai-workbench-api/internal/model"
)

func SaveFindXAgentDataArrivalEvidence(input model.FindXAgentDataArrivalEvidence) (model.FindXAgentDataArrivalEvidence, error) {
	item := normalizeFindXAgentDataArrivalEvidence(input, time.Now())
	if mysqlOK {
		if err := persistFindXAgentDataArrivalEvidence(item); err != nil {
			return copyFindXAgentDataArrivalEvidence(item), err
		}
	}
	cp := copyFindXAgentDataArrivalEvidence(item)
	mu.Lock()
	findxAgentDataArrivalEvidence[item.ID] = &cp
	mu.Unlock()
	return copyFindXAgentDataArrivalEvidence(item), nil
}

func DataArrivalEvidenceSnapshot() map[string]model.FindXAgentDataArrival {
	items, err := ListFindXAgentDataArrivalEvidence()
	if err != nil {
		return map[string]model.FindXAgentDataArrival{}
	}
	out := map[string]model.FindXAgentDataArrival{}
	for _, item := range items {
		kind := strings.TrimSpace(item.Kind)
		if kind == "" {
			continue
		}
		current := out[kind]
		current.Kind = kind
		current.EvidenceCount++
		if item.Status == model.FindXAgentDataArrivalStatusReported {
			current.Status = model.FindXAgentDataArrivalStatusReported
			current.Blocker = ""
		} else if current.Status == "" {
			current.Status = item.Status
			current.Blocker = item.Blocker
		}
		if item.UpdatedAt.After(current.LastSeen) {
			current.LastSeen = item.UpdatedAt
		}
		out[kind] = current
	}
	return out
}

func normalizeFindXAgentDataArrivalEvidence(item model.FindXAgentDataArrivalEvidence, now time.Time) model.FindXAgentDataArrivalEvidence {
	if item.ID == "" {
		item.ID = NewID()
	}
	item.Kind = strings.TrimSpace(item.Kind)
	item.AgentID = strings.TrimSpace(item.AgentID)
	item.TargetID = strings.TrimSpace(item.TargetID)
	item.Status = normalizeDataArrivalStatus(item.Status)
	item.Blocker = strings.TrimSpace(item.Blocker)
	if !model.IsFindXAgentDataArrivalKind(item.Kind) {
		item.Status = model.FindXAgentDataArrivalStatusBlocked
		if item.Blocker == "" {
			item.Blocker = "BLOCKED_BY_CONTRACT: unsupported data arrival kind"
		}
	} else if !model.IsFindXAgentReceiverBackedDataArrivalKind(item.Kind) {
		item.Status = model.FindXAgentDataArrivalStatusBlocked
		if item.Blocker == "" {
			item.Blocker = "BLOCKED_BY_CONTRACT: receiver evidence contract is not open for this data arrival kind"
		}
	}
	item.EvidenceRefs = cleanStringList(item.EvidenceRefs)
	item.Metadata = sanitizeLifecycleMetadata(item.Metadata)
	if item.CreatedAt.IsZero() {
		item.CreatedAt = now
	}
	item.UpdatedAt = now
	return item
}

func normalizeDataArrivalStatus(status string) string {
	switch strings.TrimSpace(status) {
	case model.FindXAgentDataArrivalStatusReported,
		model.FindXAgentDataArrivalStatusBlocked,
		model.FindXAgentDataArrivalStatusError:
		return strings.TrimSpace(status)
	default:
		return model.FindXAgentDataArrivalStatusBlocked
	}
}

func persistFindXAgentDataArrivalEvidence(item model.FindXAgentDataArrivalEvidence) error {
	evidenceRefs, metadata := lifecycleJSON(item.EvidenceRefs), lifecycleJSON(item.Metadata)
	_, err := db.Exec(`REPLACE INTO findx_agent_data_arrival_evidence (id,kind,agent_id,target_id,status,blocker,evidence_refs,metadata,created_at,updated_at) VALUES (?,?,?,?,?,?,?,?,?,?)`,
		item.ID, item.Kind, item.AgentID, item.TargetID, item.Status, item.Blocker, evidenceRefs, metadata, item.CreatedAt, item.UpdatedAt)
	return err
}
