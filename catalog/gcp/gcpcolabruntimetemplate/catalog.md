# GCP Colab Runtime Template

Gives your data scientists pre-approved notebook machines. A Colab Enterprise runtime template captures everything a notebook runtime needs -- the machine and GPU, the disk, a private network, encryption, the Colab image, and an idle shutdown that stops forgotten notebooks from running up the bill -- so every runtime your team starts from it follows the rules without anyone configuring a VM.

## What Gets Created

When you deploy this Cloud Resource, the IaC module provisions:

- **API enablement** -- `aiplatform.googleapis.com` on the project
- **Runtime template** -- a `colab.RuntimeTemplate` with the declared settings

## Before You Deploy

### Planton Setup

- **GCP Provider Connection** -- an active connection in the Connect module with Vertex AI admin permissions on the project. Map it as the default for your environment, or specify it explicitly when creating the Cloud Resource.
- **Planton Runner** -- required when using Runner-based credential delivery. Not needed for inline credentials or browser OAuth authentication modes.

### Optional Dependencies

- **GcpVpcNetwork** / **GcpSubnetwork** -- private networking for runtimes.
- **GcpKmsKey** -- customer-managed encryption for runtime disks.

## Deploy

### Console

Open the deployment store, find **GCP Colab Runtime Template**, and click **Deploy**. The creation wizard walks you through preset selection, environment and connection configuration, and spec fields. Start from the **Standard CPU** preset in the [Presets](#presets) tab.

### CLI

Create a manifest and apply it:

```yaml
apiVersion: gcp.planton.dev/v1alpha1
kind: GcpColabRuntimeTemplate
metadata:
  name: gpu-t4
  org: acme-corp
  env: prod
spec:
  location: us-central1
  displayName: GPU (T4)
  machineSpec:
    machineType: n1-standard-8
    acceleratorType: NVIDIA_TESLA_T4
    acceleratorCount: 1
  idleTimeout: 3600s
```

```shell
planton apply -f colab-runtime-template.yaml
```

This publishes a one-GPU template that shuts idle runtimes down after an hour. A Stack Job tracks the provisioning in real time.

### InfraChart

Reference the template's `name` output from a `GcpColabRuntime` or a `GcpColabSchedule` notebook run, a `GcpSubnetwork` from `networkSpec.subnetwork` for private runtimes, and a `GcpKmsKey` from `kmsKeyName`.

## Key Configuration

These are the most important decisions when configuring a template. Explore the full field reference in the [API Explorer](#api-explorer) tab.

**The machine** -- `machineSpec` picks the machine type and any GPU. It is fixed; publish a second template for a second size.

**Idle shutdown** -- `idleTimeout` stops runtimes nobody is using. It is the biggest lever on notebook spend.

**Network posture** -- `networkSpec` decides whether runtimes get a public internet path or live on your private subnetwork.

**Security** -- `eucDisabled` keeps users' own credentials out of the runtime, `enableSecureBoot` turns on Shielded VM Secure Boot, and `kmsKeyName` encrypts disks under your key.

## Outputs and Dependencies

### What This Component Consumes

| Dependency | Field | ValueFromRef Path |
|------------|-------|-------------------|
| **GcpProject** | `projectId` | `status.outputs.project_id` |
| **GcpVpcNetwork** | `networkSpec.network` | `status.outputs.network_id` |
| **GcpSubnetwork** | `networkSpec.subnetwork` | `status.outputs.subnetwork_self_link` |
| **GcpKmsKey** | `kmsKeyName` | `status.outputs.key_id` |

### What This Component Provides

After provisioning, `status.outputs` contains values that downstream Cloud Resources can consume via ValueFromRef:

| Output | Description | Common Downstream Use |
|--------|-------------|----------------------|
| `name` | The template's full resource name | `GcpColabRuntime.runtimeTemplate`, `GcpColabSchedule` notebook runs |
| `runtime_template_id` | The template's id | Console and SDK |
| `location` | The template's region | Regional clients |

## Common Patterns

Browse the [Presets](#presets) tab for ready-to-deploy configurations.

**Standard CPU** -- a mid-size CPU machine with a one-hour idle shutdown, for everyday notebooks. Start from the **Standard CPU** preset.

**Private GPU** -- a T4 GPU on your private subnetwork with no internet path, user credentials blocked, Secure Boot, and CMEK. Start from the **Private GPU** preset.

## Works With

- [**GCP Colab Runtime**](/cloud-catalog/gcp-colab-runtime) -- runtimes assigned from the template
- [**GCP Colab Schedule**](/cloud-catalog/gcp-colab-schedule) -- scheduled notebook runs
- [**GCP Subnetwork**](/cloud-catalog/gcp-subnetwork) -- private networking
- [**GCP KMS Key**](/cloud-catalog/gcp-kms-key) -- customer-managed encryption
