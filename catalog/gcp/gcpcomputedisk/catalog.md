# GCP Compute Disk

Deploys a zonal Compute Engine persistent disk — the durable block device behind stateful VMs: database volumes, shared read-only datasets, bootable golden images, and any data that must outlive the instance it is attached to. Supports the full pd and hyperdisk type families with in-place performance tuning, customer-managed encryption, a snapshot-before-destroy recovery net, and ValueFromRef wiring to GCP projects, KMS keys, and other disks.

## What Gets Created

When you deploy this Cloud Resource, the IaC module provisions:

- **Compute Engine Persistent Disk** -- a zonal block device in the specified project and zone, created empty (the common data-volume case) or initialized from exactly one source: an image (bootable), a snapshot (restore), an instant snapshot (fast same-region restore), a Cloud Storage disk-image file (import), or an existing GcpComputeDisk (clone)
- **Performance Configuration** -- the disk type (pd-standard, pd-balanced, pd-ssd, pd-extreme, or hyperdisk-*), with provisioned IOPS/throughput dials on the types that support tuning
- **Encryption** -- Google-managed by default; customer-managed (CMEK) when a GcpKmsKey is referenced, with optional confidential-compute mode on hyperdisk SKUs
- **Destroy-Snapshot Net** -- created only when `createSnapshotBeforeDestroy` is enabled; a timestamped final snapshot taken immediately before the disk is destroyed
- **GCP Labels** -- resource metadata labels (resource name, kind, organization, environment) applied automatically for tracking and governance

## Before You Deploy

### Planton Setup

- **GCP Provider Connection** -- an active connection in the Connect module with credentials for the target GCP project. Map it as the default for your environment, or specify it explicitly when creating the Cloud Resource.
- **Planton Runner** -- required when using Runner-based credential delivery. Not needed for inline credentials or browser OAuth authentication modes.

### GCP Project

- **A GCP project** where the disk will be created. Provide the project ID directly or reference a GcpProject Cloud Resource via ValueFromRef.
- **IAM permissions** -- `roles/compute.storageAdmin` on the target project.
- **For CMEK** -- the Compute Engine service agent (`service-<project-number>@compute-system.iam.gserviceaccount.com`) must hold `roles/cloudkms.cryptoKeyEncrypterDecrypter` on the referenced GcpKmsKey.

## Deploy

### Console

Open the deployment store, find **GCP Compute Disk**, and click **Deploy**. The creation wizard walks you through preset selection, environment and connection configuration, and spec fields. Start from the **Data Volume** preset in the [Presets](#presets) tab for the bread-and-butter empty data disk.

### CLI

Create a manifest and apply it:

```yaml
apiVersion: gcp.planton.dev/v1alpha1
kind: GcpComputeDisk
metadata:
  name: postgres-data
  org: acme-corp
  env: prod
spec:
  projectId:
    value: "acme-prod-12345"
  zone: us-central1-a
  type: pd-balanced
  sizeGb: 100
  labels:
    team: platform
```

```shell
planton apply -f disk.yaml
```

This creates an empty 100 GB pd-balanced disk ready to attach to a VM in `us-central1-a`. A Stack Job tracks the provisioning in real time.

### InfraChart

When deploying as part of a multi-resource environment, use ValueFromRef to wire the disk to a project and CMEK key deployed in the same InfraPipeline — and wire the VM to the disk:

```yaml
spec:
  projectId:
    valueFrom:
      kind: GcpProject
      name: production-project
      fieldPath: status.outputs.project_id
  kmsKey:
    valueFrom:
      kind: GcpKmsKey
      name: disk-encryption-key
      fieldPath: status.outputs.key_id
```

The consuming GcpComputeInstance then references this disk from `attachedDisks[].source` (or `bootDisk.sourceDisk`) via `status.outputs.self_link`.

## Key Configuration

These are the most important decisions when configuring a persistent disk. Explore the full field reference in the [API Explorer](#api-explorer) tab.

**Zone** -- A disk attaches only to instances in the SAME zone; put it where its VM lives. Immutable — cross-zone moves go through a snapshot round-trip.

**Source and size** -- At most one source: empty (declare `sizeGb`), `image` (makes the disk bootable), `sourceSnapshot` (restore), `sourceInstantSnapshot` (fast same-region restore), `sourceStorageObject` (import a raw disk image straight from Cloud Storage), or `sourceDisk` (clone another GcpComputeDisk). Size only GROWS — in place, with a filesystem grow from the guest; shrinking is impossible. Encrypted image/snapshot sources decrypt via `sourceImageEncryption` / `sourceSnapshotEncryption` (CMEK only).

**Type and performance** -- The pd-* family scales performance with size; the hyperdisk-* family decouples them with `provisionedIops`/`provisionedThroughput` dials that update IN PLACE (at most every 4 hours). pd-extreme and hyperdisk-extreme REQUIRE provisioned IOPS.

**Access mode** -- `READ_ONLY_MANY` shares one dataset across many VMs; `READ_WRITE_MANY` needs a multi-writer-capable type (hyperdisk-ml).

**Encryption and recovery** -- Reference a GcpKmsKey for CMEK (your revocation lever; immutable; `kmsKeyServiceAccount` overrides the requesting identity). Raw CSEK keys are deliberately not modeled — key material never flows through manifests or state. Arm `createSnapshotBeforeDestroy` on precious volumes so even a mistaken destroy leaves a timestamped snapshot, or set `deletionPolicy: PREVENT` to fail destroys outright. For cross-region disaster recovery, `asyncPrimaryDisk` pairs this disk as an async-replication secondary of a primary in another region.

## Outputs and Dependencies

### What This Component Consumes

| Dependency | Field | ValueFromRef Path |
|------------|-------|-------------------|
| **GcpProject** | `projectId` | `status.outputs.project_id` |
| **GcpComputeImage** (optional) | `image` (bootable) | `status.outputs.self_link` |
| **GcpComputeDisk** (optional) | `sourceDisk` (clone) | `status.outputs.self_link` |
| **GcpComputeDisk** (optional) | `asyncPrimaryDisk` (async replication) | `status.outputs.self_link` |
| **GcpKmsKey** (optional) | `kmsKey` | `status.outputs.key_id` |
| **GcpKmsKey** (optional) | `sourceImageEncryption.kmsKey` / `sourceSnapshotEncryption.kmsKey` | `status.outputs.key_id` |

### What This Component Provides

After provisioning, `status.outputs` contains values that downstream Cloud Resources can consume via ValueFromRef:

| Output | Description | Common Downstream Use |
|--------|-------------|----------------------|
| `self_link` | Self-link URL of the disk | GcpComputeInstance `bootDisk.sourceDisk` / `attachedDisks[].source` |
| `name` | Name of the disk in GCP | Monitoring, snapshot schedules |
| `disk_id` | Server-assigned unique numeric identifier | API references, audit logs |
| `zone` | Zone the disk lives in | Placement checks for consuming VMs |

## Common Patterns

Browse the [Presets](#presets) tab for ready-to-deploy configurations.

**Data Volume** -- An empty 100 GB pd-balanced disk — the bread-and-butter data volume for any stateful VM. Start from the **Data Volume** preset.

**Encrypted Database Volume** -- A 500 GB pd-ssd volume with CMEK encryption and the snapshot-before-destroy net armed — the posture for volumes that hold real customer data. Start from the **Encrypted Database Volume** preset.

**Hyperdisk High IOPS** -- A 1 TB hyperdisk-balanced with 6,000 provisioned IOPS and 290 MB/s throughput, tunable in place as the workload grows. Start from the **Hyperdisk High IOPS** preset.

## Works With

- [**GCP Compute Instance**](/cloud-catalog/gcp-compute-instance) -- attaches this disk as a boot or data volume by self link
- [**GCP Project**](/cloud-catalog/gcp-project) -- provides the GCP project where the disk is created
- [**GCP KMS Key**](/cloud-catalog/gcp-kms-key) -- provides the customer-managed encryption key
