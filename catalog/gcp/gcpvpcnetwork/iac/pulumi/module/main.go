package module

import (
	"github.com/pkg/errors"
	gcpvpcnetworkv1alpha1 "github.com/plantonhq/planton/catalog/gcp/gcpvpcnetwork/v1alpha1"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/provider/gcp/pulumigoogleprovider"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// Resources is the Pulumi program entry point invoked by the Planton CLI.
// It wires provider credentials, initializes locals, calls the noun‑style vpc()
// function, and surfaces any errors to the CLI.
func Resources(ctx *pulumi.Context, iacInput *gcpvpcnetworkv1alpha1.GcpVpcNetworkIacInput) error {
	// prepare useful locals (labels, metadata, credentials, etc.)
	locals := initializeLocals(ctx, iacInput)

	// create a GCP provider from the supplied credential
	gcpProvider, err := pulumigoogleprovider.Get(ctx, iacInput.ProviderConfig, iacInput.Target.GetSpec().GetProjectId().GetValue())
	if err != nil {
		return errors.Wrap(err, "failed to setup google provider")
	}

	// create the VPC network
	if _, err := vpc(ctx, locals, gcpProvider); err != nil {
		return errors.Wrap(err, "failed to create vpc network")
	}

	return nil
}
