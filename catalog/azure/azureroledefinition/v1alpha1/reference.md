# AzureRoleDefinition

> Generated from the protobuf schema by `make generate-reference` -- do not
> edit by hand. To change a fact on this page, change the proto field comment
> or validation rule it is derived from, then regenerate.

**apiVersion**: `azure.planton.dev/v1alpha1`

**AzureRoleDefinitionSpec** defines the configuration for creating a custom
Azure RBAC role: a named, reusable bundle of permissions that principals can
then be granted through role assignments.

Azure ships hundreds of built-in roles, but real organizations routinely
need permission sets the built-ins don't express -- "Contributor, except
role assignments and policy writes", "can restart VMs but not create or
delete them", "read-only on everything plus read access to storage blob
data". A custom role definition captures such a set once, with a meaningful
name, and every grant of it stays consistent as the definition evolves:
updating the definition's permissions updates what every existing
assignment of it allows.

The anatomy of a custom role:
- **permissions**: WHAT the role allows, as ARM operation patterns split
  across control plane (actions/not_actions -- managing resources) and data
  plane (data_actions/not_data_actions -- accessing the data inside them,
  like blob contents or queue messages).
- **scope**: WHERE the definition itself lives (management group,
  subscription, or resource group). This anchors the definition's ARM
  resource ID and, by default, where it can be assigned.
- **assignable_scopes**: WHERE assignments of this role may be created.
  Defaults to the definition's own scope when omitted.

A definition grants nothing by itself -- permissions only take effect when
an assignment binds the role to a principal at a scope. Compose this kind
with AzureRoleAssignment: the definition's role_definition_id output is
exactly what an assignment's role_definition_id field consumes.

Creating or updating role definitions requires
`Microsoft.Authorization/roleDefinitions/write` on the target scope --
held via Owner, User Access Administrator, or Role Based Access Control
Administrator. A tenant can hold at most 5,000 custom roles (500 in Azure
Government/China clouds); definitions are metadata and cost nothing.

## Example

```yaml
apiVersion: azure.planton.dev/v1alpha1
kind: AzureRoleDefinition
metadata:
  name: test-role-definition
  org: test-org
  env: dev
spec:
  name: test-org-vm-operator
  scope:
    value: /subscriptions/00000000-0000-0000-0000-000000000000
  description: Operate existing VMs (start/stop/restart) without create or delete rights
  permissions:
    - actions:
        - Microsoft.Compute/virtualMachines/read
        - Microsoft.Compute/virtualMachines/start/action
        - Microsoft.Compute/virtualMachines/restart/action
        - Microsoft.Compute/virtualMachines/deallocate/action
  assignableScopes:
    - value: /subscriptions/00000000-0000-0000-0000-000000000000
```

## Spec Fields

| Path | Type | Required | Default | References |
|---|---|---|---|---|
| `spec.name` | `string` | yes |  |  |
| `spec.scope` | `string \| valueFrom` | yes |  | AzureResourceGroup (`status.outputs.resource_group_id`) |
| `spec.description` | `string` |  |  |  |
| `spec.permissions` | `[]AzureRoleDefinitionPermission` |  |  |  |
| `spec.permissions[].actions` | `[]string` |  |  |  |
| `spec.permissions[].notActions` | `[]string` |  |  |  |
| `spec.permissions[].dataActions` | `[]string` |  |  |  |
| `spec.permissions[].notDataActions` | `[]string` |  |  |  |
| `spec.assignableScopes` | `[]string \| valueFrom` |  |  | AzureResourceGroup (`status.outputs.resource_group_id`) |
| `spec.roleDefinitionId` | `string` |  |  |  |

## Field Details

### spec.name

`string` · required

The role's display name -- what operators see in the portal's role picker
and what role assignments can reference by name. Must be unique within
the Azure AD tenant (Azure rejects a duplicate at create time), so prefer
org-prefixed names like "acme-vm-operator" over generic ones like
"vm-operator" that another team may already have taken. Renaming is an
in-place update; assignments track the role by GUID, not by name.

- rule: {"required":true,"string":{"minLen":"1"}}

### spec.scope

`string | valueFrom` · required

The scope the role definition is created at, which anchors its ARM
resource ID. Valid creation scopes:
- Management group: "/providers/Microsoft.Management/managementGroups/{name}"
- Subscription:     "/subscriptions/{subscription-id}"
- Resource group:   "/subscriptions/{sub}/resourceGroups/{rg-name}"

Choose the highest scope the role will ever be assigned at: a definition
is only assignable at or below the scopes in assignable_scopes, and those
must be within reach of where the definition lives. Subscription scope is
the common choice for org-wide roles; a resource-group scope keeps a
team-specific role invisible outside that group's blast radius.

Changing the scope replaces the definition (delete + create) -- the scope
is part of the definition's ARM identity.

Defaults to referencing an AzureResourceGroup's ARM ID for composed
environments; subscription or management-group scopes pass as literal
values, e.g.:
  scope:
    value: /subscriptions/00000000-0000-0000-0000-000000000000

- references: AzureResourceGroup (`status.outputs.resource_group_id`)
- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: AzureResourceGroup, name: <that resource's name>, fieldPath: status.outputs.resource_group_id}} -- a bare string does not parse

### spec.description

`string`

Free-text description shown beside the role in the portal's role picker
and returned by the authorization API. Use it to state the role's intent
and boundaries ("Operate existing VMs: start/stop/restart, no create or
delete") -- operators choosing between roles see only the name and this
text. Updatable in place.

### spec.permissions

`[]AzureRoleDefinitionPermission`

The permission blocks defining what this role allows. Azure evaluates
them as a union: an operation is permitted if any block's actions (minus
its not_actions) match it. One block is the norm; multiple blocks are
legal and merely additive -- ARM models permissions as a list, so this
does too.

A definition with no permission blocks is legal (Azure accepts it) and
grants nothing -- useful only as a placeholder being filled in
incrementally. Every real role carries at least one block.

### spec.permissions[].actions

`[]string`

Control-plane operations this role allows, e.g.
"Microsoft.Compute/virtualMachines/start/action" or the wildcard "*"
(everything -- the built-in Contributor's base). Matched against ARM
operations; grants no data-plane access.

### spec.permissions[].notActions

`[]string`

Control-plane operations subtracted from actions. NOT a deny rule: a
not_action only trims THIS role's grant, it cannot take away permissions
another assignment gives the same principal (that is Azure deny
assignments' job, a separate mechanism). The classic use is broad-grant
-minus-carve-out: actions ["*"], not_actions
["Microsoft.Authorization/*/write"] -- everything except changing RBAC.

### spec.permissions[].dataActions

`[]string`

Data-plane operations this role allows, e.g.
"Microsoft.Storage/storageAccounts/blobServices/containers/blobs/read".
Required for principals that access data inside resources -- reading
blobs, receiving queue messages, reading Key Vault secret values (in
RBAC mode) -- none of which any control-plane action grants.

### spec.permissions[].notDataActions

`[]string`

Data-plane operations subtracted from data_actions. Same trimming
semantics as not_actions: a carve-out from this role's own grant, not a
deny. Example: data_actions
["Microsoft.Storage/storageAccounts/blobServices/containers/blobs/*"]
with not_data_actions [".../blobs/delete"] -- full blob access except
deletion.

### spec.assignableScopes

`[]string | valueFrom`

The scopes at which this role may be assigned -- management groups,
subscriptions, resource groups, or individual resources. An assignment
whose scope is not at or under one of these is rejected by Azure.

When omitted, Azure defaults it to the definition's own scope (both IaC
engines inherit this defaulting from the provider), which is the right
call for most roles. Narrow it below the creation scope to pre-authorize
where a role may ever be granted -- e.g. a subscription-scoped definition
assignable only within two project resource groups. Note Azure's own
constraint: at most ONE management group may appear in this list.

Each entry is a literal ARM ID or a reference to a resource's ID output
(defaults to an AzureResourceGroup's ARM ID); literals and references
can mix freely. Updatable in place. Where a role MAY BE ASSIGNED is not
where the definition lives (that is `scope`), so on a diagram these
references are access, not placement.

- references: AzureResourceGroup (`status.outputs.resource_group_id`)
- rule: write as {value: <literal>} or {valueFrom: {kind: AzureResourceGroup, name: <that resource's name>, fieldPath: status.outputs.resource_group_id}} -- a bare string does not parse

### spec.roleDefinitionId

`string`

A stable UUID for the role definition's ARM resource name. Azure
identifies a role definition by a GUID; when omitted (recommended), a
random one is generated at deploy time. Pin it only when an
externally-known GUID must be preserved -- e.g. recreating a definition
that existing assignments or tooling reference by its full ARM ID.
Changing a pinned GUID replaces the definition.

- rule: {"ignore":"IGNORE_IF_ZERO_VALUE","string":{"uuid":true}}

## Outputs

Reference an output from another manifest as `valueFrom: {kind: AzureRoleDefinition, name: <resource-name>, fieldPath: status.outputs.<output>}`.

| Output | Type | Description |
|---|---|---|
| `status.outputs.role_definition_id` | `string` | The fully-scoped Azure Resource Manager ID of the role definition. Format: {scope}/providers/Microsoft.Authorization/roleDefinitions/{guid} This is the value an AzureRoleAssignment's role_definition_id field consumes -- reference it to grant this custom role to a principal. |
| `status.outputs.role_definition_guid` | `string` | The definition's GUID resource name -- either the pinned spec.role_definition_id or the GUID generated at deploy time. With the scope, it uniquely identifies the definition in the authorization API. |
| `status.outputs.role_name` | `string` | The role's display name as deployed (the tenant-unique name operators see in the portal's role picker). |
| `status.outputs.scope` | `string` | The scope the definition was created at, as resolved at deploy time. |
| `status.outputs.assignable_scopes` | `[]string` | The assignable scopes Azure recorded for the definition. Carries what was actually applied -- when the spec omitted assignable_scopes, this holds the provider-defaulted value (the definition's own scope). |

## References

Fields that can point at another resource's outputs:

| Field | Kind | Output |
|---|---|---|
| `spec.scope` | AzureResourceGroup | `status.outputs.resource_group_id` |
| `spec.assignableScopes` | AzureResourceGroup | `status.outputs.resource_group_id` |

## Referenced By

Fields on other kinds that can point at this resource:

| Kind | Field | Reads |
|---|---|---|
| AzureMonitorActionGroup | `spec.armRoleReceivers[].roleId` | `status.outputs.role_definition_guid` |
| AzureRoleAssignment | `spec.roleDefinitionId` | `status.outputs.role_definition_id` |

## See Also

- [Overview](../README.md)
