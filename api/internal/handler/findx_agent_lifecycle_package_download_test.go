package handler

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
)

func TestPackageDownloadBlocksProductionLikeManifestArtifact(t *testing.T) {
	gin.SetMode(gin.TestMode)
	repo := t.TempDir()
	writePackageRepositoryFile(t, repo, "artifacts/host-collector-linux.tar.gz", []byte("findx-agent-package"))
	writePackageRepositoryFile(t, repo, "checksums/host-collector-linux.sha256", []byte("8c5a\n"))
	writePackageRepositoryFile(t, repo, "signatures/host-collector-linux.sha256.sig", []byte("sig"))
	writePackageRepositoryFile(t, repo, "keys/runtime.pub", []byte("pub"))
	writeProductionPackageRepositoryManifest(t, repo, packageRepositoryManifest{
		Artifacts: []packageRepositoryManifestArtifact{
			{
				ID:                "host-collector",
				Artifact:          "artifacts/host-collector-linux.tar.gz",
				OS:                "linux",
				ChecksumFile:      "checksums/host-collector-linux.sha256",
				ChecksumSignature: "signatures/host-collector-linux.sha256.sig",
				PublicKey:         "keys/runtime.pub",
			},
		},
	})
	t.Setenv("FINDX_AGENT_PACKAGE_REPOSITORY_ROOT", repo)

	w := performAgentPackageDownloadGet(
		"/api/v1/findx-agents/package-downloads/host-collector?artifact=artifacts%2Fhost-collector-linux.tar.gz&token=secret-token",
		"host-collector",
	)
	assertPackageDownloadBlocked(t, w)
}

func TestPackageDownloadBlocksChecksumSignatureAndPublicKeyRefsWithoutTrustChain(t *testing.T) {
	gin.SetMode(gin.TestMode)
	repo := t.TempDir()
	writeInstallerDownloadPackageEvidenceFiles(t, repo)
	writeProductionPackageRepositoryManifest(t, repo, installerDownloadPackageManifest())
	t.Setenv("FINDX_AGENT_PACKAGE_REPOSITORY_ROOT", repo)

	for _, tt := range []struct {
		name string
		ref  string
	}{
		{name: "checksum", ref: "checksums/host-collector-linux.sha256"},
		{name: "signature", ref: "signatures/host-collector-linux.sha256.sig"},
		{name: "public key", ref: "keys/runtime-test.pub"},
	} {
		t.Run(tt.name, func(t *testing.T) {
			w := performAgentPackageDownloadGet("/api/v1/findx-agents/package-downloads/host-collector?artifact="+tt.ref, "host-collector")
			assertPackageDownloadBlocked(t, w)
		})
	}
}

func TestPackageDownloadExposesMissingSignatureVerificationBlocker(t *testing.T) {
	gin.SetMode(gin.TestMode)
	repo := t.TempDir()
	writeInstallerDownloadPackageEvidenceFiles(t, repo)
	writeTrustChainEvidenceFiles(t, repo)
	manifest := installerDownloadReadyManifest()
	addTrustChainEvidenceRefs(&manifest)
	manifest.SignatureVerificationRef = "trust/missing-signature-verification.json"
	writeProductionPackageRepositoryManifest(t, repo, manifest)
	t.Setenv("FINDX_AGENT_PACKAGE_REPOSITORY_ROOT", repo)

	w := performAgentPackageDownloadGet(
		"/api/v1/findx-agents/package-downloads/host-collector?artifact=artifacts%2Fhost-collector-linux.tar.gz&token=secret-token",
		"host-collector",
	)

	assertPackageDownloadBlocked(t, w)
	for _, want := range []string{
		"TRUST_CHAIN_PENDING",
		"TRUST_CHAIN_SIGNATURE_VERIFICATION_MISSING",
	} {
		if !strings.Contains(w.Body.String(), want) {
			t.Fatalf("blocked package download should include %q, body=%s", want, w.Body.String())
		}
	}
}

func TestPackageDownloadBlocksTestOnlyRepositoryEvidence(t *testing.T) {
	gin.SetMode(gin.TestMode)
	repo := t.TempDir()
	writeInstallerDownloadPackageEvidenceFiles(t, repo)
	writePackageRepositoryManifest(t, repo, installerDownloadPackageManifest())
	t.Setenv("FINDX_AGENT_PACKAGE_REPOSITORY_ROOT", repo)

	w := performAgentPackageDownloadGet("/api/v1/findx-agents/package-downloads/host-collector?artifact=artifacts%2Fhost-collector-linux.tar.gz", "host-collector")
	assertPackageDownloadBlocked(t, w)
	assertPackageDownloadBlockers(t, w,
		"PACKAGE_REPOSITORY_TEST_ONLY",
		"SIGNATURE_TEST_ONLY",
		"TRUST_CHAIN_PENDING",
	)
}

func TestPackageDownloadRejectsUnsafeAndMissingEvidence(t *testing.T) {
	gin.SetMode(gin.TestMode)
	for _, tc := range packageDownloadRejectCases() {
		t.Run(tc.name, func(t *testing.T) {
			assertPackageDownloadRejectCase(t, tc)
		})
	}
}

type packageDownloadRejectCase struct {
	name       string
	prepare    func(t *testing.T, repo string)
	path       string
	wantBlocks []string
}

func packageDownloadRejectCases() []packageDownloadRejectCase {
	return []packageDownloadRejectCase{
		{
			name: "unsafe query path",
			prepare: func(t *testing.T, repo string) {
				writeInstallerDownloadPackageEvidence(t, repo)
			},
			path:       "/api/v1/findx-agents/package-downloads/host-collector?artifact=..%2Fsecret%2Fpackage.tar.gz&password=secret-password",
			wantBlocks: []string{"PACKAGE_REPOSITORY_REF_INVALID"},
		},
		{
			name: "missing manifest",
			prepare: func(t *testing.T, repo string) {
				writePackageRepositoryTestFile(t, repo, "artifacts/host-collector-linux.tar.gz")
			},
			path:       "/api/v1/findx-agents/package-downloads/host-collector?artifact=artifacts%2Fhost-collector-linux.tar.gz&cookie=secret-cookie",
			wantBlocks: []string{"PACKAGE_REPOSITORY_MANIFEST_MISSING"},
		},
		{
			name: "missing artifact",
			prepare: func(t *testing.T, repo string) {
				writePackageRepositoryTestFile(t, repo, "checksums/host-collector-linux.sha256")
				writePackageRepositoryTestFile(t, repo, "signatures/host-collector-linux.sha256.sig")
				writePackageRepositoryTestFile(t, repo, "keys/runtime-test.pub")
				writeProductionPackageRepositoryManifest(t, repo, installerDownloadPackageManifest())
			},
			path:       "/api/v1/findx-agents/package-downloads/host-collector?artifact=artifacts%2Fhost-collector-linux.tar.gz&private_key=secret-key",
			wantBlocks: []string{"PACKAGE_REPOSITORY_ARTIFACT_MISSING"},
		},
		{
			name: "unknown package",
			prepare: func(t *testing.T, repo string) {
				writeInstallerDownloadPackageEvidence(t, repo)
			},
			path:       "/api/v1/findx-agents/package-downloads/unknown?artifact=artifacts%2Fhost-collector-linux.tar.gz&token=secret-token",
			wantBlocks: []string{"PACKAGE_NOT_FOUND"},
		},
	}
}

func assertPackageDownloadRejectCase(t *testing.T, tc packageDownloadRejectCase) {
	t.Helper()
	repo := t.TempDir()
	tc.prepare(t, repo)
	t.Setenv("FINDX_AGENT_PACKAGE_REPOSITORY_ROOT", repo)
	w := performAgentPackageDownloadGet(tc.path, packageNameFromPath(tc.path))
	assertPackageDownloadBlocked(t, w)
	assertPackageDownloadBlockers(t, w, tc.wantBlocks...)
}

func performAgentPackageDownloadGet(path, packageParam string) *httptest.ResponseRecorder {
	req := httptest.NewRequest(http.MethodGet, path, nil)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Params = gin.Params{{Key: "package", Value: packageParam}}
	c.Request = req
	DownloadFindXAgentPackageArtifact(c)
	return w
}

func packageNameFromPath(path string) string {
	parts := strings.Split(strings.TrimPrefix(path, "/api/v1/findx-agents/package-downloads/"), "?")
	return parts[0]
}

func assertPackageDownloadBlocked(t *testing.T, w *httptest.ResponseRecorder) {
	t.Helper()
	body := w.Body.String()
	if w.Code != http.StatusConflict || !strings.Contains(body, "PENDING") {
		t.Fatalf("package download should be blocked, got %d body=%s", w.Code, body)
	}
	for _, forbidden := range []string{"secret-token", "secret-password", "secret-cookie", "secret-key"} {
		if strings.Contains(body, forbidden) {
			t.Fatalf("package download block response leaked sensitive query %q: %s", forbidden, body)
		}
	}
	for _, forbidden := range []string{"queued", "running", "succeeded", "success", "applied", "rolled-back"} {
		if strings.Contains(body, forbidden) {
			t.Fatalf("package download block response must not include fake success state %q: %s", forbidden, body)
		}
	}
	var payload struct {
		Safe struct {
			CredentialEcho bool `json:"credential_echo"`
			PathTraversal  bool `json:"path_traversal"`
			SafeToRetry    bool `json:"safe_to_retry"`
		} `json:"safe"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &payload); err != nil {
		t.Fatalf("package download block response should be JSON: %v body=%s", err, body)
	}
	if payload.Safe.CredentialEcho || payload.Safe.PathTraversal || payload.Safe.SafeToRetry {
		t.Fatalf("package download safe flags must stay false for blocked contracts: %s", body)
	}
}

func assertPackageDownloadBlockers(t *testing.T, w *httptest.ResponseRecorder, wants ...string) {
	t.Helper()
	var payload struct {
		Blockers         []string `json:"blockers"`
		MissingContracts []string `json:"missing_contracts"`
		Safe             struct {
			SafeToRetry bool `json:"safe_to_retry"`
		} `json:"safe"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &payload); err != nil {
		t.Fatalf("package download block response should be JSON: %v body=%s", err, w.Body.String())
	}
	if payload.Safe.SafeToRetry {
		t.Fatalf("package download block response must not mark contract gaps safe to retry: %s", w.Body.String())
	}
	for _, want := range wants {
		if !stringSliceContains(payload.Blockers, want) {
			t.Fatalf("package download blockers should include %q, got blockers=%v missing=%v body=%s", want, payload.Blockers, payload.MissingContracts, w.Body.String())
		}
		if !stringSliceContains(payload.MissingContracts, want) {
			t.Fatalf("package download missing_contracts should include %q, got blockers=%v missing=%v body=%s", want, payload.Blockers, payload.MissingContracts, w.Body.String())
		}
	}
}

func stringSliceContains(values []string, want string) bool {
	for _, value := range values {
		if value == want {
			return true
		}
	}
	return false
}

func writePackageRepositoryFile(t *testing.T, repo, ref string, content []byte) {
	t.Helper()
	path := filepath.Join(repo, ref)
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, content, 0o644); err != nil {
		t.Fatal(err)
	}
}

func writeProductionPackageRepositoryManifest(t *testing.T, repo string, manifest packageRepositoryManifest) {
	t.Helper()
	manifest.Repository = "findx-agent-runtime-local"
	manifest.Status = "production_signature_ready"
	manifest.SignatureScope = "production-gpg"
	data, err := json.Marshal(manifest)
	if err != nil {
		t.Fatal(err)
	}
	manifestPath := filepath.Join(repo, packageRepositoryManifestPath)
	if err := os.MkdirAll(filepath.Dir(manifestPath), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(manifestPath, data, 0o644); err != nil {
		t.Fatal(err)
	}
}
