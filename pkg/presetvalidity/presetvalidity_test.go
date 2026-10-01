package presetvalidity

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
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

// TestPresetValidityGate is the CI guardrail: the live walk over every
// catalog preset must not introduce a violation outside the baseline or
// leave a stale baseline entry. On failure, fix the preset (the detail names
// the exact rejection a user copying it would hit) or -- for a preset whose
// repair is deliberately routed to its provider's sweep batch -- regenerate
// the baseline with PLANTON_REGEN_PRESET_VALIDITY_BASELINE=1 and justify the
// growth in review.
func TestPresetValidityGate(t *testing.T) {
	root := repoRoot(t)
	if _, err := os.Stat(filepath.Join(root, "catalog")); err != nil {
		t.Skip("catalog source tree not present (bazel sandbox); runs under go test and the lint.preset-validity lane")
	}

	violations, err := Check(root)
	if err != nil {
		t.Fatalf("preset walk: %v", err)
	}

	_, thisFile, _, _ := runtime.Caller(0)
	baselinePath := filepath.Join(filepath.Dir(thisFile), "baseline.yaml")
	if os.Getenv("PLANTON_REGEN_PRESET_VALIDITY_BASELINE") == "1" {
		if err := WriteBaseline(baselinePath, violations); err != nil {
			t.Fatalf("write baseline: %v", err)
		}
		t.Logf("baseline regenerated -- review the diff before committing")
		return
	}

	baseline, err := LoadBaseline(baselinePath)
	if err != nil {
		t.Fatalf("load baseline: %v", err)
	}
	res := Gate(violations, baseline)
	for _, v := range res.NewViolations {
		t.Errorf("preset shipped invalid: %s -- %s", v.ID(), v.Detail)
	}
	for _, id := range res.StaleEntries {
		t.Errorf("stale baseline entry (no longer a violation): %s -- remove it from baseline.yaml", id)
	}
	for _, id := range res.ForbiddenEntries {
		t.Errorf("baseline.yaml lists %s, but that rule accepts no baseline entries: the entry would hide a guide that misleads its reader -- remove the entry and fix the preset's guide", id)
	}
}

// torturePreset reads the torture kind's default preset -- the repository's
// canonical known-good manifest -- as the hermetic green against which each
// red below is one deliberate defect. A gate that cannot fail teaches false
// confidence.
func torturePreset(t *testing.T) string {
	t.Helper()
	path := filepath.Join(repoRoot(t), "catalog", "_test", "testcloudresourcegeneric", "presets", "01-default.yaml")
	content, err := os.ReadFile(path)
	if err != nil {
		t.Skipf("torture preset not present: %v", err)
	}
	return string(content)
}

// TestCheckPreset_HermeticFixture proves the rule fires against deliberately
// broken presets and stays quiet on the canonical valid one.
func TestCheckPreset_HermeticFixture(t *testing.T) {
	green := torturePreset(t)

	if vs := CheckPreset("catalog/_test/x/presets/01-green.yaml", []byte(green)); len(vs) != 0 {
		t.Errorf("the canonical torture preset must pass, got %v", vs)
	}

	// An angle-bracket placeholder where the schema demands a real shape:
	// the exact defect class this gate exists to make unshippable.
	brokenValue := strings.Replace(green, "int32Field: 7", `int32Field: "<replace-me>"`, 1)
	if vs := CheckPreset("catalog/_test/x/presets/02-broken.yaml", []byte(brokenValue)); len(vs) != 1 || vs[0].Rule != RuleInvalidPreset {
		t.Errorf("expected one %s violation for a type-broken placeholder, got %v", RuleInvalidPreset, vs)
	}

	// A wrong envelope: the manifest declares an apiVersion its kind does
	// not serve. The validator's envelope contract catches it here.
	brokenEnvelope := strings.Replace(green, "apiVersion: _test.planton.dev/v1alpha2", "apiVersion: _test.planton.dev/v999", 1)
	if vs := CheckPreset("catalog/_test/x/presets/03-envelope.yaml", []byte(brokenEnvelope)); len(vs) != 1 || vs[0].Rule != RuleInvalidPreset {
		t.Errorf("expected one %s violation for a wrong envelope, got %v", RuleInvalidPreset, vs)
	}

	// A required field removed: the load succeeds, validation rejects.
	brokenRequired := strings.Replace(green, "  requiredRef:\n    value: literal-required-value\n", "", 1)
	if vs := CheckPreset("catalog/_test/x/presets/04-required.yaml", []byte(brokenRequired)); len(vs) != 1 || vs[0].Rule != RuleInvalidPreset {
		t.Errorf("expected one %s violation for a missing required field, got %v", RuleInvalidPreset, vs)
	}
}

// TestGate mirrors the anatomy and catalogpage gate semantics: new drift
// fails, baselined drift passes, a fixed entry left in the baseline fails
// as stale.
func TestGate(t *testing.T) {
	v := Violation{Path: "catalog/aws/x/presets/01-a.yaml", Rule: RuleInvalidPreset}
	if res := Gate([]Violation{v}, map[string]bool{}); res.OK() || len(res.NewViolations) != 1 {
		t.Errorf("expected new drift to be detected, got %+v", res)
	}
	if res := Gate([]Violation{v}, map[string]bool{v.ID(): true}); !res.OK() {
		t.Errorf("expected baselined drift to pass, got %+v", res)
	}
	if res := Gate(nil, map[string]bool{"catalog/aws/gone/presets/01-a.yaml:invalid-preset": true}); res.OK() || len(res.StaleEntries) != 1 {
		t.Errorf("expected a stale entry to be detected, got %+v", res)
	}
	if res := Gate([]Violation{v, v}, map[string]bool{v.ID(): true}); !res.OK() {
		t.Errorf("expected duplicate-rule collapse, got %+v", res)
	}
}

// TestPhantomPlaceholderRows proves the placeholder-table rule reads the
// table the guides actually carry: the rows under "Placeholders to Replace"
// (or "Placeholders"), the placeholders in each row's first cell, and
// nothing outside that section.
func TestPhantomPlaceholderRows(t *testing.T) {
	manifest := []byte("spec:\n  region: <aws-region>\n  subnets:\n    - value: <private-subnet-a>\n    - value: <private-subnet-b>\n  # <kms-key-arn> when encryption is on\n")
	guide := []byte(`# A preset

Uses ` + "`<not-in-a-table>`" + ` in prose, which is not a row.

## Key Configuration Choices

| Field | Why |
|-------|-----|
| ` + "`<also-not-the-table>`" + ` | a different section |

## Placeholders to Replace

| Placeholder | Description | Where to Find |
|-------------|-------------|---------------|
| ` + "`<aws-region>`" + ` | The region | Console |
| ` + "`<private-subnet-a>`" + ` / ` + "`<private-subnet-b>`" + ` | Two subnets | VPC |
| ` + "`<kms-key-arn>`" + ` | Named in a manifest comment | KMS |
| ` + "`<vpc-id>`" + ` | Gone from the manifest | VPC |
| ` + "`<private-subnet-a/b>`" + ` | Two placeholders spelled as one | VPC |
| ` + "`123456789012`" + ` | A pattern-valid placeholder, not checked | IAM |

## Related Presets

| ` + "`<after-the-table>`" + ` | not a placeholder row |
`)
	rows := phantomPlaceholderRows(guide, manifest)
	want := []phantomRow{{line: 18, placeholder: "<vpc-id>"}, {line: 19, placeholder: "<private-subnet-a/b>"}}
	if len(rows) != len(want) {
		t.Fatalf("expected %v, got %v", want, rows)
	}
	for i := range want {
		if rows[i] != want[i] {
			t.Errorf("row %d: expected %v, got %v", i, want[i], rows[i])
		}
	}

	vs := checkPlaceholderTable("catalog/aws/x/presets/01-a.yaml", guide, manifest)
	if len(vs) != 1 || vs[0].Rule != RulePhantomPlaceholder {
		t.Fatalf("expected one %s violation, got %v", RulePhantomPlaceholder, vs)
	}
	for _, part := range []string{"catalog/aws/x/presets/01-a.md", "01-a.yaml", "line 18 names <vpc-id>", "line 19 names <private-subnet-a/b>", "delete each row", "reword a row", "add the placeholder to the manifest"} {
		if !strings.Contains(vs[0].Detail, part) {
			t.Errorf("detail lacks %q: %s", part, vs[0].Detail)
		}
	}

	if vs := checkPlaceholderTable("catalog/aws/x/presets/01-a.yaml", []byte("# No table\n"), manifest); len(vs) != 0 {
		t.Errorf("a guide without a placeholder table has nothing to check, got %v", vs)
	}
}

// TestGate_PhantomPlaceholderAcceptsNoBaseline proves the placeholder-table
// rule cannot be baselined: a listed violation still fails, the listing
// itself fails, and a regenerated baseline never writes one.
func TestGate_PhantomPlaceholderAcceptsNoBaseline(t *testing.T) {
	v := Violation{Path: "catalog/aws/x/presets/01-a.yaml", Rule: RulePhantomPlaceholder}
	res := Gate([]Violation{v}, map[string]bool{v.ID(): true})
	if len(res.NewViolations) != 1 || len(res.ForbiddenEntries) != 1 || res.OK() {
		t.Errorf("a baselined phantom placeholder must still fail and its entry must be refused, got %+v", res)
	}
	if res := Gate(nil, map[string]bool{v.ID(): true}); res.OK() || len(res.ForbiddenEntries) != 1 || len(res.StaleEntries) != 0 {
		t.Errorf("a baseline entry for the rule is refused, not reported stale, got %+v", res)
	}

	path := filepath.Join(t.TempDir(), "baseline.yaml")
	invalid := Violation{Path: "catalog/aws/x/presets/01-a.yaml", Rule: RuleInvalidPreset}
	if err := WriteBaseline(path, []Violation{v, invalid}); err != nil {
		t.Fatal(err)
	}
	written, err := LoadBaseline(path)
	if err != nil {
		t.Fatal(err)
	}
	if written[v.ID()] || !written[invalid.ID()] {
		t.Errorf("a regenerated baseline must keep %s and never write %s, got %v", invalid.ID(), v.ID(), written)
	}
}
