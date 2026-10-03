# Azure MSSQL Database

Deploys a single Azure SQL database onto an existing logical server (AzureMssqlServer). The database is the family's workhorse: it owns its compute (a DTU/vCore/serverless/Hyperscale SKU, or membership in an elastic pool), its storage cap, its backup and retention posture, its encryption overrides, and its lifecycle story — nine create modes cover fresh creation, copies, readable geo-replicas, point-in-time restores, and the three recovery paths. The server it attaches to owns authentication, network access, and the server-level TDE key.

## What Gets Created

When you deploy this Infra Component, the IaC module provisions:

- **SQL Database** -- on the referenced logical server, born by the chosen create mode (fresh, copy, secondary, point-in-time restore, recovery, dropped-database restore, or LTR restore), with its own SKU or elastic-pool membership
- **Bacpac Import** -- executed at creation when `import` is set on a fresh database; schema and data load from a .bacpac in blob storage
- **Retention Policies** -- created when `shortTermRetentionPolicy` / `longTermRetentionPolicy` are set; the PITR window (1-35 days) and the weekly/monthly/yearly LTR vault
- **Database-Level Encryption** -- when `transparentDataEncryptionKeyVaultKeyId` is set, this database's own customer-managed TDE key (overriding the server's), with optional automatic key-version rotation
- **Threat Detection Policy** -- created when `threatDetectionPolicy` is set; a per-database Defender for SQL override
- **Azure Tags** -- resource metadata tags applied to the database for tracking and cost allocation

## Before You Deploy

### Planton Setup

- **Azure Provider Connection** -- an active connection in the Connect module with credentials for the target Azure subscription.
- **An AzureMssqlServer** -- the logical server this database attaches to, referenced through its `server_id` output (or a raw ARM ID for servers outside Planton).

### Azure Subscription

- **SKU planning** -- blank applies Azure's serverless GP_S_Gen5_2 default. DTU tiers (Basic, S0-S12, P1-P15), vCore (GP_/BC_), serverless (GP_S_/HS_S_ — auto-pause and per-second billing), or Hyperscale (HS_ — 100 TB, up to 4 named replicas). A pooled database instead references an AzureMssqlElasticPool and carries the literal SKU `ElasticPool`.
- **The permanent set** -- name, collation, ledger mode, and enclave type are fixed at creation.

## Deploy

### Console

Open the deployment store, find **Azure MSSQL Database**, and click **Deploy**. The creation wizard walks you through preset selection, environment and connection configuration, and spec fields. Start from the **General Purpose Production Database** preset in the [Presets](#presets) tab.

### CLI

Create a manifest and apply it:

```yaml
apiVersion: azure.planton.dev/v1alpha1
kind: AzureMssqlDatabase
metadata:
  name: app-database
  org: acme-corp
  env: prod
spec:
  serverId:
    valueFrom:
      kind: AzureMssqlServer
      name: app-sql
      fieldPath: status.outputs.server_id
  databaseName: appdb
  skuName: GP_Gen5_2
  zoneRedundant: true
  shortTermRetentionPolicy:
    retentionDays: 14
```

```shell
planton apply -f mssql-database.yaml
```

This creates a fresh General Purpose database on the referenced server with zone redundancy and a 14-day point-in-time restore window. An Infra Job tracks the provisioning in real time.

### InfraChart

When the database deploys alongside its server in one chart, the `serverId` reference orders it after the server:

```yaml
spec:
  serverId:
    valueFrom:
      kind: AzureMssqlServer
      name: app-sql
      fieldPath: status.outputs.server_id
  databaseName: appdb
  skuName: GP_Gen5_2
```

A pooled database additionally references its pool (`elasticPoolId`), and a replica references its primary (`creationSourceDatabaseId`) — the InfraPipeline resolves the full dependency graph.

## Key Configuration

These are the most important decisions when configuring a SQL database. Explore the full field reference in the [API Explorer](#api-explorer) tab.

**Create mode** -- `createMode` decides how the database is born. Fresh (unset/DEFAULT) optionally seeds from `sampleName` or a bacpac `import`; COPY/SECONDARY/POINT_IN_TIME_RESTORE consume `creationSourceDatabaseId` (a SECONDARY lives on a DIFFERENT server and is the failover-group building block, with `restorePointInTime` required for PITR); RECOVERY/RESTORE/RESTORE_LONG_TERM_RETENTION_BACKUP consume the raw backup-catalog IDs (`recoverDatabaseId`/`recoveryPointId`, `restoreDroppedDatabaseId`, `restoreLongTermRetentionBackupId`).

**Standalone vs. pooled** -- setting `elasticPoolId` requires `skuName: ElasticPool` and forbids a private maintenance window (the pool owns it); the wizard holds both rules structurally. Standalone databases pick their own SKU, `licenseType` (Azure Hybrid Benefit on vCore tiers), and `maxSizeGb`.

**Serverless dials** -- `autoPauseDelayInMinutes` (-1 disables pausing; 60-10080 minutes otherwise) and `minCapacity` (0.25-40 vCores) apply only to GP_S_/HS_S_ SKUs; `readReplicaCount` (0-4) applies only to Hyperscale.

**Backups** -- `storageAccountType` picks the backup-storage redundancy (GEO, the default, is what enables geo-restore); the retention policies shape the PITR window and the LTR vault. `geoBackupEnabled` applies to Data Warehouse SKUs only — every other tier controls backup redundancy through `storageAccountType`.

**Encryption** -- `transparentDataEncryptionKeyVaultKeyId` gives this database its own CMK (requires an entry in `userAssignedIdentityIds` — the identity unwraps the key); `transparentDataEncryptionKeyAutomaticRotationEnabled` adopts new key versions automatically.

## Outputs and Dependencies

### What This Kind Consumes

| Dependency | Field | ValueFromRef Path |
|------------|-------|-------------------|
| **AzureMssqlServer** | `serverId` | `status.outputs.server_id` |
| **AzureMssqlElasticPool** | `elasticPoolId` | `status.outputs.elastic_pool_id` |
| **AzureMssqlDatabase** | `creationSourceDatabaseId` | `status.outputs.database_id` |
| **AzureUserAssignedIdentity** | `userAssignedIdentityIds` | `status.outputs.identity_id` |
| **AzureKeyVaultKey** | `transparentDataEncryptionKeyVaultKeyId` | `status.outputs.key_id` |

### What This Kind Provides

After provisioning, `status.outputs` contains values that downstream Infra Components can consume via ValueFromRef:

| Output | Description | Common Downstream Use |
|--------|-------------|----------------------|
| `database_id` | Azure resource ID of the database | AzureMssqlFailoverGroup `database_ids`, another database's `creation_source_database_id` (copies, replicas, restores) |
| `database_name` | Name of the database | Application connection strings (paired with the server's fqdn) |

## Common Patterns

Browse the [Presets](#presets) tab for ready-to-deploy configurations.

**General purpose** -- The everyday production database: a provisioned vCore SKU with zone redundancy and a widened PITR window. Start from the **General Purpose Production Database** preset.

**Serverless with auto-pause** -- Per-second billing that stops when idle — the dev/test and spiky-workload shape. Start from the **Serverless Auto-Pausing Database** preset.

**Hyperscale with replicas** -- The 100 TB tier with named read replicas for read scale-out. Start from the **Hyperscale Database with Readable Replicas** preset.

## Works With

- [**Azure MSSQL Server**](/infra-catalog/azure-mssql-server) -- the logical server this database attaches to
- [**Azure MSSQL Elastic Pool**](/infra-catalog/azure-mssql-elastic-pool) -- shared capacity the database can join
- [**Azure MSSQL Failover Group**](/infra-catalog/azure-mssql-failover-group) -- lists this database for cross-region DR
- [**Azure Key Vault Key**](/infra-catalog/azure-key-vault-key) -- the database-level customer-managed TDE key
- [**Azure User Assigned Identity**](/infra-catalog/azure-user-assigned-identity) -- unwraps the database-level CMK
