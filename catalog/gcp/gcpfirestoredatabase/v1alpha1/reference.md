# GcpFirestoreDatabase

> Generated from the protobuf schema by `make generate-reference` -- do not
> edit by hand. To change a fact on this page, change the proto field comment
> or validation rule it is derived from, then regenerate.

**apiVersion**: `gcp.planton.dev/v1alpha1`

**Guide**: [GUIDE.md](../GUIDE.md) -- authored operational judgment for this kind: conventions, trade-offs, and what pairs well with it.

GcpFirestoreDatabaseSpec defines the configuration for a Google Cloud
Firestore database.

Firestore is a fully managed, serverless, NoSQL document database. Each
GCP project can have multiple named databases in addition to the default
"(default)" database. Databases are the top-level container for
collections, documents, and indexes.

Important behavioral notes:

  - The database_name, location_id, database_edition, and kms_key_name
    fields are immutable after creation. Changing them requires recreating
    the database.

  - The type field can be changed between FIRESTORE_NATIVE and DATASTORE_MODE
    after creation, but this is a significant operational change.

  - Firestore databases do not support GCP labels. Resource Manager tags
    are supported at create time via resource_manager_tags.

  - The "(default)" database name is special -- it is the primary database
    that Firestore client libraries connect to when no database ID is
    specified. Only one "(default)" database can exist per project.

## Example

```yaml
apiVersion: gcp.planton.dev/v1alpha1
kind: GcpFirestoreDatabase
metadata:
  name: my-test-firestore
spec:
  # GCP project for the database. Replace with your project ID.
  projectId:
    value: my-gcp-project-123

  # A named database beside the project's "(default)" one.
  databaseName: my-test-firestore
  locationId: us-east1
  type: FIRESTORE_NATIVE

  # Keep the database independent of any App Engine app lifecycle.
  appEngineIntegrationMode: DISABLED

  # DELETE (the default) keeps destroys real — the raw provider would
  # ABANDON the database. Production databases pair PREVENT here with
  # deleteProtectionState: DELETE_PROTECTION_ENABLED.
  deletionPolicy: DELETE
```

## Spec Fields

| Path | Type | Required | Default | References |
|---|---|---|---|---|
| `spec.projectId` | `string \| valueFrom` |  |  | GcpProject (`status.outputs.project_id`) |
| `spec.locationId` | `string` | yes |  |  |
| `spec.databaseName` | `string` | yes |  |  |
| `spec.type` | `string` | yes |  |  |
| `spec.concurrencyMode` | `string` |  |  |  |
| `spec.pointInTimeRecoveryEnablement` | `string` |  |  |  |
| `spec.deleteProtectionState` | `string` |  | `DELETE_PROTECTION_DISABLED` |  |
| `spec.databaseEdition` | `string` |  |  |  |
| `spec.kmsKeyName` | `string \| valueFrom` |  |  | GcpKmsKey (`status.outputs.key_id`) |
| `spec.appEngineIntegrationMode` | `string` |  |  |  |
| `spec.firestoreDataAccessMode` | `string` |  |  |  |
| `spec.mongodbCompatibleDataAccessMode` | `string` |  |  |  |
| `spec.realtimeUpdatesMode` | `string` |  |  |  |
| `spec.resourceManagerTags` | `map<string, string>` |  |  |  |
| `spec.deletionPolicy` | `string` |  | `DELETE` |  |

## Field Details

### spec.projectId

`string | valueFrom`

GCP project where the Firestore database will be created.
Can be a literal project ID or a reference to a GcpProject resource.
If omitted, the provider's default project is used.
Immutable: changing the project destroys and recreates the database.

- references: GcpProject (`status.outputs.project_id`)
- rule: write as {value: <literal>} or {valueFrom: {kind: GcpProject, name: <that resource's name>, fieldPath: status.outputs.project_id}} -- a bare string does not parse

### spec.locationId

`string` · required

Location of the Firestore database. This determines where data is stored
and affects latency and availability. Immutable after creation.

Multi-region locations: "nam5" (United States), "eur3" (Europe).
Single-region locations: any supported GCP region (e.g., "us-east1",
"europe-west1").

Multi-region locations provide higher availability but at higher cost
and with slightly higher write latency.

- rule: {"required":true}

### spec.databaseName

`string` · required

Name for the Firestore database. Must be 4-63 characters, start with a
lowercase letter, contain only lowercase letters, digits, and hyphens,
and end with a letter or digit.

The special value "(default)" creates the project's default database.
Only one default database can exist per project. Client libraries connect
to the default database when no database ID is specified.

Immutable after creation.

- rule: database_name must be '(default)' or 4-63 characters: start with a letter, contain only lowercase letters, digits, and hyphens, end with a letter or digit
- rule: {"required":true}

### spec.type

`string` · required

Firestore database type. Determines the data model and API surface.

FIRESTORE_NATIVE: Modern Firestore with real-time listeners, offline
support, and the Firestore client library API. Recommended for new
applications.

DATASTORE_MODE: Legacy Datastore-compatible mode with the Datastore
client library API. Use for existing Datastore applications or workloads
that need Datastore's entity-group transactions.

- rule: type must be FIRESTORE_NATIVE or DATASTORE_MODE
- rule: {"required":true}

### spec.concurrencyMode

`string`

Concurrency control mode for the database. Determines how conflicts
between concurrent reads and writes are resolved.

OPTIMISTIC: Uses optimistic concurrency control. Reads do not block
writes, and writes are validated at commit time. Default for
FIRESTORE_NATIVE databases.

PESSIMISTIC: Uses pessimistic concurrency control. Reads block
concurrent writes to the same data. Default for DATASTORE_MODE
databases.

OPTIMISTIC_WITH_ENTITY_GROUPS: Legacy Datastore mode using entity
group-based transactions. Only valid for DATASTORE_MODE databases.

If not set, GCP applies the default for the chosen database type.

- rule: concurrency_mode must be OPTIMISTIC, PESSIMISTIC, or OPTIMISTIC_WITH_ENTITY_GROUPS

### spec.pointInTimeRecoveryEnablement

`string`

Whether to enable point-in-time recovery (PITR) for this database.

POINT_IN_TIME_RECOVERY_ENABLED: Retains 7 days of version history.
Reads can target any timestamp within the past hour or any 1-minute
snapshot within the past 7 days.

POINT_IN_TIME_RECOVERY_DISABLED: Retains 1 hour of version history.
Reads can target any timestamp within the past hour.

If not set, defaults to POINT_IN_TIME_RECOVERY_DISABLED.

- rule: point_in_time_recovery_enablement must be POINT_IN_TIME_RECOVERY_ENABLED or POINT_IN_TIME_RECOVERY_DISABLED

### spec.deleteProtectionState

`string` · optional (explicit presence)

Delete protection for the database. When enabled, the database cannot
be deleted through any interface (Console, gcloud, API, IaC tools)
until protection is disabled.

Defaults to DELETE_PROTECTION_DISABLED.

- default: `DELETE_PROTECTION_DISABLED`
- rule: delete_protection_state must be DELETE_PROTECTION_ENABLED or DELETE_PROTECTION_DISABLED

### spec.databaseEdition

`string`

Database edition. Determines the feature set and SLA tier.

STANDARD: Default edition suitable for most workloads. Provides
standard Firestore features and SLA.

ENTERPRISE: Enhanced edition with higher availability SLA, advanced
security features, and support for additional data access modes.
Requires type to be FIRESTORE_NATIVE.

Immutable after creation. If not set, defaults to STANDARD.

- rule: database_edition must be STANDARD or ENTERPRISE

### spec.kmsKeyName

`string | valueFrom`

Fully qualified name of the KMS key to use for customer-managed
encryption (CMEK). The key must exist in the same location as the
database. Immutable after creation.

For multi-region databases: nam5 requires a Cloud KMS key in the
"us" multi-region; eur3 requires a key in the "europe" multi-region.

Format: projects/{project}/locations/{location}/keyRings/{ring}/cryptoKeys/{key}

If not set, Google-managed encryption is used (default).

- references: GcpKmsKey (`status.outputs.key_id`)
- rule: write as {value: <literal>} or {valueFrom: {kind: GcpKmsKey, name: <that resource's name>, fieldPath: status.outputs.key_id}} -- a bare string does not parse

### spec.appEngineIntegrationMode

`string`

App Engine integration mode. ENABLED couples the database's
lifecycle to the project's App Engine application (a legacy
coupling: disabling the App Engine app disables the database with
it). DISABLED keeps the database independent — the right choice for
everything that is not a legacy App Engine deployment.

If not set, GCP applies its default.

- rule: app_engine_integration_mode must be ENABLED or DISABLED

### spec.firestoreDataAccessMode

`string`

Firestore API data access mode — whether the classic Firestore API
can read and write this database. ENTERPRISE edition only. Use
DATA_ACCESS_MODE_DISABLED on a database dedicated to the
MongoDB-compatible API to keep access single-protocol. At most one of
the two data-access modes can be ENABLED on a database (the API
rejects both-enabled at create), and realtime updates require THIS
mode to be explicitly ENABLED.

- rule: firestore_data_access_mode must be DATA_ACCESS_MODE_ENABLED or DATA_ACCESS_MODE_DISABLED

### spec.mongodbCompatibleDataAccessMode

`string`

MongoDB-compatible API data access mode — whether MongoDB drivers
and tools can read and write this database. ENTERPRISE edition only.
Pair with MONGODB_COMPATIBLE_API-scoped GcpFirestoreIndex indexes
for query support. Mutually exclusive with an ENABLED
firestore_data_access_mode — a database is single-protocol, so a
MongoDB-dedicated database can never carry realtime updates.

- rule: mongodb_compatible_data_access_mode must be DATA_ACCESS_MODE_ENABLED or DATA_ACCESS_MODE_DISABLED

### spec.realtimeUpdatesMode

`string`

Realtime updates mode — whether clients can subscribe to live query
snapshots on this database. ENTERPRISE edition only. Enabling it
requires firestore_data_access_mode DATA_ACCESS_MODE_ENABLED
(realtime subscriptions ride the classic Firestore API; leaving the
access mode unset does not count as enabled).

- rule: realtime_updates_mode must be REALTIME_UPDATES_MODE_ENABLED or REALTIME_UPDATES_MODE_DISABLED

### spec.resourceManagerTags

`map<string, string>`

Resource Manager tags bound to the database for org-policy and IAM
conditions. Keys in the form "tagKeys/{id}", values "tagValues/{id}".
Create-time only: changing them later replaces the database.

### spec.deletionPolicy

`string` · optional (explicit presence)

Deletion policy — what happens when this resource is destroyed:
  ""        -- same as "DELETE" (the Planton default; note the raw
               provider defaults to ABANDON, which would leave the
               database running unmanaged after a destroy)
  "DELETE"  -- the database is deleted (delete protection, if
               enabled, still blocks it)
  "PREVENT" -- destroy FAILS; belt-and-suspenders beside
               delete_protection_state
  "ABANDON" -- the database is removed from management but keeps
               running in GCP

- default: `DELETE`
- rule: deletion_policy must be one of: DELETE, PREVENT, ABANDON

## Validation Rules

- `enterprise_requires_firestore_native`: database_edition ENTERPRISE requires type to be FIRESTORE_NATIVE
- `firestore_data_access_mode_requires_enterprise`: firestore_data_access_mode can only be set on ENTERPRISE edition databases
- `mongodb_data_access_mode_requires_enterprise`: mongodb_compatible_data_access_mode can only be set on ENTERPRISE edition databases
- `realtime_updates_mode_requires_enterprise`: realtime_updates_mode can only be set on ENTERPRISE edition databases
- `data_access_modes_mutually_exclusive`: only one of firestore_data_access_mode and mongodb_compatible_data_access_mode can be DATA_ACCESS_MODE_ENABLED
- `realtime_updates_requires_firestore_access`: realtime_updates_mode REALTIME_UPDATES_MODE_ENABLED requires firestore_data_access_mode DATA_ACCESS_MODE_ENABLED (realtime subscriptions use the Firestore API; unset does not count as enabled)

## Outputs

Reference an output from another manifest as `valueFrom: {kind: GcpFirestoreDatabase, name: <resource-name>, fieldPath: status.outputs.<output>}`.

| Output | Type | Description |
|---|---|---|
| `status.outputs.database_id` | `string` | Fully qualified database ID. Format: projects/{project}/databases/{database} This is the canonical identifier used for API calls and client library connections. |
| `status.outputs.database_name` | `string` | Database name. This is the value passed during creation (e.g., "(default)" or a custom name) and is used by client libraries to select which database to connect to. |
| `status.outputs.uid` | `string` | Server-generated UUID4 for this database. Unique across all Firestore databases. |
| `status.outputs.create_time` | `string` | Timestamp at which the database was created. RFC3339 UTC format. |
| `status.outputs.earliest_version_time` | `string` | Earliest timestamp at which older versions of data can be read. Determined by the version retention period (1 hour without PITR, 7 days with PITR enabled). Useful for planning point-in-time recovery operations. RFC3339 UTC format. |
| `status.outputs.version_retention_period` | `string` | How long past versions of data are retained (e.g. "3600s" without PITR, "604800s" with PITR enabled) — the window earliest-version reads and PITR restores can target. |
| `status.outputs.key_prefix` | `string` | Key prefix for Datastore Mode app identifiers (the appid of a legacy App Engine application). Empty for databases without one. |
| `status.outputs.update_time` | `string` | Timestamp of the database's last configuration update. RFC3339 UTC format. |

## References

Fields that can point at another resource's outputs:

| Field | Kind | Output |
|---|---|---|
| `spec.projectId` | GcpProject | `status.outputs.project_id` |
| `spec.kmsKeyName` | GcpKmsKey | `status.outputs.key_id` |

## Referenced By

Fields on other kinds that can point at this resource:

| Kind | Field | Reads |
|---|---|---|
| GcpFirestoreBackupSchedule | `spec.database` | `status.outputs.database_name` |
| GcpFirestoreIndex | `spec.database` | `status.outputs.database_name` |

## See Also

- [Overview](../README.md)
