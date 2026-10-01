package runner

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/pkg/errors"
	"github.com/plantonhq/planton/e2e/framework/provider"
)

func TestParseOutOfBandRecovery(t *testing.T) {
	cases := []struct {
		value   string
		forget  []string
		wantErr string
	}{
		{value: "recreates"},
		{value: " recreates "},
		{value: "forget:stripe_coupon.this", forget: []string{"stripe_coupon.this"}},
		{value: "forget: a.this , b.this[\"k\"] ", forget: []string{"a.this", `b.this["k"]`}},
		{value: "forget:", wantErr: "names no state address"},
		{value: "forget: , ", wantErr: "names no state address"},
		{value: "recreate", wantErr: "is neither"},
		{value: "state-rm:a.this", wantErr: "is neither"},
	}
	for _, c := range cases {
		got, err := parseOutOfBandRecovery(c.value)
		if c.wantErr != "" {
			if err == nil || !strings.Contains(err.Error(), c.wantErr) {
				t.Errorf("%q: want an error containing %q, got %v", c.value, c.wantErr, err)
			}
			continue
		}
		if err != nil {
			t.Errorf("%q: unexpected error %v", c.value, err)
			continue
		}
		if strings.Join(got.forget, "|") != strings.Join(c.forget, "|") {
			t.Errorf("%q: forget = %v, want %v", c.value, got.forget, c.forget)
		}
	}
}

// The plan's detailed exit code must match the GUIDE's sentence exactly; each
// mismatch is a failure that names the sentence to fix, never a pass.
func TestJudgeDriftPlan(t *testing.T) {
	forget := &outOfBandRecovery{forget: []string{"stripe_coupon.this"}}
	recreates := &outOfBandRecovery{}
	cases := []struct {
		name     string
		recovery *outOfBandRecovery
		exitCode int
		runErr   error
		wantErr  string
	}{
		{name: "forget: the plan fails, as declared", recovery: forget, exitCode: 1},
		{name: "forget: the plan proposes a create instead", recovery: forget, exitCode: 2, wantErr: `should say "recreates"`},
		{name: "forget: the plan sees no change", recovery: forget, exitCode: 0, wantErr: "exited 0"},
		{name: "recreates: the plan proposes the create, as declared", recovery: recreates, exitCode: 2},
		{name: "recreates: the plan fails instead", recovery: recreates, exitCode: 1, wantErr: "must teach forgetting it first"},
		{name: "recreates: the plan does not notice", recovery: recreates, exitCode: 0, wantErr: "did not notice"},
		{name: "the plan could not run at all", recovery: forget, exitCode: 1, runErr: errors.New("binary missing"), wantErr: "binary missing"},
	}
	for _, c := range cases {
		err := judgeDriftPlan(c.recovery, c.exitCode, c.runErr)
		if c.wantErr == "" {
			if err != nil {
				t.Errorf("%s: unexpected error %v", c.name, err)
			}
			continue
		}
		if err == nil || !strings.Contains(err.Error(), c.wantErr) {
			t.Errorf("%s: want an error containing %q, got %v", c.name, c.wantErr, err)
		}
	}
}

type deletingHarness struct {
	bareHarness
	deleted bool
}

func (h *deletingHarness) DeleteOutOfBand(context.Context, *provider.ComponentTestContext) error {
	h.deleted = true
	return nil
}

func TestRunOutOfBandDelete_CapabilityGate(t *testing.T) {
	tc := &provider.ComponentTestContext{Component: "stripecoupon", Provider: "stripe"}

	// A harness without the capability fails loudly, naming the gap, instead
	// of passing an act that deleted nothing.
	err := runOutOfBandDelete(context.Background(), tc, bareHarness{})
	if err == nil || !strings.Contains(err.Error(), "OutOfBandDeleter") {
		t.Fatalf("expected a capability-missing error, got %v", err)
	}

	h := &deletingHarness{}
	if err := runOutOfBandDelete(context.Background(), tc, h); err != nil {
		t.Fatalf("expected the harness's delete to run, got %v", err)
	}
	if !h.deleted {
		t.Fatal("the harness's DeleteOutOfBand was not called")
	}
}

func TestRequireOutOfBandEngine(t *testing.T) {
	if err := requireOutOfBandEngine("terraform"); err != nil {
		t.Fatalf("an HCL lane must be accepted, got %v", err)
	}
	err := requireOutOfBandEngine("pulumi")
	if err == nil || !strings.Contains(err.Error(), "OpenTofu and Terraform lanes only") {
		t.Fatalf("a Pulumi lane must be refused with the reason, got %v", err)
	}
}

func TestReadLifecycleAnnotations_OutOfBand(t *testing.T) {
	dir := t.TempDir()
	write := func(name, annotation string) string {
		path := filepath.Join(dir, name)
		content := "apiVersion: stripe.planton.dev/v1\nkind: StripeCoupon\nmetadata:\n  name: x\n  annotations:\n    " +
			annotation + "\nspec: {}\n"
		if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
			t.Fatal(err)
		}
		return path
	}

	la, err := readLifecycleAnnotations(write("forget.yaml", OutOfBandDeleteAnnotation+": forget:stripe_coupon.this"))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if la.outOfBand == nil || la.outOfBand.String() != "forget:stripe_coupon.this" {
		t.Fatalf("outOfBand = %v, want forget:stripe_coupon.this", la.outOfBand)
	}

	if _, err := readLifecycleAnnotations(write("typo.yaml", OutOfBandDeleteAnnotation+": forgot:stripe_coupon.this")); err == nil {
		t.Fatal("a malformed out-of-band annotation must be refused before anything deploys")
	}

	la, err = readLifecycleAnnotations(write("none.yaml", "planton.dev/e2e-required-env: X"))
	if err != nil || la.outOfBand != nil {
		t.Fatalf("a scenario without the annotation has no act; got %v, %v", la.outOfBand, err)
	}
}
