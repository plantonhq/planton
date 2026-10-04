# AzureMachineLearningBatchEndpoint

> Generated from the protobuf schema by `make generate-reference` -- do not
> edit by hand. To change a fact on this page, change the proto field comment
> or validation rule it is derived from, then regenerate.

**apiVersion**: `azure.planton.dev/v1alpha1`

**Guide**: [GUIDE.md](../GUIDE.md) -- authored operational judgment for this kind: conventions, trade-offs, and what pairs well with it.

**AzureMachineLearningBatchEndpointSpec** defines a batch endpoint
on an Azure Machine Learning workspace (ARM:
Microsoft.MachineLearningServices/workspaces/{ws}/batchEndpoints/{name})
-- the stable address batch scoring jobs are submitted to. Invoking
the endpoint creates a JOB that runs a deployment's recipe over a
large input (files or tabular data) on pooled compute; nothing runs
or bills while no job is active.

**The contract here is the ARM specification itself** (pinned
api-version 2025-06-01): azurerm carries no resource for ML
endpoints, so both engines write the raw ARM shape and this spec's
validation rules carry the full contract burden -- there is no
provider-side schema behind them.

**The endpoint is an ARM child of its workspace** -- it has no
resource group of its own, and its name becomes part of the batch
scoring DNS (expect the same region-wide-per-subscription
reservation the online endpoint documents).

## Example

```yaml
# Offline-plan test manifest. Exercises the deep seams: the explicit
# AADToken auth value (the vocabulary's only member -- the batch
# service rejects everything else), a dual system+user identity through
# the optional identity block, the default-deployment pointer riding
# ARM's defaults object, and the ARM property dictionary.
apiVersion: azure.planton.dev/v1alpha1
kind: AzureMachineLearningBatchEndpoint
metadata:
  name: test-ml-batch-endpoint
  org: test-org
  env: dev
spec:
  workspaceId:
    value: /subscriptions/00000000-0000-0000-0000-000000000000/resourceGroups/test-rg/providers/Microsoft.MachineLearningServices/workspaces/test-ml-workspace
  name: test-nightly-scoring
  region: eastus
  authMode: AADToken
  identity:
    type: SYSTEM_AND_USER_ASSIGNED
    identityIds:
      - value: /subscriptions/00000000-0000-0000-0000-000000000000/resourceGroups/test-rg/providers/Microsoft.ManagedIdentity/userAssignedIdentities/test-uai
  defaultDeploymentName: production
  properties:
    team: ml-platform
  description: Offline-plan batch endpoint exercising the full surface
  tags:
    cost-center: ml-platform
```

## Spec Fields

| Path | Type | Required | Default | References |
|---|---|---|---|---|
| `spec.workspaceId` | `string \| valueFrom` | yes |  | AzureMachineLearningWorkspace (`status.outputs.machine_learning_workspace_id`) |
| `spec.name` | `string` | yes |  |  |
| `spec.region` | `string` | yes |  |  |
| `spec.authMode` | `string` |  |  |  |
| `spec.identity` | `AzureMachineLearningBatchEndpointIdentity` |  |  |  |
| `spec.identity.type` | `enum` | yes |  |  |
| `spec.identity.identityIds` | `[]string \| valueFrom` |  |  | AzureUserAssignedIdentity (`status.outputs.identity_id`) |
| `spec.defaultDeploymentName` | `string` |  |  |  |
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
digits, hyphens and underscores). The name is part of the batch
scoring URI's DNS -- pick something unique to the application.
Changing the name replaces the endpoint.

- rule: {"required":true,"string":{"pattern":"^[a-zA-Z0-9][a-zA-Z0-9-_]{0,254}$"}}

### spec.region

`string` · required

The Azure region the endpoint lives in, e.g. "eastus". Must be
the workspace's own region (the service provisions endpoints
beside their workspace). Fixed at creation.

- rule: {"required":true,"string":{"minLen":"1"}}

### spec.authMode

`string`

How callers authenticate job submissions (the data-plane
operation). Unspecified applies "AADToken" -- Microsoft Entra
tokens, the ONLY mode the batch service accepts: the shared ARM
enum also advertises "Key" and "AMLToken", but the service
rejects both with "AuthMode must be 'AADToken'" (unlike online
endpoints, where all three work). The vocabulary here encodes
the service's real contract and extends additively if the
service ever honors the wider enum.

- rule: {"ignore":"IGNORE_IF_ZERO_VALUE","string":{"in":["AADToken"]}}

### spec.identity

`AzureMachineLearningBatchEndpointIdentity`

The endpoint's managed identity. OPTIONAL here where the online
endpoint sibling requires one -- deliberately, because the
rationale does not transfer: batch deployments provision nothing
at create, and jobs run under the INVOKER's Entra token plus the
COMPUTE cluster's managed identity, so the endpoint's identity
sits outside the batch data path. Set one when organizational
convention grants roles to the endpoint object itself.

- rule: identity_ids is required for USER_ASSIGNED and SYSTEM_AND_USER_ASSIGNED and must be empty for SYSTEM_ASSIGNED

### spec.identity.type

`enum` · required

Identity flavor. SYSTEM_ASSIGNED is created and rotated by Azure
with the endpoint; USER_ASSIGNED brings identities you manage;
SYSTEM_AND_USER_ASSIGNED carries both.

- rule: {"required":true}

Allowed values (use exactly as shown):

- `azure_machine_learning_batch_endpoint_identity_type_unspecified` -- Not specified: rejected -- the identity block requires a flavor.
- `SYSTEM_ASSIGNED` -- Azure-managed identity created with the endpoint.
- `USER_ASSIGNED` -- Identities you create and manage (AzureUserAssignedIdentity).
- `SYSTEM_AND_USER_ASSIGNED` -- Both a system-assigned and user-assigned identities.

### spec.identity.identityIds

`[]string | valueFrom`

For USER_ASSIGNED / SYSTEM_AND_USER_ASSIGNED: the user-assigned
identities attached to the endpoint, by ARM ID. Reference
AzureUserAssignedIdentity resources so role grants can be
composed before the endpoint is created.

- references: AzureUserAssignedIdentity (`status.outputs.identity_id`)
- rule: write as {value: <literal>} or {valueFrom: {kind: AzureUserAssignedIdentity, name: <that resource's name>, fieldPath: status.outputs.identity_id}} -- a bare string does not parse

### spec.defaultDeploymentName

`string`

The deployment that answers job submissions which do not name
one -- batch's analog of the online endpoint's traffic dial
(the named deployment receives 100% of unrouted submissions).
Names a deployment that usually does not exist yet when the
endpoint is first created: create the endpoint, add deployments,
then update this pointer (updatable in place via full PUT).

- rule: {"ignore":"IGNORE_IF_ZERO_VALUE","string":{"pattern":"^[a-zA-Z0-9][a-zA-Z0-9-_]{0,254}$"}}

### spec.properties

`map<string, string>`

The endpoint's ARM property dictionary -- free-form key/value
pairs some tooling reads. ARM allows ADDING entries but never
removing or altering existing ones (the service's own contract;
removals are ignored on update).

NOTE: unlike the online endpoint, the batch surface carries NO
publicNetworkAccess property -- network reachability of batch
scoring is governed by the WORKSPACE's own network settings.

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

Reference an output from another manifest as `valueFrom: {kind: AzureMachineLearningBatchEndpoint, name: <resource-name>, fieldPath: status.outputs.<output>}`.

| Output | Type | Description |
|---|---|---|
| `status.outputs.batch_endpoint_id` | `string` | The Azure Resource Manager ID of the batch endpoint. Format: /subscriptions/{sub}/resourceGroups/{rg}/providers/Microsoft.MachineLearningServices/workspaces/{ws}/batchEndpoints/{name} |
| `status.outputs.batch_endpoint_name` | `string` | The endpoint's name -- what deployments attach to and what the default-deployment pointer routes submissions across. |
| `status.outputs.scoring_uri` | `string` | The HTTPS address batch scoring jobs are submitted to (with a Microsoft Entra token). |
| `status.outputs.swagger_uri` | `string` | The endpoint's OpenAPI (Swagger) document address -- the job submission request/response schema. EMPTY until a deployment serves behind the endpoint (live-measured on both engines: ARM returns "" for a deployment-less endpoint, unlike the online sibling whose swagger URI populates at create). |
| `status.outputs.system_assigned_identity_principal_id` | `string` | The principal (object) ID of the endpoint's system-assigned identity, when one is enabled -- what role grants bind to when organizational convention grants roles to the endpoint object. |

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
| AzureMachineLearningBatchDeployment | `spec.endpointId` | `status.outputs.batch_endpoint_id` |

## See Also

- [Overview](../README.md)
