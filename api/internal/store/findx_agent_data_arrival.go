package store

import (
	"sort"
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
	sort.SliceStable(items, func(i, j int) bool {
		if !items[i].UpdatedAt.Equal(items[j].UpdatedAt) {
			return items[i].UpdatedAt.Before(items[j].UpdatedAt)
		}
		if !items[i].CreatedAt.Equal(items[j].CreatedAt) {
			return items[i].CreatedAt.Before(items[j].CreatedAt)
		}
		return items[i].ID < items[j].ID
	})
	out := map[string]model.FindXAgentDataArrival{}
	for _, item := range items {
		kind := strings.TrimSpace(item.Kind)
		if kind == "" {
			continue
		}
		current := out[kind]
		current.Kind = kind
		current.EvidenceCount++
		item = dataArrivalEvidenceSnapshotItem(item)
		current = fillDataArrivalEvidenceSummary(current, item)
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

func dataArrivalEvidenceSnapshotItem(item model.FindXAgentDataArrivalEvidence) model.FindXAgentDataArrivalEvidence {
	item.Kind = strings.TrimSpace(item.Kind)
	item.Status = normalizeDataArrivalStatus(item.Status)
	item.EvidenceRefs = cleanStringList(item.EvidenceRefs)
	if item.Status != model.FindXAgentDataArrivalStatusReported {
		return item
	}
	if !model.IsFindXAgentReceiverBackedDataArrivalKind(item.Kind) || dataArrivalReceiverRef(item) == "" {
		item.Status = model.FindXAgentDataArrivalStatusBlocked
		if strings.TrimSpace(item.Blocker) == "" {
			item.Blocker = "PENDING: receiver evidence ref is required for reported data arrival"
		}
	}
	return item
}

func fillDataArrivalEvidenceSummary(current model.FindXAgentDataArrival, item model.FindXAgentDataArrivalEvidence) model.FindXAgentDataArrival {
	if current.FirstSeen.IsZero() || item.CreatedAt.Before(current.FirstSeen) {
		current.FirstSeen = item.CreatedAt
	}
	if item.UpdatedAt.After(current.LastSeen) || item.UpdatedAt.Equal(current.LastSeen) {
		current.LastSeen = item.UpdatedAt
		current.LastSeenAt = item.UpdatedAt
		current.SourceAgent = item.AgentID
		current.SampleEvidence = firstLifecycleValue(item.EvidenceRefs...)
		current.BackendReceiver = dataArrivalReceiverRef(item)
		current.PackageVersion = dataArrivalMetadataValue(item.Metadata, "package_version", "agent_version", "version")
		current.ConfigVersion = dataArrivalMetadataValue(item.Metadata, "config_version")
		current.RelatedIDs = dataArrivalRelatedIDs(item.Metadata)
	}
	return current
}

func dataArrivalMetadataValue(metadata map[string]string, keys ...string) string {
	for _, key := range keys {
		if value := strings.TrimSpace(metadata[key]); value != "" {
			return value
		}
	}
	return ""
}

func dataArrivalRelatedIDs(metadata map[string]string) []string {
	return cleanStringList([]string{
		metadata["related_trace_id"],
		metadata["related_log_id"],
		metadata["related_metric_id"],
		metadata["related_metric_ids"],
		metadata["trace_id"],
		metadata["log_id"],
		metadata["metric_id"],
		metadata["metric_ids"],
		metadata["span_id"],
		metadata["sample_id"],
	})
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
	item.EvidenceRefs = cleanStringList(item.EvidenceRefs)
	if !model.IsFindXAgentDataArrivalKind(item.Kind) {
		item.Status = model.FindXAgentDataArrivalStatusBlocked
		if item.Blocker == "" {
			item.Blocker = "PENDING: unsupported data arrival kind"
		}
	} else if model.IsFindXAgentReceiverBackedDataArrivalKind(item.Kind) &&
		item.Status == model.FindXAgentDataArrivalStatusReported &&
		dataArrivalReceiverRef(item) == "" {
		item.Status = model.FindXAgentDataArrivalStatusBlocked
		if item.Blocker == "" {
			item.Blocker = "PENDING: receiver evidence ref is required for reported data arrival"
		}
	} else if !model.IsFindXAgentReceiverBackedDataArrivalKind(item.Kind) {
		item.Status = model.FindXAgentDataArrivalStatusBlocked
		if item.Blocker == "" {
			item.Blocker = "PENDING: receiver/evidence/data-arrival validator is not open for this signal"
		}
	}
	item.Metadata = sanitizeLifecycleMetadata(item.Metadata)
	if item.CreatedAt.IsZero() {
		item.CreatedAt = now
	}
	item.UpdatedAt = now
	item.Metadata = normalizeDataArrivalMetadata(item)
	return item
}

func normalizeDataArrivalMetadata(item model.FindXAgentDataArrivalEvidence) map[string]string {
	metadata := copyStringMap(item.Metadata)
	if stateLooksFakeCompletion(metadata["state"]) {
		delete(metadata, "state")
	}
	if stateLooksFakeCompletion(metadata["status"]) {
		delete(metadata, "status")
	}
	if item.AgentID != "" {
		metadata["source_agent"] = item.AgentID
	} else if metadata["source_agent"] == "" {
		metadata["source_agent"] = "unknown"
	}
	metadata["signal_type"] = item.Kind
	metadata["first_seen_at"] = item.CreatedAt.Format(time.RFC3339)
	metadata["last_seen_at"] = item.UpdatedAt.Format(time.RFC3339)
	metadata["backend_receiver"] = dataArrivalReceiverRef(item)
	if metadata["backend_receiver"] == "" {
		metadata["backend_receiver"] = "none"
	}
	metadata["sample_evidence"] = firstLifecycleValue(item.EvidenceRefs...)
	if metadata["sample_evidence"] == "" {
		metadata["sample_evidence"] = "none"
	}
	normalizeDataArrivalRelatedMetadata(metadata)
	if item.Status == model.FindXAgentDataArrivalStatusBlocked && item.Blocker != "" {
		metadata["blocked_reason"] = item.Blocker
	}
	return metadata
}

func stateLooksFakeCompletion(value string) bool {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case "queued", "running", "succeeded", "success", "applied", "rolled-back", "rolled_back", "installed", "data_arrived", "service_registered":
		return true
	default:
		return false
	}
}

func normalizeDataArrivalRelatedMetadata(metadata map[string]string) {
	if metadata["related_trace_id"] == "" {
		metadata["related_trace_id"] = metadata["trace_id"]
	}
	if metadata["related_log_id"] == "" {
		metadata["related_log_id"] = metadata["log_id"]
	}
	if metadata["related_metric_ids"] == "" {
		metadata["related_metric_ids"] = firstLifecycleValue(metadata["metric_ids"], metadata["metric_id"])
	}
}

func dataArrivalReceiverRef(item model.FindXAgentDataArrivalEvidence) string {
	for _, ref := range item.EvidenceRefs {
		if strings.HasPrefix(ref, "receiver:") {
			return ref
		}
	}
	return ""
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
