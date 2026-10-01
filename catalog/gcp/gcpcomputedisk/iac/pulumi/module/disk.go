package module

import (
	"strings"

	"github.com/pkg/errors"
	"github.com/pulumi/pulumi-gcp/sdk/v9/go/gcp"
	"github.com/pulumi/pulumi-gcp/sdk/v9/go/gcp/compute"
	"github.com/pulumi/pulumi-gcp/sdk/v9/go/gcp/projects"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// computeDisk enables the Compute Engine API and creates the persistent
// disk.
//
// Sharp edges, all taught by the API rather than invented here:
//
//   - name, zone, type, sources, encryption, and architecture are
//     immutable — changing them replaces the disk and its data. size
//     grows in place but never shrinks.
//
//   - At most one source (image / snapshot / instant snapshot / storage
//     object / source_disk); none creates an empty disk — the common case
//     for data volumes.
//
//   - Deleting a disk still attached to a running instance fails; detach
//     first (or delete the instance).
//
//   - create_snapshot_before_destroy takes a final snapshot during
//     destroy — a last-resort recovery net for precious volumes (CMEK
//     disks reuse their key for the snapshot).
func computeDisk(ctx *pulumi.Context, locals *Locals, gcpProvider *gcp.Provider) (*compute.Disk, error) {
	spec := locals.GcpComputeDisk.Spec

	// Enable the Compute Engine API — the control plane that owns disks.
	// DisableOnDestroy stays false: tearing down one disk must never
	// disable the API for everything else in the project.
	computeApiArgs := &projects.ServiceArgs{
		Service:                  pulumi.String("compute.googleapis.com"),
		DisableDependentServices: pulumi.BoolPtr(true),
		DisableOnDestroy:         pulumi.BoolPtr(false),
	}
	// Honor the spec contract: an empty project_id falls back to the
	// provider's default project.
	if spec.ProjectId.GetValue() != "" {
		computeApiArgs.Project = pulumi.String(spec.ProjectId.GetValue())
	}
	createdComputeApi, err := projects.NewService(ctx,
		"gcpdisk-compute.googleapis.com", computeApiArgs, pulumi.Provider(gcpProvider))
	if err != nil {
		return nil, errors.Wrap(err, "failed to enable compute.googleapis.com api")
	}

	args := &compute.DiskArgs{
		Name:   pulumi.String(locals.DiskName),
		Zone:   pulumi.StringPtr(spec.Zone),
		Labels: pulumi.ToStringMap(locals.GcpLabels),
	}

	if spec.ProjectId.GetValue() != "" {
		args.Project = pulumi.StringPtr(spec.ProjectId.GetValue())
	}
	if spec.Description != "" {
		args.Description = pulumi.StringPtr(spec.Description)
	}
	if spec.Type != "" {
		args.Type = pulumi.StringPtr(spec.Type)
	}
	if spec.SizeGb > 0 {
		args.Size = pulumi.IntPtr(int(spec.SizeGb))
	}
	if spec.Image.GetValue() != "" {
		args.Image = pulumi.StringPtr(spec.Image.GetValue())
	}
	if spec.SourceSnapshot != "" {
		args.Snapshot = pulumi.StringPtr(spec.SourceSnapshot)
	}
	if spec.SourceInstantSnapshot != "" {
		args.SourceInstantSnapshot = pulumi.StringPtr(spec.SourceInstantSnapshot)
	}
	if spec.SourceStorageObject != "" {
		args.SourceStorageObject = pulumi.StringPtr(spec.SourceStorageObject)
	}
	if spec.SourceDisk.GetValue() != "" {
		args.SourceDisk = pulumi.StringPtr(spec.SourceDisk.GetValue())
	}
	if spec.AccessMode != "" {
		args.AccessMode = pulumi.StringPtr(spec.AccessMode)
	}
	if spec.Architecture != "" {
		args.Architecture = pulumi.StringPtr(spec.Architecture)
	}
	if spec.StoragePool != "" {
		args.StoragePool = pulumi.StringPtr(spec.StoragePool)
	}
	if spec.ProvisionedIops != nil {
		args.ProvisionedIops = pulumi.IntPtr(int(spec.GetProvisionedIops()))
	}
	if spec.ProvisionedThroughput != nil {
		args.ProvisionedThroughput = pulumi.IntPtr(int(spec.GetProvisionedThroughput()))
	}
	if spec.EnableConfidentialCompute {
		args.EnableConfidentialCompute = pulumi.BoolPtr(true)
	}
	if spec.PhysicalBlockSizeBytes != nil {
		args.PhysicalBlockSizeBytes = pulumi.IntPtr(int(spec.GetPhysicalBlockSizeBytes()))
	}
	if spec.CreateSnapshotBeforeDestroy {
		args.CreateSnapshotBeforeDestroy = pulumi.BoolPtr(true)
	}
	if spec.SnapshotBeforeDestroyPrefix != "" {
		args.CreateSnapshotBeforeDestroyPrefix = pulumi.StringPtr(spec.SnapshotBeforeDestroyPrefix)
	}

	// CMEK: the Compute Engine service agent must hold
	// roles/cloudkms.cryptoKeyEncrypterDecrypter on this key before
	// create. KmsKeyServiceAccount overrides which identity performs the
	// encryption request. Raw CSEK keys are deliberately not supported —
	// key material never flows through manifests or state.
	if spec.KmsKey.GetValue() != "" {
		diskEncryptionKeyArgs := &compute.DiskDiskEncryptionKeyArgs{
			KmsKeySelfLink: pulumi.StringPtr(spec.KmsKey.GetValue()),
		}
		if spec.KmsKeyServiceAccount != "" {
			diskEncryptionKeyArgs.KmsKeyServiceAccount = pulumi.StringPtr(spec.KmsKeyServiceAccount)
		}
		args.DiskEncryptionKey = diskEncryptionKeyArgs
	}

	// CMEK decryption of an encrypted source image / snapshot.
	if spec.SourceImageEncryption != nil {
		imageKeyArgs := &compute.DiskSourceImageEncryptionKeyArgs{
			KmsKeySelfLink: pulumi.StringPtr(spec.SourceImageEncryption.KmsKey.GetValue()),
		}
		if spec.SourceImageEncryption.KmsKeyServiceAccount != "" {
			imageKeyArgs.KmsKeyServiceAccount = pulumi.StringPtr(spec.SourceImageEncryption.KmsKeyServiceAccount)
		}
		args.SourceImageEncryptionKey = imageKeyArgs
	}
	if spec.SourceSnapshotEncryption != nil {
		snapshotKeyArgs := &compute.DiskSourceSnapshotEncryptionKeyArgs{
			KmsKeySelfLink: pulumi.StringPtr(spec.SourceSnapshotEncryption.KmsKey.GetValue()),
		}
		if spec.SourceSnapshotEncryption.KmsKeyServiceAccount != "" {
			snapshotKeyArgs.KmsKeyServiceAccount = pulumi.StringPtr(spec.SourceSnapshotEncryption.KmsKeyServiceAccount)
		}
		args.SourceSnapshotEncryptionKey = snapshotKeyArgs
	}

	// Guest OS features for bootable disks; licenses for BYOL imports.
	// Both create-time only.
	if len(spec.GuestOsFeatures) > 0 {
		guestOsFeatures := compute.DiskGuestOsFeatureArray{}
		for _, featureType := range spec.GuestOsFeatures {
			guestOsFeatures = append(guestOsFeatures, &compute.DiskGuestOsFeatureArgs{
				Type: pulumi.String(featureType),
			})
		}
		args.GuestOsFeatures = guestOsFeatures
	}
	if len(spec.Licenses) > 0 {
		args.Licenses = pulumi.ToStringArray(spec.Licenses)
	}

	// Destroy-time guard: PREVENT fails the destroy; ABANDON unmanages the
	// disk without deleting it. Unset falls back to the provider default
	// (DELETE).
	if spec.DeletionPolicy != "" {
		args.DeletionPolicy = pulumi.StringPtr(spec.DeletionPolicy)
	}

	// Async replication: pairing this disk (the SECONDARY) to its primary.
	// Replication starts when the pair is activated on the primary.
	if spec.AsyncPrimaryDisk.GetValue() != "" {
		args.AsyncPrimaryDisk = &compute.DiskAsyncPrimaryDiskArgs{
			Disk: pulumi.String(spec.AsyncPrimaryDisk.GetValue()),
		}
	}

	// Resource Manager tags bind at create time only.
	if len(spec.ResourceManagerTags) > 0 {
		args.Params = &compute.DiskParamsArgs{
			ResourceManagerTags: pulumi.ToStringMap(spec.ResourceManagerTags),
		}
	}

	createdDisk, err := compute.NewDisk(ctx,
		locals.DiskName,
		args,
		pulumi.Provider(gcpProvider),
		pulumi.DependsOn([]pulumi.Resource{createdComputeApi}),
	)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create compute disk")
	}

	// Semantic outputs — names and shapes byte-identical to the Terraform
	// module's outputs. zone exports the plain spec value; type is
	// normalized to the last path segment on BOTH engines because
	// provider lines differ on bare-name vs full-path attribute formats.
	ctx.Export(OpName, createdDisk.Name)
	ctx.Export(OpDiskId, createdDisk.DiskId)
	ctx.Export(OpSelfLink, createdDisk.SelfLink)
	ctx.Export(OpZone, pulumi.String(spec.Zone))
	ctx.Export(OpSizeGb, createdDisk.Size)
	ctx.Export(OpType, createdDisk.Type.ApplyT(func(t *string) string {
		if t == nil {
			return ""
		}
		parts := strings.Split(*t, "/")
		return parts[len(parts)-1]
	}).(pulumi.StringOutput))

	return createdDisk, nil
}
