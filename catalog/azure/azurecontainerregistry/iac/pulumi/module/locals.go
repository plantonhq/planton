package module

import (
	"strings"

	azurecontainerregistryv1alpha1 "github.com/plantonhq/planton/catalog/azure/azurecontainerregistry/v1alpha1"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/provider/azure/azuretagkeys"
	"github.com/plantonhq/planton/shared/catalogkind"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

type Locals struct {
	AzureContainerRegistry *azurecontainerregistryv1alpha1.AzureContainerRegistry

	// ResourceGroupName is a StringValueOrRef field; the platform middleware
	// resolves valueFrom references before IaC modules run, so GetValue()
	// always returns the resolved literal name.
	ResourceGroupName string

	// Sku is the ARM SKU string. The STANDARD baseline is materialized
	// here when the spec leaves the SKU unspecified (ARM requires an
	// explicit value), matching the Terraform module's locals mapping.
	Sku string

	// NetworkRuleBypassOption is the ARM string for the spec's enum, or
	// empty when unspecified so both engines let Azure apply its default
	// (AzureServices).
	NetworkRuleBypassOption string

	// IdentityType is the ARM identity-type string ("SystemAssigned",
	// "UserAssigned", "SystemAssigned, UserAssigned"), or empty when the
	// spec declares no identity.
	IdentityType string

	// IdentityIds are the resolved ARM IDs of the user-assigned identities
	// attached to the registry.
	IdentityIds []string

	// AzureTags is the metadata-derived tag map with the spec's user tags
	// merged over it (user tags win on key collision), mirroring the
	// Terraform module's merge order.
	AzureTags map[string]string
}

func initializeLocals(ctx *pulumi.Context, iacInput *azurecontainerregistryv1alpha1.AzureContainerRegistryIacInput) *Locals {
	locals := &Locals{}

	locals.AzureContainerRegistry = iacInput.Target
	target := iacInput.Target

	locals.ResourceGroupName = target.Spec.ResourceGroup.GetValue()

	switch target.Spec.Sku {
	case azurecontainerregistryv1alpha1.AzureContainerRegistrySku_BASIC:
		locals.Sku = "Basic"
	case azurecontainerregistryv1alpha1.AzureContainerRegistrySku_PREMIUM:
		locals.Sku = "Premium"
	default:
		locals.Sku = "Standard"
	}

	switch target.Spec.NetworkRuleBypassOption {
	case azurecontainerregistryv1alpha1.AzureContainerRegistryNetworkRuleBypassOption_AZURE_SERVICES:
		locals.NetworkRuleBypassOption = "AzureServices"
	case azurecontainerregistryv1alpha1.AzureContainerRegistryNetworkRuleBypassOption_NONE:
		locals.NetworkRuleBypassOption = "None"
	}

	if target.Spec.Identity != nil {
		switch target.Spec.Identity.Type {
		case azurecontainerregistryv1alpha1.AzureContainerRegistryIdentityType_SYSTEM_ASSIGNED:
			locals.IdentityType = "SystemAssigned"
		case azurecontainerregistryv1alpha1.AzureContainerRegistryIdentityType_USER_ASSIGNED:
			locals.IdentityType = "UserAssigned"
		case azurecontainerregistryv1alpha1.AzureContainerRegistryIdentityType_SYSTEM_AND_USER_ASSIGNED:
			locals.IdentityType = "SystemAssigned, UserAssigned"
		}
		for _, identityId := range target.Spec.Identity.IdentityIds {
			locals.IdentityIds = append(locals.IdentityIds, identityId.GetValue())
		}
	}

	// Metadata-derived tags first, then the user's spec tags merged over
	// them: user tags deliberately win so an org's governance conventions
	// (cost center, owner) can override the derived values where they
	// collide.
	locals.AzureTags = map[string]string{
		azuretagkeys.Resource:     "true",
		azuretagkeys.ResourceName: target.Metadata.Name,
		azuretagkeys.ResourceKind: strings.ToLower(catalogkind.CatalogKind_AzureContainerRegistry.String()),
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
