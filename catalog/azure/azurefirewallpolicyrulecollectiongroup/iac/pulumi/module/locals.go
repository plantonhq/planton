package module

import (
	azurefirewallpolicyrulecollectiongroupv1alpha1 "github.com/plantonhq/planton/catalog/azure/azurefirewallpolicyrulecollectiongroup/v1alpha1"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

type Locals struct {
	AzureFirewallPolicyRuleCollectionGroup *azurefirewallpolicyrulecollectiongroupv1alpha1.AzureFirewallPolicyRuleCollectionGroup

	// FirewallPolicyId is the parent policy's ARM id -- a StringValueOrRef
	// resolved to the literal id by the platform middleware before the
	// module runs.
	FirewallPolicyId string
}

// initializeLocals mirrors the Terraform module's locals: the group nests
// under its parent policy by ARM id, and carries no tags (ARM does not
// support tags on rule collection groups -- they are child documents of
// the policy, not top-level tracked resources).
func initializeLocals(ctx *pulumi.Context, iacInput *azurefirewallpolicyrulecollectiongroupv1alpha1.AzureFirewallPolicyRuleCollectionGroupIacInput) *Locals {
	locals := &Locals{}

	locals.AzureFirewallPolicyRuleCollectionGroup = iacInput.Target
	locals.FirewallPolicyId = iacInput.Target.Spec.FirewallPolicyId.GetValue()

	return locals
}

// filterActionWireValue maps the filter-collection action enum to the
// wire vocabulary.
func filterActionWireValue(action azurefirewallpolicyrulecollectiongroupv1alpha1.AzureFirewallPolicyFilterAction) string {
	if action == azurefirewallpolicyrulecollectiongroupv1alpha1.AzureFirewallPolicyFilterAction_DENY {
		return "Deny"
	}
	return "Allow"
}

// ruleProtocolWireValue maps the network/DNAT protocol enum to the wire
// vocabulary. Azure spells the wildcard "Any" with TCP/UDP/ICMP uppercase
// -- an irregular casing the provider validates case-sensitively.
func ruleProtocolWireValue(protocol azurefirewallpolicyrulecollectiongroupv1alpha1.AzureFirewallPolicyRuleProtocol) string {
	switch protocol {
	case azurefirewallpolicyrulecollectiongroupv1alpha1.AzureFirewallPolicyRuleProtocol_TCP:
		return "TCP"
	case azurefirewallpolicyrulecollectiongroupv1alpha1.AzureFirewallPolicyRuleProtocol_UDP:
		return "UDP"
	case azurefirewallpolicyrulecollectiongroupv1alpha1.AzureFirewallPolicyRuleProtocol_ICMP:
		return "ICMP"
	default:
		return "Any"
	}
}

// applicationProtocolTypeWireValue maps the L7 protocol type enum to the
// wire vocabulary (mixed-case Http/Https/Mssql -- the provider validates
// case-sensitively).
func applicationProtocolTypeWireValue(protocolType azurefirewallpolicyrulecollectiongroupv1alpha1.AzureFirewallPolicyApplicationProtocolType) string {
	switch protocolType {
	case azurefirewallpolicyrulecollectiongroupv1alpha1.AzureFirewallPolicyApplicationProtocolType_HTTPS:
		return "Https"
	case azurefirewallpolicyrulecollectiongroupv1alpha1.AzureFirewallPolicyApplicationProtocolType_MSSQL:
		return "Mssql"
	default:
		return "Http"
	}
}
