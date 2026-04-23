# GCP Cloud SQL

Deploys a Google Cloud SQL instance with configurable database engine (MySQL or PostgreSQL), optional private IP networking, automated backups with point-in-time recovery, and regional high availability. The component provisions a single managed database instance with SSD storage and GCP-managed labels.

## What Gets Created

When you deploy a GcpCloudSql resource, OpenMCF provisions:

- **Cloud SQL Database Instance** — a `google_sql_database_instance` with SSD storage, the specified machine tier, and database version, placed in the target GCP project and region
- **IP Configuration** — public IPv4 connectivity by default, or private IP within a VPC when `network.privateIpEnabled` is `true`
- **Authorized Networks** — created only when `network.authorizedNetworks` contains CIDR entries, allowing direct public IP access from those ranges
- **High Availability (Regional)** — configured only when `highAvailability.enabled` is `true`, promoting the instance to `REGIONAL` availability with automatic failover to the specified secondary zone
- **Automated Backups** — created only when `backup.enabled` is `true`, with configurable retention and optional point-in-time recovery
- **Database Flags** — applied only when `databaseFlags` entries are provided, passing engine-specific tuning parameters to the instance

## Prerequisites

- **GCP credentials** configured via environment variables or OpenMCF provider config
- **A GCP project** where the Cloud SQL instance will be created
- **A VPC network** with Private Services Access configured if enabling private IP connectivity
- **Service Networking API** enabled in the project if using private IP

## Quick Start

Create a file `cloudsql.yaml`:

```yaml
apiVersion: gcp.openmcf.org/v1
kind: GcpCloudSql
metadata:
  name: my-db
  labels:
    openmcf.org/provisioner: pulumi
    pulumi.openmcf.org/organization: my-org
    pulumi.openmcf.org/project: my-project
    pulumi.openmcf.org/stack.name: dev.GcpCloudSql.my-db
spec:
  projectId: my-gcp-project
  region: us-central1
  databaseEngine: POSTGRESQL
  databaseVersion: POSTGRES_15
  tier: db-n1-standard-1
  storageGb: 10
```

Deploy:

```shell
openmcf apply -f cloudsql.yaml
```

This creates a PostgreSQL 15 instance on a `db-n1-standard-1` machine with 10 GB SSD storage, public IP enabled, and no high availability.

## Configuration Reference

### Required Fields

| Field | Type | Description | Validation |
|-------|------|-------------|------------|
| `projectId` | `string` | GCP project ID where the instance is created. Can reference a GcpProject resource via `valueFrom`. | Required |
| `region` | `string` | GCP region for the instance (e.g., `us-central1`). | Pattern: `^[a-z]+-[a-z]+[0-9]$` |
| `databaseEngine` | `enum` | Database engine type. Valid values: `MYSQL`, `POSTGRESQL`. | Must be specified (cannot be `DATABASE_ENGINE_UNSPECIFIED`) |
| `databaseVersion` | `string` | Engine-specific version string (e.g., `MYSQL_8_0`, `POSTGRES_15`). | Minimum length: 1 |
| `tier` | `string` | Machine type for the instance (e.g., `db-n1-standard-1`, `db-custom-2-8192`). | Minimum length: 1 |
| `storageGb` | `int32` | Storage size in gigabytes. | 10–65536. Recommended default: `10` |

### Optional Fields

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `diskAutoResize` | `bool` | `true` | Automatically increase storage when approaching capacity. |
| `edition` | `enum` | `ENTERPRISE` | Cloud SQL edition. `ENTERPRISE`: standard, 99.95% SLA for HA instances. `ENTERPRISE_PLUS`: premium, 99.99% SLA, faster reads. |
| `deletionProtection` | `bool` | `false` | Prevents accidental deletion of the instance when enabled. |
| `queryInsightsEnabled` | `bool` | `false` | Enables Query Insights for performance monitoring and query analysis. |
| `rootPassword` | `string` | — | Initial root password for the instance. Minimum length: 8 characters. |
| `databaseFlags` | `map<string, string>` | `{}` | Engine-specific configuration flags as key-value pairs (e.g., `max_connections: "200"`). |
| `network.vpcId` | `string` | — | VPC network ID for private IP connectivity. Can reference a GcpVpc resource via `valueFrom`. |
| `network.privateIpEnabled` | `bool` | `false` | Enables private IP for the instance. Requires `network.vpcId` to be set. |
| `network.ipv4Enabled` | `bool` | `false` | Enables public IPv4 for the instance. Can be combined with private IP. |
| `network.authorizedNetworks` | `string[]` | `[]` | CIDR blocks allowed to connect via public IP. Must be unique. Pattern: `x.x.x.x/x`. |
| `highAvailability.enabled` | `bool` | `false` | Enables regional high availability with automatic failover. Requires `highAvailability.zone`. |
| `highAvailability.zone` | `string` | — | Secondary zone for HA failover (e.g., `us-central1-b`). Required when `highAvailability.enabled` is `true`. |
| `backup.enabled` | `bool` | `false` | Enables automated daily backups. Requires `backup.startTime` and `backup.retentionDays`. |
| `backup.startTime` | `string` | — | Daily backup window start time in `HH:MM` format (UTC). Required when backup is enabled. |
| `backup.retentionDays` | `int32` | `7` | Number of days to retain backups. Range: 1–365. |
| `backup.pointInTimeRecoveryEnabled` | `bool` | `false` | Enables PITR using transaction logs. Requires `backup.enabled` to be `true`. |
| `maintenanceWindow.day` | `int32` | — | Day of week for maintenance (1=Monday, 7=Sunday). |
| `maintenanceWindow.hour` | `int32` | — | Hour for maintenance start (0–23, UTC). |
| `maintenanceWindow.updateTrack` | `string` | — | Update track: `canary` for early updates, `stable` for production. |

## Examples

### MySQL with Public IP

A MySQL 8.0 instance with public IP access restricted to a single office CIDR:

```yaml
apiVersion: gcp.openmcf.org/v1
kind: GcpCloudSql
metadata:
  name: app-mysql
  labels:
    openmcf.org/provisioner: pulumi
    pulumi.openmcf.org/organization: my-org
    pulumi.openmcf.org/project: my-project
    pulumi.openmcf.org/stack.name: dev.GcpCloudSql.app-mysql
spec:
  projectId: my-gcp-project
  region: us-central1
  databaseEngine: MYSQL
  databaseVersion: MYSQL_8_0
  tier: db-n1-standard-2
  storageGb: 20
  network:
    ipv4Enabled: true
    authorizedNetworks:
      - 203.0.113.0/24
```

### PostgreSQL with Private IP and Backups

A PostgreSQL instance accessible only within a VPC, with daily backups and point-in-time recovery:

```yaml
apiVersion: gcp.openmcf.org/v1
kind: GcpCloudSql
metadata:
  name: api-postgres
  labels:
    openmcf.org/provisioner: pulumi
    pulumi.openmcf.org/organization: my-org
    pulumi.openmcf.org/project: my-project
    pulumi.openmcf.org/stack.name: staging.GcpCloudSql.api-postgres
spec:
  projectId: my-gcp-project
  region: us-central1
  databaseEngine: POSTGRESQL
  databaseVersion: POSTGRES_15
  tier: db-custom-2-8192
  storageGb: 50
  network:
    vpcId: projects/my-gcp-project/global/networks/my-vpc
    privateIpEnabled: true
  backup:
    enabled: true
    startTime: "03:00"
    retentionDays: 14
    pointInTimeRecoveryEnabled: true
```

### Production HA with Full Configuration

A production-grade PostgreSQL instance with high availability, Enterprise Plus edition, backups, scheduled maintenance, and custom database flags:

```yaml
apiVersion: gcp.openmcf.org/v1
kind: GcpCloudSql
metadata:
  name: prod-postgres
  labels:
    openmcf.org/provisioner: pulumi
    pulumi.openmcf.org/organization: my-org
    pulumi.openmcf.org/project: my-project
    pulumi.openmcf.org/stack.name: prod.GcpCloudSql.prod-postgres
spec:
  projectId: my-gcp-project
  region: us-central1
  databaseEngine: POSTGRESQL
  databaseVersion: POSTGRES_15
  tier: db-custom-4-16384
  storageGb: 200
  diskAutoResize: true
  edition: ENTERPRISE_PLUS
  deletionProtection: true
  queryInsightsEnabled: true
  rootPassword: my-secure-root-pw
  network:
    vpcId: projects/my-gcp-project/global/networks/prod-vpc
    privateIpEnabled: true
  highAvailability:
    enabled: true
    zone: us-central1-b
  backup:
    enabled: true
    startTime: "02:00"
    retentionDays: 30
    pointInTimeRecoveryEnabled: true
  maintenanceWindow:
    day: 7
    hour: 4
    updateTrack: stable
  databaseFlags:
    max_connections: "200"
    log_min_duration_statement: "1000"
```

### Using Foreign Key References

Reference other OpenMCF-managed resources instead of hardcoding IDs:

```yaml
apiVersion: gcp.openmcf.org/v1
kind: GcpCloudSql
metadata:
  name: ref-postgres
  labels:
    openmcf.org/provisioner: pulumi
    pulumi.openmcf.org/organization: my-org
    pulumi.openmcf.org/project: my-project
    pulumi.openmcf.org/stack.name: prod.GcpCloudSql.ref-postgres
spec:
  projectId:
    valueFrom:
      kind: GcpProject
      name: my-project
      fieldPath: status.outputs.project_id
  region: us-central1
  databaseEngine: POSTGRESQL
  databaseVersion: POSTGRES_15
  tier: db-n1-standard-2
  storageGb: 50
  network:
    vpcId:
      valueFrom:
        kind: GcpVpc
        name: my-vpc
        fieldPath: status.outputs.network_id
    privateIpEnabled: true
  backup:
    enabled: true
    startTime: "03:00"
    retentionDays: 7
```

## Stack Outputs

After deployment, the following outputs are available in `status.outputs`:

| Output | Type | Description |
|--------|------|-------------|
| `instance_name` | `string` | Name of the Cloud SQL instance |
| `connection_name` | `string` | Full connection name in the format `project:region:instance`, used by Cloud SQL Proxy |
| `private_ip` | `string` | Private IP address of the instance (only set when `network.privateIpEnabled` is `true`) |
| `public_ip` | `string` | Public IP address of the instance |
| `self_link` | `string` | GCP resource self link for the Cloud SQL instance |

## Related Components

- [GcpProject](/docs/catalog/gcp/gcpproject) — provides the GCP project where the instance is created
- [GcpVpc](/docs/catalog/gcp/gcpvpc) — provides the VPC network for private IP connectivity
