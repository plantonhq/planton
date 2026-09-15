# AzureCosmosdbSqlRoleDefinition

> Generated from the protobuf schema by `make generate-reference` -- do not
> edit by hand. To change a fact on this page, change the proto field comment
> or validation rule it is derived from, then regenerate.

**apiVersion**: `azure.planton.dev/v1alpha1`

**AzureCosmosdbSqlRoleDefinitionSpec** defines the configuration for
creating a Cosmos DB SQL (NoSQL) API role definition: a named, reusable
bundle of DATA-PLANE permissions that Microsoft Entra principals can then
be granted through Cosmos DB SQL role assignments.

Cosmos DB carries its own RBAC system, separate from ARM RBAC. An ARM
role assignment (even Owner on the account) governs MANAGING the account
-- it grants no ability to read or write the documents inside it. Data
access for Entra identities is governed exclusively by this system:
Cosmos-scoped role definitions bound to principals by Cosmos-scoped role
assignments. Together with the account's `local_authentication` switch,
this is Cosmos DB's keyless story -- keys off, workload identities in,
every data grant explicit and auditable.

Azure ships two built-in data roles per account:
- **Cosmos DB Built-in Data Reader** (`00000000-0000-0000-0000-000000000001`):
  read documents, run queries, read change feeds and metadata.
- **Cosmos DB Built-in Data Contributor** (`00000000-0000-0000-0000-000000000002`):
  everything the reader allows plus create/replace/upsert/delete items
  and manage containers.
Assign a built-in directly by its ID (no definition resource needed --
they already exist in every account). Create a CUSTOM definition when
neither fits: "read-only on one container", "write items but never
delete", "metadata-only for a monitoring probe".

The definition grants nothing by itself -- permissions take effect only
when an AzureCosmosdbSqlRoleAssignment binds it to a principal at a
scope. The definition's `role_definition_id` output is exactly what an
assignment's `role_definition_id` field consumes.

This RBAC surface exists on SQL (NoSQL) API accounts
(kind GLOBAL_DOCUMENT_DB); Mongo/Cassandra/Gremlin/Table accounts carry
their own mechanisms.

## Example

```yaml
apiVersion: azure.planton.dev/v1alpha1
kind: AzureCosmosdbSqlRoleDefinition
metadata:
  name: test-cosmos-sql-role-definition
spec:
  cosmosdbAccountId:
    value: /subscriptions/00000000-0000-0000-0000-000000000000/resourceGroups/test-rg/providers/Microsoft.DocumentDB/databaseAccounts/planton-hack-cosmos
  roleName: hack-container-writer
  # Exercises the explicit-type seam (unspecified deploys CustomRole).
  type: CUSTOM_ROLE
  # Exercises both assignable-scope shapes: the whole account and a
  # database path composed on the account ID.
  assignableScopes:
    - value: /subscriptions/00000000-0000-0000-0000-000000000000/resourceGroups/test-rg/providers/Microsoft.DocumentDB/databaseAccounts/planton-hack-cosmos
    - value: /subscriptions/00000000-0000-0000-0000-000000000000/resourceGroups/test-rg/providers/Microsoft.DocumentDB/databaseAccounts/planton-hack-cosmos/dbs/app-data
  # Exercises multiple additive permission blocks.
  permissions:
    - dataActions:
        - Microsoft.DocumentDB/databaseAccounts/readMetadata
        - Microsoft.DocumentDB/databaseAccounts/sqlDatabases/containers/items/*
    - dataActions:
        - Microsoft.DocumentDB/databaseAccounts/sqlDatabases/containers/executeQuery
  # Exercises the pinned-GUID seam.
  roleDefinitionId: 9b7f3f6a-2f0e-4b9a-8f0d-2f6a8f0d2f6a
```

## Spec Fields

| Path | Type | Required | Default | References |
|---|---|---|---|---|
| `spec.cosmosdbAccountId` | `string \| valueFrom` | yes |  | AzureCosmosdbAccount (`status.outputs.cosmosdb_account_id`) |
| `spec.roleName` | `string` | yes |  |  |
| `spec.type` | `enum` |  |  |  |
| `spec.assignableScopes` | `[]string \| valueFrom` | yes |  | AzureCosmosdbAccount (`status.outputs.cosmosdb_account_id`) |
| `spec.permissions` | `[]AzureCosmosdbSqlRoleDefinitionPermission` | yes |  |  |
| `spec.permissions[].dataActions` | `[]string` | yes |  |  |
| `spec.roleDefinitionId` | `string` |  |  |  |

## Field Details

### spec.cosmosdbAccountId

`string | valueFrom` · required

The Cosmos DB account the role definition lives in, by ARM ID.
References an AzureCosmosdbAccount's cosmosdb_account_id output so the
account and its RBAC surface compose in one manifest set. The account
must be a GLOBAL_DOCUMENT_DB (SQL API) account. Fixed at creation.

- references: AzureCosmosdbAccount (`status.outputs.cosmosdb_account_id`)
- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: AzureCosmosdbAccount, name: <that resource's name>, fieldPath: status.outputs.cosmosdb_account_id}} -- a bare string does not parse

### spec.roleName

`string` · required

The role's display name -- what `az cosmosdb sql role definition list`
and the portal's data-plane RBAC views show, and how operators tell
grants apart when auditing. Must be unique among the account's role
definitions. Renaming is an in-place update; assignments track the
role by GUID, not by name.

- rule: {"required":true,"string":{"minLen":"1"}}

### spec.type

`enum`

The definition's type. Leave unspecified for CUSTOM_ROLE -- the only
shape organizations author (Azure's own default). BUILT_IN_ROLE is
accepted by ARM for completeness but built-in definitions already
exist in every account; creating one is never required to assign a
built-in role (reference its well-known ID from the assignment
instead). Fixed at creation.

Allowed values (use exactly as shown):

- `azure_cosmosdb_sql_role_definition_type_unspecified` -- Not specified: deploys CustomRole -- Azure's own default and the only type organizations author.
- `CUSTOM_ROLE` -- A customer-authored role definition (the default).
- `BUILT_IN_ROLE` -- A Microsoft-curated role definition. Built-ins already exist in every account -- reference their well-known IDs from assignments instead of creating one.

### spec.assignableScopes

`[]string | valueFrom` · required

The fully-qualified scopes at or below which assignments of this role
may be created: the account itself, one of its databases
(`{account-id}/dbs/{database}`), or a single container
(`{account-id}/dbs/{database}/colls/{container}`). Scopes ABOVE the
account (subscription, resource group) are not enforceable here --
Cosmos data-plane RBAC starts at the account boundary.

At least one scope is required. Each entry is a literal path or a
reference to a resource's ID output; the default reference is the
account's own ARM ID -- assignable anywhere in the account, the
common posture. The referenced paths need not exist yet. Updatable
in place.

Database- or container-level entries are literal values composed on
the account ID (references cannot append path suffixes), e.g.:
  assignableScopes:
    - value: /subscriptions/{sub}/resourceGroups/{rg}/providers/Microsoft.DocumentDB/databaseAccounts/{account}/dbs/app-data

Where a role MAY BE ASSIGNED is not where the definition lives (that is
`cosmosdb_account_id`), so on a diagram these references are access,
not placement.

- references: AzureCosmosdbAccount (`status.outputs.cosmosdb_account_id`)
- rule: {"repeated":{"minItems":"1"}}
- rule: write as {value: <literal>} or {valueFrom: {kind: AzureCosmosdbAccount, name: <that resource's name>, fieldPath: status.outputs.cosmosdb_account_id}} -- a bare string does not parse

### spec.permissions

`[]AzureCosmosdbSqlRoleDefinitionPermission` · required

The permission blocks defining what this role allows. Azure evaluates
them as a union: an operation is permitted if any block's data actions
match it. One block is the norm; multiple blocks are legal and merely
additive -- ARM models permissions as a list, so this does too. At
least one block is required (a Cosmos role without permissions grants
nothing and ARM's contract requires the list).

- rule: {"repeated":{"minItems":"1"}}

### spec.permissions[].dataActions

`[]string` · required

The data actions this block allows. At least one is required, each a
non-empty `Microsoft.DocumentDB/databaseAccounts/...` operation
pattern (wildcards allowed on the trailing operation).

- rule: {"repeated":{"minItems":"1","items":{"string":{"minLen":"1"}}}}

### spec.roleDefinitionId

`string`

A stable UUID for the role definition's ARM resource name. Cosmos
identifies a role definition by a GUID; when omitted (recommended), a
random one is generated at deploy time. Pin it only when an
externally-known GUID must be preserved -- e.g. recreating a
definition that existing assignments or tooling reference by its full
ARM ID. Changing a pinned GUID replaces the definition.

- rule: {"ignore":"IGNORE_IF_ZERO_VALUE","string":{"uuid":true}}

## Outputs

Reference an output from another manifest as `valueFrom: {kind: AzureCosmosdbSqlRoleDefinition, name: <resource-name>, fieldPath: status.outputs.<output>}`.

| Output | Type | Description |
|---|---|---|
| `status.outputs.role_definition_id` | `string` | The fully-scoped Azure Resource Manager ID of the role definition. Format: /subscriptions/{sub}/resourceGroups/{rg}/providers/Microsoft.DocumentDB/databaseAccounts/{account}/sqlRoleDefinitions/{guid} This is the value an AzureCosmosdbSqlRoleAssignment's role_definition_id field consumes -- reference it to grant this role to a principal. |
| `status.outputs.role_definition_guid` | `string` | The definition's GUID resource name -- either the pinned spec.role_definition_id or the GUID generated at deploy time. |
| `status.outputs.role_name` | `string` | The role's display name as deployed (what the account's data-plane RBAC listings show). |
| `status.outputs.cosmosdb_account_name` | `string` | The name of the Cosmos DB account the definition lives in, parsed from the resolved account ID -- saves consumers a second reference when they need the account/definition pair. |

## References

Fields that can point at another resource's outputs:

| Field | Kind | Output |
|---|---|---|
| `spec.cosmosdbAccountId` | AzureCosmosdbAccount | `status.outputs.cosmosdb_account_id` |
| `spec.assignableScopes` | AzureCosmosdbAccount | `status.outputs.cosmosdb_account_id` |

## Referenced By

Fields on other kinds that can point at this resource:

| Kind | Field | Reads |
|---|---|---|
| AzureCosmosdbSqlRoleAssignment | `spec.roleDefinitionId` | `status.outputs.role_definition_id` |

## See Also

- [Overview](../README.md)
