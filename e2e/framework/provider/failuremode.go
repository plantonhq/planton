// Failure-mode verification capabilities -- optional Harness extensions the
// runner discovers by type assertion. They exist for scenarios whose PROOF is a
// deliberate, precisely-attributed failure: a component deployed with a
// credential that a real service rejects (the canonical case: a Planton runner
// appliance with a fake enrollment token). The evidence bar is cause-pinning
// with the provider's own APIs -- "everything worked except exactly the thing
// the scenario sabotaged" -- never a bare acceptance of any failure.

package provider

import "context"

// RuntimeCauseVerifier verifies, AFTER a successful deploy, that a workload's
// designed runtime failure has exactly the expected cause. Activated by the
// scenario annotation `planton.dev/e2e-expected-runtime-failure`; the
// annotation's value is passed as cause (e.g. "refused-join") so a harness can
// support several causes over time. Runs between VERIFY-RES and DESTROY, so
// outputs are available on tc. Implementations own their polling: runtime
// failures take time to surface (a crash-looping container needs a restart or
// two before its state and logs attest the cause).
type RuntimeCauseVerifier interface {
	VerifyRuntimeFailureCause(ctx context.Context, tc *KindTestContext, cause string) error
}

// DeployFailureVerifier verifies that a deploy which FAILED failed for exactly
// the expected reason, on substrates that gate resource creation on workload
// health (e.g. Cloud Run v2 gates service creation on first-revision
// readiness). Activated by the scenario annotation
// `planton.dev/e2e-expect-deploy-failure`; the annotation's value is passed as
// expectation. deployErr carries the engine's full error output for
// classification. Outputs do NOT exist when this runs -- implementations
// derive identity from the manifest (via ManifestPathKey on ctx or
// tc.ManifestPath) and must verify the partially-created resource's state with
// the provider's own APIs BEFORE the runner destroys it.
type DeployFailureVerifier interface {
	VerifyExpectedDeployFailure(ctx context.Context, tc *KindTestContext, expectation string, deployErr error) error
}
