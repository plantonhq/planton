# GcpComputeDisk Guide

The judgment this guide protects: a disk's identity-shaping fields (name,
zone, type, source, encryption, architecture) are replace-on-change, and a
replaced disk is replaced DATA. The spec's create-time guards
(`createSnapshotBeforeDestroy`, `deletionPolicy`) exist because the mistake
they catch is unrecoverable.

## Choosing a source

At most one source; none creates an empty volume (the common case — declare
`sizeGb`). The five sources are different restore paths, not synonyms:
`image` makes the disk bootable; `sourceSnapshot` restores durable,
cross-region snapshot data; `sourceInstantSnapshot` is the fast path but
same-region only; `sourceStorageObject` imports a raw disk file straight
from Cloud Storage without minting an intermediate compute image;
`sourceDisk` clones another `GcpComputeDisk` (same-zone data copy). Pick by
where the bytes live today and how fast the restore must be.

## Encryption posture

CMEK only. `kmsKey` (with optional `kmsKeyServiceAccount`) encrypts the
disk; `sourceImageEncryption` / `sourceSnapshotEncryption` decrypt encrypted
sources. Customer-supplied raw keys (CSEK) are deliberately NOT modeled —
the provider keeps those arguments in state as ordinary values, and key material
flowing through manifests contradicts the platform's secret posture; the
recorded exclusions live in this kind's parity manifest. If a workload
genuinely requires CSEK, that is a platform-level conversation, not a field
request. Before the first CMEK apply, the Compute Engine service agent needs
`roles/cloudkms.cryptoKeyEncrypterDecrypter` on the key — missing it fails
at apply with a KMS permission error, not at plan.

## Destroy-safety layering

Two independent nets, use them deliberately: `createSnapshotBeforeDestroy`
allows the destroy but leaves a timestamped snapshot (soft net — costs a
snapshot, saves the data); `deletionPolicy: PREVENT` fails the destroy
outright (hard net — the disk and anything orchestrating its teardown stop
there). PREVENT blocks full-lifecycle automation by design; prefer the
snapshot net for volumes that must still be tear-down-able.

## Async replication

`asyncPrimaryDisk` makes THIS disk the secondary of a primary in another
region — set it on the DR-side disk, matching the primary's size and type.
Creating the pair does not start replication: activation is an operation on
the primary (outside this component's surface today), so treat the field as
the pairing declaration, not the running replication.

## On the diagram

The disk is a first-class node; a `GcpComputeInstance` consumes its
`self_link` output via `attachedDisks[].source` or `bootDisk.sourceDisk`,
rendering the attachment as a visible edge. An `asyncPrimaryDisk` reference
renders the DR pairing as an edge between two disk nodes — one per region.

## Pairs well with

- `GcpComputeInstance` — the attachment consumer of `self_link`.
- `GcpKmsKey` — CMEK for the disk and its encrypted sources.
- `GcpComputeImage` — the golden image a bootable disk starts from
  (`image`, its `self_link`).
- A second `GcpComputeDisk` in another region — the async-replication
  primary.
