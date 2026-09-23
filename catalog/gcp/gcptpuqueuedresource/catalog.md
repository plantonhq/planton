# GCP TPU Queued Resource

Gets you TPU capacity even when Google is out of chips. A queued resource is a standing request: describe the TPU nodes you need and Google queues the request, then brings the nodes up as soon as capacity in the zone allows -- instead of failing like a direct create does. It is Google's recommended way to obtain scarce accelerators, and one request can bring up several nodes together.

## What Gets Created

When you deploy this Cloud Resource, the IaC module provisions:

- **API enablement** -- `tpu.googleapis.com` on the project
- **Queued resource** -- a `tpu.V2QueuedResource` (Google's beta-only TPU resource, used under the catalog's recorded admission)

## Before You Deploy

### Planton Setup

- **GCP Provider Connection** -- an active connection in the Connect module with Cloud TPU admin permissions on the project, and TPU quota in the zone. Map it as the default for your environment, or specify it explicitly when creating the Cloud Resource.
- **Planton Runner** -- required when using Runner-based credential delivery. Not needed for inline credentials or browser OAuth authentication modes.

## Deploy

### Console

Open the deployment store, find **GCP TPU Queued Resource**, and click **Deploy**. The creation wizard walks you through preset selection, environment and connection configuration, and spec fields. Start from the **Single v5e Node** preset in the [Presets](#presets) tab.

### CLI

Create a manifest and apply it:

```yaml
apiVersion: gcp.planton.dev/v1alpha1
kind: GcpTpuQueuedResource
metadata:
  name: pretraining-request
  org: acme-corp
  env: prod
spec:
  zone: us-west4-a
  nodeSpecs:
    - nodeId: pretrain-v5e
      node:
        runtimeVersion: v2-alpha-tpuv5-lite
        acceleratorType: v5litepod-16
```

```shell
planton apply -f tpu-queued-resource.yaml
```

This queues a request for a 16-chip v5e slice; the TPU appears when Google places it. A Stack Job tracks the provisioning in real time.

### InfraChart

Reference a `GcpSubnetwork` from each node's `networkConfig.subnetwork` to bring the nodes up on your private network.

## Key Configuration

These are the most important decisions when configuring a request. Explore the full field reference in the [API Explorer](#api-explorer) tab.

**What to request** -- one `nodeSpecs` entry per node, each with its generation and size (`acceleratorType`) and matching `runtimeVersion`.

**Where** -- `zone`; the request only waits for capacity there.

**Networking** -- each node's network, subnetwork, and whether its workers get external IPs.

## Outputs and Dependencies

### What This Component Consumes

| Dependency | Field | ValueFromRef Path |
|------------|-------|-------------------|
| **GcpProject** | `projectId` | `status.outputs.project_id` |
| **GcpVpcNetwork** | `nodeSpecs[].node.networkConfig.network` | `status.outputs.network_id` |
| **GcpSubnetwork** | `nodeSpecs[].node.networkConfig.subnetwork` | `status.outputs.subnetwork_self_link` |

### What This Component Provides

After provisioning, `status.outputs` contains values that downstream Cloud Resources can consume via ValueFromRef:

| Output | Description | Common Downstream Use |
|--------|-------------|----------------------|
| `name` | The request's full resource name | Monitoring its state |
| `queued_resource_id` | The request's id | `gcloud compute tpus queued-resources` |
| `zone` | The zone | Regional tooling |

## Common Patterns

Browse the [Presets](#presets) tab for ready-to-deploy configurations.

**Single v5e node** -- one 8-chip v5e node queued in a zone, the simplest way to get a slice when on-demand creates fail. Start from the **Single v5e Node** preset.

**Two private nodes** -- two v5e nodes requested together on a private subnetwork without external IPs. Start from the **Two Private Nodes** preset.

## Works With

- [**GCP TPU VM**](/cloud-catalog/gcp-tpu-vm) -- a TPU created directly when capacity is available
- [**GCP Subnetwork**](/cloud-catalog/gcp-subnetwork) -- the nodes' private network
