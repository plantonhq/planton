# DigitalOcean Database Cluster

Deploys a managed database cluster on DigitalOcean supporting PostgreSQL, MySQL, Redis, MongoDB, Kafka, OpenSearch, and Valkey engines with configurable node sizing, replication, optional VPC placement, custom storage with autoscale, maintenance windows, backup-restore provisioning, and engine-specific tuning. Three decisions lock in at creation: VPC placement can never be changed, storage only ever grows, and raising `engineVersion` performs an in-place major upgrade with no downgrade path. The cluster is also the anchor of a family of satellite resources -- logical databases, users, connection pools, firewall rules, replicas, and Kafka topics all reference its `cluster_id`.

## What Gets Created

When you deploy this Cloud Resource, the IaC module provisions:

- **DigitalOcean Database Cluster** -- a managed database cluster in the specified region with the configured engine, version, node size, and node count
- **VPC Network Attachment** -- configured only when `vpc` is provided; places the cluster within a private network for secure internal access
- **Custom Storage** -- configured only when `storageGib` is provided; overrides the default storage allocation for the chosen node size, with optional automatic growth via `storageAutoscale`
- **Maintenance Window** -- configured only when `maintenanceWindow` is provided; pins automatic updates to a weekly slot
- **Engine Tuning** -- `sqlMode` on MySQL clusters and `evictionPolicy` on Redis/Valkey clusters
- **Backup Restore** -- consumed only when `backupRestore` is provided; provisions the new cluster from a backup of an existing cluster (creation-time only, never read back)
- **DigitalOcean Tags** -- your `tags` plus six resource metadata tags (resource marker, name, kind, organization, environment, id) applied automatically for tracking; DigitalOcean caps the comma-joined total at 255 characters, and both provisioners check the budget before creating anything

## Before You Deploy

### Planton Setup

- **DigitalOcean Provider Connection** -- an active connection in the Connect module with a DigitalOcean API token. Map it as the default for your environment, or specify it explicitly when creating the Cloud Resource.
- **Planton Runner** -- required when using Runner-based credential delivery. Not needed for inline API token authentication.

### DigitalOcean Account

- **A target region** where managed databases are available. Not all regions support all engines -- list what is offered where via `doctl databases options regions`.
- **A VPC network** (recommended for production) in the target region. Provide the VPC UUID directly or reference a DigitalOceanVpc Cloud Resource via ValueFromRef.
- **A valid node size slug** (e.g., `"db-s-2vcpu-4gb"`) -- check available database sizes via `doctl databases options slugs`.

## Deploy

### Console

Open the deployment store, find **DigitalOcean Database Cluster**, and click **Deploy**. The creation wizard walks you through preset selection, environment and connection configuration, and spec fields. Start from the **Production PostgreSQL HA** preset in the [Presets](#presets) tab for a production-ready multi-node configuration.

### CLI

Create a manifest and apply it:

```yaml
apiVersion: digital-ocean.planton.dev/v1alpha1
kind: DigitalOceanDatabaseCluster
metadata:
  name: app-db
  org: acme-corp
  env: prod
spec:
  clusterName: app-db
  engine: pg
  engineVersion: "16"
  region: nyc3
  sizeSlug: db-s-2vcpu-4gb
  nodeCount: 3
```

```shell
planton apply -f do-database.yaml
```

This creates a 3-node PostgreSQL 16 cluster in the NYC3 region with no VPC attachment. A Stack Job tracks the provisioning in real time.

### InfraChart

When deploying as part of a multi-resource environment, use ValueFromRef to wire the database cluster to a VPC deployed in the same InfraPipeline:

```yaml
spec:
  vpc:
    valueFrom:
      kind: DigitalOceanVpc
      name: app-network
      fieldPath: status.outputs.vpc_id
```

The InfraPipeline resolves the dependency graph, deploys the VPC first, then provisions the database cluster within it.

## Key Configuration

These are the most important decisions when configuring a database cluster. Explore the full field reference in the [API Explorer](#api-explorer) tab.

**Engine selection** -- The `engine` values are DigitalOcean's own API slugs: PostgreSQL is `pg`, never `postgres`. Redis and Valkey are two slugs for one caching product line, and DigitalOcean no longer creates Redis clusters (a create with `engine: redis` is rejected) -- `redis` only adopts an existing cluster; every new cache is `valkey`. `engineVersion` must be a version DigitalOcean currently offers for the engine (`GET /v2/databases/options`; MySQL is `"8.4"` only), or the create fails with `422 invalid cluster engine version`. Raising `engineVersion` performs an in-place major upgrade on the running cluster with no downgrade and no blue-green; take a `backupRestore` copy first when application compatibility is unproven.

**Node count and high availability** -- Valid node counts are engine-specific: PostgreSQL/MySQL/MongoDB accept 1-3 (2 buys a standby without quorum, so most teams go straight to 3 for automatic failover), Kafka requires at least 3, OpenSearch scales to 15, and single-node Redis/Valkey is normal because caches tolerate a failover gap. Single-node clusters of the transactional engines have no redundancy and are suited only for development.

**Node sizing** -- The `sizeSlug` field sets CPU and memory per node; changing it later resizes the cluster in place. One interaction to watch: growing `sizeSlug` while an explicit `storageGib` is set can leave the storage value smaller than the new slug's default, which is invalid -- unset `storageGib` when upsizing and let the cluster adopt the new default.

**VPC placement** -- The `vpc` field places the cluster in a private network so database traffic stays off the public internet. It cannot be changed after creation: retrofitting means a new cluster plus a data migration (`backupRestore` gives you the copy), so decide network placement before the first deploy. When omitted, the cluster is reachable at its public hostname.

**Storage and growth** -- `storageGib` sets custom disk space beyond the size slug's default (increase-only), and `storageAutoscale` grows it automatically when usage crosses a threshold. Autoscale is applied by the Terraform provisioner; the Pulumi bridge does not support it yet and rejects it loudly rather than dropping it.

**Engine tuning** -- `sqlMode` applies only to MySQL and `evictionPolicy` only to Redis/Valkey; the manifest is rejected at validation time if they are paired with any other engine, mirroring DigitalOcean's own rules. Removing a previously set `evictionPolicy` resets the cluster to `noeviction` -- it does not keep the last value.

**Restore provisioning** -- `backupRestore.databaseName` names the SOURCE cluster to restore from; omit `backupCreatedAt` to take the newest backup. The block acts only at creation and DigitalOcean never reports it back -- it is provisioning input, not ongoing configuration.

## Outputs and Dependencies

### What This Component Consumes

| Dependency | Field | ValueFromRef Path |
|------------|-------|-------------------|
| **DigitalOceanVpc** (optional) | `vpc` | `status.outputs.vpc_id` |

### What This Component Provides

After provisioning, `status.outputs` contains values that downstream Cloud Resources can consume via ValueFromRef:

| Output | Description | Common Downstream Use |
|--------|-------------|----------------------|
| `cluster_id` | Unique database cluster identifier (UUID) | Every database satellite kind -- logical databases, users, connection pools, firewall rules, replicas, Kafka topics and schemas -- plus monitor alerts |
| `connection_uri` | Full public connection URI including credentials and database name | Application database configuration (contains the password -- treat as a secret) |
| `host` | Public hostname at which the cluster is accessible | Application connection strings |
| `port` | Network port the cluster listens on | Application connection strings |
| `database_user` | Default database user name | Application authentication |
| `database_password` | Default database user password | Application authentication (store as secret) |
| `private_host` | Private-network hostname (same-VPC access) | In-VPC application connection strings -- prefer over `host` when the app shares the VPC |
| `private_uri` | Private-network connection URI including credentials | In-VPC application configuration |
| `database_name` | Name of the default database | Application configuration |
| `ui_host` / `ui_port` / `ui_uri` / `ui_database` / `ui_user` / `ui_password` | OpenSearch Dashboards connection details (OpenSearch clusters only) | Dashboards access |

## Common Patterns

Browse the [Presets](#presets) tab for ready-to-deploy configurations.

**Production PostgreSQL HA** -- 3-node PostgreSQL 16 cluster with VPC isolation, storage autoscale, and a Sunday maintenance window: automatic failover and secure private access for mission-critical applications. Start from the **Production PostgreSQL HA** preset.

**Development PostgreSQL** -- Single-node PostgreSQL 16 on the smallest instance, no VPC. Minimal cost for development, CI/CD test databases, and staging. Start from the **Development PostgreSQL** preset.

**Valkey cache** -- Single-node Valkey 8 with VPC placement and an LRU eviction policy for low-latency caching, session storage, and pub/sub messaging. Start from the **Valkey Cache** preset.

**Kafka streaming** -- 3-node Kafka 4.2 inside a VPC, the minimum DigitalOcean Kafka topology, for event streaming between services. Start from the **Kafka Cluster** preset.

**OpenSearch analytics** -- Single-node OpenSearch 2.19 with 100 GiB storage for log analytics and full-text search; OpenSearch Dashboards connection details arrive as the `ui_*` outputs. Start from the **OpenSearch** preset.

## Works With

- [**DigitalOcean VPC**](/cloud-catalog/digital-ocean-vpc) -- provides the private network for database cluster placement
- [**DigitalOcean Logical Database**](/cloud-catalog/digital-ocean-database-db) -- additional databases inside the cluster, beyond the default one
- [**DigitalOcean Database User**](/cloud-catalog/digital-ocean-database-user) -- per-application credentials instead of sharing the default user
- [**DigitalOcean Database Connection Pool**](/cloud-catalog/digital-ocean-database-connection-pool) -- server-side connection pooling in front of the cluster
- [**DigitalOcean Database Firewall**](/cloud-catalog/digital-ocean-database-firewall) -- trusted-source rules restricting which resources may connect
- [**DigitalOcean Database Read Replica**](/cloud-catalog/digital-ocean-database-replica) -- read-only copies for read scaling and regional locality
- [**DigitalOcean Kafka Topic**](/cloud-catalog/digital-ocean-database-kafka-topic) / [**DigitalOcean Kafka Schema**](/cloud-catalog/digital-ocean-database-kafka-schema) -- topics and schemas on Kafka clusters
- [**DigitalOcean Monitor Alert**](/cloud-catalog/digital-ocean-monitor-alert) -- alerting on the cluster's resource metrics