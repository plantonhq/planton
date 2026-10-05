# AzureCosmosdbSqlDatabase - Pulumi Module

Pulumi implementation for the AzureCosmosdbSqlDatabase
kind.

## Architecture

```
cosmosdb.SqlDatabase (single resource)
```

## Key Design Decisions

- **The parent is addressed by ARM id** (`cosmosdb_account_id`) -- the
  provider addresses Cosmos children by the (resource group, account,
  name) trio, so both names are parsed from the resolved account id in
  `parse.go` with the same anchored semantics as the Terraform module's
  regexes; a malformed id fails loudly instead of computing wrong
  names. The spec models one authoritative parent reference and the
  module derives the rest.
- **Throughput is sent only when set** -- serverless accounts reject
  provisioned throughput at apply, and unset means each container
  brings its own dedicated RU/s. The spec enforces the
  fixed-XOR-autoscale contract before the plan ever runs.
- **No endpoint or credential outputs on purpose** -- connectivity and
  keys live on the ACCOUNT (AzureCosmosdbAccount's endpoint and key
  outputs); the database is addressed inside that connection by name,
  which is why the module exports `sql_database_id`,
  `sql_database_name`, and `cosmosdb_account_name` and nothing else.
- **No Azure tags**: ARM does not support tags on Cosmos child
  resources; the platform's identity tags live on the account.

## Provider

The Azure provider is built by the shared
`pulumiazureprovider.Get(ctx, iacInput.ProviderConfig)` builder, which
dispatches static client-secret, keyless web-identity (OIDC), and
ambient credential chains. Never construct a provider inline.
