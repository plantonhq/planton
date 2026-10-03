package module

import (
	"strings"

	azurefirewallpolicyv1alpha1 "github.com/plantonhq/planton/catalog/azure/azurefirewallpolicy/v1alpha1"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/provider/azure/azuretagkeys"
	"github.com/plantonhq/planton/shared/catalogkind"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

type Locals struct {
	AzureFirewallPolicy *azurefirewallpolicyv1alpha1.AzureFirewallPolicy

	// ResourceGroupName is a StringValueOrRef field; the platform middleware
	// resolves valueFrom references before IaC modules run, so GetValue()
	// always returns the resolved literal name.
	ResourceGroupName string

	// AzureTags is the metadata-derived tag map with the spec's user tags
	// merged over it (user tags win on key collision), mirroring the
	// Terraform module's merge order.
	AzureTags map[string]string
}

func initializeLocals(ctx *pulumi.Context, iacInput *azurefirewallpolicyv1alpha1.AzureFirewallPolicyIacInput) *Locals {
	locals := &Locals{}

	locals.AzureFirewallPolicy = iacInput.Target
	target := iacInput.Target

	locals.ResourceGroupName = target.Spec.ResourceGroup.GetValue()

	// Metadata-derived tags first, then the user's spec tags merged over
	// them: user tags deliberately win so an org's governance conventions
	// (cost center, owner) can override the derived values where they
	// collide.
	locals.AzureTags = map[string]string{
		azuretagkeys.Resource:     "true",
		azuretagkeys.ResourceName: target.Metadata.Name,
		azuretagkeys.ResourceKind: strings.ToLower(catalogkind.CatalogKind_AzureFirewallPolicy.String()),
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

// skuWireValue maps the spec's sku enum to the azurerm wire vocabulary.
// Unspecified deploys "Standard" -- the provider's own default, sent
// explicitly so both engines produce an identical payload.
func skuWireValue(sku azurefirewallpolicyv1alpha1.AzureFirewallPolicySku) string {
	switch sku {
	case azurefirewallpolicyv1alpha1.AzureFirewallPolicySku_BASIC:
		return "Basic"
	case azurefirewallpolicyv1alpha1.AzureFirewallPolicySku_PREMIUM:
		return "Premium"
	default:
		return "Standard"
	}
}

// threatIntelModeWireValue maps the spec's threat-intelligence mode to the
// wire vocabulary. Unspecified deploys "Alert" -- Azure's default, sent
// explicitly so both engines produce an identical payload.
func threatIntelModeWireValue(mode azurefirewallpolicyv1alpha1.AzureFirewallPolicyThreatIntelligenceMode) string {
	switch mode {
	case azurefirewallpolicyv1alpha1.AzureFirewallPolicyThreatIntelligenceMode_DENY:
		return "Deny"
	case azurefirewallpolicyv1alpha1.AzureFirewallPolicyThreatIntelligenceMode_OFF:
		return "Off"
	default:
		return "Alert"
	}
}

// idpsStateWireValue maps the IDPS state enum to the wire vocabulary. The
// proto values carry an IDPS_ prefix purely for proto scoping; the wire
// values are Azure's plain Off/Alert/Deny. Returns "" for unspecified so
// callers can omit the field.
func idpsStateWireValue(state azurefirewallpolicyv1alpha1.AzureFirewallPolicyIntrusionDetectionState) string {
	switch state {
	case azurefirewallpolicyv1alpha1.AzureFirewallPolicyIntrusionDetectionState_IDPS_OFF:
		return "Off"
	case azurefirewallpolicyv1alpha1.AzureFirewallPolicyIntrusionDetectionState_IDPS_ALERT:
		return "Alert"
	case azurefirewallpolicyv1alpha1.AzureFirewallPolicyIntrusionDetectionState_IDPS_DENY:
		return "Deny"
	default:
		return ""
	}
}

// bypassProtocolWireValue maps the IDPS bypass protocol enum to the wire
// vocabulary (the provider's enum constants are uppercase; ARM echoes
// mixed case and the provider case-suppresses the diff).
func bypassProtocolWireValue(protocol azurefirewallpolicyv1alpha1.AzureFirewallPolicyIdpsBypassProtocol) string {
	switch protocol {
	case azurefirewallpolicyv1alpha1.AzureFirewallPolicyIdpsBypassProtocol_TCP:
		return "TCP"
	case azurefirewallpolicyv1alpha1.AzureFirewallPolicyIdpsBypassProtocol_UDP:
		return "UDP"
	case azurefirewallpolicyv1alpha1.AzureFirewallPolicyIdpsBypassProtocol_ICMP:
		return "ICMP"
	default:
		return "ANY"
	}
}

// identityTypeWireValue maps the identity model enum to the wire
// vocabulary shared by every azurerm identity block.
func identityTypeWireValue(identityType azurefirewallpolicyv1alpha1.AzureFirewallPolicyIdentityType) string {
	switch identityType {
	case azurefirewallpolicyv1alpha1.AzureFirewallPolicyIdentityType_SYSTEM_ASSIGNED:
		return "SystemAssigned"
	case azurefirewallpolicyv1alpha1.AzureFirewallPolicyIdentityType_SYSTEM_AND_USER_ASSIGNED:
		return "SystemAssigned, UserAssigned"
	default:
		return "UserAssigned"
	}
}
