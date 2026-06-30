# ScalewayRdbInstance

A managed PostgreSQL or MySQL database instance on Scaleway, bundled with databases, users, privileges, and network access control.

## Overview

`ScalewayRdbInstance` provisions a fully managed relational database on Scaleway's RDB service with everything needed for applications to connect immediately. Instead of creating 5 separate Terraform resources, this single resource kind delivers a ready-to-use database.

## Bundled Resources

This composite resource creates the following Scaleway resources:

| # | Terraform Resource | Purpose |
|---|-------------------|---------|
| 1 | `scaleway_rdb_instance` | Managed database engine with admin user |
| 2 | `scaleway_rdb_database` | Logical databases (one per entry in `databases`) |
| 3 | `scaleway_rdb_user` | Additional database users (one per entry in `users`) |
| 4 | `scaleway_rdb_privilege` | User-database permission grants (one per user privilege entry) |
| 5 | `scaleway_rdb_acl` | Network ACL rules (single resource replacing all rules) |

## Features

- **PostgreSQL and MySQL** -- Supports `PostgreSQL-14`, `PostgreSQL-15`, `PostgreSQL-16`, and `MySQL-8`
- **High Availability** -- Optional multi-node HA with automatic failover
- **Private Network** -- Attach to a `ScalewayPrivateNetwork` for secure private connectivity
- **Databases and Users** -- Create application databases and users with specific permission grants
- **Network ACL** -- Restrict which IPs can connect to the public endpoint
- **Encryption at Rest** -- Optional storage encryption for compliance workloads
- **Backup Configuration** -- Configurable backup frequency and retention
- **Engine Tuning** -- Pass engine-specific settings (e.g., `work_mem`, `max_connections`)

## Quick Start

### Development PostgreSQL

```yaml
apiVersion: scaleway.planton.dev/v1
kind: ScalewayRdbInstance
metadata:
  name: dev-postgres
spec:
  region: fr-par
  engine: PostgreSQL-16
  nodeType: DB-DEV-S
  adminUser: admin
  adminPassword: my-dev-password-123
  databases:
    - name: myapp
  users:
    - name: app_user
      password: app-user-password-123
      privileges:
        - databaseName: myapp
          permission: readwrite
```

### Production PostgreSQL with HA and Private Network

```yaml
apiVersion: scaleway.planton.dev/v1
kind: ScalewayRdbInstance
metadata:
  name: prod-postgres
  org: mycompany
  env: production
spec:
  region: fr-par
  engine: PostgreSQL-16
  nodeType: db-gp-xs
  isHaCluster: true
  volumeType: bssd
  volumeSizeInGb: 100
  encryptionAtRest: true
  privateNetworkId:
    value: "xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx"
  aclRules:
    - ip: "10.0.0.0/8"
      description: "Private network range"
  adminUser: admin
  adminPassword: strong-random-password-here
  databases:
    - name: appdb
    - name: analytics
  users:
    - name: app_rw
      password: app-readwrite-password
      privileges:
        - databaseName: appdb
          permission: readwrite
    - name: analytics_ro
      password: analytics-readonly-password
      privileges:
        - databaseName: analytics
          permission: all
        - databaseName: appdb
          permission: readonly
```

## Dependencies

| Dependency | Type | Required | Field |
|-----------|------|----------|-------|
| `ScalewayPrivateNetwork` | `StringValueOrRef` | No | `privateNetworkId` |

## Stack Outputs

| Output | Description | Downstream Use |
|--------|-------------|----------------|
| `instance_id` | Regional ID of the RDB instance | Read replicas, monitoring |
| `endpoint_ip` | Public endpoint IP address | External connections |
| `endpoint_port` | Public endpoint port | External connections |
| `private_endpoint_ip` | Private Network endpoint IP | Application connections |
| `private_endpoint_port` | Private Network endpoint port | Application connections |
| `certificate` | TLS CA certificate (PEM) | Secure client connections |

## Configuration Reference

### Core Fields

| Field | Type | Required | Default | Description |
|-------|------|----------|---------|-------------|
| `region` | string | Yes | - | Scaleway region (e.g., "fr-par") |
| `engine` | string | Yes | - | Engine-Version (e.g., "PostgreSQL-16") |
| `nodeType` | string | Yes | - | Instance size (e.g., "DB-DEV-S") |
| `adminUser` | string | Yes | - | Initial admin username |
| `adminPassword` | string | Yes | - | Initial admin password (min 8 chars) |

### Optional Fields

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `privateNetworkId` | StringValueOrRef | - | Private Network for private connectivity |
| `isHaCluster` | bool | false | Enable multi-node HA |
| `volumeType` | string | "lssd" | Storage type: lssd, bssd, sbs_15k |
| `volumeSizeInGb` | uint32 | auto | Custom volume size (only increases) |
| `disableBackup` | bool | false | Disable automated backups |
| `backupScheduleFrequencyHours` | uint32 | 24 | Hours between backups |
| `backupScheduleRetentionDays` | uint32 | 7 | Days to retain backups |
| `encryptionAtRest` | bool | false | Enable storage encryption |
| `aclRules` | list | [] | Network access control rules |
| `databases` | list | [] | Databases to create |
| `users` | list | [] | Users with optional privileges |
| `settings` | map | {} | Engine runtime settings |
| `initSettings` | map | {} | Engine init-only settings |

### Permission Levels

| Permission | SQL Access | Use Case |
|-----------|-----------|----------|
| `readonly` | SELECT | Reporting, analytics, read replicas |
| `readwrite` | SELECT, INSERT, UPDATE, DELETE | Application accounts |
| `all` | Full DDL + DML | Migration tools, admin accounts |
| `none` | No access | Explicitly revoke |

### Node Types

| Type | vCPU | RAM | Use Case |
|------|------|-----|----------|
| DB-DEV-S | 2 | 2 GB | Development |
| DB-DEV-M | 4 | 4 GB | Development |
| db-gp-xs | 4 | 16 GB | Entry production |
| db-gp-s | 8 | 32 GB | Production |
| db-gp-m | 16 | 64 GB | High-traffic production |

## Best Practices

### Production Checklist

- Enable HA (`isHaCluster: true`) for zero-downtime failover
- Use Private Network for application connections
- Set ACL rules to restrict public endpoint access
- Enable encryption at rest for compliance
- Use non-admin users for application connections with minimal permissions
- Configure backup retention appropriate for your RPO

### Security

- Always set ACL rules when the database has a public endpoint
- Use `readwrite` (not `all`) for application users
- Create separate users for different services/applications
- The admin user should only be used for database administration

## Scaleway Documentation

- [RDB Overview](https://www.scaleway.com/en/docs/managed-databases/postgresql-and-mysql/)
- [Node Types](https://www.scaleway.com/en/pricing/?tags=databases)
- [Engine Versions](https://www.scaleway.com/en/docs/managed-databases/postgresql-and-mysql/reference-content/database-engine-version-policy/)
- [Private Network Integration](https://www.scaleway.com/en/docs/managed-databases/postgresql-and-mysql/how-to/connect-database-private-network/)
