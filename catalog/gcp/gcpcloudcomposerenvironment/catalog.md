# GCP Cloud Composer Environment

Deploys a Cloud Composer environment -- a managed Apache Airflow service -- with configurable environment sizing, Airflow workload resource allocation (scheduler, workers, web server, triggerer), private networking via VPC peering or Private Service Connect, CMEK encryption, maintenance windows, and scheduled snapshot recovery. Supports both Composer 2.x and Composer 3 — the two generations configure networking differently, and the image version decides which rules apply.

## What Gets Created

When you deploy this Infra Component, the IaC module provisions:

- **Cloud Composer API enablement** (`composer.googleapis.com`) on the target project (never disabled on destroy)
- **Cloud Composer Environment** -- a managed `composer.Environment` in the specified GCP project and region, with the chosen environment size (SMALL, MEDIUM, LARGE, or EXTRA_LARGE) and resilience mode
- **Managed GKE Cluster** -- provisioned and managed by Cloud Composer to run Airflow components (scheduler, workers, web server, triggerer, DAG processor)
- **Cloud SQL Metadata Database** -- an internal managed database for Airflow metadata (task history, DAG state, connections)
- **Cloud Storage DAG Bucket** -- a GCS bucket where DAG files are uploaded and synced to the Airflow scheduler
- **Software Configuration** -- Airflow image version, custom PyPI packages, Airflow configuration overrides, and environment variables applied to all Airflow components
- **Workload Resource Allocation** -- CPU, memory, and storage allocations for each Airflow component, with autoscaling bounds for workers
- **Private Environment Configuration** -- created only when `privateEnvironmentConfig` is set; configures VPC peering or Private Service Connect networking with optional private web server endpoint
- **Encryption Configuration** -- created only when `kmsKeyName` is set; encrypts all Composer-managed resources (GKE nodes, Cloud SQL, Cloud Storage) with a customer-managed key
- **Maintenance Window** -- created only when `maintenanceWindow` is set; defines when GCP may perform scheduled maintenance
- **Recovery Configuration** -- created only when `recoveryConfig` is set; enables scheduled snapshots for disaster recovery
- **Web Server Access Control** -- created only when `webServerNetworkAccessControl` is set; restricts Airflow UI access to specified IP ranges
- **GCP Labels** -- resource metadata labels applied automatically for tracking and governance

## Before You Deploy

### Planton Setup

- **GCP Provider Connection** -- an active connection in the Connect module with credentials for the target GCP project. Map it as the default for your environment, or specify it explicitly when creating the Infra Component.
- **Planton Runner** -- required when using Runner-based credential delivery. Not needed for inline credentials or browser OAuth authentication modes.

### GCP Project

- **A GCP project** where the Composer environment will be created. Provide the project ID directly or reference a GcpProject Infra Component via ValueFromRef.
- **A VPC network and subnet** (if using private networking) -- required for Composer 2.x with VPC peering. Provide self-links directly or reference GcpVpcNetwork and GcpSubnetwork Infra Components via ValueFromRef.
- **A service account** (recommended) -- a custom service account for Composer GKE nodes with permissions for BigQuery, GCS, and other GCP services your DAGs access.
- **Cloud KMS key** (if using CMEK) -- a key in the same region as the environment, with the Composer service agent granted the `cloudkms.cryptoKeyEncrypterDecrypter` role.

## Deploy

### Console

Open the deployment store, find **GCP Cloud Composer Environment**, and click **Deploy**. The creation wizard walks you through preset selection, environment and connection configuration, and spec fields. Start from the **Dev Small** preset in the [Presets](#presets) tab to pre-populate a minimal development environment.

### CLI

Create a manifest and apply it:

```yaml
apiVersion: gcp.planton.dev/v1alpha1
kind: GcpCloudComposerEnvironment
metadata:
  name: data-pipelines
  org: acme-corp
  env: prod
spec:
  projectId:
    value: "acme-prod-12345"
  region: us-central1
  environmentSize: ENVIRONMENT_SIZE_SMALL
  softwareConfig:
    imageVersion: "composer-2.9.7-airflow-2.9.3"
```

```shell
planton apply -f cloud-composer.yaml
```

This creates a small Composer 2.x environment with public endpoint access, default workload resource allocations, and no private networking or CMEK encryption. An Infra Job tracks the provisioning in real time.

### InfraChart

When deploying as part of a multi-resource environment, use ValueFromRef to wire the Composer environment to a GCP project, VPC, subnet, KMS key, and service account deployed in the same InfraPipeline:

```yaml
spec:
  projectId:
    valueFrom:
      kind: GcpProject
      name: data-project
      fieldPath: status.outputs.project_id
  nodeConfig:
    network:
      valueFrom:
        kind: GcpVpcNetwork
        name: data-vpc
        fieldPath: status.outputs.network_self_link
    subnetwork:
      valueFrom:
        kind: GcpSubnetwork
        name: composer-subnet
        fieldPath: status.outputs.subnetwork_self_link
    serviceAccount:
      valueFrom:
        kind: GcpServiceAccount
        name: composer-sa
        fieldPath: status.outputs.email
  kmsKeyName:
    valueFrom:
      kind: GcpKmsKey
      name: composer-encryption-key
      fieldPath: status.outputs.key_id
```

The InfraPipeline resolves the dependency graph, deploys the project, VPC, subnet, service account, and KMS key first, then provisions the Composer environment with private networking and CMEK encryption.

## Key Configuration

These are the most important decisions when configuring a Cloud Composer environment. Explore the full field reference in the [API Explorer](#api-explorer) tab.

**Environment size and resilience** -- `environmentSize` controls managed infrastructure capacity: SMALL for development, MEDIUM for production, LARGE for enterprise-scale workloads. Set `resilienceMode` to `HIGH_RESILIENCE` for multi-zone redundancy (available in Composer 2.1.15+). Size and resilience mode directly impact cost.

**Workload resource allocation** -- Configure CPU, memory, and storage for each Airflow component via `workloadsConfig`. Key decisions: scheduler replica count (2+ for production), worker autoscaling bounds (`minCount`/`maxCount`), and triggerer allocation for deferrable operators. Under-provisioned schedulers cause DAG parsing delays; under-provisioned workers cause task queuing.

**Private networking** -- For Composer 2.x, configure `privateEnvironmentConfig` with VPC peering (`connectionType: VPC_PEERING`) or Private Service Connect. Set `enablePrivateEndpoint: true` to restrict the Airflow web UI to private IP only. For Composer 3, use `enablePrivateEnvironment` and `nodeConfig.composerNetworkAttachment` instead.

**Software and packages** -- Set `softwareConfig.imageVersion` to pin a specific Composer/Airflow version (e.g., `"composer-2.9.7-airflow-2.9.3"`). Add custom Python packages via `pypiPackages` and override Airflow configuration via `airflowConfigOverrides`. Environment variables set via `envVariables` are available to all DAGs as plain text. Put secrets in `secretEnvVariables` instead: the module stores each value in its own Secret Manager secret, grants `nodeConfig.serviceAccount` (or the default compute account) access to it alone, and the env var holds the version's resource name for DAGs to read. Airflow's own Secret Manager backend (`secrets-backend` in `airflowConfigOverrides`) remains the way to serve Airflow connections and variables.

## Outputs and Dependencies

### What This Kind Consumes

| Dependency | Field | ValueFromRef Path |
|------------|-------|-------------------|
| **GcpProject** | `projectId` | `status.outputs.project_id` |
| **GcpVpcNetwork** (optional) | `nodeConfig.network` | `status.outputs.network_self_link` |
| **GcpSubnetwork** (optional) | `nodeConfig.subnetwork` | `status.outputs.subnetwork_self_link` |
| **GcpServiceAccount** (optional) | `nodeConfig.serviceAccount` | `status.outputs.email` |
| **GcpKmsKey** (optional) | `kmsKeyName` | `status.outputs.key_id` |
| **GcpGcsBucket** (optional) | `storageBucket` | `status.outputs.bucket_id` |

### What This Kind Provides

After provisioning, `status.outputs` contains values that downstream Infra Components can consume via ValueFromRef:

| Output | Description | Common Downstream Use |
|--------|-------------|----------------------|
| `environment_id` | Fully qualified environment ID (`projects/{project}/locations/{region}/environments/{name}`) | Monitoring dashboards, IAM bindings |
| `environment_name` | Short name of the Composer environment | Inventory tracking, script references |
| `airflow_uri` | Apache Airflow web UI URL | User access to DAG management and monitoring |
| `dag_gcs_prefix` | Cloud Storage path for DAG uploads (`gs://{bucket}/dags`) | CI/CD pipelines deploying DAG files |
| `gke_cluster` | Name of the underlying managed GKE cluster (Composer 2 only — empty for Composer 3, whose cluster lives in a Google tenant project) | Advanced debugging, monitoring |

## Common Patterns

Browse the [Presets](#presets) tab for ready-to-deploy configurations.

**Development small** -- Minimal infrastructure footprint with small resource allocations, public endpoint, and no advanced features. Optimized for cost in development and testing environments. Start from the **Dev Small** preset.

**Production private** -- Medium-sized environment with VPC peering, private endpoint, HIGH_RESILIENCE mode, scaled workloads (2 schedulers, 2 triggerers, 2-6 workers), and a weekend maintenance window. The recommended configuration for production data pipelines. Start from the **Production Private** preset.

**Enterprise encrypted** -- Large-sized environment with CMEK encryption, VPC peering, private endpoint, web server IP allowlist, scheduled daily snapshots for disaster recovery, and generous resource allocations. Designed for organizations with strict security and compliance requirements. Start from the **Enterprise Encrypted** preset.

## Works With

- [**GCP Project**](/infra-catalog/gcp-project) -- provides the GCP project where the Composer environment is created
- [**GCP VPC Network**](/infra-catalog/gcp-vpc-network) -- provides the VPC network for private Composer networking
- [**GCP Subnetwork**](/infra-catalog/gcp-subnetwork) -- provides the subnet for Composer GKE node placement
- [**GCP Service Account**](/infra-catalog/gcp-service-account) -- provides the VM identity for Composer GKE nodes
- [**GCP KMS Key**](/infra-catalog/gcp-kms-key) -- provides the customer-managed encryption key for all Composer-managed resources