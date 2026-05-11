package handler

import (
	"net/http"
	"strconv"
	"strings"

	"ai-workbench-api/internal/model"
	"ai-workbench-api/internal/store"
)

func validateBuiltinComponent(item model.MonitoringBuiltinComponent, create bool) (int, []string) {
	checks := basicBuiltinComponentChecks(item, create)
	if len(checks) > 0 {
		return http.StatusBadRequest, checks
	}
	return validateBuiltinComponentConflict(item, create)
}

func basicBuiltinComponentChecks(item model.MonitoringBuiltinComponent, create bool) []string {
	checks := []string{}
	if !create && item.ID == "" {
		checks = append(checks, "id is required")
	}
	if item.ID != "" && !validBuiltinID(item.ID) {
		checks = append(checks, "id contains unsupported characters")
	}
	if !validBuiltinIdent(item.Ident) {
		checks = append(checks, "ident is required and must use safe characters")
	}
	if item.Name == "" || len([]rune(item.Name)) > maxBuiltinNameLen {
		checks = append(checks, "name is required and must fit length limit")
	}
	if item.Logo == "" || !safeBuiltinRoute(item.Logo) {
		checks = append(checks, "logo must be a safe FindX relative route")
	}
	if len([]rune(item.Readme)) > maxBuiltinReadmeLen || unsafeBuiltinText(item.Ident+" "+item.Name+" "+item.Logo+" "+item.Readme) {
		checks = append(checks, "component fields contain blocked content")
	}
	return checks
}

func validateBuiltinComponentConflict(item model.MonitoringBuiltinComponent, create bool) (int, []string) {
	for _, existing := range store.ListMonitoringBuiltinComponents() {
		sameID := item.ID != "" && existing.ID == item.ID
		sameIdent := strings.EqualFold(existing.Ident, item.Ident)
		if create && (sameID || sameIdent) {
			return http.StatusConflict, []string{"component id or ident already exists"}
		}
		if !create && sameIdent && existing.ID != item.ID {
			return http.StatusConflict, []string{"component ident already exists"}
		}
	}
	return http.StatusOK, nil
}

func validateBuiltinPayload(item model.MonitoringBuiltinPayload, create bool) []string {
	checks := basicBuiltinPayloadChecks(item, create)
	if len(checks) > 0 {
		return checks
	}
	return builtinPayloadConflictChecks(item, create)
}

func basicBuiltinPayloadChecks(item model.MonitoringBuiltinPayload, create bool) []string {
	checks := []string{}
	if !create && item.ID == "" {
		checks = append(checks, "id is required")
	}
	if item.ID != "" && !validBuiltinID(item.ID) {
		checks = append(checks, "id contains unsupported characters")
	}
	if item.UUID != "" && !validBuiltinID(item.UUID) {
		checks = append(checks, "uuid contains unsupported characters")
	}
	if !writableBuiltinPayloadTypes[item.Type] {
		checks = append(checks, "payload type is not writable by this contract")
	}
	if !builtinComponentExists(item.ComponentID) {
		checks = append(checks, "component_id is unknown")
	}
	if item.Name == "" || len([]rune(item.Name)) > maxBuiltinNameLen {
		checks = append(checks, "name is required and must fit length limit")
	}
	if len([]rune(item.Cate)) > maxBuiltinCateLen || unsafeBuiltinText(item.Cate+" "+item.Name) {
		checks = append(checks, "payload fields contain blocked content")
	}
	return checks
}

func builtinPayloadConflictChecks(item model.MonitoringBuiltinPayload, create bool) []string {
	for _, existing := range store.ListMonitoringBuiltinPayloads(model.MonitoringBuiltinPayloadFilter{}) {
		sameID := item.ID != "" && existing.ID == item.ID
		sameUUID := item.UUID != "" && existing.UUID == item.UUID
		if create && (sameID || sameUUID) {
			return []string{"payload id or uuid already exists"}
		}
		if !create && sameUUID && existing.ID != item.ID {
			return []string{"payload uuid already exists"}
		}
	}
	return nil
}

func intAuditValue(value int) string {
	return strconv.Itoa(value)
}
