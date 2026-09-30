package module

import (
	"strings"

	azurevirtualwanv1alpha1 "github.com/plantonhq/planton/catalog/azure/azurevirtualwan/v1alpha1"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/provider/azure/azuretagkeys"
	"github.com/plantonhq/planton/shared/cloudresourcekind"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

type Locals struct {
	AzureVirtualWan *azurevirtualwanv1alpha1.AzureVirtualWan

	// ResourceGroupName is a StringValueOrRef field; the platform middleware
	// resolves valueFrom references before IaC modules run, so GetValue()
	// always returns the resolved literal name.
	ResourceGroupName string

	// AzureTags is the metadata-derived tag map with the spec's user tags
	// merged over it (user tags win on key collision), mirroring the
	// Terraform module's merge order.
	AzureTags map[string]string
}

func initializeLocals(ctx *pulumi.Context, stackInput *azurevirtualwanv1alpha1.AzureVirtualWanStackInput) *Locals {
	locals := &Locals{}

	locals.AzureVirtualWan = stackInput.Target
	target := stackInput.Target

	locals.ResourceGroupName = target.Spec.ResourceGroup.GetValue()

	// Metadata-derived tags first, then the user's spec tags merged over
	// them: user tags deliberately win so an org's governance conventions
	// (cost center, owner) can override the derived values where they
	// collide.
	locals.AzureTags = map[string]string{
		azuretagkeys.Resource:     "true",
		azuretagkeys.ResourceName: target.Metadata.Name,
		azuretagkeys.ResourceKind: strings.ToLower(cloudresourcekind.CloudResourceKind_AzureVirtualWan.String()),
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

// office365BreakoutWireValue maps the spec's optional breakout enum
// onto ARM's vocabulary, applying ARM's default (None) when the field
// is unset -- mirroring the Terraform variable default.
func office365BreakoutWireValue(category *azurevirtualwanv1alpha1.AzureVirtualWanOffice365BreakoutCategory) string {
	if category == nil {
		return "None"
	}
	switch *category {
	case azurevirtualwanv1alpha1.AzureVirtualWanOffice365BreakoutCategory_ALL:
		return "All"
	case azurevirtualwanv1alpha1.AzureVirtualWanOffice365BreakoutCategory_OPTIMIZE:
		return "Optimize"
	case azurevirtualwanv1alpha1.AzureVirtualWanOffice365BreakoutCategory_OPTIMIZE_AND_ALLOW:
		return "OptimizeAndAllow"
	default:
		return "None"
	}
}

// optionalBool returns the pointed-to value, or the default when the
// optional field is unset -- mirroring the Terraform variable default.
func optionalBool(value *bool, defaultValue bool) bool {
	if value == nil {
		return defaultValue
	}
	return *value
}

// optionalString returns the pointed-to value, or the default when the
// optional field is unset -- mirroring the Terraform variable default.
func optionalString(value *string, defaultValue string) string {
	if value == nil || *value == "" {
		return defaultValue
	}
	return *value
}
