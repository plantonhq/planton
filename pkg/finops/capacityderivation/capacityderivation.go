// Package capacityderivation loads and validates per-kind capacity
// derivations -- the catalog/_pricing/capacity/<kind>.yaml documents
// carrying the machine-executable rules that turn a cluster-capacity
// manifest's spec values into the capacity footprint it reserves from its
// target cluster (CPU/memory requests and limits, persistent volume
// storage). A capacity derivation replaces the hand-authored estimate
// model for its kind (a kind carries exactly one of the two):
// the estimate generator replays every catalog preset through the rules
// to produce the committed footprint estimates, and the same rules can
// compute a live manifest's footprint server-side. Enrollment is the
// file's presence: every discovered derivation is held to this package's
// conformance gate, with no allowlist to keep in sync.
package capacityderivation

import (
	"path/filepath"
	"sort"
	"strings"

	"github.com/pkg/errors"

	capacityv1 "github.com/plantonhq/planton/finops/catalogkindcapacityderivation/v1"
	"github.com/plantonhq/planton/pkg/protobufyaml"
)

// Dir is the capacity derivations' home, relative to the repo root --
// beside the cost derivations, models, price book, and generated
// estimates: the estimate pipeline's data in one tree.
const Dir = "catalog/_pricing/capacity"

// Path is a kind's capacity derivation location. The filename is the
// kind's identity (the same convention the derivations, models, and
// estimates use).
func Path(repoRoot, kindDir string) string {
	return filepath.Join(repoRoot, Dir, kindDir+".yaml")
}

// Discover returns the kind names that ship a capacity derivation,
// sorted.
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

// Load reads and strictly parses a kind's capacity derivation.
func Load(repoRoot, kindDir string) (*capacityv1.CatalogKindCapacityDerivation, error) {
	derivation := &capacityv1.CatalogKindCapacityDerivation{}
	if err := protobufyaml.Load(Path(repoRoot, kindDir), derivation); err != nil {
		return nil, errors.Wrapf(err, "loading capacity derivation for %s", kindDir)
	}
	return derivation, nil
}
