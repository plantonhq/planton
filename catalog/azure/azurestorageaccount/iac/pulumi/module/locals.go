package module

import (
	"strings"

	azurestorageaccountv1alpha1 "github.com/plantonhq/planton/catalog/azure/azurestorageaccount/v1alpha1"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/provider/azure/azuretagkeys"
	"github.com/plantonhq/planton/shared/catalogkind"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

type Locals struct {
	AzureStorageAccount *azurestorageaccountv1alpha1.AzureStorageAccount
	ResourceGroupName   string
	AzureTags           map[string]string
}

// accountKindStrings maps the spec's account-kind enum to ARM's values.
// Unspecified materializes StorageV2 (the spec's documented default) in
// main.go -- IaC inputs never carry proto defaults.
var accountKindStrings = map[azurestorageaccountv1alpha1.AzureStorageAccountKind]string{
	azurestorageaccountv1alpha1.AzureStorageAccountKind_STORAGE_V2:         "StorageV2",
	azurestorageaccountv1alpha1.AzureStorageAccountKind_BLOB_STORAGE:       "BlobStorage",
	azurestorageaccountv1alpha1.AzureStorageAccountKind_BLOCK_BLOB_STORAGE: "BlockBlobStorage",
	azurestorageaccountv1alpha1.AzureStorageAccountKind_FILE_STORAGE:       "FileStorage",
	azurestorageaccountv1alpha1.AzureStorageAccountKind_STORAGE:            "Storage",
}

// accountTierStrings maps the spec's tier enum to ARM's values.
var accountTierStrings = map[azurestorageaccountv1alpha1.AzureStorageAccountTier]string{
	azurestorageaccountv1alpha1.AzureStorageAccountTier_STANDARD: "Standard",
	azurestorageaccountv1alpha1.AzureStorageAccountTier_PREMIUM:  "Premium",
}

// replicationTypeStrings maps the spec's replication enum to azurerm's SKU
// suffixes (the RA_ prefix collapses: RA_GRS -> RAGRS).
var replicationTypeStrings = map[azurestorageaccountv1alpha1.AzureStorageAccountReplicationType]string{
	azurestorageaccountv1alpha1.AzureStorageAccountReplicationType_LRS:     "LRS",
	azurestorageaccountv1alpha1.AzureStorageAccountReplicationType_ZRS:     "ZRS",
	azurestorageaccountv1alpha1.AzureStorageAccountReplicationType_GRS:     "GRS",
	azurestorageaccountv1alpha1.AzureStorageAccountReplicationType_GZRS:    "GZRS",
	azurestorageaccountv1alpha1.AzureStorageAccountReplicationType_RA_GRS:  "RAGRS",
	azurestorageaccountv1alpha1.AzureStorageAccountReplicationType_RA_GZRS: "RAGZRS",
}

// accessTierStrings maps the spec's access-tier enum to ARM's values.
// Unspecified is not sent at all -- Azure computes Hot on the kinds that
// support tiers, mirroring the Terraform module's null.
var accessTierStrings = map[azurestorageaccountv1alpha1.AzureStorageAccountAccessTier]string{
	azurestorageaccountv1alpha1.AzureStorageAccountAccessTier_HOT:                 "Hot",
	azurestorageaccountv1alpha1.AzureStorageAccountAccessTier_COOL:                "Cool",
	azurestorageaccountv1alpha1.AzureStorageAccountAccessTier_COLD:                "Cold",
	azurestorageaccountv1alpha1.AzureStorageAccountAccessTier_ACCESS_TIER_PREMIUM: "Premium",
}

// minTlsVersionStrings maps the spec's TLS-floor enum; ARM's value happens
// to match the proto value name verbatim. TLS1_2 is the only floor Azure
// still provisions; the retired 1.0/1.1 values no longer exist on the
// spec enum.
var minTlsVersionStrings = map[azurestorageaccountv1alpha1.AzureStorageAccountMinTlsVersion]string{
	azurestorageaccountv1alpha1.AzureStorageAccountMinTlsVersion_TLS1_2: "TLS1_2",
}

// allowedCopyScopeStrings maps the copy-scope restriction enum. Unspecified
// is not sent -- copy stays unrestricted (Azure's default).
var allowedCopyScopeStrings = map[azurestorageaccountv1alpha1.AzureStorageAccountAllowedCopyScope]string{
	azurestorageaccountv1alpha1.AzureStorageAccountAllowedCopyScope_AAD:          "AAD",
	azurestorageaccountv1alpha1.AzureStorageAccountAllowedCopyScope_PRIVATE_LINK: "PrivateLink",
}

// dnsEndpointTypeStrings maps the DNS-architecture enum. Unspecified is not
// sent -- azurerm defaults the create-only choice to Standard itself.
var dnsEndpointTypeStrings = map[azurestorageaccountv1alpha1.AzureStorageAccountDnsEndpointType]string{
	azurestorageaccountv1alpha1.AzureStorageAccountDnsEndpointType_DNS_ENDPOINT_STANDARD: "Standard",
	azurestorageaccountv1alpha1.AzureStorageAccountDnsEndpointType_AZURE_DNS_ZONE:        "AzureDnsZone",
}

// encryptionKeyTypeStrings maps the queue/table key-scope enum. Unspecified
// is not sent -- Azure defaults to the Service scope.
var encryptionKeyTypeStrings = map[azurestorageaccountv1alpha1.AzureStorageAccountEncryptionKeyType]string{
	azurestorageaccountv1alpha1.AzureStorageAccountEncryptionKeyType_SERVICE: "Service",
	azurestorageaccountv1alpha1.AzureStorageAccountEncryptionKeyType_ACCOUNT: "Account",
}

// identityTypeStrings maps the managed-identity flavor enum to ARM's
// comma-separated vocabulary.
var identityTypeStrings = map[azurestorageaccountv1alpha1.AzureStorageAccountIdentityType]string{
	azurestorageaccountv1alpha1.AzureStorageAccountIdentityType_SYSTEM_ASSIGNED:          "SystemAssigned",
	azurestorageaccountv1alpha1.AzureStorageAccountIdentityType_USER_ASSIGNED:            "UserAssigned",
	azurestorageaccountv1alpha1.AzureStorageAccountIdentityType_SYSTEM_AND_USER_ASSIGNED: "SystemAssigned, UserAssigned",
}

// networkDefaultActionStrings maps the firewall default-action enum.
var networkDefaultActionStrings = map[azurestorageaccountv1alpha1.AzureStorageAccountNetworkDefaultAction]string{
	azurestorageaccountv1alpha1.AzureStorageAccountNetworkDefaultAction_ALLOW: "Allow",
	azurestorageaccountv1alpha1.AzureStorageAccountNetworkDefaultAction_DENY:  "Deny",
}

// networkBypassStrings maps the firewall bypass-class enum.
var networkBypassStrings = map[azurestorageaccountv1alpha1.AzureStorageAccountNetworkBypass]string{
	azurestorageaccountv1alpha1.AzureStorageAccountNetworkBypass_AZURE_SERVICES: "AzureServices",
	azurestorageaccountv1alpha1.AzureStorageAccountNetworkBypass_LOGGING:        "Logging",
	azurestorageaccountv1alpha1.AzureStorageAccountNetworkBypass_METRICS:        "Metrics",
	azurestorageaccountv1alpha1.AzureStorageAccountNetworkBypass_NONE:           "None",
}

// routingChoiceStrings maps the routing-preference enum. Unspecified
// materializes MicrosoftRouting (Azure's default) when the block is present.
var routingChoiceStrings = map[azurestorageaccountv1alpha1.AzureStorageAccountRoutingChoice]string{
	azurestorageaccountv1alpha1.AzureStorageAccountRoutingChoice_MICROSOFT_ROUTING: "MicrosoftRouting",
	azurestorageaccountv1alpha1.AzureStorageAccountRoutingChoice_INTERNET_ROUTING:  "InternetRouting",
}

// sasExpirationActionStrings maps the SAS-policy action enum. Unspecified
// materializes Log (Azure's default) when the block is present.
var sasExpirationActionStrings = map[azurestorageaccountv1alpha1.AzureStorageAccountSasExpirationAction]string{
	azurestorageaccountv1alpha1.AzureStorageAccountSasExpirationAction_LOG:   "Log",
	azurestorageaccountv1alpha1.AzureStorageAccountSasExpirationAction_BLOCK: "Block",
}

// immutabilityStateStrings maps the account-level WORM state enum.
var immutabilityStateStrings = map[azurestorageaccountv1alpha1.AzureStorageAccountImmutabilityState]string{
	azurestorageaccountv1alpha1.AzureStorageAccountImmutabilityState_DISABLED: "Disabled",
	azurestorageaccountv1alpha1.AzureStorageAccountImmutabilityState_UNLOCKED: "Unlocked",
	azurestorageaccountv1alpha1.AzureStorageAccountImmutabilityState_LOCKED:   "Locked",
}

// directoryTypeStrings maps the Azure Files directory-service enum; ARM's
// values match the proto value names verbatim.
var directoryTypeStrings = map[azurestorageaccountv1alpha1.AzureStorageAccountDirectoryServiceType]string{
	azurestorageaccountv1alpha1.AzureStorageAccountDirectoryServiceType_AADDS:   "AADDS",
	azurestorageaccountv1alpha1.AzureStorageAccountDirectoryServiceType_AADKERB: "AADKERB",
	azurestorageaccountv1alpha1.AzureStorageAccountDirectoryServiceType_AD:      "AD",
}

// defaultSharePermissionStrings maps the Azure Files default-share-permission
// enum to ARM's role-name vocabulary.
var defaultSharePermissionStrings = map[azurestorageaccountv1alpha1.AzureStorageAccountDefaultSharePermission]string{
	azurestorageaccountv1alpha1.AzureStorageAccountDefaultSharePermission_SHARE_PERMISSION_NONE:                 "None",
	azurestorageaccountv1alpha1.AzureStorageAccountDefaultSharePermission_SHARE_PERMISSION_READER:               "StorageFileDataSmbShareReader",
	azurestorageaccountv1alpha1.AzureStorageAccountDefaultSharePermission_SHARE_PERMISSION_CONTRIBUTOR:          "StorageFileDataSmbShareContributor",
	azurestorageaccountv1alpha1.AzureStorageAccountDefaultSharePermission_SHARE_PERMISSION_ELEVATED_CONTRIBUTOR: "StorageFileDataSmbShareElevatedContributor",
}

// lifecycleBlobTypeStrings maps the lifecycle blob-type enum to ARM's
// camelCase wire values.
var lifecycleBlobTypeStrings = map[azurestorageaccountv1alpha1.AzureStorageAccountLifecycleBlobType]string{
	azurestorageaccountv1alpha1.AzureStorageAccountLifecycleBlobType_BLOCK_BLOB:  "blockBlob",
	azurestorageaccountv1alpha1.AzureStorageAccountLifecycleBlobType_APPEND_BLOB: "appendBlob",
}

func initializeLocals(ctx *pulumi.Context, iacInput *azurestorageaccountv1alpha1.AzureStorageAccountIacInput) *Locals {
	locals := &Locals{}

	locals.AzureStorageAccount = iacInput.Target
	target := iacInput.Target

	locals.ResourceGroupName = target.Spec.ResourceGroup.GetValue()

	locals.AzureTags = map[string]string{
		azuretagkeys.Resource:     "true",
		azuretagkeys.ResourceName: target.Metadata.Name,
		azuretagkeys.ResourceKind: strings.ToLower(catalogkind.CatalogKind_AzureStorageAccount.String()),
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

	// The user's spec tags merge over the metadata-derived tags -- user
	// tags deliberately win so an org's governance conventions can
	// override the derived values where they collide.
	for key, value := range target.Spec.Tags {
		locals.AzureTags[key] = value
	}

	return locals
}
