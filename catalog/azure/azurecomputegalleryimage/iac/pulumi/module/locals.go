package module

import (
	"strings"

	azurecomputegalleryimagev1alpha1 "github.com/plantonhq/planton/catalog/azure/azurecomputegalleryimage/v1alpha1"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/provider/azure/azuretagkeys"
	"github.com/plantonhq/planton/shared/cloudresourcekind"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

type Locals struct {
	AzureComputeGalleryImage *azurecomputegalleryimagev1alpha1.AzureComputeGalleryImage

	// AzureTags is the metadata-derived tag map with the spec's user tags
	// merged over it (user tags win on key collision), mirroring the
	// Terraform module's merge order.
	AzureTags map[string]string

	// MetadataTags is the metadata-derived tag map WITHOUT the image's
	// spec tags -- versions merge their OWN spec tags over this base
	// instead of inheriting the image's.
	MetadataTags map[string]string
}

func initializeLocals(ctx *pulumi.Context, stackInput *azurecomputegalleryimagev1alpha1.AzureComputeGalleryImageStackInput) *Locals {
	locals := &Locals{}

	locals.AzureComputeGalleryImage = stackInput.Target
	target := stackInput.Target

	// Metadata-derived tags first, then the user's spec tags merged over
	// them: user tags deliberately win so an org's governance conventions
	// (cost center, owner) can override the derived values where they
	// collide.
	locals.MetadataTags = map[string]string{
		azuretagkeys.Resource:     "true",
		azuretagkeys.ResourceName: target.Metadata.Name,
		azuretagkeys.ResourceKind: strings.ToLower(cloudresourcekind.CloudResourceKind_AzureComputeGalleryImage.String()),
	}

	if target.Metadata.Id != "" {
		locals.MetadataTags[azuretagkeys.ResourceId] = target.Metadata.Id
	}

	if target.Metadata.Org != "" {
		locals.MetadataTags[azuretagkeys.Organization] = target.Metadata.Org
	}

	if target.Metadata.Env != "" {
		locals.MetadataTags[azuretagkeys.Environment] = target.Metadata.Env
	}

	locals.AzureTags = map[string]string{}
	for k, v := range locals.MetadataTags {
		locals.AzureTags[k] = v
	}
	for k, v := range target.Spec.Tags {
		locals.AzureTags[k] = v
	}

	return locals
}
