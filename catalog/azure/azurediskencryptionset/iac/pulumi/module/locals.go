package module

import (
	"strings"

	azurediskencryptionsetv1alpha1 "github.com/plantonhq/planton/catalog/azure/azurediskencryptionset/v1alpha1"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/provider/azure/azuretagkeys"
	"github.com/plantonhq/planton/shared/catalogkind"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

type Locals struct {
	AzureDiskEncryptionSet *azurediskencryptionsetv1alpha1.AzureDiskEncryptionSet

	ResourceGroupName string

	// KeyVaultKeyId is the resolved literal Key Vault key URL (versionless
	// for rotation-on, versioned for a pinned version).
	KeyVaultKeyId string

	// EncryptionType is the ARM string for the spec enum, or empty when
	// unspecified so both engines let Azure apply its default
	// (EncryptionAtRestWithCustomerKey).
	EncryptionType string

	// IdentityType is the ARM identity string; IdentityIds are the resolved
	// user-assigned identity ARM IDs.
	IdentityType string
	IdentityIds  []string

	AzureTags map[string]string
}

func initializeLocals(ctx *pulumi.Context, iacInput *azurediskencryptionsetv1alpha1.AzureDiskEncryptionSetIacInput) *Locals {
	locals := &Locals{}

	locals.AzureDiskEncryptionSet = iacInput.Target
	target := iacInput.Target
	spec := target.Spec

	locals.ResourceGroupName = spec.ResourceGroup.GetValue()
	locals.KeyVaultKeyId = spec.KeyVaultKeyId.GetValue()

	switch spec.EncryptionType {
	case azurediskencryptionsetv1alpha1.AzureDiskEncryptionSetEncryptionType_ENCRYPTION_AT_REST_WITH_CUSTOMER_KEY:
		locals.EncryptionType = "EncryptionAtRestWithCustomerKey"
	case azurediskencryptionsetv1alpha1.AzureDiskEncryptionSetEncryptionType_ENCRYPTION_AT_REST_WITH_PLATFORM_AND_CUSTOMER_KEYS:
		locals.EncryptionType = "EncryptionAtRestWithPlatformAndCustomerKeys"
	case azurediskencryptionsetv1alpha1.AzureDiskEncryptionSetEncryptionType_CONFIDENTIAL_VM_ENCRYPTED_WITH_CUSTOMER_KEY:
		locals.EncryptionType = "ConfidentialVmEncryptedWithCustomerKey"
	}

	if spec.Identity != nil {
		switch spec.Identity.Type {
		case azurediskencryptionsetv1alpha1.AzureDiskEncryptionSetIdentityType_SYSTEM_ASSIGNED:
			locals.IdentityType = "SystemAssigned"
		case azurediskencryptionsetv1alpha1.AzureDiskEncryptionSetIdentityType_USER_ASSIGNED:
			locals.IdentityType = "UserAssigned"
		case azurediskencryptionsetv1alpha1.AzureDiskEncryptionSetIdentityType_SYSTEM_AND_USER_ASSIGNED:
			locals.IdentityType = "SystemAssigned, UserAssigned"
		}
		for _, id := range spec.Identity.IdentityIds {
			locals.IdentityIds = append(locals.IdentityIds, id.GetValue())
		}
	}

	locals.AzureTags = map[string]string{
		azuretagkeys.Resource:     "true",
		azuretagkeys.ResourceName: target.Metadata.Name,
		azuretagkeys.ResourceKind: strings.ToLower(catalogkind.CatalogKind_AzureDiskEncryptionSet.String()),
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
	for k, v := range spec.Tags {
		locals.AzureTags[k] = v
	}

	return locals
}
