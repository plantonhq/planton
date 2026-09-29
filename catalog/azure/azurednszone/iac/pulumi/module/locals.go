package module

import (
	"strings"

	azurednszonev1alpha1 "github.com/plantonhq/planton/catalog/azure/azurednszone/v1alpha1"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/provider/azure/azuretagkeys"
	"github.com/plantonhq/planton/shared/cloudresourcekind"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

type Locals struct {
	AzureDnsZone      *azurednszonev1alpha1.AzureDnsZone
	ResourceGroupName string
	AzureTags         map[string]string
}

func initializeLocals(ctx *pulumi.Context, stackInput *azurednszonev1alpha1.AzureDnsZoneStackInput) *Locals {
	locals := &Locals{}

	locals.AzureDnsZone = stackInput.Target

	target := stackInput.Target

	// The resource_group field is a StringValueOrRef. The platform middleware resolves
	// valueFrom references before IaC modules run, so .GetValue() always returns the
	// resolved literal string.
	locals.ResourceGroupName = target.Spec.ResourceGroup.GetValue()

	// Metadata-derived tags first, then the user's spec tags merged over
	// them: user tags deliberately win so an org's governance conventions
	// (cost center, owner) can override the derived values where they
	// collide.
	locals.AzureTags = map[string]string{
		azuretagkeys.Resource:     "true",
		azuretagkeys.ResourceName: target.Metadata.Name,
		azuretagkeys.ResourceKind: strings.ToLower(cloudresourcekind.CloudResourceKind_AzureDnsZone.String()),
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

	for key, value := range target.Spec.Tags {
		locals.AzureTags[key] = value
	}

	return locals
}
