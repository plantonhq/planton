package module

import (
	"strings"

	azureapplicationgatewayv1 "github.com/plantonhq/planton/apis/dev/planton/provider/azure/azureapplicationgateway/v1"
	"github.com/plantonhq/planton/apis/dev/planton/shared/cloudresourcekind"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

type Locals struct {
	AzureApplicationGateway *azureapplicationgatewayv1.AzureApplicationGateway
	ResourceGroupName       string
	GatewayIpConfigName     string
	FrontendIpConfigName    string
	AzureTags               map[string]string
}

func initializeLocals(ctx *pulumi.Context, stackInput *azureapplicationgatewayv1.AzureApplicationGatewayStackInput) *Locals {
	locals := &Locals{}

	locals.AzureApplicationGateway = stackInput.Target
	target := stackInput.Target

	locals.ResourceGroupName = target.Spec.ResourceGroup.GetValue()

	// Auto-derive the gateway IP configuration name from the App Gateway name.
	locals.GatewayIpConfigName = target.Spec.Name + "-gw-ip-config"

	// Auto-derive the frontend IP configuration name from the App Gateway name.
	locals.FrontendIpConfigName = target.Spec.Name + "-frontend-ip-config"

	// Create Azure tags for resource tagging
	locals.AzureTags = map[string]string{
		"resource":      "true",
		"resource_name": target.Metadata.Name,
		"resource_kind": strings.ToLower(cloudresourcekind.CloudResourceKind_AzureApplicationGateway.String()),
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
