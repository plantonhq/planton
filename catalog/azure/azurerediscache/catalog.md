# Azure Redis Cache

Deploys an Azure Cache for Redis instance -- a fully managed, in-memory data store built on the open-source Redis engine, used for caching, session state, real-time leaderboards, and pub/sub messaging with sub-millisecond latency. The kind models the full current surface: the tier/capacity ladder, engine configuration, the keyless (Entra) authentication posture, managed identity, RDB/AOF persistence, VNet injection, patch schedules, and firewall rules. Know before you choose it: Azure has announced the retirement of classic Azure Cache for Redis in favor of Azure Managed Redis, and ARM has begun rejecting new cache creations region by region -- existing caches keep running and this kind remains the right surface for managing them, but prefer Azure Managed Redis for new deployments.

## What Gets Created

When you deploy this Infra Component, the IaC module provisions:

- **Redis Cache** -- an Azure Cache for Redis instance in the specified region and resource group, configured with the chosen SKU tier, capacity, Redis version, engine configuration, and optional clustering, replicas, and zone pinning
- **VNet Injection** -- created only when `subnetId` is configured (Premium only); deploys the cache inside a dedicated subnet with private IP addressing, optionally pinned to a static address
- **RDB/AOF Persistence** -- created only when the persistence switches are enabled (Premium only); snapshots or write logs flow to a storage account via SAS or the cache's managed identity
- **Patch Schedules** -- created only when `patchSchedules` entries are configured; pins weekly maintenance windows for Redis engine updates
- **Firewall Rules** -- created only when `firewallRules` entries are configured; IPv4 allow-list for the public endpoint
- **Azure Tags** -- resource metadata tags (organization, environment, resource kind, resource ID) applied automatically, merged with any user tags

## Before You Deploy

### Planton Setup

- **Azure Provider Connection** -- an active connection in the Connect module with credentials for the target Azure subscription. Map it as the default for your environment, or specify it explicitly when creating the Infra Component.
- **Planton Runner** -- required when using Runner-based credential delivery. Not needed for inline credentials or browser OAuth authentication modes.

### Azure Subscription

- **An Azure Resource Group** where the Redis cache will be created. Provide the name directly or reference an AzureResourceGroup Infra Component via ValueFromRef.
- **A dedicated subnet** (optional, Premium only) for VNet injection. The subnet must contain nothing but Redis caches. Provide the subnet resource ID directly or reference an AzureSubnet Infra Component via ValueFromRef.

## Deploy

### Console

Open the deployment store, find **Azure Redis Cache**, and click **Deploy**. The creation wizard walks you through preset selection, environment and connection configuration, and spec fields. Start from the **Standard Cache with Entra Authentication** preset in the [Presets](#presets) tab to pre-populate a production-ready configuration.

### CLI

Create a manifest and apply it:

```yaml
apiVersion: azure.planton.dev/v1alpha1
kind: AzureRedisCache
metadata:
  name: app-cache
  org: acme-corp
  env: prod
spec:
  region: eastus
  resourceGroup:
    value: "acme-prod-rg"
  cacheName: acme-app-cache
  capacity: 2
```

```shell
planton apply -f redis-cache.yaml
```

This creates a Standard-tier (Azure's default when the tier is unspecified) Redis 6 cache with 2.5 GB memory (C2), TLS-only access on port 6380, and Azure's default engine behavior. No VNet injection, clustering, persistence, or firewall rules are configured. An Infra Job tracks the provisioning in real time.

### InfraChart

When deploying as part of a multi-resource environment, use ValueFromRef to wire the Redis cache to a resource group and subnet deployed in the same InfraPipeline:

```yaml
spec:
  resourceGroup:
    valueFrom:
      kind: AzureResourceGroup
      name: production-rg
      fieldPath: status.outputs.resource_group_name
  subnetId:
    valueFrom:
      kind: AzureSubnet
      name: redis-subnet
      fieldPath: status.outputs.subnet_id
```

The InfraPipeline resolves the dependency graph, deploys the resource group and subnet first, then provisions the Redis cache with the resolved values.

## Key Configuration

These are the most important decisions when configuring a Redis cache. Explore the full field reference in the [API Explorer](#api-explorer) tab.

**SKU tier and capacity** -- `skuName` unspecified deploys Standard: a replicated primary/replica pair with a 99.9% SLA, the right answer for most production caches. Basic is a single node with no SLA (dev/test only). Premium adds VNet injection, clustering, RDB/AOF persistence, extra replicas, and geo-replication via AzureRedisLinkedServer. `capacity` selects size within the tier's family: 0-6 for Basic/Standard (C0 250 MB to C6 53 GB), 1-5 for Premium (P1 6 GB to P5 120 GB, per shard when clustering). Tier upgrades apply in place; a downgrade replaces the cache.

**Keyless authentication** -- Two independent switches shape the auth posture. `redisConfiguration.activeDirectoryAuthenticationEnabled` turns on Microsoft Entra token authentication; `accessKeysAuthenticationEnabled` controls the shared keys and may only be turned off once Entra auth is on. Keys off + Entra on = fully keyless: identities connect under AzureRedisCacheAccessPolicyAssignment grants with tokens, and no secret exists at all.

**Eviction policy** -- `redisConfiguration.maxmemoryPolicy` controls behavior when the cache reaches its memory limit, in Redis's own vocabulary. Unset leaves Azure's default `volatile-lru` (evict TTL-bearing keys, least-recently-used first). Use `allkeys-lru` for cache-only workloads where every key is expendable, `noeviction` when data must never silently disappear (writes fail when full).

**Scale-out (Premium)** -- ONE choice, never both: `shardCount` (1-10) enables OSS cluster mode with total memory = capacity x shard count (clients must speak the Redis Cluster protocol); `replicasPerPrimary` (1-3) adds read replicas to one keyspace with no client changes.

**Persistence (Premium)** -- `redisConfiguration.rdbBackupEnabled` writes periodic snapshots to a storage account (the frequency is the recovery point); `aofBackupEnabled` logs every write near-synchronously. Storage auth is SAS connection strings (secrets) or `dataPersistenceAuthenticationMethod: MANAGED_IDENTITY` with the cache's `identity` block -- no storage secret in the spec.

**Network isolation** -- Public access is on by default; harden with `firewallRules` (IPv4 allow-list -- rule names take letters, digits, and underscores only), disable `publicNetworkAccessEnabled` and attach an AzurePrivateEndpoint (recommended for new designs), or (Premium) inject the cache into a dedicated subnet via `subnetId` -- the legacy isolation mechanism.

## Outputs and Dependencies

### What This Kind Consumes

| Dependency | Field | ValueFromRef Path |
|------------|-------|-------------------|
| **AzureResourceGroup** | `resourceGroup` | `status.outputs.resource_group_name` |
| **AzureSubnet** (optional) | `subnetId` | `status.outputs.subnet_id` |
| **AzureUserAssignedIdentity** (optional) | `identity.userAssignedIdentityIds` | `status.outputs.identity_id` |

### What This Kind Provides

After provisioning, `status.outputs` contains values that downstream Infra Components can consume via ValueFromRef:

| Output | Description | Common Downstream Use |
|--------|-------------|----------------------|
| `redis_cache_id` | Azure Resource Manager ID of the cache | AzureRedisCacheAccessPolicy, AzureRedisCacheAccessPolicyAssignment, AzureRedisLinkedServer, AzurePrivateEndpoint |
| `redis_cache_name` | The cache's name (the DNS label) | Sibling resources addressing the cache within its group |
| `region` | The cache's region | AzureRedisLinkedServer's `linkedRedisCacheLocation` |
| `resource_group_name` | The cache's resource group | Sibling resource placement |
| `hostname` | `{cacheName}.redis.cache.windows.net` | Keyless (Entra) clients need only this |
| `port` | The plaintext port (6379), only open when enabled | Legacy non-TLS clients |
| `ssl_port` | The TLS port (6380) | Every production client |
| `primary_access_key` | Primary key (SECRET; empty when keys are disabled) | Redis client password |
| `secondary_access_key` | Secondary key (SECRET) | Zero-downtime key rotation |
| `primary_connection_string` | Ready-to-use connection string (SECRET) | Application environment variables |
| `secondary_connection_string` | Rotation-window connection string (SECRET) | Zero-downtime key rotation |
| `identity_principal_id` | The system-assigned identity's principal ID | Role assignments (e.g. Storage Blob Data Contributor for persistence) |

## Common Patterns

Browse the [Presets](#presets) tab for ready-to-deploy configurations.

**Standard production cache with Entra auth** -- A Standard-tier replicated cache with Entra token authentication enabled alongside the keys, ready to migrate clients toward the keyless posture. Start from the **Standard Cache with Entra Authentication** preset.

**Premium enterprise cache** -- A Premium-tier cache with zone pinning, clustering or replicas, and persistence for workloads where a cold rebuild is expensive. Start from the **Premium Enterprise Cache** preset.

**Development cache** -- A Basic-tier single-node cache for development and testing. No SLA, no replication, lowest cost. Start from the **Development Cache** preset.

## Works With

- [**Azure Resource Group**](/infra-catalog/azure-resource-group) -- provides the resource group where the Redis cache is created
- [**Azure Subnet**](/infra-catalog/azure-subnet) -- provides the VNet subnet for Premium-tier VNet injection
- [**Azure User Assigned Identity**](/infra-catalog/azure-user-assigned-identity) -- the identity for keyless persistence-storage access
- [**Azure Redis Cache Access Policy**](/infra-catalog/azure-redis-cache-access-policy) -- custom data-plane permission sets in Redis ACL syntax
- [**Azure Redis Cache Access Policy Assignment**](/infra-catalog/azure-redis-cache-access-policy-assignment) -- grants a policy (built-in or custom) to a Microsoft Entra identity
- [**Azure Redis Linked Server**](/infra-catalog/azure-redis-linked-server) -- geo-replication link pairing two Premium caches for disaster recovery
- [**Azure Private Endpoint**](/infra-catalog/azure-private-endpoint) -- private connectivity to the cache with public access disabled
- [**Azure Managed Redis**](/infra-catalog/azure-managed-redis) -- the successor service; prefer it for new Redis deployments as classic-cache creation retires region by region
