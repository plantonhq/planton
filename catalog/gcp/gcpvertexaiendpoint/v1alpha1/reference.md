# GcpVertexAiEndpoint

> Generated from the protobuf schema by `make generate-reference` -- do not
> edit by hand. To change a fact on this page, change the proto field comment
> or validation rule it is derived from, then regenerate.

**apiVersion**: `gcp.planton.dev/v1alpha1`

**Guide**: [GUIDE.md](../GUIDE.md) -- authored operational judgment for this component: conventions, trade-offs, and what pairs well with it.

GcpVertexAiEndpointSpec defines the configuration for a GCP Vertex AI
Endpoint -- a stable serving surface for deploying machine learning models.

A Vertex AI Endpoint is a lightweight resource that provides a durable
URL and networking boundary for model serving. Creating the endpoint
is an infrastructure concern; deploying models to it is an operational
step performed separately via the Vertex AI API or console.

Three networking modes are available, all mutually exclusive:

  - **Public** (default): The endpoint is accessible via the shared
    regional DNS ({region}-aiplatform.googleapis.com).

  - **VPC-peered**: The endpoint is accessible only within a peered
    VPC network. Set the `network` field. Requires Private Services
    Access configured on the VPC.

  - **Private Service Connect**: The endpoint is exposed via a PSC
    service attachment. Set `private_service_connect_config`. Provides
    the strongest network isolation without VPC peering.

The `dedicated_endpoint_enabled` flag provides a dedicated DNS name
for better performance and traffic isolation, but is incompatible
with Private Service Connect.

Immutable fields (ForceNew): location, network, kms_key_name,
endpoint_name, and the PSC block's enablement itself. Changing these
requires destroying and recreating the endpoint. Mutable in place:
display_name, description, labels, traffic_split, the request/response
logging config, and the PSC block's project_allowlist and
psc_automation_configs (the provider PATCHes them via updateMask).

## Example

```yaml
apiVersion: gcp.planton.dev/v1alpha1
kind: GcpVertexAiEndpoint
metadata:
  name: test-ml-endpoint
spec:
  projectId:
    value: my-gcp-project
  location: us-central1
  displayName: Test ML Endpoint
  description: Endpoint for serving recommendation models
  network:
    value: projects/123456789/global/networks/my-vpc
  kmsKeyName:
    value: projects/my-gcp-project/locations/us-central1/keyRings/my-ring/cryptoKeys/my-key
  dedicatedEndpointEnabled: true
  endpointName: "1234567890"
  labels:
    team: ml-platform
    cost-center: research
  requestResponseLoggingConfig:
    enabled: true
    samplingRate: 0.25
    bigqueryDestinationUri: bq://my-gcp-project.ml_logging.endpoint_requests
  # Traffic routing across deployed models (IDs are assigned at model
  # deployment, an operational step outside this resource); values must sum
  # to exactly 100. Leave empty on an endpoint with no deployed models.
  trafficSplit:
    "1234567890": 100
  # Delete the endpoint on destroy (GCP's default, made explicit). GCP
  # rejects the delete while models are still deployed — undeploy first.
  deletionPolicy: DELETE
```

## Spec Fields

| Path | Type | Required | Default | References |
|---|---|---|---|---|
| `spec.projectId` | `string \| valueFrom` |  |  | GcpProject (`status.outputs.project_id`) |
| `spec.location` | `string` | yes |  |  |
| `spec.displayName` | `string` | yes |  |  |
| `spec.description` | `string` |  |  |  |
| `spec.network` | `string \| valueFrom` |  |  | GcpVpcNetwork (`status.outputs.network_self_link`) |
| `spec.kmsKeyName` | `string \| valueFrom` |  |  | GcpKmsKey (`status.outputs.key_id`) |
| `spec.dedicatedEndpointEnabled` | `bool` |  |  |  |
| `spec.privateServiceConnectConfig` | `GcpVertexAiEndpointPrivateServiceConnectConfig` |  |  |  |
| `spec.privateServiceConnectConfig.projectAllowlist` | `[]string` |  |  |  |
| `spec.privateServiceConnectConfig.pscAutomationConfigs` | `[]GcpVertexAiEndpointPscAutomationConfig` |  |  |  |
| `spec.privateServiceConnectConfig.pscAutomationConfigs[].network` | `string \| valueFrom` | yes |  | GcpVpcNetwork (`status.outputs.network_self_link`) |
| `spec.privateServiceConnectConfig.pscAutomationConfigs[].projectId` | `string \| valueFrom` | yes |  | GcpProject (`status.outputs.project_id`) |
| `spec.privateServiceConnectConfig.enabled` | `bool` |  | `true` |  |
| `spec.endpointName` | `string` |  |  |  |
| `spec.labels` | `map<string, string>` |  |  |  |
| `spec.requestResponseLoggingConfig` | `GcpVertexAiEndpointRequestResponseLoggingConfig` |  |  |  |
| `spec.requestResponseLoggingConfig.enabled` | `bool` |  |  |  |
| `spec.requestResponseLoggingConfig.samplingRate` | `double` |  |  |  |
| `spec.requestResponseLoggingConfig.bigqueryDestinationUri` | `string` |  |  |  |
| `spec.trafficSplit` | `map<string, int32>` |  |  |  |
| `spec.deletionPolicy` | `string` |  |  |  |

## Field Details

### spec.projectId

`string | valueFrom`

GCP project where the endpoint will be created.
If omitted, the endpoint is created in the provider's default
project (from the credential or ambient configuration).

- references: GcpProject (`status.outputs.project_id`)
- rule: write as {value: <literal>} or {valueFrom: {kind: GcpProject, name: <that resource's name>, fieldPath: status.outputs.project_id}} -- a bare string does not parse

### spec.location

`string` · required

Region where the endpoint will be created (e.g., "us-central1").
Immutable after creation.

- rule: {"required":true,"string":{"minLen":"1"}}

### spec.displayName

`string` · required

Display name of the endpoint (up to 128 UTF-8 characters).
This is the primary human-readable identifier for the endpoint.

- rule: {"required":true,"string":{"minLen":"1","maxLen":"128"}}

### spec.description

`string`

Description of the endpoint.

### spec.network

`string | valueFrom`

VPC network for private endpoints via VPC peering.
Format: projects/{project}/global/networks/{network}
Requires Private Services Access configured on the VPC.
Mutually exclusive with private_service_connect_config.
Immutable after creation.

- references: GcpVpcNetwork (`status.outputs.network_self_link`)
- rule: write as {value: <literal>} or {valueFrom: {kind: GcpVpcNetwork, name: <that resource's name>, fieldPath: status.outputs.network_self_link}} -- a bare string does not parse

### spec.kmsKeyName

`string | valueFrom`

KMS key for customer-managed encryption (CMEK).
Format: projects/{project}/locations/{location}/keyRings/{ring}/cryptoKeys/{key}
If not specified, Google-managed encryption is used.
Immutable after creation.

- references: GcpKmsKey (`status.outputs.key_id`)
- rule: write as {value: <literal>} or {valueFrom: {kind: GcpKmsKey, name: <that resource's name>, fieldPath: status.outputs.key_id}} -- a bare string does not parse

### spec.dedicatedEndpointEnabled

`bool`

If true, the endpoint is exposed through a dedicated DNS name
(https://{endpointId}.{region}-{projectNumber}.prediction.vertexai.goog)
rather than the shared regional DNS. Dedicated endpoints provide
better performance, reliability, and traffic isolation.
Mutually exclusive with private_service_connect_config.

### spec.privateServiceConnectConfig

`GcpVertexAiEndpointPrivateServiceConnectConfig`

Private Service Connect configuration. When present, the endpoint
is exposed via a PSC service attachment rather than VPC peering.
Mutually exclusive with network and dedicated_endpoint_enabled.

### spec.privateServiceConnectConfig.projectAllowlist

`[]string`

Projects allowed to create forwarding rules targeting this endpoint's
service attachment. Each entry is a GCP project ID or project number.
If empty, any project in the same organization can connect.
Mutable in place.

### spec.privateServiceConnectConfig.pscAutomationConfigs

`[]GcpVertexAiEndpointPscAutomationConfig`

PSC endpoints Vertex AI creates automatically in consumer
projects/networks (instead of consumers wiring forwarding rules by
hand). Online-prediction endpoints are exactly the surface Google
documents this automation for. Mutable in place.

### spec.privateServiceConnectConfig.pscAutomationConfigs[].network

`string | valueFrom` · required

VPC network where the PSC endpoint is created, as the full relative
resource name projects/{project}/global/networks/{name} (the format
the Vertex AI API requires) — a GcpVpcNetwork reference's self-link
output is normalized to that form by both IaC modules.

- references: GcpVpcNetwork (`status.outputs.network_self_link`)
- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: GcpVpcNetwork, name: <that resource's name>, fieldPath: status.outputs.network_self_link}} -- a bare string does not parse

### spec.privateServiceConnectConfig.pscAutomationConfigs[].projectId

`string | valueFrom` · required

Project in which the PSC endpoint (forwarding rule) is created — a
project ID; a GcpProject reference resolves to it.

- references: GcpProject (`status.outputs.project_id`)
- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: GcpProject, name: <that resource's name>, fieldPath: status.outputs.project_id}} -- a bare string does not parse

### spec.privateServiceConnectConfig.enabled

`bool` · optional (explicit presence)

Whether Private Service Connect is on. Unset means on: declaring the
block has always meant enabling PSC, and this switch lets a manifest say
the opposite out loud while keeping the allowlist and automation
settings in place (the modules then render no PSC block, since the API
rejects the block with its flag off).

- default: `true`

### spec.endpointName

`string`

GCP endpoint name (numeric identifier, max 10 digits).
GCP requires Vertex AI endpoint names to be numeric only --
this is the resource ID in the fully qualified path
projects/{project}/locations/{location}/endpoints/{name}.

If not specified, the IaC module derives a stable numeric identifier
from the resource's own identity (organization, environment, and
resource name), so the same manifest always produces the same
endpoint ID on either provisioning engine. Most users should omit
this field and use display_name for human identification.
Immutable after creation.

Reservation caveat: GCP reserves a deleted endpoint's numeric ID —
destroying an endpoint and recreating the same resource identity
(or reusing the same explicit ID) fails with 409 ALREADY_EXISTS
until GCP releases the reservation. To recreate immediately, set a
different explicit endpoint_name or change the resource name.

- rule: endpoint_name must be numeric (1-10 digits, no leading zeros)

### spec.labels

`map<string, string>`

User-defined labels to organize the endpoint (cost attribution,
team ownership, environment tagging). Keys and values must follow
GCP label rules: lowercase letters, digits, underscores, and dashes,
at most 63 characters. Merged with the platform's attribution labels;
on key conflicts the platform labels win. Mutable in place.

### spec.requestResponseLoggingConfig

`GcpVertexAiEndpointRequestResponseLoggingConfig`

Request/response logging for online predictions: samples prediction
traffic into a BigQuery table for drift monitoring, debugging, and
audit. Mutable in place.

### spec.requestResponseLoggingConfig.enabled

`bool`

Enable request/response logging.

### spec.requestResponseLoggingConfig.samplingRate

`double`

Fraction of requests to log, in the range (0, 1]. Sample down
(e.g. 0.05) for high-QPS endpoints to bound BigQuery cost; use 1.0
to capture everything on low-traffic endpoints.

- rule: sampling_rate must be in the range (0, 1]

### spec.requestResponseLoggingConfig.bigqueryDestinationUri

`string`

BigQuery destination for the logged requests/responses, up to 2000
characters. Accepted forms:
  - "bq://projectId" -- GCP creates a dataset named
    logging_<endpoint-display-name>_<endpoint-id> and a
    request_response_logging table inside it.
  - "bq://projectId.bqDatasetId" -- GCP creates the table in the
    given dataset (the dataset must exist).
  - "bq://projectId.bqDatasetId.bqTableId" -- fully specified; the
    dataset must exist and the table must not.

A plain string (not a reference) because the bq:// URI scheme has no
matching stack output on the BigQuery kinds; compose by writing the
dataset's project and ID into the URI.

- rule: {"string":{"maxLen":"2000"}}

### spec.trafficSplit

`map<string, int32>`

Traffic routing across the models deployed on this endpoint: a map
from a DeployedModel's ID (assigned when a model is deployed — an
operational step outside this resource) to the percentage of traffic
it receives. GCP requires the values to add up to exactly 100, and
rejects IDs that are not currently deployed — so leave this EMPTY on
an endpoint with no deployed models (an empty map means the endpoint
accepts no traffic). Mutable in place: update it to shift traffic
between model versions (canary/blue-green serving).

- rule: {"map":{"values":{"int32":{"lte":100,"gte":1}}}}

### spec.deletionPolicy

`string`

What happens to the endpoint in GCP when this resource is destroyed.
  "DELETE"  -- (GCP's default when unset) the endpoint is deleted;
               GCP rejects the delete while models are still deployed
               to it, so undeploy first
  "PREVENT" -- destroy FAILS; protects a serving URL applications
               still call
  "ABANDON" -- the endpoint is removed from management but keeps
               serving in GCP (deployed models keep billing)

- rule: deletion_policy must be one of: DELETE, PREVENT, ABANDON

## Validation Rules

- `network_psc_mutual_exclusion`: network and private_service_connect_config are mutually exclusive; use VPC peering or PSC, not both
- `dedicated_psc_mutual_exclusion`: dedicated_endpoint_enabled and private_service_connect_config are mutually exclusive; dedicated DNS is not available with PSC

## Outputs

Reference an output from another manifest as `valueFrom: {kind: GcpVertexAiEndpoint, name: <resource-name>, fieldPath: status.outputs.<output>}`.

| Output | Type | Description |
|---|---|---|
| `status.outputs.endpoint_id` | `string` | Fully qualified endpoint resource path. Format: projects/{project}/locations/{location}/endpoints/{name} |
| `status.outputs.display_name` | `string` | Display name of the endpoint (pass-through from spec for downstream reference without needing to read the full resource). |
| `status.outputs.dedicated_endpoint_dns` | `string` | DNS of the dedicated endpoint. Populated only when dedicated_endpoint_enabled is true. A bare hostname (no scheme), as the API returns it. Format: {endpointId}.{region}-{projectNumber}.prediction.vertexai.goog |
| `status.outputs.create_time` | `string` | RFC3339 timestamp of when the endpoint was created. |
| `status.outputs.endpoint_name` | `string` | The numeric endpoint ID (the last path segment of endpoint_id) -- explicit in the spec or derived from the resource identity. This is the value model-deployment tooling passes as the endpoint reference. |

## References

Fields that can point at another resource's outputs:

| Field | Kind | Output |
|---|---|---|
| `spec.projectId` | GcpProject | `status.outputs.project_id` |
| `spec.network` | GcpVpcNetwork | `status.outputs.network_self_link` |
| `spec.kmsKeyName` | GcpKmsKey | `status.outputs.key_id` |
| `spec.privateServiceConnectConfig.pscAutomationConfigs[].network` | GcpVpcNetwork | `status.outputs.network_self_link` |
| `spec.privateServiceConnectConfig.pscAutomationConfigs[].projectId` | GcpProject | `status.outputs.project_id` |

## Referenced By

Fields on other kinds that can point at this resource:

| Kind | Field | Reads |
|---|---|---|
| GcpPubSubSubscription | `spec.messageTransforms[].aiInference.endpoint` | `status.outputs.endpoint_id` |
| GcpPubSubTopic | `spec.messageTransforms[].aiInference.endpoint` | `status.outputs.endpoint_id` |

## See Also

- [Overview](../README.md)
