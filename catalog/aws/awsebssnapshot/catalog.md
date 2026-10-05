# AWS EBS Snapshot

Deploys one EBS snapshot from exactly one of three sources: a live volume, an existing snapshot (copied same-region or cross-region, optionally re-encrypted under a different key), or a VMDK/VHD/RAW disk image imported from S3 through the VM Import/Export service. Archive tiering, fast snapshot restore per availability zone, and createVolumePermission grants to other accounts are all managed in-line as part of the snapshot. Snapshots are regional: volumes restore from them in any zone of the region.

## What Gets Created

When you deploy this Infra Component, the IaC module provisions:

- **EBS Snapshot** — exactly one of three provider resources per the configured source arm: a volume snapshot (`volumeId`), a snapshot copy (`copyFrom`), or a disk-image import (`importFrom`); all three expose the same downstream surface (id, ARN, owner, size)
- **Fast Snapshot Restore** — one per zone in `fastRestoreAvailabilityZones`, created only when the list is non-empty; billed per zone-hour while enabled
- **createVolumePermission Grants** — one per account in `shareWithAccountIds`, created only when the list is non-empty
- **AWS Tags** — resource metadata tags (organization, environment, resource kind, resource ID) applied automatically for tracking and governance

## Before You Deploy

### Planton Setup

- **AWS Provider Connection** — an active connection in the Connect module with credentials for the target AWS account, including EC2 permissions. Map it as the default for your environment, or specify it explicitly when creating the Infra Component.
- **Planton Runner** — required when using Runner-based credential delivery. Not needed for inline credentials or cross-account trust authentication modes.

### AWS Account

- **The source volume** (volume arm) — in the same region; reference an AwsEbsVolume Infra Component or pass a literal `vol-...` id.
- **The disk image and the vmimport role** (import arm) — the image staged in S3 (or reachable by URL), and the VM Import/Export service role (`vmimport` by convention, or the role named in `importFrom.roleName`) with AWS's documented trust and S3 policy. The service validates the role only when the task runs, so a missing role fails the create after several minutes, not at plan.
- **KMS key sharing** (only when sharing encrypted snapshots) — `shareWithAccountIds` grants createVolumePermission, but an encrypted snapshot is useless to the peer until the KMS key also grants them decrypt.

## Deploy

### Console

Open the deployment store, find **AWS EBS Snapshot**, and click **Deploy**. The creation wizard walks you through preset selection, environment and connection configuration, and spec fields. Start from the **Volume Backup** preset in the [Presets](#presets) tab for the pre-change checkpoint shape.

### CLI

Create a manifest and apply it:

```yaml
apiVersion: aws.planton.dev/v1alpha1
kind: AwsEbsSnapshot
metadata:
  name: db-data-backup
  org: acme-corp
  env: prod
spec:
  region: us-east-1
  volumeId:
    value: vol-0123456789abcdef0
  description: pre-upgrade db-data checkpoint
```

```shell
planton apply -f ebs-snapshot.yaml
```

This snapshots the referenced volume in place, ready to restore from before the upgrade goes sideways. An Infra Job tracks the provisioning in real time.

### InfraChart

When the snapshot deploys alongside its source volume in one chart, wire the volume reference via ValueFromRef:

```yaml
spec:
  region: us-east-1
  volumeId:
    valueFrom:
      kind: AwsEbsVolume
      name: db-data
      fieldPath: status.outputs.volume_id
  description: nightly checkpoint of db-data
```

The InfraPipeline resolves the dependency graph, provisions the volume first, then captures the snapshot from it.

## Key Configuration

These are the most important decisions when configuring an EBS snapshot. Explore the full field reference in the [API Explorer](#api-explorer) tab.

**Pick the source arm deliberately** — `volumeId`, `copyFrom`, and `importFrom` are mutually exclusive and each is fixed for life. The volume arm is the everyday backup; the copy arm is how snapshots cross regions or change keys; the import arm is how a VM image from outside AWS becomes restorable block storage. Only the volume arm is importable at the provider — copies and imports are create-only surface.

**Copying is the ONLY way to re-encrypt** — snapshots never change keys in place. Encrypting an unencrypted snapshot, or rotating to a different KMS key, means a `copyFrom` instance with `encrypted: true` and the target key. Cross-region DR copies re-encrypt under the destination region's key in the same step.

**Archive is for compliance tails, not short-lived backups** — `storageTier: archive` moves the snapshot to cheaper cold storage, but it carries a minimum billing duration and needs a restore step before the next use: archiving something you will delete next week costs more than keeping it standard. The restore dials are one-or-the-other: `temporaryRestoreDays` time-boxes access (the snapshot re-archives itself), `permanentRestore` brings it back for good.

**Fast snapshot restore is a running meter** — FSR bills per snapshot per zone-hour while enabled, whether or not anyone restores. Enable it for the restore-latency-critical window (a migration day, a DR drill), then remove the zones — `fastRestoreAvailabilityZones` makes both directions one edit.

**Sharing encrypted snapshots is two grants, not one** — `shareWithAccountIds` covers createVolumePermission; the KMS key must separately grant the peer decrypt. And the AWS-managed `aws/ebs` key can never be shared — re-encrypt under a customer key first, via the copy arm.

**Deletion is not loss** — snapshots are incremental (each stores only blocks changed since the previous one), and deleting an intermediate snapshot is safe: AWS re-parents the blocks. You pay for the unique block set, not the snapshot count — prune by policy, not by fear.

## Outputs and Dependencies

### What This Kind Consumes

| Dependency | Field | ValueFromRef Path |
|------------|-------|-------------------|
| **AwsEbsVolume** (volume arm) | `volumeId` | `status.outputs.volume_id` |
| **AwsEbsSnapshot** (copy arm) | `copyFrom.sourceSnapshotId` | `status.outputs.snapshot_id` |
| **AwsKmsKey** (optional, copy arm) | `copyFrom.kmsKeyId` | `status.outputs.key_arn` |
| **AwsKmsKey** (optional, import arm) | `importFrom.kmsKeyId` | `status.outputs.key_arn` |

### What This Kind Provides

After provisioning, `status.outputs` contains values that downstream Infra Components can consume via ValueFromRef:

| Output | Description | Common Downstream Use |
|--------|-------------|----------------------|
| `snapshot_id` | The snapshot's id (`snap-...`) | AwsEbsVolume restores (`snapshotId`) and other snapshots' copy arms (`copyFrom.sourceSnapshotId`) |
| `snapshot_arn` | The snapshot's ARN | IAM policy statements scoping snapshot actions |

`owner_id` and `volume_size_gb` are also exported — the owning account and the captured size (for imports, the size AWS derived from the image). They are observability values rather than composition inputs.

## Common Patterns

Browse the [Presets](#presets) tab for ready-to-deploy configurations.

**The pre-change checkpoint** — snapshot the data volume before an upgrade or migration, restore in minutes if it goes sideways. Cheap insurance because incremental snapshots only bill changed blocks. Start from the **Volume Backup** preset.

**Cross-region DR in one resource** — a copy-arm instance in the DR region, re-encrypted under that region's own KMS key (copying is the only way snapshots change keys). Restore volumes from it the day the primary region has a bad day. Start from the **Cross-Region DR Copy** preset.

**Recurring cadence belongs to DLM** — a hand-managed snapshot resource is a point-in-time act. For hourly/daily/weekly backups with retention, tag the volumes and let an AwsDlmLifecyclePolicy create and prune the snapshots; keep this kind for deliberate, named captures.

## Works With

- [**AWS EBS Volume**](/infra-catalog/aws-ebs-volume) — the source the volume arm captures, and the consumer that restores from `snapshot_id`
- [**AWS KMS Key**](/infra-catalog/aws-kms-key) — the key copies and imports encrypt under, wired via `copyFrom.kmsKeyId` or `importFrom.kmsKeyId`
- [**AWS Data Lifecycle Manager Policy**](/infra-catalog/aws-dlm-lifecycle-policy) — the automation for recurring snapshots; this kind covers the deliberate one-off captures DLM does not
