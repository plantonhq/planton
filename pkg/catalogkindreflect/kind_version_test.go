package catalogkindreflect

import (
	"testing"

	"github.com/plantonhq/planton/shared/catalogkind"
)

// TestKindVersionsMatchMaturityGrammar walks every registered kind and
// asserts its declared version parses under the maturity grammar
// (^v\d+((alpha|beta)\d+)?$). The version name declares the kind's
// compatibility channel, and enum-value options are compile-time data that
// protovalidate never evaluates — this gate is the enforcement point, in the
// same pattern as the registry-unambiguity and envelope-const gates beside
// it. A kind that fails here is misdeclaring its guarantees.
func TestKindVersionsMatchMaturityGrammar(t *testing.T) {
	checked := 0
	for _, kind := range KindsList() {
		if kind == catalogkind.CatalogKind_unspecified {
			continue
		}
		version, err := KindVersion(kind)
		if err != nil {
			t.Errorf("%s: %v", kind, err)
			continue
		}
		if !versionGrammar.MatchString(version) {
			t.Errorf("%s: declared version %q does not match the maturity grammar %s",
				kind, version, versionGrammar.String())
			continue
		}
		checked++
	}
	// Vacuous guard: an empty walk would pass every assertion while gating
	// nothing.
	if checked == 0 {
		t.Fatal("no kinds were checked — the registry walk is broken")
	}
}

func TestKindVersion(t *testing.T) {
	version, err := KindVersion(catalogkind.CatalogKind_AwsVpc)
	if err != nil {
		t.Fatalf("AwsVpc: %v", err)
	}
	if version != "v1alpha1" {
		t.Errorf("AwsVpc: expected version v1alpha1, got %q", version)
	}
}

func TestKindVersionDir(t *testing.T) {
	// Names arrive in manifest and directory shapes alike; all must resolve.
	for _, name := range []string{"AwsVpc", "awsvpc", "aws-vpc", "aws_vpc"} {
		version, err := KindVersionDir(name)
		if err != nil {
			t.Errorf("%q: %v", name, err)
			continue
		}
		if version != "v1alpha1" {
			t.Errorf("%q: expected version dir v1alpha1, got %q", name, version)
		}
	}

	if _, err := KindVersionDir("NoSuchKind"); err == nil {
		t.Error("unknown kind name must fail plainly instead of composing a nonexistent path")
	}
}
