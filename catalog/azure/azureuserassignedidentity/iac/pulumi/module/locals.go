package module

import (
	"strings"

	azureuserassignedidentityv1alpha1 "github.com/plantonhq/planton/catalog/azure/azureuserassignedidentity/v1alpha1"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/provider/azure/azuretagkeys"
	"github.com/plantonhq/planton/shared/catalogkind"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

type Locals struct {
	AzureUserAssignedIdentity *azureuserassignedidentityv1alpha1.AzureUserAssignedIdentity

	// ResourceGroupName is a StringValueOrRef field; the platform middleware
	// resolves valueFrom references before IaC modules run, so GetValue()
	// always returns the resolved literal.
	ResourceGroupName string

	// IsolationScope is the ARM string for the spec's enum ("Regional"), or
	// empty when unspecified so the provider applies ARM's default (no
	// isolation).
	IsolationScope string

	// AzureTags is the metadata-derived tag map with the spec's user tags
	// merged over it (user tags win on key collision), mirroring the
	// Terraform module's merge order.
	AzureTags map[string]string
}

func initializeLocals(ctx *pulumi.Context, iacInput *azureuserassignedidentityv1alpha1.AzureUserAssignedIdentityIacInput) *Locals {
	locals := &Locals{}

	locals.AzureUserAssignedIdentity = iacInput.Target
	target := iacInput.Target

	locals.ResourceGroupName = target.Spec.ResourceGroup.GetValue()

	if target.Spec.IsolationScope == azureuserassignedidentityv1alpha1.AzureUserAssignedIdentityIsolationScope_REGIONAL {
		locals.IsolationScope = "Regional"
	}

	// Metadata-derived tags first, then the user's spec tags merged over
	// them: user tags deliberately win so an org's governance conventions
	// (cost center, owner) can override the derived values where they
	// collide.
	locals.AzureTags = map[string]string{
		azuretagkeys.Resource:     "true",
		azuretagkeys.ResourceName: target.Metadata.Name,
		azuretagkeys.ResourceKind: strings.ToLower(catalogkind.CatalogKind_AzureUserAssignedIdentity.String()),
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
