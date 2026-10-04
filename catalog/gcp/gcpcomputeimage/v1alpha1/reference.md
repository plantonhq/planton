# GcpComputeImage

> Generated from the protobuf schema by `make generate-reference` -- do not
> edit by hand. To change a fact on this page, change the proto field comment
> or validation rule it is derived from, then regenerate.

**apiVersion**: `gcp.planton.dev/v1alpha1`

**Guide**: [GUIDE.md](../GUIDE.md) -- authored operational judgment for this kind: conventions, trade-offs, and what pairs well with it.

GcpComputeImageSpec defines a Compute Engine custom image
(`google_compute_image`): the golden boot image VMs, instance templates,
managed instance groups, and disks start from -- an OS hardened and
pre-loaded with an organization's agents, built once and reused.

An image is built from exactly one source: a GcpComputeDisk (the
classic golden-image path: configure a VM's boot disk, stop the VM,
image the disk), another image (copying or re-keying one), a snapshot,
or a raw disk tarball in Cloud Storage (an imported or Packer-built
image).

Important behavioral notes:

  - Everything except labels is a create-time decision: changing a
    source, the name, the family, a key, or a guest OS feature replaces
    the image. Images are therefore versioned, not edited: give each
    build its own image_name (e.g. "web-base-20261001") and the same
    family ("web-base"), and consumers that boot from
    "projects/{project}/global/images/family/web-base" always get the
    newest image of the family without changing.
  - Destroy deletes the image; VMs and disks already created from it are
    unaffected, but new ones can no longer use it.
  - Customer-supplied raw keys (CSEK) are deliberately not modeled: raw
    key material does not belong in declarative manifests or state --
    use Cloud KMS keys, where the key stays in Cloud KMS.

## Example

```yaml
apiVersion: gcp.planton.dev/v1alpha1
kind: GcpComputeImage
metadata:
  name: web-base-20261001
spec:
  projectId:
    value: my-gcp-project
  family: web-base
  description: Hardened web base image
  sourceDisk:
    value: projects/my-gcp-project/zones/us-central1-a/disks/web-build
  kmsKey:
    value: projects/my-gcp-project/locations/us/keyRings/images/cryptoKeys/images
  guestOsFeatures:
    - UEFI_COMPATIBLE
    - GVNIC
  storageLocations:
    - us
  labels:
    os: debian
```

## Spec Fields

| Path | Type | Required | Default | References |
|---|---|---|---|---|
| `spec.projectId` | `string \| valueFrom` |  |  | GcpProject (`status.outputs.project_id`) |
| `spec.imageName` | `string` |  |  |  |
| `spec.description` | `string` |  |  |  |
| `spec.family` | `string` |  |  |  |
| `spec.sourceDisk` | `string \| valueFrom` |  |  | GcpComputeDisk (`status.outputs.self_link`) |
| `spec.sourceImage` | `string \| valueFrom` |  |  | GcpComputeImage (`status.outputs.self_link`) |
| `spec.sourceSnapshot` | `string` |  |  |  |
| `spec.rawDisk` | `GcpComputeImageRawDisk` |  |  |  |
| `spec.rawDisk.source` | `string` | yes |  |  |
| `spec.rawDisk.sha1` | `string` |  |  |  |
| `spec.rawDisk.containerType` | `string` |  |  |  |
| `spec.kmsKey` | `string \| valueFrom` |  |  | GcpKmsKey (`status.outputs.key_id`), GcpKmsKeyHandle (`status.outputs.kms_key`) |
| `spec.kmsKeyServiceAccount` | `string` |  |  |  |
| `spec.sourceDiskEncryption` | `GcpComputeImageSourceEncryption` |  |  |  |
| `spec.sourceDiskEncryption.kmsKey` | `string \| valueFrom` | yes |  | GcpKmsKey (`status.outputs.key_id`) |
| `spec.sourceDiskEncryption.kmsKeyServiceAccount` | `string` |  |  |  |
| `spec.sourceImageEncryption` | `GcpComputeImageSourceEncryption` |  |  |  |
| `spec.sourceImageEncryption.kmsKey` | `string \| valueFrom` | yes |  | GcpKmsKey (`status.outputs.key_id`) |
| `spec.sourceImageEncryption.kmsKeyServiceAccount` | `string` |  |  |  |
| `spec.sourceSnapshotEncryption` | `GcpComputeImageSourceEncryption` |  |  |  |
| `spec.sourceSnapshotEncryption.kmsKey` | `string \| valueFrom` | yes |  | GcpKmsKey (`status.outputs.key_id`) |
| `spec.sourceSnapshotEncryption.kmsKeyServiceAccount` | `string` |  |  |  |
| `spec.diskSizeGb` | `int64` |  |  |  |
| `spec.guestOsFeatures` | `[]string` |  |  |  |
| `spec.licenses` | `[]string` |  |  |  |
| `spec.storageLocations` | `[]string` |  |  |  |
| `spec.shieldedInstanceInitialState` | `GcpComputeImageShieldedInstanceInitialState` |  |  |  |
| `spec.shieldedInstanceInitialState.pk` | `GcpComputeImageFileContentBuffer` |  |  |  |
| `spec.shieldedInstanceInitialState.pk.content` | `string` | yes |  |  |
| `spec.shieldedInstanceInitialState.pk.fileType` | `string` |  |  |  |
| `spec.shieldedInstanceInitialState.keks` | `[]GcpComputeImageFileContentBuffer` |  |  |  |
| `spec.shieldedInstanceInitialState.keks[].content` | `string` | yes |  |  |
| `spec.shieldedInstanceInitialState.keks[].fileType` | `string` |  |  |  |
| `spec.shieldedInstanceInitialState.dbs` | `[]GcpComputeImageFileContentBuffer` |  |  |  |
| `spec.shieldedInstanceInitialState.dbs[].content` | `string` | yes |  |  |
| `spec.shieldedInstanceInitialState.dbs[].fileType` | `string` |  |  |  |
| `spec.shieldedInstanceInitialState.dbxs` | `[]GcpComputeImageFileContentBuffer` |  |  |  |
| `spec.shieldedInstanceInitialState.dbxs[].content` | `string` | yes |  |  |
| `spec.shieldedInstanceInitialState.dbxs[].fileType` | `string` |  |  |  |
| `spec.resourceManagerTags` | `map<string, string>` |  |  |  |
| `spec.labels` | `map<string, string>` |  |  |  |
| `spec.deletionPolicy` | `string` |  |  |  |

## Field Details

### spec.projectId

`string | valueFrom`

The GCP project that owns the image: a literal project ID or a
GcpProject reference. If omitted, the provider's default project is
used. Immutable.

- references: GcpProject (`status.outputs.project_id`)
- rule: write as {value: <literal>} or {valueFrom: {kind: GcpProject, name: <that resource's name>, fieldPath: status.outputs.project_id}} -- a bare string does not parse

### spec.imageName

`string`

Name of the image, unique in the project: 1-63 lowercase letters,
digits, and hyphens, starting with a letter and not ending with a
hyphen. Version it ("web-base-20261001") and let family carry the
stable name. Defaults to metadata.name. Immutable.

- rule: image_name must be 1-63 characters of lowercase letters, numbers, and hyphens, starting with a letter and not ending with a hyphen

### spec.description

`string`

A human-readable description of the image. Immutable.

### spec.family

`string`

The image family: the stable name consumers boot from
("projects/{project}/global/images/family/{family}" resolves to the
newest non-deprecated image in it). Same naming rules as image_name.
Immutable.

- rule: family must be 1-63 characters of lowercase letters, numbers, and hyphens, starting with a letter and not ending with a hyphen

### spec.sourceDisk

`string | valueFrom`

Source: a disk to image, referenced as a GcpComputeDisk (its
self_link) or a literal disk self link. Stop or detach the VM using
the disk first so the image is consistent. Exactly one source.
Immutable.

- references: GcpComputeDisk (`status.outputs.self_link`)
- rule: write as {value: <literal>} or {valueFrom: {kind: GcpComputeDisk, name: <that resource's name>, fieldPath: status.outputs.self_link}} -- a bare string does not parse

### spec.sourceImage

`string | valueFrom`

Source: an image to copy, referenced as another GcpComputeImage (its
self_link) or a literal image self link (e.g.
"projects/debian-cloud/global/images/family/debian-12" for a public
image). Exactly one source. Immutable.

- references: GcpComputeImage (`status.outputs.self_link`)
- rule: write as {value: <literal>} or {valueFrom: {kind: GcpComputeImage, name: <that resource's name>, fieldPath: status.outputs.self_link}} -- a bare string does not parse

### spec.sourceSnapshot

`string`

Source: a snapshot to image (name or self link). Exactly one source.
Immutable.

### spec.rawDisk

`GcpComputeImageRawDisk`

Source: a raw disk tarball in Cloud Storage. Exactly one source.
Immutable.

### spec.rawDisk.source

`string` · required

The tarball's full Cloud Storage URL, e.g.
"https://storage.googleapis.com/my-bucket/images/web-base.tar.gz".
The archive holds one file named disk.raw. Required.

- rule: {"required":true}

### spec.rawDisk.sha1

`string`

The archive's SHA-1 checksum, base64-encoded; Google verifies it when
set.

### spec.rawDisk.containerType

`string`

The archive format. "TAR" is the only format Google accepts and the
default.

- rule: container_type must be TAR

### spec.kmsKey

`string | valueFrom`

Customer-managed encryption key (CMEK) for the image, referenced as a
GcpKmsKey or a GcpKmsKeyHandle (Autokey serves images). The Compute
Engine service agent
(service-<project-number>@compute-system.iam.gserviceaccount.com)
must hold roles/cloudkms.cryptoKeyEncrypterDecrypter on the key. When
omitted, Google-managed encryption is used. Immutable.

- references: GcpKmsKey (`status.outputs.key_id`), GcpKmsKeyHandle (`status.outputs.kms_key`)
- rule: write as {value: <literal>} or {valueFrom: {kind: GcpKmsKey, name: <that resource's name>, fieldPath: status.outputs.key_id}} -- a bare string does not parse

### spec.kmsKeyServiceAccount

`string`

Service account used for the encryption request of kms_key. When
omitted, the Compute Engine service agent is used. Only meaningful
together with kms_key. Immutable.

### spec.sourceDiskEncryption

`GcpComputeImageSourceEncryption`

Decrypts the source disk when it is itself CMEK-encrypted. Only valid
together with source_disk.

### spec.sourceDiskEncryption.kmsKey

`string | valueFrom` · required

The KMS key the source was encrypted with, referenced as a GcpKmsKey
or a literal self link. The service agent performing the read needs
roles/cloudkms.cryptoKeyEncrypterDecrypter on it.

- references: GcpKmsKey (`status.outputs.key_id`)
- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: GcpKmsKey, name: <that resource's name>, fieldPath: status.outputs.key_id}} -- a bare string does not parse

### spec.sourceDiskEncryption.kmsKeyServiceAccount

`string`

Service account used for the decryption request. When omitted, the
Compute Engine service agent is used.

### spec.sourceImageEncryption

`GcpComputeImageSourceEncryption`

Decrypts the source image when it is itself CMEK-encrypted. Only valid
together with source_image.

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
Compute Engine service agent is used.

### spec.sourceSnapshotEncryption

`GcpComputeImageSourceEncryption`

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
Compute Engine service agent is used.

### spec.diskSizeGb

`int64` · optional (explicit presence)

The image's size in GB, at least the source's size. Unset takes the
source's size. Immutable.

- rule: {"int64":{"gt":"0"}}

### spec.guestOsFeatures

`[]string`

Guest OS features VMs booted from this image may use, e.g.
["UEFI_COMPATIBLE", "SECURE_BOOT", "GVNIC", "VIRTIO_SCSI_MULTIQUEUE",
"SEV_CAPABLE", "SEV_SNP_CAPABLE", "TDX_CAPABLE", "IDPF",
"MULTI_IP_SUBNET", "SUSPEND_RESUME_COMPATIBLE", "WINDOWS"]. The
accepted set evolves with GCP -- see "Enabling guest operating system
features" in the Compute Engine docs. Unset inherits the source's
features. Immutable.

### spec.licenses

`[]string`

License URIs the image carries, e.g. a Windows Server or SQL Server
license. Unset inherits the source's licenses; set explicitly when
importing a raw disk that needs bring-your-own-license attribution.
Immutable.

### spec.storageLocations

`[]string`

Where the image's data is stored: one multi-region ("us", "eu",
"asia") or one region. Unset takes the multi-region nearest the
source. Immutable.

### spec.shieldedInstanceInitialState

`GcpComputeImageShieldedInstanceInitialState`

The UEFI Secure Boot keys VMs booted from this image start with. Unset
takes Google's default certificates. Immutable.

### spec.shieldedInstanceInitialState.pk

`GcpComputeImageFileContentBuffer`

The Platform Key, which authorizes changes to the key exchange keys.

### spec.shieldedInstanceInitialState.pk.content

`string` · required

The raw content, base64-encoded. Public certificates, not secrets.
Required.

- rule: {"required":true}

### spec.shieldedInstanceInitialState.pk.fileType

`string`

"X509" (a certificate) or "BIN" (raw bytes). Empty lets Google infer.

- rule: file_type must be X509 or BIN

### spec.shieldedInstanceInitialState.keks

`[]GcpComputeImageFileContentBuffer`

Key Exchange Keys, which authorize changes to db and dbx.

### spec.shieldedInstanceInitialState.keks[].content

`string` · required

The raw content, base64-encoded. Public certificates, not secrets.
Required.

- rule: {"required":true}

### spec.shieldedInstanceInitialState.keks[].fileType

`string`

"X509" (a certificate) or "BIN" (raw bytes). Empty lets Google infer.

- rule: file_type must be X509 or BIN

### spec.shieldedInstanceInitialState.dbs

`[]GcpComputeImageFileContentBuffer`

The allowed signature database: certificates of boot loaders and
drivers allowed to run.

### spec.shieldedInstanceInitialState.dbs[].content

`string` · required

The raw content, base64-encoded. Public certificates, not secrets.
Required.

- rule: {"required":true}

### spec.shieldedInstanceInitialState.dbs[].fileType

`string`

"X509" (a certificate) or "BIN" (raw bytes). Empty lets Google infer.

- rule: file_type must be X509 or BIN

### spec.shieldedInstanceInitialState.dbxs

`[]GcpComputeImageFileContentBuffer`

The forbidden signature database: revoked certificates and hashes.

### spec.shieldedInstanceInitialState.dbxs[].content

`string` · required

The raw content, base64-encoded. Public certificates, not secrets.
Required.

- rule: {"required":true}

### spec.shieldedInstanceInitialState.dbxs[].fileType

`string`

"X509" (a certificate) or "BIN" (raw bytes). Empty lets Google infer.

- rule: file_type must be X509 or BIN

### spec.resourceManagerTags

`map<string, string>`

Resource Manager tags bound to the image for org-policy and IAM
conditions. Keys in the form "tagKeys/{id}", values "tagValues/{id}".
Immutable.

### spec.labels

`map<string, string>`

User labels merged with Planton attribution labels (which win on key
conflicts). The one setting that updates in place.

### spec.deletionPolicy

`string`

Deletion policy -- what happens when this resource is destroyed:
  ""        -- same as "DELETE" (provider default)
  "DELETE"  -- the image is deleted
  "PREVENT" -- destroy FAILS; a guard for an image fleets boot from
  "ABANDON" -- the image is removed from management but kept in GCP

- rule: deletion_policy must be one of: DELETE, PREVENT, ABANDON

## Validation Rules

- `image_exactly_one_source`: an image is built from exactly one source: source_disk, source_image, source_snapshot, or raw_disk
- `source_disk_encryption_requires_source_disk`: source_disk_encryption is only valid together with source_disk
- `source_image_encryption_requires_source_image`: source_image_encryption is only valid together with source_image
- `source_snapshot_encryption_requires_source_snapshot`: source_snapshot_encryption is only valid together with source_snapshot
- `kms_key_service_account_requires_kms_key`: kms_key_service_account is only meaningful together with kms_key

## Outputs

Reference an output from another manifest as `valueFrom: {kind: GcpComputeImage, name: <resource-name>, fieldPath: status.outputs.<output>}`.

| Output | Type | Description |
|---|---|---|
| `status.outputs.name` | `string` | Name of the image in GCP. |
| `status.outputs.self_link` | `string` | Self-link URL of the image -- what a disk's image, an instance's boot disk image, or another image's source_image consumes. |
| `status.outputs.family` | `string` | The image's family, or empty when it has none. Consumers that should always boot the newest build use "projects/{project}/global/images/family/{family}". |
| `status.outputs.disk_size_gb` | `int32` | The image's size in GB. |
| `status.outputs.image_id` | `string` | The image's resource ID: projects/{project}/global/images/{name} -- the relative form, without self_link's https://www.googleapis.com/compute/v1/ prefix. Consume self_link where a field takes an image URL (a disk's image, an instance's boot disk, a managed instance group's source_image, another image's source_image); consume image_id where Google documents the relative form only, such as a GKE node pool's secondary boot disk image. |

## References

Fields that can point at another resource's outputs:

| Field | Kind | Output |
|---|---|---|
| `spec.projectId` | GcpProject | `status.outputs.project_id` |
| `spec.sourceDisk` | GcpComputeDisk | `status.outputs.self_link` |
| `spec.sourceImage` | GcpComputeImage | `status.outputs.self_link` |
| `spec.kmsKey` | GcpKmsKey | `status.outputs.key_id` |
| `spec.kmsKey` | GcpKmsKeyHandle | `status.outputs.kms_key` |
| `spec.sourceDiskEncryption.kmsKey` | GcpKmsKey | `status.outputs.key_id` |
| `spec.sourceImageEncryption.kmsKey` | GcpKmsKey | `status.outputs.key_id` |
| `spec.sourceSnapshotEncryption.kmsKey` | GcpKmsKey | `status.outputs.key_id` |

## Referenced By

Fields on other kinds that can point at this resource:

| Kind | Field | Reads |
|---|---|---|
| GcpComputeDisk | `spec.image` | `status.outputs.self_link` |
| GcpComputeImage | `spec.sourceImage` | `status.outputs.self_link` |
| GcpComputeInstance | `spec.bootDisk.image` | `status.outputs.self_link` |
| GcpComputeMig | `spec.template.disks[].sourceImage` | `status.outputs.self_link` |
| GcpGkeNodePool | `spec.nodeConfig.secondaryBootDisks[].diskImage` | `status.outputs.image_id` |

## See Also

- [Overview](../README.md)
