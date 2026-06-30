---
title: "Neptune Cluster"
description: "Neptune Cluster deployment documentation"
icon: "package"
order: 100
componentName: "awsneptunecluster"
---

# AWS Neptune Cluster

Deploys an Amazon Neptune graph database cluster with automatic subnet group creation, managed security group configuration, configurable cluster instances, optional Serverless v2 scaling, and optional parameter group customization. The component provisions both the cluster and its instances in a single resource definition. Neptune supports property-graph queries via Apache TinkerPop Gremlin and RDF queries via SPARQL.

## What Gets Created

When you deploy an AwsNeptuneCluster resource, Planton provisions:

- **Neptune Cluster** — a `neptune.Cluster` with the `neptune` engine at the specified version, encryption settings, backup configuration, IAM database authentication, optional Serverless v2 scaling, and CloudWatch log exports
- **Cluster Instances** — one `neptune.ClusterInstance` per `instanceCount` (default 1), each using the specified `instanceClass` (default `db.r6g.large`) with promotion tier assignment (primary at tier 0, replicas at tier 1)
- **Neptune Subnet Group** — a `neptune.SubnetGroup` created automatically when `subnetIds` are provided and `neptuneSubnetGroupName` is not set, placing the cluster across the specified subnets
- **Security Group** — an `ec2.SecurityGroup` created when `securityGroupIds` or `allowedCidrBlocks` are provided, with ingress rules on the cluster port from the specified sources and unrestricted egress
- **Security Group Ingress Rules** — one `ec2.SecurityGroupRule` per source security group and one for CIDR blocks, scoped to the configured `port` (default 8182)
- **Security Group Egress Rule** — an `ec2.SecurityGroupRule` allowing all outbound traffic
- **Cluster Parameter Group** — a `neptune.ClusterParameterGroup` created when `clusterParameters` are provided, with the family auto-derived from the engine version (e.g., `neptune1.3` for engine version `1.3.0.0`)

## Prerequisites

- **AWS credentials** configured via environment variables or Planton provider config
- **At least two subnets** in different Availability Zones, or an existing Neptune subnet group name
- **A VPC ID** if creating a managed security group with `securityGroupIds` or `allowedCidrBlocks`
- **A KMS key ARN** if using a customer-managed key for storage encryption
- **IAM roles** if Neptune needs to access S3 for bulk data loading

## Quick Start

Create a file `neptune.yaml`:

```yaml
apiVersion: aws.planton.dev/v1
kind: AwsNeptuneCluster
metadata:
  name: my-neptune
  labels:
    planton.dev/provisioner: pulumi
    pulumi.planton.dev/organization: my-org
    pulumi.planton.dev/project: my-project
    pulumi.planton.dev/stack.name: dev.AwsNeptuneCluster.my-neptune
spec:
  region: us-west-2
  subnetIds:
    - subnet-0a1b2c3d4e5f00001
    - subnet-0a1b2c3d4e5f00002
  storageEncrypted: true
  skipFinalSnapshot: true
```

Deploy:

```shell
planton apply -f neptune.yaml
```

This creates a single-instance Neptune 1.3.0.0 cluster with encrypted storage, IAM-ready authentication, and a `db.r6g.large` instance in the specified subnets. Unlike relational databases, Neptune does not require a master username or password — access is controlled via IAM database authentication and VPC security groups.

## Configuration Reference

### Required Fields

| Field | Type | Description | Validation |
|-------|------|-------------|------------|
| `region` | `string` | AWS region where the Neptune cluster will be created (e.g., `us-west-2`, `eu-west-1`). | Required; non-empty |
| `subnetIds` | `StringValueOrRef[]` | Subnet IDs for the Neptune subnet group. Provide at least two in distinct AZs. | Minimum 2 items unless `neptuneSubnetGroupName` is set. Can reference AwsVpc resource via `valueFrom`. |
| `finalSnapshotIdentifier` | `string` | Identifier for the final snapshot on deletion. | Required when `skipFinalSnapshot` is `false` (the default). |

### Optional Fields

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `neptuneSubnetGroupName` | `StringValueOrRef` | — | Name of an existing Neptune subnet group. When set, `subnetIds` is not required. |
| `securityGroupIds` | `StringValueOrRef[]` | `[]` | Security group IDs used to create ingress rules on the managed security group. Can reference AwsSecurityGroup resources via `valueFrom`. |
| `allowedCidrBlocks` | `string[]` | `[]` | IPv4 CIDRs to allow ingress to the managed security group. Must be unique, valid CIDR notation. |
| `vpcId` | `StringValueOrRef` | — | VPC ID for the managed security group. Can reference AwsVpc resource via `valueFrom`. |
| `engineVersion` | `string` | `"1.3.0.0"` | Neptune engine version. Examples: `1.2.1.0`, `1.3.0.0`, `1.3.1.0`. |
| `port` | `int32` | `8182` | TCP port on which the cluster accepts connections. Valid range: 1-65535. |
| `storageType` | `string` | `"standard"` | Storage I/O model. `standard` for general use; `iopt1` for I/O-Optimized storage with higher throughput and predictable pricing on read-heavy workloads. |
| `instanceCount` | `int32` | `1` | Number of instances to create. First instance is the primary writer; additional instances are read replicas. Minimum: 1. |
| `instanceClass` | `string` | `"db.r6g.large"` | Compute and memory capacity of each instance. Use `db.serverless` for Neptune Serverless (requires `serverlessV2Scaling`). Examples: `db.r6g.large`, `db.r6g.xlarge`, `db.r5.large`. |
| `serverlessV2Scaling.minCapacity` | `double` | — | Minimum Neptune Capacity Units (NCUs) for Serverless v2. Range: 1.0-128.0. |
| `serverlessV2Scaling.maxCapacity` | `double` | — | Maximum NCUs for Serverless v2. Must be >= `minCapacity`. Range: 1.0-128.0. |
| `storageEncrypted` | `bool` | `true` | Encrypts cluster storage at rest. |
| `kmsKeyId` | `StringValueOrRef` | — | KMS key ARN for storage encryption. Can reference AwsKmsKey resource via `valueFrom`. |
| `iamDatabaseAuthenticationEnabled` | `bool` | `false` | Enables IAM database authentication, allowing IAM users and roles to authenticate using temporary credentials. |
| `iamRoles` | `StringValueOrRef[]` | `[]` | IAM role ARNs to associate with Neptune for accessing other AWS services (e.g., S3 for bulk data loading). Can reference AwsIamRole resources via `valueFrom`. |
| `backupRetentionPeriod` | `int32` | `7` | Number of days to retain automated backups. Valid range: 1-35. |
| `preferredBackupWindow` | `string` | — | Daily backup window in UTC. Format: `hh24:mi-hh24:mi` (e.g., `03:00-04:00`). |
| `preferredMaintenanceWindow` | `string` | — | Weekly maintenance window in UTC. Format: `ddd:hh24:mi-ddd:hh24:mi` (e.g., `sun:05:00-sun:06:00`). |
| `deletionProtection` | `bool` | `false` | Prevents accidental cluster deletion when enabled. |
| `skipFinalSnapshot` | `bool` | `false` | When `true`, no final snapshot is created on deletion. When `false`, `finalSnapshotIdentifier` is required. |
| `enabledCloudwatchLogsExports` | `string[]` | `[]` | Log types to export to CloudWatch Logs. Valid values: `audit`, `slowquery`. |
| `applyImmediately` | `bool` | `false` | Applies modifications immediately instead of during the next maintenance window. |
| `copyTagsToSnapshot` | `bool` | `false` | Copies cluster tags to snapshots. |
| `allowMajorVersionUpgrade` | `bool` | `false` | Allows major engine version upgrades. Required when updating `engineVersion` to a new major version. |
| `clusterParameterGroupName` | `string` | — | Name of an existing Neptune cluster parameter group. |
| `clusterParameters` | `AwsNeptuneClusterParameter[]` | `[]` | Inline cluster parameters. Each entry has `name`, `value`, and optional `applyMethod` (`immediate` or `pending-reboot`). |

## Examples

### Production Graph Database with IAM Auth

A Neptune cluster with two instances for high availability, IAM authentication, encrypted storage, and audit logging:

```yaml
apiVersion: aws.planton.dev/v1
kind: AwsNeptuneCluster
metadata:
  name: knowledge-graph
  labels:
    planton.dev/provisioner: pulumi
    pulumi.planton.dev/organization: my-org
    pulumi.planton.dev/project: my-project
    pulumi.planton.dev/stack.name: prod.AwsNeptuneCluster.knowledge-graph
spec:
  region: us-west-2
  subnetIds:
    - subnet-private-az1
    - subnet-private-az2
    - subnet-private-az3
  vpcId: vpc-0a1b2c3d4e5f00001
  securityGroupIds:
    - sg-app-servers
  storageEncrypted: true
  kmsKeyId: arn:aws:kms:us-east-1:123456789012:key/abc-12345
  iamDatabaseAuthenticationEnabled: true
  instanceCount: 2
  instanceClass: db.r6g.xlarge
  storageType: iopt1
  deletionProtection: true
  backupRetentionPeriod: 14
  preferredBackupWindow: "03:00-04:00"
  preferredMaintenanceWindow: "sun:05:00-sun:06:00"
  copyTagsToSnapshot: true
  skipFinalSnapshot: false
  finalSnapshotIdentifier: knowledge-graph-final
  enabledCloudwatchLogsExports:
    - audit
    - slowquery
```

### Neptune Serverless v2

A Neptune Serverless cluster that auto-scales between 2.5 and 64 NCUs based on demand:

```yaml
apiVersion: aws.planton.dev/v1
kind: AwsNeptuneCluster
metadata:
  name: serverless-graph
  labels:
    planton.dev/provisioner: pulumi
    pulumi.planton.dev/organization: my-org
    pulumi.planton.dev/project: my-project
    pulumi.planton.dev/stack.name: dev.AwsNeptuneCluster.serverless-graph
spec:
  region: us-west-2
  subnetIds:
    - subnet-private-az1
    - subnet-private-az2
  storageEncrypted: true
  instanceClass: db.serverless
  serverlessV2Scaling:
    minCapacity: 2.5
    maxCapacity: 64.0
  iamDatabaseAuthenticationEnabled: true
  skipFinalSnapshot: true
```

### Neptune with S3 Bulk Loading

A cluster with IAM roles for loading graph data from S3:

```yaml
apiVersion: aws.planton.dev/v1
kind: AwsNeptuneCluster
metadata:
  name: data-loader
  labels:
    planton.dev/provisioner: pulumi
    pulumi.planton.dev/organization: my-org
    pulumi.planton.dev/project: my-project
    pulumi.planton.dev/stack.name: prod.AwsNeptuneCluster.data-loader
spec:
  region: us-west-2
  subnetIds:
    - subnet-private-az1
    - subnet-private-az2
  storageEncrypted: true
  iamDatabaseAuthenticationEnabled: true
  iamRoles:
    - arn:aws:iam::123456789012:role/NeptuneS3ReadRole
  instanceCount: 2
  enabledCloudwatchLogsExports:
    - audit
  skipFinalSnapshot: false
  finalSnapshotIdentifier: data-loader-final
```

### Using Foreign Key References

Reference other Planton-managed resources instead of hardcoding IDs:

```yaml
apiVersion: aws.planton.dev/v1
kind: AwsNeptuneCluster
metadata:
  name: ref-graph
  labels:
    planton.dev/provisioner: pulumi
    pulumi.planton.dev/organization: my-org
    pulumi.planton.dev/project: my-project
    pulumi.planton.dev/stack.name: prod.AwsNeptuneCluster.ref-graph
spec:
  region: us-west-2
  subnetIds:
    - valueFrom:
        kind: AwsSubnet
        name: my-private-subnet-a
        fieldPath: status.outputs.subnet_id
    - valueFrom:
        kind: AwsSubnet
        name: my-private-subnet-b
        fieldPath: status.outputs.subnet_id
  vpcId:
    valueFrom:
      kind: AwsVpc
      name: my-vpc
      field: status.outputs.vpc_id
  securityGroupIds:
    - valueFrom:
        kind: AwsSecurityGroup
        name: neptune-sg
        field: status.outputs.security_group_id
  kmsKeyId:
    valueFrom:
      kind: AwsKmsKey
      name: neptune-encryption-key
      field: status.outputs.key_arn
  iamRoles:
    - valueFrom:
        kind: AwsIamRole
        name: neptune-s3-role
        field: status.outputs.role_arn
  storageEncrypted: true
  iamDatabaseAuthenticationEnabled: true
  skipFinalSnapshot: false
  finalSnapshotIdentifier: ref-graph-final
```

## Stack Outputs

After deployment, the following outputs are available in `status.outputs`:

| Output | Type | Description |
|--------|------|-------------|
| `cluster_endpoint` | `string` | The primary writer endpoint for the Neptune cluster (Gremlin/SPARQL queries) |
| `cluster_reader_endpoint` | `string` | The reader endpoint for load-balanced read traffic across replicas |
| `cluster_id` | `string` | The AWS identifier of the Neptune cluster |
| `cluster_arn` | `string` | The Amazon Resource Name of the Neptune cluster |
| `cluster_resource_id` | `string` | The internal AWS resource identifier |
| `cluster_port` | `int32` | The port on which the cluster accepts connections (default 8182) |
| `db_subnet_group_name` | `string` | The subnet group name (only when created by the module) |
| `security_group_id` | `string` | The security group ID (only when created by the module) |
| `cluster_parameter_group_name` | `string` | The parameter group name (only when created by the module) |
| `hosted_zone_id` | `string` | The Route 53 hosted zone ID for the cluster endpoint |

## Related Components

- [AwsVpc](/docs/catalog/aws/vpc) — provides subnets and VPC ID for cluster placement
- [AwsSecurityGroup](/docs/catalog/aws/security-group) — controls network access to the cluster
- [AwsKmsKey](/docs/catalog/aws/kms-key) — provides KMS keys for storage encryption
- [AwsIamRole](/docs/catalog/aws/iam-role) — provides IAM roles for S3 bulk data loading
- [AwsRoute53Zone](/docs/catalog/aws/route53-zone) — hosts DNS zones for CNAME records pointing to the cluster endpoint
