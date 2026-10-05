# AzureDataProtectionBackupInstance Pulumi Module

## Overview

Creates a Data Protection backup instance -- the binding that puts one datasource under a backup policy's protection -- on the classic Pulumi Azure SDK (`pulumi-azure/sdk/v6`), wire-identical to the Terraform module. Exactly one variant block is set in the spec; each variant creates its own provider resource, so ONE resource exists per deployment.

## Resources Created

Exactly one of, per the spec's variant:

- `dataprotection.BackupInstanceBlogStorage` *(the SDK's own token misspells Blob as "Blog" -- a bridge artifact over the correctly named `azurerm_data_protection_backup_instance_blob_storage`; do not "fix" it)*
- `dataprotection.BackupInstanceDisk`
- `dataprotection.BackupInstanceKubernetesCluster`
- `dataprotection.BackupInstanceMysqlFlexibleServer`
- `dataprotection.BackupInstancePostgresqlFlexibleServer`
- `dataprotection.BackupInstanceDataLakeStorage`

## Outputs

- `backup_instance_id` -- the instance's full ARM ID (whichever variant ran)
- `backup_instance_name` -- the instance's name, unique on its vault

## Behavior Notes

- **Grants precede the instance**: the vault's managed identity must hold the datasource roles BEFORE create (disk: "Disk Backup Reader" on the disk + "Disk Snapshot Contributor" on the snapshot resource group; blob/Data Lake: "Storage Account Backup Contributor" on the account; Kubernetes: the AKS Backup extension + trusted access; MySQL/PostgreSQL: the vault identity's backup roles). An authorization-class create failure means missing or still-propagating role assignments, not a module defect.
- **Nearly everything is ForceNew**; only `backup_policy_id` updates in place -- and the kubernetes_cluster variant has NO update path at all (even the policy binding replaces it).
- **Blob's container list is one-way**: it can change but never be removed entirely once set; the module omits the argument when empty so operational-only protection holds exactly.
- **Destroy deletes the backup data**: with vault soft delete on, the data lingers as a soft-deleted item for 14 days and holds the vault's own deletion for that window.
- **Backup instances carry NO tags** -- the provider has no tags argument on any of the six resources.
- **Engine parity**: the classic SDK v6.38.0 carries the FULL azurerm v5 surface for all six variant resources -- zero parity exceptions.

## Required Permissions

The deploying principal's exact grant set is cataloged in [`../permissions.yaml`](../permissions.yaml) -- note it is distinct from the VAULT IDENTITY's datasource grants above, which must also exist.
