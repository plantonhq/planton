package module

import (
	"strings"

	azurepublicipprefixv1alpha1 "github.com/plantonhq/planton/catalog/azure/azurepublicipprefix/v1alpha1"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/provider/azure/azuretagkeys"
	"github.com/plantonhq/planton/shared/cloudresourcekind"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

type Locals struct {
	AzurePublicIpPrefix *azurepublicipprefixv1alpha1.AzurePublicIpPrefix

	// ResourceGroupName is a StringValueOrRef field; the platform middleware
	// resolves valueFrom references before IaC modules run, so GetValue()
	// always returns the resolved literal name.
	ResourceGroupName string

	// IpVersion, Sku, and SkuTier are the ARM strings for the spec's enums,
	// or empty when unspecified so both engines let Azure apply its
	// defaults (IPv4 / Standard / Regional).
	IpVersion string
	Sku       string
	SkuTier   string

	// AzureTags is the metadata-derived tag map with the spec's user tags
	// merged over it (user tags win on key collision), mirroring the
	// Terraform module's merge order.
	AzureTags map[string]string
}

func initializeLocals(ctx *pulumi.Context, stackInput *azurepublicipprefixv1alpha1.AzurePublicIpPrefixStackInput) *Locals {
	locals := &Locals{}

	locals.AzurePublicIpPrefix = stackInput.Target
	target := stackInput.Target

	locals.ResourceGroupName = target.Spec.ResourceGroup.GetValue()

	switch target.Spec.IpVersion {
	case azurepublicipprefixv1alpha1.AzurePublicIpPrefixIpVersion_IPV4:
		locals.IpVersion = "IPv4"
	case azurepublicipprefixv1alpha1.AzurePublicIpPrefixIpVersion_IPV6:
		locals.IpVersion = "IPv6"
	}

	switch target.Spec.Sku {
	case azurepublicipprefixv1alpha1.AzurePublicIpPrefixSku_STANDARD:
		locals.Sku = "Standard"
	case azurepublicipprefixv1alpha1.AzurePublicIpPrefixSku_STANDARD_V2:
		locals.Sku = "StandardV2"
	}

	switch target.Spec.SkuTier {
	case azurepublicipprefixv1alpha1.AzurePublicIpPrefixSkuTier_REGIONAL:
		locals.SkuTier = "Regional"
	case azurepublicipprefixv1alpha1.AzurePublicIpPrefixSkuTier_GLOBAL:
		locals.SkuTier = "Global"
	}

	// Metadata-derived tags first, then the user's spec tags merged over
	// them: user tags deliberately win so an org's governance conventions
	// (cost center, owner) can override the derived values where they
	// collide.
	locals.AzureTags = map[string]string{
		azuretagkeys.Resource:     "true",
		azuretagkeys.ResourceName: target.Metadata.Name,
		azuretagkeys.ResourceKind: strings.ToLower(cloudresourcekind.CloudResourceKind_AzurePublicIpPrefix.String()),
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
