//go:build !codegen
// +build !codegen

package secretcoverage

import (
	"testing"

	"github.com/plantonhq/planton/apis/dev/planton/shared/cloudresourcekind"
	"github.com/plantonhq/planton/pkg/crkreflect"
)

// TestSecretCoverageGate is the CI guardrail. The live scan over all production kinds
// must not introduce a gap outside the baseline, leave a stale baseline entry, or carry
// a self-contradictory annotation. On failure: run `planton secret-coverage` to see the
// detail, then annotate the field `sensitive`, exempt it with `sensitive_exempt_reason`,
// or (after annotating) regenerate the baseline with `--write-baseline`.
func TestSecretCoverageGate(t *testing.T) {
	baseline, err := LoadBaseline("baseline.yaml")
	if err != nil {
		t.Fatalf("load baseline: %v", err)
	}
	res := Gate(Analyze(), baseline)
	for _, id := range res.NewGaps {
		t.Errorf("new unannotated sensitive-looking field: %s -- annotate `sensitive` or exempt with `sensitive_exempt_reason`", id)
	}
	for _, id := range res.StaleEntries {
		t.Errorf("stale baseline entry (no longer a gap): %s -- remove it from baseline.yaml", id)
	}
	for _, f := range res.AnnotationViolations {
		for _, v := range f.Violations {
			t.Errorf("%s:%s -- %s", f.Kind, f.Path, v)
		}
	}
}

func TestClassify(t *testing.T) {
	cases := []struct {
		name           string
		fieldName      string
		isSensitive    bool
		exemptReason   string
		wantClass      Classification
		wantViolations int
	}{
		{"annotated secret", "password", true, "", Covered, 0},
		{"annotation wins over neutral name", "value", true, "", Covered, 0},
		{"exempt false positive", "token_dialect", false, "format selector", Exempt, 0},
		{"unannotated secret-looking name", "password", false, "", Gap, 0},
		{"neutral name", "region", false, "", NotSensitive, 0},
		{"sensitive + exempt is a contradiction", "password", true, "because", Covered, 1},
		{"exemption on a non-heuristic name is pointless", "region", false, "no reason", Exempt, 1},
	}
	for _, tc := range cases {
		gotClass, gotViol := classify(tc.fieldName, tc.isSensitive, tc.exemptReason)
		if gotClass != tc.wantClass {
			t.Errorf("%s: class = %q, want %q", tc.name, gotClass, tc.wantClass)
		}
		if len(gotViol) != tc.wantViolations {
			t.Errorf("%s: violations = %d, want %d (%v)", tc.name, len(gotViol), tc.wantViolations, gotViol)
		}
	}
}

// TestCollectFindings_HermeticFixture proves the descriptor walk against the permanent
// testcloudresourcegeneric fixture: a sensitive raw string AND a sensitive
// StringValueOrRef are both COVERED leaves (the StringValueOrRef is a single leaf, not
// recursed into), while non-sensitive strings, maps, repeated, and nested messages
// produce no findings.
func TestCollectFindings_HermeticFixture(t *testing.T) {
	msg, err := crkreflect.NewInstance(cloudresourcekind.CloudResourceKind_TestCloudResourceGeneric)
	if err != nil {
		t.Fatalf("new instance: %v", err)
	}
	specField := msg.ProtoReflect().Descriptor().Fields().ByName("spec")
	if specField == nil {
		t.Fatal("fixture has no spec field")
	}

	got := map[string]Classification{}
	for _, f := range CollectFindings(specField.Message(), "TestCloudResourceGeneric", "_test") {
		got[f.Path] = f.Class
	}

	want := map[string]Classification{
		"spec.sensitive_string": Covered,
		"spec.sensitive_ref":    Covered,
	}
	if len(got) != len(want) {
		t.Fatalf("findings = %v, want exactly %v", got, want)
	}
	for path, cls := range want {
		if got[path] != cls {
			t.Errorf("path %s = %q, want %q", path, got[path], cls)
		}
	}
}

func TestGate(t *testing.T) {
	gap := Finding{Kind: "AwsRdsInstance", Path: "spec.password", Class: Gap}
	contradiction := Finding{Kind: "AwsX", Path: "spec.secret", Class: Covered, Violations: []string{"contradiction"}}

	if res := Gate([]Finding{gap}, map[string]bool{}); res.OK() || len(res.NewGaps) != 1 {
		t.Errorf("expected a new gap to be detected, got %+v", res)
	}
	if res := Gate([]Finding{gap}, map[string]bool{"AwsRdsInstance:spec.password": true}); !res.OK() {
		t.Errorf("expected a baselined gap to pass, got %+v", res)
	}
	if res := Gate(nil, map[string]bool{"Old:spec.gone": true}); res.OK() || len(res.StaleEntries) != 1 {
		t.Errorf("expected a stale baseline entry to be detected, got %+v", res)
	}
	if res := Gate([]Finding{contradiction}, map[string]bool{}); res.OK() || len(res.AnnotationViolations) != 1 {
		t.Errorf("expected an annotation violation to be detected, got %+v", res)
	}
}
