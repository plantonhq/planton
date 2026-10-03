# AzureContainerAppEnvironment Pulumi Module

The Pulumi (Go) implementation of the `AzureContainerAppEnvironment` component.

## Structure

```
pulumi/
├── main.go          # Entrypoint (loads IaC input, calls module)
├── Pulumi.yaml      # Pulumi project configuration
├── Makefile         # Build/test targets
└── module/
    ├── main.go      # Environment + optional custom-domain association
    ├── locals.go    # Tag merge + enum wire-value maps
    └── outputs.go   # Output key constants
```

## Resources Created

| Resource | Pulumi Type | Condition |
|----------|-------------|-----------|
| Container App Environment | `containerapp.Environment` | Always |
| Custom DNS suffix | `containerapp.EnvironmentCustomDomain` | `spec.custom_domain` set |

## Behavior Notes

- The Azure provider comes from the shared `pulumiazureprovider.Get` builder, which resolves static client-secret, keyless web-identity, or ambient credentials from the IaC input.
- Enum wire maps are spelled out row by row in `locals.go` so a vocabulary drift fails at preview time.
- The system-assigned identity's principal id is exported empty when no system identity exists, keeping the output shape constant.

## Build

```bash
make build
```
