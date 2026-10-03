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

// bareHarness implements ONLY the base Harness interface -- the common case
// against which the failure-mode capabilities must fail loudly, never
// silently pass.
type bareHarness struct{}

func (bareHarness) Setup(context.Context) error    { return nil }
func (bareHarness) Teardown(context.Context) error { return nil }
func (bareHarness) VerifyDeployed(context.Context, string, map[string]interface{}) error {
	return nil
}
func (bareHarness) VerifyDestroyed(context.Context, string) error { return nil }

// causeHarness adds the RuntimeCauseVerifier capability and records the call.
type causeHarness struct {
	bareHarness
	gotCause string
	fail     bool
}

func (h *causeHarness) VerifyRuntimeFailureCause(_ context.Context, _ *provider.KindTestContext, cause string) error {
	h.gotCause = cause
	if h.fail {
		return errors.New("cause mismatch")
	}
	return nil
}

func TestRunVerifyRuntimeCause(t *testing.T) {
	tc := &provider.KindTestContext{Kind: "x", Provider: "azure"}

	// A harness without the capability must fail loudly, naming the gap.
	err := runVerifyRuntimeCause(context.Background(), tc, bareHarness{}, "refused-join")
	if err == nil || !strings.Contains(err.Error(), "RuntimeCauseVerifier") {
		t.Fatalf("expected a capability-missing error, got %v", err)
	}

	// A harness with the capability receives the annotation's cause value.
	h := &causeHarness{}
	if err := runVerifyRuntimeCause(context.Background(), tc, h, "refused-join"); err != nil {
		t.Fatalf("expected pass-through success, got %v", err)
	}
	if h.gotCause != "refused-join" {
		t.Fatalf("cause = %q, want refused-join", h.gotCause)
	}

	// A failing cause verification propagates.
	h.fail = true
	if err := runVerifyRuntimeCause(context.Background(), tc, h, "refused-join"); err == nil {
		t.Fatal("expected the cause-verification failure to propagate")
	}
}

func TestRunExpectDeployFailure_CapabilityGate(t *testing.T) {
	// The capability check fires BEFORE any deploy is attempted, so a
	// mis-wired scenario fails in milliseconds, not after a cloud apply.
	tc := &provider.KindTestContext{Kind: "x", Provider: "gcp"}
	err := runExpectDeployFailure(context.Background(), tc, bareHarness{}, "revision-readiness")
	if err == nil || !strings.Contains(err.Error(), "DeployFailureVerifier") {
		t.Fatalf("expected a capability-missing error, got %v", err)
	}
}

func TestFailureModeAnnotations_MutuallyExclusive(t *testing.T) {
	dir := t.TempDir()
	manifest := filepath.Join(dir, "both.yaml")
	content := `apiVersion: azure.planton.dev/v1alpha1
kind: AzurePlantonRunner
metadata:
  name: both-modes
  annotations:
    planton.dev/e2e-expect-deploy-failure: revision-readiness
    planton.dev/e2e-expected-runtime-failure: refused-join
spec: {}
`
	if err := os.WriteFile(manifest, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}

	// RepoRoot empty skips dependency deployment, so the run reaches the
	// annotation gate hermetically.
	tc := &provider.KindTestContext{
		Kind:         "azureplantonrunner",
		Provider:     "azure",
		Engine:       "pulumi",
		ManifestPath: manifest,
		RunID:        "t1",
	}
	res := RunKindTest(context.Background(), tc, bareHarness{})
	if res.Passed {
		t.Fatal("expected the run to fail on mutually exclusive annotations")
	}
	found := false
	for _, p := range res.Phases {
		if p.Error != nil && strings.Contains(p.Error.Error(), "mutually exclusive") {
			found = true
		}
	}
	if !found {
		t.Fatalf("expected the mutual-exclusion error in phases, got %+v", res.Phases)
	}
}
