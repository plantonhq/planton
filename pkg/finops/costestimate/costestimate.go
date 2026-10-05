// Package costestimate loads and validates per-kind preset cost
// estimates -- the catalog/_pricing/estimates/<kind>.yaml documents
// stating what each preset costs per month at published list prices. The
// estimates carry the VOLATILE half of cost knowledge (today's prices,
// each with source URL and retrieval date); the kind's cost.yaml
// carries the stable half (what bills at all), and the conformance gate
// binds the two: an estimate can only price meters its kind's cost
// profile declares. Enrollment is the file's presence: every discovered
// estimate is held to this package's conformance gate, with no allowlist
// to keep in sync.
package costestimate

import (
	"path/filepath"
	"sort"
	"strings"

	"github.com/pkg/errors"

	costestimatev1 "github.com/plantonhq/planton/finops/catalogkindcostestimate/v1"
	"github.com/plantonhq/planton/pkg/protobufyaml"
)

// Dir is the estimates' home, relative to the repo root. Estimates live
// centrally rather than beside their kinds because prices churn on
// their own cadence -- one tree to refresh, no touch on 676 kinds.
const Dir = "catalog/_pricing/estimates"

// Path is a kind's estimate document location. The filename is the
// kind's identity (the same convention the framework crosswalks use).
func Path(repoRoot, kindDir string) string {
	return filepath.Join(repoRoot, Dir, kindDir+".yaml")
}

// Discover returns the kind names that ship an estimate, sorted.
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

// Load reads and strictly parses a kind's estimate document.
func Load(repoRoot, kindDir string) (*costestimatev1.CatalogKindCostEstimate, error) {
	estimate := &costestimatev1.CatalogKindCostEstimate{}
	if err := protobufyaml.Load(Path(repoRoot, kindDir), estimate); err != nil {
		return nil, errors.Wrapf(err, "loading cost estimate for %s", kindDir)
	}
	return estimate, nil
}
