package module

import (
	"github.com/pkg/errors"
	azuredisksnapshotv1alpha1 "github.com/plantonhq/planton/catalog/azure/azuredisksnapshot/v1alpha1"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/provider/azure/pulumiazureprovider"
	"github.com/pulumi/pulumi-azure/sdk/v6/go/azure/compute"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

func Resources(ctx *pulumi.Context, iacInput *azuredisksnapshotv1alpha1.AzureDiskSnapshotIacInput) error {
	locals := initializeLocals(ctx, iacInput)

	// Build the Azure provider from the IaC input via the shared builder, which resolves
	// the right credential mechanism (static client secret, keyless web identity, or ambient chain).
	azureProvider, err := pulumiazureprovider.Get(ctx, iacInput.ProviderConfig)
	if err != nil {
		return errors.Wrap(err, "failed to create azure provider")
	}

	spec := locals.AzureDiskSnapshot.Spec

	// Create the snapshot. The source fields pair with create_option
	// ("Copy" reads source_resource_id; "Import" reads source_uri +
	// storage_account_id) -- the provider's own schema does not tie
	// them together and Azure validates the pairing at create time, so
	// the module sends each source field only when set. Removing
	// encryption settings from a snapshot that had them forces
	// replacement (Azure cannot disable encryption in place).
	args := &compute.SnapshotArgs{
		Name:              pulumi.String(spec.Name),
		ResourceGroupName: pulumi.String(spec.ResourceGroup.GetValue()),
		Location:          pulumi.String(spec.Region),
		CreateOption:      pulumi.String(spec.CreateOption),
		Tags:              pulumi.ToStringMap(locals.AzureTags),
	}

	if spec.SourceResourceId.GetValue() != "" {
		args.SourceResourceId = pulumi.String(spec.SourceResourceId.GetValue())
	}
	if spec.SourceUri != "" {
		args.SourceUri = pulumi.String(spec.SourceUri)
	}
	if spec.StorageAccountId.GetValue() != "" {
		args.StorageAccountId = pulumi.String(spec.StorageAccountId.GetValue())
	}
	if spec.IncrementalEnabled {
		args.IncrementalEnabled = pulumi.Bool(true)
	}
	// Unset inherits the source's size (Azure computes it).
	if spec.DiskSizeGb != nil {
		args.DiskSizeGb = pulumi.Int(int(spec.GetDiskSizeGb()))
	}
	// Unset rides the provider default, "AllowAll".
	if spec.NetworkAccessPolicy != "" {
		args.NetworkAccessPolicy = pulumi.String(spec.NetworkAccessPolicy)
	}
	if spec.DiskAccessId.GetValue() != "" {
		args.DiskAccessId = pulumi.String(spec.DiskAccessId.GetValue())
	}
	// Unset rides the provider default, true.
	if spec.PublicNetworkAccessEnabled != nil {
		args.PublicNetworkAccessEnabled = pulumi.Bool(spec.GetPublicNetworkAccessEnabled())
	}

	if encryption := spec.EncryptionSettings; encryption != nil {
		encryptionArgs := &compute.SnapshotEncryptionSettingsArgs{
			DiskEncryptionKey: &compute.SnapshotEncryptionSettingsDiskEncryptionKeyArgs{
				SecretUrl:     pulumi.String(encryption.DiskEncryptionKey.SecretUrl),
				SourceVaultId: pulumi.String(encryption.DiskEncryptionKey.SourceVaultId.GetValue()),
			},
		}
		if kek := encryption.KeyEncryptionKey; kek != nil {
			encryptionArgs.KeyEncryptionKey = &compute.SnapshotEncryptionSettingsKeyEncryptionKeyArgs{
				KeyUrl:        pulumi.String(kek.KeyUrl),
				SourceVaultId: pulumi.String(kek.SourceVaultId.GetValue()),
			}
		}
		args.EncryptionSettings = encryptionArgs
	}

	// The source fields are create-time-only by contract: a snapshot's
	// creation data is immutable history, and the provider (v5 pin)
	// never reads source_resource_id/source_uri back from Azure -- an
	// adopted (imported) snapshot holds a null source in state, and
	// without this guard every post-import plan proposes a
	// destroy+create that would delete the very artifact the user
	// adopted. Ignoring the pair also means an in-place source edit is
	// a no-op rather than a silent deletion of a backup artifact;
	// capturing a different disk is a NEW snapshot resource (the spec
	// field comments teach this). The Terraform module carries the same
	// guard -- keep the engines in step.
	createdSnapshot, err := compute.NewSnapshot(ctx,
		locals.AzureDiskSnapshot.Metadata.Name,
		args,
		pulumi.Provider(azureProvider),
		pulumi.IgnoreChanges([]string{"sourceResourceId", "sourceUri"}))
	if err != nil {
		return errors.Wrapf(err, "failed to create snapshot %s",
			locals.AzureDiskSnapshot.Metadata.Name)
	}

	ctx.Export(OpSnapshotId, createdSnapshot.ID())
	ctx.Export(OpSnapshotName, createdSnapshot.Name)

	return nil
}
