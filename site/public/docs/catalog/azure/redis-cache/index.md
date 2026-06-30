---
title: "Redis Cache"
description: "Redis Cache deployment documentation"
icon: "package"
order: 100
componentName: "azurerediscache"
---

# Azure Redis Cache

Deploys an Azure Cache for Redis instance with configurable SKU tier, capacity, eviction policy, optional VNet injection, Redis Cluster sharding, availability zones, patch schedules, and IP-based firewall rules. The component bundles the cache with its firewall rules as a single deployable unit.

## What Gets Created

When you deploy an AzureRedisCache resource, Planton provisions:

- **Redis Cache** -- a `redis.Cache` resource in the specified region and resource group, configured with the chosen SKU, capacity, Redis version, TLS settings, eviction policy, and optional clustering
- **VNet Injection** -- created only when `subnetId` is set (Premium SKU only), deploys the cache inside the specified subnet with private IP addressing
- **Firewall Rules** -- a `redis.FirewallRule` for each entry in `firewallRules`, allowing connections from specified IPv4 address ranges
- **Azure Tags** -- resource metadata tags applied to the cache for tracking and governance

## Prerequisites

- **Azure credentials** configured via environment variables or Planton provider config
- **An Azure Resource Group** where the cache will be created (can reference an AzureResourceGroup resource)
- **A globally unique cache name** -- the name becomes the endpoint `{name}.redis.cache.windows.net`
- **A dedicated subnet** if using VNet injection (Premium SKU only) -- the subnet must contain no other resources
- **SKU selection** -- choose Basic for dev/test, Standard for production, Premium for VNet injection, clustering, or zone redundancy

## Quick Start

Create a file `redis.yaml`:

```yaml
apiVersion: azure.planton.dev/v1
kind: AzureRedisCache
metadata:
  name: my-redis
  labels:
    planton.dev/provisioner: pulumi
    pulumi.planton.dev/organization: my-org
    pulumi.planton.dev/project: my-project
    pulumi.planton.dev/stack.name: dev.AzureRedisCache.my-redis
spec:
  region: eastus
  resourceGroup: my-rg
  name: my-redis
  capacity: 1
```

Deploy:

```shell
planton apply -f redis.yaml
```

This creates a Standard-tier Redis 6 cache with 1 GB capacity, SSL-only access, TLS 1.2, and `volatile-lru` eviction policy.

## Configuration Reference

### Required Fields

| Field | Type | Description | Validation |
|-------|------|-------------|------------|
| `region` | `string` | Azure region for the cache (e.g., `eastus`, `westeurope`). | Required, minimum length 1 |
| `resourceGroup` | `StringValueOrRef` | Azure Resource Group name. Can reference an AzureResourceGroup resource via `valueFrom`. | Required |
| `name` | `string` | Globally unique cache name. Becomes the endpoint `{name}.redis.cache.windows.net`. **ForceNew**: changing this destroys and recreates the cache. | Required, 1-63 characters, pattern `^[a-z][a-z0-9-]{0,62}$` |
| `capacity` | `int` | Cache size within the SKU tier. Basic/Standard: 0-6 (250 MB to 53 GB). Premium: 1-5 (6 GB to 120 GB per shard). | Required, 0-6 |

### Optional Fields

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `skuName` | `string` | `"Standard"` | SKU tier. Values: `Basic` (single node, no SLA), `Standard` (primary + replica, 99.9% SLA), `Premium` (VNet, clustering, zones). |
| `redisVersion` | `string` | `"6"` | Redis engine version. Values: `4`, `6`. |
| `subnetId` | `StringValueOrRef` | -- | Subnet ID for VNet injection (Premium only). Cache gets private IP addressing. Can reference an AzureSubnet resource via `valueFrom`. **ForceNew**. |
| `zones` | `string[]` | `[]` | Availability zones for the cache (e.g., `["1", "2", "3"]`). Requires Standard or Premium SKU. |
| `shardCount` | `int` | -- | Number of shards for Redis Cluster (Premium only). Total memory = capacity * (1 + shardCount). Range: 1-10. |
| `nonSslPortEnabled` | `bool` | `false` | Enable the non-SSL port (6379). Keep disabled for production. |
| `minimumTlsVersion` | `string` | `"1.2"` | Minimum TLS version. Values: `1.0`, `1.1`, `1.2`. |
| `publicNetworkAccessEnabled` | `bool` | `true` | Allow public internet access. Set to `false` for private-only access via VNet or Private Endpoint. |
| `maxmemoryPolicy` | `string` | `"volatile-lru"` | Eviction policy when cache is full. Values: `volatile-lru`, `allkeys-lru`, `volatile-lfu`, `allkeys-lfu`, `volatile-random`, `allkeys-random`, `volatile-ttl`, `noeviction`. |
| `patchSchedules` | `list` | `[]` | Maintenance windows. Each entry has `dayOfWeek` (required), optional `startHourUtc` (0-23), optional `maintenanceWindow` (default `PT5H`). |
| `firewallRules` | `list` | `[]` | IP-based access rules. Each entry has `name` (alphanumeric and underscores only), `startIp`, and `endIp`. Only effective with public access enabled. |

## Examples

### Development Cache

A Basic-tier cache for development and testing with minimal cost:

```yaml
apiVersion: azure.planton.dev/v1
kind: AzureRedisCache
metadata:
  name: dev-redis
  labels:
    planton.dev/provisioner: pulumi
    pulumi.planton.dev/organization: my-org
    pulumi.planton.dev/project: my-project
    pulumi.planton.dev/stack.name: dev.AzureRedisCache.dev-redis
spec:
  region: eastus
  resourceGroup: dev-rg
  name: dev-redis
  skuName: Basic
  capacity: 0
  firewallRules:
    - name: allow_dev_machine
      startIp: "203.0.113.42"
      endIp: "203.0.113.42"
```

### Production Cache with Firewall Rules

A Standard-tier cache with `allkeys-lru` eviction for a cache-only workload, firewall rules, and a scheduled maintenance window:

```yaml
apiVersion: azure.planton.dev/v1
kind: AzureRedisCache
metadata:
  name: prod-redis
  labels:
    planton.dev/provisioner: pulumi
    pulumi.planton.dev/organization: my-org
    pulumi.planton.dev/project: my-project
    pulumi.planton.dev/stack.name: prod.AzureRedisCache.prod-redis
spec:
  region: westeurope
  resourceGroup: prod-rg
  name: prod-redis
  capacity: 3
  maxmemoryPolicy: allkeys-lru
  zones:
    - "1"
    - "2"
  patchSchedules:
    - dayOfWeek: Saturday
      startHourUtc: 2
  firewallRules:
    - name: allow_office
      startIp: "203.0.113.0"
      endIp: "203.0.113.255"
    - name: allow_azure_services
      startIp: "0.0.0.0"
      endIp: "0.0.0.0"
```

### Premium Cache with VNet and Clustering

A Premium-tier cache deployed inside a VNet with Redis Cluster sharding, zone redundancy, and private-only access:

```yaml
apiVersion: azure.planton.dev/v1
kind: AzureRedisCache
metadata:
  name: enterprise-redis
  labels:
    planton.dev/provisioner: pulumi
    pulumi.planton.dev/organization: my-org
    pulumi.planton.dev/project: my-project
    pulumi.planton.dev/stack.name: prod.AzureRedisCache.enterprise-redis
spec:
  region: eastus
  resourceGroup: prod-rg
  name: enterprise-redis
  skuName: Premium
  capacity: 3
  shardCount: 3
  publicNetworkAccessEnabled: false
  subnetId: /subscriptions/00000000-0000-0000-0000-000000000000/resourceGroups/prod-rg/providers/Microsoft.Network/virtualNetworks/prod-vnet/subnets/redis
  zones:
    - "1"
    - "2"
    - "3"
  maxmemoryPolicy: noeviction
  patchSchedules:
    - dayOfWeek: Sunday
      startHourUtc: 3
      maintenanceWindow: PT3H
```

### Using Foreign Key References

Reference Planton-managed resources instead of hardcoding IDs:

```yaml
apiVersion: azure.planton.dev/v1
kind: AzureRedisCache
metadata:
  name: ref-redis
  labels:
    planton.dev/provisioner: pulumi
    pulumi.planton.dev/organization: my-org
    pulumi.planton.dev/project: my-project
    pulumi.planton.dev/stack.name: prod.AzureRedisCache.ref-redis
spec:
  region: eastus
  resourceGroup:
    valueFrom:
      kind: AzureResourceGroup
      name: my-rg
      field: status.outputs.resource_group_name
  name: ref-redis
  skuName: Premium
  capacity: 2
  subnetId:
    valueFrom:
      kind: AzureSubnet
      name: redis-subnet
      field: status.outputs.subnet_id
```

## Stack Outputs

After deployment, the following outputs are available in `status.outputs`:

| Output | Type | Description |
|--------|------|-------------|
| `redis_id` | `string` | Azure Resource Manager ID of the Redis cache. Referenced by AzurePrivateEndpoint for private connectivity. |
| `hostname` | `string` | Cache hostname (e.g., `{name}.redis.cache.windows.net`) |
| `ssl_port` | `int` | SSL port (always 6380) |
| `primary_access_key` | `string` | Primary access key for authentication (sensitive) |
| `primary_connection_string` | `string` | Ready-to-use connection string in the format `{hostname}:{port},password={key},ssl=True,abortConnect=False` (sensitive) |

## Related Components

- [AzureResourceGroup](/docs/catalog/azure/resource-group) -- provides the resource group for cache placement
- [AzureSubnet](/docs/catalog/azure/subnet) -- provides a dedicated subnet for VNet injection (Premium)
- [AzurePrivateEndpoint](/docs/catalog/azure/private-endpoint) -- establishes private connectivity to the cache
- [AzureVpc](/docs/catalog/azure/vpc-virtual-network) -- provides the virtual network containing the Redis subnet
