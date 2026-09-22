# GcpVertexAiRagEngineConfig

> Generated from the protobuf schema by `make generate-reference` -- do not
> edit by hand. To change a fact on this page, change the proto field comment
> or validation rule it is derived from, then regenerate.

**apiVersion**: `gcp.planton.dev/v1alpha1`

**Guide**: [GUIDE.md](../GUIDE.md) -- authored operational judgment for this component: conventions, trade-offs, and what pairs well with it.

GcpVertexAiRagEngineConfigSpec sets the tier of Vertex AI RAG Engine's
managed vector database (`google_vertex_ai_rag_engine_config`) for one
project and location. RAG Engine keeps every RAG corpus that uses the
managed database (RagManagedDb) in a per-location store; this
configuration is the one switch that decides how that store is
provisioned, and it is a singleton -- there is exactly one per project
per location, and Google creates it on first use.

Because the resource is a singleton Google already owns, "creating" it
is an update: applying this manifest over a location that is already
configured changes the tier in place rather than failing. The tiers:

  BASIC          -- a cost-effective, low-compute tier for experiments,
                    small corpora, latency-insensitive workloads, or when
                    RAG Engine is used only with external vector databases
                    (Vector Search, Vertex AI Feature Store, Pinecone,
                    Weaviate).
  SCALED         -- production-grade performance with autoscaling, for
                    large corpora or latency-sensitive workloads.
  UNPROVISIONED  -- disables the managed database in this location and
                    DELETES ALL DATA RAG Engine holds for it, halting its
                    billing. Moving to UNPROVISIONED is irreversible for
                    the data.

Destroying this block also PATCHes the location to UNPROVISIONED, which
deletes the managed database's data. A team that wants to stop managing
the tier from Planton while keeping its corpora sets deletion_policy to
ABANDON. Only the location and project are immutable.

## Example

```yaml
apiVersion: gcp.planton.dev/v1alpha1
kind: GcpVertexAiRagEngineConfig
metadata:
  name: rag-engine-us-central1
spec:
  projectId:
    value: my-gcp-project
  location: us-central1
  # BASIC is the experimentation tier; SCALED is production grade with
  # autoscaling; UNPROVISIONED disables the managed database AND deletes
  # its data. Destroying this block also unprovisions -- ABANDON keeps
  # the data.
  tier: BASIC
  deletionPolicy: DELETE
```

## Spec Fields

| Path | Type | Required | Default | References |
|---|---|---|---|---|
| `spec.projectId` | `string \| valueFrom` |  |  | GcpProject (`status.outputs.project_id`) |
| `spec.location` | `string` | yes |  |  |
| `spec.tier` | `string` | yes |  |  |
| `spec.deletionPolicy` | `string` |  |  |  |

## Field Details

### spec.projectId

`string | valueFrom`

The GCP project whose RAG Engine is configured: a literal project ID or
a GcpProject reference. If omitted, the provider's default project is
used.

- references: GcpProject (`status.outputs.project_id`)
- rule: write as {value: <literal>} or {valueFrom: {kind: GcpProject, name: <that resource's name>, fieldPath: status.outputs.project_id}} -- a bare string does not parse

### spec.location

`string` · required

The Vertex AI location (region) whose RAG Engine is configured, e.g.
"us-central1". One configuration exists per project per location.
Immutable.

- rule: {"required":true,"string":{"pattern":"^[a-z]+-[a-z]+[0-9]+$"}}

### spec.tier

`string` · required

The managed database tier for this location: BASIC, SCALED, or
UNPROVISIONED (see the message comment; UNPROVISIONED deletes the
managed database's data). Mutable in place -- raising BASIC to SCALED
is a live upgrade.

- rule: {"required":true,"string":{"in":["BASIC","SCALED","UNPROVISIONED"]}}

### spec.deletionPolicy

`string`

What happens to the configuration when this resource is destroyed:
  "" / "DELETE" -- the location is set to UNPROVISIONED, which deletes
                   the managed database's data (Google's delete
                   semantics for this singleton)
  "PREVENT"     -- destroy fails
  "ABANDON"     -- the configuration leaves management and the
                   location keeps its tier and data

- rule: deletion_policy must be one of: DELETE, PREVENT, ABANDON

## Outputs

Reference an output from another manifest as `valueFrom: {kind: GcpVertexAiRagEngineConfig, name: <resource-name>, fieldPath: status.outputs.<output>}`.

| Output | Type | Description |
|---|---|---|
| `status.outputs.name` | `string` | Full resource name of the singleton configuration: projects/{project}/locations/{location}/ragEngineConfig. |
| `status.outputs.location` | `string` | The location the configuration governs -- with the project, what a verifier or a downstream block rebuilds the resource path from. |

## References

Fields that can point at another resource's outputs:

| Field | Kind | Output |
|---|---|---|
| `spec.projectId` | GcpProject | `status.outputs.project_id` |

## See Also

- [Overview](../README.md)
