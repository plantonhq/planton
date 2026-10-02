# GcpCloudComposerEnvironment - Terraform Module

This Terraform module provisions a Cloud Composer environment (`google_composer_environment`) — managed Apache Airflow. It is the Terraform-side implementation of the Planton `GcpCloudComposerEnvironment` resource kind and has feature parity with the Pulumi module.

## Overview

Composer assembles a GKE cluster, Cloud SQL metadata database, web server, and DAG bucket behind the one resource — creation takes 25-45 minutes. The module enables the Cloud Composer API (`disable_on_destroy=false`) so a fresh project works first try. An empty `project_id` falls back to the provider's default project, and an empty `environment_name` falls back to `metadata.name`. User labels are merged beneath Planton's platform attribution labels (platform keys win on conflict), identically to the Pulumi module.

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
cd catalog/gcp/gcpcloudcomposerenvironment/iac/tf
terraform init
terraform plan -var-file=terraform.tfvars.json
terraform apply -var-file=terraform.tfvars.json
```

## Variables

| Name | Description | Default |
|------|-------------|---------|
| `metadata` | Resource metadata (name, labels, etc.) | — |
| `spec` | GcpCloudComposerEnvironment spec | — |

The `spec` object includes: identity (`project_id`, `region`, `environment_name`), capacity (`environment_size`, `resilience_mode`), networking (`node_config` with VPC peering or Composer 3 attachment, `ip_allocation_policy`, `enable_ip_masq_agent`; `private_environment_config`; Composer 3 flags), software (`software_config` incl. `cloud_data_lineage_integration`), workloads (`workloads_config` for all five components), security (`kms_key_name`, `web_server_network_access_control`, `master_authorized_networks_config`), operations (`maintenance_window`, `recovery_config`, `data_retention_config`), `storage_bucket`, and `labels`.

## Outputs

| Name | Description |
|------|-------------|
| `environment_id` | Fully qualified resource ID (`projects/{project}/locations/{region}/environments/{name}`) |
| `environment_name` | Short name of the Composer environment |
| `airflow_uri` | URI of the Apache Airflow web UI |
| `dag_gcs_prefix` | Cloud Storage prefix for DAG file uploads |
| `gke_cluster` | Name of the underlying GKE cluster managed by Composer |

## Lifecycle Notes

The immutables (ForceNew): `region`, `environment_name`, node networking (network, subnetwork, network attachment, IP allocation), the private environment configuration, `kms_key_name`, and `storage_bucket`. Workload sizing, environment size, resilience mode, software configuration, maintenance window, access control, retention, and labels update in place. Each `secret_env_variables` entry gets its own Secret Manager secret (user-managed replication in `region`), a pinned version with `create_before_destroy`, and a `secretmanager.secretAccessor` grant on that secret alone to `node_config.service_account`, or to the default compute account when it is empty; the environment's env var holds the version's resource name and the environment depends on the grant. The triggerer block sends `cpu`, `memory_gb`, and `count` unconditionally — the API requires all three when the block is present.
