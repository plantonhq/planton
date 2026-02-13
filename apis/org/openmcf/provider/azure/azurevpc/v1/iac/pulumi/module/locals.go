package module

import (
	"fmt"
	"strings"

	azurevpcv1 "github.com/plantonhq/openmcf/apis/org/openmcf/provider/azure/azurevpc/v1"
	"github.com/plantonhq/openmcf/apis/org/openmcf/shared/cloudresourcekind"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

type Locals struct {
	AzureVpc       *azurevpcv1.AzureVpc
	VNetName       string
	SubnetName     string
	ResourceGroup  string
	Location       string
	NatGatewayName string
	AzureTags      map[string]string
}

func initializeLocals(ctx *pulumi.Context, stackInput *azurevpcv1.AzureVpcStackInput) *Locals {
	locals := &Locals{}

	locals.AzureVpc = stackInput.Target
	target := stackInput.Target

	// Generate resource names from metadata
	locals.VNetName = fmt.Sprintf("vnet-%s", target.Metadata.Name)
	locals.SubnetName = fmt.Sprintf("subnet-nodes-%s", target.Metadata.Name)
	locals.NatGatewayName = fmt.Sprintf("natgw-%s", target.Metadata.Name)

	// Resource group and location are provided externally via spec
	locals.ResourceGroup = target.Spec.ResourceGroup.GetValue()
	locals.Location = target.Spec.Region

	// Create Azure tags for resource tagging
	locals.AzureTags = map[string]string{
		"resource":      "true",
		"resource_name": target.Metadata.Name,
		"resource_kind": strings.ToLower(cloudresourcekind.CloudResourceKind_AzureVpc.String()),
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

	// Merge user-provided tags
	for k, v := range target.Spec.Tags {
		locals.AzureTags[k] = v
	}

	return locals
}
