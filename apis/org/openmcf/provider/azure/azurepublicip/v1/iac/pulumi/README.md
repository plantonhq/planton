# AzurePublicIp Pulumi Module

Pulumi implementation for the AzurePublicIp deployment component.

## Architecture

The module creates:

- `network.PublicIp` -- Standard SKU, Static allocation Public IP Address

## Package Structure

```
iac/pulumi/
├── main.go              # Entrypoint: loads stack input, calls module
├── Pulumi.yaml          # Pulumi project configuration
├── Makefile             # Build and test targets
├── debug.sh             # Delve debugger script
├── overview.md          # Architecture overview
└── module/
    ├── main.go          # Resources(): creates Public IP
    ├── locals.go        # initializeLocals(): parses input, builds tags
    └── outputs.go       # Output key constants
```

## Resources Created

| Resource | Pulumi Type | Description |
|----------|-------------|-------------|
| Public IP | `network.PublicIp` | Standard SKU static public IPv4 address |

## Hardcoded Values

| Field | Value | Reason |
|-------|-------|--------|
| SKU | Standard | Basic SKU retired Sept 2025 |
| Allocation Method | Static | Standard SKU requires Static |

## Local Development

```bash
make deps    # Tidy Go modules
make build   # Build module and entrypoint
make test    # Run tests
```
