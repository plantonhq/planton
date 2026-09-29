package module

import (
	"strings"

	azurepublicipv1alpha1 "github.com/plantonhq/planton/catalog/azure/azurepublicip/v1alpha1"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/provider/azure/azuretagkeys"
	"github.com/plantonhq/planton/shared/cloudresourcekind"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

type Locals struct {
	AzurePublicIp *azurepublicipv1alpha1.AzurePublicIp

	// ResourceGroupName and PublicIpPrefixId are StringValueOrRef fields;
	// the platform middleware resolves valueFrom references before IaC
	// modules run, so GetValue() always returns the resolved literal.
	ResourceGroupName string
	PublicIpPrefixId  string

	// Sku, SkuTier, IpVersion, DomainNameLabelScope, and DdosProtectionMode
	// are the ARM strings for the spec's enums, or empty when unspecified so
	// both engines let Azure apply its defaults (Standard / Regional / IPv4
	// / region-unique label / VirtualNetworkInherited).
	Sku                  string
	SkuTier              string
	IpVersion            string
	DomainNameLabelScope string
	DdosProtectionMode   string

	// AzureTags is the metadata-derived tag map with the spec's user tags
	// merged over it (user tags win on key collision), mirroring the
	// Terraform module's merge order.
	AzureTags map[string]string
}

func initializeLocals(ctx *pulumi.Context, stackInput *azurepublicipv1alpha1.AzurePublicIpStackInput) *Locals {
	locals := &Locals{}

	locals.AzurePublicIp = stackInput.Target
	target := stackInput.Target

	locals.ResourceGroupName = target.Spec.ResourceGroup.GetValue()
	locals.PublicIpPrefixId = target.Spec.PublicIpPrefixId.GetValue()

	switch target.Spec.Sku {
	case azurepublicipv1alpha1.AzurePublicIpSku_STANDARD:
		locals.Sku = "Standard"
	case azurepublicipv1alpha1.AzurePublicIpSku_STANDARD_V2:
		locals.Sku = "StandardV2"
	}

	switch target.Spec.SkuTier {
	case azurepublicipv1alpha1.AzurePublicIpSkuTier_REGIONAL:
		locals.SkuTier = "Regional"
	case azurepublicipv1alpha1.AzurePublicIpSkuTier_GLOBAL:
		locals.SkuTier = "Global"
	}

	switch target.Spec.IpVersion {
	case azurepublicipv1alpha1.AzurePublicIpIpVersion_IPV4:
		locals.IpVersion = "IPv4"
	case azurepublicipv1alpha1.AzurePublicIpIpVersion_IPV6:
		locals.IpVersion = "IPv6"
	}

	switch target.Spec.DomainNameLabelScope {
	case azurepublicipv1alpha1.AzurePublicIpDomainNameLabelScope_TENANT_REUSE:
		locals.DomainNameLabelScope = "TenantReuse"
	case azurepublicipv1alpha1.AzurePublicIpDomainNameLabelScope_SUBSCRIPTION_REUSE:
		locals.DomainNameLabelScope = "SubscriptionReuse"
	case azurepublicipv1alpha1.AzurePublicIpDomainNameLabelScope_RESOURCE_GROUP_REUSE:
		locals.DomainNameLabelScope = "ResourceGroupReuse"
	case azurepublicipv1alpha1.AzurePublicIpDomainNameLabelScope_NO_REUSE:
		locals.DomainNameLabelScope = "NoReuse"
	}

	switch target.Spec.DdosProtectionMode {
	case azurepublicipv1alpha1.AzurePublicIpDdosProtectionMode_DISABLED:
		locals.DdosProtectionMode = "Disabled"
	case azurepublicipv1alpha1.AzurePublicIpDdosProtectionMode_ENABLED:
		locals.DdosProtectionMode = "Enabled"
	}

	// Metadata-derived tags first, then the user's spec tags merged over
	// them: user tags deliberately win so an org's governance conventions
	// (cost center, owner) can override the derived values where they
	// collide.
	locals.AzureTags = map[string]string{
		azuretagkeys.Resource:     "true",
		azuretagkeys.ResourceName: target.Metadata.Name,
		azuretagkeys.ResourceKind: strings.ToLower(cloudresourcekind.CloudResourceKind_AzurePublicIp.String()),
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
