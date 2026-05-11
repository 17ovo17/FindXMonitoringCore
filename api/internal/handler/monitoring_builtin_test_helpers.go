package handler

import (
	"ai-workbench-api/internal/model"
	"ai-workbench-api/internal/store"
)

func containsBuiltinString(values []string, needle string) bool {
	for _, value := range values {
		if value == needle {
			return true
		}
	}
	return false
}

func componentListed(id string) bool {
	for _, item := range modelComponentList() {
		if item.ID == id {
			return true
		}
	}
	return false
}

func modelComponentList() []model.MonitoringBuiltinComponent {
	return store.ListMonitoringBuiltinComponents()
}
