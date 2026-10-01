# GCP Compute Image

Builds a golden VM image once and reuses it everywhere: an operating system hardened and pre-loaded with your agents, captured from a configured disk, copied from a public image, or imported from a tarball. Each build is a new image in a family, so VMs that boot from the family always get the newest build.

## What Gets Created

When you deploy this Cloud Resource, the IaC module provisions:

- **API enablement** -- the Compute Engine API on the project
- **Image** -- one custom image

## Before You Deploy

### Planton Setup

- **GCP Provider Connection** -- an active connection in the Connect module with permission to create images in the project and read the source. Map it as the default for your environment, or specify it explicitly when creating the Cloud Resource.
- **Planton Runner** -- required when using Runner-based credential delivery. Not needed for inline credentials or browser OAuth authentication modes.

## Deploy

### Console

Open the deployment store, find **GCP Compute Image**, and click **Deploy**. The creation wizard walks you through preset selection, environment and connection configuration, and spec fields. Start from the **Golden Image from a Disk** preset in the [Presets](#presets) tab.

### CLI

Create a manifest and apply it:

```yaml
apiVersion: gcp.planton.dev/v1alpha1
kind: GcpComputeImage
metadata:
  name: web-base-20261001
  org: acme-corp
  env: prod
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

This captures the configured build disk as the newest image in the `web-base` family. A Stack Job tracks the provisioning in real time.

### InfraChart

Reference a disk's `status.outputs.self_link` from `sourceDisk`; consumers boot from this image's `status.outputs.self_link` or its family path.

## Key Configuration

These are the most important decisions when configuring an image. Explore the full field reference in the [API Explorer](#api-explorer) tab.

**Source** -- exactly one of a disk, an image, a snapshot, or a raw disk tarball.

**Family** -- the stable name consumers boot from; each build is a new image in it.

**Encryption** -- a Cloud KMS key or an Autokey key handle, or Google-managed by default.

## Outputs and Dependencies

### What This Component Consumes

| Dependency | Field | ValueFromRef Path |
|------------|-------|-------------------|
| **GcpProject** | `projectId` | `status.outputs.project_id` |
| **GcpComputeDisk** | `sourceDisk` | `status.outputs.self_link` |
| **GcpComputeImage** | `sourceImage` | `status.outputs.self_link` |
| **GcpKmsKey** | `kmsKey`, source encryption keys | `status.outputs.key_id` |
| **GcpKmsKeyHandle** | `kmsKey` | `status.outputs.kms_key` |

### What This Component Provides

After provisioning, `status.outputs` contains values that downstream Cloud Resources can consume via ValueFromRef:

| Output | Description | Common Downstream Use |
|--------|-------------|----------------------|
| `self_link` | The image's self link | A disk's or instance's boot image; another image's source |
| `family` | The image's family | Booting the newest build |
| `name` | The image's name | Tooling |
| `disk_size_gb` | The image's size | Sizing boot disks |

## Common Patterns

Browse the [Presets](#presets) tab for ready-to-deploy configurations.

**Golden image** -- start from the **Golden Image from a Disk** preset.

**Pinned OS** -- start from the **Copy a Public Image** preset.

**Regulated** -- start from the **CMEK Image** preset.

## Works With

- [**GCP Compute Disk**](/cloud-catalog/gcp-compute-disk) -- the build disk and the boot disks
- [**GCP Compute Instance**](/cloud-catalog/gcp-compute-instance) -- VMs that boot the image
- [**GCP KMS Key Handle**](/cloud-catalog/gcp-kms-key-handle) -- Autokey encryption
