# GcpSpannerBackupSchedule

> Generated from the protobuf schema by `make generate-reference` -- do not
> edit by hand. To change a fact on this page, change the proto field comment
> or validation rule it is derived from, then regenerate.

**apiVersion**: `gcp.planton.dev/v1alpha1`

**Guide**: [GUIDE.md](../GUIDE.md) -- authored operational judgment for this kind: conventions, trade-offs, and what pairs well with it.

GcpSpannerBackupScheduleSpec defines a backup schedule
(`google_spanner_backup_schedule`) on a Cloud Spanner database.

A backup schedule creates backups of one database on a cron cadence and
retains each backup for retention_duration. Schedules are first-class,
many-per-database resources: a database commonly carries a daily
incremental schedule AND a weekly full schedule side by side.

Important behavioral notes:

  - The schedule is OWNED BY the database: deleting the database deletes
    its schedules. The BACKUPS themselves survive — they live until
    their retention expires and are what makes destroying the instance
    fail unless GcpSpannerInstance.force_destroy is set.

  - schedule_name, instance, database, and backup_type are immutable;
    cron, retention, and encryption update in place.

  - INCREMENTAL schedules build chains where each backup stores only
    changes since the previous one (cheaper storage, same restore
    semantics) and require the instance to be ENTERPRISE or
    ENTERPRISE_PLUS edition.

## Example

```yaml
apiVersion: gcp.planton.dev/v1alpha1
kind: GcpSpannerBackupSchedule
metadata:
  name: orders-daily-backups
spec:
  # project_id omitted: falls back to the provider's default project.
  instance:
    value: orders-spanner
  database:
    value: orders
  cron: "0 2 * * *" # daily at 02:00 UTC
  retentionDuration: 2678400s # 31 days
  # Explicit DELETE (the provider default) proves the round-trip; the
  # backups already taken outlive the schedule either way.
  deletionPolicy: DELETE
```

## Spec Fields

| Path | Type | Required | Default | References |
|---|---|---|---|---|
| `spec.projectId` | `string \| valueFrom` |  |  | GcpProject (`status.outputs.project_id`) |
| `spec.instance` | `string \| valueFrom` | yes |  | GcpSpannerInstance (`status.outputs.instance_name`) |
| `spec.database` | `string \| valueFrom` | yes |  | GcpSpannerDatabase (`status.outputs.database_name`) |
| `spec.scheduleName` | `string` |  |  |  |
| `spec.cron` | `string` | yes |  |  |
| `spec.retentionDuration` | `string` | yes |  |  |
| `spec.backupType` | `string` |  | `FULL` |  |
| `spec.encryptionConfig` | `GcpSpannerBackupScheduleEncryptionConfig` |  |  |  |
| `spec.encryptionConfig.encryptionType` | `string` | yes |  |  |
| `spec.encryptionConfig.kmsKeyName` | `string \| valueFrom` |  |  | GcpKmsKey (`status.outputs.key_id`) |
| `spec.encryptionConfig.kmsKeyNames` | `[]string \| valueFrom` |  |  | GcpKmsKey (`status.outputs.key_id`) |
| `spec.deletionPolicy` | `string` |  |  |  |

## Field Details

### spec.projectId

`string | valueFrom`

The GCP project that owns the parent Spanner instance. Accepts a
literal project ID or a reference to a GcpProject resource. If
omitted, the provider's default project is used.
Immutable: changing the project destroys and recreates the schedule.

- references: GcpProject (`status.outputs.project_id`)
- rule: write as {value: <literal>} or {valueFrom: {kind: GcpProject, name: <that resource's name>, fieldPath: status.outputs.project_id}} -- a bare string does not parse

### spec.instance

`string | valueFrom` · required

The Spanner instance hosting the database. Immutable.

- references: GcpSpannerInstance (`status.outputs.instance_name`)
- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: GcpSpannerInstance, name: <that resource's name>, fieldPath: status.outputs.instance_name}} -- a bare string does not parse

### spec.database

`string | valueFrom` · required

The database this schedule backs up. Immutable.

- references: GcpSpannerDatabase (`status.outputs.database_name`)
- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: GcpSpannerDatabase, name: <that resource's name>, fieldPath: status.outputs.database_name}} -- a bare string does not parse

### spec.scheduleName

`string`

Unique name of the backup schedule within the database. Immutable.
If not specified, defaults to metadata.name. Must start with a
lowercase letter, contain lowercase letters, digits, and hyphens, and
end with a letter or digit.

- rule: {"ignore":"IGNORE_IF_ZERO_VALUE","string":{"maxLen":"63","pattern":"^[a-z][-a-z0-9]*[a-z0-9]$"}}

### spec.cron

`string` · required

Crontab expression for when backups are created, evaluated in UTC
(maps to the provider's spec.cron_spec.text). Spanner accepts a
bounded set of frequencies — every 12 hours, daily, weekly, or
monthly. Examples:
  "0 2/12 * * *" — every 12 hours at 02:00 and 14:00 UTC
  "0 2 * * *"    — daily at 02:00 UTC
  "0 2 * * 0"    — weekly on Sunday at 02:00 UTC
  "0 2 8 * *"    — monthly on the 8th at 02:00 UTC
Mutable — cadence changes apply in place.

- rule: {"required":true}

### spec.retentionDuration

`string` · required

How long each backup is retained, as a seconds duration string ending
in 's' (e.g. "86400s" = 1 day, "2678400s" = 31 days). Maximum 366 days
("31622400s"). Mutable — applies to backups created AFTER the change.

- rule: {"required":true,"string":{"pattern":"^[0-9]+(\\.[0-9]{1,9})?s$"}}

### spec.backupType

`string` · optional (explicit presence)

The kind of backups the schedule creates. Immutable.

FULL (default): every backup is a complete, self-contained copy.

INCREMENTAL: backups form chains storing only changes since the
previous backup — significantly cheaper storage at the same restore
semantics. Requires the instance to be ENTERPRISE or ENTERPRISE_PLUS
edition.

- default: `FULL`
- rule: backup_type must be FULL or INCREMENTAL

### spec.encryptionConfig

`GcpSpannerBackupScheduleEncryptionConfig`

How the backups are encrypted. If omitted, backups use
USE_DATABASE_ENCRYPTION (inherit the database's posture). Mutable.

- rule: CUSTOMER_MANAGED_ENCRYPTION requires exactly one of kms_key_name or kms_key_names
- rule: kms_key_name/kms_key_names are only valid with CUSTOMER_MANAGED_ENCRYPTION

### spec.encryptionConfig.encryptionType

`string` · required

How backups are encrypted. Required when this message is present.

- rule: encryption_type must be USE_DATABASE_ENCRYPTION, GOOGLE_DEFAULT_ENCRYPTION, or CUSTOMER_MANAGED_ENCRYPTION
- rule: {"required":true}

### spec.encryptionConfig.kmsKeyName

`string | valueFrom`

Fully qualified KMS key for single-region backup CMEK.
Format: projects/{project}/locations/{location}/keyRings/{ring}/cryptoKeys/{key}
Only valid with CUSTOMER_MANAGED_ENCRYPTION.

- references: GcpKmsKey (`status.outputs.key_id`)
- rule: write as {value: <literal>} or {valueFrom: {kind: GcpKmsKey, name: <that resource's name>, fieldPath: status.outputs.key_id}} -- a bare string does not parse

### spec.encryptionConfig.kmsKeyNames

`[]string | valueFrom`

Fully qualified KMS keys for multi-region backup CMEK — one key per
region of the instance's multi-region configuration.
Only valid with CUSTOMER_MANAGED_ENCRYPTION.

- references: GcpKmsKey (`status.outputs.key_id`)
- rule: write as {value: <literal>} or {valueFrom: {kind: GcpKmsKey, name: <that resource's name>, fieldPath: status.outputs.key_id}} -- a bare string does not parse

### spec.deletionPolicy

`string`

Deletion policy for the schedule — what happens when this resource
is destroyed. The schedule is a control-plane object: none of these
values touches the BACKUPS it already created (those live until
their retention expires):
  ""        -- same as "DELETE" (provider default)
  "DELETE"  -- the schedule is deleted; no further backups are taken
  "PREVENT" -- destroy FAILS; protects the cadence a recovery
               objective depends on from riding along with a
               stack teardown
  "ABANDON" -- the schedule is removed from management but keeps
               running (and creating backups) in GCP

- rule: deletion_policy must be one of: DELETE, PREVENT, ABANDON

## Outputs

Reference an output from another manifest as `valueFrom: {kind: GcpSpannerBackupSchedule, name: <resource-name>, fieldPath: status.outputs.<output>}`.

| Output | Type | Description |
|---|---|---|
| `status.outputs.schedule_id` | `string` | Fully qualified backup schedule ID. Format: projects/{project}/instances/{instance}/databases/{database}/backupSchedules/{name} This is the canonical identifier used for API calls. |
| `status.outputs.schedule_name` | `string` | Short backup schedule name within the database. |

## References

Fields that can point at another resource's outputs:

| Field | Kind | Output |
|---|---|---|
| `spec.projectId` | GcpProject | `status.outputs.project_id` |
| `spec.instance` | GcpSpannerInstance | `status.outputs.instance_name` |
| `spec.database` | GcpSpannerDatabase | `status.outputs.database_name` |
| `spec.encryptionConfig.kmsKeyName` | GcpKmsKey | `status.outputs.key_id` |
| `spec.encryptionConfig.kmsKeyNames` | GcpKmsKey | `status.outputs.key_id` |

## See Also

- [Overview](../README.md)
