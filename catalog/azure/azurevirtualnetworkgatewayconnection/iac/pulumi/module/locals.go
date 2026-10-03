package module

import (
	"strings"

	azurevirtualnetworkgatewayconnectionv1alpha1 "github.com/plantonhq/planton/catalog/azure/azurevirtualnetworkgatewayconnection/v1alpha1"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/provider/azure/azuretagkeys"
	"github.com/plantonhq/planton/shared/catalogkind"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

type Locals struct {
	AzureVirtualNetworkGatewayConnection *azurevirtualnetworkgatewayconnectionv1alpha1.AzureVirtualNetworkGatewayConnection

	// ResourceGroupName is a StringValueOrRef field; the platform middleware
	// resolves valueFrom references before IaC modules run, so GetValue()
	// always returns the resolved literal name.
	ResourceGroupName string

	// AzureTags is the metadata-derived tag map with the spec's user tags
	// merged over it (user tags win on key collision), mirroring the
	// Terraform module's merge order.
	AzureTags map[string]string
}

func initializeLocals(ctx *pulumi.Context, iacInput *azurevirtualnetworkgatewayconnectionv1alpha1.AzureVirtualNetworkGatewayConnectionIacInput) *Locals {
	locals := &Locals{}

	locals.AzureVirtualNetworkGatewayConnection = iacInput.Target
	target := iacInput.Target

	locals.ResourceGroupName = target.Spec.ResourceGroup.GetValue()

	// Metadata-derived tags first, then the user's spec tags merged over
	// them: user tags deliberately win so an org's governance conventions
	// (cost center, owner) can override the derived values where they
	// collide.
	locals.AzureTags = map[string]string{
		azuretagkeys.Resource:     "true",
		azuretagkeys.ResourceName: target.Metadata.Name,
		azuretagkeys.ResourceKind: strings.ToLower(catalogkind.CatalogKind_AzureVirtualNetworkGatewayConnection.String()),
	}

	if target.Metadata.Id != "" {
		locals.AzureTags[azuretagkeys.ResourceId] = target.Metadata.Id
	}

	if target.Metadata.Org != "" {
		locals.AzureTags[azuretagkeys.Organization] = target.Metadata.Org
	}

	if target.Metadata.Env != "" {
		locals.AzureTags[azuretagkeys.Environment] = target.Metadata.Env
	}

	for k, v := range target.Spec.Tags {
		locals.AzureTags[k] = v
	}

	return locals
}

// typeWireValue maps the connection-type enum to the wire vocabulary. The
// spec requires a non-zero type, so the empty default is unreachable for
// valid manifests.
func typeWireValue(connectionType azurevirtualnetworkgatewayconnectionv1alpha1.AzureVirtualNetworkGatewayConnectionType) string {
	switch connectionType {
	case azurevirtualnetworkgatewayconnectionv1alpha1.AzureVirtualNetworkGatewayConnectionType_IPSEC:
		return "IPsec"
	case azurevirtualnetworkgatewayconnectionv1alpha1.AzureVirtualNetworkGatewayConnectionType_VNET_TO_VNET:
		return "Vnet2Vnet"
	case azurevirtualnetworkgatewayconnectionv1alpha1.AzureVirtualNetworkGatewayConnectionType_EXPRESS_ROUTE:
		return "ExpressRoute"
	default:
		return ""
	}
}

// protocolWireValue maps the IKE protocol enum to the wire vocabulary.
// Returns "" for unspecified so callers omit the field -- the provider
// treats it as Computed and Azure applies its default (IKEv2).
func protocolWireValue(protocol azurevirtualnetworkgatewayconnectionv1alpha1.AzureVirtualNetworkGatewayConnectionProtocol) string {
	switch protocol {
	case azurevirtualnetworkgatewayconnectionv1alpha1.AzureVirtualNetworkGatewayConnectionProtocol_IKE_V1:
		return "IKEv1"
	case azurevirtualnetworkgatewayconnectionv1alpha1.AzureVirtualNetworkGatewayConnectionProtocol_IKE_V2:
		return "IKEv2"
	default:
		return ""
	}
}

// modeWireValue maps the connection-mode enum to the wire vocabulary.
// Unspecified deploys Default -- either side initiates, sent explicitly
// so both engines produce an identical payload.
func modeWireValue(mode azurevirtualnetworkgatewayconnectionv1alpha1.AzureVirtualNetworkGatewayConnectionMode) string {
	switch mode {
	case azurevirtualnetworkgatewayconnectionv1alpha1.AzureVirtualNetworkGatewayConnectionMode_INITIATOR_ONLY:
		return "InitiatorOnly"
	case azurevirtualnetworkgatewayconnectionv1alpha1.AzureVirtualNetworkGatewayConnectionMode_RESPONDER_ONLY:
		return "ResponderOnly"
	default:
		return "Default"
	}
}
