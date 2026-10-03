# GCP Firestore Backup Schedule

Deploys a Firestore managed backup schedule (`google_firestore_backup_schedule`) on an existing Firestore database — periodic backups with configurable retention, distinct from point-in-time recovery.

## What Gets Created

When you deploy a GcpFirestoreBackupSchedule resource, Planton provisions:

- **Backup Schedule** — a `google_firestore_backup_schedule` on the specified database with daily or weekly recurrence and retention; the Firestore API is enabled automatically

## Prerequisites

- **GCP credentials** configured via environment variables or Planton provider config
- **An existing Firestore database** (deploy via GcpFirestoreDatabase first)
- **IAM permissions** — see [`iac/permissions.yaml`](iac/permissions.yaml) for the least-privilege permission set the deploying principal needs

## Quick Start

Create a file `backup-schedule.yaml`:

```yaml
apiVersion: gcp.planton.dev/v1alpha1
kind: GcpFirestoreBackupSchedule
metadata:
  name: orders-daily-backups
spec:
  database:
    valueFrom:
      kind: GcpFirestoreDatabase
      name: prod-firestore
      fieldPath: status.outputs.database_name
  retention: 604800s
  daily: true
```

Deploy:

```shell
planton apply -f backup-schedule.yaml
```

This schedules daily backups, keeping each for 7 days, in the provider's default project.

## Configuration Reference

### Required Fields

| Field | Type | Description | Validation |
|-------|------|-------------|------------|
| `database` | `StringValueOrRef` | Database to back up. References a GcpFirestoreDatabase's `database_name` output. Immutable. | Required |
| `retention` | `string` | Per-backup retention as a seconds duration (`604800s`); up to 14 weeks (`8467200s`). Mutable. | Required, pattern `^[0-9]+s$` |

### Recurrence (exactly one)

| Field | Type | Description |
|-------|------|-------------|
| `daily` | `bool` | Take a backup every day. Immutable. |
| `weeklyRecurrence.day` | `string` | Weekly backup day (`MONDAY`–`SUNDAY`). Immutable. |

### Optional Fields

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `projectId` | `StringValueOrRef` | provider default | GCP project owning the database. |
| `deletionPolicy` | `string` | `DELETE` | `DELETE`, `PREVENT` (destroy fails — protects a compliance-mandated cadence), or `ABANDON` (unmanage; backups keep being taken). |

## Outputs

| Output | Type | Description |
|--------|------|-------------|
| `schedule_id` | `string` | Server-assigned schedule ID (last path segment of the resource name) |
| `database` | `string` | The database the schedule protects |

## Important Notes

- **Daily-plus-weekly pattern**: a database supports one daily and one weekly schedule — compose two resources on the same database.
- **Recurrence is immutable**: only `retention` updates in place.
- **Backups outlive the schedule**: deleting this resource stops future backups but never deletes existing ones — they age out per their retention.
- **No labels surface**: Firestore backup schedules do not support GCP labels — both engines skip labels identically.

## Related Kinds

- [GcpFirestoreDatabase](/docs/catalog/gcp/gcpfirestoredatabase) — the database this schedule protects
- [GcpProject](/docs/catalog/gcp/gcpproject) — the project the database lives in

## Deployment Methods

### Pulumi (Go)

See [`iac/pulumi/README.md`](iac/pulumi/README.md).

### Terraform

See [`iac/tf/README.md`](iac/tf/README.md).

For copy-paste ready manifests, see the [presets](presets/).

---

© Planton. Licensed under [Apache-2.0](https://github.com/plantonhq/planton/blob/main/LICENSE).
