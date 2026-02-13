package module

import (
	"strings"

	azurestorageaccountv1 "github.com/plantonhq/openmcf/apis/org/openmcf/provider/azure/azurestorageaccount/v1"
	"github.com/plantonhq/openmcf/apis/org/openmcf/shared/cloudresourcekind"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

type Locals struct {
	AzureStorageAccount *azurestorageaccountv1.AzureStorageAccount
	StorageAccountName  string
	ResourceGroupName   string
	AzureTags           map[string]string
}

func initializeLocals(ctx *pulumi.Context, stackInput *azurestorageaccountv1.AzureStorageAccountStackInput) *Locals {
	locals := &Locals{}

	locals.AzureStorageAccount = stackInput.Target
	target := stackInput.Target

	// The resource_group field is a StringValueOrRef. The platform middleware resolves
	// valueFrom references before IaC modules run, so .GetValue() always returns the
	// resolved literal string.
	locals.ResourceGroupName = target.Spec.ResourceGroup.GetValue()

	// Create storage account name from metadata.name
	// Azure Storage Account names must be 3-24 characters, lowercase letters and numbers only, globally unique
	storageAccountName := target.Metadata.Name
	// Remove dots, underscores, and hyphens (Azure storage accounts only allow lowercase alphanumeric)
	storageAccountName = strings.ReplaceAll(storageAccountName, ".", "")
	storageAccountName = strings.ReplaceAll(storageAccountName, "_", "")
	storageAccountName = strings.ReplaceAll(storageAccountName, "-", "")
	storageAccountName = strings.ToLower(storageAccountName)
	// Ensure it's not too long (Azure limit is 24 characters)
	if len(storageAccountName) > 24 {
		storageAccountName = storageAccountName[:24]
	}
	// Ensure minimum length of 3 characters
	if len(storageAccountName) < 3 {
		storageAccountName = storageAccountName + "stg"
	}
	locals.StorageAccountName = storageAccountName

	// Create Azure tags for resource tagging
	locals.AzureTags = map[string]string{
		"resource":      "true",
		"resource_name": target.Metadata.Name,
		"resource_kind": strings.ToLower(cloudresourcekind.CloudResourceKind_AzureStorageAccount.String()),
	}

	if target.Metadata.Id != "" {
		locals.AzureTags["resource_id"] = target.Metadata.Id
	}

	if target.Metadata.Org != "" {
		locals.AzureTags["organization"] = target.Metadata.Org
	}

	if target.Metadata.Env != "" {
		locals.AzureTags["environment"] = target.Metadata.Env
	}

	return locals
}

// getAccountKind converts proto enum to Azure SDK account kind string
func getAccountKind(kind azurestorageaccountv1.AzureStorageAccountKind) string {
	switch kind {
	case azurestorageaccountv1.AzureStorageAccountKind_STORAGE_V2:
		return "StorageV2"
	case azurestorageaccountv1.AzureStorageAccountKind_BLOB_STORAGE:
		return "BlobStorage"
	case azurestorageaccountv1.AzureStorageAccountKind_BLOCK_BLOB_STORAGE:
		return "BlockBlobStorage"
	case azurestorageaccountv1.AzureStorageAccountKind_FILE_STORAGE:
		return "FileStorage"
	case azurestorageaccountv1.AzureStorageAccountKind_STORAGE:
		return "Storage"
	case azurestorageaccountv1.AzureStorageAccountKind_ACCOUNT_KIND_UNSPECIFIED:
		return "StorageV2"
	default:
		return "StorageV2"
	}
}

// getAccountTier converts proto enum to Azure SDK tier string
func getAccountTier(tier azurestorageaccountv1.AzureStorageAccountTier) string {
	switch tier {
	case azurestorageaccountv1.AzureStorageAccountTier_PREMIUM:
		return "Premium"
	case azurestorageaccountv1.AzureStorageAccountTier_STANDARD, azurestorageaccountv1.AzureStorageAccountTier_ACCOUNT_TIER_UNSPECIFIED:
		return "Standard"
	default:
		return "Standard"
	}
}

// getReplicationType converts proto enum to Azure SDK replication type string
func getReplicationType(replication azurestorageaccountv1.AzureStorageReplicationType) string {
	switch replication {
	case azurestorageaccountv1.AzureStorageReplicationType_LRS:
		return "LRS"
	case azurestorageaccountv1.AzureStorageReplicationType_ZRS:
		return "ZRS"
	case azurestorageaccountv1.AzureStorageReplicationType_GRS:
		return "GRS"
	case azurestorageaccountv1.AzureStorageReplicationType_GZRS:
		return "GZRS"
	case azurestorageaccountv1.AzureStorageReplicationType_RA_GRS:
		return "RAGRS"
	case azurestorageaccountv1.AzureStorageReplicationType_RA_GZRS:
		return "RAGZRS"
	case azurestorageaccountv1.AzureStorageReplicationType_REPLICATION_UNSPECIFIED:
		return "LRS"
	default:
		return "LRS"
	}
}

// getAccessTier converts proto enum to Azure SDK access tier string
func getAccessTier(tier azurestorageaccountv1.AzureStorageAccessTier) string {
	switch tier {
	case azurestorageaccountv1.AzureStorageAccessTier_HOT:
		return "Hot"
	case azurestorageaccountv1.AzureStorageAccessTier_COOL:
		return "Cool"
	case azurestorageaccountv1.AzureStorageAccessTier_ACCESS_TIER_UNSPECIFIED:
		return "Hot"
	default:
		return "Hot"
	}
}

// getMinTlsVersion converts proto enum to Azure SDK TLS version string
func getMinTlsVersion(version azurestorageaccountv1.AzureTlsVersion) string {
	switch version {
	case azurestorageaccountv1.AzureTlsVersion_TLS1_0:
		return "TLS1_0"
	case azurestorageaccountv1.AzureTlsVersion_TLS1_1:
		return "TLS1_1"
	case azurestorageaccountv1.AzureTlsVersion_TLS1_2:
		return "TLS1_2"
	case azurestorageaccountv1.AzureTlsVersion_TLS_VERSION_UNSPECIFIED:
		return "TLS1_2"
	default:
		return "TLS1_2"
	}
}

// getNetworkDefaultAction converts proto enum to Azure SDK string
func getNetworkDefaultAction(action azurestorageaccountv1.AzureStorageNetworkAction) string {
	switch action {
	case azurestorageaccountv1.AzureStorageNetworkAction_ALLOW:
		return "Allow"
	case azurestorageaccountv1.AzureStorageNetworkAction_DENY, azurestorageaccountv1.AzureStorageNetworkAction_NETWORK_ACTION_UNSPECIFIED:
		return "Deny"
	default:
		return "Deny"
	}
}

// getContainerAccessType converts proto enum to Azure SDK container access type
func getContainerAccessType(access azurestorageaccountv1.AzureStorageContainerAccess) string {
	switch access {
	case azurestorageaccountv1.AzureStorageContainerAccess_PRIVATE:
		return "private"
	case azurestorageaccountv1.AzureStorageContainerAccess_BLOB:
		return "blob"
	case azurestorageaccountv1.AzureStorageContainerAccess_CONTAINER:
		return "container"
	case azurestorageaccountv1.AzureStorageContainerAccess_CONTAINER_ACCESS_UNSPECIFIED:
		return "private"
	default:
		return "private"
	}
}
