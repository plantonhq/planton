# AWS Neptune Cluster

Deploys an Amazon Neptune graph database cluster — property graphs (Apache TinkerPop Gremlin, openCypher) and RDF (SPARQL) over shared cluster storage — with a per-instance compute fleet or Neptune Serverless, IAM database authentication, encrypted storage, automated backups, CloudWatch log exports, and engine parameter management. Neptune has no master username or password by design -- access is network reachability plus optional SigV4-signed IAM authentication -- and the compute is the `instances` list, each entry managed as its own provider resource so scaling readers never touches the cluster.

## What Gets Created

When you deploy this Cloud Resource, the IaC module provisions:

- **Neptune Cluster** -- the shared-storage brain: endpoints, backups, encryption, and engine lifecycle. The cluster identifier comes from `metadata.name`
- **Neptune Instances** -- one provider resource per `instances[]` entry, keyed by name: one writer (the lowest promotion tier) plus any readers. Adding or removing a reader is an in-place update, never a cluster replacement
- **Neptune Subnet Group** -- created from the provided `subnetIds` (at least two subnets in distinct Availability Zones); skipped when an existing `neptuneSubnetGroupName` is named instead
- **Cluster Parameter Group** -- created only when inline `parameters` are configured (mutually exclusive with `neptuneClusterParameterGroupName`); requires a pinned `engineVersion`, since the group's family derives from it
- **Instance Parameter Group** -- created only when inline `instanceParameters` are configured; the instance-level twin, applied to every folded instance that does not bring its own group
- **Custom Cluster Endpoints** -- one per `customEndpoints[]` entry: stable DNS names over a chosen subset of the cluster's instances
- **AWS Tags** -- resource metadata tags (organization, environment, resource kind, resource ID) applied automatically for tracking and governance

`iamRoles` entries attach directly on the cluster resource, letting Neptune reach other AWS services (the S3 bulk loader, Neptune ML with SageMaker), and `enabledCloudwatchLogsExports` entries (`audit`, `slowquery`) make Neptune deliver those logs to CloudWatch -- neither creates a separate module resource.

## Before You Deploy

### Planton Setup

- **AWS Provider Connection** -- an active connection in the Connect module with credentials for the target AWS account. Map it as the default for your environment, or specify it explicitly when creating the Cloud Resource.
- **Planton Runner** -- required when using Runner-based credential delivery. Not needed for inline credentials or cross-account trust authentication modes.

### AWS Account

- **At least two subnets** in distinct Availability Zones. Private subnets are recommended -- Neptune is VPC-only and never internet-facing at the cluster level. Provide subnet IDs directly or reference AwsSubnet Cloud Resources via ValueFromRef. Alternatively, provide an existing `neptuneSubnetGroupName`.
- **Security groups** (optional) governing who reaches the cluster port. Ingress rules live on the referenced AwsSecurityGroup nodes -- never inside this cluster. Empty keeps the VPC's default security group.
- **A KMS key** (optional) for storage encryption beyond the default AWS-managed key. Provide the ARN directly or reference an AwsKmsKey Cloud Resource. Encryption is create-time only.
- **IAM roles** (optional) for engine features that reach other AWS services. Required for bulk loading data from S3 and for Neptune ML.

## Deploy

### Console

Open the deployment store, find **AWS Neptune Cluster**, and click **Deploy**. The creation wizard walks you through the creation source (new cluster, snapshot restore, or read replica), placement, the compute fleet (provisioned or serverless), security, backups, and audit logging. Start from the **Production Neptune (Provisioned)** preset in the [Presets](#presets) tab for the writer-plus-reader production shape.

### CLI

Create a manifest and apply it:

```yaml
apiVersion: aws.planton.dev/v1alpha1
kind: AwsNeptuneCluster
metadata:
  name: knowledge-graph
  org: acme-corp
  env: prod
spec:
  region: us-west-2
  subnetIds:
    - value: "subnet-0a1b2c3d4e5f00001"
    - value: "subnet-0a1b2c3d4e5f00002"
  instances:
    - name: writer
      instanceClass: db.r6g.large
      promotionTier: 0
    - name: reader-1
      instanceClass: db.r6g.large
      promotionTier: 1
  storageEncrypted: true
  iamDatabaseAuthenticationEnabled: true
  deletionProtection: true
  backupRetentionPeriod: 7
  skipFinalSnapshot: false
  finalSnapshotIdentifier: knowledge-graph-final
```

```shell
planton apply -f neptune-cluster.yaml
```

This creates a two-instance cluster (a writer plus one reader for high availability) on AWS's current default engine version, with encrypted storage on the AWS-managed key, IAM database authentication, deletion protection, 7-day backup retention, and a named final snapshot on deletion. A Stack Job tracks the provisioning in real time.

### InfraChart

When deploying as part of a multi-resource environment, use ValueFromRef to wire the Neptune cluster to subnets, a security group, a KMS key, and an IAM role deployed in the same InfraPipeline:

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
        name: neptune-sg
        fieldPath: status.outputs.security_group_id
  kmsKeyId:
    valueFrom:
      kind: AwsKmsKey
      name: neptune-key
      fieldPath: status.outputs.key_arn
  iamRoles:
    - valueFrom:
        kind: AwsIamRole
        name: neptune-s3-loader
        fieldPath: status.outputs.role_arn
```

The InfraPipeline resolves the dependency graph, deploys the subnets, security group, KMS key, and IAM role first, then provisions the Neptune cluster with the resolved values.

## Key Configuration

These are the most important decisions when configuring a Neptune cluster. Explore the full field reference in the [API Explorer](#api-explorer) tab.

**Creation source** -- A brand-new cluster is the everyday path. `snapshotIdentifier` restores an existing cluster snapshot into a NEW cluster; `replicationSourceIdentifier` makes this cluster a read replica of another (promote it later by clearing the field); `globalClusterIdentifier` joins a Neptune global database for cross-region reads. Restores, replicas, and global-database members may start headless -- zero instances, compute attached later. A regular cluster requires at least one instance.

**The compute fleet** -- Each `instances[]` entry is one DB instance: the lowest promotion tier that is available becomes the writer, everything else serves reads from the shared storage. Graph traversals are memory-bound; the r6g/r6i memory-optimized families are the Neptune workhorses. For variable workloads use Neptune Serverless: every instance uses `instanceClass: db.serverless` and the cluster carries `serverlessV2Scaling` NCU bounds (1-128; each NCU is ~2 GiB of memory) -- provisioned and serverless instances cannot mix.

**Engine version** -- Empty keeps AWS's current default and never goes stale. Pin a version (e.g. `"1.4.5.1"`) in production so upgrades are deliberate. Major upgrades additionally require `allowMajorVersionUpgrade: true` and `neptuneInstanceParameterGroupName` (the instance-level group AWS applies during the upgrade).

**Authentication** -- Neptune has no master username or password by design. Access is network reachability (security groups) plus, with `iamDatabaseAuthenticationEnabled`, SigV4-signed requests from IAM identities -- auditable and native to Lambda, ECS task roles, and instance profiles.

**Audit logging is two halves** -- The `audit` entry in `enabledCloudwatchLogsExports` delivers the log to CloudWatch, and the `neptune_enable_audit_log` cluster parameter (set it in `parameters`) makes the engine produce it. Without the parameter, the export delivers nothing.

**Durability posture** -- Storage encryption defaults on and is create-time only (an unencrypted cluster can never be encrypted later). `backupRetentionPeriod` bounds point-in-time recovery (production wants 7+; the AWS default is 1 day). Keep `skipFinalSnapshot: false` with a named `finalSnapshotIdentifier`, and enable `deletionProtection`, so every destructive path is a deliberate two-step.

## Outputs and Dependencies

### What This Component Consumes

| Dependency | Field | ValueFromRef Path |
|------------|-------|-------------------|
| **AwsSubnet** | `subnetIds` | `status.outputs.subnet_id` |
| **AwsSecurityGroup** (optional) | `securityGroupIds` | `status.outputs.security_group_id` |
| **AwsKmsKey** (optional) | `kmsKeyId` | `status.outputs.key_arn` |
| **AwsIamRole** (optional) | `iamRoles` | `status.outputs.role_arn` |

### What This Component Provides

After provisioning, `status.outputs` contains values that downstream Cloud Resources can consume via ValueFromRef:

| Output | Description | Common Downstream Use |
|--------|-------------|----------------------|
| `endpoint` | Writer endpoint -- reads AND writes | Application connection strings for Gremlin/openCypher/SPARQL |
| `reader_endpoint` | Load-balances read-only queries across readers | Read-heavy query distribution |
| `port` | Port the cluster accepts connections on | Application connection configuration |
| `instance_endpoints` | Per-instance endpoints, ordered as declared in `instances` | Pinning a workload to a specific reader |
| `cluster_identifier` | The cluster identifier | Monitoring dashboards, CLI operations |
| `arn` | Amazon Resource Name of the cluster | IAM policies, CloudWatch alarms, resource tagging |
| `cluster_resource_id` | Immutable resource ID (survives identifier renames) | IAM database-authentication policies, CloudWatch dimensions |
| `hosted_zone_id` | Route 53 hosted zone ID of the cluster endpoints | DNS alias record creation |
| `custom_endpoint_addresses` | Custom endpoint DNS names keyed by `customEndpoints[].name` | Workload-scoped connection strings |

`engine_version_actual`, `neptune_subnet_group_name`, `neptune_cluster_parameter_group_name`, and `neptune_instance_parameter_group_name` are also exported -- they record the resolved engine version and the groups in use, for auditing rather than downstream wiring.

## Common Patterns

Browse the [Presets](#presets) tab for ready-to-deploy configurations.

**Production graph** -- A writer plus one reader on `db.r6g.large`, a pinned engine version, encrypted storage, IAM authentication, deletion protection, 7-day retention, and the audit log paired with its `neptune_enable_audit_log` parameter. Start from the **Production Neptune (Provisioned)** preset.

**Serverless** -- A single `db.serverless` writer with NCU bounds (e.g. 1-32) for spiky or unpredictable query loads; the minimum NCU is the idle cost floor (Neptune Serverless does not pause to zero). Start from the **Neptune Serverless** preset.

## Works With

- [**AWS Subnet**](/cloud-catalog/aws-subnet) -- provides the subnets for the Neptune subnet group across multiple Availability Zones
- [**AWS Security Group**](/cloud-catalog/aws-security-group) -- provides network access control for the cluster endpoint
- [**AWS KMS Key**](/cloud-catalog/aws-kms-key) -- provides a customer-managed key for storage encryption
- [**AWS IAM Role**](/cloud-catalog/aws-iam-role) -- provides service roles for the S3 bulk loader and Neptune ML
