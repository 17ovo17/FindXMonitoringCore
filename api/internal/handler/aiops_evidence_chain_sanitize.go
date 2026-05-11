package handler

import "strings"

func cleanEvidenceChainBlocker(value string) string {
	clean := strings.TrimSpace(removeControlRunes(value))
	if clean == "" {
		return evidenceChainBlockedByContract + ": evidence contract is not open"
	}
	if looksSensitiveEvidenceChainValue(clean) || containsForbiddenEvidenceChainState(clean) {
		return evidenceChainRedactedBlocker
	}
	const maxEvidenceChainBlockerLen = 160
	runes := []rune(clean)
	if len(runes) > maxEvidenceChainBlockerLen {
		clean = string(runes[:maxEvidenceChainBlockerLen])
	}
	return clean
}

func safeEvidenceChainID(value, category, sourceType, kind string) string {
	if hasEvidenceChainControlRune(value) {
		return redactedEvidenceChainID(category, sourceType, kind)
	}
	clean := strings.TrimSpace(removeControlRunes(value))
	const maxEvidenceChainIDLen = 96
	if clean == "" || len([]rune(clean)) > maxEvidenceChainIDLen || looksSensitiveEvidenceChainValue(clean) || containsForbiddenEvidenceChainState(clean) {
		return redactedEvidenceChainID(category, sourceType, kind)
	}
	return clean
}

func hasEvidenceChainControlRune(value string) bool {
	for _, r := range value {
		if r < ' ' || r == 0x7f {
			return true
		}
	}
	return false
}

func redactedEvidenceChainID(category, sourceType, kind string) string {
	parts := []string{}
	for _, value := range []string{category, sourceType, kind} {
		if part := safeEvidenceChainIDPart(value); part != "" {
			parts = append(parts, part)
		}
	}
	if len(parts) == 0 {
		return "redacted-evidence"
	}
	return "redacted-" + strings.Join(parts, "-")
}

func safeEvidenceChainIDPart(value string) string {
	clean := strings.TrimSpace(removeControlRunes(value))
	if clean == "" || looksSensitiveEvidenceChainValue(clean) || containsForbiddenEvidenceChainState(clean) {
		return ""
	}
	out := strings.Map(func(r rune) rune {
		if r >= 'a' && r <= 'z' || r >= '0' && r <= '9' || r == '-' || r == '_' {
			return r
		}
		if r >= 'A' && r <= 'Z' {
			return r + ('a' - 'A')
		}
		return -1
	}, clean)
	if runes := []rune(out); len(runes) > 32 {
		out = string(runes[:32])
	}
	return strings.Trim(out, "-_")
}

func safeEvidenceChainRefs(values []string) []string {
	out := []string{}
	for _, value := range values {
		clean := strings.TrimSpace(value)
		if clean != "" && !looksSensitiveEvidenceChainValue(clean) {
			out = append(out, clean)
		}
	}
	return out
}

func safeEvidenceChainMetadata(input map[string]string) map[string]string {
	out := map[string]string{}
	for key, value := range input {
		if key == "" || looksSensitiveEvidenceChainValue(key) || looksSensitiveEvidenceChainValue(value) || isForbiddenEvidenceChainValue(value) {
			continue
		}
		out[strings.TrimSpace(key)] = strings.TrimSpace(value)
	}
	return out
}

func looksSensitiveEvidenceChainValue(value string) bool {
	return looksSensitive(value) || looksSensitiveReferenceValue(value)
}

func isForbiddenEvidenceChainValue(value string) bool {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case "queued", "running", "succeeded", "success", "applied", "rolled-back", "rolled_back":
		return true
	default:
		return false
	}
}

func containsForbiddenEvidenceChainState(value string) bool {
	clean := strings.ToLower(strings.TrimSpace(value))
	for _, marker := range []string{"queued", "running", "succeeded", "success", "applied", "rolled-back", "rolled_back"} {
		if strings.Contains(clean, marker) {
			return true
		}
	}
	return false
}
