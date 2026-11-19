// The release of what the provider made on the lane's behalf (see
// provider.DependentsReleaser): after the kind's last destroy, before the
// fixtures are torn down.

package runner

import (
	"context"
	"time"

	"github.com/pkg/errors"
	"github.com/plantonhq/planton/e2e/framework/provider"
)

// ReleaseDependentsAnnotation names the class of provider-created objects a
// scenario's harness releases after the kind's destroy. The class is
// harness-interpreted.
const ReleaseDependentsAnnotation = "planton.dev/e2e-release-dependents"

// PhaseReleaseDependents runs after the kind's last destroy and before
// DEPENDENCIES-DOWN, on every path that reaches the teardown.
const PhaseReleaseDependents Phase = "RELEASE-DEPENDENTS"

// releaseDependents runs the harness's release for the class the scenario
// names, and returns the phase to record (nil when the scenario names none).
// A failure fails the lane: what the provider made is still there, and the
// fixtures' teardown that follows may be refused because of it.
func releaseDependents(ctx context.Context, tc *provider.KindTestContext, harness provider.Harness, manifestPath string) *PhaseResult {
	class, err := ManifestAnnotation(manifestPath, ReleaseDependentsAnnotation)
	if err != nil || class == "" {
		return nil
	}
	start := time.Now()
	releaser, ok := harness.(provider.DependentsReleaser)
	if !ok {
		err = errors.Errorf("scenario declares %s: %q but the %s harness does not implement provider.DependentsReleaser",
			ReleaseDependentsAnnotation, class, tc.Provider)
	} else {
		lanef(tc.T, "  [release] releasing what the provider made for this lane (%s)", class)
		err = releaser.ReleaseDependents(ctx, tc, class)
	}
	return &PhaseResult{Phase: PhaseReleaseDependents, Duration: time.Since(start), Passed: err == nil, Error: err}
}
