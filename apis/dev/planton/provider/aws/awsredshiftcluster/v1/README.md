# AWS Redshift Cluster

Deploys an Amazon Redshift data warehouse cluster with optional managed networking,
security groups, parameter groups, and audit logging — everything needed for a
production-grade columnar analytics database.

## When to Use

Use an AWS Redshift Cluster to:

- **Run analytical queries**: Petabyte-scale columnar storage optimized for OLAP
  workloads, aggregations, and complex joins across large datasets.
- **Build a data warehouse**: Centralize data from S3, DynamoDB, RDS, and
  streaming sources into a single SQL-queryable warehouse.
- **Enable BI and reporting**: Connect Redshift to tools like QuickSight, Tableau,
  Looker, or dbt for dashboards and scheduled analytics.
- **Query data lakes**: Use Redshift Spectrum to query data in S3 without loading
  it into the cluster.
- **Run ETL pipelines**: Leverage COPY/UNLOAD for high-throughput data movement
  between S3 and Redshift.

## Prerequisites

- At least two private subnets in distinct Availability Zones (for the subnet group)
- A VPC ID (if providing security group IDs or CIDR blocks for managed SG creation)
- A KMS key (if using customer-managed encryption instead of the default service key)
- IAM roles (if the cluster needs to access S3, Glue, or other AWS services)

## Spec Fields

### Core

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `nodeType` | string | — (required) | Node type: `dc2.large`, `ra3.xlplus`, `ra3.4xlarge`, `ra3.16xlarge` |
| `numberOfNodes` | int32 | 1 | 1 = single-node; >1 = multi-node (separate leader + compute) |
| `databaseName` | string | `dev` | Name of the first database created in the cluster |
| `masterUsername` | string | `admin` | Admin user name (1-128 chars, starts with letter) |
| `masterPassword` | string | — | Explicit password (mutually exclusive with `manageMasterPassword`) |
| `manageMasterPassword` | bool | recommended: true | Delegate password to Secrets Manager (auto-rotate) |
| `port` | int32 | 5439 | TCP port for client connections (1115-65535) |

### Networking

| Field | Type | Description |
|-------|------|-------------|
| `subnetIds` | StringValueOrRef[] | Subnets for auto-created subnet group (≥ 2, distinct AZs) |
| `clusterSubnetGroupName` | StringValueOrRef | Use an existing subnet group instead of `subnetIds` |
| `securityGroupIds` | StringValueOrRef[] | Source SGs → creates managed SG with ingress rules on cluster port |
| `allowedCidrBlocks` | string[] | Source CIDRs → creates managed SG with ingress rules on cluster port |
| `associateSecurityGroupIds` | StringValueOrRef[] | Attach existing SGs directly to the cluster |
| `vpcId` | StringValueOrRef | Required when `securityGroupIds` or `allowedCidrBlocks` are set |
| `publiclyAccessible` | bool | Assign a public IP to the cluster |
| `enhancedVpcRouting` | bool | Force COPY/UNLOAD traffic through VPC (enables flow logs) |
| `multiAz` | bool | Multi-AZ deployment for automatic failover (RA3 nodes only) |

### Encryption

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `encrypted` | bool | true | At-rest encryption using service key or KMS |
| `kmsKeyId` | StringValueOrRef | — | Customer-managed KMS key ARN for encryption |

### IAM

| Field | Type | Description |
|-------|------|-------------|
| `iamRoles` | StringValueOrRef[] | IAM role ARNs attached to the cluster (max 10) |
| `defaultIamRoleArn` | StringValueOrRef | Default role for unqualified COPY/UNLOAD commands |

### Snapshots

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `automatedSnapshotRetentionPeriod` | int32 | 1 | Days to retain automated snapshots (0-35) |
| `skipFinalSnapshot` | bool | false | Skip final snapshot on deletion (dev/test only) |
| `finalSnapshotIdentifier` | string | — | Name for the final snapshot (required when skip is false) |

### Maintenance

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `preferredMaintenanceWindow` | string | — | Weekly maintenance window (`ddd:hh:mi-ddd:hh:mi`) |
| `allowVersionUpgrade` | bool | true | Auto-apply major version upgrades |
| `maintenanceTrackName` | string | — | `current` or `trailing` version track |
| `applyImmediately` | bool | false | Apply modifications immediately vs. next window |

### Logging

| Field | Type | Description |
|-------|------|-------------|
| `logging.logDestinationType` | string | `s3` or `cloudwatch` |
| `logging.s3BucketName` | string | S3 bucket for logs (required when type is `s3`) |
| `logging.s3KeyPrefix` | string | Optional prefix for S3 log objects |
| `logging.logExports` | string[] | Log types: `connectionlog`, `useractivitylog`, `userlog` |

### Parameter Group

| Field | Type | Description |
|-------|------|-------------|
| `clusterParameterGroupName` | string | Use an existing parameter group |
| `parameters` | Parameter[] | Inline parameters → creates a new group (family: `redshift-1.0`) |

## Stack Outputs

| Output | Description |
|--------|-------------|
| `cluster_identifier` | Unique identifier of the Redshift cluster |
| `cluster_arn` | ARN for IAM policies and cross-service references |
| `cluster_namespace_arn` | Namespace ARN for data sharing / Serverless integration |
| `endpoint` | Connection endpoint in `address:port` format |
| `dns_name` | DNS hostname (without port) |
| `database_name` | Name of the default database |
| `port` | TCP port for connections |
| `subnet_group_name` | Managed subnet group name (if created) |
| `security_group_id` | Managed security group ID (if created) |
| `parameter_group_name` | Managed parameter group name (if created) |
| `master_password_secret_arn` | Secrets Manager secret ARN (when `manageMasterPassword` is true) |

## How It Works

The AwsRedshiftCluster component bundles up to five AWS resources:

1. **`aws_redshift_cluster`** — The core cluster resource with compute, storage,
   networking, encryption, and IAM configuration.
2. **`aws_redshift_subnet_group`** (conditional) — Created when `subnetIds` are
   provided. Groups subnets across AZs for high availability.
3. **`aws_security_group`** (conditional) — Created when `securityGroupIds` or
   `allowedCidrBlocks` are provided. Adds ingress rules on the cluster port.
4. **`aws_redshift_parameter_group`** (conditional) — Created when inline
   `parameters` are provided (e.g., `require_ssl`, `enable_user_activity_logging`).
5. **`aws_redshift_logging`** (conditional) — Created when `logging` is configured.
   Routes audit logs to S3 or CloudWatch.

## Related Resources

- **AwsVpc** — VPC and subnets for cluster networking
- **AwsKmsKey** — Customer-managed encryption key
- **AwsIamRole** — Roles for S3/Glue/Spectrum access (COPY, UNLOAD, Spectrum)
- **AwsSecurityGroup** — Pre-existing security groups to attach
- **AwsS3Bucket** — Audit log destination or data lake for Spectrum queries

## References

- [Amazon Redshift Documentation](https://docs.aws.amazon.com/redshift/latest/mgmt/welcome.html)
- [Redshift Node Types](https://docs.aws.amazon.com/redshift/latest/mgmt/working-with-clusters.html#rs-node-type-info)
- [RA3 Managed Storage](https://docs.aws.amazon.com/redshift/latest/mgmt/working-with-clusters.html#rs-ra3-node-types)
- [Redshift Encryption](https://docs.aws.amazon.com/redshift/latest/mgmt/working-with-db-encryption.html)
- [Multi-AZ Deployment](https://docs.aws.amazon.com/redshift/latest/mgmt/managing-cluster-multi-az.html)
- [Audit Logging](https://docs.aws.amazon.com/redshift/latest/mgmt/db-auditing.html)
