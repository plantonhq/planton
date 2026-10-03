package module

import (
	"strings"

	azurefirewallv1alpha1 "github.com/plantonhq/planton/catalog/azure/azurefirewall/v1alpha1"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/provider/azure/azuretagkeys"
	"github.com/plantonhq/planton/shared/catalogkind"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

type Locals struct {
	AzureFirewall *azurefirewallv1alpha1.AzureFirewall

	// ResourceGroupName is a StringValueOrRef field; the platform middleware
	// resolves valueFrom references before IaC modules run, so GetValue()
	// always returns the resolved literal name.
	ResourceGroupName string

	// AzureTags is the metadata-derived tag map with the spec's user tags
	// merged over it (user tags win on key collision), mirroring the
	// Terraform module's merge order.
	AzureTags map[string]string
}

func initializeLocals(ctx *pulumi.Context, iacInput *azurefirewallv1alpha1.AzureFirewallIacInput) *Locals {
	locals := &Locals{}

	locals.AzureFirewall = iacInput.Target
	target := iacInput.Target

	locals.ResourceGroupName = target.Spec.ResourceGroup.GetValue()

	// Metadata-derived tags first, then the user's spec tags merged over
	// them: user tags deliberately win so an org's governance conventions
	// (cost center, owner) can override the derived values where they
	// collide.
	locals.AzureTags = map[string]string{
		azuretagkeys.Resource:     "true",
		azuretagkeys.ResourceName: target.Metadata.Name,
		azuretagkeys.ResourceKind: strings.ToLower(catalogkind.CatalogKind_AzureFirewall.String()),
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

// skuNameWireValue maps the deployment-model enum to the wire vocabulary.
// Unspecified deploys AZFW_VNet -- the standard hub-spoke shape, sent
// explicitly so both engines produce an identical payload.
func skuNameWireValue(skuName azurefirewallv1alpha1.AzureFirewallSkuName) string {
	if skuName == azurefirewallv1alpha1.AzureFirewallSkuName_AZFW_HUB {
		return "AZFW_Hub"
	}
	return "AZFW_VNet"
}

// skuTierWireValue maps the tier enum to the wire vocabulary. Unspecified
// deploys Standard -- the production default, sent explicitly so both
// engines produce an identical payload.
func skuTierWireValue(skuTier azurefirewallv1alpha1.AzureFirewallSkuTier) string {
	switch skuTier {
	case azurefirewallv1alpha1.AzureFirewallSkuTier_BASIC:
		return "Basic"
	case azurefirewallv1alpha1.AzureFirewallSkuTier_PREMIUM:
		return "Premium"
	default:
		return "Standard"
	}
}

// threatIntelModeWireValue maps the threat-intelligence mode to the wire
// vocabulary. Returns "" for unspecified so callers omit the field --
// the ARM field is server-defaulted (Alert) and the provider treats it
// as Computed, so omission lets Azure own the default.
func threatIntelModeWireValue(mode azurefirewallv1alpha1.AzureFirewallThreatIntelMode) string {
	switch mode {
	case azurefirewallv1alpha1.AzureFirewallThreatIntelMode_ALERT:
		return "Alert"
	case azurefirewallv1alpha1.AzureFirewallThreatIntelMode_DENY:
		return "Deny"
	case azurefirewallv1alpha1.AzureFirewallThreatIntelMode_OFF:
		return "Off"
	default:
		return ""
	}
}
