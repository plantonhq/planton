# AzureRedisCache: Deployment Landscape & Design Research

## What is Azure Cache for Redis?

Azure Cache for Redis is Microsoft's fully managed implementation of the open-source Redis
in-memory data store. It provides sub-millisecond data access for caching, session management,
real-time analytics, message brokering, and distributed locking workloads. Azure manages
patching, monitoring, scaling, and high availability.

## Deployment Methods Compared

| Method | Scope | Abstraction | Networking | Best For |
|--------|-------|-------------|------------|----------|
| Azure Portal | Single cache | GUI | Manual | Learning, one-off |
| Azure CLI / PowerShell | Single cache | Imperative | Manual | Quick scripting |
| ARM/Bicep Templates | Full stack | Declarative | In-template | Azure-native IaC |
| Terraform (`azurerm_redis_cache`) | Full stack | Declarative | In-module | Multi-cloud IaC |
| Pulumi (`redis.Cache`) | Full stack | Declarative | Programmatic | Multi-cloud, type-safe IaC |
| **Planton AzureRedisCache** | Full stack | Declarative | Composable | Multi-cloud consistency, infra charts |

### Why Planton for Redis?

Planton's value is **composability**. A Redis cache rarely exists in isolation -- it's
part of a larger deployment alongside databases, compute, and networking. Planton's
`StringValueOrRef` mechanism lets infra charts wire the Redis cache into dependency-aware
deployment DAGs alongside AzureResourceGroup, AzureSubnet, AzurePostgresqlFlexibleServer,
and other resources.

## Azure Cache for Redis Architecture

### Tier Comparison (Deep)

**Basic**: Single Redis node, no replication, no SLA. Azure allocates a VM running
the Redis process. If the VM fails, the cache is rebuilt from scratch (data loss).
The only use case is development, testing, and CI/CD pipelines.

**Standard**: Two-node deployment (primary + replica). Azure manages automatic
failover -- if the primary node fails, the replica promotes automatically. This
provides a 99.9% SLA. The replica is not directly accessible to clients; it exists
solely for failover. This is the right tier for most production workloads.

**Premium**: Same two-node base as Standard, plus advanced features:
- **VNet injection**: Deploy into a dedicated subnet for network isolation
- **Redis Cluster**: Shard data across multiple primary/replica pairs (1-10 shards)
- **Data persistence**: RDB snapshots or AOF logs to Azure Blob Storage
- **Zone redundancy**: Distribute nodes across availability zones
- **Larger sizes**: P1-P5 (6 GB to 120 GB per shard)

### Memory Management

Redis is an in-memory data store. When the cache reaches its memory limit, the
`maxmemory_policy` determines behavior:

1. **Eviction-based policies** (`volatile-lru`, `allkeys-lru`, etc.): Redis removes
   keys to make room for new writes. This is the default behavior and appropriate
   for caching workloads.

2. **No-eviction** (`noeviction`): Redis returns errors on writes when full. This
   protects against data loss but requires careful memory monitoring.

The `maxmemory_reserved` and `maxmemory_delta` settings (omitted in v1) control how
much memory Redis reserves for non-data overhead (replication buffers, fragmentation).
These are advanced tuning parameters for Premium tier.

### Networking

**Public access**: Default mode. The cache gets a public hostname
(`{name}.redis.cache.windows.net`) and is accessible from any network.
Firewall rules restrict which IPs can connect.

**VNet injection** (Premium only): The cache is deployed into a dedicated subnet
and gets a private IP. No public endpoint. This is the strongest isolation model
but requires Premium tier pricing.

**Private Endpoint**: Any tier (Standard/Premium) can use Azure Private Link via
the AzurePrivateEndpoint resource. This creates a private IP in your VNet that
maps to the cache's public endpoint. A middle-ground between public access and
full VNet injection.

## 80/20 Scoping Rationale

### What's Included (Covers 80%+ of Production Use Cases)

| Feature | Field | Rationale |
|---------|-------|-----------|
| SKU tier selection | `sku_name` | Every deployment needs to choose a tier |
| Cache sizing | `capacity` | Every deployment needs to choose a size |
| Eviction policy | `maxmemory_policy` | Critical for cache behavior in production |
| Network access | `public_network_access_enabled` | Private vs public is a baseline decision |
| VNet injection | `subnet_id` | Enterprise networking requirement |
| Zone redundancy | `zones` | Production HA requirement |
| Redis Cluster | `shard_count` | Large-scale caching requirement |
| Maintenance windows | `patch_schedules` | Production operational requirement |
| Firewall rules | `firewall_rules` | Network security baseline |
| TLS enforcement | `minimum_tls_version` | Security compliance |
| Redis version | `redis_version` | Engine version control |

### What's Deferred to v2 (Niche or Advanced)

| Feature | Rationale for Deferral |
|---------|----------------------|
| Data persistence (RDB/AOF) | Premium-only, requires storage account wiring |
| Managed identity | Advanced Key Vault / AAD auth integration |
| Memory tuning (maxmemory_reserved, delta) | Advanced DBA territory, auto-tuned by Azure |
| Replicas per primary | Premium replication fine-tuning |
| AAD-only authentication | New feature, enterprise IAM integration |
| Linked servers (geo-replication) | Complex multi-region setup |
| Access policies | New RBAC model for Redis |
| Tenant settings | Custom settings map |

## Common Deployment Patterns

### Pattern 1: Application Cache (Standard)

The most common pattern. A Standard C2 or C3 cache sitting between the application
and database, caching query results, API responses, and computed values.

**Key decisions:**
- `sku_name: Standard` (SLA, replication)
- `maxmemory_policy: allkeys-lru` (all keys are cache entries)
- `capacity: 2-3` (2.5 GB to 6 GB)

### Pattern 2: Session Store (Standard)

Web application session storage across multiple app instances. Sessions have TTL,
so `volatile-lru` (default) is appropriate.

**Key decisions:**
- `sku_name: Standard`
- `maxmemory_policy: volatile-lru` (default, sessions have TTL)
- `capacity: 1-2` (1 GB to 2.5 GB)

### Pattern 3: Enterprise Cache (Premium + VNet)

Network-isolated cache for enterprise workloads requiring VNet integration and
zone redundancy.

**Key decisions:**
- `sku_name: Premium`
- `subnet_id` pointing to a dedicated subnet
- `zones: ["1", "2"]` for zone redundancy
- `public_network_access_enabled: false`

### Pattern 4: High-Throughput Cache (Premium + Cluster)

Large-scale caching with Redis Cluster sharding for higher throughput and larger
data sets than a single node can handle.

**Key decisions:**
- `sku_name: Premium`
- `shard_count: 2-5` depending on data size
- `capacity: 2-3` per shard

## Connection String Format

Azure Cache for Redis connection strings follow this format:

```
{hostname}:{ssl_port},password={primary_access_key},ssl=True,abortConnect=False
```

Example:
```
myapp-redis.redis.cache.windows.net:6380,password=abc123...,ssl=True,abortConnect=False
```

Most Redis client libraries (.NET StackExchange.Redis, Node.js ioredis, Python redis-py)
accept this format directly.

## Infra Chart Composition

### Upstream Dependencies

- **AzureResourceGroup** (`resource_group`): Container for the cache
- **AzureSubnet** (`subnet_id`): VNet injection target (Premium only)

### Downstream Consumers

AzureRedisCache is typically a leaf resource -- applications consume it via
connection string, not via other Planton resources. The exception is
AzurePrivateEndpoint, which references `redis_id` for private connectivity.

### DAG Position

```
AzureResourceGroup (Layer 0)
  └── AzureRedisCache (Layer 1-2, alongside databases)
        └── AzurePrivateEndpoint (optional, Layer 2-3)
```

## Further Reading

- [Azure Cache for Redis Documentation](https://learn.microsoft.com/en-us/azure/azure-cache-for-redis/)
- [Redis Eviction Policies](https://redis.io/docs/reference/eviction/)
- [Azure Cache for Redis Best Practices](https://learn.microsoft.com/en-us/azure/azure-cache-for-redis/cache-best-practices)
- [Terraform azurerm_redis_cache](https://registry.terraform.io/providers/hashicorp/azurerm/latest/docs/resources/redis_cache)
- [Pulumi Azure Redis Cache](https://www.pulumi.com/registry/packages/azure/api-docs/redis/cache/)
