package handler

func agentPackageBlockers(packageID string) []string {
	if hasAgentPackageTestOnlyRepositoryEvidence(packageID) {
		return uniquePackageRepositoryBlockers(append([]string{
			"PACKAGE_REPOSITORY_TEST_ONLY",
			"SIGNATURE_TEST_ONLY",
			"PRODUCTION_PACKAGE_REPOSITORY_MISSING",
			"PRODUCTION_SIGNATURE_MISSING",
			"INSTALL_PLAN_CONTRACT_MISSING",
			"CONFIG_ROLLOUT_CONTRACT_MISSING",
		}, agentPackageInstallEnvironment(packageID).Blockers...))
	}
	return uniquePackageRepositoryBlockers(append([]string{
		"PACKAGE_REPOSITORY_MISSING",
		"SIGNATURE_MISSING",
		"INSTALL_PLAN_CONTRACT_MISSING",
		"CONFIG_ROLLOUT_CONTRACT_MISSING",
	}, agentPackageInstallEnvironment(packageID).Blockers...))
}
