# AzureNetworkSecurityGroup Pulumi Module

Pulumi implementation for the AzureNetworkSecurityGroup deployment component.

## Architecture

The module creates:

- `network.NetworkSecurityGroup` -- NSG resource (shell)
- `network.NetworkSecurityRule` -- One per security rule in the spec

## Package Structure

```
iac/pulumi/
├── main.go              # Entrypoint: loads stack input, calls module
├── Pulumi.yaml          # Pulumi project configuration
├── Makefile             # Build and test targets
├── debug.sh             # Delve debugger script
├── overview.md          # Architecture overview
└── module/
    ├── main.go          # Resources(): creates NSG + rules
    ├── locals.go        # initializeLocals(): parses input, builds tags
    └── outputs.go       # Output key constants
```

## Resources Created

| Resource | Pulumi Type | Description |
|----------|-------------|-------------|
| NSG | `network.NetworkSecurityGroup` | Network Security Group shell |
| Rules | `network.NetworkSecurityRule` | One per security rule (separate resources) |

## Local Development

```bash
make deps    # Tidy Go modules
make build   # Build module and entrypoint
make test    # Run tests
```
