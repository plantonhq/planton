# Scaleway RDB Instance

Deploys a Scaleway Managed Database instance with bundled logical databases, users with per-database privileges, and network ACL rules. Supports PostgreSQL and MySQL engines with optional high availability, Private Network attachment, encryption at rest, and automated backup configuration.

## What Gets Created

When you deploy a ScalewayRdbInstance resource, Planton provisions:

- **RDB Instance** — a `databases.Instance` resource providing a fully managed database engine (PostgreSQL or MySQL) with the specified node type, volume configuration, and admin user
- **Private Network Endpoint** — created only when `privateNetworkId` is set, attaches the instance to a Private Network with IPAM-based IP assignment
- **Logical Databases** — one `databases.Database` resource per entry in the `databases` list
- **Database Users** — one `databases.User` resource per entry in the `users` list
- **User Privileges** — one `databases.Privilege` resource per privilege entry, linking a user to a database with a specific permission level
- **Network ACL** — a `databases.Acl` resource created only when `aclRules` is non-empty, controlling which CIDR ranges can reach the public endpoint

## Prerequisites

- **Scaleway credentials** configured via environment variables or Planton provider config
- **A valid engine string** in the format `"{Engine}-{MajorVersion}"` (e.g., `"PostgreSQL-16"`, `"MySQL-8"`)
- **A Private Network** in the target region if using private connectivity (can be created via a ScalewayPrivateNetwork resource)

## Quick Start

Create a file `rdb-instance.yaml`:

```yaml
apiVersion: scaleway.planton.dev/v1
kind: ScalewayRdbInstance
metadata:
  name: my-db
  labels:
    planton.dev/provisioner: pulumi
    pulumi.planton.dev/organization: my-org
    pulumi.planton.dev/project: my-project
    pulumi.planton.dev/stack.name: dev.ScalewayRdbInstance.my-db
spec:
  region: fr-par
  engine: PostgreSQL-16
  nodeType: DB-DEV-S
  adminUser: admin
  adminPassword: change-me-strong-pw
```

Deploy:

```shell
planton apply -f rdb-instance.yaml
```

This creates a single-node PostgreSQL 16 instance with local SSD storage, automated backups enabled, and a public endpoint accessible to all IPs (no ACL rules configured).

## Configuration Reference

### Required Fields

| Field | Type | Description | Validation |
|-------|------|-------------|------------|
| `region` | `string` | Scaleway region for the instance (e.g., `"fr-par"`, `"nl-ams"`, `"pl-waw"`). Cannot be changed after creation. | Required |
| `engine` | `string` | Database engine and major version (e.g., `"PostgreSQL-16"`, `"MySQL-8"`). Cannot be changed after creation. | Required, pattern: `^(PostgreSQL\|MySQL)-[0-9]+$` |
| `nodeType` | `string` | Instance type determining CPU and RAM (e.g., `"DB-DEV-S"`, `"db-gp-xs"`, `"db-gp-m"`). Can be changed after creation. | Required |
| `adminUser` | `string` | Username for the initial admin user. Must differ from any user in the `users` list. | Required, max 63 characters |
| `adminPassword` | `string` | Password for the admin user. | Required, min 8 characters |

### Optional Fields

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `privateNetworkId` | `StringValueOrRef` | — | Private Network UUID for private connectivity. Enables IPAM-based IP assignment. Can reference a ScalewayPrivateNetwork resource via `valueFrom`. |
| `isHaCluster` | `bool` | `false` | When `true`, deploys a multi-node HA cluster with automatic failover. Doubles cost. |
| `volumeType` | `string` | `"lssd"` | Storage volume type. Options: `"lssd"` (local SSD, lowest latency), `"bssd"` (block SSD, 5K IOPS), `"sbs_15k"` (block SSD, 15K IOPS). Cannot be changed after creation. |
| `volumeSizeInGb` | `uint32` | — | Volume size in GB. If omitted, uses the node type default. Can only be increased, never decreased. |
| `disableBackup` | `bool` | `false` | When `true`, disables automated backups. |
| `backupScheduleFrequencyHours` | `uint32` | `24` | Hours between automated backups (1–24). Lower values provide finer RPO. |
| `backupScheduleRetentionDays` | `uint32` | `7` | Days to retain automated backups (1–365). |
| `encryptionAtRest` | `bool` | `false` | When `true`, encrypts all data written to disk. |
| `aclRules` | `object[]` | `[]` | Network access control rules for the public endpoint. If empty, no ACL is created (Scaleway defaults to allowing all IPs). |
| `aclRules[].ip` | `string` | — | CIDR range to allow (e.g., `"10.0.0.0/24"`, `"1.2.3.4/32"`). Required per rule. |
| `aclRules[].description` | `string` | `""` | Human-readable label for the rule (e.g., `"Office IP"`). |
| `databases` | `object[]` | `[]` | Logical databases to create on the instance. |
| `databases[].name` | `string` | — | Database name. Required per entry. Max 63 characters. Reserved names rejected by Scaleway. |
| `users` | `object[]` | `[]` | Additional database users to create. |
| `users[].name` | `string` | — | Username. Required per entry. Max 63 characters. |
| `users[].password` | `string` | — | User password. Required per entry. Min 8 characters. |
| `users[].isAdmin` | `bool` | `false` | When `true`, grants superuser-like access to all databases. |
| `users[].privileges` | `object[]` | `[]` | Per-database permission grants for this user. |
| `users[].privileges[].databaseName` | `string` | — | Target database name. Required per privilege. |
| `users[].privileges[].permission` | `string` | — | Permission level. Options: `"readonly"`, `"readwrite"`, `"all"`, `"none"`. Required per privilege. |
| `settings` | `map<string, string>` | `{}` | Engine-specific runtime settings (e.g., `"max_connections": "200"`). Applied on creation and updates. |
| `initSettings` | `map<string, string>` | `{}` | Engine-specific init-time settings. Cannot be changed after creation (e.g., `"lower_case_table_names": "1"` for MySQL). |

## Examples

### Development PostgreSQL

A minimal PostgreSQL instance for development with a single application database and user:

```yaml
apiVersion: scaleway.planton.dev/v1
kind: ScalewayRdbInstance
metadata:
  name: dev-postgres
  labels:
    planton.dev/provisioner: pulumi
    pulumi.planton.dev/organization: my-org
    pulumi.planton.dev/project: my-project
    pulumi.planton.dev/stack.name: dev.ScalewayRdbInstance.dev-postgres
spec:
  region: fr-par
  engine: PostgreSQL-16
  nodeType: DB-DEV-S
  adminUser: admin
  adminPassword: dev-admin-pw-2024
  databases:
    - name: appdb
  users:
    - name: appuser
      password: app-user-pw-2024
      privileges:
        - databaseName: appdb
          permission: readwrite
```

### Production PostgreSQL with HA and Private Network

A production-grade HA PostgreSQL instance with Private Network connectivity, encryption, ACL rules, and tuned engine settings:

```yaml
apiVersion: scaleway.planton.dev/v1
kind: ScalewayRdbInstance
metadata:
  name: prod-postgres
  labels:
    planton.dev/provisioner: pulumi
    pulumi.planton.dev/organization: my-org
    pulumi.planton.dev/project: my-project
    pulumi.planton.dev/stack.name: prod.ScalewayRdbInstance.prod-postgres
spec:
  region: fr-par
  engine: PostgreSQL-16
  nodeType: db-gp-xs
  adminUser: pgadmin
  adminPassword: strong-prod-password-2024
  isHaCluster: true
  volumeType: bssd
  volumeSizeInGb: 100
  encryptionAtRest: true
  backupScheduleFrequencyHours: 6
  backupScheduleRetentionDays: 30
  privateNetworkId: xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx
  aclRules:
    - ip: 10.0.0.0/16
      description: Internal VPC range
    - ip: 203.0.113.10/32
      description: VPN egress IP
  databases:
    - name: webapp
    - name: analytics
  users:
    - name: webapp_svc
      password: webapp-svc-pw-2024
      privileges:
        - databaseName: webapp
          permission: readwrite
    - name: analytics_ro
      password: analytics-ro-pw-2024
      privileges:
        - databaseName: analytics
          permission: readonly
        - databaseName: webapp
          permission: readonly
  settings:
    max_connections: "200"
    work_mem: "64MB"
    effective_cache_size: "4GB"
```

### MySQL with Private Network Reference

A MySQL instance referencing an Planton-managed Private Network:

```yaml
apiVersion: scaleway.planton.dev/v1
kind: ScalewayRdbInstance
metadata:
  name: mysql-db
  labels:
    planton.dev/provisioner: pulumi
    pulumi.planton.dev/organization: my-org
    pulumi.planton.dev/project: my-project
    pulumi.planton.dev/stack.name: staging.ScalewayRdbInstance.mysql-db
spec:
  region: nl-ams
  engine: MySQL-8
  nodeType: DB-DEV-M
  adminUser: root_admin
  adminPassword: mysql-admin-pw-2024
  privateNetworkId:
    valueFrom:
      kind: ScalewayPrivateNetwork
      name: app-network
      fieldPath: status.outputs.private_network_id
  databases:
    - name: ecommerce
  users:
    - name: shop_app
      password: shop-app-pw-2024
      privileges:
        - databaseName: ecommerce
          permission: readwrite
  initSettings:
    lower_case_table_names: "1"
```

## Stack Outputs

After deployment, the following outputs are available in `status.outputs`:

| Output | Type | Description |
|--------|------|-------------|
| `instance_id` | `string` | Regional ID of the created RDB instance. Referenced by downstream resources (read replicas, monitoring). |
| `endpoint_ip` | `string` | Public endpoint IP address. Subject to ACL rules. |
| `endpoint_port` | `uint32` | Public endpoint port number (typically 5432 for PostgreSQL, 3306 for MySQL). |
| `private_endpoint_ip` | `string` | Private Network endpoint IP address. Empty if no Private Network is attached. |
| `private_endpoint_port` | `uint32` | Private Network endpoint port number. Zero if no Private Network is attached. |
| `certificate` | `string` | TLS certificate in PEM format for verifying the database server's identity. Use with `sslrootcert` (PostgreSQL) or `ssl-ca` (MySQL). |

## Related Components

- [ScalewayPrivateNetwork](/docs/catalog/scaleway/scalewayprivatenetwork) — provides private connectivity between the database and application workloads
- [ScalewayKapsuleCluster](/docs/catalog/scaleway/scalewaykapsulecluster) — deploys Kubernetes clusters whose workloads connect to this database
- [ScalewayInstanceSecurityGroup](/docs/catalog/scaleway/scalewayinstancesecuritygroup) — controls network access for compute instances connecting to the database
