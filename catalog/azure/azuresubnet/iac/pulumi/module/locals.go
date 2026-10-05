package module

import (
	"strings"

	"github.com/pkg/errors"
	azuresubnetv1alpha1 "github.com/plantonhq/planton/catalog/azure/azuresubnet/v1alpha1"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

type Locals struct {
	AzureSubnet *azuresubnetv1alpha1.AzureSubnet

	// VirtualNetworkId is a StringValueOrRef field; the platform middleware
	// resolves valueFrom references before IaC modules run, so GetValue()
	// always returns the resolved literal ARM ID.
	VirtualNetworkId string

	// RouteTableId, NetworkSecurityGroupId, and NatGatewayId are the
	// resolved ARM IDs of the subnet's optional attachments, or empty when
	// the spec leaves the seam unattached.
	RouteTableId           string
	NetworkSecurityGroupId string
	NatGatewayId           string

	// PrivateEndpointNetworkPolicies is the ARM string for the spec's enum,
	// or empty when unspecified so both engines let Azure apply its default
	// (Disabled).
	PrivateEndpointNetworkPolicies string

	// SharingScope is the ARM string for the spec's enum ("Tenant"), or
	// empty when the subnet is not shared.
	SharingScope string
}

func initializeLocals(ctx *pulumi.Context, iacInput *azuresubnetv1alpha1.AzureSubnetIacInput) *Locals {
	locals := &Locals{}

	locals.AzureSubnet = iacInput.Target
	target := iacInput.Target

	locals.VirtualNetworkId = target.Spec.VirtualNetworkId.GetValue()
	locals.RouteTableId = target.Spec.RouteTableId.GetValue()
	locals.NetworkSecurityGroupId = target.Spec.NetworkSecurityGroupId.GetValue()
	locals.NatGatewayId = target.Spec.NatGatewayId.GetValue()

	switch target.Spec.PrivateEndpointNetworkPolicies {
	case azuresubnetv1alpha1.AzureSubnetPrivateEndpointNetworkPolicies_ENABLED:
		locals.PrivateEndpointNetworkPolicies = "Enabled"
	case azuresubnetv1alpha1.AzureSubnetPrivateEndpointNetworkPolicies_NETWORK_SECURITY_GROUP_ENABLED:
		locals.PrivateEndpointNetworkPolicies = "NetworkSecurityGroupEnabled"
	case azuresubnetv1alpha1.AzureSubnetPrivateEndpointNetworkPolicies_ROUTE_TABLE_ENABLED:
		locals.PrivateEndpointNetworkPolicies = "RouteTableEnabled"
	}

	if target.Spec.SharingScope == azuresubnetv1alpha1.AzureSubnetSharingScope_TENANT {
		locals.SharingScope = "Tenant"
	}

	return locals
}

// parseVirtualNetworkId extracts the resource-group name and network name
// embedded in a virtual network's ARM ID
// (/subscriptions/{sub}/resourceGroups/{rg}/providers/Microsoft.Network/virtualNetworks/{name}).
// The subnet is an ARM child of the network, but the provider SDK takes the
// network name and resource group as separate arguments -- the module parses
// rather than asking the user to restate derivable state that could then
// disagree with the referenced network.
func parseVirtualNetworkId(virtualNetworkId string) (resourceGroupName string, virtualNetworkName string, err error) {
	segments := strings.Split(strings.Trim(virtualNetworkId, "/"), "/")
	for i := 0; i < len(segments)-1; i++ {
		if strings.EqualFold(segments[i], "resourceGroups") {
			resourceGroupName = segments[i+1]
		}
		if strings.EqualFold(segments[i], "virtualNetworks") {
			virtualNetworkName = segments[i+1]
		}
	}
	if resourceGroupName == "" || virtualNetworkName == "" {
		return "", "", errors.Errorf(
			"virtual_network_id %q is not a full virtual-network ARM ID "+
				"(expected /subscriptions/{sub}/resourceGroups/{rg}/providers/Microsoft.Network/virtualNetworks/{name})",
			virtualNetworkId)
	}
	return resourceGroupName, virtualNetworkName, nil
}
