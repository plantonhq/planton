// Package controlprofile loads and validates per-catalog kind control profiles
// -- the catalog/<provider>/<kind>/controls.yaml sidecars declaring each
// kind's posture against the central control catalog
// (pkg/compliance/controlcatalog). Enrollment is the file's presence: every
// discovered controls.yaml is held to this package's conformance gate --
// including COMPLETE examination of the catalog's controls -- with no
// allowlist to keep in sync.
package controlprofile

import (
	"path/filepath"
	"sort"

	"github.com/pkg/errors"

	controlprofilev1 "github.com/plantonhq/planton/compliance/catalogkindcontrolprofile/v1"
	"github.com/plantonhq/planton/pkg/protobufyaml"
)

// FileName is the sidecar's name at the kind root.
const FileName = "controls.yaml"

// Path is a kind's control profile location.
func Path(repoRoot, provider, kindDir string) string {
	return filepath.Join(repoRoot, "catalog", provider, kindDir, FileName)
}

// Discover returns provider -> kinds that ship a control profile,
// sorted.
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

// Load reads and strictly parses a kind's control profile.
func Load(repoRoot, provider, kindDir string) (*controlprofilev1.CatalogKindControlProfile, error) {
	profile := &controlprofilev1.CatalogKindControlProfile{}
	if err := protobufyaml.Load(Path(repoRoot, provider, kindDir), profile); err != nil {
		return nil, errors.Wrapf(err, "loading control profile for %s/%s", provider, kindDir)
	}
	return profile, nil
}
