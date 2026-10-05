# DigitalOcean Logical Database

Creates an additional logical database inside a DigitalOcean managed database cluster -- the namespace an application's tables live in, separate from other workloads sharing the same cluster. Both spec fields are create-only: a rename replaces the database and drops its data, so the name you choose is the contract clients, users, and connection pools address for the database's whole life. Logical databases are free and instant, which makes one database per workload the natural shape.

## What Gets Created

When you deploy this Infra Component, the IaC module provisions:

- **Logical Database** -- a named database inside the referenced cluster, addressable by clients, users, and connection pools

## Before You Deploy

### Planton Setup

- **DigitalOcean Provider Connection** -- an active connection in the Connect module with a DigitalOcean API token. Map it as the default for your environment, or specify it explicitly when creating the Infra Component.
- **A DigitalOceanDatabaseCluster** -- the owning cluster, referenced by name (or an existing cluster's UUID as a literal).

### DigitalOcean Account

- Nothing beyond the cluster: logical databases are a free feature of managed databases.

## Deploy

### Console

Open the deployment store, find **DigitalOcean Logical Database**, and click **Deploy**. The creation wizard walks you through preset selection, environment and connection configuration, and spec fields. Start from the **Application Database** preset in the [Presets](#presets) tab to give one application its own database on a shared cluster.

### CLI

Create a manifest and apply it:

```yaml
apiVersion: digital-ocean.planton.dev/v1alpha1
kind: DigitalOceanDatabaseDb
metadata:
  name: orders-database
  org: acme-corp
  env: prod
spec:
  cluster:
    valueFrom:
      kind: DigitalOceanDatabaseCluster
      name: orders-postgres
      fieldPath: status.outputs.cluster_id
  databaseName: orders
```

```shell
planton apply -f do-database-db.yaml
```

This creates a logical database named `orders` on the referenced cluster, visible in its Users & Databases tab. An Infra Job tracks the provisioning in real time.

### InfraChart

When the database deploys alongside its cluster in one chart, wire the cluster reference via ValueFromRef:

```yaml
spec:
  cluster:
    valueFrom:
      kind: DigitalOceanDatabaseCluster
      name: orders-postgres
      fieldPath: status.outputs.cluster_id
  databaseName: orders
```

The InfraPipeline resolves the dependency graph, deploys the cluster first, then creates the logical database inside it.

## Key Configuration

These are the most important decisions when configuring a logical database. Explore the full field reference in the [API Explorer](#api-explorer) tab.

**The name is a contract, not a label** -- `databaseName` is how every client, pool, and user grant addresses this database, and both spec fields are create-only: editing either one replaces the logical database, which means the old one is DROPPED -- data and all -- and an empty successor appears. Choose names like API endpoints, and treat any rename as a data migration you plan (dump, create new, restore, cut over), never as a manifest edit.

**One database per workload** -- Logical databases are free and instant. Give each application, service, or experiment its own instead of piling tables into `defaultdb` -- deletion blast radius then matches workload boundaries exactly.

**Deletion is immediate and total** -- Deleting this resource drops the database and its data with no recycle bin of its own. The cluster's automatic backups (a DigitalOceanDatabaseCluster property) are the only recovery path -- restoring means provisioning a new cluster from a backup, not un-dropping the database.

**Composition order** -- Pools reference this database by name (their `dbName` is a plain string by design), and grants ride users. In a chart: cluster, then database, then user, then pool -- reference wiring enforces the cluster edge automatically; the database-before-pool edge is yours to keep.

## Outputs and Dependencies

### What This Kind Consumes

| Dependency | Field | ValueFromRef Path |
|------------|-------|-------------------|
| **DigitalOceanDatabaseCluster** | `cluster` | `status.outputs.cluster_id` |

### What This Kind Provides

After provisioning, `status.outputs` echoes `cluster_id` and `database_name` back -- DigitalOcean has no standalone database id, so the (cluster, name) pair IS the identity, and these echoes exist for addressing and verification rather than for downstream wiring. No other component consumes them via ValueFromRef: connection pools and application connection strings address this database by writing the same name, and credentials live on the cluster and its users, never here.

## Common Patterns

Browse the [Presets](#presets) tab for ready-to-deploy configurations.

**One database per application** -- each service gets its own logical database on the shared cluster instead of piling tables into `defaultdb`; pair it with a DigitalOceanDatabaseUser of the same service and a pool pointed at this name. Start from the **Application Database** preset.

**Analytics sidecar** -- a second database on the same cluster for reporting or ETL tables, keeping them out of the transactional database while sharing compute -- and giving schema experiments a disposable home whose deletion touches nothing else. Start from the **Analytics Sidecar Database** preset.

## Works With

- [**DigitalOcean Database Cluster**](/infra-catalog/digital-ocean-database-cluster) -- the owning cluster, wired via the `cluster` reference
- [**DigitalOcean Database User**](/infra-catalog/digital-ocean-database-user) -- the per-service credential that connects to this database
- [**DigitalOcean Database Connection Pool**](/infra-catalog/digital-ocean-database-connection-pool) -- a PgBouncer pool serving this database by name
