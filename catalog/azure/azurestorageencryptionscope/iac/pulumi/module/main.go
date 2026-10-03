package module

import (
	"fmt"
	"strings"

	"github.com/pkg/errors"
	azurestorageencryptionscopev1alpha1 "github.com/plantonhq/planton/catalog/azure/azurestorageencryptionscope/v1alpha1"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/provider/azure/pulumiazureprovider"
	"github.com/pulumi/pulumi-azure/sdk/v6/go/azure/storage"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

func Resources(ctx *pulumi.Context, iacInput *azurestorageencryptionscopev1alpha1.AzureStorageEncryptionScopeIacInput) error {
	locals := initializeLocals(ctx, iacInput)

	// Build the Azure provider from the IaC input via the shared builder, which resolves
	// the right credential mechanism (static client secret, keyless web identity, or ambient chain).
	azureProvider, err := pulumiazureprovider.Get(ctx, iacInput.ProviderConfig)
	if err != nil {
		return errors.Wrap(err, "failed to create azure provider")
	}

	spec := locals.AzureStorageEncryptionScope.Spec

	// The account name, parsed from the resolved account ARM ID for the
	// output -- consumers frequently need the account/scope name
	// pair, and this saves them a second reference. The id must END
	// with /storageAccounts/{name} (matching the Terraform module's
	// anchored regex), so a malformed or over-long id fails loudly here
	// instead of computing a wrong name.
	accountIdParts := strings.Split(locals.StorageAccountId, "/storageAccounts/")
	if len(accountIdParts) != 2 || accountIdParts[1] == "" || strings.Contains(accountIdParts[1], "/") {
		return errors.Errorf("storage_account_id %q is not a storage-account ARM id", locals.StorageAccountId)
	}
	storageAccountName := accountIdParts[1]

	// The scope is addressed by the parent account's ARM ID -- a pure
	// management-plane resource (no data-plane path exists). Deletion is
	// a SOFT-DISABLE: ARM has no true delete for scopes, so destroy
	// flips the scope's state to Disabled and the name stays reserved
	// within the account (recreating the same name re-enables it).
	// Scopes carry no Azure tags: ARM does not support tags on
	// encryptionScopes.
	scopeArgs := &storage.EncryptionScopeArgs{
		Name:             pulumi.String(spec.ScopeName),
		StorageAccountId: pulumi.String(locals.StorageAccountId),
		// Platform-managed (Microsoft.Storage) or customer-managed
		// (Microsoft.KeyVault) key ownership. With Key Vault, the
		// ACCOUNT must carry an identity with wrap/unwrap access on the
		// key's vault -- the same plumbing as the account-level
		// customer-managed key.
		Source: pulumi.String(sourceStrings[spec.Source]),
	}

	// Sent only when set -- ARM pairs the key with the Microsoft.KeyVault
	// source (the spec enforces required-when-KeyVault). The versionless
	// key URI lets rotation propagate to the scope with zero
	// intervention.
	if spec.KeyVaultKeyId.GetValue() != "" {
		scopeArgs.KeyVaultKeyId = pulumi.String(spec.KeyVaultKeyId.GetValue())
	}

	// A second, independent platform-managed encryption layer for just
	// this scope's data -- independent of the account-level
	// infrastructure encryption switch. Fixed at creation; sent only
	// when true so false means "leave it to Azure" on both engines.
	if spec.InfrastructureEncryptionRequired {
		scopeArgs.InfrastructureEncryptionRequired = pulumi.Bool(true)
	}

	createdScope, err := storage.NewEncryptionScope(ctx,
		fmt.Sprintf("%s-%s", storageAccountName, spec.ScopeName),
		scopeArgs,
		pulumi.Provider(azureProvider))
	if err != nil {
		return errors.Wrapf(err, "failed to create storage encryption scope %s", spec.ScopeName)
	}

	// Export outputs. The name is what containers
	// (default_encryption_scope), ADLS filesystems, and per-blob upload
	// options reference within the account.
	ctx.Export(OpEncryptionScopeId, createdScope.ID())
	ctx.Export(OpEncryptionScopeName, createdScope.Name)
	ctx.Export(OpStorageAccountName, pulumi.String(storageAccountName))

	return nil
}
