# AliCloud PolarDB Cluster

Deploys an Alibaba Cloud PolarDB cluster with bundled databases, accounts, and account privileges. Supports MySQL, PostgreSQL, and Oracle compatibility modes through a single component type.

## What Gets Created

When you deploy an AliCloudPolardbCluster resource, OpenMCF provisions:

- **PolarDB Cluster** -- an `alicloud_polardb_cluster` with the selected engine, node class, and node count
- **Databases** -- one `alicloud_polardb_database` per entry in the databases list
- **Accounts** -- one `alicloud_polardb_account` per entry in the accounts list
- **Account Privileges** -- one `alicloud_polardb_account_privilege` per privilege entry, granting specific access levels on databases

## Prerequisites

- **Alibaba Cloud credentials** configured via environment variables or OpenMCF provider config
- **A VSwitch** -- the PolarDB cluster is placed in a VSwitch (create one with AliCloudVswitch)
- The VSwitch's VPC and availability zone determine the cluster's network placement

## Quick Start

Create a file `polardb-cluster.yaml`:

```yaml
apiVersion: ali-cloud.openmcf.org/v1
kind: AliCloudPolardbCluster
metadata:
  name: my-polardb
spec:
  region: cn-hangzhou
  dbType: MySQL
  dbVersion: "8.0"
  dbNodeClass: polar.mysql.x4.large
  vswitchId:
    valueFrom:
      name: my-db-vswitch
  databases:
    - dbName: myapp
  accounts:
    - accountName: app_user
      accountPassword: "${DB_PASSWORD}"
      privileges:
        - dbNames: [myapp]
          accountPrivilege: ReadWrite
```

Deploy:

```shell
openmcf apply -f polardb-cluster.yaml
```

This creates a MySQL 8.0 PolarDB cluster with 2 nodes (1 primary + 1 read replica), one database, one account, and ReadWrite access.

## Configuration Reference

### Required Fields

| Field | Type | Description | Validation |
|-------|------|-------------|------------|
| `region` | string | Alibaba Cloud region (e.g., `cn-hangzhou`) | Required; non-empty |
| `dbType` | string | Database engine | Required; one of: MySQL, PostgreSQL, Oracle |
| `dbVersion` | string | Engine version (e.g., `8.0`, `14`, `11`) | Required; non-empty |
| `dbNodeClass` | string | Node instance class (e.g., `polar.mysql.x4.large`) | Required; non-empty |
| `vswitchId` | StringValueOrRef | VSwitch ID. Can reference AliCloudVswitch via `valueFrom`. | Required |

### Optional Fields

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `dbNodeCount` | int32 | `2` | Number of nodes (1 primary + N-1 read replicas); range 1-16 |
| `description` | string | metadata.name | Cluster description (2-256 chars) |
| `payType` | string | `PostPaid` | Billing: `PostPaid` or `PrePaid` |
| `period` | int32 | | Subscription period in months (for PrePaid) |
| `renewalStatus` | string | | Auto-renewal: `AutoRenewal`, `Normal`, `NotRenewal` |
| `autoRenewPeriod` | int32 | | Auto-renewal period in months |
| `zoneId` | string | | Primary availability zone |
| `securityIps` | list | | IP whitelist for access control |
| `securityGroupIds` | list | | VPC security group IDs (max 3) |
| `maintainTime` | string | | Maintenance window (e.g., `02:00Z-03:00Z`) |
| `resourceGroupId` | string | | Resource group for organizational grouping |
| `tags` | map | | Key-value tags |
| `creationCategory` | string | | Edition: `Normal`, `Basic`, `ArchiveNormal`, `NormalMultimaster`, `SENormal` |
| `subCategory` | string | | Sub-category: `Exclusive`, `General` (MySQL only) |
| `storageType` | string | | Storage: `PSL5`, `PSL4` (Enterprise), `ESSDPL0`-`ESSDPL3`, `ESSDAUTOPL` (Standard) |
| `storageSpace` | int32 | | Storage in GB (20-100000; Standard Edition only) |
| `tdeStatus` | string | | TDE: `Enabled` or `Disabled` (irreversible once enabled) |
| `encryptionKey` | string | | KMS key ID for TDE |
| `deletionLock` | int32 | | Deletion protection: `1` (locked) or `0` (unlocked) |
| `collectorStatus` | string | | Audit log: `Enable` or `Disabled` |
| `backupRetentionPolicyOnClusterDeletion` | string | | Backup on delete: `ALL`, `LATEST`, `NONE` |
| `parameters` | list | | Cluster parameter overrides |
| `databases` | list | | Databases to create (see below) |
| `accounts` | list | | Accounts to create (see below) |

### Database Fields

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `dbName` | string | | Database name (required) |
| `characterSetName` | string | engine default | Character set (e.g., `utf8`, `utf8mb4`, `UTF8`) |
| `dbDescription` | string | | Database description |
| `collate` | string | | Collation rules (PostgreSQL/Oracle only) |
| `ctype` | string | | Character type (PostgreSQL/Oracle only) |

### Account Fields

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `accountName` | string | | Login name (required; 2-16 chars) |
| `accountPassword` | string | | Password (required; 8+ chars) |
| `accountType` | string | `Normal` | `Normal` or `Super` |
| `accountDescription` | string | | Account description |
| `privileges` | list | | Database privileges (see below) |

### Privilege Fields

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `dbNames` | list | | Databases to grant access to (required; min 1) |
| `accountPrivilege` | string | `ReadOnly` | `ReadOnly`, `ReadWrite`, `DDLOnly`, `DMLOnly` |

## Stack Outputs

After deployment, the following outputs are available in `status.outputs`:

| Output | Type | Description |
|--------|------|-------------|
| `cluster_id` | string | PolarDB cluster ID (e.g., `pc-xxxxx`) |
| `connection_string` | string | Primary endpoint connection string |
| `port` | string | Database service port |
| `database_ids` | map | Map of database names to their IDs |

## Related Components

- **AliCloudVswitch** -- VSwitch where the PolarDB cluster is placed
- **AliCloudVpc** -- VPC that provides network isolation
- **AliCloudSecurityGroup** -- Network security rules for cluster access
- **AliCloudKmsKey** -- Customer-managed key for TDE encryption
- **AliCloudPrivateDnsZone** -- Private DNS resolution for the cluster endpoint
- **AliCloudRdsInstance** -- Alternative: traditional RDS for non-cloud-native workloads
