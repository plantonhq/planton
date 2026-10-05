# AzureCosmosdbSqlContainer - Pulumi Module

Pulumi implementation for the AzureCosmosdbSqlContainer
kind.

## Architecture

```
cosmosdb.SqlContainer (single resource)
```

## Key Design Decisions

- **The parent is addressed by ARM id** (`sql_database_id`) -- the spec
  models a single parent reference; the database, account, and
  resource-group names the provider requires are parsed from the
  resolved id with the same anchored semantics as the Terraform
  module's regexes, so a malformed id fails loudly on both engines
  identically.
- **Unset partition key kind materializes Hash** -- the provider's own
  default; MULTI_HASH hierarchical keys require version 2 (the spec
  enforces the pairing before the deploy runs).
- **Throughput is sent only when set** -- serverless accounts reject
  provisioned throughput, and unset means the container shares the
  database's throughput; fixed XOR autoscale is a spec rule.
- **TTL and analytical TTL are presence-guarded** -- unset is genuinely
  "off", never a zero sent to the API.
- **A declared indexing policy replaces the default wholesale** --
  tuned policies include "/*" and exclude exceptions; the module passes
  the policy through without invention.
- **No endpoint/credential outputs** -- connectivity and keys live on
  the ACCOUNT; the container addresses by name inside that connection.
- **No Azure tags** -- ARM does not support tags on Cosmos child
  resources; the platform's identity tags live on the account.

## Provider

The Azure provider is built by the shared
`pulumiazureprovider.Get(ctx, iacInput.ProviderConfig)` builder, which
dispatches static client-secret, keyless web-identity (OIDC), and
ambient credential chains. Never construct a provider inline.
