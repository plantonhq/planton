# AWS ElastiCache Redis

Deploys an ElastiCache replication group running Redis or Valkey with configurable topology (non-clustered or clustered mode), automatic failover, multi-AZ deployment, at-rest and in-transit encryption, AUTH token or Redis ACL authentication, custom parameter groups, and log delivery to CloudWatch or Kinesis. ElastiCache is a cache, not a system of record -- for durable Redis-compatible storage, the sibling AwsMemorydbCluster carries a multi-AZ transaction log.

## What Gets Created

When you deploy this Infra Component, the IaC module provisions:

- **ElastiCache Replication Group** -- a managed Redis or Valkey replication group in the specified AWS region, in either non-clustered mode (primary + read replicas) or clustered mode (sharded with data partitioning)
- **Cache Nodes** -- one or more nodes based on topology: `numCacheClusters` nodes for non-clustered mode, or `numNodeGroups` shards each with `replicasPerNodeGroup` replicas for clustered mode
- **ElastiCache Subnet Group** -- created from the provided `subnetIds`; places nodes in the specified VPC subnets
- **Parameter Group** -- created only when `parameters` entries are configured; applies engine-specific tuning using the specified `parameterGroupFamily`
- **Log Delivery** -- configured only when `logDeliveryConfigurations` entries are specified; delivers slow-log and/or engine-log to CloudWatch Logs or Kinesis Data Firehose
- **AWS Tags** -- resource metadata tags (organization, environment, resource kind, resource ID) applied automatically for tracking and governance

## Before You Deploy

### Planton Setup

- **AWS Provider Connection** -- an active connection in the Connect module with credentials for the target AWS account. Map it as the default for your environment, or specify it explicitly when creating the Infra Component.
- **Planton Runner** -- required when using Runner-based credential delivery. Not needed for inline credentials or cross-account trust authentication modes.

### AWS Account

- **Subnets** in the target VPC for the ElastiCache subnet group. Provide at least two subnets in distinct Availability Zones for multi-AZ deployments. Provide subnet IDs directly or reference AwsSubnet Infra Components via ValueFromRef.
- **Security groups** to attach to the cluster nodes for network access control. Provide security group IDs directly or reference an AwsSecurityGroup Infra Component.
- **A KMS key** (optional) for at-rest encryption with a customer-managed key instead of the default AWS-managed key. This is a ForceNew attribute -- changing it after creation destroys and recreates the cluster.
- **An SNS topic** (optional) for cluster event notifications (failover, maintenance, configuration changes).

## Deploy

### Console

Open the deployment store, find **AWS ElastiCache Redis**, and click **Deploy**. The creation wizard walks you through environment and connection configuration, and spec fields. Configure the engine, topology mode, node type, and encryption settings directly in the wizard, or start from the **Redis HA Cluster** preset in the [Presets](#presets) tab.

### CLI

Create a manifest and apply it:

```yaml
apiVersion: aws.planton.dev/v1alpha1
kind: AwsRedisElasticache
metadata:
  name: session-cache
  org: acme-corp
  env: prod
spec:
  region: us-west-2
  engine: redis
  engineVersion: "7.1"
  description: "Session cache for web application"
  nodeType: cache.r7g.large
  numCacheClusters: 3
  automaticFailoverEnabled: true
  multiAzEnabled: true
  atRestEncryptionEnabled: true
  transitEncryptionEnabled: true
  subnetIds:
    - value: "subnet-0a1b2c3d4e5f00001"
    - value: "subnet-0a1b2c3d4e5f00002"
  snapshotRetentionLimit: 7
```

```shell
planton apply -f redis-elasticache.yaml
```

This creates a non-clustered Redis 7.1 replication group with 1 primary and 2 read replicas, automatic failover across multiple AZs, encryption at rest and in transit, and 7-day snapshot retention. No AUTH token, custom parameters, or log delivery are configured. An Infra Job tracks the provisioning in real time.

### InfraChart

When deploying as part of a multi-resource environment, use ValueFromRef to wire the Redis cluster to subnets, a security group, and a KMS key deployed in the same InfraPipeline:

```yaml
spec:
  subnetIds:
    - valueFrom:
        kind: AwsSubnet
        name: cache-subnet-az1
        fieldPath: status.outputs.subnet_id
    - valueFrom:
        kind: AwsSubnet
        name: cache-subnet-az2
        fieldPath: status.outputs.subnet_id
  securityGroupIds:
    - valueFrom:
        kind: AwsSecurityGroup
        name: cache-sg
        fieldPath: status.outputs.security_group_id
  kmsKeyId:
    valueFrom:
      kind: AwsKmsKey
      name: cache-key
      fieldPath: status.outputs.key_arn
  notificationTopicArn:
    valueFrom:
      kind: AwsSnsTopic
      name: infra-alerts
      fieldPath: status.outputs.topic_arn
```

The InfraPipeline resolves the dependency graph, deploys the subnets, security group, KMS key, and SNS topic first, then provisions the Redis cluster with the resolved values.

## Key Configuration

These are the most important decisions when configuring an ElastiCache Redis cluster. Explore the full field reference in the [API Explorer](#api-explorer) tab.

**Engine and version** -- Set `engine` to `"redis"` or `"valkey"` (open-source Redis-compatible alternative). Specify `engineVersion` (e.g., `"7.1"` for Redis, `"7.2"` for Valkey). Leave `engineVersion` empty to use the provider default.

**Topology mode** -- Choose between non-clustered mode (`numCacheClusters`: 1-6 nodes with a single primary and read replicas) and clustered mode (`numNodeGroups` shards with `replicasPerNodeGroup`). Non-clustered suits workloads under ~113 GB; clustered mode supports multi-terabyte datasets and write scaling through data partitioning.

**High availability** -- Enable `automaticFailoverEnabled` for automatic promotion of a read replica when the primary fails (requires at least 2 nodes). Enable `multiAzEnabled` to distribute replicas across Availability Zones for resilience against AZ-level failures.

**Encryption** -- `atRestEncryptionEnabled` and `kmsKeyId` are ForceNew -- design these choices upfront. Enable `transitEncryptionEnabled` for TLS on all connections. Set `transitEncryptionMode` to `"required"` for full enforcement or `"preferred"` during migration from unencrypted connections.

**Authentication** -- Choose between `authToken` (Redis AUTH password, requires TLS) and `userGroupIds` (Redis ACL user groups for fine-grained command and key permissions — each entry accepts a group ID or a ValueFromRef to an AwsElasticacheUserGroup). These are mutually exclusive. When rotating a token on a running cluster, set `authTokenUpdateStrategy` (`ROTATE` for zero-downtime dual-token rollout, `SET` for immediate replacement). To migrate off AUTH entirely — the step before adopting ACL user groups — set `authTokenUpdateStrategy: DELETE` with NO `authToken`: DELETE removes the token, so supplying one alongside it is rejected.

**Global datastore** -- Set `globalReplicationGroupId` to join an existing global datastore as a cross-region secondary. The secondary inherits the primary's engine, version, node type, encryption, and parameter group — leave those fields UNSET (an explicit `false` on the encryption flags is just as illegal as `true`; the inherited settings reject the fields' presence, not just particular values) — and may only configure its own `numCacheClusters`. Start from the **Global Datastore Secondary (Cross-Region DR)** preset.

**Placement** -- In non-clustered mode, `preferredCacheClusterAzs` pins each node to an AZ (one entry per node). In clustered mode, `nodeGroupConfigurations` pins individual shards (primary AZ, replica AZs, per-shard replica count, keyspace slots). The two are mutually exclusive placement controls. `durability` (Valkey 9.0+, clustered) selects per-shard write acknowledgement (`async`/`sync`), and `clusterMode` (`enabled`/`compatible`/`disabled`) bridges live protocol migrations.

**Bring-your-own groups** -- Provide `subnetIds` to have a subnet group created, or name an existing one with `subnetGroupName` (mutually exclusive). Likewise `parameters` + `parameterGroupFamily` create a parameter group, or `parameterGroupName` attaches an existing one.

**Seed data (create-only)** -- Start the cluster from existing data with `snapshotArns` (S3 RDB files, one per shard) or `snapshotName` (an ElastiCache snapshot) — mutually exclusive, and not applicable when joining a global datastore.

**IP stack** -- `networkType` (`ipv4`, `ipv6`, `dual_stack`) fixes the protocol stack at creation; `ipDiscovery` controls which address family cluster discovery returns.

## Outputs and Dependencies

### What This Kind Consumes

| Dependency | Field | ValueFromRef Path |
|------------|-------|-------------------|
| **AwsSubnet** (optional) | `subnetIds` | `status.outputs.subnet_id` |
| **AwsSecurityGroup** (optional) | `securityGroupIds` | `status.outputs.security_group_id` |
| **AwsKmsKey** (optional) | `kmsKeyId` | `status.outputs.key_arn` |
| **AwsSnsTopic** (optional) | `notificationTopicArn` | `status.outputs.topic_arn` |
| **AwsElasticacheUserGroup** (optional) | `userGroupIds` | `status.outputs.user_group_id` |

### What This Kind Provides

After provisioning, `status.outputs` contains values that downstream Infra Components can consume via ValueFromRef:

| Output | Description | Common Downstream Use |
|--------|-------------|----------------------|
| `replication_group_id` | Replication group identifier | AWS CLI/API operations, monitoring |
| `primary_endpoint_address` | Primary (writer) endpoint DNS name | Application connection strings for read-write operations |
| `reader_endpoint_address` | Reader endpoint DNS name | Read-heavy application connection strings |
| `configuration_endpoint_address` | Cluster Mode configuration endpoint | Cluster-aware Redis client slot discovery |
| `arn` | Amazon Resource Name | IAM policies, cross-service permissions |
| `port` | Port the cluster accepts connections on | Application connection configuration |
| `subnet_group_name` | ElastiCache subnet group name | Audit, related resource lookups |
| `parameter_group_name` | Custom parameter group name (if created) | Parameter auditing |
| `engine_version_actual` | The engine version actually running (resolved from an unpinned or upgraded version) | Compatibility checks, upgrade auditing |

## Common Patterns

Browse the [Presets](#presets) tab for ready-to-deploy configurations.

**Single-node development cache** -- One node with at-rest and in-transit encryption on. The smallest footprint for development and testing. Start from the **Redis Single Node** preset.

**Highly available cache** -- Three nodes (1 primary + 2 replicas) with automatic failover, Multi-AZ, enforced TLS, snapshots, and a maintenance window. Start from the **Redis HA Cluster** preset.

**Clustered production cache** -- Three shards with two replicas each, customer-managed KMS encryption, slow-log delivery, and SNS notifications. Start from the **Redis Clustered Production** preset.

## Works With

- [**AWS Subnet**](/infra-catalog/aws-subnet) -- provides the subnets for the ElastiCache subnet group across multiple Availability Zones
- [**AWS Security Group**](/infra-catalog/aws-security-group) -- provides network access control for the Redis endpoint
- [**AWS KMS Key**](/infra-catalog/aws-kms-key) -- provides a customer-managed key for at-rest encryption
- [**AWS SNS Topic**](/infra-catalog/aws-sns-topic) -- receives cluster event notifications for failover and maintenance events
- [**AWS ElastiCache User Group**](/infra-catalog/aws-elasticache-user-group) -- provides Redis ACL user groups for fine-grained authentication