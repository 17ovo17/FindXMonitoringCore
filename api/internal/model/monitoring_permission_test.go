package model

import "testing"

func TestMonitorPermissionCMDBUserReadOnlyDefaults(t *testing.T) {
	checker := NewMonitorPermissionChecker()

	read := checker.Check("user", "cmdb.instance", "read")
	if !read.Known || !read.Allowed || read.Reason != MonitorPermissionReasonRoleAllowed {
		t.Fatalf("user cmdb.instance read = known:%v allowed:%v reason:%s; want allowed",
			read.Known, read.Allowed, read.Reason)
	}

	create := checker.Check("user", "cmdb.instance", "create")
	if !create.Known || create.Allowed || create.Reason != MonitorPermissionReasonRoleDenied {
		t.Fatalf("user cmdb.instance create = known:%v allowed:%v reason:%s; want denied",
			create.Known, create.Allowed, create.Reason)
	}
}

func TestMonitorPermissionUnknownResourceStaysUnknown(t *testing.T) {
	result := NewMonitorPermissionChecker().Check("user", "cmdb.unknown", "read")
	if result.Known || result.Allowed || result.Reason != MonitorPermissionReasonUnknown {
		t.Fatalf("unknown permission = known:%v allowed:%v reason:%s; want unknown",
			result.Known, result.Allowed, result.Reason)
	}
}

func TestMonitorPermissionCMDBHighRiskUserDeniedAndAdminAllowed(t *testing.T) {
	checker := NewMonitorPermissionChecker()
	tests := []struct {
		resource string
		action   string
	}{
		{resource: "cmdb.command", action: "exec"},
		{resource: "cmdb.terminal", action: "open"},
		{resource: "cmdb.database", action: "test"},
	}

	for _, tt := range tests {
		t.Run(tt.resource+":"+tt.action, func(t *testing.T) {
			user := checker.Check("user", tt.resource, tt.action)
			if !user.Known || user.Allowed || user.Reason != MonitorPermissionReasonRoleDenied {
				t.Fatalf("user %s:%s = known:%v allowed:%v reason:%s; want denied",
					tt.resource, tt.action, user.Known, user.Allowed, user.Reason)
			}

			admin := checker.Check("admin", tt.resource, tt.action)
			if !admin.Known || !admin.Allowed || admin.Reason != MonitorPermissionReasonRoleAllowed {
				t.Fatalf("admin %s:%s = known:%v allowed:%v reason:%s; want allowed",
					tt.resource, tt.action, admin.Known, admin.Allowed, admin.Reason)
			}
		})
	}
}

func TestMonitorPermissionAIOpsSessionAndActionDefaults(t *testing.T) {
	checker := NewMonitorPermissionChecker()

	for _, action := range []string{"read", "create"} {
		result := checker.Check("user", "aiops.session", action)
		if !result.Known || !result.Allowed || result.Reason != MonitorPermissionReasonRoleAllowed {
			t.Fatalf("user aiops.session %s = known:%v allowed:%v reason:%s; want allowed",
				action, result.Known, result.Allowed, result.Reason)
		}
	}

	action := checker.Check("user", "aiops.action", "execute")
	if !action.Known || action.Allowed || action.Reason != MonitorPermissionReasonRoleDenied {
		t.Fatalf("user aiops.action execute = known:%v allowed:%v reason:%s; want denied",
			action.Known, action.Allowed, action.Reason)
	}
}

func TestMonitorPermissionMatrixIncludesCMDBAndAIOpsResources(t *testing.T) {
	checker := NewMonitorPermissionChecker()
	matrix := checker.MatrixForRole("admin")
	expected := map[string]string{
		"cmdb.model":     "read",
		"cmdb.attribute": "read",
		"cmdb.instance":  "read",
		"cmdb.import":    "import",
		"cmdb.relation":  "read",
		"cmdb.stats":     "read",
		"cmdb.approval":  "approve",
		"cmdb.terminal":  "open",
		"cmdb.command":   "exec",
		"cmdb.file":      "upload",
		"cmdb.database":  "test",
		"aiops.session":  "create",
		"aiops.action":   "execute",
	}

	for resource, action := range expected {
		actions, ok := matrix[resource]
		if !ok {
			t.Fatalf("matrix missing resource %s", resource)
		}
		if !actions[action] {
			t.Fatalf("matrix %s:%s = false; want true", resource, action)
		}
	}
}
