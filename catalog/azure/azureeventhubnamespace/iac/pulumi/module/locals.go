package module

import (
	"strings"

	azureeventhubnamespacev1alpha1 "github.com/plantonhq/planton/catalog/azure/azureeventhubnamespace/v1alpha1"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/provider/azure/azuretagkeys"
	"github.com/plantonhq/planton/shared/catalogkind"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

type Locals struct {
	AzureEventHubNamespace *azureeventhubnamespacev1alpha1.AzureEventHubNamespace
	ResourceGroupName      string
	AzureTags              map[string]string
}

// skuStrings maps the spec's SKU enum to ARM's wire values. The unspecified
// row deploys STANDARD -- the full-featured multi-tenant tier -- because an
// unmapped enum would send the empty string, which the provider rejects.
var skuStrings = map[azureeventhubnamespacev1alpha1.AzureEventHubNamespaceSku]string{
	azureeventhubnamespacev1alpha1.AzureEventHubNamespaceSku_azure_event_hub_namespace_sku_unspecified: "Standard",
	azureeventhubnamespacev1alpha1.AzureEventHubNamespaceSku_BASIC:                                     "Basic",
	azureeventhubnamespacev1alpha1.AzureEventHubNamespaceSku_STANDARD:                                  "Standard",
	azureeventhubnamespacev1alpha1.AzureEventHubNamespaceSku_PREMIUM:                                   "Premium",
}

// identityTypeStrings maps the identity-model enum to ARM's values --
// Event Hubs namespaces support all three managed-identity models.
var identityTypeStrings = map[azureeventhubnamespacev1alpha1.AzureEventHubNamespaceIdentityType]string{
	azureeventhubnamespacev1alpha1.AzureEventHubNamespaceIdentityType_SYSTEM_ASSIGNED:          "SystemAssigned",
	azureeventhubnamespacev1alpha1.AzureEventHubNamespaceIdentityType_USER_ASSIGNED:            "UserAssigned",
	azureeventhubnamespacev1alpha1.AzureEventHubNamespaceIdentityType_SYSTEM_AND_USER_ASSIGNED: "SystemAssigned, UserAssigned",
}

// networkDefaultActionStrings maps the firewall's default-action enum to
// ARM's values. No unspecified row: the spec enum requires an explicit
// choice (Azure itself requires default_action when the rule set is
// declared).
var networkDefaultActionStrings = map[azureeventhubnamespacev1alpha1.AzureEventHubNetworkRuleSetDefaultAction]string{
	azureeventhubnamespacev1alpha1.AzureEventHubNetworkRuleSetDefaultAction_ALLOW: "Allow",
	azureeventhubnamespacev1alpha1.AzureEventHubNetworkRuleSetDefaultAction_DENY:  "Deny",
}

func initializeLocals(ctx *pulumi.Context, iacInput *azureeventhubnamespacev1alpha1.AzureEventHubNamespaceIacInput) *Locals {
	locals := &Locals{}

	locals.AzureEventHubNamespace = iacInput.Target
	target := iacInput.Target

	// The resource_group field is a StringValueOrRef. The platform middleware
	// resolves valueFrom references before IaC modules run, so .GetValue()
	// always returns the resolved literal string.
	locals.ResourceGroupName = target.Spec.ResourceGroup.GetValue()

	// Identity tags derived from metadata; user tags merge OVER these (the
	// governance surface belongs to the user).
	locals.AzureTags = map[string]string{
		azuretagkeys.Resource:     "true",
		azuretagkeys.ResourceName: target.Metadata.Name,
		azuretagkeys.ResourceKind: strings.ToLower(catalogkind.CatalogKind_AzureEventHubNamespace.String()),
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

	for key, value := range target.Spec.Tags {
		locals.AzureTags[key] = value
	}

	return locals
}
