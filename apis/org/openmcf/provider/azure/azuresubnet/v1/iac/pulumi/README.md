# AzureSubnet Pulumi Module

Pulumi implementation for the AzureSubnet deployment component.

## Architecture

The module creates:

- `network.Subnet` -- an Azure Subnet within an existing Virtual Network

## Package Structure

```
iac/pulumi/
├── main.go              # Entrypoint: loads stack input, calls module
├── Pulumi.yaml          # Pulumi project configuration
├── Makefile             # Build and test targets
├── debug.sh             # Delve debugger script
├── overview.md          # Architecture overview
└── module/
    ├── main.go          # Resources(): creates subnet with optional delegation
    ├── locals.go        # initializeLocals(): parses VNet ID, builds tags
    └── outputs.go       # Output key constants
```

## Resources Created

| Resource | Pulumi Type | Description |
|----------|-------------|-------------|
| Subnet | `network.Subnet` | Subnet within a VNet with optional delegation |

## Key Patterns

| Pattern | Detail |
|---------|--------|
| VNet name extraction | ARM ID parsed via string split in locals |
| Address prefix wrapping | Singular string wrapped in `[]string` for provider |
| Conditional delegation | nil check on `spec.Delegation` gates the block |
| Default handling | `GetPrivateEndpointNetworkPolicies()` returns "Disabled" via framework |

## Local Development

```bash
make deps    # Tidy Go modules
make build   # Build module and entrypoint
make test    # Run tests
```
