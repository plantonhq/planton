package module

import (
	"strings"

	azureapplicationinsightsv1 "github.com/plantonhq/openmcf/apis/org/openmcf/provider/azure/azureapplicationinsights/v1"
	"github.com/plantonhq/openmcf/apis/org/openmcf/shared/cloudresourcekind"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

type Locals struct {
	AzureApplicationInsights *azureapplicationinsightsv1.AzureApplicationInsights
	ResourceGroupName        string
	WorkspaceId              string
	AzureTags                map[string]string
}

func initializeLocals(ctx *pulumi.Context, stackInput *azureapplicationinsightsv1.AzureApplicationInsightsStackInput) *Locals {
	locals := &Locals{}

	locals.AzureApplicationInsights = stackInput.Target
	target := stackInput.Target

	// The resource_group field is a StringValueOrRef. The platform middleware resolves
	// valueFrom references before IaC modules run, so .GetValue() always returns the
	// resolved literal string.
	locals.ResourceGroupName = target.Spec.ResourceGroup.GetValue()

	// The workspace_id field is a StringValueOrRef referencing an
	// AzureLogAnalyticsWorkspace output.
	locals.WorkspaceId = target.Spec.WorkspaceId.GetValue()

	// Create Azure tags for resource tagging
	locals.AzureTags = map[string]string{
		"resource":      "true",
		"resource_name": target.Metadata.Name,
		"resource_kind": strings.ToLower(cloudresourcekind.CloudResourceKind_AzureApplicationInsights.String()),
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
