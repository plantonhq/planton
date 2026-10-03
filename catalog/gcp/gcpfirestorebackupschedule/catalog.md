# GCP Firestore Backup Schedule

Declares a managed, periodic backup schedule for a Cloud Firestore database — Firestore's own backup facility with retention, distinct from point-in-time recovery. PITR covers the last 7 days of versions inside the live database; scheduled backups extend protection to 14 weeks and survive even the database's deletion. A database supports one daily and one weekly schedule, and the classic production posture is both: short-retention daily backups beside a long-retention weekly one — exactly two of these resources on the same database.

## What Gets Created

When you deploy this Infra Component, the IaC module provisions:

- **Firestore Backup Schedule** -- a daily or weekly schedule attached to the target database; Firestore takes each backup automatically (you pin only the weekly day — timing within the day is Firestore's)
- **Backups over time** -- each run produces a backup kept for the retention window; backups already taken OUTLIVE the schedule and age out per their retention
- **Firestore API enablement** -- `firestore.googleapis.com` enabled in the target project (never disabled on destroy)

The recurrence (daily vs weekly, and the weekly day) is immutable after creation. Retention is the spec's only mutable field — extend or shorten protection in place.

## Before You Deploy

### Planton Setup

- **GCP Provider Connection** -- an active connection in the Connect module with credentials for the target GCP project. Map it as the default for your environment, or specify it explicitly when creating the Infra Component.
- **Planton Runner** -- required when using Runner-based credential delivery. Not needed for inline credentials or browser OAuth authentication modes.

### GCP Project

- **A Firestore database** in the target project — reference a `GcpFirestoreDatabase` Infra Component via ValueFromRef, or use `"(default)"` for the project's primary database.

## Deploy

### Console

Open the deployment store, find **GCP Firestore Backup Schedule**, and click **Deploy**. The creation wizard walks you through preset selection, environment and connection configuration, and spec fields. Start from the **Daily Short-Retention Backups** preset in the [Presets](#presets) tab for the bread-and-butter daily protection.

### CLI

Create a manifest and apply it:

```yaml
apiVersion: gcp.planton.dev/v1alpha1
kind: GcpFirestoreBackupSchedule
metadata:
  name: daily-backups
  org: acme-corp
  env: prod
spec:
  database:
    value: "(default)"
  retention: 604800s
  daily: true
```

```shell
planton apply -f firestore-backup-schedule.yaml
```

This schedules daily backups of the default database, each kept for 7 days. An Infra Job tracks the provisioning in real time.

### InfraChart

When deploying as part of a multi-resource environment, use ValueFromRef to wire the schedule to a Firestore database deployed in the same InfraPipeline:

```yaml
spec:
  database:
    valueFrom:
      kind: GcpFirestoreDatabase
      name: orders-database
      fieldPath: status.outputs.database_name
  retention: 8467200s
  weeklyRecurrence:
    day: SUNDAY
```

The InfraPipeline resolves the dependency graph, deploys the database first, then attaches its backup schedules.

## Key Configuration

These are the most important decisions when configuring a backup schedule. Explore the full field reference in the [API Explorer](#api-explorer) tab.

**Recurrence** -- Exactly one of `daily: true` or `weeklyRecurrence.day` (MONDAY through SUNDAY). Immutable after creation — Firestore chooses the time of day; only weekly schedules pin a day.

**Retention** -- A whole-seconds duration string ending in `s` (`604800s` = 7 days; `8467200s` = the 14-week maximum). The only mutable field: extend or shorten each backup's lifetime in place.

**Backups vs PITR** -- Point-in-time recovery (a `GcpFirestoreDatabase` setting) covers the last 7 days of versions for in-place reads and restores. Backups restore into a NEW database (`gcloud firestore databases restore`) and reach back 14 weeks — enable both for production data.

## Outputs and Dependencies

### What This Kind Consumes

| Dependency | Field | ValueFromRef Path |
|------------|-------|-------------------|
| **GcpProject** (optional) | `projectId` | `status.outputs.project_id` |
| **GcpFirestoreDatabase** | `database` | `status.outputs.database_name` |

### What This Kind Provides

After provisioning, `status.outputs` contains values that downstream Infra Components can consume via ValueFromRef:

| Output | Description | Common Downstream Use |
|--------|-------------|----------------------|
| `schedule_id` | Server-assigned schedule identifier within the database | Firestore Admin API calls addressing the schedule |
| `database` | The database the schedule protects | Confirms the attachment without dereferencing the spec |

## Common Patterns

Browse the [Presets](#presets) tab for ready-to-deploy configurations.

**Daily Short Retention** -- Daily backups with one-week retention: cheap, recent restore points for everyday protection. Start from the **Daily Short-Retention Backups** preset.

**Weekly Long Retention** -- Weekly Sunday backups at the 14-week maximum retention: the compliance/long-horizon layer, typically deployed BESIDE a daily schedule on the same database. Start from the **Weekly Long-Retention Archive** preset.

## Works With

- [**GCP Firestore Database**](/infra-catalog/gcp-firestore-database) -- provides the database the schedule protects
- [**GCP Project**](/infra-catalog/gcp-project) -- provides the GCP project when it differs from the connection default
- [**GCP Firestore Index**](/infra-catalog/gcp-firestore-index) -- the same database's query layer, declared side by side
