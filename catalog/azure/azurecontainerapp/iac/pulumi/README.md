# AzureContainerApp Pulumi Module

The Pulumi (Go) implementation of the `AzureContainerApp` kind.

## Structure

```
pulumi/
├── main.go          # Entrypoint (loads IaC input, calls module)
├── Pulumi.yaml      # Pulumi project configuration
├── Makefile         # Build/test targets
└── module/
    ├── main.go      # The container app (template, ingress, secrets, registries, dapr, identity)
    ├── locals.go    # Tag merge + enum wire-value maps
    └── outputs.go   # Output key constants
```

## Resources Created

| Resource | Pulumi Type | Condition |
|----------|-------------|-----------|
| Container App | `containerapp.App` | Always |

## Behavior Notes

- The Azure provider comes from the shared `pulumiazureprovider.Get` builder (static client-secret, keyless web-identity, or ambient credentials).
- Enum wire maps are spelled out row by row in `locals.go`; the ingress mTLS vocabulary is lowercase on the wire (`accept`/`require`/`ignore`).
- Per-probe-type initial-delay defaults (1 liveness, 0 readiness/startup) and the scaler/replica dials are presence-guarded -- unset fields deploy the spec's documented default, never the Go zero value.
- `ingress_fqdn` and `identity_principal_id` export empty when ingress / a system identity is absent, keeping the output shape constant.

## Build

```bash
make build
```
