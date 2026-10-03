# AzureMssqlElasticPool - Pulumi Module

Pulumi implementation for the AzureMssqlElasticPool component.

## Architecture

```
mssql.ElasticPool (single resource)
```

## Key Design Decisions

- **The server is addressed by name + resource group** (this resource's
  azurerm contract), both derived in `locals.go` from the spec's single
  parent FK (the server's ARM id) -- the parent-derivation precedent.
- **The sku tier and hardware family are derived from `sku_name`**
  through exhaustive vocabularies (pure functions of the name, verified
  against azurerm's own validation helpers), so a mismatched combination
  is unrepresentable. DTU pools carry no family.
- **`maintenance_configuration_name` is presence-guarded** to its spec
  default (`SQL_Default`): IaC inputs built from a manifest do not
  materialize proto defaults.

## Provider

The Azure provider is built by the shared
`pulumiazureprovider.Get(ctx, iacInput.ProviderConfig)` builder, which
dispatches static client-secret, keyless (web identity), and ambient
credential chains. Never construct the provider inline.

## Running Locally

```bash
# Build
make build

# Run with Pulumi
make run
```
