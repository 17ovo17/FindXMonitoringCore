package handler

import (
	"fmt"
	"net/http"
	"net/url"
	"sort"
	"strings"

	"ai-workbench-api/internal/model"

	"github.com/gin-gonic/gin"
)

const installerDownloadBlocker = agentBlocked + ": package repo, signature, script manifest, and executor contracts are not open"

func DownloadFindXAgentLinuxInstaller(c *gin.Context) {
	downloadFindXAgentInstaller(c, "linux.sh", "linux-shell", "linux")
}

func DownloadFindXAgentWindowsPowerShellInstaller(c *gin.Context) {
	downloadFindXAgentInstaller(c, "windows.ps1", "windows-powershell", "windows")
}

func DownloadFindXAgentWindowsBatchInstaller(c *gin.Context) {
	downloadFindXAgentInstaller(c, "windows.bat", "windows-cmd", "windows")
}

type installerPackageEvidence struct {
	PackageID      string
	ArtifactRef    string
	ChecksumRef    string
	SignatureRef   string
	PublicKeyRef   string
	SignatureScope string
	ToolRefs       map[string]string
}

func downloadFindXAgentInstaller(c *gin.Context, installer, platform, osName string) {
	evidence, environment, ok := findInstallerPackageEvidence(c.Query("package"), osName)
	if !ok {
		blockFindXAgentInstallerDownload(c, installer, platform, environment)
		return
	}
	script := renderInstallerScript(installer, platform, evidence)
	c.Data(http.StatusOK, installerContentType(installer), []byte(script))
}

func blockFindXAgentInstallerDownload(c *gin.Context, installer, platform string, environment model.FindXAgentInstallEnvironment) {
	c.JSON(http.StatusConflict, gin.H{
		"code":                http.StatusConflict,
		"error":               installerDownloadBlocker,
		"status":              "blocked",
		"installer":           installer,
		"platform":            platform,
		"install_environment": environment,
		"missing_contracts": []string{
			"package_repository",
			"signature",
			"script_manifest",
			"executor",
			"bundled_install_environment",
		},
		"safe": gin.H{
			"executable_script": false,
			"credential_echo":   false,
		},
	})
}

func findInstallerPackageEvidence(requestedPackage, osName string) (installerPackageEvidence, model.FindXAgentInstallEnvironment, bool) {
	packageID := strings.TrimSpace(requestedPackage)
	if packageID == "" {
		packageID = "agent-core"
	}
	if _, ok := findAgentPackage(packageID); !ok {
		return installerPackageEvidence{}, defaultBlockedInstallerEnvironment(), false
	}
	environment := defaultBlockedInstallerEnvironment()
	for _, root := range agentPackageRepositoryRoots() {
		manifest, ok := loadPackageRepositoryManifest(root)
		if !ok || isTestOnlyPackageRepositoryManifest(manifest) {
			if ok {
				environment = blockedInstallEnvironment([]string{
					"PACKAGE_REPOSITORY_TEST_ONLY",
					"SIGNATURE_TEST_ONLY",
					"BUNDLED_INSTALL_ENVIRONMENT_CONTRACT_MISSING",
				})
			}
			continue
		}
		for _, artifact := range manifest.Artifacts {
			if !installerArtifactMatches(artifact, packageID, osName) {
				continue
			}
			environment = packageRepositoryInstallEnvironmentForOS(root, manifest, osName)
			if environment.Status != "ready" {
				continue
			}
			evidence, ok := installerEvidenceFromArtifact(manifest, artifact)
			if ok && packageRepositoryTrustChainVerified(root, manifest, artifact) {
				evidence.ToolRefs = installerToolRefs(environment.Tools)
				return evidence, environment, true
			}
			environment = appendInstallerEnvironmentBlockers(
				environment,
				packageRepositoryTrustChainBlockers(root, manifest, artifact)...,
			)
		}
	}
	return installerPackageEvidence{}, environment, false
}

func defaultBlockedInstallerEnvironment() model.FindXAgentInstallEnvironment {
	return blockedInstallEnvironment([]string{
		"PACKAGE_REPOSITORY_MISSING",
		"SIGNATURE_MISSING",
		"SCRIPT_MANIFEST_REF_MISSING",
		"EXECUTOR_REF_MISSING",
		"BUNDLED_INSTALL_ENVIRONMENT_CONTRACT_MISSING",
	})
}

func appendInstallerEnvironmentBlockers(environment model.FindXAgentInstallEnvironment, blockers ...string) model.FindXAgentInstallEnvironment {
	environment.Status = "blocked"
	environment.Blockers = uniquePackageRepositoryBlockers(append(environment.Blockers, blockers...))
	return environment
}

func installerArtifactMatches(artifact packageRepositoryManifestArtifact, packageID, osName string) bool {
	if packageRepositoryArtifactID(artifact) != packageID {
		return false
	}
	artifactOS := strings.ToLower(strings.TrimSpace(artifact.OS))
	return artifactOS == "" || artifactOS == strings.ToLower(osName)
}

func installerEvidenceFromArtifact(manifest packageRepositoryManifest, artifact packageRepositoryManifestArtifact) (installerPackageEvidence, bool) {
	artifactRef := artifactPathRef(artifact)
	checksumRef := manifestBackedRef(artifact.ChecksumFile, manifest.ChecksumFile)
	signatureRef := manifestBackedRef(artifact.ChecksumSignature, manifest.ChecksumSignature)
	publicKeyRef := manifestBackedRef(artifact.PublicKey, manifest.PublicKey)
	for _, ref := range []string{artifactRef, checksumRef, signatureRef, publicKeyRef} {
		if _, ok := safePackageRepositoryRef(ref); !ok {
			return installerPackageEvidence{}, false
		}
	}
	return installerPackageEvidence{
		PackageID:      packageRepositoryArtifactID(artifact),
		ArtifactRef:    artifactRef,
		ChecksumRef:    checksumRef,
		SignatureRef:   signatureRef,
		PublicKeyRef:   publicKeyRef,
		SignatureScope: strings.TrimSpace(manifest.SignatureScope),
		ToolRefs:       map[string]string{},
	}, true
}

func installerToolRefs(tools []model.FindXAgentInstallToolEvidence) map[string]string {
	refs := map[string]string{}
	for _, tool := range tools {
		if tool.Status == "ready" && strings.TrimSpace(tool.EvidenceRef) != "" {
			refs[tool.Name] = tool.EvidenceRef
		}
	}
	return refs
}

func installerContentType(installer string) string {
	switch installer {
	case "linux.sh":
		return "text/x-shellscript; charset=utf-8"
	case "windows.ps1":
		return "text/plain; charset=utf-8"
	default:
		return "text/plain; charset=utf-8"
	}
}

func renderInstallerScript(installer, platform string, evidence installerPackageEvidence) string {
	switch installer {
	case "windows.ps1":
		return renderWindowsPowerShellInstaller(evidence)
	case "windows.bat":
		return renderWindowsBatchInstaller(evidence)
	default:
		return renderLinuxShellInstaller(platform, evidence)
	}
}

func installerArtifactDownloadPath(evidence installerPackageEvidence) string {
	return installerPackageDownloadPath(evidence.PackageID, evidence.ArtifactRef)
}

func installerPackageDownloadPath(packageID, ref string) string {
	return "/api/v1/findx-agents/package-downloads/" + url.PathEscape(packageID) +
		"?artifact=" + url.QueryEscape(ref)
}

func renderLinuxShellInstaller(platform string, evidence installerPackageEvidence) string {
	artifactURL := installerArtifactDownloadPath(evidence)
	checksumURL := installerPackageDownloadPath(evidence.PackageID, evidence.ChecksumRef)
	signatureURL := installerPackageDownloadPath(evidence.PackageID, evidence.SignatureRef)
	publicKeyURL := installerPackageDownloadPath(evidence.PackageID, evidence.PublicKeyRef)
	toolEvidence := installerBundledToolEvidenceLines("#", evidence)
	return fmt.Sprintf(`#!/usr/bin/env sh
set -eu
%s
FINDX_BASE_URL="${FINDX_BASE_URL:-http://127.0.0.1:8080}"
FINDX_TOKEN="${FINDX_TOKEN:-<TOKEN>}"
PACKAGE_URL="${FINDX_BASE_URL}%s"
CHECKSUM_URL="${FINDX_BASE_URL}%s"
SIGNATURE_URL="${FINDX_BASE_URL}%s"
PUBLIC_KEY_URL="${FINDX_BASE_URL}%s"
INSTALL_ROOT="${FINDX_INSTALL_ROOT:-/opt/findx-agent}"
SERVICE_NAME="${FINDX_SERVICE_NAME:-findx-agent}"
TMP_DIR="$(mktemp -d /tmp/findx-agent.XXXXXX)"
TMP_FILE="${TMP_DIR}/package.tar.gz"
CHECKSUM_FILE="${TMP_DIR}/SHA256SUMS"
SIGNATURE_FILE="${TMP_DIR}/SHA256SUMS.sig"
PUBLIC_KEY_FILE="${TMP_DIR}/findx-agent.pub"
curl -fL -H "Authorization: Bearer ${FINDX_TOKEN}" -o "${TMP_FILE}" "${PACKAGE_URL}"
curl -fL -H "Authorization: Bearer ${FINDX_TOKEN}" -o "${CHECKSUM_FILE}" "${CHECKSUM_URL}"
curl -fL -H "Authorization: Bearer ${FINDX_TOKEN}" -o "${SIGNATURE_FILE}" "${SIGNATURE_URL}"
curl -fL -H "Authorization: Bearer ${FINDX_TOKEN}" -o "${PUBLIC_KEY_FILE}" "${PUBLIC_KEY_URL}"
printf 'artifact=%s\nchecksum=%s\nsignature=%s\npublic_key=%s\nsignature_scope=%s\n'
(cd "${TMP_DIR}" && sha256sum -c "${CHECKSUM_FILE}")
gpg --import "${PUBLIC_KEY_FILE}"
gpg --verify "${SIGNATURE_FILE}" "${CHECKSUM_FILE}"
install -d "${INSTALL_ROOT}"
tar -xzf "${TMP_FILE}" -C "${INSTALL_ROOT}"
printf 'install FindX Agent service %%s for %s\n' "${SERVICE_NAME}"
`, toolEvidence, artifactURL, checksumURL, signatureURL, publicKeyURL, evidence.ArtifactRef, evidence.ChecksumRef, evidence.SignatureRef,
		evidence.PublicKeyRef, evidence.SignatureScope, platform)
}

func renderWindowsPowerShellInstaller(evidence installerPackageEvidence) string {
	artifactURL := installerArtifactDownloadPath(evidence)
	checksumURL := installerPackageDownloadPath(evidence.PackageID, evidence.ChecksumRef)
	signatureURL := installerPackageDownloadPath(evidence.PackageID, evidence.SignatureRef)
	publicKeyURL := installerPackageDownloadPath(evidence.PackageID, evidence.PublicKeyRef)
	toolEvidence := installerBundledToolEvidenceLines("#", evidence)
	return fmt.Sprintf(`$ErrorActionPreference = "Stop"
%s
$FindXBaseUrl = if ($env:FINDX_BASE_URL) { $env:FINDX_BASE_URL } else { "http://127.0.0.1:8080" }
$FindXToken = if ($env:FINDX_TOKEN) { $env:FINDX_TOKEN } else { "<TOKEN>" }
$PackageUrl = "$FindXBaseUrl%s"
$ChecksumUrl = "$FindXBaseUrl%s"
$SignatureUrl = "$FindXBaseUrl%s"
$PublicKeyUrl = "$FindXBaseUrl%s"
$InstallRoot = if ($env:FINDX_INSTALL_ROOT) { $env:FINDX_INSTALL_ROOT } else { "C:\Program Files\FindX Agent" }
$ServiceName = if ($env:FINDX_SERVICE_NAME) { $env:FINDX_SERVICE_NAME } else { "FindX Agent" }
$TempDir = Join-Path $env:TEMP ("findx-agent-" + [guid]::NewGuid().ToString("N"))
New-Item -ItemType Directory -Force -Path $TempDir | Out-Null
$TempFile = Join-Path $TempDir "package.zip"
$ChecksumFile = Join-Path $TempDir "SHA256SUMS"
$SignatureFile = Join-Path $TempDir "SHA256SUMS.sig"
$PublicKeyFile = Join-Path $TempDir "findx-agent.pub"
Invoke-WebRequest -Uri $PackageUrl -Headers @{ Authorization = "Bearer $FindXToken" } -OutFile $TempFile
Invoke-WebRequest -Uri $ChecksumUrl -Headers @{ Authorization = "Bearer $FindXToken" } -OutFile $ChecksumFile
Invoke-WebRequest -Uri $SignatureUrl -Headers @{ Authorization = "Bearer $FindXToken" } -OutFile $SignatureFile
Invoke-WebRequest -Uri $PublicKeyUrl -Headers @{ Authorization = "Bearer $FindXToken" } -OutFile $PublicKeyFile
Write-Output "artifact=%s checksum=%s signature=%s public_key=%s signature_scope=%s"
$Hash = Get-FileHash -Algorithm SHA256 $TempFile
$ExpectedHash = Select-String -Path $ChecksumFile -Pattern "[A-Fa-f0-9]{64}" | Select-Object -First 1 | ForEach-Object { $_.Matches[0].Value }
if (!$ExpectedHash) { throw "FindX Agent checksum file has no SHA256 value" }
if ($Hash.Hash.ToLowerInvariant() -ne $ExpectedHash.ToLowerInvariant()) { throw "FindX Agent checksum mismatch" }
if (Get-Command gpg -ErrorAction SilentlyContinue) {
  & gpg --import $PublicKeyFile
  & gpg --verify $SignatureFile $ChecksumFile
  if ($LASTEXITCODE -ne 0) { throw "FindX Agent signature verification failed" }
} else {
  throw "FindX Agent signature verifier is missing"
}
New-Item -ItemType Directory -Force -Path $InstallRoot | Out-Null
Expand-Archive -Force -Path $TempFile -DestinationPath $InstallRoot
Write-Output "install FindX Agent service $ServiceName"
`, toolEvidence, artifactURL, checksumURL, signatureURL, publicKeyURL, evidence.ArtifactRef, evidence.ChecksumRef, evidence.SignatureRef,
		evidence.PublicKeyRef, evidence.SignatureScope)
}

func renderWindowsBatchInstaller(evidence installerPackageEvidence) string {
	artifactURL := installerArtifactDownloadPath(evidence)
	checksumURL := installerPackageDownloadPath(evidence.PackageID, evidence.ChecksumRef)
	signatureURL := installerPackageDownloadPath(evidence.PackageID, evidence.SignatureRef)
	publicKeyURL := installerPackageDownloadPath(evidence.PackageID, evidence.PublicKeyRef)
	toolEvidence := installerBundledToolEvidenceLines("rem", evidence)
	return fmt.Sprintf(`@echo off
setlocal enabledelayedexpansion
%s
if "%%FINDX_BASE_URL%%"=="" set "FINDX_BASE_URL=http://127.0.0.1:8080"
if "%%FINDX_TOKEN%%"=="" set "FINDX_TOKEN=<TOKEN>"
if "%%FINDX_INSTALL_ROOT%%"=="" set "FINDX_INSTALL_ROOT=%%ProgramFiles%%\FindX Agent"
if "%%FINDX_SERVICE_NAME%%"=="" set "FINDX_SERVICE_NAME=FindX Agent"
set "PACKAGE_URL=%%FINDX_BASE_URL%%%s"
set "CHECKSUM_URL=%%FINDX_BASE_URL%%%s"
set "SIGNATURE_URL=%%FINDX_BASE_URL%%%s"
set "PUBLIC_KEY_URL=%%FINDX_BASE_URL%%%s"
set "TMP_DIR=%%TEMP%%\findx-agent-%%RANDOM%%%%RANDOM%%"
mkdir "%%TMP_DIR%%"
set "TMP_FILE=%%TMP_DIR%%\package.zip"
set "CHECKSUM_FILE=%%TMP_DIR%%\SHA256SUMS"
set "SIGNATURE_FILE=%%TMP_DIR%%\SHA256SUMS.sig"
set "PUBLIC_KEY_FILE=%%TMP_DIR%%\findx-agent.pub"
certutil -urlcache -f "%%PACKAGE_URL%%" "%%TMP_FILE%%"
certutil -urlcache -f "%%CHECKSUM_URL%%" "%%CHECKSUM_FILE%%"
certutil -urlcache -f "%%SIGNATURE_URL%%" "%%SIGNATURE_FILE%%"
certutil -urlcache -f "%%PUBLIC_KEY_URL%%" "%%PUBLIC_KEY_FILE%%"
echo artifact=%s checksum=%s signature=%s public_key=%s signature_scope=%s
certutil -hashfile "%%TMP_FILE%%" SHA256
powershell -NoProfile -ExecutionPolicy Bypass -Command "$hash=(Get-FileHash -Algorithm SHA256 '%%TMP_FILE%%').Hash.ToLowerInvariant(); $expected=(Select-String -Path '%%CHECKSUM_FILE%%' -Pattern '[A-Fa-f0-9]{64}' | Select-Object -First 1).Matches[0].Value.ToLowerInvariant(); if ($hash -ne $expected) { throw 'FindX Agent checksum mismatch' }; if (Get-Command gpg -ErrorAction SilentlyContinue) { & gpg --import '%%PUBLIC_KEY_FILE%%'; & gpg --verify '%%SIGNATURE_FILE%%' '%%CHECKSUM_FILE%%'; if ($LASTEXITCODE -ne 0) { throw 'FindX Agent signature verification failed' } } else { throw 'FindX Agent signature verifier is missing' }; New-Item -ItemType Directory -Force -Path '%%FINDX_INSTALL_ROOT%%' | Out-Null; Expand-Archive -Force -Path '%%TMP_FILE%%' -DestinationPath '%%FINDX_INSTALL_ROOT%%'"
echo install FindX Agent service %%FINDX_SERVICE_NAME%%
`, toolEvidence, artifactURL, checksumURL, signatureURL, publicKeyURL, evidence.ArtifactRef, evidence.ChecksumRef, evidence.SignatureRef,
		evidence.PublicKeyRef, evidence.SignatureScope)
}

func installerBundledToolEvidenceLines(comment string, evidence installerPackageEvidence) string {
	names := make([]string, 0, len(evidence.ToolRefs))
	for name := range evidence.ToolRefs {
		names = append(names, name)
	}
	sort.Strings(names)
	var builder strings.Builder
	builder.WriteString(comment)
	builder.WriteString(" bundled_install_environment=ready\n")
	for _, name := range names {
		builder.WriteString(comment)
		builder.WriteString(" bundled_tool ")
		builder.WriteString(name)
		builder.WriteString("=")
		builder.WriteString(evidence.ToolRefs[name])
		builder.WriteString("\n")
	}
	return strings.TrimRight(builder.String(), "\n")
}
