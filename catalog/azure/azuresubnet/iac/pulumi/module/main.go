package module

import (
	"github.com/pkg/errors"
	azuresubnetv1alpha1 "github.com/plantonhq/planton/catalog/azure/azuresubnet/v1alpha1"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/provider/azure/pulumiazureprovider"
	"github.com/pulumi/pulumi-azure/sdk/v6/go/azure/network"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

func Resources(ctx *pulumi.Context, iacInput *azuresubnetv1alpha1.AzureSubnetIacInput) error {
	locals := initializeLocals(ctx, iacInput)

	// Build the Azure provider from the IaC input via the shared builder, which resolves
	// the right credential mechanism (static client secret, keyless web identity, or ambient chain).
	azureProvider, err := pulumiazureprovider.Get(ctx, iacInput.ProviderConfig)
	if err != nil {
		return errors.Wrap(err, "failed to create azure provider")
	}

	spec := locals.AzureSubnet.Spec

	// The subnet is an ARM child of the virtual network: the network's ARM
	// ID carries the resource group and network name, and the module derives
	// both rather than modeling redundant fields that could contradict the
	// referenced network.
	resourceGroupName, virtualNetworkName, err := parseVirtualNetworkId(locals.VirtualNetworkId)
	if err != nil {
		return err
	}

	// Lifecycle notes worth knowing before operating this resource:
	// - Address prefixes, service endpoints, policies, delegations, and the
	//   route-table/NSG/NAT attachments all update IN PLACE; name and the
	//   parent network are the subnet's ARM identity, so changing either
	//   replaces the subnet and everything deployed into it.
	// - A prefix in use by deployed resources cannot shrink, and toggling
	//   default_outbound_access_enabled requires the subnet to be empty of
	//   VMs -- plan address space and egress posture before workloads land.
	subnetArgs := &network.SubnetArgs{
		Name:               pulumi.String(spec.Name),
		ResourceGroupName:  pulumi.String(resourceGroupName),
		VirtualNetworkName: pulumi.String(virtualNetworkName),
	}

	// Exactly one address source is set (spec-level validation enforces the
	// XOR): self-managed CIDR blocks, or delegated allocation from a Network
	// Manager IPAM pool that provisions the actual range at deploy time.
	if len(spec.AddressPrefixes) > 0 {
		subnetArgs.AddressPrefixes = pulumi.ToStringArray(spec.AddressPrefixes)
	}
	if spec.IpAddressPool != nil {
		subnetArgs.IpAddressPool = &network.SubnetIpAddressPoolArgs{
			Id:                  pulumi.String(spec.IpAddressPool.Id),
			NumberOfIpAddresses: pulumi.String(spec.IpAddressPool.NumberOfIpAddresses),
		}
	}

	if len(spec.ServiceEndpoints) > 0 {
		subnetArgs.ServiceEndpoints = pulumi.ToStringArray(spec.ServiceEndpoints)
	}
	if len(spec.ServiceEndpointPolicyIds) > 0 {
		subnetArgs.ServiceEndpointPolicyIds = pulumi.ToStringArray(spec.ServiceEndpointPolicyIds)
	}

	// Delegations hand the subnet to a PaaS service. An explicit action list
	// is only sent when the spec carries one; otherwise Azure applies the
	// service's default action set.
	if len(spec.Delegations) > 0 {
		delegations := network.SubnetDelegationArray{}
		for _, delegation := range spec.Delegations {
			serviceDelegation := network.SubnetDelegationServiceDelegationArgs{
				Name: pulumi.String(delegation.ServiceName),
			}
			if len(delegation.Actions) > 0 {
				serviceDelegation.Actions = pulumi.ToStringArray(delegation.Actions)
			}
			delegations = append(delegations, network.SubnetDelegationArgs{
				Name:              pulumi.String(delegation.Name),
				ServiceDelegation: serviceDelegation,
			})
		}
		subnetArgs.Delegations = delegations
	}

	// Only an explicit policy mode is ever sent, so an unspecified spec and
	// Azure's default (Disabled) deploy identically on both engines.
	if locals.PrivateEndpointNetworkPolicies != "" {
		subnetArgs.PrivateEndpointNetworkPolicies = pulumi.String(locals.PrivateEndpointNetworkPolicies)
	}

	// Presence-guarded: both optional bools default to true in the proto,
	// but the getters return false when the field is unset. An absent spec
	// value falls back to true -- the same value the Terraform module's
	// optional(bool, true) encodes -- so an untouched manifest never
	// silently disables private-link policies or subnet egress.
	if spec.PrivateLinkServiceNetworkPoliciesEnabled != nil {
		subnetArgs.PrivateLinkServiceNetworkPoliciesEnabled = pulumi.Bool(spec.GetPrivateLinkServiceNetworkPoliciesEnabled())
	} else {
		subnetArgs.PrivateLinkServiceNetworkPoliciesEnabled = pulumi.Bool(true)
	}
	if spec.DefaultOutboundAccessEnabled != nil {
		subnetArgs.DefaultOutboundAccessEnabled = pulumi.Bool(spec.GetDefaultOutboundAccessEnabled())
	} else {
		subnetArgs.DefaultOutboundAccessEnabled = pulumi.Bool(true)
	}

	// ARM only accepts sharing_scope alongside disabled default outbound
	// access (spec-level validation enforces the pairing).
	if locals.SharingScope != "" {
		subnetArgs.SharingScope = pulumi.String(locals.SharingScope)
	}

	createdSubnet, err := network.NewSubnet(ctx,
		spec.Name,
		subnetArgs,
		pulumi.Provider(azureProvider))
	if err != nil {
		return errors.Wrapf(err, "failed to create subnet %s", spec.Name)
	}

	// The attach seams. Azure models route-table/NSG/NAT attachment as
	// writes to the subnet, declared subnet-side because one table, group,
	// or gateway serves many subnets. Each association is its own ARM
	// operation; creating them here (rather than inside the referenced
	// resources' modules) keeps those resources reusable across subnets.
	if locals.RouteTableId != "" {
		if _, err := network.NewSubnetRouteTableAssociation(ctx,
			spec.Name+"-route-table",
			&network.SubnetRouteTableAssociationArgs{
				SubnetId:     createdSubnet.ID(),
				RouteTableId: pulumi.String(locals.RouteTableId),
			},
			pulumi.Provider(azureProvider)); err != nil {
			return errors.Wrapf(err, "failed to associate route table with subnet %s", spec.Name)
		}
	}

	if locals.NetworkSecurityGroupId != "" {
		if _, err := network.NewSubnetNetworkSecurityGroupAssociation(ctx,
			spec.Name+"-network-security-group",
			&network.SubnetNetworkSecurityGroupAssociationArgs{
				SubnetId:               createdSubnet.ID(),
				NetworkSecurityGroupId: pulumi.String(locals.NetworkSecurityGroupId),
			},
			pulumi.Provider(azureProvider)); err != nil {
			return errors.Wrapf(err, "failed to associate network security group with subnet %s", spec.Name)
		}
	}

	if locals.NatGatewayId != "" {
		if _, err := network.NewSubnetNatGatewayAssociation(ctx,
			spec.Name+"-nat-gateway",
			&network.SubnetNatGatewayAssociationArgs{
				SubnetId:     createdSubnet.ID(),
				NatGatewayId: pulumi.String(locals.NatGatewayId),
			},
			pulumi.Provider(azureProvider)); err != nil {
			return errors.Wrapf(err, "failed to associate nat gateway with subnet %s", spec.Name)
		}
	}

	// Export outputs from the created resource. address_prefixes comes
	// from the resource (not the spec) so IPAM-allocated subnets surface the
	// ranges the pool actually provisioned.
	ctx.Export(OpSubnetId, createdSubnet.ID())
	ctx.Export(OpSubnetName, createdSubnet.Name)
	ctx.Export(OpAddressPrefixes, createdSubnet.AddressPrefixes)
	ctx.Export(OpVirtualNetworkName, pulumi.String(virtualNetworkName))
	ctx.Export(OpResourceGroupName, pulumi.String(resourceGroupName))

	return nil
}
