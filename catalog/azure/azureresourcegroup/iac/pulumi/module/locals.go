package module

import (
	"strings"

	azureresourcegroupv1alpha1 "github.com/plantonhq/planton/catalog/azure/azureresourcegroup/v1alpha1"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/provider/azure/azuretagkeys"
	"github.com/plantonhq/planton/shared/cloudresourcekind"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

type Locals struct {
	AzureResourceGroup *azureresourcegroupv1alpha1.AzureResourceGroup
	AzureTags          map[string]string
}

func initializeLocals(ctx *pulumi.Context, stackInput *azureresourcegroupv1alpha1.AzureResourceGroupStackInput) *Locals {
	locals := &Locals{}

	locals.AzureResourceGroup = stackInput.Target
	target := stackInput.Target

	// Create Azure tags for resource tagging
	locals.AzureTags = map[string]string{
		azuretagkeys.Resource:     "true",
		azuretagkeys.ResourceName: target.Metadata.Name,
		azuretagkeys.ResourceKind: strings.ToLower(cloudresourcekind.CloudResourceKind_AzureResourceGroup.String()),
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
