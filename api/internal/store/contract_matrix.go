package store

import (
	"errors"
	"sort"
	"strings"
	"time"
	"unicode"

	"ai-workbench-api/internal/model"
)

var ErrContractMatrixValidation = errors.New("contract matrix validation failed")

func SaveContractMatrixEntry(input model.ContractMatrixRegisterRequest) (model.ContractMatrixEntry, error) {
	now := time.Now()
	item, err := normalizeContractMatrixEntry(input, now)
	if err != nil {
		return model.ContractMatrixEntry{}, err
	}
	mu.Lock()
	contractMatrixEntries[item.ID] = &item
	mu.Unlock()
	return copyContractMatrixEntry(item), nil
}

func ListContractMatrixEntries(status, domain string) ([]model.ContractMatrixEntry, error) {
	if err := EnsureMonitoringContractMatrixSeeded(); err != nil {
		return nil, err
	}
	status = strings.TrimSpace(status)
	domain = strings.TrimSpace(domain)
	if status != "" && !model.IsContractMatrixStatus(status) {
		return nil, ErrContractMatrixValidation
	}
	mu.RLock()
	defer mu.RUnlock()
	out := make([]model.ContractMatrixEntry, 0, len(contractMatrixEntries))
	for _, item := range contractMatrixEntries {
		if status != "" && item.Status != status {
			continue
		}
		if domain != "" && item.Domain != domain {
			continue
		}
		out = append(out, copyContractMatrixEntry(*item))
	}
	sort.Slice(out, func(i, j int) bool {
		if out[i].Domain == out[j].Domain {
			return out[i].ID < out[j].ID
		}
		return out[i].Domain < out[j].Domain
	})
	return out, nil
}

func GetContractMatrixEntry(id string) (model.ContractMatrixEntry, bool, error) {
	if err := EnsureMonitoringContractMatrixSeeded(); err != nil {
		return model.ContractMatrixEntry{}, false, err
	}
	id = strings.TrimSpace(id)
	if id == "" || containsContractSensitiveText(id) {
		return model.ContractMatrixEntry{}, false, ErrContractMatrixValidation
	}
	mu.RLock()
	defer mu.RUnlock()
	item, ok := contractMatrixEntries[id]
	if !ok {
		return model.ContractMatrixEntry{}, false, nil
	}
	return copyContractMatrixEntry(*item), true, nil
}

func ResetContractMatrixForTest() {
	mu.Lock()
	defer mu.Unlock()
	contractMatrixEntries = map[string]*model.ContractMatrixEntry{}
}

func normalizeContractMatrixEntry(input model.ContractMatrixRegisterRequest, now time.Time) (model.ContractMatrixEntry, error) {
	status := strings.TrimSpace(input.Status)
	item := model.ContractMatrixEntry{
		ID:            normalizeContractGapID(input.ID, input.Domain, input.Capability),
		Capability:    cleanContractText(input.Capability, 120),
		Domain:        cleanContractText(input.Domain, 80),
		Status:        status,
		Handler:       cleanContractText(input.Handler, 160),
		Backend:       cleanContractText(input.Backend, 160),
		Datasource:    cleanContractText(input.Datasource, 160),
		Executor:      cleanContractText(input.Executor, 160),
		SourceRefs:    cleanContractList(input.SourceRefs),
		EvidenceRefs:  cleanContractList(input.EvidenceRefs),
		BlockedReason: cleanContractText(input.BlockedReason, 240),
		SafeToRetry:   input.SafeToRetry,
		Metadata:      cleanContractMetadata(input.Metadata),
		CreatedAt:     now,
		UpdatedAt:     now,
	}
	if item.Status == "" || !model.IsContractMatrixStatus(item.Status) {
		return model.ContractMatrixEntry{}, ErrContractMatrixValidation
	}
	if item.ID == "" || item.Capability == "" || item.Domain == "" {
		return model.ContractMatrixEntry{}, ErrContractMatrixValidation
	}
	if item.Status == model.ContractStatusReady && !contractEntryExecutable(item) {
		return model.ContractMatrixEntry{}, ErrContractMatrixValidation
	}
	if item.Status != model.ContractStatusReady {
		item.SafeToRetry = false
	}
	if item.Status == model.ContractStatusUnsafe {
		item.BlockedReason = firstContractText(item.BlockedReason, "unsafe contract input")
	}
	if item.BlockedReason == "" && item.Status != model.ContractStatusReady {
		item.BlockedReason = defaultContractBlockReason(item.Status)
	}
	return item, nil
}

func contractEntryExecutable(item model.ContractMatrixEntry) bool {
	return item.Handler != "" && item.Backend != "" && item.Datasource != "" && item.Executor != "" && len(item.EvidenceRefs) > 0
}

func normalizeContractGapID(id, domain, capability string) string {
	raw := strings.TrimSpace(id)
	if raw == "" {
		raw = strings.Join([]string{"FX", "CONTRACT", domain, capability}, "-")
	}
	raw = strings.ToUpper(strings.ReplaceAll(raw, "_", "-"))
	parts := strings.FieldsFunc(raw, func(r rune) bool {
		return !(unicode.IsLetter(r) || unicode.IsDigit(r))
	})
	if len(parts) == 0 {
		return ""
	}
	return "FX-CONTRACT-" + strings.Join(dropContractPrefix(parts), "-")
}

func dropContractPrefix(parts []string) []string {
	start := 0
	if len(parts) > 0 && parts[0] == "FX" {
		start = 1
	}
	if len(parts) > start && parts[start] == "CONTRACT" {
		start++
	}
	return parts[start:]
}

func cleanContractText(value string, max int) string {
	clean := strings.TrimSpace(removeContractControlRunes(value))
	if clean == "" || containsContractSensitiveText(clean) || containsContractFakeSuccessText(clean) {
		return ""
	}
	runes := []rune(clean)
	if len(runes) > max {
		return string(runes[:max])
	}
	return clean
}

func cleanContractList(values []string) []string {
	out := []string{}
	seen := map[string]bool{}
	for _, value := range values {
		clean := cleanContractText(value, 180)
		if clean != "" && !seen[clean] {
			seen[clean] = true
			out = append(out, clean)
		}
	}
	return out
}

func cleanContractMetadata(input map[string]string) map[string]string {
	out := map[string]string{}
	for key, value := range input {
		cleanKey := cleanContractText(key, 80)
		cleanValue := cleanContractText(value, 180)
		if cleanKey != "" && cleanValue != "" {
			out[cleanKey] = cleanValue
		}
	}
	return out
}

func containsContractSensitiveText(value string) bool {
	normalized := strings.NewReplacer("-", "_", " ", "_", ".", "_").Replace(strings.ToLower(value))
	for _, marker := range []string{"token", "password", "passwd", "cookie", "dsn", "private_key", "privatekey", "secret", "bearer", "api_key", "apikey", "access_key", "session"} {
		if strings.Contains(normalized, marker) {
			return true
		}
	}
	return false
}

func containsContractFakeSuccessText(value string) bool {
	normalized := strings.ReplaceAll(strings.ToLower(value), "_", "-")
	for _, marker := range []string{"queued", "running", "succeeded", "success", "applied", "rolled-back", "installed", "data-arrived"} {
		if strings.Contains(normalized, marker) {
			return true
		}
	}
	return false
}

func removeContractControlRunes(value string) string {
	return strings.Map(func(r rune) rune {
		if r < 32 || r == 127 {
			return -1
		}
		return r
	}, value)
}

func defaultContractBlockReason(status string) string {
	switch status {
	case model.ContractStatusMissingBackend:
		return "backend handler contract missing"
	case model.ContractStatusMissingDatasource:
		return "datasource contract missing"
	case model.ContractStatusMissingExecutor:
		return "executor contract missing"
	case model.ContractStatusUnsafe:
		return "unsafe contract input"
	default:
		return "contract is blocked"
	}
}

func firstContractText(values ...string) string {
	for _, value := range values {
		if clean := cleanContractText(value, 240); clean != "" {
			return clean
		}
	}
	return ""
}

func copyContractMatrixEntry(item model.ContractMatrixEntry) model.ContractMatrixEntry {
	item.SourceRefs = append([]string{}, item.SourceRefs...)
	item.EvidenceRefs = append([]string{}, item.EvidenceRefs...)
	if item.Metadata != nil {
		cp := map[string]string{}
		for key, value := range item.Metadata {
			cp[key] = value
		}
		item.Metadata = cp
	}
	return item
}
