# GCP Vertex AI RAG Engine Config

Sets the tier of Vertex AI RAG Engine's managed vector database for one project and location. RAG Engine keeps every corpus that uses its managed database (RagManagedDb) in a per-location store; this block is the one switch that decides how that store is provisioned -- `BASIC` for experiments and small corpora, `SCALED` for production performance with autoscaling, `UNPROVISIONED` to turn the service off and delete its data. There is exactly one configuration per project per location, and Google creates it on first use, so applying this block over an already-configured location changes the tier in place.

## What Gets Created

When you deploy this Cloud Resource, the IaC module provisions:

- **API enablement** -- `aiplatform.googleapis.com` on the project (never disabled on destroy)
- **RAG Engine configuration** -- the `vertex_ai_rag_engine_config` singleton for the location, PATCHed to the declared tier

## Before You Deploy

### Planton Setup

- **GCP Provider Connection** -- an active connection in the Connect module with `roles/aiplatform.admin` on the project.
- **Planton Runner** -- required when using Runner-based credential delivery. Not needed for inline credentials or browser OAuth authentication modes.

### Understand the destroy

Destroying this block PATCHes the location to `UNPROVISIONED`, which deletes the managed database's data for every corpus in that location. Set `deletionPolicy: ABANDON` when the block should stop managing the tier without touching the data.

## Deploy

### CLI

```yaml
apiVersion: gcp.planton.dev/v1alpha1
kind: GcpVertexAiRagEngineConfig
metadata:
  name: rag-engine-us-central1
spec:
  location: us-central1
  tier: BASIC
  deletionPolicy: ABANDON
```

```shell
planton apply -f rag-engine-config.yaml
```

## Configuration Reference

### Required Fields

| Field | Type | Description |
|-------|------|-------------|
| `location` | `string` | The Vertex AI location (region) the configuration governs. One per project per location. Immutable. |
| `tier` | `string` | `BASIC`, `SCALED`, or `UNPROVISIONED`. Mutable in place; `UNPROVISIONED` deletes the managed database's data. |

### Optional Fields

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `projectId` | `StringValueOrRef` | provider project | The project, as a literal or a `GcpProject` reference. Immutable. |
| `deletionPolicy` | `string` | `DELETE` | `DELETE` unprovisions the location on destroy (data loss); `PREVENT` fails destroy; `ABANDON` keeps the tier and data. |

### Validation Rules

- `tier` is one of the three values Google offers.
- `location` is a region name (`us-central1`, `europe-west4`).

## Stack Outputs

| Output | Type | Description |
|--------|------|-------------|
| `name` | `string` | `projects/{project}/locations/{location}/ragEngineConfig` |
| `location` | `string` | The location the configuration governs |

## Deployment Methods

### Pulumi (Go)

See [`iac/pulumi/README.md`](iac/pulumi/README.md) for Pulumi-specific deployment instructions.

### Terraform

See [`iac/tf/README.md`](iac/tf/README.md) for Terraform-specific deployment instructions.

## Important Notes

- **A singleton Google owns.** Create and delete are both PATCHes; a re-run never fails with "already exists", and two blocks for the same location overwrite each other.
- **`UNPROVISIONED` deletes data.** So does a `destroy` under the default `deletionPolicy`. Use `ABANDON` or `PREVENT` for a location whose corpora matter.
- **Corpora on external vector databases are unaffected** by the tier; the Basic tier is enough for them.

## Examples

For a complete example, see `e2e/manifest.yaml`. Scenario variants live under `e2e/scenarios/`.

## Related Components

- **GcpVectorSearchCollection** -- an external vector database RAG Engine can use instead of its managed one
- **GcpVertexAiAgentEngine** -- the agent runtime that typically consumes a RAG corpus

---

© Planton. Licensed under [Apache-2.0](https://github.com/plantonhq/planton/blob/main/LICENSE).
