# AzureStorageContainer - Pulumi Module

Pulumi implementation for the AzureStorageContainer deployment
component.

## Architecture

```
storage.Container (single resource)
```

## Key Design Decisions

- **The parent is addressed by ARM id** (`storage_account_id`) -- the
  control-plane path and the provider's v5 direction; the account-name
  form is the legacy data-plane path and is not modeled. The account
  NAME output is parsed from the id, with a loud error if the id is not
  a storage-account ARM id.
- **Unset access type materializes `private`** -- the container is born
  locked down unless the spec says otherwise; the enum map carries
  azurerm's lowercase wire values.
- **`encryption_scope_override_enabled` is presence-guarded, not
  defaulted** -- Azure's own default (true) only applies when a scope is
  set, so an unset flag is simply not sent.
- **No URL output**: only the account knows its real blob endpoint
  (partitioned-DNS accounts differ), so URLs compose from the ACCOUNT's
  `primary_blob_endpoint` output plus this container's name.
- **No Azure tags**: ARM does not support tags on
  blobServices/containers; the platform's identity tags live on the
  parent account.

## Provider

The Azure provider is built by the shared
`pulumiazureprovider.Get(ctx, iacInput.ProviderConfig)` builder, which
dispatches static client-secret, keyless web-identity (OIDC), and
ambient credential chains. Never construct a provider inline.
