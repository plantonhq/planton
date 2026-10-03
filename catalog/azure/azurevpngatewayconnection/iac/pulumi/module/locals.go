package module

import (
	azurevpngatewayconnectionv1alpha1 "github.com/plantonhq/planton/catalog/azure/azurevpngatewayconnection/v1alpha1"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// The connection carries no tags and no resource group of its own: ARM
// addresses it as a child of the VPN gateway, and the provider's schema
// has no tags argument -- so these locals need none of the family's
// usual tag-merging fields.
type Locals struct {
	AzureVpnGatewayConnection *azurevpngatewayconnectionv1alpha1.AzureVpnGatewayConnection
}

func initializeLocals(ctx *pulumi.Context, iacInput *azurevpngatewayconnectionv1alpha1.AzureVpnGatewayConnectionIacInput) *Locals {
	locals := &Locals{}

	locals.AzureVpnGatewayConnection = iacInput.Target

	return locals
}

// protocolWireValue maps the link's protocol enum onto ARM's
// vocabulary, applying ARM's default (IKEv2) when unspecified --
// mirroring the Terraform module's handling.
func protocolWireValue(protocol azurevpngatewayconnectionv1alpha1.AzureVpnGatewayConnectionProtocol) string {
	if protocol == azurevpngatewayconnectionv1alpha1.AzureVpnGatewayConnectionProtocol_IKE_V1 {
		return "IKEv1"
	}
	return "IKEv2"
}

// connectionModeWireValue maps the link's connection-mode enum onto
// ARM's vocabulary, applying ARM's default (Default) when unspecified.
func connectionModeWireValue(mode azurevpngatewayconnectionv1alpha1.AzureVpnGatewayConnectionMode) string {
	switch mode {
	case azurevpngatewayconnectionv1alpha1.AzureVpnGatewayConnectionMode_INITIATOR_ONLY:
		return "InitiatorOnly"
	case azurevpngatewayconnectionv1alpha1.AzureVpnGatewayConnectionMode_RESPONDER_ONLY:
		return "ResponderOnly"
	default:
		return "Default"
	}
}

// optionalInt32 returns the pointed-to value, or the default when the
// optional field is unset -- mirroring the Terraform variable default.
func optionalInt32(value *int32, defaultValue int32) int32 {
	if value == nil {
		return defaultValue
	}
	return *value
}
