// Package controlcatalog loads the central control catalog and the
// framework crosswalks -- the authored compliance vocabulary under
// catalog/_compliance/. The catalog defines every technical control the
// per-catalog kind control profiles may reference; the crosswalks map external
// framework requirements (HIPAA, CIS, ...) onto those same control ids.
// Referential integrity in both directions is enforced by this package's
// conformance test, which CI runs on every catalog change.
package controlcatalog

import (
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/pkg/errors"

	controlcatalogv1 "github.com/plantonhq/planton/compliance/controlcatalog/v1"
	crosswalkv1 "github.com/plantonhq/planton/compliance/frameworkcrosswalk/v1"
	"github.com/plantonhq/planton/pkg/protobufyaml"
)

// CatalogPath is the control catalog's well-known location.
func CatalogPath(repoRoot string) string {
	return filepath.Join(repoRoot, "catalog", "_compliance", "controls-catalog.yaml")
}

// FrameworksDir is the crosswalks' well-known home.
func FrameworksDir(repoRoot string) string {
	return filepath.Join(repoRoot, "catalog", "_compliance", "frameworks")
}

// Load reads and strictly parses the control catalog.
func Load(repoRoot string) (*controlcatalogv1.ControlCatalog, error) {
	catalog := &controlcatalogv1.ControlCatalog{}
	if err := protobufyaml.Load(CatalogPath(repoRoot), catalog); err != nil {
		return nil, errors.Wrap(err, "loading control catalog")
	}
	return catalog, nil
}

// ControlIDs returns the catalog's control ids as a set.
func ControlIDs(catalog *controlcatalogv1.ControlCatalog) map[string]bool {
	ids := map[string]bool{}
	for _, c := range catalog.GetSpec().GetControls() {
		ids[c.GetId()] = true
	}
	return ids
}

// DiscoverCrosswalks returns the framework names (file basenames without
// extension) of every crosswalk on disk, sorted.
func DiscoverCrosswalks(repoRoot string) ([]string, error) {
	matches, err := filepath.Glob(filepath.Join(FrameworksDir(repoRoot), "*.yaml"))
	if err != nil {
		return nil, err
	}
	var names []string
	for _, m := range matches {
		names = append(names, strings.TrimSuffix(filepath.Base(m), ".yaml"))
	}
	sort.Strings(names)
	return names, nil
}

// LoadCrosswalk reads and strictly parses one framework crosswalk.
func LoadCrosswalk(repoRoot, framework string) (*crosswalkv1.FrameworkCrosswalk, error) {
	path := filepath.Join(FrameworksDir(repoRoot), framework+".yaml")
	if _, err := os.Stat(path); err != nil {
		return nil, errors.Wrapf(err, "crosswalk %s", framework)
	}
	crosswalk := &crosswalkv1.FrameworkCrosswalk{}
	if err := protobufyaml.Load(path, crosswalk); err != nil {
		return nil, errors.Wrapf(err, "loading crosswalk %s", framework)
	}
	return crosswalk, nil
}
