package handler

import (
	"encoding/json"
	"os"
	"path/filepath"
)

type packageRepositoryManifestDiagnostics struct {
	Manifest packageRepositoryManifest
	Blockers []string
	Ready    bool
}

func diagnosePackageRepositoryManifest(root string) packageRepositoryManifestDiagnostics {
	var manifest packageRepositoryManifest
	data, err := os.ReadFile(filepath.Join(root, packageRepositoryManifestPath))
	if err != nil {
		return packageRepositoryManifestDiagnostics{
			Blockers: []string{"PACKAGE_REPOSITORY_MANIFEST_MISSING"},
		}
	}
	if err := json.Unmarshal(data, &manifest); err != nil {
		return packageRepositoryManifestDiagnostics{
			Blockers: []string{"PACKAGE_REPOSITORY_MANIFEST_INVALID"},
		}
	}
	return packageRepositoryManifestDiagnostics{
		Manifest: manifest,
		Ready:    true,
	}
}

func packageRepositoryManifestDiagnosticBlockers(root, packageID string) []string {
	diagnostics := diagnosePackageRepositoryManifest(root)
	if !diagnostics.Ready {
		return diagnostics.Blockers
	}
	if isTestOnlyPackageRepositoryManifest(diagnostics.Manifest) {
		return packageRepositoryTestOnlyBlockers(root, diagnostics.Manifest, packageID)
	}
	return packageRepositoryProductionEvidenceBlockers(root, diagnostics.Manifest, packageID)
}

func packageRepositoryTestOnlyBlockers(root string, manifest packageRepositoryManifest, packageID string) []string {
	if !manifestHasPackageArtifact(manifest, packageID) {
		return nil
	}
	if !packageRepositoryPackageArtifactRefsSafe(manifest, packageID) {
		return packageRepositoryProductionEvidenceBlockers(root, manifest, packageID)
	}
	blockers := []string{"PACKAGE_REPOSITORY_TEST_ONLY", "SIGNATURE_TEST_ONLY"}
	blockers = append(blockers, packageRepositoryProductionEvidenceBlockers(root, manifest, packageID)...)
	return uniquePackageRepositoryBlockers(blockers)
}

func packageRepositoryProductionEvidenceBlockers(root string, manifest packageRepositoryManifest, packageID string) []string {
	blockers := []string{}
	for _, artifact := range manifest.Artifacts {
		if packageRepositoryArtifactID(artifact) != packageID {
			continue
		}
		blockers = append(blockers, packageRepositoryArtifactEvidenceBlockers(root, manifest, artifact)...)
	}
	return uniquePackageRepositoryBlockers(blockers)
}

func packageRepositoryPackageArtifactRefsSafe(manifest packageRepositoryManifest, packageID string) bool {
	for _, artifact := range manifest.Artifacts {
		if packageRepositoryArtifactID(artifact) == packageID && !packageRepositoryArtifactRefsSafe(manifest, artifact) {
			return false
		}
	}
	return true
}

func packageRepositoryArtifactRefsSafe(manifest packageRepositoryManifest, artifact packageRepositoryManifestArtifact) bool {
	for _, ref := range packageDownloadAllowedRefs(manifest, artifact) {
		if _, ok := safePackageRepositoryRef(ref); !ok {
			return false
		}
	}
	return true
}

func packageRepositoryArtifactEvidenceBlockers(root string, manifest packageRepositoryManifest, artifact packageRepositoryManifestArtifact) []string {
	blockers := []string{}
	if !safePackageRepositoryFileExists(root, artifactPathRef(artifact)) {
		blockers = append(blockers, "PACKAGE_REPOSITORY_ARTIFACT_MISSING")
	}
	if !safePackageRepositoryFileExists(root, manifestBackedRef(artifact.ChecksumFile, manifest.ChecksumFile)) {
		blockers = append(blockers, "PACKAGE_REPOSITORY_CHECKSUM_MISSING")
	}
	if !safePackageRepositoryFileExists(root, manifestBackedRef(artifact.ChecksumSignature, manifest.ChecksumSignature)) {
		blockers = append(blockers, "PRODUCTION_SIGNATURE_MISSING")
	}
	if !safePackageRepositoryFileExists(root, publicKeyOrFingerprintRef(manifest, artifact)) {
		blockers = append(blockers, "PRODUCTION_PUBLIC_KEY_MISSING")
	}
	if manifest.SignatureScope == "" || isTestOnlyPackageRepositoryManifest(manifest) {
		blockers = append(blockers, "PRODUCTION_SIGNATURE_MISSING")
	}
	return blockers
}

func packageRepositoryTrustChainDiagnosticBlockers(root, packageID string) []string {
	diagnostics := diagnosePackageRepositoryManifest(root)
	if !diagnostics.Ready {
		return diagnostics.Blockers
	}
	blockers := []string{}
	for _, artifact := range diagnostics.Manifest.Artifacts {
		if packageRepositoryArtifactID(artifact) == packageID {
			blockers = append(blockers, packageRepositoryTrustChainBlockers(root, diagnostics.Manifest, artifact)...)
		}
	}
	return uniquePackageRepositoryBlockers(blockers)
}
