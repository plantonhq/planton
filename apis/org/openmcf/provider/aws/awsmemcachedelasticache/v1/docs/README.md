# AwsMemcachedElasticache — Architecture and Design

## Overview

AwsMemcachedElasticache provisions AWS ElastiCache clusters running the Memcached engine. This component wraps the Terraform `aws_elasticache_cluster` resource (with `engine = "memcached"`) and its Pulumi equivalent `elasticache.Cluster`.

## Why a Separate Component from Redis?

During the planning phase for the AWS resource expansion project, ElastiCache was initially planned as a single component (`AwsElasticacheCluster`). Deep research into the Terraform provider revealed that Memcached and Redis use fundamentally different AWS resources:

| Aspect | Memcached | Redis/Valkey |
|--------|-----------|-------------|
| **Terraform resource** | `aws_elasticache_cluster` | `aws_elasticache_replication_group` |
| **Topology** | Flat: N independent nodes | Hierarchical: primary + replicas, optional sharding |
| **Scaling** | Add/remove nodes (1–40) | Resize, add replicas, add shards |
| **Persistence** | None | Snapshots, AOF |
| **Auth** | None | AUTH token, Redis ACL |
| **Encryption at rest** | Not supported | Supported |

The field delta is ~15+ unique fields per engine. Combining them into one component would require extensive conditional logic, confusing CEL validations, and a degraded user experience. The split into `AwsRedisElasticache` and `AwsMemcachedElasticache` produces two focused, clean components.

## Memcached Topology

Memcached uses a simple distributed topology. Keys are partitioned across nodes using consistent hashing on the client side:

```
Client Application
       │
       ├── Node 1 (us-east-1a)  ─── keys A, D, G
       ├── Node 2 (us-east-1b)  ─── keys B, E, H
       └── Node 3 (us-east-1c)  ─── keys C, F, I
```

Each key lives on exactly one node. There is no replication between nodes. If a node fails, its keys are lost and must be re-fetched from the data source (database, API, etc.).

### Auto-Discovery

Multi-node Memcached clusters expose a **configuration endpoint** that supports auto-discovery. Clients that implement the ElastiCache auto-discovery protocol connect to this endpoint and automatically discover all nodes in the cluster. When nodes are added or removed, clients detect the topology change automatically.

The configuration endpoint is exposed in stack outputs as `configuration_endpoint` (address:port format) and `cluster_address` (DNS name only).

### AZ Distribution

- **single-az**: All nodes are placed in one Availability Zone. Lower latency between nodes, but vulnerable to AZ failure.
- **cross-az**: Nodes are distributed across multiple AZs. Provides resilience against AZ-level failures at the cost of slightly higher inter-node latency (irrelevant for Memcached since nodes don't communicate with each other).

## Scaling Behavior

### Horizontal Scaling (Node Count)

- **Adding nodes**: New nodes are created and join the cluster. Existing keys remain on their current nodes. New keys may hash to the new nodes. Some existing keys may need to be re-fetched as the hash ring expands.
- **Removing nodes**: Specified nodes are removed. Keys on removed nodes are lost and must be re-fetched.
- **Non-disruptive**: Horizontal scaling does not cause downtime for existing nodes.

### Vertical Scaling (Node Type)

Memcached **does not support in-place vertical scaling**. Changing the `nodeType` forces AWS to destroy the entire cluster and create a new one. All cached data is lost.

This is a critical operational difference from Redis, which supports in-place resizing.

## Security Model

Memcached has **no authentication mechanism**. There is no password, no token, no user/role system. Security relies entirely on:

1. **VPC network isolation** — Deploy in private subnets, not publicly accessible
2. **Security groups** — Restrict inbound traffic to port 11211 from trusted sources only
3. **Transit encryption** — TLS (engine 1.6.12+) protects data in transit from network sniffing

This makes `securityGroupIds` and `subnetIds` the most important security controls for Memcached deployments.

## Encryption

| Type | Support | Notes |
|------|---------|-------|
| At rest | **Not supported** | Memcached stores data exclusively in memory; there is no disk persistence to encrypt |
| In transit | Supported (1.6.12+) | TLS encryption for client-to-node connections |

## Parameter Groups

Memcached parameter groups use families like `memcached1.4`, `memcached1.5`, `memcached1.6`. Common tuning parameters:

| Parameter | Default | Description |
|-----------|---------|-------------|
| `chunk_size` | 48 | Minimum chunk size in bytes for slab allocation |
| `chunk_size_growth_factor` | 1.25 | Growth factor for slab chunk sizes |
| `max_simultaneous_connections` | 65000 | Maximum concurrent connections |
| `binding_protocol` | auto | Protocol: `auto`, `ascii`, `binary` |
| `large_memory_pages` | no | Use large memory pages (2MB) |

## Resource Lifecycle

```
CREATE                          UPDATE                          DELETE
  │                               │                               │
  ├─ Subnet Group (conditional)   ├─ Node count change            ├─ Cluster destroyed
  ├─ Parameter Group (conditional)│   (in-place scaling)          ├─ Subnet Group removed
  └─ Cluster                      ├─ Parameter changes            └─ Parameter Group removed
                                  ├─ Maintenance window
                                  └─ Node type change
                                      (FORCES RECREATION)
```

## IaC Module Design

Both Pulumi and Terraform modules follow the same structure:

1. **Region configuration** — The `region` field on the spec determines which AWS region the provider targets
2. **Conditional subnet group** — Only created when `subnetIds` are provided
3. **Conditional parameter group** — Only created when `parameters` and `parameterGroupFamily` are provided
4. **Cluster** — Always created, references the conditional resources above

The engine is hardcoded to `"memcached"` in the IaC modules. Unlike AwsRedisElasticache where the user selects between `redis` and `valkey`, the Memcached component always creates a Memcached cluster.

## Infra Chart Composability

AwsMemcachedElasticache is designed as a leaf-layer caching resource in infra charts:

- **Depends on**: AwsVpc (subnets), AwsSecurityGroup (access control), AwsSnsTopic (notifications)
- **Depended on by**: Application-layer resources that need cache endpoints
- **Key outputs for wiring**: `configuration_endpoint` (for application connection strings), `arn` (for IAM policies)
