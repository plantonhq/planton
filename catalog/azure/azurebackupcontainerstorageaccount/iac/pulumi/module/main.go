package module

import (
	"github.com/pkg/errors"
	azurebackupcontainerstorageaccountv1alpha1 "github.com/plantonhq/planton/catalog/azure/azurebackupcontainerstorageaccount/v1alpha1"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/provider/azure/pulumiazureprovider"
	"github.com/pulumi/pulumi-azure/sdk/v6/go/azure/backup"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

func Resources(ctx *pulumi.Context, iacInput *azurebackupcontainerstorageaccountv1alpha1.AzureBackupContainerStorageAccountIacInput) error {
	locals := initializeLocals(ctx, iacInput)

	// Build the Azure provider from the IaC input via the shared builder, which resolves
	// the right credential mechanism (static client secret, keyless web identity, or ambient chain).
	azureProvider, err := pulumiazureprovider.Get(ctx, iacInput.ProviderConfig)
	if err != nil {
		return errors.Wrap(err, "failed to create azure provider")
	}

	// Register the storage account with the vault as a backup
	// container (.../vaults/{vault}/backupFabrics/Azure/
	// protectionContainers/StorageContainer;storage;{sa-rg};{sa-name})
	// -- the prerequisite for protecting any of the account's file
	// shares. Registration is free and moves no data; every argument
	// is ForceNew (ARM has no update on protection containers).
	//
	// While registered, Azure Backup places a resource lock on the
	// storage account; destroying this resource unregisters and
	// removes the lock -- and REFUSES while any of the account's
	// shares are still protected (recorded on the spec).
	containerArgs := &backup.ContainerStorageAccountArgs{
		ResourceGroupName: pulumi.String(locals.ResourceGroupName),
		RecoveryVaultName: pulumi.String(locals.RecoveryVaultName),
		StorageAccountId:  pulumi.String(locals.StorageAccountId),
	}

	createdContainer, err := backup.NewContainerStorageAccount(ctx,
		locals.AzureBackupContainerStorageAccount.Metadata.Name,
		containerArgs,
		pulumi.Provider(azureProvider))
	if err != nil {
		return errors.Wrap(err, "failed to create backup container storage account registration")
	}

	ctx.Export(OpBackupContainerId, createdContainer.ID())
	// Echo the registered account's ARM ID: protected file shares
	// reference THIS output for their source_storage_account_id so the
	// registration deploys first -- the reference carries both the
	// value and the dependency edge.
	ctx.Export(OpStorageAccountId, createdContainer.StorageAccountId)

	return nil
}
