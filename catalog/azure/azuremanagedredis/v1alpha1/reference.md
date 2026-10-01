# AzureManagedRedis

> Generated from the protobuf schema by `make generate-reference` -- do not
> edit by hand. To change a fact on this page, change the proto field comment
> or validation rule it is derived from, then regenerate.

**apiVersion**: `azure.planton.dev/v1alpha1`

**AzureManagedRedisSpec** defines the configuration for creating an
Azure Managed Redis instance -- Azure's current-generation Redis
service, built on Redis Enterprise (the Microsoft.Cache/redisEnterprise
ARM family). Azure is retiring classic Azure Cache for Redis; Managed
Redis is the target for NEW Redis deployments, and it is where the
capabilities the classic service never had live: customer-managed-key
encryption, active (multi-primary) geo-replication, Redis modules
(search, JSON, time series, bloom filters), and a keyless-by-default
authentication posture.

**Anatomy.** A Managed Redis instance is a CLUSTER (compute, load
balancer, network, TLS endpoint) plus its DEFAULT DATABASE (the Redis
process itself -- eviction, clustering, modules, persistence,
authentication). Azure maps them 1-to-1 today, so both live in this one
spec: the cluster fields at the top level and the Redis-process fields
under `default_database`.

**Sizing.** `sku_name` picks a tier family and size in one value:
BALANCED (B0-B1000) general-purpose, COMPUTE_OPTIMIZED (X3-X700) for
high ops/sec on smaller datasets, MEMORY_OPTIMIZED (M10-M2000) for
large datasets with moderate throughput, FLASH_OPTIMIZED (A250-A4500)
for very large datasets tiered onto NVMe flash. Azure permits some SKU
changes in place (it validates the target against the live instance);
when a change is not allowed the instance is replaced.

**Authentication.** Access keys are DISABLED by default -- the reverse
of classic Redis. The default posture is keyless: clients authenticate
with Microsoft Entra tokens under AzureManagedRedisAccessPolicyAssignment
grants. Set default_database.access_keys_authentication_enabled to true
only when a client genuinely cannot use Entra tokens.

**ForceNew fields** (changing these replaces the instance):
`cluster_name`, `high_availability_enabled`, and any `sku_name` change
Azure's scaling validator rejects. Inside `default_database`, changing
`clustering_policy`, `geo_replication_group_name`, or `modules`
recreates the DATABASE in place (data is lost and the endpoint is
briefly unavailable) without replacing the cluster.

## Example

```yaml
apiVersion: azure.planton.dev/v1alpha1
kind: AzureManagedRedis
metadata:
  name: test-managed-redis
spec:
  region: eastus
  resourceGroup:
    value: test-rg
  clusterName: planton-hack-managed-redis
  # Memory-optimized tier exercises a non-default SKU family mapping.
  skuName: MEMORY_OPTIMIZED_M10
  highAvailabilityEnabled: true
  publicNetworkAccessEnabled: false
  # Customer-managed key: the VERSIONED Key Vault key id; the same
  # identity is attached through the identity block (ARM pairing).
  customerManagedKey:
    keyVaultKeyId:
      value: https://planton-hack-kv.vault.azure.net/keys/redis-cmk/0123456789abcdef0123456789abcdef
    userAssignedIdentityId:
      value: /subscriptions/00000000-0000-0000-0000-000000000000/resourceGroups/test-rg/providers/Microsoft.ManagedIdentity/userAssignedIdentities/redis-cmk-identity
  identity:
    type: USER_ASSIGNED
    userAssignedIdentityIds:
      - value: /subscriptions/00000000-0000-0000-0000-000000000000/resourceGroups/test-rg/providers/Microsoft.ManagedIdentity/userAssignedIdentities/redis-cmk-identity
  defaultDatabase:
    # Keys enabled here to exercise the non-default auth path; the
    # keyless posture (the default) is exercised by the presets.
    accessKeysAuthenticationEnabled: true
    clientProtocol: ENCRYPTED
    # EnterpriseCluster + NoEviction exercise the RediSearch pairing
    # contracts alongside the module list.
    clusteringPolicy: ENTERPRISE_CLUSTER
    evictionPolicy: NO_EVICTION
    modules:
      - name: RediSearch
      - name: RedisJSON
      - name: RedisBloom
        args: ERROR_RATE 0.01 INITIAL_SIZE 400
    # RDB persistence exercises the frequency vocabulary seam.
    persistenceRedisDatabaseBackupFrequency: 6h
  tags:
    cost-center: platform
```

## Spec Fields

| Path | Type | Required | Default | References |
|---|---|---|---|---|
| `spec.region` | `string` | yes |  |  |
| `spec.resourceGroup` | `string \| valueFrom` | yes |  | AzureResourceGroup (`status.outputs.resource_group_name`) |
| `spec.clusterName` | `string` | yes |  |  |
| `spec.skuName` | `enum` |  |  |  |
| `spec.highAvailabilityEnabled` | `bool` |  | `true` |  |
| `spec.customerManagedKey` | `AzureManagedRedisCustomerManagedKey` |  |  |  |
| `spec.customerManagedKey.keyVaultKeyId` | `string \| valueFrom` | yes |  | AzureKeyVaultKey (`status.outputs.key_id`) |
| `spec.customerManagedKey.userAssignedIdentityId` | `string \| valueFrom` | yes |  | AzureUserAssignedIdentity (`status.outputs.identity_id`) |
| `spec.identity` | `AzureManagedRedisIdentity` |  |  |  |
| `spec.identity.type` | `enum` |  |  |  |
| `spec.identity.userAssignedIdentityIds` | `[]string \| valueFrom` |  |  | AzureUserAssignedIdentity (`status.outputs.identity_id`) |
| `spec.publicNetworkAccessEnabled` | `bool` |  | `true` |  |
| `spec.defaultDatabase` | `AzureManagedRedisDatabase` | yes |  |  |
| `spec.defaultDatabase.accessKeysAuthenticationEnabled` | `bool` |  |  |  |
| `spec.defaultDatabase.clientProtocol` | `enum` |  |  |  |
| `spec.defaultDatabase.clusteringPolicy` | `enum` |  |  |  |
| `spec.defaultDatabase.evictionPolicy` | `enum` |  |  |  |
| `spec.defaultDatabase.geoReplicationGroupName` | `string` |  |  |  |
| `spec.defaultDatabase.modules` | `[]AzureManagedRedisModule` |  |  |  |
| `spec.defaultDatabase.modules[].name` | `string` | yes |  |  |
| `spec.defaultDatabase.modules[].args` | `string` |  |  |  |
| `spec.defaultDatabase.persistenceAppendOnlyFileBackupFrequency` | `string` |  |  |  |
| `spec.defaultDatabase.persistenceRedisDatabaseBackupFrequency` | `string` |  |  |  |
| `spec.tags` | `map<string, string>` |  |  |  |

## Field Details

### spec.region

`string` · required

The Azure region where the Managed Redis instance will be created.
Managed Redis is newer than classic Redis and not yet in every
region -- consult Azure's product-availability table for the current
footprint. Examples: "eastus", "westus2", "westeurope".

- rule: {"required":true,"string":{"minLen":"1"}}

### spec.resourceGroup

`string | valueFrom` · required

The Azure Resource Group where the Managed Redis instance will be
created. Can be a literal string or a reference to an
AzureResourceGroup output.

- references: AzureResourceGroup (`status.outputs.resource_group_name`)
- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: AzureResourceGroup, name: <that resource's name>, fieldPath: status.outputs.resource_group_name}} -- a bare string does not parse

### spec.clusterName

`string` · required

The instance's name -- it becomes the DNS endpoint
`{cluster_name}.{region}.redis.azure.net`. 3-63 letters, digits, and
hyphens; must start and end with a letter or digit; no consecutive
hyphens. Changing the name replaces the instance.

- rule: cluster_name must be 3-63 letters, digits, and hyphens, start and end with a letter or digit, and never repeat a hyphen
- rule: {"required":true,"string":{"minLen":"3","maxLen":"63"}}

### spec.skuName

`enum`

The tier family and size in one value. BALANCED is the
general-purpose default choice; COMPUTE_OPTIMIZED trades memory for
throughput; MEMORY_OPTIMIZED trades throughput for memory;
FLASH_OPTIMIZED tiers very large datasets onto NVMe flash. The
number is the instance's memory in GB (B0 = 0.5 GB, B1 = 1 GB, and
so on). Geo-replication requires BALANCED_B3 or larger (enforced
here, exactly as Azure enforces it). Azure allows some SKU changes
in place -- it validates the target size against the live instance
at apply time and replaces the instance when the change is not
scalable (a cross-resource check that cannot be evaluated
statically).

- rule: {"enum":{"definedOnly":true,"notIn":[0]}}

Allowed values (use exactly as shown):

- `azure_managed_redis_sku_unspecified` -- Not specified -- invalid; pick an explicit tier and size.
- `BALANCED_B0` -- BALANCED (B-family): general-purpose memory-to-throughput ratio -- the right starting family for most workloads.
- `BALANCED_B1`
- `BALANCED_B3`
- `BALANCED_B5`
- `BALANCED_B10`
- `BALANCED_B20`
- `BALANCED_B50`
- `BALANCED_B100`
- `BALANCED_B150`
- `BALANCED_B250`
- `BALANCED_B350`
- `BALANCED_B500`
- `BALANCED_B700`
- `BALANCED_B1000`
- `COMPUTE_OPTIMIZED_X3` -- COMPUTE_OPTIMIZED (X-family): maximum ops/sec on smaller datasets.
- `COMPUTE_OPTIMIZED_X5`
- `COMPUTE_OPTIMIZED_X10`
- `COMPUTE_OPTIMIZED_X20`
- `COMPUTE_OPTIMIZED_X50`
- `COMPUTE_OPTIMIZED_X100`
- `COMPUTE_OPTIMIZED_X150`
- `COMPUTE_OPTIMIZED_X250`
- `COMPUTE_OPTIMIZED_X350`
- `COMPUTE_OPTIMIZED_X500`
- `COMPUTE_OPTIMIZED_X700`
- `MEMORY_OPTIMIZED_M10` -- MEMORY_OPTIMIZED (M-family): large in-memory datasets with moderate throughput needs.
- `MEMORY_OPTIMIZED_M20`
- `MEMORY_OPTIMIZED_M50`
- `MEMORY_OPTIMIZED_M100`
- `MEMORY_OPTIMIZED_M150`
- `MEMORY_OPTIMIZED_M250`
- `MEMORY_OPTIMIZED_M350`
- `MEMORY_OPTIMIZED_M500`
- `MEMORY_OPTIMIZED_M700`
- `MEMORY_OPTIMIZED_M1000`
- `MEMORY_OPTIMIZED_M1500`
- `MEMORY_OPTIMIZED_M2000`
- `FLASH_OPTIMIZED_A250` -- FLASH_OPTIMIZED (A-family): very large datasets tiered between RAM and NVMe flash -- the largest capacities at the lowest cost per GB.
- `FLASH_OPTIMIZED_A500`
- `FLASH_OPTIMIZED_A700`
- `FLASH_OPTIMIZED_A1000`
- `FLASH_OPTIMIZED_A1500`
- `FLASH_OPTIMIZED_A2000`
- `FLASH_OPTIMIZED_A4500`

### spec.highAvailabilityEnabled

`bool` · optional (explicit presence)

Whether the instance runs with a replica for high availability and
the 99.999% zone-redundant SLA. Default true -- disabling it halves
the cost but removes the replica AND the SLA, which is only
appropriate for dev/test. Fixed at creation.

- default: `true`

### spec.customerManagedKey

`AzureManagedRedisCustomerManagedKey`

Customer-managed-key (CMK) encryption for data at rest. Omit the
block for Microsoft-managed keys (the default). Bringing your own
key requires a user-assigned identity that can wrap/unwrap the key
-- that same identity must also be attached to the instance through
the identity block below (an ARM pairing enforced at apply time).

### spec.customerManagedKey.keyVaultKeyId

`string | valueFrom` · required

The Key Vault key that wraps the data-encryption key, by its
VERSIONED data-plane ID (Managed Redis pins the key version;
rotate by updating this reference). References an
AzureKeyVaultKey's key_id output.

- references: AzureKeyVaultKey (`status.outputs.key_id`)
- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: AzureKeyVaultKey, name: <that resource's name>, fieldPath: status.outputs.key_id}} -- a bare string does not parse

### spec.customerManagedKey.userAssignedIdentityId

`string | valueFrom` · required

The user-assigned identity that authenticates to Key Vault to
wrap/unwrap the key -- grant it wrap/unwrap permissions on the key
before deploying. The SAME identity must also be attached to the
instance through the identity block (an ARM pairing enforced at
apply time). References an AzureUserAssignedIdentity's ARM id.

- references: AzureUserAssignedIdentity (`status.outputs.identity_id`)
- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: AzureUserAssignedIdentity, name: <that resource's name>, fieldPath: status.outputs.identity_id}} -- a bare string does not parse

### spec.identity

`AzureManagedRedisIdentity`

The instance's managed identity. Required (with a USER_ASSIGNED
entry carrying the CMK identity) when customer_managed_key is set;
also useful on its own for future identity-based integrations.

- rule: user_assigned_identity_ids is required with USER_ASSIGNED or SYSTEM_AND_USER_ASSIGNED, and must be empty with SYSTEM_ASSIGNED

### spec.identity.type

`enum`

The identity model: SYSTEM_ASSIGNED (Azure creates and rotates a
service principal bound to the instance's lifecycle), USER_ASSIGNED
(bring identities from user_assigned_identity_ids, shareable across
resources -- what CMK requires), or SYSTEM_AND_USER_ASSIGNED (both).

- rule: {"enum":{"definedOnly":true,"notIn":[0]}}

Allowed values (use exactly as shown):

- `azure_managed_redis_identity_type_unspecified` -- Not specified -- invalid; choose an explicit identity model.
- `SYSTEM_ASSIGNED` -- Azure creates a service principal bound to the instance's lifecycle.
- `USER_ASSIGNED` -- Bring your own AzureUserAssignedIdentity entries -- shareable across resources and grantable before the instance exists. What customer-managed-key encryption requires.
- `SYSTEM_AND_USER_ASSIGNED` -- Both a system-assigned principal and user-assigned identities.

### spec.identity.userAssignedIdentityIds

`[]string | valueFrom`

The user-assigned identities to attach -- required when (and only
meaningful when) type includes USER_ASSIGNED. Each entry references
an AzureUserAssignedIdentity's ARM id.

- references: AzureUserAssignedIdentity (`status.outputs.identity_id`)
- rule: write as {value: <literal>} or {valueFrom: {kind: AzureUserAssignedIdentity, name: <that resource's name>, fieldPath: status.outputs.identity_id}} -- a bare string does not parse

### spec.publicNetworkAccessEnabled

`bool` · optional (explicit presence)

Whether the instance answers on its public endpoint. Default true.
Set false to force all traffic through an AzurePrivateEndpoint --
Managed Redis has no VNet injection or IP firewall; Private Link is
its only network-isolation mechanism.

- default: `true`

### spec.defaultDatabase

`AzureManagedRedisDatabase` · required

The Redis process itself: authentication, clustering, eviction,
modules, geo-replication membership, and persistence. Required --
an instance without its database is an administrative transient,
not a deployable target (Azure rejects creating one).

- rule: {"required":true}
- rule: AOF and RDB persistence are mutually exclusive -- set at most one backup frequency
- rule: persistence cannot be enabled on a geo-replicated database (geo_replication_group_name)
- rule: a geo-replicated database supports only the RediSearch and RedisJSON modules
- rule: the RediSearch module requires eviction_policy NO_EVICTION
- rule: the RediSearch module requires clustering_policy ENTERPRISE_CLUSTER
- rule: each module can be enabled at most once

### spec.defaultDatabase.accessKeysAuthenticationEnabled

`bool`

Whether the shared access keys authenticate clients. Default FALSE
-- Managed Redis is keyless-first: clients present Microsoft Entra
tokens under AzureManagedRedisAccessPolicyAssignment grants, and no
secret exists to leak or rotate. Enable keys only for clients that
genuinely cannot use Entra tokens; the key outputs are populated
only while this is true. Updatable in place.

### spec.defaultDatabase.clientProtocol

`enum`

Whether clients must connect over TLS. ENCRYPTED (the default) is
right for everything except legacy clients that cannot speak TLS --
PLAINTEXT sends commands and credentials unencrypted. Updatable in
place.

- rule: {"enum":{"definedOnly":true}}

Allowed values (use exactly as shown):

- `azure_managed_redis_client_protocol_unspecified` -- Not specified -- deploys ENCRYPTED, the TLS-only default.
- `ENCRYPTED` -- TLS-only connections -- the right answer for everything except legacy clients that cannot speak TLS.
- `PLAINTEXT` -- Unencrypted connections: commands AND credentials travel in the clear. Legacy clients only.

### spec.defaultDatabase.clusteringPolicy

`enum`

How keys are distributed across the cluster's shards. OSS_CLUSTER
(the default) exposes the standard Redis Cluster protocol -- best
throughput, requires cluster-aware clients. ENTERPRISE_CLUSTER
proxies all shards behind a single endpoint so any client works
(required by the RediSearch module). NO_CLUSTER disables sharding
entirely (memory <= 25 GB). CHANGING THIS RECREATES THE DATABASE:
data is lost and the endpoint is briefly unavailable.

- rule: {"enum":{"definedOnly":true}}

Allowed values (use exactly as shown):

- `azure_managed_redis_clustering_policy_unspecified` -- Not specified -- deploys OSS_CLUSTER, Azure's default.
- `ENTERPRISE_CLUSTER` -- All shards proxied behind a single endpoint: any Redis client works, at some throughput cost. Required by the RediSearch module.
- `OSS_CLUSTER` -- The standard Redis Cluster protocol: best throughput; clients must be cluster-aware. Azure's default.
- `NO_CLUSTER` -- No sharding at all -- a single Redis process. Only for databases up to 25 GB of memory.

### spec.defaultDatabase.evictionPolicy

`enum`

Eviction policy when the database reaches its memory limit, in
Azure's vocabulary. Default VOLATILE_LRU (evict TTL-bearing keys,
least-recently-used first). Guidance: ALL_KEYS_LRU for pure caches
where every key is expendable; NO_EVICTION when data must never
silently disappear (writes fail when full -- required by the
RediSearch module). Updatable in place.

- rule: {"enum":{"definedOnly":true}}

Allowed values (use exactly as shown):

- `azure_managed_redis_eviction_policy_unspecified` -- Not specified -- deploys VOLATILE_LRU, Azure's default.
- `ALL_KEYS_LFU` -- Evict any key, least-frequently-used first.
- `ALL_KEYS_LRU` -- Evict any key, least-recently-used first -- the usual choice for pure caches where every key is expendable.
- `ALL_KEYS_RANDOM` -- Evict any key, at random.
- `VOLATILE_LRU` -- Evict only TTL-bearing keys, least-recently-used first -- Azure's default.
- `VOLATILE_LFU` -- Evict only TTL-bearing keys, least-frequently-used first.
- `VOLATILE_TTL` -- Evict only TTL-bearing keys, shortest time-to-live first.
- `VOLATILE_RANDOM` -- Evict only TTL-bearing keys, at random.
- `NO_EVICTION` -- Never evict: writes fail when memory is full. Required by the RediSearch module; the right answer when data must never silently disappear.

### spec.defaultDatabase.geoReplicationGroupName

`string`

Joining a named ACTIVE geo-replication group: every Managed Redis
database created with the same group name (across regions and
subscriptions) becomes a multi-primary replica set -- all members
accept writes, with conflict-free merge semantics. Setting the name
here creates the group with this database as its first member; link
further members with AzureManagedRedisGeoReplication. Requires
BALANCED_B3 or larger, only the RediSearch/RedisJSON modules, and no
persistence (all enforced here, exactly as Azure enforces them).
1-63 letters, digits, and hyphens; must start and end with a letter
or digit; no consecutive hyphens. CHANGING THIS RECREATES THE
DATABASE.

- rule: geo_replication_group_name must be 1-63 letters, digits, and hyphens, start and end with a letter or digit, and never repeat a hyphen

### spec.defaultDatabase.modules

`[]AzureManagedRedisModule`

Redis modules to enable, up to 4: RediSearch (secondary indexing +
full-text search), RedisJSON (native JSON documents), RedisBloom
(probabilistic filters), RedisTimeSeries (time-series data).
Modules exist only on Managed Redis -- classic Redis never had
them. CHANGING THE MODULE SET RECREATES THE DATABASE.

- rule: {"repeated":{"maxItems":"4"}}

### spec.defaultDatabase.modules[].name

`string` · required

The module's name, in Azure's exact catalog spelling.

- rule: module name must be one of: RediSearch, RedisJSON, RedisBloom, RedisTimeSeries
- rule: {"required":true}

### spec.defaultDatabase.modules[].args

`string`

Optional module configuration arguments, passed through verbatim
(e.g. "ERROR_RATE 0.01 INITIAL_SIZE 400" for RedisBloom). Most
deployments leave this empty.

### spec.defaultDatabase.persistenceAppendOnlyFileBackupFrequency

`string` · optional (explicit presence)

The frequency of append-only-file (AOF) persistence backups --
every write is logged near-synchronously, giving the tightest
recovery point. Setting a value ENABLES AOF persistence; "1s" is
the only frequency Azure currently offers. Mutually exclusive with
RDB persistence and with geo-replication (enforced here, exactly as
Azure enforces it).

- rule: persistence_append_only_file_backup_frequency must be: 1s

### spec.defaultDatabase.persistenceRedisDatabaseBackupFrequency

`string` · optional (explicit presence)

The frequency of Redis-database (RDB) snapshot persistence --
periodic full snapshots, used to rebuild the database after a full
outage. Setting a value ENABLES RDB persistence. Mutually exclusive
with AOF persistence and with geo-replication (enforced here,
exactly as Azure enforces it).

- rule: persistence_redis_database_backup_frequency must be one of: 1h, 6h, 12h

### spec.tags

`map<string, string>`

Free-form tags applied to the instance, merged over the
Planton-derived resource tags (organization, environment, resource
id); a user tag with the same key wins. Tags are Azure's governance
surface -- Azure Policy enforces them and Microsoft Cost Management
groups by them. Updatable in place.

## Validation Rules

- `managed_redis_geo_replication_sku_floor`: geo_replication_group_name requires BALANCED_B3 or a larger sku (Azure does not support geo-replication on BALANCED_B0/BALANCED_B1)

## Outputs

Reference an output from another manifest as `valueFrom: {kind: AzureManagedRedis, name: <resource-name>, fieldPath: status.outputs.<output>}`. A sensitive output is a secret the resource generates: on Planton it is kept in the organization's secret store and the output holds a `$secret/` reference, so feed it only to a sensitive field.

| Output | Type | Description |
|---|---|---|
| `status.outputs.managed_redis_id` | `string` | The Azure Resource Manager ID of the Managed Redis cluster. Format: /subscriptions/{sub}/resourceGroups/{rg}/providers/Microsoft.Cache/redisEnterprise/{name} The reference target for everything that composes with the instance: AzureManagedRedisGeoReplication (group membership), AzureManagedRedisAccessPolicyAssignment (Entra data-plane grants), and AzurePrivateEndpoint (private connectivity). |
| `status.outputs.managed_redis_name` | `string` | The instance's name -- the DNS label of the endpoint and the value sibling resources address the cluster by within its resource group. |
| `status.outputs.region` | `string` | The Azure region the instance lives in. |
| `status.outputs.resource_group_name` | `string` | The resource group the instance lives in. |
| `status.outputs.hostname` | `string` | The instance's DNS hostname: {cluster_name}.{region}.redis.azure.net. Keyless (Entra) clients need only this and the port -- tokens replace keys. |
| `status.outputs.database_id` | `string` | The Azure Resource Manager ID of the default database -- the cluster ID with "/databases/default" appended. The scope Entra access-policy grants and geo-replication links operate on. |
| `status.outputs.port` | `int32` | The TCP port the database listens on (10000 -- Managed Redis does not use classic Redis's 6379/6380 ports). |
| `status.outputs.primary_access_key` | `string` (sensitive) | The primary access key (a SECRET). Used as the password in Redis connection strings. Empty when access-keys authentication is disabled -- the keyless default. |
| `status.outputs.secondary_access_key` | `string` (sensitive) | The secondary access key (a SECRET). Kept live so clients can be rotated to it while the primary is regenerated, and vice versa -- zero-downtime key rotation. Empty when access-keys authentication is disabled. |
| `status.outputs.identity_principal_id` | `string` | The system-assigned identity's principal (object) ID -- what RBAC grants target when the identity block requests SYSTEM_ASSIGNED. Empty otherwise. |

## References

Fields that can point at another resource's outputs:

| Field | Kind | Output |
|---|---|---|
| `spec.resourceGroup` | AzureResourceGroup | `status.outputs.resource_group_name` |
| `spec.customerManagedKey.keyVaultKeyId` | AzureKeyVaultKey | `status.outputs.key_id` |
| `spec.customerManagedKey.userAssignedIdentityId` | AzureUserAssignedIdentity | `status.outputs.identity_id` |
| `spec.identity.userAssignedIdentityIds` | AzureUserAssignedIdentity | `status.outputs.identity_id` |

## Referenced By

Fields on other kinds that can point at this resource:

| Kind | Field | Reads |
|---|---|---|
| AzureManagedRedisAccessPolicyAssignment | `spec.managedRedisId` | `status.outputs.managed_redis_id` |
| AzureManagedRedisGeoReplication | `spec.managedRedisId` | `status.outputs.managed_redis_id` |
| AzureManagedRedisGeoReplication | `spec.linkedManagedRedisIds` | `status.outputs.managed_redis_id` |

## See Also

- [Overview](../README.md)
