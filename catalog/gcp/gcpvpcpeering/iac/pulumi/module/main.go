package module

import (
	"github.com/pkg/errors"
	gcpvpcpeeringv1alpha1 "github.com/plantonhq/planton/catalog/gcp/gcpvpcpeering/v1alpha1"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/provider/gcp/pulumigoogleprovider"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

func Resources(ctx *pulumi.Context, stackInput *gcpvpcpeeringv1alpha1.GcpVpcPeeringStackInput) error {
	locals := initializeLocals(ctx, stackInput)

	gcpProvider, err := pulumigoogleprovider.Get(ctx, stackInput.ProviderConfig)
	if err != nil {
		return errors.Wrap(err, "failed to setup google provider")
	}

	// The form is selected by presence: a peer network means this side
	// creates the peering; none means it manages the route exchange of a
	// peering that already exists under peering_name.
	if locals.IsCreateForm {
		if err := networkPeering(ctx, locals, gcpProvider); err != nil {
			return errors.Wrap(err, "failed to create network peering")
		}
		return nil
	}
	if err := peeringRoutesConfig(ctx, locals, gcpProvider); err != nil {
		return errors.Wrap(err, "failed to configure peering routes")
	}
	return nil
}
