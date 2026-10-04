# AzureRoleAssignment

> Generated from the protobuf schema by `make generate-reference` -- do not
> edit by hand. To change a fact on this page, change the proto field comment
> or validation rule it is derived from, then regenerate.

**apiVersion**: `azure.planton.dev/v1alpha1`

**AzureRoleAssignmentSpec** defines the configuration for creating an Azure
RBAC role assignment: the grant of a role to a principal at a scope.

A role assignment is the atomic unit of authorization in Azure. Everything a
user, group, service principal, or managed identity is allowed to do is the
sum of the role assignments that target it. Because grants are the
most-repeated pattern in any Azure environment -- every identity needs
permissions on the resources it touches -- this kind models them as
first-class, composable nodes: one assignment per resource, referenceable in
infra charts, with an independent lifecycle from both the principal and the
scope it binds.

The three coordinates of every assignment:
- **scope**: WHERE the permission applies (management group, subscription,
  resource group, or an individual resource). Permissions inherit downward --
  a role assigned at a resource group applies to every resource in it.
- **role**: WHAT is permitted, referenced either by built-in role name
  (e.g. "Reader") or by role definition ID (custom roles).
- **principal_id**: WHO receives the permission (the Azure AD object ID of a
  user, group, service principal, or managed identity).

Azure role assignments are immutable: changing any field replaces the
assignment (delete + create). This matches ARM's own model, where an
assignment is an atomic grant record rather than a mutable object.

Creating role assignments requires the caller to hold
`Microsoft.Authorization/roleAssignments/write` on the target scope --
typically via the Owner, User Access Administrator, or Role Based Access
Control Administrator role. A deploy failing with "AuthorizationFailed"
almost always means the deploying credential has a data-plane or
Contributor-level role but no authorization-plane rights at that scope.

## Example

```yaml
apiVersion: azure.planton.dev/v1alpha1
kind: AzureRoleAssignment
metadata:
  name: test-role-assignment
  org: test-org
  env: dev
spec:
  scope:
    value: /subscriptions/00000000-0000-0000-0000-000000000000/resourceGroups/test-platform-rg
  roleDefinitionName: Reader
  principalId:
    value: 11111111-1111-1111-1111-111111111111
```

## Spec Fields

| Path | Type | Required | Default | References |
|---|---|---|---|---|
| `spec.scope` | `string \| valueFrom` | yes |  | AzureResourceGroup (`status.outputs.resource_group_id`) |
| `spec.roleDefinitionName` | `string` |  |  |  |
| `spec.roleDefinitionId` | `string \| valueFrom` |  |  | AzureRoleDefinition (`status.outputs.role_definition_id`) |
| `spec.principalId` | `string \| valueFrom` | yes |  | AzureUserAssignedIdentity (`status.outputs.principal_id`) |
| `spec.principalType` | `enum` |  |  |  |
| `spec.description` | `string` |  |  |  |
| `spec.condition` | `string` |  |  |  |
| `spec.conditionVersion` | `string` |  |  |  |
| `spec.delegatedManagedIdentityResourceId` | `string` |  |  |  |
| `spec.skipServicePrincipalAadCheck` | `bool` |  |  |  |
| `spec.name` | `string` |  |  |  |

## Field Details

### spec.scope

`string | valueFrom` · required

The scope the role assignment applies at. Any ARM ID is a valid scope:
- Management group: "/providers/Microsoft.Management/managementGroups/{name}"
- Subscription:     "/subscriptions/{subscription-id}"
- Resource group:   "/subscriptions/{sub}/resourceGroups/{rg-name}"
- Single resource:  ".../providers/Microsoft.KeyVault/vaults/{vault-name}"

Permissions granted at a scope inherit downward to everything under it,
so prefer the narrowest scope that satisfies the use case
(least privilege): grant on the specific vault, not the subscription.

Defaults to referencing an AzureResourceGroup's ARM ID -- the most common
grant boundary in composed environments. The scope is genuinely
polymorphic, so any other resource's ID output works via an explicit
valueFrom (kind + fieldPath override the default annotation), e.g.:
  scope:
    valueFrom:
      kind: AzureKeyVault
      name: platform-kv
      fieldPath: status.outputs.key_vault_id

- references: AzureResourceGroup (`status.outputs.resource_group_id`)
- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: AzureResourceGroup, name: <that resource's name>, fieldPath: status.outputs.resource_group_id}} -- a bare string does not parse

### spec.roleDefinitionName

`string`

The name of a built-in Azure role to assign, matched case-insensitively
against Azure's role catalog at the target scope.

Common built-in roles:
- "Reader" -- read-only access to everything at the scope
- "Contributor" -- manage everything except RBAC itself
- "Owner" -- manage everything including RBAC
- "Key Vault Secrets User" -- read Key Vault secret values (RBAC mode)
- "Storage Blob Data Contributor" -- read/write/delete blob data
- "AcrPull" / "AcrPush" -- pull/push container images
- "Network Contributor" -- manage networking resources
- "Monitoring Metrics Publisher" -- publish custom metrics

Use role_definition_id instead when assigning a custom role.

### spec.roleDefinitionId

`string | valueFrom`

The fully-scoped resource ID of the role definition to assign. This is the
way to bind custom roles -- reference an AzureRoleDefinition's
role_definition_id output (in a composed environment the reference is
also the deploy-ordering edge, so the custom role exists before the
grant that binds it), or pass a literal ID for an existing definition;
it also works for built-ins when the caller already holds the ID.
Format:
"/subscriptions/{sub}/providers/Microsoft.Authorization/roleDefinitions/{guid}"
(or "/providers/Microsoft.Authorization/roleDefinitions/{guid}" for
tenant-level built-in definitions).

- references: AzureRoleDefinition (`status.outputs.role_definition_id`)
- rule: write as {value: <literal>} or {valueFrom: {kind: AzureRoleDefinition, name: <that resource's name>, fieldPath: status.outputs.role_definition_id}} -- a bare string does not parse

### spec.principalId

`string | valueFrom` · required

The Azure AD object ID of the principal receiving the role. This is the
OBJECT ID (also called principal ID) -- not the application (client) ID.
Confusing the two is the most common role-assignment mistake: the
assignment succeeds with a client ID but grants nothing, because no
directory object has that object ID.

Defaults to referencing an AzureUserAssignedIdentity's principal_id
output -- the dominant composition on Planton (grant a managed identity
access to the resources it operates on). For users, groups, or externally
managed service principals, pass the object ID as a literal value.

- references: AzureUserAssignedIdentity (`status.outputs.principal_id`)
- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: AzureUserAssignedIdentity, name: <that resource's name>, fieldPath: status.outputs.principal_id}} -- a bare string does not parse

### spec.principalType

`enum`

The type of the principal being granted the role. Optional: Azure infers
it from the directory when omitted. Set it explicitly when the deploying
credential is constrained by ABAC rules that filter on principal type
(Azure then requires the type to be declared on the request).

Allowed values (use exactly as shown):

- `azure_role_assignment_principal_type_unspecified` -- Not specified: Azure infers the principal type from the directory object.
- `SERVICE_PRINCIPAL` -- A service principal or managed identity (managed identities ARE service principals in the directory).
- `USER` -- A directory user.
- `GROUP` -- A directory (security) group.

### spec.description

`string`

Free-text description recorded on the assignment, visible in the portal's
IAM blade and via the authorization API. Use it to record WHY the grant
exists ("CI deploy identity needs pull access to the shared registry") --
future operators auditing access will only see scope/role/principal
otherwise.

### spec.condition

`string`

An Azure attribute-based access control (ABAC) condition that narrows
when the role's permissions apply, expressed in Azure's condition syntax.
Example (only allow reading blobs tagged Project=Cascade):
  ((!(ActionMatches{'Microsoft.Storage/storageAccounts/blobServices/containers/blobs/read'}))
   OR (@Resource[Microsoft.Storage/storageAccounts/blobServices/containers/blobs/tags:Project<$key_case_sensitive$>] StringEquals 'Cascade'))
Conditions are currently supported on storage data-plane roles and a
growing set of others; an unsupported role+condition pair fails at deploy
with a clear ARM error.

### spec.conditionVersion

`string`

The version of the condition syntax. Azure currently defines "2.0" as the
only generally available version ("1.0" is legacy). When condition is set
and this is omitted, the platform's behavior matches Azure's: version
"2.0" is applied.

- rule: {"ignore":"IGNORE_IF_ZERO_VALUE","string":{"in":["1.0","2.0"]}}

### spec.delegatedManagedIdentityResourceId

`string`

For cross-tenant scenarios only: the resource ID of the managed identity
(in the delegated/managing tenant) performing this assignment on a
resource that lives in another tenant, e.g. an Azure Lighthouse managed
service provider granting roles in a customer tenant. Leave empty for
everything single-tenant -- which is almost every deployment.

### spec.skipServicePrincipalAadCheck

`bool`

Set to true when the principal is a service principal or managed identity
that was created moments before this assignment. Azure AD replicates new
principals asynchronously, and an assignment racing that replication fails
with "PrincipalNotFound"; this flag tells Azure to accept the assignment
without the directory existence check. Only valid when the principal is a
service principal (setting it forces principal_type to ServicePrincipal
on the request) -- setting it for a user or group makes the request fail.

### spec.name

`string`

A stable UUID for the assignment's ARM resource name. Azure identifies a
role assignment by a GUID; when omitted (recommended), a random one is
generated at deploy time. Pin it only when an externally-defined GUID must
be preserved, e.g. recreating an assignment that other tooling references
by its full ARM ID.

- rule: {"ignore":"IGNORE_IF_ZERO_VALUE","string":{"uuid":true}}

## Validation Rules

- `spec.role_exactly_one`: Specify the role exactly one way: either role_definition_name (built-in roles like 'Reader') or role_definition_id (custom roles), not both and not neither
- `spec.condition_version_requires_condition`: condition_version is only meaningful together with condition -- either add an ABAC condition expression or remove condition_version

## Outputs

Reference an output from another manifest as `valueFrom: {kind: AzureRoleAssignment, name: <resource-name>, fieldPath: status.outputs.<output>}`.

| Output | Type | Description |
|---|---|---|
| `status.outputs.role_assignment_id` | `string` | The fully-scoped Azure Resource Manager ID of the role assignment. Format: {scope}/providers/Microsoft.Authorization/roleAssignments/{name} This is the identifier the authorization API uses to fetch or delete the assignment. |
| `status.outputs.name` | `string` | The assignment's ARM resource name -- the GUID that (with the scope) uniquely identifies it. Either the pinned spec.name or the GUID generated at deploy time. |
| `status.outputs.scope` | `string` | The scope the role was granted at, as resolved at deploy time. |
| `status.outputs.role_definition_id` | `string` | The fully-scoped role definition ID that was bound. Populated whether the spec referenced the role by name or by ID -- when a built-in role name was given, this carries the ID Azure resolved it to. |
| `status.outputs.principal_id` | `string` | The Azure AD object ID of the principal the role was granted to. |
| `status.outputs.principal_type` | `string` | The principal type Azure recorded for the assignment (User, Group, or ServicePrincipal) -- useful when the spec omitted principal_type and Azure inferred it from the directory. |

## References

Fields that can point at another resource's outputs:

| Field | Kind | Output |
|---|---|---|
| `spec.scope` | AzureResourceGroup | `status.outputs.resource_group_id` |
| `spec.roleDefinitionId` | AzureRoleDefinition | `status.outputs.role_definition_id` |
| `spec.principalId` | AzureUserAssignedIdentity | `status.outputs.principal_id` |

## See Also

- [Overview](../README.md)
