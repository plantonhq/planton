// Package permissions loads and validates per-kind runner permission
// manifests -- the catalog/<provider>/<kind>/iac/permissions.yaml sidecars
// declaring the least-privilege permissions the IaC runner's cloud
// principal needs for the OFFICIAL modules. Every entry carries provenance
// (derived from module static analysis, or proven by live-run capture);
// the distinction is a trust feature consumers display. Enrollment is the
// file's presence: every discovered permissions.yaml is held to this
// package's conformance gate, with no allowlist to keep in sync.
package permissions

import (
	"path/filepath"
	"sort"

	"github.com/pkg/errors"

	permissionsv1 "github.com/plantonhq/planton/iac/catalogkindpermissions/v1"
	"github.com/plantonhq/planton/pkg/protobufyaml"
)

// FileName is the sidecar's name inside a kind's iac/ directory.
const FileName = "permissions.yaml"

// Path is a kind's permissions manifest location.
func Path(repoRoot, provider, kindDir string) string {
	return filepath.Join(repoRoot, "catalog", provider, kindDir, "iac", FileName)
}

// Discover returns provider -> kinds that ship a permissions manifest,
// sorted.
func Discover(repoRoot string) (map[string][]string, error) {
	matches, err := filepath.Glob(filepath.Join(repoRoot, "catalog", "*", "*", "iac", FileName))
	if err != nil {
		return nil, err
	}
	discovered := map[string][]string{}
	for _, m := range matches {
		kindDir := filepath.Base(filepath.Dir(filepath.Dir(m)))
		provider := filepath.Base(filepath.Dir(filepath.Dir(filepath.Dir(m))))
		discovered[provider] = append(discovered[provider], kindDir)
	}
	for _, kindDirs := range discovered {
		sort.Strings(kindDirs)
	}
	return discovered, nil
}

// Load reads and strictly parses a kind's permissions manifest.
func Load(repoRoot, provider, kindDir string) (*permissionsv1.CatalogKindPermissions, error) {
	manifest := &permissionsv1.CatalogKindPermissions{}
	if err := protobufyaml.Load(Path(repoRoot, provider, kindDir), manifest); err != nil {
		return nil, errors.Wrapf(err, "loading permissions for %s/%s", provider, kindDir)
	}
	return manifest, nil
}
