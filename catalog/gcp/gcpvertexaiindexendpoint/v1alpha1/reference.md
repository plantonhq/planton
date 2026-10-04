# GcpVertexAiIndexEndpoint

> Generated from the protobuf schema by `make generate-reference` -- do not
> edit by hand. To change a fact on this page, change the proto field comment
> or validation rule it is derived from, then regenerate.

**apiVersion**: `gcp.planton.dev/v1alpha1`

**Guide**: [GUIDE.md](../GUIDE.md) -- authored operational judgment for this kind: conventions, trade-offs, and what pairs well with it.

GcpVertexAiIndexEndpointSpec defines a Vertex AI Vector Search index
endpoint — the serving surface that deployed indexes answer
nearest-neighbor queries through. The endpoint owns connectivity and
capacity attachment; the indexes themselves are separate
GcpVertexAiIndex resources placed onto the endpoint by
GcpVertexAiDeployedIndex.

This is a DIFFERENT GCP resource from the online-prediction
GcpVertexAiEndpoint (which serves models): an IndexEndpoint serves
vector-search queries only.

Three connectivity modes, mutually exclusive:

  - **Public**: set `public_endpoint_enabled: true`. Queries go to a
    public domain name (surfaced as an output once created).

  - **VPC-peered**: set `network`. The endpoint is reachable only
    inside the peered VPC. Requires Private Services Access on the
    network.

  - **Private Service Connect**: set
    `private_service_connect_config`. Consumers connect through a
    service attachment; no peering needed.

Every connectivity choice is immutable (ForceNew). Mutable in place:
display_name, description, labels.

## Example

```yaml
apiVersion: gcp.planton.dev/v1alpha1
kind: GcpVertexAiIndexEndpoint
metadata:
  name: test-index-endpoint
spec:
  projectId:
    value: my-gcp-project
  location: us-central1
  displayName: Test Index Endpoint
  description: Serving surface for the test vector index
  publicEndpointEnabled: true
  deletionPolicy: DELETE
  labels:
    team: ml-platform
    cost-center: research
```

## Spec Fields

| Path | Type | Required | Default | References |
|---|---|---|---|---|
| `spec.projectId` | `string \| valueFrom` |  |  | GcpProject (`status.outputs.project_id`) |
| `spec.location` | `string` | yes |  |  |
| `spec.displayName` | `string` | yes |  |  |
| `spec.description` | `string` |  |  |  |
| `spec.publicEndpointEnabled` | `bool` |  |  |  |
| `spec.network` | `string \| valueFrom` |  |  | GcpVpcNetwork (`status.outputs.network_self_link`) |
| `spec.privateServiceConnectConfig` | `GcpVertexAiIndexEndpointPrivateServiceConnectConfig` |  |  |  |
| `spec.privateServiceConnectConfig.enablePrivateServiceConnect` | `bool` |  |  |  |
| `spec.privateServiceConnectConfig.projectAllowlist` | `[]string` |  |  |  |
| `spec.privateServiceConnectConfig.pscAutomationConfigs` | `[]GcpVertexAiIndexEndpointPscAutomationConfig` |  |  |  |
| `spec.privateServiceConnectConfig.pscAutomationConfigs[].network` | `string \| valueFrom` | yes |  | GcpVpcNetwork (`status.outputs.network_self_link`) |
| `spec.privateServiceConnectConfig.pscAutomationConfigs[].projectId` | `string \| valueFrom` | yes |  | GcpProject (`status.outputs.project_id`) |
| `spec.labels` | `map<string, string>` |  |  |  |
| `spec.kmsKeyName` | `string \| valueFrom` |  |  | GcpKmsKey (`status.outputs.key_id`) |
| `spec.deletionPolicy` | `string` |  |  |  |

## Field Details

### spec.projectId

`string | valueFrom`

GCP project where the index endpoint will be created.
If omitted, the endpoint is created in the provider's default
project (from the credential or ambient configuration).

- references: GcpProject (`status.outputs.project_id`)
- rule: write as {value: <literal>} or {valueFrom: {kind: GcpProject, name: <that resource's name>, fieldPath: status.outputs.project_id}} -- a bare string does not parse

### spec.location

`string` · required

Region where the index endpoint will be created (e.g.,
"us-central1"). Indexes deployed onto it must live in the same
region. Immutable after creation.

- rule: {"required":true,"string":{"minLen":"1"}}

### spec.displayName

`string` · required

Display name of the index endpoint (up to 128 UTF-8 characters).
The primary human-readable identifier; the numeric resource ID is
GCP-assigned.

- rule: {"required":true,"string":{"minLen":"1","maxLen":"128"}}

### spec.description

`string`

Description of the index endpoint.

### spec.publicEndpointEnabled

`bool`

If true, deployed indexes are queryable through a public domain
name (the public_endpoint_domain_name output). Mutually exclusive
with network and private_service_connect_config. Immutable.

### spec.network

`string | valueFrom`

VPC network to peer the endpoint into (private queries via Private
Services Access). The Vertex AI API expects the RELATIVE network
form projects/{project}/global/networks/{name} (with {project}
preferably the project NUMBER); both IaC modules normalize a
compute self-link URL (the GcpVpcNetwork reference's canonical
output) to that relative form. Requires Private Services Access
configured on the network. Mutually exclusive with
public_endpoint_enabled and private_service_connect_config.
Immutable.

- references: GcpVpcNetwork (`status.outputs.network_self_link`)
- rule: write as {value: <literal>} or {valueFrom: {kind: GcpVpcNetwork, name: <that resource's name>, fieldPath: status.outputs.network_self_link}} -- a bare string does not parse

### spec.privateServiceConnectConfig

`GcpVertexAiIndexEndpointPrivateServiceConnectConfig`

Private Service Connect configuration. When present, consumers
reach deployed indexes through a PSC service attachment. Mutually
exclusive with public_endpoint_enabled and network. Immutable.

- rule: psc_automation_configs is not honored on index endpoints — the Vertex AI API accepts the create but silently drops the configs (nothing is stored and no consumer-side endpoint is ever provisioned); use project_allowlist with consumer-managed forwarding rules instead

### spec.privateServiceConnectConfig.enablePrivateServiceConnect

`bool`

Must be true when this block is present — the API's enablement flag
for PSC on the endpoint. Modeled explicitly (not inferred from block
presence) because it is the API's own contract field. Immutable.

- rule: enable_private_service_connect must be true when the Private Service Connect block is present — remove the block instead of disabling it

### spec.privateServiceConnectConfig.projectAllowlist

`[]string`

Projects allowed to create forwarding rules targeting this
endpoint's service attachment. Each entry is a GCP project ID or
project number. Immutable.

### spec.privateServiceConnectConfig.pscAutomationConfigs

`[]GcpVertexAiIndexEndpointPscAutomationConfig`

PSC endpoints Vertex AI creates automatically in consumer
projects/networks (instead of consumers wiring forwarding rules by
hand). The provider models this field on index endpoints, but the
live API does NOT honor it there: a create carrying automation
configs succeeds while the stored endpoint omits them and no
consumer-side endpoint is ever provisioned (API-verified against a
live index endpoint; the provider documents the field as used by
online inference endpoints only). Because the PSC block is
immutable, that silent drop would surface to users as a perpetual
replacement diff on every re-plan — so this field is refused by
validation until Google extends automation to vector search. Use
project_allowlist with consumer-managed forwarding rules. Immutable.

### spec.privateServiceConnectConfig.pscAutomationConfigs[].network

`string | valueFrom` · required

VPC network where the PSC endpoint is created, as the full
relative resource name projects/{project}/global/networks/{name}
(the format the Vertex AI API requires) — a GcpVpcNetwork
reference's self-link output is normalized to that form by both
IaC modules. Immutable.

- references: GcpVpcNetwork (`status.outputs.network_self_link`)
- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: GcpVpcNetwork, name: <that resource's name>, fieldPath: status.outputs.network_self_link}} -- a bare string does not parse

### spec.privateServiceConnectConfig.pscAutomationConfigs[].projectId

`string | valueFrom` · required

Project in which the PSC endpoint (forwarding rule) is created —
a project ID; a GcpProject reference resolves to it. Immutable.

- references: GcpProject (`status.outputs.project_id`)
- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: GcpProject, name: <that resource's name>, fieldPath: status.outputs.project_id}} -- a bare string does not parse

### spec.labels

`map<string, string>`

User-defined labels to organize the index endpoint (cost
attribution, team ownership, environment tagging). Keys and values
must follow GCP label rules: lowercase letters, digits,
underscores, and dashes, at most 63 characters. Merged with the
platform's attribution labels; on key conflicts the platform
labels win. Mutable in place.

### spec.kmsKeyName

`string | valueFrom`

Cloud KMS key for customer-managed encryption at rest (CMEK) of
data on the endpoint's serving replicas, as the full key resource
path
projects/{project}/locations/{location}/keyRings/{ring}/cryptoKeys/{key}
— a GcpKmsKey reference resolves to it. The key must live in the
same region as the endpoint, and the Vertex AI service agent needs
roles/cloudkms.cryptoKeyEncrypterDecrypter on it. If omitted, data
is encrypted with Google-managed keys. Immutable after creation.

- references: GcpKmsKey (`status.outputs.key_id`)
- rule: write as {value: <literal>} or {valueFrom: {kind: GcpKmsKey, name: <that resource's name>, fieldPath: status.outputs.key_id}} -- a bare string does not parse

### spec.deletionPolicy

`string`

Deletion policy for the index endpoint — what happens when this
resource is destroyed:
  ""        -- same as "DELETE" (provider default)
  "DELETE"  -- the endpoint is deleted; every index deployed onto
               it stops serving
  "PREVENT" -- destroy FAILS; a guard for the serving surface all
               of this endpoint's deployed indexes depend on
  "ABANDON" -- the endpoint is removed from management but left
               standing (and billing for its replicas) in GCP

- rule: deletion_policy must be one of: DELETE, PREVENT, ABANDON

## Validation Rules

- `network_psc_mutual_exclusion`: network and private_service_connect_config are mutually exclusive; use VPC peering or PSC, not both
- `public_network_mutual_exclusion`: public_endpoint_enabled and network are mutually exclusive; a peered endpoint is private by definition
- `public_psc_mutual_exclusion`: public_endpoint_enabled and private_service_connect_config are mutually exclusive; a PSC endpoint is private by definition

## Outputs

Reference an output from another manifest as `valueFrom: {kind: GcpVertexAiIndexEndpoint, name: <resource-name>, fieldPath: status.outputs.<output>}`.

| Output | Type | Description |
|---|---|---|
| `status.outputs.index_endpoint_id` | `string` | Fully qualified index endpoint resource path — the value a GcpVertexAiDeployedIndex passes as its `index_endpoint` reference. Format: projects/{project}/locations/{location}/indexEndpoints/{id} |
| `status.outputs.index_endpoint_name` | `string` | The GCP-assigned numeric index endpoint ID (the last path segment of index_endpoint_id). |
| `status.outputs.public_endpoint_domain_name` | `string` | Domain name for querying deployed indexes over the public internet. Populated only when public_endpoint_enabled is true. |
| `status.outputs.create_time` | `string` | RFC3339 timestamp of when the index endpoint was created. |
| `status.outputs.update_time` | `string` | RFC3339 timestamp of when the index endpoint was last updated. |

## References

Fields that can point at another resource's outputs:

| Field | Kind | Output |
|---|---|---|
| `spec.projectId` | GcpProject | `status.outputs.project_id` |
| `spec.network` | GcpVpcNetwork | `status.outputs.network_self_link` |
| `spec.privateServiceConnectConfig.pscAutomationConfigs[].network` | GcpVpcNetwork | `status.outputs.network_self_link` |
| `spec.privateServiceConnectConfig.pscAutomationConfigs[].projectId` | GcpProject | `status.outputs.project_id` |
| `spec.kmsKeyName` | GcpKmsKey | `status.outputs.key_id` |

## Referenced By

Fields on other kinds that can point at this resource:

| Kind | Field | Reads |
|---|---|---|
| GcpVertexAiDeployedIndex | `spec.indexEndpoint` | `status.outputs.index_endpoint_id` |

## See Also

- [Overview](../README.md)
