package module

import (
	"strings"

	azurevirtualnetworkgatewayv1alpha1 "github.com/plantonhq/planton/catalog/azure/azurevirtualnetworkgateway/v1alpha1"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/provider/azure/azuretagkeys"
	"github.com/plantonhq/planton/shared/catalogkind"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

type Locals struct {
	AzureVirtualNetworkGateway *azurevirtualnetworkgatewayv1alpha1.AzureVirtualNetworkGateway

	// ResourceGroupName is a StringValueOrRef field; the platform middleware
	// resolves valueFrom references before IaC modules run, so GetValue()
	// always returns the resolved literal name.
	ResourceGroupName string

	// AzureTags is the metadata-derived tag map with the spec's user tags
	// merged over it (user tags win on key collision), mirroring the
	// Terraform module's merge order.
	AzureTags map[string]string
}

func initializeLocals(ctx *pulumi.Context, iacInput *azurevirtualnetworkgatewayv1alpha1.AzureVirtualNetworkGatewayIacInput) *Locals {
	locals := &Locals{}

	locals.AzureVirtualNetworkGateway = iacInput.Target
	target := iacInput.Target

	locals.ResourceGroupName = target.Spec.ResourceGroup.GetValue()

	// Metadata-derived tags first, then the user's spec tags merged over
	// them: user tags deliberately win so an org's governance conventions
	// (cost center, owner) can override the derived values where they
	// collide.
	locals.AzureTags = map[string]string{
		azuretagkeys.Resource:     "true",
		azuretagkeys.ResourceName: target.Metadata.Name,
		azuretagkeys.ResourceKind: strings.ToLower(catalogkind.CatalogKind_AzureVirtualNetworkGateway.String()),
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

// typeWireValue maps the gateway-type enum to the wire vocabulary.
// Unspecified deploys Vpn -- the site-to-site shape, sent explicitly so
// both engines produce an identical payload.
func typeWireValue(gatewayType azurevirtualnetworkgatewayv1alpha1.AzureVirtualNetworkGatewayType) string {
	if gatewayType == azurevirtualnetworkgatewayv1alpha1.AzureVirtualNetworkGatewayType_EXPRESS_ROUTE {
		return "ExpressRoute"
	}
	return "Vpn"
}

// vpnTypeWireValue maps the routing-model enum to the wire vocabulary.
// Unspecified deploys RouteBased -- the modern model, sent explicitly so
// both engines produce an identical payload.
func vpnTypeWireValue(vpnType azurevirtualnetworkgatewayv1alpha1.AzureVirtualNetworkGatewayVpnType) string {
	if vpnType == azurevirtualnetworkgatewayv1alpha1.AzureVirtualNetworkGatewayVpnType_POLICY_BASED {
		return "PolicyBased"
	}
	return "RouteBased"
}

// skuWireValue maps the SKU enum to azurerm's exact (case-sensitive)
// vocabulary. The spec requires a non-zero SKU, so the empty default is
// unreachable for valid manifests. The non-AZ VpnGw1-5 cases are gone
// with their retired spec values: ARM rejects new non-AZ VPN gateway
// creates (NonAzSkusNotAllowedForVPNGateway, live-confirmed).
func skuWireValue(sku azurevirtualnetworkgatewayv1alpha1.AzureVirtualNetworkGatewaySku) string {
	switch sku {
	case azurevirtualnetworkgatewayv1alpha1.AzureVirtualNetworkGatewaySku_BASIC:
		return "Basic"
	case azurevirtualnetworkgatewayv1alpha1.AzureVirtualNetworkGatewaySku_STANDARD:
		return "Standard"
	case azurevirtualnetworkgatewayv1alpha1.AzureVirtualNetworkGatewaySku_HIGH_PERFORMANCE:
		return "HighPerformance"
	case azurevirtualnetworkgatewayv1alpha1.AzureVirtualNetworkGatewaySku_ULTRA_PERFORMANCE:
		return "UltraPerformance"
	case azurevirtualnetworkgatewayv1alpha1.AzureVirtualNetworkGatewaySku_VPN_GW_1_AZ:
		return "VpnGw1AZ"
	case azurevirtualnetworkgatewayv1alpha1.AzureVirtualNetworkGatewaySku_VPN_GW_2_AZ:
		return "VpnGw2AZ"
	case azurevirtualnetworkgatewayv1alpha1.AzureVirtualNetworkGatewaySku_VPN_GW_3_AZ:
		return "VpnGw3AZ"
	case azurevirtualnetworkgatewayv1alpha1.AzureVirtualNetworkGatewaySku_VPN_GW_4_AZ:
		return "VpnGw4AZ"
	case azurevirtualnetworkgatewayv1alpha1.AzureVirtualNetworkGatewaySku_VPN_GW_5_AZ:
		return "VpnGw5AZ"
	case azurevirtualnetworkgatewayv1alpha1.AzureVirtualNetworkGatewaySku_ER_GW_1_AZ:
		return "ErGw1AZ"
	case azurevirtualnetworkgatewayv1alpha1.AzureVirtualNetworkGatewaySku_ER_GW_2_AZ:
		return "ErGw2AZ"
	case azurevirtualnetworkgatewayv1alpha1.AzureVirtualNetworkGatewaySku_ER_GW_3_AZ:
		return "ErGw3AZ"
	case azurevirtualnetworkgatewayv1alpha1.AzureVirtualNetworkGatewaySku_ER_GW_SCALE:
		return "ErGwScale"
	default:
		return ""
	}
}

// generationWireValue maps the generation enum to the wire vocabulary.
// Returns "" for unspecified so callers omit the field -- the provider
// treats it as Computed and Azure picks the SKU's default generation.
func generationWireValue(generation azurevirtualnetworkgatewayv1alpha1.AzureVirtualNetworkGatewayGeneration) string {
	switch generation {
	case azurevirtualnetworkgatewayv1alpha1.AzureVirtualNetworkGatewayGeneration_GENERATION1:
		return "Generation1"
	case azurevirtualnetworkgatewayv1alpha1.AzureVirtualNetworkGatewayGeneration_GENERATION2:
		return "Generation2"
	case azurevirtualnetworkgatewayv1alpha1.AzureVirtualNetworkGatewayGeneration_NONE:
		return "None"
	default:
		return ""
	}
}

// allocationWireValue maps the private-IP allocation enum to the wire
// vocabulary. Unspecified uses Dynamic -- Azure's default, sent
// explicitly so both engines produce an identical payload.
func allocationWireValue(allocation azurevirtualnetworkgatewayv1alpha1.AzureVirtualNetworkGatewayIpAllocation) string {
	if allocation == azurevirtualnetworkgatewayv1alpha1.AzureVirtualNetworkGatewayIpAllocation_STATIC {
		return "Static"
	}
	return "Dynamic"
}

// natRuleModeWireValue maps the NAT direction enum to the wire
// vocabulary. Unspecified uses EgressSnat -- the provider's default.
func natRuleModeWireValue(mode azurevirtualnetworkgatewayv1alpha1.AzureVirtualNetworkGatewayNatRuleMode) string {
	if mode == azurevirtualnetworkgatewayv1alpha1.AzureVirtualNetworkGatewayNatRuleMode_INGRESS_SNAT {
		return "IngressSnat"
	}
	return "EgressSnat"
}

// natRuleTypeWireValue maps the NAT type enum to the wire vocabulary.
// Unspecified uses Static -- the provider's default.
func natRuleTypeWireValue(ruleType azurevirtualnetworkgatewayv1alpha1.AzureVirtualNetworkGatewayNatRuleType) string {
	if ruleType == azurevirtualnetworkgatewayv1alpha1.AzureVirtualNetworkGatewayNatRuleType_DYNAMIC_NAT {
		return "Dynamic"
	}
	return "Static"
}
