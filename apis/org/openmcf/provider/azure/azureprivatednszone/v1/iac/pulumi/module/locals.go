package module

import (
	"strings"

	azureprivatednszonev1 "github.com/plantonhq/openmcf/apis/org/openmcf/provider/azure/azureprivatednszone/v1"
	"github.com/plantonhq/openmcf/apis/org/openmcf/shared/cloudresourcekind"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

type Locals struct {
	AzurePrivateDnsZone *azureprivatednszonev1.AzurePrivateDnsZone
	ResourceGroupName   string
	AzureTags           map[string]string
}

func initializeLocals(ctx *pulumi.Context, stackInput *azureprivatednszonev1.AzurePrivateDnsZoneStackInput) *Locals {
	locals := &Locals{}

	locals.AzurePrivateDnsZone = stackInput.Target
	target := stackInput.Target

	locals.ResourceGroupName = target.Spec.ResourceGroup.GetValue()

	// Create Azure tags for resource tagging
	locals.AzureTags = map[string]string{
		"resource":      "true",
		"resource_name": target.Metadata.Name,
		"resource_kind": strings.ToLower(cloudresourcekind.CloudResourceKind_AzurePrivateDnsZone.String()),
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
