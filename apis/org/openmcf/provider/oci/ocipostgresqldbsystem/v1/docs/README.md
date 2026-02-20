# OciPostgresqlDbSystem — Design Documentation

Internal design rationale, trade-offs, and deferred scope for the OciPostgresqlDbSystem component (v1).

## Design Rationale

### Why a Dedicated Component

OCI's PostgreSQL DB System (`oci_psql_db_system`) is a standalone managed service with its own API surface, distinct from the Oracle Base Database Service (`oci_database_db_system`) and MySQL DB System (`oci_mysql_mysql_db_system`). Giving it a dedicated OpenMCF kind allows:

- A schema tailored to PostgreSQL-specific concepts (DB version, PSQL configurations, backup kinds)
- Clear separation from other database components that share naming patterns but have different fields
- Independent lifecycle management — PostgreSQL DB Systems can be created, scaled, and deleted without affecting other database resources

### Credential Management via Discriminator Pattern

The `PasswordDetails` message uses a `passwordType` enum as a discriminator rather than a protobuf `oneof`. This was chosen because:

- `oneof` fields serialize differently in JSON (only one field present), which complicates YAML authoring where users expect to see the discriminator and the corresponding field together
- CEL validation rules enforce the correct pairing (`plain_text` requires `password`; `vault_secret` requires `secretId`) at the proto level
- Both approaches are functionally equivalent; the enum pattern is more explicit in YAML manifests

### Storage System Type Hardcoded

The `systemType` field on `StorageDetails` is hardcoded to `OCI_OPTIMIZED_STORAGE` in the IaC module rather than exposed in the spec. As of the current OCI API, this is the only valid value. Exposing it would add a required field with a single valid option, adding complexity with no benefit. If OCI introduces additional storage types, a field can be added in a future version.

### Shape Auto-Prefixing

The IaC module does not auto-prefix the shape string — the user provides the full shape name (e.g. "VM.Standard.E4.Flex"). The proto comment mentions that the provider prefixes "PostgreSQL." if not present, which refers to OCI API behavior, not the OpenMCF module. The module passes the shape string directly to the Pulumi resource.

### Display Name Fallback

When `displayName` is empty, the module falls back to `metadata.name`. This is implemented in `locals.go` rather than at the proto level (via a default value) because proto3 does not support non-zero defaults for string fields. The fallback is a module-level concern.

### Freeform Tags from Metadata

The module automatically constructs freeform tags from metadata fields:

| Tag Key | Source |
|---------|--------|
| `resource` | Always `"true"` |
| `resource_kind` | `CloudResourceKind_OciPostgresqlDbSystem` enum string |
| `resource_id` | `metadata.id` |
| `organization` | `metadata.org` (when non-empty) |
| `environment` | `metadata.env` (when non-empty) |
| Custom labels | All entries from `metadata.labels` |

Users cannot override the automatic tags, but can add additional tags via `metadata.labels`.

## Trade-offs

### Immutable Credentials

The entire `credentials` block is immutable after creation. Changing the username or password details forces recreation of the DB System, which causes downtime. This matches OCI's API behavior — the PostgreSQL service does not support in-place credential rotation at the infrastructure level. Password rotation should be handled at the application or Vault level.

### No Backup Restore (Source Block)

v1 only supports fresh creation. The OCI API's `source` block (for restoring from a backup OCID) is excluded because:

- Restore is a point-in-time operational concern, not a steady-state declaration
- Including it would complicate the spec with a one-time-use field that becomes irrelevant after initial creation
- Restore workflows can be handled via OCI Console or CLI as a separate operational procedure

### No Patch Operations

The `patch_operations` field (for adding/removing read replicas without recreating the system) is excluded because:

- It is an operational action, not a declarative specification
- The `instanceCount` field handles the initial replica count
- Replica management post-creation is deferred to operational tooling

### No Apply Config

The `apply_config` field (controlling how configuration changes are applied — restart vs. reload) is excluded because:

- It is a deployment-time behavioral flag, not a resource specification
- The default OCI behavior (restart when needed) is acceptable for most use cases

### Per-Instance Details Are Immutable

`instancesDetails` entries cannot be modified after creation. This is an OCI API constraint — the per-node IP assignments and identifiers are fixed at provisioning time. To change them, the DB System must be recreated.

## Deferred to Future Versions

### Source Block (Backup Restore)

Restoring a DB System from a backup OCID. Would require a `source` message with `backupId`, `isHavingRestoreConfigOverrides`, and `sourceType` fields. Deferred because fresh creation covers the primary use case and restore is an operational workflow.

### Cross-Region Backup Copy

The `backupPolicy.copyPolicy` field for replicating backups to another region. Deferred because it requires cross-region configuration that adds significant complexity and is only needed for disaster recovery scenarios.

### Defined Tags and System Tags

OCI defined tags (namespace-scoped) and system tags (Oracle-managed) are excluded. The platform manages tagging via freeform tags derived from metadata. If tag governance requirements emerge, defined tags can be added in a future version.

### Configuration Management Component

Server parameter tuning (shared_buffers, max_connections, etc.) is referenced via `configId` but the PostgreSQL Configuration resource itself is not managed by this component. A separate `OciPostgresqlConfiguration` component could be added to manage configurations declaratively.

### Connection Pooling

PgBouncer or OCI-native connection pooling is not part of this component. Connection pooling can be deployed separately as an application-level concern.
