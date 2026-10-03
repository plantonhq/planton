package module

import (
	"strings"

	azureresourcegroupv1alpha1 "github.com/plantonhq/planton/catalog/azure/azureresourcegroup/v1alpha1"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/provider/azure/azuretagkeys"
	"github.com/plantonhq/planton/shared/catalogkind"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

type Locals struct {
	AzureResourceGroup *azureresourcegroupv1alpha1.AzureResourceGroup
	AzureTags          map[string]string
}

func initializeLocals(ctx *pulumi.Context, iacInput *azureresourcegroupv1alpha1.AzureResourceGroupIacInput) *Locals {
	locals := &Locals{}

	locals.AzureResourceGroup = iacInput.Target
	target := iacInput.Target

	// Create Azure tags for resource tagging
	locals.AzureTags = map[string]string{
		azuretagkeys.Resource:     "true",
		azuretagkeys.ResourceName: target.Metadata.Name,
		azuretagkeys.ResourceKind: strings.ToLower(catalogkind.CatalogKind_AzureResourceGroup.String()),
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

	return locals
}
