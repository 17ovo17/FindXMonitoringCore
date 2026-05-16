package store

import (
	"crypto/sha1"
	"database/sql"
	"encoding/json"
	"fmt"
	"sort"
	"strings"
	"time"

	"ai-workbench-api/internal/model"
)

func UpsertMonitorTarget(target *model.MonitorTarget) (*model.MonitorTarget, error) {
	if !validMonitorStatus(target.Status) {
		return nil, fmt.Errorf("invalid monitor target status")
	}
	now := time.Now()
	normalized := normalizeMonitorTarget(target, now)
	mu.Lock()
	mergeMonitorTargetMemory(normalized, now)
	out := copyMonitorTarget(monitorTargets[normalized.ID])
	mu.Unlock()
	if mysqlOK {
		if err := persistMonitorTarget(out); err != nil {
			return out, err
		}
	}
	return out, nil
}

func validMonitorStatus(status string) bool {
	switch strings.TrimSpace(status) {
	case "", "online", "warning", "offline", "unknown", "maintenance":
		return true
	default:
		return false
	}
}

func ListMonitorTargets() []*model.MonitorTarget {
	if mysqlOK {
		if rows, err := db.Query(`SELECT id,ident,name,ip,hostname,os,arch,environment,business_group,owner,status,source,COALESCE(labels,'{}'),COALESCE(metadata,'{}'),last_seen,created_at,updated_at FROM monitor_targets ORDER BY updated_at DESC LIMIT 1000`); err == nil {
			defer rows.Close()
			return scanMonitorTargets(rows)
		}
	}
	mu.RLock()
	defer mu.RUnlock()
	out := make([]*model.MonitorTarget, 0, len(monitorTargets))
	for _, target := range monitorTargets {
		out = append(out, copyMonitorTarget(target))
	}
	sort.Slice(out, func(i, j int) bool { return out[i].UpdatedAt.After(out[j].UpdatedAt) })
	return out
}

func GetMonitorTarget(id string) (*model.MonitorTarget, bool) {
	for _, target := range ListMonitorTargets() {
		if target.ID == id || target.Ident == id {
			return target, true
		}
	}
	return nil, false
}

func DeleteMonitorTarget(id string) (bool, error) {
	found := false
	mu.Lock()
	if _, ok := monitorTargets[id]; ok {
		delete(monitorTargets, id)
		found = true
	}
	for key, target := range monitorTargets {
		if target.Ident == id {
			delete(monitorTargets, key)
			found = true
		}
	}
	mu.Unlock()
	if mysqlOK {
		res, err := db.Exec(`DELETE FROM monitor_targets WHERE id=? OR ident=?`, id, id)
		if err != nil {
			return found, err
		}
		if res == nil {
			return found, fmt.Errorf("delete target returned no result")
		}
		if rows, err := res.RowsAffected(); err == nil && rows > 0 {
			found = true
		}
	}
	return found, nil
}

func UpsertFindXAgentHeartbeat(hb model.FindXAgentHeartbeat) (*model.FindXAgent, *model.MonitorTarget, error) {
	now, err := HeartbeatTime(hb.UnixTime)
	if err != nil {
		return nil, nil, err
	}
	agent := normalizeFindXAgent(hb, now)
	target := targetFromFindXAgent(agent, now)
	target, err = UpsertMonitorTarget(target)
	if err != nil {
		return nil, target, err
	}
	agent.TargetID = target.ID
	mu.Lock()
	mergeFindXAgentMemory(agent, now)
	out := copyFindXAgent(findxAgents[agent.ID])
	mu.Unlock()
	if mysqlOK {
		if err := persistFindXAgent(out); err != nil {
			return out, target, err
		}
	}
	return out, target, nil
}

func GetFindXAgent(id string) (*model.FindXAgent, bool) {
	if mysqlOK {
		if rows, err := db.Query(`SELECT id,ident,target_id,ip,hostname,os,arch,version,collector,status,COALESCE(capabilities,'[]'),COALESCE(global_labels,'{}'),config_version,last_seen,created_at,updated_at FROM findx_agents WHERE id=?`, id); err == nil {
			defer rows.Close()
			agents := scanFindXAgents(rows)
			if len(agents) > 0 {
				return agents[0], true
			}
		}
	}
	mu.RLock()
	defer mu.RUnlock()
	agent, ok := findxAgents[id]
	if !ok {
		return nil, false
	}
	return copyFindXAgent(agent), true
}

func DeleteFindXAgent(id string) (bool, error) {
	found := false
	mu.Lock()
	if _, ok := findxAgents[id]; ok {
		delete(findxAgents, id)
		found = true
	}
	for key, agent := range findxAgents {
		if agent.Ident == id {
			delete(findxAgents, key)
			found = true
		}
	}
	mu.Unlock()
	if mysqlOK {
		res, err := db.Exec(`DELETE FROM findx_agents WHERE id=? OR ident=?`, id, id)
		if err != nil {
			return found, err
		}
		if res == nil {
			return found, fmt.Errorf("delete findx agent returned no result")
		}
		if rows, err := res.RowsAffected(); err == nil && rows > 0 {
			found = true
		}
	}
	return found, nil
}

func ListFindXAgents() []*model.FindXAgent {
	if mysqlOK {
		if rows, err := db.Query(`SELECT id,ident,target_id,ip,hostname,os,arch,version,collector,status,COALESCE(capabilities,'[]'),COALESCE(global_labels,'{}'),config_version,last_seen,created_at,updated_at FROM findx_agents ORDER BY last_seen DESC LIMIT 1000`); err == nil {
			defer rows.Close()
			return scanFindXAgents(rows)
		}
	}
	mu.RLock()
	defer mu.RUnlock()
	out := make([]*model.FindXAgent, 0, len(findxAgents))
	for _, agent := range findxAgents {
		out = append(out, copyFindXAgent(agent))
	}
	sort.Slice(out, func(i, j int) bool { return out[i].LastSeen.After(out[j].LastSeen) })
	return out
}

func normalizeMonitorTarget(target *model.MonitorTarget, now time.Time) *model.MonitorTarget {
	cp := copyMonitorTarget(target)
	cp.Ident = firstNonEmpty(cp.Ident, cp.IP, cp.Hostname, cp.ID)
	if cp.ID == "" || len(cp.ID) > 64 {
		cp.ID = stableMonitorID("mt", cp.Ident)
	}
	cp.Name = firstNonEmpty(cp.Name, cp.Hostname, cp.IP, cp.Ident)
	cp.Status = firstNonEmpty(cp.Status, "unknown")
	cp.Source = firstNonEmpty(cp.Source, "manual")
	cp.Labels = sanitizeStringMap(cp.Labels)
	cp.Metadata = sanitizeStringMap(cp.Metadata)
	if cp.CreatedAt.IsZero() {
		cp.CreatedAt = now
	}
	cp.UpdatedAt = now
	return cp
}

func mergeMonitorTargetMemory(incoming *model.MonitorTarget, now time.Time) {
	existing := monitorTargets[incoming.ID]
	if existing == nil {
		monitorTargets[incoming.ID] = incoming
		return
	}
	createdAt := existing.CreatedAt
	*existing = *incoming
	existing.CreatedAt = createdAt
	existing.UpdatedAt = now
}

func normalizeFindXAgent(hb model.FindXAgentHeartbeat, now time.Time) *model.FindXAgent {
	ident := firstNonEmpty(hb.Ident, hb.IP, hb.Hostname)
	return &model.FindXAgent{
		ID:            stableMonitorID("fa", ident),
		Ident:         ident,
		IP:            strings.TrimSpace(hb.IP),
		Hostname:      strings.TrimSpace(hb.Hostname),
		OS:            strings.TrimSpace(hb.OS),
		Arch:          strings.TrimSpace(hb.Arch),
		Version:       strings.TrimSpace(hb.Version),
		Collector:     firstNonEmpty(hb.Collector, "categraf"),
		Status:        statusByLastSeen(now),
		Capabilities:  dedupeStrings(hb.Capabilities),
		GlobalLabels:  sanitizeStringMap(hb.GlobalLabels),
		ConfigVersion: strings.TrimSpace(hb.ConfigVersion),
		LastSeen:      now,
		CreatedAt:     now,
		UpdatedAt:     now,
	}
}

func targetFromFindXAgent(agent *model.FindXAgent, now time.Time) *model.MonitorTarget {
	return &model.MonitorTarget{
		ID:       stableMonitorID("mt", agent.Ident),
		Ident:    agent.Ident,
		Name:     firstNonEmpty(agent.Hostname, agent.IP, agent.Ident),
		IP:       agent.IP,
		Hostname: agent.Hostname,
		OS:       agent.OS,
		Arch:     agent.Arch,
		Status:   statusByLastSeen(agent.LastSeen),
		Source:   "findx-agents",
		Labels:   agent.GlobalLabels,
		Metadata: map[string]string{"agent_version": agent.Version, "collector": agent.Collector},
		LastSeen: &now,
	}
}

func mergeFindXAgentMemory(incoming *model.FindXAgent, now time.Time) {
	existing := findxAgents[incoming.ID]
	if existing == nil {
		findxAgents[incoming.ID] = incoming
		return
	}
	createdAt := existing.CreatedAt
	*existing = *incoming
	existing.CreatedAt = createdAt
	existing.UpdatedAt = now
}

func persistMonitorTarget(target *model.MonitorTarget) error {
	labels, _ := json.Marshal(target.Labels)
	metadata, _ := json.Marshal(target.Metadata)
	_, err := db.Exec(`REPLACE INTO monitor_targets (id,ident,name,ip,hostname,os,arch,environment,business_group,owner,status,source,labels,metadata,last_seen,created_at,updated_at) VALUES (?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?)`, target.ID, target.Ident, target.Name, target.IP, target.Hostname, target.OS, target.Arch, target.Environment, target.BusinessGroup, target.Owner, target.Status, target.Source, string(labels), string(metadata), nullableTime(target.LastSeen), target.CreatedAt, target.UpdatedAt)
	return err
}

func persistFindXAgent(agent *model.FindXAgent) error {
	capabilities, _ := json.Marshal(agent.Capabilities)
	labels, _ := json.Marshal(agent.GlobalLabels)
	_, err := db.Exec(`REPLACE INTO findx_agents (id,ident,target_id,ip,hostname,os,arch,version,collector,status,capabilities,global_labels,config_version,last_seen,created_at,updated_at) VALUES (?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?)`, agent.ID, agent.Ident, agent.TargetID, agent.IP, agent.Hostname, agent.OS, agent.Arch, agent.Version, agent.Collector, statusByLastSeen(agent.LastSeen), string(capabilities), string(labels), agent.ConfigVersion, agent.LastSeen, agent.CreatedAt, agent.UpdatedAt)
	return err
}

func scanMonitorTargets(rows *sql.Rows) []*model.MonitorTarget {
	out := []*model.MonitorTarget{}
	for rows.Next() {
		if target, ok := scanMonitorTargetRow(rows); ok {
			out = append(out, target)
		}
	}
	if err := rows.Err(); err != nil {
		return out
	}
	return out
}

func scanFindXAgents(rows *sql.Rows) []*model.FindXAgent {
	out := []*model.FindXAgent{}
	for rows.Next() {
		if agent, ok := scanFindXAgentRow(rows); ok {
			out = append(out, agent)
		}
	}
	if err := rows.Err(); err != nil {
		return out
	}
	return out
}

type monitorTargetScanner interface {
	Scan(dest ...any) error
}

func scanMonitorTargetRow(row monitorTargetScanner) (*model.MonitorTarget, bool) {
	target := model.MonitorTarget{}
	var labels, metadata string
	var lastSeen sql.NullTime
	if err := row.Scan(&target.ID, &target.Ident, &target.Name, &target.IP, &target.Hostname, &target.OS, &target.Arch, &target.Environment, &target.BusinessGroup, &target.Owner, &target.Status, &target.Source, &labels, &metadata, &lastSeen, &target.CreatedAt, &target.UpdatedAt); err != nil {
		return nil, false
	}
	if !unmarshalStringMap(labels, &target.Labels) || !unmarshalStringMap(metadata, &target.Metadata) {
		return nil, false
	}
	if lastSeen.Valid {
		target.LastSeen = &lastSeen.Time
		target.Status = statusByLastSeen(lastSeen.Time)
	}
	return &target, true
}

type findXAgentScanner interface {
	Scan(dest ...any) error
}

func scanFindXAgentRow(row findXAgentScanner) (*model.FindXAgent, bool) {
	agent := model.FindXAgent{}
	var capabilities, labels string
	if err := row.Scan(&agent.ID, &agent.Ident, &agent.TargetID, &agent.IP, &agent.Hostname, &agent.OS, &agent.Arch, &agent.Version, &agent.Collector, &agent.Status, &capabilities, &labels, &agent.ConfigVersion, &agent.LastSeen, &agent.CreatedAt, &agent.UpdatedAt); err != nil {
		return nil, false
	}
	if !unmarshalStringSlice(capabilities, &agent.Capabilities) || !unmarshalStringMap(labels, &agent.GlobalLabels) {
		return nil, false
	}
	agent.Status = statusByLastSeen(agent.LastSeen)
	return &agent, true
}

func unmarshalStringSlice(raw string, dest *[]string) bool {
	if raw == "" {
		*dest = []string{}
		return true
	}
	return json.Unmarshal([]byte(raw), dest) == nil
}

func HeartbeatTime(unixTime int64) (time.Time, error) {
	now := time.Now()
	if unixTime <= 0 {
		return now, nil
	}
	var t time.Time
	if unixTime > 1_000_000_000_000 {
		t = time.UnixMilli(unixTime)
	} else {
		t = time.Unix(unixTime, 0)
	}
	if t.After(now.Add(5 * time.Minute)) {
		return now, fmt.Errorf("heartbeat time is too far in the future")
	}
	return t, nil
}

func statusByLastSeen(lastSeen time.Time) string {
	delta := time.Since(lastSeen)
	if delta < 2*time.Minute {
		return "online"
	}
	if delta < 5*time.Minute {
		return "warning"
	}
	return "offline"
}

func copyMonitorTarget(in *model.MonitorTarget) *model.MonitorTarget {
	if in == nil {
		return &model.MonitorTarget{}
	}
	cp := *in
	cp.Labels = copyStringMap(in.Labels)
	cp.Metadata = copyStringMap(in.Metadata)
	return &cp
}

func copyFindXAgent(in *model.FindXAgent) *model.FindXAgent {
	if in == nil {
		return &model.FindXAgent{}
	}
	cp := *in
	cp.Capabilities = append([]string{}, in.Capabilities...)
	cp.GlobalLabels = copyStringMap(in.GlobalLabels)
	return &cp
}

func copyStringMap(in map[string]string) map[string]string {
	out := map[string]string{}
	for key, value := range in {
		out[key] = value
	}
	return out
}

func sanitizeStringMap(in map[string]string) map[string]string {
	out := map[string]string{}
	for key, value := range in {
		if isSensitiveKey(key) {
			out[key] = "******"
		} else {
			out[key] = strings.TrimSpace(value)
		}
	}
	return out
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		if trimmed := strings.TrimSpace(value); trimmed != "" {
			return trimmed
		}
	}
	return ""
}

func dedupeStrings(values []string) []string {
	seen := map[string]bool{}
	out := []string{}
	for _, value := range values {
		trimmed := strings.TrimSpace(value)
		if trimmed != "" && !seen[trimmed] {
			seen[trimmed] = true
			out = append(out, trimmed)
		}
	}
	sort.Strings(out)
	return out
}

func stableMonitorID(prefix, ident string) string {
	sum := sha1.Sum([]byte(strings.TrimSpace(ident)))
	return fmt.Sprintf("%s_%x", prefix, sum)[:43]
}
