# AlicloudPolardbCluster -- Research Documentation

## Provider Resources

This component maps to the following cloud provider resources:

### Terraform

| Resource | Purpose |
|----------|---------|
| `alicloud_polardb_cluster` | The PolarDB cluster itself |
| `alicloud_polardb_database` | Databases within the cluster |
| `alicloud_polardb_account` | Database user accounts |
| `alicloud_polardb_account_privilege` | Privilege grants linking accounts to databases |

### Pulumi

| Resource | Purpose |
|----------|---------|
| `polardb.Cluster` | The PolarDB cluster itself |
| `polardb.Database` | Databases within the cluster |
| `polardb.Account` | Database user accounts |
| `polardb.AccountPrivilege` | Privilege grants linking accounts to databases |

## Architecture

PolarDB uses a shared-storage architecture where compute nodes (primary + read replicas) share a single distributed storage layer. This enables:

- **Fast read scaling** -- add read replicas in minutes (no data copy)
- **Auto-scaling storage** -- Enterprise Edition storage grows automatically
- **Fast failover** -- sub-minute failover since replicas share the same storage
- **Consistent reads** -- all nodes read from the same physical data

### Endpoint Model

PolarDB automatically creates two built-in endpoints:

1. **Primary Endpoint** -- routes all traffic to the primary (read-write) node
2. **Cluster Endpoint** -- distributes read traffic across all nodes, sends writes to primary

This component exposes the primary endpoint's connection string. Custom endpoints (for advanced read routing, connection pooling, etc.) are managed via separate `alicloud_polardb_endpoint` resources.

## Edition Differences

### Enterprise Edition (Normal)

- Shared distributed storage (PSL4/PSL5)
- Auto-scaling storage (no `storage_space` needed)
- Compute-storage separation
- Sub-minute failover
- Recommended for production workloads

### Standard Edition (SENormal)

- Local ESSD storage (ESSDPL0-3, ESSDAUTOPL)
- Pre-allocated storage via `storage_space`
- Lower cost, higher I/O performance for specific workloads
- No auto-scaling storage

### Basic Edition (Basic)

- Single node (no HA)
- Suitable for development and testing only
- Lowest cost

## Design Decisions

### DD06: PolarDB as First-Class Component

PolarDB is a separate component from AlicloudRdsInstance because:

- Different Terraform resources (`alicloud_polardb_*` vs `alicloud_db_*`)
- Different architecture (cluster vs instance, nodes vs HA categories)
- Different pricing model (per-node billing)
- Different endpoint management
- Users make an explicit architectural choice between RDS and PolarDB

### DD07: Composite Bundling

The cluster, databases, accounts, and privileges are bundled because a PolarDB cluster without any databases or accounts is not useful. The bundled flow ensures correct dependency ordering.

### Fields Excluded from Spec

The following fields were deliberately excluded to keep the component within the 80/20 scope:

- **Serverless configuration** -- specialized use case with many interrelated fields
- **Proxy configuration** -- advanced DBA feature
- **Hot standby / hot replica** -- advanced HA topology
- **Custom endpoints** -- can be managed separately
- **Global Database Network** -- multi-region replication
- **X-Engine, IMCI** -- MySQL-specific engine extensions

Users needing these features can use the `parameters` field for engine settings or manage the advanced resources separately.

## Character Set Defaults

| Engine | Default | Common Alternatives |
|--------|---------|---------------------|
| MySQL | utf8 | utf8mb4, gbk, latin1 |
| PostgreSQL | UTF8 | SQL_ASCII, EUC_CN |
| Oracle | UTF8 | GBK |

Note: PolarDB MySQL defaults to `utf8` (not `utf8mb4` like RDS). Users who need full Unicode support (including emojis) should explicitly set `characterSetName: utf8mb4`.

## Privilege Levels

| Privilege | MySQL | PostgreSQL | Oracle |
|-----------|-------|------------|--------|
| ReadOnly | SELECT | SELECT | SELECT |
| ReadWrite | SELECT, INSERT, UPDATE, DELETE | SELECT, INSERT, UPDATE, DELETE, TRUNCATE, REFERENCES, TRIGGER | SELECT, INSERT, UPDATE, DELETE |
| DDLOnly | CREATE, DROP, ALTER | CREATE, CONNECT, TEMPORARY, EXECUTE | CREATE, ALTER, DROP |
| DMLOnly | INSERT, UPDATE, DELETE | INSERT, UPDATE, DELETE | INSERT, UPDATE, DELETE |
