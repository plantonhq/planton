# GcpGkeCluster - Terraform Module

This Terraform module provisions a Google Kubernetes Engine cluster (`google_container_cluster`). It is the Terraform-side implementation of the Planton `GcpGkeCluster` resource kind and has feature parity with the Pulumi module.

## Overview

The module enables the Kubernetes Engine API (`container.googleapis.com`) so a fresh project works first try, then creates the cluster with the full released-provider surface: VPC-native IP allocation (named ranges, GKE-managed, additional subnetwork ranges, or auto-IPAM), private topology (peering- or PSC-based control planes), master authorized networks and the DNS endpoint (including token/cert serving via DNS), release channels, maintenance scheduling with disruption budgets, node auto-provisioning with upgrade rollout defaults and default compute classes, Dataplane V2 features, CMEK etcd encryption and customer-managed control-plane keys/CAs, Binary Authorization, Security Posture, RBAC binding lockdown, per-component observability with managed Prometheus, Pub/Sub notifications, BigQuery usage export, the full addon set (including Cloud Run, Ray, Lustre/Parallelstore CSI, Slurm), Secret Manager CSI rotation and the sync add-on, node-pool creation defaults (image streaming, containerd registry access), fleet registration, engine-side deletion policy, and Autopilot mode with its policy, privileged-admission, and managed-node controls.

On Standard clusters the API-mandated default node pool is removed at create time — compute comes from `GcpGkeNodePool` resources. On Autopilot clusters the node-management fields are omitted entirely (GKE owns nodes). An empty `project_id` falls back to the provider's default project; empty optional strings become null so the provider omits them from the API payload. The module runs on the plain `google` provider — every modeled field is GA on the pinned 8.x line, and the kind's `iac/provider-parity.yaml` records the full argument accounting.

`deletion_protection` defaults to true, matching GCP: a destroy plan fails until the spec explicitly sets it false.

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
cd catalog/gcp/gcpgkecluster/iac/tf
terraform init
terraform plan -var-file=terraform.tfvars.json
terraform apply -var-file=terraform.tfvars.json
```

## Variables

| Name | Description | Default |
|------|-------------|---------|
| `metadata` | Resource metadata (name, labels, etc.) | — |
| `spec` | GcpGkeCluster spec | — |

The `spec` object includes: `location` + `network` + `subnetwork` (required), `cluster_name` (empty = metadata.name), `description`, `node_locations`, `resource_labels`, `deletion_protection` + `deletion_policy`, `ip_allocation` (named ranges xor CIDR blocks, stack type, additional pod ranges and subnetwork ranges, auto-IPAM, network tier, overprovision toggle), `datapath_provider` + `dataplane_optimization_mode`, `default_max_pods_per_node`, the Dataplane V2 toggles (FQDN/Cilium policy, multi-networking, intranode visibility, in-transit encryption), `private_cluster`, `master_authorized_networks`, `control_plane_endpoints` (DNS endpoint incl. token/cert serving), `issue_client_certificate`, `node_creation_mode`, `release_channel` + `min_master_version` + `gke_auto_upgrade_patch_mode`, `maintenance_policy` (windows, exclusions, disruption budget), `cluster_autoscaling` (NAP incl. upgrade-rollout defaults and default compute classes), `enable_vertical_pod_autoscaling` + `hpa_profile`, `workload_identity_enabled`, `enable_shielded_nodes`, `database_encryption`, `user_managed_keys`, `binary_authorization_evaluation_mode`, `security_posture`, `rbac_binding_config`, `confidential_nodes`, `anonymous_authentication_mode`, `enable_kubernetes_alpha` + `k8s_beta_apis`, `logging` + `monitoring`, `notification_pubsub`, `enable_cost_management`, `resource_usage_export`, `addons` (incl. Cloud Run, Ray sub-toggles, Parallelstore/Lustre CSI, pod snapshot, agent sandbox, slice controller, Slurm), `enable_secret_manager_csi` + `secret_manager_rotation` + `secret_sync`, `node_pool_defaults`, `enable_autopilot` + `allow_net_admin` + `autopilot_policy` + `autopilot_privileged_admission_paths` + `node_pool_auto_config`, `fleet_project` + `fleet_membership_type`, `ignore_node_count_changes` + `skip_node_pool_refresh`, and `project_id` (empty falls back to the provider default project).

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
