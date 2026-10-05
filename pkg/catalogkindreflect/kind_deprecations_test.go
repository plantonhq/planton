package catalogkindreflect

import (
	"testing"

	"github.com/plantonhq/planton/shared/catalogkind"
)

// TestKindDeprecationsAreCoherent walks every registered kind and asserts the
// compile-time half of the deprecation contract: every declared deprecation
// names a grammar-valid version, no version is deprecated twice on one kind,
// and the served version is never deprecated (there would be nothing to
// upgrade to while a kind serves exactly one version). The bundle-facing half
// of the contract — the deprecated version's schema exists in the release and
// a conversion path to the served version is authored — lives in the catalog
// bundle's conformance gate, which sees the built artifact this test cannot.
//
// Enum-value options are compile-time data protovalidate never evaluates;
// this gate is the enforcement point, in the same pattern as the version
// grammar and registry-unambiguity gates beside it.
func TestKindDeprecationsAreCoherent(t *testing.T) {
	deprecationsSeen := 0
	for _, kind := range KindsList() {
		if kind == catalogkind.CatalogKind_unspecified {
			continue
		}
		deprecations, err := KindDeprecations(kind)
		if err != nil {
			t.Errorf("%s: %v", kind, err)
			continue
		}
		if len(deprecations) == 0 {
			continue
		}
		servedVersion, err := KindVersion(kind)
		if err != nil {
			t.Errorf("%s: %v", kind, err)
			continue
		}
		declared := map[string]bool{}
		for _, deprecation := range deprecations {
			deprecationsSeen++
			version := deprecation.GetVersion()
			if !versionGrammar.MatchString(version) {
				t.Errorf("%s: deprecated version %q does not match the maturity grammar %s",
					kind, version, versionGrammar.String())
			}
			if declared[version] {
				t.Errorf("%s: version %q is deprecated twice — one entry per version", kind, version)
			}
			declared[version] = true
			if version == servedVersion {
				t.Errorf("%s: the served version %q is marked deprecated — a kind serving exactly one version has nothing to upgrade to, so the served version can never be the deprecated one",
					kind, version)
			}
		}
	}
	// Vacuous guard: the torture kind carries the permanent deprecation
	// fixture, so a walk that sees zero deprecations is a broken walk (or a
	// deleted fixture), not a clean registry.
	if deprecationsSeen == 0 {
		t.Fatal("no deprecations were checked — the registry walk is broken or the permanent fixture is gone")
	}
}

// The torture kind is the permanent fixture: its v1alpha1 deprecation rides
// every gate (this registry test, bundle conformance, platform discovery), so
// the deprecation lane can never silently stop being exercised.
func TestKindDeprecationsFixture(t *testing.T) {
	deprecations, err := KindDeprecations(catalogkind.CatalogKind_TestCatalogKindGeneric)
	if err != nil {
		t.Fatalf("TestCatalogKindGeneric: %v", err)
	}
	if len(deprecations) != 1 {
		t.Fatalf("TestCatalogKindGeneric: expected exactly one deprecation (the permanent fixture), got %d", len(deprecations))
	}
	if got := deprecations[0].GetVersion(); got != "v1alpha1" {
		t.Errorf("TestCatalogKindGeneric: expected v1alpha1 deprecated, got %q", got)
	}
	if deprecations[0].GetNote() == "" {
		t.Error("TestCatalogKindGeneric: the fixture carries a note so note passthrough stays exercised end to end")
	}
}

// Kinds with no deprecations — the overwhelmingly common case — answer with
// an empty slice, never an error.
func TestKindDeprecationsEmptyForUndeprecatedKind(t *testing.T) {
	deprecations, err := KindDeprecations(catalogkind.CatalogKind_AwsVpc)
	if err != nil {
		t.Fatalf("AwsVpc: %v", err)
	}
	if len(deprecations) != 0 {
		t.Errorf("AwsVpc: expected no deprecations, got %d", len(deprecations))
	}
}
