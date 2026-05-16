package handler

import (
	"encoding/json"
	"net/http"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
)

func TestInstallerDownloadEndpointsReturnBlockedContract(t *testing.T) {
	gin.SetMode(gin.TestMode)

	for _, tt := range []struct {
		name    string
		path    string
		handler gin.HandlerFunc
	}{
		{name: "linux shell", path: "/api/v1/findx-agents/installers/linux.sh", handler: DownloadFindXAgentLinuxInstaller},
		{name: "windows powershell", path: "/api/v1/findx-agents/installers/windows.ps1", handler: DownloadFindXAgentWindowsPowerShellInstaller},
		{name: "windows batch", path: "/api/v1/findx-agents/installers/windows.bat", handler: DownloadFindXAgentWindowsBatchInstaller},
	} {
		t.Run(tt.name, func(t *testing.T) {
			w := performAgentLifecycleGet(tt.path, tt.handler)
			body := w.Body.String()
			if w.Code != http.StatusConflict {
				t.Fatalf("installer download should be blocked, got %d body=%s", w.Code, body)
			}
			for _, want := range []string{
				"PENDING",
				"package_repository",
				"signature",
				"script_manifest",
				"executor",
				"bundled_install_environment",
			} {
				if !strings.Contains(body, want) {
					t.Fatalf("blocked contract should include %q, body=%s", want, body)
				}
			}
			assertInstallerDownloadContentType(t, w.Header().Get("Content-Type"))
		})
	}
}

func TestInstallerDownloadBlocksWhenBundledEnvironmentEvidenceIsMissing(t *testing.T) {
	gin.SetMode(gin.TestMode)
	repo := t.TempDir()
	writeInstallerDownloadPackageEvidenceFiles(t, repo)
	writeInstallerDownloadBlockedEnvironmentManifest(t, repo)
	t.Setenv("FINDX_AGENT_PACKAGE_REPOSITORY_ROOT", repo)

	w := performAgentLifecycleGet("/api/v1/findx-agents/installers/linux.sh?package=host-collector", DownloadFindXAgentLinuxInstaller)
	body := w.Body.String()
	if w.Code != http.StatusConflict || !strings.Contains(body, "PENDING") {
		t.Fatalf("installer download must block when environment evidence is missing, got %d body=%s", w.Code, body)
	}
	for _, want := range []string{
		"INSTALL_ENVIRONMENT_MANIFEST_MISSING",
		"PACKAGE_REPOSITORY_REF_MISSING",
		"SIGNATURE_REF_MISSING",
		"CHECKSUM_MISSING",
		"SCRIPT_MANIFEST_REF_MISSING",
		"EXECUTOR_REF_MISSING",
		"INSTALL_COMMAND_REF_MISSING",
		"SERVICE_MANIFEST_REF_MISSING",
	} {
		if !strings.Contains(body, want) {
			t.Fatalf("blocked installer should expose environment blocker %q, body=%s", want, body)
		}
	}
	assertInstallerDownloadContentType(t, w.Header().Get("Content-Type"))
}

func TestInstallerDownloadBlocksProductionLikeReadyEnvironment(t *testing.T) {
	gin.SetMode(gin.TestMode)
	repo := t.TempDir()
	writeInstallerDownloadPackageEvidenceFiles(t, repo)
	writeInstallerDownloadReadyEnvironmentManifest(t, repo)
	t.Setenv("FINDX_AGENT_PACKAGE_REPOSITORY_ROOT", repo)

	for _, tt := range []struct {
		name    string
		path    string
		handler gin.HandlerFunc
	}{
		{
			name:    "linux shell",
			path:    "/api/v1/findx-agents/installers/linux.sh?package=host-collector&token=secret-token",
			handler: DownloadFindXAgentLinuxInstaller,
		},
		{
			name:    "windows powershell",
			path:    "/api/v1/findx-agents/installers/windows.ps1?package=host-collector&password=secret-password",
			handler: DownloadFindXAgentWindowsPowerShellInstaller,
		},
		{
			name:    "windows batch",
			path:    "/api/v1/findx-agents/installers/windows.bat?package=host-collector&cookie=secret-cookie",
			handler: DownloadFindXAgentWindowsBatchInstaller,
		},
	} {
		t.Run(tt.name, func(t *testing.T) {
			w := performAgentLifecycleGet(tt.path, tt.handler)
			body := w.Body.String()
			if w.Code != http.StatusConflict || !strings.Contains(body, "PENDING") {
				t.Fatalf("production-like installer fixture must block, got %d body=%s", w.Code, body)
			}
			for _, want := range []string{"TRUST_CHAIN_PENDING", "TRUST_CHAIN_VERIFICATION_RECEIPT_MISSING"} {
				if !strings.Contains(body, want) {
					t.Fatalf("installer blocker should include %q, body=%s", want, body)
				}
			}
			assertInstallerScriptDoesNotEchoSensitiveQuery(t, body)
			assertInstallerDownloadDoesNotExposeFakeState(t, body)
			assertInstallerDownloadContentType(t, w.Header().Get("Content-Type"))
		})
	}
}

func TestInstallerDownloadDoesNotEchoSensitiveQuery(t *testing.T) {
	gin.SetMode(gin.TestMode)

	path := "/api/v1/findx-agents/installers/linux.sh?" + strings.Join([]string{
		"package=agent-core",
		"token=%3CTOKEN%3E",
		"credential_ref=%3CCREDENTIAL_REF%3E",
		"password=%3CPASSWORD%3E",
		"session=%3CSESSION%3E",
		"cookie=%3CCOOKIE%3E",
		"private_key=%3CPRIVATE_KEY%3E",
		"DSN=%3CDB_DSN%3E",
	}, "&")
	w := performAgentLifecycleGet(path, DownloadFindXAgentLinuxInstaller)
	body := w.Body.String()

	if w.Code != http.StatusConflict || !strings.Contains(body, "PENDING") {
		t.Fatalf("installer download should stay blocked, got %d body=%s", w.Code, body)
	}
	for _, forbidden := range []string{
		"agent-core",
		"<TOKEN>",
		"<CREDENTIAL_REF>",
		"<PASSWORD>",
		"<SESSION>",
		"<COOKIE>",
		"<PRIVATE_KEY>",
		"<DB_DSN>",
		"%3CTOKEN%3E",
		"%3CCREDENTIAL_REF%3E",
		"%3CPASSWORD%3E",
		"%3CSESSION%3E",
		"%3CCOOKIE%3E",
		"%3CPRIVATE_KEY%3E",
		"%3CDB_DSN%3E",
	} {
		if strings.Contains(body, forbidden) {
			t.Fatalf("installer download must not echo sensitive query %q: %s", forbidden, body)
		}
	}
	assertInstallerDownloadContentType(t, w.Header().Get("Content-Type"))
	assertInstallerDownloadDoesNotExposeFakeState(t, body)
}

func TestInstallerDownloadBlocksTestOnlyRepositoryEvidence(t *testing.T) {
	gin.SetMode(gin.TestMode)
	repo := t.TempDir()
	writeInstallerDownloadPackageEvidenceFiles(t, repo)
	writePackageRepositoryManifest(t, repo, installerDownloadPackageManifest())
	t.Setenv("FINDX_AGENT_PACKAGE_REPOSITORY_ROOT", repo)

	w := performAgentLifecycleGet("/api/v1/findx-agents/installers/linux.sh?package=host-collector", DownloadFindXAgentLinuxInstaller)
	body := w.Body.String()
	if w.Code != http.StatusConflict || !strings.Contains(body, "PENDING") {
		t.Fatalf("test-only repository must not return executable installer, got %d body=%s", w.Code, body)
	}
	for _, want := range []string{"PACKAGE_REPOSITORY_TEST_ONLY", "SIGNATURE_TEST_ONLY"} {
		if !strings.Contains(body, want) {
			t.Fatalf("test-only installer blocker should include %q, body=%s", want, body)
		}
	}
	assertInstallerDownloadContentType(t, w.Header().Get("Content-Type"))
}

func TestInstallerDownloadRejectsUnsafeRepositoryReferences(t *testing.T) {
	gin.SetMode(gin.TestMode)

	for _, tt := range []struct {
		name     string
		artifact packageRepositoryManifestArtifact
	}{
		{name: "unsafe artifact", artifact: packageRepositoryManifestArtifact{ID: "host-collector", Artifact: "../host-collector.tar.gz", OS: "linux", ChecksumFile: "checksums/host-collector-linux.sha256", ChecksumSignature: "signatures/host-collector-linux.sha256.sig", PublicKey: "keys/runtime.pub"}},
		{name: "unsafe checksum", artifact: packageRepositoryManifestArtifact{ID: "host-collector", Artifact: "artifacts/host-collector-linux.tar.gz", OS: "linux", ChecksumFile: "../host-collector.sha256", ChecksumSignature: "signatures/host-collector-linux.sha256.sig", PublicKey: "keys/runtime.pub"}},
		{name: "unsafe signature", artifact: packageRepositoryManifestArtifact{ID: "host-collector", Artifact: "artifacts/host-collector-linux.tar.gz", OS: "linux", ChecksumFile: "checksums/host-collector-linux.sha256", ChecksumSignature: "../host-collector.sig", PublicKey: "keys/runtime.pub"}},
		{name: "unsafe public key", artifact: packageRepositoryManifestArtifact{ID: "host-collector", Artifact: "artifacts/host-collector-linux.tar.gz", OS: "linux", ChecksumFile: "checksums/host-collector-linux.sha256", ChecksumSignature: "signatures/host-collector-linux.sha256.sig", PublicKey: "../runtime.pub"}},
	} {
		t.Run(tt.name, func(t *testing.T) {
			repo := t.TempDir()
			writeInstallerDownloadPackageEvidenceFiles(t, repo)
			writeInstallerDownloadReadyEnvironmentFiles(t, repo)
			manifest := installerDownloadReadyManifest()
			manifest.Artifacts = []packageRepositoryManifestArtifact{tt.artifact}
			writeProductionPackageRepositoryManifest(t, repo, manifest)
			t.Setenv("FINDX_AGENT_PACKAGE_REPOSITORY_ROOT", repo)

			w := performAgentLifecycleGet("/api/v1/findx-agents/installers/linux.sh?package=host-collector&token=secret-token", DownloadFindXAgentLinuxInstaller)
			body := w.Body.String()
			if w.Code != http.StatusConflict || !strings.Contains(body, "PENDING") {
				t.Fatalf("unsafe repository ref must block installer, got %d body=%s", w.Code, body)
			}
			assertInstallerScriptDoesNotEchoSensitiveQuery(t, body)
		})
	}
}

func writeInstallerDownloadPackageEvidence(t *testing.T, repo string) {
	t.Helper()
	writeInstallerDownloadPackageEvidenceFiles(t, repo)
	writeProductionPackageRepositoryManifest(t, repo, installerDownloadPackageManifest())
}

func writeInstallerDownloadPackageEvidenceFiles(t *testing.T, repo string) {
	t.Helper()
	for _, ref := range []string{
		"artifacts/host-collector-linux.tar.gz",
		"artifacts/host-collector-windows.zip",
		"checksums/host-collector-linux.sha256",
		"checksums/host-collector-windows.sha256",
		"signatures/host-collector-linux.sha256.sig",
		"signatures/host-collector-windows.sha256.sig",
		"keys/runtime-test.pub",
	} {
		writePackageRepositoryTestFile(t, repo, ref)
	}
}

func installerDownloadPackageManifest() packageRepositoryManifest {
	return packageRepositoryManifest{
		Artifacts: []packageRepositoryManifestArtifact{
			{
				ID:                "host-collector",
				Artifact:          "artifacts/host-collector-linux.tar.gz",
				OS:                "linux",
				ChecksumFile:      "checksums/host-collector-linux.sha256",
				ChecksumSignature: "signatures/host-collector-linux.sha256.sig",
				PublicKey:         "keys/runtime-test.pub",
			},
			{
				ID:                "host-collector",
				Artifact:          "artifacts/host-collector-windows.zip",
				OS:                "windows",
				ChecksumFile:      "checksums/host-collector-windows.sha256",
				ChecksumSignature: "signatures/host-collector-windows.sha256.sig",
				PublicKey:         "keys/runtime-test.pub",
			},
		},
	}
}

func writeInstallerDownloadBlockedEnvironmentManifest(t *testing.T, repo string) {
	t.Helper()
	manifest := installerDownloadPackageManifest()
	writeProductionPackageRepositoryManifest(t, repo, manifest)
}

func writeInstallerDownloadReadyEnvironmentManifest(t *testing.T, repo string) {
	t.Helper()
	writeInstallerDownloadReadyEnvironmentFiles(t, repo)
	writeProductionPackageRepositoryManifest(t, repo, installerDownloadReadyManifest())
}

func writeInstallerDownloadReadyEnvironmentFiles(t *testing.T, repo string) {
	t.Helper()
	for _, ref := range []string{
		"manifests/install.sh",
		"manifests/uninstall.sh",
		"manifests/config.sh",
		"manifests/plugin.sh",
		"manifests/service.yaml",
		"manifests/rollback.yaml",
		"manifests/data-arrival-validator.sh",
		"security/common/package_repository_ref.txt",
		"security/common/signature_ref.txt",
		"security/common/checksum.txt",
		"security/common/script_manifest_ref.txt",
		"security/common/executor_ref.txt",
		"security/common/install_root_policy_ref.txt",
		"security/common/uninstall_manifest_ref.txt",
		"security/common/receiver_endpoint_ref.txt",
		"security/common/audit_ref.txt",
		"security/common/evidence_chain_ref.txt",
		"security/linux/safety_policy_path.txt",
		"security/linux/runner.txt",
		"security/linux/ssh_host_key.txt",
		"security/linux/ssh_fingerprint.txt",
		"security/linux/systemd_unit_ref.txt",
		"security/linux/curl_installer_ref.txt",
		"security/linux/env_template_ref.txt",
		"security/windows/windows_installer_ref.txt",
	} {
		writePackageRepositoryTestFile(t, repo, ref)
	}
	for _, osName := range []string{"linux", "windows"} {
		for _, tool := range packageRepositoryCriticalTools {
			writePackageRepositoryTestFile(t, repo, "tools/"+osName+"/"+tool+".ref")
		}
	}
}

func installerDownloadReadyManifest() packageRepositoryManifest {
	manifest := installerDownloadPackageManifest()
	manifest.PackageRepositoryRef = "security/common/package_repository_ref.txt"
	manifest.SignatureRef = "security/common/signature_ref.txt"
	manifest.Checksum = "security/common/checksum.txt"
	manifest.ScriptManifestRef = "security/common/script_manifest_ref.txt"
	manifest.ExecutorRef = "security/common/executor_ref.txt"
	manifest.SafetyPolicyPath = "security/linux/safety_policy_path.txt"
	manifest.Runner = "security/linux/runner.txt"
	manifest.SSHHostKey = "security/linux/ssh_host_key.txt"
	manifest.SSHFingerprint = "security/linux/ssh_fingerprint.txt"
	manifest.SystemdUnitRef = "security/linux/systemd_unit_ref.txt"
	manifest.CurlInstallerRef = "security/linux/curl_installer_ref.txt"
	manifest.EnvTemplateRef = "security/linux/env_template_ref.txt"
	manifest.WindowsInstallerRef = "security/windows/windows_installer_ref.txt"
	manifest.InstallRootPolicyRef = "security/common/install_root_policy_ref.txt"
	manifest.UninstallManifestRef = "security/common/uninstall_manifest_ref.txt"
	manifest.ReceiverEndpointRef = "security/common/receiver_endpoint_ref.txt"
	manifest.InstallCommandRef = "manifests/install.sh"
	manifest.UninstallCommandRef = "manifests/uninstall.sh"
	manifest.ConfigCommandRef = "manifests/config.sh"
	manifest.PluginCommandRef = "manifests/plugin.sh"
	manifest.ServiceManifestRef = "manifests/service.yaml"
	manifest.RollbackManifestRef = "manifests/rollback.yaml"
	manifest.DataArrivalValidatorRef = "manifests/data-arrival-validator.sh"
	manifest.AuditRef = "security/common/audit_ref.txt"
	manifest.EvidenceChainRef = "security/common/evidence_chain_ref.txt"
	manifest.Method = "linux-curl/windows-powershell/windows-cmd"
	manifest.OS = "linux/windows"
	manifest.RequiredTools = installerDownloadToolEvidence(false)
	manifest.BundledTools = installerDownloadToolEvidence(true)
	return manifest
}

func installerDownloadToolEvidence(bundled bool) []packageRepositoryToolEvidence {
	tools := []packageRepositoryToolEvidence{}
	for _, osName := range []string{"linux", "windows"} {
		for _, name := range packageRepositoryCriticalTools {
			tools = append(tools, packageRepositoryToolEvidence{
				Name:        name,
				OS:          osName,
				Arch:        "amd64",
				EvidenceRef: "tools/" + osName + "/" + name + ".ref",
				Bundled:     bundled,
			})
		}
	}
	return tools
}

func assertInstallerScriptDoesNotEchoSensitiveQuery(t *testing.T, body string) {
	t.Helper()
	for _, forbidden := range []string{"secret-token", "secret-password", "secret-cookie", "<TOKEN>", "<PASSWORD>", "<COOKIE>", "<PRIVATE_KEY>", "<DB_DSN>"} {
		if strings.Contains(body, forbidden) {
			t.Fatalf("installer script must not echo sensitive query %q: %s", forbidden, body)
		}
	}
}

func assertInstallerDownloadDoesNotExposeFakeState(t *testing.T, body string) {
	t.Helper()
	for _, forbidden := range []string{"queued", "running", "succeeded", "success", "applied", "rolled-back"} {
		if strings.Contains(body, forbidden) {
			t.Fatalf("installer block response must not include fake execution state %q: %s", forbidden, body)
		}
	}
	var payload struct {
		Safe struct {
			ExecutableScript bool `json:"executable_script"`
			CredentialEcho   bool `json:"credential_echo"`
			SafeToRetry      bool `json:"safe_to_retry"`
		} `json:"safe"`
	}
	if err := json.Unmarshal([]byte(body), &payload); err != nil {
		t.Fatalf("installer block response should be JSON: %v body=%s", err, body)
	}
	if payload.Safe.ExecutableScript || payload.Safe.CredentialEcho || payload.Safe.SafeToRetry {
		t.Fatalf("installer safe flags must stay false for blocked contracts: %s", body)
	}
}

func assertInstallerDownloadContentType(t *testing.T, contentType string) {
	t.Helper()
	if !strings.Contains(contentType, "application/json") {
		t.Fatalf("installer download should return JSON, got content-type %q", contentType)
	}
	for _, executableType := range []string{
		"text/x-shellscript",
		"application/x-sh",
		"application/x-powershell",
		"application/x-bat",
	} {
		if strings.Contains(strings.ToLower(contentType), executableType) {
			t.Fatalf("installer download must not return executable script content-type %q", contentType)
		}
	}
}
