# GCP Vertex AI RAG Engine Config

Turns Vertex AI RAG Engine's managed vector database on for a project and location, at the tier you choose. RAG Engine ingests documents into corpora and retrieves the right passages for a model at query time; its managed database is where those corpora live unless you bring an external vector store. This block is the single switch for that database -- Basic for experiments, Scaled for production, Unprovisioned to turn it off -- and because Google creates the configuration on first use, applying it changes the tier in place rather than creating anything new.

## What Gets Created

When you deploy this Cloud Resource, the IaC module provisions:

- **API enablement** -- `aiplatform.googleapis.com` on the project
- **RAG Engine configuration** -- a `vertex.AiRagEngineConfig` for the location, set to the declared tier

## Before You Deploy

### Planton Setup

- **GCP Provider Connection** -- an active connection in the Connect module with `roles/aiplatform.admin` on the project. Map it as the default for your environment, or specify it explicitly when creating the Cloud Resource.
- **Planton Runner** -- required when using Runner-based credential delivery. Not needed for inline credentials or browser OAuth authentication modes.

## Deploy

### Console

Open the deployment store, find **GCP Vertex AI RAG Engine Config**, and click **Deploy**. The creation wizard walks you through preset selection, environment and connection configuration, and spec fields. Start from the **Basic Tier** preset in the [Presets](#presets) tab.

### CLI

Create a manifest and apply it:

```yaml
apiVersion: gcp.planton.dev/v1alpha1
kind: GcpVertexAiRagEngineConfig
metadata:
  name: rag-engine-us-central1
  org: acme-corp
  env: prod
spec:
  location: us-central1
  tier: SCALED
  deletionPolicy: PREVENT
```

```shell
planton apply -f rag-engine-config.yaml
```

This sets the Scaled tier for `us-central1` and guards it against an accidental destroy. A Stack Job tracks the provisioning in real time.

### InfraChart

When deploying as part of a multi-resource environment, place this block before the agents and applications that ingest into RAG corpora in that location; nothing references its outputs, so it simply needs to exist first.

## Key Configuration

These are the most important decisions when configuring the RAG Engine tier. Explore the full field reference in the [API Explorer](#api-explorer) tab.

**Tier** -- `BASIC` is cost-effective and low-compute: experiments, small corpora, latency-insensitive retrieval, or RAG Engine paired with an external vector database. `SCALED` is production grade with autoscaling. `UNPROVISIONED` disables the managed database and deletes all its data.

**What destroy means** -- destroying the block unprovisions the location, which deletes the data. `deletionPolicy: ABANDON` stops managing the tier while keeping it; `PREVENT` makes destroy fail.

**One per location** -- the configuration is a singleton Google owns; two blocks for one location overwrite each other.

## Outputs and Dependencies

### What This Component Consumes

| Dependency | Field | ValueFromRef Path |
|------------|-------|-------------------|
| **GcpProject** | `projectId` | `status.outputs.project_id` |

### What This Component Provides

After provisioning, `status.outputs` contains values that downstream Cloud Resources can consume via ValueFromRef:

| Output | Description | Common Downstream Use |
|--------|-------------|----------------------|
| `name` | The singleton's full resource name | Audit |
| `location` | The governed location | Placing corpora |

## Common Patterns

Browse the [Presets](#presets) tab for ready-to-deploy configurations.

**Basic tier** -- Turn RAG Engine on for experiments, with `ABANDON` so leaving management keeps the data. Start from the **Basic Tier** preset.

**Scaled tier** -- Production performance guarded by `PREVENT`. Start from the **Scaled Tier** preset.

## Works With

- [**GCP Vector Search Collection**](/cloud-catalog/gcp-vector-search-collection) -- an external vector database RAG Engine can use instead of its managed one
- [**GCP Vertex AI Agent Engine**](/cloud-catalog/gcp-vertex-ai-agent-engine) -- the agent runtime that consumes RAG corpora
