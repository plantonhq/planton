# AzureMongoCluster

> Generated from the protobuf schema by `make generate-reference` -- do not
> edit by hand. To change a fact on this page, change the proto field comment
> or validation rule it is derived from, then regenerate.

**apiVersion**: `azure.planton.dev/v1alpha1`

**Guide**: [GUIDE.md](../GUIDE.md) -- authored operational judgment for this component: conventions, trade-offs, and what pairs well with it.

**AzureMongoClusterSpec** defines an Azure Cosmos DB for MongoDB
vCore cluster -- Azure's modern managed MongoDB offering: a real
MongoDB engine (wire-protocol compatible with community drivers) on
dedicated vCore-based compute, with vertical tiers from a free
sandbox to M200, optional sharding, zone-redundant high
availability, and point-in-time restore.

A cluster is created in one of three modes (create_mode): "Default"
builds a fresh cluster from the sizing fields; "GeoReplica" builds a
cross-region read replica of an existing cluster; and
"PointInTimeRestore" clones an existing cluster from a moment in its
backup history. Azure never returns the mode on reads, so CHANGING
it after create always replaces the cluster.

Update-surface honesty, recorded where it bites: adding or removing
the identity block replaces the cluster (Azure rejects in-place
identity transitions on this resource); disabling the Data API after
it is enabled replaces the cluster; and upgrades away from the
"Free" or "M25" tiers stage a tier-only change first (the provider
performs the two-step itself).

## Example

```yaml
# Deep-shape example for docs and offline validation: a Default-mode
# production-ish cluster with zone-redundant HA, Entra + native auth,
# and two firewall rules. The administrator password is a literal here
# only so the manifest validates standalone -- reference a secret
# store in real use.
apiVersion: azure.planton.dev/v1alpha1
kind: AzureMongoCluster
metadata:
  name: test-mongo-cluster
  id: test-mongo-cluster
  org: test-org
  env: test
spec:
  resourceGroup:
    value: test-rg
  name: acme-orders-db
  region: eastus
  createMode: Default
  administratorUsername: mongoadmin
  administratorPassword:
    value: Test-Passw0rd-Change-Me
  version: "8.0"
  computeTier: M30
  storageSizeInGb: 128
  storageType: PremiumSSD
  shardCount: 1
  highAvailabilityMode: ZoneRedundantPreferred
  authenticationMethods:
    - NativeAuth
    - MicrosoftEntraID
  publicNetworkAccessEnabled: true
  firewallRules:
    - name: office
      startIpAddress: 203.0.113.0
      endIpAddress: 203.0.113.255
    - name: vpn-egress
      startIpAddress: 198.51.100.7
      endIpAddress: 198.51.100.7
  tags:
    costCenter: platform
```

## Spec Fields

| Path | Type | Required | Default | References |
|---|---|---|---|---|
| `spec.resourceGroup` | `string \| valueFrom` | yes |  | AzureResourceGroup (`status.outputs.resource_group_name`) |
| `spec.name` | `string` | yes |  |  |
| `spec.region` | `string` | yes |  |  |
| `spec.createMode` | `string` |  | `Default` |  |
| `spec.administratorUsername` | `string` |  |  |  |
| `spec.administratorPassword` | `string \| valueFrom` (sensitive) |  |  |  |
| `spec.version` | `string` |  |  |  |
| `spec.computeTier` | `string` |  |  |  |
| `spec.storageSizeInGb` | `int32` |  |  |  |
| `spec.storageType` | `string` |  | `PremiumSSD` |  |
| `spec.shardCount` | `int32` |  |  |  |
| `spec.highAvailabilityMode` | `string` |  |  |  |
| `spec.authenticationMethods` | `[]string` |  |  |  |
| `spec.userAssignedIdentityIds` | `[]string \| valueFrom` |  |  | AzureUserAssignedIdentity (`status.outputs.identity_id`) |
| `spec.customerManagedKey` | `AzureMongoClusterCustomerManagedKey` |  |  |  |
| `spec.customerManagedKey.keyVaultKeyId` | `string \| valueFrom` | yes |  | AzureKeyVaultKey (`status.outputs.versionless_id`) |
| `spec.customerManagedKey.userAssignedIdentityId` | `string \| valueFrom` | yes |  | AzureUserAssignedIdentity (`status.outputs.identity_id`) |
| `spec.previewFeatures` | `[]string` |  |  |  |
| `spec.sourceServerId` | `string \| valueFrom` |  |  | AzureMongoCluster (`status.outputs.mongo_cluster_id`) |
| `spec.sourceLocation` | `string` |  |  |  |
| `spec.restore` | `AzureMongoClusterRestore` |  |  |  |
| `spec.restore.pointInTimeUtc` | `string` | yes |  |  |
| `spec.restore.sourceId` | `string \| valueFrom` | yes |  | AzureMongoCluster (`status.outputs.mongo_cluster_id`) |
| `spec.dataApiModeEnabled` | `bool` |  |  |  |
| `spec.publicNetworkAccessEnabled` | `bool` |  | `true` |  |
| `spec.firewallRules` | `[]AzureMongoClusterFirewallRule` |  |  |  |
| `spec.firewallRules[].name` | `string` | yes |  |  |
| `spec.firewallRules[].startIpAddress` | `string` | yes |  |  |
| `spec.firewallRules[].endIpAddress` | `string` | yes |  |  |
| `spec.tags` | `map<string, string>` |  |  |  |

## Field Details

### spec.resourceGroup

`string | valueFrom` · required

The Azure Resource Group the cluster lives in. Can be a literal
string or a reference to an AzureResourceGroup output.

**ForceNew**: changing this destroys and recreates the cluster.

- references: AzureResourceGroup (`status.outputs.resource_group_name`)
- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: AzureResourceGroup, name: <that resource's name>, fieldPath: status.outputs.resource_group_name}} -- a bare string does not parse

### spec.name

`string` · required

The cluster's name -- 3-40 characters; lowercase letters, numbers,
and hyphens; starts and ends with a letter or number. The name
becomes the cluster's public hostname
({name}.mongocluster.cosmos.azure.com), so it must be globally
unique -- a taken name fails at deploy time.

**ForceNew**: changing this destroys and recreates the cluster.

- rule: Cluster names must be 3-40 characters of lowercase letters, numbers, and hyphens, starting and ending with a letter or number
- rule: {"required":true}

### spec.region

`string` · required

The Azure region the cluster is created in, e.g. "eastus".

**ForceNew**: changing this destroys and recreates the cluster.

- rule: {"required":true,"string":{"minLen":"1"}}

### spec.createMode

`string` · optional (explicit presence)

How the cluster comes into existence. "Default" (the default)
builds a fresh cluster -- administrator_username, version,
compute_tier, storage_size_in_gb, shard_count, and
high_availability_mode are all required. "GeoReplica" builds a
cross-region replica -- source_server_id and source_location are
required, sizing is inherited. "PointInTimeRestore" clones from
backup history -- the restore block is required, sizing is
inherited.

Azure never returns this property, and changing it always
REPLACES the cluster (the provider forces the replacement).

- default: `Default`
- rule: {"string":{"in":["Default","GeoReplica","PointInTimeRestore"]}}

### spec.administratorUsername

`string`

The administrator login for native MongoDB authentication.
Required for "Default" mode; replicas and restores inherit the
source's administrator. Must be set together with
administrator_password.

**ForceNew**: changing the username destroys and recreates the
cluster (the password rotates in place).

### spec.administratorPassword

`string | valueFrom` · sensitive

The administrator password. Updatable in place (a password
rotation). Pass a literal only in throwaway environments --
reference a secret store (for example an AzureKeyVaultSecret
output) everywhere else. Azure never returns the password on
reads.

- rule: write as {value: <literal>} or {valueFrom: {kind: <Kind>, name: <that resource's name>, fieldPath: status.outputs.<output>}} -- a bare string does not parse

### spec.version

`string` · optional (explicit presence)

The MongoDB server version: "5.0", "6.0", "7.0", or "8.0".
Required for "Default" mode. Upgradable in place (an online major
version upgrade -- Azure performs it rolling).

- rule: {"string":{"in":["5.0","6.0","7.0","8.0"]}}

### spec.computeTier

`string` · optional (explicit presence)

The compute tier per shard: "Free" (one per subscription,
sandbox), "M10"/"M20"/"M25" (burstable dev tiers), "M30" and up
(general purpose, dedicated vCores). Required for "Default" mode;
updatable in place. "Free" and "M25" clusters cannot use
zone-redundant high availability and cannot shard beyond one
shard, and "Free" clusters cannot use MicrosoftEntraID
authentication (ARM rejects the create with 400 "Microsoft Entra
ID authentication is not supported for 'Free' cluster tier").

- rule: {"string":{"in":["Free","M10","M20","M25","M30","M40","M50","M60","M80","M200"]}}

### spec.storageSizeInGb

`int32` · optional (explicit presence)

Storage per shard in GiB, 32-32768. Required for "Default" mode;
grows in place (shrinking is not supported by the service).

- rule: {"int32":{"lte":32768,"gte":32}}

### spec.storageType

`string` · optional (explicit presence)

The storage performance class: "PremiumSSD" (the default) or
"PremiumSSDv2". Only sent alongside storage_size_in_gb -- when the
size is unset (replica/restore modes), Azure decides.

**ForceNew**: changing the storage type destroys and recreates
the cluster.

- default: `PremiumSSD`
- rule: {"string":{"in":["PremiumSSD","PremiumSSDv2"]}}

### spec.shardCount

`int32` · optional (explicit presence)

How many shards the cluster's data is split across, 1 or more.
Required for "Default" mode. "Free" and "M25" tiers allow exactly
one shard.

**ForceNew**: changing the shard count destroys and recreates the
cluster (re-sharding is not an in-place operation).

- rule: {"int32":{"gte":1}}

### spec.highAvailabilityMode

`string` · optional (explicit presence)

High availability: "Disabled" (single replica per shard) or
"ZoneRedundantPreferred" (a standby replica in another availability
zone where the region supports it). Required for "Default" mode;
updatable in place. Azure's "SameZone" mode is not yet supported
by the service on this resource. Not available on the "Free" and
"M25" tiers.

- rule: {"string":{"in":["Disabled","ZoneRedundantPreferred"]}}

### spec.authenticationMethods

`[]string`

Which authentication methods the cluster accepts: "NativeAuth"
(MongoDB username/password) and/or "MicrosoftEntraID" (Entra
principals via AzureMongoClusterUser grants). Azure defaults an
unset list to ["NativeAuth"] -- the engines send the list only
when it is set, mirroring that service default. Include
"MicrosoftEntraID" here before creating AzureMongoClusterUser
grants against the cluster. Not available on the "Free" tier:
ARM rejects a Free cluster listing "MicrosoftEntraID" at create
(400 "Microsoft Entra ID authentication is not supported for
'Free' cluster tier" -- a server-side contract no provider
schema carries); use M10 or higher when grants are needed.

- rule: {"repeated":{"unique":true,"items":{"string":{"in":["NativeAuth","MicrosoftEntraID"]}}}}

### spec.userAssignedIdentityIds

`[]string | valueFrom`

User-assigned identities attached to the cluster, by ARM ID (the
service supports ONLY user-assigned identities on this resource).
Required when customer_managed_key is set -- the key-unwrap
identity must be attached here.

**ForceNew**: ADDING the first identity or REMOVING the last one
replaces the cluster (Azure rejects the in-place transition);
changing the set of identities updates in place.

- references: AzureUserAssignedIdentity (`status.outputs.identity_id`)
- rule: write as {value: <literal>} or {valueFrom: {kind: AzureUserAssignedIdentity, name: <that resource's name>, fieldPath: status.outputs.identity_id}} -- a bare string does not parse

### spec.customerManagedKey

`AzureMongoClusterCustomerManagedKey`

Customer-managed encryption at rest: a Key Vault key plus the
user-assigned identity Azure uses to unwrap it. The identity must
also appear in user_assigned_identity_ids, and it needs
unwrap/wrap permissions on the vault BEFORE create -- Azure
validates both at deploy time.

**ForceNew**: the block cannot be added, removed, or changed after
create.

### spec.customerManagedKey.keyVaultKeyId

`string | valueFrom` · required

The Key Vault key that encrypts the cluster's data, by VERSIONLESS
key identifier (https://{vault}.vault.azure.net/keys/{name} --
Azure follows the key's rotation automatically; the provider
rejects versioned identifiers). Reference an AzureKeyVaultKey
output or pass a literal identifier.

- references: AzureKeyVaultKey (`status.outputs.versionless_id`)
- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: AzureKeyVaultKey, name: <that resource's name>, fieldPath: status.outputs.versionless_id}} -- a bare string does not parse

### spec.customerManagedKey.userAssignedIdentityId

`string | valueFrom` · required

The user-assigned identity Azure authenticates as when unwrapping
the key. It must be attached to the cluster
(user_assigned_identity_ids) and hold unwrap/wrap permissions on
the vault before the cluster is created.

- references: AzureUserAssignedIdentity (`status.outputs.identity_id`)
- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: AzureUserAssignedIdentity, name: <that resource's name>, fieldPath: status.outputs.identity_id}} -- a bare string does not parse

### spec.previewFeatures

`[]string`

Preview features enabled at create, e.g. "GeoReplicas" (required
on a source cluster before a GeoReplica can be created from it).
Azure owns the live catalog of feature names -- values are not
validated beyond being non-empty and unique.

**ForceNew**: changing the list destroys and recreates the
cluster.

- rule: {"repeated":{"unique":true,"items":{"string":{"minLen":"1"}}}}

### spec.sourceServerId

`string | valueFrom`

For "GeoReplica" mode: the source cluster this cluster replicates,
by ARM ID. Reference an AzureMongoCluster output or pass a literal
ID. The source must have the "GeoReplicas" preview feature
enabled.

**ForceNew**: changing this destroys and recreates the replica.

- references: AzureMongoCluster (`status.outputs.mongo_cluster_id`)
- rule: write as {value: <literal>} or {valueFrom: {kind: AzureMongoCluster, name: <that resource's name>, fieldPath: status.outputs.mongo_cluster_id}} -- a bare string does not parse

### spec.sourceLocation

`string`

For "GeoReplica" mode: the source cluster's region (e.g.
"eastus"). Azure never returns this property on reads.

**ForceNew**: changing this destroys and recreates the replica.

### spec.restore

`AzureMongoClusterRestore`

For "PointInTimeRestore" mode: which cluster to clone and from
which moment in its backup history.

**ForceNew**: changing this destroys and recreates the clone.

### spec.restore.pointInTimeUtc

`string` · required

The moment in the source cluster's backup history to clone from,
as an RFC 3339 UTC timestamp (e.g. "2026-08-01T12:00:00Z").

- rule: point_in_time_utc must be an RFC 3339 timestamp like 2026-08-01T12:00:00Z
- rule: {"required":true}

### spec.restore.sourceId

`string | valueFrom` · required

The cluster to clone, by ARM ID. Reference an AzureMongoCluster
output or pass a literal ID.

- references: AzureMongoCluster (`status.outputs.mongo_cluster_id`)
- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: AzureMongoCluster, name: <that resource's name>, fieldPath: status.outputs.mongo_cluster_id}} -- a bare string does not parse

### spec.dataApiModeEnabled

`bool` · optional (explicit presence)

Whether the cluster's Data API (REST access to the data plane) is
enabled. Only settable on "Default"-mode clusters. Azure can only
ENABLE it after create (the provider stages that automatically);
turning it back OFF replaces the cluster. Default: false.

### spec.publicNetworkAccessEnabled

`bool` · optional (explicit presence)

Whether the cluster accepts connections from the public internet.
Set false to restrict access to private endpoints only. Default:
true (Azure's default) -- the platform sends the value explicitly.

- default: `true`

### spec.firewallRules

`[]AzureMongoClusterFirewallRule`

Client IP firewall rules -- IPv4 ranges allowed to connect while
public network access is enabled. Each rule is a named
start/end-address range; use the same start and end for a single
address, or 0.0.0.0 to 255.255.255.255 to allow all (including
other Azure services). Rules update in place; names are the
rules' identity, so they must be unique.

### spec.firewallRules[].name

`string` · required

The rule's name -- 1-80 characters; starts with a letter or
number; letters, numbers, dots, hyphens, and underscores. The
name is the rule's identity: renaming replaces the rule.

- rule: Firewall rule names must be 1-80 characters, start with a letter or number, and contain only letters, numbers, dots, hyphens, and underscores
- rule: {"required":true}

### spec.firewallRules[].startIpAddress

`string` · required

The first IPv4 address of the allowed range (inclusive), e.g.
"203.0.113.0". Updatable in place.

- rule: {"required":true,"string":{"ipv4":true}}

### spec.firewallRules[].endIpAddress

`string` · required

The last IPv4 address of the allowed range (inclusive), e.g.
"203.0.113.255". Use the start address again for a single-address
rule. Updatable in place.

- rule: {"required":true,"string":{"ipv4":true}}

### spec.tags

`map<string, string>`

Tags to apply to the cluster, merged over the Planton-derived
metadata tags (user values win on key conflicts).

## Validation Rules

- `mongo_cluster_default_mode_requirements`: Default-mode clusters (create_mode unset or 'Default') require administrator_username, version, compute_tier, storage_size_in_gb, shard_count, and high_availability_mode
- `mongo_cluster_geo_replica_requirements`: GeoReplica-mode clusters require source_server_id and source_location
- `mongo_cluster_restore_requirements`: PointInTimeRestore-mode clusters require the restore block
- `mongo_cluster_admin_credentials_pair`: administrator_username and administrator_password must be set together
- `mongo_cluster_source_location_pairs_with_source_server`: source_location requires source_server_id
- `mongo_cluster_burstable_tier_limits`: The Free and M25 tiers cannot use ZoneRedundantPreferred high availability and cannot have more than one shard
- `mongo_cluster_free_tier_no_entra`: The Free tier does not support MicrosoftEntraID authentication -- use M10 or higher for clusters that Entra principals sign in to
- `mongo_cluster_data_api_default_mode_only`: data_api_mode_enabled can only be set on Default-mode clusters
- `mongo_cluster_storage_type_requires_size`: storage_type requires storage_size_in_gb
- `mongo_cluster_cmk_requires_identity`: customer_managed_key requires the unwrap identity to be listed in user_assigned_identity_ids
- `mongo_cluster_firewall_rule_names_unique`: firewall_rules names must be unique -- the name is the rule's identity

## Outputs

Reference an output from another manifest as `valueFrom: {kind: AzureMongoCluster, name: <resource-name>, fieldPath: status.outputs.<output>}`. A sensitive output is a secret the resource generates: on Planton it is kept in the organization's secret store and the output holds a `$secret/` reference, so feed it only to a sensitive field.

| Output | Type | Description |
|---|---|---|
| `status.outputs.mongo_cluster_id` | `string` | The cluster's Azure Resource Manager ID -- the target an AzureMongoClusterUser's mongo_cluster_id (and a replica's source_server_id) references. |
| `status.outputs.mongo_cluster_name` | `string` | The cluster's name (also the first label of its hostname, {name}.mongocluster.cosmos.azure.com). |
| `status.outputs.connection_string` | `string` (sensitive) | The cluster's primary MongoDB connection string, with the administrator credentials substituted in (Azure returns a <user>:<password> placeholder; the engines fill it from the spec). Empty when the cluster has no native administrator. |
| `status.outputs.connection_strings` | `map<string, string>` (sensitive) | Every connection string Azure publishes for the cluster, keyed by Azure's name for it (the primary plus per-replica and per-mode variants), credentials substituted as above. |

## References

Fields that can point at another resource's outputs:

| Field | Kind | Output |
|---|---|---|
| `spec.resourceGroup` | AzureResourceGroup | `status.outputs.resource_group_name` |
| `spec.userAssignedIdentityIds` | AzureUserAssignedIdentity | `status.outputs.identity_id` |
| `spec.customerManagedKey.keyVaultKeyId` | AzureKeyVaultKey | `status.outputs.versionless_id` |
| `spec.customerManagedKey.userAssignedIdentityId` | AzureUserAssignedIdentity | `status.outputs.identity_id` |
| `spec.sourceServerId` | AzureMongoCluster | `status.outputs.mongo_cluster_id` |
| `spec.restore.sourceId` | AzureMongoCluster | `status.outputs.mongo_cluster_id` |

## Referenced By

Fields on other kinds that can point at this resource:

| Kind | Field | Reads |
|---|---|---|
| AzureMongoCluster | `spec.sourceServerId` | `status.outputs.mongo_cluster_id` |
| AzureMongoCluster | `spec.restore.sourceId` | `status.outputs.mongo_cluster_id` |
| AzureMongoClusterUser | `spec.mongoClusterId` | `status.outputs.mongo_cluster_id` |

## See Also

- [Overview](../README.md)
