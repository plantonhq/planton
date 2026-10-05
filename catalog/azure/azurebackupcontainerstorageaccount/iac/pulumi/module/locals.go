package module

import (
	azurebackupcontainerstorageaccountv1alpha1 "github.com/plantonhq/planton/catalog/azure/azurebackupcontainerstorageaccount/v1alpha1"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

type Locals struct {
	AzureBackupContainerStorageAccount *azurebackupcontainerstorageaccountv1alpha1.AzureBackupContainerStorageAccount

	// The three reference fields are StringValueOrRef; the platform
	// middleware resolves valueFrom references before IaC modules run,
	// so GetValue() always returns the resolved literals.
	ResourceGroupName string
	RecoveryVaultName string
	StorageAccountId  string
}

// The protection container carries NO tags argument on the provider
// (ARM protection containers are untagged), so this module derives no
// tag map.

func initializeLocals(ctx *pulumi.Context, iacInput *azurebackupcontainerstorageaccountv1alpha1.AzureBackupContainerStorageAccountIacInput) *Locals {
	locals := &Locals{}

	locals.AzureBackupContainerStorageAccount = iacInput.Target
	target := iacInput.Target

	locals.ResourceGroupName = target.Spec.ResourceGroup.GetValue()
	locals.RecoveryVaultName = target.Spec.RecoveryVaultName.GetValue()
	locals.StorageAccountId = target.Spec.StorageAccountId.GetValue()

	return locals
}
