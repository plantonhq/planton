package module

import (
	"github.com/pkg/errors"
	azurefrontdoorendpointv1alpha1 "github.com/plantonhq/planton/catalog/azure/azurefrontdoorendpoint/v1alpha1"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/provider/azure/pulumiazureprovider"
	"github.com/pulumi/pulumi-azure/sdk/v6/go/azure/cdn"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

func Resources(ctx *pulumi.Context, iacInput *azurefrontdoorendpointv1alpha1.AzureFrontDoorEndpointIacInput) error {
	locals := initializeLocals(ctx, iacInput)

	// Build the Azure provider from the IaC input via the shared builder, which resolves
	// the right credential mechanism (static client secret, keyless web identity, or ambient chain).
	azureProvider, err := pulumiazureprovider.Get(ctx, iacInput.ProviderConfig)
	if err != nil {
		return errors.Wrap(err, "failed to create azure provider")
	}

	spec := locals.AzureFrontDoorEndpoint.Spec

	// The endpoint addresses its parent by the profile's full ARM id --
	// the provider derives the resource group and profile name from it,
	// so the spec carries one authoritative parent reference.
	endpointArgs := &cdn.FrontdoorEndpointArgs{
		Name:                  pulumi.String(spec.EndpointName),
		CdnFrontdoorProfileId: pulumi.String(locals.ProfileId),
		Tags:                  pulumi.ToStringMap(locals.AzureTags),
	}

	// Sent only when explicitly disabled: Azure's default is enabled, and
	// the platform materializes the documented default centrally (IaC
	// inputs never carry proto defaults, so an absent field means true).
	if spec.Enabled != nil {
		endpointArgs.Enabled = pulumi.Bool(spec.GetEnabled())
	}

	createdEndpoint, err := cdn.NewFrontdoorEndpoint(ctx,
		spec.EndpointName,
		endpointArgs,
		pulumi.Provider(azureProvider))
	if err != nil {
		return errors.Wrapf(err, "failed to create front door endpoint %s", spec.EndpointName)
	}

	// Export outputs. host_name is the generated *.azurefd.net
	// hostname -- the CNAME target custom-domain DNS records point at.
	ctx.Export(OpEndpointId, createdEndpoint.ID())
	ctx.Export(OpEndpointName, createdEndpoint.Name)
	ctx.Export(OpHostName, createdEndpoint.HostName)

	return nil
}
