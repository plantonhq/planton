# Azure Data Protection Backup Instance

Creates a Data Protection backup instance -- the binding that puts one datasource (a managed disk, a storage account's blobs, an AKS cluster, a MySQL or PostgreSQL flexible server, or a Data Lake storage account) under a backup policy's protection. The vault holds the backups, the policy says when and for how long; the instance is what makes a specific resource actually protected. The instance itself is a free binding object -- cost follows the backup storage the protected data consumes -- and nearly everything on it is fixed at creation.

## What Gets Created

When you deploy this Infra Component, the IaC module provisions:

- **Backup instance** -- ONE of the six datasource-specific ARM bindings (`Microsoft.DataProtection/backupVaults/{vault}/backupInstances/{name}`), per the variant block set in the spec: blob storage, managed disk, Kubernetes cluster, MySQL flexible server, PostgreSQL flexible server, or Data Lake storage

## Before You Deploy

### Planton Setup

- **Azure Provider Connection** -- an active connection in the Connect module with credentials for the target Azure subscription.
- **An AzureDataProtectionBackupVault** -- the vault that will hold the backups, WITH a system-assigned identity (the grants below bind to it).
- **An AzureDataProtectionBackupPolicy** -- of the SAME datasource type as the variant you deploy (a disk instance needs a disk policy, on the same vault).
- **The datasource component** -- the AzureManagedDisk, AzureStorageAccount, AzureAksCluster, or flexible-server resource being protected.

### Azure Subscription

- **The vault's managed identity needs datasource permissions BEFORE this deploys** -- Azure validates them at create time. Compose AzureRoleAssignment resources referencing the vault's `system_assigned_identity_principal_id` output: disk needs "Disk Backup Reader" on the disk and "Disk Snapshot Contributor" on the snapshot resource group; blob and Data Lake need "Storage Account Backup Contributor" on the storage account; AKS needs the Backup extension and trusted access; the flexible servers need the vault identity's backup roles.
- **The region must match the datasource's own region** -- Azure Backup protects a resource from within its region.
- **Destroying the instance stops protection and deletes the backup data** -- with vault soft delete on, the data lingers 14 days as a soft-deleted item and holds the vault's own deletion for that window.

## Deploy

### Console

Open the deployment store, find **Azure Data Protection Backup Instance**, and click **Deploy**. The creation wizard walks you through preset selection, environment and connection configuration, and spec fields. Start from the **Disk Backup**, **AKS Cluster Backup**, or **Blob Backup** preset in the [Presets](#presets) tab.

### CLI

Create a manifest and apply it:

```yaml
apiVersion: azure.planton.dev/v1alpha1
kind: AzureDataProtectionBackupInstance
metadata:
  name: app-disk-backup
  org: acme-corp
  env: prod
spec:
  vaultId:
    valueFrom:
      kind: AzureDataProtectionBackupVault
      name: prod-backup-vault
      fieldPath: status.outputs.backup_vault_id
  name: app-disk-backup
  region: eastus
  backupPolicyId:
    valueFrom:
      kind: AzureDataProtectionBackupPolicy
      name: daily-disk-policy
      fieldPath: status.outputs.backup_policy_id
  disk:
    diskId:
      valueFrom:
        kind: AzureManagedDisk
        name: app-data-disk
        fieldPath: status.outputs.disk_id
    snapshotResourceGroupName:
      valueFrom:
        kind: AzureResourceGroup
        name: backup-snapshots-rg
        fieldPath: status.outputs.resource_group_name
```

```shell
planton apply -f backup-instance.yaml
```

This puts one managed disk in eastus under the referenced disk policy's protection, with incremental snapshots landing in the named snapshot resource group. An Infra Job tracks the provisioning in real time.

### InfraChart

When deploying the whole protection stack as one chart -- vault, policy, role assignments, and instance -- ValueFromRef orders the graph so the grants exist before Azure validates them:

```yaml
spec:
  vaultId:
    valueFrom:
      kind: AzureDataProtectionBackupVault
      name: prod-backup-vault
      fieldPath: status.outputs.backup_vault_id
  backupPolicyId:
    valueFrom:
      kind: AzureDataProtectionBackupPolicy
      name: daily-disk-policy
      fieldPath: status.outputs.backup_policy_id
```

The InfraPipeline resolves the dependency graph -- vault, policy, and AzureRoleAssignment grants first, then this instance.

## Key Configuration

These are the most important decisions when configuring a backup instance. Explore the full field reference in the [API Explorer](#api-explorer) tab.

**Grants first, instance second -- always** -- Azure validates the vault identity's datasource permissions at create time, and role assignments propagate asynchronously (minutes, occasionally longer). An authorization-class create failure ("appropriate permissions", "does not have authorization") means missing or still-propagating grants, not a broken configuration; retrying after a few minutes is the first move.

**The policy's variant must match the instance's** -- a disk instance binds a DISK policy, a blob instance a BLOB policy, both on the same vault; Azure rejects mismatches at create. `backupPolicyId` is the instance's ONLY in-place-updatable field -- everything else (vault, datasource, name, region, snapshot settings) replaces the instance when changed. When you replace a policy (policies are immutable), update the instance's binding in the same change.

**The Kubernetes variant is a bigger commitment** -- the AKS cluster must carry the Backup extension and a trusted-access role binding to the vault before the instance can be created (cluster-side setup this component deliberately does not own), and the variant is immutable end to end: every change, including the policy binding, replaces the instance. Model AKS backup as part of cluster provisioning. Its `backupDatasourceParameters` decide namespace filters, cluster-scoped resources, and whether persistent-volume snapshots are taken (Azure's default is configuration only, no volume data).

**Blob containers are a one-way commitment** -- `storageAccountContainerNames` is required when the policy has a vault tier and must stay empty for an operational-only policy (which protects the whole account continuously). Once set, the list can change but never be cleared entirely -- clearing it replaces the instance.

**Deletion deletes the backups -- plan the exit** -- destroying the instance stops protection AND removes the backup data. With the vault's soft delete on (the default), the data lingers recoverable for 14 days -- but it also HOLDS the vault's deletion, and re-protecting the same datasource inside the window collides with the ghost. For estates that create and destroy protection frequently, a vault with soft delete Off trades the safety net for clean teardown.

**Protection configured is not a restore point** -- creating the instance registers protection; the FIRST backup runs on the policy's schedule. Until it completes there is nothing to restore, so check the vault's restore points before any destructive change to the datasource.

## Outputs and Dependencies

### What This Kind Consumes

| Dependency | Field | ValueFromRef Path |
|------------|-------|-------------------|
| **AzureDataProtectionBackupVault** | `vaultId` | `status.outputs.backup_vault_id` |
| **AzureDataProtectionBackupPolicy** | `backupPolicyId` | `status.outputs.backup_policy_id` |
| **AzureStorageAccount** (blob / Data Lake variants) | `blobStorage.storageAccountId`, `dataLakeStorage.storageAccountId` | `status.outputs.storage_account_id` |
| **AzureManagedDisk** (disk variant) | `disk.diskId` | `status.outputs.disk_id` |
| **AzureResourceGroup** (disk / AKS variants) | `disk.snapshotResourceGroupName`, `kubernetesCluster.snapshotResourceGroupName` | `status.outputs.resource_group_name` |
| **AzureAksCluster** (Kubernetes variant) | `kubernetesCluster.kubernetesClusterId` | `status.outputs.cluster_id` |
| **AzureMysqlFlexibleServer** (MySQL variant) | `mysqlFlexibleServer.serverId` | `status.outputs.server_id` |
| **AzurePostgresqlFlexibleServer** (PostgreSQL variant) | `postgresqlFlexibleServer.serverId` | `status.outputs.server_id` |

### What This Kind Provides

The instance's `status.outputs` carries `backup_instance_id` (the ARM ID of the protection binding) and `backup_instance_name` -- identifiers for operational tooling and audit, not wiring edges: no other catalog kind consumes a backup instance. The protection itself is the product; restore operations happen through the vault, not through these outputs.

## Common Patterns

Browse the [Presets](#presets) tab for ready-to-deploy configurations.

**Disk protection** -- one managed disk under scheduled incremental snapshots, kept in a dedicated snapshot resource group so backup storage is visible and governable apart from the workload. Start from the **Disk Backup** preset.

**AKS workload protection** -- scheduled cluster backups with the plumbing namespaces excluded and persistent-volume snapshots enabled; pair with the cluster-side Backup extension setup. Start from the **AKS Cluster Backup** preset.

**Blob vault-tier protection** -- named containers backed up to the vault tier (plus whatever operational-tier protection the referenced blob policy configures). Start from the **Blob Backup** preset.

**Whole-stack chart** -- vault, policies, AzureRoleAssignment grants, and instances in one InfraChart, so the reference wiring orders the grants before Azure validates them at instance create.

## Works With

- [**Azure Data Protection Backup Vault**](/infra-catalog/azure-data-protection-backup-vault) -- the vault holding this instance's backups
- [**Azure Data Protection Backup Policy**](/infra-catalog/azure-data-protection-backup-policy) -- the schedule and retention governing this instance
- [**Azure Role Assignment**](/infra-catalog/azure-role-assignment) -- the datasource grants the vault's identity needs before create
- [**Azure Managed Disk**](/infra-catalog/azure-managed-disk) -- the disk variant's datasource
- [**Azure Storage Account**](/infra-catalog/azure-storage-account) -- the blob and Data Lake variants' datasource
- [**Azure AKS Cluster**](/infra-catalog/azure-aks-cluster) -- the Kubernetes variant's datasource
- [**Azure MySQL Flexible Server**](/infra-catalog/azure-mysql-flexible-server) / [**Azure PostgreSQL Flexible Server**](/infra-catalog/azure-postgresql-flexible-server) -- the database variants' datasources
- [**Azure Resource Group**](/infra-catalog/azure-resource-group) -- where disk and AKS snapshots land
