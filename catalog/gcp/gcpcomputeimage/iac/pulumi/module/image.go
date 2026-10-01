package module

import (
	"github.com/pkg/errors"
	gcpcomputeimagev1alpha1 "github.com/plantonhq/planton/catalog/gcp/gcpcomputeimage/v1alpha1"
	"github.com/pulumi/pulumi-gcp/sdk/v9/go/gcp"
	"github.com/pulumi/pulumi-gcp/sdk/v9/go/gcp/compute"
	"github.com/pulumi/pulumi-gcp/sdk/v9/go/gcp/projects"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// image creates the custom image from its one source. Everything except
// labels is ForceNew, which is why images are versioned by name and
// consumed through their family.
func image(ctx *pulumi.Context, locals *Locals, gcpProvider *gcp.Provider) error {
	spec := locals.GcpComputeImage.Spec
	project := spec.GetProjectId().GetValue()

	// Enable the Compute Engine API. DisableOnDestroy stays false: tearing
	// down one image must never disable the API for everything else in the
	// project.
	apiArgs := &projects.ServiceArgs{
		Service:                  pulumi.String("compute.googleapis.com"),
		DisableDependentServices: pulumi.BoolPtr(true),
		DisableOnDestroy:         pulumi.BoolPtr(false),
	}
	if project != "" {
		apiArgs.Project = pulumi.String(project)
	}
	createdComputeApi, err := projects.NewService(ctx, "gcpimg-compute.googleapis.com", apiArgs, pulumi.Provider(gcpProvider))
	if err != nil {
		return errors.Wrap(err, "failed to enable compute.googleapis.com")
	}

	args := &compute.ImageArgs{
		Name:   pulumi.StringPtr(locals.ImageName),
		Labels: pulumi.ToStringMap(locals.GcpLabels),
	}
	if project != "" {
		args.Project = pulumi.StringPtr(project)
	}
	if spec.Description != "" {
		args.Description = pulumi.StringPtr(spec.Description)
	}
	if spec.Family != "" {
		args.Family = pulumi.StringPtr(spec.Family)
	}
	if spec.DeletionPolicy != "" {
		args.DeletionPolicy = pulumi.StringPtr(spec.DeletionPolicy)
	}

	// Exactly one source (the spec enforces it).
	if sourceDisk := spec.GetSourceDisk().GetValue(); sourceDisk != "" {
		args.SourceDisk = pulumi.StringPtr(sourceDisk)
	}
	if sourceImage := spec.GetSourceImage().GetValue(); sourceImage != "" {
		args.SourceImage = pulumi.StringPtr(sourceImage)
	}
	if spec.SourceSnapshot != "" {
		args.SourceSnapshot = pulumi.StringPtr(spec.SourceSnapshot)
	}
	if rawDisk := spec.RawDisk; rawDisk != nil {
		rawDiskArgs := &compute.ImageRawDiskArgs{Source: pulumi.String(rawDisk.Source)}
		if rawDisk.Sha1 != "" {
			rawDiskArgs.Sha1 = pulumi.StringPtr(rawDisk.Sha1)
		}
		if rawDisk.ContainerType != "" {
			rawDiskArgs.ContainerType = pulumi.StringPtr(rawDisk.ContainerType)
		}
		args.RawDisk = rawDiskArgs
	}

	// Optional+Computed: sent only when declared, so the values Google
	// inherits from the source never show as a diff.
	if spec.DiskSizeGb != nil {
		args.DiskSizeGb = pulumi.IntPtr(int(spec.GetDiskSizeGb()))
	}
	if len(spec.GuestOsFeatures) > 0 {
		features := compute.ImageGuestOsFeatureArray{}
		for _, feature := range spec.GuestOsFeatures {
			features = append(features, &compute.ImageGuestOsFeatureArgs{Type: pulumi.String(feature)})
		}
		args.GuestOsFeatures = features
	}
	if len(spec.Licenses) > 0 {
		args.Licenses = pulumi.ToStringArray(spec.Licenses)
	}
	if len(spec.StorageLocations) > 0 {
		args.StorageLocations = pulumi.ToStringArray(spec.StorageLocations)
	}

	if kmsKey := spec.GetKmsKey().GetValue(); kmsKey != "" {
		keyArgs := &compute.ImageImageEncryptionKeyArgs{KmsKeySelfLink: pulumi.StringPtr(kmsKey)}
		if spec.KmsKeyServiceAccount != "" {
			keyArgs.KmsKeyServiceAccount = pulumi.StringPtr(spec.KmsKeyServiceAccount)
		}
		args.ImageEncryptionKey = keyArgs
	}
	if source := spec.SourceDiskEncryption; source != nil {
		keyArgs := &compute.ImageSourceDiskEncryptionKeyArgs{KmsKeySelfLink: pulumi.StringPtr(source.GetKmsKey().GetValue())}
		if source.KmsKeyServiceAccount != "" {
			keyArgs.KmsKeyServiceAccount = pulumi.StringPtr(source.KmsKeyServiceAccount)
		}
		args.SourceDiskEncryptionKey = keyArgs
	}
	if source := spec.SourceImageEncryption; source != nil {
		keyArgs := &compute.ImageSourceImageEncryptionKeyArgs{KmsKeySelfLink: pulumi.StringPtr(source.GetKmsKey().GetValue())}
		if source.KmsKeyServiceAccount != "" {
			keyArgs.KmsKeyServiceAccount = pulumi.StringPtr(source.KmsKeyServiceAccount)
		}
		args.SourceImageEncryptionKey = keyArgs
	}
	if source := spec.SourceSnapshotEncryption; source != nil {
		keyArgs := &compute.ImageSourceSnapshotEncryptionKeyArgs{KmsKeySelfLink: pulumi.StringPtr(source.GetKmsKey().GetValue())}
		if source.KmsKeyServiceAccount != "" {
			keyArgs.KmsKeyServiceAccount = pulumi.StringPtr(source.KmsKeyServiceAccount)
		}
		args.SourceSnapshotEncryptionKey = keyArgs
	}

	if state := spec.ShieldedInstanceInitialState; state != nil {
		args.ShieldedInstanceInitialState = shieldedInstanceInitialState(state)
	}
	if len(spec.ResourceManagerTags) > 0 {
		args.Params = &compute.ImageParamsArgs{ResourceManagerTags: pulumi.ToStringMap(spec.ResourceManagerTags)}
	}

	created, err := compute.NewImage(ctx, locals.GcpComputeImage.Metadata.Name, args,
		pulumi.Provider(gcpProvider), pulumi.DependsOn([]pulumi.Resource{createdComputeApi}))
	if err != nil {
		return errors.Wrap(err, "failed to create compute image")
	}

	ctx.Export(OpName, created.Name)
	ctx.Export(OpSelfLink, created.SelfLink)
	ctx.Export(OpFamily, created.Family.ApplyT(func(family *string) string {
		if family == nil {
			return ""
		}
		return *family
	}).(pulumi.StringOutput))
	ctx.Export(OpDiskSizeGb, created.DiskSizeGb)
	// The provider's resource ID is projects/{project}/global/images/{name},
	// the same string the Terraform module exports from the resource's id.
	ctx.Export(OpImageId, created.ID().ToStringOutput())
	return nil
}

// shieldedInstanceInitialState maps the Secure Boot key databases, each
// entry's file type sent only when declared.
func shieldedInstanceInitialState(state *gcpcomputeimagev1alpha1.GcpComputeImageShieldedInstanceInitialState) *compute.ImageShieldedInstanceInitialStateArgs {
	fileType := func(value string) pulumi.StringPtrInput {
		if value == "" {
			return nil
		}
		return pulumi.StringPtr(value)
	}
	args := &compute.ImageShieldedInstanceInitialStateArgs{}
	if pk := state.Pk; pk != nil {
		args.Pk = &compute.ImageShieldedInstanceInitialStatePkArgs{Content: pulumi.String(pk.Content), FileType: fileType(pk.FileType)}
	}
	if len(state.Keks) > 0 {
		keks := compute.ImageShieldedInstanceInitialStateKekArray{}
		for _, kek := range state.Keks {
			keks = append(keks, &compute.ImageShieldedInstanceInitialStateKekArgs{Content: pulumi.String(kek.Content), FileType: fileType(kek.FileType)})
		}
		args.Keks = keks
	}
	if len(state.Dbs) > 0 {
		dbs := compute.ImageShieldedInstanceInitialStateDbArray{}
		for _, db := range state.Dbs {
			dbs = append(dbs, &compute.ImageShieldedInstanceInitialStateDbArgs{Content: pulumi.String(db.Content), FileType: fileType(db.FileType)})
		}
		args.Dbs = dbs
	}
	if len(state.Dbxs) > 0 {
		dbxs := compute.ImageShieldedInstanceInitialStateDbxArray{}
		for _, dbx := range state.Dbxs {
			dbxs = append(dbxs, &compute.ImageShieldedInstanceInitialStateDbxArgs{Content: pulumi.String(dbx.Content), FileType: fileType(dbx.FileType)})
		}
		args.Dbxs = dbxs
	}
	return args
}
