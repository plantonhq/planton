package module

import (
	"github.com/pkg/errors"
	gcpnetworkendpointgroupv1alpha1 "github.com/plantonhq/planton/catalog/gcp/gcpnetworkendpointgroup/v1alpha1"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/provider/gcp/pulumigoogleprovider"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

func Resources(ctx *pulumi.Context, stackInput *gcpnetworkendpointgroupv1alpha1.GcpNetworkEndpointGroupStackInput) error {
	locals := initializeLocals(ctx, stackInput)

	gcpProvider, err := pulumigoogleprovider.Get(ctx, stackInput.ProviderConfig)
	if err != nil {
		return errors.Wrap(err, "failed to setup google provider")
	}

	// spec.zone selects the API collection: a zone name builds the zonal
	// group with its bulk endpoint set, empty builds the global internet
	// group with one endpoint resource per member -- the same switch the
	// Terraform module makes with its count guards.
	if locals.IsZonal {
		if err := zonalNetworkEndpointGroup(ctx, locals, gcpProvider); err != nil {
			return errors.Wrap(err, "failed to create zonal network endpoint group")
		}
		return nil
	}
	if err := globalNetworkEndpointGroup(ctx, locals, gcpProvider); err != nil {
		return errors.Wrap(err, "failed to create global network endpoint group")
	}
	return nil
}
