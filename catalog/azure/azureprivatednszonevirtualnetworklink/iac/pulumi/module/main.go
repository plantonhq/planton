package module

import (
	"github.com/pkg/errors"
	azureprivatednszonevirtualnetworklinkv1alpha1 "github.com/plantonhq/planton/catalog/azure/azureprivatednszonevirtualnetworklink/v1alpha1"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/provider/azure/pulumiazureprovider"
	"github.com/pulumi/pulumi-azure/sdk/v6/go/azure/privatedns"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

func Resources(ctx *pulumi.Context, iacInput *azureprivatednszonevirtualnetworklinkv1alpha1.AzurePrivateDnsZoneVirtualNetworkLinkIacInput) error {
	locals := initializeLocals(ctx, iacInput)

	// Build the Azure provider from the IaC input via the shared builder, which resolves
	// the right credential mechanism (static client secret, keyless web identity, or ambient chain).
	azureProvider, err := pulumiazureprovider.Get(ctx, iacInput.ProviderConfig)
	if err != nil {
		return errors.Wrap(err, "failed to create azure provider")
	}

	spec := locals.AzurePrivateDnsZoneVirtualNetworkLink.Spec

	// The link is an ARM child of the zone: the zone's ARM ID carries the
	// zone name and resource group, and the module derives both rather than
	// modeling redundant fields that could contradict the referenced zone.
	zoneResourceGroupName, zoneName, err := parseZoneId(locals.PrivateDnsZoneId)
	if err != nil {
		return err
	}

	// Lifecycle notes worth knowing before operating this resource:
	// - registration_enabled, resolution_policy, and tags update IN PLACE;
	//   name, zone, and network are the link's ARM identity, so changing
	//   any of them replaces the link (a brief resolution gap for the
	//   affected network, nothing else).
	// - Azure allows only ONE registration-enabled link per virtual
	//   network; additional links to the same network must keep VM
	//   auto-registration off.
	linkArgs := &privatedns.ZoneVirtualNetworkLinkArgs{
		Name:                pulumi.String(spec.Name),
		ResourceGroupName:   pulumi.String(zoneResourceGroupName),
		PrivateDnsZoneName:  pulumi.String(zoneName),
		VirtualNetworkId:    pulumi.String(locals.VirtualNetworkId),
		RegistrationEnabled: pulumi.Bool(locals.RegistrationEnabled),
		Tags:                pulumi.ToStringMap(locals.AzureTags),
	}

	// Omitted lets Azure choose its per-zone-type default (privatelink
	// zones get their platform-managed policy); only an explicit policy is
	// ever sent, so an unspecified spec and Azure's default deploy
	// identically on both engines.
	if locals.ResolutionPolicy != "" {
		linkArgs.ResolutionPolicy = pulumi.String(locals.ResolutionPolicy)
	}

	createdLink, err := privatedns.NewZoneVirtualNetworkLink(ctx,
		spec.Name,
		linkArgs,
		pulumi.Provider(azureProvider))
	if err != nil {
		return errors.Wrapf(err, "failed to create private dns zone virtual network link %s", spec.Name)
	}

	// Export outputs from the created resource.
	ctx.Export(OpLinkId, createdLink.ID())
	ctx.Export(OpLinkName, createdLink.Name)

	return nil
}
