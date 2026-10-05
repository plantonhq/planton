# AzureBackupContainerStorageAccount Pulumi Module

## Overview

Registers a storage account with a Recovery Services vault as a backup container -- the one-time prerequisite that lets the account's file shares be protected -- via the classic Pulumi Azure provider (`pulumi-azure/sdk/v6`, bridged from azurerm). One registration per storage-account-and-vault pair. Registration is free and moves no data.

## Resources Created

- `backup.ContainerStorageAccount` -- the registration (`.../vaults/{vault}/backupFabrics/Azure/protectionContainers/StorageContainer;storage;{sa-rg};{sa-name}`)

## Outputs

- `backup_container_id` -- the registration's full ARM ID
- `storage_account_id` -- the registered account's ARM ID, echoed so protected file shares reference the REGISTRATION for their `source_storage_account_id` (the reference carries both the value and the deploy-order edge -- the provider docs' own wiring pattern)

## Behavior Notes

- **Full engine parity**: the classic SDK carries the complete v5 surface (three arguments) -- ZERO parity exceptions on this kind.
- **Everything is ForceNew** -- ARM has no update on protection containers; changing any field replaces the registration.
- **The account must live in the vault's region** (Azure Files backup is regional).
- **Azure Backup locks the storage account while registered**; unregistering removes the lock -- and REFUSES while any of the account's shares are still protected (destroy the protections first; the GUIDE carries the teardown recipe).

## Development

```bash
go build ./...
```

The module entrypoint is `main.go` at this directory's root (the release contract); the implementation lives in `module/`.
