# AWS DocumentDB

Deploys an AWS DocumentDB cluster (MongoDB-compatible) with automatic subnet group creation, managed security group configuration, configurable cluster instances, and optional parameter group customization. The component provisions both the cluster and its instances in a single resource definition.

## What Gets Created

When you deploy an AwsDocumentDb resource, OpenMCF provisions:

- **DocumentDB Cluster** — a `docdb.Cluster` with the `docdb` engine at the specified version, encryption settings, backup configuration, and CloudWatch log exports
- **Cluster Instances** — one `docdb.ClusterInstance` per `instanceCount` (default 1), each using the specified `instanceClass` with auto minor version upgrade support
- **DB Subnet Group** — a `docdb.SubnetGroup` created automatically when `subnets` are provided and `dbSubnetGroup` is not set, placing the cluster across the specified subnets
- **Security Group** — an `ec2.SecurityGroup` created when `securityGroups` or `allowedCidrs` are provided, with ingress rules on the cluster port from the specified sources and unrestricted egress
- **Security Group Ingress Rules** — one `ec2.SecurityGroupRule` per source security group and one for CIDR blocks, scoped to the configured `port` (default 27017)
- **Security Group Egress Rule** — an `ec2.SecurityGroupRule` allowing all outbound traffic
- **Cluster Parameter Group** — a `docdb.ClusterParameterGroup` created when `clusterParameters` are provided, with the family auto-derived from the engine version (e.g., `docdb5.0`)

## Prerequisites

- **AWS credentials** configured via environment variables or OpenMCF provider config
- **At least two subnets** in different Availability Zones, or an existing DB subnet group name
- **A master password** for the cluster administrator
- **A VPC ID** if creating a managed security group with `securityGroups` or `allowedCidrs`
- **A KMS key ARN** if using a customer-managed key for storage encryption

## Quick Start

Create a file `documentdb.yaml`:

```yaml
apiVersion: aws.openmcf.org/v1
kind: AwsDocumentDb
metadata:
  name: my-docdb
  labels:
    openmcf.org/provisioner: pulumi
    pulumi.openmcf.org/organization: my-org
    pulumi.openmcf.org/project: my-project
    pulumi.openmcf.org/stack.name: dev.AwsDocumentDb.my-docdb
spec:
  subnets:
    - subnet-0a1b2c3d4e5f00001
    - subnet-0a1b2c3d4e5f00002
  masterPassword: "change-me-before-deploy"
  skipFinalSnapshot: true
```

Deploy:

```shell
openmcf apply -f documentdb.yaml
```

This creates a single-instance DocumentDB 5.0 cluster with encrypted storage, a `docdbadmin` master user, and a `db.r6g.large` instance in the specified subnets.

## Configuration Reference

### Required Fields

| Field | Type | Description | Validation |
|-------|------|-------------|------------|
| `masterPassword` | `string` | Master user password for the cluster. | Must be set. |
| `subnets` | `StringValueOrRef[]` | Subnet IDs for the DB subnet group. Provide at least two in distinct AZs. | Minimum 2 items unless `dbSubnetGroup` is set. Can reference AwsVpc resource via `valueFrom`. |
| `finalSnapshotIdentifier` | `string` | Identifier for the final snapshot on deletion. | Required when `skipFinalSnapshot` is `false` (the default). |

### Optional Fields

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `dbSubnetGroup` | `StringValueOrRef` | — | Name of an existing DB subnet group. When set, `subnets` is not required. |
| `securityGroups` | `StringValueOrRef[]` | `[]` | Security group IDs used to create ingress rules on the managed security group. Can reference AwsSecurityGroup resources via `valueFrom`. |
| `allowedCidrs` | `string[]` | `[]` | IPv4 CIDRs to allow ingress to the managed security group. Must be unique, valid CIDR notation. |
| `vpc` | `StringValueOrRef` | — | VPC ID for the managed security group. Can reference AwsVpc resource via `valueFrom`. |
| `engineVersion` | `string` | `"5.0.0"` | DocumentDB engine version. Examples: `4.0.0`, `5.0.0`. |
| `port` | `int32` | `27017` | TCP port on which the cluster accepts connections. Valid range: 1-65535. |
| `masterUsername` | `string` | `"docdbadmin"` | Master user name for the cluster. |
| `instanceCount` | `int32` | `1` | Number of instances to create in the cluster. Minimum: 1. |
| `instanceClass` | `string` | `"db.r6g.large"` | Compute and memory capacity of each instance. Examples: `db.r5.large`, `db.r5.xlarge`, `db.r6g.large`. |
| `storageEncrypted` | `bool` | `true` | Encrypts cluster storage at rest. |
| `kmsKey` | `StringValueOrRef` | — | KMS key ARN for storage encryption. Can reference AwsKmsKey resource via `valueFrom`. |
| `backupRetentionPeriod` | `int32` | `7` | Number of days to retain automated backups. Valid range: 1-35. |
| `preferredBackupWindow` | `string` | — | Daily backup window in UTC. Format: `hh24:mi-hh24:mi` (e.g., `03:00-04:00`). |
| `preferredMaintenanceWindow` | `string` | — | Weekly maintenance window in UTC. Format: `ddd:hh24:mi-ddd:hh24:mi` (e.g., `sun:05:00-sun:06:00`). |
| `deletionProtection` | `bool` | `false` | Prevents accidental cluster deletion when enabled. |
| `skipFinalSnapshot` | `bool` | `false` | When `true`, no final snapshot is created on cluster deletion. When `false`, `finalSnapshotIdentifier` is required. |
| `enabledCloudwatchLogsExports` | `string[]` | `[]` | Log types to export to CloudWatch. Valid values: `audit`, `profiler`. |
| `applyImmediately` | `bool` | `false` | When `true`, modifications are applied immediately instead of during the next maintenance window. |
| `autoMinorVersionUpgrade` | `bool` | `true` | Enables automatic minor engine version upgrades for instances. |
| `clusterParameterGroupName` | `string` | — | Name of an existing cluster parameter group. |
| `clusterParameters` | `AwsDocumentDbParameter[]` | `[]` | Custom cluster parameters. Each entry has `name`, `value`, and optional `applyMethod` (`immediate` or `pending-reboot`). |

## Examples

### Single-Instance Development Cluster

A minimal DocumentDB cluster for development with final snapshot skipped:

```yaml
apiVersion: aws.openmcf.org/v1
kind: AwsDocumentDb
metadata:
  name: dev-docdb
  labels:
    openmcf.org/provisioner: pulumi
    pulumi.openmcf.org/organization: my-org
    pulumi.openmcf.org/project: my-project
    pulumi.openmcf.org/stack.name: dev.AwsDocumentDb.dev-docdb
spec:
  subnets:
    - subnet-private-az1
    - subnet-private-az2
  masterPassword: "dev-password-123"
  skipFinalSnapshot: true
  applyImmediately: true
```

### Multi-Instance Production Cluster

A three-instance cluster with encryption, backup retention, deletion protection, and audit logging:

```yaml
apiVersion: aws.openmcf.org/v1
kind: AwsDocumentDb
metadata:
  name: prod-docdb
  labels:
    openmcf.org/provisioner: pulumi
    pulumi.openmcf.org/organization: my-org
    pulumi.openmcf.org/project: my-project
    pulumi.openmcf.org/stack.name: prod.AwsDocumentDb.prod-docdb
spec:
  subnets:
    - subnet-private-az1
    - subnet-private-az2
    - subnet-private-az3
  masterUsername: prodadmin
  masterPassword: "prod-secure-password"
  instanceCount: 3
  instanceClass: db.r6g.xlarge
  storageEncrypted: true
  kmsKey: arn:aws:kms:us-east-1:123456789012:key/abc-12345
  deletionProtection: true
  backupRetentionPeriod: 14
  preferredBackupWindow: "03:00-04:00"
  preferredMaintenanceWindow: "sun:05:00-sun:06:00"
  skipFinalSnapshot: false
  finalSnapshotIdentifier: prod-docdb-final
  enabledCloudwatchLogsExports:
    - audit
    - profiler
```

### Cluster with Security Group and CIDR Access

A cluster with a managed security group allowing access from a VPC CIDR and specific application security groups:

```yaml
apiVersion: aws.openmcf.org/v1
kind: AwsDocumentDb
metadata:
  name: secured-docdb
  labels:
    openmcf.org/provisioner: pulumi
    pulumi.openmcf.org/organization: my-org
    pulumi.openmcf.org/project: my-project
    pulumi.openmcf.org/stack.name: prod.AwsDocumentDb.secured-docdb
spec:
  subnets:
    - subnet-private-az1
    - subnet-private-az2
  vpc: vpc-0a1b2c3d4e5f00001
  securityGroups:
    - sg-app-servers
  allowedCidrs:
    - "10.0.0.0/16"
  masterPassword: "secure-password-456"
  instanceCount: 2
  storageEncrypted: true
  deletionProtection: true
  skipFinalSnapshot: false
  finalSnapshotIdentifier: secured-docdb-final
```

### Cluster with Custom Parameters

A cluster using DocumentDB 5.0 with custom parameter group settings for profiler and TLS:

```yaml
apiVersion: aws.openmcf.org/v1
kind: AwsDocumentDb
metadata:
  name: custom-docdb
  labels:
    openmcf.org/provisioner: pulumi
    pulumi.openmcf.org/organization: my-org
    pulumi.openmcf.org/project: my-project
    pulumi.openmcf.org/stack.name: staging.AwsDocumentDb.custom-docdb
spec:
  subnets:
    - subnet-private-az1
    - subnet-private-az2
  masterPassword: "custom-password-789"
  engineVersion: "5.0.0"
  instanceCount: 2
  instanceClass: db.r5.large
  clusterParameters:
    - name: profiler
      value: enabled
      applyMethod: immediate
    - name: profiler_threshold_ms
      value: "100"
      applyMethod: immediate
    - name: tls
      value: enabled
      applyMethod: pending-reboot
  enabledCloudwatchLogsExports:
    - profiler
  skipFinalSnapshot: true
```

### Using Foreign Key References

Reference other OpenMCF-managed resources instead of hardcoding IDs:

```yaml
apiVersion: aws.openmcf.org/v1
kind: AwsDocumentDb
metadata:
  name: ref-docdb
  labels:
    openmcf.org/provisioner: pulumi
    pulumi.openmcf.org/organization: my-org
    pulumi.openmcf.org/project: my-project
    pulumi.openmcf.org/stack.name: prod.AwsDocumentDb.ref-docdb
spec:
  subnets:
    - valueFrom:
        kind: AwsVpc
        name: my-vpc
        field: status.outputs.private_subnets.[0].id
    - valueFrom:
        kind: AwsVpc
        name: my-vpc
        field: status.outputs.private_subnets.[1].id
  vpc:
    valueFrom:
      kind: AwsVpc
      name: my-vpc
      field: status.outputs.vpc_id
  securityGroups:
    - valueFrom:
        kind: AwsSecurityGroup
        name: docdb-sg
        field: status.outputs.security_group_id
  kmsKey:
    valueFrom:
      kind: AwsKmsKey
      name: docdb-key
      field: status.outputs.key_arn
  masterPassword: "ref-password-000"
  instanceCount: 2
  storageEncrypted: true
  skipFinalSnapshot: false
  finalSnapshotIdentifier: ref-docdb-final
```

## Stack Outputs

After deployment, the following outputs are available in `status.outputs`:

| Output | Type | Description |
|--------|------|-------------|
| `cluster_endpoint` | `string` | The primary writer endpoint for connecting to the DocumentDB cluster |
| `cluster_reader_endpoint` | `string` | The reader endpoint for load-balanced read traffic across replica instances |
| `cluster_id` | `string` | The AWS identifier of the DocumentDB cluster |
| `cluster_arn` | `string` | The Amazon Resource Name (ARN) of the DocumentDB cluster |
| `cluster_port` | `int32` | The port on which the DocumentDB cluster accepts connections |
| `cluster_resource_id` | `string` | The internal AWS resource identifier for the cluster |
| `connection_string` | `string` | A MongoDB-compatible connection string in the format `mongodb://user:<password>@endpoint:port/?tls=true&replicaSet=rs0&readPreference=secondaryPreferred&retryWrites=false` |
| `db_subnet_group_name` | `string` | The name of the DB subnet group associated with the cluster (only when created by the module) |
| `security_group_id` | `string` | The security group ID associated with the cluster (only when created by the module) |
| `cluster_parameter_group_name` | `string` | The cluster parameter group name in use (only when created by the module) |

## Related Components

- [AwsVpc](/docs/catalog/aws/awsvpc) — provides subnets and VPC ID for cluster placement
- [AwsSecurityGroup](/docs/catalog/aws/awssecuritygroup) — controls network access to the cluster
- [AwsKmsKey](/docs/catalog/aws/awskmskey) — provides KMS keys for storage encryption
