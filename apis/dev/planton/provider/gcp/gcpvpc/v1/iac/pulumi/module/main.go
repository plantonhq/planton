package module

import (
	"github.com/pkg/errors"
	gcpvpcv1 "github.com/plantonhq/planton/apis/dev/planton/provider/gcp/gcpvpc/v1"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/provider/gcp/pulumigoogleprovider"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// Resources is the Pulumi program entry point invoked by the Planton CLI.
// It wires provider credentials, initializes locals, calls the noun‑style vpc()
// function, and surfaces any errors to the CLI.
func Resources(ctx *pulumi.Context, stackInput *gcpvpcv1.GcpVpcStackInput) error {
	// prepare useful locals (labels, metadata, credentials, etc.)
	locals := initializeLocals(ctx, stackInput)

	// create a GCP provider from the supplied credential
	gcpProvider, err := pulumigoogleprovider.Get(ctx, stackInput.ProviderConfig)
	if err != nil {
		return errors.Wrap(err, "failed to setup google provider")
	}

	// create the VPC network
	if _, err := vpc(ctx, locals, gcpProvider); err != nil {
		return errors.Wrap(err, "failed to create vpc network")
	}

	return nil
}
