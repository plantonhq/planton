# GcpGkeCluster - Pulumi Module

This Pulumi (Go) module provisions a Google Kubernetes Engine cluster (`container.Cluster`). It is the Pulumi-side implementation of the Planton `GcpGkeCluster` resource kind and has feature parity with the Terraform module.

## Overview

The module enables the Kubernetes Engine API (`container.googleapis.com`) so a fresh project works first try, then creates the cluster with the full released-provider surface: VPC-native IP allocation (named ranges, GKE-managed, additional subnetwork ranges, or auto-IPAM), private topology (peering- or PSC-based control planes), master authorized networks and the DNS endpoint (including token/cert serving via DNS), release channels, maintenance scheduling with disruption budgets, node auto-provisioning with upgrade rollout defaults and default compute classes, Dataplane V2 features, CMEK etcd encryption and customer-managed control-plane keys/CAs, Binary Authorization, Security Posture, RBAC binding lockdown, per-component observability with managed Prometheus, Pub/Sub notifications, BigQuery usage export, the full addon set (including Cloud Run, Ray, Lustre/Parallelstore CSI, Slurm), Secret Manager CSI rotation and the sync add-on, node-pool creation defaults, fleet registration, engine-side deletion policy, and Autopilot mode with its policy, privileged-admission, and managed-node controls. The kind's `iac/provider-parity.yaml` records the full argument accounting against the pinned provider.

On Standard clusters the API-mandated default node pool is removed at create time — compute comes from `GcpGkeNodePool` resources. On Autopilot clusters the node-management fields are omitted entirely (GKE owns nodes). An empty `project_id` falls back to the provider's default project — the ambient-project contract every GCP kind honors. `deletion_protection` defaults to true, matching GCP: a destroy preview fails until the spec explicitly sets it false.

## Usage with Planton CLI

```shell
planton pulumi up --manifest ../../e2e/manifest.yaml --module-dir .
planton pulumi destroy --manifest ../../e2e/manifest.yaml --module-dir .
```

Credentials are provided via IaC input (by the CLI), not in the manifest `spec`. Manifest file: `../../e2e/manifest.yaml`.

## Build

```bash
cd catalog/gcp/gcpgkecluster/iac/pulumi
make build
```

## Outputs

| Name | Description |
|------|-------------|
| `endpoint` | Kubernetes API server IP (private endpoint on private-only control planes) |
| `cluster_ca_certificate` | Base64 CA certificate — public trust material clients install as a trust anchor |
| `workload_identity_pool` | `PROJECT_ID.svc.id.goog`; empty when Workload Identity is disabled on a Standard cluster |
| `cluster_id` | `projects/{project}/locations/{location}/clusters/{name}` |
| `name` | Cluster name in GCP — the handle node pools and gcloud use |
| `location` | Region or zone, exactly as provided in the spec |
| `self_link` | Server-defined URL of the cluster resource |
| `master_version` | Kubernetes version currently on the control plane |
