package module

import (
	"strings"

	azureloganalyticsworkspacev1alpha1 "github.com/plantonhq/planton/catalog/azure/azureloganalyticsworkspace/v1alpha1"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/provider/azure/azuretagkeys"
	"github.com/plantonhq/planton/shared/catalogkind"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

type Locals struct {
	AzureLogAnalyticsWorkspace *azureloganalyticsworkspacev1alpha1.AzureLogAnalyticsWorkspace
	ResourceGroupName          string
	AzureTags                  map[string]string
}

// skuStrings maps the spec's SKU enum to ARM's wire values. The unspecified
// row deploys Azure's recommended PerGB2018 pay-as-you-go tier -- an
// unmapped enum would send the empty string, which the provider rejects.
// Standard/Premium/LACluster/Unlimited are deliberately not mapped: Azure
// blocks creating workspaces on them (see the spec enum comment).
var skuStrings = map[azureloganalyticsworkspacev1alpha1.AzureLogAnalyticsWorkspaceSku]string{
	azureloganalyticsworkspacev1alpha1.AzureLogAnalyticsWorkspaceSku_azure_log_analytics_workspace_sku_unspecified: "PerGB2018",
	azureloganalyticsworkspacev1alpha1.AzureLogAnalyticsWorkspaceSku_PER_GB_2018:                                   "PerGB2018",
	azureloganalyticsworkspacev1alpha1.AzureLogAnalyticsWorkspaceSku_CAPACITY_RESERVATION:                          "CapacityReservation",
	azureloganalyticsworkspacev1alpha1.AzureLogAnalyticsWorkspaceSku_PER_NODE:                                      "PerNode",
	azureloganalyticsworkspacev1alpha1.AzureLogAnalyticsWorkspaceSku_STANDALONE:                                    "Standalone",
}

// identityTypeStrings maps the identity-model enum to ARM's values.
// Workspaces accept exactly SystemAssigned or UserAssigned -- the combined
// model does not exist on this resource.
var identityTypeStrings = map[azureloganalyticsworkspacev1alpha1.AzureLogAnalyticsWorkspaceIdentityType]string{
	azureloganalyticsworkspacev1alpha1.AzureLogAnalyticsWorkspaceIdentityType_SYSTEM_ASSIGNED: "SystemAssigned",
	azureloganalyticsworkspacev1alpha1.AzureLogAnalyticsWorkspaceIdentityType_USER_ASSIGNED:   "UserAssigned",
}

func initializeLocals(ctx *pulumi.Context, iacInput *azureloganalyticsworkspacev1alpha1.AzureLogAnalyticsWorkspaceIacInput) *Locals {
	locals := &Locals{}

	locals.AzureLogAnalyticsWorkspace = iacInput.Target
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
		azuretagkeys.ResourceKind: strings.ToLower(catalogkind.CatalogKind_AzureLogAnalyticsWorkspace.String()),
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
