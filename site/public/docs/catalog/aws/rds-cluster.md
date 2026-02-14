---
title: "RDS Cluster"
description: "RDS Cluster deployment documentation"
icon: "package"
order: 100
componentName: "awsrdscluster"
---

# AWS RDS Cluster

Deploys an Amazon Aurora DB cluster (MySQL or PostgreSQL) with automatic subnet group creation, managed security group configuration, optional Secrets Manager password management, and optional Serverless v2 scaling. The component handles cluster-level configuration; instance-level resources are managed separately.

## What Gets Created

When you deploy an AwsRdsCluster resource, OpenMCF provisions:

- **RDS Aurora Cluster** — an `rds.Cluster` with the specified engine (`aurora-mysql` or `aurora-postgresql`), encryption settings, backup configuration, and optional Serverless v2 scaling
- **DB Subnet Group** — an `rds.SubnetGroup` created automatically when `subnetIds` are provided and `dbSubnetGroupName` is not set, placing the cluster across the specified subnets
- **Security Group** — an `ec2.SecurityGroup` created when `securityGroupIds` or `allowedCidrBlocks` are provided, with ingress rules on the cluster port from the specified sources and unrestricted egress
- **Security Group Ingress Rules** — one `ec2.SecurityGroupRule` per source security group and one for CIDR blocks, scoped to the configured `port`
- **Security Group Egress Rule** — an `ec2.SecurityGroupRule` allowing all outbound traffic
- **Cluster Parameter Group** — an `rds.ClusterParameterGroup` created when inline `parameters` are provided, with the family auto-derived from the engine and engine version

## Prerequisites

- **AWS credentials** configured via environment variables or OpenMCF provider config
- **At least two subnets** in different Availability Zones, or an existing DB subnet group name
- **An Aurora-compatible engine** identifier (`aurora-mysql` or `aurora-postgresql`) and a valid engine version
- **A VPC ID** if creating a managed security group with `securityGroupIds` or `allowedCidrBlocks`
- **A KMS key ARN** if enabling storage encryption with a customer-managed key
- **An ACM Secrets Manager KMS key** if using `manageMasterUserPassword` with a custom KMS key

## Quick Start

Create a file `rds-cluster.yaml`:

```yaml
apiVersion: aws.openmcf.org/v1
kind: AwsRdsCluster
metadata:
  name: my-aurora-cluster
  labels:
    openmcf.org/provisioner: pulumi
    pulumi.openmcf.org/organization: my-org
    pulumi.openmcf.org/project: my-project
    pulumi.openmcf.org/stack.name: dev.AwsRdsCluster.my-aurora-cluster
spec:
  engine: aurora-mysql
  engineVersion: "8.0.mysql_aurora.3.05.2"
  subnetIds:
    - subnet-0a1b2c3d4e5f00001
    - subnet-0a1b2c3d4e5f00002
  manageMasterUserPassword: true
  skipFinalSnapshot: true
```

Deploy:

```shell
openmcf apply -f rds-cluster.yaml
```

This creates an Aurora MySQL cluster across two subnets with RDS-managed master password stored in AWS Secrets Manager.

## Configuration Reference

### Required Fields

| Field | Type | Description | Validation |
|-------|------|-------------|------------|
| `engine` | `string` | Aurora engine identifier. | Must be set. Examples: `aurora-mysql`, `aurora-postgresql`. |
| `engineVersion` | `string` | Engine version to deploy. | Must be set. Examples: `8.0.mysql_aurora.3.05.2`, `14.6`. |
| `subnetIds` | `StringValueOrRef[]` | Subnet IDs for the DB subnet group. Provide at least two in distinct AZs. | Minimum 2 items unless `dbSubnetGroupName` is set. Can reference AwsVpc resource via `valueFrom`. |
| `finalSnapshotIdentifier` | `string` | Identifier for the final DB snapshot on deletion. | Required when `skipFinalSnapshot` is `false`. |

### Optional Fields

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `dbSubnetGroupName` | `StringValueOrRef` | — | Name of an existing DB subnet group. When set, `subnetIds` is not required. |
| `securityGroupIds` | `StringValueOrRef[]` | `[]` | Security group IDs used to create ingress rules on the managed security group. Can reference AwsSecurityGroup resources via `valueFrom`. |
| `allowedCidrBlocks` | `string[]` | `[]` | IPv4 CIDRs to allow ingress to the managed security group. Must be unique, valid CIDR notation. |
| `associateSecurityGroupIds` | `StringValueOrRef[]` | `[]` | Existing security groups to attach directly to the cluster (in addition to the managed SG). Can reference AwsSecurityGroup resources via `valueFrom`. |
| `vpcId` | `StringValueOrRef` | — | VPC ID for the managed security group. Can reference AwsVpc resource via `valueFrom`. |
| `databaseName` | `string` | — | Name of the initial database to create in the cluster. |
| `manageMasterUserPassword` | `bool` | `false` (recommended: `true`) | When `true`, RDS manages the master password in AWS Secrets Manager. Cannot be used with `password`. |
| `masterUserSecretKmsKeyId` | `StringValueOrRef` | — | KMS key ARN for encrypting the managed master password secret. Only used when `manageMasterUserPassword` is `true`. Can reference AwsKmsKey resource via `valueFrom`. |
| `username` | `string` | `"master"` | Master database user name. |
| `password` | `string` | — | Master user password. Cannot be set when `manageMasterUserPassword` is `true`. |
| `port` | `int32` | `0` | Port on which the cluster accepts connections. Valid range: 0-65535. |
| `storageEncrypted` | `bool` | `false` | Enables encryption at rest for the cluster. |
| `kmsKeyId` | `StringValueOrRef` | — | KMS key ARN for storage encryption. Used when `storageEncrypted` is `true`. Can reference AwsKmsKey resource via `valueFrom`. |
| `enabledCloudwatchLogsExports` | `string[]` | `[]` | Log types to export to CloudWatch. Aurora MySQL: `audit`, `error`, `general`, `slowquery`. Aurora PostgreSQL: `postgresql`, `upgrade`. |
| `deletionProtection` | `bool` | `false` | Prevents accidental cluster deletion when enabled. |
| `preferredMaintenanceWindow` | `string` | — | Weekly maintenance window in UTC. Format: `ddd:hh24:mi-ddd:hh24:mi` (e.g., `sun:05:00-sun:06:00`). |
| `backupRetentionPeriod` | `int32` | `0` | Number of days to retain automated backups. Valid range: 0-35. Values greater than 0 enable automated backups. |
| `preferredBackupWindow` | `string` | — | Daily backup window in UTC. Format: `hh24:mi-hh24:mi` (e.g., `03:00-04:00`). Must not overlap the maintenance window. |
| `copyTagsToSnapshot` | `bool` | `false` | Copies cluster tags to DB snapshots. |
| `skipFinalSnapshot` | `bool` | `false` | When `true`, no final snapshot is created on cluster deletion. When `false`, `finalSnapshotIdentifier` is required. |
| `iamDatabaseAuthenticationEnabled` | `bool` | `false` | Enables IAM user/role mapping to database logins. |
| `enableHttpEndpoint` | `bool` | `false` | Enables the Data API for Aurora Serverless (where supported). |
| `serverlessV2Scaling.minCapacity` | `double` | — | Minimum Aurora Capacity Units (ACUs) for Serverless v2. Must be greater than 0. |
| `serverlessV2Scaling.maxCapacity` | `double` | — | Maximum ACUs for Serverless v2. Must be greater than or equal to `minCapacity`. |
| `snapshotIdentifier` | `string` | — | Creates the cluster from the specified DB snapshot. |
| `replicationSourceIdentifier` | `string` | — | ARN or identifier of another cluster to create a read replica. |
| `dbClusterParameterGroupName` | `string` | — | Name of an existing cluster parameter group to associate. |
| `parameters` | `AwsRdsClusterParameter[]` | `[]` | Inline cluster parameters. Each entry has `name`, `value`, and optional `applyMethod` (`immediate` or `pending-reboot`). |
| `engineMode` | `string` | — | Engine mode. Valid values: `serverless` (for Aurora Serverless v1), `provisioned`. |
| `storageType` | `string` | — | Aurora storage type. Valid values: `aurora`, `aurora-iopt1`. |

## Examples

### Aurora MySQL with Managed Password

A basic Aurora MySQL cluster that delegates password management to AWS Secrets Manager:

```yaml
apiVersion: aws.openmcf.org/v1
kind: AwsRdsCluster
metadata:
  name: app-db
  labels:
    openmcf.org/provisioner: pulumi
    pulumi.openmcf.org/organization: my-org
    pulumi.openmcf.org/project: my-project
    pulumi.openmcf.org/stack.name: dev.AwsRdsCluster.app-db
spec:
  engine: aurora-mysql
  engineVersion: "8.0.mysql_aurora.3.05.2"
  subnetIds:
    - subnet-private-az1
    - subnet-private-az2
  manageMasterUserPassword: true
  databaseName: appdb
  skipFinalSnapshot: true
```

### Aurora PostgreSQL with Encryption and Backups

A production-oriented Aurora PostgreSQL cluster with storage encryption, backup retention, and deletion protection:

```yaml
apiVersion: aws.openmcf.org/v1
kind: AwsRdsCluster
metadata:
  name: analytics-db
  labels:
    openmcf.org/provisioner: pulumi
    pulumi.openmcf.org/organization: my-org
    pulumi.openmcf.org/project: my-project
    pulumi.openmcf.org/stack.name: prod.AwsRdsCluster.analytics-db
spec:
  engine: aurora-postgresql
  engineVersion: "14.6"
  subnetIds:
    - subnet-private-az1
    - subnet-private-az2
    - subnet-private-az3
  manageMasterUserPassword: true
  databaseName: analytics
  storageEncrypted: true
  kmsKeyId: arn:aws:kms:us-east-1:123456789012:key/abc-12345
  deletionProtection: true
  backupRetentionPeriod: 14
  preferredBackupWindow: "03:00-04:00"
  preferredMaintenanceWindow: "sun:05:00-sun:06:00"
  copyTagsToSnapshot: true
  skipFinalSnapshot: false
  finalSnapshotIdentifier: analytics-db-final
  enabledCloudwatchLogsExports:
    - postgresql
    - upgrade
```

### Aurora Serverless v2

An Aurora MySQL cluster using Serverless v2 auto-scaling capacity:

```yaml
apiVersion: aws.openmcf.org/v1
kind: AwsRdsCluster
metadata:
  name: serverless-db
  labels:
    openmcf.org/provisioner: pulumi
    pulumi.openmcf.org/organization: my-org
    pulumi.openmcf.org/project: my-project
    pulumi.openmcf.org/stack.name: dev.AwsRdsCluster.serverless-db
spec:
  engine: aurora-mysql
  engineVersion: "8.0.mysql_aurora.3.05.2"
  subnetIds:
    - subnet-private-az1
    - subnet-private-az2
  manageMasterUserPassword: true
  databaseName: myapp
  serverlessV2Scaling:
    minCapacity: 0.5
    maxCapacity: 16
  skipFinalSnapshot: true
```

### Cluster with Security Group and CIDR Access

A cluster with a managed security group allowing access from specific CIDRs and existing security groups:

```yaml
apiVersion: aws.openmcf.org/v1
kind: AwsRdsCluster
metadata:
  name: secured-db
  labels:
    openmcf.org/provisioner: pulumi
    pulumi.openmcf.org/organization: my-org
    pulumi.openmcf.org/project: my-project
    pulumi.openmcf.org/stack.name: prod.AwsRdsCluster.secured-db
spec:
  engine: aurora-postgresql
  engineVersion: "15.4"
  subnetIds:
    - subnet-private-az1
    - subnet-private-az2
  vpcId: vpc-0a1b2c3d4e5f00001
  securityGroupIds:
    - sg-app-servers
  allowedCidrBlocks:
    - "10.0.0.0/16"
  associateSecurityGroupIds:
    - sg-monitoring
  manageMasterUserPassword: true
  databaseName: proddb
  port: 5432
  storageEncrypted: true
  deletionProtection: true
  iamDatabaseAuthenticationEnabled: true
  skipFinalSnapshot: false
  finalSnapshotIdentifier: secured-db-final
```

### Using Foreign Key References

Reference other OpenMCF-managed resources instead of hardcoding IDs:

```yaml
apiVersion: aws.openmcf.org/v1
kind: AwsRdsCluster
metadata:
  name: ref-db
  labels:
    openmcf.org/provisioner: pulumi
    pulumi.openmcf.org/organization: my-org
    pulumi.openmcf.org/project: my-project
    pulumi.openmcf.org/stack.name: prod.AwsRdsCluster.ref-db
spec:
  engine: aurora-mysql
  engineVersion: "8.0.mysql_aurora.3.05.2"
  subnetIds:
    - valueFrom:
        kind: AwsVpc
        name: my-vpc
        field: status.outputs.private_subnets.[0].id
    - valueFrom:
        kind: AwsVpc
        name: my-vpc
        field: status.outputs.private_subnets.[1].id
  vpcId:
    valueFrom:
      kind: AwsVpc
      name: my-vpc
      field: status.outputs.vpc_id
  securityGroupIds:
    - valueFrom:
        kind: AwsSecurityGroup
        name: db-sg
        field: status.outputs.security_group_id
  manageMasterUserPassword: true
  masterUserSecretKmsKeyId:
    valueFrom:
      kind: AwsKmsKey
      name: db-secret-key
      field: status.outputs.key_arn
  kmsKeyId:
    valueFrom:
      kind: AwsKmsKey
      name: db-storage-key
      field: status.outputs.key_arn
  storageEncrypted: true
  databaseName: myapp
  skipFinalSnapshot: false
  finalSnapshotIdentifier: ref-db-final
```

## Stack Outputs

After deployment, the following outputs are available in `status.outputs`:

| Output | Type | Description |
|--------|------|-------------|
| `rds_cluster_endpoint` | `string` | The primary writer endpoint for the DB cluster |
| `rds_cluster_reader_endpoint` | `string` | The reader endpoint for load-balanced read traffic across replicas |
| `rds_cluster_id` | `string` | The AWS identifier of the DB cluster |
| `rds_cluster_arn` | `string` | The Amazon Resource Name of the DB cluster |
| `rds_cluster_engine` | `string` | The engine used by the cluster (e.g., `aurora-mysql`, `aurora-postgresql`) |
| `rds_cluster_engine_version` | `string` | The engine version running on the cluster |
| `rds_cluster_port` | `int32` | The port on which the DB cluster accepts connections |
| `rds_subnet_group` | `string` | The name of the DB subnet group associated with the cluster (only when created by the module) |
| `rds_security_group` | `string` | The security group associated with the cluster (only when created by the module) |
| `rds_cluster_parameter_group` | `string` | The cluster parameter group in use (only when created by the module) |

## Related Components

- [AwsVpc](/docs/catalog/aws/vpc) — provides subnets and VPC ID for cluster placement
- [AwsSecurityGroup](/docs/catalog/aws/security-group) — controls network access to the cluster
- [AwsKmsKey](/docs/catalog/aws/kms-key) — provides KMS keys for storage encryption and Secrets Manager
- [AwsRoute53Zone](/docs/catalog/aws/route53-zone) — hosts DNS zones for CNAME records pointing to the cluster endpoint
