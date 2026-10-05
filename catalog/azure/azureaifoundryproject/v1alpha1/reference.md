# AzureAiFoundryProject

> Generated from the protobuf schema by `make generate-reference` -- do not
> edit by hand. To change a fact on this page, change the proto field comment
> or validation rule it is derived from, then regenerate.

**apiVersion**: `azure.planton.dev/v1alpha1`

**Guide**: [GUIDE.md](../GUIDE.md) -- authored operational judgment for this kind: conventions, trade-offs, and what pairs well with it.

**AzureAiFoundryProjectSpec** defines an Azure AI Foundry project
(ARM: Microsoft.MachineLearningServices/workspaces with kind
"Project") -- the workspace one AI team works in, created inside an
AzureAiFoundry hub. The project INHERITS the hub's posture (key
vault, storage, insights, registry, managed network, encryption)
and carries only its own identity, naming, and description -- which
is why this spec has no companion-service fields.

**The project deploys into the HUB's resource group** (the
provider derives it from the hub reference -- there is no
resource-group field here), and the hub linkage is fixed at
creation.

## Example

```yaml
# Offline-plan test manifest. Exercises the deep seams: the typed hub
# reference, a system-and-user-assigned identity with a primary
# identity, the high-business-impact flag, and the descriptive
# surface. There is deliberately NO resource group in the spec -- the
# project deploys into its hub's group (the provider derives it).
apiVersion: azure.planton.dev/v1alpha1
kind: AzureAiFoundryProject
metadata:
  name: test-ai-foundry-project
  org: test-org
  env: dev
spec:
  region: eastus
  name: test-aif-project
  aiServicesHubId:
    value: /subscriptions/00000000-0000-0000-0000-000000000000/resourceGroups/test-rg/providers/Microsoft.MachineLearningServices/workspaces/test-ai-foundry
  identity:
    type: SYSTEM_AND_USER_ASSIGNED
    identityIds:
      - value: /subscriptions/00000000-0000-0000-0000-000000000000/resourceGroups/test-rg/providers/Microsoft.ManagedIdentity/userAssignedIdentities/aifp-uai
  primaryUserAssignedIdentity:
    value: /subscriptions/00000000-0000-0000-0000-000000000000/resourceGroups/test-rg/providers/Microsoft.ManagedIdentity/userAssignedIdentities/aifp-uai
  highBusinessImpactEnabled: true
  description: Offline-plan Foundry project exercising the deep seams
  friendlyName: Test Foundry Project
  tags:
    cost-center: ai-platform
```

## Spec Fields

| Path | Type | Required | Default | References |
|---|---|---|---|---|
| `spec.region` | `string` | yes |  |  |
| `spec.name` | `string` | yes |  |  |
| `spec.aiServicesHubId` | `string \| valueFrom` | yes |  | AzureAiFoundry (`status.outputs.ai_foundry_id`) |
| `spec.identity` | `AzureAiFoundryProjectIdentity` |  |  |  |
| `spec.identity.type` | `enum` | yes |  |  |
| `spec.identity.identityIds` | `[]string \| valueFrom` |  |  | AzureUserAssignedIdentity (`status.outputs.identity_id`) |
| `spec.primaryUserAssignedIdentity` | `string \| valueFrom` |  |  | AzureUserAssignedIdentity (`status.outputs.identity_id`) |
| `spec.highBusinessImpactEnabled` | `bool` |  |  |  |
| `spec.description` | `string` |  |  |  |
| `spec.friendlyName` | `string` |  |  |  |
| `spec.tags` | `map<string, string>` |  |  |  |

## Field Details

### spec.region

`string` · required

The Azure region the project lives in, e.g. "eastus" --
typically the hub's region. Changing the region replaces the
project.

- rule: {"required":true,"string":{"minLen":"1"}}

### spec.name

`string` · required

The project's name: 3-33 characters, starting with an
alphanumeric character, then alphanumerics, hyphens or
underscores (the provider's own code regex -- its error text
understates the length by one and omits the underscore).
Changing the name replaces the project.

- rule: {"required":true,"string":{"pattern":"^[a-zA-Z0-9][a-zA-Z0-9_-]{2,32}$"}}

### spec.aiServicesHubId

`string | valueFrom` · required

The AzureAiFoundry hub the project is created in, by ARM ID.
Fixed at creation -- a project cannot move between hubs. The
project deploys into this hub's resource group.

- references: AzureAiFoundry (`status.outputs.ai_foundry_id`)
- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: AzureAiFoundry, name: <that resource's name>, fieldPath: status.outputs.ai_foundry_id}} -- a bare string does not parse

### spec.identity

`AzureAiFoundryProjectIdentity`

The project's managed identity. Optional -- most projects use
SYSTEM_ASSIGNED so per-project grants can be scoped to the team.

- rule: identity_ids is required for USER_ASSIGNED and SYSTEM_AND_USER_ASSIGNED and must be empty for SYSTEM_ASSIGNED

### spec.identity.type

`enum` · required

Identity flavor. SYSTEM_ASSIGNED is created and rotated by Azure
with the project; USER_ASSIGNED brings identities you manage --
pair with primary_user_assigned_identity;
SYSTEM_AND_USER_ASSIGNED carries both.

- rule: {"required":true}

Allowed values (use exactly as shown):

- `azure_ai_foundry_project_identity_type_unspecified` -- Not specified: rejected when the identity block is present.
- `SYSTEM_ASSIGNED` -- Azure-managed identity created with the project.
- `USER_ASSIGNED` -- Identities you create and manage (AzureUserAssignedIdentity).
- `SYSTEM_AND_USER_ASSIGNED` -- Both a system-assigned and user-assigned identities.

### spec.identity.identityIds

`[]string | valueFrom`

For USER_ASSIGNED / SYSTEM_AND_USER_ASSIGNED: the user-assigned
identities attached to the project, by ARM ID.

- references: AzureUserAssignedIdentity (`status.outputs.identity_id`)
- rule: write as {value: <literal>} or {valueFrom: {kind: AzureUserAssignedIdentity, name: <that resource's name>, fieldPath: status.outputs.identity_id}} -- a bare string does not parse

### spec.primaryUserAssignedIdentity

`string | valueFrom`

For projects with a user-assigned identity: which of the
attached identities the project uses as its primary, by ARM ID.
Only legal alongside the identity block (the provider's own
pairing, front-loaded below).

- references: AzureUserAssignedIdentity (`status.outputs.identity_id`)
- rule: write as {value: <literal>} or {valueFrom: {kind: AzureUserAssignedIdentity, name: <that resource's name>, fieldPath: status.outputs.identity_id}} -- a bare string does not parse

### spec.highBusinessImpactEnabled

`bool`

Marks the project as handling sensitive business data ("high
business impact") -- Azure reduces the diagnostic data it
collects. Fixed at creation. False means "leave it to the
service": the modules omit the property when false because the
service flips it true when hub encryption applies, and a pinned
false would fight that (the provider reads the value back).

### spec.description

`string`

What the project is for.

### spec.friendlyName

`string`

The display name shown in the Azure AI Foundry portal.

### spec.tags

`map<string, string>`

Free-form tags applied to the project, merged over the
Planton-derived resource tags (organization, environment,
resource id); a user tag with the same key wins.

## Validation Rules

- `primary_identity_requires_identity_block`: primary_user_assigned_identity requires the identity block -- add identity with type USER_ASSIGNED or SYSTEM_AND_USER_ASSIGNED

## Outputs

Reference an output from another manifest as `valueFrom: {kind: AzureAiFoundryProject, name: <resource-name>, fieldPath: status.outputs.<output>}`.

| Output | Type | Description |
|---|---|---|
| `status.outputs.ai_foundry_project_id` | `string` | The Azure Resource Manager ID of the project (projects are ARM workspaces, like their hub). Format: /subscriptions/{sub}/resourceGroups/{rg}/providers/Microsoft.MachineLearningServices/workspaces/{name} |
| `status.outputs.ai_foundry_project_name` | `string` | The name of the project. |
| `status.outputs.project_guid` | `string` | The project's immutable GUID (distinct from the ARM ID) -- what Foundry SDKs and data-plane calls identify the project by. |
| `status.outputs.system_assigned_identity_principal_id` | `string` | The principal (object) ID of the project's system-assigned identity, when one is enabled -- what per-team grants bind to. |

## References

Fields that can point at another resource's outputs:

| Field | Kind | Output |
|---|---|---|
| `spec.aiServicesHubId` | AzureAiFoundry | `status.outputs.ai_foundry_id` |
| `spec.identity.identityIds` | AzureUserAssignedIdentity | `status.outputs.identity_id` |
| `spec.primaryUserAssignedIdentity` | AzureUserAssignedIdentity | `status.outputs.identity_id` |

## See Also

- [Overview](../README.md)
