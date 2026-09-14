# AzureMysqlFlexibleServer

> Generated from the protobuf schema by `make generate-reference` -- do not
> edit by hand. To change a fact on this page, change the proto field comment
> or validation rule it is derived from, then regenerate.

**apiVersion**: `azure.planton.dev/v1alpha1`

**AzureMysqlFlexibleServerSpec** defines the configuration for creating an
Azure Database for MySQL Flexible Server: Azure's managed MySQL with
per-server compute/storage sizing, zone-redundant high availability,
Microsoft Entra (Azure AD) administration, customer-managed-key encryption,
read replicas, and point-in-time restore.

The server is the unit of management -- databases, firewall rules, server
parameters, and the Entra administrator are configured on it and live with
it, so they are folded into this spec rather than modeled as standalone
kinds (none of them has an independent lifecycle or is referenced by
anything else).

**Network access** has two mutually exclusive postures, matching Azure's
real contract: a public endpoint (with `firewall_rules` as its allowlist)
or VNet injection via `delegated_subnet_id` + `private_dns_zone_id`. A
VNet-injected server cannot have a public endpoint -- leave
`public_network_access` unset and Azure derives DISABLED.

**Authentication** always includes MySQL password auth (unlike PostgreSQL
Flexible Server, it cannot be switched off). Microsoft Entra authentication
is additive: declare the single `aad_administrator` (MySQL supports exactly
one) backed by a user-assigned identity attached to the server.

**Lifecycle modes** (`create_mode`) cover the full server story: a fresh
server (DEFAULT), a read replica (REPLICA + `source_server_id`),
point-in-time restore (POINT_IN_TIME_RESTORE + the restore timestamp), and
cross-region geo-restore of the latest geo-redundant backup (GEO_RESTORE).
A replica is promoted to a standalone primary by setting `replication_role`
to NONE.

## Example

```yaml
apiVersion: azure.planton.dev/v1alpha1
kind: AzureMysqlFlexibleServer
metadata:
  name: test-mysql
spec:
  region: eastus
  resourceGroup:
    value: test-rg
  serverName: test-mysql-server
  administratorLogin: mysqladmin
  administratorPassword:
    value: P@ssw0rd1234!
  version: "8.0.21"
  skuName: GP_Standard_D2ds_v4
  # Exercises the full storage block: fixed size with provisioned IOPS
  # (mutually exclusive with elastic IO scaling) and the slow-query-log
  # placement dial.
  storage:
    sizeGb: 128
    iops: 900
    autoGrowEnabled: true
    logOnDiskEnabled: true
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
  # Exercises the user-assigned identity list plus the Entra
  # administrator sub-resource path (MySQL's single admin, backed by an
  # attached identity).
  userAssignedIdentityIds:
    - value: /subscriptions/00000000-0000-0000-0000-000000000000/resourceGroups/test-rg/providers/Microsoft.ManagedIdentity/userAssignedIdentities/mysql-uai
  aadAdministrator:
    identityId:
      value: /subscriptions/00000000-0000-0000-0000-000000000000/resourceGroups/test-rg/providers/Microsoft.ManagedIdentity/userAssignedIdentities/mysql-uai
    login: dba-team@contoso.com
    objectId:
      value: 11111111-2222-3333-4444-555555555555
  databases:
    - name: myapp
    - name: legacy
      charset: latin1
      collation: latin1_swedish_ci
  firewallRules:
    - name: allow-azure-services
      startIpAddress: "0.0.0.0"
      endIpAddress: "0.0.0.0"
    - name: allow-office
      startIpAddress: "203.0.113.0"
      endIpAddress: "203.0.113.255"
  # Exercises the server-parameter configuration sub-resources.
  serverParameters:
    require_secure_transport: "ON"
    max_connections: "500"
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
| `spec.sourceServerId` | `string \| valueFrom` |  |  | AzureMysqlFlexibleServer (`status.outputs.server_id`) |
| `spec.pointInTimeRestoreTimeInUtc` | `string` |  |  |  |
| `spec.replicationRole` | `enum` |  |  |  |
| `spec.administratorLogin` | `string` |  |  |  |
| `spec.administratorPassword` | `string \| valueFrom` (sensitive) |  |  |  |
| `spec.version` | `string` |  | `8.0.21` |  |
| `spec.skuName` | `string` |  |  |  |
| `spec.storage` | `AzureMysqlFlexibleServerStorage` |  |  |  |
| `spec.storage.sizeGb` | `int32` |  |  |  |
| `spec.storage.iops` | `int32` |  |  |  |
| `spec.storage.autoGrowEnabled` | `bool` |  | `true` |  |
| `spec.storage.ioScalingEnabled` | `bool` |  |  |  |
| `spec.storage.logOnDiskEnabled` | `bool` |  |  |  |
| `spec.zone` | `string` |  |  |  |
| `spec.highAvailability` | `AzureMysqlFlexibleServerHighAvailability` |  |  |  |
| `spec.highAvailability.mode` | `enum` | yes |  |  |
| `spec.highAvailability.standbyAvailabilityZone` | `string` |  |  |  |
| `spec.maintenanceWindow` | `AzureMysqlFlexibleServerMaintenanceWindow` |  |  |  |
| `spec.maintenanceWindow.dayOfWeek` | `int32` |  |  |  |
| `spec.maintenanceWindow.startHour` | `int32` |  |  |  |
| `spec.maintenanceWindow.startMinute` | `int32` |  |  |  |
| `spec.backupRetentionDays` | `int32` |  | `7` |  |
| `spec.geoRedundantBackupEnabled` | `bool` |  |  |  |
| `spec.publicNetworkAccess` | `enum` |  |  |  |
| `spec.delegatedSubnetId` | `string \| valueFrom` |  |  | AzureSubnet (`status.outputs.subnet_id`) |
| `spec.privateDnsZoneId` | `string \| valueFrom` |  |  | AzurePrivateDnsZone (`status.outputs.zone_id`) |
| `spec.userAssignedIdentityIds` | `[]string \| valueFrom` |  |  | AzureUserAssignedIdentity (`status.outputs.identity_id`) |
| `spec.customerManagedKey` | `AzureMysqlFlexibleServerCustomerManagedKey` |  |  |  |
| `spec.customerManagedKey.keyVaultKeyId` | `string \| valueFrom` | yes |  | AzureKeyVaultKey (`status.outputs.versionless_id`) |
| `spec.customerManagedKey.primaryUserAssignedIdentityId` | `string \| valueFrom` |  |  | AzureUserAssignedIdentity (`status.outputs.identity_id`) |
| `spec.customerManagedKey.geoBackupKeyVaultKeyId` | `string \| valueFrom` |  |  | AzureKeyVaultKey (`status.outputs.versionless_id`) |
| `spec.customerManagedKey.geoBackupUserAssignedIdentityId` | `string \| valueFrom` |  |  | AzureUserAssignedIdentity (`status.outputs.identity_id`) |
| `spec.aadAdministrator` | `AzureMysqlFlexibleServerAadAdministrator` |  |  |  |
| `spec.aadAdministrator.identityId` | `string \| valueFrom` | yes |  | AzureUserAssignedIdentity (`status.outputs.identity_id`) |
| `spec.aadAdministrator.login` | `string` | yes |  |  |
| `spec.aadAdministrator.objectId` | `string \| valueFrom` | yes |  | AzureUserAssignedIdentity (`status.outputs.client_id`) |
| `spec.aadAdministrator.tenantId` | `string` |  |  |  |
| `spec.databases` | `[]AzureMysqlFlexibleServerDatabase` |  |  |  |
| `spec.databases[].name` | `string` | yes |  |  |
| `spec.databases[].charset` | `string` | yes | `utf8mb4` |  |
| `spec.databases[].collation` | `string` | yes | `utf8mb4_0900_ai_ci` |  |
| `spec.firewallRules` | `[]AzureMysqlFlexibleServerFirewallRule` |  |  |  |
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
({name}.mysql.database.azure.com, the fqdn output). Changing the name
replaces the server.

- rule: {"required":true,"string":{"minLen":"3","maxLen":"63","pattern":"^[a-z0-9]([a-z0-9-]*[a-z0-9])?$"}}

### spec.createMode

`enum`

How the server comes into existence. Unspecified means DEFAULT (a
fresh, empty server). The restore/replica modes consume
source_server_id (and, for point-in-time restore, the restore
timestamp) and are fixed at creation.

Allowed values (use exactly as shown):

- `azure_mysql_flexible_server_create_mode_unspecified` -- Not specified: DEFAULT (a fresh, empty server).
- `DEFAULT` -- A fresh, empty server (requires sku_name and the admin credentials).
- `POINT_IN_TIME_RESTORE` -- Restore the source server to a point in time, in the same region (requires source_server_id + point_in_time_restore_time_in_utc).
- `REPLICA` -- A read replica of the source server, asynchronously replicated (requires source_server_id; SKU/storage left unset inherit the source's).
- `GEO_RESTORE` -- Restore the source's latest geo-redundant backup into the paired region (requires source_server_id and geo_redundant_backup_enabled on the source; takes no timestamp).

### spec.sourceServerId

`string | valueFrom`

For REPLICA, POINT_IN_TIME_RESTORE, and GEO_RESTORE: the ARM ID of the
server to replicate or restore from. References another
AzureMysqlFlexibleServer's server_id output, so a primary-plus-replicas
topology composes in one manifest set. Fixed at creation.

- references: AzureMysqlFlexibleServer (`status.outputs.server_id`)
- rule: write as {value: <literal>} or {valueFrom: {kind: AzureMysqlFlexibleServer, name: <that resource's name>, fieldPath: status.outputs.server_id}} -- a bare string does not parse

### spec.pointInTimeRestoreTimeInUtc

`string`

For POINT_IN_TIME_RESTORE only: the RFC-3339 UTC instant to restore
the source server to (e.g. "2026-07-01T08:30:00Z"). Must fall inside
the source's backup retention window. GEO_RESTORE takes no timestamp --
it restores the latest geo-replicated backup. Fixed at creation.

- rule: point_in_time_restore_time_in_utc must be an RFC-3339 UTC timestamp, e.g. 2026-07-01T08:30:00Z

### spec.replicationRole

`enum`

Promotion control for a replica. Setting NONE on a server created
with create_mode REPLICA breaks replication and promotes it to a
standalone read-write primary (irreversible). Cannot be set at
creation and has no meaning on non-replica servers.

Allowed values (use exactly as shown):

- `azure_mysql_flexible_server_replication_role_unspecified` -- Not specified: leave the replication topology as created.
- `NONE` -- Break replication and promote this replica to a standalone read-write primary (irreversible).

### spec.administratorLogin

`string`

The administrator login for MySQL password authentication: 1-32
letters, digits, and underscores. Azure reserves "azure_superuser",
"admin", "administrator", "root", "guest", and "public". Required for
a fresh (DEFAULT) server; a replica or restore inherits the source's
login. The login is fixed once set.

- rule: administrator_login must contain only letters, digits, and underscores, and cannot be azure_superuser, admin, administrator, root, guest, or public
- rule: {"string":{"maxLen":"32"}}

### spec.administratorPassword

`string | valueFrom` · sensitive

The administrator password (8-128 characters, from at least three of:
uppercase, lowercase, digits, special characters). Can be a literal
value or a reference to another resource's output. Required for a
fresh (DEFAULT) server; updatable in place. The provider's write-only
variant (administrator_password_wo + its version counter) is
deliberately not modeled: it is an ephemeral input that duplicates
this field, and secret values here are already reference-resolved at
deploy time rather than stored in the manifest.

- rule: write as {value: <literal>} or {valueFrom: {kind: <Kind>, name: <that resource's name>, fieldPath: status.outputs.<output>}} -- a bare string does not parse

### spec.version

`string` · optional (explicit presence)

The MySQL version, using Azure's exact version strings. Unspecified
applies "8.0.21" (Azure's identifier for the MySQL 8.0 series -- the
production standard). In-place upgrade from 5.7 to 8.0.21 is
supported (irreversible); downgrading replaces the server.

- "5.7": approaching end of life -- legacy migrations only
- "8.0.21": the MySQL 8.0 series, recommended for new deployments
- "8.4": the newest supported LTS release

- default: `8.0.21`
- rule: version must be one of: 5.7, 8.0.21, 8.4

### spec.skuName

`string`

The compute SKU, as {TIER}_Standard_{SIZE}: B_ (Burstable -- dev/test,
e.g. "B_Standard_B1ms", "B_Standard_B2s"), GP_ (General Purpose --
production, e.g. "GP_Standard_D2ds_v4", "GP_Standard_D4ads_v5"), or
MO_ (Memory Optimized -- analytics/caching, e.g. "MO_Standard_E4ds_v4").
Required for a fresh (DEFAULT) server; a replica left unset inherits
the source's SKU. Resizable in place (brief restart). Burstable SKUs
do not support high availability or read replicas.

- rule: sku_name must be {TIER}_Standard_{SIZE} where TIER is B, GP, or MO, e.g. B_Standard_B1ms, GP_Standard_D2ds_v4, MO_Standard_E4ds_v4

### spec.storage

`AzureMysqlFlexibleServerStorage`

The server's storage profile: capacity, provisioned IOPS or elastic
IOPS scaling, auto-grow, and the slow-query-log placement. Omit to
accept Azure's defaults (20 GiB, auto-grow on, default IOPS for the
size).

- rule: iops cannot be set when io_scaling_enabled is true -- elastic scaling manages IOPS itself

### spec.storage.sizeGb

`int32` · optional (explicit presence)

The provisioned storage size in GiB, 20-16384. Unspecified applies
Azure's default of 20. Storage only grows -- shrinking replaces the
server.

- rule: {"int32":{"lte":16384,"gte":20}}

### spec.storage.iops

`int32` · optional (explicit presence)

Provisioned IOPS, 360-48000 (bounded by the SKU and storage size).
Unspecified applies Azure's default for the storage size. Cannot be
combined with io_scaling_enabled -- elastic scaling manages IOPS
itself.

- rule: {"int32":{"lte":48000,"gte":360}}

### spec.storage.autoGrowEnabled

`bool` · optional (explicit presence)

Whether storage grows automatically when free space runs low,
without downtime. Azure's default for MySQL is true (note: the
opposite of PostgreSQL Flexible Server's default).

- default: `true`

### spec.storage.ioScalingEnabled

`bool`

Elastic IOPS scaling: Azure scales IOPS up and down automatically
with workload demand instead of holding a provisioned value. Azure's
default is false. Mutually exclusive with iops.

### spec.storage.logOnDiskEnabled

`bool`

Whether the slow query log is written to the server's disk (counted
against storage) instead of Azure's default log pipeline. Azure's
default is false; enable only when a compliance regime requires
on-disk logs.

### spec.zone

`string`

The availability zone for the primary server ("1", "2", or "3").
Unset lets Azure choose. After creation the zone can only change via
a planned failover that swaps primary and standby -- Azure rejects an
independent zone change.

- rule: zone must be "1", "2", or "3"

### spec.highAvailability

`AzureMysqlFlexibleServerHighAvailability`

High availability: presence enables a standby server with synchronous
replication and automatic failover. Omit for a single-instance
server. Burstable SKUs and replica servers do not support HA.

### spec.highAvailability.mode

`enum` · required

ZONE_REDUNDANT places the standby in a different availability zone
(survives zone failure -- the production recommendation); SAME_ZONE
co-locates it (faster failover, no zone-level protection).

- rule: {"required":true}

Allowed values (use exactly as shown):

- `azure_mysql_flexible_server_high_availability_mode_unspecified` -- Not specified -- invalid; choose an explicit mode when the HA block is present.
- `ZONE_REDUNDANT` -- Standby in a different availability zone: survives zone failure. The production recommendation.
- `SAME_ZONE` -- Standby in the same zone: faster failover, no zone-level protection.

### spec.highAvailability.standbyAvailabilityZone

`string`

The availability zone for the standby ("1", "2", or "3"). Unset lets
Azure choose. For ZONE_REDUNDANT it must differ from the primary's
zone. After creation, zones only change via a planned failover that
swaps zone and standby_availability_zone.

- rule: standby_availability_zone must be "1", "2", or "3"

### spec.maintenanceWindow

`AzureMysqlFlexibleServerMaintenanceWindow`

The weekly maintenance window for Azure-managed patching. Omit for a
system-managed window (Azure picks an off-peak slot). Presence pins
patching to the declared day and start time.

### spec.maintenanceWindow.dayOfWeek

`int32` · optional (explicit presence)

Day of the week: 0 (Sunday) through 6 (Saturday).

- rule: {"int32":{"lte":6,"gte":0}}

### spec.maintenanceWindow.startHour

`int32` · optional (explicit presence)

The window's start hour, 0-23 (UTC).

- rule: {"int32":{"lte":23,"gte":0}}

### spec.maintenanceWindow.startMinute

`int32` · optional (explicit presence)

The window's start minute, 0-59.

- rule: {"int32":{"lte":59,"gte":0}}

### spec.backupRetentionDays

`int32` · optional (explicit presence)

How many days automatic backups are retained (the point-in-time
restore horizon). 1-35; unspecified applies Azure's default of 7.

- default: `7`
- rule: {"int32":{"lte":35,"gte":1}}

### spec.geoRedundantBackupEnabled

`bool`

Whether backups are replicated to the paired Azure region, enabling
cross-region GEO_RESTORE for disaster recovery. Azure's default is
false. Fixed at creation.

### spec.publicNetworkAccess

`enum`

Whether the server accepts connections on its public endpoint.
Unspecified lets Azure derive it: ENABLED for a public server,
DISABLED when the server is VNet-injected via delegated_subnet_id
(a VNet-injected server cannot have a public endpoint).

Allowed values (use exactly as shown):

- `azure_mysql_flexible_server_public_network_access_unspecified` -- Not specified: Azure derives the value -- ENABLED for a public server, DISABLED when the server is VNet-injected.
- `ENABLED` -- The server accepts connections on its public endpoint, filtered by firewall_rules.
- `DISABLED` -- The server has no public endpoint (reachable only through VNet injection or private connectivity).

### spec.delegatedSubnetId

`string | valueFrom`

For VNet injection (private access): the ARM ID of a subnet delegated
to Microsoft.DBforMySQL/flexibleServers, with no other resources in
it. Requires private_dns_zone_id. Fixed at creation. Use an
AzureSubnet with the matching delegation.

- references: AzureSubnet (`status.outputs.subnet_id`)
- rule: write as {value: <literal>} or {valueFrom: {kind: AzureSubnet, name: <that resource's name>, fieldPath: status.outputs.subnet_id}} -- a bare string does not parse

### spec.privateDnsZoneId

`string | valueFrom`

For VNet injection: the ARM ID of a private DNS zone (conventionally
ending ".mysql.database.azure.com") where the server registers its
private address, so VNet-connected clients resolve the fqdn to the
private IP. Required whenever delegated_subnet_id is set.

- references: AzurePrivateDnsZone (`status.outputs.zone_id`)
- rule: write as {value: <literal>} or {valueFrom: {kind: AzurePrivateDnsZone, name: <that resource's name>, fieldPath: status.outputs.zone_id}} -- a bare string does not parse

### spec.userAssignedIdentityIds

`[]string | valueFrom`

The user-assigned identities attached to the server, by ARM ID.
MySQL Flexible Server supports user-assigned identities only (no
system-assigned flavor). Required for customer-managed-key encryption
(the unwrapping identity must be attached here) and for the Entra
administrator (aad_administrator.identity_id must be attached here).
Reference AzureUserAssignedIdentity resources so Key Vault grants can
be composed before the server exists.

- references: AzureUserAssignedIdentity (`status.outputs.identity_id`)
- rule: write as {value: <literal>} or {valueFrom: {kind: AzureUserAssignedIdentity, name: <that resource's name>, fieldPath: status.outputs.identity_id}} -- a bare string does not parse

### spec.customerManagedKey

`AzureMysqlFlexibleServerCustomerManagedKey`

Customer-managed-key (CMK) encryption: the server's data is encrypted
with a Key Vault key you own instead of a Microsoft-managed key.
Requires a user-assigned identity (in user_assigned_identity_ids)
that has wrap/unwrap access on the key's vault.

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
Must be one of the identities in user_assigned_identity_ids, with
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

### spec.aadAdministrator

`AzureMysqlFlexibleServerAadAdministrator`

The single Microsoft Entra (Azure AD) administrator of the server --
MySQL Flexible Server supports exactly one (a group can be used to
admit a team). The grant is backed by a user-assigned identity
attached to the server, which Azure uses to validate Entra tokens.

### spec.aadAdministrator.identityId

`string | valueFrom` · required

The user-assigned identity (attached to the server via
user_assigned_identity_ids) that Azure uses to read directory
objects when validating Entra logins, by ARM ID.

- references: AzureUserAssignedIdentity (`status.outputs.identity_id`)
- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: AzureUserAssignedIdentity, name: <that resource's name>, fieldPath: status.outputs.identity_id}} -- a bare string does not parse

### spec.aadAdministrator.login

`string` · required

The administrator principal's display name as it appears in Entra
(e.g. "dba-team@contoso.com" for a user, the group name for a
group). MySQL uses it as the login name for Entra connections.

- rule: {"required":true,"string":{"minLen":"1"}}

### spec.aadAdministrator.objectId

`string | valueFrom` · required

The ID of the Entra principal being granted the administrator role.
For a user or group this is the directory object ID; for a managed
identity MySQL validates tokens against the identity's CLIENT
(application) ID -- which is why the default reference points at an
AzureUserAssignedIdentity's client_id output, not its principal_id.

- references: AzureUserAssignedIdentity (`status.outputs.client_id`)
- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: AzureUserAssignedIdentity, name: <that resource's name>, fieldPath: status.outputs.client_id}} -- a bare string does not parse

### spec.aadAdministrator.tenantId

`string` · optional (explicit presence)

The Entra tenant of the administrator principal. Leave unset to use
the deploying credential's tenant -- the correct value for virtually
every deployment.

- rule: {"string":{"uuid":true}}

### spec.databases

`[]AzureMysqlFlexibleServerDatabase`

Databases to create on the server, each a separate Azure sub-resource
with its own charset/collation. Most applications declare at least
one database here.

### spec.databases[].name

`string` · required

The database name: 1-64 characters, a valid MySQL schema identifier;
unique within the server.

- rule: {"required":true,"string":{"minLen":"1","maxLen":"64"}}

### spec.databases[].charset

`string` · required · optional (explicit presence)

The database character set. Unspecified applies "utf8mb4" (full
Unicode including supplementary characters -- the right choice for
virtually all applications). Other MySQL charsets (e.g. "latin1",
"utf8mb3") are accepted for legacy migrations.

- default: `utf8mb4`
- rule: {"string":{"minLen":"1"}}

### spec.databases[].collation

`string` · required · optional (explicit presence)

The database collation (sort order and string comparison), e.g.
"utf8mb4_0900_ai_ci" (the MySQL 8.0 default), "utf8mb4_unicode_ci"
(for MySQL 5.7), or "utf8mb4_bin".

- default: `utf8mb4_0900_ai_ci`
- rule: {"string":{"minLen":"1"}}

### spec.firewallRules

`[]AzureMysqlFlexibleServerFirewallRule`

Public-endpoint firewall allowlist: each rule admits a contiguous
IPv4 range. Only meaningful while the server has a public endpoint.
The special rule 0.0.0.0-0.0.0.0 admits Azure-internal services only
(not the internet).

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

MySQL server parameters to override (e.g. "max_connections",
"slow_query_log", "require_secure_transport"), by parameter name.
Values are applied as user overrides on Azure's per-SKU defaults;
removing an entry resets the parameter to its default. Static
(non-dynamic) parameters need a server restart to take effect --
Azure applies them but reports "pending restart" until one happens.

- rule: {"map":{"keys":{"string":{"minLen":"1"}},"values":{"string":{"minLen":"1"}}}}

### spec.tags

`map<string, string>`

Free-form tags applied to the server, merged over the Planton-derived
resource tags (organization, environment, resource id); a user tag
with the same key wins. Tags are Azure's governance surface -- Azure
Policy enforces them and Microsoft Cost Management groups by them.

## Validation Rules

- `mysql_source_matches_create_mode`: source_server_id is required for POINT_IN_TIME_RESTORE, REPLICA, and GEO_RESTORE, and must be omitted for a fresh (DEFAULT) server
- `mysql_restore_time_matches_create_mode`: point_in_time_restore_time_in_utc is required for POINT_IN_TIME_RESTORE and must be omitted for other create modes (GEO_RESTORE restores the latest geo-backup and takes no timestamp)
- `mysql_replication_role_replica_only`: replication_role (NONE, replica promotion) is only valid on a server created with create_mode REPLICA
- `mysql_default_mode_requires_credentials`: a fresh (DEFAULT) server requires administrator_login and administrator_password; only replicas and restores may omit them to inherit from the source
- `mysql_sku_required_for_default_mode`: sku_name is required for a fresh (DEFAULT) server; only replicas and restores may omit it to inherit from the source
- `mysql_vnet_injection_requires_private_dns_zone`: delegated_subnet_id (VNet injection) requires private_dns_zone_id
- `mysql_vnet_injection_forbids_public_access`: public_network_access cannot be ENABLED on a VNet-injected server (delegated_subnet_id set); leave it unset and Azure derives DISABLED
- `mysql_cmk_requires_user_assigned_identity`: customer_managed_key requires primary_user_assigned_identity_id and at least one entry in user_assigned_identity_ids
- `mysql_aad_admin_requires_user_assigned_identity`: aad_administrator requires at least one entry in user_assigned_identity_ids (its identity_id must be attached to the server)

## Outputs

Reference an output from another manifest as `valueFrom: {kind: AzureMysqlFlexibleServer, name: <resource-name>, fieldPath: status.outputs.<output>}`.

| Output | Type | Description |
|---|---|---|
| `status.outputs.server_id` | `string` | The Azure Resource Manager ID of the server. Format: /subscriptions/{sub}/resourceGroups/{rg}/providers/Microsoft.DBforMySQL/flexibleServers/{name} Referenced by AzurePrivateEndpoint (private_connection_resource_id) and by replica/restore servers (source_server_id). |
| `status.outputs.server_name` | `string` | The name of the server. |
| `status.outputs.fqdn` | `string` | The server's fully qualified domain name ({name}.mysql.database.azure.com). For a VNet-injected server it resolves to the private address through the private DNS zone. Connection strings take the shape: mysql://{login}:{password}@{fqdn}:3306/{database}?ssl-mode=REQUIRED |
| `status.outputs.administrator_login` | `string` | The administrator login, echoed so applications can construct connection strings without duplicating the value. |
| `status.outputs.database_ids` | `map<string, string>` | The ARM ID of each database declared in the spec, keyed by database name. Example valueFrom fieldPath: status.outputs.database_ids.myapp |
| `status.outputs.replica_capacity` | `int32` | How many read replicas the server can still accept -- Azure computes this from the SKU (burstable SKUs report 0). Useful when composing replica topologies. |

## References

Fields that can point at another resource's outputs:

| Field | Kind | Output |
|---|---|---|
| `spec.resourceGroup` | AzureResourceGroup | `status.outputs.resource_group_name` |
| `spec.sourceServerId` | AzureMysqlFlexibleServer | `status.outputs.server_id` |
| `spec.delegatedSubnetId` | AzureSubnet | `status.outputs.subnet_id` |
| `spec.privateDnsZoneId` | AzurePrivateDnsZone | `status.outputs.zone_id` |
| `spec.userAssignedIdentityIds` | AzureUserAssignedIdentity | `status.outputs.identity_id` |
| `spec.customerManagedKey.keyVaultKeyId` | AzureKeyVaultKey | `status.outputs.versionless_id` |
| `spec.customerManagedKey.primaryUserAssignedIdentityId` | AzureUserAssignedIdentity | `status.outputs.identity_id` |
| `spec.customerManagedKey.geoBackupKeyVaultKeyId` | AzureKeyVaultKey | `status.outputs.versionless_id` |
| `spec.customerManagedKey.geoBackupUserAssignedIdentityId` | AzureUserAssignedIdentity | `status.outputs.identity_id` |
| `spec.aadAdministrator.identityId` | AzureUserAssignedIdentity | `status.outputs.identity_id` |
| `spec.aadAdministrator.objectId` | AzureUserAssignedIdentity | `status.outputs.client_id` |

## Referenced By

Fields on other kinds that can point at this resource:

| Kind | Field | Reads |
|---|---|---|
| AzureDataProtectionBackupInstance | `spec.mysqlFlexibleServer.serverId` | `status.outputs.server_id` |
| AzureMysqlFlexibleServer | `spec.sourceServerId` | `status.outputs.server_id` |

## See Also

- [Overview](../README.md)
