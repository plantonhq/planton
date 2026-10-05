package module

import (
	"github.com/pkg/errors"
	azurevirtualnetworkv1alpha1 "github.com/plantonhq/planton/catalog/azure/azurevirtualnetwork/v1alpha1"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/provider/azure/pulumiazureprovider"
	"github.com/pulumi/pulumi-azure/sdk/v6/go/azure/network"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

func Resources(ctx *pulumi.Context, iacInput *azurevirtualnetworkv1alpha1.AzureVirtualNetworkIacInput) error {
	locals := initializeLocals(ctx, iacInput)

	// Build the Azure provider from the IaC input via the shared builder, which resolves
	// the right credential mechanism (static client secret, keyless web identity, or ambient chain).
	azureProvider, err := pulumiazureprovider.Get(ctx, iacInput.ProviderConfig)
	if err != nil {
		return errors.Wrap(err, "failed to create azure provider")
	}

	spec := locals.AzureVirtualNetwork.Spec

	// Lifecycle notes worth knowing before operating this resource:
	// - Address space, DNS servers, BGP community, DDoS attachment,
	//   encryption, flow timeout, and tags all update IN PLACE. Name,
	//   region, resource group, and edge zone are the network's ARM
	//   identity -- changing any of them replaces the network and
	//   everything inside it.
	// - Address-space blocks can be added or removed live, but a block
	//   that subnets are carved from cannot shrink below them.
	// - The network is deliberately just the network: subnets live in
	//   AzureSubnet, outbound NAT in AzureNatGateway, and private DNS
	//   attachments in AzurePrivateDnsZoneVirtualNetworkLink, each
	//   referencing this network's outputs.
	vnetArgs := &network.VirtualNetworkArgs{
		Name:              pulumi.String(spec.Name),
		Location:          pulumi.String(spec.Region),
		ResourceGroupName: pulumi.String(locals.ResourceGroupName),
		Tags:              pulumi.ToStringMap(locals.AzureTags),
	}

	// Exactly one address source is set (spec-level validation): either
	// self-managed CIDR blocks or delegated allocation from Azure Network
	// Manager IPAM pools. With IPAM pools the actual CIDR ranges are
	// provisioned at deploy time and surfaced through the address_spaces
	// output.
	if len(spec.AddressSpaces) > 0 {
		vnetArgs.AddressSpaces = pulumi.ToStringArray(spec.AddressSpaces)
	}
	if len(spec.IpAddressPools) > 0 {
		pools := make(network.VirtualNetworkIpAddressPoolArray, 0, len(spec.IpAddressPools))
		for _, pool := range spec.IpAddressPools {
			pools = append(pools, network.VirtualNetworkIpAddressPoolArgs{
				Id:                  pulumi.String(pool.Id),
				NumberOfIpAddresses: pulumi.String(pool.NumberOfIpAddresses),
			})
		}
		vnetArgs.IpAddressPools = pools
	}

	// Empty means Azure's default resolver (168.63.129.16) serves the
	// network -- required for private DNS zone resolution to work directly,
	// so custom servers are only sent when explicitly configured.
	if len(spec.DnsServers) > 0 {
		vnetArgs.DnsServers = pulumi.ToStringArray(spec.DnsServers)
	}

	if spec.BgpCommunity != "" {
		vnetArgs.BgpCommunity = pulumi.String(spec.BgpCommunity)
	}

	// The DDoS plan is a separate, shared (and billed) resource; this block
	// only attaches an existing plan. ARM keeps attachment and activation
	// distinct so a plan can stay attached with protection toggled off.
	if spec.DdosProtectionPlan != nil {
		vnetArgs.DdosProtectionPlan = &network.VirtualNetworkDdosProtectionPlanArgs{
			Id:     pulumi.String(spec.DdosProtectionPlan.Id),
			Enable: pulumi.Bool(spec.DdosProtectionPlan.Enable),
		}
	}

	// Omitted means ARM's default (encryption off); only an explicit
	// enforcement mode sends the block, so an unspecified spec and Azure's
	// default deploy identically on both engines. Note ARM currently
	// accepts only AllowUnencrypted -- DropUnencrypted is modeled because
	// the API defines it but is not yet generally available.
	if locals.EncryptionEnforcement != "" {
		vnetArgs.Encryption = &network.VirtualNetworkEncryptionArgs{
			Enforcement: pulumi.String(locals.EncryptionEnforcement),
		}
	}

	// Omitted lets Azure apply its 4-minute default; the spec constrains
	// the value to ARM's accepted 4-30 range.
	if spec.FlowTimeoutInMinutes != nil {
		vnetArgs.FlowTimeoutInMinutes = pulumi.Int(int(*spec.FlowTimeoutInMinutes))
	}

	// Omitted means ARM's default ("Disabled"); only the opt-in "Basic"
	// mode is ever sent.
	if locals.PrivateEndpointVnetPolicies != "" {
		vnetArgs.PrivateEndpointVnetPolicies = pulumi.String(locals.PrivateEndpointVnetPolicies)
	}

	if spec.EdgeZone != "" {
		vnetArgs.EdgeZone = pulumi.String(spec.EdgeZone)
	}

	createdVirtualNetwork, err := network.NewVirtualNetwork(ctx,
		spec.Name,
		vnetArgs,
		pulumi.Provider(azureProvider))
	if err != nil {
		return errors.Wrapf(err, "failed to create virtual network %s", spec.Name)
	}

	// Export outputs from the created resource. virtual_network_id is
	// the join key subnets, peerings, and DNS links attach through;
	// address_spaces reflects the ACTUAL ranges (IPAM-provisioned when
	// ip_address_pools delegate allocation); guid identifies the network to
	// features that key on ARM's stable network GUID rather than the ID.
	ctx.Export(OpVirtualNetworkId, createdVirtualNetwork.ID())
	ctx.Export(OpVirtualNetworkName, createdVirtualNetwork.Name)
	ctx.Export(OpGuid, createdVirtualNetwork.Guid)
	ctx.Export(OpAddressSpaces, createdVirtualNetwork.AddressSpaces)

	return nil
}
