# GcpVertexAiNotebook Guide

The judgment this guide protects: a Workbench instance is a stateful
workstation, not disposable compute. Its disks hold the notebooks and
data your scientists are actively working on, most of its shape is
frozen at create time, and the cheapest lever you have — stop/start —
is also the one teams forget to use.

## Pair the GPU with the machine series, not just the zone

The accelerator list runs from the Tesla family through the current
generation, and each tier binds to a machine series: `NVIDIA_L4` rides
G2, `NVIDIA_TESLA_A100`/`NVIDIA_A100_80GB` ride A2, and the
`NVIDIA_H100_80GB`/`NVIDIA_H100_MEGA_80GB`/`NVIDIA_H200_141GB`/
`NVIDIA_B200` generation requires its matching A3/A4 series. A
mismatched pair fails at create. The `_VWS` variants are
virtual-workstation (graphics-licensed) builds of the same silicon —
for visualization desktops, not training. For scarce GPUs, pair the
notebook with `reservationAffinity` so it draws from pre-purchased
capacity instead of gambling on on-demand availability.

## Disks: hyperdisk needs the newer series too

Boot and data disks accept Persistent Disk types everywhere and
hyperdisk types (`HYPERDISK_BALANCED`, `HYPERDISK_ML`, plus
`HYPERDISK_EXTREME`/`HYPERDISK_THROUGHPUT` on the data disk) on the
machine series that support them. `HYPERDISK_ML` is purpose-built for
feeding training jobs from the data disk. CMEK is derived, never set:
presence of a `kmsKey` reference means CMEK, absence means
Google-managed — there is no third state to configure.

## STOPPED is the cost lever; use it on a schedule

`desiredState: STOPPED` suspends compute billing while keeping both
disks — for a GPU notebook, that is most of the bill. Stop instances
outside working hours and let the ACTIVE flip boot them in minutes.
Storage keeps billing while stopped; that is the point — the work
survives.

## Identity: decide who the notebook IS before anyone uses it

By default notebook code acts as the VM's service account — fine for a
single-owner instance, wrong for shared ones (everyone inherits the
same permissions, audit logs say the robot did it). `enableManagedEuc`
makes JupyterLab act as the signed-in user's own identity (per-user
IAM, per-user audit); `enableThirdPartyIdentity` extends access to
workforce-federation users. Both are mutable, but flipping identity
posture on a team mid-project invalidates every assumption their
pipelines made — decide first.

## Destroy takes the disks with it

Empty/`DELETE` deletes the instance INCLUDING boot and data disks —
unsynced notebook work is gone. `deletionPolicy: PREVENT` makes destroy
fail and is the right posture for any notebook whose data disk is not
continuously synced to git/GCS. `ABANDON` keeps the VM running (and
billing) outside management. `enableDeletionProtection` adds the
API-side guard: while true, Workbench refuses to delete the instance from
any client (console, gcloud, either engine) until the flag is lifted.
Pair it with `deletionPolicy: PREVENT` for a workstation whose disks hold
work not yet pushed anywhere else — one guard on the API, one on the
engine. `dataDisk.resourcePolicies` attaches Compute Engine snapshot
schedules to the data disk, the agentless way to back up that work on a
cadence, and `minCpuPlatform` pins the VM to a CPU generation when
notebook code depends on newer instruction sets.

## What is deliberately absent

`instance_id` (a second user-settable identity handle beside the name)
is not modeled: instances are named from `instanceName`/
`metadata.name`, and a duplicate ID lever would only let the two drift.
Post-provisioning content — conda environments, git checkouts, data
loads — belongs to `metadata` startup scripts, not the spec.
