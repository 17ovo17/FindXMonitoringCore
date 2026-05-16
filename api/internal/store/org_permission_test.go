package store

import (
	"strings"
	"testing"

	"ai-workbench-api/internal/model"
)

func TestOrgCMDBAndAIOpsOperationsMapToMonitorPermissions(t *testing.T) {
	checker := model.NewMonitorPermissionChecker()
	legacyOnlyOrgOperations := map[string]bool{
		"workflow.read":   true,
		"workflow.write":  true,
		"knowledge.read":  true,
		"knowledge.write": true,
	}

	for _, group := range orgOperationGroups {
		if group.Name != "cmdb" && group.Name != "aiops" {
			continue
		}
		for _, op := range group.Ops {
			if legacyOnlyOrgOperations[op.Name] {
				continue
			}
			resource, action, ok := splitOrgOperationPermission(op.Name)
			if !ok {
				t.Fatalf("org operation %s cannot be split into resource/action", op.Name)
			}
			result := checker.Check("admin", resource, action)
			if !result.Known {
				t.Fatalf("org operation %s maps to unknown monitor permission %s:%s", op.Name, resource, action)
			}
		}
	}
}

func splitOrgOperationPermission(name string) (string, string, bool) {
	idx := strings.LastIndex(name, ".")
	if idx <= 0 || idx == len(name)-1 {
		return "", "", false
	}
	return name[:idx], name[idx+1:], true
}
