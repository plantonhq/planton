package module

import (
	azurebackupprotectedvmv1alpha1 "github.com/plantonhq/planton/catalog/azure/azurebackupprotectedvm/v1alpha1"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

type Locals struct {
	AzureBackupProtectedVm *azurebackupprotectedvmv1alpha1.AzureBackupProtectedVm

	// The four reference fields are StringValueOrRef; the platform
	// middleware resolves valueFrom references before IaC modules run,
	// so GetValue() always returns the resolved literals.
	ResourceGroupName string
	RecoveryVaultName string
	SourceVmId        string
	BackupPolicyId    string
}

// The protected item carries NO tags argument on the provider (ARM
// protected items are untagged), so this module derives no tag map.

func initializeLocals(ctx *pulumi.Context, iacInput *azurebackupprotectedvmv1alpha1.AzureBackupProtectedVmIacInput) *Locals {
	locals := &Locals{}

	locals.AzureBackupProtectedVm = iacInput.Target
	target := iacInput.Target

	locals.ResourceGroupName = target.Spec.ResourceGroup.GetValue()
	locals.RecoveryVaultName = target.Spec.RecoveryVaultName.GetValue()
	locals.SourceVmId = target.Spec.SourceVmId.GetValue()
	locals.BackupPolicyId = target.Spec.BackupPolicyId.GetValue()

	return locals
}
