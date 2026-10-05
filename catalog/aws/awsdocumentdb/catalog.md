# AWS DocumentDB

Deploys an Amazon DocumentDB cluster (MongoDB-compatible) — the shared-storage brain plus a per-instance compute fleet. The cluster owns endpoints, credentials, backups, encryption, and the engine lifecycle; the `instances` list is the compute serving queries (one writer plus any readers, each managed as its own resource keyed by name). Supports snapshot and point-in-time restores as first-class creation sources, DocumentDB Serverless capacity, AWS-managed master passwords in Secrets Manager, and ValueFromRef wiring to subnets, security groups, and KMS keys.

## What Gets Created

When you deploy this Infra Component, the IaC module provisions:

- **DocumentDB Cluster** -- the shared-storage cluster in the specified AWS region, created fresh, restored from a snapshot, or restored from another cluster's continuous backup (point-in-time, optionally as a copy-on-write fast clone)
- **DocumentDB Instances** -- one instance per `instances` entry; the lowest promotion tier that is available becomes the writer. Each entry is managed as its own provider resource keyed by name, so scaling readers never touches the cluster. Per-instance knobs cover the maintenance window, CA bundle, the CA-rotation restart deferral (`certificateRotationRestart`), snapshot tag copy, Performance Insights, and `applyImmediately`
- **DB Subnet Group** -- built from `subnetIds` (pure glue: a named list of subnets); skipped when an existing `dbSubnetGroupName` is provided instead
- **Serverless Capacity** -- configured only when `serverlessV2Scaling` is set; every instance then runs class `db.serverless` and scales independently within the DCU bounds
- **Managed Master Password** -- created only when `manageMasterUserPassword` is true (the recommended posture); AWS generates, stores, and rotates the password in Secrets Manager
- **Cluster Parameter Group** -- created only when inline `parameters` are configured (mutually exclusive with naming an existing `dbClusterParameterGroupName`)
- **CloudWatch Log Exports** -- configured only for the log types in `enabledCloudwatchLogsExports` (`audit`, `profiler`); each export also needs its matching cluster parameter to produce data
- **AWS Tags** -- resource metadata tags (organization, environment, resource kind, resource ID) applied automatically for tracking and governance

## Before You Deploy

### Planton Setup

- **AWS Provider Connection** -- an active connection in the Connect module with credentials for the target AWS account. Map it as the default for your environment, or specify it explicitly when creating the Infra Component.
- **Planton Runner** -- required when using Runner-based credential delivery. Not needed for inline credentials or cross-account trust authentication modes.

### AWS Account

- **At least two subnets** in distinct Availability Zones -- AWS rejects a subnet group covering fewer. Private subnets are the production posture. Reference AwsSubnet resources or pass literal subnet IDs; alternatively name an existing DB subnet group.
- **Security groups** (optional) -- empty keeps the VPC default group. Database ingress rules (port 27017 from your application tier) belong on the referenced AwsSecurityGroup resources.
- **A KMS key** (optional) -- storage encryption uses the AWS-managed aws/rds key unless a customer-managed key is referenced. Encryption is create-time only.

## Deploy

### Console

Open the deployment store, find **AWS DocumentDB**, and click **Deploy**. The creation wizard leads with the creation source (fresh, snapshot restore, or point-in-time restore — derived sources inherit credentials and skip that step), then walks placement, the compute fleet, and the operational posture. Start from the **Production DocumentDB (Managed Password)** preset in the [Presets](#presets) tab.

### CLI

Create a manifest and apply it:

```yaml
apiVersion: aws.planton.dev/v1alpha1
kind: AwsDocumentDb
metadata:
  name: orders-docdb
  org: acme-corp
  env: prod
spec:
  region: us-west-2
  subnetIds:
    - value: subnet-0a1b2c3d4e5f00001
    - value: subnet-0a1b2c3d4e5f00002
  masterUsername: docadmin
  manageMasterUserPassword: true
  storageEncrypted: true
  deletionProtection: true
  backupRetentionPeriod: 7
  skipFinalSnapshot: false
  finalSnapshotIdentifier: orders-docdb-final
  instances:
    - name: writer
      instanceClass: db.r6g.large
      promotionTier: 0
    - name: reader-1
      instanceClass: db.r6g.large
      promotionTier: 1
```

```shell
planton apply -f documentdb.yaml
```

This creates an encrypted two-instance cluster (writer + reader) with the master password managed in Secrets Manager, 7-day point-in-time recovery, deletion protection, and a named final snapshot. An Infra Job tracks the provisioning in real time.

### InfraChart

When deploying as part of a multi-resource environment, use ValueFromRef to wire the cluster to subnets, security groups, and a KMS key deployed in the same InfraPipeline:

```yaml
spec:
  subnetIds:
    - valueFrom:
        kind: AwsSubnet
        name: private-az1
        fieldPath: status.outputs.subnet_id
    - valueFrom:
        kind: AwsSubnet
        name: private-az2
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

The InfraPipeline resolves the dependency graph, deploys the subnets, security group, and KMS key first, then provisions the DocumentDB cluster with the resolved values.

## Key Configuration

These are the most important decisions when configuring a DocumentDB cluster. Explore the full field reference in the [API Explorer](#api-explorer) tab.

**Creation source** -- Fresh (the everyday path), `snapshotIdentifier` (restore a snapshot into a NEW cluster), or `restoreToPointInTime` (any second inside the source's retention window; `restoreType: copy-on-write` is a fast clone that shares storage with the source). Derived sources inherit the master credentials and may start with zero instances. `globalClusterIdentifier` is orthogonal: the first cluster joined becomes the global writer.

**Compute model** -- Provisioned classes (`db.r6g.large` is the workhorse) or DocumentDB Serverless: set `serverlessV2Scaling` DCU bounds and every instance runs `db.serverless`. The models cannot mix, and removing the bounds from a live cluster replaces it.

**Credentials** -- `masterUsername` is required for a fresh cluster (AWS has no default). Keep `manageMasterUserPassword: true`: the password lives only in Secrets Manager and its ARN is exported as `master_user_secret_arn`. A literal `masterPassword` is mutually exclusive and stored in IaC state.

**Encryption and lifecycle** -- `storageEncrypted` is create-time only (an unencrypted cluster can never be encrypted in place). `deletionProtection` plus the final-snapshot contract (`skipFinalSnapshot: false` requires `finalSnapshotIdentifier`) are what stand between a wrong command and data loss.

**Audit logging has two halves** -- the `audit` entry in `enabledCloudwatchLogsExports` AND the `audit_logs` cluster parameter; the export ships empty without the parameter. The same pairing applies to `profiler`. Inline `parameters` require a pinned `engineVersion` (the managed group's family derives from it).

## Outputs and Dependencies

### What This Kind Consumes

| Dependency | Field | ValueFromRef Path |
|------------|-------|-------------------|
| **AwsSubnet** | `subnetIds` | `status.outputs.subnet_id` |
| **AwsSecurityGroup** (optional) | `securityGroupIds` | `status.outputs.security_group_id` |
| **AwsKmsKey** (optional) | `kmsKeyId` | `status.outputs.key_arn` |
| **AwsKmsKey** (optional) | `instances[].performanceInsightsKmsKeyId` | `status.outputs.key_arn` |

### What This Kind Provides

After provisioning, `status.outputs` contains values that downstream Infra Components can consume via ValueFromRef:

| Output | Description | Common Downstream Use |
|--------|-------------|----------------------|
| `endpoint` | Writer endpoint — reads and writes | Application connection strings |
| `reader_endpoint` | Load-balanced read-only endpoint | Read-heavy application connection strings |
| `instance_endpoints` | Per-instance endpoints | Targeted connections, diagnostics |
| `master_user_secret_arn` | The AWS-managed master secret | Runtime credential fetch from Secrets Manager |
| `cluster_identifier` | The cluster identifier | Monitoring dashboards, CloudWatch alarms |
| `arn` | Amazon Resource Name | IAM policies, resource tagging |
| `cluster_resource_id` | Immutable cluster resource ID | IAM condition keys, CloudWatch dimensions |
| `engine_version_actual` | The running engine version | Upgrade auditing (set when the spec leaves the version to AWS) |
| `hosted_zone_id` | Route53 zone of the endpoints | DNS alias records |
| `db_subnet_group_name` | The DB subnet group in use | Audit, related resource lookups |
| `db_cluster_parameter_group_name` | The parameter group in use | Parameter auditing |

## Common Patterns

**Production writer-plus-reader** -- two provisioned instances on shared cluster storage, the reader carrying report and analytics traffic through `reader_endpoint` and doubling as the failover target (DocumentDB promotes it in seconds because it already reads from the same storage volume). Managed master password, encrypted storage, deletion protection, and audit logs to CloudWatch. Start from the **Production DocumentDB (Managed Password)** preset.

**Serverless for unknown or spiky demand** -- a single `db.serverless` instance scaling within `serverlessV2Scaling` DCU bounds, so there is no instance class to re-pick as the workload grows. Remember the floor: DocumentDB Serverless does not pause to zero, so `minCapacity` is the idle cost. Start from the **DocumentDB Serverless** preset.

**Prod-data staging clone** -- a second cluster created with `restoreToPointInTime` in `copy-on-write` mode: it shares storage with the source and pays only for divergence, giving staging an exact production dataset without a full copy. Derived clusters inherit credentials and may start headless, attaching instances later.

## Works With

- [**AWS Subnet**](/infra-catalog/aws-subnet) -- provides the subnets for the DB subnet group across multiple Availability Zones
- [**AWS Security Group**](/infra-catalog/aws-security-group) -- provides network access control for the cluster endpoint
- [**AWS KMS Key**](/infra-catalog/aws-kms-key) -- provides customer-managed keys for storage and Performance Insights encryption
