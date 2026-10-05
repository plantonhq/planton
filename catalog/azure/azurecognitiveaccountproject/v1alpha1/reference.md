# AzureCognitiveAccountProject

> Generated from the protobuf schema by `make generate-reference` -- do not
> edit by hand. To change a fact on this page, change the proto field comment
> or validation rule it is derived from, then regenerate.

**apiVersion**: `azure.planton.dev/v1alpha1`

**Guide**: [GUIDE.md](../GUIDE.md) -- authored operational judgment for this kind: conventions, trade-offs, and what pairs well with it.

**AzureCognitiveAccountProjectSpec** defines an AI Foundry project
on an Azure AI services account (ARM:
Microsoft.CognitiveServices/accounts/{account}/projects/{name}) --
the workspace a team organizes its AI work in: agents,
evaluations, files and data-plane assets live inside a project,
isolated from sibling projects on the same account.

**The parent account must allow projects**: an AzureCognitiveAccount
of kind "AIServices" with project_management_enabled true (which in
turn requires the account to carry a managed identity). The first
project created on an account becomes the account's DEFAULT project
(surfaced in the is_default output).

**The project is an ARM child of its account** -- it has no
resource group of its own (ARM derives it through the account),
though it does carry its own location, identity and tags.

**ForceNew fields**: `name` and `cognitive_account_id`. Also, ARM
cannot UPDATE `description` or `display_name` to an empty value --
clearing either replaces the project (setting or changing them is
an in-place update).

## Example

```yaml
apiVersion: azure.planton.dev/v1alpha1
kind: AzureCognitiveAccountProject
metadata:
  name: test-cognitive-account-project
spec:
  cognitiveAccountId:
    value: /subscriptions/00000000-0000-0000-0000-000000000000/resourceGroups/test-rg/providers/Microsoft.CognitiveServices/accounts/acme-foundry-test
  name: customer-support
  region: eastus
  identity:
    type: SYSTEM_ASSIGNED
  displayName: Customer Support
  description: Agents and evaluations for the customer-support team.
```

## Spec Fields

| Path | Type | Required | Default | References |
|---|---|---|---|---|
| `spec.cognitiveAccountId` | `string \| valueFrom` | yes |  | AzureCognitiveAccount (`status.outputs.cognitive_account_id`) |
| `spec.name` | `string` | yes |  |  |
| `spec.region` | `string` | yes |  |  |
| `spec.identity` | `AzureCognitiveAccountProjectIdentity` | yes |  |  |
| `spec.identity.type` | `enum` | yes |  |  |
| `spec.identity.identityIds` | `[]string \| valueFrom` |  |  | AzureUserAssignedIdentity (`status.outputs.identity_id`) |
| `spec.description` | `string` |  |  |  |
| `spec.displayName` | `string` |  |  |  |
| `spec.tags` | `map<string, string>` |  |  |  |

## Field Details

### spec.cognitiveAccountId

`string | valueFrom` · required

The Azure AI services account the project is created on, by ARM
ID. The account must be kind "AIServices" with
project_management_enabled true. Fixed at creation.

- references: AzureCognitiveAccount (`status.outputs.cognitive_account_id`)
- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: AzureCognitiveAccount, name: <that resource's name>, fieldPath: status.outputs.cognitive_account_id}} -- a bare string does not parse

### spec.name

`string` · required

The project's name, unique on the account: 2-64 characters,
starting alphanumeric, then alphanumerics, dashes, periods or
underscores (the provider's own rule). Changing the name replaces
the project.

- rule: {"required":true,"string":{"pattern":"^[a-zA-Z0-9][a-zA-Z0-9_.-]{1,63}$"}}

### spec.region

`string` · required

The Azure region the project lives in, e.g. "eastus". By
convention the account's own region.

- rule: {"required":true,"string":{"minLen":"1"}}

### spec.identity

`AzureCognitiveAccountProjectIdentity` · required

The project's managed identity. REQUIRED -- the provider's own
contract: every project carries an identity (it is what the
project's agents and evaluations act as).

- rule: {"required":true}
- rule: identity_ids is required for USER_ASSIGNED and SYSTEM_AND_USER_ASSIGNED and must be empty for SYSTEM_ASSIGNED

### spec.identity.type

`enum` · required

Identity flavor. SYSTEM_ASSIGNED is created and rotated by Azure
with the project; USER_ASSIGNED brings identities you manage;
SYSTEM_AND_USER_ASSIGNED carries both.

- rule: {"required":true}

Allowed values (use exactly as shown):

- `azure_cognitive_account_project_identity_type_unspecified` -- Not specified. Invalid -- the project requires an identity.
- `SYSTEM_ASSIGNED` -- Azure-managed identity created with the project.
- `USER_ASSIGNED` -- Identities you create and manage (AzureUserAssignedIdentity).
- `SYSTEM_AND_USER_ASSIGNED` -- Both a system-assigned and user-assigned identities.

### spec.identity.identityIds

`[]string | valueFrom`

For USER_ASSIGNED / SYSTEM_AND_USER_ASSIGNED: the user-assigned
identities attached to the project, by ARM ID.

- references: AzureUserAssignedIdentity (`status.outputs.identity_id`)
- rule: write as {value: <literal>} or {valueFrom: {kind: AzureUserAssignedIdentity, name: <that resource's name>, fieldPath: status.outputs.identity_id}} -- a bare string does not parse

### spec.description

`string`

What the project is for. ARM cannot update this to empty --
clearing a set description replaces the project.

### spec.displayName

`string`

The human-friendly name shown in the AI Foundry portal. ARM
cannot update this to empty -- clearing a set display name
replaces the project.

### spec.tags

`map<string, string>`

Free-form tags applied to the project, merged over the
Planton-derived resource tags (organization, environment,
resource id); a user tag with the same key wins.

## Outputs

Reference an output from another manifest as `valueFrom: {kind: AzureCognitiveAccountProject, name: <resource-name>, fieldPath: status.outputs.<output>}`.

| Output | Type | Description |
|---|---|---|
| `status.outputs.project_id` | `string` | The Azure Resource Manager ID of the project. Format: /subscriptions/{sub}/resourceGroups/{rg}/providers/Microsoft.CognitiveServices/accounts/{account}/projects/{name} |
| `status.outputs.project_name` | `string` | The project's name. |
| `status.outputs.endpoints` | `map<string, string>` | The project's data-plane endpoints, keyed by service label as ARM reports them (e.g. the AI Foundry API endpoint agents and SDKs call). |
| `status.outputs.is_default` | `bool` | Whether ARM made this the account's DEFAULT project (the first project created on an account becomes the default). |
| `status.outputs.system_assigned_identity_principal_id` | `string` | The principal (object) ID of the project's system-assigned identity, when one is enabled -- what data-source grants (storage, search) bind to. |

## References

Fields that can point at another resource's outputs:

| Field | Kind | Output |
|---|---|---|
| `spec.cognitiveAccountId` | AzureCognitiveAccount | `status.outputs.cognitive_account_id` |
| `spec.identity.identityIds` | AzureUserAssignedIdentity | `status.outputs.identity_id` |

## See Also

- [Overview](../README.md)
