# GCP GKE Node Pool

Deploys a Google Kubernetes Engine node pool via `google_container_node_pool` (Terraform) or Pulumi `container.NodePool` — a group of identically configured Compute Engine VMs attached to a GKE cluster, with the full node configuration surface: machine shape, disks, identity, scheduling constraints, GPU accelerators, security posture, and node-level tuning.

## Overview

Node pools are the compute half of the GKE composition boundary: the `GcpGkeCluster` owns the control plane and cluster-wide configuration; every node pool is a first-class resource with its own lifecycle, referencing the cluster by its `name` and `location` outputs. Production clusters run several pools — general-purpose on-demand, scale-to-zero Spot for fault-tolerant batch, GPU pools for ML — each sized and tainted for its workloads.

The pool inherits its project and location from the parent cluster (both resolve by reference), and honors the ambient-project contract: an empty `project_id` falls back to the provider's default project.

## Purpose

- **Heterogeneous compute, honest boundaries**: match infrastructure to workloads — pools are created, resized, and destroyed without ever touching the cluster control plane.
- **Pre-deploy coherence**: the API's cross-field rules (per-zone XOR total autoscaling limits, surge XOR blue-green settings, spot XOR preemptible, fast-socket-needs-gVNIC, specific-reservation key pairing) are enforced by validation before any cloud call.
- **Cost levers as first-class fields**: Spot VMs, scale-to-zero autoscaling, flex-start, max run duration, and reservation consumption are all modeled — not hidden behind escape hatches.

## Key Features

- **Sizing**: fixed count or cluster-autoscaler management with per-zone (`min_nodes`/`max_nodes`) or total (`total_min_nodes`/`total_max_nodes`) bounds and the `BALANCED`/`ANY` location policy
- **Upgrades & lifecycle**: surge (max_surge/max_unavailable) or blue-green strategy with batch rollout pacing and soak windows; drain pacing on pool deletion (`node_drain_config` with PDB respect — requires per-project enablement from GCP support; the API rejects it on projects without the allowlist); engine-side `deletion_policy` (DELETE/PREVENT/ABANDON); GKE-generated names via `name_prefix` for swap-without-collision workflows
- **Machine shape**: machine type, boot disk size/type — including the first-class `boot_disk` block with hyperdisk provisioned IOPS/throughput — custom node images, image type, min CPU platform, local SSDs (SCSI count, NVMe ephemeral-storage backing with data cache, raw-block NVMe, encryption modes), hyperdisk storage pools, SMT/nested-virtualization/PMU controls
- **GPUs**: accelerator type/count, GKE-managed driver installation, MIG partitioning, time-sharing/MPS GPU sharing, NCCL Fast Socket + gVNIC, GPUDirect strategies, RDMA accelerator network profiles
- **Scheduling & tenancy**: Kubernetes node labels, taints with all three effects, ARM architecture-taint behavior, network tags, Resource Manager tags, compact placement (+ TPU topology), queued provisioning (Dynamic Workload Scheduler), flex-start, max run duration, sole-tenant node groups with affinity rules, host maintenance cadence
- **Identity & security**: node service account by reference, OAuth scopes, shielded VM options, confidential nodes (SEV/SEV-SNP/TDX) and confidential storage, CMEK boot-disk encryption by `GcpKmsKey` reference, workload metadata mode, gVisor sandboxing, Windows Server pools
- **Networking**: dedicated per-pool pod ranges (create-new or use-existing), a pool-specific subnetwork, multi-networking (additional node interfaces and pod networks), per-pool private nodes override, TIER_1 egress bandwidth, pod CIDR overprovision control
- **Registry access**: containerd configuration — private registries behind custom CAs (Secret Manager-held certificates), per-registry mirrors with capabilities, timeouts, client TLS and custom headers, writable cgroups
- **Node tuning**: kubelet (CPU/memory/topology managers, CFS quota, PID limits, log rotation, image GC by threshold AND age, parallel image pulls, soft-eviction thresholds with grace periods and minimum reclaim (reclaim values are percentage-only — GKE rejects absolute quantities), crash-loop backoff caps, single-process OOM kill, unsafe sysctl allowlists, the insecure read-only port), Linux (sysctls, cgroup mode, hugepages, transparent hugepage modes, signed-kernel-module enforcement, PTP/KVM time sync, swap with sizing profiles and encryption), logging variant, image streaming (GCFS)
- **Capacity**: Spot and legacy preemptible VMs, Compute Engine reservation affinity (incl. reserve-or-fail), secondary boot disks for image preloading

## Stack Outputs

| Output | Description |
|---|---|
| `node_pool_name` | Name of the pool as created in GKE |
| `instance_group_urls` | Managed instance group URLs backing the pool (one per zone for regional clusters) |
| `min_nodes` | Effective minimum size (autoscaling minimum, or the fixed count) |
| `max_nodes` | Effective maximum size (autoscaling maximum, or the fixed count) |
| `current_node_count` | Nodes per zone at the last deploy (a snapshot — the autoscaler moves it) |
| `node_pool_id` | Fully qualified resource ID: `projects/{project}/locations/{location}/clusters/{cluster}/nodePools/{name}` |
| `location` | The pool's location (the parent cluster's region or zone), exactly as provided in the spec |
| `version` | Kubernetes version running on the pool's nodes |

## Deliberately not modeled (recorded reasons)

Every provider argument of `google_container_node_pool` is accounted for —
matched, mapped, or excluded with the reason recorded in
`iac/provider-parity.yaml`. The exclusions, in plain terms:

| Excluded Feature | Why |
|---|---|
| `maintenance_policy.exclusion_until_end_of_support.start_time` / `end_time` | Computed-only in the provider: GKE derives the exclusion window from the pool's version once `excludeUpgradesUntilEndOfSupport` is set; there is no input to author. |

## Related Components

- **GcpGkeCluster** — the control plane this pool attaches to (references its `name` and `location` outputs)
- **GcpServiceAccount** — the node identity (`service_account` reference)
- **GcpKmsKey** — CMEK key for boot-disk encryption
- **GcpGkeWorkloadIdentityBinding** — grants workloads IAM identities; pair with `GKE_METADATA` workload metadata mode
- **GcpDataprocCluster** — a Dataproc-on-GKE virtual cluster schedules Dataproc roles onto named node pools by `node_pool_id`

## Additional Resources

- [Node Pools Documentation](https://cloud.google.com/kubernetes-engine/docs/concepts/node-pools)
- [NodePools REST API](https://cloud.google.com/kubernetes-engine/docs/reference/rest/v1/projects.locations.clusters.nodePools)

---

© Planton. Licensed under [Apache-2.0](https://github.com/plantonhq/planton/blob/main/LICENSE).
