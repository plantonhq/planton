# GCP TPU Queued Resource

A Cloud TPU queued resource -- a request for TPU capacity that waits in Google's queue until the capacity exists, then provisions the TPU nodes it describes. It is Google's recommended way to get scarce TPU capacity: instead of failing when a zone is out of chips (as a direct `GcpTpuVm` create does), the request stays waiting and turns into running TPUs when Google can place them. One request can ask for several nodes, provisioned together; the nodes belong to the request. Google publishes this resource only in its beta Terraform provider, so the module uses `google-beta` for it under the catalog's recorded admission.

## What Gets Created

When you deploy this Cloud Resource, the IaC module provisions:

- **API enablement** -- `tpu.googleapis.com` on the project (never disabled on destroy)
- **Queued resource** -- a `tpu_v2_queued_resource` (google-beta) requesting the declared nodes

## Before You Deploy

### Planton Setup

- **GCP Provider Connection** -- an active connection in the Connect module with Cloud TPU admin permissions on the project, and TPU quota for the generation in the zone.
- **Planton Runner** -- required when using Runner-based credential delivery. Not needed for inline credentials or browser OAuth authentication modes.

### Optional Dependencies

- **`GcpVpcNetwork`** / **`GcpSubnetwork`** -- each node's network (`nodeSpecs[].node.networkConfig`).

## Deploy

### CLI

```yaml
apiVersion: gcp.planton.dev/v1alpha1
kind: GcpTpuQueuedResource
metadata:
  name: train-request
spec:
  zone: us-west4-a
  nodeSpecs:
    - nodeId: train-v5e
      node:
        runtimeVersion: v2-alpha-tpuv5-lite
        acceleratorType: v5litepod-8
```

```shell
planton apply -f tpu-queued-resource.yaml
```

## Configuration Reference

### Required Fields

| Field | Type | Description |
|-------|------|-------------|
| `zone` | `string` | The zone capacity is requested in. Immutable. |
| `nodeSpecs` | `object[]` | One entry per node: `nodeId` (optional; Google generates one), `node.runtimeVersion` (required), `node.acceleratorType` (default `v2-8`), `node.description`, `node.networkConfig`. |

### Optional Fields

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `projectId` | `StringValueOrRef` | provider project | The project, as a literal or a `GcpProject` reference. |
| `queuedResourceId` | `string` | `metadata.name` | The request's id. |
| `deletionPolicy` | `string` | `DELETE` | `DELETE`, `PREVENT`, or `ABANDON`. |

### Validation Rules

- At least one node; `nodeId` values are unique.
- Ids are lowercase letters, digits, and hyphens, starting with a letter.

## Stack Outputs

| Output | Type | Description |
|--------|------|-------------|
| `name` | `string` | `projects/{project}/locations/{zone}/queuedResources/{queued_resource_id}` |
| `queued_resource_id` | `string` | The request's id |
| `zone` | `string` | The zone |

## Deployment Methods

### Pulumi (Go)

See [`iac/pulumi/README.md`](iac/pulumi/README.md) for Pulumi-specific deployment instructions.

### Terraform

See [`iac/tf/README.md`](iac/tf/README.md) for Terraform-specific deployment instructions.

## Important Notes

- **Everything is immutable.** Any change replaces the request and the nodes it created.
- **The nodes belong to the request.** Destroying it deletes them; `ABANDON` leaves them running (and billing).
- **A waiting request is normal.** Google provisions the nodes when it has capacity; watch the request's state in the console or with `gcloud compute tpus queued-resources describe`.
- **At the pinned provider a request carries no spot, reservation, labels, service account, or disks.** Those need a `GcpTpuVm`.

## Examples

For a complete example, see `e2e/manifest.yaml`. Scenario variants live under `e2e/scenarios/`.

## Related Components

- **GcpTpuVm** -- create a TPU directly when capacity is available
- **GcpVpcNetwork** / **GcpSubnetwork** -- the nodes' network

---

© Planton. Licensed under [Apache-2.0](https://github.com/plantonhq/planton/blob/main/LICENSE).
