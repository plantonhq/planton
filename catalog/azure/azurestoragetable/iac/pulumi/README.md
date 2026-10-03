# AzureStorageTable - Pulumi Module

Pulumi implementation for the AzureStorageTable component.

## Architecture

```
storage.Table (single resource)
```

## Key Design Decisions

- **PARITY-EXCEPTION -- account-name addressing on this engine**: the
  spec carries the ARM-id parent like every storage child, but
  pulumi-azure v6 (verified at v6.38, the latest) has not bridged the
  table's `storageAccountId` input, so this module parses the account
  NAME from the resolved ARM id and passes `storageAccountName`. The
  created table is identical and all outputs match the Terraform
  module byte-for-byte (`table_id` carries `resource_manager_id` on
  both engines). Re-align to `StorageAccountId` when a bridge release
  carries it.
- **Table ACL windows are required** -- start + expiry are mandatory on
  table policies (Azure's contract, enforced in the spec), so they are
  passed unconditionally.
- **No URL output**: only the account knows its real table endpoint
  (partitioned-DNS accounts differ), so client URLs compose from the
  ACCOUNT's `primary_table_endpoint` output plus this table's name.
- **No Azure tags**: ARM does not support tags on tableServices/tables;
  the platform's identity tags live on the parent account.

## Operational Contract

The provider drives table creation and ACLs through the table DATA
PLANE with shared-key authorization -- the parent account must keep
`shared_access_key_enabled` true (Azure's default) for deploys to work.

## Provider

The Azure provider is built by the shared
`pulumiazureprovider.Get(ctx, iacInput.ProviderConfig)` builder, which
dispatches static client-secret, keyless web-identity (OIDC), and
ambient credential chains. Never construct a provider inline.
