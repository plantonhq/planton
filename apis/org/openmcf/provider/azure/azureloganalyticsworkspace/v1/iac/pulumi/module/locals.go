package module

import (
	"strings"

	azureloganalyticsworkspacev1 "github.com/plantonhq/openmcf/apis/org/openmcf/provider/azure/azureloganalyticsworkspace/v1"
	"github.com/plantonhq/openmcf/apis/org/openmcf/shared/cloudresourcekind"
	foreignkeyv1 "github.com/plantonhq/openmcf/apis/org/openmcf/shared/foreignkey/v1"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

type Locals struct {
	AzureLogAnalyticsWorkspace *azureloganalyticsworkspacev1.AzureLogAnalyticsWorkspace
	ResourceGroupName          string
	AzureTags                  map[string]string
}

func initializeLocals(ctx *pulumi.Context, stackInput *azureloganalyticsworkspacev1.AzureLogAnalyticsWorkspaceStackInput) *Locals {
	locals := &Locals{}

	locals.AzureLogAnalyticsWorkspace = stackInput.Target
	target := stackInput.Target

	// Resolve resource_group from StringValueOrRef.
	// At runtime, OpenMCF middleware resolves valueFrom references before IaC runs,
	// so we always get the resolved value here.
	locals.ResourceGroupName = resolveStringValueOrRef(target.Spec.ResourceGroup)

	// Create Azure tags for resource tagging
	locals.AzureTags = map[string]string{
		"resource":      "true",
		"resource_name": target.Metadata.Name,
		"resource_kind": strings.ToLower(cloudresourcekind.CloudResourceKind_AzureLogAnalyticsWorkspace.String()),
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

// resolveStringValueOrRef extracts the string value from a StringValueOrRef.
// OpenMCF middleware resolves valueFrom references before IaC modules run,
// so the Value field is always populated at this point.
func resolveStringValueOrRef(ref *foreignkeyv1.StringValueOrRef) string {
	if ref == nil {
		return ""
	}
	switch v := ref.LiteralOrRef.(type) {
	case *foreignkeyv1.StringValueOrRef_Value:
		return v.Value
	default:
		return ""
	}
}
