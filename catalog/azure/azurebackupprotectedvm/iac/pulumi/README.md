# AzureBackupProtectedVm Pulumi Module

## Overview

Registers one virtual machine under a backup policy's protection in a Recovery Services vault, via the classic Pulumi Azure provider (`pulumi-azure/sdk/v6`, bridged from azurerm). Creation only REGISTERS protection -- the first backup runs on the policy's schedule, not immediately.

## Resources Created

- `backup.ProtectedVM` -- the protected item (ARM derives its name from the VM's group and name: `VM;iaasvmcontainerv2;{vm-rg};{vm-name}`)

## Outputs

- `backup_protected_vm_id` -- the protected item's full ARM ID

## Behavior Notes

- **Full engine parity**: the classic SDK carries the complete v5 surface (source VM, policy, LUN filters, protection state) -- ZERO parity exceptions on this kind.
- **Destroy deletes backup data (deliberate)**: engine features stay at defaults, so destroying stops protection AND deletes the backup data. Retain-on-destroy is an engine-level switch for teams that need it.
- **Destroy ordering matters for the vault**: the vault refuses to delete while protected items remain -- destroy protections before their vault.
- **`protection_state` is service-managed when unset**; `BackupsSuspended` requires the VAULT to be immutable (Unlocked/Locked) -- an apply-time contract Azure checks against the live vault.

## Development

```bash
go build ./...
```

The module entrypoint is `main.go` at this directory's root (the release contract); the implementation lives in `module/`.
