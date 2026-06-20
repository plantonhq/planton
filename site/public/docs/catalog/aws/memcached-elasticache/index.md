---
title: "Memcached ElastiCache"
description: "Memcached ElastiCache deployment documentation"
icon: "package"
order: 100
componentName: "awsmemcachedelasticache"
---

# AWS Memcached ElastiCache

Deploys a fully managed AWS ElastiCache cluster running the Memcached engine with automatic subnet group and parameter group management. Memcached provides a simple, high-throughput distributed cache using consistent hashing across nodes, with no replication, no persistence, and no authentication — security relies entirely on VPC network isolation.

## What Gets Created

When you deploy an AwsMemcachedElasticache resource, OpenMCF provisions:

- **ElastiCache Memcached Cluster** — an `aws_elasticache_cluster` with the `memcached` engine, placed in the specified subnets with attached security groups
- **Subnet Group** — created automatically when `subnetIds` are provided, grouping the subnets for cluster node placement
- **Parameter Group** — created automatically when `parameters` are provided along with a `parameterGroupFamily`, enabling custom Memcached engine tuning

## Prerequisites

- **AWS credentials** configured via environment variables or OpenMCF provider config
- **A VPC with private subnets** for cluster node placement (subnets in at least two AZs when using `cross-az` mode)
- **A security group** allowing inbound traffic on the Memcached port (default 11211) — since Memcached has no authentication, security groups are the primary access control mechanism

## Quick Start

Create a file `memcached.yaml`:

```yaml
apiVersion: aws.openmcf.org/v1
kind: AwsMemcachedElasticache
metadata:
  name: my-cache
  labels:
    openmcf.org/provisioner: pulumi
    pulumi.openmcf.org/organization: my-org
    pulumi.openmcf.org/project: my-project
    pulumi.openmcf.org/stack.name: dev.AwsMemcachedElasticache.my-cache
spec:
  region: us-west-2
  engineVersion: "1.6.22"
  nodeType: cache.t3.micro
  numCacheNodes: 1
  subnetIds:
    - subnet-0a1b2c3d4e5f00001
    - subnet-0a1b2c3d4e5f00002
  securityGroupIds:
    - sg-0a1b2c3d4e5f00001
```

Deploy:

```shell
openmcf apply -f memcached.yaml
```

This creates a single-node Memcached cluster on port 11211 in the specified subnets.

## Configuration Reference

### Required Fields

| Field | Type | Description | Validation |
|-------|------|-------------|------------|
| `region` | `string` | AWS region where the ElastiCache cluster will be deployed (e.g., `"us-west-2"`, `"us-east-1"`). | Required |
| `engineVersion` | `string` | Memcached engine version. Uses three-part versioning (e.g., `"1.6.22"`, `"1.6.17"`, `"1.5.16"`). Transit encryption requires `1.6.12` or later. | Required |
| `nodeType` | `string` | ElastiCache node type determining CPU, memory, and network capacity. Examples: `cache.t3.micro` (dev), `cache.r7g.large` (production). Changing this forces cluster recreation. | Required |

### Optional Fields

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `numCacheNodes` | `int` | `1` | Number of cache nodes in the cluster (1–40). Keys are distributed across nodes via consistent hashing. |
| `azMode` | `string` | `"single-az"` | AZ distribution mode. `"single-az"` places all nodes in one AZ. `"cross-az"` distributes across AZs (requires `numCacheNodes` > 1). |
| `port` | `int` | `11211` | Port the cluster accepts connections on. ForceNew — changing this destroys and recreates the cluster. |
| `transitEncryptionEnabled` | `bool` | `false` | Enable TLS encryption for all client connections. Requires engine version `1.6.12` or later. Memcached does not support encryption at rest. |
| `subnetIds` | `StringValueOrRef[]` | `[]` | Subnet IDs for the ElastiCache subnet group. Provide subnets in at least two AZs when using `cross-az` mode. Can reference AwsVpc resources via `valueFrom`. |
| `securityGroupIds` | `StringValueOrRef[]` | `[]` | VPC security groups to attach to the cluster nodes. Can reference AwsSecurityGroup resources via `valueFrom`. |
| `parameterGroupFamily` | `string` | — | Parameter group family (e.g., `"memcached1.6"`, `"memcached1.5"`). Required when `parameters` is provided. |
| `parameters[].name` | `string` | — | Parameter name (e.g., `"chunk_size"`, `"binding_protocol"`). |
| `parameters[].value` | `string` | — | Parameter value (e.g., `"96"`, `"auto"`). |
| `maintenanceWindow` | `string` | AWS-assigned | Weekly maintenance window in UTC. Format: `"ddd:hh24:mi-ddd:hh24:mi"` (e.g., `"sun:05:00-sun:06:00"`). |
| `applyImmediately` | `bool` | `false` | Apply changes immediately instead of waiting for the next maintenance window. May cause brief downtime. |
| `autoMinorVersionUpgrade` | `bool` | `false` | Automatically apply minor engine version upgrades during maintenance windows. Recommended: `true`. |
| `notificationTopicArn` | `StringValueOrRef` | — | SNS topic ARN for cluster event notifications. Can reference an AwsSnsTopic resource via `valueFrom`. |
| `preferredAvailabilityZones` | `string[]` | `[]` | Preferred AZs for cache nodes. When provided, list length must match `numCacheNodes`. |

## Examples

### Multi-Node Cross-AZ Cluster

A production cache spread across Availability Zones for resilience:

```yaml
apiVersion: aws.openmcf.org/v1
kind: AwsMemcachedElasticache
metadata:
  name: prod-cache
  labels:
    openmcf.org/provisioner: pulumi
    pulumi.openmcf.org/organization: my-org
    pulumi.openmcf.org/project: my-project
    pulumi.openmcf.org/stack.name: prod.AwsMemcachedElasticache.prod-cache
spec:
  region: us-east-1
  engineVersion: "1.6.22"
  nodeType: cache.r7g.large
  numCacheNodes: 3
  azMode: cross-az
  transitEncryptionEnabled: true
  autoMinorVersionUpgrade: true
  maintenanceWindow: "sun:05:00-sun:06:00"
  subnetIds:
    - subnet-private-az1
    - subnet-private-az2
    - subnet-private-az3
  securityGroupIds:
    - sg-memcached-prod
  preferredAvailabilityZones:
    - us-east-1a
    - us-east-1b
    - us-east-1c
```

### Custom Parameter Tuning

A cluster with custom Memcached engine parameters and SNS notifications:

```yaml
apiVersion: aws.openmcf.org/v1
kind: AwsMemcachedElasticache
metadata:
  name: tuned-cache
  labels:
    openmcf.org/provisioner: pulumi
    pulumi.openmcf.org/organization: my-org
    pulumi.openmcf.org/project: my-project
    pulumi.openmcf.org/stack.name: staging.AwsMemcachedElasticache.tuned-cache
spec:
  region: us-west-2
  engineVersion: "1.6.22"
  nodeType: cache.m7g.large
  numCacheNodes: 2
  azMode: cross-az
  applyImmediately: true
  parameterGroupFamily: memcached1.6
  parameters:
    - name: chunk_size
      value: "96"
    - name: chunk_size_growth_factor
      value: "1.5"
    - name: binding_protocol
      value: auto
  subnetIds:
    - subnet-private-az1
    - subnet-private-az2
  securityGroupIds:
    - sg-memcached-staging
  notificationTopicArn: arn:aws:sns:us-east-1:123456789012:cache-events
```

### Using Foreign Key References

Reference other OpenMCF-managed resources instead of hardcoding IDs:

```yaml
apiVersion: aws.openmcf.org/v1
kind: AwsMemcachedElasticache
metadata:
  name: ref-cache
  labels:
    openmcf.org/provisioner: pulumi
    pulumi.openmcf.org/organization: my-org
    pulumi.openmcf.org/project: my-project
    pulumi.openmcf.org/stack.name: prod.AwsMemcachedElasticache.ref-cache
spec:
  region: us-west-2
  engineVersion: "1.6.22"
  nodeType: cache.r7g.large
  numCacheNodes: 2
  azMode: cross-az
  transitEncryptionEnabled: true
  subnetIds:
    - valueFrom:
        kind: AwsSubnet
        name: my-private-subnet-a
        fieldPath: status.outputs.subnet_id
    - valueFrom:
        kind: AwsSubnet
        name: my-private-subnet-b
        fieldPath: status.outputs.subnet_id
  securityGroupIds:
    - valueFrom:
        kind: AwsSecurityGroup
        name: memcached-sg
        field: status.outputs.security_group_id
  notificationTopicArn:
    valueFrom:
      kind: AwsSnsTopic
      name: cache-events
      field: status.outputs.topic_arn
```

## Stack Outputs

After deployment, the following outputs are available in `status.outputs`:

| Output | Type | Description |
|--------|------|-------------|
| `cluster_id` | `string` | Identifier of the ElastiCache cluster |
| `cluster_address` | `string` | DNS name of the Memcached auto-discovery endpoint (without port). Empty for single-node clusters. |
| `configuration_endpoint` | `string` | Full configuration endpoint in `address:port` format. Recommended connection endpoint for multi-node clusters. |
| `arn` | `string` | Amazon Resource Name of the ElastiCache cluster |
| `port` | `int` | Port on which the cluster accepts connections |
| `subnet_group_name` | `string` | Name of the ElastiCache subnet group. Only populated when `subnetIds` were provided. |
| `parameter_group_name` | `string` | Name of the custom parameter group. Only populated when `parameters` were provided. |

## Related Components

- [AwsVpc](/docs/catalog/aws/vpc) — provides the subnets for cluster node placement
- [AwsSecurityGroup](/docs/catalog/aws/security-group) — controls network-level access to the Memcached endpoint
- [AwsSnsTopic](/docs/catalog/aws/sns-topic) — receives cluster event notifications
- [AwsRedisElasticache](/docs/catalog/aws/redis-elasticache) — alternative cache engine with replication, persistence, and authentication
