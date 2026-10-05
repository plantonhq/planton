package module

import (
	"strings"

	azurenetworksecuritygroupv1alpha1 "github.com/plantonhq/planton/catalog/azure/azurenetworksecuritygroup/v1alpha1"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/provider/azure/azuretagkeys"
	"github.com/plantonhq/planton/shared/catalogkind"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

type Locals struct {
	AzureNetworkSecurityGroup *azurenetworksecuritygroupv1alpha1.AzureNetworkSecurityGroup

	// ResourceGroupName is a StringValueOrRef field; the platform middleware
	// resolves valueFrom references before IaC modules run, so GetValue()
	// always returns the resolved literal name.
	ResourceGroupName string

	// AzureTags is the metadata-derived tag map with the spec's user tags
	// merged over it (user tags win on key collision), mirroring the
	// Terraform module's merge order.
	AzureTags map[string]string
}

func initializeLocals(ctx *pulumi.Context, iacInput *azurenetworksecuritygroupv1alpha1.AzureNetworkSecurityGroupIacInput) *Locals {
	locals := &Locals{}

	locals.AzureNetworkSecurityGroup = iacInput.Target
	target := iacInput.Target

	locals.ResourceGroupName = target.Spec.ResourceGroup.GetValue()

	// Metadata-derived tags first, then the user's spec tags merged over
	// them: user tags deliberately win so an org's governance conventions
	// (cost center, owner) can override the derived values where they
	// collide.
	locals.AzureTags = map[string]string{
		azuretagkeys.Resource:     "true",
		azuretagkeys.ResourceName: target.Metadata.Name,
		azuretagkeys.ResourceKind: strings.ToLower(catalogkind.CatalogKind_AzureNetworkSecurityGroup.String()),
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

// directionToArm maps the spec's direction enum to ARM's
// SecurityRuleDirection string.
func directionToArm(direction azurenetworksecuritygroupv1alpha1.AzureNetworkSecurityGroupRuleDirection) string {
	switch direction {
	case azurenetworksecuritygroupv1alpha1.AzureNetworkSecurityGroupRuleDirection_INBOUND:
		return "Inbound"
	case azurenetworksecuritygroupv1alpha1.AzureNetworkSecurityGroupRuleDirection_OUTBOUND:
		return "Outbound"
	}
	return ""
}

// accessToArm maps the spec's access enum to ARM's SecurityRuleAccess
// string.
func accessToArm(access azurenetworksecuritygroupv1alpha1.AzureNetworkSecurityGroupRuleAccess) string {
	switch access {
	case azurenetworksecuritygroupv1alpha1.AzureNetworkSecurityGroupRuleAccess_ALLOW:
		return "Allow"
	case azurenetworksecuritygroupv1alpha1.AzureNetworkSecurityGroupRuleAccess_DENY:
		return "Deny"
	}
	return ""
}

// protocolToArm maps the spec's protocol enum to ARM's SecurityRuleProtocol
// string (ANY is ARM's "*").
func protocolToArm(protocol azurenetworksecuritygroupv1alpha1.AzureNetworkSecurityGroupRuleProtocol) string {
	switch protocol {
	case azurenetworksecuritygroupv1alpha1.AzureNetworkSecurityGroupRuleProtocol_ANY:
		return "*"
	case azurenetworksecuritygroupv1alpha1.AzureNetworkSecurityGroupRuleProtocol_TCP:
		return "Tcp"
	case azurenetworksecuritygroupv1alpha1.AzureNetworkSecurityGroupRuleProtocol_UDP:
		return "Udp"
	case azurenetworksecuritygroupv1alpha1.AzureNetworkSecurityGroupRuleProtocol_ICMP:
		return "Icmp"
	case azurenetworksecuritygroupv1alpha1.AzureNetworkSecurityGroupRuleProtocol_AH:
		return "Ah"
	case azurenetworksecuritygroupv1alpha1.AzureNetworkSecurityGroupRuleProtocol_ESP:
		return "Esp"
	}
	return ""
}
