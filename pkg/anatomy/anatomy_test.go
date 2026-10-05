package anatomy

import (
	"os"
	"path/filepath"
	"runtime"
	"testing"
)

// repoRoot resolves the repository root from this file's location so the
// gate works from any test working directory (including the Bazel sandbox,
// where the catalog source tree is absent -- callers skip there).
func repoRoot(t *testing.T) string {
	t.Helper()
	_, thisFile, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("cannot resolve caller location")
	}
	return filepath.Join(filepath.Dir(thisFile), "..", "..")
}

// TestKindAnatomyGate is the CI guardrail: the live walk over the whole
// catalog must not introduce a violation outside the baseline or leave a
// stale baseline entry. On failure, either fix the anatomy (the detail says
// how) or -- for a deliberate, parity-routed gap -- regenerate the baseline
// with PLANTON_REGEN_ANATOMY_BASELINE=1 and justify the growth in review.
func TestKindAnatomyGate(t *testing.T) {
	root := repoRoot(t)
	if _, err := os.Stat(filepath.Join(root, "catalog")); err != nil {
		t.Skip("catalog source tree not present (bazel sandbox); runs under go test and the lint.kind-anatomy lane")
	}

	violations, err := Check(root)
	if err != nil {
		t.Fatalf("anatomy walk: %v", err)
	}

	_, thisFile, _, _ := runtime.Caller(0)
	baselinePath := filepath.Join(filepath.Dir(thisFile), "baseline.yaml")
	if os.Getenv("PLANTON_REGEN_ANATOMY_BASELINE") == "1" {
		if err := WriteBaseline(baselinePath, violations); err != nil {
			t.Fatalf("write baseline: %v", err)
		}
		t.Logf("baseline regenerated with %d entries -- review the diff before committing", len(violations))
		return
	}

	baseline, err := LoadBaseline(baselinePath)
	if err != nil {
		t.Fatalf("load baseline: %v", err)
	}
	res := Gate(violations, baseline)
	for _, v := range res.NewViolations {
		t.Errorf("anatomy drift: %s -- %s", v.ID(), v.Detail)
	}
	for _, id := range res.StaleEntries {
		t.Errorf("stale baseline entry (no longer a violation): %s -- remove it from baseline.yaml", id)
	}
}

// TestCheck_HermeticFixture proves every rule fires, against a synthetic
// catalog that violates each one -- the gate's own deliberate red. A gate
// that cannot fail teaches false confidence.
func TestCheck_HermeticFixture(t *testing.T) {
	root := t.TempDir()
	write := func(rel string) {
		p := filepath.Join(root, rel)
		if err := os.MkdirAll(filepath.Dir(p), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(p, []byte("x"), 0o644); err != nil {
			t.Fatal(err)
		}
	}

	// A well-formed kind missing several requirements, carrying several
	// forbidden residents. awss3bucket is a real registered kind, so the
	// registry lookup resolves; every OTHER registered kind will report
	// missing-kind-dir (expected -- counted, not enumerated).
	write("catalog/aws/awss3bucket/v1alpha1/api.proto")
	write("catalog/aws/awss3bucket/v1alpha1/api.pb.go")
	write("catalog/aws/awss3bucket/v1alpha1/spec.proto")             // no stub -> missing-stub
	write("catalog/aws/awss3bucket/v1alpha1/iac/hack/manifest.yaml") // -> unexpected-entry
	write("catalog/aws/awss3bucket/README.md")
	write("catalog/aws/awss3bucket/docs/README.md") // -> unexpected-entry
	write("catalog/aws/awss3bucket/iac/pulumi/main.go")
	write("catalog/aws/awss3bucket/iac/pulumi/Makefile")      // -> forbidden-file
	write("catalog/aws/awss3bucket/iac/tf/.gitignore")        // -> forbidden-file
	write("catalog/aws/awss3bucket/iac/hack/x.sh")            // -> unexpected-entry
	write("catalog/aws/awss3bucket/iac/provider-parity.yaml") // allowed: recorded parity judgment
	write("catalog/aws/awss3bucket/iac/provider-parity.md")   // -> unexpected-entry (only the .yaml is declared)
	write("catalog/aws/awss3bucket/presets/01-basic.yaml")    // no sidecar -> missing-preset-sidecar
	write("catalog/aws/awss3bucket/cost.yaml")                // allowed: the cost profile (controls.yaml deliberately absent)
	write("catalog/aws/notakinddir/spec.proto")               // -> unregistered-kind-dir
	write("catalog/stray-file.txt")                           // -> unexpected-entry (catalog root)
	write("catalog/_compliance/controls-catalog.yaml")        // ignored: non-kind home at the catalog root
	write("catalog/_pricing/estimates/awss3bucket.yaml")      // ignored: non-kind home at the catalog root

	violations, err := Check(root)
	if err != nil {
		t.Fatal(err)
	}
	got := map[string]bool{}
	for _, v := range violations {
		got[v.ID()] = true
	}

	for _, want := range []string{
		"catalog/aws/awss3bucket/v1alpha1:missing-stub",
		"catalog/aws/awss3bucket/v1alpha1:missing-proto", // input/outputs absent
		"catalog/aws/awss3bucket/v1alpha1:missing-spec-test",
		"catalog/aws/awss3bucket/v1alpha1:missing-reference",
		"catalog/aws/awss3bucket/v1alpha1/iac:unexpected-entry",
		"catalog/aws/awss3bucket/docs:unexpected-entry",
		"catalog/aws/awss3bucket/iac/pulumi/Makefile:forbidden-file",
		"catalog/aws/awss3bucket/iac/tf/.gitignore:forbidden-file",
		"catalog/aws/awss3bucket/iac/hack:unexpected-entry",
		"catalog/aws/awss3bucket/iac/provider-parity.md:unexpected-entry",
		"catalog/aws/awss3bucket/iac/pulumi:missing-iac-readme",
		"catalog/aws/awss3bucket/iac/tf:missing-iac-readme",
		"catalog/aws/awss3bucket/presets/01-basic.yaml:missing-preset-sidecar",
		"catalog/aws/awss3bucket:missing-catalog-md",
		"catalog/aws/awss3bucket:missing-logo",
		"catalog/aws/awss3bucket:missing-control-profile",
		"catalog/aws/awss3bucket/iac:missing-permissions",
		"catalog/aws/notakinddir:unregistered-kind-dir",
		"catalog/stray-file.txt:unexpected-entry",
		"AwsEcsService:missing-kind-dir", // representative of the completeness half
	} {
		if !got[want] {
			t.Errorf("expected violation %s to fire", want)
		}
	}

	// The well-formed parts must NOT fire.
	for _, wrong := range []string{
		"catalog/aws/awss3bucket:missing-readme",
		"catalog/aws/awss3bucket:missing-iac",
		"catalog/aws/awss3bucket:missing-cost-profile",
		"catalog/aws/awss3bucket/cost.yaml:unexpected-entry",
		"catalog/aws/awss3bucket/v1alpha1/api.proto:unexpected-entry",
		"catalog/aws/awss3bucket/iac/provider-parity.yaml:unexpected-entry",
		"catalog/_compliance:unregistered-kind-dir",
		"catalog/_pricing:unregistered-kind-dir",
	} {
		if got[wrong] {
			t.Errorf("violation %s must not fire on the well-formed part", wrong)
		}
	}
}

// TestCheck_DeclaredEnginesHoldTheModuleSet proves the tree is held to a kind's
// kind_meta.provisioners in both directions: OpenFGA kinds declare tofu and
// terraform, so a Pulumi module is a violation and its absence is not, while an
// undeclared kind still owes both modules.
func TestCheck_DeclaredEnginesHoldTheModuleSet(t *testing.T) {
	root := t.TempDir()
	write := func(rel string) {
		p := filepath.Join(root, rel)
		if err := os.MkdirAll(filepath.Dir(p), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(p, []byte("x"), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	write("catalog/openfga/openfgastore/iac/tf/README.md")
	write("catalog/openfga/openfgastore/iac/pulumi/README.md") // -> module-for-undeclared-engine
	write("catalog/openfga/openfgaauthorizationmodel/iac/tf/README.md")
	write("catalog/openfga/openfgarelationshiptuple/iac/pulumi/README.md") // -> missing tf, undeclared pulumi
	write("catalog/aws/awss3bucket/iac/tf/README.md")                      // -> missing-pulumi-module

	violations, err := Check(root)
	if err != nil {
		t.Fatal(err)
	}
	got := map[string]bool{}
	for _, v := range violations {
		got[v.ID()] = true
	}
	for _, want := range []string{
		"catalog/openfga/openfgastore/iac/pulumi:module-for-undeclared-engine",
		"catalog/openfga/openfgarelationshiptuple/iac/pulumi:module-for-undeclared-engine",
		"catalog/openfga/openfgarelationshiptuple:missing-tf-module",
		"catalog/aws/awss3bucket:missing-pulumi-module",
	} {
		if !got[want] {
			t.Errorf("expected violation %s to fire", want)
		}
	}
	for _, wrong := range []string{
		"catalog/openfga/openfgastore:missing-pulumi-module",
		"catalog/openfga/openfgaauthorizationmodel:missing-pulumi-module",
		"catalog/openfga/openfgaauthorizationmodel/iac/tf:module-for-undeclared-engine",
		"catalog/aws/awss3bucket/iac/tf:module-for-undeclared-engine",
	} {
		if got[wrong] {
			t.Errorf("violation %s must not fire", wrong)
		}
	}
}

// TestGate mirrors the secretcoverage gate semantics: new drift fails,
// baselined drift passes, a fixed entry left in the baseline fails as stale.
func TestGate(t *testing.T) {
	v := Violation{Path: "catalog/aws/x", Rule: RuleMissingReadme}
	if res := Gate([]Violation{v}, map[string]bool{}); res.OK() || len(res.NewViolations) != 1 {
		t.Errorf("expected new drift to be detected, got %+v", res)
	}
	if res := Gate([]Violation{v}, map[string]bool{v.ID(): true}); !res.OK() {
		t.Errorf("expected baselined drift to pass, got %+v", res)
	}
	if res := Gate(nil, map[string]bool{"catalog/aws/gone:missing-readme": true}); res.OK() || len(res.StaleEntries) != 1 {
		t.Errorf("expected a stale entry to be detected, got %+v", res)
	}
}
