# GCP Vertex AI Deployed Index

Deploys a Vertex AI Deployed Index: the resource that places a GcpVertexAiIndex onto a GcpVertexAiIndexEndpoint and gives the placement its serving compute — the final resource of the vector-search trio, after which nearest-neighbor queries can actually be served. Many deployed indexes can share one endpoint, and one index can be deployed to many endpoints. Deploying takes tens of minutes, and only the replica bounds update in place afterwards — everything else replaces the deployment.

## What Gets Created

When you deploy this Infra Component, the IaC module provisions:

- **Deployed Index** -- the placement of the referenced index onto the referenced endpoint, addressed by your chosen deployment ID
- **Serving Compute** -- either Vertex-managed automatic resources (machine types chosen by GCP, replicas scaling between your bounds) or dedicated resources (a pinned machine type); omitting both deploys with GCP's automatic defaults
- **Private-Serving Refinements** -- on peered/PSC endpoints: the deployment group's IP-space partition, reserved peering ranges, private-endpoint access logging, and optional JWT authentication on the private query endpoint

## Before You Deploy

### Planton Setup

- **GCP Provider Connection** -- an active connection in the Connect module with credentials for the target GCP project. Map it as the default for your environment, or specify it explicitly when creating the Infra Component.
- **Planton Runner** -- required when using Runner-based credential delivery. Not needed for inline credentials or browser OAuth authentication modes.

### GCP Project

- **A GcpVertexAiIndex and a GcpVertexAiIndexEndpoint** in the SAME region — the deployment joins them and cannot cross regions.
- **Vertex AI API** enabled in the endpoint's project (the deployment inherits it — this kind carries no project field).
- **Time budget** -- deploying takes tens of minutes (the provider allows up to 45).

## Deploy

### Console

Open the deployment store, find **GCP Vertex AI Deployed Index**, and click **Deploy**. The creation wizard walks you through preset selection, environment and connection configuration, and spec fields. Start from the **Automatic Serving Compute** preset in the [Presets](#presets) tab for Vertex-managed sizing.

### CLI

Create a manifest and apply it:

```yaml
apiVersion: gcp.planton.dev/v1alpha1
kind: GcpVertexAiDeployedIndex
metadata:
  name: catalog-deployment
  org: acme-corp
  env: prod
spec:
  location: us-central1
  deployedIndexId: catalog_embeddings_v1
  index:
    value: "projects/acme-prod-12345/locations/us-central1/indexes/8123456789012345678"
  indexEndpoint:
    value: "projects/acme-prod-12345/locations/us-central1/indexEndpoints/6098765432109876543"
  automaticResources:
    minReplicaCount: 2
    maxReplicaCount: 10
```

```shell
planton apply -f deployed-index.yaml
```

This places the referenced index onto the referenced endpoint with Vertex-managed compute scaling between 2 and 10 replicas — after which nearest-neighbor queries can be served. An Infra Job tracks the provisioning in real time.

### InfraChart

The deployment is the natural LAST node of the vector-search chart — wire both joins to the index and endpoint deployed in the same InfraPipeline:

```yaml
spec:
  index:
    valueFrom:
      kind: GcpVertexAiIndex
      name: catalog-embeddings
      fieldPath: status.outputs.index_id
  indexEndpoint:
    valueFrom:
      kind: GcpVertexAiIndexEndpoint
      name: catalog-search
      fieldPath: status.outputs.index_endpoint_id
```

The InfraPipeline resolves the dependency graph, provisions the index and endpoint first, then this deployment — one chart deploys the whole serving composition.

## Key Configuration

These are the most important decisions when configuring a deployed index. Explore the full field reference in the [API Explorer](#api-explorer) tab.

**The two joins** -- `index` and `indexEndpoint` (both required, both immutable) reference the index's `index_id` and the endpoint's `index_endpoint_id` outputs. All three resources must share one region.

**Sizing arm** -- `automaticResources` (Vertex-managed) and `dedicatedResources` (pinned machine type) are mutually exclusive; omitting both deploys with automatic defaults. Dedicated machine types must be compatible with the INDEX's shard size (LARGE shards need e2-highmem-16 or n2d-standard-32). **Replica bounds are the deployment's ONLY mutable fields** — everything else replaces it.

**Deployment group + reserved ranges** -- partition a peered endpoint's IP space; the API permanently holds a non-default group's pairing with its range set. Reference GcpGlobalAddress resources (purpose VPC_PEERING) by `name`.

**JWT auth** -- `authConfig` gates the private query endpoint on tokens signed by allowed issuers (GcpServiceAccount `email` references) carrying accepted audiences. Without it, network reachability alone controls access.

## Outputs and Dependencies

### What This Kind Consumes

| Dependency | Field | ValueFromRef Path |
|------------|-------|-------------------|
| **GcpVertexAiIndex** (required) | `index` | `status.outputs.index_id` |
| **GcpVertexAiIndexEndpoint** (required) | `indexEndpoint` | `status.outputs.index_endpoint_id` |
| **GcpGlobalAddress** (optional, repeated) | `reservedIpRanges` | `status.outputs.name` |
| **GcpServiceAccount** (optional, repeated) | `authConfig.allowedIssuers` | `status.outputs.email` |

### What This Kind Provides

After provisioning, `status.outputs` contains values that downstream consumers and query clients use:

| Output | Description | Common Downstream Use |
|--------|-------------|----------------------|
| `name` | The DeployedIndex resource name as the provider reports it | Display, tooling |
| `deployed_index_id` | Pass-through of the user-chosen deployment handle | Query SDKs address the deployment by it |
| `index_endpoint` | The fully qualified endpoint path this deployment lives on | Cross-checks, tooling |
| `index_sync_time` | RFC3339 timestamp up to which this deployment reflects index updates | Data-freshness checks before critical queries |
| `match_grpc_address` | Private gRPC query address (VPC-peered endpoints only) | Private query clients inside the peered VPC |
| `service_attachment` | PSC service attachment (PSC endpoints only) | Consumer projects' forwarding rules |

## Common Patterns

Browse the [Presets](#presets) tab for ready-to-deploy configurations.

**Automatic sizing** -- Vertex-managed machine types with tuned replica bounds; the managed middle ground for most workloads. Start from the **Automatic Serving Compute** preset.

**Dedicated compute** -- A pinned machine type for predictable performance and cost on sustained query volume. Start from the **Dedicated Serving Compute** preset.

**Peered with reserved ranges** -- Private serving with pinned IP space, access logging, and JWT auth for sensitive corpora. Start from the **Peered Endpoint with Reserved Ranges and JWT Auth** preset.

## Works With

- [**GCP Vertex AI Index**](/infra-catalog/gcp-vertex-ai-index) -- the vector index this deployment serves
- [**GCP Vertex AI Index Endpoint**](/infra-catalog/gcp-vertex-ai-index-endpoint) -- the serving surface this deployment lives on
- [**GCP Global Address**](/infra-catalog/gcp-global-address) -- the reserved VPC_PEERING ranges a peered deployment pins its IP space to
- [**GCP Service Account**](/infra-catalog/gcp-service-account) -- the JWT issuers the private query endpoint trusts
