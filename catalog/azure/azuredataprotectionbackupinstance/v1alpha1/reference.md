# AzureDataProtectionBackupInstance

> Generated from the protobuf schema by `make generate-reference` -- do not
> edit by hand. To change a fact on this page, change the proto field comment
> or validation rule it is derived from, then regenerate.

**apiVersion**: `azure.planton.dev/v1alpha1`

**Guide**: [GUIDE.md](../GUIDE.md) -- authored operational judgment for this component: conventions, trade-offs, and what pairs well with it.

**AzureDataProtectionBackupInstanceSpec** defines a Data Protection
backup instance (ARM: Microsoft.DataProtection/backupVaults/{vault}/
backupInstances/{name}) -- the binding that puts ONE datasource
under a backup policy's protection. The vault holds the backups,
the policy says when and for how long; the instance is what makes a
specific resource actually protected. ONE kind covers the six
datasource types as variants: blob storage, managed disks,
Kubernetes (AKS) clusters, MySQL flexible servers, PostgreSQL
flexible servers and Data Lake storage. Exactly one variant block
is set; the block IS the datasource type.

**The vault's managed identity needs datasource permissions BEFORE
the instance is created** -- Azure validates them at create time and
the deploy fails without them (an authorization-class error here
means missing role assignments, not a module defect). The roles per
datasource: disk -> "Disk Backup Reader" on the disk + "Disk
Snapshot Contributor" on the snapshot resource group; blob storage
-> "Storage Account Backup Contributor" on the storage account;
Data Lake -> "Storage Account Backup Contributor" on the account;
Kubernetes -> the AKS Backup extension installed on the cluster plus
its trusted-access role binding and reader roles; MySQL/PostgreSQL
flexible server -> the vault identity's reader/backup roles on the
server. Compose AzureRoleAssignment resources referencing the
vault's system_assigned_identity_principal_id output.

**Nearly everything is fixed at creation**: the vault, datasource,
name, region and snapshot settings all replace the instance when
changed. Only the policy binding (backup_policy_id) updates in
place -- and on the kubernetes_cluster variant even that replaces
the instance (the provider ships no update path for it).

**The instance itself is a free binding object** -- cost follows the
backup storage the protected data consumes. Destroying the instance
stops protection and deletes the backup data (subject to the
vault's soft-delete setting, which can retain it for 14 days and
hold the vault's own deletion).

## Example

```yaml
# Offline-plan test manifest. Exercises the disk variant at full
# population (the richest single-variant shape: datasource ref,
# snapshot resource group, and the cross-subscription snapshot home).
# The other five variants are planned from per-variant manifest
# variations at proof time -- exactly one variant is ever set (the
# spec's exactly-one rule).
apiVersion: azure.planton.dev/v1alpha1
kind: AzureDataProtectionBackupInstance
metadata:
  name: test-data-protection-backup-instance
  org: test-org
  env: dev
spec:
  vaultId:
    value: /subscriptions/00000000-0000-0000-0000-000000000000/resourceGroups/backup-rg/providers/Microsoft.DataProtection/backupVaults/prod-vault
  name: app-data-disk
  region: eastus
  backupPolicyId:
    value: /subscriptions/00000000-0000-0000-0000-000000000000/resourceGroups/backup-rg/providers/Microsoft.DataProtection/backupVaults/prod-vault/backupPolicies/daily-disk
  disk:
    diskId:
      value: /subscriptions/00000000-0000-0000-0000-000000000000/resourceGroups/app-rg/providers/Microsoft.Compute/disks/app-data
    snapshotResourceGroupName:
      value: backup-snapshots-rg
    snapshotSubscriptionId: 11111111-2222-3333-4444-555555555555
```

## Spec Fields

| Path | Type | Required | Default | References |
|---|---|---|---|---|
| `spec.vaultId` | `string \| valueFrom` | yes |  | AzureDataProtectionBackupVault (`status.outputs.backup_vault_id`) |
| `spec.name` | `string` | yes |  |  |
| `spec.region` | `string` | yes |  |  |
| `spec.backupPolicyId` | `string \| valueFrom` | yes |  | AzureDataProtectionBackupPolicy (`status.outputs.backup_policy_id`) |
| `spec.blobStorage` | `AzureDataProtectionBackupInstanceBlobStorage` |  |  |  |
| `spec.blobStorage.storageAccountId` | `string \| valueFrom` | yes |  | AzureStorageAccount (`status.outputs.storage_account_id`) |
| `spec.blobStorage.storageAccountContainerNames` | `[]string` |  |  |  |
| `spec.disk` | `AzureDataProtectionBackupInstanceDisk` |  |  |  |
| `spec.disk.diskId` | `string \| valueFrom` | yes |  | AzureManagedDisk (`status.outputs.disk_id`) |
| `spec.disk.snapshotResourceGroupName` | `string \| valueFrom` | yes |  | AzureResourceGroup (`status.outputs.resource_group_name`) |
| `spec.disk.snapshotSubscriptionId` | `string` |  |  |  |
| `spec.kubernetesCluster` | `AzureDataProtectionBackupInstanceKubernetesCluster` |  |  |  |
| `spec.kubernetesCluster.kubernetesClusterId` | `string \| valueFrom` | yes |  | AzureAksCluster (`status.outputs.cluster_id`) |
| `spec.kubernetesCluster.snapshotResourceGroupName` | `string \| valueFrom` | yes |  | AzureResourceGroup (`status.outputs.resource_group_name`) |
| `spec.kubernetesCluster.backupDatasourceParameters` | `AzureDataProtectionBackupInstanceKubernetesClusterDatasourceParameters` |  |  |  |
| `spec.kubernetesCluster.backupDatasourceParameters.includedNamespaces` | `[]string` |  |  |  |
| `spec.kubernetesCluster.backupDatasourceParameters.excludedNamespaces` | `[]string` |  |  |  |
| `spec.kubernetesCluster.backupDatasourceParameters.includedResourceTypes` | `[]string` |  |  |  |
| `spec.kubernetesCluster.backupDatasourceParameters.excludedResourceTypes` | `[]string` |  |  |  |
| `spec.kubernetesCluster.backupDatasourceParameters.labelSelectors` | `[]string` |  |  |  |
| `spec.kubernetesCluster.backupDatasourceParameters.clusterScopedResourcesEnabled` | `bool` |  | `false` |  |
| `spec.kubernetesCluster.backupDatasourceParameters.volumeSnapshotEnabled` | `bool` |  | `false` |  |
| `spec.mysqlFlexibleServer` | `AzureDataProtectionBackupInstanceMysqlFlexibleServer` |  |  |  |
| `spec.mysqlFlexibleServer.serverId` | `string \| valueFrom` | yes |  | AzureMysqlFlexibleServer (`status.outputs.server_id`) |
| `spec.postgresqlFlexibleServer` | `AzureDataProtectionBackupInstancePostgresqlFlexibleServer` |  |  |  |
| `spec.postgresqlFlexibleServer.serverId` | `string \| valueFrom` | yes |  | AzurePostgresqlFlexibleServer (`status.outputs.server_id`) |
| `spec.dataLakeStorage` | `AzureDataProtectionBackupInstanceDataLakeStorage` |  |  |  |
| `spec.dataLakeStorage.storageAccountId` | `string \| valueFrom` | yes |  | AzureStorageAccount (`status.outputs.storage_account_id`) |
| `spec.dataLakeStorage.storageContainerNames` | `[]string` | yes |  |  |

## Field Details

### spec.vaultId

`string | valueFrom` · required

The Data Protection backup vault that will hold this instance's
backups, by ARM ID (an instance is an ARM child of its vault).
Fixed at creation. The vault's managed identity is the principal
Azure Backup acts as -- it must hold the datasource roles listed
above before this instance is created.

- references: AzureDataProtectionBackupVault (`status.outputs.backup_vault_id`)
- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: AzureDataProtectionBackupVault, name: <that resource's name>, fieldPath: status.outputs.backup_vault_id}} -- a bare string does not parse

### spec.name

`string` · required

The instance's name, unique on the vault. Fixed at creation.

- rule: {"required":true}

### spec.region

`string` · required

The Azure region of the instance -- must match the datasource's
own region (Azure Backup protects a resource from within its
region). Fixed at creation.

- rule: {"required":true}

### spec.backupPolicyId

`string | valueFrom` · required

The Data Protection backup policy that governs this instance, by
ARM ID. The policy must live on the SAME vault and be of the SAME
datasource type as the variant below (a disk instance needs a
disk policy). This is the instance's only in-place-updatable
field -- EXCEPT on the kubernetes_cluster variant, where changing
the policy replaces the instance (the provider ships no update
path there).

- references: AzureDataProtectionBackupPolicy (`status.outputs.backup_policy_id`)
- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: AzureDataProtectionBackupPolicy, name: <that resource's name>, fieldPath: status.outputs.backup_policy_id}} -- a bare string does not parse

### spec.blobStorage

`AzureDataProtectionBackupInstanceBlobStorage`

The blob-storage variant: protects a storage account's blob
services (operational tier and/or vault tier, per the policy).

### spec.blobStorage.storageAccountId

`string | valueFrom` · required

The storage account whose blob services are protected, by ARM ID.
Fixed at creation. The vault's identity needs the "Storage
Account Backup Contributor" role on this account before create.

- references: AzureStorageAccount (`status.outputs.storage_account_id`)
- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: AzureStorageAccount, name: <that resource's name>, fieldPath: status.outputs.storage_account_id}} -- a bare string does not parse

### spec.blobStorage.storageAccountContainerNames

`[]string`

The container names to back up. Required when the policy has a
VAULT tier (vaulted or operational+vaulted policies back up
specific containers); leave empty for an operational-only policy,
which protects the whole account continuously. ONE-WAY once set:
the list can be changed but never removed entirely -- clearing it
replaces the instance (the provider's own contract).

### spec.disk

`AzureDataProtectionBackupInstanceDisk`

The managed-disk variant: protects one managed disk with
incremental snapshots kept in a snapshot resource group.

### spec.disk.diskId

`string | valueFrom` · required

The managed disk to protect, by ARM ID. Fixed at creation. The
vault's identity needs the "Disk Backup Reader" role on this disk
before create.

- references: AzureManagedDisk (`status.outputs.disk_id`)
- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: AzureManagedDisk, name: <that resource's name>, fieldPath: status.outputs.disk_id}} -- a bare string does not parse

### spec.disk.snapshotResourceGroupName

`string | valueFrom` · required

The resource group (by name) where Azure Backup stores the disk
snapshots. Fixed at creation. The vault's identity needs the
"Disk Snapshot Contributor" role on this group before create.

- references: AzureResourceGroup (`status.outputs.resource_group_name`)
- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: AzureResourceGroup, name: <that resource's name>, fieldPath: status.outputs.resource_group_name}} -- a bare string does not parse

### spec.disk.snapshotSubscriptionId

`string` · optional (explicit presence)

The subscription holding the snapshot resource group, when it
differs from the vault's own subscription (cross-subscription
snapshot storage). Leave unset to use the vault's subscription --
the common case. Fixed at creation.

- rule: {"string":{"uuid":true}}

### spec.kubernetesCluster

`AzureDataProtectionBackupInstanceKubernetesCluster`

The Kubernetes (AKS) cluster variant: protects an AKS cluster's
workloads (with optional namespace/resource filters and volume
snapshots). This variant is IMMUTABLE end to end -- every field
including the policy binding replaces the instance when changed.

### spec.kubernetesCluster.kubernetesClusterId

`string | valueFrom` · required

The AKS cluster to protect, by ARM ID. Fixed at creation. The
cluster must carry the AKS Backup extension and its
trusted-access role binding to the vault before create -- an
apply-time contract Azure enforces, not something this spec can
check.

- references: AzureAksCluster (`status.outputs.cluster_id`)
- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: AzureAksCluster, name: <that resource's name>, fieldPath: status.outputs.cluster_id}} -- a bare string does not parse

### spec.kubernetesCluster.snapshotResourceGroupName

`string | valueFrom` · required

The resource group (by name) where Azure Backup stores the
cluster's snapshots. Fixed at creation.

- references: AzureResourceGroup (`status.outputs.resource_group_name`)
- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: AzureResourceGroup, name: <that resource's name>, fieldPath: status.outputs.resource_group_name}} -- a bare string does not parse

### spec.kubernetesCluster.backupDatasourceParameters

`AzureDataProtectionBackupInstanceKubernetesClusterDatasourceParameters`

What the backup includes. Leave unset to back up every namespace
with the service defaults (no cluster-scoped resources, no volume
snapshots). Fixed at creation.

### spec.kubernetesCluster.backupDatasourceParameters.includedNamespaces

`[]string`

Namespaces to include. Empty means all namespaces (minus any
exclusions below).

### spec.kubernetesCluster.backupDatasourceParameters.excludedNamespaces

`[]string`

Namespaces to exclude from the backup.

### spec.kubernetesCluster.backupDatasourceParameters.includedResourceTypes

`[]string`

Kubernetes resource types to include (e.g.
"deployments.apps"). Empty means all types.

### spec.kubernetesCluster.backupDatasourceParameters.excludedResourceTypes

`[]string`

Kubernetes resource types to exclude from the backup.

### spec.kubernetesCluster.backupDatasourceParameters.labelSelectors

`[]string`

Label selectors -- only resources matching these selectors are
backed up (e.g. "app=commerce").

### spec.kubernetesCluster.backupDatasourceParameters.clusterScopedResourcesEnabled

`bool` · optional (explicit presence)

Whether cluster-scoped resources (CRDs, cluster roles, ...) join
the backup. Azure's default is false (namespaced resources only), and a
switch left unset states that default; `false` says it out loud.

- default: `false`

### spec.kubernetesCluster.backupDatasourceParameters.volumeSnapshotEnabled

`bool` · optional (explicit presence)

Whether persistent-volume snapshots are taken with each backup.
Azure's default is false (configuration only, no volume data), and a
switch left unset states that default; `false` says it out loud.

- default: `false`

### spec.mysqlFlexibleServer

`AzureDataProtectionBackupInstanceMysqlFlexibleServer`

The MySQL flexible-server variant: protects one MySQL flexible
server with vault-tier full backups.

### spec.mysqlFlexibleServer.serverId

`string | valueFrom` · required

The MySQL flexible server to protect, by ARM ID. Fixed at
creation. The vault's identity needs its backup roles on the
server before create.

- references: AzureMysqlFlexibleServer (`status.outputs.server_id`)
- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: AzureMysqlFlexibleServer, name: <that resource's name>, fieldPath: status.outputs.server_id}} -- a bare string does not parse

### spec.postgresqlFlexibleServer

`AzureDataProtectionBackupInstancePostgresqlFlexibleServer`

The PostgreSQL flexible-server variant: protects one PostgreSQL
flexible server with vault-tier full backups.

### spec.postgresqlFlexibleServer.serverId

`string | valueFrom` · required

The PostgreSQL flexible server to protect, by ARM ID. Fixed at
creation. The vault's identity needs its backup roles on the
server before create.

- references: AzurePostgresqlFlexibleServer (`status.outputs.server_id`)
- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: AzurePostgresqlFlexibleServer, name: <that resource's name>, fieldPath: status.outputs.server_id}} -- a bare string does not parse

### spec.dataLakeStorage

`AzureDataProtectionBackupInstanceDataLakeStorage`

The Data Lake storage variant: protects a hierarchical-namespace
storage account's ADLS blob services, container by container.

### spec.dataLakeStorage.storageAccountId

`string | valueFrom` · required

The storage account to protect, by ARM ID. Must have hierarchical
namespace (Data Lake Gen2) enabled. Fixed at creation. The
vault's identity needs the "Storage Account Backup Contributor"
role on this account before create.

- references: AzureStorageAccount (`status.outputs.storage_account_id`)
- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: AzureStorageAccount, name: <that resource's name>, fieldPath: status.outputs.storage_account_id}} -- a bare string does not parse

### spec.dataLakeStorage.storageContainerNames

`[]string` · required

The storage containers to back up -- at least one, at most 1,000
(the provider's own bounds). Container names are 3-63 characters
of lowercase letters, digits and hyphens, never starting with a
hyphen. Updatable in place.

- rule: {"repeated":{"minItems":"1","maxItems":"1000","items":{"string":{"minLen":"3","maxLen":"63","pattern":"^[0-9a-z][0-9a-z-]*$"}}}}

## Validation Rules

- `exactly_one_variant`: exactly one of blob_storage, disk, kubernetes_cluster, mysql_flexible_server, postgresql_flexible_server or data_lake_storage must be set -- the block is the datasource type

## Outputs

Reference an output from another manifest as `valueFrom: {kind: AzureDataProtectionBackupInstance, name: <resource-name>, fieldPath: status.outputs.<output>}`.

| Output | Type | Description |
|---|---|---|
| `status.outputs.backup_instance_id` | `string` | The Azure Resource Manager ID of the backup instance. Format: /subscriptions/{sub}/resourceGroups/{rg}/providers/Microsoft.DataProtection/backupVaults/{vault}/backupInstances/{name} |
| `status.outputs.backup_instance_name` | `string` | The instance's name, unique on its vault. |

## References

Fields that can point at another resource's outputs:

| Field | Kind | Output |
|---|---|---|
| `spec.vaultId` | AzureDataProtectionBackupVault | `status.outputs.backup_vault_id` |
| `spec.backupPolicyId` | AzureDataProtectionBackupPolicy | `status.outputs.backup_policy_id` |
| `spec.blobStorage.storageAccountId` | AzureStorageAccount | `status.outputs.storage_account_id` |
| `spec.disk.diskId` | AzureManagedDisk | `status.outputs.disk_id` |
| `spec.disk.snapshotResourceGroupName` | AzureResourceGroup | `status.outputs.resource_group_name` |
| `spec.kubernetesCluster.kubernetesClusterId` | AzureAksCluster | `status.outputs.cluster_id` |
| `spec.kubernetesCluster.snapshotResourceGroupName` | AzureResourceGroup | `status.outputs.resource_group_name` |
| `spec.mysqlFlexibleServer.serverId` | AzureMysqlFlexibleServer | `status.outputs.server_id` |
| `spec.postgresqlFlexibleServer.serverId` | AzurePostgresqlFlexibleServer | `status.outputs.server_id` |
| `spec.dataLakeStorage.storageAccountId` | AzureStorageAccount | `status.outputs.storage_account_id` |

## See Also

- [Overview](../README.md)
