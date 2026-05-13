package store

import (
	"errors"
	"sort"
	"strings"
	"time"

	"ai-workbench-api/internal/model"
)

var ErrProbeValidation = errors.New("probe validation failed")

type ProbeCheckFilter struct {
	Query   string
	Type    string
	Status  string
	Enabled *bool
}

func ListProbeChecks(filter ProbeCheckFilter) ([]model.ProbeCheck, error) {
	if GormOK() {
		query := GetDB().Model(&model.ProbeCheck{})
		if filter.Type != "" {
			query = query.Where("type = ?", filter.Type)
		}
		if filter.Status != "" {
			query = query.Where("status = ?", filter.Status)
		}
		if filter.Enabled != nil {
			query = query.Where("enabled = ?", *filter.Enabled)
		}
		if filter.Query != "" {
			like := "%" + filter.Query + "%"
			query = query.Where("name LIKE ? OR target LIKE ? OR url LIKE ?", like, like, like)
		}
		var items []model.ProbeCheck
		if err := query.Order("updated_at DESC").Limit(500).Find(&items).Error; err == nil {
			return hydrateProbeChecks(items), nil
		}
	}
	mu.RLock()
	defer mu.RUnlock()
	out := make([]model.ProbeCheck, 0, len(probeChecks))
	for _, item := range probeChecks {
		cp := copyProbeCheck(*item)
		if probeCheckMatches(cp, filter) {
			out = append(out, cp)
		}
	}
	sortProbeChecks(out)
	return hydrateProbeChecks(out), nil
}

func GetProbeCheck(id string) (model.ProbeCheck, bool, error) {
	id = strings.TrimSpace(id)
	if id == "" {
		return model.ProbeCheck{}, false, nil
	}
	if GormOK() {
		var item model.ProbeCheck
		err := GetDB().First(&item, "id = ?", id).Error
		if err == nil {
			items := hydrateProbeChecks([]model.ProbeCheck{item})
			return items[0], true, nil
		}
	}
	mu.RLock()
	defer mu.RUnlock()
	item, ok := probeChecks[id]
	if !ok {
		return model.ProbeCheck{}, false, nil
	}
	return hydrateProbeChecks([]model.ProbeCheck{copyProbeCheck(*item)})[0], true, nil
}

func SaveProbeCheck(input model.ProbeCheck) (model.ProbeCheck, error) {
	item, err := normalizeProbeCheck(input, time.Now())
	if err != nil {
		return model.ProbeCheck{}, err
	}
	if GormOK() {
		if err := GetDB().Save(&item).Error; err != nil {
			return model.ProbeCheck{}, err
		}
	}
	mu.Lock()
	cp := copyProbeCheck(item)
	probeChecks[item.ID] = &cp
	mu.Unlock()
	return hydrateProbeChecks([]model.ProbeCheck{item})[0], nil
}

func DeleteProbeCheck(id string) (bool, error) {
	id = strings.TrimSpace(id)
	if id == "" {
		return false, nil
	}
	found := false
	mu.Lock()
	if _, ok := probeChecks[id]; ok {
		delete(probeChecks, id)
		found = true
	}
	for bindingID, binding := range probeNotificationBindings {
		if binding.CheckID == id {
			delete(probeNotificationBindings, bindingID)
		}
	}
	for bindingID, binding := range probeAlertBindings {
		if binding.CheckID == id {
			delete(probeAlertBindings, bindingID)
		}
	}
	mu.Unlock()
	if GormOK() {
		tx := GetDB().Delete(&model.ProbeCheck{}, "id = ?", id)
		if tx.Error != nil {
			return false, tx.Error
		}
		if err := GetDB().Delete(&model.ProbeNotificationBinding{}, "check_id = ?", id).Error; err != nil {
			return false, err
		}
		if err := GetDB().Delete(&model.ProbeAlertBinding{}, "check_id = ?", id).Error; err != nil {
			return false, err
		}
		found = found || tx.RowsAffected > 0
	}
	return found, nil
}

func SetProbeCheckEnabled(id string, enabled bool) (model.ProbeCheck, bool, error) {
	item, ok, err := GetProbeCheck(id)
	if err != nil || !ok {
		return model.ProbeCheck{}, ok, err
	}
	item.Enabled = enabled
	if enabled && item.Status == model.ProbeStatusDisabled {
		item.Status = model.ProbeStatusUnknown
	}
	if !enabled {
		item.Status = model.ProbeStatusDisabled
	}
	out, err := SaveProbeCheck(item)
	return out, err == nil, err
}

func ListProbeStatusPages() ([]model.ProbeStatusPage, error) {
	if GormOK() {
		var items []model.ProbeStatusPage
		if err := GetDB().Order("updated_at DESC").Limit(100).Find(&items).Error; err == nil {
			if len(items) > 0 {
				return items, nil
			}
		}
	}
	mu.RLock()
	defer mu.RUnlock()
	out := make([]model.ProbeStatusPage, 0, len(probeStatusPages))
	for _, item := range probeStatusPages {
		out = append(out, copyProbeStatusPage(*item))
	}
	sort.Slice(out, func(i, j int) bool { return out[i].UpdatedAt.After(out[j].UpdatedAt) })
	if len(out) == 0 {
		out = append(out, defaultProbeStatusPage())
	}
	return out, nil
}

func GetProbeStatusPage(slugOrID string) (model.ProbeStatusPage, bool, error) {
	slugOrID = normalizeProbeSlug(slugOrID)
	if GormOK() {
		var item model.ProbeStatusPage
		err := GetDB().Where("slug = ? OR id = ?", slugOrID, slugOrID).First(&item).Error
		if err == nil {
			return item, true, nil
		}
	}
	mu.RLock()
	defer mu.RUnlock()
	for _, item := range probeStatusPages {
		if item.Slug == slugOrID || item.ID == slugOrID {
			return copyProbeStatusPage(*item), true, nil
		}
	}
	if slugOrID == "main" {
		return defaultProbeStatusPage(), true, nil
	}
	return model.ProbeStatusPage{}, false, nil
}

func SaveProbeStatusPage(input model.ProbeStatusPage) (model.ProbeStatusPage, error) {
	item := normalizeProbeStatusPage(input, time.Now())
	if GormOK() {
		if err := GetDB().Save(&item).Error; err != nil {
			return model.ProbeStatusPage{}, err
		}
	}
	mu.Lock()
	cp := copyProbeStatusPage(item)
	probeStatusPages[item.ID] = &cp
	mu.Unlock()
	return item, nil
}

func ListProbeIncidents(status string) ([]model.ProbeIncident, error) {
	if GormOK() {
		query := GetDB().Model(&model.ProbeIncident{})
		if status != "" {
			query = query.Where("status = ?", status)
		}
		var items []model.ProbeIncident
		if err := query.Order("started_at DESC").Limit(200).Find(&items).Error; err == nil {
			return items, nil
		}
	}
	mu.RLock()
	defer mu.RUnlock()
	out := make([]model.ProbeIncident, 0, len(probeIncidents))
	for _, item := range probeIncidents {
		if status == "" || item.Status == status {
			out = append(out, copyProbeIncident(*item))
		}
	}
	sort.Slice(out, func(i, j int) bool { return out[i].StartedAt.After(out[j].StartedAt) })
	return out, nil
}

func SaveProbeIncident(input model.ProbeIncident) (model.ProbeIncident, error) {
	item, err := normalizeProbeIncident(input, time.Now())
	if err != nil {
		return model.ProbeIncident{}, err
	}
	if GormOK() {
		if err := GetDB().Save(&item).Error; err != nil {
			return model.ProbeIncident{}, err
		}
	}
	mu.Lock()
	cp := copyProbeIncident(item)
	probeIncidents[item.ID] = &cp
	mu.Unlock()
	return item, nil
}

func DeleteProbeIncident(id string) (bool, error) {
	id = strings.TrimSpace(id)
	if id == "" {
		return false, nil
	}
	found := false
	mu.Lock()
	if _, ok := probeIncidents[id]; ok {
		delete(probeIncidents, id)
		found = true
	}
	mu.Unlock()
	if GormOK() {
		tx := GetDB().Delete(&model.ProbeIncident{}, "id = ?", id)
		if tx.Error != nil {
			return false, tx.Error
		}
		found = found || tx.RowsAffected > 0
	}
	return found, nil
}

func ListProbeNotificationBindings(checkID string) ([]model.ProbeNotificationBinding, error) {
	checkID = strings.TrimSpace(checkID)
	if GormOK() {
		var items []model.ProbeNotificationBinding
		if err := GetDB().Where("check_id = ?", checkID).Find(&items).Error; err == nil {
			return items, nil
		}
	}
	mu.RLock()
	defer mu.RUnlock()
	out := []model.ProbeNotificationBinding{}
	for _, item := range probeNotificationBindings {
		if item.CheckID == checkID {
			out = append(out, copyProbeNotificationBinding(*item))
		}
	}
	sort.Slice(out, func(i, j int) bool { return out[i].UpdatedAt.After(out[j].UpdatedAt) })
	return out, nil
}

func SaveProbeNotificationBindings(checkID string, bindings []model.ProbeNotificationBinding) ([]model.ProbeNotificationBinding, error) {
	checkID = strings.TrimSpace(checkID)
	now := time.Now()
	out := make([]model.ProbeNotificationBinding, 0, len(bindings))
	for _, input := range bindings {
		item := normalizeProbeNotificationBinding(checkID, input, now)
		out = append(out, item)
	}
	if GormOK() {
		if err := GetDB().Delete(&model.ProbeNotificationBinding{}, "check_id = ?", checkID).Error; err != nil {
			return nil, err
		}
		for _, item := range out {
			if err := GetDB().Save(&item).Error; err != nil {
				return nil, err
			}
		}
	}
	mu.Lock()
	for id, item := range probeNotificationBindings {
		if item.CheckID == checkID {
			delete(probeNotificationBindings, id)
		}
	}
	for _, item := range out {
		cp := copyProbeNotificationBinding(item)
		probeNotificationBindings[item.ID] = &cp
	}
	mu.Unlock()
	return out, nil
}

func ListProbeAlertBindings(checkID string) ([]model.ProbeAlertBinding, error) {
	checkID = strings.TrimSpace(checkID)
	if GormOK() {
		var items []model.ProbeAlertBinding
		if err := GetDB().Where("check_id = ?", checkID).Find(&items).Error; err == nil {
			return items, nil
		}
	}
	mu.RLock()
	defer mu.RUnlock()
	out := []model.ProbeAlertBinding{}
	for _, item := range probeAlertBindings {
		if item.CheckID == checkID {
			out = append(out, copyProbeAlertBinding(*item))
		}
	}
	sort.Slice(out, func(i, j int) bool { return out[i].UpdatedAt.After(out[j].UpdatedAt) })
	return out, nil
}

func SaveProbeAlertBindings(checkID string, bindings []model.ProbeAlertBinding) ([]model.ProbeAlertBinding, error) {
	checkID = strings.TrimSpace(checkID)
	now := time.Now()
	out := make([]model.ProbeAlertBinding, 0, len(bindings))
	for _, input := range bindings {
		item := normalizeProbeAlertBinding(checkID, input, now)
		out = append(out, item)
	}
	if GormOK() {
		if err := GetDB().Delete(&model.ProbeAlertBinding{}, "check_id = ?", checkID).Error; err != nil {
			return nil, err
		}
		for _, item := range out {
			if err := GetDB().Save(&item).Error; err != nil {
				return nil, err
			}
		}
	}
	mu.Lock()
	for id, item := range probeAlertBindings {
		if item.CheckID == checkID {
			delete(probeAlertBindings, id)
		}
	}
	for _, item := range out {
		cp := copyProbeAlertBinding(item)
		probeAlertBindings[item.ID] = &cp
	}
	mu.Unlock()
	return out, nil
}

func ListProbeCheckResults(checkID string, since time.Time) ([]model.ProbeCheckResult, error) {
	checkID = strings.TrimSpace(checkID)
	if GormOK() {
		query := GetDB().Where("check_id = ?", checkID)
		if !since.IsZero() {
			query = query.Where("checked_at >= ?", since)
		}
		var items []model.ProbeCheckResult
		if err := query.Order("checked_at DESC").Limit(3000).Find(&items).Error; err == nil {
			return items, nil
		}
	}
	mu.RLock()
	defer mu.RUnlock()
	out := []model.ProbeCheckResult{}
	for _, item := range probeCheckResults {
		if item.CheckID == checkID && (since.IsZero() || !item.CheckedAt.Before(since)) {
			out = append(out, copyProbeCheckResult(*item))
		}
	}
	sort.Slice(out, func(i, j int) bool { return out[i].CheckedAt.After(out[j].CheckedAt) })
	return out, nil
}

func SaveProbeCheckResult(input model.ProbeCheckResult) (model.ProbeCheckResult, error) {
	item := normalizeProbeCheckResult(input, time.Now())
	if GormOK() {
		if err := GetDB().Save(&item).Error; err != nil {
			return model.ProbeCheckResult{}, err
		}
	}
	mu.Lock()
	cp := copyProbeCheckResult(item)
	probeCheckResults[item.ID] = &cp
	mu.Unlock()
	return item, nil
}

func BuildProbeStatusPageView(slug string) (model.ProbeStatusPageView, bool, error) {
	page, ok, err := GetProbeStatusPage(slug)
	if err != nil || !ok {
		return model.ProbeStatusPageView{}, ok, err
	}
	checks, err := ListProbeChecks(ProbeCheckFilter{})
	if err != nil {
		return model.ProbeStatusPageView{}, false, err
	}
	incidents, err := ListProbeIncidents("")
	if err != nil {
		return model.ProbeStatusPageView{}, false, err
	}
	groups := buildProbeStatusGroups(page, checks)
	summary, status, reason := summarizeProbeStatus(groups, incidents)
	return model.ProbeStatusPageView{
		Status:       status,
		StatusReason: reason,
		UpdatedAt:    time.Now(),
		Page:         &page,
		Summary:      summary,
		Groups:       groups,
		Incidents:    incidents,
		Meta: map[string]any{
			"source":              "findx_probe_store",
			"run_evidence_policy": "90 天可用性、响应耗时和全局状态只基于真实 probe_check_results 聚合；无运行记录保持 unknown/no_data。",
		},
	}, true, nil
}

func ResetProbeStoreForTest() {
	mu.Lock()
	defer mu.Unlock()
	probeChecks = map[string]*model.ProbeCheck{}
	probeCheckResults = map[string]*model.ProbeCheckResult{}
	probeStatusPages = map[string]*model.ProbeStatusPage{}
	probeIncidents = map[string]*model.ProbeIncident{}
	probeNotificationBindings = map[string]*model.ProbeNotificationBinding{}
	probeAlertBindings = map[string]*model.ProbeAlertBinding{}
}

func buildProbeStatusGroups(page model.ProbeStatusPage, checks []model.ProbeCheck) []model.ProbeStatusGroupView {
	byID := make(map[string]model.ProbeCheck, len(checks))
	for _, check := range checks {
		byID[check.ID] = check
	}
	groups := page.Groups
	if len(groups) == 0 {
		groups = []model.ProbeStatusGroup{{ID: "core", Name: "核心平台", CheckIDs: []string{}}}
	}
	out := make([]model.ProbeStatusGroupView, 0, len(groups))
	used := map[string]bool{}
	for _, group := range groups {
		view := model.ProbeStatusGroupView{ID: group.ID, Name: group.Name, Checks: []model.ProbeCheck{}}
		for _, id := range group.CheckIDs {
			if check, ok := byID[id]; ok {
				view.Checks = append(view.Checks, check)
				used[id] = true
			}
		}
		out = append(out, view)
	}
	if len(out) == 1 && len(out[0].Checks) == 0 {
		for _, check := range checks {
			out[0].Checks = append(out[0].Checks, check)
			used[check.ID] = true
		}
	}
	ungrouped := []model.ProbeCheck{}
	for _, check := range checks {
		if !used[check.ID] {
			ungrouped = append(ungrouped, check)
		}
	}
	if len(ungrouped) > 0 {
		out = append(out, model.ProbeStatusGroupView{ID: "ungrouped", Name: "未分组拨测", Checks: ungrouped})
	}
	return out
}

func summarizeProbeStatus(groups []model.ProbeStatusGroupView, incidents []model.ProbeIncident) (model.ProbeStatusSummary, string, string) {
	summary := model.ProbeStatusSummary{MissingRunEvidenceNote: "暂无真实拨测运行记录，不能计算 90 天可用性或响应耗时。"}
	totalResults, upResults, responseCount, responseSum := 0, 0, 0, 0
	hasDown, hasDegraded := false, false
	for _, group := range groups {
		for _, check := range group.Checks {
			summary.TotalChecks++
			if check.Enabled {
				summary.RunningChecks++
			}
			results, _ := ListProbeCheckResults(check.ID, time.Now().AddDate(0, 0, -90))
			for _, result := range results {
				totalResults++
				if result.Status == model.ProbeStatusUp {
					upResults++
				}
				if result.Status == model.ProbeStatusDown {
					hasDown = true
				}
				if result.Status == model.ProbeStatusDegraded {
					hasDegraded = true
				}
				if result.ResponseTimeMs > 0 && result.CheckedAt.After(time.Now().AddDate(0, 0, -30)) {
					responseCount++
					responseSum += result.ResponseTimeMs
				}
			}
		}
	}
	cutoff := time.Now().AddDate(0, 0, -90)
	for _, incident := range incidents {
		if !incident.StartedAt.Before(cutoff) {
			summary.IncidentCount90d++
		}
		if incident.Status != model.ProbeIncidentStatusResolved {
			hasDegraded = true
		}
	}
	if totalResults == 0 {
		return summary, model.ProbeStatusUnknown, summary.MissingRunEvidenceNote
	}
	summary.HasRunEvidence = true
	summary.MissingRunEvidenceNote = ""
	uptime := float64(upResults) / float64(totalResults) * 100
	summary.Uptime90d = &uptime
	if responseCount > 0 {
		avg := responseSum / responseCount
		summary.AverageResponse30dMs = &avg
	}
	switch {
	case hasDown:
		return summary, model.ProbeStatusDown, "存在真实拨测失败记录。"
	case hasDegraded:
		return summary, model.ProbeStatusDegraded, "存在未关闭事故或真实降级记录。"
	default:
		return summary, model.ProbeStatusUp, "所有启用拨测最近运行记录均正常。"
	}
}

func normalizeProbeCheck(input model.ProbeCheck, now time.Time) (model.ProbeCheck, error) {
	item := input
	item.ID = strings.TrimSpace(item.ID)
	if item.ID == "" {
		item.ID = "probe-" + NewID()
	}
	item.Name = compactStoreText(item.Name, 160)
	item.Type = strings.ToLower(strings.TrimSpace(item.Type))
	item.URL = compactStoreText(item.URL, 1024)
	item.Target = compactStoreText(item.Target, 512)
	item.BusinessGroup = compactStoreText(item.BusinessGroup, 160)
	if item.Name == "" || !validProbeCheckType(item.Type) {
		return model.ProbeCheck{}, ErrProbeValidation
	}
	if item.Type == model.ProbeCheckTypeHTTP && item.URL == "" {
		return model.ProbeCheck{}, ErrProbeValidation
	}
	if item.Type != model.ProbeCheckTypeHTTP && item.Target == "" {
		return model.ProbeCheck{}, ErrProbeValidation
	}
	if item.IntervalSeconds <= 0 {
		item.IntervalSeconds = 60
	}
	if item.IntervalSeconds < 15 || item.IntervalSeconds > 86400 {
		return model.ProbeCheck{}, ErrProbeValidation
	}
	if item.TimeoutSeconds <= 0 {
		item.TimeoutSeconds = 10
	}
	if item.TimeoutSeconds < 1 || item.TimeoutSeconds > 120 {
		return model.ProbeCheck{}, ErrProbeValidation
	}
	if item.Retries < 0 || item.Retries > 10 {
		return model.ProbeCheck{}, ErrProbeValidation
	}
	item.Status = strings.TrimSpace(item.Status)
	if item.Status == "" {
		item.Status = model.ProbeStatusUnknown
	}
	if !item.Enabled {
		item.Status = model.ProbeStatusDisabled
	}
	item.Labels = cleanProbeMap(item.Labels)
	item.HTTPConfig.Headers = cleanProbeMap(item.HTTPConfig.Headers)
	item.HTTPConfig.Method = strings.ToUpper(strings.TrimSpace(item.HTTPConfig.Method))
	if item.HTTPConfig.Method == "" {
		item.HTTPConfig.Method = "GET"
	}
	item.DNSConfig.RecordType = strings.ToUpper(strings.TrimSpace(item.DNSConfig.RecordType))
	if item.CreatedAt.IsZero() {
		item.CreatedAt = now
	}
	item.UpdatedAt = now
	return item, nil
}

func normalizeProbeCheckResult(input model.ProbeCheckResult, now time.Time) model.ProbeCheckResult {
	item := input
	item.ID = strings.TrimSpace(item.ID)
	if item.ID == "" {
		item.ID = "probe-result-" + NewID()
	}
	item.Status = normalizeProbeStatus(item.Status)
	item.Error = compactStoreText(item.Error, 512)
	item.Region = compactStoreText(item.Region, 80)
	item.EvidenceRef = compactStoreText(item.EvidenceRef, 512)
	item.Metadata = cleanProbeMap(item.Metadata)
	if item.CheckedAt.IsZero() {
		item.CheckedAt = now
	}
	if item.CreatedAt.IsZero() {
		item.CreatedAt = now
	}
	return item
}

func normalizeProbeStatusPage(input model.ProbeStatusPage, now time.Time) model.ProbeStatusPage {
	item := input
	item.ID = strings.TrimSpace(item.ID)
	if item.ID == "" {
		item.ID = "probe-page-" + NewID()
	}
	item.Slug = normalizeProbeSlug(item.Slug)
	if item.Slug == "" {
		item.Slug = "main"
	}
	item.Title = compactStoreText(firstStoreText(item.Title, "FindX 业务状态页"), 160)
	item.Description = compactStoreText(item.Description, 512)
	item.BusinessGroup = compactStoreText(item.BusinessGroup, 160)
	item.Visibility = strings.TrimSpace(item.Visibility)
	if item.Visibility == "" {
		item.Visibility = "private"
	}
	if len(item.Groups) == 0 {
		item.Groups = []model.ProbeStatusGroup{{ID: "core", Name: "核心平台", CheckIDs: []string{}}}
	}
	item.Labels = cleanProbeMap(item.Labels)
	if item.CreatedAt.IsZero() {
		item.CreatedAt = now
	}
	item.UpdatedAt = now
	return item
}

func normalizeProbeIncident(input model.ProbeIncident, now time.Time) (model.ProbeIncident, error) {
	item := input
	item.ID = strings.TrimSpace(item.ID)
	if item.ID == "" {
		item.ID = "probe-incident-" + NewID()
	}
	item.Title = compactStoreText(item.Title, 180)
	item.Status = strings.ToLower(strings.TrimSpace(item.Status))
	if item.Status == "" {
		item.Status = model.ProbeIncidentStatusInvestigating
	}
	if item.Title == "" || !validProbeIncidentStatus(item.Status) {
		return model.ProbeIncident{}, ErrProbeValidation
	}
	item.CheckID = strings.TrimSpace(item.CheckID)
	item.StatusPageID = strings.TrimSpace(item.StatusPageID)
	item.Severity = compactStoreText(item.Severity, 32)
	item.Message = compactStoreText(item.Message, 2000)
	item.BusinessGroup = compactStoreText(item.BusinessGroup, 160)
	item.Labels = cleanProbeMap(item.Labels)
	if item.StartedAt.IsZero() {
		item.StartedAt = now
	}
	if item.CreatedAt.IsZero() {
		item.CreatedAt = now
	}
	item.UpdatedAt = now
	return item, nil
}

func normalizeProbeNotificationBinding(checkID string, input model.ProbeNotificationBinding, now time.Time) model.ProbeNotificationBinding {
	item := input
	item.ID = strings.TrimSpace(item.ID)
	if item.ID == "" {
		item.ID = "probe-notify-" + NewID()
	}
	item.CheckID = checkID
	item.StatusPageID = strings.TrimSpace(item.StatusPageID)
	item.ChannelID = compactStoreText(item.ChannelID, 64)
	item.ReceiptMode = compactStoreText(firstStoreText(item.ReceiptMode, "blocked_by_contract"), 64)
	item.LastReceiptRef = ""
	item.Labels = cleanProbeMap(item.Labels)
	if item.CreatedAt.IsZero() {
		item.CreatedAt = now
	}
	item.UpdatedAt = now
	return item
}

func normalizeProbeAlertBinding(checkID string, input model.ProbeAlertBinding, now time.Time) model.ProbeAlertBinding {
	item := input
	item.ID = strings.TrimSpace(item.ID)
	if item.ID == "" {
		item.ID = "probe-alert-" + NewID()
	}
	item.CheckID = checkID
	item.AlertRuleID = compactStoreText(item.AlertRuleID, 64)
	item.LastEvidenceRef = ""
	item.Labels = cleanProbeMap(item.Labels)
	if item.CreatedAt.IsZero() {
		item.CreatedAt = now
	}
	item.UpdatedAt = now
	return item
}

func hydrateProbeChecks(items []model.ProbeCheck) []model.ProbeCheck {
	for i := range items {
		results, _ := ListProbeCheckResults(items[i].ID, time.Now().AddDate(0, 0, -90))
		items[i].StatusBar = probeStatusBars(results)
		if len(results) > 0 {
			last := results[0]
			items[i].LastResult = &last
			items[i].Status = last.Status
			if last.ResponseTimeMs > 0 {
				items[i].ResponseTimeMs = &last.ResponseTimeMs
			}
			items[i].Uptime90d = probeUptimePercent(results)
		} else if items[i].Status == "" {
			items[i].Status = model.ProbeStatusUnknown
		}
	}
	return items
}

func probeStatusBars(results []model.ProbeCheckResult) []model.ProbeStatusBarBucket {
	byDate := map[string]model.ProbeCheckResult{}
	for _, result := range results {
		key := result.CheckedAt.Format("2006-01-02")
		if existing, ok := byDate[key]; !ok || result.CheckedAt.After(existing.CheckedAt) {
			byDate[key] = result
		}
	}
	out := make([]model.ProbeStatusBarBucket, 0, 90)
	start := time.Now().AddDate(0, 0, -89)
	for i := 0; i < 90; i++ {
		day := start.AddDate(0, 0, i)
		key := day.Format("2006-01-02")
		bucket := model.ProbeStatusBarBucket{Date: key, Status: model.ProbeStatusNoData, Reason: "no probe run evidence"}
		if result, ok := byDate[key]; ok {
			bucket.Status = result.Status
			bucket.Reason = result.Error
			bucket.EvidenceRef = result.EvidenceRef
			if result.ResponseTimeMs > 0 {
				v := result.ResponseTimeMs
				bucket.ResponseTimeMs = &v
			}
			up := 0.0
			if result.Status == model.ProbeStatusUp {
				up = 100
			}
			bucket.UptimePercent = &up
		}
		out = append(out, bucket)
	}
	return out
}

func probeUptimePercent(results []model.ProbeCheckResult) *float64 {
	total := 0
	up := 0
	for _, result := range results {
		total++
		if result.Status == model.ProbeStatusUp {
			up++
		}
	}
	if total == 0 {
		return nil
	}
	value := float64(up) / float64(total) * 100
	return &value
}

func defaultProbeStatusPage() model.ProbeStatusPage {
	now := time.Now()
	return model.ProbeStatusPage{
		ID:          "probe-page-main",
		Slug:        "main",
		Title:       "FindX 业务状态页",
		Description: "展示业务拨测配置、真实运行证据和人工事故状态；没有拨测运行记录时保持未知状态。",
		Visibility:  "private",
		Groups:      []model.ProbeStatusGroup{{ID: "core", Name: "核心平台", CheckIDs: []string{}}},
		Labels:      map[string]string{"source": "findx"},
		CreatedAt:   now,
		UpdatedAt:   now,
	}
}

func probeCheckMatches(item model.ProbeCheck, filter ProbeCheckFilter) bool {
	if filter.Type != "" && item.Type != filter.Type {
		return false
	}
	if filter.Status != "" && item.Status != filter.Status {
		return false
	}
	if filter.Enabled != nil && item.Enabled != *filter.Enabled {
		return false
	}
	if filter.Query == "" {
		return true
	}
	haystack := strings.ToLower(item.Name + " " + item.URL + " " + item.Target + " " + item.BusinessGroup)
	return strings.Contains(haystack, strings.ToLower(filter.Query))
}

func sortProbeChecks(items []model.ProbeCheck) {
	sort.Slice(items, func(i, j int) bool {
		if items[i].UpdatedAt.Equal(items[j].UpdatedAt) {
			return items[i].ID < items[j].ID
		}
		return items[i].UpdatedAt.After(items[j].UpdatedAt)
	})
}

func validProbeCheckType(value string) bool {
	switch value {
	case model.ProbeCheckTypeHTTP, model.ProbeCheckTypeTCP, model.ProbeCheckTypePing, model.ProbeCheckTypeDNS:
		return true
	default:
		return false
	}
}

func validProbeIncidentStatus(value string) bool {
	switch value {
	case model.ProbeIncidentStatusInvestigating, model.ProbeIncidentStatusIdentified,
		model.ProbeIncidentStatusMonitoring, model.ProbeIncidentStatusResolved:
		return true
	default:
		return false
	}
}

func normalizeProbeStatus(value string) string {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case model.ProbeStatusUp:
		return model.ProbeStatusUp
	case model.ProbeStatusDown:
		return model.ProbeStatusDown
	case model.ProbeStatusDegraded:
		return model.ProbeStatusDegraded
	case model.ProbeStatusDisabled:
		return model.ProbeStatusDisabled
	default:
		return model.ProbeStatusUnknown
	}
}

func normalizeProbeSlug(value string) string {
	value = strings.ToLower(strings.TrimSpace(value))
	value = strings.ReplaceAll(value, " ", "-")
	if value == "" {
		return "main"
	}
	return compactStoreText(value, 120)
}

func cleanProbeMap(input map[string]string) map[string]string {
	if len(input) == 0 {
		return map[string]string{}
	}
	out := map[string]string{}
	for key, value := range input {
		key = compactStoreText(key, 80)
		if key == "" || probeSensitiveKey(key) {
			continue
		}
		out[key] = compactStoreText(value, 256)
	}
	return out
}

func probeSensitiveKey(key string) bool {
	key = strings.ToLower(key)
	return strings.Contains(key, "token") ||
		strings.Contains(key, "password") ||
		strings.Contains(key, "secret") ||
		strings.Contains(key, "cookie") ||
		strings.Contains(key, "authorization") ||
		strings.Contains(key, "dsn")
}

func compactStoreText(value string, limit int) string {
	value = strings.TrimSpace(value)
	value = strings.ReplaceAll(value, "\x00", "")
	if limit > 0 && len(value) > limit {
		return value[:limit]
	}
	return value
}

func firstStoreText(values ...string) string {
	for _, value := range values {
		if strings.TrimSpace(value) != "" {
			return value
		}
	}
	return ""
}

func copyProbeCheck(item model.ProbeCheck) model.ProbeCheck {
	item.Labels = copyProbeStringMap(item.Labels)
	item.HTTPConfig.Headers = copyProbeStringMap(item.HTTPConfig.Headers)
	return item
}

func copyProbeCheckResult(item model.ProbeCheckResult) model.ProbeCheckResult {
	item.Metadata = copyProbeStringMap(item.Metadata)
	return item
}

func copyProbeStatusPage(item model.ProbeStatusPage) model.ProbeStatusPage {
	item.Labels = copyProbeStringMap(item.Labels)
	item.Groups = append([]model.ProbeStatusGroup(nil), item.Groups...)
	for i := range item.Groups {
		item.Groups[i].CheckIDs = append([]string(nil), item.Groups[i].CheckIDs...)
	}
	return item
}

func copyProbeIncident(item model.ProbeIncident) model.ProbeIncident {
	item.Labels = copyProbeStringMap(item.Labels)
	return item
}

func copyProbeNotificationBinding(item model.ProbeNotificationBinding) model.ProbeNotificationBinding {
	item.Labels = copyProbeStringMap(item.Labels)
	return item
}

func copyProbeAlertBinding(item model.ProbeAlertBinding) model.ProbeAlertBinding {
	item.Labels = copyProbeStringMap(item.Labels)
	return item
}

func copyProbeStringMap(input map[string]string) map[string]string {
	if len(input) == 0 {
		return map[string]string{}
	}
	out := make(map[string]string, len(input))
	for key, value := range input {
		out[key] = value
	}
	return out
}
