// Package costprofile loads and validates per-kind cost profiles --
// the catalog/<provider>/<kind>/cost.yaml sidecars declaring each
// kind's cost anatomy (billing model, always-on baseline charges, and
// the spec fields that drive the bill). The profile carries the STABLE half
// of cost knowledge; unit prices live in the central price book. Enrollment
// is the file's presence: every discovered cost.yaml is held to this
// package's conformance gate, with no allowlist to keep in sync.
package costprofile

import (
	"path/filepath"
	"sort"

	"github.com/pkg/errors"

	costprofilev1 "github.com/plantonhq/planton/finops/catalogkindcostprofile/v1"
	"github.com/plantonhq/planton/pkg/protobufyaml"
)

// FileName is the sidecar's name at the kind root.
const FileName = "cost.yaml"

// Path is a kind's cost profile location.
func Path(repoRoot, provider, kindDir string) string {
	return filepath.Join(repoRoot, "catalog", provider, kindDir, FileName)
}

// Discover returns provider -> kinds that ship a cost profile, sorted.
func Discover(repoRoot string) (map[string][]string, error) {
	matches, err := filepath.Glob(filepath.Join(repoRoot, "catalog", "*", "*", FileName))
	if err != nil {
		return nil, err
	}
	discovered := map[string][]string{}
	for _, m := range matches {
		kindDir := filepath.Base(filepath.Dir(m))
		provider := filepath.Base(filepath.Dir(filepath.Dir(m)))
		discovered[provider] = append(discovered[provider], kindDir)
	}
	for _, kindDirs := range discovered {
		sort.Strings(kindDirs)
	}
	return discovered, nil
}

// Load reads and strictly parses a kind's cost profile.
func Load(repoRoot, provider, kindDir string) (*costprofilev1.CatalogKindCostProfile, error) {
	profile := &costprofilev1.CatalogKindCostProfile{}
	if err := protobufyaml.Load(Path(repoRoot, provider, kindDir), profile); err != nil {
		return nil, errors.Wrapf(err, "loading cost profile for %s/%s", provider, kindDir)
	}
	return profile, nil
}
