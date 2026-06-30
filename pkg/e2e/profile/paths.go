package profile

import "path/filepath"

const (
	providerBase = "apis/dev/planton/provider"

	// ProviderProfileRelPath is the path to the provider E2E profile relative
	// to the provider directory (e.g., kubernetes/aa_e2e/profile.yaml).
	ProviderProfileRelPath = "aa_e2e/profile.yaml"

	// ComponentProfileRelPath is the path to the component E2E profile relative
	// to the component directory (e.g., kubernetesredis/v1/e2e/profile.yaml).
	ComponentProfileRelPath = "v1/e2e/profile.yaml"

	// ComponentScenariosRelDir is the path to the component test scenarios relative
	// to the component directory.
	ComponentScenariosRelDir = "v1/e2e/scenarios"

	// ComponentFixturesRelDir is the path to the component fixture manifests relative
	// to the component directory.
	ComponentFixturesRelDir = "v1/e2e/fixtures"
)

// ProviderDir returns the absolute path to a provider's directory.
func ProviderDir(repoRoot, provider string) string {
	return filepath.Join(repoRoot, providerBase, provider)
}

// ProviderProfilePath returns the absolute path to a provider's E2E profile.
func ProviderProfilePath(repoRoot, provider string) string {
	return filepath.Join(repoRoot, providerBase, provider, ProviderProfileRelPath)
}

// ComponentProfilePath returns the absolute path to a component's E2E profile.
func ComponentProfilePath(repoRoot, provider, component string) string {
	return filepath.Join(repoRoot, providerBase, provider, component, ComponentProfileRelPath)
}

// ComponentScenariosDir returns the absolute path to a component's test scenarios directory.
func ComponentScenariosDir(repoRoot, provider, component string) string {
	return filepath.Join(repoRoot, providerBase, provider, component, ComponentScenariosRelDir)
}
