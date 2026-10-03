package module

import (
	"strings"

	"github.com/pkg/errors"
	azurevirtualnetworkpeeringv1alpha1 "github.com/plantonhq/planton/catalog/azure/azurevirtualnetworkpeering/v1alpha1"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

type Locals struct {
	AzureVirtualNetworkPeering *azurevirtualnetworkpeeringv1alpha1.AzureVirtualNetworkPeering

	// VirtualNetworkId and RemoteVirtualNetworkId are StringValueOrRef
	// fields; the platform middleware resolves valueFrom references before
	// IaC modules run, so GetValue() always returns the resolved literal
	// ARM ID.
	VirtualNetworkId       string
	RemoteVirtualNetworkId string
}

func initializeLocals(ctx *pulumi.Context, iacInput *azurevirtualnetworkpeeringv1alpha1.AzureVirtualNetworkPeeringIacInput) *Locals {
	locals := &Locals{}

	locals.AzureVirtualNetworkPeering = iacInput.Target
	target := iacInput.Target

	locals.VirtualNetworkId = target.Spec.VirtualNetworkId.GetValue()
	locals.RemoteVirtualNetworkId = target.Spec.RemoteVirtualNetworkId.GetValue()

	return locals
}

// parseVirtualNetworkId extracts the resource-group name and network name
// embedded in a virtual network's ARM ID
// (/subscriptions/{sub}/resourceGroups/{rg}/providers/Microsoft.Network/virtualNetworks/{name}).
// The peering is an ARM child of its local network, but the provider SDK
// takes the network name and resource group as separate arguments -- the
// module parses rather than asking the user to restate derivable state that
// could then disagree with the referenced network.
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
