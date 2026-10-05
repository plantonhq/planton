# Overview

The **AzureDataProtectionBackupInstance** component creates a Data Protection backup instance -- the binding that puts ONE datasource under a backup policy's protection. The vault holds the backups and the policy says when and how long; the instance is what makes a specific resource actually protected. One kind covers all six modern-backup datasource types as selectable variants: blob storage, managed disks, Kubernetes (AKS) clusters, MySQL flexible servers, PostgreSQL flexible servers, and Data Lake storage. The instance itself is a free binding object -- cost follows the backup storage the protected data consumes.

## Purpose

- **Completes the modern-backup story**: with the vault (AzureDataProtectionBackupVault) and the policy (AzureDataProtectionBackupPolicy), this is the third piece that turns "we have backup rules" into "this disk / account / cluster / database is protected".
- **One kind, six datasources**: the variant block IS the datasource type -- charts compose disk protection and blob protection from the same catalog kind.
- **Declarative protection**: which resources are backed up is reviewed and versioned like everything else, instead of living in portal clicks.

## Key Features

- Full azurerm v5 surface across all six backup-instance resources, including the AKS variant's namespace/resource filters and volume-snapshot switch.
- Typed references end to end: the vault, the policy, and every datasource (storage account, disk, AKS cluster, flexible server) wire through other components' outputs.
- Chart-ready: publishes `backup_instance_id` and echoes the instance name for downstream references.

## Use Cases

- **Disk backup**: scheduled incremental snapshots of a managed disk, kept in a dedicated snapshot resource group.
- **Blob backup**: operational (continuous, in-account) and/or vaulted protection of a storage account's containers.
- **AKS backup**: scheduled cluster backups with namespace filters and optional persistent-volume snapshots.
- **Database backup**: vault-tier full backups of MySQL/PostgreSQL flexible servers with long retention.

## Future Enhancements

- Live-proven lanes for the Kubernetes, database, and Data Lake variants as their fixture chains land in the test estate.

---

© Planton. Licensed under [Apache-2.0](https://github.com/plantonhq/planton/blob/main/LICENSE).
