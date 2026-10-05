# Azure Cosmos DB SQL Container

Deploys a SQL (NoSQL) API container inside a Cosmos DB database — the unit of storage, indexing, and scale-out. Documents live in containers; Azure distributes them across physical partitions by the partition key and indexes them by the container's indexing policy. The container references its database's `sql_database_id` output, composing account → database → container in one manifest set.

## What Gets Created

When you deploy this Infra Component, the IaC module provisions:

- **Cosmos DB SQL Container** -- a named container inside the referenced SQL database, with its partition key definition (single or hierarchical), optional unique key constraints, indexing policy, TTL settings, and conflict-resolution policy
- **Dedicated Throughput** (optional) -- fixed RU/s or an autoscale ceiling owned by this container alone, when either is declared; omit both to share the database's provisioned budget (or on serverless accounts)

## Before You Deploy

### Planton Setup

- **Azure Provider Connection** -- an active connection in the Connect module with credentials for the target Azure subscription. Map it as the default for your environment, or specify it explicitly when creating the Infra Component.
- **Planton Runner** -- required when using Runner-based credential delivery. Not needed for inline credentials or browser OAuth authentication modes.

### Azure Subscription

- **A Cosmos DB SQL database** to attach to. Reference an AzureCosmosdbSqlDatabase Infra Component via ValueFromRef, or provide the database's ARM ID directly.
- **A partition key design** -- the single most consequential decision for Cosmos DB performance and cost. Pick a property with high cardinality, even request distribution, and frequent use in query filters (tenantId, userId, deviceId). It is fixed at creation.

## Deploy

### Console

Open the deployment store, find **Azure Cosmos DB SQL Container**, and click **Deploy**. The creation wizard walks you through preset selection, environment and connection configuration, and spec fields. Start from the **Tenant-Partitioned Container** preset in the [Presets](#presets) tab for the multi-tenant shape.

### CLI

Create a manifest and apply it:

```yaml
apiVersion: azure.planton.dev/v1alpha1
kind: AzureCosmosdbSqlContainer
metadata:
  name: carts-container
  org: acme-corp
  env: prod
spec:
  sqlDatabaseId:
    valueFrom:
      kind: AzureCosmosdbSqlDatabase
      name: orders-database
      fieldPath: status.outputs.sql_database_id
  containerName: carts
  partitionKeyPaths:
    - /tenantId
  throughput: 400
```

```shell
planton apply -f cosmosdb-sql-container.yaml
```

This creates a hash-partitioned container with 400 RU/s of dedicated throughput and Azure's default indexing (everything, consistently). An Infra Job tracks the provisioning in real time.

### InfraChart

When deploying as part of a multi-resource environment, the whole chain composes: the InfraPipeline deploys the account, resolves its output into the database, then resolves the database's output into this container.

## Key Configuration

These are the most important decisions when configuring a SQL container. Explore the full field reference in the [API Explorer](#api-explorer) tab.

**Partition key** -- `partitionKeyPaths` (1-3 paths starting with `/`) plus `partitionKeyKind`: unset/`HASH` for the normal single-path key, `MULTI_HASH` for a hierarchical key of up to three levels (e.g. `/tenantId`, `/userId`) that routes queries carrying any prefix efficiently — hierarchical keys require `partitionKeyVersion: 2`. All fixed at creation: changing the key means a new container and a data migration.

**Throughput model** -- Leave both fields unset to share the database's provisioned budget (or on serverless accounts); set `throughput` (minimum 400 RU/s, increments of 100) for dedicated fixed capacity; or set `autoscaleMaxThroughput` (minimum 1000, increments of 1000). The two are mutually exclusive. Edits in place.

**TTL** -- `defaultTtl`: unset leaves TTL off; `-1` turns it on with no default expiry (documents opt in with their own ttl property); a positive value expires documents that many seconds after their last write. `analyticalStorageTtl` governs the analytical-store copy and requires analytical storage on the ACCOUNT; clearing a set value forces a replacement.

**Indexing policy** -- Unset applies Azure's default (consistent indexing of every path). A declared policy replaces the default wholesale — the root path `/*` must then appear in exactly one of `includedPaths`/`excludedPaths`. Excluding bulky payload paths is the main write-RU saver; composite indexes serve multi-property ORDER BY. Updates in place.

**Unique keys and conflict resolution** -- `uniqueKeys` constrain combined path values within each LOGICAL PARTITION (never the whole container). `conflictResolutionPolicy` matters on multi-region-write accounts: last-writer-wins on a numeric path (default `/_ts`) or a custom stored procedure. Both fixed at creation.

## Outputs and Dependencies

### What This Kind Consumes

| Dependency | Field | ValueFromRef Path |
|------------|-------|-------------------|
| AzureCosmosdbSqlDatabase | `sqlDatabaseId` | `status.outputs.sql_database_id` |

### What This Kind Provides

After provisioning, `status.outputs` contains values that downstream Infra Components can consume via ValueFromRef:

| Output | Description | Common Downstream Use |
|--------|-------------|----------------------|
| `sql_container_id` | The ARM ID of the container | The scope for container-level data-plane RBAC (AzureCosmosdbSqlRoleAssignment) |
| `sql_container_name` | The container's name | Application configuration |
| `sql_database_name` | The parent database's name | SDK addressing (`{database}/{container}`) |
| `cosmosdb_account_name` | The account's name | Connection string composition |

There are deliberately no endpoint or credential outputs here: connectivity and keys live on the account (AzureCosmosdbAccount's endpoint and key outputs); the container is addressed inside that connection by database and container name.

## Common Patterns

**Tenant-partitioned workhorse** — a single high-cardinality key (`/tenantId`), autoscale throughput that follows the daily traffic curve (billing 10% of the ceiling when idle), and an indexing policy that includes `/*` but excludes a bulky `/payload/*` subtree queries never filter on. Start from the **Tenant-Partitioned Container** preset.

**Hierarchical key for tenants that outgrow one partition** — `MULTI_HASH` with `/tenantId`, `/userId` (and `partitionKeyVersion: 2`) lifts the 20 GB per-partition-key-value ceiling of a simple key, and prefix queries route efficiently at every level. Pair with a composite index matching the feed's ORDER BY. Start from the **Hierarchical Partition Key** preset.

**Self-cleaning session store** — `defaultTtl: 86400` expires documents a day after their last write, `indexingMode: NONE` stops paying indexing RU (point reads by id only — wrong for anything that queries), and no throughput fields so the container rides the database's shared budget. The cheapest shape for ephemeral data. Start from the **TTL Session Store** preset.

## Works With

- [**Azure Cosmos DB SQL Database**](/infra-catalog/azure-cosmosdb-sql-database) — the parent database this container lives in, referenced via `sql_database_id`
- [**Azure Cosmos DB Account**](/infra-catalog/azure-cosmosdb-account) — the account that owns connectivity, keys, and network posture for everything inside
- [**Azure Cosmos DB SQL Role Assignment**](/infra-catalog/azure-cosmosdb-sql-role-assignment) — data-plane grants scoped to this container (`{account-id}/dbs/{database-name}/colls/{container-name}`)
