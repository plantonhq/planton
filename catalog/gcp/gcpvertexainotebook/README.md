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
- **GPU accelerators** -- from the Tesla family through the current generation (L4, A100, H100, H100 Mega, H200, B200, RTX6000) for training
- **Reservation affinity** -- consume pre-purchased Compute reservations to guarantee GPU capacity
- **Custom containers** -- bring your own Docker image for specialized environments
- **Private networking** -- deploy inside a VPC with no public IP, or pin a static external IP by reference
- **CMEK encryption** -- encrypt boot and data disks with customer-managed KMS keys
- **Cost management** -- stop instances when not in use (desired_state: STOPPED)
- **Shielded VM & Confidential Computing** -- Secure Boot, vTPM, integrity monitoring, and AMD SEV memory encryption
- **Per-user identity** -- managed end-user credentials so notebook code acts as the signed-in user
- **User labels** -- cost attribution and ownership tagging, merged with platform labels

## Quick Start

```yaml
apiVersion: gcp.planton.dev/v1alpha1
kind: GcpVertexAiNotebook
metadata:
  name: my-notebook
spec:
  location: us-central1-a
  machineType: e2-standard-4
```

This creates a CPU-only notebook in the provider's default project with the default deep learning VM image, accessible via JupyterLab proxy URL. Set `projectId` (literal or a GcpProject reference) to target another project.

## Configuration Highlights

### Machine Types

Choose based on workload:
- **CPU-only** (data processing, light ML): `e2-standard-4`, `e2-standard-8`
- **GPU training**: `n1-standard-8` + `NVIDIA_TESLA_T4`, `a2-highgpu-1g` (A100), or the A3/A4 series for the current H100/H200/B200 generation

### Image Selection

Two mutually exclusive options:
- **VM image** (default): omit `vmImage` for the service's latest Workbench image, or pin `cloud-notebooks-managed` / `workbench-instances`
- **Container image**: custom Docker image from any registry

### Networking

- Default: ephemeral public IP with JupyterLab accessible via proxy URL
- Private: set `disablePublicIp: true` and configure VPC network/subnet
- Pinned address: reference a `GcpAddress` in `networkInterface.externalIp` when firewall allowlists or DNS depend on the instance IP

### Storage

- **Boot disk**: OS and JupyterLab (default 150 GB PD_SSD); Persistent Disk or hyperdisk types (`HYPERDISK_BALANCED`, `HYPERDISK_BALANCED_HIGH_AVAILABILITY`, `HYPERDISK_ML` on supported machine series)
- **Data disk**: user notebooks and data (default 100 GB PD_STANDARD); additionally `HYPERDISK_EXTREME` and `HYPERDISK_THROUGHPUT`
- Both support CMEK encryption via KMS key references

### Destroy behavior

`deletionPolicy` is the client-side lever: empty/`DELETE` deletes the instance including its disks (unsynced notebook work is lost), `PREVENT` makes destroy fail, `ABANDON` removes it from management but keeps the VM running (and billing).

### Security

- **Shielded VM**: Secure Boot, vTPM, and integrity monitoring
- **Confidential Computing**: AMD SEV memory encryption (requires n2d machine types)
- **Managed EUC**: JupyterLab acts as the signed-in user's identity for per-user auditability

## Related Kinds

- **GcpProject** -- project where the notebook is created
- **GcpVpcNetwork / GcpSubnetwork** -- VPC networking for private instances
- **GcpAddress** -- static external IP pinned by reference
- **GcpServiceAccount** -- VM identity for accessing GCP resources
- **GcpKmsKey** -- encryption keys for CMEK-encrypted disks
- **GcpGcsBucket** -- storage for notebooks and datasets
- **GcpBigQueryDataset** -- data warehouse for ML pipelines

---

© Planton. Licensed under [Apache-2.0](https://github.com/plantonhq/planton/blob/main/LICENSE).
