package module

import (
	azurebackupprotectedfilesharev1alpha1 "github.com/plantonhq/planton/catalog/azure/azurebackupprotectedfileshare/v1alpha1"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

type Locals struct {
	AzureBackupProtectedFileShare *azurebackupprotectedfilesharev1alpha1.AzureBackupProtectedFileShare

	// The five reference fields are StringValueOrRef; the platform
	// middleware resolves valueFrom references before IaC modules run,
	// so GetValue() always returns the resolved literals.
	ResourceGroupName      string
	RecoveryVaultName      string
	SourceStorageAccountId string
	SourceFileShareName    string
	BackupPolicyId         string
}

// The protected item carries NO tags argument on the provider (ARM
// protected items are untagged), so this module derives no tag map.

func initializeLocals(ctx *pulumi.Context, iacInput *azurebackupprotectedfilesharev1alpha1.AzureBackupProtectedFileShareIacInput) *Locals {
	locals := &Locals{}

	locals.AzureBackupProtectedFileShare = iacInput.Target
	target := iacInput.Target

	locals.ResourceGroupName = target.Spec.ResourceGroup.GetValue()
	locals.RecoveryVaultName = target.Spec.RecoveryVaultName.GetValue()
	locals.SourceStorageAccountId = target.Spec.SourceStorageAccountId.GetValue()
	locals.SourceFileShareName = target.Spec.SourceFileShareName.GetValue()
	locals.BackupPolicyId = target.Spec.BackupPolicyId.GetValue()

	return locals
}
