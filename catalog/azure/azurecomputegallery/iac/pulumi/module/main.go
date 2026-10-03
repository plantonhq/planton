package module

import (
	"github.com/pkg/errors"
	azurecomputegalleryv1alpha1 "github.com/plantonhq/planton/catalog/azure/azurecomputegallery/v1alpha1"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/provider/azure/pulumiazureprovider"
	"github.com/pulumi/pulumi-azure/sdk/v6/go/azure/compute"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

func Resources(ctx *pulumi.Context, iacInput *azurecomputegalleryv1alpha1.AzureComputeGalleryIacInput) error {
	locals := initializeLocals(ctx, iacInput)

	// Build the Azure provider from the IaC input via the shared builder, which resolves
	// the right credential mechanism (static client secret, keyless web identity, or ambient chain).
	azureProvider, err := pulumiazureprovider.Get(ctx, iacInput.ProviderConfig)
	if err != nil {
		return errors.Wrap(err, "failed to create azure provider")
	}

	spec := locals.AzureComputeGallery.Spec

	// Create the gallery. The ENTIRE sharing tree is create-only in the
	// provider (changing it forces replacement); Community sharing
	// requires the community_gallery block (the spec's CEL front-loads
	// the provider's expand-time check). A gallery is free at rest.
	args := &compute.SharedImageGalleryArgs{
		Name:              pulumi.String(spec.Name),
		ResourceGroupName: pulumi.String(spec.ResourceGroup.GetValue()),
		Location:          pulumi.String(spec.Region),
		Tags:              pulumi.ToStringMap(locals.AzureTags),
	}

	if spec.Description != "" {
		args.Description = pulumi.String(spec.Description)
	}

	if sharing := spec.Sharing; sharing != nil {
		sharingArgs := &compute.SharedImageGallerySharingArgs{
			Permission: pulumi.String(sharing.Permission),
		}
		if community := sharing.CommunityGallery; community != nil {
			sharingArgs.CommunityGallery = &compute.SharedImageGallerySharingCommunityGalleryArgs{
				Eula:           pulumi.String(community.Eula),
				Prefix:         pulumi.String(community.Prefix),
				PublisherEmail: pulumi.String(community.PublisherEmail),
				PublisherUri:   pulumi.String(community.PublisherUri),
			}
		}
		args.Sharing = sharingArgs
	}

	createdGallery, err := compute.NewSharedImageGallery(ctx,
		locals.AzureComputeGallery.Metadata.Name,
		args,
		pulumi.Provider(azureProvider))
	if err != nil {
		return errors.Wrapf(err, "failed to create compute gallery %s",
			locals.AzureComputeGallery.Metadata.Name)
	}

	ctx.Export(OpGalleryId, createdGallery.ID())
	ctx.Export(OpGalleryName, createdGallery.Name)
	ctx.Export(OpUniqueName, createdGallery.UniqueName)
	// The community public name lives on the sharing tree; empty unless
	// Community-shared.
	ctx.Export(OpCommunityGalleryName, createdGallery.Sharing.ApplyT(func(sharing *compute.SharedImageGallerySharing) string {
		if sharing == nil || sharing.CommunityGallery == nil || sharing.CommunityGallery.Name == nil {
			return ""
		}
		return *sharing.CommunityGallery.Name
	}).(pulumi.StringOutput))

	return nil
}
