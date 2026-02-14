package module

import (
	"strings"

	azurecontainerappenvironmentv1 "github.com/plantonhq/openmcf/apis/org/openmcf/provider/azure/azurecontainerappenvironment/v1"
	"github.com/plantonhq/openmcf/apis/org/openmcf/shared/cloudresourcekind"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

type Locals struct {
	AzureContainerAppEnvironment *azurecontainerappenvironmentv1.AzureContainerAppEnvironment
	ResourceGroupName            string
	AzureTags                    map[string]string
}

func initializeLocals(ctx *pulumi.Context, stackInput *azurecontainerappenvironmentv1.AzureContainerAppEnvironmentStackInput) *Locals {
	locals := &Locals{}

	locals.AzureContainerAppEnvironment = stackInput.Target
	target := stackInput.Target

	locals.ResourceGroupName = target.Spec.ResourceGroup.GetValue()

	// Create Azure tags for resource tagging
	locals.AzureTags = map[string]string{
		"resource":      "true",
		"resource_name": target.Metadata.Name,
		"resource_kind": strings.ToLower(cloudresourcekind.CloudResourceKind_AzureContainerAppEnvironment.String()),
	}

	if target.Metadata.Id != "" {
		locals.AzureTags["resource_id"] = target.Metadata.Id
	}

	if target.Metadata.Org != "" {
		locals.AzureTags["organization"] = target.Metadata.Org
	}

	if target.Metadata.Env != "" {
		locals.AzureTags["environment"] = target.Metadata.Env
	}

	return locals
}
