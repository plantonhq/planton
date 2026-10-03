# AWS Redshift Cluster

Deploys a managed Amazon Redshift data warehouse cluster with configurable node topology, availability posture (Multi-AZ or zone relocation), managed password storage in AWS Secrets Manager, KMS encryption, audit logging, cross-region snapshot copy, and VPC placement. Cost governance is first-class: usage limits cap what Spectrum and concurrency scaling may consume, and scheduled actions pause, resume, or resize the cluster on cron schedules.

## What Gets Created

When you deploy this Infra Component, the IaC module provisions:

- **Redshift Cluster** -- a columnar data warehouse cluster running the Redshift engine with the specified node type and count, placed in the configured subnets with configurable public or private access
- **Subnet Group** -- created from the provided `subnetIds` spanning at least two Availability Zones; skipped when an existing `clusterSubnetGroupName` is specified instead
- **Parameter Group** -- created only when inline `parameters` entries are configured; applies Redshift-specific tuning parameters (family `redshift-1.0` unless `parameterGroupFamily` selects another)
- **Logging Configuration** -- created only when `logging` is configured; delivers audit logs (connection, user activity, user DDL) to S3 or CloudWatch Logs
- **Snapshot Copy Configuration** -- created only when `snapshotCopy` names a destination region; automatically copies the cluster's snapshots there for disaster recovery
- **Snapshot Schedule Association** -- created only when `snapshotScheduleIdentifier` names an existing account-level schedule; replaces the default automated snapshot cadence
- **Usage Limits** -- one per `usageLimits[]` entry; feature-scoped spend caps with escalating breach actions
- **Scheduled Actions** -- one per `scheduledActions[]` entry; pause, resume, or resize on `cron()`/`at()` schedules
- **Managed VPC Endpoints and Cross-Account Grants** -- one endpoint per `endpointAccesses[]` entry (RA3 only) and one authorization per `endpointAuthorizations[]` entry
- **AWS Tags** -- resource metadata tags (organization, environment, resource kind, resource ID) applied automatically for tracking and governance

Security groups are composed, never created: the cluster attaches the referenced `securityGroupIds`, and the ingress rules that open the warehouse port belong on those first-class AwsSecurityGroup resources.

## Before You Deploy

### Planton Setup

- **AWS Provider Connection** -- an active connection in the Connect module with credentials for the target AWS account. Map it as the default for your environment, or specify it explicitly when creating the Infra Component.
- **Planton Runner** -- required when using Runner-based credential delivery. Not needed for inline credentials or cross-account trust authentication modes.

### AWS Account

- **At least two subnets** in distinct Availability Zones. Private subnets are recommended for production. Reference AwsSubnet Infra Components via ValueFromRef or provide subnet IDs directly. Alternatively, provide an existing `clusterSubnetGroupName`.
- **Security groups** (optional) governing who reaches the warehouse port. Reference AwsSecurityGroup Infra Components or provide security group IDs; empty keeps the VPC's default group.
- **IAM roles** (optional) for COPY, UNLOAD, and Redshift Spectrum queries that access S3, DynamoDB, or the Glue Data Catalog. Provide role ARNs directly or reference AwsIamRole Infra Components.
- **A KMS key** (optional) for customer-managed encryption of cluster data and the managed master password. Provide the ARN directly or reference an AwsKmsKey Infra Component.
- **An Elastic IP** (optional) for a stable public address on a publicly accessible cluster. Reference an AwsElasticIp Infra Component's `public_ip` output or provide the IP address directly.

## Deploy

### Console

Open the deployment store, find **AWS Redshift Cluster**, and click **Deploy**. The creation wizard walks you through preset selection, environment and connection configuration, and spec fields. Start from the **Single-Node Development Cluster** preset in the [Presets](#presets) tab to pre-populate a working configuration.

### CLI

Create a manifest and apply it:

```yaml
apiVersion: aws.planton.dev/v1alpha1
kind: AwsRedshiftCluster
metadata:
  name: analytics-warehouse
  org: acme-corp
  env: prod
spec:
  region: us-west-2
  nodeType: ra3.xlplus
  numberOfNodes: 2
  subnetIds:
    - value: "subnet-0a1b2c3d4e5f00001"
    - value: "subnet-0a1b2c3d4e5f00002"
  masterUsername: admin
  manageMasterPassword: true
  encrypted: true
  skipFinalSnapshot: false
  finalSnapshotIdentifier: analytics-warehouse-final
```

```shell
planton apply -f redshift-cluster.yaml
```

This creates a two-node RA3 Redshift cluster with a managed master password (stored in Secrets Manager), encrypted storage using the default AWS-managed key, and a final-snapshot deletion contract. An Infra Job tracks the provisioning in real time.

### InfraChart

When deploying as part of a multi-resource environment, use ValueFromRef to wire the Redshift cluster to subnets, an IAM role, and a KMS key deployed in the same InfraPipeline:

```yaml
spec:
  subnetIds:
    - valueFrom:
        kind: AwsSubnet
        name: warehouse-subnet-a
        fieldPath: status.outputs.subnet_id
    - valueFrom:
        kind: AwsSubnet
        name: warehouse-subnet-b
        fieldPath: status.outputs.subnet_id
  securityGroupIds:
    - valueFrom:
        kind: AwsSecurityGroup
        name: warehouse-sg
        fieldPath: status.outputs.security_group_id
  kmsKeyId:
    valueFrom:
      kind: AwsKmsKey
      name: warehouse-key
      fieldPath: status.outputs.key_arn
  iamRoles:
    - valueFrom:
        kind: AwsIamRole
        name: redshift-service-role
        fieldPath: status.outputs.role_arn
  defaultIamRoleArn:
    valueFrom:
      kind: AwsIamRole
      name: redshift-service-role
      fieldPath: status.outputs.role_arn
```

The InfraPipeline resolves the dependency graph, deploys the subnets, KMS key, and IAM role first, then provisions the Redshift cluster with the resolved values.

## Key Configuration

These are the most important decisions when configuring a Redshift cluster. Explore the full field reference in the [API Explorer](#api-explorer) tab.

**Node type and topology** -- Choose between dc2 (legacy dense compute, node-local SSD only) and ra3 (managed storage with automatic S3 tiering). RA3 nodes are the right call for nearly all new clusters -- compute and storage scale independently, and RA3 is the only family supporting Multi-AZ and zone relocation. Set `numberOfNodes` to 1 for a single-node development cluster or 2+ for multi-node production with a dedicated leader node. Resizing later is in-place, never a replacement.

**Availability posture** -- At most one of two mutually exclusive mechanisms: `multiAz: true` runs compute in two Availability Zones with automatic failover behind a single endpoint (requires RA3 and a multi-node cluster); `availabilityZoneRelocationEnabled: true` lets AWS move the single cluster to a healthy zone during outages (requires RA3 and a port within 5431-5455 or 8191-8215).

**Creation source** -- Leave the restore fields empty for a brand-new cluster, or restore from an existing snapshot with `snapshotIdentifier` (by name, optionally disambiguated by `snapshotClusterIdentifier`) or `snapshotArn` (the cross-account/cross-region share shape). A restored cluster inherits the source's credentials and databases, so `masterUsername` stays empty; `ownerAccount` identifies a snapshot shared by another AWS account.

**Password management** -- Set `manageMasterPassword: true` (recommended) to let AWS Secrets Manager store and rotate the master password automatically; the secret's ARN is exported as `master_password_secret_arn`. Optionally provide `masterPasswordSecretKmsKeyId` to encrypt the secret with a customer-managed KMS key. When `manageMasterPassword` is `false`, provide `masterPassword` as an org-secret reference.

**Audit logging** -- Configure `logging` with `logDestinationType` set to `"s3"` or `"cloudwatch"`. CloudWatch delivery exports exactly the log types you list in `logExports` (`connectionlog`, `useractivitylog`, `userlog`); S3 delivery always writes all three to the designated bucket. The user activity log only produces data when the `enable_user_activity_logging` parameter is set.

**Snapshots and disaster recovery** -- Set `automatedSnapshotRetentionPeriod` (0-35 days; production typically wants 7+) and `manualSnapshotRetentionPeriod` (-1 retains indefinitely). Keep `skipFinalSnapshot: false` with a `finalSnapshotIdentifier` so deletion always leaves a recovery handle. Configure `snapshotCopy` with a `destinationRegion` to automatically copy snapshots to a second region -- KMS-encrypted clusters additionally need a `snapshotCopyGrantName` in the destination region. Associate an existing account-level snapshot schedule with `snapshotScheduleIdentifier` (AWS keeps one schedule per cluster) to replace the default automated cadence.

**Cost governance** -- `usageLimits` cap what individual features may consume: Spectrum and cross-region datasharing take `data-scanned` (terabytes), concurrency scaling and automatic-optimization compute take `time` (minutes) -- AWS's own pairing, enforced at validation. Breach actions escalate from `log` through `emit-metric` to `disable`. `scheduledActions` pause, resume, or resize the cluster on `cron()`/`at()` schedules (UTC) -- the nights-and-weekends lever that stops compute billing while paused. Two contracts to know: the action's IAM role must trust `scheduler.redshift.amazonaws.com` (AWS validates at create), and action names are unique per AWS ACCOUNT, so include the cluster name in them.

**Cross-VPC and cross-account access** -- `endpointAccesses` create Redshift-managed VPC endpoints (RA3 only): each entry lands in its own subnet group (typically the consuming VPC's) or reuses the cluster's, and its private address is exported in `endpoint_access_addresses`. For consumers in OTHER AWS accounts, `endpointAuthorizations` grant each account permission to create endpoints from its own side -- one grant per account, optionally restricted to specific VPCs; `forceDelete` controls whether revoking the grant tears down the grantee's live endpoints.

## Outputs and Dependencies

### What This Kind Consumes

| Dependency | Field | ValueFromRef Path |
|------------|-------|-------------------|
| **AwsSubnet** | `subnetIds` | `status.outputs.subnet_id` |
| **AwsSecurityGroup** (optional) | `securityGroupIds` | `status.outputs.security_group_id` |
| **AwsElasticIp** (optional) | `elasticIp` | `status.outputs.public_ip` |
| **AwsKmsKey** (optional) | `kmsKeyId` | `status.outputs.key_arn` |
| **AwsKmsKey** (optional) | `masterPasswordSecretKmsKeyId` | `status.outputs.key_arn` |
| **AwsIamRole** (optional) | `iamRoles` | `status.outputs.role_arn` |
| **AwsIamRole** (optional) | `defaultIamRoleArn` | `status.outputs.role_arn` |
| **AwsIamRole** (optional) | `scheduledActions[].iamRoleArn` | `status.outputs.role_arn` |
| **AwsSecurityGroup** (optional) | `endpointAccesses[].vpcSecurityGroupIds` | `status.outputs.security_group_id` |
| **AwsVpc** (optional) | `endpointAuthorizations[].vpcIds` | `status.outputs.vpc_id` |

### What This Kind Provides

After provisioning, `status.outputs` contains values that downstream Infra Components can consume via ValueFromRef:

| Output | Description | Common Downstream Use |
|--------|-------------|----------------------|
| `cluster_identifier` | Unique identifier of the Redshift cluster | Monitoring dashboards, operational scripts |
| `cluster_arn` | Amazon Resource Name of the cluster | IAM policies, CloudWatch alarms, resource tagging |
| `cluster_namespace_arn` | Namespace ARN of the cluster | Redshift data sharing, the Redshift Data API |
| `endpoint` | Connection endpoint in address:port format | Application connection strings, BI tool configuration |
| `dns_name` | DNS hostname of the cluster's leader node (without port) | DNS CNAME records, connection strings |
| `database_name` | Name of the first database in the cluster | Application connection configuration |
| `port` | TCP port for client connections | Application connection configuration, security group rules |
| `subnet_group_name` | Redshift subnet group in use (module-managed or referenced) | Audit, related resource lookups |
| `parameter_group_name` | Cluster parameter group in use (module-managed or referenced) | Parameter auditing |
| `master_password_secret_arn` | Secrets Manager secret ARN (when the password is AWS-managed) | Application secret retrieval, rotation monitoring |
| `endpoint_access_addresses` | Private DNS addresses of managed VPC endpoints, keyed by endpoint name | Connection strings for consumers in other VPCs |
| `usage_limit_ids` | AWS-generated usage-limit IDs, keyed by feature/limit-type/period | Out-of-band CLI operations, state import |

## Common Patterns

Browse the [Presets](#presets) tab for ready-to-deploy configurations.

**Single-node development** -- An RA3 single-node cluster for development and prototyping with minimal cost. No Multi-AZ, no audit logging, managed master password. Start from the **Single-Node Development Cluster** preset.

**Multi-node production** -- RA3 nodes with managed master password, encrypted storage, 7-day automated snapshot retention, and deletion protection via final snapshot. Suitable for production analytics workloads. Start from the **Multi-Node Production Data Warehouse** preset.

**Analytics workload** -- Multi-node RA3 cluster with IAM roles attached for Spectrum queries, enhanced VPC routing for network governance, CloudWatch audit logging, cross-region snapshot copy, and Multi-AZ. Designed for data lake analytics with S3 integration. Start from the **High-Performance Analytics Cluster** preset.

**Governed warehouse** -- Production cluster with daily Spectrum and concurrency-scaling spend caps, scheduled nightly pause / morning resume, a managed VPC endpoint for a BI-tooling VPC, and a cross-account endpoint grant. Start from the **Governed Warehouse with Cost Controls and Cross-VPC Access** preset.

## Works With

- [**AWS Subnet**](/infra-catalog/aws-subnet) -- provides the subnets the Redshift subnet group is built from
- [**AWS Security Group**](/infra-catalog/aws-security-group) -- provides network access control for the cluster endpoint
- [**AWS Elastic IP**](/infra-catalog/aws-elastic-ip) -- provides a stable public address for a publicly accessible cluster
- [**AWS KMS Key**](/infra-catalog/aws-kms-key) -- provides a customer-managed key for cluster encryption and Secrets Manager password encryption
- [**AWS IAM Role**](/infra-catalog/aws-iam-role) -- provides service roles for COPY, UNLOAD, and Redshift Spectrum access to S3 and Glue, and the scheduler-trusting role scheduled actions assume
- [**AWS VPC**](/infra-catalog/aws-vpc) -- scopes cross-account endpoint authorizations to specific grantee VPCs
