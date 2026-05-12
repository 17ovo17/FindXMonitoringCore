package handler

import (
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/gin-gonic/gin"
)

const packageDownloadBlocker = agentBlocked + ": package repository artifact, checksum, signature, and public key evidence are not open"

const (
	packageDownloadArtifactRefInvalid = "PACKAGE_REPOSITORY_REF_INVALID"
	packageDownloadArtifactRefMissing = "PACKAGE_REPOSITORY_ARTIFACT_REF_MISSING"
	packageDownloadPackageMissing     = "PACKAGE_REPOSITORY_PACKAGE_MISSING"
	packageDownloadRequestMissing     = "PACKAGE_REQUEST_MISSING"
	packageDownloadUnknownPackage     = "PACKAGE_NOT_FOUND"
)

type packageDownloadEvidence struct {
	FilePath    string
	ArtifactRef string
}

func DownloadFindXAgentPackageArtifact(c *gin.Context) {
	evidence, blockers, ok := findPackageDownloadEvidence(c.Param("package"), c.Query("artifact"))
	if !ok {
		blockFindXAgentPackageDownload(c, blockers)
		return
	}
	file, err := os.Open(evidence.FilePath)
	if err != nil {
		blockFindXAgentPackageDownload(c, nil)
		return
	}
	defer file.Close()
	info, err := file.Stat()
	if err != nil || info.IsDir() {
		blockFindXAgentPackageDownload(c, nil)
		return
	}
	c.Header("Content-Disposition", `attachment; filename="`+packageArtifactDownloadName(evidence.ArtifactRef)+`"`)
	c.DataFromReader(http.StatusOK, info.Size(), packageArtifactContentType(evidence.ArtifactRef), file, nil)
}

func blockFindXAgentPackageDownload(c *gin.Context, blockers []string) {
	structuredBlockers := uniquePackageRepositoryBlockers(blockers)
	c.JSON(http.StatusConflict, gin.H{
		"code":              http.StatusConflict,
		"error":             packageDownloadBlocker,
		"status":            "blocked",
		"missing_contracts": packageDownloadMissingContracts(structuredBlockers),
		"blockers":          structuredBlockers,
		"safe": gin.H{
			"credential_echo": false,
			"path_traversal":  false,
			"safe_to_retry":   false,
		},
	})
}

func packageDownloadMissingContracts(blockers []string) []string {
	base := []string{
		"package_repository_artifact",
		"checksum",
		"signature",
		"public_key",
	}
	return uniquePackageRepositoryBlockers(append(base, blockers...))
}

func findPackageDownloadEvidence(requestedPackage, requestedArtifact string) (packageDownloadEvidence, []string, bool) {
	packageID := strings.TrimSpace(requestedPackage)
	if packageID == "" {
		return packageDownloadEvidence{}, []string{packageDownloadRequestMissing}, false
	}
	if _, ok := findAgentPackage(packageID); !ok {
		return packageDownloadEvidence{}, []string{packageDownloadUnknownPackage}, false
	}
	cleanRequestedArtifact, ok := safePackageRepositoryRef(requestedArtifact)
	if !ok {
		return packageDownloadEvidence{}, []string{packageDownloadArtifactRefInvalid}, false
	}
	blockers := []string{}
	for _, root := range agentPackageRepositoryRoots() {
		evidence, rootBlockers, ok := findPackageDownloadEvidenceInRoot(root, packageID, cleanRequestedArtifact)
		if ok {
			return evidence, nil, true
		}
		blockers = append(blockers, rootBlockers...)
	}
	if len(blockers) == 0 {
		blockers = append(blockers, packageDownloadArtifactRefMissing)
	}
	return packageDownloadEvidence{}, uniquePackageRepositoryBlockers(blockers), false
}

func findPackageDownloadEvidenceInRoot(root, packageID, cleanRequestedArtifact string) (packageDownloadEvidence, []string, bool) {
	diagnostics := diagnosePackageRepositoryManifest(root)
	if !diagnostics.Ready {
		return packageDownloadEvidence{}, diagnostics.Blockers, false
	}
	manifest := diagnostics.Manifest
	if isTestOnlyPackageRepositoryManifest(manifest) {
		return packageDownloadEvidence{}, packageDownloadTestOnlyBlockers(root, manifest, packageID), false
	}
	return findProductionPackageDownloadEvidence(root, manifest, packageID, cleanRequestedArtifact)
}

func findProductionPackageDownloadEvidence(root string, manifest packageRepositoryManifest, packageID, cleanRequestedArtifact string) (packageDownloadEvidence, []string, bool) {
	packageSeen := false
	for _, artifact := range manifest.Artifacts {
		if packageRepositoryArtifactID(artifact) != packageID {
			continue
		}
		packageSeen = true
		if !packageDownloadArtifactMatches(manifest, artifact, packageID, cleanRequestedArtifact) {
			continue
		}
		blockers := packageRepositoryTrustChainBlockers(root, manifest, artifact)
		if len(blockers) > 0 {
			return packageDownloadEvidence{}, blockers, false
		}
		filePath, ok := packageRepositoryDownloadFilePath(root, cleanRequestedArtifact)
		if !ok {
			return packageDownloadEvidence{}, packageRepositoryArtifactEvidenceBlockers(root, manifest, artifact), false
		}
		return packageDownloadEvidence{FilePath: filePath, ArtifactRef: cleanRequestedArtifact}, nil, true
	}
	return packageDownloadEvidence{}, packageDownloadPackageOrArtifactBlockers(root, manifest, packageID, packageSeen), false
}

func packageDownloadPackageOrArtifactBlockers(root string, manifest packageRepositoryManifest, packageID string, packageSeen bool) []string {
	if !packageSeen {
		return []string{packageDownloadPackageMissing}
	}
	blockers := []string{packageDownloadArtifactRefMissing}
	blockers = append(blockers, packageRepositoryProductionEvidenceBlockers(root, manifest, packageID)...)
	return uniquePackageRepositoryBlockers(blockers)
}

func packageDownloadTestOnlyBlockers(root string, manifest packageRepositoryManifest, packageID string) []string {
	blockers := packageRepositoryTestOnlyBlockers(root, manifest, packageID)
	blockers = append(blockers, packageRepositoryTrustChainDiagnosticBlockers(root, packageID)...)
	if len(blockers) == 0 {
		blockers = append(blockers, packageDownloadPackageMissing)
	}
	return uniquePackageRepositoryBlockers(blockers)
}

func packageDownloadArtifactMatches(manifest packageRepositoryManifest, artifact packageRepositoryManifestArtifact, packageID, cleanRequestedArtifact string) bool {
	if packageRepositoryArtifactID(artifact) != packageID {
		return false
	}
	for _, ref := range packageDownloadAllowedRefs(manifest, artifact) {
		cleanRef, ok := safePackageRepositoryRef(ref)
		if ok && cleanRef == cleanRequestedArtifact {
			return true
		}
	}
	return false
}

func packageDownloadAllowedRefs(manifest packageRepositoryManifest, artifact packageRepositoryManifestArtifact) []string {
	return []string{
		artifactPathRef(artifact),
		manifestBackedRef(artifact.ChecksumFile, manifest.ChecksumFile),
		manifestBackedRef(artifact.ChecksumSignature, manifest.ChecksumSignature),
		publicKeyOrFingerprintRef(manifest, artifact),
	}
}

func packageRepositoryDownloadFilePath(root, cleanRef string) (string, bool) {
	rootAbs, err := filepath.Abs(filepath.Clean(root))
	if err != nil {
		return "", false
	}
	fileAbs, err := filepath.Abs(filepath.Join(rootAbs, cleanRef))
	if err != nil {
		return "", false
	}
	rel, err := filepath.Rel(rootAbs, fileAbs)
	if err != nil || rel == ".." || strings.HasPrefix(rel, ".."+string(filepath.Separator)) {
		return "", false
	}
	info, err := os.Stat(fileAbs)
	if err != nil || info.IsDir() {
		return "", false
	}
	return fileAbs, true
}

func packageArtifactContentType(ref string) string {
	lower := strings.ToLower(ref)
	switch {
	case strings.HasSuffix(lower, ".zip"):
		return "application/zip"
	case strings.HasSuffix(lower, ".tar.gz"), strings.HasSuffix(lower, ".tgz"), strings.HasSuffix(lower, ".gz"):
		return "application/gzip"
	default:
		return "application/octet-stream"
	}
}

func packageArtifactDownloadName(ref string) string {
	name := filepath.Base(ref)
	name = strings.Map(func(r rune) rune {
		if r == '"' || r == '\\' || r == '/' || r < 32 {
			return -1
		}
		return r
	}, name)
	if strings.TrimSpace(name) == "" || name == "." {
		return "findx-agent-package.bin"
	}
	return name
}
