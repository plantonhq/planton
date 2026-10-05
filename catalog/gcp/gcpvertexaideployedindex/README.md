# GcpVertexAiDeployedIndex

A GCP Vertex AI deployed index places a `GcpVertexAiIndex` onto a `GcpVertexAiIndexEndpoint` and gives the placement its serving compute -- the final resource of the vector-search trio, after which nearest-neighbor queries can actually be served. Many deployed indexes can share one endpoint, and one index can be deployed to many endpoints.

## When to Use

Use `GcpVertexAiDeployedIndex` when you need:

- To make a Vector Search index queryable (an index alone stores vectors; only a deployment serves them)
- Explicit control over serving compute (machine type, replica bounds)
- Predictable IP-space placement on a peered VPC (deployment groups + reserved ranges)
- JWT authentication on private query endpoints

## What This Kind Creates

This kind deploys one index onto one index endpoint. The index and the endpoint are separate resources -- both must exist first (the spec references them by their fully qualified resource paths).

## Key Configuration Options

### Serving Compute (sizing arms, mutually exclusive)

- **`automaticResources`** -- Vertex AI picks machine types; replicas scale between bounds. Zero configuration.
- **`dedicatedResources`** -- pin a machine type (must be compatible with the index's `shardSize`) and replica bounds. Predictable performance and cost.

The replica bounds are the **only fields that update in place** after deployment; every other change undeploys and redeploys.

### Placement Handle

`deployedIndexId` (required, immutable) is the handle queries and undeploy operations address the deployment by: letter start, letters/numbers/underscores, up to 128 characters.

### Networking Pins (peered endpoints)

- `reservedIpRanges` -- names of reserved `VPC_PEERING` global addresses to deploy into (references `GcpGlobalAddress`).
- `deploymentGroup` -- IP-space partitioning label. The API permanently pairs a non-default group with the exact set of reserved ranges it first ships with.

### Authentication

`authConfig` enables JWT auth on the private query endpoint: `allowedIssuers` are service-account emails (references `GcpServiceAccount`), `audiences` the accepted JWT audiences.

### Destroy behavior

`deletionPolicy` is the client-side lever: empty/`DELETE` undeploys the index from the endpoint, `PREVENT` makes destroy fail (a guard for a live query path), `ABANDON` removes the deployment from management but keeps it serving (and billing for its replicas).

### No Labels, No Project

The GCP API gives a DeployedIndex no labels and no project field (it lives inside the endpoint resource) -- platform label attribution is impossible on this resource class, so none is faked.

## Outputs

| Output | Description |
|--------|-------------|
| `name` | Name of the DeployedIndex resource |
| `deployed_index_id` | The user-chosen deployment handle |
| `create_time` | Creation timestamp |
| `index_sync_time` | Timestamp up to which the deployment reflects the source index's updates |
| `match_grpc_address` | Private gRPC query address (peered endpoints only) |
| `service_attachment` | PSC service attachment (PSC endpoints only) |
| `index_endpoint` | Full resource path of the endpoint this deployment lives on |

## Presets

- **automatic** -- Zero-configuration serving compute
- **dedicated** -- Pinned machine type with explicit replica bounds
- **peered-reserved-ranges** -- Deployment pinned to reserved IP ranges with JWT auth

---

© Planton. Licensed under [Apache-2.0](https://github.com/plantonhq/planton/blob/main/LICENSE).
