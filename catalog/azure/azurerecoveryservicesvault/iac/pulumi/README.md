# AzureRecoveryServicesVault Pulumi Module

## Overview

Creates a Recovery Services vault -- the safe that classic Azure Backup data (VM and file-share backups) and Site Recovery configuration live in -- via the classic Pulumi Azure provider (`pulumi-azure/sdk/v6`, bridged from azurerm). The vault is free at rest; cost follows the protected items and their backup storage.

## Resources Created

- `recoveryservices.Vault` -- the vault
- `recoveryservices.VaultResourceGuardAssociation` -- created only when `spec.resource_guard_id` is set (Multi-User Authorization; ARM pins the association's name to the literal `VaultProxy`)

## Outputs

- `recovery_services_vault_id` -- the vault's full ARM ID
- `recovery_services_vault_name` -- what backup policies and protected items address their vault by
- `system_assigned_identity_principal_id` -- for Key Vault grants under customer-managed-key encryption, when a system identity is enabled
- `resource_guard_association_id` -- the association's ARM ID when the guard arm is configured; empty otherwise

## Behavior Notes

- **PARITY-EXCEPTION -- three v5-only monitoring switches**: `alerts_for_all_failover_issues_enabled`, `alerts_for_all_replication_issues_enabled` and `email_notifications_for_site_recovery_enabled` are new in azurerm v5 and ABSENT from the classic Pulumi SDK v6.38.0. This module fails loudly when any of them is set to an explicit FALSE (the only value that diverges from the service default); leave them unset, or deploy with the Terraform module to turn them off. An explicit true is wire-equivalent to the default and passes.
- **Destroy semantics (deliberate)**: provider features stay at defaults -- deleting the vault fails while protected items remain inside it. Stop and delete protections first.
- **Immutability is a one-way street at the end**: Disabled <-> Unlocked -> Locked; leaving Locked replaces the vault. The provider stages a requested Locked through Unlocked automatically.
- **Cross-region restore is one-way ForceNew** and requires geo-redundant storage (spec CEL mirrors the provider's apply-time check).
- **CMK encryption ratchets**: once enabled it can never be disabled, `infrastructure_encryption_enabled` can never change, and the `sku` freezes. Versionless key URIs are accepted and rotate automatically -- the spec reference's default target.

## Development

```bash
go build ./...
```

The module entrypoint is `main.go` at this directory's root (the release contract); the implementation lives in `module/`.
