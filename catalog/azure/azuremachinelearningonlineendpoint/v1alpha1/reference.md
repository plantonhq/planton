# AzureMachineLearningOnlineEndpoint

> Generated from the protobuf schema by `make generate-reference` -- do not
> edit by hand. To change a fact on this page, change the proto field comment
> or validation rule it is derived from, then regenerate.

**apiVersion**: `azure.planton.dev/v1alpha1`

**Guide**: [GUIDE.md](../GUIDE.md) -- authored operational judgment for this kind: conventions, trade-offs, and what pairs well with it.

**AzureMachineLearningOnlineEndpointSpec** defines a managed online
endpoint on an Azure Machine Learning workspace (ARM:
Microsoft.MachineLearningServices/workspaces/{ws}/onlineEndpoints/{name})
-- the stable HTTPS address applications call to score against
deployed models, with a traffic dial that splits requests across the
endpoint's deployments (blue/green rollouts, mirrored shadow
traffic).

**The contract here is the ARM specification itself** (pinned
api-version 2025-06-01): azurerm carries no resource for ML
endpoints, so both engines write the raw ARM shape and this spec's
validation rules carry the full contract burden -- there is no
provider-side schema behind them.

**The endpoint is an ARM child of its workspace** -- it has no
resource group of its own, and its name is reserved REGION-WIDE per
subscription (two endpoints in the same region cannot share a name
even across workspaces; the name becomes part of the scoring DNS).

## Example

```yaml
# Offline-plan test manifest. Exercises the deep seams: the AMLToken
# auth vocabulary, a dual system+user identity, both traffic maps, the
# public-network bool -> Enabled/Disabled wire mapping, bring-your-own
# initial keys through the write-only sensitive overlay, and the ARM
# property dictionary.
apiVersion: azure.planton.dev/v1alpha1
kind: AzureMachineLearningOnlineEndpoint
metadata:
  name: test-ml-online-endpoint
  org: test-org
  env: dev
spec:
  workspaceId:
    value: /subscriptions/00000000-0000-0000-0000-000000000000/resourceGroups/test-rg/providers/Microsoft.MachineLearningServices/workspaces/test-ml-workspace
  name: test-fraud-scoring
  region: eastus
  authMode: Key
  identity:
    type: SYSTEM_AND_USER_ASSIGNED
    identityIds:
      - value: /subscriptions/00000000-0000-0000-0000-000000000000/resourceGroups/test-rg/providers/Microsoft.ManagedIdentity/userAssignedIdentities/test-uai
  traffic:
    blue: 90
    green: 10
  mirrorTraffic:
    shadow: 25
  publicNetworkAccessEnabled: false
  initialAuthKeys:
    primaryKey:
      value: offline-plan-placeholder-primary-key
    secondaryKey:
      value: offline-plan-placeholder-secondary-key
  properties:
    team: ml-platform
  description: Offline-plan endpoint exercising the full surface
  tags:
    cost-center: ml-platform
```

## Spec Fields

| Path | Type | Required | Default | References |
|---|---|---|---|---|
| `spec.workspaceId` | `string \| valueFrom` | yes |  | AzureMachineLearningWorkspace (`status.outputs.machine_learning_workspace_id`) |
| `spec.name` | `string` | yes |  |  |
| `spec.region` | `string` | yes |  |  |
| `spec.authMode` | `string` | yes |  |  |
| `spec.identity` | `AzureMachineLearningOnlineEndpointIdentity` | yes |  |  |
| `spec.identity.type` | `enum` | yes |  |  |
| `spec.identity.identityIds` | `[]string \| valueFrom` |  |  | AzureUserAssignedIdentity (`status.outputs.identity_id`) |
| `spec.traffic` | `map<string, int32>` |  |  |  |
| `spec.mirrorTraffic` | `map<string, int32>` |  |  |  |
| `spec.publicNetworkAccessEnabled` | `bool` |  | `true` |  |
| `spec.initialAuthKeys` | `AzureMachineLearningOnlineEndpointInitialAuthKeys` |  |  |  |
| `spec.initialAuthKeys.primaryKey` | `string \| valueFrom` (sensitive) |  |  |  |
| `spec.initialAuthKeys.secondaryKey` | `string \| valueFrom` (sensitive) |  |  |  |
| `spec.properties` | `map<string, string>` |  |  |  |
| `spec.description` | `string` |  |  |  |
| `spec.tags` | `map<string, string>` |  |  |  |

## Field Details

### spec.workspaceId

`string | valueFrom` · required

The Machine Learning workspace the endpoint belongs to, by ARM ID.
Fixed at creation.

- references: AzureMachineLearningWorkspace (`status.outputs.machine_learning_workspace_id`)
- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: AzureMachineLearningWorkspace, name: <that resource's name>, fieldPath: status.outputs.machine_learning_workspace_id}} -- a bare string does not parse

### spec.name

`string` · required

The endpoint's name (ARM's own rule, mirrored from the pinned
specification: starts with a letter or digit, then letters,
digits, hyphens and underscores). The name is part of the scoring
URI's DNS and is reserved region-wide per subscription -- pick
something unique to the application. Changing the name replaces
the endpoint.

- rule: {"required":true,"string":{"pattern":"^[a-zA-Z0-9][a-zA-Z0-9-_]{0,254}$"}}

### spec.region

`string` · required

The Azure region the endpoint lives in, e.g. "eastus". Must be
the workspace's own region (the service provisions endpoints
beside their workspace). Fixed at creation.

- rule: {"required":true,"string":{"minLen":"1"}}

### spec.authMode

`string` · required

How callers authenticate scoring requests (the data-plane
operation), exactly the service's vocabulary: "Key" for static
keys that never expire, "AMLToken" for Azure ML tokens that
expire and refresh, "AADToken" for Microsoft Entra tokens (the
keyless posture). Control-plane access is always Entra ID
regardless of this setting.

- rule: {"required":true,"string":{"in":["Key","AMLToken","AADToken"]}}

### spec.identity

`AzureMachineLearningOnlineEndpointIdentity` · required

The endpoint's managed identity -- how its deployments pull
container images and model artifacts from the workspace's
registry and storage. The ARM specification marks identity
optional, but an endpoint without one cannot pull anything and
every deployment on it fails at provisioning -- this spec
requires it deliberately (a recorded tightening; SYSTEM_ASSIGNED
is the service's own default posture).

- rule: {"required":true}
- rule: identity_ids is required for USER_ASSIGNED and SYSTEM_AND_USER_ASSIGNED and must be empty for SYSTEM_ASSIGNED

### spec.identity.type

`enum` · required

Identity flavor. SYSTEM_ASSIGNED is created and rotated by Azure
with the endpoint (the service's default posture); USER_ASSIGNED
brings identities you manage (grantable registry/storage access
BEFORE the endpoint exists); SYSTEM_AND_USER_ASSIGNED carries
both.

- rule: {"required":true}

Allowed values (use exactly as shown):

- `azure_machine_learning_online_endpoint_identity_type_unspecified` -- Not specified: rejected -- the identity block requires a flavor.
- `SYSTEM_ASSIGNED` -- Azure-managed identity created with the endpoint.
- `USER_ASSIGNED` -- Identities you create and manage (AzureUserAssignedIdentity).
- `SYSTEM_AND_USER_ASSIGNED` -- Both a system-assigned and user-assigned identities.

### spec.identity.identityIds

`[]string | valueFrom`

For USER_ASSIGNED / SYSTEM_AND_USER_ASSIGNED: the user-assigned
identities attached to the endpoint, by ARM ID. Reference
AzureUserAssignedIdentity resources so registry / storage grants
can be composed before the endpoint is created.

- references: AzureUserAssignedIdentity (`status.outputs.identity_id`)
- rule: write as {value: <literal>} or {valueFrom: {kind: AzureUserAssignedIdentity, name: <that resource's name>, fieldPath: status.outputs.identity_id}} -- a bare string does not parse

### spec.traffic

`map<string, int32>`

Percentage of live scoring traffic sent to each deployment,
keyed by deployment name, e.g. {"blue": 100} or
{"blue": 90, "green": 10}. Values are whole percentages; ARM
enforces that they sum to at most 100 at apply time (a sum below
100 leaves the remainder unrouted). Updatable in place -- this is
the blue/green rollout dial.

- rule: {"map":{"values":{"int32":{"lte":100,"gte":0}}}}

### spec.mirrorTraffic

`map<string, int32>`

Percentage of live traffic MIRRORED to each deployment without
returning its responses to callers -- shadow-testing a new
deployment against production traffic. Keyed by deployment name;
ARM enforces a total of at most 50 at apply time. Updatable in
place.

- rule: {"map":{"values":{"int32":{"lte":50,"gte":0}}}}

### spec.publicNetworkAccessEnabled

`bool` · optional (explicit presence)

Whether the scoring endpoint answers the public internet.
Unspecified applies true (ARM's default "Enabled"). Set false to
require private-endpoint access to the workspace for scoring.
Updatable in place.

- default: `true`

### spec.initialAuthKeys

`AzureMachineLearningOnlineEndpointInitialAuthKeys`

Static authentication keys to SET at creation (Key auth mode
only) -- bring-your-own keys from a secret store instead of the
service minting them. ARM never returns these values on any read
(retrieval is a separate listKeys action), so leave this unset to
let the service mint keys and read them with
`az ml online-endpoint get-credentials`. Fixed at creation.

- rule: the initial_auth_keys block requires at least one of primary_key or secondary_key

### spec.initialAuthKeys.primaryKey

`string | valueFrom` · sensitive

The primary scoring key. Reference a secret rather than embedding
the literal.

- rule: write as {value: <literal>} or {valueFrom: {kind: <Kind>, name: <that resource's name>, fieldPath: status.outputs.<output>}} -- a bare string does not parse

### spec.initialAuthKeys.secondaryKey

`string | valueFrom` · sensitive

The secondary scoring key -- the rotation partner. Reference a
secret rather than embedding the literal.

- rule: write as {value: <literal>} or {valueFrom: {kind: <Kind>, name: <that resource's name>, fieldPath: status.outputs.<output>}} -- a bare string does not parse

### spec.properties

`map<string, string>`

The endpoint's ARM property dictionary -- free-form key/value
pairs some tooling reads. ARM allows ADDING entries but never
removing or altering existing ones (the service's own contract;
removals are ignored on update).

### spec.description

`string`

What the endpoint serves. Updatable in place.

### spec.tags

`map<string, string>`

Free-form tags applied to the endpoint, merged over the
Planton-derived resource tags (organization, environment,
resource id); a user tag with the same key wins. Updatable in
place.

## Outputs

Reference an output from another manifest as `valueFrom: {kind: AzureMachineLearningOnlineEndpoint, name: <resource-name>, fieldPath: status.outputs.<output>}`.

| Output | Type | Description |
|---|---|---|
| `status.outputs.online_endpoint_id` | `string` | The Azure Resource Manager ID of the online endpoint. Format: /subscriptions/{sub}/resourceGroups/{rg}/providers/Microsoft.MachineLearningServices/workspaces/{ws}/onlineEndpoints/{name} |
| `status.outputs.online_endpoint_name` | `string` | The endpoint's name -- what deployments attach to and what the traffic map's keys route across. |
| `status.outputs.scoring_uri` | `string` | The HTTPS address applications POST scoring requests to. |
| `status.outputs.swagger_uri` | `string` | The endpoint's OpenAPI (Swagger) document address -- the scoring request/response schema. |
| `status.outputs.system_assigned_identity_principal_id` | `string` | The principal (object) ID of the endpoint's system-assigned identity, when one is enabled -- what registry / storage grants bind to so deployments can pull images and models. |

## References

Fields that can point at another resource's outputs:

| Field | Kind | Output |
|---|---|---|
| `spec.workspaceId` | AzureMachineLearningWorkspace | `status.outputs.machine_learning_workspace_id` |
| `spec.identity.identityIds` | AzureUserAssignedIdentity | `status.outputs.identity_id` |

## Referenced By

Fields on other kinds that can point at this resource:

| Kind | Field | Reads |
|---|---|---|
| AzureMachineLearningOnlineDeployment | `spec.endpointId` | `status.outputs.online_endpoint_id` |

## See Also

- [Overview](../README.md)
