package module

import (
	"github.com/pkg/errors"
	gcprouternatv1 "github.com/plantonhq/planton/apis/dev/planton/provider/gcp/gcprouternat/v1"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/provider/gcp/pulumigoogleprovider"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// Resources is the Pulumi program entry‑point invoked by the Planton
func Resources(
	ctx *pulumi.Context,
	stackInput *gcprouternatv1.GcpRouterNatStackInput,
) error {
	// gather locals (Terraform‑style “locals”)
	locals := initializeLocals(ctx, stackInput)

	// configure a GCP provider from the given credential
	gcpProvider, err := pulumigoogleprovider.Get(ctx, stackInput.ProviderConfig)
	if err != nil {
		return errors.Wrap(err, "failed to setup google provider")
	}

	// build router+nat
	if _, err = routerNat(ctx, locals, gcpProvider); err != nil {
		return errors.Wrap(err, "failed to create router nat resources")
	}

	return nil
}
