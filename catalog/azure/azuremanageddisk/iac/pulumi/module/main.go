package module

import (
	"github.com/pkg/errors"
	azuremanageddiskv1alpha1 "github.com/plantonhq/planton/catalog/azure/azuremanageddisk/v1alpha1"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/provider/azure/pulumiazureprovider"
	"github.com/pulumi/pulumi-azure/sdk/v6/go/azure/compute"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

func Resources(ctx *pulumi.Context, iacInput *azuremanageddiskv1alpha1.AzureManagedDiskIacInput) error {
	locals := initializeLocals(ctx, iacInput)

	// Build the Azure provider from the IaC input via the shared builder, which resolves
	// the right credential mechanism (static client secret, keyless web identity, or ambient chain).
	azureProvider, err := pulumiazureprovider.Get(ctx, iacInput.ProviderConfig)
	if err != nil {
		return errors.Wrap(err, "failed to create azure provider")
	}

	spec := locals.AzureManagedDisk.Spec

	// Lifecycle notes worth knowing before operating this resource:
	// - Name, region, zone, create option (and its source fields),
	//   logical sector size, security profile, and performance-plus are
	//   the disk's identity -- changing any of them replaces the disk
	//   AND ITS DATA.
	// - disk_size_gb can only INCREASE; growing an attached disk may
	//   briefly detach it or deallocate the VM except where Azure
	//   supports live resize.
	// - The VM-side attachment is NOT here: AzureVirtualMachine's
	//   data_disk_attachments owns which VM mounts this disk, at which
	//   LUN, with which caching -- so the disk survives VM replacement
	//   untouched.
	managedDiskArgs := &compute.ManagedDiskArgs{
		Name:               pulumi.String(spec.Name),
		Location:           pulumi.String(spec.Region),
		ResourceGroupName:  pulumi.String(locals.ResourceGroupName),
		StorageAccountType: pulumi.String(locals.StorageAccountType),
		CreateOption:       pulumi.String(locals.CreateOption),
		Tags:               pulumi.ToStringMap(locals.AzureTags),
	}

	if spec.DiskSizeGb != nil {
		managedDiskArgs.DiskSizeGb = pulumi.Int(int(spec.GetDiskSizeGb()))
	}

	// The create option's source fields; spec-level validation enforces
	// the same option-to-source pairings ARM does, so absent fields are
	// simply not sent.
	if spec.SourceResourceId != "" {
		managedDiskArgs.SourceResourceId = pulumi.String(spec.SourceResourceId)
	}
	if spec.SourceUri != "" {
		managedDiskArgs.SourceUri = pulumi.String(spec.SourceUri)
	}
	if spec.StorageAccountId != "" {
		managedDiskArgs.StorageAccountId = pulumi.String(spec.StorageAccountId)
	}
	if spec.ImageReferenceId != "" {
		managedDiskArgs.ImageReferenceId = pulumi.String(spec.ImageReferenceId)
	}
	if spec.GalleryImageReferenceId != "" {
		managedDiskArgs.GalleryImageReferenceId = pulumi.String(spec.GalleryImageReferenceId)
	}
	if spec.UploadSizeBytes != 0 {
		managedDiskArgs.UploadSizeBytes = pulumi.Int(int(spec.UploadSizeBytes))
	}

	if locals.OsType != "" {
		managedDiskArgs.OsType = pulumi.String(locals.OsType)
	}
	if locals.HyperVGeneration != "" {
		managedDiskArgs.HyperVGeneration = pulumi.String(locals.HyperVGeneration)
	}
	if spec.Zone != "" {
		managedDiskArgs.Zone = pulumi.String(spec.Zone)
	}

	// Independent performance dials -- PremiumV2/Ultra only (the
	// read-only pair budgets a shared disk's read-only mounts).
	if spec.DiskIopsReadWrite != nil {
		managedDiskArgs.DiskIopsReadWrite = pulumi.Int(int(spec.GetDiskIopsReadWrite()))
	}
	if spec.DiskMbpsReadWrite != nil {
		managedDiskArgs.DiskMbpsReadWrite = pulumi.Int(int(spec.GetDiskMbpsReadWrite()))
	}
	if spec.DiskIopsReadOnly != nil {
		managedDiskArgs.DiskIopsReadOnly = pulumi.Int(int(spec.GetDiskIopsReadOnly()))
	}
	if spec.DiskMbpsReadOnly != nil {
		managedDiskArgs.DiskMbpsReadOnly = pulumi.Int(int(spec.GetDiskMbpsReadOnly()))
	}

	// Premium SSD tier decoupling (unset = the size's default tier) and
	// on-demand bursting for >512 GiB premium disks.
	if spec.Tier != "" {
		managedDiskArgs.Tier = pulumi.String(spec.Tier)
	}
	managedDiskArgs.OnDemandBurstingEnabled = pulumi.Bool(spec.OnDemandBurstingEnabled)

	// The shared-disk seam: >1 lets several VMs attach simultaneously.
	if spec.MaxShares != nil {
		managedDiskArgs.MaxShares = pulumi.Int(int(spec.GetMaxShares()))
	}
	if spec.LogicalSectorSize != nil {
		managedDiskArgs.LogicalSectorSize = pulumi.Int(int(spec.GetLogicalSectorSize()))
	}

	// Encryption: customer-managed keys via a disk encryption set, or the
	// confidential-VM customer-key variant (mutually exclusive; spec-level
	// validation enforces the pairing with security_type).
	if spec.DiskEncryptionSetId.GetValue() != "" {
		managedDiskArgs.DiskEncryptionSetId = pulumi.String(spec.DiskEncryptionSetId.GetValue())
	}
	if spec.SecureVmDiskEncryptionSetId.GetValue() != "" {
		managedDiskArgs.SecureVmDiskEncryptionSetId = pulumi.String(spec.SecureVmDiskEncryptionSetId.GetValue())
	}
	if locals.SecurityType != "" {
		managedDiskArgs.SecurityType = pulumi.String(locals.SecurityType)
	}
	if spec.TrustedLaunchEnabled {
		managedDiskArgs.TrustedLaunchEnabled = pulumi.Bool(true)
	}

	// Network export posture: who can reach the disk's export endpoint.
	if locals.NetworkAccessPolicy != "" {
		managedDiskArgs.NetworkAccessPolicy = pulumi.String(locals.NetworkAccessPolicy)
	}
	if spec.DiskAccessId != "" {
		managedDiskArgs.DiskAccessId = pulumi.String(spec.DiskAccessId)
	}
	// Presence-guarded true-default optional bool: an absent spec value
	// explicitly falls back to Azure's default so iac-input paths that
	// bypass the manifest loader deploy identically to the Terraform
	// module's optional(bool, true) default.
	if spec.PublicNetworkAccessEnabled != nil {
		managedDiskArgs.PublicNetworkAccessEnabled = pulumi.Bool(spec.GetPublicNetworkAccessEnabled())
	} else {
		managedDiskArgs.PublicNetworkAccessEnabled = pulumi.Bool(true)
	}

	managedDiskArgs.OptimizedFrequentAttachEnabled = pulumi.Bool(spec.OptimizedFrequentAttachEnabled)
	managedDiskArgs.PerformancePlusEnabled = pulumi.Bool(spec.PerformancePlusEnabled)

	if spec.EdgeZone != "" {
		managedDiskArgs.EdgeZone = pulumi.String(spec.EdgeZone)
	}

	createdManagedDisk, err := compute.NewManagedDisk(ctx,
		spec.Name,
		managedDiskArgs,
		pulumi.Provider(azureProvider))
	if err != nil {
		return errors.Wrapf(err, "failed to create managed disk %s", spec.Name)
	}

	// Export outputs from the created resource.
	ctx.Export(OpDiskId, createdManagedDisk.ID())
	ctx.Export(OpDiskName, createdManagedDisk.Name)
	ctx.Export(OpDiskSizeGb, createdManagedDisk.DiskSizeGb)

	return nil
}
