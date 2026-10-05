# GCP Vertex AI Notebook

Deploys a Vertex AI Workbench instance -- a managed JupyterLab environment backed by a Compute Engine VM -- with configurable machine types, GPU accelerators, boot and data disk encryption, VPC networking, and pre-built or custom container images. Machine, disk, and network settings are create-time decisions; the instance can be suspended via `desiredState` to stop compute billing while disks persist.

## What Gets Created

When you deploy this Infra Component, the IaC module provisions:

- **Workbench Instance** -- a managed `workbench.Instance` in the specified GCP project and zone, configured with the chosen machine type, disk layout, and notebook environment (VM image or container image)
- **Boot Disk** -- configurable disk type (PD_SSD by default), size (150 GB default), and optional CMEK encryption via Cloud KMS
- **Data Disk** -- a single attached data disk with configurable type (PD_STANDARD by default), size (100 GB default), and optional CMEK encryption
- **GPU Accelerator** -- created only when `acceleratorConfig` is set; attaches an NVIDIA GPU (T4, V100, A100, L4, or other supported types) for ML training workloads
- **Network Interface** -- created only when `networkInterface` is set; places the instance in a specific VPC and subnet with configurable NIC type (VIRTIO_NET or GVNIC)
- **Service Account Binding** -- created only when `serviceAccount` is set; configures the VM identity for accessing GCP resources (BigQuery, GCS, Vertex AI) with `cloud-platform` scope
- **Shielded VM Configuration** -- created only when `shieldedInstanceConfig` is set; enables Secure Boot, vTPM, and integrity monitoring for enhanced security
- **GCP Labels** -- resource metadata labels (resource name, kind, organization, environment) applied automatically for tracking and governance
- **API enablement** -- `notebooks.googleapis.com` and `compute.googleapis.com` are enabled in the target project; tearing down the instance never disables the APIs

## Before You Deploy

### Planton Setup

- **GCP Provider Connection** -- an active connection in the Connect module with credentials for the target GCP project. Map it as the default for your environment, or specify it explicitly when creating the Infra Component.
- **Planton Runner** -- required when using Runner-based credential delivery. Not needed for inline credentials or browser OAuth authentication modes.

### GCP Project

- **A GCP project** where the notebook instance will be created. Provide the project ID directly or reference a GcpProject Infra Component via ValueFromRef. The module enables the Notebooks and Compute Engine APIs itself, so the connection's principal needs permission to enable services on a fresh project.
- **A VPC network and subnet** (if using private networking) -- the instance can be placed in a specific VPC/subnet with optional public IP disabled. Provide self-links directly or reference GcpVpcNetwork and GcpSubnetwork Infra Components via ValueFromRef.
- **Cloud KMS key** (if using CMEK) -- a key in the same region as the instance for boot and/or data disk encryption.
- **GPU quota** (if using accelerators) -- sufficient GPU quota in the target zone for the chosen accelerator type.

## Deploy

### Console

Open the deployment store, find **GCP Vertex AI Notebook**, and click **Deploy**. The creation wizard walks you through preset selection, environment and connection configuration, and spec fields. Start from the **Basic Notebook** preset in the [Presets](#presets) tab to pre-populate a CPU-only development notebook.

### CLI

Create a manifest and apply it:

```yaml
apiVersion: gcp.planton.dev/v1alpha1
kind: GcpVertexAiNotebook
metadata:
  name: data-science-nb
  org: acme-corp
  env: dev
spec:
  projectId:
    value: "acme-dev-12345"
  location: us-central1-a
  machineType: e2-standard-4
```

```shell
planton apply -f vertex-ai-notebook.yaml
```

This creates a Workbench instance with the default deep learning VM image, 150 GB SSD boot disk, 100 GB data disk, no GPU, and public IP access via the Vertex AI proxy. An Infra Job tracks the provisioning in real time.

### InfraChart

When deploying as part of a multi-resource environment, use ValueFromRef to wire the notebook to a GCP project, VPC, subnet, and service account deployed in the same InfraPipeline:

```yaml
spec:
  projectId:
    valueFrom:
      kind: GcpProject
      name: ml-project
      fieldPath: status.outputs.project_id
  networkInterface:
    network:
      valueFrom:
        kind: GcpVpcNetwork
        name: ml-vpc
        fieldPath: status.outputs.network_self_link
    subnet:
      valueFrom:
        kind: GcpSubnetwork
        name: ml-subnet
        fieldPath: status.outputs.subnetwork_self_link
  serviceAccount:
    valueFrom:
      kind: GcpServiceAccount
      name: ml-workbench-sa
      fieldPath: status.outputs.email
```

The InfraPipeline resolves the dependency graph, deploys the project, VPC, subnet, and service account first, then provisions the notebook instance with private networking and custom VM identity.

## Key Configuration

These are the most important decisions when configuring a Vertex AI Notebook. Explore the full field reference in the [API Explorer](#api-explorer) tab.

**Machine type and GPU** -- Choose `machineType` based on workload requirements. Use `e2-standard-4` for data exploration, `n1-standard-8` with an NVIDIA T4 for ML training, or `a2-highgpu-1g` for large model training. GPU accelerators require compatible N1 or A2 machine types -- check GCP documentation for the compatibility matrix.

**Notebook environment** -- Choose between `vmImage` (pre-built deep learning VM images with TensorFlow, PyTorch, or JAX) and `containerImage` (custom container with your own libraries). Only one can be set. If neither is specified, GCP uses the default deep learning VM image. VM images are the most common choice.

**Disk configuration** -- `bootDisk` controls the OS disk (default 150 GB PD_SSD). `dataDisk` stores notebooks and datasets (default 100 GB PD_STANDARD). Both support CMEK encryption via `kmsKey`. Disk settings are immutable after creation -- plan capacity upfront.

**Network access** -- Set `disablePublicIp: true` to restrict the instance to private VPC access only. The notebook remains accessible through the Vertex AI proxy URL or via VPN/Cloud IAP tunnel. Combine with `networkInterface` to place the instance in a specific VPC and subnet. Network settings are immutable after creation.

**Instance state** -- Set `desiredState` to `STOPPED` to suspend the instance and stop compute billing while preserving disks and configuration. Storage charges continue. Set to `ACTIVE` (default) to keep the instance running.

## Outputs and Dependencies

### What This Kind Consumes

| Dependency | Field | ValueFromRef Path |
|------------|-------|-------------------|
| **GcpProject** (optional) | `projectId` | `status.outputs.project_id` |
| **GcpVpcNetwork** (optional) | `networkInterface.network` | `status.outputs.network_self_link` |
| **GcpSubnetwork** (optional) | `networkInterface.subnet` | `status.outputs.subnetwork_self_link` |
| **GcpAddress** (optional) | `networkInterface.externalIp` | `status.outputs.address` |
| **GcpKmsKey** (optional) | `bootDisk.kmsKey` | `status.outputs.key_id` |
| **GcpKmsKey** (optional) | `dataDisk.kmsKey` | `status.outputs.key_id` |
| **GcpServiceAccount** (optional) | `serviceAccount` | `status.outputs.email` |

### What This Kind Provides

After provisioning, `status.outputs` contains values that downstream Infra Components can consume via ValueFromRef:

| Output | Description | Common Downstream Use |
|--------|-------------|----------------------|
| `instance_id` | Fully qualified instance ID (`projects/{project}/locations/{location}/instances/{id}`) | Monitoring dashboards, IAM bindings |
| `instance_name` | Short instance name | Inventory tracking, script references |
| `proxy_uri` | JupyterLab proxy URL (empty if proxy access is disabled) | User access to notebooks |
| `state` | Current instance state (ACTIVE, STOPPED, etc.) | Monitoring, lifecycle automation |
| `creator` | Email of the entity that created the instance | Audit logs, cost attribution |
| `create_time` | RFC3339 timestamp of instance creation | Lifecycle tracking |
| `health_state` | Guest-agent health (HEALTHY, AGENT_NOT_RUNNING, etc.) | The first check when JupyterLab stops responding on an ACTIVE VM |
| `update_time` | RFC3339 timestamp of the last instance update | Change tracking, drift investigation |

## Common Patterns

Browse the [Presets](#presets) tab for ready-to-deploy configurations.

**Basic CPU notebook** -- A CPU-only Workbench instance for data exploration and light ML workflows. Uses `e2-standard-4` with default disks and public access via the Vertex AI proxy. Start from the **Basic Notebook** preset.

**GPU ML notebook** -- An NVIDIA GPU-equipped instance for ML model training. Uses a compatible N1 machine type with a Tesla T4 or A100 accelerator and SSD disks for fast data loading. Start from the **GPU ML Notebook** preset.

**Secure private notebook** -- Enterprise-grade configuration with CMEK encryption on both disks, private VPC networking with no public IP, Shielded VM enabled, and a dedicated service account. Start from the **Secure Private Notebook** preset.

## Works With

- [**GCP Project**](/infra-catalog/gcp-project) -- provides the GCP project where the notebook instance is created
- [**GCP VPC Network**](/infra-catalog/gcp-vpc-network) -- provides the VPC network for private notebook access
- [**GCP Subnetwork**](/infra-catalog/gcp-subnetwork) -- provides the subnet for instance placement within the VPC
- [**GCP KMS Key**](/infra-catalog/gcp-kms-key) -- provides customer-managed encryption keys for boot and data disks
- [**GCP Service Account**](/infra-catalog/gcp-service-account) -- provides the VM identity for accessing GCP resources