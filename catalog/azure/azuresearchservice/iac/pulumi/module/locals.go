package module

import (
	"strings"

	azuresearchservicev1alpha1 "github.com/plantonhq/planton/catalog/azure/azuresearchservice/v1alpha1"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/provider/azure/azuretagkeys"
	"github.com/plantonhq/planton/shared/cloudresourcekind"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

type Locals struct {
	AzureSearchService *azuresearchservicev1alpha1.AzureSearchService

	// ResourceGroupName is a StringValueOrRef field; the platform middleware
	// resolves valueFrom references before IaC modules run, so GetValue()
	// always returns the resolved literal name.
	ResourceGroupName string

	// AzureTags is the metadata-derived tag map with the spec's user tags
	// merged over it (user tags win on key collision), mirroring the
	// Terraform module's merge order.
	AzureTags map[string]string
}

// identityTypeWire maps the spec's identity flavors to the provider's
// comma-joined wire values.
var identityTypeWire = map[azuresearchservicev1alpha1.AzureSearchServiceIdentityType]string{
	azuresearchservicev1alpha1.AzureSearchServiceIdentityType_SYSTEM_ASSIGNED:          "SystemAssigned",
	azuresearchservicev1alpha1.AzureSearchServiceIdentityType_USER_ASSIGNED:            "UserAssigned",
	azuresearchservicev1alpha1.AzureSearchServiceIdentityType_SYSTEM_AND_USER_ASSIGNED: "SystemAssigned, UserAssigned",
}

// hostingModeWire maps the spec's hosting modes to the provider's wire
// values (the ARM enum is camelCase). Unspecified is absent -- the
// property is omitted so the provider applies its default, "default".
var hostingModeWire = map[azuresearchservicev1alpha1.AzureSearchServiceHostingMode]string{
	azuresearchservicev1alpha1.AzureSearchServiceHostingMode_DEFAULT:      "default",
	azuresearchservicev1alpha1.AzureSearchServiceHostingMode_HIGH_DENSITY: "highDensity",
}

func initializeLocals(ctx *pulumi.Context, stackInput *azuresearchservicev1alpha1.AzureSearchServiceStackInput) *Locals {
	locals := &Locals{}

	locals.AzureSearchService = stackInput.Target
	target := stackInput.Target

	locals.ResourceGroupName = target.Spec.ResourceGroup.GetValue()

	// Metadata-derived tags first, then the user's spec tags merged over
	// them: user tags deliberately win so an org's governance conventions
	// (cost center, owner) can override the derived values where they
	// collide.
	locals.AzureTags = map[string]string{
		azuretagkeys.Resource:     "true",
		azuretagkeys.ResourceName: target.Metadata.Name,
		azuretagkeys.ResourceKind: strings.ToLower(cloudresourcekind.CloudResourceKind_AzureSearchService.String()),
	}

	if target.Metadata.Id != "" {
		locals.AzureTags[azuretagkeys.ResourceId] = target.Metadata.Id
	}

	if target.Metadata.Org != "" {
		locals.AzureTags[azuretagkeys.Organization] = target.Metadata.Org
	}

	if target.Metadata.Env != "" {
		locals.AzureTags[azuretagkeys.Environment] = target.Metadata.Env
	}

	for k, v := range target.Spec.Tags {
		locals.AzureTags[k] = v
	}

	return locals
}
