# AzureStorageQueue - Pulumi Module

Pulumi implementation for the AzureStorageQueue kind.

## Architecture

```
storage.Queue (single resource)
```

## Key Design Decisions

- **The parent is addressed by ARM id** (`storage_account_id`) -- the
  control-plane path and the provider's v5 direction; the account-name
  form is the legacy data-plane path and is not modeled. The account
  NAME output is parsed from the id, with a loud error if the id is not
  a storage-account ARM id.
- **No URL output**: only the account knows its real queue endpoint
  (partitioned-DNS accounts differ), so client URLs compose from the
  ACCOUNT's `primary_queue_endpoint` output plus this queue's name.
- **No Azure tags**: ARM does not support tags on queueServices/queues;
  the platform's identity tags live on the parent account.

## Provider

The Azure provider is built by the shared
`pulumiazureprovider.Get(ctx, iacInput.ProviderConfig)` builder, which
dispatches static client-secret, keyless web-identity (OIDC), and
ambient credential chains. Never construct a provider inline.
