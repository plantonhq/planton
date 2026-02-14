# AzureRedisCache

Azure Cache for Redis is a fully managed, in-memory data store based on the open-source
Redis engine. It provides sub-millisecond data access for caching, session management,
real-time analytics, and message brokering workloads.

## When to Use

Use AzureRedisCache when you need:

- **Application caching** with sub-millisecond latency for frequently accessed data
- **Session storage** for web applications (sticky sessions across app instances)
- **Real-time analytics** with Redis data structures (sorted sets, HyperLogLog)
- **Message brokering** via Redis Pub/Sub or Streams
- **Rate limiting** and distributed locking

## SKU Tiers

| Tier | Nodes | SLA | Features |
|------|-------|-----|----------|
| Basic | 1 (single) | None | Dev/test only, no replication |
| Standard | 2 (primary + replica) | 99.9% | Production recommended, automatic failover |
| Premium | 2+ (primary + replica) | 99.9% | VNet injection, clustering, persistence, zones |

**Recommendation**: Use **Standard** for most production workloads. Use **Premium** only when
you need VNet isolation, Redis Cluster sharding (>53 GB), or data persistence.

## Cache Sizes (capacity)

### Basic / Standard (C-family)

| Capacity | Size | Max Connections | Bandwidth |
|----------|------|-----------------|-----------|
| 0 | 250 MB | 256 | 5 Mbps |
| 1 | 1 GB | 1,000 | 100 Mbps |
| 2 | 2.5 GB | 2,000 | 200 Mbps |
| 3 | 6 GB | 5,000 | 400 Mbps |
| 4 | 13 GB | 10,000 | 500 Mbps |
| 5 | 26 GB | 15,000 | 1 Gbps |
| 6 | 53 GB | 20,000 | 2 Gbps |

### Premium (P-family, per shard)

| Capacity | Size | Max Connections | Bandwidth |
|----------|------|-----------------|-----------|
| 1 | 6 GB | 7,500 | 2 Gbps |
| 2 | 13 GB | 15,000 | 3 Gbps |
| 3 | 26 GB | 30,000 | 4 Gbps |
| 4 | 53 GB | 40,000 | 5 Gbps |
| 5 | 120 GB | 40,000 | 5 Gbps |

## Eviction Policies (maxmemory_policy)

The eviction policy determines what happens when Redis reaches its memory limit.
Choosing the right policy is critical for cache behavior.

| Policy | Evicts | Strategy | Use Case |
|--------|--------|----------|----------|
| `volatile-lru` | Keys with TTL | Least recently used | **Default.** Mixed workloads (some keys persist, some expire) |
| `allkeys-lru` | Any key | Least recently used | Cache-only workloads (all keys expendable) |
| `volatile-lfu` | Keys with TTL | Least frequently used | Keys with varying access frequency |
| `allkeys-lfu` | Any key | Least frequently used | Cache-only with frequency-aware eviction |
| `volatile-random` | Keys with TTL | Random | Uniform access patterns |
| `allkeys-random` | Any key | Random | Uniform access patterns |
| `volatile-ttl` | Keys with TTL | Shortest TTL first | Prioritize long-lived keys |
| `noeviction` | None | Errors on write | Data must never be lost (monitor memory!) |

**Guidance**: If all keys are cache entries (no persistent data), use `allkeys-lru`. If
you mix cached and persistent data, the default `volatile-lru` is correct. If data loss
is unacceptable, use `noeviction` and monitor memory usage closely.

## Network Access

**Public access** (default): The cache is accessible over the internet via its hostname.
Use `firewall_rules` to restrict which IP addresses can connect. Without firewall rules,
all public IPs are blocked by default.

**VNet injection** (Premium only): Set `subnet_id` to deploy the cache inside a virtual
network subnet. The subnet must be dedicated to Azure Cache for Redis. The cache gets
a private IP address and is not reachable from the public internet.

**Private Endpoint**: For Standard/Premium caches without VNet injection, use
AzurePrivateEndpoint referencing the `redis_id` output for private connectivity.

## ForceNew Fields

Changing these fields destroys and recreates the cache (data loss):

- `name` -- cache hostname
- `subnet_id` -- VNet injection configuration

## Stack Outputs

| Output | Description |
|--------|-------------|
| `redis_id` | Azure Resource Manager ID (for Private Endpoint reference) |
| `hostname` | Cache hostname (`{name}.redis.cache.windows.net`) |
| `ssl_port` | SSL port (always 6380) |
| `primary_access_key` | Primary authentication key |
| `primary_connection_string` | Ready-to-use connection string |

## Related Resources

- **AzureSubnet** -- Dedicated subnet for VNet injection (Premium only)
- **AzurePrivateEndpoint** -- Private connectivity via Private Link
- **AzureResourceGroup** -- Container for the cache

## Infra Chart Usage

This resource appears as an optional component in:

- **database-stack** -- Optional caching layer alongside PostgreSQL/MSSQL
- **container-apps-environment** -- Optional session/cache store
- **web-app-environment** -- Optional session/cache store
