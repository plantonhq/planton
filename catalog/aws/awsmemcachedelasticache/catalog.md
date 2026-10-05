# AWS ElastiCache Memcached

Deploys an ElastiCache cluster running Memcached with configurable node count, cross-AZ distribution, optional in-transit encryption, custom parameter groups, and SNS event notifications. Memcached provides a distributed volatile cache with horizontal scaling across up to 40 nodes -- no replication, no persistence, no authentication, so security is entirely network isolation and every node holds data that can vanish.

## What Gets Created

When you deploy this Infra Component, the IaC module provisions:

- **ElastiCache Cluster** -- a managed Memcached cluster in the specified AWS region with one or more cache nodes distributing keys via consistent hashing
- **Cache Nodes** -- one or more nodes based on `numCacheNodes` (range 1-40); each node holds a partition of the key space. Adding nodes is non-disruptive; removing nodes evicts keys hashed to the removed node
- **ElastiCache Subnet Group** -- created from the provided `subnetIds`, or reused when `subnetGroupName` names a group that already exists (the two are mutually exclusive); places nodes in the specified VPC subnets
- **Parameter Group** -- created only when `parameters` entries are configured (using the specified `parameterGroupFamily`), or reused when `parameterGroupName` names a group that already exists (the two are mutually exclusive); applies Memcached-specific tuning
- **AWS Tags** -- resource metadata tags (organization, environment, resource kind, resource ID) applied automatically for tracking and governance

## Before You Deploy

### Planton Setup

- **AWS Provider Connection** -- an active connection in the Connect module with credentials for the target AWS account. Map it as the default for your environment, or specify it explicitly when creating the Infra Component.
- **Planton Runner** -- required when using Runner-based credential delivery. Not needed for inline credentials or cross-account trust authentication modes.

### AWS Account

- **Subnets** in the target VPC for the ElastiCache subnet group. Provide at least two subnets in distinct Availability Zones when using `cross-az` mode. Provide subnet IDs directly or reference AwsSubnet Infra Components via ValueFromRef.
- **Security groups** to attach to the cluster nodes for network access control. Since Memcached has no built-in authentication, security groups are the primary access control mechanism. Provide security group IDs directly or reference an AwsSecurityGroup Infra Component.
- **An SNS topic** (optional) for cluster event notifications (node additions, removals, maintenance events).

## Deploy

### Console

Open the deployment store, find **AWS ElastiCache Memcached**, and click **Deploy**. The creation wizard walks you through environment and connection configuration, and spec fields. Configure the engine version, node type, node count, and network settings directly in the wizard, or start from the **Memcached Multi-Node Cross-AZ** preset in the [Presets](#presets) tab.

### CLI

Create a manifest and apply it:

```yaml
apiVersion: aws.planton.dev/v1alpha1
kind: AwsMemcachedElasticache
metadata:
  name: app-cache
  org: acme-corp
  env: prod
spec:
  region: us-west-2
  engineVersion: "1.6.22"
  nodeType: cache.r7g.large
  numCacheNodes: 3
  azMode: cross-az
  transitEncryptionEnabled: true
  subnetIds:
    - value: "subnet-0a1b2c3d4e5f00001"
    - value: "subnet-0a1b2c3d4e5f00002"
```

```shell
planton apply -f memcached-elasticache.yaml
```

This creates a 3-node Memcached 1.6.22 cluster distributed across multiple Availability Zones with in-transit encryption. No custom parameters or SNS notifications are configured. Memcached does not support encryption at rest, persistence, or authentication. An Infra Job tracks the provisioning in real time.

### InfraChart

When deploying as part of a multi-resource environment, use ValueFromRef to wire the Memcached cluster to subnets, a security group, and an SNS topic deployed in the same InfraPipeline:

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
  notificationTopicArn:
    valueFrom:
      kind: AwsSnsTopic
      name: infra-alerts
      fieldPath: status.outputs.topic_arn
```

The InfraPipeline resolves the dependency graph, deploys the subnets, security group, and SNS topic first, then provisions the Memcached cluster with the resolved values.

## Key Configuration

These are the most important decisions when configuring an ElastiCache Memcached cluster. Explore the full field reference in the [API Explorer](#api-explorer) tab.

**Engine version** -- Specify `engineVersion` with a three-part version string (e.g., `"1.6.22"`). In-transit encryption requires version 1.6.12 or later. Use the latest 1.6.x version for best performance and security support.

**Node type and scaling** -- Set `nodeType` to determine per-node CPU, memory, and network capacity. Memcached scales horizontally by adding nodes via `numCacheNodes` (1-40). Changing `nodeType` forces cluster recreation and loses all cached data -- Memcached does not support vertical scaling in-place. Size the initial node type to accommodate growth.

**AZ distribution** -- Set `azMode` to `"cross-az"` (requires `numCacheNodes > 1`) to distribute nodes across multiple Availability Zones for resilience. Use `"single-az"` for development or when AZ-level resilience is not required. For exact placement pick ONE control: `availabilityZone` pins every node to a single zone (AZ-local latency for a client fleet that lives in one zone; ForceNew), while `preferredAvailabilityZones` places nodes per-AZ in order. The two are mutually exclusive, and a single-AZ pin cannot combine with `cross-az` mode.

**Encryption** -- Enable `transitEncryptionEnabled` for TLS on all client connections. Memcached does not support encryption at rest. There is no built-in authentication -- security relies entirely on VPC network isolation via security groups and private subnets.

**Parameter tuning** -- Provide `parameterGroupFamily` (e.g., `"memcached1.6"`) with `parameters` entries to tune engine behavior, or attach an existing group with `parameterGroupName`. Common parameters include `chunk_size`, `max_simultaneous_connections`, and `binding_protocol`.

**IP stack** -- Set `networkType` (`ipv4`, `ipv6`, or `dual_stack`) to choose the cluster's protocol stack, and `ipDiscovery` (`ipv4` or `ipv6`) to control which address family cluster discovery returns to clients. Both default to IPv4 when unset; the network type is fixed at creation.

## Outputs and Dependencies

### What This Kind Consumes

| Dependency | Field | ValueFromRef Path |
|------------|-------|-------------------|
| **AwsSubnet** (optional) | `subnetIds` | `status.outputs.subnet_id` |
| **AwsSecurityGroup** (optional) | `securityGroupIds` | `status.outputs.security_group_id` |
| **AwsSnsTopic** (optional) | `notificationTopicArn` | `status.outputs.topic_arn` |

### What This Kind Provides

After provisioning, `status.outputs` contains values that downstream Infra Components can consume via ValueFromRef:

| Output | Description | Common Downstream Use |
|--------|-------------|----------------------|
| `cluster_id` | ElastiCache cluster identifier | AWS CLI/API operations, monitoring |
| `cluster_address` | Auto-discovery endpoint DNS name | Client auto-discovery of cluster nodes |
| `configuration_endpoint` | Full configuration endpoint (address:port) | Recommended connection endpoint for multi-node clusters |
| `arn` | Amazon Resource Name | IAM policies, cross-service permissions |
| `port` | Port the cluster accepts connections on | Application connection configuration |
| `subnet_group_name` | ElastiCache subnet group name | Audit, related resource lookups |
| `parameter_group_name` | Custom parameter group name (if created) | Parameter auditing |

## Common Patterns

Browse the [Presets](#presets) tab for ready-to-deploy configurations.

**Single-node development cache** -- One `cache.t3.micro` node on the latest engine version. The smallest footprint for development and testing. Start from the **Memcached Single Node (Development)** preset.

**Multi-node cross-AZ cache** -- Three nodes spread across Availability Zones with preferred AZ pinning. Keys hash across nodes, so an AZ loss costs only that node's share of the cache. Start from the **Memcached Multi-Node Cross-AZ** preset.

**Production encrypted cache** -- Three `cache.r7g.large` nodes, cross-AZ, TLS in transit, a maintenance window, and inline parameter tuning. Start from the **Memcached Production Encrypted** preset.

## Works With

- [**AWS Subnet**](/infra-catalog/aws-subnet) -- provides the subnets for the ElastiCache subnet group across multiple Availability Zones
- [**AWS Security Group**](/infra-catalog/aws-security-group) -- provides network access control for the Memcached endpoint (primary security mechanism since Memcached has no authentication)
- [**AWS SNS Topic**](/infra-catalog/aws-sns-topic) -- receives cluster event notifications for node changes and maintenance events