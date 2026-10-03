# GCP Cloud Composer Environment

Deploys a Google Cloud Composer environment (`google_composer_environment`) — managed Apache Airflow for authoring, scheduling, and monitoring data pipelines. Composer provisions and operates the underlying GKE cluster, Cloud SQL metadata database, web server, and Cloud Storage DAG bucket behind a single resource.

## Overview

Cloud Composer lets teams run production Airflow without operating Kubernetes, databases, or storage themselves. You declare the environment — sizing, networking, software, security — and upload DAGs to a bucket; Composer handles everything underneath.

This kind targets **Composer 2.x and 3**. Composer 1.x is a deprecated generation and its fields are excluded. Both networking models are covered: VPC peering (Composer 2.x) and Private Service Connect, including Composer 3's network-attachment entry point.

**Timing note**: environment creation takes 25-45 minutes — Composer assembles a GKE cluster, Cloud SQL database, and web server behind the scenes.

## Quick Start

```yaml
apiVersion: gcp.planton.dev/v1alpha1
kind: GcpCloudComposerEnvironment
metadata:
  name: dev-airflow
spec:
  region: us-central1
  environmentSize: ENVIRONMENT_SIZE_SMALL
```

```shell
planton apply -f composer.yaml
```

This creates a small public-endpoint environment in the provider's default project. The `airflow_uri` output is the web UI; upload DAGs to the `dag_gcs_prefix` path.

## Configuration Options

| Category | Options |
|----------|---------|
| **Identity** | `project_id` (optional — provider default when omitted), `region` (required, immutable), `environment_name` (defaults to `metadata.name`, immutable) |
| **Capacity** | `environment_size` (`SMALL`, `MEDIUM`, `LARGE`, `EXTRA_LARGE`), `resilience_mode` (`STANDARD_RESILIENCE` or `HIGH_RESILIENCE`) |
| **Networking (Composer 2.x)** | `node_config.network` / `subnetwork` / `service_account` / `tags`; `private_environment_config` (private endpoint, `VPC_PEERING` or `PRIVATE_SERVICE_CONNECT`, CIDR ranges) |
| **Networking (Composer 3)** | `node_config.composer_network_attachment` (+ `composer_internal_ipv4_cidr_block`), `enable_private_environment`, `enable_private_builds_only` |
| **VPC-native ranges** | `node_config.ip_allocation_policy` — pod and services ranges, each a named secondary range XOR a CIDR; `node_config.enable_ip_masq_agent` for pod-to-node SNAT |
| **Software** | `software_config` — `image_version`, `airflow_config_overrides`, `pypi_packages`, `env_variables` (plain text), `secret_env_variables` (each value stored in its own Secret Manager secret; the env var holds the version's resource name), `web_server_plugins_mode` (Composer 3), `cloud_data_lineage_integration` |
| **Workloads** | `workloads_config` — per-kind CPU/memory/storage for scheduler, web server, workers (autoscaling `min_count`/`max_count`), triggerer, and DAG processor (Composer 3) |
| **Security** | `kms_key_name` (CMEK for all Composer-managed resources), `web_server_network_access_control` (UI IP allowlist), `master_authorized_networks_config` (GKE control-plane allowlist) |
| **Operations** | `maintenance_window` (RFC3339 window + RRULE recurrence, min 12 hours), `recovery_config` (scheduled snapshots), `data_retention_config` (task log storage mode, Airflow metadata retention 30-730 days) |
| **Storage** | `storage_bucket` — an existing bucket for DAGs/plugins/data instead of the auto-created one |
| **Labels** | `labels` — user labels merged beneath platform attribution labels |

**Immutable fields** (require environment replacement if changed): `region`, `environment_name`, all node networking (`network`, `subnetwork`, network attachment, IP allocation), `private_environment_config`, `kms_key_name`, and `storage_bucket`. Workload sizing, environment size, resilience mode, software configuration, maintenance window, access control, and labels update in place.

## Outputs

| Output | Type | Description |
|--------|------|-------------|
| `environment_id` | string | Fully qualified resource ID (`projects/{project}/locations/{region}/environments/{name}`) |
| `environment_name` | string | Short name of the environment |
| `airflow_uri` | string | URL of the Apache Airflow web UI |
| `dag_gcs_prefix` | string | Cloud Storage prefix for DAG uploads (`gs://{bucket}/dags`) |
| `gke_cluster` | string | Name of the underlying GKE cluster managed by Composer |

## Important Notes

- **Creation takes 25-45 minutes** — plan pipelines and timeouts accordingly; updates that touch immutable fields recreate the environment and take just as long.
- **Plan CIDRs before deploying**: all networking is immutable. The master, Cloud SQL, Composer-internal, and pod/services ranges must not overlap anything in your network.
- **The triggerer block requires all three fields** (`cpu`, `memory_gb`, `count`) when present — the API rejects partial triggerer configs.
- **Per-range XOR in `ip_allocation_policy`**: name an existing secondary range or give a CIDR for GKE to carve — never both for the same range.
- **Composer 2.x vs 3 fields**: `private_environment_config` and `task_logs_storage_mode` are Composer 2.x surfaces; `composer_network_attachment`, `enable_private_environment`, `web_server_plugins_mode`, the DAG processor, and Airflow metadata retention are Composer 3 surfaces. The image version decides which apply.
- **`deletionPolicy` controls what a destroy does**: `DELETE` (default), `PREVENT` (destroy fails — protects the environment a data platform runs on), or `ABANDON` (drop from management, keep the environment — and its meaningful idle bill — running). The auto-created DAG bucket survives a DELETE either way.

### Deliberately not modeled (recorded reasons)

- **Composer 1.x fields** (`node_count`; node_config `zone`, `machine_type`, `disk_size_gb`, `oauth_scopes`; `ip_allocation_policy.use_ip_aliases`; software_config `python_version`, `scheduler_count`; `database_config`; `web_server_config`; private_environment_config `web_server_ipv4_cidr_block`) — Composer 1 is a deprecated generation.

## Related Kinds

- **GcpVpcNetwork** / **GcpSubnetwork** — the network and subnetwork for VPC peering deployments
- **GcpServiceAccount** — the node identity (must hold `roles/composer.worker`)
- **GcpKmsKey** — the CMEK key for encryption at rest
- **GcpGcsBucket** — a governed bucket for `storage_bucket` or snapshot storage
- **GcpCloudComposerUserWorkloadsSecret** — credentials for DAGs, delivered into this environment
- **GcpCloudComposerUserWorkloadsConfigMap** — non-secret configuration for DAGs
- **GcpProject** — provides the GCP project ID

## Deployment Methods

### Pulumi (Go)

See [`iac/pulumi/README.md`](iac/pulumi/README.md).

### Terraform

See [`iac/tf/README.md`](iac/tf/README.md).

For copy-paste ready manifests, see the [presets](presets/).

## Additional Resources

- [Cloud Composer Documentation](https://cloud.google.com/composer/docs)
- [Composer Versioning Overview](https://cloud.google.com/composer/docs/concepts/versioning/composer-versioning-overview)
- [Apache Airflow Documentation](https://airflow.apache.org/docs/)

---

© Planton. Licensed under [Apache-2.0](https://github.com/plantonhq/planton/blob/main/LICENSE).
