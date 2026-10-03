# GCP Spanner Backup Schedule

Creates backups of a Cloud Spanner database on a cron cadence and retains each backup for a configurable window. Backup schedules are first-class, many-per-database resources — a production database commonly carries a daily incremental schedule and a weekly full schedule side by side. The schedule is owned by its database, but the backups it creates outlive both the schedule and the database, surviving until their retention expires.

## What Gets Created

When you deploy this Infra Component, the IaC module provisions:

- **Spanner Backup Schedule** -- a `google_spanner_backup_schedule` on the target database that creates backups at the configured cron cadence (evaluated in UTC) and retains each backup for `retentionDuration`; `backupType` selects `FULL` (the default — every backup is a complete, self-contained copy) or `INCREMENTAL` (chains storing only changes since the previous backup — cheaper storage at identical restore semantics, requiring an ENTERPRISE or ENTERPRISE_PLUS instance)
- **Backup Encryption configuration** -- created only when `encryptionConfig` is set. Omitted entirely, backups inherit the database's own posture (a CMEK database gets CMEK backups). `CUSTOMER_MANAGED_ENCRYPTION` takes exactly one key shape: `kmsKeyName` (regional instance configurations) or `kmsKeyNames` (one key per region of a multi-region configuration)
- **Spanner API enablement** -- `spanner.googleapis.com` enabled in the target project so a fresh project can host backup schedules (never disabled on destroy)

## Before You Deploy

### Planton Setup

- **GCP Provider Connection** -- an active connection in the Connect module with credentials for the target GCP project. Map it as the default for your environment, or specify it explicitly when creating the Infra Component.
- **Planton Runner** -- required when using Runner-based credential delivery. Not needed for inline credentials or browser OAuth authentication modes.

### GCP Project

- **An existing Spanner instance and database** to attach the schedule to. Reference GcpSpannerInstance and GcpSpannerDatabase Infra Components via ValueFromRef, or provide the names directly.
- **ENTERPRISE edition or above** on the instance when using `INCREMENTAL` backups.
- **Cloud KMS key(s)** (if using CMEK) -- each key must live in the location its region requires, and the Spanner service account needs `roles/cloudkms.cryptoKeyEncrypterDecrypter` on every key.

## Deploy

### Console

Open the deployment store, find **GCP Spanner Backup Schedule**, and click **Deploy**. The creation wizard walks you through preset selection, environment and connection configuration, and spec fields. Start from the **Daily Full Backups** preset in the [Presets](#presets) tab to pre-populate a working configuration.

### CLI

Create a manifest and apply it:

```yaml
apiVersion: gcp.planton.dev/v1alpha1
kind: GcpSpannerBackupSchedule
metadata:
  name: daily-full-backups
  org: acme-corp
  env: prod
spec:
  instance:
    value: "app-spanner-prod"
  database:
    value: "app-db"
  cron: "0 2 * * *"
  retentionDuration: "2678400s"
```

```shell
planton apply -f spanner-backup-schedule.yaml
```

This creates a schedule taking a full backup daily at 02:00 UTC, keeping each backup for 31 days, with backups inheriting the database's encryption posture. An Infra Job tracks the provisioning in real time.

### InfraChart

When deploying as part of a multi-resource environment, use ValueFromRef to wire the schedule to the instance and database deployed in the same InfraPipeline:

```yaml
spec:
  instance:
    valueFrom:
      kind: GcpSpannerInstance
      name: app-spanner
      fieldPath: status.outputs.instance_name
  database:
    valueFrom:
      kind: GcpSpannerDatabase
      name: app-db
      fieldPath: status.outputs.database_name
  cron: "0 2 * * *"
  retentionDuration: "2678400s"
```

The InfraPipeline resolves the dependency graph, deploys the instance and database first, then attaches the schedule.

## Key Configuration

These are the most important decisions when configuring a backup schedule. Explore the full field reference in the [API Explorer](#api-explorer) tab.

**Cadence** -- `cron` is a standard crontab expression evaluated in UTC. Spanner accepts a bounded set of frequencies: every 12 hours (`0 2/12 * * *`), daily (`0 2 * * *`), weekly (`0 2 * * 0`), or monthly (`0 2 8 * *`). Mutable — cadence changes apply in place.

**Retention** -- `retentionDuration` is a seconds duration string ending in `s` (e.g. `"86400s"` = 1 day, `"2678400s"` = 31 days), with a 366-day ceiling (`"31622400s"`). Mutable — but changes apply only to backups created AFTER the change; existing backups keep their original expiry.

**Backup type** -- `backupType` is immutable and defaults to `FULL` when unset. `INCREMENTAL` chains cut storage costs significantly on large, frequently-backed-up databases, and require an ENTERPRISE-edition instance.

**Ownership and survival** -- the schedule is owned by its database: deleting the database deletes the schedule. The backups themselves survive until their retention expires — and they are what makes destroying the instance fail unless `GcpSpannerInstance.force_destroy` is set.

## Outputs and Dependencies

### What This Kind Consumes

| Dependency | Field | ValueFromRef Path |
|------------|-------|-------------------|
| **GcpProject** (optional) | `projectId` | `status.outputs.project_id` |
| **GcpSpannerInstance** | `instance` | `status.outputs.instance_name` |
| **GcpSpannerDatabase** | `database` | `status.outputs.database_name` |
| **GcpKmsKey** (optional) | `encryptionConfig.kmsKeyName` / `encryptionConfig.kmsKeyNames` | `status.outputs.key_id` |

### What This Kind Provides

After provisioning, `status.outputs` contains values that downstream Infra Components can consume via ValueFromRef:

| Output | Description | Common Downstream Use |
|--------|-------------|----------------------|
| `schedule_id` | Fully qualified schedule ID (`projects/{p}/instances/{i}/databases/{d}/backupSchedules/{s}`) | Direct Spanner Admin API calls, listing the backups the schedule created |
| `schedule_name` | Short schedule name within the database | Display, operational runbooks |

## Common Patterns

Browse the [Presets](#presets) tab for ready-to-deploy configurations.

**Daily full backups** -- a full backup every day at 02:00 UTC, kept 31 days. The simplest production posture. Start from the **Daily Full Backups** preset.

**Incremental enterprise** -- backups every 12 hours forming incremental chains, kept 14 days. Cheap, frequent restore points for large databases on ENTERPRISE instances. Start from the **Incremental Backups (Enterprise)** preset.

**Weekly long retention** -- a full backup every Sunday, kept 366 days (the maximum), encrypted with a customer-managed key. The compliance-archive posture. Start from the **Weekly Long-Retention Archive** preset.

## Works With

- [**GCP Spanner Instance**](/infra-catalog/gcp-spanner-instance) -- hosts the database and gates incremental backups by edition
- [**GCP Spanner Database**](/infra-catalog/gcp-spanner-database) -- the database this schedule protects
- [**GCP KMS Key**](/infra-catalog/gcp-kms-key) -- provides customer-managed keys for backup-level CMEK encryption
