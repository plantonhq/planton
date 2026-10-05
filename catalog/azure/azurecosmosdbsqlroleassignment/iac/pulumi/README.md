# AzureCosmosdbSqlRoleAssignment - Pulumi Module

Pulumi implementation for the AzureCosmosdbSqlRoleAssignment
kind.

## Architecture

```
cosmosdb.SqlRoleAssignment (single resource)
```

## Key Design Decisions

- **The parent is addressed by ARM id** (`cosmosdb_account_id`) -- the
  provider addresses Cosmos RBAC resources by the (resource group,
  account, GUID) trio, so both names are parsed from the resolved
  account id in `parse.go` with the same anchored semantics as the
  Terraform module's regexes; a malformed id fails loudly instead of
  computing wrong names.
- **The role definition id is validated before any ARM call** -- the
  Terraform provider checks the
  `.../databaseAccounts/{account}/sqlRoleDefinitions/{guid}` shape at
  plan time; the Pulumi SDK carries no equivalent, so `parse.go`
  enforces the same contract here to keep both engines failing loudly
  and early on a malformed literal instead of surfacing a late ARM
  error.
- **The pinned GUID is sent only when set** -- unset lets the provider
  generate a random GUID at create time.
- **No Azure tags**: ARM does not support tags on Cosmos child
  resources; the platform's identity tags live on the account.

## Provider

The Azure provider is built by the shared
`pulumiazureprovider.Get(ctx, iacInput.ProviderConfig)` builder, which
dispatches static client-secret, keyless web-identity (OIDC), and
ambient credential chains. Never construct a provider inline.
