# AWS ElastiCache Serverless

Deploys an ElastiCache Serverless cache with consumption-based pricing and automatic scaling of both compute (ECPU) and storage (GB). Supports Redis, Valkey, and Memcached engines with configurable scaling limits, VPC placement, customer-managed encryption, snapshots, and Redis ACL authentication. AWS owns node scaling, replication, and patching entirely -- the manifest declares bounds and posture, never topology.

## What Gets Created

When you deploy this Infra Component, the IaC module provisions:

- **Serverless Cache** -- an ElastiCache Serverless cache using the specified engine (Redis, Valkey, or Memcached), with AWS managing all node scaling, replication, and patching automatically
- **Cache Usage Limits** -- created only when scaling bounds are provided; configures minimum and maximum limits for data storage (GB) and compute (ECPU/s)
- **VPC Endpoints** -- created only when `subnetIds` are provided; places the cache endpoint in the specified subnets with traffic controlled by attached security groups
- **At-Rest Encryption** -- uses the AWS-managed key by default; uses a customer-managed KMS key when `kmsKeyId` is provided
- **Automatic Snapshots** -- created only when `dailySnapshotTime` is provided; takes daily snapshots with configurable retention (Redis/Valkey engines only)
- **AWS Tags** -- resource metadata tags (organization, environment, resource kind, resource ID) applied automatically

## Before You Deploy

### Planton Setup

- **AWS Provider Connection** -- an active connection in the Connect module with credentials for the target AWS account. Map it as the default for your environment, or specify it explicitly when creating the Infra Component.
- **Planton Runner** -- required when using Runner-based credential delivery. Not needed for inline credentials or cross-account trust authentication modes.

### AWS Account

- **VPC subnets** (recommended) -- private subnets for the cache's VPC endpoints. Provide subnet IDs directly or reference AwsSubnet Infra Components via ValueFromRef. The `subnetIds` field is ForceNew -- changing subnets destroys and recreates the cache.
- **A security group** (recommended) -- controls network access to the cache endpoint (default port 6379 for Redis/Valkey, 11211 for Memcached). Provide the ID directly or reference an AwsSecurityGroup via ValueFromRef.
- **A KMS key** (optional) -- for customer-managed at-rest encryption. The `kmsKeyId` field is ForceNew -- changing the key destroys and recreates the cache.
- **A Redis ACL user group** (optional) -- for fine-grained access control (Redis/Valkey engines only).

## Deploy

### Console

Open the deployment store, find **AWS ElastiCache Serverless**, and click **Deploy**. The creation wizard walks you through preset selection, environment and connection configuration, and spec fields. Start from the **Redis Minimal** preset in the [Presets](#presets) tab to pre-populate a zero-configuration starting point.

### CLI

Create a manifest and apply it:

```yaml
apiVersion: aws.planton.dev/v1alpha1
kind: AwsServerlessElasticache
metadata:
  name: session-cache
  org: acme-corp
  env: prod
spec:
  region: us-west-2
  engine: redis
  majorEngineVersion: "7"
```

```shell
planton apply -f serverless-cache.yaml
```

This creates a Redis 7.x serverless cache with AWS-managed scaling defaults and encryption. No VPC placement, scaling limits, or snapshots are configured. An Infra Job tracks the provisioning in real time.

### InfraChart

When deploying as part of a multi-resource environment, use ValueFromRef to wire the cache to networking and encryption resources deployed in the same InfraPipeline:

```yaml
spec:
  subnetIds:
    - valueFrom:
        kind: AwsSubnet
        name: private-subnet-a
        fieldPath: status.outputs.subnet_id
    - valueFrom:
        kind: AwsSubnet
        name: private-subnet-b
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
```

The InfraPipeline resolves the dependency graph, deploys the subnets, security group, and KMS key first, then provisions the cache with the resolved values.

## Key Configuration

These are the most important decisions when configuring ElastiCache Serverless. Explore the full field reference in the [API Explorer](#api-explorer) tab.

**Engine selection** -- Choose `redis` for persistence, replication, and ACL authentication. Choose `valkey` as an open-source Redis-compatible alternative with the same features. Choose `memcached` for volatile caching with no persistence or authentication. Switching between Redis and Valkey is in-place; switching to/from Memcached forces recreation.

**Scaling limits** -- Configure `dataStorageMinGb`/`dataStorageMaxGb` (1-5000 GB) and `ecpuMin`/`ecpuMax` (1000-15000000 ECPU/s) to control auto-scaling bounds. Minimum values guarantee capacity; maximum values cap costs. Leave as zero to use AWS defaults.

**Networking** -- Place the cache in VPC subnets with security groups for production. `subnetIds`, `networkType`, and `kmsKeyId` are ForceNew -- design networking and encryption choices upfront. `networkType` selects the endpoint addressing family (`ipv4` default, `ipv6`, or `dual_stack` -- dual-stack requires subnets with both CIDR types). Without subnets, the cache uses default networking.

**Snapshots and migration** -- Configure `dailySnapshotTime` (UTC, format HH:mm) and `snapshotRetentionLimit` (0-35 days) for Redis/Valkey engines. To migrate from a node-based cluster, snapshot it and list the snapshot ARNs in `snapshotArnsToRestore` -- the new cache starts life seeded with that data (create-time only). Memcached has no persistence and supports none of these.

**Access control** -- Reference an AwsElasticacheUserGroup in `userGroupId` for Redis ACL command-level and key-pattern permissions (Redis/Valkey only -- a serverless cache accepts exactly one group). Memcached has no authentication; its access control is entirely security-group based.

## Outputs and Dependencies

### What This Kind Consumes

| Dependency | Field | ValueFromRef Path |
|------------|-------|-------------------|
| **AwsSubnet** (optional) | `subnetIds` | `status.outputs.subnet_id` |
| **AwsSecurityGroup** (optional) | `securityGroupIds` | `status.outputs.security_group_id` |
| **AwsKmsKey** (optional) | `kmsKeyId` | `status.outputs.key_arn` |
| **AwsElasticacheUserGroup** (optional) | `userGroupId` | `status.outputs.user_group_id` |

### What This Kind Provides

After provisioning, `status.outputs` contains values that downstream Infra Components can consume via ValueFromRef:

| Output | Description | Common Downstream Use |
|--------|-------------|----------------------|
| `arn` | Amazon Resource Name of the cache | IAM policies, cross-service permissions |
| `endpoint_address` | Primary connection endpoint DNS address | Application connection strings for read-write operations |
| `endpoint_port` | Port of the primary endpoint | Application connection configuration |
| `reader_endpoint_address` | Reader endpoint DNS (Redis/Valkey only) | Read replica distribution for read-heavy workloads |
| `reader_endpoint_port` | Port of the reader endpoint | Application read connection configuration |
| `full_engine_version` | Exact engine version (e.g., 7.1.0) | Version confirmation, compatibility checks |
| `name` | Cache name (matches metadata ID) | Data source lookups, downstream references |

## Common Patterns

Browse the [Presets](#presets) tab for ready-to-deploy configurations.

**Redis minimal** -- Zero-configuration Redis 7.x cache with AWS-managed defaults. Suitable for development, prototyping, and low-traffic applications. Start from the **Redis Minimal** preset.

**Memcached with scaling limits** -- Bounded Memcached cache for web application response caching or session storage with predictable costs. No persistence or authentication. Start from the **Memcached with Scaling Limits** preset.

**Redis production** -- Full-featured Redis cache with VPC isolation, customer-managed KMS encryption, daily snapshots with 14-day retention, Redis ACL authentication, and explicit scaling bounds. All cross-resource fields use ValueFromRef for InfraChart composition. Start from the **Redis Production** preset.

## Works With

- [**AWS Subnet**](/infra-catalog/aws-subnet) -- provides the subnets for cache endpoint VPC placement
- [**AWS Security Group**](/infra-catalog/aws-security-group) -- controls network-level access to the cache endpoint
- [**AWS KMS Key**](/infra-catalog/aws-kms-key) -- provides a customer-managed key for at-rest encryption
- [**AWS ElastiCache User Group**](/infra-catalog/aws-elasticache-user-group) -- provides Redis ACL access control for Redis and Valkey caches