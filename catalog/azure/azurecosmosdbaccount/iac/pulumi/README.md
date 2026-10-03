# AzureCosmosdbAccount - Pulumi Module

Pulumi implementation for the AzureCosmosdbAccount deployment
component.

## Architecture

```
cosmosdb.Account (single resource)
```

Databases and containers are their own kinds
(AzureCosmosdbSqlDatabase / AzureCosmosdbSqlContainer /
AzureCosmosdbMongoDatabase / AzureCosmosdbMongoCollection) referencing
this account, so the module creates exactly one resource.

## Key Design Decisions

- **Enum maps are exhaustive by construction** -- every spec enum
  (kind, consistency level, capabilities, backup type/tier/redundancy,
  Mongo server version, identity types, analytical schema, create
  mode) has a complete verbatim map to ARM's wire values, including
  the two capability values that break the EnableX convention
  (`MongoDBv3.4`, `mongoEnableDocLevelTTL`).
- **Capabilities are declared, never injected** -- a MONGO_DB account
  declares ENABLE_MONGO itself; the module never rewrites what the
  user declared (most capability changes recreate the account, so
  hidden capabilities would hide recreate triggers).
- **Presence-guarded defaults** -- IaC inputs built from a manifest
  do not materialize proto defaults, so the true-default bools
  (public network access, access-key metadata writes, local auth) and
  the BoundedStaleness dials (5/100) fall back to the proto defaults
  explicitly instead of sending zero values.
- **The default identity composes the wire string** -- the spec models
  a type enum plus a real AzureUserAssignedIdentity reference; the
  module renders "UserAssignedIdentity=<resolved id>" for ARM.
- **CMK rides the versionless key URI** so Key Vault rotation
  propagates without touching the account; sent only when set
  (ForceNew).
- **Local auth is inverted once** -- the spec's
  `local_authentication_enabled` (the non-deprecated form) maps onto
  the provider's `LocalAuthenticationDisabled`.
- **The keys and connection strings are exported as the credential
  surface** -- all four keys and all eight connection strings
  (secret-bearing); they stop authenticating when local auth is off.

## Provider

The Azure provider is built by the shared
`pulumiazureprovider.Get(ctx, iacInput.ProviderConfig)` builder, which
dispatches static client-secret, keyless web-identity (OIDC), and
ambient credential chains. Never construct a provider inline.
