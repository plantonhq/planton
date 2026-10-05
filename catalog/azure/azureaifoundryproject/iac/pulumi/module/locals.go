package module

import (
	"strings"

	azureaifoundryprojectv1alpha1 "github.com/plantonhq/planton/catalog/azure/azureaifoundryproject/v1alpha1"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/provider/azure/azuretagkeys"
	"github.com/plantonhq/planton/shared/catalogkind"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

type Locals struct {
	AzureAiFoundryProject *azureaifoundryprojectv1alpha1.AzureAiFoundryProject

	// AzureTags is the metadata-derived tag map with the spec's user tags
	// merged over it (user tags win on key collision), mirroring the
	// Terraform module's merge order.
	AzureTags map[string]string
}

// identityTypeWire maps the spec's identity flavors to the provider's
// comma-joined wire values.
var identityTypeWire = map[azureaifoundryprojectv1alpha1.AzureAiFoundryProjectIdentityType]string{
	azureaifoundryprojectv1alpha1.AzureAiFoundryProjectIdentityType_SYSTEM_ASSIGNED:          "SystemAssigned",
	azureaifoundryprojectv1alpha1.AzureAiFoundryProjectIdentityType_USER_ASSIGNED:            "UserAssigned",
	azureaifoundryprojectv1alpha1.AzureAiFoundryProjectIdentityType_SYSTEM_AND_USER_ASSIGNED: "SystemAssigned, UserAssigned",
}

func initializeLocals(ctx *pulumi.Context, iacInput *azureaifoundryprojectv1alpha1.AzureAiFoundryProjectIacInput) *Locals {
	locals := &Locals{}

	locals.AzureAiFoundryProject = iacInput.Target
	target := iacInput.Target

	// Metadata-derived tags first, then the user's spec tags merged over
	// them: user tags deliberately win so an org's governance conventions
	// (cost center, owner) can override the derived values where they
	// collide.
	locals.AzureTags = map[string]string{
		azuretagkeys.Resource:     "true",
		azuretagkeys.ResourceName: target.Metadata.Name,
		azuretagkeys.ResourceKind: strings.ToLower(catalogkind.CatalogKind_AzureAiFoundryProject.String()),
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
