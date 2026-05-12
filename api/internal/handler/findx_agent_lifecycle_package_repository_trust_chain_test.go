package handler

import (
	"net/http"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
)

func TestTrustChainBlocksProductionLikePackageDownload(t *testing.T) {
	gin.SetMode(gin.TestMode)
	repo := t.TempDir()
	writeInstallerDownloadPackageEvidenceFiles(t, repo)
	writeProductionPackageRepositoryManifest(t, repo, installerDownloadPackageManifest())
	t.Setenv("FINDX_AGENT_PACKAGE_REPOSITORY_ROOT", repo)

	w := performAgentPackageDownloadGet(
		"/api/v1/findx-agents/package-downloads/host-collector?artifact=artifacts%2Fhost-collector-linux.tar.gz",
		"host-collector",
	)

	if w.Code != http.StatusConflict {
		t.Fatalf("production-like package fixture must stay blocked, got %d body=%s", w.Code, w.Body.String())
	}
	for _, want := range []string{"BLOCKED_BY_CONTRACT", "package_repository_artifact", "signature", "public_key"} {
		if !strings.Contains(w.Body.String(), want) {
			t.Fatalf("blocked package download should include %q, body=%s", want, w.Body.String())
		}
	}
}

func TestTrustChainBlocksProductionLikeInstallerDownload(t *testing.T) {
	gin.SetMode(gin.TestMode)
	repo := t.TempDir()
	writeInstallerDownloadPackageEvidenceFiles(t, repo)
	writeInstallerDownloadReadyEnvironmentFiles(t, repo)
	writeProductionPackageRepositoryManifest(t, repo, installerDownloadReadyManifest())
	t.Setenv("FINDX_AGENT_PACKAGE_REPOSITORY_ROOT", repo)

	w := performAgentLifecycleGet(
		"/api/v1/findx-agents/installers/linux.sh?package=host-collector",
		DownloadFindXAgentLinuxInstaller,
	)

	body := w.Body.String()
	if w.Code != http.StatusConflict || !strings.Contains(body, "BLOCKED_BY_CONTRACT") {
		t.Fatalf("production-like installer fixture must stay blocked, got %d body=%s", w.Code, body)
	}
	for _, want := range []string{
		"TRUST_CHAIN_VERIFICATION_RECEIPT_MISSING",
		"TRUST_CHAIN_PRODUCTION_KEY_MISSING",
		"TRUST_CHAIN_CHECKSUM_VERIFICATION_MISSING",
		"TRUST_CHAIN_SIGNATURE_VERIFICATION_MISSING",
		"TRUST_CHAIN_BLOCKED_BY_CONTRACT",
	} {
		if !strings.Contains(body, want) {
			t.Fatalf("installer blocker should include %q, body=%s", want, body)
		}
	}
}

func TestTrustChainEvidenceDoesNotAllowDownloadsWithoutVerifier(t *testing.T) {
	gin.SetMode(gin.TestMode)
	repo := t.TempDir()
	writeInstallerDownloadPackageEvidenceFiles(t, repo)
	writeTrustChainEvidenceFiles(t, repo)
	manifest := installerDownloadReadyManifest()
	addTrustChainEvidenceRefs(&manifest)
	writeInstallerDownloadReadyEnvironmentFiles(t, repo)
	writeProductionPackageRepositoryManifest(t, repo, manifest)
	t.Setenv("FINDX_AGENT_PACKAGE_REPOSITORY_ROOT", repo)

	artifact := performAgentPackageDownloadGet(
		"/api/v1/findx-agents/package-downloads/host-collector?artifact=artifacts%2Fhost-collector-linux.tar.gz",
		"host-collector",
	)
	assertTrustChainVerifierBlocked(t, "artifact download", artifact.Code, artifact.Body.String())

	installer := performAgentLifecycleGet(
		"/api/v1/findx-agents/installers/linux.sh?package=host-collector",
		DownloadFindXAgentLinuxInstaller,
	)
	assertTrustChainVerifierBlocked(t, "installer download", installer.Code, installer.Body.String())
}

func assertTrustChainVerifierBlocked(t *testing.T, name string, code int, body string) {
	t.Helper()
	if code != http.StatusConflict {
		t.Fatalf("%s must stay blocked without production verifier, got %d body=%s", name, code, body)
	}
	for _, want := range []string{
		"BLOCKED_BY_CONTRACT",
		"TRUST_CHAIN_VERIFICATION_NOT_IMPLEMENTED",
		"TRUST_CHAIN_BLOCKED_BY_CONTRACT",
	} {
		if !strings.Contains(body, want) {
			t.Fatalf("%s blocker should include %q, body=%s", name, want, body)
		}
	}
}

func TestTrustChainBlocksMissingSignatureVerification(t *testing.T) {
	gin.SetMode(gin.TestMode)
	repo := t.TempDir()
	writeInstallerDownloadPackageEvidenceFiles(t, repo)
	writeTrustChainEvidenceFiles(t, repo)
	manifest := installerDownloadReadyManifest()
	addTrustChainEvidenceRefs(&manifest)
	manifest.SignatureVerificationRef = "trust/missing-signature-verification.json"
	writeInstallerDownloadReadyEnvironmentFiles(t, repo)
	writeProductionPackageRepositoryManifest(t, repo, manifest)
	t.Setenv("FINDX_AGENT_PACKAGE_REPOSITORY_ROOT", repo)

	w := performAgentLifecycleGet(
		"/api/v1/findx-agents/installers/linux.sh?package=host-collector",
		DownloadFindXAgentLinuxInstaller,
	)

	body := w.Body.String()
	if w.Code != http.StatusConflict || !strings.Contains(body, "TRUST_CHAIN_SIGNATURE_VERIFICATION_MISSING") {
		t.Fatalf("missing signature verification must block, got %d body=%s", w.Code, body)
	}
}

func addTrustChainEvidenceRefs(manifest *packageRepositoryManifest) {
	manifest.TrustRootRef = "trust/root.pub"
	manifest.TrustRootFingerprintRef = "trust/root.fingerprint"
	manifest.VerificationReceiptRef = "trust/verification-receipt.json"
	manifest.ChecksumVerificationRef = "trust/checksum-verification.json"
	manifest.SignatureVerificationRef = "trust/signature-verification.json"
}

func writeTrustChainEvidenceFiles(t *testing.T, repo string) {
	t.Helper()
	for _, ref := range []string{
		"trust/root.pub",
		"trust/root.fingerprint",
		"trust/verification-receipt.json",
		"trust/checksum-verification.json",
		"trust/signature-verification.json",
	} {
		writePackageRepositoryTestFile(t, repo, ref)
	}
}
