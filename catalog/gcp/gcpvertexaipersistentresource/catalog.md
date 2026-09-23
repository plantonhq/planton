# GCP Vertex AI Persistent Resource

Keeps a pool of training machines warm on Vertex AI so custom training jobs and Ray on Vertex AI start in seconds instead of waiting minutes for capacity -- and so the GPUs you fought for stay yours between jobs. Declare the machine pools, optionally peer them into your VPC, and point training jobs at the resource by id. The machines bill from the moment they are provisioned until you delete the resource, which is exactly the trade: pay for readiness, get instant starts and held accelerators.

## What Gets Created

When you deploy this Cloud Resource, the IaC module provisions:

- **API enablement** -- `aiplatform.googleapis.com` on the project
- **Persistent resource** -- a `vertex.AiPersistentResource` with its machine pools, networking, and optional CMEK

## Before You Deploy

### Planton Setup

- **GCP Provider Connection** -- an active connection in the Connect module with Vertex AI admin permissions on the project. Map it as the default for your environment, or specify it explicitly when creating the Cloud Resource.
- **Planton Runner** -- required when using Runner-based credential delivery. Not needed for inline credentials or browser OAuth authentication modes.

### Optional Dependencies

- **GcpVpcNetwork** -- a network with private services access for Vertex AI, referenced by `network`.
- **GcpKmsKey** -- for customer-managed encryption, referenced by `kmsKeyName`.

## Deploy

### Console

Open the deployment store, find **GCP Vertex AI Persistent Resource**, and click **Deploy**. The creation wizard walks you through preset selection, environment and connection configuration, and spec fields. Start from the **CPU Training Pool** preset in the [Presets](#presets) tab.

### CLI

Create a manifest and apply it:

```yaml
apiVersion: gcp.planton.dev/v1alpha1
kind: GcpVertexAiPersistentResource
metadata:
  name: training-pool
  org: acme-corp
  env: prod
spec:
  location: us-central1
  resourcePools:
    - machineSpec:
        machineType: n1-standard-4
      replicaCount: 1
```

```shell
planton apply -f vertex-ai-persistent-resource.yaml
```

This keeps one `n1-standard-4` machine provisioned for training jobs in `us-central1`. A Stack Job tracks the provisioning in real time.

### InfraChart

When deploying as part of a multi-resource environment, pass the resource's `persistent_resource_id` output to the training jobs or pipelines that run on it, and reference a `GcpVpcNetwork` from `network` when jobs must reach private services.

## Key Configuration

These are the most important decisions when configuring a persistent resource. Explore the full field reference in the [API Explorer](#api-explorer) tab.

**Pools are the bill** -- every replica of every pool runs around the clock. Size `replicaCount` to what jobs need at once; `autoscalingSpec` lets a pool grow with the jobs on it, from a floor of at least one.

**Networking must match the jobs** -- a job on the resource must use the same network and key. Peer with `network` (private services access) or attach through `pscInterfaceConfig` (Private Service Connect interface); decide before creation, because both are permanent.

**Runtime identity** -- `enableCustomServiceAccount: true` requires every job to run as a user-managed service account instead of the Vertex AI Custom Code Service Agent.

## Outputs and Dependencies

### What This Component Consumes

| Dependency | Field | ValueFromRef Path |
|------------|-------|-------------------|
| **GcpProject** | `projectId` | `status.outputs.project_id` |
| **GcpVpcNetwork** | `network` | `status.outputs.network_self_link` |
| **GcpProject** | `pscInterfaceConfig.dnsPeeringConfigs[].targetProject` | `status.outputs.project_id` |
| **GcpVpcNetwork** | `pscInterfaceConfig.dnsPeeringConfigs[].targetNetwork` | `status.outputs.network_name` |
| **GcpKmsKey** | `kmsKeyName` | `status.outputs.key_id` |

### What This Component Provides

After provisioning, `status.outputs` contains values that downstream Cloud Resources can consume via ValueFromRef:

| Output | Description | Common Downstream Use |
|--------|-------------|----------------------|
| `name` | The resource's full name | SDK calls |
| `persistent_resource_id` | The resource's id | A training job's `persistent_resource_id` |
| `location` | The resource's region | Regional clients |
| `state` | The resource's state | Readiness checks |

## Common Patterns

Browse the [Presets](#presets) tab for ready-to-deploy configurations.

**CPU training pool** -- a small fixed pool for fast-starting CPU jobs. Start from the **CPU Training Pool** preset.

**Private GPU pool** -- an autoscaling L4 pool peered into your VPC, with custom service accounts, CMEK, and `PREVENT`. Start from the **Private GPU Pool** preset.

## Works With

- [**GCP VPC Network**](/cloud-catalog/gcp-vpc-network) -- the network the resource peers with
- [**GCP KMS Key**](/cloud-catalog/gcp-kms-key) -- customer-managed encryption
- [**GCP Vertex AI TensorBoard**](/cloud-catalog/gcp-vertex-ai-tensorboard) -- where jobs on the resource stream metrics
- [**GCP Vertex AI Dataset**](/cloud-catalog/gcp-vertex-ai-dataset) -- the data those jobs train on
