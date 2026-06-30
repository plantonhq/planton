---
title: "Redis Cluster"
description: "Redis Cluster deployment documentation"
icon: "package"
order: 100
componentName: "scalewayrediscluster"
---

# Scaleway Redis Cluster

Deploys a Scaleway Managed Redis cluster with configurable cluster sizing (standalone, high availability, or sharded), optional TLS encryption, network ACL rules or Private Network attachment, and custom Redis engine settings. Redis clusters are zonal resources ideal for caching, session management, real-time analytics, and message brokering.

## What Gets Created

When you deploy a ScalewayRedisCluster resource, Planton provisions:

- **Redis Cluster** — a single `redis.Cluster` resource providing a fully managed Redis instance with the specified node type, engine version, cluster size, and initial user credentials
- **ACL Rules** — inline access control rules on the cluster's public endpoint, created only when `aclRules` is non-empty and `privateNetworkId` is not set
- **Private Network Attachment** — inline Private Network configuration on the cluster, created only when `privateNetworkId` is set and `aclRules` is empty

ACL rules and Private Network are mutually exclusive. Scaleway does not support both on the same cluster.

## Prerequisites

- **Scaleway credentials** configured via environment variables or Planton provider config
- **A valid Redis engine version** in semantic version format (e.g., `"7.2.5"`, `"6.2.7"`)
- **A Private Network** in the target zone if using private connectivity (can be created via a ScalewayPrivateNetwork resource)

## Quick Start

Create a file `redis-cluster.yaml`:

```yaml
apiVersion: scaleway.planton.dev/v1
kind: ScalewayRedisCluster
metadata:
  name: my-cache
  labels:
    planton.dev/provisioner: pulumi
    pulumi.planton.dev/organization: my-org
    pulumi.planton.dev/project: my-project
    pulumi.planton.dev/stack.name: dev.ScalewayRedisCluster.my-cache
spec:
  zone: fr-par-1
  version: "7.2.5"
  nodeType: RED1-MICRO
  userName: default
  password: change-me-strong-pw
```

Deploy:

```shell
planton apply -f redis-cluster.yaml
```

This creates a single-node Redis 7.2.5 cluster with a public endpoint accessible to all IPs (no ACL rules configured) and TLS disabled.

## Configuration Reference

### Required Fields

| Field | Type | Description | Validation |
|-------|------|-------------|------------|
| `zone` | `string` | Scaleway availability zone (e.g., `"fr-par-1"`, `"nl-ams-1"`, `"pl-waw-1"`). Redis is a zonal resource. Cannot be changed after creation. | Required |
| `version` | `string` | Redis engine version in semantic version format (e.g., `"7.2.5"`, `"6.2.7"`). Can be upgraded but never downgraded. | Required, pattern: `^[0-9]+\.[0-9]+\.[0-9]+$` |
| `nodeType` | `string` | Node type determining CPU, RAM, and performance (e.g., `"RED1-MICRO"`, `"RED1-S"`, `"RED1-M"`, `"RED1-L"`, `"RED1-XL"`). Can be upgraded but never downgraded. | Required |
| `userName` | `string` | Username for the cluster's initial (and only) user. | Required, max 63 characters |
| `password` | `string` | Password for the cluster user. | Required, min 8 characters |

### Optional Fields

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `clusterSize` | `uint32` | `1` | Number of nodes. `1` = standalone, `2` = high availability (automatic failover), `3+` = cluster mode (sharded). Scaling from standalone to cluster mode destroys and recreates the cluster. |
| `tlsEnabled` | `bool` | `false` | When `true`, all client connections must use TLS. The server certificate is exported in `status.outputs.certificate`. Changing this value destroys and recreates the cluster. |
| `aclRules` | `object[]` | `[]` | Network ACL rules for the public endpoint. Mutually exclusive with `privateNetworkId`. If empty and no Private Network is set, Scaleway allows all IPs. |
| `aclRules[].ip` | `string` | — | CIDR range to allow (e.g., `"10.0.0.0/24"`, `"1.2.3.4/32"`). Required per rule. |
| `aclRules[].description` | `string` | `""` | Human-readable label for the rule (e.g., `"Office IP"`, `"VPN egress"`). |
| `privateNetworkId` | `StringValueOrRef` | — | Private Network UUID for private connectivity. When set, no public endpoint is created. Mutually exclusive with `aclRules`. Can reference a ScalewayPrivateNetwork resource via `valueFrom`. In cluster mode (3+ nodes), cannot be changed after creation. |
| `settings` | `map<string, string>` | `{}` | Redis engine configuration key-value pairs (e.g., `"maxclients": "1000"`, `"maxmemory-policy": "allkeys-lru"`, `"tcp-keepalive": "120"`, `"timeout": "300"`). Applied on creation and updates. Available settings depend on the Redis version. |

## Examples

### Development Cache

A minimal standalone Redis cluster for development and testing:

```yaml
apiVersion: scaleway.planton.dev/v1
kind: ScalewayRedisCluster
metadata:
  name: dev-cache
  labels:
    planton.dev/provisioner: pulumi
    pulumi.planton.dev/organization: my-org
    pulumi.planton.dev/project: my-project
    pulumi.planton.dev/stack.name: dev.ScalewayRedisCluster.dev-cache
spec:
  zone: fr-par-1
  version: "7.2.5"
  nodeType: RED1-MICRO
  userName: appuser
  password: dev-redis-pw-2024
  settings:
    maxclients: "100"
    timeout: "600"
```

### Production HA with TLS and ACL

A high-availability Redis cluster with TLS encryption, network ACL rules, and tuned settings for a production session store:

```yaml
apiVersion: scaleway.planton.dev/v1
kind: ScalewayRedisCluster
metadata:
  name: prod-sessions
  labels:
    planton.dev/provisioner: pulumi
    pulumi.planton.dev/organization: my-org
    pulumi.planton.dev/project: my-project
    pulumi.planton.dev/stack.name: prod.ScalewayRedisCluster.prod-sessions
spec:
  zone: fr-par-1
  version: "7.2.5"
  nodeType: RED1-M
  clusterSize: 2
  userName: session_svc
  password: strong-prod-password-2024
  tlsEnabled: true
  aclRules:
    - ip: 10.0.0.0/16
      description: Internal VPC range
    - ip: 203.0.113.10/32
      description: VPN egress IP
  settings:
    maxclients: "5000"
    tcp-keepalive: "120"
    maxmemory-policy: allkeys-lru
    timeout: "300"
```

### Sharded Cluster on Private Network

A three-node sharded Redis cluster attached to an Planton-managed Private Network for high-throughput workloads:

```yaml
apiVersion: scaleway.planton.dev/v1
kind: ScalewayRedisCluster
metadata:
  name: analytics-cache
  labels:
    planton.dev/provisioner: pulumi
    pulumi.planton.dev/organization: my-org
    pulumi.planton.dev/project: my-project
    pulumi.planton.dev/stack.name: prod.ScalewayRedisCluster.analytics-cache
spec:
  zone: nl-ams-1
  version: "7.2.5"
  nodeType: RED1-L
  clusterSize: 3
  userName: analytics_svc
  password: analytics-cache-pw-2024
  tlsEnabled: true
  privateNetworkId:
    valueFrom:
      kind: ScalewayPrivateNetwork
      name: app-network
      fieldPath: status.outputs.private_network_id
  settings:
    maxclients: "10000"
    maxmemory-policy: allkeys-lfu
    tcp-keepalive: "60"
```

## Stack Outputs

After deployment, the following outputs are available in `status.outputs`:

| Output | Type | Description |
|--------|------|-------------|
| `cluster_id` | `string` | Zonal ID of the created Redis cluster (e.g., `"fr-par-1/xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx"`). Referenced by downstream resources. |
| `public_network_port` | `uint32` | Public endpoint TCP port. Zero when using Private Network mode. |
| `public_network_ips` | `string[]` | Public endpoint IPv4 addresses. Empty when using Private Network mode. |
| `private_network_port` | `uint32` | Private Network endpoint TCP port. Zero when using public network mode. |
| `private_network_ips` | `string[]` | Private Network endpoint IPv4 addresses. Empty when using public network mode. |
| `certificate` | `string` | TLS certificate in PEM format for verifying the Redis server. Empty when TLS is disabled. |

## Related Components

- [ScalewayPrivateNetwork](/docs/catalog/scaleway/private-network) — provides private connectivity between the Redis cluster and application workloads
- [ScalewayRdbInstance](/docs/catalog/scaleway/rdb-instance) — deploys managed PostgreSQL or MySQL databases that pair with Redis as a caching layer
- [ScalewayKapsuleCluster](/docs/catalog/scaleway/kapsule-cluster) — deploys Kubernetes clusters whose workloads connect to this Redis cluster
- [ScalewayInstanceSecurityGroup](/docs/catalog/scaleway/instance-security-group) — controls network access for compute instances connecting to the cluster
