// Package costderivation loads and validates per-kind cost
// derivations -- the catalog/_pricing/derivations/<kind>.yaml
// documents carrying the machine-executable rules that turn ANY
// manifest's spec values into metered quantities and price choices. A
// derivation replaces the hand-authored estimate model for its kind
// (a kind carries exactly one of the two): the estimate generator
// replays every catalog preset through the rules to produce the
// committed estimates, and the same rules are what a server-side
// estimator evaluates against a live manifest. Enrollment is the file's
// presence: every discovered derivation is held to this package's
// conformance gate, with no allowlist to keep in sync.
package costderivation

import (
	"path/filepath"
	"sort"
	"strings"

	"github.com/pkg/errors"

	derivationv1 "github.com/plantonhq/planton/finops/catalogkindcostderivation/v1"
	"github.com/plantonhq/planton/pkg/protobufyaml"
)

// Dir is the derivations' home, relative to the repo root. Derivations
// live beside the price book and the generated estimates -- the estimate
// pipeline's data in one tree -- while the kind's cost.yaml stays
// beside the kind as its authored fact sheet.
const Dir = "catalog/_pricing/derivations"

// Path is a kind's cost derivation location. The filename is the
// kind's identity (the same convention the models and estimates use).
func Path(repoRoot, kindDir string) string {
	return filepath.Join(repoRoot, Dir, kindDir+".yaml")
}

// Discover returns the kind names that ship a cost derivation, sorted.
func Discover(repoRoot string) ([]string, error) {
	matches, err := filepath.Glob(filepath.Join(repoRoot, Dir, "*.yaml"))
	if err != nil {
		return nil, err
	}
	var kindDirs []string
	for _, m := range matches {
		kindDirs = append(kindDirs, strings.TrimSuffix(filepath.Base(m), ".yaml"))
	}
	sort.Strings(kindDirs)
	return kindDirs, nil
}

// Load reads and strictly parses a kind's cost derivation.
func Load(repoRoot, kindDir string) (*derivationv1.CatalogKindCostDerivation, error) {
	derivation := &derivationv1.CatalogKindCostDerivation{}
	if err := protobufyaml.Load(Path(repoRoot, kindDir), derivation); err != nil {
		return nil, errors.Wrapf(err, "loading cost derivation for %s", kindDir)
	}
	return derivation, nil
}
