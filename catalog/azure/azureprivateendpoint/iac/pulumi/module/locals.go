package module

import (
	"strings"

	azureprivateendpointv1alpha1 "github.com/plantonhq/planton/catalog/azure/azureprivateendpoint/v1alpha1"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/provider/azure/azuretagkeys"
	"github.com/plantonhq/planton/shared/catalogkind"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

type Locals struct {
	AzurePrivateEndpoint *azureprivateendpointv1alpha1.AzurePrivateEndpoint
	ResourceGroupName    string
	AzureTags            map[string]string
}

func initializeLocals(ctx *pulumi.Context, iacInput *azureprivateendpointv1alpha1.AzurePrivateEndpointIacInput) *Locals {
	locals := &Locals{}

	locals.AzurePrivateEndpoint = iacInput.Target
	target := iacInput.Target

	locals.ResourceGroupName = target.Spec.ResourceGroup.GetValue()

	// Metadata-derived tags first; the user's spec tags merge over them in
	// main.go so an org's governance conventions win on key collision.
	locals.AzureTags = map[string]string{
		azuretagkeys.Resource:     "true",
		azuretagkeys.ResourceName: target.Metadata.Name,
		azuretagkeys.ResourceKind: strings.ToLower(catalogkind.CatalogKind_AzurePrivateEndpoint.String()),
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
