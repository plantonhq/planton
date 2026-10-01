// Lifecycle lanes: scenarios whose PROOF is what happens to a deployed
// component AFTER its first install. The standard lifecycle deploys once,
// verifies, destroys, and verifies absence; three annotations extend it so a
// scenario can prove the promises a module makes about its second act:
//
//   - upgrade: a second manifest (a changed field, typically a version bump)
//     is deployed against the SAME stack after VERIFY-RES, then the kind's
//     verifier checks the deployed state against the second manifest. This
//     is how "bumping the version re-applies the CRDs" becomes a fact.
//   - expect-upgrade-failure: the second deploy must FAIL, for exactly the
//     class the scenario names (a refused schema downgrade, a version that
//     is not published). The engine's error goes to the harness's
//     DeployFailureVerifier, then the original inputs are restored and the
//     first install is destroyed normally.
//   - reinstall: after DESTROY and VERIFY-CLN, the same manifest is deployed
//     AGAIN, verified, destroyed, and verified absent. This is how "kept
//     resources are re-adopted on reinstall" becomes a fact.
//
// All three reuse the engine-specific input binding runValidate performs,
// so a second manifest reaches the engine exactly the way the first did.

package runner

import (
	"context"
	"os"
	"path/filepath"
	"time"

	tt "github.com/gruntwork-io/terratest/modules/terraform"
	"github.com/pkg/errors"
	"github.com/plantonhq/planton/e2e/framework/provider"
	"github.com/plantonhq/planton/pkg/failure"
)

const (
	// UpgradeManifestAnnotation names a second manifest, relative to the
	// scenario file, deployed against the same stack after VERIFY-RES. The
	// two manifests must describe the same resource (same metadata.name);
	// only fields change.
	UpgradeManifestAnnotation = "planton.dev/e2e-upgrade-manifest"

	// ExpectUpgradeFailureAnnotation, together with UpgradeManifestAnnotation,
	// declares that the second deploy must FAIL. The value names the failure
	// class the harness's DeployFailureVerifier must pin (e.g.
	// "crd-schema-downgrade").
	ExpectUpgradeFailureAnnotation = "planton.dev/e2e-expect-upgrade-failure"

	// ReinstallAnnotation ("true") deploys the scenario a second time after
	// the first destroy has been verified, then destroys and verifies again.
	ReinstallAnnotation = "planton.dev/e2e-reinstall"
)

const (
	PhaseUpgrade           Phase = "UPGRADE"
	PhaseVerifyUpgraded    Phase = "VERIFY-UPGRADED"
	PhaseUpgradeExpectFail Phase = "UPGRADE-EXPECT-FAIL"
	PhaseReinstall         Phase = "REINSTALL"
	PhaseVerifyReinstalled Phase = "VERIFY-REINSTALLED"
	PhaseDestroyAgain      Phase = "DESTROY-AGAIN"
	PhaseVerifyClnAgain    Phase = "VERIFY-CLN-AGAIN"
)

// lifecycleAnnotations reads the three lifecycle annotations and validates
// their combination.
type lifecycleAnnotations struct {
	upgradeManifest      string
	expectUpgradeFailure string
	reinstall            bool
	outOfBand            *outOfBandRecovery // nil unless the scenario declares the out-of-band act (outofband.go)
}

func readLifecycleAnnotations(manifestPath string) (lifecycleAnnotations, error) {
	var la lifecycleAnnotations
	upgrade, _ := ManifestAnnotation(manifestPath, UpgradeManifestAnnotation)
	if upgrade != "" {
		la.upgradeManifest = filepath.Join(filepath.Dir(manifestPath), upgrade)
		if _, err := os.Stat(la.upgradeManifest); err != nil {
			return la, errors.Wrapf(err, "%s names %q, which does not exist beside the scenario", UpgradeManifestAnnotation, upgrade)
		}
	}
	la.expectUpgradeFailure, _ = ManifestAnnotation(manifestPath, ExpectUpgradeFailureAnnotation)
	if la.expectUpgradeFailure != "" && la.upgradeManifest == "" {
		return la, errors.Errorf("%s requires %s to name the manifest whose deploy must fail", ExpectUpgradeFailureAnnotation, UpgradeManifestAnnotation)
	}
	reinstall, _ := ManifestAnnotation(manifestPath, ReinstallAnnotation)
	la.reinstall = reinstall == "true"
	if value, _ := ManifestAnnotation(manifestPath, OutOfBandDeleteAnnotation); value != "" {
		recovery, err := parseOutOfBandRecovery(value)
		if err != nil {
			return la, err
		}
		la.outOfBand = recovery
	}
	return la, nil
}

// bindManifest points the engine at a manifest: the Pulumi lane gets a fresh
// stack-input file, the Terraform lane a regenerated terraform.tfvars (and
// provider override) in its existing working directory. runValidate binds the
// first manifest this way; the lifecycle lanes rebind for the second and
// restore the first afterwards, so destroy always runs against the inputs
// that built what it destroys.
func bindManifest(tc *provider.ComponentTestContext, manifestPath string) error {
	providerConfig, err := laneProviderConfig(tc, manifestPath)
	if err != nil {
		return errors.Wrap(err, "provider configuration for the lane")
	}
	switch tc.Engine {
	case "pulumi":
		stackInputPath, err := BuildStackInput(manifestPath, providerConfig)
		if err != nil {
			return errors.Wrap(err, "cannot build stack input from manifest")
		}
		tc.StackInputFilePath = stackInputPath
	case "terraform":
		if tc.TerraformWorkDir == "" {
			return errors.New("terraform working directory not prepared (runValidate must run first)")
		}
		input, err := BuildTerraformInput(manifestPath, tc.TerraformWorkDir, providerConfig)
		if err != nil {
			return errors.Wrap(err, "cannot build terraform input from manifest")
		}
		tc.TerraformOpts = BuildTerratestOptions(tc.T, tc.TerraformWorkDir, input.TfvarsPath, input.EnvVars)
	default:
		return errors.Errorf("unsupported engine: %s", tc.Engine)
	}
	return nil
}

// prepareSecondAct turns the second act's file, as authored beside the
// scenario, into the manifest the upgrade deploys: its tokens expand to the
// FIRST act's values (the same run id, scenario slug and lane clock, so a name
// built from them names the same object and only the declared fields change),
// and its value_from refs resolve against the same deployed prerequisites.
// Without this, a second act could only be written for a scenario that
// carries neither tokens nor refs.
func prepareSecondAct(path, runID, firstActScenario string, clock time.Time, depOutputs DependencyOutputs) (string, error) {
	expanded, err := ExpandManifestTokens(path, runID, firstActScenario, clock)
	if err != nil {
		return "", errors.Wrap(err, "expanding the second act's tokens")
	}
	resolved, err := ResolveManifestRefs(expanded, depOutputs)
	return resolved, errors.Wrap(err, "resolving the second act's references from the prerequisites' outputs")
}

// runUpgrade deploys the second manifest against the same stack.
func runUpgrade(tc *provider.ComponentTestContext, upgradeManifest string) error {
	if err := bindManifest(tc, upgradeManifest); err != nil {
		return errors.Wrap(err, "binding the upgrade manifest")
	}
	return errors.Wrap(runDeploy(tc), "the upgrade deploy failed")
}

// runUpgradeExpectFailure deploys the second manifest expecting the engine to
// refuse it, hands the error to the harness for cause pinning, and restores
// the first manifest's inputs whatever happens, so the destroy that follows
// tears down what the first deploy built.
func runUpgradeExpectFailure(ctx context.Context, tc *provider.ComponentTestContext, harness provider.Harness,
	upgradeManifest, expectation, originalManifest string) (err error) {
	defer func() {
		if restoreErr := bindManifest(tc, originalManifest); restoreErr != nil && err == nil {
			err = errors.Wrap(restoreErr, "restoring the original manifest after the expected upgrade failure")
		}
	}()

	fv, ok := harness.(provider.DeployFailureVerifier)
	if !ok {
		return errors.Errorf("scenario expects an upgrade failure (%s) but the %s harness does not implement provider.DeployFailureVerifier",
			expectation, tc.Provider)
	}
	if bindErr := bindManifest(tc, upgradeManifest); bindErr != nil {
		return errors.Wrap(bindErr, "binding the upgrade manifest")
	}
	deployErr := failure.AnnotateError(runDeploy(tc))
	if deployErr == nil {
		return errors.Errorf("the scenario expects the upgrade to fail (%s), but it succeeded", expectation)
	}
	upgradeCtx := context.WithValue(ctx, provider.ManifestPathKey{}, upgradeManifest)
	return errors.Wrap(fv.VerifyExpectedDeployFailure(upgradeCtx, tc, expectation, deployErr),
		"the upgrade failed as expected, but the failure could not be attributed to the expected cause")
}

// runReinstall deploys the original manifest again after a verified destroy.
// The Terraform state is empty and the Pulumi stack exists with no
// resources, so this is a plain deploy from the engine's point of view; what
// it proves is on the cluster (kept resources adopted, nothing conflicting).
func runReinstall(tc *provider.ComponentTestContext) error {
	if tc.Engine == "terraform" {
		if _, ok := tc.TerraformOpts.(*tt.Options); !ok {
			return errors.New("terraform options not initialized (runValidate must run first)")
		}
	}
	return errors.Wrap(runDeploy(tc), "the reinstall deploy failed")
}
