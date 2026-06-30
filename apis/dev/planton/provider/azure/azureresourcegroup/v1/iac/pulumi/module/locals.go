package module

import (
	"strings"

	azureresourcegroupv1 "github.com/plantonhq/planton/apis/dev/planton/provider/azure/azureresourcegroup/v1"
	"github.com/plantonhq/planton/apis/dev/planton/shared/cloudresourcekind"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

type Locals struct {
	AzureResourceGroup *azureresourcegroupv1.AzureResourceGroup
	AzureTags          map[string]string
}

func initializeLocals(ctx *pulumi.Context, stackInput *azureresourcegroupv1.AzureResourceGroupStackInput) *Locals {
	locals := &Locals{}

	locals.AzureResourceGroup = stackInput.Target
	target := stackInput.Target

	// Create Azure tags for resource tagging
	locals.AzureTags = map[string]string{
		"resource":      "true",
		"resource_name": target.Metadata.Name,
		"resource_kind": strings.ToLower(cloudresourcekind.CloudResourceKind_AzureResourceGroup.String()),
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
