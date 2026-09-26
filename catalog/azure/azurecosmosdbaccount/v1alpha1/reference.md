# AzureCosmosdbAccount

> Generated from the protobuf schema by `make generate-reference` -- do not
> edit by hand. To change a fact on this page, change the proto field comment
> or validation rule it is derived from, then regenerate.

**apiVersion**: `azure.planton.dev/v1alpha1`

**AzureCosmosdbAccountSpec** defines the configuration for creating an
Azure Cosmos DB account -- the globally distributed, multi-model database
account that owns regions, consistency, network posture, encryption, and
backup for everything stored inside it.

The account is the governance boundary: which regions data lives in,
how reads are ordered (consistency), who can reach it (network rules,
public access, local auth), how it is encrypted (customer-managed keys),
and how it is backed up (periodic or continuous with point-in-time
restore) are all account-level decisions. The data containers are
first-class kinds referencing the account: AzureCosmosdbSqlDatabase /
AzureCosmosdbSqlContainer for the SQL (NoSQL) API and
AzureCosmosdbMongoDatabase / AzureCosmosdbMongoCollection for the
MongoDB API.

**API selection**: `kind` picks the wire protocol the account speaks --
GLOBAL_DOCUMENT_DB (the SQL/NoSQL API, the default) or MONGO_DB (the
MongoDB-compatible API). Cassandra, Gremlin, and Table run on a
GLOBAL_DOCUMENT_DB account through the matching capability
(ENABLE_CASSANDRA / ENABLE_GREMLIN / ENABLE_TABLE).

**Consistency**: Cosmos DB's five well-defined levels range from STRONG
(linearizable) to EVENTUAL (highest throughput). SESSION -- read-your-
writes within a session -- is Azure's recommended default for most
applications.

**Throughput**: capacity is measured in Request Units per second (RU/s)
and is provisioned on the databases and containers (or replaced by
pay-per-request serverless via ENABLE_SERVERLESS). The account-level
`capacity.total_throughput_limit` caps the sum, which is the guardrail
against runaway provisioning cost.

**ForceNew fields** (changing these destroys and recreates the account):
`account_name`, `region`, `kind`, `free_tier_enabled`, `key_vault_key_id`,
`create_mode`, and everything inside `restore`. Additionally one-way:
`analytical_storage_enabled` recreates only when DISABLING, and
`backup.type` recreates only when changing CONTINUOUS back to PERIODIC.

## Example

```yaml
apiVersion: azure.planton.dev/v1alpha1
kind: AzureCosmosdbAccount
metadata:
  name: test-cosmos-account
spec:
  region: eastus
  resourceGroup:
    value: test-rg
  accountName: planton-hack-cosmos
  # Exercises the kind enum mapping (GLOBAL_DOCUMENT_DB is also the
  # default; stated to prove the seam renders).
  kind: GLOBAL_DOCUMENT_DB
  # Exercises the BoundedStaleness dials (single-region, so the
  # multi-region floors do not apply).
  consistencyPolicy:
    consistencyLevel: BOUNDED_STALENESS
    maxIntervalInSeconds: 600
    maxStalenessPrefix: 200000
  geoLocations:
    - location: eastus
      failoverPriority: 0
      zoneRedundant: false
  # Exercises the capability enum mapping, including a wire value that
  # breaks the EnableX convention (DeleteAllItemsByPartitionKey).
  capabilities:
    - ENABLE_NO_SQL_VECTOR_SEARCH
    - DELETE_ALL_ITEMS_BY_PARTITION_KEY
  automaticFailoverEnabled: true
  # Exercises the network posture: VNet filter + subnet rule + IP filter
  # + bypass ids.
  publicNetworkAccessEnabled: true
  isVirtualNetworkFilterEnabled: true
  virtualNetworkRules:
    - subnetId:
        value: /subscriptions/00000000-0000-0000-0000-000000000000/resourceGroups/test-rg/providers/Microsoft.Network/virtualNetworks/test-vnet/subnets/data
      ignoreMissingVnetServiceEndpoint: true
  ipRangeFilter:
    - 104.42.195.92
    - 10.0.0.0/16
  networkAclBypassForAzureServices: true
  networkAclBypassIds:
    - /subscriptions/00000000-0000-0000-0000-000000000000/resourceGroups/test-rg/providers/Microsoft.Synapse/workspaces/test-syn
  # Exercises the periodic-backup enum seams (interval/retention/ZONE
  # redundancy).
  backup:
    type: PERIODIC
    intervalInMinutes: 240
    retentionInHours: 24
    storageRedundancy: ZONE
  # Exercises the identity + composed default-identity seam and the CMK
  # key reference.
  identity:
    type: USER_ASSIGNED
    identityIds:
      - value: /subscriptions/00000000-0000-0000-0000-000000000000/resourceGroups/test-rg/providers/Microsoft.ManagedIdentity/userAssignedIdentities/cosmos-cmk
  defaultIdentity:
    type: USER_ASSIGNED_DEFAULT
    userAssignedIdentityId:
      value: /subscriptions/00000000-0000-0000-0000-000000000000/resourceGroups/test-rg/providers/Microsoft.ManagedIdentity/userAssignedIdentities/cosmos-cmk
  keyVaultKeyId:
    value: https://planton-hack-vault.vault.azure.net/keys/cosmos-cmk
  # Exercises the analytical-store pairing and the capacity guardrail.
  analyticalStorageEnabled: true
  analyticalStorage:
    schemaType: WELL_DEFINED
  capacity:
    totalThroughputLimit: 20000
  # Exercises the key-posture inversions (local_authentication_disabled
  # on the wire).
  accessKeyMetadataWritesEnabled: false
  localAuthenticationEnabled: true
  # Exercises the TLS-floor enum seam. Tls12 is the only value the
  # provider accepts (the legacy Tls/Tls11 floors were retired).
  minimalTlsVersion: TLS_1_2
  burstCapacityEnabled: true
  partitionMergeEnabled: true
  # Exercises the CORS block rendering.
  corsRule:
    allowedOrigins:
      - https://app.example.com
    allowedMethods:
      - GET
      - POST
    allowedHeaders:
      - "*"
    exposedHeaders:
      - "*"
    maxAgeInSeconds: 3600
  tags:
    cost-center: data-platform
```

## Spec Fields

| Path | Type | Required | Default | References |
|---|---|---|---|---|
| `spec.region` | `string` | yes |  |  |
| `spec.resourceGroup` | `string \| valueFrom` | yes |  | AzureResourceGroup (`status.outputs.resource_group_name`) |
| `spec.accountName` | `string` | yes |  |  |
| `spec.kind` | `enum` |  |  |  |
| `spec.consistencyPolicy` | `AzureCosmosdbAccountConsistencyPolicy` | yes |  |  |
| `spec.consistencyPolicy.consistencyLevel` | `enum` |  |  |  |
| `spec.consistencyPolicy.maxIntervalInSeconds` | `int32` |  | `5` |  |
| `spec.consistencyPolicy.maxStalenessPrefix` | `int32` |  | `100` |  |
| `spec.geoLocations` | `[]AzureCosmosdbAccountGeoLocation` | yes |  |  |
| `spec.geoLocations[].location` | `string` | yes |  |  |
| `spec.geoLocations[].failoverPriority` | `int32` |  |  |  |
| `spec.geoLocations[].zoneRedundant` | `bool` |  | `false` |  |
| `spec.capabilities` | `[]enum` |  |  |  |
| `spec.freeTierEnabled` | `bool` |  | `false` |  |
| `spec.automaticFailoverEnabled` | `bool` |  | `false` |  |
| `spec.multipleWriteLocationsEnabled` | `bool` |  | `false` |  |
| `spec.publicNetworkAccessEnabled` | `bool` |  | `true` |  |
| `spec.isVirtualNetworkFilterEnabled` | `bool` |  | `false` |  |
| `spec.virtualNetworkRules` | `[]AzureCosmosdbAccountVirtualNetworkRule` |  |  |  |
| `spec.virtualNetworkRules[].subnetId` | `string \| valueFrom` | yes |  | AzureSubnet (`status.outputs.subnet_id`) |
| `spec.virtualNetworkRules[].ignoreMissingVnetServiceEndpoint` | `bool` |  | `false` |  |
| `spec.ipRangeFilter` | `[]string` |  |  |  |
| `spec.backup` | `AzureCosmosdbAccountBackup` |  |  |  |
| `spec.backup.type` | `enum` | yes |  |  |
| `spec.backup.tier` | `enum` |  |  |  |
| `spec.backup.intervalInMinutes` | `int32` |  |  |  |
| `spec.backup.retentionInHours` | `int32` |  |  |  |
| `spec.backup.storageRedundancy` | `enum` |  |  |  |
| `spec.mongoServerVersion` | `enum` |  |  |  |
| `spec.identity` | `AzureCosmosdbAccountIdentity` |  |  |  |
| `spec.identity.type` | `enum` | yes |  |  |
| `spec.identity.identityIds` | `[]string \| valueFrom` |  |  | AzureUserAssignedIdentity (`status.outputs.identity_id`) |
| `spec.defaultIdentity` | `AzureCosmosdbAccountDefaultIdentity` |  |  |  |
| `spec.defaultIdentity.type` | `enum` | yes |  |  |
| `spec.defaultIdentity.userAssignedIdentityId` | `string \| valueFrom` |  |  | AzureUserAssignedIdentity (`status.outputs.identity_id`) |
| `spec.keyVaultKeyId` | `string \| valueFrom` |  |  | AzureKeyVaultKey (`status.outputs.versionless_id`) |
| `spec.analyticalStorageEnabled` | `bool` |  | `false` |  |
| `spec.analyticalStorage` | `AzureCosmosdbAccountAnalyticalStorage` |  |  |  |
| `spec.analyticalStorage.schemaType` | `enum` | yes |  |  |
| `spec.capacity` | `AzureCosmosdbAccountCapacity` |  |  |  |
| `spec.capacity.totalThroughputLimit` | `int32` |  |  |  |
| `spec.accessKeyMetadataWritesEnabled` | `bool` |  | `true` |  |
| `spec.localAuthenticationEnabled` | `bool` |  | `true` |  |
| `spec.minimalTlsVersion` | `enum` |  |  |  |
| `spec.networkAclBypassForAzureServices` | `bool` |  | `false` |  |
| `spec.networkAclBypassIds` | `[]string` |  |  |  |
| `spec.burstCapacityEnabled` | `bool` |  | `false` |  |
| `spec.partitionMergeEnabled` | `bool` |  | `false` |  |
| `spec.corsRule` | `AzureCosmosdbAccountCorsRule` |  |  |  |
| `spec.corsRule.allowedOrigins` | `[]string` | yes |  |  |
| `spec.corsRule.allowedMethods` | `[]string` | yes |  |  |
| `spec.corsRule.allowedHeaders` | `[]string` | yes |  |  |
| `spec.corsRule.exposedHeaders` | `[]string` | yes |  |  |
| `spec.corsRule.maxAgeInSeconds` | `int32` |  |  |  |
| `spec.createMode` | `enum` |  |  |  |
| `spec.restore` | `AzureCosmosdbAccountRestore` |  |  |  |
| `spec.restore.sourceCosmosdbAccountId` | `string` | yes |  |  |
| `spec.restore.restoreTimestampInUtc` | `string` | yes |  |  |
| `spec.restore.databases` | `[]AzureCosmosdbAccountRestoreDatabase` |  |  |  |
| `spec.restore.databases[].name` | `string` | yes |  |  |
| `spec.restore.databases[].collectionNames` | `[]string` |  |  |  |
| `spec.restore.gremlinDatabases` | `[]AzureCosmosdbAccountRestoreGremlinDatabase` |  |  |  |
| `spec.restore.gremlinDatabases[].name` | `string` | yes |  |  |
| `spec.restore.gremlinDatabases[].graphNames` | `[]string` |  |  |  |
| `spec.restore.tablesToRestore` | `[]string` |  |  |  |
| `spec.tags` | `map<string, string>` |  |  |  |

## Field Details

### spec.region

`string` · required

The Azure region where the Cosmos DB account is homed. This is where
the account's metadata lives; the write and read regions themselves
are declared in `geo_locations` (the location with failover_priority
0 should match this field).
Examples: "eastus", "westus2", "westeurope", "southeastasia".

- rule: {"required":true,"string":{"minLen":"1"}}

### spec.resourceGroup

`string | valueFrom` · required

The Azure Resource Group where the Cosmos DB account will be created.
Can be a literal string or a reference to an AzureResourceGroup output.

- references: AzureResourceGroup (`status.outputs.resource_group_name`)
- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: AzureResourceGroup, name: <that resource's name>, fieldPath: status.outputs.resource_group_name}} -- a bare string does not parse

### spec.accountName

`string` · required

The account's name -- globally unique across all of Azure because it
becomes the DNS endpoint: https://{account_name}.documents.azure.com.
3-50 lowercase letters, numbers, and hyphens. Changing the name
replaces the account (and its endpoint), so treat it as permanent.

- rule: account_name must be 3-50 lowercase letters, numbers, and hyphens
- rule: {"required":true,"string":{"minLen":"3","maxLen":"50"}}

### spec.kind

`enum`

The API the account speaks. Unspecified means GLOBAL_DOCUMENT_DB --
the SQL (NoSQL) API. Fixed at creation: an account cannot switch
APIs, because the wire protocol shapes how every byte is stored.

- rule: {"enum":{"definedOnly":true}}

Allowed values (use exactly as shown):

- `azure_cosmosdb_account_kind_unspecified` -- Not specified: GlobalDocumentDB -- the SQL (NoSQL) API.
- `GLOBAL_DOCUMENT_DB` -- The SQL (NoSQL) API: SQL-like queries over JSON documents. Also the base for Cassandra, Gremlin, and Table via capabilities. Azure wire value: "GlobalDocumentDB".
- `MONGO_DB` -- The MongoDB-compatible API: existing MongoDB drivers and tools work unchanged. Azure wire value: "MongoDB".

### spec.consistencyPolicy

`AzureCosmosdbAccountConsistencyPolicy` · required

The account's default consistency policy -- how far reads may lag
writes, applied to every database and container inside. Required by
Azure at creation (there is no server-side default block).

- rule: {"required":true}

### spec.consistencyPolicy.consistencyLevel

`enum`

The consistency level. Unspecified means SESSION -- Azure's
recommendation for most applications.

- rule: {"enum":{"definedOnly":true}}

Allowed values (use exactly as shown):

- `azure_cosmosdb_account_consistency_level_unspecified` -- Not specified: Session.
- `STRONG` -- Linearizable reads -- always the latest committed write. Highest read latency; only available with a single write region.
- `BOUNDED_STALENESS` -- Reads lag writes by at most max_staleness_prefix versions or max_interval_in_seconds time -- the strongest level available to globally distributed, multi-write accounts.
- `SESSION` -- Read-your-writes within a client session -- the default and the right choice for most applications.
- `CONSISTENT_PREFIX` -- Reads never observe out-of-order writes, with no staleness bound.
- `EVENTUAL` -- No ordering or freshness guarantees -- highest throughput, lowest latency.

### spec.consistencyPolicy.maxIntervalInSeconds

`int32` · optional (explicit presence)

For BOUNDED_STALENESS only: the maximum time reads may lag writes,
in seconds. 5 to 86400; multi-region accounts require at least 300
(enforced at the spec level). Ignored by Azure on other levels.

- default: `5`
- rule: {"int32":{"lte":86400,"gte":5}}

### spec.consistencyPolicy.maxStalenessPrefix

`int32` · optional (explicit presence)

For BOUNDED_STALENESS only: the maximum number of versions reads may
lag writes. 10 to 2147483647; multi-region accounts require at least
100000 (enforced at the spec level). Ignored by Azure on other
levels.

- default: `100`
- rule: {"int32":{"gte":10}}

### spec.geoLocations

`[]AzureCosmosdbAccountGeoLocation` · required

The regions the account replicates to. At least one is required; the
location with failover_priority 0 is the write region (and should
match `region`). Priorities must be unique -- Azure promotes the
next-lowest priority on failover. Adding and removing regions is an
in-place update, which is how production accounts grow their
read-region footprint over time.

- rule: {"repeated":{"minItems":"1"}}

### spec.geoLocations[].location

`string` · required

The Azure region name, e.g. "eastus", "westeurope".

- rule: {"required":true,"string":{"minLen":"1"}}

### spec.geoLocations[].failoverPriority

`int32`

The failover priority: 0 is the write region; higher numbers are
promoted in ascending order when regions fail. Unique per account.

- rule: {"int32":{"gte":0}}

### spec.geoLocations[].zoneRedundant

`bool` · optional (explicit presence)

Replicate this region's data across availability zones for
within-region resilience. Not every region has zones.

- default: `false`

### spec.capabilities

`[]enum`

Capabilities customize what the account can do -- serverless billing,
extra APIs on a GLOBAL_DOCUMENT_DB account (Cassandra/Gremlin/Table),
MongoDB feature switches, and SQL-API search features. Most
capability CHANGES recreate the account (Azure allows only a small
set to be added or removed in place -- notably
ENABLE_MONGO_RETRYABLE_WRITES and DISABLE_RATE_LIMITING_RESPONSES),
so settle capabilities before going to production.

MONGO_DB accounts declare ENABLE_MONGO explicitly -- the capability
is part of the account's real configuration, never injected silently.

- rule: {"repeated":{"unique":true,"items":{"enum":{"definedOnly":true}}}}

Allowed values (use exactly as shown):

- `azure_cosmosdb_account_capability_unspecified`
- `ENABLE_SERVERLESS` -- Serverless billing: pay per request instead of provisioned RU/s. Databases and containers must not declare throughput. Wire value: "EnableServerless".
- `ENABLE_CASSANDRA` -- The Cassandra API on a GLOBAL_DOCUMENT_DB account. Wire value: "EnableCassandra".
- `ENABLE_GREMLIN` -- The Gremlin (graph) API on a GLOBAL_DOCUMENT_DB account. Wire value: "EnableGremlin".
- `ENABLE_TABLE` -- The Table API on a GLOBAL_DOCUMENT_DB account. Wire value: "EnableTable".
- `ENABLE_AGGREGATION_PIPELINE` -- The MongoDB aggregation pipeline. Wire value: "EnableAggregationPipeline".
- `ENABLE_MONGO` -- The MongoDB API itself -- declared explicitly on MONGO_DB accounts. Wire value: "EnableMongo".
- `ENABLE_MONGO_16MB_DOCUMENT_SUPPORT` -- Raise the MongoDB document size limit to 16 MB. Wire value: "EnableMongo16MBDocumentSupport".
- `MONGO_DB_V34` -- Pin the MongoDB 3.4 wire protocol (requires ENABLE_MONGO). Wire value: "MongoDBv3.4".
- `MONGO_ENABLE_DOC_LEVEL_TTL` -- Per-document TTL for MongoDB collections. Wire value: "mongoEnableDocLevelTTL".
- `DELETE_ALL_ITEMS_BY_PARTITION_KEY` -- The DeleteAllItemsByPartitionKey operation. Addable in place. Wire value: "DeleteAllItemsByPartitionKey".
- `DISABLE_RATE_LIMITING_RESPONSES` -- Return errors instead of 429 rate-limit responses. Addable AND removable in place. Wire value: "DisableRateLimitingResponses".
- `ALLOW_SELF_SERVE_UPGRADE_TO_MONGO36` -- Allow self-serve upgrade of legacy Mongo 3.2 accounts to 3.6. Wire value: "AllowSelfServeUpgradeToMongo36".
- `ENABLE_MONGO_RETRYABLE_WRITES` -- MongoDB retryable writes. Addable AND removable in place. Wire value: "EnableMongoRetryableWrites".
- `ENABLE_MONGO_ROLE_BASED_ACCESS_CONTROL` -- MongoDB data-plane role-based access control. Wire value: "EnableMongoRoleBasedAccessControl".
- `ENABLE_UNIQUE_COMPOUND_NESTED_DOCS` -- Unique compound indexes on nested MongoDB fields. Wire value: "EnableUniqueCompoundNestedDocs".
- `ENABLE_NO_SQL_VECTOR_SEARCH` -- Vector similarity search for the SQL API. Wire value: "EnableNoSQLVectorSearch".
- `ENABLE_NO_SQL_FULL_TEXT_SEARCH` -- Full-text search for the SQL API. Wire value: "EnableNoSQLFullTextSearch".
- `ENABLE_TTL_ON_CUSTOM_PATH` -- TTL on a custom document path (MongoDB). Wire value: "EnableTtlOnCustomPath".
- `ENABLE_PARTIAL_UNIQUE_INDEX` -- Partial unique indexes (MongoDB). Wire value: "EnablePartialUniqueIndex".
- `ENABLE_FABRIC_NETWORK_ACL_BYPASS` -- Let Microsoft Fabric bypass the account's network ACLs. Wire value: "EnableFabricNetworkAclBypass".

### spec.freeTierEnabled

`bool` · optional (explicit presence)

Enable the free tier: the first 1000 RU/s and 25 GB of storage are
free. Azure allows ONE free-tier account per subscription, and the
choice is fixed at creation.

- default: `false`

### spec.automaticFailoverEnabled

`bool` · optional (explicit presence)

Enable automatic failover: when the write region goes down, Azure
promotes the next region in failover-priority order without manual
intervention. Recommended for any multi-region account.

- default: `false`

### spec.multipleWriteLocationsEnabled

`bool` · optional (explicit presence)

Enable multi-region writes (active-active): every region in
`geo_locations` accepts writes, giving the lowest write latency
everywhere at the cost of conflict resolution (configured per SQL
container) and multi-master billing.

- default: `false`

### spec.publicNetworkAccessEnabled

`bool` · optional (explicit presence)

Whether the account answers on its public endpoint. When false the
account is reachable only through private endpoints
(AzurePrivateEndpoint) -- the locked-down posture for regulated
workloads.

- default: `true`

### spec.isVirtualNetworkFilterEnabled

`bool` · optional (explicit presence)

Enable virtual-network filtering: only traffic from the subnets in
`virtual_network_rules` (plus any allowed IPs) reaches the account.

- default: `false`

### spec.virtualNetworkRules

`[]AzureCosmosdbAccountVirtualNetworkRule`

The subnets allowed to reach the account when
`is_virtual_network_filter_enabled` is true. Each subnet needs the
"Microsoft.AzureCosmosDB" service endpoint enabled.

### spec.virtualNetworkRules[].subnetId

`string | valueFrom` · required

The subnet, by ARM ID. The subnet must have the
"Microsoft.AzureCosmosDB" service endpoint enabled (see
`ignore_missing_vnet_service_endpoint` to relax the ordering).

- references: AzureSubnet (`status.outputs.subnet_id`)
- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: AzureSubnet, name: <that resource's name>, fieldPath: status.outputs.subnet_id}} -- a bare string does not parse

### spec.virtualNetworkRules[].ignoreMissingVnetServiceEndpoint

`bool` · optional (explicit presence)

Accept the rule even if the subnet does not yet have the
Microsoft.AzureCosmosDB service endpoint -- useful when the endpoint
is being rolled out by a separate deployment; traffic flows once the
endpoint exists.

- default: `false`

### spec.ipRangeFilter

`[]string`

IP-based firewall: CIDR ranges or single IPv4 addresses allowed to
reach the account, applied alongside the virtual-network rules.

To keep the Azure portal's data explorer working on a firewalled
account, include Azure's portal addresses: "104.42.195.92",
"40.76.54.131", "52.176.6.30", "52.169.50.45", "52.187.184.26".
"0.0.0.0" admits traffic from Azure datacenter IPs (including other
customers' resources) -- use deliberately.

- rule: {"repeated":{"items":{"cel":[{"id":"cosmosdb_ip_range_entry","message":"ip_range_filter entries must be a valid IPv4 address or CIDR range (each octet 0-255, prefix /0 to /32)","expression":"this.matches('^((25[0-5]|2[0-4][0-9]|1[0-9]{2}|[1-9]?[0-9])\\\\.){3}(25[0-5]|2[0-4][0-9]|1[0-9]{2}|[1-9]?[0-9])(/(3[0-2]|[12]?[0-9]))?$')"}]}}}

### spec.backup

`AzureCosmosdbAccountBackup`

Backup configuration. Omitted means Azure's default: PERIODIC backups
every 4 hours retained 8 hours on geo-redundant storage. CONTINUOUS
enables point-in-time restore (and is required for `create_mode`
RESTORE) -- switching PERIODIC -> CONTINUOUS is an in-place upgrade,
but going back recreates the account.

- rule: interval_in_minutes, retention_in_hours, and storage_redundancy apply only to PERIODIC backups
- rule: tier applies only to CONTINUOUS backups

### spec.backup.type

`enum` · required

PERIODIC takes snapshots on an interval; CONTINUOUS keeps a rolling
point-in-time restore window (and is what `create_mode` RESTORE
restores from). PERIODIC -> CONTINUOUS upgrades in place; CONTINUOUS
-> PERIODIC recreates the account.

- rule: {"required":true,"enum":{"definedOnly":true}}

Allowed values (use exactly as shown):

- `azure_cosmosdb_account_backup_type_unspecified`
- `PERIODIC` -- Snapshot backups on a schedule. Wire value: "Periodic".
- `CONTINUOUS` -- A rolling point-in-time restore window. Wire value: "Continuous".

### spec.backup.tier

`enum`

For CONTINUOUS only: the restore-window tier. Unset lets Azure pick
its default (30 days).

- rule: {"enum":{"definedOnly":true}}

Allowed values (use exactly as shown):

- `azure_cosmosdb_account_continuous_tier_unspecified` -- Not specified: Azure's default (30 days).
- `CONTINUOUS_7_DAYS` -- A 7-day restore window -- cheaper. Wire value: "Continuous7Days".
- `CONTINUOUS_30_DAYS` -- A 30-day restore window. Wire value: "Continuous30Days".

### spec.backup.intervalInMinutes

`int32` · optional (explicit presence)

For PERIODIC only: minutes between backups, 60-1440. Azure defaults
to 240 (4 hours).

- rule: {"int32":{"lte":1440,"gte":60}}

### spec.backup.retentionInHours

`int32` · optional (explicit presence)

For PERIODIC only: hours each backup is retained, 8-720. Azure
defaults to 8. Two backups are always retained free; longer
retention bills per copy.

- rule: {"int32":{"lte":720,"gte":8}}

### spec.backup.storageRedundancy

`enum`

For PERIODIC only: where backup copies live -- GEO (paired region,
the default), LOCAL, or ZONE.

- rule: {"enum":{"definedOnly":true}}

Allowed values (use exactly as shown):

- `azure_cosmosdb_account_backup_storage_redundancy_unspecified` -- Not specified: Geo (Azure's default).
- `GEO` -- Replicated to the paired region. Wire value: "Geo".
- `LOCAL` -- Kept within the region. Wire value: "Local".
- `ZONE` -- Replicated across the region's availability zones. Wire value: "Zone".

### spec.mongoServerVersion

`enum`

The MongoDB wire-protocol version the account speaks. Only meaningful
on MONGO_DB accounts; Azure picks its current default when unset.
Applications' MongoDB drivers must be compatible with this version.

- rule: {"enum":{"definedOnly":true}}

Allowed values (use exactly as shown):

- `azure_cosmosdb_account_mongo_server_version_unspecified` -- Not specified: Azure's current default.
- `MONGO_3_2` -- Wire value: "3.2".
- `MONGO_3_6` -- Wire value: "3.6".
- `MONGO_4_0` -- Wire value: "4.0".
- `MONGO_4_2` -- Wire value: "4.2".
- `MONGO_5_0` -- Wire value: "5.0".
- `MONGO_6_0` -- Wire value: "6.0".
- `MONGO_7_0` -- Wire value: "7.0".

### spec.identity

`AzureCosmosdbAccountIdentity`

The account's managed identity, used to access other Azure services
-- most importantly to unwrap the customer-managed key when
`key_vault_key_id` is set. Leave unset for accounts that need no
identity.

- rule: identity_ids is required for USER_ASSIGNED and SYSTEM_AND_USER_ASSIGNED and must be empty for SYSTEM_ASSIGNED

### spec.identity.type

`enum` · required

Identity flavor. SYSTEM_ASSIGNED is created and rotated by Azure with
the account; USER_ASSIGNED brings identities you manage (what CMK
needs, because the key must be unwrappable BEFORE the account's own
identity exists); SYSTEM_AND_USER_ASSIGNED carries both.

- rule: {"required":true}

Allowed values (use exactly as shown):

- `azure_cosmosdb_account_identity_type_unspecified`
- `SYSTEM_ASSIGNED` -- Azure creates and rotates the identity with the account. Wire value: "SystemAssigned".
- `USER_ASSIGNED` -- Identities you create and manage. Wire value: "UserAssigned".
- `SYSTEM_AND_USER_ASSIGNED` -- Both. Wire value: "SystemAssigned, UserAssigned".

### spec.identity.identityIds

`[]string | valueFrom`

For USER_ASSIGNED / SYSTEM_AND_USER_ASSIGNED: the user-assigned
identities attached to the account, by ARM ID. Reference
AzureUserAssignedIdentity resources so grants (Key Vault crypto
access for CMK) compose before the account is created.

- references: AzureUserAssignedIdentity (`status.outputs.identity_id`)
- rule: write as {value: <literal>} or {valueFrom: {kind: AzureUserAssignedIdentity, name: <that resource's name>, fieldPath: status.outputs.identity_id}} -- a bare string does not parse

### spec.defaultIdentity

`AzureCosmosdbAccountDefaultIdentity`

Which identity the account uses BY DEFAULT when it reaches into other
Azure services (e.g. to unwrap the CMK). Unset means Azure's
first-party service identity. Set USER_ASSIGNED (with the identity
reference) to make CMK unwrapping ride an identity that exists -- and
can be granted Key Vault access -- BEFORE the account is created,
which is what composed CMK deployments need.

- rule: user_assigned_identity_id is required for USER_ASSIGNED and must be unset otherwise

### spec.defaultIdentity.type

`enum` · required

FIRST_PARTY is Azure's own service identity (the historical
default); SYSTEM_ASSIGNED uses the account's system identity;
USER_ASSIGNED uses the referenced identity -- the composable choice
for CMK, because the identity (and its Key Vault grants) exist
before the account does.

- rule: {"required":true}

Allowed values (use exactly as shown):

- `azure_cosmosdb_account_default_identity_type_unspecified`
- `FIRST_PARTY` -- Azure's first-party service identity -- the historical default. Wire value: "FirstPartyIdentity".
- `SYSTEM_ASSIGNED_DEFAULT` -- The account's system-assigned identity. Wire value: "SystemAssignedIdentity".
- `USER_ASSIGNED_DEFAULT` -- A user-assigned identity (set user_assigned_identity_id). Wire value: "UserAssignedIdentity=<identity ARM id>".

### spec.defaultIdentity.userAssignedIdentityId

`string | valueFrom`

For USER_ASSIGNED: the identity the account acts as, by ARM ID.
Must also be attached in `identity.identity_ids`.

- references: AzureUserAssignedIdentity (`status.outputs.identity_id`)
- rule: write as {value: <literal>} or {valueFrom: {kind: AzureUserAssignedIdentity, name: <that resource's name>, fieldPath: status.outputs.identity_id}} -- a bare string does not parse

### spec.keyVaultKeyId

`string | valueFrom`

Customer-managed-key encryption: the account's data is encrypted with
a key you own in Azure Key Vault instead of Microsoft's platform key.
Takes the key's VERSIONLESS Key Vault identifier
(https://{vault}.vault.azure.net/keys/{name}) so rotation propagates
automatically; defaults to referencing an AzureKeyVaultKey's
versionless_id output. The vault must have purge protection enabled,
and the unwrapping identity (see `default_identity`) needs
get/wrapKey/unwrapKey on the key. Fixed at creation.

- references: AzureKeyVaultKey (`status.outputs.versionless_id`)
- rule: write as {value: <literal>} or {valueFrom: {kind: AzureKeyVaultKey, name: <that resource's name>, fieldPath: status.outputs.versionless_id}} -- a bare string does not parse

### spec.analyticalStorageEnabled

`bool` · optional (explicit presence)

Enable the analytical store: a column-oriented copy of the data for
near-real-time analytics (Synapse Link) without touching transactional
RU budgets. Enabling is an in-place update; DISABLING recreates the
account -- so treat "on" as permanent.

- default: `false`

### spec.analyticalStorage

`AzureCosmosdbAccountAnalyticalStorage`

Analytical-store schema shape, meaningful when
`analytical_storage_enabled` is true. Unset lets Azure pick the
default for the account's API (WELL_DEFINED for SQL, FULL_FIDELITY
for MongoDB).

### spec.analyticalStorage.schemaType

`enum` · required

The analytical-store schema shape. WELL_DEFINED (the SQL-API
default) infers strict column types from the first document seen;
FULL_FIDELITY (the MongoDB default) keeps every type variant a
property has carried. Fixed once analytical storage holds data.

- rule: {"required":true,"enum":{"definedOnly":true}}

Allowed values (use exactly as shown):

- `azure_cosmosdb_account_analytical_storage_schema_type_unspecified`
- `WELL_DEFINED` -- Strict column types inferred from the first occurrence. Wire value: "WellDefined".
- `FULL_FIDELITY` -- Every type variant preserved. Wire value: "FullFidelity".

### spec.capacity

`AzureCosmosdbAccountCapacity`

Account-wide throughput guardrail.

### spec.capacity.totalThroughputLimit

`int32`

The maximum TOTAL RU/s that can be provisioned across all databases
and containers in the account -- the guardrail against runaway
provisioning cost. -1 means no limit.

- rule: {"int32":{"gte":-1}}

### spec.accessKeyMetadataWritesEnabled

`bool` · optional (explicit presence)

Whether account keys can WRITE metadata (create/change databases,
containers, throughput) through the data plane. Disabling restricts
metadata writes to ARM (Entra-authenticated) callers -- pair with
`local_authentication_enabled: false` for a fully Entra-governed
account.

- default: `true`

### spec.localAuthenticationEnabled

`bool` · optional (explicit presence)

Whether key- and connection-string-based (local) authentication works
at all. Disable to force every data-plane caller through Entra ID and
Cosmos DB's data-plane RBAC -- the keyless posture. The account keys
in the stack outputs stop authenticating when this is false.

- default: `true`

### spec.minimalTlsVersion

`enum`

The minimum TLS version the account's endpoints accept. Unset means
TLS 1.2 -- Azure's own default for all accounts since April 2023 and
the only floor Azure still provisions (the legacy 1.0/1.1 floors are
retired; see the enum).

- rule: {"enum":{"definedOnly":true}}

Allowed values (use exactly as shown):

- `azure_cosmosdb_account_minimal_tls_version_unspecified` -- Not specified: TLS 1.2 -- Azure's default since April 2023.
- `TLS_1_2` -- Accept TLS 1.2 and above -- the default and the only floor Azure still accepts. Azure wire value: "Tls12".

### spec.networkAclBypassForAzureServices

`bool` · optional (explicit presence)

Let trusted Azure services (e.g. Azure Synapse, Azure Data Factory)
bypass the account's network rules.

- default: `false`

### spec.networkAclBypassIds

`[]string`

Specific resource IDs allowed to bypass the network rules (e.g. a
Synapse workspace's ARM ID). Plain ARM IDs -- the bypass list admits
many unrelated resource types, so there is no single kind to
reference.

### spec.burstCapacityEnabled

`bool` · optional (explicit presence)

Enable burst capacity: idle provisioned throughput accumulates and
absorbs short spikes beyond the provisioned RU/s instead of
rate-limiting them.

- default: `false`

### spec.partitionMergeEnabled

`bool` · optional (explicit presence)

Enable partition merge: Azure consolidates fragmented physical
partitions after throughput scale-downs, recovering per-partition
throughput headroom on long-lived containers.

- default: `false`

### spec.corsRule

`AzureCosmosdbAccountCorsRule`

CORS configuration for browser-based access to the account's data
plane. Azure accepts one rule per account.

### spec.corsRule.allowedOrigins

`[]string` · required

The origins allowed to make cross-origin requests, e.g.
"https://app.example.com", or "*" for any origin.

- rule: {"repeated":{"minItems":"1","maxItems":"64"}}

### spec.corsRule.allowedMethods

`[]string` · required

The HTTP methods the rule admits.

- rule: {"repeated":{"minItems":"1","maxItems":"64","items":{"cel":[{"id":"cosmosdb_cors_method_valid","message":"allowed_methods entries must be one of: DELETE, GET, HEAD, MERGE, POST, OPTIONS, PUT, PATCH","expression":"this in ['DELETE', 'GET', 'HEAD', 'MERGE', 'POST', 'OPTIONS', 'PUT', 'PATCH']"}]}}}

### spec.corsRule.allowedHeaders

`[]string` · required

The request headers the browser may send, e.g. "x-ms-meta-*", or
"*" for all.

- rule: {"repeated":{"minItems":"1","maxItems":"64"}}

### spec.corsRule.exposedHeaders

`[]string` · required

The response headers exposed to the browser, e.g. "x-ms-meta-*", or
"*" for all.

- rule: {"repeated":{"minItems":"1","maxItems":"64"}}

### spec.corsRule.maxAgeInSeconds

`int32` · optional (explicit presence)

How long (seconds) the browser may cache the preflight response.

- rule: {"int32":{"lte":2147483647,"gte":1}}

### spec.createMode

`enum`

How the account is being created. Unspecified means a fresh, empty
account. RESTORE creates the account FROM a continuous-backup restore
point of another account -- `restore` must be set and `backup.type`
must be CONTINUOUS. Fixed at creation.

- rule: {"enum":{"definedOnly":true}}

Allowed values (use exactly as shown):

- `azure_cosmosdb_account_create_mode_unspecified` -- Not specified: a fresh, empty account. Wire value: "Default".
- `DEFAULT` -- A fresh, empty account, declared explicitly. Wire value: "Default".
- `RESTORE` -- The account is created FROM a continuous-backup restore point of another account (set `restore`). Wire value: "Restore".

### spec.restore

`AzureCosmosdbAccountRestore`

The restore source and scope when `create_mode` is RESTORE. Every
field is fixed at creation -- a restore happens exactly once, into a
new account.

### spec.restore.sourceCosmosdbAccountId

`string` · required

The RESTORABLE database account being restored from -- the special
Microsoft.DocumentDB/locations/{location}/restorableDatabaseAccounts/
{instanceId} ARM ID Azure assigns to every continuous-backup
account (NOT the plain account ID; list them with `az cosmosdb
restorable-database-account list`). A plain ARM ID because the
restorable-accounts catalog is Azure's own registry, not a
deployable kind.

- rule: {"required":true,"string":{"minLen":"1"}}

### spec.restore.restoreTimestampInUtc

`string` · required

The point in time to restore to, RFC 3339 UTC (e.g.
"2026-07-01T00:00:00Z"). Must fall inside the source account's
continuous-backup window.

- rule: {"required":true,"string":{"minLen":"1"}}

### spec.restore.databases

`[]AzureCosmosdbAccountRestoreDatabase`

Restore only these SQL databases (optionally only some of their
containers). Empty restores everything.

### spec.restore.databases[].name

`string` · required

The database name in the restore source.

- rule: {"required":true,"string":{"minLen":"1"}}

### spec.restore.databases[].collectionNames

`[]string`

Only these containers; empty restores the whole database.

### spec.restore.gremlinDatabases

`[]AzureCosmosdbAccountRestoreGremlinDatabase`

Restore only these Gremlin databases (optionally only some graphs).

### spec.restore.gremlinDatabases[].name

`string` · required

The database name in the restore source.

- rule: {"required":true,"string":{"minLen":"1"}}

### spec.restore.gremlinDatabases[].graphNames

`[]string`

Only these graphs; empty restores the whole database.

### spec.restore.tablesToRestore

`[]string`

Restore only these Table-API tables.

### spec.tags

`map<string, string>`

User-defined tags merged over the platform's identity tags (user
values win) on the account.

## Validation Rules

- `cosmosdb_geo_one_write_region`: exactly one geo_location must have failover_priority 0 (the write region)
- `cosmosdb_geo_unique_priorities`: geo_locations failover priorities must be unique
- `cosmosdb_geo_unique_locations`: geo_locations locations must be unique
- `cosmosdb_bounded_staleness_multi_region_floor`: multi-region BoundedStaleness requires max_staleness_prefix >= 100000 and max_interval_in_seconds >= 300 (both set explicitly)
- `cosmosdb_mongo_capabilities_require_mongo_kind`: MongoDB-only capabilities (ENABLE_MONGO and the ENABLE_MONGO_* / TTL / partial-index switches) require kind MONGO_DB
- `cosmosdb_sql_capabilities_require_sql_kind`: SQL-API capabilities (ENABLE_CASSANDRA/GREMLIN/TABLE, vector/full-text search, fabric bypass) require kind GLOBAL_DOCUMENT_DB (unspecified)
- `cosmosdb_mongo_v34_requires_enable_mongo`: capability MONGO_DB_V34 requires ENABLE_MONGO
- `cosmosdb_create_mode_requires_continuous`: create_mode can only be set on accounts with CONTINUOUS backup
- `cosmosdb_restore_mode_pairing`: the restore block is required with create_mode RESTORE and invalid otherwise

## Outputs

Reference an output from another manifest as `valueFrom: {kind: AzureCosmosdbAccount, name: <resource-name>, fieldPath: status.outputs.<output>}`. A sensitive output is a secret the resource generates: on Planton it is kept in the organization's secret store and the output holds a `$secret/` reference, so feed it only to a sensitive field.

| Output | Type | Description |
|---|---|---|
| `status.outputs.cosmosdb_account_id` | `string` | The Azure Resource Manager ID of the account. Format: /subscriptions/{sub}/resourceGroups/{rg}/providers/Microsoft.DocumentDB/databaseAccounts/{name} |
| `status.outputs.cosmosdb_account_name` | `string` | The account's name -- the globally unique DNS label. |
| `status.outputs.endpoint` | `string` | The document endpoint SDKs connect to. Format: https://{name}.documents.azure.com:443/ |
| `status.outputs.read_endpoints` | `[]string` | Per-region read endpoints, ordered by the account's failover priorities -- what latency-sensitive readers pin to. |
| `status.outputs.write_endpoints` | `[]string` | Per-region write endpoints. One entry for single-write accounts; one per region when multiple_write_locations_enabled is true. |
| `status.outputs.primary_key` | `string` (sensitive) | The primary read-write account key (secret-bearing). |
| `status.outputs.secondary_key` | `string` (sensitive) | The secondary read-write account key (secret-bearing) -- the rotation partner: applications move to the secondary, the primary is regenerated, and back. |
| `status.outputs.primary_readonly_key` | `string` (sensitive) | The primary read-only account key (secret-bearing) -- for consumers that must never write. |
| `status.outputs.secondary_readonly_key` | `string` (sensitive) | The secondary read-only account key (secret-bearing). |
| `status.outputs.primary_sql_connection_string` | `string` (sensitive) | Ready-made SQL-API connection strings (secret-bearing), populated for every account kind. Format: AccountEndpoint={endpoint};AccountKey={key}; |
| `status.outputs.secondary_sql_connection_string` | `string` (sensitive) | The secondary SQL-API connection string (secret-bearing). |
| `status.outputs.primary_readonly_sql_connection_string` | `string` (sensitive) | The read-only primary SQL-API connection string (secret-bearing). |
| `status.outputs.secondary_readonly_sql_connection_string` | `string` (sensitive) | The read-only secondary SQL-API connection string (secret-bearing). |
| `status.outputs.primary_mongodb_connection_string` | `string` (sensitive) | Ready-made MongoDB connection strings (secret-bearing), meaningful on MONGO_DB accounts. Format: mongodb://{name}:{key}@{name}.mongo.cosmos.azure.com:10255/?ssl=true... |
| `status.outputs.secondary_mongodb_connection_string` | `string` (sensitive) | The secondary MongoDB connection string (secret-bearing). |
| `status.outputs.primary_readonly_mongodb_connection_string` | `string` (sensitive) | The read-only primary MongoDB connection string (secret-bearing). |
| `status.outputs.secondary_readonly_mongodb_connection_string` | `string` (sensitive) | The read-only secondary MongoDB connection string (secret-bearing). |
| `status.outputs.identity_principal_id` | `string` | The principal ID of the account's system-assigned identity, when `identity` requests one -- the subject for role assignments the account needs against other services. |

## References

Fields that can point at another resource's outputs:

| Field | Kind | Output |
|---|---|---|
| `spec.resourceGroup` | AzureResourceGroup | `status.outputs.resource_group_name` |
| `spec.virtualNetworkRules[].subnetId` | AzureSubnet | `status.outputs.subnet_id` |
| `spec.identity.identityIds` | AzureUserAssignedIdentity | `status.outputs.identity_id` |
| `spec.defaultIdentity.userAssignedIdentityId` | AzureUserAssignedIdentity | `status.outputs.identity_id` |
| `spec.keyVaultKeyId` | AzureKeyVaultKey | `status.outputs.versionless_id` |

## Referenced By

Fields on other kinds that can point at this resource:

| Kind | Field | Reads |
|---|---|---|
| AzureCosmosdbMongoDatabase | `spec.cosmosdbAccountId` | `status.outputs.cosmosdb_account_id` |
| AzureCosmosdbSqlDatabase | `spec.cosmosdbAccountId` | `status.outputs.cosmosdb_account_id` |
| AzureCosmosdbSqlRoleAssignment | `spec.cosmosdbAccountId` | `status.outputs.cosmosdb_account_id` |
| AzureCosmosdbSqlRoleAssignment | `spec.scope` | `status.outputs.cosmosdb_account_id` |
| AzureCosmosdbSqlRoleDefinition | `spec.cosmosdbAccountId` | `status.outputs.cosmosdb_account_id` |
| AzureCosmosdbSqlRoleDefinition | `spec.assignableScopes` | `status.outputs.cosmosdb_account_id` |

## See Also

- [Overview](../README.md)
