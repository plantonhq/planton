package module

import (
	"github.com/pkg/errors"
	gcpsubnetworkv1alpha1 "github.com/plantonhq/planton/catalog/gcp/gcpsubnetwork/v1alpha1"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/provider/gcp/pulumigoogleprovider"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// Resources is the Pulumi program invoked by the Planton engine.
//
// Flow:
//  1. Derive local helpers from the IaC input.
//  2. Spin up a GCP provider from the supplied credential.
//  3. Call subnetwork() to enable necessary APIs and create the subnet.
//  4. Bubble up any error so the controller can surface it to operators.
func Resources(ctx *pulumi.Context, iacInput *gcpsubnetworkv1alpha1.GcpSubnetworkIacInput) error {
	locals := initializeLocals(ctx, iacInput)

	// (1) Provider setup – identical helper used by other Planton modules.
	gcpProvider, err := pulumigoogleprovider.Get(ctx, iacInput.ProviderConfig, iacInput.Target.GetSpec().GetProjectId().GetValue())
	if err != nil {
		return errors.Wrap(err, "failed to set up google provider")
	}

	// (2) Subnetwork creation.
	if _, err := subnetwork(ctx, locals, gcpProvider); err != nil {
		return errors.Wrap(err, "failed to create gcp subnetwork")
	}

	return nil
}
