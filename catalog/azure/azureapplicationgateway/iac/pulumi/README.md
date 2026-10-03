# AzureApplicationGateway - Pulumi Module

Pulumi implementation for the AzureApplicationGateway deployment
component.

## Architecture

```
network.ApplicationGateway (one atomic resource; sub-objects wired by name)
```

The spec's blocks translate through per-block builders (`builders.go`) --
frontends, ports, pools, backend settings, listeners (L7 + L4), routing
rules (L7 + L4), path maps, probes, certificates, SSL profiles/policies,
redirects, rewrite rule sets, Private Link, and custom error pages.

## Key Design Decisions

- **Every enum maps through an exhaustive vocabulary** in `locals.go`
  (SKU, protocols, rule types, allocations, identity types, TLS policy
  types and versions, redirect types, URL components, status codes) -- a
  missing entry would silently drop a setting.
- **The gateway IP configuration is derived** (`{name}-gateway-ip`) from
  the spec's single dedicated-subnet FK, byte-identical to the Terraform
  module's derivation, so the two engines converge on the same ARM
  object.
- **Map outputs via ApplyT** -- `backend_address_pool_ids` and
  `frontend_ip_configuration_ids` are read back from the created
  resource's computed sub-object IDs and keyed by name (the load-balancer
  precedent).
- **Presence guards on optional-with-default fields** (request timeout,
  L4 backend timeout, HTTP/2): IaC inputs built from a manifest do not
  materialize proto defaults, so unset falls back to the documented
  default explicitly.
- **Optional strings forwarded only when non-empty** so the ARM payload
  stays free of empty no-op properties.

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
