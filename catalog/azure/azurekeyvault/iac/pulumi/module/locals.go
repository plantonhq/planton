package module

import (
	"strings"

	azurekeyvaultv1alpha1 "github.com/plantonhq/planton/catalog/azure/azurekeyvault/v1alpha1"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/provider/azure/azuretagkeys"
	"github.com/plantonhq/planton/shared/cloudresourcekind"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

type Locals struct {
	AzureKeyVault *azurekeyvaultv1alpha1.AzureKeyVault

	// ResourceGroupName is a StringValueOrRef field; the platform middleware
	// resolves valueFrom references before IaC modules run, so GetValue()
	// always returns the resolved literal.
	ResourceGroupName string

	// SkuName is the ARM string for the spec's SKU enum. An unspecified spec
	// applies the STANDARD baseline on both engines (azurerm requires an
	// explicit sku_name, so the default is materialized here rather than
	// sent as empty).
	SkuName string

	// AzureTags is the metadata-derived tag map with the spec's user tags
	// merged over it (user tags win on key collision), mirroring the
	// Terraform module's merge order.
	AzureTags map[string]string
}

func initializeLocals(ctx *pulumi.Context, stackInput *azurekeyvaultv1alpha1.AzureKeyVaultStackInput) *Locals {
	locals := &Locals{}

	locals.AzureKeyVault = stackInput.Target
	target := stackInput.Target

	locals.ResourceGroupName = target.Spec.ResourceGroup.GetValue()

	if target.Spec.Sku == azurekeyvaultv1alpha1.AzureKeyVaultSku_PREMIUM {
		locals.SkuName = "premium"
	} else {
		locals.SkuName = "standard"
	}

	// Metadata-derived tags first, then the user's spec tags merged over
	// them: user tags deliberately win so an org's governance conventions
	// (cost center, owner) can override the derived values where they
	// collide.
	locals.AzureTags = map[string]string{
		azuretagkeys.Resource:     "true",
		azuretagkeys.ResourceName: target.Metadata.Name,
		azuretagkeys.ResourceKind: strings.ToLower(cloudresourcekind.CloudResourceKind_AzureKeyVault.String()),
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

// The four permission vocabularies below translate the spec's enum values to
// the exact data-plane strings Azure's API expects (which is case-sensitive
// about all of them). A missing entry would silently drop a grant, so each
// map is exhaustive over its enum by construction -- the Terraform module
// carries the same maps in locals.tf.

var keyPermissionStrings = map[azurekeyvaultv1alpha1.AzureKeyVaultKeyPermission]string{
	azurekeyvaultv1alpha1.AzureKeyVaultKeyPermission_KEY_GET:                 "Get",
	azurekeyvaultv1alpha1.AzureKeyVaultKeyPermission_KEY_LIST:                "List",
	azurekeyvaultv1alpha1.AzureKeyVaultKeyPermission_KEY_UPDATE:              "Update",
	azurekeyvaultv1alpha1.AzureKeyVaultKeyPermission_KEY_CREATE:              "Create",
	azurekeyvaultv1alpha1.AzureKeyVaultKeyPermission_KEY_IMPORT:              "Import",
	azurekeyvaultv1alpha1.AzureKeyVaultKeyPermission_KEY_DELETE:              "Delete",
	azurekeyvaultv1alpha1.AzureKeyVaultKeyPermission_KEY_RECOVER:             "Recover",
	azurekeyvaultv1alpha1.AzureKeyVaultKeyPermission_KEY_BACKUP:              "Backup",
	azurekeyvaultv1alpha1.AzureKeyVaultKeyPermission_KEY_RESTORE:             "Restore",
	azurekeyvaultv1alpha1.AzureKeyVaultKeyPermission_KEY_DECRYPT:             "Decrypt",
	azurekeyvaultv1alpha1.AzureKeyVaultKeyPermission_KEY_ENCRYPT:             "Encrypt",
	azurekeyvaultv1alpha1.AzureKeyVaultKeyPermission_KEY_UNWRAP_KEY:          "UnwrapKey",
	azurekeyvaultv1alpha1.AzureKeyVaultKeyPermission_KEY_WRAP_KEY:            "WrapKey",
	azurekeyvaultv1alpha1.AzureKeyVaultKeyPermission_KEY_VERIFY:              "Verify",
	azurekeyvaultv1alpha1.AzureKeyVaultKeyPermission_KEY_SIGN:                "Sign",
	azurekeyvaultv1alpha1.AzureKeyVaultKeyPermission_KEY_PURGE:               "Purge",
	azurekeyvaultv1alpha1.AzureKeyVaultKeyPermission_KEY_RELEASE:             "Release",
	azurekeyvaultv1alpha1.AzureKeyVaultKeyPermission_KEY_ROTATE:              "Rotate",
	azurekeyvaultv1alpha1.AzureKeyVaultKeyPermission_KEY_GET_ROTATION_POLICY: "GetRotationPolicy",
	azurekeyvaultv1alpha1.AzureKeyVaultKeyPermission_KEY_SET_ROTATION_POLICY: "SetRotationPolicy",
}

var secretPermissionStrings = map[azurekeyvaultv1alpha1.AzureKeyVaultSecretPermission]string{
	azurekeyvaultv1alpha1.AzureKeyVaultSecretPermission_SECRET_GET:     "Get",
	azurekeyvaultv1alpha1.AzureKeyVaultSecretPermission_SECRET_LIST:    "List",
	azurekeyvaultv1alpha1.AzureKeyVaultSecretPermission_SECRET_SET:     "Set",
	azurekeyvaultv1alpha1.AzureKeyVaultSecretPermission_SECRET_DELETE:  "Delete",
	azurekeyvaultv1alpha1.AzureKeyVaultSecretPermission_SECRET_RECOVER: "Recover",
	azurekeyvaultv1alpha1.AzureKeyVaultSecretPermission_SECRET_BACKUP:  "Backup",
	azurekeyvaultv1alpha1.AzureKeyVaultSecretPermission_SECRET_RESTORE: "Restore",
	azurekeyvaultv1alpha1.AzureKeyVaultSecretPermission_SECRET_PURGE:   "Purge",
}

var certificatePermissionStrings = map[azurekeyvaultv1alpha1.AzureKeyVaultCertificatePermission]string{
	azurekeyvaultv1alpha1.AzureKeyVaultCertificatePermission_CERTIFICATE_GET:             "Get",
	azurekeyvaultv1alpha1.AzureKeyVaultCertificatePermission_CERTIFICATE_LIST:            "List",
	azurekeyvaultv1alpha1.AzureKeyVaultCertificatePermission_CERTIFICATE_UPDATE:          "Update",
	azurekeyvaultv1alpha1.AzureKeyVaultCertificatePermission_CERTIFICATE_CREATE:          "Create",
	azurekeyvaultv1alpha1.AzureKeyVaultCertificatePermission_CERTIFICATE_IMPORT:          "Import",
	azurekeyvaultv1alpha1.AzureKeyVaultCertificatePermission_CERTIFICATE_DELETE:          "Delete",
	azurekeyvaultv1alpha1.AzureKeyVaultCertificatePermission_CERTIFICATE_RECOVER:         "Recover",
	azurekeyvaultv1alpha1.AzureKeyVaultCertificatePermission_CERTIFICATE_BACKUP:          "Backup",
	azurekeyvaultv1alpha1.AzureKeyVaultCertificatePermission_CERTIFICATE_RESTORE:         "Restore",
	azurekeyvaultv1alpha1.AzureKeyVaultCertificatePermission_CERTIFICATE_MANAGE_CONTACTS: "ManageContacts",
	azurekeyvaultv1alpha1.AzureKeyVaultCertificatePermission_CERTIFICATE_MANAGE_ISSUERS:  "ManageIssuers",
	azurekeyvaultv1alpha1.AzureKeyVaultCertificatePermission_CERTIFICATE_GET_ISSUERS:     "GetIssuers",
	azurekeyvaultv1alpha1.AzureKeyVaultCertificatePermission_CERTIFICATE_LIST_ISSUERS:    "ListIssuers",
	azurekeyvaultv1alpha1.AzureKeyVaultCertificatePermission_CERTIFICATE_SET_ISSUERS:     "SetIssuers",
	azurekeyvaultv1alpha1.AzureKeyVaultCertificatePermission_CERTIFICATE_DELETE_ISSUERS:  "DeleteIssuers",
	azurekeyvaultv1alpha1.AzureKeyVaultCertificatePermission_CERTIFICATE_PURGE:           "Purge",
}

var storagePermissionStrings = map[azurekeyvaultv1alpha1.AzureKeyVaultStoragePermission]string{
	azurekeyvaultv1alpha1.AzureKeyVaultStoragePermission_STORAGE_GET:            "Get",
	azurekeyvaultv1alpha1.AzureKeyVaultStoragePermission_STORAGE_LIST:           "List",
	azurekeyvaultv1alpha1.AzureKeyVaultStoragePermission_STORAGE_SET:            "Set",
	azurekeyvaultv1alpha1.AzureKeyVaultStoragePermission_STORAGE_UPDATE:         "Update",
	azurekeyvaultv1alpha1.AzureKeyVaultStoragePermission_STORAGE_DELETE:         "Delete",
	azurekeyvaultv1alpha1.AzureKeyVaultStoragePermission_STORAGE_RECOVER:        "Recover",
	azurekeyvaultv1alpha1.AzureKeyVaultStoragePermission_STORAGE_BACKUP:         "Backup",
	azurekeyvaultv1alpha1.AzureKeyVaultStoragePermission_STORAGE_RESTORE:        "Restore",
	azurekeyvaultv1alpha1.AzureKeyVaultStoragePermission_STORAGE_PURGE:          "Purge",
	azurekeyvaultv1alpha1.AzureKeyVaultStoragePermission_STORAGE_REGENERATE_KEY: "RegenerateKey",
	azurekeyvaultv1alpha1.AzureKeyVaultStoragePermission_STORAGE_GET_SAS:        "GetSAS",
	azurekeyvaultv1alpha1.AzureKeyVaultStoragePermission_STORAGE_LIST_SAS:       "ListSAS",
	azurekeyvaultv1alpha1.AzureKeyVaultStoragePermission_STORAGE_SET_SAS:        "SetSAS",
	azurekeyvaultv1alpha1.AzureKeyVaultStoragePermission_STORAGE_DELETE_SAS:     "DeleteSAS",
}
