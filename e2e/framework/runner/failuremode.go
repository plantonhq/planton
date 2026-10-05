// Failure-mode lanes: scenarios whose PROOF is a deliberate, precisely
// attributed failure. Two shapes, activated per scenario by manifest
// annotation, both dispatching to optional harness capabilities (see the
// provider package's failuremode.go):
//
//   - expected-deploy-failure, for substrates that gate resource creation on
//     workload health (Cloud Run gates service creation on first-revision
//     readiness): DEPLOY must FAIL, and the harness then proves it failed for
//     exactly the expected reason -- with the partially-created resource's
//     state and the workload's own logs -- before the stack is destroyed.
//   - expected-runtime-failure, for substrates that accept the deploy and fail
//     asynchronously (ECS tasks, Container Apps replicas, Kubernetes pods):
//     the standard lifecycle runs, plus a VERIFY-CAUSE phase after VERIFY-RES
//     asserting the workload's failure has exactly the expected cause (e.g. a
//     refused control-plane join, never an image pull failure).
//
// The evidence bar is cause-pinning with the provider's own APIs. A lane that
// merely tolerates "some failure" proves nothing; these lanes prove
// "everything worked except exactly the thing the scenario sabotaged".

package runner

import (
	"context"

	"github.com/pkg/errors"
	"github.com/plantonhq/planton/e2e/framework/provider"
	"github.com/plantonhq/planton/pkg/failure"
)

const (
	// ExpectDeployFailureAnnotation marks a scenario whose DEPLOY must fail.
	// The value names the expected failure class (passed to the harness's
	// DeployFailureVerifier), e.g. "revision-readiness".
	ExpectDeployFailureAnnotation = "planton.dev/e2e-expect-deploy-failure"

	// ExpectedRuntimeFailureAnnotation marks a scenario whose deploy succeeds
	// while its workload fails by design. The value names the expected cause
	// (passed to the harness's RuntimeCauseVerifier), e.g. "refused-join".
	ExpectedRuntimeFailureAnnotation = "planton.dev/e2e-expected-runtime-failure"
)

// runExpectDeployFailure runs DEPLOY expecting it to FAIL, then hands the
// engine's error to the harness for cause classification and post-mortem
// state assertions. A deploy that SUCCEEDS fails the phase loudly: either the
// sabotaged input no longer fails on this substrate (the scenario needs
// updating) or the substrate stopped gating on workload health (a real
// contract change worth catching). Either way the created resources are
// destroyed by the pipeline's failure-cleanup path.
func runExpectDeployFailure(ctx context.Context, tc *provider.KindTestContext, harness provider.Harness, expectation string) error {
	fv, ok := harness.(provider.DeployFailureVerifier)
	if !ok {
		return errors.Errorf("scenario expects a deploy failure (%s) but the %s harness does not implement provider.DeployFailureVerifier",
			expectation, tc.Provider)
	}

	// The engine's output is explained before the verifier sees it, so a lane
	// asserts the exact text a CLI user reads: the module's own refusal where
	// it could speak, and the runner layer's explanation of a provider's raw
	// text where it could not.
	deployErr := failure.AnnotateError(runDeploy(tc))
	if deployErr == nil {
		return errors.Errorf("the scenario expects the deploy to fail (%s), but it succeeded", expectation)
	}
	return errors.Wrap(fv.VerifyExpectedDeployFailure(ctx, tc, expectation, deployErr),
		"deploy failed as expected, but the failure could not be attributed to the expected cause")
}

// runVerifyRuntimeCause asserts a deployed-but-designed-to-fail workload's
// failure has exactly the expected cause, via the harness's optional
// RuntimeCauseVerifier capability.
func runVerifyRuntimeCause(ctx context.Context, tc *provider.KindTestContext, harness provider.Harness, cause string) error {
	rv, ok := harness.(provider.RuntimeCauseVerifier)
	if !ok {
		return errors.Errorf("scenario expects a runtime failure cause (%s) but the %s harness does not implement provider.RuntimeCauseVerifier",
			cause, tc.Provider)
	}
	return rv.VerifyRuntimeFailureCause(ctx, tc, cause)
}
