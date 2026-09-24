# GCP GKE Cluster

Deploys a Google Kubernetes Engine cluster via `google_container_cluster` (Terraform) or Pulumi `container.Cluster` — the managed Kubernetes control plane plus every cluster-wide configuration surface: networking mode, private topology, upgrades and maintenance, autoscaling, security, observability, addons, and Autopilot.

## Overview

One GcpGkeCluster resource is one cluster. Node pools are separate composable resources — `GcpGkeNodePool` — referencing this cluster by its `name` output; the cluster's default node pool is always removed at create time so every pool is an explicitly managed, first-class node. In Autopilot mode GKE manages nodes entirely (billed per pod) and no node pool resources attach.

Clusters are always VPC-native (alias IP): pods and services draw from secondary ranges on the referenced subnetwork — either ranges you name in `ip_allocation`, or ranges GKE creates and manages when none are named. Legacy routes-based networking is deliberately not modeled.

## Purpose

- **Full control-plane surface, honest boundaries**: the cluster owns cluster-wide concerns; compute lives in `GcpGkeNodePool`; NAT, addresses, and subnets are their own composable nodes.
- **Pre-deploy coherence**: the GKE API's cross-field rules (Autopilot conflicts, window exclusivity, NAP limits, CMEK key pairing, private-endpoint requirements) are enforced by validation before any cloud call — a 20-minute cluster create never fails on a rule the manifest already broke.
- **Safety defaults that match GCP**: `deletion_protection` defaults to true (a destroy fails until it is explicitly turned off), shielded nodes follow GCP's default-on posture, and Workload Identity is on by default.

## Key Features

- **Two modes, one kind**: Standard (you own node pools) and Autopilot (`enable_autopilot`; GKE owns nodes) with the API's conflict set enforced pre-deploy — plus Autopilot's own controls: conversion/posture policies (`autopilot_policy`), privileged-workload allowlists, and managed-node settings (`node_pool_auto_config`: network/resource-manager tags, cgroups, kubelet hardening)
- **Private clusters**: private nodes, peering-based (`master_ipv4_cidr_block`) or PSC-based (`private_endpoint_subnetwork`) control planes, master global access, private-only endpoints
- **Control-plane access**: master authorized networks (CIDR allowlist, GCP-public-CIDR and private-endpoint enforcement toggles), the modern DNS endpoint (`control_plane_endpoints`, incl. serving tokens and certs via DNS), legacy client-certificate stance, and node registration mode (`node_creation_mode`)
- **Upgrades on rails**: release channels incl. EXTENDED, `min_master_version`, ACCELERATED patch mode, daily/recurring maintenance windows, up to 20 maintenance exclusions with scopes and end-of-support behavior, disruption budgets between disruptive events
- **Node auto-provisioning (NAP)**: resource limits (the cost brake, required when enabled), autoscaling profiles, default compute classes, full auto-provisioning defaults (SA ref, disks, image, shielding, auto-repair/upgrade, surge/blue-green upgrade rollout)
- **IP allocation**: named or GKE-created ranges, dual-stack, additional pod ranges AND additional subnetwork ranges (with draining), auto-IPAM, network tier
- **Dataplane V2**: `datapath_provider`, FQDN / Cilium cluster-wide network policy, dataplane observability metrics and relay, in-transit pod traffic encryption
- **Security**: CMEK etcd encryption (KMS key ref, incl. all-objects encryption), customer-managed control-plane CAs and signing keys (`user_managed_keys`), Binary Authorization, Security Posture dashboard, RBAC binding lockdown, authenticator groups, confidential nodes (SEV/SEV-SNP/TDX), anonymous-auth hardening, mesh certificates
- **Secrets**: Secret Manager CSI add-on with rotation cadence, and the Secret Manager sync add-on (secrets into Kubernetes Secrets) with its own rotation
- **Observability**: per-component logging/monitoring (incl. KCP components), managed Prometheus (+ auto-monitoring scope), Pub/Sub lifecycle notifications, cost allocation, BigQuery usage export
- **Addons**: HTTP LB, HPA, PD/Filestore/GCS-Fuse/Parallelstore/Lustre CSI drivers, Backup for GKE, NodeLocal DNSCache, Config Connector, Stateful HA, Ray operator (+ logging/monitoring), Cloud Run, pod snapshots, agent sandbox, slice controller, Slurm operator
- **Node-pool defaults**: creation-time defaults for every pool (image streaming, kubelet read-only port, logging variant, containerd private-registry access)
- **Fleet registration**: `fleet_project` + membership type for multi-cluster features
- **Lifecycle & scale**: engine-side `deletion_policy` (DELETE/PREVENT/ABANDON) under `deletion_protection`, alpha clusters and beta API groups for evaluation, and read-side performance switches for very large clusters (`ignore_node_count_changes`, `skip_node_pool_refresh`)

## Stack Outputs

| Output | Description |
|---|---|
| `endpoint` | Kubernetes API server IP (the private endpoint on private-only control planes) |
| `cluster_ca_certificate` | Base64 CA certificate clients use to validate the API server's TLS certificate (public trust material) |
| `workload_identity_pool` | `PROJECT_ID.svc.id.goog` — empty when Workload Identity is disabled on a Standard cluster |
| `cluster_id` | Fully qualified resource ID: `projects/{project}/locations/{location}/clusters/{name}` |
| `name` | The cluster name as created in GCP — the handle node pools and gcloud use |
| `location` | Region (regional) or zone (zonal), exactly as provided in the spec |
| `self_link` | Server-defined URL of the cluster resource |
| `master_version` | Kubernetes version currently running on the control plane |

## Deliberately not modeled (recorded reasons)

Every provider argument of `google_container_cluster` is accounted for —
matched, mapped, or excluded with the reason recorded in
`iac/provider-parity.yaml`. The exclusions, in plain terms:

| Excluded Feature | Why |
|---|---|
| Inline `node_pool` blocks / default-pool `node_config` | Node pools are first-class `GcpGkeNodePool` resources; the default pool is always removed at create time. Inline pools mixed with external pool resources are the provider's own documented anti-pattern. |
| `networking_mode` (ROUTES), `cluster_ipv4_cidr`, `node_version`, `logging_service` / `monitoring_service`, `enable_tpu` | Legacy surfaces superseded by what the spec models: clusters are always VPC-native, node versions live on the pools, observability uses the component configs, and TPU capacity is provisioned through TPU node pools. |
| `network_policy.provider` | CALICO is the only legal value — the module wires it with `enable_network_policy`. |
| `workload_identity_config.workload_pool` | The API fixes the pool name to `PROJECT_ID.svc.id.goog`; the spec holds the on/off decision and the module composes the only possible value. |
| `maintenance_policy.recurring_maintenance_window`, `rollback_safe_upgrade` + `desired_emulated_version`, the node-readiness addon | GA at the pinned provider but not yet bridged by the pinned Pulumi SDK — modeling them only in Terraform would break cross-engine parity. Re-evaluated at every SDK bump. |
| `tpu_config`, `pod_security_policy_config`, `cluster_telemetry`, `protect_config` and other beta-only blocks | Exist only in the `google-beta` provider; GA is the parity baseline, and beta surface enters only through the catalog's admission list (`pkg/providerparity/admissions/google-beta.yaml`), which admits resources, not fields -- none of these blocks is admitted for this kind. |
| `enterprise_config` | Deprecated on the provider at the pinned version. |

## Related Components

- **GcpVpcNetwork** — the network the cluster lives in
- **GcpSubnetwork** — carries the primary node range and pod/service secondary ranges
- **GcpGkeNodePool** — compute for Standard clusters (references this cluster's `name` output)
- **GcpRouterNat** — outbound internet for private nodes
- **GcpKmsKey** — CMEK key for etcd secrets encryption and NAP boot disks
- **GcpServiceAccount** — node identity for NAP-created pools
- **GcpPubSubTopic** — receives cluster lifecycle notifications
- **GcpBigQueryDataset** — receives resource-usage export records
- **GcpGkeWorkloadIdentityBinding** — binds Kubernetes service accounts to IAM service accounts on the cluster's workload pool

## Additional Resources

- [GKE Documentation](https://cloud.google.com/kubernetes-engine/docs)
- [Clusters REST API](https://cloud.google.com/kubernetes-engine/docs/reference/rest/v1/projects.locations.clusters)

---

© Planton. Licensed under [Apache-2.0](https://github.com/plantonhq/planton/blob/main/LICENSE).
