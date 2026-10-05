package module

import (
	"strings"

	azuremanageddiskv1alpha1 "github.com/plantonhq/planton/catalog/azure/azuremanageddisk/v1alpha1"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/provider/azure/azuretagkeys"
	"github.com/plantonhq/planton/shared/catalogkind"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

type Locals struct {
	AzureManagedDisk *azuremanageddiskv1alpha1.AzureManagedDisk

	// ResourceGroupName is a StringValueOrRef field; the platform middleware
	// resolves valueFrom references before IaC modules run, so GetValue()
	// always returns the resolved literal name.
	ResourceGroupName string

	// StorageAccountType and CreateOption are the ARM strings for the
	// spec's required enums.
	StorageAccountType string
	CreateOption       string

	// OsType, HyperVGeneration, SecurityType, and NetworkAccessPolicy are
	// the ARM strings for the spec's optional enums, or empty when
	// unspecified so both engines send nothing and Azure's defaults apply.
	OsType              string
	HyperVGeneration    string
	SecurityType        string
	NetworkAccessPolicy string

	// AzureTags is the metadata-derived tag map with the spec's user tags
	// merged over it (user tags win on key collision), mirroring the
	// Terraform module's merge order.
	AzureTags map[string]string
}

func initializeLocals(ctx *pulumi.Context, iacInput *azuremanageddiskv1alpha1.AzureManagedDiskIacInput) *Locals {
	locals := &Locals{}

	locals.AzureManagedDisk = iacInput.Target
	target := iacInput.Target

	locals.ResourceGroupName = target.Spec.ResourceGroup.GetValue()

	switch target.Spec.StorageAccountType {
	case azuremanageddiskv1alpha1.AzureManagedDiskStorageAccountType_STANDARD_LRS:
		locals.StorageAccountType = "Standard_LRS"
	case azuremanageddiskv1alpha1.AzureManagedDiskStorageAccountType_STANDARD_SSD_LRS:
		locals.StorageAccountType = "StandardSSD_LRS"
	case azuremanageddiskv1alpha1.AzureManagedDiskStorageAccountType_STANDARD_SSD_ZRS:
		locals.StorageAccountType = "StandardSSD_ZRS"
	case azuremanageddiskv1alpha1.AzureManagedDiskStorageAccountType_PREMIUM_LRS:
		locals.StorageAccountType = "Premium_LRS"
	case azuremanageddiskv1alpha1.AzureManagedDiskStorageAccountType_PREMIUM_ZRS:
		locals.StorageAccountType = "Premium_ZRS"
	case azuremanageddiskv1alpha1.AzureManagedDiskStorageAccountType_PREMIUM_V2_LRS:
		locals.StorageAccountType = "PremiumV2_LRS"
	case azuremanageddiskv1alpha1.AzureManagedDiskStorageAccountType_ULTRA_SSD_LRS:
		locals.StorageAccountType = "UltraSSD_LRS"
	}

	switch target.Spec.CreateOption {
	case azuremanageddiskv1alpha1.AzureManagedDiskCreateOption_EMPTY:
		locals.CreateOption = "Empty"
	case azuremanageddiskv1alpha1.AzureManagedDiskCreateOption_COPY:
		locals.CreateOption = "Copy"
	case azuremanageddiskv1alpha1.AzureManagedDiskCreateOption_FROM_IMAGE:
		locals.CreateOption = "FromImage"
	case azuremanageddiskv1alpha1.AzureManagedDiskCreateOption_IMPORT:
		locals.CreateOption = "Import"
	case azuremanageddiskv1alpha1.AzureManagedDiskCreateOption_IMPORT_SECURE:
		locals.CreateOption = "ImportSecure"
	case azuremanageddiskv1alpha1.AzureManagedDiskCreateOption_RESTORE:
		locals.CreateOption = "Restore"
	case azuremanageddiskv1alpha1.AzureManagedDiskCreateOption_UPLOAD:
		locals.CreateOption = "Upload"
	}

	switch target.Spec.OsType {
	case azuremanageddiskv1alpha1.AzureManagedDiskOsType_LINUX:
		locals.OsType = "Linux"
	case azuremanageddiskv1alpha1.AzureManagedDiskOsType_WINDOWS:
		locals.OsType = "Windows"
	}

	switch target.Spec.HyperVGeneration {
	case azuremanageddiskv1alpha1.AzureManagedDiskHyperVGeneration_V1:
		locals.HyperVGeneration = "V1"
	case azuremanageddiskv1alpha1.AzureManagedDiskHyperVGeneration_V2:
		locals.HyperVGeneration = "V2"
	}

	switch target.Spec.SecurityType {
	case azuremanageddiskv1alpha1.AzureManagedDiskSecurityType_CONFIDENTIAL_VM_VMGUEST_STATE_ONLY_ENCRYPTED_WITH_PLATFORM_KEY:
		locals.SecurityType = "ConfidentialVM_VMGuestStateOnlyEncryptedWithPlatformKey"
	case azuremanageddiskv1alpha1.AzureManagedDiskSecurityType_CONFIDENTIAL_VM_DISK_ENCRYPTED_WITH_PLATFORM_KEY:
		locals.SecurityType = "ConfidentialVM_DiskEncryptedWithPlatformKey"
	case azuremanageddiskv1alpha1.AzureManagedDiskSecurityType_CONFIDENTIAL_VM_DISK_ENCRYPTED_WITH_CUSTOMER_KEY:
		locals.SecurityType = "ConfidentialVM_DiskEncryptedWithCustomerKey"
	}

	switch target.Spec.NetworkAccessPolicy {
	case azuremanageddiskv1alpha1.AzureManagedDiskNetworkAccessPolicy_ALLOW_ALL:
		locals.NetworkAccessPolicy = "AllowAll"
	case azuremanageddiskv1alpha1.AzureManagedDiskNetworkAccessPolicy_ALLOW_PRIVATE:
		locals.NetworkAccessPolicy = "AllowPrivate"
	case azuremanageddiskv1alpha1.AzureManagedDiskNetworkAccessPolicy_DENY_ALL:
		locals.NetworkAccessPolicy = "DenyAll"
	}

	// Metadata-derived tags first, then the user's spec tags merged over
	// them: user tags deliberately win so an org's governance conventions
	// (cost center, owner) can override the derived values where they
	// collide.
	locals.AzureTags = map[string]string{
		azuretagkeys.Resource:     "true",
		azuretagkeys.ResourceName: target.Metadata.Name,
		azuretagkeys.ResourceKind: strings.ToLower(catalogkind.CatalogKind_AzureManagedDisk.String()),
	}

	if target.Metadata.Id != "" {
		locals.AzureTags[azuretagkeys.ResourceId] = target.Metadata.Id
	}

	if target.Metadata.Org != "" {
		locals.AzureTags[azuretagkeys.Organization] = target.Metadata.Org
	}

	if target.Metadata.Env != "" {
		locals.AzureTags[azuretagkeys.Environment] = target.Metadata.Env
	}

	for k, v := range target.Spec.Tags {
		locals.AzureTags[k] = v
	}

	return locals
}
