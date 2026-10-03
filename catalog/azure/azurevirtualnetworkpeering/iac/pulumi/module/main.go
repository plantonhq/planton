package module

import (
	"github.com/pkg/errors"
	azurevirtualnetworkpeeringv1alpha1 "github.com/plantonhq/planton/catalog/azure/azurevirtualnetworkpeering/v1alpha1"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/provider/azure/pulumiazureprovider"
	"github.com/pulumi/pulumi-azure/sdk/v6/go/azure/network"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

func Resources(ctx *pulumi.Context, iacInput *azurevirtualnetworkpeeringv1alpha1.AzureVirtualNetworkPeeringIacInput) error {
	locals := initializeLocals(ctx, iacInput)

	// Build the Azure provider from the IaC input via the shared builder, which resolves
	// the right credential mechanism (static client secret, keyless web identity, or ambient chain).
	azureProvider, err := pulumiazureprovider.Get(ctx, iacInput.ProviderConfig)
	if err != nil {
		return errors.Wrap(err, "failed to create azure provider")
	}

	spec := locals.AzureVirtualNetworkPeering.Spec

	// The peering is an ARM child of its LOCAL network: the network's ARM ID
	// carries the resource group and network name, and the module derives
	// both rather than modeling redundant fields that could contradict the
	// referenced network.
	resourceGroupName, virtualNetworkName, err := parseVirtualNetworkId(locals.VirtualNetworkId)
	if err != nil {
		return err
	}

	// Lifecycle notes worth knowing before operating this resource:
	// - This resource is ONE DIRECTION of a peering; traffic only flows once
	//   the reciprocal peering exists on the remote network. Azure retries
	//   internally while the far side catches up, so the two directions can
	//   deploy concurrently.
	// - The access/forwarding/gateway flags and the subnet-name lists update
	//   IN PLACE. Name, the two networks, and the complete-vs-subnet-scoped
	//   and IPv6-only choices are the peering's identity -- changing any of
	//   them replaces it (a brief connectivity gap for this direction only).
	// - Peerings are not tracked ARM resources, so they carry no tags.
	peeringArgs := &network.VirtualNetworkPeeringArgs{
		Name:                   pulumi.String(spec.Name),
		ResourceGroupName:      pulumi.String(resourceGroupName),
		VirtualNetworkName:     pulumi.String(virtualNetworkName),
		RemoteVirtualNetworkId: pulumi.String(locals.RemoteVirtualNetworkId),

		// The four connectivity flags: proto-level defaults mirror Azure's
		// (access on; forwarding, gateway transit, and remote gateways off),
		// so both engines always send the same effective values. The
		// true-default flags are presence-guarded because the getters
		// return false when the optional field is unset -- the fallbacks
		// mirror the Terraform module's optional(bool, ...) defaults.
		AllowForwardedTraffic: pulumi.Bool(spec.GetAllowForwardedTraffic()),
		AllowGatewayTransit:   pulumi.Bool(spec.GetAllowGatewayTransit()),
		UseRemoteGateways:     pulumi.Bool(spec.GetUseRemoteGateways()),
	}

	if spec.AllowVirtualNetworkAccess != nil {
		peeringArgs.AllowVirtualNetworkAccess = pulumi.Bool(spec.GetAllowVirtualNetworkAccess())
	} else {
		peeringArgs.AllowVirtualNetworkAccess = pulumi.Bool(true)
	}
	if spec.PeerCompleteVirtualNetworksEnabled != nil {
		peeringArgs.PeerCompleteVirtualNetworksEnabled = pulumi.Bool(spec.GetPeerCompleteVirtualNetworksEnabled())
	} else {
		peeringArgs.PeerCompleteVirtualNetworksEnabled = pulumi.Bool(true)
	}

	// Subnet-scoped peering: the subnet-name lists are only meaningful when
	// complete-network peering is off (spec-level validation enforces the
	// pairing); an empty list is simply not sent.
	if len(spec.LocalSubnetNames) > 0 {
		peeringArgs.LocalSubnetNames = pulumi.ToStringArray(spec.LocalSubnetNames)
	}
	if len(spec.RemoteSubnetNames) > 0 {
		peeringArgs.RemoteSubnetNames = pulumi.ToStringArray(spec.RemoteSubnetNames)
	}

	// Only sent when explicitly chosen: ARM treats the field as a
	// creation-time property of the peering.
	if spec.GetOnlyIpv6PeeringEnabled() {
		peeringArgs.OnlyIpv6PeeringEnabled = pulumi.Bool(true)
	}

	createdPeering, err := network.NewVirtualNetworkPeering(ctx,
		spec.Name,
		peeringArgs,
		pulumi.Provider(azureProvider))
	if err != nil {
		return errors.Wrapf(err, "failed to create virtual network peering %s", spec.Name)
	}

	// Export outputs from the created resource.
	ctx.Export(OpPeeringId, createdPeering.ID())
	ctx.Export(OpPeeringName, createdPeering.Name)
	ctx.Export(OpVirtualNetworkName, pulumi.String(virtualNetworkName))
	ctx.Export(OpResourceGroupName, pulumi.String(resourceGroupName))

	return nil
}
