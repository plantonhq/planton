# AzureWebApplicationFirewallPolicy - Pulumi Module

Pulumi implementation for the AzureWebApplicationFirewallPolicy deployment
component.

## Architecture

```
waf.Policy (the single resource: custom rules + managed rules + settings)
```

The policy carries no back-references -- Application Gateways attach it by
referencing its ID, so the module creates exactly one resource.

## Key Design Decisions

- **Every enum maps through an exhaustive vocabulary** in `locals.go`
  (rule types, actions, operators, variables, transforms, rule-set types,
  selector operators, scrubbing variables, modes) -- a missing entry would
  silently drop a setting, so the maps are complete by construction.
- **Unspecified rule-set type materializes OWASP** and **unspecified mode
  materializes Prevention** -- azurerm's own defaults, sent explicitly so
  both engines produce the same ARM payload.
- **Presence guards on every optional-with-default field** (rule enabled,
  settings dials, excluded-rule-set version): IaC inputs built from a
  manifest do not materialize proto defaults, so unset falls back to the
  documented default explicitly.
- **`file_upload_enforcement` is forwarded only on explicit presence** --
  it is only honored with OWASP 3.2, and materializing a default would
  error on older rule sets (mirrored in the Terraform module).

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
