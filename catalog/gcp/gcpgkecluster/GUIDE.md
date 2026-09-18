# GcpGkeCluster Guide

The judgment this guide protects: almost every decision that matters on a
GKE cluster is IMMUTABLE — mode, location, network, IP allocation, the
private topology. A cluster is a 20-minute create and a
week-of-migrations replace, so the cost of choosing wrong is not a diff,
it is an evacuation. Decide the immutables deliberately; everything else
can wait.

## Autopilot vs Standard is the first and last decision

Autopilot removes node management entirely (and bills per pod);
Standard gives you `GcpGkeNodePool` resources with full machine control.
Converting means recreating the cluster and re-homing every workload.
Choose Autopilot when the team should not own node sizing at all; choose
Standard the moment any workload needs a specific machine family, GPUs
with driver control, sole tenancy, Windows, or kubelet/OS tuning — none
of that surface exists on Autopilot, and the spec's validation rules
will tell you so at manifest time rather than at apply.

## Size the pod range like you mean it

The pod secondary range is the classic GKE regret: it is immutable, and
every node reserves a /24 slice of it by default (~2x its max pods), so
a comfortable-looking /16 supports only ~256 nodes. Plan the range from
target node count, not current; `defaultMaxPodsPerNode` stretches it
(halving max pods halves the per-node slice); and when it still runs
out, `additionalPodRangeNames` / `additionalIpRanges` are the
post-creation escape hatches — mutable, unlike the primary range.
`podCidrOverprovisionDisabled` doubles node density at the cost of
headroom for pod churn; it is a density lever, not a free lunch.

## Private topology: pick PSC unless you have a reason not to

`privateEndpointSubnetwork` (PSC-based) needs no /28 carve-out and no
peering bookkeeping; `masterIpv4CidrBlock` (peering-based) exists for
estates that already standardized on it. Either way, private nodes
cannot pull from registries outside Google without a `GcpRouterNat` on
the network — the first symptom is ImagePullBackOff on everything
non-GCR, minutes after an otherwise green create. The DNS endpoint
(`controlPlaneEndpoints`) is the modern third path: IAM-authenticated
kubectl without bastions, peering, or public IPs — and
`enableK8sTokensViaDns` extends it to workload token auth.

## Upgrades: channels decide, windows schedule, budgets pace

A release channel (REGULAR by default) owns the version; maintenance
windows say WHEN GKE may act; exclusions freeze changes (scope them —
NO_MINOR_UPGRADES still allows patches; `endTimeBehavior:
UNTIL_END_OF_SUPPORT` rides a minor to its end of support); and the
disruption budget spaces disruptive events regardless of windows. Pin
`minMasterVersion` only on channel NONE — on a channel, the channel
wins and the pin just fights it. `gkeAutoUpgradePatchMode: ACCELERATED`
is for fleets that want CVE patches the day they ship.

Three window forms, one choice: `dailyWindow` (same four hours every
day), `recurringWindow` (an RFC 5545 rule anchored on an absolute first
start/end timestamp), or `recurringTimeWindow` (the same rule as a time
of day plus a duration, optionally delayed until a calendar date) — pick
the last when the policy reads "every Saturday at 02:00 for six hours,
starting next quarter". `rollbackSafeUpgrade.controlPlaneSoakDuration`
(6 hours to 7 days) turns control-plane minor upgrades into two steps:
the control plane upgrades but keeps emulating the old minor during the
soak, so the upgrade can be rolled back; setting
`desiredEmulatedVersion` to the target minor completes it declaratively.

## NAP is a cost brake wearing an autoscaler's clothes

Node auto-provisioning creates pools you never wrote. The spec refuses
NAP without resource limits because an unbounded NAP is an unbounded
bill — but the limits are the ceiling on GKE's judgment, not a target.
Set `autoProvisioningDefaults` as deliberately as a real pool (dedicated
SA, CMEK, shielded posture, upgrade rollout): NAP-created pools inherit
those defaults, and they are YOUR pools the moment finance asks.

## Secrets: two add-ons, two delivery models

`enableSecretManagerCsi` + `secretManagerRotation` mounts secrets as
files, re-fetched on the rotation cadence — no Kubernetes Secret objects
exist to leak. `secretSync` materializes them AS Kubernetes Secrets for
workloads that demand env-var or secretKeyRef semantics. Choose CSI when
you can (smaller audit surface); sync when the workload API forces it.

## Destroy stance

`deletionProtection` (default true) is the GKE-native guard — the API
itself refuses the delete. `deletionPolicy` layers under it: PREVENT
fails any destroying plan before it starts; ABANDON drops the cluster
from state and leaves it running unmanaged — the break-glass for state
surgery, not an operating mode. A cluster deletion takes every workload,
PV, and LoadBalancer with it; keep both guards on for anything real, and
turn them off in the same change that means it.

## On the diagram

The cluster is the hub of the GKE family: it consumes `GcpVpcNetwork`
and `GcpSubnetwork` (plus `GcpKmsKey` for etcd/NAP CMEK,
`GcpServiceAccount` for NAP identity, `GcpPubSubTopic` for lifecycle
events, `GcpBigQueryDataset` for usage export), and its `name` /
`location` outputs are what every `GcpGkeNodePool` and the workload
identity bindings attach to. Standard clusters without pools are visible
scaffolding — the diagram shows a control plane waiting for compute.

## Pairs well with

- `GcpGkeNodePool` — the compute; several per cluster is the norm.
- `GcpGkeWorkloadIdentityBinding` — KSA-to-GSA identity on the
  cluster's workload pool.
- `GcpRouterNat` — non-negotiable for private nodes that pull public
  images.
- `GcpKmsKey` — etcd CMEK, NAP boot disks, and `userManagedKeys` for
  regulated estates.
- `GcpSubnetwork` — plan the secondary ranges there first; the cluster
  only names them.
