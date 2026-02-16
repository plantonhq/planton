# GCP Vertex AI Notebook

Deploys a managed Vertex AI Workbench instance (JupyterLab notebook) on a Compute Engine VM with configurable machine type, GPU accelerators, disk encryption, VPC networking, and pre-built or custom container images. Users access notebooks through a secure proxy URL.

## What Gets Created

When you deploy a GcpVertexAiNotebook resource, OpenMCF provisions:

- **Workbench Instance** — a `google_workbench_instance` resource configured with the specified machine type, disks, networking, and image
- **Boot Disk** (optional configuration) — persistent disk for the OS and JupyterLab runtime, with optional CMEK encryption via a KMS key
- **Data Disk** (optional configuration) — persistent disk for user notebooks and datasets, with optional CMEK encryption
- **GPU Accelerator** (created only when `acceleratorConfig` is set) — an NVIDIA GPU attached to the VM for ML training workloads
- **Framework Labels** — OpenMCF resource labels applied automatically to the instance for tracking and governance

## Prerequisites

- **GCP credentials** configured via environment variables or OpenMCF provider config
- **A GCP project** with the Notebooks API enabled (`notebooks.googleapis.com`)
- **A zone** in a region that supports Workbench instances (most GCP zones)
- **A VPC network and subnet** if deploying with `disablePublicIp: true` (private networking)
- **A service account** if specifying a custom VM identity (recommended for production)
- **A KMS key** if using CMEK encryption for boot or data disks — the key must be in the same region as the instance
- **GPU quota** in the target zone if using `acceleratorConfig`

## Quick Start

Create a file `notebook.yaml`:

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

Deploy:

```shell
openmcf apply -f notebook.yaml
```

This creates a CPU-only Workbench instance with a default deep learning VM image, 150 GB boot disk, and JupyterLab accessible via the proxy URI in the stack outputs.

## Configuration Reference

### Required Fields

| Field | Type | Description | Validation |
|-------|------|-------------|------------|
| `projectId` | `StringValueOrRef` | GCP project where the instance is created. Can reference a GcpProject resource via `valueFrom`. | Required |
| `location` | `string` | GCP zone for the instance (e.g., `us-central1-a`). Immutable after creation. | Required. Pattern: `^[a-z]+-[a-z]+[0-9]-[a-z]$` |
| `machineType` | `string` | Compute Engine machine type (e.g., `e2-standard-4`, `n1-standard-8`). | Required. Min length: 1 |

### Optional Fields

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `instanceName` | `string` | `metadata.name` | Explicit GCP instance name. Immutable. Must be a valid RFC1035 hostname. |
| `instanceOwners` | `list(string)` | `[]` | Owner email(s). Currently GCP supports one owner. Sets Single User access mode. Immutable. |
| `desiredState` | `string` | `ACTIVE` | Instance state: `ACTIVE` (running) or `STOPPED` (suspended, no compute charges). |
| `disableProxyAccess` | `bool` | `false` | If `true`, no JupyterLab proxy URL is generated. Immutable. |
| `metadata` | `map(string)` | `{}` | Custom metadata key-value pairs for the VM. |
| `bootDisk.diskType` | `string` | `PD_SSD` | Boot disk type: `PD_STANDARD`, `PD_SSD`, `PD_BALANCED`, `PD_EXTREME`. |
| `bootDisk.diskSizeGb` | `int` | `150` | Boot disk size in GB. Range: 10-64000. |
| `bootDisk.kmsKey` | `StringValueOrRef` | — | KMS key for CMEK encryption. Can reference GcpKmsKey via `valueFrom`. Immutable. |
| `dataDisk.diskType` | `string` | `PD_STANDARD` | Data disk type: `PD_STANDARD`, `PD_SSD`, `PD_BALANCED`, `PD_EXTREME`. |
| `dataDisk.diskSizeGb` | `int` | `100` | Data disk size in GB. Range: 10-64000. |
| `dataDisk.kmsKey` | `StringValueOrRef` | — | KMS key for CMEK encryption. Can reference GcpKmsKey via `valueFrom`. Immutable. |
| `acceleratorConfig.type` | `string` | — | GPU type: `NVIDIA_TESLA_T4`, `NVIDIA_L4`, `NVIDIA_TESLA_A100`, `NVIDIA_A100_80GB`, etc. |
| `acceleratorConfig.coreCount` | `int` | — | Number of GPU cores (typically 1, 2, 4, or 8). |
| `networkInterface.network` | `StringValueOrRef` | default VPC | VPC network. Can reference GcpVpc via `valueFrom`. Immutable. |
| `networkInterface.subnet` | `StringValueOrRef` | — | Subnet. Can reference GcpSubnetwork via `valueFrom`. Immutable. |
| `networkInterface.nicType` | `string` | `VIRTIO_NET` | NIC type: `VIRTIO_NET` or `GVNIC`. Immutable. |
| `disablePublicIp` | `bool` | `false` | If `true`, no external IP. Instance accessible only via proxy or VPN. Immutable. |
| `enableIpForwarding` | `bool` | `false` | Enable IP forwarding on the VM. Immutable. |
| `serviceAccount` | `StringValueOrRef` | compute default SA | Service account email for VM identity. Can reference GcpServiceAccount via `valueFrom`. Immutable. |
| `tags` | `list(string)` | `[]` | Network tags for firewall rule targeting. Immutable. |
| `vmImage.project` | `string` | `deeplearning-platform-release` | Image project. |
| `vmImage.family` | `string` | — | Image family (e.g., `common-cpu-notebooks`, `tf-latest-gpu`). Mutually exclusive with `vmImage.name`. Immutable. |
| `vmImage.name` | `string` | — | Specific image name. Mutually exclusive with `vmImage.family`. Immutable. |
| `containerImage.repository` | `string` | — | Container image repo (e.g., `gcr.io/project/image`). Required if `containerImage` is set. Mutually exclusive with `vmImage`. |
| `containerImage.tag` | `string` | `latest` | Container image tag. |
| `shieldedInstanceConfig.enableSecureBoot` | `bool` | `false` | Enable Secure Boot. |
| `shieldedInstanceConfig.enableVtpm` | `bool` | `true` (GCP default) | Enable Virtual Trusted Platform Module. |
| `shieldedInstanceConfig.enableIntegrityMonitoring` | `bool` | `true` (GCP default) | Enable integrity monitoring. |

## Examples

### Basic CPU Notebook

A minimal notebook for data exploration and light ML work.

```yaml
apiVersion: gcp.openmcf.org/v1
kind: GcpVertexAiNotebook
metadata:
  name: data-explorer
spec:
  projectId:
    value: my-gcp-project
  location: us-central1-a
  machineType: e2-standard-4
  bootDisk:
    diskType: PD_SSD
    diskSizeGb: 200
```

### GPU Notebook with TensorFlow

A GPU-equipped notebook for training deep learning models.

```yaml
apiVersion: gcp.openmcf.org/v1
kind: GcpVertexAiNotebook
metadata:
  name: ml-training
spec:
  projectId:
    value: my-gcp-project
  location: us-central1-a
  machineType: n1-standard-8
  acceleratorConfig:
    type: NVIDIA_TESLA_T4
    coreCount: 1
  bootDisk:
    diskType: PD_SSD
    diskSizeGb: 200
  dataDisk:
    diskType: PD_SSD
    diskSizeGb: 500
  vmImage:
    project: deeplearning-platform-release
    family: tf-latest-gpu
```

### Private Encrypted Notebook with Foreign Key References

A security-hardened notebook inside a VPC with CMEK encryption, using `valueFrom` references for infra chart composition.

```yaml
apiVersion: gcp.openmcf.org/v1
kind: GcpVertexAiNotebook
metadata:
  name: secure-notebook
spec:
  projectId:
    valueFrom:
      kind: GcpProject
      name: ml-project
  location: us-central1-a
  machineType: e2-standard-4
  disablePublicIp: true
  networkInterface:
    network:
      valueFrom:
        kind: GcpVpc
        name: ml-vpc
    subnet:
      valueFrom:
        kind: GcpSubnetwork
        name: ml-subnet
  serviceAccount:
    valueFrom:
      kind: GcpServiceAccount
      name: notebook-sa
  bootDisk:
    diskType: PD_SSD
    diskSizeGb: 200
    kmsKey:
      valueFrom:
        kind: GcpKmsKey
        name: disk-key
  dataDisk:
    diskType: PD_BALANCED
    diskSizeGb: 500
    kmsKey:
      valueFrom:
        kind: GcpKmsKey
        name: disk-key
  shieldedInstanceConfig:
    enableSecureBoot: true
    enableVtpm: true
    enableIntegrityMonitoring: true
  tags:
    - notebook
    - no-public-ip
```

## Stack Outputs

| Output | Type | Description |
|--------|------|-------------|
| `instance_id` | `string` | Fully qualified instance ID: `projects/{project}/locations/{location}/instances/{id}` |
| `instance_name` | `string` | Short instance name (matches `instanceName` or `metadata.name`) |
| `proxy_uri` | `string` | JupyterLab proxy URL. Empty if `disableProxyAccess` is `true`. |
| `state` | `string` | Current instance state: `ACTIVE`, `STOPPED`, `INITIALIZING`, `STARTING`, `STOPPING`, etc. |
| `creator` | `string` | Email address of the entity that created the instance |
| `create_time` | `string` | RFC3339 timestamp of instance creation |

## Related Components

- [GcpProject](/docs/catalog/gcp/gcpproject) — project where the notebook is created
- [GcpVpc](/docs/catalog/gcp/gcpvpc) — VPC network for private notebook deployments
- [GcpSubnetwork](/docs/catalog/gcp/gcpsubnetwork) — subnet for VPC-connected notebooks
- [GcpServiceAccount](/docs/catalog/gcp/gcpserviceaccount) — VM identity for accessing GCP resources
- [GcpKmsKey](/docs/catalog/gcp/gcpkmskey) — encryption key for CMEK-encrypted disks
