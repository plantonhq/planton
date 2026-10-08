package module

import (
	"github.com/pkg/errors"
	gcpgkenodepoolv1alpha1 "github.com/plantonhq/planton/catalog/gcp/gcpgkenodepool/v1alpha1"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/provider/gcp/pulumigoogleprovider"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// Resources is the Pulumi entry‑point invoked by the runtime.
func Resources(ctx *pulumi.Context, iacInput *gcpgkenodepoolv1alpha1.GcpGkeNodePoolIacInput) error {
	locals := initializeLocals(ctx, iacInput)

	// Set up the Google provider from the supplied GCP credential.
	gcpProvider, err := pulumigoogleprovider.Get(ctx, iacInput.ProviderConfig, iacInput.Target.GetSpec().GetProjectId().GetValue())
	if err != nil {
		return errors.Wrap(err, "failed to configure google provider")
	}

	// Create the node pool.
	if err := nodePool(ctx, locals, gcpProvider); err != nil {
		return errors.Wrap(err, "failed to create node pool")
	}

	return nil
}
