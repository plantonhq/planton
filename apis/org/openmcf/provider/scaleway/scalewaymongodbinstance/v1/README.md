# ScalewayMongodbInstance

A managed MongoDB document database on Scaleway, bundled with application users and role-based access control.

## Overview

`ScalewayMongodbInstance` provisions a fully managed MongoDB instance on Scaleway with everything needed for applications to connect immediately. Instead of separately creating the instance and each user, this single resource kind delivers a ready-to-use document database with pre-configured access.

## Bundled Resources

This composite resource creates the following Scaleway resources:

| # | Terraform Resource | Purpose |
|---|-------------------|---------|
| 1 | `scaleway_mongodb_instance` | Managed MongoDB engine with admin user |
| 2 | `scaleway_mongodb_user` | Additional database users with roles (one per entry in `users`) |

### Why Only 2 Resource Types?

MongoDB's resource model is fundamentally different from relational databases (like ScalewayRdbInstance which bundles 5 types):

- **No database resource**: MongoDB databases are created implicitly when you first write data. There is no `scaleway_mongodb_database` Terraform resource.
- **No privilege resource**: User permissions are expressed as role assignments directly on the user resource, scoped to a database name or all databases.
- **No ACL resource**: MongoDB on Scaleway has no IP-based access control. Network security is controlled entirely by the Private Network / Public Network endpoint choice.

## Features

- **MongoDB 7.x** -- Managed MongoDB with automatic maintenance and patching
- **Replica Set HA** -- 3-node replica set with automatic failover (node_number=3)
- **Private Network** -- Attach to a `ScalewayPrivateNetwork` for private-only connectivity
- **Dual Endpoints** -- Optional public endpoint alongside Private Network for admin access
- **Users and Roles** -- Create application users with database-scoped role assignments
- **Block Storage** -- SBS 5K or 15K IOPS volume types
- **Snapshot Scheduling** -- Configurable automatic snapshot frequency and retention
- **TLS Certificate** -- Always available for secure client connections
- **Engine Settings** -- Pass MongoDB-specific configuration key-value pairs

## Quick Start

### Development Instance (Minimal)

```yaml
apiVersion: scaleway.openmcf.org/v1
kind: ScalewayMongodbInstance
metadata:
  name: dev-mongodb
spec:
  region: fr-par
  version: "7.0.12"
  nodeType: MGDB-PLAY2-NANO
  nodeNumber: 1
  adminUser: admin
  adminPassword: dev-admin-password-123
  users:
    - name: app_user
      password: app-user-password-123
      roles:
        - role: read_write
          databaseName: myapp
```

### Production Replica Set with Private Network

```yaml
apiVersion: scaleway.openmcf.org/v1
kind: ScalewayMongodbInstance
metadata:
  name: prod-mongodb
  org: mycompany
  env: production
spec:
  region: fr-par
  version: "7.0.12"
  nodeType: MGDB-POP2-8C-32G
  nodeNumber: 3
  volumeType: sbs_15k
  volumeSizeInGb: 100
  enableSnapshotSchedule: true
  snapshotScheduleFrequencyHours: 12
  snapshotScheduleRetentionDays: 30
  privateNetworkId:
    value: "xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx"
  adminUser: dbadmin
  adminPassword: very-strong-random-password-here
  users:
    - name: app_service
      password: app-service-password-random
      roles:
        - role: read_write
          databaseName: orders
        - role: read
          databaseName: analytics
    - name: analytics_reader
      password: analytics-reader-password
      roles:
        - role: read
          anyDatabase: true
```

## Dependencies

| Dependency | Type | Required | Field |
|-----------|------|----------|-------|
| `ScalewayPrivateNetwork` | `StringValueOrRef` | No | `privateNetworkId` |

## Stack Outputs

| Output | Description | Downstream Use |
|--------|-------------|----------------|
| `instance_id` | Regional ID of the MongoDB instance | Snapshots, monitoring |
| `public_dns_record` | Public endpoint DNS hostname | External connections |
| `public_port` | Public endpoint port | External connections |
| `private_dns_records` | Private Network endpoint DNS hostnames | Application connections |
| `private_ips` | Private Network endpoint IP addresses | Application connections |
| `private_port` | Private Network endpoint port | Application connections |
| `tls_certificate` | TLS CA certificate (PEM) | Secure client connections |

## Configuration Reference

### Core Fields

| Field | Type | Required | Default | Description |
|-------|------|----------|---------|-------------|
| `region` | string | Yes | - | Scaleway region (currently only "fr-par") |
| `version` | string | Yes | - | MongoDB version (e.g., "7.0.12") |
| `nodeType` | string | Yes | - | Instance size (e.g., "MGDB-PLAY2-NANO") |
| `nodeNumber` | uint32 | Yes | 1 | Number of nodes: 1 (standalone) or 3 (replica set) |
| `adminUser` | string | Yes | - | Initial admin username |
| `adminPassword` | string | Yes | - | Initial admin password (min 8 chars) |

### Optional Fields

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `privateNetworkId` | StringValueOrRef | - | Private Network for private connectivity |
| `enablePublicNetwork` | bool | false | Also create public endpoint when PN is attached |
| `volumeType` | string | "sbs_5k" | Storage type: sbs_5k, sbs_15k |
| `volumeSizeInGb` | uint32 | 5 | Volume size in GB (multiples of 5, only increases) |
| `enableSnapshotSchedule` | bool | false | Enable automatic snapshots |
| `snapshotScheduleFrequencyHours` | uint32 | auto | Hours between snapshots |
| `snapshotScheduleRetentionDays` | uint32 | auto | Days to retain snapshots |
| `users` | list | [] | Users with role assignments |
| `settings` | map | {} | MongoDB engine settings |

### Node Types

#### Shared vCPU (Cost-Optimized)

| Type | vCPUs | RAM | Use Case |
|------|-------|-----|----------|
| MGDB-PLAY2-NANO | 2 | 4 GB | Development (free trial available) |
| MGDB-PRO2-XXS | 2 | 8 GB | Light production |
| MGDB-PRO2-XS | 4 | 16 GB | Small production |
| MGDB-PRO2-S | 8 | 32 GB | Production |
| MGDB-PRO2-M | 16 | 64 GB | High-traffic production |
| MGDB-PRO2-L | 32 | 128 GB | Enterprise |

#### Dedicated vCPU (Production-Optimized)

| Type | vCPUs | RAM | Use Case |
|------|-------|-----|----------|
| MGDB-POP2-2C-8G | 2 | 8 GB | Entry production |
| MGDB-POP2-4C-16G | 4 | 16 GB | Small production |
| MGDB-POP2-8C-32G | 8 | 32 GB | Production |
| MGDB-POP2-16C-64G | 16 | 64 GB | High-traffic production |
| MGDB-POP2-32C-128G | 32 | 128 GB | Enterprise |
| MGDB-POP2-64C-256G | 64 | 256 GB | Large enterprise |

### MongoDB Roles

| Role | Description | Use Case |
|------|-------------|----------|
| `read` | Read-only access (find, count, aggregate) | Reporting, analytics |
| `read_write` | Read and write access (insert, update, delete) | Application accounts |
| `db_admin` | Database administration (indexes, collections) | Migration tools, admin |

### Deployment Modes

| node_number | Mode | Description |
|-------------|------|-------------|
| 1 | Standalone | Single node, no redundancy. Development and testing. |
| 3 | Replica Set | 1 primary + 2 secondaries, automatic failover. Production. |

## Network Security

**Important**: Unlike ScalewayRdbInstance, MongoDB has **no IP-based ACL**. Network security is controlled entirely by endpoint choice:

| Configuration | Security Level | Use Case |
|--------------|---------------|----------|
| Private Network only | Highest | Production (default when PN attached) |
| Private Network + Public | Medium | Production with admin access |
| Public only (no PN) | Lowest | Development, testing |

When `privateNetworkId` is set and `enablePublicNetwork` is false (the default), the instance is private-only -- the most secure configuration. The public endpoint is accessible from **any** IP address with no way to restrict it via ACL.

## Best Practices

### Production Checklist

- Use 3-node replica set (`nodeNumber: 3`) for automatic failover
- Attach to a Private Network for private-only connectivity
- Do **not** enable public network unless you have other network controls
- Enable snapshot scheduling with appropriate frequency and retention
- Use non-admin users for application connections with minimal roles
- Use dedicated vCPU node types (MGDB-POP2-*) for consistent performance

### Security

- Always use Private Network for application connections
- Create separate users per service with the minimum required roles
- Use `read_write` (not `db_admin`) for application users
- The admin user should only be used for database administration
- Use the TLS certificate for encrypted client connections

### MongoDB-Specific Notes

- Databases are created implicitly -- no need to pre-create them
- Role `database_name` references databases that may not exist yet (they will be created on first write)
- `any_database` roles apply to all current and future databases
- Volume size can only be increased, never decreased
- Changing `node_number` between 1 and 3 may destroy and recreate the instance

## Scaleway Documentation

- [Managed MongoDB Overview](https://www.scaleway.com/en/docs/managed-databases/mongodb/)
- [Node Types and Pricing](https://www.scaleway.com/en/pricing/?tags=databases)
- [Private Network Integration](https://www.scaleway.com/en/docs/managed-databases/mongodb/how-to/connect-mongodb-private-network/)
- [MongoDB Connection Guide](https://www.scaleway.com/en/docs/managed-databases/mongodb/how-to/connect-mongodb-instance/)
