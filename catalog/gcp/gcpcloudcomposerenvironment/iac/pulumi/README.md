# GcpCloudComposerEnvironment - Pulumi Module

This Pulumi (Go) module provisions a Cloud Composer environment (`composer.Environment`) — managed Apache Airflow. It is the Pulumi-side implementation of the Planton `GcpCloudComposerEnvironment` resource kind and has feature parity with the Terraform module.

## Overview

Composer assembles a GKE cluster, Cloud SQL metadata database, web server, and DAG bucket behind the one resource — creation takes 25-45 minutes. The module enables the Cloud Composer API so a fresh project works first try. An empty `project_id` falls back to the provider's default project, and an empty `environment_name` falls back to `metadata.name`. User labels are merged beneath Planton's platform attribution labels (platform keys win on conflict), identically to the Terraform module.

## Usage with Planton CLI

```shell
planton pulumi up --manifest ../../e2e/manifest.yaml --module-dir .
planton pulumi destroy --manifest ../../e2e/manifest.yaml --module-dir .
```

Credentials are provided via stack input (by the CLI), not in the manifest `spec`. Manifest file: `../../e2e/manifest.yaml`.

## Direct Pulumi Usage

```bash
cd catalog/gcp/gcpcloudcomposerenvironment/iac/pulumi
make build
pulumi up --stack dev
```

## Module Layout

- `main.go` — entrypoint; loads the stack input and calls the module
- `module/main.go` — provider setup and resource orchestration
- `module/locals.go` — metadata-derived values and the label merge
- `module/composer_environment.go` — API enablement + the environment resource with all configuration blocks
- `module/env_secrets.go` — one Secret Manager secret, pinned version, and accessor grant per `secret_env_variables` entry; merges the version names into `env_variables`
- `module/outputs.go` — stack output keys (must match `outputs.proto`)

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
