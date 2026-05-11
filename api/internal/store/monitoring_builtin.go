package store

import (
	"errors"
	"sort"
	"strings"

	"ai-workbench-api/internal/model"
)

const builtinUpdatedBy = "findx-builtin-catalog"

var ErrMonitoringBuiltinProtected = errors.New("protected builtin row")
var ErrMonitoringBuiltinComponentHasPayloads = errors.New("builtin component has payloads")

func ListMonitoringBuiltinComponents() []model.MonitoringBuiltinComponent {
	payloads := ListMonitoringBuiltinPayloads(model.MonitoringBuiltinPayloadFilter{})
	components := append(builtinComponentCatalog(), customMonitoringBuiltinComponents()...)
	byID := map[string]*model.MonitoringBuiltinComponent{}
	for i := range components {
		byID[components[i].ID] = &components[i]
	}
	for _, payload := range payloads {
		component := byID[payload.ComponentID]
		if component == nil {
			continue
		}
		switch payload.Type {
		case "dashboard":
			component.DashboardCount++
		case "collect":
			component.CollectCount++
		case "metric":
			component.MetricCount++
		case "alert":
			component.AlertCount++
		case "record":
			component.RecordCount++
		}
	}
	return components
}

func ListMonitoringBuiltinPayloads(filter model.MonitoringBuiltinPayloadFilter) []model.MonitoringBuiltinPayload {
	payloads := append(builtinPayloadCatalog(), customMonitoringBuiltinPayloads()...)
	out := make([]model.MonitoringBuiltinPayload, 0, len(payloads))
	for _, payload := range payloads {
		if monitoringBuiltinPayloadMatches(payload, filter) {
			out = append(out, copyMonitoringBuiltinPayload(payload))
		}
	}
	return out
}

func ListMonitoringBuiltinPayloadTypes() []string {
	seen := map[string]bool{}
	for _, payload := range append(builtinPayloadCatalog(), customMonitoringBuiltinPayloads()...) {
		seen[payload.Type] = true
	}
	out := make([]string, 0, len(seen))
	for value := range seen {
		out = append(out, value)
	}
	sort.Strings(out)
	return out
}

func GetMonitoringBuiltinPayload(id string) (model.MonitoringBuiltinPayload, bool) {
	needle := strings.TrimSpace(id)
	for _, payload := range append(builtinPayloadCatalog(), customMonitoringBuiltinPayloads()...) {
		if payload.ID == needle {
			return copyMonitoringBuiltinPayload(payload), true
		}
	}
	return model.MonitoringBuiltinPayload{}, false
}

func SaveMonitoringBuiltinComponent(input model.MonitoringBuiltinComponent, actor string) (model.MonitoringBuiltinComponent, error) {
	if isStaticMonitoringBuiltinComponent(input.ID) || isStaticMonitoringBuiltinComponentIdent(input.Ident) {
		return model.MonitoringBuiltinComponent{}, ErrMonitoringBuiltinProtected
	}
	item := normalizeMonitoringBuiltinComponent(input, actor)
	if err := persistMonitoringBuiltinComponent(item); err != nil {
		return model.MonitoringBuiltinComponent{}, err
	}
	return copyMonitoringBuiltinComponent(item), nil
}

func UpdateMonitoringBuiltinComponent(input model.MonitoringBuiltinComponent, actor string) (model.MonitoringBuiltinComponent, bool, error) {
	if isStaticMonitoringBuiltinComponent(input.ID) || isStaticMonitoringBuiltinComponentIdent(input.Ident) {
		return model.MonitoringBuiltinComponent{}, true, ErrMonitoringBuiltinProtected
	}
	if _, ok := getCustomMonitoringBuiltinComponent(input.ID); !ok {
		return model.MonitoringBuiltinComponent{}, false, nil
	}
	item := normalizeMonitoringBuiltinComponent(input, actor)
	if err := persistMonitoringBuiltinComponent(item); err != nil {
		return model.MonitoringBuiltinComponent{}, true, err
	}
	return copyMonitoringBuiltinComponent(item), true, nil
}

func DeleteMonitoringBuiltinComponents(ids []string) (bool, error) {
	if len(ids) == 0 {
		return false, nil
	}
	for _, id := range ids {
		if isStaticMonitoringBuiltinComponent(id) {
			return true, ErrMonitoringBuiltinProtected
		}
		if customMonitoringBuiltinPayloadCountByComponent(id) > 0 {
			return true, ErrMonitoringBuiltinComponentHasPayloads
		}
	}
	found := false
	for _, id := range ids {
		if mysqlOK {
			res, err := db.Exec(`DELETE FROM monitor_builtin_components WHERE id=?`, id)
			if err != nil {
				return found, err
			}
			rows, err := res.RowsAffected()
			if err != nil {
				return found, err
			}
			found = found || rows > 0
		}
		mu.Lock()
		if _, ok := monitorBuiltinComponents[id]; ok {
			delete(monitorBuiltinComponents, id)
			found = true
		}
		mu.Unlock()
	}
	return found, nil
}

func SaveMonitoringBuiltinPayload(input model.MonitoringBuiltinPayload, actor string) (model.MonitoringBuiltinPayload, error) {
	if isStaticMonitoringBuiltinPayload(input.ID) {
		return model.MonitoringBuiltinPayload{}, ErrMonitoringBuiltinProtected
	}
	item := normalizeMonitoringBuiltinPayload(input, actor)
	if err := persistMonitoringBuiltinPayload(item); err != nil {
		return model.MonitoringBuiltinPayload{}, err
	}
	return copyMonitoringBuiltinPayload(item), nil
}

func UpdateMonitoringBuiltinPayload(input model.MonitoringBuiltinPayload, actor string) (model.MonitoringBuiltinPayload, bool, error) {
	if isStaticMonitoringBuiltinPayload(input.ID) {
		return model.MonitoringBuiltinPayload{}, true, ErrMonitoringBuiltinProtected
	}
	if _, ok := getCustomMonitoringBuiltinPayload(input.ID); !ok {
		return model.MonitoringBuiltinPayload{}, false, nil
	}
	item := normalizeMonitoringBuiltinPayload(input, actor)
	if err := persistMonitoringBuiltinPayload(item); err != nil {
		return model.MonitoringBuiltinPayload{}, true, err
	}
	return copyMonitoringBuiltinPayload(item), true, nil
}

func DeleteMonitoringBuiltinPayloads(ids []string) (bool, error) {
	if len(ids) == 0 {
		return false, nil
	}
	for _, id := range ids {
		if isStaticMonitoringBuiltinPayload(id) {
			return true, ErrMonitoringBuiltinProtected
		}
	}
	found := false
	for _, id := range ids {
		if mysqlOK {
			res, err := db.Exec(`DELETE FROM monitor_builtin_payloads WHERE id=?`, id)
			if err != nil {
				return found, err
			}
			rows, err := res.RowsAffected()
			if err != nil {
				return found, err
			}
			found = found || rows > 0
		}
		mu.Lock()
		if _, ok := monitorBuiltinPayloads[id]; ok {
			delete(monitorBuiltinPayloads, id)
			found = true
		}
		mu.Unlock()
	}
	return found, nil
}

func monitoringBuiltinPayloadMatches(payload model.MonitoringBuiltinPayload, filter model.MonitoringBuiltinPayloadFilter) bool {
	componentID := strings.TrimSpace(filter.ComponentID)
	if componentID != "" && payload.ComponentID != componentID {
		return false
	}
	payloadType := strings.TrimSpace(filter.Type)
	if payloadType != "" && payload.Type != payloadType {
		return false
	}
	query := strings.ToLower(strings.TrimSpace(filter.Query))
	if query == "" {
		return true
	}
	fields := []string{payload.ID, payload.ComponentID, payload.Type, payload.Name, payload.Title, payload.Description, payload.Note}
	fields = append(fields, payload.Tags...)
	for _, field := range fields {
		if strings.Contains(strings.ToLower(field), query) {
			return true
		}
	}
	return false
}

func copyMonitoringBuiltinPayload(in model.MonitoringBuiltinPayload) model.MonitoringBuiltinPayload {
	in.Tags = append([]string{}, in.Tags...)
	in.Content = append([]byte{}, in.Content...)
	return in
}

func copyMonitoringBuiltinComponent(in model.MonitoringBuiltinComponent) model.MonitoringBuiltinComponent {
	in.Tags = append([]string{}, in.Tags...)
	return in
}

func customMonitoringBuiltinComponentExists(id string) bool {
	_, ok := getCustomMonitoringBuiltinComponent(id)
	if ok {
		return true
	}
	return isStaticMonitoringBuiltinComponent(id)
}

func customMonitoringBuiltinPayloadCountByComponent(componentID string) int {
	needle := strings.TrimSpace(componentID)
	if needle == "" {
		return 0
	}
	if mysqlOK {
		var count int
		if err := db.QueryRow(`SELECT COUNT(*) FROM monitor_builtin_payloads WHERE component_id=?`, needle).Scan(&count); err == nil {
			return count
		}
		return 0
	}
	mu.RLock()
	defer mu.RUnlock()
	count := 0
	for _, item := range monitorBuiltinPayloads {
		if item.ComponentID == needle {
			count++
		}
	}
	return count
}

func normalizeMonitoringBuiltinComponent(input model.MonitoringBuiltinComponent, actor string) model.MonitoringBuiltinComponent {
	item := copyMonitoringBuiltinComponent(input)
	if strings.TrimSpace(item.ID) == "" {
		item.ID = NewID()
	}
	item.ID = strings.TrimSpace(item.ID)
	item.Ident = strings.TrimSpace(item.Ident)
	item.Name = strings.TrimSpace(item.Name)
	item.Logo = strings.TrimSpace(item.Logo)
	item.Readme = strings.TrimSpace(item.Readme)
	if item.Disabled != 1 {
		item.Disabled = 0
	}
	item.UpdatedBy = actor
	return item
}

func normalizeMonitoringBuiltinPayload(input model.MonitoringBuiltinPayload, actor string) model.MonitoringBuiltinPayload {
	item := copyMonitoringBuiltinPayload(input)
	if strings.TrimSpace(item.ID) == "" {
		item.ID = NewID()
	}
	if strings.TrimSpace(item.UUID) == "" {
		item.UUID = item.ID
	}
	item.ID = strings.TrimSpace(item.ID)
	item.UUID = strings.TrimSpace(item.UUID)
	item.Type = strings.TrimSpace(item.Type)
	item.ComponentID = strings.TrimSpace(item.ComponentID)
	item.Cate = strings.TrimSpace(item.Cate)
	item.Name = strings.TrimSpace(item.Name)
	item.Title = item.Name
	item.UpdatedBy = actor
	return item
}

func isStaticMonitoringBuiltinComponent(id string) bool {
	needle := strings.TrimSpace(id)
	if needle == "" {
		return false
	}
	for _, item := range builtinComponentCatalog() {
		if item.ID == needle {
			return true
		}
	}
	return false
}

func isStaticMonitoringBuiltinComponentIdent(ident string) bool {
	needle := strings.TrimSpace(ident)
	if needle == "" {
		return false
	}
	for _, item := range builtinComponentCatalog() {
		if item.Ident == needle {
			return true
		}
	}
	return false
}

func isStaticMonitoringBuiltinPayload(id string) bool {
	needle := strings.TrimSpace(id)
	if needle == "" {
		return false
	}
	for _, item := range builtinPayloadCatalog() {
		if item.ID == needle {
			return true
		}
	}
	return false
}
