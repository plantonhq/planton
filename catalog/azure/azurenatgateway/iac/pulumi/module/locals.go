package module

import (
	"strings"

	azurenatgatewayv1alpha1 "github.com/plantonhq/planton/catalog/azure/azurenatgateway/v1alpha1"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/provider/azure/azuretagkeys"
	"github.com/plantonhq/planton/shared/catalogkind"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

type Locals struct {
	AzureNatGateway *azurenatgatewayv1alpha1.AzureNatGateway

	// ResourceGroupName is a StringValueOrRef field; the platform middleware
	// resolves valueFrom references before IaC modules run, so GetValue()
	// always returns the resolved literal name.
	ResourceGroupName string

	// PublicIpIds and PublicIpPrefixIds are the resolved ARM IDs of the
	// addresses/ranges the gateway SNATs through (repeated StringValueOrRef
	// fields in the spec).
	PublicIpIds       []string
	PublicIpPrefixIds []string

	// SkuName is the ARM string for the spec's enum, or empty when
	// unspecified so both engines let Azure apply its default (Standard).
	SkuName string

	// AzureTags is the metadata-derived tag map with the spec's user tags
	// merged over it (user tags win on key collision), mirroring the
	// Terraform module's merge order.
	AzureTags map[string]string
}

func initializeLocals(ctx *pulumi.Context, iacInput *azurenatgatewayv1alpha1.AzureNatGatewayIacInput) *Locals {
	locals := &Locals{}

	locals.AzureNatGateway = iacInput.Target
	target := iacInput.Target

	locals.ResourceGroupName = target.Spec.ResourceGroup.GetValue()

	for _, publicIpId := range target.Spec.PublicIpIds {
		locals.PublicIpIds = append(locals.PublicIpIds, publicIpId.GetValue())
	}
	for _, prefixId := range target.Spec.PublicIpPrefixIds {
		locals.PublicIpPrefixIds = append(locals.PublicIpPrefixIds, prefixId.GetValue())
	}

	switch target.Spec.SkuName {
	case azurenatgatewayv1alpha1.AzureNatGatewaySkuName_STANDARD:
		locals.SkuName = "Standard"
	case azurenatgatewayv1alpha1.AzureNatGatewaySkuName_STANDARD_V2:
		locals.SkuName = "StandardV2"
	}

	// Metadata-derived tags first, then the user's spec tags merged over
	// them: user tags deliberately win so an org's governance conventions
	// (cost center, owner) can override the derived values where they
	// collide.
	locals.AzureTags = map[string]string{
		azuretagkeys.Resource:     "true",
		azuretagkeys.ResourceName: target.Metadata.Name,
		azuretagkeys.ResourceKind: strings.ToLower(catalogkind.CatalogKind_AzureNatGateway.String()),
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
