# GCP Compute Disk

Deploys a zonal Google Compute Engine persistent disk (`google_compute_disk`) — the durable block device behind stateful VMs: database volumes, shared read-only datasets, bootable golden images, and any data that must outlive the instance it is attached to.

## Overview

A disk is a first-class node with its own lifecycle: create it once, attach it to a `GcpComputeInstance` by reference, and the data survives instance replacement, resizing, and rescheduling. The spec covers the full zonal-disk surface:

- **Sources** — at most one of `image` (bootable; a `GcpComputeImage` reference or a family/image path), `sourceSnapshot` (restore), or `sourceDisk` (clone another `GcpComputeDisk`); omit all three for an empty data disk (the common case, which then requires `sizeGb`).
- **Performance** — the pd family (`pd-standard`, `pd-balanced`, `pd-ssd`, `pd-extreme`) and the hyperdisk family, with `provisionedIops` and `provisionedThroughput` tunable in place on hyperdisk types (at most every 4 hours).
- **Encryption** — customer-managed keys (CMEK) via a `GcpKmsKey` reference; confidential-compute mode on hyperdisk SKUs.
- **Safety** — `createSnapshotBeforeDestroy` takes a final snapshot during destroy, a last-resort recovery net for precious volumes.

`diskName`, `zone`, `type`, sources, encryption, and `architecture` are create-time decisions — changing them replaces the disk and its data. `sizeGb` grows in place but never shrinks. Deleting a disk still attached to a running instance fails; detach first (or delete the instance).

## When to Use

- **Database volumes** — data disks for self-managed PostgreSQL/MySQL/etc. that must survive VM replacement
- **Shared datasets** — read-only volumes attached to many instances (`accessMode: READ_ONLY_MANY`)
- **Golden images** — bootable disks built from an image family and cloned or snapshotted for fleets
- **High-performance storage** — hyperdisk volumes with independently provisioned IOPS and throughput

## Prerequisites

- GCP credentials for the deploying principal — see [`iac/permissions.yaml`](iac/permissions.yaml) for the least-privilege permission set (the Compute Engine API is enabled automatically)
- For CMEK: a `GcpKmsKey` whose key the Compute Engine service agent (`service-<project-number>@compute-system.iam.gserviceaccount.com`) can use (`roles/cloudkms.cryptoKeyEncrypterDecrypter`)

## Quick Start

```yaml
apiVersion: gcp.planton.dev/v1alpha1
kind: GcpComputeDisk
metadata:
  name: app-data
spec:
  zone: us-central1-a
  type: pd-balanced
  sizeGb: 100
```

This creates an empty 100 GB pd-balanced data disk named `app-data` (the disk name falls back to `metadata.name`). Attach it to a `GcpComputeInstance` via `attachedDisks[].source` referencing this resource's `self_link` output.

## Configuration Reference

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `projectId` | StringValueOrRef | No | Owning project (reference a `GcpProject`); empty uses the provider default. Immutable |
| `diskName` | string | No | Disk name (1-63 chars, lowercase/digits/hyphens); falls back to `metadata.name`. Immutable |
| `zone` | string | Yes | Zone, e.g. `us-central1-a`. A disk attaches only to instances in the same zone. Immutable |
| `description` | string | No | Human-readable description |
| `type` | string | No | `pd-standard`, `pd-balanced` (GCP default), `pd-ssd`, `pd-extreme`, or `hyperdisk-*`. Immutable |
| `sizeGb` | int | Empty disks | Size in GB (1–65536). Required with no source; grows in place, never shrinks |
| `image` | StringValueOrRef | No | Source image — a `GcpComputeImage` reference (its `self_link`) or a literal family or image path — makes the disk bootable. Create-time only |
| `sourceSnapshot` | string | No | Snapshot (name or self link) to restore from. Create-time only |
| `sourceDisk` | StringValueOrRef | No | Existing `GcpComputeDisk` to clone (or a literal self link). Create-time only |
| `kmsKey` | StringValueOrRef | No | CMEK key (reference a `GcpKmsKey`); omitted means Google-managed encryption. Immutable |
| `provisionedIops` | int | pd-extreme, hyperdisk-extreme | Provisioned IOPS; tunable in place on hyperdisk types (every 4h at most) |
| `provisionedThroughput` | int | No | Provisioned MB/s on `hyperdisk-throughput`/`hyperdisk-balanced`; updates in place (every 4h at most) |
| `accessMode` | string | No | `READ_WRITE_SINGLE` (default), `READ_WRITE_MANY`, `READ_ONLY_MANY` |
| `architecture` | string | No | `X86_64` or `ARM64`; normally inferred from the image. Immutable |
| `enableConfidentialCompute` | bool | No | Confidential-compute disk (hyperdisk SKUs; requires `kmsKey`) |
| `physicalBlockSizeBytes` | int | No | 4096 (default) or 16384 |
| `createSnapshotBeforeDestroy` | bool | No | Take a final snapshot immediately before destroy — a last-resort recovery net |
| `snapshotBeforeDestroyPrefix` | string | No | Custom name prefix for that final snapshot |
| `storagePool` | string | No | Hyperdisk storage pool URL or name to create the disk in |
| `labels` | map | No | User labels, merged beneath platform attribution labels |
| `resourceManagerTags` | map | No | `tagKeys/{id}` → `tagValues/{id}` bindings for org policy and IAM conditions. Create-time only |
| `sourceInstantSnapshot` | string | No | Instant snapshot to create the disk from — the fast, same-region restore path. Create-time only |
| `sourceStorageObject` | string | No | Cloud Storage URI of a raw disk image to import (skips the intermediate compute image). Create-time only |
| `guestOsFeatures` | list(string) | No | Guest OS features for bootable disks (e.g. `UEFI_COMPATIBLE`, `SECURE_BOOT`, `GVNIC`). Create-time only |
| `licenses` | list(string) | No | License URIs for BYOL imports. Create-time only |
| `kmsKeyServiceAccount` | string | No | Service account for the CMEK encryption request; empty uses the Compute Engine default agent |
| `sourceImageEncryption` | object | No | CMEK key (+ optional service account) that decrypts an encrypted source `image` |
| `sourceSnapshotEncryption` | object | No | CMEK key (+ optional service account) that decrypts an encrypted `sourceSnapshot` |
| `asyncPrimaryDisk` | StringValueOrRef | No | Primary `GcpComputeDisk` this disk replicates from (makes this the async-replication secondary). Create-time only |
| `deletionPolicy` | string | No | `PREVENT` fails destroys; `ABANDON` unmanages without deleting; default `DELETE` |

At most one source (`image` / `sourceSnapshot` / `sourceInstantSnapshot` / `sourceStorageObject` / `sourceDisk`) may be set — enforced at validation time, before anything deploys. Customer-supplied (raw CSEK) keys are deliberately not modeled: raw key material never flows through manifests or state — use CMEK.

## Outputs

| Output | Description |
|--------|-------------|
| `name` | Name of the disk in GCP |
| `disk_id` | Server-assigned unique numeric identifier |
| `self_link` | Self-link URL — the composition key a `GcpComputeInstance`'s `bootDisk.sourceDisk` or `attachedDisks[].source` consumes |
| `zone` | Zone the disk lives in (plain zone name) |
| `size_gb` | Provisioned size in GB |
| `type` | Disk type (plain type name, e.g. `pd-balanced`) |

See the [presets](presets/) for remixable starting points and GUIDE.md for the deep dive.

---

© Planton. Licensed under [Apache-2.0](https://github.com/plantonhq/planton/blob/main/LICENSE).
