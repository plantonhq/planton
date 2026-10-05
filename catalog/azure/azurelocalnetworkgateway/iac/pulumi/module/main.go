package module

import (
	"github.com/pkg/errors"
	azurelocalnetworkgatewayv1alpha1 "github.com/plantonhq/planton/catalog/azure/azurelocalnetworkgateway/v1alpha1"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/provider/azure/pulumiazureprovider"
	"github.com/pulumi/pulumi-azure/sdk/v6/go/azure/network"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

func Resources(ctx *pulumi.Context, iacInput *azurelocalnetworkgatewayv1alpha1.AzureLocalNetworkGatewayIacInput) error {
	locals := initializeLocals(ctx, iacInput)

	// Build the Azure provider from the IaC input via the shared builder, which resolves
	// the right credential mechanism (static client secret, keyless web identity, or ambient chain).
	azureProvider, err := pulumiazureprovider.Get(ctx, iacInput.ProviderConfig)
	if err != nil {
		return errors.Wrap(err, "failed to create azure provider")
	}

	spec := locals.AzureLocalNetworkGateway.Spec

	gatewayArgs := &network.LocalNetworkGatewayArgs{
		Name:              pulumi.String(spec.Name),
		Location:          pulumi.String(spec.Region),
		ResourceGroupName: pulumi.String(locals.ResourceGroupName),
		Tags:              pulumi.ToStringMap(locals.AzureTags),
	}

	// The on-premises endpoint: exactly one of address or FQDN
	// (spec-validated) -- ARM stores whichever form is supplied and, for
	// FQDNs, re-resolves periodically.
	if spec.GatewayAddress != "" {
		gatewayArgs.GatewayAddress = pulumi.String(spec.GatewayAddress)
	}
	if spec.GatewayFqdn != "" {
		gatewayArgs.GatewayFqdn = pulumi.String(spec.GatewayFqdn)
	}

	// Static routing: the prefixes Azure routes into the tunnel. Empty is
	// legal only alongside BGP (spec-validated) -- learned routes carry
	// the site instead.
	if len(spec.AddressSpaces) > 0 {
		gatewayArgs.AddressSpaces = pulumi.ToStringArray(spec.AddressSpaces)
	}

	// Dynamic routing: the on-premises BGP speaker. The peering address
	// lives INSIDE the tunnel (the device's tunnel interface), not the
	// device's public endpoint.
	if spec.BgpSettings != nil {
		gatewayArgs.BgpSettings = &network.LocalNetworkGatewayBgpSettingsArgs{
			Asn:               pulumi.Int(int(spec.BgpSettings.Asn)),
			BgpPeeringAddress: pulumi.String(spec.BgpSettings.BgpPeeringAddress),
			PeerWeight:        pulumi.Int(int(spec.BgpSettings.PeerWeight)),
		}
	}

	createdGateway, err := network.NewLocalNetworkGateway(ctx,
		spec.Name,
		gatewayArgs,
		pulumi.Provider(azureProvider))
	if err != nil {
		return errors.Wrapf(err, "failed to create local network gateway %s", spec.Name)
	}

	ctx.Export(OpLocalNetworkGatewayId, createdGateway.ID())
	ctx.Export(OpLocalNetworkGatewayName, createdGateway.Name)

	return nil
}
