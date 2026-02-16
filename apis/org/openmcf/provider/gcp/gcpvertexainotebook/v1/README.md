# GCP Vertex AI Notebook

Deploys a managed Vertex AI Workbench instance (JupyterLab notebook) on Google Cloud Platform.

## Overview

GcpVertexAiNotebook provisions a [Vertex AI Workbench instance](https://cloud.google.com/vertex-ai/docs/workbench/instances/introduction) -- a managed JupyterLab environment for data science and machine learning workflows. Each instance is a Compute Engine VM pre-configured with JupyterLab, ML frameworks, and optional GPU accelerators. Users access their notebooks through a secure proxy URL.

## When to Use

- Data scientists need a managed JupyterLab environment with GPU support
- ML engineers need reproducible notebook environments for training and experimentation
- Teams need notebooks with controlled VPC networking and CMEK encryption
- Organizations want centralized management of notebook infrastructure

## Key Features

- **Pre-built ML images** -- TensorFlow, PyTorch, JAX, and other frameworks pre-installed
- **GPU accelerators** -- NVIDIA Tesla T4, A100, L4, and other GPUs for training
- **Custom containers** -- bring your own Docker image for specialized environments
- **Private networking** -- deploy inside a VPC with no public IP for security
- **CMEK encryption** -- encrypt boot and data disks with customer-managed KMS keys
- **Cost management** -- stop instances when not in use (desired_state: STOPPED)
- **Shielded VM** -- Secure Boot, vTPM, and integrity monitoring support

## Quick Start

```yaml
apiVersion: gcp.openmcf.org/v1
kind: GcpVertexAiNotebook
metadata:
  name: my-notebook
spec:
  projectId:
    value: my-gcp-project
  location: us-central1-a
  machineType: e2-standard-4
```

This creates a CPU-only notebook with the default deep learning VM image, accessible via JupyterLab proxy URL.

## Configuration Highlights

### Machine Types

Choose based on workload:
- **CPU-only** (data processing, light ML): `e2-standard-4`, `e2-standard-8`
- **GPU training** (requires N1/A2): `n1-standard-8` + `NVIDIA_TESLA_T4`, `a2-highgpu-1g`

### Image Selection

Two mutually exclusive options:
- **VM image** (default): pre-built deep learning images from `deeplearning-platform-release`
- **Container image**: custom Docker image from any registry

### Networking

- Default: public IP with JupyterLab accessible via proxy URL
- Private: set `disablePublicIp: true` and configure VPC network/subnet

### Storage

- **Boot disk**: OS and JupyterLab (default 150 GB PD_SSD)
- **Data disk**: user notebooks and data (default 100 GB PD_STANDARD)
- Both support CMEK encryption via KMS key references

## Related Components

- **GcpProject** -- project where the notebook is created
- **GcpVpc / GcpSubnetwork** -- VPC networking for private instances
- **GcpServiceAccount** -- VM identity for accessing GCP resources
- **GcpKmsKey** -- encryption keys for CMEK-encrypted disks
- **GcpGcsBucket** -- storage for notebooks and datasets
- **GcpBigQueryDataset** -- data warehouse for ML pipelines
