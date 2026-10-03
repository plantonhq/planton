# AWS EBS Volume

Deploys one EBS volume as its own resource — sized, typed, and encrypted as declared — so the disk outlives whatever instance happens to mount it today. The volume is created fresh in a chosen availability zone (empty, or restored from a snapshot) or as a copy of an existing volume, and its instance attachments are managed in-line: each entry attaches the volume to one instance at one device name, and removing the entry detaches it.

## What Gets Created

When you deploy this Infra Component, the IaC module provisions:

- **EBS Volume** — exactly one of two provider resources per the configured arm: a fresh volume in `availabilityZone` (optionally restored from `snapshotId`), or a volume copy (`copyFrom`) that lands in the source volume's zone with the source's encryption posture; both expose the same downstream surface
- **Volume Attachments** — one per `attachments[]` entry, keyed by device name and instance, with the declared detach behavior
- **AWS Tags** — resource metadata tags (organization, environment, resource kind, resource ID) applied automatically for tracking and governance

## Before You Deploy

### Planton Setup

- **AWS Provider Connection** — an active connection in the Connect module with credentials for the target AWS account, including EC2 permissions. Map it as the default for your environment, or specify it explicitly when creating the Infra Component.
- **Planton Runner** — required when using Runner-based credential delivery. Not needed for inline credentials or cross-account trust authentication modes.

### AWS Account

- **An instance in the SAME availability zone** (only for attachments) — EBS never attaches across zones; the instance's placement decides the volume's `availabilityZone`.
- **The source snapshot** (only for restores) — in the same region; reference an AwsEbsSnapshot Infra Component or pass a literal `snap-...` id.
- **A KMS key** (only for `aws:kms`-style customer-key encryption) — unset with `encrypted: true` uses the AWS-managed `aws/ebs` key.

## Deploy

### Console

Open the deployment store, find **AWS EBS Volume**, and click **Deploy**. The creation wizard walks you through preset selection, environment and connection configuration, and spec fields. Start from the **Database Data Volume** preset in the [Presets](#presets) tab for the self-managed-database baseline: an encrypted gp3 with tuned IOPS and throughput, attached at `/dev/sdf`.

### CLI

Create a manifest and apply it:

```yaml
apiVersion: aws.planton.dev/v1alpha1
kind: AwsEbsVolume
metadata:
  name: db-data
  org: acme-corp
  env: prod
spec:
  region: us-east-1
  availabilityZone: us-east-1a
  type: gp3
  sizeGb: 200
  iops: 6000
  throughputMibps: 250
  encrypted: true
  attachments:
    - deviceName: /dev/sdf
      instanceId:
        value: i-0123456789abcdef0
```

```shell
planton apply -f ebs-volume.yaml
```

This creates an encrypted 200 GiB gp3 volume with provisioned IOPS and throughput, attached to the referenced instance at `/dev/sdf` — the attachment presents the disk; the filesystem is yours. An Infra Job tracks the provisioning in real time.

### InfraChart

When the volume deploys alongside its instance in one chart, wire the instance reference via ValueFromRef:

```yaml
spec:
  region: us-east-1
  availabilityZone: us-east-1a
  type: gp3
  sizeGb: 200
  encrypted: true
  attachments:
    - deviceName: /dev/sdf
      instanceId:
        valueFrom:
          kind: AwsEc2Instance
          name: db-host
          fieldPath: status.outputs.instance_id
```

The InfraPipeline resolves the dependency graph, provisions the instance first, then creates and attaches the volume.

## Key Configuration

These are the most important decisions when configuring an EBS volume. Explore the full field reference in the [API Explorer](#api-explorer) tab.

**The zone is the contract** — a volume lives and dies in one availability zone, and EBS refuses cross-AZ attachments. Decide `availabilityZone` from the compute's placement FIRST; moving later means snapshot, restore in the target zone, re-attach — not an edit.

**Type decides which dials exist** — `gp3` is the current-generation default and the only type with a throughput dial (`throughputMibps`); `io1`/`io2` are provisioned-IOPS types that REQUIRE an explicit `iops` value and are the only multi-attach-capable types; `sc1`/`st1` are throughput HDDs. The spec enforces these pairings, so a wrong combination fails at validation, not at apply.

**In-place elasticity has a cooldown** — size, type, IOPS, and throughput all modify in place, but AWS allows roughly one modification per volume per six hours. Batch your dials into one change instead of trickling them, and grow the filesystem after growing the volume.

**Copies inherit what you cannot override** — `copyFrom` lands the copy in the SOURCE volume's zone with the source's encryption posture and snapshot lineage; the provider offers no override, so the spec forbids the create-arm fields on a copy. Need a different zone or key? Snapshot, then restore.

**Multi-attach is a filesystem decision, not a flag** — `multiAttachEnabled` (io1/io2 only) lets several instances see the same block device with NO write coordination: a regular ext4/xfs mount on two instances corrupts data silently. Enable it only for cluster-aware filesystems or applications doing their own fencing. It is fixed for life, and more than one attachment entry requires it.

**finalSnapshot is the safety net nobody sees** — it snapshots the volume on destroy, but the snapshot is untracked (config-only at AWS, invisible to imports). Budget for it in cleanup sweeps, and find the snapshots by the volume's tags when reclaiming space.

**Detach behavior is per attachment** — `skipDestroy` leaves the volume attached when the entry leaves management; `stopInstanceBeforeDetaching` is the clean option for data volumes that cannot unmount live; `forceDetach` skips the filesystem flush and is a last resort for hung instances — data loss is possible.

## Outputs and Dependencies

### What This Kind Consumes

| Dependency | Field | ValueFromRef Path |
|------------|-------|-------------------|
| **AwsEbsSnapshot** (optional, create arm) | `snapshotId` | `status.outputs.snapshot_id` |
| **AwsKmsKey** (optional, create arm) | `kmsKeyId` | `status.outputs.key_arn` |
| **AwsEbsVolume** (copy arm) | `copyFrom.sourceVolumeId` | `status.outputs.volume_id` |
| **AwsEc2Instance** (per attachment) | `attachments[].instanceId` | `status.outputs.instance_id` |

### What This Kind Provides

After provisioning, `status.outputs` contains values that downstream Infra Components can consume via ValueFromRef:

| Output | Description | Common Downstream Use |
|--------|-------------|----------------------|
| `volume_id` | The volume's id (`vol-...`) | AwsEbsSnapshot's volume arm (`volumeId`) and other volumes' copy arms (`copyFrom.sourceVolumeId`) |
| `volume_arn` | The volume's ARN | IAM policy statements scoping volume actions |
| `availability_zone` | The zone the volume actually lives in — notably useful for copies, which inherit the source's zone | Placing the compute that must sit beside the disk |

`size_gb` and `create_time` are also exported — the actual size (the snapshot's size when `sizeGb` was left unset) and the creation timestamp, observability values rather than composition inputs.

## Common Patterns

Browse the [Presets](#presets) tab for ready-to-deploy configurations.

**The database data volume** — an encrypted gp3 with tuned IOPS and throughput, attached at `/dev/sdf`, holding the data of a self-managed database. The volume outlives the instance: replace the EC2 host and re-attach the same disk, which is the whole point of declaring storage separately from compute. Start from the **Database Data Volume** preset.

**The zone move** — volumes never move zones, so "moving" one is rebuilding it from a snapshot in the target zone. Set `volumeInitializationRate` to hydrate blocks eagerly so first reads hit full performance instead of lazy-loading from S3 — worth it when the volume goes straight into service. Start from the **Snapshot Restore** preset.

## Works With

- [**AWS EC2 Instance**](/infra-catalog/aws-ec2-instance) — the compute the volume attaches to, wired via `attachments[].instanceId`
- [**AWS EBS Snapshot**](/infra-catalog/aws-ebs-snapshot) — the restore source (`snapshotId`) and the backup consumer of `volume_id`
- [**AWS KMS Key**](/infra-catalog/aws-kms-key) — the customer-managed key for at-rest encryption, wired via `kmsKeyId`
- [**AWS Data Lifecycle Manager Policy**](/infra-catalog/aws-dlm-lifecycle-policy) — tag-driven recurring backups of this volume, no per-volume wiring
