# GcpTpuVm Guide

The judgment this guide protects: TPU capacity is scarce and bills every hour it exists, and almost every choice is fixed at creation. Pick the zone and capacity model for how the job tolerates interruption, checkpoint relentlessly on spot, and destroy the slice when the job ends.

## Beta-only by Google

Google publishes `google_tpu_v2_vm` only in its beta Terraform provider (the GA `google_tpu_node` was removed in provider 7.0). The Terraform module attaches `provider = google-beta` to that one resource under the catalog's recorded admission; Pulumi serves it from its single provider. When Google promotes it, the module moves to the GA provider.

## Choosing the slice

- `acceleratorType` names the slice: generation and chip (or core) count -- `v2-8`, `v3-8`, `v4-8`, `v5litepod-8`, `v5p-8`, `v6e-8`, and larger.
- `acceleratorConfig` names it by generation (`V2`, `V3`, `V4`, `V5LITE_POD`, `V5P`, `V6E`) and chip topology (`2x2`, `2x2x1`, ...), for shapes the names do not cover.
- Set one or neither; with neither, the provider asks Google for a `v2-8`.
- `runtimeVersion` must match the generation (for example `tpu-ubuntu2204-base` on v2/v3/v4, `v2-alpha-tpuv5-lite` on v5e, `v2-alpha-tpuv6e` on v6e).

## Capacity: zone, quota, and model

Each generation exists in a handful of zones, and a project needs quota for it there. Then choose how capacity is obtained:

- **On-demand** (no `schedulingConfig`) -- yours until you delete it, if Google has chips to give.
- **Spot** (`spot: true`) -- far cheaper; Google can reclaim it at any time. Checkpoint to a data disk or Cloud Storage and design the job to resume.
- **Preemptible** -- the older reclaimable model, ended after 24 hours; prefer spot.
- **Reserved** (`reserved: true`) -- draws from a reservation made for the project and zone.

When a create fails for lack of capacity, a `GcpTpuQueuedResource` asks for the same slice and waits in Google's queue until it exists.

## Networking

`networkConfig` (one interface) or `networkConfigs` (several, for multi-NIC hosts) picks the network and subnetwork (default: the `default` network). With `enableExternalIps: false`, the subnetwork needs Private Google Access or Cloud NAT to reach Google APIs and package mirrors. `cidrBlock` (a /29) pins where addresses come from; `tags` let firewall rules target the hosts.

## Identity, disks, and startup

`serviceAccount.email` sets the hosts' identity (default: the Compute Engine default account with every Cloud API; its IAM roles govern access). `dataDisks` attach existing persistent disks -- `READ_ONLY` to share a dataset across many TPUs, `READ_WRITE` (one TPU at a time) for checkpoints -- and can change in place. `metadata.startup-script` runs on every worker at boot.

## What changes in place

Only `description`, `labels`, `metadata`, `tags`, and `dataDisks`. Everything else -- zone, id, runtime, accelerator, network, identity, scheduling, Secure Boot -- replaces the TPU.

## Destroy

A TPU bills every hour it exists, idle or not. Destroy it when the job is done; attached data disks are only detached. `ABANDON` keeps it running -- and billing -- outside management.
