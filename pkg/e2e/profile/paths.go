package profile

import (
	"path/filepath"

	"github.com/pkg/errors"
	"github.com/plantonhq/planton/pkg/catalogkindreflect"
)

const (
	catalogRoot = "catalog"

	// ProviderProfileRelPath is the path to the provider E2E profile relative
	// to the provider directory (e.g., kubernetes/aa_e2e/profile.yaml).
	ProviderProfileRelPath = "aa_e2e/profile.yaml"
)

// ProviderDir returns the absolute path to a provider's directory.
func ProviderDir(repoRoot, provider string) string {
	return filepath.Join(repoRoot, catalogRoot, provider)
}

// ProviderProfilePath returns the absolute path to a provider's E2E profile.
func ProviderProfilePath(repoRoot, provider string) string {
	return filepath.Join(repoRoot, catalogRoot, provider, ProviderProfileRelPath)
}

// validateKind checks a kind directory name against the kind
// registry. Kind E2E assets live at the kind root
// (e.g. kubernetesvalkey/e2e/...); a directory that does not resolve to a
// registered kind is not a kind.
func validateKind(kindDir string) error {
	if _, err := catalogkindreflect.KindVersionDir(kindDir); err != nil {
		return errors.Wrapf(err, "cannot locate E2E assets for kind %q", kindDir)
	}
	return nil
}

// KindProfilePath returns the absolute path to a kind's E2E profile
// (e.g. kubernetesvalkey/e2e/profile.yaml).
func KindProfilePath(repoRoot, provider, kindDir string) (string, error) {
	if err := validateKind(kindDir); err != nil {
		return "", err
	}
	return filepath.Join(repoRoot, catalogRoot, provider, kindDir, "e2e", "profile.yaml"), nil
}

// KindScenariosDir returns the absolute path to a kind's test scenarios directory.
func KindScenariosDir(repoRoot, provider, kindDir string) (string, error) {
	if err := validateKind(kindDir); err != nil {
		return "", err
	}
	return filepath.Join(repoRoot, catalogRoot, provider, kindDir, "e2e", "scenarios"), nil
}

// KindFixturesDir returns the absolute path to a kind's fixture manifests directory.
func KindFixturesDir(repoRoot, provider, kindDir string) (string, error) {
	if err := validateKind(kindDir); err != nil {
		return "", err
	}
	return filepath.Join(repoRoot, catalogRoot, provider, kindDir, "e2e", "fixtures"), nil
}
