// Package estimatemodel loads and validates per-kind estimate models
// -- the catalog/_pricing/models/<kind>.yaml documents stating each
// preset's quantity assumptions (how much of which declared meter, defended
// in prose). Models carry the AUTHORED half of an estimate; unit prices
// live in the provider's price book, and the estimate generator joins the
// two into the generated catalog/_pricing/estimates/ documents. Enrollment
// is the file's presence: every discovered model is held to this package's
// conformance gate, with no allowlist to keep in sync.
package estimatemodel

import (
	"path/filepath"
	"sort"
	"strings"

	"github.com/pkg/errors"

	estimatemodelv1 "github.com/plantonhq/planton/finops/catalogkindcostestimatemodel/v1"
	"github.com/plantonhq/planton/pkg/protobufyaml"
)

// Dir is the models' home, relative to the repo root. Models live beside
// the price book and the generated estimates -- the estimate pipeline's
// data in one tree -- while the kind's cost.yaml stays beside the
// kind as its authored fact sheet.
const Dir = "catalog/_pricing/models"

// Path is a kind's estimate model location. The filename is the
// kind's identity (the same convention the estimates use).
func Path(repoRoot, kindDir string) string {
	return filepath.Join(repoRoot, Dir, kindDir+".yaml")
}

// Discover returns the kind names that ship an estimate model, sorted.
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

// Load reads and strictly parses a kind's estimate model.
func Load(repoRoot, kindDir string) (*estimatemodelv1.CatalogKindCostEstimateModel, error) {
	model := &estimatemodelv1.CatalogKindCostEstimateModel{}
	if err := protobufyaml.Load(Path(repoRoot, kindDir), model); err != nil {
		return nil, errors.Wrapf(err, "loading estimate model for %s", kindDir)
	}
	return model, nil
}
