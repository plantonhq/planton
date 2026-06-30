# GcpDataprocVirtualCluster - Terraform Module

Terraform implementation for the GcpDataprocVirtualCluster deployment component.

## Overview

This module provisions a Dataproc on GKE virtual cluster using the `google_dataproc_cluster` resource with `virtual_cluster_config` instead of the standard `cluster_config`. Spark workloads are scheduled as Kubernetes pods on an existing GKE cluster.

## Provider

Requires Google Cloud provider `~> 6.0`.

```hcl
terraform {
  required_providers {
    google = {
      source  = "hashicorp/google"
      version = "~> 6.0"
    }
  }
}
```

## Usage

This module is designed to be called by the Planton framework. Direct usage requires providing the `spec` and `metadata` variables matching the protobuf-defined schema.

```hcl
module "dataproc_virtual_cluster" {
  source = "./iac/tf"

  metadata = {
    name = "my-spark-on-gke"
  }

  spec = {
    project_id = {
      value = "my-gcp-project"
    }
    region = "us-central1"
    gke_cluster_target = {
      value = "projects/my-gcp-project/locations/us-central1/clusters/my-gke-cluster"
    }
    software_config = {
      component_version = {
        SPARK = "3.5-dataproc-17"
      }
    }
    node_pool_targets = [
      {
        node_pool = {
          value = "default-pool"
        }
        roles = ["DEFAULT"]
      }
    ]
  }
}
```

## Resources Created

- `google_dataproc_cluster` — Dataproc cluster with `virtual_cluster_config` targeting an existing GKE cluster, including node pool role assignments, Kubernetes software configuration, and optional auxiliary services.

## Input Variables

### `spec` (required)

The `GcpDataprocVirtualClusterSpec` object:

| Field | Type | Required | Description |
|---|---|---|---|
| `project_id` | `object({ value = string })` | Yes | GCP project ID |
| `region` | `string` | Yes | GCP region (must match GKE cluster) |
| `cluster_name` | `string` | No | Explicit cluster name (defaults to metadata.name) |
| `gke_cluster_target` | `object({ value = string })` | Yes | Fully qualified GKE cluster resource ID |
| `kubernetes_namespace` | `object({ value = string })` | No | Kubernetes namespace for the virtual cluster |
| `staging_bucket` | `object({ value = string })` | No | GCS bucket for staging job artifacts |
| `software_config` | `object` | Yes | Component versions and Spark properties |
| `node_pool_targets` | `list(object)` | Yes | GKE node pool role assignments |
| `auxiliary_services_config` | `object` | No | Metastore and Spark History Server |

### `metadata` (optional)

Resource metadata for naming and labeling:

| Field | Type | Default | Description |
|---|---|---|---|
| `name` | `string` | `""` | Resource name (used as cluster name fallback) |
| `org` | `string` | `""` | Organization identifier |
| `env` | `object({ id = string })` | `null` | Environment identifier |
| `id` | `string` | `""` | Resource ID |

## Outputs

| Output | Description |
|---|---|
| `cluster_id` | Fully qualified Dataproc cluster resource name |
| `cluster_name` | Short name of the Dataproc cluster |
| `cluster_uuid` | Server-generated UUID (not exposed by provider; empty string) |

## File Structure

| File | Purpose |
|---|---|
| `provider.tf` | Google provider configuration |
| `variables.tf` | Input variable definitions |
| `locals.tf` | Computed values: cluster name, labels |
| `main.tf` | `google_dataproc_cluster` resource with `virtual_cluster_config` |
| `outputs.tf` | Output definitions |
