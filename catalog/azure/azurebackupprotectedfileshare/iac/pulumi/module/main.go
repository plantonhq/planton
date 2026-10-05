package module

import (
	"github.com/pkg/errors"
	azurebackupprotectedfilesharev1alpha1 "github.com/plantonhq/planton/catalog/azure/azurebackupprotectedfileshare/v1alpha1"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/provider/azure/pulumiazureprovider"
	"github.com/pulumi/pulumi-azure/sdk/v6/go/azure/backup"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

func Resources(ctx *pulumi.Context, iacInput *azurebackupprotectedfilesharev1alpha1.AzureBackupProtectedFileShareIacInput) error {
	locals := initializeLocals(ctx, iacInput)

	// Build the Azure provider from the IaC input via the shared builder, which resolves
	// the right credential mechanism (static client secret, keyless web identity, or ambient chain).
	azureProvider, err := pulumiazureprovider.Get(ctx, iacInput.ProviderConfig)
	if err != nil {
		return errors.Wrap(err, "failed to create azure provider")
	}

	// Register the file share under the policy's protection. Creation
	// only REGISTERS protection -- the first backup runs on the
	// policy's schedule, not immediately. The share's storage account
	// must already be REGISTERED with the vault
	// (AzureBackupContainerStorageAccount): the provider runs an
	// Inquire pass to discover protectable shares inside the
	// registered container and fails loudly when the account is not
	// registered. ARM names the protected item by the share's SYSTEM
	// name (AzureFileShare;{system-name}), not its friendly name.
	//
	// Destroy semantics kept deliberately at the engines' defaults:
	// destroying stops protection AND deletes the backup data (vault
	// soft delete may hold it 14 days) -- recorded on the spec.
	protectedFileShareArgs := &backup.ProtectedFileShareArgs{
		ResourceGroupName:      pulumi.String(locals.ResourceGroupName),
		RecoveryVaultName:      pulumi.String(locals.RecoveryVaultName),
		SourceStorageAccountId: pulumi.String(locals.SourceStorageAccountId),
		SourceFileShareName:    pulumi.String(locals.SourceFileShareName),
		// The spec's ONLY updatable field -- re-pointing the policy
		// updates in place; everything else is ForceNew on the
		// provider.
		BackupPolicyId: pulumi.String(locals.BackupPolicyId),
	}

	createdProtectedFileShare, err := backup.NewProtectedFileShare(ctx,
		locals.AzureBackupProtectedFileShare.Metadata.Name,
		protectedFileShareArgs,
		pulumi.Provider(azureProvider))
	if err != nil {
		return errors.Wrap(err, "failed to create backup protected file share")
	}

	ctx.Export(OpBackupProtectedFileShareId, createdProtectedFileShare.ID())

	return nil
}
