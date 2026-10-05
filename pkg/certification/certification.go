// Package certification is the catalog's launch-certification suite.
//
// The suite drives the synthetic torture kinds (the _test provider) through
// every capability the platform promises about API versioning and data
// safety, using the same seams real kinds flow through. The torture kinds
// exist so these proofs are hermetic and permanent: no cloud, no
// credentials, no dependence on production kind shapes that refactor over
// time.
//
// The suite grows with the machinery -- a capability lands together with its
// certification cases, never later:
//
//   - manifest lifecycle: envelope enforcement, strict loading, validation
//     (this package's initial cases)
//   - version conversion: golden-corpus round-trips once the conversion
//     engine exists
//   - bundle conformance: catalog-bundle round-trips once the bundle exists
//
// Public launch is gated on the full suite passing -- treat a failure here
// as a release blocker, not a flaky test.
package certification

import (
	"os"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/plantonhq/planton/pkg/catalogkindreflect"
)

// TortureKindRoot returns the torture kind's directory (all versions),
// resolved from this file's location so cases work from any test working
// directory.
func TortureKindRoot(t *testing.T) string {
	t.Helper()
	_, thisFile, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("cannot resolve caller location")
	}
	repoRoot := filepath.Join(filepath.Dir(thisFile), "..", "..")
	dir := filepath.Join(repoRoot, "catalog", "_test", "testcatalogkindgeneric")
	if _, err := os.Stat(dir); err != nil {
		t.Fatalf("torture kind directory missing: %v", err)
	}
	return dir
}

// TortureKindDir returns the torture kind's SERVED version directory. The
// version segment comes from the registry -- the same derivation every path
// builder in the repo uses, so certification moves with graduations
// automatically.
func TortureKindDir(t *testing.T) string {
	t.Helper()
	versionDir, err := catalogkindreflect.KindVersionDir("testcatalogkindgeneric")
	if err != nil {
		t.Fatalf("resolving the torture kind's served version: %v", err)
	}
	dir := filepath.Join(TortureKindRoot(t), versionDir)
	if _, err := os.Stat(dir); err != nil {
		t.Fatalf("torture kind served-version directory missing: %v", err)
	}
	return dir
}

// TortureManifestPath returns the canonical known-good torture manifest (the
// kind's default preset -- one source of truth for "a valid manifest").
// Presets live at the kind root: the living kind is version-agnostic,
// authored against the served (hub) version.
func TortureManifestPath(t *testing.T) string {
	t.Helper()
	return filepath.Join(TortureKindRoot(t), "presets", "01-default.yaml")
}
