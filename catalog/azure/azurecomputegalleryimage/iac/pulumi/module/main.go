package module

import (
	"github.com/pkg/errors"
	azurecomputegalleryimagev1alpha1 "github.com/plantonhq/planton/catalog/azure/azurecomputegalleryimage/v1alpha1"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/provider/azure/pulumiazureprovider"
	"github.com/pulumi/pulumi-azure/sdk/v6/go/azure/compute"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

func Resources(ctx *pulumi.Context, iacInput *azurecomputegalleryimagev1alpha1.AzureComputeGalleryImageIacInput) error {
	locals := initializeLocals(ctx, iacInput)

	// Build the Azure provider from the IaC input via the shared builder, which resolves
	// the right credential mechanism (static client secret, keyless web identity, or ambient chain).
	azureProvider, err := pulumiazureprovider.Get(ctx, iacInput.ProviderConfig)
	if err != nil {
		return errors.Wrap(err, "failed to create azure provider")
	}

	spec := locals.AzureComputeGalleryImage.Spec

	createdImage, err := createImage(ctx, locals, spec, azureProvider)
	if err != nil {
		return err
	}

	// Publish the image's versions -- one ARM child per entry, keyed by
	// the version name (its address segment under the image).
	versionIds := pulumi.StringMap{}
	for _, version := range spec.Versions {
		createdVersion, err := createVersion(ctx, locals, spec, version, createdImage, azureProvider)
		if err != nil {
			return errors.Wrapf(err, "failed to create image version %s", version.Name)
		}
		versionIds[version.Name] = createdVersion.ID().ToStringOutput()
	}

	ctx.Export(OpImageId, createdImage.ID())
	ctx.Export(OpImageName, createdImage.Name)
	ctx.Export(OpVersionIds, versionIds)

	return nil
}

// createImage creates the image definition. Almost the whole definition
// is create-only in the provider; the four security flags are a mutual-
// exclusion clique whose ConflictsWith fires on argument PRESENCE --
// each is sent ONLY when true (an explicit false alongside another flag
// is provider-rejected). Unset architecture/hyper_v_generation ride the
// provider defaults (x64 / V1).
func createImage(ctx *pulumi.Context, locals *Locals,
	spec *azurecomputegalleryimagev1alpha1.AzureComputeGalleryImageSpec,
	azureProvider pulumi.ProviderResource) (*compute.SharedImage, error) {

	args := &compute.SharedImageArgs{
		Name:              pulumi.String(spec.Name),
		GalleryName:       pulumi.String(spec.GalleryName.GetValue()),
		ResourceGroupName: pulumi.String(spec.ResourceGroup.GetValue()),
		Location:          pulumi.String(spec.Region),
		OsType:            pulumi.String(spec.OsType),
		Identifier: &compute.SharedImageIdentifierArgs{
			Publisher: pulumi.String(spec.Identifier.Publisher),
			Offer:     pulumi.String(spec.Identifier.Offer),
			Sku:       pulumi.String(spec.Identifier.Sku),
		},
		Tags: pulumi.ToStringMap(locals.AzureTags),
	}

	if spec.Specialized {
		args.Specialized = pulumi.Bool(true)
	}
	if spec.Architecture != "" {
		args.Architecture = pulumi.String(spec.Architecture)
	}
	if spec.HyperVGeneration != "" {
		args.HyperVGeneration = pulumi.String(spec.HyperVGeneration)
	}

	if spec.TrustedLaunchSupported {
		args.TrustedLaunchSupported = pulumi.Bool(true)
	}
	if spec.TrustedLaunchEnabled {
		args.TrustedLaunchEnabled = pulumi.Bool(true)
	}
	if spec.ConfidentialVmSupported {
		args.ConfidentialVmSupported = pulumi.Bool(true)
	}
	if spec.ConfidentialVmEnabled {
		args.ConfidentialVmEnabled = pulumi.Bool(true)
	}

	if spec.AcceleratedNetworkSupportEnabled {
		args.AcceleratedNetworkSupportEnabled = pulumi.Bool(true)
	}
	if spec.HibernationEnabled {
		args.HibernationEnabled = pulumi.Bool(true)
	}
	if spec.DiskControllerTypeNvmeEnabled {
		args.DiskControllerTypeNvmeEnabled = pulumi.Bool(true)
	}

	if len(spec.DiskTypesNotAllowed) > 0 {
		// The classic SDK pluralizes this field name (diskTypesNotAlloweds);
		// same provider argument underneath.
		args.DiskTypesNotAlloweds = pulumi.ToStringArray(spec.DiskTypesNotAllowed)
	}

	// Updatable, but CLEARING a previously set date forces replacement
	// (the provider's CustomizeDiff).
	if spec.EndOfLifeDate != "" {
		args.EndOfLifeDate = pulumi.String(spec.EndOfLifeDate)
	}
	if spec.Eula != "" {
		args.Eula = pulumi.String(spec.Eula)
	}
	if spec.PrivacyStatementUri != "" {
		args.PrivacyStatementUri = pulumi.String(spec.PrivacyStatementUri)
	}
	if spec.ReleaseNoteUri != "" {
		args.ReleaseNoteUri = pulumi.String(spec.ReleaseNoteUri)
	}
	if spec.Description != "" {
		args.Description = pulumi.String(spec.Description)
	}

	if plan := spec.PurchasePlan; plan != nil {
		planArgs := &compute.SharedImagePurchasePlanArgs{
			Name: pulumi.String(plan.Name),
		}
		if plan.Publisher != "" {
			planArgs.Publisher = pulumi.String(plan.Publisher)
		}
		if plan.Product != "" {
			planArgs.Product = pulumi.String(plan.Product)
		}
		args.PurchasePlan = planArgs
	}

	if spec.MinRecommendedVcpuCount != nil {
		args.MinRecommendedVcpuCount = pulumi.Int(int(spec.GetMinRecommendedVcpuCount()))
	}
	if spec.MaxRecommendedVcpuCount != nil {
		args.MaxRecommendedVcpuCount = pulumi.Int(int(spec.GetMaxRecommendedVcpuCount()))
	}
	if spec.MinRecommendedMemoryInGb != nil {
		args.MinRecommendedMemoryInGb = pulumi.Int(int(spec.GetMinRecommendedMemoryInGb()))
	}
	if spec.MaxRecommendedMemoryInGb != nil {
		args.MaxRecommendedMemoryInGb = pulumi.Int(int(spec.GetMaxRecommendedMemoryInGb()))
	}

	createdImage, err := compute.NewSharedImage(ctx,
		locals.AzureComputeGalleryImage.Metadata.Name,
		args,
		pulumi.Provider(azureProvider))
	if err != nil {
		return nil, errors.Wrapf(err, "failed to create gallery image %s",
			locals.AzureComputeGalleryImage.Metadata.Name)
	}
	return createdImage, nil
}

// createVersion publishes one version of the image. Each version has
// exactly one source (the spec's CEL enforces it); target regions and
// exclude-from-latest update in place, everything else is create-only.
// A target region's storage_account_type cannot be UPDATED by the API
// and the provider cannot force replacement for it (region-list
// membership changes in place) -- changing it on an existing region
// surfaces Azure's own error.
func createVersion(ctx *pulumi.Context, locals *Locals,
	spec *azurecomputegalleryimagev1alpha1.AzureComputeGalleryImageSpec,
	version *azurecomputegalleryimagev1alpha1.AzureComputeGalleryImageVersion,
	createdImage *compute.SharedImage,
	azureProvider pulumi.ProviderResource) (*compute.SharedImageVersion, error) {

	targetRegions := compute.SharedImageVersionTargetRegionArray{}
	for _, region := range version.TargetRegions {
		regionArgs := &compute.SharedImageVersionTargetRegionArgs{
			Name:                 pulumi.String(region.Name),
			RegionalReplicaCount: pulumi.Int(int(region.RegionalReplicaCount)),
		}
		if region.DiskEncryptionSetId.GetValue() != "" {
			regionArgs.DiskEncryptionSetId = pulumi.String(region.DiskEncryptionSetId.GetValue())
		}
		if region.ExcludeFromLatestEnabled {
			regionArgs.ExcludeFromLatestEnabled = pulumi.Bool(true)
		}
		if region.StorageAccountType != "" {
			regionArgs.StorageAccountType = pulumi.String(region.StorageAccountType)
		}
		targetRegions = append(targetRegions, regionArgs)
	}

	versionTags := map[string]string{}
	for k, v := range locals.MetadataTags {
		versionTags[k] = v
	}
	for k, v := range version.Tags {
		versionTags[k] = v
	}

	args := &compute.SharedImageVersionArgs{
		Name:              pulumi.String(version.Name),
		GalleryName:       createdImage.GalleryName,
		ImageName:         createdImage.Name,
		ResourceGroupName: pulumi.String(spec.ResourceGroup.GetValue()),
		Location:          pulumi.String(spec.Region),
		TargetRegions:     targetRegions,
		Tags:              pulumi.ToStringMap(versionTags),
	}

	if version.BlobUri != "" {
		args.BlobUri = pulumi.String(version.BlobUri)
	}
	if version.StorageAccountId.GetValue() != "" {
		args.StorageAccountId = pulumi.String(version.StorageAccountId.GetValue())
	}
	if version.OsDiskSnapshotId.GetValue() != "" {
		args.OsDiskSnapshotId = pulumi.String(version.OsDiskSnapshotId.GetValue())
	}
	if version.ManagedImageId.GetValue() != "" {
		args.ManagedImageId = pulumi.String(version.ManagedImageId.GetValue())
	}
	if version.ReplicationMode != "" {
		args.ReplicationMode = pulumi.String(version.ReplicationMode)
	}
	if version.ExcludeFromLatest {
		args.ExcludeFromLatest = pulumi.Bool(true)
	}
	if version.DeletionOfReplicatedLocationsEnabled {
		args.DeletionOfReplicatedLocationsEnabled = pulumi.Bool(true)
	}
	// Updatable, but CLEARING a previously set date forces replacement
	// (the provider's CustomizeDiff).
	if version.EndOfLifeDate != "" {
		args.EndOfLifeDate = pulumi.String(version.EndOfLifeDate)
	}

	return compute.NewSharedImageVersion(ctx,
		locals.AzureComputeGalleryImage.Metadata.Name+"-"+version.Name,
		args,
		pulumi.Provider(azureProvider),
		pulumi.Parent(createdImage))
}
