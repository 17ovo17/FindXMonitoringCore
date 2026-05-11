package handler

import (
	"sort"
	"strings"

	"ai-workbench-api/internal/model"
)

func sortEvidenceChainItems(items []aiopsEvidenceChainItem) {
	sort.SliceStable(items, func(i, j int) bool {
		return items[i].UpdatedAt.After(items[j].UpdatedAt)
	})
}

func summarizeEvidenceChainItems(items []aiopsEvidenceChainItem) aiopsEvidenceChainSummary {
	var summary aiopsEvidenceChainSummary
	summary.TotalItems = len(items)
	for _, item := range items {
		switch item.Status {
		case model.FindXAgentDataArrivalStatusReported:
			summary.ReportedItems++
		case model.FindXAgentDataArrivalStatusError:
			summary.ErrorItems++
		default:
			summary.BlockedItems++
			if item.SourceType == "contract_gap" {
				summary.MissingItems++
			}
		}
	}
	return summary
}

func summarizeEvidenceChainCategories(items []aiopsEvidenceChainItem) []aiopsEvidenceChainCategory {
	byKey := map[string]*aiopsEvidenceChainCategory{}
	for _, item := range items {
		row := byKey[item.Category]
		if row == nil {
			row = &aiopsEvidenceChainCategory{Key: item.Category}
			byKey[item.Category] = row
		}
		row.Total++
		incrementEvidenceChainCategory(row, item)
	}
	out := make([]aiopsEvidenceChainCategory, 0, len(byKey))
	for _, row := range byKey {
		row.LatestState = latestEvidenceChainState(*row)
		out = append(out, *row)
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Key < out[j].Key })
	return out
}

func incrementEvidenceChainCategory(row *aiopsEvidenceChainCategory, item aiopsEvidenceChainItem) {
	if item.SourceType == "contract_gap" {
		row.Missing++
	}
	switch item.Status {
	case model.FindXAgentDataArrivalStatusReported:
		row.Reported++
	case model.FindXAgentDataArrivalStatusError:
		row.Error++
	default:
		row.Blocked++
	}
}

func latestEvidenceChainState(row aiopsEvidenceChainCategory) string {
	if row.Error > 0 {
		return model.FindXAgentDataArrivalStatusError
	}
	if row.Blocked > 0 || row.Missing > 0 {
		return model.FindXAgentDataArrivalStatusBlocked
	}
	return model.FindXAgentDataArrivalStatusReported
}

func summarizeEvidenceChainBlockers(items []aiopsEvidenceChainItem) []aiopsEvidenceChainBlocker {
	byReason := map[string][]string{}
	for _, item := range items {
		if item.Status == model.FindXAgentDataArrivalStatusReported {
			continue
		}
		byReason[item.Blocker] = append(byReason[item.Blocker], item.ID)
	}
	out := make([]aiopsEvidenceChainBlocker, 0, len(byReason))
	for reason, ids := range byReason {
		sort.Strings(ids)
		out = append(out, aiopsEvidenceChainBlocker{Key: evidenceChainBlockerKey(reason), Reason: reason, Items: ids})
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Key < out[j].Key })
	return out
}

func evidenceChainBlockerKey(reason string) string {
	if strings.Contains(reason, evidenceChainBlockedByContract) {
		return evidenceChainBlockedByContract
	}
	return "EVIDENCE_CHAIN_ERROR"
}
