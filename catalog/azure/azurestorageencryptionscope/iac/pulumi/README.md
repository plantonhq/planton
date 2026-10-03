# AzureStorageEncryptionScope - Pulumi Module

Pulumi implementation for the AzureStorageEncryptionScope deployment
component.

## Architecture

```
storage.EncryptionScope (single resource)
```

## Key Design Decisions

- **The parent is addressed by ARM id** (`storage_account_id`) -- a
  pure management-plane resource with no legacy name form. The account
  NAME output is parsed from the id, with a loud error if the id is not
  a storage-account ARM id.
- **The key reference rides the versionless URI** -- rotation
  propagates to the scope with zero intervention (the account-CMK
  precedent); sent only when set, and the spec enforces
  required-when-KeyVault.
- **`infrastructure_encryption_required` is sent only when true** --
  the one-way-flag convention on both engines: false means "leave it to
  Azure."
- **Soft-disable delete semantics**: ARM has no true delete for scopes
  -- destroy flips the state to Disabled, the name stays reserved
  within the account, and recreating the same name re-enables it.
- **No Azure tags**: ARM does not support tags on encryptionScopes; the
  platform's identity tags live on the parent account.

## Provider

The Azure provider is built by the shared
`pulumiazureprovider.Get(ctx, iacInput.ProviderConfig)` builder, which
dispatches static client-secret, keyless web-identity (OIDC), and
ambient credential chains. Never construct a provider inline.
