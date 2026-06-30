package module

import (
	"fmt"
	"strings"

	azurenatgatewayv1 "github.com/plantonhq/planton/apis/dev/planton/provider/azure/azurenatgateway/v1"
	"github.com/plantonhq/planton/apis/dev/planton/shared/cloudresourcekind"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

type Locals struct {
	AzureNatGateway *azurenatgatewayv1.AzureNatGateway
	NatGatewayName  string
	SubnetId        string
	ResourceGroup   string
	Location        string
	AzureTags       map[string]string
}

func initializeLocals(ctx *pulumi.Context, stackInput *azurenatgatewayv1.AzureNatGatewayStackInput) *Locals {
	locals := &Locals{}

	locals.AzureNatGateway = stackInput.Target
	target := stackInput.Target

	// Generate NAT Gateway name from metadata
	locals.NatGatewayName = fmt.Sprintf("natgw-%s", target.Metadata.Name)

	// Get subnet ID (either direct value or from reference)
	locals.SubnetId = target.Spec.SubnetId.GetValue()

	// The resource_group and region fields are explicit spec fields.
	// The platform middleware resolves valueFrom references before IaC modules run,
	// so .GetValue() always returns the resolved literal string.
	locals.ResourceGroup = target.Spec.ResourceGroup.GetValue()
	locals.Location = target.Spec.Region

	// Create Azure tags for resource tagging
	locals.AzureTags = map[string]string{
		"resource":      "true",
		"resource_name": target.Metadata.Name,
		"resource_kind": strings.ToLower(cloudresourcekind.CloudResourceKind_AzureNatGateway.String()),
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

// getIdleTimeoutMinutes returns the idle timeout or default of 4
func getIdleTimeoutMinutes(spec *azurenatgatewayv1.AzureNatGatewaySpec) int {
	if spec.IdleTimeoutMinutes != nil {
		return int(*spec.IdleTimeoutMinutes)
	}
	return 4 // Azure default
}
