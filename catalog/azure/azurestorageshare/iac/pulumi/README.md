# AzureStorageShare - Pulumi Module

Pulumi implementation for the AzureStorageShare component.

## Architecture

```
storage.Share (single resource)
```

## Key Design Decisions

- **The parent is addressed by ARM id** (`storage_account_id`) -- the
  control-plane path and the provider's v5 direction; the account-name
  form is the legacy data-plane path and is not modeled. The account
  NAME output is parsed from the id, with a loud error if the id is not
  a storage-account ARM id.
- **Unset protocol materializes `SMB`** -- azurerm's own default and
  what Windows plus most Linux mounts use.
- **The tier is sent only when chosen** -- Azure's per-account-kind
  default (TransactionOptimized on standard, Premium on FileStorage)
  applies when unset.
- **ACL windows are presence-guarded** -- share policies may leave
  start/expiry to the SAS token, so unset ends are simply not sent.
- **Two id outputs on purpose**: `share_id` is the management ARM id;
  `rbac_scope_id` (the provider's own attribute) is the DIFFERENT
  `fileshares` segment Azure Files data-plane role assignments scope to.
- **No URL output**: only the account knows its real file endpoint
  (partitioned-DNS accounts differ), so mount paths compose from the
  ACCOUNT's `primary_file_endpoint` output plus this share's name.
- **No Azure tags**: ARM does not support tags on fileServices/shares;
  the platform's identity tags live on the parent account.

## Provider

The Azure provider is built by the shared
`pulumiazureprovider.Get(ctx, iacInput.ProviderConfig)` builder, which
dispatches static client-secret, keyless web-identity (OIDC), and
ambient credential chains. Never construct a provider inline.
