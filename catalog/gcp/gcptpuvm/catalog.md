# GCP TPU VM

Gives your team Google's AI chips. A Cloud TPU VM is a slice of Tensor Processing Units -- from a single 8-chip host to multi-host slices -- with its own VMs, ready for training or serving large models with JAX, PyTorch/XLA, or TensorFlow. Declare the chip generation, the size, and whether to use cheap spot capacity, and the slice comes up on your network with your identity.

## What Gets Created

When you deploy this Cloud Resource, the IaC module provisions:

- **API enablement** -- `tpu.googleapis.com` on the project
- **TPU VM** -- a `tpu.V2Vm` (Google's beta-only TPU resource, used under the catalog's recorded admission)

## Before You Deploy

### Planton Setup

- **GCP Provider Connection** -- an active connection in the Connect module with Cloud TPU admin permissions on the project, and TPU quota for the generation in the zone. Map it as the default for your environment, or specify it explicitly when creating the Cloud Resource.
- **Planton Runner** -- required when using Runner-based credential delivery. Not needed for inline credentials or browser OAuth authentication modes.

## Deploy

### Console

Open the deployment store, find **GCP TPU VM**, and click **Deploy**. The creation wizard walks you through preset selection, environment and connection configuration, and spec fields. Start from the **Spot v5e Fine-Tuning** preset in the [Presets](#presets) tab.

### CLI

Create a manifest and apply it:

```yaml
apiVersion: gcp.planton.dev/v1alpha1
kind: GcpTpuVm
metadata:
  name: finetune-v5e
  org: acme-corp
  env: prod
spec:
  zone: us-west4-a
  acceleratorType: v5litepod-8
  runtimeVersion: v2-alpha-tpuv5-lite
  schedulingConfig:
    spot: true
```

```shell
planton apply -f tpu-vm.yaml
```

This brings up an 8-chip TPU v5e slice on spot capacity. A Stack Job tracks the provisioning in real time.

### InfraChart

Reference a `GcpSubnetwork` for private networking, a `GcpServiceAccount` for the TPU's identity, and `GcpComputeDisk` resources for shared training data.

## Key Configuration

These are the most important decisions when configuring a TPU. Explore the full field reference in the [API Explorer](#api-explorer) tab.

**Generation and size** -- `acceleratorType` (e.g. `v5litepod-8`, `v6e-16`) or `acceleratorConfig` (generation plus topology). Match `runtimeVersion` to the generation.

**Where** -- `zone`: each generation exists in only a few zones.

**Capacity model** -- on-demand, `spot` (much cheaper, reclaimable any time), or `reserved`. When on-demand capacity is short, use a `GcpTpuQueuedResource` instead.

**Network and identity** -- the subnetwork, whether workers get external IPs, and the service account the hosts run as.

## Outputs and Dependencies

### What This Component Consumes

| Dependency | Field | ValueFromRef Path |
|------------|-------|-------------------|
| **GcpProject** | `projectId` | `status.outputs.project_id` |
| **GcpVpcNetwork** | `networkConfig.network`, `networkConfigs[].network` | `status.outputs.network_id` |
| **GcpSubnetwork** | `networkConfig.subnetwork`, `networkConfigs[].subnetwork` | `status.outputs.subnetwork_self_link` |
| **GcpServiceAccount** | `serviceAccount.email` | `status.outputs.email` |
| **GcpComputeDisk** | `dataDisks[].sourceDisk` | `status.outputs.self_link` |

### What This Component Provides

After provisioning, `status.outputs` contains values that downstream Cloud Resources can consume via ValueFromRef:

| Output | Description | Common Downstream Use |
|--------|-------------|----------------------|
| `name` | The TPU's full resource name | Automation, monitoring |
| `node_id` | The TPU's id | `gcloud compute tpus tpu-vm ssh` |
| `zone` | The TPU's zone | Regional tooling |

## Common Patterns

Browse the [Presets](#presets) tab for ready-to-deploy configurations.

**Spot v5e fine-tuning** -- an 8-chip v5e slice on spot capacity with a startup script, for cost-sensitive fine-tuning. Start from the **Spot v5e Fine-Tuning** preset.

**Private reserved v6e** -- a Trillium slice from a reservation on a private subnetwork with its own identity, no external IPs, Secure Boot, and a shared read-only dataset disk. Start from the **Private Reserved v6e** preset.

## Works With

- [**GCP TPU Queued Resource**](/cloud-catalog/gcp-tpu-queued-resource) -- wait for scarce capacity
- [**GCP Subnetwork**](/cloud-catalog/gcp-subnetwork) -- private networking
- [**GCP Compute Disk**](/cloud-catalog/gcp-compute-disk) -- shared data and checkpoints
- [**GCP Service Account**](/cloud-catalog/gcp-service-account) -- the TPU's identity
