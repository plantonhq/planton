# AzureCosmosdbSqlRoleDefinition - Pulumi Module

Pulumi implementation for the AzureCosmosdbSqlRoleDefinition
component.

## Architecture

```
cosmosdb.SqlRoleDefinition (single resource)
```

## Key Design Decisions

- **The parent is addressed by ARM id** (`cosmosdb_account_id`) -- the
  provider addresses Cosmos RBAC resources by the (resource group,
  account, GUID) trio, so both names are parsed from the resolved
  account id in `parse.go` with the same anchored semantics as the
  Terraform module's regexes; a malformed id fails loudly instead of
  computing wrong names.
- **The type enum maps to ARM's exact wire vocabulary**
  (CUSTOM_ROLE -> CustomRole, BUILT_IN_ROLE -> BuiltInRole) and
  unspecified sends nothing, so the provider's own CustomRole default
  applies -- identical behavior to the Terraform module.
- **The pinned GUID is sent only when set** -- unset lets the provider
  generate a random GUID at create time; assignments consume the full
  resource ID, not the GUID, so pinning is only for externally-known
  identities.
- **The `role_definition_id` output carries the fully-scoped ARM id**
  -- exactly what an AzureCosmosdbSqlRoleAssignment's
  `role_definition_id` field consumes; the bare GUID rides separately
  as `role_definition_guid`.
- **No Azure tags**: ARM does not support tags on Cosmos child
  resources; the platform's identity tags live on the account.

## Provider

The Azure provider is built by the shared
`pulumiazureprovider.Get(ctx, iacInput.ProviderConfig)` builder, which
dispatches static client-secret, keyless web-identity (OIDC), and
ambient credential chains. Never construct a provider inline.
