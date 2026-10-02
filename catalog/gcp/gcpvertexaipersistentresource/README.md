# GCP Vertex AI Persistent Resource

A Vertex AI persistent resource -- a long-running cluster of machines Vertex AI keeps provisioned so custom training jobs (and Ray on Vertex AI) start in seconds instead of waiting minutes for capacity, and so scarce accelerators stay held between jobs. Declare the pools of machines, optionally peer them into your VPC or give them a Private Service Connect interface, and have training jobs name the resource's id in `persistent_resource_id`.

## What Gets Created

When you deploy this Cloud Resource, the IaC module provisions:

- **API enablement** -- `aiplatform.googleapis.com` on the project (never disabled on destroy)
- **Persistent resource** -- a `vertex_ai_persistent_resource` with its resource pools, networking, runtime identity rule, and optional CMEK

## Before You Deploy

### Planton Setup

- **GCP Provider Connection** -- an active connection in the Connect module with Vertex AI admin permissions on the project.
- **Planton Runner** -- required when using Runner-based credential delivery. Not needed for inline credentials or browser OAuth authentication modes.

### Quota

- Custom-training quota in the region for the machine types and accelerators the pools use; Google allocates every replica before the resource reaches `RUNNING`.

### Optional Dependencies

- **`GcpVpcNetwork`** -- a network with VPC Network Peering for Vertex AI (private services access) configured, for `network`. A `GcpServiceNetworkingConnection` sets that up; `reservedIpRanges` names its allocated ranges.
- **`GcpKmsKey`** -- a key in the same region for customer-managed encryption (`kmsKeyName`).

## Deploy

### CLI

```yaml
apiVersion: gcp.planton.dev/v1alpha1
kind: GcpVertexAiPersistentResource
metadata:
  name: training-pool
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

## Configuration Reference

### Required Fields

| Field | Type | Description |
|-------|------|-------------|
| `location` | `string` | The Vertex AI region. Immutable. |
| `resourcePools[]` | `[]object` | At least one pool with a `machineSpec`. |

### Optional Fields

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `projectId` | `StringValueOrRef` | provider project | The project, as a literal or a `GcpProject` reference. |
| `persistentResourceId` | `string` | `metadata.name` | RFC 1035, up to 63 characters. Immutable. |
| `displayName`, `labels` | | | Descriptive metadata; mutable. |
| `resourcePools[].id` | `string` | Google-generated | The pool id a job's worker pool refers to. Immutable. |
| `resourcePools[].machineSpec` | `object` | -- | `machineType`, `acceleratorType` (Google's list), `acceleratorCount`. Immutable. |
| `resourcePools[].replicaCount` | `int64` | -- | Replicas kept running; mutable. |
| `resourcePools[].autoscalingSpec` | `object` | fixed | `minReplicaCount` (at least 1), `maxReplicaCount`. Immutable. |
| `resourcePools[].diskSpec` | `object` | Google's defaults | `bootDiskSizeGb` (default 100), `bootDiskType` (`pd-ssd`, `pd-standard`, `hyperdisk-balanced`). Immutable. |
| `network` | `StringValueOrRef` | no peering | A `GcpVpcNetwork` reference or literal `projects/{project}/global/networks/{name}`; the modules resolve the project number Google requires. Immutable. |
| `reservedIpRanges` | `[]string` | any range | Names of the peering's allocated ranges. Immutable. |
| `pscInterfaceConfig` | `object` | none | `networkAttachment`, `dnsPeeringConfigs[] { domain, targetProject, targetNetwork }`. Immutable. |
| `enableCustomServiceAccount` | `bool` | `false` | Require jobs to run as a user-managed service account. Immutable. |
| `kmsKeyName` | `StringValueOrRef` | Google-managed | A `GcpKmsKey` reference or literal key path. Immutable. |
| `deletionPolicy` | `string` | `DELETE` | `DELETE`, `PREVENT`, or `ABANDON`. |

### Validation Rules

- At least one pool, each with a machine spec; replica counts and the autoscaling floor are at least 1 (Google rejects a zero floor on a persistent resource).
- Accelerator types and boot disk types are Google's lists.
- DNS peering domains end with a dot.

## Stack Outputs

| Output | Type | Description |
|--------|------|-------------|
| `name` | `string` | `projects/{project}/locations/{location}/persistentResources/{persistent_resource_id}` |
| `persistent_resource_id` | `string` | What a training job's `persistent_resource_id` takes |
| `location` | `string` | The resource's location |
| `state` | `string` | The resource's state (`RUNNING` once every pool is provisioned) |

## Deployment Methods

### Pulumi (Go)

See [`iac/pulumi/README.md`](iac/pulumi/README.md) for Pulumi-specific deployment instructions.

### Terraform

See [`iac/tf/README.md`](iac/tf/README.md) for Terraform-specific deployment instructions.

## Important Notes

- **It bills from provisioning until deletion.** Every replica runs, and bills, whether or not a job is using it -- that is what makes jobs start instantly.
- **Almost everything is permanent.** Only the display name, labels, and each pool's `replicaCount` update in place; any other change replaces the resource and reprovisions every machine.
- **Jobs must match it.** A job that runs on the resource must use the same network and encryption key, or Google rejects the job.

## Examples

For a complete example, see `e2e/manifest.yaml`. Scenario variants live under `e2e/scenarios/`.

## Related Components

- **GcpVpcNetwork** -- the network the resource peers with
- **GcpServiceNetworkingConnection** -- the private services access peering the network needs
- **GcpKmsKey** -- customer-managed encryption for the machines' disks
- **GcpVertexAiTensorboard** -- where jobs on the resource stream their metrics

---

© Planton. Licensed under [Apache-2.0](https://github.com/plantonhq/planton/blob/main/LICENSE).
