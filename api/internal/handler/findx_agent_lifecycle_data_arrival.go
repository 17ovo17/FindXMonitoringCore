package handler

import "ai-workbench-api/internal/model"

func mergeDataArrivalEvidence(items []model.FindXAgentDataArrival, evidence map[string]model.FindXAgentDataArrival) []model.FindXAgentDataArrival {
	seen := map[string]bool{}
	for i := range items {
		seen[items[i].Kind] = true
		if !isReceiverBackedDataArrivalKind(items[i].Kind) {
			continue
		}
		current, ok := evidence[items[i].Kind]
		if !ok || current.EvidenceCount == 0 {
			continue
		}
		items[i].EvidenceCount = current.EvidenceCount
		items[i].Status = current.Status
		items[i].FirstSeen = current.FirstSeen
		items[i].SourceAgent = current.SourceAgent
		items[i].PackageVersion = current.PackageVersion
		items[i].ConfigVersion = current.ConfigVersion
		items[i].SampleEvidence = current.SampleEvidence
		items[i].BackendReceiver = current.BackendReceiver
		items[i].RelatedIDs = append([]string{}, current.RelatedIDs...)
		if current.LastSeen.After(items[i].LastSeen) {
			items[i].LastSeen = current.LastSeen
			items[i].LastSeenAt = current.LastSeenAt
		}
		if current.Status == model.FindXAgentDataArrivalStatusReported {
			items[i].Blocker = ""
		} else {
			items[i].Blocker = current.Blocker
		}
	}
	for _, kind := range receiverBackedDataArrivalKinds() {
		current, ok := evidence[kind]
		if !ok || seen[kind] || current.EvidenceCount == 0 {
			continue
		}
		items = append(items, model.FindXAgentDataArrival{
			Kind:            current.Kind,
			Name:            receiverBackedDataArrivalName(current.Kind),
			Status:          current.Status,
			SourceAgent:     current.SourceAgent,
			PackageVersion:  current.PackageVersion,
			ConfigVersion:   current.ConfigVersion,
			FirstSeen:       current.FirstSeen,
			LastSeen:        current.LastSeen,
			LastSeenAt:      current.LastSeenAt,
			SampleEvidence:  current.SampleEvidence,
			BackendReceiver: current.BackendReceiver,
			RelatedIDs:      append([]string{}, current.RelatedIDs...),
			Blocker:         current.Blocker,
			EvidenceCount:   current.EvidenceCount,
		})
	}
	return items
}

func receiverBackedDataArrivalKinds() []string {
	return []string{
		model.FindXAgentDataArrivalKindHeartbeat,
		model.FindXAgentDataArrivalKindMetrics,
		model.FindXAgentDataArrivalKindLogs,
		model.FindXAgentDataArrivalKindTracing,
	}
}

func isReceiverBackedDataArrivalKind(kind string) bool {
	return model.IsFindXAgentReceiverBackedDataArrivalKind(kind)
}

func receiverBackedDataArrivalName(kind string) string {
	return dataArrivalKindName(kind)
}

func dataArrivalKindName(kind string) string {
	switch kind {
	case model.FindXAgentDataArrivalKindHeartbeat:
		return "心跳"
	case model.FindXAgentDataArrivalKindMetrics:
		return "指标"
	case model.FindXAgentDataArrivalKindLogs:
		return "日志"
	case model.FindXAgentDataArrivalKindTracing:
		return "链路"
	case model.FindXAgentDataArrivalKindProfiling:
		return "性能分析"
	case model.FindXAgentDataArrivalKindInspection:
		return "巡检"
	case model.FindXAgentDataArrivalKindRUM:
		return "前端体验"
	case model.FindXAgentDataArrivalKindGatewayTrace:
		return "网关链路"
	case model.FindXAgentDataArrivalKindTopology:
		return "拓扑"
	default:
		return kind
	}
}
