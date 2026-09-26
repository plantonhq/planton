# AzureRedisCache

> Generated from the protobuf schema by `make generate-reference` -- do not
> edit by hand. To change a fact on this page, change the proto field comment
> or validation rule it is derived from, then regenerate.

**apiVersion**: `azure.planton.dev/v1alpha1`

**AzureRedisCacheSpec** defines the configuration for creating an Azure
Cache for Redis instance -- a fully managed, in-memory data store built on
the open-source Redis engine, used for caching, session state, real-time
leaderboards, and pub/sub messaging with sub-millisecond latency.

**Tiers.** The `sku_name` tier decides the capability envelope:
- BASIC: a single node with no replica and no SLA -- dev/test only.
- STANDARD (the default): a replicated primary/replica pair with a 99.9%
  SLA -- the right answer for most production caches.
- PREMIUM: everything in Standard plus VNet injection, clustering
  (sharding), data persistence (RDB/AOF), extra replicas, and
  geo-replication via AzureRedisLinkedServer.
The tier can be upgraded in place; a DOWNGRADE replaces the cache.

**Retirement notice (live-verified).** Azure has announced the retirement
of classic Azure Cache for Redis in favor of Azure Managed Redis, and ARM
has begun rejecting NEW cache creations region by region ("Azure Cache
for Redis is retiring, create Azure Managed Redis instance instead" --
observed on new PREMIUM creations in some regions while Basic/Standard
creations elsewhere still succeed). Existing caches keep running and this
kind remains the right surface for managing them; prefer
AzureManagedRedis for NEW deployments.

**Sizing.** `capacity` picks the size within the tier's family:
C0-C6 (250 MB to 53 GB) for Basic/Standard, P1-P5 (6 GB to 120 GB per
shard) for Premium. The family letter ("C" or "P") is derived from the
tier by the IaC modules -- it is never spelled twice.

**Authentication.** Two independent switches shape the auth posture:
`redis_configuration.active_directory_authentication_enabled` turns on
Microsoft Entra token authentication, and
`access_keys_authentication_enabled` controls the shared access keys.
Keys can only be turned OFF once Entra auth is ON (enforced here, exactly
as ARM enforces it) -- the fully keyless posture pairs this cache with
AzureRedisCacheAccessPolicyAssignment grants instead of secrets.

**Networking.** Public access is on by default; harden with
`firewall_rules` (IP allow-list), disable `public_network_access_enabled`
and reach the cache through an AzurePrivateEndpoint, or (Premium) inject
the cache into a dedicated subnet via `subnet_id` for private-IP-only
deployment. VNet injection is the legacy isolation mechanism -- Azure
recommends Private Link for new deployments; both are modeled.

**ForceNew fields** (changing these replaces the cache): `cache_name`,
`subnet_id`, `private_static_ip_address`, `zones`, and any `sku_name`
downgrade.

## Example

```yaml
apiVersion: azure.planton.dev/v1alpha1
kind: AzureRedisCache
metadata:
  name: test-redis
spec:
  region: eastus
  resourceGroup:
    value: test-rg
  cacheName: planton-hack-redis
  # Exercises the sku enum mapping and the P-family derivation.
  skuName: PREMIUM
  capacity: 1
  redisVersion: "6"
  zones:
    - "1"
  # Exercises VNet injection with a pinned private address.
  subnetId:
    value: /subscriptions/00000000-0000-0000-0000-000000000000/resourceGroups/test-rg/providers/Microsoft.Network/virtualNetworks/test-vnet/subnets/redis
  privateStaticIpAddress: 10.0.1.10
  # Exercises the Premium replica dial (mutually exclusive with shards).
  replicasPerPrimary: 2
  # Exercises the keyless posture: keys off is only legal with Entra on.
  accessKeysAuthenticationEnabled: false
  redisConfiguration:
    activeDirectoryAuthenticationEnabled: true
    # Exercises the eviction vocabulary and every memory dial.
    maxmemoryPolicy: allkeys-lru
    maxmemoryReserved: 100
    maxmemoryDelta: 100
    maxfragmentationmemoryReserved: 100
    notifyKeyspaceEvents: KEA
    # false is only legal inside a VNet-injected cache.
    authenticationEnabled: false
    # Exercises the persistence-auth enum mapping and the RDB dials --
    # the managed-identity path needs no storage connection string.
    dataPersistenceAuthenticationMethod: MANAGED_IDENTITY
    rdbBackupEnabled: true
    rdbBackupFrequency: 60
    rdbBackupMaxSnapshotCount: 1
  # Exercises the identity-type enum mapping (pairs with the
  # managed-identity persistence auth above).
  identity:
    type: SYSTEM_ASSIGNED
  # Exercises the day-of-week enum mapping and the window dials.
  patchSchedules:
    - dayOfWeek: SUNDAY
      startHourUtc: 2
      maintenanceWindow: PT6H
  firewallRules:
    - name: office_range
      startIp: 203.0.113.0
      endIp: 203.0.113.255
  tenantSettings:
    trace: "true"
  tags:
    cost-center: platform
```

## Spec Fields

| Path | Type | Required | Default | References |
|---|---|---|---|---|
| `spec.region` | `string` | yes |  |  |
| `spec.resourceGroup` | `string \| valueFrom` | yes |  | AzureResourceGroup (`status.outputs.resource_group_name`) |
| `spec.cacheName` | `string` | yes |  |  |
| `spec.skuName` | `enum` |  |  |  |
| `spec.capacity` | `int32` |  |  |  |
| `spec.redisVersion` | `string` |  | `6` |  |
| `spec.zones` | `[]string` |  |  |  |
| `spec.subnetId` | `string \| valueFrom` |  |  | AzureSubnet (`status.outputs.subnet_id`) |
| `spec.privateStaticIpAddress` | `string` |  |  |  |
| `spec.shardCount` | `int32` |  |  |  |
| `spec.replicasPerPrimary` | `int32` |  |  |  |
| `spec.nonSslPortEnabled` | `bool` |  |  |  |
| `spec.publicNetworkAccessEnabled` | `bool` |  | `true` |  |
| `spec.accessKeysAuthenticationEnabled` | `bool` |  | `true` |  |
| `spec.redisConfiguration` | `AzureRedisCacheConfiguration` |  |  |  |
| `spec.redisConfiguration.activeDirectoryAuthenticationEnabled` | `bool` |  |  |  |
| `spec.redisConfiguration.maxmemoryPolicy` | `string` |  | `volatile-lru` |  |
| `spec.redisConfiguration.maxmemoryReserved` | `int32` |  |  |  |
| `spec.redisConfiguration.maxmemoryDelta` | `int32` |  |  |  |
| `spec.redisConfiguration.maxfragmentationmemoryReserved` | `int32` |  |  |  |
| `spec.redisConfiguration.notifyKeyspaceEvents` | `string` |  |  |  |
| `spec.redisConfiguration.authenticationEnabled` | `bool` |  | `true` |  |
| `spec.redisConfiguration.dataPersistenceAuthenticationMethod` | `enum` |  |  |  |
| `spec.redisConfiguration.rdbBackupEnabled` | `bool` |  |  |  |
| `spec.redisConfiguration.rdbBackupFrequency` | `int32` |  |  |  |
| `spec.redisConfiguration.rdbBackupMaxSnapshotCount` | `int32` |  |  |  |
| `spec.redisConfiguration.rdbStorageConnectionString` | `string` (sensitive) |  |  |  |
| `spec.redisConfiguration.aofBackupEnabled` | `bool` |  |  |  |
| `spec.redisConfiguration.aofStorageConnectionString0` | `string` (sensitive) |  |  |  |
| `spec.redisConfiguration.aofStorageConnectionString1` | `string` (sensitive) |  |  |  |
| `spec.redisConfiguration.storageAccountSubscriptionId` | `string` |  |  |  |
| `spec.identity` | `AzureRedisCacheIdentity` |  |  |  |
| `spec.identity.type` | `enum` |  |  |  |
| `spec.identity.userAssignedIdentityIds` | `[]string \| valueFrom` |  |  | AzureUserAssignedIdentity (`status.outputs.identity_id`) |
| `spec.patchSchedules` | `[]AzureRedisCachePatchSchedule` |  |  |  |
| `spec.patchSchedules[].dayOfWeek` | `enum` |  |  |  |
| `spec.patchSchedules[].startHourUtc` | `int32` |  |  |  |
| `spec.patchSchedules[].maintenanceWindow` | `string` |  | `PT5H` |  |
| `spec.firewallRules` | `[]AzureRedisCacheFirewallRule` |  |  |  |
| `spec.firewallRules[].name` | `string` | yes |  |  |
| `spec.firewallRules[].startIp` | `string` | yes |  |  |
| `spec.firewallRules[].endIp` | `string` | yes |  |  |
| `spec.tenantSettings` | `map<string, string>` |  |  |  |
| `spec.tags` | `map<string, string>` |  |  |  |

## Field Details

### spec.region

`string` · required

The Azure region where the Redis cache will be created.
Examples: "eastus", "westus2", "westeurope", "southeastasia".

- rule: {"required":true,"string":{"minLen":"1"}}

### spec.resourceGroup

`string | valueFrom` · required

The Azure Resource Group where the Redis cache will be created.
Can be a literal string or a reference to an AzureResourceGroup output.

- references: AzureResourceGroup (`status.outputs.resource_group_name`)
- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: AzureResourceGroup, name: <that resource's name>, fieldPath: status.outputs.resource_group_name}} -- a bare string does not parse

### spec.cacheName

`string` · required

The cache's name -- GLOBALLY unique across all of Azure, because it
becomes the public DNS endpoint `{cache_name}.redis.cache.windows.net`.
1-63 letters, digits, and hyphens; must start and end with a letter or
digit; no consecutive hyphens. Deletion runs several minutes but frees
the name as soon as it completes -- there is no soft-delete hold.
Changing the name replaces the cache.

- rule: cache_name must be 1-63 letters, digits, and hyphens, start and end with a letter or digit, and never repeat a hyphen
- rule: {"required":true,"string":{"maxLen":"63"}}

### spec.skuName

`enum`

The pricing/capability tier. Unspecified deploys STANDARD -- the
replicated 99.9%-SLA tier that fits most production workloads. BASIC is
a single unreplicated node for dev/test. PREMIUM unlocks VNet injection,
clustering, persistence, extra replicas, and geo-replication. Upgrades
apply in place; a downgrade REPLACES the cache (Azure does not support
SKU downgrades on a live instance).

- rule: {"enum":{"definedOnly":true}}

Allowed values (use exactly as shown):

- `azure_redis_cache_sku_unspecified` -- Not specified -- deploys STANDARD, the production default.
- `BASIC` -- A single cache node with no replica and NO SLA -- development and testing only. Memory dials (maxmemory_reserved and friends) are not configurable on this tier.
- `STANDARD` -- A replicated primary/replica pair with a 99.9% SLA -- the right answer for most production caches.
- `PREMIUM` -- Everything in STANDARD plus VNet injection, clustering (sharding), RDB/AOF data persistence, extra replicas, and geo-replication via AzureRedisLinkedServer.

### spec.capacity

`int32`

Cache size within the tier's family. Basic/Standard (C-family):
0=250MB, 1=1GB, 2=2.5GB, 3=6GB, 4=13GB, 5=26GB, 6=53GB. Premium
(P-family): 1=6GB, 2=13GB, 3=26GB, 4=53GB, 5=120GB -- per shard when
clustering, so total memory = size x (shard_count + 1) primaries plus
replicas. Unset on Basic/Standard means C0, the smallest size. The
family letter is derived from the tier by the modules.

- rule: {"int32":{"lte":6,"gte":0}}

### spec.redisVersion

`string` · optional (explicit presence)

Redis engine major version: "4" or "6". Azure tracks the latest 6.x
patch release automatically -- only the major version is chosen here.
Default "6"; Redis 4 is end-of-life and exists only for legacy clients.
Upgrading 4 -> 6 applies in place.

- default: `6`
- rule: redis_version must be one of: 4, 6

### spec.zones

`[]string`

Availability zones to pin the cache's nodes to, e.g. ["1", "2"].
Zone-redundant deployment spreads the primary and replicas across
zones for datacenter-failure resilience. Fixed at creation.

- rule: {"repeated":{"items":{"string":{"in":["1","2","3"]}}}}

### spec.subnetId

`string | valueFrom`

PREMIUM ONLY -- the dedicated subnet to inject the cache into for
private-IP-only deployment (the subnet must contain nothing but Redis
caches). VNet injection is the legacy isolation mechanism; prefer
disabling public network access and attaching an AzurePrivateEndpoint
for new designs. Fixed at creation.
Can be a literal ARM resource ID or a reference to an AzureSubnet output.

- references: AzureSubnet (`status.outputs.subnet_id`)
- rule: write as {value: <literal>} or {valueFrom: {kind: AzureSubnet, name: <that resource's name>, fieldPath: status.outputs.subnet_id}} -- a bare string does not parse

### spec.privateStaticIpAddress

`string` · optional (explicit presence)

The static private IP to assign the cache inside its injected subnet.
Only meaningful with subnet_id; Azure picks an address when unset.
Fixed at creation.

- rule: {"string":{"ipv4":true}}

### spec.shardCount

`int32` · optional (explicit presence)

PREMIUM ONLY -- the number of shards in a clustered (OSS cluster mode)
cache, 1-10. Each shard is its own primary/replica pair, so clustering
multiplies both throughput and memory: total memory = capacity x
shard_count. Clients must speak the Redis Cluster protocol. Cannot be
combined with extra replicas_per_primary.

- rule: {"int32":{"lte":10,"gte":1}}

### spec.replicasPerPrimary

`int32` · optional (explicit presence)

PREMIUM ONLY -- extra read replicas per primary, 1-3 (the tier's
built-in replica counts as one). More replicas raise read throughput
and shrink failover windows; they cannot be combined with clustering.
ARM exposes the same setting under a legacy alias (replicasPerMaster)
that mirrors this value -- only the modern name is modeled.

- rule: {"int32":{"lte":3,"gte":1}}

### spec.nonSslPortEnabled

`bool`

Enable the plaintext non-SSL port (6379). Off by default and best left
off: the non-SSL port sends commands AND access keys unencrypted.
Exists for legacy clients that cannot speak TLS.

### spec.publicNetworkAccessEnabled

`bool` · optional (explicit presence)

Whether the cache answers on its public endpoint. Default true. Set
false to force all traffic through an AzurePrivateEndpoint (the
recommended isolation for new designs) -- firewall_rules become
irrelevant once public access is off.

- default: `true`

### spec.accessKeysAuthenticationEnabled

`bool` · optional (explicit presence)

Whether the shared access keys authenticate clients. Default true.
Turning keys OFF -- allowed only once Entra authentication is ON
(redis_configuration.active_directory_authentication_enabled) -- makes
the cache fully keyless: clients present Entra tokens under
AzureRedisCacheAccessPolicyAssignment grants, and there is no secret
to leak or rotate.

- default: `true`

### spec.redisConfiguration

`AzureRedisCacheConfiguration`

Redis engine and platform behavior: eviction policy, memory
reservations, Entra auth, keyspace notifications, and (Premium) RDB/AOF
persistence. Every field has a safe Azure default -- omit the whole
block for a standard cache.

### spec.redisConfiguration.activeDirectoryAuthenticationEnabled

`bool`

Enable Microsoft Entra (Azure AD) token authentication. Off by
default. Turning this on lets identities authenticate with Entra
tokens under access policies (see AzureRedisCacheAccessPolicyAssignment)
instead of shared keys -- and is the prerequisite for disabling the
access keys entirely (spec.access_keys_authentication_enabled).

### spec.redisConfiguration.maxmemoryPolicy

`string` · optional (explicit presence)

Eviction policy when the cache reaches its memory limit, in Redis's
own configuration vocabulary. Default volatile-lru (evict TTL-bearing
keys, least-recently-used first). Guidance: allkeys-lru for pure
caches where every key is expendable; volatile-lru for mixed
workloads; noeviction when data must never silently disappear (writes
fail when full -- watch memory).

- default: `volatile-lru`
- rule: maxmemory_policy must be one of: allkeys-lfu, allkeys-lru, allkeys-random, noeviction, volatile-lfu, volatile-lru, volatile-random, volatile-ttl

### spec.redisConfiguration.maxmemoryReserved

`int32` · optional (explicit presence)

Megabytes reserved for non-cache operations (failover, replication).
Azure sizes a default from the cache's total memory; raise it for
write-heavy workloads to keep failovers smooth. Not configurable on
the BASIC tier.

- rule: {"int32":{"gte":0}}

### spec.redisConfiguration.maxmemoryDelta

`int32` · optional (explicit presence)

Megabytes reserved per shard for the memory overhead of writes during
high-load periods. Azure sizes a default from total memory. Not
configurable on the BASIC tier.

- rule: {"int32":{"gte":0}}

### spec.redisConfiguration.maxfragmentationmemoryReserved

`int32` · optional (explicit presence)

Megabytes reserved to accommodate memory fragmentation. Azure sizes a
default from total memory; raise it for workloads with many small
keys. Not configurable on the BASIC tier.

- rule: {"int32":{"gte":0}}

### spec.redisConfiguration.notifyKeyspaceEvents

`string`

Keyspace event notification classes, in Redis's compact flag syntax
(e.g. "KEA" for everything, "Ex" for expiry events). Empty (default)
disables notifications. Clients subscribe over pub/sub to react to
key changes -- cache-invalidation fan-out, session-expiry hooks.

### spec.redisConfiguration.authenticationEnabled

`bool` · optional (explicit presence)

Whether Redis requires clients to authenticate at all. Default true;
can only be disabled on a VNet-injected cache (subnet_id set), where
the subnet boundary is the access control. Never disable on a
publicly reachable cache.

- default: `true`

### spec.redisConfiguration.dataPersistenceAuthenticationMethod

`enum`

How the cache authenticates to the storage account that holds RDB/AOF
persistence data. SAS (default) embeds a signed token in the
connection strings; MANAGED_IDENTITY uses the cache's identity block
instead -- no storage secret in the spec at all (pair with
spec.identity and grant the identity Storage Blob Data Contributor on
the account).

- rule: {"enum":{"definedOnly":true}}

Allowed values (use exactly as shown):

- `azure_redis_cache_persistence_auth_method_unspecified` -- Not specified -- Azure's default (SAS).
- `SAS` -- Signed SAS tokens embedded in the storage connection strings.
- `MANAGED_IDENTITY` -- The cache's managed identity (spec.identity) authenticates to storage -- no storage secret in the spec at all. Grant the identity "Storage Blob Data Contributor" on the persistence account.

### spec.redisConfiguration.rdbBackupEnabled

`bool`

PREMIUM ONLY -- enable RDB snapshot persistence: periodic full
snapshots written to a storage account, used to rebuild the cache
after a full outage. Requires rdb_storage_connection_string (or
MANAGED_IDENTITY persistence auth).

### spec.redisConfiguration.rdbBackupFrequency

`int32` · optional (explicit presence)

Minutes between RDB snapshots: 15, 30, 60, 360, 720, or 1440.

- rule: rdb_backup_frequency must be one of: 15, 30, 60, 360, 720, 1440 (minutes)

### spec.redisConfiguration.rdbBackupMaxSnapshotCount

`int32` · optional (explicit presence)

Maximum number of RDB snapshots retained before the oldest is
overwritten.

- rule: {"int32":{"gte":1}}

### spec.redisConfiguration.rdbStorageConnectionString

`string` · sensitive

Connection string of the storage account receiving RDB snapshots
(a secret -- it embeds the account key or SAS token). Not needed when
data_persistence_authentication_method is MANAGED_IDENTITY. Note:
Azure's API never echoes this value back.

### spec.redisConfiguration.aofBackupEnabled

`bool`

PREMIUM ONLY -- enable AOF (append-only file) persistence: every
write is logged to storage near-synchronously, giving much tighter
recovery points than RDB snapshots at the cost of write throughput.

### spec.redisConfiguration.aofStorageConnectionString0

`string` · sensitive

Connection string of the FIRST storage account receiving the AOF log
(a secret -- it embeds the account key or SAS token).

### spec.redisConfiguration.aofStorageConnectionString1

`string` · sensitive

Connection string of the SECOND storage account for AOF -- Azure
alternates between two accounts to keep log writes flowing during
storage maintenance (a secret).

### spec.redisConfiguration.storageAccountSubscriptionId

`string` · optional (explicit presence)

The subscription holding the persistence storage account, when it is
NOT the cache's own subscription (cross-subscription persistence).

- rule: {"string":{"uuid":true}}

### spec.identity

`AzureRedisCacheIdentity`

The cache's managed identity. Required (with a SystemAssigned or
UserAssigned entry) when redis_configuration selects MANAGED_IDENTITY
as the persistence auth method -- the identity is what writes RDB/AOF
snapshots to the storage account without a connection-string secret.

- rule: user_assigned_identity_ids is required with USER_ASSIGNED or SYSTEM_AND_USER_ASSIGNED, and must be empty with SYSTEM_ASSIGNED

### spec.identity.type

`enum`

The identity model: SYSTEM_ASSIGNED (Azure creates and rotates a
service principal bound to the cache's lifecycle), USER_ASSIGNED
(bring identities from user_assigned_identity_ids, shareable across
resources), or SYSTEM_AND_USER_ASSIGNED (both).

- rule: {"enum":{"definedOnly":true,"notIn":[0]}}

Allowed values (use exactly as shown):

- `azure_redis_cache_identity_type_unspecified` -- Not specified -- invalid; choose an explicit identity model.
- `SYSTEM_ASSIGNED` -- Azure creates a service principal bound to the cache's lifecycle.
- `USER_ASSIGNED` -- Bring your own AzureUserAssignedIdentity entries -- shareable across resources and grantable before the cache exists.
- `SYSTEM_AND_USER_ASSIGNED` -- Both a system-assigned principal and user-assigned identities.

### spec.identity.userAssignedIdentityIds

`[]string | valueFrom`

The user-assigned identities to attach -- required when (and only
meaningful when) type includes USER_ASSIGNED. Each entry references an
AzureUserAssignedIdentity's ARM id.

- references: AzureUserAssignedIdentity (`status.outputs.identity_id`)
- rule: write as {value: <literal>} or {valueFrom: {kind: AzureUserAssignedIdentity, name: <that resource's name>, fieldPath: status.outputs.identity_id}} -- a bare string does not parse

### spec.patchSchedules

`[]AzureRedisCachePatchSchedule`

Weekly maintenance windows during which Azure may apply Redis engine
patches and platform updates. Azure picks the window automatically when
unset; production caches should pin at least one low-traffic window.

### spec.patchSchedules[].dayOfWeek

`enum`

Day of the week for the maintenance window.

- rule: {"enum":{"definedOnly":true,"notIn":[0]}}

Allowed values (use exactly as shown):

- `azure_redis_cache_patch_schedule_day_unspecified` -- Not specified -- invalid; pick an explicit day.
- `MONDAY`
- `TUESDAY`
- `WEDNESDAY`
- `THURSDAY`
- `FRIDAY`
- `SATURDAY`
- `SUNDAY`

### spec.patchSchedules[].startHourUtc

`int32` · optional (explicit presence)

UTC hour at which the window opens (0-23). Default 0 (midnight UTC).

- rule: {"int32":{"lte":23,"gte":0}}

### spec.patchSchedules[].maintenanceWindow

`string` · optional (explicit presence)

How long the window stays open, as an ISO 8601 duration. Default
"PT5H" (five hours -- Azure's minimum recommended window).

- default: `PT5H`
- rule: maintenance_window must be an ISO 8601 duration, e.g. PT5H or PT6H30M

### spec.firewallRules

`[]AzureRedisCacheFirewallRule`

IP allow-list for the public endpoint. Only meaningful while
public_network_access_enabled is true and the cache is not
VNet-injected. Without rules, the public endpoint accepts any source
address that presents valid credentials.

### spec.firewallRules[].name

`string` · required

Rule name -- letters, digits, and underscores only (ARM rejects
hyphens in Redis firewall rule names).

- rule: firewall rule name must contain only letters, digits, and underscores (no hyphens)
- rule: {"required":true}

### spec.firewallRules[].startIp

`string` · required

First IPv4 address of the allowed range (inclusive).

- rule: {"required":true,"string":{"ipv4":true}}

### spec.firewallRules[].endIp

`string` · required

Last IPv4 address of the allowed range (inclusive). Equal to start_ip
for a single-address rule.

- rule: {"required":true,"string":{"ipv4":true}}

### spec.tenantSettings

`map<string, string>`

Tenant-level Azure platform settings passed through to the cache as
raw key/value pairs (distinct from redis_configuration, which is the
Redis engine's own configuration). Rarely needed; used by Microsoft
support scenarios and preview features.

### spec.tags

`map<string, string>`

Free-form tags applied to the cache, merged over the Planton-derived
resource tags (organization, environment, resource id); a user tag with
the same key wins. Tags are Azure's governance surface -- Azure Policy
enforces them and Microsoft Cost Management groups by them -- so carry
your org's ownership/cost-center conventions here. Updatable in place.

## Validation Rules

- `redis_cache_premium_capacity`: capacity must be 1-5 (P1-P5) on the PREMIUM tier
- `redis_cache_subnet_requires_premium`: subnet_id (VNet injection) requires the PREMIUM tier
- `redis_cache_static_ip_requires_subnet`: private_static_ip_address is only meaningful with subnet_id (VNet injection)
- `redis_cache_shard_count_requires_premium`: shard_count (clustering) requires the PREMIUM tier
- `redis_cache_replicas_require_premium`: replicas_per_primary requires the PREMIUM tier
- `redis_cache_shards_and_replicas_exclusive`: shard_count and replicas_per_primary cannot be combined
- `redis_cache_keys_off_requires_entra`: access_keys_authentication_enabled can only be false when redis_configuration.active_directory_authentication_enabled is true
- `redis_cache_auth_off_requires_subnet`: redis_configuration.authentication_enabled can only be false when the cache is VNet-injected (subnet_id set)
- `redis_cache_memory_dials_not_basic`: maxmemory_reserved, maxmemory_delta, and maxfragmentationmemory_reserved are not supported on the BASIC tier
- `redis_cache_rdb_requires_premium`: RDB persistence (rdb_backup_enabled) requires the PREMIUM tier
- `redis_cache_aof_requires_premium`: AOF persistence (aof_backup_enabled) requires the PREMIUM tier
- `redis_cache_rdb_requires_connection_string`: rdb_backup_enabled requires rdb_storage_connection_string (or data_persistence_authentication_method MANAGED_IDENTITY with an identity)
- `redis_cache_managed_identity_persistence_requires_identity`: data_persistence_authentication_method MANAGED_IDENTITY requires an identity block

## Outputs

Reference an output from another manifest as `valueFrom: {kind: AzureRedisCache, name: <resource-name>, fieldPath: status.outputs.<output>}`. A sensitive output is a secret the resource generates: on Planton it is kept in the organization's secret store and the output holds a `$secret/` reference, so feed it only to a sensitive field.

| Output | Type | Description |
|---|---|---|
| `status.outputs.redis_cache_id` | `string` | The Azure Resource Manager ID of the Redis cache. Format: /subscriptions/{sub}/resourceGroups/{rg}/providers/Microsoft.Cache/redis/{name} The reference target for everything that composes with the cache: AzureRedisLinkedServer (geo-replication), AzureRedisCacheAccessPolicy and AzureRedisCacheAccessPolicyAssignment (Entra data-plane RBAC), and AzurePrivateEndpoint (private connectivity). |
| `status.outputs.redis_cache_name` | `string` | The cache's name -- the DNS label of the endpoint and the value sibling resources address the cache by within its resource group. |
| `status.outputs.region` | `string` | The Azure region the cache lives in. AzureRedisLinkedServer references this as the linked cache's location, so geo-replication composes without hand-repeating the region. |
| `status.outputs.resource_group_name` | `string` | The resource group the cache lives in. |
| `status.outputs.hostname` | `string` | The cache's DNS hostname: {cache_name}.redis.cache.windows.net. Keyless (Entra) clients need only this -- tokens replace keys. |
| `status.outputs.port` | `int32` | The plaintext non-SSL port (6379). Only open when non_ssl_port_enabled is true. |
| `status.outputs.ssl_port` | `int32` | The TLS port (6380) -- the port every production client should use. |
| `status.outputs.primary_access_key` | `string` (sensitive) | The primary access key (a SECRET). Used as the password in Redis connection strings. Empty when access-keys authentication is disabled. |
| `status.outputs.secondary_access_key` | `string` (sensitive) | The secondary access key (a SECRET). Kept live so clients can be rotated to it while the primary is regenerated, and vice versa -- zero-downtime key rotation. |
| `status.outputs.primary_connection_string` | `string` (sensitive) | Ready-to-use primary connection string (a SECRET -- embeds the primary key): {hostname}:{ssl_port},password={key},ssl=True,abortConnect=False |
| `status.outputs.secondary_connection_string` | `string` (sensitive) | Ready-to-use secondary connection string (a SECRET -- embeds the secondary key), for the rotation window. |
| `status.outputs.identity_principal_id` | `string` | The system-assigned identity's principal (object) ID -- what RBAC grants target when the identity block requests SYSTEM_ASSIGNED. Empty otherwise. |

## References

Fields that can point at another resource's outputs:

| Field | Kind | Output |
|---|---|---|
| `spec.resourceGroup` | AzureResourceGroup | `status.outputs.resource_group_name` |
| `spec.subnetId` | AzureSubnet | `status.outputs.subnet_id` |
| `spec.identity.userAssignedIdentityIds` | AzureUserAssignedIdentity | `status.outputs.identity_id` |

## Referenced By

Fields on other kinds that can point at this resource:

| Kind | Field | Reads |
|---|---|---|
| AzureRedisCacheAccessPolicy | `spec.redisCacheId` | `status.outputs.redis_cache_id` |
| AzureRedisCacheAccessPolicyAssignment | `spec.redisCacheId` | `status.outputs.redis_cache_id` |
| AzureRedisLinkedServer | `spec.targetRedisCacheId` | `status.outputs.redis_cache_id` |
| AzureRedisLinkedServer | `spec.linkedRedisCacheId` | `status.outputs.redis_cache_id` |
| AzureRedisLinkedServer | `spec.linkedRedisCacheLocation` | `status.outputs.region` |

## See Also

- [Overview](../README.md)
