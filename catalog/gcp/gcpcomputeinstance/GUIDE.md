# GcpComputeInstance Guide

The judgment this guide protects: a VM is a COMPUTE node, not a data store.
Its identity-shaping fields (zone, boot source, NICs and their networks,
scratch disks, hostname, confidential mode, reservation affinity) are
replace-on-change, and everything durable belongs on disks that outlive
it. Configure the instance as if it will be replaced — because sooner or
later, it will be.

## The stop-to-update contract

GCP updates fall into three classes: in-place (labels, metadata,
`desiredStatus`), stop-and-restart (machine type, service account,
shielded config — the provider performs these only when
`allowStoppingForUpdate` is true), and replace (the identity fields
above). The recommended posture is `allowStoppingForUpdate: true` for any
VM whose brief restart is acceptable; without it, a machine-type change
fails the apply instead of causing downtime, which is the right default
only for VMs where an unplanned stop is worse than a failed deploy.
`desiredStatus: TERMINATED` is the other lever worth knowing: it stops
compute billing without destroying anything — disks keep billing, the
VM's identity and addresses survive.

## Provisioning-model economics

Four models, three different bargains:

- `STANDARD` — on-demand, full price, no strings.
- `SPOT` — 60–91% off; GCP may reclaim the VM at ANY time. The spec's
  single switch derives the API's legacy preemptible flag automatically.
  Right for interruption-tolerant work (CI, batch, dev boxes), wrong for
  anything a 30-second preemption notice breaks.
- `FLEX_START` — discounted capacity with a FLEXIBLE START TIME (Dynamic
  Workload Scheduler): the VM starts when capacity is available, runs at
  most `maxRunDurationSeconds` (required), and is DELETED when reclaimed.
  Right for queued batch/AI jobs that care about throughput, not latency;
  wrong for anything that must start now or must survive its run bound.
- `RESERVATION_BOUND` — consumes ONLY one named reservation (pair with
  `reservationAffinity.type: SPECIFIC_RESERVATION`); the VM's lifecycle
  follows the reservation's capacity. Right for committed-use capacity
  you already paid for.

Spot-vs-FLEX_START in one line: Spot starts now and may die any time;
FLEX_START starts when GCP is ready and dies on schedule.

## Data outlives the VM — attachment vs composition

Durable data belongs on first-class `GcpComputeDisk` resources referenced
through `attachedDisks[].source` — the disk carries its own lifecycle,
snapshots, and deletion policy, and survives this VM's replacement. The
boot disk is the one disk this component creates inline (image/snapshot
sources, size, type, tuning); `autoDelete: false` keeps it after the VM
for forensics or re-attachment, and `replicaZones` (exactly two, one the
instance's own zone) makes it a REGIONAL disk for zone-failure
resilience — but only from a `sourceSnapshot` boot source: GCP cannot
create a regional boot disk from an image (the API rejects it outright),
so the regional-boot pattern is snapshot-a-golden-disk, then boot from
the snapshot. The spec enforces the pairing before anything deploys. `forceAttach` exists for exactly one moment: regional-disk
failover, when the disk is still attached to a failed instance —
force-attaching a zonal disk is an API error, and setting the flag
replaces the VM. Scratch disks are the opposite end of the spectrum:
physically attached local SSDs whose contents vanish on stop or
preemption — caches and temp space only.

One boot-disk trap earns its own warning: `guestOsFeatures` is
all-or-nothing. The API merges what you declare with the image's own
features at create, but every later apply compares your list against the
stored (merged) set and plans a full VM replacement on any difference —
so declaring just the feature you wanted to add silently schedules your
VM for recreation on the next apply. Either leave it unset (the image's
features apply cleanly) or declare the image's complete feature set.

## deletion_protection vs deletion_policy

Two guards, different attack surfaces. `deletionProtection` is a GCP-side
latch on the instance OBJECT: any delete fails until it is flipped off —
it guards against console/API accidents but must be unset before IaC can
destroy. `deletionPolicy` governs what the IaC destroy itself may do:
`PREVENT` fails the destroy outright, `ABANDON` removes the VM from
management but leaves it running (and billing) in GCP. Neither protects
data — that is the disks' job. For a stateful VM the honest layering is:
data on referenced `GcpComputeDisk`s with their own guards, plus
whichever instance-level latch matches who you distrust (console users →
`deletionProtection`; automation → `deletionPolicy: PREVENT`).

## Networking judgment

Every interface needs exactly one attachment point: `network` (auto-mode
VPC), `subnetwork` (custom-mode — the production norm), or a Private
Service Connect `networkAttachment` alone (the consumer side of PSC
interfaces, reaching into a producer's VPC). No `accessConfigs` means no
external IP — the hardened default; pair with Cloud NAT for egress. Pin
addresses by referencing reserved `GcpAddress` resources rather than
literals so they survive VM replacement. Dual-stack needs the whole
chain IPv6-enabled: `stackType: IPV4_IPV6` here AND an IPv6-enabled
subnetwork — static `ipv6Address`/`externalIpv6` pins are only for
workloads where the address is part of the contract (DNS, allowlists).
`vlan` (2–255) declares a dynamic sub-interface; `igmpQuery` enables
multicast — both are niche, and both assume network-side design that
must exist first.

## Encryption posture

CMEK only, at every level: per-disk `kmsKey` (+`kmsKeyServiceAccount`),
`sourceImageEncryption`/`sourceSnapshotEncryption` to decrypt encrypted
boot sources, and `instanceEncryptionKey` for instance-level state.
Customer-supplied raw keys (CSEK) are deliberately NOT modeled — the
provider stores those arguments in plain-text state, and key material
flowing through manifests contradicts the platform's secret posture; the
recorded exclusions live in this component's parity manifest. Before the
first CMEK apply, the Compute Engine service agent needs
`roles/cloudkms.cryptoKeyEncrypterDecrypter` on every referenced key —
missing it fails at apply, not at plan.

## Identity

Never ship the Compute Engine default service account to production:
reference a dedicated least-privilege `GcpServiceAccount` with the single
`cloud-platform` scope and control access entirely through IAM roles.
Scopes are a legacy coarse filter; IAM is the real boundary. SSH access
follows the same logic — prefer OS Login (metadata `enable-oslogin`) over
static `sshKeys`, which OS Login ignores anyway. `workloadIdentityConfig`
goes one step further: Compute Engine issues the VM a SPIFFE identity
(and, with `identityCertificateEnabled`, rotated X.509 certificates) so
services authenticate to each other by identity rather than by shared
secret or network position. Both fields are create-time only — changing
them replaces the VM.

## Host failure is a timing decision

`scheduling.hostErrorTimeoutSeconds` (90..330 in steps of 30) sets how
long Compute Engine waits before declaring the host failed and starting
recovery. Lower recovers a hung host faster at the cost of more false
positives; leave it unset for the default timing.

## On the diagram

The instance is the consuming hub of the compute family: edges point at
the `GcpProject` it lives in, the `GcpVpcNetwork`/`GcpSubnetwork` its
NICs attach to, reserved `GcpAddress` nodes for pinned IPs, referenced
`GcpComputeDisk` nodes for boot-from-disk and every data attachment,
`GcpKmsKey` nodes for each CMEK arm, and the `GcpServiceAccount` it runs
as. Its `self_link`/`internal_ip` outputs feed DNS records, instance
groups, and firewall targeting downstream.

## Pairs well with

- `GcpComputeDisk` — durable boot and data volumes that outlive the VM.
- `GcpAddress` — stable internal/external IPs across VM replacement.
- `GcpVpcNetwork` / `GcpSubnetwork` — the network fabric.
- `GcpServiceAccount` — the least-privilege runtime identity.
- `GcpKmsKey` — CMEK for disks, sources, and instance state.
- `GcpRouterNat` — internet egress for external-IP-less VMs.
- `GcpDnsRecord` — names for the `internal_ip`/`external_ip` outputs.
