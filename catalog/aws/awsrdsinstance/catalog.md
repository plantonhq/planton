# AWS RDS Instance

Deploys a single-instance relational database on Amazon RDS supporting PostgreSQL, MySQL, MariaDB, Oracle, and SQL Server engines — with Multi-AZ failover, AWS-managed master credentials in Secrets Manager, storage encryption and autoscaling, read replicas and snapshot/point-in-time/S3-XtraBackup creation sources, instance-owned parameter and option groups, feature-scoped IAM role associations, Performance Insights and Enhanced Monitoring, Blue/Green deployments, and the extended-support opt-out. Aurora engines are the sibling AwsRdsCluster's territory — this kind covers everything RDS runs as a standalone instance, including the engines Aurora never will (Oracle and SQL Server).

## What Gets Created

When you deploy this Infra Component, the IaC module provisions:

- **RDS DB Instance** -- a managed relational database instance running the specified engine and version (or a read replica / snapshot restore / point-in-time restore of an existing one), placed in the configured VPC subnets with the selected instance class and storage
- **DB Subnet Group** -- created from the provided `subnetIds` when no existing `dbSubnetGroupName` is specified; groups subnets across Availability Zones for instance placement
- **DB Parameter Group** -- created only when inline `parameters` are configured; an instance-owned group whose family is derived from the pinned engine version
- **DB Option Group** -- created only when inline `options` are configured (Oracle and SQL Server features like TDE or native backup)
- **IAM Role Associations** -- one per `iamRoles[]` entry, each managed as its own association resource so feature roles attach and detach without touching the instance
- **AWS Tags** -- resource metadata tags (organization, environment, resource kind, resource ID) applied automatically for tracking and governance

When `manageMasterUserPassword` is `true` (the recommended posture), AWS itself generates, stores, and rotates the master password in Secrets Manager -- no secret resource appears in the module, and its ARN surfaces as the `master_user_secret_arn` output.

## Before You Deploy

### Planton Setup

- **AWS Provider Connection** -- an active connection in the Connect module with credentials for the target AWS account. Map it as the default for your environment, or specify it explicitly when creating the Infra Component.
- **Planton Runner** -- required when using Runner-based credential delivery. Not needed for inline credentials or cross-account trust authentication modes.

### AWS Account

- **At least two subnets** in distinct Availability Zones within the target VPC. Private subnets are recommended for production. Provide subnet IDs directly or reference AwsSubnet Infra Components via ValueFromRef. Alternatively, provide an existing `dbSubnetGroupName`.
- **A security group** (optional) to attach to the instance for network access control. Provide security group IDs directly or reference an AwsSecurityGroup Infra Component via ValueFromRef.
- **A KMS key** (optional) for encrypting instance storage with a customer-managed key. Provide the ARN directly or reference an AwsKmsKey Infra Component via ValueFromRef.
- **A master username** -- AWS has no default and rejects a blank one on a new instance. Prefer `manageMasterUserPassword: true` (AWS keeps the password in Secrets Manager); a supplied `password` must be an org-secret reference, never plaintext.

## Deploy

### Console

Open the deployment store, find **AWS RDS Instance**, and click **Deploy**. The creation wizard walks you through preset selection, environment and connection configuration, and spec fields. Start from the **PostgreSQL (Production Multi-AZ)** preset in the [Presets](#presets) tab to pre-populate a working configuration.

### CLI

Create a manifest and apply it:

```yaml
apiVersion: aws.planton.dev/v1alpha1
kind: AwsRdsInstance
metadata:
  name: app-database
  org: acme-corp
  env: prod
spec:
  region: us-west-2
  subnetIds:
    - value: "subnet-0a1b2c3d4e5f00001"
    - value: "subnet-0a1b2c3d4e5f00002"
  engine: postgres
  engineVersion: "16.4"
  instanceClass: db.m6g.large
  allocatedStorageGb: 50
  maxAllocatedStorageGb: 200
  storageType: gp3
  storageEncrypted: true
  username: dbadmin
  manageMasterUserPassword: true
  multiAz: true
  backupRetentionPeriod: 7
  deletionProtection: true
  skipFinalSnapshot: false
  finalSnapshotIdentifier: app-database-final
  performanceInsightsEnabled: true
```

```shell
planton apply -f rds-instance.yaml
```

This creates a Multi-AZ PostgreSQL instance with encrypted gp3 storage that autoscales to 200 GiB, an AWS-managed master password in Secrets Manager, 7-day backups, and deletion protection. An Infra Job tracks the provisioning in real time.

### InfraChart

When deploying as part of a multi-resource environment, use ValueFromRef to wire the RDS instance to subnets, a security group, and a KMS key deployed in the same InfraPipeline:

```yaml
spec:
  subnetIds:
    - valueFrom:
        kind: AwsSubnet
        name: db-subnet-az1
        fieldPath: status.outputs.subnet_id
    - valueFrom:
        kind: AwsSubnet
        name: db-subnet-az2
        fieldPath: status.outputs.subnet_id
  securityGroupIds:
    - valueFrom:
        kind: AwsSecurityGroup
        name: database-sg
        fieldPath: status.outputs.security_group_id
  kmsKeyId:
    valueFrom:
      kind: AwsKmsKey
      name: database-key
      fieldPath: status.outputs.key_arn
```

The InfraPipeline resolves the dependency graph, deploys the subnets, security group, and KMS key first, then provisions the RDS instance with the resolved values.

## Key Configuration

These are the most important decisions when configuring an RDS instance. Explore the full field reference in the [API Explorer](#api-explorer) tab.

**Engine choice** -- Set `engine` to `postgres`, `mysql`, `mariadb`, an `oracle-*` edition, or a `sqlserver-*` edition and pin `engineVersion` (empty floats with AWS's current default). The engine choice is immutable after creation. Aurora engines deploy through AwsRdsCluster instead.

**Creation source** -- A new instance configures everything; `replicateSourceDb` creates a read replica (engine, storage, and credentials inherited — clearing it later promotes the replica to standalone), `snapshotIdentifier` restores a snapshot, `restoreToPointInTime` restores another instance's continuous backup to any moment in its retention window, and `s3Import` restores a Percona XtraBackup from S3 (MySQL migration on-ramp). The four sources are mutually exclusive.

**Credentials** -- `manageMasterUserPassword: true` (recommended) keeps the master password in Secrets Manager, generated and rotated by AWS, exported as the `master_user_secret_arn` output. A supplied `password` is mutually exclusive with the managed strategy.

**Instance sizing** -- Choose `instanceClass` per workload (burstable `db.t4g.*` for dev, `db.m6g.*` general purpose, `db.r6g.*` memory-heavy). `allocatedStorageGb` sets the initial size; `maxAllocatedStorageGb` enables storage autoscaling; `storageType: gp3` tunes `iops` and `storageThroughput` independently.

**High availability** -- Set `multiAz: true` for production to deploy a synchronous standby replica in a different Availability Zone. Failover is automatic and typically completes within 60 seconds. Multi-AZ roughly doubles the instance cost.

**Encryption** -- Set `storageEncrypted: true` to encrypt data at rest (create-time only — it can never be enabled in place). Optionally provide `kmsKeyId` for a customer-managed KMS key instead of the default AWS-managed key.

**Observability** -- `performanceInsightsEnabled` is free at the 7-day retention; `monitoringInterval` + `monitoringRoleArn` stream OS-level metrics; `enabledCloudwatchLogsExports` ships the engine's own logs (legal types vary by engine family).

**Engine configuration and roles** -- Inline `parameters` and `options` manage instance-owned parameter and option groups (family and major version derived from the pinned engine version; mutually exclusive with `parameterGroupName`/`optionGroupName`); `iamRoles` attaches feature-scoped engine roles (one association per entry, `featureName` required and engine-specific — e.g. `s3Import`, `s3Export`, `Lambda`; the role's trust policy must allow `rds.amazonaws.com` to assume it — AWS validates that server-side at association time and rejects the deploy otherwise); `engineLifecycleSupport: open-source-rds-extended-support-disabled` opts out of paid extended support.

## Outputs and Dependencies

### What This Kind Consumes

| Dependency | Field | ValueFromRef Path |
|------------|-------|-------------------|
| **AwsSubnet** | `subnetIds` | `status.outputs.subnet_id` |
| **AwsSecurityGroup** (optional) | `securityGroupIds` | `status.outputs.security_group_id` |
| **AwsKmsKey** (optional) | `kmsKeyId`, `masterUserSecretKmsKeyId`, `performanceInsightsKmsKeyId` | `status.outputs.key_arn` |
| **AwsIamRole** (optional) | `monitoringRoleArn`, `iamRoles[].role`, `s3Import.ingestionRole` | `status.outputs.role_arn` |
| **AwsSecurityGroup** (optional, options) | `options[].vpcSecurityGroupMemberships` | `status.outputs.security_group_id` |

### What This Kind Provides

After provisioning, `status.outputs` contains values that downstream Infra Components can consume via ValueFromRef:

| Output | Description | Common Downstream Use |
|--------|-------------|----------------------|
| `endpoint` | `address:port` connection endpoint | Application connection strings |
| `address` | Instance hostname | DNS CNAME records |
| `port` | Port the instance accepts connections on | Application connection configuration |
| `instance_identifier` | RDS instance identifier | Monitoring dashboards, CloudWatch alarms |
| `arn` | Amazon Resource Name | IAM policies, resource tagging |
| `resource_id` | Immutable dbi-resource ID | Point-in-time restore sources, IAM auth policies |
| `hosted_zone_id` | The endpoint's Route 53 hosted zone | Alias records |
| `master_user_secret_arn` | Secrets Manager ARN of the AWS-managed master secret | Application credential reads |

`engine_version_actual`, `db_subnet_group_name`, `db_parameter_group_name`, and `option_group_name` are also exported -- they record the resolved engine version and the groups in use, for auditing rather than downstream wiring.

## Common Patterns

Browse the [Presets](#presets) tab for ready-to-deploy configurations.

**PostgreSQL production** -- Multi-AZ deployment with encrypted, autoscaling gp3 storage, an AWS-managed master password, 7-day backups, and Performance Insights. Start from the **PostgreSQL (Production Multi-AZ)** preset.

**MySQL production** -- Same resilience posture as PostgreSQL but configured for MySQL 8.0. Suitable for applications requiring MySQL-compatible SQL or migrating from on-premises MySQL. Start from the **MySQL (Production Multi-AZ)** preset.

**Read replica** -- Scales reads and stages cross-region DR; engine, storage, and credentials inherit from the source. Start from the **Read Replica** preset.

## Works With

- [**AWS Subnet**](/infra-catalog/aws-subnet) -- provides the subnets for the DB subnet group across multiple Availability Zones
- [**AWS Security Group**](/infra-catalog/aws-security-group) -- provides network access control for the instance endpoint
- [**AWS KMS Key**](/infra-catalog/aws-kms-key) -- provides customer-managed keys for storage, the master secret, and Performance Insights
- [**AWS IAM Role**](/infra-catalog/aws-iam-role) -- the role Enhanced Monitoring publishes through, plus feature-scoped engine roles and the S3 import ingestion role
- [**AWS RDS Cluster**](/infra-catalog/aws-rds-cluster) -- the cluster sibling for Aurora engines and Multi-AZ community clusters