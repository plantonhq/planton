# GCP TPU VM

A Cloud TPU VM -- a slice of Google's AI accelerators (v2 through v6e / Trillium) with its host VMs, for training and serving large models with JAX, PyTorch/XLA, or TensorFlow. The accelerator type or topology decides the slice size; larger slices span several hosts, each with its own VM and network endpoint. Google publishes this resource only in its beta Terraform provider, so the module uses `google-beta` for it under the catalog's recorded admission.

## What Gets Created

When you deploy this Cloud Resource, the IaC module provisions:

- **API enablement** -- `tpu.googleapis.com` on the project (never disabled on destroy)
- **TPU VM** -- a `tpu_v2_vm` (google-beta) with the chosen accelerator, runtime, network, identity, scheduling, disks, and labels

## Before You Deploy

### Planton Setup

- **GCP Provider Connection** -- an active connection in the Connect module with Cloud TPU admin permissions on the project, and TPU quota for the generation in the chosen zone.
- **Planton Runner** -- required when using Runner-based credential delivery. Not needed for inline credentials or browser OAuth authentication modes.

### Optional Dependencies

- **`GcpVpcNetwork`** / **`GcpSubnetwork`** -- the network (`networkConfig` or `networkConfigs`).
- **`GcpServiceAccount`** -- the host VMs' identity (`serviceAccount.email`).
- **`GcpComputeDisk`** -- data disks to attach (`dataDisks[].sourceDisk`).

## Deploy

### CLI

```yaml
apiVersion: gcp.planton.dev/v1alpha1
kind: GcpTpuVm
metadata:
  name: train-v5e
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

## Configuration Reference

### Required Fields

| Field | Type | Description |
|-------|------|-------------|
| `zone` | `string` | A zone that offers the TPU generation, e.g. `us-central1-a`. Immutable. |
| `runtimeVersion` | `string` | The TPU software image, e.g. `tpu-ubuntu2204-base`, `v2-alpha-tpuv5-lite`. Immutable. |

### Optional Fields

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `projectId` | `StringValueOrRef` | provider project | The project, as a literal or a `GcpProject` reference. |
| `nodeId` | `string` | `metadata.name` | The TPU's name. Immutable. |
| `acceleratorType` | `string` | `v2-8` | The slice by name (`v2-8`, `v5litepod-8`, `v6e-8`, ...). Immutable. |
| `acceleratorConfig` | `object` | none | The slice by `type` (`V2` ... `V6E`) and `topology` (`2x2`, `2x2x1`). Immutable. |
| `description` | `string` | none | Mutable. |
| `cidrBlock` | `string` | Google chooses | A /29 block. Immutable. |
| `networkConfig` / `networkConfigs` | `object` / `object[]` | default network | One interface or several: `network`, `subnetwork`, `enableExternalIps`, `canIpForward`, `queueCount`. Immutable. |
| `serviceAccount` | `object` | Compute default | `email` (`GcpServiceAccount` ref), `scopes`. Immutable. |
| `schedulingConfig` | `object` | on-demand | `spot`, `preemptible`, `reserved`. Immutable. |
| `dataDisks` | `object[]` | none | `sourceDisk` (`GcpComputeDisk` ref), `mode` (`READ_WRITE` / `READ_ONLY`). Mutable. |
| `enableSecureBoot` | `bool` | `false` | Shielded VM Secure Boot. Immutable. |
| `labels` / `metadata` / `tags` | map / map / list | none | Mutable; `metadata` carries `startup-script`. |
| `deletionPolicy` | `string` | `DELETE` | `DELETE`, `PREVENT`, or `ABANDON`. |

### Validation Rules

- At most one of `acceleratorType` and `acceleratorConfig`; at most one of `networkConfig` and `networkConfigs`.
- `cidrBlock` is an IPv4 /29; `zone` is a zone.

## Stack Outputs

| Output | Type | Description |
|--------|------|-------------|
| `name` | `string` | `projects/{project}/locations/{zone}/nodes/{node_id}` |
| `node_id` | `string` | The TPU's id (for `gcloud compute tpus tpu-vm ssh`) |
| `zone` | `string` | The TPU's zone |

## Deployment Methods

### Pulumi (Go)

See [`iac/pulumi/README.md`](iac/pulumi/README.md) for Pulumi-specific deployment instructions.

### Terraform

See [`iac/tf/README.md`](iac/tf/README.md) for Terraform-specific deployment instructions.

## Important Notes

- **Capacity is the constraint.** TPUs live in specific zones, need quota, and are often scarce. If a create fails for lack of capacity, request the slice through a `GcpTpuQueuedResource`, which waits for it.
- **Spot is far cheaper and can be reclaimed at any time.** Checkpoint to a data disk or Cloud Storage often.
- **Almost everything is fixed.** Only the description, labels, metadata, tags, and data disks change in place.
- **A TPU bills every hour it exists.** Destroy it when the job is done; `ABANDON` keeps it billing outside management.

## Examples

For a complete example, see `e2e/manifest.yaml`. Scenario variants live under `e2e/scenarios/`.

## Related Components

- **GcpTpuQueuedResource** -- request TPU capacity that waits in Google's queue
- **GcpVpcNetwork** / **GcpSubnetwork** -- the TPU's network
- **GcpComputeDisk** -- shared training data and checkpoints
- **GcpServiceAccount** -- the host VMs' identity

---

© Planton. Licensed under [Apache-2.0](https://github.com/plantonhq/planton/blob/main/LICENSE).
