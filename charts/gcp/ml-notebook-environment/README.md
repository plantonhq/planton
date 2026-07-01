# GCP ML Notebook Environment

Provisions a complete machine learning development environment with a Vertex AI Workbench notebook, BigQuery dataset for analytics, GCS bucket for datasets and model artifacts, a dedicated service account with appropriate IAM roles, and private VPC networking.

A data scientist can start working immediately after deployment: the notebook has BigQuery and GCS access pre-configured, a private network prevents accidental data exposure, and optional GPU acceleration is available for training workloads.

## Architecture

```
┌──────────────────────────────────────────────────────────────────┐
│  Network (optional)                                              │
│                                                                  │
│  ┌──────────────┐       ┌─────────────────────────┐             │
│  │   GcpVpc     │──────▶│    GcpSubnetwork         │             │
│  │              │       │  Private Google Access    │             │
│  └──────────────┘       └────────────┬─────────────┘             │
│                                      │                           │
└──────────────────────────────────────│───────────────────────────┘
                                       │
                                       ▼
                         ┌─────────────────────────┐
                         │  GcpVertexAiNotebook     │
                         │  (JupyterLab + ML libs)  │
                         │  Optional GPU            │
                         └─────────────┬───────────┘
                                       │ uses
                          ┌────────────┼────────────┐
                          ▼            ▼            ▼
                   ┌────────────┐ ┌─────────┐ ┌──────────────────┐
                   │GcpGcsBucket│ │GcpBigQ. │ │GcpServiceAccount │
                   │ (artifacts)│ │ Dataset │ │ (notebook SA)    │
                   └────────────┘ └─────────┘ └──────────────────┘
```

## Dependency Graph

```
Layer 0 (parallel):  GcpVpc, GcpServiceAccount, GcpGcsBucket, GcpBigQueryDataset
Layer 1 (dep VPC):   GcpSubnetwork
Layer 2 (dep all):   GcpVertexAiNotebook
```

## Included Cloud Resources

| Resource | Kind | Group | Purpose |
|----------|------|-------|---------|
| VPC Network | `GcpVpc` | network | Private networking (optional) |
| Subnetwork | `GcpSubnetwork` | network | Subnet with Private Google Access (optional) |
| Service Account | `GcpServiceAccount` | identity | Notebook VM identity with BigQuery, GCS, Vertex AI access |
| GCS Bucket | `GcpGcsBucket` | storage | Datasets, model artifacts, training outputs |
| BigQuery Dataset | `GcpBigQueryDataset` | storage | Analytics data for exploration and feature engineering |
| Vertex AI Notebook | `GcpVertexAiNotebook` | compute | JupyterLab notebook with ML frameworks |

## Parameters

| Parameter | Description | Default | Required |
|-----------|-------------|---------|----------|
| `gcp_project_id` | GCP project ID | `my-gcp-project` | Yes |
| `region` | GCP region for networking, storage, BigQuery | `us-central1` | Yes |
| `zone` | GCP zone for the notebook (must be in region) | `us-central1-a` | Yes |
| `networkingEnabled` | Create VPC and subnet | `true` | No |
| `vpc_name` | VPC network name | `ml-notebook-vpc` | No |
| `subnet_cidr` | Subnet CIDR range | `10.0.0.0/24` | No |
| `notebook_name` | Notebook instance name | `ml-notebook` | Yes |
| `machine_type` | VM machine type | `e2-standard-4` | Yes |
| `gpuEnabled` | Attach GPU accelerator | `false` | No |
| `accelerator_type` | GPU type (if gpuEnabled) | `NVIDIA_TESLA_T4` | No |
| `accelerator_core_count` | GPU cores (if gpuEnabled) | `1` | No |
| `boot_disk_size_gb` | Boot disk size (0 = default 150 GB) | `0` | No |
| `data_disk_size_gb` | Data disk size (0 = default 100 GB) | `0` | No |
| `service_account_id` | Service account ID | `ml-notebook-sa` | Yes |
| `bucket_name` | GCS bucket name (globally unique) | `my-project-ml-artifacts` | Yes |
| `dataset_id` | BigQuery dataset ID | `ml_dataset` | Yes |

## Service Account Roles

The notebook service account is granted:

| Role | Purpose |
|------|---------|
| `roles/bigquery.dataEditor` | Read/write BigQuery tables |
| `roles/bigquery.jobUser` | Run BigQuery queries |
| `roles/storage.objectAdmin` | Read/write GCS bucket objects |
| `roles/aiplatform.user` | Submit Vertex AI training jobs and use endpoints |

## GPU Configuration

To enable GPU acceleration, set `gpuEnabled: true` and choose an appropriate machine type:

| Machine Type | Compatible GPUs |
|-------------|-----------------|
| `n1-standard-*` | NVIDIA_TESLA_T4, NVIDIA_TESLA_P100, NVIDIA_TESLA_V100 |
| `a2-highgpu-1g` | NVIDIA_TESLA_A100 (1 GPU) |
| `g2-standard-*` | NVIDIA_L4 |

Example GPU configuration:

```yaml
machine_type: n1-standard-8
gpuEnabled: true
accelerator_type: NVIDIA_TESLA_T4
accelerator_core_count: 1
```

## Networking

When `networkingEnabled: true` (default), the chart creates:
- A custom-mode VPC with a single subnet
- **Private Google Access** enabled on the subnet (GCS, BigQuery, and Vertex AI APIs are accessed through Google's internal network, not the public internet)
- **No public IP** on the notebook instance (accessed via the Vertex AI proxy URL)

When `networkingEnabled: false`, the notebook uses GCP's default network. This is suitable for quick experiments but not recommended for production.

## Important Notes

- The notebook instance is accessed via the **proxy URI** in the outputs, not a public IP. Use the Vertex AI console or the proxy URL to access JupyterLab.
- `notebook_name`, `zone`, and network configuration are **immutable** after creation.
- BigQuery tables are managed by notebooks and pipelines, not by this chart. The chart creates the **dataset** (container) only.
- The GCS bucket name must be **globally unique** across all GCP projects.
