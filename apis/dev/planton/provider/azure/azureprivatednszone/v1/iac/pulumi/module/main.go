package module

import (
	"fmt"

	"github.com/pkg/errors"
	azureprivatednszonev1 "github.com/plantonhq/planton/apis/dev/planton/provider/azure/azureprivatednszone/v1"
	"github.com/pulumi/pulumi-azure/sdk/v6/go/azure"
	"github.com/pulumi/pulumi-azure/sdk/v6/go/azure/privatedns"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

func Resources(ctx *pulumi.Context, stackInput *azureprivatednszonev1.AzurePrivateDnsZoneStackInput) error {
	locals := initializeLocals(ctx, stackInput)
	azureProviderConfig := stackInput.ProviderConfig

	// Create azure provider using the credentials from the input
	azureProvider, err := azure.NewProvider(ctx,
		"azure",
		&azure.ProviderArgs{
			ClientId:       pulumi.String(azureProviderConfig.ClientId),
			ClientSecret:   pulumi.String(azureProviderConfig.ClientSecret),
			SubscriptionId: pulumi.String(azureProviderConfig.SubscriptionId),
			TenantId:       pulumi.String(azureProviderConfig.TenantId),
		})
	if err != nil {
		return errors.Wrap(err, "failed to create azure provider")
	}

	spec := locals.AzurePrivateDnsZone.Spec

	// Create the Private DNS Zone.
	// Private DNS zones are global resources (no location/region parameter).
	zone, err := privatedns.NewZone(ctx,
		spec.Name,
		&privatedns.ZoneArgs{
			Name:              pulumi.String(spec.Name),
			ResourceGroupName: pulumi.String(locals.ResourceGroupName),
			Tags:              pulumi.ToStringMap(locals.AzureTags),
		},
		pulumi.Provider(azureProvider))
	if err != nil {
		return errors.Wrapf(err, "failed to create private DNS zone %s", spec.Name)
	}

	// Create the VNet link.
	// A private DNS zone without a VNet link is unreachable (DD03 bundling).
	// The link name is derived from the resource metadata name to ensure uniqueness.
	vnetLinkName := fmt.Sprintf("%s-vnet-link", locals.AzurePrivateDnsZone.Metadata.Name)
	vnetId := spec.VnetId.GetValue()

	_, err = privatedns.NewZoneVirtualNetworkLink(ctx,
		vnetLinkName,
		&privatedns.ZoneVirtualNetworkLinkArgs{
			Name:                pulumi.String(vnetLinkName),
			ResourceGroupName:   pulumi.String(locals.ResourceGroupName),
			PrivateDnsZoneName:  zone.Name,
			VirtualNetworkId:    pulumi.String(vnetId),
			RegistrationEnabled: pulumi.Bool(spec.GetRegistrationEnabled()),
			Tags:                pulumi.ToStringMap(locals.AzureTags),
		},
		pulumi.Provider(azureProvider),
		pulumi.DependsOn([]pulumi.Resource{zone}))
	if err != nil {
		return errors.Wrapf(err, "failed to create VNet link %s", vnetLinkName)
	}

	// Export stack outputs
	ctx.Export(OpZoneId, zone.ID())
	ctx.Export(OpZoneName, zone.Name)

	return nil
}
