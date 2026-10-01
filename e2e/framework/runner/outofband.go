// The out-of-band act: someone deletes the deployed object in the provider's
// own console, behind the engine's back. Every GUIDE states what the next plan
// does with that -- either it notices the object is gone and offers to create
// it again, or (a provider without "not found" handling) the read fails and
// the plan errors until the object is forgotten from state. This act proves
// the sentence the GUIDE wrote, and then proves the recovery it teaches brings
// the declared object back.
//
// The phases slot after VERIFY-RES (and the import round-trip, when on):
//
//	OUT-OF-BAND-DELETE  the harness deletes the object through the provider's
//	                    API (provider.OutOfBandDeleter) and confirms it is gone
//	DRIFT-PLAN          the engine plans against the now-stale state; the
//	                    outcome must match the declared recovery
//	RECOVER             the GUIDE's recovery: forget the named addresses
//	                    (state rm) when declared, then apply
//	VERIFY-RECOVERED    outputs and the kind's verifier again, on the object
//	                    the recovery created
//
// DESTROY and VERIFY-CLN then run on the recovered object as usual.

package runner

import (
	"context"
	"fmt"
	"strings"

	tt "github.com/gruntwork-io/terratest/modules/terraform"
	"github.com/pkg/errors"
	"github.com/plantonhq/planton/e2e/framework/provider"
)

// OutOfBandDeleteAnnotation opts a scenario into the out-of-band act. Its
// value is the recovery the kind's GUIDE teaches, in one of two forms:
//
//	recreates                 the plan proposes creating the object again,
//	                          and a plain apply brings it back
//	forget:<addr>[,<addr>...] the plan FAILS reading the missing object; the
//	                          recovery forgets those state addresses, then
//	                          applies (the GUIDE's `tofu state rm` + apply)
//
// The act runs on HCL engines only: it reads the engine's plan exit code and
// edits its state, which have no Pulumi equivalent in this runner yet, so a
// Pulumi lane carrying the annotation is refused rather than silently skipped.
const OutOfBandDeleteAnnotation = "planton.dev/e2e-out-of-band-delete"

const (
	PhaseOutOfBandDelete Phase = "OUT-OF-BAND-DELETE"
	PhaseDriftPlan       Phase = "DRIFT-PLAN"
	PhaseRecover         Phase = "RECOVER"
	PhaseVerifyRecovered Phase = "VERIFY-RECOVERED"
)

// The annotation's two value forms.
const (
	outOfBandRecreates    = "recreates"
	outOfBandForgetPrefix = "forget:"
)

// outOfBandRecovery is the parsed annotation: either the plan recreates, or
// the plan fails and the named addresses are forgotten before the apply.
type outOfBandRecovery struct {
	forget []string // empty means "recreates"
}

func (r outOfBandRecovery) String() string {
	if len(r.forget) == 0 {
		return outOfBandRecreates
	}
	return outOfBandForgetPrefix + strings.Join(r.forget, ",")
}

// parseOutOfBandRecovery reads the annotation's value. An empty value, an
// unknown form, or a forget with no address is an authoring error, reported
// before anything deploys.
func parseOutOfBandRecovery(value string) (*outOfBandRecovery, error) {
	value = strings.TrimSpace(value)
	switch {
	case value == outOfBandRecreates:
		return &outOfBandRecovery{}, nil
	case strings.HasPrefix(value, outOfBandForgetPrefix):
		var addrs []string
		for _, a := range strings.Split(strings.TrimPrefix(value, outOfBandForgetPrefix), ",") {
			if a = strings.TrimSpace(a); a != "" {
				addrs = append(addrs, a)
			}
		}
		if len(addrs) == 0 {
			return nil, errors.Errorf("%s: %q names no state address to forget", OutOfBandDeleteAnnotation, value)
		}
		return &outOfBandRecovery{forget: addrs}, nil
	default:
		return nil, errors.Errorf("%s: %q is neither %q nor %q followed by state addresses",
			OutOfBandDeleteAnnotation, value, outOfBandRecreates, outOfBandForgetPrefix)
	}
}

// requireOutOfBandEngine refuses the act on an engine it cannot drive.
func requireOutOfBandEngine(engine string) error {
	if engine != "terraform" {
		return errors.Errorf("%s runs on OpenTofu and Terraform lanes only (it reads the plan's exit code and edits state); this lane's engine is %s",
			OutOfBandDeleteAnnotation, engine)
	}
	return nil
}

// runOutOfBandDelete hands the deletion to the harness, which must own the
// capability: a scenario that declares the act on a harness without it fails
// loudly instead of passing without deleting anything.
func runOutOfBandDelete(ctx context.Context, tc *provider.ComponentTestContext, harness provider.Harness) error {
	deleter, ok := harness.(provider.OutOfBandDeleter)
	if !ok {
		return errors.Errorf("scenario declares %s but the %s harness does not implement provider.OutOfBandDeleter",
			OutOfBandDeleteAnnotation, tc.Provider)
	}
	return errors.Wrap(deleter.DeleteOutOfBand(ctx, tc), "deleting the object outside the engine")
}

// runDriftPlan plans against the state that still records the deleted object
// and judges the outcome against the declared recovery.
func runDriftPlan(tc *provider.ComponentTestContext, recovery *outOfBandRecovery) error {
	opts, ok := tc.TerraformOpts.(*tt.Options)
	if !ok || opts == nil {
		return errors.New("terraform options not initialized (runValidate must run first)")
	}
	exitCode, err := tt.PlanExitCodeE(tc.T, opts)
	return judgeDriftPlan(recovery, exitCode, err)
}

// judgeDriftPlan compares a plan's detailed exit code (0 no change, 1 error,
// 2 changes) with what the GUIDE promised. A mismatch names the sentence that
// is wrong, so the fix lands in the GUIDE (and this annotation), never in a
// tolerance.
func judgeDriftPlan(recovery *outOfBandRecovery, exitCode int, runErr error) error {
	if runErr != nil {
		return errors.Wrap(runErr, "running the plan after the out-of-band delete")
	}
	if len(recovery.forget) > 0 {
		switch exitCode {
		case 1:
			fmt.Printf("  [out-of-band] the plan fails on the missing object, as the GUIDE says; the recovery forgets %s\n",
				strings.Join(recovery.forget, ", "))
			return nil
		default:
			return errors.Errorf("the scenario declares %q (the plan fails until the object is forgotten), but the plan exited %d: "+
				"the provider handles the missing object, so the GUIDE's recovery and this annotation should say %q",
				recovery, exitCode, outOfBandRecreates)
		}
	}
	switch exitCode {
	case 2:
		fmt.Println("  [out-of-band] the plan proposes creating the missing object again, as the GUIDE says")
		return nil
	case 1:
		return errors.Errorf("the scenario declares %q, but the plan failed on the missing object: "+
			"the GUIDE must teach forgetting it first (%s<address>)", outOfBandRecreates, outOfBandForgetPrefix)
	default:
		return errors.Errorf("the scenario declares %q, but the plan exited %d and proposes nothing: "+
			"the engine did not notice the object is gone", outOfBandRecreates, exitCode)
	}
}

// runRecover performs the GUIDE's recovery: forget the declared addresses,
// then apply the unchanged manifest.
func runRecover(tc *provider.ComponentTestContext, recovery *outOfBandRecovery) error {
	opts, ok := tc.TerraformOpts.(*tt.Options)
	if !ok || opts == nil {
		return errors.New("terraform options not initialized (runValidate must run first)")
	}
	if len(recovery.forget) > 0 {
		args := append([]string{"state", "rm", "-lock=false"}, recovery.forget...)
		if _, err := tt.RunTerraformCommandE(tc.T, opts, args...); err != nil {
			return errors.Wrapf(err, "forgetting %s", strings.Join(recovery.forget, ", "))
		}
	}
	return errors.Wrap(runDeploy(tc), "the recovery apply failed")
}
