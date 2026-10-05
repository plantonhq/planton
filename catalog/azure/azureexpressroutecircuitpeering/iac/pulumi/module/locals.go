package module

import (
	azureexpressroutecircuitpeeringv1alpha1 "github.com/plantonhq/planton/catalog/azure/azureexpressroutecircuitpeering/v1alpha1"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

type Locals struct {
	AzureExpressRouteCircuitPeering *azureexpressroutecircuitpeeringv1alpha1.AzureExpressRouteCircuitPeering

	// ResourceGroupName and ExpressRouteCircuitName are StringValueOrRef
	// fields; the platform middleware resolves valueFrom references before
	// IaC modules run, so GetValue() always returns the resolved literal.
	ResourceGroupName       string
	ExpressRouteCircuitName string
}

// NOTE: no tag locals -- an ExpressRoute circuit peering is an ARM child
// of the circuit and carries no tags of its own (the provider schema has
// no tags argument; governance tags live on the parent circuit).

func initializeLocals(ctx *pulumi.Context, iacInput *azureexpressroutecircuitpeeringv1alpha1.AzureExpressRouteCircuitPeeringIacInput) *Locals {
	locals := &Locals{}

	locals.AzureExpressRouteCircuitPeering = iacInput.Target
	locals.ResourceGroupName = iacInput.Target.Spec.ResourceGroup.GetValue()
	locals.ExpressRouteCircuitName = iacInput.Target.Spec.ExpressRouteCircuitName.GetValue()

	return locals
}

// peeringTypeWireValue maps the spec's peering-type enum onto ARM's
// vocabulary -- the value is also the ARM child's NAME on the circuit.
func peeringTypeWireValue(peeringType azureexpressroutecircuitpeeringv1alpha1.AzureExpressRouteCircuitPeeringType) string {
	switch peeringType {
	case azureexpressroutecircuitpeeringv1alpha1.AzureExpressRouteCircuitPeeringType_AZURE_PRIVATE_PEERING:
		return "AzurePrivatePeering"
	case azureexpressroutecircuitpeeringv1alpha1.AzureExpressRouteCircuitPeeringType_AZURE_PUBLIC_PEERING:
		return "AzurePublicPeering"
	case azureexpressroutecircuitpeeringv1alpha1.AzureExpressRouteCircuitPeeringType_MICROSOFT_PEERING:
		return "MicrosoftPeering"
	}
	// Unreachable: the spec's peering_type_required contract rejects
	// unspecified before the module runs.
	return ""
}
