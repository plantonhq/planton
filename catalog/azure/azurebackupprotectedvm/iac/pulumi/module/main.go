package module

import (
	"github.com/pkg/errors"
	azurebackupprotectedvmv1alpha1 "github.com/plantonhq/planton/catalog/azure/azurebackupprotectedvm/v1alpha1"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/provider/azure/pulumiazureprovider"
	"github.com/pulumi/pulumi-azure/sdk/v6/go/azure/backup"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

func Resources(ctx *pulumi.Context, iacInput *azurebackupprotectedvmv1alpha1.AzureBackupProtectedVmIacInput) error {
	locals := initializeLocals(ctx, iacInput)

	// Build the Azure provider from the IaC input via the shared builder, which resolves
	// the right credential mechanism (static client secret, keyless web identity, or ambient chain).
	azureProvider, err := pulumiazureprovider.Get(ctx, iacInput.ProviderConfig)
	if err != nil {
		return errors.Wrap(err, "failed to create azure provider")
	}

	spec := locals.AzureBackupProtectedVm.Spec

	// Register the VM under the policy's protection. Creation only
	// REGISTERS protection -- the first backup runs on the policy's
	// schedule, not immediately. ARM derives the protected item's own
	// name from the VM's group and name
	// (VM;iaasvmcontainerv2;{vm-rg};{vm-name}).
	//
	// Destroy semantics kept deliberately at the engines' defaults:
	// destroying stops protection AND deletes the backup data (the
	// retain-on-destroy features stay off) -- recorded on the spec.
	protectedVmArgs := &backup.ProtectedVMArgs{
		ResourceGroupName: pulumi.String(locals.ResourceGroupName),
		RecoveryVaultName: pulumi.String(locals.RecoveryVaultName),
		SourceVmId:        pulumi.String(locals.SourceVmId),
		BackupPolicyId:    pulumi.String(locals.BackupPolicyId),
	}

	// Mutually exclusive on the spec (CEL) and the provider
	// (ConflictsWith): the disks to skip OR the disks to keep.
	if len(spec.ExcludeDiskLuns) > 0 {
		luns := pulumi.IntArray{}
		for _, lun := range spec.ExcludeDiskLuns {
			luns = append(luns, pulumi.Int(int(lun)))
		}
		protectedVmArgs.ExcludeDiskLuns = luns
	}
	if len(spec.IncludeDiskLuns) > 0 {
		luns := pulumi.IntArray{}
		for _, lun := range spec.IncludeDiskLuns {
			luns = append(luns, pulumi.Int(int(lun)))
		}
		protectedVmArgs.IncludeDiskLuns = luns
	}

	// Optional+Computed on the provider: unset lets Azure manage the
	// posture (transient states like IRPending read back as Protected).
	// BackupsSuspended additionally requires the VAULT to be immutable
	// -- an apply-time contract Azure checks against the live vault
	// (recorded on the spec field).
	if spec.ProtectionState != "" {
		protectedVmArgs.ProtectionState = pulumi.String(spec.ProtectionState)
	}

	createdProtectedVm, err := backup.NewProtectedVM(ctx,
		locals.AzureBackupProtectedVm.Metadata.Name,
		protectedVmArgs,
		pulumi.Provider(azureProvider))
	if err != nil {
		return errors.Wrap(err, "failed to create backup protected vm")
	}

	ctx.Export(OpBackupProtectedVmId, createdProtectedVm.ID())

	return nil
}
