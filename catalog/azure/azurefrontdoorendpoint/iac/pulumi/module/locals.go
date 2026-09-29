package module

import (
	"strings"

	azurefrontdoorendpointv1alpha1 "github.com/plantonhq/planton/catalog/azure/azurefrontdoorendpoint/v1alpha1"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/provider/azure/azuretagkeys"
	"github.com/plantonhq/planton/shared/cloudresourcekind"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

type Locals struct {
	AzureFrontDoorEndpoint *azurefrontdoorendpointv1alpha1.AzureFrontDoorEndpoint
	ProfileId              string
	AzureTags              map[string]string
}

func initializeLocals(ctx *pulumi.Context, stackInput *azurefrontdoorendpointv1alpha1.AzureFrontDoorEndpointStackInput) *Locals {
	locals := &Locals{}

	locals.AzureFrontDoorEndpoint = stackInput.Target
	target := stackInput.Target

	locals.ProfileId = target.Spec.ProfileId.GetValue()

	// Metadata-derived tags first, then the user's spec tags merged over
	// them: user tags deliberately win so an org's governance conventions
	// (cost center, owner) can override the derived values where they
	// collide.
	locals.AzureTags = map[string]string{
		azuretagkeys.Resource:     "true",
		azuretagkeys.ResourceName: target.Metadata.Name,
		azuretagkeys.ResourceKind: strings.ToLower(cloudresourcekind.CloudResourceKind_AzureFrontDoorEndpoint.String()),
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
