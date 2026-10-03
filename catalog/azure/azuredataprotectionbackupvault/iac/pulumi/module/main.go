package module

import (
	"github.com/pkg/errors"
	azuredataprotectionbackupvaultv1alpha1 "github.com/plantonhq/planton/catalog/azure/azuredataprotectionbackupvault/v1alpha1"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/provider/azure/pulumiazureprovider"
	"github.com/pulumi/pulumi-azure/sdk/v6/go/azure/dataprotection"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

func Resources(ctx *pulumi.Context, iacInput *azuredataprotectionbackupvaultv1alpha1.AzureDataProtectionBackupVaultIacInput) error {
	locals := initializeLocals(ctx, iacInput)

	// Build the Azure provider from the IaC input via the shared builder, which resolves
	// the right credential mechanism (static client secret, keyless web identity, or ambient chain).
	azureProvider, err := pulumiazureprovider.Get(ctx, iacInput.ProviderConfig)
	if err != nil {
		return errors.Wrap(err, "failed to create azure provider")
	}

	spec := locals.AzureDataProtectionBackupVault.Spec

	// Create the Data Protection backup vault -- the safe that modern
	// Azure Backup data (disks, blobs, AKS clusters, flexible-server
	// databases, Data Lake storage) lives in. The vault is free at
	// rest; cost follows the protected instances and their backup
	// storage.
	//
	// Destroy note: Azure's delete call returns before the vault is
	// fully gone; the provider polls until the name is actually free
	// (its own workaround for the service bug), so destroy runs a
	// little longer than the API suggests.
	vaultArgs := &dataprotection.BackupVaultArgs{
		Name:              pulumi.String(spec.Name),
		ResourceGroupName: pulumi.String(locals.ResourceGroupName),
		Location:          pulumi.String(spec.Region),
		// Both ForceNew on the provider; required in the spec.
		DatastoreType: pulumi.String(spec.DatastoreType),
		Redundancy:    pulumi.String(spec.Redundancy),
		Tags:          pulumi.ToStringMap(locals.AzureTags),
	}

	// Sent ONLY when true: the provider errors when this argument is
	// EXPLICITLY present (even as false) on a non-GeoRedundant vault,
	// so an unset spec value must never reach the provider as false.
	// Enabling is in-place; DISABLING replaces the vault (the
	// provider's one-way ForceNew, recorded on the spec field).
	if spec.CrossRegionRestoreEnabled {
		vaultArgs.CrossRegionRestoreEnabled = pulumi.Bool(true)
	}

	// Soft-delete retention window (days). Unset lets the provider
	// default (14) apply.
	if spec.RetentionDurationInDays != nil {
		vaultArgs.RetentionDurationInDays = pulumi.Float64(float64(*spec.RetentionDurationInDays))
	}

	// Unset lets the provider defaults apply (On / Disabled). Both
	// carry one-way doors -- AlwaysOn and Locked are permanent
	// (leaving either replaces the vault; the provider's ForceNew
	// transitions, recorded on the spec fields).
	if spec.SoftDelete != nil {
		vaultArgs.SoftDelete = pulumi.String(*spec.SoftDelete)
	}
	if spec.Immutability != nil {
		vaultArgs.Immutability = pulumi.String(*spec.Immutability)
	}

	if spec.Identity != nil {
		identityArgs := &dataprotection.BackupVaultIdentityArgs{
			Type: pulumi.String(identityTypeWire[spec.Identity.Type]),
		}
		if len(spec.Identity.IdentityIds) > 0 {
			identityIds := pulumi.StringArray{}
			for _, identityId := range spec.Identity.IdentityIds {
				identityIds = append(identityIds, pulumi.String(identityId.GetValue()))
			}
			identityArgs.IdentityIds = identityIds
		}
		vaultArgs.Identity = identityArgs
	}

	createdVault, err := dataprotection.NewBackupVault(ctx,
		spec.Name,
		vaultArgs,
		pulumi.Provider(azureProvider))
	if err != nil {
		return errors.Wrapf(err, "failed to create data protection backup vault %s", spec.Name)
	}

	ctx.Export(OpBackupVaultId, createdVault.ID())
	ctx.Export(OpBackupVaultName, createdVault.Name)
	ctx.Export(OpSystemAssignedIdentityPrincipalId, createdVault.Identity.PrincipalId())

	// Customer-managed-key encryption, composed from the provider's
	// sibling resource (which rewrites the vault's own security
	// settings -- one ARM object, one authoritative shape). Azure
	// unwraps the key with the vault's SYSTEM-assigned identity (the
	// provider hardcodes it; the spec's CEL requires that identity
	// flavor).
	//
	// ONE-WAY DOOR: once enabled, CMK can never be removed -- the
	// provider's delete for this resource is a documented no-op; only
	// deleting the vault removes the encryption. The KEY itself
	// rotates in place (the one updatable part), and versionless key
	// URIs (the reference's default) make rotation automatic.
	if spec.Encryption != nil {
		_, err := dataprotection.NewBackupVaultCustomerManagedKey(ctx,
			spec.Name+"-cmk",
			&dataprotection.BackupVaultCustomerManagedKeyArgs{
				DataProtectionBackupVaultId: createdVault.ID(),
				KeyVaultKeyId:               pulumi.String(spec.Encryption.KeyId.GetValue()),
			},
			pulumi.Provider(azureProvider),
			pulumi.Parent(createdVault))
		if err != nil {
			return errors.Wrap(err, "failed to create backup vault customer managed key")
		}
	}

	return nil
}
