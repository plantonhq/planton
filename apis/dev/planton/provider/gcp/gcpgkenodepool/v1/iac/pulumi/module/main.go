package module

import (
	"github.com/pkg/errors"
	gcpgkenodepoolv1 "github.com/plantonhq/planton/apis/dev/planton/provider/gcp/gcpgkenodepool/v1"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/provider/gcp/pulumigoogleprovider"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// Resources is the Pulumi entry‑point invoked by the runtime.
func Resources(ctx *pulumi.Context, stackInput *gcpgkenodepoolv1.GcpGkeNodePoolStackInput) error {
	locals := initializeLocals(ctx, stackInput)

	// Set up the Google provider from the supplied GCP credential.
	gcpProvider, err := pulumigoogleprovider.Get(ctx, stackInput.ProviderConfig)
	if err != nil {
		return errors.Wrap(err, "failed to configure google provider")
	}

	// Create the node pool.
	if err := nodePool(ctx, locals, gcpProvider); err != nil {
		return errors.Wrap(err, "failed to create node pool")
	}

	return nil
}
