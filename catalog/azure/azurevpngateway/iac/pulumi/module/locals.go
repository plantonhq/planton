package module

import (
	"strings"

	azurevpngatewayv1alpha1 "github.com/plantonhq/planton/catalog/azure/azurevpngateway/v1alpha1"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/provider/azure/azuretagkeys"
	"github.com/plantonhq/planton/shared/cloudresourcekind"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

type Locals struct {
	AzureVpnGateway *azurevpngatewayv1alpha1.AzureVpnGateway

	// ResourceGroupName is a StringValueOrRef field; the platform middleware
	// resolves valueFrom references before IaC modules run, so GetValue()
	// always returns the resolved literal name.
	ResourceGroupName string

	// AzureTags is the metadata-derived tag map with the spec's user tags
	// merged over it (user tags win on key collision), mirroring the
	// Terraform module's merge order.
	AzureTags map[string]string
}

func initializeLocals(ctx *pulumi.Context, stackInput *azurevpngatewayv1alpha1.AzureVpnGatewayStackInput) *Locals {
	locals := &Locals{}

	locals.AzureVpnGateway = stackInput.Target
	target := stackInput.Target

	locals.ResourceGroupName = target.Spec.ResourceGroup.GetValue()

	// Metadata-derived tags first, then the user's spec tags merged over
	// them: user tags deliberately win so an org's governance conventions
	// (cost center, owner) can override the derived values where they
	// collide.
	locals.AzureTags = map[string]string{
		azuretagkeys.Resource:     "true",
		azuretagkeys.ResourceName: target.Metadata.Name,
		azuretagkeys.ResourceKind: strings.ToLower(cloudresourcekind.CloudResourceKind_AzureVpnGateway.String()),
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

// routingPreferenceWireValue maps the spec's optional routing preference
// enum onto ARM's vocabulary, applying ARM's default ("Microsoft
// Network" -- note the space) when the field is unset -- mirroring the
// Terraform module's null handling.
func routingPreferenceWireValue(preference *azurevpngatewayv1alpha1.AzureVpnGatewayRoutingPreference) string {
	if preference != nil && *preference == azurevpngatewayv1alpha1.AzureVpnGatewayRoutingPreference_INTERNET {
		return "Internet"
	}
	return "Microsoft Network"
}

// natRuleModeWireValue maps the NAT rule mode enum onto ARM's
// vocabulary, applying ARM's default (EgressSnat) when unspecified.
func natRuleModeWireValue(mode azurevpngatewayv1alpha1.AzureVpnGatewayNatRuleMode) string {
	if mode == azurevpngatewayv1alpha1.AzureVpnGatewayNatRuleMode_INGRESS_SNAT {
		return "IngressSnat"
	}
	return "EgressSnat"
}

// natRuleTypeWireValue maps the NAT rule type enum onto ARM's
// vocabulary, applying ARM's default (Static) when unspecified.
func natRuleTypeWireValue(natRuleType azurevpngatewayv1alpha1.AzureVpnGatewayNatRuleType) string {
	if natRuleType == azurevpngatewayv1alpha1.AzureVpnGatewayNatRuleType_DYNAMIC_NAT {
		return "Dynamic"
	}
	return "Static"
}

// optionalInt32 returns the pointed-to value, or the default when the
// optional field is unset -- mirroring the Terraform variable default.
func optionalInt32(value *int32, defaultValue int32) int32 {
	if value == nil {
		return defaultValue
	}
	return *value
}
