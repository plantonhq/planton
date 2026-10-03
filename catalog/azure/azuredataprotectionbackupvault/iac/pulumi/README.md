# AzureDataProtectionBackupVault Pulumi Module

## Overview

Creates a Data Protection backup vault -- the safe that modern Azure Backup data (managed disks, blob storage, AKS clusters, MySQL/PostgreSQL flexible servers, Data Lake storage) lives in -- on the classic Pulumi Azure SDK (`pulumi-azure/sdk/v6`), wire-identical to the Terraform module. The vault is free at rest; cost follows the protected instances and their backup storage.

## Resources Created

- `dataprotection.BackupVault` -- the vault
- `dataprotection.BackupVaultCustomerManagedKey` -- created only when `spec.encryption` is set (the provider's sibling resource that rewrites the vault's own security settings)

## Outputs

- `backup_vault_id` -- the vault's full ARM ID; what backup policies and backup instances reference their vault by
- `backup_vault_name` -- the vault's name
- `system_assigned_identity_principal_id` -- for Key Vault grants under customer-managed-key encryption and datasource RBAC grants, when a system identity is enabled

## Behavior Notes

- **`crossRegionRestoreEnabled` ships only when true**: the provider errors when the argument is explicitly present -- even as false -- on a non-GeoRedundant vault, so an unset spec value never reaches the provider as false. Enabling is in-place; disabling replaces the vault (one-way ForceNew).
- **Two more one-way doors**: `immutability: Locked` and `softDelete: AlwaysOn` are permanent -- leaving either replaces the vault (the provider's ForceNew transitions). `Disabled <-> Unlocked` and `On <-> Off` move freely.
- **CMK encryption can never be removed**: the composed CMK resource's delete is a documented provider no-op -- only deleting the vault removes the encryption. The KEY rotates in place (the one updatable part); versionless key URIs are accepted -- the spec reference's default target, so rotation propagates automatically. Azure unwraps the key with the vault's SYSTEM-assigned identity (the provider hardcodes it).
- **Deletion outlives the API's answer**: the provider polls past Azure's premature "deleted" response until the vault is actually gone -- destroy runs a little longer than the API suggests.
- **Engine parity**: the classic SDK v6.38.0 carries the FULL azurerm v5 surface for this kind -- zero parity exceptions.

## Required Permissions

Least-privilege runner permissions for this component are declared in [`../permissions.yaml`](../permissions.yaml).
