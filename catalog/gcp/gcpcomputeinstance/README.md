# GCP Compute Instance

Deploys a Google Compute Engine virtual machine (`google_compute_instance`) — the general-purpose compute node behind databases-on-VM, stateful application servers, GPU workers, and bastion hosts — with first-class references to the disks, addresses, networks, and service accounts it composes with.

## Overview

The spec covers the full single-instance surface:

- **Boot and data** — exactly one boot source (an image for a fresh install, a snapshot for a restore, or an existing bootable `GcpComputeDisk`); regional boot disks replicated across two zones (`replicaZones`); guest OS features; attached data disks are first-class `GcpComputeDisk` resources referenced by self link, so they carry their own lifecycle and survive the VM (regional disks can be force-attached during failover). Ephemeral local-SSD scratch disks are modeled separately.
- **Networking** — one or more vNICs, each on a VPC network or subnetwork (referenceable as `GcpVpcNetwork` / `GcpSubnetwork`) or a Private Service Connect `networkAttachment` on its own, static internal IPs and external NAT IPs as `GcpAddress` references, alias IP ranges, dual-stack/IPv6 with static internal and external IPv6 addresses, VLAN-tagged dynamic sub-interfaces, IGMP multicast queries, gVNIC, and TIER_1 egress bandwidth.
- **Cost and lifecycle** — Spot and FLEX_START provisioning via a single switch (`scheduling.provisioningModel`; both engines derive the API's legacy preemptible flag for Spot and force automatic restart off), reservation-bound capacity, run-duration limits, `desiredStatus` to start (`RUNNING`), suspend (`SUSPENDED`), or stop (`TERMINATED`) the VM in place — compute billing stops while terminated; disks keep billing — and `deletionPolicy` to guard (`PREVENT`) or orphan (`ABANDON`) the VM at destroy time.
- **Security** — Shielded VM (secure boot, vTPM, integrity monitoring), Confidential VM (SEV / SEV_SNP / TDX; requires `onHostMaintenance: TERMINATE`), CMEK encryption at every level via `GcpKmsKey` (boot disk, attached disks, encrypted image/snapshot sources, and instance-level state), and a dedicated service-account identity referenced as `GcpServiceAccount`.
- **Placement** — GPUs (also require `TERMINATE` maintenance), sole-tenant node affinities, reservation affinity, and minimum CPU platform.

Zone, boot source, NIC count and their networks, scratch disks, hostname, confidential mode, and reservation affinity are create-time decisions — changing them replaces the VM. Machine type, service account, shielded config, and several others update by stopping the VM, which the provider only does when `allowStoppingForUpdate` is true (recommended).

`deletionProtection` defaults to false: it guards only the VM object. Data protection lives on the disks — the boot disk's `autoDelete` and each `GcpComputeDisk`'s own lifecycle.

## When to Use

- **Databases and stateful servers on VMs** — attach durable `GcpComputeDisk` data disks with `deletionProtection` on the instance
- **Bastion / jump hosts** — small shielded VM, OS Login, no external IP
- **Dev/test and batch workers** — Spot VMs at deep discount for interruption-tolerant work
- **GPU and specialized compute** — guest accelerators, local SSD scratch, sole tenancy

## Prerequisites

- GCP credentials for the deploying principal — see [`iac/permissions.yaml`](iac/permissions.yaml) for the least-privilege permission set (the Compute Engine API is enabled automatically)
- For CMEK boot disks: a `GcpKmsKey` the Compute Engine service agent can use (`roles/cloudkms.cryptoKeyEncrypterDecrypter`)

## Quick Start

```yaml
apiVersion: gcp.planton.dev/v1alpha1
kind: GcpComputeInstance
metadata:
  name: dev-vm
spec:
  zone: us-central1-a
  machineType: e2-medium
  bootDisk:
    image: debian-cloud/debian-12
  networkInterfaces:
    - network:
        value: default
      accessConfigs:
        - ephemeral: true
  scheduling:
    provisioningModel: SPOT
    instanceTerminationAction: DELETE
```

This creates a Spot Debian 12 VM on the default network with an ephemeral external IP — the cheapest way to get a shell on GCP. When omitted, the instance name falls back to `metadata.name`.

## Configuration Reference

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `projectId` | StringValueOrRef | No | Owning project (reference a `GcpProject`); empty uses the provider default. Immutable |
| `instanceName` | string | No | Instance name; falls back to `metadata.name`. Immutable |
| `zone` | string | Yes | Zone, e.g. `us-central1-a`. Immutable |
| `machineType` | string | Yes | e.g. `e2-medium`, `n2-standard-4`, `custom-6-20480`. Changing it stops/restarts the VM |
| `description` | string | No | Human-readable description |
| `hostname` | string | No | Custom FQDN; default `<name>.c.<project>.internal`. Immutable |
| `bootDisk` | object | Yes | Exactly one source: `image`, `sourceSnapshot`, or `sourceDisk` (a `GcpComputeDisk` reference); plus size, type, `autoDelete`, CMEK `kmsKey` (+`kmsKeyServiceAccount`), hyperdisk tuning, `guestOsFeatures`, regional `replicaZones` (exactly two), `resourceManagerTags`, attachment `mode`/`interface`/`forceAttach`, and CMEK decryption of encrypted sources (`sourceImageEncryption` / `sourceSnapshotEncryption`) |
| `attachedDisks[]` | list | No | Existing `GcpComputeDisk` references (`source`, `deviceName`, `mode`, `kmsKey` + `kmsKeyServiceAccount`, `forceAttach` for regional-disk takeover) — each disk keeps its own lifecycle |
| `scratchDisks[]` | list | No | Ephemeral local SSDs (`interface` NVME/SCSI, 375 GB units). Contents lost on stop/preemption. Create-time only |
| `networkInterfaces[]` | list | Yes | Each needs an attachment point: `network` (`GcpVpcNetwork`), `subnetwork` (`GcpSubnetwork`), or a PSC `networkAttachment` alone; static `networkIp` references `GcpAddress`; `accessConfigs[].natIp` (external IP, also `GcpAddress`), IPv6 (`stackType`, `ipv6Address`, `ipv6AccessConfigs[].externalIpv6`), `nicType`, `vlan` (dynamic sub-interface), `igmpQuery`, `queueCount`, alias ranges |
| `serviceAccount` | object | No | `email` (a `GcpServiceAccount` reference) + `scopes` (modern practice: single `cloud-platform` scope, control via IAM). Omitted = Compute Engine default account |
| `scheduling` | object | No | `provisioningModel` (STANDARD/SPOT/FLEX_START/RESERVATION_BOUND), `automaticRestart`, `onHostMaintenance` (MIGRATE/TERMINATE), `instanceTerminationAction` (STOP/DELETE — for reclaimable models and timed-run VMs), run-duration limits, sole-tenant node affinities |
| `shieldedInstanceConfig` | object | No | `enableSecureBoot` (GCP default false), `enableVtpm`, `enableIntegrityMonitoring` (both default true) |
| `confidentialInstanceConfig` | object | No | `confidentialInstanceType`: SEV / SEV_SNP / TDX. Requires `onHostMaintenance: TERMINATE`. Create-time only |
| `advancedMachineFeatures` | object | No | Nested virtualization, threads per core, visible core count, UEFI networking, PMU, turbo mode |
| `guestAccelerators[]` | list | No | GPU `type` + `count`. Requires `onHostMaintenance: TERMINATE` |
| `reservationAffinity` | object | No | `ANY_RESERVATION` / `SPECIFIC_RESERVATION` / `NO_RESERVATION` |
| `totalEgressBandwidthTier` | string | No | `DEFAULT` or `TIER_1` (gVNIC + ≥30 vCPUs on supported shapes) |
| `metadata` | map | No | Guest metadata (e.g. `enable-oslogin`, `startup-script-url`) |
| `startupScript` | string | No | Runs on every boot (dedicated metadata surface) |
| `sshKeys[]` | list | No | `username:ssh-rsa ...` entries, folded into metadata `ssh-keys` identically on both engines. Ignored under OS Login |
| `labels` | map | No | User labels, merged beneath platform attribution labels |
| `tags[]` | list | No | Network tags used by firewall rules and routes |
| `resourceManagerTags` | map | No | `tagKeys/{id}` → `tagValues/{id}` bindings. Create-time only |
| `resourcePolicies[]` | list | No | Attached resource-policy self links (max 1, e.g. an instance schedule) |
| `minCpuPlatform` | string | No | e.g. `Intel Ice Lake`, `AMD Milan` |
| `canIpForward` | bool | No | Required for router/NAT/VPN VMs. Create-time only |
| `enableDisplay` | bool | No | Virtual display device |
| `deletionProtection` | bool | No | Guards the VM object only (default false — data levers live on disks) |
| `desiredStatus` | string | No | `RUNNING`, `SUSPENDED`, or `TERMINATED` — starts/suspends/stops in place |
| `allowStoppingForUpdate` | bool | No | Let the provider stop/restart for updates that need it (recommended true) |
| `keyRevocationActionType` | string | No | `NONE` or `STOP` when a protecting KMS key is revoked |
| `instanceEncryptionKey` | object | No | Instance-level CMEK (`kmsKey` reference + optional `kmsKeyServiceAccount`), distinct from per-disk keys. Create-time only |
| `deletionPolicy` | string | No | `DELETE` (default), `PREVENT` (destroy fails), or `ABANDON` (VM left running, removed from management) |

## Stack Outputs

| Output | Description |
|--------|-------------|
| `instance_name` | Name of the instance |
| `instance_id` | Server-assigned unique numeric identifier |
| `self_link` | API self link — what consumers reference |
| `internal_ip` | Primary internal IP (first interface) |
| `external_ip` | External IP of the first interface; `""` when the VM is private |
| `status` | Current status (RUNNING, TERMINATED, ...) |
| `zone` | Zone where the instance runs |
| `machine_type` | Machine type |
| `cpu_platform` | CPU platform the instance landed on |

See the [presets](presets/) for remixable starting points and GUIDE.md for the deep dive.

---

© Planton. Licensed under [Apache-2.0](https://github.com/plantonhq/planton/blob/main/LICENSE).
