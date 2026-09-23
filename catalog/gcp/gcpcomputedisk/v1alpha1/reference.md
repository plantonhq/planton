# GcpComputeDisk

> Generated from the protobuf schema by `make generate-reference` -- do not
> edit by hand. To change a fact on this page, change the proto field comment
> or validation rule it is derived from, then regenerate.

**apiVersion**: `gcp.planton.dev/v1alpha1`

**Guide**: [GUIDE.md](../GUIDE.md) -- authored operational judgment for this component: conventions, trade-offs, and what pairs well with it.

GcpComputeDiskSpec defines the configuration for a zonal Google Compute
Engine persistent disk — the durable block device behind stateful VMs:
database volumes, shared read-only datasets, bootable golden images,
and any data that must outlive the instance it is attached to.

A disk is a first-class node with its own lifecycle: create it once,
attach it to a GcpComputeInstance by reference, and the data survives
instance replacement, resizing, and rescheduling.

Important behavioral notes:

  - disk_name, zone, type, source (image / snapshot / instant
    snapshot / storage object / disk), encryption, and architecture
    are create-time decisions — changing them replaces the disk (and
    its data). size_gb grows in place but never shrinks.
  - At most one source may be set. With no source the disk is created
    empty (the common case for data volumes).
  - Deleting a disk that is still attached to a running instance fails;
    detach first (or delete the instance).
  - Regional (dual-zone replicated) disks are a separate GCP resource
    with a materially different surface and are not modeled here.

## Example

```yaml
apiVersion: gcp.planton.dev/v1alpha1
kind: GcpComputeDisk
metadata:
  name: test-disk
spec:
  projectId:
    value: planton-testing
  zone: us-central1-a
  type: pd-balanced
  sizeGb: 50
  labels:
    env: testing
```

## Spec Fields

| Path | Type | Required | Default | References |
|---|---|---|---|---|
| `spec.projectId` | `string \| valueFrom` |  |  | GcpProject (`status.outputs.project_id`) |
| `spec.diskName` | `string` |  |  |  |
| `spec.zone` | `string` | yes |  |  |
| `spec.description` | `string` |  |  |  |
| `spec.type` | `string` |  |  |  |
| `spec.sizeGb` | `int32` |  |  |  |
| `spec.image` | `string` |  |  |  |
| `spec.sourceSnapshot` | `string` |  |  |  |
| `spec.sourceDisk` | `string \| valueFrom` |  |  | GcpComputeDisk (`status.outputs.self_link`) |
| `spec.kmsKey` | `string \| valueFrom` |  |  | GcpKmsKey (`status.outputs.key_id`) |
| `spec.provisionedIops` | `int64` |  |  |  |
| `spec.provisionedThroughput` | `int64` |  |  |  |
| `spec.accessMode` | `string` |  |  |  |
| `spec.architecture` | `string` |  |  |  |
| `spec.enableConfidentialCompute` | `bool` |  |  |  |
| `spec.physicalBlockSizeBytes` | `int32` |  |  |  |
| `spec.createSnapshotBeforeDestroy` | `bool` |  |  |  |
| `spec.snapshotBeforeDestroyPrefix` | `string` |  |  |  |
| `spec.storagePool` | `string` |  |  |  |
| `spec.labels` | `map<string, string>` |  |  |  |
| `spec.resourceManagerTags` | `map<string, string>` |  |  |  |
| `spec.guestOsFeatures` | `[]string` |  |  |  |
| `spec.licenses` | `[]string` |  |  |  |
| `spec.sourceInstantSnapshot` | `string` |  |  |  |
| `spec.sourceStorageObject` | `string` |  |  |  |
| `spec.deletionPolicy` | `string` |  |  |  |
| `spec.kmsKeyServiceAccount` | `string` |  |  |  |
| `spec.sourceImageEncryption` | `GcpComputeDiskSourceEncryption` |  |  |  |
| `spec.sourceImageEncryption.kmsKey` | `string \| valueFrom` | yes |  | GcpKmsKey (`status.outputs.key_id`) |
| `spec.sourceImageEncryption.kmsKeyServiceAccount` | `string` |  |  |  |
| `spec.sourceSnapshotEncryption` | `GcpComputeDiskSourceEncryption` |  |  |  |
| `spec.sourceSnapshotEncryption.kmsKey` | `string \| valueFrom` | yes |  | GcpKmsKey (`status.outputs.key_id`) |
| `spec.sourceSnapshotEncryption.kmsKeyServiceAccount` | `string` |  |  |  |
| `spec.asyncPrimaryDisk` | `string \| valueFrom` |  |  | GcpComputeDisk (`status.outputs.self_link`) |

## Field Details

### spec.projectId

`string | valueFrom`

The GCP project that owns the disk.
Can be a literal project ID or a reference to a GcpProject resource.
If omitted, the provider's default project is used.
Immutable after creation.

- references: GcpProject (`status.outputs.project_id`)
- rule: write as {value: <literal>} or {valueFrom: {kind: GcpProject, name: <that resource's name>, fieldPath: status.outputs.project_id}} -- a bare string does not parse

### spec.diskName

`string`

Name of the disk. 1-63 characters, lowercase letters, numbers, and
hyphens; must start with a letter and cannot end with a hyphen. When
omitted, metadata.name is used. Immutable after creation.

- rule: disk_name must be 1-63 characters of lowercase letters, numbers, and hyphens, starting with a letter and not ending with a hyphen

### spec.zone

`string` · required

Zone the disk lives in, e.g. "us-central1-a". A disk attaches only to
instances in the same zone. Immutable after creation.

- rule: {"required":true,"string":{"pattern":"^[a-z]+-[a-z]+[0-9]+-[a-z]$"}}

### spec.description

`string`

Human-readable description of the disk.

### spec.type

`string`

Disk type: "pd-standard" (HDD), "pd-balanced" (GCP's default and the
sensible general choice), "pd-ssd" (high IOPS), "pd-extreme"
(provisioned-IOPS pd), or a hyperdisk type ("hyperdisk-balanced",
"hyperdisk-extreme", "hyperdisk-throughput", "hyperdisk-ml") on
supported machine families. Immutable after creation.

### spec.sizeGb

`int32`

Size in GB. Required for empty disks; optional with a source (the
source's size is used, and a larger value grows the disk). Grows in
place; shrinking is impossible.

- rule: {"ignore":"IGNORE_IF_ZERO_VALUE","int32":{"lte":65536,"gte":1}}

### spec.image

`string`

Source image to initialize the disk from — makes the disk bootable.
Accepts an image family ("debian-cloud/debian-12") or a specific
image self link. Create-time only.

### spec.sourceSnapshot

`string`

Source snapshot to restore the disk from (name or self link).
Create-time only.

### spec.sourceDisk

`string | valueFrom`

Existing disk to clone, referenced as another GcpComputeDisk (or a
literal self link). Create-time only.

- references: GcpComputeDisk (`status.outputs.self_link`)
- rule: write as {value: <literal>} or {valueFrom: {kind: GcpComputeDisk, name: <that resource's name>, fieldPath: status.outputs.self_link}} -- a bare string does not parse

### spec.kmsKey

`string | valueFrom`

Customer-managed encryption key (CMEK), referenced as a GcpKmsKey.
The Compute Engine service agent
(service-<project-number>@compute-system.iam.gserviceaccount.com)
must hold roles/cloudkms.cryptoKeyEncrypterDecrypter on the key.
When omitted, Google-managed encryption is used. Immutable after
creation.

- references: GcpKmsKey (`status.outputs.key_id`)
- rule: write as {value: <literal>} or {valueFrom: {kind: GcpKmsKey, name: <that resource's name>, fieldPath: status.outputs.key_id}} -- a bare string does not parse

### spec.provisionedIops

`int64` · optional (explicit presence)

Provisioned IOPS. Required for pd-extreme and hyperdisk-extreme;
tunable on hyperdisk-balanced. Hyperdisk types update in place
(at most every 4 hours); other types cannot set this.

- rule: {"int64":{"gt":"0"}}

### spec.provisionedThroughput

`int64` · optional (explicit presence)

Provisioned throughput in MB/s. Tunable on hyperdisk-throughput and
hyperdisk-balanced; updates in place (at most every 4 hours).

- rule: {"int64":{"gt":"0"}}

### spec.accessMode

`string`

Access mode of the disk:
  ""                     -- same as "READ_WRITE_SINGLE" (GCP default)
  "READ_WRITE_SINGLE"    -- one instance read-write
  "READ_WRITE_MANY"      -- many instances read-write (hyperdisk-ml
                            and supported multi-writer types)
  "READ_ONLY_MANY"       -- many instances read-only

- rule: access_mode must be READ_WRITE_SINGLE, READ_WRITE_MANY, or READ_ONLY_MANY

### spec.architecture

`string`

CPU architecture the disk's contents target: "X86_64" or "ARM64".
Relevant for bootable disks; normally inferred from the image.
Immutable after creation.

- rule: architecture must be X86_64 or ARM64

### spec.enableConfidentialCompute

`bool`

Create the disk in confidential-compute mode (hyperdisk SKUs only;
requires kms_key).

### spec.physicalBlockSizeBytes

`int32` · optional (explicit presence)

Physical block size in bytes: 4096 (default) or 16384.

- rule: {"int32":{"in":[4096,16384]}}

### spec.createSnapshotBeforeDestroy

`bool`

Take a snapshot of the disk immediately before it is destroyed — a
last-resort recovery net for precious data volumes. The snapshot is
named "<disk-name>-YYYYMMDD-HHmmss" unless
snapshot_before_destroy_prefix overrides the prefix.

### spec.snapshotBeforeDestroyPrefix

`string`

Custom name prefix for the snapshot taken by
create_snapshot_before_destroy.

### spec.storagePool

`string`

URL or name of the storage pool to create the disk in (hyperdisk
storage pools).

### spec.labels

`map<string, string>`

User labels merged with Planton attribution labels (which win on key
conflicts).

### spec.resourceManagerTags

`map<string, string>`

Resource Manager tags bound to the disk for org-policy and IAM
conditions. Keys in the form "tagKeys/{id}", values "tagValues/{id}".
Create-time only.

### spec.guestOsFeatures

`[]string`

Guest OS features to enable when this disk is used as a boot disk,
e.g. ["UEFI_COMPATIBLE", "SECURE_BOOT", "GVNIC", "MULTI_IP_SUBNET",
"WINDOWS"]. The accepted set evolves with GCP — see "Enabling guest
operating system features" in the Compute Engine docs. Create-time
only.

### spec.licenses

`[]string`

License URIs applicable to this disk, e.g. a Windows Server or
SQL Server license
("https://www.googleapis.com/compute/v1/projects/windows-cloud/global/licenses/windows-server-core").
Normally inherited from the source image; set explicitly when
importing raw disks that need bring-your-own-license attribution.
Create-time only.

### spec.sourceInstantSnapshot

`string`

Source INSTANT snapshot to create the disk from (name, partial or
full URL). Instant snapshots are near-instant, same-region-only
point-in-time copies — the fast restore path, vs. standard
snapshots' cross-region durability. Create-time only.

### spec.sourceStorageObject

`string`

Full Google Cloud Storage URI of a raw disk image to create the
disk from (e.g. "https://storage.googleapis.com/bucket/image.vmdk"
or a gs:// URI) — the import-a-disk-file path, skipping the
intermediate compute image. Create-time only.

### spec.deletionPolicy

`string`

Deletion policy — what happens when this resource is destroyed:
  ""        -- same as "DELETE" (provider default)
  "DELETE"  -- the disk (and its data) is deleted
  "PREVENT" -- destroy FAILS; a guard rail for precious data volumes
               (create_snapshot_before_destroy is the softer net)
  "ABANDON" -- the disk is removed from management but left running
               in GCP

- rule: deletion_policy must be one of: DELETE, PREVENT, ABANDON

### spec.kmsKeyServiceAccount

`string`

Service account used for the encryption request of kms_key (CMEK).
When omitted, the Compute Engine default service agent is used.
Only meaningful together with kms_key. Immutable after creation.

### spec.sourceImageEncryption

`GcpComputeDiskSourceEncryption`

Decrypts the source image when it is itself CMEK-encrypted. Only
valid together with image.

### spec.sourceImageEncryption.kmsKey

`string | valueFrom` · required

The KMS key the source was encrypted with, referenced as a GcpKmsKey
or a literal self link. The service agent performing the read needs
roles/cloudkms.cryptoKeyEncrypterDecrypter on it.

- references: GcpKmsKey (`status.outputs.key_id`)
- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: GcpKmsKey, name: <that resource's name>, fieldPath: status.outputs.key_id}} -- a bare string does not parse

### spec.sourceImageEncryption.kmsKeyServiceAccount

`string`

Service account used for the decryption request. When omitted, the
Compute Engine default service agent is used.

### spec.sourceSnapshotEncryption

`GcpComputeDiskSourceEncryption`

Decrypts the source snapshot when it is itself CMEK-encrypted. Only
valid together with source_snapshot.

### spec.sourceSnapshotEncryption.kmsKey

`string | valueFrom` · required

The KMS key the source was encrypted with, referenced as a GcpKmsKey
or a literal self link. The service agent performing the read needs
roles/cloudkms.cryptoKeyEncrypterDecrypter on it.

- references: GcpKmsKey (`status.outputs.key_id`)
- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: GcpKmsKey, name: <that resource's name>, fieldPath: status.outputs.key_id}} -- a bare string does not parse

### spec.sourceSnapshotEncryption.kmsKeyServiceAccount

`string`

Service account used for the decryption request. When omitted, the
Compute Engine default service agent is used.

### spec.asyncPrimaryDisk

`string | valueFrom`

Makes this disk an ASYNC REPLICATION SECONDARY of the referenced
primary disk (another GcpComputeDisk, or a literal disk self link in
another region). Replication starts when the pair is activated on
the primary; the secondary must match the primary's size and type.
Create-time only.

- references: GcpComputeDisk (`status.outputs.self_link`)
- rule: write as {value: <literal>} or {valueFrom: {kind: GcpComputeDisk, name: <that resource's name>, fieldPath: status.outputs.self_link}} -- a bare string does not parse

## Validation Rules

- `disk_at_most_one_source`: at most one disk source may be set: image, source_snapshot, source_instant_snapshot, source_storage_object, or source_disk — omit all for an empty disk
- `empty_disk_requires_size`: an empty disk (no image, snapshot, or source disk) must declare size_gb
- `image_encryption_requires_image`: source_image_encryption is only valid together with image
- `snapshot_encryption_requires_snapshot`: source_snapshot_encryption is only valid together with source_snapshot
- `confidential_compute_requires_kms_key`: confidential-compute disks require customer-managed encryption — set kms_key when enable_confidential_compute is true

## Outputs

Reference an output from another manifest as `valueFrom: {kind: GcpComputeDisk, name: <resource-name>, fieldPath: status.outputs.<output>}`.

| Output | Type | Description |
|---|---|---|
| `status.outputs.name` | `string` | Name of the disk in GCP. |
| `status.outputs.disk_id` | `string` | Server-assigned unique numeric identifier of the disk. |
| `status.outputs.self_link` | `string` | Self-link URL of the disk — the composition key a GcpComputeInstance's boot_disk.source_disk or attached_disks[].source consumes. |
| `status.outputs.zone` | `string` | Zone the disk lives in (plain zone name, e.g. "us-central1-a"). |
| `status.outputs.size_gb` | `int32` | Provisioned size of the disk in GB. |
| `status.outputs.type` | `string` | Disk type (plain type name, e.g. "pd-balanced"). |

## References

Fields that can point at another resource's outputs:

| Field | Kind | Output |
|---|---|---|
| `spec.projectId` | GcpProject | `status.outputs.project_id` |
| `spec.sourceDisk` | GcpComputeDisk | `status.outputs.self_link` |
| `spec.kmsKey` | GcpKmsKey | `status.outputs.key_id` |
| `spec.sourceImageEncryption.kmsKey` | GcpKmsKey | `status.outputs.key_id` |
| `spec.sourceSnapshotEncryption.kmsKey` | GcpKmsKey | `status.outputs.key_id` |
| `spec.asyncPrimaryDisk` | GcpComputeDisk | `status.outputs.self_link` |

## Referenced By

Fields on other kinds that can point at this resource:

| Kind | Field | Reads |
|---|---|---|
| GcpComputeDisk | `spec.sourceDisk` | `status.outputs.self_link` |
| GcpComputeDisk | `spec.asyncPrimaryDisk` | `status.outputs.self_link` |
| GcpComputeInstance | `spec.bootDisk.sourceDisk` | `status.outputs.self_link` |
| GcpComputeInstance | `spec.attachedDisks[].source` | `status.outputs.self_link` |
| GcpComputeMig | `spec.template.disks[].source` | `status.outputs.self_link` |
| GcpComputeMig | `spec.perInstanceConfigs[].preservedState.disks[].source` | `status.outputs.self_link` |
| GcpTpuVm | `spec.dataDisks[].sourceDisk` | `status.outputs.self_link` |

## See Also

- [Overview](../README.md)
