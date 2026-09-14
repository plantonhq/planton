# AzurePostgresqlFlexibleServer

> Generated from the protobuf schema by `make generate-reference` -- do not
> edit by hand. To change a fact on this page, change the proto field comment
> or validation rule it is derived from, then regenerate.

**apiVersion**: `azure.planton.dev/v1alpha1`

**AzurePostgresqlFlexibleServerSpec** defines the configuration for creating
an Azure Database for PostgreSQL Flexible Server: Azure's managed PostgreSQL
with per-server compute/storage sizing, zone-redundant high availability,
Microsoft Entra (Azure AD) authentication, customer-managed-key encryption,
and point-in-time restore.

The server is the unit of management -- databases, firewall rules, server
parameters, and Entra administrators are configured on it and live with it,
so they are folded into this spec rather than modeled as standalone kinds
(none of them has an independent lifecycle or is referenced by anything
else).

**Network access** has two independent dials, matching Azure's real
contract: `public_network_access_enabled` controls the public endpoint
(with `firewall_rules` as its allowlist), and `delegated_subnet_id` +
`private_dns_zone_id` inject the server into a virtual network. Azure
requires public access OFF when the server is VNet-injected; validation
enforces the pairing up front.

**Authentication** supports PostgreSQL password auth (the default),
Microsoft Entra auth (set `authentication.active_directory_auth_enabled`
and declare `aad_administrators`), or both. When password auth is disabled,
the admin login/password pair must be omitted -- Azure rejects it.

**Lifecycle modes** (`create_mode`) cover the full server story: a fresh
server (DEFAULT), a read replica (REPLICA + `source_server_id`),
point-in-time restore and cross-region geo-restore (with
`point_in_time_restore_time_in_utc`), and reviving a dropped server. A
replica is promoted to a standalone primary by setting `replication_role`
to NONE.

## Example

```yaml
apiVersion: azure.planton.dev/v1alpha1
kind: AzurePostgresqlFlexibleServer
metadata:
  name: test-pg
spec:
  region: eastus
  resourceGroup:
    value: test-rg
  serverName: test-pg-server
  administratorLogin: pgadmin
  administratorPassword:
    value: P@ssw0rd1234!
  version: "16"
  skuName: GP_Standard_D2s_v3
  storageMb: 131072
  # Exercises the storage-tier enum mapping (P30 is valid for 128 GiB).
  storageTier: P30
  autoGrowEnabled: true
  zone: "1"
  # Exercises the HA-mode enum mapping.
  highAvailability:
    mode: ZONE_REDUNDANT
    standbyAvailabilityZone: "2"
  maintenanceWindow:
    dayOfWeek: 6
    startHour: 2
    startMinute: 30
  backupRetentionDays: 14
  # Exercises the authentication block with Entra auth alongside password
  # auth, which activates the AAD administrator sub-resource path.
  authentication:
    activeDirectoryAuthEnabled: true
    passwordAuthEnabled: true
  # Exercises the principal-type enum mapping and the administrator
  # sub-resource.
  aadAdministrators:
    - objectId:
        value: 11111111-2222-3333-4444-555555555555
      principalName: dba-team@contoso.com
      principalType: GROUP
  # Exercises the identity-type enum mapping.
  identity:
    type: SYSTEM_ASSIGNED
  databases:
    - name: myapp
    - name: legacy
      charset: SQL_ASCII
      collation: C
  firewallRules:
    - name: allow-azure-services
      startIpAddress: "0.0.0.0"
      endIpAddress: "0.0.0.0"
    - name: allow-office
      startIpAddress: "203.0.113.0"
      endIpAddress: "203.0.113.255"
  # Exercises the server-parameter configuration sub-resources.
  serverParameters:
    azure.extensions: PGCRYPTO
    max_connections: "120"
  tags:
    team: data
```

## Spec Fields

| Path | Type | Required | Default | References |
|---|---|---|---|---|
| `spec.region` | `string` | yes |  |  |
| `spec.resourceGroup` | `string \| valueFrom` | yes |  | AzureResourceGroup (`status.outputs.resource_group_name`) |
| `spec.serverName` | `string` | yes |  |  |
| `spec.createMode` | `enum` |  |  |  |
| `spec.sourceServerId` | `string \| valueFrom` |  |  | AzurePostgresqlFlexibleServer (`status.outputs.server_id`) |
| `spec.pointInTimeRestoreTimeInUtc` | `string` |  |  |  |
| `spec.replicationRole` | `enum` |  |  |  |
| `spec.administratorLogin` | `string` |  |  |  |
| `spec.administratorPassword` | `string \| valueFrom` (sensitive) |  |  |  |
| `spec.version` | `string` |  | `16` |  |
| `spec.skuName` | `string` |  |  |  |
| `spec.storageMb` | `int32` |  |  |  |
| `spec.storageTier` | `enum` |  |  |  |
| `spec.autoGrowEnabled` | `bool` |  |  |  |
| `spec.zone` | `string` |  |  |  |
| `spec.highAvailability` | `AzurePostgresqlFlexibleServerHighAvailability` |  |  |  |
| `spec.highAvailability.mode` | `enum` | yes |  |  |
| `spec.highAvailability.standbyAvailabilityZone` | `string` |  |  |  |
| `spec.maintenanceWindow` | `AzurePostgresqlFlexibleServerMaintenanceWindow` |  |  |  |
| `spec.maintenanceWindow.dayOfWeek` | `int32` |  |  |  |
| `spec.maintenanceWindow.startHour` | `int32` |  |  |  |
| `spec.maintenanceWindow.startMinute` | `int32` |  |  |  |
| `spec.backupRetentionDays` | `int32` |  | `7` |  |
| `spec.geoRedundantBackupEnabled` | `bool` |  |  |  |
| `spec.publicNetworkAccessEnabled` | `bool` |  | `true` |  |
| `spec.delegatedSubnetId` | `string \| valueFrom` |  |  | AzureSubnet (`status.outputs.subnet_id`) |
| `spec.privateDnsZoneId` | `string \| valueFrom` |  |  | AzurePrivateDnsZone (`status.outputs.zone_id`) |
| `spec.authentication` | `AzurePostgresqlFlexibleServerAuthentication` |  |  |  |
| `spec.authentication.activeDirectoryAuthEnabled` | `bool` |  |  |  |
| `spec.authentication.passwordAuthEnabled` | `bool` |  | `true` |  |
| `spec.authentication.tenantId` | `string` |  |  |  |
| `spec.aadAdministrators` | `[]AzurePostgresqlFlexibleServerAadAdministrator` |  |  |  |
| `spec.aadAdministrators[].objectId` | `string \| valueFrom` | yes |  | AzureUserAssignedIdentity (`status.outputs.principal_id`) |
| `spec.aadAdministrators[].principalName` | `string` | yes |  |  |
| `spec.aadAdministrators[].principalType` | `enum` | yes |  |  |
| `spec.identity` | `AzurePostgresqlFlexibleServerIdentity` |  |  |  |
| `spec.identity.type` | `enum` | yes |  |  |
| `spec.identity.identityIds` | `[]string \| valueFrom` |  |  | AzureUserAssignedIdentity (`status.outputs.identity_id`) |
| `spec.customerManagedKey` | `AzurePostgresqlFlexibleServerCustomerManagedKey` |  |  |  |
| `spec.customerManagedKey.keyVaultKeyId` | `string \| valueFrom` | yes |  | AzureKeyVaultKey (`status.outputs.versionless_id`) |
| `spec.customerManagedKey.primaryUserAssignedIdentityId` | `string \| valueFrom` |  |  | AzureUserAssignedIdentity (`status.outputs.identity_id`) |
| `spec.customerManagedKey.geoBackupKeyVaultKeyId` | `string \| valueFrom` |  |  | AzureKeyVaultKey (`status.outputs.versionless_id`) |
| `spec.customerManagedKey.geoBackupUserAssignedIdentityId` | `string \| valueFrom` |  |  | AzureUserAssignedIdentity (`status.outputs.identity_id`) |
| `spec.cluster` | `AzurePostgresqlFlexibleServerCluster` |  |  |  |
| `spec.cluster.size` | `int32` | yes |  |  |
| `spec.cluster.defaultDatabaseName` | `string` | yes | `postgres` |  |
| `spec.databases` | `[]AzurePostgresqlFlexibleServerDatabase` |  |  |  |
| `spec.databases[].name` | `string` | yes |  |  |
| `spec.databases[].charset` | `string` | yes | `UTF8` |  |
| `spec.databases[].collation` | `string` | yes | `en_US.utf8` |  |
| `spec.firewallRules` | `[]AzurePostgresqlFlexibleServerFirewallRule` |  |  |  |
| `spec.firewallRules[].name` | `string` | yes |  |  |
| `spec.firewallRules[].startIpAddress` | `string` | yes |  |  |
| `spec.firewallRules[].endIpAddress` | `string` | yes |  |  |
| `spec.serverParameters` | `map<string, string>` |  |  |  |
| `spec.tags` | `map<string, string>` |  |  |  |

## Field Details

### spec.region

`string` · required

The Azure region where the server will be created (e.g. "eastus",
"westeurope"). Must match the region of the delegated subnet when the
server is VNet-injected. Changing the region replaces the server.

- rule: {"required":true,"string":{"minLen":"1"}}

### spec.resourceGroup

`string | valueFrom` · required

The Azure resource group the server will be created in. Can be a
literal resource-group name or a reference to an AzureResourceGroup's
name output. Changing it replaces the server.

- references: AzureResourceGroup (`status.outputs.resource_group_name`)
- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: AzureResourceGroup, name: <that resource's name>, fieldPath: status.outputs.resource_group_name}} -- a bare string does not parse

### spec.serverName

`string` · required

The server's name: 3-63 lowercase letters, digits, and hyphens,
starting and ending with a letter or digit -- and GLOBALLY unique
across Azure, because it becomes the server's DNS name
({name}.postgres.database.azure.com, the fqdn output). Changing the
name replaces the server.

- rule: {"required":true,"string":{"minLen":"3","maxLen":"63","pattern":"^[a-z0-9]([a-z0-9-]*[a-z0-9])?$"}}

### spec.createMode

`enum`

How the server comes into existence. Unspecified means DEFAULT (a
fresh, empty server). The restore/replica modes consume
source_server_id (and, for the restore modes, the restore timestamp)
and are fixed at creation. Azure's internal "Update" mode is not a
declarative state and is deliberately not modeled.

Allowed values (use exactly as shown):

- `azure_postgresql_flexible_server_create_mode_unspecified` -- Not specified: DEFAULT (a fresh, empty server).
- `DEFAULT` -- A fresh, empty server (requires sku_name and, with password auth, the admin credentials).
- `POINT_IN_TIME_RESTORE` -- Restore the source server to a point in time, in the same region (requires source_server_id + point_in_time_restore_time_in_utc).
- `REPLICA` -- A read replica of the source server, asynchronously replicated (requires source_server_id; SKU/storage left unset inherit the source's).
- `GEO_RESTORE` -- Restore the source's geo-redundant backup into the paired region (requires source_server_id + point_in_time_restore_time_in_utc, and geo_redundant_backup_enabled on the source).
- `REVIVE_DROPPED` -- Revive a soft-deleted (dropped) server (requires source_server_id).

### spec.sourceServerId

`string | valueFrom`

For REPLICA, POINT_IN_TIME_RESTORE, GEO_RESTORE, and REVIVE_DROPPED:
the ARM ID of the server to replicate or restore from. References
another AzurePostgresqlFlexibleServer's server_id output, so a
primary-plus-replicas topology composes in one manifest set. Fixed at
creation.

- references: AzurePostgresqlFlexibleServer (`status.outputs.server_id`)
- rule: write as {value: <literal>} or {valueFrom: {kind: AzurePostgresqlFlexibleServer, name: <that resource's name>, fieldPath: status.outputs.server_id}} -- a bare string does not parse

### spec.pointInTimeRestoreTimeInUtc

`string`

For POINT_IN_TIME_RESTORE and GEO_RESTORE: the RFC-3339 UTC instant
to restore the source server to (e.g. "2026-07-01T08:30:00Z"). Must
fall inside the source's backup retention window. Fixed at creation.

- rule: point_in_time_restore_time_in_utc must be an RFC-3339 UTC timestamp, e.g. 2026-07-01T08:30:00Z

### spec.replicationRole

`enum`

Promotion control for a replica. Setting NONE on a server created
with create_mode REPLICA breaks replication and promotes it to a
standalone read-write primary (irreversible). Cannot be set at
creation and has no meaning on non-replica servers.

Allowed values (use exactly as shown):

- `azure_postgresql_flexible_server_replication_role_unspecified` -- Not specified: leave the replication topology as created.
- `NONE` -- Break replication and promote this replica to a standalone read-write primary (irreversible).

### spec.administratorLogin

`string`

The administrator login for PostgreSQL password authentication.
Required for a fresh (DEFAULT) server with password auth enabled;
must be OMITTED when password auth is disabled. Azure reserves
"azure_superuser", "azure_pg_admin", "admin", "administrator",
"root", "guest", "public", and any name starting with "pg_". The
login is fixed once set.

- rule: administrator_login cannot be azure_superuser, azure_pg_admin, admin, administrator, root, guest, or public, and cannot start with pg_

### spec.administratorPassword

`string | valueFrom` · sensitive

The administrator password (8-128 characters, from at least three of:
uppercase, lowercase, digits, special characters). Can be a literal
value or a reference to another resource's output. Required for a
fresh server with password auth enabled; must be OMITTED when
password auth is disabled. Updatable in place. The provider's
write-only variant (administrator_password_wo + its version counter)
is deliberately not modeled: it is an ephemeral input that duplicates
this field, and secret values here are already reference-resolved at
deploy time rather than stored in the manifest.

- rule: write as {value: <literal>} or {valueFrom: {kind: <Kind>, name: <that resource's name>, fieldPath: status.outputs.<output>}} -- a bare string does not parse

### spec.version

`string` · optional (explicit presence)

The PostgreSQL major version. Unspecified applies "16" (the current
production standard). In-place upgrades to a HIGHER major version are
supported (irreversible, brief downtime); downgrading replaces the
server. Elastic clusters require 17 or later.

- "11"/"12"/"13": end-of-life community versions -- legacy migrations only
- "14"/"15": active, approaching end of standard support
- "16": recommended default for new deployments
- "17": latest GA -- required for elastic clusters
- "18": newest supported release

- default: `16`
- rule: version must be one of: 11, 12, 13, 14, 15, 16, 17, 18

### spec.skuName

`string`

The compute SKU, as {TIER}_Standard_{SIZE}: B_ (Burstable -- dev/test,
e.g. "B_Standard_B1ms", "B_Standard_B2s"), GP_ (General Purpose --
production, e.g. "GP_Standard_D2s_v3", "GP_Standard_D4ds_v5"), or
MO_ (Memory Optimized -- analytics/caching, e.g. "MO_Standard_E4s_v3").
Required for a fresh (DEFAULT) server; a replica left unset inherits
the source's SKU. Resizable in place (brief restart), but switching
between confidential-compute (C-series, e.g. GP_Standard_DC4ads_v5)
and regular compute requires a migration. Burstable SKUs do not
support high availability or read replicas.

- rule: sku_name must be {TIER}_Standard_{SIZE} where TIER is B, GP, or MO, e.g. B_Standard_B1ms, GP_Standard_D2s_v3, MO_Standard_E4s_v3

### spec.storageMb

`int32` · optional (explicit presence)

The provisioned storage size in MB, from Azure's fixed ladder:
32768 (32 GiB), 65536, 131072, 262144, 524288, 1048576 (1 TiB),
2097152, 4193280, 4194304 (4 TiB), 8388608 (8 TiB), 16777216 (16 TiB),
33553408 (32 TiB). Unspecified applies Azure's default of 32768; a
replica inherits the source's size. Storage only grows -- shrinking
replaces the server.

- rule: storage_mb must be one of Azure's supported sizes: 32768, 65536, 131072, 262144, 524288, 1048576, 2097152, 4193280, 4194304, 8388608, 16777216, 33553408

### spec.storageTier

`enum`

The storage performance tier (IOPS class). Unspecified applies the
default tier for the chosen storage_mb. Higher tiers buy more IOPS
without growing capacity, but each storage size supports a bounded
tier range (validated here, mirroring Azure's matrix). The tier can
be changed once every 12 hours.

Allowed values (use exactly as shown):

- `azure_postgresql_flexible_server_storage_tier_unspecified` -- Not specified: the default tier for the chosen storage_mb.
- `P4` -- 120 IOPS -- the default for 32 GiB.
- `P6` -- 240 IOPS -- the default for 64 GiB.
- `P10` -- 500 IOPS -- the default for 128 GiB.
- `P15` -- 1,100 IOPS -- the default for 256 GiB.
- `P20` -- 2,300 IOPS -- the default for 512 GiB.
- `P30` -- 5,000 IOPS -- the default for 1 TiB.
- `P40` -- 7,500 IOPS -- the default for 2 TiB.
- `P50` -- 7,500 IOPS -- the default (and only) tier for 4 TiB.
- `P60` -- 16,000 IOPS -- the default for 8 TiB.
- `P70` -- 18,000 IOPS -- the default for 16 TiB.
- `P80` -- 20,000 IOPS -- the default (and only) tier for 32 TiB.

### spec.autoGrowEnabled

`bool`

Whether storage grows automatically when free space runs low,
stepping up the storage_mb ladder without downtime. Azure's default
is false (storage stays at the provisioned size).

### spec.zone

`string`

The availability zone for the primary server ("1", "2", or "3").
Unset lets Azure choose. With ZONE_REDUNDANT high availability, set
this (and standby_availability_zone) to control placement; after
creation the zone can only change via a planned failover that swaps
primary and standby.

- rule: zone must be "1", "2", or "3"

### spec.highAvailability

`AzurePostgresqlFlexibleServerHighAvailability`

High availability: presence enables a standby server with synchronous
replication and automatic failover. Omit for a single-instance
server. Burstable SKUs do not support HA.

### spec.highAvailability.mode

`enum` · required

ZONE_REDUNDANT places the standby in a different availability zone
(survives zone failure -- the production recommendation); SAME_ZONE
co-locates it (faster failover, no zone-level protection).

- rule: {"required":true}

Allowed values (use exactly as shown):

- `azure_postgresql_flexible_server_high_availability_mode_unspecified` -- Not specified -- invalid; choose an explicit mode when the HA block is present.
- `ZONE_REDUNDANT` -- Standby in a different availability zone: survives zone failure. The production recommendation.
- `SAME_ZONE` -- Standby in the same zone: faster failover, no zone-level protection.

### spec.highAvailability.standbyAvailabilityZone

`string`

The availability zone for the standby ("1", "2", or "3"). Unset lets
Azure choose. For ZONE_REDUNDANT it must differ from the primary's
zone. Fixed at creation -- after that, zones only change via planned
failover.

- rule: standby_availability_zone must be "1", "2", or "3"

### spec.maintenanceWindow

`AzurePostgresqlFlexibleServerMaintenanceWindow`

The weekly maintenance window for Azure-managed patching. Omit for
a system-managed window (Azure picks an off-peak slot). Presence
pins patching to the declared day and start time.

### spec.maintenanceWindow.dayOfWeek

`int32` · optional (explicit presence)

Day of the week: 0 (Sunday) through 6 (Saturday).

- rule: {"int32":{"lte":6,"gte":0}}

### spec.maintenanceWindow.startHour

`int32` · optional (explicit presence)

The window's start hour, 0-23 (server-local UTC).

- rule: {"int32":{"lte":23,"gte":0}}

### spec.maintenanceWindow.startMinute

`int32` · optional (explicit presence)

The window's start minute, 0-59.

- rule: {"int32":{"lte":59,"gte":0}}

### spec.backupRetentionDays

`int32` · optional (explicit presence)

How many days automatic backups are retained (the point-in-time
restore horizon). 7-35; unspecified applies Azure's default of 7.

- default: `7`
- rule: {"int32":{"lte":35,"gte":7}}

### spec.geoRedundantBackupEnabled

`bool`

Whether backups are replicated to the paired Azure region, enabling
cross-region GEO_RESTORE for disaster recovery. Azure's default is
false. Fixed at creation.

### spec.publicNetworkAccessEnabled

`bool` · optional (explicit presence)

Whether the server accepts connections on its public endpoint.
Azure's default is true, with firewall_rules as the allowlist. Must
be explicitly false when the server is VNet-injected via
delegated_subnet_id -- Azure rejects a public endpoint on an injected
server.

- default: `true`

### spec.delegatedSubnetId

`string | valueFrom`

For VNet injection (private access): the ARM ID of a subnet delegated
to Microsoft.DBforPostgreSQL/flexibleServers, with no other resources
in it. Requires private_dns_zone_id and public network access OFF.
Fixed at creation. Use an AzureSubnet with the matching delegation.

- references: AzureSubnet (`status.outputs.subnet_id`)
- rule: write as {value: <literal>} or {valueFrom: {kind: AzureSubnet, name: <that resource's name>, fieldPath: status.outputs.subnet_id}} -- a bare string does not parse

### spec.privateDnsZoneId

`string | valueFrom`

For VNet injection: the ARM ID of a private DNS zone (conventionally
ending ".postgres.database.azure.com") where the server registers its
private address, so VNet-connected clients resolve the fqdn to the
private IP. Required whenever delegated_subnet_id is set.

- references: AzurePrivateDnsZone (`status.outputs.zone_id`)
- rule: write as {value: <literal>} or {valueFrom: {kind: AzurePrivateDnsZone, name: <that resource's name>, fieldPath: status.outputs.zone_id}} -- a bare string does not parse

### spec.authentication

`AzurePostgresqlFlexibleServerAuthentication`

Which authentication mechanisms the server accepts. Omit for Azure's
default: password auth on, Microsoft Entra auth off.

- rule: at least one of password_auth_enabled and active_directory_auth_enabled must be true
- rule: tenant_id is only meaningful when active_directory_auth_enabled is true

### spec.authentication.activeDirectoryAuthEnabled

`bool`

Whether Microsoft Entra (Azure AD) authentication is enabled --
clients connect with Entra tokens instead of passwords, and
aad_administrators become grantable. Azure's default is false.

### spec.authentication.passwordAuthEnabled

`bool` · optional (explicit presence)

Whether PostgreSQL password authentication is enabled. Azure's
default is true. Disabling it (for an Entra-only posture) requires
active_directory_auth_enabled and forbids the admin login/password
pair on the spec.

- default: `true`

### spec.authentication.tenantId

`string` · optional (explicit presence)

The Entra tenant of the aad_administrators principals. Leave unset
to use the deploying credential's tenant -- the correct value for
virtually every deployment.

- rule: {"string":{"uuid":true}}

### spec.aadAdministrators

`[]AzurePostgresqlFlexibleServerAadAdministrator`

Microsoft Entra (Azure AD) principals granted the server's Entra
administrator role -- they can connect as admins with Entra tokens
and manage other Entra roles inside PostgreSQL. Requires
authentication.active_directory_auth_enabled. Each entry is its own
Azure sub-resource keyed by the principal's object ID.

### spec.aadAdministrators[].objectId

`string | valueFrom` · required

The object ID of the Entra principal being granted the administrator
role (a user, group, or service principal / managed identity).
Defaults to referencing an AzureUserAssignedIdentity's principal_id
output in composed environments -- note this is the identity's
PRINCIPAL id (the directory object), not its client id.

- references: AzureUserAssignedIdentity (`status.outputs.principal_id`)
- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: AzureUserAssignedIdentity, name: <that resource's name>, fieldPath: status.outputs.principal_id}} -- a bare string does not parse

### spec.aadAdministrators[].principalName

`string` · required

The principal's display name as it appears in Entra (e.g.
"dba-team@contoso.com" for a user, the group or identity name
otherwise). PostgreSQL uses it as the role name to connect as.

- rule: {"required":true,"string":{"minLen":"1"}}

### spec.aadAdministrators[].principalType

`enum` · required

What kind of Entra principal this is. Azure validates the type
against the object ID, so it must match the directory object.

- rule: {"required":true}

Allowed values (use exactly as shown):

- `azure_postgresql_flexible_server_aad_principal_type_unspecified` -- Not specified -- invalid; declare the principal's directory type.
- `USER` -- A directory user.
- `GROUP` -- A directory group (every member becomes an Entra admin).
- `SERVICE_PRINCIPAL` -- A service principal or managed identity.

### spec.identity

`AzurePostgresqlFlexibleServerIdentity`

The server's managed identity. Required (with a user-assigned
identity) for customer-managed-key encryption; also usable to grant
the server access to other Azure resources.

- rule: identity_ids is required for USER_ASSIGNED and SYSTEM_AND_USER_ASSIGNED and must be empty for SYSTEM_ASSIGNED

### spec.identity.type

`enum` · required

Identity flavor. SYSTEM_ASSIGNED is created and rotated by Azure with
the server; USER_ASSIGNED brings identities you manage and share
across resources (required for customer-managed-key encryption);
SYSTEM_AND_USER_ASSIGNED carries both.

- rule: {"required":true}

Allowed values (use exactly as shown):

- `azure_postgresql_flexible_server_identity_type_unspecified` -- Not specified: the server has no managed identity.
- `SYSTEM_ASSIGNED` -- Azure-managed identity created and rotated with the server.
- `USER_ASSIGNED` -- Bring-your-own user-assigned identities (set identity_ids) -- required for customer-managed-key encryption.
- `SYSTEM_AND_USER_ASSIGNED` -- Both a system-assigned identity and the listed user-assigned ones.

### spec.identity.identityIds

`[]string | valueFrom`

For USER_ASSIGNED / SYSTEM_AND_USER_ASSIGNED: the user-assigned
identities attached to the server, by ARM ID. Reference
AzureUserAssignedIdentity resources so Key Vault grants can be
composed before the server exists.

- references: AzureUserAssignedIdentity (`status.outputs.identity_id`)
- rule: write as {value: <literal>} or {valueFrom: {kind: AzureUserAssignedIdentity, name: <that resource's name>, fieldPath: status.outputs.identity_id}} -- a bare string does not parse

### spec.customerManagedKey

`AzurePostgresqlFlexibleServerCustomerManagedKey`

Customer-managed-key (CMK) encryption: the server's data is encrypted
with a Key Vault key you own instead of a Microsoft-managed key.
Requires a user-assigned identity (in identity.identity_ids) that has
wrap/unwrap access on the key's vault. Fixed at creation.

- rule: geo_backup_key_vault_key_id and geo_backup_user_assigned_identity_id must be set together

### spec.customerManagedKey.keyVaultKeyId

`string | valueFrom` · required

The Key Vault key that encrypts the server's data, by data-plane key
ID. Defaults to referencing an AzureKeyVaultKey's versionless_id
output so key rotations propagate automatically; pin a versioned ID
only when a compliance regime demands an immutable key version. The
key's vault must have purge protection enabled.

- references: AzureKeyVaultKey (`status.outputs.versionless_id`)
- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: AzureKeyVaultKey, name: <that resource's name>, fieldPath: status.outputs.versionless_id}} -- a bare string does not parse

### spec.customerManagedKey.primaryUserAssignedIdentityId

`string | valueFrom`

The user-assigned identity Azure uses to unwrap the key, by ARM ID.
Must be one of the identities in identity.identity_ids, with
wrap/unwrap access on the key's vault (a "Key Vault Crypto Service
Encryption User" role assignment, or the equivalent access policy).

- references: AzureUserAssignedIdentity (`status.outputs.identity_id`)
- rule: write as {value: <literal>} or {valueFrom: {kind: AzureUserAssignedIdentity, name: <that resource's name>, fieldPath: status.outputs.identity_id}} -- a bare string does not parse

### spec.customerManagedKey.geoBackupKeyVaultKeyId

`string | valueFrom`

For geo-redundant backups: the Key Vault key (in the paired region's
vault) that encrypts the geo-replicated backup data. Requires
geo_backup_user_assigned_identity_id.

- references: AzureKeyVaultKey (`status.outputs.versionless_id`)
- rule: write as {value: <literal>} or {valueFrom: {kind: AzureKeyVaultKey, name: <that resource's name>, fieldPath: status.outputs.versionless_id}} -- a bare string does not parse

### spec.customerManagedKey.geoBackupUserAssignedIdentityId

`string | valueFrom`

The user-assigned identity that unwraps the geo-backup key, by ARM
ID. Required together with geo_backup_key_vault_key_id.

- references: AzureUserAssignedIdentity (`status.outputs.identity_id`)
- rule: write as {value: <literal>} or {valueFrom: {kind: AzureUserAssignedIdentity, name: <that resource's name>, fieldPath: status.outputs.identity_id}} -- a bare string does not parse

### spec.cluster

`AzurePostgresqlFlexibleServerCluster`

Elastic cluster: presence provisions the server as a sharded,
horizontally-distributed cluster (PostgreSQL 17+, citus-based) of the
declared node count instead of a single node. Fixed at creation and
only valid with create_mode DEFAULT.

### spec.cluster.size

`int32` · required

The number of nodes in the cluster, 1-20. Grows in place; shrinking
replaces the cluster.

- rule: {"required":true,"int32":{"lte":20,"gte":1}}

### spec.cluster.defaultDatabaseName

`string` · required · optional (explicit presence)

The name of the cluster's default database. Unspecified applies
Azure's default ("postgres"). Fixed at creation.

- default: `postgres`
- rule: {"string":{"minLen":"1"}}

### spec.databases

`[]AzurePostgresqlFlexibleServerDatabase`

Databases to create on the server, each a separate Azure sub-resource
with its own charset/collation. Azure always creates the built-in
"postgres" database; most applications declare at least one database
here.

### spec.databases[].name

`string` · required

The database name: 1-63 characters, starting with a letter,
containing letters, digits, hyphens, and underscores; unique within
the server.

- rule: {"required":true,"string":{"minLen":"1","maxLen":"63","pattern":"^[a-zA-Z][a-zA-Z0-9-_]*$"}}

### spec.databases[].charset

`string` · required · optional (explicit presence)

The database character set. Unspecified applies "UTF8" (the right
choice for virtually all applications). Other PostgreSQL charsets
(e.g. "SQL_ASCII", "LATIN1", "EUC_JP", "WIN1252") are accepted for
legacy migrations.

- default: `UTF8`
- rule: {"string":{"minLen":"1"}}

### spec.databases[].collation

`string` · required · optional (explicit presence)

The database collation (sort order and string comparison), e.g.
"en_US.utf8" (the default), "C", "POSIX", or an ICU collation like
"de-x-icu".

- default: `en_US.utf8`
- rule: {"string":{"minLen":"1"}}

### spec.firewallRules

`[]AzurePostgresqlFlexibleServerFirewallRule`

Public-endpoint firewall allowlist: each rule admits a contiguous
IPv4 range. Only meaningful while public_network_access_enabled is
true. The special rule 0.0.0.0-0.0.0.0 admits Azure-internal services
only (not the internet).

### spec.firewallRules[].name

`string` · required

The rule's name: 1-128 letters, digits, hyphens, and underscores;
unique within the server. E.g. "allow-office", "allow-ci".

- rule: {"required":true,"string":{"minLen":"1","maxLen":"128","pattern":"^[a-zA-Z0-9-_]+$"}}

### spec.firewallRules[].startIpAddress

`string` · required

The first IPv4 address of the admitted range (inclusive). Use
0.0.0.0 for both start and end to admit Azure-internal services.

- rule: {"required":true,"string":{"ipv4":true}}

### spec.firewallRules[].endIpAddress

`string` · required

The last IPv4 address of the admitted range (inclusive). Equal to
start_ip_address for a single-address rule.

- rule: {"required":true,"string":{"ipv4":true}}

### spec.serverParameters

`map<string, string>`

PostgreSQL server parameters to override (e.g. "shared_preload_libraries",
"max_connections", "azure.extensions"), by parameter name. Values are
applied as user overrides on Azure's per-SKU defaults; removing an
entry resets the parameter to its default. Static (non-dynamic)
parameters need a server restart to take effect -- Azure applies them
but reports "pending restart" until one happens.

- rule: {"map":{"keys":{"string":{"minLen":"1"}},"values":{"string":{"minLen":"1"}}}}

### spec.tags

`map<string, string>`

Free-form tags applied to the server, merged over the Planton-derived
resource tags (organization, environment, resource id); a user tag
with the same key wins. Tags are Azure's governance surface -- Azure
Policy enforces them and Microsoft Cost Management groups by them.

## Validation Rules

- `postgres_source_matches_create_mode`: source_server_id is required for REPLICA, POINT_IN_TIME_RESTORE, GEO_RESTORE, and REVIVE_DROPPED, and must be omitted for a fresh (DEFAULT) server
- `postgres_restore_time_matches_create_mode`: point_in_time_restore_time_in_utc is required for POINT_IN_TIME_RESTORE and GEO_RESTORE, and must be omitted for other create modes
- `postgres_replication_role_replica_only`: replication_role (NONE, replica promotion) is only valid on a server created with create_mode REPLICA
- `postgres_password_auth_requires_credentials`: a fresh (DEFAULT) server with password authentication enabled requires administrator_login and administrator_password
- `postgres_no_credentials_without_password_auth`: administrator_login and administrator_password must be omitted when authentication.password_auth_enabled is false
- `postgres_sku_required_for_default_mode`: sku_name is required for a fresh (DEFAULT) server; only replicas and restores may omit it to inherit from the source
- `postgres_aad_admins_require_aad_auth`: aad_administrators requires authentication.active_directory_auth_enabled = true
- `postgres_cmk_requires_user_assigned_identity`: customer_managed_key requires primary_user_assigned_identity_id and an identity block that includes a user-assigned identity (USER_ASSIGNED or SYSTEM_AND_USER_ASSIGNED)
- `postgres_cluster_requires_default_mode`: cluster (elastic cluster) is only valid on a fresh (DEFAULT) server -- not on replicas or restores
- `postgres_cluster_requires_version_17`: cluster (elastic cluster) requires PostgreSQL version 17 or 18, set explicitly
- `postgres_vnet_injection_requires_private_dns_zone`: delegated_subnet_id (VNet injection) requires private_dns_zone_id
- `postgres_vnet_injection_requires_public_access_off`: delegated_subnet_id (VNet injection) requires public_network_access_enabled = false, set explicitly
- `postgres_storage_tier_matches_storage_mb`: storage_tier is outside the supported range for storage_mb: 32GB starts at P4, 64GB at P6, 128GB at P10, 256GB at P15, 512GB at P20, 1TB at P30, 2TB at P40, 4TB is P50 only, 8TB is P60-P80, 16TB is P70-P80, 32TB is P80 only

## Outputs

Reference an output from another manifest as `valueFrom: {kind: AzurePostgresqlFlexibleServer, name: <resource-name>, fieldPath: status.outputs.<output>}`.

| Output | Type | Description |
|---|---|---|
| `status.outputs.server_id` | `string` | The Azure Resource Manager ID of the server. Format: /subscriptions/{sub}/resourceGroups/{rg}/providers/Microsoft.DBforPostgreSQL/flexibleServers/{name} Referenced by AzurePrivateEndpoint (private_connection_resource_id) and by replica/restore servers (source_server_id). |
| `status.outputs.server_name` | `string` | The name of the server. |
| `status.outputs.fqdn` | `string` | The server's fully qualified domain name ({name}.postgres.database.azure.com). For a VNet-injected server it resolves to the private address through the private DNS zone. Connection strings take the shape: postgresql://{login}:{password}@{fqdn}:5432/{database}?sslmode=require |
| `status.outputs.administrator_login` | `string` | The administrator login, echoed so applications can construct connection strings without duplicating the value. Empty on an Entra-only server (password auth disabled). |
| `status.outputs.database_ids` | `map<string, string>` | The ARM ID of each database declared in the spec, keyed by database name. Example valueFrom fieldPath: status.outputs.database_ids.myapp |
| `status.outputs.identity_principal_id` | `string` | The principal (directory object) ID of the server's system-assigned managed identity -- the subject for AzureRoleAssignment grants. Empty unless the identity type includes SYSTEM_ASSIGNED. |

## References

Fields that can point at another resource's outputs:

| Field | Kind | Output |
|---|---|---|
| `spec.resourceGroup` | AzureResourceGroup | `status.outputs.resource_group_name` |
| `spec.sourceServerId` | AzurePostgresqlFlexibleServer | `status.outputs.server_id` |
| `spec.delegatedSubnetId` | AzureSubnet | `status.outputs.subnet_id` |
| `spec.privateDnsZoneId` | AzurePrivateDnsZone | `status.outputs.zone_id` |
| `spec.aadAdministrators[].objectId` | AzureUserAssignedIdentity | `status.outputs.principal_id` |
| `spec.identity.identityIds` | AzureUserAssignedIdentity | `status.outputs.identity_id` |
| `spec.customerManagedKey.keyVaultKeyId` | AzureKeyVaultKey | `status.outputs.versionless_id` |
| `spec.customerManagedKey.primaryUserAssignedIdentityId` | AzureUserAssignedIdentity | `status.outputs.identity_id` |
| `spec.customerManagedKey.geoBackupKeyVaultKeyId` | AzureKeyVaultKey | `status.outputs.versionless_id` |
| `spec.customerManagedKey.geoBackupUserAssignedIdentityId` | AzureUserAssignedIdentity | `status.outputs.identity_id` |

## Referenced By

Fields on other kinds that can point at this resource:

| Kind | Field | Reads |
|---|---|---|
| AzureDataProtectionBackupInstance | `spec.postgresqlFlexibleServer.serverId` | `status.outputs.server_id` |
| AzurePostgresqlFlexibleServer | `spec.sourceServerId` | `status.outputs.server_id` |

## See Also

- [Overview](../README.md)
