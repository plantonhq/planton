# ScalewayRedisCluster

A managed Redis cluster on Scaleway with configurable cluster sizing, TLS, ACL-based or Private Network-based access control, and Redis engine tuning.

## Overview

`ScalewayRedisCluster` provisions a fully managed Redis cluster on Scaleway's Managed Redis service. This is a standalone resource (not composite) -- it wraps a single `scaleway_redis_cluster` Terraform resource with inline ACL and Private Network configuration.

Redis clusters are **zonal** resources, deployed to a specific availability zone (e.g., `fr-par-1`).

## Deployment Modes

The `clusterSize` field determines the deployment mode:

| Mode | cluster_size | Description |
|------|-------------|-------------|
| Standalone | 1 | Single node, no redundancy. Development and testing. |
| HA | 2 | 1 main + 1 standby with automatic failover. Production workloads. |
| Cluster | 3+ | Data sharded across nodes. High-throughput production workloads. |

## Networking (Mutually Exclusive)

Scaleway Redis supports two networking modes, but **they cannot be used simultaneously**:

| Mode | Field | Endpoint | Use Case |
|------|-------|----------|----------|
| Public + ACL | `aclRules` | Public IPs + port | External access with IP restrictions |
| Private Network | `privateNetworkId` | Private IPs + port | Internal-only access from same network |

This mutual exclusivity is enforced at the proto schema level via CEL validation and by the Scaleway API at apply time.

## Features

- **Three deployment modes** -- Standalone, HA, and Cluster (sharding)
- **TLS encryption** -- Optional encrypted client connections with certificate export
- **ACL rules** -- IP-based access control for public endpoint
- **Private Network** -- Attach to a `ScalewayPrivateNetwork` for secure internal connectivity
- **Redis settings** -- Engine-specific tuning (maxclients, eviction policy, keepalive, etc.)
- **Online scaling** -- Version and node type upgrades via online migration (no downtime)

## Quick Start

### Development Redis (Standalone)

```yaml
apiVersion: scaleway.openmcf.org/v1
kind: ScalewayRedisCluster
metadata:
  name: dev-cache
  env: development
spec:
  zone: fr-par-1
  version: "7.2.5"
  nodeType: RED1-MICRO
  clusterSize: 1
  userName: admin
  password: dev-cache-password-123
  aclRules:
    - ip: "0.0.0.0/0"
      description: "Allow all (dev only)"
```

### Production HA with Private Network

```yaml
apiVersion: scaleway.openmcf.org/v1
kind: ScalewayRedisCluster
metadata:
  name: prod-cache
  org: mycompany
  env: production
spec:
  zone: fr-par-1
  version: "7.2.5"
  nodeType: RED1-M
  clusterSize: 2
  tlsEnabled: true
  userName: cache_admin
  password: strong-random-password-here
  privateNetworkId:
    value: "xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx"
  settings:
    maxclients: "10000"
    tcp-keepalive: "120"
    maxmemory-policy: allkeys-lru
```

## Dependencies

| Dependency | Type | Required | Field |
|-----------|------|----------|-------|
| `ScalewayPrivateNetwork` | `StringValueOrRef` | No | `privateNetworkId` |

## Stack Outputs

| Output | Description | Downstream Use |
|--------|-------------|----------------|
| `cluster_id` | Zonal ID of the Redis cluster | Monitoring, management |
| `public_network_port` | Public endpoint port (0 if using PN) | Client connections |
| `public_network_ips` | Public endpoint IPs (empty if using PN) | Client connections |
| `private_network_port` | PN endpoint port (0 if not using PN) | Application connections |
| `private_network_ips` | PN endpoint IPs (empty if not using PN) | Application connections |
| `certificate` | TLS CA certificate PEM (empty if TLS off) | Secure client connections |

## Configuration Reference

### Required Fields

| Field | Type | Description |
|-------|------|-------------|
| `zone` | string | Availability zone (e.g., "fr-par-1") |
| `version` | string | Redis version (e.g., "7.2.5") |
| `nodeType` | string | Instance size (e.g., "RED1-MICRO") |
| `userName` | string | Initial user name |
| `password` | string | Initial user password (min 8 chars) |

### Optional Fields

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `clusterSize` | uint32 | 1 | Number of nodes (1=standalone, 2=HA, 3+=cluster) |
| `tlsEnabled` | bool | false | Enable TLS (forces recreation if changed) |
| `aclRules` | list | [] | Network ACL rules (conflicts with privateNetworkId) |
| `privateNetworkId` | StringValueOrRef | - | Private Network (conflicts with aclRules) |
| `settings` | map | {} | Redis engine settings |

### Node Types

| Type | Description | Use Case |
|------|-------------|----------|
| RED1-MICRO | Smallest available | Development, testing |
| RED1-S | Small | Light production |
| RED1-M | Medium | Standard production |
| RED1-L | Large | High-throughput production |
| RED1-XL | Extra large | Demanding production workloads |

## Lifecycle Warnings

### Destructive Changes

These changes **destroy and recreate** the cluster (data loss unless backed up externally):

- Changing `tlsEnabled`
- Changing from Standalone (1) to Cluster mode (3+)
- Reducing `clusterSize` in Cluster mode

### Safe Changes (Online Migration)

- Upgrading `version` (cannot downgrade)
- Upgrading `nodeType` (cannot downgrade)
- Increasing `clusterSize` within Cluster mode (3 -> 5, etc.)

### Immutable in Cluster Mode

- Private Network attachment cannot be changed after cluster creation in Cluster mode (3+ nodes)

## Best Practices

### Production Checklist

- Use HA mode (`clusterSize: 2`) or Cluster mode (`clusterSize: 3+`) for redundancy
- Enable TLS for encrypted connections (`tlsEnabled: true`)
- Use Private Network for application connections (no public exposure)
- Tune `maxmemory-policy` based on your caching strategy
- Set `tcp-keepalive` to detect dead connections
- Use strong, randomly generated passwords

### Security

- Prefer Private Network over ACL for production workloads
- Never use `0.0.0.0/0` ACL rules in production
- Enable TLS for compliance and defense-in-depth
- Rotate passwords regularly through your secrets management workflow

## Scaleway Documentation

- [Redis Overview](https://www.scaleway.com/en/docs/managed-databases/redis/)
- [Node Types & Pricing](https://www.scaleway.com/en/pricing/?tags=databases)
- [Private Network Integration](https://www.scaleway.com/en/docs/managed-databases/redis/how-to/connect-to-redis-cluster-private-network/)
- [Redis Settings](https://www.scaleway.com/en/docs/managed-databases/redis/reference-content/configuration/)
