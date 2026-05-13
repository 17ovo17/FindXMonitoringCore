package handler

import (
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
)

func TestFindXAgentPackageActionsPersistBlockedTasks(t *testing.T) {
	gin.SetMode(gin.TestMode)
	for _, tt := range []struct {
		name       string
		action     string
		missingRef string
	}{
		{name: "sync repository", action: "sync_package_repository", missingRef: "manifest_ref"},
		{name: "publish package", action: "publish_package", missingRef: "release_manifest_ref"},
		{name: "download package", action: "download_package", missingRef: "artifact_ref"},
		{name: "verify signature", action: "verify_package_signature", missingRef: "verifier_ref"},
	} {
		t.Run(tt.name, func(t *testing.T) {
			resetAgentLifecycleRecordsForTest(t)
			body := strings.NewReader(`{
				"action":"` + tt.action + `",
				"package_id":"host-collector",
				"target_ids":["package-repository-control-plane"],
				"metadata":{
					"package_id":"host-collector",
					"target_os":"control-plane",
					"transport":"local-control-plane",
					"audit_ref":"package-action-audit"
				}
			}`)
			w := performAgentLifecyclePost("/api/v1/findx-agents/tasks", body, CreateFindXAgentTask)
			payload := decodeAgentTaskResponse(t, w)

			if w.Code != 409 {
				t.Fatalf("%s should stay blocked, got %d body=%s", tt.action, w.Code, w.Body.String())
			}
			if payload.Data.Action != tt.action || payload.Data.Status != "blocked" {
				t.Fatalf("%s should persist blocked task, got %#v", tt.action, payload.Data)
			}
			if !strings.Contains(payload.Data.Blocker, "BLOCKED_BY_CONTRACT") ||
				!strings.Contains(payload.Data.Blocker, tt.missingRef) {
				t.Fatalf("%s blocker should name missing %s: %s", tt.action, tt.missingRef, payload.Data.Blocker)
			}
			for _, forbidden := range []string{`"status":"queued"`, `"status":"running"`, `"status":"succeeded"`} {
				if strings.Contains(w.Body.String(), forbidden) {
					t.Fatalf("%s must not fake execution state %s: %s", tt.action, forbidden, w.Body.String())
				}
			}
			assertNoCredentialEcho(t, w.Body.String())
		})
	}
}
