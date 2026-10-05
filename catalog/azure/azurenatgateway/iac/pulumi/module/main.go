package module

import (
	"fmt"

	"github.com/pkg/errors"
	azurenatgatewayv1alpha1 "github.com/plantonhq/planton/catalog/azure/azurenatgateway/v1alpha1"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/provider/azure/pulumiazureprovider"
	"github.com/pulumi/pulumi-azure/sdk/v6/go/azure/network"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

func Resources(ctx *pulumi.Context, iacInput *azurenatgatewayv1alpha1.AzureNatGatewayIacInput) error {
	locals := initializeLocals(ctx, iacInput)

	// Build the Azure provider from the IaC input via the shared builder, which resolves
	// the right credential mechanism (static client secret, keyless web identity, or ambient chain).
	azureProvider, err := pulumiazureprovider.Get(ctx, iacInput.ProviderConfig)
	if err != nil {
		return errors.Wrap(err, "failed to create azure provider")
	}

	spec := locals.AzureNatGateway.Spec

	// Lifecycle notes worth knowing before operating this resource:
	// - Idle timeout, tags, and the IP/prefix associations update IN
	//   PLACE. Name, SKU, and zone are the gateway's identity -- changing
	//   any of them replaces it, briefly interrupting egress for every
	//   attached subnet.
	// - The gateway is just the gateway: its addresses are referenced
	//   first-class resources, and the subnets it serves attach themselves
	//   (AzureSubnet's nat_gateway_id). A gateway with no associated
	//   addresses deploys but cannot translate anything.
	natGatewayArgs := &network.NatGatewayArgs{
		Name:              pulumi.String(spec.Name),
		Location:          pulumi.String(spec.Region),
		ResourceGroupName: pulumi.String(locals.ResourceGroupName),
		Tags:              pulumi.ToStringMap(locals.AzureTags),
	}

	// Presence-guarded: the getter returns 0 when the optional field is
	// unset, which azurerm's 4-120 validation rejects. An absent spec value
	// falls back to Azure's default (4) -- the same value the Terraform
	// module's optional(number, 4) encodes.
	if spec.IdleTimeoutInMinutes != nil {
		natGatewayArgs.IdleTimeoutInMinutes = pulumi.Int(int(spec.GetIdleTimeoutInMinutes()))
	} else {
		natGatewayArgs.IdleTimeoutInMinutes = pulumi.Int(4)
	}

	// Only an explicit SKU choice is sent, so an unspecified spec and
	// Azure's default (Standard) deploy identically on both engines.
	if locals.SkuName != "" {
		natGatewayArgs.SkuName = pulumi.String(locals.SkuName)
	}

	// A STANDARD gateway is zonal; STANDARD_V2 is zone-redundant and
	// forbids zones (spec-level validation enforces the pairing).
	if len(spec.Zones) > 0 {
		natGatewayArgs.Zones = pulumi.ToStringArray(spec.Zones)
	}

	createdNatGateway, err := network.NewNatGateway(ctx,
		spec.Name,
		natGatewayArgs,
		pulumi.Provider(azureProvider))
	if err != nil {
		return errors.Wrapf(err, "failed to create nat gateway %s", spec.Name)
	}

	// Associate the referenced addresses and ranges. Each association is
	// its own ARM operation; creating them here (rather than inside the
	// public IP modules) keeps the addresses reusable and makes the gateway
	// the owner of which addresses it SNATs through.
	for i, publicIpId := range locals.PublicIpIds {
		if _, err := network.NewNatGatewayPublicIpAssociation(ctx,
			fmt.Sprintf("%s-public-ip-%d", spec.Name, i),
			&network.NatGatewayPublicIpAssociationArgs{
				NatGatewayId:      createdNatGateway.ID(),
				PublicIpAddressId: pulumi.String(publicIpId),
			},
			pulumi.Provider(azureProvider)); err != nil {
			return errors.Wrapf(err, "failed to associate public ip %d with nat gateway %s", i, spec.Name)
		}
	}
	for i, prefixId := range locals.PublicIpPrefixIds {
		if _, err := network.NewNatGatewayPublicIpPrefixAssociation(ctx,
			fmt.Sprintf("%s-public-ip-prefix-%d", spec.Name, i),
			&network.NatGatewayPublicIpPrefixAssociationArgs{
				NatGatewayId:     createdNatGateway.ID(),
				PublicIpPrefixId: pulumi.String(prefixId),
			},
			pulumi.Provider(azureProvider)); err != nil {
			return errors.Wrapf(err, "failed to associate public ip prefix %d with nat gateway %s", i, spec.Name)
		}
	}

	// Export outputs from the created resource.
	ctx.Export(OpNatGatewayId, createdNatGateway.ID())
	ctx.Export(OpNatGatewayName, createdNatGateway.Name)
	ctx.Export(OpResourceGuid, createdNatGateway.ResourceGuid)

	return nil
}
