# GcpGkeNodePool - Pulumi Module

This Pulumi (Go) module provisions a Google Kubernetes Engine node pool (`container.NodePool`). It is the Pulumi-side implementation of the Planton `GcpGkeNodePool` resource kind and has feature parity with the Terraform module.

## Overview

The module enables the Kubernetes Engine API (`container.googleapis.com`) so a fresh project works first try, then creates the node pool with the full released-provider surface: fixed or autoscaled sizing (per-zone or total bounds, scale-to-zero), surge and blue-green upgrade strategies, compact placement and queued provisioning, per-pool pod ranges, subnetwork and multi-networking attachments, node drain pacing, engine-side deletion policy, and the complete node configuration — machine shape, boot disk (including hyperdisk provisioning), custom node images, local SSDs with encryption modes, storage pools, GPUs with GKE-managed drivers and GPUDirect, Spot/preemptible capacity, shielded and confidential VMs, confidential storage, CMEK boot-disk encryption, reservation affinity, sole tenancy, gVisor sandboxing, Windows pools, secondary boot disks, containerd registry configuration, full kubelet tuning (CPU/memory/topology managers, eviction thresholds, crash-loop backoff, image GC), and Linux OS tuning (sysctls, cgroups, hugepages, transparent hugepages, kernel module policy, PTP time sync, swap). The kind's `iac/provider-parity.yaml` records the full argument accounting against the pinned provider.

The pool addresses its parent by the cluster's `name` and `location` outputs (resolved to plain strings before the module runs) — no lookup. An empty `project_id` falls back to the provider's default project — the ambient-project contract every GCP kind honors. For autoscaled pools, `IgnoreChanges` on `nodeCount` keeps previews from fighting the cluster autoscaler. GKE's `disable-legacy-endpoints=true` node metadata is enforced beneath any user metadata; platform attribution labels are merged over user resource labels.

## Usage with Planton CLI

```shell
planton pulumi up --manifest ../../e2e/manifest.yaml --module-dir .
planton pulumi destroy --manifest ../../e2e/manifest.yaml --module-dir .
```

Credentials are provided via IaC input (by the CLI), not in the manifest `spec`. Manifest file: `../../e2e/manifest.yaml`.

## Build

```bash
cd catalog/gcp/gcpgkenodepool/iac/pulumi
make build
```

## Outputs

| Name | Description |
|------|-------------|
| `node_pool_name` | Pool name as created in GKE |
| `instance_group_urls` | Managed instance group URLs backing the pool (one per zone) |
| `min_nodes` | Effective minimum size (autoscaling minimum, or the fixed count) |
| `max_nodes` | Effective maximum size (autoscaling maximum, or the fixed count) |
| `current_node_count` | Nodes per zone at the last deploy (autoscaler-owned at runtime) |
| `node_pool_id` | `projects/{project}/locations/{location}/clusters/{cluster}/nodePools/{name}` |
| `location` | The pool's region or zone, exactly as provided in the spec |
| `version` | Kubernetes version running on the pool's nodes |
