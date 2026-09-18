# GcpDataprocCluster Guide

The judgment this guide protects: a Dataproc cluster is
create-mostly-immutable compute wrapped around jobs that are not. The
in-place levers are exactly four — labels, worker counts, the
autoscaling-policy attachment, and the lifecycle TTLs — everything else
is a replace. So the operating question is never "how do I mutate this
cluster" but "should this cluster exist between jobs at all".

## Ephemeral-first is the cost model

The TTLs are the difference between a data platform and a standing bill.
`idleDeleteTtl` deletes an idle cluster; the stop pair —
`idleStopTtl` / `autoStopTime` — shuts the VMs down but keeps the
cluster restartable, the middle ground for clusters whose
configuration is expensive to recreate but whose compute should not idle
overnight. Persistent state belongs OUTSIDE the cluster: a Metastore
service for schemas, GCS for data, `stagingBucket`/`tempBucket` you own
rather than the auto-created ones (which Dataproc never deletes and
which accumulate per-creator ACLs).

## Size the groups by role, not symmetry

Masters are the control plane: 1 or 3 (HA), never scaled. Primary
workers carry HDFS — scale them deliberately (`numInstances` /
`minNumInstances` update in place, with `gracefulDecommissionTimeout`
draining YARN on shrink). Secondary workers are stateless burst
capacity: spot-priced, aggressively autoscaled, first to be preempted.
The flexibility policy is the capacity-crunch answer — ranked
machine-type fallbacks on ALL three groups, with the standard/spot
`provisioningModelMix` on secondaries only (masters and primary workers
are always on-demand; the spec's validation enforces the boundary).
A flexibility policy REPLACES `machineType` — never set both on one
group. The API accepts the pairing but silently drops the machine type
from the stored cluster, and because that field is create-only, every
subsequent apply then plans a whole-cluster replacement. The spec
rejects the pairing outright; put every acceptable type in the ranked
`instanceSelectionList` instead (the first live run of this arm watched
Dataproc pick the list's second type when the first was capacity-tight —
the fallback working exactly as designed).

Disks follow the same role-first thinking. `diskConfig.attachedDisks`
adds persistent disks to every node of a role beyond the boot disk —
for HDFS data or shuffle spill that should not share the boot volume,
or for Hyperdisk tiers the boot disk cannot use — and is fixed at
creation. Inside a flexibility policy each selection entry can carry its
own `diskConfig`, so the fallback machine types get the boot disk and
local-SSD count that fit them rather than inheriting the preferred
type's shape. `gceConfig.confidentialInstanceConfig.confidentialInstanceType`
(SEV, SEV_SNP, TDX) names the Confidential Compute technology for every
node; the older `enableConfidentialCompute` boolean stays for existing
manifests.

## Autoscaling is a policy attachment, not a property

The autoscaler lives in a first-class `GcpDataprocAutoscalingPolicy`
resource — one region-wide policy can govern many clusters, and
ATTACHING/detaching (`autoscalingPolicyUri`) updates in place. Tune
scaling in the policy, not the cluster — and remember the policy cannot
be deleted while any cluster still references it, so teardown order is
clusters first.

## Networking earns two live-proven prerequisites

On a custom-mode VPC there are no default firewall rules, and a
multi-node cluster will hang (not fail) at create because master and
workers cannot talk — an intra-cluster allow rule is part of the
cluster's minimum composition. Hardened projects grant the default
compute identity nothing: give the cluster a dedicated
`GcpServiceAccount` holding `roles/dataproc.worker`, or the control
plane rejects the create outright. `internalIpOnly` clusters
additionally need NAT or Private Google Access to reach anything public.

## The virtual arm is a different animal

`virtualClusterConfig` runs Spark as pods on an existing GKE cluster —
no VMs, no disk configs, no Kerberos, and NO update paths at all (any
change replaces the virtual cluster; the GKE substrate is untouched).
User labels are rejected by the API on this arm. Choose it when the
estate is already Kubernetes-operated and the GKE cluster is the
capacity plane; choose the GCE arm when jobs need OS-level control,
local SSDs, or GPU shapes.

## Destroy stance

`deletionPolicy: PREVENT` fails destroying plans; `ABANDON` releases the
cluster from IaC management while it keeps running — with an ephemeral
cluster this is rarely what you want, but it is the honest lever for
handing a long-lived shared cluster to out-of-band operation. The
auto-created staging/temp buckets survive every destroy by design; sweep
them when an identity change locks new clusters out of the old buckets.

## On the diagram

The cluster consumes `GcpProject`, `GcpVpcNetwork`/`GcpSubnetwork`,
`GcpServiceAccount` (VM identity), `GcpGcsBucket` (staging/temp),
`GcpKmsKey` (CMEK), `GcpDataprocAutoscalingPolicy` (attached policy),
and — on the virtual arm — a `GcpGkeCluster` and its node pools. Its
`cluster_id` output feeds Spark History Server composition on other
clusters.

## Pairs well with

- `GcpDataprocAutoscalingPolicy` — scaling behavior, tuned once, shared
  region-wide.
- `GcpFirewallRule` — the intra-cluster allow rule custom VPCs require.
- `GcpServiceAccount` — the `roles/dataproc.worker` identity hardened
  projects must provide.
- `GcpGcsBucket` — owned staging/temp buckets instead of the immortal
  auto-created ones.
- `GcpGkeCluster` + `GcpGkeNodePool` — the substrate of the virtual arm.
