package module

import (
	"strings"

	azuresubnetv1 "github.com/plantonhq/openmcf/apis/org/openmcf/provider/azure/azuresubnet/v1"
	"github.com/plantonhq/openmcf/apis/org/openmcf/shared/cloudresourcekind"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

type Locals struct {
	AzureSubnet       *azuresubnetv1.AzureSubnet
	ResourceGroupName string
	VnetName          string
	AzureTags         map[string]string
}

func initializeLocals(ctx *pulumi.Context, stackInput *azuresubnetv1.AzureSubnetStackInput) *Locals {
	locals := &Locals{}

	locals.AzureSubnet = stackInput.Target
	target := stackInput.Target

	locals.ResourceGroupName = target.Spec.ResourceGroup.GetValue()

	// Extract the VNet name from the ARM resource ID.
	// ARM ID format: /subscriptions/{sub}/resourceGroups/{rg}/providers/Microsoft.Network/virtualNetworks/{name}
	vnetId := target.Spec.VnetId.GetValue()
	parts := strings.Split(vnetId, "/")
	locals.VnetName = parts[len(parts)-1]

	// Create Azure tags for resource tagging
	locals.AzureTags = map[string]string{
		"resource":      "true",
		"resource_name": target.Metadata.Name,
		"resource_kind": strings.ToLower(cloudresourcekind.CloudResourceKind_AzureSubnet.String()),
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
