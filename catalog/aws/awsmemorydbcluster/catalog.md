# AWS MemoryDB Cluster

Deploys a fully managed Amazon MemoryDB cluster -- a Redis-compatible, durable in-memory database with multi-AZ transaction log replication, microsecond reads, and single-digit millisecond writes. The cluster supports sharded topology, ACL-based authentication ([AwsMemorydbAcl](/infra-catalog/aws-memorydb-acl)), TLS encryption, customer-managed KMS keys, automatic snapshots, snapshot restore (MemoryDB snapshots or S3 RDB files), IPv6/dual-stack networking, multi-Region membership, and optional data tiering to SSD. Unlike ElastiCache, MemoryDB is a primary database: the multi-AZ transaction log makes every acknowledged write durable, so it can be the system of record rather than a cache in front of one.

## What Gets Created

When you deploy this Infra Component, the IaC module provisions:

- **MemoryDB Cluster** -- a sharded in-memory database cluster running Redis or Valkey with the specified node type, shard count, and replicas per shard, using ACL-based authentication
- **Subnet Group** -- created from the provided `subnetIds` spanning at least two Availability Zones for multi-AZ durability; alternatively the cluster joins an existing group via `subnetGroupName` (the two are mutually exclusive)
- **Parameter Group** -- created only when `parameters` entries are configured (with the required `parameterGroupFamily`); alternatively the cluster uses an existing group via `parameterGroupName` (mutually exclusive with inline parameters)
- **AWS Tags** -- resource metadata tags (organization, environment, resource kind, resource ID) applied automatically for tracking and governance

## Before You Deploy

### Planton Setup

- **AWS Provider Connection** -- an active connection in the Connect module with credentials for the target AWS account. Map it as the default for your environment, or specify it explicitly when creating the Infra Component.
- **Planton Runner** -- required when using Runner-based credential delivery. Not needed for inline credentials or cross-account trust authentication modes.

### AWS Account

- **At least two subnets** in distinct Availability Zones within the target VPC for multi-AZ durability. Private subnets are recommended. Provide subnet IDs directly or reference AwsSubnet Infra Components via ValueFromRef.
- **Security groups** (optional) to control network-level access to the MemoryDB endpoint. Must allow inbound traffic on the cluster port (default 6379). Provide security group IDs directly or reference AwsSecurityGroup Infra Components.
- **A KMS key** (optional) for customer-managed at-rest encryption. MemoryDB always encrypts data at rest; this field specifies your own key instead of the AWS-managed key. The KMS key is ForceNew and cannot be changed after creation. Provide the ARN directly or reference an AwsKmsKey Infra Component.
- **An SNS topic** (optional) for cluster event notifications (failover, maintenance, configuration changes). Provide the topic ARN directly or reference an AwsSnsTopic Infra Component.

## Deploy

### Console

Open the deployment store, find **AWS MemoryDB Cluster**, and click **Deploy**. The creation wizard walks you through preset selection, environment and connection configuration, and spec fields. Start from the **Dev Single Shard** preset in the [Presets](#presets) tab to pre-populate a working configuration.

### CLI

Create a manifest and apply it:

```yaml
apiVersion: aws.planton.dev/v1alpha1
kind: AwsMemorydbCluster
metadata:
  name: session-store
  org: acme-corp
  env: prod
spec:
  region: us-west-2
  engine: redis
  engineVersion: "7.1"
  nodeType: db.r7g.large
  numShards: 2
  numReplicasPerShard: 1
  aclName:
    value: open-access
  subnetIds:
    - value: "subnet-0a1b2c3d4e5f00001"
    - value: "subnet-0a1b2c3d4e5f00002"
  tlsEnabled: true
```

```shell
planton apply -f memorydb-cluster.yaml
```

This creates a two-shard MemoryDB cluster with one replica per shard (4 total nodes), TLS encryption enabled, the built-in `open-access` ACL stated explicitly (swap in a real ACL for production), and AWS-managed at-rest encryption. An Infra Job tracks the provisioning in real time.

### InfraChart

When deploying as part of a multi-resource environment, use ValueFromRef to wire the MemoryDB cluster to a VPC, security group, and KMS key deployed in the same InfraPipeline:

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
        name: memorydb-sg
        fieldPath: status.outputs.security_group_id
  aclName:
    valueFrom:
      kind: AwsMemorydbAcl
      name: prod-services
      fieldPath: status.outputs.acl_name
  kmsKeyArn:
    valueFrom:
      kind: AwsKmsKey
      name: memorydb-key
      fieldPath: status.outputs.key_arn
  snsTopicArn:
    valueFrom:
      kind: AwsSnsTopic
      name: infra-alerts
      fieldPath: status.outputs.topic_arn
```

The InfraPipeline resolves the dependency graph, deploys the subnets, security group, ACL, KMS key, and SNS topic first, then provisions the MemoryDB cluster with the resolved values.

## Key Configuration

These are the most important decisions when configuring a MemoryDB cluster. Explore the full field reference in the [API Explorer](#api-explorer) tab.

**Engine choice** -- Set `engine` to `"redis"` or `"valkey"` (open-source Redis-compatible alternative). Specify `engineVersion` (e.g., `"7.1"` for Redis, `"7.2"` for Valkey) or leave empty for the provider default. Unlike ElastiCache, MemoryDB provides full data durability through a multi-AZ transaction log.

**Topology** -- `numShards` controls data partitions (each holding a portion of the keyspace). `numReplicasPerShard` (0-5) controls read replicas within each shard. Total nodes = `numShards` x (`numReplicasPerShard` + 1). For production, use at least 2 shards with 1 replica each for high availability and read scaling.

**Authentication** -- Every cluster requires an `aclName`. Use the built-in `"open-access"` ACL for development (no authentication) or reference an [AwsMemorydbAcl](/infra-catalog/aws-memorydb-acl) whose member users ([AwsMemorydbUser](/infra-catalog/aws-memorydb-user)) carry per-application permissions for production. Swapping ACLs applies in place. When `tlsEnabled` is `false`, only `"open-access"` is allowed.

**Encryption** -- TLS is on by default, and omitting `tlsEnabled` is identical to setting it `true` (the provider enforces that default itself); explicit `false` is the only way to disable it. MemoryDB always encrypts data at rest; provide `kmsKeyArn` to use a customer-managed KMS key instead of the AWS-managed key. Both `tlsEnabled` and `kmsKeyArn` are ForceNew -- changing them destroys and recreates the cluster.

**Maintenance and snapshots** -- `maintenanceWindow` and `snapshotWindow` are one-way once applied: leaving them empty lets AWS assign a window, but removing them later keeps the current window rather than reverting -- change a window by naming the new one explicitly. `snapshotRetentionLimit: 0` (the default) disables automatic snapshots. `finalSnapshotName` (1-255 lowercase alphanumeric/hyphen characters, no consecutive or trailing hyphens) preserves the data as a snapshot when the cluster is deleted -- without it, deletion is final.

**Snapshot restore** -- seed a new cluster's data from a named MemoryDB snapshot (`snapshotName`) or S3-hosted RDB files (`snapshotArns`, S3 object ARNs -- the migrate-from-self-managed path). The two are mutually exclusive and both are create-time decisions.

**Network addressing** -- `networkType` (`ipv4`/`ipv6`/`dual_stack`, ForceNew) sets which IP families the cluster serves; `ipDiscovery` (updates in place) sets which family discovery commands return -- the dual-stack + flip-discovery combination migrates clients to IPv6 without replacement. `multiRegionClusterName` (ForceNew) joins an existing multi-Region cluster for active-active replication.

**Data tiering** -- Set `dataTiering: true` on db.r6gd.* node types to automatically move less-frequently-accessed data to SSD storage, reducing cost for large datasets. ForceNew -- cannot be changed after creation.

## Outputs and Dependencies

### What This Kind Consumes

| Dependency | Field | ValueFromRef Path |
|------------|-------|-------------------|
| **AwsSubnet** | `subnetIds` | `status.outputs.subnet_id` |
| **AwsMemorydbAcl** | `aclName` | `status.outputs.acl_name` |
| **AwsSecurityGroup** (optional) | `securityGroupIds` | `status.outputs.security_group_id` |
| **AwsKmsKey** (optional) | `kmsKeyArn` | `status.outputs.key_arn` |
| **AwsSnsTopic** (optional) | `snsTopicArn` | `status.outputs.topic_arn` |

### What This Kind Provides

After provisioning, `status.outputs` contains values that downstream Infra Components can consume via ValueFromRef:

| Output | Description | Common Downstream Use |
|--------|-------------|----------------------|
| `cluster_endpoint_address` | DNS address of the cluster endpoint | Application connection strings, Redis client configuration |
| `cluster_endpoint_port` | Port of the cluster endpoint | Application connection configuration |
| `cluster_arn` | Amazon Resource Name of the cluster | IAM policies, CloudWatch alarms, resource tagging |
| `cluster_name` | MemoryDB cluster name | Operational scripts, monitoring dashboards |
| `engine_patch_version` | Actual engine patch version running | Compatibility verification, patching audits |
| `subnet_group_name` | Subnet group name (if created) | Audit, related resource lookups |
| `parameter_group_name` | Parameter group name (if created) | Parameter auditing |

## Common Patterns

Browse the [Presets](#presets) tab for ready-to-deploy configurations.

**Development single-shard** -- A single-shard cluster with db.t4g.small nodes, one replica, TLS enabled, and the `open-access` ACL. Minimal cost for development and testing. Start from the **Dev Single Shard** preset.

**Production high-availability** -- Multi-shard cluster with db.r7g.large nodes, replicas per shard, TLS enabled, customer-managed KMS encryption, snapshot retention, and SNS event notifications. Designed for production workloads requiring durability and failover. Start from the **Production HA** preset.

**High-throughput** -- Multi-shard cluster with db.r7g.xlarge nodes, multiple replicas for read scaling, data tiering on db.r6gd.* node types, and custom parameters for memory management. Optimized for high-throughput workloads with large datasets. Start from the **High-Throughput with Data Tiering** preset.

## Works With

- [**AWS MemoryDB ACL**](/infra-catalog/aws-memorydb-acl) -- the access control list the cluster attaches; its member [**AWS MemoryDB User**](/infra-catalog/aws-memorydb-user) entries carry per-application permissions
- [**AWS Subnet**](/infra-catalog/aws-subnet) -- provides subnets for the MemoryDB subnet group across multiple Availability Zones
- [**AWS Security Group**](/infra-catalog/aws-security-group) -- provides network access control for the cluster endpoint
- [**AWS KMS Key**](/infra-catalog/aws-kms-key) -- provides a customer-managed key for at-rest encryption
- [**AWS SNS Topic**](/infra-catalog/aws-sns-topic) -- provides event notification delivery for cluster operations