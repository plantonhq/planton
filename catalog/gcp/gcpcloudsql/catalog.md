# GCP Cloud SQL

Deploys a fully managed MySQL, PostgreSQL, or SQL Server database instance on Google Cloud SQL — with explicit connectivity (private VPC IP, public IPv4, and/or Private Service Connect), regional high availability with automatic failover, automated backups with point-in-time recovery, read replicas, customer-managed encryption, and engine tuning flags. The GCP project, the VPC network, the CMEK key, and the replication primary all wire by ValueFromRef, so the instance composes into an InfraChart without hand-copied identifiers.

## What Gets Created

When you deploy this Infra Component, the IaC module provisions:

- **Cloud SQL Database Instance** -- a managed `google_sql_database_instance` in the specified GCP project and region, configured with the chosen engine (MySQL, PostgreSQL, or SQL Server), machine tier, edition, and data disk
- **Connectivity** -- public IPv4 (optionally restricted by authorized-network CIDR rules), private IP inside a VPC (requires private services access on the network), and/or Private Service Connect exposure; TLS posture and server CA mode per your spec
- **High Availability** -- `availabilityType: REGIONAL` provisions a synchronous standby in a second zone with automatic failover
- **Backups & Recovery** -- daily automated backups with retention policies, plus point-in-time recovery (transaction logs on PostgreSQL/SQL Server, binary logs on MySQL)
- **Read Replica wiring** -- when `masterInstanceName` is set, the instance is created as a read replica of the named primary
- **Observability & Performance** -- Query Insights telemetry, managed connection pooling, and engine-specific database flags
- **Security posture** -- root password bootstrap, password validation policy, customer-managed encryption (CMEK), and the deletion guards
- **SQL Server surface** -- collation, time zone, threads-per-core, Managed AD domain join, and audit-to-GCS when the engine is SQL Server

## Before You Deploy

### Planton Setup

- **GCP Provider Connection** -- an active connection in the Connect module with credentials for the target GCP project. Map it as the default for your environment, or specify it explicitly when creating the Infra Component.
- **Planton Runner** -- required when using Runner-based credential delivery. Not needed for inline credentials or browser OAuth authentication modes.

### GCP Project

- **A GCP project** where the Cloud SQL instance will be created. Provide the project ID directly or reference a GcpProject Infra Component via ValueFromRef.
- **Private Services Access** (if using private IP) -- the VPC network must already carry a service networking connection: compose a [GcpGlobalAddress](/infra-catalog/gcp-global-address) with `purpose: VPC_PEERING` and a [GcpServiceNetworkingConnection](/infra-catalog/gcp-service-networking-connection) on the network. This is a one-time setup per VPC; instance creation fails without it.
- **Cloud SQL Admin API** enabled in the target project (the IaC module enables it).

## Deploy

### Console

Open the deployment store, find **GCP Cloud SQL**, and click **Deploy**. The creation wizard walks the decisions in the order a database owner thinks: engine → identity → fresh-or-replica → machine → connectivity → availability → backups → maintenance → observability → security (with a SQL Server-only options step when that engine is chosen). The [Presets](#presets) tab offers three starting configurations: **Production PostgreSQL (Private IP)**, **High-Availability MySQL (Auth Proxy Access)**, and **PostgreSQL Read Replica**.

### CLI

Create a manifest and apply it:

```yaml
apiVersion: gcp.planton.dev/v1alpha1
kind: GcpCloudSql
metadata:
  name: app-database
  org: acme-corp
  env: prod
spec:
  projectId:
    value: "acme-prod-12345"
  instanceName: app-database-prod
  region: us-central1
  databaseEngine: POSTGRESQL
  databaseVersion: "POSTGRES_16"
  tier: "db-custom-2-7680"
  disk:
    sizeGb: 50
  backup:
    enabled: true
    pointInTimeRecoveryEnabled: true
```

```shell
planton apply -f cloud-sql.yaml
```

This creates a PostgreSQL 16 instance with GCP's safe connectivity default (public IPv4 reachable only through the Auth Proxy / connectors), daily backups, and point-in-time recovery. An Infra Job tracks the provisioning in real time.

### InfraChart

When deploying as part of a multi-resource environment, use ValueFromRef to wire the Cloud SQL instance to a GCP project and VPC deployed in the same InfraPipeline:

```yaml
spec:
  projectId:
    valueFrom:
      kind: GcpProject
      name: production-project
      fieldPath: status.outputs.project_id
  network:
    privateNetwork:
      valueFrom:
        kind: GcpVpcNetwork
        name: production-vpc
        fieldPath: status.outputs.network_id
```

The InfraPipeline resolves the dependency graph, deploys the project and VPC first, then provisions the Cloud SQL instance with private IP connectivity.

## Key Configuration

These are the most important decisions when configuring a Cloud SQL instance. Explore the full field reference in the [API Explorer](#api-explorer) tab.

**Database engine and version** -- `databaseEngine` is `MYSQL`, `POSTGRESQL`, or `SQLSERVER`; `databaseVersion` carries the engine-prefixed version string (e.g. `"POSTGRES_16"`, `"MYSQL_8_0"`, `"SQLSERVER_2022_STANDARD"`). The engine is immutable; in-place major version upgrades ARE supported (take a backup first). SQL Server additionally requires `rootPassword`.

**Connectivity** -- Setting `network.privateNetwork` (a VPC self-link or a GcpVpcNetwork reference) is what ENABLES private IP. `network.ipv4Enabled` adds a public address — safe with an empty `authorizedNetworks` list (Auth Proxy only). `network.psc.enabled` exposes the instance over Private Service Connect instead of peering. At least one path must be enabled when the network block is present; omitting the block entirely gives the safe public-plus-proxy default.

**High availability** -- `availabilityType: REGIONAL` runs a standby in a second zone with automatic failover (roughly double cost). Requires `backup.enabled: true` — and on MySQL, `backup.binaryLogEnabled: true`.

**Backups and recovery** -- `backup.enabled` with `startTime` (UTC `"HH:MM"`), `retainedBackups` (count of daily backups kept), and `transactionLogRetentionDays` (how far point-in-time recovery reaches, 1-35). PITR is `pointInTimeRecoveryEnabled` on PostgreSQL/SQL Server and `binaryLogEnabled` on MySQL.

**Read replicas and read pools** -- Set `masterInstanceName` (a name or a GcpCloudSql reference) to create this instance as a read replica. Immutable; the primary must have backups (and binary logs on MySQL) enabled. `replicaConfiguration` tunes failover/cascading and carries the replication channel for external sources. Promote a replica to a standalone primary with `instanceType: CLOUD_SQL_INSTANCE` (clearing `masterInstanceName` in the same change). `instanceType: READ_POOL_INSTANCE` turns the resource into a read pool — a single endpoint over `nodeCount` read-serving nodes, optionally autoscaled between bounds with `readPoolAutoScale`.

**Restore provenance** -- An instance can be born from existing data: `clone` copies another instance (optionally at a point in time), `restoreBackupContext` restores a specific backup run, and `pointInTimeRestoreContext` / `backupdrBackup` restore from Backup and DR. These are triggers as much as configuration — adding or changing them on an existing instance runs the restore.

**Instance sizing** -- `tier` sets CPU and memory (`"db-custom-4-15360"` = 4 vCPU / 15 GB; `"db-f1-micro"` for shared-core dev). The `disk` block sets type (`PD_SSD` default), size (10-65536 GB, grows but never shrinks), auto-resize, the auto-resize limit, and — on `HYPERDISK_BALANCED` — provisioned IOPS and throughput decoupled from size.

**Deletion guards** -- `deletionProtection` makes the IaC engines refuse a destroy; `deletionProtectionEnabled` makes GCP itself reject deletion from every surface; `retainBackupsOnDelete` keeps backups alive after deletion; `finalBackup` takes a last backup during the delete itself. Set the guards on production, and use `deletionPolicy: ABANDON` when an instance must leave IaC management without being destroyed.

## Outputs and Dependencies

### What This Kind Consumes

| Dependency | Field | ValueFromRef Path |
|------------|-------|-------------------|
| **GcpProject** | `projectId` | `status.outputs.project_id` |
| **GcpVpcNetwork** (if private IP) | `network.privateNetwork` | `status.outputs.network_id` |
| **GcpKmsKey** (if CMEK) | `encryptionKeyName` | `status.outputs.key_id` |
| **GcpCloudSql** (if read replica) | `masterInstanceName` | `status.outputs.instance_name` |

### What This Kind Provides

After provisioning, `status.outputs` contains values that downstream Infra Components can consume via ValueFromRef:

| Output | Description | Common Downstream Use |
|--------|-------------|----------------------|
| `instance_name` | Cloud SQL instance name | GcpCloudSqlDatabase / GcpCloudSqlUser `instance` references, read replicas |
| `connection_name` | Full connection identity (`project:region:instance`) | Cloud SQL Auth Proxy, language connectors, Cloud Run Cloud SQL volumes |
| `private_ip` | Private IP address (when private IP is enabled) | In-VPC application connection strings |
| `public_ip` | Public IPv4 address (when enabled) | Auth Proxy target, authorized-network rules |
| `self_link` | GCP resource self link | IAM bindings, audit log filters |
| `service_account_email` | The instance's service account | GCS bucket grants (SQL Server audit), CMEK key access |
| `dns_name` | DNS name (PSC instances) | Consumer application endpoints |
| `psc_service_attachment_link` | PSC service attachment | Consumer-side PSC endpoint wiring |

## Common Patterns

Browse the [Presets](#presets) tab for ready-to-deploy configurations.

**Production PostgreSQL (private IP)** -- Private-IP-only connectivity inside your VPC, regional high availability, point-in-time recovery, Query Insights, and both delete guards. Start from the **Production PostgreSQL (Private IP)** preset.

**MySQL high availability** -- The MySQL equivalent: regional failover with binary logging (MySQL's PITR and replica prerequisite). Start from the **High-Availability MySQL (Auth Proxy Access)** preset.

**Read replica** -- Scale reads by attaching a replica to an existing primary via `masterInstanceName`. Start from the **PostgreSQL Read Replica** preset.

## Works With

- [**GCP Project**](/infra-catalog/gcp-project) -- provides the GCP project where the Cloud SQL instance is created
- [**GCP VPC Network**](/infra-catalog/gcp-vpc-network) -- provides the VPC network for private IP connectivity (requires private services access)
- [**GCP Service Networking Connection**](/infra-catalog/gcp-service-networking-connection) -- the private services access peering that private IP depends on
- [**GCP KMS Key**](/infra-catalog/gcp-kms-key) -- customer-managed encryption for the instance's storage
- [**GCP Cloud SQL Database**](/infra-catalog/gcp-cloud-sql-database) -- create application databases on this instance
- [**GCP Cloud SQL User**](/infra-catalog/gcp-cloud-sql-user) -- create per-application users instead of sharing the admin user
- [**GCP Cloud Run**](/infra-catalog/gcp-cloud-run) -- mounts this instance via Cloud SQL volumes using the connection name output
