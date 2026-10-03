# GCP Compute Image

Creates a Compute Engine custom image: the golden boot image VMs, instance templates, managed instance groups, and disks start from. Built from exactly one source -- a `GcpComputeDisk`, another image, a snapshot, or a raw disk tarball in Cloud Storage -- and rolled forward through image families.

## What Gets Created

When you deploy this Infra Component, the IaC module provisions:

- **API enablement** -- `compute.googleapis.com` on the project (never disabled on destroy)
- **Image** -- one `compute_image`

## Before You Deploy

### Planton Setup

- **GCP Provider Connection** -- an active connection in the Connect module with `roles/compute.storageAdmin` (or the image permissions in `iac/permissions.yaml`) on the project, and read access to the source.
- **Planton Runner** -- required when using Runner-based credential delivery. Not needed for inline credentials or browser OAuth authentication modes.

### Optional Dependencies

- **`GcpComputeDisk`** -- the source disk (`sourceDisk`, its `self_link`).
- **`GcpComputeImage`** -- a source image (`sourceImage`).
- **`GcpKmsKey`** / **`GcpKmsKeyHandle`** -- a customer-managed key (`kmsKey`).
- **`GcpProject`** -- the image's project (`projectId`).

## Deploy

### CLI

```yaml
apiVersion: gcp.planton.dev/v1alpha1
kind: GcpComputeImage
metadata:
  name: web-base-20261001
spec:
  family: web-base
  sourceDisk:
    valueFrom:
      kind: GcpComputeDisk
      name: web-build-disk
      fieldPath: status.outputs.self_link
```

```shell
planton apply -f compute-image.yaml
```

## Configuration Reference

### Optional Fields

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `projectId` | `string` / ref | provider project | The image's project (`GcpProject` ref). |
| `imageName` | `string` | `metadata.name` | One name per build, e.g. `web-base-20261001`. |
| `description` | `string` | -- | |
| `family` | `string` | -- | The stable name consumers boot from. |
| `sourceDisk` / `sourceImage` / `sourceSnapshot` / `rawDisk` | one of | -- | Exactly one source. `rawDisk` is `source` (a Cloud Storage URL), `sha1`, `containerType` (`TAR`). |
| `kmsKey` / `kmsKeyServiceAccount` | `string` / ref | Google-managed | Customer-managed key (`GcpKmsKey` or `GcpKmsKeyHandle`). |
| `sourceDiskEncryption` / `sourceImageEncryption` / `sourceSnapshotEncryption` | object | -- | The key a CMEK-encrypted source was encrypted with; only with that source. |
| `diskSizeGb` | `int64` | the source's size | |
| `guestOsFeatures` / `licenses` / `storageLocations` | `string[]` | inherited / nearest multi-region | |
| `shieldedInstanceInitialState` | object | Google's certificates | Secure Boot `pk`, `keks`, `dbs`, `dbxs` (`content`, `fileType` `X509` or `BIN`). |
| `resourceManagerTags` | `map` | -- | `tagKeys/{id}` to `tagValues/{id}`. |
| `labels` | `map` | -- | The one setting that updates in place. |
| `deletionPolicy` | `string` | `DELETE` | `DELETE`, `PREVENT`, or `ABANDON`. |

### Validation Rules

- Exactly one source; a source decryption key only with its source; `kmsKeyServiceAccount` only with `kmsKey`.
- `imageName` and `family` are 1-63 lowercase letters, digits, or hyphens, starting with a letter.

## Outputs

| Output | Type | Description |
|--------|------|-------------|
| `name` | `string` | The image's name |
| `self_link` | `string` | What disks, instances, and other images boot or copy from |
| `family` | `string` | The image's family, or empty |
| `disk_size_gb` | `int32` | The image's size in GB |
| `image_id` | `string` | `projects/{project}/global/images/{name}` -- the relative form GKE node pools' secondary boot disks consume |

## Deployment Methods

### Pulumi (Go)

See [`iac/pulumi/README.md`](iac/pulumi/README.md) for Pulumi-specific deployment instructions.

### Terraform

See [`iac/tf/README.md`](iac/tf/README.md) for Terraform-specific deployment instructions.

## Important Notes

- **Version, don't edit.** Everything except labels replaces the image. Give each build its own `imageName` and the same `family`; consumers booting `projects/{project}/global/images/family/{family}` always get the newest build.
- **Customer-supplied raw keys are not modeled.** Raw key material does not belong in manifests or state; use Cloud KMS keys, as `GcpComputeDisk` and `GcpComputeInstance` do.
- **Images bill storage until deleted**, whether or not any VM boots from them -- retire old builds.

## Examples

For a complete example, see `e2e/manifest.yaml`. Scenario variants live under `e2e/scenarios/`.

## Related Kinds

- **GcpComputeDisk** -- the source of a golden image, and a consumer of one
- **GcpComputeInstance**, **GcpComputeMig** -- boot from the image or its family
- **GcpKmsKey**, **GcpKmsKeyHandle** -- customer-managed encryption

---

© Planton. Licensed under [Apache-2.0](https://github.com/plantonhq/planton/blob/main/LICENSE).
