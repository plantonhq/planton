# GcpVertexAiDeployedIndex

> Generated from the protobuf schema by `make generate-reference` -- do not
> edit by hand. To change a fact on this page, change the proto field comment
> or validation rule it is derived from, then regenerate.

**apiVersion**: `gcp.planton.dev/v1alpha1`

**Guide**: [GUIDE.md](../GUIDE.md) -- authored operational judgment for this kind: conventions, trade-offs, and what pairs well with it.

GcpVertexAiDeployedIndexSpec places a GcpVertexAiIndex onto a
GcpVertexAiIndexEndpoint and gives the placement its serving compute —
the final resource of the vector-search trio, after which queries can
be served. Many deployed indexes can share one endpoint, and one index
can be deployed to many endpoints.

Deploying takes tens of minutes (the provider allows up to 45); the
only fields that update in place afterwards are the replica bounds
inside the sizing arm. Everything else — the placement itself, the
auth config, networking pins, deployment group — replaces the
deployment (undeploy + redeploy).

This resource class carries NO labels and NO project field in the
GCP API: platform label attribution is impossible here (the deployed
index inherits the endpoint's project and lives inside the endpoint
resource), so none is faked.

## Example

```yaml
apiVersion: gcp.planton.dev/v1alpha1
kind: GcpVertexAiDeployedIndex
metadata:
  name: test-deployed-index
spec:
  location: us-central1
  deployedIndexId: products_v1
  index:
    value: projects/my-gcp-project/locations/us-central1/indexes/1234567890
  indexEndpoint:
    value: projects/my-gcp-project/locations/us-central1/indexEndpoints/9876543210
  displayName: Products v1 Deployment
  dedicatedResources:
    machineType: e2-standard-16
    minReplicaCount: 2
    maxReplicaCount: 5
  deploymentGroup: prod
  enableAccessLogging: true
  reservedIpRanges:
    - value: vertex-ai-range-a
  authConfig:
    allowedIssuers:
      - value: query-sa@my-gcp-project.iam.gserviceaccount.com
    audiences:
      - vector-search-clients
  deletionPolicy: DELETE
```

## Spec Fields

| Path | Type | Required | Default | References |
|---|---|---|---|---|
| `spec.location` | `string` | yes |  |  |
| `spec.deployedIndexId` | `string` | yes |  |  |
| `spec.index` | `string \| valueFrom` | yes |  | GcpVertexAiIndex (`status.outputs.index_id`) |
| `spec.indexEndpoint` | `string \| valueFrom` | yes |  | GcpVertexAiIndexEndpoint (`status.outputs.index_endpoint_id`) |
| `spec.displayName` | `string` |  |  |  |
| `spec.automaticResources` | `GcpVertexAiDeployedIndexAutomaticResources` |  |  |  |
| `spec.automaticResources.minReplicaCount` | `int32` |  | `2` |  |
| `spec.automaticResources.maxReplicaCount` | `int32` |  |  |  |
| `spec.dedicatedResources` | `GcpVertexAiDeployedIndexDedicatedResources` |  |  |  |
| `spec.dedicatedResources.machineType` | `string` |  |  |  |
| `spec.dedicatedResources.minReplicaCount` | `int32` | yes |  |  |
| `spec.dedicatedResources.maxReplicaCount` | `int32` |  |  |  |
| `spec.deploymentGroup` | `string` |  | `default` |  |
| `spec.enableAccessLogging` | `bool` |  |  |  |
| `spec.reservedIpRanges` | `[]string \| valueFrom` |  |  | GcpGlobalAddress (`status.outputs.name`) |
| `spec.authConfig` | `GcpVertexAiDeployedIndexAuthConfig` |  |  |  |
| `spec.authConfig.allowedIssuers` | `[]string \| valueFrom` |  |  | GcpServiceAccount (`status.outputs.email`) |
| `spec.authConfig.audiences` | `[]string` |  |  |  |
| `spec.deletionPolicy` | `string` |  |  |  |

## Field Details

### spec.location

`string` · required

Region of the index endpoint (e.g., "us-central1"). Must match the
endpoint's own region — the Vertex AI API host is regional
(https://{region}-aiplatform.googleapis.com) and a deployment
cannot cross regions. Immutable after creation.

- rule: {"required":true,"string":{"minLen":"1"}}

### spec.deployedIndexId

`string` · required

User-chosen ID of this deployment, unique within the project: up to
128 characters, starting with a letter, containing only letters,
numbers, and underscores. This is the handle queries and undeploy
operations address the deployment by. Immutable after creation.

- rule: deployed_index_id must start with a letter and contain only letters, numbers, and underscores (up to 128 characters)
- rule: {"required":true}

### spec.index

`string | valueFrom` · required

The index being deployed — the fully qualified index resource path
(projects/{project}/locations/{location}/indexes/{indexId}); a
GcpVertexAiIndex reference resolves to it. The index must live in
the same region as the endpoint. Immutable after creation.

- references: GcpVertexAiIndex (`status.outputs.index_id`)
- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: GcpVertexAiIndex, name: <that resource's name>, fieldPath: status.outputs.index_id}} -- a bare string does not parse

### spec.indexEndpoint

`string | valueFrom` · required

The index endpoint being deployed onto — the fully qualified
endpoint resource path
(projects/{project}/locations/{location}/indexEndpoints/{id}); a
GcpVertexAiIndexEndpoint reference resolves to it. Immutable after
creation.

- references: GcpVertexAiIndexEndpoint (`status.outputs.index_endpoint_id`)
- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: GcpVertexAiIndexEndpoint, name: <that resource's name>, fieldPath: status.outputs.index_endpoint_id}} -- a bare string does not parse

### spec.displayName

`string`

Display name of the deployment (up to 128 UTF-8 characters).
Unusually for a display name, the API treats it as IMMUTABLE on a
deployed index — changing it replaces the deployment.

- rule: {"string":{"maxLen":"128"}}

### spec.automaticResources

`GcpVertexAiDeployedIndexAutomaticResources`

Vertex-managed serving compute (machine types chosen by GCP,
replicas scale between bounds). Mutually exclusive with
dedicated_resources; omitting both lets GCP deploy with automatic
defaults.

### spec.automaticResources.minReplicaCount

`int32` · optional (explicit presence)

Minimum replicas always running. GCP's default is 2 (no SLA is
provided at 1). Mutable in place — replica bounds are the only
post-deploy tuning knobs.

- default: `2`
- rule: {"int32":{"gte":1}}

### spec.automaticResources.maxReplicaCount

`int32`

Maximum replicas under load, up to 1000. Defaults to
min_replica_count when unset. Mutable in place.

- rule: {"int32":{"lte":1000,"gte":0}}

### spec.dedicatedResources

`GcpVertexAiDeployedIndexDedicatedResources`

Explicitly pinned serving compute (machine type + replica bounds).
Mutually exclusive with automatic_resources.

### spec.dedicatedResources.machineType

`string`

Machine type serving the index (e.g. "e2-standard-16"). Must be
compatible with the index's shard_size. If omitted, the API applies
its own default. Immutable.

### spec.dedicatedResources.minReplicaCount

`int32` · required

Minimum replicas always running, at least 1 (no SLA at 1). Required
by the API for dedicated sizing. Mutable in place — replica bounds
are the only post-deploy tuning knobs.

- rule: {"required":true,"int32":{"gte":1}}

### spec.dedicatedResources.maxReplicaCount

`int32`

Maximum replicas under load, up to 1000. Defaults to
min_replica_count when unset. Mutable in place.

- rule: {"int32":{"lte":1000,"gte":0}}

### spec.deploymentGroup

`string` · optional (explicit presence)

Deployment group for IP-space partitioning, up to 64 characters
(e.g. "test", "prod"); GCP's default group is "default". Pairing
groups with reserved_ip_ranges gives each group a predictable IP
space when the peered network has multiple peering ranges — and the
API HOLDS the pairing: a non-default group, once used with a set of
reserved ranges, can only ever be used with exactly that set again.
At most 5 groups besides "default". Immutable after creation.

- default: `default`
- rule: {"string":{"maxLen":"64"}}

### spec.enableAccessLogging

`bool`

If true, private-endpoint access logs are sent to Cloud Logging.
Immutable after creation.

### spec.reservedIpRanges

`[]string | valueFrom`

Names of reserved compute address ranges under the endpoint's
peered VPC network to deploy into (e.g. ["vertex-ai-ip-range"]) —
Vertex peering ranges are global INTERNAL VPC_PEERING addresses, so
a GcpGlobalAddress reference resolves to its name. If omitted, the
index may deploy to any range under the network. Only meaningful on
a VPC-peered endpoint. Immutable after creation.

- references: GcpGlobalAddress (`status.outputs.name`)
- rule: write as {value: <literal>} or {valueFrom: {kind: GcpGlobalAddress, name: <that resource's name>, fieldPath: status.outputs.name}} -- a bare string does not parse

### spec.authConfig

`GcpVertexAiDeployedIndexAuthConfig`

JWT authentication for the private query endpoint. If omitted, the
endpoint relies on network reachability alone. Immutable after
creation.

### spec.authConfig.allowedIssuers

`[]string | valueFrom`

Service accounts whose signed JWTs are accepted, each in the form
service-account-name@project-id.iam.gserviceaccount.com. Immutable.

- references: GcpServiceAccount (`status.outputs.email`)
- rule: write as {value: <literal>} or {valueFrom: {kind: GcpServiceAccount, name: <that resource's name>, fieldPath: status.outputs.email}} -- a bare string does not parse

### spec.authConfig.audiences

`[]string`

JWT audiences accepted on the query endpoint; a JWT carrying any of
them is accepted. Immutable.

### spec.deletionPolicy

`string`

Deletion policy for the deployment — what happens when this
resource is destroyed:
  ""        -- same as "DELETE" (provider default)
  "DELETE"  -- the index is undeployed from the endpoint; queries
               against this deployment stop being served
  "PREVENT" -- destroy FAILS; a guard for a serving path whose
               disappearance would break live query traffic
  "ABANDON" -- the deployment is removed from management but keeps
               serving (and billing for its replicas) in GCP

- rule: deletion_policy must be one of: DELETE, PREVENT, ABANDON

## Validation Rules

- `at_most_one_sizing_arm`: automatic_resources and dedicated_resources are mutually exclusive; let Vertex AI manage the compute or pin a machine type, not both

## Outputs

Reference an output from another manifest as `valueFrom: {kind: GcpVertexAiDeployedIndex, name: <resource-name>, fieldPath: status.outputs.<output>}`.

| Output | Type | Description |
|---|---|---|
| `status.outputs.name` | `string` | Name of the DeployedIndex resource as the provider reports it. |
| `status.outputs.deployed_index_id` | `string` | The user-chosen deployment ID (pass-through from spec for downstream reference without needing to read the full resource). |
| `status.outputs.create_time` | `string` | RFC3339 timestamp of when the deployment was created. |
| `status.outputs.index_sync_time` | `string` | RFC3339 timestamp up to which this deployment reflects the source index's updates: if it is at least the index's update_time, the deployment is in sync with the index contents. |
| `status.outputs.match_grpc_address` | `string` | Private gRPC address for match queries inside the peered VPC. Populated only when the endpoint is VPC-peered. |
| `status.outputs.service_attachment` | `string` | PSC service attachment consumers target with forwarding rules. Populated only when the endpoint uses Private Service Connect. |
| `status.outputs.index_endpoint` | `string` | Fully qualified resource path of the index endpoint this deployment lives on (projects/{project}/locations/{location}/indexEndpoints/{id}). Query clients need the endpoint path and the deployment ID together — this output carries the pair's other half. |

## References

Fields that can point at another resource's outputs:

| Field | Kind | Output |
|---|---|---|
| `spec.index` | GcpVertexAiIndex | `status.outputs.index_id` |
| `spec.indexEndpoint` | GcpVertexAiIndexEndpoint | `status.outputs.index_endpoint_id` |
| `spec.reservedIpRanges` | GcpGlobalAddress | `status.outputs.name` |
| `spec.authConfig.allowedIssuers` | GcpServiceAccount | `status.outputs.email` |

## See Also

- [Overview](../README.md)
