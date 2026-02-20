# OciMysqlDbSystem — Design Documentation

## Design Rationale

### Single-Resource Component

OciMysqlDbSystem creates exactly one OCI resource: `oci_mysql_mysql_db_system`. This follows the OpenMCF principle that each component maps to one primary infrastructure resource. The DB System resource itself includes the primary endpoint, HA replicas (when enabled), and backup configuration as sub-resources managed by the OCI API.

HeatWave cluster, replication channels, and MySQL configurations are intentionally separate components. They have independent lifecycles — a HeatWave cluster can be attached, detached, and resized without affecting the underlying DB System, and a MySQL Configuration can be shared across multiple DB Systems.

### Why HeatWave Cluster Is Not Bundled

OCI treats the HeatWave cluster as a separate resource (`oci_mysql_heat_wave_cluster`) with its own shape, node count, and lifecycle. Bundling it would:

1. Force recreation of the entire DB System when only the HeatWave node count changes
2. Couple the analytics tier lifecycle to the transactional database lifecycle
3. Prevent attaching HeatWave to an existing DB System provisioned outside OpenMCF

By keeping them separate, operators can add or remove HeatWave capability independently and manage costs separately.

### Compute Shape Selection

The `shapeName` field takes a literal OCI shape name (e.g., `MySQL.VM.Standard.E4.1.8GB`). The component does not abstract or validate shape names because:

- OCI shapes change frequently as new hardware becomes available
- Shape availability varies by region and availability domain
- Abstracting shapes would add a mapping layer that drifts from reality

The shape determines CPU, memory, network bandwidth, and which MySQL Configurations are compatible.

## Spec Field Decisions

### Foreign Key References

Six fields use `StringValueOrRef`: `compartmentId`, `subnetId`, `configurationId`, `nsgIds` (repeated), `encryptData.keyId`, and `secureConnections.certificateId`. Each has a `default_kind` and `default_kind_field_path` annotation so that `valueFrom` references can omit the `fieldPath` when referencing the most common output.

For example, `subnetId` defaults to `OciSubnet` / `status.outputs.subnetId`, so a reference only needs `kind` and `name`:

```yaml
subnetId:
  valueFrom:
    kind: OciSubnet
    name: db-subnet
```

### Admin Credentials

`adminUsername` and `adminPassword` are optional strings rather than required fields because:

- The OCI API has provider-level defaults for admin username
- Forcing a password in the proto would mean it must be present in the YAML manifest; real deployments should inject secrets via environment variables or secret references

Both fields trigger recreation when changed — this matches OCI API behavior.

### Data Storage Block vs. Legacy Field

The OCI API has both a legacy `data_storage_size_in_gb` top-level field and a newer `data_storage` block. This component uses only the `data_storage` block because:

- The block supports auto-expand configuration (`isAutoExpandStorageEnabled`, `maxStorageSizeInGbs`)
- The legacy field is deprecated in the OCI Terraform provider
- Using both would create ambiguity about which value wins

### Backup Policy and PITR

Backup configuration is a single nested `backupPolicy` message with an embedded `pitrPolicy`. This mirrors the OCI API structure. PITR requires backups to be enabled — this constraint is documented but not enforced in proto validation because the OCI API returns a clear error if violated.

### Maintenance Window

The `maintenance` message includes `windowStartTime` (required when the block is present), `maintenanceScheduleType`, `versionPreference`, and `versionTrackPreference`. These are all enums in the proto but passed as uppercase strings to the OCI API (the Go module handles the conversion).

The format for `windowStartTime` is `{day-of-week} {time-of-day}` (e.g., `sun 04:00`), matching the OCI API expectation.

### Deletion Policy

`deletionPolicy` maps to the OCI `deletion_policy` block (note: the Pulumi provider uses `deletion_policies` as an array, but this component wraps a single policy into a one-element array internally). The three controls are:

- `automaticBackupRetention` — `DELETE` or `RETAIN` existing backups
- `finalBackup` — `REQUIRE_FINAL_BACKUP` or `SKIP_FINAL_BACKUP`
- `isDeleteProtected` — boolean lock that must be set to `false` before deletion

### Encryption and TLS

Both `encryptData` and `secureConnections` use an enum-plus-optional-OCID pattern:

- `keyGenerationType: system` — Oracle-managed keys, no `keyId` needed
- `keyGenerationType: byok` — customer must provide `keyId`
- `certificateGenerationType: system_cert` — Oracle-managed TLS
- `certificateGenerationType: byoc` — customer must provide `certificateId`

Cross-field validation is enforced at the proto level via CEL expressions:

- `encrypt_data.key_id` is required when `key_generation_type` is `byok`
- `secure_connections.certificate_id` is required when `certificate_generation_type` is `byoc`

Note: `system_cert` maps to `SYSTEM` (not `SYSTEM_CERT`) in the OCI API. The Go module handles this mapping in `buildSecureConnections`.

### Read Endpoint

The `readEndpoint` message enables a separate DNS endpoint for read scaling across HA replicas. It is only meaningful when `isHighlyAvailable` is `true`, but this constraint is not enforced in proto validation — OCI simply ignores the read endpoint on non-HA systems.

### Database Console and REST API

Both `databaseConsole` and `rest` are newer OCI features exposed as optional nested messages. They control the web-based management UI and MySQL Router REST API respectively. Port values are validated by OCI (443 or 1024-65535).

## What Is Deferred to Future Versions

| Feature | Reason |
|---------|--------|
| Source block (BACKUP, PITR, IMPORTURL) | Only fresh creation is supported in v1; restore-from-backup requires additional workflow orchestration |
| `shutdown_type`, `state` | Operational lifecycle controls belong to a separate operations layer, not declarative IaC |
| `access_mode`, `database_mode` | Runtime toggles that change during operations, not at provisioning time |
| `security_attributes` (ZPR) | Zero-trust Packet Routing is a newer OCI feature with limited availability |
| `backup_policy.copy_policies` | Cross-region backup copy adds significant complexity (target region, vault, key) |
| `backup_policy.soft_delete` | Newer feature not yet widely adopted |
| `maintenance.maintenance_disabled_windows` | Advanced scheduling for blocking specific maintenance windows |
| HeatWave cluster | Separate OCI resource with independent lifecycle — planned as a dedicated component |
| Replication channels | Separate OCI resource — planned as a dedicated component |
| MySQL Configuration | Separate OCI resource — allows sharing across DB Systems |

## Trade-Offs

### Single Availability Domain

The DB System is placed in a single availability domain. HA provides fault domain diversity within that AD but not cross-AD redundancy. Cross-AD or cross-region disaster recovery requires replication channels (deferred).

### Password in Manifest

`adminPassword` is a plain string in the spec. This is a known limitation — the proto does not have a secret type. Operators should avoid committing passwords to version control and instead use environment variable substitution or a secrets management integration.

### No Shape Validation

The component does not validate `shapeName` against available shapes. An invalid shape name will produce an OCI API error at deployment time, not at validation time. This is intentional — the set of valid shapes changes frequently.

### Enum String Mapping

Proto enums use lowercase values (`byok`, `system_cert`, `early`, `regular`) while the OCI API expects uppercase (`BYOK`, `SYSTEM`, `EARLY`, `REGULAR`). The Go module handles the conversion with `strings.ToUpper()` and special-cases `system_cert` to `SYSTEM`. This mapping is invisible to the user but must be maintained if new enum values are added.

## Freeform Tags

The component automatically applies the following freeform tags to the DB System:

| Tag | Value |
|-----|-------|
| `resource` | `true` |
| `resource_kind` | `OciMysqlDbSystem` |
| `resource_id` | `metadata.id` |
| `organization` | `metadata.org` (when set) |
| `environment` | `metadata.env` (when set) |

Additional tags from `metadata.labels` are merged into the freeform tags map.
