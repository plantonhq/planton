# GcpGkeNodePool Guide

The judgment this guide protects: node pools are cattle in a way clusters
never are — nearly all of `nodeConfig` is immutable, and GKE replaces the
NODES on any change to it. The working posture is therefore never "edit
the pool" but "add the right pool next to it, drain, delete" — and the
spec's levers (name prefixes, drain pacing, taints, scale-to-zero) exist
to make that rotation boring.

## One pool per workload class, fenced

The production shape is several purpose-built pools — general on-demand,
scale-to-zero Spot for batch, GPU for ML — each with a taint and matching
labels, so nothing lands where it shouldn't. A pool without a taint is a
default pool by another name: everything schedules there, and its
replacement disrupts everyone. GKE taints GPU and Arm pools automatically
(`architectureTaintBehavior` controls the Arm one); everything else is
your fence to build.

## Rotation, not mutation

Changing machine type, image, disks, or almost anything in `nodeConfig`
replaces the pool's nodes. For pools serving live traffic, prefer
explicit rotation: create the successor pool (a `namePrefix` pool gets a
fresh GKE-generated name each time, so blue and green never collide),
let scheduling drain the old one, then delete it — with
`nodeDrainConfig` pacing the teardown (grace period, PDB respect with a
timeout, so one stuck budget cannot hang the delete forever). One gate to
know before you reach for it: customized node drain is allowlisted per
project — GCP support has to enable it, and until they do the API rejects
any pool carrying the block with "customized node drain timeout is not
enabled for this project". Blue-green `upgradeSettings` gives you the
same shape for VERSION changes within one pool; rotation covers
everything else.

## Spot is a discipline, not a discount

Spot nodes vanish with 30 seconds' notice. The pattern that works is a
package: `spot: true` + scale-to-zero autoscaling + `locationPolicy:
ANY` (drains fewer reclaimed zones) + a `capacity=spot` taint so only
tolerating workloads land. Skipping any leg of that is how Spot outages
become production outages. `maxRunDuration` and flex-start extend the
same economics to hard-to-get GPU capacity; `reservationAffinity` with
ANY_RESERVATION_THEN_FAIL is the opposite stance — consume committed
capacity or fail loudly rather than fall back to on-demand billing.
Give reclaimed nodes a graceful exit:
`kubeletConfig.shutdownGracePeriodSeconds` (10..10000, Spot and
preemptible pools only) is the total time a node delays shutdown so pods
terminate cleanly, and `shutdownGracePeriodCriticalPodsSeconds` reserves
part of it for system-critical pods after ordinary pods have had their
share.

## Holding a version, and preparing the host

`excludeUpgradesUntilEndOfSupport: true` pins the pool on its current
Kubernetes version until that version's end-of-support date — the lever
for a workload that must be re-qualified before it moves minors (GKE
reports the resulting window in the pool's status; the cluster's own
maintenance windows still govern everything else).
`linuxNodeConfig.customNodeInit` runs a boot-time script from Cloud
Storage (pin it with `gcsGeneration` so a later upload does not silently
change what new nodes run) or Secret Manager (for scripts that embed
credentials) before the node joins the cluster — host-level setup no
DaemonSet can do.

## Kubelet tuning: set only what you can defend

Every kubelet field unset means GKE's default — which is right for
almost everyone. The tuning surface exists for pools with a measured
problem: static CPU manager and topology manager for latency-critical
pinning (Guaranteed-QoS pods only), eviction thresholds when node-pressure
OOMs precede kubelet reaction, crash-loop backoff caps for dev clusters
where fast restart matters more than churn, swap for memory-elastic
workloads that prefer latency to OOM kills. Each of these shifts failure
modes rather than removing them; tune from evidence, and leave the rest
of the pool's fields at GKE defaults so re-plans stay clean.

## Private registries live on the pool

`containerdConfig` is where private-CA registries and mirrors are wired
— per pool, not per cluster (the cluster's `nodePoolDefaults` seeds new
pools, but existing pools keep their own). The CA and client TLS
material comes from Secret Manager by URI: the nodes fetch it, the
manifest never carries it. When a pull fails against an internal
registry, check the pool's containerd config before the registry — the
per-pool scope means one pool can be configured and its neighbor not.

## Destroy stance

`deletionPolicy: PREVENT` suits pools that are the only compute under
stateful workloads — a destroy fails at the pool instead of evicting
everything pod-hosted. ABANDON leaves the nodes running unmanaged
(break-glass only). DELETE is right for the rotation pattern above and
for e2e fixtures. Remember the drain settings apply on the way out:
`respectPdbDuringNodePoolDeletion` makes even the delete honor
application availability budgets, bounded by `pdbTimeoutDuration`.

## On the diagram

The pool hangs off its `GcpGkeCluster` by the cluster's `name` and
`location` outputs — the edge every pool shares. `GcpServiceAccount`
(node identity) and `GcpKmsKey` (boot-disk CMEK) are its other typed
edges; multi-networking adds visible edges to extra `GcpVpcNetwork` /
`GcpSubnetwork` nodes. A cluster with one giant untainted pool reads as
exactly the risk it is.

## Pairs well with

- `GcpGkeCluster` — the control plane; the pool inherits project and
  location through its outputs.
- `GcpServiceAccount` — a minimal dedicated node SA beats the Compute
  Engine default everywhere it matters.
- `GcpKmsKey` — CMEK for boot disks (and local-SSD ephemeral-key
  encryption for data that must die with the node).
- `GcpGkeWorkloadIdentityBinding` — workload identity for the pods this
  pool runs; pair with `workloadMetadataMode: GKE_METADATA`.
