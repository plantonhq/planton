# GcpFirestoreBackupSchedule

> Generated from the protobuf schema by `make generate-reference` -- do not
> edit by hand. To change a fact on this page, change the proto field comment
> or validation rule it is derived from, then regenerate.

**apiVersion**: `gcp.planton.dev/v1alpha1`

**Guide**: [GUIDE.md](../GUIDE.md) -- authored operational judgment for this kind: conventions, trade-offs, and what pairs well with it.

GcpFirestoreBackupScheduleSpec defines a backup schedule for a
Firestore database — Firestore's own managed, periodic backups with
retention, distinct from point-in-time recovery (PITR covers the last
7 days of versions; backups extend protection to 14 weeks and survive
database deletion).

A database supports one daily and one weekly schedule — the
daily-plus-weekly pattern (short-retention daily backups beside
longer-retention weekly ones) is exactly two of these resources on the
same database.

Important behavioral notes:

  - The recurrence (daily or weekly, and the weekly day) is immutable;
    only retention updates in place.
  - Backups already taken OUTLIVE the schedule: deleting this resource
    stops future backups but never deletes existing ones — they age
    out per their retention.
  - Backup timing within the day is chosen by Firestore; only weekly
    schedules pin a day.

## Example

```yaml
apiVersion: gcp.planton.dev/v1alpha1
kind: GcpFirestoreBackupSchedule
metadata:
  name: daily-backups
spec:
  # GCP project owning the database. Replace with your project ID.
  projectId:
    value: my-gcp-project-123

  # Firestore database to back up — the database name.
  database:
    value: my-firestore-db

  # Keep each backup for 7 days.
  retention: 604800s

  # Daily recurrence — exactly one of daily or weeklyRecurrence.
  daily: true

  # PREVENT protects a compliance-mandated cadence from accidental
  # teardown; DELETE (the default) keeps destroys real.
  deletionPolicy: DELETE
```

## Spec Fields

| Path | Type | Required | Default | References |
|---|---|---|---|---|
| `spec.projectId` | `string \| valueFrom` |  |  | GcpProject (`status.outputs.project_id`) |
| `spec.database` | `string \| valueFrom` | yes |  | GcpFirestoreDatabase (`status.outputs.database_name`) |
| `spec.retention` | `string` | yes |  |  |
| `spec.daily` | `bool` |  |  |  |
| `spec.weeklyRecurrence` | `GcpFirestoreBackupScheduleWeeklyRecurrence` |  |  |  |
| `spec.weeklyRecurrence.day` | `string` | yes |  |  |
| `spec.deletionPolicy` | `string` |  |  |  |

## Field Details

### spec.projectId

`string | valueFrom`

GCP project owning the database. Can be a literal project ID or a
reference to a GcpProject resource. If omitted, the provider's
default project is used.
Immutable: changing the project destroys and recreates the schedule.

- references: GcpProject (`status.outputs.project_id`)
- rule: write as {value: <literal>} or {valueFrom: {kind: GcpProject, name: <that resource's name>, fieldPath: status.outputs.project_id}} -- a bare string does not parse

### spec.database

`string | valueFrom` · required

The Firestore database to back up — the database name (a
GcpFirestoreDatabase reference resolves to it). Immutable after
creation.

- references: GcpFirestoreDatabase (`status.outputs.database_name`)
- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: GcpFirestoreDatabase, name: <that resource's name>, fieldPath: status.outputs.database_name}} -- a bare string does not parse

### spec.retention

`string` · required

How long each backup is kept, as a seconds duration string (e.g.
"604800s" for 7 days). Maximum 14 weeks ("8467200s"). The only
mutable field — extend or shorten protection in place.

- rule: {"required":true,"string":{"pattern":"^[0-9]+s$"}}

### spec.daily

`bool`

Take a backup every day. Exactly one of daily or weekly_recurrence
must be set. Immutable after creation.

### spec.weeklyRecurrence

`GcpFirestoreBackupScheduleWeeklyRecurrence`

Take a backup every week on the given day. Exactly one of daily or
weekly_recurrence must be set. Immutable after creation.

### spec.weeklyRecurrence.day

`string` · required

Day of the week the weekly backup runs.

- rule: {"required":true,"string":{"in":["MONDAY","TUESDAY","WEDNESDAY","THURSDAY","FRIDAY","SATURDAY","SUNDAY"]}}

### spec.deletionPolicy

`string`

Deletion policy — what happens when this resource is destroyed:
  ""        -- same as "DELETE" (provider default)
  "DELETE"  -- the schedule is deleted (backups already taken
               outlive it either way, aging out per retention)
  "PREVENT" -- destroy FAILS; protects a compliance-mandated backup
               cadence from accidental teardown
  "ABANDON" -- the schedule is removed from management but keeps
               taking backups in GCP

- rule: deletion_policy must be one of: DELETE, PREVENT, ABANDON

## Validation Rules

- `exactly_one_recurrence`: set exactly one recurrence: daily true, or a weekly_recurrence day

## Outputs

Reference an output from another manifest as `valueFrom: {kind: GcpFirestoreBackupSchedule, name: <resource-name>, fieldPath: status.outputs.<output>}`.

| Output | Type | Description |
|---|---|---|
| `status.outputs.schedule_id` | `string` | Server-assigned schedule ID (the last path segment of the schedule's resource name) — what Admin API calls address the schedule by. |
| `status.outputs.database` | `string` | The database the schedule protects — confirms the parent without chasing the reference chain. |

## References

Fields that can point at another resource's outputs:

| Field | Kind | Output |
|---|---|---|
| `spec.projectId` | GcpProject | `status.outputs.project_id` |
| `spec.database` | GcpFirestoreDatabase | `status.outputs.database_name` |

## See Also

- [Overview](../README.md)
