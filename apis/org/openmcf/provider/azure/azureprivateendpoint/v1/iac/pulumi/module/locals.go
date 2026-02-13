package module

import (
	"strings"

	azureprivateendpointv1 "github.com/plantonhq/openmcf/apis/org/openmcf/provider/azure/azureprivateendpoint/v1"
	"github.com/plantonhq/openmcf/apis/org/openmcf/shared/cloudresourcekind"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

type Locals struct {
	AzurePrivateEndpoint *azureprivateendpointv1.AzurePrivateEndpoint
	ResourceGroupName     string
	AzureTags             map[string]string
}

func initializeLocals(ctx *pulumi.Context, stackInput *azureprivateendpointv1.AzurePrivateEndpointStackInput) *Locals {
	locals := &Locals{}

	locals.AzurePrivateEndpoint = stackInput.Target
	target := stackInput.Target

	locals.ResourceGroupName = target.Spec.ResourceGroup.GetValue()

	// Create Azure tags for resource tagging
	locals.AzureTags = map[string]string{
		"resource":      "true",
		"resource_name": target.Metadata.Name,
		"resource_kind": strings.ToLower(cloudresourcekind.CloudResourceKind_AzurePrivateEndpoint.String()),
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
