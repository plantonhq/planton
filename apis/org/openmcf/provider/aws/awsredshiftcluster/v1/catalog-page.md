# AWS Redshift Cluster

Deploys an Amazon Redshift data warehouse cluster with automatic subnet group creation, managed security group configuration, optional Secrets Manager password management, KMS encryption, audit logging, and inline parameter group support. Redshift is a petabyte-scale columnar data warehouse for analytical (OLAP) queries on structured and semi-structured data.

## What Gets Created

When you deploy an AwsRedshiftCluster resource, OpenMCF provisions:

- **Redshift Cluster** — a `redshift.Cluster` with the specified node type, node count, encryption settings, snapshot configuration, and optional Multi-AZ deployment
- **Subnet Group** — a `redshift.SubnetGroup` created automatically when `subnetIds` are provided and `clusterSubnetGroupName` is not set, placing the cluster across the specified subnets
- **Security Group** — an `ec2.SecurityGroup` created when `securityGroupIds` or `allowedCidrBlocks` are provided, with ingress rules on the cluster port from the specified sources and unrestricted egress
- **Security Group Ingress Rules** — one `ec2.SecurityGroupRule` per source security group and one for CIDR blocks, scoped to the configured `port` (default 5439)
- **Security Group Egress Rule** — an `ec2.SecurityGroupRule` allowing all outbound traffic
- **Parameter Group** — a `redshift.ParameterGroup` (family `redshift-1.0`) created when inline `parameters` are provided
- **Logging Configuration** — a `redshift.LoggingConfiguration` created when `logging` is specified, sending audit logs to S3 or CloudWatch Logs

## Prerequisites

- **AWS credentials** configured via environment variables or OpenMCF provider config
- **At least two subnets** in different Availability Zones, or an existing Redshift subnet group name
- **A VPC ID** if creating a managed security group with `securityGroupIds` or `allowedCidrBlocks`
- **A KMS key ARN** if enabling encryption with a customer-managed key or encrypting the managed password secret
- **IAM role ARNs** if the cluster needs to access S3, DynamoDB, Glue Data Catalog, or other AWS services

## Quick Start

Create a file `redshift-cluster.yaml`:

```yaml
apiVersion: aws.openmcf.org/v1
kind: AwsRedshiftCluster
metadata:
  name: my-warehouse
  labels:
    openmcf.org/provisioner: pulumi
    pulumi.openmcf.org/organization: my-org
    pulumi.openmcf.org/project: my-project
    pulumi.openmcf.org/stack.name: dev.AwsRedshiftCluster.my-warehouse
spec:
  nodeType: dc2.large
  subnetIds:
    - value: "<private-subnet-id-az1>"
    - value: "<private-subnet-id-az2>"
  manageMasterPassword: true
  encrypted: true
  skipFinalSnapshot: true
```

Deploy:

```shell
openmcf apply -f redshift-cluster.yaml
```

This creates a single-node dc2.large Redshift cluster across two subnets with AWS-managed encryption and a master password stored in AWS Secrets Manager.

## Configuration Reference

### Required Fields

| Field | Type | Description |
|-------|------|-------------|
| `nodeType` | `string` | Compute and storage capacity of each node. Common values: `dc2.large`, `ra3.xlplus`, `ra3.4xlarge`, `ra3.16xlarge`. RA3 nodes are recommended for most workloads (decoupled compute and storage). |
| `subnetIds` | `StringValueOrRef[]` | Subnet IDs for automatic Redshift subnet group creation. Provide at least two in distinct AZs. Can reference `AwsVpc` outputs via `valueFrom`. Not required when `clusterSubnetGroupName` is set. |
| `finalSnapshotIdentifier` | `string` | Identifier for the final snapshot created on cluster deletion. Required when `skipFinalSnapshot` is `false`. |

### Optional Fields

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `numberOfNodes` | `int32` | `1` | Cluster size. `1` = single-node (leader+compute combined); `>1` = multi-node (dedicated leader + N compute nodes). |
| `databaseName` | `string` | `"dev"` | Name of the first database created in the cluster. 1-64 characters, lowercase alphanumeric and underscores, starts with a letter or underscore. |
| `masterUsername` | `string` | `"admin"` | Admin user for the cluster. 1-128 characters, starts with a letter. |
| `masterPassword` | `string` | — | Admin password (8-64 chars, mixed case + digit). Mutually exclusive with `manageMasterPassword`. |
| `manageMasterPassword` | `bool` | `false` (recommended: `true`) | When `true`, AWS Secrets Manager generates, rotates, and stores the master password. Mutually exclusive with `masterPassword`. |
| `masterPasswordSecretKmsKeyId` | `StringValueOrRef` | — | KMS key ARN to encrypt the Secrets Manager secret holding the managed password. Only used when `manageMasterPassword` is `true`. Can reference `AwsKmsKey` via `valueFrom`. |
| `port` | `int32` | `5439` | TCP port for client connections. Valid range: 1115-65535. |
| `clusterSubnetGroupName` | `StringValueOrRef` | — | Name of an existing Redshift subnet group. When set, `subnetIds` is not required. |
| `securityGroupIds` | `StringValueOrRef[]` | `[]` | Security group IDs used to create ingress rules on the managed security group. Can reference `AwsSecurityGroup` via `valueFrom`. |
| `allowedCidrBlocks` | `string[]` | `[]` | IPv4 CIDRs to allow ingress on the cluster port. Must be unique, valid CIDR notation. |
| `associateSecurityGroupIds` | `StringValueOrRef[]` | `[]` | Existing security groups attached directly to the cluster alongside the managed SG. Can reference `AwsSecurityGroup` via `valueFrom`. |
| `vpcId` | `StringValueOrRef` | — | VPC ID for the managed security group. Required when `securityGroupIds` or `allowedCidrBlocks` are provided. Can reference `AwsVpc` via `valueFrom`. |
| `publiclyAccessible` | `bool` | `false` | When `true`, the cluster gets a public IP and is reachable from outside the VPC. |
| `enhancedVpcRouting` | `bool` | `false` | Forces all COPY/UNLOAD traffic through the VPC, enabling VPC flow logs and endpoint policies. |
| `multiAz` | `bool` | `false` | Enables Multi-AZ deployment with automatic failover. Requires RA3 node types. |
| `encrypted` | `bool` | `true` | Enables at-rest encryption. Uses the AWS-managed Redshift service key unless `kmsKeyId` is specified. |
| `kmsKeyId` | `StringValueOrRef` | — | Customer-managed KMS key ARN for cluster encryption. Requires `encrypted: true`. Can reference `AwsKmsKey` via `valueFrom`. |
| `iamRoles` | `StringValueOrRef[]` | `[]` | IAM roles attached to the cluster for accessing S3, DynamoDB, Glue, etc. Maximum 10 roles. Can reference `AwsIamRole` via `valueFrom`. |
| `defaultIamRoleArn` | `StringValueOrRef` | — | IAM role used by default when SQL commands do not specify a role. Can reference `AwsIamRole` via `valueFrom`. |
| `automatedSnapshotRetentionPeriod` | `int32` | `1` | Days to retain automated snapshots. `0` disables automated snapshots. Maximum: 35. |
| `skipFinalSnapshot` | `bool` | `false` | When `true`, no final snapshot is created on deletion. Set to `true` only for ephemeral dev/test clusters. |
| `preferredMaintenanceWindow` | `string` | — | Weekly UTC maintenance window. Format: `ddd:hh:mi-ddd:hh:mi` (e.g., `sat:03:00-sat:04:00`). |
| `allowVersionUpgrade` | `bool` | `true` | Permits AWS to apply major engine version upgrades during the maintenance window. |
| `maintenanceTrackName` | `string` | — | Cluster maintenance track. `"current"` applies the latest approved version; `"trailing"` uses the previous major version. |
| `applyImmediately` | `bool` | `false` | When `true`, modifications apply immediately instead of during the next maintenance window. |
| `logging` | `object` | — | Audit logging configuration. See sub-fields below. |
| `logging.logDestinationType` | `string` | — | Where audit logs are delivered. `"s3"` or `"cloudwatch"`. Required when `logging` is set. |
| `logging.s3BucketName` | `string` | — | S3 bucket for log delivery. Required when `logDestinationType` is `"s3"`. |
| `logging.s3KeyPrefix` | `string` | — | Prefix for log objects in the S3 bucket. |
| `logging.logExports` | `string[]` | `[]` | Log types to export: `connectionlog`, `useractivitylog`, `userlog`. Required when `logDestinationType` is `"cloudwatch"`. |
| `clusterParameterGroupName` | `string` | — | Name of an existing parameter group to associate. Ignored when inline `parameters` are provided. |
| `parameters` | `AwsRedshiftClusterParameter[]` | `[]` | Inline parameters (family: `redshift-1.0`). Each entry has `name` and `value`. Common parameters: `require_ssl`, `enable_user_activity_logging`, `max_concurrency_scaling_clusters`. |

## Examples

### Single-Node Development Cluster

A minimal single-node cluster for development and testing. Uses dc2.large for low cost, skips the final snapshot, and retains automated snapshots for 1 day:

```yaml
apiVersion: aws.openmcf.org/v1
kind: AwsRedshiftCluster
metadata:
  name: dev-warehouse
  labels:
    openmcf.org/provisioner: pulumi
    pulumi.openmcf.org/organization: my-org
    pulumi.openmcf.org/project: my-project
    pulumi.openmcf.org/stack.name: dev.AwsRedshiftCluster.dev-warehouse
spec:
  nodeType: dc2.large
  numberOfNodes: 1
  databaseName: dev
  masterUsername: admin
  manageMasterPassword: true
  subnetIds:
    - value: "<private-subnet-id-az1>"
    - value: "<private-subnet-id-az2>"
  encrypted: true
  skipFinalSnapshot: true
  automatedSnapshotRetentionPeriod: 1
```

### Production Multi-Node with Encryption and Logging

A 2-node RA3 cluster with customer-managed KMS encryption, SSL enforcement, enhanced VPC routing, CloudWatch audit logging, and a 7-day snapshot retention policy:

```yaml
apiVersion: aws.openmcf.org/v1
kind: AwsRedshiftCluster
metadata:
  name: prod-warehouse
  labels:
    openmcf.org/provisioner: pulumi
    pulumi.openmcf.org/organization: my-org
    pulumi.openmcf.org/project: my-project
    pulumi.openmcf.org/stack.name: prod.AwsRedshiftCluster.prod-warehouse
spec:
  nodeType: ra3.xlplus
  numberOfNodes: 2
  databaseName: analytics
  masterUsername: admin
  manageMasterPassword: true
  masterPasswordSecretKmsKeyId:
    value: "<kms-key-arn>"
  subnetIds:
    - value: "<private-subnet-id-az1>"
    - value: "<private-subnet-id-az2>"
  encrypted: true
  kmsKeyId:
    value: "<kms-key-arn>"
  enhancedVpcRouting: true
  automatedSnapshotRetentionPeriod: 7
  skipFinalSnapshot: false
  finalSnapshotIdentifier: prod-warehouse-final
  preferredMaintenanceWindow: "sat:03:00-sat:04:00"
  allowVersionUpgrade: true
  logging:
    logDestinationType: cloudwatch
    logExports:
      - connectionlog
      - useractivitylog
      - userlog
  parameters:
    - name: require_ssl
      value: "true"
    - name: enable_user_activity_logging
      value: "true"
  iamRoles:
    - value: "<redshift-s3-access-role-arn>"
```

### Analytics Workload with Multi-AZ and Spectrum IAM Roles

A 4-node ra3.4xlarge cluster for large-scale analytics. Multi-AZ provides automatic failover, concurrency scaling handles query bursts with up to 5 additional transient clusters, and two IAM roles are attached for S3 data loading and Redshift Spectrum external table queries:

```yaml
apiVersion: aws.openmcf.org/v1
kind: AwsRedshiftCluster
metadata:
  name: analytics-cluster
  labels:
    openmcf.org/provisioner: pulumi
    pulumi.openmcf.org/organization: my-org
    pulumi.openmcf.org/project: my-project
    pulumi.openmcf.org/stack.name: prod.AwsRedshiftCluster.analytics-cluster
spec:
  nodeType: ra3.4xlarge
  numberOfNodes: 4
  databaseName: datalake
  masterUsername: admin
  manageMasterPassword: true
  masterPasswordSecretKmsKeyId:
    value: "<kms-key-arn>"
  subnetIds:
    - value: "<private-subnet-id-az1>"
    - value: "<private-subnet-id-az2>"
  encrypted: true
  kmsKeyId:
    value: "<kms-key-arn>"
  enhancedVpcRouting: true
  multiAz: true
  automatedSnapshotRetentionPeriod: 14
  skipFinalSnapshot: false
  finalSnapshotIdentifier: analytics-cluster-final
  preferredMaintenanceWindow: "sun:02:00-sun:04:00"
  allowVersionUpgrade: true
  applyImmediately: false
  logging:
    logDestinationType: cloudwatch
    logExports:
      - connectionlog
      - useractivitylog
      - userlog
  parameters:
    - name: require_ssl
      value: "true"
    - name: enable_user_activity_logging
      value: "true"
    - name: max_concurrency_scaling_clusters
      value: "5"
  iamRoles:
    - value: "<redshift-s3-access-role-arn>"
    - value: "<redshift-spectrum-role-arn>"
  defaultIamRoleArn:
    value: "<redshift-s3-access-role-arn>"
```

## Stack Outputs

After deployment, the following outputs are available in `status.outputs`:

| Output | Type | Description |
|--------|------|-------------|
| `clusterIdentifier` | `string` | The unique identifier of the Redshift cluster |
| `clusterArn` | `string` | The Amazon Resource Name of the cluster, used for IAM policies and cross-service references |
| `clusterNamespaceArn` | `string` | The namespace ARN, used for Redshift data sharing and Serverless integration |
| `endpoint` | `string` | The connection endpoint in `address:port` format for SQL client connections |
| `dnsName` | `string` | The DNS hostname of the cluster (without port), for use in connection strings |
| `databaseName` | `string` | The name of the default database in the cluster |
| `port` | `int32` | The TCP port on which the cluster accepts connections |
| `subnetGroupName` | `string` | The name of the Redshift subnet group (only when created by this component) |
| `securityGroupId` | `string` | The ID of the managed security group (only when created by this component) |
| `parameterGroupName` | `string` | The name of the parameter group (only when created by this component) |
| `masterPasswordSecretArn` | `string` | The ARN of the Secrets Manager secret containing the master password (only when `manageMasterPassword` is `true`) |

## Related Components

- [AwsVpc](/docs/catalog/aws/awsvpc) — provides subnets and VPC ID for cluster placement
- [AwsSecurityGroup](/docs/catalog/aws/awssecuritygroup) — controls network access to the cluster
- [AwsKmsKey](/docs/catalog/aws/awskmskey) — provides KMS keys for cluster encryption and Secrets Manager
- [AwsIamRole](/docs/catalog/aws/awsiamrole) — provides IAM roles for S3 access, Spectrum, and other AWS service integrations
- [AwsS3Bucket](/docs/catalog/aws/awss3bucket) — stores data for COPY/UNLOAD operations and audit logs (when using S3 log destination)
