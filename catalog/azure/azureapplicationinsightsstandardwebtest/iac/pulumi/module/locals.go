package module

import (
	"strings"

	azureappinsightswebtestv1 "github.com/plantonhq/planton/catalog/azure/azureapplicationinsightsstandardwebtest/v1alpha1"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/provider/azure/azuretagkeys"
	"github.com/plantonhq/planton/shared/cloudresourcekind"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

type Locals struct {
	AzureApplicationInsightsStandardWebTest *azureappinsightswebtestv1.AzureApplicationInsightsStandardWebTest

	ResourceGroupName     string
	ApplicationInsightsId string

	AzureTags map[string]string
}

func initializeLocals(ctx *pulumi.Context, stackInput *azureappinsightswebtestv1.AzureApplicationInsightsStandardWebTestStackInput) *Locals {
	locals := &Locals{}

	locals.AzureApplicationInsightsStandardWebTest = stackInput.Target
	target := stackInput.Target
	spec := target.Spec

	locals.ResourceGroupName = spec.ResourceGroup.GetValue()
	locals.ApplicationInsightsId = spec.ApplicationInsightsId.GetValue()

	locals.AzureTags = map[string]string{
		azuretagkeys.Resource:     "true",
		azuretagkeys.ResourceName: target.Metadata.Name,
		azuretagkeys.ResourceKind: strings.ToLower(cloudresourcekind.CloudResourceKind_AzureApplicationInsightsStandardWebTest.String()),
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
	for k, v := range spec.Tags {
		locals.AzureTags[k] = v
	}

	return locals
}
