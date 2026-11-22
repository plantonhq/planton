// The step a cloud requires of a resource's owners before its destroy (see
// provider.DestroyPreparer), taken before every destroy of the kind under
// test.

package runner

import (
	"context"

	"github.com/pkg/errors"
	"github.com/plantonhq/planton/e2e/framework/provider"
)

// PrepareDestroyAnnotation names the class of owner step a scenario's harness
// takes before each destroy of the kind under test. The class is
// harness-interpreted.
const PrepareDestroyAnnotation = "planton.dev/e2e-prepare-destroy"

// destroyPreparation returns the step to take before each destroy attempt of
// the kind under test, or nil when the scenario declares none. ctx carries
// the manifest the step reads (ManifestPathKey). A harness without the
// capability yields a step that fails, so the destroy it guards fails
// loudly instead of the annotation being ignored.
func destroyPreparation(ctx context.Context, tc *provider.KindTestContext, harness provider.Harness, manifestPath string) func() error {
	class, err := ManifestAnnotation(manifestPath, PrepareDestroyAnnotation)
	if err != nil || class == "" {
		return nil
	}
	preparer, ok := harness.(provider.DestroyPreparer)
	if !ok {
		return func() error {
			return errors.Errorf("scenario declares %s: %q but the %s harness does not implement provider.DestroyPreparer",
				PrepareDestroyAnnotation, class, tc.Provider)
		}
	}
	return func() error {
		lanef(tc.T, "  [prepare-destroy] taking the owners' step before destroy (%s)", class)
		return errors.Wrapf(preparer.PrepareDestroy(ctx, tc, class), "preparing the destroy (%s)", class)
	}
}
