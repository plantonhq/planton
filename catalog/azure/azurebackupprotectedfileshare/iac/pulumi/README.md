# AzureBackupProtectedFileShare Pulumi Module

## Overview

Puts one Azure Files share under a backup policy's protection in a Recovery Services vault, via the classic Pulumi Azure provider (`pulumi-azure/sdk/v6`, bridged from azurerm). Creating the binding only REGISTERS protection -- the first backup runs on the policy's schedule, not immediately.

## Resources Created

- `backup.ProtectedFileShare` -- the protected item (`.../protectionContainers/StorageContainer;storage;{sa-rg};{sa-name}/protectedItems/AzureFileShare;{system-name}`)

## Outputs

- `backup_protected_file_share_id` -- the protected item's full ARM ID (Azure names it by the share's SYSTEM name, not its friendly name)

## Behavior Notes

- **Full engine parity**: the classic SDK carries the complete v5 surface (five arguments) -- ZERO parity exceptions on this kind.
- **The storage account must already be registered with the vault** (AzureBackupContainerStorageAccount). The provider runs an Inquire pass to discover protectable shares and fails loudly when the account is not registered. The spec's default reference wires `source_storage_account_id` THROUGH the registration so it always deploys first.
- **`backup_policy_id` is the only updatable field** -- everything else replaces the protection.
- **Create and delete run up to 80 minutes** (the provider's own timeout class).
- **Destroying deletes the backup data** (engines' default destroy behavior; vault soft delete -- always on -- may hold it 14 days).

## Development

```bash
go build ./...
```

The module entrypoint is `main.go` at this directory's root (the release contract); the implementation lives in `module/`.
