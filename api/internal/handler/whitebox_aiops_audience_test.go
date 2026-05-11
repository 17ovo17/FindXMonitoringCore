package handler

import "testing"

func TestWhiteboxAIOpsAudienceSummaryAndHandoff(t *testing.T) {
	msg := runAIOpsDiagnosis("audience-whitebox", "10.0.1.21 CPU ?????????", nil, "oncall")
	if msg.Audience != "oncall" {
		t.Fatalf("expected oncall audience, got %s", msg.Audience)
	}
	if msg.SummaryCard.Problem == "" || msg.SummaryCard.NextStep == "" || msg.SummaryCard.Severity == "" {
		t.Fatalf("summaryCard should be populated: %+v", msg.SummaryCard)
	}
	if msg.HandoffNote.Summary == "" || len(msg.HandoffNote.OpenQuestions) == 0 || msg.HandoffNote.EscalationPolicy == "" {
		t.Fatalf("handoffNote should be populated for oncall: %+v", msg.HandoffNote)
	}
	foundSummary := false
	foundHandoff := false
	for _, action := range msg.SuggestedActions {
		if action.ID == "copy-summary" && action.Type == "command" && action.Command != "" {
			foundSummary = true
		}
		if action.ID == "copy-handoff" && action.Type == "command" && action.Command != "" {
			foundHandoff = true
		}
	}
	if !foundSummary || !foundHandoff {
		t.Fatalf("expected copy summary and handoff actions, got %+v", msg.SuggestedActions)
	}
}

func TestWhiteboxAIOpsAudienceAutoDetection(t *testing.T) {
	cases := map[string]string{
		"manager business impact": "manager",
		"show PromQL evidence":    "ops",
		"oncall escalation p1":    "oncall",
		"is service affected":     "user",
	}
	for input, want := range cases {
		if got := normalizeAIOpsAudience("", input, "diagnostic"); got != want {
			t.Fatalf("normalizeAIOpsAudience(%q)=%s want %s", input, got, want)
		}
	}
	if got := normalizeAIOpsAudience("", "\u7ed9\u9886\u5bfc\u770b\u4e1a\u52a1\u5f71\u54cd", "diagnostic"); got != "manager" {
		t.Fatalf("expected Chinese manager intent, got %s", got)
	}
	if got := normalizeAIOpsAudience("", "\u9700\u8981\u503c\u73ed\u4ea4\u63a5", "diagnostic"); got != "oncall" {
		t.Fatalf("expected Chinese oncall intent, got %s", got)
	}
}
