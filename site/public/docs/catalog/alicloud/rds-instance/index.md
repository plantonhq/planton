---
title: "RDS Instance"
description: "RDS Instance deployment documentation"
icon: "package"
order: 100
componentName: "alicloudrdsinstance"
---

# AliCloud RDS Instance

Deploys an Alibaba Cloud RDS (Relational Database Service) instance with bundled databases, accounts, and account privileges. Supports MySQL, PostgreSQL, SQL Server, MariaDB, and PPAS engines through a single component type.

## What Gets Created

When you deploy an AliCloudRdsInstance resource, OpenMCF provisions:

- **RDS Instance** -- an `alicloud_db_instance` with the selected engine, instance class, and storage
- **Databases** -- one `alicloud_db_database` per entry in the databases list
- **Accounts** -- one `alicloud_rds_account` per entry in the accounts list
- **Account Privileges** -- one `alicloud_db_account_privilege` per privilege entry, granting specific access levels on databases

## Prerequisites

- **Alibaba Cloud credentials** configured via environment variables or OpenMCF provider config
- **A VSwitch** -- the RDS instance is placed in a VSwitch (create one with AliCloudVswitch)
- The VSwitch's VPC and availability zone determine the instance's network placement

## Quick Start

Create a file `rds-instance.yaml`:

```yaml
apiVersion: alicloud.openmcf.org/v1
kind: AliCloudRdsInstance
metadata:
  name: my-mysql
spec:
  region: cn-hangzhou
  engine: MySQL
  engineVersion: "8.0"
  instanceType: rds.mysql.s2.large
  instanceStorage: 50
  vswitchId:
    valueFrom:
      name: my-db-vswitch
  databases:
    - name: myapp
  accounts:
    - accountName: app_user
      accountPassword: "${DB_PASSWORD}"
      privileges:
        - databaseNames: [myapp]
          privilege: ReadWrite
```

Deploy:

```shell
openmcf apply -f rds-instance.yaml
```

This creates a MySQL 8.0 HA instance with one database, one account, and ReadWrite access.

## Configuration Reference

### Required Fields

| Field | Type | Description | Validation |
|-------|------|-------------|------------|
| `region` | string | Alibaba Cloud region (e.g., `cn-hangzhou`) | Required; non-empty |
| `engine` | string | Database engine | Required; one of: MySQL, PostgreSQL, SQLServer, MariaDB, PPAS |
| `engineVersion` | string | Engine version (e.g., `8.0`, `16.0`) | Required; non-empty |
| `instanceType` | string | Instance class (e.g., `rds.mysql.s2.large`) | Required; non-empty |
| `instanceStorage` | int32 | Storage size in GB | Required; > 0 |
| `vswitchId` | StringValueOrRef | VSwitch ID. Can reference AliCloudVswitch via `valueFrom`. | Required |

### Optional Fields

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `instanceName` | string | metadata.name | Instance display name (2-256 chars) |
| `instanceChargeType` | string | `Postpaid` | Billing: `Postpaid` or `Prepaid` |
| `category` | string | `HighAvailability` | Architecture: `Basic`, `HighAvailability`, `AlwaysOn`, `Finance`, `cluster` |
| `dbInstanceStorageType` | string | | Storage type: `local_ssd`, `cloud_ssd`, `cloud_essd`, `cloud_essd2`, `cloud_essd3` |
| `zoneId` | string | | Primary availability zone |
| `zoneIdSlaveA` | string | | Standby availability zone (for HA) |
| `securityIps` | list | | IP whitelist for access control |
| `securityGroupIds` | list | | VPC security group IDs |
| `monitoringPeriod` | int32 | | Monitoring interval: 5, 10, 60, 300 seconds |
| `maintainTime` | string | | Maintenance window (e.g., `02:00Z-06:00Z`) |
| `deletionProtection` | bool | `false` | Prevent accidental deletion |
| `sslAction` | string | | SSL: `Open` or `Close` |
| `tdeStatus` | string | | TDE: `Enabled` or `Disabled` |
| `encryptionKey` | string | | KMS key ID for disk encryption |
| `autoRenew` | bool | `false` | Auto-renewal for Prepaid |
| `autoRenewPeriod` | int32 | | Auto-renewal period in months (1-12) |
| `period` | int32 | | Subscription period in months |
| `resourceGroupId` | string | | Resource group for organizational grouping |
| `tags` | map | | Key-value tags |
| `parameters` | list | | Database engine parameter overrides |
| `databases` | list | | Databases to create (see below) |
| `accounts` | list | | Accounts to create (see below) |

### Database Fields

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `name` | string | | Database name (required) |
| `characterSet` | string | engine default | Character set (e.g., `utf8mb4`, `UTF8`) |
| `description` | string | | Database description |

### Account Fields

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `accountName` | string | | Login name (required) |
| `accountPassword` | string | | Password (required; 8+ chars) |
| `accountType` | string | `Normal` | `Normal` or `Super` |
| `accountDescription` | string | | Account description |
| `privileges` | list | | Database privileges (see below) |

### Privilege Fields

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `databaseNames` | list | | Databases to grant access to (required; min 1) |
| `privilege` | string | `ReadOnly` | `ReadOnly`, `ReadWrite`, `DDLOnly`, `DMLOnly`, `DBOwner` |

## Stack Outputs

After deployment, the following outputs are available in `status.outputs`:

| Output | Type | Description |
|--------|------|-------------|
| `instance_id` | string | RDS instance ID (e.g., `rm-xxxxx`) |
| `connection_string` | string | Intranet (VPC-internal) connection endpoint |
| `port` | string | Database service port |
| `database_ids` | map | Map of database names to their IDs |

## Related Components

- **AliCloudVswitch** -- VSwitch where the RDS instance is placed
- **AliCloudVpc** -- VPC that provides network isolation
- **AliCloudSecurityGroup** -- Network security rules for instance access
- **AliCloudKmsKey** -- Customer-managed key for disk/TDE encryption
- **AliCloudPrivateDnsZone** -- Private DNS resolution for the instance endpoint
