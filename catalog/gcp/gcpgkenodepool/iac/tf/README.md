# GcpGkeNodePool - Terraform Module

This Terraform module provisions a Google Kubernetes Engine node pool (`google_container_node_pool`). It is the Terraform-side implementation of the Planton `GcpGkeNodePool` resource kind and has feature parity with the Pulumi module.

## Overview

The module enables the Kubernetes Engine API (`container.googleapis.com`) so a fresh project works first try, then creates the node pool with the full released-provider surface: fixed or autoscaled sizing (per-zone or total bounds, scale-to-zero), surge and blue-green upgrade strategies, compact placement and queued provisioning, per-pool pod ranges, subnetwork and multi-networking attachments, node drain pacing, engine-side deletion policy, and the complete node configuration — machine shape, boot disk (including hyperdisk provisioning), custom node images, local SSDs with encryption modes, storage pools, GPUs with GKE-managed drivers and GPUDirect, Spot/preemptible capacity, shielded and confidential VMs, confidential storage, CMEK boot-disk encryption, reservation affinity, sole tenancy, gVisor sandboxing, Windows pools, secondary boot disks, containerd registry configuration (private CAs, mirrors, writable cgroups), full kubelet tuning (CPU/memory/topology managers, eviction thresholds, crash-loop backoff, image GC), and Linux OS tuning (sysctls, cgroups, hugepages, transparent hugepages, kernel module policy, PTP time sync, swap).

The pool addresses its parent by the cluster's `name` and `location` outputs (resolved to plain strings before the module runs) — no data-source lookup. An empty `project_id` falls back to the provider's default project; empty optional strings become null so the provider omits them from the API payload. The module runs on the plain `google` provider — every modeled field is GA on the pinned 8.x line, and the kind's `iac/provider-parity.yaml` records the full argument accounting.

For autoscaled pools, `lifecycle.ignore_changes` on `node_count` keeps plans from fighting the cluster autoscaler. GKE's `disable-legacy-endpoints=true` node metadata is enforced beneath any user metadata; platform attribution labels are merged over user resource labels.

## Usage with Planton CLI

```shell
planton tofu init --manifest ../../e2e/manifest.yaml
planton tofu plan --manifest ../../e2e/manifest.yaml
planton tofu apply --manifest ../../e2e/manifest.yaml --auto-approve
planton tofu destroy --manifest ../../e2e/manifest.yaml --auto-approve
```

Credentials are provided via stack input (by the CLI), not in the manifest `spec`. Manifest file: `../../e2e/manifest.yaml`.

## Direct Terraform Usage

```bash
cd catalog/gcp/gcpgkenodepool/iac/tf
terraform init
terraform plan -var-file=terraform.tfvars.json
terraform apply -var-file=terraform.tfvars.json
```

## Variables

| Name | Description | Default |
|------|-------------|---------|
| `metadata` | Resource metadata (name, labels, etc.) | — |
| `spec` | GcpGkeNodePool spec | — |

The `spec` object includes: `cluster_name` + `location` (required; resolved from the cluster's outputs), `node_pool_name` XOR `name_prefix` (empty = metadata.name), `node_locations`, `version`, `max_pods_per_node`, `initial_node_count`, `node_count` XOR `autoscaling` (per-zone or total bounds + location policy), `management` (auto-repair/auto-upgrade, both default true), `upgrade_settings` (surge dials or blue-green rollout), `placement_policy`, `queued_provisioning_enabled`, `node_drain_config`, `deletion_policy`, `ignore_node_count_changes`, `network_config` (pod range, private nodes, subnetwork, accelerator network profile, additional node/pod networks, egress tier, overprovision), `node_config` (machine/disks/boot-disk block/node image/identity/scopes/labels/taints/spot/GPUs incl. GPUDirect/shielded/confidential/CMEK/local SSDs incl. encryption mode/storage pools/sole tenancy/sandbox/Windows/GCFS/gVNIC/fast-socket/workload-metadata/reservations/secondary disks/containerd config/kubelet incl. managers, eviction, crash-loop backoff/Linux incl. hugepages, swap, kernel-module policy/host maintenance/architecture taints/logging-variant/flex-start/max-run-duration/resource-manager tags/confidential storage), and `project_id` (empty falls back to the provider default project).

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
