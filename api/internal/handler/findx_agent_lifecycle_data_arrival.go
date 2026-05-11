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
		if current.LastSeen.After(items[i].LastSeen) {
			items[i].LastSeen = current.LastSeen
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
			Kind:          current.Kind,
			Name:          receiverBackedDataArrivalName(current.Kind),
			Status:        current.Status,
			LastSeen:      current.LastSeen,
			Blocker:       current.Blocker,
			EvidenceCount: current.EvidenceCount,
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
