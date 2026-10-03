package module

import (
	azurestoragesharev1alpha1 "github.com/plantonhq/planton/catalog/azure/azurestorageshare/v1alpha1"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

type Locals struct {
	AzureStorageShare *azurestoragesharev1alpha1.AzureStorageShare
	StorageAccountId  string
}

// accessTierStrings maps the spec's tier enum to azurerm's wire values.
// Unspecified is never sent -- Azure's per-account-kind default
// (TransactionOptimized on standard, Premium on FileStorage) applies.
var accessTierStrings = map[azurestoragesharev1alpha1.AzureStorageShareAccessTier]string{
	azurestoragesharev1alpha1.AzureStorageShareAccessTier_TRANSACTION_OPTIMIZED: "TransactionOptimized",
	azurestoragesharev1alpha1.AzureStorageShareAccessTier_HOT:                   "Hot",
	azurestoragesharev1alpha1.AzureStorageShareAccessTier_COOL:                  "Cool",
	azurestoragesharev1alpha1.AzureStorageShareAccessTier_PREMIUM:               "Premium",
}

// enabledProtocolStrings maps the spec's protocol enum to azurerm's wire
// values. Unspecified materializes SMB in main.go -- azurerm's own default.
var enabledProtocolStrings = map[azurestoragesharev1alpha1.AzureStorageShareProtocol]string{
	azurestoragesharev1alpha1.AzureStorageShareProtocol_SMB: "SMB",
	azurestoragesharev1alpha1.AzureStorageShareProtocol_NFS: "NFS",
}

func initializeLocals(ctx *pulumi.Context, iacInput *azurestoragesharev1alpha1.AzureStorageShareIacInput) *Locals {
	locals := &Locals{}

	locals.AzureStorageShare = iacInput.Target
	locals.StorageAccountId = iacInput.Target.Spec.StorageAccountId.GetValue()

	// No Azure tags: ARM does not support tags on fileServices/shares,
	// so the platform's identity tags live on the parent account.

	return locals
}
