package module

import (
	"strings"

	azureloadbalancerv1 "github.com/plantonhq/openmcf/apis/org/openmcf/provider/azure/azureloadbalancer/v1"
	"github.com/plantonhq/openmcf/apis/org/openmcf/shared/cloudresourcekind"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

type Locals struct {
	AzureLoadBalancer      *azureloadbalancerv1.AzureLoadBalancer
	ResourceGroupName      string
	FrontendConfigName     string
	AzureTags              map[string]string
}

func initializeLocals(ctx *pulumi.Context, stackInput *azureloadbalancerv1.AzureLoadBalancerStackInput) *Locals {
	locals := &Locals{}

	locals.AzureLoadBalancer = stackInput.Target
	target := stackInput.Target

	locals.ResourceGroupName = target.Spec.ResourceGroup.GetValue()

	// Auto-derive the frontend IP configuration name from the LB name.
	// This is the internal Azure name used by rules to reference the frontend.
	locals.FrontendConfigName = target.Spec.Name + "-frontend"

	// Create Azure tags for resource tagging
	locals.AzureTags = map[string]string{
		"resource":      "true",
		"resource_name": target.Metadata.Name,
		"resource_kind": strings.ToLower(cloudresourcekind.CloudResourceKind_AzureLoadBalancer.String()),
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
