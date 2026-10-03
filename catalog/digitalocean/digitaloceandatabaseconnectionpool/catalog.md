# DigitalOcean Database Connection Pool

Creates a PgBouncer connection pool on a DigitalOcean managed PostgreSQL cluster -- the endpoint applications connect to when direct connections would exhaust the cluster's limit. Pools run in transaction, session, or statement mode, authenticate as one dedicated cluster user or proxy each client's own credentials, and can point at any logical database on the cluster. Every spec field is create-only -- the provider registers no update path for pools -- so a later edit replaces the pool and drops its live connections; the mode and size decisions below are worth making once, deliberately.

## What Gets Created

When you deploy this Infra Component, the IaC module provisions:

- **Connection Pool** -- a named PgBouncer pool on the referenced PostgreSQL cluster, with its own endpoint port
- **Pool Identity** -- configured only when `user` is set; the pool authenticates as that cluster user. Omitted creates DigitalOcean's inbound-user pool where clients bring their own credentials

## Before You Deploy

### Planton Setup

- **DigitalOcean Provider Connection** -- an active connection in the Connect module with a DigitalOcean API token. Map it as the default for your environment, or specify it explicitly when creating the Infra Component.
- **A DigitalOceanDatabaseCluster** -- the owning PostgreSQL cluster, referenced by name (or an existing cluster's UUID as a literal). Pools exist only on PostgreSQL clusters.

### DigitalOcean Account

- Nothing beyond the cluster: pools are a free feature of managed PostgreSQL.

## Deploy

### Console

Open the deployment store, find **DigitalOcean Database Connection Pool**, and click **Deploy**. The creation wizard walks you through preset selection, environment and connection configuration, and spec fields. Start from the **Transaction Pool for One Service** preset in the [Presets](#presets) tab for the workhorse shape: transaction mode with a dedicated service user.

### CLI

Create a manifest and apply it:

```yaml
apiVersion: digital-ocean.planton.dev/v1alpha1
kind: DigitalOceanDatabaseConnectionPool
metadata:
  name: orders-pool
  org: acme-corp
  env: prod
spec:
  cluster:
    valueFrom:
      kind: DigitalOceanDatabaseCluster
      name: orders-postgres
      fieldPath: status.outputs.cluster_id
  poolName: orders-pool
  mode: transaction
  size: 20
  dbName: orders
  user: orders-service
```

```shell
planton apply -f do-connection-pool.yaml
```

This creates a 20-connection transaction-mode pool on the referenced cluster, serving the `orders` database as the `orders-service` user. An Infra Job tracks the provisioning in real time.

### InfraChart

When the pool deploys alongside its cluster in one chart, wire the cluster reference via ValueFromRef:

```yaml
spec:
  cluster:
    valueFrom:
      kind: DigitalOceanDatabaseCluster
      name: orders-postgres
      fieldPath: status.outputs.cluster_id
  poolName: orders-pool
  mode: transaction
  size: 20
  dbName: orders
```

The InfraPipeline resolves the dependency graph, deploys the cluster first, then provisions the pool on it.

## Key Configuration

These are the most important decisions when configuring a connection pool. Explore the full field reference in the [API Explorer](#api-explorer) tab.

**Every change is an outage** -- The provider has no update path for pools: `size`, `mode`, `dbName`, `user` -- all of it is create-only, and any edit replaces the pool, dropping its live connections. Size generously up front, and schedule pool edits like brief connection outages with client retry in place.

**Pooling mode** -- `transaction` is the right default for web apps and APIs: a server connection per transaction, maximum reuse -- but session state dies (LISTEN/NOTIFY, session-scoped prepared statements, advisory locks held across transactions). `session` preserves all of that at the cost of one server connection per connected client -- size for concurrent CLIENTS, not transactions. `statement` is for autocommit-style workloads only; multi-statement transactions fail on it.

**Pool size** -- `size` is backend connections HELD OPEN, and the cluster's connection limit scales with its node size (DigitalOcean reserves a few connections for itself). Pools on the same cluster share that budget with direct connections. If the pool needs to grow past the budget, the real fix is a bigger cluster size slug, not a bigger pool.

**Dedicated user or inbound-user** -- Omit `user` and the pool proxies each client's own credentials: per-client identity, grants, and rotation stay intact, which is why it is the safer default for shared pools (the `password` output is legitimately empty in this shape). Name a user only when exactly one service owns the pool.

**Database wiring** -- `dbName` is a plain name by design: compose it with a DigitalOceanDatabaseDb resource by writing the same name, or point it at the cluster's built-in `defaultdb`. In a chart, the cluster edge is enforced by reference wiring, but the database-before-pool ordering is yours to keep.

**Connect to the pool's port, not the cluster's** -- The pool listens on its own port beside the cluster's (both on the same hosts), and clients connect to `poolName` as if it were a database name. The `uri` / `private_uri` outputs have this right already; hand-built connection strings routinely get it wrong.

## Outputs and Dependencies

### What This Kind Consumes

| Dependency | Field | ValueFromRef Path |
|------------|-------|-------------------|
| **DigitalOceanDatabaseCluster** | `cluster` | `status.outputs.cluster_id` |

### What This Kind Provides

After provisioning, `status.outputs` contains values that downstream Infra Components can consume via ValueFromRef:

| Output | Description | Common Downstream Use |
|--------|-------------|----------------------|
| `pool_name` | Name of the pool -- the "database name" clients connect to | Application connection configuration |
| `host` / `private_host` | Public and private-network hostnames of the pool endpoint | Application connection strings (private for same-VPC workloads) |
| `port` | The pool's own port -- distinct from the cluster's | Application connection strings |
| `uri` / `private_uri` | Full connection URIs including credentials (secrets) | Application configuration wired as secrets |
| `password` | The pool user's password (secret); empty for inbound-user pools | Application authentication |

`cluster_id` is also echoed: DigitalOcean has no standalone pool id, so the (cluster, name) pair is how the API addresses the pool.

## Common Patterns

Browse the [Presets](#presets) tab for ready-to-deploy configurations.

**Per-service transaction pool** -- transaction mode, a dedicated service user, pointed at the service's own logical database. Maximum connection reuse for web applications with many short transactions; pair it with a DigitalOceanDatabaseDb and DigitalOceanDatabaseUser of the same service. Start from the **Transaction Pool for One Service** preset.

**Shared inbound-user session pool** -- `user` omitted so every client authenticates with its own credentials, in session mode for workloads that need session state (LISTEN/NOTIFY, advisory locks). Trades connection reuse for per-client identity and audit trails. Start from the **Shared Inbound-User Session Pool** preset.

## Works With

- [**DigitalOcean Database Cluster**](/infra-catalog/digital-ocean-database-cluster) -- the PostgreSQL cluster the pool runs on, wired via the `cluster` reference
- [**DigitalOcean Logical Database**](/infra-catalog/digital-ocean-database-db) -- the database the pool serves; compose by writing the same name in `dbName`
- [**DigitalOcean Database User**](/infra-catalog/digital-ocean-database-user) -- the dedicated identity a single-service pool authenticates as
